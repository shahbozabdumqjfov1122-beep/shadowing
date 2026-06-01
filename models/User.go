package models

import "time"

type User struct {
	ID        uint
	Name      string
	Phone     string
	Photo     string
	IsAdmin   bool
	CreatedAt time.Time
	IsBanned  bool
}
