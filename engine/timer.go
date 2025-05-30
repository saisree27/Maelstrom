package engine

import "time"

type Timer struct {
	startTime time.Time
	timeLimit time.Duration

	stopSearch bool
}
