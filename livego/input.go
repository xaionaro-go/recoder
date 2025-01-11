package livego

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/recoder"
)

type Input struct {
	URL string
}

var _ recoder.Input = (*Input)(nil)

func (r *Recoder) NewInputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.InputConfig,
) (recoder.Input, error) {
	if authKey != "" {
		return nil, fmt.Errorf("not implemented")
	}
	return &Input{
		URL: url,
	}, nil
}

func (r *Input) Close() error {
	return nil
}
