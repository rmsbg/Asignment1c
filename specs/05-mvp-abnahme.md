# BikeRental – Phase 5: MVP-Gesamtprüfung und Dokumentation

## Status und Ziel

**Status:** Implementiert, technisch geprüft und am 21.10.2026 ausdrücklich vom Auftraggeber abgenommen.
Die manuelle Prüfung führte zur beauftragten [Änderung Sofortbeginn](06-sofortbeginn.md).
Diese Nachbesserung ist von der ausdrücklichen Phase-5-Abnahme umfasst.
Phase 4 wurde ausdrücklich abgenommen und mit `340a53b` gesichert.
Grundlagen sind [Mission](mission.md), [technische Verfassung](tech-stack.md),
[Roadmap](roadmap.md) und die Feature-Spezifikationen der Phasen 0 bis 4.

Phase 5 liefert einen reproduzierbaren Nachweis für B-01 bis B-06. Sie ergänzt
keine neuen Geschäftsregeln, APIs, Datenfelder oder Infrastruktur. Festgestellte
Fehler werden im vereinbarten Umfang behoben und erneut geprüft. Technische
Fertigstellung und ausdrückliche Auftraggeberabnahme bleiben getrennt.

## Aufgaben und Akzeptanzkriterien

1. Anforderungen mit konkreten Fach-/HTTP-Tests und Browsernachweisen verknüpfen.
2. Den HTTP-Gesamtfluss über alle drei Typen, beide Mietarten und Buchungen mit
   und ohne Extras prüfen: Bild, Person, Preisvorschau, Verfügbarkeit, Buchung,
   Ausgabe, Tarifänderung, Rückgabe und unveränderter Preisverlauf.
3. Einen im Repository gespeicherten Browserprüflauf bereitstellen. Er startet
   das kompilierte Binary außerhalb des Quellbaums, prüft Normal- und Demomodus,
   manuelle und fotografische Ausweisnachweise, beide Mietarten und alle Typen.
4. Echte Prozessneustarts müssen Daten einschließlich Uploads, Buchungen und
   Tarifänderungen verwerfen; Demo-Neustart stellt ausschließlich Beispieldaten her.
   Das Binary muss HTML, JavaScript, CSS und Demoillustrationen eingebettet ausliefern.
5. Build, Tests, Race Detector, Vet und JavaScript-Syntaxprüfung bestehen.
   Bestehende Grenzfall-/Nebenläufigkeitsprüfungen bleiben Teil der Gesamtprüfung.
6. README enthält tatsächliche Verzeichnisse, Voraussetzungen, Start- und
   Prüfkommandos, Bedienung und Grenzen. Screenshots verwenden synthetische Daten.
7. Abweichungen und verbleibende Grenzen ausdrücklich dokumentieren. Nach
   Auftraggeberabnahme folgt `Phase 5: MVP-Gesamtprüfung und Dokumentation abschließen`.

## Prüfverfahren

Die Go-Prüfungen bleiben ohne zusätzliche Go-Abhängigkeiten ausführbar. Ein
Browserprogramm unter `BikeRental/tools/mvp-browser.cjs` verwendet optional Node.js
und Playwright ausschließlich zur Prüfung. Diese sind keine Laufzeitabhängigkeiten
der Anwendung. Port 8080 muss vor dem Prüflauf frei sein; ein bestehender Server
wird vom Skript nicht beendet. Temporäre Server und Browser werden aufgeräumt.

Desktop, mobile Browseransicht und Tastatur werden geprüft. Die umfassenden
Fehler-/Dashboard-Nachweise aus [Phase 4](04-dashboard-bedienung.md) ergänzen
den abschließenden Gesamtfluss. Die nachträgliche Prüfung des Sofortbeginns
ist unter [06-sofortbeginn.md](06-sofortbeginn.md) dokumentiert. Physische Mobilgeräte und Screenreader sind
damit nicht geprüft. Keine Behauptung vollständiger Fehlerfreiheit.

## Anforderungsmatrix und Prüfnachweis

Alle Anforderungen B-01 bis B-06 sind im vereinbarten lokalen Prototyp umgesetzt
und technisch geprüft. Die Matrix unterscheidet automatisierte Fach-/HTTP-Prüfung
und Browserbeobachtung; sie ist keine ausdrückliche Auftraggeberabnahme.

| Anforderung | Automatisierter Nachweis | Browsernachweis |
| --- | --- | --- |
| B-01 Fahrradverleih | `TestMVPLifecycleMatrix` in [mvp_matrix_test.go](../BikeRental/internal/handler/mvp_matrix_test.go): Bild/Person → Vorschau → Buchung → Ausgabe → Tarifänderung → Rückgabe → Verlauf; [dashboard_test.go](../BikeRental/internal/handler/dashboard_test.go) prüft Bestandszahlen | `workflow` im [MVP-Prüfprogramm](../BikeRental/tools/mvp-browser.cjs): alle 24 Vorgänge einschließlich getrennter Ausgabe/Rückgabe |
| B-02 Stunden-/Tagesmiete | [period_test.go](../BikeRental/internal/domain/period_test.go): Rundung, Kalender, Mitternacht, Sommer-/Winterzeit; [mixed_rental_test.go](../BikeRental/internal/store/mixed_rental_test.go): gemischte Grenzen, Anschluss, parallele Konflikte; HTTP-Matrix prüft 75 Minuten als zwei Stunden | Stundenmiete „Sofort beginnen“ ab Buchungsabschluss und Tagesmiete heute für jeden Typ mit und ohne Extras |
| B-03 Typabhängige Preise | [typed_pricing_test.go](../BikeRental/internal/handler/typed_pricing_test.go), HTTP-Matrix und [pricing_test.go](../BikeRental/internal/domain/pricing_test.go): drei Typen, Stundenpreis, Tagesstaffel und Zusatztag | Alle drei Typen mit erwarteten Demo-Preisen, Bild und korrektem Gesamtpreis |
| B-04 Zusatzleistungen | `TestHTTPAllExtraSelectionsForEveryModeAndType` in [extras_test.go](../BikeRental/internal/handler/extras_test.go): alle acht Kombinationen; [Store-Prüfungen](../BikeRental/internal/store/extras_test.go): Preisbindung und parallele Gebührenänderung | Je Typ/Mietart einmal ohne Extras und einmal alle drei gewählt; nächste Buchung ohne Vorauswahl, Einzelpauschalen und Verlauf sichtbar |
| B-05 Fahrradbilder | [bike_test.go](../BikeRental/internal/handler/bike_test.go): Upload/Ersetzen/Beibehalten und atomare Fehler; [validation_test.go](../BikeRental/internal/domain/validation_test.go): Inhalt, Format, Größen-/Pixelgrenzen | Sichtbare und dekodierte Bilder in jeder Auswahl; [Auswahlbild](screenshots/05-mvp-fahrradauswahl.png); Bildfehler zusätzlich in Phase 4 geprüft |
| B-06 Übersichtlichkeit | [dashboard_test.go](../BikeRental/internal/handler/dashboard_test.go): vier Kennzahlen, Datenschutz, fünf Einstiege und SSR; [acceptance_test.go](../BikeRental/internal/handler/acceptance_test.go): beide Ausweisvarianten einschließlich PNG/JPEG | Personenvorauswahl, Tastatur für Extras/Fahrrad/Speicherlink, direkte Buchungskarte, mobile Karte; [Phase 4](04-dashboard-bedienung.md) ergänzt Leer-/Fehlerzustände, Fokus und Dashboard-Erholung |

### Gesamtergebnis am 21.10.2026

Umgebung: macOS/arm64, Go **1.27.3**, Node **24.19.0**, Playwright **1.62.1**,
lokales Google Chrome. Browserzeitzone **America/New_York**, Geschäftsdarstellung
weiterhin ausdrücklich Wiener Zeit.

| Prüfung aus `BikeRental/` | Ergebnis |
| --- | --- |
| `go fmt ./...` | Bestanden; Go-Dateien formatiert |
| `go build ./...` | Bestanden |
| `go test -count=1 ./...` | Bestanden, alle Domain-/Handler-/Store-Prüfungen einschließlich neuer HTTP-Matrix |
| `go test -race -count=1 ./...` | Bestanden, keine gemeldeten Datenrennen |
| `go vet ./...` | Bestanden |
| `node --check` für alle vier Web-JavaScript-Dateien und `tools/mvp-browser.cjs` | Bestanden |
| `tools/mvp-browser.cjs`, Normalmodus | Bestanden: zwölf vollständige Verleihabläufe und echter Neustart |
| `tools/mvp-browser.cjs`, Demomodus | Bestanden: zwölf vollständige Verleihabläufe und echter Neustart |

Der im Repository gespeicherte Browserlauf baut selbst ein Binary und startet es
außerhalb des Quellverzeichnisses. HTML, vier JavaScript-Dateien, CSS und Demo-PNG
werden ohne benachbarte Asset-Dateien ausgeliefert. Neue Fahrräder mit Pflichtbild
und beide Personenvarianten entstehen über die Oberfläche. Nach jedem Buchen
bleibt der ausgegebene Bestand unverändert; Ausgabe erhöht und Rückgabe senkt ihn.
Alle zwölf gespeicherten Vorgänge je Modus sind im Verlauf vorhanden und behalten
Referenzen, Extras und Preis. Eine anschließende Gebührenänderung verändert den
Verlauf nicht. Personenlisten enthalten nur ID/Name, keine Ausweisnummern.

Nach vollständigem Prozessneustart sind neue Personen, Fahrräder, Uploads und
Buchungen verschwunden. Der Normalmodus hat wieder Tarifversion 0 und leere Listen;
der Demomodus Tarifversion 1, drei Fahrräder mit eingebetteten Bildern und eine
synthetische Person. Die geänderte Helmgebühr wird verworfen. Kein automatisches
Wiederherstellen von Nutzereingaben und keine unbemerkte Persistenz.

Desktop **1440 × 1000**, mobile Karte **390 × 844**: Bilder, Einzelpreise und
Schaltflächen bleiben lesbar, kein horizontaler Überlauf. Keine unbehandelten
JavaScript-Fehler. Die detaillierten Fehler-/Dashboard-Prüfungen aus Phase 4 gelten
weiter; die Dashboard-Produktionslogik wurde in Phase 5 nicht verändert. Die ergänzte
Stundenbuchung ist gesondert in der Änderung Sofortbeginn dokumentiert.

### Nachprüfung des Sofortbeginns

Nach dem manuellen Fund wurden die volle Go-/Race-/Vet-Prüfung und das
MVP-Browserprogramm erneut erfolgreich ausgeführt. Der Browserlauf enthält nun
zusätzlich Sofortabschluss nach 65 Sekunden echter Wartezeit, volle Dauer,
unmittelbare Ausgabe und feste zukünftige Reservierung mit Ausgabesperre.
Ergebnisse, API-Ergänzung und Screenshots: [06-sofortbeginn.md](06-sofortbeginn.md).
Die ergänzte Regel ersetzt den früheren starren „Mietbeginn jetzt“-Zeitstempel.

### Bildnachweise

- [Fahrradauswahl mit Bildern, Name und Typ](screenshots/05-mvp-fahrradauswahl.png)
- [Preisverlauf nach Rückgabe mit allen Extras](screenshots/05-mvp-preisverlauf.png)
- [Zurückgegebene Buchung, mobile Karte](screenshots/05-mvp-mobil.png)
- [Sauberer Demostart nach echtem Prozessneustart](screenshots/05-mvp-neustart.png)
- Ergänzende [Desktop-, Mobil- und Fehlernachweise aus Phase 4](04-dashboard-bedienung.md)

Wiederholung: [README-Anleitung](../README.md#browser-gesamtprüfung-wiederholen).
Die Screenshots enthalten ausschließlich synthetische Daten und selbst erstellte
Demoillustrationen. Frühere temporäre Skripte sind für diesen Gesamtprüflauf
nicht erforderlich.

## Abweichungen und verbleibende Grenzen

Keine funktionale Abweichung vom vereinbarten MVP festgestellt. Die folgenden
Grenzen sind Teil des vereinbarten Umfangs und bleiben sichtbar dokumentiert:

- Flüchtiger In-Memory-Store einschließlich Uploads und Tarifen; keine Datenbank.
- Lokaler Mitarbeiter-Prototyp auf Loopback ohne Benutzerkonten; kein produktiver Betrieb.
- Keine Zahlungen, Rechnungen, Kundenportal, Zubehörbestände oder externe Versicherungsabwicklung.
- JavaScript für Verwaltung/Buchung erforderlich; Dashboard-Anfangswerte serverseitig.
- Mobile Darstellung im Browser emuliert; keine Prüfung auf physischem Gerät oder mit Screenreader.
- Das separat referenzierte `SkiRental/` bleibt unverändert und ist im obersten Git
  ausgeschlossen; bei Weitergabe dieser historischen Referenz separat bereitstellen.

Die ursprüngliche Phase-5-Arbeit umfasste Prüfungen, Dokumentation und aktuelle
Phasenkennzeichnung. Aus der manuellen Prüfung kam die ausdrücklich beauftragte
Nachbesserung für Sofortbeginn hinzu: [Regel, API-Ergänzung und Prüfnachweis](06-sofortbeginn.md).
Die feste Terminbuchung und der übrige MVP-Umfang bleiben erhalten.

## Auftraggeberabnahme

Am **21.10.2026** ausdrücklich erteilt: „phase 5 ist hiermit abgenommen“.
Die Abnahme umfasst den MVP einschließlich der dokumentierten Sofortbeginn-Nachbesserung.
Der technisch geprüfte und abgenommene Stand wird mit
`Phase 5: MVP-Gesamtprüfung und Dokumentation abschließen` lokal committet.
