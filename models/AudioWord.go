package models

type AudioWord struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	VideoAudioID uint   `gorm:"index" json:"video_audio_id"` // Qaysi audioga tegishli ekanligi (Foreign Key)
	English      string `gorm:"type:text" json:"english"`
	Uzbek        string `gorm:"type:text" json:"uzbek"`
	Path         string // ← bu yerda bo'sh joy bilan saqlanган

}
