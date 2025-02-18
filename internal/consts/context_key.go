package consts

import "errors"

type ContextKey string

var (
	ErrInvalidContextKey = errors.New("invalid context key")
)

const (
	KeyStreamFunc ContextKey = "stream_func"
	KeySSEEvent   ContextKey = "sse_event"
)
