package ratelimit

import (
	"testing"
)

func TestRateLimiterWithIncorrectParams(t *testing.T) {

	t.Run("return error on negative max rpm", func(t *testing.T) {
		_, err := NewTokenBucketRateLimiter(-1, 5)
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
		expected := "maxRpm must be at least 1"
		if err.Error() != expected {
			t.Errorf("got message: %q. Wanted %q", err.Error(), expected)
		}
	})

	t.Run("return error on negative ramp up time", func(t *testing.T) {
		_, err := NewTokenBucketRateLimiter(100, -1)
		if err == nil {
			t.Errorf("Expected error but got nil")
		}
		expected := "ramp up minutes cannot be negative"
		if err.Error() != expected {
			t.Errorf("got message: %q. Wanted %q", err.Error(), expected)
		}
	})

	t.Run("rampUpDelta is set to 0 when rampUpMinutes is set to 0", func(t *testing.T) {

	})
}

func TestRampUpMinutesZero(t *testing.T) {
	maxRpm := 100
	rl, err := NewTokenBucketRateLimiter(maxRpm, 0)
	if err != nil {
		t.Errorf("Got unexpected error: %q", err)
	}

	if rl.rampingDelta != 0 {
		t.Errorf("Expected delta to be 0 in case of 0 minute ramp-up, but it was %d", rl.rampingDelta)
	}
	if rl.currentCapacity != maxRpm {
		t.Errorf("Expected initial capacity to be equal to maxRpm in case of 0 minute ramp-up, but it was %d", rl.currentCapacity)
	}
}

func TestRampUpDeltaAndInitialCapacityInitialization(t *testing.T) {
	maxRpm := 100
	rl, err := NewTokenBucketRateLimiter(maxRpm, 1)
	if err != nil {
		t.Errorf("Got unexpected error: %q", err)
	}

	if rl.rampingDelta != 16 { // =100/6, because the first 6 ramp up iterations happen within a minute and that should bring us to the maxRpm capacity
		t.Errorf("Expected delta to be 16 but it was %d", rl.rampingDelta)
	}
	if rl.currentCapacity != 10 {
		t.Errorf("Expected initial capacity to be equal to 10 (0.1 * maxRpm) but it was %d", rl.currentCapacity)
	}
}

func TestRateLimitingLogic(t *testing.T) {
	t.Run("test that rate limit is enforced", func(t *testing.T) {
		rl, _ := NewTokenBucketRateLimiter(20, 1)

		if rl.currentCapacity != 2 || rl.tokens != 2 {
			t.Errorf("Expected to start off with 2 tokens (10 percent of max), but currentCapacity=%d and tokens=%d", rl.currentCapacity, rl.tokens)
		}
		if rl.Accept() != true {
			t.Errorf("Consuming first token failed unexpectedly")
		}
		if rl.Accept() != true {
			t.Errorf("Consuming the second token failed unexpectedly")
		}
		if rl.Accept() != false {
			t.Errorf("Consuming third token was allowed unexpectedly")
		}
	})
}
