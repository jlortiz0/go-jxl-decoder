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
const DecodeSingleImg16CRC = 0xBA789AE9
const DecodeSingleImgGCRC = 0x5A60E6F5
const DecodeSingleImgG16CRC = 0x7A9459F0
const DecodeVideoName = "tests/vid.jxl"
const DecodeVideoFirstCRC = 0x3E164F85
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

func TestDecode16(t *testing.T) {
	f, err := os.Open("tests/single16.jxl")
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
	if out.Sum32() != DecodeSingleImg16CRC {
		t.Error("crc does not match")
		t.Error(out.Sum32())
	}
}

func TestDecodeG(t *testing.T) {
	f, err := os.Open("tests/singleG.jxl")
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
	if out.Sum32() != DecodeSingleImgGCRC {
		t.Error("crc does not match")
	}
}

func TestDecodeG16(t *testing.T) {
	f, err := os.Open("tests/singleG16.jxl")
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
	if out.Sum32() != DecodeSingleImgG16CRC {
		t.Error("crc does not match")
		t.Error(out.Sum32())
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
