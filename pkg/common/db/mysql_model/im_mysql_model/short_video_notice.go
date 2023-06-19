package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
)

func AddShortVideoNotice(data db.ShortVideoNotice) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	err = dbConn.Table(db.ShortVideoNotice{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("initShortVideoUserCount err", err)
		return 0
	}

	return data.Id
}

func GetShortVideoNoticeList(userId string, noticeType int8, state int8, pageNumber, showNumber int32) ([]db.ShortVideoNotice, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	condition := map[string]interface{}{}
	if userId != "" {
		condition["user_id"] = userId
	}
	if noticeType != 0 {
		condition["type"] = noticeType
	}
	if state != 0 {
		condition["state"] = state
	}

	var list []db.ShortVideoNotice
	var count int64

	err = dbConn.Table(db.ShortVideoNotice{}.TableName()).Where(condition).Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoNoticeList err", err)
		return nil, 0, err
	}
	if count == 0 {
		return list, count, nil
	}

	err = dbConn.Table(db.ShortVideoNotice{}.TableName()).Where(condition).Order("id DESC").
		Offset(int((pageNumber - 1) * showNumber)).Limit(int(showNumber)).Find(&list).Error
	if err != nil {
		log.NewError("GetShortVideoNoticeList err", err)
		return nil, 0, err
	}

	return list, count, nil
}

func UpdateShortVideoStateByIdList(idList []int64, state int8) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err = dbConn.Table(db.ShortVideoNotice{}.TableName()).Where("id in (?)", idList).Update("state", state).Error
	if err != nil {
		log.NewError("UpdateShortVideoStateByIdList err", err)
		return false
	}

	return true
}
