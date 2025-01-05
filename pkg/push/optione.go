package push

type PushOption interface {
	Apply(option *PushOptions)
}

// PushOptions 用于设置发送单个消息选项的结构体
type PushOptions struct {
	DryRun      bool `json:"dry_run,omitempty"`
	Development bool `json:"development,omitempty"`
}

func (s *PushOptions) Apply(option *PushOptions) {
	if s.DryRun {
		option.DryRun = true
	}
	if s.Development {
		option.Development = true
	}
}

func (s *PushOptions) ApplyOptions(opts []PushOption) *PushOptions {
	for _, opt := range opts {
		opt.Apply(s)
	}
	return s
}
