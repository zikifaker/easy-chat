package router

import (
	"easy-chat/controller"
	"easy-chat/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	r.POST("/api/login", controller.UserLoginAPI)
	r.POST("/api/register", controller.UserRegisterAPI)

	r.Use(middleware.AuthMiddleware())

	r.POST("/api/chat-session", controller.CreateChatSessionAPI)
	r.GET("/api/chat-session/:username", controller.GetUserChatSessionAPI)
	r.DELETE("/api/chat-session/:session_id", controller.DeleteChatSessionAPI)
	r.GET("/api/chat-history/:session_id", controller.GetChatHistoryAPI)
	r.POST("/api/chat", middleware.RetrieveMiddleware, controller.ChatAPI)

	return r
}
