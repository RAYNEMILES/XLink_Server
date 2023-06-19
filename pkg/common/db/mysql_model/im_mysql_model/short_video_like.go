package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"gorm.io/gorm"
)

func InsertShortVideoLike(fileId, userId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		insertErr := tx.Table(db.ShortVideoLike{}.TableName()).Debug().Create(&db.ShortVideoLike{
			FileId: fileId,
			UserId: userId,
		}).Error
		if insertErr != nil {
			return insertErr
		}

		updateErr := tx.Table(db.ShortVideo{}.TableName()).Debug().Where("file_id = ?", fileId).Update("like_num", gorm.Expr("like_num + ?", 1)).Error
		if updateErr != nil {
			return updateErr
		}

		return nil
	})
	if transactionErr != nil {
		log.NewError("InsertShortVideoLike", "InsertShortVideoLike", transactionErr.Error())
		return false
	}
	return true
}

func DeleteShortVideoLike(fileId, userId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		deleteErr := tx.Table(db.ShortVideoLike{}.TableName()).Debug().Where("file_id = ? and user_id = ?", fileId, userId).Delete(&db.ShortVideoLike{}).Error
		if deleteErr != nil {
			return deleteErr
		}

		updateErr := tx.Table(db.ShortVideo{}.TableName()).Debug().Where("file_id = ?", fileId).Update("like_num", gorm.Expr("like_num - ?", 1)).Error
		if updateErr != nil {
			return updateErr
		}

		return nil
	})
	if transactionErr != nil {
		log.NewError("DeleteShortVideoLike", "DeleteShortVideoLike", transactionErr.Error())
		return false
	}
	return true
}

func GetLikeCountByFileId(fileId []string) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	shortVideoLike := &db.ShortVideoLike{}
	var count int64
	err = dbConn.Table(shortVideoLike.TableName()).Debug().Where("file_id in (?)", fileId).Count(&count).Error
	if err != nil {
		log.NewError("GetLikeCountByFileId", "GetLikeCountByFileId", err.Error())
		return 0
	}
	return count
}

func GetLikeListByFileId(fileId []string, pageNumber, showNumber int64) ([]*db.ShortVideoLike, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	shortVideoLike := &db.ShortVideoLike{}
	var shortVideoLikeList []*db.ShortVideoLike
	err = dbConn.Table(shortVideoLike.TableName()).Where("file_id in (?)", fileId).Offset(int((pageNumber - 1) * showNumber)).
		Order("id ASC").Limit(int(showNumber)).Find(&shortVideoLikeList).Error
	if err != nil {
		log.NewError("GetLikeListByFileId", "GetLikeListByFileId", err.Error())
		return nil, err
	}

	return shortVideoLikeList, err
}

func MultiDeleteLikeByLikeIdList(idList []int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		deleteErr := tx.Table(db.ShortVideoLike{}.TableName()).Debug().Where("id in (?)", idList).Delete(&db.ShortVideoLike{}).Error
		if deleteErr != nil {
			return deleteErr
		}

		return nil
	})
	if transactionErr != nil {
		log.NewError("MultiDeleteLikeByLikeIdList", "MultiDeleteLikeByLikeIdList", transactionErr.Error())
		return false
	}
	return true
}

func GetShortVideoLikeCountByUserId(userId string) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	shortVideoLike := &db.ShortVideoLike{}
	var count int64
	err = dbConn.Table(shortVideoLike.TableName()).Where("user_id = ?", userId).Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoLikeCountByUserId", "GetShortVideoLikeCountByUserId", err.Error())
		return 0
	}
	return count
}

func GetLikeListByUserId(userId string, pageNumber, showNumber int64) ([]*db.ShortVideoLike, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	shortVideoLike := &db.ShortVideoLike{}
	var shortVideoLikeList []*db.ShortVideoLike
	err = dbConn.Table(shortVideoLike.TableName()).Where("user_id=?", userId).Offset(int((pageNumber - 1) * showNumber)).
		Order("id ASC").Limit(int(showNumber)).Find(&shortVideoLikeList).Error
	if err != nil {
		log.NewError("GetLikeListByUserId", "GetLikeListByUserId", err.Error())
		return nil, err
	}

	return shortVideoLikeList, err
}

func GetLikeShortVideoFileIdList(userId string, isSelf, isFriend bool, pageNumber, showNumber int32) ([]string, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	shortVideoLike := &db.ShortVideoLike{}
	shortVideo := &db.ShortVideo{}
	condition := map[string]interface{}{}
	if userId != "" {
		condition[shortVideoLike.TableName()+"."+"user_id"] = userId
	}
	orderBy := shortVideoLike.TableName() + ".create_time desc"

	status := []int32{constant.ShortVideoTypeNormal}
	if isSelf {
		status = append(status, constant.ShortVideoTypePrivate)
	}
	if isFriend {
		status = append(status, constant.ShortVideoTypeFriend)
	}

	var count int64
	err = dbConn.Table(shortVideoLike.TableName()).Where(condition).
		Joins("left join "+shortVideo.TableName()+" s on s.file_id = "+shortVideoLike.TableName()+".file_id").
		Where("s.status in (?)", status).Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoLikeList", "GetShortVideoLikeList", err.Error())
		return nil, 0, err
	}

	var shortVideoFiledIdList []string
	err = dbConn.Table(shortVideoLike.TableName()).Where(condition).Debug().
		Joins("left join "+shortVideo.TableName()+" s on s.file_id = "+shortVideoLike.TableName()+".file_id").
		Where("s.status in (?)", status).Order(orderBy).Offset(int((pageNumber-1)*showNumber)).
		Limit(int(showNumber)).Pluck(shortVideoLike.TableName()+".file_id", &shortVideoFiledIdList).Error
	if err != nil {
		log.NewError("GetShortVideoLikeList", "GetShortVideoLikeList", err.Error())
		return nil, 0, err
	}

	return shortVideoFiledIdList, count, err
}

func GetShortVideoLikeList(userIdList, likeUserIdList []string, fileId string, status int32, desc string, emptyDesc int32, startTime, endTime int64, pageNumber, showNumber int32) ([]*db.ShortVideoLike, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	shortVideoLike := &db.ShortVideoLike{}
	shortVideo := &db.ShortVideo{}
	condition := map[string]interface{}{}
	likeCondition := map[string]interface{}{}
	if len(userIdList) > 0 {
		condition[shortVideo.TableName()+"."+"user_id"] = userIdList
	}
	if len(likeUserIdList) > 0 {
		condition[shortVideoLike.TableName()+"."+"user_id"] = likeUserIdList
	}
	if fileId != "" {
		condition[shortVideoLike.TableName()+"."+"file_id"] = fileId
	}
	if status != 0 {
		condition[shortVideo.TableName()+"."+"status"] = status
	}
	if emptyDesc == 1 {
		condition[shortVideo.TableName()+"."+"desc"] = ""
	} else {
		if desc != "" {
			likeCondition[shortVideo.TableName()+"."+"desc"] = "%" + desc + "%"
		}
	}
	orderBy := shortVideoLike.TableName() + ".create_time desc"

	var count int64
	sub := dbConn.Table(shortVideoLike.TableName()).Where(condition).Debug()
	if startTime > 0 {
		sub = sub.Where(shortVideoLike.TableName()+".create_time >= ?", startTime)
	}
	if endTime > 0 {
		sub = sub.Where(shortVideoLike.TableName()+".create_time <= ?", endTime)
	}
	if len(likeCondition) > 0 {
		for k, v := range likeCondition {
			sub = sub.Where(k+" like ?", v)
		}
	}
	err = sub.Joins("left join " + shortVideo.TableName() + " " + shortVideo.TableName() + " on " + shortVideo.TableName() + ".file_id = " + shortVideoLike.TableName() + ".file_id").
		Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoLikeList", "GetShortVideoLikeList", err.Error())
		return nil, 0, err
	}

	var shortVideoLikeList []*db.ShortVideoLike
	mainSub := dbConn.Table(shortVideoLike.TableName()).Where(condition)
	if len(likeCondition) > 0 {
		for k, v := range likeCondition {
			mainSub = mainSub.Where(k+" like ?", v)
		}
	}
	if startTime > 0 {
		mainSub = mainSub.Where(shortVideoLike.TableName()+".create_time >= ?", startTime)
	}
	if endTime > 0 {
		mainSub = mainSub.Where(shortVideoLike.TableName()+".create_time <= ?", endTime)
	}
	err = mainSub.Joins("left join " + shortVideo.TableName() + " " + shortVideo.TableName() + " on " + shortVideo.TableName() + ".file_id = " + shortVideoLike.TableName() + ".file_id").
		Order(orderBy).Offset(int((pageNumber - 1) * showNumber)).
		Limit(int(showNumber)).Find(&shortVideoLikeList).Error
	if err != nil {
		log.NewError("GetShortVideoLikeList", "GetShortVideoLikeList", err.Error())
		return nil, 0, err
	}

	return shortVideoLikeList, count, err
}

func GetShortVideoLikeByLikeIdList(idList []int64) ([]*db.ShortVideoLike, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var shortVideoLikeList []*db.ShortVideoLike

	shortVideoLike := &db.ShortVideoLike{}
	if len(idList) == 0 {
		return shortVideoLikeList, nil
	}

	err = dbConn.Table(shortVideoLike.TableName()).Where("id in (?)", idList).Find(&shortVideoLikeList).Error
	if err != nil {
		log.NewError("GetShortVideoLikeByLikeIdList", "GetShortVideoLikeByLikeIdList", err.Error())
		return nil, err
	}

	return shortVideoLikeList, err
}
