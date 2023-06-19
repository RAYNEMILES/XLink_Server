package apiThird

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRTCInvitationInfo(c *gin.Context) {
	var (
		req  api.GetRTCInvitationInfoReq
		resp api.GetRTCInvitationInfoResp
	)
	if err := c.Bind(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	var err error
	invitationInfo, err := db.DB.GetSignalInfoFromCacheByClientMsgID(req.ClientMsgID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetSignalInfoFromCache", err.Error(), req)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	if err := db.DB.DelUserSignalList(userID); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DelUserSignalList result:", err.Error())
	}

	resp.Data.OpUserID = invitationInfo.OpUserID
	resp.Data.Invitation.RoomID = invitationInfo.Invitation.RoomID
	resp.Data.Invitation.SessionType = invitationInfo.Invitation.SessionType
	resp.Data.Invitation.GroupID = invitationInfo.Invitation.GroupID
	resp.Data.Invitation.InviterUserID = invitationInfo.Invitation.InviterUserID
	resp.Data.Invitation.InviteeUserIDList = invitationInfo.Invitation.InviteeUserIDList
	resp.Data.Invitation.MediaType = invitationInfo.Invitation.MediaType
	resp.Data.Invitation.Timeout = invitationInfo.Invitation.Timeout
	resp.Data.Invitation.InitiateTime = invitationInfo.Invitation.InitiateTime
	resp.Data.Invitation.PlatformID = invitationInfo.Invitation.PlatformID
	resp.Data.Invitation.CustomData = invitationInfo.Invitation.CustomData
	c.JSON(http.StatusOK, resp)
}

func GetRTCInvitationInfoStartApp(c *gin.Context) {
	var (
		req  api.GetRTCInvitationInfoStartAppReq
		resp api.GetRTCInvitationInfoStartAppResp
	)
	if err := c.Bind(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req)

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	invitationInfo, err := db.DB.GetAvailableSignalInvitationInfo(userID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetSignalInfoFromCache", err.Error(), req)
		c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": err.Error(), "data": struct{}{}})
		return
	}
	resp.Data.OpUserID = invitationInfo.OpUserID
	resp.Data.Invitation.RoomID = invitationInfo.Invitation.RoomID
	resp.Data.Invitation.SessionType = invitationInfo.Invitation.SessionType
	resp.Data.Invitation.GroupID = invitationInfo.Invitation.GroupID
	resp.Data.Invitation.InviterUserID = invitationInfo.Invitation.InviterUserID
	resp.Data.Invitation.InviteeUserIDList = invitationInfo.Invitation.InviteeUserIDList
	resp.Data.Invitation.MediaType = invitationInfo.Invitation.MediaType
	resp.Data.Invitation.Timeout = invitationInfo.Invitation.Timeout
	resp.Data.Invitation.InitiateTime = invitationInfo.Invitation.InitiateTime
	resp.Data.Invitation.PlatformID = invitationInfo.Invitation.PlatformID
	resp.Data.Invitation.CustomData = invitationInfo.Invitation.CustomData
	c.JSON(http.StatusOK, resp)

}
