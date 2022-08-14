package main

import (
	"easyRide/activities"
	"easyRide/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	log.Println("Main worker Starting...")
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "worker-group-1", worker.Options{})
	w.RegisterWorkflow(workflows.MainWorkFlow)
	w.RegisterActivity(activities.InTrip)
	w.RegisterActivity(activities.Arrive)
	w.RegisterActivity(activities.PassengerEndTrip)
	w.RegisterActivity(activities.Rate)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln(err)
	}
}
