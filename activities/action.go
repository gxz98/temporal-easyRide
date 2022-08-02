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
	DB   = "test"
)

// InTrip is the mock process of riding.
func InTrip(ctx context.Context, name string) error {
	log.Printf("Passenger %s is on a trip to destination....", name)
	// generate random number as trip duration time
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Int63n(30) + 60
	time.Sleep(time.Duration(randInt) * time.Second)
	return nil
}

// Arrive marks the passenger has arrived at the destination, update the driver status.
func Arrive(ctx context.Context, name string) error {
	log.Printf("Passenger %s arrive the destination...", name)
	// update the driver status
	db, err := data.Initialize(USR, PASS, DB)
	if err != nil {
		log.Fatal("Cannot connect to database.")
	}
	driverID, err := db.GetMatchedDriver(name)
	// change the drive availability to true
	if err := db.UpdateDriverStatus(driverID, &models.Passenger{}); err != nil {
		return err
	}
	return nil
}

// UpdatePassengerInfo updates the passenger's rating after a trip.
func UpdatePassengerInfo(ctx context.Context, name string) error {
	log.Printf("Updating passenger %s profile...", name)
	db, err := data.Initialize(USR, PASS, DB)
	if err != nil {
		log.Fatal("Cannot connect to database.")
	}
	err = db.UploadPassengerRating(name)
	if err != nil {
		return err
	}
	err = db.DeletePassenger(name)
	if err != nil {
		return err
	}
	return nil
}
