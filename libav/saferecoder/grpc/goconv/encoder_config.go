package goconv

import (
	"fmt"

	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/libav/saferecoder/grpc/go/recoder_grpc"
)

func EncoderConfigToThrift(
	cfg recoder.EncoderConfig,
) *recoder_grpc.EncoderConfig {
	return &recoder_grpc.EncoderConfig{
		OutputAudioTracks: OutputAudioTracksToThrift(cfg.OutputAudioTracks),
		OutputVideoTracks: OutputVideoTracksToThrift(cfg.OutputVideoTracks),
	}
}

func convertSlice[IN, OUT any](
	tracks []IN,
	convFunc func(IN) OUT,
) []OUT {
	result := make([]OUT, 0, len(tracks))
	for _, item := range tracks {
		result = append(result, convFunc(item))
	}
	return result
}

func OutputAudioTracksToThrift(
	tracks []recoder.AudioTrackConfig,
) []*recoder_grpc.OutputAudioTrack {
	return convertSlice(tracks, OutputAudioTrackToThrift)
}

func OutputAudioTrackToThrift(
	cfg recoder.AudioTrackConfig,
) *recoder_grpc.OutputAudioTrack {
	return &recoder_grpc.OutputAudioTrack{
		InputID:       uint64(cfg.InputID),
		InputTrackIDs: convertSlice(cfg.InputTrackIDs, func(in int) uint64 { return uint64(in) }),
		Encode:        EncodeAudioConfigToThrift(cfg.EncodeAudioConfig),
	}
}

func EncodeAudioConfigToThrift(
	cfg recoder.EncodeAudioConfig,
) *recoder_grpc.EncodeAudioConfig {
	return &recoder_grpc.EncodeAudioConfig{
		Codec:   AudioCodecToThrift(cfg.Codec),
		Quality: AudioQualityToThrift(cfg.Quality),
	}
}

func AudioCodecToThrift(
	codec recoder.AudioCodec,
) recoder_grpc.AudioCodec {
	switch codec {
	case recoder.AudioCodecAAC:
		return recoder_grpc.AudioCodec_AudioCodecAAC
	case recoder.AudioCodecVorbis:
		return recoder_grpc.AudioCodec_AudioCodecVorbis
	case recoder.AudioCodecOpus:
		return recoder_grpc.AudioCodec_AudioCodecOpus
	default:
		panic(fmt.Errorf("unexpected codec: '%v'", codec))
	}
}

func AudioQualityToThrift(
	q recoder.AudioQuality,
) *recoder_grpc.AudioQuality {
	switch q := q.(type) {
	case *recoder.AudioQualityConstantBitrate:
		return &recoder_grpc.AudioQuality{
			AudioQuality: &recoder_grpc.AudioQuality_ConstantBitrate{
				ConstantBitrate: uint32(*q),
			},
		}
	default:
		panic(fmt.Errorf("unexpected audio quality type: '%T' (%v)", q, q))
	}
}

func OutputVideoTracksToThrift(
	tracks []recoder.VideoTrackConfig,
) []*recoder_grpc.OutputVideoTrack {
	return convertSlice(tracks, OutputVideoTrackToThrift)
}

func OutputVideoTrackToThrift(
	cfg recoder.VideoTrackConfig,
) *recoder_grpc.OutputVideoTrack {
	return &recoder_grpc.OutputVideoTrack{
		InputID:       uint64(cfg.InputID),
		InputTrackIDs: convertSlice(cfg.InputTrackIDs, func(in int) uint64 { return uint64(in) }),
		Encode:        EncodeVideoConfigToThrift(cfg.EncodeVideoConfig),
	}
}

func EncodeVideoConfigToThrift(
	cfg recoder.EncodeVideoConfig,
) *recoder_grpc.EncodeVideoConfig {
	return &recoder_grpc.EncodeVideoConfig{
		Codec:   VideoCodecToThrift(cfg.Codec),
		Quality: VideoQualityToThrift(cfg.Quality),
	}
}

func VideoCodecToThrift(
	codec recoder.VideoCodec,
) recoder_grpc.VideoCodec {
	switch codec {
	case recoder.VideoCodecH264:
		return recoder_grpc.VideoCodec_VideoCodecH264
	case recoder.VideoCodecHEVC:
		return recoder_grpc.VideoCodec_VideoCodecHEVC
	case recoder.VideoCodecAV1:
		return recoder_grpc.VideoCodec_VideoCodecAV1
	default:
		panic(fmt.Errorf("unexpected codec: '%v'", codec))
	}
}

func VideoQualityToThrift(
	q recoder.VideoQuality,
) *recoder_grpc.VideoQuality {
	switch q := q.(type) {
	case *recoder.VideoQualityConstantBitrate:
		return &recoder_grpc.VideoQuality{
			VideoQuality: &recoder_grpc.VideoQuality_ConstantBitrate{
				ConstantBitrate: uint32(*q),
			},
		}
	case *recoder.VideoQualityConstantQuality:
		return &recoder_grpc.VideoQuality{
			VideoQuality: &recoder_grpc.VideoQuality_ConstantQuality{
				ConstantQuality: uint32(*q),
			},
		}
	default:
		panic(fmt.Errorf("unexpected video quality type: '%T' (%v)", q, q))
	}
}
