package agents

const defaultMaxStep = 5

type Options struct {
	MaxStep int
}

func GetDefaultOptions() *Options {
	return &Options{
		MaxStep: defaultMaxStep,
	}
}

type Option func(*Options)

func WithMaxStep(maxStep int) Option {
	return func(o *Options) {
		o.MaxStep = maxStep
	}
}
