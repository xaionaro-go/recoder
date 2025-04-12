package recoder

import (
	"encoding/json"
	"fmt"
	"strings"

	"maps"

	"gopkg.in/yaml.v3"
)

type DecodersConfig struct {
	InputAudioTracks []AudioTrackDecodingConfig
	InputVideoTracks []VideoTrackDecodingConfig
}

type EncodersConfig struct {
	OutputAudioTracks []AudioTrackEncodingConfig `json:"output_audio_tracks,omitempty" yaml:"output_audio_tracks,omitempty"`
	OutputVideoTracks []VideoTrackEncodingConfig `json:"output_video_tracks,omitempty" yaml:"output_video_tracks,omitempty"`
}

type VideoTrackDecodingConfig struct {
	DecodeVideoConfig
}

type DecodeVideoConfig struct {
	Codec         VideoCodec    `json:"codec,omitempty"   yaml:"codec,omitempty"`
	CustomOptions CustomOptions `json:"custom_options,omitempty" yaml:"custom_options,omitempty"`
}

func (cfg DecodeVideoConfig) GetCustomOptions() CustomOptions {
	return cfg.CustomOptions
}

type VideoTrackEncodingConfig struct {
	InputTrackIDs []int             `json:"input_track_ids,omitempty" yaml:"input_track_ids,omitempty"`
	Config        EncodeVideoConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type EncodeVideoConfig struct {
	Codec         VideoCodec    `json:"codec,omitempty"   yaml:"codec,omitempty"`
	Quality       VideoQuality  `json:"quality,omitempty" yaml:"quality,omitempty"`
	CustomOptions CustomOptions `json:"custom_options,omitempty" yaml:"custom_options,omitempty"`
	// TODO: add video filters
}

func (c *EncodeVideoConfig) UnmarshalJSON(b []byte) (_err error) {
	c.Quality = videoQualitySerializable{}
	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("unable to un-JSON-ize: %w", err)
	}
	if c.Quality != nil {
		c.Quality, err = c.Quality.(videoQualitySerializable).Convert()
		if err != nil {
			return fmt.Errorf("unable to convert the 'quality' field: %w", err)
		}
	}
	return nil
}

func (c *EncodeVideoConfig) UnmarshalYAML(b []byte) (_err error) {
	m := map[string]any{}
	err := yaml.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("unable to unmarshal EncodeVideoConfig bytes to a map: %w", err)
	}
	quality := m["quality"]
	m["quality"] = nil
	b, err = yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("unable to remarshal back to EncodeVideoConfig from the map: %w", err)
	}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("unable to un-JSON-ize: %w", err)
	}
	if quality != nil {
		sb, err := yaml.Marshal(quality)
		if err != nil {
			return fmt.Errorf("unable to remarshal back to EncodeVideoConfig from the map: %w", err)
		}
		s := videoQualitySerializable{}
		err = yaml.Unmarshal(sb, &s)
		if err != nil {
			return fmt.Errorf("unable to un-JSON-ize: %w", err)
		}
		c.Quality, err = s.Convert()
		if err != nil {
			return fmt.Errorf("unable to convert the 'quality' field: %w", err)
		}
	}
	return nil
}

func (c *EncodeVideoConfig) MarshalYAML() ([]byte, error) {
	cpy := *c
	if cpy.Quality != nil {
		cpy.Quality = cpy.Quality.serializable()
	}
	return yaml.Marshal(cpy)
}

type AudioTrackDecodingConfig struct {
	DecodeAudioConfig
}

type DecodeAudioConfig struct {
	Codec         AudioCodec    `json:"codec,omitempty"   yaml:"codec,omitempty"`
	CustomOptions CustomOptions `json:"custom_options,omitempty" yaml:"custom_options,omitempty"`
}

func (cfg DecodeAudioConfig) GetCustomOptions() CustomOptions {
	return cfg.CustomOptions
}

type AudioTrackEncodingConfig struct {
	InputTrackIDs []int             `json:"input_track_ids,omitempty" yaml:"input_track_ids,omitempty"`
	Config        EncodeAudioConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

type EncodeAudioConfig struct {
	Codec         AudioCodec    `json:"codec,omitempty"   yaml:"codec,omitempty"`
	Quality       AudioQuality  `json:"quality,omitempty" yaml:"quality,omitempty"`
	CustomOptions CustomOptions `json:"custom_options,omitempty" yaml:"custom_options,omitempty"`
	// TODO: add audio filters
}

func (c *EncodeAudioConfig) UnmarshalJSON(b []byte) (_err error) {
	c.Quality = audioQualitySerializable{}
	err := json.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("unable to un-JSON-ize: %w", err)
	}
	if c.Quality != nil {
		c.Quality, err = c.Quality.(audioQualitySerializable).Convert()
		if err != nil {
			return fmt.Errorf("unable to convert the 'quality' field: %w", err)
		}
	}
	return nil
}

func (c *EncodeAudioConfig) UnmarshalYAML(b []byte) (_err error) {
	m := map[string]any{}
	err := yaml.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("unable to unmarshal EncodeAudioConfig bytes to a map: %w", err)
	}
	quality := m["quality"]
	m["quality"] = nil
	b, err = yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("unable to remarshal back to EncodeAudioConfig from the map: %w", err)
	}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return fmt.Errorf("unable to un-JSON-ize: %w", err)
	}
	if quality != nil {
		sb, err := yaml.Marshal(quality)
		if err != nil {
			return fmt.Errorf("unable to remarshal back to EncodeAudioConfig from the map: %w", err)
		}
		s := audioQualitySerializable{}
		err = yaml.Unmarshal(sb, &s)
		if err != nil {
			return fmt.Errorf("unable to un-JSON-ize: %w", err)
		}
		c.Quality, err = s.Convert()
		if err != nil {
			return fmt.Errorf("unable to convert the 'quality' field: %w", err)
		}
	}
	return nil
}

func (c *EncodeAudioConfig) MarshalYAML() ([]byte, error) {
	cpy := *c
	if cpy.Quality != nil {
		cpy.Quality = cpy.Quality.serializable()
	}
	return yaml.Marshal(cpy)
}

type AudioQuality interface {
	audioQuality()
	typeName() string
	serializable() audioQualitySerializable
	setValues(vq audioQualitySerializable) error
}

type AudioQualityConstantBitrate uint

func (AudioQualityConstantBitrate) typeName() string {
	return "constant_bitrate"
}

func (AudioQualityConstantBitrate) audioQuality() {}

func (aq AudioQualityConstantBitrate) serializable() audioQualitySerializable {
	return map[string]any{
		"type":    aq.typeName(),
		"bitrate": uint(aq),
	}
}

func (aq AudioQualityConstantBitrate) MarshalJSON() ([]byte, error) {
	return json.Marshal(aq.serializable())
}

func (aq AudioQualityConstantBitrate) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(aq.serializable())
}

func (aq *AudioQualityConstantBitrate) setValues(in audioQualitySerializable) error {
	bitrateR := in["bitrate"]
	bitrate, ok := bitrateR.(int)
	if !ok {
		return fmt.Errorf("have not found int value using key 'bitrate' in %#+v, found %T, instead", in, bitrateR)
	}

	*aq = AudioQualityConstantBitrate(bitrate)
	return nil
}

type audioQualitySerializable map[string]any

func (audioQualitySerializable) audioQuality() {}

func (aq audioQualitySerializable) typeName() string {
	result, _ := aq["type"].(string)
	return result
}

func (aq audioQualitySerializable) serializable() audioQualitySerializable {
	return aq
}

func (aq audioQualitySerializable) setValues(in audioQualitySerializable) error {
	for k := range aq {
		delete(aq, k)
	}
	maps.Copy(aq, in)
	return nil
}

func (aq audioQualitySerializable) Convert() (AudioQuality, error) {
	typeName, ok := aq["type"].(string)
	if !ok {
		return nil, nil
	}

	var r AudioQuality
	for _, sample := range []AudioQuality{
		ptr(AudioQualityConstantBitrate(0)),
	} {
		if sample.typeName() == typeName {
			r = sample
			break
		}
	}
	if r == nil {
		return nil, fmt.Errorf("unknown type '%s'", typeName)
	}

	if err := r.setValues(aq); err != nil {
		return nil, fmt.Errorf("unable to convert the value (aq): %w", err)
	}
	return r, nil
}

type AudioCodec uint

const (
	AudioCodecUndefined = AudioCodec(iota)
	AudioCodecCopy
	AudioCodecAAC
	AudioCodecVorbis
	AudioCodecOpus
	EndOfAudioCodec
)

func (ac *AudioCodec) String() string {
	if ac == nil {
		return "null"
	}

	switch *ac {
	case AudioCodecUndefined:
		return "<undefined>"
	case AudioCodecCopy:
		return "<copy>"
	case AudioCodecAAC:
		return "aac"
	case AudioCodecVorbis:
		return "vorbis"
	case AudioCodecOpus:
		return "opus"
	}
	return fmt.Sprintf("unexpected_audio_codec_id_%d", uint(*ac))
}

func (ac AudioCodec) MarshalJSON() ([]byte, error) {
	return []byte(`"` + ac.String() + `"`), nil
}

func (ac *AudioCodec) UnmarshalJSON(b []byte) error {
	if ac == nil {
		return fmt.Errorf("AudioCodec is nil")
	}
	s := strings.ToLower(strings.Trim(string(b), `"`))
	for cmp := AudioCodecUndefined; cmp < EndOfAudioCodec; cmp++ {
		if cmp.String() == s {
			*ac = cmp
			return nil
		}
	}
	return fmt.Errorf("unknown value of the AudioCodec: '%s'", s)
}

type VideoQuality interface {
	videoQuality()
	typeName() string
	serializable() videoQualitySerializable
	setValues(vq videoQualitySerializable) error
}

type VideoQualityConstantBitrate uint

func (VideoQualityConstantBitrate) typeName() string {
	return "constant_bitrate"
}

func (VideoQualityConstantBitrate) videoQuality() {}

func (vq VideoQualityConstantBitrate) serializable() videoQualitySerializable {
	return videoQualitySerializable{
		"type":    vq.typeName(),
		"bitrate": uint(vq),
	}
}

func (vq VideoQualityConstantBitrate) MarshalJSON() ([]byte, error) {
	return json.Marshal(vq.serializable())
}

func (vq VideoQualityConstantBitrate) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(vq.serializable())
}

func (vq *VideoQualityConstantBitrate) setValues(in videoQualitySerializable) error {
	bitrate, ok := in["bitrate"].(int)
	if !ok {
		return fmt.Errorf("have not found int value using key 'bitrate' in %#+v", in)
	}

	*vq = VideoQualityConstantBitrate(bitrate)
	return nil
}

type VideoQualityConstantQuality uint8

func (VideoQualityConstantQuality) typeName() string {
	return "constant_quality"
}

func (VideoQualityConstantQuality) videoQuality() {}

func (vq VideoQualityConstantQuality) serializable() videoQualitySerializable {
	return videoQualitySerializable{
		"type":    vq.typeName(),
		"quality": uint(vq),
	}
}

func (vq VideoQualityConstantQuality) MarshalJSON() ([]byte, error) {
	return json.Marshal(vq.serializable())
}

func (vq VideoQualityConstantQuality) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(vq.serializable())
}

func (vq *VideoQualityConstantQuality) setValues(in videoQualitySerializable) error {
	bitrate, ok := in["quality"].(float64)
	if !ok {
		return fmt.Errorf("have not found float64 value using key 'quality' in %#+v", in)
	}

	*vq = VideoQualityConstantQuality(bitrate)
	return nil
}

type videoQualitySerializable map[string]any

func (videoQualitySerializable) videoQuality() {}

func (vq videoQualitySerializable) typeName() string {
	result, _ := vq["type"].(string)
	return result
}

func (vq videoQualitySerializable) serializable() videoQualitySerializable {
	return vq
}

func (vq videoQualitySerializable) setValues(in videoQualitySerializable) error {
	for k := range vq {
		delete(vq, k)
	}
	for k, v := range in {
		vq[k] = v
	}
	return nil
}

func (vq videoQualitySerializable) Convert() (VideoQuality, error) {
	typeName, ok := vq["type"].(string)
	if !ok {
		return nil, nil
	}

	var r VideoQuality
	for _, sample := range []VideoQuality{
		ptr(VideoQualityConstantBitrate(0)),
		ptr(VideoQualityConstantQuality(0)),
	} {
		if sample.typeName() == typeName {
			r = sample
			break
		}
	}
	if r == nil {
		return nil, fmt.Errorf("unknown type '%s'", typeName)
	}

	if err := r.setValues(vq); err != nil {
		return nil, fmt.Errorf("unable to convert the value (vq): %w", err)
	}
	return r, nil
}

func ptr[T any](in T) *T {
	return &in
}

type VideoCodec uint

const (
	VideoCodecUndefined = VideoCodec(iota)
	VideoCodecCopy
	VideoCodecH264
	VideoCodecHEVC
	VideoCodecAV1
	EndOfVideoCodec
)

func (vc *VideoCodec) String() string {
	if vc == nil {
		return "null"
	}

	switch *vc {
	case VideoCodecUndefined:
		return "<undefined>"
	case VideoCodecCopy:
		return "<copy>"
	case VideoCodecH264:
		return "h264"
	case VideoCodecHEVC:
		return "hevc"
	case VideoCodecAV1:
		return "av1"
	}
	return fmt.Sprintf("unexpected_video_codec_id_%d", uint(*vc))
}

func (vc VideoCodec) MarshalJSON() ([]byte, error) {
	return []byte(`"` + vc.String() + `"`), nil
}

func (vc *VideoCodec) UnmarshalJSON(b []byte) error {
	if vc == nil {
		return fmt.Errorf("VideoCodec is nil")
	}
	s := strings.ToLower(strings.Trim(string(b), `"`))
	for cmp := VideoCodecUndefined; cmp < EndOfVideoCodec; cmp++ {
		if cmp.String() == s {
			*vc = cmp
			return nil
		}
	}
	return fmt.Errorf("unknown value of the VideoCodec: '%s'", s)
}
