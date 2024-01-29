package internal

type (
	Food struct {
		ID                 int64
		DetailsFromSources map[int64]FoodsDetail
		DetailsFromUsers   map[int64]FoodsDetail
		ImagesFromSources  map[int64]FoodsImage
		ImagesFromUsers    map[int64]FoodsImage
		Locales            map[int64]string
	}
	FoodsDetail struct {
		Kcal float64
	}
	FoodsImage struct {
		Type   string
		Width  int64
		Height int64
		URI    string
	}
)
