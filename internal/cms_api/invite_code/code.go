package inviteCode

import (
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils3 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AddCode(c *gin.Context) {
	var (
		apiReq cms_api_struct.AddInviteCodeRequest

		rpcReq pbAdmin.AddInviteCodeRequest
	)

	if err := c.BindJSON(&apiReq); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiReq.OperationID, utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiReq.OperationID)
	if userID == "" {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, apiReq.OperationID)
	if etcdConn == nil {
		errMsg := apiReq.OperationID + "getcdv3.GetConn == nil"
		log.NewError(apiReq.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	rpcReq.OperationID = apiReq.OperationID
	rpcReq.UserId = apiReq.UserId

	utils.CopyStructFields(&rpcReq, &apiReq)

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	res, err1 := client.AddInviteCode(context.Background(), &rpcReq)
	if err1 != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "AddInviteCode failed ", err1)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: res.CommonResp.ErrCode, ErrMsg: res.CommonResp.ErrMsg}, nil)
	return
}

func EditCode(c *gin.Context) {
	var (
		apiReq cms_api_struct.EditInviteCodeRequest
	)

	if err := c.BindJSON(&apiReq); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiReq.OperationID, utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiReq.OperationID)
	if userID == "" {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	// code is exist
	if imdb.CodeIsExpired(apiReq.Code) == false {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "code is exist!", apiReq.Code)
		openIMHttp.RespHttp200(c, constant.ErrEditInviteCodeIsNotExist, nil)
		return
	}

	var data db.InviteCode
	utils.CopyStructFields(&data, &apiReq)
	result := imdb.EditInviteCodeByCode(apiReq.Code, data)
	if result == false {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "EditInviteCodeByCode failed!", apiReq.Code)
		openIMHttp.RespHttp200(c, constant.ErrEditInviteCodeFailed, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.OK, nil)
	return
}

func Switch(c *gin.Context) {
	var (
		apiRequest cms_api_struct.SwitchRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiRequest.OperationID)
	if userID == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := apiRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	SetInviteCodeSwitchRequest := &pbAdmin.SetInviteCodeSwitchRequest{
		OperationID:    apiRequest.OperationID,
		State:          apiRequest.State,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.SetInviteCodeSwitch(context.Background(), SetInviteCodeSwitchRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: rpcRes.CommonResp.ErrCode, ErrMsg: rpcRes.CommonResp.ErrMsg}, nil)
	return
}

func Limit(c *gin.Context) {
	var (
		apiRequest cms_api_struct.LimitRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiRequest.OperationID)
	if userID == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := apiRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	SetInviteCodeLimitRequest := &pbAdmin.SetInviteCodeLimitRequest{
		OperationID:    apiRequest.OperationID,
		State:          apiRequest.State,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.SetInviteCodeLimit(context.Background(), SetInviteCodeLimitRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: rpcRes.CommonResp.ErrCode, ErrMsg: rpcRes.CommonResp.ErrMsg}, nil)
	return
}

func MultiDelete(c *gin.Context) {
	var (
		apiRequest cms_api_struct.MultiDeleteRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiRequest.OperationID)
	if userID == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := apiRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	MultiDeleteInviteCodeRequest := &pbAdmin.MultiDeleteInviteCodeRequest{
		OperationID:    apiRequest.OperationID,
		Code:           apiRequest.InviteCodes,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.MultiDeleteInviteCode(context.Background(), MultiDeleteInviteCodeRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: rpcRes.CommonResp.ErrCode, ErrMsg: rpcRes.CommonResp.ErrMsg}, nil)
	return
}

func SwitchCodeState(c *gin.Context) {
	var (
		apiReq cms_api_struct.SwitchInviteCodeStateRequest
	)

	if err := c.BindJSON(&apiReq); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiReq.OperationID, utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, apiReq.OperationID)
	if userID == "" {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	code := imdb.CodeIsExpired(apiReq.InviteCode)
	if code == false {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "SwitchCodeState failed ", code)
		openIMHttp.RespHttp200(c, constant.ErrInviteCodeInexistence, nil)
		return
	}

	res := imdb.SwitchCodeState(apiReq.InviteCode, apiReq.State)
	if res == false {
		log.NewError(apiReq.OperationID, utils.GetSelfFuncName(), "SwitchCodeState failed ", res)
		openIMHttp.RespHttp200(c, constant.ErrInviteCode, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, res)
}

func GetCodeList(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetInviteCodeListRequest
		apiResponse cms_api_struct.GetInviteCodeListResponse
	)
	if err := c.ShouldBind(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), apiRequest, err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}
	token := c.GetHeader("token")
	if token == "" {
		log.NewError("", utils2.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, userID, _ := token_verify.GetUserIDFromToken(token, "")
	if userID == "" {
		log.NewError("", utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp200(c, constant.ErrTokenInvalid, nil)
		return
	}

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError("", utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := apiRequest.OperationID + "getcdv3.GetConn == nil"
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := pbAdmin.NewAdminCMSClient(etcdConn)
	GetInviteCodeListRequest := pbAdmin.GetInviteCodeListRequest{
		OperationID: apiRequest.OperationID,
		Pagination: &sdk.RequestPagination{
			PageNumber: int32(apiRequest.RequestPagination.PageNumber),
			ShowNumber: int32(apiRequest.RequestPagination.ShowNumber),
		},
		Conditions: &pbAdmin.GetInviteCodeListConditions{
			Code:     apiRequest.Code,
			UserName: apiRequest.UserName,
			UserId:   apiRequest.UserId,
			State:    utils.IntToString(apiRequest.State),
			Note:     apiRequest.Note,
			OrderBy:  apiRequest.OrderBy,
		},
	}
	rpcRes, rpcErr := client.GetInviteCodeList(context.Background(), &GetInviteCodeListRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	utils.CopyStructFields(&apiResponse, &rpcRes)
	apiResponse.BaseLink = rpcRes.InviteCodeBaseLink
	apiResponse.CurrentPage = int(rpcRes.CurrentNumber)

	var InviteCodeCode []cms_api_struct.InviteCodeCode
	for _, code := range rpcRes.List {
		InviteCodeCode = append(InviteCodeCode, cms_api_struct.InviteCodeCode{
			ID:       utils.StringToInt(code.Id),
			Code:     code.Code,
			UserId:   code.UserId,
			UserName: code.UserName,
			State:    int32(utils.StringToInt(code.State)),
			Note:     code.Note,
			Greeting: code.Greeting,
		})
	}
	apiResponse.InviteCodeList = InviteCodeCode

	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
}
