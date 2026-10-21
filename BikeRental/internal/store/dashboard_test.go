package store

import (
	"reflect"
	"sync"
	"testing"

	"bikerental/internal/domain"
)

func assertCounts(t *testing.T, s *Store, want DashboardCounts) {
	t.Helper()
	if got := s.Dashboard(); got != want {
		t.Fatalf("counts: %+v want %+v", got, want)
	}
}
func TestDashboardLifecycle(t *testing.T) {
	s := New(false)
	assertCounts(t, s, DashboardCounts{})
	assertCounts(t, New(true), DashboardCounts{AvailableSkis: 2, TotalSkis: 2, RegisteredPeople: 1})
	p, err := s.SavePerson("", manual(), false)
	if err != nil {
		t.Fatal(err)
	}
	assertCounts(t, s, DashboardCounts{RegisteredPeople: 1})
	a, _ := s.SaveSki("", "A")
	b, _ := s.SaveSki("", "B")
	assertCounts(t, s, DashboardCounts{AvailableSkis: 2, TotalSkis: 2, RegisteredPeople: 1})
	first, err := createIssued(s, p.ID, a.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertCounts(t, s, DashboardCounts{AvailableSkis: 1, RentedSkis: 1, TotalSkis: 2, RegisteredPeople: 1})
	_, err = createIssued(s, p.ID, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	assertCounts(t, s, DashboardCounts{RentedSkis: 2, TotalSkis: 2, RegisteredPeople: 1})
	before := s.Dashboard()
	_, _ = createIssued(s, p.ID, a.ID)
	_, _ = createIssued(s, "missing", b.ID)
	_, _ = s.SavePerson("", domain.Person{}, false)
	_, _ = s.SaveSki("", " ")
	_, _ = s.ReturnRental("missing")
	if !reflect.DeepEqual(before, s.Dashboard()) {
		t.Fatal("failed operations changed dashboard")
	}
	_, _ = s.SavePerson(p.ID, domain.Person{Name: "Renamed"}, true)
	_, _ = s.SaveSki(a.ID, "Renamed")
	assertCounts(t, s, before)
	_, _ = s.ReturnRental(first.Rental.ID)
	assertCounts(t, s, DashboardCounts{AvailableSkis: 1, RentedSkis: 1, TotalSkis: 2, RegisteredPeople: 1})
	_, _ = s.ReturnRental(first.Rental.ID)
	assertCounts(t, s, DashboardCounts{AvailableSkis: 1, RentedSkis: 1, TotalSkis: 2, RegisteredPeople: 1})
	_, _ = createIssued(s, p.ID, a.ID)
	_, _ = s.ReturnRental(first.Rental.ID)
	assertCounts(t, s, DashboardCounts{RentedSkis: 2, TotalSkis: 2, RegisteredPeople: 1})
	copy := s.Dashboard()
	copy.TotalSkis = 999
	assertCounts(t, s, DashboardCounts{RentedSkis: 2, TotalSkis: 2, RegisteredPeople: 1})
	assertCounts(t, New(false), DashboardCounts{})
}
func TestConcurrentDashboard(t *testing.T) {
	s := New(true)
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			p, err := s.SavePerson("", manual(), false)
			if err != nil {
				t.Error(err)
				return
			}
			ski, err := s.SaveSki("", "Parallel pair")
			if err != nil {
				t.Error(err)
				return
			}
			for range 30 {
				r, err := createIssued(s, p.ID, ski.ID)
				if err != nil {
					t.Error(err)
					return
				}
				counts := s.Dashboard()
				if counts.AvailableSkis < 0 || counts.RentedSkis < 0 || counts.AvailableSkis+counts.RentedSkis != counts.TotalSkis || counts.RegisteredPeople < 1 {
					t.Error("inconsistent counts")
				}
				_, _ = s.ReturnRental(r.Rental.ID)
			}
		})
	}
	wg.Wait()
	assertCounts(t, s, DashboardCounts{AvailableSkis: 14, TotalSkis: 14, RegisteredPeople: 13})
}
