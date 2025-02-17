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

func NewInputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.InputConfig,
) (*Input, error) {
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
