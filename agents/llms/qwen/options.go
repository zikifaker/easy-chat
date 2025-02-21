package qwen

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
