package group

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/group"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	pbGroup "Open_IM/pkg/proto/group"

	"github.com/gin-gonic/gin"
)

func GetGroupById(c *gin.Context) {
	var (
		req   cms_api_struct.GetGroupByIdRequest
		resp  cms_api_struct.GetGroupByIdResponse
		reqPb pbGroup.GetGroupByIdReq
	)
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupId = req.GroupId
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.GetGroupById(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetGroupById failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	// check the group is not existed
	if respPb.CMSGroup.GroupInfo.CreatorUserID == "" {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "group is not existed")
		openIMHttp.RespHttp200(c, constant.OK, nil)
		return
	}

	for _, member := range respPb.CMSGroup.MemberList {
		resp.MemberList = append(resp.MemberList, cms_api_struct.MemberSimple{UserName: member.UserName, UserID: member.UserID})
	}
	resp.GroupName = respPb.CMSGroup.GroupInfo.GroupName
	resp.GroupID = respPb.CMSGroup.GroupInfo.GroupID
	resp.Notification = respPb.CMSGroup.GroupInfo.Notification
	resp.Introduction = respPb.CMSGroup.GroupInfo.Introduction
	resp.CreateTime = (utils.UnixSecondToTime(int64(respPb.CMSGroup.GroupInfo.CreateTime))).String()
	resp.ProfilePhoto = respPb.CMSGroup.GroupInfo.FaceURL
	resp.GroupMasterName = respPb.CMSGroup.GroupMasterName
	resp.GroupMasterId = respPb.CMSGroup.GroupMasterId
	resp.IsOpen = respPb.CMSGroup.IsOpen
	resp.Remark = respPb.CMSGroup.GroupInfo.Remark
	resp.IsBanChat = constant.GroupIsBanChat(respPb.CMSGroup.GroupInfo.Status)
	resp.IsBanPrivateChat = constant.GroupIsBanPrivateChat(respPb.CMSGroup.GroupInfo.Status)
	log.NewInfo("", utils.GetSelfFuncName(), "req: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetGroups(c *gin.Context) {
	var (
		req   cms_api_struct.GetGroupsRequest
		resp  cms_api_struct.GetGroupsResponse
		reqPb pbGroup.GetGroupsReq
	)
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.Pagination = &commonPb.RequestPagination{}
	utils.CopyStructFields(&reqPb.Pagination, req)
	utils.CopyStructFields(&reqPb, req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.GetGroups(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetUserInfo failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	for _, v := range respPb.CMSGroups {
		resp.Groups = append(resp.Groups, cms_api_struct.GroupResponse{
			GroupName:        v.GroupInfo.GroupName,
			GroupID:          v.GroupInfo.GroupID,
			GroupMasterName:  v.GroupMasterName,
			GroupMasterId:    v.GroupMasterId,
			CreateTime:       (utils.UnixSecondToTime(int64(v.GroupInfo.CreateTime))).String(),
			IsBanChat:        constant.GroupIsBanChat(v.GroupInfo.Status),
			IsBanPrivateChat: constant.GroupIsBanPrivateChat(v.GroupInfo.Status),
			ProfilePhoto:     v.GroupInfo.FaceURL,
			Remark:           v.GroupInfo.Remark,
			Introduction:     v.GroupInfo.Introduction,
			Notification:     v.GroupInfo.Notification,
			VideoStatus:      int8(v.GroupInfo.VideoStatus),
			AudioStatus:      int8(v.GroupInfo.AudioStatus),
			Members:          v.GroupInfo.MemberCount,
			IsOpen:           v.IsOpen,
		})
	}
	resp.GroupNums = int(respPb.GroupNum)
	resp.CurrentPage = int(respPb.Pagination.PageNumber)
	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
	log.NewInfo("", utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetGroupByName(c *gin.Context) {
	var (
		req   cms_api_struct.GetGroupRequest
		resp  cms_api_struct.GetGroupResponse
		reqPb pbGroup.GetGroupReq
	)
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupName = req.GroupName
	reqPb.Pagination = &commonPb.RequestPagination{}
	utils.CopyStructFields(&reqPb.Pagination, req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.GetGroup(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetGroup failed", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}
	for _, v := range respPb.CMSGroups {
		resp.Groups = append(resp.Groups, cms_api_struct.GroupResponse{
			GroupName:        v.GroupInfo.GroupName,
			GroupID:          v.GroupInfo.GroupID,
			GroupMasterName:  v.GroupMasterName,
			GroupMasterId:    v.GroupMasterId,
			CreateTime:       (utils.UnixSecondToTime(int64(v.GroupInfo.CreateTime))).String(),
			IsBanChat:        constant.GroupIsBanChat(v.GroupInfo.Status),
			IsBanPrivateChat: constant.GroupIsBanPrivateChat(v.GroupInfo.Status),
			ProfilePhoto:     v.GroupInfo.FaceURL,
		})
	}
	resp.CurrentPage = int(respPb.Pagination.PageNumber)
	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
	resp.GroupNums = int(respPb.GroupNums)
	log.NewInfo("", utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func CreateGroup(c *gin.Context) {
	var (
		req   cms_api_struct.CreateGroupRequest
		_     cms_api_struct.CreateGroupResponse
		reqPb pbGroup.CreateGroupReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupInfo = &commonPb.GroupInfo{}
	reqPb.GroupInfo.GroupName = req.GroupName
	reqPb.GroupInfo.Introduction = req.Introduction
	reqPb.GroupInfo.Notification = req.Notification
	reqPb.GroupInfo.CreatorUserID = req.GroupMasterId
	reqPb.OwnerUserID = req.GroupMasterId
	reqPb.OpUserID = req.GroupMasterId
	reqPb.GroupInterest = req.Interests

	reqPb.GroupInfo.GroupID = req.GroupId
	reqPb.GroupInfo.GroupType = req.GroupType
	reqPb.GroupInfo.Remark = req.Remark
	reqPb.GroupInfo.FaceURL = req.FaceURL
	reqPb.IsOpen = req.IsOpen
	for _, v := range req.GroupMembers {
		reqPb.InitMemberList = append(reqPb.InitMemberList, &pbGroup.GroupAddMemberInfo{
			UserID:    v,
			RoleLevel: 1,
		})
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	// group id is exists
	if req.GroupId != "" {
		result, getErr := client.GetGroupById(context.Background(), &pbGroup.GetGroupByIdReq{
			OperationID: reqPb.OperationID,
			GroupId:     req.GroupId,
		})

		if getErr == nil && (result.CMSGroup.GroupInfo.GroupID == req.GroupId && result.CMSGroup.GroupInfo.CreatorUserID != "") {
			openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: 501, ErrMsg: "group id is exist"}, nil)
			return
		}
	}

	re, err := client.CreateGroup(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "CreateGroup failed", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: re.ErrCode, ErrMsg: re.ErrMsg}, nil)
}

func BanGroupChat(c *gin.Context) {
	var (
		req   cms_api_struct.BanGroupChatRequest
		reqPb pbGroup.OperateGroupStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupId = req.GroupId
	reqPb.Status = constant.GroupBanChat
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateGroupStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BanGroupChat failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func BanPrivateChat(c *gin.Context) {
	var (
		req   cms_api_struct.BanPrivateChatRequest
		reqPb pbGroup.OperateGroupStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupId = req.GroupId
	reqPb.Status = constant.GroupBanPrivateChat
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateGroupStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "OperateGroupStatus failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func OpenGroupChat(c *gin.Context) {
	var (
		req   cms_api_struct.BanPrivateChatRequest
		reqPb pbGroup.OperateGroupStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupId = req.GroupId
	reqPb.Status = constant.GroupOk
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateGroupStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "OperateGroupStatus failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func OpenPrivateChat(c *gin.Context) {
	var (
		req   cms_api_struct.BanPrivateChatRequest
		reqPb pbGroup.OperateGroupStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "OpenPrivateChat failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupId = req.GroupId
	reqPb.Status = constant.GroupOk
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateGroupStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "OperateGroupStatus failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetGroupMembers(c *gin.Context) {
	var (
		req   cms_api_struct.GetGroupMembersRequest
		reqPb pbGroup.GetGroupMembersCMSReq
		resp  cms_api_struct.GetGroupMembersResponse
	)
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.Pagination = &commonPb.RequestPagination{
		PageNumber: int32(req.PageNumber),
		ShowNumber: int32(req.ShowNumber),
	}
	_ = utils.CopyStructFields(&reqPb, req)
	if req.RoleLevel != "" {
		err := json.Unmarshal([]byte(req.RoleLevel), &reqPb.RoleLevel)
		if err != nil {
			log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "role level failed", err.Error())
			openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
			return
		}
	}
	if req.Status != "" {
		err := json.Unmarshal([]byte(req.Status), &reqPb.Status)
		if err != nil {
			log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "Status failed", err.Error())
			openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
			return
		}
	}
	if req.Permission != "" {
		err := json.Unmarshal([]byte(req.Permission), &reqPb.Permission)
		if err != nil {
			log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "Permission failed", err.Error())
			openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
			return
		}
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.GetGroupMembersCMS(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetGroupMembersCMS failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.ResponsePagination = cms_api_struct.ResponsePagination{
		CurrentPage: int(respPb.Pagination.CurrentPage),
		ShowNumber:  int(respPb.Pagination.ShowNumber),
	}
	resp.MemberNums = int(respPb.MemberNums)
	for _, groupMembers := range respPb.Members {
		resp.GroupMembers = append(resp.GroupMembers, cms_api_struct.GroupMemberResponse{
			MemberPosition: int(groupMembers.RoleLevel),
			MemberNickName: groupMembers.Nickname,
			MemberId:       groupMembers.UserID,
			JoinTime:       utils.UnixSecondToTime(int64(groupMembers.JoinTime)).String(),
			MemberFaceURL:  groupMembers.FaceURL,
			RoleLevel:      groupMembers.RoleLevel,
			JoinSource:     groupMembers.JoinSource,
			MuteEndTime:    groupMembers.MuteEndTime,
			Remark:         groupMembers.Remark,
			VideoStatus:    int8(groupMembers.VideoStatus),
			AudioStatus:    int8(groupMembers.AudioStatus),
		})
	}
	log.NewInfo("", utils.GetSelfFuncName(), "req: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AddGroupMembers(c *gin.Context) {
	var (
		req   cms_api_struct.RemoveGroupMembersRequest
		resp  cms_api_struct.RemoveGroupMembersResponse
		reqPb pbGroup.AddGroupMembersCMSReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationId, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationId = utils.OperationIDGenerator()
	log.NewInfo("", utils.GetSelfFuncName(), "req: ", req)

	reqPb.UserIds = req.Members
	reqPb.GroupId = req.GroupId
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationId)
	if etcdConn == nil {
		errMsg := reqPb.OperationId + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationId, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.AddGroupMembersCMS(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationId, utils.GetSelfFuncName(), "AddGroupMembersCMS failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.Success = respPb.Success
	resp.Failed = respPb.Failed
	log.NewInfo("", utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func RemoveGroupMembers(c *gin.Context) {
	var (
		req   cms_api_struct.RemoveGroupMembersRequest
		resp  cms_api_struct.RemoveGroupMembersResponse
		reqPb pbGroup.RemoveGroupMembersCMSReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.UserIds = req.Members
	reqPb.GroupId = req.GroupId
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.RemoveGroupMembersCMS(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "RemoveGroupMembersCMS failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.Success = respPb.Success
	resp.Failed = respPb.Failed
	log.NewInfo("", utils.GetSelfFuncName(), "req: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func DeleteGroup(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteGroupRequest
		_     cms_api_struct.DeleteGroupResponse
		reqPb pbGroup.DeleteGroupReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.GroupId = req.GroupId
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.DeleteGroup(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SetGroupMaster(c *gin.Context) {
	var (
		req   cms_api_struct.SetGroupMasterRequest
		_     cms_api_struct.SetGroupMasterResponse
		reqPb pbGroup.OperateUserRoleReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	reqPb.GroupId = req.GroupId
	reqPb.UserId = req.UserId
	reqPb.RoleLevel = constant.GroupOwner
	reqPb.OpUserID = userID
	reqPb.OpFrom = constant.OpFromAdmin
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateUserRole(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SetGroupOrdinaryUsers(c *gin.Context) {
	var (
		req   cms_api_struct.SetGroupMemberRequest
		_     cms_api_struct.AdminLoginResponse
		reqPb pbGroup.OperateUserRoleReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.GroupId = req.GroupId
	reqPb.UserId = req.UserId
	reqPb.RoleLevel = constant.GroupOrdinaryUsers
	reqPb.OpUserID = userID
	reqPb.OpFrom = constant.OpFromAdmin
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateUserRole(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SetGroupAdmin(c *gin.Context) {
	var (
		req   cms_api_struct.SetGroupAdminRequest
		_     cms_api_struct.SetGroupAdminResponse
		reqPb pbGroup.OperateUserRoleReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.GroupId = req.GroupId
	reqPb.UserId = req.UserId
	reqPb.RoleLevel = constant.GroupAdmin
	reqPb.OpUserID = userID
	reqPb.OpFrom = constant.OpFromAdmin
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.OperateUserRole(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterGroupInfo(c *gin.Context) {
	var (
		req   cms_api_struct.AlterGroupInfoRequest
		_     cms_api_struct.SetGroupMasterResponse
		reqPb pbGroup.SetGroupInfoReq
	)
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OpUserID = c.MustGet("userID").(string)
	reqPb.GroupInfo = &commonPb.GroupInfo{
		GroupID:      req.GroupID,
		GroupName:    req.GroupName,
		Introduction: req.Introduction,
		Notification: req.Notification,
		FaceURL:      req.ProfilePhoto,
		GroupType:    int32(req.GroupType),
		Remark:       req.Remark,
	}
	reqPb.GroupInterest = req.Interests
	reqPb.IsAdmin = true
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.SetGroupInfo(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func MuteGroupMember(c *gin.Context) {
	params := cms_api_struct.MuteGroupMemberReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &pbGroup.MuteGroupMemberReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, exsited := c.Get("userID")
	if exsited {
		req.OpUserID = userIDInter.(string)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "err_msg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	reply, err := client.MuteGroupMember(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "err_msg": err.Error()})
		return
	}

	resp := cms_api_struct.MuteGroupMemberResp{CommResp: cms_api_struct.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func CancelMuteGroupMember(c *gin.Context) {
	params := cms_api_struct.CancelMuteGroupMemberReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.CancelMuteGroupMemberReq{}
	utils.CopyStructFields(req, &params)

	var ok bool
	var errInfo string
	ok, req.OpUserID, errInfo = token_verify.GetUserIDFromToken(c.Request.Header.Get("token"), req.OperationID)
	if !ok {
		errMsg := req.OperationID + " " + "GetUserIDFromToken failed " + errInfo + " token:" + c.Request.Header.Get("token")
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	//check user status
	if err := utils2.CheckUserPermissions(req.OpUserID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "user is banned!", req.OpUserID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	reply, err := client.CancelMuteGroupMember(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "err_msg": err.Error()})
		return
	}

	resp := cms_api_struct.CancelMuteGroupMemberResp{CommResp: cms_api_struct.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	c.JSON(http.StatusOK, gin.H{"code": reply.CommonResp.ErrCode, "err_msg": reply.CommonResp.ErrMsg})
}

func SetVideoAudioStatus(c *gin.Context) {
	var (
		req   cms_api_struct.SetVideoAudioStatusRequest
		_     cms_api_struct.SetVideoAudioStatusResponse
		reqPb pbGroup.SetVideoAudioStatusReq
	)
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OpUserID = c.MustGet("userID").(string)
	utils.CopyStructFields(&reqPb, req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.SetVideoAudioStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SetUserVideoAudioStatus(c *gin.Context) {
	var (
		req   cms_api_struct.SetUserVideoAudioStatusRequest
		_     cms_api_struct.SetUserVideoAudioStatusResponse
		reqPb pbGroup.SetUserVideoAudioStatusReq
	)
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OpUserID = c.MustGet("userID").(string)
	utils.CopyStructFields(&reqPb, req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	_, err := client.SetUserVideoAudioStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "DeleteGroup failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func DismissGroup(c *gin.Context) {
	params := api.DismissGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.DismissGroupReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}
	//Check if the user is admin or not
	isAdminUserInterface, existed := c.Get("isAdminUser")
	if existed {
		req.IsAdminUser = isAdminUserInterface.(bool)
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	reply, err := client.DismissGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.DismissGroupResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetUsersByGroup(c *gin.Context) {
	var (
		params cms_api_struct.GetUsersByGroupReq
		resp   cms_api_struct.GetUsersByGroupResp
		reqPb  pbGroup.GetUsersByGroupReq
	)

	if err := c.BindQuery(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", params)

	reqPb.Pagination = &commonPb.RequestPagination{
		PageNumber: int32(params.PageNumber),
		ShowNumber: int32(params.ShowNumber),
	}
	_ = utils.CopyStructFields(&reqPb, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbGroup.NewGroupClient(etcdConn)
	respPb, err := client.GetUsersByGroup(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetGroupMembersCMS failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.ResponsePagination = cms_api_struct.ResponsePagination{
		CurrentPage: int(respPb.Pagination.CurrentPage),
		ShowNumber:  int(respPb.Pagination.ShowNumber),
	}
	resp.UserNums = respPb.UserNums
	for _, user := range respPb.Users {
		resp.Users = append(resp.Users, cms_api_struct.UserIDAndName{
			UserID:   user.UserID,
			UserName: user.UserName,
		})
	}
	log.NewInfo("", utils.GetSelfFuncName(), "req: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}
