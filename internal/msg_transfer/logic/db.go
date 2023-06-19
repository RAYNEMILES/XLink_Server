package logic

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	pbMsg "Open_IM/pkg/proto/chat"
	"Open_IM/pkg/utils"
	"strings"
)

func saveUserChat(uid string, chatType int, msg *pbMsg.MsgDataToMQ) error {
	_ = utils.GetCurrentTimestampByMill()
	seq, err := db.DB.IncrUserSeq(uid)
	if err != nil {
		log.NewError(msg.OperationID, "data insert to redis err", err.Error(), msg.String())
		return err
	}
	msg.MsgData.Seq = uint32(seq)
	pbSaveData := pbMsg.MsgDataToDB{}
	pbSaveData.MsgData = msg.MsgData
	//log.NewInfo(msg.OperationID, "IncrUserSeq cost time", utils.GetCurrentTimestampByMill()-time)
	return db.DB.SaveUserChatMongo2(uid, chatType, pbSaveData.MsgData.SendTime, &pbSaveData)
	//	return db.DB.SaveUserChatMongo2(uid, pbSaveData.MsgData.SendTime, &pbSaveData)
}

func saveUserChatList(userID string, chatType int, msgList []*pbMsg.MsgDataToMQ, operationID string) (error, uint64) {
	log.Info(operationID, utils.GetSelfFuncName(), "args ", userID, len(msgList))

	if chatType == constant.GroupChatType {
		if strings.Contains(userID, "group-") {
			userID = strings.TrimPrefix(userID, "group-")
		}
	}
	//return db.DB.BatchInsertChat(userID, msgList, operationID)
	return db.DB.BatchInsertChat2Cache(userID, chatType, msgList, operationID)
}
