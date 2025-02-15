package memory

import (
	"context"
)

const (
	MessageRoleAI   = "ai"
	MessageRoleUser = "user"
)

type Message struct {
	Role    string
	Content string
}

type Memory interface {
	AddMessage(ctx context.Context, messages []Message)
	GetMessages(ctx context.Context) []Message
	Clear(ctx context.Context)
}
