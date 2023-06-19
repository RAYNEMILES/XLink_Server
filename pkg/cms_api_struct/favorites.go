package cms_api_struct

type Favorite struct {
	FavoriteId         string `json:"favorite_id"`
	UserID             string `json:"user_id"`
	UserName           string `json:"user_name"`
	ContentType        int32  `json:"content_type"`
	Content            string `json:"content"`
	CreateTime         int64  `json:"create_time"`
	PublishUser        string `json:"publish_user"`
	PublishTime        int64  `json:"publish_time"`
	Remark             string `json:"remark"`
	EditUser           string `json:"edit_user"`
	UpdateBy           string `json:"update_by"`
	UpdateTime         int64  `json:"update_time"`
	ContentCreatorName string `json:"content_creator_name"`
	FileSize           int64  `json:"file_size"`
	SourceType         int32  `json:"source_type"`
	CreateBy           string `json:"create_by"`
}

type GetFavoritesRequest struct {
	RequestPagination
	Account     string `form:"account" binding:"omitempty"`
	Content     string `form:"content" binding:"omitempty"`
	ContentType string `form:"content_type" binding:"omitempty"`
	PublishUser string `form:"publish_user" binding:"omitempty"`
	TimeType    int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime   string `form:"start_time" binding:"omitempty,numeric"`
	EndTime     string `form:"end_time" binding:"omitempty,numeric"`
	OrderBy     string `form:"order_by" binding:"omitempty,oneof=create_time:asc create_time:desc"`
}

type GetFavoritesResponse struct {
	ResponsePagination
	Favorites    []*Favorite `json:"favorites"`
	FavoriteNums int64       `json:"favorite_nums"`
}

type DeleteFavoritesRequest struct {
	OperationId string   `json:"operation_id"`
	FavoriteIds []string `json:"favorite_ids"`
}

type DeleteFavoritesResponse struct {
	CommResp
}

type AlterFavoritesRequest struct {
	OperationId      string `json:"operation_id"`
	UserID           string `json:"user_id"`
	ContentType      int32  `json:"content_type"`
	ExKeywords       string `json:"ex_keywords"`
	Content          string `json:"content"`
	ContentID        string `json:"content_id"`
	ContentCreatorID string `json:"content_creator_id"`
	SourceType       int32  `json:"source_type"`
	ContentGroupID   string `json:"content_group_id"`
	FavoriteId       string `json:"favorite_id"`
	Remark           string `json:"remark"`
}

type AlterFavoritesResponse struct {
	CommResp
}
