package rasterm

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
func IsKittyCapable() bool {

	// TODO: more rigorous check
	V := GetEnvIdentifiers()
	return (len(V["KITTY_WINDOW_ID"]) > 0) || (V["TERM_PROGRAM"] == "wezterm") || (V["TERM_PROGRAM"] == "ghostty")
}

func (self KittyEncoder) EncodeLocal(location string, opts KittyImgOpts) error {
	writer, err := self.EncodeLocalRaw(opts)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(location))

	return errors.Join(
		err,
		writer.Close(),
	)
}

func (self KittyEncoder) EncodeLocalRaw(opts KittyImgOpts) (io.WriteCloser, error) {
	localWriter, err := self.EncodeLocalBase64(opts)
	if err != nil {
		return nil, err
	}

	base64Writer := base64.NewEncoder(base64.RawStdEncoding, self.writer)

	return DualWriteCloser{
		inner: base64Writer,
		a:     base64Writer,
		b:     localWriter,
	}, nil
}

func (self KittyEncoder) pathHeader(opts KittyImgOpts) error {
	_, err := fmt.Fprint(self.writer, opts.ToHeader("a=T", "f=100", "t=f"))
	return err
}

type KittyLocalEncoder struct {
	inner io.Writer
}

func (self KittyLocalEncoder) Write(buffer []byte) (int, error) {
	return self.inner.Write(buffer)
}

func (self KittyLocalEncoder) Close() error {
	_, err := fmt.Fprint(self.inner, KITTY_IMG_FTR)
	return err
}

func (self KittyEncoder) EncodeLocalBase64(opts KittyImgOpts) (io.WriteCloser, error) {
	if err := self.pathHeader(opts); err != nil {
		return nil, err
	}

	return KittyLocalEncoder{
		inner: self.writer,
	}, nil
}

func (self KittyEncoder) EncodeImage(image image.Image, opts KittyImgOpts) error {
	writer, err := self.EncodeImageRaw(opts)
	if err != nil {
		return err
	}

	return errors.Join(
		png.Encode(writer, image),
		writer.Close(),
	)
}

type DualWriteCloser struct {
	inner io.Writer
	a     io.Closer
	b     io.Closer
}

func (self DualWriteCloser) Write(buffer []byte) (int, error) {
	return self.inner.Write(buffer)
}

func (self DualWriteCloser) Close() error {
	return errors.Join(
		self.a.Close(),
		self.b.Close(),
	)
}

func (self KittyEncoder) EncodeImageRaw(opts KittyImgOpts) (io.WriteCloser, error) {
	imageWriter, err := self.EncodeImageBase64(opts)
	if err != nil {
		return nil, err
	}

	base64Writer := base64.NewEncoder(base64.RawStdEncoding, imageWriter)

	return DualWriteCloser{
		inner: base64Writer,
		a:     base64Writer,
		b:     imageWriter,
	}, nil
}

func (self KittyEncoder) imageHeader(opts KittyImgOpts) error {
	_, err := fmt.Fprint(self.writer, opts.ToHeader("a=T", "f=100", "t=d", "m=1"), KITTY_IMG_FTR)
	return err
}

func (self KittyEncoder) EncodeImageBase64(opts KittyImgOpts) (io.WriteCloser, error) {
	if err := self.imageHeader(opts); err != nil {
		return nil, err
	}

	return self.base64Writer(), nil
}

func (self KittyEncoder) base64Writer() io.WriteCloser {
	writer := KittyImageWriter{
		chunkSize: 4096,
		inner:     self.writer,
	}

	return writer
}
