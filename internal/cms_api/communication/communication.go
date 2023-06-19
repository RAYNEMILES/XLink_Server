package communication

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	officePb "Open_IM/pkg/proto/office"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetCommunications(c *gin.Context) {
	var (
		req   cms_api_struct.GetCommunicationsReq
		resp  cms_api_struct.GetCommunicationsResp
		reqPb officePb.GetCommunicationsReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	reqPb.Pagination = &commonPb.RequestPagination{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "Copy request failed" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := officePb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.GetCommunications(context.Background(), &reqPb)
	if err != nil {
		errMsg := reqPb.OperationID + "get favorites rpc call error" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	err = utils.CopyStructFields(&resp.CommunicationList, &rpcResp.CommunicationList)
	if err != nil {
		errMsg := reqPb.OperationID + "Copy rpc result failed" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.Communications = rpcResp.Communications

	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func DeleteCommunications(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteCommunicationsReq
		resp  cms_api_struct.DeleteCommunicationsResp
		reqPb officePb.DeleteCommunicationsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
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
	reqPb.OpUserID = userID
	reqPb.CommunicatIDs = req.CommunicatIDs

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := officePb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.DeleteCommunications(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	resp.CommResp = cms_api_struct.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func SetRemark(c *gin.Context) {
	var (
		req   cms_api_struct.SetRemarkReq
		resp  cms_api_struct.SetRemarkResp
		reqPb officePb.SetRemarkReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
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
	reqPb.OpUserID = userID
	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "Copy request failed" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := officePb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.SetRemark(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	resp.CommResp = cms_api_struct.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func InterruptPersonalCommunications(c *gin.Context) {
	var (
		req   cms_api_struct.InterruptCommunicationReq
		resp  cms_api_struct.InterruptCommunicationResp
		reqPb officePb.InterruptPersonalCommunicationsReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
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
	reqPb.OpUserID = userID
	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "Copy request failed" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := officePb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.InterruptPersonalCommunications(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	resp.CommResp = cms_api_struct.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}
	openIMHttp.RespHttp200(c, constant.OK, resp)
}
