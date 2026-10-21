package store

import (
	"reflect"
	"sync"
	"testing"
)

func TestStartModes(t *testing.T) {
	empty := New(false)
	empty.SeedDemo()
	if empty.IsDemo() || len(empty.Snapshot().People) != 0 || len(empty.Snapshot().Bikes) != 0 || len(empty.Snapshot().Rentals) != 0 {
		t.Fatal("normal mode must start and remain empty")
	}
	demo := New(true)
	got := demo.Snapshot()
	if !demo.IsDemo() || len(got.People) != 1 || len(got.Bikes) != 3 || len(got.Rentals) != 0 {
		t.Fatalf("unexpected demo state: %+v", got)
	}
	person := got.People[0]
	if person.Name != "Demo Person" || person.DocumentType != "Sonstiger Ausweis" || person.DocumentDescription != "Demoausweis" || person.DocumentNumber != "DEMO-0001" {
		t.Fatalf("unexpected synthetic identity: %+v", person)
	}
	for i, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		b := got.Bikes[i]
		if b.Name != "Demo "+kind || b.Type != kind || len(b.Photo) == 0 || b.PhotoURL == "" {
			t.Fatalf("invalid demo bike: %+v", b)
		}
	}

	demo.SeedDemo()
	if !reflect.DeepEqual(got, demo.Snapshot()) {
		t.Fatal("repeated seeding changed the demo data")
	}
}

func TestSnapshotsAreIndependentAndRestartResetsState(t *testing.T) {
	s := New(true)
	original := s.Snapshot()
	copy := s.Snapshot()
	copy.People[0].DocumentNumber = "CHANGED"
	copy.Bikes[0].Name = "CHANGED"
	if !reflect.DeepEqual(original, s.Snapshot()) {
		t.Fatal("snapshot exposed mutable internal state")
	}

	// Simulate a change from a future business operation without adding a CRUD API.
	s.mu.Lock()
	person := s.people["demo-person-01"]
	person.Name = "Edited Demo"
	s.people[person.ID] = person
	delete(s.bikes, "demo-bike-02")
	s.mu.Unlock()
	s.SeedDemo()
	if s.Snapshot().People[0].Name != "Edited Demo" {
		t.Fatal("seeding overwrote an existing customer")
	}
	if !reflect.DeepEqual(original, New(true).Snapshot()) {
		t.Fatal("a restarted demo store inherited old state")
	}
	fresh := New(false).Snapshot()
	if len(fresh.People)+len(fresh.Bikes)+len(fresh.Rentals) != 0 {
		t.Fatal("a restarted normal store inherited old state")
	}
}

func TestConcurrentSeedingAndSnapshots(t *testing.T) {
	s := New(true)
	expected := s.Snapshot()
	var wg sync.WaitGroup
	for range 24 {
		wg.Go(func() {
			for range 100 {
				s.SeedDemo()
				if !reflect.DeepEqual(expected, s.Snapshot()) {
					t.Error("concurrent access changed demo state or produced duplicates")
					return
				}
			}
		})
	}
	wg.Wait()
}
