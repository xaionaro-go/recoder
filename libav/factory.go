package libav

import (
	"context"
	"errors"
	"fmt"

	"github.com/xaionaro-go/recoder"
)

type Factory struct {
	Process *Process
	Context *Context
}

var _ recoder.Factory = (*Factory)(nil)

func NewFactory(ctx context.Context) (*Factory, error) {
	process, err := NewProcess(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize the process: %w", err)
	}

	contextInstance, err := process.NewContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize the recoder: %w", err)
	}

	return &Factory{
		Process: process,
		Context: contextInstance,
	}, nil
}

func (f *Factory) Close() error {
	return errors.Join(
		f.Context.Close(),
		f.Process.Kill(),
	)
}

func (f *Factory) NewRecoder(
	ctx context.Context,
) (recoder.Recoder, error) {
	return &Recoder{
		Context: f.Context,
	}, nil
}

type Encoder struct{}

func (Encoder) Close() error {
	return nil
}

func (f *Factory) NewEncoder(
	ctx context.Context,
	cfg recoder.EncodersConfig,
) (recoder.Encoder, error) {
	return &Encoder{}, nil
}

func (f *Factory) NewInputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.InputConfig,
) (recoder.Input, error) {
	return f.Process.NewInputFromURL(ctx, url, authKey, cfg)
}

func (f *Factory) NewOutputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.OutputConfig,
) (recoder.Output, error) {
	return f.Process.NewOutputFromURL(ctx, url, authKey, cfg)
}
