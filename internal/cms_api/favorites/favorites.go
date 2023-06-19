package favorites

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/office"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

func GetFavorites(c *gin.Context) {
	var (
		req   cms_api_struct.GetFavoritesRequest
		resp  cms_api_struct.GetFavoritesResponse
		reqPb pb.GetFavoritesReq
	)

	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
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

	reqPb.ContentType = make([]int32, 0)
	err = json.Unmarshal([]byte(req.ContentType), &reqPb.ContentType)
	if err != nil {
		errMsg := reqPb.OperationID + "Content Type json error" + err.Error()
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
	client := pb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.GetFavorites(context.Background(), &reqPb)
	if err != nil {
		errMsg := reqPb.OperationID + "get favorites rpc call error" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	err = utils.CopyStructFields(&resp.Favorites, rpcResp.Favorites)
	if err != nil {
		errMsg := reqPb.OperationID + "Copy rpc result failed" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	log.Debug("", "resp.Favorites: ", resp.Favorites)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.FavoriteNums = rpcResp.FavoriteNums

	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func DeleteFavorites(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteFavoritesRequest
		resp  cms_api_struct.DeleteFavoritesResponse
		reqPb pb.RemoveFavoriteReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "Copy request failed" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	userID := ""
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID
	reqPb.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.RemoveFavorite(context.Background(), &reqPb)
	if err != nil {
		errMsg := reqPb.OperationID + "get favorites rpc call error" + err.Error()
		log.NewError(reqPb.OperationID, errMsg)
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		c.JSON(http.StatusBadRequest, gin.H{"errCode": resp.ErrCode, "errMsg": resp.ErrMsg, "data": resp})
		return
	}

	resp.CommResp = cms_api_struct.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}
	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})

}

func AlterFavorites(c *gin.Context) {
	var (
		req   cms_api_struct.AlterFavoritesRequest
		resp  cms_api_struct.AlterFavoritesResponse
		reqPb pb.AlterFavoritesReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}

	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationId + "Copy request failed" + err.Error()
		log.NewError(reqPb.OperationId, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	userID := ""
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	reqPb.OpUserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, reqPb.OperationId)
	if etcdConn == nil {
		errMsg := reqPb.OperationId + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationId, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pb.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.AlterFavorites(context.Background(), &reqPb)
	if err != nil {
		errMsg := reqPb.OperationId + "get favorites rpc call error" + err.Error()
		log.NewError(reqPb.OperationId, errMsg)
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		c.JSON(http.StatusBadRequest, gin.H{"errCode": resp.ErrCode, "errMsg": resp.ErrMsg, "data": resp})
		return
	}

	resp.CommResp = cms_api_struct.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}
	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})

}
