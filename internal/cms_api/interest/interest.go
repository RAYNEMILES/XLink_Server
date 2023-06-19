package interest

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/admin_cms"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetInterests(c *gin.Context) {

	var (
		req   cms_api_struct.GetInterestsRequest
		resp  cms_api_struct.GetInterestsResponse
		reqPb pb.GetInterestsReq
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	log.Debug("xx", "begin copy")
	utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	log.Debug("xx", "Copy ok Get conn")
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.Interests, &respPb.Interests)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.InterestNums = respPb.InterestNums
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func DeleteInterests(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteInterestsRequest
		reqPb pb.DeleteInterestsReq
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
	reqPb.Interests = req.Interests

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func AlterInterests(c *gin.Context) {
	var (
		req   cms_api_struct.AlterInterestsRequest
		resp  cms_api_struct.AlterInterestsResponse
		reqPb pb.AlterInterestReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AlterInterest(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func ChangeInterestStatus(c *gin.Context) {

	var (
		req   cms_api_struct.ChangeInterestStatusRequest
		resp  cms_api_struct.ChangeInterestStatusResponse
		reqPb pb.ChangeInterestStatusReq
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
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.ChangeInterestStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func AddInterests(c *gin.Context) {
	var (
		req   cms_api_struct.AddInterestsRequest
		resp  cms_api_struct.AddInterestsResponse
		reqPb pb.AddInterestsReq
	)

	if err := c.BindJSON(&req); err != nil {
		log.Debug("0", "should bind query error, request is: ", req)
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
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
	reqPb.Interests = []*pb.InterestReq{}
	for _, interest := range req.Interests {
		interestPb := pb.InterestReq{}
		utils.CopyStructFields(&interestPb, &interest)
		reqPb.Interests = append(reqPb.Interests, &interestPb)
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AddInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AddOneInterest(c *gin.Context) {
	var (
		req   cms_api_struct.AddOneInterestRequest
		resp  cms_api_struct.AddOneInterestResponse
		reqPb pb.AddInterestsReq
	)

	if err := c.BindJSON(&req); err != nil {
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
	interestPb := pb.InterestReq{}
	utils.CopyStructFields(&interestPb, &req)
	reqPb.Interests = []*pb.InterestReq{}
	reqPb.Interests = append(reqPb.Interests, &interestPb)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AddInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetUserInterests(c *gin.Context) {

	var (
		req   cms_api_struct.GetUserInterestsRequest
		resp  cms_api_struct.GetUserInterestsResponse
		reqPb pb.GetUserInterestsReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.Pagination = &commonPb.RequestPagination{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetUserInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Interests, &respPb.Interests)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.InterestNums = respPb.InterestNums
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AlterUserInterests(c *gin.Context) {

	var (
		req   cms_api_struct.AlterUserInterestsRequest
		reqPb pb.AlterUserInterestsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
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
	reqPb.OpUserId = userID

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	_, err := client.AlterUserInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func DeleteUserInterests(c *gin.Context) {

	var (
		req   cms_api_struct.DeleteUserInterestsRequest
		reqPb pb.DeleteUserInterestsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
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
	_, err := client.DeleteUserInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}

func GetGroupInterests(c *gin.Context) {

	var (
		req   cms_api_struct.GetGroupInterestsRequest
		resp  cms_api_struct.GetGroupInterestsResponse
		reqPb pb.GetGroupInterestsReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.Pagination = &commonPb.RequestPagination{}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetGroupInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}

	utils.CopyStructFields(&resp.Interests, &respPb.Interests)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.InterestNums = respPb.InterestNums
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func AlterGroupInterests(c *gin.Context) {

	var (
		req   cms_api_struct.AlterGroupInterestsRequest
		reqPb pb.AlterGroupInterestsReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
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
	_, err := client.AlterGroupInterests(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)

}
