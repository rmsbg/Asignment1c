# BikeRental – Phase 2: Mietzeiten und typabhängige Preise

## Status und Umfang

**Status:** Implementiert und technisch geprüft am 21.10.2026.
Am 21.10.2026 ausdrücklich vom Auftraggeber abgenommen.
Phase 1 wurde ausdrücklich abgenommen und mit `b8901d2` gesichert.

Grundlage sind [Phase 0](00-grundlagen.md), [Mission](mission.md) und
[Roadmap](roadmap.md). B-02 und B-03 werden umgesetzt; Fahrradbilder,
Personenverwaltung, Ausgabe/Rückgabe, Stornierung und Dashboard bleiben erhalten.
Zusatzleistungen folgen in Phase 3. Speicherung bleibt flüchtig.

## Geschäftsregeln und Nutzerablauf

Mitarbeiter wählen Person, Fahrradtyp, Mietart und Beginn/Ende. Tagesmiete zählt
beide Randtage und reserviert ganze Kalendertage. Stundenmiete akzeptiert
minutengenaue lokale Uhrzeiten, mindestens eine berechnete Stunde, mit Aufrundung
angefangener Stunden; Buchungen über Mitternacht sind zulässig. Der gewählte
Zeitraum wird durch die Preisrundung nicht verlängert. Keine automatische Wahl
eines günstigeren Tarifs.

Alle Eingaben richten sich nach Europe/Vienna, unabhängig von der Browserzeitzone.
Stunden beginnen nicht vor der aktuellen Wiener Minute (Sekunden werden für
diese Eingabeprüfung abgeschnitten), Tagesmiete darf ab heute beginnen. Ende muss
nach Beginn liegen; für Tagesmieten darf dasselbe Datum gewählt werden.
Nicht existierende/doppeldeutige Uhrzeiten werden abgelehnt. Tatsächlich
verstrichene Stunden und Kalendertage werden getrennt berechnet.

Stunden- und Tagesmieten verwenden UTC-Intervalle `[Beginn, Ende)`. Für Tagesmiete
ist Ende die Wiener Mitternacht nach dem letzten Miettag. Direkter Anschluss ohne
Puffer ist erlaubt; Überschneidungen desselben Fahrrads sind atomar ausgeschlossen.
Tatsächliche Ausgabe nur im Intervall und ohne noch offene vorherige Ausgabe.
Am exklusiven Endzeitpunkt wird eine nicht ausgegebene Buchung abgelaufen und
eine noch ausgegebene Ausleihe überfällig. Kein automatischer Zustandswechsel
zur Rückgabe. Wiederholte Ausgabe/Rückgabe und Stornierung bleiben idempotent.

Je Typ werden ein Stundenpreis, Gesamtpreise für 1 bis N Tage und ein Preis je
Zusatztag verwaltet. Ein gemeinsamer Versionszähler schützt die atomare Änderung
aller drei Tarife. Nur vollständig eingerichtete Tarife sind buchbar. Die
Preisvorschau zeigt Mietart, Typ, berechnete Dauer, Preis und Tarifversion.
Bei Änderungen von Zeit, Typ oder Mietart werden Auswahl und Vorschau verworfen.
Nach Tarif-/Bestandskonflikten muss der aktualisierte Preis erneut gewählt und
bestätigt werden. Typänderungen am Inventar zwischen Vorschau und Buchung erzeugen
ebenfalls einen Konflikt. Bestehende Buchungen bewahren den gebuchten Typ,
Mietart, Zeitraum, Dauer und Gesamtpreis unabhängig von späteren Änderungen.

Demo-Beispielpreise, ausdrücklich keine vorgegebenen Kundentarife:

| Typ | Stunde | 1 / 2 / 3 Tage | Zusatztag |
| --- | --- | --- | --- |
| Downhill | 12 EUR | 40 / 70 / 95 EUR | 25 EUR |
| MTB | 8 EUR | 25 / 45 / 60 EUR | 15 EUR |
| Rennrad | 10 EUR | 35 / 60 / 80 EUR | 20 EUR |

Ganze Cent, nichtnegative Werte, Grenze `9007199254740991` einschließlich
Überlaufprüfung. Keine automatische Nachberechnung/Erstattung bei abweichender
Ausgabe oder Rückgabe. Normalstart ohne Tarif; Demo mit obigen Preisen.

## Schnittstellen und Daten

Dieser Abschnitt beschreibt den abgenommenen Phase-2-Stand (`be2d6de`).
Die explizite Extraauswahl und Preisaufteilung werden in
[Phase 3](03-zusatzleistungen.md) ergänzt.

- `GET/PUT /api/tarife`: `{version, types:{Downhill:{hourlyCents, pricesCents,
  extraDayCents}, MTB:{...}, Rennrad:{...}}}`. PUT verlangt alle drei vollständigen
  Typen und den aktuellen Versionszähler; fehlende/null/ungültige Felder werden
  abgelehnt. Konflikt: 409. Alte gemeinsame Tarif-Schreibform entfällt.
- `POST /api/ausleihen/preis`: `{bikeType, mode, startsOn, endsOn}`; `mode` ist
  `day` oder `hour`. Datumsformat `YYYY-MM-DD`, Stundenformat `YYYY-MM-DDTHH:mm`
  ohne Offset. Antwort enthält Typ/Mietart, `durationDays` oder `durationHours`,
  UTC-`startsAt`/`endsAt`, `totalPriceCents`, `tariffVersion`.
- `GET /api/fahrraeder/verfuegbar`: mit Zeitraum sind `mode`, `startsOn`, `endsOn`
  erforderlich; optionaler `bikeType` filtert die Karten. Ohne Zeitraum bleibt
  die physische Verfügbarkeit abrufbar. Typ ohne Zeitraum ist ungültig.
- `POST /api/ausleihen`: `{personID, bikeID, bikeType, mode, startsOn, endsOn,
  tariffVersion}`. Referenzen, tatsächlicher Typ, Zeit, Tarifversion, Preis und
  Verfügbarkeit werden zusammen unter der Store-Sperre geprüft. Keine vom Client
  vorgegebenen Preise. Anlage 201; ungültig 400, unbekannte Referenzen 404,
  Konflikte 409, zu große Requests 413, falscher Medientyp 415.
- Buchungsantworten ergänzen `bikeType`, `mode`, `startsAt`, `endsAt`,
  `durationHours`, `canIssue`; Tagesdauer und vorhandene Status-/Ereignisfelder
  bleiben. Zeitanzeige erfolgt ausdrücklich in Europe/Vienna. Ausgabeaktionen
  bleiben serverseitig autoritativ; `canIssue` ist nur eine Anzeigehilfe.

Keine Datenmigration: Es gibt weder dauerhafte Daten noch externe Clients.
Ein Neustart verwirft den bisherigen Speicher. Vorhandene Tests werden auf die
neue explizite Schnittstelle angepasst und um Stunden-/Mischfälle erweitert.

## Aufgaben und Akzeptanzprüfung

1. Zeitintervalle, Zeitzonenvalidierung und Preisregeln im Domain-Paket implementieren.
2. Atomare typabhängige Tarife, gemischte Konfliktprüfung und Preissnapshots im Store.
3. HTTP-Verträge und Fehlerfälle; vorhandene Personen-/Bildfunktionen erhalten.
4. Tarifverwaltung für alle Typen und Mietart-/Typauswahl im Buchungsablauf.
5. Prüfungen: beide Mietarten × drei Typen, 15/60/75 Minuten, Mitternacht,
   Schaltjahr und Zeitumstellungen, Tagesgrenzen und erlaubte Anschlüsse,
   gemischte parallele Doppelbuchungen, veraltete Tarife/Inventartypen, früher/später
   ausgegeben/zurückgegeben, Überfälligkeit und idempotente Aktionen.
6. Build, Tests, Race und Vet sowie Browserprüfung Normal-/Demomodus auf Desktop
   und mobil; Tastaturwahl, Fehlerfokus, Neuprüfung nach Änderung, Preisbestätigung.
7. README, Verfassung, Roadmap und Prüfnachweis aktualisieren. Erst nach ausdrücklicher
   Auftraggeberabnahme lokaler Commit `Phase 2: Mietzeiten und typabhängige Preise abschließen`.

## Prüfnachweis und Auftraggeberabnahme

Technische Prüfung am **21.10.2026**, Go **1.27.3**, `darwin/arm64`.
Toolchain `/private/tmp/skirental-toolchain/go/bin/go`,
`GOCACHE=/private/tmp/skirental-go-cache`. Keine neuen Projektabhängigkeiten.

| Prüfung | Ergebnis |
| --- | --- |
| `go build ./...` | Bestanden |
| `go test -count=1 ./...` | Domain, Handler und Store bestanden |
| `go test -race -count=1 ./...` | Bestanden; keine Race-Meldung |
| `go vet ./...` | Bestanden |
| `node --check` für Ausleihe und Tarifverwaltung | Bestanden |
| Browserprüfung Normal-/Demomodus | Beide bestanden |
| Desktop 1440 × 1000 / mobil 390 × 844 | Bilder und Formulare sichtbar, kein horizontaler Überlauf |

**Domain:** Erwartungsfälle für 15/60/75 Minuten, Mitternacht, Schaltjahr,
23-/25-Stunden-Tage, beide Zeitumstellungen, mehrdeutige/nicht existierende
Eingabezeiten, Sekunden-/Offseteingaben, Beginn vor der aktuellen Minute,
exklusive Endgrenzen, Ganzcentgrenzen und Preisüberlauf. Alle drei vollständigen
Typen sind erforderlich; unbekannte und unvollständige Tarife werden abgelehnt.

**Store/HTTP:** Tagesmiete und Stundenmiete für alle drei Typen, typgefilterte
Bildauswahl, UTC-Intervalle, fixe Preissnapshots und Versionen, atomare Ablehnung
von Überschneidungen und gemischten parallelen Buchungen. Direkter Stundenanschluss
und Tages-/Stundenübergang um Mitternacht sind zulässig. Ausgabe vor Beginn oder
nach Ende sowie bei verspätetem Vorgänger scheitert; Rückgabe und Stornierung
bleiben idempotent. Vergangene Buchungen laufen an der exklusiven Endgrenze ab.

Bestehende Regressionstests für Personen, Bilder, Dashboard und vollständige
Tagesabläufe wurden erhalten und auf die neue Schnittstelle angepasst. Der
Gesamtfluss prüft beide Startmodi jeweils mit manuellem Nachweis, PNG und JPEG.
Tarif- und Inventartypkonflikte erzeugen keine Teilbuchungen. Eine tatsächliche
Inventartypänderung nach Vorschau erhält ältere Buchungen und erzwingt eine neue
Bestätigung; dieser ergänzte Fall und unabhängige Tarif-Maps wurden zusätzlich
gezielt unter Race geprüft. Aktualisierungen verwenden unabhängige Preisstaffeln.
HTTP-Tests prüfen unter anderem fehlende/null/gebrochene oder negative Preise,
Teilkonfigurationen, unbekannte Typen und veraltete Versionen.

**Browser:** Vorhandenes Playwright mit separatem lokalem Chrome, ohne Installation
neuer Pakete. Der jeweilige Build startete aus einem temporären Verzeichnis
außerhalb des Repositorys in Normal- und Demomodus. Im Normalmodus wurden Bilder
und Person mit synthetischen Daten eingerichtet und alle drei Tarife über das
Formular gespeichert. EUR-Eingaben mit zu vielen Nachkommastellen wurden mit
Fehlerfokus und erhaltenen Eingaben abgelehnt.

Alle sechs Kombinationen aus Typ/Mietart zeigten die erwarteten Preise und passende
Fahrradbilder. Typ-/Mietart-/Zeitänderungen verwarfen Auswahl und Preisvorschau.
Nach externer Tarifänderung wurde die alte Buchung abgelehnt, der neue Preis
angezeigt und eine erneute Auswahl/Bestätigung verlangt. Stundenbuchung ab der
aktuellen Minute, Ausgabe, Rückgabe, Stornierung, gemischte Konflikte und erlaubter
Mitternachtsanschluss wurden geprüft. Zwei gleichzeitig geöffnete Tarifeditoren
verhinderten das Überschreiben einer neueren Änderung; Eingaben blieben erhalten.
Bereits bestätigte Preise blieben trotz späterer Tarifänderung unverändert.

Zusätzlich lief der Browser ausdrücklich in **America/New_York**; Eingaben und
Anzeigen blieben auf **Europe/Vienna** bezogen. Desktop-/Mobilansichten und
Tastaturauswahl wurden geprüft; keine unbehandelten JavaScript-Fehler. Die
Prüfserver und separaten Browserprozesse wurden nach dem Durchlauf beendet.
Der lokale Demostand wird anschließend für die Auftraggeberprüfung frisch gestartet.

Screenshots mit synthetischen Daten:

- [Tarifverwaltung Desktop](screenshots/02-tarife-desktop.png)
- [Tarifverwaltung mobil](screenshots/02-tarife-mobil.png)
- [Stundenmiete Desktop](screenshots/02-stunden-desktop.png)
- [Stundenmiete mobil](screenshots/02-stunden-mobil.png)
- [Erneute Preisbestätigung nach Tarifkonflikt](screenshots/02-tarifkonflikt.png)

Temporäre Prüfhilfen: `/private/tmp/bikerental-phase2-browser.cjs` und
`/private/tmp/bikerental-phase2-checks/`. Dauerhafte Nachweise sind dieses Dokument,
die automatisierten Go-Tests und die Screenshots. Kein physischer Mobilgerätetest
oder eigener Screenreader-Test. SkiRental und das oberste `AGENTS.md` bleiben
unverändert. Extras sind weiterhin nicht Teil dieser Phase.

**Auftraggeberabnahme:** Am 21.10.2026 ausdrücklich erteilt:
„phase 2 abgenommen. implementiere phase 3“. Der abgenommene Stand wird mit dem
folgenden lokalen Commit gesichert: `Phase 2: Mietzeiten und typabhängige Preise abschließen`.
