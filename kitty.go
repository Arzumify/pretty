package pritty

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"math/rand"
	"os"
	"path"
	"strings"
)

// See https://sw.kovidgoyal.net/kitty/graphics-protocol.html for more details.

const (
	KITTY_IMG_HDR = "\x1b_G"
	KITTY_IMG_FTR = "\x1b\\"
	// base64 decode of "pritty"
	KITTY_PROBE_ID = 2797120951
)

var (
	KITTY_PROBE_RESPONSE = fmt.Sprintf(
		"%si=%d;",
		KITTY_IMG_HDR,
		KITTY_PROBE_ID,
	)
	KITTY_PROBE_OK = []byte("OK")
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

const (
	KittySupported = 1 << iota
	KittyLocal
)

type KittySupport = uint8

// checks if terminal supports kitty image protocol
func GetKittySupport(stdin io.Reader, stdout io.Writer) (support KittySupport, err error) {
	path := path.Join(os.TempDir(), fmt.Sprintf("pritty-probe-%d", rand.Uint32()))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return 0, err
	}

	_, err = file.Write([]byte{0, 0, 0})
	if err != nil {
		return 0, errors.Join(
			os.RemoveAll(path),
			err,
		)
	}

	fmt.Fprintf(
		stdout,
		"%s%s%s",
		KittyImgOpts{
			SrcWidth:  1,
			SrcHeight: 1,
			ImageId:   KITTY_PROBE_ID,
		}.ToHeader("a=q", "s=1", "v=1", "t=f", "f=24"),
		base64.StdEncoding.EncodeToString([]byte(path)),
		KITTY_IMG_FTR,
	)

	err = awaitTerminalResponse(stdin, []byte(KITTY_PROBE_RESPONSE))
	if err != nil {
		if errors.Is(err, E_TIMED_OUT) {
			return 0, os.RemoveAll(path)
		} else {
			return 0, errors.Join(
				os.RemoveAll(path),
				err,
			)
		}
	}

	buffer := make([]byte, 2)
	_, err = io.ReadFull(stdin, buffer)
	if err != nil {
		return KittySupported, errors.Join(
			os.RemoveAll(path),
			err,
		)
	}

	if bytes.Equal(buffer, KITTY_PROBE_OK) {
		return KittySupported | KittyLocal, nil
	} else {
		return KittySupported, os.RemoveAll(path)
	}
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
