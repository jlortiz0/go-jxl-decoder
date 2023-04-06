package gojxl

// #include <jxl/encode.h>
// #include <jxl/codestream_header.h>
// #include <jxl/types.h>
// #include <jxl/resizable_parallel_runner.h>
// #include <stdint.h>
import "C"
import (
	"image"
	"image/color"
	"io"
	"unsafe"
)

type EncodeError string

func (e EncodeError) Error() string { return "jxl encode error: " + string(e) }

const EncodeInfoError DecodeError = "failed to set info"
const EncodeUnsupportedError DecodeError = "image type not supported"
const EncodeClosedError DecodeError = "encoder is closed"
const EncodeUninitializedError DecodeError = "info not set before writing"
const EncodeDataError DecodeError = "unknown"

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
		e.Write(buf)
	}
	C.JxlEncoderDestroy(e.encoder)
	C.JxlResizableParallelRunnerDestroy(e.runner)
}

func (e *JxlEncoder) NextIsLast() {
	e.shouldClose = true
}

func (e *JxlEncoder) SetInfo(x, y int, m color.Model, fps int) bool {
	var info C.JxlBasicInfo
	C.JxlEncoderInitBasicInfo(&info)
	info.xsize = C.uint32_t(x)
	info.ysize = C.uint32_t(y)
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
		info.alpha_premultiplied = C.JXL_TRUE
	case color.NRGBA64Model:
		info.bits_per_sample = 16
		fallthrough
	case color.NRGBAModel:
		info.alpha_bits = info.bits_per_sample
	}
	if fps != 0 {
		info.have_animation = C.JXL_TRUE
		info.animation.tps_numerator = 10000
		info.animation.tps_denominator = C.uint32_t(fps)
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
		var bDepth C.JxlBitDepth
		bDepth.bits_per_sample = info.bits_per_sample
		bDepth._type = C.JXL_BIT_DEPTH_FROM_PIXEL_FORMAT
		ok = C.JxlEncoderSetFrameBitDepth(e.settings, &bDepth)
	}
	return ok == C.JXL_ENC_SUCCESS
}

func (e *JxlEncoder) Write(b []byte) error {
	if e.x == 0 {
		return EncodeUninitializedError
	}
	if e.closed {
		return EncodeClosedError
	}
	C.JxlEncoderAddImageFrame(e.settings, &e.pxFormat, unsafe.Pointer(&b[0]), C.size_t(len(b)))
	if e.shouldClose {
		C.JxlEncoderCloseInput(e.encoder)
		e.closed = true
	}
	buf := make([]byte, block_size)
	sz := C.size_t(len(buf))
	bp := (*C.uchar)(unsafe.Pointer(&buf[0]))
	status := C.JxlEncoderProcessOutput(e.encoder, &bp, &sz)
	for status == C.JXL_ENC_NEED_MORE_OUTPUT {
		n := 0
		l := len(buf) - int(sz)
		for n < l {
			n2, err := e.w.Write(buf[n : l-n])
			if err != nil {
				return err
			}
			n += n2
		}
		bp = (*C.uchar)(unsafe.Pointer(&buf[0]))
		sz = C.size_t(len(buf))
		status = C.JxlEncoderProcessOutput(e.encoder, &bp, &sz)
	}
	if status == C.JXL_ENC_ERROR {
		return EncodeDataError
	}
	n := 0
	l := len(buf) - int(sz)
	for n < l {
		n2, err := e.w.Write(buf[n : l-n])
		if err != nil {
			return err
		}
		n += n2
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
