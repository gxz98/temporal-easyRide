package db

import (
	"database/sql"
	"easyRide/models"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

const (
	HOST = "localhost"
	PORT = 5432
	USR  = "temporal"
	PASS = "temporal"
	DB   = "postgres"
)

type Database struct {
	Conn *sql.DB
}

// ErrNoMatch is returned when we request a row that doesn't exist
var ErrNoMatch = fmt.Errorf("no matching record")

// Initialize will establish a db connection.
func Initialize(username, password, database string) (Database, error) {
	db := Database{}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		HOST, PORT, username, password, database)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return db, err
	}
	db.Conn = conn
	err = db.Conn.Ping()
	if err != nil {
		return db, err
	}
	log.Println("Database connection established")
	return db, nil
}

// Driver database

func (db *Database) AddDriver(name string, password string) error {
	query := `INSERT INTO drivers (name, password) VALUES ($1, $2)`
	_, err := db.Conn.Query(query, name, password)
	if err != nil {
		return err
	}
	return nil
}

// GetAvailableDrivers fetch all available drivers in descending order of their waiting time.
func (db *Database) GetAvailableDrivers() (models.DriverList, error) {
	list := models.DriverList{}
	query := `SELECT * FROM drivers WHERE available=TRUE AND loc>=0 ORDER BY last_trip_end_at ASC`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var driver models.Driver
		if err := rows.Scan(&driver.ID, &driver.Name, &driver.Password, &driver.Loc,
			&driver.Available, &driver.Rating, &driver.WithPassenger, &driver.LastTripEndAt); err != nil {
			return list, err
		}
		list.Drivers = append(list.Drivers, driver)
	}
	return list, nil
}

// UpdateDriverStatus update driver available status.
func (db *Database) UpdateDriverStatus(driverId int, withPassenger *models.Passenger, status bool) error {
	query := `UPDATE drivers SET available=$3, with_passenger=$2 WHERE id=$1;`
	_, err := db.Conn.Exec(query, driverId, (*withPassenger).ID, status)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// UpdateDriverLoc updates the driver's location regularly.
func (db *Database) UpdateDriverLoc(driverID int, loc int) error {
	query := `UPDATE drivers SET loc=$1 WHERE id=$2;`
	_, err := db.Conn.Exec(query, loc, driverID)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) SetDriverOffline(DriverID int) error {
	query := `UPDATE drivers SET loc=-100,available=FALSE WHERE id=$1;`
	_, err := db.Conn.Exec(query, DriverID)
	if err != nil {
		return err
	}
	return nil
}

// UpdateLastTripEndTime updates the driver's last trip end time.
func (db *Database) UpdateLastTripEndTime(driverId int) error {
	query := `UPDATE drivers SET last_trip_end_at=$1 WHERE id=$2;`
	_, err := db.Conn.Exec(query, time.Now(), driverId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// UpdateDriverRating updates the driver's rating when receiving passengers' feedback.
func (db *Database) UpdateDriverRating(driverId int, newRating float64) error {
	var prevRating float64
	err := db.Conn.QueryRow(`SELECT rating FROM drivers WHERE id=$1`, driverId).Scan(&prevRating)
	if err != nil {
		prevRating = 5.0
	}
	query := `UPDATE drivers SET rating=$1 WHERE id=$2;`
	_, err = db.Conn.Exec(query, (newRating+prevRating)/2, driverId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db *Database) GetMatchedPassenger(driverId int) (passengerID int, e error) {
	query := `SELECT with_passenger FROM drivers WHERE id=$1`
	err := db.Conn.QueryRow(query, driverId).Scan(&passengerID)
	if err != nil {
		return 0, err
	}
	return passengerID, nil
}

// Passenger database

func (db *Database) AddPassenger(name string, password string) error {
	query := `INSERT INTO passengers (name, password) VALUES ($1, $2)`
	_, err := db.Conn.Query(query, name, password)
	if err != nil {
		return err
	}
	return nil
}

// GetWaitingPassengers fetch all unmatched passengers in descending order of their waiting time.
func (db *Database) GetWaitingPassengers() (models.PassengerList, error) {
	list := models.PassengerList{}
	query := `SELECT * FROM passengers WHERE in_ride=FALSE AND drop_loc>=0 ORDER BY created_at ASC`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var passenger models.Passenger
		if err := rows.Scan(&passenger.ID, &passenger.Name, &passenger.Password, &passenger.PickupLoc,
			&passenger.DropLoc, &passenger.Rating, &passenger.WorkflowID, &passenger.InRide,
			&passenger.WithDriver, &passenger.CreatedAt); err != nil {
			return list, err
		}
		list.Passengers = append(list.Passengers, passenger)
	}
	return list, nil
}

// UpdatePassengerStatus update passengers' status.
func (db *Database) UpdatePassengerStatus(passengerId int, withDriver *models.Driver, status bool) error {
	query := `UPDATE passengers SET in_ride=$3, with_driver=$2 WHERE id=$1;`
	_, err := db.Conn.Exec(query, passengerId, (*withDriver).ID, status)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db *Database) SetPassengerTripEnd(passengerID int) error {
	query := `UPDATE passengers SET drop_loc=-100,in_ride=FALSE WHERE id=$1;`
	_, err := db.Conn.Exec(query, passengerID)
	if err != nil {
		return err
	}
	return nil
}

// UpdatePassengerRating updates passengers' rating when receiving feedback from drivers.
func (db *Database) UpdatePassengerRating(passengerId int, newRating float64) error {
	var prevRating float64
	err := db.Conn.QueryRow(`SELECT rating FROM passengers WHERE id=$1`, passengerId).Scan(&prevRating)
	if err != nil {
		prevRating = 5.0
	}
	query := `UPDATE passengers SET rating=$1 WHERE id=$2;`
	_, err = db.Conn.Exec(query, (newRating+prevRating)/2, passengerId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db *Database) UpdateWorkFlowID(passengerId int, workflowId string) error {
	query := `UPDATE passengers SET workflow_id=$1 WHERE id=$2;`
	_, err := db.Conn.Exec(query, workflowId, passengerId)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetWorkFlowID(passengerId int) (id string, e error) {
	query := `SELECT workflow_id FROM passengers WHERE id=$1;`
	err := db.Conn.QueryRow(query, passengerId).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (db *Database) GetMatchedDriver(passengerId int) (driverID int, e error) {
	query := `SELECT with_driver FROM passengers WHERE id=$1`
	err := db.Conn.QueryRow(query, passengerId).Scan(&driverID)
	if err != nil {
		return 0, err
	}
	return driverID, nil
}

func (db *Database) GetDestination(passengerId int) (destination int, e error) {
	query := `SELECT drop_loc FROM passengers WHERE id=$1`
	err := db.Conn.QueryRow(query, passengerId).Scan(&destination)
	if err != nil {
		return 0, err
	}
	return destination, nil
}

func (db *Database) GetPassword(userName string, table string) (password string, id int, e error) {
	var query string
	if table == "passenger" {
		query = `SELECT password,id FROM passengers WHERE name=$1`
	} else {
		query = `SELECT password,id FROM drivers WHERE name=$1`
	}
	err := db.Conn.QueryRow(query, userName).Scan(&password, &id)
	if err != nil {
		return "", 0, err
	}
	return password, id, nil
}

func (db *Database) UpdatePassengerLoc(body *models.PassengerRequestBody) error {
	query := `UPDATE passengers SET pick_up_loc=$1, drop_loc=$2 WHERE id=$3;`
	_, err := db.Conn.Exec(query, body.PickupLoc, body.DropLoc, body.ID)
	if err != nil {
		return err
	}
	return nil
}

//func main() {
//	db, err := Initialize(USR, PASS, DB)
//	if err != nil {
//		log.Println("Database connection failed")
//		log.Println(err)
//	}
//
//	err = db.UpdatePassengerStatus(1, &models.Driver{}, true)
//	if err != nil {
//		return
//	}
//
//	ps, _ := db.GetWaitingPassengers()
//	fmt.Println(ps)
//	dr, _ := db.GetAvailableDrivers()
//	fmt.Println(dr)
//
//	passenger1 := &models.Passenger{}
//	passenger1.Init("passenger_1", 3, 6, 4.8, time.Now().String())
//	err = db.AddPassenger(passenger1)
//	if err != nil {
//		return
//	}
//	log.Println("Adding client", passenger1)
//
//	passenger2 := &models.Passenger{}
//	passenger2.Init("passenger_2", 1, 9, 4.6, time.Now().String())
//	err = db.AddPassenger(passenger2)
//	if err != nil {
//		return
//	}
//	log.Println("Adding client", passenger2)
//
//	driver1 := &models.Driver{}
//	driver1.Init("driver_1", 4, 4.9, time.Now().String())
//	err = db.AddDriver(driver1)
//	if err != nil {
//		return
//	}
//	log.Println("Adding driver", driver1)
//
//	driver2 := &models.Driver{}
//	driver2.Init("driver_2", 7, 4.9, time.Now().String())
//	err = db.AddDriver(driver2)
//	if err != nil {
//		return
//	}
//	log.Println("Adding driver", driver2)
//
//	// driver1 match with passenger1
//	log.Println("start trip.....")
//	err = db.UpdatePassengerStatus(passenger1.ID, driver1)
//	if err != nil {
//		return
//	}
//	log.Println("Changed client state", passenger1)
//
//	err = db.UpdateDriverStatus(driver1.ID, passenger1)
//	if err != nil {
//		return
//	}
//	log.Println("Changed driver state", driver1)
//
//	// check unmatched passengers and drivers
//	drivers, _ := db.GetAvailableDrivers()
//	fmt.Println("available drivers: ", drivers.Drivers[0].Name)
//
//	passengers, _ := db.GetWaitingPassengers()
//	fmt.Println("waiting passengers ", passengers.Passengers[0].Name)
//
//	//// trip end
//	log.Println("end trip.....")
//	err = db.UpdateDriverStatus(driver1.ID, passenger1)
//	if err != nil {
//		return
//	}
//	log.Println("Changed driver state", driver1)
//	err = db.UpdateDriverLoc(driver1.ID, 6)
//	if err != nil {
//		return
//	}
//	log.Println("Changed driver location", driver1)
//	err = db.DeletePassenger(passenger1.ID)
//	if err != nil {
//		return
//	}
//	log.Println("delete client", passenger1)
//
//}
