package base_info

type GetPrivacyRequest struct {
	OperationID string `json:"operationID" binding:"required"`
}

type GetPrivacyResponse struct {
	CommResp
	Data []PrivacySetting `json:"data"`
}

type PrivacySetting struct {
	SettingKey string `json:"setting_key" binding:"required,oneof=privacy_add_by_phone privacy_add_by_account privacy_add_by_email privacy_see_wooms privacy_private_chat privacy_add_by_group privacy_add_by_qr privacy_add_by_contact_card"`
	SettingVal string `json:"setting_val" binding:"required,oneof=0 1"`
}

type SetPrivacyRequest struct {
	OperationID string            `json:"operationID" binding:"required"`
	Data        []*PrivacySetting `json:"data" binding:"dive"`
}

type SetPrivacyResponse struct {
	CommResp
}
