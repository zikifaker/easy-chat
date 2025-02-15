package agents

import "easy-chat/internal/agents/memory"

const defaultMaxStep = 5

type Options struct {
	Memory  memory.Memory
	MaxStep int
}

func GetDefaultOptions() *Options {
	return &Options{
		Memory:  memory.NewFullMemory(),
		MaxStep: defaultMaxStep,
	}
}

type Option func(*Options)

func WithMemory(memory memory.Memory) Option {
	return func(o *Options) {
		o.Memory = memory
	}
}

func WithMaxStep(maxStep int) Option {
	return func(o *Options) {
		o.MaxStep = maxStep
	}
}
