package workflows

import (
	"easyRide/activities"
	"go.temporal.io/sdk/workflow"
	"time"
)

// CronResult is used to pass data from one cron run to the next
type CronResult struct {
	RunTime time.Time
}

// MatchWorkFlow executes on the given schedule
// The schedule is provided when starting the workflow
func MatchWorkFlow(ctx workflow.Context) (*CronResult, error) {

	workflow.GetLogger(ctx).Info("Match Workflow started.", "StartTime", workflow.Now(ctx))

	// Define maximum time of a match round
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    60 * time.Second,
		ScheduleToCloseTimeout: 120 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	// Start from 0 to first cron job
	lastRunTime := time.Time{}
	// Update last run time if there was a previous successful job
	if workflow.HasLastCompletionResult(ctx) {
		var lastResult CronResult
		if err := workflow.GetLastCompletionResult(ctx, &lastResult); err == nil {
			lastRunTime = lastResult.RunTime
		}
	}
	thisRunTime := workflow.Now(ctx)

	err := workflow.ExecuteActivity(ctx1, activities.Match, lastRunTime, thisRunTime).Get(ctx, nil)
	if err != nil {
		// Match job failed
		workflow.GetLogger(ctx).Error("Match job failed.", "Error", err)
		return nil, err
	}

	return &CronResult{RunTime: thisRunTime}, nil
}
