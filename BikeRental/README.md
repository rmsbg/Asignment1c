# BikeRental – Arbeitskopie

Das Go-Modul `bikerental` liegt hier. Verfassung, Roadmap, Startanleitung und
Phasennachweise werden zentral im übergeordneten Projekt gepflegt:

- [README und Startanleitung](../README.md)
- [Mission](../specs/mission.md)
- [Technische Verfassung](../specs/tech-stack.md)
- [Phase 0](../specs/00-grundlagen.md)

**Phase 5:** MVP-Gesamtprüfung mit Anforderungsmatrix, HTTP-Prüfung aller Typen,
Mietarten und Extras mit/ohne Auswahl sowie reproduzierbarem Browserprogramm
`tools/mvp-browser.cjs`. Normal-/Demostart, eigenständiges Binary und Datenverlust
beim Prozessneustart sind Bestandteil der Prüfung. Keine neuen Laufzeitabhängigkeiten.

- [MVP-Anforderungsmatrix und Prüfnachweis](../specs/05-mvp-abnahme.md)
- [Nachbesserung: Sofortbeginn ab Buchungsabschluss](../specs/06-sofortbeginn.md)
- [Browserprüfung wiederholen](../README.md#browser-gesamtprüfung-wiederholen)

Phase 5 einschließlich Sofortbeginn wurde am 21.10.2026 ausdrücklich abgenommen. Phase 4 ist mit `340a53b` gesichert.
**Phase 6 – Buchen und ausgeben:** Implementiert, technisch geprüft und am
22.10.2026 ausdrücklich vom Auftraggeber abgenommen. Bei Tagesmiete ab heute und
Stundenmiete mit „Sofort beginnen“ stehen „Buchen und ausgeben“ und „Nur buchen“
zur Verfügung; für spätere Termine „Buchen“.
[Verhalten, Abnahme und Prüfnachweis](../specs/07-buchen-und-ausgeben.md).

Start aus diesem Verzeichnis: `go run ./cmd/server`, optional mit `-demo`.
