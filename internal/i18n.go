package internal

var Langs = []Lang{
	{Name: "en"},
	{Name: "en_US"},
	{Name: "en_GB"},
	{Name: "es"},
	{Name: "es_ES"},
	{Name: "es_MX"},
}

type (
	Lang struct {
		ID   int64
		Name string
	}
	Locale struct {
		ID    int64
		Value string
	}
)
