package qwen

import (
	"bufio"
	"bytes"
	"context"
	"easy-chat/internal/agents/llms"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var (
	ErrMissedAPIKey                = errors.New("missed api key")
	ErrFailedToRequest             = errors.New("failed to request")
	ErrFailedToParseStreamResponse = errors.New("failed to parse stream response")
	ErrWhileCallingStreamFunc      = errors.New("error while calling stream function")
)

type client struct {
	httpClient *http.Client
	apiKey     string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Input struct {
	Messages []Message `json:"messages"`
}

type Parameters struct {
	ResultFormat      string `json:"result_format"`
	IncrementalOutput bool   `json:"incremental_output"`
}

type ChatRequest struct {
	Model      string     `json:"model"`
	Input      Input      `json:"input"`
	Parameters Parameters `json:"parameters"`

	StreamFunc llms.StreamFunc `json:"-"`
}

type ChatResponse struct {
	RequestID string `json:"request_id"`
	Output    struct {
		Choices []struct {
			FinishReason string  `json:"finish_reason"`
			Message      Message `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
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

func (c *client) createChat(ctx context.Context, chatRequest *ChatRequest) (*ChatResponse, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"

	reqBody, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, chatRequest.StreamFunc != nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w with status code %d", ErrFailedToRequest, resp.StatusCode)
	}

	if chatRequest.StreamFunc != nil {
		return handleStreamResponse(ctx, chatRequest, resp)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var chatResponse ChatResponse
	if err := json.Unmarshal(respBody, &chatResponse); err != nil {
		return nil, err
	}

	return &chatResponse, nil
}

func (c *client) setHeaders(req *http.Request, isStream bool) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if isStream {
		req.Header.Set("X-DashScope-SSE", "enable")
	}
}

func handleStreamResponse(ctx context.Context, chatRequest *ChatRequest, resp *http.Response) (*ChatResponse, error) {
	const bufferSize = 10
	dataChan := make(chan *ChatResponse, bufferSize)

	go func() {
		defer close(dataChan)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data:") {
				continue
			}

			line = strings.TrimPrefix(line, "data:")
			var partialResponse ChatResponse
			if err := json.Unmarshal([]byte(line), &partialResponse); err != nil {
				log.Printf("%v: %v", ErrFailedToParseStreamResponse, err)
			} else {
				dataChan <- &partialResponse
			}
		}
	}()

	var completeResponse *ChatResponse
	var contentBuilder strings.Builder
	for partialResponse := range dataChan {
		content := partialResponse.Output.Choices[0].Message.Content
		if err := callStreamFunc(ctx, chatRequest.StreamFunc, content); err != nil {
			log.Printf("%v: %v", ErrWhileCallingStreamFunc, err)
		}

		contentBuilder.WriteString(content)
		completeResponse = partialResponse
	}

	completeResponse.Output.Choices[0].Message.Content = contentBuilder.String()

	return completeResponse, nil
}

func callStreamFunc(ctx context.Context, streamFunc llms.StreamFunc, content string) error {
	if streamFunc == nil || content == "" {
		return nil
	}
	return streamFunc(ctx, []byte(content))
}
