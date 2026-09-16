package rasterm

import (
	"io"

	sixel "github.com/mattn/go-sixel"
)

type SixelEncoder = *sixel.Encoder

func NewSixelEncoder(writer io.Writer) SixelEncoder {
	return sixel.NewEncoder(writer)
}
