package store

import "bikerental/internal/domain"

// Preserve the original lifecycle scenarios using the new book-then-issue workflow.
func createIssued(s *Store, person, ski string) (RentalDetails, error) {
	tariff := s.Tariff()
	if tariff.Version == 0 {
		var err error
		tariff, err = s.SaveTariff(domain.Tariff{PricesCents: []int64{2500}, ExtraDayCents: 1500})
		if err != nil {
			return RentalDetails{}, err
		}
	}
	today := domain.Today(s.now())
	r, err := s.CreateRental(person, ski, today, today, tariff.Version)
	if err != nil {
		return RentalDetails{}, err
	}
	return s.IssueRental(r.Rental.ID)
}
