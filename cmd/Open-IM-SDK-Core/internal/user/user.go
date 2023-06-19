package user

import (
	comm "Open_IM/cmd/Open-IM-SDK-Core/internal/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/pkg/common/log"
	"github.com/google/go-cmp/cmp"

	//"github.com/mitchellh/mapstructure"
	ws "Open_IM/cmd/Open-IM-SDK-Core/internal/interaction"
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"

	sdk "Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	api "Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	sdk2 "Open_IM/pkg/proto/sdk_ws"
)

type User struct {
	*db.DataBase
	p           *ws.PostApi
	loginUserID string
	listener    open_im_sdk_callback.OnUserListener
	loginTime   int64
	platformID  int32
}

func (u *User) LoginTime() int64 {
	return u.loginTime
}

func (u *User) SetLoginTime(loginTime int64) {
	u.loginTime = loginTime
}

func (u *User) SetListener(listener open_im_sdk_callback.OnUserListener) {
	u.listener = listener
}

func NewUser(dataBase *db.DataBase, p *ws.PostApi, loginUserID string, platformID int32) *User {
	return &User{DataBase: dataBase, p: p, loginUserID: loginUserID, platformID: platformID}
}

func (u *User) DoNotification(msg *sdk2.MsgData) {
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg)
	if u.listener == nil {
		log.Error(operationID, "listener == nil")
		return
	}

	if msg.SendTime < u.loginTime {
		log.Warn(operationID, "ignore notification ", msg.ClientMsgID, msg.ServerMsgID, msg.Seq, msg.ContentType)
		return
	}

	go func() {
		switch msg.ContentType {
		case constant.UserInfoUpdatedNotification:
			u.userInfoUpdatedNotification(msg, operationID)
		default:
			log.Error(operationID, "type failed ", msg.ClientMsgID, msg.ServerMsgID, msg.ContentType)
		}
	}()
}

func (u *User) userInfoUpdatedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	var detail sdk2.UserInfoUpdatedTips
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg.Content)
		return
	}
	if detail.UserID == u.loginUserID {
		log.Info(operationID, "detail.UserID == u.loginUserID, SyncLoginUserInfo", detail.UserID)
		u.SyncLoginUserInfo(operationID)
		user, err := u.GetLoginUser()
		if err != nil {
			go u.updateMsgSenderInfo(user.Nickname, user.FaceURL, operationID)
		}
	} else {
		log.Info(operationID, "detail.UserID != u.loginUserID, do nothing", detail.UserID, u.loginUserID)
		//from svr
		notIn := make([]string, 0)
		notIn = append(notIn, detail.UserID)
		if len(notIn) > 0 {
			publicList, err := u.GetUsersInfoFromSvrNoCallback(notIn, operationID)
			if err != nil {
				return
			}
			go func() {
				for _, v := range publicList {
					//Update the faceURL and nickname information of the local chat history with non-friends
					_ = u.UpdateMsgSenderFaceURLAndSenderNickname(v.UserID, v.FaceURL, v.Nickname, constant.SingleChatType)
					_ = u.UpdateMsgSenderFaceURLAndSenderNickname(v.UserID, v.FaceURL, v.Nickname, constant.GroupChatType)

				}
			}()
		}
	}
}

func (u *User) SyncLoginUserInfo(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svr, err := u.GetSelfUserInfoFromSvr(operationID)
	if err != nil {
		log.Error(operationID, "GetSelfUserInfoFromSvr failed", err.Error())
		return
	}
	onServer := common.TransferToLocalUserInfo(svr)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "onServer: ", *onServer)
	onLocal, err := u.GetLoginUser()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "onLocal: ", *onLocal)
	if err != nil {
		log.Warn(operationID, "GetLoginUser failed", err.Error())
		onLocal = &model_struct.LocalUser{}
	}
	if !cmp.Equal(onServer, onLocal) {
		if onLocal.UserID == "" {
			log.NewInfo(operationID, utils.GetSelfFuncName(), "InsertLoginUser")
			if err = u.InsertLoginUser(onServer); err != nil {
				log.Error(operationID, "InsertLoginUser failed ", *onServer, err.Error())
			}
			return
		}
		err = u.UpdateLoginUserByMap(onServer, map[string]interface{}{"name": onServer.Nickname, "face_url": onServer.FaceURL,
			"gender": onServer.Gender, "phone_number": onServer.PhoneNumber, "birth": onServer.Birth, "email": onServer.Email, "create_time": onServer.CreateTime, "app_manger_level": onServer.AppMangerLevel, "ex": onServer.Ex, "attached_info": onServer.AttachedInfo, "global_recv_msg_opt": onServer.GlobalRecvMsgOpt})
		//fmt.Println("UpdateLoginUser ", *onServer, svr)
		if err != nil {
			log.Error(operationID, "UpdateLoginUser failed ", *onServer, err.Error())
			return
		}
		callbackData := sdk.SelfInfoUpdatedCallback(*onServer)
		if u.listener == nil {
			log.Error(operationID, "u.listener == nil")
			return
		}
		u.listener.OnSelfInfoUpdated(utils.StructToJsonString(callbackData))
	}
}

func (u *User) GetUsersInfoFromSvr(callback open_im_sdk_callback.Base, UserIDList sdk.GetUsersInfoParam, operationID string) []*sdk2.PublicUserInfo {
	apiReq := api.GetUsersInfoReq{}
	apiReq.OperationID = operationID
	apiReq.UserIDList = UserIDList
	apiResp := api.GetUsersInfoResp{}
	u.p.PostFatalCallback(callback, constant.GetUsersInfoRouter, apiReq, &apiResp.UserInfoList, apiReq.OperationID)
	return apiResp.UserInfoList
}

func (u *User) GetUsersInfoFromSvrNoCallback(UserIDList sdk.GetUsersInfoParam, operationID string) ([]*sdk2.PublicUserInfo, error) {
	apiReq := api.GetUsersInfoReq{}
	apiReq.OperationID = operationID
	apiReq.UserIDList = UserIDList
	apiResp := api.GetUsersInfoResp{}
	err := u.p.PostReturn(constant.GetUsersInfoRouter, apiReq, &apiResp.UserInfoList)
	return apiResp.UserInfoList, err
}

func (u *User) GetUsersInfoFromCacheSvr(UserIDList sdk.GetUsersInfoParam, operationID string) ([]*sdk2.PublicUserInfo, error) {
	apiReq := api.GetUsersInfoReq{}
	apiReq.OperationID = operationID
	apiReq.UserIDList = UserIDList
	apiResp := api.GetUsersInfoResp{}
	err := u.p.PostReturn(constant.GetUsersInfoFromCacheRouter, apiReq, &apiResp.UserInfoList)
	return apiResp.UserInfoList, err
}

func (u *User) getSelfUserInfo(callback open_im_sdk_callback.Base, operationID string) sdk.GetSelfUserInfoCallback {
	userInfo, err := u.GetLoginUser()
	common.CheckDBErrCallback(callback, err, operationID)
	return userInfo
}

func (u *User) updateSelfUserInfo(callback open_im_sdk_callback.Base, userInfo sdk.SetSelfUserInfoParam, operationID string) {
	apiReq := api.UpdateSelfUserInfoReq{}
	apiReq.OperationID = operationID
	apiReq.ApiUserInfo = api.ApiUserInfo(userInfo)
	apiReq.UserID = u.loginUserID
	u.p.PostFatalCallback(callback, constant.UpdateSelfUserInfoRouter, apiReq, nil, apiReq.OperationID)
	u.SyncLoginUserInfo(operationID)
}

func (u *User) removeUserFaceUrl(callback open_im_sdk_callback.Base, userInfo sdk.RemoveFaceUrlParam, operationID string) {
	apiReq := api.UpdateSelfUserInfoReq{}
	apiReq.OperationID = operationID
	apiReq.ApiUserInfo = api.ApiUserInfo(userInfo)
	apiReq.UserID = u.loginUserID
	u.p.PostFatalCallback(callback, constant.RemoveUserFaceUrlRouter, apiReq, nil, apiReq.OperationID)
	u.SyncLoginUserInfo(operationID)
}

func (u *User) GetSelfUserInfoFromSvr(operationID string) (*sdk2.UserInfo, error) {
	log.Debug(operationID, utils.GetSelfFuncName())

	apiReq := api.GetSelfUserInfoReq{}
	apiReq.OperationID = operationID
	apiReq.UserID = u.loginUserID
	apiResp := api.GetSelfUserInfoResp{UserInfo: &sdk2.UserInfo{}}
	err := u.p.PostReturn(constant.GetSelfUserInfoRouter, apiReq, &apiResp.UserInfo)
	if err != nil {
		log.Error(operationID, utils.GetSelfFuncName(), "apiResp Error:", err.Error())
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "apiResp.UserInfo:", apiReq, apiResp.UserInfo)
	return apiResp.UserInfo, nil
}

func (u *User) DoUserNotification(msg *sdk2.MsgData) {
	if msg.SendID == u.loginUserID && msg.SenderPlatformID == sdk_struct.SvrConf.Platform {
		return
	}
}

func (u *User) ParseTokenFromSvr(operationID string) (uint32, error) {
	apiReq := api.ParseTokenReq{}
	apiReq.OperationID = operationID
	apiReq.UserID = u.loginUserID
	apiReq.PlatformID = u.platformID
	apiResp := api.ParseTokenResp{}
	err := u.p.PostReturn(constant.ParseTokenRouter, apiReq, &apiResp.ExpireTime)
	if err != nil {
		return 0, utils.Wrap(err, apiReq.OperationID)
	}
	log.Info(operationID, "apiResp.ExpireTime.ExpireTimeSeconds ", apiResp.ExpireTime)
	return apiResp.ExpireTime.ExpireTimeSeconds, nil
}

func (u *User) UpdateUserIPandStatus(operationID, ipAddress string) {
	apiReq := api.UpdateUserIPAndStatusReq{}
	apiReq.OperationID = operationID
	apiReq.UserID = u.loginUserID
	apiReq.IPaddress = ipAddress

	apiResp := api.UpdateUserIPAndStatusResp{}
	err := u.p.PostReturn(constant.UpdateUserIPandStatus, apiReq, &apiResp)
	if err != nil {
		log.Info(operationID, "Update User IP on server failed ")
		return

	}
	log.Info(operationID, "Update User IP on server success ")

}

func (u *User) StartingWelcomeMessagesFromSvr(operationID string) (map[string]interface{}, error) {
	apiReq := api.GetInviteAndChannelReq{}
	apiReq.OperationID = operationID
	apiResp := api.GetInviteAndChannelResp{}
	err := u.p.PostReturn(constant.StartingWelcomeMessagesRouter, apiReq, &apiResp.Data)
	return apiResp.Data, err
}

func (u *User) NotifyMomentNotification(msg *sdk2.MsgData) {
	detail := sdk2.MomentCommentNotificationTips{}
	if err := comm.UnmarshalTipsSync(msg, &detail); err != nil {
		log.Error("UnmarshalTips failed ", err.Error(), msg)
		return
	}
	log.NewInfo(utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID, msg.String(), "detail : ", detail.String())
	u.listener.OnMomentNotification(utils.StructToJsonString(detail))
}
