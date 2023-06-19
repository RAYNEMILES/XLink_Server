package user

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
)

//func (u *User) GetUsersInfo(callback open_im_sdk_callback.Base, userIDList string, operationID string) {
//	fName := utils.GetSelfFuncName()
//	go func() {
//		log.NewInfo(operationID, fName, "args: ", userIDList)
//		var unmarshalParam sdk_params_callback.GetUsersInfoParam
//		common.JsonUnmarshalAndArgsValidate(userIDList, &unmarshalParam, callback, operationID)
//		result := u.GetUsersInfoFromSvr(callback, unmarshalParam, operationID)
//		callback.OnSuccess(utils.StructToJsonStringDefault(result))
//		log.NewInfo(operationID, fName, "callback: ", utils.StructToJsonStringDefault(result))
//	}()
//}

func (u *User) GetSelfUserInfo(callback open_im_sdk_callback.Base, operationID string) {
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ")
		result := u.getSelfUserInfo(callback, operationID)
		callback.OnSuccess(utils.StructToJsonString(result))
		log.NewInfo(operationID, fName, "callback: ", utils.StructToJsonString(result))
	}()
}

func (u *User) SetSelfInfo(callback open_im_sdk_callback.Base, userInfo string, operationID string) {
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", userInfo)
		var unmarshalParam sdk_params_callback.SetSelfUserInfoParam
		common.JsonUnmarshalAndArgsValidate(userInfo, &unmarshalParam, callback, operationID)
		u.updateSelfUserInfo(callback, unmarshalParam, operationID)
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.SetSelfUserInfoCallback))
		log.NewInfo(operationID, fName, "callback: ", utils.StructToJsonString(sdk_params_callback.SetSelfUserInfoCallback))
	}()
}

func (u *User) RemoveFaceUrl(callback open_im_sdk_callback.Base, userInfo string, operationID string) {
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", userInfo)
		var unmarshalParam sdk_params_callback.RemoveFaceUrlParam
		common.JsonUnmarshalAndArgsValidate(userInfo, &unmarshalParam, callback, operationID)
		u.removeUserFaceUrl(callback, unmarshalParam, operationID)
		callback.OnSuccess(userInfo)
		log.NewInfo(operationID, fName, "callback: ", utils.StructToJsonString(sdk_params_callback.SetSelfUserInfoCallback))
	}()
}

func (u *User) updateMsgSenderInfo(nickname, faceURL string, operationID string) {
	if nickname != "" {
		err := u.DataBase.UpdateMsgSenderNickname(u.loginUserID, nickname, constant.SingleChatType)
		if err != nil {
			log.Error(operationID, "UpdateMsgSenderNickname failed ", err.Error(), u.loginUserID, nickname, constant.SingleChatType)
		}
	}
	if faceURL != "" {
		err := u.DataBase.UpdateMsgSenderFaceURL(u.loginUserID, faceURL, constant.SingleChatType)
		if err != nil {
			log.Error(operationID, "UpdateMsgSenderFaceURL failed ", err.Error(), u.loginUserID, faceURL, constant.SingleChatType)
		}
	}
}
