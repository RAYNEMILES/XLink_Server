package register

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/utils"
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type resetPasswordRequest struct {
	VerificationCode string `json:"verificationCode" binding:"required"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phoneNumber"`
	NewPassword      string `json:"newPassword" binding:"required"`
	OperationID      string `json:"operationID"`
}

func ResetPassword(c *gin.Context) {
	var (
		req resetPasswordRequest
	)
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var account string
	var user *db.User
	var err error
	if req.Email != "" {
		account = req.Email
		user, err = im_mysql_model.GetRegisterFromEmail(account)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "get register error", err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "user not register!"})
			return
		}

	} else {
		account = req.PhoneNumber
		user, err = im_mysql_model.GetRegisterFromPhone(account)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "get register error", err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.NotRegistered, "errMsg": "user not register!"})
			return
		}
	}
	if req.VerificationCode != config.Config.Demo.SuperCode || config.Config.Environment == constant.PROD {
		accountKey := account + "_" + constant.VerificationCodeForResetSuffix
		v, err := db.DB.GetAccountCode(accountKey)
		if err != nil || v != req.VerificationCode {
			log.NewError(req.OperationID, "password Verification code error", account, req.VerificationCode, v)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code error!"})
			return
		}
	}

	//check user status
	if err := utils2.CheckUserPermissions(user.UserID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "user is banned!", user.UserID)
		c.JSON(http.StatusOK, gin.H{"errCode": err.ErrCode, "errMsg": err.ErrMsg})
		return
	}

	newPasswordFirst := req.NewPassword + user.Salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	newPassword := fmt.Sprintf("%x", has)

	if err := im_mysql_model.ResetPassword(user.UserID, newPassword); err != nil {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ResetPasswordFailed, "errMsg": "reset password failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "reset password success"})
}
