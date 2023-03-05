package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/VladMinzatu/go-projects/rate-limiter/ratelimit"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsAcceptedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "requests_accepted",
		Help: "Requests that the RateLimiter allowed to pass through",
	})
	requestsRejectedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "requests_rejected",
		Help: "Requests that the RateLimiter blocked from passing through",
	})
)

func makeRequestsAtConstantRpm(rpm int, rl *ratelimit.TokenBucketRateLimiter) {
	go func() {
		startTime := time.Now()
		for {
			if rl.Accept() {
				requestsAcceptedCounter.Inc()
			} else {
				requestsRejectedCounter.Inc()
			}
			time.Sleep(time.Minute / time.Duration(rpm))
			if time.Since(startTime) > 10*time.Minute {
				time.Sleep(10 * time.Minute)
				startTime = time.Now()
			}
		}
	}()
}

func main() {
	rl, err := ratelimit.NewTokenBucketRateLimiter(100, 5)
	if err != nil {
		fmt.Printf("Error initializing rate limiter: %s\n", err.Error())
		os.Exit(1)
	}
	rl.Start()

	go makeRequestsAtConstantRpm(90, rl)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
