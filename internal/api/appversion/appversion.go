package appversion

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/front_api_struct"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/proto/appversion"
	"Open_IM/pkg/utils"
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetLatestAppVersion(c *gin.Context) {
	//params init
	var (
		req      front_api_struct.GetLatestAppVersionReq
		resp     front_api_struct.FrontApiResp
		respData front_api_struct.GetLatestAppVersionResp
		reqPb    appversion.GetLatestAppVersionReq
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

	//request the gRPC client
	reqPb.Client = req.Client
	reqPb.OperationID = req.OperationID
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

	client := admin_cms.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetLatestAppVersion(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		resp = front_api_struct.FrontApiResp{
			ErrCode: constant.ErrGetVersion.ErrCode,
			ErrMsg:  constant.ErrGetVersion.ErrMsg,
			Data:    nil,
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	//return response data
	respData = front_api_struct.GetLatestAppVersionResp{
		ID:          respPb.Appversion.ID,
		Version:     respPb.Appversion.Version,
		Type:        int(respPb.Appversion.Type),
		Status:      int(respPb.Appversion.Status),
		Isforce:     int(respPb.Appversion.Isforce),
		Title:       respPb.Appversion.Title,
		DownloadUrl: respPb.Appversion.DownloadUrl,
		Content:     respPb.Appversion.Content,
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
