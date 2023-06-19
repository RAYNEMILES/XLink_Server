package huawei

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

// WSSE_HEADER_FORMAT 无需修改,用于格式化鉴权头域,给"X-WSSE"参数赋值
const WSSE_HEADER_FORMAT = "UsernameToken Username=\"%s\",PasswordDigest=\"%s\",Nonce=\"%s\",Created=\"%s\""

// AUTH_HEADER_VALUE 无需修改,用于格式化鉴权头域,给"Authorization"参数赋值
const AUTH_HEADER_VALUE = "WSSE realm=\"SDP\",profile=\"UsernameToken\",type=\"Appkey\""

const CODE_SUCCESS = "000000"

type HuaWei struct {
	ApiAddress string
	Signature  string
}

// SendMessage
// msgType : 1登陆2找回密码
func (h *HuaWei) SendMessage(operationID string, phoneNumber string, code string, msgType int, language string) interface{} {

	appKey := config.Config.Sms.Huawei.En.AppKey       // APP_Key
	appSecret := config.Config.Sms.Huawei.En.AppSecret // APP_Secret
	sender := config.Config.Sms.Huawei.En.Sender       // 国内短信签名通道号或国际/港澳台短信通道号
	templateId := ""

	// nation, message language, message type then get message template code.
	var templateMap = map[string]map[string]map[int]string {
		"cn": {
			"en": {
				constant.SendMsgRegister: config.Config.Sms.Huawei.Cn.SmsRegisterEn,
				constant.SendMsgResetPassword: config.Config.Sms.Huawei.Cn.SmsResetPasswordEn,
				constant.SendMsgDeleteAccount: config.Config.Sms.Huawei.Cn.SmsDeleteAccountEn,
			},
			"cn": {
				constant.SendMsgRegister: config.Config.Sms.Huawei.Cn.SmsRegisterCn,
				constant.SendMsgResetPassword: config.Config.Sms.Huawei.Cn.SmsResetPasswordCn,
				constant.SendMsgDeleteAccount: config.Config.Sms.Huawei.Cn.SmsDeleteAccountCn,
			},
		},
		"en": {
			"en": {
				constant.SendMsgRegister: config.Config.Sms.Huawei.En.SmsRegisterEn,
				constant.SendMsgResetPassword: config.Config.Sms.Huawei.En.SmsResetPasswordEn,
				constant.SendMsgDeleteAccount: config.Config.Sms.Huawei.En.SmsDeleteAccountEn,
			},
			"cn": {
				constant.SendMsgRegister: config.Config.Sms.Huawei.En.SmsRegisterCn,
				constant.SendMsgResetPassword: config.Config.Sms.Huawei.En.SmsResetPasswordCn,
				constant.SendMsgDeleteAccount: config.Config.Sms.Huawei.En.SmsDeleteAccountCn,
			},
		},
	}
	// chinese
	//if string([]byte(phoneNumber)[:3]) == "+86" {
	//	appKey = config.Config.Sms.Huawei.Cn.AppKey
	//	appSecret = config.Config.Sms.Huawei.Cn.AppSecret
	//	sender = config.Config.Sms.Huawei.Cn.Sender
	//	if msgType == constant.SendMsgRegister {
	//		if language == "en" {
	//			templateId = config.Config.Sms.Huawei.Cn.SmsRegisterEn
	//		} else if language == "cn" {
	//			templateId = config.Config.Sms.Huawei.Cn.SmsRegisterCn
	//		}
	//	} else if msgType == constant.SendMsgResetPassword {
	//		if language == "en" {
	//			templateId = config.Config.Sms.Huawei.Cn.SmsResetPasswordEn
	//		} else if language == "cn" {
	//			templateId = config.Config.Sms.Huawei.Cn.SmsResetPasswordCn
	//		}
	//	}
	//} else {
	//	// english
	//}
	if string([]byte(phoneNumber)[:3]) == "+86" {
		// chinese
		appKey = config.Config.Sms.Huawei.Cn.AppKey
		appSecret = config.Config.Sms.Huawei.Cn.AppSecret
		sender = config.Config.Sms.Huawei.Cn.Sender
		templateId = templateMap["cn"][language][msgType]
	} else {
		// english
		templateId = templateMap["en"][language][msgType]
	}

	apiAddress := h.ApiAddress // APP接入地址(在控制台"应用管理"页面获取)+接口访问URI

	// 条件必填,国内短信关注,当templateId指定的模板类型为通用模板时生效且必填,必须是已审核通过的,与模板类型一致的签名名称
	// 国际/港澳台短信不用关注该参数
	signature := h.Signature

	// 必填,全局号码格式(包含国家码),示例:+8615123456789,多个号码之间用英文逗号分隔
	receiver := phoneNumber // 短信接收人号码

	// 选填,短信状态报告接收地址,推荐使用域名,为空或者不填表示不接收状态报告
	statusCallBack := ""

	/*
	 * 选填,使用无变量模板时请赋空值 string templateParas = "";
	 * 单变量模板示例:模板内容为"您的验证码是${1}"时,templateParas可填写为"[\"369751\"]"
	 * 双变量模板示例:模板内容为"您有${1}件快递请到${2}领取"时,templateParas可填写为"[\"3\",\"人民公园正门\"]"
	 * 模板中的每个变量都必须赋值，且取值不能为空
	 * 查看更多模板和变量规范:产品介绍>模板和变量规范
	 */
	templateParas := "[\"" + code + "\",\"" + strconv.Itoa(int(time.Duration(config.Config.Demo.ExpireTTL)*time.Second/time.Minute)) + "\"]" // 模板变量，此处以单变量验证码短信为例，请客户自行生成6位验证码，并定义为字符串类型，以杜绝首位0丢失的问题（例如：002569变成了2569）。

	body := buildRequestBody(sender, receiver, templateId, templateParas, statusCallBack, signature)
	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"
	headers["Authorization"] = AUTH_HEADER_VALUE
	headers["X-WSSE"] = buildWsseHeader(appKey, appSecret)

	log.NewInfo(operationID, utils.GetSelfFuncName(), apiAddress, utils.StructToJsonString(body))

	resp, err := post(apiAddress, []byte(body), headers)
	log.NewInfo(operationID, utils.GetSelfFuncName(), resp, err)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), apiAddress, utils.StructToJsonString(body), err)
		return err
	}

	m := make(map[string]interface{})
	err2 := json.Unmarshal([]byte(resp), &m)
	if err2 != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), apiAddress, utils.StructToJsonString(body), resp, err2)
		return err
	}

	errorCode := m["code"]
	if errorCode != CODE_SUCCESS {
		log.NewError(operationID, utils.GetSelfFuncName(), apiAddress, utils.StructToJsonString(body), resp)
		return m["description"]
	}

	return nil
}

/**
 * sender,receiver,templateId不能为空
 */
func buildRequestBody(sender, receiver, templateId, templateParas, statusCallBack, signature string) string {
	param := "from=" + url.QueryEscape(sender) + "&to=" + url.QueryEscape(receiver) + "&templateId=" + url.QueryEscape(templateId)
	if templateParas != "" {
		param += "&templateParas=" + url.QueryEscape(templateParas)
	}
	if statusCallBack != "" {
		param += "&statusCallback=" + url.QueryEscape(statusCallBack)
	}
	if signature != "" {
		param += "&signature=" + url.QueryEscape(signature)
	}
	return param
}

func post(url string, param []byte, headers map[string]string) (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(param))
	if err != nil {
		return "", err
	}
	for key, header := range headers {
		req.Header.Set(key, header)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func buildWsseHeader(appKey, appSecret string) string {
	var cTime = time.Now().Format("2006-01-02T15:04:05Z")
	var nonce = uuid.NewV4().String()
	nonce = strings.ReplaceAll(nonce, "-", "")

	h := sha256.New()
	h.Write([]byte(nonce + cTime + appSecret))
	passwordDigestBase64Str := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf(WSSE_HEADER_FORMAT, appKey, passwordDigestBase64Str, nonce, cTime)
}
