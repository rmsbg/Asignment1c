package handler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func TestPhase5HTTPContracts(t *testing.T) {
	s := store.New(false)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	empty := call(t, app, "GET", "/api/tarife", "", nil, 200)
	if !strings.Contains(empty.Body.String(), `"version":0`) || !strings.Contains(empty.Body.String(), `"pricesCents":[]`) {
		t.Fatal("invalid empty tariff")
	}
	today := domain.Today(time.Now())
	dates := fmt.Sprintf(`{"startsOn":%q,"endsOn":%q}`, today, today)
	call(t, app, "POST", "/api/ausleihen/preis", "application/json", strings.NewReader(dates), 400)
	for _, body := range []string{`null`, `{}`, `{"version":0,"pricesCents":[],"extraDayCents":1}`, `{"version":0,"pricesCents":[-1],"extraDayCents":1}`, `{"version":0,"pricesCents":[1.5],"extraDayCents":1}`, `{"version":0,"pricesCents":[null],"extraDayCents":1}`, `{"version":0,"pricesCents":[1],"extraDayCents":null}`, `{"version":0,"pricesCents":[1],"extraDayCents":1,"surprise":2}`} {
		call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(body), 400)
		if s.Tariff().Version != 0 {
			t.Fatal("invalid tariff changed store")
		}
	}
	tariff := `{"version":0,"pricesCents":[2500,4500,6000],"extraDayCents":1500}`
	call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(tariff), 200)
	call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(tariff), 409)
	quote := call(t, app, "POST", "/api/ausleihen/preis", "application/json", strings.NewReader(dates), 200)
	var price domain.Quote
	_ = json.Unmarshal(quote.Body.Bytes(), &price)
	if price.TotalPriceCents != 2500 || price.DurationDays != 1 || price.TariffVersion != 1 {
		t.Fatal("wrong quote")
	}
	call(t, app, "GET", "/api/fahrraeder/verfuegbar?startsOn="+today, "", nil, 400)
	call(t, app, "GET", "/api/fahrraeder/verfuegbar?startsOn=&endsOn=", "", nil, 400)
	call(t, app, "POST", "/api/ausleihen/preis", "application/json", strings.NewReader(`{"startsOn":"2020-01-01","endsOn":"2020-01-02"}`), 400)
	call(t, app, "POST", "/api/ausleihen/preis", "text/plain", strings.NewReader(dates), 415)
	call(t, app, "POST", "/api/ausleihen/preis", "application/json", strings.NewReader(`{"startsOn":"`+strings.Repeat("x", 70000)+`"}`), 413)
	call(t, app, "POST", "/api/tarife", "", nil, 405)
}
func TestPhase5BookingHTTPFlow(t *testing.T) {
	s := store.New(true)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	body := bookingBody("demo-person-01", "demo-bike-01", 1)
	created := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
	var booked rentalResponse
	_ = json.Unmarshal(created.Body.Bytes(), &booked)
	if booked.Status != "booked" || booked.RentedAt != nil || booked.TotalPriceCents != 2500 || booked.DurationDays != 1 {
		t.Fatal("wrong booked response")
	}
	if s.Dashboard().RentedBikes != 0 {
		t.Fatal("booking counted as issue")
	}
	today := domain.Today(time.Now())
	query := "/api/fahrraeder/verfuegbar?startsOn=" + today + "&endsOn=" + today
	available := call(t, app, "GET", query, "", nil, 200)
	if strings.Contains(available.Body.String(), "demo-bike-01") {
		t.Fatal("reserved pair offered for same date")
	}
	unissued := call(t, app, "GET", "/api/fahrraeder/verfuegbar", "", nil, 200)
	if !strings.Contains(unissued.Body.String(), "demo-bike-01") {
		t.Fatal("physical availability includes reservation")
	}
	path := "/api/ausleihen/" + booked.ID
	before := s.Snapshot()
	call(t, app, "PUT", path+"/rueckgabe", "", nil, 409)
	call(t, app, "PUT", path+"/ausgabe", "application/json", strings.NewReader(`{}`), 400)
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("bad actions changed state")
	}
	cancelled := call(t, app, "PUT", path+"/stornierung", "", nil, 200)
	again := call(t, app, "PUT", path+"/stornierung", "", nil, 200)
	if cancelled.Body.String() != again.Body.String() {
		t.Fatal("cancel not idempotent")
	}
	call(t, app, "PUT", path+"/ausgabe", "", nil, 409)
	created = call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
	_ = json.Unmarshal(created.Body.Bytes(), &booked)
	path = "/api/ausleihen/" + booked.ID
	issued := issueHTTP(t, app, booked.ID)
	if issued.Status != "issued" || issued.RentedAt == nil || s.Dashboard().RentedBikes != 1 {
		t.Fatal("wrong issue state")
	}
	repeat := issueHTTP(t, app, booked.ID)
	if !reflect.DeepEqual(issued, repeat) {
		t.Fatal("issue changed timestamp")
	}
	call(t, app, "PUT", path+"/stornierung", "", nil, 409)
	call(t, app, "PUT", "/api/tarife", "application/json", strings.NewReader(`{"version":1,"pricesCents":[3000],"extraDayCents":1000}`), 200)
	staleBody := bookingBody("demo-person-01", "demo-bike-02", 1)
	before = s.Snapshot()
	call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(staleBody), 409)
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("stale tariff created rental")
	}
	returned := call(t, app, "PUT", path+"/rueckgabe", "", nil, 200)
	var done rentalResponse
	_ = json.Unmarshal(returned.Body.Bytes(), &done)
	if done.Status != "returned" || done.TotalPriceCents != 2500 || done.TariffVersion != 1 || s.Dashboard().RentedBikes != 0 {
		t.Fatal("return changed booked price")
	}
	for _, action := range []string{"ausgabe", "rueckgabe", "stornierung"} {
		call(t, app, "PUT", "/api/ausleihen/missing/"+action, "", nil, 404)
	}
	// A future booking can be cancelled, but never issued before its first date.
	tomorrow := time.Now().AddDate(0, 0, 1)
	future := domain.Today(tomorrow)
	futureBody := fmt.Sprintf(`{"personID":"demo-person-01","bikeID":"demo-bike-01","startsOn":%q,"endsOn":%q,"tariffVersion":2}`, future, future)
	response := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(futureBody), 201)
	var futureRental rentalResponse
	_ = json.Unmarshal(response.Body.Bytes(), &futureRental)
	call(t, app, "PUT", "/api/ausleihen/"+futureRental.ID+"/ausgabe", "", nil, 409)
}
