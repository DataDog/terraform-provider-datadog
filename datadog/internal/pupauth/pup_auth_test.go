package pupauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testOrg   = "11111111-1111-4111-8111-111111111111"
	otherOrg  = "22222222-2222-4222-8222-222222222222"
	testUser  = "33333333-3333-4333-8333-333333333333"
	otherUser = "44444444-4444-4444-8444-444444444444"
)

// TestHelperPup is not a test: a stand-in `pup` script runs the test binary
// into it. It prints a JWT for PUP_TEST_ORG and PUP_TEST_USER expiring
// PUP_TEST_TTL seconds ahead, counts its runs in PUP_TEST_COUNT, and fails its
// first PUP_TEST_FAIL_FIRST runs.
func TestHelperPup(t *testing.T) {
	if os.Getenv("PUP_TEST_HELPER") != "1" {
		return
	}
	n := 1
	if b, err := os.ReadFile(os.Getenv("PUP_TEST_COUNT")); err == nil {
		prev, _ := strconv.Atoi(string(b))
		n = prev + 1
	}
	_ = os.WriteFile(os.Getenv("PUP_TEST_COUNT"), []byte(strconv.Itoa(n)), 0o600)
	if failFirst, _ := strconv.Atoi(os.Getenv("PUP_TEST_FAIL_FIRST")); n <= failFirst {
		fmt.Fprint(os.Stderr, "Error: token is expired — run 'pup auth login' to refresh")
		os.Exit(1)
	}
	if os.Getenv("PUP_TEST_OUTPUT") != "" {
		fmt.Println(os.Getenv("PUP_TEST_OUTPUT"))
		os.Exit(0)
	}
	ttl, _ := strconv.Atoi(os.Getenv("PUP_TEST_TTL"))
	fmt.Println(testJWT(time.Now().Add(time.Duration(ttl)*time.Second), os.Getenv("PUP_TEST_ORG"), os.Getenv("PUP_TEST_USER")))
	os.Exit(0)
}

func testJWT(exp time.Time, org, user string) string {
	payload := fmt.Sprintf(`{"exp":%d,"ddauth":{"org_uuid":%q,"user_uuid":%q}}`, exp.Unix(), org, user)
	return "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".c2ln"
}

// fakePup points command at a script that runs TestHelperPup, logged in as
// testUser in testOrg with an hour-long token, and resets the shared state.
func fakePup(t *testing.T) (runs func() int) {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "pup")
	require.NoError(t, os.WriteFile(script, []byte(fmt.Sprintf("#!/bin/sh\nexec %q -test.run='^TestHelperPup$' -- \"$@\"\n", os.Args[0])), 0o700))
	counter := filepath.Join(dir, "count")
	t.Setenv("PUP_TEST_HELPER", "1")
	t.Setenv("PUP_TEST_COUNT", counter)
	t.Setenv("PUP_TEST_ORG", testOrg)
	t.Setenv("PUP_TEST_USER", testUser)
	t.Setenv("PUP_TEST_TTL", "3600")
	t.Setenv("PUP_TEST_FAIL_FIRST", "0")
	t.Setenv("PUP_TEST_OUTPUT", "")

	previousCommand, previousDelays := command, retryDelays
	command, retryDelays = script, []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	ResetForTesting()
	t.Cleanup(func() {
		command, retryDelays = previousCommand, previousDelays
		ResetForTesting()
	})
	return func() int {
		b, err := os.ReadFile(counter)
		if os.IsNotExist(err) {
			return 0
		}
		require.NoError(t, err)
		n, err := strconv.Atoi(string(b))
		require.NoError(t, err)
		return n
	}
}

func authenticate(t *testing.T, orgUUID string) (string, time.Time, error) {
	t.Helper()
	config := GetDelegatedTokenConfig(orgUUID)
	creds, err := config.ProviderAuth.Authenticate(context.Background(), config)
	if err != nil {
		return "", time.Time{}, err
	}
	return creds.DelegatedToken, creds.Expiration, nil
}

// expire makes the cached token stale, as the passage of time would.
func expire() {
	stateMu.Lock()
	defer stateMu.Unlock()
	state.expiration = time.Now().Add(-time.Second)
}

func TestAuthenticateRunsPupAndCaches(t *testing.T) {
	runs := fakePup(t)
	token, exp, err := authenticate(t, "")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(token, "eyJ"))
	require.WithinDuration(t, time.Now().Add(time.Hour-expirySkew), exp, 5*time.Second)

	// A second configuration, as the other half of the muxed provider builds, shares the token.
	again, _, err := authenticate(t, "")
	require.NoError(t, err)
	require.Equal(t, token, again)
	require.Equal(t, 1, runs())
}

func TestAuthenticateRunsPupAgainAfterExpiry(t *testing.T) {
	runs := fakePup(t)
	_, _, err := authenticate(t, "")
	require.NoError(t, err)
	expire()
	_, _, err = authenticate(t, "")
	require.NoError(t, err)
	require.Equal(t, 2, runs())
}

func TestAuthenticateChecksOrgUUID(t *testing.T) {
	fakePup(t)
	_, _, err := authenticate(t, testOrg)
	require.NoError(t, err)

	ResetForTesting()
	_, _, err = authenticate(t, otherOrg)
	require.ErrorContains(t, err, "pup is logged in to organization")
	require.ErrorContains(t, err, otherOrg)
}

func TestAuthenticateStopsWhenUserChanges(t *testing.T) {
	runs := fakePup(t)
	_, _, err := authenticate(t, "")
	require.NoError(t, err)

	t.Setenv("PUP_TEST_USER", otherUser)
	expire()
	_, _, err = authenticate(t, "")
	require.ErrorContains(t, err, "pup is now logged in as a different user")
	require.Equal(t, 2, runs())

	// Logging back in as the original user lets the run continue.
	t.Setenv("PUP_TEST_USER", testUser)
	_, _, err = authenticate(t, "")
	require.NoError(t, err)
}

func TestAuthenticateRetriesAFailedRefresh(t *testing.T) {
	runs := fakePup(t)
	t.Setenv("PUP_TEST_FAIL_FIRST", "2")
	_, _, err := authenticate(t, "")
	require.NoError(t, err)
	require.Equal(t, 3, runs(), "two failures, then the retry finds the token")
}

func TestAuthenticateGivesUpAfterRetries(t *testing.T) {
	runs := fakePup(t)
	t.Setenv("PUP_TEST_FAIL_FIRST", "100")
	_, _, err := authenticate(t, "")
	require.ErrorContains(t, err, "token is expired")
	require.Equal(t, 1+len(retryDelays), runs())
}

func TestAuthenticateRejectsTokensWithoutClaims(t *testing.T) {
	for name, output := range map[string]string{
		"not a JWT":     "ddpat_opaque",
		"no ddauth":     "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"exp":4102444800}`)) + ".c2ln",
		"no exp":        "eyJhbGciOiJQUzI1NiJ9." + base64.RawURLEncoding.EncodeToString([]byte(`{"ddauth":{"org_uuid":"o","user_uuid":"u"}}`)) + ".c2ln",
		"payload bytes": "a.!!!.c",
	} {
		t.Run(name, func(t *testing.T) {
			fakePup(t)
			t.Setenv("PUP_TEST_OUTPUT", output)
			_, _, err := authenticate(t, "")
			require.Error(t, err)
			require.NotContains(t, err.Error(), output, "pup's output may be a token and is never quoted")
		})
	}
}

func TestAuthenticateWithoutPup(t *testing.T) {
	runs := fakePup(t)
	command = "pup"
	t.Setenv("PATH", t.TempDir())
	_, _, err := authenticate(t, "")
	require.ErrorContains(t, err, "use_pup is set but pup is not installed")
	require.Equal(t, 0, runs(), "a missing pup is not retried")
}
