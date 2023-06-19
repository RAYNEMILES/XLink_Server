package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/chat"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/rand"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AllowGuestLogin(c *gin.Context) {
	data := map[string]interface{}{}
	data["allow_guest_login"] = im_mysql_model.GetAllowGuestLogin()
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": data})
}

func Login(c *gin.Context) {
	params := ParamsLogin{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	var user *db.User
	var err error
	if params.Email != "" {
		account = params.Email
		user, err = im_mysql_model.GetRegisterFromEmailOrUserIdOrPhone(account)
		if err != nil {
			log.NewError(params.OperationID, "user have not register", params.Password, account, err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}
	} else if params.UserId != "" {
		account = params.UserId
		user, err = im_mysql_model.GetRegisterFromEmailOrUserIdOrPhone(account)
		if err != nil {
			log.NewError(params.OperationID, "user have not register", params.Password, account, err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}
	} else {
		account = params.PhoneNumber
		user, err = im_mysql_model.GetRegisterFromEmailOrUserIdOrPhone(account)
		if err != nil {
			log.NewError(params.OperationID, "user have not register", params.Password, account, err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register..."})
			return
		}
	}

	// user is deleted
	if user.DeleteTime != 0 {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register...."})
		return
	}

	newPasswordFirst := params.Password + user.Salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	password := fmt.Sprintf("%x", has)

	if user.Password != password {
		log.NewError(params.OperationID, "password  err", newPasswordFirst, params.Password, user.Password, password, user.Salt, account)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.PasswordErr, "errMsg": "password error!"})
		return
	}
	userID := user.UserID

	//check user status
	if err := utils2.CheckUserPermissions(userID); err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "user is banned!", userID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	var domain string
	var domainPre = "http://"
	if config.Config.Environment != constant.DEV {
		domain = c.Request.Host
		domainPre = "https://"
	}
	if domain == "" {
		domain = fmt.Sprintf("%s:%d", utils.ServerIP, config.Config.Api.GinPort[0])
	} else {
		domain = fmt.Sprintf("%s/api", domain)
	}

	url := fmt.Sprintf("%s%s/auth/user_token", domainPre, domain)
	openIMGetUserToken := api.UserTokenReq{}
	openIMGetUserToken.OperationID = params.OperationID
	openIMGetUserToken.Platform = params.Platform
	openIMGetUserToken.Secret = password
	openIMGetUserToken.UserID = userID
	openIMGetUserToken.GAuthTypeToken = false
	openIMGetUserToken.LoginIp = c.ClientIP()
	openIMGetUserTokenResp := api.UserTokenResp{}
	bMsg, err := http2.Post(url, openIMGetUserToken, constant.ApiTimeOutSeconds)
	if err != nil {
		log.NewError(params.OperationID, "request openIM get user token error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMGetUserTokenResp)
	if err != nil || openIMGetUserTokenResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request get user token", account, "err", "")
		c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": ""})
		return
	}
	// if user.SuperUserStatus == 1 {
	// 	openIMGetUserTokenResp.UserToken.IsSuperUser = true
	// }

	//store groupid list to redis
	groupIDList, err := im_mysql_model.GetJoinedGroupIDListByUserID(userID)
	if err == nil {
		var tmpGroupIDList []interface{}
		for _, s := range groupIDList {
			if s != "" {
				tmpGroupIDList = append(tmpGroupIDList, s)
			}
		}
		if len(tmpGroupIDList) > 0 {
			_ = db.DB.SaveGroupIDListForUser(userID, tmpGroupIDList)
		}
	}

	openIMGetUserTokenResp.UserToken.Extend = api.UserTokenInfoExtend{
		IsChangeName:    user.Nickname == "",
		IsChangePwd:     user.Password == "",
		IsNeedBindPhone: user.PhoneNumber == "",
		IsChangeFace:    user.FaceURL == "",
	}

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMGetUserTokenResp.UserToken})

}

// GenerateInAppPin used for generating in-app OTP code for login
func GenerateInAppPin(c *gin.Context) {
	params := GenerateInAppPinRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	var user *db.User
	var err error
	if params.UserName != "" {
		account = params.UserName
		user, err = im_mysql_model.GetRegisterFromEmailOrUserIdOrPhone(account)
		if err != nil {
			log.NewError("user have not register", params.UserName, account, err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
			return
		}
	}

	generatedCode := rangeIn(10000, 99999)
	db.DB.SaveInAppLoginPin(user.UserID, generatedCode)

	msg := strings.Replace(constant.PinGeneratedMessage, "@nickname@", user.Nickname, -1)
	msg = strings.Replace(msg, "@pincode@", strconv.Itoa(generatedCode), -1)
	//message := sdk_struct.MsgStruct{}
	//createTextMsg(&message, constant.SysMsgType, constant.Text)
	//message.Content = msg

	paramsSendMsg := paramsUserSendMsg{}
	paramsSendMsg.SendID = "xlink"
	paramsSendMsg.SenderNickName = "X-Link"
	paramsSendMsg.SenderFaceURL = ""
	paramsSendMsg.Data.ClientMsgID = utils.GetMsgID("xlink")
	paramsSendMsg.Data.SessionType = constant.NotificationChatType
	paramsSendMsg.Data.MsgFrom = constant.SysMsgType
	paramsSendMsg.Data.ContentType = constant.Text
	paramsSendMsg.Data.RecvID = user.UserID
	paramsSendMsg.Data.Content = []byte(msg)
	paramsSendMsg.Data.CreateTime = time.Now().Unix()

	pbData := newUserSendMsgReq("", &paramsSendMsg)
	log.Info(paramsSendMsg.OperationID, "", "api SendMsg call start..., [data: %s]", pbData.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, paramsSendMsg.OperationID)
	if etcdConn == nil {
		errMsg := paramsSendMsg.OperationID + "getcdv3.GetConn == nil"
		log.NewError(paramsSendMsg.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbChat.NewChatClient(etcdConn)

	log.Info("", "", "api SendMsg call, api call rpc...")

	reply, err := client.SendMsg(context.Background(), pbData)
	if err != nil {
		log.NewError(paramsSendMsg.OperationID, "SendMsg rpc failed, ", paramsSendMsg, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": "SendMsg rpc failed, " + err.Error()})
		return
	}
	log.Info(paramsSendMsg.OperationID, "", "api SendMsg call end..., [data: %s] [reply: %s]", pbData.String(), reply.String())

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": "code generated and sent"})

}

// LoginWithInAppOTP login with in-app OTP
func LoginWithInAppOTP(c *gin.Context) {
	params := ParamsLogin{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	var user *db.User
	var err error

	account = params.UserId
	user, err = im_mysql_model.GetRegisterFromEmailOrUserIdOrPhone(account)
	if err != nil {
		log.NewError(params.OperationID, "user have not register", params.Password, account, err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "No user found! Kindly Register"})
		return
	}

	pinCode, err := db.DB.GetInAppLoginPin(user.UserID)
	if err != nil || pinCode == "" {
		log.NewError(params.OperationID, "pin code is not correct", params.Password, account, err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.PasswordErr, "errMsg": "pinCode error!"})
		return
	}

	if params.Password != pinCode {
		log.NewError(params.OperationID, "pin code is not correct", params.Password)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.PasswordErr, "errMsg": "pinCode error!"})
		return
	}
	userID := user.UserID

	//check user status
	if err := utils2.CheckUserPermissions(userID); err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "user is banned!", userID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	var domain string
	var domainPre = "http://"
	if config.Config.Environment != constant.DEV {
		domain = c.Request.Host
		domainPre = "https://"
	}
	if domain == "" {
		domain = fmt.Sprintf("%s:%d", utils.ServerIP, config.Config.Api.GinPort[0])
	} else {
		domain = fmt.Sprintf("%s/api", domain)
	}

	url := fmt.Sprintf("%s%s/auth/user_token", domainPre, domain)
	openIMGetUserToken := api.UserTokenReq{}
	openIMGetUserToken.OperationID = params.OperationID
	openIMGetUserToken.Platform = params.Platform
	openIMGetUserToken.Secret = user.Password
	openIMGetUserToken.UserID = userID
	openIMGetUserToken.GAuthTypeToken = false
	openIMGetUserToken.LoginIp = c.ClientIP()
	openIMGetUserTokenResp := api.UserTokenResp{}
	bMsg, err := http2.Post(url, openIMGetUserToken, constant.ApiTimeOutSeconds)
	if err != nil {
		log.NewError(params.OperationID, "request openIM get user token error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMGetUserTokenResp)
	if err != nil || openIMGetUserTokenResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request get user token", account, "err", "")
		c.JSON(http.StatusOK, gin.H{"errCode": constant.GetIMTokenErr, "errMsg": ""})
		return
	}
	// if user.SuperUserStatus == 1 {
	// 	openIMGetUserTokenResp.UserToken.IsSuperUser = true
	// }

	//store groupid list to redis
	groupIDList, err := im_mysql_model.GetJoinedGroupIDListByUserID(userID)
	if err == nil {
		var tmpGroupIDList []interface{}
		for _, s := range groupIDList {
			if s != "" {
				tmpGroupIDList = append(tmpGroupIDList, s)
			}
		}
		if len(tmpGroupIDList) > 0 {
			_ = db.DB.SaveGroupIDListForUser(userID, tmpGroupIDList)
		}
	}

	openIMGetUserTokenResp.UserToken.Extend = api.UserTokenInfoExtend{
		IsChangeName:    user.Nickname == "",
		IsChangePwd:     user.Password == "",
		IsNeedBindPhone: user.PhoneNumber == "",
		IsChangeFace:    user.FaceURL == "",
	}

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMGetUserTokenResp.UserToken})

}

func Test(c *gin.Context) {
	buf := make([]byte, 102400)
	n, _ := c.Request.Body.Read(buf)
	body := string(buf[0:n])

	log.NewError("", utils.GetSelfFuncName(), "body", body)
}

func rangeIn(low, hi int) int {
	return low + rand.Intn(hi-low)
}

/*
/
/	Data Models
/
*/
type ParamsLogin struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	Password    string `json:"password"`
	Platform    int32  `json:"platform"`
	UserId      string `json:"userId"`
	OperationID string `json:"operationID" binding:"required"`
}

type GenerateInAppPinRequest struct {
	UserName string `json:"userName"`
}

type paramsUserSendMsg struct {
	SenderPlatformID int32  `json:"senderPlatformID" binding:"required"`
	SendID           string `json:"sendID" binding:"required"`
	SenderNickName   string `json:"senderNickName"`
	SenderFaceURL    string `json:"senderFaceUrl"`
	OperationID      string `json:"operationID" binding:"required"`
	Data             struct {
		SessionType int32                        `json:"sessionType" binding:"required"`
		MsgFrom     int32                        `json:"msgFrom" binding:"required"`
		ContentType int32                        `json:"contentType" binding:"required"`
		RecvID      string                       `json:"recvID" `
		GroupID     string                       `json:"groupID" `
		ForceList   []string                     `json:"forceList"`
		Content     []byte                       `json:"content" binding:"required"`
		Options     map[string]bool              `json:"options" `
		ClientMsgID string                       `json:"clientMsgID" binding:"required"`
		CreateTime  int64                        `json:"createTime" binding:"required"`
		OffLineInfo *open_im_sdk.OfflinePushInfo `json:"offlineInfo" `
	}
}

func newUserSendMsgReq(token string, params *paramsUserSendMsg) *pbChat.SendMsgReq {
	pbData := pbChat.SendMsgReq{
		Token:       token,
		OperationID: params.OperationID,
		MsgData: &open_im_sdk.MsgData{
			SendID:           params.SendID,
			RecvID:           params.Data.RecvID,
			GroupID:          params.Data.GroupID,
			ClientMsgID:      params.Data.ClientMsgID,
			SenderPlatformID: params.SenderPlatformID,
			SenderNickname:   params.SenderNickName,
			SenderFaceURL:    params.SenderFaceURL,
			SessionType:      params.Data.SessionType,
			MsgFrom:          params.Data.MsgFrom,
			ContentType:      params.Data.ContentType,
			Content:          params.Data.Content,
			CreateTime:       params.Data.CreateTime,
			Options:          params.Data.Options,
			OfflinePushInfo:  params.Data.OffLineInfo,
		},
	}
	return &pbData
}
