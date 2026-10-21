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


## Entwicklungsstand – Phase 0

**Phase 0 ist technisch abgeschlossen, geprüft und am 21.10.2026 ausdrücklich vom Auftraggeber abgenommen.**

Die separate Arbeitskopie liegt in `BikeRental/` (Go-Modul `bikerental`). Sie
übernimmt den SkiRental-Stand bis einschließlich Phase 5. **Die Fachfunktionen
arbeiten noch mit Ski und bisherigen Tagestarifen.** Jede Seite kennzeichnet
diesen Ausgangsstand. Fahrradtypen, Fahrradbilder, Stundenmiete und Extras
folgen in den nächsten Phasen.

`SkiRental/` bleibt unveränderte Referenz mit eigenem Git-Repository. Das neue
Git-Repository im obersten Verzeichnis umfasst BikeRental und diese Dokumentation;
die Referenz ist darin ausgeschlossen. Für eine Weitergabe der ursprünglichen
Referenz muss `SkiRental/` separat mitgeliefert werden.

### Voraussetzungen und lokaler Start

Go **1.27.3 oder neuer**, für Race-Tests zusätzlich ein C-Compiler, unter macOS
beispielsweise die Xcode Command Line Tools. Keine Datenbank, zusätzlichen Go-Pakete
oder Frontend-Buildwerkzeuge erforderlich.

```sh
cd BikeRental
go version
go run ./cmd/server
# Alternativ mit synthetischen SkiRental-Beispieldaten:
go run ./cmd/server -demo
```

Die Anwendung ist unter **http://127.0.0.1:8080** erreichbar; beenden mit `Ctrl+C`.
Die beiden Modi werden nacheinander gestartet. Bei belegtem Port zuerst den
anderen Dienst beenden. Der Server bindet nur an Loopback.

Normalstart: leerer Store ohne Tarif. Demostart in Phase 0: zwei Ski-Paare, eine
synthetische Person und der bisherige Beispieltagestarif. Alle Änderungen und
Uploads gehen beim Neustart verloren. Es gibt keine Benutzerkonten und keinen
produktiven Betrieb. Nur synthetische Kundendaten verwenden.

In dieser Arbeitsumgebung liegt die zur Prüfung verwendete Toolchain temporär
unter `/private/tmp/skirental-toolchain/go`. Falls Go nicht im PATH liegt:

```sh
export PATH="/private/tmp/skirental-toolchain/go/bin:$PATH"
```

Für spätere Entwicklung ist eine reguläre Go-Installation erforderlich.

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

HTML, CSS und JavaScript sind eingebettet; das Binary benötigt beim Start keine
Asset-Dateien aus dem Repository. Nach Asset-Änderungen neu starten.

### Spezifikation und Abnahme

- [Mission und Produktverfassung](specs/mission.md)
- [Technische Verfassung](specs/tech-stack.md)
- [Roadmap](specs/roadmap.md)
- [Phase 0: Entscheidungen, Wiederverwendung und Prüfnachweis](specs/00-grundlagen.md)

Nach technischem Abschluss und ausdrücklicher Auftraggeberabnahme jeder Phase
wird ein lokaler Git-Commit erstellt. Ein Implementierungsauftrag allein gilt
nicht als Abnahme.
