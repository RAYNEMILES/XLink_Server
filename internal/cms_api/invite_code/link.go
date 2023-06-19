package inviteCode

import (
	apiStruct "Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/admin_cms"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetBaseLink(c *gin.Context) {
	var reqPb = admin_cms.GetInviteCodeBaseLinkReq{}
	reqPb.OperationID = utils2.OperationIDGenerator()

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, reqPb.OperationID)
	if userID == "" {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// check the permissions
	if err := utils2.CheckAdminPermissions(userID); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	res, _ := client.GetInviteCodeBaseLink(context.Background(), &reqPb)

	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetInviteCodeBaseLink failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, apiStruct.InviteCodeBashLinkResponse{BaseLink: res.InviteCodeBaseLink})
}

func SetBaseLink(c *gin.Context) {
	var (
		apiReq apiStruct.SetInviteCodeBashLinkRequest

		rpcReq pbAdmin.SetInviteCodeBaseLinkReq
	)
	if err := c.BindJSON(&apiReq); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiReq.OperationID)
	if userID == "" {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// check the permissions
	if err := utils2.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, apiReq.OperationID)
	if etcdConn == nil {
		errMsg := apiReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(apiReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	rpcReq.OperationID = apiReq.OperationID
	rpcReq.Value = apiReq.BaseLink

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	res, _ := client.SetInviteCodeBaseLink(context.Background(), &rpcReq)

	if err != nil {
		log.NewError(rpcReq.OperationID, utils.GetSelfFuncName(), "SetInviteCodeBaseLink failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, res)
}
