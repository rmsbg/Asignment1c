package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

// TestMVPLifecycleMatrix checks the public workflow across type, mode and extras,
// including a tariff edit between issue and return. Lower-level tests cover all
// eight extra combinations and clock boundaries separately.
func TestMVPLifecycleMatrix(t *testing.T) {
	for _, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		for _, mode := range []string{"day", "hour"} {
			for _, withExtras := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/extras=%t", kind, mode, withExtras), func(t *testing.T) {
					s := store.New(false)
					app, err := New(s)
					if err != nil {
						t.Fatal(err)
					}
					decode := func(data []byte, target any) {
						t.Helper()
						if err := json.Unmarshal(data, target); err != nil {
							t.Fatal(err)
						}
					}
					send := func(method, path string, value any, status int) []byte {
						t.Helper()
						data, err := json.Marshal(value)
						if err != nil {
							t.Fatal(err)
						}
						return call(t, app, method, path, "application/json", bytes.NewReader(data), status).Body.Bytes()
					}
					tariff := store.New(true).Tariff()
					tariff.Version = 0
					send("PUT", "/api/tarife", tariff, 200)
					var bike domain.Bike
					decode(bikeRequest(t, app, "POST", "/api/fahrraeder", map[string]string{"name": "MVP " + kind, "type": kind, "image": "replace"}, bikePNG(t), "image/png", 201).Body.Bytes(), &bike)
					call(t, app, "GET", bike.PhotoURL, "", nil, 200)
					fields := map[string]string{"name": "Synthetische MVP-Person", "proof": "manual", "documentType": "Reisepass", "documentNumber": "TEST-MVP-MATRIX"}
					var photo []byte
					if withExtras {
						fields = map[string]string{"name": "Synthetische MVP-Person", "proof": "photo"}
						photo, err = os.ReadFile("../../testdata/ausweis-demo.png")
						if err != nil {
							t.Fatal(err)
						}
					}
					var person domain.Person
					decode(personRequest(t, app, "POST", "/api/personen", fields, photo, "image/png", 201).Body.Bytes(), &person)
					zone, err := time.LoadLocation("Europe/Vienna")
					if err != nil {
						t.Fatal(err)
					}
					now := time.Now().In(zone).Truncate(time.Minute)
					start, end := domain.Today(now), domain.Today(now)
					base := map[string]int64{"Downhill": 4000, "MTB": 2500, "Rennrad": 3500}[kind]
					if mode == "hour" {
						start = now.Format("2006-01-02T15:04")
						end = now.Add(75 * time.Minute).Format("2006-01-02T15:04")
						base = map[string]int64{"Downhill": 2400, "MTB": 1600, "Rennrad": 2000}[kind]
					}
					extras := domain.Extras{Helmet: withExtras, Protection: withExtras, Insurance: withExtras}
					var quote domain.Quote
					decode(send("POST", "/api/ausleihen/preis", map[string]any{"bikeType": kind, "mode": mode, "startsOn": start, "endsOn": end, "extras": extras}, 200), &quote)
					extraPrice := int64(0)
					if withExtras {
						extraPrice = 2300
					}
					if quote.BasePriceCents != base || quote.ExtrasPriceCents != extraPrice || quote.TotalPriceCents != base+extraPrice {
						t.Fatalf("wrong quote: %+v", quote)
					}
					query := "/api/fahrraeder/verfuegbar?" + url.Values{"bikeType": {kind}, "mode": {mode}, "startsOn": {start}, "endsOn": {end}}.Encode()
					var available []domain.Bike
					decode(call(t, app, "GET", query, "", nil, 200).Body.Bytes(), &available)
					if len(available) != 1 || available[0].ID != bike.ID || available[0].PhotoURL != bike.PhotoURL {
						t.Fatal("wrong bike selection")
					}
					input := domain.BookingInput{PersonID: person.ID, BikeID: bike.ID, BikeType: kind, Mode: mode, StartsOn: start, EndsOn: end, TariffVersion: quote.TariffVersion, Extras: extras}
					var booked rentalResponse
					decode(send("POST", "/api/ausleihen", input, 201), &booked)
					if s.Dashboard().RentedBikes != 0 || booked.Status != "booked" {
						t.Fatal("booking counted as issue")
					}
					send("POST", "/api/ausleihen", input, 409)
					issued := issueHTTP(t, app, booked.ID)
					if issued.Status != "issued" || s.Dashboard().RentedBikes != 1 {
						t.Fatal("issue missing")
					}
					tariff = s.Tariff()
					rate := tariff.Types[kind]
					rate.HourlyCents++
					rate.PricesCents[0]++
					tariff.Types[kind] = rate
					tariff.Extras.HelmetCents++
					send("PUT", "/api/tarife", tariff, 200)
					var returned rentalResponse
					decode(call(t, app, "PUT", "/api/ausleihen/"+booked.ID+"/rueckgabe", "", nil, 200).Body.Bytes(), &returned)
					if returned.Status != "returned" || s.Dashboard().AvailableBikes != 1 || s.Dashboard().RentedBikes != 0 {
						t.Fatal("return counts wrong")
					}
					var history []rentalResponse
					decode(call(t, app, "GET", "/api/ausleihen", "", nil, 200).Body.Bytes(), &history)
					if len(history) != 1 {
						t.Fatal("partial or missing booking")
					}
					for _, r := range []rentalResponse{booked, issued, returned, history[0]} {
						if r.BasePriceCents != base || r.ExtrasPriceCents != extraPrice || r.TotalPriceCents != base+extraPrice || r.Extras != extras || r.ExtraPrices != quote.ExtraPrices || r.TariffVersion != 1 || r.BikeType != kind || r.Mode != mode {
							t.Fatalf("price snapshot changed: %+v", r)
						}
					}
				})
			}
		}
	}
}
