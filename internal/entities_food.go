package internal

type (
	Food struct {
		ID             int64
		DetailBySource map[ /*source_id*/ int64]FoodDetail
		DetailByUser   map[ /*user_id*/ int64]FoodDetail
		ImageBySource  map[ /*source_id*/ int64]Image
		ImageByUser    map[ /*user_id*/ int64]Image
		Locales        map[ /*lang_id*/ int64]Locale
	}
	FoodDetail struct {
		KCal float64
	}
	Image struct {
		Kind string
		URI  string
	}
)

func (f Food) Name(langID int64) string {
	if v, ok := f.Locales[langID]; ok && v.Value != "" {
		return v.Value
	}

	if v, ok := f.Locales[LangsID["en"]]; ok && v.Value != "" {
		return v.Value
	}

	for _, v := range f.Locales {
		if v.Value != "" {
			return v.Value
		}
	}

	return ""
}
