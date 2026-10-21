// Package store provides the process-local, concurrency-safe application state.
package store

import (
	"bikerental/web"
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
	nextBike   uint64
	nextRental uint64
	demo       bool
	people     map[string]domain.Person
	bikes      map[string]domain.Bike
	rentals    map[string]domain.Rental
}

// Snapshot contains independent value copies in stable ID order.
type Snapshot struct {
	People  []domain.Person
	Bikes   []domain.Bike
	Rentals []domain.Rental
}

func New(demo bool) *Store {
	s := &Store{
		demo:    demo,
		now:     time.Now,
		tariff:  domain.Tariff{Types: map[string]domain.Rate{}},
		people:  make(map[string]domain.Person),
		bikes:   make(map[string]domain.Bike),
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
		s.tariff = domain.Tariff{Version: 1, Types: map[string]domain.Rate{
			"Downhill": {HourlyCents: 1200, PricesCents: []int64{4000, 7000, 9500}, ExtraDayCents: 2500},
			"MTB":      {HourlyCents: 800, PricesCents: []int64{2500, 4500, 6000}, ExtraDayCents: 1500},
			"Rennrad":  {HourlyCents: 1000, PricesCents: []int64{3500, 6000, 8000}, ExtraDayCents: 2000},
		}}
	}
	person := domain.Person{
		ID: "demo-person-01", Name: "Demo Person",
		DocumentType: "Sonstiger Ausweis", DocumentDescription: "Demoausweis",
		DocumentNumber: "DEMO-0001",
	}
	if _, exists := s.people[person.ID]; !exists {
		s.people[person.ID] = person
	}
	for i, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		id := fmt.Sprintf("demo-bike-%02d", i+1)
		if _, exists := s.bikes[id]; !exists {
			photo, err := web.Files.ReadFile(fmt.Sprintf("assets/demo/%s.png", kind))
			if err != nil {
				panic(fmt.Sprintf("Eingebettetes Demobild fehlt: %v", err))
			}
			s.bikes[id] = domain.Bike{ID: id, Name: "Demo " + kind, Type: kind, Photo: photo, PhotoType: "image/png"}
		}
	}
}

func (s *Store) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := Snapshot{
		People:  make([]domain.Person, 0, len(s.people)),
		Bikes:   make([]domain.Bike, 0, len(s.bikes)),
		Rentals: make([]domain.Rental, 0, len(s.rentals)),
	}
	for _, person := range s.people {
		result.People = append(result.People, clonePerson(person))
	}
	for _, bike := range s.bikes {
		result.Bikes = append(result.Bikes, cloneBike(bike))
	}
	for _, rental := range s.rentals {
		result.Rentals = append(result.Rentals, rental)
	}
	sort.Slice(result.People, func(i, j int) bool { return result.People[i].ID < result.People[j].ID })
	sort.Slice(result.Bikes, func(i, j int) bool { return result.Bikes[i].ID < result.Bikes[j].ID })
	sort.Slice(result.Rentals, func(i, j int) bool { return result.Rentals[i].ID < result.Rentals[j].ID })
	return result
}

var ErrNotFound = errors.New("Datensatz nicht gefunden")

func cloneBike(b domain.Bike) domain.Bike {
	b.Photo = bytes.Clone(b.Photo)
	b.PhotoURL = "/api/fahrraeder/" + b.ID + "/foto"
	return b
}

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
func (s *Store) Bike(id string) (domain.Bike, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bikes[id]
	if !ok {
		return domain.Bike{}, ErrNotFound
	}
	return cloneBike(b), nil
}

// SaveBike validates and replaces all fields atomically, retaining the old image
// only when explicitly requested without an accompanying upload.
func (s *Store) SaveBike(id string, b domain.Bike, keepPhoto bool) (domain.Bike, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, exists := s.bikes[id]
	if id != "" && !exists {
		return domain.Bike{}, ErrNotFound
	}
	if keepPhoto {
		if !exists || len(b.Photo) > 0 || b.PhotoType != "" {
			return domain.Bike{}, domain.ValidationErrors{"image": "Ein vorhandenes Bild kann nur ohne neue Datei beibehalten werden."}
		}
		b.Photo, b.PhotoType = old.Photo, old.PhotoType
	}
	b, err := domain.ValidateBike(b)
	if err != nil {
		return domain.Bike{}, err
	}
	if id == "" {
		s.nextBike++
		id = fmt.Sprintf("bike-%06d", s.nextBike)
	}
	b.ID = id
	s.bikes[id] = b
	return cloneBike(b), nil
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
func (s *Store) Bikes() []domain.Bike {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]domain.Bike, 0, len(s.bikes))
	for _, p := range s.bikes {
		result = append(result, cloneBike(p))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
