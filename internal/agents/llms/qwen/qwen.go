package qwen

import (
	"context"
	"easy-chat/internal/agents/llms"
	"errors"
)

var _ llms.LLM = (*LLM)(nil)

var (
	ErrMissedModelName = errors.New("missed model name")
)

type LLM struct {
	client    *client
	ModelName string
}

func New(options ...Option) (*LLM, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	if opts.ModelName == "" {
		return nil, ErrMissedModelName
	}

	client, err := newClient(opts.APIKey)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client:    client,
		ModelName: opts.ModelName,
	}, nil
}

func (l *LLM) GenerateContent(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	opts := &llms.CallOptions{}

	for _, opt := range options {
		opt(opts)
	}

	chatRequest := &ChatRequest{
		Model: l.ModelName,
		Input: Input{
			Messages: []Message{{Role: "user", Content: prompt}},
		},
		Parameters: Parameters{
			ResultFormat:      "message",
			IncrementalOutput: opts.StreamFunc != nil,
		},
		StreamFunc: opts.StreamFunc,
	}

	result, err := l.client.createChat(ctx, chatRequest)
	if err != nil {
		return "", err
	}

	content := result.Output.Choices[0].Message.Content

	return content, nil
}
