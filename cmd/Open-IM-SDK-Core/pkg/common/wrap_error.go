package common

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"github.com/mitchellh/mapstructure"
)

func GetGroupMemberListByGroupID(callback open_im_sdk_callback.Base, operationID string, db *db.DataBase, groupID string) []*model_struct.LocalGroupMember {
	memberList, err := db.GetGroupMemberListByGroupID(groupID)
	CheckDBErrCallback(callback, err, operationID)
	return memberList
}

func MapstructureDecode(input interface{}, output interface{}, callback open_im_sdk_callback.Base, oprationID string) {
	err := mapstructure.Decode(input, output)
	CheckDataErrCallback(callback, err, oprationID)
}
