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
		fields = domain.ValidationErrors{"skiID": store.ErrUnavailable.Error()}
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
	mux.HandleFunc("GET /api/ski", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.Skis()) })
	mux.HandleFunc("GET /api/ski/{id}", func(w http.ResponseWriter, r *http.Request) {
		p, err := s.Ski(r.PathValue("id"))
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, p)
	})
	skiWrite := func(w http.ResponseWriter, r *http.Request) {
		media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if media != "application/json" {
			writeJSON(w, 415, map[string]any{"errors": domain.ValidationErrors{"form": "JSON erwartet."}})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
		var input struct {
			Name string `json:"name"`
		}
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&input); err != nil {
			apiError(w, err)
			return
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			if err == nil {
				err = errors.New("extra JSON")
			}
			apiError(w, err)
			return
		}
		p, err := s.SaveSki(r.PathValue("id"), input.Name)
		if err != nil {
			apiError(w, err)
			return
		}
		status := 200
		if r.Method == "POST" {
			status = 201
			w.Header().Set("Location", "/api/ski/"+p.ID)
		}
		writeJSON(w, status, p)
	}
	mux.HandleFunc("POST /api/ski", skiWrite)
	mux.HandleFunc("PUT /api/ski/{id}", skiWrite)
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

func readPerson(w http.ResponseWriter, r *http.Request) (domain.Person, bool, error) {
	var p domain.Person
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxPhotoBytes+64*1024)
	reader, err := r.MultipartReader()
	if err != nil {
		return p, false, err
	}
	fields := map[string]string{}
	seen := map[string]bool{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return p, false, err
		}
		name := part.FormName()
		if seen[name] {
			return p, false, domain.ValidationErrors{"form": "Doppelte Formularfelder sind ungültig."}
		}
		seen[name] = true
		limit := int64(8192)
		if name == "photo" {
			limit = domain.MaxPhotoBytes + 1
		}
		data, err := io.ReadAll(io.LimitReader(part, limit+1))
		if err != nil {
			return p, false, err
		}
		if int64(len(data)) > limit || (name == "photo" && len(data) > domain.MaxPhotoBytes) {
			return p, false, &http.MaxBytesError{Limit: domain.MaxPhotoBytes}
		}
		if name == "photo" {
			p.Photo = data
			p.PhotoType, _, _ = mime.ParseMediaType(part.Header.Get("Content-Type"))
		} else {
			switch name {
			case "name", "proof", "documentType", "documentDescription", "documentNumber":
				fields[name] = string(data)
			default:
				return p, false, domain.ValidationErrors{"form": "Unbekanntes Formularfeld."}
			}
			if part.FileName() != "" {
				return p, false, domain.ValidationErrors{"form": "Unerwartete Datei."}
			}
		}
		_ = part.Close()
	}
	p.Name = fields["name"]
	p.DocumentType = fields["documentType"]
	p.DocumentDescription = fields["documentDescription"]
	p.DocumentNumber = fields["documentNumber"]
	switch fields["proof"] {
	case "keep":
		if seen["photo"] {
			return p, false, domain.ValidationErrors{"proof": "Zum Beibehalten bitte keinen neuen Nachweis senden."}
		}
		return p, true, nil
	case "manual":
		if seen["photo"] {
			return p, false, domain.ValidationErrors{"proof": "Bitte genau eine Nachweisart verwenden."}
		}
	case "photo":
		if len(p.Photo) == 0 {
			return p, false, domain.ValidationErrors{"photo": "Bitte ein Foto auswählen."}
		}
	default:
		return p, false, domain.ValidationErrors{"proof": "Bitte eine Nachweisart wählen."}
	}
	return p, false, nil
}
