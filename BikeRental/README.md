# BikeRental – Arbeitskopie

Das Go-Modul `bikerental` liegt hier. Verfassung, Roadmap, Startanleitung und
Phasennachweise werden zentral im übergeordneten Projekt gepflegt:

- [README und Startanleitung](../README.md)
- [Mission](../specs/mission.md)
- [Technische Verfassung](../specs/tech-stack.md)
- [Phase 0](../specs/00-grundlagen.md)

**Phase 4:** Dashboard mit fünf Schnellzugriffen, Hinweisen beim leeren Start und
automatisch aktualisierten Kennzahlen. Der Speicherlink führt direkt zur Buchung
für Ausgabe und Rückgabe; Verwaltungsfehler lassen sich erneut laden.
Stunden-/Tagesmiete und Zusatzleistungen mit gespeicherter Preisaufteilung bleiben erhalten.
Spezifikation und Prüfnachweis: [Phase 4](../specs/04-dashboard-bedienung.md).
Phase 4 wurde am 21.10.2026 ausdrücklich abgenommen.

Start aus diesem Verzeichnis: `go run ./cmd/server`, optional mit `-demo`.
