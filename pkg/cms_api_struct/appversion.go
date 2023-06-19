package cms_api_struct

type GetAppVersionByIDReq struct {
	OperationID string `json:"operation_id"`
	ID          string `json:"id"`
}

type GetAppVersionByIDResp struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Type        int    `json:"type"`
	Status      int    `json:"status"`
	Isforce     int    `json:"isforce"`
	Title       string `json:"title"`
	DownloadUrl string `json:"download_url"`
	Content     string `json:"content"`
	CreateTime  int64  `json:"create_time"`
	CreateUser  string `json:"create_user"`
	UpdateTime  int64  `json:"update_time"`
	UpdateUser  string `json:"update_user"`
}

type GetAppVersionsReq struct {
	OperationID     string `json:"operationID"`
	Client          string `json:"client"`
	Status          string `json:"status"`
	CreateTimeBegin string `json:"create_time_begin"`
	CreateTimeEnd   string `json:"create_time_end"`
	PageNumber      int    `json:"page_number"`
	ShowNumber      int    `json:"show_number"`
	OrderBy         string `json:"order_by" binding:"omitempty,oneof=create_time:asc create_time:desc"`
}

type GetAppVersionsResp struct {
	ResponsePagination
	Appversions []*GetAppVersionByIDResp `json:"appversions"`
	Total       int64                    `json:"total"`
}

type AddAppVersionReq struct {
	OperationID string `json:"operation_id"`
	Version     string `json:"version"`
	Client      string `json:"client"`
	Status      string `json:"status"`
	Isforce     string `json:"isforce"`
	Title       string `json:"title"`
	DownloadUrl string `json:"download_url"`
	Remark      string `json:"remark"`
}

type EditAppVersionReq struct {
	OperationID string `json:"operation_id"`
	ID          string `json:"id"`
	Version     string `json:"version"`
	Client      string `json:"client"`
	Status      string `json:"status"`
	Isforce     string `json:"isforce"`
	Title       string `json:"title"`
	DownloadUrl string `json:"download_url"`
	Remark      string `json:"remark"`
}

type DeleteAppVersionReq struct {
	OperationID string `json:"operation_id"`
	ID          string `json:"id"`
}
