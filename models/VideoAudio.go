package models

import "time"

type VideoAudio struct {
	ID      uint `gorm:"primaryKey"`
	VideoID uint `gorm:"index"` // qidiruvlar tezroq ishlashi uchun indeks qo'ydik

	Path string
	Text string `gorm:"type:text"`

	// constraint:OnDelete:CASCADE qo'shildi.
	// Bu - admin paneldan audioni o'chirganda unga tegishli so'zlar bazada yetim qolib,
	// joy egallab yotmasligi uchun ularni ham avtomat o'chirib yuboradi.
	Words []AudioWord `gorm:"foreignKey:VideoAudioID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
}
