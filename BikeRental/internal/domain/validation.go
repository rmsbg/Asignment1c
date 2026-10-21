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
		if len(p.Photo) > MaxPhotoBytes {
			e["photo"] = "Das Foto darf höchstens 5 MiB groß sein."
		} else {
			cfg, format, err := image.DecodeConfig(bytes.NewReader(p.Photo))
			if err != nil || (format != "jpeg" && format != "png") || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 20_000_000 {
				e["photo"] = "Bitte ein gültiges JPEG oder PNG mit höchstens 20 Millionen Pixeln wählen."
			} else if _, _, err := image.Decode(bytes.NewReader(p.Photo)); err != nil {
				e["photo"] = "Das Foto ist beschädigt und kann nicht gelesen werden."
			} else if p.PhotoType != "image/"+format {
				e["photo"] = "Der angegebene Dateityp stimmt nicht mit dem Bildinhalt überein."
			}
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
