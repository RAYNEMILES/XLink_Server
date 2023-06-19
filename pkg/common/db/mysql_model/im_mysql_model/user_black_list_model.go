package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/utils"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func InsertInToUserBlackList(black db.Black) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	black.CreateTime = time.Now()
	err = dbConn.Table("blacks").Create(black).Error
	return err
}

//type Black struct {
//	OwnerUserID    string    `gorm:"column:owner_user_id;primaryKey;"`
//	BlockUserID    string    `gorm:"column:block_user_id;primaryKey;"`
//	CreateTime     time.Time `gorm:"column:create_time"`
//	AddSource      int32     `gorm:"column:add_source"`
//	OperatorUserID int32     `gorm:"column:operator_user_id"`
//	Ex             string    `gorm:"column:ex"`
//}

func CheckBlack(ownerUserID, blockUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	var black db.Black
	err = dbConn.Table("blacks").Where("owner_user_id=? and block_user_id=?", ownerUserID, blockUserID).Find(&black).Error
	return err
}

func RemoveBlackList(ownerUserID, blockUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table("blacks").Where("owner_user_id=? and block_user_id=?", ownerUserID, blockUserID).Delete(db.Black{}).Error
	return utils.Wrap(err, "RemoveBlackList failed")
}

func GetBlackListByUserID(ownerUserID string) ([]db.Black, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var blackListUsersInfo []db.Black
	err = dbConn.Table("blacks").Where("owner_user_id=?", ownerUserID).Find(&blackListUsersInfo).Error
	if err != nil {
		return nil, err
	}
	return blackListUsersInfo, nil
}

func GetBlackIDListByUserID(ownerUserID string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var blackIDList []string
	err = dbConn.Table("blacks").Where("owner_user_id=?", ownerUserID).Pluck("block_user_id", &blackIDList).Error
	if err != nil {
		return nil, err
	}
	return blackIDList, nil
}

func GetDistinctUsersFromBlackMoments() (blackListOwner []db.BlackForMoment, err error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return blackListOwner, err
	}
	err = dbConn.Table(db.BlackForMoment{}.TableName()).Distinct("owner_user_id").Find(&blackListOwner).Error
	return blackListOwner, err
}

func InsertUserBlackList(blacks []db.Black) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	nowTime := time.Now()
	for index := range blacks {
		blacks[index].CreateTime = nowTime
	}
	err = dbConn.Table("blacks").Create(blacks).Error
	return err
}

func InsertUserMomentsBlackList(blacks []db.BlackForMoment) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	nowTime := time.Now()
	for index := range blacks {
		blacks[index].CreateTime = nowTime
	}
	err = dbConn.Table(db.BlackForMoment{}.TableName()).Create(blacks).Error
	return err
}

func GetBlackIDForMomentListByUserID(ownerUserID string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var blackIDList []string
	err = dbConn.Table(db.BlackForMoment{}.TableName()).Where("owner_user_id=?", ownerUserID).Pluck("block_user_id", &blackIDList).Error
	if err != nil {
		return nil, err
	}
	return blackIDList, nil
}

func RemoveBlackListUsers(ownerUserID string, blockUsersID []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table("blacks").Where("owner_user_id=?", ownerUserID).Where("block_user_id IN (?)", blockUsersID).Delete(&db.Black{}).Error
	return err
}

func RemoveMomentsBlackListUsers(ownerUserID string, blockUsersID []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(db.BlackForMoment{}.TableName()).Where("owner_user_id=?", ownerUserID).Where("block_user_id IN (?)", blockUsersID).Delete(&db.Black{}).Error
	return err
}

func AlterBlackRemark(OwnerID, BlackID, Remark, OpUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	black := db.Black{
		OwnerUserID: OwnerID,
		BlockUserID: BlackID,
		Remark:      Remark,
		UpdateUser:  OpUserID,
		UpdateTime:  time.Now().Unix(),
	}

	return dbConn.Table(black.TableName()).Where("owner_user_id=?", OwnerID).Where("block_user_id=?", BlackID).Updates(&black).Error
}

func GetBlacksByWhere(where map[string]interface{}, pageNumber, showNumber int32, orderBy string) ([]db.BlackRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var blackList []db.BlackRes
	var count int64
	if err != nil {
		return blackList, count, err
	}

	var ownerUsers []db.User
	var ownerUserIDList []string
	var ownerUsersMap map[string]*db.User

	var blackUsers []db.User
	var blackUserIDList []string
	var blackUsersMap map[string]*db.User

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["create_time"] = "create_time"

	// query owner user
	queryOwnerUser := false
	if ownerUserWhere, ok := where["owner_user"]; ok {
		ownerUser := ownerUserWhere.(string)
		if ownerUser != "" {
			queryOwnerUser = true
			ownerUsers = GetUserByAllCondition(ownerUser)
			ownerUsersMap = make(map[string]*db.User, len(ownerUsers))
			ownerUserIDList = make([]string, len(ownerUsers))
			for index, user := range ownerUsers {
				ownerUserIDList[index] = user.UserID
				ownerUsersMap[user.UserID] = &ownerUsers[index]
			}
		}
	}

	// query black friend list
	queryBlockUser := false
	if blockUserWhere, ok := where["block_user"]; ok {
		blockUser := blockUserWhere.(string)
		if blockUser != "" {
			queryBlockUser = true
			blackUsers = GetUserByAllCondition(blockUser)
			blackUsersMap = make(map[string]*db.User, len(blackUsers))
			blackUserIDList = make([]string, len(blackUsers))
			for index, user := range blackUsers {
				blackUserIDList[index] = user.UserID
				blackUsersMap[user.UserID] = &blackUsers[index]
			}
		}
	}

	// query black list
	dbSub := dbConn.Table(db.Black{}.TableName())
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
	if queryOwnerUser {
		dbSub = dbSub.Where("owner_user_id IN (?)", ownerUserIDList)
	}
	if queryBlockUser {
		dbSub = dbSub.Where("block_user_id IN (?)", blackUserIDList)
	}
	if remarkWhere, ok := where["remark"]; ok {
		remark := remarkWhere.(string)
		if remark != "" {
			dbSub = dbSub.Where("remark LIKE ?", "%"+remark+"%")
		}
	}
	if startTimeWhere, ok := where["start_time"]; ok {
		startTimeStr := startTimeWhere.(string)
		startTime, _ := strconv.ParseInt(startTimeStr, 10, 64)
		if startTimeStr != "" {
			dbSub = dbSub.Where("create_time >= ?", time.Unix(startTime, 0))
		}
	}
	if endTimeWhere, ok := where["end_time"]; ok {
		endTimeStr := endTimeWhere.(string)
		endTime, _ := strconv.ParseInt(endTimeStr, 10, 64)
		if endTimeStr != "" {
			dbSub = dbSub.Where("create_time <= ?", time.Unix(endTime, 0))
		}
	}
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}

	var blackDBList = make([]db.Black, 0)
	dbSub.Debug().Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&blackDBList)

	if !queryOwnerUser || !queryBlockUser {
		for _, black := range blackDBList {
			if !queryOwnerUser {
				ownerUserIDList = append(ownerUserIDList, black.OwnerUserID)
			}
			if !queryBlockUser {
				blackUserIDList = append(blackUserIDList, black.BlockUserID)
			}
		}
		if !queryOwnerUser {
			err = dbConn.Table("users").Where("user_id IN (?)", ownerUserIDList).Where("delete_time", 0).Find(&ownerUsers).Error
			if err != nil {
				return blackList, 0, err
			}

			ownerUsersMap = make(map[string]*db.User, len(ownerUsers))
			for index, user := range ownerUsers {
				ownerUsersMap[user.UserID] = &ownerUsers[index]
			}
		}
		if !queryBlockUser {
			err = dbConn.Table("users").Where("user_id IN (?)", blackUserIDList).Where("delete_time", 0).Find(&blackUsers).Error
			if err != nil {
				return blackList, 0, err
			}

			blackUsersMap = make(map[string]*db.User, len(blackUsers))
			for index, user := range blackUsers {
				blackUsersMap[user.UserID] = &blackUsers[index]
			}
		}
	}

	blackList = make([]db.BlackRes, len(blackDBList))
	for index, black := range blackDBList {
		blackRes := db.BlackRes{}
		_ = utils.CopyStructFields(&blackRes, black)
		if ownerUser, ok := ownerUsersMap[black.OwnerUserID]; ok {
			blackRes.OwnerUserName = ownerUser.Nickname
			blackRes.OwnerProfilePhoto = ownerUser.FaceURL
		}
		if blackUser, ok := blackUsersMap[black.BlockUserID]; ok {
			blackRes.BlockUserName = blackUser.Nickname
		}
		blackList[index] = blackRes
	}

	return blackList, count, nil
}
