package foodcomposition

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

func Ingest(ctx context.Context, csvPath string) error {
	f, err := os.OpenFile(csvPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = 37
	r.ReuseRecord = true
	r.TrimLeadingSpace = true

	for {
		_, err := r.Read() // todo:
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("cannot read record: %w", err)
		}

	}
}
