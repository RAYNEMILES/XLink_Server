package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"gorm.io/gorm"
)

func InsertShortVideoCommentLike(fileId, userId string, commentId int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		insertErr := tx.Table(db.ShortVideoCommentLike{}.TableName()).Create(&db.ShortVideoCommentLike{
			UserId:    userId,
			CommentId: commentId,
			FileId:    fileId,
		}).Error
		if insertErr != nil {
			return insertErr
		}

		updateErr := tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", commentId).Update("comment_like_count", gorm.Expr("comment_like_count + ?", 1)).Error
		if updateErr != nil {
			return updateErr
		}

		updateErr = tx.Table(db.ShortVideo{}.TableName()).Where("file_id = ?", fileId).Update("comment_like_num", gorm.Expr("comment_like_num + ?", 1)).Error
		if updateErr != nil {
			return updateErr
		}

		return nil
	})
	if transactionErr != nil {
		log.NewError("InsertShortVideoCommentLike", "InsertShortVideoCommentLike", transactionErr.Error())
		return false
	}
	return true
}

func DeleteShortVideoCommentLike(fileId, userId string, commentId int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		deleteErr := tx.Table(db.ShortVideoCommentLike{}.TableName()).Where("user_id = ? and comment_id = ?", userId, commentId).Delete(&db.ShortVideoCommentLike{}).Error
		if deleteErr != nil {
			return deleteErr
		}

		updateErr := tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", commentId).Update("comment_like_count", gorm.Expr("comment_like_count - ?", 1)).Error
		if updateErr != nil {
			return updateErr
		}

		updateErr = tx.Table(db.ShortVideo{}.TableName()).Where("file_id = ?", fileId).Update("comment_like_num", gorm.Expr("comment_like_num - ?", 1)).Error
		if updateErr != nil {
			return updateErr
		}
		return nil
	})
	if transactionErr != nil {
		log.NewError("DeleteShortVideoCommentLike", "DeleteShortVideoCommentLike", transactionErr.Error())
		return false
	}
	return true
}

func UpdateShortVideoCommentLike(like *db.ShortVideoCommentLike) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(like.TableName()).Updates(like).Error
	if err != nil {
		return err
	}
	if like.Remark == "" {
		specifyMap := make(map[string]interface{})
		specifyMap["remark"] = gorm.Expr("remark", "")
		err = dbConn.Table(like.TableName()).Debug().Updates(specifyMap).
			Error
		if err != nil {
			return err
		}
	}
	return nil
}

func GetCommentLikeCountByFileId(fileId []string) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	var count int64
	err = dbConn.Table(db.ShortVideoCommentLike{}.TableName()).Where("file_id in (?)", fileId).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

func GetCommentLikeListByFileId(fileId []string, pageNumber, showNumber int64) ([]db.ShortVideoCommentLike, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var likeList []db.ShortVideoCommentLike
	err = dbConn.Table(db.ShortVideoCommentLike{}.TableName()).Where("file_id in (?)", fileId).Order("id asc").
		Offset(int((pageNumber - 1) * showNumber)).Limit(int(showNumber)).Find(&likeList).Error
	if err != nil {
		log.NewError("GetCommentLikeListByFileId", "GetCommentLikeListByFileId", err.Error())
	}

	return likeList, err
}

func MultiDeleteCommentLikeByLikeIdList(likeIdList []int64) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(db.ShortVideoCommentLike{}.TableName()).Where("id in (?)", likeIdList).Delete(&db.ShortVideoCommentLike{}).Error
	if err != nil {
		return err
	}
	return nil
}

func GetShortVideoCommentLike(LikeId int64) (db.ShortVideoCommentLike, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return db.ShortVideoCommentLike{}, nil
	}
	like := db.ShortVideoCommentLike{}
	err = dbConn.Table(like.TableName()).Where("id=?", LikeId).First(&like).Error

	return like, err
}
