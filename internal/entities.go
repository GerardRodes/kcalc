package internal

type (
	User struct {
		ID            int64
		Role          string
		Email         string
		VerifiedEmail bool
		Lang          int64
		Family        *Family
	}
	Family struct {
		ID    int64
		Name  string
		Users []User
	}
	Source struct {
		ID   int64
		Name string
	}
	Session struct {
		User User
	}
)
