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
	req, exists := c.Get(string(consts.KeyChatRequest))
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "request not found"})
		return
	}

	chatReq, ok := req.(request.ChatRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse request"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, consts.KeyStreamFunc, buildSSECallback(c))

	err := service.Chat(ctx, &chatReq)
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
