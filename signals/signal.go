package signals

import (
	"context"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"log"
)

// signal definitions

const (
	MATCH_SIGNAL   = "signal_match"
	SIGNAL_PAYMENT = "signal_payment"
)

func SendMatchSignal(workflowID string, matchStatus bool) error {
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Println("Unable to create Temporal client", err)
		return err
	}
	err = temporalClient.SignalWorkflow(context.Background(), workflowID, "", MATCH_SIGNAL, matchStatus)
	if err != nil {
		log.Println("Error signaling workflow in execution ", err)
		return err
	}
	return nil
}

func ReceiveSignal(ctx workflow.Context, signalName string) (status bool) {
	res := workflow.GetSignalChannel(ctx, signalName).Receive(ctx, &status)
	if !res {
		log.Fatalf("%s channel is closed. Cannot receive status. ", signalName)
	}
	return
}

func SendPaymentSignal(workflowID string, paymentStatus bool) {
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client ", err)
	}
	err = temporalClient.SignalWorkflow(context.Background(), workflowID, "", SIGNAL_PAYMENT, paymentStatus)
	if err != nil {
		log.Fatalln("Error signaling workflow in execution ", err)
	}
}
