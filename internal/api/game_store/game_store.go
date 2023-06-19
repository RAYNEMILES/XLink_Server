package game_store

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbGameStore "Open_IM/pkg/proto/game_store"
	rpc "Open_IM/pkg/proto/game_store"
	sdk_ws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func BannerGames(c *gin.Context) {
	var (
		params api.BannerGamesReq
		resp api.BannerGamesResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	resp.CommRespCode = api.CommRespCode{}
	rpcReq := &pbGameStore.BannerGamesReq{}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.BannerGames(context.Background(), rpcReq)
	if err != nil {
		errMsg := "BannerGames failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.BannerGameList = []api.BannerGameInfo{}
	_ = utils.CopyStructFields(&resp.BannerGameList, rpcResp.BannerGameList)

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)

}

func TodayRecommendations(c *gin.Context) {
	var (
		params api.TodayRecommendationsReq
		resp api.TodayRecommendationsResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	resp.CommRespCode = api.CommRespCode{}

	rpcReq := &pbGameStore.TodayRecommendationsReq{Pagination: &sdk_ws.RequestPagination{PageNumber: params.PageNumber, ShowNumber: params.ShowNumber}}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.TodayRecommendations(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.RecommendationGameList = []api.GameListInfo{}
	_ = utils.CopyStructFields(&resp.RecommendationGameList, rpcResp.RecommendationGameList)

	resp.ShowNumber = params.ShowNumber
	resp.CurrentPage = params.PageNumber
	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)

}

func PopularGames(c *gin.Context) {
	var (
		params api.PopularGamesReq
		resp api.PopularGamesResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	rpcReq := &pbGameStore.PopularGamesReq{Pagination: &sdk_ws.RequestPagination{PageNumber: params.PageNumber, ShowNumber: params.ShowNumber}}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.PopularGames(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.PopularGameList = []api.GameListInfo{}
	resp.CommRespCode = api.CommRespCode{}
	_ = utils.CopyStructFields(&resp.PopularGameList, rpcResp.PopularGameList)

	resp.ShowNumber = params.ShowNumber
	resp.CurrentPage = params.PageNumber
	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)
}

func AllGames(c *gin.Context) {
	var (
		params api.AllGamesReq
		resp api.AllGamesResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	rpcReq := &pbGameStore.AllGamesReq{Pagination: &sdk_ws.RequestPagination{PageNumber: params.PageNumber, ShowNumber: params.ShowNumber}}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.AllGames(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.AllGameList = []api.GameListInfo{}
	resp.CommRespCode = api.CommRespCode{}
	_ = utils.CopyStructFields(&resp.AllGameList, rpcResp.AllGameList)

	resp.ShowNumber = params.ShowNumber
	resp.CurrentPage = params.PageNumber
	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)
}

func SearchName(c *gin.Context) {
	var (
		params api.SearchNameReq
		resp api.SearchNameResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	rpcReq := &pbGameStore.SearchNameReq{}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.SearchName(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.GameList = []*api.GameNameInfo{}
	resp.CommRespCode = api.CommRespCode{}

	_ = utils.CopyStructFields(&resp.GameList, rpcResp.GameList)

	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)
}

func SearchGameListByName(c *gin.Context) {
	var (
		params api.SearchGameListByNameReq
		resp api.SearchGameListByNameResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	rpcReq := &pbGameStore.SearchGameListByNameReq{Pagination: &sdk_ws.RequestPagination{PageNumber: params.PageNumber, ShowNumber: params.ShowNumber}}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.SearchGameListByName(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.GameList = []api.GameListInfo{}
	resp.CommRespCode = api.CommRespCode{}
	_ = utils.CopyStructFields(&resp.GameList, rpcResp.GameList)

	resp.ShowNumber = params.ShowNumber
	resp.CurrentPage = params.PageNumber
	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)
}

func GetCategories(c *gin.Context) {
	var (
		params api.GetCategoriesReq
		resp api.GetCategoriesResp
	)

	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}

	rpcReq := &pbGameStore.GetCategoriesReq{}
	_ = utils.CopyStructFields(&rpcReq, params)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.GetCategories(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	for _, category := range rpcResp.Categories {
		categoryResp := &api.Categories{Id: category.Id}
		_ = utils.CopyStructFields(&categoryResp.Categories, category.Detail)
		resp.Categories = append(resp.Categories, categoryResp)
	}

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)

}

func PlayGame(c *gin.Context) {
	var (
		params api.PlayGameReq
		resp api.PlayGameResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	//TODO Check authorized... or other check

	//
	rpcReq := &pbGameStore.PlayGameRecordReq{}
	rpcReq.GameCode = params.GameCode
	rpcReq.OperationID = params.OperationID
	rpcReq.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.PlayGameRecord(context.Background(), rpcReq)
	if err != nil {
		errMsg := "PlayGameRecord failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	c.JSON(http.StatusOK, resp)

}

func GetGameHistory(c *gin.Context) {
	var (
		params api.GetGameHistoryReq
		resp api.GetGameHistoryResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}


	rpcReq := &pbGameStore.GetHistoryReq{Pagination: &sdk_ws.RequestPagination{PageNumber: params.PageNumber, ShowNumber: params.ShowNumber}}
	_ = utils.CopyStructFields(&rpcReq, params)
	rpcReq.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.GetHistory(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.GameList = []api.GameListInfo{}
	resp.CommRespCode = api.CommRespCode{}
	_ = utils.CopyStructFields(&resp.GameList, rpcResp.GameList)

	resp.ShowNumber = params.ShowNumber
	resp.CurrentPage = params.PageNumber
	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)

}

func GetGameFavorites(c *gin.Context) {
	var (
		params api.GetGameFavoritesReq
		resp api.GetGameFavoritesResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	rpcReq := &pbGameStore.GetFavoritesReq{Pagination: &sdk_ws.RequestPagination{PageNumber: params.PageNumber, ShowNumber: params.ShowNumber}}
	_ = utils.CopyStructFields(&rpcReq, params)
	rpcReq.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.GetFavorites(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.GameList = []api.GameListInfo{}
	resp.CommRespCode = api.CommRespCode{}
	_ = utils.CopyStructFields(&resp.GameList, rpcResp.GameList)

	resp.ShowNumber = params.ShowNumber
	resp.CurrentPage = params.PageNumber
	resp.GameNums = rpcResp.GameNums

	resp.CommRespCode.ErrMsg = "fetch success"
	c.JSON(http.StatusOK, resp)

}

func RemoveGameFavorite(c *gin.Context) {
	var (
		params api.RemoveGameFavoriteReq
		resp api.RemoveGameFavoriteResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	rpcReq := &pbGameStore.RemoveGameFavoriteReq{}
	_ = utils.CopyStructFields(&rpcReq, params)
	rpcReq.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.RemoveGameFavorite(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	c.JSON(http.StatusOK, resp)

}

func AddGameFavorite(c *gin.Context) {
	var (
		params api.AddGameFavoriteReq
		resp api.AddGameFavoriteResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	rpcReq := &pbGameStore.AddGameFavoriteReq{}
	_ = utils.CopyStructFields(&rpcReq, params)
	rpcReq.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.AddGameFavorite(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	c.JSON(http.StatusOK, resp)

}

func GameDetails(c *gin.Context) {
	var (
		params api.GameDetailsReq
		resp api.GameDetailsResp
	)
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	rpcReq := &pbGameStore.GameDetailsReq{}
	_ = utils.CopyStructFields(&rpcReq, params)
	rpcReq.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGameStoreName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGameStoreClient(etcdConn)
	rpcResp, err := client.GameDetails(context.Background(), rpcReq)
	if err != nil {
		errMsg := "TodayRecommendations failed " + err.Error()
		log.NewError(rpcReq.OperationID, errMsg, rpcReq.String())
		resp.ErrMsg = errMsg
		resp.ErrCode = constant.ErrServer.ErrCode
		c.JSON(http.StatusOK, resp)
		return
	}

	log.Debug("", "rpcResp.SupportPlatform: ", rpcResp.SupportPlatform)

	_ = utils.CopyStructFields(&resp, rpcResp)
	_ = json.Unmarshal([]byte(rpcResp.SupportPlatform), &resp.PlatformLink)

	resp.CommRespCode.ErrCode = rpcResp.CommonResp.ErrCode
	resp.CommRespCode.ErrMsg = rpcResp.CommonResp.ErrMsg
	c.JSON(http.StatusOK, resp)

}
