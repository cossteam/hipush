package adapter

import (
	"encoding/json"
	"fmt"

	v1 "github.com/cossteam/hipush/api/http/v1"
	"github.com/cossteam/hipush/pkg/push"
)

var _ push.Message = &MessageAdapter{}

// MessageAdapter 将 MessageRequest 适配为 Message 接口
type MessageAdapter struct {
	req *v1.MessageRequest
}

// NewMessageAdapter 创建消息适配器
func NewMessageAdapter(req *v1.MessageRequest) *MessageAdapter {
	return &MessageAdapter{req: req}
}

func (m *MessageAdapter) GetTitle() string {
	if m.req.Title == nil {
		return ""
	}
	return *m.req.Title
}

func (m *MessageAdapter) GetContent() string {
	if m.req.Content == nil {
		return ""
	}
	return *m.req.Content
}

func (m *MessageAdapter) GetCategory() string {
	if m.req.Category == nil {
		return ""
	}
	return *m.req.Category
}

func (m *MessageAdapter) GetPriority() push.Priority {
	if m.req.Priority == nil {
		return push.Normal
	}
	return push.Priority(*m.req.Priority)
}

func (m *MessageAdapter) GetNotifyType() int {
	if m.req.NotifyType == nil {
		return 0
	}
	return *m.req.NotifyType
}

func (m *MessageAdapter) GetIcon() string {
	if m.req.Icon == nil {
		return ""
	}
	return *m.req.Icon
}

func (m *MessageAdapter) GetSound() string {
	if m.req.Sound == nil {
		return ""
	}
	return *m.req.Sound
}

func (m *MessageAdapter) GetTTL() int {
	if m.req.TTL == nil {
		return 0
	}
	return *m.req.TTL
}

func (m *MessageAdapter) GetForeground() bool {
	if m.req.Foreground == nil {
		return false
	}
	return *m.req.Foreground
}

func (m *MessageAdapter) GetClickAction() push.ClickAction {
	if m.req.ClickAction == nil {
		return push.ClickAction{}
	}
	return push.ClickAction{
		Action:     push.ClickActionAction(m.req.ClickAction.Action),
		Activity:   m.req.ClickAction.Activity,
		Parameters: m.req.ClickAction.Parameters,
		URL:        m.req.ClickAction.URL,
	}
}

func (m *MessageAdapter) GetData() map[string]interface{} {
	if m.req.Data == nil {
		return nil
	}
	return *m.req.Data
}

func (m *MessageAdapter) String() string {
	marshal, err := json.Marshal(m.req)
	if err != nil {
		fmt.Println("ee" + err.Error())
		return ""
	}
	return string(marshal)
}
