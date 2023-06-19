package im_mysql_model

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	//init managers
	for k, v := range config.Config.Manager.AppManagerUid {
		user, err := GetUserByUserID(v)
		if err != nil {
			fmt.Println("GetUserByUserID failed ", err.Error(), v, user)
		} else {
			continue
		}
		var appMgr db.User
		appMgr.UserID = v
		appMgr.Password = config.Config.Manager.Secrets[k]
		if k == 0 {
			appMgr.Nickname = config.Config.Manager.AppSysNotificationName
		} else {
			appMgr.Nickname = "AppManager" + utils.IntToString(k+1)
		}
		appMgr.AppMangerLevel = constant.AppAdmin
		err = UserRegister(appMgr)
		if err != nil {
			fmt.Println("AppManager insert error", err.Error(), appMgr, "time: ", appMgr.Birth.Unix())
		}

	}
}

func UserRegister(user db.User) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	user.ID = int(db.DB.SnowFlake.Generate().Int64())
	user.Salt = utils.RandomString(10)
	newPasswordFirst := user.Password + user.Salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	if user.Password == "" {
		has = [16]byte{}
	}
	user.Password = fmt.Sprintf("%x", has)
	user.CreateTime = time.Now().Unix()
	if user.AppMangerLevel == 0 {
		user.AppMangerLevel = constant.AppOrdinaryUsers
	}
	if user.Birth.Unix() < 0 {
		user.Birth = utils.UnixSecondToTime(0)
	}
	err = dbConn.Table("users").Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(userID, reason, opUserID string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	updateInterface := make(map[string]interface{})
	updateInterface["delete_time"] = time.Now().Unix()
	updateInterface["delete_user"] = opUserID
	updateInterface["name"] = "Deleted Account"
	updateInterface["face_url"] = ""
	updateInterface["delete_reason"] = reason
	updateInterface["account_status"] = 2

	i = dbConn.Table("users").Where("user_id=?", userID).UpdateColumns(updateInterface).RowsAffected
	return i
}

func GetAllUserExcludeUser(users []string) ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var userList []db.User
	if users != nil && len(users) != 0 {
		err = dbConn.Table("users").Where("user_id NOT IN (?)", users).Where("delete_time=0").Find(&userList).Error
	} else {
		err = dbConn.Table("users").Where("delete_time=0").Find(&userList).Error
	}

	return userList, err
}

func GetAllUser() ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var userList []db.User
	err = dbConn.Table("users").Find(&userList).Error
	return userList, err
}

func GetAllUserCount() (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64
	err = dbConn.Table("users").Where("delete_time=0").Count(&count).Error
	return count, err
}

func GetAllOfficialFollowersData(officialID int64) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var idList []string
	if err != nil {
		return idList, err
	}
	err = dbConn.Table(db.OfficialFollow{}.TableName()).Select("user_id").Where("official_id=?", officialID).Find(&idList).Error
	return idList, err
}

func GetUserByUserID(userID string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("user_id=? and delete_time=0", userID).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUserIDAll(userID string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("user_id=?", userID).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUserIDEvenDeleted(userID string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("user_id=? ", userID).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByOfficialID(officialID int64) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("official_id", officialID).Where("delete_time", 0).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUserIdOrPhoneOrNicknameOrEmail(key string) ([]*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var userList []*db.User

	err = dbConn.Table("users").Where("user_id=? or phone_number=? or name=? or email=? and delete_time=0", key, key, key, key).Find(&userList).Error

	if err != nil {
		return nil, err
	}
	return userList, nil
}

func GetUserByPhoneNumber(phoneNumber string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("phone_number=? and delete_time=0", phoneNumber).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserNameByUserID(userID string) (string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return "", err
	}
	var user db.User
	err = dbConn.Table("users").Select("name").Where("user_id=?", userID).First(&user).Error
	if err != nil {
		return "", err
	}
	return user.Nickname, nil
}

func GetUserInfoByUserIDs(userIDs []string) ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var userList []db.User
	err = dbConn.Table("users").Where("user_id in (?) and delete_time=0", userIDs).Find(&userList).Error
	if err != nil {
		return nil, err
	}
	return userList, nil
}

func DeleteUserPhoneNumber(userId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}
	err = dbConn.Table("users").Where("user_id=?", userId).Update("phone_number", "").Error
	if err != nil {
		return false
	}
	return true
}

func UpdateUserInfo(user db.User) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	mapData := map[string]interface{}{}

	if user.SourceCode != "" {
		mapData["source_code"] = user.SourceCode
	}
	if user.SourceId != 0 {
		mapData["source_id"] = user.SourceId
	}
	mapData["remark"] = user.Remark
	mapData["email"] = user.Email
	if user.PhoneNumber != "" {
		mapData["phone_number"] = user.PhoneNumber
	}
	if user.FaceURL != "" {
		mapData["face_url"] = user.FaceURL

	}
	if user.Nickname != "" {
		mapData["name"] = user.Nickname
	}
	if user.Password != "" {
		mapData["password"] = user.Password
	}
	if user.Gender != 0 {
		mapData["gender"] = user.Gender
	}
	if !user.Birth.IsZero() {
		mapData["birth"] = user.Birth
	}

	err = dbConn.Table("users").Where("user_id=? and delete_time=0", user.UserID).Updates(&mapData).Error

	return err
}

func UpdateUserStatus(user db.User) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(user.TableName()).Where("user_id=? and delete_time=0", user.UserID).Debug().Updates(&user).Error

	return err
}

// func UpdateSuperUserStatus(user db.User) error {
// 	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
// 	if err != nil {
// 		return err
// 	}

// 	err = dbConn.Table("users").Where("user_id=? and delete_time=0", user.UserID).Update("add_friend_status", user.SuperUserStatus).Error

// 	return err
// }

func RemoveUserFaceUrl(user db.User) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table("users").Where("user_id=?", user.UserID).Update("face_url", "").Error
	return err
}

func UpdateUserInfoByMap(user db.User, m map[string]interface{}) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table("users").Where("user_id=? and delete_time=0", user.UserID).Updates(m).Error
	return err
}

func SelectAllUserID() ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var resultArr []string
	err = dbConn.Table("users").Pluck("user_id", &resultArr).Error
	if err != nil {
		return nil, err
	}
	return resultArr, nil
}

func SelectSomeUserID(userIDList []string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()

	if err != nil {
		return nil, err
	}
	var resultArr []string
	err = dbConn.Table("users").Where("user_id IN (?) ", userIDList).Pluck("user_id", &resultArr).Error

	if err != nil {
		return nil, err
	}
	return resultArr, nil
}

func GetSomeUserNameByUserId(userIdList []string) map[string]string {
	dbConn, _ := db.DB.MysqlDB.DefaultGormDB()
	var users []db.User
	dbConn.Table("users").Where("user_id IN (?) ", userIdList).Select([]string{"user_id", "name"}).Find(&users)

	userNameMap := make(map[string]string, len(users))
	for _, user := range users {
		userNameMap[user.UserID] = user.Nickname
	}
	return userNameMap
}

func GetUsers(showNumber, pageNumber int32) ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.User
	if err != nil {
		return users, err
	}

	err = dbConn.Table("users").Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&users).Error
	if err != nil {
		return users, err
	}
	return users, err
}

func GetUsersByWhere(where map[string]string, statusList []string, showNumber, pageNumber int32, orderBy string) ([]db.User, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.User
	var count int64
	if err != nil {
		return users, count, err
	}
	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"

	if orderBy != "" {
		direction := "DESC"
		sort := strings.Split(orderBy, ":")
		if len(sort) == 2 {
			if sort[1] == "asc" {
				direction = "ASC"
			}
		}
		col, ok := sortMap[sort[0]]
		if ok {
			orderByClause = fmt.Sprintf("%s %s ", col, direction)
		}

	}

	dbSub := dbConn.Table("users").Where("delete_time=0")
	if sourceId, ok := where["source_id"]; ok {
		if sourceId != "" {
			dbSub = dbSub.Where("source_id=?", sourceId)
		}
	}
	if sourceCode, ok := where["source_code"]; ok {
		if sourceCode != "" {
			dbSub = dbSub.Where("source_code like ?", "%"+sourceCode+"%")
		}
	}
	if remark, ok := where["remark"]; ok {
		if remark != "" {
			dbSub = dbSub.Where("remark like ?", "%"+remark+"%")
		}
	}
	if lastLoginDevice, ok := where["last_login_device"]; ok {
		if lastLoginDevice != "" && lastLoginDevice != "0" {
			if lastLoginDevice == "3" {
				dbSub = dbSub.Where("last_login_device IN (3,4,7)")
			} else {
				dbSub = dbSub.Where("last_login_device=?", lastLoginDevice)
			}
		}
	}
	if gender, ok := where["gender"]; ok {
		if gender != "" && gender != "0" {
			dbSub = dbSub.Where("gender=?", gender)
		}
	}
	if statusList != nil {
		for _, status := range statusList {
			if status == "1" {
				dbSub = dbSub.Where("status=1")
			} else if status == "2" {
				dbSub = dbSub.Where("status=2")
			} else if status == "3" {
				dbSub = dbSub.Where("video_status=1")
			} else if status == "4" {
				dbSub = dbSub.Where("video_status=2")
			} else if status == "5" {
				dbSub = dbSub.Where("audio_status=1")
			} else if status == "6" {
				dbSub = dbSub.Where("audio_status=2")
			}
		}
	}

	switch where["type"] {
	case "0":
		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				if endTime, ok := where["end_time"]; ok {
					if endTime != "" {
						dbSub = dbSub.Where("(last_login_time>=? and last_login_time<=?) or (create_time>=? and create_time<=?)", startTime, endTime, startTime, endTime)
					}
				}

			}
			break
		}

		// all
		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				dbSub = dbSub.Where("(last_login_time>=? or create_time>=?)", startTime, startTime)
			}
		}
		if endTime, ok := where["end_time"]; ok {
			if endTime != "" {
				dbSub = dbSub.Where("(last_login_time<=? or create_time<=?)", endTime, endTime)
			}
		}
		break
	case "1":
		// create
		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				dbSub = dbSub.Where("create_time>=?", startTime)
			}
		}
		if endTime, ok := where["end_time"]; ok {
			if endTime != "" {
				dbSub = dbSub.Where("create_time<=?", endTime)
			}
		}
		break
	case "2":
		// last login
		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				dbSub = dbSub.Where("last_login_time>=?", startTime)
			}
		}
		if endTime, ok := where["end_time"]; ok {
			if endTime != "" {
				dbSub = dbSub.Where("last_login_time<=?", endTime)
			}
		}
		break
	}

	if userId, ok := where["user_id"]; ok {
		if userId != "" {
			dbSub = dbSub.Where("user_id like ? or email like ? or name like ? or phone_number like ? or uuid like ? or login_ip like ?", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%")
		}
	}
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}

	err = dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&users).Error
	if err != nil {
		return users, count, err
	}
	return users, count, nil
}

func AddUser(userId, phoneNumber, name, password, opUser, sourceId, code, remark string, gender int32, email, faceURL string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	salt := utils.RandomString(10)
	newPasswordFirst := password + salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	newPassword := fmt.Sprintf("%x", has)

	phoneUser, _ := GetUserByPhone(phoneNumber)
	if phoneUser != nil {
		return errors.New("105")
	}

	if email != "" {
		emailUser, _ := GetUserByEmail(email)
		if emailUser != nil {
			return errors.New("106")
		}
	}

	user := db.User{
		PhoneNumber: phoneNumber,
		Birth:       time.Now(),
		UserID:      userId,
		Nickname:    name,
		Password:    newPassword,
		Email:       email,
		Salt:        salt,
		CreateUser:  opUser,
		CreateTime:  time.Now().Unix(),
		SourceId:    utils.StringToInt64(sourceId),
		SourceCode:  code,
		Remark:      remark,
		Gender:      gender,
		FaceURL:     faceURL,
	}
	result := dbConn.Table("users").Create(&user)
	defer func() {
		if result.Error == nil {
			dbConn.Table("users").Debug().Where("user_id=?", userId).Update("last_login_device", "0")
		}
	}()

	return result.Error
}

func GetUserByPhone(phoneNumber string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("phone_number=? and delete_time=0", phoneNumber).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func GetUserByEmail(email string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var user db.User
	err = dbConn.Table("users").Where("email=? and delete_time=0", email).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UserIsBlock(userId string) (bool, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false, err
	}
	var user db.BlackList
	rows := dbConn.Table("black_lists").Where("uid=?", userId).First(&user).RowsAffected
	if rows >= 1 {
		return true, nil
	}
	return false, nil
}

func BlockUser(userId, endDisableTime string) error {
	user, err := GetUserByUserID(userId)
	if err != nil || user.UserID == "" {
		return err
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	end, err := time.Parse("2006-01-02 15:04:05", endDisableTime)
	if err != nil {
		return err
	}
	if end.Before(time.Now()) {
		return constant.ErrDB
	}
	var blockUser db.BlackList
	dbConn.Table("black_lists").Where("uid=?", userId).First(&blockUser)
	if blockUser.UserId != "" {
		dbConn.Model(&blockUser).Where("uid=?", blockUser.UserId).Update("end_disable_time", end)
		return nil
	}
	blockUser = db.BlackList{
		UserId:           userId,
		BeginDisableTime: time.Now(),
		EndDisableTime:   end,
	}
	result := dbConn.Create(&blockUser)
	return result.Error
}

func UnBlockUser(userId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	result := dbConn.Where("uid=?", userId).Delete(&db.BlackList{})
	return result.Error
}

type BlockUserInfo struct {
	User             db.User
	BeginDisableTime time.Time
	EndDisableTime   time.Time
}

func GetBlockUserById(userId string) (BlockUserInfo, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var blockUserInfo BlockUserInfo
	blockUser := db.BlackList{
		UserId: userId,
	}
	if err != nil {
		return blockUserInfo, err
	}
	if err = dbConn.Table("black_lists").Where("uid=?", userId).Find(&blockUser).Error; err != nil {
		return blockUserInfo, err
	}
	user := db.User{
		UserID: blockUser.UserId,
	}
	if err := dbConn.Find(&user).Error; err != nil {
		return blockUserInfo, err
	}
	blockUserInfo.User.UserID = user.UserID
	blockUserInfo.User.FaceURL = user.FaceURL
	blockUserInfo.User.Nickname = user.Nickname
	blockUserInfo.BeginDisableTime = blockUser.BeginDisableTime
	blockUserInfo.EndDisableTime = blockUser.EndDisableTime
	return blockUserInfo, nil
}

func GetBlockUsers(showNumber, pageNumber int32) ([]BlockUserInfo, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var blockUserInfos []BlockUserInfo
	var blockUsers []db.BlackList
	if err != nil {
		return blockUserInfos, err
	}

	if err = dbConn.Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&blockUsers).Error; err != nil {
		return blockUserInfos, err
	}
	for _, blockUser := range blockUsers {
		var user db.User
		if err := dbConn.Table("users").Where("user_id=?", blockUser.UserId).First(&user).Error; err == nil {
			blockUserInfos = append(blockUserInfos, BlockUserInfo{
				User: db.User{
					UserID:   user.UserID,
					Nickname: user.Nickname,
					FaceURL:  user.FaceURL,
				},
				BeginDisableTime: blockUser.BeginDisableTime,
				EndDisableTime:   blockUser.EndDisableTime,
			})
		}
	}
	return blockUserInfos, nil
}

func GetUserByName(userName string, showNumber, pageNumber int32) ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.User
	if err != nil {
		return users, err
	}

	err = dbConn.Table("users").Where(fmt.Sprintf(" name like '%%%s%%' ", userName)).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&users).Error
	return users, err
}

func GetUserByNameOrUserId(userName string, showNumber, pageNumber int32) ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.User
	if err != nil {
		return users, err
	}

	err = dbConn.Table("users").Where(fmt.Sprintf(" name like '%%%s%%' ", userName)).Or("user_id = ?", userName).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&users).Error
	return users, err
}

func GetUsersCount(user db.User) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := dbConn.Table("users").Where(fmt.Sprintf(" name like '%%%s%%' ", user.Nickname)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetBlockUsersNumCount() (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := dbConn.Model(&db.BlackList{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetUserByUserIDList(userId []string) (map[string]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	dbConn.Debug()
	type UserListOne struct {
		UserID string
		Name   string
	}
	userList := []UserListOne{}
	userMap := map[string]string{}
	err = dbConn.Table("users").Select("user_id", "name").Where("delete_time=0").Where("user_id IN ?", userId).Find(&userList).Error
	if err != nil {
		return userMap, err
	}
	for _, v := range userList {
		if v.UserID != "" {
			userMap[v.UserID] = v.Name
		}
	}
	return userMap, nil
}
func GetValidUserIdListByUserIdList(userIdList []string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var users []db.User
	if err := dbConn.Table("users").Where("delete_time=0").Where("user_id in (?)", userIdList).Find(&users).Error; err != nil {
		return nil, err
	}

	var validUserIdList []string
	for _, user := range users {
		validUserIdList = append(validUserIdList, user.UserID)
	}
	return validUserIdList, nil
}

func GetUserIdByCondition(where map[string]string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var usersIds []string

	dbSub := dbConn.Table("users")
	if userId, ok := where["user_id"]; ok {
		if userId != "" {
			dbSub = dbSub.Where("user_id like ? or name like ? or phone_number like ?", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%")
		}
	}
	dbSub.Pluck("user_id", &usersIds)

	return usersIds
}

func GetUserIDListByExcludeIDList(excludeIdList []string) ([]db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []db.User{}, err
	}
	var users []db.User
	dbConn.Table("users").Where("user_id NOT IN (?)", excludeIdList).Pluck("user_id", &users)
	return users, nil
}

func GetUserIdByAllCondition(userKey string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var usersIds []string

	dbSub := dbConn.Table("users").Where("user_id like ? or name like ? or phone_number like ? or email like ?", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%")
	dbSub.Pluck("user_id", &usersIds)

	return usersIds
}

func GetUserByAllCondition(userKey string) []db.User {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []db.User{}
	}

	var usersIds []db.User

	dbSub := dbConn.Table("users").Where("user_id like ? or name like ? or phone_number like ? or email like ?", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%")
	dbSub.Find(&usersIds)

	return usersIds
}

func GetUsersByPhoneList(phoneList []string) []db.User {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var users []db.User
	if err != nil {
		return users
	}

	sql := "delete_time = 0"
	if len(phoneList) > 0 {
		sql += " and ("
		for _, phone := range phoneList {
			sql += "phone_number like '%" + phone + "%' or "
		}

		sql = strings.TrimRight(sql, " or ")
		sql += ")"
	}

	err = dbConn.Table(db.User{}.TableName()).Where(sql).Find(&users).Error
	if err != nil {
		return users
	}
	return users
}

func GetUsersThirdInfoByWhere(where map[string]string, showNumber, pageNumber int32) ([]map[string]interface{}, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, nil
	}

	var result []map[string]interface{}

	sub := dbConn.Table(db.User{}.TableName()).Debug().Joins("left join oauth_client on users.user_id = oauth_client.user_id").Where("users.delete_time=0")
	sub = sub.Group("users.user_id")
	if userId, ok := where["user_id"]; ok {
		if userId != "" {
			sub = sub.Where("users.user_id like ? or users.name like ? or users.phone_number like ? or users.email like ?", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%")
		}
	}
	if thirdType, ok := where["third_type"]; ok {
		if thirdType != "" {
			sub = sub.Where("oauth_client.type=?", thirdType)
		}
	}
	if thirdUserName, ok := where["third_name"]; ok {
		if thirdUserName != "" {
			sub = sub.Where("oauth_client.third_user_name=?", thirdUserName)
		}
	}

	err = sub.Offset(int((pageNumber-1)*showNumber)).Limit(int(showNumber)).Select("users.user_id", "users.name", "users.phone_number", "users.email").Scan(&result).Error

	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	if len(result) > 0 {
		userIdArray := make([]string, len(result))
		for _, v := range result {
			userIdArray = append(userIdArray, v["user_id"].(string))
		}

		thirdInfo := make([]map[string]interface{}, 0)
		dbConn.Table("oauth_client").Where("user_id in (?)", userIdArray).Scan(&thirdInfo)

		for i, v := range result {

			result[i]["apple"] = ""
			result[i]["facebook"] = ""
			result[i]["google"] = ""
			result[i]["official"] = ""
			for _, third := range thirdInfo {
				if v["user_id"] == third["user_id"] {
					switch third["type"] {
					case utils.StringToInt64("5"):
						result[i]["official"] = third["third_user_name"]
						continue
					case utils.StringToInt64("3"):
						result[i]["apple"] = third["third_user_name"]
						continue
					case utils.StringToInt64("1"):
						result[i]["facebook"] = third["third_user_name"]
						continue
					case utils.StringToInt64("2"):
						result[i]["google"] = third["third_user_name"]
						continue
					}
				}
			}
		}
	}

	return result, nil
}

func GetUsersThirdInfoCountByWhere(where map[string]string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, nil
	}

	var result int64

	sub := dbConn.Table(db.User{}.TableName()).Joins("left join oauth_client on users.user_id = oauth_client.user_id").Where("users.delete_time=0")
	sub = sub.Group("users.user_id")
	if userId, ok := where["user_id"]; ok {
		if userId != "" {
			sub = sub.Where("users.user_id like ? or users.name like ? or users.phone_number like ? or users.email like ?", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%", "%"+userId+"%")
		}
	}
	if thirdType, ok := where["third_type"]; ok {
		if thirdType != "" {
			sub = sub.Where("oauth_client.type=?", thirdType)
		}
	}
	if thirdUserName, ok := where["third_name"]; ok {
		if thirdUserName != "" {
			sub = sub.Where("oauth_client.third_user_name=?", thirdUserName)
		}
	}
	err = sub.Count(&result).Error
	if err != nil {
		return 0, err
	}
	return result, nil
}

func GetYouKnowUsersByContactList(userId string, contactList []string) []db.User {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	var users []db.User

	var friendList []string
	dbConn.Table(db.Friend{}.TableName()).Where("owner_user_id = ?", userId).Pluck("friend_user_id", &friendList)

	var excludeList []string
	dbConn.Table(db.ContactExclude{}.TableName()).Where("user_id = ?", userId).Pluck("to_user_id", &excludeList)

	// contact friend
	if len(contactList) > 0 {
		or := ""
		for _, s := range contactList {
			or += "users.phone_number = '" + s + "' or "
		}
		or = strings.TrimRight(or, " or ")

		sub := dbConn.Table(db.User{}.TableName()+" users").Limit(20).Where("users.delete_time=0").Where("users.user_id != ?", userId).
			Where(or).Order("users.create_time ASC").Group("users.user_id")
		if len(friendList) > 0 {
			sub = sub.Where("users.user_id not in (?)", friendList)
		}
		if len(excludeList) > 0 {
			sub = sub.Where("users.user_id not in (?)", excludeList)
		}
		err = sub.Find(&users).Error
		if err != nil {
			return users
		}
	}

	// friend's friend has same interest
	if len(users) < 20 && len(friendList) > 0 {
		userInterestList, _ := GetUserInterestList(userId)
		if len(userInterestList) == 0 {
			userInterestList = append(userInterestList, constant.InterestDefault)
		}

		var existUser []string
		if len(users) > 0 {
			for _, v := range users {
				existUser = append(existUser, v.UserID)
			}
		}

		var friendsFriend []string
		err = dbConn.Table(db.Friend{}.TableName()).Where("owner_user_id in (?)", friendList).Where("friend_user_id != ?", userId).Pluck("friend_user_id", &friendsFriend).Error
		if err != nil && len(friendsFriend) == 0 {
			return users
		}

		var interestUser []db.User
		iSub := dbConn.Table(db.User{}.TableName()+" users").Limit(20-len(users)).Where("users.delete_time=0").
			Where("users.user_id != ?", userId).Where("users.user_id in (?)", friendsFriend).Group("users.user_id")
		if len(friendList) > 0 {
			iSub = iSub.Where("users.user_id not in (?)", friendList)
		}
		if len(excludeList) > 0 {
			iSub = iSub.Where("users.user_id not in (?)", excludeList)
		}
		if len(existUser) > 0 {
			iSub = iSub.Where("users.user_id not in (?)", existUser)
		}

		err = iSub.Joins("left join interest_user ui on users.user_id = ui.user_id").Where("ui.interest_id in (?)", userInterestList).
			Order("users.create_time desc").
			Find(&interestUser).Error
		if err == nil {
			for _, user := range interestUser {
				users = append(users, user)
			}
		}
	}

	return users
}

func GetUsersByUserIDList(userIDList []string) ([]db.User, error) {
	if len(userIDList) == 0 {
		result := make([]db.User, 0)
		return result, nil
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var users []db.User
	err = dbConn.
		Table("users").
		Where("user_id", userIDList).
		Where("delete_time", 0).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUsersMapByUserIDList(userIDList []string) (map[string]db.User, error) {
	if len(userIDList) == 0 {
		usersMap := make(map[string]db.User, 0)
		return usersMap, nil
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var users []db.User
	err = dbConn.
		Table("users").
		Where("user_id", userIDList).
		Where("delete_time", 0).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	usersMap := make(map[string]db.User, len(users))
	for _, user := range users {
		usersMap[user.UserID] = user
	}

	return usersMap, nil
}

func GetDeletedUsersByWhere(where map[string]string, pageNumber, showNumber int32) ([]db.User, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}
	var result []db.User
	var count int64

	timeTypeMap := map[int]string{}
	timeTypeMap[1] = "create_time"
	timeTypeMap[2] = "delete_time"

	dbSub := dbConn.Table(db.User{}.TableName()).Where("delete_time != 0")

	if userKey, ok := where["user"]; ok {
		if userKey != "" {
			dbSub = dbSub.Where("user_id like ? or name like ? or phone_number like ? or email like ? or uuid like ? or login_ip like ?",
				"%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%", "%"+userKey+"%")
		}
	}
	if gender, ok := where["gender"]; ok {
		if gender != "" && gender != "0" {
			dbSub = dbSub.Where("gender=?", gender)
		}
	}
	if reason, ok := where["reason"]; ok {
		if reason != "" {
			dbSub = dbSub.Where("delete_reason like ?", "%"+reason+"%")
		}
	}
	//if location, ok := where["location"]; ok {
	//	if location != "" {
	//		dbSub = dbSub.Where("reason like ?", "%"+reason+"%")
	//	}
	//}

	if device, ok := where["last_login_device"]; ok {
		if device != "" && device != "0" {
			dbSub = dbSub.Where("last_login_device=?", device)
		}
	}
	if deletedBy, ok := where["deleted_by"]; ok {
		if deletedBy != "" {
			dbSub = dbSub.Where("delete_user like ?", "%"+deletedBy+"%")
		}
	}

	startTime := where["start_time"]
	endTime := where["end_time"]
	timeTypeStr := where["time_type"]
	timeType, _ := strconv.Atoi(timeTypeStr)

	if timeType == 0 {
		if startTime != "" || endTime != "" {
			var sql = ""
			var paramList []interface{}
			isFirst := true
			for _, v := range timeTypeMap {
				beenOr := false
				if startTime != "" {
					if isFirst {
						sql += fmt.Sprintf("%s>=?", v)
						paramList = append(paramList, startTime)
					} else {
						sql += fmt.Sprintf(" OR %s>=?", v)
						paramList = append(paramList, startTime)
						beenOr = true
					}
				}
				if endTime != "" {
					if isFirst || beenOr {
						sql += fmt.Sprintf(" AND %s<=?", v)
						paramList = append(paramList, endTime)
					} else {
						sql += fmt.Sprintf(" OR %s<=?", v)
						paramList = append(paramList, endTime)
					}
				}
				isFirst = false
			}

			log.Debug("", "time sql: ", sql, " param: ", paramList)
			dbSub = dbSub.Where(sql, paramList...)
		}
	} else {
		if startTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s>=?", timeTypeMap[timeType]), startTime)
		}
		if endTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s<=?", timeTypeMap[timeType]), endTime)
		}
	}

	dbSub = dbSub.Order("delete_time DESC")

	res := dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Debug().Find(&result)

	return result, count, res.Error

}
