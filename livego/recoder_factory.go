package livego

import (
	"context"

	"github.com/xaionaro-go/recoder"
)

type RecoderFactory struct{}

var _ recoder.Factory = (*RecoderFactory)(nil)

func NewRecoderFactory() *RecoderFactory {
	return &RecoderFactory{}
}

func (RecoderFactory) New(ctx context.Context) (recoder.Recoder, error) {
	return &Recoder{}, nil
}
