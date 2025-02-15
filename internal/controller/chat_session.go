package controller

import (
	"easy-chat/internal/dao"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CreateChatSessionAPI(c *gin.Context) {
	var request struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionID, err := dao.CreateChatSession(request.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"session_id": sessionID})
}

func DeleteChatSessionAPI(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "miss parameter 'session_id'"})
		return
	}

	if err := dao.DeleteChatSession(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "chat session deleted successfully"})
}

func GetUserChatSessionAPI(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "miss parameter 'username'"})
		return
	}

	sessions, err := dao.GetChatSessionByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]struct {
		SessionID   string `json:"session_id"`
		SessionName string `json:"session_name"`
	}, len(sessions))

	for i := 0; i < len(sessions); i++ {
		response[i].SessionID = sessions[i].SessionID
		response[i].SessionName = sessions[i].SessionName
	}

	c.JSON(http.StatusOK, response)
}
