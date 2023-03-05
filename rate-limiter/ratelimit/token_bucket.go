/*
A token-bucket rate limiter with ramp-up functionality.

Set a maximum number of requests per minute to be supported and the capacity will scale up and down with the demand,
smoothly over time, according to the ramp-up interval.
*/
package ratelimit

import (
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

var (
	rcurrentCapacityGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "current_capacity",
		Help: "The current capacity of the bucket measured in rpm",
	})
)

const (
	refillIntervalSeconds     = 10
	refillsPerMinute          = 60 / refillIntervalSeconds
	scaleUpThreshold          = 0.4 // available tokens vs capacity
	scaleDownThreshold        = 0.9 // available tokens vs capacity
	initialCapacityPercentage = 0.1 // start off with 10% of the maxRpm
)

type TokenBucketRateLimiter struct {
	maxRpm        int // peak number of requests per minute allowed e.g. 60 * 500rps = 30_000
	rampUpMinutes int // number of minutes over which to smoothly ramp up to the max rpm

	tokens          int // current number of tokens in the bucket
	currentCapacity int
	rampingDelta    int
	stop            chan bool
}

func NewTokenBucketRateLimiter(maxRpm, rampUpMinutes int) (*TokenBucketRateLimiter, error) {
	if maxRpm < 1 {
		return nil, errors.New("maxRpm must be at least 1")
	}
	if rampUpMinutes < 0 {
		return nil, errors.New("ramp up minutes cannot be negative")
	}
	var rampingDelta int
	var startCapacity int
	if rampUpMinutes > 0 {
		rampingDelta = max(maxRpm/(rampUpMinutes*refillsPerMinute), 1)
		startCapacity = max(int(float64(maxRpm)*initialCapacityPercentage), 1)
	} else {
		rampingDelta = 0
		startCapacity = maxRpm
	}
	return &TokenBucketRateLimiter{maxRpm: maxRpm, rampUpMinutes: rampUpMinutes, currentCapacity: startCapacity, tokens: startCapacity, rampingDelta: rampingDelta}, nil
}

func (rl *TokenBucketRateLimiter) Accept() bool {
	if rl.tokens > 0 {
		rl.tokens -= 1
		log.Debugf("Token retrieved from bucket. Tokens left: %d", rl.tokens)
		return true
	}
	log.Debug("No tokens available")
	return false
}

func (rl *TokenBucketRateLimiter) Start() {
	go rl.refill()
}

func (rl *TokenBucketRateLimiter) Stop() {
	rl.stop <- true
}

func (rl *TokenBucketRateLimiter) refill() {
	for {
		select {
		case <-rl.stop:
			return
		case <-time.After(refillIntervalSeconds * time.Second):
			rl.rampUp()
			tokensToAdd := max(rl.currentCapacity/refillsPerMinute, 1)
			rl.tokens = min(rl.tokens+tokensToAdd, rl.currentCapacity)
			log.Debugf("Adding %d tokens to bucket. New capacity: %d", tokensToAdd, rl.currentCapacity)
		}
	}
}

// Adjust the current capacity up or down depending on the rate of consumption of tokens in the bucket and the ramp-up rate configured
func (rl *TokenBucketRateLimiter) rampUp() {
	if float64(rl.tokens) > scaleDownThreshold*float64(rl.currentCapacity) {
		rl.currentCapacity = max(rl.currentCapacity-rl.rampingDelta, 1)
		log.Debugf("Scaled down capacity to %d", rl.currentCapacity)
	} else if float64(rl.tokens) < scaleUpThreshold*float64(rl.currentCapacity) {
		rl.currentCapacity = min(rl.currentCapacity+rl.rampingDelta, rl.maxRpm)
		log.Debugf("Scaled up capacity to %d", rl.currentCapacity)
	}
	rcurrentCapacityGauge.Set(float64(rl.currentCapacity))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
