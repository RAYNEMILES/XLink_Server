package local_database

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/http"
	"Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/local_database"
	"Open_IM/pkg/utils"
	"context"
	"github.com/panjf2000/ants"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
)

type localDataServer struct {
	rpcPort           int
	rpcRegisterName   string
	etcdSchema        string
	etcdAddr          []string
	dataDir           string
	syncGroupProducer *kafka.Producer
	pool              *ants.Pool
}

func NewLocalDataServer(port int) *localDataServer {
	localDataServer := localDataServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImLocalDataName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
		dataDir:         config.Config.LocalData.DataDir,
	}
	localDataServer.syncGroupProducer = kafka.NewKafkaProducer(config.Config.Kafka.MsgToSyncData.Addr, config.Config.Kafka.MsgToSyncData.Topic)
	localDataServer.pool, _ = ants.NewPool(10000)
	return &localDataServer
}

func (s *localDataServer) Run() {
	log.NewInfo("", "local_database rpc start")
	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}

	defer func() {
		if s.pool != nil {
			s.pool.Release()
		}
	}()
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("", "listen network success,", address, listener)
	defer listener.Close()

	server := grpc.NewServer()
	defer server.GracefulStop()

	//register server to etcd
	local_database.RegisterLocalDataBaseServer(server, s)

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
		log.NewError("", "Register Failed:", err.Error())
		return
	}
	log.NewInfo("", "RegisterEtcd ", s.etcdSchema, strings.Join(s.etcdAddr, ""), rpcRegisterIP, s.rpcPort, s.rpcRegisterName)
	err = server.Serve(listener)
	if err != nil {
		log.NewError("", "Server failed:", err.Error())
		return
	}
	log.NewInfo("", "group rpc success")

}

func (s *localDataServer) SyncData(c context.Context, req *local_database.SyncDataReq) (*local_database.CommonResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())

	respPb := &local_database.CommonResp{}
	operationID := req.OperationID

	msgType := req.MsgType
	memberIDList := req.MemberIDList
	if memberIDList == nil {
		respPb.ErrCode = constant.ErrArgs.ErrCode
		respPb.ErrMsg = "memberIDList is nil"
		log.NewError(operationID, utils.GetSelfFuncName(), respPb.ErrMsg)
		return respPb, http.WrapError(constant.ErrArgs)
	}
	log.NewError(req.OperationID, utils.GetSelfFuncName(), "msgType", msgType, "memberIDList", memberIDList)

	if req.GroupInfo == nil && req.GroupRequest == nil && req.FriendInfo == nil && req.Conversation == nil && req.FriendRequest == nil && req.BlackInfo == nil && req.UserInfo == nil {
		respPb.ErrCode = constant.ErrArgs.ErrCode
		respPb.ErrMsg = "args are nil"
		log.NewError(operationID, utils.GetSelfFuncName(), respPb.ErrMsg)
		return respPb, http.WrapError(constant.ErrArgs)
	}

	switch msgType {
	case constant.SyncCreateGroup, constant.SyncUpdateGroup, constant.SyncDeleteGroup,
		constant.SyncCancelMuteGroup, constant.SyncMuteGroup:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync create group:", req.String())

		localGroup := model_struct.LocalGroup{}
		err := utils.CopyStructFields(&localGroup, &req.GroupInfo)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localGroup
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})
		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncInvitedGroupMember, constant.SyncKickedGroupMember, constant.SyncMuteGroupMember, constant.SyncCancelMuteGroupMember, constant.SyncGroupMemberInfo:
		log.NewError(operationID, utils.GetSelfFuncName(), "sync create group:", req.String())

		var memberList []model_struct.LocalGroupMember
		for _, info := range req.MemberInfoList {
			localGroupMember := model_struct.LocalGroupMember{
				GroupID:        info.GroupID,
				UserID:         info.UserID,
				Nickname:       info.Nickname,
				FaceURL:        info.FaceURL,
				RoleLevel:      info.RoleLevel,
				JoinTime:       uint32(info.JoinTime),
				JoinSource:     info.JoinSource,
				MuteEndTime:    info.MuteEndTime,
				OperatorUserID: info.OperatorUserID,
				Ex:             info.Ex,
			}
			err := utils.CopyStructFields(&localGroupMember, &info)
			if err == nil {
				memberList = append(memberList, localGroupMember)
			}
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := memberList
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})
		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncConversation:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync conversation:", req.String())

		localConv := model_struct.LocalConversation{}
		err := utils.CopyStructFields(&localConv, &req.Conversation)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localConv
			p := operationID
			s.pool.Submit(func() {
				log.NewError(p, utils.GetSelfFuncName(), "memberID", m)
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})

		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncGroupRequest:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync group request:", req.String())

		localGroupReq := model_struct.LocalGroupRequest{}
		err := utils.CopyStructFields(&localGroupReq, &req.GroupRequest)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		localGroupReq.GroupID = req.GroupRequest.GroupInfo.GroupID
		localGroupReq.GroupName = req.GroupRequest.GroupInfo.GroupName
		localGroupReq.UserID = req.GroupRequest.UserInfo.UserID
		localGroupReq.UserFaceURL = req.GroupRequest.UserInfo.FaceURL

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localGroupReq
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})
		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncAdminGroupRequest:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync Admin Group Request:", req.String())

		localAdmGroupReq := model_struct.LocalAdminGroupRequest{}
		err := utils.CopyStructFields(&localAdmGroupReq, &req.GroupRequest)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		localAdmGroupReq.GroupID = req.GroupRequest.GroupInfo.GroupID
		localAdmGroupReq.GroupName = req.GroupRequest.GroupInfo.GroupName
		localAdmGroupReq.UserID = req.GroupRequest.UserInfo.UserID
		localAdmGroupReq.UserFaceURL = req.GroupRequest.UserInfo.FaceURL

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localAdmGroupReq
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})
		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncFriendRequest, constant.SyncSelfFriendRequest:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync friend request:", req.String())

		localFriendReq := model_struct.LocalFriendRequest{}
		err := utils.CopyStructFields(&localFriendReq, &req.FriendRequest)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localFriendReq
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})

		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncFriendInfo:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync friend request:", req.String())

		localFriend := model_struct.LocalFriend{}
		err := utils.CopyStructFields(&localFriend, &req.FriendInfo)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localFriend
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})
		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncAddBlackList, constant.SyncDeleteBlackList:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync black list:", req.String())

		localBlack := model_struct.LocalBlack{}
		err := utils.CopyStructFields(&localBlack, &req.BlackInfo)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localBlack
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})
		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncUserInfo:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync userinfo request:", req.String())

		localUser := model_struct.LocalUser{}
		err := utils.CopyStructFields(&localUser, &req.UserInfo)
		if err != nil {
			log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
			respPb.ErrCode = constant.ErrArgs.ErrCode
			respPb.ErrMsg = "CopyStructFields failed!"
			return respPb, err
		}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := localUser
			p := operationID
			s.pool.Submit(func() {
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})

		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	case constant.SyncWelcomeMessageFromChannel:
		log.NewInfo(operationID, utils.GetSelfFuncName(), "sync conversation:", req.String())

		//var localChatLogs []*model_struct.LocalChatLog
		//err := utils.CopyStructFields(&localChatLogs, &req.MsgList)
		//if err != nil {
		//	log.Error(operationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.String(), err.Error())
		//	respPb.ErrCode = constant.ErrArgs.ErrCode
		//	respPb.ErrMsg = "CopyStructFields failed!"
		//	return respPb, err
		//}

		for _, memberID := range memberIDList {
			m := memberID
			t := msgType
			l := req.MsgList
			p := operationID
			s.pool.Submit(func() {
				log.NewError(p, utils.GetSelfFuncName(), "memberID", m)
				pid, offset, err := s.syncGroupProducer.SendMessage(
					&local_database.SyncDataMsg{MsgType: t, UserID: m, MsgData: utils.StructToJsonString(l)},
					m,
					p)
				if err != nil {
					log.Error(operationID, "kafka send failed", "send data", "pid", pid, "offset", offset, "err", err.Error())
				}
			})

		}

		respPb.ErrCode = constant.OK.ErrCode
		respPb.ErrMsg = "kafka send success!"
		return respPb, nil
	default:
		log.NewError(operationID, utils.GetSelfFuncName(), "msgType error!")
		respPb.ErrCode = constant.ErrArgs.ErrCode
		respPb.ErrMsg = "msgType error!"
		return respPb, http.WrapError(constant.ErrArgs)
	}
}

//func (s *localDataServer) InitDataBase(c context.Context, req *local_database.InitDataBaseReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	userID := req.UserID
//	if userID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "userID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	//create database if not exist
//	_, err := db2.NewDataBase(userID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) SyncGroupMemberInfoToLocal(c context.Context, req *local_database.SyncGroupMemberInfoReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	userID := req.UserID
//	groupID := req.GroupID
//	if userID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "userID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//	if groupID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "groupID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	//get data from mysql
//	memberInfo, err2 := imdb.GetGroupMemberInfoByGroupIDAndUserID(groupID, userID)
//	if err2 != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "get data from mysql is failed:", err2.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	//create database if not exist
//	db, err := db2.NewDataBase(userID, s.dataDir)
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroupMemberInfoByGroupIDUserID0 ", groupID, userID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	//update local database
//	localMemberInfo := &model_struct.LocalGroupMember{
//		GroupID:        groupID,
//		UserID:         userID,
//		Nickname:       memberInfo.Nickname,
//		FaceURL:        memberInfo.FaceURL,
//		RoleLevel:      memberInfo.RoleLevel,
//		JoinTime:       uint32(memberInfo.JoinTime.Unix()),
//		JoinSource:     memberInfo.JoinSource,
//		MuteEndTime:    uint32(memberInfo.MuteEndTime.Unix()),
//		OperatorUserID: memberInfo.OperatorUserID,
//		Ex:             memberInfo.Ex,
//	}
//
//	currentMemberInfo, _ := db.GetGroupMemberInfoByGroupIDUserID(groupID, userID)
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroupMemberInfoByGroupIDUserID1 ", currentMemberInfo.GroupID, currentMemberInfo.UserID)
//	//currentMemberInfo.Nickname = memberInfo.Nickname
//	//currentMemberInfo.FaceURL = memberInfo.FaceURL
//	//currentMemberInfo.RoleLevel = memberInfo.RoleLevel
//	//currentMemberInfo.JoinTime = uint32(memberInfo.JoinTime.Unix())
//	//currentMemberInfo.JoinSource = memberInfo.JoinSource
//	//currentMemberInfo.MuteEndTime = uint32(memberInfo.MuteEndTime.Unix())
//	//currentMemberInfo.OperatorUserID = memberInfo.OperatorUserID
//	//currentMemberInfo.Ex = memberInfo.Ex
//	err = db.UpdateGroupMember(localMemberInfo)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update Local Group Member is failed:", err.Error(), localMemberInfo.MuteEndTime, currentMemberInfo.MuteEndTime)
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetGroupMemberInfoByGroupIDUserID2 ", currentMemberInfo.GroupID, currentMemberInfo.UserID, localMemberInfo.MuteEndTime)
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncGroupMemberInfoToLocal is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) SyncGroupMemerListToLocal(c context.Context, req *local_database.SyncGroupMemberListReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	userID := req.UserID
//	memberListStr := req.MemberList
//	if userID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "userID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//	if memberListStr == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "memberListStr is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	//create database if not exist
//	db, err := db2.NewDataBase(userID, s.dataDir)
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncGroupMemerListToLocal ", userID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	//update local database
//	memberList := []model_struct.LocalGroupMember{}
//	err = utils.JsonStringToStruct(memberListStr, &memberList)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "memberListStr JsonStringToStruct failed:", err.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	for _, member := range memberList {
//		groupID := member.GroupID
//		userID := member.UserID
//		_, err := db.GetGroupMemberInfoByGroupIDUserID(groupID, userID)
//		if err == nil {
//			//update member
//			log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "update member")
//			localMemberInfo := &model_struct.LocalGroupMember{
//				GroupID:        groupID,
//				UserID:         userID,
//				Nickname:       member.Nickname,
//				FaceURL:        member.FaceURL,
//				RoleLevel:      member.RoleLevel,
//				JoinTime:       member.JoinTime,
//				JoinSource:     member.JoinSource,
//				MuteEndTime:    member.MuteEndTime,
//				OperatorUserID: member.OperatorUserID,
//				Ex:             member.Ex,
//			}
//			err = db.UpdateGroupMember(localMemberInfo)
//			if err != nil {
//				log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateGroupMember failed:", err.Error())
//				respPb.ErrCode = constant.ErrDB.ErrCode
//				respPb.ErrMsg = constant.ErrDB.ErrMsg
//				return respPb, http.WrapError(constant.ErrDB)
//			}
//		} else {
//			//create member
//			log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "create member")
//			localMemberInfo := &model_struct.LocalGroupMember{
//				GroupID:        groupID,
//				UserID:         userID,
//				Nickname:       member.Nickname,
//				FaceURL:        member.FaceURL,
//				RoleLevel:      member.RoleLevel,
//				JoinTime:       member.JoinTime,
//				JoinSource:     member.JoinSource,
//				MuteEndTime:    member.MuteEndTime,
//				OperatorUserID: member.OperatorUserID,
//				Ex:             member.Ex,
//			}
//			err = db.InsertGroupMember(localMemberInfo)
//			if err != nil {
//				log.NewError(req.OperationID, utils.GetSelfFuncName(), "InsertGroupMember failed:", err.Error())
//				respPb.ErrCode = constant.ErrDB.ErrCode
//				respPb.ErrMsg = constant.ErrDB.ErrMsg
//				return respPb, http.WrapError(constant.ErrDB)
//			}
//		}
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncGroupMemerListToLocal is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) SyncUserInfoToLocal(c context.Context, req *local_database.SyncUserInfoReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	userID := req.UserID
//	if userID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "userID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	//create database if not exist
//	db, err := db2.NewDataBase(userID, s.dataDir)
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncUserInfoToLocal ", userID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//		return respPb, err
//	}
//
//	//get userinfo from mysql
//	userInfo, err := imdb.GetUserByUserID(userID)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByUserID is failed:", err.Error())
//		return respPb, err
//	}
//
//	//update local database
//	localUser := &model_struct.LocalUser{}
//	err = utils.CopyStructFields(localUser, &userInfo)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed:", err.Error())
//		return respPb, err
//	}
//
//	//set birth
//	localUser.Birth = utils.GetTimeStringFromTime(userInfo.Birth)
//
//	_, err = db.GetLoginUserWithParams(&model_struct.LocalUser{UserID: userID})
//	if err == nil {
//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "UpdateLoginUser")
//		err = db.UpdateLoginUser(localUser)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateLoginUser failed:", err.Error())
//			return respPb, err
//		}
//	} else {
//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "InsertLoginUser")
//		err = db.InsertLoginUser(localUser)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "InsertLoginUser failed:", err.Error())
//			return respPb, err
//		}
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncUserInfoToLocal is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) SyncAddFriendToLocal(c context.Context, req *local_database.SyncAddFriendReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	userID := req.UserID
//	friendID := req.FriendID
//	operatorID := req.OperatorID
//
//	if userID == "" || friendID == "" || operatorID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "userID or friendID or operatorID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	//create database if not exist
//	db, err := db2.NewDataBase(userID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase for user is failed:", userID, err.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//	db2, err := db2.NewDataBase(friendID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase for friend is failed:", friendID, err.Error())
//		return respPb, http.WrapError(constant.ErrDB)
//	}
//
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncAddFriendToLocal ", userID, friendID, operatorID, s.dataDir)
//
//	_, err = db.GetFriendInfoByFriendUserID(friendID)
//	if err != nil {
//		//create Friend
//		friendInfo, err := imdb.GetUserByUserID(friendID)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByUserID failed:", friendID, err.Error())
//			respPb.ErrCode = constant.ErrDB.ErrCode
//			respPb.ErrMsg = constant.ErrDB.ErrMsg
//			return respPb, err
//		}
//
//		localFriendToUser := &model_struct.LocalFriend{
//			OwnerUserID:    userID,
//			FriendUserID:   friendID,
//			OperatorUserID: operatorID,
//			CreateTime:     uint32(friendInfo.CreateTime),
//			Nickname:       friendInfo.Nickname,
//			FaceURL:        friendInfo.FaceURL,
//			Gender:         friendInfo.Gender,
//			PhoneNumber:    friendInfo.PhoneNumber,
//			Birth:          utils.GetTimeStringFromTime(friendInfo.Birth),
//			Email:          friendInfo.Email,
//			Ex:             friendInfo.Ex,
//		}
//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "user add Friend")
//		err = db.InsertFriend(localFriendToUser)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "create Friend to User failed:", err.Error())
//			respPb.ErrCode = constant.ErrDB.ErrCode
//			respPb.ErrMsg = constant.ErrDB.ErrMsg
//			return respPb, http.WrapError(constant.ErrDB)
//		}
//	}
//
//	_, err = db2.GetFriendInfoByFriendUserID(userID)
//	if err != nil {
//		//create Friend
//		friendInfo, err := imdb.GetUserByUserID(userID)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByUserID failed:", userID, err.Error())
//			respPb.ErrCode = constant.ErrDB.ErrCode
//			respPb.ErrMsg = constant.ErrDB.ErrMsg
//			return respPb, err
//		}
//
//		localFriendToFriend := &model_struct.LocalFriend{
//			OwnerUserID:    friendID,
//			FriendUserID:   userID,
//			OperatorUserID: operatorID,
//			CreateTime:     uint32(friendInfo.CreateTime),
//			Nickname:       friendInfo.Nickname,
//			FaceURL:        friendInfo.FaceURL,
//			Gender:         friendInfo.Gender,
//			PhoneNumber:    friendInfo.PhoneNumber,
//			Birth:          utils.GetTimeStringFromTime(friendInfo.Birth),
//			Email:          friendInfo.Email,
//			Ex:             friendInfo.Ex,
//		}
//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "friend add user")
//		err = db2.InsertFriend(localFriendToFriend)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "create User to Friend failed:", err.Error())
//			respPb.ErrCode = constant.ErrDB.ErrCode
//			respPb.ErrMsg = constant.ErrDB.ErrMsg
//			return respPb, http.WrapError(constant.ErrDB)
//		}
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "SyncAddFriendToLocal is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) KickGroupMemerForAllToLocal(c context.Context, req *local_database.KickGroupMemberForAllReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	groupID := req.GroupID
//	kickedID := req.KickedID
//
//	if groupID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "groupID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	memberList, err := imdb.GetGroupMemberIDListByGroupID(groupID)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "memberList is nil")
//		return respPb, errors.New("memberList is nil")
//	}
//
//	kickedDatabase, err := db2.NewDataBase(kickedID, sdk_struct.SvrConf.DataDir)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "kickedDatabase is failed")
//		return respPb, errors.New("kickedDatabase is failed")
//	}
//	err = kickedDatabase.DeleteGroup(groupID)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete group for kicked user is failed")
//		return respPb, errors.New("delete group for kicked user is failed")
//	}
//
//	for _, memberID := range memberList {
//		//create database if not exist
//		db, err := db2.NewDataBase(memberID, s.dataDir)
//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), groupID, memberID, s.dataDir)
//		if err != nil {
//			respPb.ErrCode = constant.ErrDB.ErrCode
//			respPb.ErrMsg = constant.ErrDB.ErrMsg
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//			return respPb, errors.New("NewDataBase is failed")
//		}
//
//		localMember, _ := db.GetGroupMemberInfoByGroupIDUserID(groupID, kickedID)
//		if localMember != nil {
//			db.DeleteGroupMember(groupID, kickedID)
//		}
//
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "DeleteGroupMemersForAllToLocal is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) DeleteGroupMemersForAllToLocal(c context.Context, req *local_database.DeleteGroupMembersForAllReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	groupID := req.GroupID
//
//	if groupID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "groupID is nil")
//		return respPb, http.WrapError(constant.ErrArgs)
//	}
//
//	memberList, err := imdb.GetGroupMemberIDListByGroupID(groupID)
//	if err != nil {
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "memberList is nil")
//		return respPb, errors.New("memberList is nil")
//	}
//
//	for _, memberID := range memberList {
//		//create database if not exist
//		db, err := db2.NewDataBase(memberID, s.dataDir)
//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), groupID, memberID, s.dataDir)
//		if err != nil {
//			respPb.ErrCode = constant.ErrDB.ErrCode
//			respPb.ErrMsg = constant.ErrDB.ErrMsg
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//			return respPb, errors.New("NewDataBase is failed")
//		}
//		db.DeleteGroupAllMembers(groupID)
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "DeleteGroupMemersForAllToLocal is success")
//	return respPb, nil
//
//}
//
//func (s *localDataServer) SyncConversationToLocal(c context.Context, req *local_database.SyncConversationReq) (*local_database.CommonResp, error) {
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "args ", req.String())
//
//	respPb := &local_database.CommonResp{}
//	ownerID := req.OwnerID
//	conversationID := req.ConversationID
//
//	if ownerID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "ownerID is nil")
//		return respPb, errors.New("ownerID is nil")
//	}
//
//	if conversationID == "" {
//		respPb.ErrCode = constant.ErrArgs.ErrCode
//		respPb.ErrMsg = constant.ErrArgs.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "conversationID is nil")
//		return respPb, errors.New("conversationID is nil")
//	}
//
//	conversationSvr, err := imdb.GetConversation(ownerID, conversationID)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//		return respPb, errors.New("conversationSvr is nil")
//	}
//
//	localConversation := model_struct.LocalConversation{}
//	utils.CopyStructFields(&localConversation, &conversationSvr)
//
//	//create database if not exist
//	db, err := db2.NewDataBase(ownerID, s.dataDir)
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), ownerID, conversationID, s.dataDir)
//	if err != nil {
//		respPb.ErrCode = constant.ErrDB.ErrCode
//		respPb.ErrMsg = constant.ErrDB.ErrMsg
//		log.NewError(req.OperationID, utils.GetSelfFuncName(), "NewDataBase is failed:", err.Error())
//		return respPb, errors.New("NewDataBase is failed")
//	}
//
//	oldConversation, err := db.GetConversation(conversationID)
//	if oldConversation != nil && oldConversation.ConversationID != "" {
//		err = db.UpdateConversation(&localConversation)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateConversationForSync is failed", err.Error(), localConversation)
//			return respPb, errors.New("UpdateConversationForSync is failed")
//		}
//	} else {
//		err = db.InsertConversation(&localConversation)
//		if err != nil {
//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "InsertConversation is failed", err.Error(), localConversation)
//			return respPb, errors.New("InsertConversation is failed")
//		}
//	}
//
//	respPb.ErrCode = constant.OK.ErrCode
//	respPb.ErrMsg = constant.OK.ErrMsg
//	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "DeleteGroupMemersForAllToLocal is success")
//	return respPb, nil
//
//}
