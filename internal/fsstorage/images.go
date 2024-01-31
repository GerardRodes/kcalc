package fsstorage

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/segmentio/ksuid"
)

var (
	prevL sync.Mutex
	prev  = ksuid.New()
)

func StoreImage(rawURL string) (foodImage internal.FoodImage, outErr error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return internal.FoodImage{}, fmt.Errorf("parse url: %w", err)
	}

	res, deleteCache, err := CachedGet(parsedURL.String())
	if err != nil {
		return internal.FoodImage{}, fmt.Errorf("download img: %w", err)
	}
	defer func() {
		if outErr != nil {
			deleteCache()
		}
	}()

	imgData, err := io.ReadAll(res.Body)
	if err != nil {
		return internal.FoodImage{}, fmt.Errorf("read all: %w", err)
	}

	if err := res.Body.Close(); err != nil {
		return internal.FoodImage{}, fmt.Errorf("close body: %w", err)
	}

	var img image.Image
	var imgConfig image.Config
	switch mmt := mime.TypeByExtension(filepath.Ext(rawURL)); mmt {
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewBuffer(imgData))
		if err != nil {
			return internal.FoodImage{}, fmt.Errorf("decode jpeg: %w", err)
		}

		imgConfig, err = jpeg.DecodeConfig(bytes.NewBuffer(imgData))
		if err != nil {
			return internal.FoodImage{}, fmt.Errorf("decode jpeg config: %w", err)
		}
	case "image/png":
		img, err = png.Decode(bytes.NewBuffer(imgData))
		if err != nil {
			return internal.FoodImage{}, fmt.Errorf("decode png: %w", err)
		}

		imgConfig, err = png.DecodeConfig(bytes.NewBuffer(imgData))
		if err != nil {
			return internal.FoodImage{}, fmt.Errorf("decode png config: %w", err)
		}
	default:
		// todo:
		return internal.FoodImage{}, fmt.Errorf("unsupported image type %q", mmt)
	}

	var id ksuid.KSUID
	{
		prevL.Lock()
		id = prev.Next()
		prev = id
		prevL.Unlock()
	}

	output, err := os.Create(filepath.Join(internal.RootDir, "images", id.String()+".webp"))
	if err != nil {
		return internal.FoodImage{}, fmt.Errorf("create webp: %w", err)
	}
	defer output.Close()

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetPhoto, 90)
	if err != nil {
		return internal.FoodImage{}, fmt.Errorf("create lossy encoder: %w", err)
	}

	if err := webp.Encode(output, img, options); err != nil {
		return internal.FoodImage{}, fmt.Errorf("encode webp: %w", err)
	}

	return internal.FoodImage{
		URI:    filepath.Join("images", id.String()+".webp"),
		Width:  int64(imgConfig.Width),
		Height: int64(imgConfig.Height),
	}, nil
}
