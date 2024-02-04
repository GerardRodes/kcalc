package internal

type (
	Food struct {
		ID             int64
		DetailBySource map[ /*source_id*/ int64]FoodDetail
		DetailByUser   map[ /*user_id*/ int64]FoodDetail
		ImageBySource  map[ /*source_id*/ int64]FoodImage
		ImageByUser    map[ /*user_id*/ int64]FoodImage
		Locales        map[ /*lang_id*/ int64]Locale
	}
	FoodDetail struct {
		KCal float64
	}
	FoodImage struct {
		Kind string
		URI  string
	}
)
