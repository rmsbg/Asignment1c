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
		tariff, err = s.SaveTariff(testTariff(0, []int64{2500}, 1500))
		if err != nil {
			return RentalDetails{}, err
		}
	}
	today := domain.Today(s.now())
	r, err := createDay(s, person, bike, today, today, tariff.Version)
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

func testTariff(version uint64, prices []int64, extra int64) domain.Tariff {
	types := map[string]domain.Rate{}
	for _, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		types[kind] = domain.Rate{HourlyCents: 1000, PricesCents: append([]int64{}, prices...), ExtraDayCents: extra}
	}
	return domain.Tariff{Version: version, Types: types}
}
func createDay(s *Store, person, bike, start, end string, version uint64) (RentalDetails, error) {
	kind := "MTB"
	if b, err := s.Bike(bike); err == nil {
		kind = b.Type
	}
	return s.CreateRental(domain.BookingInput{PersonID: person, BikeID: bike, BikeType: kind, Mode: "day", StartsOn: start, EndsOn: end, TariffVersion: version})
}
