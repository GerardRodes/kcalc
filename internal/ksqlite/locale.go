package ksqlite

func GetLang(name string) (int64, error) {
	return RQueryOne[int64]("select id from langs where name like ?", name)
}

func ListLangsByID() (map[int64]string, error) {
	type rowt struct {
		ID   int64
		Name string
	}
	rows, err := RQuery[rowt]("select id, name from langs")
	if err != nil {
		return nil, err
	}

	out := make(map[int64]string, len(rows))
	for _, row := range rows {
		out[row.ID] = row.Name
	}

	return out, nil
}
