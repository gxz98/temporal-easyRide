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
	DB   = "test"
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

// AddPassenger add a client who starts a trip into the database.
func (db *Database) AddPassenger(passenger *models.Passenger) error {
	var id int
	var createdAt string
	var inRide bool
	query := `INSERT INTO passengers (name, pick_up_loc, drop_loc, rating) VALUES ($1, $2, $3, $4) RETURNING id, created_at, in_ride`
	err := db.Conn.QueryRow(query, passenger.Name, passenger.PickupLoc, passenger.DropLoc, passenger.Rating).Scan(&id, &createdAt, &inRide)
	if err != nil {
		return err
	}
	passenger.ID = id
	passenger.CreatedAt = createdAt
	passenger.InRide = inRide
	return nil
}

// AddDriver add a driver who starts to work into the database.
func (db *Database) AddDriver(driver *models.Driver) error {
	var id int
	var available bool
	var lastTripEndAt string
	query := `INSERT INTO drivers (name, loc, rating) VALUES ($1, $2, $3) RETURNING id, available, last_trip_end_at`
	err := db.Conn.QueryRow(query, driver.Name, driver.Loc, driver.Rating).Scan(&id, &available, &lastTripEndAt)
	if err != nil {
		return err
	}
	driver.ID = id
	driver.Available = available
	driver.LastTripEndAt = lastTripEndAt
	return nil
}

// every 10 second, fetch unmatched passengers and drivers from db, and run matching algorithm

// GetAvailableDrivers fetch all available drivers in descending order of their waiting time.
func (db *Database) GetAvailableDrivers() (models.DriverList, error) {
	list := models.DriverList{}
	query := `SELECT * FROM drivers WHERE available=TRUE ORDER BY last_trip_end_at ASC`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var driver models.Driver
		if err := rows.Scan(&driver.ID, &driver.Name, &driver.Loc,
			&driver.Available, &driver.Rating, &driver.WithPassenger, &driver.LastTripEndAt); err != nil {
			return list, err
		}
		list.Drivers = append(list.Drivers, driver)
	}
	return list, nil
}

// GetWaitingPassengers fetch all unmatched passengers in descending order of their waiting time.
func (db *Database) GetWaitingPassengers() (models.PassengerList, error) {
	list := models.PassengerList{}
	query := `SELECT * FROM passengers WHERE in_ride=FALSE ORDER BY created_at ASC`
	rows, err := db.Conn.Query(query)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var passenger models.Passenger
		if err := rows.Scan(&passenger.ID, &passenger.Name, &passenger.PickupLoc,
			&passenger.DropLoc, &passenger.Rating, &passenger.InRide, &passenger.WithDriver, &passenger.CreatedAt); err != nil {
			return list, err
		}
		list.Passengers = append(list.Passengers, passenger)
	}
	return list, nil
}

// DeletePassenger delete the client after the trip is end.
func (db *Database) DeletePassenger(passengerName string) error {
	query := `DELETE FROM passengers WHERE name = $1;`
	_, err := db.Conn.Exec(query, passengerName)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// DeleteDriver delete the driver when they choose to offline.
func (db *Database) DeleteDriver(driverName string) error {
	query := `DELETE FROM drivers WHERE name = $1;`
	_, err := db.Conn.Exec(query, driverName)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// UpdateDriverStatus update driver available status.
func (db *Database) UpdateDriverStatus(driverId int, withPassenger *models.Passenger) error {
	query := `UPDATE drivers SET available=NOT available, with_passenger=$2 WHERE id=$1;`
	_, err := db.Conn.Exec(query, driverId, (*withPassenger).ID)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// UpdateDriverLoc updates the driver's location regularly.
func (db *Database) UpdateDriverLoc(driverId int, newLoc int) error {
	query := `UPDATE drivers SET loc=$1 WHERE id=$2;`
	_, err := db.Conn.Exec(query, newLoc, driverId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
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

// UpdatePassengerStatus update passengers' status.
func (db *Database) UpdatePassengerStatus(passengerId int, withDriver *models.Driver) error {
	query := `UPDATE passengers SET in_ride=NOT in_ride, with_driver=$2 WHERE id=$1;`
	_, err := db.Conn.Exec(query, passengerId, (*withDriver).ID)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

// UpdatePassengerRating updates passengers' rating when receiving feedback from drivers.
func (db *Database) UpdatePassengerRating(passengerId int, newRating float64) error {
	var prevRating float64
	err := db.Conn.QueryRow(`SELECT rating FROM passengers WHERE id=$1`, passengerId).Scan(&prevRating)
	if err != nil {
		prevRating = 5.0
	}
	query := `UPDATE passengers SET rating=$1 WHERE id=$2;`
	_, err = db.Conn.Exec(query, newRating, passengerId)
	switch err {
	case sql.ErrNoRows:
		return ErrNoMatch
	default:
		return err
	}
}

func (db *Database) AddUser(userName string, password string, role string) error {
	query := `INSERT INTO users (name, password, role) VALUES ($1, $2, $3)`
	_, err := db.Conn.Exec(query, userName, password, role)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetUserPassword(userName string) (storedCreds *models.Credentials, e error) {
	query := `SELECT password FROM users WHERE username=$1`
	err := db.Conn.QueryRow(query, userName).Scan(storedCreds.Password)
	if err != nil {
		return storedCreds, err
	}
	return storedCreds, nil
}

func (db *Database) UpdateWorkFlowID(userName string, workflowId string) error {
	query := `UPDATE users SET workflow_id=$1 WHERE username=$2;`
	_, err := db.Conn.Exec(query, workflowId, userName)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetWorkFlowID(userName string) (id string, e error) {
	query := `SELECT workflow_id FROM users WHERE username=$1;`
	err := db.Conn.QueryRow(query, userName).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (db *Database) GetRating(userName string) (rating float64, e error) {
	query := `SELECT rating FROM users WHERE username=$1`
	err := db.Conn.QueryRow(query, userName).Scan(&rating)
	if err != nil {
		return rating, err
	}
	return rating, nil
}

func (db *Database) UploadPassengerRating(name string) error {
	var rating float64
	err := db.Conn.QueryRow(`SELECT rating FROM passengers WHERE name=$1`, name).Scan(&rating)
	if err != nil {
		return err
	}
	query := `UPDATE users SET rating=$1 WHERE username=$2`
	_, err = db.Conn.Exec(query, rating, name)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) UploadDriverRating(name string) error {
	var rating float64
	err := db.Conn.QueryRow(`SELECT rating FROM drivers WHERE name=$1`, name).Scan(&rating)
	if err != nil {
		return err
	}
	query := `UPDATE users SET rating=$1 WHERE username=$2`
	_, err = db.Conn.Exec(query, rating, name)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetRole(userName string) (role string, e error) {
	query := `SELECT role FROM users WHERE username=$1`
	err := db.Conn.QueryRow(query, userName).Scan(&role)
	if err != nil {
		return role, err
	}
	return role, nil
}

func (db *Database) GetPassengerStatus(userName string) (status bool, e error) {
	query := `SELECT in_ride FROM passengers WHERE username=$1`
	err := db.Conn.QueryRow(query, userName).Scan(&status)
	if err != nil {
		return false, err
	}
	return status, nil
}

func (db *Database) GetMatchedDriver(pName string) (driverID int, e error) {
	query := `SELECT with_driver FROM passengers WHERE username=$1`
	err := db.Conn.QueryRow(query, pName).Scan(&driverID)
	if err != nil {
		return 0, err
	}
	return driverID, nil
}

func (db *Database) GetMatchedPassenger(dName string) (passengerID int, e error) {
	query := `SELECT with_passenger FROM drivers WHERE username=$1`
	err := db.Conn.QueryRow(query, dName).Scan(&passengerID)
	if err != nil {
		return 0, err
	}
	return passengerID, nil
}

//func main() {
//	db, err := Initialize(USR, PASS, DB)
//	if err != nil {
//		log.Println("Database connection failed")
//		log.Println(err)
//	}
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
