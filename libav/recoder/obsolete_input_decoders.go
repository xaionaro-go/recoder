package recoder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/asticode/go-astiav"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/recoder"
)

type DecodersConfig = recoder.DecodersConfig

type OBSOLETEInputDecoder struct {
	Locker        sync.Mutex
	Config        DecodersConfig
	InputFrame    *astiav.Frame
	SoftwareFrame *astiav.Frame
	Packet        *astiav.Packet

	Decoders map[int]*decoderWithStream
}

// OBSOLETE: use individual Encoder/Decoder-s instead of using this "smart" entangled mess.
func OBSOLETENewInputDecoder(cfg DecodersConfig) (*OBSOLETEInputDecoder, error) {
	result := &OBSOLETEInputDecoder{
		Config:        cfg,
		InputFrame:    astiav.AllocFrame(),
		SoftwareFrame: astiav.AllocFrame(),
		Packet:        astiav.AllocPacket(),
		Decoders:      make(map[int]*decoderWithStream),
	}
	return result, nil
}

func (d *OBSOLETEInputDecoder) Close() error {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	for _, s := range d.Decoders {
		s.Close()
	}
	d.SoftwareFrame.Free()
	d.InputFrame.Free()
	d.Packet.Free()
	return nil
}

func (d *OBSOLETEInputDecoder) lazyInitStreamDecoders(
	ctx context.Context,
	input *Input,
) (_err error) {

	for _, stream := range input.FormatContext.Streams() {
		if _, ok := d.Decoders[stream.Index()]; ok {
			logger.Debugf(ctx, "video codec for stream %d is already initialized, skipping", stream.Index())
			continue
		}

		var cfg interface{ GetCustomOptions() recoder.CustomOptions }
		switch stream.CodecParameters().MediaType() {
		case astiav.MediaTypeVideo:
			if stream.Index() < len(d.Config.InputVideoTracks) {
				cfg = d.Config.InputVideoTracks[stream.Index()]
			}
		case astiav.MediaTypeAudio:
			if stream.Index() < len(d.Config.InputAudioTracks) {
				cfg = d.Config.InputAudioTracks[stream.Index()]
			}
		default:
			logger.Debugf(ctx, "stream %d is not an audio/video stream, skipping", stream.Index())
			continue
		}
		if cfg != nil {
			logger.Debugf(ctx, "decoder config: %#+v", cfg)
		} else {
			logger.Debugf(ctx, "no decoder config")
		}

		hardwareDeviceType, _ := recoder.GetCustomOption[HardwareDeviceType](cfg.GetCustomOptions())
		hardwareDeviceName, _ := recoder.GetCustomOption[HardwareDeviceName](cfg.GetCustomOptions())
		videoDecoder, err := d.newDecoder(ctx, input, stream, hardwareDeviceType, hardwareDeviceName)
		if err != nil {
			return fmt.Errorf("unable to initialize a decoder for video stream #%d:%#+v, using config %#+v: %w", stream.Index(), stream, d.Config, err)
		}
		d.Decoders[stream.Index()] = videoDecoder

	}
	return nil
}

func (d *OBSOLETEInputDecoder) getStreamDecoder(
	ctx context.Context,
	input *Input,
	streamIndex int,
) (*decoderWithStream, error) {
	streamDecoder, ok := d.Decoders[d.Packet.StreamIndex()]
	if ok {
		return streamDecoder, nil
	}

	if err := d.lazyInitStreamDecoders(ctx, input); err != nil {
		return nil, fmt.Errorf("unable to initialize stream decoders: %w", err)
	}

	streamDecoder, ok = d.Decoders[d.Packet.StreamIndex()]
	if ok {
		return streamDecoder, nil
	}

	return nil, fmt.Errorf("internal error: stream decoder not found for stream #%d", streamIndex)
}

func (d *OBSOLETEInputDecoder) ReadFrame(
	ctx context.Context,
	input *Input,
	frameReader FrameReader,
) error {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	logger.Tracef(ctx, "ReadFrame")
	err := input.FormatContext.ReadFrame(d.Packet)
	logger.Tracef(ctx, "/ReadFrame: %v", err)
	if err != nil {
		if errors.Is(err, astiav.ErrEof) {
			return io.EOF
		}
		return fmt.Errorf("unable to read a frame: %w", err)
	}

	streamDecoder, err := d.getStreamDecoder(ctx, input, d.Packet.StreamIndex())
	if err != nil {
		return fmt.Errorf("unable to get a stream decoder: %w", err)
	}

	decoderCtx := streamDecoder.CodecContext()
	if err := decoderCtx.SendPacket(d.Packet); err != nil {
		return fmt.Errorf("unable to send packet to the decoder: %w", err)
	}

	frame := Frame{
		InputStream:        streamDecoder.Stream,
		InputFormatContext: input.FormatContext,
		Decoder:            streamDecoder.Decoder,
		Packet:             d.Packet,
		RAMFrame:           d.SoftwareFrame,
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := decoderCtx.ReceiveFrame(d.InputFrame)

		switch {
		case err == nil:
		case errors.Is(err, astiav.ErrEof):
			return io.EOF
		case errors.Is(err, astiav.ErrEagain):
			return nil
		default:
			return fmt.Errorf("unable to receive a frame: %w", err)
		}

		frame.Frame = d.InputFrame
		err = frameReader.ReadFrame(&frame)
		if err != nil {
			return fmt.Errorf("the FrameReader returned an error: %w", err)
		}
	}
}

type decoderWithStream struct {
	*Decoder
	Stream *astiav.Stream
}

func (d *OBSOLETEInputDecoder) newDecoder(
	ctx context.Context,
	input *Input,
	stream *astiav.Stream,
	hardwareDeviceType astiav.HardwareDeviceType,
	hardwareDeviceName HardwareDeviceName,
) (_ret *decoderWithStream, _err error) {
	if stream.CodecParameters().MediaType() == astiav.MediaTypeVideo {
		if stream.CodecParameters().FrameRate().Num() == 0 {
			stream.CodecParameters().SetFrameRate(input.FormatContext.GuessFrameRate(stream, nil))
		}
	}
	decoder, err := NewDecoder(
		ctx,
		"",
		stream.CodecParameters(),
		hardwareDeviceType,
		hardwareDeviceName,
		nil,
		0,
	)
	if err != nil {
		return nil, err
	}
	return &decoderWithStream{
		Decoder: decoder,
		Stream:  stream,
	}, nil
}
