package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"sync"
)

var rwMutex sync.RWMutex

func UpdateContactByUserId(userId string, phone []string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	rwMutex.Lock()
	defer rwMutex.Unlock()

	// 删除userid对应的所有联系人
	dbConn.Where("user_id = ?", userId).Delete(&db.Contact{})

	if len(phone) == 0 {
		return true
	}

	var data []db.Contact
	for _, s := range phone {
		data = append(data, db.Contact{UserId: userId, Phone: s})
	}
	err = dbConn.Table(db.Contact{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("", "UpdateContactByUserId", "err", err, "data", data)
		return false
	}
	db.DB.DelUsersInfoByUserIdCache(userId)
	return true
}

func GetContactByUserId(userId string) (data []db.Contact) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return
	}
	dbConn.Where("user_id = ?", userId).Find(&data)
	return
}
