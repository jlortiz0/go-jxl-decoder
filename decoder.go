package gojxl

import (
	"image"
	"image/color"
	"io"
	"time"
	"unsafe"
)

// #cgo linux darwin pkg-config: libjxl libjxl_threads
// #cgo windows LDFLAGS: jxl.dll jxl_threads.dll
// #include <jxl/decode.h>
// #include <jxl/codestream_header.h>
// #include <jxl/types.h>
// #include <jxl/resizable_parallel_runner.h>
// #include <stdint.h>
import "C"

const jxlHeader = "\xff\x0a"
const block_size = 4096 * 4
const config_block_size = 2048

type DecodeError string

func (e DecodeError) Error() string { return "jxl decode error: " + string(e) }

const DecodeHeaderError DecodeError = "invalid header"
const DecodeInputError DecodeError = "unable to set input"
const DecodeDataError DecodeError = "invalid body"

func init() {
	image.RegisterFormat("jxl", jxlHeader, Decode, DecodeConfig)
}

type JxlDecoder struct {
	decoder *C.JxlDecoder
	runner  unsafe.Pointer
	buf     []byte
	r       io.Reader
	hasInfo bool
	hitEnd  bool
}

type JxlInfo struct {
	H, W               int
	BitDepth           int
	Channels           int
	Alpha              int
	AlphaPremult       bool
	Orientation        int
	PreviewH, PreviewW int
	Animated           bool
	FrameDelay         time.Duration
}

func NewJxlDecoder(r io.Reader, blockSize int) *JxlDecoder {
	if blockSize == 0 {
		blockSize = block_size
	}
	d := new(JxlDecoder)
	runner, err := C.JxlResizableParallelRunnerCreate(nil)
	if runner == nil {
		panic(err)
	}
	d.runner = runner
	d2, err := C.JxlDecoderCreate(nil)
	if d2 == nil {
		panic(err)
	}
	C.JxlDecoderSetParallelRunner(d2, (*[0]byte)(C.JxlResizableParallelRunner), runner)
	C.JxlDecoderSubscribeEvents(d2, C.JXL_DEC_BASIC_INFO|C.JXL_DEC_FULL_IMAGE)
	d.decoder = d2
	d.buf = make([]byte, blockSize)
	d.r = r
	return d
}

func (d *JxlDecoder) Destroy() {
	C.JxlDecoderDestroy(d.decoder)
	C.JxlResizableParallelRunnerDestroy(d.runner)
}

func (d *JxlDecoder) nextInput() error {
	remain := int(C.JxlDecoderReleaseInput(d.decoder))
	if remain > 0 {
		copy(d.buf, d.buf[len(d.buf)-remain:])
	}
	n, err := io.ReadFull(d.r, d.buf[remain:])
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}
	n += remain
	status := C.JxlDecoderSetInput(d.decoder, (*C.uchar)(unsafe.Pointer(&d.buf[0])), C.size_t(n))
	if status != C.JXL_DEC_SUCCESS {
		return DecodeInputError
	}
	return nil
}

func (d *JxlDecoder) Info() (JxlInfo, error) {
	for !d.hasInfo {
		err := d.nextInput()
		if err != nil {
			return JxlInfo{}, err
		}
		status := C.JxlDecoderProcessInput(d.decoder)
		if status == C.JXL_DEC_ERROR {
			return JxlInfo{}, DecodeHeaderError
		} else if status != C.JXL_DEC_NEED_MORE_INPUT {
			d.hasInfo = true
			break
		}
	}
	var info C.JxlBasicInfo
	C.JxlDecoderGetBasicInfo(d.decoder, &info)
	var output JxlInfo
	output.Alpha = int(info.alpha_bits)
	output.AlphaPremult = info.alpha_premultiplied != 0
	output.Animated = info.have_animation != 0
	output.BitDepth = int(info.bits_per_sample)
	output.Channels = int(info.num_color_channels)
	output.PreviewW = int(info.preview.xsize)
	output.PreviewH = int(info.preview.ysize)
	output.W, output.H = int(info.xsize), int(info.ysize)
	output.Orientation = int(info.orientation)
	if output.Animated {
		output.FrameDelay = time.Second / time.Duration(info.animation.tps_numerator) * time.Duration(info.animation.tps_denominator)
	}
	return output, nil
}

func (d *JxlDecoder) Read() ([]byte, error) {
	if d.hitEnd {
		return nil, nil
	}
	info, err := d.Info()
	if err != nil {
		return nil, err
	}
	sz := info.Channels
	if sz != 1 {
		sz += 1
	}
	var fmt C.JxlPixelFormat
	fmt.endianness = C.JXL_NATIVE_ENDIAN
	fmt.num_channels = C.uint32_t(sz)
	fmt.data_type = C.JXL_TYPE_UINT8
	if info.BitDepth == 16 {
		sz *= 2
		fmt.data_type = C.JXL_TYPE_UINT16
	}
	outbuf := make([]byte, sz*info.H*info.W)
	status := C.JxlDecoderSetImageOutBuffer(d.decoder, &fmt, unsafe.Pointer(&outbuf[0]), C.size_t(len(outbuf)))
	for status != C.JXL_DEC_SUCCESS {
		err = d.nextInput()
		if err != nil {
			return nil, err
		}
		status = C.JxlDecoderProcessInput(d.decoder)
		if status == C.JXL_DEC_NEED_IMAGE_OUT_BUFFER {
			status = C.JxlDecoderSetImageOutBuffer(d.decoder, &fmt, unsafe.Pointer(&outbuf[0]), C.size_t(len(outbuf)))
		}
	}
	status = C.JxlDecoderProcessInput(d.decoder)
	if status == C.JXL_DEC_SUCCESS {
		d.hitEnd = true
		return nil, nil
	}
	for status != C.JXL_DEC_SUCCESS && status != C.JXL_DEC_FULL_IMAGE {
		if status == C.JXL_DEC_NEED_MORE_INPUT {
			err = d.nextInput()
			if err != nil {
				return nil, err
			}
		} else if status == C.JXL_DEC_ERROR {
			return nil, DecodeDataError
		}
		status = C.JxlDecoderProcessInput(d.decoder)
	}
	return outbuf, nil
}

func (d *JxlDecoder) Reset(r io.Reader) {
	C.JxlDecoderReleaseInput(d.decoder)
	C.JxlDecoderReset(d.decoder)
	d.r = r
	d.hasInfo = false
	d.hitEnd = false
	C.JxlDecoderSetParallelRunner(d.decoder, (*[0]byte)(C.JxlResizableParallelRunner), d.runner)
	C.JxlDecoderSubscribeEvents(d.decoder, C.JXL_DEC_BASIC_INFO|C.JXL_DEC_FULL_IMAGE)
}

func (d *JxlDecoder) Rewind() {
	C.JxlDecoderReleaseInput(d.decoder)
	C.JxlDecoderRewind(d.decoder)
	d.hitEnd = false
	d.hasInfo = false
	C.JxlDecoderSubscribeEvents(d.decoder, C.JXL_DEC_BASIC_INFO|C.JXL_DEC_FULL_IMAGE)
}

func Decode(r io.Reader) (image.Image, error) {
	d := NewJxlDecoder(r, 0)
	defer d.Destroy()
	info, err := d.Info()
	if err != nil {
		return nil, err
	}
	buf, err := d.Read()
	if err != nil {
		return nil, err
	}
	rect := image.Rectangle{Max: image.Point{X: info.W, Y: info.H}}
	if info.Channels == 1 {
		if info.BitDepth == 16 {
			img := new(image.Gray16)
			img.Rect = rect
			img.Stride = 2 * info.W
			img.Pix = buf
			return img, nil
		} else {
			img := new(image.Gray)
			img.Rect = rect
			img.Stride = info.W
			img.Pix = buf
			return img, nil
		}
	} else if info.AlphaPremult {
		if info.BitDepth == 16 {
			img := new(image.RGBA64)
			img.Rect = rect
			img.Stride = 8 * info.W
			img.Pix = buf
			return img, nil
		} else {
			img := new(image.RGBA)
			img.Rect = rect
			img.Stride = 4 * info.W
			img.Pix = buf
			return img, nil
		}
	} else {
		if info.BitDepth == 16 {
			img := new(image.NRGBA64)
			img.Rect = rect
			img.Stride = 8 * info.W
			img.Pix = buf
			return img, nil
		} else {
			img := new(image.NRGBA)
			img.Rect = rect
			img.Stride = 4 * info.W
			img.Pix = buf
			return img, nil
		}
	}
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	d := NewJxlDecoder(r, config_block_size)
	defer d.Destroy()
	info, err := d.Info()
	if err != nil {
		return image.Config{}, err
	}
	cfg := image.Config{Width: info.W, Height: info.H}
	if info.Channels == 1 {
		if info.BitDepth == 16 {
			cfg.ColorModel = color.Gray16Model
		} else {
			cfg.ColorModel = color.GrayModel
		}
	} else if info.AlphaPremult {
		if info.BitDepth == 16 {
			cfg.ColorModel = color.RGBA64Model
		} else {
			cfg.ColorModel = color.RGBAModel
		}
	} else {
		if info.BitDepth == 16 {
			cfg.ColorModel = color.NRGBA64Model
		} else {
			cfg.ColorModel = color.NRGBAModel
		}
	}
	return cfg, nil
}
