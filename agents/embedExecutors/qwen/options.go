package qwen

const (
	ModelNameTextEmbeddingV1 = "text-embedding-v1"
	ModelNameTextEmbeddingV2 = "text-embedding-v2"
)

type Options struct {
	ModelName string
	APIKey    string
}

type Option func(*Options)

func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

func WithModelName(model string) Option {
	return func(o *Options) {
		o.ModelName = model
	}
}
