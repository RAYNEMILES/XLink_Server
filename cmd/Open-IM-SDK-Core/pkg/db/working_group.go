package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
)

func (d *DataBase) GetJoinedWorkingGroupIDList() ([]string, error) {
	groupList, err := d.GetJoinedGroupList()
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var groupIDList []string
	for _, v := range groupList {
		if v.GroupType == constant.WorkingGroup {
			groupIDList = append(groupIDList, v.GroupID)
		}
	}
	return groupIDList, nil
}

func (d *DataBase) GetJoinedWorkingGroupList() ([]*model_struct.LocalGroup, error) {
	groupList, err := d.GetJoinedGroupList()
	var transfer []*model_struct.LocalGroup
	for _, v := range groupList {
		if v.GroupType == constant.WorkingGroup {
			transfer = append(transfer, v)
		}
	}
	return transfer, utils.Wrap(err, "GetJoinedSuperGroupList failed ")
}
