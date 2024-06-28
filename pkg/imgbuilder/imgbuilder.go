package imgbuilder

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/jpeg"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type ImgBuilder struct {
	image *image.RGBA
}

func NewImgBuilder(width, height int) *ImgBuilder {
	return &ImgBuilder{
		image: image.NewRGBA(image.Rect(0, 0, width, height)),
	}
}

func (i *ImgBuilder) AddLabel(label string, fontSize float64, x, y int) {
	col := color.RGBA{200, 100, 0, 255}
	point := fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}

	fFace, err := truetype.Parse(goregular.TTF)

	if err != nil {
		panic(err)
	}

	h := font.HintingNone
	d := &font.Drawer{
		Dst: i.image,
		Src: image.NewUniform(col),
		Face: truetype.NewFace(fFace, &truetype.Options{
			Size:    fontSize,
			DPI:     72,
			Hinting: h,
		}),
		Dot: point,
	}
	d.DrawString(label)
}

func (i *ImgBuilder) Generate() ([]byte, error) {
	var b bytes.Buffer
	f := bufio.NewWriter(&b)

	err := jpeg.Encode(f, i.image, &jpeg.Options{Quality: 100})
	return b.Bytes(), err
}
