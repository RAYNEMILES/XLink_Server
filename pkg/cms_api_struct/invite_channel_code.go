package cms_api_struct

import apiStruct "Open_IM/pkg/base_info"

type AddInviteChannelCodeRequest struct {
	OperationID string   `json:"operationID" binding:"required"`
	Code        string   `json:"code" binding:"omitempty"`
	FriendId    []string `json:"friend_id" binding:"required"`
	GroupId     []string `json:"group_id" binding:"required"`
	Greeting    string   `json:"greeting" binding:"omitempty,max=255"`
	Note        string   `json:"note" binding:"omitempty,max=255"`
	SourceId    int64    `json:"source_id" binding:"required,oneof=1 3"`
}
type AddInviteChannelCodeResponse struct {
	apiStruct.CommResp
}

type SwitchInviteChannelCodeRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	Code        string `json:"code" binding:"required"`
	State       int    `json:"state" binding:"required,oneof=1 2 3"`
}
type ChannelCodeSwitchRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	State       string `json:"state" binding:"required,oneof=0 1"`
}
type ChannelCodeLimitRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	State       string `json:"state" binding:"required,oneof=0 1"`
}
type ChannelCodeMultiDeleteRequest struct {
	OperationID  string   `json:"operationID" binding:"required"`
	ChannelCodes []string `json:"channel_codes" binding:"required"`
}
type GetInviteChannelCodeListRequest struct {
	OperationID string `form:"operationID" binding:"required"`
	RequestPagination
	Code       string `form:"code" binding:"omitempty"`
	FriendId   string `form:"friend_id" binding:"omitempty,alphanum,min=1,max=32"`
	GroupId    string `form:"group_id" binding:"omitempty,alphanum,min=1,max=32"`
	Note       string `form:"note" binding:"omitempty,max=255"`
	State      string `form:"state" binding:"omitempty,oneof=1 2 3"`
	IsOfficial string `form:"is_official" binding:"omitempty,oneof=0 1"`
	OrderBy    string `form:"order_by" binding:"omitempty,oneof=id:asc id:desc"`
}

type GetInviteChannelCodeListResponse struct {
	ResponsePagination
	IsOpen          int                  `json:"is_open"`
	IsLimit         int                  `json:"limit"`
	Total           int64                `json:"total"`
	ChannelCodeList []*InviteChannelCode `json:"list"`
}
type InviteChannelCode struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	FriendIdArr []string  `json:"friend_id"`
	GroupIdArr  []string  `json:"group_id"`
	Greeting    string    `json:"greeting"`
	Note        string    `json:"note"`
	State       int32     `json:"state"`
	SourceId    int64     `json:"source_id"`
	FriendsList []Friends `json:"friend_list"`
	GroupList   []Groups  `json:"group_list"`
}
type Friends struct {
	ID   string `json:"user_id"`
	Name string `json:"nick_name"`
}
type Groups struct {
	ID   string `json:"group_id"`
	Name string `json:"group_name"`
}
