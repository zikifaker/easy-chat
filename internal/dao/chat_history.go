package dao

import (
	"easy-chat/internal/agents/memory"
	"easy-chat/internal/entity"
	"easy-chat/internal/request"
)

func GetChatHistoryBySessionID(sessionID string) ([]*entity.ChatHistory, error) {
	var chatHistories []*entity.ChatHistory

	result := DB.Where("session_id = ?", sessionID).Find(&chatHistories)
	if result.Error != nil {
		return nil, result.Error
	}

	return chatHistories, nil
}

func SaveChatHistory(chatRequest *request.ChatRequest, messages []memory.Message) error {
	user, err := GetUserByUsername(chatRequest.Username)
	if err != nil {
		return err
	}

	for _, message := range messages {
		chatHistory := &entity.ChatHistory{
			UserID:      user.ID,
			SessionID:   chatRequest.SessionID,
			MessageType: message.Role,
			Content:     message.Content,
		}
		result := DB.Create(chatHistory)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
