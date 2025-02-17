package livego

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/recoder"
)

type Output struct {
	URL string
}

func NewOutputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	cfg recoder.OutputConfig,
) (*Output, error) {
	if authKey != "" {
		return nil, fmt.Errorf("not implemented")
	}
	return &Output{
		URL: url,
	}, nil
}

func (r *Output) Close() error {
	return nil
}
