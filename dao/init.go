package dao

import (
	"context"
	"easy-chat/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB          *gorm.DB
	RedisClient *redis.Client
)

func Init() error {
	if err := initDB(); err != nil {
		return err
	}
	if err := initRedisClient(); err != nil {
		return err
	}
	return nil
}

func initDB() error {
	dsn := buildDSN()
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

func initRedisClient() error {
	cfg := config.Get()
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:" + cfg.DataBase.Redis.Port,
		Password: cfg.DataBase.Redis.Password,
		DB:       0,
	})

	ctx := context.Background()
	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
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
