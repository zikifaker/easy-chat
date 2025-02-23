package main

import (
	"easy-chat/config"
	"easy-chat/dao"
	"easy-chat/router"
	"easy-chat/service/mq"
	"log"
)

const (
	configFilePath = "config-dev.yaml"
)

func main() {
	if err := config.Init(configFilePath); err != nil {
		log.Fatal(err)
	}

	if err := dao.Init(); err != nil {
		log.Fatal(err)
	}

	if err := mq.Init(); err != nil {
		log.Fatal(err)
	}

	r := router.SetupRouter()
	if err := r.Run(":8088"); err != nil {
		log.Fatal(err)
	}
}
