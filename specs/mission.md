# BikeRental – Mission und Produktverfassung

## Zweck und Geltungsbereich

BikeRental unterstützt Mitarbeiter eines Fahrradverleihs beim Verwalten von
Fahrrädern und Kunden sowie beim Buchen, Ausgeben und Zurücknehmen von Fahrrädern.
Einfache Bedienung, Übersichtlichkeit und kostensparende Wiederverwendung stehen
im Mittelpunkt.

Diese Verfassung bildet die Grundlage der spezifikationsgetriebenen Entwicklung
(SDD). Verbindliche Kundenanforderungen stammen aus [README.md](../README.md).
Ausgangspunkt ist der vorhandene SkiRental-Stand einschließlich Terminbuchungen
und Preisen aus Phase 5, beschrieben in [dessen Verfassung](../SkiRental/specs/mission.md).
[tech-stack.md](tech-stack.md) legt die technischen Leitplanken fest,
[roadmap.md](roadmap.md) die Umsetzung und Abnahme. Diese Dokumente beschreiben
den BikeRental-Zielzustand; sie behaupten keine bereits erfolgte Umsetzung oder Abnahme.

## Verbindliche Prinzipien

1. **Wiederverwenden vor Neuentwickeln:** Vorhandene Abläufe, Komponenten und
   Tests werden übernommen und gezielt angepasst. Neue Infrastruktur und ein
   vollständiger Neubau widersprechen dem Kostenziel.
2. **Bedienbarkeit zuerst:** Häufige Aufgaben sind direkt erreichbar. Deutsche
   Beschriftungen, erkennbare Fahrräder und verständliche Fehler unterstützen
   Mitarbeiter ohne technische Vorkenntnisse.
3. **Korrekte Buchungen und Preise:** Verfügbarkeit und Preis werden serverseitig
   geprüft. Ungültige oder konkurrierende Aktionen dürfen keine Teiländerungen erzeugen.
4. **Spezifikation vor Implementierung:** Jede Phase erhält überprüfbare
   Akzeptanzkriterien. Offene Geschäftsregeln werden vor der betroffenen Umsetzung geklärt.
5. **Datensparsam entwickeln:** Tests und Vorführungen verwenden synthetische
   Kundendaten und selbst erstellte oder zulässig verwendbare Fahrradbilder.

## MVP aus dem Kundengespräch

- **B-01 – Fahrradverleih:** Inventar, Buchung, Ausgabe, Rückgabe und Dashboard
  beziehen sich auf einzelne Fahrräder statt Ski-Paare.
- **B-02 – Mietdauer:** Mitarbeiter können eine stundenweise oder tageweise
  Vermietung wählen. Verfügbarkeit und Preis berücksichtigen die gewählte Mietart.
- **B-03 – Fahrradtypen:** Jedes Fahrrad gehört zu Downhill, MTB oder Rennrad.
  Die Preisberechnung unterscheidet diese Typen.
- **B-04 – Zusatzleistungen:** Helm, Schutzausrüstung und Versicherung werden bei
  jeder Buchung ausdrücklich zur Auswahl angeboten. Die Auswahl ist vor Abschluss
  zusammen mit ihren Preisfolgen sichtbar und danach in der Buchung nachvollziehbar.
- **B-05 – Fahrradbilder:** Bei der Fahrradauswahl im Verleihvorgang ist zu jedem
  Fahrrad sein Bild sichtbar; Name und Typ ergänzen die Darstellung.
- **B-06 – Übersichtlichkeit:** Ein durchgängiger Ablauf führt von Person und
  Mietzeit über Fahrradauswahl und Extras zur verständlichen Buchungsübersicht.

## Übernommene Grundlagen und ergänzende Leitplanken

Personenverwaltung mit automatisch erzeugten IDs und den bisherigen alternativen
Ausweisnachweisen bleibt erhalten. Ein Inventareintrag bezeichnet genau ein
Fahrrad. Eine Buchung verbindet eine Person und ein Fahrrad; mehrere Buchungen
einer Person sind möglich.

Buchung, tatsächliche Ausgabe, Rückgabe und Stornierung bleiben getrennte
Schritte. Überschneidende offene Buchungen desselben Fahrrads sind ausgeschlossen,
auch zwischen Stunden- und Tagesmieten. Eine verspätete Rückgabe verhindert eine
erneute tatsächliche Ausgabe. Es gibt keine automatische Ausgabe oder Rückgabe.

Vereinbarte Preise und Zusatzleistungen werden bei Buchung festgehalten; spätere
Tarifänderungen verändern bestehende Buchungen nicht. Das Dashboard zählt
tatsächlich ausgegebene Fahrräder: verfügbar plus ausgeliehen ergibt den
Gesamtbestand. Terminbezogene Verfügbarkeit wird gesondert im Buchungsablauf geprüft.

## Festlegungen aus Phase 0

Die Entscheidungen aus dem Projektgespräch und die technischen Präzisierungen
stehen mit Beispielen in [00-grundlagen.md](00-grundlagen.md):

- Stundenmiete: minutengenaue Eingabe, mindestens eine Stunde, angefangene Stunden
  aufrunden; auch über Mitternacht. Tatsächlich verstrichene Zeit ist maßgeblich.
- Tagesmiete: Kalendertage einschließlich beider Randtage wie bei SkiRental.
  Stunden- und Tagesmieten teilen dieselbe Verfügbarkeitsprüfung. Direkte Anschlüsse
  ohne Puffer sind zulässig; tatsächliche Rückgabe bleibt Voraussetzung für neue Ausgabe.
- Lokale Zeit: `Europe/Vienna`; nicht existierende oder doppeldeutige Eingabezeiten
  bei Zeitumstellungen werden zurückgewiesen.
- Preise: je Fahrradtyp ein bearbeitbarer Stundenpreis und eine Tagesstaffel mit
  Zusatztag; keine automatische Tarifwahl. Demo-Beispielbeträge werden in der
  Tarifphase festgelegt. Der bestätigte Preis bleibt auch bei abweichender Ausgabe
  oder Rückgabe unverändert.
- Extras: Helm, Schutzausrüstung und Versicherung unabhängig wählbar, jeweils eine
  bearbeitbare Pauschale pro Buchung; keine eigene Bestandsverwaltung und keine
  externe Versicherungsabwicklung.
- Bilder: Pflichtbild beim Anlegen, per Upload ersetzbar, JPEG/PNG bis 5 MiB und
  20 Millionen Pixel; wie übrige Daten nur im Arbeitsspeicher.

Die Regeln sind die Grundlage der Folgephasen und in Phase 0 noch nicht als
Fahrradfunktionen implementiert. Detail-Spezifikationen dürfen sie nicht
stillschweigend ändern.

## Grenzen und Erfolgskriterien

Wie SkiRental bleibt BikeRental zunächst ein lokaler Mitarbeiter-Prototyp mit
flüchtiger Speicherung. Benutzerkonten, Kundenportal, Zahlungen, Rechnungen,
dauerhafte Speicherung und produktiver Betrieb sind nicht Teil dieses MVP.
Eine Versicherungsoption dokumentiert die Buchungsauswahl; die Anbindung eines
Versicherungsanbieters ist nicht beauftragt.

Erfolg bedeutet: Mitarbeiter können ein Fahrrad mit Bild auswählen, beide
Mietarten nutzen, typabhängige Preise prüfen, Extras wählen und den vollständigen
Ausgabe-/Rückgabeablauf mit korrektem Dashboard durchführen. Doppelbuchungen
müssen auch bei gleichzeitigen Anfragen scheitern.

## Umgang mit Änderungen

Änderungen an Umfang oder Prinzipien werden zuerst in der Verfassung begründet.
Feature-Spezifikationen müssen dazu passen. Technische Prüfung und ausdrückliche
Auftraggeberabnahme werden getrennt dokumentiert; SkiRental-Abnahmen gelten
nicht automatisch für BikeRental.
