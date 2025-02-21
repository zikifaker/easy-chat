package dao

import (
	"easy-chat/entity"
	"github.com/google/uuid"
)

func CreateChatSession(username string) (string, error) {
	sessionID := uuid.New().String()
	user, err := GetUserByUsername(username)
	if err != nil {
		return "", nil
	}

	session := &entity.ChatSession{
		SessionID: sessionID,
		UserID:    user.ID,
	}
	result := db.Create(session)
	if result.Error != nil {
		return "", result.Error
	}

	return sessionID, nil
}

func DeleteChatSession(sessionID string) error {
	return db.Where("session_id = ?", sessionID).Delete(&entity.ChatSession{}).Error
}

func GetChatSessionByUsername(username string) ([]*entity.ChatSession, error) {
	user, err := GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	var sessions []*entity.ChatSession
	result := db.Where("user_id = ?", user.ID).Find(&sessions)
	if result.Error != nil {
		return nil, result.Error
	}
	return sessions, nil
}
