package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_CreateReservoir_should_create_an_empty_reservoir(t *testing.T) {
	reservoir := CreateReservoir()

	assert.NotNil(t, reservoir)
	assert.Equal(t, 0, reservoir.Count)
	assert.Equal(t, 0, len(reservoir.Measurements))
}

func Test_Increment_should_add_a_measurement(t *testing.T) {
	reservoir := CreateReservoir()

	increment(reservoir)

	assert.Equal(t, 1, len(reservoir.Measurements))
	for key := range reservoir.Measurements {
		assert.Equal(t, int64(1), reservoir.Measurements[key])
	}
}

func Test_GetMeasurements_should_provide_map_values(t *testing.T) {
	reservoir := CreateReservoir()

	increment(reservoir)
	increment(reservoir)

	measurements := GetMeasurements(reservoir)
	assert.Equal(t, []int64{1, 1}, measurements)
}

func Test_Sum_should_sum_all_values(t *testing.T) {
	reservoir := CreateReservoir()

	increment(reservoir)
	increment(reservoir)
	increment(reservoir)

	sum := Sum(reservoir)
	assert.Equal(t, int64(3), sum)
}
