package group

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	api2 "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/utils"
	"errors"
	"sync"
	"time"

	"github.com/jinzhu/copier"

	comm "Open_IM/cmd/Open-IM-SDK-Core/internal/common"
	ws "Open_IM/cmd/Open-IM-SDK-Core/internal/interaction"
	sdk "Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	api "Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	pbGroup "Open_IM/pkg/proto/group"
	sdk2 "Open_IM/pkg/proto/sdk_ws"
)

// //utils.GetCurrentTimestampByMill()
type Group struct {
	listener              open_im_sdk_callback.OnGroupListener
	loginUserID           string
	db                    *db.DataBase
	p                     *ws.PostApi
	loginTime             int64
	joinedSuperGroupCh    chan common.Cmd2Value
	heartbeatCmdCh        chan common.Cmd2Value
	groupMemberSyncStatus map[string]pbGroup.GroupMemberSyncStatusModel
	sync.Mutex
}

func (g *Group) LoginTime() int64 {
	return g.loginTime
}

func (g *Group) SetLoginTime(loginTime int64) {
	g.loginTime = loginTime
}

func NewGroup(loginUserID string, db *db.DataBase, p *ws.PostApi, joinedSuperGroupCh chan common.Cmd2Value, heartbeatCmdCh chan common.Cmd2Value) *Group {
	return &Group{loginUserID: loginUserID, db: db, p: p, joinedSuperGroupCh: joinedSuperGroupCh, heartbeatCmdCh: heartbeatCmdCh, groupMemberSyncStatus: make(map[string]pbGroup.GroupMemberSyncStatusModel)}
}

func (g *Group) DoNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value) {
	if g.listener == nil {
		return
	}
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID, msg.ContentType)
	if msg.SendTime < g.loginTime {
		log.Warn(operationID, "ignore notification ", msg.ClientMsgID, msg.ServerMsgID, msg.Seq, msg.ContentType)
		return
	}
	go func() {
		switch msg.ContentType {
		case constant.GroupCreatedNotification:
			g.groupCreatedNotification(msg, operationID)
		case constant.GroupInfoSetNotification:
			g.groupInfoSetNotification(msg, conversationCh, operationID)
		case constant.GroupAnnouncementNotification:
			g.groupAnnouncementSetNotification(msg, conversationCh, operationID)
		case constant.JoinGroupApplicationNotification:
			g.joinGroupApplicationNotification(msg, operationID)
		case constant.MemberQuitNotification:
			g.memberQuitNotification(msg, operationID)
		case constant.GroupApplicationAcceptedNotification:
			g.groupApplicationAcceptedNotification(msg, operationID)
		case constant.GroupApplicationRejectedNotification:
			g.groupApplicationRejectedNotification(msg, operationID)
		case constant.GroupOwnerTransferredNotification:
			g.groupOwnerTransferredNotification(msg, operationID)
		case constant.MemberKickedNotification:
			g.memberKickedNotification(msg, operationID)
		case constant.MemberInvitedNotification:
			g.memberInvitedNotification(msg, operationID)
		case constant.MemberEnterNotification:
			g.memberEnterNotification(msg, operationID)
		case constant.GroupDismissedNotification:
			g.groupDismissNotification(msg, operationID)
		case constant.GroupMemberMutedNotification:
			g.groupMemberMuteChangedNotification(msg, false, operationID)
		case constant.GroupMemberCancelMutedNotification:
			g.groupMemberMuteChangedNotification(msg, true, operationID)
		case constant.GroupMutedNotification:
			fallthrough
		case constant.GroupCancelMutedNotification:
			g.groupMuteChangedNotification(msg, operationID)
		case constant.GroupMemberInfoSetNotification:
			g.groupMemberInfoSetNotification(msg, operationID)
		case constant.GroupMemberSetToAdminNotification:
			g.groupMemberInfoSetNotification(msg, operationID)
		case constant.GroupMemberSetToOrdinaryUserNotification:
			g.groupMemberInfoSetNotification(msg, operationID)
		case constant.ConversationChangeNotification:
			g.batchSetConversation(msg, operationID)
		case constant.GroupDeleteNotification:
			g.groupDeleteNotification(msg, operationID)
		case constant.GroupMemberSyncSilentSDKNotification:
			detail := sdk2.MemberSyncNotificationTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}, InvitedUserList: make([]*sdk2.GroupMemberFullInfo, 0)}
			if err := comm.UnmarshalTipsSync(msg, &detail); err != nil {
				log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
				return
			}
			log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID, msg.String(), "detail : ", detail.String())
			g.groupMemberSyncSilentSDKNotification(&detail, operationID)
		default:
			log.Error(operationID, "ContentType tip failed ", msg.ContentType)
		}
	}()
}

func (g *Group) groupCreatedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupCreatedTips{Group: &sdk2.GroupInfo{}}
	comm.UnmarshalTips(msg, &detail)
	g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, false)
	g.SyncJoinedGroupList(operationID)
	g.listener.OnGroupMembersSyncedStarted(utils.StructToJsonString(detail.Group))
	log.NewInfo(operationID, utils.GetSelfFuncName(), "Group Create Notification Sync Started", msg.ClientMsgID, msg.ServerMsgID)
}

func (g *Group) groupInfoSetNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupInfoSetTips{Group: &sdk2.GroupInfo{}}
	comm.UnmarshalTips(msg, &detail)
	g.SyncJoinedGroupList(operationID) //todo,  sync some group info
	conversationID := utils.GetConversationIDBySessionType(detail.Group.GroupID, constant.GroupChatType)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UpdateFaceUrlAndNickName, Args: common.SourceIDAndSessionType{SourceID: detail.Group.GroupID, SessionType: constant.GroupChatType}}, conversationCh)
}

func (g *Group) groupAnnouncementSetNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupInfoSetTips{Group: &sdk2.GroupInfo{}}
	comm.UnmarshalTips(msg, &detail)
	g.SyncJoinedGroupList(operationID) //todo,  sync some group info
	conversationID := utils.GetConversationIDBySessionType(detail.Group.GroupID, constant.GroupChatType)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.PostGroupAnnouncement, Args: common.SourceIDAndSessionType{SourceID: detail.Group.GroupID, SessionType: constant.GroupChatType}}, conversationCh)
}

func (g *Group) joinGroupApplicationNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.JoinGroupApplicationTips{Group: &sdk2.GroupInfo{}, Applicant: &sdk2.PublicUserInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.Applicant.UserID == g.loginUserID {
		g.SyncSelfGroupApplication(operationID)
	} else {
		g.SyncAdminGroupApplication(operationID)
	}
}

func (g *Group) memberQuitNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.MemberQuitTips{Group: &sdk2.GroupInfo{}, QuitUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.QuitUser.UserID == g.loginUserID {
		g.SyncJoinedGroupList(operationID)
		g.db.DeleteGroupAllMembers(detail.Group.GroupID)
	} else {
		g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
		g.updateMemberCount(detail.Group.GroupID, operationID)
	}
}

func (g *Group) groupApplicationAcceptedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupApplicationAcceptedTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.OpUser.UserID == g.loginUserID {
		g.SyncAdminGroupApplication(operationID)
	} else {
		g.SyncSelfGroupApplication(operationID)
		g.SyncJoinedGroupList(operationID)
	}
	//g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID)
}

func (g *Group) groupApplicationRejectedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupApplicationRejectedTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.OpUser.UserID == g.loginUserID {
		g.SyncAdminGroupApplication(operationID)
	} else {
		g.SyncSelfGroupApplication(operationID)
	}
}

func (g *Group) groupOwnerTransferredNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupOwnerTransferredTips{Group: &sdk2.GroupInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	g.SyncJoinedGroupList(operationID)
	g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
	g.SyncAdminGroupApplication(operationID)
}

func (g *Group) memberKickedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.MemberKickedTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}

	log.Info(operationID, "KickedUserList ", detail.KickedUserList)
	for _, v := range detail.KickedUserList {
		if v.UserID == g.loginUserID {
			g.SyncJoinedGroupList(operationID)
			g.db.DeleteGroupAllMembers(detail.Group.GroupID)
			return
		}
	}
	g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
	g.updateMemberCount(detail.Group.GroupID, operationID)
}

func (g *Group) memberInvitedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.MemberInvitedTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}

	for _, v := range detail.InvitedUserList {
		if v.UserID == g.loginUserID {
			g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, false)
			g.SyncJoinedGroupList(operationID)

			return
		}
	}
	g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
	g.updateMemberCount(detail.Group.GroupID, operationID)
}

func (g *Group) memberEnterNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.MemberEnterTips{Group: &sdk2.GroupInfo{}, EntrantUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	if detail.EntrantUser.UserID == g.loginUserID {
		g.SyncJoinedGroupList(operationID)
		g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, false)
	} else {
		g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
		g.updateMemberCount(detail.Group.GroupID, operationID)
	}
}

func (g *Group) groupDismissNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupDismissedTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	g.SyncJoinedGroupList(operationID)
	g.db.DeleteGroupAllMembers(detail.Group.GroupID)

}

func (g *Group) groupDeleteNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	detail := sdk2.GroupDismissedTips{Group: &sdk2.GroupInfo{}, OpUser: &sdk2.GroupMemberFullInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	g.SyncJoinedGroupList(operationID)
	g.db.DeleteGroupAllMembers(detail.Group.GroupID)

}

func (g *Group) groupMemberMuteChangedNotification(msg *sdk2.MsgData, isCancel bool, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	var syncGroupID string
	if isCancel {
		detail := sdk2.GroupMemberCancelMutedTips{Group: &sdk2.GroupInfo{}}
		if err := comm.UnmarshalTips(msg, &detail); err != nil {
			log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
			return
		}
		syncGroupID = detail.Group.GroupID
	} else {
		detail := sdk2.GroupMemberMutedTips{Group: &sdk2.GroupInfo{}}
		if err := comm.UnmarshalTips(msg, &detail); err != nil {
			log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
			return
		}
		syncGroupID = detail.Group.GroupID
	}
	g.syncGroupMemberByGroupID(syncGroupID, operationID, true)
}

func (g *Group) groupMuteChangedNotification(msg *sdk2.MsgData, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID)
	g.SyncJoinedGroupList(operationID)
}

func (g *Group) groupMemberInfoSetNotification(msg *sdk2.MsgData, operationID string) {
	detail := sdk2.GroupMemberInfoSetTips{Group: &sdk2.GroupInfo{}}
	if err := comm.UnmarshalTips(msg, &detail); err != nil {
		log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
		return
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID, msg.String(), "detail : ", detail.String())
	g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
	_ = g.db.UpdateMsgSenderFaceURLAndSenderNickname(detail.ChangedUser.UserID, detail.ChangedUser.FaceURL, detail.ChangedUser.Nickname, constant.GroupChatType)
}

func (g *Group) groupMemberSyncSilentSDKNotification(detail *sdk2.MemberSyncNotificationTips, operationID string) {

	if detail == nil {
		return
	}
	//todo #1 store the current provided data and update the local version number with the highest version in group members
	//log.Error(operationID, "Get response from Sync Friend Request ", detail)
	var updatedVersion int32 = 0
	if detail.InvitedUserList != nil {
		for _, info := range detail.InvitedUserList {
			if info.DeleteTime == 0 {
				//todo new user invited, add in local db
				var insertGroupMember model_struct.LocalGroupMember
				err := utils2.CopyStructFields(&insertGroupMember, info)
				if err != nil {
					log.Error(operationID, "Copy Tips UserInvited to LocalGroup ", err.Error())
				} else {
					if insertGroupMember.AccountStatus == 2 {
						g.db.UpdateUserNameInGroupLocalChatLogs(info.GroupID, insertGroupMember.UserID)
					}
					err := g.db.UpsertGroupMember(&insertGroupMember)
					if err != nil {
						log.NewError(operationID, "InsertGroupMemberSync failed ", err.Error(), insertGroupMember)
						continue
					}

				}
			} else {
				//todo need to remove this user from local db
				err := g.db.DeleteGroupMember(info.GroupID, info.UserID)
				if err != nil {
					log.NewError(operationID, "DeleteGroupMemberSync failed ", err.Error(), info)
					continue
				}
			}
			if updatedVersion < info.UpdateVersion {
				updatedVersion = info.UpdateVersion
			}
		}
	} else {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "detail.InvitedUserList is nil", "detail : ", detail.String())
	}
	if detail.PageNumber == 1 {
		g.listener.OnGroupMembersSyncedStarted(utils.StructToJsonString(detail.Group))
		g.Lock()
		g.groupMemberSyncStatus[detail.Group.GroupID] = pbGroup.GroupMemberSyncStatusModel{GroupID: detail.Group.GroupID, PageNumber: 1, TotalCount: int32(detail.Group.MemberCount), SyncStartTime: int32(time.Now().Unix()), Priority: 0}
		g.Unlock()
		//TODO update group details
		if detail.Group != nil {
			localGroup, err := g.db.GetGroupInfoByGroupID(detail.Group.GroupID)
			if err == nil && localGroup != nil {
				localGroup.MemberCount = int32(detail.Group.MemberCount)
				if err == nil && localGroup.GroupID != "" {
					err = g.db.UpdateGroup(localGroup)
					if err == nil {
						callbackData := sdk.GroupInfoChangedCallback(*localGroup)
						g.listener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
					}
				}
			}
		}
	}
	//Update by callback to FE to update the Group Member List
	if detail.NeedNextPageFetch {
		//todo #2 request for next page
		apiReq := api.CheckLocalGroupUpdatesVersionsWithSrvReq{}
		apiReq.OperationID = operationID
		apiReq.GroupID = detail.Group.GroupID
		apiReq.GroupVersion = 0
		apiReq.PageNumber = detail.PageNumber + 1
		apiReq.PageSize = detail.PageSize
		apiReq.NeedNextPageFetch = true
		apiReq.ResponseBackHTTP = detail.ResponseBackHTTP
		var result *pbGroup.GroupUpdatesVersionsRes
		log.Debug(operationID, "api args: ", apiReq)
		err := g.p.PostReturn(constant.CheckLocalGroupUpdatesVersionsWithSrv, apiReq, &result)
		if err != nil {
			log.Info(operationID, "syncGroupUpdatesVersionByGroupIDV2", err.Error())
		}
		g.Lock()
		groupSyncStatusModel := g.groupMemberSyncStatus[detail.Group.GroupID]
		groupSyncStatusModel.PageNumber = apiReq.PageNumber
		g.groupMemberSyncStatus[detail.Group.GroupID] = groupSyncStatusModel
		g.Unlock()
		log.Error(operationID, "GroupMemberSyncStatusModel", groupSyncStatusModel)
		//Calling Back for next Page Response Handler
		if apiReq.ResponseBackHTTP && result != nil {
			g.groupMemberSyncSilentSDKNotification(result.MemberSyncNotificationTips, operationID)
		}
	} else {
		g.listener.OnGroupMembersSyncedFinished(utils.StructToJsonString(detail.Group))
		g.Lock()
		delete(g.groupMemberSyncStatus, detail.Group.GroupID)
		g.Unlock()
	}
	localDBVersion, _ := g.db.GetGroupUpdateVersionByID(detail.Group.GroupID)
	if localDBVersion.VersionNumber < int64(updatedVersion) {
		g.db.UpdateGroupUpdateVersionByID(detail.Group.GroupID, updatedVersion)
	}

}

func (g *Group) batchSetConversation(msg *sdk2.MsgData, operationID string) {
	log.Error("Jumpy Jumpy Yolo")
	// detail := sdk2.GroupMemberInfoSetTips{Group: &sdk2.GroupInfo{}}
	// if err := comm.UnmarshalTips(msg, &detail); err != nil {
	// 	log.Error(operationID, "UnmarshalTips failed ", err.Error(), msg)
	// 	return
	// }
	// log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", msg.ClientMsgID, msg.ServerMsgID, msg.String(), "detail : ", detail.String())
	// g.syncGroupMemberByGroupID(detail.Group.GroupID, operationID, true)
	// _ = g.db.UpdateMsgSenderFaceURLAndSenderNickname(detail.ChangedUser.UserID, detail.ChangedUser.FaceURL, detail.ChangedUser.Nickname, constant.GroupChatType)
}

func (g *Group) createGroup(callback open_im_sdk_callback.Base, group sdk.CreateGroupBaseInfoParam, memberList sdk.CreateGroupMemberRoleParam, operationID string) *sdk.CreateGroupCallback {
	apiReq := api.CreateGroupReq{}
	apiReq.OperationID = operationID
	apiReq.OwnerUserID = g.loginUserID
	apiReq.MemberList = memberList
	apiReq.IsOpen = group.IsOpen
	apiReq.GroupInterest = group.GroupInterest

	for _, v := range apiReq.MemberList {
		if v.RoleLevel == 0 {
			v.RoleLevel = 1
		}
	}
	copier.Copy(&apiReq, &group)
	realData := api.CreateGroupResp{}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "api req args: ", apiReq)
	g.p.PostFatalCallback(callback, constant.CreateGroupRouter, apiReq, &realData.GroupInfo, apiReq.OperationID)
	m := utils.JsonDataOne(&realData.GroupInfo)
	g.SyncJoinedGroupList(operationID)
	g.syncGroupMemberByGroupID(realData.GroupInfo.GroupID, operationID, true)

	//for _, member := range memberList {
	//	go func() {
	//		g.SyncJoinedGroupForUser(operationID, realData.GroupInfo.GroupID, member.UserID)
	//	}()
	//}
	//
	//go func() {
	//	g.syncGroupMemberByGroupIDForAll(realData.GroupInfo.GroupID, operationID)
	//}()
	return (*sdk.CreateGroupCallback)(&m)
}

func (g *Group) joinGroup(groupID, reqMsg string, joinSource int32, callback open_im_sdk_callback.Base, operationID string) {
	apiReq := api.JoinGroupReq{}
	apiReq.OperationID = operationID
	apiReq.ReqMessage = reqMsg
	apiReq.GroupID = groupID
	apiReq.JoinSource = joinSource
	g.p.PostFatalCallback(callback, constant.JoinGroupRouter, apiReq, nil, apiReq.OperationID)
	g.SyncSelfGroupApplication(operationID)

	//svrList, _ := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	//if svrList != nil {
	//	log.NewInfo(operationID, "getGroupAllMemberByGroupIDFromSvr ", svrList)
	//	for _, member := range svrList {
	//		if member.RoleLevel > constant.GroupOrdinaryUsers {
	//			log.NewInfo(operationID, utils.GetSelfFuncName(), "member.UserID:", member.UserID, member.RoleLevel)
	//			go func() {
	//				g.SyncAdminGroupApplicationForMember(operationID, member.UserID)
	//			}()
	//		}
	//
	//	}
	//}
}
func (g *Group) GetGroupOwnerIDAndAdminIDList(groupID, operationID string) (ownerID string, adminIDList []string, err error) {
	localGroup, err := g.db.GetGroupInfoByGroupID(groupID)
	if err != nil {
		return "", nil, err
	}
	adminIDList, err = g.db.GetGroupAdminID(groupID)
	if err != nil {
		return "", nil, err
	}
	return localGroup.OwnerUserID, adminIDList, nil
}
func (g *Group) quitGroup(groupID string, callback open_im_sdk_callback.Base, operationID string) {
	apiReq := api.QuitGroupReq{}
	apiReq.OperationID = operationID
	apiReq.GroupID = groupID
	g.p.PostFatalCallback(callback, constant.QuitGroupRouter, apiReq, nil, apiReq.OperationID)
	//	g.syncGroupMemberByGroupID(groupID, operationID, false) //todo
	g.SyncJoinedGroupList(operationID)
}

func (g *Group) dismissGroup(groupID string, callback open_im_sdk_callback.Base, operationID string) {
	apiReq := api.DismissGroupReq{}
	apiReq.OperationID = operationID
	apiReq.GroupID = groupID
	g.p.PostFatalCallback(callback, constant.DismissGroupRouter, apiReq, nil, apiReq.OperationID)
	g.SyncJoinedGroupList(operationID)
	//go func() {
	//	g.syncGroupMemberByGroupID(groupID, operationID, true)
	//}()
}

func (g *Group) changeGroupMute(groupID string, isMute bool, callback open_im_sdk_callback.Base, operationID string) {
	if isMute {
		apiReq := api.MuteGroupReq{}
		apiReq.OperationID = operationID
		apiReq.GroupID = groupID
		g.p.PostFatalCallback(callback, constant.MuteGroupRouter, apiReq, nil, apiReq.OperationID)
	} else {
		apiReq := api.CancelMuteGroupReq{}
		apiReq.OperationID = operationID
		apiReq.GroupID = groupID
		g.p.PostFatalCallback(callback, constant.CancelMuteGroupRouter, apiReq, nil, apiReq.OperationID)
	}
	g.SyncJoinedGroupList(operationID)
	//go func() {
	//	g.SyncJoinedGroupForAll(operationID, groupID)
	//}()
}

func (g *Group) changeGroupMemberMute(groupID, userID string, mutedSeconds uint32, callback open_im_sdk_callback.Base, operationID string) {
	if mutedSeconds == 0 {
		apiReq := api.CancelMuteGroupMemberReq{}
		apiReq.OperationID = operationID
		apiReq.GroupID = groupID
		apiReq.UserID = userID
		g.p.PostFatalCallback(callback, constant.CancelMuteGroupMemberRouter, apiReq, nil, apiReq.OperationID)
	} else {
		apiReq := api.MuteGroupMemberReq{}
		apiReq.OperationID = operationID
		apiReq.GroupID = groupID
		apiReq.UserID = userID
		apiReq.MutedSeconds = mutedSeconds
		g.p.PostFatalCallback(callback, constant.MuteGroupMemberRouter, apiReq, nil, apiReq.OperationID)
	}
	//g.SyncGroupMemberInfoForOwnerAndAdmins(operationID, userID, groupID)

}

func (g *Group) setGroupMemberRoleLevel(callback open_im_sdk_callback.Base, groupID, userID string, roleLevel int, operationID string) {
	apiReq := api.SetGroupMemberRoleLevelReq{
		SetGroupMemberInfoReq: api.SetGroupMemberInfoReq{
			OperationID: operationID,
			UserID:      userID,
			GroupID:     groupID,
		},
		RoleLevel: roleLevel,
	}
	g.p.PostFatalCallback(callback, constant.SetGroupMemberInfoRouter, apiReq, nil, apiReq.OperationID)
	g.syncGroupMemberByGroupID(groupID, operationID, true)
	//go func() {
	//	g.SyncGroupMemberInfoForAllMembers(operationID, userID, groupID)
	//}()

}

func (g *Group) getJoinedGroupList(callback open_im_sdk_callback.Base, operationID string) sdk.GetJoinedGroupListCallback {
	groupList, err := g.db.GetJoinedGroupList()
	log.Info("this is rpc", groupList)
	common.CheckDBErrCallback(callback, err, operationID)
	return groupList
}

func (g *Group) GetGroupInfoFromLocal2Svr(groupID string) (*model_struct.LocalGroup, error) {
	localGroup, err := g.db.GetGroupInfoByGroupID(groupID)
	if err == nil {
		return localGroup, nil
	}
	groupIDList := []string{groupID}
	operationID := utils.OperationIDGenerator()
	svrGroup, err := g.getGroupsInfoFromSvr(groupIDList, operationID)
	if err == nil && len(svrGroup) == 1 {
		transfer := common.TransferToLocalGroupInfo(svrGroup)
		return transfer[0], nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	} else {
		return nil, utils.Wrap(errors.New("no group"), "")
	}
}

func (g *Group) GetGroupInfoFromSvr(groupID string) (*model_struct.LocalGroup, error) {
	groupIDList := []string{groupID}
	operationID := utils.OperationIDGenerator()
	svrGroup, err := g.getGroupsInfoFromSvr(groupIDList, operationID)
	if err == nil && len(svrGroup) == 1 {
		transfer := common.TransferToLocalGroupInfo(svrGroup)
		return transfer[0], nil
	}
	if err != nil {
		return nil, utils.Wrap(err, "")
	} else {
		return nil, utils.Wrap(errors.New("no group"), "")
	}
}

func (g *Group) searchGroups(callback open_im_sdk_callback.Base, param sdk.SearchGroupsParam, operationID string) sdk.SearchGroupsCallback {
	if len(param.KeywordList) == 0 || (!param.IsSearchGroupName && !param.IsSearchGroupID) {
		common.CheckAnyErrCallback(callback, 201, errors.New("keyword is null or search field all false"), operationID)
	}
	localGroup, err := g.db.GetAllGroupInfoByGroupIDOrGroupName(param.KeywordList[0], param.IsSearchGroupID, param.IsSearchGroupName)
	common.CheckDBErrCallback(callback, err, operationID)
	return localGroup
}

func (g *Group) getGroupsInfo(groupIDList sdk.GetGroupsInfoParam, callback open_im_sdk_callback.Base, operationID string) sdk.GetGroupsInfoCallback {
	groupList, err := g.db.GetJoinedGroupList()
	common.CheckDBErrCallback(callback, err, operationID)
	var result sdk.GetGroupsInfoCallback
	var notInDB []string

	for _, v := range groupList {
		in := false
		for _, k := range groupIDList {
			if v.GroupID == k {
				in = true
				break
			}
		}
		if in {
			result = append(result, v)
		}
	}

	for _, v := range groupIDList {
		in := false
		for _, k := range result {
			if v == k.GroupID {
				in = true
				break
			}
		}
		if !in {
			notInDB = append(notInDB, v)
		}
	}
	if len(notInDB) > 0 {
		groupsInfoSvr, err := g.getGroupsInfoFromSvr(notInDB, operationID)
		log.Info(operationID, "getGroupsInfoFromSvr groupsInfoSvr", groupsInfoSvr)
		common.CheckArgsErrCallback(callback, err, operationID)
		transfer := common.TransferToLocalGroupInfo(groupsInfoSvr)
		result = append(result, transfer...)
	}

	return result
}

func (g *Group) getGroupsInfoFromSvr(groupIDList []string, operationID string) ([]*sdk2.GroupInfo, error) {
	apiReq := api.GetGroupInfoReq{}
	apiReq.GroupIDList = groupIDList
	apiReq.OperationID = operationID
	var groupInfoList []*sdk2.GroupInfo
	err := g.p.PostReturn(constant.GetGroupsInfoRouter, apiReq, &groupInfoList)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return groupInfoList, nil
}

func (g *Group) setGroupInfo(callback open_im_sdk_callback.Base, groupInfo sdk.SetGroupInfoParam, groupID, operationID string) {
	apiReq := api.SetGroupInfoReq{}
	apiReq.GroupName = groupInfo.GroupName
	apiReq.FaceURL = groupInfo.FaceURL
	apiReq.Notification = groupInfo.Notification
	apiReq.Introduction = groupInfo.Introduction
	apiReq.Ex = groupInfo.Ex
	apiReq.OperationID = operationID
	apiReq.GroupID = groupID
	apiReq.NeedVerification = groupInfo.NeedVerification
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupInfo, groupID)
	g.p.PostFatalCallback(callback, constant.SetGroupInfoRouter, apiReq, nil, apiReq.OperationID)
	g.SyncJoinedGroupList(operationID)
	//g.SyncJoinedGroupForAll(operationID, groupID)
	g.listener.OnGroupInfoChanged(utils.StructToJsonString(groupInfo))
}
func (g *Group) modifyGroupInfo(callback open_im_sdk_callback.Base, apiReq api.SetGroupInfoReq, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", apiReq)
	g.p.PostFatalCallback(callback, constant.SetGroupInfoRouter, apiReq, nil, apiReq.OperationID)
	g.SyncJoinedGroupList(operationID)
}

// todo
func (g *Group) getGroupMemberList(callback open_im_sdk_callback.Base, groupID string, filter, offset, count int32, nameSearch string, operationID string) sdk.GetGroupMemberListCallback {
	//g.syncGroupMemberByGroupID(groupID, operationID, true)

	groupInfoList, err := g.db.GetGroupMemberListSplit(groupID, filter, int(offset), int(count), nameSearch)
	var waitgroup sync.WaitGroup
	if len(groupInfoList) <= 1 && nameSearch != "" {
		groupInfoList = fetchGroupMemberListServer(g, &waitgroup, groupID, filter, offset, count, nameSearch, operationID)
		// if len(groupInfoList) > 0 {
		// 	go g.syncGroupMemberByGroupIDAfterSearchOnServer(groupID, groupInfoList, operationID, false)
		// }
	}

	common.CheckDBErrCallback(callback, err, operationID)
	// log.Error(operationID, "SandmMi", "After fetching", len(groupInfoList), len(groupInfoList))
	return groupInfoList
}

func fetchGroupMemberListServer(g *Group, waitgroup *sync.WaitGroup, groupID string, filter, offset, count int32, searchName string, operationID string) []*model_struct.LocalGroupMember {
	var groupInfoList []*model_struct.LocalGroupMember
	apiReq := api.GetGroupMemberListSrvReq{}
	apiReq.GroupId = groupID
	apiReq.SearchName = searchName
	apiReq.PageNumber = int(offset)
	apiReq.ShowNumber = int(count)
	var respML *pbGroup.GetGroupMembersResV2
	err := g.p.PostReturn(constant.GetGroupAllMemberListRouterV2, apiReq, &respML)
	if err != nil {
		log.Error(operationID, "Get Group Member list from server error", err.Error())
	}
	for _, gm := range respML.Members {
		gmR := model_struct.LocalGroupMember{}
		gmR.GroupID = groupID
		gmR.UserID = gm.UserID
		gmR.Nickname = gm.Nickname
		gmR.FaceURL = gm.FaceURL
		// gmR.JoinTime = gm.JoinTime
		gmR.RoleLevel = gm.RoleLevel
		gmR.JoinSource = gm.JoinSource
		gmR.MuteEndTime = gm.MuteEndTime
		groupInfoList = append(groupInfoList, &gmR)
		// log.Error(operationID, "SandmMi", "fetchGroupMemberListServer get", fmt.Sprintf("%v", gm))
	}
	return groupInfoList
}

func (g *Group) syncGroupMemberByGroupIDAfterSearchOnServer(groupID string, onServer []*model_struct.LocalGroupMember, operationID string, onGroupMemberNotification bool) {
	// log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID)
	// svrList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	// if err != nil {
	// 	log.NewError(operationID, "getGroupAllMemberByGroupIDFromSvr failed ", err.Error(), groupID)
	// 	//return
	// }
	// log.NewInfo(operationID, "getGroupAllMemberByGroupIDFromSvr ", svrList)
	// onServer := common.TransferToLocalGroupMember(svrList)
	onLocal, err := g.db.GetGroupMemberListByGroupID(groupID)
	if err != nil {
		log.NewError(operationID, "GetGroupMemberListByGroupID failed ", err.Error(), groupID)
		return
	}
	log.Error(operationID, "SandmMi svrList onServer onLocal", len(onServer), len(onLocal))
	aInBNot, bInANot, sameA, _ := common.CheckGroupMemberDiff(onServer, onLocal)
	log.Info(operationID, "getGroupAllMemberByGroupIDFromSvr  diff ", aInBNot, bInANot, sameA)
	var insertGroupMemberList []*model_struct.LocalGroupMember
	for _, index := range aInBNot {
		// if onGroupMemberNotification == false {
		// 	insertGroupMemberList = append(insertGroupMemberList, onServer[index])
		// 	continue
		// }
		log.NewInfo(operationID, utils.GetSelfFuncName(), "aInBNot:", onServer[index])
		err := g.db.InsertGroupMember(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertGroupMember failed ", err.Error(), *onServer[index])

		} else {
			log.NewError(operationID, "InsertGroupMember Successed ", *onServer[index])
		}
		// if onGroupMemberNotification == true {
		// 	callbackData := sdk.GroupMemberAddedCallback(*onServer[index])
		// 	g.listener.OnGroupMemberAdded(utils.StructToJsonString(callbackData))
		// 	log.Debug(operationID, "OnGroupMemberAdded", utils.StructToJsonString(callbackData))
		// }
	}
	if len(insertGroupMemberList) > 0 {
		split := 1000
		idx := 0
		remain := len(insertGroupMemberList) % split
		log.Warn(operationID, "BatchInsertGroupMember all: ", len(insertGroupMemberList))
		for idx = 0; idx < len(insertGroupMemberList)/split; idx++ {
			sub := insertGroupMemberList[idx*split : (idx+1)*split]
			err = g.db.BatchInsertGroupMember(sub)
			log.Warn(operationID, "BatchInsertGroupMember len: ", len(sub))
			if err != nil {
				log.Error(operationID, "BatchInsertGroupMember failed ", err.Error(), len(sub))
				for again := 0; again < len(sub); again++ {
					if err = g.db.InsertGroupMember(sub[again]); err != nil {
						log.Error(operationID, "InsertGroupMember failed ", err.Error(), sub[again])
					}
				}
			}
		}
		if remain > 0 {
			sub := insertGroupMemberList[idx*split:]
			log.Warn(operationID, "BatchInsertGroupMember len: ", len(sub))
			if err != nil {
				log.Error(operationID, "BatchInsertGroupMember failed ", err.Error(), len(sub))
				for again := 0; again < len(sub); again++ {
					if err = g.db.InsertGroupMember(sub[again]); err != nil {
						log.Error(operationID, "InsertGroupMember failed ", err.Error(), sub[again])
					}
				}
			}
		}
	}

	for _, index := range sameA {
		err := g.db.UpdateGroupMember(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateGroupMember failed ", err.Error(), *onServer[index])
			continue
		}

		// callbackData := sdk.GroupMemberInfoChangedCallback(*onServer[index])
		// g.listener.OnGroupMemberInfoChanged(utils.StructToJsonString(callbackData))
		// log.Info(operationID, "OnGroupMemberInfoChanged", utils.StructToJsonString(callbackData))
	}
	// for _, index := range bInANot {
	// 	err := g.db.DeleteGroupMember(onLocal[index].GroupID, onLocal[index].UserID)
	// 	if err != nil {
	// 		log.NewError(operationID, "DeleteGroupMember failed ", err.Error(), onLocal[index].GroupID, onLocal[index].UserID)
	// 		continue
	// 	}
	// 	if onGroupMemberNotification == true {
	// 		callbackData := sdk.GroupMemberDeletedCallback(*onLocal[index])
	// 		g.listener.OnGroupMemberDeleted(utils.StructToJsonString(callbackData))
	// 		log.Info(operationID, "OnGroupMemberDeleted", utils.StructToJsonString(callbackData))
	// 	}
	// }
}

func (g *Group) getGroupMemberOwnerAndAdmin(callback open_im_sdk_callback.Base, groupID string, operationID string) sdk.GetGroupMemberListCallback {
	groupInfoList, err := g.db.GetGroupMemberOwnerAndAdmin(groupID)
	common.CheckDBErrCallback(callback, err, operationID)
	return groupInfoList
}
func (g *Group) getGroupMemberListByJoinTimeFilter(callback open_im_sdk_callback.Base, groupID string, offset, count int32, joinTimeBegin, joinTimeEnd int64, userIDList []string, operationID string) sdk.GetGroupMemberListCallback {
	if joinTimeEnd == 0 {
		joinTimeEnd = utils.GetCurrentTimestampBySecond()
	}
	groupInfoList, err := g.db.GetGroupMemberListSplitByJoinTimeFilter(groupID, int(offset), int(count), joinTimeBegin, joinTimeEnd, userIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	return groupInfoList
}

// todo
func (g *Group) getGroupMembersInfo(callback open_im_sdk_callback.Base, groupID string, userIDList sdk.GetGroupMembersInfoParam, operationID string) sdk.GetGroupMembersInfoCallback {
	groupInfoList, err := g.db.GetGroupSomeMemberInfo(groupID, userIDList)
	common.CheckDBErrCallback(callback, err, operationID)
	return groupInfoList
}

func (g *Group) GetOneGroupMemberInfo(groupID, userID string) (*model_struct.LocalGroupMember, error) {
	member, err := g.db.GetGroupMemberInfoByGroupIDUserID(groupID, userID)
	return member, err
}

func (g *Group) kickGroupMember(callback open_im_sdk_callback.Base, groupID string, memberList sdk.KickGroupMemberParam, reason string, operationID string) sdk.KickGroupMemberCallback {
	apiReq := api.KickGroupMemberReq{}
	apiReq.GroupID = groupID
	apiReq.KickedUserIDList = memberList
	apiReq.Reason = reason
	apiReq.OperationID = operationID
	realData := api.KickGroupMemberResp{}
	g.p.PostFatalCallback(callback, constant.KickGroupMemberRouter, apiReq, &realData.UserIDResultList, apiReq.OperationID)
	g.SyncJoinedGroupList(operationID)
	g.syncGroupMemberByGroupID(groupID, operationID, true)
	//for _, member := range memberList {
	//	go func() {
	//		g.deleteJoinedGroupForUser(operationID, groupID, member)
	//	}()
	//}
	//go func() {
	//	g.syncGroupMemberByGroupIDForAll(groupID, operationID)
	//}()
	return realData.UserIDResultList
}

// //1
func (g *Group) transferGroupOwner(callback open_im_sdk_callback.Base, groupID, newOwnerUserID string, operationID string) {
	apiReq := api.TransferGroupOwnerReq{}
	apiReq.GroupID = groupID
	apiReq.NewOwnerUserID = newOwnerUserID
	apiReq.OperationID = operationID
	apiReq.OldOwnerUserID = g.loginUserID
	g.p.PostFatalCallback(callback, constant.TransferGroupRouter, apiReq, nil, apiReq.OperationID)
	//g.SyncJoinedGroupMember(operationID)
	g.syncGroupMemberByGroupID(groupID, operationID, true)
	//go func() {
	//	g.syncGroupMemberByGroupIDForAll(groupID, operationID)
	//}()
}

func (g *Group) inviteUserToGroup(callback open_im_sdk_callback.Base, groupID, reason string, userList sdk.InviteUserToGroupParam, operationID string) sdk.InviteUserToGroupCallback {
	apiReq := api.InviteUserToGroupReq{}
	apiReq.GroupID = groupID
	apiReq.Reason = reason
	apiReq.InvitedUserIDList = userList
	apiReq.OperationID = operationID
	var realData sdk.InviteUserToGroupCallback
	g.p.PostFatalCallback(callback, constant.InviteUserToGroupRouter, apiReq, &realData, apiReq.OperationID)
	g.SyncJoinedGroupList(operationID)
	//for _, userID := range userList {
	//	go func() {
	//		g.SyncJoinedGroupForUser(operationID, groupID, userID)
	//	}()
	//}

	g.syncGroupMemberByGroupID(groupID, operationID, true)
	//go func() {
	//	g.syncGroupMemberByGroupIDForAll(groupID, operationID)
	//}()

	return realData
}

// //1
func (g *Group) getRecvGroupApplicationList(callback open_im_sdk_callback.Base, operationID string) sdk.GetGroupApplicationListCallback {
	applicationList, err := g.db.GetAdminGroupApplication()
	common.CheckDBErrCallback(callback, err, operationID)
	return applicationList
}

func (g *Group) getSendGroupApplicationList(callback open_im_sdk_callback.Base, operationID string) sdk.GetSendGroupApplicationListCallback {
	applicationList, err := g.db.GetSendGroupApplication()
	common.CheckDBErrCallback(callback, err, operationID)
	return applicationList
}

func (g *Group) getRecvGroupApplicationListFromSvr(operationID string) ([]*sdk2.GroupRequest, error) {
	apiReq := api.GetGroupApplicationListReq{}
	apiReq.FromUserID = g.loginUserID
	apiReq.OperationID = operationID
	var realData []*sdk2.GroupRequest
	err := g.p.PostReturn(constant.GetRecvGroupApplicationListRouter, apiReq, &realData)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return realData, nil
}

func (g *Group) getRecvGroupApplicationListForUserFromSvr(operationID, userID string) ([]*sdk2.GroupRequest, error) {
	apiReq := api.GetGroupApplicationListReq{}
	apiReq.FromUserID = userID
	apiReq.OperationID = operationID
	var realData []*sdk2.GroupRequest
	err := g.p.PostReturn(constant.GetRecvGroupApplicationListRouter, apiReq, &realData)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return realData, nil
}

func (g *Group) getSendGroupApplicationListFromSvr(operationID string) ([]*sdk2.GroupRequest, error) {
	apiReq := api.GetUserReqGroupApplicationListReq{}
	apiReq.UserID = g.loginUserID
	apiReq.OperationID = operationID
	var realData []*sdk2.GroupRequest
	err := g.p.PostReturn(constant.GetSendGroupApplicationListRouter, apiReq, &realData)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return realData, nil
}

func (g *Group) processGroupApplication(callback open_im_sdk_callback.Base, groupID, fromUserID, handleMsg string, handleResult int32, operationID string) {
	apiReq := api.ApplicationGroupResponseReq{}
	apiReq.GroupID = groupID
	apiReq.OperationID = operationID
	apiReq.FromUserID = fromUserID
	apiReq.HandleResult = handleResult
	apiReq.HandledMsg = handleMsg
	if handleResult == constant.GroupResponseAgree {
		g.p.PostFatalCallback(callback, constant.AcceptGroupApplicationRouter, apiReq, nil, apiReq.OperationID)
		g.syncGroupMemberByGroupID(groupID, operationID, true)
		//go func() {
		//	g.syncGroupMemberByGroupIDForAll(groupID, operationID)
		//}()
		//go func() {
		//	g.SyncJoinedGroupForUser(operationID, groupID, fromUserID)
		//}()

	} else if handleResult == constant.GroupResponseRefuse {
		g.p.PostFatalCallback(callback, constant.RefuseGroupApplicationRouter, apiReq, nil, apiReq.OperationID)
	}
	g.SyncAdminGroupApplication(operationID)
	//go func() {
	//	g.SyncGroupApplicationForUserAndGroup(operationID, groupID, fromUserID)
	//}()
	//go func() {
	//	g.SyncGroupApplicationToGroupAdmins(operationID, groupID, fromUserID)
	//}()
}

func (g *Group) getJoinedGroupListFromSvr(operationID string) ([]*sdk2.GroupInfo, error) {
	apiReq := api.GetJoinedGroupListReq{}
	apiReq.OperationID = operationID
	apiReq.FromUserID = g.loginUserID
	var result []*sdk2.GroupInfo
	log.Debug(operationID, "api args: ", apiReq)
	err := g.p.PostReturn(constant.GetJoinedGroupListRouter, apiReq, &result)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return result, nil
}

//func (g *Group) getGroupUpdatesVersionsFromSrv(operationID string, groupIDs []string) (*pbGroup.GroupUpdatesVersionsRes, error) {
//	apiReq := api.GroupUpdatesVersionsReq{}
//	apiReq.OperationID = operationID
//	apiReq.GroupIDs = append(apiReq.GroupIDs, groupIDs...)
//	var result *pbGroup.GroupUpdatesVersionsRes
//	log.Debug(operationID, "api args: ", apiReq)
//	err := g.p.PostReturn(constant.GroupUpdateVersion, apiReq, &result)
//	if err != nil {
//		return nil, utils.Wrap(err, apiReq.OperationID)
//	}
//	return result, nil
//}

func (g *Group) getJoinedGroupListForUserIDFromSvr(operationID, userID string) ([]*sdk2.GroupInfo, error) {
	apiReq := api.GetJoinedGroupListReq{}
	apiReq.OperationID = operationID
	apiReq.FromUserID = userID
	var result []*sdk2.GroupInfo
	log.Debug(operationID, "api args: ", apiReq)
	err := g.p.PostReturn(constant.GetJoinedGroupListRouter, apiReq, &result)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return result, nil
}

func (g *Group) GetJoinedDiffusionGroupIDListFromSvr(operationID string) ([]string, error) {
	result, err := g.getJoinedGroupListFromSvr(operationID)
	if err != nil {
		return nil, utils.Wrap(err, "working group get err")
	}
	var groupIDList []string
	for _, v := range result {
		if v.GroupType == constant.WorkingGroup {
			groupIDList = append(groupIDList, v.GroupID)
		}
	}
	return groupIDList, nil
}

func (g *Group) GetJoinedNormalGroupIDListFromSvr(operationID string) ([]string, error) {
	result, err := g.getJoinedGroupListFromSvr(operationID)
	if err != nil {
		return nil, utils.Wrap(err, "working group get err")
	}
	var groupIDList []string
	for _, v := range result {
		if v.GroupType == constant.NormalGroup {
			groupIDList = append(groupIDList, v.GroupID)
		}
	}
	return groupIDList, nil
}

func (g *Group) updateMemberCount(groupID string, operationID string) {
	memberCount, err := g.db.GetGroupMemberCount(groupID)
	if err != nil {
		log.Error(operationID, "GetGroupMemberCount failed ", err.Error(), groupID)
		return
	}
	groupInfo, err := g.db.GetGroupInfoByGroupID(groupID)
	if err != nil {
		log.Error(operationID, "GetGroupInfoByGroupID failed ", err.Error(), groupID)
		return
	}
	if groupInfo.MemberCount != int32(memberCount) {
		groupInfo.MemberCount = int32(memberCount)
		log.Info(operationID, "OnGroupInfoChanged , update group info", groupInfo)
		g.db.UpdateGroup(groupInfo)
		g.listener.OnGroupInfoChanged(utils.StructToJsonString(groupInfo))
	}
}

func (g *Group) getGroupMembersInfoFromSvr(groupID string, memberList []string) ([]*sdk2.GroupMemberFullInfo, error) {
	var apiReq api.GetGroupMembersInfoReq
	apiReq.OperationID = utils.OperationIDGenerator()
	apiReq.GroupID = groupID
	apiReq.MemberList = memberList
	var realData []*sdk2.GroupMemberFullInfo
	err := g.p.PostReturn(constant.GetGroupMembersInfoRouter, apiReq, &realData)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return realData, nil
}

func (g *Group) SyncSelfGroupApplication(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := g.getSendGroupApplicationListFromSvr(operationID)
	if err != nil {
		log.NewError(operationID, "getSendGroupApplicationListFromSvr failed ", err.Error())
		return
	}
	onServer := common.TransferToLocalSendGroupRequest(svrList)
	onLocal, err := g.db.GetSendGroupApplication()
	if err != nil {
		log.NewError(operationID, "GetSendGroupApplication failed ", err.Error())
		return
	}

	log.NewInfo(operationID, "svrList onServer onLocal", svrList, onServer, onLocal)
	aInBNot, bInANot, sameA, sameB := common.CheckGroupRequestDiff(onServer, onLocal)
	log.Info(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	for _, index := range aInBNot {
		err := g.db.InsertGroupRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertGroupRequest failed ", err.Error(), *onServer[index])
			continue
		}

		callbackData := *onServer[index]
		g.listener.OnGroupApplicationAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnGroupApplicationAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := g.db.UpdateGroupRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateGroupRequest failed ", err.Error())
			continue
		}

		if onServer[index].HandleResult == constant.GroupResponseRefuse {
			callbackData := *onServer[index]
			g.listener.OnGroupApplicationRejected(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnGroupApplicationRejected", utils.StructToJsonString(callbackData))

		} else if onServer[index].HandleResult == constant.GroupResponseAgree {
			callbackData := *onServer[index]
			g.listener.OnGroupApplicationAccepted(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnGroupApplicationAccepted", utils.StructToJsonString(callbackData))
		} else {
			callbackData := *onServer[index]
			g.listener.OnGroupApplicationAdded(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnGroupApplicationAdded", utils.StructToJsonString(callbackData))

		}
	}
	for _, index := range bInANot {
		err := g.db.DeleteGroupRequest(onLocal[index].GroupID, onLocal[index].UserID)
		if err != nil {
			log.NewError(operationID, "DeleteGroupRequest failed ", err.Error())
			continue
		}
		callbackData := *onLocal[index]
		g.listener.OnGroupApplicationDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnGroupApplicationDeleted", utils.StructToJsonString(callbackData))
	}
}

func (g *Group) SyncGroupApplicationForUserAndGroup(operationID, groupID, userID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID, userID)
	svrList, err := g.getRecvGroupApplicationListForUserFromSvr(operationID, userID)
	if err != nil {
		log.NewError(operationID, "getSendGroupApplicationListFromSvr failed ", err.Error())
		return
	}

	onServer := common.TransferToLocalSendGroupRequest(svrList)
	var localRequst *model_struct.LocalGroupRequest
	for _, request := range onServer {
		if request.UserID == userID && request.GroupID == groupID {
			localRequst = request
			break
		}
	}

	if localRequst == nil {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "localRequst is nil", localRequst)
		return
	}

	database, err := db.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		log.NewError(operationID, "NewDataBase failed ", err.Error())
		return
	}

	onLocal, _ := database.GetGroupApplicationWithUserIDAndGroupID(userID, groupID)
	if onLocal != nil {
		err := database.UpdateGroupRequest(localRequst)
		log.NewError(operationID, "UpdateGroupRequest failed ", err.Error())
		return
	} else {
		err := database.InsertGroupRequest(localRequst)
		log.NewError(operationID, "InsertGroupRequest failed ", err.Error())
		return
	}
}

func (g *Group) SyncGroupApplicationToGroupAdmins(operationID, groupID, userID string) {

	svrList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	if err != nil {
		log.NewError(operationID, "getGroupAllMemberByGroupIDFromSvr failed ", err.Error(), groupID)
		return
	}
	log.NewInfo(operationID, "getGroupAllMemberByGroupIDFromSvr ", svrList)

	request, err := g.db.GetAdminGroupApplicationWithUserAndGroup(groupID, userID)
	if err != nil {
		log.NewError(operationID, "GetAdminGroupApplicationWithUserAndGroup failed ", err.Error(), groupID)
		return
	}

	log.NewInfo(operationID, utils.GetSelfFuncName(), "request: ", request)

	for _, member := range svrList {
		if member.RoleLevel > constant.GroupOrdinaryUsers {
			log.NewInfo(operationID, utils.GetSelfFuncName(), "member.UserID:", member.UserID, member.RoleLevel)
			toUserDB, err := db.NewDataBase(member.UserID, sdk_struct.SvrConf.DataDir)
			if err != nil {
				log.NewError(operationID, "init ToUserDB failed ", member.UserID, err.Error(), *request)
				continue
			} else {
				adminRequest, _ := toUserDB.GetAdminGroupApplicationWithUserAndGroup(groupID, userID)
				if adminRequest != nil {
					err := toUserDB.UpdateAdminGroupRequest(request)
					if err != nil {
						log.NewError(operationID, "UpdateAdminGroupRequest to admin failed ", err.Error(), member.UserID, request)
					}
				} else {
					err := toUserDB.InsertAdminGroupRequest(request)
					if err != nil {
						log.NewError(operationID, "InsertAdminGroupRequest to admin failed ", err.Error(), member.UserID, request)
					}
				}
			}
		}

	}

}

func (g *Group) SyncAdminGroupApplicationForMember(operationID, userID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userID)
	svrList, err := g.getRecvGroupApplicationListForUserFromSvr(operationID, userID)
	if err != nil {
		log.NewError(operationID, "getRecvGroupApplicationListFromSvr failed ", err.Error())
		return
	}
	onServer := common.TransferToLocalAdminGroupRequest(svrList)
	database, err := db.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		log.NewError(operationID, "NewDataBase failed ", err.Error())
		return
	}

	onLocal, err := database.GetAdminGroupApplication()
	if err != nil {
		log.NewError(operationID, "GetAdminGroupApplication failed ", err.Error())
		return
	}

	log.NewInfo(operationID, "svrList onServer onLocal", svrList, onServer, onLocal)
	aInBNot, bInANot, sameA, sameB := common.CheckAdminGroupRequestDiff(onServer, onLocal)
	log.Info(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	for _, index := range aInBNot {
		err := database.InsertAdminGroupRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertGroupRequest failed ", err.Error(), *onServer[index])
			continue
		}
		//callbackData := sdk.GroupApplicationAddedCallback(*onServer[index])
		//g.listener.OnGroupApplicationAdded(utils.StructToJsonString(callbackData))
		//log.Info(operationID, "OnReceiveJoinGroupApplicationAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := database.UpdateAdminGroupRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateGroupRequest failed ", err.Error())
			continue
		}
		//if onServer[index].HandleResult == constant.GroupResponseRefuse {
		//	callbackData := sdk.GroupApplicationRejectCallback(*onServer[index])
		//	g.listener.OnGroupApplicationRejected(utils.StructToJsonString(callbackData))
		//	log.Info(operationID, "OnGroupApplicationRejected", utils.StructToJsonString(callbackData))
		//
		//} else if onServer[index].HandleResult == constant.GroupResponseAgree {
		//	callbackData := sdk.GroupApplicationAcceptCallback(*onServer[index])
		//	g.listener.OnGroupApplicationAccepted(utils.StructToJsonString(callbackData))
		//	log.Info(operationID, "OnGroupApplicationAccepted", utils.StructToJsonString(callbackData))
		//} else {
		//	callbackData := sdk.GroupApplicationAcceptCallback(*onServer[index])
		//	g.listener.OnGroupApplicationAdded(utils.StructToJsonString(callbackData))
		//	log.Info(operationID, "OnReceiveJoinGroupApplicationAdded", utils.StructToJsonString(callbackData))
		//}
	}
	for _, index := range bInANot {
		err := database.DeleteAdminGroupRequest(onLocal[index].GroupID, onLocal[index].UserID)
		if err != nil {
			log.NewError(operationID, "DeleteGroupRequest failed ", err.Error())
			continue
		}
		//callbackData := sdk.GroupApplicationDeletedCallback(*onLocal[index])
		//g.listener.OnGroupApplicationDeleted(utils.StructToJsonString(callbackData))
		//log.Info(operationID, "OnReceiveJoinGroupApplicationDeleted", utils.StructToJsonString(callbackData))
	}
}

func (g *Group) SyncAdminGroupApplication(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := g.getRecvGroupApplicationListFromSvr(operationID)
	if err != nil {
		log.NewError(operationID, "getRecvGroupApplicationListFromSvr failed ", err.Error())
		return
	}
	onServer := common.TransferToLocalAdminGroupRequest(svrList)
	onLocal, err := g.db.GetAdminGroupApplication()
	if err != nil {
		log.NewError(operationID, "GetAdminGroupApplication failed ", err.Error())
		return
	}

	log.NewInfo(operationID, "svrList onServer onLocal", svrList, onServer, onLocal)
	aInBNot, bInANot, sameA, sameB := common.CheckAdminGroupRequestDiff(onServer, onLocal)
	log.Info(operationID, "diff ", aInBNot, bInANot, sameA, sameB)
	for _, index := range aInBNot {
		err := g.db.InsertAdminGroupRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertGroupRequest failed ", err.Error(), *onServer[index])
			continue
		}
		callbackData := sdk.GroupApplicationAddedCallback(*onServer[index])
		g.listener.OnGroupApplicationAdded(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnReceiveJoinGroupApplicationAdded", utils.StructToJsonString(callbackData))
	}
	for _, index := range sameA {
		err := g.db.UpdateAdminGroupRequest(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateGroupRequest failed ", err.Error())
			continue
		}
		if onServer[index].HandleResult == constant.GroupResponseRefuse {
			callbackData := sdk.GroupApplicationRejectCallback(*onServer[index])
			g.listener.OnGroupApplicationRejected(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnGroupApplicationRejected", utils.StructToJsonString(callbackData))

		} else if onServer[index].HandleResult == constant.GroupResponseAgree {
			callbackData := sdk.GroupApplicationAcceptCallback(*onServer[index])
			g.listener.OnGroupApplicationAccepted(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnGroupApplicationAccepted", utils.StructToJsonString(callbackData))
		} else {
			callbackData := sdk.GroupApplicationAcceptCallback(*onServer[index])
			g.listener.OnGroupApplicationAdded(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnReceiveJoinGroupApplicationAdded", utils.StructToJsonString(callbackData))
		}
	}
	for _, index := range bInANot {
		err := g.db.DeleteAdminGroupRequest(onLocal[index].GroupID, onLocal[index].UserID)
		if err != nil {
			log.NewError(operationID, "DeleteGroupRequest failed ", err.Error())
			continue
		}
		callbackData := sdk.GroupApplicationDeletedCallback(*onLocal[index])
		g.listener.OnGroupApplicationDeleted(utils.StructToJsonString(callbackData))
		log.Info(operationID, "OnReceiveJoinGroupApplicationDeleted", utils.StructToJsonString(callbackData))
	}
}

//	func transferGroupInfo(input []*api.GroupInfo) []*api.GroupInfo{
//		var result []*api.GroupInfo
//		for _, v := range input {
//			t := &api.GroupInfo{}
//			copier.Copy(t, &v)
//			if v.NeedVerification != nil {
//				t.NeedVerification = v.NeedVerification.Value
//			}
//			result = append(result, t)
//		}
//		return result
//	}

func (g *Group) SyncJoinedGroupList(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := g.getJoinedGroupListFromSvr(operationID)
	log.Info(operationID, "getJoinedGroupListFromSvr", svrList, g.loginUserID)
	if err != nil {
		log.NewError(operationID, "getJoinedGroupListFromSvr failed ", err.Error())
		return
	}
	onServer := common.TransferToLocalGroupInfo(svrList)
	onLocal, err := g.db.GetJoinedGroupList()
	if err != nil {
		log.NewError(operationID, "GetJoinedGroupList failed ", err.Error())
		return
	}

	log.NewInfo(operationID, " onLocal ", onLocal, g.loginUserID)
	aInBNot, bInANot, sameA, sameB := common.CheckGroupInfoDiff(onServer, onLocal)
	log.Info(operationID, "diff ", aInBNot, bInANot, sameA, sameB, g.loginUserID)
	var isReadDiffusion bool
	for _, index := range aInBNot {
		fallThrough := false
		err := g.db.InsertGroup(onServer[index])
		if err != nil {
			log.NewError(operationID, "InsertGroup failed ", err.Error(), *onServer[index])
			fallThrough = true
		}
		//todo update conversation if already exists
		conversationID := utils.GetConversationIDBySessionType(onServer[index].GroupID, constant.GroupChatType)
		localConv, err := g.db.GetConversation(conversationID)
		if err == nil && localConv != nil {
			if localConv.GroupID != "" {
				paramsGetMaxSeq := api2.ParamsGetMaxSeq{}
				paramsGetMaxSeq.OperationID = operationID
				paramsGetMaxSeq.OpUserID = g.loginUserID
				paramsGetMaxSeq.GroupID = localConv.GroupID

				response := map[string]interface{}{}
				response["maxSeq"] = 0
				response[localConv.GroupID] = 0
				g.p.PostReturn(constant.GetGroupMinSeqRouter, paramsGetMaxSeq, &response)

				if seqNumberMaxInterf, ok := response[localConv.GroupID]; ok {
					seqNumberMax := seqNumberMaxInterf.(float64)
					localSeqSynced := &model_struct.LocalSeqSynced{}
					localSeqSynced.ConversationID = utils.GetConversationIDBySessionType(localConv.GroupID, constant.GroupChatType)
					localSeqSynced.Seq = uint32(seqNumberMax)
					if localSeqSynced.Seq > 0 {
						g.db.UpdateLocalSeqNumber(localSeqSynced)
					}
				}

			}

			localConv.ShowName = onServer[index].GroupName
			localConv.FaceURL = onServer[index].FaceURL
			_ = g.db.UpdateConversationForGroupIfExists(localConv)
		}
		if fallThrough {
			continue
		}

		callbackData := sdk.JoinedGroupAddedCallback(*onServer[index])
		if (*onServer[index]).GroupType == int32(constant.WorkingGroup) {
			isReadDiffusion = true
		}

		if g.listener != nil {
			g.listener.OnJoinedGroupAdded(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnJoinedGroupAdded", utils.StructToJsonString(callbackData))
		}
	}
	for _, index := range sameA {
		err := g.db.UpdateGroup(onServer[index])
		if err != nil {
			log.NewError(operationID, "UpdateGroup failed ", err.Error(), onServer[index])
			continue
		}
		//todo update conversation if already exists
		conversationID := utils.GetConversationIDBySessionType(onServer[index].GroupID, constant.GroupChatType)
		localConv, err := g.db.GetConversation(conversationID)
		if err == nil && localConv != nil {
			localConv.ShowName = onServer[index].GroupName
			localConv.FaceURL = onServer[index].FaceURL
			_ = g.db.UpdateConversationForGroupIfExists(localConv)

			//if localConv.GroupID != "" {
			//	paramsGetMaxSeq := api2.ParamsGetMaxSeq{}
			//	paramsGetMaxSeq.OperationID = operationID
			//	paramsGetMaxSeq.OpUserID = g.loginUserID
			//	paramsGetMaxSeq.GroupID = localConv.GroupID
			//
			//	response := map[string]interface{}{}
			//	response["maxSeq"] = 0
			//	response[localConv.GroupID] = 0
			//	g.p.PostReturn(constant.GetMaxSeqRouter, paramsGetMaxSeq, &response)
			//
			//	if seqNumberMaxInterf, ok := response[localConv.GroupID]; ok {
			//		seqNumberMax := seqNumberMaxInterf.(float64)
			//		localSeqSynced := &model_struct.LocalSeqSynced{}
			//		localSeqSynced.ConversationID = utils.GetConversationIDBySessionType(localConv.GroupID, constant.GroupChatType)
			//		localSeqSynced.Seq = uint32(seqNumberMax)
			//		if localSeqSynced.Seq > 0 {
			//			g.db.UpdateLocalSeqNumber(localSeqSynced)
			//		}
			//	}
			//
			//}
		}
		callbackData := sdk.GroupInfoChangedCallback(*onServer[index])
		if g.listener != nil {
			g.listener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnGroupInfoChanged", utils.StructToJsonString(callbackData))
		}

	}

	for _, index := range bInANot {
		log.Info(operationID, "DeleteGroup: ", onLocal[index].GroupID, g.loginUserID)
		err := g.db.DeleteGroup(onLocal[index].GroupID)
		if err != nil {
			log.NewError(operationID, "DeleteGroup failed ", err.Error(), onLocal[index].GroupID)
			continue
		}
		if (*onLocal[index]).GroupType == int32(constant.WorkingGroup) {
			isReadDiffusion = true
		}
		g.db.DeleteGroupAllMembers(onLocal[index].GroupID)
		localGroup := onLocal[index]
		localGroup.Status = constant.GroupStatusDismissed
		callbackData := sdk.JoinedGroupDeletedCallback(*localGroup)
		if g.listener != nil {
			g.listener.OnJoinedGroupDeleted(utils.StructToJsonString(callbackData))
			log.Info(operationID, "OnJoinedGroupDeleted", utils.StructToJsonString(callbackData))
		}
	}
	if isReadDiffusion {
		cmd := sdk_struct.CmdJoinedSuperGroup{OperationID: operationID}
		err := common.TriggerCmdJoinedSuperGroup(cmd, g.joinedSuperGroupCh)
		if err != nil {
			log.Error(operationID, "TriggerCmdJoinedSuperGroup failed ", err.Error())
		}
		err = common.TriggerCmdWakeUp(g.heartbeatCmdCh)
		if err != nil {
			log.Error(operationID, "TriggerCmdWakeUp failed ", err.Error())
		}
	}
}

func (g *Group) SyncJoinedGroupForAll(operationID, groupID string) {

	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID)
	svrList, err := g.getGroupsInfoFromSvr([]string{groupID}, operationID)
	log.Info(operationID, "getJoinedGroupListForUserIDFromSvr", svrList)
	if err != nil {
		log.NewError(operationID, "getJoinedGroupListForUserIDFromSvr failed ", err.Error())
		return
	}

	onServer := common.TransferToLocalGroupInfo(svrList)
	memberList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	if err != nil {
		log.NewError(operationID, "getGroupAllMemberByGroupIDFromSvr failed ", err.Error())
		return
	}

	for _, member := range memberList {
		database, err := db.NewDataBase(member.UserID, sdk_struct.SvrConf.DataDir)
		if err != nil {
			log.NewError(operationID, "NewDataBase failed ", err.Error())
			return
		}

		onLocal, _ := database.GetGroupInfoByGroupID(groupID)
		if onLocal != nil && onLocal.GroupID != "" {
			err := database.UpdateGroup(onServer[0])
			if err != nil {
				log.NewError(operationID, "UpdateGroup failed ", err.Error(), onServer[0])
				return
			}
		} else {
			err := database.InsertGroup(onServer[0])
			if err != nil {
				log.NewError(operationID, "InsertGroup failed ", err.Error(), *onServer[0])
				return
			}
		}
	}

}

func (g *Group) SyncJoinedGroupForUser(operationID, groupID, userID string) {

	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	svrList, err := g.getGroupsInfoFromSvr([]string{groupID}, operationID)
	log.Info(operationID, "getJoinedGroupListForUserIDFromSvr", svrList, userID)
	if err != nil {
		log.NewError(operationID, "getJoinedGroupListForUserIDFromSvr failed ", err.Error())
		return
	}

	onServer := common.TransferToLocalGroupInfo(svrList)
	database, err := db.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		log.NewError(operationID, "NewDataBase failed ", err.Error())
		return
	}

	onLocal, _ := database.GetGroupInfoByGroupID(groupID)
	if onLocal != nil && onLocal.GroupID != "" {
		err := database.UpdateGroup(onServer[0])
		if err != nil {
			log.NewError(operationID, "UpdateGroup failed ", err.Error(), onServer[0])
			return
		}
	} else {
		err := database.InsertGroup(onServer[0])
		if err != nil {
			log.NewError(operationID, "InsertGroup failed ", err.Error(), *onServer[0])
			return
		}
	}

}

func (g *Group) syncGroupMemberByGroupIDForAll(groupID, operationID string) {
	//log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID)
	//svrList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	//if err != nil {
	//	log.NewError(operationID, "getGroupAllMemberByGroupIDFromSvr failed ", err.Error(), groupID)
	//	return
	//}
	//log.NewInfo(operationID, "getGroupAllMemberByGroupIDFromSvr ", svrList)
	//onServer := common.TransferToLocalGroupMember(svrList)
	//
	//for _, member := range svrList {
	//	database, err := db.NewDataBase(member.UserID, sdk_struct.SvrConf.DataDir)
	//	if err != nil {
	//		log.NewError(operationID, "NewDataBase failed ", err.Error(), groupID, member.UserID)
	//		return
	//	}
	//
	//	onLocal, err := database.GetGroupMemberListByGroupID(groupID)
	//	if err != nil {
	//		log.NewError(operationID, "GetGroupMemberListByGroupID failed ", err.Error(), groupID, member.UserID)
	//		return
	//	}
	//	//log.NewInfo(operationID, "svrList onServer onLocal", svrList, onServer, onLocal)
	//	aInBNot, bInANot, sameA, _ := common.CheckGroupMemberDiff(onServer, onLocal)
	//	log.Info(operationID, "getGroupAllMemberByGroupIDFromSvr  diff ", aInBNot, bInANot, sameA)
	//	for _, index := range aInBNot {
	//		log.NewInfo(operationID, utils.GetSelfFuncName(), "aInBNot:", onServer[index])
	//		err := database.InsertGroupMember(onServer[index])
	//		if err != nil {
	//			log.NewError(operationID, "InsertGroupMember failed ", err.Error(), *onServer[index])
	//			continue
	//		}
	//	}
	//	for _, index := range sameA {
	//		err := database.UpdateGroupMember(onServer[index])
	//		if err != nil {
	//			log.NewError(operationID, "UpdateGroupMember failed ", err.Error(), *onServer[index])
	//			continue
	//		}
	//	}
	//	for _, index := range bInANot {
	//		err := database.DeleteGroupMember(onLocal[index].GroupID, onLocal[index].UserID)
	//		if err != nil {
	//			log.NewError(operationID, "DeleteGroupMember failed ", err.Error(), onLocal[index].GroupID, onLocal[index].UserID)
	//			continue
	//		}
	//	}
	//}
}

func (g *Group) deleteJoinedGroupForUser(operationID, groupID, userID string) {

	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID, userID)

	database, err := db.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		log.NewError(operationID, "NewDataBase failed ", err.Error())
		return
	}

	err = database.DeleteGroup(groupID)
	if err != nil {
		log.NewError(operationID, "DeleteGroup failed ", err.Error(), userID, groupID)
		return
	}

}

func (g *Group) SyncGroupMemberByGroupID(groupID string, operationID string, onGroupMemberNotification bool) {
	g.newSyncGroupMemberByGroupID(groupID, operationID, onGroupMemberNotification)
}

func (g *Group) SyncMySelfInTheGroup(groupID string, operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args:", groupID)
	syncMemberList, err := g.getGroupMembersInfoFromSvr(groupID, []string{g.loginUserID})
	if err != nil {
		log.Error(operationID, "SyncGroupMemberInfoForAllMembers failed ", err.Error())
		return
	}

	if len(syncMemberList) > 0 {
		//log.NewError(operationID, utils.GetSelfFuncName(), "MuteEndTime:", syncMemberList[0].MuteEndTime)
		localInfo := common.TransferToLocalGroupMember(syncMemberList)
		if localInfo != nil && len(localInfo) > 0 {
			myInfo := localInfo[0]
			_, err := g.db.GetGroupMemberInfoByGroupIDUserID(myInfo.GroupID, myInfo.UserID)
			if err != nil {
				err = g.db.InsertGroupMember(myInfo)
				if err != nil {
					log.Error(operationID, "SyncMySelfInTheGroup failed ", err.Error())
				}
			} else {
				err = g.db.UpdateGroupMember(myInfo)
				if err != nil {
					log.Error(operationID, "SyncMySelfInTheGroup failed ", err.Error())
				}
			}

		}
	}
}

func (g *Group) syncGroupMemberByGroupID(groupID string, operationID string, onGroupMemberNotification bool) {
	g.syncGroupUpdatesVersionByGroupIDV2(groupID, operationID)
	//log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID)
	//svrList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	//if err != nil {
	//	log.NewError(operationID, "getGroupAllMemberByGroupIDFromSvr failed ", err.Error(), groupID)
	//	//return
	//}
	//log.NewInfo(operationID, "getGroupAllMemberByGroupIDFromSvr ", svrList)
	//onServer := common.TransferToLocalGroupMember(svrList)
	//onLocal, err := g.db.GetGroupMemberListByGroupID(groupID)
	//if err != nil {
	//	log.NewError(operationID, "GetGroupMemberListByGroupID failed ", err.Error(), groupID)
	//	return
	//}
	////log.NewInfo(operationID, "svrList onServer onLocal", svrList, onServer, onLocal)
	//aInBNot, bInANot, sameA, _ := common.CheckGroupMemberDiff(onServer, onLocal)
	//log.Info(operationID, "getGroupAllMemberByGroupIDFromSvr  diff ", aInBNot, bInANot, sameA)
	//var insertGroupMemberList []*model_struct.LocalGroupMember
	//for _, index := range aInBNot {
	//	if onGroupMemberNotification == false {
	//		insertGroupMemberList = append(insertGroupMemberList, onServer[index])
	//		continue
	//	}
	//	log.NewInfo(operationID, utils.GetSelfFuncName(), "aInBNot:", onServer[index])
	//	err := g.db.InsertGroupMember(onServer[index])
	//	if err != nil {
	//		log.NewError(operationID, "InsertGroupMember failed ", err.Error(), *onServer[index])
	//		continue
	//	}
	//	if onGroupMemberNotification == true {
	//		callbackData := sdk.GroupMemberAddedCallback(*onServer[index])
	//		g.listener.OnGroupMemberAdded(utils.StructToJsonString(callbackData))
	//		log.Debug(operationID, "OnGroupMemberAdded", utils.StructToJsonString(callbackData))
	//	}
	//}
	//if len(insertGroupMemberList) > 0 {
	//	split := 1000
	//	idx := 0
	//	remain := len(insertGroupMemberList) % split
	//	log.Warn(operationID, "BatchInsertGroupMember all: ", len(insertGroupMemberList))
	//	for idx = 0; idx < len(insertGroupMemberList)/split; idx++ {
	//		sub := insertGroupMemberList[idx*split : (idx+1)*split]
	//		err = g.db.BatchInsertGroupMember(sub)
	//		log.Warn(operationID, "BatchInsertGroupMember len: ", len(sub))
	//		if err != nil {
	//			log.Error(operationID, "BatchInsertGroupMember failed ", err.Error(), len(sub))
	//			for again := 0; again < len(sub); again++ {
	//				if err = g.db.InsertGroupMember(sub[again]); err != nil {
	//					log.Error(operationID, "InsertGroupMember failed ", err.Error(), sub[again])
	//				}
	//			}
	//		}
	//	}
	//	if remain > 0 {
	//		sub := insertGroupMemberList[idx*split:]
	//		log.Warn(operationID, "BatchInsertGroupMember len: ", len(sub))
	//		if err != nil {
	//			log.Error(operationID, "BatchInsertGroupMember failed ", err.Error(), len(sub))
	//			for again := 0; again < len(sub); again++ {
	//				if err = g.db.InsertGroupMember(sub[again]); err != nil {
	//					log.Error(operationID, "InsertGroupMember failed ", err.Error(), sub[again])
	//				}
	//			}
	//		}
	//	}
	//}
	//
	//for _, index := range sameA {
	//	err := g.db.UpdateGroupMember(onServer[index])
	//	if err != nil {
	//		log.NewError(operationID, "UpdateGroupMember failed ", err.Error(), *onServer[index])
	//		continue
	//	}
	//
	//	callbackData := sdk.GroupMemberInfoChangedCallback(*onServer[index])
	//	g.listener.OnGroupMemberInfoChanged(utils.StructToJsonString(callbackData))
	//	log.Info(operationID, "OnGroupMemberInfoChanged", utils.StructToJsonString(callbackData))
	//}
	//for _, index := range bInANot {
	//	err := g.db.DeleteGroupMember(onLocal[index].GroupID, onLocal[index].UserID)
	//	if err != nil {
	//		log.NewError(operationID, "DeleteGroupMember failed ", err.Error(), onLocal[index].GroupID, onLocal[index].UserID)
	//		continue
	//	}
	//	if onGroupMemberNotification == true {
	//		callbackData := sdk.GroupMemberDeletedCallback(*onLocal[index])
	//		g.listener.OnGroupMemberDeleted(utils.StructToJsonString(callbackData))
	//		log.Info(operationID, "OnGroupMemberDeleted", utils.StructToJsonString(callbackData))
	//	}
	//}
}

func (g *Group) newSyncGroupMemberByGroupID(groupID string, operationID string, onGroupMemberNotification bool) {
	g.syncGroupUpdatesVersionByGroupIDV2(groupID, operationID)
	//log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", groupID)
	//svrList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	//if err != nil {
	//	svrList = []*sdk2.GroupMemberFullInfo{}
	//	log.NewError(operationID, "getGroupAllMemberByGroupIDFromSvr failed ", err.Error(), groupID)
	//	//return
	//}
	//log.NewInfo(operationID, "getGroupAllMemberByGroupIDFromSvr ", svrList)
	//
	//memberIDListSvr := []string{}
	//onServer := common.TransferToLocalGroupMember(svrList)
	//
	//for _, member := range onServer {
	//	memberIDListSvr = append(memberIDListSvr, member.UserID)
	//	tmpMember, err := g.db.GetGroupMemberInfoByGroupIDUserID(groupID, member.UserID)
	//	if err == nil && tmpMember.UserID != "" {
	//		//update member
	//		if !reflect.DeepEqual(tmpMember, member) {
	//			err := g.db.UpdateGroupMember(member)
	//			if err != nil {
	//				log.NewError(operationID, "UpdateGroupMember failed ", err.Error(), *member)
	//				continue
	//			}
	//			if onGroupMemberNotification == true {
	//				callbackData := sdk.GroupMemberInfoChangedCallback(*member)
	//				g.listener.OnGroupMemberInfoChanged(utils.StructToJsonString(callbackData))
	//				log.Info(operationID, "OnGroupMemberInfoChanged", utils.StructToJsonString(callbackData))
	//			}
	//		}
	//
	//	} else {
	//		//insert member
	//		err := g.db.InsertGroupMember(member)
	//		if err != nil {
	//			log.NewError(operationID, "InsertGroupMember failed ", err.Error(), *member)
	//			continue
	//		}
	//		if onGroupMemberNotification == true {
	//			callbackData := sdk.GroupMemberAddedCallback(*member)
	//			g.listener.OnGroupMemberAdded(utils.StructToJsonString(callbackData))
	//			log.Debug(operationID, "OnGroupMemberAdded", utils.StructToJsonString(callbackData))
	//		}
	//	}
	//}
	//
	////delete member not in the group
	//onLocal, err := g.db.GetGroupMemberListByGroupID(groupID)
	//if err != nil {
	//	log.NewError(operationID, "GetGroupMemberListByGroupID failed ", err.Error(), groupID)
	//	return
	//}
	//for _, member := range onLocal {
	//	if !utils.IsContain(member.UserID, memberIDListSvr) {
	//		err = g.db.DeleteGroupMember(groupID, member.UserID)
	//		if err != nil {
	//			log.NewError(operationID, "DeleteGroupMember failed ", err.Error(), member.GroupID, member.UserID)
	//			continue
	//		}
	//		if onGroupMemberNotification == true {
	//			callbackData := sdk.GroupMemberDeletedCallback(*member)
	//			g.listener.OnGroupMemberDeleted(utils.StructToJsonString(callbackData))
	//			log.Info(operationID, "OnGroupMemberDeleted", utils.StructToJsonString(callbackData))
	//		}
	//	}
	//}
}

func (g *Group) SyncGroupMemberInfoForMembers(operationID, syncUserID, groupID string, members []string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args:", syncUserID, groupID, members)
	syncMemberList, err := g.getGroupMembersInfoFromSvr(groupID, []string{syncUserID})
	if err != nil {
		log.Error(operationID, "SyncGroupMemberInfoForAllMembers failed ", err.Error())
		return
	}

	if len(syncMemberList) > 0 {
		log.NewError(operationID, utils.GetSelfFuncName(), "MuteEndTime:", syncMemberList[0].MuteEndTime)
		localInfo := common.TransferToLocalGroupMember(syncMemberList)

		for _, member := range members {
			database, err := db.NewDataBase(member, sdk_struct.SvrConf.DataDir)
			if err != nil {
				log.Error(operationID, "newdatabase failed ", member, err.Error())
				continue
			}

			err = database.UpdateGroupMember(localInfo[0])
			if err != nil {
				log.Error(operationID, "UpdateGroupMember failed ", localInfo[0].UserID, err.Error())
				continue
			}
		}
	}

}

func (g *Group) SyncGroupMemberInfoForOwnerAndAdmins(operationID, syncUserID, groupID string) {
	memberList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	if err != nil {
		log.Error(operationID, "SyncGroupMemberInfoForAllMembers failed ", err.Error())
		return
	}

	syncMemberList, err := g.getGroupMembersInfoFromSvr(groupID, []string{syncUserID})
	if err != nil {
		log.Error(operationID, "SyncGroupMemberInfoForAllMembers failed ", err.Error())
		return
	}

	if len(syncMemberList) > 0 {
		localInfo := common.TransferToLocalGroupMember(syncMemberList)

		for _, member := range memberList {
			if member.RoleLevel > 1 {
				database, err := db.NewDataBase(member.UserID, sdk_struct.SvrConf.DataDir)
				if err != nil {
					log.Error(operationID, "newdatabase failed ", member.UserID, err.Error())
					continue
				}

				err = database.UpdateGroupMember(localInfo[0])
				if err != nil {
					log.Error(operationID, "UpdateGroupMember failed ", localInfo[0].UserID, err.Error())
					continue
				}
			}
		}
	}

}

func (g *Group) SyncGroupMemberInfoForAllMembers(operationID, syncUserID, groupID string) {
	memberList, err := g.getGroupAllMemberByGroupIDFromSvr(groupID, operationID)
	if err != nil {
		log.Error(operationID, "SyncGroupMemberInfoForAllMembers failed ", err.Error())
		return
	}

	syncMemberList, err := g.getGroupMembersInfoFromSvr(groupID, []string{syncUserID})
	if err != nil {
		log.Error(operationID, "SyncGroupMemberInfoForAllMembers failed ", err.Error())
		return
	}

	if len(syncMemberList) > 0 {
		localInfo := common.TransferToLocalGroupMember(syncMemberList)

		for _, member := range memberList {
			database, err := db.NewDataBase(member.UserID, sdk_struct.SvrConf.DataDir)
			if err != nil {
				log.Error(operationID, "newdatabase failed ", member.UserID, err.Error())
				continue
			}

			err = database.UpdateGroupMember(localInfo[0])
			if err != nil {
				log.Error(operationID, "UpdateGroupMember failed ", localInfo[0].UserID, err.Error())
				continue
			}
		}
	}

}

// syncType: 1 kicked, 2 invited, 3 updated
func (g *Group) SyncOneGroupMemberInfo(operationID, userID, groupID string, syncType int) {

	switch syncType {
	case 1:
		//delete group member
		err := g.db.DeleteGroupMember(groupID, userID)
		if err != nil {
			log.Error(operationID, "DeleteGroupMember failed ", err.Error())
		}
	case 2:
		//insert group member
		syncMemberList, err := g.getGroupMembersInfoFromSvr(groupID, []string{userID})
		if err != nil {
			log.Error(operationID, "getGroupMembersInfoFromSvr failed ", err.Error())
			return
		}

		if len(syncMemberList) > 0 {
			localInfo := common.TransferToLocalGroupMember(syncMemberList)

			err = g.db.InsertGroupMember(localInfo[0])
			if err != nil {
				log.Error(operationID, "InsertGroupMember failed ", localInfo[0].UserID, err.Error())
			}
		}
	case 3:
		//update group member
		syncMemberList, err := g.getGroupMembersInfoFromSvr(groupID, []string{userID})
		if err != nil {
			log.Error(operationID, "getGroupMembersInfoFromSvr failed ", err.Error())
			return
		}

		if len(syncMemberList) > 0 {
			localInfo := common.TransferToLocalGroupMember(syncMemberList)

			err = g.db.UpdateGroupMember(localInfo[0])
			if err != nil {
				log.Error(operationID, "UpdateGroupMember failed ", localInfo[0].UserID, err.Error())
			}
		}
	}
}

func (g *Group) SyncJoinedGroupMember(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	groupListOnServer, err := g.getJoinedGroupListFromSvr(operationID)
	if err != nil {
		log.Error(operationID, "getJoinedGroupListFromSvr failed ", err.Error())
		return
	}
	var wg sync.WaitGroup
	if len(groupListOnServer) == 0 {
		return
	}
	wg.Add(len(groupListOnServer))
	log.Info(operationID, "syncGroupMemberByGroupID begin", len(groupListOnServer))
	for _, v := range groupListOnServer {
		go func(groupID, operationID string) {
			g.syncGroupMemberByGroupID(groupID, operationID, true)
			wg.Done()
		}(v.GroupID, operationID)
	}

	wg.Wait()
	log.Info(operationID, "syncGroupMemberByGroupID end")
}

func (g *Group) SyncMyInfoInAllGroupForFirstLogin(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	groupListOnServer, err := g.getJoinedGroupListFromSvr(operationID)
	if err != nil {
		log.Error(operationID, "getJoinedGroupListFromSvr failed ", err.Error())
		return
	}
	var wg sync.WaitGroup
	if len(groupListOnServer) == 0 {
		return
	}
	wg.Add(len(groupListOnServer))
	log.Info(operationID, "syncGroupMemberByGroupID begin", len(groupListOnServer))
	for _, v := range groupListOnServer {
		go func(groupID, operationID string) {
			g.SyncMySelfInTheGroup(groupID, operationID)
			wg.Done()
		}(v.GroupID, operationID)
	}

	wg.Wait()
	log.Info(operationID, "syncGroupMemberByGroupID end")
}

func (g *Group) SyncJoinedGroupMemberForFirstLogin(operationID string) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ")
	groupListOnServer, err := g.getJoinedGroupListFromSvr(operationID)
	if err != nil {
		log.Error(operationID, "getJoinedGroupListFromSvr failed ", err.Error())
		return
	}
	var wg sync.WaitGroup
	if len(groupListOnServer) == 0 {
		return
	}
	wg.Add(len(groupListOnServer))
	log.Info(operationID, "syncGroupMemberByGroupID begin", len(groupListOnServer))
	for _, v := range groupListOnServer {
		go func(groupID, operationID string) {
			g.syncGroupMemberByGroupID(groupID, operationID, true)
			wg.Done()
		}(v.GroupID, operationID)
	}

	wg.Wait()
	log.Info(operationID, "syncGroupMemberByGroupID end")
}

func (g *Group) getGroupAllMemberByGroupIDFromSvr(groupID string, operationID string) ([]*sdk2.GroupMemberFullInfo, error) {
	var apiReq api.GetGroupAllMemberReq
	apiReq.OperationID = operationID
	apiReq.GroupID = groupID
	var realData []*sdk2.GroupMemberFullInfo
	err := g.p.PostReturn(constant.GetGroupAllMemberListRouter, apiReq, &realData)
	if err != nil {
		return nil, utils.Wrap(err, apiReq.OperationID)
	}
	return realData, nil
}

func (g *Group) setGroupMemberNickname(callback open_im_sdk_callback.Base, groupID, userID string, GroupMemberNickname string, operationID string) {
	var apiReq api.SetGroupMemberNicknameReq
	apiReq.OperationID = operationID
	apiReq.GroupID = groupID
	apiReq.UserID = userID
	apiReq.Nickname = GroupMemberNickname
	g.p.PostFatalCallback(callback, constant.SetGroupMemberNicknameRouter, apiReq, nil, apiReq.OperationID)
	g.syncGroupMemberByGroupID(groupID, operationID, true)
}

func (g *Group) searchGroupMembers(callback open_im_sdk_callback.Base, searchParam sdk.SearchGroupMembersParam, operationID string) sdk.SearchGroupMembersCallback {
	if len(searchParam.KeywordList) == 0 {
		log.Error(operationID, "len keywordList == 0")
		common.CheckArgsErrCallback(callback, errors.New("no keyword"), operationID)
	}
	members, err := g.db.SearchGroupMembers(searchParam.KeywordList[0], searchParam.GroupID, searchParam.IsSearchMemberNickname, searchParam.IsSearchUserID, searchParam.Offset, searchParam.Count)
	common.CheckDBErrCallback(callback, err, operationID)
	return members
}

//func (g *Group) syncGroupUpdatesVerion(operationID string) {
//	onLocal, err := g.db.GetJoinedGroupList()
//	if err != nil {
//		log.NewError(operationID, "GetJoinedGroupList failed ", err.Error())
//		return
//	}
//	var usrGroupIds []string
//	for _, groupLocal := range onLocal {
//		usrGroupIds = append(usrGroupIds, groupLocal.GroupID)
//	}
//	groupUpdateVersions, err := g.getGroupUpdatesVersionsFromSrv(operationID, usrGroupIds)
//	if err == nil && groupUpdateVersions != nil {
//		for _, groupUpdateVersionServ := range groupUpdateVersions.GroupUpdateVersionByID {
//			if groupUpdateVersionServ != nil {
//				log.Info(operationID, "SyncGroupUpdatesVerion", groupUpdateVersionServ.GroupIDs, "Recived from Server")
//				groupID := groupUpdateVersionServ.GroupIDs
//				versionNumber := groupUpdateVersionServ.VersionNumber
//				gUVonLocal, _ := g.db.GetGroupUpdateVersionByID(groupID)
//				log.Info(operationID, "SyncGroupUpdatesVerion", groupID, versionNumber, gUVonLocal.GroupID, gUVonLocal.VersionNumber, "Recived from Server Values")
//				if gUVonLocal.VersionNumber < versionNumber {
//					g.SyncGroupMemberByGroupID(groupID, operationID, false)
//					g.db.UpdateGroupUpdateVersionByID(groupID, versionNumber)
//				}
//				utils.RemoveStringFromSlice(usrGroupIds, groupID)
//			}
//		}
//	}
//
//	for _, localGroupID_WO_versionSrv := range usrGroupIds {
//		log.Info(operationID, "SyncGroupUpdatesVerion", localGroupID_WO_versionSrv, "Syncing to server")
//		g.SyncGroupMemberByGroupID(localGroupID_WO_versionSrv, operationID, false)
//		g.db.UpdateGroupUpdateVersionByID(localGroupID_WO_versionSrv, 0)
//	}
//}

//func (g *Group) syncGroupUpdatesVerionByGroupID(groupID, operationID string) {
//
//	var usrGroupIds []string
//	usrGroupIds = append(usrGroupIds, groupID)
//	groupUpdateVersions, err := g.getGroupUpdatesVersionsFromSrv(operationID, usrGroupIds)
//	if err == nil && groupUpdateVersions != nil {
//		for _, groupUpdateVersionServ := range groupUpdateVersions.GroupUpdateVersionByID {
//			if groupUpdateVersionServ != nil && groupUpdateVersionServ.GroupIDs == groupID {
//				log.Info(operationID, "SyncGroupUpdatesVerion", groupUpdateVersionServ.GroupIDs, "Received from Server")
//				groupID := groupUpdateVersionServ.GroupIDs
//				versionNumber := groupUpdateVersionServ.VersionNumber
//				gUVonLocal, _ := g.db.GetGroupUpdateVersionByID(groupID)
//				log.Info(operationID, "SyncGroupUpdatesVerion", groupID, versionNumber, gUVonLocal.GroupID, gUVonLocal.VersionNumber, "Received from Server Values")
//				if gUVonLocal.VersionNumber < versionNumber {
//					g.SyncGroupMemberByGroupID(groupID, operationID, false)
//					localGroup, err := g.db.GetGroupInfoByGroupID(groupID)
//					if err != nil {
//						localGroup.GroupID = groupID
//					}
//					g.listener.OnGroupMembersSyncedStarted(utils.StructToJsonString(localGroup))
//					g.db.UpdateGroupUpdateVersionByID(groupID, versionNumber)
//				}
//				utils.RemoveStringFromSlice(usrGroupIds, groupID)
//			}
//		}
//	}
//	for _, localGroupID_WO_versionSrv := range usrGroupIds {
//		log.Info(operationID, "SyncGroupUpdatesVerion", localGroupID_WO_versionSrv, "Syncing to server")
//		gUVonLocal, _ := g.db.GetGroupUpdateVersionByID(groupID)
//		if gUVonLocal.VersionNumber < 0 {
//			g.SyncGroupMemberByGroupID(localGroupID_WO_versionSrv, operationID, true)
//			g.db.UpdateGroupUpdateVersionByID(localGroupID_WO_versionSrv, 0)
//		}
//	}
//}

func (g *Group) syncGroupUpdatesVersionByGroupIDV2(groupID, operationID string) {
	g.Lock()
	groupSyncStatusModel := g.groupMemberSyncStatus[groupID]
	g.Unlock()
	if groupSyncStatusModel.GroupID == groupID {
		log.Debug(operationID, "syncGroupUpdatesVersionByGroupIDV2", groupID, "Sync is already in progress wait ")
		return
	}

	var groupVersionNumber int64 = 0
	//TODO Sync Group members by versioning
	//#1 get group version form local
	localGroupVersion, err := g.db.GetGroupUpdateVersionByID(groupID)
	if err != nil {
		log.Info(operationID, "syncGroupUpdatesVersionByGroupIDV2", groupID, "#1 get group version form local ", err.Error())
		groupVersionNumber = 0
	}
	groupVersionNumber = localGroupVersion.VersionNumber
	//#2 send group version to server for check
	apiReq := api.CheckLocalGroupUpdatesVersionsWithSrvReq{}
	apiReq.OperationID = operationID
	apiReq.GroupID = groupID
	apiReq.GroupVersion = groupVersionNumber
	apiReq.PageNumber = 1
	apiReq.PageSize = 300
	if groupVersionNumber == 0 {
		apiReq.NeedNextPageFetch = true
	}
	//Below Flag is how you want to get the Response, Back ON HTTP request or through WS
	apiReq.ResponseBackHTTP = true
	var result *pbGroup.GroupUpdatesVersionsRes
	log.Debug(operationID, "api args: ", apiReq)
	err = g.p.PostReturn(constant.CheckLocalGroupUpdatesVersionsWithSrv, apiReq, &result)
	if err != nil {
		log.Info(operationID, "syncGroupUpdatesVersionByGroupIDV2", groupID, err.Error())
	} else {
		//Server will figure out what has to be done for sync, will create sync methods for sdk
		//Calling Back for next Page Response Handler
		if apiReq.ResponseBackHTTP && result != nil {
			g.groupMemberSyncSilentSDKNotification(result.MemberSyncNotificationTips, operationID)
		}
	}

}

func (g *Group) getFriendListNotMemberOfTheGroup(groupID, operationID string) []model_struct.LocalFriend {
	return g.db.GetFriendListNotMemberOfTheGroup(groupID)
}
