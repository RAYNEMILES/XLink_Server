package apiChat

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/chat"
	sdk_ws "Open_IM/pkg/proto/sdk_ws"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type paramsUserNewestSeq struct {
	ReqIdentifier int    `json:"reqIdentifier" binding:"required"`
	SendID        string `json:"sendID" binding:"required"`
	OperationID   string `json:"operationID" binding:"required"`
	MsgIncr       int    `json:"msgIncr" binding:"required"`
}

//type ParamsGetMaxSeq struct {
//	OpUserID    string `json:"opUserID" binding:"required"`
//	GroupID     string `json:"groupID"`
//	OperationID string `json:"operationID" binding:"required"`
//}

func GetSeq(c *gin.Context) {
	params := paramsUserNewestSeq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		params.SendID = userIDInter.(string)
	}

	pbData := sdk_ws.GetMaxAndMinSeqReq{}
	pbData.UserID = params.SendID
	pbData.OperationID = params.OperationID
	grpcConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, pbData.OperationID)
	if grpcConn == nil {
		errMsg := pbData.OperationID + " getcdv3.GetConn == nil"
		log.NewError(pbData.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	msgClient := pbChat.NewChatClient(grpcConn)
	reply, err := msgClient.GetMaxAndMinSeq(context.Background(), &pbData)
	if err != nil {
		log.NewError(params.OperationID, "UserGetSeq rpc failed, ", params, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "UserGetSeq rpc failed, " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errCode":       reply.ErrCode,
		"errMsg":        reply.ErrMsg,
		"msgIncr":       params.MsgIncr,
		"reqIdentifier": params.ReqIdentifier,
		"data": gin.H{
			"maxSeq": reply.MaxSeq,
			"minSeq": reply.MinSeq,
		},
	})

}

func GetMaxSeq(c *gin.Context) {
	params := api.ParamsGetMaxSeq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		params.OpUserID = userIDInter.(string)
	}

	pbData := sdk_ws.GetMaxAndMinSeqReq{}
	pbData.UserID = params.OpUserID
	pbData.OperationID = params.OperationID
	if params.GroupID != "" {
		pbData.GroupIDList = append(pbData.GroupIDList, params.GroupID)
	}

	grpcConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, pbData.OperationID)
	if grpcConn == nil {
		errMsg := pbData.OperationID + " getcdv3.GetConn == nil"
		log.NewError(pbData.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	msgClient := pbChat.NewChatClient(grpcConn)
	reply, err := msgClient.GetMaxAndMinSeq(context.Background(), &pbData)
	if err != nil {
		log.NewError(params.OperationID, "UserGetSeq rpc failed, ", params, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "UserGetSeq rpc failed, " + err.Error()})
		return
	}
	maxSeq := reply.MaxSeq
	dataMap := gin.H{
		"maxSeq": maxSeq,
	}
	if reply.GroupMaxAndMinSeq != nil {
		if groupSeqMax, ok := reply.GroupMaxAndMinSeq[params.GroupID]; ok {
			dataMap[params.GroupID] = groupSeqMax.MaxSeq
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": reply.ErrCode,
		"errMsg":  reply.ErrMsg,
		"data":    dataMap,
	})

}

func GetGroupMinSeq(c *gin.Context) {
	params := api.ParamsGetMaxSeq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	userIDInter, existed := c.Get("userID")
	if existed {
		params.OpUserID = userIDInter.(string)
	}

	pbData := sdk_ws.GetMaxAndMinSeqReq{}
	pbData.UserID = params.OpUserID
	pbData.OperationID = params.OperationID
	if params.GroupID != "" {
		pbData.GroupIDList = append(pbData.GroupIDList, params.GroupID)
	}

	grpcConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, pbData.OperationID)
	if grpcConn == nil {
		errMsg := pbData.OperationID + " getcdv3.GetConn == nil"
		log.NewError(pbData.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	msgClient := pbChat.NewChatClient(grpcConn)
	reply, err := msgClient.GetMaxAndMinSeq(context.Background(), &pbData)
	if err != nil {
		log.NewError(params.OperationID, "UserGetSeq rpc failed, ", params, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "UserGetSeq rpc failed, " + err.Error()})
		return
	}
	maxSeq := reply.MaxSeq
	dataMap := gin.H{
		"maxSeq": maxSeq,
	}
	if reply.GroupMaxAndMinSeq != nil {
		if groupSeqMax, ok := reply.GroupMaxAndMinSeq[params.GroupID]; ok {
			dataMap[params.GroupID] = groupSeqMax.MinSeq
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"errCode": reply.ErrCode,
		"errMsg":  reply.ErrMsg,
		"data":    dataMap,
	})

}
