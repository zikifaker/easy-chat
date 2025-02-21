package controller

import (
	"easy-chat/dao"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetChatHistoryAPI(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "miss parameter 'session_id'"})
		return
	}

	chatHistories, err := dao.GetChatHistoryBySessionID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response = make([]struct {
		MessageType string `json:"message_type"`
		Content     string `json:"content"`
	}, len(chatHistories))

	for i := 0; i < len(chatHistories); i++ {
		response[i].MessageType = chatHistories[i].MessageType
		response[i].Content = chatHistories[i].Content
	}

	c.JSON(http.StatusOK, response)
}
