package user

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	pb "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"fmt"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUserById(c *gin.Context) {
	var (
		req   cms_api_struct.GetUserRequest
		resp  cms_api_struct.GetUserResponse
		reqPb pb.GetUserByIdReq
	)
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pb.NewUserClient(etcdConn)
	respPb, err := client.GetUserById(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	if respPb.User.UserId == "" {
		openIMHttp.RespHttp200(c, constant.OK, nil)
		return
	}
	utils.CopyStructFields(&resp, respPb.User)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetUsersByName(c *gin.Context) {
	var (
		req   cms_api_struct.GetUsersByNameRequest
		resp  cms_api_struct.GetUsersByNameResponse
		reqPb pb.GetUsersByNameReq
	)
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.UserName = req.UserName
	reqPb.Pagination = &commonPb.RequestPagination{
		PageNumber: int32(req.PageNumber),
		ShowNumber: int32(req.ShowNumber),
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	respPb, err := client.GetUsersByName(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "rpc", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	utils.CopyStructFields(&resp.Users, respPb.Users)
	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
	resp.CurrentPage = int(respPb.Pagination.CurrentPage)
	resp.UserNums = int64(respPb.UserNums)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetUsers(c *gin.Context) {
	var (
		req   cms_api_struct.GetUsersRequest
		resp  cms_api_struct.GetUsersResponse
		reqPb pb.GetUsersRequest
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	reqPb.SourceCode = req.Code
	reqPb.Pagination.PageNumber = int32(req.PageNumber)
	reqPb.Pagination.ShowNumber = int32(req.ShowNumber)
	reqPb.LastLoginDevice = int32(req.LastLoginDevice)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	respPb, err := client.GetUsers(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	utils.CopyStructFields(&resp.Users, respPb.User)
	resp.ShowNumber = req.ShowNumber
	resp.CurrentPage = req.PageNumber
	resp.UserNums = int64(respPb.UserNums)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)

}

func GetUsersThirdInfo(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetUsersThirdInfoRequest
		apiResponse cms_api_struct.GetUsersThirdInfoResponse
		rpcRequest  pb.GetUsersThirdInfoRequest
	)
	rpcRequest.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&apiRequest); err != nil {
		log.NewError("0", "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	rpcRequest.OperationID = utils.OperationIDGenerator()

	utils.CopyStructFields(&rpcRequest, &apiRequest)
	rpcRequest.Pagination.PageNumber = int32(apiRequest.PageNumber)
	rpcRequest.Pagination.ShowNumber = int32(apiRequest.ShowNumber)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := rpcRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	respPb, err := client.GetUsersThirdInfo(context.Background(), &rpcRequest)
	if err != nil {
		openIMHttp.RespHttp200(c, err, apiResponse)
		return
	}

	log.NewInfo(rpcRequest.OperationID, utils.GetSelfFuncName(), "resp: ", respPb)

	utils.CopyStructFields(&apiResponse.Users, respPb.UserThirdInfo)
	apiResponse.ShowNumber = apiRequest.ShowNumber
	apiResponse.CurrentPage = apiRequest.PageNumber
	apiResponse.UserNums = int64(respPb.UserNums)
	log.NewInfo(apiRequest.OperationID, utils.GetSelfFuncName(), "resp: ", apiResponse)
	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}

func ResignUser(c *gin.Context) {
	var (
		req   cms_api_struct.ResignUserRequest
		resp  cms_api_struct.ResignUserResponse
		reqPb pb.ResignUserReq
	)
	if err := c.ShouldBind(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": http.StatusBadRequest, "errMsg": err.Error()})
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	_, err := client.ResignUser(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func AlterUser(c *gin.Context) {
	var (
		req   cms_api_struct.AlterUserRequest
		resp  cms_api_struct.AlterUserResponse
		reqPb pb.AlterUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	//get the userId from token

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	reqPb.SourceCode = req.Code
	_, err := client.AlterUser(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, "microserver failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func AddUser(c *gin.Context) {
	var (
		req   cms_api_struct.AddUserRequest
		resp  cms_api_struct.AddUserResponse
		reqPb pb.AddUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	req.PhoneNumber = utils.CompressStr(strings.TrimSpace(req.PhoneNumber))

	reqPb.OperationID = utils.OperationIDGenerator()

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	//check phone number
	if !strings.Contains(req.PhoneNumber, "+") {
		log.NewError(reqPb.OperationID, "The phone number should has + at the head, please check.")
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	official := im_mysql_model.GetChannelCodeListByOfficialAndState(constant.InviteChannelCodeStateValid)
	officialCode := ""
	if len(official) > 0 {
		officialCode = official[0].Code
	}

	// invite code is valid
	if utils.StringToInt(req.SourceId) == constant.UserRegisterSourceTypeInvite {
		if req.Code == "" {
			req.SourceId = utils.IntToString(constant.UserRegisterSourceTypeOfficial)
		} else {
			invite := im_mysql_model.GetCodeInfoByCode(req.Code)
			if invite == nil || invite.State != constant.InviteChannelCodeStateValid {
				openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrInviteCodeInvalid.ErrCode, ErrMsg: req.Code + " : " + constant.ErrInviteCodeInvalid.ErrMsg}, nil)
				return
			}
		}
	}

	// channel code is valid
	if utils.StringToInt(req.SourceId) == constant.UserRegisterSourceTypeChannel {
		if req.Code == "" {
			req.SourceId = utils.IntToString(constant.UserRegisterSourceTypeOfficial)
		} else {
			channel, _ := im_mysql_model.GetInviteChannelCodeByCode(req.Code)
			if channel == nil {
				openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrChannelCodeNotExist.ErrCode, ErrMsg: req.Code + " : " + constant.ErrChannelCodeNotExist.ErrMsg}, nil)
				return
			}
			if channel.State != constant.InviteChannelCodeStateValid {
				openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrChannelCodeInvalid.ErrCode, ErrMsg: req.Code + " : " + constant.ErrChannelCodeInvalid.ErrMsg}, nil)
				return
			}
		}
	}

	if utils.StringToInt(req.SourceId) == constant.UserRegisterSourceTypeOfficial {
		req.Code = officialCode
	}

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}

	reqPb.OpUserId = userID
	client := pb.NewUserClient(etcdConn)
	rpcResp, err := client.AddUser(context.Background(), &reqPb)
	if err != nil {
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		openIMHttp.RespHttp200(c, err, resp.CommResp)
		return
	}
	resp.ErrCode = rpcResp.CommonResp.ErrCode
	resp.ErrMsg = rpcResp.CommonResp.ErrMsg

	if resp.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": resp.CommResp.ErrCode, "err_msg": resp.CommResp.ErrMsg, "data": resp})
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func MultiAddUser(c *gin.Context) {
	var (
		req   cms_api_struct.MultiAddUserRequest
		reqPb pb.AddUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	for i, item := range req.Users {
		req.Users[i].PhoneNumber = utils.CompressStr(strings.TrimSpace(item.PhoneNumber))
		req.Users[i].UserId = utils.CompressStr(strings.TrimSpace(item.UserId))

		if !strings.Contains(req.Users[i].PhoneNumber, "+") {
			log.NewError(reqPb.OperationID, "The phone number should has + at the head, please check.")
			openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
			return
		}
	}

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	reqPb.OperationID = utils.OperationIDGenerator()
	reqPb.OpUserId = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}

	// get official code
	official := im_mysql_model.GetChannelCodeListByOfficialAndState(constant.InviteChannelCodeStateValid)
	officialCode := ""
	if len(official) > 0 {
		officialCode = official[0].Code
	}

	client := pb.NewUserClient(etcdConn)
	for i, user := range req.Users {
		// check the user id is existed
		idIsExist, _ := client.UserIdIsExist(context.Background(), &pb.UserIdIsExistRequest{
			UserId: user.UserId,
		})
		if idIsExist.IsExist == true {
			openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrUserExistUserIdExist.ErrCode, ErrMsg: user.UserId + " : " + constant.ErrUserExistUserIdExist.ErrMsg}, nil)
			return
		}

		// user is existed
		resp, _ := client.ExistsUser(context.Background(), &pb.ExistsUserRequest{
			OperationID: reqPb.OperationID,
			PhoneNumber: user.PhoneNumber,
			UserId:      user.UserId,
		})
		if resp != nil && resp.CommonResp.ErrCode != constant.OK.ErrCode {
			if resp.CommonResp.ErrCode == constant.ErrUserExistUserIdExist.ErrCode {
				openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: resp.CommonResp.ErrCode, ErrMsg: user.UserId + " : " + resp.CommonResp.ErrMsg}, nil)
				return
			}
			if resp.CommonResp.ErrCode == constant.ErrUserExistPhoneNumberExist.ErrCode {
				openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: resp.CommonResp.ErrCode, ErrMsg: user.PhoneNumber + " : " + resp.CommonResp.ErrMsg}, nil)
				return
			}
			openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: resp.CommonResp.ErrCode, ErrMsg: resp.CommonResp.ErrMsg}, nil)
			return
		}

		// invite code is valid
		if utils.StringToInt(user.SourceId) == constant.UserRegisterSourceTypeInvite {
			if user.Code == "" {
				req.Users[i].SourceId = utils.IntToString(constant.UserRegisterSourceTypeOfficial)
			} else {
				invite := im_mysql_model.GetCodeInfoByCode(user.Code)
				if invite == nil || invite.State != constant.InviteChannelCodeStateValid {
					openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrInviteCodeInvalid.ErrCode, ErrMsg: user.Code + " : " + constant.ErrInviteCodeInvalid.ErrMsg}, nil)
					return
				}
			}
		}

		// channel code is valid
		if utils.StringToInt(user.SourceId) == constant.UserRegisterSourceTypeChannel {
			if user.Code == "" {
				req.Users[i].SourceId = utils.IntToString(constant.UserRegisterSourceTypeOfficial)
			} else {
				channel, _ := im_mysql_model.GetInviteChannelCodeByCode(user.Code)
				if channel == nil {
					openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrChannelCodeNotExist.ErrCode, ErrMsg: user.Code + " : " + constant.ErrChannelCodeNotExist.ErrMsg}, nil)
					return
				}
				if channel.State != constant.InviteChannelCodeStateValid {
					openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: constant.ErrChannelCodeInvalid.ErrCode, ErrMsg: user.Code + " : " + constant.ErrChannelCodeInvalid.ErrMsg}, nil)
					return
				}
			}
		}

		if utils.StringToInt(user.SourceId) == constant.UserRegisterSourceTypeOfficial {
			req.Users[i].Code = officialCode
		}
	}

	for _, user := range req.Users {
		utils.CopyStructFields(&reqPb, &user)
		reqPb.OpUserId = userID
		client := pb.NewUserClient(etcdConn)
		_, _ = client.AddUser(context.Background(), &reqPb)
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
	return
}

func Exists(c *gin.Context) {
	var (
		request    cms_api_struct.ExistsUserRequest
		rpcRequest pb.ExistsUserRequest
	)
	if err := c.BindJSON(&request); err != nil {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	//get the userId from token
	token := c.GetHeader("token")
	if token == "" {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, request.OperationID)
	if userID == "" {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	//check the permissions
	if err := utils2.CheckAdminPermissions(userID); err != nil {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	utils.CopyStructFields(&rpcRequest, &request)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, request.OperationID)
	if etcdConn == nil {
		errMsg := request.OperationID + "getcdv3.GetConn == nil"
		log.NewError(request.OperationID, errMsg)
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}
	client := pb.NewUserClient(etcdConn)
	resp, rpcErr := client.ExistsUser(context.Background(), &rpcRequest)

	if rpcErr != nil {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: resp.CommonResp.ErrCode, ErrMsg: resp.CommonResp.ErrMsg}, nil)
	return
}

func BlockUser(c *gin.Context) {
	var (
		req   cms_api_struct.BlockUserRequest
		resp  cms_api_struct.BlockUserResponse
		reqPb pb.BlockUserReq
	)
	if err := c.BindJSON(&req); err != nil {
		fmt.Println(err)
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	fmt.Println(reqPb)
	_, err := client.BlockUser(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func UnblockUser(c *gin.Context) {
	var (
		req   cms_api_struct.UnblockUserRequest
		resp  cms_api_struct.UnBlockUserResponse
		reqPb pb.UnBlockUserReq
	)
	if err := c.ShouldBind(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb, &req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	_, err := client.UnBlockUser(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetBlockUsers(c *gin.Context) {
	var (
		req    cms_api_struct.GetBlockUsersRequest
		resp   cms_api_struct.GetBlockUsersResponse
		reqPb  pb.GetBlockUsersReq
		respPb *pb.GetBlockUsersResp
	)
	reqPb.Pagination = &commonPb.RequestPagination{}
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	utils.CopyStructFields(&reqPb.Pagination, &req)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "blockUsers", reqPb.Pagination, req)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	respPb, err := client.GetBlockUsers(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "GetBlockUsers rpc", err.Error())
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	for _, v := range respPb.BlockUsers {
		resp.BlockUsers = append(resp.BlockUsers, cms_api_struct.BlockUser{
			UserResponse: cms_api_struct.UserResponse{
				UserId:       v.User.UserId,
				ProfilePhoto: v.User.ProfilePhoto,
				Nickname:     v.User.Nickname,
				IsBlock:      v.User.IsBlock,
				CreateTime:   v.User.CreateTime,
			},
			BeginDisableTime: v.BeginDisableTime,
			EndDisableTime:   v.EndDisableTime,
		})
	}
	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
	resp.CurrentPage = int(respPb.Pagination.CurrentPage)
	resp.UserNums = int64(respPb.UserNums)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func GetBlockUserById(c *gin.Context) {
	var (
		req   cms_api_struct.GetBlockUserRequest
		resp  cms_api_struct.GetBlockUserResponse
		reqPb pb.GetBlockUserByIdReq
	)
	if err := c.ShouldBindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	reqPb.OperationID = utils.OperationIDGenerator()
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)

	reqPb.UserId = req.UserId
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	respPb, err := client.GetBlockUserById(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, "GetBlockUserById rpc failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp.EndDisableTime = respPb.BlockUser.EndDisableTime
	resp.BeginDisableTime = respPb.BlockUser.BeginDisableTime
	utils.CopyStructFields(&resp, respPb.BlockUser.User)
	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

func DeleteUser(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteUserRequest
		reqPb pb.DeleteUserReq
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

	reqPb.UserId = req.UserId
	reqPb.OpUserId = userID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	_, err := client.DeleteUser(context.Background(), &reqPb)
	if err != nil {
		log.NewError(reqPb.OperationID, "DeleteUser rpc failed ", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func SwitchStatus(c *gin.Context) {
	var (
		req   cms_api_struct.SwitchStatusRequest
		reqPb pb.SwitchStatusReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	reqPb.OperationID = utils.OperationIDGenerator()

	// get the userId from middleware
	userID := ""
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	//check the parameters
	if req.UserId == "" || req.Status < 1 || req.Status > 2 {
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
	err := utils.CopyStructFields(&reqPb, &req)
	if err != nil {
		return
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(reqPb.OperationID, errMsg)
		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
		return
	}

	reqPb.OpUserId = userID
	reqPb.UserId = req.UserId
	reqPb.Status = int32(req.Status)
	client := pb.NewUserClient(etcdConn)
	_, err = client.SwitchStatus(context.Background(), &reqPb)
	if err != nil {
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetDeletedUsers(c *gin.Context) {

	var (
		req    cms_api_struct.GetDeletedUsersRequest
		resp   cms_api_struct.GetDeletedUsersResponse
		rpcReq pb.GetDeletedUsersReq
	)
	if err := c.BindQuery(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		resp.ErrCode = constant.ErrArgs.ErrCode
		resp.ErrMsg = constant.ErrArgs.ErrMsg
		openIMHttp.RespHttp200(c, constant.ErrArgs, resp)
		return
	}
	rpcReq.OperationID = utils.OperationIDGenerator()
	log.NewInfo(rpcReq.OperationID, utils.GetSelfFuncName(), "req: ", req)

	_ = utils.CopyStructFields(&rpcReq, req)
	rpcReq.Pagination = &commonPb.RequestPagination{PageNumber: int32(req.PageNumber), ShowNumber: int32(req.ShowNumber)}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, rpcReq.OperationID)
	if etcdConn == nil {
		errMsg := rpcReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(rpcReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pb.NewUserClient(etcdConn)
	rpcResp, err := client.GetDeletedUsers(context.Background(), &rpcReq)
	if err != nil {
		s, _ := status.FromError(err)
		resp.ErrCode = int32(s.Code())
		resp.ErrMsg = s.Message()
		openIMHttp.RespHttp200(c, err, resp)
		return
	}
	_ = utils.CopyStructFields(&resp.DeletedUsers, rpcResp.DeletedUsers)
	resp.ShowNumber = int(rpcResp.Pagination.ShowNumber)
	resp.CurrentPage = int(rpcResp.Pagination.CurrentPage)
	resp.DeletedUsersCount = rpcResp.DeletedUsersCount

	openIMHttp.RespHttp200(c, constant.OK, resp)
}

// AlterAddFriendStatus
// func AlterAddFriendStatus(c *gin.Context) {
// 	var (
// 		req   cms_api_struct.SwitchStatusRequest
// 		reqPb pb.SwitchStatusReq
// 	)
// 	if err := c.BindJSON(&req); err != nil {
// 		log.NewError(reqPb.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
// 		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
// 		return
// 	}

// 	reqPb.OperationID = utils.OperationIDGenerator()

// 	// get the userId from middleware
// 	userID := ""
// 	userIDInter, exsited := c.Get("userID")
// 	if exsited {
// 		userID = userIDInter.(string)
// 	}

// 	//check the parameters
// 	if req.UserId == "" || req.Status < 1 || req.Status > 2 {
// 		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
// 		return
// 	}

// 	log.NewInfo(reqPb.OperationID, utils.GetSelfFuncName(), "req: ", req)
// 	utils.CopyStructFields(&reqPb, &req)
// 	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, reqPb.OperationID)
// 	if etcdConn == nil {
// 		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
// 		log.NewError(reqPb.OperationID, errMsg)
// 		openIMHttp.RespHttp200(c, constant.ErrServer, nil)
// 		return
// 	}

// 	reqPb.OpUserId = userID
// 	reqPb.UserId = req.UserId
// 	reqPb.Status = int32(req.Status)
// 	client := pb.NewUserClient(etcdConn)
// 	_, err := client.AlterAddFriendStatus(context.Background(), &reqPb)
// 	if err != nil {
// 		openIMHttp.RespHttp200(c, err, nil)
// 		return
// 	}
// 	openIMHttp.RespHttp200(c, constant.OK, nil)
// }
