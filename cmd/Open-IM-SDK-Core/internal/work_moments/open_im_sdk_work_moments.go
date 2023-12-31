package workMoments

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
)

func (w *WorkMoments) SetListener(callback open_im_sdk_callback.OnWorkMomentsListener) {
	if callback == nil {
		log.NewError("", "callback is null")
		return
	}
	log.NewDebug("", "callback set success")
	w.listener = callback
}

func (w *WorkMoments) GetWorkMomentsUnReadCount(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName)
		result := w.getWorkMomentsNotificationUnReadCount(callback, operationID)
		callback.OnSuccess(utils.StructToJsonString(result))
		log.NewInfo(operationID, fName)
	}()
}

func (w *WorkMoments) GetWorkMomentsNotification(callback open_im_sdk_callback.Base, offset, count int, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, offset, count)
		result := w.getWorkMomentsNotification(offset, count, callback, operationID)
		callback.OnSuccess(utils.StructToJsonString(result))
		log.NewInfo(operationID, fName)
	}()
}

func (w *WorkMoments) ClearWorkMomentsNotification(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName)
		w.clearWorkMomentsNotification(callback, operationID)
		callback.OnSuccess(sdk_params_callback.ClearWorkMomentsMessageCallback)
		log.NewInfo(operationID, fName)
	}()
}
