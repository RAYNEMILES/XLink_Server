package domain

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

func GetAllDomains(c *gin.Context) {
	var (
		req   apiStruct.GetAllDomainsReq
		resp  apiStruct.GetAllDomainsResp
		reqPb pbAdmin.GetAllDomainsReq
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
	respPb, err := client.GetAllDomains(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	resp.Domains = respPb.Domains
	resp.CommResp = apiStruct.CommResp{}
	resp.CommResp.ErrCode = respPb.CommonResp.ErrCode
	resp.CommResp.ErrMsg = respPb.CommonResp.ErrMsg

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)

	c.JSON(http.StatusOK, resp)
}

func SaveAllDomains(c *gin.Context) {
	var (
		req   apiStruct.SaveAllDomainsReq
		reqPb pbAdmin.SaveAllDomainsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.Domains = req.Domains
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbAdmin.NewAdminCMSClient(etcdConn)
	_, err := client.SaveAllDomains(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}
