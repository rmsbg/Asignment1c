// Package domain defines the BikeRental data model.
package domain

import "time"

// Person represents a customer. Identity data must not appear in overview pages.
type Person struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	DocumentType        string `json:"documentType,omitempty"`
	DocumentDescription string `json:"documentDescription,omitempty"`
	DocumentNumber      string `json:"documentNumber,omitempty"`
	Photo               []byte `json:"-"`
	PhotoType           string `json:"-"`
	PhotoURL            string `json:"photoURL,omitempty"`
}

// Bike represents a pair of bikes rented and returned together.
type Bike struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Photo     []byte `json:"-"`
	PhotoType string `json:"-"`
	PhotoURL  string `json:"photoURL"`
}

// Rental separates the agreed calendar period from actual issue and return.
type Rental struct {
	ID              string
	CreatedAt       time.Time
	StartsOn        string
	EndsOn          string
	DurationDays    int
	TotalPriceCents int64
	TariffVersion   uint64
	CancelledAt     time.Time
	PersonID        string
	BikeID          string
	RentedAt        time.Time
	ReturnedAt      time.Time
}
