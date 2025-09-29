package server

import (
	"context"

	"github.com/asticode/go-astiav"
	"github.com/xaionaro-go/avpipeline/kernel"
	"github.com/xaionaro-go/avpipeline/packetorframe"
	"github.com/xaionaro-go/recoder"
)

type frameStreamsMerger struct {
	encodersConfig         recoder.EncodersConfig
	allowedVideoTrackIDMap map[int]int
	allowedAudioTrackIDMap map[int]int
}

var _ kernel.StreamIndexAssigner = (*frameStreamsMerger)(nil)

func newFrameStreamsMerger(cfg recoder.EncodersConfig) (*frameStreamsMerger, error) {
	m := &frameStreamsMerger{
		encodersConfig:         cfg,
		allowedVideoTrackIDMap: map[int]int{},
		allowedAudioTrackIDMap: map[int]int{},
	}
	for idx, t := range cfg.OutputVideoTracks {
		for _, id := range t.InputTrackIDs {
			m.allowedVideoTrackIDMap[id] = idx
		}
	}
	for idx, t := range cfg.OutputAudioTracks {
		for _, id := range t.InputTrackIDs {
			m.allowedAudioTrackIDMap[id] = idx
		}
	}
	return m, nil
}

func (m *frameStreamsMerger) StreamIndexAssign(
	ctx context.Context,
	input packetorframe.InputUnion,
) ([]int, error) {
	streamIndex := input.GetStreamIndex()
	switch input.GetMediaType() {
	case astiav.MediaTypeVideo:
		if len(m.encodersConfig.OutputVideoTracks) == 0 {
			return []int{}, nil
		}
		return []int{
			m.allowedVideoTrackIDMap[streamIndex],
		}, nil
	case astiav.MediaTypeAudio:
		if len(m.encodersConfig.OutputAudioTracks) == 0 {
			return []int{}, nil
		}
		return []int{
			len(m.allowedVideoTrackIDMap) + m.allowedAudioTrackIDMap[streamIndex],
		}, nil
	default:
		return []int{
			len(m.allowedVideoTrackIDMap) + len(m.allowedAudioTrackIDMap) + streamIndex,
		}, nil
	}
}
