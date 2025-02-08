package recoder

import (
	"context"
	"fmt"

	"github.com/asticode/go-astiav"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/recoder"
)

type DecoderHardware struct {
	codec                 *astiav.Codec
	codecContext          *astiav.CodecContext
	hardwareDeviceContext *astiav.HardwareDeviceContext
	hardwarePixelFormat   astiav.PixelFormat
}

func (d *DecoderHardware) Codec() *astiav.Codec {
	return d.codec
}

func (d *DecoderHardware) CodecContext() *astiav.CodecContext {
	return d.codecContext
}

func (d *DecoderHardware) Close() error {
	d.codecContext.Free()
	return nil
}

func (d *InputDecoder) newHardwareDecoder(
	ctx context.Context,
	_ *Input,
	stream *astiav.Stream,
) (_ret *decoderWithStream, _err error) {
	decoder, err := NewDecoderHardware(
		ctx,
		stream.CodecParameters(),
		d.HardwareDeviceType,
		d.Config.HardwareDeviceTypeName,
		d.Config.CodecName,
	)
	if err != nil {
		return nil, err
	}
	return &decoderWithStream{
		decoder: decoder,
		Stream:  stream,
	}, nil
}

func NewDecoderHardware(
	ctx context.Context,
	codecParameters *astiav.CodecParameters,
	hardwareDeviceType astiav.HardwareDeviceType,
	hardwareDeviceTypeName recoder.HardwareDeviceTypeName,
	codecName recoder.CodecName,
) (_ret *DecoderHardware, _err error) {
	if codecParameters.MediaType() != astiav.MediaTypeVideo {
		return nil, fmt.Errorf("currently hardware decoding is supported only for video streams")
	}

	decoder := &DecoderHardware{}
	defer func() {
		if _err != nil {
			_ = decoder.Close()
		}
	}()

	if codecName != "" {
		decoder.codec = astiav.FindDecoderByName(string(codecName))
	} else {
		decoder.codec = astiav.FindDecoder(codecParameters.CodecID())
	}
	if decoder.codec == nil {
		return nil, fmt.Errorf("unable to find a codec using codec ID %v", codecParameters.CodecID())
	}

	if decoder.codecContext = astiav.AllocCodecContext(decoder.codec); decoder.codecContext == nil {
		return nil, fmt.Errorf("unable to allocate codec context")
	}

	for _, p := range decoder.codec.HardwareConfigs() {
		if p.MethodFlags().Has(astiav.CodecHardwareConfigMethodFlagHwDeviceCtx) && p.HardwareDeviceType() == hardwareDeviceType {
			decoder.hardwarePixelFormat = p.PixelFormat()
			break
		}
	}

	if decoder.hardwarePixelFormat == astiav.PixelFormatNone {
		return nil, fmt.Errorf("hardware device type '%v' is not supported", hardwareDeviceType)
	}

	if err := codecParameters.ToCodecContext(decoder.codecContext); err != nil {
		return nil, fmt.Errorf("CodecParameters().ToCodecContext(...) returned error: %w", err)
	}

	var err error
	decoder.hardwareDeviceContext, err = astiav.CreateHardwareDeviceContext(
		hardwareDeviceType,
		string(hardwareDeviceTypeName),
		nil,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create hardware device context: %w", err)
	}

	decoder.codecContext.SetHardwareDeviceContext(decoder.hardwareDeviceContext)
	decoder.codecContext.SetPixelFormatCallback(func(pfs []astiav.PixelFormat) astiav.PixelFormat {
		for _, pf := range pfs {
			if pf == decoder.hardwarePixelFormat {
				return pf
			}
		}

		logger.Errorf(ctx, "unable to find appropriate pixel format")
		return astiav.PixelFormatNone
	})

	// TODO: figure out if we need to do GuessFrameRate here

	if err := decoder.codecContext.Open(decoder.codec, nil); err != nil {
		return nil, fmt.Errorf("unable to open codec context: %w", err)
	}

	return decoder, nil
}
