package main

import (
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
	"sync"
)

func Test_Submit_should_run_function(t *testing.T) {
	rateLimiter := CreateRateLimiter(1000, time.Hour)

	var triggered = false
	rateLimiter.Submit(func() {
		triggered = true
	})

	assert.Equal(t, true, triggered)
}

func Test_Submit_should_run_function_until_threshold_is_reached(t *testing.T) {
	threshold := int64(3)
	rateLimiter := CreateRateLimiter(threshold, time.Hour)

	var triggerCount = 0
	for i := int64(0); i < threshold+1; i++ {
		rateLimiter.Submit(func() {
			triggerCount += 1
		})
	}
	assert.Equal(t, int(threshold), triggerCount)
}

func Test_concurrent_calls(t *testing.T) {
	const iterations = 2000
	threshold := int64(1000)
	rateLimiter := CreateRateLimiter(threshold, time.Hour)

	wg := sync.WaitGroup{}
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			rateLimiter.Submit(func() {
				// Whatever
			})
			defer wg.Done()
		}()
	}
	wg.Wait()

	assert.Equal(t, threshold, rateLimiter.reservoir.Sum())
}