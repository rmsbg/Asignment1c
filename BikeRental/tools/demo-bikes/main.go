// Command demo-bikes draws original, synthetic bicycle illustrations using only
// the Go standard library. Run from BikeRental: go run ./tools/demo-bikes.
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type canvas struct{ *image.RGBA }

func (c canvas) disc(x, y, r int, ink color.RGBA) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				c.SetRGBA(x+dx, y+dy, ink)
			}
		}
	}
}

func (c canvas) line(x, y, xx, yy, width int, ink color.RGBA) {
	n := max(abs(xx-x), abs(yy-y))
	for i := 0; i <= n; i++ {
		c.disc(x+(xx-x)*i/max(n, 1), y+(yy-y)*i/max(n, 1), width/2, ink)
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func main() {
	if err := os.MkdirAll("web/assets/demo", 0755); err != nil {
		panic(err)
	}
	for i, kind := range []string{"Downhill", "MTB", "Rennrad"} {
		c := canvas{image.NewRGBA(image.Rect(0, 0, 800, 450))}
		ink := color.RGBA{35, 54, 52, 255}
		frame := []color.RGBA{{198, 95, 59, 255}, {60, 114, 91, 255}, {66, 106, 148, 255}}[i]
		for y := 0; y < 450; y++ {
			for x := 0; x < 800; x++ {
				c.SetRGBA(x, y, color.RGBA{uint8(246 - y/60), uint8(243 - y/60), uint8(230 - y/60), 255})
			}
		}
		c.line(90, 363, 710, 363, 3, color.RGBA{198, 204, 185, 255})
		for _, x := range []int{225, 575} {
			r := 88
			c.disc(x, 270, r, ink)
			c.disc(x, 270, r-[]int{14, 11, 6}[i], color.RGBA{220, 225, 210, 255})
			for n := range 12 {
				a := float64(n) * math.Pi / 6
				c.line(x, 270, x+int(float64(r-17)*math.Cos(a)), 270+int(float64(r-17)*math.Sin(a)), 1, color.RGBA{143, 158, 146, 255})
			}
			c.disc(x, 270, 7, ink)
		}
		c.line(225, 270, 350, 270, 9, frame)
		c.line(350, 270, 295, 155, 12, frame)
		c.line(295, 155, 225, 270, 9, frame)
		c.line(295, 155, 510, 145, 11, frame)
		c.line(510, 145, 350, 270, 14, frame)
		c.line(510, 145, 575, 270, 10, ink)
		c.line(295, 155, 285, 130, 7, ink)
		c.line(260, 127, 310, 127, 12, ink)
		c.line(510, 145, 500, 115, 7, ink)
		if i == 2 {
			c.line(500, 115, 543, 115, 7, ink)
			c.line(543, 115, 553, 130, 7, ink)
			c.line(553, 130, 542, 142, 7, ink)
		} else {
			c.line(477, 109, 535, 106, 8, ink)
			c.line(533, 186, 552, 221, 16, frame)
		}
		if i == 0 {
			c.line(340, 180, 390, 205, 14, ink)
		}
		c.disc(350, 270, 19, ink)
		c.disc(350, 270, 12, color.RGBA{180, 191, 174, 255})
		c.line(350, 270, 376, 292, 6, ink)
		c.line(364, 293, 393, 293, 6, ink)
		f, err := os.Create(filepath.Join("web/assets/demo", kind+".png"))
		if err != nil {
			panic(err)
		}
		if err := png.Encode(f, c); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}
