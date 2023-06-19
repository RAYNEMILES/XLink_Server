package inviteChannelCode

import (
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
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
	"time"

	"github.com/gin-gonic/gin"
)

func AddCode(c *gin.Context) {
	var (
		apiRequest cms_api_struct.AddInviteChannelCodeRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// if sourceId is channel, then check the code
	if apiRequest.SourceId == constant.UserRegisterSourceTypeChannel {
		if apiRequest.Code == "" {
			log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "code is nil")
			openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
			return
		}
	}

	// set code to default value
	if apiRequest.Code == "" {
		// TODO 如果操作频繁，会导致生成的code重复
		apiRequest.Code = utils.GenerateChannelCode(time.Now().Unix())
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is nil")
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

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp200(c, constant.ErrUserBanned, nil)
		return
	}

	channelCode, _ := im_mysql_model.GetInviteChannelAllCodeByCode(apiRequest.Code)
	if channelCode != nil {
		if channelCode.State == constant.InviteChannelCodeStateDelete {
			openIMHttp.RespHttp200(c, constant.ErrChannelCodeIsDelete, nil)
			return
		}
		openIMHttp.RespHttp200(c, constant.ErrChannelCodeInexistence, nil)
		return
	}

	// check the official channel，only one
	if apiRequest.SourceId == constant.UserRegisterSourceTypeOfficial {
		isExist := im_mysql_model.IsExistOfficialChannelCode()
		if isExist {
			openIMHttp.RespHttp200(c, constant.OnlyOneOfficialChannelCode, nil)
			return
		}
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
	AddChannelCodeRequest := &pbAdmin.AddChannelCodeRequest{
		OperationID:    apiRequest.OperationID,
		Code:           apiRequest.Code,
		Greeting:       apiRequest.Greeting,
		Note:           apiRequest.Note,
		FriendIds:      strings.Join(utils.RemoveDuplicatesAndEmpty(apiRequest.FriendId), ","),
		GroupIds:       strings.Join(utils.RemoveDuplicatesAndEmpty(apiRequest.GroupId), ","),
		SourceId:       apiRequest.SourceId,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.AddChannelCode(context.Background(), AddChannelCodeRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: rpcRes.CommonResp.ErrCode, ErrMsg: rpcRes.CommonResp.ErrMsg}, nil)
	return
}

func EditCode(c *gin.Context) {
	var (
		apiRequest cms_api_struct.AddInviteChannelCodeRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is nil")
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

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
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
	EditChannelCodeRequest := &pbAdmin.EditChannelCodeRequest{
		OperationID:    apiRequest.OperationID,
		Code:           apiRequest.Code,
		Greeting:       apiRequest.Greeting,
		Note:           apiRequest.Note,
		FriendIds:      strings.Join(utils.RemoveDuplicatesAndEmpty(apiRequest.FriendId), ","),
		GroupIds:       strings.Join(utils.RemoveDuplicatesAndEmpty(apiRequest.GroupId), ","),
		OperatorUserId: userID,
		SourceId:       apiRequest.SourceId,
	}
	rpcRes, rpcErr := client.EditChannelCode(context.Background(), EditChannelCodeRequest)

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
		apiRequest cms_api_struct.SwitchInviteChannelCodeRequest
	)
	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is nil")
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

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
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
	SwitchChannelCodeStateRequest := &pbAdmin.SwitchChannelCodeStateRequest{
		OperationID:    apiRequest.OperationID,
		Code:           apiRequest.Code,
		State:          utils.IntToString(apiRequest.State),
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.SwitchChannelCodeState(context.Background(), SwitchChannelCodeStateRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: rpcRes.CommonResp.ErrCode, ErrMsg: rpcRes.CommonResp.ErrMsg}, nil)
	return
}

func Switch(c *gin.Context) {
	var (
		apiRequest cms_api_struct.ChannelCodeSwitchRequest
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
	SetChannelCodeSwitchRequest := &pbAdmin.SetChannelCodeSwitchRequest{
		OperationID:    apiRequest.OperationID,
		State:          apiRequest.State,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.SetChannelCodeSwitch(context.Background(), SetChannelCodeSwitchRequest)

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
		apiRequest cms_api_struct.ChannelCodeLimitRequest
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
	SetChannelCodeLimitRequest := &pbAdmin.SetChannelCodeLimitRequest{
		OperationID:    apiRequest.OperationID,
		State:          apiRequest.State,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.SetChannelCodeLimit(context.Background(), SetChannelCodeLimitRequest)

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
		apiRequest cms_api_struct.ChannelCodeMultiDeleteRequest
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
	MultiDeleteChannelCodeRequest := &pbAdmin.MultiDeleteChannelCodeRequest{
		OperationID:    apiRequest.OperationID,
		Code:           apiRequest.ChannelCodes,
		OperatorUserId: userID,
	}
	rpcRes, rpcErr := client.MultiDeleteChannelCode(context.Background(), MultiDeleteChannelCodeRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	openIMHttp.RespHttp200(c, constant.ErrInfo{ErrCode: rpcRes.CommonResp.ErrCode, ErrMsg: rpcRes.CommonResp.ErrMsg}, nil)
	return
}

func GetCodeList(c *gin.Context) {
	var (
		apiRequest  cms_api_struct.GetInviteChannelCodeListRequest
		apiResponse cms_api_struct.GetInviteChannelCodeListResponse
	)
	if err := c.ShouldBindQuery(&apiRequest); err != nil {
		log.NewInfo("0", utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	token := c.GetHeader("token")
	if token == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is nil")
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

	// check the permissions
	if err := utils3.CheckAdminPermissions(userID); err != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "admin is banned!", userID)
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
	GetChannelCodeListRequest := pbAdmin.GetChannelCodeListRequest{
		OperationID: apiRequest.OperationID,
		Pagination: &sdk.RequestPagination{
			PageNumber: int32(apiRequest.RequestPagination.PageNumber),
			ShowNumber: int32(apiRequest.RequestPagination.ShowNumber),
		},
		Conditions: &pbAdmin.GetChannelCodeListConditions{
			Code:       apiRequest.Code,
			FriendId:   apiRequest.FriendId,
			GroupId:    apiRequest.GroupId,
			State:      apiRequest.State,
			Note:       apiRequest.Note,
			IsOfficial: apiRequest.IsOfficial,
			OrderBy:    apiRequest.OrderBy,
		},
	}
	rpcRes, rpcErr := client.GetChannelCodeList(context.Background(), &GetChannelCodeListRequest)

	if rpcErr != nil {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "rpcErr", rpcErr)
		openIMHttp.RespHttp200(c, constant.ErrRPC, nil)
		return
	}

	apiResponse.CurrentPage = int(rpcRes.CurrentNumber)
	apiResponse.ShowNumber = int(rpcRes.ShowNumber)
	apiResponse.IsOpen = int(rpcRes.IsOpen)
	apiResponse.IsLimit = int(rpcRes.Limit)
	apiResponse.Total = rpcRes.Total

	var InviteChannelCode []*cms_api_struct.InviteChannelCode
	for _, code := range rpcRes.ChannelCodes {
		var flist []cms_api_struct.Friends
		for _, v := range code.Friends {
			flist = append(flist, cms_api_struct.Friends{
				ID:   v.Id,
				Name: v.Name,
			})
		}
		var glist []cms_api_struct.Groups
		for _, v := range code.Groups {
			glist = append(glist, cms_api_struct.Groups{
				ID:   v.Id,
				Name: v.Name,
			})
		}

		InviteChannelCode = append(InviteChannelCode, &cms_api_struct.InviteChannelCode{
			ID:          utils.StringToInt(code.Id),
			Code:        code.Code,
			FriendIdArr: code.FriendIds,
			GroupIdArr:  code.GroupIds,
			State:       int32(utils.StringToInt(code.State)),
			Note:        code.Note,
			Greeting:    code.Greeting,
			SourceId:    code.SourceId,
			FriendsList: flist,
			GroupList:   glist,
		})
	}
	apiResponse.ChannelCodeList = InviteChannelCode

	openIMHttp.RespHttp200(c, constant.OK, apiResponse)
	return
}
