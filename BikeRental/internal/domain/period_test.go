package domain

import (
	"testing"
	"time"
)

func TestMixedPeriodCalendarAndClockRules(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		mode, start, end     string
		days, hours, elapsed int
	}{
		{"hour", "2026-10-21T10:00", "2026-10-21T10:15", 0, 1, 15 * 60},
		{"hour", "2026-10-21T10:00", "2026-10-21T11:00", 0, 1, 3600},
		{"hour", "2026-10-21T10:00", "2026-10-21T11:15", 0, 2, 75 * 60},
		{"hour", "2026-10-21T23:30", "2026-10-22T00:30", 0, 1, 3600},
		{"hour", "2027-03-28T01:30", "2027-03-28T03:30", 0, 1, 3600},
		{"hour", "2026-10-25T01:30", "2026-10-25T03:30", 0, 3, 3 * 3600},
		{"day", "2027-03-28", "2027-03-28", 1, 0, 23 * 3600},
		{"day", "2026-10-25", "2026-10-25", 1, 0, 25 * 3600},
		{"day", "2026-11-02", "2026-11-03", 2, 0, 48 * 3600},
		{"day", "2028-02-28", "2028-03-01", 3, 0, 72 * 3600},
	} {
		t.Run(tc.mode+tc.start+tc.end, func(t *testing.T) {
			p, err := BuildPeriod(tc.mode, tc.start, tc.end, now)
			if err != nil || p.DurationDays != tc.days || p.DurationHours != tc.hours || p.EndsAt.Unix()-p.StartsAt.Unix() != int64(tc.elapsed) {
				t.Fatalf("period=%+v err=%v", p, err)
			}
			if p.StartsAt.Location() != time.UTC || p.EndsAt.Location() != time.UTC {
				t.Fatal("interval is not UTC")
			}
		})
	}
	for _, tc := range [][3]string{
		{"hour", "2027-03-28T02:30", "2027-03-28T04:00"},
		{"hour", "2026-10-25T02:30", "2026-10-25T04:00"},
		{"hour", "2026-10-25T01:30", "2026-10-25T02:30"},
		{"hour", "2026-10-21T10:00", "2026-10-21T10:00"},
		{"hour", "2026-10-21T10:00", "2026-10-21T09:00"},
		{"hour", "2026-10-21T10:00Z", "2026-10-21T12:00"},
		{"hour", "2026-10-21T10:00:00", "2026-10-21T12:00"},
		{"hour", "0000-01-01T10:00", "2026-10-21T12:00"},
		{"invalid", "2026-10-21", "2026-10-22"},
		{"day", "2026-10-22", "2026-10-21"},
	} {
		if _, err := BuildPeriod(tc[0], tc[1], tc[2], now); err == nil {
			t.Fatalf("accepted invalid period: %v", tc)
		}
	}
	now = time.Date(2026, 10, 21, 9, 0, 59, 0, time.UTC)
	if _, err := BuildPeriod("hour", "2026-10-21T11:00", "2026-10-21T12:00", now); err != nil {
		t.Fatal("current minute rejected", err)
	}
	if _, err := BuildPeriod("hour", "2026-10-21T10:59", "2026-10-21T12:00", now); err == nil {
		t.Fatal("past minute accepted")
	}
}

func TestCompleteTypedTariff(t *testing.T) {
	valid := Tariff{Version: 1, Types: map[string]Rate{}}
	for i, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		valid.Types[kind] = Rate{HourlyCents: int64(i+1) * 100, PricesCents: []int64{1000}, ExtraDayCents: 500}
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	delete(valid.Types, "MTB")
	if valid.Validate() == nil {
		t.Fatal("partial tariff accepted")
	}
	valid.Types["MTB"] = Rate{PricesCents: []int64{1000}}
	valid.Types["Other"] = Rate{PricesCents: []int64{1000}}
	if valid.Validate() == nil {
		t.Fatal("unknown type accepted")
	}
}
