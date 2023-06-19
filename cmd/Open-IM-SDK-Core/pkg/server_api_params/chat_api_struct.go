package server_api_params

import server_api_params "Open_IM/pkg/proto/sdk_ws"

type DeleteMsgReq struct {
	OpUserID    string   `json:"opUserID"`
	UserID      string   `json:"userID"`
	ChatType    int64    `json:"chatType"`
	SeqList     []uint32 `json:"seqList"`
	OperationID string   `json:"operationID"`
}

type DeleteMsgResp struct {
}

type CleanUpMsgReq struct {
	UserID      string `json:"userID"  binding:"required"`
	OperationID string `json:"operationID"  binding:"required"`
}

type CleanUpMsgResp struct {
	CommResp
}
type DelSuperGroupMsgReq struct {
	UserID      string   `json:"userID,omitempty" binding:"required"`
	GroupID     string   `json:"groupID,omitempty" binding:"required"`
	SeqList     []uint32 `json:"seqList,omitempty"`
	IsAllDelete bool     `json:"isAllDelete"`
	OperationID string   `json:"operationID,omitempty" binding:"required"`
}
type DelSuperGroupMsgResp struct {
	CommResp
}
type MsgDeleteNotificationElem struct {
	GroupID     string   `json:"groupID"`
	IsAllDelete bool     `json:"isAllDelete"`
	SeqList     []uint32 `json:"seqList"`
}

type GetBroadcastStatusReq struct {
	UserID      string `json:"user_id" binding:"required"`
	OperationID string `json:"operation_id" binding:"required"`
}

type GetBroadcastStatusResp struct {
	CommResp
	Status int8 `json:"status"`
}

type ExtensionResult struct {
	CommResp
	server_api_params.KeyValue
}

type SetMessageReactionExtensionsReq struct {
	OperationID           string                                 `json:"operationID" validate:"required"`
	ClientMsgID           string                                 `json:"clientMsgID" validate:"required"`
	SourceID              string                                 `json:"sourceID" validate:"required"`
	SessionType           int32                                  `json:"sessionType" validate:"required"`
	ReactionExtensionList map[string]*server_api_params.KeyValue `json:"reactionExtensionList"`
	IsReact               bool                                   `json:"isReact,omitempty"`
	IsExternalExtensions  bool                                   `json:"isExternalExtensions,omitempty"`
	MsgFirstModifyTime    int64                                  `json:"msgFirstModifyTime,omitempty"`
}
