package callback

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbShortVideo "Open_IM/pkg/proto/short_video"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func Vod(c *gin.Context) {
	var (
		callbackRequest         api.VodCallbackRequest
		rpcFileUpload           pbShortVideo.FileUploadCallBackRequest
		rpcFileDeleted          pbShortVideo.FileDeletedCallBackRequest
		rpcProcedureStateChange pbShortVideo.ProcedureStateChangeCallBackRequest
	)

	operationID := utils.OperationIDGenerator()
	err := c.ShouldBindJSON(&callbackRequest)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	log.NewInfo(operationID, "callbackRequest", callbackRequest)

	client, err := createShortVideoGRPConnection(operationID)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
		return
	}

	switch callbackRequest.EventType {
	case "FileDeleted":
		err := utils.CopyStructFields(&rpcFileDeleted, &callbackRequest)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
			return
		}

		rpcFileDeleted.OperationID = operationID
		result, err := client.FileDeletedCallBack(context.Background(), &rpcFileDeleted)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusOK, api.CommResp{ErrCode: result.CommonResp.ErrCode, ErrMsg: result.CommonResp.ErrMsg})
			return
		}
		c.JSON(http.StatusOK, api.CommResp{ErrCode: result.CommonResp.ErrCode, ErrMsg: result.CommonResp.ErrMsg})
		return
	case "NewFileUpload":
		err := utils.CopyStructFields(&rpcFileUpload, &callbackRequest)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
			return
		}

		rpcFileUpload.OperationID = operationID
		result, err := client.FileUploadCallBack(context.Background(), &rpcFileUpload)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusOK, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, api.CommResp{ErrCode: result.CommonResp.ErrCode, ErrMsg: result.CommonResp.ErrMsg})
		return
	case "ProcedureStateChanged":
		err := utils.CopyStructFields(&rpcProcedureStateChange, &callbackRequest)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
			return
		}

		rpcFileUpload.OperationID = operationID
		result, err := client.ProcedureStateChangeCallBack(context.Background(), &rpcProcedureStateChange)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			c.JSON(http.StatusOK, api.CommResp{ErrCode: 400, ErrMsg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, api.CommResp{ErrCode: result.CommonResp.ErrCode, ErrMsg: result.CommonResp.ErrMsg})
		return
	default:
		log.NewError(operationID, utils.GetSelfFuncName(), "Event type not support yet!")
		c.JSON(http.StatusBadRequest, api.CommResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "Event type not support yet!"})
		return
	}
}

func createShortVideoGRPConnection(OperationID string) (pbShortVideo.ShortVideoClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImShortVideoName, OperationID)
	if etcdConn == nil {
		errMsg := "etcd3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbShortVideo.NewShortVideoClient(etcdConn)
	return client, nil
}
