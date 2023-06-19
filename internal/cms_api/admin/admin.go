package admin

import (
	apiStruct "Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"net/http"
	url2 "net/url"
	"sort"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	gotp "github.com/diebietse/gotp/v2"

	"github.com/gin-gonic/gin"
)

var (
	minioClient *minio.Client
)

func init() {
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "minio config: ", config.Config.Credential.Minio)
	var initUrl string
	if config.Config.Credential.Minio.EndpointInnerEnable {
		initUrl = config.Config.Credential.Minio.EndpointInner
	} else {
		initUrl = config.Config.Credential.Minio.Endpoint
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "use initUrl: ", initUrl)
	minioUrl, err := url2.Parse(initUrl)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "parse failed, please check config/config.yaml", err.Error())
		return
	}
	opts := &minio.Options{
		Creds: credentials.NewStaticV4(config.Config.Credential.Minio.AccessKeyID, config.Config.Credential.Minio.SecretAccessKey, ""),
	}
	if minioUrl.Scheme == "http" {
		opts.Secure = false
	} else if minioUrl.Scheme == "https" {
		opts.Secure = true
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), "Parse ok ", config.Config.Credential.Minio)
	minioClient, err = minio.New(minioUrl.Host, opts)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "new ok ", config.Config.Credential.Minio)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "init minio client failed", err.Error())
		return
	}

	adminRolesAndPermissions := im_mysql_model.GetAdminAllRolesPermissions()
	db.DB.StoreAllAdminRolesInRedis(adminRolesAndPermissions)
	log.NewDebug("Sandman InIt Roles called", utils.GetSelfFuncName())
}

// Deprecated
func AdminLogin(c *gin.Context) {
	var (
		req   apiStruct.AdminLoginRequest
		resp  apiStruct.AdminLoginResponse
		reqPb pbAdmin.AdminLoginReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.Secret = req.Secret
	reqPb.AdminID = req.AdminName
	reqPb.OperationID = utils.OperationIDGenerator()
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.AdminLogin(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.Token = respPb.Token
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AdminLogin_v2(c *gin.Context) {

	var (
		req   apiStruct.AdminLoginRequest
		resp  apiStruct.AdminLoginResponse
		reqPb pbAdmin.AdminLoginReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.Secret = req.Secret
	reqPb.AdminID = req.AdminName
	reqPb.OperationID = utils.OperationIDGenerator()
	reqPb.GAuthTypeToken = true
	reqPb.RequestIP = c.ClientIP()
	log.NewError("Login Request Admin IP Address ", reqPb.RequestIP)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.AdminLoginV2(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc failed", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.Token = respPb.Token
	resp.GAuthEnabled = respPb.GAuthEnabled
	resp.GAuthSetupRequired = respPb.GAuthSetupRequired
	resp.GAuthSetupProvUri = respPb.GAuthSetupProvUri
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

// VerifyTOTPAdminUser verify TOTP after setup or after Login
func VerifyTOTPAdminUser(c *gin.Context) {
	var (
		resp  apiStruct.AdminLoginResponse
		reqPb pbAdmin.AdminLoginReq
	)
	params := apiStruct.ParamsTOTPVerify{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var ok bool
	var errInfo string
	ok, userID, errInfo := token_verify.GetAdminUserIDFromToken(c.Request.Header.Get("token"), params.OperationID, true)
	if !ok {
		errMsg := params.OperationID + " " + "GetUserIDFromToken failed " + errInfo + " token:" + c.Request.Header.Get("token")
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}

	if userID != "" {
		var user *db.AdminUser
		var err error
		userID := userID
		user, err = im_mysql_model.GetRegAdminUsrByUID(userID)

		if err != nil {
			log.NewError(params.OperationID, "Admin user have not register", userID, err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}
		totp := genrateTOTPForNow(*user)

		if params.TOTP == totp {
			reqPb.Secret = user.Password
			reqPb.SecretHashd = true
			reqPb.AdminID = user.UserID
			reqPb.OperationID = utils.OperationIDGenerator()
			reqPb.GAuthTypeToken = false
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
			if etcdConn == nil {
				errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
				log.NewError(reqPb.OperationID, errMsg)
				c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
				return
			}

			reqPb.RequestIP = c.ClientIP()
			log.NewError("Login Request Admin IP Address ", reqPb.RequestIP)
			client := pbAdmin.NewAdminCMSClient(etcdConn)
			respPb, err := client.AdminLoginV2(context.Background(), &reqPb)
			if err != nil {
				log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc failed", err.Error())
				openIMHttp.RespHttp200(c, err, nil)
				return
			}

			resp.Token = respPb.Token
			openIMHttp.RespHttp200(c, constant.OK, resp)
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": constant.PasswordErr, "err_msg": "TOTP is not correct", "data": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": constant.ErrTokenInvalid, "err_msg": constant.TokenInvalidMsg, "data": nil})
}

func AddAdminUser(c *gin.Context) {
	var (
		req   apiStruct.AddAdminUserRequest
		reqPb pbAdmin.AddAdminUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}

	reqPb.OpUserId = userID
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AddAdminUser(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeleteAdminUser(c *gin.Context) {
	var (
		req   apiStruct.DeleteAdminUserRequest
		reqPb pbAdmin.DeleteAdminUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
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
	reqPb.UserId = req.UserId
	reqPb.OpUserId = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteAdminUser(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, "DeleteUser rpc failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterAdminUser(c *gin.Context) {
	var (
		req   apiStruct.AlterAdminUserRequest
		reqPb pbAdmin.AlterAdminUserRequest
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}

	reqPb.OpUserId = userID
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AlterAdminUser(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetAdminUsers(c *gin.Context) {
	var (
		req   apiStruct.GetAdminUsersRequest
		resp  apiStruct.GetAdminUsersResponse
		reqPb pbAdmin.GetAdminUsersReq
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb.Pagination, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetAdminUsers(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.Users, respPb.User)
	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
	resp.CurrentPage = int(respPb.Pagination.CurrentPage)
	resp.UserNums = int64(respPb.UserNums)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func SearchAdminUsers(c *gin.Context) {
	var (
		req   apiStruct.SearchAdminUsersRequest
		reqPb pbAdmin.SearchAdminUsersRequest
		resp  apiStruct.SearchAdminUsersResponse
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	utils.CopyStructFields(&reqPb, req)
	operationID := utils.OperationIDGenerator()
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, operationID)
	if etcdConn == nil {
		errMsg := operationID + "getcdv3.GetConn == nil"
		log.NewError(operationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.SearchAdminUsers(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	resp.PageNumber = reqPb.PageNumber
	resp.PageSizeLimit = reqPb.PageLimit
	resp.TotalRecCount = int64(respPb.UserNums)

	utils.CopyStructFields(&resp.Users, respPb.User)
	log.NewInfo(operationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func SwitchAdminUserStatus(c *gin.Context) {
	var (
		req   apiStruct.SwitchStatusRequest
		reqPb pbAdmin.SwitchAdminUserStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()

	//get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	//check the parameters
	if req.UserId == "" || req.Status < 1 || req.Status > 2 {
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}

	reqPb.OpUserId = userID
	reqPb.UserId = req.UserId
	reqPb.Status = int32(req.Status)
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.SwitchAdminUserStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

// Change password with verify TOTP
func ChangePasswordTOTPAdminUser(c *gin.Context) {
	var (
		resp  apiStruct.AdminPasswordChangeResponse
		reqPb pbAdmin.ChangeAdminUserPasswordReq
	)
	params := apiStruct.AdminPasswordChangeRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	OperationID := utils.OperationIDGenerator()
	//get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	if userID != "" {
		reqPb.Secret = params.Secret
		reqPb.NewSecret = params.NewSecret
		reqPb.TOTP = params.TOTP
		reqPb.UserId = userID
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
		if etcdConn == nil {
			errMsg := OperationID + "getcdv3.GetConn == nil"
			log.NewError(OperationID, errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
			return
		}
		client := pbAdmin.NewAdminCMSClient(etcdConn)
		respPb, err := client.ChangeAdminUserPassword(context.Background(), &reqPb)
		if err != nil {
			log.NewError(OperationID, utils.GetSelfFuncName(), "rpc failed", err.Error())
			openIMHttp.RespHttp200(c, err, nil)
			return
		}
		resp.Token = respPb.Token
		resp.PasswordUpdated = respPb.PasswordUpdated
		openIMHttp.RespHttp200(c, constant.OK, resp)
		return

	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrTokenInvalid, "errMsg": constant.TokenInvalidMsg, "data": nil})
}

func AlterGAuthStatus(c *gin.Context) {
	var (
		req   apiStruct.AlterGAuthStatusRequest
		reqPb pbAdmin.AlterGAuthStatusReq
		resp  apiStruct.AlterGAuthStatusResponse
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	OperationID := utils.OperationIDGenerator()
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"errCode": 403, "errMsg": "UserID not exsisted in token"})
		return
	}
	reqPb.UserId = userID
	reqPb.UserGAuthStatus = req.Status

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, OperationID)
	if etcdConn == nil {
		errMsg := "getcdv3.GetConn == nil"
		log.NewError(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.AlterGAuthStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.UsergAuthStatus = respPb.GAuthStatus
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetgAuthQrCode(c *gin.Context) {
	var (
		reqPb pbAdmin.GetgAuthQrCodeReq
		resp  apiStruct.GetgAuthQrCodeResponse
	)
	OperationID := utils.OperationIDGenerator()
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"errCode": 403, "errMsg": "UserID not exsisted in token"})
		return
	}
	reqPb.UserId = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, OperationID)
	if etcdConn == nil {
		errMsg := "getcdv3.GetConn == nil"
		log.NewError(errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetgAuthQrCode(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.GAuthSetupProvUri = respPb.GAuthSetupProvUri
	resp.UsergAuthStatus = respPb.UsergAuthStatus
	resp.GAuthAccountID = respPb.GAuthAccountID
	resp.GAuthKey = respPb.GAuthKey
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

// GetPermissionsOfAdminUser current login user permissions only
func GetPermissionsOfAdminUser(c *gin.Context) {
	var (
		reqPb pbAdmin.AdminPermissionsReq
	)

	OperationID := utils.OperationIDGenerator()
	//get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	if userID != "" {
		reqPb.UserId = userID
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, OperationID)
		if etcdConn == nil {
			errMsg := OperationID + "getcdv3.GetConn == nil"
			log.NewError(OperationID, errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
			return
		}
		client := pbAdmin.NewAdminCMSClient(etcdConn)
		respPb, err := client.GetAdminPermissionReq(context.Background(), &reqPb)
		if err != nil {
			log.NewError(OperationID, utils.GetSelfFuncName(), "rpc failed", err.Error())
			openIMHttp.RespHttp200(c, err, nil)
			return
		}
		//traverse
		response := triversPagesByParents(respPb)
		openIMHttp.RespHttp200(c, constant.OK, response)
		return

	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrTokenInvalid, "errMsg": constant.TokenInvalidMsg, "data": nil})
}

// GetPermissionsOfAdminUserByID user permissions only
func GetPermissionsOfAdminUserByID(c *gin.Context) {
	var (
		req   apiStruct.AdminPermissionByAdminIDReq
		reqPb pbAdmin.AdminPermissionsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	OperationID := utils.OperationIDGenerator()
	//get the userId from middleware
	userID := req.AdminUserID
	if userID != "" {
		reqPb.UserId = userID
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, OperationID)
		if etcdConn == nil {
			errMsg := OperationID + "getcdv3.GetConn == nil"
			log.NewError(OperationID, errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
			return
		}
		client := pbAdmin.NewAdminCMSClient(etcdConn)
		respPb, err := client.GetAdminPermissionReq(context.Background(), &reqPb)
		if err != nil {
			log.NewError(OperationID, utils.GetSelfFuncName(), "rpc failed", err.Error())
			openIMHttp.RespHttp200(c, err, nil)
			return
		}
		//traverse
		response := triversPagesByParents(respPb)
		openIMHttp.RespHttp200(c, constant.OK, response)
		return

	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrTokenInvalid, "errMsg": constant.TokenInvalidMsg, "data": nil})
}

// Admin Role API
func AddAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AddAdminRoleRequest
		reqPb pbAdmin.AddAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)

	reqPb.AdminRoleName = req.AdminRoleName
	reqPb.Status = req.AdminRoleStatus
	reqPb.AdminAPIsIDs = req.AdminAPIsIDs
	reqPb.AdminPagesIDs = req.AdminPagesIDs
	reqPb.OperationID = req.OperationID
	reqPb.AdminRoleDiscription = req.AdminRoleDiscription
	reqPb.CreateUser = userID
	reqPb.CreateTime = time.Now().Unix()
	reqPb.AdminRoleRemarks = req.AdminRoleRemarks

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AddAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterAdminRoleRequest
		reqPb pbAdmin.AlterAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.AdminRoleID = req.AdminRoleID
	reqPb.AdminRoleName = req.AdminRoleName
	reqPb.Status = req.AdminRoleStatus
	reqPb.AdminAPIsIDs = req.AdminAPIsIDs
	reqPb.AdminPagesIDs = req.AdminPagesIDs
	reqPb.AdminRoleDiscription = req.AdminRoleDiscription
	reqPb.UpdateUser = userID
	reqPb.UpdateTime = time.Now().Unix()
	reqPb.AdminRoleRemarks = req.AdminRoleRemarks
	reqPb.OperationID = req.OperationID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AlterAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeleteAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterAdminRoleRequest
		reqPb pbAdmin.AlterAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.AdminRoleID = req.AdminRoleID
	reqPb.CreateUser = userID
	reqPb.OperationID = req.OperationID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetAllAdminRoles(c *gin.Context) {
	var (
		req   apiStruct.GetAllAdminRolesRequest
		resp  apiStruct.GetAllAdminRolesResponse
		reqPb pbAdmin.GetAllAdminRolesReq
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb.Pagination, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetAllAdminRoles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.AdminRoles, respPb.AdminRoles)
	resp.PageSizeLimit = int64(respPb.Pagination.ShowNumber)
	resp.PageNumber = int64(respPb.Pagination.CurrentPage)
	resp.TotalRecCount = int64(respPb.AdminRolesNums)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func SearchAminRoles(c *gin.Context) {
	var (
		req   apiStruct.SearchAminRolesRequest
		reqPb pbAdmin.SearchAminRolesRequest
		resp  apiStruct.GetAllAdminRolesResponse
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	OperationID := utils.OperationIDGenerator()
	log.NewInfo(OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, OperationID)
	if etcdConn == nil {
		errMsg := OperationID + "getcdv3.GetConn == nil"
		log.NewError(OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.SearchAminRoles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.AdminRoles, respPb.AdminRoles)
	resp.PageSizeLimit = int64(respPb.Pagination.ShowNumber)
	resp.PageNumber = int64(respPb.Pagination.CurrentPage)
	resp.TotalRecCount = int64(respPb.AdminRolesNums)
	log.NewInfo(OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

// Admin API Path CRUD APIs
func AddApiAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AddApiInAdminRoleRequest
		reqPb pbAdmin.AddApiAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)

	reqPb.OperationID = req.OperationID
	reqPb.ApiName = req.ApiName
	reqPb.ApiPath = req.ApiPath
	reqPb.Status = int64(req.Status)
	reqPb.CreateUser = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AddApiAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterApiAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterApiInAdminRoleRequest
		reqPb pbAdmin.AlterApiAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.ApiID = req.ApiID
	reqPb.ApiName = req.ApiName
	reqPb.Status = int32(req.Status)
	reqPb.OperationID = req.OperationID
	reqPb.CreateUser = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AlterApiAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func EnDisableApiAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterApiInAdminRoleRequest
		reqPb pbAdmin.AlterApiAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.ApiID = req.ApiID
	reqPb.Status = int32(req.Status)
	reqPb.OperationID = req.OperationID
	reqPb.CreateUser = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AlterApiAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeleteApiAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterApiInAdminRoleRequest
		reqPb pbAdmin.AlterApiAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.ApiID = req.ApiID
	reqPb.CreateUser = userID
	reqPb.OperationID = req.OperationID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteApiAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetAllApiAdminRoles(c *gin.Context) {
	var (
		req   apiStruct.GetAllApiInAdminRoleRequest
		resp  apiStruct.GetAllApiInAdminRoleResponse
		reqPb pbAdmin.GetAllApiAdminRolesReq
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb.Pagination, &req)
	reqPb.OperationID = utils.OperationIDGenerator()

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetAllApiAdminRoles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.ApiInAdminRoles, respPb.ApisAdminRole)
	resp.PageNumber = int64(respPb.Pagination.CurrentPage)
	resp.PageSizeLimit = int64(respPb.Pagination.ShowNumber)
	resp.TotalRecCount = int64(respPb.ApiNums)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

// SearchApiAdminRoles
func SearchApiAdminRoles(c *gin.Context) {
	var (
		req   apiStruct.SearchApiAdminRoleRequest
		reqPb pbAdmin.SearchApiAdminRoleRequest
		resp  apiStruct.GetAllApiInAdminRoleResponse
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	utils.CopyStructFields(&reqPb, req)
	operationID := utils.OperationIDGenerator()
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, operationID)
	if etcdConn == nil {
		errMsg := operationID + "getcdv3.GetConn == nil"
		log.NewError(operationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.SearchApiAdminRoles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.ApiInAdminRoles, respPb.ApisAdminRole)
	resp.PageNumber = int64(respPb.Pagination.CurrentPage)
	resp.PageSizeLimit = int64(respPb.Pagination.ShowNumber)
	resp.TotalRecCount = int64(respPb.ApiNums)
	log.NewInfo(utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

// Admin Pages Path CRUD APIs
func AddPageAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AddPageInAdminRoleRequest
		reqPb pbAdmin.AddPageAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)

	reqPb.OperationID = req.OperationID
	reqPb.PageName = req.PageName
	reqPb.PagePath = req.PagePath
	reqPb.Status = int64(req.Status)
	reqPb.CreateUser = userID
	reqPb.FatherPageID = req.FatherPageID
	reqPb.IsButton = req.IsButton
	reqPb.IsMenu = req.IsMenu
	reqPb.SortPriority = req.SortPriority
	reqPb.AdminAPIsIDs = req.AdminAPIsIDs

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AddPageAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterPageAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterPageInAdminRoleRequest
		reqPb pbAdmin.AlterPageAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.PageID = req.PageID
	reqPb.PageName = req.PageName
	reqPb.Status = int32(req.Status)
	reqPb.OperationID = req.OperationID
	reqPb.CreateUser = userID
	reqPb.FatherPageID = req.FatherPageID
	reqPb.IsButton = req.IsButton
	reqPb.IsMenu = req.IsMenu
	reqPb.SortPriority = req.SortPriority
	reqPb.AdminAPIsIDs = req.AdminAPIsIDs
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.AlterPageAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeletePageAdminRole(c *gin.Context) {
	var (
		req   apiStruct.AlterPageInAdminRoleRequest
		reqPb pbAdmin.AlterPageAdminRoleRequest
	)
	req.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.PageID = req.PageID
	reqPb.CreateUser = userID
	reqPb.OperationID = req.OperationID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.DeletePageAdminRole(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetAllPageAdminRoles(c *gin.Context) {
	var (
		req   apiStruct.GetAllPageInAdminRoleRequest
		resp  apiStruct.GetAllPageInAdminRoleResponse
		reqPb pbAdmin.GetAllPageAdminRolesReq
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	reqPb.FatherIDFilter = req.FatherIDFilter
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb.Pagination, &req)
	log.NewError(utils.GetSelfFuncName(), "SchMax pre", req.FatherIDFilter, reqPb.FatherIDFilter)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetAllPageAdminRoles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.PageInAdminRoles, respPb.PagesAdminRole)
	utils.CopyStructFields(&resp.ApiInPageAdminRoles, respPb.ApisAdminRole)
	utils.CopyStructFields(&resp.FatherPageInAdminRoles, respPb.FatherPagesAdminRole)
	resp.PageSizeLimit = int64(respPb.Pagination.ShowNumber)
	resp.PageNumber = int64(respPb.Pagination.CurrentPage)
	resp.TotalRecCount = int64(respPb.TotalRecCount)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

// SearchPageAdminRoles
func SearchPageAdminRoles(c *gin.Context) {
	var (
		req   apiStruct.SearchPageAdminRolesRequest
		reqPb pbAdmin.SearchPageAdminRolesRequest
		resp  apiStruct.GetAllPageInAdminRoleResponse
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	utils.CopyStructFields(&reqPb, req)
	operationID := utils.OperationIDGenerator()
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, operationID)
	if etcdConn == nil {
		errMsg := operationID + "getcdv3.GetConn == nil"
		log.NewError(operationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.SearchPageAdminRoles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.PageInAdminRoles, respPb.PagesAdminRole)
	utils.CopyStructFields(&resp.ApiInPageAdminRoles, respPb.ApisAdminRole)
	utils.CopyStructFields(&resp.FatherPageInAdminRoles, respPb.FatherPagesAdminRole)
	resp.PageNumber = int64(respPb.Pagination.CurrentPage)
	resp.PageSizeLimit = int64(respPb.Pagination.ShowNumber)
	resp.TotalRecCount = int64(respPb.TotalRecCount)
	log.NewInfo(utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

// SearchOperationLogs
func SearchOperationLogs(c *gin.Context) {
	var (
		req   apiStruct.SearchOperationLogsRequest
		reqPb pbAdmin.SearchOperationLogsRequest
		resp  apiStruct.SearchOperationLogsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	utils.CopyStructFields(&reqPb, req)
	operationID := utils.OperationIDGenerator()
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, operationID)
	if etcdConn == nil {
		errMsg := operationID + "getcdv3.GetConn == nil"
		log.NewError(operationID, errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.SearchOperationLogs(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.OperationLogs, respPb.OperationLogs)
	resp.PageNumber = respPb.PageNumber
	resp.PageSizeLimit = respPb.PageLimit
	resp.TotalRecCount = respPb.OperationLogsCount

	log.NewInfo(utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

// Admin Action CURD APIs

// func AddAdminAction(c *gin.Context) {
// 	var (
// 		req   apiStruct.AddAdminActionRequest
// 		reqPb pbAdmin.AddAdminActionRequest
// 	)
// 	req.OperationID = utils.OperationIDGenerator()
// 	if err := c.BindJSON(&req); err != nil {
// 		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
// 		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
// 		return
// 	}
// 	userID := ""
// 	userIDInter, exsited := c.Get("userID")
// 	if exsited {
// 		userID = userIDInter.(string)
// 	}
// 	reqPb.AdminActionName = req.AdminActionName
// 	reqPb.AdminAPIsIDs = req.AdminAPIsIDs
// 	reqPb.AdminPagesIDs = req.AdminPagesIDs
// 	reqPb.CreateUser = userID
// 	reqPb.OperationID = req.OperationID

// 	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
// 	utils.CopyStructFields(&reqPb, &req)
// 	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
// 	if etcdConn == nil {
// 		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
// 		log.NewError(reqPb.OperationID, errMsg)
// 		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
// 		return
// 	}
// 	client := pbAdmin.NewAdminCMSClient(etcdConn)
// 	_, err := client.AddAdminAction(context.Background(), &reqPb)
// 	if err != nil {
// 		openIMHttp.RespHttp200(c, err, nil)
// 		return
// 	}
// 	openIMHttp.RespHttp200(c, constant.OK, nil)
// }

// func AlterAdminAction(c *gin.Context) {
// 	var (
// 		req   apiStruct.AlterAdminActionRequest
// 		reqPb pbAdmin.AlterAdminActionRequest
// 	)
// 	req.OperationID = utils.OperationIDGenerator()
// 	if err := c.BindJSON(&req); err != nil {
// 		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
// 		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
// 		return
// 	}
// 	userID := ""
// 	userIDInter, exsited := c.Get("userID")
// 	if exsited {
// 		userID = userIDInter.(string)
// 	}
// 	reqPb.AdminActionID = req.AdminActionID
// 	reqPb.AdminActionName = req.AdminActionName
// 	reqPb.AdminAPIsIDs = req.AdminAPIsIDs
// 	reqPb.AdminPagesIDs = req.AdminPagesIDs
// 	reqPb.CreateUser = userID
// 	reqPb.OperationID = req.OperationID
// 	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
// 	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
// 	if etcdConn == nil {
// 		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
// 		log.NewError(reqPb.OperationID, errMsg)
// 		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
// 		return
// 	}
// 	client := pbAdmin.NewAdminCMSClient(etcdConn)
// 	_, err := client.AlterAdminAction(context.Background(), &reqPb)
// 	if err != nil {
// 		openIMHttp.RespHttp200(c, err, nil)
// 		return
// 	}
// 	openIMHttp.RespHttp200(c, constant.OK, nil)
// }

// func DeleteAdminAction(c *gin.Context) {
// 	var (
// 		req   apiStruct.AlterAdminActionRequest
// 		reqPb pbAdmin.AlterAdminActionRequest
// 	)
// 	req.OperationID = utils.OperationIDGenerator()
// 	if err := c.BindJSON(&req); err != nil {
// 		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
// 		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
// 		return
// 	}
// 	userID := ""
// 	userIDInter, exsited := c.Get("userID")
// 	if exsited {
// 		userID = userIDInter.(string)
// 	}
// 	reqPb.AdminActionID = req.AdminActionID
// 	reqPb.CreateUser = userID
// 	reqPb.OperationID = req.OperationID
// 	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
// 	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
// 	if etcdConn == nil {
// 		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
// 		log.NewError(reqPb.OperationID, errMsg)
// 		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
// 		return
// 	}
// 	client := pbAdmin.NewAdminCMSClient(etcdConn)
// 	_, err := client.DeleteAdminAction(context.Background(), &reqPb)
// 	if err != nil {
// 		openIMHttp.RespHttp200(c, err, nil)
// 		return
// 	}
// 	openIMHttp.RespHttp200(c, constant.OK, nil)
// }

// func GetAllAdminAction(c *gin.Context) {
// 	var (
// 		req   apiStruct.GetAllAdminActionsRequest
// 		resp  apiStruct.GetAllAdminActionsResponse
// 		reqPb pbAdmin.GetAllAdminActionReq
// 	)
// 	reqPb.Pagination = &commonPb.RequestPagination{}
// 	if err := c.ShouldBindQuery(&req); err != nil {
// 		log.NewError("0", "ShouldBindQuery failed ", err.Error())
// 		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
// 		return
// 	}
// 	reqPb.OperationID = utils.OperationIDGenerator()
// 	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

// 	utils.CopyStructFields(&reqPb.Pagination, &req)
// 	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
// 	if etcdConn == nil {
// 		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
// 		log.NewError(reqPb.OperationID, errMsg)
// 		c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
// 		return
// 	}
// 	client := pbAdmin.NewAdminCMSClient(etcdConn)
// 	respPb, err := client.GetAllAdminAction(context.Background(), &reqPb)
// 	if err != nil {
// 		openIMHttp.RespHttp200(c, err, resp)
// 		return
// 	}
// 	utils.CopyStructFields(&resp.AdminActionss, respPb.AdminActions)
// 	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
// 	resp.CurrentPage = int(respPb.Pagination.CurrentPage)
// 	resp.ActionsNums = int64(respPb.AdminActionNums)
// 	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
// 	openIMHttp.RespHttp200(c, constant.OK, resp)

// }

// Utils Function blewThis line
//
// genrateTOTPForNow genrate TOTP code for varification
//
//	func genrateTOTPForNow(user db.AdminUser) string {
//		totp := gotp.NewDefaultTOTP(user.Google2fSecretKey)
//		totpCode := totp.Now()
//		return totpCode
//	}
func genrateTOTPForNow(user db.AdminUser) string {
	// totp := gotp.NewDefaultTOTP(user.Google2fSecretKey)
	secret, _ := gotp.DecodeBase32(user.Google2fSecretKey)
	totp, _ := gotp.NewTOTP(secret)
	totpCode, err := totp.Now()
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Genrating TOTP Failed", err.Error())
	}
	return totpCode
}

// triversPagesByParents maping data
func triversPagesByParents(adminPermissions *pbAdmin.AdminPermissionsResp) map[string]interface{} {
	response := make(map[string]interface{})
	admin_Role := make(map[string]interface{})
	var allowedApiSlice []pbAdmin.AdminApiPath
	var allowedPagesSlice []apiStruct.AdminPagePath
	var allowedPagesSliceFinal []apiStruct.AdminPagePath
	triversMap := make(map[int64]apiStruct.AdminPagePath)

	// for _, adminAction := range adminPermissions.AdminRole.AdminActions {
	for _, allowedApi := range adminPermissions.AdminRole.AllowedApis {
		allowedApiSlice = append(allowedApiSlice, *allowedApi)
	}
	for _, allowedPage := range adminPermissions.AdminRole.AllowedPages {
		var temp apiStruct.AdminPagePath
		temp.FatherPageID = allowedPage.FatherPageID
		temp.Id = allowedPage.Id
		temp.PageName = allowedPage.PageName
		temp.IsMenu = allowedPage.IsMenu
		temp.SortPriority = allowedPage.SortPriority
		temp.PagePath = allowedPage.PagePath
		temp.IsButton = allowedPage.IsButton
		temp.AdminAPIsIDs = allowedPage.AdminAPIsIDs
		if _, ok := triversMap[allowedPage.Id]; !ok {
			triversMap[allowedPage.Id] = temp
			allowedPagesSlice = append(allowedPagesSlice, temp)
		}

	}
	// }

	for _, expChild := range allowedPagesSlice {
		if expChild.FatherPageID != 0 {
			fatherAllowdPage := triversMap[expChild.FatherPageID]
			fatherAllowdPage.Childs = append(fatherAllowdPage.Childs, expChild)
			fatherAllowdPage.ChildsCount++
			triversMap[fatherAllowdPage.Id] = fatherAllowdPage
		}
	}

	for key := range triversMap {
		allowdPage := triversMap[key]
		if allowdPage.FatherPageID == 0 {
			allowedPageFinal := getChildsOfAllowedPages(triversMap, allowdPage)
			allowedPagesSliceFinal = append(allowedPagesSliceFinal, allowedPageFinal)
		}
	}
	sort.Slice(allowedPagesSliceFinal, func(i, j int) bool {
		return allowedPagesSliceFinal[i].SortPriority < allowedPagesSliceFinal[j].SortPriority
	})

	admin_Role["allowedApis"] = allowedApiSlice
	admin_Role["allowedPages"] = allowedPagesSliceFinal
	response["adminRole"] = admin_Role
	return response
}

func getChildsOfAllowedPages(allowedPageMap map[int64]apiStruct.AdminPagePath, child apiStruct.AdminPagePath) apiStruct.AdminPagePath {

	for position, allowedPageChild := range child.Childs {
		childInMap := allowedPageMap[allowedPageChild.Id]
		if len(childInMap.Childs) > 0 {
			childInMap = getChildsOfAllowedPages(allowedPageMap, childInMap)
		}
		child.Childs[position] = childInMap
		sort.Slice(child.Childs, func(i, j int) bool {
			return child.Childs[i].SortPriority < child.Childs[j].SortPriority
		})
	}
	return child
}
