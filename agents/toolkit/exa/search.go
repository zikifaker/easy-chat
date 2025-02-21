package exa

import (
	"bytes"
	"context"
	"easy-chat/agents/toolkit"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var _ toolkit.Tool = (*SearchTool)(nil)

var (
	ErrMissedAPIKey    = errors.New("missed api key")
	ErrFailedToRequest = errors.New("failed to request")
)

type SearchTool struct {
	APIKey string
}

type SearchRequest struct {
	Query         string `json:"query"`
	UseAutoprompt bool   `json:"useAutoprompt"`
	Type          string `json:"type"`
	NumResults    int    `json:"numResults"`
	Contents      struct {
		Text bool `json:"text"`
	} `json:"contents"`
}

type SearchResponse struct {
	Results []struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		Text   string `json:"text"`
	} `json:"results"`
	SearchType string `json:"searchType"`
}

func NewSearchTool(apiKey string) (*SearchTool, error) {
	if apiKey == "" {
		return nil, ErrMissedAPIKey
	}

	return &SearchTool{
		APIKey: apiKey,
	}, nil
}

func (s *SearchTool) Name() string {
	return "Exa Search API"
}

func (s *SearchTool) Description() string {
	return "Search the web with an Exa prompt-engineered query."
}

func (s *SearchTool) Execute(ctx context.Context, input string) (string, error) {
	searchResponse, err := s.search(ctx, input)
	if err != nil {
		return "", err
	}

	return buildSearchResult(searchResponse), nil
}

func (s *SearchTool) search(ctx context.Context, query string) (*SearchResponse, error) {
	url := "https://api.exa.ai/search"

	searchRequest := &SearchRequest{
		Query:         query,
		UseAutoprompt: true,
		Type:          "auto",
		NumResults:    10,
		Contents: struct {
			Text bool `json:"text"`
		}{
			Text: true,
		},
	}

	reqBody, err := json.Marshal(searchRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	s.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
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

	var searchResponse SearchResponse
	if err := json.Unmarshal(respBody, &searchResponse); err != nil {
		return nil, err
	}

	return &searchResponse, nil
}

func (s *SearchTool) setHeaders(req *http.Request) {
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.APIKey)
}

func buildSearchResult(searchResponse *SearchResponse) string {
	var result strings.Builder

	for _, res := range searchResponse.Results {
		result.WriteString("Title: " + res.Title + "\n")
		result.WriteString("Author: " + res.Author + "\n")
		result.WriteString("Text: " + res.Text + "\n\n")
	}

	return result.String()
}
