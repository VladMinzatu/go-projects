package main

import (
	"math"
	"testing"
	"time"
)

func TestCreatingAlertFromBurnRate(t *testing.T) {
	tests := []struct {
		slo                           float64
		alertWindowSize               time.Duration
		burnRate                      float64
		expectedError                 error
		expectedPercentBudgetConsumed float64
	}{
		{0.99, 1 * time.Hour, 2.0, nil, 0.002976},
		{0.99, 1 * time.Hour, 5.0, nil, 0.007440},
		{-1.0, 1 * time.Hour, 2.0, ErrSLOOutOfRange, 0.0},
		{1.1, 1 * time.Hour, 2.0, ErrSLOOutOfRange, 0.0},
		{0.99, 1 * time.Minute, 2.0, ErrAlertTimeWindowOutOfRange, 0.0},
		{0.99, 25 * time.Hour, 2.0, ErrAlertTimeWindowOutOfRange, 0.0},
		{0.99, 1 * time.Hour, -1.0, ErrBurnRateOutOfRange, 0.0},
		{0.99, 1 * time.Hour, 101.0, ErrBurnRateOutOfRange, 0.0},
	}
	for _, test := range tests {
		alert, err := NewSLOAlertFromBurnRate(test.slo, test.alertWindowSize, test.burnRate)
		if err != test.expectedError {
			t.Errorf("NewSLOAlertFromBurnRate(%f, %s, %f) returned error: %v", test.slo, test.alertWindowSize, test.burnRate, err)
		}
		if err == nil && math.Abs(alert.PercentErrorBudgetConsumed-test.expectedPercentBudgetConsumed) > 1e-6 {
			t.Errorf("NewSLOAlertFromBurnRate(%f, %s, %f) should have had consumed error %f but was %f",
				test.slo, test.alertWindowSize, test.burnRate, test.expectedPercentBudgetConsumed, alert.PercentErrorBudgetConsumed)
		}
	}
}

func TestCreatingAlertFromrBudgetUsed(t *testing.T) {
	tests := []struct {
		slo              float64
		alertWindowSize  time.Duration
		errorBudgetUsed  float64
		expectedError    error
		expectedBurnRate float64
	}{
		{0.99, 1 * time.Hour, 0.03, nil, 20.16},
		{0.99, 1 * time.Hour, 0.1, nil, 67.2},
		{-1.0, 1 * time.Hour, 0.03, ErrSLOOutOfRange, 0.0},
		{1.1, 1 * time.Hour, 0.03, ErrSLOOutOfRange, 0.0},
		{0.99, 1 * time.Minute, 0.03, ErrAlertTimeWindowOutOfRange, 0.0},
		{0.99, 25 * time.Hour, 0.03, ErrAlertTimeWindowOutOfRange, 0.0},
		{0.99, 1 * time.Hour, -0.01, ErrErrorBudgetUsedOutOfRange, 0.0},
		{0.99, 1 * time.Hour, 1.1, ErrErrorBudgetUsedOutOfRange, 0.0},
		{0.99, 1 * time.Hour, 0.001, ErrBurnRateOutOfRange, 0.0},
		{0.99, 1 * time.Hour, 0.90, ErrBurnRateOutOfRange, 0.0},
	}
	for _, test := range tests {
		alert, err := NewSLOAlertFromBudgetUsed(test.slo, test.alertWindowSize, test.errorBudgetUsed)
		if err != test.expectedError {
			t.Errorf("NewSLOAlertFromPercentageUsed(%f, %s, %f) returned error: %v", test.slo, test.alertWindowSize, test.errorBudgetUsed, err)
		}
		if err == nil && math.Abs(alert.BurnRate-test.expectedBurnRate) > 1e-6 {
			t.Errorf("NewSLOAlertFromBudgetUsed(%f, %s, %f) should have had burn rate %f but was %f",
				test.slo, test.alertWindowSize, test.errorBudgetUsed, test.expectedBurnRate, alert.BurnRate)
		}
	}
}

func TestCreatingNewScenario(t *testing.T) {
	alert, _ := NewSLOAlertFromBurnRate(0.99, 1*time.Hour, 2.0)
	tests := []struct {
		errorRate     float64
		expectedError error
	}{
		{0.01, nil},
		{0.03, nil},
		{-0.01, ErrErrorRateOutOfRange},
		{1.01, ErrErrorRateOutOfRange},
	}
	for _, test := range tests {
		_, err := NewScenario(alert, test.errorRate)
		if err != test.expectedError {
			t.Errorf("NewScenario(%f) returned error: %v", test.errorRate, err)
		}
	}
}

func TestAlertCondition(t *testing.T) {
	alert, _ := NewSLOAlertFromBurnRate(0.99, 1*time.Hour, 2.0)
	if scenario, _ := NewScenario(alert, 0.01); scenario.Check() {
		t.Errorf("Alert triggered when it should not have (error rate: 1%%)")
	}
	if scenario, _ := NewScenario(alert, 0.03); !scenario.Check() {
		t.Errorf("Alert failed to trigger when it should have (error rate: 3%%)")
	}
}

func TestDetectionTime(t *testing.T) {
	alert, _ := NewSLOAlertFromBurnRate(0.99, 1*time.Hour, 2.0)
	if scenario, _ := NewScenario(alert, 1.0); scenario.DetectionTime() != 1*time.Minute+12*time.Second {
		t.Errorf("Scenario.DetectionTime() not as expected (1m12s)")
	}
	if scenario, _ := NewScenario(alert, 0.5); scenario.DetectionTime() != 2*time.Minute+24*time.Second {
		t.Errorf("Scenario.DetectionTime() not as expected (2m24s)")
	}
	if scenario, _ := NewScenario(alert, 0.01); scenario.DetectionTime() != -1 {
		t.Errorf("Scenario.DetectionTime() did not return -1 when alert was not triggered")
	}
}
