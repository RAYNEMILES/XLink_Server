package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/oauth"
	rpc "Open_IM/pkg/proto/auth"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type FaceBookOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	AccessToken string `json:"access_token" Binding:"required"`
}

type GoogleOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	Code        string `json:"code" Binding:"required"`
	State       string `json:"state" Binding:"required"`
	IdToken     string `json:"idToken" Binding:"required"`
}

type AppleOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	Code        string `json:"access_token" Binding:"required"`
	RedirectURI string `json:"redirect_uri" Binding:"required"`
}

func FaceBookLogin(c *gin.Context) {
	var params FaceBookOauthParamsRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	thirdUserInfo, err := oauth.ParseFaceBookAccessToken(params.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	user := imdb.GetUserByTypeAndThirdId(constant.OauthTypeFaceBook, thirdUserInfo["id"].(string))

	// insert
	if user.ThirdId == "" {
		imdb.InsertThirdInfo(constant.OauthTypeFaceBook, thirdUserInfo["id"].(string), thirdUserInfo["name"].(string))
	}

	// Registered
	if user.UserId != "" {
		// update third user info
		imdb.UpdateThirdUserNameByThirdInfo(constant.OauthTypeFaceBook, thirdUserInfo["id"].(string), thirdUserInfo["name"].(string))

		userInfo, err := imdb.GetRegisterFromUserId(user.UserId)
		if err != nil {
			log.NewError("", "user have not register", err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}

		if userInfo.DeleteTime != 0 {
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register...."})
			return
		}

		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, params.OperationID)
		if etcdConn == nil {
			errMsg := params.OperationID + " etcd3.GetConn == nil"
			log.NewError(params.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}

		client := rpc.NewAuthClient(etcdConn)
		req := &rpc.UserTokenReq{Platform: params.Platform, FromUserID: user.UserId, OperationID: params.OperationID, GAuthTypeToken: false}
		reply, err := client.UserToken(context.Background(), req)
		if err != nil {
			errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
			log.NewError(req.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}

		imdb.UpdateUserLastLoginIP(user.UserId, c.ClientIP())

		resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg},
			UserToken: api.UserTokenInfo{
				UserID:      req.FromUserID,
				Token:       reply.Token,
				ExpiredTime: reply.ExpiredTime,
				UserSign:    reply.UserSign,
				Extend: api.UserTokenInfoExtend{
					IsNeedBindPhone: userInfo.PhoneNumber == "",
					IsChangePwd:     userInfo.Password == "",
					IsChangeName:    userInfo.Nickname == "" || userInfo.Nickname == userInfo.UserID,
					IsChangeFace:    userInfo.FaceURL == "",
				}}}
		c.JSON(http.StatusOK, resp)
		return
	}

	// create user
	userID := utils.Md5(params.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
	bi := big.NewInt(0)
	bi.SetString(userID[0:8], 16)
	userID = bi.String()

	password := ""
	// register
	var domain string
	var domainPre = "http://"
	if config.Config.Environment != constant.DEV {
		domain = c.Request.Host
		domainPre = "https://"
	}

	if domain == "" {
		domain = fmt.Sprintf("%s:%d", utils.ServerIP, config.Config.Api.GinPort[0])
	} else {
		domain = fmt.Sprintf("%s/api", domain)
	}
	url := fmt.Sprintf("%s%s/auth/user_register", domainPre, domain)
	openIMRegisterReq := api.UserRegisterReq{}
	openIMRegisterReq.OperationID = params.OperationID
	openIMRegisterReq.Platform = params.Platform
	openIMRegisterReq.UserID = userID
	openIMRegisterReq.Nickname = userID
	openIMRegisterReq.Secret = password
	openIMRegisterReq.FaceURL = ""
	openIMRegisterReq.PhoneNumber = ""
	openIMRegisterReq.Email = ""
	openIMRegisterReq.Ex = ""
	openIMRegisterReq.SourceCode = ""
	openIMRegisterReq.CreateIp = c.ClientIP()
	openIMRegisterReq.UpdateIp = c.ClientIP()
	openIMRegisterReq.Uuid = ""
	openIMRegisterResp := api.UserRegisterResp{}
	bMsg, err := http2.Post(url, openIMRegisterReq, 2)
	if err != nil {
		log.NewError(params.OperationID, "request openIM register error", userID, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMRegisterResp)
	if err != nil || openIMRegisterResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request openIM register error", userID, "err", "resp: ", openIMRegisterResp.ErrCode)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed: " + openIMRegisterResp.ErrMsg})
		return
	}

	// binding
	imdb.UpdateUserIdByThirdInfo(constant.OauthTypeFaceBook, thirdUserInfo["id"].(string), userID)

	resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: openIMRegisterResp.ErrCode, ErrMsg: openIMRegisterResp.ErrMsg},
		UserToken: api.UserTokenInfo{
			UserID:      userID,
			Token:       openIMRegisterResp.UserToken.Token,
			ExpiredTime: openIMRegisterResp.UserToken.ExpiredTime,
			UserSign:    openIMRegisterResp.UserToken.UserSign,
			Extend:      api.UserTokenInfoExtend{IsNeedBindPhone: true, IsChangePwd: true, IsChangeName: true, IsChangeFace: true}}}
	c.JSON(http.StatusOK, resp)
	return
}

func GoogleLogin(c *gin.Context) {
	var params GoogleOauthParamsRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	if params.State != "random" {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": "state is not random"})
		return
	}

	thirdUserInfo, err := oauth.ParseGoogleCode(int(params.Platform), params.Code, params.IdToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	user := imdb.GetUserByTypeAndThirdId(constant.OauthTypeGoogle, thirdUserInfo["sub"].(string))

	// insert
	if user.ThirdId == "" {
		imdb.InsertThirdInfo(constant.OauthTypeGoogle, thirdUserInfo["sub"].(string), thirdUserInfo["email"].(string))
	}

	// Registered
	if user.UserId != "" {
		// update third user info
		imdb.UpdateThirdUserNameByThirdInfo(constant.OauthTypeGoogle, thirdUserInfo["sub"].(string), thirdUserInfo["email"].(string))

		userInfo, err := imdb.GetRegisterFromUserId(user.UserId)
		if err != nil {
			log.NewError("", "user have not register", err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}

		if userInfo.DeleteTime != 0 {
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register...."})
			return
		}

		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, params.OperationID)
		if etcdConn == nil {
			errMsg := params.OperationID + " etcd3.GetConn == nil"
			log.NewError(params.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}

		client := rpc.NewAuthClient(etcdConn)
		req := &rpc.UserTokenReq{Platform: params.Platform, FromUserID: user.UserId, OperationID: params.OperationID, GAuthTypeToken: false}
		reply, err := client.UserToken(context.Background(), req)
		if err != nil {
			errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
			log.NewError(req.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}

		imdb.UpdateUserLastLoginIP(user.UserId, c.ClientIP())

		resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg},
			UserToken: api.UserTokenInfo{
				UserID:      req.FromUserID,
				Token:       reply.Token,
				ExpiredTime: reply.ExpiredTime,
				UserSign:    reply.UserSign,
				Extend: api.UserTokenInfoExtend{
					IsNeedBindPhone: userInfo.PhoneNumber == "",
					IsChangePwd:     userInfo.Password == "",
					IsChangeName:    userInfo.Nickname == "" || userInfo.Nickname == userInfo.UserID,
					IsChangeFace:    userInfo.FaceURL == "",
				}}}
		c.JSON(http.StatusOK, resp)
		return
	}

	// create user
	userID := utils.Md5(params.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
	bi := big.NewInt(0)
	bi.SetString(userID[0:8], 16)
	userID = bi.String()

	password := ""
	// register
	var domain string
	var domainPre = "http://"
	if config.Config.Environment != constant.DEV {
		domain = c.Request.Host
		domainPre = "https://"
	}

	if domain == "" {
		domain = fmt.Sprintf("%s:%d", utils.ServerIP, config.Config.Api.GinPort[0])
	} else {
		domain = fmt.Sprintf("%s/api", domain)
	}
	url := fmt.Sprintf("%s%s/auth/user_register", domainPre, domain)
	openIMRegisterReq := api.UserRegisterReq{}
	openIMRegisterReq.OperationID = params.OperationID
	openIMRegisterReq.Platform = params.Platform
	openIMRegisterReq.UserID = userID
	openIMRegisterReq.Nickname = userID
	openIMRegisterReq.Secret = password
	openIMRegisterReq.FaceURL = ""
	openIMRegisterReq.PhoneNumber = ""
	openIMRegisterReq.Email = ""
	openIMRegisterReq.Ex = ""
	openIMRegisterReq.SourceCode = ""
	openIMRegisterReq.CreateIp = c.ClientIP()
	openIMRegisterReq.UpdateIp = c.ClientIP()
	openIMRegisterReq.Uuid = ""
	openIMRegisterResp := api.UserRegisterResp{}
	bMsg, err := http2.Post(url, openIMRegisterReq, 2)
	if err != nil {
		log.NewError(params.OperationID, "request openIM register error", userID, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMRegisterResp)
	if err != nil || openIMRegisterResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request openIM register error", userID, "err", "resp: ", openIMRegisterResp.ErrCode)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed: " + openIMRegisterResp.ErrMsg})
		return
	}

	// binding
	imdb.UpdateUserIdByThirdInfo(constant.OauthTypeGoogle, thirdUserInfo["sub"].(string), userID)

	resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: openIMRegisterResp.ErrCode, ErrMsg: openIMRegisterResp.ErrMsg},
		UserToken: api.UserTokenInfo{
			UserID:      userID,
			Token:       openIMRegisterResp.UserToken.Token,
			ExpiredTime: openIMRegisterResp.UserToken.ExpiredTime,
			UserSign:    openIMRegisterResp.UserToken.UserSign,
			Extend:      api.UserTokenInfoExtend{IsNeedBindPhone: true, IsChangePwd: true, IsChangeName: true, IsChangeFace: true}}}
	c.JSON(http.StatusOK, resp)
	return
}

func AppleLogin(c *gin.Context) {
	var params AppleOauthParamsRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	thirdUserInfo, err := oauth.ParseAppleCode(int(params.Platform), params.Code, params.RedirectURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	user := imdb.GetUserByTypeAndThirdId(constant.OauthTypeApple, thirdUserInfo["id"].(string))

	// insert
	if user.ThirdId == "" {
		imdb.InsertThirdInfo(constant.OauthTypeApple, thirdUserInfo["id"].(string), thirdUserInfo["email"].(string))
	}

	// Registered
	if user.UserId != "" {
		// update third user info
		imdb.UpdateThirdUserNameByThirdInfo(constant.OauthTypeApple, thirdUserInfo["id"].(string), thirdUserInfo["email"].(string))

		userInfo, err := imdb.GetRegisterFromUserId(user.UserId)
		if err != nil {
			log.NewError("", "user have not register", err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}

		if userInfo.DeleteTime != 0 {
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register...."})
			return
		}

		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, params.OperationID)
		if etcdConn == nil {
			errMsg := params.OperationID + " etcd3.GetConn == nil"
			log.NewError(params.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}

		client := rpc.NewAuthClient(etcdConn)
		req := &rpc.UserTokenReq{Platform: params.Platform, FromUserID: user.UserId, OperationID: params.OperationID, GAuthTypeToken: false}
		reply, err := client.UserToken(context.Background(), req)
		if err != nil {
			errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
			log.NewError(req.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}

		imdb.UpdateUserLastLoginIP(user.UserId, c.ClientIP())

		resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg},
			UserToken: api.UserTokenInfo{
				UserID:      req.FromUserID,
				Token:       reply.Token,
				ExpiredTime: reply.ExpiredTime,
				UserSign:    reply.UserSign,
				Extend: api.UserTokenInfoExtend{
					IsNeedBindPhone: userInfo.PhoneNumber == "",
					IsChangePwd:     userInfo.Password == "",
					IsChangeName:    userInfo.Nickname == "" || userInfo.Nickname == userInfo.UserID,
					IsChangeFace:    userInfo.FaceURL == "",
				}}}
		c.JSON(http.StatusOK, resp)
		return
	}

	// create user
	userID := utils.Md5(params.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
	bi := big.NewInt(0)
	bi.SetString(userID[0:8], 16)
	userID = bi.String()

	password := ""
	// register
	var domain string
	var domainPre = "http://"
	if config.Config.Environment != constant.DEV {
		domain = c.Request.Host
		domainPre = "https://"
	}

	if domain == "" {
		domain = fmt.Sprintf("%s:%d", utils.ServerIP, config.Config.Api.GinPort[0])
	} else {
		domain = fmt.Sprintf("%s/api", domain)
	}
	url := fmt.Sprintf("%s%s/auth/user_register", domainPre, domain)
	openIMRegisterReq := api.UserRegisterReq{}
	openIMRegisterReq.OperationID = params.OperationID
	openIMRegisterReq.Platform = params.Platform
	openIMRegisterReq.UserID = userID
	openIMRegisterReq.Nickname = userID
	openIMRegisterReq.Secret = password
	openIMRegisterReq.FaceURL = ""
	openIMRegisterReq.PhoneNumber = ""
	openIMRegisterReq.Email = ""
	openIMRegisterReq.Ex = ""
	openIMRegisterReq.SourceCode = ""
	openIMRegisterReq.CreateIp = c.ClientIP()
	openIMRegisterReq.UpdateIp = c.ClientIP()
	openIMRegisterReq.Uuid = ""
	openIMRegisterResp := api.UserRegisterResp{}
	bMsg, err := http2.Post(url, openIMRegisterReq, 2)
	if err != nil {
		log.NewError(params.OperationID, "request openIM register error", userID, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMRegisterResp)
	if err != nil || openIMRegisterResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request openIM register error", userID, "err", "resp: ", openIMRegisterResp.ErrCode)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed: " + openIMRegisterResp.ErrMsg})
		return
	}

	// binding
	imdb.UpdateUserIdByThirdInfo(constant.OauthTypeApple, thirdUserInfo["id"].(string), userID)

	resp := api.UserTokenResp{CommResp: api.CommResp{ErrCode: openIMRegisterResp.ErrCode, ErrMsg: openIMRegisterResp.ErrMsg},
		UserToken: api.UserTokenInfo{
			UserID:      userID,
			Token:       openIMRegisterResp.UserToken.Token,
			UserSign:    openIMRegisterResp.UserToken.UserSign,
			ExpiredTime: openIMRegisterResp.UserToken.ExpiredTime,
			Extend:      api.UserTokenInfoExtend{IsNeedBindPhone: true, IsChangePwd: true, IsChangeName: true, IsChangeFace: true}}}
	c.JSON(http.StatusOK, resp)
	return
}
