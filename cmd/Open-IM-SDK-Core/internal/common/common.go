package common

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/common/log"
	sdk "Open_IM/pkg/proto/sdk_ws"
	"encoding/json"
	"github.com/golang/protobuf/proto"
)

func UnmarshalTips(msg *sdk.MsgData, detail proto.Message) error {
	var tips sdk.TipsComm
	if err := proto.Unmarshal(msg.Content, &tips); err != nil {
		return utils.Wrap(err, "")
	}
	if err := proto.Unmarshal(tips.Detail, detail); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func UnmarshalTipsSync(msg *sdk.MsgData, detail proto.Message) error {
	var tips sdk.TipsComm
	if err := proto.Unmarshal(msg.Content, &tips); err != nil {
		return utils.Wrap(err, "")
	}
	log.Error("UnmarshalTipsSync Content 1", tips.JsonDetail)

	err := json.Unmarshal([]byte(tips.JsonDetail), &detail)
	if err != nil {
		return utils.Wrap(err, "")
	}
	log.Error("UnmarshalTipsSync Content 2", detail)
	return nil
}
