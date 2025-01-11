package xaionarogortmp

import (
	"github.com/xaionaro-go/recoder"
)

type RecoderFactory struct{}

var _ recoder.Factory = (*RecoderFactory)(nil)

func NewRecoderFactory() *RecoderFactory {
	return &RecoderFactory{}
}
