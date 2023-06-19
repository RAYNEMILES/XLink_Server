package discover

import (
	apiStruct "Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// get discover url
func GetDiscoverUrl(c *gin.Context) {
	var (
		req   apiStruct.GetDiscoverUrlReq
		resp  apiStruct.GetDiscoverUrlResp
		reqPb pbAdmin.GetDiscoverUrlReq
	)

	//check the params form request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	//request gRPC client
	reqPb.PlatformID = req.PlatformID
	reqPb.OperationID = req.OperationID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + " getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetDiscoverUrl(c, &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp = apiStruct.GetDiscoverUrlResp{
		ID:         uint(respPb.Url.ID),
		Url:        respPb.Url.Url,
		Status:     int(respPb.Url.Status),
		PlatformId: int(respPb.Url.PlatformId),
		CreateTime: respPb.Url.CreateTime,
		CreateUser: respPb.Url.CreateUser,
		UpdateTime: respPb.Url.UpdateTime,
		UpdateUser: respPb.Url.UpdateUser,
	}
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

// save discover url
func SaveDiscoverUrl(c *gin.Context) {

	var (
		req   apiStruct.SaveDiscoverUrlReq
		reqPb pbAdmin.SaveDiscoverUrlReq
	)

	//check the prams form request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)

	}

	//request gRPC client
	reqPb.PlatformID = req.PlatformID
	reqPb.OperationID = req.OperationID
	reqPb.Url = req.Url
	reqPb.UserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + " getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.SaveDiscoverUrl(c, &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

// switch the visible status of discover page
func SwitchDiscoverPage(c *gin.Context) {
	var (
		req   apiStruct.SwitchDiscoverStatusReq
		reqPb pbAdmin.SwitchDiscoverStatusReq
	)

	//check the parms from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)

	}

	//request gRPC client
	reqPb.PlatformID = req.PlatformID
	reqPb.OperationID = req.OperationID
	reqPb.Status = uint32(req.Status)
	reqPb.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + " getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.SwitchDiscoverStatus(c, &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}
