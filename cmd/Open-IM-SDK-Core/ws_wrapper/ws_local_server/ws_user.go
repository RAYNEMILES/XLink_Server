package ws_local_server

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk"
)

func (wsRouter *WsFuncRouter) GetUsersInfo(userIDList string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, userIDList, operationID, runFuncName(), nil) {
		return
	}
	userWorker.Full().GetUsersInfo(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, userIDList, operationID)
}

func (wsRouter *WsFuncRouter) SetSelfInfo(userInfo string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, userInfo, operationID, runFuncName(), nil) {
		return
	}
	userWorker.User().SetSelfInfo(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, userInfo, operationID)
}

func (wsRouter *WsFuncRouter) RemoveFaceUrl(userInfo string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, userInfo, operationID, runFuncName(), nil) {
		return
	}
	userWorker.User().RemoveFaceUrl(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, userInfo, operationID)
}

func (wsRouter *WsFuncRouter) GetSelfUserInfo(input string, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), nil) {
		return
	}
	userWorker.User().GetSelfUserInfo(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, operationID)
}

type UserCallback struct {
	uid string
}

func (u *UserCallback) OnSelfInfoUpdated(userInfo string) {
	SendOneUserMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", userInfo, "0"}, u.uid)
}
func (u *UserCallback) OnMomentNotification(momentInfoIDAction string) {
	SendOneUserMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", momentInfoIDAction, "0"}, u.uid)
}

func (wsRouter *WsFuncRouter) SetUserListener() {
	var u UserCallback
	u.uid = wsRouter.uId
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	userWorker.SetUserListener(&u)
}
