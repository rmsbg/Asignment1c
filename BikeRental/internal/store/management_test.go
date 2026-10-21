package store

import (
	"bytes"
	"image"
	"image/png"
	"reflect"
	"sync"
	"testing"

	"bikerental/internal/domain"
)

func manual() domain.Person {
	return domain.Person{Name: " Test Person ", DocumentType: "Reisepass", DocumentNumber: " TEST-1 "}
}
func photo() domain.Person {
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	return domain.Person{Name: "Photo Person", Photo: b.Bytes(), PhotoType: "image/png"}
}
func TestManagementValidationAndAtomicReplacement(t *testing.T) {
	s := New(false)
	for _, p := range []domain.Person{{}, {Name: "X"}, {Name: "X", DocumentType: "Andere", DocumentNumber: "X"}, {Name: "X", DocumentType: "Sonstiger Ausweis", DocumentNumber: "X"}} {
		if _, err := s.SavePerson("", p, false); err == nil {
			t.Fatal("accepted invalid person")
		}
	}
	if len(s.People()) != 0 {
		t.Fatal("invalid creation changed state")
	}
	p, err := s.SavePerson("", manual(), false)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Test Person" || p.DocumentNumber != "TEST-1" {
		t.Fatal("missing trim")
	}
	for _, kind := range []string{"Personalausweis", "Reisepass", "Führerschein", "Sonstiger Ausweis"} {
		candidate := manual()
		candidate.DocumentType = kind
		candidate.DocumentDescription = " Testausweis "
		updated, err := s.SavePerson(p.ID, candidate, false)
		if err != nil || updated.ID != p.ID {
			t.Fatalf("manual type failed: %v", err)
		}
	}
	p, _ = s.Person(p.ID)
	if _, err := s.SavePerson(p.ID, domain.Person{Name: " "}, true); err == nil {
		t.Fatal("accepted blank name")
	}
	unchanged, _ := s.Person(p.ID)
	if !reflect.DeepEqual(p, unchanged) {
		t.Fatal("failed edit changed state")
	}
	updated, err := s.SavePerson(p.ID, domain.Person{Name: " Neuer Name "}, true)
	if err != nil || updated.DocumentNumber != p.DocumentNumber || updated.Name != "Neuer Name" {
		t.Fatal("keep failed")
	}
	pic := photo()
	updated, err = s.SavePerson(p.ID, pic, false)
	if err != nil || len(updated.Photo) == 0 || updated.DocumentNumber != "" {
		t.Fatalf("photo switch: %v", err)
	}
	pic.Photo[0] ^= 1
	updated.Photo[1] ^= 1
	got, _ := s.Person(p.ID)
	if got.Photo[0] != 137 {
		t.Fatal("input/output photo aliased")
	}
	got.Photo[0] = 0
	snapshot := s.Snapshot()
	snapshot.People[0].Photo[0] = 0
	got, _ = s.Person(p.ID)
	if got.Photo[0] != 137 {
		t.Fatal("read photo aliased")
	}
	if _, err := s.SavePerson(p.ID, domain.Person{Name: "X", Photo: []byte("broken"), PhotoType: "image/png"}, false); err == nil {
		t.Fatal("broken image accepted")
	}
	gotAgain, _ := s.Person(p.ID)
	if !reflect.DeepEqual(got, gotAgain) {
		t.Fatal("failed proof replacement mutated state")
	}
	kept, err := s.SavePerson(p.ID, domain.Person{Name: "Renamed"}, true)
	if err != nil || !bytes.Equal(kept.Photo, got.Photo) {
		t.Fatal("keep photo failed")
	}
	updated, err = s.SavePerson(p.ID, manual(), false)
	if err != nil || len(updated.Photo) != 0 || updated.PhotoType != "" {
		t.Fatal("manual switch retained photo")
	}
	summary := s.People()[0]
	if summary.DocumentNumber != "" || len(summary.Photo) != 0 {
		t.Fatal("summary exposes proof")
	}
	for _, name := range []string{"", "  "} {
		if _, err := s.SaveSki("", name); err == nil {
			t.Fatal("blank ski accepted")
		}
	}
	ski, err := s.SaveSki("", " Ski ")
	if err != nil || ski.Name != "Ski" {
		t.Fatal("ski creation")
	}
	changed, err := s.SaveSki(ski.ID, " Changed ")
	if err != nil || changed.ID != ski.ID || changed.Name != "Changed" {
		t.Fatal("ski edit")
	}
	if _, err := s.SaveSki(ski.ID, " "); err == nil {
		t.Fatal("invalid edit accepted")
	}
	gotSki, _ := s.Ski(ski.ID)
	if gotSki != changed {
		t.Fatal("invalid ski edit mutated state")
	}
	if _, err := s.Person("missing"); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := s.Ski("missing"); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := s.SavePerson("missing", manual(), false); err != ErrNotFound {
		t.Fatal(err)
	}
	if _, err := s.SaveSki("missing", "X"); err != ErrNotFound {
		t.Fatal(err)
	}
}
func TestConcurrentManagement(t *testing.T) {
	s := New(true)
	var wg sync.WaitGroup
	for range 24 {
		wg.Go(func() {
			for range 15 {
				p, err := s.SavePerson("", manual(), false)
				if err != nil {
					t.Error(err)
					return
				}
				if _, err := s.SavePerson(p.ID, domain.Person{Name: "Edited"}, true); err != nil {
					t.Error(err)
				}
				ski, err := s.SaveSki("", "Same name")
				if err != nil {
					t.Error(err)
					return
				}
				if _, err := s.SaveSki(ski.ID, "Edited"); err != nil {
					t.Error(err)
				}
				_, _ = s.Person(p.ID)
				_, _ = s.Ski(ski.ID)
				_ = s.Snapshot()
				_ = s.People()
				_ = s.Skis()
				s.SeedDemo()
			}
		})
	}
	wg.Wait()
	snap := s.Snapshot()
	if len(snap.People) != 361 || len(snap.Skis) != 362 {
		t.Fatal("lost records / duplicate IDs")
	}
}

func TestConcurrentEditsPreserveCompleteProof(t *testing.T) {
	s := New(false)
	p, err := s.SavePerson("", manual(), false)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for worker := range 12 {
		wg.Go(func() {
			for range 30 {
				if worker%2 == 0 {
					_, localErr := s.SavePerson(p.ID, manual(), false)
					if localErr != nil {
						t.Error(localErr)
					}
				} else {
					_, localErr := s.SavePerson(p.ID, domain.Person{Name: "Renamed"}, true)
					if localErr != nil {
						t.Error(localErr)
					}
				}
				got, readErr := s.Person(p.ID)
				if readErr != nil || got.ID != p.ID || got.Name == "" || got.DocumentNumber == "" {
					t.Error("incomplete concurrent edit")
				}
			}
		})
	}
	wg.Wait()
}
