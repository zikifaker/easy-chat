package llms

import "context"

type StreamFunc func(ctx context.Context, chunk []byte) error

// CallOptions Common parameters while calling LLM
type CallOptions struct {
	StreamFunc StreamFunc
}

type CallOption func(*CallOptions)

func WithStreamFunc(streamFunc StreamFunc) CallOption {
	return func(o *CallOptions) {
		o.StreamFunc = streamFunc
	}
}
