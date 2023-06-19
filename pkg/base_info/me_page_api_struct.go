package base_info

type GetMePageURLReq struct {
	OperationID string `json:"operation_id"`
}

type GetMePageURLResp struct {
	Url    map[string]string `json:"url"`
	Status int               `json:"status"`
}

type MePageObj struct {
	Status int8              `json:"status"`
	URLMap map[string]string `json:"urls"`
}

type GetMePageURLsResp struct {
	MePageURL map[string]MePageObj
}
