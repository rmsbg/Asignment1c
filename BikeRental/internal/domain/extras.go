package domain

// Extras is an explicit selection; each option can occur at most once per booking.
type Extras struct {
	Helmet     bool `json:"helmet"`
	Protection bool `json:"protection"`
	Insurance  bool `json:"insurance"`
}

// ExtraPrices is a value snapshot, independent of mutable tariff collections.
type ExtraPrices struct {
	HelmetCents     int64 `json:"helmetCents"`
	ProtectionCents int64 `json:"protectionCents"`
	InsuranceCents  int64 `json:"insuranceCents"`
}

func (p ExtraPrices) Validate() error {
	fields := ValidationErrors{}
	for key, value := range map[string]int64{
		"extras.helmetCents":     p.HelmetCents,
		"extras.protectionCents": p.ProtectionCents,
		"extras.insuranceCents":  p.InsuranceCents,
	} {
		if value < 0 || value > MaxMoneyCents {
			fields[key] = "Extrapauschalen müssen nichtnegative ganze Cent im zulässigen Bereich sein."
		}
	}
	if len(fields) > 0 {
		return fields
	}
	return nil
}

// AddTo checks every addition, including the base price, before producing a quote.
func (p ExtraPrices) AddTo(base int64, choice Extras) (charged ExtraPrices, extras, total int64, err error) {
	if err = p.Validate(); err != nil {
		return
	}
	if base < 0 || base > MaxMoneyCents {
		err = ValidationErrors{"extras": "Der Fahrradpreis liegt außerhalb des zulässigen Bereichs."}
		return
	}
	if choice.Helmet {
		charged.HelmetCents = p.HelmetCents
	}
	if choice.Protection {
		charged.ProtectionCents = p.ProtectionCents
	}
	if choice.Insurance {
		charged.InsuranceCents = p.InsuranceCents
	}
	total = base
	for _, value := range []int64{charged.HelmetCents, charged.ProtectionCents, charged.InsuranceCents} {
		if value > MaxMoneyCents-total {
			return ExtraPrices{}, 0, 0, ValidationErrors{"extras": "Der Gesamtpreis mit Zusatzleistungen ist zu groß."}
		}
		total += value
		extras += value
	}
	return
}
