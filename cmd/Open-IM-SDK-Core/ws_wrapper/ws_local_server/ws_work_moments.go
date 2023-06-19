package ws_local_server

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"encoding/json"
)

type WorkMomentsCallback struct {
	uid string
}

func (wsRouter *WsFuncRouter) SetWorkMomentsListener() {
	var cb WorkMomentsCallback
	cb.uid = wsRouter.uId
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	userWorker.SetWorkMomentsListener(&cb)
}

func (w *WorkMomentsCallback) OnRecvNewNotification() {
	SendOneUserMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", "", "0"}, w.uid)
}

func (wsRouter *WsFuncRouter) GetWorkMomentsUnReadCount(input, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), nil) {
		return
	}
	userWorker.WorkMoments().GetWorkMomentsUnReadCount(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, operationID)
}

func (wsRouter *WsFuncRouter) GetWorkMomentsNotification(input, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
	}
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), nil) {
		return
	}
	userWorker.WorkMoments().GetWorkMomentsNotification(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, int(m["offset"].(float64)), int(m["count"].(float64)), operationID)
}

func (wsRouter *WsFuncRouter) ClearWorkMomentsNotification(input, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), nil) {
		return
	}
	userWorker.WorkMoments().ClearWorkMomentsNotification(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId}, operationID)
}
