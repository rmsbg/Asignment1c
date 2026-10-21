package domain

import (
	"testing"
	"time"
)

func TestImmediatePeriodValidationAndDST(t *testing.T) {
	for _, now := range []time.Time{time.Date(2026, 3, 29, 0, 45, 22, 0, time.UTC), time.Date(2026, 10, 25, 0, 45, 22, 0, time.UTC)} {
		p, err := BuildBookingPeriod("hour", "", "", StartOptions{StartMode: "now", RentalHours: 2}, now)
		if err != nil || p.StartsAt != now || p.EndsAt.Sub(p.StartsAt) != 2*time.Hour || p.DurationHours != 2 {
			t.Fatalf("wrong elapsed period: %+v %v", p, err)
		}
	}
	now := time.Date(2026, 10, 21, 8, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		mode, start, end string
		options          StartOptions
	}{
		{"day", "", "", StartOptions{StartMode: "now", RentalHours: 1}},
		{"hour", "2026-10-21T10:00", "", StartOptions{StartMode: "now", RentalHours: 1}},
		{"hour", "", "2026-10-21T11:00", StartOptions{StartMode: "now", RentalHours: 1}},
		{"hour", "", "", StartOptions{StartMode: "now"}},
		{"hour", "", "", StartOptions{StartMode: "now", RentalHours: -1}},
		{"hour", "", "", StartOptions{StartMode: "now", RentalHours: 2562048}},
		{"hour", "2026-10-21T10:00", "2026-10-21T11:00", StartOptions{StartMode: "scheduled", RentalHours: 1}},
		{"hour", "2026-10-21T10:00", "2026-10-21T11:00", StartOptions{StartMode: "unknown"}},
	} {
		if _, err := BuildBookingPeriod(tc.mode, tc.start, tc.end, tc.options, now); err == nil {
			t.Fatalf("accepted invalid timing: %+v", tc)
		}
	}
	a, err := BuildBookingPeriod("hour", "2026-10-21T10:00", "2026-10-21T11:00", StartOptions{}, now)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BuildBookingPeriod("hour", "2026-10-21T10:00", "2026-10-21T11:00", StartOptions{StartMode: "scheduled"}, now)
	if err != nil || a != b {
		t.Fatal("legacy fixed timing changed")
	}
	if _, err := BuildBookingPeriod("hour", "2026-10-21T09:59", "2026-10-21T11:00", StartOptions{}, now); err == nil {
		t.Fatal("fixed past start accepted")
	}
}
