package activities

import (
	"context"
	data "easyRide/db"
	"easyRide/models"
	"log"
	"math/rand"
	"time"
)

const (
	HOST = "localhost"
	PORT = 5432
	USR  = "temporal"
	PASS = "temporal"
	DB   = "postgres"
)

// InTrip is the mock process of riding.
func InTrip(ctx context.Context, passengerID int) error {
	log.Printf("Passenger %d is on a trip to destination....", passengerID)
	// generate random number as trip duration time
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Int63n(30) + 60
	time.Sleep(time.Duration(randInt) * time.Second)
	return nil
}

// Arrive marks the passenger has arrived at the destination, update the driver status.
func Arrive(ctx context.Context, passengerID int) error {
	log.Printf("Passenger %d arrive the destination...", passengerID)
	// update the driver status
	db, err := data.Initialize(USR, PASS, DB)
	if err != nil {
		log.Fatal("Cannot connect to database.")
	}
	driverID, err := db.GetMatchedDriver(passengerID)
	destination, err := db.GetDestination(passengerID)
	if err != nil {
		return err
	}
	// change the drive availability to true
	if err := db.UpdateDriverStatus(driverID, &models.Passenger{}); err != nil {
		return err
	}
	// change the last trip time of driver
	if err := db.UpdateLastTripEndTime(driverID); err != nil {
		return nil
	}
	// change the driver loc
	if err := db.UpdateDriverLoc(driverID, destination); err != nil {
		return nil
	}
	return nil
}

func PassengerEndTrip(ctx context.Context, passengerID int) error {
	db, err := data.Initialize(USR, PASS, DB)
	if err != nil {
		log.Fatal("Cannot connect to database.")
	}
	err = db.SetPassengerTripEnd(passengerID)
	if err != nil {
		return err
	}
	return nil
}
