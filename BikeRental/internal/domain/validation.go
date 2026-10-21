package domain

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

const MaxPhotoBytes = 5 * 1024 * 1024

type ValidationErrors map[string]string

func (e ValidationErrors) Error() string { return "Bitte die markierten Felder prüfen." }

func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ValidationErrors{"name": "Bitte einen Namen oder eine Bezeichnung eingeben."}
	}
	return name, nil
}

// ValidatePerson normalizes and validates a complete replacement before storage.
func ValidatePerson(p Person) (Person, error) {
	p.Name = strings.TrimSpace(p.Name)
	p.DocumentType = strings.TrimSpace(p.DocumentType)
	p.DocumentDescription = strings.TrimSpace(p.DocumentDescription)
	p.DocumentNumber = strings.TrimSpace(p.DocumentNumber)
	e := ValidationErrors{}
	if p.Name == "" {
		e["name"] = "Bitte einen Namen eingeben."
	}
	if len(p.Photo) > 0 {
		if p.DocumentType != "" || p.DocumentDescription != "" || p.DocumentNumber != "" {
			e["proof"] = "Bitte genau eine Nachweisart verwenden."
		}
		if err := ValidatePhoto(p.Photo, p.PhotoType); err != nil {
			e["photo"] = err.(ValidationErrors)["photo"]
		}
	} else {
		switch p.DocumentType {
		case "Personalausweis", "Reisepass", "Führerschein", "Sonstiger Ausweis":
		default:
			e["documentType"] = "Bitte eine gültige Ausweisart wählen."
		}
		if p.DocumentNumber == "" {
			e["documentNumber"] = "Bitte eine Ausweisnummer eingeben."
		}
		if p.DocumentType == "Sonstiger Ausweis" && p.DocumentDescription == "" {
			e["documentDescription"] = "Bitte den sonstigen Ausweis bezeichnen."
		}
		if p.DocumentType != "Sonstiger Ausweis" {
			p.DocumentDescription = ""
		}
		p.PhotoType = ""
	}
	if len(e) > 0 {
		return Person{}, e
	}
	p.Photo = bytes.Clone(p.Photo)
	return p, nil
}

// ValidatePhoto checks both the declared type and the actual decoded image.
func ValidatePhoto(data []byte, media string) error {
	message := ""
	if len(data) == 0 {
		message = "Bitte ein Foto auswählen."
	} else if len(data) > MaxPhotoBytes {
		message = "Das Foto darf höchstens 5 MiB groß sein."
	} else {
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || (format != "jpeg" && format != "png") || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 20_000_000 {
			message = "Bitte ein gültiges JPEG oder PNG mit höchstens 20 Millionen Pixeln wählen."
		} else if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
			message = "Das Foto ist beschädigt und kann nicht gelesen werden."
		} else if media != "image/"+format {
			message = "Der angegebene Dateityp stimmt nicht mit dem Bildinhalt überein."
		}
	}
	if message != "" {
		return ValidationErrors{"photo": message}
	}
	return nil
}

func ValidateBike(b Bike) (Bike, error) {
	b.Name = strings.TrimSpace(b.Name)
	b.Type = strings.TrimSpace(b.Type)
	fields := ValidationErrors{}
	if b.Name == "" {
		fields["name"] = "Bitte eine Bezeichnung eingeben."
	}
	switch b.Type {
	case "Downhill", "MTB", "Rennrad":
	default:
		fields["type"] = "Bitte Downhill, MTB oder Rennrad auswählen."
	}
	if err := ValidatePhoto(b.Photo, b.PhotoType); err != nil {
		fields["photo"] = err.(ValidationErrors)["photo"]
	}
	if len(fields) > 0 {
		return Bike{}, fields
	}
	b.Photo = bytes.Clone(b.Photo)
	b.PhotoURL = ""
	return b, nil
}
