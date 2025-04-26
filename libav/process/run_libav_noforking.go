//go:build with_libav && noforking
// +build with_libav,noforking

package process

import (
	"context"
	"fmt"
	"net"

	"github.com/xaionaro-go/observability"
	"github.com/xaionaro-go/recoder/libav/process/client"
)

func run(
	ctx context.Context,
) (*Recoder, error) {
	errCh := make(chan error, 1)
	addrCh := make(chan net.Addr, 1)
	observability.Go(ctx, func() {
		defer close(errCh)
		defer close(addrCh)
		err := runEncoderServer(ctx, func(ctx context.Context, addr net.Addr) error {
			addrCh <- addr
			return nil
		})
		if err != nil {
			errCh <- err
		}
	})

	select {
	case addr := <-addrCh:
		c := client.New(addr.String())
		level := observability.LogLevelFilter.GetLevel()
		if err := c.SetLoggingLevel(ctx, level); err != nil {
			return nil, fmt.Errorf("unable to set the logging level to %s: %w", level, err)
		}
		return &Recoder{
			Client: c,
		}, nil
	case err := <-errCh:
		return nil, fmt.Errorf("unable to initialize a server: %w", err)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *Recoder) Kill(ctx context.Context) error {
	return r.Client.Die(ctx)
}

func (r *Recoder) Wait(ctx context.Context) error {
	return nil
}
