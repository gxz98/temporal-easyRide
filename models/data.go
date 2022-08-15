package models

// Passenger data model

type Passenger struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Password   string  `json:"password"`
	PickupLoc  int     `json:"pick_up_loc"`
	DropLoc    int     `json:"drop_loc"`
	Rating     float64 `json:"rating"`
	WorkflowID string  `json:"workflow_id"`
	InRide     bool    `json:"in_ride"`
	WithDriver int     `json:"with_driver"`
	CreatedAt  string  `json:"created_at"`
}

type PassengerList struct {
	Passengers []Passenger `json:"passengers"`
}

func (p *Passenger) Init(name string, pickupLoc int,
	dropLoc int, rating float64, createdAt string) {
	p.Name = name
	p.PickupLoc = pickupLoc
	p.DropLoc = dropLoc
	p.Rating = rating
	p.CreatedAt = createdAt
}

// Driver data model

type Driver struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Password      string  `json:"password"`
	Loc           int     `json:"loc"`
	Available     bool    `json:"available"`
	Rating        float64 `json:"rating"`
	WithPassenger int     `json:"with_passenger"`
	LastTripEndAt string  `json:"last_trip_end_at"`
}

type DriverList struct {
	Drivers []Driver `json:"drivers"`
}

func (d *Driver) Init(name string, loc int, rating float64, lastTripEndAt string) {
	d.Name = name
	d.Loc = loc
	d.Rating = rating
	d.LastTripEndAt = lastTripEndAt
}

type Credentials struct {
	ID       int    `json:"id"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type PassengerRequestBody struct {
	Name      string `json:"name"`
	ID        int    `json:"id"`
	PickupLoc int    `json:"pick_up_loc"`
	DropLoc   int    `json:"drop_loc"`
}

type DriverRequestBody struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	Loc  int    `json:"loc"`
}
