package server

import (
	"runtime"

	"github.com/xaionaro-go/recoder"
)

func optimalVideoCodec(codec recoder.VideoCodec) string {
	switch codec {
	case recoder.VideoCodecH264:
		if runtime.GOOS == "android" {
			return "h264_mediacodec"
		} else {
			return "h264_nvenc"
		}
	default:
		return codec.String()
	}
}
