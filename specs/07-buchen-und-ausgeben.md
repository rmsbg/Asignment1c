# Phase 6 – Buchen und ausgeben

Status: implementiert und technisch geprüft am 22.10.2026.
Am 22.10.2026 ausdrücklich vom Auftraggeber abgenommen.

## Verhalten

Übernahme der Erweiterung aus SkiRental (`Assignment1ab/specs/06-buchen-und-ausgeben.md`)
mit Anpassung an die beiden BikeRental-Mietarten:

- Tagesmiete ab heute in Europe/Vienna: „Buchen und ausgeben“ und „Nur buchen“.
- Stundenmiete mit „Sofort beginnen“: ebenfalls beide Aktionen. Die volle gewählte
  Dauer beginnt beim erfolgreichen Buchungsabschluss. Bei Sofortausgabe stimmen
  Mietbeginn, Buchungszeit und Ausgabezeit überein.
- Tagesmiete mit späterem Beginn sowie Stundenmiete mit „Für später reservieren“:
  „Buchen“. Die tatsächliche Ausgabe wird später separat erfasst.

Alle Buchungsaktionen bestätigen den zuvor angezeigten Preis einschließlich
gewählter Extras. Beide Buttons bleiben während der Anfrage gesperrt.
Änderungen an Zeiten, Typ oder Extras verwerfen Auswahl und Vorschau.
Nach Sofortausgabe führt „Zur Buchungsübersicht“ direkt zur ausgegebenen Buchung;
nach reiner Buchung bleibt „Zur Ausgabe und Buchungsübersicht“ erhalten.

## Schnittstelle und Speicherung

`POST /api/ausleihen` akzeptiert das optionale Boolean `issueImmediately`
(Standard `false`). Bei `true` prüft und speichert `CreateAndIssueRental`
Buchung und Ausgabe unter derselben Store-Sperre. Erfolg: HTTP 201 mit Location,
Status `issued` und `rentedAt`. Ohne Flag oder bei `false` bleibt es bei einer
reinen Buchung über `CreateRental`.

Preisversion, Typ, Überschneidungen, Ausgabezeitraum und tatsächlich ausgegebene
Vorgänger werden vor jeder Zustandsänderung geprüft. Konflikte liefern HTTP 409,
ungültige Eingaben 400. Es bleibt keine Teilbuchung zurück, auch keine verbrauchte
Buchungsnummer. Nach Konflikten lädt das Formular eine neue Preis-/Verfügbarkeits-
vorschau und verlangt erneute Auswahl und Bestätigung.

Die separate Ausgabe verwendet dieselben Ausgaberegeln. Preis, gebuchte Extras
und Dauer bleiben nach der Ausgabe unverändert. Keine neuen Abhängigkeiten.

## Prüfnachweis

Erfolgreich:

- `go build ./...`
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`

Store-Tests: beide Mietarten mit Extras und Dashboardzahlen; volle Stundendauer;
Wiener Tageswechsel; zukünftige Tages-/Stundenmieten, geänderte Tarife oder Typen,
Überschneidungen und überfällige Vorgänger ohne Teiländerungen;
32 konkurrierende Sofortausgaben mit genau einem Erfolg.
HTTP-Tests: neues Flag, fehlend/false abwärtskompatibel, Status, Ausgabezeitpunkt,
Location, ungültige Flag-Typen, Doppelbuchung sowie zukünftige Sofortausgabe mit
409 ohne Teilbuchung und anschließend erfolgreicher reiner Reservierung.
Vorhandene Tests laufen weiterhin erfolgreich.

Browserprüfung mit Chrome: Desktop 1280 × 1000 und mobil 390 × 844; Tages-/
Stundenmiete jeweils mit Sofortausgabe und reiner Buchung; Tastaturbedienung;
beide Buttons während einer angehaltenen Anfrage gesperrt; Rückgabe und Dashboard;
Umschalten bei Datums-/Startmodusänderung; Tarifkonflikt ohne gespeicherte Buchung
und mit erneuter Auswahl. Keine unbehandelten JavaScript-Fehler, kein horizontaler
Überlauf. Mobilprüfung per Emulation.

Reproduzierbar gegen einen frisch gestarteten Demoserver aus `BikeRental/` mit
`node tools/book-issue-browser.cjs`; Playwright und Chromium wie in der README
konfigurieren (`PLAYWRIGHT_MODULE`, `CHROME_EXECUTABLE_PATH`). Der Prüflauf verändert
synthetische Buchungen und eine Demopauschale; danach den Demoserver neu starten.
Der vorhandene MVP-Browserprüflauf verwendet jetzt „Nur buchen“ bzw. „Buchen“
anstelle der früheren Beschriftung. Er wurde in diesem Schritt nicht erneut ausgeführt.

Screenshots:

- [Desktop](screenshots/07-buchen-ausgeben-desktop.png)
- [Mobil](screenshots/07-buchen-ausgeben-mobil.png)

Nach Abschluss der Prüfung wurde der Demoserver mit frischen synthetischen Daten
gestartet und seine Erreichbarkeit geprüft. Alle Daten bleiben flüchtig.


## Auftraggeberabnahme am 22.10.2026

- [x] Tagesmiete ab heute bietet „Buchen und ausgeben“ und „Nur buchen“.
- [x] Stundenmiete mit „Sofort beginnen“ bietet ebenfalls beide Aktionen.
- [x] Spätere Termine behalten die reine Buchung und separate Ausgabe.
- [x] Buchung und Sofortausgabe erfolgen gemeinsam, ohne Teilbuchung bei Konflikten.

Der Auftraggeber hat die Änderung ausdrücklich mit „änderung ist abgenommen.
bitte mach einen commit und aktualisiere die dokumente entsprechend“ abgenommen.
Damit ist Phase 6 abgeschlossen. Die Abnahme beruht auf dieser ausdrücklichen
Bestätigung und wird nicht allein aus technischen Prüfungen abgeleitet.

Abnahme und Prüfnachweis sind vor dem lokalen Abschluss-Commit
`Phase 6: Buchung mit sofortiger Ausgabe abschließen` dokumentiert.
