package store

import "bikerental/internal/domain"

type Conflict struct{ Field, Message string }

func (e Conflict) Error() string { return e.Message }

func cloneTariff(t domain.Tariff) domain.Tariff {
	result := domain.Tariff{Version: t.Version, Types: map[string]domain.Rate{}}
	for kind, rate := range t.Types {
		rate.PricesCents = append([]int64{}, rate.PricesCents...)
		result.Types[kind] = rate
	}
	return result
}
func (s *Store) Tariff() domain.Tariff {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneTariff(s.tariff)
}
func (s *Store) SaveTariff(t domain.Tariff) (domain.Tariff, error) {
	if err := t.Validate(); err != nil {
		return domain.Tariff{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.Version != s.tariff.Version {
		return domain.Tariff{}, Conflict{"version", "Der Tarif wurde inzwischen geändert. Bitte neu laden und erneut prüfen."}
	}
	t.Version++
	s.tariff = cloneTariff(t)
	return cloneTariff(t), nil
}
func (s *Store) quoteLocked(kind, mode, start, end string) (domain.Quote, error) {
	if !domain.ValidBikeType(kind) {
		return domain.Quote{}, domain.ValidationErrors{"bikeType": "Bitte Downhill, MTB oder Rennrad auswählen."}
	}
	period, err := domain.BuildPeriod(mode, start, end, s.now())
	if err != nil {
		return domain.Quote{}, err
	}
	if s.tariff.Version == 0 {
		return domain.Quote{}, domain.ValidationErrors{"tariffVersion": "Bitte zuerst alle drei Fahrradtarife einrichten."}
	}
	rate := s.tariff.Types[kind]
	units := period.DurationDays
	if mode == "hour" {
		units = period.DurationHours
	}
	price, err := rate.Price(mode, units)
	if err != nil {
		return domain.Quote{}, err
	}
	return domain.Quote{BikeType: kind, Mode: mode, DurationDays: period.DurationDays, DurationHours: period.DurationHours, StartsAt: period.StartsAt, EndsAt: period.EndsAt, TotalPriceCents: price, TariffVersion: s.tariff.Version}, nil
}
func (s *Store) Quote(kind, mode, start, end string) (domain.Quote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.quoteLocked(kind, mode, start, end)
}
