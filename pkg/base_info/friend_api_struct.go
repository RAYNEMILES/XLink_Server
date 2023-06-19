package base_info

import open_im_sdk "Open_IM/pkg/proto/sdk_ws"

type ParamsCommFriend struct {
	OperationID string `json:"operationID" binding:"required"`
	ToUserID    string `json:"toUserID" binding:"required"`
	FromUserID  string `json:"fromUserID" binding:"required"`
}

type BlackUserInfo struct {
	UserID   string `json:"userID"`
	Nickname string `json:"nickname"`
	FaceURL  string `json:"faceURL"`
	Gender   int32  `json:"gender"`
	Ex       string `json:"ex"`
}

type AddBlacklistReq struct {
	ParamsCommFriend
}
type AddBlacklistResp struct {
	CommResp
}

type ImportFriendReq struct {
	FriendUserIDList []string `json:"friendUserIDList" binding:"required"`
	OperationID      string   `json:"operationID" binding:"required"`
	FromUserID       string   `json:"fromUserID" binding:"required"`
}
type UserIDResult struct {
	UserID string `json:"userID"`
	Result int32  `json:"result"`
}
type ImportFriendResp struct {
	CommResp
	UserIDResultList []UserIDResult `json:"data"`
}

type AddFriendReq struct {
	ParamsCommFriend
	ReqMsg string `json:"reqMsg"`
	Source string `json:"source"`
}
type AddFriendResp struct {
	CommResp
}

type AddFriendResponseReq struct {
	ParamsCommFriend
	Flag      int32  `json:"flag" binding:"required,oneof=-1 0 1"`
	HandleMsg string `json:"handleMsg"`
}
type AddFriendResponseResp struct {
	CommResp
}

type DeleteFriendReq struct {
	ParamsCommFriend
}
type DeleteFriendResp struct {
	CommResp
}

type GetBlackListReq struct {
	OperationID string `json:"operationID" binding:"required"`
	FromUserID  string `json:"fromUserID" binding:"required"`
}
type GetBlackListResp struct {
	CommResp
	BlackUserInfoList []*open_im_sdk.PublicUserInfo `json:"-"`
	Data              []map[string]interface{}      `json:"data"`
}

//type PublicUserInfo struct {
//	UserID   string `json:"userID"`
//	Nickname string `json:"nickname"`
//	FaceUrl  string `json:"faceUrl"`
//	Gender   int32  `json:"gender"`
//}

type SetFriendRemarkReq struct {
	ParamsCommFriend
	Remark string `json:"remark"`
}
type SetFriendRemarkResp struct {
	CommResp
}

type RemoveBlackListReq struct {
	ParamsCommFriend
}
type RemoveBlackListResp struct {
	CommResp
}

type IsFriendReq struct {
	ParamsCommFriend
}
type Response struct {
	Friend bool `json:"isFriend"`
}
type IsFriendResp struct {
	CommResp
	Response Response `json:"data"`
}

type GetFriendsInfoReq struct {
	OperationID   string   `json:"operationID" binding:"required"`
	FriendUserIDs []string `json:"FriendUserIDs" binding:"required"`
}
type GetFriendsInfoResp struct {
	CommResp
	FriendInfoList []*open_im_sdk.FriendInfo `json:"-"`
	Data           []map[string]interface{}  `json:"data"`
}

type GetFriendListReq struct {
	OperationID string `json:"operationID" binding:"required"`
	FromUserID  string `json:"fromUserID" binding:"required"`
}
type GetFriendListResp struct {
	CommResp
	FriendInfoList []*open_im_sdk.FriendInfo `json:"-"`
	Data           []map[string]interface{}  `json:"data"`
}

type GetFriendApplyListReq struct {
	OperationID string `json:"operationID" binding:"required"`
	FromUserID  string `json:"fromUserID" binding:"required"`
}
type GetFriendApplyListResp struct {
	CommResp
	FriendRequestList []*open_im_sdk.FriendRequest `json:"-"`
	Data              []map[string]interface{}     `json:"data"`
}

type GetSelfApplyListReq struct {
	OperationID string `json:"operationID" binding:"required"`
	FromUserID  string `json:"fromUserID" binding:"required"`
}
type GetSelfApplyListResp struct {
	CommResp
	FriendRequestList []*open_im_sdk.FriendRequest `json:"-"`
	Data              []map[string]interface{}     `json:"data"`
}
type GetFriendRemarkOrNickReq struct {
	ForUserID string `json:"ForUserID" binding:"required"`
	GroupID   string `json:"groupID"`
}
type GetFriendRemarkOrNickResp struct {
	CommResp
	GetFriendRemarkOrNick GetFriendRemarkOrNick `json:"data"`
}
type GetFriendRemarkOrNick struct {
	RemarkNickName string `json:"remarkNickName"`
}

type AddBlackFriendsReq struct {
	OperationID string   `json:"operationID" binding:"required"`
	ToUsersID   []string `json:"toUsersID" binding:"required"`
	FromUserID  string   `json:"fromUserID" binding:"required"`
}

type AddBlackFriendsResp struct {
	CommResp
}

type GetBlackFriendsReq struct {
	OperationID string `json:"operationID" binding:"required"`
	FromUserID  string `json:"fromUserID" binding:"required"`
}

type GetBlackFriendsResp struct {
	CommResp
	Data []BlackUserInfo `json:"data"`
}

type RemoveBlackFriendsReq struct {
	OperationID string   `json:"operationID" binding:"required"`
	ToUsersID   []string `json:"toUsersID" binding:"required"`
	FromUserID  string   `json:"fromUserID" binding:"required"`
}

type RemoveBlackFriendsResp struct {
	CommResp
}
