// Package store provides the process-local, concurrency-safe application state.
package store

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"bikerental/internal/domain"
)

type Store struct {
	mu         sync.RWMutex
	now        func() time.Time
	tariff     domain.Tariff
	nextPerson uint64
	nextSki    uint64
	nextRental uint64
	demo       bool
	people     map[string]domain.Person
	skis       map[string]domain.Ski
	rentals    map[string]domain.Rental
}

// Snapshot contains independent value copies in stable ID order.
type Snapshot struct {
	People  []domain.Person
	Skis    []domain.Ski
	Rentals []domain.Rental
}

func New(demo bool) *Store {
	s := &Store{
		demo:    demo,
		now:     time.Now,
		tariff:  domain.Tariff{PricesCents: []int64{}},
		people:  make(map[string]domain.Person),
		skis:    make(map[string]domain.Ski),
		rentals: make(map[string]domain.Rental),
	}
	if demo {
		s.SeedDemo()
	}
	return s
}

// IsDemo is immutable for the lifetime of the store.
func (s *Store) IsDemo() bool { return s.demo }

// SeedDemo is idempotent and does not overwrite existing records.
// Normal-mode stores remain empty unless future business operations add data.
func (s *Store) SeedDemo() {
	if !s.demo {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tariff.Version == 0 {
		s.tariff = domain.Tariff{Version: 1, PricesCents: []int64{2500, 4500, 6000}, ExtraDayCents: 1500}
	}
	person := domain.Person{
		ID: "demo-person-01", Name: "Demo Person",
		DocumentType: "Sonstiger Ausweis", DocumentDescription: "Demoausweis",
		DocumentNumber: "DEMO-0001",
	}
	if _, exists := s.people[person.ID]; !exists {
		s.people[person.ID] = person
	}
	for _, ski := range []domain.Ski{
		{ID: "demo-ski-01", Name: "Demo Ski-Paar 01"},
		{ID: "demo-ski-02", Name: "Demo Ski-Paar 02"},
	} {
		if _, exists := s.skis[ski.ID]; !exists {
			s.skis[ski.ID] = ski
		}
	}
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := Snapshot{
		People:  make([]domain.Person, 0, len(s.people)),
		Skis:    make([]domain.Ski, 0, len(s.skis)),
		Rentals: make([]domain.Rental, 0, len(s.rentals)),
	}
	for _, person := range s.people {
		result.People = append(result.People, clonePerson(person))
	}
	for _, ski := range s.skis {
		result.Skis = append(result.Skis, ski)
	}
	for _, rental := range s.rentals {
		result.Rentals = append(result.Rentals, rental)
	}
	sort.Slice(result.People, func(i, j int) bool { return result.People[i].ID < result.People[j].ID })
	sort.Slice(result.Skis, func(i, j int) bool { return result.Skis[i].ID < result.Skis[j].ID })
	sort.Slice(result.Rentals, func(i, j int) bool { return result.Rentals[i].ID < result.Rentals[j].ID })
	return result
}

var ErrNotFound = errors.New("Datensatz nicht gefunden")

func clonePerson(p domain.Person) domain.Person { p.Photo = bytes.Clone(p.Photo); return p }

func (s *Store) Person(id string) (domain.Person, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.people[id]
	if !ok {
		return domain.Person{}, ErrNotFound
	}
	return clonePerson(p), nil
}
func (s *Store) Ski(id string) (domain.Ski, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.skis[id]
	if !ok {
		return domain.Ski{}, ErrNotFound
	}
	return p, nil
}
func (s *Store) SaveSki(id, name string) (domain.Ski, error) {
	name, err := domain.ValidateName(name)
	if err != nil {
		return domain.Ski{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == "" {
		s.nextSki++
		id = fmt.Sprintf("ski-%06d", s.nextSki)
	} else if _, ok := s.skis[id]; !ok {
		return domain.Ski{}, ErrNotFound
	}
	p := domain.Ski{ID: id, Name: name}
	s.skis[id] = p
	return p, nil
}

// SavePerson validates the entire candidate under the same lock as replacement.
// keepProof permits name-only edits without sending the previous identity data.
func (s *Store) SavePerson(id string, p domain.Person, keepProof bool) (domain.Person, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, exists := s.people[id]
	if id != "" && !exists {
		return domain.Person{}, ErrNotFound
	}
	if keepProof {
		if id == "" || len(p.Photo) > 0 || p.DocumentNumber != "" || p.DocumentType != "" || p.DocumentDescription != "" {
			return domain.Person{}, domain.ValidationErrors{"proof": "Unverändert ist nur ohne neuen Nachweis bei vorhandenen Personen möglich."}
		}
		name := p.Name
		p = old
		p.Name = name
	}
	p, err := domain.ValidatePerson(p)
	if err != nil {
		return domain.Person{}, err
	}
	if id == "" {
		s.nextPerson++
		id = fmt.Sprintf("person-%06d", s.nextPerson)
	}
	p.ID = id
	p.PhotoURL = ""
	s.people[id] = p
	return clonePerson(p), nil
}

// People exposes only the fields needed by overview callers, without photo copies.
func (s *Store) People() []domain.Person {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Person, 0, len(s.people))
	for _, p := range s.people {
		result = append(result, domain.Person{ID: p.ID, Name: p.Name})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
func (s *Store) Skis() []domain.Ski {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Ski, 0, len(s.skis))
	for _, p := range s.skis {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
