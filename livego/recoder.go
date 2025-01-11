package livego

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/xsync"
)

type Recoder struct {
	Locker  xsync.Mutex
	Encoder *Encoder
}

var _ recoder.Recoder = (*Recoder)(nil)

func (r *Recoder) Close() error {
	return fmt.Errorf("not implemented, yet")
}

func (r *Recoder) NewEncoder(
	ctx context.Context,
	cfg recoder.EncoderConfig,
) (recoder.Encoder, error) {
	return NewEncoder(cfg)
}

func (r *Recoder) StartRecoding(
	ctx context.Context,
	encoder recoder.Encoder,
	input recoder.Input,
	output recoder.Output,
) error {
	return xsync.DoR1(ctx, &r.Locker, func() error {
		if r.Encoder != nil {
			return fmt.Errorf("already recoding")
		}
		var ok bool
		r.Encoder, ok = encoder.(*Encoder)
		if !ok {
			return fmt.Errorf("expected an encoder of type %T, but received %T", r.Encoder, encoder)
		}
		return r.Encoder.StartRecoding(ctx, input, output)
	})
}

func (r *Recoder) WaitForRecodingEnd(
	ctx context.Context,
) error {
	encoder := xsync.DoR1(ctx, &r.Locker, func() *Encoder {
		return r.Encoder
	})
	if r.Encoder == nil {
		return fmt.Errorf("not recoding")
	}
	return encoder.WaitForRecordingEnd(ctx)
}

func (r *Recoder) GetStats(
	ctx context.Context,
) (*recoder.Stats, error) {
	return nil, fmt.Errorf("not implemented")
}
