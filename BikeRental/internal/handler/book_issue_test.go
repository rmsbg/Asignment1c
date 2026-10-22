package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"bikerental/internal/store"
)

func TestBookAndIssueHTTP(t *testing.T) {
	for _, flag := range []string{"", `,"issueImmediately":false`, `,"issueImmediately":true`} {
		t.Run(flag, func(t *testing.T) {
			s := store.New(true)
			app, err := New(s)
			if err != nil {
				t.Fatal(err)
			}
			body := `{"personID":"demo-person-01","bikeID":"demo-bike-02","bikeType":"MTB","mode":"hour","startMode":"now","rentalHours":1,"tariffVersion":1,"extras":{"helmet":true,"protection":false,"insurance":false}` + flag + `}`
			response := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
			var r rentalResponse
			if err := json.Unmarshal(response.Body.Bytes(), &r); err != nil {
				t.Fatal(err)
			}
			issued := strings.Contains(flag, "true")
			if (r.Status == "issued") != issued || (r.RentedAt != nil) != issued || response.Header().Get("Location") != "/api/ausleihen/"+r.ID {
				t.Fatalf("bad response: %+v", r)
			}
			call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 409)
			if len(s.Rentals()) != 1 {
				t.Fatal("duplicate booking")
			}
		})
	}
	app, _ := New(store.New(true))
	for _, flag := range []string{`"true"`, `1`} {
		call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(`{"issueImmediately":`+flag+`}`), 400)
	}
}

func TestFutureBookAndIssueHTTPIsAtomic(t *testing.T) {
	s := store.New(true)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	zone, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		t.Fatal(err)
	}
	day := time.Now().In(zone).AddDate(0, 0, 1).Format(time.DateOnly)
	body := `{"personID":"demo-person-01","bikeID":"demo-bike-02","bikeType":"MTB","mode":"day","startsOn":"` + day + `","endsOn":"` + day + `","tariffVersion":1,"issueImmediately":true,"extras":{"helmet":false,"protection":false,"insurance":false}}`
	call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 409)
	if len(s.Rentals()) != 0 {
		t.Fatal("future handover left a booking")
	}
	call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(strings.Replace(body, `"issueImmediately":true`, `"issueImmediately":false`, 1)), 201)
}
