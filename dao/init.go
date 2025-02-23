package dao

import (
	"easy-chat/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func Init() error {
	if err := initDB(); err != nil {
		return err
	}
	return nil
}

func initDB() error {
	dsn := buildDSN()
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

func buildDSN() string {
	cfg := config.Get()
	port := cfg.DataBase.Mysql.Port
	username := cfg.DataBase.Mysql.Username
	password := cfg.DataBase.Mysql.Password
	return username + ":" + password + "@tcp(127.0.0.1:" + port + ")/chat?charset=utf8mb4&parseTime=True&loc=Local"
}
