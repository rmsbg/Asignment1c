package store

import (
	"bikerental/internal/domain"
	"bikerental/web"
)

// Preserve the original lifecycle scenarios using the new book-then-issue workflow.
func createIssued(s *Store, person, bike string) (RentalDetails, error) {
	tariff := s.Tariff()
	if tariff.Version == 0 {
		var err error
		tariff, err = s.SaveTariff(domain.Tariff{PricesCents: []int64{2500}, ExtraDayCents: 1500})
		if err != nil {
			return RentalDetails{}, err
		}
	}
	today := domain.Today(s.now())
	r, err := s.CreateRental(person, bike, today, today, tariff.Version)
	if err != nil {
		return RentalDetails{}, err
	}
	return s.IssueRental(r.Rental.ID)
}

// saveBike supplies valid inventory data for the existing lifecycle scenarios.
func saveBike(s *Store, id, name string) (domain.Bike, error) {
	photo, err := web.Files.ReadFile("assets/demo/MTB.png")
	if err != nil {
		panic(err)
	}
	return s.SaveBike(id, domain.Bike{Name: name, Type: "MTB", Photo: photo, PhotoType: "image/png"}, false)
}
