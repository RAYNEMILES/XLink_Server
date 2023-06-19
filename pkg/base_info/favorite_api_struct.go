package base_info

type Favorite struct {
	FavoriteId       string                 `json:"favorite_id"`
	FileSize         int64                  `json:"file_size"`
	ExKeywords       string                 `json:"ex_keywords"`
	ContentType      int32                  `json:"content_type"`
	Content          map[string]interface{} `json:"content"`
	ContentID        string                 `json:"content_id"`
	ContentCreatorID string                 `json:"content_creator_id"`
	CreatorDetail    map[string]interface{} `json:"creator_detail"`
	SourceType       int32                  `json:"source_type"`
	PublishTime      int32                  `json:"publish_time"`
	CreateTime       int64                  `json:"create_time"`
}

type AddFavoriteReq struct {
	OperationID string                 `json:"operation_id"`
	UserID      string                 `json:"user_id"`
	ContentType int32                  `json:"content_type"`
	ExKeywords  string                 `json:"ex_keywords"`
	Content     map[string]interface{} `json:"content"`
	ContentID   string                 `json:"content_id"`
	SourceType  int32                  `json:"source_type"`
	PublishTime int64                  `json:"publish_time"`
}

type AddFavoriteResp struct {
	CommResp
}

type GetFavoriteListReq struct {
	OperationID string `json:"operation_id"`
	ContentType int32  `json:"content_type"`
	PageNumber  int32  `json:"page_number"`
	ShowNumber  int32  `json:"show_number"`
}

type GetFavoriteListResp struct {
	CommResp
	Favorites []*Favorite `json:"data"`
}

type RemoveFavoriteReq struct {
	OperationID string   `json:"operation_id"`
	FavoriteIds []string `json:"favorite_ids"`
}

type RemoveFavoriteResp struct {
	CommResp
}
