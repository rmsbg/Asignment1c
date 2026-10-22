package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"bikerental/internal/domain"
	"bikerental/internal/store"
)

type rentalResponse struct {
	StartMode        string             `json:"startMode"`
	BasePriceCents   int64              `json:"basePriceCents"`
	Extras           domain.Extras      `json:"extras"`
	ExtraPrices      domain.ExtraPrices `json:"extraPrices"`
	ExtrasPriceCents int64              `json:"extrasPriceCents"`
	BikeType         string             `json:"bikeType"`
	Mode             string             `json:"mode"`
	StartsAt         time.Time          `json:"startsAt"`
	EndsAt           time.Time          `json:"endsAt"`
	DurationHours    int                `json:"durationHours"`
	CanIssue         bool               `json:"canIssue"`
	ID               string             `json:"id"`
	PersonID         string             `json:"personID"`
	BikeID           string             `json:"bikeID"`
	PersonName       string             `json:"personName"`
	BikeName         string             `json:"bikeName"`
	CreatedAt        time.Time          `json:"createdAt"`
	StartsOn         string             `json:"startsOn"`
	EndsOn           string             `json:"endsOn"`
	DurationDays     int                `json:"durationDays"`
	TotalPriceCents  int64              `json:"totalPriceCents"`
	TariffVersion    uint64             `json:"tariffVersion"`
	Status           string             `json:"status"`
	Overdue          bool               `json:"overdue"`
	RentedAt         *time.Time         `json:"rentedAt"`
	ReturnedAt       *time.Time         `json:"returnedAt"`
	CancelledAt      *time.Time         `json:"cancelledAt"`
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
func exposeRental(d store.RentalDetails) rentalResponse {
	r := d.Rental
	return rentalResponse{StartMode: r.StartMode, BasePriceCents: r.BasePriceCents, Extras: r.Extras, ExtraPrices: r.ExtraPrices, ExtrasPriceCents: r.ExtrasPriceCents, BikeType: r.BikeType, Mode: r.Mode, StartsAt: r.StartsAt, EndsAt: r.EndsAt, DurationHours: r.DurationHours,
		CanIssue: r.Status(d.Now) == "booked" && !d.Now.Before(r.StartsAt) && d.Now.Before(r.EndsAt), ID: r.ID, PersonID: r.PersonID, BikeID: r.BikeID, PersonName: d.PersonName, BikeName: d.BikeName,
		CreatedAt: r.CreatedAt, StartsOn: r.StartsOn, EndsOn: r.EndsOn, DurationDays: r.DurationDays, TotalPriceCents: r.TotalPriceCents,
		TariffVersion: r.TariffVersion, Status: r.Status(d.Now), Overdue: r.Active() && !d.Now.Before(r.EndsAt),
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

type extrasInput struct {
	Helmet     *bool `json:"helmet"`
	Protection *bool `json:"protection"`
	Insurance  *bool `json:"insurance"`
}

func readExtras(input *extrasInput) (domain.Extras, error) {
	if input == nil || input.Helmet == nil || input.Protection == nil || input.Insurance == nil {
		return domain.Extras{}, domain.ValidationErrors{"extras": "Bitte Helm, Schutzausrüstung und Versicherung ausdrücklich auswählen oder abwählen."}
	}
	return domain.Extras{Helmet: *input.Helmet, Protection: *input.Protection, Insurance: *input.Insurance}, nil
}

func registerRentals(mux *http.ServeMux, s *store.Store) {
	mux.HandleFunc("GET /api/tarife", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.Tariff()) })
	mux.HandleFunc("PUT /api/tarife", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Extras *struct {
				HelmetCents     *int64 `json:"helmetCents"`
				ProtectionCents *int64 `json:"protectionCents"`
				InsuranceCents  *int64 `json:"insuranceCents"`
			} `json:"extras"`
			Version *uint64 `json:"version"`
			Types   map[string]struct {
				HourlyCents   *int64   `json:"hourlyCents"`
				PricesCents   []*int64 `json:"pricesCents"`
				ExtraDayCents *int64   `json:"extraDayCents"`
			} `json:"types"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		if input.Version == nil {
			apiError(w, domain.ValidationErrors{"version": "Tarifversion erforderlich."})
			return
		}
		if input.Extras == nil || input.Extras.HelmetCents == nil || input.Extras.ProtectionCents == nil || input.Extras.InsuranceCents == nil {
			apiError(w, domain.ValidationErrors{"extras": "Bitte alle drei Extrapauschalen in ganzen Cent einrichten."})
			return
		}
		tariff := domain.Tariff{Version: *input.Version, Types: map[string]domain.Rate{}, Extras: domain.ExtraPrices{HelmetCents: *input.Extras.HelmetCents, ProtectionCents: *input.Extras.ProtectionCents, InsuranceCents: *input.Extras.InsuranceCents}}
		for kind, rate := range input.Types {
			if rate.HourlyCents == nil || rate.ExtraDayCents == nil {
				apiError(w, domain.ValidationErrors{"types": "Stundenpreis und Zusatztagspreis sind je Fahrradtyp erforderlich."})
				return
			}
			r := domain.Rate{HourlyCents: *rate.HourlyCents, ExtraDayCents: *rate.ExtraDayCents}
			for _, price := range rate.PricesCents {
				if price == nil {
					apiError(w, domain.ValidationErrors{"types": "Jeder Tagespreis muss eine ganze Centzahl sein."})
					return
				}
				r.PricesCents = append(r.PricesCents, *price)
			}
			tariff.Types[kind] = r
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
			domain.StartOptions
			Extras   *extrasInput `json:"extras"`
			BikeType string       `json:"bikeType"`
			Mode     string       `json:"mode"`
			StartsOn string       `json:"startsOn"`
			EndsOn   string       `json:"endsOn"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		extras, err := readExtras(input.Extras)
		if err != nil {
			apiError(w, err)
			return
		}
		result, err := s.QuoteWithStart(input.BikeType, input.Mode, input.StartsOn, input.EndsOn, extras, input.StartOptions)
		if err != nil {
			apiError(w, err)
			return
		}
		writeJSON(w, 200, result)
	})
	mux.HandleFunc("GET /api/fahrraeder/verfuegbar", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Has("startsOn") || q.Has("endsOn") || q.Has("mode") || q.Has("bikeType") || q.Has("startMode") || q.Has("rentalHours") {
			kind := q.Get("bikeType")
			if q.Has("bikeType") && !domain.ValidBikeType(kind) {
				apiError(w, domain.ValidationErrors{"bikeType": "Bitte einen gültigen Fahrradtyp wählen."})
				return
			}
			options := domain.StartOptions{StartMode: q.Get("startMode")}
			if q.Has("rentalHours") {
				hours, err := strconv.Atoi(q.Get("rentalHours"))
				if err != nil {
					apiError(w, domain.ValidationErrors{"rentalHours": "Bitte eine ganze Stundenzahl angeben."})
					return
				}
				options.RentalHours = hours
			}
			result, err := s.AvailableWithStart(q.Get("mode"), q.Get("startsOn"), q.Get("endsOn"), options)
			if err != nil {
				apiError(w, err)
				return
			}
			if kind != "" {
				filtered := []domain.Bike{}
				for _, b := range result {
					if b.Type == kind {
						filtered = append(filtered, b)
					}
				}
				result = filtered
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
			domain.BookingInput
			IssueImmediately bool         `json:"issueImmediately"`
			Extras           *extrasInput `json:"extras"`
		}
		if !readJSON(w, r, &input) {
			return
		}
		if _, _, err := domain.ValidateRentalReferences(input.PersonID, input.BikeID); err != nil {
			apiError(w, err)
			return
		}

		extras, err := readExtras(input.Extras)
		if err != nil {
			apiError(w, err)
			return
		}
		input.BookingInput.Extras = extras
		create := s.CreateRental
		if input.IssueImmediately {
			create = s.CreateAndIssueRental
		}
		result, err := create(input.BookingInput)
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
