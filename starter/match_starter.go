package starter

import (
	"easyRide/workflows"
	"go.temporal.io/sdk/client"
	"golang.org/x/net/context"
	"log"
)

func StartMatching(workflowID string) {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// add workflow options
	workflowOptions := client.StartWorkflowOptions{
		ID:           workflowID,
		TaskQueue:    "matching",
		CronSchedule: "@every 30s",
	}

	w, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflows.MatchWorkFlow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started matching workflow", "WorkflowID", w.GetID(), "RunID", w.GetRunID())
}
