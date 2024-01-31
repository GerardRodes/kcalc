package ksqlite

func GetLang(name string) (int64, error) {
	return RQueryOne[int64]("select id from langs where name like ?", name)
}
