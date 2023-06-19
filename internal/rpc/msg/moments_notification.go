package msg

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"

	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func MomentNotification(operationID, sendID, recvID string, m proto.Message) {
	var tips sdk.TipsComm
	var err error
	marshaler := jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  false,
		EmitDefaults: false,
	}
	tips.JsonDetail, _ = marshaler.MarshalToString(m)
	n := &NotificationMsg{
		SendID:      sendID,
		RecvID:      recvID,
		MsgFrom:     constant.UserMsgType,
		ContentType: constant.MomentNotification,
		SessionType: constant.SingleChatType,
		OperationID: operationID,
	}
	n.Content, err = proto.Marshal(&tips)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "proto.Marshal failed")
		return
	}
	log.NewInfo(operationID, utils.GetSelfFuncName(), string(n.Content))
	Notification(n)
}

func SendMomentInteractionNotification(operationID, momentID, sendID string, actionType int32, actionObject interface{}) {

	momentIDP, _ := primitive.ObjectIDFromHex(momentID)
	interactedUserList, err := db.DB.GetMomentInteractedUsers(momentIDP)
	senderFriendList, err1 := db.DB.GetFriendIDListFromCache(sendID)
	if err == nil && err1 == nil {
		for _, interactedUser := range interactedUserList {
			receiverID := interactedUser.(string)
			if receiverID != sendID && utils.IsContain(receiverID, senderFriendList) {
				momentCommentNotificationTips := open_im_sdk.MomentCommentNotificationTips{}
				momentCommentNotificationTips.MomentID = momentID
				momentCommentNotificationTips.ActionType = actionType
				actionObjectData, err := json.Marshal(actionObject)
				if err == nil {
					momentCommentNotificationTips.ActionObject = string(actionObjectData[:])
				}
				MomentNotification(operationID, sendID, receiverID, &momentCommentNotificationTips)
			}
		}

	}

}
