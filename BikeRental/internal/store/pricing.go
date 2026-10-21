package store

import "bikerental/internal/domain"

type Conflict struct{ Field, Message string }

func (e Conflict) Error() string { return e.Message }

func cloneTariff(t domain.Tariff) domain.Tariff {
	t.PricesCents = append([]int64{}, t.PricesCents...)
	return t
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
func (s *Store) quoteLocked(start, end string) (domain.Quote, error) {
	days, err := domain.PeriodDays(start, end, domain.Today(s.now()))
	if err != nil {
		return domain.Quote{}, err
	}
	if s.tariff.Version == 0 {
		return domain.Quote{}, domain.ValidationErrors{"tariffVersion": "Bitte zuerst einen Tarif einrichten."}
	}
	return s.tariff.Price(days)
}
func (s *Store) Quote(start, end string) (domain.Quote, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.quoteLocked(start, end)
}
