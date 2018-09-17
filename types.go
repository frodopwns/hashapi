package main

import (
	"fmt"
	"sync"
)

// Env wraps the shared items each handler may need
type Env struct {
	HashMap     *HashMap
	Stats       *Stats
	Terminating bool
}

// HashMap is a threadsafe map implementation for storing hashes
type HashMap struct {
	sync.RWMutex
	data map[int]string
}

// NewHashMap returns a new HashMap
func NewHashMap() *HashMap {
	return &HashMap{
		data: make(map[int]string),
	}
}

// Get returns an item from the map
func (hm *HashMap) Get(key int) (value string, ok bool) {
	hm.RLock()
	result, ok := hm.data[key-1]
	hm.RUnlock()
	return result, ok
}

// Save adds a new key to the map and returns its index
func (hm *HashMap) Save(value string) int {
	hm.Lock()
	// auto incrementing key
	key := len(hm.data)
	hm.data[key] = value
	hm.Unlock()

	// first index is 0 but first allowed id is 1
	return key + 1
}

// Update changes the values of an existing index
func (hm *HashMap) Update(key int, value string) {
	hm.Lock()
	hm.data[key-1] = value
	hm.Unlock()
}

// Stats is a threadsafe struct for storing statistics
type Stats struct {
	sync.Mutex
	requests    int     // total number of POST requests
	requestTime float64 // total time of all POST requests
}

// NewStats returns a new Stats struct
func NewStats() *Stats {
	return &Stats{}
}

// Update adds the request time to the total request time and increments the total requests
func (s *Stats) Update(incr float64) {
	s.Lock()
	s.requests++
	s.requestTime = s.requestTime + incr
	s.Unlock()
}

// JSON calculates average request time and renders stats in json
func (s *Stats) JSON() string {
	return fmt.Sprintf(
		`{"total": %d, "average": %f}`,
		s.requests,
		s.requestTime/float64(s.requests),
	)
}
