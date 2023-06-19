package base_info

type CommRespCode struct {
	ErrCode int32  `json:"code"`
	ErrMsg  string `json:"errMsg"`
}

type CategoryMultiLanguage struct {
	Category map[string]string `json:"category"`
}

type GameLink struct {
	PackageURL string `json:"package_url"`
	PlayURL    string `json:"play_url"`
}

type BannerGameInfo struct {
	GameCode string            `json:"game_code"`
	GameName map[string]string `json:"game_name"`
	CoverImg string            `json:"cover_img"`
}

type Categories struct {
	Id         int64             `json:"id"`
	Categories map[string]string `json:"categories"`
}

type GameNameInfo struct {
	GameCode string            `json:"game_code"`
	CoverImg string            `json:"cover_img"`
	GameName map[string]string `json:"game_name"`
}

type GameListInfo struct {
	GameCode   string                  `json:"game_code"`
	GameName   map[string]string       `json:"game_name"`
	CoverImg   string                  `json:"cover_img"`
	Categories []CategoryMultiLanguage `json:"categories"`
	Hot        int32                   `json:"hot"`
	CreateTime int64                   `json:"create_time"`
	DeleteTime int64                   `json:"delete_time"`
}

type BannerGamesReq struct {
	OperationID string `json:"operation_id" binding:"required"`
}

type BannerGamesResp struct {
	CommRespCode
	BannerGameList []BannerGameInfo `json:"banner_game_list"`
}

type TodayRecommendationsReq struct {
	PageNumber  int32   `json:"page_number" binding:"required"`
	ShowNumber  int32   `json:"show_number" binding:"required"`
	OperationID string  `json:"operation_id" binding:"required"`
	Categories  []int64 `json:"categories"`
}

type TodayRecommendationsResp struct {
	CommRespCode
	CurrentPage            int32          `json:"current_page"`
	ShowNumber             int32          `json:"show_number"`
	RecommendationGameList []GameListInfo `json:"recommendation_game_list"`
	GameNums               int32          `json:"game_nums"`
}

type PopularGamesReq struct {
	PageNumber  int32   `json:"page_number" binding:"required"`
	ShowNumber  int32   `json:"show_number" binding:"required"`
	OperationID string  `json:"operation_id" binding:"required"`
	Categories  []int64 `json:"categories"`
}

type PopularGamesResp struct {
	CommRespCode
	CurrentPage     int32          `json:"current_page"`
	ShowNumber      int32          `json:"show_number"`
	PopularGameList []GameListInfo `json:"popular_game_list"`
	GameNums        int32          `json:"game_nums"`
}

type AllGamesReq struct {
	PageNumber  int32   `json:"page_number" binding:"required"`
	ShowNumber  int32   `json:"show_number" binding:"required"`
	OperationID string  `json:"operation_id" binding:"required"`
	Categories  []int64 `json:"categories"`
}

type AllGamesResp struct {
	CommRespCode
	CurrentPage int32          `json:"current_page"`
	ShowNumber  int32          `json:"show_number"`
	AllGameList []GameListInfo `json:"all_game_list"`
	GameNums    int32          `json:"game_nums"`
}

type SearchNameReq struct {
	OperationID string `json:"operation_id"`
	Name        string `json:"name"`
	NameType    string `json:"name_type"`
	GetCount    int32  `json:"get_count"`
}

type SearchNameResp struct {
	CommRespCode
	GameList []*GameNameInfo `json:"game_list"`
	GameNums int32           `json:"game_nums"`
}

type SearchGameListByNameReq struct {
	OperationID string  `json:"operation_id"`
	PageNumber  int32   `json:"page_number" binding:"required"`
	ShowNumber  int32   `json:"show_number" binding:"required"`
	Name        string  `json:"name"`
	NameType    string  `json:"name_type"`
	Categories  []int64 `json:"categories"`
}

type SearchGameListByNameResp struct {
	CommRespCode
	CurrentPage int32          `json:"current_page"`
	ShowNumber  int32          `json:"show_number"`
	GameList    []GameListInfo `json:"game_list"`
	GameNums    int32          `json:"game_nums"`
}

type GetCategoriesReq struct {
	OperationID string `json:"operation_id"`
}

type GetCategoriesResp struct {
	CommRespCode
	Categories []*Categories `json:"categories"`
}

type PlayGameReq struct {
	OperationID string `json:"operation_id"`
	GameCode    string `json:"game_code"`
}

type PlayGameResp struct {
	CommRespCode
}

type GetGameHistoryReq struct {
	OperationID string `json:"operation_id"`
	PageNumber  int32  `json:"page_number" binding:"required"`
	ShowNumber  int32  `json:"show_number" binding:"required"`
}

type GetGameHistoryResp struct {
	CommRespCode
	CurrentPage int32          `json:"current_page"`
	ShowNumber  int32          `json:"show_number"`
	GameList    []GameListInfo `json:"game_list"`
	GameNums    int32          `json:"game_nums"`
}

type GetGameFavoritesReq struct {
	OperationID string `json:"operation_id"`
	PageNumber  int32  `json:"page_number" binding:"required"`
	ShowNumber  int32  `json:"show_number" binding:"required"`
}

type GetGameFavoritesResp struct {
	CommRespCode
	CurrentPage int32          `json:"current_page"`
	ShowNumber  int32          `json:"show_number"`
	GameList    []GameListInfo `json:"game_list"`
	GameNums    int32          `json:"game_nums"`
}

type RemoveGameFavoriteReq struct {
	OperationID string `json:"operation_id"`
	GameCode    string `json:"game_code"`
}

type RemoveGameFavoriteResp struct {
	CommRespCode
}

type AddGameFavoriteReq struct {
	OperationID string `json:"operation_id"`
	GameCode    string `json:"game_code"`
}

type AddGameFavoriteResp struct {
	CommRespCode
}

type GameDetailsReq struct {
	OperationID string `json:"operation_id"`
	GameCode    string `json:"game_code"`
}

type GameDetailsResp struct {
	CommRespCode
	GameCode        string                  `json:"game_code"`
	GameName        map[string]string       `json:"game_name"`
	CoverImg        string                  `json:"cover_img"`
	HorizontalCover []string                `json:"horizontal_cover"`
	PlatformLink    map[int32]GameLink      `json:"platform_link"`
	Categories      []CategoryMultiLanguage `json:"categories"`
	Description     map[string]string       `json:"description"`
	Hot             int32                   `json:"hot"`
	HasFavorite     bool                    `json:"has_favorite"`
	Publisher       string                  `json:"publisher"`
}
