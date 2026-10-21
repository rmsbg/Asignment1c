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

type Tariff struct {
	Version       uint64  `json:"version"`
	PricesCents   []int64 `json:"pricesCents"`
	ExtraDayCents int64   `json:"extraDayCents"`
}

type Quote struct {
	DurationDays    int    `json:"durationDays"`
	TotalPriceCents int64  `json:"totalPriceCents"`
	TariffVersion   uint64 `json:"tariffVersion"`
}

func (t Tariff) Validate() error {
	fields := ValidationErrors{}
	if len(t.PricesCents) == 0 {
		fields["pricesCents"] = "Bitte mindestens einen Tagespreis hinterlegen."
	}
	for _, price := range t.PricesCents {
		if price < 0 || price > MaxMoneyCents {
			fields["pricesCents"] = "Tagespreise müssen nichtnegative ganze Cent im zulässigen Bereich sein."
		}
	}
	if t.ExtraDayCents < 0 || t.ExtraDayCents > MaxMoneyCents {
		fields["extraDayCents"] = "Bitte einen gültigen nichtnegativen Zusatztagspreis eingeben."
	}
	if len(fields) > 0 {
		return fields
	}
	return nil
}
func (t Tariff) Price(days int) (Quote, error) {
	if err := t.Validate(); err != nil {
		return Quote{}, err
	}
	if days < 1 {
		return Quote{}, ValidationErrors{"endsOn": "Die Verleihdauer muss mindestens einen Tag betragen."}
	}
	index := min(days, len(t.PricesCents)) - 1
	price := t.PricesCents[index]
	extra := int64(days - index - 1)
	if extra > 0 && t.ExtraDayCents > (MaxMoneyCents-price)/extra {
		return Quote{}, ValidationErrors{"endsOn": "Der Preis für diesen Zeitraum ist zu groß."}
	}
	return Quote{days, price + extra*t.ExtraDayCents, t.Version}, nil
}
