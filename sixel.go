package pritty

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

	sixel "github.com/mattn/go-sixel"
)

const (
	SIXEL_DA1_REQUEST  = "\x1b[c"
	SIXEL_DA1_RESPONSE = "\x1b[?"
)

func GetSixelSupport(stdin io.Reader, stdout io.Writer) (bool, error) {
	fmt.Fprint(stdout, SIXEL_DA1_REQUEST)
	err := awaitTerminalResponse(stdin, []byte(SIXEL_DA1_RESPONSE))
	if err != nil {
		return false, err
	}

	buffer, err := bufio.NewReader(stdin).ReadBytes('c')
	if err != nil {
		return false, err
	}

	return slices.Contains(strings.Split(strings.TrimSuffix(string(buffer), "c"), ";"), "4"), nil
}

type SixelEncoder = *sixel.Encoder

func NewSixelEncoder(writer io.Writer) SixelEncoder {
	return sixel.NewEncoder(writer)
}
