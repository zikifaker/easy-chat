package qwen

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrMissedAPIKey    = errors.New("missed api key")
	ErrFailedToRequest = errors.New("failed to request")
)

type client struct {
	httpClient *http.Client
	apiKey     string
}

func newClient(apiKey string) (*client, error) {
	if apiKey == "" {
		return nil, ErrMissedAPIKey
	}

	return &client{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}, nil
}

type Input struct {
	Texts []string `json:"texts"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input Input  `json:"input"`
}

type EmbeddingResponse struct {
	RequestID string `json:"request_id"`
	Output    struct {
		Embeddings []struct {
			Embedding []float32 `json:"embedding"`
			TextIndex int       `json:"text_index"`
		} `json:"embeddings"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *client) createEmbedding(ctx context.Context, embeddingRequest *EmbeddingRequest) (*EmbeddingResponse, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/services/embeddings/text-embedding/text-embedding"

	reqBody, err := json.Marshal(embeddingRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w with status code %d", ErrFailedToRequest, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var embeddingResponse EmbeddingResponse
	if err := json.Unmarshal(respBody, &embeddingResponse); err != nil {
		return nil, err
	}

	return &embeddingResponse, nil
}

func (c *client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}
