package main

import "time"

const ReservoirTrimThreshold = 10

type Reservoir struct {
	Measurements map[time.Time]int64
	Count        int
}

func CreateReservoir() *Reservoir {
	return &Reservoir{
		Measurements: make(map[time.Time]int64),
	}
}

func increment(reservoir *Reservoir) {
	reservoir.Count += 1
	if reservoir.Count%ReservoirTrimThreshold == 0 {
		trim(reservoir)
	}

	key := time.Now()
	value, exists := reservoir.Measurements[key]
	if exists {
		reservoir.Measurements[key] = value + 1
	} else {
		reservoir.Measurements[key] = 1
	}
}

func trim(reservoir *Reservoir) *Reservoir {
	var keys []time.Time

	for k := range reservoir.Measurements {
		keys = append(keys, k)
	}
	for _, key := range keys {
		if time.Since(key) > time.Hour {
			delete(reservoir.Measurements, key)
		}
	}

	return reservoir
}

func GetMeasurements(reservoir *Reservoir) []int64 {
	var values []int64
	measurements := trim(reservoir).Measurements

	for key := range measurements {
		values = append(values, measurements[key])
	}

	return values
}

func Sum(reservoir *Reservoir) int64 {
	sum := int64(0)
	for _, v := range GetMeasurements(reservoir) {
		sum += v
	}
	return sum
}
