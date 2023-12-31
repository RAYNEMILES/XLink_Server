package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"errors"
)

func (d *DataBase) InsertGroupRequest(groupRequest *model_struct.LocalGroupRequest) error {
	return utils.Wrap(d.conn.Create(groupRequest).Error, "InsertGroupRequest failed")
}
func (d *DataBase) DeleteGroupRequest(groupID, userID string) error {
	return utils.Wrap(d.conn.Where("group_id=? and user_id=?", groupID, userID).Delete(&model_struct.LocalGroupRequest{}).Error, "DeleteGroupRequest failed")
}
func (d *DataBase) UpdateGroupRequest(groupRequest *model_struct.LocalGroupRequest) error {
	t := d.conn.Model(groupRequest).Select("*").Updates(*groupRequest)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "")
}

func (d *DataBase) GetSendGroupApplication() ([]*model_struct.LocalGroupRequest, error) {
	var groupRequestList []model_struct.LocalGroupRequest
	err := utils.Wrap(d.conn.Order("create_time DESC").Find(&groupRequestList).Error, "")
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var transfer []*model_struct.LocalGroupRequest
	for _, v := range groupRequestList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, nil
}

func (d *DataBase) GetGroupApplicationWithUserIDAndGroupID(userID, groupID string) (*model_struct.LocalGroupRequest, error) {
	var groupRequest model_struct.LocalGroupRequest
	err := utils.Wrap(d.conn.Where("group_id=? and user_id=?", groupID, userID).Find(&groupRequest).Error, "")
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	return &groupRequest, nil
}
