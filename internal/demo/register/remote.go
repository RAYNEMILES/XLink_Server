package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/auth"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type assignRemoteTokenParamsRequest struct {
	OperationID string `json:"operationID"`
	Token       string `json:"token" binding:"required"`
}

type consumeRemoteTokenParamsRequest struct {
	Token       string `json:"token" binding:"required"`
	Secret      string `json:"secret" binding:"required"`
	Platform    int32  `json:"platform" binding:"required,eq=8"` // remote login is only available for official
	OperationID string `json:"operationID"`
}

type createRemoteTokenResponseData struct {
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

type createRemoteTokenResponse struct {
	api.CommResp
	Data createRemoteTokenResponseData `json:"data"`
}

var remoteToken = db.RemoteTokenModel{}

func CreateRemoteToken(c *gin.Context) {
	token, secret, err := remoteToken.Create()
	if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 500, ErrMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, createRemoteTokenResponse{
		CommResp: api.CommResp{
			ErrCode: 0,
			ErrMsg:  "",
		},
		Data: createRemoteTokenResponseData{
			Token:  token,
			Secret: secret,
		},
	})
}

func AssignRemoteToken(c *gin.Context) {
	var params assignRemoteTokenParamsRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.FormattingError, ErrMsg: err.Error()})
		return
	}

	ok, userID, errInfo := token_verify.GetUserIDFromToken(c.Request.Header.Get("token"), params.OperationID)
	if !ok {
		errMsg := params.OperationID + " " + "GetUserIDFromToken failed " + errInfo + " token:" + c.Request.Header.Get("token")
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	if err := remoteToken.Assign(params.Token, userID); err == db.RemoteTokenNotFoundError {
		c.JSON(http.StatusBadRequest, api.CommResp{
			ErrCode: constant.RemoteTokenExpiredErrorCode,
			ErrMsg:  err.Error(),
		})
		return
	} else if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{
			ErrCode: 500,
			ErrMsg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, api.CommResp{
		ErrCode: 0,
		ErrMsg:  "",
	})
}

func ConsumeRemoteToken(c *gin.Context) {
	var params consumeRemoteTokenParamsRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.FormattingError, ErrMsg: err.Error()})
		return
	}

	userID, err := remoteToken.Consume(params.Token, params.Secret)
	if err == db.RemoteTokenNotFoundError {
		c.JSON(http.StatusBadRequest, api.CommResp{
			ErrCode: constant.RemoteTokenExpiredErrorCode,
			ErrMsg:  err.Error(),
		})
		return
	} else if err == db.RemoteTokenNotNotAssignedError {
		c.JSON(http.StatusBadRequest, api.CommResp{
			ErrCode: constant.RemoteTokenNotAssignedErrorCode,
			ErrMsg:  err.Error(),
		})
		return
	} else if err != nil {
		c.JSON(http.StatusBadRequest, api.CommResp{
			ErrCode: 500,
			ErrMsg:  err.Error(),
		})
		return
	}

	req := &rpc.UserTokenReq{Platform: params.Platform, FromUserID: userID, OperationID: params.OperationID, GAuthTypeToken: false}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + " getcdv3.GetConn == nil"
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	client := rpc.NewAuthClient(etcdConn)

	reply, err := client.UserToken(context.Background(), req)
	if err != nil {
		errMsg := req.OperationID + " UserToken failed " + err.Error() + req.String()
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: errMsg})
		return
	}

	c.JSON(http.StatusOK, api.UserTokenResp{
		CommResp: api.CommResp{
			ErrCode: reply.CommonResp.ErrCode, ErrMsg: reply.CommonResp.ErrMsg,
		},
		UserToken: api.UserTokenInfo{
			UserID:      req.FromUserID,
			Token:       reply.Token,
			ExpiredTime: reply.ExpiredTime,
		},
	})
}
