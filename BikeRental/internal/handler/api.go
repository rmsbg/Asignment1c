package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func apiError(w http.ResponseWriter, err error) {
	status := 400
	fields := domain.ValidationErrors{"form": "Die Anfrage ist ungültig."}
	var validation domain.ValidationErrors
	var size *http.MaxBytesError
	var missing store.MissingReference
	var conflict store.Conflict
	if errors.As(err, &validation) {
		fields = validation
	} else if errors.As(err, &missing) {
		status = 404
		fields = domain.ValidationErrors{missing.Field: missing.Error()}
	} else if errors.As(err, &conflict) {
		status = 409
		fields = domain.ValidationErrors{conflict.Field: conflict.Message}
	} else if errors.Is(err, store.ErrUnavailable) {
		status = 409
		fields = domain.ValidationErrors{"bikeID": store.ErrUnavailable.Error()}
	} else if errors.Is(err, store.ErrNotFound) {
		status = 404
		fields["form"] = "Der Datensatz wurde nicht gefunden."
	} else if errors.As(err, &size) {
		status = 413
		fields = domain.ValidationErrors{"form": "Die Anfrage ist zu groß. Bitte die Eingaben verkleinern."}
		if size.Limit == domain.MaxPhotoBytes {
			fields = domain.ValidationErrors{"photo": "Das Foto darf höchstens 5 MiB groß sein."}
		}
	}
	writeJSON(w, status, map[string]any{"errors": fields})
}
func exposePerson(p domain.Person) domain.Person {
	if len(p.Photo) > 0 {
		p.PhotoURL = "/api/personen/" + p.ID + "/foto"
	}
	return p
}
func registerAPI(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("GET /api/fahrraeder", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.Bikes()) })
	mux.HandleFunc("GET /api/fahrraeder/{id}", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.Bike(r.PathValue("id"))
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, p)
	})
	mux.HandleFunc("GET /api/fahrraeder/{id}/foto", func(w http.ResponseWriter, r *http.Request) {
		b, err := s.Bike(r.PathValue("id"))
		if err != nil {
			apiError(w, err)
			return
		}
		w.Header().Set("Content-Type", b.PhotoType)
		if r.Method != http.MethodHead {
			_, _ = w.Write(b.Photo)
		}
	})
	bikeWrite := func(w http.ResponseWriter, r *http.Request) {
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if media != "multipart/form-data" {
			writeJSON(w, 415, map[string]any{"errors": domain.ValidationErrors{"form": "Multipart-Formular erwartet."}})
			return
		}
		fields, photo, photoType, seenPhoto, err := readMultipart(w, r, []string{"name", "type", "image"})
		if err != nil {
			apiError(w, err)
			return
		}
		keep := fields["image"] == "keep"
		if (keep && seenPhoto) || (!keep && fields["image"] != "replace") {
			apiError(w, domain.ValidationErrors{"image": "Bitte Bild beibehalten oder durch eine neue Datei ersetzen."})
			return
		}
		b, err := s.SaveBike(r.PathValue("id"), domain.Bike{Name: fields["name"], Type: fields["type"], Photo: photo, PhotoType: photoType}, keep)
		if err != nil {
			apiError(w, err)
			return
		}
		status := 200
		if r.Method == "POST" {
			status = 201
			w.Header().Set("Location", "/api/fahrraeder/"+b.ID)
		}
		writeJSON(w, status, b)
	}
	mux.HandleFunc("POST /api/fahrraeder", bikeWrite)
	mux.HandleFunc("PUT /api/fahrraeder/{id}", bikeWrite)
	mux.HandleFunc("GET /api/personen", func(w http.ResponseWriter, r *http.Request) {
		// A dedicated DTO guarantees that even empty identity fields are absent.
		type summary struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		list := []summary{}
		for _, p := range s.People() {
			list = append(list, summary{p.ID, p.Name})
		}
		writeJSON(w, 200, list)
	})
	mux.HandleFunc("GET /api/personen/{id}", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.Person(r.PathValue("id"))
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, exposePerson(p))
	})
	mux.HandleFunc("GET /api/personen/{id}/foto", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.Person(r.PathValue("id"))
		if err == nil && len(p.Photo) == 0 {
			err = store.ErrNotFound
		}
		if err != nil {
			apiError(w, err)
			return
		}
		w.Header().Set("Content-Type", p.PhotoType)
		if r.Method != "HEAD" {
			_, _ = w.Write(p.Photo)
		}
	})
	personWrite := func(w http.ResponseWriter, r *http.Request) {
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if media != "multipart/form-data" {
			writeJSON(w, 415, map[string]any{"errors": domain.ValidationErrors{"form": "Multipart-Formular erwartet."}})
			return
		}
		p, keep, err := readPerson(w, r)
		if err != nil {
			apiError(w, err)
			return
		}
		p, err = s.SavePerson(r.PathValue("id"), p, keep)
		if err != nil {
			apiError(w, err)
			return
		}
		status := 200
		if r.Method == "POST" {
			status = 201
			w.Header().Set("Location", "/api/personen/"+p.ID)
		}
		writeJSON(w, status, exposePerson(p))
	}
	mux.HandleFunc("POST /api/personen", personWrite)
	mux.HandleFunc("PUT /api/personen/{id}", personWrite)
}

// readMultipart reads a bounded form directly into memory, never temporary files.
func readMultipart(w http.ResponseWriter, r *http.Request, allowed []string) (map[string]string, []byte, string, bool, error) {
	fields := map[string]string{}
	seen := map[string]bool{}
	var photo []byte
	media := ""
	fail := func(err error) (map[string]string, []byte, string, bool, error) { return nil, nil, "", false, err }
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxPhotoBytes+64*1024)
	reader, err := r.MultipartReader()
	if err != nil {
		return fail(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fail(err)
		}
		name := part.FormName()
		if seen[name] {
			return fail(domain.ValidationErrors{"form": "Doppelte Formularfelder sind ungültig."})
		}
		seen[name] = true
		valid := name == "photo"
		for _, key := range allowed {
			if name == key {
				valid = true
			}
		}
		if !valid {
			return fail(domain.ValidationErrors{"form": "Unbekanntes Formularfeld."})
		}
		limit := int64(8192)
		if name == "photo" {
			limit = domain.MaxPhotoBytes
		}
		data, err := io.ReadAll(io.LimitReader(part, limit+1))
		if err != nil {
			return fail(err)
		}
		if int64(len(data)) > limit {
			return fail(&http.MaxBytesError{Limit: domain.MaxPhotoBytes})
		}
		if name == "photo" {
			photo = data
			media, _, _ = mime.ParseMediaType(part.Header.Get("Content-Type"))
		} else {
			if part.FileName() != "" {
				return fail(domain.ValidationErrors{"form": "Unerwartete Datei."})
			}
			fields[name] = string(data)
		}
		if err := part.Close(); err != nil {
			return fail(err)
		}
	}
	return fields, photo, media, seen["photo"], nil
}

func readPerson(w http.ResponseWriter, r *http.Request) (domain.Person, bool, error) {
	fields, photo, media, seenPhoto, err := readMultipart(w, r, []string{"name", "proof", "documentType", "documentDescription", "documentNumber"})
	if err != nil {
		return domain.Person{}, false, err
	}
	p := domain.Person{Name: fields["name"], DocumentType: fields["documentType"], DocumentDescription: fields["documentDescription"], DocumentNumber: fields["documentNumber"], Photo: photo, PhotoType: media}
	switch fields["proof"] {
	case "keep":
		if seenPhoto {
			return p, false, domain.ValidationErrors{"proof": "Zum Beibehalten bitte keinen neuen Nachweis senden."}
		}
		return p, true, nil
	case "manual":
		if seenPhoto {
			return p, false, domain.ValidationErrors{"proof": "Bitte genau eine Nachweisart verwenden."}
		}
	case "photo":
		if len(photo) == 0 {
			return p, false, domain.ValidationErrors{"photo": "Bitte ein Foto auswählen."}
		}
	default:
		return p, false, domain.ValidationErrors{"proof": "Bitte eine Nachweisart wählen."}
	}
	return p, false, nil
}
