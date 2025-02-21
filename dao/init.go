package dao

import (
	"context"
	"easy-chat/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	cache *redis.Client
)

func GetDBClient() *gorm.DB {
	return db
}

func GetCacheClient() *redis.Client {
	return cache
}

func Init() error {
	if err := initDB(); err != nil {
		return err
	}
	if err := initCache(); err != nil {
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

func initCache() error {
	cache = redis.NewClient(&redis.Options{
		Addr: "localhost:" + config.Get().DataBase.Redis.Port,
		DB:   0,
	})

	ctx := context.Background()
	if _, err := cache.Ping(ctx).Result(); err != nil {
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
