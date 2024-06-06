package sse

import (
	"context"
	"crypto/tls"
	"gopkg.in/cenkalti/backoff.v1"
	"net/http"
	"time"
)

func WithReconnectNotify(events chan *Event, cancelCtx context.CancelFunc) func(c *Client) {
	return func(c *Client) {
		c.ReconnectNotify = func(err error, duration time.Duration) {
			// on reconnect backoff failed
			if duration == backoff.Stop {
				c.Unsubscribe(events)
				cancelCtx()
			}
		}
	}
}

func WithInsecureTransport(client *Client) {
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

func WithReconnectStrategy(client *Client) {
	// will try to reconnect every 100ms after connection closed, and stop after 3 retry
	client.ReconnectStrategy = NewThrottledBackoff(100*time.Millisecond, 2)
}

// ThrottledBackOff implement backoff interface. Will retry definitively unless time between interrupt is lower than interval for n time.
type ThrottledBackOff struct {
	interval      time.Duration
	retry         int
	maxRetry      int
	lastInterrupt time.Time
}

func (t *ThrottledBackOff) NextBackOff() time.Duration {
	now := time.Now()
	if now.Sub(t.lastInterrupt) >= t.interval {
		t.lastInterrupt = now
		t.retry = 0
		return t.interval
	}

	t.retry++
	if t.retry >= t.maxRetry {
		return backoff.Stop
	}

	t.lastInterrupt = now
	return t.interval
}

func (t *ThrottledBackOff) Reset() {
	t.lastInterrupt = time.Now()
	t.retry = 0
}

func NewThrottledBackoff(interval time.Duration, retry int) *ThrottledBackOff {
	return &ThrottledBackOff{
		interval:      interval,
		retry:         0,
		maxRetry:      retry,
		lastInterrupt: time.Now(),
	}
}
