package signals

import (
	"context"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"log"
)

// signal definitions

const (
	SIGNAL_MATCH   = "signal_match"
	SIGNAL_PAYMENT = "signal_payment"
)

func SendMatchSignal(workflowID string, matchStatus bool) {
	temporalClient, err := client.NewLazyClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client ", err)
	}
	err = temporalClient.SignalWorkflow(context.Background(), workflowID, "", SIGNAL_MATCH, matchStatus)
	if err != nil {
		log.Fatalln("Error signaling workflow in execution ", err)
	}
}

func ReceiveMatchSignal(ctx workflow.Context, signalName string) (matchStatus bool) {
	res := workflow.GetSignalChannel(ctx, signalName).Receive(ctx, &matchStatus)
	if !res {
		log.Fatalln("Match signal channel is closed. Cannot receive match status. ")
	}
	return
}

func SendPaymentSignal(workflowID string, paymentStatus bool) {
	temporalClient, err := client.NewLazyClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client ", err)
	}
	err = temporalClient.SignalWorkflow(context.Background(), workflowID, "", SIGNAL_PAYMENT, paymentStatus)
	if err != nil {
		log.Fatalln("Error signaling workflow in execution ", err)
	}
}

func ReceivePaymentSignal(ctx workflow.Context, signalName string) (paymentStatus bool) {
	res := workflow.GetSignalChannel(ctx, signalName).Receive(ctx, &paymentStatus)
	if !res {
		log.Fatalln("Payment signal channel is closed. Cannot receive match status. ")
	}
	return
}
