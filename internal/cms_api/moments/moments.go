package moments

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/moments"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetMoments(c *gin.Context) {

	var (
		req   cms_api_struct.GetMomentsRequest
		resp  cms_api_struct.GetMomentsResponse
		reqPb pb.GetMomentsReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.Pagination = &commonPb.RequestPagination{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	respPb, err := client.GetMoments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Moments, &respPb.Moments)

	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.MomentsNums = int64(respPb.MomentsNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func DeleteMoments(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteMomentsRequest
		reqPb pb.DeleteMomentsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, req)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.DeleteMoments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func AlterMoment(c *gin.Context) {

	var (
		req   cms_api_struct.AlterMomentRequest
		resp  cms_api_struct.AlterMomentResponse
		reqPb pb.AlterMomentReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.AlterMoment(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func ChangeMomentStatus(c *gin.Context) {

	var (
		req   cms_api_struct.ChangeMomentStatusRequest
		resp  cms_api_struct.ChangeMomentStatusResponse
		reqPb pb.ChangeMomentStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.ChangeMomentStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func ModifyVisibility(c *gin.Context) {

	var (
		req   cms_api_struct.ModifyVisibilityRequest
		resp  cms_api_struct.ModifyVisibilityResponse
		reqPb pb.ModifyVisibilityReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.ModifyVisibility(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetMomentDetails(c *gin.Context) {

	var (
		req   cms_api_struct.GetMomentDetailsRequest
		resp  cms_api_struct.GetMomentDetailsResponse
		reqPb pb.GetMomentDetailsReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.Pagination = &commonPb.RequestPagination{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	respPb, err := client.GetMomentDetails(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.MomentDetail, &respPb.MomentDetails)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.MomentsNums = int64(respPb.MomentsNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func CtlMomentComment(c *gin.Context) {

	var (
		req   cms_api_struct.CtlMomentCommentRequest
		resp  cms_api_struct.CtlMomentCommentResponse
		reqPb pb.CtlMomentCommentReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.CtlMomentComment(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func GetComments(c *gin.Context) {

	var (
		req   cms_api_struct.GetCommentsRequest
		resp  cms_api_struct.GetCommentsResponse
		reqPb pb.GetCommentsReq
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
		return
	}
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	respPb, err := client.GetComments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	_ = utils.CopyStructFields(&resp.Comments, &respPb.Comments)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.CommentsNums = int64(respPb.CommentsNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func RemoveComments(c *gin.Context) {

	var (
		req   cms_api_struct.RemoveCommentsRequest
		reqPb pb.RemoveCommentsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.RemoveComments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterComment(c *gin.Context) {

	var (
		req   cms_api_struct.AlterCommentRequest
		reqPb pb.AlterCommentReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.AlterComment(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SwitchCommentHideState(c *gin.Context) {

	var (
		req   cms_api_struct.SwitchCommentHideStateRequest
		reqPb pb.SwitchCommentHideStateReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.SwitchCommentHideState(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetReplayComments(c *gin.Context) {

	var (
		req   cms_api_struct.GetReplayCommentsRequest
		resp  cms_api_struct.GetReplayCommentsResponse
		reqPb pb.GetCommentsReq
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
		return
	}
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	reqPb.CommentType = "2"

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	respPb, err := client.GetComments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	_ = utils.CopyStructFields(&resp.Comments, &respPb.Comments)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.CommentsNums = int64(respPb.CommentsNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetLikes(c *gin.Context) {

	var (
		req   cms_api_struct.GetLikesRequest
		resp  cms_api_struct.GetLikesResponse
		reqPb pb.GetLikesReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.Pagination = &commonPb.RequestPagination{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	respPb, err := client.GetLikes(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Likes, &respPb.Likes)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.LikesNums = int64(respPb.LikeNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func RemoveLikes(c *gin.Context) {

	var (
		req   cms_api_struct.RemoveLikesRequest
		reqPb pb.RemoveLikesReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.RemoveLikes(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SwitchLikeHideState(c *gin.Context) {

	var (
		req   cms_api_struct.SwitchLikeHideStateRequest
		reqPb pb.SwitchLikeHideStateReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewMomentsClient(etcdConn)
	_, err := client.SwitchLikeHideState(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}
