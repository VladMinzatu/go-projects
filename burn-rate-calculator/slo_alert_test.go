package main

import (
	"math"
	"testing"
	"time"
)

func TestCreatingAlertFromBurnRate(t *testing.T) {
	tests := []struct {
		slo             float64
		alertWindowSize time.Duration
		burnRate        float64
		expectedError   error
	}{
		{0.99, 1 * time.Hour, 2.0, nil},
		{-1.0, 1 * time.Hour, 2.0, ErrSLOOutOfRange},
		{1.1, 1 * time.Hour, 2.0, ErrSLOOutOfRange},
		{0.99, 1 * time.Minute, 2.0, ErrAlertTimeWindowOutOfRange},
		{0.99, 25 * time.Hour, 2.0, ErrAlertTimeWindowOutOfRange},
		{0.99, 1 * time.Hour, -1.0, ErrBurnRateOutOfRange},
		{0.99, 1 * time.Hour, 101.0, ErrBurnRateOutOfRange},
	}
	for _, test := range tests {
		_, err := NewSLOAlert(test.slo, test.alertWindowSize, test.burnRate)
		if err != test.expectedError {
			t.Errorf("NewSLOAlert(%f, %s, %f) returned error: %v", test.slo, test.alertWindowSize, test.burnRate, err)
		}
	}
}

func TestCreatingAlertFromErrorBusgetUsed(t *testing.T) {
	tests := []struct {
		slo             float64
		alertWindowSize time.Duration
		errorBudgetUsed float64
		expectedError   error
	}{
		{0.99, 1 * time.Hour, 0.03, nil},
		{-1.0, 1 * time.Hour, 0.03, ErrSLOOutOfRange},
		{1.1, 1 * time.Hour, 0.03, ErrSLOOutOfRange},
		{0.99, 1 * time.Minute, 0.03, ErrAlertTimeWindowOutOfRange},
		{0.99, 25 * time.Hour, 0.03, ErrAlertTimeWindowOutOfRange},
		{0.99, 1 * time.Hour, -0.01, ErrErrorBudgetUsedOutOfRange},
		{0.99, 1 * time.Hour, 1.1, ErrErrorBudgetUsedOutOfRange},
		{0.99, 1 * time.Hour, 0.001, ErrBurnRateOutOfRange},
		{0.99, 1 * time.Hour, 0.90, ErrBurnRateOutOfRange},
	}
	for _, test := range tests {
		_, err := NewSLOAlertFromPercentageUsed(test.slo, test.alertWindowSize, test.errorBudgetUsed)
		if err != test.expectedError {
			t.Errorf("NewSLOAlertFromPercentageUsed(%f, %s, %f) returned error: %v", test.slo, test.alertWindowSize, test.errorBudgetUsed, err)
		}
	}
}

func TestCreatingNewScenario(t *testing.T) {
	alert, _ := NewSLOAlert(0.99, 1*time.Hour, 2.0)
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
	alert, _ := NewSLOAlert(0.99, 1*time.Hour, 2.0)
	if scenario, _ := NewScenario(alert, 0.01); scenario.Check() {
		t.Errorf("Alert triggered when it should not have (error rate: 1%%)")
	}
	if scenario, _ := NewScenario(alert, 0.03); !scenario.Check() {
		t.Errorf("Alert failed to trigger when it should have (error rate: 3%%)")
	}
}

func TestDetectionTime(t *testing.T) {
	alert, _ := NewSLOAlert(0.99, 1*time.Hour, 2.0)
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

func TestErrorBudgetConsumedBeforeTriggering(t *testing.T) {
	errorBudget := 0.03
	alert, _ := NewSLOAlertFromPercentageUsed(0.99, 1*time.Hour, errorBudget)
	if consumed := alert.ErrorBudgetConsumedBeforeTriggering(); math.Abs(consumed-errorBudget) > 1e-9 {
		t.Errorf("Scenario.ErrorBudgetConsumedBeforeTriggering returned %f, expected %f", consumed, errorBudget)
	}
}
