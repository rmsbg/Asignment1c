package domain

import "testing"

func TestExtraPauschalenAndOverflow(t *testing.T) {
	fees := ExtraPrices{HelmetCents: 500, ProtectionCents: 1000, InsuranceCents: 800}
	wants := []int64{0, 500, 1000, 1500, 800, 1300, 1800, 2300}
	for mask, want := range wants {
		choice := Extras{Helmet: mask&1 != 0, Protection: mask&2 != 0, Insurance: mask&4 != 0}
		charged, sum, total, err := fees.AddTo(4000, choice)
		if err != nil || sum != want || total != 4000+want {
			t.Fatalf("mask %d: charged=%+v sum=%d total=%d err=%v", mask, charged, sum, total, err)
		}
		if !choice.Helmet && charged.HelmetCents != 0 || !choice.Protection && charged.ProtectionCents != 0 || !choice.Insurance && charged.InsuranceCents != 0 {
			t.Fatal("unselected option charged")
		}
	}
	for _, fees := range []ExtraPrices{{HelmetCents: -1}, {ProtectionCents: MaxMoneyCents + 1}, {InsuranceCents: -1}} {
		if fees.Validate() == nil {
			t.Fatal("invalid fees accepted")
		}
	}
	if _, _, _, err := fees.AddTo(MaxMoneyCents-499, Extras{Helmet: true}); err == nil {
		t.Fatal("base plus fee overflow accepted")
	}
	if _, sum, total, err := fees.AddTo(MaxMoneyCents-500, Extras{Helmet: true}); err != nil || sum != 500 || total != MaxMoneyCents {
		t.Fatal("exact money boundary rejected", err)
	}
	if _, _, _, err := (ExtraPrices{HelmetCents: MaxMoneyCents, ProtectionCents: 1}).AddTo(0, Extras{Helmet: true, Protection: true}); err == nil {
		t.Fatal("multiple fee overflow accepted")
	}
	if _, _, total, err := (ExtraPrices{}).AddTo(MaxMoneyCents, Extras{Helmet: true}); err != nil || total != MaxMoneyCents {
		t.Fatal("free selected option rejected", err)
	}
}
