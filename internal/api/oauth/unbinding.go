package oauth

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UnbindingOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" Binding:"required"`
	ThirdType   string `json:"third_type" Binding:"required,oneof=apple google facebook email"`
}

func Unbinding(c *gin.Context) {
	params := UnbindingOauthParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	oauthType := constant.OauthTypeNameToId(params.ThirdType)
	if oauthType == constant.OauthTypeEmail {
		// update user info
		err := imdb.UpdateUserInfoByMap(db.User{
			UserID: userID,
		}, map[string]interface{}{"email": ""})
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthUnbindingFail.ErrCode, "errMsg": constant.ErrOauthUnbindingFail.ErrMsg})
			return
		}
		return
	} else {
		result := imdb.UnbindingOauthByUserIdAndThirdId(oauthType, userID)
		if result {
			c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "success"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthUnbindingFail.ErrCode, "errMsg": constant.ErrOauthUnbindingFail.ErrMsg})
	return
}
