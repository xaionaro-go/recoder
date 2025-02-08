package recoder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/asticode/go-astiav"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/recoder"
)

type DecoderConfig = recoder.DecoderConfig

type InputDecoder struct {
	Locker             sync.Mutex
	Config             DecoderConfig
	HardwareDeviceType astiav.HardwareDeviceType
	InputFrame         *astiav.Frame
	SoftwareFrame      *astiav.Frame
	Packet             *astiav.Packet

	HWDecoders map[int]*decoderWithStream
	SWDecoders map[int]*decoderWithStream
}

func NewInputDecoder(cfg DecoderConfig) (*InputDecoder, error) {
	result := &InputDecoder{
		Config:        cfg,
		InputFrame:    astiav.AllocFrame(),
		SoftwareFrame: astiav.AllocFrame(),
		Packet:        astiav.AllocPacket(),
		HWDecoders:    make(map[int]*decoderWithStream),
		SWDecoders:    make(map[int]*decoderWithStream),
	}

	hardwareDeviceType := astiav.FindHardwareDeviceTypeByName(string(cfg.HardwareDeviceTypeName))
	if hardwareDeviceType == astiav.HardwareDeviceTypeNone {
		logger.Errorf(context.TODO(), "the hardware device '%s' not found", cfg.HardwareDeviceTypeName)
	}
	result.HardwareDeviceType = hardwareDeviceType

	return result, nil
}

func (d *InputDecoder) Close() error {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	for _, s := range d.HWDecoders {
		s.CodecContext().Free()
	}
	d.SoftwareFrame.Free()
	d.InputFrame.Free()
	d.Packet.Free()
	return nil
}

func (d *InputDecoder) lazyInitStreamDecoders(
	ctx context.Context,
	input *Input,
) (_err error) {

	// Initialize video codecs
	for _, stream := range input.FormatContext.Streams() {
		switch stream.CodecParameters().MediaType() {
		case astiav.MediaTypeVideo:
			if _, ok := d.HWDecoders[stream.Index()]; ok {
				logger.Debugf(ctx, "video codec for stream %d is already initialized, skipping", stream.Index())
				continue
			}

			if d.HardwareDeviceType != astiav.HardwareDeviceTypeNone {
				videoDecoderHW, err := d.newHardwareDecoder(ctx, input, stream)
				if err == nil {
					d.HWDecoders[stream.Index()] = videoDecoderHW
					continue
				}
				logger.Warnf(ctx, "unable to initialize a hardware decoder for video stream #%d: %#+v: %v", stream.Index(), stream, err)
			}
			videoDecoderSW, err := d.newSoftwareDecoder(ctx, input, stream)
			if err != nil {
				return fmt.Errorf("unable to initialize a decoder for video stream #%d:%#+v, using config %#+v: %w", stream.Index(), stream, d.Config, err)
			}
			d.SWDecoders[stream.Index()] = videoDecoderSW
		case astiav.MediaTypeAudio:
			if _, ok := d.SWDecoders[stream.Index()]; ok {
				logger.Debugf(ctx, "audio codec for stream %d is already initialized, skipping", stream.Index())
				continue
			}

			audioDecoder, err := d.newSoftwareDecoder(ctx, input, stream)
			if err != nil {
				return fmt.Errorf("unable to initialize a decoder for audio stream #%d: %#+v: ", stream.Index(), stream)
			}
			d.SWDecoders[stream.Index()] = audioDecoder
		default:
			logger.Debugf(ctx, "stream %d is not an audio/video stream, skipping", stream.Index())
			continue
		}

	}
	return nil
}

func (d *InputDecoder) getStreamDecoder(
	ctx context.Context,
	input *Input,
	streamIndex int,
) (*decoderWithStream, error) {
	streamDecoderHW, ok := d.HWDecoders[d.Packet.StreamIndex()]
	if ok {
		return streamDecoderHW, nil
	}

	streamDecoderSW, ok := d.SWDecoders[d.Packet.StreamIndex()]
	if ok {
		return streamDecoderSW, nil
	}

	if err := d.lazyInitStreamDecoders(ctx, input); err != nil {
		return nil, fmt.Errorf("unable to initialize stream decoders: %w", err)
	}

	streamDecoderHW, ok = d.HWDecoders[d.Packet.StreamIndex()]
	if ok {
		return streamDecoderHW, nil
	}

	streamDecoderSW, ok = d.SWDecoders[d.Packet.StreamIndex()]
	if ok {
		return streamDecoderSW, nil
	}

	return nil, fmt.Errorf("internal error: stream decoder not found for stream #%d", streamIndex)
}

func (d *InputDecoder) ReadFrame(
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
		DecoderContext:     decoderCtx,
		Packet:             d.Packet,
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

		var resultingFrame *astiav.Frame
		switch streamDecoder := streamDecoder.decoder.(type) {
		case *DecoderHardware:
			// Get final frame
			if d.InputFrame.PixelFormat() == streamDecoder.hardwarePixelFormat {
				// Transfer hardware data
				if err := d.InputFrame.TransferHardwareData(d.SoftwareFrame); err != nil {
					log.Fatal(fmt.Errorf("main: transferring hardware data failed: %w", err))
				}

				// Update pts
				d.SoftwareFrame.SetPts(d.InputFrame.Pts())

				// Update final frame
				resultingFrame = d.SoftwareFrame
			}
		case *DecoderSoftware:
			resultingFrame = d.InputFrame
		}

		frame.Frame = resultingFrame
		err = frameReader.ReadFrame(&frame)
		if err != nil {
			return fmt.Errorf("the FrameReader returned an error: %w", err)
		}
	}
}
