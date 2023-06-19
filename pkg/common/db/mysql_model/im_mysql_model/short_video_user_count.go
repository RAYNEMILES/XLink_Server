package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"gorm.io/gorm"
)

func initShortVideoUserCount(userId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	data := &db.ShortVideoUserCount{
		UserId: userId,
	}
	err = dbConn.Table(db.ShortVideoUserCount{}.TableName()).Create(data).Error
	if err != nil {
		log.NewError("initShortVideoUserCount err", err)
		return false
	}

	return true
}

func GetShortVideoUserCountByUserId(userId string) (*db.ShortVideoUserCount, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	data := &db.ShortVideoUserCount{}
	err = dbConn.Table(db.ShortVideoUserCount{}.TableName()).Where("user_id = ?", userId).First(data).Error
	if err != nil {
		initShortVideoUserCount(userId)
		data.UserId = userId
		return data, nil
	}

	return data, nil
}

func UpdateShortVideoUserCountByUserId(userId, field string, count int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	var model db.ShortVideoUserCount
	dbConn.FirstOrCreate(&model, db.ShortVideoUserCount{UserId: userId})

	switch field {
	case "work_num", "like_num", "comment_num", "comment_like_num", "fans_num", "follow_num":
		err = dbConn.Model(&model).Update(field, count).Error
		break
	default:
		return false
	}

	if err != nil {
		log.NewError("UpdateShortVideoUserCountByUserId err", err)
		return false
	}
	return true
}

func IncrShortVideoUserCountByUserId(userId, field string, count int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	var model db.ShortVideoUserCount
	dbConn.FirstOrCreate(&model, db.ShortVideoUserCount{UserId: userId})

	switch field {
	case "work_num", "like_num", "harvested_likes_number", "comment_num", "comment_like_num", "fans_num", "follow_num":
		err = dbConn.Table(model.TableName()).Where("user_id=?", userId).Update(field, gorm.Expr(field+" + ?", count)).Error
		break
	default:
		return false
	}

	if err != nil {
		log.NewError("UpdateShortVideoUserCountByUserId err", err)
		return false
	}
	return true
}
