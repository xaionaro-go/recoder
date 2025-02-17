package livego

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/recoder"
)

type Factory struct{}

var _ recoder.Factory = (*Factory)(nil)

func NewFactory() *Factory {
	return &Factory{}
}

func (Factory) New(ctx context.Context) (recoder.Factory, error) {
	return &Factory{}, nil
}

func (Factory) Close() error {
	return fmt.Errorf("not implemented, yet")
}

func (Factory) NewEncoder(
	ctx context.Context,
	cfg recoder.EncodersConfig,
) (recoder.Encoder, error) {
	return &Encoder{}, nil
}

func (Factory) NewRecoder(
	ctx context.Context,
) (recoder.Recoder, error) {
	return NewRecoder(ctx), nil
}

func (Factory) NewInputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.InputConfig,
) (recoder.Input, error) {
	return NewInputFromURL(ctx, url, authKey, cfg)
}

func (Factory) NewOutputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.OutputConfig,
) (recoder.Output, error) {
	return NewOutputFromURL(ctx, url, authKey, cfg)
}
