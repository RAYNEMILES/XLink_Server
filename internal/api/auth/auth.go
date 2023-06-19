package apiAuth

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/auth"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func UserRegister(c *gin.Context) {
	params := api.UserRegisterReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	//if params.Secret != config.Config.Secret {
	//	errMsg := " params.Secret != config.Config.Secret "
	//	log.NewError(params.OperationID, errMsg, params.Secret, config.Config.Secret)
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": errMsg})
	//	return
	//}

	req := &rpc.UserRegisterReq{UserInfo: &open_im_sdk.UserInfo{}}
	utils.CopyStructFields(req.UserInfo, &params)
	//copier.Copy(req.UserInfo, &params)
	req.OperationID = params.OperationID
	req.Password = params.Secret
	req.UserInfo.Uuid = params.Uuid
	log.NewInfo(req.OperationID, "UserRegister args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + " getcdv3.GetConn == nil"
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewAuthClient(etcdConn)
	reply, err := client.UserRegister(context.Background(), req)
	if err != nil {
		errMsg := req.OperationID + " " + "UserRegister failed . " + err.Error() + " ."
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg+req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	if reply.CommonResp.ErrCode != 0 {
		errMsg := req.OperationID + " " + " UserRegister failed . " + reply.CommonResp.ErrMsg + " .."
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg+req.String())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": reply.CommonResp.ErrMsg})
		return
	}

	pbDataToken := &rpc.UserTokenReq{Platform: params.Platform, FromUserID: params.UserID, OperationID: params.OperationID}
	replyToken, err := client.UserToken(context.Background(), pbDataToken)
	if err != nil {
		errMsg := req.OperationID + " " + " client.UserToken failed " + err.Error() + pbDataToken.String()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": pbDataToken.String()})
		return
	}
	resp := api.UserRegisterResp{CommResp: api.CommResp{ErrCode: replyToken.CommonResp.ErrCode, ErrMsg: replyToken.CommonResp.ErrMsg},
		UserToken: api.UserTokenInfo{UserID: req.UserInfo.UserID, Token: replyToken.Token, ExpiredTime: replyToken.ExpiredTime, UserSign: replyToken.UserSign}}

	IsChangeName := false
	if params.Nickname == "" || params.Nickname == params.UserID {
		IsChangeName = true
	}

	resp.UserToken.Extend = api.UserTokenInfoExtend{
		IsChangeName:    IsChangeName,
		IsChangePwd:     params.Secret == "",
		IsNeedBindPhone: params.PhoneNumber == "",
		IsChangeFace:    params.FaceURL == "",
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "UserRegister return ", resp)
	c.JSON(http.StatusOK, resp)

}

func UserToken(c *gin.Context) {
	params := api.UserTokenReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	//check user status
	if err := utils2.CheckUserPermissions(params.UserID); err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "user is banned!", params.UserID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	req := &rpc.UserTokenReq{Platform: params.Platform, FromUserID: params.UserID, OperationID: params.OperationID, GAuthTypeToken: params.GAuthTypeToken}
	log.NewInfo(req.OperationID, "UserToken args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + " getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewAuthClient(etcdConn)
	reply, err := client.UserToken(context.Background(), req)
	if err != nil {
		errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
		//log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	if reply.Token != "" {
		imdb.UpdateUserLastLoginIP(params.UserID, params.LoginIp)
		imdb.UpdateUserLastLoginDevice(params.UserID, params.Platform)
	}

	resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg},
		UserToken: api.UserTokenInfo{UserID: req.FromUserID, Token: reply.Token, ExpiredTime: reply.ExpiredTime, UserSign: reply.UserSign}}
	log.NewInfo(req.OperationID, "UserToken return ", resp)
	c.JSON(http.StatusOK, resp)
}

func ParseToken(c *gin.Context) {
	params := api.ParseTokenReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	token := c.Request.Header.Get("token")
	//_, err, errMsg := token_verify.WsVerifyToken(token, params.UserID, strconv.Itoa(int(params.PlatformID)), params.OperationID)
	//if err != nil {
	//	log.NewError(params.OperationID, errMsg)
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTokenInvalid, "errMsg": errMsg})
	//	return
	//}

	var ok bool
	var errInfo string
	var expireTime int64
	ok, _, errInfo, expireTime = token_verify.GetUserIDFromTokenExpireTime(token, params.OperationID)
	if !ok {
		errMsg := params.OperationID + " " + "GetUserIDFromTokenExpireTime failed " + errInfo + " token:" + c.Request.Header.Get("token")
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	resp := api.ParseTokenResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, ExpireTime: api.ExpireTime{ExpireTimeSeconds: uint32(expireTime)}}
	resp.Data = structs.Map(&resp.ExpireTime)
	log.NewInfo(params.OperationID, "ParseToken return ", resp)
	c.JSON(http.StatusOK, resp)
}

func ForceLogout(c *gin.Context) {
	params := api.ForceLogoutReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	req := &rpc.ForceLogoutReq{}
	utils.CopyStructFields(req, &params)

	userIDInter, existed := c.Get("userID")
	if existed {
		req.OpUserID = userIDInter.(string)
	}

	log.NewInfo(req.OperationID, "ForceLogout args ", req.String())
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + " getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewAuthClient(etcdConn)
	reply, err := client.ForceLogout(context.Background(), req)
	if err != nil {
		errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	resp := api.ForceLogoutResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewInfo(params.OperationID, utils.GetSelfFuncName(), " return ", resp)
	c.JSON(http.StatusOK, resp)
}

func UpdateUserIpLocation(c *gin.Context) {

	params := api.UpdateUserIPandStatusReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	req := &rpc.UpdateUserIPReq{}

	userIDInter, existed := c.Get("userID")
	if existed {
		req.UserID = userIDInter.(string)
	}
	ip := c.ClientIP()
	req.IPaddress = ip
	if params.IPaddress != "" {
		req.IPaddress = params.IPaddress
	}

	if err := utils2.CheckUserPermissions(req.UserID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "user is banned!", req.UserID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + " getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewAuthClient(etcdConn)
	reply, err := client.UpdateUserIPandStatus(context.Background(), req)
	if err != nil {
		errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	resp := api.UpdateUserIPandStatus{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	// log.NewInfo(params.OperationID, utils.GetSelfFuncName(), " return ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetUserIpLocation(c *gin.Context) {
	params := api.GetUserIPandStatusReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	req := &rpc.GetUserIPReq{}
	req.ForUserID = params.ForUserID
	req.FromUserID = params.UserID
	req.OperationID = params.OperationID

	userIDInter, existed := c.Get("userID")
	if existed {
		req.FromUserID = userIDInter.(string)
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + " getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewAuthClient(etcdConn)
	reply, err := client.GetUserIPandStatus(context.Background(), req)
	if err != nil {
		errMsg := err.Error()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusUnauthorized, gin.H{"errCode": 401, "errMsg": errMsg})
		return
	}
	log.NewError(req.OperationID, "Before Response parsed")
	resp := api.GetUserIPandStatusResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg}}
	log.NewError(req.OperationID, "After Response parsed 1")
	resp.Data.IPaddress = reply.IPaddress
	resp.Data.City = reply.City
	resp.Data.LastOnlineTime = reply.LastOnlineTime
	resp.Data.OnlineDifValue = reply.OnlineDifValue
	resp.Data.OperationID = reply.OperationID
	resp.Data.IsOnline = reply.IsOnline
	resp.Data.UserID = reply.UserID
	c.JSON(http.StatusOK, resp)
}

func ChangePassword(c *gin.Context) {
	var (
		apiRequest  = api.ChangePasswordRequest{}
		apiResponse = api.ChangePasswordResponse{}
		rpcRequest  = rpc.ChangePasswordRequest{}
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	var userId string
	userIDInter, existed := c.Get("userID")
	if existed {
		userId = userIDInter.(string)
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := apiRequest.OperationID + " getcdv3.GetConn == nil"
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	utils.CopyStructFields(&rpcRequest, &apiRequest)
	rpcRequest.UserID = userId
	client := rpc.NewAuthClient(etcdConn)
	reply, err := client.ChangePassword(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := apiRequest.OperationID + " UserToken failed " + err.Error() + rpcRequest.String()
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	apiResponse.ErrCode = reply.CommonResp.ErrCode
	apiResponse.ErrMsg = reply.CommonResp.ErrMsg
	c.JSON(http.StatusOK, apiResponse)
	return
}
