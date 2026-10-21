# BikeRental – Phase 1: Fahrradverwaltung und Bilder

## Status und Umfang

**Status:** Implementiert und technisch geprüft am 21.10.2026.
Am 21.10.2026 ausdrücklich vom Auftraggeber abgenommen.

Grundlage: [Mission](mission.md), [Phase 0](00-grundlagen.md) und
[Roadmap](roadmap.md). Phase 1 setzt B-01/B-03/B-05 für Inventar und Bildauswahl um.
Stundenmiete, typabhängige Preise und Extras folgen in Phase 2/3. Bis dahin bleibt
die bestehende gemeinsame Tagesstaffel ausdrücklich als Übergangstarif erhalten.
SkiRental bleibt unverändert; gearbeitet wird in `BikeRental/`.

## Nutzerablauf und Daten

„Fahrräder verwalten“ zeigt Bild, ID, Name und Typ. „Fahrrad anlegen“ verlangt
Bezeichnung, Typ (Downhill, MTB, Rennrad) und genau ein JPEG-/PNG-Bild. Die ID wird
automatisch erzeugt. Details zeigen das Bild; „Bearbeiten“ erlaubt Namens- und
Typänderung sowie wahlweisen Bildersatz. Ohne neue Datei bleibt das Bild erhalten.
„Abbrechen“ verwirft Eingaben. Fehler verändern keine Daten und erhalten die
Formulareingaben; Speichern zeigt eine Bestätigung und die Detailansicht.

Ein Fahrrad enthält `id`, `name`, `type`, `photoURL` sowie intern Bildbytes und
MIME-Typ. Namen werden getrimmt, dürfen nicht leer sein und müssen nicht eindeutig
sein. Typen sind exakt die drei genannten Werte. Bilder sind Pflicht, höchstens
5 MiB und 20 Millionen Pixel groß. MIME-Deklaration, dekodierbarer JPEG-/PNG-Inhalt
und tatsächliche Größe werden geprüft. Bild- und Ausweisdaten bleiben getrennt.
Lesezugriffe dürfen keine veränderbaren internen Bildpuffer offenlegen.

## Schnittstellen und Konsistenz

Dieser Abschnitt beschreibt den abgenommenen Phase-1-Stand (`b8901d2`).
Mietzeiten und Tarifverträge werden in [Phase 2](02-mietzeiten-preise.md) erweitert.

- `GET /api/fahrraeder`: nach ID sortierte Liste; `GET /api/fahrraeder/{id}`: Detail.
  JSON enthält Bild-URL, keine eingebetteten Bildbytes.
- `POST /api/fahrraeder` / `PUT /api/fahrraeder/{id}`: Multipart mit `name`, `type`,
  `image=replace` und Datei `photo`. Beim Bearbeiten alternativ `image=keep` ohne
  Datei. Beim Anlegen ist `keep` ungültig. Kein Entfernen des Pflichtbilds.
- Anlage liefert 201 und `Location`, Bearbeitung 200. Neue IDs: `bike-000001` usw.
- `GET /api/fahrraeder/{id}/foto`: Bild mit geprüftem Content-Type; HEAD ohne Body.
- `GET /api/fahrraeder/verfuegbar` übernimmt die bestehende Terminabfrage und
  liefert dieselben Fahrradmetadaten einschließlich Bild-URL.
- Buchungsreferenzen heißen `bikeID`/`bikeName`, Dashboard-Felder
  `availableBikes`, `rentedBikes`, `totalBikes`. Ski-Endpunkte werden entfernt;
  es gibt keine externen Clients oder persistenten Daten zu migrieren.
- Die Personen-API bleibt unverändert. Unbekannte IDs: 404, falscher Medientyp:
  415, ungültige/mehrfache/unbekannte Formularfelder oder Bilder: 400, zu große
  Requests/Bilder: 413; Fehlerformat weiterhin `{ "errors": { "feld": "Text" } }`.
- Uploads werden begrenzt und ohne temporäre Dateien gelesen. Vollständige
  Validierung geht der atomaren Speicherung voraus. IDs werden nur bei Erfolg
  vergeben. Änderungen und Lesen verwenden die vorhandenen Store-Sperren.

## Darstellung und Demomodus

Nach der Zeitraumprüfung zeigt die Ausleihe verfügbare Fahrräder als beschriftete
Radio-Karten mit zugehörigem Bild, Name, Typ und ID. Die Auswahl bleibt per
Tastatur möglich. Bei Datumsänderung wird die bisherige Auswahl verworfen. Bei
Bildladefehlern wird ein sichtbarer Hinweis gezeigt; es wird kein fremdes Bild
als Ersatz zugeordnet. Fehler verhindern nicht die Lesbarkeit der übrigen Daten.

Der Demomodus enthält drei Fahrräder, je eines pro Typ, mit eigens erstellten
Fahrradillustrationen und einer synthetischen Person. Illustrationen und Herkunft
werden im Projekt mitgeliefert; keine externen Bilddienste oder neuen Abhängigkeiten.
Normalmodus bleibt leer. Daten und Uploads gehen beim Neustart verloren.

## Akzeptanzkriterien und Prüfung

1. Alle drei Typen können mit Bild angelegt, aufgelistet, angezeigt und bearbeitet
   werden. IDs sind eindeutig und bleiben bei Bearbeitung unverändert.
2. Fehlender Name/Typ/Bild, ungültige Typen, MIME-Abweichungen, beschädigte oder
   übergroße Bilder und fehlerhafte Multipart-Anfragen verändern keine Daten.
3. Bild beibehalten und ersetzen funktionieren atomar. Ein- und Ausgabepuffer
   sind unabhängig; parallele Zugriffe bestehen die Race-Prüfung.
4. Ausleihe zeigt die richtigen Bilder und Typen für freie Fahrräder; bestehende
   Tagesbuchung, Konflikte, Ausgabe und Rückgabe bleiben funktionsfähig.
5. Personenverwaltung funktioniert weiterhin mit manuellen Nachweisen und PNG/JPEG;
   Ausweisnachweise erscheinen nicht in Fahrradlisten oder Buchungsübersichten.
6. Desktop, schmaler Bildschirm und Tastaturablauf werden im Browser geprüft.
   Name/Typ sind auch ohne Bild lesbar; Fehler und Bestätigungen erhalten Fokus.
7. Build, sämtliche Go-Tests, Race und Vet bestehen. README erklärt den aktuellen
   Funktionsumfang und die noch fehlenden Folgephasen.

## Prüfnachweis und Abnahme

Technische Prüfung am **21.10.2026**, Go **1.27.3**, `darwin/arm64`.
Verwendete Toolchain: `/private/tmp/skirental-toolchain/go/bin/go`,
`GOCACHE=/private/tmp/skirental-go-cache`. Keine neuen Go-Abhängigkeiten.

| Prüfung | Ergebnis |
| --- | --- |
| `go build ./...` | Bestanden |
| `go test -count=1 ./...` | Domain, Handler und Store bestanden |
| `go test -race -count=1 ./...` | Bestanden, keine Race-Meldung |
| `go vet ./...` | Bestanden |
| JavaScript-Syntaxprüfung mit `node --check` | Verwaltung und Ausleihe bestanden |
| Browserfluss Normal-/Demomodus | Beide bestanden |
| Desktop 1440 × 1000 / mobil 390 × 844 | Kein horizontaler Überlauf, Bilder sichtbar |

**Automatisierte Fach- und HTTP-Prüfung:** Alle bisherigen Personen-, Preis-,
Buchungs-, Dashboard- und Gesamtflusstests wurden erhalten und auf Fahrrad-IDs,
Multipart-Inventar sowie drei Demo-Fahrräder angepasst. Der Gesamtfluss prüft
Normal-/Demomodus jeweils mit manuellem Nachweis, PNG und JPEG.

Ergänzende Tests prüfen alle drei Fahrradtypen, Bild-URLs und Bildantworten
(einschließlich HEAD), PNG/JPEG-Wechsel, Beibehalten, Pflichtbilder, ungültige Typen,
Namen, MIME-Abweichungen, beschädigte Bilder, 5-MiB-Grenze, Überschreitung und
25 Millionen Pixel. Unbekannte IDs, doppelte/unbekannte Formularfelder, defekte
Multipart-Bodies und unzulässige Schreibmethoden sind geprüft. Ungültige Änderungen
lassen den Snapshot unverändert. Tests prüfen unabhängige Bildpuffer bei Anlage,
Detail, Listen, Snapshot und Verfügbarkeit sowie gleichzeitigen Bildersatz und
Beibehalten. Fehlerhafte Anlagen verbrauchen keine ID.

**Browserprüfung:** Separater lokaler Chrome-Prozess über das bereits in der
Arbeitsumgebung vorhandene Playwright, ohne neue Projektabhängigkeit. Das Binary
wurde aus einem temporären Verzeichnis außerhalb des Repositorys gestartet,
jeweils mit leerem Store und mit `-demo`.

In beiden Modi geprüft: drei Fahrradtypen per Formular anlegen; fehlendes Bild
mit erhaltenen Eingaben und Fokus auf der Fehlerzusammenfassung; Name und Typ
ändern, Bild beibehalten/ersetzen und Änderungen abbrechen; Person mit manuellem
Nachweis registrieren und auf Fotobeleg wechseln; Ausleihe mit vorausgewählter
Person öffnen; passende Bilder laden; Auswahl per Leertaste/Pfeiltaste; Auswahl
nach Datumsänderung verwerfen; Tagesbuchung, Ausgabe und Rückgabe durchführen.
Das Dashboard zählt nach Rückgabe wieder alle Fahrräder als verfügbar. Bildfehler
wurden gezielt simuliert; ein lesbarer Hinweis erscheint statt eines falschen
Ersatzbilds. Keine unbehandelten JavaScript-Fehler. Alle Prüfserver und der separate
Browserprozess wurden beendet.

Screenshots mit ausschließlich synthetischen Daten:

- [Fahrradverwaltung Desktop](screenshots/01-fahrraeder-desktop.png)
- [Fahrradverwaltung mobil](screenshots/01-fahrraeder-mobil.png)
- [Bildauswahl Desktop](screenshots/01-auswahl-desktop.png)
- [Bildauswahl mobil](screenshots/01-auswahl-mobil.png)
- [Pflichtbildfehler im Formular](screenshots/01-pflichtbild-fehler.png)

Das temporäre Prüfskript liegt unter `/private/tmp/bikerental-phase1-browser.cjs`,
Serverlogs unter `/private/tmp/bikerental-phase1-checks/`. Diese Hilfsdateien und
Playwright sind keine Voraussetzungen zum Bauen oder Starten der Anwendung.
Die Screenshots und dieser Prüfnachweis sind die dauerhaften Abnahmeunterlagen.
Ein physischer Mobilgerätetest und eigener Screenreader-Test sind nicht erfolgt.

**Grenzen:** Gemeinsame Tagesstaffel als ausdrücklich angezeigter Übergangstarif;
Stundenmiete und typabhängige Preise folgen in Phase 2, Extras in Phase 3. Speicher
bleibt flüchtig; keine Löschfunktion, Datenbank oder externe Bild-/Versicherungsdienste.
SkiRental und das oberste `AGENTS.md` bleiben unverändert.

**Auftraggeberabnahme:** Am 21.10.2026 ausdrücklich erteilt:
„phase 1 ist hiermit abgenommen. implementiere phase 2“. Der abgenommene Stand
wird mit dem folgenden lokalen Commit gesichert:
`Phase 1: Fahrradverwaltung mit Bildern abschließen`.
