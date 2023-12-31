package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"errors"
)

func (d *DataBase) InsertAdminGroupRequest(groupRequest *model_struct.LocalAdminGroupRequest) error {
	return utils.Wrap(d.conn.Create(groupRequest).Error, "InsertAdminGroupRequest failed")
}

func (d *DataBase) DeleteAdminGroupRequest(groupID, userID string) error {
	return utils.Wrap(d.conn.Where("group_id=? and user_id=?", groupID, userID).Delete(&model_struct.LocalAdminGroupRequest{}).Error, "DeleteAdminGroupRequest failed")
}

func (d *DataBase) UpdateAdminGroupRequest(groupRequest *model_struct.LocalAdminGroupRequest) error {
	t := d.conn.Model(groupRequest).Select("*").Updates(*groupRequest)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "")
}

func (d *DataBase) GetAdminGroupApplication() ([]*model_struct.LocalAdminGroupRequest, error) {
	var groupRequestList []model_struct.LocalAdminGroupRequest
	err := utils.Wrap(d.conn.Order("create_time DESC").Find(&groupRequestList).Error, "")
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var transfer []*model_struct.LocalAdminGroupRequest
	for _, v := range groupRequestList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, nil
}

func (d *DataBase) GetAdminGroupApplicationWithUserAndGroup(groupID, userID string) (*model_struct.LocalAdminGroupRequest, error) {
	var groupRequest model_struct.LocalAdminGroupRequest
	err := utils.Wrap(d.conn.Where("group_id=? and user_id=?", groupID, userID).Find(&groupRequest).Error, "")
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	return &groupRequest, nil
}
