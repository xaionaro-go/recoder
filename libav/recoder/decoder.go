package recoder

import (
	"github.com/asticode/go-astiav"
)

type decoder interface {
	CodecContext() *astiav.CodecContext
}
