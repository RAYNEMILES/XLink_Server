package base_info

import "mime/multipart"

type MinioStorageCredentialReq struct {
	OperationID string `json:"operationID"`
}

type MiniostorageCredentialResp struct {
	SecretAccessKey string `json:"secretAccessKey"`
	AccessKeyID     string `json:"accessKeyID"`
	SessionToken    string `json:"sessionToken"`
	BucketName      string `json:"bucketName"`
	StsEndpointURL  string `json:"stsEndpointURL"`
}

type MinioUploadFileReq struct {
	OperationID string `form:"operationID" binding:"required"`
	FileType    int    `form:"fileType" binding:"required"`
}

type MinioUploadFileResp struct {
	URL             string `json:"URL"`
	NewName         string `json:"newName"`
	SnapshotURL     string `json:"snapshotURL,omitempty"`
	SnapshotNewName string `json:"snapshotName,omitempty"`
	ImageWidth      int    `json:"imageWidth,omitempty"`
	ImageHeight     int    `json:"imageHeight,omitempty"`
}

type UploadUpdateAppReq struct {
	OperationID string                `form:"operationID" binding:"required"`
	Type        int                   `form:"type" binding:"required"`
	Version     string                `form:"version"  binding:"required"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
	Yaml        *multipart.FileHeader `form:"yaml"`
	ForceUpdate bool                  `form:"forceUpdate"`
	UpdateLog   string                `form:"updateLog" binding:"required"`
}

type UploadUpdateAppResp struct {
	CommResp
}

type GetDownloadURLReq struct {
	OperationID string `json:"operationID" binding:"required"`
	Type        int    `json:"type" binding:"required"`
	Version     string `json:"version" binding:"required"`
}

type GetDownloadURLResp struct {
	CommResp
	Data struct {
		HasNewVersion bool   `json:"hasNewVersion"`
		ForceUpdate   bool   `json:"forceUpdate"`
		FileURL       string `json:"fileURL"`
		YamlURL       string `json:"yamlURL"`
		Version       string `json:"version"`
		UpdateLog     string `json:"update_log"`
	} `json:"data"`
}

type GetRTCInvitationInfoReq struct {
	OperationID string `json:"operationID" binding:"required"`
	ClientMsgID string `json:"clientMsgID" binding:"required"`
}

type GetRTCInvitationInfoResp struct {
	CommResp
	Data struct {
		OpUserID   string `json:"opUserID"`
		Invitation struct {
			InviterUserID     string   `json:"inviterUserID"`
			InviteeUserIDList []string `json:"inviteeUserIDList"`
			GroupID           string   `json:"groupID"`
			RoomID            string   `json:"roomID"`
			Timeout           int32    `json:"timeout"`
			MediaType         string   `json:"mediaType"`
			SessionType       int32    `json:"sessionType"`
			InitiateTime      int32    `json:"initiateTime"`
			PlatformID        int32    `json:"platformID"`
			CustomData        string   `json:"customData"`
		} `json:"invitation"`
		OfflinePushInfo struct{} `json:"offlinePushInfo"`
	} `json:"data"`
}

type GetRTCInvitationInfoStartAppReq struct {
	OperationID string `json:"operationID" binding:"required"`
}

type GetRTCInvitationInfoStartAppResp struct {
	GetRTCInvitationInfoResp
}

type CheckStatusReq struct {
	OperationID string `json:"operationID" binding:"required"`
	UserID      string `json:"user_id"`
	GroupID     string `json:"group_id"`
	ChatType    int8   `json:"chat_type"`
}

type CheckStatusResp struct {
	CommResp
}

type StartCommunicationReq struct {
	OperationID        string `json:"operationID" binding:"required"`
	OriginatorPlatform int32  `json:"originator_platform" binding:"required"`
	RoomID             string `json:"room_id"`
	RoomIdType         uint64 `json:"room_id_type"`
	GroupID            string `json:"group_id"`
	Receiver           string `json:"receiver"`
	ChatType           int8   `json:"chat_type" binding:"required"`
	Supporter          string `json:"supporter" binding:"required"`
}

type StartCommunicationResp struct {
	CommResp
	CommunicationID int64 `json:"communication_id"`
}

type EndCommunicationReq struct {
	OperationID string `json:"operationID" binding:"required"`
	RoomId      string `json:"room_id"`
	RoomIdType  int64  `json:"room_id_type"`
	ErrCode     int64  `json:"err_code"`
}

type EndCommunicationResp struct {
	CommResp
}

type GetMembersByCommunicationIDReq struct {
	OperationID     string `json:"operationID" binding:"required"`
	CommunicationID int64  `json:"communication_id" binding:"required"`
}

type GetMembersByCommunicationIDResp struct {
	CommResp
	MasterID  string   `json:"master_id"`
	GroupID   string   `json:"group_id"`
	MembersID []string `json:"members_id"`
}

type StartRecordReq struct {
	OperationID string `json:"operationID" binding:"required"`
	RoomId      string `json:"room_id"`
	RoomIdType  uint64 `json:"room_id_type"`
}

type StartRecordResp struct {
	CommResp
}

type EndRecordReq struct {
	OperationID string `json:"operationID" binding:"required"`
	RoomId      string `json:"room_id"`
}

type EndRecordResp struct {
}

type JoinCommunicationReq struct {
	OperationID string `json:"operationID" binding:"required"`
	RoomID      string `json:"room_id"`
	RoomIDType  int64  `json:"room_id_type"`
	ChatType    int32  `json:"chat_type"`
	GroupID     string `json:"group_id"`
}

type JoinCommunicationResp struct {
	CommResp
}

type RecordCallBackReq struct {
	EventGroupId uint64 `json:"EventGroupId"`
	EventType    uint64 `json:"EventType"`
	CallbackTs   uint64 `json:"CallbackTs"`
	EventInfo    struct {
		RoomId  uint64 `json:"RoomId"`
		EventTs uint64 `json:"EventTs"`
		UserId  string `json:"UserId"`
		TaskId  string `json:"UniqueId"`
		Payload struct {
			Status   uint64   `json:"Status"`
			FileList []string `json:"FileList"`
		} `json:"Payload"`
	} `json:"EventInfo"`
}

type SaveMediaReq struct {
	OperationID string `json:"operation_id"`
	CallID      string `json:"call_id"`
	RecordURL   string `json:"record_url"`
}

type SaveMediaResp struct {
	CommResp
}
