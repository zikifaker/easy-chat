package entity

import "time"

type ChatHistory struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	CreateTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;update:CURRENT_TIMESTAMP"`
	UserID      uint      `gorm:"not_null;index"`
	SessionID   string    `gorm:"type:char(36);not_null;index"`
	MessageType string    `gorm:"type:varchar(10);not_null"`
	Content     string    `gorm:"type:text"`
}

func (ChatHistory) TableName() string {
	return "chat_history"
}
