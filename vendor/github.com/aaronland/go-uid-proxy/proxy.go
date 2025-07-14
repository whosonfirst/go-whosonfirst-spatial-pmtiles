package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aaronland/go-pool/v2"
	"github.com/aaronland/go-uid"
)

const PROXY_SCHEME string = "proxy"

func init() {
	ctx := context.Background()
	uid.RegisterProvider(ctx, PROXY_SCHEME, NewProxyProvider)
}

type ProxyProvider struct {
	uid.Provider
	provider  uid.Provider
	workers   int
	minimum   int
	pool      pool.Pool
	refilling *atomic.Bool
}

func NewProxyProvider(ctx context.Context, uri string) (uid.Provider, error) {

	workers := 10
	minimum := 0

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	status_monitor := false

	if q.Has("status-monitor") {

		v, err := strconv.ParseBool(q.Get("status-monitor"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?status-monitor= parameter, %w", err)
		}

		status_monitor = v
	}

	source_uri := q.Get("provider")

	if source_uri == "" {
		return nil, fmt.Errorf("Missing ?provider parameter")
	}

	pool_uri := q.Get("pool")

	if pool_uri == "" {
		pool_uri = "memory://"
	}

	str_workers := q.Get("workers")

	if str_workers != "" {

		v, err := strconv.Atoi(str_workers)

		if err != nil {
			return nil, fmt.Errorf("Invalid ?workers parameter")
		}

		workers = v
	}

	str_minimum := q.Get("minimum")

	if str_minimum != "" {

		v, err := strconv.Atoi(str_minimum)

		if err != nil {
			return nil, fmt.Errorf("Invalid ?minimum parameter")
		}

		minimum = v
	}

	source_pr, err := uid.NewProvider(ctx, source_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create provider, %w", err)
	}

	pl, err := pool.NewPool(ctx, pool_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create pool, %w", err)
	}

	pr := &ProxyProvider{
		provider:  source_pr,
		pool:      pl,
		workers:   workers,
		minimum:   minimum,
		refilling: new(atomic.Bool),
	}

	go pr.monitor(ctx)

	if status_monitor {
		slog.Debug("Starting status monitor")
		go pr.status(ctx)
	}

	return pr, nil
}

func (pr *ProxyProvider) UID(ctx context.Context, args ...interface{}) (uid.UID, error) {

	if pr.pool.Length(ctx) == 0 {

		slog.Debug("Pool length is 0 so fetching integer from source")

		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		done_ch := make(chan bool)

		go func() {
			pr.refillPool(ctx)
			done_ch <- true
		}()

		for {
			select {
			case <-done_ch:
				return pr.provider.UID(ctx, args...)
			case <-ticker.C:
				count := pr.pool.Length(ctx)
				if count > 0 {
					slog.Debug("Pool has count, try again", "count", count)
					return pr.provider.UID(ctx, args...)
				}
			}
		}
	}

	v, ok := pr.pool.Pop(ctx)

	if !ok {

		slog.Warn("Failed to pop UID from pool")

		done_ch := make(chan bool)

		go func() {
			pr.refillPool(ctx)
			done_ch <- true
		}()

		<-done_ch
		return pr.provider.UID(ctx, args...)
	}

	return v.(uid.UID), nil
}

func (pr *ProxyProvider) status(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			slog.Debug("Status monitor received done signal, exiting")
			return
		case <-time.After(5 * time.Second):
			slog.Debug("Status", "pool length", pr.pool.Length(ctx))
		}
	}
}

func (pr *ProxyProvider) monitor(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			if pr.pool.Length(ctx) < int64(pr.minimum) {
				go pr.refillPool(ctx)
			}
		}

	}
}

func (pr *ProxyProvider) refillPool(ctx context.Context) {

	if pr.refilling.Load() {
		return
	}

	pr.refilling.Swap(true)
	defer pr.refilling.Swap(false)

	if pr.minimum == 0 {
		pr.minimum = 1
	}

	t1 := time.Now()

	// Figure out how many integers we need to get *at this moment* which when
	// the service is under heavy load is a misleading number at best. It might
	// be worth adjusting this by a factor of (n) depending on the current load.
	// But that also means tracking what we think the current load means so we
	// aren't going to do that now...

	todo := int64(pr.minimum) - pr.pool.Length(ctx)

	workers := pr.workers

	if workers == 0 {
		workers = int(pr.minimum / 2)
	}

	if workers == 0 {
		workers = 1
	}

	slog.Debug("Start refilling pool.", "workers", workers, "to do", todo)

	// Now we're going to set up two simultaneous queues. One (the work group) is
	// just there to keep track of all the requests for new integers we need to
	// make. The second (the throttle) is there to make sure we don't exhaust all
	// the filehandles or network connections.

	th := make(chan bool, workers)

	for i := 0; i < workers; i++ {
		th <- true
	}

	wg := new(sync.WaitGroup)

	slog.Debug("Refill pool", "count", todo, "workers", workers)

	success := 0
	failed := 0

	for j := 0; int64(j) < todo; j++ {

		// Wait for the throttle to open a slot. Also record whether
		// the operation was successful.

		rsp := <-th

		if rsp == true {
			success += 1
		} else {
			failed += 1
		}

		// First check that we still actually need to keep fetching integers

		if pr.pool.Length(ctx) >= int64(pr.minimum) {
			slog.Debug("Pool is full", "count", pr.pool.Length(ctx), "iterations", j)
			break
		}

		// Standard work group stuff

		wg.Add(1)

		// Sudo make me a sandwitch. Note the part where we ping the throttle with
		// the return value at the end both to signal an available slot and to record
		// whether the integer harvesting was successful.

		go func(ctx context.Context, pr *ProxyProvider) {
			defer wg.Done()
			th <- pr.addToPool(ctx)
		}(ctx, pr)
	}

	// More standard work group stuff

	wg.Wait()

	t2 := time.Since(t1)
	slog.Debug("Pool refilled", "count", todo, "successful", success, "failed", failed, "total", pr.pool.Length(ctx), "time to complete", fmt.Sprintf("%v", t2))

}

func (pr *ProxyProvider) addToPool(ctx context.Context) bool {

	i, err := pr.provider.UID(ctx)

	if err != nil {
		slog.Error("Failed to create new UID to add to pool", "error", err)
		return false
	}

	pr.pool.Push(ctx, i)
	return true
}
