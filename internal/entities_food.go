package internal

type (
	Food struct {
		ID                 int64
		DetailsFromSources map[ /*source_id*/ int64]FoodDetail
		DetailsFromUsers   map[ /*user_id*/ int64]FoodDetail
		ImagesFromSources  map[ /*source_id*/ int64][]FoodImage
		ImagesFromUsers    map[ /*user_id*/ int64][]FoodImage
		Locales            map[ /*lang_id*/ int64]Locale
	}
	FoodDetail struct {
		KCal float64
	}
	FoodImage struct {
		Kind string
		URI  string
	}
)
