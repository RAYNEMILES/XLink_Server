package register

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pb "Open_IM/pkg/proto/admin_cms"
	commonPb "Open_IM/pkg/proto/sdk_ws"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type InterestListRequest struct {
	OperationID  string `json:"operationID" binding:"required"`
	LanguageType string `json:"languageType" binding:"required"`
}

func InterestList(c *gin.Context) {
	params := InterestListRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	data := make([]map[string]interface{}, 0)
	if params.LanguageType == "zh" {
		params.LanguageType = "cn"
	}

	// cache
	cache, err := db.DB.GetInterestListByLanguage(params.LanguageType)
	if err == nil {
		json.Unmarshal([]byte(cache), &data)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "data": data})
		return
	}

	operationID := params.OperationID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, operationID)
	if etcdConn == nil {
		errMsg := operationID + "getcdv3.GetConn == nil"
		log.NewError(operationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	require := &pb.GetInterestsReq{}
	require.Status = "1"
	require.Pagination = &commonPb.RequestPagination{}
	require.Pagination.PageNumber = int32(1)
	require.Pagination.ShowNumber = int32(-1)
	client := pb.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetInterests(context.Background(), require)

	if err != nil {
		errMsg := operationID + "client.GetInterests err:" + err.Error()
		log.NewError(operationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	for _, interest := range respPb.Interests {
		name := ""
		for _, interestName := range interest.Name {
			if interestName.LanguageType == params.LanguageType {
				name = interestName.Name
				break
			}
		}

		if name == "" {
			continue
		}

		data = append(data, map[string]interface{}{
			"id":   interest.Id,
			"name": name,
		})
	}

	// redis
	marshal, _ := json.Marshal(data)
	db.DB.SetInterestListByLanguage(params.LanguageType, string(marshal))

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": data})
	return
}
