package domain

import "time"

type Item struct {
	ID        string
	Title     string
	Username  string
	Password  string
	Website   string
	Notes     string
	Category  string
	Favorite  bool
	Tags      []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateItemInput struct {
	Title    string
	Username string
	Password string
	Website  string
	Notes    string
	Category string
	Favorite bool
	Tags     []string
}

type UpdateItemInput struct {
	ID       string
	Title    string
	Username string
	Password string
	Website  string
	Notes    string
	Category string
	Favorite bool
	Tags     []string
}

type ListItemsFilter struct {
	Keyword  string
	Tag      string
	Favorite *bool
	Category string
}
