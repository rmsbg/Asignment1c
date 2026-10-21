package handler

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func TestHTTPAllExtraSelectionsForEveryModeAndType(t *testing.T) {
	wants := []int64{0, 500, 1000, 1500, 800, 1300, 1800, 2300}
	for _, tc := range []struct {
		kind, bike string
		day, hour  int64
	}{
		{"Downhill", "demo-bike-01", 4000, 2400}, {"MTB", "demo-bike-02", 2500, 1600}, {"Rennrad", "demo-bike-03", 3500, 2000},
	} {
		for _, mode := range []string{"day", "hour"} {
			for mask, want := range wants {
				s := store.New(true)
				app, err := New(s)
				if err != nil {
					t.Fatal(err)
				}
				start := domain.Today(time.Now().AddDate(0, 0, 1))
				end := start
				base := tc.day
				if mode == "hour" {
					end = start + "T11:15"
					start += "T10:00"
					base = tc.hour
				}
				choice := domain.Extras{Helmet: mask&1 != 0, Protection: mask&2 != 0, Insurance: mask&4 != 0}
				quoteBody, _ := json.Marshal(map[string]any{"bikeType": tc.kind, "mode": mode, "startsOn": start, "endsOn": end, "extras": choice})
				quote := call(t, app, "POST", "/api/ausleihen/preis", "application/json", bytes.NewReader(quoteBody), 200)
				var q domain.Quote
				_ = json.Unmarshal(quote.Body.Bytes(), &q)
				if q.BasePriceCents != base || q.ExtrasPriceCents != want || q.TotalPriceCents != base+want || q.Extras != choice || q.ExtraRates.HelmetCents != 500 {
					t.Fatalf("%s %s %d: quote=%+v", tc.kind, mode, mask, q)
				}
				input := domain.BookingInput{PersonID: "demo-person-01", BikeID: tc.bike, BikeType: tc.kind, Mode: mode, StartsOn: start, EndsOn: end, TariffVersion: 1, Extras: choice}
				body, _ := json.Marshal(input)
				created := call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 201)
				var r rentalResponse
				_ = json.Unmarshal(created.Body.Bytes(), &r)
				if r.BasePriceCents != base || r.ExtrasPriceCents != want || r.TotalPriceCents != q.TotalPriceCents || r.Extras != choice || r.ExtraPrices != q.ExtraPrices {
					t.Fatal("booking lost price breakdown")
				}
				before := s.Snapshot()
				call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 409)
				if !reflect.DeepEqual(before, s.Snapshot()) {
					t.Fatal("conflict partially stored extras")
				}
				call(t, app, "PUT", "/api/ausleihen/"+r.ID+"/stornierung", "", nil, 200)
				detail := call(t, app, "GET", "/api/ausleihen/"+r.ID, "", nil, 200)
				var stored rentalResponse
				_ = json.Unmarshal(detail.Body.Bytes(), &stored)
				if stored.Status != "cancelled" || stored.TotalPriceCents != base+want || stored.Extras != choice || stored.ExtraPrices != q.ExtraPrices {
					t.Fatal("history lost extras")
				}
			}
		}
	}
}

func TestHTTPRejectsInvalidExtrasAndFeeConfiguration(t *testing.T) {
	s := store.New(true)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	today := domain.Today(time.Now())
	input := map[string]any{"personID": "demo-person-01", "bikeID": "demo-bike-02", "bikeType": "MTB", "mode": "day", "startsOn": today, "endsOn": today, "tariffVersion": 1}
	for _, extras := range []any{nil, map[string]any{}, map[string]any{"helmet": false, "protection": false},
		map[string]any{"helmet": nil, "protection": false, "insurance": false},
		map[string]any{"helmet": "yes", "protection": false, "insurance": false},
		map[string]any{"helmet": false, "protection": false, "insurance": false, "unknown": true}, true,
	} {
		input["extras"] = extras
		body, _ := json.Marshal(input)
		before := s.Snapshot()
		call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 400)
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("invalid selection changed state")
		}
		quoteBody, _ := json.Marshal(map[string]any{"bikeType": "MTB", "mode": "day", "startsOn": today, "endsOn": today, "extras": extras})
		call(t, app, "POST", "/api/ausleihen/preis", "application/json", bytes.NewReader(quoteBody), 400)
	}
	delete(input, "extras")
	body, _ := json.Marshal(input)
	call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 400)
	input["extras"] = domain.Extras{}
	input["totalPriceCents"] = 1
	body, _ = json.Marshal(input)
	call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 400)
	for _, extras := range []any{nil, map[string]any{}, map[string]any{"helmetCents": 500, "protectionCents": 1000},
		map[string]any{"helmetCents": -1, "protectionCents": 1000, "insuranceCents": 800},
		map[string]any{"helmetCents": 1.5, "protectionCents": 1000, "insuranceCents": 800},
		map[string]any{"helmetCents": nil, "protectionCents": 1000, "insuranceCents": 800},
		map[string]any{"helmetCents": domain.MaxMoneyCents + 1, "protectionCents": 1000, "insuranceCents": 800},
		map[string]any{"helmetCents": 500, "protectionCents": 1000, "insuranceCents": 800, "unknown": 1},
	} {
		before := s.Tariff()
		body, _ := json.Marshal(map[string]any{"version": before.Version, "types": before.Types, "extras": extras})
		call(t, app, "PUT", "/api/tarife", "application/json", bytes.NewReader(body), 400)
		if !reflect.DeepEqual(before, s.Tariff()) {
			t.Fatal("invalid fees changed tariff")
		}
	}
}

func TestHTTPExtrasOverflowAndFreeSelection(t *testing.T) {
	s := store.New(true)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	tariff := s.Tariff()
	r := tariff.Types["MTB"]
	r.PricesCents = []int64{domain.MaxMoneyCents}
	tariff.Types["MTB"] = r
	if _, err := s.SaveTariff(tariff); err != nil {
		t.Fatal(err)
	}
	today := domain.Today(time.Now())
	input := domain.BookingInput{PersonID: "demo-person-01", BikeID: "demo-bike-02", BikeType: "MTB", Mode: "day", StartsOn: today, EndsOn: today, TariffVersion: 2, Extras: domain.Extras{Helmet: true}}
	body, _ := json.Marshal(input)
	call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 400)
	if len(s.Rentals()) != 0 {
		t.Fatal("overflow created a booking")
	}
	tariff = s.Tariff()
	tariff.Extras.HelmetCents = 0
	if _, err := s.SaveTariff(tariff); err != nil {
		t.Fatal(err)
	}
	input.TariffVersion = 3
	body, _ = json.Marshal(input)
	created := call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(body), 201)
	var saved rentalResponse
	_ = json.Unmarshal(created.Body.Bytes(), &saved)
	if !saved.Extras.Helmet || saved.ExtrasPriceCents != 0 || saved.TotalPriceCents != domain.MaxMoneyCents {
		t.Fatal("free selected helmet not retained")
	}
}
