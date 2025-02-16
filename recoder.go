package recoder

import (
	"context"
	"io"
)

type Publisher interface {
	ClosedChan() <-chan struct{}
}

type Recoder interface {
	io.Closer

	NewEncoder(context.Context, EncodersConfig) (Encoder, error)
	NewInputFromURL(context.Context, string, string, InputConfig) (Input, error)
	NewOutputFromURL(context.Context, string, string, OutputConfig) (Output, error)
	StartRecoding(context.Context, Encoder, Input, Output) error
	WaitForRecodingEnd(context.Context) error
	GetStats(context.Context) (*Stats, error)
}

type NewInputFromPublisherer interface {
	NewInputFromPublisher(context.Context, Publisher, InputConfig) (Input, error)
}

type Factory interface {
	New(context.Context) (Recoder, error)
}
