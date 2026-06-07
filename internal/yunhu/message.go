package yunhu

type SendMessageRequest struct {
	RecvID      string         `json:"recvId"`
	RecvType    string         `json:"recvType"`
	ContentType string         `json:"contentType"`
	Content     SendContent    `json:"content"`
}

type SendContent struct {
	Text     string   `json:"text,omitempty"`
	ImageKey string   `json:"imageKey,omitempty"`
	FileKey  string   `json:"fileKey,omitempty"`
	VideoKey string   `json:"videoKey,omitempty"`
	Buttons  []Button `json:"buttons,omitempty"`
}

type Button struct {
	Text       string `json:"text"`
	ActionType int    `json:"actionType"`
	URL        string `json:"url,omitempty"`
	Value      string `json:"value,omitempty"`
}

type SendMessageResponse struct {
	Code int               `json:"code"`
	Msg  string            `json:"msg"`
	Data *SendMessageData  `json:"data,omitempty"`
}

type SendMessageData struct {
	MessageInfo *MessageInfo `json:"messageInfo,omitempty"`
}

type MessageInfo struct {
	MsgID    string `json:"msgId"`
	RecvID   string `json:"recvId"`
	RecvType string `json:"recvType"`
}

type EditMessageRequest struct {
	MsgID       string      `json:"msgId"`
	RecvID      string      `json:"recvId"`
	RecvType    string      `json:"recvType"`
	ContentType string      `json:"contentType"`
	Content     SendContent `json:"content"`
}

type EditMessageResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data *SendMessageData `json:"data,omitempty"`
}
