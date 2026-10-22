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


## Entwicklungsstand – Phase 6

**Phase 5 einschließlich Sofortbeginn ist implementiert, technisch geprüft und
am 21.10.2026 ausdrücklich vom Auftraggeber abgenommen. [Anforderungsmatrix und Prüfnachweis](specs/05-mvp-abnahme.md).**
Phase 0 ist mit `28aa829`, Phase 1 mit `b8901d2` und die ausdrücklich abgenommene
Phase 2 mit `be2d6de` gesichert. Phase 3 wurde ausdrücklich abgenommen und mit
`2eb73e9` committet. Die ausdrücklich abgenommene Phase 4 ist mit `340a53b` gesichert.

Die Arbeitskopie liegt in `BikeRental/` (Go-Modul `bikerental`) und baut auf
SkiRental auf. Fahrräder mit Typ und Pflichtbild werden stundenweise oder tageweise
zu typabhängigen Preisen gebucht. Helm, Schutzausrüstung und Versicherung sind
jetzt unabhängig wählbare Zusatzleistungen mit Pauschalen pro Buchung.

`SkiRental/` bleibt unveränderte Referenz mit eigenem Git-Repository. Das oberste
Git-Repository umfasst BikeRental und Dokumentation; die Referenz ist darin
wie vereinbart ausgeschlossen und bei einer Weitergabe separat bereitzustellen.

Die Erweiterung **Buchen und ausgeben** (Phase 6) ist implementiert, technisch
geprüft und am 22.10.2026 ausdrücklich vom Auftraggeber abgenommen.
[Verhalten und Prüfnachweis](specs/07-buchen-und-ausgeben.md).

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

- **Stundenmiete:** Feste Termine werden minutengenau in Wiener Zeit eingegeben,
  angefangene Stunden aufgerundet. 10:00–11:15 kostet zwei Stunden, reserviert
  aber nur bis 11:15. Über Mitternacht möglich. Bei **Sofort beginnen** die gewünschte positive ganze Stundenzahl eingeben
  (voreingestellt 1). Die volle Dauer startet erst beim erfolgreichen Buchungsabschluss;
  die Vorschau zeigt beispielsweise **1 Stunde ab Buchungsabschluss · 8 €**.
  Die verbindlichen Uhrzeiten erscheinen nach dem Speichern, und die Ausgabe ist
  sofort möglich. Die Mietzeit läuft bereits ab Buchung, auch bei späterer Ausgabe.
  **Für später reservieren** verwendet feste Beginn-/Endtermine; die Ausgabe ist
  erst ab dem gebuchten Beginn möglich. Eingaben dürfen dort nicht in der Vergangenheit liegen.
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

Bei Tagesmiete mit Beginn heute (Europe/Vienna) und bei Stundenmiete mit
**Sofort beginnen** gibt es **Buchen und ausgeben** sowie **Nur buchen**.
Beide Aktionen bestätigen den angezeigten Preis. **Buchen und ausgeben** speichert
Buchung und tatsächliche Ausgabe gemeinsam; bei Konflikten wird nichts angelegt.
**Nur buchen** lässt die spätere Ausgabe über **Fahrrad ausgeben** offen.
Für spätere Tagesmieten und feste Stundenreservierungen gibt es **Buchen**.
Die Rückgabe erfolgt über **Rückgabe erfassen**. Der Bestätigungslink
**Zur Ausgabe und Buchungsübersicht** (bei Sofortausgabe **Zur Buchungsübersicht**) öffnet direkt die gerade angelegte Buchung
und setzt den Fokus auf deren Karte. **Buchungen und Rückgaben** ist auch direkt
vom Dashboard und Buchungsformular erreichbar. Ausgabe nur innerhalb des
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
### Dashboard und Bedienung

Das Dashboard zeigt verfügbare und tatsächlich ausgegebene Fahrräder, Gesamtbestand
und registrierte Personen. Reservierungen zählen noch nicht als Ausgabe.
Fünf Schnellzugriffe führen zu Inventar, Personen, neuer Ausleihe, Buchungen und
Rückgaben sowie Tarifen. Bei leerem Start verlinken Hinweise die ersten Schritte.

Die Kennzahlen aktualisieren sich alle fünf Sekunden und bei Rückkehr zum Tab.
Die letzte Aktualisierung zeigt Wiener Zeit. Bei Fehlern bleiben die letzten
Werte sichtbar und sind als veraltet markiert; die Aktualisierung erholt sich
automatisch oder über **Jetzt aktualisieren**. Ohne JavaScript erscheinen die
serverseitigen Anfangswerte; die interaktiven Verwaltungsabläufe benötigen JavaScript.

Fehlgeschlagene Verwaltungsabfragen bieten **Erneut laden**. Nach korrigierter
Preisprüfung verschwinden alte Fehlermarkierungen. Bei belegten Fahrrädern führt
**Rückgaben ansehen** zur Übersicht; Änderungen am Zeitraum verlangen eine neue Prüfung.

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

### Browser-Gesamtprüfung wiederholen

Das mitgelieferte `BikeRental/tools/mvp-browser.cjs` prüft 24 vollständige
Verleihabläufe (Normal/Demo × drei Typen × zwei Mietarten × mit/ohne Extras)
plus Prüfungen für verzögert bestätigten Sofortbeginn und feste Reservierung,
manuellen/fotografischen Ausweisnachweis, Bilder, Preisverlauf, Dashboard und echte
Prozessneustarts. Es baut ein Binary, startet es in einem temporären Verzeichnis,
beendet seine Server wieder und schreibt synthetische Screenshots nach `specs/screenshots/`.
Ein bereits laufender Server wird nicht beendet; Port 8080 vorher freigeben.

Nur für diese optionale Prüfung werden zusätzlich Node.js, Playwright und ein
Chromium-Browser benötigt. Geprüft mit Node **24.19.0**, Playwright **1.62.1**
und lokalem Google Chrome. Go muss im PATH liegen; alternativ `GO_BINARY` setzen.
Die Anwendung selbst benötigt diese Werkzeuge nicht.

Beispiel unter macOS, aus `BikeRental/`, Prüfabhängigkeiten außerhalb des Repositories:

```sh
npm install --prefix /tmp/bikerental-browser --no-save playwright@1.62.1
PLAYWRIGHT_MODULE=/tmp/bikerental-browser/node_modules/playwright/test \
CHROME_EXECUTABLE_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
node tools/mvp-browser.cjs
```

Alternativ mit dem von Playwright bereitgestellten Chromium:
`/tmp/bikerental-browser/node_modules/.bin/playwright install chromium`, anschließend
dieselbe Prüfung ohne `CHROME_EXECUTABLE_PATH` starten. Beide Varianten verwenden
denselben Prüflauf. Nach der Prüfung den gewünschten lokalen Server erneut starten.
Für einen echten Wartezeittest zusätzlich `MVP_CONFIRM_DELAY_MS=65000` setzen:
Im Normalmodus wird dann vor dem Buchungsabschluss länger als eine Minute gewartet.
Die kontrollierte Sieben-Minuten-Prüfung im Store läuft bereits mit `go test ./...`.

### Grenzen des MVP

Speicherung und Bilder sind flüchtig; ein Neustart im Normalmodus beginnt leer,
ein Demo-Neustart stellt ausschließlich Beispieldaten wieder her. Ohne JavaScript
zeigt das Dashboard seine Anfangswerte; die Verwaltungsabläufe benötigen JavaScript.
Benutzerkonten, dauerhafte Speicherung, produktiver Betrieb, Zahlungen, Rechnungen,
Kundenportal, Zubehörbestände und externe Versicherungsabwicklung sind nicht umgesetzt.
Die Mobilprüfung verwendet Browseremulation; physische Geräte und Screenreader
sind nicht geprüft. Bekannte Grenzen und technische Nachweise stehen in Phase 5.

### Spezifikation und Abnahme

- [Mission und Produktverfassung](specs/mission.md)
- [Technische Verfassung](specs/tech-stack.md)
- [Roadmap](specs/roadmap.md)
- [Phase 0: Grundlagen und Prüfnachweis](specs/00-grundlagen.md)
- [Phase 1: Fahrradverwaltung, API und Prüfnachweis](specs/01-fahrradverwaltung.md)
- [Phase 2: Mietzeiten, Tarife und Prüfnachweis](specs/02-mietzeiten-preise.md)
- [Phase 3: Zusatzleistungen, API und Prüfnachweis](specs/03-zusatzleistungen.md)
- [Phase 4: Dashboard, Bedienung und Prüfnachweis](specs/04-dashboard-bedienung.md)
- [Phase 5: MVP-Anforderungsmatrix und Gesamtprüfung](specs/05-mvp-abnahme.md)
- [Änderung zur MVP-Abnahme: Sofortbeginn](specs/06-sofortbeginn.md)
- [Erweiterung: Buchen und ausgeben](specs/07-buchen-und-ausgeben.md)

Phase 5 bündelt die Anforderungen B-01 bis B-06 mit automatisierten Prüfungen
und Browsernachweisen. Die technische Prüfung ersetzt keine Auftraggeberabnahme.

Nach technischem Abschluss und ausdrücklicher Auftraggeberabnahme jeder Phase
wird ein lokaler Git-Commit erstellt. Ein Implementierungsauftrag allein gilt
nicht als Abnahme.
