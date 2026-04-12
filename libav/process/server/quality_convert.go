// quality_convert.go provides conversions from recoder quality types to
// avpipeline quality types used by the encoder factory.

package server

import (
	"fmt"

	"github.com/xaionaro-go/avpipeline/quality"
	"github.com/xaionaro-go/recoder"
)

func videoQualityToCodecQuality(q recoder.VideoQuality) (quality.Quality, error) {
	switch q := q.(type) {
	case nil:
		return nil, nil
	case *recoder.VideoQualityConstantBitrate:
		return quality.ConstantBitrate(*q), nil
	case *recoder.VideoQualityConstantQuality:
		return quality.ConstantQuality(*q), nil
	default:
		return nil, fmt.Errorf("unsupported video quality type: %T", q)
	}
}

func audioQualityToCodecQuality(q recoder.AudioQuality) (quality.Quality, error) {
	switch q := q.(type) {
	case nil:
		return nil, nil
	case *recoder.AudioQualityConstantBitrate:
		return quality.ConstantBitrate(*q), nil
	default:
		return nil, fmt.Errorf("unsupported audio quality type: %T", q)
	}
}
