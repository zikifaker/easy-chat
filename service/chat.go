package service

import (
	"bytes"
	"context"
	"easy-chat/agents"
	qwen2 "easy-chat/agents/embedExecutors/qwen"
	"easy-chat/agents/llms"
	"easy-chat/agents/llms/qwen"
	"easy-chat/agents/memory"
	"easy-chat/agents/toolkit"
	"easy-chat/agents/toolkit/exa"
	"easy-chat/config"
	"easy-chat/consts"
	"easy-chat/dao"
	"easy-chat/request"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
)

const (
	ModeNormal = "normal"
	ModeAgent  = "agent"
)

const qaIndexPrefix = "qa:"

var (
	ErrInvalidMode         = errors.New("invalid mode")
	ErrFailedToSavedQAPair = errors.New("failed to save QA pair")
)

func Chat(ctx context.Context, request *request.ChatRequest) error {
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

	if err := saveQAPair(ctx, request.Query, result); err != nil {
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

func saveQAPair(ctx context.Context, query, answer string) error {
	cfg := config.Get()
	embedExecutor, err := qwen2.New(
		qwen2.WithModelName(qwen2.ModelNameTextEmbeddingV2),
		qwen2.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		return err
	}

	embeddings, err := embedExecutor.GenerateEmbeddings(ctx, []string{query})
	if err != nil {
		return err
	}
	queryEmbedding := embeddings[0]

	queryEmbeddingBytes, err := convertFloat32SliceToBytes(queryEmbedding)
	if err != nil {
		return err
	}

	qaKey := qaIndexPrefix + uuid.New().String()
	client := dao.GetCacheClient()
	if err := client.HSet(ctx, qaKey, map[string]interface{}{
		"query":     query,
		"answer":    answer,
		"embedding": queryEmbeddingBytes,
	}).Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrFailedToSavedQAPair, err)
	}

	return nil
}

func convertFloat32SliceToBytes(embedding []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, embedding)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
