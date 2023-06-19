package front_api_struct

type FrontApiResp struct {
	ErrCode int32                  `json:"errCode"`
	ErrMsg  string                 `json:"errMsg"`
	Data    map[string]interface{} `json:"data"`
}
