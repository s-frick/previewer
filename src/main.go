package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

var (
	srcDir      string
	templateDir string
)

func init() {
	flag.StringVar(&srcDir, "src", "", "Directory containing source files")
	flag.StringVar(&templateDir, "templates", "", "Directory containing template files to render")
	flag.Parse()
}

func main() {

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			relPath := strings.Replace(path, srcDir+"/", "", 1)
			relDir := strings.Replace(relPath, "/"+d.Name(), "", 1)

			fmt.Println(relDir, d.Name())
		}

		return nil
	})
	if err != nil {
		log.Println(err)
	}

	imagePath := "/home/sfrick/git/private/projects/previewer/resources/areana-tech-talk_driver.jpg"
	input, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("no such file: %s, %v", imagePath, err)
	}
	defer input.Close()

	os.MkdirAll("/home/sfrick/git/private/projects/previewer/resources/out", 0777)
	output, err := os.Create("/home/sfrick/git/private/projects/previewer/resources/out/areana-tech-talk_driver.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	src, err := jpeg.Decode(input)
	if err != nil {
		log.Fatalf("image is not a jpeg, %v", err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, 640, 640))

	draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	jpeg.Encode(output, dst, nil)

}
