package workflows

import (
	"easyRide/activities"
	"easyRide/signals"
	"go.temporal.io/sdk/workflow"
	"log"
	"time"
)

// MainWorkFlow starts after the passenger logging in.
func MainWorkFlow(ctx workflow.Context, passengerID int) error {
	ao := workflow.ActivityOptions{
		// The trip can be long-time, set heartbeat instead of time to close
		HeartbeatTimeout: 1 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for {
		status := signals.ReceiveSignal(ctx, signals.MATCH_SIGNAL)
		if status == true {
			break
		} else {
			log.Printf("Cannot match passenger %d, trying again.", passengerID)
		}
	}

	log.Printf("Succesfully found driver for passenger %d", passengerID)
	err := workflow.ExecuteActivity(ctx, activities.InTrip, passengerID).Get(ctx, nil)
	if err != nil {
		return err
	}
	log.Printf("Driver please rate passenger %d", passengerID)
	// driver rate passenger
	err = workflow.Sleep(ctx, 15*time.Second)
	if err != nil {
		return err
	}
	err = workflow.ExecuteActivity(ctx, activities.Arrive, passengerID).Get(ctx, nil)
	if err != nil {
		return err
	}

	for {
		status := signals.ReceiveSignal(ctx, signals.SIGNAL_PAYMENT)
		if status == true {
			break
		} else {
			log.Println("Payment cannot be completed.")
		}
	}

	// passenger rate driver
	log.Printf("Passenger %d please rate driver", passengerID)
	err = workflow.Sleep(ctx, 15*time.Second)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, activities.PassengerEndTrip, passengerID).Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
