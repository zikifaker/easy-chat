package service

import (
	"context"
	"easy-chat/config"
	"easy-chat/internal/agents/llms"
	"easy-chat/internal/agents/llms/qwen"
	"easy-chat/internal/agents/memory"
	"easy-chat/internal/consts"
	"easy-chat/internal/dao"
	"easy-chat/internal/request"
	"errors"
	"fmt"
	"log"
	"strings"
)

var (
	ErrInvalidContextKey   = errors.New("invalid context key")
	ErrFailedToBuildPrompt = errors.New("failed to build prompt")
)

func Chat(ctx context.Context, request *request.ChatRequest) error {
	cfg := config.Get()

	llm, err := qwen.New(
		qwen.WithModelName(request.Model),
		qwen.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		return err
	}

	streamFunc, exists := ctx.Value(consts.ContextKeyStreamFunc).(llms.StreamFunc)
	if !exists {
		return fmt.Errorf("%w: %s", ErrInvalidContextKey, consts.ContextKeyStreamFunc)
	}

	prompt, err := buildPrompt(request)
	if err != nil {
		return err
	}

	result, err := llm.GenerateContent(ctx, prompt, llms.WithStreamFunc(streamFunc))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToBuildPrompt, err)
	}

	log.Println("result: ", result)

	err = dao.SaveChatHistory(request, []memory.Message{
		{Role: memory.MessageRoleUser, Content: request.Query},
		{Role: memory.MessageRoleAI, Content: result},
	})
	if err != nil {
		return err
	}

	return nil
}

func buildPrompt(request *request.ChatRequest) (string, error) {
	var result strings.Builder

	chatHistories, err := dao.GetChatHistoryBySessionID(request.SessionID)
	if err != nil {
		return "", err
	}

	result.WriteString("Chat History:\n")
	for _, chatHistory := range chatHistories {
		result.WriteString(chatHistory.MessageType + ": " + chatHistory.Content + "\n")
	}

	result.WriteString("User Query:\n")
	result.WriteString(request.Query)

	return result.String(), nil
}
