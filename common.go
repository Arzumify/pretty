package pritty

import (
	"bytes"
	"errors"
	"go/types"
	"image"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type GraphicsHandle struct {
	Sixel bool
	Iterm ItermSupport
	Kitty KittySupport
}

func GetGraphicsHandle(stdin *os.File, stdout *os.File) (_ GraphicsHandle, err error) {
	fd := int(stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return GraphicsHandle{
			Sixel: false,
			Iterm: 0,
			Kitty: 0,
		}, err
	}

	defer func() {
		err = term.Restore(fd, state)
	}()

	iterm := GetItermSupport()

	sixel, err := GetSixelSupport(stdin, stdout)
	if err != nil {
		return GraphicsHandle{
			Sixel: false,
			Iterm: iterm,
			Kitty: 0,
		}, err
	}

	kitty, err := GetKittySupport(stdin, stdout)
	if err != nil {
		return GraphicsHandle{
			Sixel: sixel,
			Iterm: iterm,
			Kitty: 0,
		}, err
	}

	return GraphicsHandle{
		Sixel: sixel,
		Iterm: iterm,
		Kitty: kitty,
	}, nil
}

// Used to get rid of the EncodeLocal functions on Kitty if unsupported
type EncodeWrapper struct {
	inner Encoder
}

func (wrapper EncodeWrapper) Encode(image image.Image) error {
	return wrapper.inner.Encode(image)
}

type Encoder interface {
	Encode(image image.Image) error
}

type EncodeLocal interface {
	Encoder
	EncodeLocal(path string) error
}

func (handle GraphicsHandle) Optimal(writer io.Writer) (Encoder, error) {
	if handle.Kitty&KittySupported != 0 {
		if handle.Kitty&KittyLocal != 0 {
			return NewKittyEncoder(writer), nil
		} else {
			return EncodeWrapper{inner: NewKittyEncoder(writer)}, nil
		}
	} else if handle.Iterm&ItermSupported != 0 {
		return NewItermEncoder(writer, handle.Iterm&ItermStreaming != 0), nil
	} else if handle.Sixel {
		return NewSixelEncoder(writer), nil
	} else {
		return nil, E_NO_PROTOCOL
	}
}

func (handle GraphicsHandle) KittyEncoder(writer io.Writer) (Encoder, error) {
	if handle.Kitty&KittySupported != 0 {
		if handle.Kitty&KittyLocal != 0 {
			return NewKittyEncoder(writer), nil
		} else {
			return EncodeWrapper{inner: NewKittyEncoder(writer)}, nil
		}
	} else {
		return nil, E_UNSUPPORTED
	}
}

func (handle GraphicsHandle) ItermEncoder(writer io.Writer) (Encoder, error) {
	if handle.Iterm&ItermSupported != 0 {
		return NewItermEncoder(writer, handle.Iterm&ItermStreaming != 0), nil
	} else {
		return nil, E_UNSUPPORTED
	}
}

func (handle GraphicsHandle) SixelEncoder(writer io.Writer) (Encoder, error) {
	if handle.Sixel {
		return NewSixelEncoder(writer), nil
	} else {
		return nil, E_UNSUPPORTED
	}
}

const (
	ESC_ERASE_DISPLAY = "\x1b[2J\x1b[0;0H"
)

var (
	E_NON_TTY     = errors.New("not a tty")
	E_TIMED_OUT   = errors.New("terminal response timed out")
	E_UNSUPPORTED = errors.New("graphics method not supported")
	E_NO_PROTOCOL = errors.New("no supported protocols")
)

func IsTmuxScreen() bool {
	TERM := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	return strings.HasPrefix(TERM, "screen")
}

type DualWriteCloser struct {
	writer io.Writer
	a      io.Closer
	b      io.Closer
}

func (writeCloser DualWriteCloser) Write(buffer []byte) (int, error) {
	return writeCloser.writer.Write(buffer)
}

func (writeCloser DualWriteCloser) Close() error {
	return errors.Join(
		writeCloser.a.Close(),
		writeCloser.b.Close(),
	)
}

func pollTerminal(stdin io.Reader, header []byte, expected []byte) error {
	for !bytes.Equal(header, expected) {
		if _, err := io.ReadFull(stdin, header); err != nil {
			return err
		}
	}
	return nil
}

func awaitTerminalResponse(stdin io.Reader, expected []byte) error {
	header := make([]byte, len(expected))
	done := make(chan error)
	timeout := make(chan types.Nil)
	go func() {
		done <- pollTerminal(stdin, header, expected)
	}()
	go func() {
		time.Sleep(time.Millisecond * 100)
		done <- nil
	}()
	select {
	case result := <-done:
		return result
	case <-timeout:
		return E_TIMED_OUT
	}
}

func lcaseEnv(k string) string {
	return strings.ToLower(strings.TrimSpace(os.Getenv(k)))
}

func GetEnvIdentifiers() map[string]string {

	KEYS := []string{"TERM", "TERM_PROGRAM", "LC_TERMINAL", "VIM_TERMINAL", "KITTY_WINDOW_ID"}
	V := make(map[string]string)
	for _, K := range KEYS {
		V[K] = lcaseEnv(K)
	}

	return V
}
