package news

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbNews "Open_IM/pkg/proto/news"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

// createNewsGRPCconnection create connection for API request
func createNewsGRPCconnection(OperationID string) (pbNews.NewsClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, OperationID)
	if etcdConn == nil {
		errMsg := "getcdv3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbNews.NewNewsClient(etcdConn)
	return client, nil
}

// RegisterOfficial register new official account
func RegisterOfficial(c *gin.Context) {
	var (
		req   api.OfficialRegisterRequest
		res   api.CommResp
		reqPb pbNews.RegisterOfficialRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	if len(req.Nickname) > 36 {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrNickNameLength.ErrCode, "errMsg": constant.ErrNickNameLength.ErrMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	reqPb.UserID = userIDInterface.(string)

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.RegisterOfficial(context.Background(), &reqPb)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

// GetSelfInfo get official user self info
func GetSelfInfo(c *gin.Context) {
	var (
		req   api.GetSelfOfficialInfoRequest
		res   api.GetSelfOfficialInfoResponse
		reqPb pbNews.GetSelfOfficialInfoRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	reqPb.UserID = userIDInterface.(string)
	reqPb.OperationID = req.OperationID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.GetSelfOfficialInfo(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetSelfOfficialInfo rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg
	if err = utils.CopyStructFields(&res.Data, respPb.Data); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func SetSelfInfo(c *gin.Context) {
	var (
		req   api.SetSelfOfficialInfoRequest
		res   api.CommResp
		reqPb pbNews.SetSelfOfficialInfoRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	reqPb.UserID = userIDInterface.(string)

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.SetSelfOfficialInfo(context.Background(), &reqPb)
	if err != nil {
		errMsg := "SetSelfOfficialInfo rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func FollowOfficial(c *gin.Context) {
	var (
		req   api.FollowOfficialRequest
		res   api.CommResp
		reqPb pbNews.FollowOfficialAccountRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.OfficialID = req.OfficialID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.FollowOfficialAccount(context.Background(), &reqPb)
	if err != nil {
		errMsg := "follow official account failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UnfollowOfficial(c *gin.Context) {
	var (
		req   api.UnfollowOfficialRequest
		res   api.CommResp
		reqPb pbNews.UnfollowOfficialAccountRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.OfficialID = req.OfficialID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.UnfollowOfficialAccount(context.Background(), &reqPb)
	if err != nil {
		errMsg := "unfollow official account failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UpdateOfficialFollowSettings(c *gin.Context) {
	var (
		req   api.UpdateOfficialFollowSettingsRequest
		res   api.CommResp
		reqPb pbNews.UpdateOfficialFollowSettingsRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.OfficialID = req.OfficialID
	reqPb.Muted = req.Muted
	reqPb.Enabled = req.Enabled

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.UpdateOfficialFollowSettings(context.Background(), &reqPb)
	if err != nil {
		errMsg := "update official follow settings failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func OfficialFollowSettingsByOfficialID(c *gin.Context) {
	var (
		req   api.OfficialFollowSettingsByOfficialIDRequest
		res   api.OfficialFollowSettingsByOfficialIDResponse
		reqPb pbNews.OfficialFollowSettingsByOfficialIDRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.ReqUserID = userIDInterface.(string)
	}
	reqPb.OfficialID = req.OfficialID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.GetOfficialFollowSettingsByOfficialID(context.Background(), &reqPb)
	if err != nil {
		errMsg := "update official follow settings failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.CommonResp.ErrCode
	res.ErrMsg = respPb.CommonResp.ErrMsg
	if respPb.UserFollow != nil {
		res.UserFollowedOfficialAccSetting = new(api.UserFollowedOfficialAccSetting)
		err = utils.CopyStructFields(res.UserFollowedOfficialAccSetting, respPb.UserFollow)
		if err != nil {
			log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "Copy UserFlow object have error : ", err.Error())
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UserFollowList(c *gin.Context) {
	var (
		req   api.UserFollowListRequest
		res   api.UserFollowListResponse
		reqPb pbNews.GetUserFollowListRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit
	reqPb.Keyword = req.Keyword

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.GetUserFollowList(context.Background(), &reqPb)
	if err != nil {
		errMsg := "follow official account failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		res.ErrCode = respPb.CommonResp.ErrCode
		res.ErrMsg = respPb.CommonResp.ErrMsg
	}

	res.Data = api.UserFollowListData{
		Count:   respPb.Count,
		Entries: make([]api.UserFollow, 0),
	}

	if respPb.Follows != nil {
		res.Data.Entries = make([]api.UserFollow, len(respPb.Follows))
		for i, follow := range respPb.Follows {
			res.Data.Entries[i] = api.UserFollow{
				OfficialID: follow.OfficialID,
				Nickname:   follow.Nickname,
				FaceURL:    follow.FaceURL,
				Bio:        follow.Bio,
				Type:       follow.Type,
				FollowTime: follow.FollowTime,
				Muted:      follow.Muted,
				Enabled:    follow.Enabled,
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func BlockOfficialFollows(c *gin.Context) {
	var (
		req   api.BlockOfficialFollowsRequest
		res   api.CommResp
		reqPb pbNews.BlockOfficialFollowsRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.OfficialUserID = userIDInterface.(string)
	}
	reqPb.UserIDList = req.UserIDList

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.BlockOfficialFollows(context.Background(), &reqPb)
	if err != nil {
		errMsg := "block official account follow failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UnblockOfficialFollows(c *gin.Context) {
	var (
		req   api.UnblockOfficialFollowsRequest
		res   api.CommResp
		reqPb pbNews.UnblockOfficialFollowsRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.OfficialUserID = userIDInterface.(string)
	}

	reqPb.UserIDList = req.UserIDList

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.UnblockOfficialFollows(context.Background(), &reqPb)
	if err != nil {
		errMsg := "unblock official account follow failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func DeleteOfficialFollows(c *gin.Context) {
	var (
		req   api.DeleteOfficialFollowsRequest
		res   api.CommResp
		reqPb pbNews.DeleteOfficialFollowsRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.OfficialUserID = userIDInterface.(string)
	}

	reqPb.UserIDList = req.UserIDList

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.DeleteOfficialFollows(context.Background(), &reqPb)
	if err != nil {
		errMsg := "delete official account follows failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func LikeArticle(c *gin.Context) {
	var (
		req   api.LikeArticleRequest
		res   api.CommResp
		reqPb pbNews.LikeArticleRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.ArticleID = req.ArticleID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.LikeArticle(context.Background(), &reqPb)
	if err != nil {
		errMsg := "like article failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UnlikeArticle(c *gin.Context) {
	var (
		req   api.UnlikeArticleRequest
		res   api.CommResp
		reqPb pbNews.UnlikeArticleRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.ArticleID = req.ArticleID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.UnlikeArticle(context.Background(), &reqPb)
	if err != nil {
		errMsg := "unlike failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func DeleteArticleLike(c *gin.Context) {
	var (
		req   api.DeleteArticleLikeRequest
		res   api.CommResp
		reqPb pbNews.DeleteArticleLikeRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.OpUserID = userIDInterface.(string)
	}

	reqPb.UserID = req.UserID
	reqPb.ArticleID = req.ArticleID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.DeleteArticleLike(context.Background(), &reqPb)
	if err != nil {
		errMsg := "delete article like failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func AddArticleComment(c *gin.Context) {
	var (
		req   api.AddArticleCommentRequest
		resp  api.AddArticleCommentResponse
		reqPb pbNews.AddArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInter.(string)
		reqPb.OpUserID = userIDInter.(string)
	}

	reqPb.ArticleID = req.ArticleID
	reqPb.ParentCommentID = req.ParentCommentID
	reqPb.ReplyUserID = req.ReplyUserID
	reqPb.ReplyOfficialID = req.ReplyOfficialID
	reqPb.Content = req.Content

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.AddArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "add article comment failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		resp.ErrMsg = respPb.CommonResp.ErrMsg
		resp.ErrCode = respPb.CommonResp.ErrCode
	}
	resp.Data = respPb.CommentID

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func LikeArticleComment(c *gin.Context) {
	var (
		req   api.LikeArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.LikeArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.LikeArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "LikeArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UnlikeArticleComment(c *gin.Context) {
	var (
		req   api.UnlikeArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.UnlikeArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.UnlikeArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "UnlikeArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func OfficialLikeArticleComment(c *gin.Context) {
	var (
		req   api.OfficialLikeArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.OfficialLikeArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.OfficialLikeArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "OfficialLikeArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func OfficialUnlikeArticleComment(c *gin.Context) {
	var (
		req   api.OfficialUnlikeArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.OfficialUnlikeArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.OfficialUnlikeArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "OfficialUnlikeArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func OfficialDeleteArticleComment(c *gin.Context) {
	var (
		req   api.OfficialDeleteArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.OfficialDeleteArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.OfficialDeleteArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "OfficialDeleteArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func OfficialShowArticleComment(c *gin.Context) {
	var (
		req   api.OfficialShowArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.OfficialShowArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.OfficialShowArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "OfficialShowArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func OfficialHideArticleComment(c *gin.Context) {
	var (
		req   api.OfficialHideArticleCommentRequest
		res   api.CommResp
		reqPb pbNews.OfficialHideArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.OfficialHideArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "OfficialHideArticleComment rpc failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func AddOfficialArticleComment(c *gin.Context) {
	var (
		req   api.AddArticleCommentRequest
		resp  api.AddArticleCommentResponse
		reqPb pbNews.AddArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		reqPb.OpUserID = userIDInter.(string)
	}

	reqPb.ArticleID = req.ArticleID
	reqPb.ParentCommentID = req.ParentCommentID
	reqPb.ReplyUserID = req.ReplyUserID
	reqPb.ReplyOfficialID = req.ReplyOfficialID
	reqPb.Content = req.Content

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.AddArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "add article comment failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		resp.ErrMsg = respPb.CommonResp.ErrMsg
		resp.ErrCode = respPb.CommonResp.ErrCode
	}
	resp.Data = respPb.CommentID

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func ListArticlesTimeLine(c *gin.Context) {
	var (
		req   api.ListArticlesTimeLineRequest
		reqPb pbNews.ListArticlesTimeLineRequest
		res   api.ListArticlesTimeLineResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.Source = req.Source
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit
	reqPb.OfficialID = req.OfficialID

	userIDInter, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInter.(string)
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListArticlesTimeLine(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListArticlesTimeLine rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
	}

	res.Data = &api.ListArticlesTimeLineData{
		Count:   resPb.Count,
		Entries: make([]api.ListArticlesTimeLineEntry, 0),
	}

	if resPb.Entries != nil {
		res.Data.Entries = make([]api.ListArticlesTimeLineEntry, len(resPb.Entries))
		if err = utils.CopyStructFields(&res.Data.Entries, resPb.Entries); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func CreateArticle(c *gin.Context) {
	var (
		req   api.CreateArticleReq
		reqPb pbNews.CreateArticleReq
		res   api.CommResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	reqPb.UserID = userIDInterface.(string)

	if err := utils.CopyStructFields(&reqPb, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.CreateArticle(context.Background(), &reqPb)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func UpdateArticle(c *gin.Context) {
	var (
		req   api.UpdateArticleReq
		reqPb pbNews.UpdateArticleReq
		res   api.CommResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	reqPb.UserID = userIDInterface.(string)

	if err := utils.CopyStructFields(&reqPb, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.UpdateArticle(context.Background(), &reqPb)
	if err != nil {
		errMsg := "UpdateArticle rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func DeleteArticle(c *gin.Context) {
	var (
		req   api.DeleteArticleReq
		reqPb pbNews.DeleteArticleReq
		res   api.CommResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	reqPb.UserID = userIDInterface.(string)

	if err := utils.CopyStructFields(&reqPb, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.DeleteArticle(context.Background(), &reqPb)
	if err != nil {
		errMsg := "DeleteArticle rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ListArticles(c *gin.Context) {
	var (
		req       api.ListArticlesReq
		listReqPb pbNews.ListOfficialArticlesReq
		res       api.ListArticlesResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, _ := c.Get("userID")
	listReqPb.UserID = userIDInterface.(string)

	if err := utils.CopyStructFields(&listReqPb, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	listResp, err := client.ListOfficialArticles(context.Background(), &listReqPb)
	if err != nil {
		errMsg := "ListOfficialArticles rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, listReqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = listResp.Count
	res.Data.Entries = make([]api.ArticleSummaryEntry, 0)
	if listResp.Articles != nil {
		res.Data.Entries = make([]api.ArticleSummaryEntry, len(listResp.Articles))
		for i, article := range listResp.Articles {
			if err = utils.CopyStructFields(&res.Data.Entries[i], article); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func GetOfficialArticle(c *gin.Context) {
	var (
		req   api.GetOfficialArticleReq
		reqPb pbNews.GetOfficialArticleReq
		res   api.GetOfficialArticleResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if err := utils.CopyStructFields(&reqPb, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	pbRes, err := client.GetOfficialArticle(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetOfficialArticle rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if pbRes.Article != nil {
		var article api.ArticleEntry
		if err = utils.CopyStructFields(&article, pbRes.Article); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
		res.Data = &article
	}

	if pbRes.CommonResp != nil {
		res.ErrCode = pbRes.CommonResp.ErrCode
		res.ErrMsg = pbRes.CommonResp.ErrMsg
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ListOfficialSelfFollows(c *gin.Context) {
	var (
		req   api.ListOfficialSelfFollowsRequest
		reqPb pbNews.ListSelfOfficialFollowsRequest
		res   api.ListOfficialSelfFollowsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, exists := c.Get("userID")
	if exists {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.MinFollowTime = req.MinFollowTime
	reqPb.MinBlockTime = req.MinBlockTime
	reqPb.BlockFilter = req.BlockFilter
	reqPb.OrderBy = req.OrderBy
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListSelfOfficialFollows(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListOfficialArticles rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = resPb.Count
	res.Data.Entries = make([]api.OfficialFollowEntry, 0)
	if resPb.Follows != nil {
		res.Data.Entries = make([]api.OfficialFollowEntry, len(resPb.Follows))
		for i, follow := range resPb.Follows {
			if err = utils.CopyStructFields(&res.Data.Entries[i], follow); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ListArticleLikes(c *gin.Context) {
	var (
		req   api.ListArticleLikesRequest
		reqPb pbNews.ListArticleLikesRequest
		res   api.ListArticleLikesResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.ArticleID = req.ArticleID
	reqPb.Keyword = req.Keyword
	reqPb.MinCreateTime = req.MinCreateTime
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListArticleLikes(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListArticleLikes rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = resPb.Count
	res.Data.Entries = make([]api.ArticleLikeEntry, 0)
	if resPb.Likes != nil {
		res.Data.Entries = make([]api.ArticleLikeEntry, len(resPb.Likes))
		for i, like := range resPb.Likes {
			if err = utils.CopyStructFields(&res.Data.Entries[i], like); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func copyArticleCommentEntry(comment *api.CommentEntry, newComment *pbNews.CommentEntry) error {
	if comment == nil || newComment == nil {
		return nil
	}

	if newComment.UserID != "" {
		comment.Nickname = newComment.UserNickname
		comment.FaceURL = newComment.UserFaceURL
	} else if newComment.OfficialID != 0 {
		comment.Nickname = newComment.OfficialNickname
		comment.FaceURL = newComment.OfficialFaceURL
	}

	if newComment.ReplyUserID != "" {
		comment.ReplyNickname = newComment.ReplyUserNickname
		comment.ReplyFaceURL = newComment.ReplyUserFaceURL
	} else if newComment.ReplyOfficialID != 0 {
		comment.ReplyNickname = newComment.ReplyOfficialNickname
		comment.ReplyFaceURL = newComment.ReplyOfficialFaceURL
	}

	return utils.CopyStructFields(&comment, newComment)
}

func copyArticleCommentEntryWithReplies(comment *api.CommentEntryWithReplies, newComment *pbNews.CommentEntry) error {
	if comment == nil || newComment == nil {
		return nil
	}

	if newComment.UserID != "" {
		comment.Nickname = newComment.UserNickname
		comment.FaceURL = newComment.UserFaceURL
	} else if newComment.OfficialID != 0 {
		comment.Nickname = newComment.OfficialNickname
		comment.FaceURL = newComment.OfficialFaceURL
	}

	if newComment.ReplyUserID != "" {
		comment.ReplyNickname = newComment.ReplyUserNickname
		comment.ReplyFaceURL = newComment.ReplyUserFaceURL
	} else if newComment.ReplyOfficialID != 0 {
		comment.ReplyNickname = newComment.ReplyOfficialNickname
		comment.ReplyFaceURL = newComment.ReplyOfficialFaceURL
	}

	return utils.CopyStructFields(&comment, newComment)
}

func ListArticleComments(c *gin.Context) {
	var (
		req   api.ListArticleCommentsRequest
		reqPb pbNews.ListArticleCommentsRequest
		res   api.ListArticleCommentsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.ArticleID = req.ArticleID
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit
	reqPb.ReplyLimit = req.ReplyLimit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListArticleComments(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListArticleComments rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = resPb.Count
	res.Data.Entries = make([]api.CommentEntryWithReplies, 0)

	if resPb.Comments != nil {
		res.Data.Entries = make([]api.CommentEntryWithReplies, len(resPb.Comments))
		for i, comment := range resPb.Comments {
			if err = copyArticleCommentEntryWithReplies(&res.Data.Entries[i], comment.Comment); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "copyArticleCommentEntryWithReplies failed", err.Error())
			}
			replies := make([]api.CommentEntry, len(comment.Replies.Replies))
			for j, reply := range comment.Replies.Replies {
				if err = copyArticleCommentEntry(&replies[j], reply); err != nil {
					log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "copyArticleCommentEntry failed", err.Error())
				}
			}
			res.Data.Entries[i].Replies.Count = resPb.Comments[i].Replies.Count
			res.Data.Entries[i].Replies.Entries = replies
		}
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ListArticleCommentReplies(c *gin.Context) {
	var (
		req   api.ListArticleCommentRepliesRequest
		reqPb pbNews.ListArticleCommentRepliesRequest
		res   api.ListArticleCommentRepliesResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.ParentCommentID = req.ParentCommentID
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListArticleCommentReplies(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListArticleCommentReplies rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = resPb.Count
	res.Data.Entries = make([]api.CommentEntry, 0)

	if resPb.Replies != nil {
		res.Data.Entries = make([]api.CommentEntry, len(resPb.Replies))
		for i, reply := range resPb.Replies {
			if err = copyArticleCommentEntry(&res.Data.Entries[i], reply); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "copyArticleCommentEntry failed", err.Error())
			}
		}
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func copyUserArticleCommentEntry(comment *api.UserArticleCommentEntry, newComment *pbNews.CommentEntry) error {
	if comment == nil || newComment == nil {
		return nil
	}

	if newComment.UserID != "" {
		comment.Nickname = newComment.UserNickname
		comment.FaceURL = newComment.UserFaceURL
	} else if newComment.OfficialID != 0 {
		comment.Nickname = newComment.OfficialNickname
		comment.FaceURL = newComment.OfficialFaceURL
	}

	if newComment.ReplyUserID != "" {
		comment.ReplyNickname = newComment.ReplyUserNickname
		comment.ReplyFaceURL = newComment.ReplyUserFaceURL
	} else if newComment.ReplyOfficialID != 0 {
		comment.ReplyNickname = newComment.ReplyOfficialNickname
		comment.ReplyFaceURL = newComment.ReplyOfficialFaceURL
	}

	return utils.CopyStructFields(&comment, newComment)
}

func copyUserArticleCommentEntryWithTopReplies(comment *api.UserArticleCommentEntryWithTopReplies, newComment *pbNews.UserArticleCommentEntry) error {
	if comment == nil || newComment == nil {
		return nil
	}

	if newComment.Comment.UserID != "" {
		comment.Nickname = newComment.Comment.UserNickname
		comment.FaceURL = newComment.Comment.UserFaceURL
	} else if newComment.Comment.OfficialID != 0 {
		comment.Nickname = newComment.Comment.OfficialNickname
		comment.FaceURL = newComment.Comment.OfficialFaceURL
	}

	if newComment.Comment.ReplyUserID != "" {
		comment.ReplyNickname = newComment.Comment.ReplyUserNickname
		comment.ReplyFaceURL = newComment.Comment.ReplyUserFaceURL
	} else if newComment.Comment.ReplyOfficialID != 0 {
		comment.ReplyNickname = newComment.Comment.ReplyOfficialNickname
		comment.ReplyFaceURL = newComment.Comment.ReplyOfficialFaceURL
	}

	comment.TopReplies = make([]api.UserArticleCommentEntry, 0)
	if newComment.TopReply != nil {
		comment.TopReplies = make([]api.UserArticleCommentEntry, 1)
		if err := copyUserArticleCommentEntry(&comment.TopReplies[0], newComment.TopReply); err != nil {
			return err
		}
	}

	return utils.CopyStructFields(&comment, newComment.Comment)
}

func ListUserArticleComments(c *gin.Context) {
	var (
		req   api.ListUserArticleCommentsRequest
		reqPb pbNews.ListUserArticleCommentsRequest
		res   api.ListUserArticleCommentsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.ArticleID = req.ArticleID
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListUserArticleComments(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListArticleComments rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = resPb.Count
	res.Data.Entries = make([]api.UserArticleCommentEntryWithTopReplies, 0)

	if resPb.Comments != nil {
		res.Data.Entries = make([]api.UserArticleCommentEntryWithTopReplies, len(resPb.Comments))
		for i, comment := range resPb.Comments {
			if err = copyUserArticleCommentEntryWithTopReplies(&res.Data.Entries[i], comment); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "copyUserArticleCommentEntryWithTopReplies failed", err.Error())
			}
		}

		// merge top replies
		for i := 0; i < len(res.Data.Entries); i++ {
			if len(res.Data.Entries[i].TopReplies) > 0 {
				var deleteFlag bool
				for i2, i3 := range res.Data.Entries[i].TopReplies {
					for j := 0; j < i; j++ {
						if i3.ParentCommentID == res.Data.Entries[j].CommentID {
							res.Data.Entries[j].TopReplies = append(res.Data.Entries[j].TopReplies, res.Data.Entries[i].TopReplies[i2])
							res.Data.Entries[i].TopReplies = append(res.Data.Entries[i].TopReplies[:i2], res.Data.Entries[i].TopReplies[i2+1:]...)
							deleteFlag = true
							break
						}
					}
				}
				if deleteFlag {
					res.Data.Entries = append(res.Data.Entries[:i], res.Data.Entries[i+1:]...)
				}
			}
		}
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ListUserArticleCommentReplies(c *gin.Context) {
	var (
		req   api.ListUserArticleCommentRepliesRequest
		reqPb pbNews.ListUserArticleCommentRepliesRequest
		res   api.ListUserArticleCommentRepliesResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.CommentID = req.CommentID
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.ListUserArticleCommentReplies(context.Background(), &reqPb)
	if err != nil {
		errMsg := "ListUserArticleCommentReplies rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.Data.Count = resPb.Count
	res.Data.Entries = make([]api.UserArticleCommentEntry, 0)

	if resPb.Comments != nil {
		res.Data.Entries = make([]api.UserArticleCommentEntry, len(resPb.Comments))
		for i, comment := range resPb.Comments {
			if err = copyUserArticleCommentEntry(&res.Data.Entries[i], comment); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "copyUserArticleCommentEntryWithTopReplies failed", err.Error())
			}
		}
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func GetOfficialProfile(c *gin.Context) {
	var (
		req   api.GetOfficialProfileRequest
		reqPb pbNews.GetOfficialProfileRequest
		res   api.GetOfficialProfileResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, exists := c.Get("userID")
	if exists {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.OfficialID = req.OfficialID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.GetOfficialProfile(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetOfficialProfile rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
	}

	if resPb.Follow != nil {
		res.Data = &api.UserFollow{
			OfficialID: resPb.Follow.OfficialID,
			Nickname:   resPb.Follow.Nickname,
			FaceURL:    resPb.Follow.FaceURL,
			Bio:        resPb.Follow.Bio,
			Type:       resPb.Follow.Type,
			FollowTime: resPb.Follow.FollowTime,
			Muted:      resPb.Follow.Muted,
			Enabled:    resPb.Follow.Enabled,
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func GetRecentAnalytics(c *gin.Context) {
	var (
		req   api.GetRecentAnalyticsRequest
		reqPb pbNews.GetOfficialRecentAnalyticsByGenderRequest
		res   api.GetRecentAnalyticsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, exists := c.Get("userID")
	if exists {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.StartTime = req.StartTime
	reqPb.EndTime = req.EndTime

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.GetOfficialRecentAnalyticsByGender(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetOfficialRecentAnalyticsByGender rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
	}

	if resPb.Current != nil {
		if err = utils.CopyStructFields(&res.Data.Current, resPb.Current); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}

	}

	if resPb.Previous != nil {
		if err = utils.CopyStructFields(&res.Data.Previous, resPb.Previous); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func GetAnalyticsByDay(c *gin.Context) {
	var (
		req   api.GetAnalyticsByDayRequest
		reqPb pbNews.GetOfficialAnalyticsByDayRequest
		res   api.GetAnalyticsByDayResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, exists := c.Get("userID")
	if exists {
		reqPb.UserID = userIDInterface.(string)
	}

	reqPb.StartTime = (req.StartDate.Unix() / 86400) * 86400
	reqPb.EndTime = ((req.EndDate.Unix() / 86400) * 86400) + 86399
	daysNumber := ((reqPb.EndTime - reqPb.StartTime) / 86400) + 1

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	resPb, err := client.GetOfficialAnalyticsByDay(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetOfficialAnalyticsByDay rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if resPb.CommonResp != nil {
		res.CommResp.ErrMsg = resPb.CommonResp.ErrMsg
		res.CommResp.ErrCode = resPb.CommonResp.ErrCode
	}

	var resMap map[int64]*pbNews.AnalyticsByDayEntry

	if resPb.Entries != nil {
		resMap = make(map[int64]*pbNews.AnalyticsByDayEntry, len(resPb.Entries))
		for _, entry := range resPb.Entries {
			resMap[entry.Day] = entry
		}
	}

	res.Data = make([]api.GetAnalyticsByDayEntry, daysNumber)
	for i := 0; i < int(daysNumber); i++ {
		dayUnix := reqPb.StartTime + (int64(i) * 86400)
		res.Data[i] = api.GetAnalyticsByDayEntry{
			Date: time.Unix(dayUnix, 0).UTC().Format(time.RFC3339),
		}
		if item, ok := resMap[dayUnix]; ok {
			res.Data[i].Comments = item.Comments
			res.Data[i].Follows = item.Follows
			res.Data[i].Likes = item.Likes
			res.Data[i].Reads = item.Reads
			res.Data[i].UniqueReads = item.UniqueReads
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func SearchOfficialAccounts(c *gin.Context) {
	var (
		req   api.SearchOfficialAccountsRequest
		reqPb pbNews.SearchOfficialAccountsRequest
		res   api.SearchOfficialAccountsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.Keyword = req.Keyword
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	//if strings.TrimSpace(req.Keyword) == "" {
	//	c.JSON(http.StatusOK, res)
	//	return
	//}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.SearchOfficialAccounts(context.Background(), &reqPb)
	if err != nil {
		errMsg := "SearchOfficialAccounts rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		res.ErrCode = respPb.CommonResp.ErrCode
		res.ErrMsg = respPb.CommonResp.ErrMsg
	}

	res.Data = api.SearchOfficialAccountsData{
		Count:   respPb.Count,
		Entries: make([]api.UserFollow, 0),
	}

	if respPb.Entries != nil {
		res.Data.Entries = make([]api.UserFollow, len(respPb.Entries))
		for i, follow := range respPb.Entries {
			res.Data.Entries[i] = api.UserFollow{
				OfficialID: follow.OfficialID,
				Nickname:   follow.Nickname,
				FaceURL:    follow.FaceURL,
				Bio:        follow.Bio,
				Type:       follow.Type,
				FollowTime: follow.FollowTime,
				Muted:      follow.Muted,
				Enabled:    follow.Enabled,
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func SearchArticles(c *gin.Context) {
	var (
		req   api.SearchArticlesRequest
		reqPb pbNews.SearchArticlesRequest
		res   api.SearchArticlesResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	//if strings.TrimSpace(req.Keyword) == "" {
	//	c.JSON(http.StatusOK, res)
	//	return
	//}

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.Keyword = req.Keyword
	reqPb.OfficialID = req.OfficialID
	reqPb.MinReadTime = req.MinReadTime
	reqPb.MaxReadTime = req.MaxReadTime
	reqPb.MinCreateTime = req.MinCreateTime
	reqPb.MaxCreateTime = req.MaxCreateTime
	reqPb.Sort = req.Sort
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.SearchArticles(context.Background(), &reqPb)
	if err != nil {
		errMsg := "SearchArticles rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		res.CommResp.ErrCode = respPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
	}

	res.Data = &api.SearchArticlesData{
		Count:   respPb.Count,
		Entries: make([]api.SearchArticlesEntry, 0),
	}

	if respPb.Entries != nil {
		res.Data.Entries = make([]api.SearchArticlesEntry, len(respPb.Entries))
		if err = utils.CopyStructFields(&res.Data.Entries, respPb.Entries); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func GetUserArticleByArticleID(c *gin.Context) {
	var (
		req   api.GetUserArticleByArticleIDRequest
		reqPb pbNews.GetUserArticleByArticleIDRequest
		res   api.GetUserArticleByArticleIDResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		userID := userIDInterface.(string)
		reqPb.UserID = &userID
	}
	reqPb.ArticleID = req.ArticleID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.GetUserArticleByArticleID(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetUserArticleByArticleID rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		res.CommResp.ErrCode = respPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
	}

	if respPb.Article != nil || respPb.Official != nil {
		res.Data = &api.GetUserArticleByArticleIDData{}

		if respPb.Article != nil {
			if err = utils.CopyStructFields(&res.Data.Article, respPb.Article); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
			}
		}

		if respPb.Official != nil {
			if err = utils.CopyStructFields(&res.Data.Official, respPb.Official); err != nil {
				log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func InsertArticleRead(c *gin.Context) {
	var (
		req   api.InsertArticleReadRequest
		reqPb pbNews.InsertArticleReadRequest
		res   api.CommResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInterface, existed := c.Get("userID")
	if existed {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.ArticleID = req.ArticleID

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.InsertArticleRead(context.Background(), &reqPb)
	if err != nil {
		errMsg := "InsertArticleRead rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	res.ErrCode = respPb.ErrCode
	res.ErrMsg = respPb.ErrMsg

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ListUserArticleReads(c *gin.Context) {
	var (
		req   api.ListUserArticleReadsRequest
		reqPb pbNews.ListUserArticleReadsRequest
		res   api.ListUserArticleReadsResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if userIDInterface, ok := c.Get("userID"); ok {
		reqPb.UserID = userIDInterface.(string)
	}
	reqPb.Offset = req.Offset
	reqPb.Limit = req.Limit
	reqPb.MinCreateTime = time.Now().Unix() - 604800

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.ListUserArticleReads(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetUserArticleByArticleID rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb.CommonResp != nil {
		res.CommResp.ErrCode = respPb.CommonResp.ErrCode
		res.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
	}

	res.Data.Count = respPb.Count
	res.Data.Entries = make([]api.ListUserArticleReadsEntry, 0)
	if respPb.Entries != nil {
		res.Data.Entries = make([]api.ListUserArticleReadsEntry, len(respPb.Entries))
		if err = utils.CopyStructFields(&res.Data.Entries, respPb.Entries); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed "+err.Error())
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func ClearUserArticleReads(c *gin.Context) {
	var (
		req   api.ClearUserArticleReadsRequest
		reqPb pbNews.ClearUserArticleReadsRequest
		res   api.CommResp
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if userIDInterface, ok := c.Get("userID"); ok {
		reqPb.UserID = userIDInterface.(string)
	}

	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.ClearUserArticleReads(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetUserArticleByArticleID rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb != nil {
		res.ErrCode = respPb.ErrCode
		res.ErrMsg = respPb.ErrMsg
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}
func DeleteArticleComment(c *gin.Context) {
	var (
		req   api.DeleteArticleCommentRequest
		resp  api.DeleteArticleCommentResponse
		reqPb pbNews.DeleteArticleCommentRequest
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userIDInter, existed := c.Get("userID")
	if existed {
		reqPb.ReqUserID = userIDInter.(string)
	}

	reqPb.CommentID = req.CommentID
	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	respPb, err := client.DeleteArticleComment(context.Background(), &reqPb)
	if err != nil {
		errMsg := "delete article comment failed" + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if respPb != nil {
		resp.ErrMsg = respPb.ErrMsg
		resp.ErrCode = respPb.ErrCode
	} else {
		resp.ErrMsg = "rpc failed to return response"
		resp.ErrCode = 400
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetFollowedOfficialConversation(c *gin.Context) {
	var (
		req   api.FollowedOfficialConversationRequest
		reqPb pbNews.FollowedOfficialConversationRequest
		res   api.FollowedOfficialConversationResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if userIDInterface, ok := c.Get("userID"); ok {
		reqPb.ReqUserID = userIDInterface.(string)
	}
	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}
	respPb, err := client.FollowedOfficialConversation(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetUserArticleByArticleID rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		res.CommResp = api.CommResp{ErrCode: 400, ErrMsg: errMsg}
		c.JSON(http.StatusBadRequest, res)
		return
	}
	res.CommResp = api.CommResp{ErrCode: 0, ErrMsg: ""}
	for _, article := range respPb.Articles {
		res.Data = append(res.Data, *article)
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}

func GetOfficialIDNumberAvailability(c *gin.Context) {
	var (
		req   api.GetOfficialIDNumberAvailabilityRequest
		reqPb pbNews.GetOfficialIDNumberAvailabilityRequest
		res   api.GetOfficialIDNumberAvailabilityResponse
	)
	if err := c.BindJSON(&req); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb = pbNews.GetOfficialIDNumberAvailabilityRequest{IDNumber: req.IDNumber, IDType: req.IDType}
	client, err := createNewsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}
	resp, err := client.GetOfficialIDNumberAvailability(context.Background(), &reqPb)
	if err != nil {
		errMsg := "GetUserArticleByArticleID rpc failed " + err.Error()
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, reqPb.String())
		res.CommResp = api.CommResp{ErrCode: 400, ErrMsg: errMsg}
		c.JSON(http.StatusBadRequest, res)
		return
	}
	res.CommResp = api.CommResp{ErrCode: 0, ErrMsg: ""}
	res.Data.IDNumber = resp.IDNumber
	res.Data.IsAvailable = resp.IsAvailable

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", res)
	c.JSON(http.StatusOK, res)
}
