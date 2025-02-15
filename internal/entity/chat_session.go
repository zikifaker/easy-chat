package entity

import "time"

type ChatSession struct {
	SessionID   string    `gorm:"primaryKey;type:char(36)"`
	CreateTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;update:CURRENT_TIMESTAMP"`
	UserID      uint      `gorm:"not_null;index"`
	SessionName string    `gorm:"type:varchar(50)"`
}

func (ChatSession) TableName() string {
	return "chat_session"
}
