package pritty

import (
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"strconv"
	"strings"
)

const (
	ITERM_IMG_HDR = "\x1b]1337;"
	ITERM_IMG_FTR = "\a"
)

type ItermEncoder struct {
	writer    io.Writer
	streaming bool
}

func NewItermEncoder(writer io.Writer, streaming bool) ItermEncoder {
	return ItermEncoder{
		writer: writer,
	}
}

type ItermOpts struct {
	// Filename. Defaults to "Unnamed file".
	Name string

	// Width to render. See notes below.
	Width string

	// Height to render. See notes below.
	Height string

	// The width and height are given as a number followed by a unit, or the word "auto".
	//
	//   - N: N character cells.
	//   - Npx: N pixels.
	//   - N%: N percent of the session's width or height.
	//   - auto: The image's inherent size will be used to determine an appropriate dimension.

	// File size in bytes. Optional; this is only used by the progress indicator.
	Size int64

	// If set, the file will be displayed inline. Otherwise, it will be downloaded
	// with no visual representation in the terminal session.
	DisplayInline bool

	// If set, the image's inherent aspect ratio will not be respected.
	IgnoreAspectRatio bool
}

func (o ItermOpts) ToHeader(streaming bool) string {

	var opts []string

	if o.Name != "" {
		opts = append(opts, "name="+base64.StdEncoding.EncodeToString([]byte(o.Name)))
	}

	if o.Width != "" {
		opts = append(opts, "width="+o.Width)
	}

	if o.Height != "" {
		opts = append(opts, "height="+o.Height)
	}

	if o.Size > 0 {
		opts = append(opts, "size="+strconv.FormatInt(o.Size, 10))
	}

	// default: inline=0
	if o.DisplayInline {
		opts = append(opts, "inline=1")
	}

	// default: preserveAspectRatio=1
	if o.IgnoreAspectRatio {
		opts = append(opts, "preserveAspectRatio=0")
	}

	if streaming {
		return ITERM_IMG_HDR + "MultipartFile=" + strings.Join(opts, ";") + ITERM_IMG_FTR
	} else {
		return ITERM_IMG_HDR + "File=" + strings.Join(opts, ";") + ":"
	}
}

type ItermSupport = uint8

const (
	ItermSupported = 1 << iota
	ItermStreaming
)

// NOTE: uses $TERM_PROGRAM, which isn't passed through tmux or ssh
// checks if iterm inline image protocol is supported
func GetItermSupport() ItermSupport {

	V := GetEnvIdentifiers()

	if V["TERM"] == "mintty" {
		return ItermSupported
	}

	if V["LC_TERMINAL"] == "iterm2" {
		return ItermSupported | ItermStreaming
	}

	if V["TERM_PROGRAM"] == "wezterm" {
		return ItermSupported
	}

	if V["TERM_PROGRAM"] == "rio" {
		return ItermSupported
	}

	return 0
}

func (encoder ItermEncoder) Encode(image image.Image) error {
	return encoder.EncodeOpts(image, ItermOpts{DisplayInline: true})
}

func (encoder ItermEncoder) EncodeOpts(image image.Image, opts ItermOpts) error {
	writer, err := encoder.EncodeRaw(opts)
	if err != nil {
		return err
	}

	return errors.Join(
		png.Encode(writer, image),
		writer.Close(),
	)
}

func (encoder ItermEncoder) EncodeRaw(opts ItermOpts) (io.WriteCloser, error) {
	imageWriter, err := encoder.EncodeBase64(opts)
	if err != nil {
		return nil, err
	}

	base64Writer := base64.NewEncoder(base64.StdEncoding, imageWriter)

	return DualWriteCloser{
		writer: base64Writer,
		a:      base64Writer,
		b:      imageWriter,
	}, nil
}

func (encoder ItermEncoder) EncodeBase64(opts ItermOpts) (io.WriteCloser, error) {
	if _, err := fmt.Fprint(encoder.writer, opts.ToHeader(encoder.streaming)); err != nil {
		return nil, err
	}

	if encoder.streaming {
		return ItermStreamingWriter{
			chunkSize: 4096,
			writer:    encoder.writer,
		}, nil
	} else {
		return ItermStandardWriter{
			writer: encoder.writer,
		}, nil
	}
}
