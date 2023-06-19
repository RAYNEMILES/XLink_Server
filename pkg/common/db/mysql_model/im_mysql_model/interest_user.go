package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func SetUserInterestType(userId string, interestType []int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	rwMutex.Lock()
	defer rwMutex.Unlock()

	dbConn.Where("user_id = ?", userId).Delete(&db.InterestUser{})

	//if len(interestType) == 0 {
	//	// default
	//	interestType = append(interestType, constant.InterestDefault)
	//}

	var data []db.InterestUser
	now := time.Now().Unix()
	for _, interestId := range interestType {
		data = append(data, db.InterestUser{UserId: userId, InterestId: interestId, UpdateTime: now})
	}
	err = dbConn.Table(db.InterestUser{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("", "UpdateInterestUser", "err", err, "data", data)
		return false
	}
	return true
}

func GetUserInterestList(userId string) ([]int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var interests []int64
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		err = tx.Table(db.InterestUser{}.TableName()).Where("user_id", userId).Pluck("interest_id", &interests).Error
		if err != nil {
			return err
		}
		if len(interests) == 0 {
			err = tx.Table("interest_type").Where("is_default = ?", constant.InterestIsDefaultType).Where("delete_time=0").Pluck("id", &interests).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return interests, nil
}

func GetDefaultInterestUserIDList() ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var result []string
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		var interestUserIDList []string
		err = tx.Table("interest_user").Select("user_id").Distinct("user_id").Find(&interestUserIDList).Error
		if err != nil {
			return err
		}
		err = tx.Table("users").Select("user_id").Where("user_id NOT IN (?)", interestUserIDList).Find(&result).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetUserInterestsByWhereV3(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.UserInterests, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var interestsRes []*db.UserInterests
	var count int64
	if err != nil {
		return interestsRes, count, err
	}
	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "update_time"

	var userList []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User, 0)

	var queryInterestUserIDList []string

	var resultUserList []db.User
	var resultUserIdList []string

	var interestIdList []int64
	var interestLanguageList []db.InterestLanguage

	var userInterestResMap map[string][]*db.InterestTypeRes

	defaultStatus := 0

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


	notDefaultInterestIDList := GetNotDefaultInterestIdList()
	defaultUserIDList, err := GetDefaultInterestUserIDList()
	var notDefaultInterestIDMap = make(map[int64]struct{})
	for _, interestid := range notDefaultInterestIDList {
		notDefaultInterestIDMap[interestid] = struct{}{}
	}
	var notDefaultUserIDMap = make(map[string]struct{})
	for _, userid := range defaultUserIDList {
		notDefaultUserIDMap[userid] = struct{}{}
	}

	if userID, ok := where["user_id"]; ok && userID != "" {
		dbConn.Table("users").Where("delete_time=0").Where("user_id = ?", userID).
			Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Order("create_time DESC").Debug().Find(&resultUserList)
	} else {
		// if searching is default tags, then set the default status is 1 or 2, 2 is default, 1 is personal
		if isDefaultStr, ok := where["default"]; ok {
			if isDefaultStr != "" {
				defaultStatus, _ = strconv.Atoi(isDefaultStr)
			}
		}

		// Query user list
		queryUser := false
		dbSub := dbConn.Table("users")
		if account, ok := where["account"]; ok {
			if account != "" {
				queryUser = true
				userList = GetUserByAllCondition(account)
				for index, user := range userList {
					userIdList = append(userIdList, user.UserID)
					userMap[user.UserID] = &userList[index]
				}
			}
		}

		if err != nil {
			log.NewError("", "query default user id list failed")
			return nil, 0, err
		}

		switch defaultStatus {
		case 0:
			hasDefault := false
			queryInterestsList := false
			dbSub = dbConn.Table("interest_language")
			if interestName, ok := where["interest_name"]; ok {
				if interestName != "" {
					queryInterestsList = true
					dbSub = dbSub.Where("name LIKE ?", "%"+interestName+"%")
				}
			}
			if queryInterestsList {
				dbSub.Debug().Find(&interestLanguageList)
				for _, language := range interestLanguageList {
					if !hasDefault {
						if _, ok := notDefaultInterestIDMap[language.InterestId]; !ok && len(defaultUserIDList) > 0 {
							hasDefault = true
						}
					}
					interestIdList = append(interestIdList, language.InterestId)
				}
			}

			dbSub = dbConn.Table("users").Where("delete_time=0")
			if queryUser && !queryInterestsList {
				dbSub = dbSub.Where("user_id IN (?)", userIdList)
			} else if queryUser && queryInterestsList {
				dbConn.Table("interest_user").Distinct().Select("user_id").Where("interest_id IN (?)", interestIdList).Find(&queryInterestUserIDList)
				dbSub = dbSub.Where("user_id IN (?)", userIdList)
				if hasDefault {
					dbSub = dbSub.Where("user_id IN (?) OR user_id IN (?)", queryInterestUserIDList, defaultUserIDList)
				} else {
					dbSub = dbSub.Where("user_id IN (?)", queryInterestUserIDList)
				}
			} else if !queryUser && queryInterestsList {
				dbConn.Table("interest_user").Distinct().Select("user_id").Where("interest_id IN (?)", interestIdList).Find(&queryInterestUserIDList)
				if hasDefault {
					dbSub = dbSub.Where("user_id IN (?) OR user_id IN (?)", queryInterestUserIDList, defaultUserIDList)
				} else {
					dbSub = dbSub.Where("user_id IN (?)", queryInterestUserIDList)
				}
			}
			//
			//sql := ""
			//var params []interface{}
			//if queryInterestsList {
			//	dbConn.Table("interest_user").Distinct().Select("user_id").Where("interest_id IN (?)", interestIdList).Find(&queryInterestUserIDList)
			//	sql = "user_id IN (?)"
			//	params = append(params, queryInterestUserIDList)
			//}
			//fmt.Println("queryInterestsList: ", queryInterestsList)
			//fmt.Println("hasDefault: ", hasDefault)
			//fmt.Println("queryUser: ", queryUser)
			//if !queryInterestsList || hasDefault {
			//	if sql == "" {
			//		sql = "user_id IN (?)"
			//	}
			//	params = append(params, defaultUserIDList)
			//}
			//fmt.Println(sql)
			//if sql != "" {
			//	dbSub = dbSub.Where(sql, params...)
			//}
			dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Order("create_time DESC").Debug().Find(&resultUserList)
		case 1:
			// personal
			queryInterestsList := false
			dbSub = dbConn.Table("interest_language")
			if interestName, ok := where["interest_name"]; ok {
				if interestName != "" {
					queryInterestsList = true
					dbSub = dbSub.Where("name LIKE ?", "%"+interestName+"%")
				}
			}
			if defaultStatus == 1 {
				queryInterestsList = true
				dbSub = dbSub.Where("interest_id IN (?)", notDefaultInterestIDList)
			}
			if queryInterestsList {
				dbSub.Debug().Find(&interestLanguageList)
				for _, language := range interestLanguageList {
					interestIdList = append(interestIdList, language.InterestId)
				}
			}
			dbSub = dbConn.Table("users").Where("delete_time=0")
			interUserDBSub := dbConn.Table("interest_user").Distinct().Select("user_id")
			if queryInterestsList {
				interUserDBSub.Where("interest_id IN (?)", interestIdList).Find(&queryInterestUserIDList)
			} else {
				interUserDBSub.Find(&queryInterestUserIDList)
			}
			dbSub = dbSub.Where("user_id IN (?)", queryInterestUserIDList)
			if queryUser {
				dbSub = dbSub.Where("user_id IN (?)", userIdList)
			}
			if orderByClause != "" {
				dbSub.Order(orderByClause)
			}
			dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Order("create_time DESC").Debug().Find(&resultUserList)
		case 2:
			// default
			dbSub = dbConn.Table("users").Where("delete_time=0").Where("user_id IN (?)", defaultUserIDList)
			if queryUser {
				dbSub = dbSub.Where("user_id IN (?)", userIdList)
			}
			if orderByClause != "" {
				dbSub.Order(orderByClause)
			}
			dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Debug().Find(&resultUserList)
		}

	}

	resultUserIdList = make([]string, len(resultUserList))
	for index, user := range resultUserList {
		resultUserIdList[index] = user.UserID
	}

	userInterestResMap = GetInterestsByUserIds(resultUserIdList)

	interestsRes = make([]*db.UserInterests, len(resultUserList))
	for index, user := range resultUserList {
		interRes := &db.UserInterests{}
		if _, ok := notDefaultUserIDMap[user.UserID]; ok {
			interRes.Type = 0
		} else {
			interRes.Type = 1
		}
		interRes.UserID = user.UserID
		interRes.Username = user.Nickname

		interRes.Interests = userInterestResMap[user.UserID]

		interestsRes[index] = interRes
	}

	return interestsRes, count, nil
}

func GetUserInterestsByWhereV2(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.UserInterests, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var interestsRes []*db.UserInterests
	var count int64
	if err != nil {
		return interestsRes, count, err
	}
	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "update_time"

	var userList []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User, 0)

	var resultUserList []db.User
	var resultUserIdList []string
	var resultUserMap = make(map[string]*db.User, 0)

	var interestIdList []int64
	var interestLanguageList []db.InterestLanguage

	var userInterestResMap map[string][]*db.InterestTypeRes

	defaultStatus := 0

	// if searching is default tags, then set the default status is 1 or 2, 2 is default, 1 is personal
	if isDefaultStr, ok := where["default"]; ok {
		if isDefaultStr != "" {
			defaultStatus, _ = strconv.Atoi(isDefaultStr)
		}
	}

	// Query user list
	queryUser := false
	dbSub := dbConn.Table("users")
	if account, ok := where["account"]; ok {
		if account != "" {
			queryUser = true
			userList = GetUserByAllCondition(account)
			for index, user := range userList {
				userIdList = append(userIdList, user.UserID)
				userMap[user.UserID] = &userList[index]
			}
		}
	}

	// Query interests list by interest language
	queryInterestsList := false
	dbSub = dbConn.Table("interest_language")
	if interestName, ok := where["interest_name"]; ok {
		if interestName != "" {
			queryInterestsList = true
			dbSub = dbSub.Where("name LIKE ?", "%"+interestName+"%")
		}
	}
	if queryInterestsList {
		dbSub.Debug().Find(&interestLanguageList)
		def := false
		defIntList := GetDefaultInterestTypeList()
		defIntMap := map[int64]struct{}{}
		for _, id := range defIntList {
			defIntMap[id] = struct{}{}
		}
		for _, language := range interestLanguageList {
			interestIdList = append(interestIdList, language.InterestId)
			if _, ok := defIntMap[language.InterestId]; ok && !def {
				interestIdList = append(interestIdList, -1)
				def = true
			}
		}
	}

	// Query interests-user list's user id
	var limitUserId []string
	dbSub = dbConn.Table("interest_user").Distinct().Select("user_id")

	if queryUser {
		dbSub = dbSub.Where("user_id IN (?)", userIdList)
	}
	if queryInterestsList {
		dbSub = dbSub.Where("interest_id IN (?)", interestIdList)
	}

	// if is personal, query interest id isn't -1
	if defaultStatus == 1 {
		// not default
		queryInterestsList = true
		dbSub = dbSub.Where("interest_id != -1")
	} else if defaultStatus == 2 {
		// default
		queryInterestsList = true
		dbSub = dbSub.Where("interest_id = -1")
	}

	// limit user id
	if queryInterestsList {
		dbSub.Debug().Find(&limitUserId)
	}

	// fmt.Println(queryInterestsList)
	dbSub = dbConn.Table("users").Where("delete_time=0")

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

	if queryInterestsList {
		dbSub = dbSub.Where("user_id IN (?)", limitUserId)
	}

	if queryUser {
		dbSub = dbSub.Where("user_id IN (?)", userIdList)
	}

	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}

	dbSub.Debug().Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Debug().Find(&resultUserList)

	resultUserIdList = make([]string, len(resultUserList))
	for index, user := range resultUserList {
		resultUserIdList[index] = user.UserID
		resultUserMap[user.UserID] = &resultUserList[index]
	}
	// get user's interests, include -1
	userInterestResMap = GetInterestsByUserIds(resultUserIdList)
	defaultUserId, err := GetDefaultInterestUserIDList()
	if err != nil {
		return nil, count, err
	}
	defaultUserMap := map[string]struct{}{}
	for _, id := range defaultUserId {
		defaultUserMap[id] = struct{}{}
	}
	interestsRes = make([]*db.UserInterests, len(resultUserList))
	for index, user := range resultUserList {
		interestsRes[index] = &db.UserInterests{}
		if _, ok := defaultUserMap[user.UserID]; ok {
			interestsRes[index].Type = 0
		} else {
			interestsRes[index].Type = 1
		}
		interestsRes[index].UserID = user.UserID
		interestsRes[index].Username = user.Nickname
		interestsRes[index].Interests = userInterestResMap[user.UserID]
	}

	log.Debug("", "interestsRes len: ", len(interestsRes))

	return interestsRes, count, nil
}

func GetUserInterestsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.UserInterests, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var interestsRes []*db.UserInterests
	var count int64
	if err != nil {
		return interestsRes, count, err
	}
	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "update_time"

	var userList []db.User
	var userIdList []string
	var userMap = make(map[string]*db.User, 0)

	// var defaultInterestsIdList []int64

	var interestIdList []int64

	var interestUserList []db.InterestUser
	var userInterestResMap map[string][]*db.InterestTypeRes

	defaultStatus := 0

	// if searching is default tags, then set the default status is 1 or 2, 2 is default, 1 is personal
	if isDefaultStr, ok := where["default"]; ok {
		if isDefaultStr != "" {
			defaultStatus, _ = strconv.Atoi(isDefaultStr)
		}
	}

	// Query user list
	queryUser := false
	dbSub := dbConn.Table("users")
	if account, ok := where["account"]; ok {
		if account != "" {
			queryUser = true
			userList = GetUserByAllCondition(account)
			for index, user := range userList {
				userIdList = append(userIdList, user.UserID)
				userMap[user.UserID] = &userList[index]
			}
		}
	}

	// Get default interests id list
	//if defaultStatus != 0 {
	//	dbConn.Table("interest_type").Debug().Select("id").
	//		Where("is_default=2").
	//		Where("delete_time=0").
	//		Order("id ASC").
	//		Find(&defaultInterestsIdList)
	//}

	// Query interests list by interest language
	queryInterestsList := false
	dbSub = dbConn.Table("interest_language").Select("interest_id").Order("interest_id ASC")
	if interestName, ok := where["interest_name"]; ok {
		if interestName != "" {
			queryInterestsList = true
			dbSub = dbSub.Where("name LIKE ?", "%"+interestName+"%")
		}
	}
	if queryInterestsList {
		dbSub.Find(&interestIdList)
	}

	// if there are interest id in default interests, need to search interest id equals -1 in interest user
	//hasDefault := false
	//defaultInterestLen := len(defaultInterestsIdList)
	//interestIdLen := len(interestIdList)
	//if interestIdLen > 0 && defaultInterestLen > 0 {
	//	interestIdIndex := 0
	//	defaultInterestIndex := 0
	//	for interestIdIndex < interestIdLen && defaultInterestIndex < defaultInterestLen {
	//		if interestIdList[interestIdIndex] > defaultInterestsIdList[defaultInterestIndex] {
	//			defaultInterestIndex++
	//		} else if interestIdList[interestIdIndex] < defaultInterestsIdList[defaultInterestIndex] {
	//			interestIdIndex++
	//		} else {
	//			hasDefault = true
	//			break
	//		}
	//	}
	//}

	// Query interests-user list's user id
	var limitUserId []string
	dbSub = dbConn.Table("interest_user").Distinct().Select("user_id")

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

	if queryUser {
		dbSub = dbSub.Where("user_id IN (?)", userIdList)
	}
	if queryInterestsList {
		dbSub = dbSub.Where("interest_id IN (?)", interestIdList)
	}

	// if is personal, query interest id isn't -1
	if defaultStatus == 1 {
		// not default
		queryInterestsList = true
		dbSub = dbSub.Where("interest_id != -1")
	} else if defaultStatus == 2 {
		// default
		queryInterestsList = true
		dbSub = dbSub.Where("interest_id = -1")
	}
	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}

	// limit user id
	dbSub.Debug().Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&limitUserId)

	// query result list
	dbSub = dbConn.Table("interest_user").Where("user_id IN (?)", limitUserId)

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
	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}

	// query
	dbSub.Debug().Find(&interestUserList)

	// if not query user list, then query users by interest-user's user id list
	if !queryUser {
		for _, userInt := range interestUserList {
			userIdList = append(userIdList, userInt.UserId)

		}
		dbConn.Table("users").Debug().Where("user_id IN (?)", userIdList).Find(&userList)
		for index, user := range userList {
			userMap[user.UserID] = &userList[index]
		}
	}

	// Query all interests
	var userInterestTypeMap = make(map[string]int32)
	var interestUserIdList []string
	for _, intUser := range interestUserList {
		if _, ok := userInterestTypeMap[intUser.UserId]; !ok {
			// didn't record this user, then add user id and set user interest type to 1 or 0
			interestUserIdList = append(interestUserIdList, intUser.UserId)
			userInterestTypeMap[intUser.UserId] = 1
			if intUser.InterestId == -1 {
				userInterestTypeMap[intUser.UserId] = 0
			}
		} else {
			// has record this user, and interest type hasn't default and the tag is default, then set default
			if userInterestTypeMap[intUser.UserId] != 0 && intUser.InterestId == -1 {
				userInterestTypeMap[intUser.UserId] = 0
			}
		}
	}

	log.Debug("xxxxxx")

	// get user's interests, include -1
	userInterestResMap = GetInterestsByUserIds(interestUserIdList)

	// Link result
	for _, userId := range interestUserIdList {
		interestRes := &db.UserInterests{}
		log.Debug("", "user id: ", userId)
		if interestId, ok := userInterestTypeMap[userId]; ok {
			interestRes.Type = interestId
		}
		interestRes.UserID = userId
		if user, ok := userMap[userId]; ok {
			interestRes.Username = user.Nickname
		} else {
			log.Debug("didn't find user map: ", userId)
		}

		if _, ok := userInterestResMap[userId]; ok {
			interestRes.Interests = userInterestResMap[userId]
		} else {
			log.Debug("didn't find user userInterestResMap: ", userId)
		}

		interestsRes = append(interestsRes, interestRes)
	}

	log.Debug("", "interestsRes len: ", len(interestsRes))

	return interestsRes, count, nil
}

func AlterUserInterests(userId string, interestList []int64) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	nowTime := time.Now().Unix()

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		//{
		//	i += tx.Table("interest_user").Create(&db.InterestUser{
		//		UserId:     userId,
		//		InterestId: -1,
		//		UpdateTime: nowTime,
		//	}).RowsAffected
		//} else
		_ = tx.Table("interest_user").Debug().Where("user_id=?", userId).Delete(&db.InterestUser{}).RowsAffected
		if len(interestList) != 0  {
			var interestUserList []db.InterestUser
			for _, interestId := range interestList {
				interestUserList = append(interestUserList, db.InterestUser{
					UserId:     userId,
					InterestId: interestId,
					UpdateTime: nowTime,
				})
			}
			i += tx.Table("interest_user").Debug().Create(&interestUserList).RowsAffected
		}
		return nil
	})
	if err != nil {
		return 0
	}

	go func() {
		_ = db.DB.DelUsersInfoByUserIdCache(userId)
		_ = db.DB.DeleteInterestGroupINfoListByUserId(userId)
	}()

	return i
}

func GetInterestGroupByInterestIdList(userId string, interestIdList []int64) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}
	month := time.Now().Format("0601")
	if time.Now().Day() < 2 {
		month = time.Now().AddDate(0, 0, -3).Format("0601")
	}

	excludeList := GetInterestGroupExcludeListByUserId(userId)
	joinedGroupList, _ := GetMyJoinedGroupIdByUserID(userId)

	// wait join group
	waitJoinGroupList, _ := GetUserReqGroupByUserID(userId)
	var waitJoinGroupMap []string
	if len(waitJoinGroupList) > 0 {
		for _, group := range waitJoinGroupList {
			if group.HandleResult == 0 {
				waitJoinGroupMap = append(waitJoinGroupMap, group.GroupID)
			}
		}
	}

	data := make([]string, 10)
	sql := dbConn.Table(db.InterestGroup{}.TableName()).Select("interest_group.group_id").
		Joins("left join group_heat on group_heat.group_id = interest_group.group_id").
		Joins("left join groups as ig on ig.group_id = interest_group.group_id").Where("ig.status = 0 and ig.is_open = 1").
		Where("interest_group.interest_id in (?)", interestIdList).Where("group_heat.month = ?", month)
	if len(excludeList) > 0 {
		sql = sql.Where("interest_group.group_id not in (?)", excludeList)
	}
	if len(joinedGroupList) > 0 {
		sql = sql.Where("group_heat.group_id not in (?)", joinedGroupList)
	}
	if len(waitJoinGroupMap) > 0 {
		sql = sql.Where("group_heat.group_id not in (?)", waitJoinGroupMap)
	}

	err = sql.Order("group_heat.heat desc").Limit(20).Pluck("group_id", &data).Error

	if err != nil || len(data) == 0 {
		sub := dbConn.Table(db.GroupHeat{}.TableName()).Select("group_heat.group_id").
			Where("group_heat.month = ?", time.Now().Format("0601")).
			Joins("left join groups as ig on ig.group_id = group_heat.group_id").Where("ig.status = 0 and ig.is_open = 1").
			Joins("left join interest_group on group_heat.group_id = interest_group.group_id").
			Where("interest_group.interest_id in (?)", interestIdList)

		if len(excludeList) > 0 {
			sub = sub.Where("group_heat.group_id not in (?)", excludeList)
		}
		if len(joinedGroupList) > 0 {
			sql = sql.Where("group_heat.group_id not in (?)", joinedGroupList)
		}
		if len(waitJoinGroupMap) > 0 {
			sql = sql.Where("group_heat.group_id not in (?)", waitJoinGroupMap)
		}
		sub.Order("group_heat.heat desc").Limit(20).Debug().Pluck("group_id", &data)
	}
	return data
}

func GetStringListByUserId(userId string) []string {
	userInterest, _ := GetUserInterestList(userId)

	//// get default interest type id list
	//if len(userInterest) == 0 || (len(userInterest) == 1 && userInterest[0] == constant.InterestDefault) {
	//	userInterest = GetDefaultInterestTypeList()
	//}

	var interestIdList []string
	for _, interestId := range userInterest {
		interestIdList = append(interestIdList, strconv.FormatInt(interestId, 10))
	}

	return interestIdList
}

func DeleteUserInterests(userId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	interestUser := db.InterestUser{}
	err = dbConn.Table(interestUser.TableName()).Where("user_id", userId).Delete(&interestUser).Error

	if err != nil {
		return err
	}
	return nil
}

