package domain

import "strings"

func (r Rental) Active() bool { return !r.RentedAt.IsZero() && r.ReturnedAt.IsZero() }

func ValidateRentalReferences(personID, skiID string) (string, string, error) {
	personID, skiID = strings.TrimSpace(personID), strings.TrimSpace(skiID)
	fields := ValidationErrors{}
	if personID == "" {
		fields["personID"] = "Bitte eine Person auswählen."
	}
	if skiID == "" {
		fields["skiID"] = "Bitte ein verfügbares Ski-Paar auswählen."
	}
	if len(fields) > 0 {
		return "", "", fields
	}
	return personID, skiID, nil
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
