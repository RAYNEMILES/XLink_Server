package server_api_params

type Conversation struct {
	OwnerUserID      string `json:"ownerUserID" binding:"required"`
	ConversationID   string `json:"conversationID" binding:"required"`
	ConversationType int32  `json:"conversationType" binding:"required"`
	UserID           string `json:"userID"`
	GroupID          string `json:"groupID"`
	RecvMsgOpt       int32  `json:"recvMsgOpt"`
	UnreadCount      int32  `json:"unreadCount" `
	DraftTextTime    int64  `json:"draftTextTime"`
	IsPinned         bool   `json:"isPinned"`
	PinnedTime       int64  `json:"pinnedTime"`
	IsPrivateChat    bool   `json:"isPrivateChat"`
	GroupAtType      int32  `json:"groupAtType"`
	IsNotInGroup     bool   `json:"isNotInGroup"`
	AttachedInfo     string `json:"attachedInfo"`
	Ex               string `json:"ex"`
}

type SetConversationReq struct {
	Conversation
	NotificationType int    `json:"notificationType"`
	OperationID      string `json:"operationID" binding:"required"`
}

type SetConversationResp struct {
}
type ModifyConversationFieldReq struct {
	Conversation
	FieldType   int32    `json:"fieldType" binding:"required"`
	UserIDList  []string `json:"userIDList" binding:"required"`
	OperationID string   `json:"operationID" binding:"required"`
}
type ModifyConversationFieldResp struct {
	CommResp
}
type BatchSetConversationsReq struct {
	Conversations    []Conversation `json:"conversations" binding:"required"`
	OwnerUserID      string         `json:"ownerUserID" binding:"required"`
	NotificationType int            `json:"notificationType"`
	OperationID      string         `json:"operationID" binding:"required"`
}

type BatchSetConversationsResp struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
}

type GetConversationReq struct {
	ConversationID string `json:"conversationID" binding:"required"`
	OwnerUserID    string `json:"ownerUserID" binding:"required"`
	OperationID    string `json:"operationID" binding:"required"`
}

type GetConversationResp struct {
	Conversation Conversation `json:"data"`
}

type GetAllConversationsReq struct {
	OwnerUserID string `json:"ownerUserID" binding:"required"`
	OperationID string `json:"operationID" binding:"required"`
}

type GetAllConversationsResp struct {
	Conversations []Conversation `json:"data"`
}

type GetConversationsReq struct {
	ConversationIDs []string `json:"conversationIDs" binding:"required"`
	OwnerUserID     string   `json:"ownerUserID" binding:"required"`
	OperationID     string   `json:"operationID" binding:"required"`
}

type GetConversationsResp struct {
	CommResp
	Conversations []Conversation `json:"data"`
}

type GetConversationRecvMessageOptResp struct {
	ConversationID string `json:"conversationID"`
	Result         *int32 `json:"result"`
}

type GetUserIPAndStatusReq struct {
	ForUserID   string `json:"forUserID" binding:"required"`
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"userID" binding:"required"`
	IPAddress   string `json:"ipAddress"`
}
type GetUserIPAndStatusResp struct {
	UserID         string `json:"userID"`
	IPaddress      string `json:"iPaddress"`
	City           string `json:"city"`
	LastOnlineTime string `json:"lastOnlineTime"`
	OnlineDifValue int32  `json:"onlineDifValue"`
	OperationID    string `json:"operationID"`
}

type OfficialAccountConversationRequest struct {
	OperationID string `json:"operationID"`
}

type OfficialAccountConversationResponse struct {
	Articles []ArticleForLocalConv `json:"data"`
}

type ArticleForLocalConv struct {
	ArticleID          int64  `json:"ArticleID,omitempty"`
	OfficialID         int64  `json:"OfficialID,omitempty"`
	Title              string `json:"Title,omitempty"`
	CoverPhoto         string `json:"CoverPhoto,omitempty"`
	Content            string `json:"Content,omitempty"`
	TextContent        string `json:"TextContent,omitempty"`
	OfficialName       string `json:"OfficialName,omitempty"`
	OfficialProfileImg string `json:"OfficialProfileImg,omitempty"`
	CreatedBy          string `json:"CreatedBy,omitempty"`
	CreateTime         int64  `json:"CreateTime,omitempty"`
	UpdatedBy          string `json:"UpdatedBy,omitempty"`
	UpdateTime         int64  `json:"UpdateTime,omitempty"`
	DeletedBy          string `json:"DeletedBy,omitempty"`
	DeleteTime         int64  `json:"DeleteTime,omitempty"`
	Status             int32  `json:"Status,omitempty"`
	Privacy            int32  `json:"Privacy,omitempty"`
	CommentCounts      int64  `json:"CommentCounts,omitempty"`
	LikeCounts         int64  `json:"LikeCounts,omitempty"`
	ReadCounts         int64  `json:"ReadCounts,omitempty"`
	UniqueReadCounts   int64  `json:"UniqueReadCounts,omitempty"`
	RepostCounts       int64  `json:"RepostCounts,omitempty"`
	LastLoginIp        string `json:"LastLoginIp,omitempty"`
	LastLoginTime      string `json:"LastLoginTime,omitempty"`
	OfficialType       int32  `json:"OfficialType,omitempty"`
}
