package base_info

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// UserID               string   `protobuf:"bytes,1,opt,name=UserID" json:"UserID,omitempty"`
//	Nickname             string   `protobuf:"bytes,2,opt,name=Nickname" json:"Nickname,omitempty"`
//	FaceUrl              string   `protobuf:"bytes,3,opt,name=FaceUrl" json:"FaceUrl,omitempty"`
//	Gender               int32    `protobuf:"varint,4,opt,name=Gender" json:"Gender,omitempty"`
//	PhoneNumber          string   `protobuf:"bytes,5,opt,name=PhoneNumber" json:"PhoneNumber,omitempty"`
//	Birth                string   `protobuf:"bytes,6,opt,name=Birth" json:"Birth,omitempty"`
//	Email                string   `protobuf:"bytes,7,opt,name=Email" json:"Email,omitempty"`
//	Ex                   string   `protobuf:"bytes,8,opt,name=Ex" json:"Ex,omitempty"`

type UserRegisterReq struct {
	Secret   string `json:"secret" binding:"omitempty,max=32"`
	Platform int32  `json:"platform" binding:"required,min=1,max=7"`
	ApiUserInfo
	OperationID string `json:"operationID" binding:"required"`
	SourceCode  string `json:"source_code" binding:"omitempty,max=32"`
	Uuid        string `json:"uuid" binding:"omitempty,max=32"`
	CreateIp    string `json:"create_ip" binding:"omitempty,max=32"`
	UpdateIp    string `json:"update_ip" binding:"omitempty,max=32"`
}

type RegisterReq struct {
	Platform    int32  `json:"platform" binding:"required,min=1,max=7"`
	OperationID string `json:"operationID" binding:"required"`

	Email       string `json:"email" binding:"omitempty,email"`
	PhoneNumber string `json:"phoneNumber" binding:"omitempty,max=32"`
	UserId      string `json:"userId" binding:"omitempty,alphanum,max=32,min=6"`

	InviteCode     string `json:"invite_code" binding:"omitempty"`
	VerificationId string `json:"verification_id" binding:"omitempty,alphanum,min=1,max=50"`

	VerificationCode string `json:"verification_code" binding:"omitempty,alphanum"`
	Nickname         string `json:"nickname"`
	Password         string `json:"password" binding:"required,max=32"`
	Ex               string `json:"ex"`
	FaceURL          string `json:"faceURL"`
}

type UserTokenInfo struct {
	UserID string `json:"userID"`
	Token  string `json:"token"`
	// IsSuperUser bool                `json:"isSuperUser"`
	ExpiredTime int64               `json:"expiredTime"`
	Extend      UserTokenInfoExtend `json:"extend"`
	UserSign    string              `json:"user_sign"`

	// GAuthEnabled       bool   `json:"gAuthEnabled"`
	// GAuthSetupRequired bool   `json:"gAuthSetupRequired"`
	// GAuthSetupProvUri  string `json:"gAuthSetupProvUri"`
}

type UserTokenInfoExtend struct {
	IsNeedBindPhone bool `json:"isNeedBindPhone"`
	IsChangePwd     bool `json:"isChangePwd"`
	IsChangeName    bool `json:"isChangeName"`
	IsChangeFace    bool `json:"isChangeFace"`
}

type UserRegisterResp struct {
	CommResp
	UserToken UserTokenInfo `json:"data"`
}

type UserTokenReq struct {
	Secret         string `json:"secret" binding:"required,max=32"`
	Platform       int32  `json:"platform" binding:"required,min=1,max=8"`
	UserID         string `json:"userID" binding:"required,min=1,max=64"`
	OperationID    string `json:"operationID" binding:"required"`
	LoginIp        string `json:"login_ip" binding:"omitempty,max=32"`
	GAuthTypeToken bool   `json:"gAuthTypeToken"`
}

type UserTokenResp struct {
	CommResp
	UserToken UserTokenInfo `json:"data"`
}

type ForceLogoutReq struct {
	Platform    int32  `json:"platform" binding:"required,min=1,max=8"`
	FromUserID  string `json:"fromUserID" binding:"required,min=1,max=64"`
	OperationID string `json:"operationID" binding:"required"`
}

type GetVerificationCodeReq struct {
	Platform    int32  `json:"platform" binding:"required,min=1,max=8"`
	OperationID string `json:"operationID" binding:"required"`
}

type ForceLogoutResp struct {
	CommResp
}

type ParseTokenReq struct {
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"userID" binding:"required"`
	PlatformID  int32  `json:"platformID" binding:"required"`
}

// type ParseTokenResp struct {
//	CommResp
//	ExpireTime int64 `json:"expireTime" binding:"required"`
// }

type ExpireTime struct {
	ExpireTimeSeconds uint32 `json:"expireTimeSeconds" `
}

type ParseTokenResp struct {
	CommResp
	Data       map[string]interface{} `json:"data"`
	ExpireTime ExpireTime             `json:"-"`
}
type GetVerificationCodeResp struct {
	CommResp
	Data GetVerificationCodeData `json:"data"`
}

type GetVerificationCodeData struct {
	InviteCode string `json:"invite_code"`
	Image      string `json:"image"`
}

type InviteCodeReq struct {
	OperationID string `json:"operationID" binding:"required"`

	Code        string `json:"code" binding:"omitempty"`
	Timezone    string `json:"timezone" binding:"required,utc,min=1,max=8"`
	Mobile      string `json:"mobile" binding:"omitempty,min=1,max=64"`
	Os          string `json:"os" binding:"required,alpha,min=1,max=64"`
	Version     string `json:"version" binding:"required,version_regular,min=1,max=64"`
	Webkit      string `json:"webkit" binding:"omitempty,version_regular,min=1,max=64"`
	ScreenWidth int    `json:"screen_width" binding:"required,number,min=1,max=9999"`
	Language    string `json:"language" binding:"required,language,min=1,max=8"`

	Test string `json:"test" binding:"omitempty,alphanum"`

	Ip         string `json:"ip"`
	CreateTime int64  `json:"create_time"`
}

var VersionRegular validator.Func = func(fl validator.FieldLevel) bool {
	data := fl.Field().Interface().(string)
	if ok, _ := regexp.Match("^\\d+(\\.\\d+)+$", []byte(data)); ok {
		return true
	} else {
		return false
	}
}
var Utc validator.Func = func(fl validator.FieldLevel) bool {
	data := fl.Field().Interface().(string)
	if ok, _ := regexp.Match("^UTC[+-]\\d{1,2}:\\d{1,3}$", []byte(data)); ok {
		return true
	} else {
		return false
	}
}
var Language validator.Func = func(fl validator.FieldLevel) bool {
	data := fl.Field().Interface().(string)
	if ok, _ := regexp.Match("^[A-Za-z]{1,2}$", []byte(data)); ok {
		return true
	} else {
		return false
	}
}

type PushCodeResp struct {
	CommResp
	Data PushCodeData `json:"data"`
}
type PushCodeData struct {
	InviteCode string `json:"invite_code"`
}

type UpdateUserIPandStatus struct {
	CommResp
}

type UpdateUserIPandStatusReq struct {
	UserID      string `json:"userID"`
	IPaddress   string `json:"ipAddress"`
	OperationID string `json:"operationID"`
}

type GetUserIPandStatusResp struct {
	CommResp
	Data GetUserIPandStatus `json:"data"`
}
type GetUserIPandStatus struct {
	UserID         string `json:"userID"`
	IPaddress      string `json:"ipAddress"`
	City           string `json:"city"`
	LastOnlineTime string `json:"lastOnlineTime"`
	OnlineDifValue int32  `json:"onlineDifValue"`
	OperationID    string `json:"operationID"`
	IsOnline       bool   `json:"isOnline"`
}

type GetUserIPandStatusReq struct {
	ForUserID   string `json:"forUserID" binding:"required"`
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"userID" binding:"required"`
}

type ChangePasswordRequest struct {
	OperationID string `json:"operationID" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=1,max=20"`
}

type ChangePasswordResponse struct {
	CommResp
}
