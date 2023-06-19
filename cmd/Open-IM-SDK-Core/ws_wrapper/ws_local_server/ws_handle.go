package ws_local_server

import (
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/ws_wrapper/utils"
	"Open_IM/pkg/common/log"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type EventData struct {
	Event       string `json:"event"`
	ErrCode     int32  `json:"errCode"`
	ErrMsg      string `json:"errMsg"`
	Data        string `json:"data"`
	OperationID string `json:"operationID"`
}

type BaseSuccessFailed struct {
	funcName    string //e.g open_im_sdk/open_im_sdk.Login
	operationID string
	uid         string
}

// e.g open_im_sdk/open_im_sdk.Login ->Login
func cleanUpfuncName(funcName string) string {
	end := strings.LastIndex(funcName, ".")
	if end == -1 {
		log.Info("", "funcName not include.", funcName)
		return ""
	}
	return funcName[end+1:]
}

func (b *BaseSuccessFailed) OnError(errCode int32, errMsg string) {
	log.Info("", "!!!!!!!OnError ", b.uid, b.operationID, b.funcName)
	SendOneUserMessage(EventData{cleanUpfuncName(b.funcName), errCode, errMsg, "", b.operationID}, b.uid)
}

func (b *BaseSuccessFailed) OnSuccess(data string) {
	log.Info("", "!!!!!!!OnSuccess ", b.uid, b.operationID, b.funcName)
	SendOneUserMessage(EventData{cleanUpfuncName(b.funcName), 0, "", data, b.operationID}, b.uid)
}

func runFuncName() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func int32ToString(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}

//uid->funcname->func

type WsFuncRouter struct {
	uId      string
	connAddr string
	//conn *UserConn
}

//func DelUserRouter(uid, platformID, connAddr string) {
//	log.Info("", "DelUserRouter ", uid)
//	platformIDInt, _ := strconv.ParseInt(platformID, 10, 64)
//	sub := " " + utils.PlatformIDToName(int32(platformIDInt))
//	idx := strings.LastIndex(uid, sub)
//	if idx == -1 {
//		log.NewError("", "err uid, not Web", uid, sub)
//		return
//	}
//
//	uid = uid[:idx]
//
//	UserRouteRwLock.Lock()
//	defer UserRouteRwLock.Unlock()
//	urm, ok := UserRouteMap[uid]
//	operationID := utils2.OperationIDGenerator()
//	if ok {
//
//		if router, ok := urm[connAddr]; ok {
//			log.Info(operationID, "DelUserRouter logout, UnInitSDK ", uid, operationID)
//
//			router.wsRouter.LogoutNoCallback(uid, operationID)
//			router.refName = make(map[string]reflect.Value)
//
//			delete(urm, connAddr)
//			if len(urm) != 0 {
//				UserRouteMap[uid] = urm
//				log.NewError(operationID, utils2.GetSelfFuncName(), "urm length", len(urm), uid, connAddr)
//
//			} else {
//				router.wsRouter.UnInitSDKNew()
//				delete(UserRouteMap, uid)
//			}
//
//			log.NewError(operationID, utils2.GetSelfFuncName(), "UserRouteMap length", len(UserRouteMap), uid, connAddr)
//
//		} else {
//			log.Info(operationID, "no found UserRouteMap for connADDr: ", connAddr)
//		}
//
//	} else {
//		log.Info(operationID, "no found UserRouteMap for user: ", uid)
//	}
//	log.Info(operationID, "DelUserRouter delete ", uid, connAddr)
//
//}

func DelUserRouterNew(uid, platformID, connAddr string) {
	log.Info("", "DelUserRouter ", uid)
	platformIDInt, _ := strconv.ParseInt(platformID, 10, 64)
	sub := " " + utils.PlatformIDToName(int32(platformIDInt))
	idx := strings.LastIndex(uid, sub)
	if idx == -1 {
		log.NewError("", "err uid, not Web", uid, sub)
		return
	}

	uid = uid[:idx]

	UserRouteRwLock.Lock()
	defer UserRouteRwLock.Unlock()
	_, ok := UserRouteMapNew[uid]
	operationID := utils2.OperationIDGenerator()
	if ok {
		delete(UserRouteMapNew, uid)
		log.Debug(operationID, utils2.GetSelfFuncName(), "UserRouteMap length", len(UserRouteMapNew), uid, connAddr)

	} else {
		log.Info(operationID, "no found UserRouteMap for user: ", uid)
	}
	log.Info(operationID, "DelUserRouter delete ", uid, connAddr)

}

//func GenUserRouterNoLock(uid, connAddr string, batchMsg int, operationID string) *RefRouter {
//	urm, ok := UserRouteMap[uid]
//	if urm != nil && ok {
//		if _, ok := urm[connAddr]; ok {
//			if ok {
//				return nil
//			}
//		}
//	}
//
//	RouteMap1 := make(map[string]reflect.Value, 0)
//	var wsRouter1 WsFuncRouter
//	wsRouter1.uId = uid
//	wsRouter1.connAddr = connAddr
//
//	vf := reflect.ValueOf(&wsRouter1)
//	vft := vf.Type()
//
//	mNum := vf.NumMethod()
//	for i := 0; i < mNum; i++ {
//		mName := vft.Method(i).Name
//		log.Info(operationID, "index:", i, " MethodName:", mName)
//		RouteMap1[mName] = vf.Method(i)
//	}
//	wsRouter1.InitSDKNew(ConfigSvr, operationID)
//	log.Info(operationID, "SetAdvancedMsgListener() ", uid)
//	wsRouter1.SetAdvancedMsgListener()
//	if batchMsg == 1 {
//		log.Info(operationID, "SetBatchMsgListener() ", uid)
//		wsRouter1.SetBatchMsgListener()
//	}
//	wsRouter1.SetConversationListener()
//	log.Info(operationID, "SetFriendListener() ", uid)
//	wsRouter1.SetFriendListener()
//	log.Info(operationID, "SetGroupListener() ", uid)
//	wsRouter1.SetGroupListener()
//	log.Info(operationID, "SetUserListener() ", uid)
//	wsRouter1.SetUserListener()
//	log.Info(operationID, "SetSignalingListener() ", uid)
//	wsRouter1.SetSignalingListener()
//	log.Info(operationID, "setWorkMomentsListener", uid)
//	wsRouter1.SetWorkMomentsListener()
//	var rr RefRouter
//	rr.refName = RouteMap1
//	rr.wsRouter = &wsRouter1
//
//	var routerMap map[string]RefRouter
//	if routerMap, ok = UserRouteMap[uid]; ok {
//		routerMap[connAddr] = rr
//	} else {
//		routerMap = map[string]RefRouter{}
//		routerMap[connAddr] = rr
//	}
//	UserRouteMap[uid] = routerMap
//	log.Info("", "insert UserRouteMap: ", uid)
//	return &rr
//}

//func GenUserRouterWithLock(uid, connAddr string, batchMsg int, operationID string) *RefRouter {
//	UserRouteRwLock.RLock()
//	defer UserRouteRwLock.RUnlock()
//
//	if uid == "" {
//		if _, ok := UserRouteMap[uid]; ok {
//			log.NewError(operationID, utils2.GetSelfFuncName(), "delete uid none router!!")
//			delete(UserRouteMap, uid)
//		}
//		return nil
//	}
//
//	urm, ok := UserRouteMap[uid]
//	if urm != nil && ok {
//		if _, ok := urm[connAddr]; ok {
//			if ok {
//				return nil
//			}
//		}
//	}
//
//	RouteMap1 := make(map[string]reflect.Value, 0)
//	var wsRouter1 WsFuncRouter
//	wsRouter1.uId = uid
//	wsRouter1.connAddr = connAddr
//
//	vf := reflect.ValueOf(&wsRouter1)
//	vft := vf.Type()
//
//	mNum := vf.NumMethod()
//	for i := 0; i < mNum; i++ {
//		mName := vft.Method(i).Name
//		log.Info(operationID, "index:", i, " MethodName:", mName)
//		RouteMap1[mName] = vf.Method(i)
//	}
//	wsRouter1.InitSDKNew(ConfigSvr, operationID)
//	log.Info(operationID, "SetAdvancedMsgListener() ", uid)
//	wsRouter1.SetAdvancedMsgListener()
//	if batchMsg == 1 {
//		log.Info(operationID, "SetBatchMsgListener() ", uid)
//		wsRouter1.SetBatchMsgListener()
//	}
//	wsRouter1.SetConversationListener()
//	log.Info(operationID, "SetFriendListener() ", uid)
//	wsRouter1.SetFriendListener()
//	log.Info(operationID, "SetGroupListener() ", uid)
//	wsRouter1.SetGroupListener()
//	log.Info(operationID, "SetUserListener() ", uid)
//	wsRouter1.SetUserListener()
//	log.Info(operationID, "SetSignalingListener() ", uid)
//	wsRouter1.SetSignalingListener()
//	log.Info(operationID, "setWorkMomentsListener", uid)
//	wsRouter1.SetWorkMomentsListener()
//	var rr RefRouter
//	rr.refName = RouteMap1
//	rr.wsRouter = &wsRouter1
//
//	var routerMap map[string]RefRouter
//	if routerMap, ok = UserRouteMap[uid]; ok {
//		routerMap[connAddr] = rr
//	} else {
//		routerMap = map[string]RefRouter{}
//		routerMap[connAddr] = rr
//	}
//	UserRouteMap[uid] = routerMap
//	log.Info("", "insert UserRouteMap: ", uid)
//
//	log.NewError(operationID, utils2.GetSelfFuncName(), "routerMap length", len(routerMap), uid, connAddr)
//	log.NewError(operationID, utils2.GetSelfFuncName(), "UserRouteMap length", len(UserRouteMap), uid, connAddr)
//	return &rr
//}

func GenUserRouterWithLockNew(uid, connAddr string, batchMsg int, operationID string) *RefRouter {
	UserRouteRwLock.Lock()
	defer UserRouteRwLock.Unlock()

	if uid == "" {
		if _, ok := UserRouteMapNew[uid]; ok {
			log.Debug(operationID, utils2.GetSelfFuncName(), "delete uid none router!!")
			delete(UserRouteMapNew, uid)
		}
		return nil
	}

	if urm, ok := UserRouteMapNew[uid]; ok {
		return &urm
	}

	RouteMap1 := make(map[string]reflect.Value, 0)
	var wsRouter1 WsFuncRouter
	wsRouter1.uId = uid
	wsRouter1.connAddr = connAddr

	vf := reflect.ValueOf(&wsRouter1)
	vft := vf.Type()

	mNum := vf.NumMethod()
	for i := 0; i < mNum; i++ {
		mName := vft.Method(i).Name
		log.Info(operationID, "index:", i, " MethodName:", mName)
		RouteMap1[mName] = vf.Method(i)
	}
	wsRouter1.InitSDKNew(ConfigSvr, operationID)
	log.Info(operationID, "SetAdvancedMsgListener() ", uid)
	wsRouter1.SetAdvancedMsgListener()
	if batchMsg == 1 {
		log.Info(operationID, "SetBatchMsgListener() ", uid)
		wsRouter1.SetBatchMsgListener()
	}
	wsRouter1.SetConversationListener()
	log.Info(operationID, "SetFriendListener() ", uid)
	wsRouter1.SetFriendListener()
	log.Info(operationID, "SetGroupListener() ", uid)
	wsRouter1.SetGroupListener()
	log.Info(operationID, "SetUserListener() ", uid)
	wsRouter1.SetUserListener()
	log.Info(operationID, "SetSignalingListener() ", uid)
	wsRouter1.SetSignalingListener()
	log.Info(operationID, "setWorkMomentsListener", uid)
	wsRouter1.SetWorkMomentsListener()
	var rr RefRouter
	rr.refName = RouteMap1
	rr.wsRouter = &wsRouter1

	UserRouteMapNew[uid] = rr
	log.Info("", "insert UserRouteMap: ", uid)

	log.Debug(operationID, utils2.GetSelfFuncName(), "UserRouteMap length", len(UserRouteMapNew), uid, connAddr)
	return &rr
}

func (wsRouter *WsFuncRouter) GlobalSendMessage(data interface{}) {
	SendOneUserMessage(data, wsRouter.uId)
}

// listener
func SendOneUserMessage(data interface{}, uid string) {
	var chMsg ChanMsg
	chMsg.data, _ = json.Marshal(data)
	chMsg.uid = uid
	err := send2Ch(WS.ch, &chMsg, 2)
	if err != nil {
		log.Info("", "send2ch failed, ", err, string(chMsg.data), uid)
		return
	}
	log.Info("", "send response to web: ", string(chMsg.data))
}

func SendOneUserMessageForTest(data interface{}, uid string) {
	d, err := json.Marshal(data)
	log.Info("", "Marshal ", string(d))
	var chMsg ChanMsg
	chMsg.data = d
	chMsg.uid = uid
	err = send2ChForTest(WS.ch, chMsg, 2)
	if err != nil {
		log.Info("", "send2ch failed, ", err, string(chMsg.data), uid)
		return
	}
	log.Info("", "send response to web: ", string(chMsg.data))
}

func SendOneConnMessage(data interface{}, conn *UserConn) {
	bMsg, _ := json.Marshal(data)
	err := WS.writeMsg(conn, websocket.TextMessage, bMsg)
	log.Info("", "send response to web: ", string(bMsg))
	if err != nil {
		log.Info("", "WS WriteMsg error", "", "userIP", conn.RemoteAddr().String(), "error", err, "data", data)
	} else {
		log.Info("", "WS WriteMsg ok", "data", data)
	}
}

func send2ChForTest(ch chan ChanMsg, value ChanMsg, timeout int64) error {
	var t ChanMsg
	t = value
	log.Info("", "test uid ", t.uid)
	return nil
}

func send2Ch(ch chan ChanMsg, value *ChanMsg, timeout int64) error {
	var flag = 0
	select {
	case ch <- *value:
		flag = 1
	case <-time.After(time.Second * time.Duration(timeout)):
		flag = 2
	}
	if flag == 1 {
		return nil
	} else {
		log.Info("", "send cmd timeout, ", timeout, value)
		return errors.New("send cmd timeout")
	}
}
