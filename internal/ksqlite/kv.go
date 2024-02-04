package ksqlite

import (
	"encoding/binary"
)

// todo: make KV* generic

func KVSetUInt64(k string, v uint64) error {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], v)
	return Exec(`
		insert into kv (k, v)
		values (?, ?)
		on conflict do update
		set v = excluded.v;
		`, k, data[:])
}

func KVGetUInt64(k string) (uint64, error) {
	data, err := RQueryOne[[]byte]("select v from kv where k = ?", k)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(data), nil
}
