package front_api_struct

type GetDiscoverUrlReq struct {
	OperationID string `json:"string"`
}

type GetDiscoverUrlResp struct {
	Url        string `json:"url"`
	Status     int    `json:"status"`
	PlatformId int    `json:"platform_id"`
}
