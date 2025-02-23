package controller

import (
	"easy-chat/consts"
	"easy-chat/request"
	"easy-chat/service/mq"
	"github.com/gin-gonic/gin"
)

func ChatAPI(c *gin.Context) {
	setHeaders(c)

	var req request.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.SSEvent(consts.SSEventError, err.Error())
		c.Writer.Flush()
		return
	}

	if err := mq.PublishChatRequest(c, &req); err != nil {
		c.SSEvent(consts.SSEventError, err.Error())
		c.Writer.Flush()
		return
	}

	<-mq.ChatRequestDone
}

func setHeaders(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
}
