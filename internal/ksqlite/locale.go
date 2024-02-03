package ksqlite

import (
	"github.com/GerardRodes/kcalc/internal/ksqlite/gen"
)

func GetLang(name string) (int64, error) {
	return RQueryOne[int64]("select id from langs where name like ?", name)
}

func ListLangsByID() (map[int64]string, error) {
	rows, err := RQuery[gen.Lang]("select * from langs")
	if err != nil {
		return nil, err
	}

	out := make(map[int64]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Name
	}

	return out, nil
}
