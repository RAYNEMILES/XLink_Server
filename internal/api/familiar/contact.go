package familiar

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/front_api_struct"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
)

func GetFamiliarList(c *gin.Context) {
	params := api.GetFamiliarListRequest{}
	response := api.GetFamiliarListResponse{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	contactList := imdb.GetContactByUserId(userID)
	if len(contactList) == 0 {
		c.JSON(http.StatusOK, response)
		return
	}

	var phoneList []string
	for _, contact := range contactList {
		phoneList = append(phoneList, contact.Phone)
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, params.OperationID)
	if etcdConn == nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "etcdConn is nil"})
		return
	}
	client := pb.NewUserClient(etcdConn)
	result, err := client.GetYouKnowUsersByContactList(context.Background(), &pb.GetYouKnowUsersByContactListRequest{
		PhoneNumber: phoneList,
		UserId:      userID,
		OperationID: params.OperationID,
	})

	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		resp := front_api_struct.FrontApiResp{
			ErrCode: constant.ErrRPC.ErrCode,
			ErrMsg:  constant.ErrRPC.ErrMsg,
			Data:    nil,
		}
		c.JSON(http.StatusOK, resp)
		return
	}

	var usersInfo []api.GetFamiliarListData
	for _, i := range result.User {
		usersInfo = append(usersInfo, api.GetFamiliarListData{
			UserID:   i.UserId,
			UserName: i.Nickname,
			Avatar:   i.ProfilePhoto,
			Gender:   i.Gender,
			Contact:  1,
			Friend:   0,
			Meet:     0,
		})
	}
	response.Data = usersInfo

	c.JSON(http.StatusOK, response)
	return
}

func SyncContact(c *gin.Context) {
	params := api.SyncContactRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	// data cleaning
	var phoneNumber []string
	reg, _ := regexp.Compile("[^+\\d]")
	for _, s := range params.ContactList {
		if s != "" {
			s = reg.ReplaceAllString(strings.TrimLeft(s, "0"), "")
			if s == "" {
				continue
			}
			phoneNumber = append(phoneNumber, s)
			continue
		}
	}

	imdb.UpdateContactByUserId(userID, utils.RemoveDuplicatesAndEmpty(phoneNumber))

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, params.OperationID)
	if etcdConn == nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "etcdConn is nil"})
		return
	}

	client := pb.NewUserClient(etcdConn)
	result, _ := client.GetYouKnowUsersByContactList(context.Background(), &pb.GetYouKnowUsersByContactListRequest{
		PhoneNumber: phoneNumber,
		UserId:      userID,
		OperationID: params.OperationID,
	})

	data := make([]map[string]interface{}, 0)
	if len(result.User) > 0 {
		for _, v := range result.User {
			data = append(data, map[string]interface{}{
				"userId":   v.UserId,
				"userName": v.Nickname,
				"avatar":   v.ProfilePhoto,
				"contact":  1,
				"friend":   0,
				"meet":     0,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "success", "data": data})
	return
}

func RemoveUser(c *gin.Context) {
	params := api.RemoveFamiliarUserRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	imdb.AddContactExcludeByUserId(userID, params.UserId)

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "success"})
	return
}
