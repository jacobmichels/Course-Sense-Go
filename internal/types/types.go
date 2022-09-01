package types

type CourseSection struct {
	CourseCode  int
	Department  string
	SectionCode string
	Term        string
	Watchers    []User
}

type User struct {
	Email string
	Phone string
}
