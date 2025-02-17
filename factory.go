package recoder

import (
	"context"
	"io"
)

type Factory interface {
	io.Closer

	NewRecoder(context.Context) (Recoder, error)
	NewEncoder(context.Context, EncodersConfig) (Encoder, error)
	NewInputFromURL(context.Context, string, string, InputConfig) (Input, error)
	NewOutputFromURL(context.Context, string, string, OutputConfig) (Output, error)
}
