# BikeRental – Änderung zur MVP-Abnahme: Sofortbeginn

## Anlass und Status

**Status:** Implementiert und am 21.10.2026 technisch geprüft. Mit Phase 5 am 21.10.2026 ausdrücklich abgenommen.
Beim manuellen Prüfen veraltete „Mietbeginn jetzt“ während der Formulareingabe.
Ein zukünftiger Ersatztermin ließ anschließend keine sofortige Ausgabe zu.
Der Auftraggeber hat den Vorschlag „Sofort beginnen“ mit gewünschter Dauer sowie
„Für später reservieren“ mit festen Terminen zur Umsetzung beauftragt.

## Verbindliche Regel

Bei Stundenmiete wird ausdrücklich zwischen Sofortbeginn und festen Terminen
gewählt. Sofortbeginn ist die Voreinstellung, die gewünschte Dauer eine positive
ganze Stundenzahl (zunächst 1). Die Preisvorschau nennt „ab Buchungsabschluss“.
Erst beim erfolgreichen Speichern setzt der Server unter der Buchungssperre
Beginn auf seine aktuelle Zeit und Ende auf Beginn plus Dauer. Die volle Dauer
bleibt beim Warten vor Abschluss erhalten. Beginn und Buchungszeitpunkt sind gleich.

Die getrennte Ausgabe ist unmittelbar möglich, sofern das Fahrrad tatsächlich
zurückgegeben wurde. Mietzeit läuft bereits ab Buchung, auch wenn die Ausgabe
später erfasst wird. Ausgabe verschiebt weder Zeit noch Preis. Nach Speicherung
zeigt die Übersicht die verbindlichen Wiener Uhrzeiten. Vorschau reserviert nichts.
Verfügbarkeit, Typ, Tarifversion und Extras werden beim Abschluss erneut atomar
geprüft; ein inzwischen kollidierender Zeitraum erzeugt keine Teilbuchung.

Feste Termine und Tagesmieten behalten die bisherigen Regeln. Ein fester
zukünftiger Beginn erlaubt die Ausgabe weiterhin erst ab diesem Zeitpunkt.
Ein fester Stundenbeginn in der Vergangenheit bleibt ungültig. Lokale mehrdeutige
oder nicht existierende Termine werden weiter abgelehnt. Sofortbeginn verwendet
eindeutige Serverzeit und tatsächlich verstrichene Stunden, auch bei Zeitumstellung.

## Schnittstellen

Preisvorschau, Verfügbarkeitsabfrage und Buchung erhalten optional `startMode`
(`now` oder `scheduled`) und `rentalHours` (ganze positive Zahl bei `now`).
Fehlendes/leeres `startMode` bedeutet aus Kompatibilitätsgründen feste Termine.
Bei `now` muss `mode=hour` gelten, `startsOn`/`endsOn` bleiben leer. Bei festen
Terminen ist `rentalHours=0`; widersprüchliche Angaben werden abgelehnt.
Die Dauer darf den Wertebereich von Go `time.Duration` nicht überschreiten
(maximal 2.562.047 ganze Stunden).
Preis und Buchungsantwort nennen den Startmodus, Buchungen speichern ihn im Verlauf.

## Akzeptanz und Prüfplan

- Kontrollierte Store-Uhr: Preisprüfung, danach mehrere Minuten Wartezeit,
  unveränderter Preis, volle Dauer ab Abschluss und sofortige Ausgabe.
- Kollision erst durch das spätere Ende, konkurrierende Sofortbuchungen sowie
  Tarifänderung nach Vorschau müssen atomar scheitern.
- Ungültige Dauer, Mischangaben, Tagesmiete mit Sofortbeginn und unbekannter
  Startmodus werden zurückgewiesen. Zeitumstellung verändert verstrichene Dauer nicht.
- Browser: längere Wartezeit vor Abschluss simulieren, neue Auswahl und Vorschau,
  direkte Ausgabe, feste zukünftige Reservierung und mobile Bedienung.
- MVP-Prüfprogramm, Anforderungsmatrix, Mission, Roadmap und README aktualisieren;
  Go-Tests, Race, Vet und Browser-Gesamtprüfung erneut ausführen.

## Prüfnachweis

Technische Prüfung am **21.10.2026**, Go **1.27.3**, macOS/arm64:

- `go test -count=1 ./...`, `go test -race -count=1 ./...`, `go build ./...`
  und `go vet ./...` bestanden. JavaScript-Syntax aller vier Web-Assets und des
  MVP-Prüfprogramms geprüft. Keine neuen Abhängigkeiten für die Anwendung.
- [Store-Prüfungen](../BikeRental/internal/store/immediate_test.go): sieben Minuten
  und dreizehn Sekunden nach Vorschau buchen, volle Stunde ab aktuellem Abschluss,
  13 EUR mit Helm, Ausgabe zwei Minuten später ohne Zeit-/Preisverschiebung.
  Erst durch das spätere Ende entstehende Kollision und inzwischen geänderter
  Tarif scheitern ohne Teilbuchung. Zwölf parallele Sofortbuchungen: genau eine erfolgreich.
- [Domain-Prüfungen](../BikeRental/internal/domain/immediate_test.go): positive
  ganze Dauer, Überlaufgrenze, unzulässige Mischangaben, falsche Mietart und
  unbekannter Startmodus. Tatsächlich verstrichene Dauer bleibt über beide
  Zeitumstellungen erhalten. Bestehende feste Termine einschließlich Vergangenheitsprüfung unverändert.
- [HTTP-Prüfungen](../BikeRental/internal/handler/immediate_test.go): Vorschau,
  Verfügbarkeit und Abschluss mit Sofortbeginn; fehlerhafte Zahlen/JSON abgelehnt.
  `startsAt == createdAt`, volle Dauer, `canIssue=true`, unmittelbare Ausgabe.
- [MVP-Browserprogramm](../BikeRental/tools/mvp-browser.cjs), Google Chrome,
  Playwright **1.62.1**, Node **24.19.0**: die bisherigen 24 vollständigen Abläufe
  in Normal-/Demomodus bestanden, Stundenmieten jetzt mit Sofortbeginn.
  Zusätzlich je Modus eine Zwei-Stunden-Sofortmiete mit Helm (21 EUR), direkter
  Ausgabe und Rückgabe sowie feste zukünftige Reservierung mit Ausgabesperre
  und Stornierung. Ungültige Dauer bzw. fester vergangener Beginn erzeugen Fehler,
  korrigierte Angaben entfernen diese. Mobile Eingabe bei 390 × 844 ohne Überlauf.
- Im Normalmodus **65 Sekunden echte Wartezeit zwischen Vorschau und Abschluss**
  (`MVP_CONFIRM_DELAY_MS=65000`): Startzeit kommt aus dem Abschluss, Ende exakt
  zwei Stunden später, 21 EUR unverändert und Ausgabe unmittelbar möglich.
  Der Demomodus prüft denselben Ablauf ohne zusätzliche Wartezeit.
- Prozessneustarts und eingebettete Assets bestehen weiterhin. Browserzeitzone
  America/New_York, Anzeige weiterhin Wiener Zeit. Keine unbehandelten JavaScript-Fehler.

Bildnachweise: [Buchung mit verbindlichen Zeiten und Ausgabeaktion](screenshots/06-sofortbeginn.png),
[mobile Dauereingabe mit Erklärung](screenshots/06-sofortbeginn-mobil.png).

**Erneuter manueller Check:** Stundenmiete → „Sofort beginnen“ → 1 Stunde und Helm
wählen → Preis prüfen (MTB 13 EUR) → mindestens eine Minute warten → buchen →
Bestätigungslink öffnen → sofort ausgeben. Beginn muss dem Abschluss entsprechen,
Ende eine volle Stunde danach liegen. Bei „Für später reservieren“ bleibt die
Ausgabe bis zum gewählten Beginn gesperrt.

**Abnahme:** Mit Phase 5 am **21.10.2026** ausdrücklich erteilt:
„phase 5 ist hiermit abgenommen“. Die Änderung wird im Phase-5-Abschlusscommit gesichert.
