package interest

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/group"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type SetInterestRequest struct {
	OperationID  string  `json:"operationID"`
	InterestList []int64 `json:"interestList" binding:"required"`
}

func SetInterest(c *gin.Context) {
	apiRequest := SetInterestRequest{}
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	token := c.GetHeader("token")
	_, userID, _ := token_verify.GetUserIDFromToken(token, apiRequest.OperationID)
	if userID == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is illegal")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token is illegal"})
		return
	}

	imdb.SetUserInterestType(userID, apiRequest.InterestList)

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "success"})
	return
}

func GetInterestGroup(c *gin.Context) {
	token := c.GetHeader("token")
	OperationID := utils.OperationIDGenerator()
	_, userID, _ := token_verify.GetUserIDFromToken(token, OperationID)
	if userID == "" {
		log.NewError(OperationID, utils.GetSelfFuncName(), "token is illegal")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "token is illegal"})
		return
	}

	//check user status
	if err := utils2.CheckUserPermissions(OperationID); err != nil {
		log.NewError(OperationID, utils.GetSelfFuncName(), "user is banned!", userID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, OperationID)
	if etcdConn == nil {
		errMsg := OperationID + "getcdv3.GetConn == nil"
		log.NewError(OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewGroupClient(etcdConn)
	RpcResp, err := client.GetInterestGroupListByUserId(context.Background(), &rpc.GetInterestGroupListByUserIdRequest{
		OperationID: OperationID,
		FromUserID:  userID,
	})
	if err != nil {
		log.NewError(OperationID, "GetGroupMemberList failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	data := make([]map[string]string, 0)
	for _, i := range RpcResp.GroupList {
		data = append(data, map[string]string{
			"groupID":      i.GroupID,
			"groupName":    i.GroupName,
			"FaceURL":      i.FaceURL,
			"Notification": i.Notification,
			"Introduction": i.Introduction,
			"MemberNum":    i.Ex,
		})
	}

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "success", "data": data})
	return
}

func RemoveGroup(c *gin.Context) {
	params := api.RemoveInterestGroupRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	userID := ""
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	imdb.AddInterestGroupExcludeListByUserId(userID, params.Group)

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "success"})
	return
}
