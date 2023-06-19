package discover

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/front_api_struct"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetDiscoverUrl(c *gin.Context) {
	//params init
	var (
		req      front_api_struct.GetDiscoverUrlReq
		resp     front_api_struct.FrontApiResp
		respData front_api_struct.GetDiscoverUrlResp
		reqPb    pbAdmin.GetDiscoverUrlReq
	)

	//check the params from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		resp = front_api_struct.FrontApiResp{
			ErrCode: http.StatusBadRequest,
			ErrMsg:  err.Error(),
			Data:    nil,
		}
		c.JSON(int(resp.ErrCode), resp)
		return
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	//check the token
	token := c.GetHeader("token")
	if token == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "token is nil")
		resp = front_api_struct.FrontApiResp{
			ErrCode: constant.ErrArgs.ErrCode,
			ErrMsg:  constant.ErrArgs.ErrMsg,
			Data:    nil,
		}
		c.JSON(int(resp.ErrCode), resp)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, req.OperationID)
	if userID == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "token is illegal")
		resp = front_api_struct.FrontApiResp{
			ErrCode: constant.ErrArgs.ErrCode,
			ErrMsg:  constant.ErrArgs.ErrMsg,
			Data:    nil,
		}
		c.JSON(int(resp.ErrCode), resp)
		return
	}

	//check user status
	if err := utils2.CheckUserPermissions(userID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "user is banned!", userID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	//request the gRPC client
	reqPb.PlatformID = "1"
	reqPb.OperationID = req.OperationID
	reqPb.UserId = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "etcd connect failed")
		resp = front_api_struct.FrontApiResp{
			ErrCode: constant.ErrServer.ErrCode,
			ErrMsg:  "etcd connect failed",
			Data:    nil,
		}
		c.JSON(int(resp.ErrCode), resp)
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetDiscoverUrl(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		resp = front_api_struct.FrontApiResp{
			ErrCode: constant.ErrRPC.ErrCode,
			ErrMsg:  constant.ErrRPC.ErrMsg,
			Data:    nil,
		}
		c.JSON(int(resp.ErrCode), resp)
		return
	}

	//return response data
	respData = front_api_struct.GetDiscoverUrlResp{
		Url:        respPb.Url.Url,
		Status:     int(respPb.Url.Status),
		PlatformId: int(respPb.Url.PlatformId),
	}
	data := structs.Map(respData)
	resp = front_api_struct.FrontApiResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
		Data:    data,
	}
	c.JSON(int(resp.ErrCode), resp)
	return
}
