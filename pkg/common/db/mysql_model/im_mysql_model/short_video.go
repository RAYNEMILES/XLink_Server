package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

func InsertShortVideo(shortVideo *db.ShortVideo) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table(shortVideo.TableName()).Create(shortVideo).Error
}

func IncrShortVideoForwardNum(fileId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table(db.ShortVideo{}.TableName()).Where("file_id=?", fileId).Update("forward_num", gorm.Expr("forward_num + ?", 1)).Error
}

func GetShortVideoByFileId(fileId string) (*db.ShortVideo, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	shortVideo := &db.ShortVideo{}
	err = dbConn.Table(shortVideo.TableName()).Where("file_id = ?", fileId).First(shortVideo).Error
	return shortVideo, err
}

func GetShortVideoByFileIdList(fileId []string) ([]*db.ShortVideo, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var shortVideo []*db.ShortVideo
	if len(fileId) == 0 {
		return shortVideo, nil
	}

	fileIdStr := "'" + strings.Join(fileId, "','") + "'"
	err = dbConn.Raw("select * from short_video where file_id in (" + fileIdStr + ") order by field(file_id, " + fileIdStr + ")").Scan(&shortVideo).Error
	return shortVideo, err
}

func GetTopLikeShortVideoList() ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var shortVideoFiledIdList []string
	err = dbConn.Table(db.ShortVideo{}.TableName()).Where("status=?", constant.ShortVideoTypeNormal).Debug().Order("like_num desc").Limit(30).Pluck("file_id", &shortVideoFiledIdList).Error
	return shortVideoFiledIdList, err
}

func SearchShortVideoByKeyword(keyword []string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	shortVideo := &db.ShortVideo{}

	var shortVideoFiledIdList []string

	if len(keyword) == 0 {
		return shortVideoFiledIdList, nil
	}

	// Get possible users id
	var usersId []string
	usersIdSql := dbConn.Table(db.User{}.TableName()).Where("delete_time=0")
	usersIdSubSql := ""
	for _, v := range keyword {
		usersIdSubSql += fmt.Sprintf(" (user_id like '%%%s%%' or name like '%%%s%%' or phone_number like '%%%s%%' or email like '%%%s%%') OR", v, v, v, v)
	}
	usersIdSql.Where(usersIdSubSql[1:len(usersIdSubSql)-3]).Pluck("user_id", &usersId)

	// main sql
	mainSql := dbConn.Table(shortVideo.TableName()).Debug().Where("status=?", constant.ShortVideoTypeNormal).Order("like_num desc,create_time desc").Limit(30)
	mainSubSql := ""
	for i, v := range keyword {
		if i == 0 {
			if len(usersId) > 0 {
				usersIdStr := "'" + strings.Join(usersId, "','") + "'"
				mainSubSql += fmt.Sprintf("((user_id in (%s)) OR", usersIdStr)
			}
			mainSubSql += fmt.Sprintf(" (`name` like '%%%s%%' or `desc` like '%%%s%%') OR", v, v)
		} else {
			mainSubSql += fmt.Sprintf(" (`name` like '%%%s%%' or `desc` like '%%%s%%') OR", v, v)
		}
	}
	mainSql = mainSql.Where(mainSubSql[1 : len(mainSubSql)-3])
	mainSql.Pluck("file_id", &shortVideoFiledIdList)

	return shortVideoFiledIdList, nil
}

func UpdateShortVideoCommentReply(comment *db.ShortVideoComment) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	i := dbConn.Table(comment.TableName()).Debug().Updates(comment).
		RowsAffected
	if comment.Remark == "" {
		specifyMap := make(map[string]interface{})
		specifyMap["remark"] = gorm.Expr("remark", "")
		i += dbConn.Table(comment.TableName()).Debug().Updates(specifyMap).
			RowsAffected
	}

	return i, nil
}

func GetAllShortVideoByUserId(userId string) ([]*db.ShortVideo, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var shortVideo []*db.ShortVideo
	err = dbConn.Table(db.ShortVideo{}.TableName()).Where("user_id=?", userId).Order("create_time desc").Find(&shortVideo).Error

	return shortVideo, err
}

func GetShortVideoFileIdListByUserId(userId string, isSelf, isFriend bool, pageNumber, showNumber int32) ([]string, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	shortVideo := &db.ShortVideo{}
	status := []int32{constant.ShortVideoTypeNormal}
	if isSelf {
		status = append(status, constant.ShortVideoTypePrivate)
	}
	if isFriend {
		status = append(status, constant.ShortVideoTypeFriend)
	}

	var count int64
	err = dbConn.Table(shortVideo.TableName()).Where("user_id=?", userId).
		Where("status in (?)", status).Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoList", "GetShortVideoList", err.Error())
		return nil, 0, err
	}

	var shortVideoFiledIdList []string
	err = dbConn.Table(shortVideo.TableName()).Where("user_id=?", userId).Where("status in (?)", status).
		Order("create_time desc").Offset(int((pageNumber-1)*showNumber)).Limit(int(showNumber)).
		Pluck("file_id", &shortVideoFiledIdList).Error
	if err != nil {
		log.NewError("GetShortVideoList", "GetShortVideoList", err.Error())
		return nil, 0, err
	}

	return shortVideoFiledIdList, count, err
}

func GetFollowShortVideoFileIdListByUserId(userId string, pageNumber, showNumber int32) ([]string, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	status := []int32{constant.ShortVideoTypeNormal}

	var count = make(map[string]any, 1)
	var shortVideoFiledIdList []string

	err = dbConn.Debug().Raw("SELECT COUNT(*) as count FROM( ( SELECT s.file_id,s.user_id FROM short_video_follow f LEFT JOIN short_video s ON f.user_id = s.user_id WHERE f.fans_id = ? AND s.STATUS IN ? ) UNION ( SELECT s.file_id, s.user_id FROM short_video s WHERE s.user_id = ? AND s.STATUS IN ?)) AS temp", userId, status, userId, status).
		First(&count).Error

	if err != nil {
		log.NewError("GetFollowShortVideoList", "GetFollowShortVideoList", err.Error())
		return nil, 0, err
	}
	if count["count"].(int64) == 0 {
		return shortVideoFiledIdList, 0, nil
	}

	err = dbConn.Debug().Raw("SELECT file_id FROM( ( SELECT s.file_id,s.user_id,s.create_time FROM short_video_follow f LEFT JOIN short_video s ON f.user_id = s.user_id WHERE f.fans_id = ? AND s.STATUS IN ? ) UNION ( SELECT s.file_id, s.user_id,s.create_time FROM short_video s WHERE s.user_id = ? AND s.STATUS = ?)) AS temp GROUP BY temp.file_id ORDER BY temp.create_time desc LIMIT ?,?", userId, status, userId, status, (pageNumber-1)*showNumber, showNumber).Debug().Pluck("file_id", &shortVideoFiledIdList).Error

	return shortVideoFiledIdList, count["count"].(int64), err
}

func ManageGetShortVideoList(userIdList []string, status int8, desc string, EmptyDesc int64, isBlock int8, startTime, endTime int64, pageNumber, showNumber int32, interestList []int64) ([]*db.ShortVideo, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	shortVideo := &db.ShortVideo{}
	mainSql := dbConn.Table(shortVideo.TableName()).Debug()
	if len(userIdList) > 0 {
		mainSql = mainSql.Where("user_id in (?)", userIdList)
	}
	mainSql = mainSql.Where("`status`!=?", constant.ShortVideoTypeDeleted)
	if status != 0 {
		mainSql = mainSql.Where("`status`=?", status)
	}
	if EmptyDesc == 1 {
		mainSql = mainSql.Where("`desc`=?", "")
	} else {
		if desc != "" {
			mainSql = mainSql.Where("`desc` like ?", "%"+desc+"%")
		}
	}
	if startTime != 0 {
		mainSql = mainSql.Where("create_time>=?", startTime)
	}
	if endTime != 0 {
		mainSql = mainSql.Where("create_time<=?", endTime)
	}
	if len(interestList) > 0 {
		for _, v := range interestList {
			mainSql = mainSql.Where("interest_id like ?", "%"+strconv.FormatInt(v, 10)+"%")
		}
	}

	var count int64
	var shortVideoList []*db.ShortVideo
	err = mainSql.Count(&count).Debug().Error
	if err != nil {
		log.NewError("ManageGetShortVideoList", "ManageGetShortVideoList", err.Error())
		return shortVideoList, 0, err
	}
	if count == 0 {
		return shortVideoList, 0, nil
	}

	err = mainSql.Order("create_time desc").Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&shortVideoList).Error
	if err != nil {
		log.NewError("ManageGetShortVideoList", "ManageGetShortVideoList", err.Error())
		return shortVideoList, 0, err
	}

	return shortVideoList, count, nil
}

func MultiDeleteShortVideoByFileId(fileIdList []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	shortVideo := &db.ShortVideo{}
	err = dbConn.Table(shortVideo.TableName()).Where("file_id in (?)", fileIdList).Update("status", constant.ShortVideoTypeDeleted).Error
	if err != nil {
		log.NewError("MultiDeleteShortVideoByFileId", "MultiDeleteShortVideoByFileId", err.Error())
		return err
	}

	return nil
}

func DeleteShortVideoByFileIdList(fileIdList []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	shortVideo := &db.ShortVideo{}
	err = dbConn.Table(shortVideo.TableName()).Where("file_id in (?)", fileIdList).Updates(map[string]interface{}{
		"status":           constant.ShortVideoTypeDeleted,
		"like_num":         0,
		"comment_num":      0,
		"reply_num":        0,
		"forward_num":      0,
		"comment_like_num": 0,
	}).Error
	if err != nil {
		log.NewError("DeleteShortVideoByFileIdList", "DeleteShortVideoByFileIdList", err.Error())
		return err
	}

	return nil
}

func UpdateShortVideoInfoByFileId(fileId string, shortVideo *db.ShortVideo) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	updateMap := make(map[string]interface{})
	updateMap["desc"] = shortVideo.Desc
	updateMap["status"] = shortVideo.Status
	updateMap["remark"] = shortVideo.Remark
	err = dbConn.Table(shortVideo.TableName()).Where("file_id=?", fileId).Updates(updateMap).Error
	if err != nil {
		log.NewError("UpdateShortVideoInfoByFileId", "UpdateShortVideoInfoByFileId", err.Error())
		return err
	}

	return nil
}

func UpdateShortVideoInterestByFileId(fileId string, interestId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	updateMap := make(map[string]interface{})
	updateMap["interest_id"] = interestId
	err = dbConn.Table(db.ShortVideo{}.TableName()).Where("file_id=?", fileId).Updates(updateMap).Error
	if err != nil {
		log.NewError("UpdateShortVideoInterestByFileId", "UpdateShortVideoInterestByFileId", err.Error())
		return err
	}

	return nil
}

func UpdateShortVideoMediaUrlByFileId(fileId string, mediaUrl string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	if fileId == "" || mediaUrl == "" {
		return errors.New("fileId or mediaUrl is empty")
	}

	updateMap := make(map[string]interface{})
	updateMap["media_url"] = mediaUrl
	err = dbConn.Table(db.ShortVideo{}.TableName()).Where("file_id=?", fileId).Updates(updateMap).Error
	if err != nil {
		log.NewError("UpdateShortVideoMediaUrlByFileId", "UpdateShortVideoMediaUrlByFileId", err.Error())
		return err
	}

	return nil
}

func AlterShortVideo(info *db.ShortVideo) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(info.TableName()).Where("file_id=?", info.FileId).Updates(info).Error
	if err != nil {
		log.NewError("AlterShortVideo", "AlterShortVideo", err.Error())
		return err
	}

	return nil
}
func IncrShortVideoCountByFileId(fileId, field string, count int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	var model db.ShortVideo
	switch field {
	case "like_num", "comment_num", "comment_like_num", "reply_num", "forward_num":
		err = dbConn.Table(model.TableName()).Where("file_id=?", fileId).Update(field, gorm.Expr(field+" + ?", count)).Error
		break
	default:
		return false
	}

	if err != nil {
		log.NewError("IncrShortVideoCountByFileId err", err)
		return false
	}
	return true
}
