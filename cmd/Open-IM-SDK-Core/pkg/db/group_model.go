package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/common/log"
	"errors"
	"fmt"
	"time"
)

func (d *DataBase) InsertGroup(groupInfo *model_struct.LocalGroup) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(groupInfo).Error, "InsertGroup failed")
}
func (d *DataBase) DeleteGroup(groupID string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	localGroup := model_struct.LocalGroup{GroupID: groupID}
	return utils.Wrap(d.conn.Delete(&localGroup).Error, "DeleteGroup failed")
}
func (d *DataBase) UpdateGroup(groupInfo *model_struct.LocalGroup) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()

	t := d.conn.Model(groupInfo).Select("*").Updates(*groupInfo)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "")

}

func (d *DataBase) UpdateUserNameInGroupLocalChatLogs(groupID, userID string) error {

	return utils.Wrap(d.conn.Table(utils.GetSuperGroupTableName(groupID)).Where(
		"send_id = ? ", userID).Updates(
		map[string]interface{}{"sender_face_url": "", "sender_nick_name": "Deleted Account", "account_status": 2}).Error, utils.GetSelfFuncName()+" failed")
}
func (d *DataBase) GetJoinedGroupList() ([]*model_struct.LocalGroup, error) {
	var groupList []model_struct.LocalGroup
	var sqlQry = "SELECT * FROM local_groups INNER JOIN local_group_members ON '" + d.loginUserID + "'=local_group_members.user_id AND local_groups.group_id = local_group_members.group_id ORDER BY local_group_members.join_time DESC;"
	log.Debug("Joined Groups", sqlQry)
	err := d.conn.Raw(sqlQry).Scan(&groupList).Error
	var transfer []*model_struct.LocalGroup
	for _, v := range groupList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, utils.Wrap(err, "GetJoinedGroupList failed ")
}

func (d *DataBase) GetJoinedGroupIDList() ([]string, error) {
	groupList, err := d.GetJoinedGroupList()
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var groupIDList []string
	for _, v := range groupList {
		groupIDList = append(groupIDList, v.GroupID)
	}
	return groupIDList, nil
}

func (d *DataBase) GetGroupInfoByGroupID(groupID string) (*model_struct.LocalGroup, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var g model_struct.LocalGroup
	return &g, utils.Wrap(d.conn.Where("group_id = ?", groupID).Take(&g).Error, "GetGroupList failed")
}
func (d *DataBase) GetAllGroupInfoByGroupIDOrGroupName(keyword string, isSearchGroupID bool, isSearchGroupName bool) ([]*model_struct.LocalGroup, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var groupList []model_struct.LocalGroup
	var condition string
	if isSearchGroupID {
		if isSearchGroupName {
			condition = fmt.Sprintf("group_id like %q or name like %q", "%"+keyword+"%", "%"+keyword+"%")
		} else {
			condition = fmt.Sprintf("group_id like %q ", "%"+keyword+"%")
		}
	} else {
		condition = fmt.Sprintf("name like %q ", "%"+keyword+"%")
	}
	err := d.conn.Where(condition).Order("create_time DESC").Find(&groupList).Error
	var transfer []*model_struct.LocalGroup
	for _, v := range groupList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, utils.Wrap(err, "GetAllGroupInfoByGroupIDOrGroupName failed ")
}

func (d *DataBase) GetGroupUpdateVersionByID(groupID string) (model_struct.LocalGroupUpdatesVersion, error) {
	var localgroupUpdateVersion model_struct.LocalGroupUpdatesVersion
	localgroupUpdateVersion.GroupID = groupID
	err := d.conn.Where("group_id = ?", groupID).Find(&localgroupUpdateVersion).Error
	if err != nil {
		var localgroupUpdateVersion model_struct.LocalGroupUpdatesVersion
		localgroupUpdateVersion.GroupID = groupID
		localgroupUpdateVersion.VersionNumber = int64(0)
		localgroupUpdateVersion.UpdateTime = time.Now()
		d.conn.Create(&localgroupUpdateVersion)
		log.Info("Record not found Localy set version number to -1")
		localgroupUpdateVersion.VersionNumber = 0
	}
	return localgroupUpdateVersion, utils.Wrap(err, "GetGroupUpdateVersionByID failed ")
}
func (d *DataBase) UpdateGroupUpdateVersionByID(groupID string, versionNumber int32) model_struct.LocalGroupUpdatesVersion {
	var localgroupUpdateVersion model_struct.LocalGroupUpdatesVersion
	localgroupUpdateVersion.GroupID = groupID
	localgroupUpdateVersion.VersionNumber = int64(versionNumber)
	localgroupUpdateVersion.UpdateTime = time.Now()
	log.Info("Local Group Version Update ", localgroupUpdateVersion)
	if d.conn.Where("group_id = ?", groupID).Updates(localgroupUpdateVersion).RowsAffected == 0 {
		d.conn.Create(&localgroupUpdateVersion) // create new record from newUser
		log.Info("Local Group Version insert ", localgroupUpdateVersion)
	}
	return localgroupUpdateVersion
}

func (d *DataBase) GetFriendListNotMemberOfTheGroup(groupID string) []model_struct.LocalFriend {
	var localFriendsList = new([]model_struct.LocalFriend)
	sqlQuery := "select * from local_friends AS lf where not EXISTS (Select local_group_members.user_id from local_group_members  where lf.friend_user_id = local_group_members.user_id and local_group_members.group_id  = " + groupID + ")"
	trx := d.conn.Raw(sqlQuery)
	err := trx.Scan(&localFriendsList).Error
	if err != nil {
		log.Info("GetFriendListNotMemberOfTheGroupSQL ", sqlQuery, err.Error())
	}
	return *localFriendsList
}
