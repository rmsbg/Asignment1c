package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

type rentalResponse struct {
	ID              string     `json:"id"`
	PersonID        string     `json:"personID"`
	BikeID          string     `json:"bikeID"`
	PersonName      string     `json:"personName"`
	BikeName        string     `json:"bikeName"`
	CreatedAt       time.Time  `json:"createdAt"`
	StartsOn        string     `json:"startsOn"`
	EndsOn          string     `json:"endsOn"`
	DurationDays    int        `json:"durationDays"`
	TotalPriceCents int64      `json:"totalPriceCents"`
	TariffVersion   uint64     `json:"tariffVersion"`
	Status          string     `json:"status"`
	Overdue         bool       `json:"overdue"`
	RentedAt        *time.Time `json:"rentedAt"`
	ReturnedAt      *time.Time `json:"returnedAt"`
	CancelledAt     *time.Time `json:"cancelledAt"`
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
func exposeRental(d store.RentalDetails) rentalResponse {
	r := d.Rental
	return rentalResponse{ID: r.ID, PersonID: r.PersonID, BikeID: r.BikeID, PersonName: d.PersonName, BikeName: d.BikeName,
		CreatedAt: r.CreatedAt, StartsOn: r.StartsOn, EndsOn: r.EndsOn, DurationDays: r.DurationDays, TotalPriceCents: r.TotalPriceCents,
		TariffVersion: r.TariffVersion, Status: r.Status(d.Today), Overdue: r.Active() && r.EndsOn < d.Today,
		RentedAt: optionalTime(r.RentedAt), ReturnedAt: optionalTime(r.ReturnedAt), CancelledAt: optionalTime(r.CancelledAt)}
}

// readJSON bounds all writes and rejects fields the endpoint does not own.
func readJSON(w http.ResponseWriter, r *http.Request, input any) bool {
	media, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if media != "application/json" {
		writeJSON(w, 415, map[string]any{"errors": domain.ValidationErrors{"form": "JSON erwartet."}})
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(input); err != nil {
		apiError(w, err)
		return false
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		if err == nil {
			err = errors.New("extra JSON")
		}
		apiError(w, err)
		return false
	}
	return true
}
func registerRentals(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("GET /api/tarife", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.Tariff()) })
	mux.HandleFunc("PUT /api/tarife", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Version       *uint64  `json:"version"`
			PricesCents   []*int64 `json:"pricesCents"`
			ExtraDayCents *int64   `json:"extraDayCents"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		if input.Version == nil || input.ExtraDayCents == nil {
			apiError(w, domain.ValidationErrors{"form": "Tarifversion und Zusatztagspreis sind erforderlich."})
			return
		}
		tariff := domain.Tariff{Version: *input.Version, ExtraDayCents: *input.ExtraDayCents}
		for _, price := range input.PricesCents {
			if price == nil {
				apiError(w, domain.ValidationErrors{"pricesCents": "Jeder Tagespreis muss eine ganze Centzahl sein."})
				return
			}
			tariff.PricesCents = append(tariff.PricesCents, *price)
		}
		result, err := s.SaveTariff(tariff)
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("POST /api/ausleihen/preis", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			StartsOn string `json:"startsOn"`
			EndsOn   string `json:"endsOn"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		result, err := s.Quote(input.StartsOn, input.EndsOn)
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("GET /api/fahrraeder/verfuegbar", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Has("startsOn") || q.Has("endsOn") {
			result, err := s.AvailableFor(q.Get("startsOn"), q.Get("endsOn"))
			if err != nil {
				apiError(w, err)
				return
			}
			writeJSON(w, 200, result)
		} else {
			writeJSON(w, 200, s.AvailableBikes())
		}
	})
	mux.HandleFunc("GET /api/ausleihen", func(w http.ResponseWriter, r *http.Request) {
		result := []rentalResponse{}
		for _, detail := range s.Rentals() {
			result = append(result, exposeRental(detail))
		}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("GET /api/ausleihen/{id}", func(w http.ResponseWriter, r *http.Request) {
		result, err := s.Rental(r.PathValue("id"))
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, exposeRental(result))
	})
	mux.HandleFunc("POST /api/ausleihen", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			PersonID      string  `json:"personID"`
			BikeID        string  `json:"bikeID"`
			StartsOn      string  `json:"startsOn"`
			EndsOn        string  `json:"endsOn"`
			TariffVersion *uint64 `json:"tariffVersion"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		if _, _, err := domain.ValidateRentalReferences(input.PersonID, input.BikeID); err != nil {
			apiError(w, err)
			return
		}
		if input.TariffVersion == nil {
			apiError(w, domain.ValidationErrors{"tariffVersion": "Bitte zuerst den Preis prüfen und bestätigen."})
			return
		}
		result, err := s.CreateRental(input.PersonID, input.BikeID, input.StartsOn, input.EndsOn, *input.TariffVersion)
		if err != nil {
			apiError(w, err)
			return
		}
		w.Header().Set("Location", "/api/ausleihen/"+result.Rental.ID)
		writeJSON(w, 201, exposeRental(result))
	})
	for action, operation := range map[string]func(string) (store.RentalDetails, error){"ausgabe": s.IssueRental, "rueckgabe": s.ReturnRental, "stornierung": s.CancelRental} {
		mux.HandleFunc("PUT /api/ausleihen/{id}/"+action, func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64*1024))
			if err != nil {
				apiError(w, err)
				return
			}
			if len(body) != 0 {
				apiError(w, errors.New("unexpected action body"))
				return
			}
			result, err := operation(r.PathValue("id"))
			if err != nil {
				apiError(w, err)
				return
			}
			writeJSON(w, 200, exposeRental(result))
		})
	}
}
