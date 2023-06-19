package me_page

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

func GetURL(c *gin.Context) {
	var (
		req   apiStruct.GetURLReq
		resp  apiStruct.GetURLResp
		reqPb pbAdmin.GetMePageURLReq
	)

	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.PageType = int64(req.Type)
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
	respPb, err := client.GetMePageURL(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	_ = utils.CopyStructFields(&resp, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func SaveUrl(c *gin.Context) {
	var (
		req   apiStruct.SaveMePageUrlReq
		resp  apiStruct.SaveMePageUrlResp
		reqPb pbAdmin.SaveMePageURLReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	reqPb.PageType = int64(req.Type)
	reqPb.OpUserID = userID
	reqPb.Url = req.Url
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
	respPb, err := client.SaveMePageURL(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.CommResp = apiStruct.CommResp{}
	resp.CommResp.ErrCode = respPb.CommonResp.ErrCode
	resp.CommResp.ErrMsg = respPb.CommonResp.ErrMsg

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp.CommResp})

}

func SwitchStatus(c *gin.Context) {
	var (
		req   apiStruct.SwitchStatusReq
		resp  apiStruct.SwitchStatusResp
		reqPb pbAdmin.SwitchMePageURLReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	reqPb.OpUserID = userID
	reqPb.PageType = int64(req.Type)
	reqPb.Status = int32(req.Status)
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
	respPb, err := client.SwitchMePageURL(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	resp.CommResp = apiStruct.CommResp{}
	resp.CommResp.ErrCode = respPb.CommonResp.ErrCode
	resp.CommResp.ErrMsg = respPb.CommonResp.ErrMsg

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp.CommResp})

}
