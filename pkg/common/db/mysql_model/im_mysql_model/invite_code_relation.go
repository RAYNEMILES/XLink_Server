package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"time"
)

func AddFriendRalation(ownerUserId, friendUserId string, RalationType int) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	dbConn.Table(db.InviteCodeRelation{}.TableName()).Create(&db.InviteCodeRelation{
		OwnerUserId: ownerUserId,
		InviteUser:  friendUserId,
		Type:        RalationType,
		CreateTime:  time.Now().Unix(),
	})

	return true
}

func GetCountByOwnerUserId(ownerUserId string) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	var count int64
	dbConn.Table(db.InviteCodeRelation{}.TableName()).Where("owner_user_id = ?", ownerUserId).Count(&count)

	return count
}
