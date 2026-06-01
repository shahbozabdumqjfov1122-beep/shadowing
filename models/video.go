package models

import "time"

type Video struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Thumbnail   string `gorm:"size:500"` // Muqova surati yo'li
	VideoPath   string `gorm:"size:500"` // .mp4 fayl yo'li

	// Zanjirsimon o'chirish o'rnatilgan aloqador modellar
	Audios     []VideoAudio    `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`
	Recordings []UserRecording `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`

	Level int `gorm:"not null"`
	Views int `gorm:"default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
