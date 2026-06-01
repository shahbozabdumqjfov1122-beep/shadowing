package models

import "time"

type Room struct {
	ID        uint
	Name      string
	Image     string
	CreatedAt time.Time
}
