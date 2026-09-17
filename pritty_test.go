package pritty

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
)

var testFiles []string
var graphics GraphicsHandle

func init() {
	graphics = GraphicsHandle{
		// NOTE: go test captures stdin/stdout to where they are no longer TTYs.
		// This prevents GetSixelSupport() from operating, so always attempting
		// sixels from the test, whether the terminal is capable or not.
		// https://github.com/golang/go/issues/18153
		Sixel: true,
		Iterm: GetItermSupport(),
		Kitty: GetKittySupport(),
	}

	files, error := os.ReadDir("./test_images")
	if error != nil {
		panic(error)
	}

	for _, file := range files {
		switch file.Name() {
		default:
			testFiles = append(testFiles, file.Name())
		}
	}

	os.Stdout.Write([]byte(ESC_ERASE_DISPLAY))
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
	Log(...interface{})
	Logf(string, ...interface{})
}

func encodeImage(path string, encoder Encoder) error {
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

	if localEncoder, ok := encoder.(EncodeLocal); ok {
		fmt.Println("Inline Image")
		inlineErr := localEncoder.Encode(image)
		fmt.Println("\nLocal Image")
		localErr := localEncoder.EncodeLocal(path)
		return errors.Join(inlineErr, localErr)
	} else {
		return encoder.Encode(image)
	}
}

func testImage(logger TestLogger, encoder Encoder, testFiles []string) error {
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

func TestItermWez(test *testing.T) {
	if graphics.Iterm&ItermSupported == 0 {
		test.SkipNow()
	}

	fmt.Println("ITERM/WEZ/MINTTY")
	encoder, err := graphics.ItermEncoder(os.Stdout)
	if err != nil {
		test.Fatal(err)
	}

	if err := testImage(test, encoder, testFiles); err != nil {
		test.Fatal(err)
	}
}

func TestKitty(test *testing.T) {
	if !graphics.Kitty {
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
