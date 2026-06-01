package models

import "time"

type UserRecording struct {
	ID        uint `gorm:"primaryKey"`
	VideoID   uint
	FilePath  string
	CreatedAt time.Time
}
