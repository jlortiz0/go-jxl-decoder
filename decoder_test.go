package gojxl_test

import (
	"encoding/binary"
	"image"
	"os"
	"testing"

	"github.com/devedge/imagehash"
	jxl "github.com/jlortiz0/go-jxl-decoder"
)

const DecodeSingleImgName = "tests/single.jxl"
const DecodeSingleImgHash uint64 = 0xe8c8989c9d98dcdb
const DecodeVideoName = "tests/vid.jxl"
const DecodeVideoFirstHash uint64 = 0xb269c9cccccc6830
const DecodeVideoLastHash uint64 = 0x5a192d2c2c1cf870

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
	h2, _ := imagehash.DhashHorizontal(img, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeSingleImgHash {
		t.Error("crc does not match", DecodeSingleImgHash, h)
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
	h2, _ := imagehash.DhashHorizontal(img, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeSingleImgHash {
		t.Error("crc does not match", DecodeSingleImgHash, h)
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
	h2, _ := imagehash.DhashHorizontal(img, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeSingleImgHash {
		t.Error("crc does not match", DecodeSingleImgHash, h)
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
	h2, _ := imagehash.DhashHorizontal(img, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeSingleImgHash {
		t.Error("crc does not match", DecodeSingleImgHash, h)
	}
}

func TestHitEnd(t *testing.T) {
	f, err := os.Open(DecodeSingleImgName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f)
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
	d := jxl.NewJxlDecoder(f)
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
	h2, _ := imagehash.DhashHorizontal(&image.NRGBA{Rect: image.Rect(0, 0, info.W, info.H), Pix: prev, Stride: info.W * 4}, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeVideoLastHash {
		t.Error("crc does not match", DecodeVideoLastHash, h)
	}
}

func TestRewind(t *testing.T) {
	f, err := os.Open(DecodeVideoName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f)
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
	h2, _ := imagehash.DhashHorizontal(&image.NRGBA{Rect: image.Rect(0, 0, info.W, info.H), Pix: n, Stride: info.W * 4}, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeVideoFirstHash {
		t.Error("crc does not match", DecodeVideoFirstHash, h)
	}
}

func TestReset(t *testing.T) {
	f, err := os.Open(DecodeVideoName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f)
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
	h2, _ := imagehash.DhashHorizontal(&image.NRGBA{Rect: image.Rect(0, 0, info.W, info.H), Pix: n, Stride: info.W * 4}, 8)
	h := binary.BigEndian.Uint64(h2)
	if h != DecodeSingleImgHash {
		t.Error("crc does not match", DecodeSingleImgHash, h)
	}
}
