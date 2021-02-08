package datadog

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// Retry calls the call function for count times every interval while it returns nil
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
		} else if errors.Is(err, retryErrorType) {
			log.Printf(err.Error())
			time.Sleep(interval)
		} else if errors.Is(err, fatalErrorType) {
			log.Printf(err.Error())
			return err
		}
	}
	return &FatalError{prob: fmt.Sprintf("Retry error: failed to satisfy the condition after %d times\n", count)}
}

// RetryableError represents a transient error and means its safe to try the request again
type RetryableError struct {
	prob string
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("[INFO] Failed with retry-able error: %s", e.prob)
}

// FatalError should be considered final and not trigger any retries
type FatalError struct {
	prob string
}

func (e *FatalError) Error() string {
	return fmt.Sprintf("[ERROR] Failed with an error and shouldn't retry: %s", e.prob)
}
