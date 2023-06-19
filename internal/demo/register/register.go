package register

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Register(c *gin.Context) {
	params := api.RegisterReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)

		regex := regexp.MustCompile(`for '(.*?)'`)
		value := regex.FindStringSubmatch(err.Error())[1]

		log.NewError("0", value)

		switch value {
		case "UserId":
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrRegisterParamUidErr.ErrCode, "errMsg": constant.ErrRegisterParamUidErr.ErrMsg})
			return
		case "PhoneNumber":
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrRegisterParamPhoneErr.ErrCode, "errMsg": constant.ErrRegisterParamPhoneErr.ErrMsg})
			return
		case "Email":
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrRegisterParamEmailErr.ErrCode, "errMsg": constant.ErrRegisterParamEmailErr.ErrMsg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": err.Error()})
		return
	}

	if len(params.Nickname) > 36 {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrNickNameLength.ErrCode, "errMsg": constant.ErrNickNameLength.ErrMsg})
		return
	}

	userID := utils.Md5(params.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
	bi := big.NewInt(0)
	bi.SetString(userID[0:8], 16)
	userID = bi.String()

	var account string
	var needVerifyCode bool
	if params.Email != "" {
		account = params.Email
		if !strings.Contains(account, "@") {
			log.NewError(params.OperationID, "The email address should contain @, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The email address should contain @, please check."})
			return
		}
		userCheck, _ := imdb.GetRegisterFromEmail(account)
		if userCheck != nil && userCheck.UserID != "" {
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrRegisterParamEmailExist.ErrCode, "errMsg": constant.ErrRegisterParamEmailExist.ErrMsg})
			return
		}
		needVerifyCode = true
	} else if params.PhoneNumber != "" {
		account = params.PhoneNumber
		if !strings.Contains(account, "+") {
			log.NewError(params.OperationID, "The phone number should has + at the head, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The phone number should has + at the head, please check."})
			return
		}
		userCheck, _ := imdb.GetRegisterFromPhone(account)
		if userCheck != nil && userCheck.UserID != "" {
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrRegisterParamPhoneExist.ErrCode, "errMsg": constant.ErrRegisterParamPhoneExist.ErrMsg})
			return
		}
		needVerifyCode = true
	} else {
		account = params.UserId
		userID = params.UserId
		userCheck, _ := imdb.GetHistoryRegisterFromUserId(userID)
		if userCheck != nil && userCheck.UserID != "" {
			if userCheck.DeleteTime > 0 {
				c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccountNotAvailable.ErrCode, "errMsg": constant.ErrAccountNotAvailable.ErrMsg})
				return
			}

			log.NewError(params.OperationID, utils.GetSelfFuncName(), "user exist!！", userCheck.UserID)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrRegisterParamUidExist.ErrCode, "errMsg": constant.ErrRegisterParamUidExist.ErrMsg})
			return
		}
		needVerifyCode = false
	}

	if needVerifyCode && (params.VerificationCode != config.Config.Demo.SuperCode || config.Config.Environment == constant.PROD) {
		switch params.Platform {
		case constant.WebPlatformID, constant.MiniWebPlatformID:
			if utils2.CaptVerify(params.VerificationId, strings.ToLower(params.VerificationCode)) == false {
				c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrCaptchaError.ErrCode, "errMsg": constant.ErrCaptchaError.ErrMsg})
				return
			}
			break
		default:
			accountKey := account + "_" + constant.VerificationCodeForRegisterSuffix
			v, err := db.DB.GetAccountCode(accountKey)
			if err != nil || v != params.VerificationCode {
				log.NewError(params.OperationID, "password Verification code error", account, params.VerificationCode)
				data := make(map[string]interface{})
				data["Email"] = params.Email
				data["PhoneNumber"] = params.PhoneNumber
				c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code error!", "data": data})
				return
			}
		}
	}

	if params.Nickname == "" {
		params.Nickname = userID
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
	url := fmt.Sprintf("%s%s/auth/user_register", domainPre, domain)
	openIMRegisterReq := api.UserRegisterReq{}
	openIMRegisterReq.OperationID = params.OperationID
	openIMRegisterReq.Platform = params.Platform
	openIMRegisterReq.UserID = userID
	openIMRegisterReq.Nickname = params.Nickname
	openIMRegisterReq.Secret = params.Password
	openIMRegisterReq.FaceURL = params.FaceURL
	openIMRegisterReq.PhoneNumber = params.PhoneNumber
	openIMRegisterReq.Email = params.Email
	openIMRegisterReq.Ex = params.Ex
	openIMRegisterReq.SourceCode = params.InviteCode
	openIMRegisterReq.CreateIp = c.ClientIP()
	openIMRegisterReq.UpdateIp = c.ClientIP()
	openIMRegisterResp := api.UserRegisterResp{}
	bMsg, err := http2.Post(url, openIMRegisterReq, 2)
	if err != nil {
		log.NewError(params.OperationID, "request openIM register error", account, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
		return
	}
	err = json.Unmarshal(bMsg, &openIMRegisterResp)
	if err != nil || openIMRegisterResp.ErrCode != 0 {
		log.NewError(params.OperationID, "request openIM register error", account, "err", "resp: ", openIMRegisterResp.ErrCode)
		if err != nil {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed: " + openIMRegisterResp.ErrMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMRegisterResp.UserToken})
	return
}
