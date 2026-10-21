package handler

import (
	"bikerental/internal/domain"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func bookingBody(person, ski string, version uint64) string {
	today := domain.Today(time.Now())
	data, _ := json.Marshal(map[string]any{"personID": person, "skiID": ski, "startsOn": today, "endsOn": today, "tariffVersion": version})
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
