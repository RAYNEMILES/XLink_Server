package interaction

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
)

func (u *WsConn) syncDataForUser(operationID, userID, msgType, msgData string) {

	log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", userID, msgData)

	if u.syncCh != nil {
		u.syncCh <- common.Cmd2Value{Cmd: msgType, Value: msgData}
	}

}
