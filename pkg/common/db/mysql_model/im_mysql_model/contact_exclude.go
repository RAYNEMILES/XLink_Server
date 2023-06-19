package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
)

func AddContactExcludeByUserId(userId string, exclude []string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	exclude = utils.RemoveDuplicatesAndEmpty(exclude)

	if len(exclude) == 0 {
		return true
	}

	var data []db.ContactExclude
	for _, s := range exclude {
		data = append(data, db.ContactExclude{FromUserId: userId, ToUserId: s})
		dbConn.Delete(&db.ContactExclude{}, "from_user_id = ? and to_user_id = ?", userId, s)
	}
	err = dbConn.Table(db.ContactExclude{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("", "UpdateContactByUserId", "err", err, "data", data)
		return false
	}
	db.DB.DelUsersInfoByUserIdCache(userId)
	return true
}

func GetContactExcludeByUserId(userId string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var data []db.ContactExclude
	dbConn.Where("from_user_id = ?", userId).Find(&data)
	var result []string
	for _, v := range data {
		result = append(result, v.ToUserId)
	}
	return result
}
