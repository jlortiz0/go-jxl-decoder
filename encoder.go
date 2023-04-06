package gojxl

import (
	"image"
	"image/color"
	"io"
	"math"
	"unsafe"
)

// #cgo LDFLAGS: -lm
// #include <jxl/encode.h>
// #include <jxl/codestream_header.h>
// #include <jxl/types.h>
// #include <jxl/resizable_parallel_runner.h>
// #include <stdint.h>
// #include <math.h>
// JxlEncoderStatus encoderProcess(JxlEncoder *p0, unsigned char *p1, size_t *p2) {
//     return JxlEncoderProcessOutput(p0, &p1, p2);
// }
import "C"

type EncodeError string

func (e EncodeError) Error() string { return "jxl encode error: " + string(e) }

const EncodeInfoError EncodeError = "failed to set info"
const EncodeUnsupportedError EncodeError = "image type not supported"
const EncodeClosedError EncodeError = "encoder is closed"
const EncodeUninitializedError EncodeError = "info not set before writing"
const EncodeInputError EncodeError = "failed to set input"
const EncodeDataError EncodeError = "unknown"

type JxlEncoder struct {
	encoder     *C.JxlEncoder
	runner      unsafe.Pointer
	settings    *C.JxlEncoderFrameSettings
	x, y        int
	pxFormat    C.JxlPixelFormat
	closed      bool
	shouldClose bool
	w           io.Writer
}

func NewJxlEncoder(w io.Writer) *JxlEncoder {
	e := new(JxlEncoder)
	runner, err := C.JxlResizableParallelRunnerCreate(nil)
	if runner == nil {
		panic(err)
	}
	e.runner = runner
	e2, err := C.JxlEncoderCreate(nil)
	if e2 == nil {
		panic(err)
	}
	C.JxlEncoderSetParallelRunner(e2, (*[0]byte)(C.JxlResizableParallelRunner), runner)
	e.encoder = e2
	e.w = w
	return e
}

func (e *JxlEncoder) Destroy() {
	if !e.closed {
		var fdata C.JxlFrameHeader
		C.JxlEncoderSetFrameHeader(e.settings, &fdata)
		buf := make([]byte, e.x*e.y*int(e.pxFormat.num_channels))
		e.shouldClose = true
		e.Write(buf)
	}
	C.JxlEncoderDestroy(e.encoder)
	C.JxlResizableParallelRunnerDestroy(e.runner)
}

func (e *JxlEncoder) NextIsLast() {
	e.shouldClose = true
}

func (e *JxlEncoder) SetInfo(x, y int, m color.Model, fps float64) bool {
	var info C.JxlBasicInfo
	C.JxlEncoderInitBasicInfo(&info)
	info.xsize = C.uint32_t(x)
	info.ysize = C.uint32_t(y)
	info.intensity_target = 255
	info.intrinsic_xsize = info.xsize
	info.intrinsic_ysize = info.ysize
	e.x, e.y = x, y
	switch m {
	case color.Gray16Model:
		info.bits_per_sample = 16
		fallthrough
	case color.GrayModel:
		info.alpha_bits = 0
		info.num_color_channels = 1
	case color.RGBA64Model:
		info.bits_per_sample = 16
		fallthrough
	case color.RGBAModel:
		info.alpha_bits = info.bits_per_sample
		info.num_extra_channels = 1
		info.alpha_premultiplied = C.JXL_TRUE
	case color.NRGBA64Model:
		info.bits_per_sample = 16
		fallthrough
	case color.NRGBAModel:
		info.alpha_bits = info.bits_per_sample
		info.num_extra_channels = 1
	}
	if fps > 0 {
		info.have_animation = C.JXL_TRUE
		var exp C.int
		fl := float64(C.frexp(C.double(fps), &exp))
		for fl != math.Floor(fl) {
			fl *= 2
			exp--
		}
		info.animation.tps_numerator = C.uint32_t(fl)
		info.animation.tps_denominator = 1
		if exp > 0 {
			info.animation.tps_numerator <<= exp
		} else {
			info.animation.tps_denominator <<= -exp
		}
		e.shouldClose = false
	} else {
		e.shouldClose = true
	}
	var pxFormat C.JxlPixelFormat
	pxFormat.num_channels = info.num_color_channels
	if info.alpha_bits != 0 {
		pxFormat.num_channels++
	}
	pxFormat.data_type = C.JXL_TYPE_UINT8
	if info.bits_per_sample == 16 {
		pxFormat.data_type = C.JXL_TYPE_UINT16
	}
	pxFormat.endianness = C.JXL_NATIVE_ENDIAN
	e.pxFormat = pxFormat
	ok := C.JxlEncoderSetBasicInfo(e.encoder, &info)
	if ok == C.JXL_ENC_SUCCESS {
		e.settings = C.JxlEncoderFrameSettingsCreate(e.encoder, nil)
		if e.settings == nil {
			return false
		}
		var bDepth C.JxlBitDepth
		bDepth.bits_per_sample = info.bits_per_sample
		bDepth._type = C.JXL_BIT_DEPTH_FROM_PIXEL_FORMAT
		ok = C.JxlEncoderSetFrameBitDepth(e.settings, &bDepth)
		if ok == C.JXL_ENC_SUCCESS && !e.shouldClose {
			var fdata C.JxlFrameHeader
			fdata.duration = 1
			ok = C.JxlEncoderSetFrameHeader(e.settings, &fdata)
		}
	}
	return ok == C.JXL_ENC_SUCCESS
}

func writeHelper(w io.Writer, b []byte) error {
	n := 0
	l := len(b)
	for n < l {
		n2, err := w.Write(b[n : l-n])
		if err != nil {
			return err
		}
		n += n2
	}
	return nil
}

func (e *JxlEncoder) Write(b []byte) error {
	if e.x == 0 {
		return EncodeUninitializedError
	}
	if e.closed {
		return EncodeClosedError
	}
	status := C.JxlEncoderAddImageFrame(e.settings, &e.pxFormat, unsafe.Pointer(&b[0]), C.size_t(len(b)))
	if status != C.JXL_ENC_SUCCESS {
		return EncodeInputError
	}
	if e.shouldClose {
		C.JxlEncoderCloseInput(e.encoder)
		e.closed = true
	}
	buf := make([]byte, block_size)
	sz := C.size_t(len(buf))
	status = C.encoderProcess(e.encoder, (*C.uchar)(unsafe.Pointer(&buf[0])), &sz)
	for status == C.JXL_ENC_NEED_MORE_OUTPUT {
		err := writeHelper(e.w, buf[:len(buf)-int(sz)])
		if err != nil {
			return err
		}
		sz = C.size_t(len(buf))
		status = C.encoderProcess(e.encoder, (*C.uchar)(unsafe.Pointer(&buf[0])), &sz)
	}
	if status == C.JXL_ENC_ERROR {
		return EncodeDataError
	}
	err := writeHelper(e.w, buf[:len(buf)-int(sz)])
	if err != nil {
		return err
	}
	return nil
}

func Encode(w io.Writer, img image.Image) error {
	var buf []uint8
	switch i := img.(type) {
	case *image.Gray:
		buf = i.Pix
	case *image.Gray16:
		buf = i.Pix
	case *image.NRGBA:
		buf = i.Pix
	case *image.NRGBA64:
		buf = i.Pix
	case *image.RGBA:
		buf = i.Pix
	case *image.RGBA64:
		buf = i.Pix
	default:
		return EncodeUnsupportedError
	}
	e := NewJxlEncoder(w)
	defer e.Destroy()
	rect := img.Bounds()
	if !e.SetInfo(rect.Dx(), rect.Dy(), img.ColorModel(), 0) {
		return EncodeInfoError
	}
	return e.Write(buf)
}
