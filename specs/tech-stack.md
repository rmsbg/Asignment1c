# BikeRental – Technische Verfassung

## Technische Basis

Der Kunde verlangt exakt den Technologie-Stack von SkiRental. Die vorhandene
Anwendung wird angepasst; neue Frameworks, Datenbanken oder Build-Systeme sind
für den beschriebenen Umfang nicht erforderlich.

| Bereich | Entscheidung | Begründung |
| --- | --- | --- |
| Backend | Go, vorhandene Modulvorgabe 1.27.3; `net/http` | Vorhandene Server- und Fachlogik weiterverwenden |
| Frontend | HTML, CSS, JavaScript ohne Framework | Bestehende Oberfläche gezielt anpassen |
| Auslieferung | Go-Templates und Assets mit `go:embed` | Ein ausführbares Programm ohne separaten Frontend-Server |
| Kommunikation | HTTP-API mit JSON; Multipart für Bild-/Formulardaten | Bestehende API-Muster übernehmen |
| Speicherung | In-Memory-Store mit Mutex | Bestehende Architektur ohne Datenbank oder ORM |
| Tests | `testing`, `net/http/httptest`, Race Detector | Fachregeln, HTTP und Nebenläufigkeit prüfen |
| Versionierung | Git | Änderungen und Abnahmen nachvollziehbar sichern |

## Ausgangsbasis und Projektstruktur

Die BikeRental-Verfassung und Feature-Spezifikationen liegen im obersten
`specs/`-Verzeichnis, das neue Go-Modul `bikerental` in `BikeRental/`.
Das oberste Git-Repository versioniert Arbeitskopie und Dokumentation.
`SkiRental/` bleibt unveränderte Referenz mit eigener Historie und wird im
obersten Repository ausgeschlossen. Herkunft und Wiederverwendung sind in
[Phase 0](00-grundlagen.md) festgehalten.

Seit Phase 1 verwendet `BikeRental/` Fahrradmodelle, Fahrrad-API-Felder und
Demodaten mit drei Typen und Bildern. Eine gemeinsame Tagesstaffel bleibt als
Übergangslösung bis Phase 2 erhalten; Stundenmiete und Extras sind noch nicht
implementiert. Der folgende Strukturbaum ist relativ zu `BikeRental/`;
`specs/` wird zentral eine Ebene darüber gepflegt.

Die übernommene Schichtung bleibt erhalten:

```text
cmd/server/         Programmeinstieg und HTTP-Server
internal/domain/    Personen, Fahrräder, Mietzeiten, Preise und Validierung
internal/store/     Atomare Geschäftsoperationen und flüchtiger Zustand
internal/handler/   HTTP-Eingaben, Antworten und API-Fehler
web/                Eingebettete Templates, CSS und JavaScript
../specs/           Verfassung, Feature-Spezifikationen und Prüfnachweise
testdata/           Synthetische Prüf- und Bilddaten
```

Tests bleiben als `*_test.go` neben dem Code. Handler verwenden die vorhandenen
Store-Operationen; Fachregeln werden nicht zusätzlich im Frontend implementiert.
Bestehende Tests werden fachlich angepasst und um BikeRental-Fälle erweitert.

## Wiederverwendung und gezielte Erweiterungen

Personenverwaltung einschließlich Ausweisprüfung, Serverstart, Demomodus,
Navigation, Dashboard-Aktualisierung und Buchungslebenszyklus werden übernommen.
Inventarmodell, Bezeichnungen, API-Felder, Beispiele und Tests werden konsistent
auf Fahrräder umgestellt. Interne Ski-Namen dürfen nicht ungeprüft in neue
Schnittstellen oder die sichtbare Oberfläche übernommen werden.

Das Inventarmodell erhält Fahrradtyp und Bildzuordnung. Die Tariflogik wird von
einem gemeinsamen Tagestarif auf typ- und mietartabhängige Preise erweitert.
Buchungen benötigen im Zielzustand Mietart, vereinbarten Zeitraum, Zusatzleistungen
und einen nachvollziehbaren Preisstand. Phase 1 verwendet vorerst weiterhin
Kalenderdaten und den gemeinsamen Tagestarif. Exakte Felder und Endpunkte legt die jeweilige
Feature-Spezifikation vor der Implementierung fest.

## Zeit, Preise und Konsistenz

- Geschäftskalender und lokale Eingaben richten sich nach `Europe/Vienna`.
  Ereigniszeitpunkte werden wie bisher in UTC gespeichert. Stundenmieten
  erfordern Uhrzeiten; die bisherigen reinen Datumsfelder reichen dafür nicht.
- Stunden- und Tagesmieten verwenden eine gemeinsame Verfügbarkeitsprüfung.
  Intervalle sind `[Beginn, Ende)` ohne Puffer. Tagesmieten reichen von lokaler
  Mitternacht bis Mitternacht nach dem letzten Tag. Mehrdeutige oder nicht
  existierende lokale Eingabezeiten werden abgelehnt; Stundenabrechnung verwendet
  tatsächlich verstrichene Zeit. Beispiele stehen in [Phase 0](00-grundlagen.md).
- EUR-Beträge werden als ganze Cent mit Überlaufprüfung verarbeitet. Die bisherige
  Begrenzung auf exakt in JavaScript darstellbare Ganzzahlen bleibt erhalten.
  Das Backend berechnet Grundpreis, Extras und Gesamtpreis.
- Tarifversion und Preisvorschau werden beim Buchen erneut geprüft. Änderungen
  erfordern eine aktualisierte Vorschau und erneute Bestätigung.
- Referenzen, Tarifstand und Verfügbarkeit werden unter derselben Schreibsperre
  wie die Buchungsanlage geprüft. Rückgabe und Stornierung bleiben idempotent.
  Lesezugriffe liefern unabhängige Kopien einschließlich Bilddaten.
- Buchungen bewahren die vereinbarte Preisaufteilung und ausgewählten Extras.
  Extras sind unabhängig wählbare Pauschalen pro Buchung ohne eigene Bestände.
  Eine externe Versicherungsabwicklung ist nicht Bestandteil des MVP.

## Bilder, Fehlerbehandlung und Datenschutz

Fahrradbilder und Ausweisfotos bleiben fachlich und technisch getrennt. Die
vorhandene Bildvalidierung kann wiederverwendet werden. Für Fahrrad-Uploads gilt
als technische Leitplanke JPEG/PNG mit maximal 5 MiB und 20 Millionen Pixeln;
Dateigröße und dekodierter Inhalt werden serverseitig geprüft. Bild und
Stammdaten werden bei gemeinsamen Schreiboperationen atomar validiert.

Fahrräder benötigen beim Anlegen ein Bild; es kann später per Upload ersetzt
werden. Die API ist in [Phase 1](01-fahrradverwaltung.md) spezifiziert: Multipart für
Fahrrad-Schreibzugriffe, JSON für Metadaten und separate geprüfte Bildantworten. Bilder erscheinen mit
verständlichem Alternativtext und angepasster Größe im Buchungsablauf. Fehler
dürfen keine falsche Fahrradzuordnung oder unbemerkte Teilspeicherung erzeugen.

Ausweisnachweise gehören weder in Inventarbilder noch Buchungslisten, Dashboard
oder Logs. Die vorhandenen Schutzheader und Upload-Begrenzungen bleiben erhalten.
HTTP-Antworten unterscheiden ungültige Eingaben, fehlende Datensätze und Konflikte;
die Oberfläche zeigt verständliche deutsche Fehlermeldungen.

## Lokaler Betrieb und Qualitätsregeln

Loopback-Adresse `http://127.0.0.1:8080`, leerer Normalstart und optionaler
`-demo`-Start bleiben erhalten. BikeRental-Demodaten sollen alle drei Fahrradtypen,
geeignete Bilder, synthetische Personen und nachvollziehbare Beispieltarife
abdecken. Seit Phase 1 sind die drei Typen mit Bildern vorhanden; der bisherige
gemeinsame Demotagestarif bleibt bis Phase 2 bestehen. Die Oberfläche kennzeichnet Demodaten und Datenverlust
beim Neustart.

Aus dem Go-Modulverzeichnis `BikeRental/`:

```sh
go run ./cmd/server
go run ./cmd/server -demo
go build ./...
go test ./...
go test -race ./...
go vet ./...
go fmt ./...
```

Go-Code folgt `gofmt`; zusätzliche Abhängigkeiten benötigen eine begründete
Entscheidung nach den Repository-Regeln. Es gibt keine feste Coverage-Quote.
Geänderte Regeln benötigen aussagekräftige Fach- und HTTP-Tests; Speicheränderungen
zusätzlich Race-Prüfungen. Browserprüfungen umfassen Desktop, mobile Darstellung,
Tastaturbedienung, Bilder und verständliche Preis-/Fehleranzeigen. Nach Änderungen
eingebetteter Assets ist ein Neustart erforderlich.
