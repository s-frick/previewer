package resizer

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"

	"golang.org/x/image/draw"
)

type ResizerOptions struct {
	Type      string
	MaxWidth  int
	MaxHeight int
}

// FIXME: bug when decoding/encoding jpg
// 2023/11/14 23:57:26 error while resizing image jpg images/logos/tech-talk/comic_things_2.jpg: unexpected EOF
func ResizeImage(input io.Reader, output io.Writer, options ResizerOptions) error {
	src, err := decode(input, options.Type)
	if err != nil {
		return err
	}

	width, height := scaleKeepRatio(src.Bounds().Dx(), src.Bounds().Dy(), options.MaxHeight, options.MaxWidth)
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	draw.ApproxBiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

	err = encode(output, dst, options.Type)
	if err != nil {
		return err
	}
	return nil
}
func decode(input io.Reader, t string) (image.Image, error) {
	switch t {
	case "png":
		return png.Decode(input)
	case "jpg":
		return jpeg.Decode(input)
	default:
		return nil, errors.New("unsupported file type")
	}
}

func encode(output io.Writer, img image.Image, t string) error {
	switch t {
	case "png":
		return png.Encode(output, img)
	case "jpg":
		return jpeg.Encode(output, img, nil)
	default:
		return errors.New("unsupported file type")
	}
}
func scaleKeepRatio(oldWidth int, oldHeight int, maxWidth int, maxHeight int) (width int, height int) {
	if maxWidth == 0 && maxHeight == 0 {
		maxWidth = oldWidth / 2
		maxHeight = oldHeight / 2
	}
	if maxWidth == 0 {
		maxWidth = math.MaxInt
	}
	if maxHeight == 0 {
		maxHeight = math.MaxInt
	}
	ratioX := float32(maxWidth) / float32(oldWidth)
	ratioY := float32(maxHeight) / float32(oldHeight)
	ratio := min(ratioX, ratioY)
	width = int(float32(oldWidth) * ratio)
	height = int(float32(oldHeight) * ratio)
	return width, height
}
