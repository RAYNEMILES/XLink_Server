package gate

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	constant2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/proto/local_database"
	"Open_IM/pkg/utils"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type UserConn struct {
	*websocket.Conn
	w            *sync.Mutex
	PushedMaxSeq uint32
	uid          string
	platformID   int
	isClose      bool
	ticker       *time.Ticker
	pingCh       chan common.Cmd2Value
	errTimes     int
	pingErrTimes int
}

type WServer struct {
	wsAddr       string
	wsMaxConnNum int
	wsUpGrader   *websocket.Upgrader
	//wsConnToUser map[*UserConn]map[int]string
	//wsUserToConn map[string]map[int]*UserConn
	wsUserToConnNew map[string]map[string]*UserConn
}

type GeneralWsResp struct {
	ReqIdentifier int    `json:"reqIdentifier"`
	ErrCode       int    `json:"errCode"`
	ErrMsg        string `json:"errMsg"`
	MsgIncr       string `json:"msgIncr"`
	OperationID   string `json:"operationID"`
	Data          []byte `json:"data"`
}

type SyncData struct {
	MsgType string `json:"msgType"`
	UserID  string `json:"userID"`
	MsgData string `json:"msgData"`
}

var (
	myWs *WServer
	sh   SyncDataConsumerHandler
)

func (ws *WServer) onInit(wsPort int) {
	ws.wsAddr = ":" + utils.IntToString(wsPort)
	ws.wsMaxConnNum = config.Config.LongConnSvr.WebsocketMaxConnNum
	//ws.wsConnToUser = make(map[*UserConn]map[int]string)
	//ws.wsUserToConn = make(map[string]map[int]*UserConn)
	ws.wsUserToConnNew = make(map[string]map[string]*UserConn)
	ws.wsUpGrader = &websocket.Upgrader{
		HandshakeTimeout: time.Duration(config.Config.LongConnSvr.WebsocketTimeOut) * time.Second,
		ReadBufferSize:   config.Config.LongConnSvr.WebsocketMaxMsgLen,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}

	myWs = ws
	sh.Init(msgHandler)
}

func msgHandler(msg *sarama.ConsumerMessage) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		log.NewError("", utils.GetSelfFuncName(), " panic is ", r)
	//		buf := make([]byte, 1<<20)
	//		runtime.Stack(buf, true)
	//		log.NewError("", utils.GetSelfFuncName(), "panic", "call", string(buf))
	//	}
	//}()

	memberID := string(msg.Key)

	log.NewInfo("", utils.GetSelfFuncName(), "ws_server handle msg", memberID, string(msg.Value))
	syncDataMsg := local_database.SyncDataMsg{}
	err := proto.Unmarshal(msg.Value, &syncDataMsg)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "proto.Unmarshal failed")
		return
	}

	msgData := SyncData{
		MsgType: syncDataMsg.MsgType,
		UserID:  syncDataMsg.UserID,
		MsgData: syncDataMsg.MsgData,
	}
	msgDataByte, err := json.Marshal(msgData)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "json.Marshal failed")
		return
	}
	//wsUserToConn := myWs.GetWsUserToConnNew(memberID)
	if userToConn, ok := myWs.GetWsUserToConnNew(memberID); ok {
		for _, conn := range userToConn {
			genWsResp := GeneralWsResp{
				ReqIdentifier: constant.WsSyncDataMsg,
				Data:          msgDataByte,
			}
			msgValue := utils.StructToJsonBytes(genWsResp)
			err := myWs.SetWriteTimeoutWriteMsg(conn, 1, msgValue, 10)
			if err == nil {
				//log.NewError("", utils.GetSelfFuncName(), "send sync group message", string(msgValue), conn.RemoteAddr(), memberID)
			} else {
				log.Debug("", utils.GetSelfFuncName(), "SetWriteTimeoutWriteMsg is failed", conn, string(msgValue))
			}
		}
	} else {
		//log.Error("", utils.GetSelfFuncName(), "wsUserToConn is nil", memberID)
	}
}

func (ws *WServer) run() {
	//defer func() {
	//	if r := recover(); r != nil {
	//		log.NewError("", utils.GetSelfFuncName(), " panic is ", r)
	//		buf := make([]byte, 1<<20)
	//		runtime.Stack(buf, true)
	//		log.NewError("", utils.GetSelfFuncName(), "panic", "call", string(buf))
	//	}
	//}()

	go sh.syncGroupConsumerHandler.RegisterHandleAndConsumer(&sh)

	http.HandleFunc("/", ws.wsHandler)         //Get request from client to handle by wsHandler
	err := http.ListenAndServe(ws.wsAddr, nil) //Start listening
	if err != nil {
		panic("Ws listening err:" + err.Error())
	}
}

// msg_gateway websocket
func (ws *WServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	operationID := ""
	if len(query["operationID"]) != 0 {
		operationID = query["operationID"][0]
	} else {
		operationID = utils.OperationIDGenerator()
	}
	log.Debug(operationID, utils.GetSelfFuncName(), " args: ", query)
	if ws.headerCheck(w, r, operationID) {
		conn, err := ws.wsUpGrader.Upgrade(w, r, nil) //Conn is obtained through the upgraded escalator
		if err != nil {
			log.Error(operationID, "upgrade http conn err", err.Error(), query)
			return
		} else {
			newConn := &UserConn{conn, new(sync.Mutex), 0, query["sendID"][0], utils.StringToInt(query["platformID"][0]), false, time.NewTicker(3 * time.Second), make(chan common.Cmd2Value, 10), 0, 0}
			conn.SetCloseHandler(func(code int, text string) error {
				log.Debug("+++++", utils.GetSelfFuncName(), "SetCloseHandler", code, text)
				newConn.isClose = true
				//conn.Close()
				return nil
			})
			userCount++
			ws.addUserConn(query["sendID"][0], utils.StringToInt(query["platformID"][0]), newConn, query["token"][0], operationID)
			go ws.RunPing(newConn)
			go ws.readMsg(newConn)
		}
	} else {
		log.Error(operationID, "headerCheck failed ")
	}
}

func (ws *WServer) readMsg(conn *UserConn) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		log.NewError("", utils.GetSelfFuncName(), " panic is ", r)
	//		buf := make([]byte, 1<<20)
	//		runtime.Stack(buf, true)
	//		log.NewError("", utils.GetSelfFuncName(), "panic", "call", string(buf))
	//	}
	//}()
	//for {
	//	//if conn.isClose {
	//	//	log.NewError("", utils.GetSelfFuncName(), "conn.isClose!!!")
	//	//	userCount--
	//	//	ws.delUserConn(conn)
	//	//	return
	//	//}
	//
	//	//log.NewError("", utils.GetSelfFuncName(), conn.RemoteAddr().String(), conn.uid)
	//	_, msg, err := conn.ReadMessage()
	//	//if messageType == websocket.PingMessage {
	//	//	log.NewInfo("", "this is a  pingMessage")
	//	//}
	//	//
	//	//if err != nil {
	//	//	log.Error("", "WS ReadMsg error ", " userIP", conn.RemoteAddr().String(), "error", err.Error())
	//	//	//|| strings.Contains(err.Error(), "close 1006 (abnormal closure)")
	//	//	if strings.Contains(err.Error(), "close 1001 (going away)") {
	//	//		log.Error("", utils.GetSelfFuncName(), "IsUnexpectedCloseError")
	//	//		userCount--
	//	//		ws.delUserConn(conn)
	//	//		//runtime.Goexit()
	//	//	}
	//	//	return
	//	//}
	//
	//	//err = conn.WriteMessage(1, utils.String2Bytes("ping"))
	//	//if err != nil {
	//	//	log.NewError("", utils.GetSelfFuncName(), "WriteMessage err=======", err.Error())
	//	//}
	//
	//	if err == nil {
	//		ws.msgParse(conn, msg)
	//	}
	//
	//}

	for {
		if conn != nil {
			_, msg, err := conn.ReadMessage()
			if err == nil {
				//log.NewError("", utils.GetSelfFuncName(), "read message success:", utils.Bytes2String(msg), conn.uid)
				ws.msgParse(conn, msg)
			} else {
				log.Debug("", utils.GetSelfFuncName(), "read message error, delete connection!!!", conn.uid, err.Error())
				ws.delUserConn(conn)
				break
			}
		} else {
			log.NewError("", utils.GetSelfFuncName(), "read message error, conn is nil")
			break
		}
	}
}

func (ws *WServer) SetWriteTimeout(conn *UserConn, timeout int) {
	conn.w.Lock()
	defer conn.w.Unlock()
	conn.SetWriteDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
}

func (ws *WServer) writeMsg(conn *UserConn, a int, msg []byte) error {
	conn.w.Lock()
	defer conn.w.Unlock()
	conn.SetWriteDeadline(time.Now().Add(time.Duration(60) * time.Second))
	return conn.WriteMessage(a, msg)
}

func (ws *WServer) SetWriteTimeoutWriteMsg(conn *UserConn, a int, msg []byte, timeout int) error {
	conn.w.Lock()
	defer conn.w.Unlock()
	conn.SetWriteDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	return conn.WriteMessage(a, msg)
}

//func (ws *WServer) MultiTerminalLoginChecker(uid string, platformID int, newConn *UserConn, token string, operationID string) {
//	switch config.Config.MultiLoginPolicy {
//	case constant.AllLoginButSameTermKick:
//		if oldConnMap, ok := ws.wsUserToConn[uid]; ok { // user->map[platform->conn]
//			if oldConn, ok := oldConnMap[platformID]; ok {
//				log.NewDebug(operationID, uid, platformID, "kick old conn")
//				//	ws.sendKickMsg(oldConn, newConn)
//				m, err := db.DB.GetTokenMapByUidPid(uid, constant.PlatformIDToName(platformID))
//				if err != nil && err != go_redis.Nil {
//					log.NewError(operationID, "get token from redis err", err.Error(), uid, constant.PlatformIDToName(platformID))
//					return
//				}
//				if m == nil {
//					log.NewError(operationID, "get token from redis err", "m is nil", uid, constant.PlatformIDToName(platformID))
//					return
//				}
//				log.NewDebug(operationID, "get token map is ", m, uid, constant.PlatformIDToName(platformID))
//
//				for k, _ := range m {
//					if k != token {
//						m[k] = constant.KickedToken
//					}
//				}
//				log.NewDebug(operationID, "set token map is ", m, uid, constant.PlatformIDToName(platformID))
//				err = db.DB.SetTokenMapByUidPid(uid, platformID, m)
//				if err != nil {
//					log.NewError(operationID, "SetTokenMapByUidPid err", err.Error(), uid, platformID, m)
//					return
//				}
//				err = oldConn.Close()
//				delete(oldConnMap, platformID)
//				ws.wsUserToConn[uid] = oldConnMap
//				if len(oldConnMap) == 0 {
//					delete(ws.wsUserToConn, uid)
//				}
//				//delete(ws.wsConnToUser, oldConn)
//				if err != nil {
//					log.NewError(operationID, "conn close err", err.Error(), uid, platformID)
//				}
//			} else {
//				log.NewWarn(operationID, "abnormal uid-conn  ", uid, platformID, oldConnMap[platformID])
//			}
//
//		} else {
//			log.NewDebug(operationID, "no other conn", ws.wsUserToConn, uid, platformID)
//		}
//
//	case constant.SingleTerminalLogin:
//	case constant.WebAndOther:
//	}
//}

func (ws *WServer) sendKickMsg(oldConn, newConn *UserConn) {
	mReply := Resp{
		ReqIdentifier: constant.WSKickOnlineMsg,
		ErrCode:       constant.ErrTokenInvalid.ErrCode,
		ErrMsg:        constant.ErrTokenInvalid.ErrMsg,
	}
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(mReply)
	if err != nil {
		log.NewError(mReply.OperationID, mReply.ReqIdentifier, mReply.ErrCode, mReply.ErrMsg, "Encode Msg error", oldConn.RemoteAddr().String(), newConn.RemoteAddr().String(), err.Error())
		return
	}
	err = ws.writeMsg(oldConn, websocket.BinaryMessage, b.Bytes())
	if err != nil {
		log.NewError(mReply.OperationID, mReply.ReqIdentifier, mReply.ErrCode, mReply.ErrMsg, "sendKickMsg WS WriteMsg error", oldConn.RemoteAddr().String(), newConn.RemoteAddr().String(), err.Error())
	}
}

func (ws *WServer) addUserConn(uid string, platformID int, conn *UserConn, token string, operationID string) {
	log.Debug("", utils.GetSelfFuncName(), "======", conn.RemoteAddr())
	rwLock.Lock()
	defer func() {
		rwLock.Unlock()
		//if r := recover(); r != nil {
		//	log.NewError(operationID, utils.GetSelfFuncName(), " panic is ", r)
		//	buf := make([]byte, 1<<20)
		//	runtime.Stack(buf, true)
		//	log.NewError(operationID, utils.GetSelfFuncName(), "panic", "call", string(buf))
		//}
	}()

	log.NewInfo(operationID, utils.GetSelfFuncName(), "addUserConn args: ", uid, platformID, conn, token)
	callbackResp := callbackUserOnline(operationID, uid, platformID, token)
	if callbackResp.ErrCode != 0 {
		log.NewError(operationID, utils.GetSelfFuncName(), "callbackUserOnline resp:", callbackResp)
	}
	//remove old connection
	//ws.MultiTerminalLoginChecker(uid, platformID, conn, token, operationID)
	if oldConnMap, ok := ws.wsUserToConnNew[uid]; ok {
		oldConnMap[conn.RemoteAddr().String()] = conn
		ws.wsUserToConnNew[uid] = oldConnMap
		log.Debug(operationID, "user not first come in, add conn ", uid, platformID, conn, oldConnMap)
	} else {
		i := make(map[string]*UserConn)
		i[conn.RemoteAddr().String()] = conn
		ws.wsUserToConnNew[uid] = i
		log.Debug(operationID, "user first come in, new user, conn", uid, platformID, conn, ws.wsUserToConnNew[uid])
	}
	//if oldStringMap, ok := ws.wsConnToUser[conn]; ok {
	//	oldStringMap[platformID] = uid
	//	ws.wsConnToUser[conn] = oldStringMap
	//} else {
	//	i := make(map[int]string)
	//	i[platformID] = uid
	//	ws.wsConnToUser[conn] = i
	//}
	count := 0
	for _, v := range ws.wsUserToConnNew {
		count = count + len(v)
	}
	log.Debug(operationID, utils.GetSelfFuncName(), "msg_gateway", "uid", uid, "connection_platform", constant.PlatformIDToName(platformID), "online_user_num", len(ws.wsUserToConnNew), "ws.wsUserToConnNew-len1", len(ws.wsUserToConnNew[uid]), "online_conn_num", count)
}

func (ws *WServer) delUserConn(conn *UserConn) {
	log.Debug("", utils.GetSelfFuncName(), "======", conn.RemoteAddr())
	rwLock.Lock()
	defer func() {
		rwLock.Unlock()
		//if r := recover(); r != nil {
		//	log.NewError("", utils.GetSelfFuncName(), " panic is ", r)
		//	buf := make([]byte, 1<<20)
		//	runtime.Stack(buf, true)
		//	log.NewError("", utils.GetSelfFuncName(), "panic", "call", string(buf))
		//}
	}()
	operationID := utils.OperationIDGenerator()
	//var uid string
	//var platform int
	//if oldStringMap, ok := ws.wsConnToUser[conn]; ok {
	//	for k, v := range oldStringMap {
	//		platform = k
	//		uid = v
	//	}
	uid := conn.uid
	platform := conn.platformID

	if oldConnMap, ok := ws.wsUserToConnNew[uid]; ok {
		if _, ok := oldConnMap[conn.RemoteAddr().String()]; ok {
			defer func() {
				err := common.TriggerCmdLogout(conn.pingCh)
				if err != nil {
					log.NewError("", "logout err", err.Error())
				}

				//err = conn.Close()
				//if err != nil {
				//	log.NewError("close err", "conn", conn)
				//}
				//
				//err = conn.Conn.Close()
				//if err != nil {
				//	log.NewError("close err", "conn", conn, "err", err.Error())
				//}

				count := 0
				for _, v := range ws.wsUserToConnNew {
					count = count + len(v)
				}
				log.Debug(operationID, "WS delete operation", "", "wsUser deleted", ws.wsUserToConnNew, "disconnection_uid", uid, "disconnection_platform", platform, "online_user_num", len(ws.wsUserToConnNew), "online_conn_num", count)
			}()

			conn.Close()
			delete(oldConnMap, conn.RemoteAddr().String())
			ws.wsUserToConnNew[uid] = oldConnMap
			if len(oldConnMap) == 0 {
				oldConnMap = nil
				delete(ws.wsUserToConnNew, uid)
			}
		}
	} else {
		log.Debug(operationID, "WS delete operation", "", "wsUser deleted", ws.wsUserToConnNew, "disconnection_uid", uid, "disconnection_platform", platform, "online_user_num", len(ws.wsUserToConnNew))
	}
	//	delete(ws.wsConnToUser, conn)
	//
	//}

	//err := conn.Close()
	//if err != nil {
	//	log.Error(operationID, " close err", "", "uid", uid, "platform", platform, err.Error())
	//}
	callbackResp := callbackUserOffline(operationID, uid, platform)
	if callbackResp.ErrCode != 0 {
		log.NewError(operationID, utils.GetSelfFuncName(), "callbackUserOffline failed", callbackResp)
	}

	//log.NewError(operationID, utils.GetSelfFuncName(), "msg_gateway", "ws.wsUserToConn-len", len(ws.wsUserToConn[uid]), "ws.wsConnToUser-len", len(ws.wsConnToUser[conn]), "uid", uid, "connection_platform", constant.PlatformIDToName(platform), "online_user_num", len(ws.wsUserToConn), "online_conn_num", len(ws.wsConnToUser), "ws.wsConnToUser", ws.wsConnToUser)
	//log.Debug(operationID, utils.GetSelfFuncName(), "msg_gateway", "ws.wsUserToConn-len", len(ws.wsUserToConnNew[uid]), "uid", uid, "connection_platform", constant.PlatformIDToName(platform), "online_user_num", len(ws.wsUserToConnNew))
}

func (ws *WServer) getUserConnNew(uid string, platform int) []*UserConn {
	rwLock.RLock()
	defer rwLock.RUnlock()
	list := []*UserConn{}
	if connMap, ok := ws.wsUserToConnNew[uid]; ok {
		for _, conn := range connMap {
			if conn.platformID == platform {
				list = append(list, conn)
			}
		}
	}
	return list
}

func (ws *WServer) getUserAllConsNew(uid string) map[string]*UserConn {
	rwLock.RLock()
	defer rwLock.RUnlock()
	if connMap, ok := ws.wsUserToConnNew[uid]; ok {
		tmpMap := make(map[string]*UserConn, len(connMap))
		for s, conn := range connMap {
			tmpMap[s] = conn
		}
		return tmpMap
	}
	return nil
}

//func (ws *WServer) getUserConn(uid string, platform int) *UserConn {
//	rwLock.RLock()
//	defer rwLock.RUnlock()
//	if connMap, ok := ws.wsUserToConnNew[uid]; ok {
//		if conn, flag := connMap[platform]; flag {
//			return conn
//		}
//	}
//	return nil
//}
//
//func (ws *WServer) getUserAllCons(uid string) map[int]*UserConn {
//	rwLock.RLock()
//	defer rwLock.RUnlock()
//	if connMap, ok := ws.wsUserToConn[uid]; ok {
//		return connMap
//	}
//	return nil
//}

//	func (ws *WServer) getUserUid(conn *UserConn) (uid string, platform int) {
//		rwLock.RLock()
//		defer rwLock.RUnlock()
//
//		if stringMap, ok := ws.wsConnToUser[conn]; ok {
//			for k, v := range stringMap {
//				platform = k
//				uid = v
//			}
//			return uid, platform
//		}
//		return "", 0
//	}
func (ws *WServer) headerCheck(w http.ResponseWriter, r *http.Request, operationID string) bool {
	status := http.StatusUnauthorized
	query := r.URL.Query()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "request url:", r.RemoteAddr, r.Header, r.URL.String())
	if len(query["token"]) != 0 && len(query["sendID"]) != 0 && len(query["platformID"]) != 0 {
		//if ok, err, msg := token_verify.WsVerifyToken(query["token"][0], query["sendID"][0], query["platformID"][0], operationID); !ok {
		//	log.Error(operationID, "Token verify failed ", "query ", query, msg, err.Error())
		//	w.Header().Set("Sec-Websocket-Version", "13")
		//	w.Header().Set("ws_err_msg", err.Error())
		//	http.Error(w, err.Error(), status)
		//	return false
		//}
		//else {
		//	//check permissions
		//	err2 := utils2.CheckUserPermissions(query["sendID"][0])
		//	if err2 != nil {
		//		http.Error(w, err2.Error(), status)
		//		return false
		//	}
		//
		//	log.Info(operationID, "Connection Authentication Success", "", "token ", query["token"][0], "userID ", query["sendID"][0], "platformID ", query["platformID"][0])
		//	return true
		//}

		log.Info(operationID, "Connection Authentication Success", "", "token ", query["token"][0], "userID ", query["sendID"][0], "platformID ", query["platformID"][0])
		return true

	} else {
		log.Error(operationID, "Args err ", "query ", query)
		w.Header().Set("Sec-Websocket-Version", "13")
		errMsg := "args err, need token, sendID, platformID"
		w.Header().Set("ws_err_msg", errMsg)
		http.Error(w, errMsg, status)
		return false
	}
}

//func (ws *WServer) GetWsUserToConn() map[string]map[int]*UserConn {
//	rwLock.Lock()
//	defer rwLock.Unlock()
//	return ws.wsUserToConn
//}

func (ws *WServer) GetWsUserToConn() map[string]map[string]*UserConn {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return ws.wsUserToConnNew
}

func (ws *WServer) GetWsUserToConnNew(memberID string) (map[string]*UserConn, bool) {
	rwLock.Lock()
	defer rwLock.Unlock()
	tmpUserConn := map[string]*UserConn{}
	userConn, ok := ws.wsUserToConnNew[memberID]
	if ok {
		for s, conn := range userConn {
			tmpUserConn[s] = conn
		}
	}

	return tmpUserConn, ok
}

func (ws *WServer) RunPing(u *UserConn) {
	//defer func() {
	//	if r := recover(); r != nil {
	//		log.NewError("", "RunPing panic", " panic is ", r)
	//		buf := make([]byte, 1<<20)
	//		runtime.Stack(buf, true)
	//		log.NewError("", "panic", "call", string(buf))
	//		ws.RunPing(u)
	//		log.Info("", "goroutine RunPing restart")
	//	}
	//}()
	for {
		select {
		case r := <-u.pingCh:
			if r.Cmd == constant2.CmdLogout {
				log.NewError("", "recv RunPing logout cmd, close conn,  set logout state, Goexit...")
				if u.ticker != nil {
					u.ticker.Stop()
				}
				//runtime.Goexit()
				return
			}
		case <-u.ticker.C:
			ws.SendPingMsg2(u)
		}
	}
	if u.ticker != nil {
		u.ticker.Stop()
	}
}

func (ws *WServer) SendPingMsg2(u *UserConn) error {
	u.w.Lock()
	defer u.w.Unlock()
	if u == nil {
		log.Debug("", utils2.GetSelfFuncName(), "conn == nil")
		return utils2.Wrap(errors.New("conn == nil"), "")
	}
	ping := "msg_gateway ping"
	err := u.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.NewError("", utils2.GetSelfFuncName(), "SetWriteTimeout", err.Error())
		return utils2.Wrap(err, "SetWriteDeadline failed")
	}
	err = u.WriteMessage(websocket.PingMessage, []byte(ping))
	if err != nil {
		log.NewError("", utils2.GetSelfFuncName(), "WriteMessage", u.uid, err.Error())
		//u.ticker.Stop()
		//if strings.Contains(err.Error(), "i/o timeout") {
		//	ws.delUserConn(u)
		//	u.ticker.Stop()
		//}
		//u.pingErrTimes++
		//if u.pingErrTimes > 10 {
		//	ws.delUserConn(u)
		//	u.ticker.Stop()
		//}

		//if strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "i/o timeout") {
		//	ws.delUserConn(u)
		//	u.ticker.Stop()
		//}

		u.ticker.Stop()
		ws.delUserConn(u)

		//common.TriggerCmdLogout(ws.cmdCh)
		return utils2.Wrap(err, "WriteMessage failed")
	}
	//log.Debug("", utils2.GetSelfFuncName(), "SendPingMsg2 success!", u.uid, u.RemoteAddr().String())
	return nil
}
