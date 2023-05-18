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
	SLO                        float64
	AlertWindowSize            time.Duration
	BurnRate                   float64
	PercentErrorBudgetConsumed float64
}

// A scenario models how an alert behaves when a certain error rate starts being observed in the system
type Scenario struct {
	Alert     *SLOAlert
	ErrorRate float64
}

func NewSLOAlertFromBurnRate(slo float64, alertWindowSize time.Duration, burnRate float64) (*SLOAlert, error) {
	err := verifyAlertConfiguration(slo, alertWindowSize, burnRate)
	if err != nil {
		return nil, err
	}

	percentErrorBudgetConsumed := burnRate * float64(alertWindowSize) / float64(SLOWindowSize)
	return &SLOAlert{
		SLO:                        slo,
		AlertWindowSize:            alertWindowSize,
		BurnRate:                   burnRate,
		PercentErrorBudgetConsumed: percentErrorBudgetConsumed,
	}, nil
}

func NewSLOAlertFromBudgetUsed(slo float64, alertWindowSize time.Duration, percentageErrorBudgetUsed float64) (*SLOAlert, error) {
	if percentageErrorBudgetUsed < MinErrorBudgetUsed || percentageErrorBudgetUsed > MaxErrorBudgetUsed {
		return nil, ErrErrorBudgetUsedOutOfRange
	}

	burnRate := percentageErrorBudgetUsed * float64(SLOWindowSize) / float64(alertWindowSize)
	err := verifyAlertConfiguration(slo, alertWindowSize, burnRate)
	if err != nil {
		return nil, err
	}
	return NewSLOAlertFromBurnRate(slo, alertWindowSize, burnRate)
}

func verifyAlertConfiguration(slo float64, alertWindowSize time.Duration, burnRate float64) error {
	if slo < MinSLO || slo > MaxSLO {
		return ErrSLOOutOfRange
	}
	if alertWindowSize < MinAlertTimeWindow || alertWindowSize > MaxAlertTimeWindow {
		return ErrAlertTimeWindowOutOfRange
	}
	if burnRate < MinBurnRate || burnRate > MaxBurnRate {
		return ErrBurnRateOutOfRange
	}
	return nil
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
	if !s.Check() { // equivalent to duration > AlertWindowSize (easily provable by substituting in equations)
		return -1
	}
	duration := (1.0 - s.Alert.SLO) / s.ErrorRate * float64(s.Alert.AlertWindowSize) * float64(s.Alert.BurnRate)
	return time.Duration(duration)
}
