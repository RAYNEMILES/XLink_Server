package game_store

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/game_store"
	sdk_ws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"
)

func GetGameList(c *gin.Context) {

	var (
		req   cms_api_struct.GetGameListReq
		resp  cms_api_struct.GetGameListResp
		reqPb pb.GetGameListReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		resp.ErrCode = constant.ErrArgs.ErrCode
		resp.ErrMsg = constant.ErrArgs.ErrMsg
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	resp.CommRespCode = cms_api_struct.CommRespCode{}
	err := utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination = &sdk_ws.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}
	reqPb.Categories = req.Category

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.GetGameList(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	resp.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.ErrCode = rpcResp.CommonResp.ErrCode

	if resp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
		return
	}

	_ = utils.CopyStructFields(&resp.GameList, rpcResp.GameList)

	for index, info := range rpcResp.GameList {
		resp.GameList[index].PlatformLink = make(map[int32]cms_api_struct.GameLink, 0)
		_ = json.Unmarshal([]byte(info.SupportPlatform), &resp.GameList[index].PlatformLink)
		log.Debug("", "get game list: ", info.Categories)
	}

	resp.GameNum = rpcResp.GameNum

	resp.ResponsePagination = cms_api_struct.ResponsePagination{CurrentPage: int(rpcResp.Pagination.CurrentPage), ShowNumber: int(rpcResp.Pagination.ShowNumber)}

	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func EditGame(c *gin.Context) {

	var (
		req   cms_api_struct.EditGameReq
		resp  cms_api_struct.EditGameResp
		reqPb pb.EditGameReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	resp.CommRespCode = cms_api_struct.CommRespCode{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	reqPb.Game = &pb.GameBackgroundInfo{}
	_ = utils.CopyStructFields(reqPb.Game, req)
	platform, err := json.Marshal(req.PlatformLink)
	if err != nil {
		errMsg := reqPb.OperationID + "The platform params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	reqPb.Game.SupportPlatform = string(platform)

	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.EditGame(context.Background(), &reqPb)
	if err != nil {
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	if resp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})

}

func AddGame(c *gin.Context) {
	var (
		req   cms_api_struct.AddGameReq
		resp  cms_api_struct.AddGameResp
		reqPb pb.AddGameReq
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

	resp.CommRespCode = cms_api_struct.CommRespCode{}
	reqPb.OperationID = utils.OperationIDGenerator()
	reqPb.Game = &pb.GameBackgroundInfo{}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	err := utils.CopyStructFields(reqPb.Game, req)
	if err != nil {
		errMsg := reqPb.OperationID + "CopyStructFields error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	platform, err := json.Marshal(req.PlatformLink)
	if err != nil {
		errMsg := reqPb.OperationID + "The platform params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	reqPb.Game.SupportPlatform = string(platform)

	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.AddGame(context.Background(), &reqPb)
	if err != nil {
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	if resp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
}

func DeleteGames(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteGamesReq
		resp  cms_api_struct.DeleteGamesResp
		reqPb pb.DeleteGamesReq
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

	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "The request params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.DeleteGames(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.ErrCode = rpcResp.CommonResp.ErrCode

	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})

}

func GetCategory(c *gin.Context) {
	//TODO
	var (
		req   cms_api_struct.GetCategoryReq
		resp  cms_api_struct.GetCategoryResp
		reqPb pb.GetCategoryReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	resp.CommRespCode = cms_api_struct.CommRespCode{}
	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "The request params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	reqPb.Pagination = &sdk_ws.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.GetCategory(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.ErrCode = rpcResp.CommonResp.ErrCode

	if resp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
		return
	}

	resp.CategoryList = make([]cms_api_struct.CategoryInfo, len(rpcResp.CategoryList))

	for index, info := range rpcResp.CategoryList {
		log.Debug("", "info: ", info)
		_ = utils.CopyStructFields(&resp.CategoryList[index], info)
		resp.CategoryList[index].CategoryID = info.CategoryID
		resp.CategoryList[index].State = int8(info.State)
		resp.CategoryList[index].SortPriority = info.SortPriority
		resp.CategoryList[index].Editor = info.Editor
		resp.CategoryList[index].UsedAmount = info.UsedAmount
		resp.CategoryList[index].CreateTime = info.CreateTime
		resp.CategoryList[index].Categories = info.Categories
	}

	resp.CategoriesNum = rpcResp.CategoriesNum

	resp.ResponsePagination = cms_api_struct.ResponsePagination{CurrentPage: int(rpcResp.Pagination.CurrentPage), ShowNumber: int(rpcResp.Pagination.ShowNumber)}

	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func AddCategory(c *gin.Context) {

	var (
		req   cms_api_struct.AddCategoryReq
		resp  cms_api_struct.AddCategoryResp
		reqPb pb.AddCategoryReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	resp.CommRespCode = cms_api_struct.CommRespCode{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "The request params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.AddCategory(context.Background(), &reqPb)
	if err != nil {
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	if resp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func EditCategory(c *gin.Context) {

	var (
		req   cms_api_struct.EditCategoryReq
		resp  cms_api_struct.EditCategoryResp
		reqPb pb.EditCategoryReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}

	resp.CommRespCode = cms_api_struct.CommRespCode{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		errMsg := reqPb.OperationID + "The request params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.EditCategory(context.Background(), &reqPb)
	if err != nil {
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	if resp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
}

func SetCategoryStatus(c *gin.Context) {

	var (
		req   cms_api_struct.SetCategoryStatusReq
		resp  cms_api_struct.SetCategoryStatusResp
		reqPb pb.SetCategoryStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}

	resp.CommRespCode = cms_api_struct.CommRespCode{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "The request params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.SetCategoryStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.ErrCode = rpcResp.CommonResp.ErrCode

	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})

}

func DeleteCategory(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteCategoryReq
		resp  cms_api_struct.DeleteCategoryResp
		reqPb pb.DeleteCategoryReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}

	resp.CommRespCode = cms_api_struct.CommRespCode{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		errMsg := reqPb.OperationID + "The request params error"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	reqPb.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewGameStoreClient(etcdConn)
	rpcResp, err := client.DeleteCategory(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	resp.ErrMsg = rpcResp.CommonResp.ErrMsg
	resp.ErrCode = rpcResp.CommonResp.ErrCode

	c.JSON(http.StatusOK, gin.H{"code": resp.ErrCode, "err_msg": resp.ErrMsg, "data": resp})
}
