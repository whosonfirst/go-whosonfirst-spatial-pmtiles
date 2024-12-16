package pmtiles

// Simple package to manage atomic counters in a map (dictionary) until something
// like is added to the default sync or sync/atomic packages.
// If anything begged to use generics it is this...

import (
	"sync"
)

type Counter struct {
	lookup map[string]int32
	mu     *sync.RWMutex
}

func NewCounter() *Counter {

	lookup := make(map[string]int32)
	mu := new(sync.RWMutex)

	c := &Counter{
		lookup: lookup,
		mu:     mu,
	}

	return c
}

func (c *Counter) Count(key string) int32 {

	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.lookup[key]

	if !ok {
		v = int32(0)
	}

	return v
}

func (c *Counter) Increment(key string, i int32) int32 {

	c.mu.Lock()
	defer c.mu.Unlock()

	v, ok := c.lookup[key]

	if !ok {
		v = int32(0)
	}

	new_v := v + i

	if new_v < 0 {
		delete(c.lookup, key)
		new_v = 0
	}

	c.lookup[key] = new_v
	return new_v
}
