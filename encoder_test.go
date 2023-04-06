package gojxl_test

import (
	"hash/crc32"
	"image"
	"image/color"
	"io"
	"os"
	"testing"
    "bytes"
    _ "image/png"

	jxl "github.com/jlortiz0/go-jxl-decoder"
)

const EncodeSingleImageName = "tests/input.png"
const EncodeSingleImageCRC = 0x7CFC11DC
const EncodeVideoCRC = 0xC5061331

func TestEncode(t *testing.T) {
	f, err := os.Open(EncodeSingleImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
    buf := new(bytes.Buffer)
	err = jxl.Encode(buf, i)
	if err != nil {
		t.Fatal(err)
	}
    i, err = jxl.Decode(buf)
    i2 := i.(*image.NRGBA)
    h := crc32.ChecksumIEEE(i2.Pix)
	if h != EncodeSingleImageCRC {
		t.Error("crc does not match", EncodeSingleImageCRC, h)
	}
}

func TestEncoderBadBuffer(t *testing.T) {
	e := jxl.NewJxlEncoder(io.Discard)
	e.SetInfo(256, 256, color.RGBAModel, 0)
	err := e.Write(make([]byte, 128))
	if err != jxl.EncodeInputError {
		t.Error("expected EncodeInputError, got", err)
	}
	e.Destroy()
}

func TestEncoderDoubleWriteOneImage(t *testing.T) {
	f, err := os.Open(EncodeSingleImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	i2 := i.(*image.RGBA)
	e := jxl.NewJxlEncoder(io.Discard)
	defer e.Destroy()
	e.SetInfo(i.Bounds().Dx(), i.Bounds().Dy(), color.RGBAModel, 0)
	err = e.Write(i2.Pix)
	if err != nil {
		t.Fatal(err)
	}
	err = e.Write(i2.Pix)
	if err != jxl.EncodeClosedError {
		t.Error("expected EncodeClosedError, got", err)
	}
}

func TestEncoderVideo(t *testing.T) {
	f, err := os.Open(EncodeSingleImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	i2 := i.(*image.RGBA)
	buf := new(bytes.Buffer)
	e := jxl.NewJxlEncoder(buf)
	e.SetInfo(i.Bounds().Dx(), i.Bounds().Dy(), color.RGBAModel, 0.5)
	err = e.Write(i2.Pix)
	if err != nil {
        e.Destroy()
		t.Fatal(err)
	}
	flipped := make([]byte, len(i2.Pix))
	for i := 0; i < i2.Bounds().Dy(); i++ {
		offset := i2.PixOffset(0, i)
		offset2 := i2.PixOffset(0, i2.Bounds().Dy()-i-1)
		copy(flipped[offset2:], i2.Pix[offset:offset+i2.Stride])
	}
	e.NextIsLast()
	err = e.Write(flipped)
	if err != nil {
        e.Destroy()
		t.Fatal(err)
	}
    e.Destroy()
    d := jxl.NewJxlDecoder(buf)
    defer d.Destroy()
    d.Read()
    b, err := d.Read()
    if err != nil {
        t.Fatal(err)
    }
    h := crc32.ChecksumIEEE(b)
	if h != EncodeVideoCRC {
		t.Error("crc does not match", EncodeVideoCRC, h)
	}
}

func TestEncoderVideoWOFinalize(t *testing.T) {
	f, err := os.Open(EncodeSingleImageName)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	i, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	i2 := i.(*image.RGBA)
	buf := new(bytes.Buffer)
	e := jxl.NewJxlEncoder(buf)
	defer e.Destroy()
	e.SetInfo(i.Bounds().Dx(), i.Bounds().Dy(), color.RGBAModel, 0.5)
	err = e.Write(i2.Pix)
	if err != nil {
		t.Fatal(err)
	}
	flipped := make([]byte, len(i2.Pix))
	for i := 0; i < i2.Bounds().Dy(); i++ {
		offset := i2.PixOffset(0, i)
		offset2 := i2.PixOffset(0, i2.Bounds().Dy()-i-1)
		copy(flipped[offset2:], i2.Pix[offset:offset+i2.Stride])
	}
	err = e.Write(flipped)
	if err != nil {
		t.Fatal(err)
	}
    e.Destroy()
    d := jxl.NewJxlDecoder(buf)
    defer d.Destroy()
    d.Read()
    _, err = d.Read()
    if err != nil {
        t.Fatal(err)
    }
    n, err := d.Read()
    if err != nil {
        t.Fatal(err)
    }
    if n == nil {
        t.Error("expected black frame, got nil")
    }
}
