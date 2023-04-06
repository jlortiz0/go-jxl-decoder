package gojxl_test

import (
	"hash/crc32"
	"image"
	"image/png"
	"os"
	"testing"

	jxl "github.com/jlortiz0/go-jxl-decoder"
)

const DecodeSingleImgName = "tests/single.jxl"
const DecodeSingleImgCRC = 0xA8803D68
const DecodeVideoName = "tests/vid.jxl"
const DecodeVideoFirstCRC = 0x897C9D81
const DecodeVideoLastCRC = 0x28E5F171

func TestDecodeConfig(t *testing.T) {
	f, err := os.Open(DecodeSingleImgName)
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
	f, err := os.Open(DecodeSingleImgName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := jxl.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	out := crc32.NewIEEE()
	err = png.Encode(out, img)
	if err != nil {
		t.Fatal(err)
	}
	if out.Sum32() != DecodeSingleImgCRC {
		t.Error("crc does not match")
	}
}

func TestHitEnd(t *testing.T) {
	f, err := os.Open(DecodeSingleImgName)
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
	f, err := os.Open(DecodeSingleImgName)
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
	f, err := os.Open(DecodeVideoName)
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
	out := crc32.NewIEEE()
	err = png.Encode(out, &image.RGBA{Pix: prev, Stride: 4 * info.W, Rect: image.Rect(0, 0, info.W, info.H)})
	if err != nil {
		t.Fatal(err)
	}
	if out.Sum32() != DecodeVideoLastCRC {
		t.Error("crc does not match")
	}
}

func TestRewind(t *testing.T) {
	f, err := os.Open(DecodeVideoName)
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
	out := crc32.NewIEEE()
	err = png.Encode(out, &image.RGBA{Pix: n, Stride: 4 * info.W, Rect: image.Rect(0, 0, info.W, info.H)})
	if err != nil {
		t.Fatal(err)
	}
	if out.Sum32() != DecodeVideoFirstCRC {
		t.Error("crc does not match")
	}
}

func TestReset(t *testing.T) {
	f, err := os.Open(DecodeVideoName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f, 0)
	_, err = d.Read()
	if err != nil {
		t.Fatal(err)
	}
	f, err = os.Open(DecodeSingleImgName)
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
	out := crc32.NewIEEE()
	err = png.Encode(out, &image.RGBA{Pix: n, Stride: 4 * info.W, Rect: image.Rect(0, 0, info.W, info.H)})
	if err != nil {
		t.Fatal(err)
	}
	if out.Sum32() != DecodeSingleImgCRC {
		t.Error("crc does not match")
	}
}
