package services

import "time"

// MediaRetryDelay returns a bounded exponential delay for retry scheduling.
// Attempts are expected to start at 1 after the first processing claim.
func MediaRetryDelay(attempts int) time.Duration {
	if attempts <= 0 {
		return time.Second
	}

	const maxDelay = 30 * time.Minute
	delay := time.Second * time.Duration(1<<(attempts-1))
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

// MediaJobRetryable determines whether a failed attempt may return to the queue.
func MediaJobRetryable(attempts int) bool {
	return attempts < MaxMediaAttempts
}
