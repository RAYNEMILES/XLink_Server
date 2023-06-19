package server_api_params

type GetMembersByRoomIDReq struct {
	OperationID     string `json:"operationID" binding:"required"`
	CommunicationID int64  `json:"communication_id"`
}

type GetMembersByRoomIDResp struct {
	CommResp
	MasterID  string   `json:"master_id"`
	GroupID   string   `json:"group_id"`
	MembersID []string `json:"members_id"`
}
