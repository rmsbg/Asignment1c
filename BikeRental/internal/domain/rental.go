package domain

import "strings"

func (r Rental) Active() bool { return !r.RentedAt.IsZero() && r.ReturnedAt.IsZero() }

func ValidateRentalReferences(personID, bikeID string) (string, string, error) {
	personID, bikeID = strings.TrimSpace(personID), strings.TrimSpace(bikeID)
	fields := ValidationErrors{}
	if personID == "" {
		fields["personID"] = "Bitte eine Person auswählen."
	}
	if bikeID == "" {
		fields["bikeID"] = "Bitte ein verfügbares Fahrrad auswählen."
	}
	if len(fields) > 0 {
		return "", "", fields
	}
	return personID, bikeID, nil
}

// BlocksPeriod remains true for issued rentals until actual return.
func (r Rental) BlocksPeriod() bool { return r.ReturnedAt.IsZero() && r.CancelledAt.IsZero() }

func (r Rental) Status(today string) string {
	if !r.CancelledAt.IsZero() {
		return "cancelled"
	}
	if !r.ReturnedAt.IsZero() {
		return "returned"
	}
	if !r.RentedAt.IsZero() {
		return "issued"
	}
	if r.EndsOn < today {
		return "expired"
	}
	return "booked"
}
