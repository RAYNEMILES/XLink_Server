package shortVideo

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/admin_cms"
	pb "Open_IM/pkg/proto/admin_cms"
	server_api_params "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetShortVideoList(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetShortVideoListRequest
		apiResponse cms_api_struct.GetShortVideoListResponse
		rpcRequest  admin_cms.ManagementShortVideoRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindQuery(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)
	rpcRequest.Pagination = &server_api_params.RequestPagination{PageNumber: int32(apiRequest.PageNumber), ShowNumber: int32(apiRequest.ShowNumber)}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.ManagementShortVideo(context.Background(), &rpcRequest)
	if err != nil {
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	_ = utils.CopyStructFields(&apiResponse.ShortVideoList, &respPb.ShortVideoInfo)
	apiResponse.ShowNumber = int(respPb.Pagination.ShowNumber)
	apiResponse.CurrentPage = int(respPb.Pagination.CurrentPage)
	apiResponse.ShortVideoCount = respPb.TotalCount

	for i, v := range apiResponse.ShortVideoList {
		if v.InterestId != "" {
			stringArr := strings.Split(v.InterestId, ",")
			for _, v1 := range stringArr {
				apiResponse.ShortVideoList[i].InterestArray = append(apiResponse.ShortVideoList[i].InterestArray, utils.StringToInt64(v1))
			}
		}
	}

	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func DeleteShortVideo(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.DeleteShortVideoRequest
		apiResponse cms_api_struct.DeleteShortVideoResponse
		rpcRequest  admin_cms.DeleteShortVideoRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.DeleteShortVideo(context.Background(), &rpcRequest)
	if err != nil {
		if respPb != nil {
			apiResponse.CommResp.ErrCode = respPb.CommonResp.ErrCode
			apiResponse.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
		}
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	apiResponse.ErrMsg = respPb.CommonResp.ErrMsg
	apiResponse.ErrCode = respPb.CommonResp.ErrCode

	if apiResponse.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": apiResponse.ErrCode, "err_msg": apiResponse.ErrMsg, "data": apiResponse})
		return
	}

	_ = utils.CopyStructFields(&apiResponse, &respPb)
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func AlterShortVideo(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.AlterShortVideoRequest
		apiResponse cms_api_struct.AlterShortVideoResponse
		rpcRequest  admin_cms.AlterShortVideoRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.AlterShortVideo(context.Background(), &rpcRequest)
	if err != nil {
		if respPb != nil {
			apiResponse.CommResp.ErrCode = respPb.CommonResp.ErrCode
			apiResponse.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
		}
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	apiResponse.ErrMsg = respPb.CommonResp.ErrMsg
	apiResponse.ErrCode = respPb.CommonResp.ErrCode

	if apiResponse.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": apiResponse.ErrCode, "err_msg": apiResponse.ErrMsg, "data": apiResponse})
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func GetShortVideoLikeList(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetShortVideoLikeListRequest
		apiResponse cms_api_struct.GetShortVideoLikeListResponse
		rpcRequest  admin_cms.GetShortVideoLikeListRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindQuery(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)
	rpcRequest.Pagination = &server_api_params.RequestPagination{PageNumber: int32(apiRequest.PageNumber), ShowNumber: int32(apiRequest.ShowNumber)}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetShortVideoLikeList(context.Background(), &rpcRequest)
	if err != nil {
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	_ = utils.CopyStructFields(&apiResponse.ShortVideoLikeList, &respPb.ShortVideoLike)
	apiResponse.ShowNumber = int(respPb.Pagination.ShowNumber)
	apiResponse.CurrentPage = int(respPb.Pagination.CurrentPage)
	apiResponse.ShortVideoLikeCount = respPb.TotalCount
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func DeleteShortVideoLike(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.DeleteShortVideoLikeRequest
		apiResponse cms_api_struct.DeleteShortVideoLikeResponse
		rpcRequest  admin_cms.DeleteShortVideoLikeRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.DeleteShortVideoLike(context.Background(), &rpcRequest)
	if err != nil {
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	apiResponse.ErrMsg = respPb.CommonResp.ErrMsg
	apiResponse.ErrCode = respPb.CommonResp.ErrCode

	if apiResponse.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": apiResponse.ErrCode, "err_msg": apiResponse.ErrMsg, "data": apiResponse})
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func GetShortVideoCommentList(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetShortVideoCommentListRequest
		apiResponse cms_api_struct.GetShortVideoCommentListResponse
		rpcRequest  admin_cms.GetShortVideoCommentListRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindQuery(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)
	rpcRequest.Pagination = &server_api_params.RequestPagination{PageNumber: int32(apiRequest.PageNumber), ShowNumber: int32(apiRequest.ShowNumber)}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	rpcRequest.Content = apiRequest.Context
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetShortVideoCommentList(context.Background(), &rpcRequest)
	if err != nil {
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	_ = utils.CopyStructFields(&apiResponse.ShortVideoCommentList, &respPb.ShortVideoComment)
	apiResponse.ShowNumber = int(respPb.Pagination.ShowNumber)
	apiResponse.CurrentPage = int(respPb.Pagination.CurrentPage)
	apiResponse.ShortVideoCommentCount = respPb.TotalCount
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func DeleteShortVideoComment(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.DeleteShortVideoCommentRequest
		apiResponse cms_api_struct.DeleteShortVideoCommentResponse
		rpcRequest  admin_cms.DeleteShortVideoCommentRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.DeleteShortVideoComment(context.Background(), &rpcRequest)
	if err != nil {
		if respPb != nil {
			apiResponse.CommResp.ErrCode = respPb.CommonResp.ErrCode
			apiResponse.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
		}
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	apiResponse.ErrMsg = respPb.CommonResp.ErrMsg
	apiResponse.ErrCode = respPb.CommonResp.ErrCode

	if apiResponse.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": apiResponse.ErrCode, "err_msg": apiResponse.ErrMsg, "data": apiResponse})
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func AlterShortVideoComment(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.AlterShortVideoCommentRequest
		apiResponse cms_api_struct.AlterShortVideoCommentResponse
		rpcRequest  admin_cms.AlterShortVideoCommentRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.AlterShortVideoComment(context.Background(), &rpcRequest)
	if err != nil {
		if respPb != nil {
			apiResponse.CommResp.ErrCode = respPb.CommonResp.ErrCode
			apiResponse.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
		}
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	apiResponse.ErrMsg = respPb.CommonResp.ErrMsg
	apiResponse.ErrCode = respPb.CommonResp.ErrCode

	if apiResponse.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": apiResponse.ErrCode, "err_msg": apiResponse.ErrMsg, "data": apiResponse})
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func GetShortVideoInterestLabelList(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetShortVideoInterestLabelListRequest
		apiResponse cms_api_struct.GetShortVideoInterestLabelListResponse
		rpcRequest  admin_cms.GetShortVideoInterestLabelListRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindQuery(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)
	rpcRequest.Pagination = &server_api_params.RequestPagination{PageNumber: int32(apiRequest.PageNumber), ShowNumber: int32(apiRequest.ShowNumber)}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetShortVideoInterestLabelList(context.Background(), &rpcRequest)
	if err != nil {
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	_ = utils.CopyStructFields(&apiResponse.ShortVideoList, &respPb.ShortVideoInterestLabel)
	apiResponse.ShowNumber = int(respPb.Pagination.ShowNumber)
	apiResponse.CurrentPage = int(respPb.Pagination.CurrentPage)
	apiResponse.ShortVideoCount = respPb.TotalCount
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func AlterShortVideoInterestLabel(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.AlterShortVideoInterestLabelRequest
		apiResponse cms_api_struct.AlterShortVideoInterestLabelResponse
		rpcRequest  admin_cms.AlterShortVideoInterestLabelRequest
	)
	//check the params from request
	apiRequest.OperationID = utils.OperationIDGenerator()
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), err.Error())
		apiResponse.ErrMsg = constant.ErrArgs.ErrMsg
		apiResponse.ErrCode = constant.ErrArgs.ErrCode
		openIMHttp.RespHttp200(c, constant.ErrArgs, apiResponse)
		return
	}

	_ = utils.CopyStructFields(&rpcRequest, &apiRequest)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.AlterShortVideoInterestLabel(context.Background(), &rpcRequest)
	if err != nil {
		if respPb != nil {
			apiResponse.CommResp.ErrCode = respPb.CommonResp.ErrCode
			apiResponse.CommResp.ErrMsg = respPb.CommonResp.ErrMsg
		}
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	apiResponse.ErrMsg = respPb.CommonResp.ErrMsg
	apiResponse.ErrCode = respPb.CommonResp.ErrCode

	c.JSON(http.StatusBadRequest, gin.H{"code": apiResponse.ErrCode, "err_msg": apiResponse.ErrMsg, "data": apiResponse})
}

func GetShortVideoCommentReplies(c *gin.Context) {

	var (
		req   cms_api_struct.GetShortVideoCommentRepliesRequest
		resp  cms_api_struct.GetShortVideoCommentRepliesResponse
		reqPb pb.GetShortVideoCommentRepliesReq
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

	reqPb.Pagination = &server_api_params.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}

	_ = utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	rpcResp, err := client.GetShortVideoCommentReplies(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	_ = utils.CopyStructFields(&resp.CommentReplies, rpcResp.CommentReplies)

	resp.RepliesCount = rpcResp.RepliesCount
	resp.ResponsePagination = cms_api_struct.ResponsePagination{
		CurrentPage: int(rpcResp.Pagination.CurrentPage),
		ShowNumber:  int(rpcResp.Pagination.ShowNumber),
	}
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AlterReply(c *gin.Context) {

	var (
		req   cms_api_struct.AlterReplyRequest
		resp  cms_api_struct.AlterReplyResponse
		reqPb pb.AlterReplyReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AlterReply(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeleteReplies(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteRepliesRequest
		resp  cms_api_struct.DeleteRepliesResponse
		reqPb pb.DeleteRepliesReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteReplies(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetShortVideoCommentLikes(c *gin.Context) {

	var (
		req   cms_api_struct.GetShortVideoCommentLikesRequest
		resp  cms_api_struct.GetShortVideoCommentLikesResponse
		reqPb pb.GetShortVideoCommentLikesReq
	)
	if err := c.BindQuery(&req); err != nil {
		fmt.Println("BindJSON failed ", err.Error())
		resp.ErrCode = constant.ErrArgs.ErrCode
		resp.ErrMsg = constant.ErrArgs.ErrMsg
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.Pagination = &server_api_params.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}

	_ = utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	rpcResp, err := client.GetShortVideoCommentLikes(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	_ = utils.CopyStructFields(&resp.CommentLikes, rpcResp.CommentLikes)

	resp.LikesCount = rpcResp.LikesCount
	resp.ResponsePagination = cms_api_struct.ResponsePagination{
		CurrentPage: int(rpcResp.Pagination.CurrentPage),
		ShowNumber:  int(rpcResp.Pagination.ShowNumber),
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AlterLike(c *gin.Context) {

	var (
		req   cms_api_struct.AlterLikeRequest
		resp  cms_api_struct.AlterLikeResponse
		reqPb pb.AlterLikeReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AlterLike(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeleteLikes(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteLikesRequest
		resp  cms_api_struct.DeleteLikesResponse
		reqPb pb.DeleteLikesReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteLikes(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetFollowers(c *gin.Context) {

	var (
		req    cms_api_struct.GetFollowersRequest
		resp   cms_api_struct.GetFollowersResponse
		rpcReq pb.GetFollowersReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, req)
	rpcReq.Pagination = &server_api_params.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	rpcResp, err := client.GetFollowers(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	_ = utils.CopyStructFields(&resp.Followers, rpcResp.Followers)
	resp.ShowNumber = int(rpcResp.Pagination.ShowNumber)
	resp.CurrentPage = int(rpcResp.Pagination.CurrentPage)
	resp.FollowersCount = rpcResp.FollowersCount

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AlterFollower(c *gin.Context) {

	var (
		req    cms_api_struct.AlterFollowerRequest
		resp   cms_api_struct.AlterFollowerResponse
		rpcReq pb.AlterFollowerReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AlterFollower(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func DeleteFollowers(c *gin.Context) {

	var (
		req    cms_api_struct.DeleteFollowersRequest
		resp   cms_api_struct.DeleteFollowersResponse
		rpcReq pb.DeleteFollowersReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteFollowers(context.Background(), &rpcReq)

	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

