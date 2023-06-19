package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	pbMoments "Open_IM/pkg/proto/moments"
	"Open_IM/pkg/utils"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func InsertMoment(moment db.MomentSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	moment.MCreateTime = time.Now().Local().Unix()
	log.Debug("", "add moment to mysql", moment)
	err = dbConn.Table(moment.TableName()).Create(&moment).Error
	if err != nil {
		return err
	}
	return nil
}

func InsertMomentLike(momentLike db.MomentLikeSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(momentLike.TableName()).Create(&momentLike).Error
	if err != nil {
		return err
	}
	return nil
}

func CancelMomentLike(momentLike db.MomentLikeSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	momentLike.DeleteTime = time.Now().Unix()
	momentLike.DeletedBy = momentLike.CreateBy
	momentLike.Status = 0
	err = dbConn.Table(momentLike.TableName()).Where("moment_id  = ? and user_id = ? ", momentLike.MomentID, momentLike.CreateBy).Updates(&momentLike).Error
	if err != nil {
		return err
	}
	return nil
}

func InsertMomentComment(momentComment db.MomentCommentSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(momentComment.TableName()).Create(&momentComment).Error
	if err != nil {
		return err
	}
	return nil
}

func GetMomentsByUserID(userId string) ([]*db.MomentSQL, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var moments []*db.MomentSQL
	err = dbConn.Table("moments").Where("user_id = ?", userId).Find(&moments).Error
	if err != nil {
		return nil, err
	}
	return moments, err
}

func GetMoment(momentIdList []string) []*db.MomentSQL {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var moments []*db.MomentSQL
	if err != nil {
		return moments
	}
	dbConn.Table("moments").Debug().Where("moment_id IN (?)", momentIdList).Where("delete_time=0").Find(&moments)

	return moments

}

func GetMomentComment(commentIds []string) []*db.MomentCommentSQL {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var comment []*db.MomentCommentSQL
	if err != nil {
		return comment
	}
	dbConn.Table("moments_comments").Where("comment_id IN (?)", commentIds).Where("delete_time=0").Find(&comment)

	return comment

}

func GetMomentLikes(momentIdList []string, userIdList []string) []*db.MomentLikeSQL {
	if len(momentIdList) != len(userIdList) {
		return []*db.MomentLikeSQL{}
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var likes []*db.MomentLikeSQL
	if err != nil {
		return likes
	}
	for index, momentId := range momentIdList {
		momentLike := &db.MomentLikeSQL{}
		dbConn.Table("moments_like").Where("moment_id=?", momentId).Where("user_id=?", userIdList[index]).Where("delete_time=0").Where("status=1").Find(&momentLike)
		likes = append(likes, momentLike)
	}

	return likes

}

func UpdateMoment(moment *db.MomentSQL) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var RowsAffected int64 = 0
	var err1 error
	if err != nil {
		return 0, err
	}
	tx := dbConn.Table("moments").Debug().Where("moment_id=?", moment.MomentID).Updates(moment)
	RowsAffected = tx.RowsAffected
	err1 = tx.Error
	if err1 != nil {
		return RowsAffected, err1
	}

	return RowsAffected, err1
}

func UpdateMomentV2(moment *db.MomentSQL) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var rowsAffected int64 = 0
	if err != nil {
		return rowsAffected, err
	}
	var err1 error

	//likeStr := fmt.Sprintf("m_likes_count+(%d)", moment.MLikesCount)
	//momentStr := fmt.Sprintf("m_comments_count+(%d)", moment.MCommentsCount)
	//repostStr := fmt.Sprintf("m_repost_count+(%d)", moment.MRepostCount)
	//
	//log.Debug("", "likeStr: ", likeStr, " momentStr: ", momentStr, " repostStr: ", repostStr)
	//
	specifyMap := make(map[string]interface{})
	specifyMap["m_likes_count"] = gorm.Expr("m_likes_count+(?)", moment.MLikesCount)
	specifyMap["m_comments_count"] = gorm.Expr("m_comments_count+(?)", moment.MCommentsCount)
	specifyMap["m_repost_count"] = gorm.Expr("m_repost_count+(?)", moment.MRepostCount)

	tx := dbConn.Table("moments").Debug().
		Where("moment_id=?", moment.MomentID).Updates(specifyMap)

	//UpdateColumn("m_likes_count", gorm.Expr("m_likes_count+(?)", moment.MLikesCount)).
	//UpdateColumn("m_comments_count", gorm.Expr("m_comments_count+(?)", moment.MCommentsCount)).
	//UpdateColumn("m_repost_count", gorm.Expr("m_repost_count+(?)", moment.MRepostCount))
	err1 = tx.Error
	rowsAffected = tx.RowsAffected
	if err1 != nil {
		return rowsAffected, err1
	}

	return rowsAffected, err1
}

func UpdateMomentByUserId(moment *db.MomentSQL) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var RowsAffected int64 = 0
	var err1 error
	if err != nil {
		return 0, err
	}
	tx := dbConn.Table("moments").Debug().Where("user_id=?", moment.UserID).Updates(moment)
	RowsAffected = tx.RowsAffected
	err1 = tx.Error
	if err1 != nil {
		return RowsAffected, err1
	}

	return RowsAffected, err1
}

func UpdateMomentCommentV2(comment *db.MomentCommentSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	specifyMap := make(map[string]interface{})
	specifyMap["comment_replies"] = gorm.Expr("comment_replies+(?)", comment.CommentReplies)
	specifyMap["like_counts"] = gorm.Expr("like_counts+(?)", comment.LikeCounts)

	return dbConn.Table("moments_comments").Debug().
		Where("comment_id=?", comment.CommentID).Updates(specifyMap).Error
}

func UpdateMomentComment(comment *db.MomentCommentSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table("moments_comments").Where("user_id=?", comment.UserID).Updates(comment).Error
	if err != nil {
		return err
	}

	return nil
}

func UpdateMomentLikes(like *db.MomentLikeSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table("moments_like").Where("user_id=?", like.UserID).Updates(like).Error
	if err != nil {
		return err
	}

	return nil
}

func GetMomentsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.MomentsRes, int64, error) {

	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var momentsResList []*db.MomentsRes
	var count int64
	if err != nil {
		return momentsResList, count, err
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "m_create_time"

	var users []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User)

	var officialList []db.Official
	var officialIDList []int64
	var officialMap = make(map[int64]*db.Official)

	var articles []db.ArticleSQL
	var articleIDList []int64
	var articleMap = make(map[int64]*db.ArticleSQL)

	var moments []db.MomentSQL

	// Query user info by account, get user id list.
	queryUsers := false
	if account, ok := where["account"]; ok {
		if account != "" {
			queryUsers = true
			users = GetUserByAllCondition(account)
		}
	}
	if queryUsers {
		for index, user := range users {
			userIdList = append(userIdList, user.UserID)
			userMap[user.UserID] = &users[index]
		}
	}

	// Query original user info by original account.
	queryOriginalUser := false
	dbSub := dbConn.Table("official")
	if originalUser, ok := where["original_user"]; ok {
		if originalUser != "" {
			queryOriginalUser = true
			likeUser := "%" + originalUser + "%"
			dbSub = dbSub.Where("nickname like ? OR initial_nickname like ? OR user_id like ?", likeUser, likeUser, likeUser)
		}
	}
	if queryOriginalUser {
		dbSub.Debug().Find(&officialList)
		for index, official := range officialList {
			officialIDList = append(officialIDList, official.Id)
			officialMap[official.Id] = &officialList[index]
		}
		// find articles
		dbConn.Table("article").Where("official_id IN (?)", officialIDList).Find(&articles)
		for index, article := range articles {
			articleIDList = append(articleIDList, article.ArticleID)
			articleMap[article.ArticleID] = &articles[index]
		}
	}

	// Query moments by user id, privacy, media, content, original, status, create time.
	dbSub = dbConn.Table("moments").Where("delete_time=0")
	if queryUsers {
		dbSub = dbSub.Where("creator_id IN (?)", userIdList)
	}
	if queryOriginalUser {
		dbSub = dbSub.Where("article_id IN (?)", articleIDList)
	}
	if orderBy != "" {
		direction := "DESC"
		sort := strings.Split(orderBy, ":")
		if len(sort) == 2 {
			if sort[1] == "asc" {
				direction = "ASC"
			}
		}
		col, ok := sortMap[sort[0]]
		if ok {
			orderByClause = fmt.Sprintf("%s %s ", col, direction)
		}
	}
	if privacy, ok := where["privacy"]; ok {
		if privacy != "" {
			privacyInt, _ := strconv.Atoi(privacy)
			if privacyInt != 0 {
				dbSub = dbSub.Where("privacy=?", privacyInt)
			}
		}
	}

	contentType, ok1 := where["content_type"]
	content, ok2 := where["content"]
	if ok1 && ok2 {
		if contentType == "" || contentType == "0" {
			// all
			if content != "" {
				dbSub = dbSub.Where("m_content_text LIKE ?", "%"+content+"%")
			}
		} else {
			// empty
			dbSub = dbSub.Where("m_content_text =''")
		}
	}
	if mediaType, ok := where["media_type"]; ok {
		if mediaType != "" {
			mediaTypeInt, _ := strconv.Atoi(mediaType)
			if mediaTypeInt == 1 {
				// image
				dbSub = dbSub.Where("m_content_images_array IS NOT NULL AND m_content_images_array != '' AND m_content_images_array != '[]' AND m_content_images_array != 'null'")
			} else if mediaTypeInt == 2 {
				// video
				dbSub = dbSub.Where("m_content_videos_array IS NOT NULL AND m_content_videos_array != '' AND m_content_videos_array != '[]' AND m_content_videos_array != 'null'")
			}
		}
	}
	if isReposted, ok := where["is_reposted"]; ok {
		if isReposted != "" {
			if isReposted == "1" {
				// original
				dbSub = dbSub.Where("article_id=0")
			} else if isReposted == "2" {
				// reposted
				dbSub = dbSub.Where("article_id!=0")
			}
		}
	}
	if isBlocked, ok := where["is_blocked"]; ok {
		if isBlocked != "" {
			isBlockedInt, _ := strconv.Atoi(isBlocked)
			if isBlockedInt != 0 {
				dbSub = dbSub.Where("status=?", isBlockedInt)
			}
		}
	}
	if startTime, ok := where["start_time"]; ok {
		if startTime != "" {
			dbSub = dbSub.Where("m_create_time>=?", startTime)
		}
	}
	if endTime, ok := where["end_time"]; ok {
		if endTime != "" {
			dbSub = dbSub.Where("m_create_time<=?", endTime)
		}
	}
	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}

	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&moments)

	// if there aren't the conditions of create user and original user, need to query user info by moments.
	if !queryOriginalUser || !queryUsers {
		for _, moment := range moments {
			articleIDList = append(articleIDList, moment.ArticleID)
			userIdList = append(userIdList, moment.UserID)
		}

		if !queryOriginalUser {
			dbConn.Table("article").Where("article_id IN (?)", articleIDList).Find(&articles)
			for index, article := range articles {
				officialIDList = append(officialIDList, article.OfficialID)
				articleMap[article.ArticleID] = &articles[index]
			}
			dbConn.Table("official").Where("id IN (?)", officialIDList).Find(&officialList)
			for index, official := range officialList {
				officialMap[official.Id] = &officialList[index]
			}
		}
		if !queryUsers {
			dbConn.Table("users").Where("user_id IN (?)", userIdList).Find(&users)
			for index, user := range users {
				userMap[user.UserID] = &users[index]
			}
		}
	}

	// Query user interest types.
	userInterests := GetInterestsByUserIds(userIdList)

	for _, moment := range moments {
		momentRes := db.MomentsRes{}
		_ = utils.CopyStructFields(&momentRes, &moment)
		log.Debug("", "moment.UserID", moment.UserID)
		if interests, ok := userInterests[moment.UserID]; ok {
			_ = utils.CopyStructFields(&momentRes.Interests, &interests)
		}
		if user, ok := userMap[moment.UserID]; ok {
			if user.DeleteTime == 0 {
				momentRes.UserName = user.Nickname
			}
		}
		if article, ok := articleMap[moment.ArticleID]; ok {
			if official, ok := officialMap[article.OfficialID]; ok {
				momentRes.OriginalCreatorName = official.Nickname
				if official.Nickname == "" {
					momentRes.OriginalCreatorName = official.InitialNickname
				}
				momentRes.OriginalCreatorID = official.UserID
			}
		}

		momentsResList = append(momentsResList, &momentRes)
	}

	return momentsResList, count, nil

}

func DeleteMoments(momentIdList []string, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	nowTime := time.Now().Unix()
	i = 0
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		i += dbConn.Table("moments").
			Where("moment_id IN (?)", momentIdList).
			Where("delete_time=0").
			Updates(db.MomentSQL{DeletedBy: opUserId, DeleteTime: nowTime}).RowsAffected

		return nil
	})

	return i
}

func AlterMoment(moment db.MomentSQL) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	log.Error("Moments update Body", moment)

	moment.MUpdateTime = time.Now().Unix()

	trx := dbConn.Table(moment.TableName()).Updates(&moment)
	if trx.Error != nil {
		log.Error("Moments update failed", trx.Error.Error())
	}
	i = trx.RowsAffected
	return i
}

func ChangeMomentStatus(momentIdList []string, status int8) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	log.Debug("", "status: ", status)
	i = dbConn.Table("moments").Where("moment_id IN (?)", momentIdList).
		Debug().
		Updates(&db.MomentSQL{Status: status, MUpdateTime: time.Now().Unix()}).RowsAffected

	return i
}

func ModifyVisibility(momentIdList []string, privacy int32) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	i = dbConn.Table("moments").Where("moment_id IN (?)", momentIdList).
		Debug().
		Updates(&db.MomentSQL{Privacy: privacy, MUpdateTime: time.Now().Unix()}).RowsAffected

	return i
}

func GetMomentDetailsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.MomentDetailRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var details []*db.MomentDetailRes
	var count int64
	if err != nil {
		return details, count, err
	}
	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "m_create_time"

	var users []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User)

	var moments []db.MomentSQL

	// Query user list by user account
	queryUserInfo := false
	dbSub := dbConn.Table("users")
	if account, ok := where["account"]; ok {
		if account != "" {
			queryUserInfo = true
			likeUser := "%" + account + "%"
			dbSub = dbSub.Where("name like ? OR phone_number like ? OR user_id like ? OR email like ?", likeUser, likeUser, likeUser, likeUser)
		}
	}
	if queryUserInfo {
		dbSub.Find(&users)
		for index, user := range users {
			userIdList = append(userIdList, user.UserID)
			userMap[user.UserID] = &users[index]
		}
	}
	// Query moment
	dbSub = dbConn.Table("moments").Where("delete_time=0")
	momentSearch := false
	if momentId, ok := where["moment_id"]; ok {
		if momentId != "" {
			momentSearch = true
			dbSub = dbSub.Where("moment_id=?", momentId)
		}
	}
	if !momentSearch {
		if orderBy != "" {
			direction := "DESC"
			sort := strings.Split(orderBy, ":")
			if len(sort) == 2 {
				if sort[1] == "asc" {
					direction = "ASC"
				}
			}
			col, ok := sortMap[sort[0]]
			if ok {
				orderByClause = fmt.Sprintf("%s %s ", col, direction)
			}
		}
		if originalId, ok := where["original_id"]; ok {
			if originalId != "" {
				dbSub = dbSub.Where("orignal_id=?", originalId)
			}
		}
		contentType, ok1 := where["content_type"]
		content, ok2 := where["content"]
		if ok1 && ok2 {
			if contentType == "" || contentType == "0" {
				// all
				if content != "" {
					dbSub = dbSub.Where("m_content_text LIKE ?", "%"+content+"%")
				}
			} else {
				// empty
				dbSub = dbSub.Where("m_content_text =''")
			}
		}

		if privacy, ok := where["privacy"]; ok {
			if privacy != "" {
				privacyInt, _ := strconv.Atoi(privacy)
				if privacyInt != 0 {
					dbSub = dbSub.Where("privacy=?", privacyInt)
				}
			}
		}

		if mediaType, ok := where["media_type"]; ok {
			if mediaType != "" {
				mediaTypeInt, _ := strconv.Atoi(mediaType)
				if mediaTypeInt == 1 {
					// image
					dbSub = dbSub.Where("m_content_images_array IS NOT NULL AND m_content_images_array != '' AND m_content_images_array != '[]' AND m_content_images_array != 'null'")
				} else if mediaTypeInt == 2 {
					// video
					dbSub = dbSub.Where("m_content_videos_array IS NOT NULL AND m_content_videos_array != '' AND m_content_videos_array != '[]' AND m_content_videos_array != 'null'")
				}
			}
		}
		if startTime, ok := where["start_time"]; ok {
			if startTime != "" {
				dbSub = dbSub.Where("m_create_time>=?", startTime)
			}
		}
		if endTime, ok := where["end_time"]; ok {
			if endTime != "" {
				dbSub = dbSub.Where("m_create_time<=?", endTime)
			}
		}
		if queryUserInfo {
			dbSub = dbSub.Where("creator_id IN (?)", userIdList)
		}
		if orderByClause != "" {
			log.Debug("inter", "DB orderByClause", orderByClause)
			dbSub = dbSub.Order(orderByClause)
		}
	}
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&moments)

	if !queryUserInfo {
		for _, moment := range moments {
			userIdList = append(userIdList, moment.CreatorID)
		}
		dbConn.Table("users").Where("user_id IN (?)", userIdList).Find(&users)
		for index, user := range users {
			userMap[user.UserID] = &users[index]
		}
	}

	// Link
	for _, moment := range moments {
		momentDetail := &db.MomentDetailRes{}
		utils.CopyStructFields(&momentDetail, &moment)
		if user, ok := userMap[moment.CreatorID]; ok {
			momentDetail.LastLoginIp = user.UpdateIp
		}
		details = append(details, momentDetail)
	}

	return details, count, nil
}

func CtlMomentComment(momentId string, commentCtl int32) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	moment := &db.MomentSQL{
		MomentID:    momentId,
		CommentCtl:  commentCtl,
		MUpdateTime: time.Now().Unix(),
	}
	i = dbConn.Table("moments").Updates(moment).RowsAffected
	return i
}

func GetMomentCommentsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.MomentCommentRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var commentsResList []*db.MomentCommentRes
	var count int64
	if err != nil {
		return commentsResList, count, err
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"

	var userList []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User)

	var commentedUserList []db.User
	var commentedUserIdList []string
	var commentedUserMap = make(map[string]*db.User)

	var commentedCommentList []db.MomentCommentSQL
	var commentedCommentIdList []string
	var commentedUserCommentMap = make(map[string]*db.MomentCommentSQL)

	var commentUserList []db.User
	var commentUserIdList []string
	var commentUserMap = make(map[string]*db.User)

	var momentList []db.MomentSQL
	var momentIdList []string
	var momentMap = make(map[string]*db.MomentSQL)

	var commentList []db.MomentCommentSQL

	// Comment type
	commentType := ""
	ok := false
	if commentType, ok = where["comment_type"]; !ok || commentType == "" {
		commentType = "0"
	}

	// Commented user list
	queryCommented := false
	if commentedUser, ok := where["commented_user"]; ok {
		if commentedUser != "" {
			queryCommented = true
			commentedUserList = GetUserByAllCondition(commentedUser)
			for index, user := range commentedUserList {
				commentedUserIdList = append(commentedUserIdList, user.UserID)
				commentedUserMap[user.UserID] = &commentedUserList[index]
			}
			dbConn.Table("moments_comments").Where("user_id IN (?)", commentedUserIdList).Find(&commentedCommentList)
			for index, comment := range commentedCommentList {
				commentedCommentIdList = append(commentedCommentIdList, comment.CommentID)
				commentedUserCommentMap[comment.CommentID] = &commentedCommentList[index]
			}
		}
	}

	// Query user list, get user id list by user name.
	queryUser := false
	dbSub := dbConn.Table("users")
	if publishUser, ok := where["publish_user"]; ok {
		if publishUser != "" {
			queryUser = true
			userList = GetUserByAllCondition(publishUser)
			for index, user := range userList {
				userIdList = append(userIdList, user.UserID)
				userMap[user.UserID] = &userList[index]
			}
		}
	}

	// Query comment user info list.
	queryCommentUser := false
	dbSub = dbConn.Table("users")
	if commentUser, ok := where["comment_user"]; ok {
		if commentUser != "" {
			queryCommentUser = true
			commentLike := "%" + commentUser + "%"
			dbSub = dbSub.Where("user_id LIKE ? OR name LIKE ?", commentLike, commentLike)
		}
	}
	if queryCommentUser {
		dbSub.Find(&commentUserList)
		for index, user := range commentUserList {
			commentUserIdList = append(commentUserIdList, user.UserID)
			commentUserMap[user.UserID] = &commentUserList[index]
		}
	}

	// Query moment info list by user id, content, type, origin, original user account, publish time.
	queryMoment := false
	dbSub = dbConn.Table("moments")
	if queryUser {
		queryMoment = true
		dbSub = dbSub.Where("creator_id IN (?)", userIdList)
	}
	momentSearch := false
	if momentId, ok := where["moment_id"]; ok {
		if momentId != "" {
			momentSearch = true
			queryMoment = true
			dbSub = dbSub.Where("moment_id=?", momentId)
		}
	}

	if !momentSearch {
		if mediaType, ok := where["media_type"]; ok {
			if mediaType != "" {
				mediaTypeInt, _ := strconv.Atoi(mediaType)
				if mediaTypeInt == 1 {
					// image
					dbSub = dbSub.Where("m_content_images_array IS NOT NULL AND m_content_images_array != '' AND m_content_images_array != '[]' AND m_content_images_array != 'null'")
				} else if mediaTypeInt == 2 {
					// video
					dbSub = dbSub.Where("m_content_videos_array IS NOT NULL AND m_content_videos_array != '' AND m_content_videos_array != '[]' AND m_content_videos_array != 'null'")
				}
			}
		}
		contentType, ok1 := where["content_type"]
		content, ok2 := where["m_content_text"]
		if ok1 && ok2 {
			queryMoment = true
			if contentType == "" || contentType == "0" {
				// all
				if content != "" {
					dbSub = dbSub.Where("m_content_text LIKE ?", "%"+content+"%")
				}
			} else {
				// empty
				dbSub = dbSub.Where("m_content_text =''")
			}
		}
		if privacy, ok := where["privacy"]; ok {
			if privacy != "" {
				privacyInt, _ := strconv.Atoi(privacy)
				if privacyInt != 0 {
					queryMoment = true
					dbSub = dbSub.Where("privacy=?", privacyInt)
				}
			}
		}
		if timeType, ok := where["time_type"]; ok {
			if timeType != "" {
				timeTypeInt, _ := strconv.Atoi(timeType)
				// all or publish time
				if timeTypeInt == 0 || timeTypeInt == 1 {
					if startTime, ok := where["start_time"]; ok {
						if startTime != "" {
							queryMoment = true
							dbSub = dbSub.Where("m_create_time>=?", startTime)
						}
					}
					if endTime, ok := where["end_time"]; ok {
						if endTime != "" {
							queryMoment = true
							dbSub = dbSub.Where("m_create_time<=?", endTime)
						}
					}
				}
			}
		}
	}

	if queryMoment {
		dbSub.Find(&momentList)
		for index, moment := range momentList {
			momentIdList = append(momentIdList, moment.MomentID)
			momentMap[moment.MomentID] = &momentList[index]
			userIdList = append(userIdList, moment.CreatorID)
		}
	}

	// Query comment list by comment time, comment user, comment content.
	dbSub = dbConn.Table("moments_comments").Where("delete_time=0")

	// Search by parent id and not moment's comment.
	if replyCommentId, ok := where["reply_comment_id"]; ok && replyCommentId != "" {
		dbSub = dbSub.Where("reply_comment_id=?", replyCommentId)
	}

	if orderBy != "" {
		direction := "DESC"
		sort := strings.Split(orderBy, ":")
		if len(sort) == 2 {
			if sort[1] == "asc" {
				direction = "ASC"
			}
		}
		col, ok := sortMap[sort[0]]
		if ok {
			orderByClause = fmt.Sprintf("%s %s ", col, direction)
		}
	}
	if queryMoment {
		dbSub = dbSub.Where("moment_id IN (?)", momentIdList)
	}
	if momentSearch {
		dbSub = dbSub.Where("comment_parent_id=''")
	}
	if queryCommentUser {
		dbSub = dbSub.Where("user_id IN (?)", commentUserIdList)
	}

	if timeType, ok := where["time_type"]; ok {
		if timeType != "" {
			timeTypeInt, _ := strconv.Atoi(timeType)
			// all or publish time
			if timeTypeInt == 0 || timeTypeInt == 2 {
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
			}
		}
	}

	if commentContent, ok := where["comment_content"]; ok {
		if commentContent != "" {
			dbSub = dbSub.Where("comment_content LIKE ?", "%"+commentContent+"%")
		}
	}

	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}
	dbSub.Debug().Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&commentList)

	if !queryUser || !queryCommentUser || !queryMoment || queryCommented {
		for _, comment := range commentList {
			if !queryCommentUser {
				commentUserIdList = append(commentUserIdList, comment.CreateBy)
			}
			if !queryMoment {
				momentIdList = append(momentIdList, comment.MomentID)
			}
			if !queryCommented {
				commentedCommentIdList = append(commentedCommentIdList, comment.CommentParentID)
			}
		}
		if !queryCommentUser {
			// comment user
			dbConn.Table("users").Where("user_id IN (?)", commentUserIdList).Find(&commentUserList)
			for index, user := range commentUserList {
				commentUserIdList = append(commentUserIdList, user.UserID)
				commentUserMap[user.UserID] = &commentUserList[index]
			}
		}
		if !queryMoment {
			// moment
			dbConn.Table("moments").Where("moment_id IN (?)", momentIdList).Find(&momentList)
			for index, moment := range momentList {
				momentIdList = append(momentIdList, moment.MomentID)
				momentMap[moment.MomentID] = &momentList[index]
				userIdList = append(userIdList, moment.CreatorID)
			}
		}
		if !queryUser {
			dbConn.Table("users").Where("user_id IN (?)", userIdList).Find(&userList)
			for index, user := range userList {
				userIdList = append(userIdList, user.UserID)
				userMap[user.UserID] = &userList[index]
			}
		}
		if !queryCommented {
			dbConn.Table("moments_comments").Where("comment_id IN (?)", commentedCommentIdList).Find(&commentedCommentList)

			for index, comment := range commentedCommentList {
				commentedUserCommentMap[comment.CommentID] = &commentedCommentList[index]
			}
		}
	}

	// Link res var commentsResList []*db.MomentCommentRes
	for _, comment := range commentList {
		commentRes := &db.MomentCommentRes{}
		_ = utils.CopyStructFields(&commentRes, &comment)
		commentRes.CommentParentId = comment.CommentParentID
		commentRes.ReplyCommentId = comment.ReplyCommentID
		log.Debug("", "moment id: ", comment.MomentID)
		if moment, ok := momentMap[comment.MomentID]; ok {
			commentRes.MCreateTime = moment.MCreateTime
			commentRes.MContentText = moment.MContentText
			commentRes.MContentImagesArray = moment.MContentImagesArray
			commentRes.MContentVideosArray = moment.MContentVideosArray
			commentRes.Privacy = moment.Privacy
			log.Debug("", "[moment.UserID: ", moment.UserID)
			if user, ok := userMap[moment.UserID]; ok {
				commentRes.PublishName = user.Nickname
				commentRes.PublishAccount = user.UserID
			}
		}
		if replyComment, ok := commentedUserCommentMap[comment.CommentParentID]; ok {
			commentRes.CommentedUseID = replyComment.UserID
			commentRes.CommentedUserName = replyComment.UserName
		}

		commentsResList = append(commentsResList, commentRes)
	}

	return commentsResList, count, nil
}

func DeleteMomentComments(commentIdList []string, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	i = dbConn.Table("moments_comments").
		Where("comment_id IN (?)", commentIdList).
		Where("delete_time=0").Debug().
		Updates(db.MomentCommentSQL{DeletedBy: opUserId, DeleteTime: time.Now().Unix()}).
		RowsAffected

	return i
}

func AlterMomentComment(param *pbMoments.AlterCommentReq) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	comment := &db.MomentCommentSQL{
		CommentContent: param.Content,
		UpdateBy:       param.OpUserId,
		UpdatedTime:    time.Now().Unix(),
	}
	i := dbConn.Table("moments_comments").Where("comment_id=?", param.CommentId).Where("delete_time=0").Updates(&comment).RowsAffected
	return i
}

func SwitchMomentCommentHideState(param *pbMoments.SwitchCommentHideStateReq) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	comment := &db.MomentCommentSQL{
		UpdatedTime: time.Now().Unix(),
		Status:      int8(param.Status),
		UpdateBy:    param.OpUserId,
	}

	i := dbConn.Table("moments_comments").Where("comment_id=?", param.CommentId).Debug().Updates(comment).RowsAffected
	return i
}

func GetMomentLikesByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.MomentLikeRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var likeResList []*db.MomentLikeRes
	var count int64
	if err != nil {
		return likeResList, count, err
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"

	var userList []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User)

	var likeUserList []db.User
	var likeUserIdList []string
	var likeUserMap = make(map[string]*db.User)

	var momentList []db.MomentSQL
	var momentIdList []string
	var momentMap = make(map[string]*db.MomentSQL)

	var likeList []db.MomentLikeSQL

	// Query moment user list, get user id list by user name.
	queryUser := false
	if publishUser, ok := where["publish_user"]; ok {
		if publishUser != "" {
			queryUser = true
			userList = GetUserByAllCondition(publishUser)
		}
	}
	if queryUser {
		for index, user := range userList {
			userIdList = append(userIdList, user.UserID)
			userMap[user.UserID] = &userList[index]
		}
	}

	// Query like user info list.
	queryLikeUser := false
	dbSub := dbConn.Table("users")
	if likeUser, ok := where["like_user"]; ok {
		if likeUser != "" {
			queryLikeUser = true
			likeLike := "%" + likeUser + "%"
			dbSub = dbSub.Where("user_id LIKE ? OR name LIKE ?", likeLike, likeLike)
		}
	}
	if queryLikeUser {
		dbSub.Find(&likeUserList)
		for index, user := range likeUserList {
			likeUserIdList = append(likeUserIdList, user.UserID)
			likeUserMap[user.UserID] = &likeUserList[index]
		}
	}

	// Query moment info list by user id, content, type, origin, original user account, publish time.
	queryMoment := false
	dbSub = dbConn.Table("moments")
	momentSearch := false
	if momentId, ok := where["moment_id"]; ok {
		if momentId != "" {
			queryMoment = true
			momentSearch = true
			dbSub = dbSub.Where("moment_id=?", momentId)
		}
	}
	if !momentSearch {
		if queryUser {
			queryMoment = true
			dbSub = dbSub.Where("creator_id IN (?)", userIdList)
		}
		contentType, ok1 := where["content_type"]
		content, ok2 := where["m_content_text"]
		if ok1 && ok2 {
			queryMoment = true
			if contentType == "" || contentType == "0" {
				// all
				if content != "" {
					dbSub = dbSub.Where("m_content_text LIKE ?", "%"+content+"%")
				}
			} else {
				// empty
				dbSub = dbSub.Where("m_content_text =''")
			}
		}
		if mediaType, ok := where["media_type"]; ok {
			if mediaType != "" {
				mediaTypeInt, _ := strconv.Atoi(mediaType)
				if mediaTypeInt == 1 {
					// image
					dbSub = dbSub.Where("m_content_images_array IS NOT NULL AND m_content_images_array != '' AND m_content_images_array != '[]' AND m_content_images_array != 'null'")
				} else if mediaTypeInt == 2 {
					// video
					dbSub = dbSub.Where("m_content_videos_array IS NOT NULL AND m_content_videos_array != '' AND m_content_videos_array != '[]' AND m_content_videos_array != 'null'")
				}
			}
		}
		if privacy, ok := where["privacy"]; ok {
			if privacy != "" {
				queryMoment = true
				privacyInt, _ := strconv.Atoi(privacy)
				if privacyInt != 0 {
					dbSub = dbSub.Where("privacy=?", privacyInt)
				}
			}
		}
		if timeType, ok := where["time_type"]; ok {
			if timeType != "" {
				timeTypeInt, _ := strconv.Atoi(timeType)
				// all or publish time
				if timeTypeInt == 0 || timeTypeInt == 1 {
					if startTime, ok := where["start_time"]; ok {
						if startTime != "" {
							queryMoment = true
							dbSub = dbSub.Where("m_create_time>=?", startTime)
						}
					}
					if endTime, ok := where["end_time"]; ok {
						if endTime != "" {
							queryMoment = true
							dbSub = dbSub.Where("m_create_time<=?", endTime)
						}
					}
				}
			}
		}

	}
	if queryMoment {
		dbSub.Find(&momentList)
		for index, moment := range momentList {
			momentIdList = append(momentIdList, moment.MomentID)
			momentMap[moment.MomentID] = &momentList[index]
			userIdList = append(userIdList, moment.CreatorID)
		}
	}

	// Query comment list by comment time, comment user, comment content.
	dbSub = dbConn.Table("moments_like").Where("delete_time=0")
	if orderBy != "" {
		direction := "DESC"
		sort := strings.Split(orderBy, ":")
		if len(sort) == 2 {
			if sort[1] == "asc" {
				direction = "ASC"
			}
		}
		col, ok := sortMap[sort[0]]
		if ok {
			orderByClause = fmt.Sprintf("%s %s ", col, direction)
		}
	}
	if queryMoment {
		dbSub = dbSub.Where("moment_id IN (?)", momentIdList)
	}
	if queryLikeUser {
		dbSub = dbSub.Where("user_id IN (?)", likeUserIdList)
	}

	if timeType, ok := where["time_type"]; ok {
		if timeType != "" {
			timeTypeInt, _ := strconv.Atoi(timeType)
			// all or publish time
			if timeTypeInt == 0 || timeTypeInt == 2 {
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
			}
		}
	}

	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&likeList)
	if !queryUser || !queryLikeUser || !queryMoment {
		for _, like := range likeList {
			if !queryLikeUser {
				likeUserIdList = append(likeUserIdList, like.CreateBy)
			}
			if !queryMoment {
				momentIdList = append(momentIdList, like.MomentID)
			}
		}
		if !queryLikeUser {
			// comment user
			dbConn.Table("users").Where("user_id IN (?)", likeUserIdList).Find(&likeUserList)
			for index, user := range likeUserList {
				likeUserIdList = append(likeUserIdList, user.UserID)
				likeUserMap[user.UserID] = &likeUserList[index]
			}
		}
		if !queryMoment {
			// moment
			dbConn.Table("moments").Where("moment_id IN (?)", momentIdList).Find(&momentList)
			for index, moment := range momentList {
				momentIdList = append(momentIdList, moment.MomentID)
				momentMap[moment.MomentID] = &momentList[index]
				userIdList = append(userIdList, moment.CreatorID)
			}
		}
		if !queryUser {
			dbConn.Table("users").Where("user_id IN (?)", userIdList).Find(&userList)
			for index, user := range userList {
				userIdList = append(userIdList, user.UserID)
				userMap[user.UserID] = &userList[index]
			}
		}
	}

	// Link res
	for _, like := range likeList {
		likeRes := &db.MomentLikeRes{}
		_ = utils.CopyStructFields(&likeRes, &like)
		if moment, ok := momentMap[like.MomentID]; ok {
			likeRes.MCreateTime = moment.MCreateTime
			likeRes.MContentText = moment.MContentText
			likeRes.MContentImagesArray = moment.MContentImagesArray
			likeRes.MContentVideosArray = moment.MContentVideosArray
			likeRes.Privacy = moment.Privacy
			if user, ok := userMap[moment.UserID]; ok && user.DeleteTime == 0 {
				likeRes.Account = user.UserID
				likeRes.AccountNickname = user.Nickname
			}
		}

		likeResList = append(likeResList, likeRes)
	}

	return likeResList, count, nil
}

func RemoveMomentsLikes(momentIdList []string, userIdList []string, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	nowTime := time.Now().Unix()

	for index, momentId := range momentIdList {
		like := &db.MomentLikeSQL{
			DeleteTime: nowTime,
			DeletedBy:  opUserId,
		}
		i += dbConn.Table("moments_like").
			Where("moment_id=?", momentId).
			Where("user_id=?", userIdList[index]).
			Where("delete_time=0").
			Updates(&like).RowsAffected
	}
	return i
}

func SwitchMomentLikeHideState(param *pbMoments.SwitchLikeHideStateReq) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	nowTime := time.Now().Unix()
	i = dbConn.Table("moments_like").Debug().
		Where("moment_id=?", param.MomentId).
		Where("user_id=?", param.UserId).
		Updates(db.MomentLikeSQL{
			UpdatedTime: nowTime,
			UpdateBy:    param.OpUserId,
			Status:      int8(param.Status),
		}).RowsAffected

	return i
}

func DeleteMomentCommentByID(commentId string, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	i = dbConn.Table("moments_comments").
		Where("comment_id =? ", commentId).
		Updates(db.MomentCommentSQL{DeletedBy: opUserId, DeleteTime: time.Now().Unix()}).
		RowsAffected

	return i
}
