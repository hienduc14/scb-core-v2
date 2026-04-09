package entities

import "time"

type CleanupPolicy struct {
	CompletedRetention time.Duration
	FailedRetention    time.Duration
	CheckInterval      time.Duration
}
