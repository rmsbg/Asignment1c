package store

import (
	"bikerental/internal/domain"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestRentalLifecycle(t *testing.T) {
	s := New(true)
	before := s.Snapshot()
	for _, ids := range [][2]string{{"", ""}, {"missing", "demo-ski-01"}, {"demo-person-01", "missing"}} {
		if _, err := createIssued(s, ids[0], ids[1]); err == nil {
			t.Fatal("invalid rental accepted")
		}
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("failed request changed state")
		}
	}
	start := time.Now().UTC()
	first, err := createIssued(s, "demo-person-01", "demo-ski-01")
	if err != nil {
		t.Fatal(err)
	}
	r := first.Rental
	if r.ID == "" || !r.Active() || r.RentedAt.Before(start) || r.RentedAt.After(time.Now().UTC()) || r.RentedAt.Location() != time.UTC {
		t.Fatal("invalid generated rental fields")
	}
	if len(s.AvailableSkis()) != 1 {
		t.Fatal("rented pair still available")
	}
	before = s.Snapshot()
	if _, err := createIssued(s, "demo-person-01", "demo-ski-01"); !errors.Is(err, ErrUnavailable) {
		t.Fatal("double rental accepted")
	}
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("conflict mutated state")
	}
	second, err := createIssued(s, "demo-person-01", "demo-ski-02")
	if err != nil || second.Rental.ID == r.ID || len(s.AvailableSkis()) != 0 {
		t.Fatal("multiple rentals per person failed")
	}
	_, _ = s.SaveSki("demo-ski-01", "Renamed Ski")
	_, _ = s.SavePerson("demo-person-01", domain.Person{Name: "Renamed Person"}, true)
	details, err := s.Rental(r.ID)
	if err != nil || details.PersonName != "Renamed Person" || details.SkiName != "Renamed Ski" || details.Rental.PersonID != r.PersonID {
		t.Fatal("names or references inconsistent")
	}
	returned, err := s.ReturnRental(r.ID)
	if err != nil || returned.Rental.Active() || returned.Rental.ReturnedAt.Before(r.RentedAt) || len(s.AvailableSkis()) != 1 {
		t.Fatal("return failed")
	}
	again, err := s.ReturnRental(r.ID)
	if err != nil || !reflect.DeepEqual(returned, again) {
		t.Fatal("repeated return changed timestamp")
	}
	replacement, err := createIssued(s, "demo-person-01", "demo-ski-01")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = s.ReturnRental(r.ID)
	current, _ := s.Rental(replacement.Rental.ID)
	if !current.Rental.Active() || len(s.AvailableSkis()) != 0 {
		t.Fatal("old return released a later rental")
	}
	list := s.Rentals()
	list[0].Rental.PersonID = "mutated"
	list[0].PersonName = "mutated"
	snap := s.Snapshot()
	snap.Rentals[0].ID = "mutated"
	got, _ := s.Rental(r.ID)
	if got.Rental.PersonID != r.PersonID || got.PersonName != "Renamed Person" {
		t.Fatal("rental read aliased")
	}
	_, _ = s.ReturnRental(second.Rental.ID)
	available := s.AvailableSkis()
	available[0].Name = "mutated"
	if s.AvailableSkis()[0].Name == "mutated" {
		t.Fatal("availability aliased")
	}
	if _, err := s.Rental("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := s.ReturnRental("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if len(New(true).Rentals()) != 0 || len(New(false).Rentals()) != 0 {
		t.Fatal("new store inherited rentals")
	}
}

func TestConcurrentRentalAndReturn(t *testing.T) {
	s := New(true)
	var wg sync.WaitGroup
	results := make(chan RentalDetails, 32)
	failures := make(chan error, 32)
	for range 32 {
		wg.Go(func() {
			r, err := createIssued(s, "demo-person-01", "demo-ski-01")
			if err != nil {
				failures <- err
			} else {
				results <- r
			}
			_ = s.AvailableSkis()
			_ = s.Rentals()
		})
	}
	wg.Wait()
	close(results)
	close(failures)
	if len(results) != 1 || len(failures) != 31 {
		t.Fatalf("successes=%d failures=%d", len(results), len(failures))
	}
	for err := range failures {
		if !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
	}
	first := <-results
	returns := make(chan RentalDetails, 32)
	for range 32 {
		wg.Go(func() {
			r, err := s.ReturnRental(first.Rental.ID)
			if err != nil {
				t.Error(err)
				return
			}
			returns <- r
			_ = s.AvailableSkis()
		})
	}
	wg.Wait()
	close(returns)
	var returned time.Time
	for r := range returns {
		if returned.IsZero() {
			returned = r.Rental.ReturnedAt
		}
		if r.Rental.ReturnedAt.IsZero() || !r.Rental.ReturnedAt.Equal(returned) {
			t.Fatal("inconsistent return timestamp")
		}
	}
	if len(s.AvailableSkis()) != 2 {
		t.Fatal("return did not restore availability")
	}
	// Independent pairs may be rented concurrently, all IDs must remain unique.
	ids := make(chan string, 24)
	for range 24 {
		wg.Go(func() {
			ski, err := s.SaveSki("", "Parallel Ski")
			if err != nil {
				t.Error(err)
				return
			}
			r, err := createIssued(s, "demo-person-01", ski.ID)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- r.Rental.ID
		})
	}
	wg.Wait()
	close(ids)
	seen := map[string]bool{first.Rental.ID: true}
	for id := range ids {
		if seen[id] {
			t.Fatal("duplicate rental ID")
		}
		seen[id] = true
	}
	if len(seen) != 25 {
		t.Fatal("lost rentals")
	}
}
