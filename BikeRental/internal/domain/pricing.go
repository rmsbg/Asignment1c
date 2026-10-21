package domain

import (
	"time"
	_ "time/tzdata" // Keep Europe/Vienna available in standalone builds.
)

var businessZone = func() *time.Location {
	zone, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		panic(err)
	}
	return zone
}()

func Today(now time.Time) string { return now.In(businessZone).Format(time.DateOnly) }

// PeriodDays uses UTC calendar midnights, never elapsed local 24-hour intervals.
func PeriodDays(start, end, today string) (int, error) {
	fields := ValidationErrors{}
	a, err := time.Parse(time.DateOnly, start)
	if err != nil || start < "0001-01-01" || start < today {
		fields["startsOn"] = "Bitte einen gültigen Beginn ab heute wählen."
	}
	b, err := time.Parse(time.DateOnly, end)
	if err != nil || end < "0001-01-01" || end < start {
		fields["endsOn"] = "Bitte ein gültiges Rückgabedatum ab Verleihbeginn wählen."
	}
	if len(fields) > 0 {
		return 0, fields
	}
	// Unix subtraction supports the full four-digit year range without Duration overflow.
	return int((b.Unix()-a.Unix())/86400) + 1, nil
}

const MaxMoneyCents int64 = 9007199254740991 // Exact JSON/JavaScript integer range.

// Rate contains the editable hourly price and the existing daily price ladder.
type Rate struct {
	HourlyCents   int64   `json:"hourlyCents"`
	PricesCents   []int64 `json:"pricesCents"`
	ExtraDayCents int64   `json:"extraDayCents"`
}

type Tariff struct {
	Extras  ExtraPrices     `json:"extras"`
	Version uint64          `json:"version"`
	Types   map[string]Rate `json:"types"`
}

type Quote struct {
	StartMode        string      `json:"startMode"`
	BasePriceCents   int64       `json:"basePriceCents"`
	Extras           Extras      `json:"extras"`
	ExtraPrices      ExtraPrices `json:"extraPrices"`
	ExtraRates       ExtraPrices `json:"extraRates"`
	ExtrasPriceCents int64       `json:"extrasPriceCents"`
	BikeType         string      `json:"bikeType"`
	Mode             string      `json:"mode"`
	DurationDays     int         `json:"durationDays"`
	DurationHours    int         `json:"durationHours"`
	StartsAt         time.Time   `json:"startsAt"`
	EndsAt           time.Time   `json:"endsAt"`
	TotalPriceCents  int64       `json:"totalPriceCents"`
	TariffVersion    uint64      `json:"tariffVersion"`
}

func ValidBikeType(kind string) bool { return kind == "Downhill" || kind == "MTB" || kind == "Rennrad" }

func (r Rate) Validate() error {
	fields := ValidationErrors{}
	if r.HourlyCents < 0 || r.HourlyCents > MaxMoneyCents {
		fields["hourlyCents"] = "Bitte einen nichtnegativen Stundenpreis in ganzen Cent eingeben."
	}
	if len(r.PricesCents) == 0 {
		fields["pricesCents"] = "Bitte mindestens einen Tagespreis hinterlegen."
	}
	for _, price := range r.PricesCents {
		if price < 0 || price > MaxMoneyCents {
			fields["pricesCents"] = "Tagespreise müssen nichtnegative ganze Cent im zulässigen Bereich sein."
		}
	}
	if r.ExtraDayCents < 0 || r.ExtraDayCents > MaxMoneyCents {
		fields["extraDayCents"] = "Bitte einen gültigen nichtnegativen Zusatztagspreis eingeben."
	}
	if len(fields) > 0 {
		return fields
	}
	return nil
}
func (t Tariff) Validate() error {
	fields := ValidationErrors{}
	if err := t.Extras.Validate(); err != nil {
		for key, msg := range err.(ValidationErrors) {
			fields[key] = msg
		}
	}
	for _, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		r, ok := t.Types[kind]
		if !ok {
			fields[kind] = "Bitte den Tarif für " + kind + " vollständig einrichten."
			continue
		}
		if err := r.Validate(); err != nil {
			for key, msg := range err.(ValidationErrors) {
				fields[kind+"."+key] = kind + ": " + msg
			}
		}
	}
	for kind := range t.Types {
		if !ValidBikeType(kind) {
			fields["types"] = "Es sind nur Downhill, MTB und Rennrad zulässig."
		}
	}
	if len(fields) > 0 {
		return fields
	}
	return nil
}
func (r Rate) Price(mode string, units int) (int64, error) {
	if err := r.Validate(); err != nil {
		return 0, err
	}
	if units < 1 {
		return 0, ValidationErrors{"endsOn": "Die Mietdauer muss positiv sein."}
	}
	if mode == "hour" {
		if r.HourlyCents > MaxMoneyCents/int64(units) {
			return 0, ValidationErrors{"endsOn": "Der Preis für diesen Zeitraum ist zu groß."}
		}
		return r.HourlyCents * int64(units), nil
	}
	if mode != "day" {
		return 0, ValidationErrors{"mode": "Bitte Stundenmiete oder Tagesmiete wählen."}
	}
	index := min(units, len(r.PricesCents)) - 1
	price := r.PricesCents[index]
	extra := int64(units - index - 1)
	if extra > 0 && r.ExtraDayCents > (MaxMoneyCents-price)/extra {
		return 0, ValidationErrors{"endsOn": "Der Preis für diesen Zeitraum ist zu groß."}
	}
	return price + extra*r.ExtraDayCents, nil
}

// Period unifies calendar bookings and exact local-minute bookings as UTC intervals.
type Period struct {
	StartsAt, EndsAt            time.Time
	DurationDays, DurationHours int
}

func localMinute(value, field string) (time.Time, error) {
	const layout = "2006-01-02T15:04"
	naive, err := time.Parse(layout, value)
	if err != nil || len(value) != 16 || value[:4] < "0001" {
		return time.Time{}, ValidationErrors{field: "Bitte Datum und Uhrzeit minutengenau eingeben."}
	}
	candidates := map[int64]time.Time{}
	for _, hours := range []int{-48, -24, 0, 24, 48} {
		_, offset := naive.Add(time.Duration(hours) * time.Hour).In(businessZone).Zone()
		candidate := naive.Add(-time.Duration(offset) * time.Second)
		if candidate.In(businessZone).Format(layout) == value {
			candidates[candidate.Unix()] = candidate
		}
	}
	if len(candidates) != 1 {
		return time.Time{}, ValidationErrors{field: "Diese Wiener Uhrzeit ist wegen einer Zeitumstellung nicht eindeutig oder existiert nicht. Bitte eine andere Uhrzeit wählen."}
	}
	for _, candidate := range candidates {
		return candidate.UTC(), nil
	}
	panic("unreachable local time")
}

func BuildPeriod(mode, start, end string, now time.Time) (Period, error) {
	if mode == "day" {
		days, err := PeriodDays(start, end, Today(now))
		if err != nil {
			return Period{}, err
		}
		a, _ := time.ParseInLocation(time.DateOnly, start, businessZone)
		b, _ := time.ParseInLocation(time.DateOnly, end, businessZone)
		return Period{StartsAt: a.UTC(), EndsAt: b.AddDate(0, 0, 1).UTC(), DurationDays: days}, nil
	}
	if mode != "hour" {
		return Period{}, ValidationErrors{"mode": "Bitte Stundenmiete oder Tagesmiete wählen."}
	}
	a, err := localMinute(start, "startsOn")
	if err != nil {
		return Period{}, err
	}
	b, err := localMinute(end, "endsOn")
	if err != nil {
		return Period{}, err
	}
	if a.Before(now.Truncate(time.Minute)) {
		return Period{}, ValidationErrors{"startsOn": "Der Beginn darf nicht vor der aktuellen Wiener Minute liegen."}
	}
	if !b.After(a) {
		return Period{}, ValidationErrors{"endsOn": "Die Rückgabe muss nach dem Verleihbeginn liegen."}
	}
	seconds := b.Unix() - a.Unix()
	return Period{StartsAt: a, EndsAt: b, DurationHours: int((seconds + 3599) / 3600)}, nil
}

// StartOptions distinguishes a duration starting at booking from fixed dates.
type StartOptions struct {
	StartMode   string `json:"startMode"`
	RentalHours int    `json:"rentalHours"`
}

func (o StartOptions) Name() string {
	if o.StartMode == "" {
		return "scheduled"
	}
	return o.StartMode
}

// BuildBookingPeriod resolves immediate starts only against the server clock.
func BuildBookingPeriod(mode, start, end string, options StartOptions, now time.Time) (Period, error) {
	switch options.Name() {
	case "scheduled":
		if options.RentalHours != 0 {
			return Period{}, ValidationErrors{"rentalHours": "Bei festen Terminen bitte Beginn und Ende statt einer Dauer angeben."}
		}
		return BuildPeriod(mode, start, end, now)
	case "now":
		if mode != "hour" {
			return Period{}, ValidationErrors{"startMode": "Sofortbeginn ist nur bei Stundenmiete möglich."}
		}
		if start != "" || end != "" {
			return Period{}, ValidationErrors{"startMode": "Bei Sofortbeginn keine festen Termine angeben."}
		}
		if options.RentalHours < 1 || int64(options.RentalHours) > int64((1<<63-1)/time.Hour) {
			return Period{}, ValidationErrors{"rentalHours": "Bitte eine positive ganze Stundenzahl im zulässigen Bereich angeben."}
		}
		begins := now.UTC()
		return Period{StartsAt: begins, EndsAt: begins.Add(time.Duration(options.RentalHours) * time.Hour), DurationHours: options.RentalHours}, nil
	default:
		return Period{}, ValidationErrors{"startMode": "Bitte Sofortbeginn oder feste Termine wählen."}
	}
}
