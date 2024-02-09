package internal

type Cooking struct {
	ID            string
	Name          string
	GAfterCooking float64
	Foods         []CookingFood
	SubCookings   []SubCooking
}

type CookingFood struct {
	Food Food
	G    float64
}

type SubCooking struct {
	Cooking Cooking
	G       float64
}
