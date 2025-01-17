package libav

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/libav/saferecoder"
)

type RecoderFactory struct{}

var _ recoder.Factory = (*RecoderFactory)(nil)

func NewRecoderFactory() *RecoderFactory {
	return &RecoderFactory{}
}

func (RecoderFactory) New(
	ctx context.Context,
) (recoder.Recoder, error) {
	process, err := saferecoder.NewProcess(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize the process: %w", err)
	}

	recoderInstance, err := process.NewRecoder(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize the recoder: %w", err)
	}

	return &Recoder{
		Process: process,
		Recoder: recoderInstance,
	}, nil
}
