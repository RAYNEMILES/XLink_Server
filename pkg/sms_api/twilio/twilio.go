package twilio

import (
	"Open_IM/pkg/common/constant"
	http2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"net/http"
	"net/url"
	"strings"
)

type Twilio struct {
	BaseURL    string
	AccountSID string
	ApiSID     string
	ApiSecret  string
	SendNumber string
}

type TwilioSMSReq struct {
	Body string `json:"Body"`
	From string `json:"From"`
	To   string `json:"To"`
}

//msgType: 1 sms 2 email
func (t *Twilio) SendMessage(operationID string, msg string, phoneNumber string, msgType int) error {
	apiURL := t.BaseURL
	accountSID := t.AccountSID
	apiSID := t.ApiSID
	apiSecret := t.ApiSecret

	apiURL = strings.Replace(apiURL, "<api_sid>", apiSID, 1)
	apiURL = strings.Replace(apiURL, "<api_secret>", apiSecret, 1)
	apiURL = strings.Replace(apiURL, "<account_sid>", accountSID, 1)

	params := map[string]interface{}{
		"Body": msg,
		"From": t.SendNumber,
		"To":   phoneNumber,
	}

	//params := TwilioSMSReq{
	//	Body: msg,
	//	From: t.SendNumber,
	//	To:   phoneNumber,
	//}

	log.NewInfo(operationID, utils.GetSelfFuncName(), apiURL, utils.StructToJsonString(params))
	result, err := http.PostForm(apiURL, url.Values{
		"Body": {msg},
		"From": {t.SendNumber},
		"To":   {phoneNumber},
	})
	//result, err := http2.Post(apiURL, params, 60)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		return err
	}

	resultData, err2 := utils.ParseResponse(result)
	if err2 != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err2.Error())
		return err2
	}

	errorCode := resultData["error_code"]
	if errorCode != nil {
		errorMessage := resultData["error_message"]
		log.NewError(operationID, utils.GetSelfFuncName(), errorMessage)
		return http2.WrapError(constant.ErrServer)
	}

	log.NewInfo(operationID, utils.GetSelfFuncName(), utils.MapToJsonString(resultData))
	return nil
}
