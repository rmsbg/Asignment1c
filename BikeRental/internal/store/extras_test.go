package store

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"bikerental/internal/domain"
)

func TestExtrasAllCombinationsForEveryTypeAndMode(t *testing.T) {
	wants := []int64{0, 500, 1000, 1500, 800, 1300, 1800, 2300}
	for _, tc := range []struct {
		kind, bike string
		day, hour  int64
	}{
		{"Downhill", "demo-bike-01", 7000, 2400}, {"MTB", "demo-bike-02", 4500, 1600}, {"Rennrad", "demo-bike-03", 6000, 2000},
	} {
		for _, mode := range []string{"day", "hour"} {
			for mask, want := range wants {
				s := datedStore()
				selection := domain.Extras{Helmet: mask&1 != 0, Protection: mask&2 != 0, Insurance: mask&4 != 0}
				start, end, base := "2026-10-22", "2026-10-23", tc.day
				if mode == "hour" {
					start, end, base = "2026-10-22T10:00", "2026-10-22T11:15", tc.hour
				}
				quote, err := s.Quote(tc.kind, mode, start, end, selection)
				if err != nil || quote.BasePriceCents != base || quote.ExtrasPriceCents != want || quote.TotalPriceCents != base+want || quote.Extras != selection {
					t.Fatalf("%s %s %d: quote=%+v err=%v", tc.kind, mode, mask, quote, err)
				}
				r, err := s.CreateRental(domain.BookingInput{PersonID: "demo-person-01", BikeID: tc.bike, BikeType: tc.kind, Mode: mode, StartsOn: start, EndsOn: end, TariffVersion: 1, Extras: selection})
				if err != nil || r.Rental.BasePriceCents != base || r.Rental.ExtrasPriceCents != want || r.Rental.TotalPriceCents != base+want || r.Rental.Extras != selection || r.Rental.ExtraPrices != quote.ExtraPrices {
					t.Fatal("booking did not preserve quote", err)
				}
				copy := s.Snapshot()
				copy.Rentals[0].Extras.Helmet = !selection.Helmet
				copy.Rentals[0].ExtraPrices.HelmetCents = -1
				got, _ := s.Rental(r.Rental.ID)
				if got.Rental.Extras != selection || got.Rental.ExtraPrices != quote.ExtraPrices {
					t.Fatal("snapshot changed stored extras")
				}
			}
		}
	}
}

func TestExtraPriceChangeRequiresConfirmationAndPreservesHistory(t *testing.T) {
	s := datedStore()
	input := domain.BookingInput{PersonID: "demo-person-01", BikeID: "demo-bike-01", BikeType: "Downhill", Mode: "day", StartsOn: "2026-10-21", EndsOn: "2026-10-21", TariffVersion: 1, Extras: domain.Extras{Helmet: true, Protection: true, Insurance: true}}
	r, err := s.CreateRental(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.IssueRental(r.Rental.ID); err != nil {
		t.Fatal(err)
	}
	rates := s.Tariff()
	rates.Extras.HelmetCents = 900
	if _, err := s.SaveTariff(rates); err != nil {
		t.Fatal(err)
	}
	input.BikeID = "demo-bike-02"
	input.BikeType = "MTB"
	before := s.Snapshot()
	var conflict Conflict
	if _, err := s.CreateRental(input); !errors.As(err, &conflict) || conflict.Field != "tariffVersion" {
		t.Fatal("stale extra fee accepted", err)
	}
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("failed confirmation created partial booking")
	}
	input.TariffVersion = 2
	updated, err := s.CreateRental(input)
	if err != nil || updated.Rental.TotalPriceCents != 5200 || updated.Rental.ExtrasPriceCents != 2700 {
		t.Fatal("confirmed extras incorrect", err)
	}
	s.now = func() time.Time { return time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) }
	returned, err := s.ReturnRental(r.Rental.ID)
	if err != nil || returned.Rental.BasePriceCents != 4000 || returned.Rental.ExtrasPriceCents != 2300 || returned.Rental.TotalPriceCents != 6300 || returned.Rental.ExtraPrices.HelmetCents != 500 {
		t.Fatal("late return or fee change altered agreement", err)
	}
	again, _ := s.ReturnRental(r.Rental.ID)
	if !reflect.DeepEqual(returned.Rental, again.Rental) {
		t.Fatal("repeated return changed fees")
	}
	cancelled, err := s.CancelRental(updated.Rental.ID)
	if err != nil || cancelled.Rental.TotalPriceCents != 5200 || cancelled.Rental.Extras != input.Extras {
		t.Fatal("cancellation lost agreement", err)
	}
}

func TestExtrasAtomicDuringConcurrentFeeEdits(t *testing.T) {
	s := datedStore()
	tariff := s.Tariff()
	tariff.Extras.HelmetCents = 200
	if _, err := s.SaveTariff(tariff); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 40 {
			current := s.Tariff()
			current.Extras.HelmetCents = int64(current.Version+1) * 100
			if _, err := s.SaveTariff(current); err != nil {
				t.Error(err)
			}
		}
	})
	for range 16 {
		wg.Go(func() {
			bike, err := saveBike(s, "", "Concurrent fee test")
			if err != nil {
				t.Error(err)
				return
			}
			for {
				choice := domain.Extras{Helmet: true}
				q, err := s.Quote("MTB", "day", "2026-10-21", "2026-10-21", choice)
				if err != nil {
					t.Error(err)
					return
				}
				r, err := s.CreateRental(domain.BookingInput{PersonID: "demo-person-01", BikeID: bike.ID, BikeType: "MTB", Mode: "day", StartsOn: "2026-10-21", EndsOn: "2026-10-21", TariffVersion: q.TariffVersion, Extras: choice})
				if err != nil {
					var c Conflict
					if errors.As(err, &c) {
						continue
					}
					t.Error(err)
					return
				}
				if r.Rental.BasePriceCents != 2500 || r.Rental.ExtrasPriceCents != int64(r.Rental.TariffVersion)*100 || r.Rental.TotalPriceCents != 2500+r.Rental.ExtrasPriceCents {
					t.Error("mixed fee snapshot")
				}
				return
			}
		})
	}
	wg.Wait()
	if len(s.Rentals()) != 16 {
		t.Fatal("lost concurrent bookings")
	}
}
