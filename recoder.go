package recoder

import (
	"context"
)

type Publisher interface {
	ClosedChan() <-chan struct{}
}

type NewInputFromPublisherer interface {
	NewInputFromPublisher(context.Context, Publisher, InputConfig) (Input, error)
}

type Recoder interface {
	Start(context.Context, Encoder, Input, Output) error
	Wait(context.Context) error
	GetStats(context.Context) (*Stats, error)
}
