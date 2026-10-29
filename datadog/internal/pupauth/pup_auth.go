// Package pupauth gets a Bearer token from the pup CLI (`pup auth token`), the
// OAuth session a developer starts with `pup auth login`, and gets a fresh one
// when it expires.
package pupauth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
)

const (
	// Provider is the DelegatedTokenConfig provider name for pup tokens.
	Provider = "pup"

	// commandTimeout bounds one run of pup. An interactive prompt would
	// otherwise hang the provider.
	commandTimeout = 60 * time.Second
	// expirySkew renews a token this long before its exp claim, so a request
	// started just before expiry still carries a valid token.
	expirySkew = time.Minute
	// maxStderr caps how much of pup's stderr is quoted in an error.
	maxStderr = 1024
)

// command is the program run for a token. Tests point it at a stand-in.
var command = "pup"

// retryDelays are the waits before each retry of a failed `pup auth token`.
// When several pup processes refresh the same expired token at once, only one
// refresh succeeds; a new pup process run moments later finds the token it saved.
var retryDelays = []time.Duration{500 * time.Millisecond, time.Second, 2 * time.Second}

// The SDKv2 and framework providers each build a DelegatedTokenConfig, so the
// token and the user it belongs to live here: one run of pup serves both, and
// the user is pinned for the life of the provider process.
var (
	stateMu sync.Mutex
	state   struct {
		token      string
		expiration time.Time
		userUUID   string
	}
)

// GetDelegatedTokenConfig returns a configuration whose token comes from
// `pup auth token`. A nonempty orgUUID must match the org of every token pup
// returns. The SDK calls Authenticate whenever the token is missing or past its
// expiration.
func GetDelegatedTokenConfig(orgUUID string) *datadog.DelegatedTokenConfig {
	return &datadog.DelegatedTokenConfig{
		OrgUUID:      orgUUID,
		Provider:     Provider,
		ProviderAuth: &pupAuth{orgUUID: orgUUID},
	}
}

type pupAuth struct {
	orgUUID string
}

func (a *pupAuth) Authenticate(ctx context.Context, _ *datadog.DelegatedTokenConfig) (*datadog.DelegatedTokenCredentials, error) {
	// Serialise runs, so concurrent requests at expiry run pup once.
	stateMu.Lock()
	defer stateMu.Unlock()
	if state.token == "" || !time.Now().Before(state.expiration) {
		token, err := runWithRetries(ctx)
		if err != nil {
			return nil, err
		}
		c, err := parseClaims(token)
		if err != nil {
			return nil, err
		}
		if a.orgUUID != "" && !strings.EqualFold(a.orgUUID, c.orgUUID) {
			return nil, fmt.Errorf("pup is logged in to organization %q, but org_uuid is %q", c.orgUUID, a.orgUUID)
		}
		if state.userUUID != "" && state.userUUID != c.userUUID {
			return nil, fmt.Errorf("pup is now logged in as a different user (%s) than when this run started (%s); stopping rather than continuing as someone else", c.userUUID, state.userUUID)
		}
		state.token = token
		state.expiration = c.expiration.Add(-expirySkew)
		state.userUUID = c.userUUID
	}
	return &datadog.DelegatedTokenCredentials{
		OrgUUID:        a.orgUUID,
		DelegatedToken: state.token,
		Expiration:     state.expiration,
	}, nil
}

func runWithRetries(ctx context.Context) (string, error) {
	token, err := run(ctx)
	for _, delay := range retryDelays {
		if err == nil || errors.Is(err, exec.ErrNotFound) {
			break
		}
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return "", err
		}
		token, err = run(ctx)
	}
	return token, err
}

func run(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, command, "auth", "token") // #nosec G204 -- fixed command, not configuration
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("use_pup is set but pup is not installed: %w", err)
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("`pup auth token` did not finish within %s", commandTimeout)
		}
		// stdout is never quoted: it may hold a token.
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > maxStderr {
			msg = msg[:maxStderr] + "..."
		}
		if msg != "" {
			return "", fmt.Errorf("`pup auth token` failed: %v: %s", err, msg)
		}
		return "", fmt.Errorf("`pup auth token` failed: %v", err)
	}
	token := strings.TrimSpace(stdout.String())
	if token == "" {
		return "", errors.New("`pup auth token` printed no token; run `pup auth login`")
	}
	return token, nil
}

type claims struct {
	expiration time.Time
	orgUUID    string
	userUUID   string
}

// parseClaims reads the expiry, organization and user from a pup access token.
// pup's tokens are JWTs carrying the org and user in a ddauth claim.
func parseClaims(token string) (claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims{}, errors.New("`pup auth token` did not print a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims{}, errors.New("`pup auth token` printed a JWT whose payload does not decode")
	}
	var raw struct {
		Exp    json.Number `json:"exp"`
		DDAuth struct {
			OrgUUID  string `json:"org_uuid"`
			UserUUID string `json:"user_uuid"`
		} `json:"ddauth"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return claims{}, errors.New("`pup auth token` printed a JWT whose payload is not JSON")
	}
	exp, err := raw.Exp.Int64()
	if err != nil || exp <= 0 {
		return claims{}, errors.New("pup's token has no exp claim")
	}
	if raw.DDAuth.OrgUUID == "" || raw.DDAuth.UserUUID == "" {
		return claims{}, errors.New("pup's token has no ddauth.org_uuid or ddauth.user_uuid claim")
	}
	return claims{
		expiration: time.Unix(exp, 0),
		orgUUID:    raw.DDAuth.OrgUUID,
		userUUID:   raw.DDAuth.UserUUID,
	}, nil
}

// ResetForTesting forgets the cached token and pinned user, so each test
// starts as a new provider process would.
func ResetForTesting() {
	stateMu.Lock()
	defer stateMu.Unlock()
	state.token, state.expiration, state.userUUID = "", time.Time{}, ""
}
