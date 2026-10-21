package handler

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func call(t *testing.T, app http.Handler, method, path, kind string, body io.Reader, status int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, body)
	if kind != "" {
		r.Header.Set("Content-Type", kind)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, status, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("missing cache protection")
	}
	return w
}
func personRequest(t *testing.T, app http.Handler, method, path string, fields map[string]string, photo []byte, mime string, status int) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	for k, v := range fields {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if photo != nil {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", `form-data; name="photo"; filename="synthetic.png"`)
		h.Set("Content-Type", mime)
		part, err := writer.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = part.Write(photo)
	}
	_ = writer.Close()
	return call(t, app, method, path, writer.FormDataContentType(), &b, status)
}
func TestManagementAPI(t *testing.T) {
	s := store.New(false)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	bike := bikeRequest(t, app, "POST", "/api/fahrraeder", map[string]string{"name": " Test Bike ", "type": "MTB", "image": "replace"}, bikePNG(t), "image/png", 201)
	var sk domain.Bike
	_ = json.Unmarshal(bike.Body.Bytes(), &sk)
	if sk.Name != "Test Bike" || bike.Header().Get("Location") != "/api/fahrraeder/"+sk.ID {
		t.Fatal("bike creation")
	}
	call(t, app, "GET", "/api/fahrraeder/"+sk.ID, "", nil, 200)
	bikeRequest(t, app, "PUT", "/api/fahrraeder/"+sk.ID, map[string]string{"name": "Edited", "type": "Rennrad", "image": "keep"}, nil, "", 200)
	bikeRequest(t, app, "PUT", "/api/fahrraeder/"+sk.ID, map[string]string{"name": " ", "type": "Rennrad", "image": "keep"}, nil, "", 400)
	call(t, app, "POST", "/api/fahrraeder", "application/json", strings.NewReader(`{"name":"X"}`), 415)
	call(t, app, "GET", "/api/fahrraeder/missing", "", nil, 404)
	bikeRequest(t, app, "PUT", "/api/fahrraeder/missing", map[string]string{"name": "X", "type": "MTB", "image": "keep"}, nil, "", 404)
	fields := map[string]string{"name": " Synthetic Person ", "proof": "manual", "documentType": "Sonstiger Ausweis", "documentDescription": "Testdokument", "documentNumber": "TEST-42"}
	created := personRequest(t, app, "POST", "/api/personen", fields, nil, "", 201)
	var p domain.Person
	_ = json.Unmarshal(created.Body.Bytes(), &p)
	path := "/api/personen/" + p.ID
	if p.Name != "Synthetic Person" || created.Header().Get("Location") != path {
		t.Fatal("person creation")
	}
	detail := call(t, app, "GET", path, "", nil, 200)
	if !strings.Contains(detail.Body.String(), "TEST-42") {
		t.Fatal("detail missing proof")
	}
	list := call(t, app, "GET", "/api/personen", "", nil, 200)
	var summaries []map[string]any
	_ = json.Unmarshal(list.Body.Bytes(), &summaries)
	if len(summaries) != 1 || len(summaries[0]) != 2 {
		t.Fatal("list exposes extra fields")
	}
	call(t, app, "GET", path+"/foto", "", nil, 404)
	personRequest(t, app, "PUT", path, map[string]string{"name": "Renamed", "proof": "keep"}, nil, "", 200)
	before := s.Snapshot()
	for _, bad := range []map[string]string{
		{"name": " ", "proof": "keep"}, {"name": "X", "proof": "manual", "documentType": "Unknown", "documentNumber": "X"},
		{"name": "X", "proof": "manual", "documentType": "Reisepass"}, {"name": "X", "proof": "manual", "documentType": "Sonstiger Ausweis", "documentNumber": "X"},
		{"name": "X", "proof": "keep", "documentNumber": "X"}, {"name": "X", "proof": "photo"},
	} {
		personRequest(t, app, "PUT", path, bad, nil, "", 400)
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("invalid edit mutated state")
		}
	}
	var pic bytes.Buffer
	_ = png.Encode(&pic, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	photoFields := map[string]string{"name": "Photo Person", "proof": "photo"}
	for _, tc := range []struct {
		data   []byte
		mime   string
		status int
	}{
		{[]byte("broken"), "image/png", 400}, {pic.Bytes(), "image/jpeg", 400}, {pic.Bytes()[:len(pic.Bytes())-15], "image/png", 400}, {make([]byte, domain.MaxPhotoBytes+1), "image/png", 413},
	} {
		personRequest(t, app, "PUT", path, photoFields, tc.data, tc.mime, tc.status)
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("invalid photo mutated state")
		}
	}
	personRequest(t, app, "PUT", path, photoFields, pic.Bytes(), "image/png", 200)
	imageResponse := call(t, app, "GET", path+"/foto", "", nil, 200)
	if !bytes.Equal(imageResponse.Body.Bytes(), pic.Bytes()) || imageResponse.Header().Get("Content-Type") != "image/png" {
		t.Fatal("wrong photo response")
	}
	detail = call(t, app, "GET", path, "", nil, 200)
	if strings.Contains(detail.Body.String(), "TEST-42") || !strings.Contains(detail.Body.String(), "photoURL") {
		t.Fatal("photo switch retained manual proof")
	}
	personRequest(t, app, "PUT", path, map[string]string{"name": "Photo renamed", "proof": "keep"}, nil, "", 200)
	imageResponse = call(t, app, "GET", path+"/foto", "", nil, 200)
	if !bytes.Equal(imageResponse.Body.Bytes(), pic.Bytes()) {
		t.Fatal("keep changed photo")
	}
	personRequest(t, app, "PUT", path, fields, nil, "", 200)
	call(t, app, "GET", path+"/foto", "", nil, 404)
	personRequest(t, app, "PUT", "/api/personen/missing", fields, nil, "", 404)
	personRequest(t, app, "POST", "/api/personen", map[string]string{"name": "X", "proof": "keep"}, nil, "", 400)
	call(t, app, "GET", "/api/personen/missing", "", nil, 404)
	call(t, app, "GET", "/api/personen/missing/foto", "", nil, 404)
	call(t, app, "DELETE", path, "", nil, 405)
	call(t, app, "POST", "/api/personen", "application/json", strings.NewReader(`{}`), 415)
	// A fresh process-local state never inherits edits or photos.
	if len(store.New(false).People()) != 0 {
		t.Fatal("fresh store inherited data")
	}
}

func TestMultipartBoundariesAndPhotoCreation(t *testing.T) {
	s := store.New(false)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	var pic bytes.Buffer
	_ = png.Encode(&pic, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	fields := map[string]string{"name": "Photo creation", "proof": "photo"}
	created := personRequest(t, app, "POST", "/api/personen", fields, pic.Bytes(), "image/png", 201)
	var p domain.Person
	_ = json.Unmarshal(created.Body.Bytes(), &p)
	if p.PhotoURL == "" {
		t.Fatal("photo creation missing URL")
	}
	before := s.Snapshot()
	path := "/api/personen/" + p.ID
	personRequest(t, app, "PUT", path, map[string]string{"name": "X", "proof": "keep"}, []byte{}, "image/png", 400)
	personRequest(t, app, "PUT", path, map[string]string{"name": "X", "proof": "photo", "documentType": "Reisepass", "documentNumber": "X"}, pic.Bytes(), "image/png", 400)
	personRequest(t, app, "PUT", path, map[string]string{"name": "X", "proof": "manual", "id": "override"}, nil, "", 400)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("name", "A")
	_ = writer.WriteField("name", "B")
	_ = writer.Close()
	call(t, app, "PUT", path, writer.FormDataContentType(), &body, 400)
	call(t, app, "PUT", path, "multipart/form-data; boundary=broken", strings.NewReader("truncated"), 400)
	personRequest(t, app, "PUT", path, map[string]string{"name": strings.Repeat("x", 9000), "proof": "keep"}, nil, "", 413)
	bikeRequest(t, app, "POST", "/api/fahrraeder", map[string]string{"name": strings.Repeat("x", 70000), "type": "MTB", "image": "replace"}, nil, "", 413)
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("malformed request mutated state")
	}
	call(t, app, "GET", "/api/fahrraeder", "", nil, 200)
	call(t, app, "GET", path+"/foto", "", nil, 200)
}
