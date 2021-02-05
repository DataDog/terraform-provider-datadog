package datadog

import (
	"fmt"
	"time"
)

// Retry calls the call function for count times every interval while it returns true
// Call function should return (false, nil) to indicate a success
// (true, err) to indicate there was an error but we should retry
// and (false, err) to indicate there was a fatal error
func Retry(interval time.Duration, count int, call func() (bool, error)) error {
	for i := 0; i < count; i++ {
		shouldRetry, err := call()
		if !shouldRetry && err == nil {
			return nil
		} else if shouldRetry {
			fmt.Errorf("Failed with retry-able error: %s", err)
			time.Sleep(interval)
		} else {
			fmt.Errorf("Failed with an error and shouldn't retry: %s", err)
			return err
		}
	}
	return fmt.Errorf("Retry error: failed to satisfy the condition after %d times", count)
}
