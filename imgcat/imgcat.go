package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/akamensky/argparse"
	"github.com/arzumify/pritty"
)

func main() {
	parser := argparse.NewParser("imgcat", "Prints images to the console")
	path := parser.StringPositional(&argparse.Options{
		Help: "Path to the image",
	})
	forceInline := parser.Flag("f", "force-inline", &argparse.Options{
		Default: false,
		Help:    "Forces all images to be sent as inline rather than as paths",
	})
	info := parser.Flag("i", "info", &argparse.Options{
		Default: false,
		Help:    "Does not display images and instead reports information about the terminal",
	})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}

	graphics, err := pritty.GetGraphicsHandle(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}

	if *info {
		fmt.Printf(
			"Sixel support: %t\n"+
				"Iterm support: %t (streaming: %t)\n"+
				"Kitty support: %t (local: %t)",
			graphics.Sixel,
			graphics.Iterm&pritty.ItermSupported != 0,
			graphics.Iterm&pritty.ItermStreaming != 0,
			graphics.Kitty&pritty.KittySupported != 0,
			graphics.Kitty&pritty.KittyLocal != 0,
		)
		return
	}

	if *path == "" {
		fmt.Println(parser.Usage("Error: no path provided"))
		os.Exit(1)
	}

	encoder, err := graphics.Optimal(os.Stdout)
	if err != nil {
		panic(err)
	}

	if !*forceInline {
		if localEncoder, ok := encoder.(pritty.EncodeLocal); ok {
			absolutePath, err := filepath.Abs(*path)
			if err != nil {
				panic(err)
			}

			if err := localEncoder.EncodeLocal(absolutePath); err != nil {
				panic(err)
			}

			return
		}
	}

	file, err := os.Open(*path)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	image, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	if err := encoder.Encode(image); err != nil {
		panic(err)
	}
}
