# Synthetische Testdaten

`ausweis-demo.png` ist ein erzeugtes geometrisches Farbmuster (400 × 240 Pixel,
PNG). Es enthält weder einen echten Ausweis noch eine Person, Nummer oder reale
Kundendaten. Das Muster dient ausschließlich der Upload- und Ablaufprüfung.

Für die Browservorführung im Personenformular „Foto hochladen“ wählen und diese
Datei auswählen. Die HTTP-MVP-Tests verwenden dasselbe PNG und erzeugen daraus
zusätzlich eine JPEG-Variante im Arbeitsspeicher. Testfälle legen eigene synthetische
Stammdaten an und benötigen keinen laufenden Server.
