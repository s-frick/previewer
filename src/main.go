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
	"text/template"

	"golang.org/x/image/draw"
)

type Img struct {
	AbsPath  string
	RelPath  string
	Filename string
}
type Image struct {
	Origin  Img
	Resized Img
}

type Tpl struct {
	template *template.Template
	path     string
}

var (
	srcDir      string
	templateDir string
	outDir      string
)

func initFlags() {
	flag.StringVar(&srcDir, "src", "", "Directory containing source files")
	flag.StringVar(&templateDir, "templates", "", "Directory containing template files to render")
	flag.StringVar(&outDir, "outDir", "", "Resized image target directory.")
	flag.Parse()
	if outDir == "" {
		outDir = srcDir
	}
	log.Println("OUTDIR: ", outDir)
	log.Println("SRCDIR: ", srcDir)
}

func main() {
	initFlags()
	items := findImages(srcDir)
	resizedImg := resizeImages(items, outDir)
	tpls := findTemplates(templateDir)
	renderTemplates(tpls, resizedImg)
}

func findTemplates(templateDir string) []Tpl {
	var items []Tpl

	err := filepath.WalkDir(templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if strings.Contains(path, ".tpl") {
				renderPath := strings.Replace(path, ".tpl", "", 1)
				tpl, err := template.ParseFiles(path)
				if err != nil {
					return err
				}
				items = append(items, Tpl{template: tpl, path: renderPath})
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return items
}

func renderTemplates(tpls []Tpl, imgs []Image) {
	for _, tpl := range tpls {
		output, err := os.Create(tpl.path)
		if err != nil {
			log.Println(err)
		}
		defer output.Close()

		if err := tpl.template.Execute(output, imgs); err != nil {
			log.Printf("error while render template: %v", err)
			return
		}
	}
}

func findImages(srcDir string) []Img {
	var items []Img

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && !strings.Contains(d.Name(), "__th__") {
			relPath := strings.Replace(path, srcDir+"/", "", 1)
			relDir := strings.Replace(relPath, "/"+d.Name(), "", 1)
			items = append(items, Img{AbsPath: path, RelPath: relDir, Filename: d.Name()})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return items
}

func resizeImages(items []Img, targetPath string) []Image {
	var resizedImg []Image
	for _, item := range items {
		// maxWidth maxHeigth
		newFilename, absPathOut := resizeImage(item, targetPath)

		resizedImg = append(resizedImg, Image{Origin: item, Resized: Img{AbsPath: absPathOut, RelPath: item.RelPath, Filename: newFilename}})
	}
	return resizedImg
}

func resizeImage(item Img, targetPath string) (string, string) {
	input, err := os.Open(item.AbsPath)
	if err != nil {
		log.Printf("no such file: %s, %v", item.AbsPath, err)
	}
	defer input.Close()

	targetDir := fmt.Sprintf("%s/%s", targetPath, item.RelPath)
	if err := os.MkdirAll(targetDir, 0777); err != nil {
		log.Println(err)
	}
	log.Printf(targetDir)

	newFilename := strings.Replace(item.Filename, ".jpg", "__th__.jpg", 1)
	log.Println(newFilename)
	absPathOut := fmt.Sprintf("%s/%s", targetDir, newFilename)
	output, err := os.Create(absPathOut)
	if err != nil {
		log.Println(err)
	}
	defer output.Close()

	src, err := jpeg.Decode(input)
	if err != nil {
		log.Printf("image is not a jpeg, %v", err)
	}

	dst := image.NewRGBA(image.Rect(0, 0, 160, 160))

	draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	jpeg.Encode(output, dst, nil)
	return newFilename, absPathOut
}
