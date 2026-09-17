package main

import (
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/arzumify/pritty"
)

var testFiles []string
var graphics pritty.GraphicsHandle

const (
	ESC_ERASE_DISPLAY = "\x1b[2J\x1b[0;0H"
)

func main() {
	var err error
	graphics, err = pritty.GetGraphicsHandle(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}

	files, err := os.ReadDir("./test_images")
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		switch file.Name() {
		default:
			testFiles = append(testFiles, file.Name())
		}
	}

	os.Stdout.Write([]byte(ESC_ERASE_DISPLAY))

	testing.Main(nil, []testing.InternalTest{
		{
			Name: "sixel",
			F:    TestSixel,
		},
		{
			Name: "iterm",
			F:    TestIterm,
		},
		{
			Name: "kitty",
			F:    TestKitty,
		},
	}, nil, nil)
}

func getFile(fpath string) (*os.File, int64, error) {

	pF, E := os.Open(fpath)
	if E != nil {
		return nil, 0, E
	}

	fInf, E := pF.Stat()
	if E != nil {
		pF.Close()
		return nil, 0, E
	}

	return pF, fInf.Size(), nil
}

type TestLogger interface {
	Log(...any)
	Logf(string, ...any)
}

func encodeImage(path string, encoder pritty.Encoder) error {
	file, size, err := getFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println(path)

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	image, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	fmt.Printf("[FMT: %s, W: %d, H: %d, LEN: %d, IMG: %T]\n", format, config.Width, config.Height, size, image)

	defer fmt.Println("")

	if localEncoder, ok := encoder.(pritty.EncodeLocal); ok {
		fmt.Println("Inline Image")
		inlineErr := localEncoder.Encode(image)
		fmt.Println("\nLocal Image")
		localErr := localEncoder.EncodeLocal(path)
		return errors.Join(inlineErr, localErr)
	} else {
		return encoder.Encode(image)
	}
}

func testImage(logger TestLogger, encoder pritty.Encoder, testFiles []string) error {
	baseDir, err := filepath.Abs("./test_images")
	if err != nil {
		return err
	}

	for _, file := range testFiles {
		path := baseDir + "/" + file
		logger.Log(path)

		err := encodeImage(path, encoder)
		if err != nil {
			logger.Log(err)
		}
	}

	return nil
}

func TestSixel(test *testing.T) {
	if !graphics.Sixel {
		test.SkipNow()
	}

	fmt.Println("SIXEL")
	encoder, err := graphics.SixelEncoder(os.Stdout)
	if err != nil {
		test.Fatal(err)
	}

	if err := testImage(test, encoder, testFiles); err != nil {
		test.Fatal(err)
	}
}

func TestIterm(test *testing.T) {
	if graphics.Iterm&pritty.ItermSupported == 0 {
		test.SkipNow()
	}

	fmt.Println("ITERM")
	encoder, err := graphics.ItermEncoder(os.Stdout)
	if err != nil {
		test.Fatal(err)
	}

	if err := testImage(test, encoder, testFiles); err != nil {
		test.Fatal(err)
	}
}

func TestKitty(test *testing.T) {
	if graphics.Kitty&pritty.KittySupported == 0 {
		test.SkipNow()
	}

	fmt.Println("KITTY")
	encoder, err := graphics.KittyEncoder(os.Stdout)
	if err != nil {
		test.Fatal(err)
	}

	if err := testImage(test, encoder, testFiles); err != nil {
		test.Fatal(err)
	}
}
