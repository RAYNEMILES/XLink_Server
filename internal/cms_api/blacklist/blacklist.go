package blacklist

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/friend"
	server_api_params "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetBlacks(c *gin.Context) {
	var (
		req    cms_api_struct.GetBlacksReq
		resp   cms_api_struct.GetBlacksResp
		rpcReq pb.GetBlacksReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, &req)
	rpcReq.Pagination = &server_api_params.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewFriendClient(etcdConn)
	rpcResp, err := client.GetBlacks(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	err = utils.CopyStructFields(&resp.BlackList, rpcResp.BlackList)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	log.Debug("", len(resp.BlackList), "rpcResp.BlackList :", len(rpcResp.BlackList))
	resp.ListNumber = rpcResp.ListNumber
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber

	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func RemoveBlack(c *gin.Context) {
	var (
		req    cms_api_struct.RemoveBlackReq
		resp   cms_api_struct.RemoveBlackResp
		rpcReq pb.RemoveBlackReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	_ = utils.CopyStructFields(&rpcReq, &req)

	rpcReq.OpUserID = userID
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewFriendClient(etcdConn)
	rpcResp, err := client.RemoveBlack(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, rpcResp.CommResp.ErrMsg)

}

func AlterRemark(c *gin.Context) {
	var (
		req    cms_api_struct.AlterRemarkReq
		resp   cms_api_struct.AlterRemarkResp
		rpcReq pb.AlterRemarkReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	rpcReq.Remark = req.Remark
	rpcReq.BlackID = req.BlockUser
	rpcReq.OwnerID = req.OwnerUser
	rpcReq.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewFriendClient(etcdConn)
	rpcResp, err := client.AlterRemark(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, rpcResp.CommResp.ErrMsg)

}
