package game_store

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbGameStore "Open_IM/pkg/proto/game_store"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type rpcGameStore struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewRpcGameStoreServer(port int) *rpcGameStore {
	return &rpcGameStore{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImGameStoreName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (rpc *rpcGameStore) Run() {
	log.NewPrivateLog(constant.OpenImGameStoreLog)
	log.NewInfo("0", "rpc moments start...")

	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(rpc.rpcPort)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + rpc.rpcRegisterName)
	}
	log.NewInfo("0", "listen network success, ", address, listener)
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd
	pbGameStore.RegisterGameStoreServer(srv, rpc)
	rpcRegisterIP := ""
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(),
			rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)
		return
	}
	log.NewInfo("0", "RegisterMomentsServer ok ", rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "rpc Moments ok")
}

func (rpc *rpcGameStore) BannerGames(_ context.Context, req *pbGameStore.BannerGamesReq) (*pbGameStore.BannerGamesResp, error) {
	resp := &pbGameStore.BannerGamesResp{CommonResp: &pbGameStore.CommonResp{}}

	where := make(map[string]interface{}, 0)
	where["classification"] = `["1"]`
	where["state"] = 1
	where["delete"] = 1

	bannerGames, _, err := imdb.GetGamesByWhere(where, 1, -1, "")
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.BannerGameList = []*pbGameStore.BannerGameInfo{}
	err = utils.CopyStructFields(&resp.BannerGameList, bannerGames)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}

	for index, game := range bannerGames {
		resp.BannerGameList[index].GameName = make(map[string]string, 2)
		resp.BannerGameList[index].GameName["en"] = game.GameNameEN
		resp.BannerGameList[index].GameName["cn"] = game.GameNameCN
	}

	return resp, nil
}

func (rpc *rpcGameStore) TodayRecommendations(_ context.Context, req *pbGameStore.TodayRecommendationsReq) (*pbGameStore.TodayRecommendationsResp, error) {
	resp := &pbGameStore.TodayRecommendationsResp{CommonResp: &pbGameStore.CommonResp{}}

	where := make(map[string]interface{}, 0)
	where["classification"] = `["2"]`
	where["categories"] = req.Categories
	where["state"] = 1
	where["delete"] = 1

	orderBy := "hot:DESC,create_time:DESC"

	recommendationGames, gameCounts, err := imdb.GetGamesByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, orderBy)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.RecommendationGameList = []*pbGameStore.GameListInfo{}
	err = utils.CopyStructFields(&resp.RecommendationGameList, recommendationGames)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}

	var categories = make([]string, 0)
	for index, recommendationGame := range recommendationGames {
		_ = json.Unmarshal([]byte(recommendationGame.Categories), &categories)
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}

		var categoriesStructList []*pbGameStore.CategoryMultiLanguage
		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}

		resp.RecommendationGameList[index].GameName = make(map[string]string, 2)
		resp.RecommendationGameList[index].GameName["en"] = recommendationGame.GameNameEN
		resp.RecommendationGameList[index].GameName["cn"] = recommendationGame.GameNameCN

		resp.RecommendationGameList[index].Categories = categoriesStructList
	}

	resp.GameNums = int32(gameCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcGameStore) PopularGames(_ context.Context, req *pbGameStore.PopularGamesReq) (*pbGameStore.PopularGamesResp, error) {
	resp := &pbGameStore.PopularGamesResp{}

	where := make(map[string]interface{}, 0)
	where["categories"] = req.Categories
	where["state"] = 1
	where["delete"] = 1

	orderBy := "click_counts:DESC,create_time:DESC"

	games, gameCounts, err := imdb.GetGamesByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, orderBy)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.PopularGameList = []*pbGameStore.GameListInfo{}
	err = utils.CopyStructFields(&resp.PopularGameList, games)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}

	var categories = make([]string, 0)
	for index, game := range games {
		_ = json.Unmarshal([]byte(game.Categories), &categories)
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}

		var categoriesStructList []*pbGameStore.CategoryMultiLanguage
		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}

		resp.PopularGameList[index].GameName = make(map[string]string, 2)
		resp.PopularGameList[index].GameName["en"] = game.GameNameEN
		resp.PopularGameList[index].GameName["cn"] = game.GameNameCN

		resp.PopularGameList[index].Categories = categoriesStructList
	}

	resp.GameNums = int32(gameCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil

}

func (rpc *rpcGameStore) AllGames(_ context.Context, req *pbGameStore.AllGamesReq) (*pbGameStore.AllGamesResp, error) {
	resp := &pbGameStore.AllGamesResp{CommonResp: &pbGameStore.CommonResp{}}

	where := make(map[string]interface{}, 0)
	where["categories"] = req.Categories
	where["state"] = 1
	where["delete"] = 1

	orderBy := "priority:DESC,create_time:DESC"

	games, gameCounts, err := imdb.GetGamesByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, orderBy)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.AllGameList = []*pbGameStore.GameListInfo{}
	err = utils.CopyStructFields(&resp.AllGameList, games)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}
	var categories = make([]string, 0)
	for index, game := range games {
		_ = json.Unmarshal([]byte(game.Categories), &categories)
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}

		var categoriesStructList []*pbGameStore.CategoryMultiLanguage

		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}

		resp.AllGameList[index].GameName = make(map[string]string, 2)
		resp.AllGameList[index].GameName["en"] = game.GameNameEN
		resp.AllGameList[index].GameName["cn"] = game.GameNameCN

		resp.AllGameList[index].Categories = categoriesStructList
	}

	resp.GameNums = int32(gameCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcGameStore) SearchName(_ context.Context, req *pbGameStore.SearchNameReq) (*pbGameStore.SearchNameResp, error) {
	resp := &pbGameStore.SearchNameResp{CommonResp: &pbGameStore.CommonResp{}}

	where := make(map[string]interface{}, 0)
	where["game_name"] = req.Name
	where["name_type"] = req.NameType
	where["state"] = 1
	where["delete"] = 1

	orderBy := ""
	if req.NameType == "en" {
		orderBy += "LENGTH(game_name_en):ASC,"
	} else if req.NameType == "cn" {
		orderBy += "LENGTH(game_name_cn):ASC,"
	} else {
		orderBy += "LENGTH(game_name_en):ASC,LENGTH(game_name_cn):ASC,"
	}
	orderBy += "create_time:DESC"

	games, gameCounts, err := imdb.GetGamesByWhere(where, 1, req.GetCount, orderBy)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.GameList = []*pbGameStore.GameNameInfo{}
	err = utils.CopyStructFields(&resp.GameList, games)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}

	for index, g := range games {
		resp.GameList[index].GameName = make(map[string]string, 2)
		resp.GameList[index].GameName["en"] = g.GameNameEN
		resp.GameList[index].GameName["cn"] = g.GameNameCN
	}

	resp.GameNums = int32(gameCounts)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcGameStore) SearchGameListByName(_ context.Context, req *pbGameStore.SearchGameListByNameReq) (*pbGameStore.SearchGameListByNameResp, error) {
	resp := &pbGameStore.SearchGameListByNameResp{CommonResp: &pbGameStore.CommonResp{}}

	where := make(map[string]interface{}, 0)
	where["game_name"] = req.Name
	where["categories"] = req.Categories
	where["name_type"] = req.NameType
	where["state"] = 1
	where["delete"] = 1

	orderBy := "priority:DESC,create_time:DESC"

	games, gameCounts, err := imdb.GetGamesByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, orderBy)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.GameList = []*pbGameStore.GameListInfo{}
	err = utils.CopyStructFields(&resp.GameList, games)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}

	var categories = make([]string, 0)
	for index, game := range games {
		_ = json.Unmarshal([]byte(game.Categories), &categories)
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}

		var categoriesStructList []*pbGameStore.CategoryMultiLanguage
		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}

		resp.GameList[index].GameName = make(map[string]string, 2)
		resp.GameList[index].GameName["en"] = game.GameNameEN
		resp.GameList[index].GameName["cn"] = game.GameNameCN
		resp.GameList[index].Categories = categoriesStructList
	}

	resp.GameNums = int32(gameCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcGameStore) GetCategories(_ context.Context, req *pbGameStore.GetCategoriesReq) (*pbGameStore.GetCategoriesResp, error) {
	resp := &pbGameStore.GetCategoriesResp{CommonResp: &pbGameStore.CommonResp{}}
	list, err := imdb.GetCategoriesList()
	if err != nil {
		errMsg := "GetCategoriesList error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetCategoriesList error" + err.Error()
		return resp, err
	}

	resp.Categories = []*pbGameStore.Categories{}
	for _, categories := range list {
		category := &pbGameStore.Categories{Id: categories.Id, Detail: map[string]string{
			"en": categories.CategoryNameEN,
			"cn": categories.CategoryNameCN,
		}}
		resp.Categories = append(resp.Categories, category)
	}

	resp.CommonResp.ErrMsg = "Get categories success"
	return resp, nil

}

func (rpc *rpcGameStore) PlayGameRecord(_ context.Context, req *pbGameStore.PlayGameRecordReq) (*pbGameStore.PlayGameRecordResp, error) {

	resp := &pbGameStore.PlayGameRecordResp{CommonResp: &pbGameStore.CommonResp{}}

	game, err := imdb.GetGameByGameCode(req.GameCode)
	if err != nil {
		errMsg := "GetGameByGameCode error"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "get game error"
		return resp, err
	}
	if game.Id == 0 {
		errMsg := "UpdateGameV2 error"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = "The game has expired"
		return resp, err
	}

	game = &db.Game{GameCode: req.GameCode, PlayCounts: 1}
	row, err := imdb.UpdateGameV2(game)
	if row == 0 || err != nil {
		errMsg := "UpdateGameV2 error"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "UpdateGameV2 error"
		return resp, err
	}

	gameHistory := &db.GamePlayHistory{GameCode: req.GameCode, UserID: req.UserID}
	err = imdb.InsertGameHistory(gameHistory)
	if err != nil {
		errMsg := "InsertGameHistory error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "InsertGameHistory error" + err.Error()
		return resp, err
	}

	resp.CommonResp.ErrMsg = "play pass success"
	return resp, nil
}

func (rpc *rpcGameStore) GetHistory(_ context.Context, req *pbGameStore.GetHistoryReq) (*pbGameStore.GetHistoryResp, error) {
	resp := &pbGameStore.GetHistoryResp{CommonResp: &pbGameStore.CommonResp{}}

	games, err := imdb.GetUserHistoriesGameList(req.UserID, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.GameList = []*pbGameStore.GameListInfo{}
	err = utils.CopyStructFields(&resp.GameList, games)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}
	history := &db.GamePlayHistory{}
	gameCounts, err := imdb.GetUserHistoryCounts(history, req.UserID)
	if err != nil {
		errMsg := "GetGameCountsByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGameCountsByWhere error" + err.Error()
		return resp, err
	}
	var categories = make([]string, 0)
	for index, game := range games {
		_ = json.Unmarshal([]byte(game.Categories), &categories)
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}

		var categoriesStructList []*pbGameStore.CategoryMultiLanguage
		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}

		resp.GameList[index].GameName = make(map[string]string, 2)
		resp.GameList[index].GameName["en"] = game.GameNameEN
		resp.GameList[index].GameName["cn"] = game.GameNameCN
		resp.GameList[index].Categories = categoriesStructList
		resp.GameList[index].DeleteTime = game.DeleteTime
		resp.GameList[index].CreateTime = game.CreateTime
	}

	resp.GameNums = int32(gameCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp.ErrMsg = "get history success"
	return resp, nil
}

func (rpc *rpcGameStore) GetFavorites(_ context.Context, req *pbGameStore.GetFavoritesReq) (*pbGameStore.GetFavoritesResp, error) {
	resp := &pbGameStore.GetFavoritesResp{CommonResp: &pbGameStore.CommonResp{}}

	games, err := imdb.GetUserFavoritesGameList(req.UserID, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		errMsg := "GetGamesByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGamesByWhere error" + err.Error()
		return resp, err
	}
	resp.GameList = []*pbGameStore.GameListInfo{}
	err = utils.CopyStructFields(&resp.GameList, games)
	if err != nil {
		errMsg := "CopyStructFields error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "CopyStructFields error" + err.Error()
		return resp, err
	}
	favorite := &db.GameFavorites{}
	gameCounts, err := imdb.GetUserFavoritesCounts(favorite, req.UserID)
	if err != nil {
		errMsg := "GetGameCountsByWhere error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetGameCountsByWhere error" + err.Error()
		return resp, err
	}
	var categories = make([]string, 0)
	for index, game := range games {
		_ = json.Unmarshal([]byte(game.Categories), &categories)
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}

		var categoriesStructList []*pbGameStore.CategoryMultiLanguage
		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}

		resp.GameList[index].GameName = map[string]string{
			"en": game.GameNameEN,
			"cn": game.GameNameCN,
		}
		resp.GameList[index].DeleteTime = game.DeleteTime
		resp.GameList[index].Categories = categoriesStructList
		resp.GameList[index].CreateTime = game.CreateTime
	}

	resp.GameNums = int32(gameCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	resp.CommonResp.ErrMsg = "get favorites success"
	return resp, nil
}

func (rpc *rpcGameStore) RemoveGameFavorite(_ context.Context, req *pbGameStore.RemoveGameFavoriteReq) (*pbGameStore.RemoveGameFavoriteResp, error) {
	resp := &pbGameStore.RemoveGameFavoriteResp{CommonResp: &pbGameStore.CommonResp{}}

	favorite := &db.GameFavorites{UserID: req.UserID, GameCode: req.GameCode}

	row, err := imdb.RemoveGameFavorite(favorite)
	if err != nil {
		errMsg := "RemoveGameFavorite error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "RemoveGameFavorite error" + err.Error()
		return resp, err
	}
	if row == 0 {
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = "Don't have this favorites"
		return resp, err
	}

	resp.CommonResp.ErrMsg = "Remove success"

	return resp, nil
}

func (rpc *rpcGameStore) AddGameFavorite(_ context.Context, req *pbGameStore.AddGameFavoriteReq) (*pbGameStore.AddGameFavoriteResp, error) {
	resp := &pbGameStore.AddGameFavoriteResp{CommonResp: &pbGameStore.CommonResp{}}

	favored, err := imdb.HasFavored(req.UserID, req.GameCode)
	if err != nil {
		errMsg := "HasFavored error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	if favored {
		errMsg := "this game has been favored"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	err = imdb.AddGameFavorite(&db.GameFavorites{UserID: req.UserID, GameCode: req.GameCode})
	if err != nil {
		errMsg := "AddGameFavorite error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	resp.CommonResp.ErrMsg = "Add favorite success"

	return resp, nil
}

func (rpc *rpcGameStore) GameDetails(_ context.Context, req *pbGameStore.GameDetailsReq) (*pbGameStore.GameDetailsResp, error) {
	resp := &pbGameStore.GameDetailsResp{CommonResp: &pbGameStore.CommonResp{}}

	game, err := imdb.GetGameByGameCode(req.GameCode)
	if err != nil {
		errMsg := "GetGameByGameCode error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "Get game by gameCode error" + err.Error()
		return resp, err
	}

	countGame := &db.Game{GameCode: game.GameCode, ClickCounts: 1}
	row, err := imdb.UpdateGameV2(countGame)
	if row == 0 || err != nil {
		errMsg := "The game isn't exist"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	_ = utils.CopyStructFields(resp, game)

	// horizontal cover
	if game.HorizontalCover != "" {
		err = json.Unmarshal([]byte(game.HorizontalCover), &resp.HorizontalCover)
		if err != nil {
			errMsg := "HorizontalCover json error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
			resp.CommonResp.ErrMsg = "HorizontalCover json error" + err.Error()
			return resp, err
		}
	}

	// categories
	var categories = make([]string, 0)
	_ = json.Unmarshal([]byte(game.Categories), &categories)
	if len(categories) > 0 {
		categoriesList, err := imdb.GetCategoriesByIdList(categories)
		if err != nil {
			errMsg := "GetCategoriesByIdList error" + err.Error()
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "GetCategoriesByIdList error" + err.Error()
			return resp, err
		}
		var categoriesStructList []*pbGameStore.CategoryMultiLanguage
		for _, gameCategories := range categoriesList {
			category := &pbGameStore.CategoryMultiLanguage{Category: map[string]string{
				"en": gameCategories.CategoryNameEN,
				"cn": gameCategories.CategoryNameCN,
			}}
			categoriesStructList = append(categoriesStructList, category)
		}
		resp.Categories = categoriesStructList
	}

	// game name
	resp.GameName = map[string]string{
		"en": game.GameNameEN,
		"cn": game.GameNameCN,
	}

	// Description
	resp.Description = map[string]string{
		"en": game.DescriptionEN,
		"cn": game.DescriptionCN,
	}

	// Support platform
	gameLinks, err := imdb.GetGameLink(req.GameCode)
	if err != nil {
		errMsg := "GetGameLink error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "Game get error" + err.Error()
		return resp, err
	}

	linkMap := map[int32]interface{}{}
	for _, link := range gameLinks {
		linkMap[link.Platform] = struct {
			PackageURL string `json:"package_url"`
			PlayURL    string `json:"play_url"`
		}{
			PackageURL: link.PackageURL,
			PlayURL:    link.PlayURL,
		}
	}
	linkByte, err := json.Marshal(linkMap)
	if err != nil {
		errMsg := "link map json error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
		resp.CommonResp.ErrMsg = "link map json error" + err.Error()
		return resp, err
	}
	log.Debug("", "linkMap::", linkMap)

	resp.SupportPlatform = string(linkByte)

	resp.HasFavorite, err = imdb.HasFavored(req.UserID, req.GameCode)
	if err != nil {
		errMsg := "GetFavorite error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "GetFavorite error" + err.Error()
		return resp, err
	}

	resp.CommonResp.ErrMsg = "get game detail success"
	return resp, nil
}

func (rpc *rpcGameStore) GetGameList(_ context.Context, req *pbGameStore.GetGameListReq) (*pbGameStore.GetGameListResp, error) {
	resp := &pbGameStore.GetGameListResp{CommonResp: &pbGameStore.CommonResp{}, Pagination: &sdkws.ResponsePagination{}}

	gameCodeList := make([]string, 0)
	if req.Platform != 0 {
		var err error
		if req.Platform != 1 && req.Platform != 2 && req.Platform != 3 {
			errMsg := "Not support platform"
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, nil
		}
		gameCodeList, err = imdb.GetGameCodeListByPlatform(req.Platform)
		if err != nil {
			errMsg := "find by platform error"
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, nil
		}
	}

	if req.Platform != 0 && len(gameCodeList) == 0 {
		resp.CommonResp.ErrMsg = "No record"
	} else {
		where := map[string]interface{}{}
		where["game_code"] = req.GameCode
		where["name_type"] = req.NameType
		where["game_name"] = req.GameName
		if req.Categories != 0 {
			where["categories"] = []int64{req.Categories}
		}
		if req.Classification == 1 || req.Classification == 2 {
			classByte, _ := json.Marshal([]string{strconv.FormatInt(req.Classification, 10)})
			where["classification"] = string(classByte)
		}
		where["publisher"] = req.Publisher
		where["hot"] = req.Hot
		where["state"] = req.State
		where["start_time"] = req.StartTime
		where["end_time"] = req.EndTime
		where["platform_game_codes"] = gameCodeList
		where["state"] = -1
		where["delete"] = 1

		orderBy := "create_time:DESC"
		log.Debug("", "where: ", where)

		games, gameCount, err := imdb.GetGamesByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, orderBy)
		if err != nil {
			errMsg := "Fetch games error"
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, nil
		}
		resp.GameList = []*pbGameStore.GameBackgroundListInfo{}
		log.Debug("", " ", len(games))
		_ = utils.CopyStructFields(&resp.GameList, games)
		for index, game := range games {
			respGame := resp.GameList[index]
			respGame.SortPriority = game.Priority
			// game name
			respGame.GameName = map[string]string{
				"en": game.GameNameEN,
				"cn": game.GameNameCN,
			}

			// game description
			respGame.Description = map[string]string{
				"en": game.DescriptionEN,
				"cn": game.DescriptionCN,
			}

			// horizontal cover
			respGame.HorizontalCover = make([]string, 0)
			_ = json.Unmarshal([]byte(game.HorizontalCover), &respGame.HorizontalCover)
			// categories
			var categories []string
			_ = json.Unmarshal([]byte(game.Categories), &categories)

			log.Debug("", "game.Categories: ", game.Categories)
			log.Debug("", "categories: ", categories)
			categoriesList, _ := imdb.GetCategoriesByIdList(categories)
			respGame.Categories = make([]*pbGameStore.CategoryMultiLanguage, len(categoriesList))
			for index, gameCategories := range categoriesList {
				log.Debug("", "gameCategories: ", gameCategories)
				categoryID := strconv.FormatInt(gameCategories.Id, 10)
				respGame.Categories[index] = &pbGameStore.CategoryMultiLanguage{
					Category: map[string]string{
						"id": categoryID,
						"en": gameCategories.CategoryNameEN,
						"cn": gameCategories.CategoryNameCN,
					},
				}
			}
			// classification
			log.Debug("", "game.Classification: ", game.Classification)
			_ = json.Unmarshal([]byte(game.Classification), &respGame.Classification)
			if respGame.Classification == nil {
				respGame.Classification = []string{}
			}
			// support platform
			gameLinks, err := imdb.GetGameLink(game.GameCode)
			if err != nil {
				errMsg := "GetGameLink error" + err.Error()
				log.Error(req.OperationID, errMsg)
				resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
				resp.CommonResp.ErrMsg = "Game get error" + err.Error()
				return resp, nil
			}

			log.Debug("", "req.GameCode: ", game.GameCode)
			log.Debug("", "gameLinks: ", gameLinks)
			linkMap := map[int32]interface{}{}
			for _, link := range gameLinks {
				uploadTime := link.CreateTime
				if link.UpdateTime != 0 {
					uploadTime = link.UpdateTime
				}
				linkMap[link.Platform] = struct {
					PackageURL string `json:"package_url"`
					PlayURL    string `json:"play_url"`
					Size       int64  `json:"size"`
					FileName   string `json:"file_name"`
					UploadTime int64  `json:"upload_time"`
				}{
					PackageURL: link.PackageURL,
					PlayURL:    link.PlayURL,
					Size:       link.PackageSize,
					FileName:   link.PackageName,
					UploadTime: uploadTime,
				}
			}
			linkByte, err := json.Marshal(linkMap)
			if err != nil {
				errMsg := "link map json error" + err.Error()
				log.Error(req.OperationID, errMsg)
				resp.CommonResp.ErrCode = constant.ErrServer.ErrCode
				resp.CommonResp.ErrMsg = "link map json error" + err.Error()
				return resp, nil
			}
			respGame.SupportPlatform = string(linkByte)
		}

		resp.GameNum = gameCount
	}

	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	resp.Pagination.CurrentPage = req.Pagination.PageNumber
	resp.CommonResp.ErrMsg = "Get game list success"
	return resp, nil
}

func (rpc *rpcGameStore) EditGame(_ context.Context, req *pbGameStore.EditGameReq) (*pbGameStore.EditGameResp, error) {
	resp := &pbGameStore.EditGameResp{CommonResp: &pbGameStore.CommonResp{}}

	game, err := imdb.GetGameAllByGameCode(req.Game.GameCode)
	if err != nil || game.GameCode == "" {
		errMsg := "This game isn't exist."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	game.GameCode = req.Game.GameCode
	// verify if the game info can be updated.
	if err = verifyParam(req.Game, game); err != nil {
		log.Error(req.OperationID, err.Error())
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = err.Error()
		return resp, nil
	}

	// platform
	var gameLinks []*db.GameLink
	var platformLink = make(map[int32]struct {
		PackageURL string `json:"package_url"`
		PlayURL    string `json:"play_url"`
		FileName   string `json:"file_name"`
		Size       int64  `json:"size"`
	}, 0)
	err = json.Unmarshal([]byte(req.Game.SupportPlatform), &platformLink)
	if err != nil {
		log.Error(req.OperationID, "platform param error"+err.Error())
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "platform param error" + err.Error()
		return resp, nil
	}
	if len(platformLink) != 0 {
		nowTime := time.Now().Unix()
		for k, v := range platformLink {
			gl := &db.GameLink{
				GameCode:    game.GameCode,
				Platform:    k,
				PackageURL:  v.PackageURL,
				PackageName: v.FileName,
				PlayURL:     v.PlayURL,
				PackageSize: v.Size,
				CreateUser:  req.OpUserID,
			}
			if v.Size != 0 {
				gl.CreateTime = nowTime
			}
			gameLinks = append(gameLinks, gl)
		}
	} else {
		log.Error(req.OperationID, "you must select support platform")
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "you must select support platform"
		return resp, nil
	}

	game.Publisher = req.Game.Publisher
	game.UpdateBy = req.OpUserID
	var row int64
	if row, err = imdb.UpdateGame(game); err != nil {
		errMsg := "db error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "update error, record isn't exist"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	gameLink := &db.GameLink{}
	_, err = imdb.RealDeleteGameLinksByGameCode(gameLink, game.GameCode)
	if err != nil {
		errMsg := "db error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	row, err = imdb.InsertGameLinks(gameLinks)
	if err != nil {
		errMsg := "db error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "insert game error"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	resp.CommonResp.ErrMsg = "Edit game success"
	return resp, nil
}

func (rpc *rpcGameStore) AddGame(_ context.Context, req *pbGameStore.AddGameReq) (*pbGameStore.AddGameResp, error) {
	resp := &pbGameStore.AddGameResp{CommonResp: &pbGameStore.CommonResp{}}

	game := &db.Game{GameCode: req.Game.GameCode}

	queryGame, err := imdb.GetGameAllByGameCode(req.Game.GameCode)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = err.Error()
		return resp, nil
	}

	if queryGame.GameCode != "" {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "the game has been exist"
		return resp, nil
	}

	// verify if the game info can be updated.
	if err = verifyParam(req.Game, game); err != nil {
		log.Error(req.OperationID, err.Error())
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = err.Error()
		return resp, nil
	}

	// platform
	var gameLinks []*db.GameLink
	var platformLink = make(map[int32]struct {
		PackageURL string `json:"package_url"`
		PlayURL    string `json:"play_url"`
		FileName   string `json:"file_name"`
		Size       int64  `json:"size"`
	}, 0)
	err = json.Unmarshal([]byte(req.Game.SupportPlatform), &platformLink)
	if err != nil {
		log.Error(req.OperationID, "platform param error"+err.Error())
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "platform param error" + err.Error()
		return resp, nil
	}
	if len(platformLink) != 0 {
		nowTime := time.Now().Unix()
		for k, v := range platformLink {
			gl := &db.GameLink{
				GameCode:    game.GameCode,
				Platform:    k,
				PackageURL:  v.PackageURL,
				PackageName: v.FileName,
				PlayURL:     v.PlayURL,
				PackageSize: v.Size,
				CreateUser:  req.OpUserID,
			}
			if v.Size != 0 {
				gl.CreateTime = nowTime
			}
			gameLinks = append(gameLinks, gl)
		}
	} else {
		log.Error(req.OperationID, "you must select support platform")
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "you must select support platform"
		return resp, nil
	}

	game.Publisher = req.Game.Publisher
	game.UpdateBy = req.OpUserID
	var row int64
	if row, err = imdb.AddGame(game); err != nil {
		errMsg := "db error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "Add error"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	gameLink := &db.GameLink{}
	_, err = imdb.RealDeleteGameLinksByGameCode(gameLink, game.GameCode)
	if err != nil {
		errMsg := "db error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	row, err = imdb.InsertGameLinks(gameLinks)
	if err != nil {
		errMsg := "db error" + err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "insert game error"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	resp.CommonResp.ErrMsg = "Add game success"
	return resp, nil
}

func (rpc *rpcGameStore) DeleteGames(_ context.Context, req *pbGameStore.DeleteGamesReq) (*pbGameStore.DeleteGamesResp, error) {

	resp := &pbGameStore.DeleteGamesResp{CommonResp: &pbGameStore.CommonResp{}}
	row, err := imdb.DeleteGamesByCodes(req.GamesCode, req.OpUserID)
	if err != nil {
		errMsg := "Delete games failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "Delete games failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	resp.CommonResp.ErrMsg = "delete games success"
	return resp, nil

}

func (rpc *rpcGameStore) AddCategory(_ context.Context, req *pbGameStore.AddCategoryReq) (*pbGameStore.AddCategoryResp, error) {
	resp := &pbGameStore.AddCategoryResp{CommonResp: &pbGameStore.CommonResp{}}

	if len(req.CategoryName) == 0 {
		errMsg := "you must typing category name."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	gameCategory := &db.GameCategories{}
	for k, language := range req.CategoryName {
		if utf8.RuneCountInString(language) > 50 {
			errMsg := "the category max length is 50"
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, nil
		}
		if k == "en" {
			gameCategory.CategoryNameEN = language
		} else if k == "cn" {
			gameCategory.CategoryNameCN = language
		} else {
			errMsg := "the language " + k + " isn't support"
			return resp, errors.New(errMsg)
		}
	}

	// priority
	if req.SortPriority >= 0 && req.SortPriority <= 100 {
		gameCategory.Priority = req.SortPriority
	} else {
		errMsg := "the sort priority must be between 0 and 100"
		return resp, errors.New(errMsg)
	}

	// state
	if req.State == 1 || req.State == 2 {
		gameCategory.Status = int8(req.State)
	} else {
		errMsg := "state must be 1 on or 2 off"
		return resp, errors.New(errMsg)
	}

	gameCategory.CreateUser = req.OpUserID
	row, err := imdb.InsertCategory(gameCategory)
	if err != nil {
		errMsg := err.Error()
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "Add category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	resp.CommonResp.ErrMsg = "Add category success"
	return resp, nil
}

func (rpc *rpcGameStore) GetCategory(_ context.Context, req *pbGameStore.GetCategoryReq) (*pbGameStore.GetCategoryResp, error) {
	resp := &pbGameStore.GetCategoryResp{CommonResp: &pbGameStore.CommonResp{}, Pagination: &sdkws.ResponsePagination{}}

	//var editorUsers []db.User
	//var creatorUsers []db.User
	// get editor list
	//if req.Editor != "" {
	//	editorUsers = imdb.GetUserByAllCondition(req.Editor)
	//}
	//if req.Creator != "" {
	//	creatorUsers = imdb.GetUserByAllCondition(req.Creator)
	//}
	//if (req.Editor != "" && len(editorUsers) == 0) || (req.Creator != "" && len(creatorUsers) == 0) {
	//	errMsg := "Not record"
	//	log.Error(req.OperationID, errMsg)
	//	resp.CommonResp.ErrMsg = errMsg
	//	return resp, nil
	//} else {
	//
	//}

	//var editorUserIdList []string
	//var editorUserMap = make(map[string]*db.User, 0)
	//
	//var creatorUserIdList []string
	//var creatorUserMap = make(map[string]*db.User, 0)

	//if req.Editor != "" {
	//	for index, user := range editorUsers {
	//		editorUserIdList = append(editorUserIdList, user.UserID)
	//		editorUserMap[user.UserID] = &editorUsers[index]
	//	}
	//}
	//if req.Creator != "" {
	//	for index, user := range creatorUsers {
	//		creatorUserIdList = append(creatorUserIdList, user.UserID)
	//		creatorUserMap[user.UserID] = &creatorUsers[index]
	//	}
	//}
	where := map[string]interface{}{}
	where["category_name_type"] = req.NameType
	where["category_name"] = req.CategoryName
	//if req.Editor != "" {
	//	where["editor"] = editorUserIdList
	//}
	//if req.Creator != "" {
	//	where["creator"] = creatorUserIdList
	//}
	where["editor"] = req.Editor
	where["creator"] = req.Creator
	where["create_start_time"] = req.CreateStartTime
	where["create_end_time"] = req.CreateEndTime
	where["edited_start_time"] = req.EditedStartTime
	where["edited_end_time"] = req.EditedEndTime
	where["state"] = req.State

	orderBy := "create_time:DESC"

	categories, categoryCount, err := imdb.GetCategoriesByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, orderBy)
	if err != nil {
		errMsg := "Get category failed"
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, err
	}

	resp.CategoryList = []*pbGameStore.CategoryDetailInfo{}

	_ = utils.CopyStructFields(&resp.CategoryList, categories)
	for index, category := range categories {
		resp.CategoryList[index].CategoryID = category.Id
		resp.CategoryList[index].UsedAmount, _ = imdb.GetCategoryUsedAmount(category.Id)
		resp.CategoryList[index].Creator = category.CreateUser
		resp.CategoryList[index].Editor = category.UpdateUser
		resp.CategoryList[index].EditTime = category.UpdateTime
		resp.CategoryList[index].State = int32(category.Status)
		resp.CategoryList[index].SortPriority = category.Priority

		resp.CategoryList[index].Categories = map[string]string{
			"en": category.CategoryNameEN,
			"cn": category.CategoryNameCN,
		}
	}

	resp.CategoriesNum = categoryCount

	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	resp.Pagination.CurrentPage = req.Pagination.PageNumber
	resp.CommonResp.ErrMsg = "fetch success"

	return resp, nil
}

func (rpc *rpcGameStore) EditCategory(_ context.Context, req *pbGameStore.EditCategoryReq) (*pbGameStore.EditCategoryResp, error) {
	resp := &pbGameStore.EditCategoryResp{CommonResp: &pbGameStore.CommonResp{}}

	if len(req.CategoryName) == 0 {
		errMsg := "you must typing category name."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	gameCategory := &db.GameCategories{}
	gameCategory.Id = req.CategoryID
	for k, language := range req.CategoryName {
		if utf8.RuneCountInString(language) > 50 {
			errMsg := "the category max length is 50"
			log.Error(req.OperationID, errMsg)
			resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
			resp.CommonResp.ErrMsg = errMsg
			return resp, nil
		}
		if k == "en" {
			gameCategory.CategoryNameEN = language
		} else if k == "cn" {
			gameCategory.CategoryNameCN = language
		} else {
			errMsg := "the language " + k + " isn't support"
			return resp, errors.New(errMsg)
		}
	}

	// priority
	if req.SortPriority >= 0 && req.SortPriority <= 100 {
		gameCategory.Priority = req.SortPriority
	} else {
		errMsg := "the sort priority must be between 0 and 100"
		return resp, errors.New(errMsg)
	}

	// state
	if req.State == 1 || req.State == 2 {
		gameCategory.Status = int8(req.State)
	} else {
		errMsg := "state must be 1 on or 2 off"
		return resp, errors.New(errMsg)
	}

	gameCategory.UpdateUser = req.OpUserID
	row, err := imdb.UpdateCategory(gameCategory)
	if err != nil {
		errMsg := "Edit category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "Edit category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	resp.CommonResp.ErrMsg = "Update category success"
	return resp, nil
}

func (rpc *rpcGameStore) SetCategoryStatus(_ context.Context, req *pbGameStore.SetCategoryStatusReq) (*pbGameStore.SetCategoryStatusResp, error) {
	resp := &pbGameStore.SetCategoryStatusResp{CommonResp: &pbGameStore.CommonResp{}}

	category := &db.GameCategories{}
	// state
	if req.State == 1 || req.State == 2 {
		category.Status = int8(req.State)
	} else {
		errMsg := "state must be 1 on or 2 off"
		return resp, errors.New(errMsg)
	}
	category.Id = req.CategoryID
	category.UpdateUser = req.OpUserID
	row, err := imdb.UpdateCategoryStatus(category)
	if err != nil {
		errMsg := "Update category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "Update category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}

	resp.CommonResp.ErrMsg = "Update category success"
	return resp, nil
}

func (rpc *rpcGameStore) DeleteCategory(_ context.Context, req *pbGameStore.DeleteCategoryReq) (*pbGameStore.DeleteCategoryResp, error) {
	resp := &pbGameStore.DeleteCategoryResp{CommonResp: &pbGameStore.CommonResp{}}
	row, err := imdb.DeleteCategoryById(req.CategoryID, req.OpUserID)
	if err != nil {
		errMsg := "Delete category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	if row == 0 {
		errMsg := "Delete category failed."
		log.Error(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	resp.CommonResp.ErrMsg = "delete category success"
	return resp, nil
}

func verifyParam(params *pbGameStore.GameBackgroundInfo, game *db.Game) error {

	// game name cover can't be empty, len < 50
	if len(params.GameName) != 0 {
		for k, language := range params.GameName {
			if utf8.RuneCountInString(language) > 50 {
				return errors.New("the game name max length is 50")
			}
			existed, err := imdb.HasGameByGameName(language, game.GameCode)
			if err != nil {
				return constant.ErrDB
			}
			if existed {
				errMsg := "the game " + language + " has existed"
				return errors.New(errMsg)
			}

			if k == "en" {
				game.GameNameEN = language
			} else if k == "cn" {
				game.GameNameCN = language
			} else {
				errMsg := "the language " + k + " isn't support"
				return errors.New(errMsg)
			}
		}
	} else {
		return errors.New("the game name can't be empty")
	}

	// description len < 500
	if len(params.Description) != 0 {
		for k, description := range params.Description {
			if utf8.RuneCountInString(description) > 500 {
				return errors.New("the game description max length is 500")
			}
			if k == "en" {
				game.DescriptionEN = description
			} else if k == "cn" {
				game.DescriptionCN = description
			} else {
				errMsg := "the language " + k + " isn't support"
				return errors.New(errMsg)
			}
		}
	}

	// horizontal cover can't be empty
	if len(params.HorizontalCover) != 0 {
		if len(params.HorizontalCover) > 5 {
			return errors.New("the horizontal cover page max count is 5")
		}
		coverByte, err := json.Marshal(params.HorizontalCover)
		if err != nil {
			return errors.New("the game horizontal cover format incorrect")
		}
		game.HorizontalCover = string(coverByte)
	} else {
		return errors.New("the game horizontal cover can't be empty")
	}

	// game list cover
	if len(params.CoverImg) != 0 {
		game.CoverImg = params.CoverImg
	} else {
		return errors.New("you must select a game cover image")
	}

	// categories
	if len(params.Categories) != 0 {
		categoriesByte, _ := json.Marshal(params.Categories)
		game.Categories = string(categoriesByte)
	} else {
		return errors.New("you must select at least one category")
	}

	// Hot
	if params.Hot > 0 && params.Hot <= 5 {
		game.Hot = params.Hot
	} else {
		return errors.New("you must set the game hot")
	}

	// classification
	classificationByte, err := json.Marshal(params.Classification)
	if err != nil {
		return err
	}
	game.Classification = string(classificationByte)

	// priority
	if params.SortPriority >= 0 && params.SortPriority <= 100 {
		game.Priority = params.SortPriority
	} else {
		return errors.New("the sort priority must be between 0 and 100")
	}

	// state
	if params.State == 1 || params.State == 2 {
		game.State = int8(params.State)
	} else {
		return errors.New("state must be 1 on or 2 off")
	}

	return nil

}
