package models

// ShrURL contains url alias and URL
type ShrURL struct {
	Alias   string
	URL     string
	UserID  string
	Deleted bool
}

// UserDeleteTask presents user tasks to be deleted
type UserDeleteTask struct {
	UID     string
	Aliases []string
}
