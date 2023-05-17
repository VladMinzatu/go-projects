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

func TestAlertCondition(t *testing.T) {
	alert, _ := NewSLOAlert(0.99, 1*time.Hour, 2.0)
	if check, _ := alert.Check(0.01); check {
		t.Errorf("Alert triggered when it should not have (error rate: 1%%)")
	}
	if check, _ := alert.Check(0.03); !check {
		t.Errorf("Alert failed to trigger when it should have (error rate: 3%%)")
	}
	if _, err := alert.Check(-0.01); err != ErrErrorRateOutOfRange {
		t.Errorf("Alert.Check(-0.01) returned error: %v", err)
	}
	if _, err := alert.Check(1.01); err != ErrErrorRateOutOfRange {
		t.Errorf("Alert.Check(1.01) returned error: %v", err)
	}
}

func TestDetectionTime(t *testing.T) {
	alert, _ := NewSLOAlert(0.99, 1*time.Hour, 2.0)
	if _, err := alert.DetectionTime(-0.01); err != ErrErrorRateOutOfRange {
		t.Errorf("Alert.DetectionTime(0.01) returned error: %v", err)
	}
	if _, err := alert.DetectionTime(1.01); err != ErrErrorRateOutOfRange {
		t.Errorf("Alert.DetectionTime(1.01) returned error: %v", err)
	}
	if detectionTime, _ := alert.DetectionTime(1.0); detectionTime != 1*time.Minute+12*time.Second {
		t.Errorf("Alert.DetectionTime(1.0) returned %s, expected 1m12s", detectionTime)
	}
	if detectionTime, _ := alert.DetectionTime(0.5); detectionTime != 2*time.Minute+24*time.Second {
		t.Errorf("Alert.DetectionTime(1.0) returned %s, expected 2m24s", detectionTime)
	}
}

func TestErrorBudgetConsumedBeforeTriggering(t *testing.T) {
	errorBudget := 0.03
	alert, _ := NewSLOAlertFromPercentageUsed(0.99, 1*time.Hour, errorBudget)
	if consumed := alert.ErrorBudgetConsumedBeforeTriggering(); math.Abs(consumed-errorBudget) > 1e-9 {
		t.Errorf("Alert.ErrorBudgetConsumedBeforeTriggering returned %f, expected %f", consumed, errorBudget)
	}
}
