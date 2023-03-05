package main

import (
	"time"

	"github.com/VladMinzatu/go-projects/rate-limiter/ratelimit"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)

	// create a token bucket with a peak capacity of 20 tokens per minute and a ramp-up period of 5 minutes
	rl, _ := ratelimit.NewTokenBucketRateLimiter(20, 5)

	// simulate incoming requests
	for i := 1; i <= 1000; i++ {
		rl.Accept()
		time.Sleep(500 * time.Millisecond)
	}
}
