package store

// DashboardCounts is one consistent, anonymous snapshot of the current inventory.
type DashboardCounts struct {
	AvailableBikes   int `json:"availableBikes"`
	RentedBikes      int `json:"rentedBikes"`
	TotalBikes       int `json:"totalBikes"`
	RegisteredPeople int `json:"registeredPeople"`
}

func (s *Store) Dashboard() DashboardCounts {
	s.mu.RLock()
	defer s.mu.RUnlock()
	active := make(map[string]bool)
	for _, r := range s.rentals {
		if r.Active() {
			active[r.BikeID] = true
		}
	}
	return DashboardCounts{
		AvailableBikes: len(s.bikes) - len(active), RentedBikes: len(active),
		TotalBikes: len(s.bikes), RegisteredPeople: len(s.people),
	}
}
