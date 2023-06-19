package front_api_struct

type GetLatestAppVersionReq struct {
	OperationID string `json:"operation_id"`
	Client      string `json:"client"`
}

type GetLatestAppVersionResp struct {
	ID          string `json:"id"`
	Version     string `json:"url"`
	Type        int    `json:"type"`
	Status      int    `json:"status"`
	Isforce     int    `json:"isforce"`
	Title       string `json:"title"`
	DownloadUrl string `json:"download_url"`
	Content     string `json:"content"`
}
