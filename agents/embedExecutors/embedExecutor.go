package embedExecutors

import "context"

type EmbedExecutor interface {
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
}
