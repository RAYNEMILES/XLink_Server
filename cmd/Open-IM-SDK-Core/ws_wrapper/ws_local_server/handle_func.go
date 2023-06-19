/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/9/8 14:54).
 */
package ws_local_server

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"encoding/json"
	"reflect"
	"runtime"
)

type Req struct {
	ReqFuncName string `json:"reqFuncName" `
	OperationID string `json:"operationID"`
	Data        string `json:"data"`
	UserID      string `json:"userID"`
	Batch       int    `json:"batchMsg"`
}

//func (ws *WServer) DoLogin(m Req, conn *UserConn) {
//	if m.UserID == "" || conn == nil {
//		return
//	}
//	UserRouteRwLock.RLock()
//	defer func() {
//		UserRouteRwLock.RUnlock()
//		if r := recover(); r != nil {
//			log.NewError("", "DoLogin panic", " panic is ", r)
//			buf := make([]byte, 1<<20)
//			runtime.Stack(buf, true)
//			log.NewError("", "DoLogin panic", "call", string(buf))
//			ws.getMsgAndSend()
//			log.Info("", "goroutine getMsgAndSend restart")
//		}
//	}()
//	urm, ok := UserRouteMap[m.UserID]
//	log.NewInfo("", utils.GetSelfFuncName(), "urm:", urm)
//	if !ok {
//		log.Info("", "login", "user first login: ", m, "conn:", conn.RemoteAddr().String())
//		refR := GenUserRouterWithLock(m.UserID, conn.RemoteAddr().String(), m.Batch, m.OperationID)
//		params := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
//		vf, ok := (refR.refName)[m.ReqFuncName]
//		if ok {
//			vf.Call(params)
//		} else {
//			log.Info("", "login", "no func name: ", m.ReqFuncName, m)
//			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
//		}
//
//	} else {
//		if urm != nil {
//			if router, ok := urm[conn.RemoteAddr().String()]; ok {
//				if router.wsRouter.getMyLoginStatus() == constant.LoginSuccess {
//					//send ok
//					SendOneConnMessage(EventData{"Login", 0, "ok", "", m.OperationID}, conn)
//				} else {
//					log.Info("", "login", "login status pending, try after 5 second ", router.wsRouter.getMyLoginStatus(), m.UserID)
//					SendOneConnMessage(EventData{"Login", StatusLoginPending, StatusText(StatusLoginPending), "", m.OperationID}, conn)
//				}
//			} else {
//				log.Info("", "login", "user first login: ", m, conn.RemoteAddr().String())
//				refR := GenUserRouterWithLock(m.UserID, conn.RemoteAddr().String(), m.Batch, m.OperationID)
//				params := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
//				vf, ok := (refR.refName)[m.ReqFuncName]
//				if ok {
//					vf.Call(params)
//				} else {
//					log.Info("", "login", "no func name-1: ", m.ReqFuncName, m)
//					SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
//				}
//			}
//		} else {
//			log.Info("", "login", "user first login: ", m, conn.RemoteAddr().String())
//			refR := GenUserRouterWithLock(m.UserID, conn.RemoteAddr().String(), m.Batch, m.OperationID)
//			params := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
//			vf, ok := (refR.refName)[m.ReqFuncName]
//			if ok {
//				vf.Call(params)
//			} else {
//				log.Info("", "login", "no func name-1: ", m.ReqFuncName, m)
//				SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
//			}
//		}
//
//	}
//}

func (ws *WServer) DoLoginNew(m Req, conn *UserConn) {
	if m.UserID == "" || conn == nil {
		return
	}

	//defer func() {
	//	if r := recover(); r != nil {
	//		log.NewError("", "DoLogin panic", " panic is ", r)
	//		buf := make([]byte, 1<<20)
	//		runtime.Stack(buf, true)
	//		log.NewError("", "DoLogin panic", "call", string(buf))
	//		log.Info("", "goroutine getMsgAndSend restart")
	//	}
	//}()
	UserRouteRwLock.RLock()
	urm, ok := UserRouteMapNew[m.UserID]
	UserRouteRwLock.RUnlock()

	log.NewInfo("", utils.GetSelfFuncName(), "urm:", urm)
	if !ok {
		log.Info("", "login", "user first login: ", m, "conn:", conn.RemoteAddr().String())
		refR := GenUserRouterWithLockNew(m.UserID, conn.RemoteAddr().String(), m.Batch, m.OperationID)
		params := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
		vf, ok := (refR.refName)[m.ReqFuncName]
		if ok {
			vf.Call(params)
		} else {
			log.Info("", "login", "no func name: ", m.ReqFuncName, m)
			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
		}

	} else {
		if urm.wsRouter.getMyLoginStatus() == constant.LoginSuccess {
			//send ok
			SendOneConnMessage(EventData{"Login", 0, "ok", "", m.OperationID}, conn)
		} else {
			log.NewError("", "login", "login status pending, try after 5 second ", urm.wsRouter.getMyLoginStatus(), m.UserID)
			SendOneConnMessage(EventData{"Login", StatusLoginPending, StatusText(StatusLoginPending), "", m.OperationID}, conn)
		}

	}
}

//func (ws *WServer) msgParse(conn *UserConn, jsonMsg []byte) {
//	m := Req{}
//	if err := json.Unmarshal(jsonMsg, &m); err != nil {
//		SendOneConnMessage(EventData{"error", 100, "Unmarshal failed", "", ""}, conn)
//		return
//	}
//
//	defer func() {
//		if r := recover(); r != nil {
//			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
//			log.Info("", "msgParse", "bad request, panic is ", r)
//			buf := make([]byte, 1<<16)
//			runtime.Stack(buf, true)
//			log.Info("", "msgParse", "call", string(buf))
//		}
//	}()
//
//	log.Info("", "msgParse", "recv request from web: ", "reqFuncName ", m.ReqFuncName, "data ", m.Data, "recv jsonMsg: ", string(jsonMsg))
//
//	//check token and platform
//	if conn.token == "" || conn.platformID == "" {
//		log.NewError("", "msgParse", "token or platformID is nil: ", conn.token, conn.platformID)
//		SendOneConnMessage(EventData{"Login", StatusNoLogin, StatusText(StatusNoLogin), "", m.OperationID}, conn)
//		return
//	}
//
//	//_, err, _ := token_verify.WsVerifyToken(conn.token, m.UserID, conn.platformID, m.OperationID)
//	//if err != nil {
//	//	log.NewError("", "msgParse", "WsVerifyToken is failed!: ", conn.token, m.UserID, conn.platformID, err.Error())
//	//	SendOneConnMessage(EventData{"Login", StatusForbidden, StatusText(StatusForbidden), "", m.OperationID}, conn)
//	//	return
//	//}
//	////check permissions
//	//err2 := utils2.CheckUserPermissions(m.UserID)
//	//if err2 != nil {
//	//	log.NewError("", "msgParse", "CheckUserPermissions is failed!: ", conn.token, m.UserID, conn.platformID)
//	//	SendOneConnMessage(EventData{"Login", StatusForbidden, StatusText(StatusForbidden), "", m.OperationID}, conn)
//	//	return
//	//}
//
//	db, err := db.NewDataBase(conn.userID, sdk_struct.SvrConf.DataDir)
//	if err != nil {
//		log.NewError("", "msgParse", "NewDataBase is failed!: ", conn.token, m.UserID, conn.platformID)
//		SendOneConnMessage(EventData{"Login", StatusInternalServerError, StatusText(StatusInternalServerError), "", m.OperationID}, conn)
//		return
//	}
//
//	if db != nil {
//		localUser, _ := db.GetLoginUser()
//		if localUser != nil && localUser.Status == 2 {
//			log.NewError("", "msgParse", "CheckUserPermissions is failed!: ", conn.token, m.UserID, conn.platformID)
//			SendOneConnMessage(EventData{"Login", StatusForbidden, StatusText(StatusForbidden), "", m.OperationID}, conn)
//			return
//		}
//	}
//
//	if m.ReqFuncName == "Login" {
//		//	rwLock.Lock()
//		ws.DoLogin(m, conn)
//		log.Info("", "msgParse", m)
//		//	rwLock.Unlock()
//		return
//	}
//
//	UserRouteRwLock.RLock()
//	defer UserRouteRwLock.RUnlock()
//	//	rwLock.RLock()
//	//	defer rwLock.RUnlock()
//	urm, ok := UserRouteMap[m.UserID]
//
//	if !ok {
//		log.Info("", "msgParse", "user not login failed, must login first: ", m.UserID)
//		SendOneConnMessage(EventData{"Login", StatusNoLogin, StatusText(StatusNoLogin), "", m.OperationID}, conn)
//		return
//	}
//
//	parms := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}
//
//	if router, ok := urm[conn.RemoteAddr().String()]; ok {
//		vf, ok := (router.refName)[m.ReqFuncName]
//		if ok {
//			vf.Call(parms)
//		} else {
//			log.Info("", "msgParse", "no func ", m.ReqFuncName)
//			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
//		}
//	} else {
//		log.Info("", "msgParse", "can't find router", m.UserID, conn.RemoteAddr().String())
//		SendOneConnMessage(EventData{"Login", StatusNoLogin, StatusText(StatusNoLogin), "", m.OperationID}, conn)
//		return
//	}
//
//}

func (ws *WServer) msgParseNew(conn *UserConn, jsonMsg []byte) {
	m := Req{}
	if err := json.Unmarshal(jsonMsg, &m); err != nil {
		SendOneConnMessage(EventData{"error", 100, "Unmarshal failed", "", ""}, conn)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
			log.NewError("", "msgParse", "bad request, panic is ", r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "msgParse", "call", string(buf))
		}
	}()

	log.Info("", "msgParse", "recv request from web: ", "reqFuncName ", m.ReqFuncName, "data ", m.Data, "recv jsonMsg: ", string(jsonMsg))

	//check token and platform
	if conn.token == "" || conn.platformID == "" {
		log.NewError("", "msgParse", "token or platformID is nil: ", conn.token, conn.platformID)
		SendOneConnMessage(EventData{"Login", StatusNoLogin, StatusText(StatusNoLogin), "", m.OperationID}, conn)
		return
	}

	db, err := db.NewDataBase(conn.userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		log.NewError("", "msgParse", "NewDataBase is failed!: ", conn.token, m.UserID, conn.platformID)
		SendOneConnMessage(EventData{"Login", StatusBadRequest, StatusText(StatusBadRequest), "", m.OperationID}, conn)
		return
	}

	if db != nil {
		localUser, _ := db.GetLoginUser()
		if localUser != nil && localUser.Status == 2 {
			log.Debug("", "msgParse", "CheckUserPermissions is failed!: ", conn.token, m.UserID, conn.platformID)
			SendOneConnMessage(EventData{"Login", StatusForbidden, StatusText(StatusForbidden), "", m.OperationID}, conn)
			return
		}
	}

	if m.ReqFuncName == "Login" {
		//	rwLock.Lock()
		m.Batch = 1 //open batch messages listener
		ws.DoLoginNew(m, conn)
		log.Info("", "msgParse", m)
		//	rwLock.Unlock()
		return
	}

	UserRouteRwLock.RLock()
	router, ok := UserRouteMapNew[m.UserID]
	UserRouteRwLock.RUnlock()
	if !ok {
		log.Info("", "msgParse", "user login failed, must login first: ", m.UserID)
		SendOneConnMessage(EventData{"Login", StatusNoLogin, StatusText(StatusNoLogin), "", m.OperationID}, conn)
		return
	}

	parms := []reflect.Value{reflect.ValueOf(m.Data), reflect.ValueOf(m.OperationID)}

	vf, ok := (router.refName)[m.ReqFuncName]
	if ok {
		vf.Call(parms)
	} else {
		log.Info("", "msgParse", "no func ", m.ReqFuncName)
		SendOneConnMessage(EventData{m.ReqFuncName, StatusBadParameter, StatusText(StatusBadParameter), "", m.OperationID}, conn)
	}

}
