package user

import (
	jsonData "Open_IM/internal/utils"
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	cacheRpc "Open_IM/pkg/proto/cache"
	pbRelay "Open_IM/pkg/proto/relay"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	pbUser "Open_IM/pkg/proto/user"
	rpc "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetUsersInfoFromCache(c *gin.Context) {
	params := api.GetUsersInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	log.NewInfo(params.OperationID, "GetUsersInfoFromCache req: ", params)
	req := &rpc.GetUserInfoReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.GetUserInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetUserInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	var publicUserInfoList []*open_im_sdk.PublicUserInfo
	for _, v := range RpcResp.UserInfoList {
		publicUserInfoList = append(publicUserInfoList,
			&open_im_sdk.PublicUserInfo{UserID: v.UserID, Nickname: v.Nickname, FaceURL: v.FaceURL, Gender: v.Gender, Ex: v.Ex})
	}
	resp := api.GetUsersInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}, UserInfoList: publicUserInfoList}
	resp.Data = jsonData.JsonDataList(resp.UserInfoList)
	log.NewInfo(req.OperationID, "GetUserInfo api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetFriendIDListFromCache(c *gin.Context) {
	var (
		req    api.GetFriendIDListFromCacheReq
		resp   api.GetFriendIDListFromCacheResp
		reqPb  cacheRpc.GetFriendIDListFromCacheReq
		respPb *cacheRpc.GetFriendIDListFromCacheResp
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req)
	reqPb.OperationID = req.OperationID

	userIDInter, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInter.(string)
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := cacheRpc.NewCacheClient(etcdConn)
	respPb, err := client.GetFriendIDListFromCache(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetFriendIDListFromCache", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed:" + err.Error()})
		return
	}
	resp.UserIDList = respPb.UserIDList
	resp.CommResp = api.CommResp{ErrMsg: respPb.CommonResp.ErrMsg, ErrCode: respPb.CommonResp.ErrCode}
	c.JSON(http.StatusOK, resp)
}

func GetBlackIDListFromCache(c *gin.Context) {
	var (
		req    api.GetBlackIDListFromCacheReq
		resp   api.GetBlackIDListFromCacheResp
		reqPb  cacheRpc.GetBlackIDListFromCacheReq
		respPb *cacheRpc.GetBlackIDListFromCacheResp
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)
	reqPb.OperationID = req.OperationID

	userIDInter, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInter.(string)
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := cacheRpc.NewCacheClient(etcdConn)
	respPb, err := client.GetBlackIDListFromCache(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetFriendIDListFromCache", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed:" + err.Error()})
		return
	}
	resp.UserIDList = respPb.UserIDList
	resp.CommResp = api.CommResp{ErrMsg: respPb.CommonResp.ErrMsg, ErrCode: respPb.CommonResp.ErrCode}
	c.JSON(http.StatusOK, resp)
}

func GetUsersInfo(c *gin.Context) {
	params := api.GetUsersInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetUserInfoReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(params.OperationID, "GetUserInfo args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.GetUserInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetUserInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}

	log.NewInfo(req.OperationID, "GetUserInfo rpc return ", RpcResp.String())

	var publicUserInfoList []*open_im_sdk.PublicUserInfo
	for _, v := range RpcResp.UserInfoList {
		publicUserInfoList = append(publicUserInfoList,
			&open_im_sdk.PublicUserInfo{UserID: v.UserID, Nickname: v.Nickname, FaceURL: v.FaceURL, Gender: v.Gender, Ex: v.Ex, MomentsPicArray: v.MomentsPicArray, IsFriend: v.IsFriend, MomentCount: v.MomentsCount})
	}

	resp := api.GetUsersInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}, UserInfoList: publicUserInfoList}
	resp.Data = jsonData.JsonDataList(resp.UserInfoList)
	log.NewInfo(req.OperationID, "GetUserInfo api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func SearchUser(c *gin.Context) {
	params := api.GetUsersInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	req := &rpc.SearchUserRequest{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(params.OperationID, "GetUserInfo args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.SearchUser(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetUserInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	var publicUserInfoList []*open_im_sdk.PublicUserInfo
	for _, v := range RpcResp.UserInfoList {
		publicUserInfoList = append(publicUserInfoList,
			&open_im_sdk.PublicUserInfo{UserID: v.UserID, Nickname: v.Nickname, FaceURL: v.FaceURL, Gender: v.Gender, Ex: v.Ex, MomentsPicArray: v.MomentsPicArray})
	}

	resp := api.GetUsersInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}, UserInfoList: publicUserInfoList}
	resp.Data = jsonData.JsonDataList(resp.UserInfoList)
	log.NewInfo(req.OperationID, "GetUserInfo api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func UpdateUserInfo(c *gin.Context) {
	params := api.UpdateSelfUserInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	// check separately cuz need custom error code for this to show in FE
	if len(params.Nickname) > 36 {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrNickNameLength.ErrCode, "errMsg": constant.ErrNickNameLength.ErrMsg})
		return
	}

	req := &rpc.UpdateUserInfoReq{UserInfo: &open_im_sdk.UserInfo{}}
	utils.CopyStructFields(req.UserInfo, &params)
	req.OperationID = params.OperationID

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	if req.UserInfo.Email != "" {
		if !strings.Contains(req.UserInfo.Email, "@") {
			log.NewError(params.OperationID, "The email address should contain @, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The email address should contain @, please check."})
			return
		}
	}
	if req.UserInfo.PhoneNumber != "" {
		if !strings.Contains(req.UserInfo.PhoneNumber, "+") {
			log.NewError(params.OperationID, "The phone number should has + at the head, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The phone number should has + at the head, please check."})
			return
		}
		oldUser, err := imdb.GetRegisterFromPhone(req.UserInfo.PhoneNumber)
		if err == nil && oldUser.UserID != params.ApiUserInfo.UserID {
			log.NewError(params.OperationID, "The phone number has been used by other user, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrUserExistPhoneNumberExist.ErrCode, "errMsg": constant.ErrUserExistPhoneNumberExist.ErrMsg})
			return
		}
	}

	log.NewInfo(params.OperationID, "UpdateUserInfo args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.UpdateUserInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "UpdateUserInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	resp := api.UpdateUserInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "UpdateUserInfo api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func RemoveUserFaceUrl(c *gin.Context) {
	params := api.UpdateSelfUserInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.UpdateUserInfoReq{UserInfo: &open_im_sdk.UserInfo{}}
	utils.CopyStructFields(req.UserInfo, &params)
	req.OperationID = params.OperationID

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(params.OperationID, "RemoveUserFaceUrl args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.RemoveUserFaceUrl(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "RemoveUserFaceUrl failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	resp := api.UpdateUserInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "RemoveUserFaceUrl api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func SetGlobalRecvMessageOpt(c *gin.Context) {
	params := api.SetGlobalRecvMessageOptReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &rpc.SetGlobalRecvMessageOptReq{}
	utils.CopyStructFields(req, &params)
	req.OperationID = params.OperationID

	userIDInter, existed := c.Get("userID")
	if existed {
		req.UserID = userIDInter.(string)
	}

	log.NewInfo(params.OperationID, "SetGlobalRecvMessageOpt args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.SetGlobalRecvMessageOpt(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "SetGlobalRecvMessageOpt failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	resp := api.UpdateUserInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "SetGlobalRecvMessageOpt api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetSelfUserInfo(c *gin.Context) {
	params := api.GetSelfUserInfoReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	req := &rpc.GetUserInfoReq{}

	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	req.UserIDList = append(req.UserIDList, req.OpUserID)
	log.NewInfo(params.OperationID, "GetUserInfo args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	RpcResp, err := client.GetUserInfo(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetUserInfo failed ", err.Error(), req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call  rpc server failed"})
		return
	}
	if len(RpcResp.UserInfoList) == 1 {
		userInfo := RpcResp.UserInfoList[0]
		//s, ok := userInfo.CreateTime.(string)
		//if ok {
		//	strconv.ParseInt(s)
		//}
		resp := api.GetSelfUserInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}, UserInfo: userInfo}
		resp.Data = jsonData.JsonDataOne(resp.UserInfo)
		//createTime := resp.Data["CreateTime"]
		//s, _ := createTime.(string)
		//resp.Data["CreateTime"], err = strconv.ParseInt(s, 10, 64)

		log.NewInfo(req.OperationID, "GetUserInfo api return ", resp)
		c.JSON(http.StatusOK, resp)
	} else {
		resp := api.GetSelfUserInfoResp{CommResp: api.CommResp{ErrCode: RpcResp.CommonResp.ErrCode, ErrMsg: RpcResp.CommonResp.ErrMsg}}
		log.NewInfo(req.OperationID, "GetUserInfo api return ", resp)
		c.JSON(http.StatusOK, resp)
	}

}

func GetUsersOnlineStatus(c *gin.Context) {
	params := api.GetUsersOnlineStatusReq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	req := &pbRelay.GetUsersOnlineStatusReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	if len(config.Config.Manager.AppManagerUid) == 0 {
		log.NewError(req.OperationID, "Manager == 0")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "Manager == 0"})
		return
	}
	req.OpUserID = config.Config.Manager.AppManagerUid[0]

	log.NewInfo(params.OperationID, "GetUsersOnlineStatus args ", req.String())
	var wsResult []*pbRelay.GetUsersOnlineStatusResp_SuccessResult
	var respResult []*pbRelay.GetUsersOnlineStatusResp_SuccessResult
	flag := false
	grpcCons := getcdv3.GetConn4Unique(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOnlineMessageRelayName)
	for _, v := range grpcCons {
		client := pbRelay.NewOnlineMessageRelayServiceClient(v)
		reply, err := client.GetUsersOnlineStatus(context.Background(), req)
		if err != nil {
			log.NewError(params.OperationID, "GetUsersOnlineStatus rpc  err", req.String(), err.Error())
			continue
		} else {
			if reply != nil {
				log.NewInfo(params.OperationID, utils.GetSelfFuncName(), "grpcCons:", v, reply.ErrCode, reply.ErrMsg, reply.SuccessResult)
				if reply.ErrCode == 0 {
					wsResult = append(wsResult, reply.SuccessResult...)
				}
			}

		}
	}
	log.NewInfo(params.OperationID, "call GetUsersOnlineStatus rpc server is success", wsResult)
	//Online data merge of each node
	for _, v1 := range params.UserIDList {
		flag = false
		temp := new(pbRelay.GetUsersOnlineStatusResp_SuccessResult)
		for _, v2 := range wsResult {
			if v2.UserID == v1 {
				flag = true
				temp.UserID = v1
				temp.Status = constant.OnlineStatus
				temp.DetailPlatformStatus = append(temp.DetailPlatformStatus, v2.DetailPlatformStatus...)
			}

		}
		if !flag {
			temp.UserID = v1
			temp.Status = constant.OfflineStatus
		}
		respResult = append(respResult, temp)
	}
	resp := api.GetUsersOnlineStatusResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, SuccessResult: respResult}
	if len(respResult) == 0 {
		resp.SuccessResult = []*pbRelay.GetUsersOnlineStatusResp_SuccessResult{}
	}
	log.NewInfo(req.OperationID, "GetUsersOnlineStatus api return", resp)
	c.JSON(http.StatusOK, resp)
}

func DeleteUser(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteSelfUserRequest
		reqPb pbUser.DeleteUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.UserId = userID
	reqPb.OpUserId = userID
	reqPb.Reason = req.Reason
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbUser.NewUserClient(etcdConn)
	_, err := client.DeleteUser(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, "DeleteUser rpc failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}
