# BikeRental – Phase 0: Grundlagen und Wiederverwendungsplan

**Spätere Ergänzung zur MVP-Prüfung:** [Sofortbeginn](06-sofortbeginn.md) ergänzt
Stundenmieten um volle Dauer ab Buchungsabschluss. Die hier dokumentierten festen
Termine behalten ihre Regeln; „Mietbeginn jetzt“ als fester Zeitstempel wurde ersetzt.


## Status und Ziel

**Status:** Technisch abgeschlossen, geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen.

Phase 0 schafft eine getrennte, startfähige Entwicklungsbasis und verbindliche
Geschäftsregeln für die folgenden Phasen. Die Beauftragung dieser Phase ist keine
Abnahme. Fahrradverwaltung, Stundenmieten, neue Tarife, Bilder und Extras werden
hier spezifiziert, aber noch nicht implementiert.

Grundlagen: [Mission](mission.md), [Technik](tech-stack.md),
[Roadmap](roadmap.md) und die Entscheidungen aus dem Projektgespräch.

## Herkunft, Arbeitsort und Versionierung

- Unveränderte Referenz: `SkiRental/`, Ausgangscommit
  `0050ea6bf47cbb74ee3ad420d246a8c022489870`
  (`Phase 5: Verleihtermine und Tarifverwaltung abschließen`).
- Arbeitskopie: `BikeRental/`, Go-Modul `bikerental`. Übernommen sind
  `cmd/`, `internal/`, `web/`, `testdata/` und `go.mod`; interne Go-Imports wurden
  angepasst. Es sind keine neuen Abhängigkeiten hinzugekommen.
- Das neue Git-Repository liegt im obersten Verzeichnis und umfasst BikeRental,
  README, Verfassung und Spezifikationen. Die eigenständige SkiRental-Referenz
  wird ausgeschlossen und behält ihre eigene Git-Historie. Sie muss bei einer
  späteren Weitergabe separat bereitgestellt werden; ihre Herkunft ist oben festgehalten.
- Alte SkiRental-Abnahmeberichte und Agentenanweisungen wurden nicht in die
  Arbeitskopie übernommen. Das oberste `AGENTS.md` bleibt unverändert. Seine
  Beschreibung des bisherigen Codeorts ist historisch; aktueller Arbeitsort
  und Befehle stehen in README und technischer Verfassung.
- Der Phase-0-Start verwendet noch die Ski-Fachfunktionen und deren Demodaten.
  Jede HTML-Seite und das Startprotokoll kennzeichnen diesen Ausgangsstand.

## Wiederverwendungsübersicht

| Bereich | Übernehmen | Gezielte Anpassung | Phase |
| --- | --- | --- | --- |
| Server und Assets | Loopback, Shutdown, `go:embed`, Schutzheader | Modulname und Entwicklungskennzeichnung | 0 |
| Personen | IDs, CRUD ohne Löschen, manuelle Nachweise und Fotoprüfung | Fahrradbezogene Verweise, Regression prüfen | 1–5 |
| Inventar | Verwaltung, IDs, atomare Änderungen | Einzelnes Fahrrad, Typ Downhill/MTB/Rennrad, Pflichtbild | 1 |
| Buchungen | Buchung, Ausgabe, Rückgabe, Stornierung, Konfliktprüfung | Gemeinsame Zeitintervalle für Stunden-/Tagesmiete | 2 |
| Preise | Ganze Cent, Tagesstaffel, Versionierung, Preisbestätigung | Stundenpreis und Tagesstaffel je Fahrradtyp | 2 |
| Extras | Validierung und atomare Buchungsoperation | Drei Optionen mit Pauschalen und Preisaufteilung | 3 |
| Dashboard | Vier Kennzahlen, konsistenter Snapshot, Aktualisierung | Fahrradzählung und Fahrradnavigation | 4 |
| Tests und Demo | Fach-/HTTP-/Race-Tests, synthetische Person | Fahrradfälle, Bilder, Typen und Beispielpreise | 1–5 |

## Verbindliche Geschäftsregeln für die Folgephasen

### Mietzeit und gemeinsame Verfügbarkeit

**Stundenmiete:** Beginn und Ende werden minutengenau in `Europe/Vienna`
eingegeben. Das Ende muss nach dem Beginn liegen. Abgerechnet wird mindestens
eine Stunde, danach jede angefangene Stunde. Buchungen über Mitternacht sind
zulässig. Die Abrechnung verwendet tatsächlich verstrichene Zeit; das reservierte
Intervall bleibt der gewählte Zeitraum und wird nicht auf ganze Stunden verlängert.

**Tagesmiete:** Beginn und Ende sind lokale Kalenderdaten; beide Randtage zählen.
Reserviert wird ab 00:00 Uhr des ersten Tags bis 00:00 Uhr nach dem letzten Tag.
Ein Kalendertag bleibt auch bei Zeitumstellungen ein abrechenbarer Tag.

Beide Mietarten werden auf Intervalle `[Beginn, Ende)` abgebildet. Zwei offene
Buchungen überschneiden sich genau dann, wenn `A.Beginn < B.Ende` und
`B.Beginn < A.Ende`. Direkte Anschlüsse sind ohne Puffer erlaubt. Ein Fahrrad
darf erst erneut tatsächlich ausgegeben werden, wenn die vorherige Ausgabe
zurückgenommen wurde; ein verstrichenes Planende genügt nicht.

Geschäftsdatum und Eingaben richten sich nach `Europe/Vienna`; Ereignisse und
normalisierte Intervallgrenzen werden als UTC-Zeitpunkte behandelt. Nicht
existierende oder doppeldeutige lokale Eingabezeiten bei Zeitumstellungen werden
mit verständlichem Fehler zurückgewiesen, statt stillschweigend verschoben zu werden.
Der bestehende Ausschluss vergangener Buchungsbeginne bleibt erhalten; eine
Tagesbuchung ab heute ist wie bisher möglich, eine Stundenbuchung beginnt nicht
vor der aktuellen Zeit. Die Feature-Spezifikation konkretisiert die minutengenaue Prüfung.

| Fall | Erwartetes Ergebnis |
| --- | --- |
| 02.11.2026, 10:00–10:15 | Eine Stunde berechnen |
| 02.11.2026, 10:00–11:15 | Zwei Stunden berechnen; Reservierung endet 11:15 |
| 02.11.2026, 23:30 bis 03.11.2026, 00:30 | Eine Stunde berechnen |
| Montag 02.11. bis Dienstag 03.11.2026 | Zwei Tage; reserviert bis 04.11., 00:00 |
| Stundenmiete bis 11:00, nächste ab 11:00 | Zeitlich erlaubt; Ausgabe erst nach tatsächlicher Rückgabe |
| Tagesmiete 02.11., Stundenmiete 02.11. um 15:00 | Konflikt beim selben Fahrrad |
| Tagesmiete 02.11., Stundenmiete 03.11. ab 00:00 | Zeitlich erlaubt |
| 28.03.2027, 02:30 in Wien | Nicht existierende Uhrzeit: ablehnen |
| 25.10.2026, 02:30 in Wien | Doppeldeutige Uhrzeit: ablehnen |
| 28.03.2027, 01:30–03:30 in Wien | Eine tatsächlich verstrichene Stunde berechnen |
| 25.10.2026, 01:30–03:30 in Wien | Drei tatsächlich verstrichene Stunden berechnen |
| Tagesmiete am Zeitumstellungstag | Ein Tag unabhängig von 23/25 Stunden |

### Tarife und Festpreis

Je Typ Downhill, MTB und Rennrad werden ein Stundenpreis, Gesamtpreise für 1 bis
N Kalendertage und ein Preis je Zusatztag bearbeitbar. Die Stundenmiete verwendet
Stundenzahl mal Stundenpreis; es gibt keinen automatischen Wechsel zur Tagesmiete.
Die bestehende Staffelberechnung wird für Tagesmieten weiterverwendet.

Der Server berechnet die Vorschau. Die Buchung bestätigt Preis und Tarifversion
atomar; zwischenzeitliche Änderungen verlangen eine neue Bestätigung. Grundpreis,
Extras und Gesamtpreis werden bei Buchung festgehalten. Spätere Ausgabe, frühere
oder verspätete Rückgabe erzeugen weder Erstattung noch Nachberechnung.
Verspätete Rückgaben werden weiterhin markiert.

Es sind keine echten Kundenbeträge vorgegeben. Die spätere Tarif-Spezifikation
legt klar gekennzeichnete synthetische Beispielpreise für die Demo fest. Im
Normalmodus müssen die erforderlichen Tarife vor einer Buchung eingerichtet sein.

### Zusatzleistungen und Bilder

Helm, Schutzausrüstung und Versicherung sind unabhängig kombinierbare Optionen,
jeweils mit bearbeitbarer EUR-Pauschale pro Buchung, unabhängig von Dauer und Typ.
Alle Optionen werden bei jeder neuen Buchung angeboten; eine Buchung ohne Extras
ist möglich. Es gibt keine Zubehörbestände. Versicherung dokumentiert ausschließlich
Auswahl und Preis, ohne externe Abwicklung oder Zusage eines Versicherungsschutzes.

Beim Anlegen eines Fahrrads ist ein JPEG-/PNG-Bild erforderlich, später ersetzbar.
Grenzen: 5 MiB und 20 Millionen Pixel; serverseitig Dateiinhalt und Dekodierbarkeit
prüfen. Stammdaten und Bild werden atomar gespeichert. Die Auswahl im Verleih
zeigt das tatsächliche Bild zusammen mit Name und Typ. Ein Platzhalter ersetzt
das Pflichtbild nicht. Bilder verbleiben wie andere Daten im Arbeitsspeicher.
Die Demo erhält mitgelieferte zulässige Fahrradbilder in den zuständigen Folgephasen.

## Umsetzungsschritte und Abnahmekriterien

1. Herkunft und unveränderten Referenzstand festhalten.
2. Geschäftsregeln in Verfassung und dieser Spezifikation dokumentieren.
3. Separate Codebasis anlegen, Modul/Imports anpassen und Ausgangsstand kennzeichnen.
4. Oberstes Git-Repository mit passenden Ausschlüssen anlegen; noch kein Phasencommit.
5. Build, Tests, Race und Vet für Referenz und Arbeitskopie protokollieren.
6. Beide Startmodi einschließlich eingebetteter Oberfläche, Assets und Kennzahlen prüfen.
7. README, Verweise und Konsistenz prüfen; technische Fertigstellung dokumentieren.
8. Nach ausdrücklicher Auftraggeberabnahme Status aktualisieren und lokal committen.

Phase 0 ist technisch erfüllt, wenn diese Schritte bis einschließlich 7 erfolgreich
sind. Neue Fahrrad-APIs und die Regeln dieses Dokuments sind bis zu den jeweiligen
Folgephasen ausdrücklich noch keine implementierten Fähigkeiten.

## Prüfnachweis

Prüfung am **21.10.2026**, Go **1.27.3**, `darwin/arm64`.
Verwendete Toolchain: `/private/tmp/skirental-toolchain/go/bin/go`;
`GOCACHE=/private/tmp/skirental-go-cache`. Keine zusätzlichen Abhängigkeiten.
Die folgenden Befehle liefen jeweils im betreffenden Go-Modulverzeichnis:

| Prüfung | SkiRental-Referenz | BikeRental-Arbeitskopie |
| --- | --- | --- |
| `go build ./...` | Bestanden | Bestanden |
| `go test -count=1 ./...` | Domain, Handler und Store bestanden | Domain, Handler und Store bestanden |
| `go test -race -count=1 ./...` | Bestanden, keine Race-Meldung | Bestanden, keine Race-Meldung |
| `go vet ./...` | Bestanden | Bestanden |
| Eigenständiger Binary-Start, Normalmodus | Bestanden | Bestanden |
| Eigenständiger Binary-Start, `-demo` | Bestanden | Bestanden |
| Normalmodus nach Testanlage und Neustart leer | Bestanden | Bestanden |

**Startprüfung:** Für beide Module wurde `go build -o <temporäres Binary>
./cmd/server` ausgeführt. Beide Modi wurden nacheinander auf `127.0.0.1:8080`
gestartet, jeweils aus einem temporären Verzeichnis außerhalb des Repositorys.
Geprüft wurden HTTP 200 und erwarteter Inhalt für `/`, `/ski`, `/personen`,
`/tarife`, `/ausleihe` sowie alle fünf CSS-/JavaScript-Assets. Der Phase-0-Hinweis
war auf allen BikeRental-Seiten vorhanden. Demokennzeichnung, Speicherhinweis,
`no-store` und `nosniff` waren korrekt.

Das Dashboard lieferte im Normalmodus vier Nullwerte, im Demomodus zwei verfügbare
Ski-Paare, null ausgegebene, zwei insgesamt und eine registrierte Person. Ein
synthetischer Inventareintrag wurde im Normalmodus per API angelegt; nach einem
Neustart desselben Modus war der Bestand wieder leer. Alle gestarteten
Serverprozesse wurden anschließend mit erfolgreichem Exit beendet.

Die Sandbox verweigerte zunächst das Binden des lokalen Ports. Nach erteilter
Ausführungsfreigabe bestand dieselbe Startprüfung vollständig. Das temporäre
Prüfskript liegt unter `/private/tmp/bikerental-phase0-smoke.py`, Build-/Test- und
Serverlogs unter `/private/tmp/bikerental-phase0-checks/`; diese Hilfsdateien sind
keine dauerhaften Projektvoraussetzungen. Die relevanten Ergebnisse sind hier
festgehalten. Ein visueller Browser-/Gerätetest ist für Phase 0 nicht erfolgt;
die fachlichen UI-Prüfungen sind Bestandteil der Folgephasen.

**Weitere Kontrollen:** Lokale Markdown-Verweise sind auflösbar. Die Beispiele
für Zeitumstellungen wurden gegen `Europe/Vienna` geprüft; das belegt ihre
fachliche Erwartung, noch keine implementierte Stundenmietfunktion. Go-Imports
verweisen auf `bikerental`; keine verschachtelte `.git`-Struktur oder alten
Agentenanweisungen wurden kopiert. Git-Ausschlüsse für Referenz und Build-Dateien
wurden geprüft. `git -C SkiRental status --porcelain` blieb leer und der
Referenzcommit unverändert. Das oberste `AGENTS.md` wurde nicht verändert.

**Ergebnis:** Alle Phase-0-Kriterien bis einschließlich technischer Prüfung sind
erfüllt. Die unveränderten Ski-Fachfunktionen, alten API-Feldnamen und Ski-Demodaten
sind der bewusst übernommene Ausgangsstand. B-01 bis B-06 werden erst in den
Folgephasen umgesetzt; dafür wird hier keine Abnahme beansprucht.

## Auftraggeberabnahme

**Am 21.10.2026 ausdrücklich erteilt.** Bestätigung im Projektgespräch:
„phase 0 ist hiermit abgenommen“.

Der abgenommene Stand wird mit dem lokalen Commit
`Phase 0: Grundlagen und Wiederverwendungsbasis abschließen` gesichert.
