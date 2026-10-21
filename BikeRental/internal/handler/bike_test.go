package handler

import (
	"bytes"
	"encoding/json"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"reflect"
	"strings"
	"testing"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func TestBikeAPIImagesAndInvalidReplacements(t *testing.T) {
	s := store.New(false)
	app, err := New(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		w := bikeRequest(t, app, "POST", "/api/fahrraeder", map[string]string{"name": " Test " + kind + " ", "type": kind, "image": "replace"}, bikePNG(t), "image/png", 201)
		var b domain.Bike
		if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
			t.Fatal(err)
		}
		path := "/api/fahrraeder/" + b.ID
		if b.Name != "Test "+kind || b.Type != kind || b.PhotoURL != path+"/foto" || bytes.Contains(w.Body.Bytes(), []byte(`"Photo"`)) {
			t.Fatal("invalid bike response")
		}
		got := call(t, app, "GET", b.PhotoURL, "", nil, 200)
		if !bytes.Equal(got.Body.Bytes(), bikePNG(t)) || got.Header().Get("Content-Type") != "image/png" {
			t.Fatal("incorrect image")
		}
		head := call(t, app, "HEAD", b.PhotoURL, "", nil, 200)
		if head.Body.Len() != 0 || head.Header().Get("Content-Type") != "image/png" || head.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatal("incorrect HEAD response")
		}
		before := s.Snapshot()
		for _, fields := range []map[string]string{
			{"name": " ", "type": kind, "image": "keep"},
			{"name": "X", "type": "Unknown", "image": "keep"},
			{"name": "X", "type": kind, "image": "remove"},
			{"name": "X", "type": kind, "image": "keep", "id": "override"},
		} {
			bikeRequest(t, app, "PUT", path, fields, nil, "", 400)
			if !reflect.DeepEqual(before, s.Snapshot()) {
				t.Fatal("invalid fields changed inventory")
			}
		}
		for _, tc := range []struct {
			data   []byte
			media  string
			status int
		}{
			{nil, "", 400}, {[]byte("broken"), "image/png", 400},
			{bikePNG(t), "image/jpeg", 400}, {make([]byte, domain.MaxPhotoBytes+1), "image/png", 413},
			{bikePNG(t)[:len(bikePNG(t))-15], "image/png", 400},
		} {
			bikeRequest(t, app, "PUT", path, map[string]string{"name": "Changed", "type": kind, "image": "replace"}, tc.data, tc.media, tc.status)
			if !reflect.DeepEqual(before, s.Snapshot()) {
				t.Fatal("invalid image changed inventory")
			}
		}
		bikeRequest(t, app, "PUT", path, map[string]string{"name": "Kept", "type": kind, "image": "keep"}, nil, "", 200)
		if !bytes.Equal(call(t, app, "GET", b.PhotoURL, "", nil, 200).Body.Bytes(), bikePNG(t)) {
			t.Fatal("keep lost image")
		}
		bikeRequest(t, app, "PUT", path, map[string]string{"name": "Mixed", "type": kind, "image": "keep"}, []byte{}, "image/png", 400)
		img, err := png.Decode(bytes.NewReader(bikePNG(t)))
		if err != nil {
			t.Fatal(err)
		}
		var jpegData bytes.Buffer
		if err := jpeg.Encode(&jpegData, img, nil); err != nil {
			t.Fatal(err)
		}
		bikeRequest(t, app, "PUT", path, map[string]string{"name": "Replaced", "type": kind, "image": "replace"}, jpegData.Bytes(), "image/jpeg", 200)
		got = call(t, app, "GET", b.PhotoURL, "", nil, 200)
		if got.Header().Get("Content-Type") != "image/jpeg" || !bytes.Equal(got.Body.Bytes(), jpegData.Bytes()) {
			t.Fatal("replacement not served")
		}
	}
	list := call(t, app, "GET", "/api/fahrraeder", "", nil, 200)
	var bikes []domain.Bike
	if err := json.Unmarshal(list.Body.Bytes(), &bikes); err != nil || len(bikes) != 3 {
		t.Fatal("missing inventory")
	}
	available := call(t, app, "GET", "/api/fahrraeder/verfuegbar", "", nil, 200)
	if !strings.Contains(available.Body.String(), `"photoURL"`) || strings.Contains(available.Body.String(), "documentNumber") {
		t.Fatal("wrong selection metadata")
	}
	call(t, app, "GET", "/api/ski", "", nil, 404)
	call(t, app, "GET", "/api/fahrraeder/missing/foto", "", nil, 404)
	call(t, app, "DELETE", "/api/fahrraeder/"+bikes[0].ID, "", nil, 405)
	before := s.Snapshot()
	for _, fields := range []map[string]string{
		{"name": "X", "type": "MTB", "image": "keep"},
		{"name": "X", "type": "MTB", "image": "replace"},
	} {
		bikeRequest(t, app, "POST", "/api/fahrraeder", fields, nil, "", 400)
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("name", "A")
	_ = w.WriteField("name", "B")
	_ = w.Close()
	call(t, app, "POST", "/api/fahrraeder", w.FormDataContentType(), &body, 400)
	call(t, app, "POST", "/api/fahrraeder", "multipart/form-data; boundary=broken", strings.NewReader("truncated"), 400)
	if !reflect.DeepEqual(before, s.Snapshot()) {
		t.Fatal("malformed creation changed state")
	}
}
