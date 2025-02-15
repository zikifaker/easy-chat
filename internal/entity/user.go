package entity

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	CreateTime time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP"`
	UpDateTime time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;update:CURRENT_TIMESTAMP"`
	Username   string    `gorm:"type:varchar(50);not_null;unique"`
	Email      string    `gorm:"type:varchar(100);not_null;unique"`
	Password   string    `gorm:"type:varchar(100);not_null"`
	LastLogin  time.Time `gorm:"type:datetime;default:null"`
}

func (User) TableName() string {
	return "user"
}
