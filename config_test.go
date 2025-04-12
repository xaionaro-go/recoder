package recoder

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestConfigMarshalUnmarshal(t *testing.T) {
	cfg := &EncodersConfig{
		OutputAudioTracks: []AudioTrackEncodingConfig{{
			InputTrackIDs: []int{1, 2, 3},
			Config: EncodeAudioConfig{
				Codec:   AudioCodecCopy,
				Quality: ptr(AudioQualityConstantBitrate(2)),
			},
		}},
		OutputVideoTracks: []VideoTrackEncodingConfig{{
			InputTrackIDs: []int{1, 2, 3},
			Config: EncodeVideoConfig{
				Codec:   VideoCodecCopy,
				Quality: ptr(VideoQualityConstantBitrate(2)),
			},
		}},
	}

	b, err := yaml.Marshal(cfg)
	require.NoError(t, err)

	defer func() {
		r := recover()
		if r != nil {
			require.Nil(t, r, string(b))
		}
	}()

	var cfgDup EncodersConfig
	err = yaml.Unmarshal(b, &cfgDup)
	require.NoError(t, err)

	require.Equal(t, cfg, &cfgDup)
}
