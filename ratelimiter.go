package main

import (
	"time"
	"sync"
)

type RateLimiter struct {
	reservoir      *Reservoir
	threshold      int64
	thresholdMutex *sync.Mutex
}

func CreateRateLimiter(threshold int64, timeWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		reservoir:      CreateReservoir(timeWindow),
		threshold:      threshold,
		thresholdMutex: &sync.Mutex{},
	}
}

func (rateLimiter *RateLimiter) Submit(function func()) {
	rateLimiter.thresholdMutex.Lock()
	execute := rateLimiter.reservoir.Sum() < rateLimiter.threshold
	rateLimiter.thresholdMutex.Unlock()

	if execute {
		rateLimiter.reservoir.Increment()
		function()
	}
}
