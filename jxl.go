package gojxl

import (
	"image"
	"image/color"
	"io"
	"unsafe"
)

// #cgo pkg-config: libjxl
// #cgo windows LDFLAGS: jxl.dll
// #include <jxl/decode.h>
// #include <jxl/codestream_header.h>
// #include <jxl/types.h>
import "C"

const jxlHeader = "\xff\x0a"
const block_size = 4096 * 4
const config_block_size = 2048

type FormatError string

func (e FormatError) Error() string { return "invalid JPEG-XL format: " + string(e) }

func init() {
	image.RegisterFormat("jxl", jxlHeader, Decode, DecodeConfig)
}

func Decode(r io.Reader) (image.Image, error) {
	data := make([]byte, block_size)
	n, err := io.ReadFull(r, data)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	d, err := C.JxlDecoderCreate(nil)
	if d == nil {
		return nil, err
	}
	defer C.JxlDecoderDestroy(d)
	C.JxlDecoderSubscribeEvents(d, C.JXL_DEC_BASIC_INFO|C.JXL_DEC_COLOR_ENCODING|C.JXL_DEC_FULL_IMAGE)
	C.JxlDecoderSetInput(d, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(n))
	status := C.JxlDecoderProcessInput(d)
	for status != C.JXL_DEC_BASIC_INFO {
		if status == C.JXL_DEC_NEED_MORE_INPUT {
			remain := int(C.JxlDecoderReleaseInput(d))
			if remain > 0 {
				copy(data, data[len(data)-remain:])
			}
			n, err = io.ReadFull(r, data[remain:])
			if err != nil && err != io.ErrUnexpectedEOF {
				return nil, err
			}
			n += remain
			C.JxlDecoderSetInput(d, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(n))
		} else if status == C.JXL_DEC_ERROR {
			return nil, FormatError("header error")
		}
		status = C.JxlDecoderProcessInput(d)
	}
	var info C.JxlBasicInfo
	C.JxlDecoderGetBasicInfo(d, &info)
	rect := image.Rectangle{Max: image.Point{X: int(info.xsize), Y: int(info.ysize)}}
	var img image.Image
	var fmt C.JxlPixelFormat
	fmt.endianness = C.JXL_NATIVE_ENDIAN
	var outbuf []byte
	if info.num_color_channels == 1 {
		fmt.num_channels = 1
		if info.bits_per_sample == 16 {
			fmt.data_type = C.JXL_TYPE_UINT16
			img2 := image.NewGray16(rect)
			outbuf = img2.Pix
			img = img2
		} else {
			fmt.data_type = C.JXL_TYPE_UINT8
			img2 := image.NewGray(rect)
			outbuf = img2.Pix
			img = img2
		}
	} else if info.num_color_channels == 3 {
		fmt.num_channels = 4
		if int(info.alpha_premultiplied) != 0 {
			if info.bits_per_sample == 16 {
				fmt.data_type = C.JXL_TYPE_UINT16
				img2 := image.NewRGBA64(rect)
				outbuf = img2.Pix
				img = img2
			} else {
				fmt.data_type = C.JXL_TYPE_UINT8
				img2 := image.NewRGBA(rect)
				outbuf = img2.Pix
				img = img2
			}
		} else {
			if info.bits_per_sample == 16 {
				fmt.data_type = C.JXL_TYPE_UINT16
				img2 := image.NewNRGBA64(rect)
				outbuf = img2.Pix
				img = img2
			} else {
				fmt.data_type = C.JXL_TYPE_UINT8
				img2 := image.NewNRGBA(rect)
				outbuf = img2.Pix
				img = img2
			}
		}
	} else {
		fmt.num_channels = 1
		if info.bits_per_sample == 16 {
			fmt.data_type = C.JXL_TYPE_UINT16
			img2 := image.NewAlpha16(rect)
			outbuf = img2.Pix
			img = img2
		} else {
			fmt.data_type = C.JXL_TYPE_UINT8
			img2 := image.NewAlpha(rect)
			outbuf = img2.Pix
			img = img2
		}
	}
	for status != C.JXL_DEC_SUCCESS && status != C.JXL_DEC_FULL_IMAGE {
		if status == C.JXL_DEC_NEED_IMAGE_OUT_BUFFER {
			status = C.JxlDecoderSetImageOutBuffer(d, &fmt, unsafe.Pointer(&outbuf[0]), C.size_t(len(outbuf)))
			if status == C.JXL_DEC_ERROR {
				return nil, FormatError("output buffer error")
			}
		} else if status == C.JXL_DEC_NEED_MORE_INPUT {
			remain := int(C.JxlDecoderReleaseInput(d))
			if remain > 0 {
				copy(data, data[len(data)-remain:])
			}
			n, err = io.ReadFull(r, data[remain:])
			if err != nil && err != io.ErrUnexpectedEOF {
				return nil, err
			}
			n += remain
			C.JxlDecoderSetInput(d, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(n))
		} else if status == C.JXL_DEC_ERROR {
			return nil, FormatError("decoding error")
		}
		status = C.JxlDecoderProcessInput(d)
	}
	return img, nil
}

func DecodeConfig(r io.Reader) (image.Config, error) {
	data := make([]byte, config_block_size)
	n, err := io.ReadFull(r, data)
	if err != nil && err != io.ErrUnexpectedEOF {
		return image.Config{}, err
	}
	d, err := C.JxlDecoderCreate(nil)
	if d == nil {
		return image.Config{}, err
	}
	defer C.JxlDecoderDestroy(d)
	C.JxlDecoderSetInput(d, (*C.uchar)(unsafe.Pointer(&data[0])), C.size_t(n))
	C.JxlDecoderCloseInput(d)
	status := C.JxlDecoderProcessInput(d)
	if status == C.JXL_DEC_ERROR {
		return image.Config{}, FormatError("header error")
	}
	var info C.JxlBasicInfo
	C.JxlDecoderGetBasicInfo(d, &info)
	cfg := image.Config{Width: int(info.xsize), Height: int(info.ysize)}
	if info.num_color_channels == 1 {
		if info.bits_per_sample == 16 {
			cfg.ColorModel = color.Gray16Model
		} else {
			cfg.ColorModel = color.GrayModel
		}
	} else if info.num_color_channels == 3 {
		if int(info.alpha_premultiplied) != 0 {
			if info.bits_per_sample == 16 {
				cfg.ColorModel = color.RGBA64Model
			} else {
				cfg.ColorModel = color.RGBAModel
			}
		} else {
			if info.bits_per_sample == 16 {
				cfg.ColorModel = color.NRGBA64Model
			} else {
				cfg.ColorModel = color.NRGBAModel
			}
		}
	} else {
		if info.bits_per_sample == 16 {
			cfg.ColorModel = color.Alpha16Model
		} else {
			cfg.ColorModel = color.AlphaModel
		}
	}
	return cfg, nil
}
