package starter

import (
	"easyRide/workflows"
	"go.temporal.io/sdk/client"
	"golang.org/x/net/context"
	"log"
)

func StartMatchWorkflow() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// add workflow options
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue:    "matching",
		CronSchedule: "@every 30s",
	}

	w, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflows.MatchWorkFlow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started matching workflow", "WorkflowID", w.GetID(), "RunID", w.GetRunID())
}

func StartMainWorkflow(workflowID string, passengerName string) {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// add workflow options
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "worker-group-1",
		ID:        workflowID,
	}

	w, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflows.MainWorkFlow, passengerName)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started main workflow", "WorkflowID", w.GetID(), "RunID", w.GetRunID())
}
