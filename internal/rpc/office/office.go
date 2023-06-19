package office

import (
	"Open_IM/internal/push/sdk/tpns-server-sdk-go/go/req"
	"Open_IM/internal/rpc/admin_cms"
	"Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbCache "Open_IM/pkg/proto/cache"
	pbOffice "Open_IM/pkg/proto/office"
	pbCommon "Open_IM/pkg/proto/sdk_ws"
	sdk_ws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"google.golang.org/grpc"
)

type officeServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
	ch              chan tagSendStruct
}

func NewOfficeServer(port int) *officeServer {
	ch := make(chan tagSendStruct, 100000)
	return &officeServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImOfficeName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
		ch:              ch,
	}
}

func (s *officeServer) Run() {
	log.NewPrivateLog(constant.OpenImOfficeLog)
	log.NewInfo("0", "officeServer rpc start ")
	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)
	//listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("0", "listen network success, ", address, listener)
	defer listener.Close()
	//grpc server
	recvSize := 1024 * 1024 * 30
	sendSize := 1024 * 1024 * 30
	var options = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(recvSize),
		grpc.MaxSendMsgSize(sendSize),
	}
	srv := grpc.NewServer(options...)
	defer srv.GracefulStop()
	//Service registers with etcd
	pbOffice.RegisterOfficeServiceServer(srv, s)
	rpcRegisterIP := ""
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error())
		return
	}
	go s.sendTagMsgRoutine()
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "message cms rpc success")
}

type tagSendStruct struct {
	operationID      string
	user             *db.User
	userID           string
	content          string
	senderPlatformID int32
}

func (s *officeServer) sendTagMsgRoutine() {
	log.NewInfo("", utils.GetSelfFuncName(), "start")
	for {
		select {
		case v := <-s.ch:
			msg.TagSendMessage(v.operationID, v.user, v.userID, v.content, v.senderPlatformID)
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (s *officeServer) GetUserTags(_ context.Context, req *pbOffice.GetUserTagsReq) (resp *pbOffice.GetUserTagsResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req ", req.String())
	resp = &pbOffice.GetUserTagsResp{
		CommonResp: &pbOffice.CommonResp{},
		Tags:       []*pbOffice.Tag{},
	}
	tags, err := db.DB.GetUserTags(req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserTags failed", err.Error())
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "tags: ", tags)
	for _, v := range tags {
		tag := &pbOffice.Tag{
			TagID:   v.TagID,
			TagName: v.TagName,
		}
		for _, userID := range v.UserList {
			UserName, err := im_mysql_model.GetUserNameByUserID(userID)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID failed", err.Error())
				continue
			}
			tag.UserList = append(tag.UserList, &pbOffice.TagUser{
				UserID:   userID,
				UserName: UserName,
			})
		}
		resp.Tags = append(resp.Tags, tag)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp ", resp.String())
	return resp, nil
}

func (s *officeServer) CreateTag(_ context.Context, req *pbOffice.CreateTagReq) (resp *pbOffice.CreateTagResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "CreateTag req", req.String())
	userIDList := utils.RemoveRepeatedStringInList(req.UserIDList)
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "userIDList: ", userIDList)
	resp = &pbOffice.CreateTagResp{CommonResp: &pbOffice.CommonResp{}}
	if err := db.DB.CreateTag(req.UserID, req.TagName, userIDList); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserTags failed", err.Error())
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp", resp.String())
	return resp, nil
}

func (s *officeServer) DeleteTag(_ context.Context, req *pbOffice.DeleteTagReq) (resp *pbOffice.DeleteTagResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.DeleteTagResp{CommonResp: &pbOffice.CommonResp{}}
	if err := db.DB.DeleteTag(req.UserID, req.TagID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteTag failed", err.Error())
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) SetTag(_ context.Context, req *pbOffice.SetTagReq) (resp *pbOffice.SetTagResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.SetTagResp{CommonResp: &pbOffice.CommonResp{}}
	increaseUserIDList := utils.RemoveRepeatedStringInList(req.IncreaseUserIDList)
	reduceUserIDList := utils.RemoveRepeatedStringInList(req.ReduceUserIDList)
	if err := db.DB.SetTag(req.UserID, req.TagID, req.NewName, increaseUserIDList, reduceUserIDList); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetTag failed", increaseUserIDList, reduceUserIDList, err.Error())
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) SendMsg2Tag(_ context.Context, req *pbOffice.SendMsg2TagReq) (resp *pbOffice.SendMsg2TagResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.SendMsg2TagResp{CommonResp: &pbOffice.CommonResp{}}
	var tagUserIDList []string
	for _, tagID := range req.TagList {
		userIDList, err := db.DB.GetUserIDListByTagID(req.SendID, tagID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserIDListByTagID failed", err.Error())
			continue
		}
		tagUserIDList = append(tagUserIDList, userIDList...)
	}
	var groupUserIDList []string
	for _, groupID := range req.GroupList {
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrInternal.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, nil
		}

		cacheClient := pbCache.NewCacheClient(etcdConn)
		req := pbCache.GetGroupMemberIDListFromCacheReq{
			OperationID: req.OperationID,
			GroupID:     groupID,
		}
		getGroupMemberIDListFromCacheResp, err := cacheClient.GetGroupMemberIDListFromCache(context.Background(), &req)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupMemberIDListFromCache rpc call failed ", err.Error(), req.String())
			resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
			resp.CommonResp.ErrMsg = err.Error()
			return resp, nil
		}
		if getGroupMemberIDListFromCacheResp.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, "GetGroupMemberIDListFromCache rpc logic call failed ", getGroupMemberIDListFromCacheResp.CommonResp.ErrCode)
			resp.CommonResp.ErrCode = getGroupMemberIDListFromCacheResp.CommonResp.ErrCode
			resp.CommonResp.ErrMsg = getGroupMemberIDListFromCacheResp.CommonResp.ErrMsg
			return resp, nil
		}
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), getGroupMemberIDListFromCacheResp.UserIDList)
		groupUserIDList = append(groupUserIDList, getGroupMemberIDListFromCacheResp.UserIDList...)
	}
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), groupUserIDList, req.GroupList)
	var userIDList []string
	userIDList = append(userIDList, tagUserIDList...)
	userIDList = append(userIDList, groupUserIDList...)
	userIDList = append(userIDList, req.UserList...)
	userIDList = utils.RemoveRepeatedStringInList(userIDList)
	for i, userID := range userIDList {
		if userID == req.SendID || userID == "" {
			userIDList = append(userIDList[:i], userIDList[i+1:]...)
		}
	}
	if unsafe.Sizeof(userIDList) > 1024*1024 {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "size", unsafe.Sizeof(userIDList))
		resp.CommonResp.ErrMsg = constant.ErrSendLimit.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrSendLimit.ErrCode
		return
	}
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "total userIDList result: ", userIDList)
	user, err := imdb.GetUserByUserID(req.SendID)
	if err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.SendID)
		resp.CommonResp.ErrMsg = err.Error()
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	var successUserIDList []string
	for _, userID := range userIDList {
		t := tagSendStruct{
			operationID:      req.OperationID,
			user:             user,
			userID:           userID,
			content:          req.Content,
			senderPlatformID: req.SenderPlatformID,
		}
		select {
		case s.ch <- t:
			log.NewDebug(t.operationID, utils.GetSelfFuncName(), "msg: ", t, "send success")
			successUserIDList = append(successUserIDList, userID)
		// if channel is full, return grpc req
		case <-time.After(1 * time.Second):
			log.NewError(t.operationID, utils.GetSelfFuncName(), s.ch, "channel is full")
			resp.CommonResp.ErrCode = constant.ErrSendLimit.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrSendLimit.ErrMsg
			return resp, nil
		}
	}

	var tagSendLogs db.TagSendLog
	var wg sync.WaitGroup
	wg.Add(len(successUserIDList))
	var lock sync.Mutex
	for _, userID := range successUserIDList {
		go func(userID string) {
			defer wg.Done()
			userName, err := im_mysql_model.GetUserNameByUserID(userID)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID failed", err.Error(), userID)
				return
			}
			lock.Lock()
			tagSendLogs.UserList = append(tagSendLogs.UserList, db.TagUser{
				UserID:   userID,
				UserName: userName,
			})
			lock.Unlock()
		}(userID)
	}
	wg.Wait()
	tagSendLogs.SendID = req.SendID
	tagSendLogs.Content = req.Content
	tagSendLogs.SenderPlatformID = req.SenderPlatformID
	tagSendLogs.SendTime = time.Now().Unix()
	if err := db.DB.SaveTagSendLog(&tagSendLogs); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SaveTagSendLog failed", tagSendLogs, err.Error())
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) GetTagSendLogs(_ context.Context, req *pbOffice.GetTagSendLogsReq) (resp *pbOffice.GetTagSendLogsResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.GetTagSendLogsResp{
		CommonResp: &pbOffice.CommonResp{},
		Pagination: &pbCommon.ResponsePagination{
			CurrentPage: req.Pagination.PageNumber,
			ShowNumber:  req.Pagination.ShowNumber,
		},
		TagSendLogs: []*pbOffice.TagSendLog{},
	}
	tagSendLogs, err := db.DB.GetTagSendLogs(req.UserID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetTagSendLogs", err.Error())
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	if err := utils.CopyStructFields(&resp.TagSendLogs, tagSendLogs); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) GetUserTagByID(_ context.Context, req *pbOffice.GetUserTagByIDReq) (resp *pbOffice.GetUserTagByIDResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.GetUserTagByIDResp{
		CommonResp: &pbOffice.CommonResp{},
		Tag:        &pbOffice.Tag{},
	}
	tag, err := db.DB.GetTagByID(req.UserID, req.TagID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetTagByID failed", err.Error())
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, nil
	}
	for _, userID := range tag.UserList {
		userName, err := im_mysql_model.GetUserNameByUserID(userID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID failed", err.Error())
			continue
		}
		resp.Tag.UserList = append(resp.Tag.UserList, &pbOffice.TagUser{
			UserID:   userID,
			UserName: userName,
		})
	}
	resp.Tag.TagID = tag.TagID
	resp.Tag.TagName = tag.TagName
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) CreateOneWorkMoment(_ context.Context, req *pbOffice.CreateOneWorkMomentReq) (resp *pbOffice.CreateOneWorkMomentResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.CreateOneWorkMomentResp{CommonResp: &pbOffice.CommonResp{}}
	workMoment := db.WorkMoment{
		Comments:           []*db.Comment{},
		LikeUserList:       []*db.WorkMomentUser{},
		PermissionUserList: []*db.WorkMomentUser{},
	}
	createUser, err := imdb.GetUserByUserID(req.WorkMoment.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByUserID", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(&workMoment, req.WorkMoment); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	workMoment.UserName = createUser.Nickname
	workMoment.FaceURL = createUser.FaceURL
	workMoment.PermissionUserIDList = s.getPermissionUserIDList(req.OperationID, req.WorkMoment.PermissionGroupList, req.WorkMoment.PermissionUserList)
	workMoment.PermissionUserList = []*db.WorkMomentUser{}
	for _, userID := range workMoment.PermissionUserIDList {
		userName, err := imdb.GetUserNameByUserID(userID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID failed", err.Error())
			continue
		}
		workMoment.PermissionUserList = append(workMoment.PermissionUserList, &db.WorkMomentUser{
			UserID:   userID,
			UserName: userName,
		})
	}
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "workMoment to create", workMoment)
	err = db.DB.CreateOneWorkMoment(&workMoment)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateOneWorkMoment", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}

	// send notification to at users
	for _, atUser := range req.WorkMoment.AtUserList {
		workMomentNotificationMsg := &pbOffice.WorkMomentNotificationMsg{
			NotificationMsgType: constant.WorkMomentAtUserNotification,
			WorkMomentID:        workMoment.WorkMomentID,
			WorkMomentContent:   workMoment.Content,
			UserID:              workMoment.UserID,
			FaceURL:             createUser.FaceURL,
			UserName:            createUser.Nickname,
			CreateTime:          workMoment.CreateTime,
		}
		msg.WorkMomentSendNotification(req.OperationID, atUser.UserID, workMomentNotificationMsg)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) DeleteComment(_ context.Context, req *pbOffice.DeleteCommentReq) (resp *pbOffice.DeleteCommentResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.DeleteCommentResp{CommonResp: &pbOffice.CommonResp{}}
	err = db.DB.DeleteComment(req.WorkMomentID, req.ContentID, req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetWorkMomentByID failed", err.Error())
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

// count and distinct permission users
func (s *officeServer) getPermissionUserIDList(operationID string, groupList []*pbOffice.PermissionGroup, userList []*pbOffice.WorkMomentUser) []string {
	var permissionUserIDList []string
	for _, group := range groupList {
		groupMemberIDList, err := imdb.GetGroupMemberIDListByGroupID(group.GroupID)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), "GetGroupMemberIDListByGroupID failed", group, err.Error())
			continue
		}
		log.NewDebug(operationID, utils.GetSelfFuncName(), "groupMemberIDList: ", groupMemberIDList)
		permissionUserIDList = append(permissionUserIDList, groupMemberIDList...)
	}
	var userIDList []string
	for _, user := range userList {
		userIDList = append(userIDList, user.UserID)
	}
	permissionUserIDList = append(permissionUserIDList, userIDList...)
	permissionUserIDList = utils.RemoveRepeatedStringInList(permissionUserIDList)
	return permissionUserIDList
}

func (s *officeServer) DeleteOneWorkMoment(_ context.Context, req *pbOffice.DeleteOneWorkMomentReq) (resp *pbOffice.DeleteOneWorkMomentResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.DeleteOneWorkMomentResp{CommonResp: &pbOffice.CommonResp{}}
	workMoment, err := db.DB.GetWorkMomentByID(req.WorkMomentID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetWorkMomentByID failed", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "workMoment", workMoment)
	if workMoment.UserID != req.UserID {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "workMoment.UserID != req.WorkMomentID, delete failed", workMoment, req.WorkMomentID)
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}
		return resp, nil
	}
	err = db.DB.DeleteOneWorkMoment(req.WorkMomentID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteOneWorkMoment", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func isUserCanSeeWorkMoment(userID string, workMoment db.WorkMoment) bool {
	if userID != workMoment.UserID {
		switch workMoment.Permission {
		case constant.WorkMomentPublic:
			return true
		case constant.WorkMomentPrivate:
			return false
		case constant.WorkMomentPermissionCanSee:
			return utils.IsContain(userID, workMoment.PermissionUserIDList)
		case constant.WorkMomentPermissionCantSee:
			return !utils.IsContain(userID, workMoment.PermissionUserIDList)
		}
		return false
	}
	return true
}

func (s *officeServer) LikeOneWorkMoment(_ context.Context, req *pbOffice.LikeOneWorkMomentReq) (resp *pbOffice.LikeOneWorkMomentResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.LikeOneWorkMomentResp{CommonResp: &pbOffice.CommonResp{}}
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID failed", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	workMoment, like, err := db.DB.LikeOneWorkMoment(req.UserID, user.Nickname, req.WorkMomentID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "LikeOneWorkMoment failed ", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	workMomentNotificationMsg := &pbOffice.WorkMomentNotificationMsg{
		NotificationMsgType: constant.WorkMomentLikeNotification,
		WorkMomentID:        workMoment.WorkMomentID,
		WorkMomentContent:   workMoment.Content,
		UserID:              user.UserID,
		FaceURL:             user.FaceURL,
		UserName:            user.Nickname,
		CreateTime:          int32(time.Now().Unix()),
	}
	// send notification
	if like && workMoment.UserID != req.UserID {
		msg.WorkMomentSendNotification(req.OperationID, workMoment.UserID, workMomentNotificationMsg)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) CommentOneWorkMoment(_ context.Context, req *pbOffice.CommentOneWorkMomentReq) (resp *pbOffice.CommentOneWorkMomentResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.CommentOneWorkMomentResp{CommonResp: &pbOffice.CommonResp{}}
	commentUser, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID commentUserName failed", req.UserID, err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	var replyUserName string
	if req.ReplyUserID != "" {
		replyUserName, err = imdb.GetUserNameByUserID(req.ReplyUserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserNameByUserID get replyUserName failed", req.ReplyUserID, err.Error())
			resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
			return resp, nil
		}
	}
	comment := &db.Comment{
		UserID:        req.UserID,
		UserName:      commentUser.Nickname,
		ReplyUserID:   req.ReplyUserID,
		ReplyUserName: replyUserName,
		Content:       req.Content,
		CreateTime:    int32(time.Now().Unix()),
	}
	workMoment, err := db.DB.CommentOneWorkMoment(comment, req.WorkMomentID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CommentOneWorkMoment failed", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	workMomentNotificationMsg := &pbOffice.WorkMomentNotificationMsg{
		NotificationMsgType: constant.WorkMomentCommentNotification,
		WorkMomentID:        workMoment.WorkMomentID,
		WorkMomentContent:   workMoment.Content,
		UserID:              commentUser.UserID,
		FaceURL:             commentUser.FaceURL,
		UserName:            commentUser.Nickname,
		ReplyUserID:         comment.ReplyUserID,
		ReplyUserName:       comment.ReplyUserName,
		ContentID:           comment.ContentID,
		Content:             comment.Content,
		CreateTime:          comment.CreateTime,
	}
	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "msg: ", *workMomentNotificationMsg)
	if req.UserID != workMoment.UserID {
		msg.WorkMomentSendNotification(req.OperationID, workMoment.UserID, workMomentNotificationMsg)
	}
	if req.ReplyUserID != "" && req.ReplyUserID != workMoment.UserID && req.ReplyUserID != req.UserID {
		msg.WorkMomentSendNotification(req.OperationID, req.ReplyUserID, workMomentNotificationMsg)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) GetWorkMomentByID(_ context.Context, req *pbOffice.GetWorkMomentByIDReq) (resp *pbOffice.GetWorkMomentByIDResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.GetWorkMomentByIDResp{
		WorkMoment: &pbOffice.WorkMoment{},
		CommonResp: &pbOffice.CommonResp{},
	}
	workMoment, err := db.DB.GetWorkMomentByID(req.WorkMomentID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetWorkMomentByID failed", err.Error())
		return resp, nil
	}
	canSee := isUserCanSeeWorkMoment(req.OpUserID, *workMoment)
	log.Debug(req.OperationID, utils.GetSelfFuncName(), canSee, req.OpUserID, *workMoment)
	if !canSee {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "workMoments not access to user", canSee, workMoment, req.OpUserID)
	}

	if err := utils.CopyStructFields(resp.WorkMoment, workMoment); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields", err.Error())
	}
	user, err := imdb.GetUserByUserID(workMoment.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByUserID failed", err.Error())
	}
	if user != nil {
		resp.WorkMoment.FaceURL = user.FaceURL
		resp.WorkMoment.UserName = user.Nickname
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) GetUserWorkMoments(_ context.Context, req *pbOffice.GetUserWorkMomentsReq) (resp *pbOffice.GetUserWorkMomentsResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.GetUserWorkMomentsResp{CommonResp: &pbOffice.CommonResp{}, WorkMoments: []*pbOffice.WorkMoment{}}
	resp.Pagination = &pbCommon.ResponsePagination{CurrentPage: req.Pagination.PageNumber, ShowNumber: req.Pagination.ShowNumber}
	var workMoments []db.WorkMoment
	if req.UserID == req.OpUserID {
		workMoments, err = db.DB.GetUserSelfWorkMoments(req.UserID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	} else {
		workMoments, err = db.DB.GetUserWorkMoments(req.OpUserID, req.UserID, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	}
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserWorkMoments failed", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(&resp.WorkMoments, workMoments); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	for _, v := range resp.WorkMoments {
		user, err := imdb.GetUserByUserID(v.UserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		if user != nil {
			v.UserName = user.Nickname
			v.FaceURL = user.FaceURL
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) GetUserFriendWorkMoments(_ context.Context, req *pbOffice.GetUserFriendWorkMomentsReq) (resp *pbOffice.GetUserFriendWorkMomentsResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.GetUserFriendWorkMomentsResp{CommonResp: &pbOffice.CommonResp{}, WorkMoments: []*pbOffice.WorkMoment{}}
	resp.Pagination = &pbCommon.ResponsePagination{CurrentPage: req.Pagination.PageNumber, ShowNumber: req.Pagination.ShowNumber}
	workMoments, err := db.DB.GetUserFriendWorkMoments(req.Pagination.ShowNumber, req.Pagination.PageNumber, req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserFriendWorkMoments", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(&resp.WorkMoments, workMoments); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	for _, v := range resp.WorkMoments {
		user, err := imdb.GetUserByUserID(v.UserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		if user != nil {
			v.UserName = user.Nickname
			v.FaceURL = user.FaceURL
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) SetUserWorkMomentsLevel(_ context.Context, req *pbOffice.SetUserWorkMomentsLevelReq) (resp *pbOffice.SetUserWorkMomentsLevelResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.SetUserWorkMomentsLevelResp{CommonResp: &pbOffice.CommonResp{}}
	if err := db.DB.SetUserWorkMomentsLevel(req.UserID, req.Level); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetUserWorkMomentsLevel failed", err.Error())
		resp.CommonResp = &pbOffice.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *officeServer) ChangeWorkMomentPermission(_ context.Context, req *pbOffice.ChangeWorkMomentPermissionReq) (resp *pbOffice.ChangeWorkMomentPermissionResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.ChangeWorkMomentPermissionResp{CommonResp: &pbOffice.CommonResp{}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func addTencentCloudTagForObj(client *cos.Client, obj string) error {
	opt := &cos.ObjectPutTaggingOptions{
		TagSet: []cos.ObjectTaggingTag{
			{
				Key:   "delete",
				Value: "1",
			},
		},
	}
	_, err := client.Object.PutTagging(context.Background(), obj, opt)
	if err != nil {
		log.NewError("", "put tag is failed:", err.Error())
		return err
	}
	return nil
}

func removeTencentCloudTagForObj(client *cos.Client, obj string) error {
	_, err := client.Object.DeleteTagging(context.Background(), obj)
	if err != nil {
		log.NewError("", "delete tag is failed:", err.Error())
		return err
	}
	return nil
}

func getObjSize(client *cos.Client, obj string) (int64, error) {
	objMeta, err := client.Object.Head(context.Background(), obj, nil)
	if err != nil {
		log.NewError("", "get object meta failed, err:", err.Error(), "obj name: ", obj)
		return 0, err
	}
	contentLength, _ := strconv.ParseInt(objMeta.Header.Get("Content-Length"), 10, 64)
	return contentLength, nil
}

func addFavoriteFile(favorites *db.FavoritesSQL) error {
	// it's file, image, audio, video, record the counts
	if favorites.ContentType == constant.FavoriteContentTypeFile ||
		favorites.ContentType == constant.FavoriteContentTypeMedia || favorites.ContentType == constant.FavoriteContentTypeAudio {
		count, err := db.DB.AddFavoriteMedia(favorites.ObjName)
		if err != nil {
			log.NewError("", "increase count err: ", err.Error(), " content:", req.Content)
			return err
		}

		log.Debug("", "after increased, the count: ", count)
		if favorites.SourceType == constant.FavoriteSourceTypeChatting && count == 1 {
			var client *cos.Client
			client, err = admin_cms.GetTencentCloudClient(false)
			if err != nil {
				log.NewError("", "add tencent cloud tag get client err, err:", err.Error())
				return err
			}
			// this is first time to favorite this file, need to remove the tag "delete" for remove files after 30days
			err = removeTencentCloudTagForObj(client, favorites.ObjName)
			if err != nil {
				log.NewError("", "remove object tag failed, err: ", err.Error())
				return nil
			}
			if favorites.ThumbnailObjName != "" {
				err = removeTencentCloudTagForObj(client, favorites.ThumbnailObjName)
				if err != nil {
					log.NewError("", "remove ThumbnailObjName object tag failed, err: ", err.Error())
					return nil
				}
			}
		}
	} else if favorites.SourceType == constant.FavoriteSourceCombineChatting {
		client, err := admin_cms.GetTencentCloudClient(false)
		if err != nil {
			log.NewError("", "add tencent cloud tag get client err, err:", err.Error())
			return err
		}
		var mediasObjName []string
		_ = json.Unmarshal([]byte(favorites.MediasObjName), &mediasObjName)
		for _, objName := range mediasObjName {
			count, err := db.DB.AddFavoriteMedia(objName)
			if err != nil {
				log.NewError("", "increase count err: ", err.Error(), " content:", req.Content)
				return err
			}

			log.Debug("", "after increased, the count: ", count)
			if count == 1 {
				err = removeTencentCloudTagForObj(client, objName)
				if err != nil {
					fmt.Println("", "remove object tag failed, err: ", err.Error())
					log.NewError("", "remove object tag failed, err: ", err.Error())
					continue
				}
			}
		}
	} else if favorites.SourceType == constant.FavoriteSourceCombineChatting {
		client, err := admin_cms.GetTencentCloudClient(false)
		if err != nil {
			log.NewError("", "add tencent cloud tag get client err, err:", err.Error())
			return err
		}
		var mediasObjName []string
		_ = json.Unmarshal([]byte(favorites.MediasObjName), &mediasObjName)
		for _, objName := range mediasObjName {
			count, err := db.DB.AddFavoriteMedia(objName)
			if err != nil {
				log.NewError("", "increase count err: ", err.Error(), " content:", req.Content)
				return err
			}

			log.Debug("", "after increased, the count: ", count)
			if count == 1 {
				err = removeTencentCloudTagForObj(client, objName)
				if err != nil {
					fmt.Println("", "remove object tag failed, err: ", err.Error())
					log.NewError("", "remove object tag failed, err: ", err.Error())
					continue
				}
			}
		}
	}

	return nil
}

func removeFavoriteFile(favorites []*db.FavoritesSQL, userID string) error {

	// it's a file
	var obs []cos.Object
	client := &cos.Client{}
	for _, favorite := range favorites {
		log.Debug("", "favorite.UserID: ", favorite.UserID, " userID: ", userID)
		if favorite.UserID != userID {
			continue
		}
		log.Debug("", "favorite.ContentType: ", favorite.ContentType)
		if favorite.ContentType == constant.FavoriteContentTypeFile ||
			favorite.ContentType == constant.FavoriteContentTypeMedia || favorite.ContentType == constant.FavoriteContentTypeAudio {
			var count int64 = 0
			var err error
			// it's file, image, audio, video, record the counts
			count, err = db.DB.RemoveFavoriteMedia(favorite.ObjName)
			if err != nil {
				log.NewError("", "decrease count err: ", err.Error(), " content:", favorite.Content)
				return err
			}

			log.Debug("", "after decreased, the count: ", count, " favorite.SourceType: ", favorite.SourceType)
			if count == 0 {
				// the file haven't any one favorite yet, add tag for deleting.
				if favorite.SourceType == constant.FavoriteSourceTypeChatting {
					client, err = admin_cms.GetTencentCloudClient(false)
					if err != nil {
						return err
					}
					// chatting type, no persistent, just add the tag
					err = addTencentCloudTagForObj(client, favorite.ObjName)
					if err != nil {
						log.NewError("", "delete tag failed, err: ", err.Error(), " content:", favorite.Content)
						return nil
					}
				} else {
					client, err = admin_cms.GetTencentCloudClient(true)
					if err != nil {
						return err
					}
					// other type persistent, check if the source was deleted
					deleted := false

					switch favorite.SourceType {
					case constant.FavoriteSourceTypeMoment:
						var momentID primitive.ObjectID
						var moment = &db.Moment{}
						if momentID, err = primitive.ObjectIDFromHex(favorite.ContentID); err != nil {
							return err
						}
						if moment, err = db.DB.GetMoment(momentID); err != nil {
							return err
						}
						log.Debug("", "moment.DeleteTime: ", moment.DeleteTime)
						if moment != nil && moment.DeleteTime != 0 {
							deleted = true
						}
					case constant.FavoriteSourceTypeArticle:
						var article = &db.Article{}
						var articleId int64
						if articleId, err = strconv.ParseInt(favorite.ContentID, 10, 64); err != nil {
							return err
						}
						if article, err = db.DB.GetArticlesByID(articleId); err != nil {
							return err
						}
						if article != nil && article.DeleteTime != 0 {
							deleted = true
						}
					case constant.FavoriteSourceTypeShortVideo:
						var shortVideo = &db.ShortVideo{}
						if shortVideo, err = imdb.GetShortVideoByFileId(favorite.ContentID); err != nil {
							return err
						}
						if shortVideo != nil && shortVideo.Status != 5 {
							deleted = true
						}
					}
					log.Debug("", "deleted: ", deleted)
					if deleted {
						// delete the file
						obs = append(obs, cos.Object{Key: favorite.ObjName})
						if favorite.ThumbnailObjName != "" {
							obs = append(obs, cos.Object{Key: favorite.ThumbnailObjName})
						}
						log.Debug("", "delete obj: ", favorite.ObjName, " thumbnail obj name: ", favorite.ThumbnailObjName)
					}
				}
			}
		} else if favorite.SourceType == constant.FavoriteSourceCombineChatting {
			if favorite.MediasObjName != "" {
				var medias []string
				_ = json.Unmarshal([]byte(favorite.MediasObjName), &medias)
				var removeMedias []string
				for _, media := range medias {
					var count int64 = 0
					var err error
					// it's file, image, audio, video, record the counts
					count, err = db.DB.RemoveFavoriteMedia(media)
					if err != nil {
						log.NewError("", "decrease count err: ", err.Error(), " content:", media)
						return err
					}
					if count == 0 {
						removeMedias = append(removeMedias, media)
					}
				}
				if len(removeMedias) > 0 {
					var err error
					client, err = admin_cms.GetTencentCloudClient(false)
					if err != nil {
						return err
					}
					for _, media := range removeMedias {
						// chatting type, no persistent, just add the tag
						err = addTencentCloudTagForObj(client, media)
						if err != nil {
							log.NewError("", "delete tag failed, err: ", err.Error(), " content:", media)
							return err
						}
					}
				}
			}
		}
	}
	log.Debug("", "obs len: ", len(obs))
	if len(obs) > 0 {
		deleteOpt := &cos.ObjectDeleteMultiOptions{
			Objects: obs,
		}
		result, _, err := client.Object.DeleteMulti(context.Background(), deleteOpt)
		if err != nil {
			return err
		}
		if len(result.Errors) > 0 {
			for _, errxx := range result.Errors {
				log.NewError("", "err key: ", errxx.Key, " err code: ", errxx.Code, " err msg: ", errxx.Message)
			}
		}
	}

	return nil
}

func checkCanAddAndGetTotal(favorites *db.Favorites) error {
	contentObj := struct {
		Thumbnail string `json:"thumbnail"`
		Url       string `json:"url"`
	}{}
	// from chatting, all files are temporary
	if favorites.ContentType == constant.FavoriteContentTypeFile ||
		favorites.ContentType == constant.FavoriteContentTypeMedia || favorites.ContentType == constant.FavoriteContentTypeAudio {
		client := &cos.Client{}
		persistent := false
		var err error
		if favorites.SourceType != constant.FavoriteSourceTypeChatting {
			persistent = true
		}
		err = json.Unmarshal([]byte(favorites.Content), &contentObj)
		if err != nil {
			log.NewError("", "Content parse error", err.Error(), "Content: ", favorites.Content)
			return err
		}

		favorites.ObjName = admin_cms.GetObjNameByURL(contentObj.Url, persistent)
		favorites.ThumbnailObjName = admin_cms.GetObjNameByURL(contentObj.Thumbnail, persistent)

		client, err = admin_cms.GetTencentCloudClient(persistent)
		if err != nil {
			log.NewError("", "add tencent cloud tag get client err, err:", err.Error())
			return err
		}

		// get object size, if total size exceed 2G, failed.
		contentLength, err := getObjSize(client, favorites.ObjName)
		if err != nil {
			log.NewError("", "contentLength get error", err.Error(), favorites.ObjName)
			return err
		}

		log.Debug("", "obj len: ", contentLength)
		var capacity int64 = 0
		capacity, err = db.DB.GetUsedCapacity(favorites.UserID)
		if err != nil {
			log.NewError("", "get used capacity failed, err:", err.Error(), "user: ", favorites.UserID)
			capacity = 0
		}
		favorites.FileSize = contentLength
		if capacity+contentLength > config.Config.Favorite.MaxCapacity {
			log.NewError("", "Favorite are full, please release some space")
			return constant.ErrSendLimit
		}
	} else if favorites.SourceType == constant.FavoriteSourceCombineChatting {
		// get all files
		regeTool := regexp.MustCompile(fmt.Sprintf(`https:\/\/%s\.cos\..*?\.myqcloud\.com\/[^"]*`, config.Config.Credential.Tencent.Bucket))
		// the header won't he deleted, here needn't to deal with it.
		var medias = make(map[string]int64)
		urls := regeTool.FindAllString(favorites.Content, -1)
		for _, url := range urls {
			medias[url] = medias[url] + 1
		}

		client, err := admin_cms.GetTencentCloudClient(false)
		if err != nil {
			log.NewError("", "add tencent cloud tag get client err, err:", err.Error())
			return err
		}
		// calculate size
		var total int64 = 0
		var mediasObjName []string
		for url, count := range medias {
			objName := admin_cms.GetObjNameByURL(url, false)
			mediasObjName = append(mediasObjName, objName)
			contentLength, err := getObjSize(client, objName)
			if err != nil {
				log.NewError("", "contentLength get error", err.Error(), objName)
				return err
			}
			total += contentLength * count
		}
		mediasByte, _ := json.Marshal(mediasObjName)
		favorites.MediasObjName = string(mediasByte)

		var capacity int64 = 0
		capacity, err = db.DB.GetUsedCapacity(favorites.UserID)
		if err != nil {
			log.NewError("", "get used capacity failed, err:", err.Error(), "user: ", favorites.UserID)
			capacity = 0
		}
		if capacity+total > config.Config.Favorite.MaxCapacity {
			log.NewError("", "Favorite are full, please release some space")
			return constant.ErrSendLimit
		}
	}
	return nil
}

func (s *officeServer) AddFavorite(_ context.Context, req *pbOffice.AddFavoriteReq) (resp *pbOffice.AddFavoriteResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.AddFavoriteResp{CommonResp: &pbOffice.CommonResp{}}

	// set favorite
	favorite := db.Favorites{}

	favorite.CreateBy = req.OpUserID
	err = utils.CopyStructFields(&favorite, req)
	if err != nil {
		log.NewError("", "Copy favorite error: ", err.Error())
		errMsg := "Copy favorite error: " + err.Error()
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	// if the type is file, video, audio or image need to set the source to persistent.
	if err = checkCanAddAndGetTotal(&favorite); err != nil {
		if err == constant.ErrSendLimit {
			log.NewError("", "Add favorite failed, err: ", err.Error())
			errMsg := "Favorite are full, please release some space"
			resp.CommonResp.ErrCode = constant.ErrSendLimit.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, err
		}
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
		return resp, err
	}

	switch favorite.SourceType {
	case constant.FavoriteSourceCombineChatting:
		// set ex key words
		contentMap := map[string]interface{}{}
		_ = json.Unmarshal([]byte(favorite.Content), &contentMap)
		if content, ok := contentMap["key"]; ok {
			//type ContentTemp struct {
			//	Profile     string `json:"profile"`
			//	Username    string `json:"username"`
			//	Message     string `json:"message"`
			//	ContentType int32  `json:"content_type"`
			//	SendTime    int64  `json:"send_time"`
			//}
			contentList := content.([]interface{})
			for _, conInter := range contentList {
				con := conInter.(map[string]interface{})
				contentType := int64(con["content_type"].(float64))
				username := con["username"].(string)
				message := con["message"].(string)

				switch contentType {
				case constant.Text:
					if favorite.ExKeywords != "" {
						favorite.ExKeywords += fmt.Sprintf("-%s-%s",
							strings.ReplaceAll(username, "-", "--"),
							strings.ReplaceAll(message, "-", "--"))
					} else {
						favorite.ExKeywords += fmt.Sprintf("%s-%s",
							strings.ReplaceAll(username, "-", "--"),
							strings.ReplaceAll(message, "-", "--"))
					}
				case constant.AtText:
				case constant.Card:
				}
			}
		}
		fallthrough
	case constant.FavoriteSourceTypeChatting:
		chatLog := &db.ChatLog{}
		chatLog, err = imdb.GetChatLogWithClientMsgID(favorite.ContentID)
		if err != nil {
			log.Debug("", "get msg failed", favorite.ContentID)
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			return resp, err
		}
		favorite.ContentCreatorID = chatLog.SendID
		if chatLog.SessionType == constant.GroupChatType {
			favorite.ContentGroupID = chatLog.RecvID
		}
	case constant.FavoriteSourceTypeMoment:
		momentList := imdb.GetMoment([]string{favorite.ContentID})
		if len(momentList) != 1 {
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			return resp, err
		}
		favorite.ContentCreatorID = momentList[0].CreatorID
	case constant.FavoriteSourceTypeArticle:
		var articleID, _ = strconv.ParseInt(favorite.ContentID, 10, 64)
		article, err := imdb.GetAllArticleByArticleID(articleID)
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			return resp, err
		}
		favorite.ContentCreatorID = strconv.FormatInt(article.OfficialID, 10)
	case constant.FavoriteSourceTypeWeb:
		break
	case constant.FavoriteSourceTypeShortVideo:
		shortVideo, err := imdb.GetShortVideoByFileId(favorite.ContentID)
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			return resp, err
		}
		favorite.ContentCreatorID = shortVideo.UserId
	default:
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
		return resp, err
	}

	mongoID := primitive.NewObjectIDFromTimestamp(time.Now())
	favorite.FavoriteId = mongoID

	// save favorite to mongo db
	err = db.DB.AddFavorite(favorite)
	if err != nil {
		log.NewError("", "Add favorite failed, err: ", err.Error())
		errMsg := "Add favorite failed, err: " + err.Error()
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	// success, then save data to mysql and update media and update tencent, if need
	go func() {

		favoriteSQL := &db.FavoritesSQL{}
		err = utils.CopyStructFields(favoriteSQL, favorite)
		if err != nil {
			log.NewError("", "copy req to favorite is failed, err: ", err.Error())
			return
		}

		favoriteSQL.FavoriteId = mongoID.Hex()
		favoriteSQL.ContentID = req.ContentID

		// save favorite file
		err = addFavoriteFile(favoriteSQL)
		if err != nil {
			log.NewError("", "Add favorite failed, err: ", err.Error())
			return
		}

		// save favorite to mysql
		err = imdb.AddFavorites(favoriteSQL, req.OpUserID)
		if err != nil {
			log.NewError("", "Add favorite failed, err: ", err.Error())
			return
		}
	}()

	return resp, nil
}

func (s *officeServer) GetFavoriteList(_ context.Context, req *pbOffice.GetFavoriteListReq) (resp *pbOffice.GetFavoriteListResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.GetFavoriteListResp{CommonResp: &pbOffice.CommonResp{}, Favorites: []*pbOffice.Favorite{}}

	favorites, err := db.DB.GetFavorites(req.UserID, req.ContentType)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
		return resp, err
	}

	err = utils.CopyStructFields(&resp.Favorites, favorites)
	if err != nil {
		log.NewError("", "copy result failed, err: ", err.Error())
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
		return resp, err
	}

	for index, favorite := range favorites {
		resp.Favorites[index].FavoriteId = favorite.FavoriteId.Hex()
		user := &db.User{}
		creatorInfo := struct {
			ContentCreatorFaceURL string `json:"content_creator_face_url"`
			ContentGroupName      string `json:"content_group_name"`
			ContentCreatorName    string `json:"content_creator_name"`
		}{}
		if favorite.ContentCreatorID != "" {
			log.Debug("favorite.ContentCreatorID: ", "favorite.ContentCreatorID: ", favorite.ContentCreatorID)
			log.Debug("", "favorite: ", favorite)
			if favorite.SourceType == 3 {
				// from article, should find the official account.
				officialId, _ := strconv.ParseInt(favorite.ContentCreatorID, 10, 64)
				officialAccount, err := imdb.GetOfficialByOfficialIDAll(officialId)
				if err != nil {
					log.NewError("", "GetOfficialByOfficialID failed, err: ", err.Error(), " officialAccount: ", officialAccount)
					continue
				}
				creatorInfo.ContentCreatorName = officialAccount.Nickname
				creatorInfo.ContentCreatorFaceURL = officialAccount.FaceURL
			} else {
				user, err = imdb.GetUserByUserIDAll(favorite.ContentCreatorID)
				if user != nil {
					creatorInfo.ContentCreatorName = user.Nickname
					creatorInfo.ContentCreatorFaceURL = user.FaceURL
				}
			}
		}

		if favorite.ContentGroupID != "" {
			group, err := imdb.GetGroupInfoByGroupID(favorite.ContentGroupID)
			if err != nil {
				log.NewError("", "GetGroupInfoByGroupID failed, err: ", err.Error())
				continue
			}
			creatorInfo.ContentGroupName = group.GroupName
		}
		contentCreatorInfo, err := json.Marshal(creatorInfo)
		if err != nil {
			log.NewError("", "contentCreatorInfo failed, err: ", err.Error())
			continue
		}
		resp.Favorites[index].ContentCreatorName = string(contentCreatorInfo)
	}

	return resp, err
}

func (s *officeServer) RemoveFavorite(_ context.Context, req *pbOffice.RemoveFavoriteReq) (resp *pbOffice.RemoveFavoriteResp, err error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp = &pbOffice.RemoveFavoriteResp{CommonResp: &pbOffice.CommonResp{}}

	favorite := db.Favorites{}
	favorite.UserID = req.UserID
	favorite.DeletedBy = req.OpUserId

	// remove from mongo db
	for _, id := range req.FavoriteIds {
		favorite.FavoriteId, _ = primitive.ObjectIDFromHex(id)
		err = db.DB.RemoveFavorites(favorite)
		if err != nil {
			log.NewError("", "remove favorite from mongoDB failed, err: ", err.Error(), " favorite id:", favorite.FavoriteId)
			resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
			return resp, err
		}
	}

	// Remove favorite from mysql and deal with files
	go func() {

		favorites, err := imdb.GetFavoritesByFavIds(req.FavoriteIds)
		if err != nil {
			log.NewError("", "GetFavoritesByFavIds failed, err: ", err.Error())
			return
		}

		err = removeFavoriteFile(favorites, favorite.UserID)
		if err != nil {
			log.NewError("", "removeFavoriteFile failed, err: ", err.Error())
			return
		}

		err = imdb.DeleteFavorites(req.FavoriteIds)
		if err != nil {
			log.NewError("", "delete tag from mysql failed, err: ", err.Error())
			return
		}

	}()

	return resp, nil

}

func (s *officeServer) GetFavorites(_ context.Context, req *pbOffice.GetFavoritesReq) (*pbOffice.GetFavoritesResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbOffice.GetFavoritesResp{Favorites: []*pbOffice.FavoriteManager{}}

	where := map[string]string{}
	where["account"] = req.Account
	where["content"] = req.Content
	where["publish_user"] = req.PublishUser
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	favoriteDBList, favoriteCounts, err := imdb.GetFavoritesByWhere(where, req.ContentType, req.Pagination.PageNumber, req.Pagination.ShowNumber, req.OrderBy)

	if err != nil {
		log.NewError("", "query favorites err: ", err.Error())
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
		return resp, err
	}

	err = utils.CopyStructFields(&resp.Favorites, favoriteDBList)
	if err != nil {
		log.NewError("", "copy result failed, err: ", err.Error())
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
		return resp, err
	}

	resp.FavoriteNums = favoriteCounts
	resp.Pagination = &sdk_ws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil

}

func (s *officeServer) AlterFavorites(_ context.Context, req *pbOffice.AlterFavoritesReq) (*pbOffice.AlterFavoritesResp, error) {
	log.NewInfo(req.OperationId, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbOffice.AlterFavoritesResp{CommonResp: &pbOffice.CommonResp{}}

	// set favorite
	favorite := db.FavoritesSQL{}
	_, err := db.DB.GetFavoritesById(req.FavoriteId)
	if err != nil {
		log.NewError("", "the favorite isn't exist: ", err.Error())
		errMsg := "the favorite isn't exist: " + err.Error()
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	favorite.FavoriteId = req.FavoriteId
	favorite.Remark = req.Remark
	// update MySql
	err = imdb.UpdateFavorite(&favorite, req.OpUserID)
	if err != nil {
		log.NewError("", "Update error: ", err.Error(), " Content: ", favorite.FavoriteId)
		errMsg := "Update error: " + err.Error() + " Content: " + favorite.FavoriteId
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return nil, err
	}

	return resp, nil
}

//func (s *officeServer) AlterFavorites(_ context.Context, req *pbOffice.AlterFavoritesReq) (*pbOffice.AlterFavoritesResp, error) {
//	log.NewInfo(req.OperationId, utils.GetSelfFuncName(), "req: ", req.String())
//	resp := &pbOffice.AlterFavoritesResp{CommonResp: &pbOffice.CommonResp{}}
//
//	// set favorite
//	favorite := db.FavoritesSQL{}
//	oldFavorite, err := db.DB.GetFavoritesById(req.FavoriteId)
//	if err != nil {
//		log.NewError("", "the favorite isn't exist: ", err.Error())
//		errMsg := "the favorite isn't exist: " + err.Error()
//		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
//		resp.CommonResp.ErrMsg = errMsg
//		return resp, err
//	}
//
//	favorite.CreateBy = req.OpUserID
//	err = utils.CopyStructFields(&favorite, req)
//	if err != nil {
//		log.NewError("", "Copy favorite error: ", err.Error())
//		errMsg := "Copy favorite error: " + err.Error()
//		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
//		resp.CommonResp.ErrMsg = errMsg
//		return resp, err
//	}
//
//	client, err := GetTencentCloudClient()
//	if err != nil {
//		log.NewError("", "add tencent cloud tag get client err, err:", err.Error())
//		errMsg := "add tencent cloud tag get client err, err:" + err.Error()
//		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
//		resp.CommonResp.ErrMsg = errMsg
//		return resp, err
//	}
//
//	var totalUsed int64 = 0
//	contentObj := struct{
//		Thumbnail string `json:"thumbnail"`
//		Url       string `json:"url"`
//	}{}
//	isFile := false
//	// if the type is file, video, audio or image need to set the source to persistent.
//	if favorite.SourceType == 1 {
//		// from chatting, all files are temporary
//		if favorite.ContentType == 1 || favorite.ContentType == 3 || favorite.ContentType == 4 {
//			isFile = true
//			dir := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.Bucket, config.Config.Credential.Tencent.Region)
//			err = json.Unmarshal([]byte(favorite.Content), &contentObj)
//			if err != nil {
//				log.NewError("", "Content parse error", err.Error(), "Content: ", favorite.Content)
//				errMsg := "Content parse error" + err.Error() + "Content: " + favorite.Content
//				resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
//				resp.CommonResp.ErrMsg = errMsg
//				return nil, err
//			}
//			favorite.ObjName = strings.Replace(contentObj.Url, dir+"/", "", 1)
//			favorite.ThumbnailObjName = strings.Replace(contentObj.Thumbnail, dir+"/", "", 1)
//
//			// get object size, if total size exceed 2G, failed.
//			contentLength, err := getObjSize(client, favorite.ObjName)
//			if err != nil {
//				log.NewError("", "contentLength get error", err.Error(), favorite.ObjName)
//				errMsg := "contentLength get error" + err.Error() + favorite.ObjName
//				resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
//				resp.CommonResp.ErrMsg = errMsg
//				return nil, err
//			}
//			oldContentLength, _ := getObjSize(client, oldFavorite.ObjName)
//
//			log.Debug("", "obj len: ", contentLength)
//			var capacity int64 = 0
//			capacity, err = db.DB.GetFavoriteUsedCapacity(req.UserID)
//			if err != nil {
//				log.NewError("", "get used capacity failed, err:", err.Error(), "user: ", req.UserID)
//				capacity = 0
//			}
//			if req.ContentType == 1 {
//				favorite.FileSize = contentLength
//			}
//
//			totalUsed = capacity - oldContentLength + contentLength
//			if totalUsed > config.Config.Favorite.MaxCapacity {
//				log.NewError("","Favorite are full, please release some space")
//				errMsg := "Favorite are full, please release some space"
//				resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
//				resp.CommonResp.ErrMsg = errMsg
//				return resp, nil
//			}
//		}
//	}
//
//	// update MySql
//	err = imdb.UpdateFavorite(&favorite, req.OpUserID)
//	if err != nil {
//		log.NewError("", "Update error: ", err.Error(), " Content: ", favorite.FavoriteId)
//		errMsg := "Update error: " + err.Error() + " Content: " + favorite.FavoriteId
//		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
//		resp.CommonResp.ErrMsg = errMsg
//		return nil, err
//	}
//
//	go func() {
//		favoriteMongo := db.Favorites{}
//		err = utils.CopyStructFields(&favoriteMongo, favorite)
//		if err != nil {
//			log.NewError("", "copy favorite to mongo bean failed", err.Error())
//			return
//		}
//		favoriteMongo.FavoriteId, _ = primitive.ObjectIDFromHex(favorite.FavoriteId)
//
//		err = db.DB.UpdateFavorite(favoriteMongo, req.OpUserID)
//		if err != nil {
//			log.NewError("", "UpdateFavorite failed", err.Error())
//			return
//		}
//
//		// remove favorite from tencent cloud and update redis
//		err = removeFavoriteFile(client, &favorite)
//		if err != nil {
//			log.NewError("", "removeFavoriteFile failed", err.Error())
//			return
//		}
//
//		// add new favorite
//		if isFile {
//			err = addFavoriteFile(client, &favorite, totalUsed)
//			if err != nil {
//				log.NewError("", "addFavoriteFile failed", err.Error())
//				return
//			}
//		}
//	}()
//
//	return resp, nil
//}

func (s *officeServer) GetCommunications(_ context.Context, req *pbOffice.GetCommunicationsReq) (*pbOffice.GetCommunicationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbOffice.GetCommunicationsResp{CommunicationList: []*pbOffice.Communication{}}

	where := map[string]string{}
	where["originator"] = req.Originator
	where["member"] = req.Member
	where["originator_platform"] = strconv.Itoa(int(req.OriginatorPlatform))
	where["chat_type"] = strconv.Itoa(int(req.ChatType))
	where["duration"] = strconv.FormatInt(req.Duration, 10)
	where["status"] = strconv.Itoa(int(req.Status))
	where["remark"] = req.Remark
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["room_id"] = req.RoomID

	communicationList, recordCounts, err := imdb.GetCommunicationsByWhere(where, req.CommunicationType, req.Pagination.PageNumber, req.Pagination.ShowNumber, req.OrderBy)

	if err != nil {
		log.NewError("", "query favorites err: ", err.Error())
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
		return resp, err
	}

	err = utils.CopyStructFields(&resp.CommunicationList, communicationList)
	if err != nil {
		log.NewError("", "copy result failed, err: ", err.Error())
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrServer.ErrMsg
		return resp, err
	}

	resp.Communications = int32(recordCounts)
	resp.Pagination = &sdk_ws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil

}

func (s *officeServer) DeleteCommunications(_ context.Context, req *pbOffice.DeleteCommunicationsReq) (*pbOffice.DeleteCommunicationsResp, error) {
	log.Debug(req.OperationID, "Delete official req: ", req.String())
	resp := &pbOffice.DeleteCommunicationsResp{CommonResp: &pbOffice.CommonResp{}}
	if len(req.CommunicatIDs) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	if err := imdb.DeleteCommunications(req.CommunicatIDs, req.OpUserID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete err:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (s *officeServer) SetRemark(_ context.Context, req *pbOffice.SetRemarkReq) (*pbOffice.SetRemarkResp, error) {
	resp := &pbOffice.SetRemarkResp{CommonResp: &pbOffice.CommonResp{}}
	record := &db.VideoAudioCommunicationRecord{CommunicationID: req.CommunicatID, Remark: req.Remark, UpdateBy: req.OpUserID, UpdateTime: time.Now().Unix()}
	log.Debug("", "alter communicate remark: ", req.CommunicatID)

	if err := imdb.UpdateCommunication(record); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "update failed", "update err:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (s *officeServer) InterruptPersonalCommunications(_ context.Context, req *pbOffice.InterruptPersonalCommunicationsReq) (*pbOffice.InterruptPersonalCommunicationsResp, error) {
	resp := &pbOffice.InterruptPersonalCommunicationsResp{}

	nowTime := time.Now().Unix()
	record := &db.VideoAudioCommunicationRecord{
		CommunicationID: req.CommunicatID,
		Status:          4,
		ErrCode:         12,
		ErrMsgEN:        "Communication interrupted",
		EndTime:         nowTime,
		UpdateBy:        req.OpUserID,
		UpdateTime:      nowTime,
	}
	log.Debug("", "alter communicate remark: ", req.CommunicatID)

	if err := imdb.UpdateCommunication(record); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "update failed", "update err:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil

}
