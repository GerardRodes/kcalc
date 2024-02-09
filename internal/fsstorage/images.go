package fsstorage

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite"
	"github.com/jxskiss/base62"
	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/nfnt/resize"
	"github.com/rs/zerolog/log"
)

var (
	lastIDKey = "fsstorage_images_last_id"
	lastID    atomic.Uint64
	b62enc    = base62.NewEncoding("l5ON9XsidVFxGJST20gEuBa4fhvkqUK1cjboDnMCALIp3zPQ8YWwy6ZemrRHt7") // random base62
)

func Init() error {
	id, err := ksqlite.KVGet[uint64](lastIDKey)
	if err != nil && !errors.Is(err, internal.ErrNotFound) {
		return fmt.Errorf("KVGet: %w", err)
	}

	if !lastID.CompareAndSwap(0, id) {
		return errors.New("cannot swap last id")
	}

	log.Debug().Uint64(lastIDKey, lastID.Load()).Msg("kv loaded")

	return nil
}

func DownloadAndStoreImage(rawURL string) (foodImage internal.Image, outErr error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return internal.Image{}, fmt.Errorf("parse url: %w", err)
	}

	res, deleteCache, err := CachedGet(parsedURL.String())
	if err != nil {
		return internal.Image{}, fmt.Errorf("download img: %w", err)
	}
	defer func() {
		if outErr != nil {
			deleteCache()
		}
	}()

	imgData, err := io.ReadAll(res.Body)
	if err != nil {
		return internal.Image{}, fmt.Errorf("read all: %w", err)
	}

	if err := res.Body.Close(); err != nil {
		return internal.Image{}, fmt.Errorf("close body: %w", err)
	}

	uri, err := StoreImage(imgData, mime.TypeByExtension(filepath.Ext(rawURL)))
	if err != nil {
		return internal.Image{}, err
	}

	return internal.Image{
		URI: uri,
		// Kind: "TODO",
	}, nil
}

func StoreImage(data []byte, mimetype string) (uri string, outErr error) {
	var err error
	var img image.Image

	buf := bytes.NewBuffer(data)
	switch mimetype {
	case "image/jpeg":
		img, err = jpeg.Decode(buf)
	case "image/png":
		img, err = png.Decode(buf)
	case "image/webp":
		img, err = webp.Decode(buf, &decoder.Options{
			UseThreads: true,
		})
	default:
		return "", fmt.Errorf("%w: not supported image type %q", internal.ErrInvalid, mimetype)
	}

	if err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	img = resize.Thumbnail(640, 640, img, resize.NearestNeighbor)

	prevID := lastID.Load()
	newID := lastID.Add(1)
	defer func() {
		if outErr != nil {
			// try to recover some ids
			lastID.CompareAndSwap(newID, prevID)
		} else {
			outErr = ksqlite.KVSet(lastIDKey, newID)
		}
	}()
	name := string(b62enc.FormatUint(newID))
	uri = filepath.Join("images", name+".webp")
	output, err := os.Create(filepath.Join(internal.RootDir, "content", uri))
	if err != nil {
		return "", fmt.Errorf("create webp: %w", err)
	}
	defer output.Close()

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetPhoto, 90)
	if err != nil {
		return "", fmt.Errorf("create lossy encoder: %w", err)
	}

	if err := webp.Encode(output, img, options); err != nil {
		return "", fmt.Errorf("encode webp: %w", err)
	}

	return uri, nil
}
