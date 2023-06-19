package base_info

import (
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
)

type GetUsersInfoReq struct {
	OperationID string   `json:"operationID" binding:"required"`
	UserIDList  []string `json:"userIDList" binding:"required"`
}
type GetUsersInfoResp struct {
	CommResp
	UserInfoList []*open_im_sdk.PublicUserInfo `json:"-"`
	Data         []map[string]interface{}      `json:"data"`
}

type UpdateSelfUserInfoReq struct {
	ApiUserInfo
	OperationID string `json:"operationID" binding:"required"`
}
type SetGlobalRecvMessageOptReq struct {
	OperationID      string `json:"operationID" binding:"required"`
	GlobalRecvMsgOpt *int32 `json:"globalRecvMsgOpt" binding:"omitempty,oneof=0 1 2"`
}
type SetGlobalRecvMessageOptResp struct {
	CommResp
}
type UpdateUserInfoResp struct {
	CommResp
}

type GetSelfUserInfoReq struct {
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"userID" binding:"required"`
}
type GetSelfUserInfoResp struct {
	CommResp
	UserInfo *open_im_sdk.UserInfo  `json:"-"`
	Data     map[string]interface{} `json:"data"`
}

type GetFriendIDListFromCacheReq struct {
	OperationID string `json:"operationID" binding:"required"`
}

type GetFriendIDListFromCacheResp struct {
	CommResp
	UserIDList []string `json:"userIDList" binding:"required"`
}

type GetBlackIDListFromCacheReq struct {
	OperationID string `json:"operationID" binding:"required"`
}

type GetBlackIDListFromCacheResp struct {
	CommResp
	UserIDList []string `json:"userIDList" binding:"required"`
}

type GetInviteCodeLinkRequest struct {
	OperationID string `json:"operationID" binding:"required"`
}
type GetInviteCodeLinkResponse struct {
	CommResp
	Data map[string]interface{} `json:"data"`
}
type GetInvitionTotalRespone struct {
	CommResp
	Data map[string]interface{} `json:"data"`
}

type BindingListResponse struct {
	CommResp
	Data []map[string]interface{} `json:"data"`
}

type SyncContactRequest struct {
	OperationID string   `json:"operationID" binding:"required"`
	ContactList []string `json:"contactList" binding:"required"`
}

type RemoveFamiliarUserRequest struct {
	OperationID string   `json:"operationID" binding:"required"`
	UserId      []string `json:"userId" binding:"required"`
}

type RemoveInterestGroupRequest struct {
	OperationID string   `json:"operationID" binding:"required"`
	Group       []string `json:"group" binding:"required"`
}

type GetFamiliarListRequest struct {
	OperationID string `json:"operationID" binding:"required"`
}

type GetFamiliarListResponse struct {
	CommResp
	Data []GetFamiliarListData `json:"data"`
}
type GetFamiliarListData struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	Avatar   string `json:"avatar"`
	Gender   int32  `json:"gender"`
	Contact  int    `json:"contact"`
	Friend   int    `json:"friend"`
	Meet     int    `json:"meet"`
}
type VideoAudioStatusReq struct {
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"userId" binding:"required"`
	GroupID     string `json:"group_id"`
}
type VideoAudioStatusResp struct {
	CommResp
	Data map[string]interface{} `json:"data"`
}
