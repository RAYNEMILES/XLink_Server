package group

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbGroup "Open_IM/pkg/proto/group"
	rpc "Open_IM/pkg/proto/group"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"

	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/gin-gonic/gin"

	"net/http"
	"strings"

	jsonData "Open_IM/internal/utils"
)

func KickGroupMember(c *gin.Context) {
	params := api.KickGroupMemberReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	req := &rpc.KickGroupMemberReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "KickGroupMember args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.KickGroupMember(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberList failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	var memberListResp api.KickGroupMemberResp
	memberListResp.ErrMsg = RpcResp.ErrMsg
	memberListResp.ErrCode = RpcResp.ErrCode
	for _, v := range RpcResp.Id2ResultList {
		memberListResp.UserIDResultList = append(memberListResp.UserIDResultList, &api.UserIDResult{UserID: v.UserID, Result: v.Result})
	}
	if len(memberListResp.UserIDResultList) == 0 {
		memberListResp.UserIDResultList = []*api.UserIDResult{}
	}

	log.NewInfo(req.OperationID, "KickGroupMember api return ", memberListResp)
	c.JSON(http.StatusOK, memberListResp)
}

func GetGroupMembersInfo(c *gin.Context) {
	params := api.GetGroupMembersInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetGroupMembersInfoReq{}
	utils.CopyStructFields(req, params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetGroupMembersInfo args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)

	RpcResp, err := client.GetGroupMembersInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberList failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	memberListResp := api.GetGroupMembersInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}, MemberList: RpcResp.MemberList}
	memberListResp.Data = jsonData.JsonDataList(RpcResp.MemberList)
	log.NewInfo(req.OperationID, "GetGroupMembersInfo api return ", memberListResp)
	c.JSON(http.StatusOK, memberListResp)
}

func GetGroupMemberList(c *gin.Context) {
	params := api.GetGroupMemberListReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetGroupMemberListReq{}
	utils.CopyStructFields(req, params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetGroupMemberList args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)

	RpcResp, err := client.GetGroupMemberList(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupMemberList failed, ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	memberListResp := api.GetGroupMemberListResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}, MemberList: RpcResp.MemberList, NextSeq: RpcResp.NextSeq}
	memberListResp.Data = jsonData.JsonDataList(memberListResp.MemberList)

	log.NewInfo(req.OperationID, "GetGroupMemberList api return ", memberListResp)
	c.JSON(http.StatusOK, memberListResp)
}

func GetGroupMemberListV2(c *gin.Context) {

	var (
		params api.GetGroupMemberListSrvReq
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	reqPb := pbGroup.GetGroupMembersReqV2{}
	reqPb.Filter = params.Filter
	reqPb.GroupId = params.GroupId
	reqPb.Limit = params.Limit
	reqPb.Offset = params.Offset
	reqPb.SearchName = params.SearchName

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)

	RpcResp, err := client.GetGroupMemberListV2(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, "GetGroupMemberList failed, ", err.Error(), reqPb.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	// memberListResp := api.GetGroupMemberListSrvRes{}
	// for _, gm := range RpcResp.Members {
	// 	gmObj := api.GroupMemberResponse{}
	// 	// utils.CopyStructFields(gmObj, gm)
	// 	gmObj.MemberId = gm.UserID
	// 	gmObj.MemberNickName = gm.Nickname
	// 	gmObj.JoinTime = utils.UnixSecondToTime(int64(gm.JoinTime)).String()
	// 	gmObj.MemberFaceURL = gm.FaceURL
	// 	gmObj.RoleLevel = gm.RoleLevel
	// 	gmObj.JoinSource = gm.JoinSource
	// 	gmObj.MuteEndTime = gm.MuteEndTime

	// 	memberListResp.GroupMembers = append(memberListResp.GroupMembers, gmObj)
	// }
	customeResponse := make(map[string]interface{})
	customeResponse["errCode"] = 0
	customeResponse["errMsg"] = ""
	customeResponse["data"] = jsonData.JsonDataOne(RpcResp)
	// log.NewInfo(reqPb.OperationID, "GetGroupMemberList api return ", memberListResp)

	c.JSON(http.StatusOK, customeResponse)
}

func GetGroupAllMemberList(c *gin.Context) {
	params := api.GetGroupAllMemberReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetGroupAllMemberReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetGroupAllMember args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.GetGroupAllMember(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupAllMember failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	memberListResp := api.GetGroupAllMemberResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}, MemberList: RpcResp.MemberList}
	memberListResp.Data = jsonData.JsonDataList(memberListResp.MemberList)
	log.NewInfo(req.OperationID, "GetGroupAllMember api return ", memberListResp)
	c.JSON(http.StatusOK, memberListResp)
}

func GetJoinedGroupList(c *gin.Context) {
	params := api.GetJoinedGroupListReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetJoinedGroupListReq{}
	utils.CopyStructFields(req, params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetJoinedGroupList args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.GetJoinedGroupList(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetJoinedGroupList failed  ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	GroupListResp := api.GetJoinedGroupListResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}, GroupInfoList: RpcResp.GroupList}
	GroupListResp.Data = jsonData.JsonDataList(GroupListResp.GroupInfoList)
	log.NewInfo(req.OperationID, "GetJoinedGroupList api return ", GroupListResp)
	c.JSON(http.StatusOK, GroupListResp)
}

func InviteUserToGroup(c *gin.Context) {
	params := api.InviteUserToGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.InviteUserToGroupReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "InviteUserToGroup args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.InviteUserToGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "InviteUserToGroup failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.InviteUserToGroupResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}}
	for _, v := range RpcResp.Id2ResultList {
		resp.UserIDResultList = append(resp.UserIDResultList, &api.UserIDResult{UserID: v.UserID, Result: v.Result})
	}

	if len(resp.UserIDResultList) == 0 {
		resp.UserIDResultList = *new([]*api.UserIDResult)
	}

	log.NewInfo(req.OperationID, "InviteUserToGroup api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func CreateGroup(c *gin.Context) {

	token := c.Request.Header.Get("token")
	recentGroupID, _ := db.DB.GetTimeLimitOnActionType(db.GroupCreatedCacheType, token)
	if recentGroupID != "" {
		log.NewError("User have recently created group not more then 10 sec")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": "you have to wait for creating new group"})
		return
	}

	params := api.CreateGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	//
	req := &rpc.CreateGroupReq{GroupInfo: &open_im_sdk.GroupInfo{}}
	utils.CopyStructFields(req.GroupInfo, &params)

	for _, v := range params.MemberList {
		req.InitMemberList = append(req.InitMemberList, &rpc.GroupAddMemberInfo{UserID: v.UserID, RoleLevel: v.RoleLevel})
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	req.OwnerUserID = params.OwnerUserID
	req.OperationID = params.OperationID
	req.GroupInterest = params.GroupInterest
	req.IsOpen = params.IsOpen
	log.NewInfo(req.OperationID, "CreateGroup args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.CreateGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "CreateGroup failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}

	if RpcResp.ErrCode == constant.OK.ErrCode {
		if err := db.DB.SetTimeLimitOnActionType(db.GroupCreatedCacheType, token, RpcResp.GroupInfo.GroupID); err != nil && RpcResp.GroupInfo != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, Insertion new group ID failed", err.Error())
		}
	}
	resp := api.CreateGroupResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}}
	if RpcResp.ErrCode == 0 {
		utils.CopyStructFields(&resp.GroupInfo, RpcResp.GroupInfo)
		resp.Data = jsonData.JsonDataOne(&resp.GroupInfo)

		resp.Data["createTime"] = 0
		resp.Data["status"] = 0
		resp.Data["groupType"] = 0
	}
	log.NewInfo(req.OperationID, "CreateGroup api return ", resp)
	c.JSON(http.StatusOK, resp)
}

// 群主或管理员收到的
func GetRecvGroupApplicationList(c *gin.Context) {
	params := api.GetGroupApplicationListReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetGroupApplicationListReq{}
	utils.CopyStructFields(req, params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetGroupApplicationList args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	reply, err := client.GetGroupApplicationList(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupApplicationList failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.GetGroupApplicationListResp{CommResp: api.CommResp{ErrCode: reply.ErrCode, ErrMsg: reply.ErrMsg}, GroupRequestList: reply.GroupRequestList}
	resp.Data = jsonData.JsonDataList(resp.GroupRequestList)
	log.NewInfo(req.OperationID, "GetGroupApplicationList api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetUserReqGroupApplicationList(c *gin.Context) {
	var params api.GetUserReqGroupApplicationListReq
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetUserReqApplicationListReq{}
	utils.CopyStructFields(req, params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetGroupsInfo args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.GetUserReqApplicationList(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupsInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	log.NewInfo(req.OperationID, RpcResp)
	resp := api.GetGroupApplicationListResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}, GroupRequestList: RpcResp.GroupRequestList}
	log.NewInfo(req.OperationID, "GetGroupApplicationList api return ", resp)
	resp.Data = jsonData.JsonDataList(resp.GroupRequestList)
	c.JSON(http.StatusOK, resp)
}

func GetGroupsInfo(c *gin.Context) {
	params := api.GetGroupInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetGroupsInfoReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "GetGroupsInfo args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.GetGroupsInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetGroupsInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}

	resp := api.GetGroupInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}, GroupInfoList: RpcResp.GroupInfoList}
	resp.Data = jsonData.JsonDataList(resp.GroupInfoList)
	log.NewInfo(req.OperationID, "GetGroupsInfo api return ", resp)
	c.JSON(http.StatusOK, resp)
}

// process application
func ApplicationGroupResponse(c *gin.Context) {
	params := api.ApplicationGroupResponseReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.GroupApplicationResponseReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "ApplicationGroupResponse args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	reply, err := client.GroupApplicationResponse(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GroupApplicationResponse failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.ApplicationGroupResponseResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "ApplicationGroupResponse api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func JoinGroup(c *gin.Context) {
	params := api.JoinGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.JoinGroupReq{}
	utils.CopyStructFields(req, params)

	token := c.Request.Header.Get("token")
	recentJoinedGroupID, _ := db.DB.GetTimeLimitOnActionType(db.GroupJoinCacheType, token+params.GroupID)
	if recentJoinedGroupID != "" {
		log.NewError("User have recently joined group not more then 10 sec")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": "you have to wait for joining new group"})
		return
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "JoinGroup args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)

	RpcResp, err := client.JoinGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "JoinGroup failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	if RpcResp.CommonResp.ErrCode == constant.OK.ErrCode {
		if err := db.DB.SetTimeLimitOnActionType(db.GroupJoinCacheType, token+params.GroupID, params.GroupID); err != nil && RpcResp.CommonResp != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, Insertion new join group failed", err.Error())
		}
	}
	resp := api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}
	log.NewInfo(req.OperationID, "JoinGroup api return", RpcResp.String())
	c.JSON(http.StatusOK, resp)
}

func QuitGroup(c *gin.Context) {
	params := api.QuitGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.QuitGroupReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "QuitGroup args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.QuitGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "call quit group rpc server failed,err=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	resp := api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}
	log.NewInfo(req.OperationID, "QuitGroup api return", RpcResp.String())
	c.JSON(http.StatusOK, resp)
}

func SetGroupInfo(c *gin.Context) {
	params := api.SetGroupInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.SetGroupInfoReq{GroupInfo: &open_im_sdk.GroupInfo{}}
	utils.CopyStructFields(req.GroupInfo, &params)
	req.OperationID = params.OperationID
	req.IsAdmin = false

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "SetGroupInfo args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.SetGroupInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "SetGroupInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	resp := api.SetGroupInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}}
	c.JSON(http.StatusOK, resp)
	log.NewInfo(req.OperationID, "SetGroupInfo api return ", resp)
}

func TransferGroupOwner(c *gin.Context) {
	params := api.TransferGroupOwnerReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.TransferGroupOwnerReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "TransferGroupOwner args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	reply, err := client.TransferGroupOwner(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "TransferGroupOwner failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.TransferGroupOwnerResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "TransferGroupOwner api return ", resp)
	c.JSON(http.StatusOK, resp)
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
	c.JSON(http.StatusOK, resp)
}

func MuteGroupMember(c *gin.Context) {
	params := api.MuteGroupMemberReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.MuteGroupMemberReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc args ", req.String())
	opFlag, err := getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		errMsg := req.OperationID + " getGroupUserLevel failed " + req.GroupID + req.OpUserID + err.Error()
		log.Error(req.OperationID, errMsg)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": constant.ErrAccess.ErrMsg})
		return
	}
	if opFlag == 0 {
		errMsg := req.OperationID + "opFlag == 0  " + req.GroupID + req.OpUserID
		log.Error(req.OperationID, errMsg)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": constant.ErrAccess.ErrMsg})
		return
	}

	mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.UserID)
	if err != nil {
		errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID failed " + req.GroupID + req.UserID + err.Error()
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}
	if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
		errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupOwner " + req.GroupID + req.UserID
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}
	if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
		errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupAdmin " + req.GroupID + req.UserID
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
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
	reply, err := client.MuteGroupMember(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.MuteGroupMemberResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func getGroupUserLevel(groupID, userID string) (int, error) {
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

func CancelMuteGroupMember(c *gin.Context) {
	params := api.CancelMuteGroupMemberReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.CancelMuteGroupMemberReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api args ", req.String())

	opFlag, err := getGroupUserLevel(req.GroupID, req.OpUserID)
	if err != nil {
		errMsg := req.OperationID + " getGroupUserLevel failed " + req.GroupID + req.OpUserID + err.Error()
		log.Error(req.OperationID, errMsg)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}
	if opFlag == 0 {
		errMsg := req.OperationID + "opFlag == 0  " + req.GroupID + req.OpUserID
		log.Error(req.OperationID, errMsg)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}

	mutedInfo, err := imdb.GetGroupMemberInfoByGroupIDAndUserID(req.GroupID, req.UserID)
	if err != nil {
		errMsg := req.OperationID + " GetGroupMemberInfoByGroupIDAndUserID failed " + req.GroupID + req.UserID + err.Error()
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}
	if mutedInfo.RoleLevel == constant.GroupOwner && opFlag != 1 {
		errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupOwner " + req.GroupID + req.UserID
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}
	if mutedInfo.RoleLevel == constant.GroupAdmin && opFlag == 3 {
		errMsg := req.OperationID + " mutedInfo.RoleLevel == constant.GroupAdmin " + req.GroupID + req.UserID
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": errMsg})
		return
	}

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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.CancelMuteGroupMemberResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func MuteGroup(c *gin.Context) {
	params := api.MuteGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.MuteGroupReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
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
	reply, err := client.MuteGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.MuteGroupResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func CancelMuteGroup(c *gin.Context) {
	params := api.CancelMuteGroupReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.CancelMuteGroupReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
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
	reply, err := client.CancelMuteGroup(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.CancelMuteGroupResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	c.JSON(http.StatusOK, resp)
}

//SetGroupMemberNickname

func SetGroupMemberNickname(c *gin.Context) {
	params := api.SetGroupMemberNicknameReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.SetGroupMemberNicknameReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
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
	reply, err := client.SetGroupMemberNickname(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp := api.SetGroupMemberNicknameResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func SetGroupMemberInfo(c *gin.Context) {
	var (
		req  api.SetGroupMemberInfoReq
		resp api.SetGroupMemberInfoResp
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req)
	var opUserID string
	userIDInter, existed := c.Get("userID")
	if existed {
		opUserID = userIDInter.(string)
	}

	reqPb := &rpc.SetGroupMemberInfoReq{
		GroupID:     req.GroupID,
		UserID:      req.UserID,
		OperationID: req.OperationID,
		OpUserID:    opUserID,
	}
	if req.Nickname != nil {
		reqPb.Nickname = &wrappers.StringValue{Value: *req.Nickname}
	}
	if req.FaceURL != nil {
		reqPb.FaceURL = &wrappers.StringValue{Value: *req.FaceURL}
	}
	if req.Ex != nil {
		reqPb.Ex = &wrappers.StringValue{Value: *req.Ex}
	}
	if req.RoleLevel != nil {
		reqPb.RoleLevel = &wrappers.Int32Value{Value: *req.RoleLevel}
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	respPb, err := client.SetGroupMemberInfo(context.Background(), reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	resp.ErrMsg = respPb.CommonResp.ErrMsg
	resp.ErrCode = respPb.CommonResp.ErrCode
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " api args ", resp)
	c.JSON(http.StatusOK, resp)
}

func CheckGroupUpdateVersionsFromLocal(c *gin.Context) {
	var (
		req api.CheckGroupUpdateVersionsFromLocalReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req)
	var opUserID string
	userIDInter, existed := c.Get("userID")
	if existed {
		opUserID = userIDInter.(string)
	}

	reqPb := &rpc.CheckGroupUpdateVersionsFromLocalReq{
		GroupID:           req.GroupID,
		GroupVersion:      req.GroupVersion,
		OperationID:       req.OperationID,
		PageNumber:        req.PageNumber,
		PageSize:          req.PageSize,
		NeedNextPageFetch: req.NeedNextPageFetch,
		OpUserID:          opUserID,
		ResponseBackHTTP:  req.ResponseBackHTTP,
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	respPb, err := client.CheckGroupUpdateVersionsFromLocal(context.Background(), reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), " failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	customeResponse := make(map[string]interface{})
	customeResponse["errCode"] = 0
	customeResponse["errMsg"] = ""
	customeResponse["data"] = respPb //jsonData.JsonDataOne(respPb)
	c.JSON(http.StatusOK, customeResponse)
}
