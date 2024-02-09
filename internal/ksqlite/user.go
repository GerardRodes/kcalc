package ksqlite

import "github.com/GerardRodes/kcalc/internal"

// todo: GetUserWithPashHash
// todo: GetUserBySessionToken

func GetUser(id int64) (internal.User, error) {
	type rowt struct {
		Role  string
		Email string
		Lang  int64
	}
	row, err := RQueryOne[rowt](`
		select role, email, lang
		from users
		where id = ?
	`, id)
	if err != nil {
		return internal.User{}, err
	}

	// todo: get family

	return internal.User{
		ID:    id,
		Role:  row.Role,
		Email: row.Email,
		Lang:  row.Lang,
		// todo: Family: *internal.Family,
	}, nil
}
