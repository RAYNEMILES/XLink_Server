package news

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	momentPb "Open_IM/pkg/proto/moments"
	pb "Open_IM/pkg/proto/news"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetOfficialAccounts(c *gin.Context) {

	var (
		req   cms_api_struct.GetOfficialAccountsRequest
		resp  cms_api_struct.GetOfficialAccountsResponse
		reqPb pb.GetOfficialAccountsReq
	)
	if err := c.ShouldBindQuery(&req); err != nil {
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

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	respPb, err := client.GetOfficialAccounts(context.Background(), &reqPb)
	if err != nil {
		log.NewError("", "db err: ", err.Error())
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Official, &respPb.OfficialAccount)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.OfficialNums = int64(respPb.OfficialNums)
	resp.PendingNums = respPb.PendingNums

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func DeleteOfficialAccounts(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteOfficialAccountsRequest
		reqPb pb.DeleteOfficialAccountsReq
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

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.DeleteOfficialAccounts(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterOfficialAccount(c *gin.Context) {

	var (
		req   cms_api_struct.AlterOfficialAccountRequest
		reqPb pb.AlterOfficialAccountReq
		resp  cms_api_struct.AlterOfficialAccountResponse
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "AlterOfficialAccount xx BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserID = userID

	reqPb.Official = &pb.Official{}
	utils.CopyStructFields(&reqPb.Official, &req)
	reqPb.Interests = req.Interests
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	rpcResp, err := client.AlterOfficialAccount(context.Background(), &reqPb)
	if err != nil {
		if rpcResp != nil {
			resp.CommResp.ErrCode = rpcResp.CommonResp.ErrCode
			resp.CommResp.ErrMsg = rpcResp.CommonResp.ErrMsg
		}
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	resp.CommResp.ErrCode = rpcResp.CommonResp.ErrCode
	resp.CommResp.ErrMsg = rpcResp.CommonResp.ErrMsg

	if resp.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.CommResp.ErrCode, "err_msg": resp.CommResp.ErrMsg, "data": resp})
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AddOfficialAccount(c *gin.Context) {
	var (
		req   cms_api_struct.AddOfficialAccountRequest
		reqPb pb.AddOfficialAccountReq
		resp  cms_api_struct.AddOfficialAccountResponse
	)

	if err := c.BindJSON(&req); err != nil {
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	if req.InitialNickname == "" {
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	if len(req.Nickname) > 36 {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrNickNameLength.ErrCode, "errMsg": constant.ErrNickNameLength.ErrMsg})
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	reqPb.OpUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	rpcResp, err := client.AddOfficialAccount(context.Background(), &reqPb)
	if err != nil {
		if rpcResp != nil {
			resp.CommResp.ErrCode = rpcResp.CommonResp.ErrCode
			resp.CommResp.ErrMsg = rpcResp.CommonResp.ErrMsg
		}
		c.JSON(http.StatusBadRequest, gin.H{"err_code": resp.CommResp.ErrCode, "err_msg": resp.CommResp.ErrMsg})
		return
	}

	resp.CommResp.ErrCode = rpcResp.CommonResp.ErrCode
	resp.CommResp.ErrMsg = rpcResp.CommonResp.ErrMsg
	if resp.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.CommResp.ErrCode, "err_msg": resp.CommResp.ErrMsg, "data": resp})
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func Process(c *gin.Context) {
	var (
		req   cms_api_struct.ProcessRequest
		resp  cms_api_struct.ProcessResponse
		reqPb pb.ProcessReq
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
	reqPb.OpUserId = userID
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.Process(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func GetNews(c *gin.Context) {

	var (
		req   cms_api_struct.GetNewsRequest
		resp  cms_api_struct.GetNewsResponse
		reqPb pb.GetNewsReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	respPb, err := client.GetNews(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Articles, &respPb.Articles)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.NewsNums = int64(respPb.ArticlesNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func DeleteNews(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteNewsRequest
		reqPb pb.DeleteNewsReq
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
	reqPb.Articles = req.Articles

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.DeleteNews(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterNews(c *gin.Context) {

	var (
		req   cms_api_struct.AlterNewsRequest
		resp  cms_api_struct.AlterNewsResponse
		reqPb pb.AlterNewsReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.AlterNews(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func ChangePrivacy(c *gin.Context) {

	var (
		req   cms_api_struct.ChangePrivacyRequest
		resp  cms_api_struct.ChangePrivacyResponse
		reqPb pb.ChangePrivacyReq
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

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.ChangePrivacy(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetNewsComments(c *gin.Context) {

	var (
		req   cms_api_struct.GetNewsCommentsRequest
		resp  cms_api_struct.GetNewsCommentsResponse
		reqPb pb.GetNewsCommentsReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	respPb, err := client.GetNewsComments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Comments, &respPb.Comments)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.CommentsNums = int64(respPb.CommentsNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func RemoveNewsComments(c *gin.Context) {

	var (
		req   cms_api_struct.RemoveNewsCommentsRequest
		reqPb pb.RemoveNewsCommentsReq
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

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.RemoveNewsComments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AlterNewsComment(c *gin.Context) {

	var (
		req   cms_api_struct.AlterNewsCommentRequest
		resp  cms_api_struct.AlterNewsCommentResponse
		reqPb pb.AlterNewsCommentReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.AlterNewsComment(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func ChangeNewsCommentStatus(c *gin.Context) {

	var (
		req   cms_api_struct.ChangeNewsCommentStatusRequest
		resp  cms_api_struct.ChangeNewsCommentStatusResponse
		reqPb pb.ChangeNewsCommentStatusReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.ChangeNewsCommentStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetNewsLikes(c *gin.Context) {

	var (
		req   cms_api_struct.GetNewsLikesRequest
		resp  cms_api_struct.GetNewsLikesResponse
		reqPb pb.GetNewsLikesReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	respPb, err := client.GetNewsLikes(context.Background(), &reqPb)
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

func RemoveNewsLikes(c *gin.Context) {

	var (
		req   cms_api_struct.RemoveNewsLikesRequest
		reqPb pb.RemoveNewsLikesReq
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

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.RemoveNewsLikes(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func ChangeNewsLikeStatus(c *gin.Context) {

	var (
		req   cms_api_struct.ChangeNewsLikeStatusRequest
		resp  cms_api_struct.ChangeNewsLikeStatusResponse
		reqPb pb.ChangeNewsLikeStatusReq
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

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.ChangeNewsLikeStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetRepostArticles(c *gin.Context) {

	var (
		req   cms_api_struct.GetRepostArticlesRequest
		resp  cms_api_struct.GetRepostArticlesResponse
		reqPb pb.GetRepostArticlesReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	respPb, err := client.GetRepostArticles(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Repost, &respPb.Reposts)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.RepostNums = int64(respPb.RepostNums)

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func ChangeRepostPrivacy(c *gin.Context) {

	var (
		req   cms_api_struct.ChangeRepostPrivacyRequest
		reqPb momentPb.ModifyVisibilityReq
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

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := momentPb.NewMomentsClient(etcdConn)
	_, err := client.ModifyVisibility(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func DeleteReposts(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteRepostsRequest
		reqPb momentPb.DeleteMomentsReq
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
	reqPb.Moments = req.MomentIds
	reqPb.ArticleIDs = req.Articles
	//reqPb.MomentsType = make([]int32, len(reqPb.Moments))
	//for index, _ := range reqPb.Moments {
	//	reqPb.MomentsType[index] = 1
	//}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImMoemntsName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := momentPb.NewMomentsClient(etcdConn)
	_, err := client.DeleteMoments(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func GetOfficialFollowers(c *gin.Context) {

	var (
		req    cms_api_struct.GetOfficialFollowersRequest
		resp   cms_api_struct.GetOfficialFollowersResponse
		rpcReq pb.GetOfficialFollowersReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, req)
	rpcReq.Pagination = &commonPb.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	rpcResp, err := client.GetOfficialFollowers(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	_ = utils.CopyStructFields(&resp.OfficialFollowers, rpcResp.OfficialFollowers)
	resp.ShowNumber = int(rpcResp.Pagination.ShowNumber)
	resp.CurrentPage = int(rpcResp.Pagination.CurrentPage)
	resp.OfficialFollowersCount = rpcResp.OfficialFollowersCount

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func BlockFollower(c *gin.Context) {

	var (
		req  cms_api_struct.BlockFollowerRequest
		resp cms_api_struct.BlockFollowerResponse

		rpcReq pb.BlockFollowerReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.BlockFollower(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func MuteFollower(c *gin.Context) {

	var (
		req    cms_api_struct.MuteFollowerRequest
		resp   cms_api_struct.MuteFollowerResponse
		rpcReq pb.MuteFollowerReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, &req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	rpcReq.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.MuteFollower(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func RemoveFollowers(c *gin.Context) {

	var (
		req  cms_api_struct.RemoveFollowersRequest
		resp cms_api_struct.RemoveFollowersResponse

		rpcReq pb.RemoveFollowersReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, &req)
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}
	rpcReq.OpUserID = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImNewsName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewNewsClient(etcdConn)
	_, err := client.RemoveFollowers(context.Background(), &rpcReq)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
}

