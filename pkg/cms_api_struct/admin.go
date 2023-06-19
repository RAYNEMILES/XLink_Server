package cms_api_struct

import (
	apiStruct "Open_IM/pkg/base_info"
)

type AdminUserResponse struct {
	Name             string `json:"name"`
	NickName         string `json:"nick_name"`
	UserId           string `json:"user_id"`
	CreateTime       int64  `json:"create_time"`
	TwoFactorEnabled int64  `json:"two_factor_enabled"`
	Status           int32  `json:"status"`
	Role             int64  `json:"role_id"`
	CreateUser       string `json:"create_user"`
	UpdateUser       string `json:"update_user"`
	IPRangeStart     string `json:"ip_range_start"`
	IPRangeEnd       string `json:"ip_range_end"`
	LastLoginIP      string `json:"last_login_ip"`
	Remarks          string `json:"remarks"`
	User2FAuthEnable int64  `json:"user_2fauth_enable"`
	LastLoginTime    int64  `json:"last_login_time"`
}

type AdminLoginRequest struct {
	AdminName string `json:"admin_name" binding:"required"`
	Secret    string `json:"secret" binding:"required"`
}

type AdminLoginResponse struct {
	Token              string `json:"token"`
	GAuthEnabled       bool   `json:"gAuthEnabled"`
	GAuthSetupRequired bool   `json:"gAuthSetupRequired"`
	GAuthSetupProvUri  string `json:"gAuthSetupProvUri"`
}

type InviteCodeBashLinkResponse struct {
	BaseLink string `json:"base_link"`
}
type SetInviteCodeBashLinkRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	BaseLink    string `json:"base_link" binding:"required,url"`
}

type UploadUpdateAppReq struct {
	OperationID string `form:"operationID" binding:"required"`
	Type        int    `form:"type" binding:"required"`
	Version     string `form:"version"  binding:"required"`
	// File        *multipart.FileHeader `form:"file" binding:"required"`
	// Yaml        *multipart.FileHeader `form:"yaml" binding:"required"`
	ForceUpdate bool `form:"forceUpdate"  binding:"required"`
}

type UploadUpdateAppResp struct {
	apiStruct.CommResp
}

type GetDownloadURLReq struct {
	OperationID string `json:"operationID" binding:"required"`
	Type        int    `json:"type" binding:"required"`
	Version     string `json:"version" binding:"required"`
}

type GetDownloadURLResp struct {
	apiStruct.CommResp
	Data struct {
		HasNewVersion bool   `json:"hasNewVersion"`
		ForceUpdate   bool   `json:"forceUpdate"`
		FileURL       string `json:"fileURL"`
		YamlURL       string `json:"yamlURL"`
	} `json:"data"`
}

type ParamsTOTPVerify struct {
	TOTP        string `json:"totp" binding:"required"`
	OperationID string `json:"operationID"`
}

type AddAdminUserRequest struct {
	UserId           string `json:"user_id" binding:"required"`
	Name             string `json:"name"`
	Nickname         string `json:"nick_name"`
	Password         string `json:"password"`
	PhoneNumber      string `json:"phoneNumber"`
	TwoFactorEnabled int64  `json:"two_factor_enabled"`
	User2FAuthEnable int64  `json:"user_2fauth_enable"`
	IPRangeStart     string `json:"ip_range_start"`
	IPRangeEnd       string `json:"ip_range_end"`
	Role             int16  `json:"role"`
	Status           int64  `json:"status"`
	Remarks          string `json:"remarks"`
}

type DeleteAdminUserRequest struct {
	UserId string `json:"user_id" binding:"required"`
}

type DeleteAdminUserResponse struct {
}

type AlterAdminUserRequest struct {
	UserId           string `json:"user_id" binding:"required"`
	Name             string `json:"name"`
	Nickname         string `json:"nick_name"`
	Password         string `json:"password"`
	PhoneNumber      string `json:"phoneNumber"`
	TwoFactorEnabled int64  `json:"two_factor_enabled"`
	User2FAuthEnable int64  `json:"user_2fauth_enable"`
	IPRangeStart     string `json:"ip_range_start"`
	IPRangeEnd       string `json:"ip_range_end"`
	Role             int16  `json:"role"`
	Status           int64  `json:"status"`
	Remarks          string `json:"remarks"`
}

type GetAdminUsersRequest struct {
	RequestPagination
}

type GetAdminUsersResponse struct {
	Users []*AdminUserResponse `json:"users"`
	ResponsePagination
	UserNums int64 `json:"user_nums"`
}
type SearchAdminUsersResponse struct {
	Users []*AdminUserResponse `json:"users"`
	ResponsePaginationV2
}

type AdminPasswordChangeRequest struct {
	Secret    string `json:"secret" binding:"required"`
	NewSecret string `json:"new_secret" binding:"required"`
	TOTP      string `json:"totp"`
}

type AdminPasswordChangeResponse struct {
	Token           string `json:"token"`
	PasswordUpdated bool   `json:"password_updated"`
}
type AdminPermissionsResponse struct {
	Token string `json:"token"`
}

type AdminRole struct {
	AdminActions []AdminAction `json:"adminActions"`
	Status       int64         `json:"status"`
}

type AdminAction struct {
	ActionName   string          `json:"actionName"`
	AllowedApis  []AdminApiPath  `json:"allowedApis"`
	AllowedPages []AdminPagePath `json:"AllowedPages"`
	Status       int64           `json:"status"`
}
type AdminApiPath struct {
	ApiName string `json:"apiName"`
	ApiPath string `json:"apiPath"`
	Status  int64  `json:"status"`
}

type AdminPagePath struct {
	Id           int64           `json:"id"`
	PageName     string          `json:"pageName"`
	PagePath     string          `json:"pagePath"`
	FatherPageID int64           `json:"fatherPageID"`
	IsMenu       int64           `json:"isMenu"`
	SortPriority int64           `json:"sortPriority"`
	Childs       []AdminPagePath `json:"childs"`
	ChildsCount  int64           `json:"childsCount"`
	Status       int64           `json:"status"`
	AdminAPIsIDs []int64         `json:"adminAPIsIDs"`
	IsButton     int64           `json:"isButton"`
}

type AddAdminRoleRequest struct {
	OperationID     string `json:"operationID"`
	AdminRoleName   string `json:"adminRoleName" binding:"required"`
	AdminRoleStatus int64  `json:"status" binding:"required"`
	// AdminActionsIDs string `json:"adminActionsIDs"`
	AdminAPIsIDs         string `json:"admin_api_ids"`
	AdminPagesIDs        string `json:"admin_pages_ids"`
	CreateUser           string `json:"createUser"`
	AdminRoleDiscription string `json:"admin_role_discription"`
	AdminRoleRemarks     string `json:"admin_role_remarks"`
}

type AlterAdminRoleRequest struct {
	OperationID     string `json:"operationID"`
	AdminRoleID     int64  `json:"adminRoleID" binding:"required"`
	AdminRoleName   string `json:"adminRoleName"`
	AdminRoleStatus int64  `json:"status"`
	// AdminActionsIDs string `json:"adminActionsIDs"`
	AdminAPIsIDs         string `json:"admin_api_ids"`
	AdminPagesIDs        string `json:"admin_pages_ids"`
	AdminRoleDiscription string `json:"admin_role_discription"`
	AdminRoleRemarks     string `json:"admin_role_remarks"`
}

type AdminRoleResponse struct {
	AdminRoleID          int64  `json:"adminRoleID" binding:"required"`
	AdminRoleName        string `json:"adminRoleName"`
	AdminRoleStatus      int64  `json:"status"`
	AdminAPIsIDs         string `json:"admin_api_ids"`
	AdminPagesIDs        string `json:"admin_pages_ids"`
	AdminRoleDiscription string `json:"admin_role_discription"`
	CreateUser           string `json:"create_user"`
	CreateTime           int64  `json:"create_time"`
	UpdateUser           string `json:"update_user"`
	UpdateTime           int64  `json:"update_time"`
	AdminRoleRemarks     string `json:"admin_role_remarks"`
}

type GetAllAdminRolesRequest struct {
	RequestPagination
}
type SearchAminRolesRequest struct {
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	PageNumber  int64  `json:"page_number"`
	PageLimit   int64  `json:"page_limit"`
}
type GetAllAdminRolesResponse struct {
	AdminRoles []*AdminRoleResponse `json:"admin_roles"`
	ResponsePaginationV2
}

// AddApiInAdminRole Models
type AddApiInAdminRoleRequest struct {
	OperationID string `json:"operationID"`
	ApiName     string `json:"api_name" binding:"required"`
	ApiPath     string `json:"api_path" binding:"required"`
	Status      int8   `json:"status" `
	CreateUser  string `json:"createUser"`
}

type AlterApiInAdminRoleRequest struct {
	OperationID string `json:"operation_id"`
	ApiID       int64  `json:"api_id"`
	ApiName     string `json:"api_name"`
	Status      int8   `json:"status"`
	CreateUser  string `json:"create_user"`
}

type ApiInAdminRoleResponse struct {
	ApiID      int64  `json:"api_id"`
	ApiName    string `json:"api_name"`
	ApiPath    string `json:"api_path"`
	CreateUser string `json:"create_user"`
	UpdateUser string `json:"update_user"`
	Status     int32  `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

type GetAllApiInAdminRoleRequest struct {
	RequestPagination
}
type GetAllApiInAdminRoleResponse struct {
	ApiInAdminRoles []*ApiInAdminRoleResponse `json:"api_in_admin"`
	ResponsePaginationV2
}

// AddPageInAdminRole Models
type AddPageInAdminRoleRequest struct {
	OperationID  string `json:"operationID"`
	PageName     string `json:"page_name" binding:"required"`
	PagePath     string `json:"page_path" binding:"required"`
	Status       int8   `json:"status" `
	CreateUser   string `json:"createUser"`
	FatherPageID int64  `json:"father_page_id"`
	IsMenu       int64  `json:"is_menu"`
	SortPriority int64  `json:"sort_priority"`
	AdminAPIsIDs string `json:"admin_apis_ids"`
	IsButton     int64  `json:"is_button"`
}

type AlterPageInAdminRoleRequest struct {
	OperationID  string `json:"operation_id"`
	PageID       int64  `json:"page_id"`
	PageName     string `json:"page_name"`
	Status       int8   `json:"status" `
	CreateUser   string `json:"create_user"`
	FatherPageID int64  `json:"father_page_id"`
	IsMenu       int64  `json:"is_menu"`
	SortPriority int64  `json:"sort_priority"`
	AdminAPIsIDs string `json:"admin_apis_ids"`
	IsButton     int64  `json:"is_button"`
}

type PageInAdminRoleResponse struct {
	PageID       int64  `json:"page_id"`
	PageName     string `json:"page_name"`
	PagePath     string `json:"page_path"`
	CreateUser   string `json:"create_user"`
	UpdateUser   string `json:"update_user"`
	FatherPageID int64  `json:"father_page_id"`
	IsMenu       int64  `json:"is_menu"`
	SortPriority int64  `json:"sort_priority"`
	AdminAPIsIDs string `json:"admin_apis_ids"`
	IsButton     int64  `json:"is_button"`
	Status       int32  `json:"status"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time"`
}

type GetAllPageInAdminRoleRequest struct {
	FatherIDFilter int32 `form:"father_id_filter"`
	RequestPagination
}
type GetAllPageInAdminRoleResponse struct {
	PageInAdminRoles       []*PageInAdminRoleResponse `json:"pages_in_admin"`
	ApiInPageAdminRoles    []*ApiInAdminRoleResponse  `json:"api_in_admin"`
	FatherPageInAdminRoles []*PageInAdminRoleResponse `json:"father_pages_in_admin"`
	ResponsePaginationV2
}

// admin action model
type AddAdminActionRequest struct {
	OperationID       string `json:"operationID"`
	AdminActionName   string `json:"admin_action_name" binding:"required"`
	AdminActionStatus int64  `json:"status" binding:"required"`
	AdminAPIsIDs      string `json:"admin_api_ids"`
	AdminPagesIDs     string `json:"admin_pages_ids"`
	CreateUser        string `json:"createUser"`
}

type AlterAdminActionRequest struct {
	OperationID       string `json:"operationID"`
	AdminActionID     int64  `json:"admin_action_id" binding:"required"`
	AdminActionName   string `json:"admin_action_name"`
	AdminActionStatus int64  `json:"status"`
	AdminAPIsIDs      string `json:"admin_api_ids"`
	AdminPagesIDs     string `json:"admin_pages_ids"`
	CreateUser        string `json:"createUser"`
}

type AdminActionResponse struct {
	AdminActionID     int64  `json:"admin_action_id" binding:"required"`
	AdminActionName   string `json:"admin_action_name"`
	AdminActionStatus int64  `json:"status"`
	AdminAPIsIDs      string `json:"admin_api_ids"`
	AdminPagesIDs     string `json:"admin_pages_ids"`
	CreateUser        string `json:"create_user"`
}

type GetAllAdminActionsRequest struct {
	RequestPagination
}
type GetAllAdminActionsResponse struct {
	AdminActionss []*AdminActionResponse `json:"admin_actions"`
	ResponsePagination
	ActionsNums int64 `json:"actions_nums"`
}

type GetgAuthQrCodeResponse struct {
	GAuthSetupProvUri string `json:"gAuthSetupProvUri"`
	UsergAuthStatus   bool   `json:"usergAuthStatus"`
	GAuthAccountID    string `json:"gAuthAccountID"`
	GAuthKey          string `json:"gAuthKey"`
}

type AlterGAuthStatusResponse struct {
	UsergAuthStatus bool `json:"usergAuthStatus"`
}

type AlterGAuthStatusRequest struct {
	Status bool `json:"status"`
}

type AdminPermissionByAdminIDReq struct {
	OperationID string `json:"operationID"`
	AdminUserID string `json:"admin_user_id" binding:"required"`
}

type SearchAdminUsersRequest struct {
	AccountName           string `json:"account_name"`
	RoleID                int64  `json:"role_id"`
	GAuthStatus           int64  `json:"g_auth_status"`
	Status                int64  `json:"status"`
	Remarks               string `json:"remarks"`
	CreateTimeOrLastLogin int64  `json:"create_time_last_login"`
	IPAddress             string `json:"login_ip"`
	DateStart             int64  `json:"date_start"`
	DateEnd               int64  `json:"date_end"`
	PageNumber            int64  `json:"page_number"`
	PageLimit             int64  `json:"page_limit"`
}

type SearchApiAdminRoleRequest struct {
	ApiName    string `json:"api_name"`
	ApiPath    string `json:"api_path"`
	AddedBy    string `json:"added_by_user_id"`
	DateStart  int64  `json:"date_start"`
	DateEnd    int64  `json:"date_end"`
	PageNumber int64  `json:"page_number"`
	PageLimit  int64  `json:"page_limit"`
}
type SearchPageAdminRolesRequest struct {
	PageName   string `json:"page_name"`
	PagePath   string `json:"page_path"`
	AddedBy    string `json:"added_by_user_id"`
	DateStart  int64  `json:"date_start"`
	DateEnd    int64  `json:"date_end"`
	Status     int64  `json:"status"`
	PageNumber int64  `json:"page_number"`
	PageLimit  int64  `json:"page_limit"`
}

type SearchOperationLogsRequest struct {
	Operator   string `json:"operator"`
	Action     string `json:"action"`
	Executee   string `json:"executee"`
	DateStart  int64  `json:"date_start"`
	DateEnd    int64  `json:"date_end"`
	PageNumber int64  `json:"page_number"`
	PageLimit  int64  `json:"page_limit"`
}

type OperationLog struct {
	Operator   string `json:"operator"`
	Action     string `json:"action"`
	Payload    string `json:"payload"`
	OperatorIP string `json:"operatorIP"`
	Executee   string `json:"executee"`
	CreateTime int64  `json:"create_time"`
}

type SearchOperationLogsResponse struct {
	OperationLogs []OperationLog `json:"operationLogs"`
	ResponsePaginationV2
}

type GetURLReq struct {
	OperationID string `form:"operationID" binding:"required"`
	Type        uint   `form:"type" binding:"required"`
}

type GetURLResp struct {
	ID         uint              `json:"id"`
	Url        map[string]string `json:"url"`
	Status     int               `json:"status"`
	CreateTime int64             `json:"create_time"`
	CreateUser string            `json:"create_user"`
	UpdateTime int64             `json:"update_time"`
	UpdateUser string            `json:"update_user"`
	DeleteTime int64             `json:"delete_time"`
	DeleteUser string            `json:"delete_user"`
}

type SaveMePageUrlReq struct {
	OperationID string            `json:"operationID" binding:"required"`
	Type        uint              `json:"type"`
	Url         map[string]string `json:"url"`
}

type SaveMePageUrlResp struct {
	CommResp
}

type SwitchStatusReq struct {
	OperationID string `json:"operationID" binding:"required"`
	Type        uint   `json:"type"`
	Status      int    `json:"status"`
}

type SwitchStatusResp struct {
	CommResp
}
