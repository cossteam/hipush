package push

// Priority 消息优先级
type Priority string

const (
	High   Priority = "high"
	Normal Priority = "normal"
	Low    Priority = "low"
)

type ClickActionAction string

const (
	OpenApp      ClickActionAction = "openApp"
	OpenActivity ClickActionAction = "openActivity"
	OpenURL      ClickActionAction = "openURL"
)

type ClickAction struct {
	// Action 点击动作类型。
	Action ClickActionAction `json:"action,omitempty"`

	// Activity 指定打开的 Activity 类名。
	Activity string `json:"activity,omitempty"`

	// URL 指定打开的 URL 地址。
	URL string `json:"url,omitempty"`

	// Parameters 附加参数，可以传递自定义数据。
	Parameters map[string]interface{} `json:"Parameters,omitempty"`
}
