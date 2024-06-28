package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	ha "github.com/mkelcik/go-ha-client"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype/truetype"
)

var (
	sensors = [][]string{
		{"AQI", "sensor.outdoor_aqi"},
		{"Outside", "sensor.gw2000b_outdoor_temperature"},
		{"Upstairs", "sensor.upstairs_temperature_temperature"},
		{"Lake", "sensor.gw2000b_soil_temperature_1"},
	}
)

const (
	fontSize       = 32
	paddingBetween = 32 / 2
)

type content struct {
	fname string
	ftype string
	fdata []byte
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{200, 100, 0, 255}
	point := fixed.Point26_6{fixed.I(x), fixed.I(y)}

	fFace, err := truetype.Parse(goregular.TTF)

	if err != nil {
		panic(err)
	}

	h := font.HintingNone
	d := &font.Drawer{
		Dst: img,
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

func main() {
	client := ha.NewClient(ha.ClientConfig{Token: os.Getenv("HA_TOKEN"), Host: os.Getenv("HA_HOST")}, &http.Client{
		Timeout: 30 * time.Second,
	})

	img := image.NewRGBA(image.Rect(0, 0, 240, 240))

	offset := fontSize

	for _, slic := range sensors {
		name, entityId := slic[0], slic[1]
		stateStr := "Unknown"
		state, err := client.GetStateForEntity(context.Background(), entityId)
		if err == nil {
			stateStr = state.State
		}

		addLabel(img, 1, offset, fmt.Sprintf("%s: %s", name, stateStr))

		offset += fontSize + paddingBetween
	}

	var b bytes.Buffer
	f := bufio.NewWriter(&b)

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 100}); err != nil {
		panic(err)
	}

	err := retry.Do(
		func() error {
			_, err := sendPostRequest("http://192.168.50.175/doUpload?dir=/image/", content{
				fname: "hello-go.jpg",
				ftype: "image/jpeg",
				fdata: b.Bytes(),
			})
			return err
		},
		retry.Attempts(10),
		retry.Delay(5*time.Second),
	)

	if err != nil {
		panic(err)
	}
}

func sendPostRequest(url string, files ...content) ([]byte, error) {
	var (
		buf = new(bytes.Buffer)
		w   = multipart.NewWriter(buf)
	)

	for _, f := range files {
		part, err := w.CreateFormFile(f.ftype, filepath.Base(f.fname))
		if err != nil {
			return []byte{}, err
		}

		_, err = part.Write(f.fdata)
		if err != nil {
			return []byte{}, err
		}
	}

	err := w.Close()
	if err != nil {
		return []byte{}, err
	}

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	cnt, err := io.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	return cnt, nil
}
