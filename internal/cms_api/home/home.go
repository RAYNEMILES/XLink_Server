package home

import (
	apiStruct "Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetStatus(c *gin.Context) {
	var (
		req   apiStruct.GetStatusReq
		resp  apiStruct.GetStatusResp
		reqPb pbAdmin.GetStatusReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	resp.StatusList = make(map[string]int8)
	for k, v := range respPb.StatusList {
		resp.StatusList[k] = int8(v)
	}
	resp.CommResp = apiStruct.CommResp{}
	resp.CommResp.ErrCode = respPb.CommonResp.ErrCode
	resp.CommResp.ErrMsg = respPb.CommonResp.ErrMsg

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)

}

func SetStatus(c *gin.Context) {
	var (
		req   apiStruct.SetStatusReq
		reqPb pbAdmin.SetStatusReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.StatusList = make(map[string]int32)
	for k, v := range req.StatusList {
		reqPb.StatusList[k] = int32(v)
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.SetStatus(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": ""})

}
