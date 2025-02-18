package controller

import (
	"context"
	"easy-chat/internal/agents/llms"
	"easy-chat/internal/consts"
	"easy-chat/internal/request"
	"easy-chat/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ChatAPI(c *gin.Context) {
	var req request.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, consts.KeyStreamFunc, buildSSECallback(c))

	err := service.Chat(ctx, &req)
	if err != nil {
		c.SSEvent("error", err.Error())
		c.Writer.Flush()
		return
	}
}

func buildSSECallback(c *gin.Context) llms.StreamFunc {
	return func(ctx context.Context, chunk []byte) error {
		c.SSEvent("result", string(chunk))
		c.Writer.Flush()
		return nil
	}
}
