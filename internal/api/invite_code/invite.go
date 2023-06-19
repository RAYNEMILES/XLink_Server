package invite_code

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/front_api_struct"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpc "Open_IM/pkg/proto/user"
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetInviteCodeLink(c *gin.Context) {
	apiRequest := api.GetInviteCodeLinkRequest{}
	apiResponse := front_api_struct.FrontApiResp{}
	if err := c.BindJSON(&apiRequest); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	if userID == "" {
		log.NewError(apiRequest.OperationID, utils.GetSelfFuncName(), "token is illegal")
		apiResponse = front_api_struct.FrontApiResp{
			ErrCode: constant.ErrArgs.ErrCode,
			ErrMsg:  constant.ErrArgs.ErrMsg,
			Data:    nil,
		}
		c.JSON(int(apiResponse.ErrCode), apiResponse)
		return
	}

	// Function open or not
	if config.Config.Invite.IsOpen != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invite code function is not open"})
		return
	}

	// etcd
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, apiRequest.OperationID)
	if etcdConn == nil {
		errMsg := apiRequest.OperationID + " getcdv3.GetConn == nil"
		log.NewError(apiRequest.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := rpc.NewUserClient(etcdConn)
	link, err := client.GetIviteLink(context.Background(), &rpc.GetInviteLinkRequest{
		OperationID: apiRequest.OperationID,
		UserId:      userID,
	})
	if err != nil {
		log.NewError(apiRequest.OperationID, "call invite link  rpc server failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call invite link users rpc server failed"})
		return
	}

	data := make(map[string]interface{})
	data["link"] = link.InviteLink
	c.JSON(http.StatusOK, api.GetInviteCodeLinkResponse{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: data})
	return
}

func GetTotalInvitation(c *gin.Context) {
	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	if userID == "" {
		log.NewError("", utils.GetSelfFuncName(), "token is illegal")
		apiResponse := front_api_struct.FrontApiResp{
			ErrCode: constant.ErrArgs.ErrCode,
			ErrMsg:  constant.ErrArgs.ErrMsg,
			Data:    nil,
		}
		c.JSON(int(apiResponse.ErrCode), apiResponse)
		return
	}

	total := imdb.GetCountByOwnerUserId(userID)
	data := make(map[string]interface{})
	data["total"] = total
	c.JSON(http.StatusOK, api.GetInvitionTotalRespone{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: data})
	return
}
