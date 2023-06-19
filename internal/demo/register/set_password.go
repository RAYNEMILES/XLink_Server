package register

import (
	constant2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
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
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ParamsSetPassword struct {
	Email            string `json:"email"`
	Nickname         string `json:"nickname"`
	PhoneNumber      string `json:"phoneNumber"`
	Password         string `json:"password" binding:"required"`
	VerificationCode string `json:"verificationCode"`
	Platform         int32  `json:"platform" binding:"required,min=1,max=7"`
	Ex               string `json:"ex"`
	FaceURL          string `json:"faceURL"`
	OperationID      string `json:"operationID" binding:"required"`
}

func GetRegisterType(c *gin.Context) {
	data := make(map[string]interface{})
	data["type"] = config.Config.PhoneRegisterType

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": data})
	return
}

func InviteCodeSwitch(c *gin.Context) {
	sw := 0
	if (config.Config.Invite.IsOpen == 1 && imdb.GetInviteCodeIsOpen()) || (config.Config.Channel.IsOpen == 1 && imdb.GetChannelCodeIsOpen()) {
		sw = 1
	}

	data := make(map[string]interface{})
	data["switch"] = sw

	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": data})
	return
}

func SetPassword(c *gin.Context) {
	params := ParamsSetPassword{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "bind json failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	// allow register type
	if !utils.IsContainInt(constant2.RegisterTypeByMobile, config.Config.PhoneRegisterType) {
		// c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.NotAllowRegisterType, "errMsg": "This method of registration is not allowed"})
		// return
	}

	var account string
	if params.Email != "" {
		account = params.Email
		if !strings.Contains(account, "@") {
			log.NewError(params.OperationID, "The email address should contain @, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The email address should contain @, please check."})
			return
		}
	} else {
		account = params.PhoneNumber
		if !strings.Contains(account, "+") {
			log.NewError(params.OperationID, "The phone number should has + at the head, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The phone number should has + at the head, please check."})
			return
		}
	}

	if params.VerificationCode != config.Config.Demo.SuperCode || config.Config.Environment == constant.PROD {
		accountKey := account + "_" + constant.VerificationCodeForRegisterSuffix
		v, err := db.DB.GetAccountCode(accountKey)
		if err != nil || v != params.VerificationCode {
			log.NewError(params.OperationID, "password Verification code error", account, params.VerificationCode)
			data := make(map[string]interface{})
			data["PhoneNumber"] = account
			c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code error!", "data": data})
			return
		}
	}
	// userID := utils.Base64Encode(account)

	userID := utils.Md5(params.OperationID + strconv.FormatInt(time.Now().UnixNano(), 10))
	bi := big.NewInt(0)
	bi.SetString(userID[0:8], 16)
	userID = bi.String()

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
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": "register failed"})
		return
	}

	// merge register table to users table
	// log.Info(params.OperationID, "begin store mysql", account, params.Password)
	// err = im_mysql_model.SetPassword(account, params.Password, params.Ex, userID)
	// if err != nil {
	//	log.NewError(params.OperationID, "set phone number password error", account, "err", err.Error())
	//	c.JSON(http.StatusOK, gin.H{"errCode": constant.RegisterFailed, "errMsg": err.Error()})
	//	return
	// }
	// log.Info(params.OperationID, "end setPassword", account, params.Password)

	// demo onboarding
	// onboardingProcess(params.OperationID, userID, params.Nickname)
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "", "data": openIMRegisterResp.UserToken})
	return
}

func GetVerificationCode(c *gin.Context) {
	params := api.GetVerificationCodeReq{}

	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	// create verification code
	id, b64s, _ := utils2.CaptMake()
	resp := api.GetVerificationCodeResp{
		CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""},
		Data: api.GetVerificationCodeData{
			InviteCode: id,
			Image:      b64s,
		},
	}

	c.JSON(http.StatusOK, resp)
}

func PushInviteCode(c *gin.Context) {
	params := api.InviteCodeReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	// lowercase
	params.Mobile = strings.ToLower(params.Mobile)
	params.Os = strings.ToLower(params.Os)
	params.Language = strings.ToLower(params.Language)

	params.Ip = c.ClientIP()
	params.CreateTime = time.Now().Unix()

	// Pure storage
	req := &db.InviteCodeLog{}
	utils.CopyStructFields(req, &params)
	imdb.PushCode(req)

	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": ""})
	return
}

func GetInviteCode(c *gin.Context) {
	params := api.InviteCodeReq{}
	if err := c.BindJSON(&params); err != nil {
		errMsg := " BindJSON failed " + err.Error()
		log.NewError("0", errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	// TEST
	if params.Test != "" {
		c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{InviteCode: params.Test}})
		return
	}

	// lowercase
	params.Mobile = strings.ToLower(params.Mobile)
	params.Os = strings.ToLower(params.Os)
	params.Language = strings.ToLower(params.Language)

	params.Ip = c.ClientIP()
	params.CreateTime = time.Now().Unix()

	list, err := imdb.GetCodeByOr(&params)
	if err != nil {
		if imdb.CodeIsExpired(params.Code) == true {
			c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{InviteCode: params.Code}})
			return
		}
		log.NewError(params.OperationID, utils.GetSelfFuncName(), params, err.Error())
		c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{
			InviteCode: "",
		}})
		return
	}
	if list == nil || len(list) == 0 {
		// verify param.code,maybe valid
		if imdb.CodeIsExpired(params.Code) == true {
			c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{InviteCode: params.Code}})
			return
		}
		c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{InviteCode: ""}})
		return
	}

	// way 1 : matched
	// c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{ChannelCode: list[0].Code}})

	// way 2 ：weighting
	weighting := []int{10, 10, 10, 10, 10, 10, 10, 10, 10}
	//scores := make(map[string]int)
	code := list[0].Code
	num := 0
	for _, item := range list {
		// score
		codeNum := 0
		if item.Code == params.Code {
			codeNum += weighting[0]
		}
		if item.Timezone == params.Timezone {
			codeNum += weighting[1]
		}
		if item.Mobile == params.Mobile {
			codeNum += weighting[2]
		}
		if item.Os == params.Os {
			codeNum += weighting[3]
		}
		if item.Version == params.Version {
			codeNum += weighting[4]
		}
		if item.Webkit == params.Webkit {
			codeNum += weighting[5]
		}
		if item.ScreenWidth == params.ScreenWidth {
			codeNum += weighting[6]
		}
		if item.Language == params.Language {
			codeNum += weighting[7]
		}
		if item.Ip == params.Ip {
			codeNum += weighting[8]
		}

		if codeNum > num {
			code = item.Code
			num = codeNum
		}
	}

	// code is expired
	if imdb.CodeIsExpired(code) == false && imdb.ChannelCodeIsExpired(code) == false {
		c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{InviteCode: ""}})
		return
	}

	// code := utils2.FindMaxInMap(scores)
	c.JSON(http.StatusOK, api.PushCodeResp{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: api.PushCodeData{InviteCode: code}})
	return
}
