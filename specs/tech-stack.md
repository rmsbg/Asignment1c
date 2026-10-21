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

Seit Phase 2 verwendet `BikeRental/` Stunden- und Tagesmieten mit typabhängigen
Tarifen, gemeinsamen UTC-Intervallen und festgehaltenen Buchungspreisen. Die
Fahrradverwaltung und Bilder aus Phase 1 bleiben erhalten. Seit Phase 3 ergänzen
drei feste boolesche Extraoptionen und einmalige Pauschalen die Buchung; die
Preisaufteilung wird als unabhängiger Wert gespeichert. Der folgende Strukturbaum ist relativ zu `BikeRental/`; `specs/` wird
zentral eine Ebene darüber gepflegt.

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

Das Inventarmodell enthält Fahrradtyp und Bildzuordnung. Die Tariflogik verwendet
je Typ einen Stundenpreis und eine Tagesstaffel mit Zusatztag; die Tarife werden
als vollständige Sammlung mit gemeinsamem Versionszähler geändert. Buchungen
halten Typ, Mietart, Zeitraum, normalisierte UTC-Grenzen, Dauer und Preis fest.
Zusatzleistungen ergänzen diese Daten seit Phase 3 um Auswahl, berechnete
Einzelpauschalen, Grundpreis und Extrasumme. Exakte Felder und Endpunkte legt die jeweilige
Feature-Spezifikation vor der Implementierung fest.

## Zeit, Preise und Konsistenz

- Geschäftskalender und lokale Eingaben richten sich nach `Europe/Vienna`.
  Ereigniszeitpunkte werden wie bisher in UTC gespeichert. Stundenmieten
  verwenden lokale Uhrzeiten mit Minutengenauigkeit; Eingaben mit Zeitzonenoffset
  oder Sekunden werden abgelehnt. Für Tagesmieten bleiben reine Datumsfelder.
- Stunden- und Tagesmieten verwenden eine gemeinsame Verfügbarkeitsprüfung.
  Intervalle sind `[Beginn, Ende)` ohne Puffer. Tagesmieten reichen von lokaler
  Mitternacht bis Mitternacht nach dem letzten Tag. Mehrdeutige oder nicht
  existierende lokale Eingabezeiten werden abgelehnt; Stundenabrechnung verwendet
  tatsächlich verstrichene Zeit. Beginn darf bei Stundenmiete nicht vor der aktuellen
  Minute liegen. Die serverseitige Prüfung schneidet dafür Sekunden ab. Beispiele
  stehen in [Phase 0](00-grundlagen.md), Verträge in [Phase 2](02-mietzeiten-preise.md).
- EUR-Beträge werden als ganze Cent mit Überlaufprüfung verarbeitet. Die bisherige
  Begrenzung auf exakt in JavaScript darstellbare Ganzzahlen bleibt erhalten.
  Das Backend berechnet Grundpreis, Extras und Gesamtpreis. Auch die Addition
  jeder Extrapauschale wird vor Speicherung auf Überlauf geprüft.
- Tarifversion und Preisvorschau werden beim Buchen erneut geprüft. Ein gemeinsamer
  Versionszähler gilt für die drei Fahrradtarife und alle Extrapauschalen. Änderungen
  erfordern eine aktualisierte Vorschau und erneute Bestätigung.
- Referenzen, Tarifstand und Verfügbarkeit werden unter derselben Schreibsperre
  wie die Buchungsanlage geprüft. Rückgabe und Stornierung bleiben idempotent.
  Lesezugriffe liefern unabhängige Kopien einschließlich Bilddaten.
- Buchungen bewahren die vereinbarte Preisaufteilung und ausgewählten Extras.
  Extras sind unabhängig wählbare Pauschalen pro Buchung ohne eigene Bestände.
  Eine externe Versicherungsabwicklung ist nicht Bestandteil des MVP. Die HTTP-API
  verlangt eine explizite boolesche Auswahl aller drei Optionen; fehlende/null oder
  unbekannte Felder sind ungültig. Tarifänderungen, Ausgabe, Rückgabe und Stornierung
  ändern bestehende Auswahl und Preisaufteilung nicht. Details: [Phase 3](03-zusatzleistungen.md).

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

## Dashboard und Navigation

Seit Phase 4 verknüpfen fünf Schnellzugriffe und dynamische Leerhinweise die
Verwaltung mit dem Verleihablauf. `/ausleihe#uebersicht/{id}` fokussiert die
zugehörige Buchungskarte; unbekannte oder fehlerhaft kodierte Ziele lassen die
Übersicht bedienbar. Verwaltungsfehler bieten einen erneuten Ladeversuch,
korrigierte Preisprüfungen entfernen alte Validierungsfehler.

Die Dashboard-API bleibt unverändert mit vier Kennzahlen aus einem gesperrten
Store-Snapshot. Serverseitige Anfangswerte funktionieren ohne JavaScript. Der
Browser aktualisiert alle fünf Sekunden und bei Rückkehr, pausiert verborgen
und begrenzt Abfragen auf jeweils eine mit fünf Sekunden Timeout. Ungültige
Antworten ersetzen keine gültigen Werte; die Oberfläche markiert sie als veraltet
und versucht es erneut. Zeitangaben verwenden ausdrücklich Europe/Vienna.
Details und Nachweise: [Phase 4](04-dashboard-bedienung.md).

## Lokaler Betrieb und Qualitätsregeln

Loopback-Adresse `http://127.0.0.1:8080`, leerer Normalstart und optionaler
`-demo`-Start bleiben erhalten. BikeRental-Demodaten sollen alle drei Fahrradtypen,
geeignete Bilder, synthetische Personen und nachvollziehbare Beispieltarife
abdecken. Seit Phase 2 sind je Typ Stundenpreis und Tagesstaffel eingerichtet;
konkrete synthetische Beispielbeträge stehen in README und Phase-2-Spezifikation.
Seit Phase 3 kommen Demo-Extrapauschalen von 5/10/8 EUR für Helm/Schutzausrüstung/
Versicherung hinzu. Auch diese Gebühren sind bearbeitbar und flüchtig. Die Oberfläche kennzeichnet Demodaten und Datenverlust
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
