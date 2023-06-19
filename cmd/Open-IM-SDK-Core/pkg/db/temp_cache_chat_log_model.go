package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
)

func (d *DataBase) BatchInsertTempCacheMessageList(MessageList []*model_struct.TempCacheLocalChatLog) error {
	if MessageList == nil {
		return nil
	}
	return utils.Wrap(d.conn.Create(MessageList).Error, "BatchInsertTempCacheMessageList failed")
}
func (d *DataBase) InsertTempCacheMessage(Message *model_struct.TempCacheLocalChatLog) error {

	return utils.Wrap(d.conn.Create(Message).Error, "InsertTempCacheMessage failed")

}
