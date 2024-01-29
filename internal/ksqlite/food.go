package ksqlite

import (
	"fmt"

	"github.com/GerardRodes/kcalc/internal"
)

func AddFoods(foods ...internal.Food) error {
	c, unlock := WConn()
	defer unlock()
	for _, food := range foods {
		if err := addFood(c, food); err != nil {
			return fmt.Errorf("add food: %w", err)
		}
	}

	return nil
}

func addFood(c *Conn, food internal.Food) error {
	return c.conn.WithTx(func() error {
		return nil
	})
}
