package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
)

func AddInterestGroupExcludeListByUserId(userId string, groupId []string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	if len(groupId) == 0 {
		return true
	}

	var data []db.InterestGroupExclude
	for _, s := range groupId {
		data = append(data, db.InterestGroupExclude{GroupId: s, UserId: userId})
	}
	err = dbConn.Table(db.InterestGroupExclude{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("", "UpdateInterestGroup", "err", err, "data", data)
		return false
	}
	db.DB.DeleteInterestGroupINfoListByUserId(userId)

	return true
}

func GetInterestGroupExcludeListByUserId(userId string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var data []db.InterestGroupExclude
	dbConn.Table(db.InterestGroupExclude{}.TableName()).Where("user_id = ?", userId).Find(&data)
	var result []string
	for _, v := range data {
		result = append(result, v.GroupId)
	}
	return result
}
