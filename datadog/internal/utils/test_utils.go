package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

// Retry calls the call function for count times every interval while it returns a RetryableError
// Call function should return either: nil, RetryableError, or FatalError
// nil indicates a success condition was met
// RetryableError means we'll retry up to the count
// FatalError indicates we shouldn't try the request again
func Retry(interval time.Duration, count int, call func() error) error {
	var retryErrorType *RetryableError
	var fatalErrorType *FatalError
	for i := 0; i < count; i++ {
		err := call()
		if err == nil {
			return nil
		} else if errors.As(err, &retryErrorType) {
			log.Print(err.Error())
			if os.Getenv("RECORD") == "false" {
				// Skip sleep in replay mode to go faster
				continue
			}
			time.Sleep(interval)
		} else if errors.As(err, &fatalErrorType) {
			log.Print(err.Error())
			return err
		}
	}
	return &FatalError{Prob: fmt.Sprintf("failed to satisfy the condition after %d times", count)}
}

// RetryableError represents a transient error and means its safe to try the request again
type RetryableError struct {
	Prob string
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("[INFO] Failed with retry-able error: %s", e.Prob)
}

// FatalError should be considered final and not trigger any retries
type FatalError struct {
	Prob string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("[ERROR] Failed with an error and shouldn't retry: %s", e.Prob)
}
