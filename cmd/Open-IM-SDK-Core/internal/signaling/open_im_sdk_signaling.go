package signaling

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	sdk2 "Open_IM/pkg/proto/sdk_ws"
)

func (s *LiveSignaling) SetDefaultReq(req *sdk2.InvitationInfo) {
	if req.RoomID == "" {
		req.RoomID = utils.OperationIDGenerator()
	}
	if req.Timeout == 0 {
		req.Timeout = 60 * 60
	}
}

func (s *LiveSignaling) InviteInGroup(callback open_im_sdk_callback.Base, signalInviteInGroupReq string, operationID string) {
	if callback == nil {
		log.Error(operationID, "callback is nil")
		return
	}
	if s.listener == nil {
		log.Error(operationID, "listener is nil")
		callback.OnError(3004, "listener is nil")
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", signalInviteInGroupReq)
		req := &sdk2.SignalReq_InviteInGroup{InviteInGroup: &sdk2.SignalInviteInGroupReq{Invitation: &sdk2.InvitationInfo{}, OfflinePushInfo: &sdk2.OfflinePushInfo{}}}
		var signalReq sdk2.SignalReq
		common.JsonUnmarshalCallback(signalInviteInGroupReq, req.InviteInGroup, callback, operationID)
		s.SetDefaultReq(req.InviteInGroup.Invitation)
		req.InviteInGroup.Invitation.InviterUserID = s.loginUserID
		req.InviteInGroup.OpUserID = s.loginUserID
		req.InviteInGroup.Invitation.InitiateTime = int32(utils.GetCurrentTimestampBySecond())
		signalReq.Payload = req
		req.InviteInGroup.Participant = s.getSelfParticipant(req.InviteInGroup.Invitation.GroupID, callback, operationID)
		s.handleSignaling(&signalReq, callback, operationID)
		log.NewInfo(operationID, fName, " callback: finished")
	}()
}

func (s *LiveSignaling) Invite(callback open_im_sdk_callback.Base, signalInviteReq string, operationID string) {
	if callback == nil {
		log.Error(operationID, "callback is nil")
		return
	}
	if s.listener == nil {
		log.Error(operationID, "listener is nil")
		callback.OnError(3004, "listener is nil")
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", signalInviteReq)
		req := &sdk2.SignalReq_Invite{Invite: &sdk2.SignalInviteReq{Invitation: &sdk2.InvitationInfo{}, OfflinePushInfo: &sdk2.OfflinePushInfo{}}}
		var signalReq sdk2.SignalReq
		common.JsonUnmarshalCallback(signalInviteReq, req.Invite, callback, operationID)
		s.SetDefaultReq(req.Invite.Invitation)
		req.Invite.Invitation.InviterUserID = s.loginUserID
		req.Invite.OpUserID = s.loginUserID
		req.Invite.Invitation.InitiateTime = int32(utils.GetCurrentTimestampBySecond())
		signalReq.Payload = req
		req.Invite.Participant = s.getSelfParticipant(req.Invite.Invitation.GroupID, callback, operationID)
		s.handleSignaling(&signalReq, callback, operationID)
		log.NewInfo(operationID, fName, " callback: finished")
	}()
}

func (s *LiveSignaling) Accept(callback open_im_sdk_callback.Base, signalAcceptReq string, operationID string) {
	if callback == nil {
		log.Error(operationID, "callback is nil")
		return
	}
	if s.listener == nil {
		log.Error(operationID, "listener is nil")
		callback.OnError(3004, "listener is nil")
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", signalAcceptReq)
		req := &sdk2.SignalReq_Accept{Accept: &sdk2.SignalAcceptReq{Invitation: &sdk2.InvitationInfo{}, OfflinePushInfo: &sdk2.OfflinePushInfo{}}}
		var signalReq sdk2.SignalReq
		common.JsonUnmarshalCallback(signalAcceptReq, req.Accept, callback, operationID)
		s.SetDefaultReq(req.Accept.Invitation)
		req.Accept.OpUserID = s.loginUserID
		req.Accept.Invitation.InitiateTime = int32(utils.GetCurrentTimestampBySecond())
		signalReq.Payload = req
		req.Accept.Participant = s.getSelfParticipant(req.Accept.Invitation.GroupID, callback, operationID)
		req.Accept.OpUserPlatformID = s.platformID
		s.handleSignaling(&signalReq, callback, operationID)
		log.NewInfo(operationID, fName, " callback finished")
	}()
}

func (s *LiveSignaling) Reject(callback open_im_sdk_callback.Base, signalRejectReq string, operationID string) {
	if callback == nil {
		log.NewError(operationID, "callback is nil")
		return
	}
	if s.listener == nil {
		log.Error(operationID, "listener is nil")
		callback.OnError(3004, "listener is nil")
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", signalRejectReq)
		req := &sdk2.SignalReq_Reject{Reject: &sdk2.SignalRejectReq{Invitation: &sdk2.InvitationInfo{}, OfflinePushInfo: &sdk2.OfflinePushInfo{}}}
		var signalReq sdk2.SignalReq
		common.JsonUnmarshalCallback(signalRejectReq, req.Reject, callback, operationID)
		s.SetDefaultReq(req.Reject.Invitation)
		req.Reject.OpUserID = s.loginUserID
		req.Reject.Invitation.InitiateTime = int32(utils.GetCurrentTimestampBySecond())
		signalReq.Payload = req
		req.Reject.OpUserPlatformID = s.platformID
		req.Reject.Participant = s.getSelfParticipant(req.Reject.Invitation.GroupID, callback, operationID)
		s.handleSignaling(&signalReq, callback, operationID)
		log.NewInfo(operationID, fName, " callback finished")
	}()
}

func (s *LiveSignaling) Cancel(callback open_im_sdk_callback.Base, signalCancelReq string, operationID string) {
	if callback == nil {
		log.NewError(operationID, "callback is nil")
	}
	if s.listener == nil {
		log.Error(operationID, "listener is nil")
		callback.OnError(3004, "listener is nil")
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", signalCancelReq)
		req := &sdk2.SignalReq_Cancel{Cancel: &sdk2.SignalCancelReq{Invitation: &sdk2.InvitationInfo{}, OfflinePushInfo: &sdk2.OfflinePushInfo{}}}
		var signalReq sdk2.SignalReq
		common.JsonUnmarshalCallback(signalCancelReq, req.Cancel, callback, operationID)
		s.SetDefaultReq(req.Cancel.Invitation)
		req.Cancel.OpUserID = s.loginUserID
		req.Cancel.Invitation.InitiateTime = int32(utils.GetCurrentTimestampBySecond())
		signalReq.Payload = req
		s.handleSignaling(&signalReq, callback, operationID)
		log.NewInfo(operationID, fName, " callback finished")
	}()
}

func (s *LiveSignaling) HungUp(callback open_im_sdk_callback.Base, signalHungUpReq string, operationID string) {
	if callback == nil {
		log.NewError(operationID, "callback is nil")
	}
	if s.listener == nil {
		log.Error(operationID, "listener is nil")
		callback.OnError(3004, "listener is nil")
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", signalHungUpReq)
		req := &sdk2.SignalReq_HungUp{HungUp: &sdk2.SignalHungUpReq{Invitation: &sdk2.InvitationInfo{}, OfflinePushInfo: &sdk2.OfflinePushInfo{}}}
		var signalReq sdk2.SignalReq
		common.JsonUnmarshalCallback(signalHungUpReq, req.HungUp, callback, operationID)
		s.SetDefaultReq(req.HungUp.Invitation)
		req.HungUp.OpUserID = s.loginUserID
		req.HungUp.Invitation.InitiateTime = int32(utils.GetCurrentTimestampBySecond())
		signalReq.Payload = req
		s.handleSignaling(&signalReq, callback, operationID)
		log.NewInfo(operationID, fName, " callback finished")
	}()
}
