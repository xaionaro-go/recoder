package recoder

import (
	"context"
	"fmt"

	"github.com/asticode/go-astiav"
	"github.com/asticode/go-astikit"
	"github.com/facebookincubator/go-belt/tool/logger"
)

type Codec struct {
	codec                 *astiav.Codec
	codecContext          *astiav.CodecContext
	hardwareDeviceContext *astiav.HardwareDeviceContext
	hardwarePixelFormat   astiav.PixelFormat
	closer                astikit.Closer
}

func (c *Codec) Codec() *astiav.Codec {
	return c.codec
}

func (c *Codec) CodecContext() *astiav.CodecContext {
	return c.codecContext
}

func (c *Codec) HardwareDeviceContext() *astiav.HardwareDeviceContext {
	return c.hardwareDeviceContext
}

func (c *Codec) HardwarePixelFormat() astiav.PixelFormat {
	return c.hardwarePixelFormat
}

func (c *Codec) Close() error {
	return c.closer.Close()
}

func newCodec(
	ctx context.Context,
	codecName string,
	codecParameters *astiav.CodecParameters,
	isEncoder bool, // otherwise: decoder
	hardwareDeviceType astiav.HardwareDeviceType,
	hardwareDeviceName HardwareDeviceName,
	options *astiav.Dictionary,
	flags int,
) (_ret *Codec, _err error) {
	c := &Codec{}
	defer func() {
		if _err != nil {
			_ = c.Close()
		}
	}()

	if isEncoder {
		if codecName != "" {
			c.codec = astiav.FindEncoderByName(string(codecName))
		} else {
			c.codec = astiav.FindEncoder(codecParameters.CodecID())
		}
	} else {
		if codecName != "" {
			c.codec = astiav.FindDecoderByName(string(codecName))
		} else {
			c.codec = astiav.FindDecoder(codecParameters.CodecID())
		}
	}
	if c.codec == nil {
		return nil, fmt.Errorf("unable to find a codec using name '%s' or codec ID %v", codecName, codecParameters.CodecID())
	}

	c.codecContext = astiav.AllocCodecContext(c.codec)
	if c.codecContext == nil {
		return nil, fmt.Errorf("unable to allocate codec context")
	}
	c.closer.Add(c.codecContext.Free)

	if hardwareDeviceType != astiav.HardwareDeviceTypeNone {
		if codecParameters.MediaType() != astiav.MediaTypeVideo {
			return nil, fmt.Errorf("currently hardware encoding/decoding is supported only for video streams")
		}

		for _, p := range c.codec.HardwareConfigs() {
			if p.MethodFlags().Has(astiav.CodecHardwareConfigMethodFlagHwDeviceCtx) && p.HardwareDeviceType() == hardwareDeviceType {
				c.hardwarePixelFormat = p.PixelFormat()
				break
			}
		}

		if c.hardwarePixelFormat == astiav.PixelFormatNone {
			return nil, fmt.Errorf("hardware device type '%v' is not supported", hardwareDeviceType)
		}
	}

	err := codecParameters.ToCodecContext(c.codecContext)
	if err != nil {
		return nil, fmt.Errorf("codecParameters.ToCodecContext(...) returned error: %w", err)
	}

	if codecParameters.MediaType() == astiav.MediaTypeVideo {
		if frameRate := codecParameters.FrameRate(); frameRate.Num() != 0 {
			c.codecContext.SetFramerate(frameRate)
		}
	}

	if hardwareDeviceType != astiav.HardwareDeviceTypeNone {
		c.hardwareDeviceContext, err = astiav.CreateHardwareDeviceContext(
			hardwareDeviceType,
			string(hardwareDeviceName),
			options,
			flags,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create hardware device context: %w", err)
		}
		c.closer.Add(c.hardwareDeviceContext.Free)

		c.codecContext.SetHardwareDeviceContext(c.hardwareDeviceContext)
		c.codecContext.SetPixelFormatCallback(func(pfs []astiav.PixelFormat) astiav.PixelFormat {
			for _, pf := range pfs {
				if pf == c.hardwarePixelFormat {
					return pf
				}
			}

			logger.Errorf(ctx, "unable to find appropriate pixel format")
			return astiav.PixelFormatNone
		})
	}

	if err := c.codecContext.Open(c.codec, options); err != nil {
		return nil, fmt.Errorf("unable to open codec context: %w", err)
	}

	return c, nil
}
