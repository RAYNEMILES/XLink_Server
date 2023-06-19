package cronTask

import (
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/proto/short_video"
	"Open_IM/pkg/utils"
	"context"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	v20180717 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod/v20180717"
)

func newFileUpload(rpc short_video.ShortVideoClient, event *v20180717.EventContent) {
	operationId := getCronTaskOperationID()
	log.NewInfo(operationId, "newFileUpload", event)

	FileUploadEvent := &short_video.FileUploadEventMessage{}

	err := utils.CopyStructFields(&FileUploadEvent, &event.FileUploadEvent)
	if err != nil {
		log.NewError(operationId, "newFileUpload", err.Error())
		return
	}

	result, err := rpc.FileUploadCallBack(context.Background(), &short_video.FileUploadCallBackRequest{
		OperationID:     operationId,
		EventType:       "NewFileUpload",
		FileUploadEvent: FileUploadEvent,
	})
	if err != nil {
		log.NewError(operationId, "newFileUpload", err.Error())
		return
	}
	log.NewInfo(operationId, "newFileUpload", result)
}

func procedureStateChanged(rpc short_video.ShortVideoClient, event *v20180717.EventContent) {
	operationId := getCronTaskOperationID()
	log.NewInfo(operationId, "procedureStateChanged", event)

	procedureStateChangeEventMessage := &short_video.ProcedureStateChangeEventMessage{}

	err := utils.CopyStructFields(&procedureStateChangeEventMessage, &event.ProcedureStateChangeEvent)
	if err != nil {
		log.NewError(operationId, "newFileUpload", err.Error())
		return
	}

	result, err := rpc.ProcedureStateChangeCallBack(context.Background(), &short_video.ProcedureStateChangeCallBackRequest{
		OperationID:               operationId,
		EventType:                 "ProcedureStateChanged",
		ProcedureStateChangeEvent: procedureStateChangeEventMessage,
	})
	if err != nil {
		log.NewError(operationId, "ProcedureStateChanged", err.Error())
		return
	}
	log.NewInfo(operationId, "ProcedureStateChanged", result)
}

func fileDeleted(rpc short_video.ShortVideoClient, event *v20180717.EventContent) {
	operationId := getCronTaskOperationID()
	log.NewInfo(operationId, "fileDeleted", event)

	fileDeletedEventMessage := &short_video.FileDeleteEventMessage{}

	err := utils.CopyStructFields(&fileDeletedEventMessage, &event.FileDeleteEvent)
	fileDeletedEventMessage.FileIdSet = common.StringValues(event.FileDeleteEvent.FileIdSet)
	if err != nil {
		log.NewError(operationId, "fileDeleted", err.Error())
		return
	}

	result, err := rpc.FileDeletedCallBack(context.Background(), &short_video.FileDeletedCallBackRequest{
		OperationID:     operationId,
		EventType:       "FileDeleted",
		FileDeleteEvent: fileDeletedEventMessage,
	})
	if err != nil {
		log.NewError(operationId, "fileDeleted", err.Error())
		return
	}
	log.NewInfo(operationId, "fileDeleted", result)
}
