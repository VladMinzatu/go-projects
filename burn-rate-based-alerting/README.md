# SLO Burn Rate Based Alerting

This is a succinct explanation of alerting based on SLO burn rates. This content is based on the information inside the ["Alerting on SLOs" chapter in the SRE workbook](https://sre.google/workbook/alerting-on-slos/).

### Setting the scene

*I will assume the reader knows what SLOs and error budgets are. If you need a detailed introduction to the topic, [Google's SRE Books](https://sre.google/books/) are the go-to reference, of course.*

We know that alerting should ideally be based on the SLOs that you define for your application (assuming you have defined enough of them and chosen them wisely).

SLOs are defined over longer fixed-size time windows and should preferably be expressed as a target for the ratio of "successful requests" to "total requests" that the application processes.
For example, `in this 28 day window, 99% of client requests completed successfully`.

Note that this doesn't restrict us to availability SLOs: latency targets can also be expressed as success ratios, by distinguishing between those requests that meet the target and those who don't. Having this standardized way of expressing SLOs comes in handy when defining SLO-based alerting.

### Simple SLO based alerting

Suppose we do have a target to serve 99% of requests successfully in our 28 day window. How can we set up prompt alerting so that we are notified when we are in risk of breaking our SLO?

The first approach we could consider is to pick a small time window and alert if we are above the target. The alert condition would look like this:
```
failure_ratio[10m] > 1.0 - SLO
```
We'll get notified as soon as we are not within our SLO for the 10m window that just ended. We won't miss a much, for sure. But that is the main drawback of this strategy: it has high recall, meaning it will cause us to waste time looking into small glitches that actually just consume a very small portion of our total error budget. You may be alerted many times daily, while you're actually staying within budget.

To mitigate that issue, another approach would be to simply define an alert based on a much larger time window:
```
failure_ratio[24h] > 1.0 - SLO
```
In case of a serious outage, this approach would still alert us fairly quickly, but the main drawback of this approach is that it has a big reset time, meaning the alert stays active for a long time even after an issue is resolved, so it's not very practical.

### Burn rate based alerting
This brings us to alerting based on burn rate. The burn rate represents how fast, relative to the SLO, the service consumes the error budget.

This sounds more complicated than it is, really. The whole idea is that we pick a reasonably small time window (like 1h or 30m) and a constant "burn_rate", and the alert condition just becomes:
```
failure_ratio[1h] > burn_rate * (1.0 - SLO)
```
So the burn_rate is just a positive multiplier, larger than 1, applied to the overall acceptable error budget. We are essentially acting as though we have have a higher tolerable error rate within any particular small window over which we are alerting.

### How does this solve our problems?

Because we don't use a large window, we don't have a big reset time problem. 

And because we have our burn_rate multiplier as part of the alert condition, we will not be getting bombarded with false positive alerts.

### How to choose the right burn rate?

So far, it isn't at all clear what constitutes a good burn_rate, or even that this is really a promising approach at all. Fortunately, we can run some calculations that give us more insight into how such alerting behaves before even putting it to work.

First of all, the burn_rate has an interesting relationship to the consumption of global error budget within the chosen window. Using the code in this repo, we can perform such calculations:
```
sloAlert, _ := NewSLOAlertFromBurnRate(0.99, 1*time.Hour, 10.0)
fmt.Println(sloAlert.PercentErrorBudgetConsumed)
```
This will output ~0.015, which tells us that if we have a 99% availability SLO (over 28 days) and we define an alert with a burn_rate=10 over the past hour, we will have consumed 1.5% of our total error budget by the time the alert fires.

In fact, it is possible to define your alerts in the form "Alert me when X% of the total error budget has been consumed in the past hour". This is just an alternative way of expressing the same kind of alert! Using the code in this repo, you could do it like this:
```
sloAlert, _ := NewSLOAlertFromBudgetUsed(0.99, 1*time.Hour, 0.03)
fmt.Println(sloAlert.BurnRate)
```
In the code above we have configured an alert to trigger when 3% of the total error budget has been used in the past hour. The output tells us that an alert with a burn_rate of just over 20 will achieve that.

What's more, once we have an alert configured, we can calculate whether the alert will fire and how long it would take for that alert to fire when we start seeing certain error rates. For example:
```
sloAlert, _ := NewSLOAlertFromBudgetUsed(0.99, 1*time.Hour, 0.03)
scenario, _ := NewScenario(sloAlert, 0.5)
fmt.Printf("Alert fires: %t\n", scenario.Check())
fmt.Printf("Detection time: %v\n", scenario.DetectionTime())
```
Here, we have our 99% availability SLO over 28 days and we configured an alert to fire whenever we use up more than 3% of our budget in 1h. Then we simulate a scenario in which we start to see a 50% error rate. The output is:
```
Alert fires: true
Detection time: 24m11.52s
```
So the alert will fire and it will take it 24 minutes after the start of the outage to do so. If we experience a total outage:
```
...
scenario, _ := NewScenario(sloAlert, 1.0)
...
```
Then the response time will be cut to 12m. If you feel this is a little too long, you can lower the percentage of error budget used slightly and this will give you an alert with a lower burn rate and that will trigger more quickly.

### What params should you choose?
using code in this repo, for configuration and scenario, print time to alerting.
Tune, no right answer - multi window explained in the book

### This is not the end of the story

While burn rate based alerting is clearly an improvement over the naive SLO-based alerting strategies mentioned in the beginning, this is not the end of the story.

You can build on top of burn rate based alerting by having multiple burn rate alerts and even multiwindow, multi-burn-rate alerts. You can read all about the pros and cons of each [here](https://sre.google/workbook/alerting-on-slos/).