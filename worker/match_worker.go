package main

import (
	"easyRide/activities"
	"easyRide/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

// register workflows and activities to the worker
func main() {
	log.Println("Match worker Starting...")
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	cronWorker := worker.New(c, "matching", worker.Options{})

	cronWorker.RegisterWorkflow(workflows.MatchWorkFlow)
	cronWorker.RegisterActivity(activities.Match)

	if err := cronWorker.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
