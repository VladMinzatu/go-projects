package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// Just a demo of how to use the SLOAlert:
	sloAlert, err := NewSLOAlertFromPercentageUsed(0.99, 1*time.Hour, 0.02) // alerting on 2% error budget used in the past hour (for our 99% SLO)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v:\n", sloAlert)

	errorRate := 1.0
	fmt.Printf("We start seeing Error Rate: %.2f%%\n", errorRate*100)
	sloCheck, err := sloAlert.Check(errorRate)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("  Alert Will Trigger: %t\n", sloCheck)
	if sloCheck {
		detectionTime, _ := sloAlert.DetectionTime(errorRate)
		fmt.Printf("  Detection Time: %s\n", detectionTime)
		errorBudgetConsumed := sloAlert.ErrorBudgetConsumedBeforeTriggering()
		fmt.Printf("  Error Budget Consumed Before Triggering: %.2f%% \n", errorBudgetConsumed*100)
	}
}
