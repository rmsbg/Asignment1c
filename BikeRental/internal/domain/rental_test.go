package domain

import (
	"testing"
	"time"
)

func TestRentalReferencesAndActivity(t *testing.T) {
	for _, ids := range [][2]string{{"", ""}, {" ", "bike"}, {"person", " "}} {
		if _, _, err := ValidateRentalReferences(ids[0], ids[1]); err == nil {
			t.Fatal("missing reference accepted")
		}
	}
	p, s, err := ValidateRentalReferences(" person ", " bike ")
	if err != nil || p != "person" || s != "bike" {
		t.Fatal("references not trimmed")
	}
	r := Rental{}
	if r.Active() {
		t.Fatal("booking counted as issued")
	}
	r.RentedAt = time.Now()
	if !r.Active() {
		t.Fatal("issued rental inactive")
	}
	r.ReturnedAt = time.Now()
	if r.Active() {
		t.Fatal("returned rental active")
	}
}
