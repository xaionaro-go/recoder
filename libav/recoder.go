package libav

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/avpipeline/packet"
	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/libav/process"
)

type Packet = packet.Commons

type ContextID = process.ContextID

type Recoder struct {
	Context *Context
}

var _ recoder.Recoder = (*Recoder)(nil)

func (r *Recoder) Close() error {
	return nil
}

func (r *Recoder) Start(
	ctx context.Context,
	_ recoder.Encoder,
	input recoder.Input,
	output recoder.Output,
) error {
	err := r.Context.Process.processBackend.Client.StartRecoding(
		ctx,
		r.Context.ID,
		input.(*Input).ID,
		output.(*Output).ID,
	)
	if err != nil {
		return fmt.Errorf("got an error while starting the recording: %w", err)
	}

	return nil
}

func (r *Recoder) Wait(ctx context.Context) error {
	ch, err := r.Context.Process.Client.RecodingEndedChan(ctx, r.Context.ID)
	if err != nil {
		return err
	}
	<-ch
	return nil
}

func (r *Recoder) GetStats(
	ctx context.Context,
) (*recoder.Stats, error) {
	read, wrote, err := r.Context.Process.GetStats(ctx, r.Context.ID)
	if err != nil {
		return nil, err
	}
	return &recoder.Stats{
		BytesCountRead:  read,
		BytesCountWrote: wrote,
	}, nil
}
