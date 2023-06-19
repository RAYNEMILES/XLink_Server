package favorite

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/office"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AddFavorite(c *gin.Context) {
	params := api.AddFavoriteReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	contentByte, err := json.Marshal(params.Content)
	if err != nil {
		errMsg := "AddFavorite failed, copy struct" + err.Error()
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	req := &rpc.AddFavoriteReq{}
	err = utils.CopyStructFields(req, params)
	if err != nil {
		errMsg := "AddFavorite failed, copy struct" + err.Error()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	req.Content = string(contentByte)
	req.UserID = userID
	log.NewInfo(params.OperationID, "AddFavorite args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, req.OperationID)
	client := rpc.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.AddFavorite(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "AddFavorite failed ", err.Error())
		errMsg := "AddFavorite failed " + err.Error()
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	resp := api.AddFavoriteResp{CommResp: api.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "AddFavorite api return ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetFavoriteList(c *gin.Context) {
	params := api.GetFavoriteListReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	req := &rpc.GetFavoriteListReq{}
	err := utils.CopyStructFields(req, params)
	if err != nil {
		errMsg := "GetFavoriteList failed, copy struct" + err.Error()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	req.UserID = userID

	log.NewInfo(params.OperationID, "GetFavoriteList args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, req.OperationID)
	client := rpc.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.GetFavoriteList(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "GetFavoriteList failed ", err.Error())
		errMsg := "GetFavoriteList rpc failed, err:" + err.Error()
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	resp := api.GetFavoriteListResp{CommResp: api.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}, Favorites: []*api.Favorite{}}
	err = utils.CopyStructFields(&resp.Favorites, rpcResp.Favorites)
	if err != nil {
		errMsg := "GetFavoriteList failed, copy result struct" + err.Error()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	for index, _ := range resp.Favorites {
		contentMap := make(map[string]interface{}, 0)
		contentCreatorNameMap := make(map[string]interface{}, 0)
		err = json.Unmarshal([]byte(rpcResp.Favorites[index].Content), &contentMap)
		if err != nil {
			errMsg := "content json error: " + rpcResp.Favorites[index].Content
			log.NewError(req.OperationID, errMsg)
			continue
		}
		err = json.Unmarshal([]byte(rpcResp.Favorites[index].ContentCreatorName), &contentCreatorNameMap)
		if err != nil {
			errMsg := "content json error: " + rpcResp.Favorites[index].Content
			log.NewError(req.OperationID, errMsg)
			continue
		}

		resp.Favorites[index].Content = contentMap
		resp.Favorites[index].CreatorDetail = contentCreatorNameMap
	}

	log.NewInfo(req.OperationID, "GetFavoriteList api return ", resp)
	c.JSON(http.StatusOK, resp)

}

func RemoveFavorite(c *gin.Context) {
	params := api.RemoveFavoriteReq{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	userID := ""
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	req := &rpc.RemoveFavoriteReq{}
	err := utils.CopyStructFields(req, params)
	if err != nil {
		errMsg := "RemoveFavorite failed, copy struct" + err.Error()
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	req.UserID = userID

	log.NewInfo(params.OperationID, "RemoveFavorite args ", req.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfficeName, req.OperationID)
	client := rpc.NewOfficeServiceClient(etcdConn)
	rpcResp, err := client.RemoveFavorite(context.Background(), req)
	if err != nil {
		log.NewError(req.OperationID, "RemoveFavorite failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call add blacklist rpc server failed"})
		return
	}
	resp := api.AddFavoriteResp{CommResp: api.CommResp{ErrCode: rpcResp.CommonResp.ErrCode, ErrMsg: rpcResp.CommonResp.ErrMsg}}
	log.NewInfo(req.OperationID, "RemoveFavorite api return ", resp)
	c.JSON(http.StatusOK, resp)

}


