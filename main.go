package main

import (
	"easy-chat/config"
	"easy-chat/internal/dao"
	"easy-chat/internal/router"
	"log"
)

func main() {
	if err := config.Init("config-dev.yaml"); err != nil {
		log.Fatal(err)
	}

	if err := dao.Init(); err != nil {
		log.Fatal(err)
	}

	r := router.SetupRouter()
	if err := r.Run(":8088"); err != nil {
		log.Fatal(err)
	}
}
