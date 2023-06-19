package cms_api_struct

type Communication struct {
	CommunicationID    int64    `json:"communication_id"`
	Originator         string   `json:"originator"`
	OriginatorName     string   `json:"originator_name"`
	CallID             string   `json:"call_id"`
	OriginatorPlatform int32    `json:"originator_platform"`
	GroupID            string   `json:"group_id"`
	GroupName          string   `json:"group_name"`
	MemberIDs          []string `json:"member_ids"`
	MemberIDNames      []string `json:"member_id_names"`
	Status             int32    `json:"status"`
	Duration           int64    `json:"duration"`
	StartTime          int64    `json:"start_time"`
	EndTime            int64    `json:"end_time"`
	RecordURL          string   `json:"record_url"`
	ChatType           int8     `json:"chat_type"`
	Supporter          string   `json:"supporter"`
	Remark             string   `json:"remark"`
	DeleteBy           string   `json:"delete_by"`
	DeleteTime         int64    `json:"delete_time"`

	RecordVideoURL string `json:"record_video_url"`
	ErrCode        string `json:"err_code"`
	ErrMsg         string `json:"err_msg"`
}

type GetCommunicationsReq struct {
	RequestPagination
	OperationID        string `form:"operation_id" binding:"omitempty"`
	Originator         string `form:"originator" binding:"omitempty"`
	Member             string `form:"member" binding:"omitempty"`
	OriginatorPlatform int32  `form:"originator_platform" binding:"omitempty,numeric"`
	ChatType           int8   `form:"chat_type" binding:"omitempty,numeric"`
	Duration           int64  `form:"duration" binding:"omitempty,numeric"`
	TimeType           int32  `form:"time_type" binding:"omitempty,numeric"`
	StartTime          string `form:"start_time" binding:"omitempty,numeric"`
	EndTime            string `form:"end_time" binding:"omitempty,numeric"`
	Status             int32  `form:"status" binding:"omitempty"`
	Remark             string `form:"remark" binding:"omitempty"`
	OrderBy            string `form:"order_by" binding:"omitempty"`
	CommunicationType  int32  `form:"communication_type" binding:"omitempty,numeric"`
	RoomID             string `form:"room_id" binding:"omitempty"`
}

type GetCommunicationsResp struct {
	CommunicationList []*Communication `json:"communication_list"`
	Communications    int32            `json:"communications"`
	ResponsePagination
}

type DeleteCommunicationsReq struct {
	OperationID   string  `json:"operation_id"`
	CommunicatIDs []int64 `json:"communicat_ids"`
}

type DeleteCommunicationsResp struct {
	CommResp
}

type SetRemarkReq struct {
	OperationID  string `json:"operation_id"`
	CommunicatID int64  `json:"communicat_id"`
	Remark       string `json:"remark"`
}

type SetRemarkResp struct {
	CommResp
}

type InterruptCommunicationReq struct {
	OperationID  string `json:"operation_id"`
	CommunicatID int64  `json:"communicat_id"`
}

type InterruptCommunicationResp struct {
	CommResp
}
