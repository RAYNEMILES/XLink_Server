package open_im_sdk

import (
	"Open_IM/cmd/Open-IM-SDK-Core/internal/login"
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"errors"
	"reflect"
	"runtime"
	"sync"
)

func init() {
	UserSDKRwLock.Lock()
	defer UserSDKRwLock.Unlock()
	//UserRouterMap = make(map[string]map[string]*login.LoginMgr, 0)
	UserRouterMapNew = make(map[string]*login.LoginMgr, 0)
}

var UserSDKRwLock sync.RWMutex

//var UserRouterMap map[string]map[string]*login.LoginMgr
var UserRouterMapNew map[string]*login.LoginMgr

var userForSDK *login.LoginMgr

//func GetUserWorker(uid, connAddr string) *login.LoginMgr {
//	UserSDKRwLock.Lock()
//	defer UserSDKRwLock.Unlock()
//	v, ok := UserRouterMap[uid]
//	if ok {
//		if worker, ok := v[connAddr]; ok {
//			if worker != nil {
//				return worker
//			}
//		}
//		v[connAddr] = new(login.LoginMgr)
//		UserRouterMap[uid] = v
//		return v[connAddr]
//	}
//
//	loginMgr := new(login.LoginMgr)
//	UserRouterMap[uid] = map[string]*login.LoginMgr{connAddr: loginMgr}
//
//	return loginMgr
//}

func GetUserWorkerNew(uid string) *login.LoginMgr {
	UserSDKRwLock.RLock()
	v, ok := UserRouterMapNew[uid]
	if ok {
		UserSDKRwLock.RUnlock()
		return v
	}
	UserSDKRwLock.RUnlock()
	UserSDKRwLock.Lock()
	defer UserSDKRwLock.Unlock()
	loginMgr := new(login.LoginMgr)
	UserRouterMapNew[uid] = loginMgr

	return loginMgr
}

type Caller interface {
	BaseCaller(funcName interface{}, base open_im_sdk_callback.Base, args ...interface{})
	SendMessageCaller(funcName interface{}, messageCallback open_im_sdk_callback.SendMsgCallBack, args ...interface{})
}

type name struct {
}

var ErrNotSetCallback = errors.New("not set callback to call")
var ErrNotSetFunc = errors.New("not set func to call")

func BaseCaller(funcName interface{}, callback open_im_sdk_callback.Base, args ...interface{}) {
	var operationID string
	if len(args) <= 0 {
		callback.OnError(constant.ErrArgs.ErrCode, constant.ErrArgs.ErrMsg)
		return
	}
	if v, ok := args[len(args)-1].(string); ok {
		operationID = v
	} else {
		callback.OnError(constant.ErrArgs.ErrCode, constant.ErrArgs.ErrMsg)
		return
	}
	if err := CheckResourceLoad(userForSDK); err != nil {
		log.Error(operationID, "resource loading is not completed ", err.Error())
		callback.OnError(constant.ErrResourceLoadNotComplete.ErrCode, constant.ErrResourceLoadNotComplete.ErrMsg)
		return
	}
	defer func() {
		if rc := recover(); rc != nil {
			log.Error(operationID, "err:", rc)
		}
	}()
	if funcName == nil {
		panic(ErrNotSetFunc)
	}
	var refFuncName reflect.Value
	var values []reflect.Value
	refFuncName = reflect.ValueOf(funcName)
	if callback != nil {
		values = append(values, reflect.ValueOf(callback))
	} else {
		log.Error("AsyncCallWithCallback", "not set callback")
		panic(ErrNotSetCallback)
	}
	for i := 0; i < len(args); i++ {
		values = append(values, reflect.ValueOf(args[i]))
	}
	pc, _, _, _ := runtime.Caller(1)
	funcNameString := utils.CleanUpfuncName(runtime.FuncForPC(pc).Name())
	log.Debug(operationID, funcNameString, "input args:", args)
	go refFuncName.Call(values)
}
