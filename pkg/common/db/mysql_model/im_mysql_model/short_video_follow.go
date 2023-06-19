package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
)

func InsertShortVideoFollow(userId, fansId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	data := db.ShortVideoFollow{
		UserId: userId,
		FansId: fansId,
	}
	err = dbConn.Table(db.ShortVideoFollow{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("InsertShortVideoFollow", "InsertShortVideoFollow", err.Error())
		return false
	}
	return true
}

func GetFollowList(userId, fansId string, pageNumber, showNumber int32) ([]*db.ShortVideoFollow, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	condition := map[string]interface{}{}
	if userId != "" {
		condition["user_id"] = userId
	}
	if fansId != "" {
		condition["fans_id"] = fansId
	}

	shortVideoFollow := &db.ShortVideoFollow{}
	var count int64

	err = dbConn.Table(shortVideoFollow.TableName()).Where(condition).Count(&count).Error
	if err != nil {
		log.NewError("GetShortVideoCommentList", "GetShortVideoCommentList", err.Error())
		return nil, 0, err
	}

	var ShortVideoFollowList []*db.ShortVideoFollow
	err = dbConn.Table(shortVideoFollow.TableName()).Where(condition).Order("id DESC").
		Offset(int((pageNumber - 1) * showNumber)).Limit(int(showNumber)).Find(&ShortVideoFollowList).Error
	if err != nil {
		log.NewError("GetShortVideoFollowList", "GetShortVideoFollowList", err.Error())
		return nil, 0, err
	}

	return ShortVideoFollowList, count, nil
}

func GetShortVideoFollowUserIdByFansId(userId string) ([]interface{}, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var data []string
	err = dbConn.Table(db.ShortVideoFollow{}.TableName()).Where("fans_id = ?", userId).Pluck("user_id", &data).Error
	if err != nil {
		log.NewError("GetShortVideoFollowUserIdByFansId", "GetShortVideoFollowUserIdByFansId", err.Error())
		return nil, err
	}

	result := utils.StringArrayToInterfaceArray(data)
	return result, nil
}

func GetFansIdByUserId(userId string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var data []string
	err = dbConn.Table(db.ShortVideoFollow{}.TableName()).Where("user_id = ?", userId).Pluck("fans_id", &data).Error
	if err != nil {
		log.NewError("GetFollowerUserIdByFollowUserId", "GetFollowerUserIdByFollowUserId", err.Error())
		return nil, err
	}
	return data, nil
}

func DeleteShortVideoFollow(userId, fansId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err = dbConn.Table(db.ShortVideoFollow{}.TableName()).Where("user_id = ? and fans_id = ?", userId, fansId).Delete(&db.ShortVideoFollow{}).Error
	if err != nil {
		log.NewError("DeleteShortVideoFollow", "DeleteShortVideoFollow", err.Error())
		return false
	}
	return true
}

func GetFollowersByWhere(where map[string]string, showNumber int32, pageNumber int32) ([]db.ShortVideoFollowRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var result []db.ShortVideoFollowRes
	var count int64
	if err != nil {
		return result, count, err
	}

	var followerIdList []string
	var followerUsers []db.User
	var followerUserMap map[string]*db.User

	var followedIdList []string
	var followedUsers []db.User
	var followedUserMap map[string]*db.User

	queryFollower := false
	{
		if follower, ok := where["follower"]; ok {
			if follower != "" {
				queryFollower = true
				followerUsers = GetUserByAllCondition(follower)
			}
		}
		if queryFollower {
			followerIdList = make([]string, len(followerUsers))
			followerUserMap = make(map[string]*db.User, len(followerUsers))
			for index, user := range followerUsers {
				followerIdList[index] = user.UserID
				followerUserMap[user.UserID] = &followerUsers[index]
			}
		}
	}

	queryFollowed := false
	{
		if followedUser, ok := where["followed_user"]; ok {
			if followedUser != "" {
				queryFollowed = true
				followedUsers = GetUserByAllCondition(followedUser)
			}
		}
		if queryFollowed {
			followedIdList = make([]string, len(followedUsers))
			followedUserMap = make(map[string]*db.User, len(followedUsers))

			for index, user := range followedUsers {
				followedIdList[index] = user.UserID
				followedUserMap[user.UserID] = &followedUsers[index]
			}
		}
	}

	dbSub := dbConn.Table("short_video_follow")
	if queryFollower {
		dbSub = dbSub.Where("user_id IN (?)", followerIdList)
	}
	if queryFollowed {
		dbSub = dbSub.Where("fans_id IN (?)", followedIdList)
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

	var followerDBRes []db.ShortVideoFollow
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Order("create_time DESC").Debug().Find(&followerDBRes)

	for _, follower := range followerDBRes {
		if !queryFollower {
			followerIdList = append(followerIdList, follower.UserId)
		}
		if !queryFollowed {
			followedIdList = append(followedIdList, follower.FansId)
		}
	}
	if !queryFollower {
		followerUsers, _ = GetUsersByUserIDList(followerIdList)
		followerUserMap = make(map[string]*db.User, len(followerUsers))
		for index, user := range followerUsers {
			followerUserMap[user.UserID] = &followerUsers[index]
		}
	}
	if !queryFollowed {
		followedUsers, _ = GetUsersByUserIDList(followedIdList)

		followedUserMap = make(map[string]*db.User, len(followedUsers))
		for index, user := range followedUsers {
			followedUserMap[user.UserID] = &followedUsers[index]
		}
	}

	result = make([]db.ShortVideoFollowRes, len(followerDBRes))
	for index, follower := range followerDBRes {
		follow := db.ShortVideoFollowRes{}
		_ = utils.CopyStructFields(&follow, follower)
		if user, ok := followerUserMap[follow.UserId]; ok {
			follow.UserName = user.Nickname
			follow.UserFace = user.FaceURL
		} else {
			follow.UserName = "deleted user"
		}
		if user, ok := followedUserMap[follow.FansId]; ok {
			follow.FansName = user.Nickname
			follow.FansFace = user.FaceURL
		} else {
			follow.UserName = "deleted user"
		}

		result[index] = follow
	}

	return result, count, nil
}

func AlterFollower(follower *db.ShortVideoFollow) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	result := dbConn.Table("short_video_follow").Updates(follower)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, result.Error
}

func GetFollowersByIdList(id []int64) ([]db.ShortVideoFollow, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var result []db.ShortVideoFollow
	err = dbConn.Table("short_video_follow").Where("id IN (?)", id).Find(&result).Error
	if err != nil {
		return result, err
	}
	return result, err
}
