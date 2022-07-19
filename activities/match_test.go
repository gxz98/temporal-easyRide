package activities

import (
	"easyRide/activities/hungarian"
	"easyRide/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func setUp() (models.PassengerList, models.DriverList) {
	passenger1 := &models.Passenger{}
	passenger1.Init("passenger_1", 3, 6, 5.0, time.Now().String())

	passenger2 := &models.Passenger{}
	passenger2.Init("passenger_2", 9, 6, 5.0, time.Now().String())

	driver1 := &models.Driver{}
	driver1.Init("driver_1", 8, 5.0, time.Now().String())

	driver2 := &models.Driver{}
	driver2.Init("driver_2", 15, 5.0, time.Now().String())

	pl := models.PassengerList{Passengers: []models.Passenger{*passenger1, *passenger2}}
	dl := models.DriverList{Drivers: []models.Driver{*driver1, *driver2}}
	return pl, dl
}

func TestConstructGraph(t *testing.T) {
	pl, dl := setUp()
	res := constructGraph(pl, dl)
	expected := [][]float64{
		{0.5, 1.2},
		{0.1, 0.6},
	}
	assert.Equal(t, expected, res)
}

func TestHungarianAlgo(t *testing.T) {
	graph := [][]float64{
		{0.5, 1.2, 0.3}, // 0.3
		{0.1, 0.6, 0.4}, // 0.6
		{0.2, 1.4, 0.5}, // 0.2
	}
	res, _ := hungarian.Solve(graph)
	expected := []int{2, 1, 0}
	assert.Equal(t, expected, res)
}
