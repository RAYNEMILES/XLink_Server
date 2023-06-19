package cms_api_struct

type UserResponse struct {
	ProfilePhoto string `json:"profile_photo"`
	Nickname     string `json:"nick_name"`
	UserId       string `json:"user_id"`
	CreateTime   int64  `json:"create_time,omitempty"`
	IsBlock      bool   `json:"is_block"`
	Status       int32  `json:"status"`
	SourceId     string `json:"source_id"`
	SourceCode   string `json:"source_code"`
	// SuperUserStatus int64  `json:"super_user_status"`
	UpdateIp        string `json:"update_ip"`
	PhoneNumber     string `json:"phone_number"`
	Address         string `json:"address"`
	LastLoginTime   int64  `json:"last_login_time"`
	Email           string `json:"email"`
	Uuid            string `json:"uuid"`
	LastLoginDevice int8   `json:"last_login_device"`
	VideoStatus     int8   `json:"video_status"`
	AudioStatus     int8   `json:"audio_status"`
	Gender          int32  `json:"gender"`
	Remark          string `json:"remark"`
	LoginIp         string `json:"login_ip"`
}

type UserThirdInfoResponse struct {
	Nickname     string `json:"nick_name"`
	PhoneNumber  string `json:"phone_number"`
	UserId       string `json:"user_id"`
	Email        string `json:"email"`
	OfficialName string `json:"official_name"`
	Wallet       string `json:"wallet"`
	Facebook     string `json:"facebook"`
	Google       string `json:"google"`
	Apple        string `json:"apple"`
}

type GetUserRequest struct {
	UserId    string `form:"user_id" binding:"omitempty,alphanum,max=32"`
	SourceId  string `form:"source_id" binding:"omitempty,oneof=1 2 3"`
	Code      string `form:"code" binding:"omitempty,alphanum,len=6"`
	StartTime string `form:"start_time" binding:"omitempty,numeric"`
	EndTime   string `form:"end_time" binding:"omitempty,numeric"`
}

type GetUserResponse struct {
	UserResponse
}

type SwitchGuestStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=0 1"`
}

type GetUsersRequest struct {
	RequestPagination
	UserId          string `form:"user_id" binding:"omitempty,max=32"`
	SourceId        string `form:"source_id" binding:"omitempty,oneof=1 2 3"`
	Code            string `form:"code" binding:"omitempty"`
	StartTime       string `form:"start_time" binding:"omitempty,numeric"`
	EndTime         string `form:"end_time" binding:"omitempty,numeric"`
	Type            string `form:"type" binding:"omitempty,oneof=0 1 2"`
	AccountStatus   string `form:"account_status" binding:"omitempty"`
	Remark          string `form:"remark" binding:"omitempty"`
	LastLoginDevice int8   `form:"last_login_device" binding:"omitempty,numeric"`
	OrderBy         string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
	Gender          int32  `form:"gender" binding:"omitempty"`
}

type GetUsersResponse struct {
	Users []*UserResponse `json:"users"`
	ResponsePagination
	UserNums int64 `json:"user_nums"`
}

type GetUsersThirdInfoRequest struct {
	RequestPagination
	OperationID string `form:"operationID" binding:"omitempty"`
	UserId      string `form:"user_id" binding:"omitempty"`
	ThirdType   string `form:"third_type" binding:"omitempty,oneof=0 1 2 3 4 5"`
	ThirdName   string `form:"third_name" binding:"omitempty,min=1,max=32"`
}

type GetUsersThirdInfoResponse struct {
	Users []*UserThirdInfoResponse `json:"users"`
	ResponsePagination
	UserNums int64 `json:"user_nums"`
}

type GetUsersByNameRequest struct {
	UserName string `form:"user_name" binding:"required"`
	RequestPagination
}

type GetUsersByNameResponse struct {
	Users []*UserResponse `json:"users"`
	ResponsePagination
	UserNums int64 `json:"user_nums"`
}

type ResignUserRequest struct {
	UserId string `json:"user_id"`
}

type ResignUserResponse struct {
}

type AlterUserRequest struct {
	UserId      string   `json:"user_id" binding:"required"`
	Nickname    string   `json:"nickname"`
	PhoneNumber string   `json:"phone_number"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	Gender      int32    `json:"gender"`
	Interests   []string `json:"interests"`
	Remark      string   `json:"remark"`
	SourceId    string   `json:"source_id"`
	Code        string   `json:"code"`
	FaceURL     string   `json:"face_url"`
}

type AlterUserResponse struct {
}

type AddUserRequest struct {
	PhoneNumber string  `json:"phone_number" binding:"required"`
	UserId      string  `json:"user_id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Password    string  `json:"password" binding:"required"`
	Email       string  `json:"email" binding:"omitempty,email"`
	Code        string  `json:"code" binding:"omitempty"`
	SourceId    string  `json:"source_id" binding:"required"`
	Gender      int32   `json:"gender"`
	Interests   []int64 `json:"interests"`
	Remark      string  `json:"remark"`
	FaceURL     string  `json:"face_url"`
}

type AddUserResponse struct {
	CommResp
}

type MultiAddUserRequest struct {
	Users []*AddUserRequest `json:"users" binding:"required"`
}

type MultiAddUserResponse struct {
}

type BlockUser struct {
	UserResponse
	BeginDisableTime string `json:"begin_disable_time"`
	EndDisableTime   string `json:"end_disable_time"`
}

type BlockUserRequest struct {
	UserId         string `json:"user_id" binding:"required"`
	EndDisableTime string `json:"end_disable_time" binding:"required"`
}

type BlockUserResponse struct {
}

type UnblockUserRequest struct {
	UserId string `json:"user_id" binding:"required"`
}

type UnBlockUserResponse struct {
}

type ExistsUserRequest struct {
	OperationID string `json:"operationID" binding:"omitempty"`
	UserId      string `json:"user_id" binding:"omitempty,alphanum,max=32"`
	PhoneNumber string `json:"phone_number" binding:"omitempty"`
}

type GetBlockUsersRequest struct {
	RequestPagination
}

type GetBlockUsersResponse struct {
	BlockUsers []BlockUser `json:"block_users"`
	ResponsePagination
	UserNums int64 `json:"user_nums"`
}

type GetBlockUserRequest struct {
	UserId string `form:"user_id" binding:"required"`
}

type GetBlockUserResponse struct {
	BlockUser
}

type DeleteUserRequest struct {
	UserId string `json:"user_id" binding:"required"`
}

type DeleteSelfUserRequest struct {
	Reason string `json:"reason"`
}

type DeleteUserResponse struct {
}

type SwitchStatusRequest struct {
	UserId     string `json:"user_id" binding:"required"`
	StatusType int    `json:"status_type" binding:"required"`
	Status     int    `json:"status" binding:"required"`
}

type SwitchStatusResponse struct {
}

type GetDeletedUsersRequest struct {
	RequestPagination
	OperationID     string `form:"operationID"`
	User            string `form:"user"`
	Gender          int32  `form:"gender"`
	Reason          string `form:"reason"`
	Location        string `form:"location"`
	LastLoginDevice int32  `form:"last_login_device"`
	DeletedBy       string `form:"deleted_by"`
	TimeType        int32  `form:"time_type"`
	StartTime       string `form:"start_time"`
	EndTime         string `form:"end_time"`
}

type GetDeletedUsersResponse struct {
	CommResp
	ResponsePagination
	DeletedUsers []struct {
		UserID       string `json:"user_id"`
		Username     string `json:"username"`
		ProfilePhoto string `json:"profile_photo"`
		PhoneNumber  string `json:"phone_number"`
		Gender       int32  `json:"gender"`
		LastLoginIP  string `json:"last_login_ip"`
		Location     string `json:"location"`
		CreateTime   int64  `json:"create_time"`
		DeleteTime   int64  `json:"delete_time"`
		DeletedBy    string `json:"deleted_by"`
		Reason       string `json:"reason"`
	} `json:"deleted_users"`
	DeletedUsersCount int64 `json:"deleted_users_count"`
}
