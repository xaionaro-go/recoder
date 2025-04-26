//go:build with_libav
// +build with_libav

package process

import (
	"context"
	"os/exec"

	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/libav/process/client"
)

const (
	debugDisableForking = true
)

type ContextID = client.ContextID

type InputID = client.InputID
type InputConfig = recoder.InputConfig

type OutputID = client.OutputID
type OutputConfig = recoder.OutputConfig

type RecoderStats = recoder.Stats

type Recoder struct {
	*client.Client
	Cmd *exec.Cmd
}

func Run(
	ctx context.Context,
) (*Recoder, error) {
	return run(ctx)
}
