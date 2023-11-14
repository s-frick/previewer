package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/s-frick/previewer/resizer"
)

type File struct {
	AbsPath  string
	RelPath  string
	Filename string
	Type     string
}
type Img struct {
	AbsPath  string
	RelPath  string
	Filename string
}
type Image struct {
	Origin  File
	Resized File
}

type Tpl struct {
	template *template.Template
	path     string
}

var (
	srcDir      string
	templateDir string
	outDir      string
	maxWidth    int
	maxHeight   int
)

func initFlags() {
	flag.StringVar(&srcDir, "src", "", "Directory containing source files")
	flag.StringVar(&templateDir, "templates", "", "Directory containing template files to render")
	flag.StringVar(&outDir, "outDir", "", "Resized image target directory.")
	flag.IntVar(&maxWidth, "maxWidth", 0, "maxWidth")
	flag.IntVar(&maxHeight, "maxHeight", 0, "maxHeight")
	flag.Parse()
	if outDir == "" {
		outDir = srcDir
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %d", name, elapsed.Milliseconds())
}

func main() {
	defer timeTrack(time.Now(), "previewer")
	initFlags()
	items := findImages(srcDir)

	resizedImg := resizeImages(items, outDir)

	if templateDir != "" {
		tpls := findTemplates(templateDir)
		renderTemplates(tpls, resizedImg)
	}
	// time.Sleep(10 * time.Second)
}

func findTemplates(templateDir string) []Tpl {
	items := findFiles(templateDir, func(path string, d fs.DirEntry) bool {
		return !d.IsDir() && strings.Contains(path, ".tpl")
	})
	var result []Tpl

	for _, e := range items {
		renderPath := strings.Replace(e.AbsPath, ".tpl", "", 1)
		tpl, err := template.ParseFiles(e.AbsPath)
		if err != nil {
			continue
		}
		result = append(result, Tpl{template: tpl, path: renderPath})
	}

	return result
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

func isJpg(name string) bool {
	return strings.HasSuffix(name, "jpg") || strings.HasSuffix(name, "jpeg")
}
func isPng(name string) bool {
	return strings.HasSuffix(name, "png")
}
func isImage(name string) bool {
	return isJpg(name) || isPng(name)
}

func findFiles(srcDir string, predicate func(path string, d fs.DirEntry) bool) []File {
	var items []File

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if predicate(path, d) {
			relPath := strings.Replace(path, srcDir+"/", "", 1)
			relDir := strings.Replace(relPath, "/"+d.Name(), "", 1)
			items = append(items, File{AbsPath: path, RelPath: relDir, Filename: d.Name()})
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return items
}

func findImages(srcDir string) []File {
	return findFiles(srcDir, func(path string, d fs.DirEntry) bool {
		return !d.IsDir() && !strings.Contains(d.Name(), "__th__") && isImage(d.Name())
	})
}

func resizeImages(items []File, targetPath string) []Image {
	var resizedImg []Image
	var wg sync.WaitGroup
	for _, item := range items {

		//FIXME: bug when files in srcDir
		targetDir := fmt.Sprintf("%s/%s", targetPath, item.RelPath)
		if err := os.MkdirAll(targetDir, 0777); err != nil {
			log.Println(err)
		}
		log.Printf(targetDir)

		var newFilename string
		var typ string
		if isJpg(item.Filename) {
			newFilename = strings.Replace(item.Filename, ".jpg", "__th__.jpg", 1)
			newFilename = strings.Replace(item.Filename, ".jpeg", "__th__.jpeg", 1)
			typ = "jpg"
		} else if isPng(item.Filename) {
			newFilename = strings.Replace(item.Filename, ".png", "__th__.png", 1)
			typ = "png"
		}
		log.Println(newFilename)
		absPathOut := fmt.Sprintf("%s/%s", targetDir, newFilename)

		wg.Add(1)
		// TODO: we should take a channel and buffer/limit the tasks
		go func(item File, absPathOut string) {
			defer wg.Done()
			resizeImageIO(item, absPathOut, typ)
		}(item, absPathOut)

		resizedImg = append(resizedImg, Image{Origin: item, Resized: File{AbsPath: absPathOut, RelPath: item.RelPath, Filename: newFilename}})
	}
	wg.Wait()
	return resizedImg
}

func resizeImageIO(item File, absPathOut string, typ string) {
	input, err := os.Open(item.AbsPath)
	if err != nil {
		log.Printf("no such file: %s, %v", item.AbsPath, err)
	}
	defer input.Close()

	output, err := os.Create(absPathOut)
	if err != nil {
		log.Println(err)
	}
	defer output.Close()

	err = resizer.ResizeImage(input, output, resizer.ResizerOptions{MaxWidth: maxWidth, MaxHeight: maxHeight, Type: typ})
	if err != nil {
		log.Printf("error while resizing image %s %s: %v", typ, item.AbsPath, err)
	}
}
