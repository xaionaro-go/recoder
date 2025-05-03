//go:build !with_libav
// +build !with_libav

package process

import (
	"context"
	"fmt"

	"github.com/xaionaro-go/recoder"
)

type Recoder struct {
	*Client
}

func (r *Recoder) Kill(ctx context.Context) error {
	return fmt.Errorf("not compiled with libav support")
}

func (r *Recoder) Wait(ctx context.Context) error {
	return fmt.Errorf("not compiled with libav support")
}

func Run(
	ctx context.Context,
) (*Recoder, error) {
	return nil, fmt.Errorf("not compiled with libav support")
}

type Client struct{}

type ContextID uint64

type InputID uint64
type InputConfig = recoder.InputConfig

type OutputID uint64
type OutputConfig = recoder.OutputConfig

func (c *Client) NewInputFromURL(
	ctx context.Context,
	url string,
	authKey string,
	config InputConfig,
) (InputID, error) {
	return 0, fmt.Errorf("not compiled with libav support")
}

func (c *Client) NewOutputFromURL(
	ctx context.Context,
	url string,
	streamKey string,
	config OutputConfig,
) (OutputID, error) {
	return 0, fmt.Errorf("not compiled with libav support")
}

func (c *Client) StartRecoding(
	ctx context.Context,
	contextID ContextID,
	inputID InputID,
	outputID OutputID,
	config *recoder.EncodersConfig,
) error {
	return fmt.Errorf("not compiled with libav support")
}

func (c *Client) NewContext(
	ctx context.Context,
) (ContextID, error) {
	return 0, fmt.Errorf("not compiled with libav support")
}

func (c *Client) GetStats(
	ctx context.Context,
	contextID ContextID,
) (uint64, uint64, error) {
	return 0, 0, fmt.Errorf("not compiled with libav support")
}

func (c *Client) RecodingEndedChan(
	ctx context.Context,
	contextID ContextID,
) (<-chan struct{}, error) {
	return nil, fmt.Errorf("not compiled with libav support")
}

func (c *Client) CloseInput(
	ctx context.Context,
	inputID InputID,
) error {
	return fmt.Errorf("not compiled with libav support")
}

func (c *Client) CloseOutput(
	ctx context.Context,
	outputID OutputID,
) error {
	return fmt.Errorf("not compiled with libav support")
}

func (c *Client) CloseContext(
	ctx context.Context,
	contextID ContextID,
) error {
	return fmt.Errorf("not compiled with libav support")
}
