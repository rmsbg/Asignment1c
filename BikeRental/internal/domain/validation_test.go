package domain

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestPhotoValidation(t *testing.T) {
	var pngData, jpegData bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	_ = png.Encode(&pngData, img)
	_ = jpeg.Encode(&jpegData, img, nil)
	for _, tc := range []struct {
		name  string
		data  []byte
		mime  string
		valid bool
	}{
		{"png", pngData.Bytes(), "image/png", true}, {"jpeg", jpegData.Bytes(), "image/jpeg", true},
		{"wrong declaration", pngData.Bytes(), "image/jpeg", false}, {"damaged header", []byte("png"), "image/png", false},
		{"truncated", pngData.Bytes()[:len(pngData.Bytes())-15], "image/png", false},
		{"oversize", make([]byte, MaxPhotoBytes+1), "image/png", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidatePerson(Person{Name: "Test", Photo: tc.data, PhotoType: tc.mime})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t err=%v", tc.valid, err)
			}
		})
	}
	_, err := ValidatePerson(Person{Name: "Test", Photo: pngData.Bytes(), PhotoType: "image/png", DocumentNumber: "X"})
	if err == nil {
		t.Fatal("mixed proofs accepted")
	}
}
