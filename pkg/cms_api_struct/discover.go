package cms_api_struct

type GetDiscoverUrlReq struct {
	OperationID string `form:"operationID" binding:"required"`
	PlatformID  string `json:"platform_id"`
}

type GetDiscoverUrlResp struct {
	ID         uint   `json:"id"`
	Url        string `json:"url"`
	Status     int    `json:"status"`
	PlatformId int    `json:"platform_id"`
	CreateTime int64  `json:"create_time"`
	CreateUser string `json:"create_user"`
	UpdateTime int64  `json:"update_time"`
	UpdateUser string `json:"update_user"`
}

type SaveDiscoverUrlReq struct {
	OperationID string `form:"operationID" binding:"required"`
	PlatformID  string `json:"platform_id"`
	Url         string `json:"url"`
}

type SwitchDiscoverStatusReq struct {
	OperationID string `form:"operationID" binding:"required"`
	PlatformID  string `json:"platform_id"`
	Status      int    `json:"status"`
}
