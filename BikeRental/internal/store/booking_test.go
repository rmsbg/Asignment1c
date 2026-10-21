package store

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"bikerental/internal/domain"
)

func datedStore() *Store {
	s := New(true)
	s.now = func() time.Time { return time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC) }
	return s
}
func book(t *testing.T, s *Store, bike, start, end string) RentalDetails {
	t.Helper()
	r, err := s.CreateRental("demo-person-01", bike, start, end, s.Tariff().Version)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func TestBookingOverlapAndLifecycle(t *testing.T) {
	s := datedStore()
	first := book(t, s, "demo-bike-01", "2026-10-21", "2026-10-23")
	if first.Rental.Active() || first.Rental.TotalPriceCents != 6000 || first.Rental.Status(first.Today) != "booked" {
		t.Fatal("wrong booking")
	}
	assertCounts(t, s, DashboardCounts{AvailableBikes: 3, TotalBikes: 3, RegisteredPeople: 1})
	available, _ := s.AvailableFor("2026-10-23", "2026-10-24")
	if len(available) != 2 {
		t.Fatal("inclusive availability failed")
	}
	before := s.Snapshot()
	for _, period := range [][2]string{{"2026-10-21", "2026-10-21"}, {"2026-10-22", "2026-10-24"}, {"2026-10-23", "2026-10-23"}, {"2026-10-21", "2026-10-30"}} {
		if _, err := s.CreateRental("demo-person-01", "demo-bike-01", period[0], period[1], 1); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("overlap accepted: %v", err)
		}
	}
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("conflict changed state")
	}
	if _, err := s.ReturnRental(first.Rental.ID); err == nil {
		t.Fatal("return before issue accepted")
	}
	issued, err := s.IssueRental(first.Rental.ID)
	if err != nil || !issued.Rental.Active() {
		t.Fatal(err)
	}
	repeated, err := s.IssueRental(first.Rental.ID)
	if err != nil || !issued.Rental.RentedAt.Equal(repeated.Rental.RentedAt) {
		t.Fatal("issue is not idempotent")
	}
	if _, err := s.CancelRental(first.Rental.ID); err == nil {
		t.Fatal("issued booking cancelled")
	}
	assertCounts(t, s, DashboardCounts{AvailableBikes: 2, RentedBikes: 1, TotalBikes: 3, RegisteredPeople: 1})
	next := book(t, s, "demo-bike-01", "2026-10-24", "2026-10-25")
	if _, err := s.IssueRental(next.Rental.ID); err == nil {
		t.Fatal("early issue accepted")
	}
	// The previous pair is still out, even beyond its scheduled end.
	s.now = func() time.Time { return time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC) }
	if _, err := s.IssueRental(next.Rental.ID); err == nil {
		t.Fatal("late predecessor did not block issue")
	}
	returned, err := s.ReturnRental(first.Rental.ID)
	if err != nil || returned.Rental.TotalPriceCents != 6000 {
		t.Fatal("late return changed price")
	}
	actual, err := s.IssueRental(next.Rental.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = s.ReturnRental(first.Rental.ID)
	if len(s.AvailableBikes()) != 2 {
		t.Fatal("old return released new issue")
	}
	_, _ = s.ReturnRental(next.Rental.ID)
	again, err := s.IssueRental(next.Rental.ID)
	if err != nil || !actual.Rental.RentedAt.Equal(again.Rental.RentedAt) {
		t.Fatal("repeated issue after return changed timestamp")
	}
}
func TestCancellationExpiryAndEarlyReturn(t *testing.T) {
	s := datedStore()
	r := book(t, s, "demo-bike-01", "2026-10-21", "2026-10-22")
	cancelled, err := s.CancelRental(r.Rental.ID)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.CancelRental(r.Rental.ID)
	if err != nil || !reflect.DeepEqual(cancelled, again) {
		t.Fatal("cancel not idempotent")
	}
	if _, err := s.IssueRental(r.Rental.ID); err == nil {
		t.Fatal("cancelled booking issued")
	}
	replacement := book(t, s, "demo-bike-01", "2026-10-21", "2026-10-22")
	if _, err := s.IssueRental(replacement.Rental.ID); err != nil {
		t.Fatal(err)
	}
	returned, _ := s.ReturnRental(replacement.Rental.ID)
	if returned.Rental.TotalPriceCents != 4500 {
		t.Fatal("early return changed price")
	}
	expired := book(t, s, "demo-bike-01", "2026-10-21", "2026-10-21")
	s.now = func() time.Time { return time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) }
	detail, _ := s.Rental(expired.Rental.ID)
	if detail.Rental.Status(detail.Today) != "expired" {
		t.Fatal("booking did not expire")
	}
	if _, err := s.IssueRental(expired.Rental.ID); err == nil {
		t.Fatal("expired booking issued")
	}
	if _, err := s.CancelRental(expired.Rental.ID); err != nil {
		t.Fatal(err)
	}
}
func TestTariffSnapshotsAndConflicts(t *testing.T) {
	s := datedStore()
	r := book(t, s, "demo-bike-01", "2026-10-21", "2026-10-24")
	tariff := s.Tariff()
	tariff.PricesCents[0] = 999
	if s.Tariff().PricesCents[0] != 2500 {
		t.Fatal("tariff aliased")
	}
	saved, err := s.SaveTariff(tariff)
	if err != nil || saved.Version != 2 {
		t.Fatal(err)
	}
	saved.PricesCents[0] = 777
	if s.Tariff().PricesCents[0] != 999 {
		t.Fatal("saved tariff aliased")
	}
	s.SeedDemo()
	if s.Tariff().Version != 2 {
		t.Fatal("seeding overwrote tariff")
	}
	before := s.Snapshot()
	oldTariff := s.Tariff()
	if _, err := s.CreateRental("demo-person-01", "demo-bike-02", "2026-10-21", "2026-10-21", 1); err == nil {
		t.Fatal("stale price accepted")
	}
	if _, err := s.SaveTariff(tariff); err == nil {
		t.Fatal("stale tariff edit accepted")
	}
	if !reflect.DeepEqual(before, s.Snapshot()) || !reflect.DeepEqual(oldTariff, s.Tariff()) {
		t.Fatal("conflict mutated state")
	}
	existing, _ := s.Rental(r.Rental.ID)
	if existing.Rental.TotalPriceCents != 7500 || existing.Rental.TariffVersion != 1 {
		t.Fatal("stored quote changed")
	}
	normal := New(false)
	if normal.Tariff().Version != 0 {
		t.Fatal("normal mode inherited tariff")
	}
	if _, err := normal.Quote("9999-01-01", "9999-01-01"); err == nil {
		t.Fatal("missing tariff quoted")
	}
}
func TestConcurrentBookingsAndTariffEdits(t *testing.T) {
	s := datedStore()
	var wg sync.WaitGroup
	results := make(chan error, 32)
	for range 32 {
		wg.Go(func() {
			_, err := s.CreateRental("demo-person-01", "demo-bike-01", "2026-10-21", "2026-10-23", 1)
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
		t.Fatalf("success=%d", success)
	}
	edits := make(chan error, 16)
	for range 16 {
		wg.Go(func() {
			_, err := s.SaveTariff(domain.Tariff{Version: 1, PricesCents: []int64{100}, ExtraDayCents: 50})
			edits <- err
		})
	}
	wg.Wait()
	close(edits)
	success = 0
	for err := range edits {
		if err == nil {
			success++
		}
	}
	if success != 1 || s.Tariff().Version != 2 {
		t.Fatal("tariff update lost concurrency control")
	}
}

func TestAtomicPriceSnapshotDuringEdits(t *testing.T) {
	s := datedStore()
	// Every tariff version has a distinct, predictable daily price.
	tariff := s.Tariff()
	tariff.PricesCents = []int64{200}
	tariff.ExtraDayCents = 0
	if _, err := s.SaveTariff(tariff); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 50 {
			current := s.Tariff()
			current.PricesCents = []int64{int64(current.Version+1) * 100}
			if _, err := s.SaveTariff(current); err != nil {
				t.Error(err)
			}
		}
	})
	for range 24 {
		wg.Go(func() {
			bike, err := saveBike(s, "", "Price test pair")
			if err != nil {
				t.Error(err)
				return
			}
			for {
				q, err := s.Quote("2026-10-21", "2026-10-21")
				if err != nil {
					t.Error(err)
					return
				}
				r, err := s.CreateRental("demo-person-01", bike.ID, "2026-10-21", "2026-10-21", q.TariffVersion)
				if err != nil {
					var conflict Conflict
					if errors.As(err, &conflict) {
						continue
					}
					t.Error(err)
					return
				}
				if r.Rental.TotalPriceCents != int64(r.Rental.TariffVersion)*100 {
					t.Error("mixed tariff snapshot")
				}
				return
			}
		})
	}
	wg.Wait()
	if len(s.Rentals()) != 24 {
		t.Fatal("lost bookings")
	}
}
