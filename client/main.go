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
)

const (
	HOST = "localhost"
	PORT = 5432
	USR  = "temporal"
	PASS = "temporal"
	DB   = "postgres"
)

var db data.Database

func main() {
	router := mux.NewRouter()

	var err error
	db, err = data.Initialize(USR, PASS, DB)
	if err != nil {
		panic(err)
	}

	router.HandleFunc("/about", GetAbout)
	router.HandleFunc("/start-engine", Start)
	// User sign up
	router.HandleFunc("/passenger/signup", PassengerSignUpHandler)
	router.HandleFunc("/driver/signup", DriverSignUpHandler)

	// Sign in: validate the login info in the database
	router.HandleFunc("/passenger/login", PassengerLogInHandler)
	router.HandleFunc("/driver/login", DriverLogInHandler)

	// passenger request a trip
	router.HandleFunc("/passenger/start-trip", StartTripHandler)
	// driver start serving passenger
	router.HandleFunc("/driver/start-work", StartWorkHandler)

	// After trip, rate and pay
	router.HandleFunc("/passenger/payment/{pay}", PaymentHandler)
	router.HandleFunc("/passenger/rating/{rating}", PassengerRatingHandler)
	router.HandleFunc("/driver/rating/{rating}", DriverRatingHandler)

	router.HandleFunc("/driver/end-work", EndWorkHandler)
	// more features
	//router.HandleFunc("/driver/confirm-trip/{confirm}", ConfirmTripHandler)
	//router.HandleFunc("/passenger/report-danger", DangerHandler)
	//router.HandleFunc("/passenger/cancel", CancelHandler)
	//router.HandleFunc("/passenger/change-destination", DestinationChangeHandler)
	log.Fatal(http.ListenAndServe(":3310", router))
}

func Start(writer http.ResponseWriter, request *http.Request) {
	starter.StartMatchWorkflow()
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
	err = db.AddPassenger(creds.Username, string(hashedPassword))
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
	err = db.AddDriver(creds.Username, string(hashedPassword))
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// credentials are correctly stored in the database, send default status of 200
}

func PassengerLogInHandler(writer http.ResponseWriter, request *http.Request) {
	creds := &models.Credentials{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	// Get the existing entry in the database for the given username
	storedPassword, id, err := db.GetPassword(creds.Username, "passenger")
	if err != nil {
		// If the username does not exist
		if err == sql.ErrNoRows {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)); err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
	}

	// Log in successfully, start the workflow
	workFlowUUID, err := uuid.NewUUID()
	if err != nil {
		writer.Write([]byte(err.Error()))
		return
	}
	// Update user workflow ID
	err = db.UpdateWorkFlowID(id, workFlowUUID.String())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}

	starter.StartMainWorkflow(workFlowUUID.String(), id)
	writer.WriteHeader(http.StatusOK)
}

func DriverLogInHandler(writer http.ResponseWriter, request *http.Request) {
	creds := &models.Credentials{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	// Get the existing entry in the database for the given username
	storedPassword, _, err := db.GetPassword(creds.Username, "driver")
	if err != nil {
		// If the username does not exist
		if err == sql.ErrNoRows {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)); err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
	}
}

func StartTripHandler(writer http.ResponseWriter, request *http.Request) {
	passenger := &models.PassengerRequestBody{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(passenger)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	err = db.UpdatePassengerLoc(passenger)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
}

func StartWorkHandler(writer http.ResponseWriter, request *http.Request) {
	driver := &models.DriverRequestBody{}
	// Decode the request body into a new Credential struct
	err := json.NewDecoder(request.Body).Decode(driver)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}

	err = db.UpdateDriverLoc(driver.ID, driver.Loc)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(err.Error()))
		return
	}
}

func PaymentHandler(writer http.ResponseWriter, request *http.Request) {
	passenger := &models.PassengerRequestBody{}
	if err := json.NewDecoder(request.Body).Decode(passenger); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	workflowID, err := db.GetWorkFlowID(passenger.ID)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	vars := mux.Vars(request)
	actualPay, _ := strconv.ParseFloat(vars["pay"], 64)
	expectedPay := math.Abs(float64(passenger.PickupLoc - passenger.DropLoc))
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
	passenger := &models.PassengerRequestBody{}
	if err := json.NewDecoder(request.Body).Decode(passenger); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	rating, _ := strconv.ParseFloat(vars["rating"], 64)
	driverID, err := db.GetMatchedDriver(passenger.ID)
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
	driver := &models.DriverRequestBody{}
	if err := json.NewDecoder(request.Body).Decode(driver); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	rating, _ := strconv.ParseFloat(vars["rating"], 64)
	passengerID, err := db.GetMatchedPassenger(driver.ID)
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
	driver := &models.DriverRequestBody{}
	if err := json.NewDecoder(request.Body).Decode(driver); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
	err := db.SetDriverOffline(driver.ID)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(err.Error()))
		return
	}
}
