package im_mysql_model

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	pbAdminCMS "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const AdminTableName = "admin_users"

func init() {
	//init managers
	for k, v := range config.Config.Manager.AppManagerUid {
		user, err := GetRegAdminUsrByUID(v)
		if err != nil {
			fmt.Println("GetUserByUserID failed ", err.Error(), v, user)
		} else {
			continue
		}
		var appMgr db.AdminUser
		appMgr.UserID = v
		appMgr.Password = config.Config.Manager.Secrets[k]
		if k == 0 {
			appMgr.Nickname = config.Config.Manager.AppSysNotificationName
		} else {
			appMgr.Nickname = "AppManager" + utils.IntToString(k+1)
		}
		appMgr.AppMangerLevel = constant.AppAdmin
		err = AdminUserRegister(appMgr)
		if err != nil {
			fmt.Println("AppManager insert error", err.Error(), appMgr, "time: ")
		}

	}

}

func GetRegAdminUsrByUID(userID string) (*db.AdminUser, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var r db.AdminUser
	return &r, dbConn.Table(AdminTableName).Where("user_id = ? and delete_time = 0",
		userID).Take(&r).Error
}

func AdminUserRegister(user db.AdminUser) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	user.Salt = utils.RandomString(10)
	user.Google2fSecretKey = strings.ToUpper(utils.RandomString(16))
	newPasswordFirst := user.Password + user.Salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	user.Password = fmt.Sprintf("%x", has)
	user.CreateTime = time.Now().Unix()
	user.IPRangeStart = "172.18.0.0"
	user.IPRangeEnd = "172.18.0.200"
	if user.AppMangerLevel == 0 {
		user.AppMangerLevel = constant.AppOrdinaryUsers
	}
	err = dbConn.Table(AdminTableName).Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func AddAdminUser(req *pbAdminCMS.AddAdminUserReq) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	salt := utils.RandomString(10)
	newPasswordFirst := req.Password + salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	newPassword := fmt.Sprintf("%x", has)
	google2fSecretKey := strings.ToUpper(utils.RandomString(16))

	user := db.AdminUser{
		UserID:            req.UserId,
		Nickname:          req.NickName,
		Name:              req.Name,
		Password:          newPassword,
		Google2fSecretKey: google2fSecretKey,
		Salt:              salt,
		CreateUser:        req.OpUserId,
		Role:              int64(req.Role),
		CreateTime:        time.Now().Unix(),
		Remarks:           req.Remarks,
		Status:            req.Status,
		User2FAuthEnable:  req.User2FAuthEnable,
		IPRangeStart:      req.IPRangeStart,
		IPRangeEnd:        req.IPRangeEnd,
	}
	result := dbConn.Table(AdminTableName).Create(&user)
	return result.Error
}

func DeleteAdminUser(userID string, OpUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	user := db.AdminUser{
		DeleteTime: time.Now().Unix(),
		DeleteUser: OpUserId,
	}
	if dbConn.Model(&user).Where("user_id = ?", userID).Updates(&user).RowsAffected == 0 {
		return 0
	}
	return 1
}

func AlterAdminUser(req *pbAdminCMS.AlterAdminUserRequest) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	user := db.AdminUser{
		UpdateUser: req.OpUserId,
		UpdateTime: time.Now().Unix(),
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.NickName != "" {
		user.Nickname = req.NickName
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.Password != "" {
		userDB, err := GetRegAdminUsrByUID(req.UserId)
		if err != nil {
			return 0
		}
		newPasswordFirst := req.Password + userDB.Salt
		passwordData := []byte(newPasswordFirst)
		has := md5.Sum(passwordData)
		password := fmt.Sprintf("%x", has)
		user.Password = password
	}
	user.TwoFactorEnabled = req.TwoFactorEnabled
	user.User2FAuthEnable = req.User2FAuthEnable
	user.IPRangeStart = req.IPRangeStart
	user.IPRangeEnd = req.IPRangeEnd
	user.Role = int64(req.Role)
	user.Remarks = req.Remarks
	user.LastLoginIP = req.LastLoginIP
	user.Status = req.Status
	log.NewError("Login Request Admin IP Address Alter", user.LastLoginIP)
	if dbConn.Model(&user).Where("user_id = ?", req.UserId).Updates(&user).RowsAffected == 0 {
		return 0
	}
	return 1
}

func AlterAdminUserLoginIP(userID, loginIP string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	user := db.AdminUser{
		UpdateUser:    userID,
		LastLoginIP:   loginIP,
		LastLoginTime: time.Now().Unix(),
	}

	log.NewError("Login Request Admin IP Address Alter", user.LastLoginIP)
	if dbConn.Model(&user).Where("user_id = ?", userID).Updates(&user).RowsAffected == 0 {
		return 0
	}
	return 1
}
func AlterUserGAuthSatus(req *pbAdminCMS.AlterGAuthStatusReq) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(db.AdminUser{}.TableName()).Where("user_id=?", req.UserId).Update("user_two_factor_control_status", req.UserGAuthStatus).Error
	return err
}

func GetAdminUsers(showNumber, pageNumber int32) ([]db.AdminUser, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.AdminUser
	if err != nil {
		return users, err
	}

	err = dbConn.Table(AdminTableName).Where("delete_time=0").Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&users).Error
	if err != nil {
		return users, err
	}
	return users, err
}

func GetAdminsByIds(ids []string) ([]db.AdminUser, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.AdminUser
	if err != nil {
		return users, err
	}

	err = dbConn.Table(AdminTableName).Where("delete_time=0").Where("user_id IN (?)", ids).Find(&users).Error
	if err != nil {
		return users, err
	}
	return users, err
}

func GetRowsCountCount(tableName string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64
	if err := dbConn.Table(tableName).Where("delete_time=0").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func UpdateAdminUserInfo(user db.AdminUser) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table(AdminTableName).Where("user_id=? and delete_time=0", user.UserID).Updates(&user).Error

	return err
}

func UpdateAdminUserPassword(user db.AdminUser, userID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(AdminTableName).Where("user_id=? and delete_time=0", userID).Updates(&user).Error

	return err
}
func GetAdminPermission(user db.AdminUser) pbAdminCMS.AdminRole {

	adminPermission := pbAdminCMS.AdminRole{}
	adminRole, err := GetAdminRole(user)
	if err != nil || adminRole == nil {
		return adminPermission
	}
	// adminActions := GetAdminActions(*adminRole)

	// for _, adminAction := range adminActions {
	// adminAction_pbACMS := pbAdminCMS.AdminAction{}
	// adminAction_pbACMS.ActionName = adminAction.AdminActionName
	// adminAction_pbACMS.Id = adminAction.AdminActionID

	apiIDs := adminRole.AdminAPIsIDs
	var adminAPIIds []int
	if err := json.Unmarshal([]byte(apiIDs), &adminAPIIds); err != nil {
		log.NewError(utils.GetSelfFuncName(), "Parsing String of IDs to int slice", err.Error())
	}
	admin_apis_DBM := GetAdminAPIs(adminAPIIds)
	adminPermission.AllowedApis = convertAdminAPIModel(admin_apis_DBM)

	pagesIDs := adminRole.AdminPagesIDs
	var adminPagesIds []int
	if err := json.Unmarshal([]byte(pagesIDs), &adminPagesIds); err != nil {
		log.NewError(utils.GetSelfFuncName(), "Parsing String of IDs to int slice", err.Error())
	}
	admin_pages_DBM := GetAdminPages(adminPagesIds)
	adminPermission.AllowedPages = ConvertAdminPagesModel(admin_pages_DBM)

	// adminPermission.AdminActions = append(adminPermission.AdminActions, &adminAction_pbACMS)
	// }
	return adminPermission

}

func GetAdminRole(user db.AdminUser) (*db.AdminRole, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	adminRole := db.AdminRole{}

	err = dbConn.Table(adminRole.TableName()).Where("admin_role_id=? and delete_time=0", user.Role).Take(&adminRole).Error
	if err != nil {
		return nil, err
	}

	return &adminRole, err
}

// func GetAdminActions(adminRole db.AdminRole) []db.AdminActions {
// 	var adminActions []db.AdminActions
// 	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
// 	if err != nil {
// 		log.NewError(utils.GetSelfFuncName(), "Db connection", err.Error())
// 		return adminActions
// 	}
// 	var adminActionsIds []int
// 	if err := json.Unmarshal([]byte(adminRole.AdminActionsIDs), &adminActionsIds); err != nil {
// 		log.NewError(utils.GetSelfFuncName(), "Parsing String of IDs to int slice", err.Error())
// 		return adminActions
// 	}
// 	err = dbConn.Table(db.AdminActions{}.TableName()).Where("admin_action_id in ? and delete_time=0", adminActionsIds).Find(&adminActions).Error
// 	if err != nil {
// 		return adminActions
// 	}
// 	return adminActions
// }

func GetAdminAPIs(api_ids_intSlice []int) []db.AdminAPIs {
	var adminAPIs []db.AdminAPIs
	if len(api_ids_intSlice) == 0 {
		return adminAPIs
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Db connection", err.Error())
		return adminAPIs
	}
	err = dbConn.Table(db.AdminAPIs{}.TableName()).Where("api_id in ? and delete_time=0", api_ids_intSlice).Find(&adminAPIs).Error
	if err != nil {
		return adminAPIs
	}
	return adminAPIs
}
func convertAdminAPIModel(adminAPIs_DBM []db.AdminAPIs) []*pbAdminCMS.AdminApiPath {
	var adminAPIs []*pbAdminCMS.AdminApiPath
	for _, adminAPI_DBM := range adminAPIs_DBM {
		adminAPI := pbAdminCMS.AdminApiPath{}
		adminAPI.ApiName = adminAPI_DBM.ApiName
		adminAPI.ApiPath = adminAPI_DBM.ApiPath
		adminAPI.Id = adminAPI_DBM.ApiID
		adminAPIs = append(adminAPIs, &adminAPI)
	}
	return adminAPIs
}

func GetAdminPages(pages_ids_intSlice []int) []db.AdminPages {
	var adminPages []db.AdminPages
	if len(pages_ids_intSlice) == 0 {
		return adminPages
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Db connection", err.Error())
		return adminPages
	}

	err = dbConn.Table(db.AdminPages{}.TableName()).Where("page_id in ? and delete_time=0", pages_ids_intSlice).Find(&adminPages).Error
	if err != nil {
		return adminPages
	}

	return adminPages
}

func ConvertAdminPagesModel(adminPages_DBM []db.AdminPages) []*pbAdminCMS.AdminPagePath {
	var adminPages []*pbAdminCMS.AdminPagePath
	for _, adminAPI_DBM := range adminPages_DBM {
		adminAPI := pbAdminCMS.AdminPagePath{}
		adminAPI.PageName = adminAPI_DBM.PageName
		adminAPI.PagePath = adminAPI_DBM.PagePath
		adminAPI.Id = adminAPI_DBM.PageID
		adminAPI.FatherPageID = adminAPI_DBM.FatherPageID
		adminAPI.IsMenu = adminAPI_DBM.IsMenu
		adminAPI.SortPriority = adminAPI_DBM.SortPriority
		adminAPI.IsButton = adminAPI_DBM.IsButton
		var adminAPIIds []int64
		if adminAPI_DBM.AdminAPIsIDs != "" {
			if err := json.Unmarshal([]byte(adminAPI_DBM.AdminAPIsIDs), &adminAPIIds); err != nil {
				log.NewError(utils.GetSelfFuncName(), "Parsing String of IDs to int slice", err.Error())
			}
			adminAPI.AdminAPIsIDs = adminAPIIds
		}

		adminPages = append(adminPages, &adminAPI)
	}
	return adminPages
}

func GetAdminAllRolesPermissions() []pbAdminCMS.AdminRole {
	var adminRoles []db.AdminRole
	var adminPermissionSlice []pbAdminCMS.AdminRole
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return adminPermissionSlice
	}

	err = dbConn.Table(db.AdminRole{}.TableName()).Where("delete_time=0").Find(&adminRoles).Error
	if err != nil {
		return adminPermissionSlice
	}
	for _, adminRole := range adminRoles {
		adminPermission := pbAdminCMS.AdminRole{}
		adminPermission.Id = adminRole.AdminRoleID
		// adminActions := GetAdminActions(adminRole)

		// for _, adminAction := range adminActions {
		// adminAction_pbACMS := pbAdminCMS.AdminAction{}
		// adminAction_pbACMS.ActionName = adminAction.AdminActionName
		// adminAction_pbACMS.Id = adminAction.AdminActionID

		apiIDs := adminRole.AdminAPIsIDs
		var adminAPIIds []int
		if err := json.Unmarshal([]byte(apiIDs), &adminAPIIds); err != nil {
			log.NewError(utils.GetSelfFuncName(), "Parsing String of IDs to int slice", err.Error())
		}
		admin_apis_DBM := GetAdminAPIs(adminAPIIds)
		adminPermission.AllowedApis = convertAdminAPIModel(admin_apis_DBM)

		pagesIDs := adminRole.AdminPagesIDs
		var adminPagesIds []int
		if err := json.Unmarshal([]byte(pagesIDs), &adminPagesIds); err != nil {
			log.NewError(utils.GetSelfFuncName(), "Parsing String of IDs to int slice", err.Error())
		}
		admin_pages_DBM := GetAdminPages(adminPagesIds)
		adminPermission.AllowedPages = ConvertAdminPagesModel(admin_pages_DBM)

		// adminPermission.AdminActions = append(adminPermission.AdminActions, &adminAction_pbACMS)
		// }
		adminPermissionSlice = append(adminPermissionSlice, adminPermission)
	}

	return adminPermissionSlice
}

// Admin Roles CRUD Operations
func AddAdminRole(req *pbAdminCMS.AddAdminRoleRequest) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	adminRole := db.AdminRole{}
	adminRole.AdminRoleName = req.AdminRoleName
	adminRole.AdminAPIsIDs = req.AdminAPIsIDs
	adminRole.AdminPagesIDs = req.AdminPagesIDs
	adminRole.Status = int(req.Status)
	adminRole.CreateUser = req.CreateUser
	adminRole.CreateTime = time.Now().Unix()
	adminRole.AdminRoleDiscription = req.AdminRoleDiscription
	adminRole.AdminRoleRemarks = req.AdminRoleRemarks

	result := dbConn.Table(adminRole.TableName()).Create(&adminRole)
	return result.Error
}

func AlterAdminRole(req *pbAdminCMS.AlterAdminRoleRequest) int {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	adminRole := db.AdminRole{}
	adminRole.AdminRoleName = req.AdminRoleName
	adminRole.AdminAPIsIDs = req.AdminAPIsIDs
	adminRole.AdminPagesIDs = req.AdminPagesIDs
	adminRole.Status = int(req.Status)
	adminRole.UpdateUser = req.UpdateUser
	adminRole.UpdateTime = time.Now().Unix()
	adminRole.AdminRoleDiscription = req.AdminRoleDiscription
	adminRole.AdminRoleRemarks = req.AdminRoleRemarks

	if dbConn.Model(&adminRole).Where("admin_role_id = ?", req.AdminRoleID).Updates(&adminRole).RowsAffected == 0 {
		log.NewError(utils.GetSelfFuncName(), "Sandman # 1 No record Updated", adminRole)
		return 0
	}
	return 1
}

func DeletedminRole(req *pbAdminCMS.AlterAdminRoleRequest) int {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	adminRole := db.AdminRole{}
	adminRole.Status = 2
	adminRole.DeleteUser = req.CreateUser
	adminRole.DeleteTime = time.Now().Unix()

	if dbConn.Model(&adminRole).Where("admin_role_id = ?", req.AdminRoleID).Updates(&adminRole).RowsAffected == 0 {
		log.NewError(utils.GetSelfFuncName(), "Sandman # 1 No record Updated", adminRole)
		return 0
	}
	return 1
}
func CheckAdminRoleAssigned(AdminRoleID int64) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	var assignedRoleCount int64 = 0
	dbConn.Table(db.AdminUser{}.TableName()).Where("role = ?", AdminRoleID).Count(&assignedRoleCount)
	return assignedRoleCount
}

func GetAllAdminRoles(showNumber, pageNumber int32) ([]db.AdminRole, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var adminRoles []db.AdminRole
	if err != nil {
		return adminRoles, err
	}

	err = dbConn.Table(db.AdminRole{}.TableName()).Where("delete_time = ?", 0).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&adminRoles).Error
	if err != nil {
		return adminRoles, err
	}
	return adminRoles, err
}
func SearchAminRoles(RoleName, Description string, pageNumber, pageLimit int64) ([]db.AdminRole, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var totalRecordsInDb int64 = 0
	var adminRoles []db.AdminRole
	if err != nil {
		return adminRoles, totalRecordsInDb, err
	}
	searchQuery := "delete_time = 0 "
	if RoleName != "" {
		searchQuery = searchQuery + " and admin_role_name like '%" + RoleName + "%'"
	}
	if Description != "" {
		searchQuery = searchQuery + " and admin_role_discription like '%" + Description + "%'"
	}

	err = dbConn.Table(db.AdminRole{}.TableName()).Where(searchQuery).Count(&totalRecordsInDb).Limit(int(pageLimit)).Offset(int(pageLimit * (pageNumber - 1))).Find(&adminRoles).Error
	if err != nil {
		return adminRoles, totalRecordsInDb, err
	}
	return adminRoles, totalRecordsInDb, err
}

// API Admin Roles CRUD Operations
func AddApiAdminRole(req *pbAdminCMS.AddApiAdminRoleRequest) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	apiAdminRole := db.AdminAPIs{}
	apiAdminRole.ApiName = req.ApiName
	apiAdminRole.ApiPath = req.ApiPath
	apiAdminRole.Status = int(req.Status)
	apiAdminRole.CreateUser = req.CreateUser
	apiAdminRole.CreateTime = time.Now().Unix()

	result := dbConn.Table(apiAdminRole.TableName()).Create(&apiAdminRole)
	return result.Error
}

func AlterApiAdminRole(req *pbAdminCMS.AlterApiAdminRoleRequest) int {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	apiAdminRole := db.AdminAPIs{}
	apiAdminRole.ApiName = req.ApiName
	apiAdminRole.ApiPath = req.ApiPath
	apiAdminRole.Status = int(req.Status)
	apiAdminRole.UpdateUser = req.CreateUser
	apiAdminRole.UpdateTime = time.Now().Unix()

	if dbConn.Model(&apiAdminRole).Where("api_id = ?", req.ApiID).Updates(&apiAdminRole).RowsAffected == 0 {
		log.NewError(utils.GetSelfFuncName(), "No record Updated", apiAdminRole)
		return 0
	}
	return 1
}

func DeleteApiAdminRole(req *pbAdminCMS.AlterApiAdminRoleRequest) int {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	apiAdminRole := db.AdminAPIs{}
	apiAdminRole.Status = 2
	apiAdminRole.DeleteUser = req.CreateUser
	apiAdminRole.DeleteTime = time.Now().Unix()

	if dbConn.Model(&apiAdminRole).Where("api_id = ?", req.ApiID).Updates(&apiAdminRole).RowsAffected == 0 {
		log.NewError(utils.GetSelfFuncName(), "No record Deleted", apiAdminRole)
		return 0
	}
	return 1
}

func GetAllApiAdminRoles(showNumber, pageNumber int32) ([]db.AdminAPIs, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var totalRecordsInDb int64 = 0
	var apiAdminRole []db.AdminAPIs
	if err != nil {
		return apiAdminRole, totalRecordsInDb, err
	}

	err = dbConn.Table(db.AdminAPIs{}.TableName()).Where("delete_time = ?", 0).Count(&totalRecordsInDb).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&apiAdminRole).Error
	if err != nil {
		return apiAdminRole, totalRecordsInDb, err
	}
	return apiAdminRole, totalRecordsInDb, err
}

func SearchApiAdminRoles(req *pbAdminCMS.SearchApiAdminRoleRequest) ([]db.AdminAPIs, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var totalRecordsInDb int64 = 0
	var apiAdminRole []db.AdminAPIs
	if err != nil {
		return apiAdminRole, totalRecordsInDb, err
	}
	searchQuery := " delete_time = 0 "
	if req.ApiName != "" {
		searchQuery = searchQuery + " and api_name like '%" + req.ApiName + "%'"
	}
	if req.ApiPath != "" {
		searchQuery = searchQuery + " and api_path like '%" + req.ApiPath + "%'"
	}
	if req.AddedBy != "" {
		searchQuery = searchQuery + " and create_user like '%" + req.AddedBy + "%'"
	}
	if req.DateStart > 0 && req.DateEnd > 0 {
		searchQuery = searchQuery + " and create_time >= " + fmt.Sprint(req.DateStart) + " and create_time <= " + fmt.Sprint(req.DateEnd)
	}
	log.NewError(utils.GetSelfFuncName(), "Search Admin API # 1", searchQuery)
	err = dbConn.Table(db.AdminAPIs{}.TableName()).Where(searchQuery).Count(&totalRecordsInDb).Limit(int(req.PageLimit)).Offset(int(req.PageLimit * (req.PageNumber - 1))).Find(&apiAdminRole).Error
	if err != nil {
		return apiAdminRole, totalRecordsInDb, err
	}
	return apiAdminRole, totalRecordsInDb, err
}

// Page Admin Roles CRUD Operations
func AddPageAdminRole(req *pbAdminCMS.AddPageAdminRoleRequest) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	pageAdminRole := db.AdminPages{}
	pageAdminRole.PageName = req.PageName
	pageAdminRole.PagePath = req.PagePath
	pageAdminRole.Status = int(req.Status)
	pageAdminRole.CreateUser = req.CreateUser
	pageAdminRole.CreateTime = time.Now().Unix()
	pageAdminRole.FatherPageID = req.FatherPageID
	pageAdminRole.IsButton = req.IsButton
	pageAdminRole.IsMenu = req.IsMenu
	pageAdminRole.SortPriority = req.SortPriority
	pageAdminRole.AdminAPIsIDs = req.AdminAPIsIDs

	result := dbConn.Table(pageAdminRole.TableName()).Create(&pageAdminRole)
	return result.Error
}

func AlterPageAdminRole(req *pbAdminCMS.AlterPageAdminRoleRequest) int {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	pageAdminRole := db.AdminPages{}
	pageAdminRole.PageName = req.PageName
	pageAdminRole.PagePath = req.PagePath
	pageAdminRole.Status = int(req.Status)
	pageAdminRole.UpdateUser = req.CreateUser
	pageAdminRole.UpdateTime = time.Now().Unix()
	pageAdminRole.FatherPageID = req.FatherPageID
	pageAdminRole.IsButton = req.IsButton
	pageAdminRole.IsMenu = req.IsMenu
	pageAdminRole.SortPriority = req.SortPriority
	pageAdminRole.AdminAPIsIDs = req.AdminAPIsIDs

	if dbConn.Model(&pageAdminRole).Where("page_id = ?", req.PageID).Updates(&pageAdminRole).RowsAffected == 0 {
		log.NewError(utils.GetSelfFuncName(), "No record Updated", pageAdminRole)
		return 0
	}
	return 1
}

func DeletePageAdminRole(req *pbAdminCMS.AlterPageAdminRoleRequest) int {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	pageAdminRole := db.AdminPages{}
	pageAdminRole.Status = 2
	pageAdminRole.DeleteUser = req.CreateUser
	pageAdminRole.DeleteTime = time.Now().Unix()

	if dbConn.Model(&pageAdminRole).Where("page_id = ?", req.PageID).Updates(&pageAdminRole).RowsAffected == 0 {
		log.NewError(utils.GetSelfFuncName(), "No record Deleted", pageAdminRole)
		return 0
	}
	return 1
}

func GetAllPageAdminRoles(fatherIdFilter, showNumber, pageNumber int32) ([]db.AdminPages, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var pageAdminRoles []db.AdminPages
	var totalRecordsInDb int64 = 0
	if err != nil {
		return pageAdminRoles, totalRecordsInDb, err
	}
	if fatherIdFilter >= 0 {
		err = dbConn.Table(db.AdminPages{}.TableName()).Where("delete_time = ? and father_page_id = ?", 0, fatherIdFilter).Find(&pageAdminRoles).Error
	} else {
		err = dbConn.Table(db.AdminPages{}.TableName()).Where("delete_time = ? ", 0).Count(&totalRecordsInDb).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&pageAdminRoles).Error
	}
	if err != nil {
		return pageAdminRoles, totalRecordsInDb, err
	}
	return pageAdminRoles, totalRecordsInDb, err
}

func SearchPageAdminRoles(req *pbAdminCMS.SearchPageAdminRolesRequest) ([]db.AdminPages, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var pageAdminRoles []db.AdminPages
	var totalRecordsInDb int64 = 0
	if err != nil {
		return pageAdminRoles, totalRecordsInDb, err
	}
	searchQuery := "delete_time = 0 " //and create_user like '%" + req.AddedBy + "%'"
	if req.PageName != "" {
		searchQuery = searchQuery + " and page_name like '%" + req.PageName + "%'"
	}
	if req.PagePath != "" {
		searchQuery = searchQuery + " and page_path like '%" + req.PagePath + "%'"
	}
	if req.DateStart > 0 && req.DateEnd > 0 {
		searchQuery = searchQuery + " and create_time >= " + fmt.Sprint(req.DateStart) + " and create_time <= " + fmt.Sprint(req.DateEnd)
	}
	if req.Status > 0 {
		searchQuery = searchQuery + " and status = " + fmt.Sprint(req.Status)
	}
	log.NewError(utils.GetSelfFuncName(), "Search Admin Page # 1", searchQuery)
	err = dbConn.Table(db.AdminPages{}.TableName()).Where(searchQuery).Count(&totalRecordsInDb).Limit(int(req.PageLimit)).Offset(int(req.PageLimit * (req.PageNumber - 1))).Find(&pageAdminRoles).Error
	if err != nil {
		return pageAdminRoles, totalRecordsInDb, err
	}
	return pageAdminRoles, totalRecordsInDb, err
}

func SearchAdminUsers(AccountName string, RoleID, GAuthStatus, Status int32, IPAddress string, startDate, endDate, PageNumber, PageLimit, CreateTimeOrLastLogin int64, remarks string) ([]db.AdminUser, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var totalRecordsInDb int64 = 0
	var users []db.AdminUser
	if err != nil {
		return users, totalRecordsInDb, err
	}
	var searchQuery string
	searchQuery = " delete_time = 0 "
	if IPAddress != "" {
		searchQuery = searchQuery + " and last_login_ip like '%" + IPAddress + "%'"
	}
	if AccountName != "" {
		searchQuery = searchQuery + " and name like '%" + AccountName + "%'"
	}
	if remarks != "" {
		searchQuery = searchQuery + " and remarks like '%" + remarks + "%'"
	}
	if RoleID != 0 {
		searchQuery = searchQuery + " and role = " + fmt.Sprint(RoleID)
	}
	if GAuthStatus != 0 {
		searchQuery = searchQuery + " and user_two_factor_control_status = " + fmt.Sprint(GAuthStatus)
	}
	if Status != 0 {
		searchQuery = searchQuery + " and status = " + fmt.Sprint(Status)
	}
	if startDate > 0 && endDate > 0 {
		if CreateTimeOrLastLogin == 1 {
			searchQuery = searchQuery + " and create_time >= " + fmt.Sprint(startDate) + " and create_time <= " + fmt.Sprint(endDate)
		} else {
			searchQuery = searchQuery + " and last_login_time >= " + fmt.Sprint(startDate) + " and last_login_time <= " + fmt.Sprint(endDate)
		}
	}
	err = dbConn.Table(AdminTableName).Where(searchQuery).Count(&totalRecordsInDb).Limit(int(PageLimit)).Offset(int(PageLimit * (PageNumber - 1))).Find(&users).Error
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Search Admin Users # 2", err.Error())
		return users, totalRecordsInDb, err
	}
	return users, totalRecordsInDb, err
}

// Admin Actions CRUD Operations
// func AddAdminAction(req *pbAdminCMS.AddAdminActionRequest) error {
// 	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
// 	if err != nil {
// 		return err
// 	}
// 	adminAction := db.AdminActions{}
// 	adminAction.AdminActionName = req.AdminActionName
// 	adminAction.AdminAPIsIDs = req.AdminAPIsIDs
// 	adminAction.AdminPagesIDs = req.AdminPagesIDs
// 	adminAction.Status = int(req.Status)
// 	adminAction.CreateUser = req.CreateUser
// 	adminAction.CreateTime = time.Now().Unix()

// 	result := dbConn.Table(adminAction.TableName()).Create(&adminAction)
// 	return result.Error
// }

// func AlterAdminAction(req *pbAdminCMS.AlterAdminActionRequest) int {
// 	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
// 	if err != nil {
// 		return 0
// 	}
// 	adminAction := db.AdminActions{}
// 	adminAction.AdminActionName = req.AdminActionName
// 	adminAction.AdminAPIsIDs = req.AdminAPIsIDs
// 	adminAction.AdminPagesIDs = req.AdminPagesIDs
// 	adminAction.Status = int(req.Status)
// 	adminAction.UpdateUser = req.CreateUser
// 	adminAction.UpdateTime = time.Now().Unix()

// 	if dbConn.Model(&adminAction).Where("admin_action_id = ?", req.AdminActionID).Updates(&adminAction).RowsAffected == 0 {
// 		log.NewError(utils.GetSelfFuncName(), "Sandman # 1 No record Updated", adminAction)
// 		return 0
// 	}
// 	return 1
// }

// func DeletedminAction(req *pbAdminCMS.AlterAdminActionRequest) int {
// 	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
// 	if err != nil {
// 		return 0
// 	}
// 	adminAction := db.AdminActions{}
// 	adminAction.Status = 2
// 	adminAction.DeleteUser = req.CreateUser
// 	adminAction.DeleteTime = time.Now().Unix()

// 	if dbConn.Model(&adminAction).Where("admin_action_id = ?", req.AdminActionID).Updates(&adminAction).RowsAffected == 0 {
// 		log.NewError(utils.GetSelfFuncName(), "Sandman # 1 No record Updated", adminAction)
// 		return 0
// 	}
// 	return 1
// }

// func GetAllAdminAction(showNumber, pageNumber int32) ([]db.AdminActions, error) {
// 	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
// 	var adminActions []db.AdminActions
// 	if err != nil {
// 		return adminActions, err
// 	}

// 	err = dbConn.Table(db.AdminActions{}.TableName()).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&adminActions).Error
// 	if err != nil {
// 		return adminActions, err
// 	}
// 	return adminActions, err
// }
