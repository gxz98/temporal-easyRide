package activities

import (
	"context"
	"easyRide/activities/hungarian"
	postgres "easyRide/db"
	"easyRide/models"
	"go.temporal.io/sdk/activity"
	"math"
	"time"
)

const (
	HOST = "localhost"
	PORT = 5432
	USR  = "temporal"
	PASS = "temporal"
	DB   = "test"
)

func Match(ctx context.Context, lastRunTime, thisRunTime time.Time) error {
	activity.GetLogger(ctx).Info("Match job running.", "lastRunTime_exclude", lastRunTime, "thisRunTime_include", thisRunTime)
	db, err := postgres.Initialize(USR, PASS, DB)
	if err != nil {
		activity.GetLogger(ctx).Error("Database connection failed", "Error", err)
	}

	// Fetch unmatched passengers and driver
	p, errP := db.GetWaitingPassengers()
	if errP != nil {
		activity.GetLogger(ctx).Error("Cannot fetch waiting passengers", "Error", errP)
	}
	d, errD := db.GetAvailableDrivers()
	if errP != nil {
		activity.GetLogger(ctx).Error("Cannot fetch available driver", "Error", errD)
	}
	graph := constructGraph(p, d)

	// apply hungarian algo
	res, errG := hungarian.Solve(graph)
	if errG != nil {
		return errG
	}
	// update passenger and driver status in the database
	for p_idx, d_idx := range res {
		passenger := p.Passengers[p_idx]
		driver := d.Drivers[d_idx]
		errPS := db.UpdatePassengerStatus(passenger.ID, &driver)
		if errPS != nil {
			return errPS
		}
		errDS := db.UpdateDriverAvailability(driver.ID, &passenger)
		if errDS != nil {
			return errDS
		}

	}
	return nil
}

func constructGraph(p models.PassengerList, d models.DriverList) [][]float64 {
	nodeNum := min(10, min(len(p.Passengers), len(d.Drivers)))
	passenger := p.Passengers[:nodeNum]
	driver := d.Drivers[:nodeNum]

	// calculate the graph weight
	graph := make([][]float64, nodeNum)
	for row := range graph {
		graph[row] = make([]float64, nodeNum)
	}
	for i, ps := range passenger {
		for j, dr := range driver {
			// metric: distance/rating sum
			graph[i][j] = math.Abs(float64(ps.PickupLoc-dr.Loc)) / (ps.Rating + dr.Rating)
		}
	}
	return graph
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
