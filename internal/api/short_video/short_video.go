package short_video

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbMoment "Open_IM/pkg/proto/moments"
	pbShortVideo "Open_IM/pkg/proto/short_video"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetUpdateShortVideoSign(c *gin.Context) {
	var (
		apiRequest  api.GetUpdateShortVideoSignRequest
		apiResponse api.GetUpdateShortVideoSignResponse
		rpcRequest  pbShortVideo.GetUpdateShortVideoSignRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")

	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Desc = apiRequest.Desc
	rpcRequest.InterestId = apiRequest.InterestId
	rpcResponse, err := connection.GetUpdateShortVideoSign(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.Sign = rpcResponse.Sign

	c.JSON(http.StatusOK, apiResponse)
	return
}

func CreateShortVideo(c *gin.Context) {
	var (
		apiRequest  api.CreateShortVideoSignRequest
		apiResponse api.CreateShortVideoSignResponse
		rpcRequest  pbShortVideo.CreateShortVideoRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")

	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Name = apiRequest.Name
	rpcRequest.Desc = apiRequest.Desc
	rpcRequest.InterestIds = apiRequest.InterestId
	rpcRequest.MediaUrl = apiRequest.MediaUrl
	rpcRequest.CoverUrl = apiRequest.CoverUrl
	rpcRequest.FileId = apiRequest.FileId
	rpcResponse, err := connection.CreateShortVideo(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.FileId = rpcResponse.FileId

	c.JSON(http.StatusOK, apiResponse)
	return
}

func SearchShortVideo(c *gin.Context) {
	var (
		apiRequest  api.SearchShortVideoRequest
		apiResponse api.SearchShortVideoResponse
		rpcRequest  pbShortVideo.SearchShortVideoRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Keyword = apiRequest.KeyWord
	rpcResponse, err := connection.SearchShortVideo(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(rpcRequest.OperationID, utils.GetSelfFuncName(), rpcResponse.String())
	if rpcResponse.CommonResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	_ = utils.CopyStructFields(&apiResponse.Data.ShortVideoList, rpcResponse.ShortVideoInfoList)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetUserCount(c *gin.Context) {
	var (
		apiRequest  api.GetUserCountRequest
		apiResponse api.GetUserCountResponse
		rpcRequest  pbShortVideo.GetShortVideoUserCountByUserIdRequest

		momentRequest pbMoment.GetUserMomentCountRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")

	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = apiRequest.UserID
	rpcRequest.OperateUserId = userIDInterface.(string)
	rpcResponse, err := connection.GetShortVideoUserCountByUserId(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if rpcResponse.CommonResp.ErrCode != 0 {
		c.JSON(http.StatusOK, apiResponse)
		return
	}
	_ = utils.CopyStructFields(&apiResponse.Data.ShortVideo, rpcResponse)

	// get user moment count
	apiResponse.Data.Moment.WorkNum = 0
	apiResponse.Data.Moment.LikeNum = 0
	client, err := createMomentsGRPConnection(apiRequest.OperationID)
	if err == nil {
		momentRequest.OperationID = apiRequest.OperationID
		momentRequest.CurrentUserId = userIDInterface.(string)
		momentRequest.UserId = apiRequest.UserID
		respPb, err := client.GetUserMomentCount(context.Background(), &momentRequest)
		if err == nil {
			apiResponse.Data.Moment.WorkNum = respPb.Posts
			apiResponse.Data.Moment.LikeNum = respPb.Likes
		}
	}

	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetUserNotices(c *gin.Context) {
	var (
		apiRequest  api.GetUserNoticesRequest
		apiResponse api.GetUserNoticesResponse
		rpcRequest  pbShortVideo.GetShortVideoNoticesRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")

	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.State = int32(apiRequest.State)
	rpcRequest.NoticeType = apiRequest.Type
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		ShowNumber: apiRequest.ShowNumber,
		PageNumber: apiRequest.PageNumber,
	}
	rpcResponse, err := connection.GetShortVideoNoticeList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	log.NewInfo(rpcRequest.OperationID, utils.GetSelfFuncName(), rpcResponse.String())

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if rpcResponse.CommonResp.ErrCode != 0 {
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	apiResponse.Data.NoticeCount = rpcResponse.NoticeCount
	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	_ = utils.CopyStructFields(&apiResponse.Data.NoticeList, rpcResponse.ShortVideoNoticeList)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetShortVideoByFileId(c *gin.Context) {
	var (
		apiRequest  api.GetShortVideoByFileIdRequest
		apiResponse api.GetShortVideoByFileIdResponse
		rpcRequest  pbShortVideo.GetShortVideoByFileIdRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.FileId = apiRequest.FileId
	rpcResponse, err := connection.GetShortVideoByFieldId(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if rpcResponse.CommonResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	utils.CopyStructFields(&apiResponse.Data, rpcResponse.ShortVideoInfo)
	apiResponse.Data.UpUserInfo.UserId = rpcResponse.ShortVideoInfo.UpUserInfo.UserId
	apiResponse.Data.UpUserInfo.NickName = rpcResponse.ShortVideoInfo.UpUserInfo.Nickname
	apiResponse.Data.UpUserInfo.AvatarUrl = rpcResponse.ShortVideoInfo.UpUserInfo.FaceURL
	apiResponse.Data.SelfOperation.IsLike = rpcResponse.ShortVideoInfo.SelfInfo.IsLike
	apiResponse.Data.SelfOperation.IsComment = rpcResponse.ShortVideoInfo.SelfInfo.IsComment

	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetLikeShortVideoList(c *gin.Context) {
	var (
		apiRequest  api.GetLikeShortVideoRequest
		apiResponse api.GetLikeShortVideoResponse
		rpcRequest  pbShortVideo.GetLikeShortVideoListRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = apiRequest.UserId
	rpcRequest.OperatorUserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		PageNumber: apiRequest.PageNumber,
		ShowNumber: apiRequest.ShowNumber,
	}
	rpcResponse, err := connection.GetLikeShortVideoList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(rpcRequest.OperationID, utils.GetSelfFuncName(), rpcResponse.String())
	if rpcResponse.CommonResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.ShortVideoCount = rpcResponse.ShortVideoCount
	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	utils.CopyStructFields(&apiResponse.Data.ShortVideoList, rpcResponse.ShortVideoInfoList)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetShortVideoListByUserId(c *gin.Context) {
	var (
		apiRequest  api.GetShortVideoListByUserIdRequest
		apiResponse api.GetLikeShortVideoResponse
		rpcRequest  pbShortVideo.GetShortVideoListByUserIdRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = apiRequest.UserId
	rpcRequest.OperatorUserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		PageNumber: apiRequest.PageNumber,
		ShowNumber: apiRequest.ShowNumber,
	}
	rpcResponse, err := connection.GetShortVideoListByUserId(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}
	log.NewInfo(rpcRequest.OperationID, utils.GetSelfFuncName(), rpcResponse.String())
	if rpcResponse.CommonResp.ErrCode != 0 {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.ShortVideoCount = rpcResponse.ShortVideoCount
	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	_ = utils.CopyStructFields(&apiResponse.Data.ShortVideoList, rpcResponse.ShortVideoInfoList)
	_ = utils.CopyStructFields(&apiResponse.Data.IsUpFollow, rpcResponse.IsUpFollow)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func ShortVideoLike(c *gin.Context) {
	var (
		apiRequest  api.ShortVideoLikeRequest
		apiResponse api.ShortVideoLikeResponse
		rpcRequest  pbShortVideo.ShortVideoLikeRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	// Current limiting
	if db.DB.ShortVideoLikeLimit(userIDInterface.(string), apiRequest.FileId, *apiRequest.Like) == false {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.ErrLikeLimit.ErrCode, ErrMsg: constant.ErrLikeLimit.ErrMsg})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.FileId = apiRequest.FileId
	rpcRequest.Like = *apiRequest.Like

	rpcResponse, err := connection.ShortVideoLike(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func ShortVideoFollow(c *gin.Context) {
	var (
		apiRequest  api.ShortVideoFollowRequest
		apiResponse api.ShortVideoFollowResponse
		rpcRequest  pbShortVideo.FollowRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	// Current limiting
	if db.DB.ShortVideoFollowLimit(userIDInterface.(string)) == false {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.ErrFollowLimit.ErrCode, ErrMsg: constant.ErrFollowLimit.ErrMsg})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.FollowUserId = apiRequest.FollowUserId
	rpcRequest.Follow = *apiRequest.Follow

	rpcResponse, err := connection.Follow(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func ShortVideoFollowList(c *gin.Context) {
	var (
		apiRequest  api.ShortVideoFollowListRequest
		apiResponse api.ShortVideoFollowListResponse
		rpcRequest  pbShortVideo.GetFollowListRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.OperateUserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		ShowNumber: apiRequest.ShowNumber,
		PageNumber: apiRequest.PageNumber,
	}
	rpcResponse, err := connection.GetFollowList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.FollowCount = rpcResponse.FollowCount
	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	apiResponse.Data.FollowList = make([]api.UserInfo, 0)
	_ = utils.CopyStructFields(&apiResponse.Data.FollowList, rpcResponse.FollowList)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func ShortVideoFansList(c *gin.Context) {
	var (
		apiRequest  api.ShortVideoFansListRequest
		apiResponse api.ShortVideoFansListResponse
		rpcRequest  pbShortVideo.GetFansListRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.OperateUserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		ShowNumber: apiRequest.ShowNumber,
		PageNumber: apiRequest.PageNumber,
	}
	rpcResponse, err := connection.GetFansList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.FansCount = rpcResponse.FansCount
	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	apiResponse.Data.FansList = make([]api.UserInfo, 0)
	_ = utils.CopyStructFields(&apiResponse.Data.FansList, rpcResponse.FansList)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetRecommendList(c *gin.Context) {
	var (
		apiRequest  api.GetRecommendListRequest
		apiResponse api.GetRecommendListResponse
		rpcRequest  pbShortVideo.GetRecommendShortVideoListRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	// Current limiting
	if db.DB.ShortVideoRecommendLimit(userIDInterface.(string)) == false {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.ErrRecommendLimit.ErrCode, ErrMsg: constant.ErrRecommendLimit.ErrMsg})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Size = apiRequest.Size

	rpcResponse, err := connection.GetRecommendShortVideoList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if apiResponse.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusOK, apiResponse)
		return
	}
	utils.CopyStructFields(&apiResponse.Data.ShortVideoFileIdList, rpcResponse.FileIdList)
	utils.CopyStructFields(&apiResponse.Data.ShortVideoInfoList, rpcResponse.ShortVideoInfoList)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetFollowShortVideoList(c *gin.Context) {
	var (
		apiRequest  api.GetFollowShortVideoListRequest
		apiResponse api.GetFollowShortVideoListResponse
		rpcRequest  pbShortVideo.GetFollowShortVideoListRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		ShowNumber: apiRequest.ShowNumber,
		PageNumber: apiRequest.PageNumber,
	}

	rpcResponse, err := connection.GetFollowShortVideoList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if apiResponse.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusOK, apiResponse)
		return
	}
	utils.CopyStructFields(&apiResponse.Data.ShortVideoList, rpcResponse.FileIdList)
	apiResponse.Data.ShortVideoInfoList = make([]api.ShortVideoInfo, 0)
	utils.CopyStructFields(&apiResponse.Data.ShortVideoInfoList, rpcResponse.ShortVideoInfoList)

	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	apiResponse.Data.ShortVideoCount = rpcResponse.ShortVideoCount
	c.JSON(http.StatusOK, apiResponse)
	return
}

func ShortVideoCommentLike(c *gin.Context) {
	var (
		apiRequest  api.ShortVideoCommentLikeRequest
		apiResponse api.ShortVideoLikeCommentResponse
		rpcRequest  pbShortVideo.ShortVideoCommentLikeRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	// Current limiting
	if db.DB.ShortVideoLikeLimit(userIDInterface.(string), apiRequest.FileId, *apiRequest.Like) == false {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.ErrLikeLimit.ErrCode, ErrMsg: constant.ErrLikeLimit.ErrMsg})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.FileId = apiRequest.FileId
	rpcRequest.CommentId = apiRequest.CommentId
	rpcRequest.Like = *apiRequest.Like
	rpcResponse, err := connection.ShortVideoCommentLike(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func ShortVideoComment(c *gin.Context) {
	var (
		apiRequest  api.ShortVideoCommentRequest
		apiResponse api.ShortVideoCommentResponse
		rpcRequest  pbShortVideo.ShortVideoCommentRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	if db.DB.ShortVideoCommentLimit(userIDInterface.(string), apiRequest.FileId) == false {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.ErrCommentLimit.ErrCode, ErrMsg: constant.ErrCommentLimit.ErrMsg})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.FileId = apiRequest.FileId
	rpcRequest.ParentId = apiRequest.ParentId
	rpcRequest.Content = apiRequest.Content
	rpcResponse, err := connection.ShortVideoComment(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	utils.CopyStructFields(&apiResponse.Data, rpcResponse)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func DeleteShortVideoComment(c *gin.Context) {
	var (
		apiRequest  api.DeleteShortVideoCommentRequest
		apiResponse api.DeleteShortVideoCommentResponse
		rpcRequest  pbShortVideo.DeleteShortVideoRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.FileId = apiRequest.FileId
	rpcRequest.CommentId = apiRequest.CommentId
	rpcResponse, err := connection.DeleteShortVideoComment(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	c.JSON(http.StatusOK, apiResponse)
	return
}

func GetShortVideoComment(c *gin.Context) {
	var (
		apiRequest  api.GetShortVideoCommentListRequest
		apiResponse api.GetShortVideoCommentListResponse
		rpcRequest  pbShortVideo.GetShortVideoCommentListRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.FileId = apiRequest.FileId
	rpcRequest.ParentId = apiRequest.ParentId
	rpcRequest.OrderBy = apiRequest.Order
	rpcRequest.SourceCommentId = apiRequest.SourceCommentId
	rpcRequest.OperationUserID = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		PageNumber: apiRequest.PageNumber,
		ShowNumber: apiRequest.ShowNumber,
	}
	rpcResponse, err := connection.GetShortVideoCommentList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if rpcResponse.CommonResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	apiResponse.Data.CommentCount = rpcResponse.CommentCount
	apiResponse.Data.Level0CommentCount = rpcResponse.Level0CommentCount
	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber

	// delete Modify some information about deleted comments
	for i, comment := range rpcResponse.CommentList {
		if comment.Status == constant.ShortVideoCommentStatusDeleted {
			rpcResponse.CommentList[i].Content = ""
			rpcResponse.CommentList[i].CommentUserInfo.UserId = ""
			rpcResponse.CommentList[i].CommentUserInfo.Nickname = ""
			rpcResponse.CommentList[i].CommentUserInfo.FaceURL = ""
		}
		for j, reply := range comment.ReplyComment {
			if reply.Status == constant.ShortVideoCommentStatusDeleted {
				rpcResponse.CommentList[i].ReplyComment[j].Content = ""
				rpcResponse.CommentList[i].ReplyComment[j].CommentUserInfo.UserId = ""
				rpcResponse.CommentList[i].ReplyComment[j].CommentUserInfo.Nickname = ""
				rpcResponse.CommentList[i].ReplyComment[j].CommentUserInfo.FaceURL = ""
			}
		}
	}
	_ = utils.CopyStructFields(&apiResponse.Data.CommentList, rpcResponse.CommentList)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func CommentPage(c *gin.Context) {
	var (
		apiRequest  api.GetCommentPageRequest
		apiResponse api.GetCommentPageResponse
		rpcRequest  pbShortVideo.GetCommentPageRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if apiRequest.CommentId == 0 && apiRequest.SourceCommentId == 0 {
		errMsg := "commentId and sourceCommentId can't be empty at the same time"
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.CommentId = apiRequest.CommentId
	rpcRequest.SourceCommentId = apiRequest.SourceCommentId
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		PageNumber: apiRequest.PageNumber,
		ShowNumber: apiRequest.ShowNumber,
	}

	rpcResponse, err := connection.GetCommentPage(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if rpcResponse.CommonResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	apiResponse.Data.ReplyCount = rpcResponse.ReplyCount
	apiResponse.Data.CreatorUserId = rpcResponse.CreatorUserId
	apiResponse.Data.TotalReplyNum = rpcResponse.TotalReplyNum

	if rpcResponse.CommentInfo.Status == constant.ShortVideoCommentStatusDeleted {
		rpcResponse.CommentInfo.Content = ""
		rpcResponse.CommentInfo.CommentUserInfo.UserId = ""
		rpcResponse.CommentInfo.CommentUserInfo.Nickname = ""
		rpcResponse.CommentInfo.CommentUserInfo.FaceURL = ""
	}
	for i, reply := range rpcResponse.ReplyList {
		if reply.Status == constant.ShortVideoCommentStatusDeleted {
			rpcResponse.ReplyList[i].Content = ""
			rpcResponse.ReplyList[i].CommentUserInfo.UserId = ""
			rpcResponse.ReplyList[i].CommentUserInfo.Nickname = ""
			rpcResponse.ReplyList[i].CommentUserInfo.FaceURL = ""
		}
		for j, ReplyComment := range reply.ReplyComment {
			if ReplyComment.Status == constant.ShortVideoCommentStatusDeleted {
				rpcResponse.ReplyList[i].ReplyComment[j].Content = ""
				rpcResponse.ReplyList[i].ReplyComment[j].CommentUserInfo.UserId = ""
				rpcResponse.ReplyList[i].ReplyComment[j].CommentUserInfo.Nickname = ""
				rpcResponse.ReplyList[i].ReplyComment[j].CommentUserInfo.FaceURL = ""
			}
		}
	}

	_ = utils.CopyStructFields(&apiResponse.Data.ReplyList, rpcResponse.ReplyList)
	_ = utils.CopyStructFields(&apiResponse.Data.CommentInfo, rpcResponse.CommentInfo)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func CommentPageReplyList(c *gin.Context) {
	var (
		apiRequest  api.GetCommentPageReplyListRequest
		apiResponse api.GetCommentPageReplyListResponse

		rpcRequest pbShortVideo.GetCommentPageReplyListRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.CommentId = apiRequest.CommentId
	rpcRequest.SourceCommentId = apiRequest.SourceCommentId
	rpcRequest.UserId = userIDInterface.(string)
	rpcRequest.Pagination = &pbShortVideo.RequestPagination{
		PageNumber: apiRequest.PageNumber,
		ShowNumber: apiRequest.ShowNumber,
	}

	rpcResponse, err := connection.GetCommentPageReplyList(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if rpcResponse.CommonResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	apiResponse.Data.CurrentPage = rpcResponse.Pagination.CurrentPage
	apiResponse.Data.ShowNumber = rpcResponse.Pagination.ShowNumber
	apiResponse.Data.ReplyCount = rpcResponse.ReplyCount

	for i, reply := range rpcResponse.ReplyList {
		if reply.Status == constant.ShortVideoCommentStatusDeleted {
			rpcResponse.ReplyList[i].Content = ""
			rpcResponse.ReplyList[i].CommentUserInfo.UserId = ""
			rpcResponse.ReplyList[i].CommentUserInfo.Nickname = ""
			rpcResponse.ReplyList[i].CommentUserInfo.FaceURL = ""
		}
	}

	_ = utils.CopyStructFields(&apiResponse.Data.ReplyList, rpcResponse.ReplyList)

	c.JSON(http.StatusOK, apiResponse)
	return
}

func BlockShortVideo(c *gin.Context) {
	var (
		apiRequest  api.BlockShortVideoRequest
		apiResponse api.BlockShortVideoResponse
		rpcRequest  pbShortVideo.BlockShortVideoRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := CreateShortVideoGRPConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationID
	rpcRequest.FileId = apiRequest.FileId
	rpcRequest.UserId = userIDInterface.(string)
	rpcResponse, err := connection.BlockShortVideo(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if rpcResponse.CommonResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: rpcResponse.CommonResp.ErrCode, ErrMsg: rpcResponse.CommonResp.ErrMsg})
		return
	}

	c.JSON(http.StatusOK, apiResponse)
	return
}

func CreateShortVideoGRPConnection(OperationID string) (pbShortVideo.ShortVideoClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImShortVideoName, OperationID)
	if etcdConn == nil {
		errMsg := "etcd3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbShortVideo.NewShortVideoClient(etcdConn)
	return client, nil
}

func createMomentsGRPConnection(OperationID string) (pbMoment.MomentsClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, OperationID)
	if etcdConn == nil {
		errMsg := "getcdv3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbMoment.NewMomentsClient(etcdConn)
	return client, nil
}
