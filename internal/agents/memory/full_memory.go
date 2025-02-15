package memory

import (
	"context"
)

var _ Memory = (*FullMemory)(nil)

type FullMemory struct {
	messages []Message
}

func NewFullMemory() Memory {
	return &FullMemory{
		messages: make([]Message, 0),
	}
}

func (f *FullMemory) AddMessage(ctx context.Context, messages []Message) {
	f.messages = append(f.messages, messages...)
}

func (f *FullMemory) GetMessages(ctx context.Context) []Message {
	return f.messages
}

func (f *FullMemory) Clear(ctx context.Context) {
	f.messages = make([]Message, 0)
}
