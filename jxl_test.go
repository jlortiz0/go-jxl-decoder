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

func TestHitEnd(t *testing.T) {
	f, err := os.Open("test.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f, 0)
	_, err = d.Read()
	if err != nil {
		t.Fatal(err)
	}
	out, err := d.Read()
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Error("expected nil, got buffer")
	}
	out, err = d.Read()
	if err != nil {
		t.Fatal(err)
	}
	if out != nil {
		t.Error("expected nil, got buffer")
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

func TestDecodeVideo(t *testing.T) {
	f, err := os.Open("yeah.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f, 0)
	info, _ := d.Info()
	prev, err := d.Read()
	n := prev
	for n != nil {
		prev = n
		n, err = d.Read()
	}
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Create("yeah.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, &image.RGBA{Pix: prev, Stride: 4 * info.W, Rect: image.Rect(0, 0, info.W, info.H)})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRewind(t *testing.T) {
	f, err := os.Open("yeah.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f, 0)
	info, _ := d.Info()
	n, err := d.Read()
	for n != nil {
		n, err = d.Read()
	}
	if err != nil {
		t.Fatal(err)
	}
	f.Seek(0, 0)
	d.Rewind()
	n, err = d.Read()
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Create("yeah2.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, &image.RGBA{Pix: n, Stride: 4 * info.W, Rect: image.Rect(0, 0, info.W, info.H)})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReset(t *testing.T) {
	f, err := os.Open("yeah.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f, 0)
	_, err = d.Read()
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Open("test.jxl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d.Reset(f)
	info, _ := d.Info()
	n, err := d.Read()
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Create("test2.png")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	err = png.Encode(f, &image.RGBA{Pix: n, Stride: 4 * info.W, Rect: image.Rect(0, 0, info.W, info.H)})
	if err != nil {
		t.Fatal(err)
	}
}
