package base_info

type DelMsgReq struct {
	OpUserID    string   `json:"opUserID,omitempty"`
	UserID      string   `json:"userID,omitempty"`
	ChatType    int64    `json:"chatType" binding:"required"`
	SeqList     []uint32 `json:"seqList,omitempty"`
	OperationID string   `json:"operationID,omitempty"`
}

type DelMsgResp struct {
	CommResp
}

type CleanUpMsgReq struct {
	UserID      string `json:"userID"  binding:"required"`
	OperationID string `json:"operationID"  binding:"required"`
}

type CleanUpMsgResp struct {
	CommResp
}

type SwitchBroadcastReq struct {
	UserID      string `json:"user_id" binding:"required"`
	OperationID string `json:"operation_id" binding:"required"`
	Status      int8   `json:"status"`
}

type SwitchBroadcastResp struct {
	CommResp
}

type GetBroadcastStatusReq struct {
	OperationID string `json:"operation_id" binding:"required"`
}

type GetBroadcastStatusResp struct {
	CommResp
	Data   []map[string]interface{} `json:"data"`
	Status int8                     `json:"status"`
}

type ParamsGetMaxSeq struct {
	OpUserID    string `json:"opUserID" binding:"required"`
	GroupID     string `json:"groupID"`
	OperationID string `json:"operationID" binding:"required"`
}

type GetMaxSeqResponse struct {
	CommResp
	GroupID     string `json:"groupID"`
	OperationID string `json:"operationID" binding:"required"`
}

type CommDataRespOne struct {
	CommResp
	Data map[string]interface{} `json:"data"`
}
