package register

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	huawei2 "Open_IM/pkg/sms_api/huawei"
	"Open_IM/pkg/sms_api/twilio"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gin-gonic/gin"
	"gopkg.in/gomail.v2"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type paramsVerificationCode struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	OperationID string `json:"operationID" binding:"required"`
	UsedFor     int    `json:"usedFor"`
	Language    string `json:"language"`
}

func SendVerificationCode(c *gin.Context) {
	params := paramsVerificationCode{}
	if err := c.BindJSON(&params); err != nil {
		log.NewError("", "BindJSON failed", "err:", err.Error(), "phoneNumber", params.PhoneNumber, "email", params.Email)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
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
	var accountKey string
	if params.UsedFor == 0 {
		params.UsedFor = constant.VerificationCodeForRegister
	}

	var language string
	if params.Language == "en" {
		language = "en"
	} else if params.Language == "cn" {
		language = "cn"
	} else {
		language = "en"
	}

	var messageContent string
	var msgType = constant.SendMsgRegister

	switch params.UsedFor {
	case constant.VerificationCodeForBindEmail:
		_, err := im_mysql_model.GetRegisterFromEmail(account)
		if err == nil {
			log.NewError(params.OperationID, "The email address has been registered", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "The email address has been registered"})
			return
		}
		accountKey = account + "_" + constant.VerificationCodeForBindEmailSuffix

		if language == "cn" {
			messageContent = config.Config.Sms.SmsBindEmailCn
		} else {
			messageContent = config.Config.Sms.SmsBindEmailEn
		}
	case constant.VerificationCodeForRegister:
		//check the account
		if params.Email != "" {
			_, err := im_mysql_model.GetRegisterFromEmail(account)
			if err == nil {
				log.NewError(params.OperationID, "The email address has been registered", params)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "The email address has been registered"})
				return
			}
		} else {
			_, err := im_mysql_model.GetRegisterFromPhone(account)
			if err == nil {
				log.NewError(params.OperationID, "The phone number has been registered", params)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "The phone number has been registered"})
				return
			}

		}

		accountKey = account + "_" + constant.VerificationCodeForRegisterSuffix

		if language == "cn" {
			messageContent = config.Config.Sms.SmsRegisterCN
		} else {
			messageContent = config.Config.Sms.SmsRegisterEN
		}
	case constant.VerificationCodeForBindPhone:
		//check the account
		_, err := im_mysql_model.GetRegisterFromPhone(account)
		if err == nil {
			log.NewError(params.OperationID, "The phone number has been registered", params)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "The phone number has been bound"})
			return
		}

		accountKey = account + "_" + constant.VerificationCodeForBindPhoneSuffix

		if language == "cn" {
			messageContent = config.Config.Sms.SmsRegisterCN
		} else {
			messageContent = config.Config.Sms.SmsRegisterEN
		}
	case constant.VerificationCodeForReset:
		//check the account
		if params.Email != "" {
			_, err := im_mysql_model.GetRegisterFromEmail(account)
			if err != nil {
				log.NewError(params.OperationID, "The email address has not been registered", params)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "No user found! Kindly Register."})
				return
			}
		} else {
			_, err := im_mysql_model.GetRegisterFromPhone(account)
			if err != nil {
				log.NewError(params.OperationID, "The phone number has not been registered", params)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "No user found! Kindly Register."})
				return
			}

		}
		msgType = constant.SendMsgResetPassword
		accountKey = account + "_" + constant.VerificationCodeForResetSuffix

		if language == "cn" {
			messageContent = config.Config.Sms.SmsResetpasswordCN
		} else {
			messageContent = config.Config.Sms.SmsResetpasswordEN
		}
	case constant.VerificationCodeForDeleteAccount:
		if params.PhoneNumber == "" {
			log.NewError(params.OperationID, "You must provide a phone number.")
			c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "You must provide a phone number."})
			return
		} else {
			deleteUser, err := im_mysql_model.GetUserByPhoneNumber(params.PhoneNumber)
			if err != nil || deleteUser == nil {
				log.NewError(params.OperationID, "Phone number isn't available")
				c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Phone number isn't available"})
				return
			}
		}

		msgType = constant.SendMsgDeleteAccount
		accountKey = account + "_" + constant.VerificationCodeForDeleteAccountSuffix
	}

	//check if the verification code was sent in 1 minutes
	ok, err := db.DB.JudgeAccountEXISTS(accountKey + "_repeat")
	if ok || err != nil {
		log.NewError(params.OperationID, "The verification code cannot be sent repeatedly within 1 minutes", params)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.RepeatSendCode, "errMsg": "The verification code cannot be sent repeatedly within 1 minutes"})
		return
	}

	rand.Seed(time.Now().UnixNano())
	code := 100000 + rand.Intn(900000)
	log.NewInfo(params.OperationID, params.UsedFor, "begin store redis", accountKey, code)
	err = db.DB.SetAccountCode(accountKey, code, config.Config.Demo.ExpireTTL)
	err = db.DB.SetAccountCode(accountKey+"_repeat", code, config.Config.Demo.CodeTTL)
	if err != nil {
		fmt.Println("failed", err.Error())
		log.NewError(params.OperationID, "set redis error", accountKey, "err", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Failed to send verification code"})
		return
	}

	//set the message content
	messageContent = strings.Replace(messageContent, "[]", fmt.Sprintf("[%d]", code), 1)

	log.NewDebug("", config.Config.Demo)
	if params.Email != "" {
		m := gomail.NewMessage()
		m.SetHeader(`From`, config.Config.Demo.Mail.SenderMail)
		m.SetHeader(`To`, []string{account}...)
		m.SetHeader(`Subject`, config.Config.Demo.Mail.Title)
		m.SetBody(`text/html`, messageContent)
		if err := gomail.NewDialer(config.Config.Demo.Mail.SmtpAddr, config.Config.Demo.Mail.SmtpPort, config.Config.Demo.Mail.SenderMail, config.Config.Demo.Mail.SenderAuthorizationCode).DialAndSend(m); err != nil {
			log.Error(params.OperationID, "send mail error", account, err.Error())
			c.JSON(http.StatusOK, gin.H{"errCode": constant.MailSendCodeErr, "errMsg": ""})
			return
		}
	} else {
		if config.Config.Sms.Api == "alismsverify" {
			client, err := CreateClient(tea.String(config.Config.Demo.AliSMSVerify.AccessKeyID), tea.String(config.Config.Demo.AliSMSVerify.AccessKeySecret))
			if err != nil {
				log.NewError(params.OperationID, "create sendSms client err", "err", err.Error())
				c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Failed to send verification code"})
				return
			}

			sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
				PhoneNumbers:  tea.String(account),
				SignName:      tea.String(config.Config.Demo.AliSMSVerify.SignName),
				TemplateCode:  tea.String(config.Config.Demo.AliSMSVerify.VerificationCodeTemplateCode),
				TemplateParam: tea.String(messageContent),
			}

			response, err := client.SendSms(sendSmsRequest)
			if err != nil {
				log.NewError(params.OperationID, "sendSms error", account, "err", err.Error())
				c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Failed to send verification code"})
				return
			}
			if *response.Body.Code != "OK" {
				log.NewError(params.OperationID, "alibabacloud sendSms error", account, "err", response.Body.Code, response.Body.Message)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Failed to send verification code"})
				return
			}
		} else if config.Config.Sms.Api == "twilio" {
			twilio := twilio.Twilio{
				BaseURL:    config.Config.Sms.Twilio.ApiUrl,
				AccountSID: config.Config.Sms.Twilio.AccountSID,
				ApiSID:     config.Config.Sms.Twilio.ApiSID,
				ApiSecret:  config.Config.Sms.Twilio.ApiSecret,
				SendNumber: config.Config.Sms.Twilio.SendNumber,
			}

			log.NewInfo(params.OperationID, "send twilio sms", account, messageContent)
			err = twilio.SendMessage(params.OperationID, messageContent, account, 1)
			if err != nil {
				log.NewError(params.OperationID, "send twilio sms error", account, "err", err.Error())
				c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Failed to send verification code"})
				return
			}
		} else if config.Config.Sms.Api == "huawei" {
			huawei := huawei2.HuaWei{
				ApiAddress: config.Config.Sms.Huawei.Api,
				Signature:  config.Config.Sms.Huawei.Signature,
			}

			err := huawei.SendMessage(params.OperationID, account, strconv.Itoa(code), msgType, language)
			if err != nil {
				log.NewError(params.OperationID, "send huawei sms error", account, "err", err)
				c.JSON(http.StatusOK, gin.H{"errCode": constant.SmsSendCodeErr, "errMsg": "Failed to send verification code..."})
				return
			}
		}

	}

	data := make(map[string]interface{})
	data["account"] = account
	c.JSON(http.StatusOK, gin.H{"errCode": constant.NoError, "errMsg": "Verification code has been sent!", "data": data})
}

func CreateClient(accessKeyId *string, accessKeySecret *string) (result *dysmsapi20170525.Client, err error) {
	c := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: accessKeyId,
		// 您的AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}

	// 访问的域名
	c.Endpoint = tea.String("dysmsapi.aliyuncs.com")
	result = &dysmsapi20170525.Client{}
	result, err = dysmsapi20170525.NewClient(c)
	return result, err
}
