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
		Required: true,
	})
	forceInline := parser.Flag("i", "force-inline", &argparse.Options{
		Default: false,
	})

	if err := parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
		os.Exit(1)
	}

	if *path == "" {
		fmt.Println(parser.Usage("Error: no path provided"))
		os.Exit(1)
	}

	graphics, err := pritty.GetGraphicsHandle(os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
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
