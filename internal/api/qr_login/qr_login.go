package qr_login

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAuth "Open_IM/pkg/proto/auth"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetQrCode(c *gin.Context) {
	var (
		apiRequest  api.GetQrCodeRequest
		apiResponse api.GetQrCodeResponse

		rpcRequest pbAuth.GetDeviceLoginQrCodeRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationId, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	connection, err := createAuthGrpcConnection(apiRequest.OperationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationId
	rpcRequest.DeviceID = apiRequest.DeviceId
	rpcRequest.Platform = apiRequest.Platform

	rpcResponse, err := connection.GetDeviceLoginQrCode(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if apiResponse.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	apiResponse.Data.QrCode = rpcResponse.QrCode
	c.JSON(http.StatusOK, apiResponse)
	return
}

func CheckState(c *gin.Context) {
	var (
		apiRequest  api.CheckStateRequest
		apiResponse api.CheckStateResponse

		rpcRequest pbAuth.CheckDeviceLoginQrCodeStateRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationId, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	connection, err := createAuthGrpcConnection(apiRequest.OperationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationId
	rpcRequest.DeviceID = apiRequest.DeviceId
	rpcRequest.QrCode = apiRequest.QrCodeId
	rpcRequest.Platform = apiRequest.PlatForm

	rpcResponse, err := connection.CheckDeviceLoginQrCodeState(context.Background(), &rpcRequest)
	if err != nil {
		errMsg := "RegisterOfficial rpc failed " + err.Error()
		log.NewError(rpcRequest.OperationID, utils.GetSelfFuncName(), errMsg, rpcRequest.String())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	_ = utils.CopyStructFields(&apiResponse.CommResp, rpcResponse.CommonResp)
	if apiResponse.CommResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	apiResponse.Data.State = rpcResponse.State

	// Confirmed login
	if rpcResponse.State == constant.QrLoginStateConfirmed {
		if err := utils2.CheckUserPermissions(rpcResponse.UserId); err != nil {
			log.NewError(apiRequest.OperationId, utils.GetSelfFuncName(), "user is banned!", rpcResponse.UserId)
			c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
			return
		}

		reply, err := connection.UserToken(context.Background(), &pbAuth.UserTokenReq{
			OperationID:    apiRequest.OperationId,
			Platform:       apiRequest.PlatForm,
			OpUserID:       rpcResponse.UserId,
			FromUserID:     rpcResponse.UserId,
			GAuthTypeToken: false,
		})

		if err != nil {
			errMsg := apiRequest.OperationId + " UserToken failed " + err.Error()
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}
		if reply.CommonResp.ErrCode != constant.OK.ErrCode {
			log.NewError(apiRequest.OperationId, "request get user token", rpcResponse.UserId, "err", "")
			c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": ""})
			return
		}

		userInfo, err := imdb.GetUserByUserID(rpcResponse.UserId)

		apiResponse.Data.UserID = rpcResponse.UserId
		apiResponse.Data.Token = reply.Token
		apiResponse.Data.ExpiredTime = reply.ExpiredTime
		apiResponse.Data.UserSign = reply.UserSign
		apiResponse.Data.Extend = api.UserTokenInfoExtend{
			IsNeedBindPhone: userInfo.PhoneNumber == "",
			IsChangePwd:     userInfo.Password == "",
			IsChangeName:    userInfo.Nickname == "",
			IsChangeFace:    userInfo.FaceURL == "",
		}
	}

	c.JSON(http.StatusOK, apiResponse)
	return
}

func PushQrCode(c *gin.Context) {
	var (
		apiRequest  api.PushQrCodeRequest
		apiResponse api.PushQrCodeResponse

		rpcRequest pbAuth.PushDeviceLoginQrCodeRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationId, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := createAuthGrpcConnection(apiRequest.OperationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationId
	rpcRequest.Platform = apiRequest.PlatForm
	rpcRequest.QrCode = apiRequest.QrCodeId
	rpcRequest.UserId = userIDInterface.(string)
	rpcResponse, err := connection.PushDeviceLoginQrCode(context.Background(), &rpcRequest)
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

	apiResponse.Data.TemporaryToken = rpcResponse.TemporaryToken
	c.JSON(http.StatusOK, apiResponse)
	return
}

func ConfirmQrCode(c *gin.Context) {
	var (
		apiRequest  api.ConfirmQrCodeRequest
		apiResponse api.ConfirmQrCodeResponse

		rpcRequest pbAuth.ConfirmDeviceLoginQrCodeRequest
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationId, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")
	connection, err := createAuthGrpcConnection(apiRequest.OperationId)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcRequest.OperationID = apiRequest.OperationId
	rpcRequest.Platform = apiRequest.PlatForm
	rpcRequest.QrCode = apiRequest.QrCodeId
	rpcRequest.TemporaryToken = apiRequest.TemporaryToken
	rpcRequest.Agree = *apiRequest.Agree
	rpcRequest.UserId = userIDInterface.(string)
	rpcResponse, err := connection.ConfirmDeviceLoginQrCode(context.Background(), &rpcRequest)
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

func createAuthGrpcConnection(OperationId string) (pbAuth.AuthClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, OperationId)
	if etcdConn == nil {
		errMsg := "etcd3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbAuth.NewAuthClient(etcdConn)
	return client, nil
}
