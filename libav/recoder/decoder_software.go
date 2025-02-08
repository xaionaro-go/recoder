package recoder

import (
	"context"
	"fmt"

	"github.com/asticode/go-astiav"
)

type DecoderSoftware struct {
	codec        *astiav.Codec
	codecContext *astiav.CodecContext
}

func (d *DecoderSoftware) Codec() *astiav.Codec {
	return d.codec
}

func (d *DecoderSoftware) CodecContext() *astiav.CodecContext {
	return d.codecContext
}

func (d *DecoderSoftware) Close() error {
	d.codecContext.Free()
	return nil
}

type decoderWithStream struct {
	decoder
	Stream *astiav.Stream
}

func (d *InputDecoder) newSoftwareDecoder(
	ctx context.Context,
	input *Input,
	stream *astiav.Stream,
) (_ret *decoderWithStream, _err error) {
	var frameRate astiav.Rational
	if stream.CodecParameters().MediaType() == astiav.MediaTypeVideo {
		frameRate = input.FormatContext.GuessFrameRate(stream, nil)
	}
	decoder, err := NewDecoderSoftware(
		ctx,
		stream.CodecParameters(),
		frameRate,
	)
	if err != nil {
		return nil, err
	}
	return &decoderWithStream{
		decoder: decoder,
		Stream:  stream,
	}, nil
}

func NewDecoderSoftware(
	_ context.Context,
	codecParameters *astiav.CodecParameters,
	frameRate astiav.Rational,
) (_ret *DecoderSoftware, _err error) {
	decoder := &DecoderSoftware{}
	defer func() {
		if _err != nil {
			_ = decoder.Close()
		}
	}()

	decoder.codec = astiav.FindDecoder(codecParameters.CodecID())
	if decoder.codec == nil {
		return nil, fmt.Errorf("unable to find a codec using codec ID %v", codecParameters.CodecID())
	}

	decoder.codecContext = astiav.AllocCodecContext(decoder.codec)
	if decoder.codecContext == nil {
		return nil, fmt.Errorf("unable to allocate codec context")
	}

	err := codecParameters.ToCodecContext(decoder.codecContext)
	if err != nil {
		return nil, fmt.Errorf("codecParameters.ToCodecContext(...) returned error: %w", err)
	}

	if codecParameters.MediaType() == astiav.MediaTypeVideo {
		decoder.codecContext.SetFramerate(frameRate)
	}

	if err := decoder.codecContext.Open(decoder.codec, nil); err != nil {
		return nil, fmt.Errorf("unable to open codec context: %w", err)
	}

	return decoder, nil
}
