package goconv

import (
	"fmt"

	"github.com/xaionaro-go/recoder"
	"github.com/xaionaro-go/recoder/libav/grpc/go/recoder_grpc"
)

func EncoderConfigFromThrift(
	cfg *recoder_grpc.EncoderConfig,
) (recoder.EncodersConfig, bool) {
	return recoder.EncodersConfig{
		OutputAudioTracks: OutputAudioTracksFromThrift(cfg.OutputAudioTracks),
		OutputVideoTracks: OutputVideoTracksFromThrift(cfg.OutputVideoTracks),
	}, cfg.Enable
}

func OutputAudioTracksFromThrift(
	tracks []*recoder_grpc.OutputAudioTrack,
) []recoder.AudioTrackEncodingConfig {
	return convertSlice(tracks, OutputAudioTrackFromThrift)
}

func OutputAudioTrackFromThrift(
	cfg *recoder_grpc.OutputAudioTrack,
) recoder.AudioTrackEncodingConfig {
	return recoder.AudioTrackEncodingConfig{
		InputTrackIDs: convertSlice(cfg.InputTrackIDs, func(in uint64) int { return int(in) }),
		Config:        EncodeAudioConfigFromThrift(cfg.Encode),
	}
}

func EncodeAudioConfigFromThrift(
	cfg *recoder_grpc.EncodeAudioConfig,
) recoder.EncodeAudioConfig {
	return recoder.EncodeAudioConfig{
		Codec:   AudioCodecFromThrift(cfg.Codec),
		Quality: AudioQualityFromThrift(cfg.Quality),
	}
}

func AudioCodecFromThrift(
	codec recoder_grpc.AudioCodec,
) recoder.AudioCodec {
	switch codec {
	case recoder_grpc.AudioCodec_AudioCodecAAC:
		return recoder.AudioCodecAAC
	case recoder_grpc.AudioCodec_AudioCodecVorbis:
		return recoder.AudioCodecVorbis
	case recoder_grpc.AudioCodec_AudioCodecOpus:
		return recoder.AudioCodecOpus
	default:
		panic(fmt.Errorf("unexpected codec: '%v'", codec))
	}
}

func AudioQualityFromThrift(
	q *recoder_grpc.AudioQuality,
) recoder.AudioQuality {

	switch q := q.GetAudioQuality().(type) {
	case *recoder_grpc.AudioQuality_ConstantBitrate:
		return ptr(recoder.AudioQualityConstantBitrate(q.ConstantBitrate))
	default:
		panic(fmt.Errorf("unexpected audio quality type: '%T' (%v)", q, q))
	}
}

func OutputVideoTracksFromThrift(
	tracks []*recoder_grpc.OutputVideoTrack,
) []recoder.VideoTrackEncodingConfig {
	return convertSlice(tracks, OutputVideoTrackFromThrift)
}

func OutputVideoTrackFromThrift(
	cfg *recoder_grpc.OutputVideoTrack,
) recoder.VideoTrackEncodingConfig {
	return recoder.VideoTrackEncodingConfig{
		InputTrackIDs: convertSlice(cfg.InputTrackIDs, func(in uint64) int { return int(in) }),
		Config:        EncodeVideoConfigFromThrift(cfg.Encode),
	}
}

func EncodeVideoConfigFromThrift(
	cfg *recoder_grpc.EncodeVideoConfig,
) recoder.EncodeVideoConfig {
	return recoder.EncodeVideoConfig{
		Codec:   VideoCodecFromThrift(cfg.Codec),
		Quality: VideoQualityFromThrift(cfg.Quality),
	}
}

func VideoCodecFromThrift(
	codec recoder_grpc.VideoCodec,
) recoder.VideoCodec {
	switch codec {
	case recoder_grpc.VideoCodec_VideoCodecH264:
		return recoder.VideoCodecH264
	case recoder_grpc.VideoCodec_VideoCodecHEVC:
		return recoder.VideoCodecHEVC
	case recoder_grpc.VideoCodec_VideoCodecAV1:
		return recoder.VideoCodecAV1
	default:
		panic(fmt.Errorf("unexpected codec: '%v'", codec))
	}
}

func VideoQualityFromThrift(
	q *recoder_grpc.VideoQuality,
) recoder.VideoQuality {
	switch q := q.GetVideoQuality().(type) {
	case *recoder_grpc.VideoQuality_ConstantBitrate:
		return ptr(recoder.VideoQualityConstantBitrate(q.ConstantBitrate))
	case *recoder_grpc.VideoQuality_ConstantQuality:
		return ptr(recoder.VideoQualityConstantQuality(q.ConstantQuality))
	default:
		panic(fmt.Errorf("unexpected video quality type: '%T' (%v)", q, q))
	}
}
