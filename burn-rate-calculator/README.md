# burn-rate-calculator

Small project to demo how alerting based on SLO error budget burn rate can be configured and how it would behave, following the guidelines in https://sre.google/workbook/alerting-on-slos/.

For your given SLO over a 28 day period (e.g. 99% success rate), you can define the alert as "Alert me when X% of the error budget is consumed in Y time".
Then, for a given error rate that you might start observing you can query:
- whether the alert will be triggered
- how long it will take for the alert to trigger once the errors start happening at the given rate
- what percentage of the error budget will have been consumed by the time the alert is triggered (spoiler! it's the percentage you set to alert on, but you also have the option to define your alert in terms of burn rate and then query this percentage.)

This can be used to understand the behavior of the alerting and tune the parameters to your expectations before putting the alert live.