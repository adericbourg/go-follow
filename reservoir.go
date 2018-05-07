package main

import (
	"time"
	"sync"
)

const ReservoirTrimThreshold = 10

type Reservoir struct {
	Measurements map[time.Time]int64
	Count        int
	Mutex        *sync.Mutex
}

func CreateReservoir() *Reservoir {
	return &Reservoir{
		Measurements: make(map[time.Time]int64),
		Mutex:        &sync.Mutex{},
	}
}

func increment(reservoir *Reservoir) {
	reservoir.Count += 1
	if reservoir.Count%ReservoirTrimThreshold == 0 {
		trim(reservoir)
	}

	key := time.Now()
	Synchronized(reservoir, func() {
		value, exists := reservoir.Measurements[key]
		if exists {
			reservoir.Measurements[key] = value + 1
		} else {
			reservoir.Measurements[key] = 1
		}
	})
}

func trim(reservoir *Reservoir) *Reservoir {
	Synchronized(reservoir, func() {
		var keys []time.Time

		for k := range reservoir.Measurements {
			keys = append(keys, k)
		}
		for _, key := range keys {
			if time.Since(key) > time.Hour {
				delete(reservoir.Measurements, key)
			}
		}
	})
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

func Synchronized(reservoir *Reservoir, f func()) {
	reservoir.Mutex.Lock()
	f()
	reservoir.Mutex.Unlock()
}
