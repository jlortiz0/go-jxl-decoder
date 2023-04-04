package gojxl_test

import (
	"image"
	"image/png"
	"os"
	"testing"

	jxl "github.com/jlortiz0/go-jxl-decoder"
)

func TestDecodeConfig(t *testing.T) {
	f, err := os.Open("test.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	conf, err := jxl.DecodeConfig(f)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(conf)
}

func TestDecode(t *testing.T) {
	f, err := os.Open("test.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := jxl.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Create("test.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRegister(t *testing.T) {
	f, err := os.Open("test.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	conf, s, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatal(err)
	}
	if s != "jxl" {
		t.Error("expected jxl, got", s)
	}
	t.Log(conf)
}
