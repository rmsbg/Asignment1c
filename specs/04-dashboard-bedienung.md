# BikeRental – Phase 4: Dashboard und durchgängige Bedienung

## Status und Umfang

**Status:** Implementiert und am 21.10.2026 technisch geprüft. Vom Auftraggeber am 21.10.2026 ausdrücklich abgenommen.
Phase 3 ist ausdrücklich abgenommen und mit `2eb73e9` gesichert.
Grundlagen: [Mission](mission.md), [Roadmap](roadmap.md) und die abgenommenen
Feature-Spezifikationen der Phasen 1 bis 3. Phase 4 vervollständigt B-05/B-06
ohne Änderung der vereinbarten Miet-, Preis-, Zusatzleistungs- oder Speicherregeln.

## Dashboard und Navigation

Die bestehenden vier Kennzahlen bleiben maßgeblich: verfügbare und tatsächlich
ausgegebene Fahrräder, Gesamtbestand und registrierte Personen. Verfügbar plus
ausgegeben ergibt Gesamtbestand. Bloße Buchungen und Stornierungen verändern
den ausgegebenen Bestand nicht; Extras beeinflussen keine Bestandskennzahl.
Ein konsistenter Store-Snapshot liefert die Werte unter derselben Lesesperre.
Ausweisnachweise und Personenlisten gehören nicht ins Dashboard.

Die API `GET /api/dashboard` bleibt unverändert mit genau vier Ganzzahlfeldern.
Serverseitiges HTML enthält gültige Anfangswerte, auch ohne JavaScript. Mit
JavaScript wird sofort, alle fünf Sekunden und bei Rückkehr zum Tab/Fenster
aktualisiert. Nur eine Abfrage gleichzeitig, fünf Sekunden Timeout, Pause bei
verborgenem Tab. Fehlerhafte, negative, unvollständige oder inkonsistente
Antworten überschreiben die letzten gültigen Werte nicht. Eine sichtbare
Veraltet-Meldung bietet automatische Erholung und „Jetzt aktualisieren“.
Die Zeitangabe verwendet ausdrücklich Europe/Vienna.

Schnellzugriffe führen zu Fahrrädern, Personen, neuer Ausleihe, Tarifen und
„Buchungen und Rückgaben“. Bei leerem Inventar oder fehlenden Personen zeigt das
Dashboard konkrete Links zum Anlegen. Diese Hinweise folgen den aktualisierten
Kennzahlen. Der Speicher- und Demohinweis bleibt sichtbar.

## Durchgängiger Verleihablauf

1. Fahrrad mit Pflichtbild anlegen und Person registrieren. Aus dem Personendetail
   die neue Ausleihe mit vorausgewählter Person öffnen.
2. Mietart, Typ, Zeitraum und Extras wählen, Preis prüfen, Fahrrad mit Bild
   auswählen und zum angezeigten Preis buchen.
3. Die Speicherbestätigung bietet einen direkten Link zur gerade angelegten
   Buchung in der Übersicht. Die Buchungskarte erhält Fokus und ist ohne Suche
   für Ausgabe oder Stornierung erreichbar.
4. Ausgabe und Rückgabe bleiben ausdrückliche Aktionen. Rückkehr zum Dashboard
   aktualisiert die Kennzahlen. Auch aus dem Buchungsformular ist die Übersicht
   unmittelbar erreichbar; bei fehlender Verfügbarkeit führt ein Link zu Rückgaben.

Die bestehende Preisvorschau muss nach einer gültigen Neuprüfung korrigierte
Fehlertexte und `aria-invalid` entfernen. Bei Fehlern bleiben Eingaben erhalten;
es gibt keine Teilbuchung. Erfolgs- und Fehlermeldungen besitzen Status-/Alert-Rollen
und einen passenden Fokus. Tastatur bedient Navigation, Formulare, Bilderauswahl,
Extras und Aktionen. Bildfehler bleiben verständlich sichtbar, Name und Typ lesbar.

Fehlgeschlagene Listen-/Detailabfragen der Verwaltung bieten „Erneut laden“
ohne Browserneustart. Leere Listen, fehlende Tarife, ungültige Eingaben und
Buchungskonflikte zeigen klare nächste Schritte. Abbrechen verwirft Eingaben.

## Schnittstellen und Umsetzung

Keine neue Backend-API und keine neuen Abhängigkeiten. Die existierende fachliche
Dashboard-Berechnung und Tarif-/Buchungsverträge bleiben erhalten. Änderungen
betreffen eingebettetes HTML, JavaScript und CSS sowie relevante Prüffälle.

Die bestehende Ansicht `/ausleihe#uebersicht` bleibt nutzbar. Zusätzlich führt
`/ausleihe#uebersicht/{id}` unmittelbar zur fokussierten Buchungskarte. Die ID
wird als URL-Komponente kodiert und über DOM-ID gesucht. Unbekannte oder defekte
Ziele dürfen keinen JavaScript-Abbruch verursachen; die Übersicht bleibt bedienbar.
Direkte Links liefern keine Ausweisdaten.

## Aufgaben und Akzeptanzprüfung

1. Dashboard-Zustände und Schnellzugriffe einschließlich direktem Rückgabenweg
   ergänzen; Bestandslogik und automatisch aktualisierte Kennzahlen erhalten.
2. Speicherbestätigung mit direktem Buchungsziel, Fehlerbereinigung nach Neuprüfung
   und Wiederholen nach Verwaltungsfehlern implementieren.
3. HTTP-/Fachprüfungen für Bestandsinvarianten mit Stunden-/Tagesbuchung, Extras,
   Stornierung, Ausgabe und wiederholter Rückgabe; Datenschutz und SSR erhalten.
4. Browserprüfung Normal-/Demomodus, Desktop/mobil und Tastatur: leerer Start,
   Tarifeinrichtung, vollständiger Verleihfluss und direkte Buchungskarte.
5. Automatische Aktualisierung ohne Seitenneustart, Veraltet-Meldung bei
   Verbindungs-/Datenfehlern, Erholung und keine überlappenden Abfragen prüfen.
6. Fehlende Tarife, keine freien Fahrräder, Bildfehler und Verwaltungs-/Buchungsfehler
   prüfen; korrigierte Eingaben müssen alte Fehlermarkierungen entfernen.
7. Build, Tests, Race, Vet, Screenshots und Dokumentation. Erst nach ausdrücklicher
   Auftraggeberabnahme Commit `Phase 4: Dashboard und durchgängige Bedienung abschließen`.

## Prüfnachweis und Auftraggeberabnahme

Technische Prüfung am **21.10.2026**, Go **1.27.3**, macOS/arm64:

- `go fmt ./...`, `go build ./...`, `go test -count=1 ./...`,
  `go test -race -count=1 ./...` und `go vet ./...` bestanden.
  `node --check` für die drei geänderten JavaScript-Dateien bestanden.
- HTTP-Prüfungen bestätigen unveränderten Dashboard-Vertrag, Datenschutz,
  serverseitige Anfangswerte/Leerhinweise und alle fünf Schnellzugriffe.
  Künftige Stundenbuchung mit allen Extras und anschließende Stornierung verändern
  den ausgegebenen Bestand nicht. Tagesbuchung, Ausgabe und wiederholte Rückgabe
  halten die Bestandsinvariante ein.
- Browserprüfung mit lokalem Google Chrome/Playwright im Normal- und Demomodus:
  Tarife einrichten, Fahrrad mit Pflichtbild anlegen, Person registrieren,
  vorausgewählte Person übernehmen, Stundenmiete mit Helm buchen, direkten
  Speicherlink per Tastatur öffnen, Karte fokussieren, ausgeben und zurücknehmen.
  8 EUR Fahrradpreis plus 5 EUR Helm ergeben unverändert 13 EUR.
- Tatsächliche Buchungskollision durch einen zweiten Client erzeugt keine
  Teilbuchung. Nach Stornierung und neuer Preisprüfung gelingt die Erholung;
  alte Fehlertexte und `aria-invalid` verschwinden. Fehlender Tarif bleibt als
  Hinweis erhalten. Leerer Start, fehlende Personen, fehlende Verfügbarkeit,
  abgebrochener Bildabruf und 503-Verwaltungsfehler mit Wiederholen sind geprüft.
- Unabhängiger Client ändert Personenbestand bei offenem Dashboard: Werte folgen
  ohne Klick/Neuladen. Bloße Buchungen bleiben ungezählt; Ausgabe und Rückgabe
  aktualisieren den Bestand. Fehler 503, ungültige Kennzahlen und Timeout behalten
  letzte gültige Werte, markieren sie als veraltet und erholen sich automatisch.
  Mehrfache Aktualisierungsaktionen erzeugen keine parallelen Dashboard-Abfragen.
- Ohne JavaScript liefert das Dashboard korrekte Anfangswerte samt Hinweis.
  Browserzeitzone America/New_York verändert Wiener Zeitangaben nicht.
  Unbekannte und defekt kodierte Buchungsziele verursachen keinen JavaScript-Fehler.
- Desktop **1440 × 1000**, mobile Ansicht **390 × 844**, Skip-Link,
  Tastaturauswahl und Fokusziele geprüft; kein horizontaler Überlauf und keine
  unbehandelten JavaScript-Fehler. Die mobile Ansicht ist eine Browseremulation,
  kein Test auf physischer Hardware; ein Screenreader-Test wurde nicht durchgeführt.

Temporäres Prüfskript: `/private/tmp/bikerental-phase4-browser.cjs`;
Serverlogs und Testbinary: `/private/tmp/bikerental-phase4-checks/`.
Diese Hilfsdateien sind keine Repository-Bestandteile. Die Go-Prüffälle bleiben
im Repository. Das Browserprogramm baut das Binary selbst und startet jeden Modus
mit frischem Store; es verwendet ausschließlich synthetische Daten.

Bildnachweise:

- [Leeres Dashboard mit nächsten Schritten](screenshots/04-dashboard-leer.png)
- [Dashboard nach Ausgabe, Desktop](screenshots/04-dashboard-desktop.png)
- [Direkt fokussierte Buchung für Ausgabe](screenshots/04-buchung-direktziel.png)
- [Dashboard mit veralteten Werten](screenshots/04-dashboard-veraltet.png)
- [Verwaltungsfehler mit erneutem Ladeversuch](screenshots/04-verwaltung-fehler.png)
- [Dashboard, mobile Ansicht](screenshots/04-dashboard-mobil.png)
- [Buchungsübersicht nach Rückgabe, mobile Ansicht](screenshots/04-rueckgaben-mobil.png)

**Auftraggeberabnahme:** Am 21.10.2026 ausdrücklich erteilt:
„phase 4 ist hiermit abgenommen. implementiere phase 5“.
Der abgenommene Stand wird mit `Phase 4: Dashboard und durchgängige Bedienung abschließen` lokal committet.
Die abschließende Anforderungsmatrix und MVP-Gesamtprüfung gehören zu Phase 5.
