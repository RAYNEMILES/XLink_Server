package cms_api_struct

type InterestLanguage struct {
	LanguageType string `json:"language_type"`
	Name         string `json:"name"`
}

type InterestType struct {
	Id         int64              `json:"id"`
	Status     int32              `json:"status"`
	Remark     string             `json:"remark"`
	CreateUser string             `json:"create_user"`
	CreateTime int64              `json:"create_time"`
	UpdateUser string             `json:"update_user"`
	UpdateTime int64              `json:"update_time"`
	DeleteTime int64              `json:"delete_time"`
	IsDefault  int8               `json:"is_default"`
	Name       []InterestLanguage `json:"name"`
}

type UserInterests struct {
	Username  string         `json:"username"`
	UserID    string         `json:"user_id"`
	Type      int32          `json:"type"`
	Interests []InterestType `json:"interests"`
}

type GroupInterests struct {
	GroupID   string         `json:"group_id"`
	GroupName string         `json:"group_name"`
	GroupType int32          `json:"group_type"`
	Interests []InterestType `json:"interests"`
}

type GetInterestsRequest struct {
	RequestPagination
	IsDefault  int8   `form:"is_default"`
	Name       string `form:"name" binding:"omitempty"`
	CreateUser string `form:"create_user" binding:"omitempty"`
	Remark     string `form:"remark" binding:"omitempty"`
	TimeType   int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime  string `form:"start_time" binding:"omitempty,numeric"`
	EndTime    string `form:"end_time" binding:"omitempty,numeric"`
	Status     string `form:"status" binding:"omitempty"`
	OrderBy    string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetInterestsResponse struct {
	Interests    []*InterestType `json:"interests"`
	InterestNums int32           `json:"interest_nums"`
	ResponsePagination
}

type DeleteInterestsRequest struct {
	Interests string `json:"interests"`
}

type AlterInterestsRequest struct {
	Id        int64              `json:"id"`
	Status    int8               `json:"status"`
	Remark    string             `json:"remark"`
	IsDefault int8               `json:"is_default"`
	Name      []InterestLanguage `json:"name"`
}

type AlterInterestsResponse struct {
}

type ChangeInterestStatusRequest struct {
	InterestId int64 `json:"interest_id"`
	Status     int32 `json:"status"`
}

type ChangeInterestStatusResponse struct {
}

type AddInterestsRequest struct {
	Interests []InterestType `json:"interests"`
}

type AddInterestsResponse struct {
}

type AddOneInterestRequest struct {
	InterestType
}

type AddOneInterestResponse struct {
}

type GetUserInterestsRequest struct {
	RequestPagination
	UserID       string `form:"user_id" binding:"omitempty"`
	Account      string `form:"account" binding:"omitempty"`
	InterestName string `form:"interest_name" binding:"omitempty"`
	Default      int8   `form:"default" binding:"omitempty"`
	OrderBy      string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetUserInterestsResponse struct {
	ResponsePagination
	Interests    []*UserInterests `json:"interests"`
	InterestNums int32            `json:"interest_nums"`
}

type AlterUserInterestsRequest struct {
	UserID    string `json:"user_id"`
	Interests string `json:"interests"`
}

type DeleteUserInterestsRequest struct {
	UsersID string `json:"users_id"`
}

type GetGroupInterestsRequest struct {
	RequestPagination
	GroupID      string `form:"group_id" binding:"omitempty"`
	CreatorUser  string `form:"creator_user" binding:"omitempty"`
	GroupName    string `form:"group_name" binding:"omitempty"`
	InterestName string `form:"interest_name" binding:"omitempty"`
	OrderBy      string `form:"order_by" binding:"omitempty,oneof=start_time:asc start_time:desc"`
}

type GetGroupInterestsResponse struct {
	ResponsePagination
	Interests    []*GroupInterests `json:"interests"`
	InterestNums int32             `json:"interest_nums"`
}

type AlterGroupInterestsRequest struct {
	GroupID   string `json:"group_id"`
	Interests string `json:"interests"`
}
