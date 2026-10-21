package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bikerental/internal/store"
)

func TestPagesInBothModes(t *testing.T) {
	for _, demo := range []bool{false, true} {
		app, err := New(store.New(demo))
		if err != nil {
			t.Fatal(err)
		}
		for path, title := range map[string]string{
			"/": "Dashboard", "/ski": "Ski verwalten",
			"/tarife": "Tarife verwalten", "/personen": "Personen verwalten", "/ausleihe": "Neue Ausleihe",
		} {
			t.Run(fmt.Sprintf("%s/demo=%t", path, demo), func(t *testing.T) {
				r := httptest.NewRecorder()
				app.ServeHTTP(r, httptest.NewRequest(http.MethodGet, path, nil))
				if r.Code != http.StatusOK || !strings.Contains(r.Header().Get("Content-Type"), "text/html") {
					t.Fatalf("unexpected response: %d, %v", r.Code, r.Header())
				}
				body := r.Body.String()
				for _, text := range []string{
					"<h1>" + title + "</h1>", `lang="de"`, `href="#inhalt"`,
					`href="/"`, `href="/ski"`, `href="/personen"`, `href="/ausleihe"`,
					"BikeRental · Phase 0.", "Technischer Ausgangsstand von SkiRental.",
					"Arbeitsspeicher", "beim Neustart verloren", `href="` + path + `" aria-current="page"`,
				} {
					if !strings.Contains(body, text) {
						t.Errorf("missing %q", text)
					}
				}
				if strings.Contains(body, "Demomodus · Beispieldaten") != demo {
					t.Error("incorrect demo indicator")
				}
				for _, private := range []string{"DEMO-0001", "Demo Person", "Demoausweis", "demo-person-01"} {
					if strings.Contains(body, private) {
						t.Errorf("page exposed customer data: %s", private)
					}
				}
				if path == "/" && strings.Count(body, `data-metric=`) != 4 {
					t.Error("dashboard must display four real metrics")
				}
				if path == "/ausleihe" && !strings.Contains(body, `id="ausleihe"`) {
					t.Error("rental interface missing")
				}
			})
		}
	}
}

func TestHTTPBoundariesAndAssets(t *testing.T) {
	app, err := New(store.New(false))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/unbekannt", 404},
		{"GET", "/ski/unbekannt", 404},
		{"GET", "/page.gohtml", 404},
		{"POST", "/", 405},
		{"POST", "/ausleihe", 405},
		{"DELETE", "/personen", 405},
		{"HEAD", "/", 200},
		{"GET", "/assets/styles.css", 200},
		{"GET", "/assets/missing.css", 404},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			r := httptest.NewRecorder()
			app.ServeHTTP(r, httptest.NewRequest(tc.method, tc.path, nil))
			if r.Code != tc.status {
				t.Fatalf("got %d, want %d", r.Code, tc.status)
			}
			if tc.method == "HEAD" && r.Body.Len() != 0 {
				t.Error("HEAD response must not include a body")
			}
			if tc.path == "/assets/styles.css" && (!strings.Contains(r.Header().Get("Content-Type"), "text/css") || !strings.Contains(r.Body.String(), "@media")) {
				t.Error("expected embedded responsive stylesheet")
			}
		})
	}
}
