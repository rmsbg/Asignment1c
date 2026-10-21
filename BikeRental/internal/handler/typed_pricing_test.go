package handler

import (
	"bytes"
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func TestTypedHourlyAndDailyHTTPContracts(t *testing.T) {
	for _, tc := range []struct {
		kind, bike string
		hour, day  int64
	}{
		{"Downhill", "demo-bike-01", 1200, 4000}, {"MTB", "demo-bike-02", 800, 2500}, {"Rennrad", "demo-bike-03", 1000, 3500},
	} {
		for _, mode := range []string{"day", "hour"} {
			t.Run(tc.kind+mode, func(t *testing.T) {
				s := store.New(true)
				app, err := New(s)
				if err != nil {
					t.Fatal(err)
				}
				start := domain.Today(time.Now().AddDate(0, 0, 1))
				end := start
				price := tc.day
				if mode == "hour" {
					end = start + "T11:15"
					start += "T10:00"
					price = 2 * tc.hour
				}
				quoteBody, _ := json.Marshal(map[string]any{"bikeType": tc.kind, "mode": mode, "startsOn": start, "endsOn": end})
				w := call(t, app, "POST", "/api/ausleihen/preis", "application/json", bytes.NewReader(quoteBody), 200)
				var q domain.Quote
				_ = json.Unmarshal(w.Body.Bytes(), &q)
				if q.TotalPriceCents != price || q.Mode != mode || q.BikeType != tc.kind || q.TariffVersion != 1 || !q.EndsAt.After(q.StartsAt) {
					t.Fatalf("wrong quote: %+v", q)
				}
				query := "/api/fahrraeder/verfuegbar?" + url.Values{"mode": {mode}, "bikeType": {tc.kind}, "startsOn": {start}, "endsOn": {end}}.Encode()
				available := call(t, app, "GET", query, "", nil, 200)
				var bikes []domain.Bike
				_ = json.Unmarshal(available.Body.Bytes(), &bikes)
				if len(bikes) != 1 || bikes[0].ID != tc.bike || bikes[0].PhotoURL == "" {
					t.Fatal("type filter failed")
				}
				input := domain.BookingInput{PersonID: "demo-person-01", BikeID: tc.bike, BikeType: tc.kind, Mode: mode, StartsOn: start, EndsOn: end, TariffVersion: 1}
				data, _ := json.Marshal(input)
				created := call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(data), 201)
				var r rentalResponse
				_ = json.Unmarshal(created.Body.Bytes(), &r)
				if r.TotalPriceCents != price || r.Mode != mode || r.BikeType != tc.kind || r.CanIssue || r.Status != "booked" {
					t.Fatalf("wrong rental %+v", r)
				}
				if mode == "hour" && (r.DurationHours != 2 || r.DurationDays != 0) {
					t.Fatal("hour duration wrong")
				}
				if mode == "day" && (r.DurationDays != 1 || r.DurationHours != 0) {
					t.Fatal("day duration wrong")
				}
				before := s.Snapshot()
				call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(data), 409)
				if !reflect.DeepEqual(before, s.Snapshot()) {
					t.Fatal("conflict mutated state")
				}
				call(t, app, "PUT", "/api/ausleihen/"+r.ID+"/ausgabe", "", nil, 409)
				call(t, app, "PUT", "/api/ausleihen/"+r.ID+"/stornierung", "", nil, 200)
				available = call(t, app, "GET", query, "", nil, 200)
				if !strings.Contains(available.Body.String(), tc.bike) {
					t.Fatal("cancellation did not release interval")
				}
			})
		}
	}
}

func TestTypedTariffRejectsIncompleteAndNullPrices(t *testing.T) {
	s := store.New(false)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	valid := tariffBody(0, []int64{2500}, 1500)
	variants := []string{
		strings.Replace(valid, `"hourlyCents":800`, `"hourlyCents":null`, 1),
		strings.Replace(valid, `"hourlyCents":800`, `"hourlyCents":1.5`, 1),
		strings.Replace(valid, `"hourlyCents":800`, `"hourlyCents":-1`, 1),
		strings.Replace(valid, `"pricesCents":[2500]`, `"pricesCents":[null]`, 1),
		strings.Replace(valid, `"pricesCents":[2500]`, `"pricesCents":[]`, 1),
		strings.Replace(valid, `"extraDayCents":1500`, `"extraDayCents":null`, 1),
		strings.Replace(valid, `"MTB"`, `"Other"`, 1),
		`{"version":0,"types":{"MTB":{"hourlyCents":800,"pricesCents":[2500],"extraDayCents":1500}}}`,
		`{"version":0,"types":{"Downhill":null,"MTB":null,"Rennrad":null}}`,
	}
	before := s.Tariff()
	for _, body := range variants {
		call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(body), 400)
		if !reflect.DeepEqual(before, s.Tariff()) {
			t.Fatal("invalid tariff saved")
		}
	}
	call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(valid), 200)
	call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(valid), 409)
}
