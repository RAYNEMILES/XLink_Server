package me_page_url

import (
	"Open_IM/pkg/base_info"
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

func GetOTCUrl(c *gin.Context) {
	//params init
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLResp
		reqPb    pbAdmin.GetMePageURLReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.PageType = constant.MePageTypeOTC
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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData})

}

func DepositURL(c *gin.Context) {
	//params init
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLResp
		reqPb    pbAdmin.GetMePageURLReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.PageType = constant.MePageTypeDeposit
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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData})

}

func MarketURL(c *gin.Context) {
	//params init
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLResp
		reqPb    pbAdmin.GetMePageURLReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.PageType = constant.MePageTypeMarket
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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData})

}

func WithdrawURL(c *gin.Context) {
	//params init
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLResp
		reqPb    pbAdmin.GetMePageURLReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.PageType = constant.MePageTypeWithdraw
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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData})

}

func ExchangeURL(c *gin.Context) {
	//params init
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLResp
		reqPb    pbAdmin.GetMePageURLReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.PageType = constant.MePageTypeExchange
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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData})

}

func EarnURL(c *gin.Context) {
	//params init
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLResp
		reqPb    pbAdmin.GetMePageURLReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = req.OperationID
	reqPb.PageType = constant.MePageTypeEarn
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
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData})

}

func GetMePageURLs(c *gin.Context) {
	var (
		req      base_info.GetMePageURLReq
		respData base_info.GetMePageURLsResp
		reqPb    pbAdmin.GetMePageURLsReq
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
	respPb, err := client.GetMePageURLs(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetActiveGroup failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	_ = utils.CopyStructFields(&respData, respPb)

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", respData)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": respData.MePageURL})

}
