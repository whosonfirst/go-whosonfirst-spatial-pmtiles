package timer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Timing struct {
	Created     time.Time
	Description string
	Duration    time.Duration
}

func (t *Timing) String() string {
	return fmt.Sprintf("%s: %v", t.Description, t.Duration)
}

type Timer struct {
	mu      *sync.RWMutex
	Timings map[string][]*Timing
}

func NewTimer() *Timer {

	timings := make(map[string][]*Timing)
	mu := new(sync.RWMutex)

	t := &Timer{
		mu:      mu,
		Timings: timings,
	}

	return t
}

func (t *Timer) Add(ctx context.Context, group string, description string, duration time.Duration) error {

	now := time.Now()

	tm := &Timing{
		Created:     now,
		Description: description,
		Duration:    duration,
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	timings, ok := t.Timings[group]

	if !ok {
		timings = make([]*Timing, 0)
	}

	timings = append(timings, tm)
	t.Timings[group] = timings

	return nil
}
