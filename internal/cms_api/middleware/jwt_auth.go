package middleware

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdmin "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		operationID := utils.OperationIDGenerator()
		userID, errValidation := ValidateAdminUser(c, operationID)
		if errValidation != nil {
			log.NewError("", "GetUserIDFromToken false ", c.Request.Header.Get("token"))
			c.Abort()
			return
		}
		log.NewInfo("0", utils.GetSelfFuncName(), "userID: ", userID)
		c.Set("userID", userID)
		c.Set("isAdminUser", true)
		body, err := ioutil.ReadAll(c.Request.Body)
		go LogRequestDataInDb(*c, body, err, userID)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		c.Next()

	}
}

// ValidateAdminUser validate Header Token and get user ID from token also check is user is from mangmnet or not
//
//	Also check user si blocked or not
func ValidateAdminUser(c *gin.Context, operationID string) (string, error) {
	//get the userId from token
	var userID = ""
	token := c.GetHeader("token")
	if token == "" {
		log.NewError(operationID, utils.GetSelfFuncName(), "token is nil")
		openIMHttp.RespHttp403(c, constant.ErrTokenInvalid, nil)
		return userID, errors.New("token is nil")
	}

	_, userID, _ = token_verify.GetUserIDFromToken(token, "")
	if userID == "" {
		log.NewError(operationID, utils.GetSelfFuncName(), "token is illegal")
		openIMHttp.RespHttp403(c, constant.ErrTokenInvalid, nil)
		return userID, errors.New("token is illegal")
	}

	_, err := token_verify.VerifyManagementToken(token, userID)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "token is wrong!")
		openIMHttp.RespHttp403(c, constant.ErrTokenInvalid, nil)
		return userID, errors.New("token is wrong")
	}

	//check the permissions
	if err := utils2.CheckAdminPermissions(userID); err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "admin is banned!", userID)
		openIMHttp.RespHttp403(c, constant.ErrUserBanned, nil)
		return userID, errors.New("admin is banned")
	}

	user, err := im_mysql_model.GetRegAdminUsrByUID(userID)
	if err != nil {
		log.NewError(operationID, "Admin user have not register", userID, err.Error())
		openIMHttp.RespHttp403(c, constant.ErrUserBanned, nil)
		return userID, errors.New("no user found Kindly Register")
	}
	if user.Status != 1 {
		openIMHttp.RespHttp403(c, constant.ErrUserBanned, nil)
		return userID, errors.New("user is disabled/banned")
	}
	// clientIP := c.ClientIP()
	// InRange := checkIPRange(clientIP, user.IPRangeStart, user.IPRangeEnd)
	// if !InRange {
	// 	log.NewError(operationID, "Admin user have not register", userID)
	// 	openIMHttp.RespHttp403(c, constant.ErrUserBanned2, nil)
	// 	return userID, errors.New("user IP is out of range")
	// }

	// fullPath := c.FullPath()

	// if !db.DB.CheckAdminRolesAllowedInRedis(fullPath, strconv.Itoa(user.Role)) {
	// 	log.NewError("%v path not exsist in redis, ", fullPath, strconv.Itoa(user.Role))
	// 	openIMHttp.RespHttp403(c, constant.ErrUserBanned3, nil)
	// 	return userID, errors.New("you dont have permission")
	// }

	return userID, nil
}

func checkIPRange(clientIPs string, RangeStarts string, RangeEnds string) bool {
	clientIP := net.ParseIP(clientIPs)
	RangeStart := net.ParseIP(RangeStarts)
	RangeEnd := net.ParseIP(RangeEnds)
	if clientIP.To4() == nil || RangeStart.To4() == nil || RangeEnd.To4() == nil {
		log.NewError("Clinet IP :%v ,RangeStart : %v and RangeEnd : %v is not an IPv4 address\n", clientIP, RangeStart, RangeEnd)
		return false
	}
	if bytes.Compare(clientIP, RangeStart) >= 0 && bytes.Compare(clientIP, RangeEnd) <= 0 {
		log.NewError("%v is between %v and %v\n", clientIP, RangeStart, RangeEnd)
		return true
	}
	log.NewError("%v is NOT between %v and %v\n", clientIP, RangeStart, RangeEnd)
	return false
}

func LogRequestDataInDb(c gin.Context, req []byte, err error, userID string) {
	//get the userId from token
	// req, err := ioutil.ReadAll(c.Request.Body)
	// var req map[string]interface{}
	// err := c.BindJSON(&req)
	if err == nil {
		//if strings.Contains(c.FullPath(), "add") || strings.Contains(c.FullPath(), "create") || strings.Contains(c.FullPath(), "edit") ||
		//	strings.Contains(c.FullPath(), "remove") || strings.Contains(c.FullPath(), "alter") ||
		//	strings.Contains(c.FullPath(), "delete") {

		log.NewError("Request Binding for loging 1", req)
		var reqPb pbAdmin.OperationLogRequest
		OperationID := utils.OperationIDGenerator()
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, OperationID)
		if etcdConn == nil {
			errMsg := OperationID + "getcdv3.GetConn == nil"
			log.NewError(OperationID, errMsg)
			c.JSON(http.StatusInternalServerError, gin.H{"errCode": 500, "errMsg": errMsg})
			return
		}
		var x map[string]interface{}
		json.Unmarshal(req, &x)
		reqPb.Operator = userID
		reqPb.Action = c.FullPath()
		reqPb.Payload = string(req)
		reqPb.OperatorIP = c.ClientIP()
		var userIDInterafce = x["user_id"]
		if userIDInterafce != nil {
			reqPb.Executee = userIDInterafce.(string)
		} else {
			var deleteUserIDListIDInterface = x["deleteUserIDList"]
			if deleteUserIDListIDInterface != nil {
				reqPb.Executee = fmt.Sprint(deleteUserIDListIDInterface)
			}
			var group_members = x["group_name"]
			if group_members != nil {
				reqPb.Executee = fmt.Sprint(group_members)
			}
		}

		client := pbAdmin.NewAdminCMSClient(etcdConn)
		_, err := client.OperationLog(context.Background(), &reqPb)
		if err != nil {
			log.NewError("Request Binding faield for loging", err.Error())
			return
		}
		//}

	} else {
		log.NewError("Request Binding faield - BindJSON ", err.Error())
	}
}
