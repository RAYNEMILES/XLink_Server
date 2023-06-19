package moments

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbMoment "Open_IM/pkg/proto/moments"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// createGRPCconnection create connection for API request
func createMomentsGRPCconnection(OperationID string) (pbMoment.MomentsClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, OperationID)
	if etcdConn == nil {
		errMsg := "getcdv3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbMoment.NewMomentsClient(etcdConn)
	return client, nil
}

// CreateMoment create moment API
func CreateMoment(c *gin.Context) {
	var (
		req    api.MomentCreateRequest
		resp   api.MomentCreatResp
		reqPb  pbMoment.Moment
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.CreateMoment(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	if respPb.Moment != nil {
		resp.Moment = new(api.Moment)
		err = utils.CopyStructFields(resp.Moment, respPb.Moment)
		if err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
		if respPb.Moment.ArticleDetailsInMoment != nil {
			resp.Moment.ArticleDetailsInMoment = &api.GetUserArticleByArticleIDData{}
			err = utils.CopyStructFields(resp.Moment.ArticleDetailsInMoment, respPb.Moment.ArticleDetailsInMoment)
		}
		if respPb.Moment.WoomDetails != nil {
			resp.Moment.WoomDetails = &api.ShortVideoInfo{}
			err = utils.CopyStructFields(resp.Moment.WoomDetails, respPb.Moment.WoomDetails)
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// CreateMomentLike Like a moment
func CreateMomentLike(c *gin.Context) {
	var (
		req    api.MomentLikeRequest
		resp   api.MomentCreatResp
		reqPb  pbMoment.MomentLike
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.CreateMomentLike(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// CancelMomentLike Cancel Like done on a moment
func CancelMomentLike(c *gin.Context) {
	var (
		req    api.MomentCacelLikeRequest
		resp   api.MomentCreatResp
		reqPb  pbMoment.MomentCancelLike
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.CancelMomentLike(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// CreateMomentCommentcreate a comment on a moment
func CreateMomentComment(c *gin.Context) {
	var (
		req    api.MomentCommentCreateRequest
		resp   api.MomentCommentCreateResp
		reqPb  pbMoment.MomentComment
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.CreateMomentComment(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	if respPb.Comment != nil {
		resp.MomentComment = new(api.MomentCommentResp)
		err = utils.CopyStructFields(resp.MomentComment, respPb.Comment)
		if err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// CreateReplyOfMomentComment reply a comment on a moment
func CreateReplyOfMomentComment(c *gin.Context) {
	var (
		req    api.CreateReplyOfMomentCommentRequest
		resp   api.CreateReplyOfMomentCommentResp
		reqPb  pbMoment.ReplyOfMomentComment
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.CreateReplyOfMomentComment(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	if respPb.Comment != nil {
		resp.MomentComment = new(api.MomentCommentResp)
		err = utils.CopyStructFields(resp.MomentComment, respPb.Comment)
		if err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// GetListHomeTimeLineOfMoments list down the moments on home with comments
func GetListHomeTimeLineOfMoments(c *gin.Context) {
	var (
		req    api.ListHomeTimeLineOfMomentsRequest
		reqPb  pbMoment.ListHomeTimeLineOfMomentsReq
		resp   api.ListHomeTimeLineOfMomentsResp
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	reqPb.CreatorID = userID

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	homeTimeLineOfMoments, err := client.GetListHomeTimeLineOfMoments(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetListHomeTimeLineOfMoments rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "GetListHomeTimeLineOfMoments rpc failed, " + err.Error()})
		return
	}

	err = utils.CopyStructFields(&resp, homeTimeLineOfMoments)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	resp.PageNumber = req.PageNumber
	resp.MomentLimit = req.MomentLimit
	resp.CommentsLimit = req.CommentsLimit
	c.JSON(http.StatusOK, resp)
}

// GetMomentDetailsByID get the details of a moment and 100 comments
func GetMomentDetailsByID(c *gin.Context) {
	var (
		req    api.GetMomentDetailsByIDRequest
		reqPb  pbMoment.GetMomentDetailsByIDRequest
		resp   api.GetMomentDetailsByIDResponse
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	reqPb.CreatorID = userID

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	getMomentDetailsByIDResponse, err := client.GetMomentDetailsByID(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}

	err = utils.CopyStructFields(&resp, getMomentDetailsByIDResponse)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	if resp.MomentComments != nil {
		resp.PageNumber = int64(len(resp.MomentComments) / 20)
	}
	resp.CommentsLimit = 20
	c.JSON(http.StatusOK, resp)
}

// GetMomentDetailsByID get the details of a moment and 100 comments
func GetMomentCommentsByID(c *gin.Context) {
	var (
		req    api.GetMomentCommentsByIDRequest
		reqPb  pbMoment.GetMomentCommentsByIDRequest
		resp   api.GetMomentCommentsByIDResponse
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	reqPb.CreatorID = userID

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	getMomentCommentsByIDResponse, err := client.GetMomentCommentsByID(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}

	err = utils.CopyStructFields(&resp, getMomentCommentsByIDResponse)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)

	resp.PageNumber = req.PageNumber
	resp.CommentsLimit = req.CommentsLimit
	c.JSON(http.StatusOK, resp)
}

// RepostAMoment repost a moment
func RepostAMoment(c *gin.Context) {
	var (
		req   api.RepostAMomentRequest
		reqPb pbMoment.RepostAMomentRequest
		resp  api.RepostAMomentResp

		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.RepostAMoment(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	if respPb.Moment != nil {
		resp.Moment = new(api.Moment)
		err = utils.CopyStructFields(resp.Moment, respPb.Moment)
		if err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// DeleteMoment Delete moment by owner
func DeleteMoment(c *gin.Context) {
	var (
		req    api.DeleteMomentRequest
		resp   api.DeleteMomentResp
		reqPb  pbMoment.DeleteMomentRequest
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.DeleteMoment(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "DeleteMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

// DeleteMomentComment Delete moment comment by owner
func DeleteMomentComment(c *gin.Context) {
	var (
		req    api.DeleteMomentCommentRequest
		resp   api.DeleteMomentResp
		reqPb  pbMoment.DeleteMomentCommentRequest
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	req.CreatorID = userID

	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.DeleteMomentComment(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteMomentComment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "DeleteMoment rpc failed, " + err.Error()})
		return
	}
	resp.ErrMsg = respPb.ErrMsg
	resp.ErrCode = respPb.ErrCode
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetAnyUserMomentsByID(c *gin.Context) {

	var (
		req   api.GetAnyUserMomentsByIDRequest
		resp  api.GetAnyUserMomentsByIDResp
		reqPb pbMoment.GetAnyUserMomentsByIDRequest
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.CurrentUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	reqPb.UserId = req.UserId

	reqPb.PageNumber = req.PageNumber
	reqPb.ShowNumber = req.ShowNumber

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.GetAnyUserMomentsByID(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "DeleteMoment rpc failed, " + err.Error()})
		return
	}
	err = utils.CopyStructFields(&resp.Moments, respPb.Moments)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	resp.ErrMsg = respPb.CommonResp.ErrMsg
	resp.ErrCode = respPb.CommonResp.ErrCode

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func GetUserMomentCount(c *gin.Context) {

	var (
		req   api.GetUserMomentCountRequest
		resp  api.GetUserMomentCountResp
		reqPb pbMoment.GetUserMomentCountRequest
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.CurrentUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	reqPb.UserId = req.UserId

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	respPb, err := client.GetUserMomentCount(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "DeleteMoment rpc failed, " + err.Error()})
		return
	}

	resp.Posts = respPb.Posts
	resp.ErrMsg = respPb.CommonResp.ErrMsg
	resp.ErrCode = respPb.CommonResp.ErrCode

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func GlobalSearchInMoments(c *gin.Context) {

	var (
		req    api.GlobalSearchInMomentsRequest
		reqPb  pbMoment.GlobalSearchInMomentsRequest
		resp   api.GlobalSearchInMomentsResp
		userID string
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	req.CommentsLimit = 1
	req.MomentLimit = 20
	err := utils.CopyStructFields(&reqPb, req)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	reqPb.CreatorID = userID

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	homeTimeLineOfMoments, err := client.GlobalSearchInMoments(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CreateMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "CreateMoment rpc failed, " + err.Error()})
		return
	}

	err = utils.CopyStructFields(&resp, homeTimeLineOfMoments)
	if err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	resp.PageNumber = req.PageNumber
	resp.MomentLimit = req.MomentLimit
	resp.CommentsLimit = req.CommentsLimit
	c.JSON(http.StatusOK, resp)
}

func GetMomentAnyUserMediaByID(c *gin.Context) {
	var (
		req   api.GetMomentAnyUserMediaByIDRequest
		resp  api.GetMomentAnyUserMediaByIDResp
		reqPb pbMoment.GetMomentAnyUserMediaByIDRequest
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "bind json failed " + err.Error()})
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()
	reqPb.LastCount = req.LastCount
	reqPb.UserID = req.UserID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	client, err := createMomentsGRPCconnection(req.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	rpcResp, err := client.GetMomentAnyUserMediaByID(context.Background(), &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteMoment rpc failed, ", reqPb.String(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "DeleteMoment rpc failed, " + err.Error()})
		return
	}

	for _, pic := range rpcResp.Pics {
		resp.Pics = append(resp.Pics, struct {
			URL  string `json:"url"`
			Type int8   `json:"type"`
		}{URL: pic.URL, Type: int8(pic.Type)})
	}
	resp.AllMediaMomentCount = rpcResp.AllMediaMomentCount

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, gin.H{"data": resp})

}
