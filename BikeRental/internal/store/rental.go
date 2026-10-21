package store

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"bikerental/internal/domain"
)

var ErrUnavailable = errors.New("Das Fahrrad ist im gewählten Zeitraum bereits gebucht. Bitte ein anderes Fahrrad oder einen anderen Zeitraum auswählen.")

type MissingReference struct{ Field string }

func (e MissingReference) Error() string { return "Der ausgewählte Datensatz wurde nicht gefunden." }
func (e MissingReference) Unwrap() error { return ErrNotFound }

// RentalDetails joins current names under the same lock without identity evidence.
type RentalDetails struct {
	Now        time.Time
	Rental     domain.Rental
	PersonName string
	BikeName   string
}

func (s *Store) rentalDetailsLocked(r domain.Rental) RentalDetails {
	return RentalDetails{Now: s.now().UTC(), Rental: r, PersonName: s.people[r.PersonID].Name, BikeName: s.bikes[r.BikeID].Name}
}
func (s *Store) availablePeriodLocked(bikeID string, start, end time.Time) bool {
	for _, r := range s.rentals {
		if r.BikeID == bikeID && r.BlocksPeriod() && r.StartsAt.Before(end) && start.Before(r.EndsAt) {
			return false
		}
	}
	return true
}
func (s *Store) AvailableBikes() []domain.Bike {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Build once rather than scanning the rental history for every pair.
	active := make(map[string]bool)
	for _, r := range s.rentals {
		if r.Active() {
			active[r.BikeID] = true
		}
	}
	result := make([]domain.Bike, 0, len(s.bikes))
	for id, bike := range s.bikes {
		if !active[id] {
			result = append(result, cloneBike(bike))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// CreateRental validates price, references and overlap in one atomic operation.
func (s *Store) CreateRental(input domain.BookingInput) (RentalDetails, error) {
	personID, bikeID, err := domain.ValidateRentalReferences(input.PersonID, input.BikeID)
	if err != nil {
		return RentalDetails{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.people[personID]; !ok {
		return RentalDetails{}, MissingReference{Field: "personID"}
	}
	bike, ok := s.bikes[bikeID]
	if !ok {
		return RentalDetails{}, MissingReference{Field: "bikeID"}
	}
	if !domain.ValidBikeType(input.BikeType) {
		return RentalDetails{}, domain.ValidationErrors{"bikeType": "Bitte einen gültigen Fahrradtyp auswählen."}
	}
	if bike.Type != input.BikeType {
		return RentalDetails{}, Conflict{"bikeType", "Der Fahrradtyp wurde geändert. Bitte Verfügbarkeit und Preis erneut prüfen."}
	}
	quote, err := s.quoteLocked(bike.Type, input.Mode, input.StartsOn, input.EndsOn)
	if err != nil {
		return RentalDetails{}, err
	}
	if input.TariffVersion == 0 {
		return RentalDetails{}, domain.ValidationErrors{"tariffVersion": "Bitte zuerst den Preis prüfen und bestätigen."}
	}
	if input.TariffVersion != quote.TariffVersion {
		return RentalDetails{}, Conflict{"tariffVersion", "Der Tarif wurde geändert. Bitte den neuen Preis prüfen und erneut bestätigen."}
	}
	if !s.availablePeriodLocked(bikeID, quote.StartsAt, quote.EndsAt) {
		return RentalDetails{}, ErrUnavailable
	}
	s.nextRental++
	r := domain.Rental{ID: fmt.Sprintf("ausleihe-%06d", s.nextRental), PersonID: personID, BikeID: bikeID,
		BikeType: bike.Type, Mode: input.Mode, StartsAt: quote.StartsAt, EndsAt: quote.EndsAt, DurationHours: quote.DurationHours,
		CreatedAt: s.now().UTC(), StartsOn: input.StartsOn, EndsOn: input.EndsOn, DurationDays: quote.DurationDays,
		TotalPriceCents: quote.TotalPriceCents, TariffVersion: quote.TariffVersion}
	s.rentals[r.ID] = r
	return s.rentalDetailsLocked(r), nil
}

func (s *Store) AvailableFor(mode, start, end string) ([]domain.Bike, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	period, err := domain.BuildPeriod(mode, start, end, s.now())
	if err != nil {
		return nil, err
	}
	result := []domain.Bike{}
	for id, bike := range s.bikes {
		if s.availablePeriodLocked(id, period.StartsAt, period.EndsAt) {
			result = append(result, cloneBike(bike))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
func (s *Store) IssueRental(id string) (RentalDetails, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rentals[id]
	if !ok {
		return RentalDetails{}, ErrNotFound
	}
	if !r.CancelledAt.IsZero() {
		return RentalDetails{}, Conflict{"form", "Eine stornierte Buchung kann nicht ausgegeben werden."}
	}
	if !r.RentedAt.IsZero() {
		return s.rentalDetailsLocked(r), nil
	}
	now := s.now()
	if now.Before(r.StartsAt) || !now.Before(r.EndsAt) {
		return RentalDetails{}, Conflict{"form", "Die Ausgabe ist nur innerhalb des gebuchten Zeitraums möglich."}
	}
	for _, other := range s.rentals {
		if other.BikeID == r.BikeID && other.Active() {
			return RentalDetails{}, Conflict{"form", "Das Fahrrad wurde aus einer vorherigen Ausleihe noch nicht zurückgegeben."}
		}
	}
	r.RentedAt = s.now().UTC()
	s.rentals[id] = r
	return s.rentalDetailsLocked(r), nil
}
func (s *Store) CancelRental(id string) (RentalDetails, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rentals[id]
	if !ok {
		return RentalDetails{}, ErrNotFound
	}
	if !r.RentedAt.IsZero() {
		return RentalDetails{}, Conflict{"form", "Eine bereits ausgegebene Ausleihe kann nicht storniert werden."}
	}
	if r.CancelledAt.IsZero() {
		r.CancelledAt = s.now().UTC()
		s.rentals[id] = r
	}
	return s.rentalDetailsLocked(r), nil
}
func (s *Store) ReturnRental(id string) (RentalDetails, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.rentals[id]
	if !ok {
		return RentalDetails{}, ErrNotFound
	}
	if r.RentedAt.IsZero() {
		return RentalDetails{}, Conflict{"form", "Eine Rückgabe ist erst nach der Ausgabe möglich."}
	}
	if r.Active() {
		r.ReturnedAt = s.now().UTC()
		s.rentals[id] = r
	}
	return s.rentalDetailsLocked(r), nil
}
func (s *Store) Rental(id string) (RentalDetails, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.rentals[id]
	if !ok {
		return RentalDetails{}, ErrNotFound
	}
	return s.rentalDetailsLocked(r), nil
}
func (s *Store) Rentals() []RentalDetails {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]RentalDetails, 0, len(s.rentals))
	for _, r := range s.rentals {
		result = append(result, s.rentalDetailsLocked(r))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Rental.ID < result[j].Rental.ID })
	return result
}
