package privacy

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func GetPrivacy(c *gin.Context) {
	var (
		apiRequest  api.GetPrivacyRequest
		apiResponse api.GetPrivacyResponse
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")

	client, err := createUserGRPCConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	rpcResponse, err := client.GetPrivacy(context.Background(), &pbUser.GetUserPrivacyRequest{
		OperationID: apiRequest.OperationID,
		UserId:      userIDInterface.(string),
	})
	if err != nil {
		errMsg := "rpc failed" + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	apiResponse.ErrCode = rpcResponse.CommonResp.ErrCode
	apiResponse.ErrMsg = rpcResponse.CommonResp.ErrMsg
	if rpcResponse.CommonResp.ErrCode != constant.OK.ErrCode {
		c.JSON(http.StatusOK, apiResponse)
		return
	}

	for _, v := range rpcResponse.Privacy {
		apiResponse.Data = append(apiResponse.Data, api.PrivacySetting{
			SettingKey: v.SettingKey,
			SettingVal: v.SettingValue,
		})
	}

	c.JSON(http.StatusOK, apiResponse)
	return
}

func SetPrivacy(c *gin.Context) {
	var (
		apiRequest  api.SetPrivacyRequest
		apiResponse api.SetPrivacyResponse
	)

	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := "bind json failed " + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	userIDInterface, _ := c.Get("userID")

	client, err := createUserGRPCConnection(apiRequest.OperationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	privacy := make([]*pbUser.Privacy, 0)
	for _, v := range apiRequest.Data {
		privacy = append(privacy, &pbUser.Privacy{
			SettingKey:   v.SettingKey,
			SettingValue: v.SettingVal,
		})
	}

	rpcResponse, err := client.SetPrivacy(context.Background(), &pbUser.SetUserPrivacyRequest{
		OperationID: apiRequest.OperationID,
		UserId:      userIDInterface.(string),
		Privacy:     privacy,
	})
	if err != nil {
		errMsg := "rpc failed" + err.Error()
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), errMsg)
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	apiResponse.ErrCode = rpcResponse.CommonResp.ErrCode
	apiResponse.ErrMsg = rpcResponse.CommonResp.ErrMsg

	c.JSON(http.StatusOK, apiResponse)
	return
}

func createUserGRPCConnection(OperationID string) (pbUser.UserClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, OperationID)
	if etcdConn == nil {
		errMsg := "getcdv3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbUser.NewUserClient(etcdConn)
	return client, nil
}
