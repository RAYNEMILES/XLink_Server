package cms_api_struct

type BroadcastRequest struct {
	Message string `json:"message"`
}

type BroadcastResponse struct {
}

type MassSendMassageRequest struct {
	Message string   `json:"message"`
	Users   []string `json:"users"`
}

type MassSendMassageResponse struct {
}

type GetChatLogsRequest struct {
	SessionType  int    `form:"session_type"`
	ContentType  int    `form:"content_type"`
	Content      string `form:"content"`
	UserId       string `form:"user_id"`
	GroupId      string `form:"group_id"`
	Date         string `form:"date"`
	OrderBy      string `form:"order_by" binding:"omitempty,oneof=create_time:asc create_time:desc"`
	ContentTypes string `form:"content_types"`
	RequestPagination
}

type GetChatLogsV1Request struct {
	SessionType int    `form:"session_type"`
	ContentType int    `form:"content_type"`
	Content     string `form:"content"`
	UserId      string `form:"user_id"`
	GroupId     string `form:"group_id"`

	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`

	RequestPagination
}

type ChatLog struct {
	SessionType    int    `json:"session_type"`
	ContentType    int    `json:"content_type"`
	SenderNickName string `json:"sender_nick_name"`
	SenderId       string `json:"sender_id"`
	SearchContent  string `json:"search_content"`
	WholeContent   string `json:"whole_content"`
	ClientMsgId    string `json:"client_msg_id"`
	Status         int    `json:"status"`

	ReceiverNickName string `json:"receiver_nick_name,omitempty"`
	ReceiverID       string `json:"receiver_id,omitempty"`

	GroupName string `json:"group_name,omitempty"`
	GroupId   string `json:"group_id,omitempty"`

	Date string `json:"date"`
}

type GetChatLogsResponse struct {
	ChatLogs    []ChatLog `json:"chat_logs"`
	ChatLogsNum int       `json:"log_nums"`
	ResponsePagination
}
