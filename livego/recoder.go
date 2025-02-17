package livego

import (
	"context"
	"fmt"

	"github.com/gwuhaolin/livego/protocol/rtmp/rtmprelay"
	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/xsync"
)

type Recoder struct {
	Locker xsync.Mutex
	Relay  *rtmprelay.RtmpRelay
}

var _ recoder.Factory = (*Factory)(nil)

func NewRecoder(
	ctx context.Context,
) *Recoder {
	return &Recoder{}
}

func (r *Recoder) Close() error {
	ctx := context.TODO()
	return xsync.DoR1(ctx, &r.Locker, func() error {
		if r.Relay != nil {
			return fmt.Errorf("recoder is not started (or is closed)")
		}
		r.Relay.Stop()
		r.Relay = nil
		return nil
	})
}

func (r *Recoder) Start(
	ctx context.Context,
	_ recoder.Encoder,
	inputIface recoder.Input,
	outputIface recoder.Output,
) error {
	input, ok := inputIface.(*Input)
	if !ok {
		return fmt.Errorf("expected Input of type %T, but received %T", input, inputIface)
	}

	output, ok := outputIface.(*Output)
	if !ok {
		return fmt.Errorf("expected Input of type %T, but received %T", output, outputIface)
	}

	return xsync.DoR1(ctx, &r.Locker, func() error {
		relay := rtmprelay.NewRtmpRelay(&input.URL, &output.URL)
		if err := relay.Start(); err != nil {
			return fmt.Errorf(
				"unable to start RTMP relay from '%s' to '%s': %w",
				input.URL,
				output.URL,
				err,
			)
		}

		r.Relay = relay
		return nil
	})
}

func (r *Recoder) Wait(
	ctx context.Context,
) error {
	return xsync.DoR1(ctx, &r.Locker, func() error {
		if r.Relay != nil {
			return fmt.Errorf("recoder is not started (or is closed)")
		}
		panic("do not know how to implement this, without changing the upstream library, yet")
	})
}

func (r *Recoder) GetStats(
	ctx context.Context,
) (*recoder.Stats, error) {
	return &recoder.Stats{}, nil
}
