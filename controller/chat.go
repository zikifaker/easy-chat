package controller

import (
	"context"
	"easy-chat/agents/llms"
	"easy-chat/consts"
	"easy-chat/request"
	"easy-chat/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ChatAPI(c *gin.Context) {
	var request request.ChatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, consts.KeyStreamFunc, buildSSECallback(c))

	err := service.Chat(ctx, &request)
	if err != nil {
		c.SSEvent(consts.SSEventError, err.Error())
		c.Writer.Flush()
		return
	}
}

func buildSSECallback(c *gin.Context) llms.StreamFunc {
	return func(ctx context.Context, chunk []byte) error {
		c.SSEvent(consts.SSEventResult, string(chunk))
		c.Writer.Flush()
		return nil
	}
}
