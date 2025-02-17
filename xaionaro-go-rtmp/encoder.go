package xaionarogortmp

import (
	"github.com/xaionaro-go/recoder"
)

type Encoder struct {
}

var _ recoder.Encoder = (*Encoder)(nil)

func NewEncoder(
	cfg recoder.EncodersConfig,
) (*Encoder, error) {
	return &Encoder{}, nil
}

func (*Encoder) Close() error {
	return nil
}
