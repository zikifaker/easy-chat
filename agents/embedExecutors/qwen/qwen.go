package qwen

import (
	"context"
	"easy-chat/agents/embedExecutors"
	"errors"
)

var _ embedExecutors.EmbedExecutor = (*EmbedExecutor)(nil)

var (
	ErrMissedModelName = errors.New("missed model name")
)

type EmbedExecutor struct {
	client    *client
	ModelName string
}

func New(options ...Option) (*EmbedExecutor, error) {
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

	return &EmbedExecutor{
		client:    client,
		ModelName: opts.ModelName,
	}, nil
}

func (e *EmbedExecutor) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddingRequest := &EmbeddingRequest{
		Model: e.ModelName,
		Input: Input{
			Texts: texts,
		},
	}

	result, err := e.client.createEmbedding(ctx, embeddingRequest)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, 0, len(texts))
	for _, embedding := range result.Output.Embeddings {
		embeddings = append(embeddings, embedding.Embedding)
	}

	return embeddings, nil
}
