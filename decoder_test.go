package gojxl_test

import (
	"hash/crc32"
	"image"
	"os"
	"testing"

	jxl "github.com/jlortiz0/go-jxl-decoder"
)

const DecodeSingleImgName = "tests/single.jxl"
const DecodeSingleImgCRC = 0xB91B2433
const DecodeSingleImg16CRC = 0xD390EA1C
const DecodeSingleImgGCRC = 0xBEF53CF7
const DecodeSingleImgG16CRC = 0x9F6D9AE9
const DecodeVideoName = "tests/vid.jxl"
const DecodeVideoFirstCRC = 0x59F24A92
const DecodeVideoLastCRC = 0x9719908E

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
    i := img.(*image.NRGBA)
    h := crc32.ChecksumIEEE(i.Pix) 
    if h != DecodeSingleImgCRC {
		t.Error("crc does not match", DecodeSingleImgCRC, h)
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
    i := img.(*image.NRGBA64)
    h := crc32.ChecksumIEEE(i.Pix) 
    if h != DecodeSingleImg16CRC {
		t.Error("crc does not match", DecodeSingleImg16CRC, h)
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
    i := img.(*image.Gray)
    h := crc32.ChecksumIEEE(i.Pix) 
    if h != DecodeSingleImgGCRC {
		t.Error("crc does not match", DecodeSingleImgGCRC, h)
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
    i := img.(*image.Gray16)
    h := crc32.ChecksumIEEE(i.Pix) 
    if h != DecodeSingleImgG16CRC {
		t.Error("crc does not match", DecodeSingleImgG16CRC, h)
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
	prev, err := d.Read()
	n := prev
	for n != nil {
		prev = n
		n, err = d.Read()
	}
	if err != nil {
		t.Fatal(err)
	}
    h := crc32.ChecksumIEEE(prev) 
    if h != DecodeVideoLastCRC {
		t.Error("crc does not match", DecodeVideoLastCRC, h)
	}
}

func TestRewind(t *testing.T) {
	f, err := os.Open(DecodeVideoName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := jxl.NewJxlDecoder(f)
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
    h := crc32.ChecksumIEEE(n) 
    if h != DecodeVideoFirstCRC {
		t.Error("crc does not match", DecodeVideoFirstCRC, h)
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
	n, err := d.Read()
	if err != nil {
		t.Fatal(err)
	}
    h := crc32.ChecksumIEEE(n) 
    if h != DecodeSingleImgCRC {
		t.Error("crc does not match", DecodeSingleImgCRC, h)
	}
}
