package store

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"bikerental/internal/domain"
)

func immediateInput() domain.BookingInput {
	return domain.BookingInput{PersonID: "demo-person-01", BikeID: "demo-bike-02", BikeType: "MTB", Mode: "hour", StartMode: "now", RentalHours: 1, TariffVersion: 1, Extras: domain.Extras{Helmet: true}}
}

func TestImmediateBookingAfterDelayedConfirmation(t *testing.T) {
	s := New(true)
	now := time.Date(2026, 10, 21, 8, 0, 25, 123, time.UTC)
	s.now = func() time.Time { return now }
	options := domain.StartOptions{StartMode: "now", RentalHours: 1}
	quote, err := s.QuoteWithStart("MTB", "hour", "", "", domain.Extras{Helmet: true}, options)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(7*time.Minute + 13*time.Second)
	booked, err := s.CreateRental(immediateInput())
	if err != nil {
		t.Fatal(err)
	}
	r := booked.Rental
	if !r.StartsAt.Equal(now) || !r.CreatedAt.Equal(now) || !r.EndsAt.Equal(now.Add(time.Hour)) || r.StartMode != "now" || r.TotalPriceCents != 1300 || r.TotalPriceCents != quote.TotalPriceCents {
		t.Fatalf("wrong agreement after delay: %+v", r)
	}
	if s.Dashboard().RentedBikes != 0 {
		t.Fatal("booked bike counted as issued")
	}
	now = now.Add(2 * time.Minute)
	issued, err := s.IssueRental(r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if issued.Rental.StartsAt != r.StartsAt || issued.Rental.EndsAt != r.EndsAt || issued.Rental.TotalPriceCents != r.TotalPriceCents || !issued.Rental.RentedAt.Equal(now) {
		t.Fatal("issue changed the agreed period or price")
	}
}

func TestImmediateShiftRechecksAvailabilityAndTariff(t *testing.T) {
	for _, cause := range []string{"overlap", "tariff"} {
		t.Run(cause, func(t *testing.T) {
			s := New(true)
			now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
			s.now = func() time.Time { return now }
			options := domain.StartOptions{StartMode: "now", RentalHours: 1}
			if cause == "overlap" {
				future := immediateInput()
				future.StartMode = "scheduled"
				future.RentalHours = 0
				future.StartsOn = "2026-10-21T11:03"
				future.EndsOn = "2026-10-21T12:00"
				if _, err := s.CreateRental(future); err != nil {
					t.Fatal(err)
				}
			}
			bikes, err := s.AvailableWithStart("hour", "", "", options)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, b := range bikes {
				found = found || b.ID == "demo-bike-02"
			}
			if !found {
				t.Fatal("initial hour should be available")
			}
			if _, err := s.QuoteWithStart("MTB", "hour", "", "", domain.Extras{Helmet: true}, options); err != nil {
				t.Fatal(err)
			}
			now = now.Add(5 * time.Minute)
			if cause == "tariff" {
				tariff := s.Tariff()
				tariff.Extras.HelmetCents++
				if _, err := s.SaveTariff(tariff); err != nil {
					t.Fatal(err)
				}
			}
			before := s.Snapshot()
			if _, err := s.CreateRental(immediateInput()); err == nil {
				t.Fatal("changed availability/tariff must require a new confirmation")
			}
			if !reflect.DeepEqual(before, s.Snapshot()) {
				t.Fatal("failed immediate booking changed state")
			}
		})
	}
}

func TestConcurrentImmediateBookings(t *testing.T) {
	s := New(true)
	now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	var wg sync.WaitGroup
	results := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := s.CreateRental(immediateInput()); results <- err }()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrUnavailable) {
			t.Fatal(err)
		}
	}
	if successes != 1 || len(s.Rentals()) != 1 {
		t.Fatalf("got %d successful overlapping bookings", successes)
	}
}
