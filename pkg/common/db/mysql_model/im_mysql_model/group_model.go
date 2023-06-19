package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/utils"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

//type Group struct {
//	GroupID       string    `gorm:"column:group_id;primaryKey;"`
//	GroupName     string    `gorm:"column:name"`
//	Introduction  string    `gorm:"column:introduction"`
//	Notification  string    `gorm:"column:notification"`
//	FaceUrl       string    `gorm:"column:face_url"`
//	CreateTime    time.Time `gorm:"column:create_time"`
//	Status        int32     `gorm:"column:status"`
//	CreatorUserID string    `gorm:"column:creator_user_id"`
//	GroupType     int32     `gorm:"column:group_type"`
//	Ex            string    `gorm:"column:ex"`
//}

func InsertIntoGroup(groupInfo db.Group) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	if groupInfo.GroupName == "" {
		groupInfo.GroupName = "Group Chat"
	}
	groupInfo.CreateTime = time.Now()
	err = dbConn.Table("groups").Create(&groupInfo).Error
	if err != nil {
		return err
	}
	return nil
}

func GetGroupInfoByGroupID(groupId string) (*db.Group, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var groupInfo db.Group
	err = dbConn.Table("groups").Where("group_id=?", groupId).Take(&groupInfo).Error
	if err != nil {
		return nil, err
	}
	return &groupInfo, nil
}

func SetGroupInfo(groupInfo db.Group) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table("groups").Debug().Where("group_id=?", groupInfo.GroupID).Updates(&groupInfo).Error
	if err == nil {
		dbConn.Table("groups").Where("group_id=?", groupInfo.GroupID).Updates(map[string]interface{}{"notification": groupInfo.Notification})
	}

	return err
}

func GetGroupsByName(groupName string, pageNumber, showNumber int32) ([]db.Group, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var groups []db.Group
	if err != nil {
		return groups, err
	}

	err = dbConn.Table("groups").Where(fmt.Sprintf(" name like '%%%s%%' ", groupName)).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&groups).Error
	return groups, err
}

func GetGroupsByWholeName(groupName string, pageNumber, showNumber int32) ([]*db.Group, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var groups []*db.Group
	if err != nil {
		return groups, err
	}

	err = dbConn.Table("groups").Where(" name = ?", groupName).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&groups).Error
	return groups, err
}

func GetGroups(pageNumber, showNumber int, orderBy string) ([]db.Group, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var groups []db.Group
	if err != nil {
		return groups, err
	}

	if showNumber == -1 {
		if err = dbConn.Table("groups").Find(&groups).Error; err != nil {
			return groups, err
		}
		return groups, nil
	}
	sortMap := map[string]string{}
	var orderByClause string
	sortMap["create_time"] = "create_time"

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
		dbConn = dbConn.Order(orderByClause)
	}
	if err = dbConn.Table("groups").Limit(showNumber).Offset(showNumber * (pageNumber - 1)).Find(&groups).Error; err != nil {
		return groups, err
	}
	return groups, nil
}

func GetGroupsByWhere(where map[string]interface{}, statusList []string, pageNumber, showNumber int, orderBy string) ([]db.Group, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var groups []db.Group
	var count int64
	if err != nil {
		return groups, count, err
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["create_time"] = "create_time"

	var ownGroupList []string
	var creatorList []string

	queryOwner := false
	if ownerGet, ok := where["owner"]; ok {
		owner := ownerGet.(string)
		if owner != "" {
			var ownerList []string
			dbConn.Table("users").Select("user_id").
				Where("user_id like ? or name like ? or phone_number like ? or email like ?", "%"+owner+"%", "%"+owner+"%", "%"+owner+"%", "%"+owner+"%").
				Find(&ownerList)
			dbConn.Table("group_members").Select("group_id").
				Where("role_level=2").
				Where("user_id IN (?)", ownerList).
				Find(&ownGroupList)
			queryOwner = true
		}
	}

	queryCreator := false
	if creatorGet, ok := where["creator"]; ok {
		creator := creatorGet.(string)
		if creator != "" {
			dbConn.Table("users").Select("user_id").
				Where("user_id like ? or name like ? or phone_number like ? or email like ?", "%"+creator+"%", "%"+creator+"%", "%"+creator+"%", "%"+creator+"%").
				Find(&creatorList)
			queryCreator = true
		}
	}

	if queryOwner {
		dbConn = dbConn.Where("group_id IN (?)", ownGroupList)
	}

	if queryCreator {
		dbConn = dbConn.Where("creator_user_id IN (?)", creatorList)
	}

	if remarkWhere, ok := where["remark"]; ok {
		remark := remarkWhere.(string)
		if remark != "" {
			dbConn = dbConn.Where("remark LIKE ?", "%"+remark+"%")
		}
	}

	if startTimeWhere, ok := where["start_time"]; ok {
		startTime := startTimeWhere.(string)
		if startTime != "" {
			startTimeInt64, _ := strconv.ParseInt(startTime, 10, 64)
			dbConn = dbConn.Where("create_time>=?", time.Unix(startTimeInt64, 0))
		}
	}

	if endTimeWhere, ok := where["end_time"]; ok {
		endTime := endTimeWhere.(string)
		if endTime != "" {
			startTimeInt64, _ := strconv.ParseInt(endTime, 10, 64)
			dbConn = dbConn.Where("create_time<=?", time.Unix(startTimeInt64, 0))
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
		dbConn = dbConn.Order(orderByClause)
	}
	if isOpenGet, ok := where["is_open"]; ok {
		isOpen := isOpenGet.(int32)
		if isOpen != 0 {
			dbConn = dbConn.Where("is_open=?", isOpen-1)
		}
	}
	if statusList != nil {
		for _, status := range statusList {
			if status == "1" {
				dbConn = dbConn.Where("status=1")
			} else if status == "2" {
				dbConn = dbConn.Where("status=0")
			} else if status == "3" {
				dbConn = dbConn.Where("video_status=1")
			} else if status == "4" {
				dbConn = dbConn.Where("video_status=0")
			} else if status == "5" {
				dbConn = dbConn.Where("audio_status=1")
			} else if status == "6" {
				dbConn = dbConn.Where("audio_status=0")
			}
		}
	}

	if group, ok := where["group"]; ok {
		if group != "" {
			dbConn = dbConn.Where("group_id like ? or name like ?", "%"+group.(string)+"%", "%"+group.(string)+"%")
		}
	}
	if groupIdArray, ok := where["group_id_array"]; ok {
		if len(groupIdArray.([]string)) > 0 {
			dbConn = dbConn.Where("group_id in (?)", groupIdArray)
		}
	}

	if err = dbConn.Table("groups").Debug().Count(&count).Limit(showNumber).Offset(showNumber * (pageNumber - 1)).Find(&groups).Error; err != nil {
		return groups, count, err
	}
	return groups, count, nil
}

func OperateGroupStatus(groupId string, groupStatus int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	var updatedCount int64 = 0
	err = dbConn.Table("groups").Where("group_id=?", groupId).UpdateColumn("status", groupStatus).Count(&updatedCount).Error
	if updatedCount == 0 {
		return errors.New("group not found")
	}
	return err
}

func DeleteGroup(groupId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	var group db.Group
	var groupMembers []db.GroupMember
	if err := dbConn.Table("groups").Where("group_id=?", groupId).Delete(&group).Error; err != nil {
		return err
	}
	if err := dbConn.Table("group_members").Where("group_id=?", groupId).Delete(groupMembers).Error; err != nil {
		return err
	}
	return nil
}

func OperateGroupRole(userId, groupId string, roleLevel int32) (string, string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return "", "", err
	}
	groupMember := db.GroupMember{
		UserID:  userId,
		GroupID: groupId,
	}
	updateInfo := db.GroupMember{
		RoleLevel: roleLevel,
	}
	groupMaster := db.GroupMember{}
	switch roleLevel {
	case constant.GroupOwner:
		err = dbConn.Transaction(func(tx *gorm.DB) error {
			result := dbConn.Table("group_members").Where("group_id = ? and role_level = ?", groupId, constant.GroupOwner).First(&groupMaster).Updates(&db.GroupMember{
				RoleLevel: constant.GroupOrdinaryUsers,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New(fmt.Sprintf("user %s not exist in group %s or already operate", userId, groupId))
			}

			result = dbConn.Table("group_members").First(&groupMember).Updates(updateInfo)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New(fmt.Sprintf("user %s not exist in group %s or already operate", userId, groupId))
			}
			return nil
		})

	case constant.GroupOrdinaryUsers:
		err = dbConn.Transaction(func(tx *gorm.DB) error {
			result := dbConn.Table("group_members").Where("group_id = ? and role_level = ?", groupId, constant.GroupOwner).First(&groupMaster)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New(fmt.Sprintf("user %s not exist in group %s or already operate", userId, groupId))
			}
			if groupMaster.UserID == userId {
				return errors.New(fmt.Sprintf("user %s is master of %s, cant set to ordinary user", userId, groupId))
			} else {
				result = dbConn.Table("group_members").Find(&groupMember).Updates(updateInfo)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return errors.New(fmt.Sprintf("user %s not exist in group %s or already operate", userId, groupId))
				}
			}
			return nil
		})
	}
	return "", "", nil
}

func GetGroupsCountNum(group db.Group) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := dbConn.Table("groups").Where(fmt.Sprintf(" name like '%%%s%%' ", group.GroupName)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func GetGroupsCountNumByWhere(where map[string]interface{}, statusList []string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64

	dbSub := dbConn.Table("groups")
	if group, ok := where["group"]; ok {
		if group != "" {
			dbSub = dbSub.Where("group_id like ? or name like ?", "%"+group.(string)+"%", "%"+group.(string)+"%")
		}
	}
	if groupIdArray, ok := where["group_id_array"]; ok {
		if len(groupIdArray.([]string)) > 0 {
			dbSub = dbSub.Where("group_id in (?)", groupIdArray)
		}
	}
	err = dbSub.Count(&count).Error

	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetGroupById(groupId string) (db.Group, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	group := db.Group{
		GroupID: groupId,
	}
	if err != nil {
		return group, err
	}

	if err := dbConn.Table("groups").Find(&group).Error; err != nil {
		return group, err
	}
	return group, nil
}

func GetGroupMaster(groupId string) (db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	groupMember := db.GroupMember{}
	if err != nil {
		return groupMember, err
	}

	if err := dbConn.Table("group_members").Where("role_level=? and group_id=?", constant.GroupOwner, groupId).Find(&groupMember).Error; err != nil {
		return groupMember, err
	}
	return groupMember, nil
}

func UpdateGroupInfoDefaultZero(groupID string, args map[string]interface{}) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table("groups").Where("group_id = ? ", groupID).Updates(args).Error
}

func GetAllGroupIDList() ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupIDList []string
	err = dbConn.Table("groups").Pluck("group_id", &groupIDList).Error
	return groupIDList, err
}

func GetGroupByIDList(groupID []string) (map[string]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()

	type GroupList struct {
		GroupID string
		Name    string
	}
	groupList := []GroupList{}
	groupMap := map[string]string{}
	err = dbConn.Table("groups").Select("group_id", "name").Where("group_id IN ?", groupID).Find(&groupList).Error
	if err != nil {
		return groupMap, err
	}
	for _, v := range groupList {
		if v.GroupID != "" {
			groupMap[v.GroupID] = v.Name
		}
	}
	return groupMap, nil
}

func GetGroupInfoByGroupIDList(groupID []string) ([]db.Group, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var groupList []db.Group
	err = dbConn.Table("groups").Where("group_id IN ?", groupID).Find(&groupList).Error
	return groupList, err
}

func GetValidGroupIdListByGroupIdList(idList []string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var groupIDList []string
	dbConn.Table("groups").Where("group_id in ?", idList).Pluck("group_id", &groupIDList)
	return groupIDList
}

func GetSomeGroupNameByGroupId(groupIdList []string) map[string]string {
	dbConn, _ := db.DB.MysqlDB.DefaultGormDB()
	var group []db.Group
	dbConn.Table("groups").Where("group_id in ?", groupIdList).Select([]string{"group_id", "name"}).Find(&group)

	userNameMap := make(map[string]string, len(group))
	for _, v := range group {
		userNameMap[v.GroupID] = v.GroupName
	}
	return userNameMap
}

func GetGroupIdByCondition(where map[string]string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var groupIds []string

	dbSub := dbConn.Table("groups")
	if groupId, ok := where["group_id"]; ok {
		if groupId != "" {
			dbSub = dbSub.Where("group_id like ? or name like ?", "%"+groupId+"%", "%"+groupId+"%")
		}
	}
	dbSub.Pluck("group_id", &groupIds)

	return groupIds
}
