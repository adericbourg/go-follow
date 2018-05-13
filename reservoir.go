package main

import (
	"time"
	"sync"
)

const ReservoirTrimThreshold = 10

type Reservoir struct {
	Measurements map[time.Time]int64
	Count        int
	TimeWindow   time.Duration
	Mutex        *sync.Mutex
}

func CreateDefaultReservoir() *Reservoir {
	return CreateReservoir(time.Hour)
}

func CreateReservoir(timeWindow time.Duration) *Reservoir {
	return &Reservoir{
		Measurements: make(map[time.Time]int64),
		TimeWindow:   timeWindow,
		Mutex:        &sync.Mutex{},
	}
}

func (reservoir *Reservoir) Increment() {
	reservoir.Count += 1
	if reservoir.Count%ReservoirTrimThreshold == 0 {
		reservoir.trim()
	}

	key := time.Now()
	reservoir.synchronized(func() {
		value, exists := reservoir.Measurements[key]
		if exists {
			reservoir.Measurements[key] = value + 1
		} else {
			reservoir.Measurements[key] = 1
		}
	})
}

func (reservoir *Reservoir) trim() {
	reservoir.synchronized(func() {
		var keys []time.Time

		for k := range reservoir.Measurements {
			keys = append(keys, k)
		}
		for _, key := range keys {
			if time.Since(key) > reservoir.TimeWindow {
				delete(reservoir.Measurements, key)
			}
		}
	})
}

func (reservoir *Reservoir) GetMeasurements() []int64 {
	var values []int64
	reservoir.trim()
	measurements := reservoir.Measurements

	for key := range measurements {
		values = append(values, measurements[key])
	}

	return values
}

func (reservoir *Reservoir) Sum() int64 {
	sum := int64(0)
	for _, v := range reservoir.GetMeasurements() {
		sum += v
	}
	return sum
}

func (reservoir *Reservoir) synchronized(f func()) {
	reservoir.Mutex.Lock()
	f()
	reservoir.Mutex.Unlock()
}
