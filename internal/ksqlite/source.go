package ksqlite

import (
	"errors"
	"fmt"

	"github.com/GerardRodes/kcalc/internal"
	"github.com/GerardRodes/kcalc/internal/ksqlite/gen"
)

func AddSource(src internal.Source) (int64, error) {
	sourceID, err := GetSourceID(src.Name)
	if errors.Is(err, internal.ErrNotFound) {
		return WQueryOne[int64](`
			insert into sources (name)
			values (?)
			on conflict do update
			set id = id
			returning id;
		`, src.Name)
	}

	if err != nil {
		return 0, fmt.Errorf("get source id: %w", err)
	}

	return sourceID, nil
}

func GetSourceID(name string) (int64, error) {
	return RQueryOne[int64]("select id from sources where name = ?", name)
}

func ListSourcesByID() (map[int64]string, error) {
	rows, err := RQuery[gen.Source]("select * from sources")
	if err != nil {
		return nil, err
	}

	out := make(map[int64]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Name
	}

	return out, nil
}
