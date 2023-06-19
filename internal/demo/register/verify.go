package register

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"strings"

	"github.com/gin-gonic/gin"
	"net/http"
)

type paramsCertification struct {
	Email            string `json:"email"`
	PhoneNumber      string `json:"phoneNumber"`
	VerificationCode string `json:"verificationCode"`
	OperationID      string `json:"operationID" binding:"required"`
	UsedFor          int    `json:"usedFor"`
}

func Verify(c *gin.Context) {
	params := paramsCertification{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("", "request params json parsing failed", "", "err", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	log.NewInfo("recv req: ", params)

	if params.UsedFor == 0 {
		params.UsedFor = constant.VerificationCodeForRegister
	}

	var account string
	if params.Email != "" {
		account = params.Email
		if !strings.Contains(account, "@") {
			log.NewError(params.OperationID, "The email address should contain @, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The email address should contain @, please check."})
			return
		}
		// need to check user account whether exist
		if params.UsedFor == constant.VerificationCodeForReset {
			_, err := im_mysql_model.GetRegisterFromEmail(account)
			if err != nil {
				log.NewError(params.OperationID, "The email address has not been registered", params)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "No user found! Kindly Register."})
				return
			}
		}
	} else {
		account = params.PhoneNumber
		if !strings.Contains(account, "+") {
			log.NewError(params.OperationID, "The phone number should has + at the head, please check.", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrArgs.ErrCode, "errMsg": "The phone number should has + at the head, please check."})
			return
		}
		// need to check user account whether exist
		if params.UsedFor == constant.VerificationCodeForReset {
			_, err := im_mysql_model.GetRegisterFromPhone(account)
			if err != nil {
				log.NewError(params.OperationID, "The phone number has not been registered", params)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "No user found! Kindly Register."})
				return
			}
		}
	}

	if params.VerificationCode == config.Config.Demo.SuperCode && config.Config.Environment != constant.PROD {
		log.NewInfo(params.OperationID, "Super Code Verified successfully", account)
		data := make(map[string]interface{})
		data["account"] = account
		data["verificationCode"] = params.VerificationCode
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "Verified successfully!", "data": data})
		return
	}
	log.NewInfo(params.OperationID, " params.VerificationCode != config.Config.Demo.SuperCode", params.VerificationCode, config.Config.Demo)
	log.NewInfo(params.OperationID, "begin get form redis", account)
	if params.UsedFor == 0 {
		params.UsedFor = constant.VerificationCodeForRegister
	}
	var accountKey string
	switch params.UsedFor {
	case constant.VerificationCodeForRegister:
		accountKey = account + "_" + constant.VerificationCodeForRegisterSuffix
	case constant.VerificationCodeForReset:
		accountKey = account + "_" + constant.VerificationCodeForResetSuffix
	case constant.VerificationCodeForDeleteAccount:
		accountKey = account + "_" + constant.VerificationCodeForDeleteAccountSuffix
	}

	// first whether the key exist
	if db.DB.AccountCodeIsExists(accountKey) == false {
		log.NewError(params.OperationID, "Verification code expired", accountKey)
		data := make(map[string]interface{})
		data["account"] = account
		c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code expired!", "data": data})
		return
	}

	code, err := db.DB.GetAccountCode(accountKey)
	log.NewInfo(params.OperationID, "redis phone number and verificating Code", accountKey, code, params)
	if err != nil {
		log.NewError(params.OperationID, "Verification code expired", accountKey, "err", err.Error())
		data := make(map[string]interface{})
		data["account"] = account
		c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code expired!", "data": data})
		return
	}
	if params.VerificationCode == code {
		log.Info(params.OperationID, "Verified successfully", account)
		data := make(map[string]interface{})
		data["account"] = account
		data["verificationCode"] = params.VerificationCode
		c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "Verified successfully!", "data": data})
		return
	} else {
		log.Info(params.OperationID, "Verification code error", account, params.VerificationCode)
		data := make(map[string]interface{})
		data["account"] = account
		c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Invalid verification code! Please try again", "data": data})
		return
	}
}
