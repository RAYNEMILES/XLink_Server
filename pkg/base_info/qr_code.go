package base_info

type GetQrCodeRequest struct {
	OperationId string `json:"operationId" binding:"required"`
	Platform    int32  `json:"platform" binding:"required,min=1,max=7"`
	DeviceId    string `json:"deviceId" binding:"required"`
}

type GetQrCodeResponse struct {
	CommResp
	Data struct {
		QrCode string `json:"qrCode"`
	} `json:"data"`
}

type PushQrCodeRequest struct {
	OperationId string `json:"operationId" binding:"required"`
	PlatForm    int32  `json:"platform" binding:"required,min=1,max=7"`
	QrCodeId    string `json:"qrCodeId" binding:"required,uuid4"`
}

type PushQrCodeResponse struct {
	CommResp
	Data struct {
		TemporaryToken string `json:"temporaryToken"`
	} `json:"data"`
}

type ConfirmQrCodeRequest struct {
	OperationId    string `json:"operationId" binding:"required"`
	PlatForm       int32  `json:"platform" binding:"required,min=1,max=7"`
	QrCodeId       string `json:"qrCodeId" binding:"required,uuid4"`
	TemporaryToken string `json:"temporaryToken" binding:"required,uuid5"`
	Agree          *bool  `json:"agree" binding:"required"`
}

type ConfirmQrCodeResponse struct {
	CommResp
}

type CheckStateRequest struct {
	OperationId string `json:"operationId" binding:"required"`
	PlatForm    int32  `json:"platform" binding:"required,min=1,max=7"`
	QrCodeId    string `json:"qrCodeId" binding:"required,uuid4"`
	DeviceId    string `json:"deviceId" binding:"required"`
}

type CheckStateResponse struct {
	CommResp
	Data struct {
		UserID      string              `json:"userID"`
		State       int32               `json:"state"`
		Token       string              `json:"token"`
		ExpiredTime int64               `json:"expiredTime"`
		Extend      UserTokenInfoExtend `json:"extend"`
		UserSign    string              `json:"user_sign"`
	} `json:"data"`
}
