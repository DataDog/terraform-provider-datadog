package commandauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestHelperCommand is not a test: the tests below run the test binary as the
// token command. It prints COMMAND_AUTH_OUTPUT, writes COMMAND_AUTH_STDERR to
// stderr, counts its runs in COMMAND_AUTH_COUNT and exits COMMAND_AUTH_EXIT.
func TestHelperCommand(t *testing.T) {
	if os.Getenv("COMMAND_AUTH_HELPER") != "1" {
		return
	}
	if f := os.Getenv("COMMAND_AUTH_COUNT"); f != "" {
		b, _ := os.ReadFile(f)
		n, _ := strconv.Atoi(string(b))
		_ = os.WriteFile(f, []byte(strconv.Itoa(n+1)), 0o600)
	}
	fmt.Fprint(os.Stdout, os.Getenv("COMMAND_AUTH_OUTPUT"))
	fmt.Fprint(os.Stderr, os.Getenv("COMMAND_AUTH_STDERR"))
	code, _ := strconv.Atoi(os.Getenv("COMMAND_AUTH_EXIT"))
	os.Exit(code)
}

// helperArgv returns a command that runs TestHelperCommand. name keeps each
// test's cache entry separate.
func helperArgv(t *testing.T, name string) []string {
	t.Helper()
	t.Setenv("COMMAND_AUTH_HELPER", "1")
	return []string{os.Args[0], "-test.run=^TestHelperCommand$", "--", t.Name() + "/" + name}
}

func runs(t *testing.T, counter string) int {
	t.Helper()
	b, err := os.ReadFile(counter)
	if os.IsNotExist(err) {
		return 0
	}
	require.NoError(t, err)
	n, err := strconv.Atoi(string(b))
	require.NoError(t, err)
	return n
}

func jwt(exp time.Time) string {
	payload := fmt.Sprintf(`{"exp":%d}`, exp.Unix())
	return "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".c2ln"
}

func TestAuthenticateRunsCommandAndCaches(t *testing.T) {
	counter := filepath.Join(t.TempDir(), "count")
	t.Setenv("COMMAND_AUTH_COUNT", counter)
	exp := time.Now().Add(time.Hour).Truncate(time.Second)
	token := jwt(exp)
	t.Setenv("COMMAND_AUTH_OUTPUT", token+"\n")

	config, err := GetDelegatedTokenConfig(helperArgv(t, "cache"))
	require.NoError(t, err)
	require.Equal(t, Provider, config.Provider)

	for range 3 {
		creds, err := config.ProviderAuth.Authenticate(context.Background(), config)
		require.NoError(t, err)
		require.Equal(t, token, creds.DelegatedToken)
		require.Equal(t, exp.Add(-expirySkew), creds.Expiration)
	}
	require.Equal(t, 1, runs(t, counter), "a valid cached token must not rerun the command")

	// A second configuration for the same command, as the other half of the
	// muxed provider builds, shares the cache.
	other, err := GetDelegatedTokenConfig(helperArgv(t, "cache"))
	require.NoError(t, err)
	_, err = other.ProviderAuth.Authenticate(context.Background(), other)
	require.NoError(t, err)
	require.Equal(t, 1, runs(t, counter))
}

func TestAuthenticateRerunsAfterExpiry(t *testing.T) {
	counter := filepath.Join(t.TempDir(), "count")
	t.Setenv("COMMAND_AUTH_COUNT", counter)
	// Within the skew of expiry, so the cached token is already stale.
	t.Setenv("COMMAND_AUTH_OUTPUT", jwt(time.Now().Add(30*time.Second)))

	config, err := GetDelegatedTokenConfig(helperArgv(t, "expiry"))
	require.NoError(t, err)
	for range 2 {
		_, err := config.ProviderAuth.Authenticate(context.Background(), config)
		require.NoError(t, err)
	}
	require.Equal(t, 2, runs(t, counter))
}

func TestExpiration(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	exp := now.Add(time.Hour)
	require.Equal(t, exp.Add(-expirySkew), expiration(jwt(exp), now))
	require.Equal(t, now.Add(opaqueTokenLifetime), expiration("ddpat_opaque", now))
	require.Equal(t, now.Add(opaqueTokenLifetime), expiration("a.not-base64!.c", now))
	noExp := "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"x"}`)) + ".c2ln"
	require.Equal(t, now.Add(opaqueTokenLifetime), expiration(noExp, now))
}

func TestAuthenticateErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		output, stderr, exit, want string
	}{
		"exit status": {output: "secret-on-stdout", stderr: "not logged in", exit: "3", want: "not logged in"},
		"no output":   {output: "  \n", want: "printed no token"},
		"two lines":   {output: "first\nsecond\n", want: "only the token"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("COMMAND_AUTH_OUTPUT", tc.output)
			t.Setenv("COMMAND_AUTH_STDERR", tc.stderr)
			t.Setenv("COMMAND_AUTH_EXIT", tc.exit)
			config, err := GetDelegatedTokenConfig(helperArgv(t, name))
			require.NoError(t, err)
			_, err = config.ProviderAuth.Authenticate(context.Background(), config)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.want)
			require.NotContains(t, err.Error(), "secret-on-stdout", "stdout may hold a token and is never quoted")
			require.NotContains(t, err.Error(), "first")
		})
	}

	t.Run("missing command", func(t *testing.T) {
		config, err := GetDelegatedTokenConfig([]string{filepath.Join(t.TempDir(), "no-such-command")})
		require.NoError(t, err)
		_, err = config.ProviderAuth.Authenticate(context.Background(), config)
		require.Error(t, err)
	})

	t.Run("empty argv", func(t *testing.T) {
		_, err := GetDelegatedTokenConfig(nil)
		require.Error(t, err)
		_, err = GetDelegatedTokenConfig([]string{""})
		require.Error(t, err)
	})
}

func TestParseEnv(t *testing.T) {
	require.Equal(t, []string{"pup", "auth", "token"}, ParseEnv("  pup auth\ttoken \n"))
	require.Empty(t, ParseEnv("   "))
}
