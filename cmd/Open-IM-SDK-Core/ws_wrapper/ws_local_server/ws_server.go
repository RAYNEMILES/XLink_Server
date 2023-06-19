/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/9/8 14:42).
 */
package ws_local_server

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/pkg/common/log"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/ws_wrapper/utils"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

const POINTNUM = 100

var (
	rwLock *sync.RWMutex
	WS     WServer
)

type UserConn struct {
	*websocket.Conn
	w          *sync.Mutex
	userID     string
	token      string
	platformID string
	ticker     *time.Ticker
	pingCh     chan common.Cmd2Value
}
type ChanMsg struct {
	data []byte
	uid  string
}
type WServer struct {
	wsAddr       string
	wsMaxConnNum int
	wsUpGrader   *websocket.Upgrader
	wsConnToUser map[*UserConn]map[string]string
	wsUserToConn map[string]map[string]*UserConn
	ch           chan ChanMsg
	cmdCh        chan common.Cmd2Value
}

func (ws *WServer) OnInit(wsPort int) {
	//ip := utils.ServerIP
	ws.wsAddr = ":" + utils.IntToString(wsPort)
	ws.wsMaxConnNum = 10000
	ws.wsConnToUser = make(map[*UserConn]map[string]string)
	ws.wsUserToConn = make(map[string]map[string]*UserConn)
	ws.ch = make(chan ChanMsg, 100000)
	ws.cmdCh = make(chan common.Cmd2Value, 10)
	rwLock = new(sync.RWMutex)
	ws.wsUpGrader = &websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   4096,
		CheckOrigin:      func(r *http.Request) bool { return true },
	}
}

func (ws *WServer) Run() {
	go ws.getMsgAndSend()
	http.HandleFunc("/", ws.wsHandler)         //Get request from client to handle by wsHandler
	err := http.ListenAndServe(ws.wsAddr, nil) //Start listening
	if err != nil {
		log.Info("", "Ws listening err", "", "err", err.Error())
	}
}

func (ws *WServer) getMsgAndSend() {

	for {
		select {
		case r := <-ws.ch:
			go func() {
				operationID := utils2.OperationIDGenerator()
				log.Info(operationID, "getMsgAndSend channel: ", string(r.data), r.uid)
				conns := ws.getUserConnByUserID(r.uid)
				if conns == nil {
					log.Error(operationID, "uid no conn, failed ", "uid", r.uid)
					r.data = nil

				} else {
					for _, conn := range conns {
						if conn != nil {
							log.Info(operationID, "conn", conn, "userIP", conn.RemoteAddr().String(), "userUid", r.uid, "platformID", conn.platformID, "token", conn.token)
							err := WS.writeMsg(conn, websocket.TextMessage, r.data)
							if err != nil {
								//log.Error(operationID, "wsUserToConn", ws.wsUserToConn)
								//log.Error(operationID, "WS WriteMsg error, the connection will be delete", "", "userIP", conn.RemoteAddr().String(), "userUid", r.uid, "platformID", conn.platformID, "token", conn.token, "conn", conn, "error", err.Error())
								//ws.delUserConn(conn)
							} else {
								log.Info(operationID, "writeMsg  ", conn.RemoteAddr(), string(r.data), r.uid)
							}
						} else {
							log.Error(operationID, "Conn is nil, failed")
						}
					}
				}

				r.data = nil
			}()
		}
	}

}

// jssdk
func (ws *WServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	operationID := utils2.OperationIDGenerator()
	log.Info(operationID, "wsHandler ", r.URL.Query())
	if ws.headerCheck(w, r) {
		query := r.URL.Query()
		conn, err := ws.wsUpGrader.Upgrade(w, r, nil) //Conn is obtained through the upgraded escalator
		if err != nil {
			log.Info(operationID, "upgrade http conn err", "", "err", err)
			return
		} else {
			//Connection mapping relationship,
			//userID+" "+platformID->conn
			SendID := query["sendID"][0] + " " + utils.PlatformIDToName(int32(utils.StringToInt64(query["platformID"][0])))
			log.Info(operationID, "wsHandler uid:", SendID, query["token"][0])
			newConn := &UserConn{conn, new(sync.Mutex), query["sendID"][0], query["token"][0], query["platformID"][0], time.NewTicker(3 * time.Second), make(chan common.Cmd2Value, 10)}
			ws.addUserConn(SendID, newConn, operationID)

			setPongAndCloseHandlerWS(conn, w, r, ws, query)

			go ws.RunPing(newConn)
			go ws.readMsg(newConn)
		}
	} else {
		log.Info(operationID, "headerCheck failed")
	}
}

func setPongAndCloseHandlerWS(conn *websocket.Conn, w http.ResponseWriter, r *http.Request, ws *WServer, query url.Values) {
	conn.SetPingHandler(func(pongMsg string) error {
		return nil
	})
	conn.SetCloseHandler(func(code int, text string) error {
		//log.Info("Sandman Napster", "WebSocket status : ", code)
		//if code == websocket.CloseGoingAway || code == websocket.CloseNoStatusReceived || code == websocket.CloseAbnormalClosure || code == websocket.CloseServiceRestart {
		//	conn, err := ws.wsUpGrader.Upgrade(w, r, nil) //Conn is obtained through the upgraded escalator
		//	if err == nil {
		//		SendID := query["sendID"][0] + " " + utils.PlatformIDToName(int32(utils.StringToInt64(query["platformID"][0])))
		//		newConn := &UserConn{conn, new(sync.Mutex), query["sendID"][0], query["token"][0], query["platformID"][0], time.NewTicker(3 * time.Second), make(chan common.Cmd2Value, 10)}
		//		ws.addUserConn(SendID, newConn, "operationID")
		//		setPongAndCloseHandlerWS(conn, w, r, ws, query)
		//	}
		//
		//}
		return nil
	})
}

func pMem() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("mem for test ", m)
	fmt.Println("mem for test os ", m.Sys)
	fmt.Println("mem for test HeapAlloc ", m.HeapAlloc)
}
func (ws *WServer) readMsg(conn *UserConn) {

	for {
		if conn != nil {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.NewError("", "ReadMessage error", "uid", conn.userID, "userIP", conn.RemoteAddr().String(), "error", err)
				ws.delUserConn(conn)
				break
			}
			m := Req{}
			json.Unmarshal(msg, &m)
			ws.msgParseNew(conn, msg)

		} else {
			break
		}
	}

}

func (ws *WServer) writeMsg(conn *UserConn, a int, msg []byte) error {
	if conn == nil || msg == nil {
		return errors.New("args nil")
	}
	conn.w.Lock()
	defer conn.w.Unlock()
	return conn.WriteMessage(a, msg)

}
func (ws *WServer) addUserConn(uid string, conn *UserConn, operationID string) {
	rwLock.Lock()
	defer func() {
		rwLock.Unlock()
	}()

	log.NewInfo(operationID, utils2.GetSelfFuncName(), "uid:", uid)

	if oldConnMap, ok := ws.wsUserToConn[uid]; ok {

		oldConnMap[conn.RemoteAddr().String()] = conn
		ws.wsUserToConn[uid] = oldConnMap
		log.Info(operationID, "this user is not first login", "", "uid", uid)

	} else {
		i := make(map[string]*UserConn)
		i[conn.RemoteAddr().String()] = conn
		ws.wsUserToConn[uid] = i
		log.Info(operationID, "this user is first login", "", "uid", uid)
	}
	if oldStringMap, ok := ws.wsConnToUser[conn]; ok {
		oldStringMap[conn.RemoteAddr().String()] = uid
		ws.wsConnToUser[conn] = oldStringMap
		log.Info(operationID, "find failed", "", "uid", uid, conn)

	} else {
		i := make(map[string]string)
		i[conn.RemoteAddr().String()] = uid
		ws.wsConnToUser[conn] = i
		log.Info(operationID, "this user is first login-1", "", "uid", uid, conn)
	}
	log.NewError(operationID, utils2.GetSelfFuncName(), "test user connections: after add", "uid", uid, "ws.wsUserToConn-len", len(ws.wsUserToConn[uid]), "ws.wsConnToUser-len", len(ws.wsConnToUser[conn]))

}
func (ws *WServer) getConnNum(uid string) int {
	rwLock.Lock()
	defer rwLock.Unlock()
	log.Info("", "getConnNum uid: ", uid)
	if connMap, ok := ws.wsUserToConn[uid]; ok {
		log.Info("", "uid->conn ", connMap)
		return len(connMap)
	} else {
		return 0
	}

}

func (ws *WServer) delUserConn(conn *UserConn) {

	rwLock.Lock()
	defer func() {
		rwLock.Unlock()
	}()
	if conn == nil {
		return
	}
	if conn.platformID == "" {
		return
	}
	if oldStringMap, ok := ws.wsConnToUser[conn]; ok {
		if uidPlatform, ok := oldStringMap[conn.RemoteAddr().String()]; ok {
			defer func() {
				err := common.TriggerCmdLogout(conn.pingCh)
				if err != nil {
					log.NewError("", "logout err", err.Error())
				}
			}()

			conn.Close()
			if oldConnMap, ok := ws.wsUserToConn[uidPlatform]; ok {
				log.Info("old map : ", oldConnMap, "conn: ", conn.RemoteAddr().String())
				if _, ok := oldConnMap[conn.RemoteAddr().String()]; ok {
					delete(oldConnMap, conn.RemoteAddr().String())
					ws.wsUserToConn[uidPlatform] = oldConnMap
					log.NewError("WS delete operation", "", "wsUser deleted", ws.wsUserToConn, "uid", uidPlatform, "online_num", len(ws.wsUserToConn))
					if len(oldConnMap) == 0 {
						log.Info("no conn delete user router ", uidPlatform)
						log.NewError("DelUserRouter ", uidPlatform)
						DelUserRouterNew(uidPlatform, conn.platformID, conn.RemoteAddr().String())
						//ws.wsUserToConn[uidPlatform] = make(map[string]*UserConn)
						oldConnMap = nil
						delete(ws.wsUserToConn, uidPlatform)

						userWorker := open_im_sdk.GetUserWorkerNew(conn.userID)
						if userWorker != nil {
							log.NewError("", utils2.GetSelfFuncName(), "userWorker logout", userWorker)
							ws := userWorker.Ws()
							if ws != nil {
								ws.Logout("")
							}
						}
					}
				}
			} else {
				log.Info("uid not exist", "", "wsUser deleted", ws.wsUserToConn, "uid", uidPlatform, "online_num", len(ws.wsUserToConn))
			}
			oldStringMap = nil
			delete(ws.wsConnToUser, conn)
		}

	}

}

func (ws *WServer) getUserConn(uid string) (w []*UserConn) {
	rwLock.RLock()
	defer rwLock.RUnlock()
	t := ws.wsUserToConn

	if connMap, ok := t[uid]; ok {
		for _, v := range connMap {
			w = append(w, v)
		}
		return w
	}
	return nil
}

func getKeysFromUserConnMap(m map[string]map[string]*UserConn) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (ws *WServer) getUserConnByUserID(uid string) (w []*UserConn) {
	rwLock.Lock()
	defer rwLock.Unlock()
	t := ws.wsUserToConn

	keys := getKeysFromUserConnMap(t)
	for _, key := range keys {
		if strings.Contains(key, uid+" ") {
			if connMap, ok := t[key]; ok {
				for _, conn := range connMap {
					w = append(w, conn)
				}
			}

		}
	}

	return w
}

func (ws *WServer) getUserUid(conn *UserConn) string {
	rwLock.RLock()
	defer rwLock.RUnlock()
	if oldStringMap, ok := ws.wsConnToUser[conn]; ok {
		if oldStringMap != nil {
			if uid, ok := oldStringMap[conn.RemoteAddr().String()]; ok {
				log.Info("", "getUserUid:", uid, conn.RemoteAddr())
				return uid
			}
		}
	}
	log.Error("", "getUserUid failed!", conn.RemoteAddr())
	return "getUserUid"
}

func (ws *WServer) headerCheck(w http.ResponseWriter, r *http.Request) bool {

	status := http.StatusUnauthorized
	query := r.URL.Query()
	log.Info("headerCheck: ", query["token"], query["platformID"], query["sendID"])
	if len(query["token"]) != 0 && len(query["sendID"]) != 0 && len(query["platformID"]) != 0 {
		SendID := query["sendID"][0] + " " + utils.PlatformIDToName(int32(utils.StringToInt64(query["platformID"][0])))
		if ws.getConnNum(SendID) >= POINTNUM {
			log.Info("Over quantity failed", query, ws.getConnNum(SendID), SendID)
			w.Header().Set("Sec-Websocket-Version", "13")
			http.Error(w, "Over quantity", status)
			return false
		}
		tmpPlatformID, _ := strconv.ParseInt(query["platformID"][0], 10, 64)
		checkFlag := open_im_sdk.CheckToken(query["sendID"][0], query["token"][0], int32(tmpPlatformID))
		if checkFlag != nil {
			log.Info("check token failed", query["sendID"][0], query["token"][0], checkFlag.Error())
			w.Header().Set("Sec-Websocket-Version", "13")
			http.Error(w, http.StatusText(status), status)
			return false
		}
		log.Info("Connection Authentication Success", "", "token", query["token"][0], "userID", query["sendID"][0], "platformID", query["platformID"][0])
		return true

	} else {
		log.Info("Args err", "", "query", query)
		w.Header().Set("Sec-Websocket-Version", "13")
		http.Error(w, http.StatusText(status), StatusBadRequest)
		return false
	}
}

func (ws *WServer) RunPing(u *UserConn) {
	for {
		select {
		case r := <-u.pingCh:
			if r.Cmd == constant.CmdLogout {
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
		log.NewError("", utils2.GetSelfFuncName(), "conn == nil")
		return utils2.Wrap(errors.New("conn == nil"), "")
	}
	ping := "try ping"
	err := u.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		log.NewError("", utils2.GetSelfFuncName(), "SetWriteTimeout", err.Error())
		return utils2.Wrap(err, "SetWriteDeadline failed")
	}
	err = u.WriteMessage(websocket.PingMessage, []byte(ping))

	if err != nil {
		log.NewError("", utils2.GetSelfFuncName(), "WriteMessage", u.userID, err.Error())

		if u.ticker != nil {
			u.ticker.Stop()
		}
		ws.delUserConn(u)

		return utils2.Wrap(err, "WriteMessage failed")
	}
	return nil
}
