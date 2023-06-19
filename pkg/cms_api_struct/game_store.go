package cms_api_struct

type CommRespCode struct {
	ErrCode int32  `json:"code"`
	ErrMsg  string `json:"errMsg"`
}

type CategoryInfo struct {
	Categories   map[string]string `json:"categories"`
	CategoryID   int64             `json:"category_id"`
	UsedAmount   int64             `json:"used_amount"`
	Creator      string            `json:"creator"`
	CreateTime   int64             `json:"create_time"`
	Editor       string            `json:"editor"`
	EditTime     int64             `json:"edit_time"`
	SortPriority int32             `json:"sort_priority"`
	State        int8              `json:"state"`
}

type GameInfo struct {
	GameCode        string             `json:"game_code" binding:"required"`
	GameName        map[string]string  `json:"game_name" binding:"required"`
	Description     map[string]string  `json:"description"`
	HorizontalCover []string           `json:"horizontal_cover" binding:"required"`
	CoverImg        string             `json:"cover_img" binding:"required"`
	Hot             int32              `json:"hot" binding:"required"`
	Classification  []string           `json:"classification"`
	PlatformLink    map[int32]GameLink `json:"platform_link" binding:"required"`
	SortPriority    int32              `json:"sort_priority"`
	State           int8               `json:"state"`
	Publisher       string             `json:"publisher"`
}

type GameListInfo struct {
	GameInfo
	Categories []CategoryMultiLanguage `json:"categories" binding:"required"`
	CreateTime int64                   `json:"create_time"`
}

type GameLink struct {
	PackageURL string `json:"package_url"`
	PlayURL    string `json:"play_url"`
	FileName   string `json:"file_name"`
	Size       int64  `json:"size"`
	UploadTime int64  `json:"upload_time"`
}

type CategoryMultiLanguage struct {
	Category map[string]string `json:"category"`
}

type GetGameListReq struct {
	RequestPagination
	OperationID    string `form:"operation_id" binding:"required"`
	GameCode       string `form:"game_code" binding:"omitempty"`
	NameType       string `form:"name_type" binding:"omitempty"`
	GameName       string `form:"game_name" binding:"omitempty"`
	Category       int64  `form:"category" binding:"omitempty"`
	Classification int64  `form:"classification" binding:"omitempty"`
	Platform       int32  `form:"platform" binding:"omitempty"`
	Publisher      string `form:"publisher" binding:"omitempty"`
	Hot            int32  `form:"hot" binding:"omitempty"`
	State          int8   `form:"state" binding:"omitempty"`
	StartTime      string `form:"start_time" binding:"omitempty,numeric"`
	EndTime        string `form:"end_time" binding:"omitempty,numeric"`
}

type GetGameListResp struct {
	CommRespCode
	ResponsePagination
	GameList []GameListInfo `json:"game_list"`
	GameNum  int64          `json:"game_num"`
}

type EditGameReq struct {
	OperationID string `json:"operation_id" binding:"required"`
	GameInfo
	Categories []string `json:"categories" binding:"required"`
}

type EditGameResp struct {
	CommRespCode
}

type AddGameReq struct {
	OperationID string `json:"operation_id" binding:"required"`
	GameInfo
	Categories []string `json:"categories" binding:"required"`
}

type AddGameResp struct {
	CommRespCode
}

type DeleteGamesReq struct {
	OperationID string   `json:"operation_id" binding:"required"`
	GamesCode   []string `json:"games_code" binding:"required"`
}

type DeleteGamesResp struct {
	CommRespCode
}

type GetCategoryReq struct {
	RequestPagination
	OperationID     string `form:"operation_id" binding:"required"`
	NameType        string `form:"name_type"`
	CategoryName    string `form:"category_name"`
	Creator         string `form:"creator"`
	CreateStartTime string `form:"create_start_time" binding:"omitempty,numeric"`
	CreateEndTime   string `form:"create_end_time" binding:"omitempty,numeric"`
	EditedStartTime string `form:"edited_start_time" binding:"omitempty,numeric"`
	EditedEndTime   string `form:"edited_end_time" binding:"omitempty,numeric"`
	Editor          string `form:"editor"`
	State           int8   `form:"state"`
}

type GetCategoryResp struct {
	CommRespCode
	ResponsePagination
	CategoryList  []CategoryInfo `json:"category_list"`
	CategoriesNum int64          `json:"categories_num"`
}

type AddCategoryReq struct {
	OperationID  string            `json:"operation_id" binding:"required"`
	CategoryName map[string]string `json:"category_name"`
	SortPriority int32             `json:"sort_priority"`
	State        int8              `json:"state"`
}

type AddCategoryResp struct {
	CommRespCode
}

type EditCategoryReq struct {
	OperationID  string            `json:"operation_id" binding:"required"`
	CategoryID   int64             `json:"category_id" binding:"required"`
	CategoryName map[string]string `json:"category_name"`
	SortPriority int32             `json:"sort_priority"`
	State        int8              `json:"state"`
}

type EditCategoryResp struct {
	CommRespCode
}

type SetCategoryStatusReq struct {
	OperationID string `json:"operation_id" binding:"required"`
	CategoryID  int64  `json:"category_id"`
	State       int8   `json:"state"`
}

type SetCategoryStatusResp struct {
	CommRespCode
}

type DeleteCategoryReq struct {
	OperationID string `json:"operation_id" binding:"required"`
	CategoryID  int64  `json:"category_id"`
}

type DeleteCategoryResp struct {
	CommRespCode
}
