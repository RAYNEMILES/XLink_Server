package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"gorm.io/gorm/clause"
)

func SaveGroupHeat(groupId, month string, msgCount, userCount, hate int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	data := db.GroupHeat{
		GroupId:   groupId,
		Month:     month,
		MsgCount:  msgCount,
		UserCount: userCount,
		Heat:      hate,
	}
	err = dbConn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "month"}},
		DoUpdates: clause.AssignmentColumns([]string{"msg_count", "user_count", "heat"}),
	}).Create(&data).Error
	if err != nil {
		log.NewError("", "SaveGroupHeat", "SaveGroupHeat", err.Error())
		return false
	}

	return true
}
