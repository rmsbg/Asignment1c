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


## Entwicklungsstand – Phase 3

**Phase 3 ist implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. [Prüfnachweis und Screenshots](specs/03-zusatzleistungen.md).**
Phase 0 ist mit `28aa829`, Phase 1 mit `b8901d2` und die ausdrücklich abgenommene
Phase 2 mit `be2d6de` gesichert.

Die Arbeitskopie liegt in `BikeRental/` (Go-Modul `bikerental`) und baut auf
SkiRental auf. Fahrräder mit Typ und Pflichtbild werden stundenweise oder tageweise
zu typabhängigen Preisen gebucht. Helm, Schutzausrüstung und Versicherung sind
jetzt unabhängig wählbare Zusatzleistungen mit Pauschalen pro Buchung.

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

Normalstart: leerer Store ohne Tarife. Demo: je ein Downhill-, MTB- und Rennrad
mit Illustration, eine synthetische Person und folgende Beispielpreise:

| Typ | Stunde | Gesamtpreis 1 / 2 / 3 Tage | Zusatztag |
| --- | --- | --- | --- |
| Downhill | 12 EUR | 40 / 70 / 95 EUR | 25 EUR |
| MTB | 8 EUR | 25 / 45 / 60 EUR | 15 EUR |
| Rennrad | 10 EUR | 35 / 60 / 80 EUR | 20 EUR |

Demo-Extrapauschalen je Buchung: **Helm 5 EUR**, **Schutzausrüstung 10 EUR**,
**Versicherung 8 EUR**. Diese Beträge sind synthetische Beispielpreise.

Alle Änderungen und Uploads gehen beim Neustart verloren. Im Normalmodus vor der
ersten Buchung unter **Tarife verwalten** die drei vollständigen Tarifstaffeln
inklusive Stundenpreis und alle drei Extrapauschalen einrichten. Alle EUR-Beträge akzeptieren Komma oder Punkt
und höchstens zwei Nachkommastellen; die Preise sind bearbeitbar. Eine ausdrücklich
eingerichtete Pauschale von 0 EUR ist zulässig.

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
Kunden. Unter **Neue Ausleihe** Person, Fahrradtyp und Mietart wählen, dann Beginn
und geplante Rückgabe eingeben. Unter **Zusatzleistungen** Helm, Schutzausrüstung
und/oder Versicherung auswählen; alle Optionen sind zunächst abgewählt.
**Verfügbarkeit und Preis prüfen** zeigt Fahrradpreis, Einzelpauschalen, Extrasumme
und Gesamtpreis sowie passende freie Fahrräder mit Bild, Name, Typ und ID. Auswahl per Klick oder
Tabulator/Leertaste/Pfeiltasten. Änderungen an Zeit, Typ, Mietart oder Extras verwerfen Auswahl und Preisvorschau
und verlangen eine erneute Prüfung.

- **Stundenmiete:** Eingabe minutengenau in Wiener Zeit, mindestens eine berechnete
  Stunde, angefangene Stunden aufrunden. 10:00–11:15 kostet zwei Stunden, reserviert
  aber nur bis 11:15. Über Mitternacht möglich. **Mietbeginn jetzt** füllt die
  aktuelle Wiener Minute und eine geplante Stunde aus.
- **Tagesmiete:** Beide Randtage zählen vollständig. Montag bis Dienstag sind zwei
  Tage, reserviert bis Mittwoch 00:00. Nach der Tagesstaffel gilt der Zusatztagspreis.
- **Zeitzone:** Europe/Vienna, auch wenn der Browser eine andere Zeitzone verwendet.
  Nicht existierende/doppeldeutige Eingabezeiten bei Zeitumstellungen sind unzulässig.
  Tagespreise zählen Kalendertage; Stundenpreise die tatsächlich verstrichene Zeit.

Extrapauschalen gelten genau einmal pro Buchung, unabhängig von Stunden/Tagen
und Fahrradtyp. Ohne Auswahl zeigt die Vorschau **Keine Zusatzleistungen**.
Nach erfolgreicher Buchung ist die nächste Eingabe wieder ohne vorausgewählte
Extras. Die Versicherungsoption dokumentiert Auswahl und Aufpreis; der lokale
Prototyp führt keine externe Versicherungsabwicklung durch. Zubehörbestände
werden nicht geführt.

Nach **Zum angezeigten Preis buchen** erfolgt die Ausgabe separat über
**Fahrrad ausgeben**, später **Rückgabe erfassen**. Ausgabe nur innerhalb des
gebuchten Zeitraums und nach tatsächlicher Rückgabe des vorherigen Vorgangs.
Direkte Anschlussbuchungen sind möglich, Doppelbuchungen beider Mietarten werden
verhindert. Noch nicht ausgegebene Buchungen sind stornierbar. Abgelaufene und
überfällige Vorgänge werden markiert; **Aktualisieren** lädt deren aktuellen Stand.

Bei zwischenzeitlicher Tarif- oder Typänderung, auch nur einer Extrapauschale,
muss der neue Preis nochmals geprüft
und bestätigt werden. Bestehende Buchungen behalten den gebuchten Typ, Mietart,
Dauer, gewählte Extras, Einzelpauschalen und Gesamtpreis auch bei späterer Ausgabe
oder früher/später Rückgabe. Die Preisaufteilung bleibt in Übersicht und Verlauf
sichtbar, auch nach Stornierung.
Das Dashboard zählt tatsächlich ausgegebene Fahrräder. Die interaktiven Abläufe
benötigen JavaScript.

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
- [Phase 2: Mietzeiten, Tarife und Prüfnachweis](specs/02-mietzeiten-preise.md)
- [Phase 3: Zusatzleistungen, API und Prüfnachweis](specs/03-zusatzleistungen.md)

Nach technischem Abschluss und ausdrücklicher Auftraggeberabnahme jeder Phase
wird ein lokaler Git-Commit erstellt. Ein Implementierungsauftrag allein gilt
nicht als Abnahme.
