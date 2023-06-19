package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

func SetGroupInterestType(groupId string, interestType []int64) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	rwMutex.Lock()
	defer rwMutex.Unlock()

	dbConn.Where("group_id = ?", groupId).Delete(&db.InterestGroup{})

	if len(interestType) == 0 {
		return true
	}

	var data []db.InterestGroup
	now := time.Now().Unix()
	for _, s := range interestType {
		data = append(data, db.InterestGroup{GroupId: groupId, InterestId: int64(s), UpdateTime: now})
	}
	err = dbConn.Table(db.InterestGroup{}.TableName()).Create(&data).Error
	if err != nil {
		log.NewError("", "UpdateInterestGroup", "err", err, "data", data)
		return false
	}
	return true
}

func GetGroupInterestsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.GroupInterests, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var interestsRes []*db.GroupInterests
	var count int64
	if err != nil {
		return interestsRes, count, err
	}
	sortMap := map[string]string{}
	sortMap["start_time"] = "create_time"
	var orderByClause string

	var groupIdList []string
	var groupList []db.Group
	var groupMap = make(map[string]*db.Group)

	var interestIdList []int64
	var queryGroupList []string
	var interestGroup []db.InterestGroup
	var interestGroupMap map[string][]int64

	queryInterest := false
	if groupId, ok := where["group_id"]; ok && groupId != "" {
		dbConn.Table("groups").Where("status != 2").Where("is_open = 1").Where("group_id=?", groupId).
			Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&groupList)
	} else {
		// Query interests list by interest name.
		if interestName, ok := where["interest_name"]; ok {
			if interestName != "" {
				queryInterest = true
				dbConn.Table("interest_language").Select("interest_id").Where("name LIKE ?", "%"+interestName+"%").Find(&interestIdList)
			}
		}

		if queryInterest {
			// Query interest_group.
			dbSub := dbConn.Table("interest_group")
			if queryInterest {
				dbSub = dbSub.Where("interest_id IN (?)", interestIdList)
			}
			dbSub.Find(&interestGroup)
			queryGroupList = make([]string, len(interestGroup))
			interestGroupMap = make(map[string][]int64)
			for _, ig := range interestGroup {
				queryGroupList = append(queryGroupList, ig.GroupId)
				if _, ok := interestGroupMap[ig.GroupId]; !ok {
					interestGroupMap[ig.GroupId] = make([]int64, 0)
				}
				interestGroupMap[ig.GroupId] = append(interestGroupMap[ig.GroupId], ig.InterestId)
			}
		}

		// Query group list by group name.
		dbSub := dbConn.Table("groups").Where("status != 2").Where("is_open = 1")
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
		if queryInterest {
			dbSub = dbSub.Where("group_id IN (?)", queryGroupList)
		}
		if orderByClause != "" {
			log.Debug("inter", "DB orderByClause", orderByClause)
			dbSub = dbSub.Order(orderByClause)
		}
		if groupName, ok := where["group_name"]; ok {
			if groupName != "" {
				dbSub = dbSub.Where("name LIKE ? OR group_id LIKE ? ", "%"+groupName+"%", "%"+groupName+"%")
			}
		}
		dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&groupList)
	}

	for index, group := range groupList {
		groupIdList = append(groupIdList, group.GroupID)
		groupMap[group.GroupID] = &groupList[index]
	}

	// get all groups interest id list
	if !queryInterest {
		dbConn.Table("interest_group").Where("group_id IN (?)", groupIdList).Find(&interestGroup)
		interestIdList = make([]int64, len(interestGroup))
		interestGroupMap = make(map[string][]int64)
		for _, ig := range interestGroup {
			interestIdList = append(interestIdList, ig.InterestId)
			if _, ok := interestGroupMap[ig.GroupId]; !ok {
				interestGroupMap[ig.GroupId] = make([]int64, 0)
			}
			interestGroupMap[ig.GroupId] = append(interestGroupMap[ig.GroupId], ig.InterestId)
		}
	}

	interestTypeResList, err := GetInterestsById(interestIdList)
	if err != nil {
		return nil, count, err
	}

	var interestTypeMap = make(map[int64]*db.InterestTypeRes)
	for index, interestType := range interestTypeResList {
		interestTypeMap[interestType.Id] = &interestTypeResList[index]
	}

	// Link list and get result to return  var interestsRes []*db.GroupInteres
	interestsRes = make([]*db.GroupInterests, len(groupList))
	for index, group := range groupList {
		var value = &db.GroupInterests{}
		_ = utils.CopyStructFields(&value, group)
		value.GroupType = group.IsOpen

		// if don't have group record pointer, create a new record.
		if gInterList, ok := interestGroupMap[group.GroupID]; ok {
			for _, inter := range gInterList {
				if interestRes, ok := interestTypeMap[inter]; ok {
					value.Interests = append(value.Interests, *interestRes)
				}
			}
		}
		interestsRes[index] = value
	}

	return interestsRes, count, nil

}

func AlterGroupInterests(groupId string, interestList []int64) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	nowTime := time.Now().Unix()

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		tx.Table("interest_group").Debug().Where("group_id=?", groupId).Delete(&db.InterestGroup{})
		var interestUserList []db.InterestGroup
		for _, interestId := range interestList {
			interestUserList = append(interestUserList, db.InterestGroup{
				GroupId:    groupId,
				InterestId: interestId,
				UpdateTime: nowTime,
			})
		}
		i += tx.Table("interest_group").Debug().Create(&interestUserList).RowsAffected
		return nil
	})
	if err != nil {
		return 0
	}
	return i
}

func DeleteGroupInterests(groupId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	interestGroup := db.InterestGroup{}
	err = dbConn.Table(interestGroup.TableName()).Where("group_id", groupId).Delete(&interestGroup).Error

	if err != nil {
		return err
	}
	return nil
}

