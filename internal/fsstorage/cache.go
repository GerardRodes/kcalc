package fsstorage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/rs/zerolog/log"
)

func CachedGet(url string) (resp *http.Response, deleteCache func(), err error) {
	fp := filepath.Join("data/cache", internal.MustNormalizeStr(url))
	_, err = os.Stat(fp)

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("stat: %w", err)
	}

	if errors.Is(err, os.ErrNotExist) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, nil, fmt.Errorf("http get: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, nil, fmt.Errorf("http status: %q", resp.Status)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("read all: %w", err)
		}

		if err := resp.Body.Close(); err != nil {
			return nil, nil, fmt.Errorf("close body: %w", err)
		}

		if err := os.WriteFile(fp, data, os.ModePerm); err != nil {
			return nil, nil, fmt.Errorf("write cache: %w", err)
		}

		resp.Body = io.NopCloser(bytes.NewBuffer(data))

		return resp, nil, nil
	} else {
		f, err := os.Open(fp)
		if err != nil {
			return nil, nil, fmt.Errorf("open cache: %w", err)
		}

		return &http.Response{Body: f}, func() {
			if err := os.Remove(fp); err != nil {
				log.Err(err).Msg("remove cache")
			} else {
				log.Debug().Str("fp", fp).Msg("removed from cache")
			}
		}, nil
	}
}
