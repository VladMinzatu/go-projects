package main

import (
	"fmt"
	"time"
)

// We're assuming an SLO window size of 28 days below
const SLOWindowSize = 28 * 24 * time.Hour

const MinSLO = 0.0
const MaxSLO = 1.0
const MinAlertTimeWindow = 10 * time.Minute
const MaxAlertTimeWindow = 24 * time.Hour
const MinBurnRate = 1.0
const MaxBurnRate = 100.0
const MinErrorBudgetUsed = 0.0
const MaxErrorBudgetUsed = 1.0
const MinErrorRate = 0.0
const MaxErrorRate = 1.0

var ErrSLOOutOfRange = fmt.Errorf("SLO must be between %f and %f", MinSLO, MaxSLO)
var ErrAlertTimeWindowOutOfRange = fmt.Errorf("alertWindowSize must be between %v and %v", MinAlertTimeWindow, MaxAlertTimeWindow)
var ErrBurnRateOutOfRange = fmt.Errorf("burnRate must be between %f and %f", MinBurnRate, MaxBurnRate)
var ErrErrorBudgetUsedOutOfRange = fmt.Errorf("errorBudgetUsed must be between %f and %f", MinErrorBudgetUsed, MaxErrorBudgetUsed)
var ErrErrorRateOutOfRange = fmt.Errorf("errorRate must be between %f and %f", MinErrorRate, MaxErrorRate)

type SLOAlert struct {
	SLO             float64
	AlertWindowSize time.Duration
	BurnRate        float64
}

// A scenario models the alert behavior in a given circumstance, described by observing a fixed error rate in the system
type Scenario struct {
	Alert     *SLOAlert
	ErrorRate float64
}

func NewSLOAlert(slo float64, alertWindowSize time.Duration, burnRate float64) (*SLOAlert, error) {
	if slo < MinSLO || slo > MaxSLO {
		return nil, ErrSLOOutOfRange
	}
	if alertWindowSize < MinAlertTimeWindow || alertWindowSize > MaxAlertTimeWindow {
		return nil, ErrAlertTimeWindowOutOfRange
	}
	if burnRate < MinBurnRate || burnRate > MaxBurnRate {
		return nil, ErrBurnRateOutOfRange
	}
	return &SLOAlert{
		SLO:             slo,
		AlertWindowSize: alertWindowSize,
		BurnRate:        burnRate,
	}, nil
}

func NewSLOAlertFromPercentageUsed(slo float64, alertWindowSize time.Duration, errorBudgetUsed float64) (*SLOAlert, error) {
	if errorBudgetUsed < MinErrorBudgetUsed || errorBudgetUsed > MaxErrorBudgetUsed {
		return nil, ErrErrorBudgetUsedOutOfRange
	}
	burnRate := errorBudgetUsed * float64(SLOWindowSize) / float64(alertWindowSize)
	return NewSLOAlert(slo, alertWindowSize, burnRate)
}

func (a *SLOAlert) ErrorBudgetConsumedBeforeTriggering() float64 {
	return a.BurnRate * float64(a.AlertWindowSize) / float64(SLOWindowSize)
}

func NewScenario(alert *SLOAlert, errorRate float64) (*Scenario, error) {
	if errorRate < MinErrorRate || errorRate > MaxErrorRate {
		return nil, ErrErrorRateOutOfRange
	}
	return &Scenario{
		Alert:     alert,
		ErrorRate: errorRate,
	}, nil
}

func (s *Scenario) Check() bool {
	errorBudgetPercentage := 1.0 - s.Alert.SLO
	return s.ErrorRate > s.Alert.BurnRate*errorBudgetPercentage
}

func (s *Scenario) DetectionTime() time.Duration {
	if !s.Check() {
		return -1
	}
	duration := (1.0 - s.Alert.SLO) / s.ErrorRate * float64(s.Alert.AlertWindowSize) * float64(s.Alert.BurnRate)
	return time.Duration(duration)
}
