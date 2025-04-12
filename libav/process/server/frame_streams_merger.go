package server

import (
	"context"

	"github.com/asticode/go-astiav"
	"github.com/xaionaro-go/avpipeline/kernel"
	"github.com/xaionaro-go/avpipeline/types"
	"github.com/xaionaro-go/typing"
)

type frameStreamsMerger struct{}

var _ kernel.StreamIndexAssigner = (*frameStreamsMerger)(nil)

func (frameStreamsMerger) StreamIndexAssign(
	ctx context.Context,
	input types.InputPacketOrFrameUnion,
) (typing.Optional[int], error) {
	switch input.GetMediaType() {
	case astiav.MediaTypeVideo:
		return typing.Opt(1), nil
	case astiav.MediaTypeAudio:
		return typing.Opt(2), nil
	default:
		return typing.Opt(input.GetStreamIndex() + 2), nil
	}
}
