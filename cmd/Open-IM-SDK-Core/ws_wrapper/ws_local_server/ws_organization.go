package ws_local_server

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk"
	"Open_IM/pkg/common/log"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"encoding/json"
)

type OrganizationCallback struct {
	uid string
}

func (g *OrganizationCallback) OnOrganizationUpdated() {
	SendOneUserMessage(EventData{cleanUpfuncName(runFuncName()), 0, "", "", "0"}, g.uid)
}

func (wsRouter *WsFuncRouter) SetOrganizationListener() {
	var g OrganizationCallback
	g.uid = wsRouter.uId
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	userWorker.SetOrganizationListener(&g)
}

func (wsRouter *WsFuncRouter) GetSubDepartment(input, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
		return
	}

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), m, "departmentID", "offset", "count") {
		return
	}
	//callback open_im_sdk_callback.Base, departmentID string, offset, count int, operationID string
	userWorker.Organization().GetSubDepartment(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		m["departmentID"].(string), int(m["offset"].(float64)), int(m["count"].(float64)),
		operationID)
}

func (wsRouter *WsFuncRouter) GetDepartmentMember(input, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
		return
	}

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), m, "departmentID", "offset", "count") {
		return
	}
	//callback open_im_sdk_callback.Base, departmentID string, offset, count int, operationID string
	userWorker.Organization().GetDepartmentMember(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		m["departmentID"].(string), int(m["offset"].(float64)), int(m["count"].(float64)),
		operationID)
}

func (wsRouter *WsFuncRouter) GetUserInDepartment(input, operationID string) {
	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	//callback open_im_sdk_callback.Base, departmentID string, offset, count int, operationID string
	userWorker.Organization().GetUserInDepartment(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		input, operationID)
}

func (wsRouter *WsFuncRouter) GetDepartmentMemberAndSubDepartment(input, operationID string) {

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	//callback open_im_sdk_callback.Base, departmentID string, offset, count int, operationID string
	userWorker.Organization().GetDepartmentMemberAndSubDepartment(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		input, operationID)
}

func (wsRouter *WsFuncRouter) GetParentDepartmentList(input, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
		return
	}

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), m, "departmentID") {
		return
	}
	//callback open_im_sdk_callback.Base, departmentID string, offset, count int, operationID string
	userWorker.Organization().GetParentDepartmentList(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		m["departmentID"].(string), operationID)
}

func (wsRouter *WsFuncRouter) GetDepartmentInfo(input, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
		return
	}

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), m, "departmentID") {
		return
	}
	userWorker.Organization().GetDepartmentInfo(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		m["departmentID"].(string), operationID)
}

func (wsRouter *WsFuncRouter) SearchOrganization(input, operationID string) {
	m := make(map[string]interface{})
	if err := json.Unmarshal([]byte(input), &m); err != nil {
		log.Info(operationID, utils.GetSelfFuncName(), "unmarshal failed", input, err.Error())
		wsRouter.GlobalSendMessage(EventData{cleanUpfuncName(runFuncName()), StatusBadParameter, "unmarshal failed", "", operationID})
		return
	}

	userWorker := open_im_sdk.GetUserWorkerNew(wsRouter.uId)
	if !wsRouter.checkResourceLoadingAndKeysIn(userWorker, input, operationID, runFuncName(), m, "input", "offset", "count") {
		return
	}
	userWorker.Organization().SearchOrganization(&BaseSuccessFailed{runFuncName(), operationID, wsRouter.uId},
		m["input"].(string), int(m["offset"].(float64)), int(m["count"].(float64)), operationID)
}
