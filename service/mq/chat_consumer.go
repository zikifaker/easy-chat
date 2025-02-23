package mq

import (
	"context"
	"easy-chat/agents/llms"
	"easy-chat/consts"
	"easy-chat/request"
	"easy-chat/service"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"sync"
)

var (
	sseConnection      = make(map[string]*gin.Context)
	sseConnectionMutex = &sync.Mutex{}
	ChatRequestDone    = make(chan struct{})
)

var (
	ErrMissedCorrelationID                = errors.New("missed correlation id")
	ErrSSEContextNotFoundForCorrelationID = errors.New("sse context not found for correlation id")
)

const (
	chatRequestQueue       = "chat_queue"
	chatRequestConsumerNum = 5
)

func PublishChatRequest(c *gin.Context, request *request.ChatRequest) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		return err
	}

	correlationID := uuid.New().String()
	addSSEConnection(correlationID, c)

	return rabbitMQChannel.PublishWithContext(
		c,
		"",
		chatRequestQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: correlationID,
			Body:          requestJson,
		},
	)
}

func startChatRequestConsumer() {
	messages, err := rabbitMQChannel.Consume(
		chatRequestQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	for d := range messages {
		var req request.ChatRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			log.Printf("%v", err)
			continue
		}

		correlationID := d.CorrelationId
		if correlationID == "" {
			log.Printf("%v", ErrMissedCorrelationID)
			continue
		}

		sseCtx, exists := getSSEContext(correlationID)
		if !exists {
			log.Printf("%v: %s", ErrSSEContextNotFoundForCorrelationID, correlationID)
			continue
		}

		ctx := sseCtx.Request.Context()
		ctx = context.WithValue(ctx, consts.KeyStreamFunc, buildSSECallback(sseCtx))

		if err := service.HandleChat(ctx, &req); err != nil {
			sseCtx.SSEvent(consts.SSEventError, err.Error())
			sseCtx.Writer.Flush()
		}

		deleteSSEConnection(correlationID)

		ChatRequestDone <- struct{}{}
	}
}

func getSSEContext(correlationID string) (*gin.Context, bool) {
	sseConnectionMutex.Lock()
	sseCtx, exists := sseConnection[correlationID]
	sseConnectionMutex.Unlock()
	return sseCtx, exists
}

func addSSEConnection(correlationID string, c *gin.Context) {
	sseConnectionMutex.Lock()
	sseConnection[correlationID] = c
	sseConnectionMutex.Unlock()
}

func deleteSSEConnection(correlationID string) {
	sseConnectionMutex.Lock()
	delete(sseConnection, correlationID)
	sseConnectionMutex.Unlock()
}

func buildSSECallback(c *gin.Context) llms.StreamFunc {
	return func(ctx context.Context, chunk []byte) error {
		c.SSEvent(consts.SSEventResult, string(chunk))
		c.Writer.Flush()
		return nil
	}
}
