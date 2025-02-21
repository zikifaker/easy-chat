package middleware

import (
	"bytes"
	"context"
	"easy-chat/agents/embedExecutors/qwen"
	"easy-chat/config"
	"easy-chat/consts"
	"easy-chat/dao"
	"easy-chat/request"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var (
	ErrFailedToParseResult          = errors.New("failed to parse result")
	ErrNoSimilarQAFound             = errors.New("no similar QA found")
	ErrFailedToParseFirstResult     = errors.New("failed to parse first result")
	ErrFailedToParseExtraAttributes = errors.New("failed to parse extra_attributes")
	ErrAnswerFieldISNotAString      = errors.New("answer field is not a string")
)

const (
	maxKNNSearchResult = 1
	minSearchScore     = 0.8
)

func RetrieveMiddleware(c *gin.Context) {
	var req request.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Set(string(consts.KeyChatRequest), req)

	answer, found := retrieveSimilarQA(c.Request.Context(), &req)
	if found {
		pushAnswer(c, answer)
		c.Abort()
		return
	}

	c.Next()
}

func retrieveSimilarQA(ctx context.Context, request *request.ChatRequest) (string, bool) {
	cfg := config.Get()
	embedExecutor, err := qwen.New(
		qwen.WithModelName(qwen.ModelNameTextEmbeddingV2),
		qwen.WithAPIKey(cfg.APIKey.Qwen),
	)
	if err != nil {
		log.Printf("%v", err)
		return "", false
	}

	embeddings, err := embedExecutor.GenerateEmbeddings(ctx, []string{request.Query})
	if err != nil {
		log.Printf("%v", err)
		return "", false
	}
	queryEmbedding := embeddings[0]

	queryEmbeddingBytes, err := convertFloat32SliceToBytes(queryEmbedding)
	if err != nil {
		log.Printf("%v", err)
		return "", false
	}

	query := fmt.Sprintf(
		"*=>[KNN %d @embedding $query_vector AS score]",
		maxKNNSearchResult,
	)
	args := []interface{}{
		"FT.SEARCH", "qa_index", query,
		"PARAMS", "2", "query_vector", queryEmbeddingBytes,
		"RETURN", "1", "answer",
		"DIALECT", "2",
	}

	client := dao.GetCacheClient()
	result, err := client.Do(ctx, args...).Result()
	if err != nil {
		log.Printf("%v", err)
		return "", false
	}

	return extractAnswer(result)
}

func convertFloat32SliceToBytes(embedding []float32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, embedding)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func extractAnswer(result interface{}) (string, bool) {
	resMap, ok := result.(map[interface{}]interface{})
	if !ok {
		log.Printf("%v", ErrFailedToParseResult)
		return "", false
	}

	results, ok := resMap["results"].([]interface{})
	if !ok || len(results) == 0 {
		log.Printf("%v", ErrNoSimilarQAFound)
		return "", false
	}

	firstResult, ok := results[0].(map[interface{}]interface{})
	if !ok {
		log.Printf("%v", ErrFailedToParseFirstResult)
		return "", false
	}

	extraAttributes, ok := firstResult["extra_attributes"].(map[interface{}]interface{})
	if !ok {
		log.Printf("%v", ErrFailedToParseExtraAttributes)
		return "", false
	}

	answer, ok := extraAttributes["answer"].(string)
	if !ok {
		log.Printf("%v", ErrAnswerFieldISNotAString)
		return "", false
	}

	return answer, true
}

func pushAnswer(c *gin.Context, answer string) {
	c.SSEvent(consts.SSEventResult, answer)
	c.Writer.Flush()
}
