package middleware

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/network"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/utils"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		operationID := utils.OperationIDGenerator()
		userID, errValidation := ValidateUser(c, operationID)
		if errValidation != nil {
			log.NewError("", "GetUserIDFromToken false ", c.Request.Header.Get("token"))
			c.Abort()
			return
		}
		log.NewInfo("0", utils.GetSelfFuncName(), "userID: ", userID)
		c.Set("userID", userID)
		c.Next()
	}
}

// ValidateUser validate Header Token and get user ID from token also check is user
//
//	Also check user si blocked or not
func ValidateUser(c *gin.Context, operationID string) (string, error) {
	//get the userId from token
	token := c.GetHeader("token")
	if token == "" {
		log.NewError(operationID, utils.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp403(c, constant.ErrTokenInvalid, nil)
		return "", errors.New("token is nil")
	}

	_, userID, officialID, _ := token_verify.GetUserIDFromTokenV2(token, "")
	if userID == "" {
		log.NewError(operationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp403(c, constant.ErrTokenInvalid, nil)
		return userID, errors.New("token is illegal")
	}
	if err := utils2.CheckUserPermissions(userID); err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "user is banned!", userID)
		openIMHttp.RespHttp403(c, constant.ErrTokenInvalid, nil)
		return userID, errors.New("token is illegal")
	}
	if officialID != 0 {
		go updateOfficialAccountLastActivityIPAndRegion(officialID, c.ClientIP())
	}
	return userID, nil
}

func updateOfficialAccountLastActivityIPAndRegion(officialID int64, clientIP string) {
	official, err := imdb.GetOfficialByOfficialID(officialID)
	if err == nil && official.ProcessStatus == 1 && official.LastLoginIp != clientIP {
		var country, city string
		requestURl := config.Config.LocationIpServerAddressPreFix + clientIP + config.Config.LocationIpServerAddressPostFix
		responseBytes, err := network.DoGetRequest(requestURl)
		if err != nil {
			log.NewError("", "updateOfficialAccountLastActivityIPAndRegion ", officialID, clientIP, err.Error())
			return
		}
		if len(responseBytes) > 0 {
			jsonInterface := make(map[string]interface{})
			err := json.Unmarshal(responseBytes, &jsonInterface)
			if err != nil {
				log.NewError("", "updateOfficialAccountLastActivityIPAndRegion ", officialID, clientIP, err.Error())
				return
			}
			if val, ok := jsonInterface["country"]; ok {
				country = val.(string)
			}
			if val, ok := jsonInterface["city"]; ok {
				city = val.(string)
			}
			err = imdb.UpdateOfficialAccountIPAndLocationInfo(officialID, clientIP, country, city)
			if err != nil {
				log.NewError("", "updateOfficialAccountLastActivityIPAndRegion ", officialID, clientIP, err.Error())
				return
			}

		}

	}
}
