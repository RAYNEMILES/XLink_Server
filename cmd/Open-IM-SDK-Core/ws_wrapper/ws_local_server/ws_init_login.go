package ws_local_server

import (
	"Open_IM/cmd/Open-IM-SDK-Core/internal/login"
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"encoding/json"

	//	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	//	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
)

type InitCallback struct {
	uid string
}

func (i *InitCallback) OnConnecting() {
	var ed EventData
	ed.Event = cleanUpfuncName(runFuncName())
	ed.ErrCode = 0
	SendOneUserMessage(ed, i.uid)
}

func (i *InitCallback) OnConnectSuccess() {
	var ed EventData
	ed.Event = cleanUpfuncName(runFuncName())
	ed.ErrCode = 0
	SendOneUserMessage(ed, i.uid)
}

func (i *InitCallback) OnConnectFailed(ErrCode int32, ErrMsg string) {
	var ed EventData
	ed.Event = cleanUpfuncName(runFuncName())
	ed.ErrCode = ErrCode
	ed.ErrMsg = ErrMsg
	SendOneUserMessage(ed, i.uid)
}

func (i *InitCallback) OnKickedOffline() {
	var ed EventData
	ed.Event = cleanUpfuncName(runFuncName())
	ed.ErrCode = 0
	SendOneUserMessage(ed, i.uid)
}

func (i *InitCallback) OnUserTokenExpired() {
	var ed EventData
	ed.Event = cleanUpfuncName(runFuncName())
	ed.ErrCode = 0
	SendOneUserMessage(ed, i.uid)
}

func (i *InitCallback) OnSelfInfoUpdated(userInfo string) {
	var ed EventData
	ed.Data = userInfo
	ed.Event = cleanUpfuncName(runFuncName())
	ed.ErrCode = 0
	SendOneUserMessage(ed, i.uid)
}

var ConfigSvr string

//func (wsRouter *WsFuncRouter) InitSDK(config string, operationID string) {
//	var initcb InitCallback
//	initcb.uid = wsRouter.uId
//	log.Info(operationID, "Initsdk uid: ", initcb.uid, config)
//	c := sdk_struct.IMConfig{}
//	json.Unmarshal([]byte(config), &c)
//
//	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
//	log.NewError("", utils.GetSelfFuncName(), "generate userWorker:", wsRouter.uId, wsRouter.connAddr)
//	log.NewError("", utils.GetSelfFuncName(), "after generate length:", wsRouter.uId, len(open_im_sdk.UserRouterMap[wsRouter.uId]))
//	log.NewError("", utils.GetSelfFuncName(), "after generate UserRouterMap length:", wsRouter.uId, len(open_im_sdk.UserRouterMap))
//
//	if userWorker.InitSDK(c, &initcb, operationID) {
//		//	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", "", operationID})
//	} else {
//		//	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), open_im_sdk.ErrCodeInitLogin, "init config failed", "", operationID})
//	}
//}

func (wsRouter *WsFuncRouter) InitSDKNew(config string, operationID string) {
	var initcb InitCallback
	initcb.uid = wsRouter.uId
	log.Info(operationID, "Initsdk uid: ", initcb.uid, config)
	c := sdk_struct.IMConfig{}
	json.Unmarshal([]byte(config), &c)

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	log.Debug("", utils.GetSelfFuncName(), "generate userWorker:", wsRouter.uId, wsRouter.connAddr)
	log.Debug("", utils.GetSelfFuncName(), "after generate UserRouterMap length:", wsRouter.uId, len(open_im_sdk.UserRouterMapNew))

	if userWorker.InitSDK(c, &initcb, operationID) {
		//	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", "", operationID})
	} else {
		//	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), open_im_sdk.ErrCodeInitLogin, "init config failed", "", operationID})
	}
}

//func (wsRouter *WsFuncRouter) UnInitSDK() {
//	log.NewError("", "UnInitSDK uid: ", wsRouter.uId, wsRouter.connAddr)
//	open_im_sdk.UserSDKRwLock.Lock()
//	if v, ok := open_im_sdk.UserRouterMap[wsRouter.uId]; ok {
//		if worker, ok := v[wsRouter.connAddr]; ok {
//			if worker != nil {
//				conv := worker.Conversation()
//				if conv != nil {
//					conv.Pool.Release()
//				}
//
//				worker = nil
//				delete(v, wsRouter.connAddr)
//				if len(v) != 0 {
//					log.NewError("", utils.GetSelfFuncName(), "delete v:", wsRouter.uId, wsRouter.connAddr)
//					log.NewError("", utils.GetSelfFuncName(), "after delete length:", wsRouter.uId, len(v))
//					log.NewError("", utils.GetSelfFuncName(), "after delete UserRouterMap length:", wsRouter.uId, len(open_im_sdk.UserRouterMap))
//					open_im_sdk.UserRouterMap[wsRouter.uId] = v
//				} else {
//					delete(open_im_sdk.UserRouterMap, wsRouter.uId)
//					log.NewError("", utils.GetSelfFuncName(), "after delete UserRouterMap length:", wsRouter.uId, len(open_im_sdk.UserRouterMap))
//				}
//
//			}
//		}
//	}
//	log.NewError("", "delete UnInitSDK uid: ", wsRouter.uId, wsRouter.connAddr)
//	open_im_sdk.UserSDKRwLock.Unlock()
//}

func (wsRouter *WsFuncRouter) UnInitSDKNew() {
	log.Debug("", "UnInitSDK uid: ", wsRouter.uId, wsRouter.connAddr)
	open_im_sdk.UserSDKRwLock.Lock()
	if worker, ok := open_im_sdk.UserRouterMapNew[wsRouter.uId]; ok {
		if worker != nil {
			//conv := worker.Conversation()
			//if conv != nil && conv.Pool != nil {
			//	conv.Pool.Release()
			//}

			worker = nil
			delete(open_im_sdk.UserRouterMapNew, wsRouter.uId)
			log.Debug("", utils.GetSelfFuncName(), "after delete UserRouterMap length:", wsRouter.uId, len(open_im_sdk.UserRouterMapNew))

		}
	}
	log.Debug("", "delete UnInitSDK uid: ", wsRouter.uId, wsRouter.connAddr)
	open_im_sdk.UserSDKRwLock.Unlock()
}

func (wsRouter *WsFuncRouter) checkResourceLoadingAndKeysIn(mgr *login.LoginMgr, input, operationID, funcName string, m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		_, ok := m[k]
		if !ok {
			log.Info(operationID, "key not in", keys, input, operationID, funcName)
			wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(funcName), StatusBadParameter, "key not in", "", operationID})
			return false
		}
	}

	if err := open_im_sdk.CheckResourceLoad(mgr); err != nil {
		log.Info(operationID, "Resource Loading ", mgr, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(funcName), StatusResourceNotCompleted, "resource loading is not completed", "", operationID})
		return false
	}
	return true
}

func (wsRouter *WsFuncRouter) checkKeysIn(input, operationID, funcName string, m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		_, ok := m[k]
		if !ok {
			log.Info(operationID, "key not in", keys, input, funcName, m)
			wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(funcName), StatusBadParameter, "key not in", "", operationID})
			return false
		}
	}
	return true
}

func (wsRouter *WsFuncRouter) Login(input string, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
		return
	}
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkKeysIn(input, operationID, runFuncName(), m, "userID", "token") {
		log.NewError(operationID, utils.GetSelfFuncName(), "checkKeysIn failed!")
		return
	}
	userWorker.Login(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, m["userID"].(string), m["token"].(string), operationID)
}

func (wsRouter *WsFuncRouter) Logout(input string, operationID string) {
	//userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	//userWorker.Logout(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId})
	//todo just send response
	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", "", operationID})
}

func (wsRouter *WsFuncRouter) LogoutNoCallback(input string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	userWorker.Logout(nil, operationID)
}

func (wsRouter *WsFuncRouter) GetLoginStatus(input string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), nil) {
		return
	}
	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", int32ToString(int32(userWorker.GetLoginStatus())), operationID})
}

// 1
func (wsRouter *WsFuncRouter) getMyLoginStatus() int32 {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, "", "", runFuncName(), nil) {
		return constant.SdkInit
	}
	return userWorker.GetLoginStatus()
}

// 1
func (wsRouter *WsFuncRouter) GetLoginUser(input string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), nil) {
		return
	}
	wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", userWorker.GetLoginUser(), operationID})
}

func InitServer(config *sdk_struct.IMConfig) {
	data, _ := json.Marshal(config)
	ConfigSvr = string(data)
	UserRouteMapNew = make(map[string]RefRouter, 0)
	open_im_sdk.InitOnce(config)
}
