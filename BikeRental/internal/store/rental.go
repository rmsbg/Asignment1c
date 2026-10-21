package store

import (
	"errors"
	"fmt"
	"sort"

	"bikerental/internal/domain"
)

var ErrUnavailable = errors.New("Das Ski-Paar ist im gewählten Zeitraum bereits gebucht. Bitte ein anderes Paar oder einen anderen Zeitraum auswählen.")

type MissingReference struct{ Field string }

func (e MissingReference) Error() string { return "Der ausgewählte Datensatz wurde nicht gefunden." }
func (e MissingReference) Unwrap() error { return ErrNotFound }

// RentalDetails joins current names under the same lock without identity evidence.
type RentalDetails struct {
	Today      string
	Rental     domain.Rental
	PersonName string
	SkiName    string
}

func (s *Store) rentalDetailsLocked(r domain.Rental) RentalDetails {
	return RentalDetails{Today: domain.Today(s.now()), Rental: r, PersonName: s.people[r.PersonID].Name, SkiName: s.skis[r.SkiID].Name}
}
func (s *Store) availablePeriodLocked(skiID, start, end string) bool {
	for _, r := range s.rentals {
		if r.SkiID == skiID && r.BlocksPeriod() && r.StartsOn <= end && start <= r.EndsOn {
			return false
		}
	}
	return true
}
func (s *Store) AvailableSkis() []domain.Ski {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Build once rather than scanning the rental history for every pair.
	active := make(map[string]bool)
	for _, r := range s.rentals {
		if r.Active() {
			active[r.SkiID] = true
		}
	}
	result := make([]domain.Ski, 0, len(s.skis))
	for id, ski := range s.skis {
		if !active[id] {
			result = append(result, ski)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// CreateRental validates price, references and overlap in one atomic operation.
func (s *Store) CreateRental(personID, skiID, start, end string, version uint64) (RentalDetails, error) {
	personID, skiID, err := domain.ValidateRentalReferences(personID, skiID)
	if err != nil {
		return RentalDetails{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.people[personID]; !ok {
		return RentalDetails{}, MissingReference{Field: "personID"}
	}
	if _, ok := s.skis[skiID]; !ok {
		return RentalDetails{}, MissingReference{Field: "skiID"}
	}
	quote, err := s.quoteLocked(start, end)
	if err != nil {
		return RentalDetails{}, err
	}
	if version != quote.TariffVersion {
		return RentalDetails{}, Conflict{"tariffVersion", "Der Tarif wurde geändert. Bitte den neuen Preis prüfen und erneut bestätigen."}
	}
	if !s.availablePeriodLocked(skiID, start, end) {
		return RentalDetails{}, ErrUnavailable
	}
	s.nextRental++
	r := domain.Rental{ID: fmt.Sprintf("ausleihe-%06d", s.nextRental), PersonID: personID, SkiID: skiID,
		CreatedAt: s.now().UTC(), StartsOn: start, EndsOn: end, DurationDays: quote.DurationDays,
		TotalPriceCents: quote.TotalPriceCents, TariffVersion: quote.TariffVersion}
	s.rentals[r.ID] = r
	return s.rentalDetailsLocked(r), nil
}

func (s *Store) AvailableFor(start, end string) ([]domain.Ski, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, err := domain.PeriodDays(start, end, domain.Today(s.now())); err != nil {
		return nil, err
	}
	result := []domain.Ski{}
	for id, ski := range s.skis {
		if s.availablePeriodLocked(id, start, end) {
			result = append(result, ski)
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
	today := domain.Today(s.now())
	if today < r.StartsOn || today > r.EndsOn {
		return RentalDetails{}, Conflict{"form", "Die Ausgabe ist nur innerhalb des gebuchten Zeitraums möglich."}
	}
	for _, other := range s.rentals {
		if other.SkiID == r.SkiID && other.Active() {
			return RentalDetails{}, Conflict{"form", "Das Ski-Paar wurde aus einer vorherigen Ausleihe noch nicht zurückgegeben."}
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
