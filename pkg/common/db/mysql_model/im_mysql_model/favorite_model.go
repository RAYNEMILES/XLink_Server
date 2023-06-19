package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func AddFavorites(favorites *db.FavoritesSQL, opUserId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	favorites.CreateTime = time.Now().Unix()
	favorites.CreateBy = opUserId

	err = dbConn.Table("favorites").Create(favorites).Error

	return err
}

func GetFavoritesByFavIds(favorites []string) ([]*db.FavoritesSQL, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var res []*db.FavoritesSQL
	dbConn.Table("favorites").Where("favorite_id IN (?)", favorites).Where("delete_time=0").Find(&res)

	return res, nil
}

func DeleteFavorites(favorites []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table("favorites").Where("favorite_id IN (?)", favorites).UpdateColumn("delete_time", time.Now().Unix()).Error
	return err
}

func UpdateFavorite(favorite *db.FavoritesSQL, opUserId string) error {
	favorite.UpdateTime = time.Now().Unix()
	favorite.UpdateBy = opUserId
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table(favorite.TableName()).Where("favorite_id=?", favorite.FavoriteId).Updates(favorite).Error
	return err
}

func GetFavoritesByWhere(where map[string]string, contentType []int32, pageNumber int32, showNumber int32, orderBy string) ([]*db.FavoriteRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var favorites []*db.FavoriteRes
	var count int64
	if err != nil {
		return favorites, count, err
	}
	timeTypeMap := map[int]string{}
	timeTypeMap[1] = "create_time"
	timeTypeMap[2] = "publish_time"

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["create_time"] = "create_time"

	var favoriteUser []db.User
	var favoriteUserMap = make(map[string]*db.User)
	var favoriteUserIdList []string

	var publishUser []db.User
	var publishUserMap = make(map[string]*db.User)
	var publishUserIdList []string

	var officialUser []*db.Official
	var officialUserMap = make(map[int64]*db.Official)
	var officialUserIdList []int64

	// favorite user
	queryFavoriteUser := false
	if account, ok := where["account"]; ok {
		if account != "" {
			queryFavoriteUser = true
			favoriteUser = GetUserByAllCondition(account)
		}
		for index, user := range favoriteUser {
			favoriteUserMap[user.UserID] = &favoriteUser[index]
			favoriteUserIdList = append(favoriteUserIdList, user.UserID)
		}
	}

	// publish user
	queryPublishUser := false
	if publishAccount, ok := where["publish_user"]; ok {
		if publishAccount != "" {
			queryPublishUser = true
			publishUser = GetUserByAllCondition(publishAccount)
			officialUser, _ = GetOfficialByUserName(publishAccount)
		}
		for index, user := range publishUser {
			publishUserMap[user.UserID] = &publishUser[index]
			publishUserIdList = append(publishUserIdList, user.UserID)
		}
		for index, official := range officialUser {
			officialUserMap[official.Id] = officialUser[index]
			officialUserIdList = append(officialUserIdList, official.Id)
		}
	}

	// publish official account

	// favorites
	dbSub := dbConn.Table(db.FavoritesSQL{}.TableName()).Where("delete_time=0")
	if queryFavoriteUser {
		dbSub = dbSub.Where("user_id IN (?)", favoriteUserIdList)
	}
	if queryPublishUser {
		dbSub = dbSub.Where("content_creator_id IN (?) OR content_creator_id IN (?)", publishUserIdList, officialUserIdList)
	}

	// sort content type
	typeLen := len(contentType)
	if typeLen == 1 && contentType[0] == 0 {
		// all favorites, no type, recently used
		t := time.Now().AddDate(0, 0, -6)
		sevenDaysAgo := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()
		dbSub = dbSub.Where("create_time>?", sevenDaysAgo)
	} else if typeLen != 0 {
		// all types checked
		dbSub = dbSub.Where("content_type IN (?)", contentType)
	}

	// add content searching conditions
	if content, ok := where["content"]; ok {
		if content != "" {
			//var sql = ""
			//var paramList []interface{}
			var likeStr = "%" + content + "%"
			newKey := strings.ReplaceAll(content, "-", "--")
			mergerLike := "%" + newKey + "%"

			dbSub = dbSub.Where("(source_type = ? AND ex_keywords LIKE ?) OR (source_type != ? AND ex_keywords LIKE ?)",
				constant.FavoriteSourceCombineChatting, mergerLike, constant.FavoriteSourceCombineChatting, likeStr)

			//setLike := false
			//for _, cType := range contentType {
			//	if cType == constant.FavoriteContentTypeChats {
			//		newKey := strings.ReplaceAll(content, "-", "--")
			//		mergerLike := "%" + newKey + "%"
			//		dbSub = dbSub.Where("(source_type = ? AND ex_keywords LIKE ?) OR (source_type != ? AND ex_keywords LIKE ?)",
			//			constant.FavoriteSourceCombineChatting, mergerLike, constant.FavoriteSourceCombineChatting, likeStr)
			//		setLike = true
			//		break
			//	}
			//}
			//for _, cType := range contentType {
			//	if cType == constant.FavoriteContentTypeMedia || cType == constant.FavoriteContentTypeAudio {
			//		dbSub = dbSub.Where("content_type != ?", cType)
			//	} else if !setLike {
			//		dbSub = dbSub.Where("ex_keywords LIKE ?", likeStr)
			//		setLike = true
			//	}
			//}
			//dbSub = dbSub.Where(sql, paramList...)
		}
	}

	startTime := where["start_time"]
	endTime := where["end_time"]
	timeTypeStr := where["time_type"]
	timeType, _ := strconv.Atoi(timeTypeStr)
	if timeType == 0 {
		if startTime != "" || endTime != "" {
			var sql = ""
			var paramList []interface{}
			isFirst := true
			for _, v := range timeTypeMap {
				beenOr := false
				if startTime != "" {
					if isFirst {
						sql += fmt.Sprintf("%s>=?", v)
						paramList = append(paramList, startTime)
					} else {
						sql += fmt.Sprintf(" OR %s>=?", v)
						paramList = append(paramList, startTime)
						beenOr = true
					}
				}
				if endTime != "" {
					if isFirst || beenOr {
						sql += fmt.Sprintf(" AND %s<=?", v)
						paramList = append(paramList, endTime)
					} else {
						sql += fmt.Sprintf(" OR %s<=?", v)
						paramList = append(paramList, endTime)
					}
				}
				isFirst = false
			}

			dbSub = dbSub.Where(sql, paramList...)
		}
	} else {
		if startTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s>=?", timeTypeMap[timeType]), startTime)
		}
		if endTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s<=?", timeTypeMap[timeType]), endTime)
		}
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
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}

	var favoriteSQL []db.FavoritesSQL
	dbSub.Debug().Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&favoriteSQL)

	if !queryFavoriteUser || !queryPublishUser {
		for _, fav := range favoriteSQL {
			if !queryFavoriteUser {
				favoriteUserIdList = append(favoriteUserIdList, fav.UserID)
			}
			if !queryPublishUser {
				if fav.SourceType == constant.FavoriteSourceTypeArticle {
					officialID, _ := strconv.ParseInt(fav.ContentCreatorID, 10, 64)
					officialUserIdList = append(officialUserIdList, officialID)
				} else {
					publishUserIdList = append(publishUserIdList, fav.ContentCreatorID)
				}
			}
		}
		if !queryFavoriteUser {
			dbConn.Table(db.User{}.TableName()).Debug().Where("user_id IN (?)", favoriteUserIdList).Find(&favoriteUser)
			for index, user := range favoriteUser {
				favoriteUserMap[user.UserID] = &favoriteUser[index]
			}
		}
		if !queryPublishUser {
			tempMap, _ := GetOfficialsByOfficialIDList(officialUserIdList)
			for k, v := range tempMap {
				official := db.Official{}
				_ = utils.CopyStructFields(&official, v)
				officialUserMap[k] = &official
			}
			dbConn.Table(db.User{}.TableName()).Debug().Where("user_id IN (?)", publishUserIdList).Find(&publishUser)
			for index, user := range publishUser {
				publishUserMap[user.UserID] = &publishUser[index]
			}
		}
	}

	// link favorites
	for _, favorite := range favoriteSQL {
		favRes := db.FavoriteRes{}
		err = utils.CopyStructFields(&favRes, favorite)
		if err != nil {
			return nil, count, err
		}
		if user, ok := favoriteUserMap[favorite.UserID]; ok {
			favRes.UserName = user.Nickname
		}
		if favorite.UpdateBy != "" {
			updateUser := &db.User{}
			updateUser, err = GetUserByUserID(favorite.UpdateBy)
			if err == nil {
				favRes.UpdateBy = updateUser.Nickname
			}
		}
		gGroup := &db.Group{}
		creatorInfoStr := struct {
			ContentGroupName   string
			ContentCreatorName string
		}{}

		if favorite.ContentCreatorID != "" {
			if favorite.SourceType == constant.FavoriteSourceTypeArticle {
				officialID, _ := strconv.ParseInt(favorite.ContentCreatorID, 10, 64)
				if official, ok := officialUserMap[officialID]; ok {
					creatorInfoStr.ContentCreatorName = official.Nickname
				}
			} else {
				if user, ok := publishUserMap[favorite.ContentCreatorID]; ok {
					creatorInfoStr.ContentCreatorName = user.Nickname
				}
			}
		}

		if favorite.ContentGroupID != "" {
			gGroup, _ = GetGroupInfoByGroupID(favorite.ContentGroupID)
			if gGroup != nil {
				creatorInfoStr.ContentGroupName = gGroup.GroupName
			}
		}

		contentCreatorInfo, err := json.Marshal(creatorInfoStr)
		if err != nil {
			return nil, count, err
		}
		favRes.ContentCreatorName = string(contentCreatorInfo)

		favorites = append(favorites, &favRes)
	}

	return favorites, count, nil
}
