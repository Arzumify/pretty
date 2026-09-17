package pritty

import (
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"strings"
)

// See https://sw.kovidgoyal.net/kitty/graphics-protocol.html for more details.

const (
	KITTY_IMG_HDR = "\x1b_G"
	KITTY_IMG_FTR = "\x1b\\"
)

type KittyEncoder struct {
	writer io.Writer
}

func NewKittyEncoder(writer io.Writer) KittyEncoder {
	return KittyEncoder{
		writer: writer,
	}
}

type KittyImgOpts struct {
	SrcX        uint32 // x=
	SrcY        uint32 // y=
	SrcWidth    uint32 // w=
	SrcHeight   uint32 // h=
	CellOffsetX uint32 // X= (pixel x-offset inside terminal cell)
	CellOffsetY uint32 // Y= (pixel y-offset inside terminal cell)
	DstCols     uint32 // c= (display width in terminal columns)
	DstRows     uint32 // r= (display height in terminal rows)
	ZIndex      int32  // z=
	ImageId     uint32 // i=
	ImageNo     uint32 // I=
	PlacementId uint32 // p=
}

func (o KittyImgOpts) ToHeader(opts ...string) string {

	type fldmap struct {
		pv   *uint32
		code rune
	}
	sFld := []fldmap{
		{&o.SrcX, 'x'},
		{&o.SrcY, 'y'},
		{&o.SrcWidth, 'w'},
		{&o.SrcHeight, 'h'},
		{&o.CellOffsetX, 'X'},
		{&o.CellOffsetY, 'Y'},
		{&o.DstCols, 'c'},
		{&o.DstRows, 'r'},
		{&o.ImageId, 'i'},
		{&o.ImageNo, 'I'},
		{&o.PlacementId, 'p'},
	}

	for _, f := range sFld {
		if *f.pv != 0 {
			opts = append(opts, fmt.Sprintf("%c=%d", f.code, *f.pv))
		}
	}

	if o.ZIndex != 0 {
		opts = append(opts, fmt.Sprintf("z=%d", o.ZIndex))
	}

	return KITTY_IMG_HDR + strings.Join(opts, ",") + ";"
}

// checks if terminal supports kitty image protocols
func GetKittySupport() bool {

	// TODO: more rigorous check
	V := GetEnvIdentifiers()
	return (len(V["KITTY_WINDOW_ID"]) > 0) || (V["TERM_PROGRAM"] == "wezterm") || (V["TERM_PROGRAM"] == "ghostty")
}

func (encoder KittyEncoder) EncodeLocal(path string) error {
	return encoder.EncodeLocalOpts(path, KittyImgOpts{})
}

func (encoder KittyEncoder) EncodeLocalOpts(path string, opts KittyImgOpts) error {
	writer, err := encoder.EncodeLocalRaw(opts)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(path))

	return errors.Join(
		err,
		writer.Close(),
	)
}

func (encoder KittyEncoder) EncodeLocalRaw(opts KittyImgOpts) (io.WriteCloser, error) {
	localWriter, err := encoder.EncodeLocalBase64(opts)
	if err != nil {
		return nil, err
	}

	base64Writer := base64.NewEncoder(base64.StdEncoding, encoder.writer)

	return DualWriteCloser{
		writer: base64Writer,
		a:      base64Writer,
		b:      localWriter,
	}, nil
}

func (encoder KittyEncoder) pathHeader(opts KittyImgOpts) error {
	_, err := fmt.Fprint(encoder.writer, opts.ToHeader("a=T", "f=100", "t=f"))
	return err
}

func (encoder KittyEncoder) EncodeLocalBase64(opts KittyImgOpts) (io.WriteCloser, error) {
	if err := encoder.pathHeader(opts); err != nil {
		return nil, err
	}

	return KittyLocalWriter{
		writer: encoder.writer,
	}, nil
}

func (encoder KittyEncoder) Encode(image image.Image) error {
	return encoder.EncodeOpts(image, KittyImgOpts{})
}

func (encoder KittyEncoder) EncodeOpts(image image.Image, opts KittyImgOpts) error {
	writer, err := encoder.EncodeRaw(opts)
	if err != nil {
		return err
	}

	return errors.Join(
		png.Encode(writer, image),
		writer.Close(),
	)
}

func (encoder KittyEncoder) EncodeRaw(opts KittyImgOpts) (io.WriteCloser, error) {
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

func (encoder KittyEncoder) imageHeader(opts KittyImgOpts) error {
	_, err := fmt.Fprint(encoder.writer, opts.ToHeader("a=T", "f=100", "t=d", "m=1"), KITTY_IMG_FTR)
	return err
}

func (encoder KittyEncoder) EncodeBase64(opts KittyImgOpts) (io.WriteCloser, error) {
	if err := encoder.imageHeader(opts); err != nil {
		return nil, err
	}

	return encoder.base64Writer(), nil
}

func (encoder KittyEncoder) base64Writer() io.WriteCloser {
	return KittyImageWriter{
		chunkSize: 4096,
		writer:    encoder.writer,
	}
}
