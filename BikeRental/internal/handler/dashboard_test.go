package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"bikerental/internal/store"
)

func TestDashboardAPIAndHTML(t *testing.T) {
	for _, demo := range []bool{false, true} {
		s := store.New(demo)
		app, err := New(s)
		if err != nil {
			t.Fatal(err)
		}
		want := store.DashboardCounts{}
		if demo {
			want = store.DashboardCounts{AvailableSkis: 2, TotalSkis: 2, RegisteredPeople: 1}
		}
		w := call(t, app, "GET", "/api/dashboard", "", nil, 200)
		var got store.DashboardCounts
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatal("wrong starting counts")
		}
		var fields map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &fields)
		if len(fields) != 4 {
			t.Fatal("unexpected fields in dashboard response")
		}
		for _, key := range []string{"availableSkis", "rentedSkis", "totalSkis", "registeredPeople"} {
			if _, ok := fields[key]; !ok {
				t.Fatal("missing count")
			}
		}
		html := call(t, app, "GET", "/", "", nil, 200).Body.String()
		for _, secret := range []string{"Demo Person", "DEMO-0001", "demo-person-01", "Demoausweis"} {
			if strings.Contains(html, secret) || strings.Contains(w.Body.String(), secret) {
				t.Fatal("dashboard exposes identity data")
			}
		}
		for _, text := range []string{`id="dashboard"`, `data-metric="availableSkis"`, `data-metric="rentedSkis"`, `data-metric="totalSkis"`, `data-metric="registeredPeople"`, `/assets/dashboard.js`, `id="dashboard-status"`, `href="/ski"`, `href="/personen"`, `href="/ausleihe"`} {
			if !strings.Contains(html, text) {
				t.Fatalf("missing dashboard UI: %s", text)
			}
		}
		if strings.Contains(html, "Kennzahl folgt") || strings.Contains(html, "Noch nicht verfügbar") {
			t.Fatal("placeholder remained")
		}
		head := call(t, app, "HEAD", "/api/dashboard", "", nil, 200)
		if head.Body.Len() != 0 {
			t.Fatal("HEAD body")
		}
		call(t, app, "POST", "/api/dashboard", "", nil, 405)
		call(t, app, "GET", "/api/dashboard/missing", "", nil, 404)
	}
}
func TestDashboardTracksHTTPWorkflow(t *testing.T) {
	app, err := New(store.New(true))
	if err != nil {
		t.Fatal(err)
	}
	check := func(want store.DashboardCounts) {
		t.Helper()
		w := call(t, app, "GET", "/api/dashboard", "", nil, 200)
		var got store.DashboardCounts
		_ = json.Unmarshal(w.Body.Bytes(), &got)
		if got != want {
			t.Fatalf("got %+v want %+v", got, want)
		}
	}
	personRequest(t, app, "POST", "/api/personen", map[string]string{"name": "Dashboard Testperson", "proof": "manual", "documentType": "Reisepass", "documentNumber": "SYNTHETISCH-DASH"}, nil, "", 201)
	check(store.DashboardCounts{AvailableSkis: 2, TotalSkis: 2, RegisteredPeople: 2})
	call(t, app, "POST", "/api/ski", "application/json", strings.NewReader(`{"name":"Dashboard Testpaar"}`), 201)
	check(store.DashboardCounts{AvailableSkis: 3, TotalSkis: 3, RegisteredPeople: 2})
	body := bookingBody("demo-person-01", "demo-ski-01", 1)
	w := call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 201)
	var rental rentalResponse
	_ = json.Unmarshal(w.Body.Bytes(), &rental)
	issueHTTP(t, app, rental.ID)
	check(store.DashboardCounts{AvailableSkis: 2, RentedSkis: 1, TotalSkis: 3, RegisteredPeople: 2})
	call(t, app, "POST", "/api/ausleihen", "application/json", strings.NewReader(body), 409)
	check(store.DashboardCounts{AvailableSkis: 2, RentedSkis: 1, TotalSkis: 3, RegisteredPeople: 2})
	call(t, app, "PUT", "/api/ausleihen/"+rental.ID+"/rueckgabe", "", nil, 200)
	call(t, app, "PUT", "/api/ausleihen/"+rental.ID+"/rueckgabe", "", nil, 200)
	check(store.DashboardCounts{AvailableSkis: 3, TotalSkis: 3, RegisteredPeople: 2})
	page := httptest.NewRecorder()
	app.ServeHTTP(page, httptest.NewRequest("GET", "/", nil))
	if !strings.Contains(page.Body.String(), `data-metric="totalSkis">3</strong>`) {
		t.Fatal("SSR counts stale")
	}
}
