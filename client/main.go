package main

import (
	"database/sql"
	data "easyRide/db"
	models "easyRide/models"
	"easyRide/signals"
	"easyRide/starter"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	HOST = "localhost"
	PORT = 5432
	USR  = "temporal"
	PASS = "temporal"
	DB   = "test"
)

var db data.Database

func main() {
	router := mux.NewRouter()

	var err error
	db, err = data.Initialize(USR, PASS, DB)
	if err != nil {
		panic(err)
	}

	// Endpoints(URL) design
	router.HandleFunc("/about", GetAbout)
	// initialize user info and upload to database
	// TODO: validate the user name
	router.HandleFunc("/passenger/signup", PassengerSignUpHandler)
	router.HandleFunc("/driver/signup", DriverSignUpHandler)
	// validate the login info in the database
	router.HandleFunc("/login", LogInHandler)
	// after login, start a workflow when passenger starting a trip
	router.HandleFunc("/passenger/start-trip", StartTripHandler)
	router.HandleFunc("/driver/start-work", StartWorkHandler)
	// after trip, rating passenger and driver
	router.HandleFunc("/passenger/payment/{pay}", PaymentHandler)
	router.HandleFunc("/passenger/rating/{rating}", PassengerRatingHandler)
	router.HandleFunc("/driver/rating/{rating}", DriverRatingHandler)
	router.HandleFunc("/driver/end-work", EndWorkHandler)
	// more features
	//router.HandleFunc("/driver/confirm-trip/{confirm}", ConfirmTripHandler)
	//router.HandleFunc("/passenger/report-danger", DangerHandler)
	//router.HandleFunc("/passenger/cancel", CancelHandler)
	//router.HandleFunc("/passenger/change-destination", DestinationChangeHandler)
	log.Fatal(http.ListenAndServe("3310", router))
}

func GetAbout(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte("Easy Ride: Make transportation more convenient"))
	if err != nil {
		log.Fatalln(err)
	}
}

func PassengerSignUpHandler(writer http.ResponseWriter, request *http.Request) {
	creds := &models.Credentials{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		panic(err)
	}
	err = db.AddUser(creds.Username, string(hashedPassword), "passenger")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// credentials are correctly stored in the database, send default status of 200
}

func DriverSignUpHandler(writer http.ResponseWriter, request *http.Request) {
	creds := &models.Credentials{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	// Hash the password
	var hashedPassword []byte
	hashedPassword, err = bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		panic(err)
	}
	err = db.AddUser(creds.Username, string(hashedPassword), "driver")
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// credentials are correctly stored in the database, send default status of 200
}

func LogInHandler(writer http.ResponseWriter, request *http.Request) {
	creds := &models.Credentials{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	// Get the existing entry in the database for the given username
	storedCreds, err := db.GetUserPassword(creds.Username)
	if err != nil {
		// If the username does not exist
		if err == sql.ErrNoRows {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
	}

	// Only passenger start the workflow
	if role, err := db.GetRole(creds.Username); err != nil {
		writer.Write([]byte(err.Error()))
		return
	} else {
		if role == "driver" {
			return
		}
	}

	// Log in successfully, start the workflow
	workFlowUUID, err := uuid.NewUUID()
	if err != nil {
		writer.Write([]byte(err.Error()))
		return
	}
	// Update user workflow ID
	err = db.UpdateWorkFlowID(creds.Username, workFlowUUID.String())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}

	starter.StartMainWorkflow(workFlowUUID.String(), creds.Username)
	writer.WriteHeader(http.StatusOK)
}

// StartTripHandler add passenger to the database for matching
func StartTripHandler(writer http.ResponseWriter, request *http.Request) {
	loc := &models.PassengerLocation{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(loc)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	rating, err := db.GetRating(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	// add the passenger in matching database
	passenger := &models.Passenger{}
	passenger.Init(loc.Name, loc.PickupLoc, loc.DropLoc, rating, time.Now().String())
	err = db.AddPassenger(passenger)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
}

// StartWorkHandler add driver to the database for matching
func StartWorkHandler(writer http.ResponseWriter, request *http.Request) {
	loc := &models.DriverLocation{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(loc)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	rating, err := db.GetRating(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}

	// add the passenger in matching database
	driver := &models.Driver{}
	driver.Init(loc.Name, loc.Loc, rating, time.Now().String())
	err = db.AddDriver(driver)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
}

func PaymentHandler(writer http.ResponseWriter, request *http.Request) {
	loc := &models.PassengerLocation{}
	if err := json.NewDecoder(request.Body).Decode(loc); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	workflowID, err := db.GetWorkFlowID(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	vars := mux.Vars(request)
	actualPay, _ := strconv.ParseFloat(vars["pay"], 64)
	expectedPay := math.Abs(float64(loc.PickupLoc - loc.DropLoc))
	if actualPay < expectedPay {
		signals.SendPaymentSignal(workflowID, false)
		return
	}
	signals.SendPaymentSignal(workflowID, true)
}

func PassengerRatingHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if vars["rating"] == "" {
		return
	}
	loc := &models.PassengerLocation{}
	if err := json.NewDecoder(request.Body).Decode(loc); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	rating, _ := strconv.ParseFloat(vars["rating"], 64)
	driverID, err := db.GetMatchedDriver(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.UpdateDriverRating(driverID, rating)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// DriverRatingHandler is for drivers to rate their passengers.
// Before the system mark this trip as "arrive".
func DriverRatingHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	if vars["rating"] == "" {
		return
	}
	loc := &models.DriverLocation{}
	if err := json.NewDecoder(request.Body).Decode(loc); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	rating, _ := strconv.ParseFloat(vars["rating"], 64)
	passengerID, err := db.GetMatchedPassenger(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Driver don't response to the rating popup window, miss the chance to rate the passenger.
	if passengerID == 0 {
		return
	}
	err = db.UpdatePassengerRating(passengerID, rating)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// EndWorkHandler is used by drivers to get offline.
func EndWorkHandler(writer http.ResponseWriter, request *http.Request) {
	loc := &models.DriverLocation{}
	if err := json.NewDecoder(request.Body).Decode(loc); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	err := db.UploadDriverRating(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
	err = db.DeleteDriver(loc.Name)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
}
