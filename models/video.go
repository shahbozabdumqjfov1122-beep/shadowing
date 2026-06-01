package models

import "time"

type Video struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:255;not null"`
	Description string `gorm:"type:text"`
	Thumbnail   string `gorm:"size:500"`
	VideoPath   string `gorm:"size:500"`

	// constraint:OnDelete:CASCADE qo'shildi.
	// Agar admin paneldan butunlay bitta VIDEONI o'chirib tashlasangiz,
	// unga tegishli barcha audiolarni va o'sha audiolarga tegishli barcha so'zlarni
	// bazaning o'zi avtomat ravishda "zanjirsimon" tozalab yuboradi (Baza "axlat"ga to'lmaydi).
	Audios []VideoAudio `gorm:"foreignKey:VideoID;constraint:OnDelete:CASCADE"`

	Level int `gorm:"not null"`
	Views int `gorm:"default:0"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
