package cms_api_struct

type BlackListRes struct {
	OwnerUserName     string `json:"owner_user_name"`
	OwnerUserID       string `json:"owner_user_id"`
	OwnerProfilePhoto string `json:"owner_profile"`
	BlockUserName     string `json:"block_user_name"`
	BlockUserID       string `json:"block_user_id"`
	CreateTime        int64  `json:"create_time"`
	Remark            string `json:"remark"`
	EditUser          string `json:"edit_user"`
	EditTime          int64  `json:"edit_time"`
	Ex                string `json:"ex"`
}

type GetBlacksReq struct {
	RequestPagination
	OwnerUser string `form:"owner_user" binding:"omitempty"`
	BlockUser string `form:"block_user" binding:"omitempty"`
	Remark    string `form:"remark" binding:"omitempty"`
	StartTime string `form:"start_time" binding:"omitempty"`
	EndTime   string `form:"end_time" binding:"omitempty"`
	OrderBy   string `form:"order_by" binding:"omitempty"`
}

type GetBlacksResp struct {
	ResponsePagination
	ListNumber int64          `json:"list_number"`
	BlackList  []BlackListRes `json:"black_list"`
}

type RemoveBlackReq struct {
	FriendList []struct {
		OwnerID string `json:"owner_id"`
		BlackID string `json:"black_id"`
	} `json:"friend_list"`
}

type RemoveBlackResp struct {
	CommResp
}

type AlterRemarkReq struct {
	OwnerUser string `json:"owner_user"`
	BlockUser string `json:"block_user"`
	Remark    string `json:"remark"`
}

type AlterRemarkResp struct {
	CommResp
}
