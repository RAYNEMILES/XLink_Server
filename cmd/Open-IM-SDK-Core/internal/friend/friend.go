// Copyright 2021 OpenIM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package friend

import (
	comm "Open_IM/cmd/Open-IM-SDK-Core/internal/common"
	ws "Open_IM/cmd/Open-IM-SDK-Core/internal/interaction"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/user"
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	sdk "Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	api "Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	"Open_IM/pkg/common/log"
	sdk2 "Open_IM/pkg/proto/sdk_ws"
	"errors"
)

type Friend struct {
	friendListener open_im_sdk_callback.OnFriendshipListener
	loginUserID    string
	db             *db.DataBase
	user           *user.User
	p              *ws.PostApi
	loginTime      int64
}

func (f *Friend) LoginTime() int64 {
	return f.loginTime
}

func (f *Friend) SetLoginTime(loginTime int64) {
	f.loginTime = loginTime
}

func (f *Friend) Db() *db.DataBase {
	return f.db
}

func NewFriend(loginUserID string, db *db.DataBase, user *user.User, p *ws.PostApi) *Friend {
	return &Friend{loginUserID: loginUserID, db: db, user: user, p: p}
}

func (f *Friend) SetListener(listener open_im_sdk_callback.OnFriendshipListener) {
	f.friendListener = listener
}

func (f *Friend) getDesignatedFriendsInfo(callback open_im_sdk_callback.Base, friendUserIDList sdk.GetDesignatedFriendsInfoParams, operationID string) sdk.GetDesignatedFriendsInfoCallback {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", friendUserIDList)

	localFriendList, err := f.db.GetFriendInfoList(friendUserIDList)
	common.CheckDBErrCallback(callback, err, operationID)

	blackList, err := f.db.GetBlackInfoList(friendUserIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	for _, v := range blackList {
		log.Info(operationID, "GetBlackInfoList ", *v)
	}

	r := common.MergeFriendBlackResult(localFriendList, blackList)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", r)
	return r
}

func (f *Friend) GetUserNameAndFaceUrlByUid(friendUserID, operationID string) (faceUrl, name string, accountStatus int8, err error, isFromSvr bool) {
	isFromSvr = false
	friendInfo, err := f.db.GetFriendInfoByFriendUserID(friendUserID)
	if err == nil {
		if friendInfo.Remark != "" {
			return friendInfo.FaceURL, friendInfo.Remark, friendInfo.AccountStatus, nil, isFromSvr
		} else {
			return friendInfo.FaceURL, friendInfo.Nickname, friendInfo.AccountStatus, nil, isFromSvr
		}
	} else {
		if operationID == "" {
			operationID = utils.OperationIDGenerator()
		}
		userInfos, err := f.user.GetUsersInfoFromSvrNoCallback([]string{friendUserID}, operationID)
		if err != nil {
			return "", "", 1, err, isFromSvr
		}
		for _, v := range userInfos {
			isFromSvr = true
			//change server response with Account Status
			return v.FaceURL, v.Nickname, 1, nil, isFromSvr
		}
		log.Info(operationID, "GetUsersInfoFromSvr ", friendUserID)
	}
	return "", "", 1, errors.New("getUserNameAndFaceUrlByUid err"), isFromSvr
}

func (f *Friend) GetDesignatedFriendListInfo(callback open_im_sdk_callback.Base, friendUserIDList []string, operationID string) []*model_struct.LocalFriend {
	friendList, err := f.db.GetFriendInfoList(friendUserIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	return friendList
}

func (f *Friend) GetDesignatedBlackListInfo(callback open_im_sdk_callback.Base, blackIDList []string, operationID string) []*model_struct.LocalBlack {
	blackList, err := f.db.GetBlackInfoList(blackIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	return blackList
}

func (f *Friend) addFriend(callback open_im_sdk_callback.Base, userIDReqMsg sdk.AddFriendParams, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userIDReqMsg)
	apiReq := api.AddFriendReq{}
	apiReq.ToUserID = userIDReqMsg.ToUserID
	apiReq.FromUserID = f.loginUserID
	apiReq.ReqMsg = userIDReqMsg.ReqMsg
	apiReq.OperationID = operationID
	apiReq.Source = userIDReqMsg.Source
	f.p.PostFatalCallback(callback, constant.AddFriendRouter, apiReq, nil, operationID)
	f.SyncFriendApplication(operationID)
	//f.SyncSelfFriendApplication(operationID)
}

func (f *Friend) getRecvFriendApplicationList(callback open_im_sdk_callback.Base, operationID string) sdk.GetRecvFriendApplicationListCallback {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	friendApplicationList, err := f.db.GetRecvFriendApplication()
	common.CheckDBErrCallback(callback, err, operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", friendApplicationList)
	return friendApplicationList
}

func (f *Friend) getSendFriendApplicationList(callback open_im_sdk_callback.Base, operationID string) sdk.GetSendFriendApplicationListCallback {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	friendApplicationList, err := f.db.GetSendFriendApplication()
	common.CheckDBErrCallback(callback, err, operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", friendApplicationList)
	return friendApplicationList
}

func (f *Friend) getFriendApplicationList(callback open_im_sdk_callback.Base, operationID string) sdk.GetSendFriendApplicationListCallback {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	friendApplicationList, err := f.db.GetFriendApplication()
	common.CheckDBErrCallback(callback, err, operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", friendApplicationList)
	return friendApplicationList
}

func (f *Friend) processFriendApplication(callback open_im_sdk_callback.Base, userIDHandleMsg sdk.ProcessFriendApplicationParams, handleResult int32, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userIDHandleMsg, handleResult)
	apiReq := api.AddFriendResponseReq{}
	apiReq.FromUserID = f.loginUserID
	apiReq.ToUserID = userIDHandleMsg.ToUserID
	apiReq.Flag = handleResult
	apiReq.OperationID = operationID
	apiReq.HandleMsg = userIDHandleMsg.HandleMsg
	f.p.PostFatalCallback(callback, constant.AddFriendResponse, apiReq, nil, operationID)
	f.SyncFriendApplication(operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ")
}

func (f *Friend) checkFriend(callback open_im_sdk_callback.Base, friendUserIDList sdk.CheckFriendParams, operationID string) sdk.CheckFriendCallback {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", friendUserIDList)
	friendList, err := f.db.GetFriendInfoList(friendUserIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	blackList, err := f.db.GetBlackInfoList(friendUserIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	var checkFriendCallback sdk.CheckFriendCallback
	for _, v := range friendUserIDList {
		var r api.UserIDResult
		isBlack := false
		isFriend := false
		for _, b := range blackList {
			if v == b.BlockUserID {
				isBlack = true
				break
			}
		}
		for _, f := range friendList {
			if v == f.FriendUserID {
				isFriend = true
				break
			}
		}
		r.UserID = v
		if isFriend && !isBlack {
			r.Result = 1
		} else {
			r.Result = 0
		}
		checkFriendCallback = append(checkFriendCallback, r)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", checkFriendCallback)
	return checkFriendCallback
}

func (f *Friend) deleteFriend(friendUserID sdk.DeleteFriendParams, callback open_im_sdk_callback.Base, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", friendUserID)
	apiReq := api.DeleteFriendReq{}
	apiReq.ToUserID = string(friendUserID)
	apiReq.FromUserID = f.loginUserID
	apiReq.OperationID = operationID
	apiResp := api.DeleteFriendResp{}
	f.p.PostFatalCallback(callback, constant.DeleteFriendRouter, apiReq, &apiResp, operationID)
	f.SyncFriendList(operationID)
	//if apiResp.ErrCode == 0 {
	//	log.NewInfo(operationID, utils.GetSelfFuncName(), apiResp.ErrCode, apiResp.ErrMsg)
	//	f.DeleteMeAsFriendForUser(operationID, apiReq.ToUserID)
	//}

	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ")
}

func (f *Friend) getFriendList(callback open_im_sdk_callback.Base, operationID string) sdk.GetFriendListCallback {
	localFriendList, err := f.db.GetAllFriendList()
	common.CheckDBErrCallback(callback, err, operationID)
	localBlackList, err := f.db.GetBlackList()
	common.CheckDBErrCallback(callback, err, operationID)
	return common.MergeFriendBlackResult(localFriendList, localBlackList)
}
func (f *Friend) searchFriends(callback open_im_sdk_callback.Base, param sdk.SearchFriendsParam, operationID string) sdk.SearchFriendsCallback {
	if len(param.KeywordList) == 0 || (!param.IsSearchNickname && !param.IsSearchUserID && !param.IsSearchRemark) {
		common.CheckAnyErrCallback(callback, 201, errors.New("keyword is null or search field all false"), operationID)
	}
	localFriendList, err := f.db.SearchFriendList(param.KeywordList[0], param.IsSearchUserID, param.IsSearchNickname, param.IsSearchRemark)
	common.CheckDBErrCallback(callback, err, operationID)
	localBlackList, err := f.db.GetBlackList()
	common.CheckDBErrCallback(callback, err, operationID)
	return mergeFriendBlackSearchResult(localFriendList, localBlackList)
}
func mergeFriendBlackSearchResult(base []*model_struct.LocalFriend, add []*model_struct.LocalBlack) (result []*sdk.SearchFriendItem) {
	blackUserIDList := func(bl []*model_struct.LocalBlack) (result []string) {
		for _, v := range bl {
			result = append(result, v.BlockUserID)
		}
		return result
	}(add)
	for _, v := range base {
		node := sdk.SearchFriendItem{}
		node.OwnerUserID = v.OwnerUserID
		node.FriendUserID = v.FriendUserID
		node.Remark = v.Remark
		node.CreateTime = v.CreateTime
		node.AddSource = v.AddSource
		node.OperatorUserID = v.OperatorUserID
		node.Nickname = v.Nickname
		node.FaceURL = v.FaceURL
		node.Gender = v.Gender
		node.PhoneNumber = v.PhoneNumber
		node.Birth = v.Birth
		node.Email = v.Email
		node.Ex = v.Ex
		node.AttachedInfo = v.AttachedInfo
		if !utils.IsContain(v.FriendUserID, blackUserIDList) {
			node.Relationship = constant.FriendRelationship
		}
		result = append(result, &node)
	}
	return result
}
func (f *Friend) getBlackList(callback open_im_sdk_callback.Base, operationID string) sdk.GetBlackListCallback {
	localBlackList, err := f.db.GetBlackList()
	common.CheckDBErrCallback(callback, err, operationID)

	localFriendList, err := f.db.GetAllFriendList()
	common.CheckDBErrCallback(callback, err, operationID)

	return common.MergeBlackFriendResult(localBlackList, localFriendList)
}

func (f *Friend) setFriendRemark(userIDRemark sdk.SetFriendRemarkParams, callback open_im_sdk_callback.Base, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userIDRemark)
	apiReq := api.SetFriendRemarkReq{}
	apiReq.OperationID = operationID
	apiReq.ToUserID = userIDRemark.ToUserID
	apiReq.FromUserID = f.loginUserID
	apiReq.Remark = userIDRemark.Remark
	f.p.PostFatalCallback(callback, constant.SetFriendRemark, apiReq, nil, operationID)
	f.SyncFriendList(operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ")
}

func (f *Friend) getFriendRemarkOrNick(userIDRemark sdk.GetFriendRemarkOrNickParams, callback open_im_sdk_callback.Base, operationID string) (api.GetFriendRemarkOrNickResp, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userIDRemark)
	apiReq := api.GetFriendRemarkOrNickReq{}
	apiReq.OperationID = operationID
	apiReq.ForUserID = userIDRemark.ForUserID
	apiReq.GroupID = userIDRemark.GroupID
	apiRes := api.GetFriendRemarkOrNickResp{}
	f.p.PostFatalCallback(callback, constant.GetFriendRemarkOrNick, apiReq, &apiRes, operationID)
	// if err != nil {
	// 	return apiRes, utils.Wrap(err, apiReq.OperationID)
	// }
	// f.SyncFriendList(operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", apiRes)
	return apiRes, nil
}

func (f *Friend) getServerFriendList(operationID string) ([]*sdk2.FriendInfo, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	apiReq := api.GetFriendListReq{OperationID: operationID, FromUserID: f.loginUserID}
	realData := api.GetFriendListResp{}
	err := f.p.PostReturn(constant.GetFriendListRouter, apiReq, &realData.FriendInfoList)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", realData.FriendInfoList)
	return realData.FriendInfoList, nil
}

func (f *Friend) GetFriendInfoFromSvr(operationID string, friendUserID []string) ([]*sdk2.FriendInfo, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", friendUserID)
	apiReq := api.GetFriendsInfoReq{OperationID: operationID, FriendUserIDs: friendUserID}
	realData := api.GetFriendsInfoResp{}
	err := f.p.PostReturn(constant.GetFriendsInfo, apiReq, &realData.FriendInfoList)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", realData.FriendInfoList)
	return realData.FriendInfoList, nil
}

func (f *Friend) getServerFriendListForUser(operationID, userID string) ([]*sdk2.FriendInfo, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	apiReq := api.GetFriendListReq{OperationID: operationID, FromUserID: userID}
	realData := api.GetFriendListResp{}
	err := f.p.PostReturn(constant.GetFriendListRouter, apiReq, &realData.FriendInfoList)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", realData.FriendInfoList)
	return realData.FriendInfoList, nil
}

func (f *Friend) getServerBlackList(operationID string) ([]*sdk2.PublicUserInfo, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	apiReq := api.GetBlackListReq{OperationID: operationID, FromUserID: f.loginUserID}
	realData := api.GetBlackListResp{}
	err := f.p.PostReturn(constant.GetBlackListRouter, apiReq, &realData.BlackUserInfoList)
	if err != nil {
		return nil, utils.Wrap(err, operationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", realData.BlackUserInfoList)
	return realData.BlackUserInfoList, nil
}

// recv
func (f *Friend) getFriendApplicationFromServer(operationID string) ([]*sdk2.FriendRequest, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	apiReq := api.GetFriendApplyListReq{OperationID: operationID, FromUserID: f.loginUserID}
	realData := api.GetFriendApplyListResp{}
	err := f.p.PostReturn(constant.GetFriendApplicationListRouter, apiReq, &realData.FriendRequestList)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", realData.FriendRequestList)
	return realData.FriendRequestList, nil
}

// send
func (f *Friend) getSelfFriendApplicationFromServer(operationID string) ([]*sdk2.FriendRequest, error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	apiReq := api.GetSelfFriendApplyListReq{OperationID: operationID, FromUserID: f.loginUserID}
	realData := api.GetSelfFriendApplyListResp{}
	err := f.p.PostReturn(constant.GetSelfFriendApplicationListRouter, apiReq, &realData.FriendRequestList)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ", realData.FriendRequestList)
	return realData.FriendRequestList, nil
}

func (f *Friend) addBlack(callback open_im_sdk_callback.Base, blackUserID sdk.AddBlackParams, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", blackUserID)
	apiReq := api.AddBlacklistReq{}
	apiReq.ToUserID = string(blackUserID)
	apiReq.FromUserID = f.loginUserID
	apiReq.OperationID = operationID
	f.p.PostFatalCallback(callback, constant.AddBlackRouter, apiReq, nil, operationID)
	f.SyncBlackList(operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ")
}

func (f *Friend) removeBlack(callback open_im_sdk_callback.Base, blackUserID sdk.RemoveBlackParams, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", blackUserID)
	apiReq := api.RemoveBlackListReq{}
	apiReq.ToUserID = string(blackUserID)
	apiReq.FromUserID = f.loginUserID
	apiReq.OperationID = operationID
	f.p.PostFatalCallback(callback, constant.RemoveBlackRouter, apiReq, nil, operationID)
	f.SyncBlackList(operationID)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "return: ")
}

func (f *Friend) SyncSelfFriendApplication(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := f.getSelfFriendApplicationFromServer(operationID)
	if err != nil {
		log.NewError(operationID, "getSelfFriendApplicationFromServer failed ", err.Error())
		return
	}
	onServer := common.TransferToLocalFriendRequest(svrList)
	onLocal, err := f.db.GetSendFriendApplication()
	if err != nil {
		log.NewError(operationID, "GetSendFriendApplication failed ", err.Error())
		return
	}
	log.NewInfo(operationID, "list", svrList, onServer, onLocal)

	aInBNot, bInANot, sameA, sameB := common.CheckFriendRequestDiff(onServer, onLocal)
	log.Debug(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	for _, index := range aInBNot {
		err := f.db.InsertFriendRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertFriendRequest failed ", err.Error())
			continue
		}

		toUserDB, err := db.NewDataBase(onServer[index].ToUserID, sdk_struct.SvrConf.DataDir)
		if err != nil {
			log.NewError(operationID, "init ToUserDB failed ", onServer[index].ToUserID, err.Error(), *onServer[index])
			continue
		} else {
			_, err := toUserDB.GetFriendApplicationByBothID(onServer[index].FromUserID, onServer[index].ToUserID)
			if err == nil {
				err := toUserDB.UpdateFriendRequest(onServer[index])
				if err != nil {
					log.NewError(operationID, "UpdateSelfFriendRequest to ToUser failed ", onServer[index].ToUserID, err.Error(), *onServer[index])
					continue
				}
			} else {
				err := toUserDB.InsertFriendRequest(onServer[index])
				if err != nil {
					log.NewError(operationID, "InsertFriendRequest to ToUser failed ", onServer[index].ToUserID, err.Error(), *onServer[index])
					continue
				}
			}
		}

		callbackData := sdk.FriendApplicationAddedCallback(*onServer[index])
		f.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnFriendApplicationAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := f.db.UpdateFriendRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateSelfFriendRequest failed ", err.Error(), *onServer[index])
			continue
		} else {

			toUserDB, err := db.NewDataBase(onServer[index].ToUserID, sdk_struct.SvrConf.DataDir)
			if err != nil {
				log.NewError(operationID, "init ToUserDB failed ", onServer[index].ToUserID, err.Error(), *onServer[index])
				continue
			} else {
				_, err := toUserDB.GetFriendApplicationByBothID(onServer[index].FromUserID, onServer[index].ToUserID)
				if err == nil {
					err := toUserDB.UpdateFriendRequest(onServer[index])
					if err != nil {
						log.NewError(operationID, "UpdateSelfFriendRequest to ToUser failed ", onServer[index].ToUserID, err.Error(), *onServer[index])
						continue
					}
				} else {
					err := toUserDB.InsertFriendRequest(onServer[index])
					if err != nil {
						log.NewError(operationID, "InsertFriendRequest to ToUser failed ", onServer[index].ToUserID, err.Error(), *onServer[index])
						continue
					}
				}
			}

			log.NewInfo(operationID, utils.GetSelfFuncName(), "UpdateSelfFriendRequest success!")
			if onServer[index].HandleResult == constant.FriendResponseRefuse {
				callbackData := sdk.FriendApplicationRejectCallback(*onServer[index])
				f.friendListener.OnFriendApplicationRejected(utils.StructToJsonString(callbackData))
				log.Info(operationID, "OnFriendApplicationRejected", utils.StructToJsonString(callbackData))

			} else if onServer[index].HandleResult == constant.FriendResponseAgree {
				callbackData := sdk.FriendApplicationAcceptCallback(*onServer[index])
				f.friendListener.OnFriendApplicationAccepted(utils.StructToJsonString(callbackData))
				log.Info(operationID, "OnFriendApplicationAccepted", utils.StructToJsonString(callbackData))
			} else {
				callbackData := sdk.FriendApplicationAddedCallback(*onServer[index])
				f.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
				log.Info(operationID, "OnFriendApplicationAdded", utils.StructToJsonString(callbackData))
			}

		}
	}
	for _, index := range bInANot {
		err := f.db.DeleteFriendRequestBothUserID(onLocal[index].FromUserID, onLocal[index].ToUserID)
		if err != nil {
			log.NewError(operationID, "_deleteFriendRequestBothUserID failed ", err.Error(), onLocal[index].FromUserID, onLocal[index].ToUserID)
			continue
		}
		callbackData := sdk.FriendApplicationDeletedCallback(*onLocal[index])
		f.friendListener.OnFriendApplicationDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnFriendApplicationDeleted", utils.StructToJsonString(callbackData))
	}
}

// recv
func (f *Friend) SyncFriendApplication(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := f.getFriendApplicationFromServer(operationID)
	if err != nil {
		log.NewError(operationID, "getFriendApplicationFromServer failed ", err.Error())
		return
	}
	onServer := common.TransferToLocalFriendRequest(svrList)
	onLocal, err := f.db.GetRecvFriendApplication()
	if err != nil {
		log.NewError(operationID, "GetRecvFriendApplication failed ", err.Error())
		return
	}
	log.NewInfo(operationID, "list", svrList, onServer, onLocal)

	aInBNot, bInANot, sameA, sameB := common.CheckFriendRequestDiff(onServer, onLocal)
	log.Debug(operationID, "diff ", aInBNot, bInANot, sameA, sameB)

	for _, index := range aInBNot {
		err := f.db.InsertFriendRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertFriendRequest failed ", err.Error())
			continue
		}

		fromUserDB, err := db.NewDataBase(onServer[index].FromUserID, sdk_struct.SvrConf.DataDir)
		if err != nil {
			log.NewError(operationID, "init FromUserDB failed ", onServer[index].FromUserID, err.Error(), *onServer[index])
			continue
		} else {
			_, err := fromUserDB.GetFriendApplicationByBothID(onServer[index].FromUserID, onServer[index].ToUserID)
			if err == nil {
				err := fromUserDB.UpdateFriendRequest(onServer[index])
				if err != nil {
					log.NewError(operationID, "UpdateFriendRequest to FromUser failed ", onServer[index].FromUserID, err.Error(), *onServer[index])
					continue
				}
			} else {
				err := fromUserDB.InsertFriendRequest(onServer[index])
				if err != nil {
					log.NewError(operationID, "InsertFriendRequest to FromUser failed ", onServer[index].FromUserID, err.Error(), *onServer[index])
					continue
				}
			}
		}

		callbackData := sdk.FriendApplicationAddedCallback(*onServer[index])
		//f.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
		f.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnReceiveFriendApplicationAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {

		err := f.db.UpdateFriendRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateFriendRequest failed ", err.Error(), *onServer[index])
			continue
		} else {

			fromUserDB, err := db.NewDataBase(onServer[index].FromUserID, sdk_struct.SvrConf.DataDir)
			if err != nil {
				log.NewError(operationID, "init FromUserDB failed ", onServer[index].FromUserID, err.Error(), *onServer[index])
				continue
			} else {
				_, err := fromUserDB.GetFriendApplicationByBothID(onServer[index].FromUserID, onServer[index].ToUserID)
				if err == nil {
					err := fromUserDB.UpdateFriendRequest(onServer[index])
					if err != nil {
						log.NewError(operationID, "UpdateFriendRequest to FromUser failed ", onServer[index].FromUserID, err.Error(), *onServer[index])
						continue
					}
				} else {
					err := fromUserDB.InsertFriendRequest(onServer[index])
					if err != nil {
						log.NewError(operationID, "InsertFriendRequest to FromUser failed ", onServer[index].FromUserID, err.Error(), *onServer[index])
						continue
					}
				}
			}

			log.NewInfo(operationID, utils.GetSelfFuncName(), "UpdateFriendRequest success!")
			if onServer[index].HandleResult == constant.FriendResponseRefuse {
				callbackData := sdk.FriendApplicationRejectCallback(*onServer[index])
				f.friendListener.OnFriendApplicationRejected(utils.StructToJsonString(callbackData))
				log.Info(operationID, "OnFriendApplicationRejected", utils.StructToJsonString(callbackData))
			} else if onServer[index].HandleResult == constant.FriendResponseAgree {
				callbackData := sdk.FriendApplicationAcceptCallback(*onServer[index])
				f.friendListener.OnFriendApplicationAccepted(utils.StructToJsonString(callbackData))
				log.Info(operationID, "OnFriendApplicationAccepted", utils.StructToJsonString(callbackData))
			} else {
				callbackData := sdk.FriendApplicationAddedCallback(*onServer[index])
				f.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
				log.Info(operationID, "OnReceiveFriendApplicationAdded", utils.StructToJsonString(callbackData))
			}

		}
	}
	for _, index := range bInANot {
		err := f.db.DeleteFriendRequestBothUserID(onLocal[index].FromUserID, onLocal[index].ToUserID)
		if err != nil {
			log.NewError(operationID, "_deleteFriendRequestBothUserID failed ", err.Error(), onLocal[index].FromUserID, onLocal[index].ToUserID)
			continue
		}
		callbackData := sdk.FriendApplicationDeletedCallback(*onLocal[index])
		f.friendListener.OnFriendApplicationDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnReceiveFriendApplicationDeleted", utils.StructToJsonString(callbackData))
	}
}

func (f *Friend) SyncFriendList(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := f.getServerFriendList(operationID)
	if err != nil {
		log.NewError(operationID, "getServerFriendList failed ", err.Error())
		return
	}
	friendsInfoOnServer := common.TransferToLocalFriend(svrList)
	friendsInfoOnLocal, err := f.db.GetAllFriendList()
	if err != nil {
		log.NewError(operationID, "_getAllFriendList failed ", err.Error())
		return
	}
	log.NewInfo(operationID, "list ", svrList, friendsInfoOnServer, friendsInfoOnLocal)
	for _, v := range friendsInfoOnServer {
		log.NewDebug(operationID, "friendsInfoOnServer ", *v)
	}
	aInBNot, bInANot, sameA, sameB := common.CheckFriendListDiff(friendsInfoOnServer, friendsInfoOnLocal)
	log.NewInfo(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	f.db.DeleteFriend("")
	for _, index := range aInBNot {
		// user info is null
		if friendsInfoOnServer[index].FriendUserID == "" {
			continue
		}

		err := f.db.InsertFriend(friendsInfoOnServer[index])
		if err != nil {
			log.NewError(operationID, "_insertFriend failed ", err.Error())
			continue
		}
		callbackData := sdk.FriendAddedCallback(*friendsInfoOnServer[index])
		f.friendListener.OnFriendAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnFriendAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := f.db.UpdateFriend(friendsInfoOnServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateFriendRequest failed ", err.Error(), *friendsInfoOnServer[index])
			continue
			//if !strings.Contains(err.Error(), "RowsAffected == 0") {
			//
			//}

		} else {
			callbackData := sdk.FriendInfoChangedCallback(*friendsInfoOnServer[index])
			f.friendListener.OnFriendInfoChanged(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnFriendInfoChanged", utils.StructToJsonString(callbackData))
		}
		localFriend := friendsInfoOnServer[index]
		go func() {
			conversationID := utils.GetConversationIDBySessionType(localFriend.FriendUserID, constant.SingleChatType)
			conversation, _ := f.db.GetConversation(conversationID)
			if conversation != nil {
				log.Error(operationID, "Sync Frind And Conv", utils.StructToJsonString(conversation))
				conversation.FaceURL = localFriend.FaceURL
				if localFriend.Remark == "" {
					conversation.ShowName = localFriend.Nickname
				}
				err := f.db.UpdateConversation(conversation)
				if err != nil {
					log.Error(operationID, "Sync Frind And Conv", err.Error())
				}

			}
		}()
	}
	for _, index := range bInANot {
		err := f.db.DeleteFriend(friendsInfoOnLocal[index].FriendUserID)
		if err != nil {
			log.NewError(operationID, "_deleteFriend failed ", err.Error())
			continue
		}
		callbackData := sdk.FriendDeletedCallback(*friendsInfoOnLocal[index])
		f.friendListener.OnFriendDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnFriendDeleted", utils.StructToJsonString(callbackData))
	}
}

func (f *Friend) SyncFriendInfo(operationID string, friendUserIDs []string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", friendUserIDs)
	svrList, err := f.GetFriendInfoFromSvr(operationID, friendUserIDs)
	if err != nil {
		log.NewError(operationID, "getServerFriendList failed ", err.Error())
		return
	}
	friendsInfoOnServer := common.TransferToLocalFriend(svrList)
	friendsInfoOnLocal, err := f.db.GetAllFriendList()
	if err != nil {
		log.NewError(operationID, "_getAllFriendList failed ", err.Error())
		return
	}
	log.NewInfo(operationID, "list ", svrList, friendsInfoOnServer, friendsInfoOnLocal)
	for _, v := range friendsInfoOnServer {
		log.NewDebug(operationID, "friendsInfoOnServer ", *v)
	}
	aInBNot, bInANot, sameA, sameB := common.CheckFriendListDiff(friendsInfoOnServer, friendsInfoOnLocal)
	log.NewInfo(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	for _, index := range aInBNot {
		err := f.db.InsertFriend(friendsInfoOnServer[index])
		if err != nil {
			log.NewError(operationID, "_insertFriend failed ", err.Error())
			continue
		}
		callbackData := sdk.FriendAddedCallback(*friendsInfoOnServer[index])
		f.friendListener.OnFriendAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnFriendAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := f.db.UpdateFriend(friendsInfoOnServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateFriendRequest failed ", err.Error(), *friendsInfoOnServer[index])
			continue
			//if !strings.Contains(err.Error(), "RowsAffected == 0") {
			//
			//}

		} else {
			callbackData := sdk.FriendInfoChangedCallback(*friendsInfoOnServer[index])
			f.friendListener.OnFriendInfoChanged(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnFriendInfoChanged", utils.StructToJsonString(callbackData))
		}
	}
	for _, index := range bInANot {
		err := f.db.DeleteFriend(friendsInfoOnLocal[index].FriendUserID)
		if err != nil {
			log.NewError(operationID, "_deleteFriend failed ", err.Error())
			continue
		}
		callbackData := sdk.FriendDeletedCallback(*friendsInfoOnLocal[index])
		f.friendListener.OnFriendDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnFriendDeleted", utils.StructToJsonString(callbackData))
	}
}

func (f *Friend) DeleteMeAsFriendForUser(operationID, userID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userID)
	database, _ := db.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if database != nil {
		err := database.DeleteFriendForUser(userID, f.loginUserID)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		}
	} else {
		log.NewError(operationID, utils.GetSelfFuncName(), "database is nil", userID)
	}
}

func (f *Friend) SyncBlackList(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := f.getServerBlackList(operationID)
	if err != nil {
		log.NewError(operationID, "getServerBlackList failed ", err.Error())
		return
	}
	blackListOnServer := common.TransferToLocalBlack(svrList, f.loginUserID)
	blackListOnLocal, err := f.db.GetBlackList()
	if err != nil {
		log.NewError(operationID, "_getBlackList failed ", err.Error())
		return
	}
	log.NewInfo(operationID, "list ", svrList, blackListOnServer, blackListOnLocal)
	aInBNot, bInANot, sameA, sameB := common.CheckBlackListDiff(blackListOnServer, blackListOnLocal)
	log.NewInfo(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	for _, index := range aInBNot {
		err := f.db.InsertBlack(blackListOnServer[index])
		if err != nil {
			log.NewError(operationID, "_insertFriend failed ", err.Error())
			continue
		}
		callbackData := sdk.BlackAddCallback(*blackListOnServer[index])
		f.friendListener.OnBlackAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnBlackAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := f.db.UpdateBlack(blackListOnServer[index])
		if err != nil {
			log.NewError(operationID, "_updateFriend failed ", err.Error())
			continue
		}
		//todo : add black info update callback
		log.Info(operationID, "black info update, do nothing ", blackListOnServer[index])
	}
	for _, index := range bInANot {
		err := f.db.DeleteBlack(blackListOnLocal[index].BlockUserID)
		if err != nil {
			log.NewError(operationID, "_deleteFriend failed ", err.Error())
			continue
		}
		callbackData := sdk.BlackDeletedCallback(*blackListOnLocal[index])
		f.friendListener.OnBlackDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnBlackDeleted", utils.StructToJsonString(callbackData))
	}
}

func (f *Friend) DoNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value) {
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg)
	if f.friendListener == nil {
		log.Error(operationID, "f.friendListener == nil")
		return
	}
	if msg.SendTime < f.loginTime {
		log.Warn(operationID, "ignore notification ", msg.ClientMsgID, msg.ServerMsgID, msg.Seq, msg.ContentType)
		return
	}
	go func() {
		switch msg.ContentType {
		case constant.FriendApplicationNotification:
			f.friendApplicationNotification(msg, operationID)
		case constant.FriendApplicationApprovedNotification:
			f.friendApplicationApprovedNotification(msg, operationID)
		case constant.FriendApplicationRejectedNotification:
			f.friendApplicationRejectedNotification(msg, operationID)
		case constant.FriendAddedNotification:
			f.friendAddedNotification(msg, operationID)
		case constant.FriendDeletedNotification:
			f.friendDeletedNotification(msg, operationID)
		case constant.FriendRemarkSetNotification:
			f.friendRemarkNotification(msg, conversationCh, operationID)
		case constant.UserInfoUpdatedNotification:
			f.friendInfoChangedNotification(msg, conversationCh, operationID)
		case constant.BlackAddedNotification:
			f.blackAddedNotification(msg, operationID)
		case constant.BlackDeletedNotification:
			f.blackDeletedNotification(msg, operationID)
		default:
			log.Error(operationID, "type failed ", msg.ClientMsgID, msg.ServerMsgID, msg.ContentType)
		}
	}()
}

func (f *Friend) blackDeletedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.BlackDeletedTips{FromToUserID: &sdk2.FromToUserID{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg.Content)
		return
	}
	if detail.FromToUserID.FromUserID == f.loginUserID {
		f.SyncBlackList(operationID)
	}
}

func (f *Friend) blackAddedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.BlackAddedTips{FromToUserID: &sdk2.FromToUserID{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg.Content)
		return
	}
	if detail.FromToUserID.FromUserID == f.loginUserID {
		f.SyncBlackList(operationID)
	}
}

func (f *Friend) friendRemarkNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	var detail sdk2.FriendInfoChangedTips
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg.Content)
		return
	}
	if detail.FromToUserID.FromUserID == f.loginUserID {
		f.SyncFriendList(operationID)
		conversationID := utils.GetConversationIDBySessionType(detail.FromToUserID.ToUserID, constant.SingleChatType)
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UpdateFaceUrlAndNickName, Args: common.SourceIDAndSessionType{SourceID: detail.FromToUserID.ToUserID, SessionType: constant.SingleChatType}}, conversationCh)
	}
}

func (f *Friend) friendInfoChangedNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	var detail sdk2.UserInfoUpdatedTips
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg.Content)
		return
	}
	if detail.UserID != f.loginUserID {
		f.SyncFriendList(operationID)
		conversationID := utils.GetConversationIDBySessionType(detail.UserID, constant.SingleChatType)
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UpdateFaceUrlAndNickName, Args: common.SourceIDAndSessionType{SourceID: detail.UserID, SessionType: constant.SingleChatType}}, conversationCh)
		go func() {
			friendInfo, err := f.db.GetFriendInfoByFriendUserID(detail.UserID)
			if err == nil {
				_ = f.db.UpdateMsgSenderFaceURLAndSenderNickname(detail.UserID, friendInfo.FaceURL, friendInfo.Nickname, constant.SingleChatType)
			}
		}()

	} else {
		f.user.SyncLoginUserInfo(operationID)
		go func() {
			loginUserInfo, err := f.db.GetLoginUser()
			if err == nil {
				_ = f.db.UpdateMsgSenderFaceURLAndSenderNickname(detail.UserID, loginUserInfo.FaceURL, loginUserInfo.Nickname, constant.SingleChatType)
			}
		}()
	}
}

func (f *Friend) friendDeletedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.FriendDeletedTips{FromToUserID: &sdk2.FromToUserID{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.FromToUserID.FromUserID == f.loginUserID {
		f.SyncFriendList(operationID)
		return
	}
}

func (f *Friend) friendAddedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.FriendAddedTips{Friend: &sdk2.FriendInfo{}, OpUser: &sdk2.PublicUserInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg)
		return
	}
	log.Info("detail: ", detail.Friend)
	f.SyncFriendList(operationID)
	if detail.Friend.OwnerUserID == f.loginUserID || detail.Friend.FriendUser.UserID == f.loginUserID {
		f.SyncFriendList(operationID)
		return
	}
}

func (f *Friend) friendApplicationNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.FriendApplicationTips{FromToUserID: &sdk2.FromToUserID{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.FromToUserID.FromUserID == f.loginUserID {
		log.Info(operationID, "SyncSelfFriendApplication ", detail.FromToUserID.FromUserID)
		f.SyncSelfFriendApplication(operationID)
		return
	}
	if detail.FromToUserID.ToUserID == f.loginUserID {
		log.Info(operationID, "SyncFriendApplication ", detail.FromToUserID.FromUserID, detail.FromToUserID.ToUserID)
		f.SyncFriendApplication(operationID)
		return
	}
	log.Error(operationID, "FromToUserID failed ", detail.FromToUserID)
}

func (f *Friend) friendApplicationRejectedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.FriendApplicationRejectedTips{FromToUserID: &sdk2.FromToUserID{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if f.loginUserID == detail.FromToUserID.FromUserID {
		f.SyncFriendApplication(operationID)
		return
	}
	if f.loginUserID == detail.FromToUserID.ToUserID {
		f.SyncSelfFriendApplication(operationID)
		return
	}
	log.Error(operationID, "FromToUserID failed ", detail.FromToUserID)
}

func (f *Friend) friendApplicationApprovedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.FriendApplicationApprovedTips{FromToUserID: &sdk2.FromToUserID{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "comm.UnmarshalTips failed ", err.Error(), msg)
		return
	}

	//f.SyncFriendList(operationID)
	if f.loginUserID == detail.FromToUserID.FromUserID {
		f.SyncFriendApplication(operationID)
		return
	}
	if f.loginUserID == detail.FromToUserID.ToUserID {
		f.SyncSelfFriendApplication(operationID)
		return
	}
	log.Error(operationID, "FromToUserID failed ", detail.FromToUserID)
}
