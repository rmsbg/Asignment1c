package domain

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
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
				t.Fatalf("person valid=%t err=%v", tc.valid, err)
			}
			_, err = ValidateBike(Bike{Name: "Testfahrrad", Type: "MTB", Photo: tc.data, PhotoType: tc.mime})
			if (err == nil) != tc.valid {
				t.Fatalf("bike valid=%t err=%v", tc.valid, err)
			}
		})
	}
	_, err := ValidatePerson(Person{Name: "Test", Photo: pngData.Bytes(), PhotoType: "image/png", DocumentNumber: "X"})
	if err == nil {
		t.Fatal("mixed proofs accepted")
	}
}

func TestBikePixelAndByteLimits(t *testing.T) {
	var data bytes.Buffer
	_ = png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	huge := bytes.Clone(data.Bytes())
	binary.BigEndian.PutUint32(huge[16:20], 5000)
	binary.BigEndian.PutUint32(huge[20:24], 5000)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	if err := ValidatePhoto(huge, "image/png"); err == nil {
		t.Fatal("25 million pixels accepted")
	}
	exact := append(bytes.Clone(data.Bytes()), make([]byte, MaxPhotoBytes-data.Len())...)
	if err := ValidatePhoto(exact, "image/png"); err != nil {
		t.Fatalf("exact 5 MiB boundary rejected: %v", err)
	}
	if err := ValidatePhoto(append(exact, 0), "image/png"); err == nil {
		t.Fatal("over 5 MiB accepted")
	}
}
