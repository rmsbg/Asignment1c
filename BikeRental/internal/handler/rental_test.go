package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"bikerental/internal/store"
)

func TestRentalAPI(t *testing.T) {
	s := store.New(true)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	body := bookingBody("demo-person-01", "demo-ski-01", 1)
	empty := call(t, app, "GET", "/api/ausleihen", "", nil, 200)
	if strings.TrimSpace(empty.Body.String()) != "[]" {
		t.Fatal("expected empty list")
	}
	before := s.Snapshot()
	for _, tc := range []struct {
		body   string
		status int
		field  string
	}{
		{`{}`, 400, "personID"}, {`{"personID":"demo-person-01"}`, 400, "skiID"},
		{strings.Replace(body, "demo-person-01", "missing", 1), 404, "personID"},
		{strings.Replace(body, "demo-ski-01", "missing", 1), 404, "skiID"},
		{`{"personID":"demo-person-01","skiID":"demo-ski-01","id":"override"}`, 400, "form"},
		{`{"personID":"demo-person-01","skiID":"demo-ski-01","rentedAt":"override"}`, 400, "form"},
		{body + ` {}`, 400, "form"}, {`broken`, 400, "form"}, {`{"personID":"` + strings.Repeat("x", 70000) + `"}`, 413, "form"},
	} {
		w := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(tc.body), tc.status)
		if !strings.Contains(w.Body.String(), `"`+tc.field+`"`) {
			t.Fatalf("missing field error %s", tc.field)
		}
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("invalid request mutated state")
		}
	}
	call(t, app, "POST", "/api/ausleihen", "text/plain", strings.NewReader(body), 415)
	created := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
	var rental rentalResponse
	if err := json.Unmarshal(created.Body.Bytes(), &rental); err != nil {
		t.Fatal(err)
	}
	path := "/api/ausleihen/" + rental.ID
	if created.Header().Get("Location") != path || rental.ReturnedAt != nil || rental.RentedAt != nil || rental.PersonName != "Demo Person" {
		t.Fatal("invalid created response")
	}
	call(t, app, "GET", path, "", nil, 200)
	issueHTTP(t, app, rental.ID)
	call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 409)
	available := call(t, app, "GET", "/api/ski/verfuegbar", "", nil, 200)
	if strings.Contains(available.Body.String(), "demo-ski-01") {
		t.Fatal("rented ski offered")
	}
	// Multiple active rentals for the same person are valid.
	second := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(strings.Replace(body, "demo-ski-01", "demo-ski-02", 1)), 201)
	var secondRental rentalResponse
	_ = json.Unmarshal(second.Body.Bytes(), &secondRental)
	issueHTTP(t, app, secondRental.ID)
	available = call(t, app, "GET", "/api/ski/verfuegbar", "", nil, 200)
	if strings.TrimSpace(available.Body.String()) != "[]" {
		t.Fatal("expected no available skis")
	}
	before = s.Snapshot()
	call(t, app, "PUT", path+"/rueckgabe", "application/json", strings.NewReader(`{"returnedAt":"override"}`), 400)
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("invalid return mutated state")
	}
	returned := call(t, app, "PUT", path+"/rueckgabe", "", nil, 200)
	again := call(t, app, "PUT", path+"/rueckgabe", "", nil, 200)
	if returned.Body.String() != again.Body.String() {
		t.Fatal("repeated return changed response")
	}
	replacement := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
	var replacementRental rentalResponse
	_ = json.Unmarshal(replacement.Body.Bytes(), &replacementRental)
	issueHTTP(t, app, replacementRental.ID)
	call(t, app, "PUT", path+"/rueckgabe", "", nil, 200)
	available = call(t, app, "GET", "/api/ski/verfuegbar", "", nil, 200)
	if strings.TrimSpace(available.Body.String()) != "[]" {
		t.Fatal("old return released new rental")
	}
	for _, p := range []string{"/api/ausleihen", path} {
		w := call(t, app, "GET", p, "", nil, 200)
		for _, private := range []string{"DEMO-0001", "Demoausweis", "documentNumber", "photoURL"} {
			if strings.Contains(w.Body.String(), private) {
				t.Fatal("identity evidence leaked")
			}
		}
	}
	call(t, app, "GET", "/api/ausleihen/missing", "", nil, 404)
	call(t, app, "PUT", "/api/ausleihen/missing/rueckgabe", "", nil, 404)
	call(t, app, "DELETE", path, "", nil, 405)
	call(t, app, "POST", path+"/rueckgabe", "", nil, 405)
}
func TestConcurrentRentalHTTP(t *testing.T) {
	app, err := New(store.New(true))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	codes := make(chan int, 24)
	for range 24 {
		wg.Go(func() {
			r := httptest.NewRequest(http.MethodPost, "/api/ausleihen", strings.NewReader(bookingBody("demo-person-01", "demo-ski-01", 1)))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			app.ServeHTTP(w, r)
			codes <- w.Code
		})
	}
	wg.Wait()
	close(codes)
	counts := map[int]int{}
	for code := range codes {
		counts[code]++
	}
	if counts[201] != 1 || counts[409] != 23 {
		t.Fatal(fmt.Sprint(counts))
	}
}
