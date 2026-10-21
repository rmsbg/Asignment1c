# BikeRental – SDD-Roadmap

## Vorgehen und aktueller Stand

Phase 0 ist technisch abgeschlossen, geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. Phase 1 ist implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. Phase 2 ist implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. Phase 3 ist implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. Phase 4 ist implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen. Phase 5 ist noch offen. Als technische Ausgangsbasis dient der vorhandene
SkiRental-Stand bis einschließlich Phase 5. Dessen dokumentierte Abnahmen sind
historische Nachweise für SkiRental, keine BikeRental-Abnahmen. Diese Roadmap
ordnet Arbeit nach Abhängigkeiten und enthält keine Terminzusagen.

Für jede Phase gilt: **Spezifizieren → Aufgaben ableiten → Implementieren →
Prüfen → Abnehmen**. Vor Codeänderungen entsteht eine Feature-Spezifikation in
BikeRentals `specs/` mit Nutzerablauf, Datenmodell, Geschäftsregeln, API-Verhalten,
Fehlerfällen und überprüfbaren Akzeptanzkriterien. Die nachfolgend genannten
Feature-Dateien ab Phase 5 sind geplant und noch nicht angelegt. Spezifikationen
und Prüfnachweise der Phasen 0 bis 4 sind bereits vorhanden.

Technische Prüfergebnisse und ausdrückliche Auftraggeberabnahme werden getrennt
festgehalten. Nach Abschluss und ausdrücklicher Auftraggeberabnahme jeder Phase
aktualisiert der ausführende Agent Prüfnachweis und Status und erstellt ohne
weitere Rückfrage einen lokalen Git-Commit mit den zugehörigen Änderungen.
Dies gilt für alle Phasen einschließlich Phase 0. Technischer Abschluss allein
ersetzt die Abnahme nicht. Commit-Nachrichten benennen Phase und Ergebnis,
beispielsweise `Phase 1: Fahrradverwaltung mit Bildern abschließen`.

## Phase 0 – Grundlagen und Wiederverwendungsplan

**Status:** Technisch abgeschlossen, geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen: „phase 0 ist hiermit abgenommen“.
Der Stand wird mit `Phase 0: Grundlagen und Wiederverwendungsbasis abschließen`
lokal committet.
**Spezifikation und Prüfnachweis:** [00-grundlagen.md](00-grundlagen.md).

**Ergebnis:** Abgestimmte Geschäftsregeln und ein reproduzierbarer Ausgangsstand.
Die Arbeitskopie liegt in `BikeRental/` (Go-Modul `bikerental`), die Referenz bleibt
in `SkiRental/`; Git für BikeRental und Verfassung liegt im obersten Verzeichnis.

**Abnahme:**
- Wiederverwendung und notwendige Änderungen sind je Bereich dokumentiert:
  Personen, Inventar, Buchung, Tarife, Extras, Bilder und Dashboard.
- Arbeitsverzeichnis, Modulname und Git-Struktur für BikeRental sind entschieden;
  der SkiRental-Ausgangsstand bleibt nachvollziehbar gesichert.
- Die in Phase 0 festgelegten Regeln zu Mietdauer, Tarifgestaltung, Extras,
  Versicherung und Bildpflege sind mit konkreten Beispielen dokumentiert.
- Stunden-/Tagesüberschneidungen, Buchungsgrenzen und Zeitumstellungen haben
  eindeutige erwartete Ergebnisse. Ungeklärte Regeln blockieren die betroffene
  Implementierung, nicht unabhängige Arbeiten.
- Ausgangs-Build, Tests, Race und Vet sind mit verwendeter Toolchain protokolliert.

## Phase 1 – Fahrradverwaltung und Bilder

**Status:** Implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen: „phase 1 ist hiermit abgenommen“.
Gesichert mit `b8901d2` (`Phase 1: Fahrradverwaltung mit Bildern abschließen`).
**Spezifikation und Prüfnachweis:** [01-fahrradverwaltung.md](01-fahrradverwaltung.md).

**Ergebnis:** Wiederverwendete Inventarverwaltung für Fahrräder und bestehende
Personenverwaltung als Grundlage für B-01, B-03 und B-05.

**Abnahme:**
- Fahrräder können mit eindeutiger ID, Bezeichnung, einem der drei Typen
  Downhill/MTB/Rennrad und Bild angelegt, angezeigt und bearbeitet werden.
- Bildpflege und fehlende Bilder folgen der in Phase 0 vereinbarten Regel.
  Ungültige Eingaben oder Bilder verändern keine gespeicherten Daten.
- Die Fahrradauswahl zeigt das zum jeweiligen Fahrrad gehörende Bild, Namen und
  Typ. Personenverwaltung und beide Ausweisvarianten funktionieren weiterhin.
- Oberfläche und Beispieldaten verwenden Fahrradbegriffe; Bilder sind auf
  Desktop und kleinen Bildschirmen erkennbar und zugänglich beschriftet.

## Phase 2 – Stunden-/Tagesmiete und typabhängige Preise

**Status:** Implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen: „phase 2 abgenommen“.
Gesichert mit `be2d6de` (`Phase 2: Mietzeiten und typabhängige Preise abschließen`).
**Spezifikation und Prüfnachweis:** [02-mietzeiten-preise.md](02-mietzeiten-preise.md).

**Ergebnis:** Durchgängige Buchung, Ausgabe und Rückgabe mit den Mietarten und
Preisen aus B-01, B-02 und B-03.

**Abnahme:**
- Beide Mietarten und alle drei Fahrradtypen liefern für die vereinbarten
  Beispieldauern den erwarteten Preis und eine verständliche Vorschau.
- Ungültige Zeiträume, Rundungsgrenzen, Mitternacht und Zeitumstellungen sind
  entsprechend den beschlossenen Regeln geprüft.
- Überschneidende Stunden-/Tagesbuchungen desselben Fahrrads scheitern auch bei
  parallelen Anfragen; erlaubte Anschlussbuchungen gelingen.
- Tarifänderungen zwischen Vorschau und Abschluss erfordern erneute Bestätigung;
  bestehende Buchungen behalten ihren vereinbarten Preis.
- Ausgabe, Rückgabe, Stornierung und Verlauf funktionieren weiterhin.
  Überfällige Rückgaben verhindern eine weitere tatsächliche Ausgabe.

## Phase 3 – Zusatzleistungen im Buchungsablauf

**Status:** Implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen: „phase 3 ist abgenommen“.
Gesichert mit `2eb73e9` (`Phase 3: Zusatzleistungen im Buchungsablauf abschließen`).
**Spezifikation und Prüfnachweis:** [03-zusatzleistungen.md](03-zusatzleistungen.md).

**Ergebnis:** Helm, Schutzausrüstung und Versicherung sind gemäß B-04 in den
Verleihvorgang integriert.

**Abnahme:**
- Alle drei Optionen werden bei jeder neuen Buchung sichtbar abgefragt.
  Eine Buchung ohne Extras sowie jede erlaubte Kombination sind möglich.
- Vorschau und Abschluss zeigen Fahrradpreis, gewählte Extras und Gesamtpreis;
  Buchungsdetails erhalten die vereinbarte Auswahl und Preisaufteilung.
- Änderungs- und Konfliktfälle erzeugen keine Teilbuchungen. Jede Option wird
  genau einmal pro Buchung berechnet; Zubehörbestände werden nicht geführt.
- Die Versicherungsdarstellung entspricht dem vereinbarten Umfang und behauptet
  keine nicht implementierte externe Versicherungsabwicklung.

## Phase 4 – Dashboard und durchgängige Bedienung

**Status:** Implementiert, technisch geprüft und am 21.10.2026 ausdrücklich
vom Auftraggeber abgenommen: „phase 4 ist hiermit abgenommen“.
**Spezifikation und Prüfnachweis:** [04-dashboard-bedienung.md](04-dashboard-bedienung.md).

**Ergebnis:** Übersichtlicher Gesamtfluss gemäß B-05/B-06 und konsistente Kennzahlen.

**Abnahme:**
- Dashboard: verfügbare und ausgeliehene Fahrräder, Gesamtbestand und registrierte
  Personen. Verfügbar plus ausgeliehen entspricht dem Gesamtbestand; bloße
  Buchungen zählen nicht als ausgegeben.
- Vorhandene automatische Aktualisierung und Kennzeichnung veralteter Werte
  bleiben funktionsfähig; Schnellzugriffe führen in die passenden Fahrradansichten.
- Person registrieren, Mietart wählen, Fahrrad mit Bild auswählen, Extras prüfen,
  buchen, ausgeben und zurücknehmen ist ohne Umwege möglich.
- Leere Listen, fehlende Tarife, Bildprobleme und Buchungskonflikte werden
  verständlich behandelt. Desktop, mobile Darstellung und Tastaturbedienung
  sind im Browser mit Prüfnachweis kontrolliert.

## Phase 5 – MVP-Gesamtprüfung und Dokumentation

**Status:** Offen. **Geplante Spezifikation:** `05-mvp-abnahme.md`.

**Ergebnis:** Vorführbarer BikeRental-Prototyp mit nachvollziehbarem Prüfnachweis
für alle Anforderungen B-01 bis B-06.

**Abnahme:**
- Eine Anforderungsmatrix verknüpft B-01 bis B-06 mit Tests und Browsernachweisen.
  Der Gesamtfluss deckt beide Mietarten, alle Fahrradtypen, Extras mit/ohne
  Auswahl sowie beide Ausweisvarianten ab.
- Normal- und Demomodus, Datenverlust beim Neustart und ein eigenständig
  startbarer Build mit eingebetteter Oberfläche sind geprüft.
- `go build ./...`, `go test ./...`, `go test -race ./...` und `go vet ./...`
  bestehen. Fachliche Grenzfälle und unveränderte Personenfunktionen sind abgedeckt.
- README dokumentiert tatsächlichen Arbeitsort, Voraussetzungen, Start,
  Prüfkommandos, Bedienung und Grenzen. Screenshots verwenden synthetische Daten.
- Abweichungen sind explizit dokumentiert; technische Fertigstellung ersetzt
  keine ausdrückliche Auftraggeberabnahme.

## Spätere Erweiterungen

Dauerhafte Speicherung, Benutzerkonten, Kundenportal, Zahlungen, Rechnungen und
Anbindungen an Versicherungsanbieter benötigen eine eigene Priorisierung und
eine bewusste Erweiterung der Mission.
