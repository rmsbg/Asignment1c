package store

import (
	"bytes"
	"reflect"
	"sync"
	"testing"

	"bikerental/internal/domain"
)

func TestBikeImagesAreAtomicAndIndependent(t *testing.T) {
	s := New(false)
	image := photo()
	input := domain.Bike{Name: " Testfahrrad ", Type: "MTB", Photo: image.Photo, PhotoType: image.PhotoType}
	b, err := s.SaveBike("", input, false)
	if err != nil || b.ID != "bike-000001" || b.Name != "Testfahrrad" {
		t.Fatalf("create: %+v, %v", b, err)
	}
	want := bytes.Clone(b.Photo)
	input.Photo[0] ^= 1
	b.Photo[0] ^= 1
	got, _ := s.Bike(b.ID)
	if !bytes.Equal(got.Photo, want) {
		t.Fatal("input or return image aliases stored data")
	}
	before := s.Snapshot()
	for _, candidate := range []domain.Bike{
		{Name: "", Type: "MTB", Photo: want, PhotoType: "image/png"},
		{Name: "Changed", Type: "Unbekannt", Photo: want, PhotoType: "image/png"},
		{Name: "Changed", Type: "MTB"},
		{Name: "Changed", Type: "MTB", Photo: []byte("broken"), PhotoType: "image/png"},
	} {
		if _, err := s.SaveBike(b.ID, candidate, false); err == nil {
			t.Fatal("invalid edit accepted")
		}
		if !reflect.DeepEqual(before, s.Snapshot()) {
			t.Fatal("invalid edit changed a record")
		}
	}
	if _, err := s.SaveBike("", domain.Bike{Name: "No image", Type: "MTB"}, true); err == nil {
		t.Fatal("keep on create accepted")
	}
	if _, err := s.SaveBike(b.ID, domain.Bike{Name: "Mixed", Type: "MTB", Photo: want, PhotoType: "image/png"}, true); err == nil {
		t.Fatal("keep with replacement accepted")
	}
	for _, bikes := range [][]domain.Bike{s.Bikes(), s.AvailableBikes(), s.Snapshot().Bikes} {
		bikes[0].Photo[0] ^= 1
	}
	available, err := s.AvailableFor("day", domain.Today(s.now()), domain.Today(s.now()))
	if err != nil {
		t.Fatal(err)
	}
	available[0].Photo[0] ^= 1
	got, _ = s.Bike(b.ID)
	if !bytes.Equal(got.Photo, want) {
		t.Fatal("list or availability image aliases stored data")
	}
	kept, err := s.SaveBike(b.ID, domain.Bike{Name: "Renamed", Type: "Rennrad"}, true)
	if err != nil || kept.ID != b.ID || !bytes.Equal(kept.Photo, want) || kept.Type != "Rennrad" {
		t.Fatalf("keep: %v", err)
	}
	newImage := photo()
	replaced, err := s.SaveBike(b.ID, domain.Bike{Name: "Replaced", Type: "Downhill", Photo: newImage.Photo, PhotoType: newImage.PhotoType}, false)
	if err != nil || !bytes.Equal(replaced.Photo, newImage.Photo) {
		t.Fatal("replacement failed")
	}
	newBike, err := s.SaveBike("", domain.Bike{Name: "Same name", Type: "MTB", Photo: want, PhotoType: "image/png"}, false)
	if err != nil || newBike.ID != "bike-000002" {
		t.Fatal("failed creation consumed an ID")
	}
}

func TestConcurrentBikeImageEdits(t *testing.T) {
	s := New(true)
	var wg sync.WaitGroup
	for worker := range 8 {
		wg.Go(func() {
			for range 15 {
				var err error
				if worker%2 == 0 {
					_, err = saveBike(s, "demo-bike-01", "Replacement")
				} else {
					_, err = s.SaveBike("demo-bike-01", domain.Bike{Name: "Keep", Type: "Downhill"}, true)
				}
				if err != nil {
					t.Error(err)
				}
				b, err := s.Bike("demo-bike-01")
				if err != nil || domain.ValidatePhoto(b.Photo, b.PhotoType) != nil {
					t.Error("incomplete concurrent image")
				}
			}
		})
	}
	wg.Wait()
}
