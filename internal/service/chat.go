package service

import (
	"context"
	"easy-chat/config"
	"easy-chat/internal/agents"
	"easy-chat/internal/agents/llms"
	"easy-chat/internal/agents/llms/qwen"
	"easy-chat/internal/agents/memory"
	"easy-chat/internal/agents/toolkit"
	"easy-chat/internal/agents/toolkit/exa"
	"easy-chat/internal/consts"
	"easy-chat/internal/dao"
	"easy-chat/internal/request"
	"errors"
	"fmt"
	"strings"
)

const (
	ModeNormal = "normal"
	ModeAgent  = "agent"
)

var (
	ErrInvalidMode = errors.New("invalid mode")
)

func Chat(ctx context.Context, request *request.ChatRequest) error {
	switch request.Mode {
	case ModeNormal:
		return handleNormalChat(ctx, request)
	case ModeAgent:
		return handleAgentChat(ctx, request)
	default:
		return fmt.Errorf("%w: %s", ErrInvalidMode, request.Mode)
	}
}

func handleNormalChat(ctx context.Context, request *request.ChatRequest) error {
	cfg := config.Get()
	llm, err := qwen.New(
		qwen.WithModelName(request.Model),
		qwen.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		return err
	}

	streamFunc, exists := ctx.Value(consts.KeyStreamFunc).(llms.StreamFunc)
	if !exists {
		return fmt.Errorf("%w: %s", consts.ErrInvalidContextKey, consts.KeyStreamFunc)
	}

	prompt, err := buildPrompt(request)
	if err != nil {
		return err
	}

	result, err := llm.GenerateContent(ctx, prompt, llms.WithStreamFunc(streamFunc))
	if err != nil {
		return err
	}

	if err := dao.SaveChatHistory(request, []memory.Message{
		{Role: memory.MessageRoleUser, Content: request.Query},
		{Role: memory.MessageRoleAI, Content: result},
	}); err != nil {
		return err
	}

	return nil
}

func handleAgentChat(ctx context.Context, request *request.ChatRequest) error {
	cfg := config.Get()
	llm, err := qwen.New(
		qwen.WithModelName(request.Model),
		qwen.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		return err
	}

	searchTool, err := exa.NewSearchTool(cfg.APIKey.Exa)
	if err != nil {
		return err
	}

	tools := []toolkit.Tool{searchTool}

	agent, err := agents.NewAgent(llm, tools)
	if err != nil {
		return err
	}

	result, err := agent.Execute(ctx, request)
	if err != nil {
		return err
	}

	if err := dao.SaveChatHistory(request, []memory.Message{
		{Role: memory.MessageRoleUser, Content: request.Query},
		{Role: memory.MessageRoleAI, Content: result},
	}); err != nil {
		return err
	}

	return nil
}

func buildPrompt(request *request.ChatRequest) (string, error) {
	var prompt strings.Builder

	chatHistories, err := dao.GetChatHistoryBySessionID(request.SessionID)
	if err != nil {
		return "", err
	}

	prompt.WriteString("Chat History:\n")
	for _, chatHistory := range chatHistories {
		prompt.WriteString(chatHistory.MessageType + ": " + chatHistory.Content + "\n")
	}

	prompt.WriteString("User Query:\n")
	prompt.WriteString(request.Query)

	return prompt.String(), nil
}
