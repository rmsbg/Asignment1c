# BikeRental
Der Kunde hat das Verleihsystem „SkiRental“ gesehen und möchte so ein System für seinen Fahrradverleih haben.
Das System soll kostensparend entwickelt werden und daher zu möglichst großen Teilen auf „SkiRental“ aufbauen. Entwicklung neuer Features nur dort wo es notwendig ist, nach Möglichkeit Code von „SkiRental“ wiederverwenden und gegebenenfalls anpassen.


## Folgende Features und Änderungen im Vergleich zu „SkiRental“ wünscht sich der Kunde
- Fahrradverleih statt Skiverleih
- Die Fahrräder werden tageweise oder stundenweise verliehen
- Es gibt verschiedene Fahrradtypen mit verschiedenen Preisen: Downhill, MTB, Rennrad
- Es können auch Helm, Schutzausrüstung und Versicherung dazugebucht werden, das soll alles beim Verleihvorgang abgefragt werden
- Einfache Bedienbarkeit und Übersichtlichkeit sind dem Kunden sehr wichtig!
- Außerdem soll von den Fahrrädern jeweils ein Bild sichtbar sein beim Ausleihvorgang (nicht nur der Name des Rads, so wie das derzeit bei „SkiRental“ ist)


## Tech-Stack
- exakt wie bei „SkiRental“


## Entwicklungsstand – Phase 1

**Phase 1 ist implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. [Prüfnachweis und Screenshots](specs/01-fahrradverwaltung.md).**
Phase 0 ist am 21.10.2026 ausdrücklich abgenommen und mit `28aa829` gesichert.

Die Arbeitskopie liegt in `BikeRental/` (Go-Modul `bikerental`). Sie baut auf dem
SkiRental-Stand bis einschließlich dessen Phase 5 auf. Fahrräder werden jetzt mit
Typ und Pflichtbild angelegt und bei der Ausleihe mit Bild ausgewählt.
**Derzeit gilt für alle Typen dieselbe Tagesstaffel.** Stundenmiete und Preise je
Fahrradtyp folgen in Phase 2, Helm/Schutzausrüstung/Versicherung in Phase 3.

`SkiRental/` bleibt unveränderte Referenz mit eigenem Git-Repository. Das oberste
Git-Repository umfasst BikeRental und Dokumentation; die Referenz ist darin
wie vereinbart ausgeschlossen und bei einer Weitergabe separat bereitzustellen.

### Voraussetzungen und lokaler Start

Go **1.27.3 oder neuer**, für Race-Tests zusätzlich ein C-Compiler, unter macOS
beispielsweise die Xcode Command Line Tools. Keine Datenbank, zusätzlichen Go-Pakete
oder Frontend-Buildwerkzeuge erforderlich.

```sh
cd BikeRental
go version
go run ./cmd/server
# Alternativ mit synthetischen Beispieldaten:
go run ./cmd/server -demo
```

Öffnen: **http://127.0.0.1:8080**, beenden mit `Ctrl+C`. Die Modi werden
nacheinander gestartet; ist Port 8080 belegt, zuerst den anderen Dienst beenden.
Der Server bindet nur an Loopback. Es gibt keine Benutzerkonten und keinen
produktiven Betrieb; ausschließlich synthetische Kundendaten verwenden.

Normalstart: leerer Store ohne Tarif. Demo: je ein Downhill-, MTB- und Rennrad mit
mitgelieferter Illustration, eine synthetische Person und der Übergangstagestarif
25/45/60 EUR für 1/2/3 Tage plus 15 EUR je Zusatztag. Alle Änderungen und Uploads
gehen beim Neustart verloren. Für Normalmodus vor dem Buchen einen Tarif einrichten.

In dieser Arbeitsumgebung liegt die geprüfte Toolchain temporär unter
`/private/tmp/skirental-toolchain/go`. Falls Go nicht im PATH liegt:

```sh
export PATH="/private/tmp/skirental-toolchain/go/bin:$PATH"
```

Für spätere Entwicklung ist eine reguläre Go-Installation erforderlich.

### Fahrräder verwalten und auswählen

Unter **Fahrräder verwalten → Fahrrad anlegen** Bezeichnung, Typ Downhill/MTB/Rennrad
und ein JPEG-/PNG-Bild angeben. Maximal 5 MiB und 20 Millionen Pixel. Nach dem
Speichern erscheinen generierte ID, Name, Typ und Bild. In der Detailansicht führt
**Bearbeiten** zum Formular: Bild beibehalten oder durch Upload ersetzen.
**Abbrechen** verwirft Eingaben; ungültige Eingaben verändern keine gespeicherten Daten.

Die Personenverwaltung bietet weiterhin manuellen Ausweisnachweis oder ein
JPEG-/PNG-Foto. Nach Registrierung führt **Fahrrad ausleihen** zum vorausgewählten
Kunden. Unter **Neue Ausleihe** Person und Datumsbereich wählen, dann **Verfügbarkeit
und Preis prüfen**. Freie Fahrräder erscheinen mit Bild, Name, Typ und ID als
Auswahlkarten. Auswahl per Klick oder Tabulator, Leertaste und Pfeiltasten.
Datumsänderungen verwerfen die bisherige Auswahl und verlangen erneute Prüfung.

Nach **Zum angezeigten Preis buchen** erfolgt die Ausgabe separat über
**Fahrrad ausgeben**, später **Rückgabe erfassen**. Offene Buchungen können vor
Ausgabe storniert werden. Beide Randtage zählen; Änderungen am Übergangstarif
ändern bestehende Preise nicht. Das Dashboard zählt tatsächlich ausgegebene Fahrräder.
Die Verwaltung und die interaktiven Abläufe benötigen JavaScript.

### Build und Prüfung

Aus `BikeRental/`:

```sh
go build ./...
go test ./...
go test -race ./...
go vet ./...
go fmt ./...
mkdir -p bin
go build -o bin/bikerental ./cmd/server
./bin/bikerental
# Alternativ: ./bin/bikerental -demo
```

HTML, CSS, JavaScript und Demoillustrationen sind eingebettet. Das Binary benötigt
beim Start keine Asset-Dateien aus dem Repository. Nach Asset-Änderungen neu starten.
Die Demoillustrationen können ohne externe Pakete mit `go run ./tools/demo-bikes`
aus ihrem mitgelieferten Zeichencode neu erzeugt werden.

### Spezifikation und Abnahme

- [Mission und Produktverfassung](specs/mission.md)
- [Technische Verfassung](specs/tech-stack.md)
- [Roadmap](specs/roadmap.md)
- [Phase 0: Grundlagen und Prüfnachweis](specs/00-grundlagen.md)
- [Phase 1: Fahrradverwaltung, API und Prüfnachweis](specs/01-fahrradverwaltung.md)

Nach technischem Abschluss und ausdrücklicher Auftraggeberabnahme jeder Phase
wird ein lokaler Git-Commit erstellt. Ein Implementierungsauftrag allein gilt
nicht als Abnahme.
