package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"testing"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

// TestMVPAcceptanceWorkflow follows the public HTTP interface from fresh inventory
// through registration, rental, return and reuse for every proof/start variant.
func TestMVPAcceptanceWorkflow(t *testing.T) {
	for _, demo := range []bool{false, true} {
		for _, proof := range []string{"manual", "png", "jpeg"} {
			t.Run(fmt.Sprintf("demo=%t/%s", demo, proof), func(t *testing.T) {
				t.Parallel()
				s := store.New(demo)
				app, err := New(s)
				if err != nil {
					t.Fatal(err)
				}
				baseline := s.Dashboard()
				expect := func(addedSkis, addedPeople, rented int) {
					t.Helper()
					w := call(t, app, "GET", "/api/dashboard", "", nil, 200)
					var got store.DashboardCounts
					if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
						t.Fatal(err)
					}
					want := store.DashboardCounts{AvailableSkis: baseline.TotalSkis + addedSkis - rented, RentedSkis: rented, TotalSkis: baseline.TotalSkis + addedSkis, RegisteredPeople: baseline.RegisteredPeople + addedPeople}
					if got != want {
						t.Fatalf("dashboard=%+v want=%+v", got, want)
					}
				}
				expect(0, 0, 0)
				skiResponse := call(t, app, "POST", "/api/ski", "application/json", strings.NewReader(`{"name":"MVP Testpaar"}`), 201)
				var ski domain.Ski
				if err := json.Unmarshal(skiResponse.Body.Bytes(), &ski); err != nil {
					t.Fatal(err)
				}
				if ski.ID == "" || ski.Name != "MVP Testpaar" {
					t.Fatal("missing generated ski data")
				}
				expect(1, 0, 0)
				fields := map[string]string{"name": "MVP Testperson", "proof": "manual", "documentType": "Reisepass", "documentNumber": "SYNTHETISCH-MVP"}
				var photo []byte
				mime := ""
				if proof != "manual" {
					fields = map[string]string{"name": "MVP Testperson", "proof": "photo"}
					photo, err = os.ReadFile("../../testdata/ausweis-demo.png")
					if err != nil {
						t.Fatal(err)
					}
					mime = "image/png"
					if proof == "jpeg" {
						img, err := png.Decode(bytes.NewReader(photo))
						if err != nil {
							t.Fatal(err)
						}
						var encoded bytes.Buffer
						if err := jpeg.Encode(&encoded, img, nil); err != nil {
							t.Fatal(err)
						}
						photo = encoded.Bytes()
						mime = "image/jpeg"
					}
				}
				response := personRequest(t, app, "POST", "/api/personen", fields, photo, mime, 201)
				var person domain.Person
				if err := json.Unmarshal(response.Body.Bytes(), &person); err != nil {
					t.Fatal(err)
				}
				if person.ID == "" || person.Name != "MVP Testperson" {
					t.Fatal("missing generated person data")
				}
				expect(1, 1, 0)
				detail := call(t, app, "GET", "/api/personen/"+person.ID, "", nil, 200)
				if proof == "manual" {
					if !strings.Contains(detail.Body.String(), "SYNTHETISCH-MVP") {
						t.Fatal("manual proof missing")
					}
				} else {
					if person.PhotoURL == "" {
						t.Fatal("photo URL missing")
					}
					gotPhoto := call(t, app, "GET", person.PhotoURL, "", nil, 200)
					if gotPhoto.Header().Get("Content-Type") != mime || !bytes.Equal(photo, gotPhoto.Body.Bytes()) {
						t.Fatal("photo not preserved")
					}
				}
				people := call(t, app, "GET", "/api/personen", "", nil, 200)
				for _, private := range []string{"documentNumber", "documentType", "photoURL", "SYNTHETISCH-MVP"} {
					if strings.Contains(people.Body.String(), private) {
						t.Fatal("overview exposed proof")
					}
				}
				if !demo {
					call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(`{"version":0,"pricesCents":[2500,4500,6000],"extraDayCents":1500}`), 200)
				}
				input := []byte(bookingBody(person.ID, ski.ID, 1))
				create := func() rentalResponse {
					t.Helper()
					w := call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(input), 201)
					var rental rentalResponse
					if err := json.Unmarshal(w.Body.Bytes(), &rental); err != nil {
						t.Fatal(err)
					}
					if rental.PersonID != person.ID || rental.SkiID != ski.ID || rental.PersonName != person.Name || rental.SkiName != ski.Name || rental.ReturnedAt != nil {
						t.Fatal("rental references wrong records")
					}
					return issueHTTP(t, app, rental.ID)
				}
				first := create()
				expect(1, 1, 1)
				call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader(input), 409)
				expect(1, 1, 1)
				available := call(t, app, "GET", "/api/ski/verfuegbar", "", nil, 200)
				if strings.Contains(available.Body.String(), ski.ID) {
					t.Fatal("rented pair still selectable")
				}
				path := "/api/ausleihen/" + first.ID + "/rueckgabe"
				returned := call(t, app, "PUT", path, "", nil, 200)
				expect(1, 1, 0)
				repeat := call(t, app, "PUT", path, "", nil, 200)
				if returned.Body.String() != repeat.Body.String() {
					t.Fatal("repeated return changed timestamp")
				}
				expect(1, 1, 0)
				available = call(t, app, "GET", "/api/ski/verfuegbar", "", nil, 200)
				if !strings.Contains(available.Body.String(), ski.ID) {
					t.Fatal("returned pair not available")
				}
				next := create()
				if next.ID == first.ID {
					t.Fatal("rental ID reused")
				}
				expect(1, 1, 1)
				call(t, app, "PUT", path, "", nil, 200)
				expect(1, 1, 1)
				call(t, app, "PUT", "/api/ausleihen/"+next.ID+"/rueckgabe", "", nil, 200)
				expect(1, 1, 0)
				history := call(t, app, "GET", "/api/ausleihen", "", nil, 200)
				var records []rentalResponse
				if err := json.Unmarshal(history.Body.Bytes(), &records); err != nil {
					t.Fatal(err)
				}
				if len(records) != 2 || records[0].ReturnedAt == nil || records[1].ReturnedAt == nil {
					t.Fatal("returned history incomplete")
				}
				page := call(t, app, "GET", "/", "", nil, 200)
				if !strings.Contains(page.Body.String(), fmt.Sprintf(`data-metric="totalSkis">%d</strong>`, baseline.TotalSkis+1)) {
					t.Fatal("HTML dashboard out of sync")
				}
				if store.New(demo).Dashboard() != baseline {
					t.Fatal("fresh start inherited scenario data")
				}
			})
		}
	}
}
