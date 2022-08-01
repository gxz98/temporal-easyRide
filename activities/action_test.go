package activities

import (
	"context"
	. "easyRide/models"
	"log"
	"testing"
)

func TestInTrip(t *testing.T) {
	p := Passenger{ID: 1, Name: "xuezhou", DropLoc: 5}
	name := PassengerName("me")
	ctx := context.WithValue(context.Background(), name, p.Name)
	err := InTrip(ctx, name)
	if err != nil {
		log.Fatal(err)
	}
}
