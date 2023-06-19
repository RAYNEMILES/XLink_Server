package interaction

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/internal/msg_gateway/gate"
	"Open_IM/pkg/common/log"
	"encoding/json"
	log2 "log"
	"runtime"
	"strings"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	sdk "Open_IM/pkg/proto/sdk_ws"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"time"
)

type Ws struct {
	*WsRespAsyn
	*WsConn
	//*db.DataBase
	//conversationCh chan common.Cmd2Value
	cmdCh              chan common.Cmd2Value //waiting logout cmd
	pushMsgAndMaxSeqCh chan common.Cmd2Value //recv push msg  -> channel
	cmdHeartbeatCh     chan common.Cmd2Value //
	JustOnceFlag       bool
}

func NewWs(wsRespAsyn *WsRespAsyn, wsConn *WsConn, cmdCh chan common.Cmd2Value, pushMsgAndMaxSeqCh chan common.Cmd2Value, cmdHeartbeatCh chan common.Cmd2Value) *Ws {
	p := Ws{WsRespAsyn: wsRespAsyn, WsConn: wsConn, cmdCh: cmdCh, pushMsgAndMaxSeqCh: pushMsgAndMaxSeqCh, cmdHeartbeatCh: cmdHeartbeatCh}
	go p.ReadData()
	return &p
}

//func (w *Ws) SeqMsg() map[int32]server_api_params.MsgData {
//	w.seqMsgMutex.RLock()
//	defer w.seqMsgMutex.RUnlock()
//	return w.seqMsg
//}
//
//func (w *Ws) SetSeqMsg(seqMsg map[int32]server_api_params.MsgData) {
//	w.seqMsgMutex.Lock()
//	defer w.seqMsgMutex.Unlock()
//	w.seqMsg = seqMsg
//}

func (w *Ws) WaitResp(ch chan GeneralWsResp, timeout int, operationID string, connSend *websocket.Conn) (*GeneralWsResp, error) {
	//t := time.Now()
	select {
	case r := <-ch:
		log.Debug(operationID, "ws ch recvMsg success, code ", r.ErrCode)
		if r.ErrCode != 0 {
			log.Error(operationID, "ws ch recvMsg failed, code, err msg: ", r.ErrCode, r.ErrMsg)
			switch r.ErrCode {
			case int(constant.ErrInBlackList.ErrCode):
				return nil, &constant.ErrInBlackList
			case int(constant.ErrNotFriend.ErrCode):
				return nil, &constant.ErrNotFriend
			}
			return nil, errors.New(utils.IntToString(r.ErrCode) + ":" + r.ErrMsg)
		} else {
			return &r, nil
		}

	case <-time.After(time.Second * time.Duration(timeout)):
		log.Error(operationID, "ws ch recvMsg err, timeout")
		if connSend == nil {
			return nil, errors.New("ws ch recvMsg err, timeout")
		}
		if connSend != w.WsConn.conn {
			return nil, constant.WsRecvConnDiff
		} else {
			return nil, constant.WsRecvConnSame
		}
	}
}

func (w *Ws) SendReqWaitResp(m proto.Message, reqIdentifier int32, timeout, retryTimes int, senderID, operationID string) (*GeneralWsResp, error) {
	//t := time.Now()
	var wsReq GeneralWsReq
	var connSend *websocket.Conn
	var err error
	wsReq.ReqIdentifier = reqIdentifier
	wsReq.OperationID = operationID
	msgIncr, ch := w.AddCh(senderID)
	log.Debug(wsReq.OperationID, "SendReqWaitResp AddCh msgIncr:", msgIncr, reqIdentifier)
	defer w.DelCh(msgIncr)
	defer log.Debug(wsReq.OperationID, "SendReqWaitResp DelCh msgIncr:", msgIncr, reqIdentifier)
	wsReq.SendID = senderID
	wsReq.PlatformID = w.WsConn.platformID
	wsReq.Token = w.WsConn.token
	wsReq.MsgIncr = msgIncr
	wsReq.Data, err = proto.Marshal(m)
	if err != nil {
		return nil, utils.Wrap(err, "proto marshal err")
	}
	flag := 0
	for i := 0; i < retryTimes+1; i++ {
		connSend, err = w.writeBinaryMsg(wsReq)
		if err != nil {
			if !w.IsWriteTimeout(err) {
				//log.Error(operationID, "Not send timeout, failed, close conn, writeBinaryMsg again ", err.Error())
				if strings.Contains(err.Error(), "broken pipe") {
					w.SetLoginState(constant.Disconnected)
				} else {
					w.SetLoginState(constant.Logout)
				}

				time.Sleep(time.Duration(1) * time.Second)
				continue
			} else {
				w.SetLoginState(constant.Disconnected)
				return nil, utils.Wrap(err, "writeBinaryMsg timeout")
			}
		}
		flag = 1
		break
	}
	if flag == 1 {
		log.Debug(operationID, "send ok wait resp")
		r1, r2 := w.WaitResp(ch, timeout, wsReq.OperationID, connSend)
		return r1, r2
	} else {
		log.Error(operationID, "send failed")
		err := errors.New("send failed")
		return nil, utils.Wrap(err, "SendReqWaitResp failed")
	}
}
func (w *Ws) SendReqTest(m proto.Message, reqIdentifier int32, timeout int, senderID, operationID string) bool {
	var wsReq GeneralWsReq
	var connSend *websocket.Conn
	var err error
	wsReq.ReqIdentifier = reqIdentifier
	wsReq.OperationID = operationID
	msgIncr, ch := w.AddCh(senderID)
	defer w.DelCh(msgIncr)
	wsReq.SendID = senderID
	wsReq.PlatformID = w.platformID
	wsReq.MsgIncr = msgIncr
	wsReq.Data, err = proto.Marshal(m)
	if err != nil {
		return false
	}
	connSend, err = w.writeBinaryMsg(wsReq)
	if err != nil {
		log2.Println(operationID, "writeBinaryMsg timeout", m.String(), senderID, err.Error())
		log.Debug(operationID, "writeBinaryMsg timeout", m.String(), senderID, err.Error())
		return false
	} else {
		log2.Println(operationID, "writeBinaryMsg success", m.String(), senderID)
		log.Debug(operationID, "writeBinaryMsg success", m.String(), senderID)
	}
	startTime := time.Now()
	result := w.WaitTest(ch, timeout, wsReq.OperationID, connSend, m, senderID)
	log2.Println(operationID, "ws Response time：", time.Since(startTime), m.String(), senderID, result)
	log.Debug(operationID, "ws Response time：", time.Since(startTime), m.String(), senderID, result)
	return result
}
func (w *Ws) WaitTest(ch chan GeneralWsResp, timeout int, operationID string, connSend *websocket.Conn, m proto.Message, senderID string) bool {
	select {
	case r := <-ch:
		if r.ErrCode != 0 {
			log.Debug(operationID, "ws ch recvMsg success, code ", r.ErrCode, r.ErrMsg, m.String(), senderID)
			return false
		} else {
			log.Debug(operationID, "ws ch recvMsg send success, code ", m.String(), senderID)

			return true
		}

	case <-time.After(time.Second * time.Duration(timeout)):
		log.Debug(operationID, "ws ch recvMsg err, timeout ", m.String(), senderID)

		return false
	}
}
func (w *Ws) reConnSleep(operationID string, sleep int32) (error, bool) {
	_, err, isNeedReConn := w.WsConn.ReConn()
	if err != nil {
		log.Error(operationID, "ReConn failed ", err.Error(), "is need re connect ", isNeedReConn)
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return err, isNeedReConn
}

func (w *Ws) ReadData() {
	defer func() {
		if r := recover(); r != nil {
			log.NewError("", "ReadData panic", " panic is ", r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "ReadData panic", "panic:", string(buf))
		}
	}()
	isErrorOccurred := false
	for {
		operationID := utils.OperationIDGenerator()
		if isErrorOccurred {
			select {
			case r := <-w.cmdCh:
				if r.Cmd == constant.CmdLogout {
					log.Warn(operationID, "recv CmdLogout, return, close conn")
					log.Warn(operationID, "close ws read channel ", w.cmdCh)
					//		close(w.cmdCh)
					w.SetLoginState(constant.Logout)
					return
				}
				log.Warn(operationID, "other cmd ...", r.Cmd)
			case <-time.After(time.Microsecond * time.Duration(100)):
				log.Warn(operationID, "timeout(ms)... ", 100)
			}
		}
		isErrorOccurred = false
		if w.WsConn.conn == nil || w.LoginState() == constant.Disconnected {
			w.SetLoginState(constant.Disconnected)
			//log.Warn(operationID, "conn == nil, ReConn")
			//err, isNeedReConnect := w.reConnSleep(operationID, 1)
			//if err != nil && isNeedReConnect == false {
			//	log.Warn(operationID, "token failed, don't connect again")
			//	return
			//}
			continue
		}

		//w.WsConn.SetReadTimeout(5)
		msgType, message, err := w.WsConn.conn.ReadMessage()
		if err != nil {
			isErrorOccurred = true
			if w.loginState == constant.Logout {
				log.Warn(operationID, "loginState == logout ")
				log.Warn(operationID, "close ws read channel ", w.cmdCh)
				//	close(w.cmdCh)
				return
			}
			if w.WsConn.IsFatalError(err) {
				//log.Error(operationID, "IsFatalError ", err.Error(), "ReConn")
				//err, isNeedReConnect := w.reConnSleep(operationID, 5)
				//if err != nil && isNeedReConnect == false {
				//	log.Error(operationID, "token failed, don't connect again")
				//	return
				//}
				//return
			} else {
				log.Error(operationID, "timeout failed ", err.Error())

			}
			w.SetLoginState(constant.Disconnected)
			continue
		}
		log.NewInfo(operationID, utils.GetSelfFuncName(), "msgType", msgType, "message", string(message))
		if msgType == websocket.CloseMessage {
			w.SetLoginState(constant.Disconnected)
			log.Error(operationID, "type websocket.CloseMessage, ReConn")
			err, isNeedReConnect := w.reConnSleep(operationID, 1)
			if err != nil && isNeedReConnect == false {
				log.Warn(operationID, "token failed, don't connect again")
				return
			}
			continue
		} else if msgType == websocket.TextMessage {
			log.Warn(operationID, "type websocket.TextMessage")
			go w.doTextMsg(message)
		} else if msgType == websocket.BinaryMessage {
			go w.doWsMsg(message)
		} else {
			log.Warn(operationID, "recv other type ", msgType)
		}
	}
}

func (w *Ws) doWsMsg(message []byte) {
	wsResp, err := w.decodeBinaryWs(message)
	if err != nil {
		log.Error("decodeBinaryWs err", err.Error())
		return
	}
	log.NewDebug(wsResp.OperationID, "ws recv msg, code: ", wsResp.ErrCode, wsResp.ReqIdentifier)
	switch wsResp.ReqIdentifier {
	case constant.WSGetNewestSeq:
		if err = w.doWSGetNewestSeq(*wsResp); err != nil {
			log.Error(wsResp.OperationID, "doWSGetNewestSeq failed ", err.Error(), wsResp.ReqIdentifier, wsResp.MsgIncr)
		}
	case constant.WSPullMsgBySeqList:
		if err = w.doWSPullMsg(*wsResp); err != nil {
			log.Error(wsResp.OperationID, "doWSPullMsg failed ", err.Error())
		}
	case constant.WSPushMsg:
		if err = w.doWSPushMsg(*wsResp); err != nil {
			log.Error(wsResp.OperationID, "doWSPushMsg failed ", err.Error())
		}
		//if err = w.doWSPushMsgForTest(*wsResp); err != nil {
		//	log.Error(wsResp.OperationID, "doWSPushMsgForTest failed ", err.Error())
		//}

	case constant.WSSendMsg:
		if err = w.doWSSendMsg(*wsResp); err != nil {
			w.SetLoginState(constant.Disconnected)
			log.Error(wsResp.OperationID, "doWSSendMsg failed ", err.Error(), wsResp.ReqIdentifier, wsResp.MsgIncr)
		}
	case constant.WSSendBroadcastMsg:
		if err = w.doWSSendMsg(*wsResp); err != nil {
			w.SetLoginState(constant.Disconnected)
			log.Error(wsResp.OperationID, "doWSSendBroadcastMsg failed ", err.Error(), wsResp.ReqIdentifier, wsResp.MsgIncr)
		}
	case constant.WSKickOnlineMsg:
		log.Warn(wsResp.OperationID, "kick...  logout")
		w.kickOnline(*wsResp)
		w.Logout(wsResp.OperationID)

	case constant.WsLogoutMsg:
		log.Warn(wsResp.OperationID, "logout... ")
	case constant.WSSendSignalMsg:
		log.Info(wsResp.OperationID, "signaling...")
		w.DoWSSignal(*wsResp)
	default:
		log.Error(wsResp.OperationID, "type failed, ", wsResp.ReqIdentifier)
		return
	}
}

func (w *Ws) doTextMsg(message []byte) {
	wsResp := &GeneralWsResp{}
	err := json.Unmarshal(message, wsResp)
	if err != nil {
		log.Error("doTextMsg Unmarshal err", err.Error())
		return
	}
	log.Debug(wsResp.OperationID, "ws recv msg, code: ", wsResp.ErrCode, wsResp.ReqIdentifier)
	switch wsResp.ReqIdentifier {
	case constant.WsSyncDataMsg:
		log.Info(wsResp.OperationID, "syncdata...", string(wsResp.Data))
		w.doDataSynchronization(wsResp)
	default:
		log.Error(wsResp.OperationID, "type failed, ", wsResp.ReqIdentifier)
		return
	}
}

func (w *Ws) Logout(operationID string) {
	w.SetLoginState(constant.Logout)
	w.CloseConn()
	log.Warn(operationID, "TriggerCmdLogout ws...")
	err := common.TriggerCmdLogout(w.cmdCh)
	if err != nil {
		log.Error(operationID, "TriggerCmdLogout failed ", err.Error())
	}
	log.Info(operationID, "TriggerCmdLogout heartbeat...")
	err = common.TriggerCmdLogout(w.cmdHeartbeatCh)
	if err != nil {
		log.Error(operationID, "TriggerCmdLogout failed ", err.Error())
	}
}

func (w *Ws) doWSGetNewestSeq(wsResp GeneralWsResp) error {
	if err := w.notifyResp(wsResp); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (w *Ws) doWSPullMsg(wsResp GeneralWsResp) error {
	if err := w.notifyResp(wsResp); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (w *Ws) doWSSendMsg(wsResp GeneralWsResp) error {
	if err := w.notifyResp(wsResp); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (w *Ws) DoWSSignal(wsResp GeneralWsResp) error {
	if err := w.notifyResp(wsResp); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (w *Ws) doWSPushMsg(wsResp GeneralWsResp) error {
	if wsResp.ErrCode != 0 {
		return utils.Wrap(errors.New("errCode"), wsResp.ErrMsg)
	}
	var msg sdk.MsgData
	err := proto.Unmarshal(wsResp.Data, &msg)
	if err != nil {
		return utils.Wrap(err, "Unmarshal failed")
	}
	return utils.Wrap(common.TriggerCmdPushMsg(sdk_struct.CmdPushMsgToMsgSync{Msg: &msg, OperationID: wsResp.OperationID}, w.pushMsgAndMaxSeqCh), "")
}

func (w *Ws) doWSPushMsgForTest(wsResp GeneralWsResp) error {
	if wsResp.ErrCode != 0 {
		return utils.Wrap(errors.New("errCode"), wsResp.ErrMsg)
	}
	var msg sdk.MsgData
	err := proto.Unmarshal(wsResp.Data, &msg)
	if err != nil {
		return utils.Wrap(err, "Unmarshal failed")
	}
	log.Debug(wsResp.OperationID, "recv push doWSPushMsgForTest")
	return nil
	//	return utils.Wrap(common.TriggerCmdPushMsg(sdk_struct.CmdPushMsgToMsgSync{Msg: &msg, OperationID: wsResp.OperationID}, w.pushMsgAndMaxSeqCh), "")
}

func (w *Ws) doDataSynchronization(wsResp *GeneralWsResp) error {
	m := gate.SyncData{}
	err := json.Unmarshal(wsResp.Data, &m)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
		return err
	}

	log.NewInfo("doDataSynchronization begin ", m.UserID)

	operationID := utils.OperationIDGenerator()
	w.syncDataForUser(operationID, m.UserID, m.MsgType, m.MsgData)
	return nil
}

func (w *Ws) kickOnline(msg GeneralWsResp) {
	w.listener.OnKickedOffline()
}

func (w *Ws) SendSignalingReqWaitResp(req *sdk.SignalReq, operationID string) (*sdk.SignalResp, error) {
	resp, err := w.SendReqWaitResp(req, constant.WSSendSignalMsg, 10, 12, w.loginUserID, operationID)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var signalResp sdk.SignalResp
	err = proto.Unmarshal(resp.Data, &signalResp)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	return &signalResp, nil
}

func (w *Ws) SignalingWaitPush(inviterUserID, inviteeUserID, roomID string, timeout int32, operationID string) (*sdk.SignalReq, error) {
	msgIncr := inviterUserID + inviteeUserID + roomID
	log.Info(operationID, "add msgIncr: ", msgIncr)
	ch := w.AddChByIncr(msgIncr)
	defer w.DelCh(msgIncr)

	resp, err := w.WaitResp(ch, int(timeout), operationID, nil)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	var signalReq sdk.SignalReq
	err = proto.Unmarshal(resp.Data, &signalReq)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}

	return &signalReq, nil
}
