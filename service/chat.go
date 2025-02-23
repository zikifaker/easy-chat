package service

import (
	"context"
	"easy-chat/agents"
	"easy-chat/agents/llms"
	"easy-chat/agents/llms/qwen"
	"easy-chat/agents/memory"
	"easy-chat/agents/toolkit"
	"easy-chat/agents/toolkit/exa"
	"easy-chat/config"
	"easy-chat/consts"
	"easy-chat/dao"
	"easy-chat/request"
	"errors"
	"fmt"
	"strings"
)

const (
	ModeNormal = "normal"
	ModeAgent  = "agent"
)

var ErrInvalidMode = errors.New("invalid mode")

func HandleChat(ctx context.Context, request *request.ChatRequest) error {
	var result string
	var err error

	switch request.Mode {
	case ModeNormal:
		result, err = handleNormalChat(ctx, request)
		if err != nil {
			return err
		}
	case ModeAgent:
		result, err = handleAgentChat(ctx, request)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidMode, request.Mode)
	}

	if err := dao.SaveChatHistory(request, []memory.Message{
		{Role: memory.MessageRoleUser, Content: request.Query},
		{Role: memory.MessageRoleAI, Content: result},
	}); err != nil {
		return err
	}

	return nil
}

func handleNormalChat(ctx context.Context, request *request.ChatRequest) (string, error) {
	cfg := config.Get()
	llm, err := qwen.New(
		qwen.WithModelName(request.Model),
		qwen.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		return "", err
	}

	prompt, err := buildPrompt(request)
	if err != nil {
		return "", err
	}

	streamFunc, exists := ctx.Value(consts.KeyStreamFunc).(llms.StreamFunc)
	if !exists {
		return "", fmt.Errorf("%w: %s", consts.ErrInvalidContextKey, consts.KeyStreamFunc)
	}

	result, err := llm.GenerateContent(ctx, prompt, llms.WithStreamFunc(streamFunc))
	if err != nil {
		return "", err
	}

	return result, nil
}

func handleAgentChat(ctx context.Context, request *request.ChatRequest) (string, error) {
	cfg := config.Get()
	llm, err := qwen.New(
		qwen.WithModelName(request.Model),
		qwen.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		return "", err
	}

	searchTool, err := exa.NewSearchTool(cfg.APIKey.Exa)
	if err != nil {
		return "", err
	}

	tools := []toolkit.Tool{searchTool}

	agent, err := agents.NewAgent(llm, tools)
	if err != nil {
		return "", err
	}

	result, err := agent.Execute(ctx, request)
	if err != nil {
		return "", err
	}

	return result, nil
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
