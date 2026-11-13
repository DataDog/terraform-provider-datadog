// Package commandauth obtains a bearer token by running a command, such as
// `pup auth token`, and runs it again when the token expires.
package commandauth

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
	// Provider is the DelegatedTokenConfig provider name for command-sourced tokens.
	Provider = "bearer_token_command"

	// commandTimeout bounds one run of the command. An interactive login prompt
	// would otherwise hang the provider.
	commandTimeout = 60 * time.Second
	// expirySkew renews a JWT this long before its exp claim, so a request
	// started just before expiry still carries a valid token.
	expirySkew = time.Minute
	// opaqueTokenLifetime is how long a token that is not a JWT is reused before
	// the command runs again. The command, not the provider, knows its lifetime.
	opaqueTokenLifetime = 5 * time.Minute
	// maxStderr caps how much of the command's stderr is quoted in an error.
	maxStderr = 1024
)

// The SDKv2 and framework providers each build a DelegatedTokenConfig, so the
// token cache lives here: one run of the command serves both.
var (
	cacheMu sync.Mutex
	cache   = map[string]*cachedToken{}
)

type cachedToken struct {
	mu         sync.Mutex
	token      string
	expiration time.Time
}

// GetDelegatedTokenConfig returns a configuration whose token comes from
// running argv. The SDK calls Authenticate whenever the token is missing or
// past its expiration.
func GetDelegatedTokenConfig(argv []string) (*datadog.DelegatedTokenConfig, error) {
	if len(argv) == 0 || argv[0] == "" {
		return nil, errors.New("bearer_token_command must name a command to run")
	}
	return &datadog.DelegatedTokenConfig{
		Provider:     Provider,
		ProviderAuth: &commandAuth{argv: append([]string(nil), argv...)},
	}, nil
}

// ParseEnv splits an environment variable's value into a command and its
// arguments at whitespace. No shell runs and no quoting is interpreted.
func ParseEnv(value string) []string {
	return strings.Fields(value)
}

type commandAuth struct {
	argv []string
}

func (a *commandAuth) Authenticate(ctx context.Context, _ *datadog.DelegatedTokenConfig) (*datadog.DelegatedTokenCredentials, error) {
	key := strings.Join(a.argv, "\x00")
	cacheMu.Lock()
	entry, ok := cache[key]
	if !ok {
		entry = &cachedToken{}
		cache[key] = entry
	}
	cacheMu.Unlock()

	// Serialise runs, so concurrent requests at expiry run the command once.
	entry.mu.Lock()
	defer entry.mu.Unlock()
	if entry.token == "" || !time.Now().Before(entry.expiration) {
		token, err := run(ctx, a.argv)
		if err != nil {
			return nil, err
		}
		entry.token = token
		entry.expiration = expiration(token, time.Now())
	}
	return &datadog.DelegatedTokenCredentials{
		DelegatedToken: entry.token,
		Expiration:     entry.expiration,
	}, nil
}

func run(ctx context.Context, argv []string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, commandTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) // #nosec G204 -- the command is the user's provider configuration
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		// stdout is never quoted: a command may print a token before failing.
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > maxStderr {
			msg = msg[:maxStderr] + "..."
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("bearer_token_command %q did not finish within %s", argv[0], commandTimeout)
		}
		if msg != "" {
			return "", fmt.Errorf("bearer_token_command %q failed: %v: %s", argv[0], err, msg)
		}
		return "", fmt.Errorf("bearer_token_command %q failed: %v", argv[0], err)
	}
	token := strings.TrimSpace(stdout.String())
	if token == "" {
		return "", fmt.Errorf("bearer_token_command %q printed no token", argv[0])
	}
	if strings.ContainsAny(token, " \t\r\n") {
		return "", fmt.Errorf("bearer_token_command %q printed more than one line or word; it must print only the token", argv[0])
	}
	return token, nil
}

// expiration reads a JWT's exp claim, less expirySkew. A token that is not a
// JWT, or has no exp, is reused for opaqueTokenLifetime.
func expiration(token string, now time.Time) time.Time {
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		if payload, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
			var claims struct {
				Exp json.Number `json:"exp"`
			}
			if json.Unmarshal(payload, &claims) == nil {
				if exp, err := claims.Exp.Int64(); err == nil && exp > 0 {
					return time.Unix(exp, 0).Add(-expirySkew)
				}
			}
		}
	}
	return now.Add(opaqueTokenLifetime)
}
