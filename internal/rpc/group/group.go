package group

import (
	"Open_IM/internal/rpc/admin_cms"
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	cp "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbCache "Open_IM/pkg/proto/cache"
	pbConversation "Open_IM/pkg/proto/conversation"
	pbGroup "Open_IM/pkg/proto/group"
	"Open_IM/pkg/proto/local_database"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
)

type groupServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewGroupServer(port int) *groupServer {
	return &groupServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImGroupName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *groupServer) Run() {
	log.NewPrivateLog(constant.OpenImGroupLog)
	log.NewInfo("", "group rpc start ")
	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)
	// listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("", "listen network success, ", address, listener)
	defer listener.Close()
	// grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	// Service registers with etcd
	pbGroup.RegisterGroupServer(srv, s)

	rpcRegisterIP := config.Config.RpcRegisterIP
	if rpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("", "RegisterEtcd failed ", err.Error())
		return
	}
	log.Info("", "RegisterEtcd ", s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName)
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("", "group rpc success")
}

func (s *groupServer) CreateGroup(ctx context.Context, req *pbGroup.CreateGroupReq) (*pbGroup.CreateGroupResp, error) {

	//log.NewInfo(req.OperationID, "CreateGroup, args ", req.String())
	resp := &pbGroup.CreateGroupResp{GroupInfo: &open_im_sdk.GroupInfo{}}
	if len(req.InitMemberList) > config.Config.MembersInGroupMaxLimit {
		log.NewError(req.OperationID, "Defined Limit of group memebers is excedeed")
		resp.ErrCode = constant.LimitExceeded.ErrCode
		resp.ErrMsg = "defined limit of group memebers is exceeded"
		return resp, nil
	}
	// if req.OpUserID != "" {
	// 	us, err := imdb.GetUserByUserID(req.OpUserID)
	// 	if err != nil {
	// 		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.OpUserID)
	// 		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
	// 	}
	// 	if us != nil && us.SuperUserStatus != 1 {
	// 		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrNotASuperUser.ErrCode, ErrMsg: constant.ErrNotASuperUser.ErrMsg}, nil
	// 	}
	// }

	groupId := req.GroupInfo.GroupID
	if groupId == "" {
		groupId = utils.Md5(req.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
		bi := big.NewInt(0)
		bi.SetString(groupId[0:8], 16)
		groupId = bi.String()
	}
	// to group
	groupInfo := db.Group{}
	utils.CopyStructFields(&groupInfo, req.GroupInfo)
	groupInfo.CreatorUserID = req.OpUserID
	groupInfo.GroupID = groupId
	groupInfo.IsOpen = req.IsOpen
	err := imdb.InsertIntoGroup(groupInfo)
	if err != nil {
		log.NewError(req.OperationID, "InsertIntoGroup failed, ", err.Error(), groupInfo)
		return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
	}

	// set group interest type
	if len(req.GroupInterest) > 0 {
		imdb.SetGroupInterestType(groupId, req.GroupInterest)
	}

	var okUserIDList []string

	groupMember := db.GroupMember{}
	us := &db.User{}
	if req.OwnerUserID != "" {
		us, err = imdb.GetUserByUserID(req.OwnerUserID)
		if err != nil {
			log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.OwnerUserID)
			return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
		}
		// to group member
		groupMember = db.GroupMember{GroupID: groupId, RoleLevel: constant.GroupOwner, OperatorUserID: req.OpUserID}
		utils.CopyStructFields(&groupMember, us)
		err = imdb.InsertIntoGroupMember(groupMember)
		if err != nil {
			log.NewError(req.OperationID, "InsertIntoGroupMember failed ", err.Error(), groupMember)
			return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, http.WrapError(constant.ErrDB)
		}
		// set user group min seq to group max seq, no need to show previous messages to user
		maxSeq, _ := db.DB.GetGrupMaxSeq(groupMember.GroupID)
		if maxSeq > 0 {
			_ = db.DB.SetGroupMinSeq(groupMember.UserID, groupMember.GroupID, uint32(maxSeq))
		}
	}
	if req.GroupInfo.GroupType != constant.SuperGroup {
		// to group member
		// set user group min seq to group max seq, no need to show previous messages to user
		maxSeq, _ := db.DB.GetGrupMaxSeq(groupMember.GroupID)
		for _, user := range req.InitMemberList {
			us, err := imdb.GetUserByUserID(user.UserID)
			if err != nil {
				log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), user.UserID)
				continue
			}
			if user.RoleLevel == constant.GroupOwner {
				log.NewError(req.OperationID, "only one owner, failed ", user)
				continue
			}
			groupMember.RoleLevel = user.RoleLevel
			utils.CopyStructFields(&groupMember, us)
			err = imdb.InsertIntoGroupMember(groupMember)
			if err != nil {
				log.NewError(req.OperationID, "InsertIntoGroupMember failed ", err.Error(), groupMember)
				continue
			}
			// set user group min seq to group max seq, no need to show previous messages to user
			if maxSeq > 0 {
				_ = db.DB.SetGroupMinSeq(groupMember.UserID, groupMember.GroupID, uint32(maxSeq))
			}
			okUserIDList = append(okUserIDList, user.UserID)
		}
		group, err := imdb.GetGroupInfoByGroupID(groupId)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupId)
			resp.ErrCode = constant.ErrDB.ErrCode
			resp.ErrMsg = err.Error()
			return resp, nil
		}
		utils.CopyStructFields(resp.GroupInfo, group)
		membernum, err := imdb.GetGroupMemberNumByGroupID(groupId)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberNumByGroupID failed ", err.Error(), groupId)
			resp.ErrCode = constant.ErrDB.ErrCode
			resp.ErrMsg = err.Error()
			return resp, nil
		}
		resp.GroupInfo.MemberCount = uint32(membernum)
		if req.OwnerUserID != "" {
			resp.GroupInfo.OwnerUserID = req.OwnerUserID
			okUserIDList = append(okUserIDList, req.OwnerUserID)
		}
	} else {
		for _, v := range req.InitMemberList {
			okUserIDList = append(okUserIDList, v.UserID)
		}
		if err := db.DB.CreateSuperGroup(groupId, okUserIDList, len(okUserIDList)); err != nil {
			log.NewError(req.OperationID, "GetGroupMemberNumByGroupID failed ", err.Error(), groupId)
			resp.ErrCode = constant.ErrDB.ErrCode
			resp.ErrMsg = err.Error() + ": CreateSuperGroup failed"
		}
	}

	if len(okUserIDList) != 0 {
		addGroupMemberToCacheReq := &pbCache.AddGroupMemberToCacheReq{
			UserIDList:  okUserIDList,
			GroupID:     groupId,
			OperationID: req.OperationID,
		}
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return &pbGroup.CreateGroupResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}, nil
		}
		cacheClient := pbCache.NewCacheClient(etcdConn)
		cacheResp, err := cacheClient.AddGroupMemberToCache(context.Background(), addGroupMemberToCacheReq)
		if err != nil {
			log.NewError(req.OperationID, "AddGroupMemberToCache rpc call failed ", err.Error())
			return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, "AddGroupMemberToCache rpc logic call failed ", cacheResp.String())
			return &pbGroup.CreateGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
		}

		//data synchronization
		// syncGroupToLocal(req.OperationID, groupId, constant.SyncCreateGroup, nil, nil, nil)

		//log.NewInfo(req.OperationID, "/ ", resp.String())
		if req.GroupInfo.GroupType != constant.SuperGroup {
			chat.GroupCreatedNotification(req.OperationID, req.OpUserID, groupId, okUserIDList)
		} else {
			chat.SuperGroupNotification(req.OperationID, req.OpUserID, groupId)

		}
		return resp, nil
	} else {
		log.NewInfo(req.OperationID, "rpc CreateGroup return ", resp.String())
		return resp, nil
	}
}

func (s *groupServer) GetJoinedGroupList(ctx context.Context, req *pbGroup.GetJoinedGroupListReq) (*pbGroup.GetJoinedGroupListResp, error) {
	//log.NewInfo(req.OperationID, "GetJoinedGroupList, args ", req.String())
	// group list
	joinedGroupList, err := imdb.GetJoinedGroupIDListByUserID(req.FromUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetJoinedGroupIDListByUserID failed ", err.Error(), req.FromUserID)
		return &pbGroup.GetJoinedGroupListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	//log.NewInfo(req.OperationID, "joinedGroupList", joinedGroupList, "length", len(joinedGroupList))

	var resp pbGroup.GetJoinedGroupListResp
	for _, v := range joinedGroupList {
		var groupNode open_im_sdk.GroupInfo
		num, err := imdb.GetGroupMemberNumByGroupID(v)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberNumByGroupID failed ", v, err.Error())
			continue
		}
		owner, err2 := imdb.GetGroupOwnerInfoByGroupID(v)
		if err2 != nil {
			log.NewError(req.OperationID, "GetGroupOwnerInfoByGroupID failed ", v, err2.Error())
			continue
		}
		group, err3 := imdb.GetGroupInfoByGroupID(v)
		if err3 != nil {
			log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", v, err3.Error())
			continue
		}
		if num > 0 && owner != nil && group != nil {
			utils.CopyStructFields(&groupNode, group)
			groupNode.CreateTime = uint32(group.CreateTime.Unix())
			groupNode.MemberCount = uint32(num)
			groupNode.OwnerUserID = owner.UserID
			resp.GroupList = append(resp.GroupList, &groupNode)
		} else {
			log.NewError(req.OperationID, "check nil ", num, owner, group)
			continue
		}
		//log.NewDebug(req.OperationID, "joinedGroup ", groupNode)
	}
	log.NewInfo(req.OperationID, "GetJoinedGroupList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) InviteUserToGroup(ctx context.Context, req *pbGroup.InviteUserToGroupReq) (*pbGroup.InviteUserToGroupResp, error) {
	//log.NewInfo(req.OperationID, "InviteUserToGroup args ", req.String())

	membersCount, _ := imdb.GetGroupMemberNumByGroupID(req.GroupID)
	membersCountInt := int(membersCount) + len(req.InvitedUserIDList)
	if membersCountInt > config.Config.MembersInGroupMaxLimit {
		log.NewError(req.OperationID, "Defined Limit of group memebers is excedeed")
		respErrMsg := "defined limit of group memebers is exceeded"
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.LimitExceeded.ErrCode, ErrMsg: respErrMsg}, nil
	}

	if !imdb.IsExistGroupMember(req.GroupID, req.OpUserID) && !token_verify.IsManagerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "no permission InviteUserToGroup ", req.GroupID, req.OpUserID)
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupID, err)
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		errMsg := " group status is dismissed "
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}, nil
	}
	//
	// from User:  invite: applicant
	// to user:  invite: invited
	var okUserIDList []string
	if groupInfo.GroupType != constant.SuperGroup {
		var resp pbGroup.InviteUserToGroupResp
		for _, v := range req.InvitedUserIDList {
			var resultNode pbGroup.Id2Result
			resultNode.UserID = v
			resultNode.Result = 0
			toUserInfo, err := imdb.GetUserByUserID(v)
			if err != nil {
				log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), v)
				resultNode.Result = -1
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
			}

			if imdb.IsExistGroupMember(req.GroupID, v) {
				log.NewError(req.OperationID, "IsExistGroupMember ", req.GroupID, v)
				resultNode.Result = -1
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
			}
			var toInsertInfo db.GroupMember
			utils.CopyStructFields(&toInsertInfo, toUserInfo)
			toInsertInfo.GroupID = req.GroupID
			toInsertInfo.RoleLevel = constant.GroupOrdinaryUsers
			toInsertInfo.OperatorUserID = req.OpUserID
			err = imdb.InsertIntoGroupMember(toInsertInfo)
			if err != nil {
				log.NewError(req.OperationID, "InsertIntoGroupMember failed ", req.GroupID, toUserInfo.UserID, toUserInfo.Nickname, toUserInfo.FaceURL)
				resultNode.Result = -1
				resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
				continue
			}
			// set user group min seq to group max seq, no need to show previous messages to user
			maxSeq, _ := db.DB.GetGrupMaxSeq(toInsertInfo.GroupID)
			if maxSeq > 0 {
				_ = db.DB.SetGroupMinSeq(toInsertInfo.UserID, toInsertInfo.GroupID, uint32(maxSeq))
			}
			okUserIDList = append(okUserIDList, v)
			err = db.DB.AddGroupMember(req.GroupID, toUserInfo.UserID)
			if err != nil {
				log.NewError(req.OperationID, "AddGroupMember failed ", err.Error(), req.GroupID, toUserInfo.UserID)
			}
			resp.Id2ResultList = append(resp.Id2ResultList, &resultNode)
		}
		var haveConUserID []string
		conversations, err := imdb.GetConversationsByConversationIDMultipleOwner(okUserIDList, utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType))
		if err != nil {
			log.NewError(req.OperationID, "GetConversationsByConversationIDMultipleOwner failed ", err.Error(), req.GroupID, constant.GroupChatType)
		}
		for _, v := range conversations {
			haveConUserID = append(haveConUserID, v.OwnerUserID)
		}
		var reqPb pbUser.SetConversationReq
		var c pbUser.Conversation
		for _, v := range conversations {
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = v.OwnerUserID
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.RecvMsgOpt = v.RecvMsgOpt
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsPinned = v.IsPinned
			c.AttachedInfo = v.AttachedInfo
			c.IsPrivateChat = v.IsPrivateChat
			c.GroupAtType = v.GroupAtType
			c.IsNotInGroup = false
			c.Ex = v.Ex
			reqPb.Conversation = &c
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if etcdConn == nil {
				errMsg := req.OperationID + "getcdv3.GetConn == nil"
				log.NewError(req.OperationID, errMsg)
				return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}, nil
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v.OwnerUserID)
			} else {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v.OwnerUserID)
			}
		}
		for _, v := range utils.DifferenceString(haveConUserID, okUserIDList) {
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = v
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsNotInGroup = false
			reqPb.Conversation = &c
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if etcdConn == nil {
				errMsg := req.OperationID + "getcdv3.GetConn == nil"
				log.NewError(req.OperationID, errMsg)
				return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}, nil
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v)
			} else {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
			}
		}
	} else {
		okUserIDList = req.InvitedUserIDList
		if err := db.DB.AddUserToSuperGroup(req.GroupID, req.InvitedUserIDList); err != nil {
			log.NewError(req.OperationID, "AddUserToSuperGroup failed ", req.GroupID, err)
			return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}, nil
		}

	}
	addGroupMemberToCacheReq := &pbCache.AddGroupMemberToCacheReq{
		UserIDList:  okUserIDList,
		GroupID:     req.GroupID,
		OperationID: req.OperationID,
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}, nil
	}
	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.AddGroupMemberToCache(context.Background(), addGroupMemberToCacheReq)
	if err != nil {
		log.NewError(req.OperationID, "AddGroupMemberToCache rpc call failed ", err.Error())
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationID, "AddGroupMemberToCache rpc logic call failed ", cacheResp.String())
		return &pbGroup.InviteUserToGroupResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	//data synchronization
	// syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncInvitedGroupMember, nil, req.InvitedUserIDList, nil)

	if groupInfo.GroupType != constant.SuperGroup {
		chat.MemberInvitedNotification(req.OperationID, req.GroupID, req.OpUserID, req.Reason, okUserIDList)
	} else {
		chat.SuperGroupNotification(req.OperationID, req.OpUserID, req.GroupID)
	}

	log.NewInfo(req.OperationID, "InviteUserToGroup rpc return ")
	return &pbGroup.InviteUserToGroupResp{}, nil
}

func (s *groupServer) ChannelCodeInviteUserToGroup(ctx context.Context, request *pbGroup.ChannelCodeInviteUserToGroupRequest) (*pbGroup.ChannelCodeInviteUserToGroupResponse, error) {
	//log.NewInfo(request.OperationID, "ChannelCodeInviteUserToGroup args ", request.String())

	// user is existed
	userInfo, err := imdb.GetUserByUserID(request.UserId)
	if err != nil || userInfo == nil {
		errMsg := "user is not existed"
		return &pbGroup.ChannelCodeInviteUserToGroupResponse{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}, nil
	}

	response := pbGroup.ChannelCodeInviteUserToGroupResponse{}
	for _, v := range request.InvitedGroupIDList {
		var responseNode pbGroup.GroupId2Result
		responseNode.GroupID = v
		responseNode.Result = 0

		groupInfo, err := imdb.GetGroupInfoByGroupID(v)
		if err != nil {
			log.NewError(request.OperationID, "GetGroupInfoByGroupID failed ", v, err)
			responseNode.Result = -1
			response.GroupId2ResultList = append(response.GroupId2ResultList, &responseNode)
			continue
		}
		if groupInfo.Status == constant.GroupStatusDismissed {
			responseNode.Result = -1
			response.GroupId2ResultList = append(response.GroupId2ResultList, &responseNode)
			continue
		}
		if imdb.IsExistGroupMember(v, request.UserId) {
			log.NewError(request.OperationID, "IsExistGroupMember ", v, request.UserId)
			responseNode.Result = -1
			response.GroupId2ResultList = append(response.GroupId2ResultList, &responseNode)
			continue
		}

		// No super group for now
		if groupInfo.GroupType == constant.NormalGroup {

			// overcrowding
			count, _ := imdb.GetGroupMembersCount(v, "")
			if int(count) >= config.Config.MembersInGroupMaxLimit {
				log.NewError(request.OperationID, "channel InsertIntoGroupMember failed,overcrowding ", groupInfo.GroupID, request.UserId)
				responseNode.Result = -1
				response.GroupId2ResultList = append(response.GroupId2ResultList, &responseNode)
				continue
			}

			groupOwnerInfo, _ := imdb.GetGroupOwnerInfoByGroupID(v)
			// add group
			var toInsertInfo db.GroupMember
			utils.CopyStructFields(&toInsertInfo, userInfo)
			toInsertInfo.GroupID = v
			toInsertInfo.RoleLevel = constant.GroupOrdinaryUsers
			toInsertInfo.OperatorUserID = ""
			if groupOwnerInfo != nil {
				toInsertInfo.OperatorUserID = groupOwnerInfo.UserID
			}
			err = imdb.InsertIntoGroupMember(toInsertInfo)
			if err != nil {
				log.NewError(request.OperationID, "channel InsertIntoGroupMember failed ", groupInfo.GroupID, request.UserId, err)
				responseNode.Result = -1
				response.GroupId2ResultList = append(response.GroupId2ResultList, &responseNode)
				continue
			}
			response.GroupId2ResultList = append(response.GroupId2ResultList, &responseNode)

			// set user group min seq to group max seq, no need to show previous messages to user
			maxSeq, _ := db.DB.GetGrupMaxSeq(toInsertInfo.GroupID)
			if maxSeq > 0 {
				_ = db.DB.SetGroupMinSeq(toInsertInfo.UserID, toInsertInfo.GroupID, uint32(maxSeq))
			}
			// add conversation
			var reqPb pbUser.SetConversationReq
			var c pbUser.Conversation
			reqPb.OperationID = request.OperationID
			c.OwnerUserID = request.UserId
			c.ConversationID = utils.GetConversationIDBySessionType(v, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = v
			c.IsNotInGroup = false
			reqPb.Conversation = &c
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, request.OperationID)
			if etcdConn == nil {
				errMsg := request.OperationID + "getcdv3.GetConn == nil"
				log.NewError(request.OperationID, errMsg, request.UserId, v)
			} else {
				client := pbUser.NewUserClient(etcdConn)
				respPb, err := client.SetConversation(context.Background(), &reqPb)
				if err != nil {
					log.NewError(request.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v)
				} else {
					log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
				}
			}

			// cache
			addGroupMemberToCacheReq := &pbCache.AddGroupMemberToCacheReq{
				UserIDList:  []string{request.UserId},
				GroupID:     v,
				OperationID: request.OperationID,
			}
			etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, request.OperationID)
			if etcdConn == nil {
				errMsg := request.OperationID + "getcdv3.GetConn == nil"
				log.NewError(request.OperationID, errMsg, request.UserId, v)
			} else {
				cacheClient := pbCache.NewCacheClient(etcdConn)
				cacheResp, err := cacheClient.AddGroupMemberToCache(context.Background(), addGroupMemberToCacheReq)
				if err != nil {
					log.NewError(request.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v)
				} else {
					log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "SetConversation success", cacheResp.String(), v)
				}
			}

			// chat
			groupOwnerInfo, _ = imdb.GetGroupOwnerInfoByGroupID(v)
			chat.MemberInvitedNotification(request.OperationID, v, toInsertInfo.OperatorUserID, request.Reason, []string{request.UserId})

			//data synchronization
			//syncGroupToLocal(request.OperationID, groupInfo.GroupID, constant.SyncInvitedGroupMember, nil, []string{request.UserId}, nil)
		}
	}

	log.NewInfo(request.OperationID, "channel InviteUserToGroup rpc return ")
	return &response, nil
}

func (s *groupServer) GetGroupAllMember(ctx context.Context, req *pbGroup.GetGroupAllMemberReq) (*pbGroup.GetGroupAllMemberResp, error) {
	//log.NewInfo(req.OperationID, "GetGroupAllMember, args ", req.String())
	var resp pbGroup.GetGroupAllMemberResp
	opUserID := req.OpUserID

	_, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, opUserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error(), req.GroupID)
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		return &resp, nil
	}

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error(), req.GroupID)
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		return &resp, nil
	}
	if groupInfo.GroupType != constant.SuperGroup {
		memberList, err := imdb.GetGroupMemberListByGroupID(req.GroupID)
		if err != nil {
			resp.ErrCode = constant.ErrDB.ErrCode
			resp.ErrMsg = constant.ErrDB.ErrMsg
			log.NewError(req.OperationID, "GetGroupMemberListByGroupID failed,", err.Error(), req.GroupID)
			return &resp, nil
		}

		for _, v := range memberList {
			// log.Debug(req.OperationID, v)
			var node open_im_sdk.GroupMemberFullInfo
			cp.GroupMemberDBCopyOpenIM(&node, &v)
			// log.Debug(req.OperationID, "db value:", v.MuteEndTime, "seconds: ", v.MuteEndTime.Unix())
			// log.Debug(req.OperationID, "cp value: ", node)
			resp.MemberList = append(resp.MemberList, &node)
		}
	}
	log.NewInfo(req.OperationID, "GetGroupAllMember rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupMemberList(ctx context.Context, req *pbGroup.GetGroupMemberListReq) (*pbGroup.GetGroupMemberListResp, error) {
	//log.NewInfo(req.OperationID, "GetGroupMemberList args ", req.String())
	var resp pbGroup.GetGroupMemberListResp
	memberList, err := imdb.GetGroupMemberByGroupID(req.GroupID, req.Filter, req.NextSeq, 30)
	if err != nil {
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, "GetGroupMemberByGroupId failed,", req.GroupID, req.Filter, req.NextSeq, 30)
		return &resp, nil
	}

	for _, v := range memberList {
		var node open_im_sdk.GroupMemberFullInfo
		utils.CopyStructFields(&node, &v)
		resp.MemberList = append(resp.MemberList, &node)
	}
	// db operate  get db sorted by join time
	if int32(len(memberList)) < 30 {
		resp.NextSeq = 0
	} else {
		resp.NextSeq = req.NextSeq + int32(len(memberList))
	}

	resp.ErrCode = 0
	log.NewInfo(req.OperationID, "GetGroupMemberList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupMemberListV2(ctx context.Context, req *pbGroup.GetGroupMembersReqV2) (*pbGroup.GetGroupMembersResV2, error) {
	//log.Error(req.OperationID, "SandiV2 GetGroupMemberList args ", req.String())
	var resp pbGroup.GetGroupMembersResV2
	memberList, err := imdb.GetGroupMemberByGroupIDV2(req.GroupId, req.Filter, req.SearchName, req.Limit, req.Offset)
	if err != nil {
		log.NewError(err.Error())
		return &resp, nil
	}

	for _, v := range memberList {
		var node open_im_sdk.GroupMemberFullInfo
		node.UserID = v.UserID
		node.Nickname = v.Nickname
		node.JoinTime = int32(v.JoinTime.Unix())
		node.FaceURL = v.FaceURL
		node.RoleLevel = v.RoleLevel
		node.JoinSource = v.JoinSource
		node.MuteEndTime = uint32(v.MuteEndTime.Unix())

		resp.Members = append(resp.Members, &node)
	}
	log.Error(req.OperationID, "SandiV2 GetGroupMemberList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) getGroupUserLevel(groupID, userID string) (int, error) {
	opFlag := 0
	opInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(groupID, userID)
	if err != nil {
		return opFlag, utils.Wrap(err, "")
	}
	if opInfo.RoleLevel == constant.GroupOrdinaryUsers {
		opFlag = 0
	} else if opInfo.RoleLevel == constant.GroupOwner {
		opFlag = 2 // owner
	} else {
		opFlag = 3 // admin
	}

	return opFlag, nil
}

func (s *groupServer) KickGroupMember(ctx context.Context, req *pbGroup.KickGroupMemberReq) (*pbGroup.KickGroupMemberResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc args ", req.String())
	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupInfoByGroupID", req.GroupID, err.Error())
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	var okUserIDList []string
	var kickedMemberList []*db.GroupMember
	var resp pbGroup.KickGroupMemberResp
	if groupInfo.GroupType != constant.SuperGroup {
		opFlag := 0
		if !token_verify.IsManagerUserID(req.OpUserID) {
			opInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
			if err != nil {
				errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID  failed " + err.Error() + req.GroupID + req.OpUserID
				log.Error(req.OperationID, errMsg)
				return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
			}
			if opInfo.RoleLevel == constant.GroupOrdinaryUsers {
				errMsg := req.OperationID + " opInfo.RoleLevel == constant.GroupOrdinaryUsers " + opInfo.UserID + opInfo.GroupID
				log.Error(req.OperationID, errMsg)
				return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
			} else if opInfo.RoleLevel == constant.GroupOwner {
				opFlag = 2 // owner
			} else {
				opFlag = 3 // admin
			}
		} else {
			opFlag = 1 // app manager
		}

		// op is group owner?
		if len(req.KickedUserIDList) == 0 {
			log.NewError(req.OperationID, "failed, kick list 0")
			return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}, nil
		}

		// remove
		for _, v := range req.KickedUserIDList {
			kickedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, v)
			if err != nil {
				log.NewError(req.OperationID, " GetGroupMemberInfoByGroupIDAndUserID failed ", req.GroupID, v, err.Error())
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
				continue
			}

			if kickedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
				log.Error(req.OperationID, "is constant.GroupAdmin, can't kicked ", v)
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
				continue
			}
			if kickedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
				log.NewDebug(req.OperationID, "is constant.GroupOwner, can't kicked ", v)
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
				continue
			}

			member, err := imdb.GetGroupMemberByUserIDGroupID(req.GroupID, v)
			if err == nil && member.GroupID != "" {
				kickedMemberList = append(kickedMemberList, member)
			}

			err = imdb.RemoveGroupMember(req.GroupID, v)
			if err != nil {
				log.NewError(req.OperationID, "RemoveGroupMember failed ", err.Error(), req.GroupID, v)
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: -1})
			} else {
				log.NewDebug(req.OperationID, "kicked ", v)
				resp.Id2ResultList = append(resp.Id2ResultList, &pbGroup.Id2Result{UserID: v, Result: 0})
				okUserIDList = append(okUserIDList, v)
			}
		}
		var reqPb pbUser.SetConversationReq
		var c pbUser.Conversation
		for _, v := range okUserIDList {
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = v
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsNotInGroup = true
			reqPb.Conversation = &c
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if etcdConn == nil {
				errMsg := req.OperationID + "getcdv3.GetConn == nil"
				log.NewError(req.OperationID, errMsg)
				resp.ErrCode = constant.ErrInternal.ErrCode
				resp.ErrMsg = errMsg
				return &resp, nil
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v)
			} else {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
			}

			//Kick Group Member
			//reqLocalData := local_database.KickGroupMemberForAllReq{
			//	OperationID: req.OperationID,
			//	GroupID:     req.GroupID,
			//	KickedID:    v,
			//}
			//etcdConnLocalData := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.OperationID)
			//if etcdConnLocalData == nil {
			//	errMsg := req.OperationID + "getcdv3.GetConn == nil"
			//	log.NewError(req.OperationID, errMsg)
			//	resp.ErrCode = constant.ErrInternal.ErrCode
			//	resp.ErrMsg = errMsg
			//	return &resp, nil
			//}
			//localDataClient := local_database.NewLocalDataBaseClient(etcdConnLocalData)
			//respLocalData, err := localDataClient.KickGroupMemerForAllToLocal(context.Background(), &reqLocalData)
			//if err != nil {
			//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "KickGroupMemerForAllToLocal rpc failed, ", respLocalData.String(), err.Error(), v)
			//} else {
			//	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "KickGroupMemerForAllToLocal success", respLocalData.String(), v)
			//}

		}
	} else {
		okUserIDList = req.KickedUserIDList
		if err := db.DB.RemoverUserFromSuperGroup(req.GroupID, okUserIDList); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), req.GroupID, req.KickedUserIDList, err.Error())
			resp.ErrCode = constant.ErrDB.ErrCode
			resp.ErrMsg = constant.ErrDB.ErrMsg
			return &resp, nil
		}
	}

	reduceGroupMemberFromCacheReq := &pbCache.ReduceGroupMemberFromCacheReq{
		UserIDList:  okUserIDList,
		GroupID:     req.GroupID,
		OperationID: req.OperationID,
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}, nil
	}
	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.ReduceGroupMemberFromCache(context.Background(), reduceGroupMemberFromCacheReq)
	if err != nil {
		log.NewError(req.OperationID, "ReduceGroupMemberFromCache rpc call failed ", err.Error())
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationID, "ReduceGroupMemberFromCache rpc logic call failed ", cacheResp.String())
		return &pbGroup.KickGroupMemberResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	//data synchronization
	//syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncKickedGroupMember, okUserIDList, nil, kickedMemberList)
	//log.NewError(req.OperationID, utils.GetSelfFuncName(), groupInfo.GroupID, constant.SyncKickedGroupMember, req.KickedUserIDList)

	if groupInfo.GroupType != constant.SuperGroup {
		chat.MemberKickedNotification(req, okUserIDList)
	} else {
		chat.SuperGroupNotification(req.OperationID, req.OpUserID, req.GroupID)
	}
	log.NewInfo(req.OperationID, "GetGroupMemberList rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupMembersInfo(ctx context.Context, req *pbGroup.GetGroupMembersInfoReq) (*pbGroup.GetGroupMembersInfoResp, error) {
	//log.NewInfo(req.OperationID, "GetGroupMembersInfo args ", req.String())

	var resp pbGroup.GetGroupMembersInfoResp

	for _, v := range req.MemberList {
		var memberNode open_im_sdk.GroupMemberFullInfo
		memberInfo, err := imdb.GetMemberInfoByID(req.GroupID, v)
		memberNode.UserID = v
		if err != nil {
			log.NewError(req.OperationID, "GetMemberInfoById failed ", err.Error(), req.GroupID, v)
			continue
		} else {
			utils.CopyStructFields(&memberNode, memberInfo)
			memberNode.JoinTime = int32(memberInfo.JoinTime.Unix())
			memberNode.MuteEndTime = uint32(memberInfo.MuteEndTime.Unix())
			resp.MemberList = append(resp.MemberList, &memberNode)
		}
	}
	resp.ErrCode = 0
	log.NewInfo(req.OperationID, "GetGroupMembersInfo rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GetGroupApplicationList(_ context.Context, req *pbGroup.GetGroupApplicationListReq) (*pbGroup.GetGroupApplicationListResp, error) {
	//log.NewInfo(req.OperationID, "GetGroupMembersInfo args ", req.String())
	reply, err := imdb.GetGroupApplicationList(req.FromUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupApplicationList failed ", err.Error(), req.FromUserID)
		return &pbGroup.GetGroupApplicationListResp{ErrCode: 701, ErrMsg: "GetGroupApplicationList failed"}, nil
	}

	//log.NewDebug(req.OperationID, "GetGroupApplicationList reply ", reply)
	resp := pbGroup.GetGroupApplicationListResp{}
	for _, v := range reply {
		node := open_im_sdk.GroupRequest{UserInfo: &open_im_sdk.PublicUserInfo{}, GroupInfo: &open_im_sdk.GroupInfo{}}
		group, err := imdb.GetGroupInfoByGroupID(v.GroupID)
		if err != nil {
			log.Error(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), v.GroupID)
			continue
		}
		user, err := imdb.GetUserByUserID(v.UserID)
		if err != nil {
			log.Error(req.OperationID, "GetUserByUserID failed ", err.Error(), v.UserID)
			continue
		}

		cp.GroupRequestDBCopyOpenIM(&node, &v)
		cp.UserDBCopyOpenIMPublicUser(node.UserInfo, user)
		cp.GroupDBCopyOpenIM(node.GroupInfo, group)
		log.NewDebug(req.OperationID, "node ", node, "v ", v)
		resp.GroupRequestList = append(resp.GroupRequestList, &node)
	}
	if resp.GroupRequestList != nil {
		sort.Slice(resp.GroupRequestList, func(i, j int) bool {
			return resp.GroupRequestList[i].ReqTime > resp.GroupRequestList[j].ReqTime
		})
	}

	//log.NewInfo(req.OperationID, "GetGroupMembersInfo rpc return ", resp)
	return &resp, nil
}

func (s *groupServer) GetGroupsInfo(ctx context.Context, req *pbGroup.GetGroupsInfoReq) (*pbGroup.GetGroupsInfoResp, error) {
	//log.NewInfo(req.OperationID, "GetGroupsInfo args ", req.String())
	groupsInfoList := make([]*open_im_sdk.GroupInfo, 0)
	for _, groupID := range req.GroupIDList {
		groupInfoFromMysql, err := imdb.GetGroupInfoByGroupID(groupID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupID)
			continue
		}
		var groupInfo open_im_sdk.GroupInfo
		cp.GroupDBCopyOpenIM(&groupInfo, groupInfoFromMysql)
		groupsInfoList = append(groupsInfoList, &groupInfo)
	}

	// use group name search
	for _, groupID := range req.GroupIDList {
		groupInfoFromMysql, err := imdb.GetGroupsByWholeName(groupID, 0, 0)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupID)
			continue
		}
		if len(groupInfoFromMysql) > 0 {
			for _, v := range groupInfoFromMysql {
				var groupInfo open_im_sdk.GroupInfo
				cp.GroupDBCopyOpenIM(&groupInfo, v)
				groupsInfoList = append(groupsInfoList, &groupInfo)
			}
		}
	}

	resp := pbGroup.GetGroupsInfoResp{GroupInfoList: groupsInfoList}
	log.NewInfo(req.OperationID, "GetGroupsInfo rpc return ", resp.String())
	return &resp, nil
}

func (s *groupServer) GroupApplicationResponse(_ context.Context, req *pbGroup.GroupApplicationResponseReq) (*pbGroup.GroupApplicationResponseResp, error) {
	//log.NewInfo(req.OperationID, "GroupApplicationResponse args ", req.String())

	groupRequest := db.GroupRequest{}
	utils.CopyStructFields(&groupRequest, req)
	groupRequest.UserID = req.FromUserID
	groupRequest.HandleUserID = req.OpUserID
	groupRequest.HandledTime = time.Now()
	if !token_verify.IsManagerUserID(req.OpUserID) && !imdb.IsGroupOwnerAdmin(req.GroupID, req.OpUserID) {
		log.NewError(req.OperationID, "IsManagerUserID IsGroupOwnerAdmin false ", req.GroupID, req.OpUserID)
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	err := imdb.UpdateGroupRequest(groupRequest)
	if err != nil {
		// {openIM002 7836e478bc43ce1d3b8889cac983f59b 1  ok 0001-01-01 00:00:00 +0000 UTC openIM001 0001-01-01 00:00:00 +0000 UTC }
		log.NewError(req.OperationID, "GroupApplicationResponse failed ", err.Error(), groupRequest)
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	membersCount, _ := imdb.GetGroupMemberNumByGroupID(req.GroupID)
	membersCountInt := int(membersCount) + 1
	if membersCountInt > config.Config.MembersInGroupMaxLimit {
		log.NewError(req.OperationID, "Defined Limit of group memebers is excedeed")
		respErrMsg := "defined limit of group memebers is exceeded"
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.LimitExceeded.ErrCode, ErrMsg: respErrMsg}}, nil
	}

	if req.HandleResult == constant.GroupResponseAgree {
		user, err := imdb.GetUserByUserID(req.FromUserID)
		if err != nil {
			log.NewError(req.OperationID, "GroupApplicationResponse failed ", err.Error(), req.FromUserID)
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
		member := db.GroupMember{}
		member.GroupID = req.GroupID
		member.UserID = req.FromUserID
		member.RoleLevel = constant.GroupOrdinaryUsers
		member.OperatorUserID = req.OpUserID
		member.FaceURL = user.FaceURL
		member.Nickname = user.Nickname

		err = imdb.InsertIntoGroupMember(member)
		if err != nil {
			log.NewError(req.OperationID, "GroupApplicationResponse failed ", err.Error(), member)
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
		// set user group min seq to group max seq, no need to show previous messages to user
		maxSeq, _ := db.DB.GetGrupMaxSeq(member.GroupID)
		if maxSeq > 0 {
			_ = db.DB.SetGroupMinSeq(member.UserID, member.GroupID, uint32(maxSeq))
		}
		var reqPb pbUser.SetConversationReq
		reqPb.OperationID = req.OperationID
		var c pbUser.Conversation
		conversation, err := imdb.GetConversation(req.FromUserID, utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType))
		if err != nil {
			c.OwnerUserID = req.FromUserID
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsNotInGroup = false
		} else {
			c.OwnerUserID = conversation.OwnerUserID
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.RecvMsgOpt = conversation.RecvMsgOpt
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsPinned = conversation.IsPinned
			c.AttachedInfo = conversation.AttachedInfo
			c.IsPrivateChat = conversation.IsPrivateChat
			c.GroupAtType = conversation.GroupAtType
			c.IsNotInGroup = false
			c.Ex = conversation.Ex
		}
		reqPb.Conversation = &c
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
		}
		client := pbUser.NewUserClient(etcdConn)
		respPb, err := client.SetConversation(context.Background(), &reqPb)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error())
		} else {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String())
		}
		addGroupMemberToCacheReq := &pbCache.AddGroupMemberToCacheReq{OperationID: req.OperationID, GroupID: req.GroupID, UserIDList: []string{req.FromUserID}}
		etcdCacheConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
		}
		cacheClient := pbCache.NewCacheClient(etcdCacheConn)
		cacheResp, err := cacheClient.AddGroupMemberToCache(context.Background(), addGroupMemberToCacheReq)
		if err != nil {
			log.NewError(req.OperationID, "AddGroupMemberToCache rpc call failed ", err.Error())
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, "AddGroupMemberToCache rpc logic call failed ", cacheResp.String())
			return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}

		//data synchronization
		syncGroupRequestToLocal(req.OperationID, req.GroupID, req.FromUserID, constant.SyncGroupRequest)

		// cache
		_ = db.DB.DeleteInterestGroupINfoListByUserId(req.FromUserID)

		chat.GroupApplicationAcceptedNotification(req)
		chat.MemberEnterNotification(req)
	} else if req.HandleResult == constant.GroupResponseRefuse {
		//data synchronization
		syncGroupRequestToLocal(req.OperationID, req.GroupID, req.FromUserID, constant.SyncGroupRequest)

		chat.GroupApplicationRejectedNotification(req)
	} else {
		log.Error(req.OperationID, "HandleResult failed ", req.HandleResult)
		return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}

	log.NewInfo(req.OperationID, "rpc GroupApplicationResponse return ", pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{}})
	return &pbGroup.GroupApplicationResponseResp{CommonResp: &pbGroup.CommonResp{}}, nil
}

func (s *groupServer) JoinGroup(ctx context.Context, req *pbGroup.JoinGroupReq) (*pbGroup.JoinGroupResp, error) {
	//log.NewInfo(req.OperationID, "JoinGroup args ", req.String())
	_, err := imdb.GetUserByUserID(req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.OpUserID)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupID, err)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		errMsg := " group status is dismissed "
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
	}

	var groupRequest db.GroupRequest
	groupRequest.UserID = req.OpUserID
	groupRequest.ReqMsg = req.ReqMessage
	groupRequest.GroupID = req.GroupID

	err = imdb.InsertIntoGroupRequest(groupRequest)
	if err != nil {
		log.NewError(req.OperationID, "UpdateGroupRequest ", err.Error(), groupRequest)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	_, err = imdb.GetGroupMemberListByGroupIDAndRoleLevel(req.GroupID, constant.GroupOwner)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberListByGroupIDAndRoleLevel failed ", err.Error(), req.GroupID, constant.GroupOwner)
		return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
	}

	//data synchronization
	syncGroupRequestToLocal(req.OperationID, req.GroupID, req.OpUserID, constant.SyncGroupRequest)

	chat.JoinGroupApplicationNotification(req)

	log.NewInfo(req.OperationID, "ReceiveJoinApplicationNotification rpc return ")
	return &pbGroup.JoinGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) QuitGroup(ctx context.Context, req *pbGroup.QuitGroupReq) (*pbGroup.QuitGroupResp, error) {
	//log.NewInfo(req.OperationID, "QuitGroup args ", req.String())
	_, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed ", err.Error(), req.GroupID, req.OpUserID)
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	var deletedMemberList []*db.GroupMember
	member, err := imdb.GetGroupMemberByUserIDGroupID(req.GroupID, req.OpUserID)
	if err == nil && member.GroupID != "" {
		deletedMemberList = append(deletedMemberList, member)
	}

	err = imdb.DeleteGroupMemberByGroupIDAndUserID(req.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "DeleteGroupMemberByGroupIdAndUserId failed ", err.Error(), req.GroupID, req.OpUserID)
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//modify quitter conversation info
	var reqPb pbUser.SetConversationReq
	var c pbUser.Conversation
	reqPb.OperationID = req.OperationID
	c.OwnerUserID = req.OpUserID
	c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
	c.ConversationType = constant.GroupChatType
	c.GroupID = req.GroupID
	c.IsNotInGroup = true
	reqPb.Conversation = &c
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}
	client := pbUser.NewUserClient(etcdConn)
	respPb, err := client.SetConversation(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error())
	} else {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String())
	}

	reduceGroupMemberFromCacheReq := &pbCache.ReduceGroupMemberFromCacheReq{
		UserIDList:  []string{req.OpUserID},
		GroupID:     req.GroupID,
		OperationID: req.OperationID,
	}
	etcdConnCache := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}
	cacheClient := pbCache.NewCacheClient(etcdConnCache)
	cacheResp, err := cacheClient.ReduceGroupMemberFromCache(context.Background(), reduceGroupMemberFromCacheReq)
	if err != nil {
		log.NewError(req.OperationID, "ReduceGroupMemberFromCache rpc call failed ", err.Error())
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationID, "ReduceGroupMemberFromCache rpc logic call failed ", cacheResp.String())
		return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupID, constant.SyncKickedGroupMember, []string{req.OpUserID}, nil, deletedMemberList)

	chat.MemberQuitNotification(req)
	log.NewInfo(req.OperationID, "rpc QuitGroup return ", pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}})
	return &pbGroup.QuitGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func hasAccess(req *pbGroup.SetGroupInfoReq) bool {
	if utils.IsContain(req.OpUserID, config.Config.Manager.AppManagerUid) {
		return true
	}
	groupUserInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupInfo.GroupID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed, ", err.Error(), req.GroupInfo.GroupID, req.OpUserID)
		return false

	}
	if groupUserInfo.RoleLevel == constant.GroupOwner || groupUserInfo.RoleLevel == constant.GroupAdmin {
		return true
	}
	return false
}

func (s *groupServer) SetGroupInfo(ctx context.Context, req *pbGroup.SetGroupInfoReq) (*pbGroup.SetGroupInfoResp, error) {
	log.NewInfo(req.OperationID, "SetGroupInfo args ", req.String())
	if !hasAccess(req) && req.IsAdmin == false {
		log.NewError(req.OperationID, "no access ", req)
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	group, err := imdb.GetGroupInfoByGroupID(req.GroupInfo.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), req.GroupInfo.GroupID)
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, http.WrapError(constant.ErrDB)
	}

	if group.Status == constant.GroupStatusDismissed {
		errMsg := " group status is dismissed "
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
	}

	// get group owner
	groupOwner, err := imdb.GetGroupMaster(req.GroupInfo.GroupID)

	// //bitwise operators: 0001:groupName; 0010:Notification  0100:Introduction; 1000:FaceUrl; 10000:owner
	var changedType int32
	groupName := ""
	notification := ""
	introduction := ""
	faceURL := ""
	if group.GroupName != req.GroupInfo.GroupName && req.GroupInfo.GroupName != "" {
		changedType = 1
		groupName = req.GroupInfo.GroupName
	}
	if group.Notification != req.GroupInfo.Notification && req.GroupInfo.Notification != "" {
		changedType = changedType | (1 << 1)
		notification = req.GroupInfo.Notification
	}
	if group.Introduction != req.GroupInfo.Introduction && req.GroupInfo.Introduction != "" {
		changedType = changedType | (1 << 2)
		introduction = req.GroupInfo.Introduction
	}
	if group.FaceURL != req.GroupInfo.FaceURL && req.GroupInfo.FaceURL != "" {
		changedType = changedType | (1 << 3)
		faceURL = req.GroupInfo.FaceURL
	}
	// only administrators can set group information
	var groupInfo db.Group
	utils.CopyStructFields(&groupInfo, req.GroupInfo)
	groupInfo.GroupName = req.GroupInfo.GroupName
	err = imdb.SetGroupInfo(groupInfo)
	if err != nil {
		log.NewError(req.OperationID, "SetGroupInfo failed ", err.Error(), groupInfo)
		return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, http.WrapError(constant.ErrDB)
	}
	log.NewInfo(req.OperationID, "SetGroupInfo rpc return ", pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}})
	if changedType != 0 {
		var opFrom int32 = 0
		if req.IsAdmin {
			opFrom = constant.OpFromAdmin
		}
		chat.GroupInfoSetNotification(req.OperationID, groupOwner.UserID, req.GroupInfo.GroupID, groupName, notification, introduction, faceURL, opFrom)
	}

	// face url
	if faceURL != "" {
		client, err := admin_cms.GetTencentCloudClient(true)
		if err == nil {
			admin_cms.RemoveDeleteTagForPersistent(client, []string{faceURL})
		}
	}

	// master id

	// set group interest type
	if len(req.GroupInterest) > 0 {
		imdb.SetGroupInterestType(req.GroupInfo.GroupID, req.GroupInterest)
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupInfo.GroupID, constant.SyncUpdateGroup, nil, nil, nil)

	if req.GroupInfo.Notification != "" {
		// get group member user id
		getGroupMemberIDListFromCacheReq := &pbCache.GetGroupMemberIDListFromCacheReq{OperationID: req.OperationID, GroupID: req.GroupInfo.GroupID}
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, http.WrapError(constant.ErrInternal)
		}
		client := pbCache.NewCacheClient(etcdConn)
		cacheResp, err := client.GetGroupMemberIDListFromCache(context.Background(), getGroupMemberIDListFromCacheReq)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberIDListFromCache rpc call failed ", err.Error())
			return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}}, nil
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, "GetGroupMemberIDListFromCache rpc logic call failed ", cacheResp.String())
			return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}}, nil
		}
		var conversationReq pbConversation.ModifyConversationFieldReq

		conversation := pbConversation.Conversation{
			OwnerUserID:      groupOwner.UserID,
			ConversationID:   utils.GetConversationIDBySessionType(req.GroupInfo.GroupID, constant.GroupChatType),
			ConversationType: constant.GroupChatType,
			GroupID:          req.GroupInfo.GroupID,
		}
		if req.IsAdmin {
			conversationReq.OpFrom = constant.OpFromAdmin
		}
		conversationReq.Conversation = &conversation
		conversationReq.OperationID = req.OperationID
		conversationReq.FieldType = constant.FieldGroupAtType
		conversation.GroupAtType = constant.GroupNotification
		conversationReq.UserIDList = cacheResp.UserIDList
		nEtcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImConversationName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, http.WrapError(constant.ErrInternal)
		}
		nClient := pbConversation.NewConversationClient(nEtcdConn)
		conversationReply, err := nClient.ModifyConversationField(context.Background(), &conversationReq)
		if err != nil {
			log.NewError(conversationReq.OperationID, "ModifyConversationField rpc failed, ", conversationReq.String(), err.Error())
		} else if conversationReply.CommonResp.ErrCode != 0 {
			log.NewError(conversationReq.OperationID, "ModifyConversationField rpc failed, ", conversationReq.String(), conversationReply.String())
		}
	}

	return &pbGroup.SetGroupInfoResp{CommonResp: &pbGroup.CommonResp{}}, nil
}

func (s *groupServer) TransferGroupOwner(_ context.Context, req *pbGroup.TransferGroupOwnerReq) (*pbGroup.TransferGroupOwnerResp, error) {
	//log.NewInfo(req.OperationID, "TransferGroupOwner ", req.String())

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupID, err)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		errMsg := " group status is dismissed "
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
	}

	if req.OldOwnerUserID == req.NewOwnerUserID {
		log.NewError(req.OperationID, "same owner ", req.OldOwnerUserID, req.NewOwnerUserID)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}

	// check oldOwner and OpUser
	oldOwner, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OldOwnerUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed ", oldOwner)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, err
	}

	// if the user is not the owner will no be allowed
	if oldOwner != nil && oldOwner.RoleLevel != 2 {
		log.NewError(req.OperationID, "oldOwner is not the owner ", oldOwner)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, err
	}

	if req.OpFrom == constant.OpFromFrontend {
		opUser, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID opUser failed ", opUser)
			return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, err
		}

		// if the user is not the owner will no be allowed
		if opUser != nil && opUser.RoleLevel != 2 {
			log.NewError(req.OperationID, "opUser is not the owner ", opUser)
			return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
		}
	}

	// if the group has another owner
	ownerNums, err := imdb.GetOwnerNumByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "get ownerNums failed", req.OpUserID)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "get ownerNums failed"}}, err
	}
	if ownerNums > 1 {
		log.NewError(req.OperationID, "the group has another owner", req.OpUserID)
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "the group has another owner"}}, nil
	}

	err = imdb.TransferGroupOwner(req.GroupID, req.OldOwnerUserID, req.NewOwnerUserID)
	if err != nil {
		log.NewError(req.OperationID, "TransferGroupOwner failed ", err.Error())
		return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}}, nil
	}

	//data synchronization
	//syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncUpdateGroup, nil, nil, nil)
	//syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncGroupMemberInfo, nil, []string{req.OldOwnerUserID, req.NewOwnerUserID}, nil)

	chat.GroupOwnerTransferredNotification(req)

	return &pbGroup.TransferGroupOwnerResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil

}

func (s *groupServer) TransferGroupAdminCMS(_ context.Context, req *pbGroup.TransferGroupAdminReq) (*pbGroup.TransferGroupAdminResp, error) {
	//log.NewInfo(req.OperationID, "TransferGroupOwner ", req.String())

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupID, err)
		return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		errMsg := " group status is dismissed "
		return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
	}

	if req.OwnerUserID == req.UserID {
		log.NewError(req.OperationID, "The owner can't be set as admin", req.OwnerUserID, req.UserID)
		return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}

	if req.OpFrom == constant.OpFromFrontend {
		// check oldOwner and OpUser
		oldOwner, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed ", oldOwner)
			return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, err
		}

		// if the user is not the owner will no be allowed
		if oldOwner != nil && oldOwner.RoleLevel != constant.GroupOwner {
			log.NewError(req.OperationID, "oldOwner is not the owner ", oldOwner)
			return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
		}
	}

	err = imdb.TransferGroupAdmin(req.GroupID, req.UserID)
	if err != nil {
		log.NewError(req.OperationID, "TransferGroupAdmin failed ", err.Error())
		return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}}, err
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncUpdateGroup, nil, nil, nil)
	syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncGroupMemberInfo, nil, []string{req.UserID, req.OpUserID}, nil)

	chat.GroupMemberRoleLevelChangeNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, constant.GroupMemberSetToAdminNotification, req.OpFrom)

	return &pbGroup.TransferGroupAdminResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil

}

func (s *groupServer) TransferGroupOrdinaryCMS(_ context.Context, req *pbGroup.TransferGroupOrdinaryReq) (*pbGroup.TransferGroupOrdinaryResp, error) {
	//log.NewInfo(req.OperationID, "TransferGroupOwner ", req.String())

	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupID, err)
		return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	if groupInfo.Status == constant.GroupStatusDismissed {
		errMsg := " group status is dismissed "
		return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrStatus.ErrCode, ErrMsg: errMsg}}, nil
	}

	if req.OwnerUserID == req.UserID {
		log.NewError(req.OperationID, "The owner can't be set as ordinary", req.OwnerUserID, req.UserID)
		return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}

	if req.OpFrom == constant.OpFromFrontend {
		// check oldOwner and OpUser
		opUser, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.OpUserID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberInfoByGroupIDAndUserID failed ", opUser)
			return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, err
		}

		// if the user is not the owner will no be allowed
		if opUser != nil && opUser.RoleLevel != constant.GroupOwner && opUser.RoleLevel != constant.GroupAdmin {
			log.NewError(req.OperationID, "oldOwner is not the owner ", opUser)
			return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
		}
	}

	err = imdb.TransferGroupOrdinary(req.GroupID, req.UserID)
	if err != nil {
		log.NewError(req.OperationID, "TransferGroupOrdinary failed ", err.Error())
		return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}}, err
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncUpdateGroup, nil, nil, nil)
	syncGroupToLocal(req.OperationID, groupInfo.GroupID, constant.SyncGroupMemberInfo, nil, []string{req.UserID, req.OpUserID}, nil)

	chat.GroupMemberRoleLevelChangeNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, constant.GroupMemberSetToOrdinaryUserNotification, req.OpFrom)

	return &pbGroup.TransferGroupOrdinaryResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil

}

func (s *groupServer) GetGroupById(_ context.Context, req *pbGroup.GetGroupByIdReq) (*pbGroup.GetGroupByIdResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbGroup.GetGroupByIdResp{CMSGroup: &pbGroup.CMSGroup{
		GroupInfo: &open_im_sdk.GroupInfo{},
	}}
	group, err := imdb.GetGroupById(req.GroupId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupById error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.CMSGroup.GroupInfo = &open_im_sdk.GroupInfo{
		GroupID:       group.GroupID,
		GroupName:     group.GroupName,
		FaceURL:       group.FaceURL,
		OwnerUserID:   group.CreatorUserID,
		MemberCount:   0,
		Introduction:  group.Introduction,
		Notification:  group.Notification,
		Status:        group.Status,
		CreatorUserID: group.CreatorUserID,
		GroupType:     group.GroupType,
		CreateTime:    uint32(group.CreateTime.Unix()),
		Remark:        group.Remark,
		VideoStatus:   int32(group.VideoStatus),
		AudioStatus:   int32(group.AudioStatus),
	}
	groupMember, err := imdb.GetGroupMaster(group.GroupID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	members, _, err := imdb.GetGroupMembersByGroupIdCMS(group.GroupID, "", -1, 1)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMembersByGroupIdCMS", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.CMSGroup.MemberList = make([]*pbGroup.MemberSimple, len(members))
	for index, member := range members {
		resp.CMSGroup.MemberList[index] = &pbGroup.MemberSimple{UserID: member.UserID, UserName: member.Nickname}
	}

	resp.CMSGroup.IsOpen = group.IsOpen
	resp.CMSGroup.GroupMasterName = groupMember.Nickname
	resp.CMSGroup.GroupMasterId = groupMember.UserID
	resp.CMSGroup.GroupInfo.CreatorUserID = group.CreatorUserID
	count, _ := imdb.GetGroupMembersCount(group.GroupID, "")
	resp.CMSGroup.GroupInfo.MemberCount = uint32(count)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *groupServer) GetGroup(_ context.Context, req *pbGroup.GetGroupReq) (*pbGroup.GetGroupResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbGroup.GetGroupResp{
		CMSGroups: []*pbGroup.CMSGroup{},
	}
	groups, err := imdb.GetGroupsByName(req.GroupName, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupsByName error", req.String())
		return resp, http.WrapError(constant.ErrDB)
	}
	nums, err := imdb.GetGroupsCountNum(db.Group{GroupName: req.GroupName})
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupsCountNum error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	resp.GroupNums = int32(nums)
	resp.Pagination = &open_im_sdk.RequestPagination{
		PageNumber: req.Pagination.PageNumber,
		ShowNumber: req.Pagination.ShowNumber,
	}
	for _, v := range groups {
		groupMember, err := imdb.GetGroupMaster(v.GroupID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster error", err.Error())
			continue
		}
		resp.CMSGroups = append(resp.CMSGroups, &pbGroup.CMSGroup{
			GroupInfo: &open_im_sdk.GroupInfo{
				GroupID:       v.GroupID,
				GroupName:     v.GroupName,
				FaceURL:       v.FaceURL,
				OwnerUserID:   v.CreatorUserID,
				Status:        v.Status,
				CreatorUserID: v.CreatorUserID,
				CreateTime:    uint32(v.CreateTime.Unix()),
			},
			GroupMasterName: groupMember.Nickname,
			GroupMasterId:   groupMember.UserID,
		})
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *groupServer) GetGroups(_ context.Context, req *pbGroup.GetGroupsReq) (*pbGroup.GetGroupsResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroups ", req.String())
	resp := &pbGroup.GetGroupsResp{
		CMSGroups:  []*pbGroup.CMSGroup{},
		Pagination: &open_im_sdk.RequestPagination{},
	}

	where := map[string]interface{}{}
	where["is_open"] = req.IsOpen
	where["owner"] = req.Owner
	where["creator"] = req.Creator

	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["remark"] = req.Remark

	if req.Group != "" {
		where["group"] = req.Group
	}
	var statusList []string
	if req.GroupStatus != "" {
		err := json.Unmarshal([]byte(req.GroupStatus), &statusList)
		if err != nil {
			return resp, err
		}
	}

	if req.Member != "" {
		users, _, _ := imdb.GetUsersByWhere(map[string]string{"user_id": req.Member}, nil, -1, 1, req.OrderBy)
		if len(users) > 0 {
			var usersArr []string
			for _, v := range users {
				usersArr = append(usersArr, v.UserID)
			}
			if len(usersArr) > 0 {
				where["group_id_array"] = imdb.GetGroupIdByMemberIds(usersArr)
			}
		}
	}

	groups, groupsCountNum, err := imdb.GetGroupsByWhere(where, statusList, int(req.Pagination.PageNumber), int(req.Pagination.ShowNumber), req.OrderBy)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroups error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}

	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "groupsCountNum ", groupsCountNum)
	resp.GroupNum = int32(groupsCountNum)
	resp.Pagination.PageNumber = req.Pagination.PageNumber
	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	for _, v := range groups {
		groupMember, err := imdb.GetGroupMaster(v.GroupID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster failed", err.Error(), v)
			continue
		}
		count, err := imdb.GetGroupMembersCount(v.GroupID, "")
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "Member count failed", err.Error(), v)
			continue
		}
		resp.CMSGroups = append(resp.CMSGroups, &pbGroup.CMSGroup{
			GroupInfo: &open_im_sdk.GroupInfo{
				GroupID:       v.GroupID,
				GroupName:     v.GroupName,
				FaceURL:       v.FaceURL,
				OwnerUserID:   v.CreatorUserID,
				Status:        v.Status,
				CreatorUserID: v.CreatorUserID,
				CreateTime:    uint32(v.CreateTime.Unix()),
				Ex:            v.Ex,
				GroupType:     v.GroupType,
				Remark:        v.Remark,
				MemberCount:   uint32(count),
				VideoStatus:   int32(v.VideoStatus),
				AudioStatus:   int32(v.AudioStatus),
				Notification:  v.Notification,
				Introduction:  v.Introduction,
			},
			GroupMasterId:   groupMember.UserID,
			GroupMasterName: groupMember.Nickname,
			IsOpen:          v.IsOpen,
		})
	}
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroups ", resp.String())
	return resp, nil
}

func (s *groupServer) OperateGroupStatus(_ context.Context, req *pbGroup.OperateGroupStatusReq) (*pbGroup.OperateGroupStatusResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbGroup.OperateGroupStatusResp{}
	if err := imdb.OperateGroupStatus(req.GroupId, req.Status); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "OperateGroupStatus", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *groupServer) DeleteGroup(_ context.Context, req *pbGroup.DeleteGroupReq) (*pbGroup.DeleteGroupResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())

	resp := &pbGroup.DeleteGroupResp{}

	memberIDList, _ := imdb.GetGroupMemberIDListByGroupID(req.GroupId)
	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupId)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.GroupId, err.Error())
		return resp, err
	}

	groupInfo.Status = constant.GroupStatusDismissed

	// okUserIDList := []string{}
	if groupInfo.GroupType != constant.SuperGroup {
		// memberList, err := imdb.GetGroupMemberListByGroupID(req.GroupId)
		//if err != nil {
		//	log.NewError(req.OperationID, "GetGroupMemberListByGroupID failed,", err.Error(), req.GroupId)
		//}
		//modify quitter conversation info
		// var reqPb pbUser.SetConversationReq
		// var c pbUser.Conversation
		// for _, v := range memberList {
		// 	okUserIDList = append(okUserIDList, v.UserID)
		// 	reqPb.OperationID = req.OperationID
		// 	c.OwnerUserID = v.UserID
		// 	c.ConversationID = utils.GetConversationIDBySessionType(req.GroupId, constant.GroupChatType)
		// 	c.ConversationType = constant.GroupChatType
		// 	c.GroupID = req.GroupId
		// 	c.IsNotInGroup = true
		// 	reqPb.Conversation = &c
		// 	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
		// 	if etcdConn == nil {
		// 		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		// 		log.NewError(req.OperationID, errMsg)
		// 		return resp, http.WrapError(constant.ErrRPC)
		// 	}
		// 	client := pbUser.NewUserClient(etcdConn)
		// 	respPb, err := client.SetConversation(context.Background(), &reqPb)
		// 	if err != nil {
		// 		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v.UserID)
		// 	} else {
		// 		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v.UserID)
		// 	}
		// }
		err = imdb.DeleteGroupMemberByGroupID(req.GroupId)
		if err != nil {
			log.NewError(req.OperationID, "DeleteGroupMemberByGroupID failed ", req.GroupId)
			return resp, err
		}

		err = imdb.DeleteGroupInterests(req.GroupId)
		if err != nil {
			log.NewError(req.OperationID, "DeleteGroupInterests failed ", req.GroupId)
			return resp, err
		}

		chat.GroupDeleteNotification(req)

	}

	if err := imdb.DeleteGroup(req.GroupId); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteGroup error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}

	//data synchronization
	if memberIDList != nil {
		syncDeleteGroupToLocal(req.OperationID, groupInfo, memberIDList)
	}

	return resp, nil
}

func (s *groupServer) OperateUserRole(_ context.Context, req *pbGroup.OperateUserRoleReq) (*pbGroup.OperateUserRoleResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.OperateUserRoleResp{}
	oldOwnerMember, err := imdb.GetGroupMaster(req.GroupId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMaster failed", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	if oldOwnerMember.GroupID == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "The group doesn't exist")
		return resp, http.WrapError(constant.ErrArgs)
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return resp, http.WrapError(constant.ErrInternal)
	}
	client := pbGroup.NewGroupClient(etcdConn)
	switch req.RoleLevel {
	case constant.GroupOwner:
		var reqPb pbGroup.TransferGroupOwnerReq
		reqPb.OperationID = req.OperationID
		reqPb.NewOwnerUserID = req.UserId
		reqPb.GroupID = req.GroupId
		reqPb.OpUserID = oldOwnerMember.UserID
		reqPb.OldOwnerUserID = oldOwnerMember.UserID
		reqPb.OpFrom = req.OpFrom
		reply, err := client.TransferGroupOwner(context.Background(), &reqPb)
		if err != nil || reply.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "TransferGroupOwner rpc failed")
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			}
		}
	case constant.GroupAdmin:
		var reqPb pbGroup.TransferGroupAdminReq
		reqPb.OperationID = req.OperationID
		reqPb.OwnerUserID = oldOwnerMember.UserID
		reqPb.UserID = req.UserId
		reqPb.GroupID = req.GroupId
		reqPb.OpUserID = req.OpUserID
		reqPb.OpFrom = req.OpFrom
		reply, err := client.TransferGroupAdminCMS(context.Background(), &reqPb)
		if err != nil || reply.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "TransferGroupOwner rpc failed")
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			}
		}
	case constant.GroupOrdinaryUsers:
		var reqPb pbGroup.TransferGroupOrdinaryReq
		reqPb.OperationID = req.OperationID
		reqPb.OwnerUserID = oldOwnerMember.UserID
		reqPb.GroupID = req.GroupId
		reqPb.UserID = req.UserId
		reqPb.OpUserID = req.OpUserID
		reqPb.OpFrom = req.OpFrom
		reply, err := client.TransferGroupOrdinaryCMS(context.Background(), &reqPb)
		if err != nil || reply.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "TransferGroupOwner rpc failed")
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			}
		}
	default:
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupId, constant.SyncUpdateGroup, nil, nil, nil)
	syncGroupToLocal(req.OperationID, req.GroupId, constant.SyncGroupMemberInfo, nil, []string{req.UserId, oldOwnerMember.UserID}, nil)

	return resp, err
}

func (s *groupServer) GetGroupMembersCMS(_ context.Context, req *pbGroup.GetGroupMembersCMSReq) (*pbGroup.GetGroupMembersCMSResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.GetGroupMembersCMSResp{}

	where := map[string]interface{}{}
	where["member"] = req.Member
	where["remark_name"] = req.RemarkName
	where["remark"] = req.Remark
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["permission"] = req.Permission
	where["status"] = req.Status
	where["role_level"] = req.RoleLevel

	groupMembers, groupMembersCount, err := imdb.GetGroupMemberByWhere(req.GroupID, where, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMembersByGroupIdCMS Error", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}

	//log.NewInfo(req.OperationID, groupMembersCount)
	resp.MemberNums = int32(groupMembersCount)
	g, _ := imdb.GetGroupById(req.GroupID)
	resp.GroupName = g.GroupName
	nowTime := time.Now().Unix()
	for _, groupMember := range groupMembers {
		var muteTime uint32 = 0
		if nowTime < groupMember.MuteEndTime.Unix() {
			muteTime = uint32(groupMember.MuteEndTime.Unix() - nowTime)
		}
		userName, _ := imdb.GetUserNameByUserID(groupMember.UserID)
		resp.Members = append(resp.Members, &open_im_sdk.GroupMemberFullInfo{
			GroupID:     req.GroupID,
			MemberName:  userName,
			UserID:      groupMember.UserID,
			RoleLevel:   groupMember.RoleLevel,
			JoinTime:    int32(groupMember.JoinTime.Unix()),
			Nickname:    groupMember.Nickname,
			FaceURL:     groupMember.FaceURL,
			JoinSource:  groupMember.JoinSource,
			Remark:      groupMember.Remark,
			VideoStatus: int32(groupMember.VideoStatus),
			AudioStatus: int32(groupMember.AudioStatus),
			MuteEndTime: muteTime,
		})
	}
	resp.Pagination = &open_im_sdk.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (s *groupServer) RemoveGroupMembersCMS(_ context.Context, req *pbGroup.RemoveGroupMembersCMSReq) (*pbGroup.RemoveGroupMembersCMSResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args:", req.String())
	resp := &pbGroup.RemoveGroupMembersCMSResp{}
	var kickedMemberList []*db.GroupMember
	for _, userId := range req.UserIds {
		member, err := imdb.GetGroupMemberByUserIDGroupID(req.GroupId, userId)
		if err == nil && member.GroupID != "" {
			kickedMemberList = append(kickedMemberList, member)
		}

		err = imdb.RemoveGroupMember(req.GroupId, userId)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			resp.Failed = append(resp.Failed, userId)
		} else {
			resp.Success = append(resp.Success, userId)
		}
	}
	reqKick := &pbGroup.KickGroupMemberReq{
		GroupID:          req.GroupId,
		KickedUserIDList: resp.Success,
		Reason:           "admin kick",
		OperationID:      req.OperationID,
		OpUserID:         req.OpUserId,
	}
	var reqPb pbUser.SetConversationReq
	var c pbUser.Conversation
	for _, v := range resp.Success {
		reqPb.OperationID = req.OperationID
		c.OwnerUserID = v
		c.ConversationID = utils.GetConversationIDBySessionType(req.GroupId, constant.GroupChatType)
		c.ConversationType = constant.GroupChatType
		c.GroupID = req.GroupId
		c.IsNotInGroup = true
		reqPb.Conversation = &c
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return resp, http.WrapError(constant.ErrInternal)
		}
		client := pbUser.NewUserClient(etcdConn)
		respPb, err := client.SetConversation(context.Background(), &reqPb)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v)
		} else {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
		}
	}

	reduceGroupMemberFromCacheReq := &pbCache.ReduceGroupMemberFromCacheReq{
		UserIDList:  resp.Success,
		GroupID:     req.GroupId,
		OperationID: req.OperationID,
	}
	etcdConnCache := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConnCache == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return resp, http.WrapError(constant.ErrInternal)
	}
	cacheClient := pbCache.NewCacheClient(etcdConnCache)
	cacheResp, err := cacheClient.ReduceGroupMemberFromCache(context.Background(), reduceGroupMemberFromCacheReq)
	if err != nil {
		log.NewError(req.OperationID, "ReduceGroupMemberFromCache rpc call failed ", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationID, "ReduceGroupMemberFromCache rpc logic call failed ", cacheResp.String())
		return resp, http.WrapError(constant.ErrDB)
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupId, constant.SyncKickedGroupMember, req.UserIds, nil, kickedMemberList)

	chat.MemberKickedNotification(reqKick, resp.Success)
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	return resp, nil
}

func (s *groupServer) AddGroupMembersCMS(_ context.Context, req *pbGroup.AddGroupMembersCMSReq) (*pbGroup.AddGroupMembersCMSResp, error) {
	//log.NewInfo(req.OperationId, utils.GetSelfFuncName(), "args:", req.String())

	membersCount, _ := imdb.GetGroupMemberNumByGroupID(req.GroupId)

	resp := &pbGroup.AddGroupMembersCMSResp{}
	for _, userId := range req.UserIds {
		membersCount = membersCount + 1
		if int(membersCount) > config.Config.MembersInGroupMaxLimit {
			log.NewError(req.OperationId, "Defined Limit of group memebers is excedeed", userId, req.GroupId)
			resp.Failed = append(resp.Failed, userId)
			continue
		}
		if isExist := imdb.IsExistGroupMember(req.GroupId, userId); isExist {
			log.NewError(req.OperationId, utils.GetSelfFuncName(), "user is exist in group", userId, req.GroupId)
			resp.Failed = append(resp.Failed, userId)
			continue
		}
		user, err := imdb.GetUserByUserID(userId)
		if err != nil {
			log.NewError(req.OperationId, utils.GetSelfFuncName(), "GetUserByUserID", err.Error())
			resp.Failed = append(resp.Failed, userId)
			continue
		}
		if user != nil {
			groupMember := db.GroupMember{
				GroupID:        req.GroupId,
				UserID:         userId,
				Nickname:       user.Nickname,
				FaceURL:        user.FaceURL,
				RoleLevel:      1,
				JoinTime:       time.Time{},
				JoinSource:     constant.JoinByAdmin,
				OperatorUserID: "CmsAdmin",
				Ex:             "",
			}
			if err := imdb.InsertIntoGroupMember(groupMember); err != nil {
				log.NewError(req.OperationId, utils.GetSelfFuncName(), "InsertIntoGroupMember failed", req.String())
				resp.Failed = append(resp.Failed, userId)
			} else {
				resp.Success = append(resp.Success, userId)
				// set user group min seq to group max seq, no need to show previous messages to user
				maxSeq, _ := db.DB.GetGrupMaxSeq(groupMember.GroupID)
				if maxSeq > 0 {
					_ = db.DB.SetGroupMinSeq(groupMember.UserID, groupMember.GroupID, uint32(maxSeq))
				}
			}
		}
	}

	//modify quitter conversation info
	var reqPb pbUser.SetConversationReq
	var c pbUser.Conversation
	for _, v := range resp.Success {
		reqPb.OperationID = req.OperationId
		c.OwnerUserID = v
		c.ConversationID = utils.GetConversationIDBySessionType(req.GroupId, constant.GroupChatType)
		c.ConversationType = constant.GroupChatType
		c.GroupID = req.GroupId
		c.IsNotInGroup = false
		reqPb.Conversation = &c
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationId)
		if etcdConn == nil {
			errMsg := req.OperationId + "getcdv3.GetConn == nil"
			log.NewError(req.OperationId, errMsg)
			return resp, http.WrapError(constant.ErrInternal)
		}
		client := pbUser.NewUserClient(etcdConn)
		respPb, err := client.SetConversation(context.Background(), &reqPb)
		if err != nil {
			log.NewError(req.OperationId, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v)
		} else {
			log.NewDebug(req.OperationId, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
		}
	}

	//data synchronization
	syncGroupToLocal(req.OperationId, req.GroupId, constant.SyncInvitedGroupMember, nil, req.UserIds, nil)

	addGroupMemberToCacheReq := &pbCache.AddGroupMemberToCacheReq{
		UserIDList:  resp.Success,
		GroupID:     req.GroupId,
		OperationID: req.OperationId,
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationId)
	if etcdConn == nil {
		errMsg := req.OperationId + "getcdv3.GetConn == nil"
		log.NewError(req.OperationId, errMsg)
		return resp, http.WrapError(constant.ErrInternal)
	}
	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.AddGroupMemberToCache(context.Background(), addGroupMemberToCacheReq)
	if err != nil {
		log.NewError(req.OperationId, "AddGroupMemberToCache rpc call failed ", err.Error())
		return resp, http.WrapError(constant.ErrDB)
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationId, "AddGroupMemberToCache rpc logic call failed ", cacheResp.String())
		return resp, http.WrapError(constant.ErrDB)
	}

	chat.MemberInvitedNotification(req.OperationId, req.GroupId, req.OpUserId, "admin add you to group", resp.Success)
	return resp, nil
}

func (s *groupServer) GetUserReqApplicationList(_ context.Context, req *pbGroup.GetUserReqApplicationListReq) (*pbGroup.GetUserReqApplicationListResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbGroup.GetUserReqApplicationListResp{}
	groupRequests, err := imdb.GetUserReqGroupByUserID(req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserReqGroupByUserID failed ", err.Error())
		resp.CommonResp = &pbGroup.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return resp, nil
	}
	for _, groupReq := range groupRequests {
		node := open_im_sdk.GroupRequest{UserInfo: &open_im_sdk.PublicUserInfo{}, GroupInfo: &open_im_sdk.GroupInfo{}}
		group, err := imdb.GetGroupInfoByGroupID(groupReq.GroupID)
		if err != nil {
			log.Error(req.OperationID, "GetGroupInfoByGroupID failed ", err.Error(), groupReq.GroupID)
			continue
		}
		user, err := imdb.GetUserByUserID(groupReq.UserID)
		if err != nil {
			log.Error(req.OperationID, "GetUserByUserID failed ", err.Error(), groupReq.UserID)
			continue
		}
		cp.GroupRequestDBCopyOpenIM(&node, &groupReq)
		cp.UserDBCopyOpenIMPublicUser(node.UserInfo, user)
		cp.GroupDBCopyOpenIM(node.GroupInfo, group)
		resp.GroupRequestList = append(resp.GroupRequestList, &node)
	}
	resp.CommonResp = &pbGroup.CommonResp{
		ErrCode: 0,
		ErrMsg:  "",
	}
	return resp, nil
}

func (s *groupServer) DismissGroup(ctx context.Context, req *pbGroup.DismissGroupReq) (*pbGroup.DismissGroupResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc args ", req.String())
	if !req.IsAdminUser && !imdb.IsGroupOwnerAdmin(req.GroupID, req.OpUserID) {
		log.NewError(req.OperationID, "verify failed ", req.OpUserID, req.GroupID)
		return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	err := imdb.OperateGroupStatus(req.GroupID, constant.GroupStatusDismissed)
	if err != nil {
		log.NewError(req.OperationID, "OperateGroupStatus failed ", req.GroupID, constant.GroupStatusDismissed)
		return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	groupInfo, err := imdb.GetGroupInfoByGroupID(req.GroupID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error(), req.GroupID)
		return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	okUserIDList := []string{}
	if groupInfo.GroupType != constant.SuperGroup {
		memberList, err := imdb.GetGroupMemberListByGroupID(req.GroupID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberListByGroupID failed,", err.Error(), req.GroupID)
		}
		// modify quitter conversation info
		var reqPb pbUser.SetConversationReq
		var c pbUser.Conversation
		for _, v := range memberList {
			okUserIDList = append(okUserIDList, v.UserID)
			reqPb.OperationID = req.OperationID
			c.OwnerUserID = v.UserID
			c.ConversationID = utils.GetConversationIDBySessionType(req.GroupID, constant.GroupChatType)
			c.ConversationType = constant.GroupChatType
			c.GroupID = req.GroupID
			c.IsNotInGroup = true
			reqPb.Conversation = &c
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
			if etcdConn == nil {
				errMsg := req.OperationID + "getcdv3.GetConn == nil"
				log.NewError(req.OperationID, errMsg)
				return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
			}
			client := pbUser.NewUserClient(etcdConn)
			respPb, err := client.SetConversation(context.Background(), &reqPb)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", reqPb.String(), err.Error(), v.UserID)
			} else {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v.UserID)
			}
		}
		err = imdb.DeleteGroupMemberByGroupID(req.GroupID)
		if err != nil {
			log.NewError(req.OperationID, "DeleteGroupMemberByGroupID failed ", req.GroupID)
			return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}

		//data synchronization
		syncDeleteGroupToLocal(req.OperationID, groupInfo, okUserIDList)

		chat.GroupDismissedNotification(req)

	} else {
		err = db.DB.DeleteSuperGroup(req.GroupID)
		if err != nil {
			log.NewError(req.OperationID, "DeleteGroupMemberByGroupID failed ", req.GroupID)
			return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil

		}
	}
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return ", pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""})
	return &pbGroup.DismissGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

//  rpc MuteGroupMember(MuteGroupMemberReq) returns(MuteGroupMemberResp);
//  rpc CancelMuteGroupMember(CancelMuteGroupMemberReq) returns(CancelMuteGroupMemberResp);
//  rpc MuteGroup(MuteGroupReq) returns(MuteGroupResp);
//  rpc CancelMuteGroup(CancelMuteGroupReq) returns(CancelMuteGroupResp);

func (s *groupServer) MuteGroupMember(ctx context.Context, req *pbGroup.MuteGroupMemberReq) (*pbGroup.MuteGroupMemberResp, error) {

	groupMemberInfo := db.GroupMember{GroupID: req.GroupID, UserID: req.UserID}

	groupMemberInfo.MuteEndTime = time.Unix(int64(time.Now().Second())+int64(req.MutedSeconds), time.Now().UnixNano())
	err := imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		log.Error(req.OperationID, "UpdateGroupMemberInfo failed ", err.Error(), groupMemberInfo)
		return &pbGroup.MuteGroupMemberResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupID, constant.SyncMuteGroupMember, nil, []string{req.UserID}, nil)

	chat.GroupMemberMutedNotification(ctx, req.OperationID, req.OpUserID, req.GroupID, req.UserID, req.MutedSeconds)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return ", pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""})
	return &pbGroup.MuteGroupMemberResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) CancelMuteGroupMember(ctx context.Context, req *pbGroup.CancelMuteGroupMemberReq) (*pbGroup.CancelMuteGroupMemberResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc args ", req.String())

	groupMemberInfo := db.GroupMember{GroupID: req.GroupID, UserID: req.UserID}
	groupMemberInfo.MuteEndTime = time.Unix(0, 0)
	err := imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		log.Error(req.OperationID, "UpdateGroupMemberInfo failed ", err.Error(), groupMemberInfo)
		return &pbGroup.CancelMuteGroupMemberResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupID, constant.SyncCancelMuteGroupMember, nil, []string{req.UserID}, nil)

	chat.GroupMemberCancelMutedNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return ", pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""})
	return &pbGroup.CancelMuteGroupMemberResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) MuteGroup(ctx context.Context, req *pbGroup.MuteGroupReq) (*pbGroup.MuteGroupResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc args ", req.String())

	opFlag, err := s.getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		errMsg := req.OperationID + " getGroupUserLevel failed " + req.GroupID + req.OpUserID + err.Error()
		log.Error(req.OperationID, errMsg)
		return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	}
	if opFlag == 0 {
		errMsg := req.OperationID + "opFlag == 0  " + req.GroupID + req.OpUserID
		log.Error(req.OperationID, errMsg)
		return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	}

	// mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.UserID)
	// if err != nil {
	//	errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID failed " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	// }
	// if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupOwner " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	// }
	// if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupAdmin " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	// }

	err = imdb.OperateGroupStatus(req.GroupID, constant.GroupStatusMuted)
	if err != nil {
		log.Error(req.OperationID, "OperateGroupStatus failed ", err.Error(), req.GroupID, constant.GroupStatusMuted)
		return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupID, constant.SyncMuteGroup, nil, nil, nil)

	chat.GroupMutedNotification(req.OperationID, req.OpUserID, req.GroupID)
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return ", pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""})
	return &pbGroup.MuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) CancelMuteGroup(ctx context.Context, req *pbGroup.CancelMuteGroupReq) (*pbGroup.CancelMuteGroupResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc args ", req.String())

	opFlag, err := s.getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		errMsg := req.OperationID + " getGroupUserLevel failed " + req.GroupID + req.OpUserID + err.Error()
		log.Error(req.OperationID, errMsg)
		return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	}
	if opFlag == 0 {
		errMsg := req.OperationID + "opFlag == 0  " + req.GroupID + req.OpUserID
		log.Error(req.OperationID, errMsg)
		return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	}

	// mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.)
	// if err != nil {
	//	errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID failed " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	// }
	// if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupOwner " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	// }
	// if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
	//	errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupAdmin " + req.GroupID + req.OpUserID + err.Error()
	//	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	// }
	log.Debug(req.OperationID, "UpdateGroupInfoDefaultZero ", req.GroupID, map[string]interface{}{"status": constant.GroupOk})
	err = imdb.UpdateGroupInfoDefaultZero(req.GroupID, map[string]interface{}{"status": constant.GroupOk})
	if err != nil {
		log.Error(req.OperationID, "UpdateGroupInfoDefaultZero failed ", err.Error(), req.GroupID)
		return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupID, constant.SyncCancelMuteGroup, nil, nil, nil)

	chat.GroupCancelMutedNotification(req.OperationID, req.OpUserID, req.GroupID)
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return ", pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""})
	return &pbGroup.CancelMuteGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) SetGroupMemberNickname(ctx context.Context, req *pbGroup.SetGroupMemberNicknameReq) (*pbGroup.SetGroupMemberNicknameResp, error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc args ", req.String())
	if req.OpUserID != req.UserID && !token_verify.IsManagerUserID(req.OpUserID) {
		errMsg := req.OperationID + " verify failed " + req.OpUserID + req.GroupID
		log.Error(req.OperationID, errMsg)
		return &pbGroup.SetGroupMemberNicknameResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}

	groupMemberInfo := db.GroupMember{}
	groupMemberInfo.UserID = req.UserID
	groupMemberInfo.GroupID = req.GroupID
	if req.Nickname == "" {
		userNickname, err := imdb.GetUserNameByUserID(groupMemberInfo.UserID)
		if err != nil {
			errMsg := req.OperationID + " GetUserNameByUserID failed " + err.Error()
			log.Error(req.OperationID, errMsg)
			return &pbGroup.SetGroupMemberNicknameResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
		groupMemberInfo.Nickname = userNickname
	} else {
		groupMemberInfo.Nickname = req.Nickname
	}
	err := imdb.UpdateGroupMemberInfo(groupMemberInfo)
	if err != nil {
		errMsg := req.OperationID + " UpdateGroupMemberInfo failed " + err.Error()
		log.Error(req.OperationID, errMsg)
		return &pbGroup.SetGroupMemberNicknameResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return ", pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""})

	go func() {
		_ = db.DB.SetUserNickNameByGroup(db.NickNameByGroupID, req.GroupID, req.UserID, groupMemberInfo.Nickname)
	}()
	return &pbGroup.SetGroupMemberNicknameResp{CommonResp: &pbGroup.CommonResp{ErrCode: 0, ErrMsg: ""}}, nil
}

func (s *groupServer) SetGroupMemberInfo(ctx context.Context, req *pbGroup.SetGroupMemberInfoReq) (resp *pbGroup.SetGroupMemberInfoResp, err error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbGroup.SetGroupMemberInfoResp{CommonResp: &pbGroup.CommonResp{}}
	groupMember := db.GroupMember{
		GroupID: req.GroupID,
		UserID:  req.UserID,
	}
	m := make(map[string]interface{})
	if req.RoleLevel != nil {
		m["role_level"] = req.RoleLevel.Value
	}
	if req.FaceURL != nil {
		m["user_group_face_url"] = req.FaceURL.Value
	}
	if req.Nickname != nil {
		m["nickname"] = req.Nickname.Value
	}
	if req.Ex != nil {
		m["ex"] = req.Ex.Value
	}

	// check the group member role level
	member, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetGroupMemberInfoByGroupIDAndUserID failed", err.Error())
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg + ":" + err.Error()
		return resp, nil
	}

	if member != nil && member.RoleLevel == 2 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "no allow to set group owner role level in this api")
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg + ":" + "no allow to set group owner role level in this api"
		return resp, nil
	}

	if req.RoleLevel != nil && req.RoleLevel.Value == 2 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "no allow to change group owner in this api")
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg + ":" + "no allow to change group owner in this api"
		return resp, nil
	}

	err = imdb.UpdateGroupMemberInfoByMap(groupMember, m)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetGroupMemberInfo failed", err.Error())
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg + ":" + err.Error()
		return resp, nil
	}

	//data synchronization
	syncGroupToLocal(req.OperationID, req.GroupID, constant.SyncGroupMemberInfo, nil, []string{req.UserID}, nil)

	if req.RoleLevel != nil {
		switch req.RoleLevel.Value {
		case constant.GroupOrdinaryUsers:
			// chat.GroupMemberRoleLevelChangeNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, constant.GroupMemberSetToOrdinaryUserNotification)
			chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
		case constant.GroupAdmin, constant.GroupOwner:
			// chat.GroupMemberRoleLevelChangeNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID, constant.GroupMemberSetToAdminNotification)
			chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
		}
	} else {
		chat.GroupMemberInfoSetNotification(req.OperationID, req.OpUserID, req.GroupID, req.UserID)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func syncGroupToLocal(operationID, groupID, msgType string, extUserIDList []string, opMemberIDList []string, opMemberList []*db.GroupMember) error {
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)
	newestGroupInfo, err := imdb.GetGroupInfoByGroupID(groupID)
	if err != nil {
		log.NewError(operationID, "GetGroupInfoByGroupID failed", groupID)
		return err
	}

	localGroupInfo := open_im_sdk.GroupInfo{}
	err = utils.CopyStructFields(&localGroupInfo, &newestGroupInfo)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}
	okUserIDList, err := imdb.GetGroupMemberIDListByGroupID(groupID)
	if err != nil {
		log.NewError(operationID, "GetGroupMemberIDListByGroupID failed")
		//return err
	}
	if extUserIDList != nil {
		okUserIDList = append(okUserIDList, extUserIDList...)
	}

	//set members num and owner id
	membernum, err := imdb.GetGroupMemberNumByGroupID(groupID)
	if err != nil {
		log.NewError(operationID, "GetGroupMemberNumByGroupID failed ", err.Error(), groupID)
		return err
	}
	localGroupInfo.MemberCount = uint32(membernum)
	owner, err := imdb.GetGroupOwnerInfoByGroupID(groupID)
	if err != nil {
		log.NewError(operationID, "GetGroupOwnerInfoByGroupID failed ", err.Error(), groupID)
		return err
	}
	localGroupInfo.OwnerUserID = owner.UserID

	//set member info list
	var memberInfoList []*open_im_sdk.GroupMemberFullInfo
	if opMemberIDList != nil {
		for _, s := range opMemberIDList {
			gm, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(groupID, s)
			if err == nil && gm.GroupID != "" {
				fullInfo := open_im_sdk.GroupMemberFullInfo{}
				err = utils.CopyStructFields(&fullInfo, &gm)
				if err == nil {
					memberInfoList = append(memberInfoList, &fullInfo)
				} else {
					log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
				}
			} else {
				log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			}
		}
	}
	if opMemberList != nil {
		for _, gm := range opMemberList {
			fullInfo := open_im_sdk.GroupMemberFullInfo{}
			err = utils.CopyStructFields(&fullInfo, &gm)
			if err == nil {
				memberInfoList = append(memberInfoList, &fullInfo)
			} else {
				log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			}
		}
	}

	syncGroupsReq := &local_database.SyncDataReq{
		OperationID:    operationID,
		MsgType:        msgType,
		MemberIDList:   okUserIDList,
		GroupInfo:      &localGroupInfo,
		MemberInfoList: memberInfoList,
	}
	localDataResp, err := localDataClient.SyncData(context.Background(), syncGroupsReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}
	if localDataResp.ErrCode != 0 {
		log.NewError(operationID, "SyncData rpc logic call failed ", localDataResp.String())
		return errors.New("SyncData rpc logic call failed")
	}
	return nil
}

func syncGroupRequestToLocal(operationID, groupID, userID, msgType string) error {
	//log.NewError(operationID, utils.GetSelfFuncName(), groupID, userID, msgType)
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)
	groupReqInfo, err := imdb.GetGroupRequestByGroupIDAndUserID(groupID, userID)
	if err != nil {
		log.NewError(operationID, "GetGroupRequestByGroupIDAndUserID failed")
		return err
	}

	localGroupReq := open_im_sdk.GroupRequest{}
	err = utils.CopyStructFields(&localGroupReq, &groupReqInfo)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}

	if groupReqInfo.GroupID != "" {
		groupInfo, err := imdb.GetGroupInfoByGroupID(groupReqInfo.GroupID)
		if err == nil {
			localGroupInfo := open_im_sdk.GroupInfo{}
			err = utils.CopyStructFields(&localGroupInfo, &groupInfo)
			if err != nil {
				log.NewError(operationID, "CopyStructFields failed")
				return err
			}
			localGroupReq.GroupInfo = &localGroupInfo
		}
	}

	if groupReqInfo.UserID != "" {
		userInfo, err := imdb.GetUserByUserID(groupReqInfo.UserID)
		if err == nil {
			localUserInfo := open_im_sdk.PublicUserInfo{}
			err = utils.CopyStructFields(&localUserInfo, &userInfo)
			if err != nil {
				log.NewError(operationID, "CopyStructFields failed")
				return err
			}
			localGroupReq.UserInfo = &localUserInfo
		}
	}

	managerIDList, err := imdb.GetGroupOwnerAdminIDListByGroupID(groupID)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}
	userIDList := []string{userID}

	syncGroupsReq := &local_database.SyncDataReq{
		OperationID:  operationID,
		MsgType:      msgType,
		MemberIDList: userIDList,
		GroupRequest: &localGroupReq,
	}
	localDataResp, err := localDataClient.SyncData(context.Background(), syncGroupsReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}

	syncAdminGroupsReq := &local_database.SyncDataReq{
		OperationID:  operationID,
		MsgType:      constant.SyncAdminGroupRequest,
		MemberIDList: managerIDList,
		GroupRequest: &localGroupReq,
	}
	localDataResp, err = localDataClient.SyncData(context.Background(), syncAdminGroupsReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}

	if localDataResp.ErrCode != 0 {
		log.NewError(operationID, "SyncData rpc logic call failed ", localDataResp.String())
		return errors.New("SyncData rpc logic call failed")
	}
	return nil
}

func syncDeleteGroupToLocal(operationID string, groupInfo *db.Group, userIDList []string) error {
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)

	localGroupInfo := open_im_sdk.GroupInfo{}
	err := utils.CopyStructFields(&localGroupInfo, &groupInfo)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}

	//set members num and owner id
	localGroupInfo.MemberCount = 0

	syncGroupsReq := &local_database.SyncDataReq{
		OperationID:  operationID,
		MsgType:      constant.SyncDeleteGroup,
		MemberIDList: userIDList,
		GroupInfo:    &localGroupInfo,
	}
	localDataResp, err := localDataClient.SyncData(context.Background(), syncGroupsReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}
	if localDataResp.ErrCode != 0 {
		log.NewError(operationID, "SyncData rpc logic call failed ", localDataResp.String())
		return errors.New("SyncData rpc logic call failed")
	}
	return nil
}

func (s *groupServer) CheckGroupUpdateVersionsFromLocal(ctx context.Context, req *pbGroup.CheckGroupUpdateVersionsFromLocalReq) (resp *pbGroup.GroupUpdatesVersionsRes, err error) {
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbGroup.GroupUpdatesVersionsRes{}

	//todo update group members through sdk sync
	//todo #1 get group version from group Member list
	groupVersionUpdate, err := imdb.GetGroupUpdatesVersionNumberGroupID(req.GroupID)
	//log.Error(req.OperationID, "GetGroupUpdatesVersionNumberGroupID ", groupVersionUpdate)
	//todo #2 compare version with user local version of this group
	if err == nil {
		if groupVersionUpdate.VersionNumber > req.GroupVersion {
			//todo #3 if user requested version number is ZERO then we have to send all contacts by 1000/page
			if req.GroupVersion == 0 || req.NeedNextPageFetch {
				//todo push all group members by paging 1000/page
				//log.Error(req.OperationID, "MemberSyncNotification called with Group Version 0")
				//*open_im_sdk.MemberSyncNotificationTips
				resp.MemberSyncNotificationTips = chat.MemberSyncNotification(req.OperationID, req.GroupID, req.OpUserID, req.NeedNextPageFetch, req.PageNumber, req.PageSize, nil, req.ResponseBackHTTP)
			} else {
				//todo #4 if not then sync only group members which have updated version number
				updatedUsersIDList, err := imdb.GetGroupMemberIDListByVersionGrater(req.GroupID, req.GroupVersion)
				if err == nil {
					//log.Error(req.OperationID, "GetGroupMemberIDListByVersionGrater ", updatedUsersIDList)
					resp.MemberSyncNotificationTips = chat.MemberSyncNotification(req.OperationID, req.GroupID, req.OpUserID, false, 0, 0, updatedUsersIDList, req.ResponseBackHTTP)
				} else {
					log.Error(req.OperationID, "GetGroupMemberIDListByVersionGrater Error : ", err.Error())
				}

			}

			//Sync function on sdk will be updated due to old sync process (remove which are not
		}
	}
	return resp, nil
}

func (s *groupServer) GetInterestGroupListByUserId(ctx context.Context, request *pbGroup.GetInterestGroupListByUserIdRequest) (*pbGroup.GetInterestGroupListByUserIdResponse, error) {
	userId := request.FromUserID
	resp := &pbGroup.GetInterestGroupListByUserIdResponse{}

	// cache
	cache, _ := db.DB.GetInterestGroupINfoListByUserId(userId)
	if cache == "" {
		err := json.Unmarshal([]byte(cache), &resp)
		if err == nil {
			return resp, nil
		}
	}

	// get user interest
	userInterest, _ := imdb.GetUserInterestList(userId)

	// get default interest type id list
	if len(userInterest) == 0 || (len(userInterest) == 1 && userInterest[0] == constant.InterestDefault) {
		userInterest = imdb.GetDefaultInterestTypeList()
	}

	if len(userInterest) == 0 || userInterest == nil {
		return resp, nil
	}

	groupIdList := imdb.GetInterestGroupByInterestIdList(userId, userInterest)
	if groupIdList == nil || len(groupIdList) == 0 {
		return resp, nil
	}

	groupList, err := imdb.GetGroupInfoByGroupIDList(groupIdList)
	if err != nil || groupList == nil {
		return resp, err
	}

	for _, group := range groupList {
		groupInfo := &open_im_sdk.GroupInfo{}
		err := utils.CopyStructFields(&groupInfo, group)
		if err != nil {
			return resp, err
		}

		groupMemberNum, _ := imdb.GetGroupMemberNumByGroupID(group.GroupID)
		groupInfo.Ex = utils.Int64ToString(groupMemberNum)

		resp.GroupList = append(resp.GroupList, groupInfo)
	}

	ca, _ := json.Marshal(resp)
	db.DB.SaveInterestGroupINfoListByUserId(userId, string(ca))

	return resp, nil
}

func (s *groupServer) SetVideoAudioStatus(_ context.Context, req *pbGroup.SetVideoAudioStatusReq) (*pbGroup.SetVideoAudioStatusResp, error) {
	resp := &pbGroup.SetVideoAudioStatusResp{CommonResp: &pbGroup.CommonResp{}}

	group := &db.Group{}
	group.GroupID = req.GroupID
	if req.StatusType == 1 {
		group.VideoStatus = int8(req.Status)
	} else if req.StatusType == 2 {
		group.AudioStatus = int8(req.Status)
	}

	if row := imdb.UpdateGroupStatus(group); row == 0 {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, nil
	}

	return resp, nil
}

func (s *groupServer) SetUserVideoAudioStatus(_ context.Context, req *pbGroup.SetUserVideoAudioStatusReq) (*pbGroup.SetUserVideoAudioStatusResp, error) {
	resp := &pbGroup.SetUserVideoAudioStatusResp{CommonResp: &pbGroup.CommonResp{}}

	member := &db.GroupMember{}
	member.GroupID = req.GroupID
	member.UserID = req.MemberID
	if req.StatusType == 1 {
		member.VideoStatus = int8(req.Status)
	} else if req.StatusType == 2 {
		member.AudioStatus = int8(req.Status)
	}

	if row := imdb.UpdateGroupMemberStatus(member); row == 0 {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, nil
	}

	return resp, nil
}

func (s *groupServer) GetUsersByGroup(_ context.Context, req *pbGroup.GetUsersByGroupReq) (*pbGroup.GetUsersByGroupResp, error) {
	resp := &pbGroup.GetUsersByGroupResp{CommonResp: &pbGroup.CommonResp{ErrCode: constant.OK.ErrCode, ErrMsg: ""}}

	resp.Pagination = &open_im_sdk.ResponsePagination{}

	if req.GroupID == "" {
		where := map[string]string{}
		members, count, err := imdb.GetUsersByWhere(where, nil, req.Pagination.ShowNumber, req.Pagination.PageNumber, "create_time:DESC")
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = err.Error()
			return resp, nil
		}
		resp.Users = make([]*pbGroup.UserIDAndName, len(members))
		for index, member := range members {
			uan := &pbGroup.UserIDAndName{
				UserName: member.Nickname,
				UserID:   member.UserID,
			}
			resp.Users[index] = uan
		}
		resp.UserNums = count
	} else {
		if req.GetType == 1 {
			// is member
			members, memberCounts, err := imdb.GetGroupMembersByGroupIdCMS(req.GroupID, "", req.Pagination.ShowNumber, req.Pagination.PageNumber)
			if err != nil {
				resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
				resp.CommonResp.ErrMsg = err.Error()
				return resp, nil
			}
			resp.Users = make([]*pbGroup.UserIDAndName, len(members))
			for index, member := range members {
				uan := &pbGroup.UserIDAndName{
					UserName: member.Nickname,
					UserID:   member.UserID,
				}
				resp.Users[index] = uan
			}
			resp.UserNums = memberCounts
		} else if req.GetType == 2 {
			// not member
			members, count, err := imdb.GetNoMembers(req.GroupID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
			if err != nil {
				return resp, err
			}
			resp.Users = make([]*pbGroup.UserIDAndName, len(members))
			for index, member := range members {
				uan := &pbGroup.UserIDAndName{
					UserName: member.Nickname,
					UserID:   member.UserID,
				}
				resp.Users[index] = uan
			}
			resp.UserNums = count
		}
	}

	return resp, nil
}
