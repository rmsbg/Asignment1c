# BikeRental – Phase 3: Zusatzleistungen im Buchungsablauf

## Status und Umfang

**Status:** Implementiert und technisch geprüft am 21.10.2026.
Am 21.10.2026 ausdrücklich vom Auftraggeber abgenommen.
Phase 2 ist ausdrücklich abgenommen und mit `be2d6de` gesichert.
Grundlagen: [Mission](mission.md), [Phase 0](00-grundlagen.md),
[Phase 2](02-mietzeiten-preise.md) und [Roadmap](roadmap.md).

B-04 ergänzt jede neue Buchung um Helm, Schutzausrüstung und Versicherung.
Die bestehenden Fahrrad-, Personen-, Zeit-, Preis- und Verfügbarkeitsregeln
bleiben erhalten. Keine eigenen Zubehörbestände, Zahlungen, externe Anbieter
oder dauerhafte Speicherung.

## Nutzerablauf und Geschäftsregeln

Im Buchungsformular erscheinen alle drei Optionen als beschriftete Checkboxen,
anfangs nicht ausgewählt. Jede Kombination einschließlich keiner Auswahl ist
zulässig. Jede Option wird höchstens einmal und ausschließlich pro Buchung
berechnet, unabhängig von Mietdauer oder Fahrradtyp. Die Versicherungsoption
dokumentiert Auswahl und Aufpreis; der lokale Prototyp führt keine externe
Versicherungsabwicklung durch.

„Verfügbarkeit und Preis prüfen“ zeigt Fahrradpreis, gewählte Extras mit ihren
Einzelpauschalen, Extrasumme und Gesamtpreis. Ohne Extras steht ausdrücklich
„Keine Zusatzleistungen“. Änderung einer Option verwirft Preisvorschau und
Fahrradauswahl; vor Abschluss ist erneut zu prüfen. Nach erfolgreicher Buchung
startet die nächste Eingabe ohne vorausgewählte Extras. Abbrechen speichert nichts.

In Buchungsübersicht und Verlauf bleiben Auswahl, Einzelpauschalen, Fahrradpreis,
Extrasumme und Gesamtpreis sichtbar, auch nach Ausgabe, Rückgabe oder Stornierung.
Tarifänderungen wirken nur auf neue Buchungen. Auch spätere Ausgabe oder
frühere/spätere Rückgabe verändern die bestätigte Preisaufteilung nicht.

Alle drei Pauschalen werden zusammen mit den Fahrradtarifen unter „Tarife
verwalten“ bearbeitet. Gemeinsamer Versionszähler: Auch die Änderung ausschließlich
einer Extrapauschale erfordert bei älteren Vorschauen eine erneute Preisbestätigung.
Demo-Beispielpauschalen: Helm **5 EUR**, Schutzausrüstung **10 EUR**, Versicherung
**8 EUR**. Keine vom Kunden vorgegebenen Echtpreise. Im Normalmodus müssen alle
Tarife und drei Pauschalen ausdrücklich eingerichtet werden; JSON-`null` ist kein Preis,
eine explizite Pauschale von 0 Cent ist zulässig.

## Daten und Schnittstellen

- `Tariff.extras`: `{helmetCents, protectionCents, insuranceCents}` als ganze,
  nichtnegative Cent bis `9007199254740991`. GET liefert auch im leeren Zustand
  alle drei Werte (0), Version 0 kennzeichnet „nicht eingerichtet“. PUT verlangt
  zusätzlich zu den bisherigen Feldern alle drei vollständigen Pauschalen.
- Preisvorschau und Buchungsanfrage verlangen `extras` mit drei expliziten
  booleschen Feldern `{helmet, protection, insurance}`. Fehlende/null/unbekannte
  Felder und falsche Typen werden abgelehnt. Drei `false` bedeuten ohne Extras.
- Quote und Buchungsantwort ergänzen `basePriceCents`, `extras`, `extraPrices`
  (tatsächlich berechnete Einzelpauschalen; nicht gewählte Optionen 0),
  `extrasPriceCents` und den bisherigen `totalPriceCents`.
- Nur die Quote ergänzt `extraRates` mit den drei aktuell angebotenen Pauschalen,
  damit Beschriftungen nach einer Neuprüfung auch bei Tarifänderungen aktuell sind.
- Server berechnet sämtliche Beträge. Gesamtpreis ist Fahrradpreis plus einmalige
  Summe der gewählten Extras. Jede Addition wird gegen die Ganzzahlgrenze geprüft;
  Überlauf erzeugt einen Feldfehler zu `extras` und keinerlei Buchung.
- Referenzen, Auswahl, aktueller Tarifstand, Preis und Fahrradverfügbarkeit werden
  unter derselben Store-Sperre wie die Buchungsanlage verarbeitet. Snapshot-Daten
  sind Werte ohne veränderbare Auswahl-/Preispuffer. Kein separates Speichern
  von Extras und damit keine unvollständigen Teilbuchungen.
- Fehlerformat und Status bleiben erhalten: ungültig 400, unbekannt 404,
  Konflikt 409, zu groß 413, falscher Medientyp 415. Vom Client übermittelte Preise
  werden nicht akzeptiert. Keine Migration, da keine persistenten Daten existieren.

## Aufgaben und Akzeptanzprüfung

1. Auswahl, Pauschalen und geprüfte Preisaddition im Domain-Modell ergänzen.
2. Atomare Tarifänderungen, Vorschau und unveränderliche Buchungssnapshots erweitern.
3. HTTP-Verträge mit expliziten Auswahlfeldern und serverseitiger Validierung.
4. Optionen in jedem Buchungsformular, erneute Prüfung nach Änderungen,
   Preisaufschlüsselung in Vorschau und Verlauf, Pauschalen in Tarifverwaltung.
5. Alle acht Kombinationen für Stunden-/Tagesmiete und drei Fahrradtypen prüfen;
   Dauer darf Extras nicht vervielfachen. Auch 0-Cent-Optionen bleiben als gewählt sichtbar.
6. Fehlerhafte Auswahl/Preise, Preisüberlauf, veraltete Vorschau nach Pauschaländerung,
   gleichzeitige Buchungen/Tarifänderungen und dauerhafte Snapshots innerhalb der
   Sitzung prüfen. Stornierung, Ausgabe und Rückgabe bleiben korrekt und idempotent.
7. Build, Tests, Race, Vet und Browserablauf in Normal-/Demomodus, Desktop/mobil,
   Tastaturauswahl, Fehlermeldungen und Konfliktbestätigung prüfen. Personen- und
   Bildfunktionen sowie gemischte Verfügbarkeit müssen weiter funktionieren.
8. README, Verfassung, Roadmap, Screenshots und Prüfnachweis aktualisieren. Nach
   ausdrücklicher Abnahme Commit `Phase 3: Zusatzleistungen im Buchungsablauf abschließen`.

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
| `node --check` für Ausleihe und Tarife | Bestanden |
| Browserablauf Normal-/Demomodus | Beide bestanden |
| Desktop 1440 × 1000 / mobil 390 × 844 | Bilder/Formulare sichtbar, kein horizontaler Überlauf |

**Domain und Store:** Alle acht Auswahlen, beide Mietarten und alle drei Typen
wurden für Preisvorschau und Buchung geprüft (48 Kombinationen je Store-/HTTP-Suite).
Eine mehrtägige oder mehrstündige Miete vervielfacht die Pauschalen nicht.
Grenzfälle prüfen ungültige Gebühren, genaue Ganzzahlgrenze, Überlauf bei
Grundpreis plus Extrakosten und beim Summieren mehrerer Optionen sowie ausgewählte
0-Cent-Optionen. Ungültige Vorgänge erzeugen keine Buchung. Auswahl- und Gebührenwerte
in Snapshots sind unabhängig; spätere Tarifänderungen ändern bestehende Buchungen
auch bei verspäteter Rückgabe nicht.

Die Race-Prüfung deckt gleichzeitige Extragebührenänderungen und Buchungen ab:
Tarifversion, Einzelgebühren und Gesamtsumme stammen immer aus demselben Stand.
Bereits vorhandene Tests für gemischte Doppelbuchungen, Bildpuffer, Personen,
Zeitzonen, Ausgabe/Rückgabe, Stornierung und Dashboard bestehen weiterhin.

**HTTP:** Fehlende/null/unvollständige oder falsch typisierte Extraauswahl,
unbekannte Optionen, unvollständige, negative, gebrochene, null oder übergroße
Gebühren sowie vom Client vorgegebene Preise werden abgelehnt. Alle acht
Auswahlen bleiben nach Stornierung in Details abrufbar. Stale Tarifversionen,
auch nach reiner Extragebührenänderung, werden ohne Teilbuchung abgelehnt.
Eine ausgewählte kostenlose Option bleibt ausdrücklich als gewählt erhalten.

**Browser:** Vorhandenes Playwright mit separatem lokalen Chrome, ohne neue
Paketinstallation. Beide Builds starteten aus einem temporären Verzeichnis
außerhalb des Repositorys. Im Normalmodus wurden synthetische Person/Fahrräder
und sämtliche Tarife einschließlich der drei Extrapauschalen eingerichtet.
Ungültige EUR-Pauschalen mit zu vielen Nachkommastellen führten zu Fehlerfokus
und ließen Werte sowie Store unverändert.

Alle acht Kombinationen wurden mit Stunden- und mehrtägiger Tagesmiete geprüft.
Auswahländerungen per Tastatur verwarfen Vorschau und Fahrradwahl; Abbrechen
speicherte nichts. Ohne Extras wurde dies ausdrücklich angezeigt. Eine reine
Helmpreisänderung zwischen Vorschau und Abschluss erzwang neue Prüfung und
Bestätigung; Optionen und Preisaufschlüsselung verwendeten die neuen Gebühren.
Nach Buchung waren die Optionen für die nächste Eingabe abgewählt. Ausgabe,
Rückgabe und Verlauf bewahrten alle Einzelpreise. Ein kostenlos gewählter Helm
blieb auch nach Stornierung sichtbar und wurde nicht als „keine Extras“ behandelt.
Zwei gleichzeitig geöffnete Gebühreneditoren verhinderten das Überschreiben
einer neueren Änderung und erhielten bei Konflikten die Formulareingaben.

Desktop-/Mobilansichten und Preisaufteilung wurden visuell kontrolliert; im
Browser war ausdrücklich `America/New_York` eingestellt, während die Fachzeiten
weiterhin `Europe/Vienna` verwendeten. Keine unbehandelten JavaScript-Fehler.
Die Prüfserver und separaten Browserprozesse wurden nach dem Durchlauf beendet;
anschließend wird ein frischer lokaler Demostand für die Auftraggeberprüfung gestartet.

Screenshots verwenden ausschließlich synthetische Daten und zeigen teilweise
im Test bearbeitete Pauschalen einschließlich eines kostenlosen Helms:

- [Vorschau mit allen Extras](screenshots/03-extras-vorschau-desktop.png)
- [Extras mobil und Verlauf](screenshots/03-extras-mobil.png)
- [Verlauf mit unveränderlichen Einzelpreisen](screenshots/03-extras-verlauf-desktop.png)
- [Extratarife Desktop](screenshots/03-extratarife-desktop.png)
- [Extratarife mobil](screenshots/03-extratarife-mobil.png)
- [Erneute Bestätigung nach Gebührenänderung](screenshots/03-extrakonflikt.png)

Temporäre Prüfhilfen: `/private/tmp/bikerental-phase3-browser.cjs` und
`/private/tmp/bikerental-phase3-checks/`. Dauerhafte Nachweise sind dieses Dokument,
die Go-Tests und Screenshots. Kein eigener Screenreader-Test oder physischer
Mobilgerätetest. SkiRental und das oberste `AGENTS.md` bleiben unverändert.

**Auftraggeberabnahme:** Am 21.10.2026 ausdrücklich erteilt:
„phase 3 ist abgenommen. implementiere phase 4“. Der abgenommene Stand wird mit dem
folgenden lokalen Commit gesichert: `Phase 3: Zusatzleistungen im Buchungsablauf abschließen`.
