package store

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"bikerental/internal/domain"
)

func hourInput(bike, kind, start, end string) domain.BookingInput {
	return domain.BookingInput{PersonID: "demo-person-01", BikeID: bike, BikeType: kind, Mode: "hour", StartsOn: start, EndsOn: end, TariffVersion: 1}
}

func TestHourlySnapshotsAndMixedBoundaries(t *testing.T) {
	s := datedStore()
	s.now = func() time.Time { return time.Date(2026, 10, 21, 8, 0, 30, 0, time.UTC) } // 10:00 Vienna
	for _, tc := range []struct {
		kind, bike string
		hour, day  int64
	}{
		{"Downhill", "demo-bike-01", 1200, 4000}, {"MTB", "demo-bike-02", 800, 2500}, {"Rennrad", "demo-bike-03", 1000, 3500},
	} {
		q, err := s.Quote(tc.kind, "hour", "2026-10-21T10:00", "2026-10-21T11:15")
		if err != nil || q.TotalPriceCents != 2*tc.hour || q.DurationHours != 2 {
			t.Fatalf("hour quote %+v %v", q, err)
		}
		q, err = s.Quote(tc.kind, "day", "2026-10-21", "2026-10-21")
		if err != nil || q.TotalPriceCents != tc.day || q.DurationDays != 1 {
			t.Fatalf("day quote %+v %v", q, err)
		}
	}
	input := hourInput("demo-bike-01", "Downhill", "2026-10-21T10:00", "2026-10-21T11:15")
	r, err := s.CreateRental(input)
	if err != nil {
		t.Fatal(err)
	}
	if r.Rental.TotalPriceCents != 2400 || r.Rental.BikeType != "Downhill" {
		t.Fatal("wrong snapshot")
	}
	if _, err := createDay(s, input.PersonID, input.BikeID, "2026-10-21", "2026-10-21", 1); !errors.Is(err, ErrUnavailable) {
		t.Fatal("day over hour accepted", err)
	}
	if _, err := s.CreateRental(hourInput(input.BikeID, input.BikeType, "2026-10-21T11:14", "2026-10-21T12:00")); !errors.Is(err, ErrUnavailable) {
		t.Fatal("hour overlap accepted", err)
	}
	next, err := s.CreateRental(hourInput(input.BikeID, input.BikeType, "2026-10-21T11:15", "2026-10-21T12:00"))
	if err != nil {
		t.Fatal("exact non-rounded boundary rejected", err)
	}
	if _, err := s.IssueRental(next.Rental.ID); err == nil {
		t.Fatal("issued before start")
	}
	if _, err := s.IssueRental(r.Rental.ID); err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return time.Date(2026, 10, 21, 9, 15, 0, 0, time.UTC) }
	if _, err := s.IssueRental(next.Rental.ID); err == nil {
		t.Fatal("late predecessor did not block issue")
	}
	q := s.Tariff()
	newRate := q.Types["Downhill"]
	newRate.HourlyCents = 9999
	q.Types["Downhill"] = newRate
	if _, err := s.SaveTariff(q); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ReturnRental(r.Rental.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.IssueRental(next.Rental.ID); err != nil {
		t.Fatal(err)
	}
	returned, _ := s.Rental(r.Rental.ID)
	if returned.Rental.TotalPriceCents != 2400 || returned.Rental.DurationHours != 2 || returned.Rental.TariffVersion != 1 {
		t.Fatal("actual return or tariff changed snapshot")
	}
	if _, err := s.ReturnRental(r.Rental.ID); err != nil {
		t.Fatal(err)
	}
	current, _ := s.Rental(next.Rental.ID)
	if !current.Rental.Active() {
		t.Fatal("old return affected new issue")
	}
}

func TestDayToHourAndExactExpiry(t *testing.T) {
	s := datedStore()
	r := book(t, s, "demo-bike-01", "2026-10-21", "2026-10-21")
	if _, err := s.CreateRental(hourInput("demo-bike-01", "Downhill", "2026-10-21T23:59", "2026-10-22T01:00")); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	next, err := s.CreateRental(hourInput("demo-bike-01", "Downhill", "2026-10-22T00:00", "2026-10-22T01:00"))
	if err != nil {
		t.Fatal("midnight handoff rejected", err)
	}
	s.now = func() time.Time { return time.Date(2026, 10, 21, 22, 0, 0, 0, time.UTC) }
	old, _ := s.Rental(r.Rental.ID)
	if old.Rental.Status(old.Now) != "expired" {
		t.Fatal("not expired at exclusive end")
	}
	if _, err := s.IssueRental(r.Rental.ID); err == nil {
		t.Fatal("expired booking issued")
	}
	if _, err := s.IssueRental(next.Rental.ID); err != nil {
		t.Fatal(err)
	}
}

func TestTypeConflictAndConcurrentMixedBookings(t *testing.T) {
	s := datedStore()
	before := s.Snapshot()
	bad := hourInput("demo-bike-01", "MTB", "2026-10-21T11:00", "2026-10-21T12:00")
	var conflict Conflict
	if _, err := s.CreateRental(bad); !errors.As(err, &conflict) || conflict.Field != "bikeType" {
		t.Fatal("stale type accepted", err)
	}
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("conflict changed state")
	}
	var wg sync.WaitGroup
	results := make(chan error, 32)
	for i := range 32 {
		wg.Go(func() {
			var err error
			if i%2 == 0 {
				_, err = s.CreateRental(hourInput("demo-bike-01", "Downhill", "2026-10-21T11:00", "2026-10-21T12:00"))
			} else {
				_, err = createDay(s, "demo-person-01", "demo-bike-01", "2026-10-21", "2026-10-21", 1)
			}
			results <- err
		})
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("got %d successful mixed bookings", success)
	}
}

func TestInventoryTypeChangePreservesAgreedPrice(t *testing.T) {
	s := datedStore()
	booked, err := s.CreateRental(hourInput("demo-bike-01", "Downhill", "2026-10-21T11:00", "2026-10-21T12:00"))
	if err != nil {
		t.Fatal(err)
	}
	// A preview belongs to the selected type, even if the inventory is edited later.
	_, err = s.Quote("Downhill", "hour", "2026-10-22T11:00", "2026-10-22T12:00")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SaveBike("demo-bike-01", domain.Bike{Name: "Changed category", Type: "MTB"}, true); err != nil {
		t.Fatal(err)
	}
	before := s.Snapshot()
	if _, err := s.CreateRental(hourInput("demo-bike-01", "Downhill", "2026-10-22T11:00", "2026-10-22T12:00")); err == nil {
		t.Fatal("stale preview accepted after inventory change")
	}
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("stale preview modified state")
	}
	old, _ := s.Rental(booked.Rental.ID)
	if old.Rental.BikeType != "Downhill" || old.Rental.TotalPriceCents != 1200 {
		t.Fatal("inventory edit changed existing price or category")
	}
	newBooking, err := s.CreateRental(hourInput("demo-bike-01", "MTB", "2026-10-22T11:00", "2026-10-22T12:00"))
	if err != nil || newBooking.Rental.TotalPriceCents != 800 {
		t.Fatal("fresh type confirmation did not use current rate", err)
	}
	copy := s.Tariff()
	delete(copy.Types, "Downhill")
	if _, ok := s.Tariff().Types["Downhill"]; !ok {
		t.Fatal("tariff map aliases stored state")
	}
}
