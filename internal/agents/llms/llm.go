package llms

import "context"

type LLM interface {
	GenerateContent(ctx context.Context, prompt string, options ...CallOption) (string, error)
}
