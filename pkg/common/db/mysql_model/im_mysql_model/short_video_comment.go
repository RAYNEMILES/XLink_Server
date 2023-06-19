package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"fmt"
	"gorm.io/gorm"
)

func InsertShortVideoComment(shortVideoComment *db.ShortVideoComment) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		insertErr := tx.Table(db.ShortVideoComment{}.TableName()).Create(shortVideoComment).Error
		if insertErr != nil {
			return insertErr
		}

		if shortVideoComment.ParentId != 0 {
			updateErr := tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", shortVideoComment.ParentId).Update("comment_reply_count", gorm.Expr("comment_reply_count + ?", 1)).Error
			if updateErr != nil {
				return updateErr
			}

			updateErr = tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", shortVideoComment.Level0CommentId).Update("total_reply_count", gorm.Expr("total_reply_count + ?", 1)).Error
			if updateErr != nil {
				return updateErr
			}
		}

		// update short video comment num
		updateShortVideoErr := tx.Table(db.ShortVideo{}.TableName()).Where("file_id = ?", shortVideoComment.FileId).Update("comment_num", gorm.Expr("comment_num + ?", 1)).Error
		if updateShortVideoErr != nil {
			return updateShortVideoErr
		}

		// update short video reply num
		if shortVideoComment.ParentId != 0 {
			updateShortVideoErr = tx.Table(db.ShortVideo{}.TableName()).Where("file_id = ?", shortVideoComment.FileId).Update("reply_num", gorm.Expr("reply_num + ?", 1)).Error
			if updateShortVideoErr != nil {
				return updateShortVideoErr
			}
		}

		return nil
	})
	if transactionErr != nil {
		log.NewError("InsertShortVideoComment", "InsertShortVideoComment", transactionErr.Error())
		return 0, transactionErr
	}
	return shortVideoComment.CommentId, nil

}

func UpdateShortVideoCommentStatus(commentId int64, status int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	shortVideoComment := &db.ShortVideoComment{}
	err = dbConn.Table(shortVideoComment.TableName()).Where("comment_id = ?", commentId).Update("status", status).Error
	if err != nil {
		log.NewError("UpdateShortVideoCommentStatus", "UpdateShortVideoCommentStatus", err.Error())
	}

	return err
}

func GetCommentCountByFileId(fileId []string) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	shortVideoComment := &db.ShortVideoComment{}
	var count int64
	err = dbConn.Table(shortVideoComment.TableName()).Where("file_id in (?)", fileId).Count(&count).Error
	if err != nil {
		log.NewError("GetCommentCountByFileId", "GetCommentCountByFileId", err.Error())
	}

	return count
}

func GetCommentListByFileId(fileId []string, pageNumber, showNumber int64) ([]*db.ShortVideoComment, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	shortVideoComment := &db.ShortVideoComment{}
	var commentList []*db.ShortVideoComment
	err = dbConn.Table(shortVideoComment.TableName()).Where("file_id in (?)", fileId).Order("comment_id asc").
		Offset(int((pageNumber - 1) * showNumber)).Limit(int(showNumber)).Find(&commentList).Error
	if err != nil {
		log.NewError("GetCommentListByFileId", "GetCommentListByFileId", err.Error())
	}

	return commentList, err
}

func MultiDeleteCommentByCommentIdList(commentIdList []int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		deleteErr := tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id in (?)", commentIdList).Update("status", constant.ShortVideoCommentStatusDeleted).Error
		if deleteErr != nil {
			return deleteErr
		}
		return nil
	})
	if transactionErr != nil {
		log.NewError("MultiDeleteCommentByCommentIdList", "MultiDeleteCommentByCommentIdList", transactionErr.Error())
		return false
	}
	return true
}

func DeleteShortVideoCommentByCommentList(commentList []*db.ShortVideoComment) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	transactionErr := dbConn.Transaction(func(tx *gorm.DB) error {
		for _, comment := range commentList {
			deleteErr := tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", comment.CommentId).Update("status", constant.ShortVideoCommentStatusDeleted).Error
			if deleteErr != nil {
				return deleteErr
			}

			if comment.ParentId != 0 {
				updateErr := tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", comment.ParentId).Update("comment_reply_count", gorm.Expr("comment_reply_count - ?", 1)).Error
				if updateErr != nil {
					return updateErr
				}

				updateErr = tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id = ?", comment.Level0CommentId).Update("total_reply_count", gorm.Expr("total_reply_count - ?", 1)).Error
				if updateErr != nil {
					return updateErr
				}
			}

			// update short video comment num
			updateShortVideoErr := tx.Table(db.ShortVideo{}.TableName()).Where("file_id = ?", comment.FileId).Update("comment_num", gorm.Expr("comment_num - ?", 1)).Error
			if updateShortVideoErr != nil {
				return updateShortVideoErr
			}

			// update short video reply num
			if comment.ParentId != 0 {
				updateShortVideoErr = tx.Table(db.ShortVideo{}.TableName()).Where("file_id = ?", comment.FileId).Update("reply_num", gorm.Expr("reply_num - ?", 1)).Error
				if updateShortVideoErr != nil {
					return updateShortVideoErr
				}
			}
		}

		return nil
	})
	if transactionErr != nil {
		log.NewError("DeleteShortVideoComment", "DeleteShortVideoComment", transactionErr.Error())
		return transactionErr
	}
	return nil
}

func UpdateShortVideoCommentByCommentId(commentId int64, content, remark string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	shortVideoComment := &db.ShortVideoComment{}
	err = dbConn.Table(shortVideoComment.TableName()).Where("comment_id = ?", commentId).Updates(map[string]interface{}{"content": content, "remark": remark}).Error
	if err != nil {
		log.NewError("UpdateShortVideoCommentByCommentId", "UpdateShortVideoCommentByCommentId", err.Error())
	}

	return err
}

func GetShortVideoCommentByCommentId(commentId int64) (*db.ShortVideoComment, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	shortVideoComment := &db.ShortVideoComment{}
	err = dbConn.Table(shortVideoComment.TableName()).Where("comment_id = ?", commentId).First(shortVideoComment).Error
	if err != nil {
		log.NewError("GetShortVideoCommentByCommentId", "GetShortVideoCommentByCommentId", err.Error())
	}

	return shortVideoComment, err
}

// GetAllReplyCommentIdByCommentId
// Does not contain it self
func GetAllReplyCommentIdByCommentId(commentId int64) ([]int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	shortVideoComment := &db.ShortVideoComment{}
	var commentIds []int64
	err = dbConn.Table(shortVideoComment.TableName()).Where("parent_id = ?", commentId).Pluck("comment_id", &commentIds).Error
	if err != nil {
		log.NewError("GetAllReplyCommentIdByCommentId", "GetAllReplyCommentIdByCommentId", err.Error())
		return nil, err
	}

	for _, commentId := range commentIds {
		childCommentIds, err := GetAllReplyCommentIdByCommentId(commentId)
		if err != nil {
			return nil, err
		}
		commentIds = append(commentIds, childCommentIds...)
	}

	return commentIds, nil
}

func GetCommentInfoListByCommentId(commentIds []int64) ([]*db.ShortVideoComment, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	shortVideoComment := &db.ShortVideoComment{}
	var commentList []*db.ShortVideoComment
	err = dbConn.Table(shortVideoComment.TableName()).Where("comment_id in (?)", commentIds).Find(&commentList).Error
	if err != nil {
		log.NewError("GetCommentInfoListByCommentId", "GetCommentInfoListByCommentId", err.Error())
		return nil, err
	}

	return commentList, nil
}

func GetShortVideoCommentList(status int64, fileId, userId string, commentId, parentId, levelId, level1CommentId int64, pageNumber, showNumber, order int32, sourceCommentId ...int64) ([]*db.ShortVideoComment, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	condition := map[string]interface{}{}
	if levelId != -1 {
		condition["level_id"] = levelId
	}
	if level1CommentId != 0 {
		condition["level1_comment_id"] = level1CommentId
	}
	if fileId != "" {
		condition["file_id"] = fileId
	}
	if userId != "" {
		condition["user_id"] = userId
	}
	if parentId != -1 {
		condition["parent_id"] = parentId
	}
	if commentId != 0 {
		condition["comment_id"] = commentId
	}
	if status != 0 {
		condition["status"] = status
	}

	sourceOrder := ""
	if len(sourceCommentId) > 0 {
		if sourceCommentId[0] > 0 {
			sourceOrder = fmt.Sprintf("comment_id = %d desc,", sourceCommentId[0])
		}
	}
	orderBy := sourceOrder + "create_time desc"
	if order == 1 {
		orderBy = sourceOrder + "create_time asc"
	} else if order == 2 {
		orderBy = sourceOrder + "comment_like_count desc"
	}

	shortVideoComment := &db.ShortVideoComment{}
	var count int64
	err = dbConn.Table(shortVideoComment.TableName()).Where(condition).Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoCommentList", "GetShortVideoCommentList", err.Error())
		return nil, 0, err
	}

	var shortVideoCommentList []*db.ShortVideoComment
	err = dbConn.Table(shortVideoComment.TableName()).Debug().Where(condition).Order(orderBy).Offset(int((pageNumber - 1) * showNumber)).Limit(int(showNumber)).Find(&shortVideoCommentList).Error
	if err != nil {
		log.NewError("GetShortVideoCommentList", "GetShortVideoCommentList", err.Error())
		return nil, 0, err
	}

	return shortVideoCommentList, count, err
}

func ManagementGetShortVideoCommentList(parentId int64, userIdList, commentUserId []string, fileId string, status int32, desc string, content string, emptyDesc int64, startTime, endTime int64, pageNumber, showNumber int32) ([]*db.ShortVideoCommentResult, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	dbSubQuery := dbConn.Table(db.ShortVideoComment{}.TableName()).Debug().
		Joins("left join " + db.ShortVideo{}.TableName() + " as short_video on short_video_comment.file_id = short_video.file_id").
		Select("short_video_comment.comment_id, short_video_comment.file_id, short_video_comment.user_id,short_video_comment.create_time, short_video_comment.content,short_video_comment.comment_reply_count,short_video_comment.comment_like_count,short_video_comment.remark,short_video.cover_url,short_video.media_url,short_video.user_id as post_user_id,short_video.desc,short_video.status")
	dbSubQuery = dbSubQuery.Where("short_video_comment.status != ?", constant.ShortVideoCommentStatusDeleted)
	if parentId != -1 {
		dbSubQuery = dbSubQuery.Where("short_video_comment.parent_id = ?", parentId)
	}
	if len(userIdList) > 0 {
		dbSubQuery = dbSubQuery.Where("short_video.user_id in (?)", userIdList)
	}
	if len(commentUserId) > 0 {
		dbSubQuery = dbSubQuery.Where("short_video_comment.user_id in (?)", commentUserId)
	}
	if fileId != "" {
		dbSubQuery = dbSubQuery.Where("short_video_comment.file_id = ?", fileId)
	}
	if status != 0 {
		dbSubQuery = dbSubQuery.Where("short_video.status = ?", status)
	}
	if emptyDesc == 1 {
		dbSubQuery = dbSubQuery.Where("short_video.desc = ''")
	} else {
		if desc != "" {
			dbSubQuery = dbSubQuery.Where("short_video.desc like ?", "%"+desc+"%")
		}
	}
	if content != "" {
		dbSubQuery = dbSubQuery.Where("short_video_comment.content like ?", "%"+content+"%")
	}
	if startTime != 0 {
		dbSubQuery = dbSubQuery.Where("short_video_comment.create_time >= ?", startTime)
	}
	if endTime != 0 {
		dbSubQuery = dbSubQuery.Where("short_video_comment.create_time <= ?", endTime)
	}

	var count int64
	err = dbSubQuery.Count(&count).Error
	if err != nil {
		log.NewError("ManagementGetShortVideoCommentList", "ManagementGetShortVideoCommentList", err.Error())
		return nil, 0, err
	}

	var shortVideoCommentList []*db.ShortVideoCommentResult
	err = dbSubQuery.Order("short_video_comment.create_time desc").Offset(int((pageNumber - 1) * showNumber)).Limit(int(showNumber)).Scan(&shortVideoCommentList).Error
	if err != nil {
		log.NewError("ManagementGetShortVideoCommentList", "ManagementGetShortVideoCommentList", err.Error())
		return nil, 0, err
	}

	return shortVideoCommentList, count, err
}

func ManagementGetShortVideoCommentCountByFileIdList(fileIdList []string) (map[string]int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	shortVideoComment := &db.ShortVideoComment{}
	var shortVideoCommentList []*db.ShortVideoComment
	err = dbConn.Table(shortVideoComment.TableName()).Where("file_id in (?)", fileIdList).Debug().
		Where("status!=3 and parent_id=0").Find(&shortVideoCommentList).Error
	if err != nil {
		log.NewError("ManagementGetShortVideoCommentCountByFileIdList", "ManagementGetShortVideoCommentCountByFileIdList", err.Error())
		return nil, err
	}

	commentCountMap := make(map[string]int64)
	for _, shortVideoComment := range shortVideoCommentList {
		commentCountMap[shortVideoComment.FileId]++
	}

	return commentCountMap, nil
}

func GetShortVideoCommentByCommentIdList(commentIdList []int64) ([]*db.ShortVideoComment, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	shortVideoComment := &db.ShortVideoComment{}
	var shortVideoCommentList []*db.ShortVideoComment
	err = dbConn.Table(shortVideoComment.TableName()).Where("comment_id in (?)", commentIdList).Find(&shortVideoCommentList).Error
	if err != nil {
		log.NewError("GetShortVideoCommentByCommentIdList", "GetShortVideoCommentByCommentIdList", err.Error())
	}

	return shortVideoCommentList, err
}

func GetCommentCountByLevel1CommentId(commentId int64) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	var count int64
	err = dbConn.Table(db.ShortVideoComment{}.TableName()).Where("level1_comment_id=?", commentId).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

func GetCommentCountByLevel0CommentId(commentId int64) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	var count int64
	err = dbConn.Table(db.ShortVideoComment{}.TableName()).Debug().Where("level0_comment_id=?", commentId).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

func GetShortVideoCommentNoPointerByCommentIdList(commentId []int64) ([]db.ShortVideoComment, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var shortVideoComment []db.ShortVideoComment
	err = dbConn.Table("short_video_comment").Where("comment_id IN (?)", commentId).Find(&shortVideoComment).Error
	if err != nil {
		log.NewError("GetShortVideoCommentByCommentIdList", "GetShortVideoCommentByCommentIdList", err.Error())
	}

	return shortVideoComment, err
}

func DeleteShortVideoReply(shortVideoComment *db.ShortVideoComment) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	shortVideoComment.Status = constant.ShortVideoCommentStatusDeleted
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		comment := db.ShortVideoComment{}
		err = tx.Table(shortVideoComment.TableName()).
			Where("comment_id=? AND status != ?", shortVideoComment.CommentId, constant.ShortVideoCommentStatusDeleted).
			First(&comment).Error
		if err != nil {
			return err
		}
		if comment.ParentId == 0 {
			return constant.ErrAccess
		} else {
			specifyMap := make(map[string]interface{})
			specifyMap["comment_reply_count"] = gorm.Expr("comment_reply_count-1")
			tx.Table(db.ShortVideoComment{}.TableName()).Where("comment_id=?", comment.ParentId).Updates(specifyMap)
		}

		return tx.Table(shortVideoComment.TableName()).Updates(shortVideoComment).Error
	})

	return err
}

func GetShortVideoCommentRepliesByWhere(where map[string]string, showNumber int32, pageNumber int32) ([]db.ShortVideoCommentRepliesRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()

	var result []db.ShortVideoCommentRepliesRes
	var count int64
	if err != nil {
		return result, count, err
	}

	var shortVideoPublisher []db.User
	//var shortVideoPublisherMap map[string]*db.User
	var shortVideoPublisherList []string

	var shortVideoList []db.ShortVideo
	var shortVideoMap map[string]*db.ShortVideo
	var shortVideoFileIDList []string

	var commentUsers []db.User
	var commentUserIdList []string
	var commentUserMap map[string]*db.User

	var parentComment []db.ShortVideoComment
	var parentCommentList []int64
	var parentCommentMap map[int64]*db.ShortVideoComment

	var replyUsers []db.User
	var replyUserIdList []string
	var replyUserMap map[string]*db.User

	var replyComment []db.ShortVideoComment

	// query publisher
	queryPublisher := false
	{
		if publisher, ok := where["publisher"]; ok {
			if publisher != "" {
				queryPublisher = true
				shortVideoPublisher = GetUserByAllCondition(publisher)
				shortVideoPublisherList = make([]string, len(shortVideoPublisher))
				//shortVideoPublisherMap = make(map[string]*db.User)
				for index, user := range shortVideoPublisher {
					//shortVideoPublisherMap[user.UserID] = &shortVideoPublisher[index]
					shortVideoPublisherList[index] = user.UserID
				}
			}
		}
	}

	// query short video id list
	queryShortVideo := false
	{
		dbSub := dbConn.Table("short_video")
		if queryPublisher {
			queryShortVideo = true
			dbSub = dbSub.Where("user_id IN (?)", shortVideoPublisherList)
		}
		if privacy, ok := where["privacy"]; ok {
			if privacy != "" && privacy != "0" {
				queryShortVideo = true
				dbSub = dbSub.Where("status=?", privacy)
			}
		}
		contentType := where["content_type"]
		if contentType == "1" {
			if content, ok := where["desc"]; ok {
				queryShortVideo = true
				dbSub = dbSub.Where("`desc` LIKE ?", "%"+content+"%")
			}
		} else if contentType == "2" {
			queryShortVideo = true
			dbSub = dbSub.Where("`desc`=''")
		}
		if queryShortVideo {
			err = dbSub.Debug().Find(&shortVideoList).Error
			if err != nil {
				return nil, 0, err
			}
			shortVideoMap = make(map[string]*db.ShortVideo)
			shortVideoFileIDList = make([]string, len(shortVideoList))
			for index, video := range shortVideoList {
				shortVideoMap[video.FileId] = &shortVideoList[index]
				shortVideoFileIDList[index] = video.FileId
			}
		}
	}

	// query comment user
	queryCommentUser := false
	{
		if commentUser, ok := where["comment_user"]; ok {
			if commentUser != "" {
				queryCommentUser = true
				commentUsers = GetUserByAllCondition(commentUser)
			}
		}
		if queryCommentUser {
			commentUserIdList = make([]string, len(commentUsers))
			commentUserMap = make(map[string]*db.User)
			for index, user := range commentUsers {
				commentUserIdList[index] = user.UserID
				commentUserMap[user.UserID] = &commentUsers[index]
			}
		}
	}

	// parent id list
	queryParentId := false
	{
		dbSub := dbConn.Table("short_video_comment")
		if queryCommentUser {
			queryParentId = true
			dbSub = dbSub.Where("user_id in (?)", commentUserIdList)
		}
		if comment, ok := where["comment"]; ok {
			if comment != "" {
				queryParentId = true
				dbSub = dbSub.Where("content LIKE ?", "%"+comment+"%")
			}
		}
		if queryParentId {
			err = dbSub.Debug().Find(&parentComment).Error
			if err != nil {
				return nil, 0, err
			}
			parentCommentMap = make(map[int64]*db.ShortVideoComment)
			parentCommentList = make([]int64, len(parentComment))
			for index, comment := range parentComment {
				parentCommentList[index] = comment.CommentId
				parentCommentMap[comment.CommentId] = &parentComment[index]
			}
		}
	}

	// reply user, content
	queryReplyUser := false
	{
		if replyUser, ok := where["reply_user"]; ok {
			if replyUser != "" {
				queryReplyUser = true
				replyUsers = GetUserByAllCondition(replyUser)
			}
		}
		if queryReplyUser {
			replyUserIdList = make([]string, len(replyUsers))
			replyUserMap = make(map[string]*db.User)
			for index, user := range replyUsers {
				replyUserIdList[index] = user.UserID
				replyUserMap[user.UserID] = &replyUsers[index]
			}
		}
	}

	// query reply comment(finally)
	{
		dbSub := dbConn.Table("short_video_comment").Where("parent_id != 0").Where("status != 3")
		if commentId, ok := where["comment_id"]; ok {
			if commentId != "" && commentId != "0" {
				dbSub = dbSub.Where("parent_id = ?", commentId)
			}
		}
		if queryShortVideo {
			dbSub = dbSub.Where("file_id IN (?)", shortVideoFileIDList)
		}
		if queryParentId {
			dbSub = dbSub.Where("parent_id IN (?)", parentCommentList)
		}
		if queryReplyUser {
			dbSub = dbSub.Where("user_id IN (?)", replyUserIdList)
		}
		if replyContent, ok := where["reply_content"]; ok {
			if replyContent != "" {
				dbSub = dbSub.Where("content LIKE ?", "%"+replyContent+"%")
			}
		}

		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				dbSub = dbSub.Where("create_time>=?", startTime)
			}
		}
		if endTime, ok := where["end_time"]; ok {
			if endTime != "" {
				dbSub = dbSub.Where("create_time<=?", endTime)
			}
		}

		dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Order("create_time DESC").Find(&replyComment)
	}

	// before link
	{
		for _, comment := range replyComment {
			if !queryShortVideo {
				shortVideoFileIDList = append(shortVideoFileIDList, comment.FileId)
			}
			if !queryParentId {
				parentCommentList = append(parentCommentList, comment.ParentId)
			}
			if !queryReplyUser {
				replyUserIdList = append(replyUserIdList, comment.UserID)
			}
		}
		if !queryShortVideo {
			list, err := GetShortVideoByFileIdList(shortVideoFileIDList)
			if err != nil {
				log.NewError("", "GetShortVideoByFileIdList query error ", err.Error())
				return nil, 0, err
			}
			shortVideoMap = make(map[string]*db.ShortVideo)
			for index, video := range list {
				shortVideoMap[video.FileId] = list[index]
			}
		}
		if !queryParentId {
			parentComment, err = GetShortVideoCommentNoPointerByCommentIdList(parentCommentList)
			if err != nil {
				log.NewError("", "GetShortVideoCommentByCommentIdList query error ", err.Error())
				return nil, 0, err
			}
			parentCommentMap = make(map[int64]*db.ShortVideoComment)
			if !queryCommentUser {
				commentUserIdList = make([]string, len(parentComment))
			}
			for index, comment := range parentComment {
				parentCommentMap[comment.CommentId] = &parentComment[index]
				if !queryCommentUser {
					commentUserIdList[index] = comment.UserID
				}
			}
			//if !queryCommentUser {
			//	commentUsers, err = GetUsersByUserIDList(commentUserIdList)
			//	if err != nil {
			//		log.NewError("", "GetUsersByUserIDList query error ", err.Error())
			//		return nil, 0, err
			//	}
			//	commentUserMap = make(map[string]*db.User)
			//	for index, user := range commentUsers {
			//		commentUserMap[user.UserID] = &commentUsers[index]
			//	}
			//}
		}
		if !queryReplyUser {
			replyUsers, err = GetUsersByUserIDList(replyUserIdList)
			if err != nil {
				log.NewError("", "GetUsersByUserIDList query error ", err.Error())
				return nil, 0, err
			}
			replyUserMap = make(map[string]*db.User)
			for index, user := range replyUsers {
				replyUserMap[user.UserID] = &replyUsers[index]
			}
		}
	}

	// link result
	{
		result = make([]db.ShortVideoCommentRepliesRes, len(replyComment))
		for index, comment := range replyComment {
			videoRes := db.ShortVideoCommentRepliesRes{}
			if parComment, ok := parentCommentMap[comment.ParentId]; ok {
				videoRes.CommentId = parComment.CommentId
				videoRes.CommentContent = parComment.Content
				videoRes.CommentStatus = int64(parComment.Status)
				if parComment.Status == constant.ShortVideoCommentStatusDeleted {
					videoRes.CommentContent = "deleted comment"
				}
			}
			if vid, ok := shortVideoMap[comment.FileId]; ok {
				videoRes.FileId = vid.FileId
				pubUser, _ := GetUserByUserID(vid.UserId)
				if pubUser != nil {
					videoRes.PublishUserID = pubUser.UserID
					videoRes.PublishUser = pubUser.Nickname
				}
				videoRes.ShortVideoStatus = int32(vid.Status)
				videoRes.Content = vid.Desc
				videoRes.CoverUrl = vid.CoverUrl
				videoRes.MediaUrl = vid.MediaUrl
				videoRes.Size = vid.Size
				videoRes.Height = vid.Height
				videoRes.Width = vid.Width
			}

			videoRes.ReplyCommentId = comment.CommentId
			if replyUser, ok := replyUserMap[comment.UserID]; ok {
				videoRes.ReplyUserID = replyUser.UserID
				videoRes.ReplyUserName = replyUser.Nickname
				if replyUser.DeleteTime != 0 {
					videoRes.ReplyUserID = ""
					videoRes.ReplyUserName = "deleted user"
				}
			} else {
				videoRes.ReplyUserID = ""
				videoRes.ReplyUserName = "deleted user"
			}
			videoRes.ReplyCommentContent = comment.Content
			videoRes.ReplyTime = comment.CreateTime
			videoRes.LikeCount = comment.CommentLikeCount
			videoRes.CommentCount = comment.CommentReplyCount
			videoRes.Remark = comment.Remark
			videoRes.Status = int64(comment.Status)

			result[index] = videoRes
		}
	}

	return result, count, nil
}

func GetShortVideoCommentLikesByWhere(where map[string]string, showNumber int32, pageNumber int32) ([]db.ShortVideoCommentLikeRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()

	var result []db.ShortVideoCommentLikeRes
	var count int64
	if err != nil {
		return result, count, err
	}

	var shortVideoPublisher []db.User
	//var shortVideoPublisherMap map[string]*db.User
	var shortVideoPublisherList []string

	var shortVideoList []db.ShortVideo
	var shortVideoMap map[string]*db.ShortVideo
	var shortVideoFileIDList []string

	var commentUsers []db.User
	var commentUserIdList []string
	var commentUserMap map[string]*db.User

	var parentComment []db.ShortVideoComment
	var parentCommentList []int64
	var parentCommentMap map[int64]*db.ShortVideoComment

	var replyUsers []db.User
	var replyUserIdList []string
	var replyUserMap map[string]*db.User

	var likeUsers []db.User
	var likeUserIdList []string
	var likeUserMap map[string]*db.User

	var replyComment []db.ShortVideoComment
	var replyCommentIDList []int64
	var replyCommentMap map[int64]*db.ShortVideoComment

	var likes []db.ShortVideoCommentLike

	// query publisher
	queryPublisher := false
	{
		if publisher, ok := where["publisher"]; ok {
			if publisher != "" {
				queryPublisher = true
				shortVideoPublisher = GetUserByAllCondition(publisher)
				shortVideoPublisherList = make([]string, len(shortVideoPublisher))
				//shortVideoPublisherMap = make(map[string]*db.User)
				for index, user := range shortVideoPublisher {
					//shortVideoPublisherMap[user.UserID] = &shortVideoPublisher[index]
					shortVideoPublisherList[index] = user.UserID
				}
			}
		}
	}

	// query short video id list
	queryShortVideo := false
	{
		dbSub := dbConn.Table("short_video")
		if queryPublisher {
			queryShortVideo = true
			dbSub = dbSub.Where("user_id IN (?)", shortVideoPublisherList)
		}
		if privacy, ok := where["privacy"]; ok {
			if privacy != "" && privacy != "0" {
				queryShortVideo = true
				dbSub = dbSub.Where("status=?", privacy)
			}
		}
		contentType := where["content_type"]
		if contentType == "1" {
			if content, ok := where["desc"]; ok {
				queryShortVideo = true
				dbSub = dbSub.Where("`desc` LIKE ?", "%"+content+"%")
			}
		} else if contentType == "2" {
			queryShortVideo = true
			dbSub = dbSub.Where("`desc`=''")
		}
		if queryShortVideo {
			err = dbSub.Debug().Find(&shortVideoList).Error
			if err != nil {
				return nil, 0, err
			}
			shortVideoMap = make(map[string]*db.ShortVideo)
			shortVideoFileIDList = make([]string, len(shortVideoList))
			for index, video := range shortVideoList {
				shortVideoMap[video.FileId] = &shortVideoList[index]
				shortVideoFileIDList[index] = video.FileId
			}
		}
	}

	// query comment user
	queryCommentUser := false
	{
		if commentUser, ok := where["comment_user"]; ok {
			if commentUser != "" {
				queryCommentUser = true
				commentUsers = GetUserByAllCondition(commentUser)
			}
		}
		if queryCommentUser {
			commentUserIdList = make([]string, len(commentUsers))
			commentUserMap = make(map[string]*db.User)
			for index, user := range commentUsers {
				commentUserIdList[index] = user.UserID
				commentUserMap[user.UserID] = &commentUsers[index]
			}
		}
	}

	// parent id list
	queryParentId := false
	{
		dbSub := dbConn.Table("short_video_comment")
		if queryCommentUser {
			queryParentId = true
			dbSub = dbSub.Where("user_id in (?)", commentUserIdList)
		}
		if comment, ok := where["comment"]; ok {
			if comment != "" {
				queryParentId = true
				dbSub = dbSub.Where("content LIKE ?", "%"+comment+"%")
			}
		}

		if queryParentId {
			err = dbSub.Debug().Find(&parentComment).Error
			if err != nil {
				return nil, 0, err
			}
			parentCommentMap = make(map[int64]*db.ShortVideoComment)
			parentCommentList = make([]int64, len(parentComment))
			for index, comment := range parentComment {
				parentCommentList[index] = comment.CommentId
				parentCommentMap[comment.CommentId] = &parentComment[index]
			}
		}
	}

	// reply user, content
	queryReplyUser := false
	{
		if replyUser, ok := where["reply_user"]; ok {
			if replyUser != "" {
				queryReplyUser = true
				replyUsers = GetUserByAllCondition(replyUser)
			}
		}
		if queryReplyUser {
			replyUserIdList = make([]string, len(replyUsers))
			replyUserMap = make(map[string]*db.User)
			for index, user := range replyUsers {
				replyUserIdList[index] = user.UserID
				replyUserMap[user.UserID] = &replyUsers[index]
			}
		}
	}

	// like user
	queryLikeUser := false
	{
		if likeUser, ok := where["like_user"]; ok {
			if likeUser != "" {
				queryLikeUser = true
				likeUsers = GetUserByAllCondition(likeUser)
			}
		}
		if queryLikeUser {
			likeUserIdList = make([]string, len(likeUsers))
			likeUserMap = make(map[string]*db.User)
			for index, user := range likeUsers {
				likeUserIdList[index] = user.UserID
				likeUserMap[user.UserID] = &likeUsers[index]
			}
		}
	}

	// query reply comment
	queryReplyComment := false
	{
		dbSub := dbConn.Table("short_video_comment").Where("parent_id != 0")
		if queryReplyUser {
			queryReplyComment = true
			dbSub = dbSub.Where("user_id IN (?)", replyUserIdList)
		}
		if replyContent, ok := where["reply_content"]; ok {
			if replyContent != "" {
				queryReplyComment = true
				dbSub = dbSub.Where("content LIKE ?", "%"+replyContent+"%")
			}
		}
		if queryReplyComment && queryParentId {
			queryReplyComment = true
			dbSub = dbSub.Where("parent_id IN (?)", parentCommentList)
		}
		if queryReplyComment {
			dbSub.Debug().Find(&replyComment)
			replyCommentIDList = make([]int64, len(replyComment))
			replyCommentMap = make(map[int64]*db.ShortVideoComment, 0)
			for index, comment := range replyComment {
				replyCommentIDList[index] = comment.CommentId
				replyCommentMap[comment.CommentId] = &replyComment[index]
			}
		}
	}

	// query likes (finally)
	{
		dbSub := dbConn.Table("short_video_comment_like")
		if commentId, ok := where["comment_id"]; ok {
			if commentId != "" && commentId != "0" {
				dbSub = dbSub.Where("comment_id=?", commentId)
			}
		}
		if queryParentId && !queryReplyComment {
			dbSub = dbSub.Where("comment_id IN (?)", parentCommentList)
		} else if queryReplyComment {
			dbSub = dbSub.Where("comment_id IN (?)", replyCommentIDList)
		}

		if queryShortVideo {
			dbSub = dbSub.Where("file_id IN (?)", shortVideoFileIDList)
		}

		if queryLikeUser {
			dbSub = dbSub.Where("user_id IN (?)", likeUserIdList)
		}

		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				dbSub = dbSub.Where("create_time>=?", startTime)
			}
		}
		if endTime, ok := where["end_time"]; ok {
			if endTime != "" {
				dbSub = dbSub.Where("create_time<=?", endTime)
			}
		}

		dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Order("create_time DESC").Debug().Find(&likes)
	}

	// before link
	{
		for _, like := range likes {
			if !queryReplyComment {
				replyCommentIDList = append(replyCommentIDList, like.CommentId)
			}
			if !queryLikeUser {
				likeUserIdList = append(likeUserIdList, like.UserId)
			}
			if !queryShortVideo {
				shortVideoFileIDList = append(shortVideoFileIDList, like.FileId)
			}
		}
		if !queryReplyComment {
			dbConn.Table("short_video_comment").Where("comment_id in (?)", replyCommentIDList).Find(&replyComment)
			replyCommentMap = make(map[int64]*db.ShortVideoComment)
			for index, comment := range replyComment {
				replyCommentMap[comment.CommentId] = &replyComment[index]
			}
		}
		if !queryLikeUser {
			likeUsers, _ = GetUserInfoByUserIDs(likeUserIdList)
			likeUserMap = make(map[string]*db.User)
			for index, user := range likeUsers {
				likeUserMap[user.UserID] = &likeUsers[index]
			}
		}

		for _, comment := range replyComment {

			if !queryParentId {
				parentCommentList = append(parentCommentList, comment.ParentId)
			}
			if !queryReplyUser {
				replyUserIdList = append(replyUserIdList, comment.UserID)
			}
		}
		if !queryShortVideo {
			list, err := GetShortVideoByFileIdList(shortVideoFileIDList)
			if err != nil {
				log.NewError("", "GetShortVideoByFileIdList query error ", err.Error())
				return nil, 0, err
			}
			shortVideoMap = make(map[string]*db.ShortVideo)
			for index, video := range list {
				shortVideoMap[video.FileId] = list[index]
			}
		}
		if !queryParentId {
			parentComment, err = GetShortVideoCommentNoPointerByCommentIdList(parentCommentList)
			if err != nil {
				log.NewError("", "GetShortVideoCommentByCommentIdList query error ", err.Error())
				return nil, 0, err
			}
			parentCommentMap = make(map[int64]*db.ShortVideoComment)
			if !queryCommentUser {
				commentUserIdList = make([]string, len(parentComment))
			}
			for index, comment := range parentComment {
				parentCommentMap[comment.CommentId] = &parentComment[index]
				if !queryCommentUser {
					commentUserIdList[index] = comment.UserID
				}
			}
			//if !queryCommentUser {
			//	commentUsers, err = GetUsersByUserIDList(commentUserIdList)
			//	if err != nil {
			//		log.NewError("", "GetUsersByUserIDList query error ", err.Error())
			//		return nil, 0, err
			//	}
			//	commentUserMap = make(map[string]*db.User)
			//	for index, user := range commentUsers {
			//		commentUserMap[user.UserID] = &commentUsers[index]
			//	}
			//}
		}
		if !queryReplyUser {
			replyUsers, err = GetUsersByUserIDList(replyUserIdList)
			if err != nil {
				log.NewError("", "GetUsersByUserIDList query error ", err.Error())
				return nil, 0, err
			}
			replyUserMap = make(map[string]*db.User)
			for index, user := range replyUsers {
				replyUserMap[user.UserID] = &replyUsers[index]
			}
		}
	}

	// link result
	{
		fmt.Println(parentCommentMap)
		result = make([]db.ShortVideoCommentLikeRes, len(likes))
		for index, like := range likes {
			likeRes := db.ShortVideoCommentLikeRes{}
			if vid, ok := shortVideoMap[like.FileId]; ok {
				likeRes.FileId = vid.FileId
				fmt.Println(vid.UserId)
				pubUser, err := GetUserByUserID(vid.UserId)
				if err != nil {
					likeRes.PublishUserID = ""
					likeRes.PublishUser = "deleted user"
				} else {
					if pubUser != nil {
						likeRes.PublishUserID = pubUser.UserID
						likeRes.PublishUser = pubUser.Nickname
					}
				}
				likeRes.ShortVideoStatus = int32(vid.Status)
				likeRes.Content = vid.Desc
				likeRes.CoverUrl = vid.CoverUrl
				likeRes.MediaUrl = vid.MediaUrl
				likeRes.Size = vid.Size
				likeRes.Height = vid.Height
				likeRes.Width = vid.Width
			}

			parComment := &db.ShortVideoComment{}
			if queryReplyComment {
				if comment, ok := replyCommentMap[like.CommentId]; ok {
					likeRes.ReplyCommentId = comment.CommentId
					likeRes.ReplyCommentContent = comment.Content
					if commUser, ok := replyUserMap[comment.UserID]; ok {
						likeRes.ReplyUserName = commUser.Nickname
						likeRes.ReplyUserName = commUser.UserID
					}
					parComment, err = GetShortVideoCommentByCommentId(comment.ParentId)
					if err != nil {
						continue
					}
				}
			} else if queryParentId {
				parComment = parentCommentMap[like.CommentId]
			} else {
				var ok bool
				var comment *db.ShortVideoComment
				if comment, ok = replyCommentMap[like.CommentId]; !ok {
					comment = parentCommentMap[like.CommentId]
				}
				if comment == nil {
					comment, err = GetShortVideoCommentByCommentId(like.CommentId)
					if err != nil {
						continue
					}
				}
				likeRes.ReplyCommentId = comment.CommentId
				likeRes.ReplyCommentContent = comment.Content
				if commUser, ok := replyUserMap[comment.UserID]; ok {
					likeRes.ReplyUserName = commUser.Nickname
					likeRes.ReplyUserName = commUser.UserID
				}
				parComment, err = GetShortVideoCommentByCommentId(comment.ParentId)
				if err != nil {
					continue
				}
			}
			if parComment != nil {
				likeRes.CommentId = parComment.CommentId
				likeRes.CommentContent = parComment.Content
				var parUser *db.User
				var ok bool
				if parUser, ok = commentUserMap[parComment.UserID]; !ok {
					parUser, _ = GetUserByUserID(parComment.UserID)
				}
				if parUser != nil {
					likeRes.CommentUserName = parUser.Nickname
					likeRes.CommentUserID = parUser.UserID
				} else {
					likeRes.CommentUserID = ""
					likeRes.CommentUserName = "deleted user"
				}
			}
			//if comment, ok := replyCommentMap[like.CommentId]; ok {
			//	if comment.ParentId != 0 {
			//		okp := false
			//		if parComment, okp = parentCommentMap[comment.ParentId]; !okp || parComment == nil {
			//			continue
			//		}
			//		likeRes.ReplyCommentId = comment.CommentId
			//		likeRes.ReplyCommentContent = comment.Content
			//		if commUser, ok := replyUserMap[comment.UserID]; ok {
			//			likeRes.ReplyUserName = commUser.Nickname
			//			likeRes.ReplyUserName = commUser.UserID
			//		}
			//	} else {
			//		parComment = comment
			//	}
			//} else {
			//	parComment = parentCommentMap[like.CommentId]
			//}

			likeRes.LikeId = like.Id
			likeRes.LikeTime = like.CreateTime
			likeRes.Remark = like.Remark

			if likeU, ok := likeUserMap[like.UserId]; ok {
				likeRes.LikeUserID = likeU.UserID
				likeRes.LikeUserName = likeU.Nickname
			}

			result[index] = likeRes
		}
	}

	return result, count, nil
}
