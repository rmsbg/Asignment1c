package handler

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func TestImmediateHTTPContract(t *testing.T) {
	app, err := New(store.New(true))
	if err != nil {
		t.Fatal(err)
	}
	timing := `"bikeType":"MTB","mode":"hour","startMode":"now","rentalHours":1,"extras":{"helmet":true,"protection":false,"insurance":false}`
	quote := call(t, app, "POST", "/api/ausleihen/preis", "application/json", strings.NewReader(`{`+timing+`}`), 200)
	var q domain.Quote
	if err := json.Unmarshal(quote.Body.Bytes(), &q); err != nil {
		t.Fatal(err)
	}
	if q.StartMode != "now" || q.DurationHours != 1 || q.TotalPriceCents != 1300 {
		t.Fatalf("bad immediate quote: %+v", q)
	}
	for _, bad := range []string{`null`, `0`, `-1`, `1.5`, `2562048`, `"1"`} {
		call(t, app, "POST", "/api/ausleihen/preis", "application/json", strings.NewReader(`{`+strings.Replace(timing, `"rentalHours":1`, `"rentalHours":`+bad, 1)+`}`), 400)
	}
	call(t, app, "GET", "/api/fahrraeder/verfuegbar?mode=hour&bikeType=MTB&startMode=now&rentalHours=1", "", nil, 200)
	for _, bad := range []string{"", "0", "-1", "1.5", "999999999999999999999999"} {
		call(t, app, "GET", "/api/fahrraeder/verfuegbar?mode=hour&startMode=now&rentalHours="+bad, "", nil, 400)
	}
	before := time.Now().UTC()
	body := `{"personID":"demo-person-01","bikeID":"demo-bike-02","tariffVersion":1,` + timing + `}`
	created := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
	var r rentalResponse
	if err := json.Unmarshal(created.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	if r.StartMode != "now" || !r.CanIssue || r.StartsAt.Before(before) || r.StartsAt.After(time.Now()) || r.StartsAt != r.CreatedAt || r.EndsAt.Sub(r.StartsAt) != time.Hour || r.TotalPriceCents != q.TotalPriceCents {
		t.Fatalf("bad immediate booking: %+v", r)
	}
	issued := issueHTTP(t, app, r.ID)
	if issued.Status != "issued" || issued.StartsAt != r.StartsAt || issued.EndsAt != r.EndsAt {
		t.Fatal("immediate issue shifted period")
	}
	call(t, app, "POST", "/api/ausleihen", "application/json", bytes.NewReader([]byte(body)), 409)
}
