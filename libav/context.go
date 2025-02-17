package libav

import (
	"context"
)

type Context struct {
	Process *Process
	ID      ContextID
}

func (p *Process) NewContext(
	ctx context.Context,
) (*Context, error) {
	recoderID, err := p.processBackend.Client.NewContext(ctx)
	if err != nil {
		return nil, err
	}
	return &Context{
		Process: p,
		ID:      ContextID(recoderID),
	}, nil
}

func (ctx *Context) Close() error {
	return ctx.Process.Client.CloseContext(context.Background(), ctx.ID)
}
