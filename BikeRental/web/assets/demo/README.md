# Synthetische Fahrradillustrationen

`Downhill.png`, `MTB.png` und `Rennrad.png` sind eigens für dieses Projekt
programmatisch gezeichnete Fahrradillustrationen (800 × 450 Pixel, PNG).
Sie enthalten keine Personen, Kundendaten oder fremden Marken. Keine externen
Bildquellen oder Bilddienste werden verwendet.

Der vollständige Zeichencode liegt in `tools/demo-bikes/main.go`. Aus dem
Go-Modulverzeichnis `BikeRental/` reproduzierbar erzeugen:

```sh
go run ./tools/demo-bikes
```

Die Dateien werden über `go:embed` mitgeliefert und beim Demostart den drei
Fahrrädern zugeordnet. Hochgeladene Inventarbilder sind davon unabhängig und
bleiben ausschließlich im Arbeitsspeicher. Für HTTP-/Browserprüfungen können
diese synthetischen Dateien auch als Upload verwendet werden.
