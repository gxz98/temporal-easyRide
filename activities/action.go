package activities

import (
	"context"
	data "easyRide/db"
	"log"
	"math/rand"
	"time"
)

const (
	HOST = "localhost"
	PORT = 5432
	USR  = "temporal"
	PASS = "temporal"
	DB   = "test"
)

type PassengerName string

var db data.Database

func InTrip(ctx context.Context, name PassengerName) error {
	log.Printf("Passenger %s is on a trip to destination....", ctx.Value(name))
	// generate random number as trip duration time
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Int63n(30) + 60
	time.Sleep(time.Duration(randInt) * time.Second)
	return nil
}
