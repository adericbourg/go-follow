package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"sync"
)

func Test_CreateReservoir_should_create_an_empty_reservoir(t *testing.T) {
	reservoir := CreateDefaultReservoir()

	assert.NotNil(t, reservoir)
	assert.Equal(t, 0, reservoir.Count)
	assert.Equal(t, 0, len(reservoir.Measurements))
}

func Test_Increment_should_add_a_measurement(t *testing.T) {
	reservoir := CreateDefaultReservoir()

	reservoir.Increment()

	assert.Equal(t, 1, len(reservoir.Measurements))
	for key := range reservoir.Measurements {
		assert.Equal(t, int64(1), reservoir.Measurements[key])
	}
}

func Test_GetMeasurements_should_provide_map_values(t *testing.T) {
	reservoir := CreateDefaultReservoir()

	reservoir.Increment()
	reservoir.Increment()

	measurements := reservoir.GetMeasurements()
	assert.Equal(t, []int64{1, 1}, measurements)
}

func Test_Sum_should_sum_all_values(t *testing.T) {
	reservoir := CreateDefaultReservoir()

	reservoir.Increment()
	reservoir.Increment()
	reservoir.Increment()

	sum := reservoir.Sum()
	assert.Equal(t, int64(3), sum)
}

func Test_concurrency(t *testing.T) {
	const iterations = 1000
	reservoir := CreateDefaultReservoir()

	wg := sync.WaitGroup{}
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func() {
			reservoir.Increment()
			defer wg.Done()
		}()
	}
	wg.Wait()

	sum := reservoir.Sum()
	assert.Equal(t, int64(iterations), sum)
}
