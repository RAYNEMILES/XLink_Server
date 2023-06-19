package cms_api_struct

type MemberSimple struct {
	UserName string `json:"nick_name"`
	UserID   string `json:"user_id"`
}

type GroupResponse struct {
	GroupName        string `json:"group_name"`
	GroupID          string `json:"group_id"`
	GroupMasterName  string `json:"group_master_name"`
	GroupMasterId    string `json:"group_master_id"`
	CreateTime       string `json:"create_time"`
	IsBanChat        bool   `json:"is_ban_chat"`
	IsBanPrivateChat bool   `json:"is_ban_private_chat"`
	ProfilePhoto     string `json:"profile_photo"`
	Notification     string `json:"group_notification"`
	Introduction     string `json:"group_introduction"`
	Remark           string `json:"remark"`
	VideoStatus      int8   `json:"video_status"`
	AudioStatus      int8   `json:"audio_status"`
	Status           int8   `json:"status"`
	IsOpen           int32  `json:"is_open"`
	Members          uint32 `json:"members"`
}

type GetGroupByIdRequest struct {
	GroupId string `form:"group_id" binding:"required"`
}

type GetGroupByIdResponse struct {
	GroupResponse
	MemberList []MemberSimple `json:"member_list"`
}

type GetGroupRequest struct {
	GroupName string `form:"group_name" binding:"required"`
	RequestPagination
}

type GetGroupResponse struct {
	Groups    []GroupResponse `json:"groups"`
	GroupNums int             `json:"group_nums"`
	ResponsePagination
}

type GetGroupsRequest struct {
	RequestPagination
	OrderBy     string `form:"order_by" binding:"omitempty,oneof=create_time:asc create_time:desc"`
	Group       string `form:"group" binding:"omitempty,max=32"`
	Member      string `form:"member" binding:"omitempty,max=32"`
	IsOpen      int32  `form:"is_open" binding:"omitempty,numeric,oneof=0 1 2"`
	Remark      string `form:"remark" binding:"omitempty"`
	GroupStatus string `form:"group_status" binding:"omitempty"`
	Owner       string `form:"owner" binding:"omitempty"`
	Creator     string `form:"creator" binding:"omitempty"`
	StartTime   string `form:"start_time" binding:"omitempty,numeric"`
	EndTime     string `form:"end_time" binding:"omitempty,numeric"`
}

type GetGroupsResponse struct {
	Groups    []GroupResponse `json:"groups"`
	GroupNums int             `json:"group_nums"`
	ResponsePagination
}

type CreateGroupRequest struct {
	GroupName     string   `json:"group_name" binding:"required"`
	GroupMasterId string   `json:"group_master_id" binding:"required,min=1,max=20"`
	GroupMembers  []string `json:"group_members" binding:"required"`
	Notification  string   `json:"group_notification"`
	Introduction  string   `json:"group_introduction"`
	Interests     []int64  `json:"interests"`

	GroupId   string `json:"group_id" binding:"omitempty,alphanum"`
	GroupType int32  `json:"group_type" binding:"omitempty,numeric"`
	Remark    string `json:"remark" binding:"omitempty,max=256"`
	FaceURL   string `json:"face_url" binding:"omitempty,max=256,url"`
	IsOpen    int32  `json:"isOpen" binding:"omitempty,oneof=0 1"`
}

type CreateGroupResponse struct {
}

type SetGroupMasterRequest struct {
	GroupId string `json:"group_id" binding:"required"`
	UserId  string `json:"user_id" binding:"required"`
}

type SetGroupMasterResponse struct {
}

type SetGroupMemberRequest struct {
	GroupId string `json:"group_id" binding:"required"`
	UserId  string `json:"user_id" binding:"required"`
}

type SetGroupMemberRespones struct {
}

type SetGroupAdminRequest struct {
	GroupId string `json:"group_id" binding:"required"`
	UserId  string `json:"user_id" binding:"required"`
}

type SetGroupAdminResponse struct {
}

type BanGroupChatRequest struct {
	GroupId string `json:"group_id" binding:"required"`
}

type BanGroupChatResponse struct {
}

type BanPrivateChatRequest struct {
	GroupId string `json:"group_id" binding:"required"`
}

type BanPrivateChatResponse struct {
}

type DeleteGroupRequest struct {
	GroupId string `json:"group_id" binding:"required"`
}

type DeleteGroupResponse struct {
}

type GetGroupMembersRequest struct {
	GroupID    string `form:"group_id"`
	RoleLevel  string `form:"role_level"`
	Member     string `form:"member"`
	RemarkName string `form:"remark_name"`
	Remark     string `form:"remark"`
	StartTime  string `form:"start_time" binding:"omitempty,numeric"`
	EndTime    string `form:"end_time" binding:"omitempty,numeric"`
	Permission string `form:"permission"`
	Status     string `form:"status"`
	RequestPagination
}

type GroupMemberResponse struct {
	MemberPosition int    `json:"member_position"`
	MemberName     string `json:"member_name"`
	MemberNickName string `json:"member_nick_name"`
	MemberId       string `json:"member_id"`
	MemberFaceURL  string `json:"member_group_face_url"`
	RoleLevel      int32  `json:"role_level"`
	JoinTime       string `json:"join_time"`
	JoinSource     int32  `json:"join_source"`
	InviterUserID  string `json:"inviter_user_id"`
	MuteEndTime    uint32 `json:"mute_end_time"`
	Remark         string `json:"remark"`
	VideoStatus    int8   `json:"video_status"`
	AudioStatus    int8   `json:"audio_status"`
}

type GetGroupMembersResponse struct {
	GroupMembers []GroupMemberResponse `json:"group_members"`
	ResponsePagination
	GroupName  string `json:"group_name"`
	MemberNums int    `json:"member_nums"`
}

type GroupMemberRequest struct {
	GroupId string   `json:"group_id" binding:"required"`
	Members []string `json:"members" binding:"required"`
}

type GroupMemberOperateResponse struct {
	Success []string `json:"success"`
	Failed  []string `json:"failed"`
}

type AddGroupMembersRequest struct {
	GroupMemberRequest
}

type AddGroupMembersResponse struct {
	GroupMemberOperateResponse
}

type RemoveGroupMembersRequest struct {
	GroupMemberRequest
}

type RemoveGroupMembersResponse struct {
	GroupMemberOperateResponse
}

type AlterGroupInfoRequest struct {
	GroupID      string `json:"group_id"`
	GroupName    string `json:"group_name"`
	Notification string `json:"group_notification"`
	Introduction string `json:"group_introduction"`
	ProfilePhoto string `json:"profile_photo"`
	GroupType    int    `json:"group_type"`
	Remark       string `json:"remark"`
	CanAddFriend int8   `json:"can_add_friend"`

	Interests []int64 `json:"interests"`
}

type AlterGroupInfoResponse struct {
}

type MuteGroupMemberReq struct {
	OperationID  string `json:"operationID" binding:"required"`
	GroupID      string `json:"groupID" binding:"required"`
	UserID       string `json:"userID" binding:"required"`
	MutedSeconds uint32 `json:"mutedSeconds" binding:"required"`
}
type MuteGroupMemberResp struct {
	CommResp
}
type CancelMuteGroupMemberReq struct {
	OperationID string `json:"operationID" binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
	UserID      string `json:"userID" binding:"required"`
}
type CancelMuteGroupMemberResp struct {
	CommResp
}
type CommResp struct {
	ErrCode int32  `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
}

type SetVideoAudioStatusRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
	StatusType  int32  `json:"status_type" binding:"required"`
	Status      int32  `json:"status"`
}

type SetVideoAudioStatusResponse struct {
}

type SetUserVideoAudioStatusRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	GroupID     string `json:"groupID" binding:"required"`
	MemberID    string `json:"member_id" binding:"required"`
	StatusType  int32  `json:"status_type" binding:"required"`
	Status      int32  `json:"status"`
}

type SetUserVideoAudioStatusResponse struct {
}

type UserIDAndName struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

type GetUsersByGroupReq struct {
	RequestPagination
	GroupID     string `form:"groupID" binding:"omitempty"`
	OperationID string `form:"operationID"`
	GetType     int32  `form:"get_type" binding:"omitempty,numeric"`
}

type GetUsersByGroupResp struct {
	ResponsePagination
	UserNums int64           `json:"user_nums"`
	Users    []UserIDAndName `json:"users"`
}
