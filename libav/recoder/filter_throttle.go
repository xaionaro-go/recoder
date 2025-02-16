package recoder

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/asticode/go-astiav"
	"github.com/facebookincubator/go-belt/tool/experimental/errmon"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/observability"
)

type FilterThrottle struct {
	AverageBitRate         uint64
	BitrateAveragingPeriod time.Duration

	locker                      sync.Mutex
	skippedVideoFrame           bool
	videoAveragerBufferConsumed int64
	prevEncodeTS                time.Time
	isClosed                    bool
	inputChan                   chan InputPacket
	outputChan                  chan OutputPacket
}

var _ Filter = (*FilterThrottle)(nil)

func NewFilterThrottle(
	ctx context.Context,
	averageBitRate uint64,
	bitrateAveragingPeriod time.Duration,
) (*FilterThrottle, error) {
	if averageBitRate != 0 && bitrateAveragingPeriod == 0 {
		bitrateAveragingPeriod = time.Second * 10
		logger.Warnf(ctx, "AveragingPeriod is not set, defaulting to %v", bitrateAveragingPeriod)
	}

	f := &FilterThrottle{
		AverageBitRate:         averageBitRate,
		BitrateAveragingPeriod: bitrateAveragingPeriod,
		inputChan:              make(chan InputPacket, 100),
		outputChan:             make(chan OutputPacket, 1),
	}

	observability.Go(ctx, func() {
		err := f.readerLoop(ctx)
		if err != nil {
			errmon.ObserveErrorCtx(ctx, err)
		}
	})
	return f, nil
}

func (f *FilterThrottle) readerLoop(
	ctx context.Context,
) error {
	return readerLoop(ctx, f.inputChan, f)
}

type sendPacketer interface {
	SendPacket(
		ctx context.Context,
		input InputPacket,
	) error
}

func (f *FilterThrottle) Close() error {
	f.locker.Lock()
	defer f.locker.Unlock()
	close(f.outputChan)
	f.isClosed = true
	return nil
}

func (f *FilterThrottle) SendPacketChan() chan<- InputPacket {
	return f.inputChan
}

func (f *FilterThrottle) SendPacket(
	ctx context.Context,
	input InputPacket,
) error {
	f.locker.Lock()
	defer f.locker.Unlock()
	if f.isClosed {
		return io.ErrClosedPipe
	}

	if f.AverageBitRate == 0 {
		f.videoAveragerBufferConsumed = 0
		f.outputChan <- OutputPacket{
			Packet: ClonePacketAsWritable(input.Packet),
		}
		return nil
	}

	now := time.Now()
	prevTS := f.prevEncodeTS
	f.prevEncodeTS = now

	tsDiff := now.Sub(prevTS)
	allowMoreBits := 1 + int64(tsDiff.Seconds()*float64(f.AverageBitRate))

	f.videoAveragerBufferConsumed -= allowMoreBits
	if f.videoAveragerBufferConsumed < 0 {
		f.videoAveragerBufferConsumed = 0
	}

	pktSize := input.Packet.Size()
	averagingBuffer := int64(f.BitrateAveragingPeriod.Seconds() * float64(f.AverageBitRate))
	consumedWithPacket := f.videoAveragerBufferConsumed + int64(pktSize)*8
	if consumedWithPacket > averagingBuffer {
		f.skippedVideoFrame = true
		logger.Tracef(ctx, "skipping a frame to reduce the bitrate: %d > %d", consumedWithPacket, averagingBuffer)
		return nil
	}

	if f.skippedVideoFrame {
		isKeyFrame := input.Packet.Flags().Has(astiav.PacketFlagKey)
		if !isKeyFrame {
			logger.Tracef(ctx, "skipping a non-key frame (BTW, the consumedWithPacket is %d/%d)", consumedWithPacket, averagingBuffer)
			return nil
		}
	}

	f.skippedVideoFrame = false
	f.videoAveragerBufferConsumed = consumedWithPacket
	f.outputChan <- OutputPacket{
		Packet: ClonePacketAsWritable(input.Packet),
	}
	return nil
}

func (f *FilterThrottle) OutputPacketsChan() <-chan OutputPacket {
	return f.outputChan
}
