package internal

type (
	Unit struct {
		ID       int64
		Symbol   string
		Quantity UnitQuantity
		Locales  map[string]Locale
	}
	UnitQuantity struct {
		ID      int64
		Locales map[string]Locale
	}
)
