package handler

import (
	"bikerental/internal/domain"
	"bikerental/web"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func bookingBody(person, bike string, version uint64) string {
	today := domain.Today(time.Now())
	data, _ := json.Marshal(map[string]any{"personID": person, "bikeID": bike, "startsOn": today, "endsOn": today, "tariffVersion": version})
	return string(data)
}
func issueHTTP(t *testing.T, app http.Handler, id string) rentalResponse {
	t.Helper()
	w := call(t, app, "PUT", "/api/ausleihen/"+id+"/ausgabe", "", nil, 200)
	var r rentalResponse
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func bikePNG(t *testing.T) []byte {
	t.Helper()
	data, err := web.Files.ReadFile("assets/demo/MTB.png")
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func bikeRequest(t *testing.T, app http.Handler, method, path string, fields map[string]string, data []byte, kind string, status int) *httptest.ResponseRecorder {
	t.Helper()
	return personRequest(t, app, method, path, fields, data, kind, status)
}
func createBikeHTTP(t *testing.T, app http.Handler, name string) *httptest.ResponseRecorder {
	t.Helper()
	return bikeRequest(t, app, "POST", "/api/fahrraeder", map[string]string{"name": name, "type": "MTB", "image": "replace"}, bikePNG(t), "image/png", 201)
}
