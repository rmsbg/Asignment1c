package domain

import (
	"testing"
	"time"
)

func TestCalendarDuration(t *testing.T) {
	for _, tc := range []struct {
		start, end string
		days       int
	}{
		{"2028-02-28", "2028-03-01", 3}, {"2026-12-31", "2027-01-01", 2},
		{"2026-03-28", "2026-03-30", 3}, {"2026-10-24", "2026-10-26", 3},
		{"2026-10-21", "2026-10-21", 1}, {"0001-01-01", "9999-12-31", 3652059},
	} {
		t.Run(tc.start+"/"+tc.end, func(t *testing.T) {
			days, err := PeriodDays(tc.start, tc.end, tc.start)
			if err != nil || days != tc.days {
				t.Fatalf("days=%d err=%v", days, err)
			}
		})
	}
	for _, dates := range [][2]string{{"2026-02-29", "2026-03-01"}, {"2026-1-01", "2026-01-02"}, {"2026-10-20", "2026-10-21"}, {"2026-10-22", "2026-10-21"}, {"", ""}, {"0000-01-01", "2026-10-21"}, {"10000-01-01", "10000-01-02"}} {
		if _, err := PeriodDays(dates[0], dates[1], "2026-10-21"); err == nil {
			t.Fatalf("accepted %v", dates)
		}
	}
	utc := time.Date(2026, 10, 20, 22, 30, 0, 0, time.UTC)
	if Today(utc) != "2026-10-21" {
		t.Fatal("business date is not Vienna")
	}
	if Today(time.Date(2026, 12, 20, 23, 30, 0, 0, time.UTC)) != "2026-12-21" {
		t.Fatal("winter business date incorrect")
	}
}
func TestTariffPrices(t *testing.T) {
	tariff := Tariff{Version: 7, PricesCents: []int64{2500, 4500, 6000}, ExtraDayCents: 1500}
	for days, want := range map[int]int64{1: 2500, 2: 4500, 3: 6000, 4: 7500, 8: 13500} {
		got, err := tariff.Price(days)
		if err != nil || got.TotalPriceCents != want || got.DurationDays != days || got.TariffVersion != 7 {
			t.Fatalf("quote=%+v err=%v", got, err)
		}
	}
	for _, bad := range []Tariff{{}, {PricesCents: []int64{-1}}, {PricesCents: []int64{1}, ExtraDayCents: -1}, {PricesCents: []int64{MaxMoneyCents + 1}}} {
		if bad.Validate() == nil {
			t.Fatal("invalid tariff accepted")
		}
	}
	limit := Tariff{PricesCents: []int64{MaxMoneyCents}, ExtraDayCents: 1}
	if _, err := limit.Price(1); err != nil {
		t.Fatal(err)
	}
	if _, err := limit.Price(2); err == nil {
		t.Fatal("price overflow accepted")
	}
	if _, err := tariff.Price(0); err == nil {
		t.Fatal("zero duration accepted")
	}
	if _, err := (Tariff{PricesCents: []int64{0}}).Price(4); err != nil {
		t.Fatal("zero price rejected")
	}
}
