package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"strings"
	"time"

	"gorm.io/gorm"
)

//type GroupMember struct {
//	GroupID            string    `gorm:"column:group_id;primaryKey;"`
//	UserID             string    `gorm:"column:user_id;primaryKey;"`
//	NickName           string    `gorm:"column:nickname"`
//	FaceUrl            string    `gorm:"user_group_face_url"`
//	RoleLevel int32     `gorm:"column:role_level"`
//	JoinTime           time.Time `gorm:"column:join_time"`
//	JoinSource int32 `gorm:"column:join_source"`
//	OperatorUserID  string `gorm:"column:operator_user_id"`
//	Ex string `gorm:"column:ex"`
//}

func InsertIntoGroupMember(toInsertInfo db.GroupMember) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	toInsertInfo.JoinTime = time.Now()
	if toInsertInfo.RoleLevel == 0 {
		toInsertInfo.RoleLevel = constant.GroupOrdinaryUsers
	}
	toInsertInfo.MuteEndTime = time.Unix(int64(time.Now().Second()), 0)
	toInsertInfo.DeleteTime = 0
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(toInsertInfo.GroupID)
	toInsertInfo.UpdateVersion = int32(updatedVersion.VersionNumber + 1)
	trx := dbConn.Table("group_members").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "user_id"}},                                                                                                                                                    // key colume
		DoUpdates: clause.AssignmentColumns([]string{"nickname", "user_group_face_url", "role_level", "join_time", "join_source", "operator_user_id", "mute_end_time", "ex", "delete_time", "deleted_by", "update_version"}), // column needed to be updated
	}).Create(&toInsertInfo)
	if trx != nil && trx.Error != nil {
		return trx.Error
	}
	db.DB.SaveGroupIDForUser(toInsertInfo.UserID, toInsertInfo.GroupID)
	//Update Group Version Update - We check that for sync process
	go UpdateGroupUpdatesVersionNumber(toInsertInfo.GroupID)
	return nil
}

func GetGroupMemberListByUserID(userID string) ([]db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []db.GroupMember
	err = dbConn.Table("group_members").Where("user_id=? and delete_time = 0", userID).Find(&groupMemberList).Error
	//err = dbConn.Table("group_members").Where("user_id=?", userID).Take(&groupMemberList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func GetMyJoinedGroupIdByUserID(userID string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []string
	err = dbConn.Table("group_members").Where("user_id=? and delete_time = 0", userID).Pluck("group_id", &groupMemberList).Error
	//err = dbConn.Table("group_members").Where("user_id=?", userID).Take(&groupMemberList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func GetGroupMemberListByGroupID(groupID string) ([]db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []db.GroupMember
	err = dbConn.Table("group_members").Where("group_id=? and delete_time = 0", groupID).Find(&groupMemberList).Error

	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func GetGroupMemberIDListByGroupID(groupID string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var groupMemberIDList []string
	err = dbConn.Table("group_members").Where("group_id=? and delete_time = 0", groupID).Pluck("user_id", &groupMemberIDList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberIDList, nil
}

func GetGroupOwnerAdminIDListByGroupID(groupID string) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var groupMemberIDList []string
	err = dbConn.Table("group_members").Where("group_id=? and role_level > 1 and delete_time = 0", groupID).Pluck("user_id", &groupMemberIDList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberIDList, nil
}

func GetGroupMemberListByGroupIDAndRoleLevel(groupID string, roleLevel int32) ([]db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []db.GroupMember
	err = dbConn.Table("group_members").Where("group_id=? and role_level=? and delete_time = 0", groupID, roleLevel).Find(&groupMemberList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func GetGroupMemberInfoByGroupIDAndUserID(groupID, userID string) (*db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMember db.GroupMember
	err = dbConn.Table("group_members").Where("group_id=? and user_id=? and delete_time = 0", groupID, userID).Limit(1).Take(&groupMember).Error
	if err != nil {
		return nil, err
	}
	return &groupMember, nil
}

// GetGroupMemberInfoByGroupIDAndUserIDForSync -> ForSync we also fetched deleted users
func GetGroupMemberInfoByGroupIDAndUserIDForSync(groupID, userID string) (*db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMember db.GroupMember
	query := "group_id = '" + groupID + "' and user_id = '" + userID + "'"
	log.Error("GetGroupMemberInfoByGroupIDAndUserIDForSync Query ", query)
	err = dbConn.Table("group_members").Where(query).Find(&groupMember).Error
	if err != nil {
		return nil, err
	}
	return &groupMember, nil
}

// -> ForSync we also fetched deleted users
func GetGroupMemberInfoByGroupIDPagingForSync(groupID string, pageNUmber, pageSize int32) ([]db.GroupMember, int32, error) {
	var totalCount int64 = 0
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, int32(totalCount), err
	}
	var groupMembers []db.GroupMember

	query := "group_id = '" + groupID + "'"
	log.Error("GetGroupMemberInfoByGroupIDPagingForSync Query ", query)
	err = dbConn.Table("group_members").Where(query).Count(&totalCount).Order("role_level desc").Limit(int(pageSize)).Offset(int(pageNUmber-1) * int(pageSize)).Find(&groupMembers).Error
	if err != nil {
		return nil, int32(totalCount), err
	}
	return groupMembers, int32(totalCount), nil
}

func GetGroupMemberListByGroupIDAndRoleLevelV2(groupID string, roleLevel int32, nameSearch string, limit int32, offset int32) ([]db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []db.GroupMember
	err = dbConn.Table("group_members").Where("group_id = ? And role_level > ? And nickname like ? and delete_time = 0", groupID, roleLevel, "%"+nameSearch+"%").Order("join_time ASC").Offset(int(offset)).Limit(int(limit)).Find(&groupMemberList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func GetGroupMemberListByGroupIDV2(groupID string, nameSearch string, limit int32, offset int32) ([]db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []db.GroupMember
	err = dbConn.Table("group_members").Where("group_id=? and nickname like ? and delete_time = 0", groupID, "%"+nameSearch+"%").Offset(int(offset)).Limit(int(limit)).Find(&groupMemberList).Error

	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func DeleteGroupMemberByGroupIDAndUserID(groupID, userID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupID)
	updateInterface := make(map[string]interface{})
	updateInterface["delete_time"] = time.Now().Unix()
	updateInterface["update_version"] = updatedVersion.VersionNumber + 1
	err = dbConn.Table("group_members").Where("group_id=? and user_id=? ", groupID, userID).UpdateColumns(updateInterface).Error
	if err != nil {
		return err
	}
	db.DB.RemoveGroupIDForUser(userID, groupID)
	//Update Group Version Update - We check that for sync process
	go UpdateGroupUpdatesVersionNumber(groupID)
	return nil
}

func DeleteGroupMemberByGroupID(groupID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	userIDList, err := GetGroupMemberIDListByGroupID(groupID)
	if err != nil {
		return err
	}

	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupID)
	updateInterface := make(map[string]interface{})
	updateInterface["delete_time"] = time.Now().Unix()
	updateInterface["update_version"] = updatedVersion.VersionNumber + 1

	err = dbConn.Table("group_members").Where("group_id=?  ", groupID).UpdateColumns(updateInterface).Error
	if err != nil {
		return err
	}
	for _, s := range userIDList {
		db.DB.RemoveGroupIDForUser(s, groupID)
	}
	//Update Group Version Update - We check that for sync process
	go UpdateGroupUpdatesVersionNumber(groupID)
	return nil
}

func UpdateGroupMemberInfo(groupMemberInfo db.GroupMember) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var updateCount int64 = 0
	if err != nil {
		return err
	}
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupMemberInfo.GroupID)
	groupMemberInfo.UpdateVersion = int32(updatedVersion.VersionNumber + 1)

	err = dbConn.Table("group_members").Where("group_id=? and user_id=?", groupMemberInfo.GroupID, groupMemberInfo.UserID).Updates(&groupMemberInfo).Count(&updateCount).Error
	if err != nil {
		return err
	}
	//Update Group Version Update - We check that for sync process
	if updateCount == 0 {
		return errors.New("record not found")
	}
	go UpdateGroupUpdatesVersionNumber(groupMemberInfo.GroupID)
	return nil
}

func TransferGroupOwner(groupID, oldOwnerID, newOwnerID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	member := db.GroupMember{}
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupID)
	updateInterface := make(map[string]interface{})
	updateInterface["update_version"] = updatedVersion.VersionNumber + 1

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		//For Old User
		updateInterface["role_level"] = constant.GroupOrdinaryUsers
		err = tx.Model(&member).Where("group_id=? and user_id=?", groupID, oldOwnerID).UpdateColumns(updateInterface).Error
		if err != nil {
			log.NewError("", utils.GetSelfFuncName(), "Transaction oldOwnerID failed", err.Error())
			return err
		}
		//New User
		updateInterface["role_level"] = constant.GroupOwner
		err = tx.Model(&member).Where("group_id=? and user_id=?", groupID, newOwnerID).UpdateColumns(updateInterface).Error
		if err != nil {
			log.NewError("", utils.GetSelfFuncName(), "Transaction newOwnerID failed", err.Error())
			return err
		}
		return nil
	})

	if err == nil {
		//Update Group Version Update - We check that for sync process
		go UpdateGroupUpdatesVersionNumber(groupID)
	}
	return err
}

func TransferGroupAdmin(groupID, userID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	member := db.GroupMember{}
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupID)
	updateInterface := make(map[string]interface{})
	updateInterface["update_version"] = updatedVersion.VersionNumber + 1
	updateInterface["role_level"] = constant.GroupAdmin
	err = dbConn.Model(&member).Where("group_id=? and user_id=?", groupID, userID).UpdateColumns(updateInterface).Error

	if err == nil {
		//Update Group Version Update - We check that for sync process
		go UpdateGroupUpdatesVersionNumber(groupID)
	}
	return err
}

func TransferGroupOrdinary(groupID, userID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	member := db.GroupMember{}
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupID)
	updateInterface := make(map[string]interface{})
	updateInterface["update_version"] = updatedVersion.VersionNumber + 1
	updateInterface["role_level"] = constant.GroupOrdinaryUsers
	err = dbConn.Model(&member).Where("group_id=? and user_id=?", groupID, userID).UpdateColumns(updateInterface).Error

	if err == nil {
		//Update Group Version Update - We check that for sync process
		go UpdateGroupUpdatesVersionNumber(groupID)
	}
	return err
}

func UpdateGroupMemberInfoByMap(groupMemberInfo db.GroupMember, m map[string]interface{}) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupMemberInfo.GroupID)
	m["update_version"] = updatedVersion.VersionNumber + 1
	err = dbConn.Table("group_members").Where("group_id=? and user_id=?", groupMemberInfo.GroupID, groupMemberInfo.UserID).Updates(m).Error
	if err != nil {
		return err
	}
	//Update Group Version Update - We check that for sync process
	go UpdateGroupUpdatesVersionNumber(groupMemberInfo.GroupID)
	return nil
}

func GetOwnerManagerByGroupID(groupID string) ([]db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMemberList []db.GroupMember
	err = dbConn.Table("group_members").Where("group_id=? and role_level>? and delete_time = 0", groupID, constant.GroupOrdinaryUsers).Find(&groupMemberList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberList, nil
}

func GetGroupMemberNumByGroupID(groupID string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, utils.Wrap(err, "DefaultGormDB failed")
	}
	var number int64
	err = dbConn.Table("group_members").Where("group_id=? and delete_time = 0", groupID).Count(&number).Error
	if err != nil {
		return 0, utils.Wrap(err, "")
	}
	return number, nil
}

func GetGroupOwnerInfoByGroupID(groupID string) (*db.GroupMember, error) {
	//omList, err := GetOwnerManagerByGroupID(groupID)
	//if err != nil {
	//	return nil, err
	//}
	//for _, v := range omList {
	//	if v.RoleLevel == constant.GroupOwner {
	//		return &v, nil
	//	}
	//}

	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, utils.Wrap(err, "DefaultGormDB failed")
	}
	var ownerInfo = db.GroupMember{}
	err = dbConn.Table("group_members").Where("group_id=? and role_level=2 and delete_time = 0", groupID).Limit(1).Take(&ownerInfo).Error
	if err != nil {
		return nil, err
	}
	return &ownerInfo, nil
}

func GetOwnerNumByGroupID(groupID string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, utils.Wrap(err, "DefaultGormDB failed")
	}
	var number int64
	err = dbConn.Table("group_members").Where("group_id=? and role_level=2 and delete_time = 0", groupID).Count(&number).Error
	if err != nil {
		return 0, utils.Wrap(err, "")
	}
	return number, nil
}

func IsExistGroupMember(groupID, userID string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}
	var number int64
	err = dbConn.Table("group_members").Where("group_id = ? and user_id = ? and delete_time = 0", groupID, userID).Count(&number).Error
	if err != nil {
		return false
	}
	if number != 1 {
		return false
	}
	return true
}

func RemoveGroupMember(groupID string, UserID string) error {
	db.DB.RemoveGroupIDForUser(UserID, groupID)
	err := DeleteGroupMemberByGroupIDAndUserID(groupID, UserID)
	if err == nil {
		//Update Group Version Update - We check that for sync process
		go UpdateGroupUpdatesVersionNumber(groupID)
	}
	return err

}

func GetMemberInfoByID(groupID string, userID string) (*db.GroupMember, error) {
	return GetGroupMemberInfoByGroupIDAndUserID(groupID, userID)
}

func GetGroupMemberByGroupID(groupID string, filter int32, begin int32, maxNumber int32) ([]db.GroupMember, error) {
	var memberList []db.GroupMember
	var err error
	if filter >= 0 {
		memberList, err = GetGroupMemberListByGroupIDAndRoleLevel(groupID, filter) //sorted by join time
	} else {
		memberList, err = GetGroupMemberListByGroupID(groupID)
	}

	if err != nil {
		return nil, err
	}
	if begin >= int32(len(memberList)) {
		return nil, nil
	}

	var end int32
	if begin+int32(maxNumber) < int32(len(memberList)) {
		end = begin + maxNumber
	} else {
		end = int32(len(memberList))
	}
	return memberList[begin:end], nil
}

func GetGroupMemberByGroupIDV2(groupID string, filter int32, searchName string, limit int32, offset int32) ([]db.GroupMember, error) {
	var memberList []db.GroupMember
	var err error
	if filter >= 0 {
		memberList, err = GetGroupMemberListByGroupIDAndRoleLevelV2(groupID, filter, searchName, limit, offset) //sorted by join time
	} else {
		memberList, err = GetGroupMemberListByGroupIDV2(groupID, searchName, limit, offset)
	}
	return memberList, err
}

func GetJoinedGroupIDListByUserID(userID string) ([]string, error) {
	memberList, err := GetGroupMemberListByUserID(userID)
	if err != nil {
		return nil, err
	}

	groupIDList := []string{}
	for _, v := range memberList {
		if v.GroupID != "" {
			groupIDList = append(groupIDList, v.GroupID)
		}
	}
	return groupIDList, nil
}

func IsGroupOwnerAdmin(groupID, UserID string) bool {
	groupMemberList, err := GetOwnerManagerByGroupID(groupID)
	if err != nil {
		return false
	}
	for _, v := range groupMemberList {
		if v.UserID == UserID && v.RoleLevel > constant.GroupOrdinaryUsers {
			return true
		}
	}
	return false
}

func GetGroupMembersByGroupIdCMS(groupId string, userName string, showNumber, pageNumber int32) ([]db.GroupMember, int64, error) {
	var groupMembersCount int64
	var groupMembers []db.GroupMember
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return groupMembers, groupMembersCount, err
	}

	sub := dbConn.Table("group_members").Where("group_members.group_id=? and group_members.delete_time = 0", groupId)
	if userName != "" {
		// .Where(fmt.Sprintf(" nickname like '%%%s%%' ", userName))
		or := fmt.Sprintf(" (users.name like '%%%s%%' or group_members.nickname like '%%%s%%' or group_members.user_id like '%%%s%%' or users.phone_number like '%%%s%%') ", userName, userName, userName, userName)

		lowerUserName := strings.ToLower(userName)
		if lowerUserName == "群成员" || lowerUserName == "member" {
			or = or + fmt.Sprintf(" or group_members.role_level = %d", 1)
		}
		if lowerUserName == "群主" || lowerUserName == "owner" {
			or = or + fmt.Sprintf(" or group_members.role_level = %d", 2)
		}
		if lowerUserName == "管理员" || lowerUserName == "admin" {
			or = or + fmt.Sprintf(" or group_members.role_level = %d", 3)
		}

		sub = sub.Joins(" left join users on users.user_id = group_members.user_id").Where(or)
	}

	err = sub.Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Count(&groupMembersCount).Find(&groupMembers).Error
	if err != nil {
		return nil, groupMembersCount, err
	}
	return groupMembers, groupMembersCount, nil
}

func GetNoMembers(groupId string, showNumber, pageNumber int32) ([]db.User, int64, error) {
	var membersCount int64
	var users []db.User
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return users, membersCount, err
	}

	memberUsers := []string{}

	dbConn.Table("group_members").Select("user_id").Debug().Where("group_members.group_id!=? and group_members.delete_time = 0", groupId).Find(&memberUsers)

	sub := dbConn.Table("users").Debug().Where("user_id NOT IN (?) and delete_time = 0", memberUsers)

	err = sub.Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Count(&membersCount).Find(&users).Error
	if err != nil {
		return nil, membersCount, err
	}
	return users, membersCount, nil
}

func GetGroupMemberByWhere(groupId string, where map[string]interface{}, showNumber, pageNumber int32) ([]db.GroupMember, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var count int64
	if err != nil {
		return nil, count, err
	}

	var members []db.User
	var memberIDs []string
	var memberMap map[string]*db.User

	var result []db.GroupMember

	// query member
	queryMember := false
	if memberWhere, ok := where["member"]; ok {
		member := memberWhere.(string)
		if member != "" {
			queryMember = true
			members = GetUserByAllCondition(member)
			memberIDs = make([]string, len(members))
			memberMap = make(map[string]*db.User, len(members))
			for index, user := range members {
				memberIDs[index] = user.UserID
				memberMap[user.UserID] = &members[index]
			}
		}
	}

	// query group member
	dbSub := dbConn.Table("group_members").Where("delete_time=0")
	if queryMember {
		dbSub = dbSub.Where("user_id IN (?)", memberIDs)
	}
	dbSub = dbSub.Where("group_id = ?", groupId)
	if remarkNameWhere, ok := where["remark_name"]; ok {
		remark := remarkNameWhere.(string)
		if remark != "" {
			dbSub = dbSub.Where("nickname LIKE ?", "%"+remark+"%")
		}
	}
	if remarkWhere, ok := where["remark"]; ok {
		remark := remarkWhere.(string)
		if remark != "" {
			dbSub = dbSub.Where("remark LIKE ?", "%"+remark+"%")
		}
	}
	if roleLevelWhere, ok := where["role_level"]; ok {
		roleLevel := roleLevelWhere.([]int32)
		if len(roleLevel) > 0 {
			dbSub = dbSub.Where("role_level IN (?)", roleLevel)
		}
	}
	if startTimeWhere, ok := where["start_time"]; ok {
		startTime := startTimeWhere.(string)
		if startTime != "" {
			dbSub = dbSub.Where("join_time >= ?", startTime)
		}
	}
	if endTimeWhere, ok := where["end_time"]; ok {
		endTime := endTimeWhere.(string)
		if endTime != "" {
			dbSub = dbSub.Where("join_time <= ?", endTime)
		}
	}
	if permissionWhere, ok := where["permission"]; ok {
		permission := permissionWhere.([]int32)
		if len(permission) > 0 {
			var roleLevelList []int32
			for _, p := range permission {
				if p == 7 {
					roleLevelList = append(roleLevelList, constant.GroupOwner)
				} else if p == 5 {
					roleLevelList = append(roleLevelList, constant.GroupOwner, constant.GroupAdmin)
				} else if p == 3 {
					roleLevelList = append(roleLevelList, constant.GroupOrdinaryUsers, constant.GroupOwner, constant.GroupAdmin)
				}
			}
			dbSub = dbSub.Where("role_level IN (?)", roleLevelList)
		}
	}
	if statusWhere, ok := where["status"]; ok {
		status := statusWhere.([]int32)
		if len(status) > 0 {
			sql := ""
			isFirst := true
			for _, s := range status {
				if s >= 1 && s <= 4 {
					// TODO private or group chatting status
				} else {
					if s == 5 {
						if isFirst {
							isFirst = false
							sql += " audio_status = 1"
						} else {
							sql += " OR audio_status = 1"
						}
					} else if s == 6 {
						if isFirst {
							isFirst = false
							sql += " audio_status = 2"
						} else {
							sql += " OR audio_status = 2"
						}
					} else if s == 7 {
						if isFirst {
							isFirst = false
							sql += " video_status = 1"
						} else {
							sql += " OR video_status = 1"
						}
					} else if s == 8 {
						if isFirst {
							isFirst = false
							sql += " video_status = 2"
						} else {
							sql += " OR video_status = 2"
						}
					}
				}
			}
			dbSub = dbSub.Where(sql)
		}
	}
	dbSub = dbSub.Order(fmt.Sprintf("role_level=%d DESC, join_time ASC", constant.GroupOwner))
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Debug().Find(&result)

	return result, count, nil

}

func GetGroupMembersCount(groupId, userName string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var count int64
	if err != nil {
		return count, err
	}

	sub := dbConn.Table("group_members").Where("group_id=? and delete_time = 0", groupId)
	if userName != "" {
		// .Where(fmt.Sprintf(" nickname like '%%%s%%' ", userName))
		sub = sub.Joins("left join users on users.user_id = group_members.user_id").
			Where("users.name like ? OR group_members.nickname like ? or group_members.user_id like ? or users.phone_number like ?", "%"+userName+"%", "%"+userName+"%", "%"+userName+"%", "%"+userName+"%")
	}
	err = sub.Count(&count).Error

	if err != nil {
		return count, err
	}
	return count, nil
}

func UpdateGroupMemberInfoDefaultZero(groupMemberInfo db.GroupMember, args map[string]interface{}) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	updatedVersion, _ := GetGroupUpdatesVersionNumberGroupID(groupMemberInfo.GroupID)
	args["update_version"] = updatedVersion.VersionNumber + 1
	err = dbConn.Model(groupMemberInfo).Updates(args).Error
	if err == nil {
		//Update Group Version Update - We check that for sync process
		go UpdateGroupUpdatesVersionNumber(groupMemberInfo.GroupID)
	}
	return err
}

func GetGroupMemberByUserIDGroupID(groupId, userID string) (*db.GroupMember, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var groupMember *db.GroupMember
	err = dbConn.Table("group_members").Where("group_id=? and user_id=? and delete_time = 0", groupId, userID).Find(&groupMember).Error
	//err = dbConn.Table("group_members").Where("user_id=?", userID).Take(&groupMemberList).Error
	if err != nil {
		return nil, err
	}
	return groupMember, nil
}

func GetGroupIdByMemberIds(memberIdList []string) []string {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return []string{}
	}

	var groupIDList []string
	dbConn.Table("group_members").Where("user_id in ? and delete_time = 0", memberIdList).Group("group_id").Pluck("group_id", &groupIDList)
	return groupIDList
}

func UpdateGroupUpdatesVersionNumber(groupID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		var gVersion db.GroupUpdatesVersion
		if err := tx.Where("group_id = ?", groupID).First(&gVersion).Error; err != nil {
			log.NewError("UpdateGroupUpdatesVersionNumber", utils.GetSelfFuncName(), "Find ", err.Error())
			gVersion.GroupID = groupID
			gVersion.VersionNumber = 1
			gVersion.UpdateTime = time.Now()
			err = tx.Create(gVersion).Error
			if err != nil {
				log.NewError("UpdateGroupUpdatesVersionNumber", utils.GetSelfFuncName(), "Insert after Find ", err.Error())
			}
			return err
		}
		if gVersion.VersionNumber == 0 {
			gVersion.GroupID = groupID
			gVersion.VersionNumber = 1
			gVersion.UpdateTime = time.Now()
			err = tx.Create(gVersion).Error
			if err != nil {
				log.NewError("UpdateGroupUpdatesVersionNumber", utils.GetSelfFuncName(), "Insert after version code zero ", err.Error())
			}
			return err
		} else if err := tx.Model(&gVersion).Update("version_number", gVersion.VersionNumber+1).Error; err != nil {
			log.NewError("UpdateGroupUpdatesVersionNumber", utils.GetSelfFuncName(), "Update ", err.Error())
			return err
		}
		return nil
	})
	return err
}

func GetGroupUpdatesVersionNumber(groupIDs []string) ([]db.GroupUpdatesVersion, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var gVersions []db.GroupUpdatesVersion
	if err != nil {
		return gVersions, err
	}
	err = dbConn.Table(db.GroupUpdatesVersion{}.TableName()).Where("group_id in (" + strings.Join(groupIDs, ",") + ")").Find(&gVersions).Error
	return gVersions, err
}

func GetGroupUpdatesVersionNumberGroupID(groupID string) (db.GroupUpdatesVersion, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var gVersions = db.GroupUpdatesVersion{}
	if err != nil {
		return gVersions, err
	}
	err = dbConn.Table(db.GroupUpdatesVersion{}.TableName()).Where("group_id = ? ", groupID).Find(&gVersions).Error
	return gVersions, err
}

// GetGroupMemberIDListByVersionGrater -> ForSync we also fetch user deleted status
func GetGroupMemberIDListByVersionGrater(groupID string, updateVersion int64) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var groupMemberIDList []string
	query := "group_id=" + groupID + " and update_version > " + fmt.Sprintf("%d", updateVersion)
	log.Error("GetGroupMemberIDListByVersionGrater Query", query)
	err = dbConn.Table("group_members").Where(query).Pluck("user_id", &groupMemberIDList).Error
	if err != nil {
		return nil, err
	}
	return groupMemberIDList, nil
}

func UpdateGroupStatus(g *db.Group) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	row := dbConn.Table(g.TableName()).Where("group_id=?", g.GroupID).Updates(g).RowsAffected
	return row
}

func UpdateGroupMemberStatus(g *db.GroupMember) int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	row := dbConn.Table(g.TableName()).Where("group_id=? and user_id=?", g.GroupID, g.UserID).Updates(g).RowsAffected
	return row
}

//
//func SelectGroupList(groupID string) ([]string, error) {
//	var groupUserID string
//	var groupList []string
//	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
//	if err != nil {
//		return groupList, err
//	}
//
//	rows, err := dbConn.Model(&GroupMember{}).Where("group_id = ?", groupID).Select("user_id").Rows()
//	if err != nil {
//		return groupList, err
//	}
//	defer rows.Close()
//	for rows.Next() {
//		rows.Scan(&groupUserID)
//		groupList = append(groupList, groupUserID)
//	}
//	return groupList, nil
//}
