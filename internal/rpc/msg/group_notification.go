package msg

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	pbGroup "Open_IM/pkg/proto/group"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

//message GroupCreatedTips{
//  GroupInfo Group = 1;
//  GroupMemberFullInfo Creator = 2;
//  repeated GroupMemberFullInfo MemberList = 3;
//  uint64 OperationTime = 4;
//} creator->group

func setOpUserInfo(opUserID, groupID string, groupMemberInfo *open_im_sdk.GroupMemberFullInfo) error {
	if token_verify.IsManagerUserID(opUserID) {
		u, err := imdb.GetUserByUserID(opUserID)
		if err != nil {
			return utils.Wrap(err, "GetUserByUserID failed")
		}
		utils.CopyStructFields(groupMemberInfo, u)
		groupMemberInfo.GroupID = groupID
	} else {
		u, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(groupID, opUserID)
		if err == nil {
			if err = utils2.GroupMemberDBCopyOpenIM(groupMemberInfo, u); err != nil {
				return utils.Wrap(err, "")
			}
		}
		user, err := imdb.GetUserByUserID(opUserID)
		if err != nil {
			return utils.Wrap(err, "")
		}
		groupMemberInfo.GroupID = groupID
		groupMemberInfo.UserID = user.UserID
		groupMemberInfo.Nickname = user.Nickname
		groupMemberInfo.AppMangerLevel = user.AppMangerLevel
		groupMemberInfo.FaceURL = user.FaceURL
		groupMemberInfo.AccountStatus = int32(user.AccountStatus)
	}
	return nil
}

func setGroupInfo(groupID string, groupInfo *open_im_sdk.GroupInfo) error {
	group, err := imdb.GetGroupInfoByGroupID(groupID)
	if err != nil {
		return utils.Wrap(err, "GetGroupInfoByGroupID failed")
	}
	err = utils2.GroupDBCopyOpenIM(groupInfo, group)
	if err != nil {
		return utils.Wrap(err, "GetGroupMemberNumByGroupID failed")
	}
	return nil
}

func setGroupMemberInfo(groupID, userID string, groupMemberInfo *open_im_sdk.GroupMemberFullInfo) error {
	groupMember, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(groupID, userID)
	if err == nil && groupMember != nil {
		return utils.Wrap(utils2.GroupMemberDBCopyOpenIM(groupMemberInfo, groupMember), "")
	}

	user, err := imdb.GetUserByUserID(userID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	groupMemberInfo.GroupID = groupID
	groupMemberInfo.UserID = user.UserID
	groupMemberInfo.Nickname = user.Nickname
	groupMemberInfo.AppMangerLevel = user.AppMangerLevel
	groupMemberInfo.FaceURL = user.FaceURL
	return nil
}

func setGroupMemberInfoForSync(groupID, userID string, groupMemberInfo *open_im_sdk.GroupMemberFullInfo) error {
	groupMember, err := imdb.GetGroupMemberInfoByGroupIDAndUserIDForSync(groupID, userID)
	if err == nil && groupMember != nil {
		err := utils2.GroupMemberDBCopyOpenIM(groupMemberInfo, groupMember)
		return utils.Wrap(err, "")
	}
	log.Error("setGroupMemberInfoForSync not found group member in GetGroupMemberInfoByGroupIDAndUserIDForSync ")
	user, err := imdb.GetUserByUserID(userID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	groupMemberInfo.GroupID = groupID
	groupMemberInfo.UserID = user.UserID
	groupMemberInfo.Nickname = user.Nickname
	groupMemberInfo.AppMangerLevel = user.AppMangerLevel
	groupMemberInfo.FaceURL = user.FaceURL
	return nil
}

func setGroupMemberInfoByPagingForSync(groupID string, pageNumber, pageSize int32) ([]*open_im_sdk.GroupMemberFullInfo, int32, error) {
	groupMembersFullInfo := make([]*open_im_sdk.GroupMemberFullInfo, 0)
	groupMembers, totalCount, err := imdb.GetGroupMemberInfoByGroupIDPagingForSync(groupID, pageNumber, pageSize)
	if err == nil {
		for _, member := range groupMembers {
			groupMemberInfo := new(open_im_sdk.GroupMemberFullInfo)
			err := utils2.GroupMemberDBCopyOpenIM(groupMemberInfo, &member)
			if err == nil {
				groupMembersFullInfo = append(groupMembersFullInfo, groupMemberInfo)
			}
		}

	}
	log.Error("setGroupMemberInfoForSync not found group member in GetGroupMemberInfoByGroupIDAndUserIDForSync ")
	return groupMembersFullInfo, totalCount, nil
}

func setGroupOwnerInfo(groupID string, groupMemberInfo *open_im_sdk.GroupMemberFullInfo) error {
	groupMember, err := imdb.GetGroupOwnerInfoByGroupID(groupID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	if err = utils2.GroupMemberDBCopyOpenIM(groupMemberInfo, groupMember); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func setPublicUserInfo(userID string, publicUserInfo *open_im_sdk.PublicUserInfo) error {
	user, err := imdb.GetUserByUserID(userID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	utils2.UserDBCopyOpenIMPublicUser(publicUserInfo, user)
	return nil
}

func groupNotification(contentType int32, m proto.Message, sendID, groupID, recvUserID, operationID string, OpFrom ...int32) {
	log.Info(operationID, utils.GetSelfFuncName(), "args: ", contentType, sendID, groupID, recvUserID)

	var err error
	var tips open_im_sdk.TipsComm
	tips.Detail, err = proto.Marshal(m)
	if err != nil {
		log.Error(operationID, "Marshal failed ", err.Error(), m.String())
		return
	}
	marshaler := jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  false,
		EmitDefaults: false,
	}

	tips.JsonDetail, _ = marshaler.MarshalToString(m)
	var nickname string

	from, err := imdb.GetUserByUserIDEvenDeleted(sendID)
	if err != nil {
		log.Error(operationID, "GetUserByUserID failed ", err.Error(), sendID)
	}
	if from != nil {
		nickname = from.Nickname
	}

	to, err := imdb.GetUserByUserID(recvUserID)
	if err != nil {
		log.NewWarn(operationID, "GetUserByUserID failed ", err.Error(), recvUserID)
	}
	toNickname := ""
	if to != nil {
		toNickname = to.Nickname
	}

	cn := config.Config.Notification
	switch contentType {
	case constant.GroupCreatedNotification:
		tips.DefaultTips = nickname + " " + cn.GroupCreated.DefaultTips.Tips
	case constant.GroupInfoSetNotification:
		tips.DefaultTips = nickname + " " + cn.GroupInfoSet.DefaultTips.Tips
	case constant.GroupAnnouncementNotification:
		tips.DefaultTips = nickname + " " + cn.GroupInfoSet.DefaultTips.Tips
	case constant.JoinGroupApplicationNotification:
		tips.DefaultTips = nickname + " " + cn.JoinGroupApplication.DefaultTips.Tips
	case constant.MemberQuitNotification:
		tips.DefaultTips = nickname + " " + cn.MemberQuit.DefaultTips.Tips
	case constant.GroupApplicationAcceptedNotification: //
		tips.DefaultTips = toNickname + " " + cn.GroupApplicationAccepted.DefaultTips.Tips
	case constant.GroupApplicationRejectedNotification: //
		tips.DefaultTips = toNickname + " " + cn.GroupApplicationRejected.DefaultTips.Tips
	case constant.GroupOwnerTransferredNotification: //
		tips.DefaultTips = toNickname + " " + cn.GroupOwnerTransferred.DefaultTips.Tips
	case constant.MemberKickedNotification: //
		tips.DefaultTips = toNickname + " " + cn.MemberKicked.DefaultTips.Tips
	case constant.MemberInvitedNotification: //
		tips.DefaultTips = toNickname + " " + cn.MemberInvited.DefaultTips.Tips
	case constant.MemberEnterNotification:
		tips.DefaultTips = toNickname + " " + cn.MemberEnter.DefaultTips.Tips
	case constant.GroupDismissedNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupDismissed.DefaultTips.Tips
	case constant.GroupMutedNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMuted.DefaultTips.Tips
	case constant.GroupCancelMutedNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupCancelMuted.DefaultTips.Tips
	case constant.GroupMemberMutedNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMemberMuted.DefaultTips.Tips
	case constant.GroupMemberCancelMutedNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMemberCancelMuted.DefaultTips.Tips
	case constant.GroupMemberInfoSetNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMemberInfoSet.DefaultTips.Tips
	case constant.GroupMemberSetToAdminNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMemberSetToAdmin.DefaultTips.Tips
	case constant.GroupMemberSetToOrdinaryUserNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMemberSetToOrdinary.DefaultTips.Tips
	case constant.GroupDeleteNotification:
		tips.DefaultTips = toNickname + "" + cn.GroupMemberSetToOrdinary.DefaultTips.Tips
	default:
		log.Error(operationID, "contentType failed ", contentType)
		return
	}

	var n NotificationMsg
	n.SendID = sendID
	if groupID != "" {
		n.RecvID = groupID
		n.SessionType = constant.GroupChatType
	} else {
		n.RecvID = recvUserID
		n.SessionType = constant.SingleChatType
	}
	n.ContentType = contentType
	n.OperationID = operationID
	if len(OpFrom) > 0 {
		n.OpFrom = OpFrom[0]
	}
	n.Content, err = proto.Marshal(&tips)
	if err != nil {
		log.Error(operationID, "Marshal failed ", err.Error(), tips.String())
		return
	}
	Notification(&n)
}

func groupSilentSDKNotification(contentType int32, operationID, sendID, recvID string, m proto.Message) {
	var tips open_im_sdk.TipsComm
	var err error
	marshaler := jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  false,
		EmitDefaults: false,
	}
	tips.JsonDetail, _ = marshaler.MarshalToString(m)
	n := &NotificationMsg{
		SendID:      sendID,
		RecvID:      recvID,
		MsgFrom:     constant.UserMsgType,
		ContentType: contentType,
		SessionType: constant.SingleChatType,
		OperationID: operationID,
	}
	n.Content, err = proto.Marshal(&tips)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "proto.Marshal failed")
		return
	}
	// log.NewError(utils.GetSelfFuncName(), "Notification Silent SDK Not, ", string(n.Content))
	Notification(n)
}

// 创建群后调用
// This create Group Notification, Where we were sending all the group members
// Now we moved it to sdk side, on reading this notification msg we will sync group members like we do in heartbeat
func GroupCreatedNotification(operationID, opUserID, groupID string, initMemberList []string) {
	GroupCreatedTips := open_im_sdk.GroupCreatedTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}, GroupOwnerUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setOpUserInfo(opUserID, groupID, GroupCreatedTips.OpUser); err != nil {
		log.NewError(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID, GroupCreatedTips.OpUser)
		return
	}
	err := setGroupInfo(groupID, GroupCreatedTips.Group)
	if err != nil {
		log.Error(operationID, "setGroupInfo failed ", groupID, GroupCreatedTips.Group)
		return
	}
	imdb.GetGroupOwnerInfoByGroupID(groupID)
	if err := setGroupOwnerInfo(groupID, GroupCreatedTips.GroupOwnerUser); err != nil {
		log.Error(operationID, "setGroupOwnerInfo failed", err.Error(), groupID)
		return
	}

	// var wg sync.WaitGroup
	// var split = 100
	// var count = len(initMemberList)/split + 1

	// for i := 0; i < count; i++ {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		start := i * split
	// 		end := (i + 1) * split
	// 		if end > len(initMemberList) {
	// 			end = len(initMemberList)
	// 		}
	// 		for _, v := range initMemberList[start:end] {
	// 			var member open_im_sdk.GroupMemberFullInfo
	// 			if err := setGroupMemberInfo(groupID, v, &member); err != nil {
	// 				log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID, v)
	// 				continue
	// 			}
	// 			GroupCreatedTips.MemberList = append(GroupCreatedTips.MemberList, &member)
	// 		}
	// 		groupNotification(constant.GroupCreatedNotification, &GroupCreatedTips, opUserID, groupID, "", operationID)
	// 	}(i)
	// }
	// wg.Wait()

	go groupNotification(constant.GroupCreatedNotification, &GroupCreatedTips, opUserID, groupID, "", operationID)
}

// 群信息改变后掉用
// groupName := ""
//
//	notification := ""
//	introduction := ""
//	faceURL := ""
func GroupInfoSetNotification(operationID, opUserID, groupID string, groupName, notification, introduction, faceURL string, OpFrom int32) {
	GroupInfoChangedTips := open_im_sdk.GroupInfoSetTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, GroupInfoChangedTips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	GroupInfoChangedTips.Group.GroupName = groupName
	GroupInfoChangedTips.Group.Notification = notification
	GroupInfoChangedTips.Group.Introduction = introduction
	GroupInfoChangedTips.Group.FaceURL = faceURL
	if err := setOpUserInfo(opUserID, groupID, GroupInfoChangedTips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	if notification != "" {
		groupNotification(constant.GroupAnnouncementNotification, &GroupInfoChangedTips, opUserID, groupID, "", operationID, OpFrom)
	} else {
		groupNotification(constant.GroupInfoSetNotification, &GroupInfoChangedTips, opUserID, groupID, "", operationID, OpFrom)
	}

}

func GroupMutedNotification(operationID, opUserID, groupID string) {
	tips := open_im_sdk.GroupMutedTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, tips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, tips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	groupNotification(constant.GroupMutedNotification, &tips, opUserID, groupID, "", operationID)
}

func GroupCancelMutedNotification(operationID, opUserID, groupID string) {
	tips := open_im_sdk.GroupCancelMutedTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, tips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, tips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	groupNotification(constant.GroupCancelMutedNotification, &tips, opUserID, groupID, "", operationID)
}

func GroupMemberMutedNotification(c context.Context, operationID, opUserID, groupID, groupMemberUserID string, mutedSeconds uint32) {
	tips := open_im_sdk.GroupMemberMutedTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}, MutedUser: &open_im_sdk.GroupMemberFullInfo{}}
	tips.MutedSeconds = mutedSeconds
	if err := setGroupInfo(groupID, tips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, tips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	if err := setGroupMemberInfo(groupID, groupMemberUserID, tips.MutedUser); err != nil {
		log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID, groupMemberUserID)
		return
	}
	groupNotification(constant.GroupMemberMutedNotification, &tips, opUserID, groupID, "", operationID)
}

func GroupMemberInfoSetNotification(operationID, opUserID, groupID, groupMemberUserID string) {
	tips := open_im_sdk.GroupMemberInfoSetTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}, ChangedUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, tips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, tips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	if err := setGroupMemberInfo(groupID, groupMemberUserID, tips.ChangedUser); err != nil {
		log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID, groupMemberUserID)
		return
	}
	groupNotification(constant.GroupMemberInfoSetNotification, &tips, opUserID, groupID, "", operationID)
}

func GroupMemberRoleLevelChangeNotification(operationID, opUserID, groupID, groupMemberUserID string, notificationType, OpFrom int32) {
	if notificationType != constant.GroupMemberSetToAdminNotification && notificationType != constant.GroupMemberSetToOrdinaryUserNotification {
		log.NewError(operationID, utils.GetSelfFuncName(), "invalid notificationType: ", notificationType)
		return
	}
	tips := open_im_sdk.GroupMemberInfoSetTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}, ChangedUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, tips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, tips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	if err := setGroupMemberInfo(groupID, groupMemberUserID, tips.ChangedUser); err != nil {
		log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID, groupMemberUserID)
		return
	}
	groupNotification(notificationType, &tips, opUserID, groupID, "", operationID, OpFrom)
}

func GroupMemberCancelMutedNotification(operationID, opUserID, groupID, groupMemberUserID string) {
	tips := open_im_sdk.GroupMemberCancelMutedTips{Group: &open_im_sdk.GroupInfo{},
		OpUser: &open_im_sdk.GroupMemberFullInfo{}, MutedUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, tips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, tips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	if err := setGroupMemberInfo(groupID, groupMemberUserID, tips.MutedUser); err != nil {
		log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID, groupMemberUserID)
		return
	}
	groupNotification(constant.GroupMemberCancelMutedNotification, &tips, opUserID, groupID, "", operationID)
}

//	message ReceiveJoinApplicationTips{
//	 GroupInfo Group = 1;
//	 PublicUserInfo Applicant  = 2;
//	 string 	Reason = 3;
//	}  apply->all managers GroupID              string   `protobuf:"bytes,1,opt,name=GroupID" json:"GroupID,omitempty"`
//
//	ReqMessage           string   `protobuf:"bytes,2,opt,name=ReqMessage" json:"ReqMessage,omitempty"`
//	OpUserID             string   `protobuf:"bytes,3,opt,name=OpUserID" json:"OpUserID,omitempty"`
//	OperationID          string   `protobuf:"bytes,4,opt,name=OperationID" json:"OperationID,omitempty"`
//
// 申请进群后调用
func JoinGroupApplicationNotification(req *pbGroup.JoinGroupReq) {
	JoinGroupApplicationTips := open_im_sdk.JoinGroupApplicationTips{Group: &open_im_sdk.GroupInfo{}, Applicant: &open_im_sdk.PublicUserInfo{}}
	err := setGroupInfo(req.GroupID, JoinGroupApplicationTips.Group)
	if err != nil {
		log.Error(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID)
		return
	}
	if err = setPublicUserInfo(req.OpUserID, JoinGroupApplicationTips.Applicant); err != nil {
		log.Error(req.OperationID, "setPublicUserInfo failed ", err.Error(), req.OpUserID)
		return
	}
	JoinGroupApplicationTips.ReqMsg = req.ReqMessage

	managerList, err := imdb.GetOwnerManagerByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetOwnerManagerByGroupId failed ", err.Error(), req.GroupID)
		return
	}
	for _, v := range managerList {
		groupNotification(constant.JoinGroupApplicationNotification, &JoinGroupApplicationTips, req.OpUserID, "", v.UserID, req.OperationID)
		log.NewInfo(req.OperationID, "Notification ", v)
	}
}

func MemberQuitNotification(req *pbGroup.QuitGroupReq) {
	MemberQuitTips := open_im_sdk.MemberQuitTips{Group: &open_im_sdk.GroupInfo{}, QuitUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(req.GroupID, MemberQuitTips.Group); err != nil {
		log.Error(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID)
		return
	}
	if err := setOpUserInfo(req.OpUserID, req.GroupID, MemberQuitTips.QuitUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed ", err.Error(), req.OpUserID, req.GroupID)
		return
	}

	groupNotification(constant.MemberQuitNotification, &MemberQuitTips, req.OpUserID, req.GroupID, "", req.OperationID)
	//	groupNotification(constant.MemberQuitNotification, &MemberQuitTips, req.OpUserID, "", req.OpUserID, req.OperationID)

}

//	message ApplicationProcessedTips{
//	 GroupInfo Group = 1;
//	 GroupMemberFullInfo OpUser = 2;
//	 int32 Result = 3;
//	 string 	Reason = 4;
//	}
//
// 处理进群请求后调用
func GroupApplicationAcceptedNotification(req *pbGroup.GroupApplicationResponseReq) {
	GroupApplicationAcceptedTips := open_im_sdk.GroupApplicationAcceptedTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}, HandleMsg: req.HandledMsg}
	if err := setGroupInfo(req.GroupID, GroupApplicationAcceptedTips.Group); err != nil {
		log.NewError(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID, GroupApplicationAcceptedTips.Group)
		return
	}
	if err := setOpUserInfo(req.OpUserID, req.GroupID, GroupApplicationAcceptedTips.OpUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed", req.OpUserID, req.GroupID, GroupApplicationAcceptedTips.OpUser)
		return
	}
	groupNotification(constant.GroupApplicationAcceptedNotification, &GroupApplicationAcceptedTips, req.OpUserID, "", req.FromUserID, req.OperationID)
}

func GroupApplicationRejectedNotification(req *pbGroup.GroupApplicationResponseReq) {
	GroupApplicationRejectedTips := open_im_sdk.GroupApplicationRejectedTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}, HandleMsg: req.HandledMsg}
	if err := setGroupInfo(req.GroupID, GroupApplicationRejectedTips.Group); err != nil {
		log.NewError(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID, GroupApplicationRejectedTips.Group)
		return
	}
	if err := setOpUserInfo(req.OpUserID, req.GroupID, GroupApplicationRejectedTips.OpUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed", req.OpUserID, req.GroupID, GroupApplicationRejectedTips.OpUser)
		return
	}
	groupNotification(constant.GroupApplicationRejectedNotification, &GroupApplicationRejectedTips, req.OpUserID, "", req.FromUserID, req.OperationID)
}

func GroupOwnerTransferredNotification(req *pbGroup.TransferGroupOwnerReq) {
	GroupOwnerTransferredTips := open_im_sdk.GroupOwnerTransferredTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}, NewGroupOwner: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(req.GroupID, GroupOwnerTransferredTips.Group); err != nil {
		log.NewError(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID)
		return
	}
	if err := setOpUserInfo(req.OpUserID, req.GroupID, GroupOwnerTransferredTips.OpUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed", req.OpUserID, req.GroupID)
		return
	}
	if err := setGroupMemberInfo(req.GroupID, req.NewOwnerUserID, GroupOwnerTransferredTips.NewGroupOwner); err != nil {
		log.Error(req.OperationID, "setGroupMemberInfo failed", req.GroupID, req.NewOwnerUserID)
		return
	}
	groupNotification(constant.GroupOwnerTransferredNotification, &GroupOwnerTransferredTips, req.OpUserID, req.GroupID, "", req.OperationID, req.OpFrom)
}

func GroupDismissedNotification(req *pbGroup.DismissGroupReq) {
	tips := open_im_sdk.GroupDismissedTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(req.GroupID, tips.Group); err != nil {
		log.NewError(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID)
		return
	}
	if err := setOpUserInfo(req.OpUserID, req.GroupID, tips.OpUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed", req.OpUserID, req.GroupID)
		return
	}
	groupNotification(constant.GroupDismissedNotification, &tips, req.OpUserID, req.GroupID, "", req.OperationID)
}

func GroupDeleteNotification(req *pbGroup.DeleteGroupReq) {
	tips := open_im_sdk.GroupDismissedTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(req.GroupId, tips.Group); err != nil {
		log.NewError(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupId)
		return
	}
	// if err := setOpUserInfo(req.OpUserID, req.GroupId, tips.OpUser); err != nil {
	// 	log.Error(req.OperationID, "setOpUserInfo failed", req.Op, req.GroupId)
	// 	return
	// }
	groupNotification(constant.GroupDeleteNotification, &tips, tips.Group.GetOwnerUserID(), req.GroupId, "", req.OperationID)
}

//	message MemberKickedTips{
//	 GroupInfo Group = 1;
//	 GroupMemberFullInfo OpUser = 2;
//	 GroupMemberFullInfo KickedUser = 3;
//	 uint64 OperationTime = 4;
//	}
//
// 被踢后调用
func MemberKickedNotification(req *pbGroup.KickGroupMemberReq, kickedUserIDList []string) {
	MemberKickedTips := open_im_sdk.MemberKickedTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(req.GroupID, MemberKickedTips.Group); err != nil {
		log.Error(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID)
		return
	}
	if err := setOpUserInfo(req.OpUserID, req.GroupID, MemberKickedTips.OpUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed ", err.Error(), req.OpUserID)
		return
	}
	for _, v := range kickedUserIDList {
		var groupMemberInfo open_im_sdk.GroupMemberFullInfo
		if err := setGroupMemberInfo(req.GroupID, v, &groupMemberInfo); err != nil {
			log.Error(req.OperationID, "setGroupMemberInfo failed ", err.Error(), req.GroupID, v)
			continue
		}
		MemberKickedTips.KickedUserList = append(MemberKickedTips.KickedUserList, &groupMemberInfo)
	}
	groupNotification(constant.MemberKickedNotification, &MemberKickedTips, req.OpUserID, req.GroupID, "", req.OperationID)
	//
	//for _, v := range kickedUserIDList {
	//	groupNotification(constant.MemberKickedNotification, &MemberKickedTips, req.OpUserID, "", v, req.OperationID)
	//}
}

//	message MemberInvitedTips{
//	 GroupInfo Group = 1;
//	 GroupMemberFullInfo OpUser = 2;
//	 GroupMemberFullInfo InvitedUser = 3;
//	 uint64 OperationTime = 4;
//	}
//
// 被邀请进群后调用
func MemberInvitedNotification(operationID, groupID, opUserID, reason string, invitedUserIDList []string) {
	MemberInvitedTips := open_im_sdk.MemberInvitedTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, MemberInvitedTips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return
	}
	if err := setOpUserInfo(opUserID, groupID, MemberInvitedTips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return
	}
	for _, v := range invitedUserIDList {
		var groupMemberInfo open_im_sdk.GroupMemberFullInfo
		if err := setGroupMemberInfo(groupID, v, &groupMemberInfo); err != nil {
			log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID)
			continue
		}
		MemberInvitedTips.InvitedUserList = append(MemberInvitedTips.InvitedUserList, &groupMemberInfo)
	}

	groupNotification(constant.MemberInvitedNotification, &MemberInvitedTips, opUserID, groupID, "", operationID)
}

func MemberSyncNotification(operationID, groupID, opUserID string, needNextPageFetch bool, pageNumber, pageSize int32, invitedUserIDList []string, responseBackHTTP bool) *open_im_sdk.MemberSyncNotificationTips {
	MemberSyncNotificationTips := open_im_sdk.MemberSyncNotificationTips{Group: &open_im_sdk.GroupInfo{}, OpUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(groupID, MemberSyncNotificationTips.Group); err != nil {
		log.Error(operationID, "setGroupInfo failed ", err.Error(), groupID)
		return &MemberSyncNotificationTips
	}
	if err := setOpUserInfo(opUserID, groupID, MemberSyncNotificationTips.OpUser); err != nil {
		log.Error(operationID, "setOpUserInfo failed ", err.Error(), opUserID, groupID)
		return &MemberSyncNotificationTips
	}
	if !needNextPageFetch {
		// Selected Users depending on Version code
		for _, v := range invitedUserIDList {
			var groupMemberInfo open_im_sdk.GroupMemberFullInfo
			if err := setGroupMemberInfoForSync(groupID, v, &groupMemberInfo); err != nil {
				log.Error(operationID, "setGroupMemberInfo failed ", err.Error(), groupID)
				continue
			}
			MemberSyncNotificationTips.InvitedUserList = append(MemberSyncNotificationTips.InvitedUserList, &groupMemberInfo)
		}
	} else {
		// have to fetch user by paging
		groupMembersFullInfo, totalCount, _ := setGroupMemberInfoByPagingForSync(groupID, pageNumber, pageSize)
		log.Error("setGroupMemberInfoByPagingForSync called and Result are", totalCount, groupMembersFullInfo)
		MemberSyncNotificationTips.InvitedUserList = append(MemberSyncNotificationTips.InvitedUserList, groupMembersFullInfo...)
		MemberSyncNotificationTips.TotalCount = totalCount
		MemberSyncNotificationTips.PageSize = pageSize
		MemberSyncNotificationTips.PageNumber = pageNumber
		MemberSyncNotificationTips.ResponseBackHTTP = responseBackHTTP
		if totalCount > (pageSize * pageNumber) {
			MemberSyncNotificationTips.NeedNextPageFetch = true
		} else {
			MemberSyncNotificationTips.NeedNextPageFetch = false
		}

	}
	//if !responseBackHTTP {
	//	groupSilentSDKNotification(constant.GroupMemberSyncSilentSDKNotification, operationID, "sandman", opUserID, &MemberSyncNotificationTips)
	//
	//}
	return &MemberSyncNotificationTips
}

//message GroupInfoChangedTips{
//  int32 ChangedType = 1; //bitwise operators: 1:groupName; 10:Notification  100:Introduction; 1000:FaceUrl
//  GroupInfo Group = 2;
//  GroupMemberFullInfo OpUser = 3;
//}

//message MemberLeaveTips{
//  GroupInfo Group = 1;
//  GroupMemberFullInfo LeaverUser = 2;
//  uint64 OperationTime = 3;
//}

//群成员退群后调用

//	message MemberEnterTips{
//	 GroupInfo Group = 1;
//	 GroupMemberFullInfo EntrantUser = 2;
//	 uint64 OperationTime = 3;
//	}
//
// 群成员主动申请进群，管理员同意后调用，
func MemberEnterNotification(req *pbGroup.GroupApplicationResponseReq) {
	MemberEnterTips := open_im_sdk.MemberEnterTips{Group: &open_im_sdk.GroupInfo{}, EntrantUser: &open_im_sdk.GroupMemberFullInfo{}}
	if err := setGroupInfo(req.GroupID, MemberEnterTips.Group); err != nil {
		log.Error(req.OperationID, "setGroupInfo failed ", err.Error(), req.GroupID, MemberEnterTips.Group)
		return
	}
	if err := setGroupMemberInfo(req.GroupID, req.FromUserID, MemberEnterTips.EntrantUser); err != nil {
		log.Error(req.OperationID, "setOpUserInfo failed ", err.Error(), req.OpUserID, req.GroupID, MemberEnterTips.EntrantUser)
		return
	}
	groupNotification(constant.MemberEnterNotification, &MemberEnterTips, req.OpUserID, req.GroupID, "", req.OperationID)

}
