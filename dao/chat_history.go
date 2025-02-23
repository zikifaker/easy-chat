package dao

import (
	"easy-chat/agents/memory"
	"easy-chat/entity"
	"easy-chat/request"
)

func GetChatHistoryBySessionID(sessionID string) ([]*entity.ChatHistory, error) {
	var chatHistories []*entity.ChatHistory

	result := db.Where("session_id = ?", sessionID).Find(&chatHistories)
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
		result := db.Create(chatHistory)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}
