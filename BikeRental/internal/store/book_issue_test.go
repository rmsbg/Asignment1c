package store

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"bikerental/internal/domain"
)

func TestCreateAndIssueRental(t *testing.T) {
	for _, mode := range []string{"day", "hour"} {
		t.Run(mode, func(t *testing.T) {
			s := datedStore()
			input := immediateInput()
			if mode == "day" {
				input.Mode = "day"
				input.StartMode = "scheduled"
				input.RentalHours = 0
				input.StartsOn = "2026-10-21"
				input.EndsOn = "2026-10-21"
			}
			q, err := s.QuoteWithStart(input.BikeType, input.Mode, input.StartsOn, input.EndsOn, input.Extras, domain.StartOptions{StartMode: input.StartMode, RentalHours: input.RentalHours})
			if err != nil {
				t.Fatal(err)
			}
			r, err := s.CreateAndIssueRental(input)
			if err != nil {
				t.Fatal(err)
			}
			if !r.Rental.Active() || r.Rental.RentedAt != r.Rental.CreatedAt || r.Rental.TotalPriceCents != q.TotalPriceCents {
				t.Fatalf("bad handover: %+v", r)
			}
			if mode == "hour" && (r.Rental.StartsAt != r.Rental.RentedAt || r.Rental.EndsAt.Sub(r.Rental.StartsAt) != time.Hour) {
				t.Fatal("immediate period shifted")
			}
			assertCounts(t, s, DashboardCounts{AvailableBikes: 2, RentedBikes: 1, TotalBikes: 3, RegisteredPeople: 1})
		})
	}
}

func TestCreateAndIssueConflictsLeaveNoBooking(t *testing.T) {
	for _, cause := range []string{"future-day", "future-hour", "tariff", "overlap", "overdue", "type"} {
		t.Run(cause, func(t *testing.T) {
			s := datedStore()
			input := immediateInput()
			switch cause {
			case "future-day":
				input.Mode = "day"
				input.StartMode = "scheduled"
				input.RentalHours = 0
				input.StartsOn = "2026-10-22"
				input.EndsOn = "2026-10-22"
			case "future-hour":
				input.StartMode = "scheduled"
				input.RentalHours = 0
				input.StartsOn = "2026-10-21T12:00"
				input.EndsOn = "2026-10-21T13:00"
			case "tariff":
				tariff := s.Tariff()
				tariff.Extras.HelmetCents++
				if _, err := s.SaveTariff(tariff); err != nil {
					t.Fatal(err)
				}
			case "overlap":
				if _, err := s.CreateRental(input); err != nil {
					t.Fatal(err)
				}
			case "overdue":
				if _, err := s.CreateAndIssueRental(input); err != nil {
					t.Fatal(err)
				}
				now := s.now().Add(2 * time.Hour)
				s.now = func() time.Time { return now }
			case "type":
				input.BikeType = "Downhill"
			}
			before := s.Snapshot()
			next := s.nextRental
			if _, err := s.CreateAndIssueRental(input); err == nil {
				t.Fatal("conflict accepted")
			}
			if !reflect.DeepEqual(before, s.Snapshot()) || next != s.nextRental {
				t.Fatal("partial booking on failure")
			}
		})
	}
}

func TestConcurrentCreateAndIssue(t *testing.T) {
	s := datedStore()
	var wg sync.WaitGroup
	results := make(chan error, 32)
	for range 32 {
		wg.Go(func() { _, err := s.CreateAndIssueRental(immediateInput()); results <- err })
	}
	wg.Wait()
	close(results)
	success := 0
	for err := range results {
		if err == nil {
			success++
		}
	}
	if success != 1 || len(s.Rentals()) != 1 || !s.Rentals()[0].Rental.Active() {
		t.Fatal("concurrent issue not atomic")
	}
}

func TestCreateAndIssueViennaDayBoundary(t *testing.T) {
	s := datedStore()
	s.now = func() time.Time { return time.Date(2026, 10, 21, 22, 30, 0, 0, time.UTC) }
	input := immediateInput()
	input.Mode = "day"
	input.StartMode = "scheduled"
	input.RentalHours = 0
	input.StartsOn = "2026-10-22"
	input.EndsOn = "2026-10-22"
	if _, err := s.CreateAndIssueRental(input); err != nil {
		t.Fatal(err)
	}
}
