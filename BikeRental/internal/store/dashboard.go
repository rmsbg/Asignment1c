package store

// DashboardCounts is one consistent, anonymous snapshot of the current inventory.
type DashboardCounts struct {
	AvailableSkis    int `json:"availableSkis"`
	RentedSkis       int `json:"rentedSkis"`
	TotalSkis        int `json:"totalSkis"`
	RegisteredPeople int `json:"registeredPeople"`
}

func (s *Store) Dashboard() DashboardCounts {
	s.mu.RLock()
	defer s.mu.RUnlock()
	active := make(map[string]bool)
	for _, r := range s.rentals {
		if r.Active() {
			active[r.SkiID] = true
		}
	}
	return DashboardCounts{
		AvailableSkis: len(s.skis) - len(active), RentedSkis: len(active),
		TotalSkis: len(s.skis), RegisteredPeople: len(s.people),
	}
}
