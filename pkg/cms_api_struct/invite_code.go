package cms_api_struct

type AddInviteCodeRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	UserId      string `json:"user_id" binding:"required,alphanum,min=1,max=32"`
	Code        string `json:"code" binding:"omitempty,alphanum,len=6"`
	Greeting    string `json:"greeting" binding:"omitempty,max=255"`
	Note        string `json:"note" binding:"omitempty,max=255"`
}

type AddInviteCodeResponse struct {
	InviteCode string `json:"invite_code"`
}

type EditInviteCodeRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	Code        string `json:"code" binding:"omitempty,alphanum,len=6"`
	Greeting    string `json:"greeting" binding:"omitempty,max=255"`
	Note        string `json:"note" binding:"omitempty,max=255"`
}

type SwitchInviteCodeStateRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	InviteCode  string `json:"invite_code" binding:"required,alphanum,min=1,max=32"`
	State       int    `json:"state" binding:"required,oneof=1 2 3"`
}

type SwitchRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	State       string `json:"state" binding:"required,oneof=0 1"`
}
type LimitRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	State       string `json:"state" binding:"required,oneof=0 1"`
}
type MultiDeleteRequest struct {
	OperationID string   `json:"operationID" binding:"required"`
	InviteCodes []string `json:"invite_codes" binding:"required"`
}

type GetInviteCodeListRequest struct {
	OperationID string `json:"operationID" binding:"omitempty"`
	RequestPagination
	UserName string `form:"user_name" binding:"omitempty,min=1,max=32"`
	UserId   string `form:"user_id" binding:"omitempty,alphanum,min=1,max=32"`
	Code     string `form:"code" binding:"omitempty,alphanum,min=1,max=32"`
	Note     string `form:"note" binding:"omitempty,max=255"`
	State    int    `form:"state" binding:"omitempty,oneof=1 2"`
	OrderBy  string `form:"order_by" binding:"omitempty,oneof=id:asc id:desc"`
}
type GetInviteCodeListResponse struct {
	ResponsePagination
	InviteCodeList []InviteCodeCode `json:"list"`
	Total          int64            `json:"total"`
	IsOpen         int              `json:"is_open"`
	Limit          int              `json:"limit"`
	BaseLink       string           `json:"base_link"`
}
type InviteCodeCode struct {
	ID       int    `json:"id"`
	Code     string `json:"code"`
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
	Note     string `json:"note"`
	Greeting string `json:"greeting"`
	State    int32  `json:"state"`
}
