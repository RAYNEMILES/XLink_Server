package signaling

import (
	ws "Open_IM/cmd/Open-IM-SDK-Core/internal/interaction"
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	sdk2 "Open_IM/pkg/proto/sdk_ws"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/copier"
	"strings"
)

type LiveSignaling struct {
	*ws.Ws
	listener    open_im_sdk_callback.OnSignalingListener
	loginUserID string
	*db.DataBase
	platformID int32
	isCanceled bool
}

func NewLiveSignaling(ws *ws.Ws, listener open_im_sdk_callback.OnSignalingListener, loginUserID string, platformID int32, db *db.DataBase) *LiveSignaling {
	if ws == nil || listener == nil {
		log.Error("", "ws or listener is nil")
		return nil
	}
	return &LiveSignaling{Ws: ws, listener: listener, loginUserID: loginUserID, platformID: platformID, DataBase: db}
}

func (s *LiveSignaling) waitPush(req *sdk2.SignalReq, operationID string) {
	var invt *sdk2.InvitationInfo
	switch payload := req.Payload.(type) {
	case *sdk2.SignalReq_Invite:
		invt = payload.Invite.Invitation
	case *sdk2.SignalReq_InviteInGroup:
		invt = payload.InviteInGroup.Invitation
	}

	for _, v := range invt.InviteeUserIDList {
		go func(invitee string) {
			push, err := s.SignalingWaitPush(invt.InviterUserID, invitee, invt.RoomID, invt.Timeout, operationID)
			if err != nil {
				if strings.Contains(err.Error(), "timeout") {
					log.Error(operationID, "wait push timeout ", err.Error(), invt.InviterUserID, v, invt.RoomID, invt.Timeout)
					switch payload := req.Payload.(type) {
					case *sdk2.SignalReq_Invite:
						if !s.isCanceled {
							s.listener.OnInvitationTimeout(utils.StructToJsonString(payload.Invite))
						}
					case *sdk2.SignalReq_InviteInGroup:
						if !s.isCanceled {
							s.listener.OnInvitationTimeout(utils.StructToJsonString(payload.InviteInGroup))
						}
					}

				} else {
					log.Error(operationID, "other failed ", err.Error(), invt.InviterUserID, v, invt.RoomID, invt.Timeout)
				}
				return
			}
			log.Info(operationID, "SignalingWaitPush ", push.String(), invt.InviterUserID, v, invt.RoomID, invt.Timeout)
			s.doSignalPush(push, operationID)
		}(v)
	}
}
func (s *LiveSignaling) doSignalPush(req *sdk2.SignalReq, operationID string) {
	switch payload := req.Payload.(type) {
	//case *api.SignalReq_Invite:
	//	log.Info(operationID, "recv signal push ", payload.Invite.String())
	//	s.listener.OnReceiveNewInvitation(utils.StructToJsonString(payload.Invite))
	case *sdk2.SignalReq_Accept:
		log.Info(operationID, "recv signal push Accept ", payload.Accept.String())
		s.listener.OnInviteeAccepted(utils.StructToJsonString(payload.Accept))
	case *sdk2.SignalReq_Reject:
		log.Info(operationID, "recv signal push Reject ", payload.Reject.String())
		s.listener.OnInviteeRejected(utils.StructToJsonString(payload.Reject))
	//case *api.SignalReq_HungUp:
	//	log.Info(operationID, "recv signal push HungUp ", payload.HungUp.String())
	//	s.listener.OnHangUp(utils.StructToJsonString(payload.HungUp))
	//case *api.SignalReq_Cancel:
	//	log.Info(operationID, "recv signal push ", payload.Cancel.String())
	//	s.listener.OnInvitationCancelled(utils.StructToJsonString(payload.Cancel))
	//case *api.SignalReq_InviteInGroup:
	//	log.Info(operationID, "recv signal push ", payload.InviteInGroup.String())
	default:
		log.Error(operationID, "payload type failed ", payload)
	}
}

func (s *LiveSignaling) SetListener(listener open_im_sdk_callback.OnSignalingListener, operationID string) {
	s.listener = listener
}

func (s *LiveSignaling) getSelfParticipant(groupID string, callback open_im_sdk_callback.Base, operationID string) *sdk2.ParticipantMetaData {
	log.Info(operationID, utils.GetSelfFuncName(), "args ", groupID)
	p := sdk2.ParticipantMetaData{GroupInfo: &sdk2.GroupInfo{}, GroupMemberInfo: &sdk2.GroupMemberFullInfo{}, UserInfo: &sdk2.PublicUserInfo{}}
	if groupID != "" {
		g, err := s.GetGroupInfoByGroupID(groupID)
		common.CheckDBErrCallback(callback, err, operationID)
		copier.Copy(p.GroupInfo, g)
		mInfo, err := s.GetGroupMemberInfoByGroupIDUserID(groupID, s.loginUserID)
		common.CheckDBErrCallback(callback, err, operationID)
		copier.Copy(p.GroupMemberInfo, mInfo)
	}

	sf, err := s.GetLoginUser()
	common.CheckDBErrCallback(callback, err, operationID)
	copier.Copy(p.UserInfo, sf)
	log.Info(operationID, utils.GetSelfFuncName(), "return ", p)
	return &p
}

func (s *LiveSignaling) DoNotification(msg *sdk2.MsgData, conversationCh chan common.Cmd2Value, operationID string) {
	log.Info(operationID, utils.GetSelfFuncName(), "args ", msg.String())
	var resp sdk2.SignalReq
	err := proto.Unmarshal(msg.Content, &resp)
	if err != nil {
		log.Error(operationID, "Unmarshal failed")
		return
	}
	switch payload := resp.Payload.(type) {
	case *sdk2.SignalReq_Accept:
		log.Info(operationID, "signaling response ", payload.Accept.String())
		if payload.Accept.Invitation.InviterUserID == s.loginUserID && payload.Accept.Invitation.PlatformID == s.platformID {
			var wsResp ws.GeneralWsResp
			wsResp.ReqIdentifier = constant.WSSendSignalMsg
			wsResp.Data = msg.Content
			wsResp.MsgIncr = s.loginUserID + payload.Accept.OpUserID + payload.Accept.Invitation.RoomID
			log.Info(operationID, "search msgIncr: ", wsResp.MsgIncr)
			s.DoWSSignal(wsResp)
			return
		}
		if payload.Accept.OpUserPlatformID != s.platformID && payload.Accept.OpUserID == s.loginUserID {
			s.listener.OnInviteeAcceptedByOtherDevice(utils.StructToJsonString(payload.Accept))
			return
		}
	case *sdk2.SignalReq_Reject:
		log.Info(operationID, "signaling response ", payload.Reject.String())
		if payload.Reject.Invitation.InviterUserID == s.loginUserID && payload.Reject.Invitation.PlatformID == s.platformID {
			var wsResp ws.GeneralWsResp
			wsResp.ReqIdentifier = constant.WSSendSignalMsg
			wsResp.Data = msg.Content
			wsResp.MsgIncr = s.loginUserID + payload.Reject.OpUserID + payload.Reject.Invitation.RoomID
			log.Info(operationID, "search msgIncr: ", wsResp.MsgIncr)
			s.DoWSSignal(wsResp)
			return
		}
		if payload.Reject.OpUserPlatformID != s.platformID && payload.Reject.OpUserID == s.loginUserID {
			s.listener.OnInviteeRejectedByOtherDevice(utils.StructToJsonString(payload.Reject))
			return
		}

	case *sdk2.SignalReq_HungUp:
		log.Info(operationID, "signaling response HungUp", payload.HungUp.String())
		if s.loginUserID != payload.HungUp.OpUserID {
			s.listener.OnHangUp(utils.StructToJsonString(payload.HungUp))
		}
	case *sdk2.SignalReq_Cancel:
		log.Info(operationID, "signaling response ", payload.Cancel.String())
		if utils.IsContain(s.loginUserID, payload.Cancel.Invitation.InviteeUserIDList) {
			s.listener.OnInvitationCancelled(utils.StructToJsonString(payload.Cancel))
		}
	case *sdk2.SignalReq_Invite:
		log.Info(operationID, "signaling response ", payload.Invite.String())
		if utils.IsContain(s.loginUserID, payload.Invite.Invitation.InviteeUserIDList) {
			//	if s.loginUserID == payload.Invite.Invitation.InviterUserID {
			s.listener.OnReceiveNewInvitation(utils.StructToJsonString(payload.Invite))
		}

	case *sdk2.SignalReq_InviteInGroup:
		log.Info(operationID, "signaling response ", payload.InviteInGroup.String())
		if utils.IsContain(s.loginUserID, payload.InviteInGroup.Invitation.InviteeUserIDList) {
			//	if s.loginUserID == payload.InviteInGroup.Invitation.InviterUserID {
			s.listener.OnReceiveNewInvitation(utils.StructToJsonString(payload.InviteInGroup))
		}
	default:
		log.Error(operationID, "resp payload type failed ", payload)
	}
}

func (s *LiveSignaling) handleSignaling(req *sdk2.SignalReq, callback open_im_sdk_callback.Base, operationID string) {
	log.Info(operationID, utils.GetSelfFuncName(), "args ", req.String())
	resp, err := s.SendSignalingReqWaitResp(req, operationID)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "SendSignalingReqWaitResp error", err.Error())
		common.CheckAnyErrCallback(callback, 3003, errors.New("timeout"), operationID)
	}
	common.CheckAnyErrCallback(callback, 3001, err, operationID)
	switch payload := resp.Payload.(type) {
	case *sdk2.SignalResp_Accept:
		log.Info(operationID, "signaling response ", payload.Accept.String())
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.AcceptCallback(payload.Accept)))
	case *sdk2.SignalResp_Reject:
		log.Info(operationID, "signaling response ", payload.Reject.String())
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.RejectCallback(payload.Reject)))
	case *sdk2.SignalResp_HungUp:
		log.Info(operationID, "signaling response ", payload.HungUp.String())
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.HungUpCallback(payload.HungUp)))
	case *sdk2.SignalResp_Cancel:
		s.isCanceled = true
		log.Info(operationID, "signaling response ", payload.Cancel.String())
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.CancelCallback(payload.Cancel)))
	case *sdk2.SignalResp_Invite:
		s.isCanceled = false
		log.Info(operationID, "signaling response ", payload.Invite.String())
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.InviteCallback(payload.Invite)))
	case *sdk2.SignalResp_InviteInGroup:
		s.isCanceled = false
		log.Info(operationID, "signaling response ", payload.InviteInGroup.String())
		callback.OnSuccess(utils.StructToJsonString(sdk_params_callback.InviteInGroupCallback(payload.InviteInGroup)))
	default:
		log.Error(operationID, "resp payload type failed ", payload)
		common.CheckAnyErrCallback(callback, 3002, errors.New("resp payload type failed"), operationID)
	}
	switch req.Payload.(type) {
	case *sdk2.SignalReq_Invite:
		log.Info(operationID, "wait push ", req.String())
		s.waitPush(req, operationID)
	case *sdk2.SignalReq_InviteInGroup:
		log.Info(operationID, "wait push ", req.String())
		s.waitPush(req, operationID)
	}
}
