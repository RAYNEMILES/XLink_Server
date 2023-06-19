package oauth

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/oauth"
	"github.com/gin-gonic/gin"
	"net/http"
)

type BindingListParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
}

type EmailBindingParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,min=2,max=8,alphanum"`
}

type PhoneBindingParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	Phone       string `json:"Phone" binding:"required"`
	Code        string `json:"code" binding:"required,min=2,max=8,alphanum"`
}

type FaceBookOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	AccessToken string `json:"access_token" Binding:"required"`
}

type GoogleOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	Code        string `json:"code" Binding:"omitempty"`
	IdToken     string `json:"idToken" Binding:"omitempty"`
	State       string `json:"state" Binding:"required"`
}

type AppleOauthParamsRequest struct {
	Platform    int32  `json:"platform"`
	OperationID string `json:"operationID" binding:"required"`
	Code        string `json:"access_token" Binding:"required"`
	RedirectURI string `json:"redirect_uri" Binding:"required"`
}

func BindingFacebook(c *gin.Context) {
	params := FaceBookOauthParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	thirdUserInfo, err := oauth.ParseFaceBookAccessToken(params.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	user := imdb.GetUserByTypeAndThirdId(constant.OauthTypeFaceBook, thirdUserInfo["id"].(string))
	if user.UserId != "" {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBoundOver.ErrCode, "errMsg": constant.ErrOauthBoundOver.ErrMsg})
		return
	}

	// binding
	result := imdb.UpdateUserIdByThirdInfo(constant.OauthTypeFaceBook, thirdUserInfo["id"].(string), userID, thirdUserInfo["name"].(string))
	if result == false {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBindingFailed.ErrCode, "errMsg": constant.ErrOauthBindingFailed.ErrMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": constant.OK.ErrMsg})
	return
}

func BindingEmail(c *gin.Context) {
	params := EmailBindingParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	// check email
	_, err := imdb.GetRegisterFromEmail(params.Email)
	if err == nil {
		log.NewError(params.OperationID, "The email address has been registered", params)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "The email has been registered"})
		return
	}

	// check verify code
	verifyKey := params.Email + "_" + constant.VerificationCodeForBindEmailSuffix
	code, _ := db.DB.GetAccountCode(verifyKey)
	if code != params.Code {
		log.NewError(params.OperationID, "The verification code is incorrect", params)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code expired!"})
		return
	}

	// update email
	err = imdb.UpdateUserInfoByMap(db.User{
		UserID: userID,
	}, map[string]interface{}{"email": params.Email})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBindingFailed.ErrCode, "errMsg": constant.ErrOauthBindingFailed.ErrMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": constant.OK.ErrMsg})
	return
}

func BindingPhone(c *gin.Context) {
	params := PhoneBindingParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	// check phone
	_, err := imdb.GetRegisterFromPhone(params.Phone)
	if err == nil {
		log.NewError(params.OperationID, "The phone number has been bound", params)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.HasRegistered, "errMsg": "The phone number has been registered"})
		return
	}

	// check verify code
	verifyKey := params.Phone + "_" + constant.VerificationCodeForBindPhoneSuffix
	code, _ := db.DB.GetAccountCode(verifyKey)
	if code != params.Code && (params.Code != config.Config.Demo.SuperCode || config.Config.Environment == constant.PROD) {
		log.NewError(params.OperationID, "The verification code is incorrect", params)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.CodeInvalidOrExpired, "errMsg": "Verification code expired!"})
		return
	}

	// update email
	err = imdb.UpdateUserInfoByMap(db.User{
		UserID: userID,
	}, map[string]interface{}{"phone_number": params.Phone})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBindingFailed.ErrCode, "errMsg": constant.ErrOauthBindingFailed.ErrMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": constant.OK.ErrMsg})
	return
}

func BindingGoogle(c *gin.Context) {
	params := GoogleOauthParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	if params.State != "random" {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": "state is not random"})
		return
	}

	thirdUserInfo, err := oauth.ParseGoogleCode(int(params.Platform), params.Code, params.IdToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	user := imdb.GetUserByTypeAndThirdId(constant.OauthTypeGoogle, thirdUserInfo["sub"].(string))
	if user.UserId != "" {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBoundOver.ErrCode, "errMsg": constant.ErrOauthBoundOver.ErrMsg})
		return
	}

	// binding
	result := imdb.UpdateUserIdByThirdInfo(constant.OauthTypeGoogle, thirdUserInfo["sub"].(string), userID, thirdUserInfo["email"].(string))
	if result == false {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBindingFailed.ErrCode, "errMsg": constant.ErrOauthBindingFailed.ErrMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": constant.OK.ErrMsg})
	return
}

func BindingApple(c *gin.Context) {
	params := AppleOauthParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	thirdUserInfo, err := oauth.ParseAppleCode(int(params.Platform), params.Code, params.RedirectURI)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}

	user := imdb.GetUserByTypeAndThirdId(constant.OauthTypeApple, thirdUserInfo["id"].(string))
	if user.UserId != "" {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBoundOver.ErrCode, "errMsg": constant.ErrOauthBoundOver.ErrMsg})
		return
	}

	// binding
	result := imdb.UpdateUserIdByThirdInfo(constant.OauthTypeApple, thirdUserInfo["id"].(string), userID, thirdUserInfo["email"].(string))
	if result == false {
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrOauthBindingFailed.ErrCode, "errMsg": constant.ErrOauthBindingFailed.ErrMsg})
		return
	}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": constant.OK.ErrMsg})
	return
}

func BindingList(c *gin.Context) {
	params := BindingListParamsRequest{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.FormattingError, "errMsg": err.Error()})
		return
	}
	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	bindingList := imdb.GetBindingList(userID)
	data := make(map[string]interface{})
	for _, v := range bindingList {
		data[constant.OauthTypeIdToName(int(v.Type))] = v.Type
	}

	// email
	user, _ := imdb.GetUserByUserID(userID)
	if user.Email != "" {
		data[constant.OauthTypeEmailStr] = user.Email
	}

	c.JSON(http.StatusOK, api.GetInvitionTotalRespone{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: data})
	return
}
