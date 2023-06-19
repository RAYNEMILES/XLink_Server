package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	"Open_IM/pkg/common/log"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func (d *DataBase) BatchInsertMessageList(MessageList []*model_struct.LocalChatLog) error {
	if MessageList == nil {
		return nil
	}
	go func() {
		localSeqSynced := &model_struct.LocalSeqSynced{}
		for _, chatLog := range MessageList {
			localSeqSynced.ConversationID = utils.GetConversationIDBySessionType(chatLog.RecvID, constant.SingleChatType)
			localSeqSynced.Seq = chatLog.Seq
			if localSeqSynced.Seq > 0 {
				recordUpdated := d.conn.Updates(localSeqSynced).RowsAffected
				if recordUpdated == 0 {
					d.conn.Create(localSeqSynced)
				}
			}
		}

	}()
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(MessageList).Error, "BatchInsertMessageList failed")
}
func (d *DataBase) BatchInsertMessageListController(MessageList []*model_struct.LocalChatLog) error {
	if len(MessageList) == 0 {
		return nil
	}
	switch MessageList[len(MessageList)-1].SessionType {
	case constant.GroupChatType:
		return d.SuperGroupBatchInsertMessageList(MessageList, MessageList[len(MessageList)-1].RecvID)
	default:
		return d.BatchInsertMessageList(MessageList)
	}
}
func (d *DataBase) InsertMessage(Message *model_struct.LocalChatLog) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	go func() {
		localSeqSynced := &model_struct.LocalSeqSynced{}
		localSeqSynced.ConversationID = utils.GetConversationIDBySessionType(Message.RecvID, constant.SingleChatType)
		localSeqSynced.Seq = Message.Seq
		if localSeqSynced.Seq > 0 {
			recordUpdated := d.conn.Updates(localSeqSynced).RowsAffected
			if recordUpdated == 0 {
				d.conn.Create(localSeqSynced)
			}
		}
	}()
	return utils.Wrap(d.conn.Create(Message).Error, "InsertMessage failed")
}
func (d *DataBase) InsertMessageController(message *model_struct.LocalChatLog) error {
	switch message.SessionType {
	case constant.GroupChatType:
		return d.SuperGroupInsertMessage(message, message.RecvID)
	default:
		return d.InsertMessage(message)
	}
}
func (d *DataBase) SearchMessageByKeyword(contentType []int, hasLink int, keywordList []string, keywordListMatchType int, sourceID string, startTime, endTime int64, sessionType, offset, count int) (result []*model_struct.LocalChatLog, err error) {
	var messageList []model_struct.LocalChatLog
	var condition string
	var subCondition string
	if keywordListMatchType == constant.KeywordMatchOr {
		for i := 0; i < len(keywordList); i++ {
			if i == 0 {
				subCondition += "And ("
			}
			if i+1 >= len(keywordList) {
				subCondition += "content like " + "'%" + keywordList[i] + "%') "
			} else {
				subCondition += "content like " + "'%" + keywordList[i] + "%' " + "or "

			}
		}
	} else {
		for i := 0; i < len(keywordList); i++ {
			if i == 0 {
				subCondition += "And ("
			}
			if i+1 >= len(keywordList) {
				subCondition += "content like " + "'%" + keywordList[i] + "%') "
			} else {
				subCondition += "content like " + "'%" + keywordList[i] + "%' " + "and "
			}
		}
	}
	switch sessionType {
	case constant.SingleChatType, constant.NotificationChatType:
		condition = fmt.Sprintf("session_type==%d And (send_id==%q OR recv_id==%q) And send_time  between %d and %d AND status <=%d  And content_type IN ? ", constant.SingleChatType, sourceID, sourceID, startTime, endTime, constant.MsgStatusSendFailed)
	case constant.GroupChatType:
		condition = fmt.Sprintf("session_type==%d And recv_id==%q And send_time between %d and %d AND status <=%d  And content_type IN ? ", constant.GroupChatType, sourceID, startTime, endTime, constant.MsgStatusSendFailed)
	default:
		//condition = fmt.Sprintf("(send_id==%q OR recv_id==%q) And send_time between %d and %d AND status <=%d And content_type == %d And content like %q", sourceID, sourceID, startTime, endTime, constant.MsgStatusSendFailed, constant.Text, "%"+keyword+"%")
		return nil, err
	}
	if hasLink != 0 {
		condition += fmt.Sprintf("has_link=%d", hasLink-1)
	}
	condition += subCondition
	err = utils.Wrap(d.conn.Where(condition, contentType).Order("send_time DESC").Offset(offset).Limit(count).Find(&messageList).Error, "InsertMessage failed")

	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}
func (d *DataBase) SearchMessageByKeywordController(contentType []int, hasLink int, keywordList []string, keywordListMatchType int, sourceID string, startTime, endTime int64, sessionType, offset, count int) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupSearchMessageByKeyword(contentType, hasLink, keywordList, keywordListMatchType, sourceID, startTime, endTime, sessionType, offset, count)
	default:
		return d.SearchMessageByKeyword(contentType, hasLink, keywordList, keywordListMatchType, sourceID, startTime, endTime, sessionType, offset, count)
	}
}

func (d *DataBase) SearchMessageByContentType(contentType []int, hasLink int, sourceID string, startTime, endTime int64, sessionType, offset, count int) (result []*model_struct.LocalChatLog, err error) {
	var messageList []model_struct.LocalChatLog
	var condition string
	switch sessionType {
	case constant.SingleChatType, constant.NotificationChatType:
		condition = fmt.Sprintf("session_type==%d And (send_id==%q OR recv_id==%q) And send_time between %d and %d AND status <=%d And content_type IN ?", constant.SingleChatType, sourceID, sourceID, startTime, endTime, constant.MsgStatusSendFailed)
	case constant.GroupChatType:
		condition = fmt.Sprintf("session_type==%d And recv_id==%q And send_time between %d and %d AND status <=%d And content_type IN ?", constant.GroupChatType, sourceID, startTime, endTime, constant.MsgStatusSendFailed)
	default:
		return nil, err
	}
	if hasLink != 0 {
		condition += fmt.Sprintf("has_link=%d", hasLink-1)
	}

	err = utils.Wrap(d.conn.Where(condition, contentType).Order("send_time DESC").Offset(offset).Limit(count).Find(&messageList).Error, "SearchMessage failed")
	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}

func (d *DataBase) SearchMessageByContentTypeController(contentType []int, hasLink int, sourceID string, startTime, endTime int64, sessionType, offset, count int) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupSearchMessageByContentType(contentType, hasLink, sourceID, startTime, endTime, sessionType, offset, count)
	default:
		return d.SearchMessageByContentType(contentType, hasLink, sourceID, startTime, endTime, sessionType, offset, count)
	}
}

func (d *DataBase) SearchMessageByContentTypeAndKeyword(contentType []int, hasLink int, keywordList []string, keywordListMatchType int, startTime, endTime int64) (result []*model_struct.LocalChatLog, err error) {
	var messageList []model_struct.LocalChatLog
	var condition string
	var subCondition string
	if keywordListMatchType == constant.KeywordMatchOr {
		for i := 0; i < len(keywordList); i++ {
			if i == 0 {
				subCondition += "And ("
			}
			if i+1 >= len(keywordList) {
				subCondition += "content like " + "'%" + keywordList[i] + "%') "
			} else {
				subCondition += "content like " + "'%" + keywordList[i] + "%' " + "or "

			}
		}
	} else {
		for i := 0; i < len(keywordList); i++ {
			if i == 0 {
				subCondition += "And ("
			}
			if i+1 >= len(keywordList) {
				subCondition += "content like " + "'%" + keywordList[i] + "%') "
			} else {
				subCondition += "content like " + "'%" + keywordList[i] + "%' " + "and "
			}
		}
	}
	condition = fmt.Sprintf("send_time between %d and %d AND status <=%d  And content_type IN ? ", startTime, endTime, constant.MsgStatusSendFailed)
	if hasLink != 0 {
		condition += fmt.Sprintf("has_link=%d", hasLink-1)
	}
	condition += subCondition
	log.Info("key owrd", condition)
	err = utils.Wrap(d.conn.Where(condition, contentType).Order("send_time DESC").Find(&messageList).Error, "SearchMessage failed")
	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}
func (d *DataBase) SearchMessageByContentTypeAndKeywordController(contentType []int, hasLink int, keywordList []string, keywordListMatchType int, startTime, endTime int64, operationID string) (result []*model_struct.LocalChatLog, err error) {
	list, err := d.SearchMessageByContentTypeAndKeyword(contentType, hasLink, keywordList, keywordListMatchType, startTime, endTime)
	if err != nil {
		return nil, err
	}
	groupIDList, err := d.GetJoinedGroupIDList()
	if err != nil {
		return nil, err
	}
	for _, v := range groupIDList {
		sList, err := d.SuperGroupSearchMessageByContentTypeAndKeyword(contentType, hasLink, keywordList, keywordListMatchType, startTime, endTime, v)
		if err != nil {
			log.Error(operationID, "search message in group err", err.Error(), v)
			continue
		}
		list = append(list, sList...)
	}
	//workingGroupIDList, err := d.GetJoinedWorkingGroupIDList()
	//if err != nil {
	//	return nil, err
	//}
	//for _, v := range workingGroupIDList {
	//	sList, err := d.SuperGroupSearchMessageByContentTypeAndKeyword(contentType, keywordList, keywordListMatchType, startTime, endTime, v)
	//	if err != nil {
	//		log.Error(operationID, "search message in group err", err.Error(), v)
	//		continue
	//	}
	//	list = append(list, sList...)
	//}
	return list, nil
}
func (d *DataBase) BatchUpdateMessageList(MessageList []*model_struct.LocalChatLog) error {
	if MessageList == nil {
		return nil
	}

	for _, v := range MessageList {
		v1 := new(model_struct.LocalChatLog)
		v1.ClientMsgID = v.ClientMsgID
		v1.Seq = v.Seq
		v1.Status = v.Status
		v1.RecvID = v.RecvID
		v1.SessionType = v.SessionType
		v1.ServerMsgID = v.ServerMsgID
		err := d.UpdateMessageController(v1)
		if err != nil {
			return utils.Wrap(err, "BatchUpdateMessageList failed")
		}

	}
	return nil
}

func (d *DataBase) BatchSpecialUpdateMessageList(MessageList []*model_struct.LocalChatLog) error {
	if MessageList == nil {
		return nil
	}

	for _, v := range MessageList {
		v1 := new(model_struct.LocalChatLog)
		v1.ClientMsgID = v.ClientMsgID
		v1.ServerMsgID = v.ServerMsgID
		v1.SendID = v.SendID
		v1.RecvID = v.RecvID
		v1.SenderPlatformID = v.SenderPlatformID
		v1.SenderNickname = v.SenderNickname
		v1.SenderFaceURL = v.SenderFaceURL
		v1.SessionType = v.SessionType
		v1.MsgFrom = v.MsgFrom
		v1.ContentType = v.ContentType
		v1.Content = v.Content
		v1.Seq = v.Seq
		v1.SendTime = v.SendTime
		v1.CreateTime = v.CreateTime
		v1.AttachedInfo = v.AttachedInfo
		v1.Ex = v.Ex
		err := d.UpdateMessageController(v1)
		if err != nil {
			log.Error("", "update single message failed", *v)
			return utils.Wrap(err, "BatchUpdateMessageList failed")
		}

	}
	return nil
}

func (d *DataBase) MessageIfExists(ClientMsgID string) (bool, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var count int64
	t := d.conn.Model(&model_struct.LocalChatLog{}).Where("client_msg_id = ?",
		ClientMsgID).Count(&count)
	if t.Error != nil {
		return false, utils.Wrap(t.Error, "MessageIfExists get failed")
	}
	if count != 1 {
		return false, nil
	} else {
		return true, nil
	}
}
func (d *DataBase) IsExistsInErrChatLogBySeq(seq int64) bool {
	return true
}
func (d *DataBase) MessageIfExistsBySeq(seq int64) (bool, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var count int64
	t := d.conn.Model(&model_struct.LocalChatLog{}).Where("seq = ?",
		seq).Count(&count)
	if t.Error != nil {
		return false, utils.Wrap(t.Error, "MessageIfExistsBySeq get failed")
	}
	if count != 1 {
		return false, nil
	} else {
		return true, nil
	}
}
func (d *DataBase) GetMessage(ClientMsgID string) (*model_struct.LocalChatLog, error) {
	//d.mRWMutex.RLock()
	//defer d.mRWMutex.RUnlock()
	var c model_struct.LocalChatLog
	return &c, utils.Wrap(d.conn.Where("client_msg_id = ?",
		ClientMsgID).Debug().Take(&c).Error, "GetMessage failed")
}
func (d *DataBase) GetMessageController(msg *sdk_struct.MsgStruct) (*model_struct.LocalChatLog, error) {
	switch msg.SessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMessage(msg)
	default:
		return d.GetMessage(msg.ClientMsgID)
	}
}

func (d *DataBase) GetAllUnDeleteMessageSeqList() ([]uint32, error) {
	var seqList []uint32
	return seqList, utils.Wrap(d.conn.Model(&model_struct.LocalChatLog{}).Where("status != ?", constant.MsgStatusHasDeleted).Select("seq").Find(&seqList).Error, "")
}
func (d *DataBase) UpdateColumnsMessageList(clientMsgIDList []string, args map[string]interface{}) error {
	c := model_struct.LocalChatLog{}
	t := d.conn.Model(&c).Where("client_msg_id IN", clientMsgIDList).Updates(args)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateColumnsConversation failed")
}
func (d *DataBase) UpdateColumnsMessage(ClientMsgID string, args map[string]interface{}) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalChatLog{ClientMsgID: ClientMsgID}
	t := d.conn.Model(&c).Updates(args)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateColumnsConversation failed")
}
func (d *DataBase) UpdateColumnsMessageController(ClientMsgID string, groupID string, sessionType int32, args map[string]interface{}) error {
	switch sessionType {
	case constant.GroupChatType:
		return utils.Wrap(d.SuperGroupUpdateColumnsMessage(ClientMsgID, groupID, args), "")
	default:
		return utils.Wrap(d.UpdateColumnsMessage(ClientMsgID, args), "")
	}
}
func (d *DataBase) UpdateMessage(c *model_struct.LocalChatLog) error {
	t := d.conn.Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update ")
	}
	return utils.Wrap(t.Error, "UpdateMessage failed")
}
func (d *DataBase) UpdateMessageController(c *model_struct.LocalChatLog) error {
	switch c.SessionType {
	case constant.GroupChatType:
		return utils.Wrap(d.SuperGroupUpdateMessage(c), "")
	default:
		return utils.Wrap(d.UpdateMessage(c), "")
	}
}

func (d *DataBase) DeleteAllMessage() error {
	m := model_struct.LocalChatLog{Status: constant.MsgStatusHasDeleted, Content: ""}
	err := d.conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Select("status", "content").Updates(m).Error
	return utils.Wrap(err, "delete all message error")
}
func (d *DataBase) UpdateMessageStatusBySourceID(sourceID string, status, sessionType int32) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var condition string
	if sourceID == d.loginUserID && sessionType == constant.SingleChatType {
		condition = "send_id=? And recv_id=? AND session_type=?"
	} else {
		condition = "(send_id=? or recv_id=?)AND session_type=?"
	}
	t := d.conn.Model(model_struct.LocalChatLog{}).Where(condition, sourceID, sourceID, sessionType).Updates(model_struct.LocalChatLog{Status: status})
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateMessageStatusBySourceID failed")
}
func (d *DataBase) UpdateMessageStatusBySourceIDController(sourceID string, status, sessionType int32) error {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupUpdateMessageStatusBySourceID(sourceID, status, sessionType)
	default:
		return d.UpdateMessageStatusBySourceID(sourceID, status, sessionType)
	}
}
func (d *DataBase) UpdateMessageTimeAndStatus(clientMsgID string, serverMsgID string, sendTime int64, status int32) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Where("client_msg_id=? And seq=?", clientMsgID, 0).Updates(model_struct.LocalChatLog{Status: status, SendTime: sendTime, ServerMsgID: serverMsgID}).Error, "UpdateMessageStatusBySourceID failed")

}
func (d *DataBase) UpdateMessageTimeAndStatusController(msg *sdk_struct.MsgStruct) error {
	switch msg.SessionType {
	case constant.GroupChatType:
		return d.SuperGroupUpdateMessageTimeAndStatus(msg)
	default:
		return d.UpdateMessageTimeAndStatus(msg.ClientMsgID, msg.ServerMsgID, msg.SendTime, msg.Status)
	}
}

func (d *DataBase) UpdateBroadcastMessageTimeAndStatusController(msg *sdk_struct.BroadcastMsgStruct) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Model(model_struct.LocalBroadcastChatLog{}).Where("client_msg_id=? ", msg.ClientMsgID, 0).Updates(model_struct.LocalBroadcastChatLog{Status: msg.Status, SendTime: msg.SendTime, ServerMsgID: msg.ServerMsgID}).Error, "UpdateMessageStatusBySourceID failed")

}

func (d *DataBase) UpdateBroadcastMessageController(msg *model_struct.LocalBroadcastChatLog) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Model(model_struct.LocalBroadcastChatLog{}).Where("client_msg_id=? ", msg.ClientMsgID, 0).Updates(msg).Error, "UpdateMessageStatusBySourceID failed")

}

func (d *DataBase) GetMessageListByStartMsgId(sourceID string, sessionType int, startTime, endTime int64) (result []*model_struct.LocalChatLog, err error) {
	var condition, timeOrder, seqOrder string
	timeOrder = "send_time DESC"
	seqOrder = "seq DESC"
	if sessionType == constant.SingleChatType && sourceID == d.loginUserID {
		condition = "send_id = ? And recv_id = ? AND status <=? And session_type = ? And send_time < ? AND send_time >= ?"
	} else {
		condition = "(send_id = ? OR recv_id = ?) AND status <=? And session_type = ? And send_time < ? AND send_time >= ?"
	}
	log.NewInfo("", utils.GetSelfFuncName(), condition, timeOrder, seqOrder)
	err = utils.Wrap(d.conn.Where(condition, sourceID, sourceID, constant.MsgStatusSendFailed, sessionType, startTime, endTime).
		Order(timeOrder).Find(&result).Error, "GetMessageList failed")

	return result, nil
}
func (d *DataBase) GetMessageListByStartMsgIdController(sourceID string, sessionType int, startTime, endTime int64) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMessageListByStartMsgId(sourceID, sessionType, startTime, endTime)
	default:
		return d.GetMessageListByStartMsgId(sourceID, sessionType, startTime, endTime)
	}
}
func (d *DataBase) GetMessageListByStartMsgIdNoStartTimeController(sourceID string, sessionType int, endTime int64) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMessageListByStartMsgIdNoTime(sourceID, sessionType, endTime)
	default:
		return d.GetMessageListByMsgIdNoTime(sourceID, sessionType, endTime)
	}
}
// group ,index_recv_id and index_send_time only one can be used,when index_recv_id be used,temp B tree use for order by,Query speed decrease
func (d *DataBase) GetMessageList(sourceID string, sessionType, count int, startTime int64, isReverse bool) (result []*model_struct.LocalChatLog, err error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var messageList []model_struct.LocalChatLog
	var condition, timeOrder, timeSymbol, seqOrder string
	if isReverse {
		timeOrder = "send_time ASC"
		timeSymbol = ">"
		seqOrder = "seq ASC"
	} else {
		timeOrder = "send_time DESC"
		timeSymbol = "<"
		seqOrder = "seq DESC"
	}
	if sessionType == constant.SingleChatType && sourceID == d.loginUserID {
		condition = "send_id = ? And recv_id = ? AND status <=? And session_type = ? And send_time " + timeSymbol + " ?"
	} else {
		condition = "(send_id = ? OR recv_id = ?) AND status <=? And session_type = ? And send_time " + timeSymbol + " ?"
	}

	log.NewInfo("", utils.GetSelfFuncName(), condition, timeOrder, seqOrder)
	err = utils.Wrap(d.conn.Where(condition, sourceID, sourceID, constant.MsgStatusSendFailed, sessionType, startTime).
		Order(timeOrder).Offset(0).Limit(count).Find(&messageList).Error, "GetMessageList failed")
	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}
func (d *DataBase) GetMessageListController(sourceID string, sessionType, count int, startTime int64, isReverse bool) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMessageList(sourceID, sessionType, count, startTime, isReverse)
	default:
		return d.GetMessageList(sourceID, sessionType, count, startTime, isReverse)
	}
}
func (d *DataBase) GetMessageListNoTime(sourceID string, sessionType, count int, isReverse bool) (result []*model_struct.LocalChatLog, err error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var messageList []model_struct.LocalChatLog
	var condition, timeOrder, seqOrder string
	if isReverse {
		timeOrder = "send_time ASC"
		seqOrder = "seq ASC"
	} else {
		timeOrder = "send_time DESC"
		seqOrder = "seq DESC"
	}
	/*
		send_time DESC have issues when user fetch messages in 2nd page. first we return messages ASC order and then Desc, it return
		same message twice in it. told front end to call Reverse list for now
	*/

	log.NewInfo("", utils.GetSelfFuncName(), timeOrder, seqOrder)

	switch sessionType {
	case constant.SingleChatType:
		if sourceID == d.loginUserID {
			condition = "send_id = ? And recv_id = ? AND status <=? And session_type = ?"
		} else {
			condition = "(send_id = ? OR recv_id = ?) AND status <=? And session_type = ?"
		}
		err = utils.Wrap(d.conn.Where(condition, sourceID, sourceID, constant.MsgStatusSendFailed, sessionType).
			Order(timeOrder).Offset(0).Limit(count).Find(&messageList).Error, "GetMessageList failed")
	case constant.GroupChatType:
		condition = " recv_id = ? AND status <=? AND session_type = ?"
		err = utils.Wrap(d.conn.Where(condition, sourceID, constant.MsgStatusSendFailed, sessionType).
			Order(timeOrder).Offset(0).Limit(count).Find(&messageList).Error, "GetMessageList failed")
	default:
		condition = "(send_id = ? OR recv_id = ?) AND status <=? AND session_type = ?"
		err = utils.Wrap(d.conn.Where(condition, sourceID, sourceID, constant.MsgStatusSendFailed, sessionType).
			Order(timeOrder).Offset(0).Limit(count).Find(&messageList).Error, "GetMessageList failed")
	}

	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}
func (d *DataBase) GetMessageListByMsgIdNoTime(sourceID string, sessionType int, endTime int64) (result []*model_struct.LocalChatLog, err error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var condition, timeOrder, seqOrder string
	timeOrder = "send_time DESC"
	seqOrder = "seq DESC"
	/*
		send_time DESC have issues when user fetch messages in 2nd page. first we return messages ASC order and then Desc, it return
		same message twice in it. told front end to call Reverse list for now
	*/

	log.NewInfo("", utils.GetSelfFuncName(), timeOrder, seqOrder)

	switch sessionType {
	case constant.SingleChatType:
		if sourceID == d.loginUserID {
			condition = "send_id = ? And recv_id = ? AND status <=? And session_type = ? AND send_time > ?"
		} else {
			condition = "(send_id = ? OR recv_id = ?) AND status <=? And session_type = ? AND send_time > ?"
		}
		err = utils.Wrap(d.conn.Where(condition, sourceID, sourceID, constant.MsgStatusSendFailed, sessionType, endTime).
			Order(timeOrder).Find(&result).Error, "GetMessageList failed")
	case constant.GroupChatType:
		condition = " recv_id = ? AND status <=? AND session_type = ?"
		err = utils.Wrap(d.conn.Where(condition, sourceID, constant.MsgStatusSendFailed, sessionType, endTime).
			Order(timeOrder).Find(&result).Error, "GetMessageList failed")
	default:
		condition = "(send_id = ? OR recv_id = ?) AND status <=? AND session_type = ?"
		err = utils.Wrap(d.conn.Where(condition, sourceID, sourceID, constant.MsgStatusSendFailed, sessionType, endTime).
			Order(timeOrder).Find(&result).Error, "GetMessageList failed")
	}

	return result, err
}

func (d *DataBase) GetMessageListNoTimeController(sourceID string, sessionType, count int, isReverse bool) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMessageListNoTime(sourceID, sessionType, count, isReverse)
	default:
		return d.GetMessageListNoTime(sourceID, sessionType, count, isReverse)
	}
}

func (d *DataBase) GetSendingMessageList() (result []*model_struct.LocalChatLog, err error) {
	//d.mRWMutex.RLock()
	//defer d.mRWMutex.RUnlock()
	var messageList []model_struct.LocalChatLog
	err = utils.Wrap(d.conn.Where("status = ?", constant.MsgStatusSending).Find(&messageList).Error, "GetMessageList failed")
	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}
func (d *DataBase) UpdateSingleMessageHasRead(sendID string, msgIDList []string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Model(model_struct.LocalChatLog{}).Where("send_id=?  AND session_type=? AND client_msg_id in ?", sendID, constant.SingleChatType, msgIDList).Update("is_read", constant.HasRead)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateMessageStatusBySourceID failed")
}
func (d *DataBase) UpdateGroupMessageHasRead(msgIDList []string, sessionType int32) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Model(model_struct.LocalChatLog{}).Where("session_type=? AND client_msg_id in ?", sessionType, msgIDList).Update("is_read", constant.HasRead)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateMessageStatusBySourceID failed")
}
func (d *DataBase) UpdateGroupMessageHasReadController(msgIDList []string, groupID string, sessionType int32) error {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupUpdateGroupMessageHasRead(msgIDList, groupID)
	default:
		return d.UpdateGroupMessageHasRead(msgIDList, sessionType)
	}
}
func (d *DataBase) GetMultipleMessage(msgIDList []string) (result []*model_struct.LocalChatLog, err error) {
	var messageList []model_struct.LocalChatLog
	err = utils.Wrap(d.conn.Where("client_msg_id IN ?", msgIDList).Find(&messageList).Error, "GetMultipleMessage failed")
	for _, v := range messageList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}
func (d *DataBase) GetMultipleMessageController(msgIDList []string, groupID string, sessionType int32) (result []*model_struct.LocalChatLog, err error) {
	switch sessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMultipleMessage(msgIDList, groupID)
	default:
		return d.GetMultipleMessage(msgIDList)
	}
}

func (d *DataBase) GetNormalMsgSeq() (uint32, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var seq uint32
	var localSeqSynced *model_struct.LocalSeqSynced
	conversationID := utils.GetConversationIDBySessionType(d.loginUserID, constant.SingleChatType)
	err := d.conn.Model(model_struct.LocalSeqSynced{}).Where("conversation_id = ? ", conversationID).Find(&localSeqSynced).Error
	if err != nil {
		err := d.conn.Model(model_struct.LocalChatLog{}).Select("IFNULL(max(seq),0)").Find(&seq).Error
		return seq, utils.Wrap(err, "GetNormalMsgSeq")
	}
	if localSeqSynced != nil {
		seq = localSeqSynced.Seq
	}
	return seq, utils.Wrap(err, "GetNormalMsgSeq")

}
func (d *DataBase) GetLostMsgSeqList(minSeqInSvr uint32) ([]uint32, error) {
	var hasSeqList []uint32
	err := d.conn.Model(model_struct.LocalChatLog{}).Select("seq").Order("seq ASC").Find(&hasSeqList).Error
	if err != nil {
		return nil, err
	}
	seqLength := len(hasSeqList)
	if seqLength == 0 {
		return nil, nil
	}
	var normalSeqList []uint32

	var i uint32
	for i = 1; i <= hasSeqList[seqLength-1]; i++ {
		normalSeqList = append(normalSeqList, i)
	}
	lostSeqList := utils.DifferenceSubset(normalSeqList, hasSeqList)
	if len(lostSeqList) == 0 {
		return nil, nil
	}
	abnormalSeqList, err := d.GetAbnormalMsgSeqList()
	if err != nil {
		return nil, err
	}
	if len(abnormalSeqList) == 0 {
		return lostSeqList, nil
	}
	return utils.DifferenceSubset(lostSeqList, utils.Intersect(lostSeqList, abnormalSeqList)), nil

}
func (d *DataBase) GetSuperGroupNormalMsgSeq(groupID string) (uint32, error) {
	d.initSuperLocalChatLog(groupID)
	var seq uint32
	var localSeqSynced *model_struct.LocalSeqSynced
	conversationID := utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
	err := d.conn.Model(model_struct.LocalSeqSynced{}).Where("conversation_id = ? ", conversationID).Find(&localSeqSynced).Error
	if err != nil {
		err := d.conn.Table(utils.GetSuperGroupTableName(groupID)).Select("IFNULL(max(seq),0)").Find(&seq).Error
		return seq, utils.Wrap(err, "GetSuperGroupNormalMsgSeq")
	}
	if localSeqSynced != nil {
		seq = localSeqSynced.Seq
	}
	return seq, utils.Wrap(err, "GetSuperGroupNormalMsgSeq")
}

func (d *DataBase) GetTestMessage(seq uint32) (*model_struct.LocalChatLog, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var c model_struct.LocalChatLog
	return &c, utils.Wrap(d.conn.Where("seq = ?",
		seq).Find(&c).Error, "GetTestMessage failed")
}

func (d *DataBase) UpdateMsgSenderNickname(sendID, nickname string, sType int) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Where(
		"send_id = ? and session_type = ? and sender_nick_name != ? ", sendID, sType, nickname).Updates(
		map[string]interface{}{"sender_nick_name": nickname}).Error, utils.GetSelfFuncName()+" failed")
}

func (d *DataBase) UpdateMsgSenderFaceURL(sendID, faceURL string, sType int) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Where(
		"send_id = ? and session_type = ? and sender_face_url != ? ", sendID, sType, faceURL).Updates(
		map[string]interface{}{"sender_face_url": faceURL}).Error, utils.GetSelfFuncName()+" failed")
}
func (d *DataBase) UpdateMsgSenderFaceURLAndSenderNickname(sendID, faceURL, nickname string, sessionType int) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Where(
		"send_id = ? and session_type = ?", sendID, sessionType).Updates(
		map[string]interface{}{"sender_face_url": faceURL, "sender_nick_name": nickname}).Error, utils.GetSelfFuncName()+" failed")
}

func (d *DataBase) GetMsgSeqByClientMsgID(clientMsgID string) (uint32, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var seq uint32
	err := utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Select("seq").Where("client_msg_id=?", clientMsgID).First(&seq).Error, utils.GetSelfFuncName()+" failed")
	return seq, err
}

func (d *DataBase) GetMsgSeqByClientMsgIDController(m *sdk_struct.MsgStruct) (uint32, error) {
	switch m.SessionType {
	case constant.GroupChatType:
		return d.SuperGroupGetMsgSeqByClientMsgID(m.ClientMsgID, m.GroupID)
	default:
		return d.GetMsgSeqByClientMsgID(m.ClientMsgID)
	}
}

func (d *DataBase) GetMsgSeqListByGroupID(groupID string) ([]uint32, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var seqList []uint32
	err := utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Select("seq").Where("recv_id=?", groupID).Find(&seqList).Error, utils.GetSelfFuncName()+" failed")
	return seqList, err
}

func (d *DataBase) GetMsgSeqListByPeerUserID(userID string) ([]uint32, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var seqList []uint32
	err := utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Select("seq").Where("recv_id=? or send_id=?", userID, userID).Find(&seqList).Error, utils.GetSelfFuncName()+" failed")
	return seqList, err
}

func (d *DataBase) GetMsgSeqListBySelfUserID(userID string) ([]uint32, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var seqList []uint32
	err := utils.Wrap(d.conn.Model(model_struct.LocalChatLog{}).Select("seq").Where("recv_id=? and send_id=?", userID, userID).Find(&seqList).Error, utils.GetSelfFuncName()+" failed")
	return seqList, err
}

func (d *DataBase) InsertBroadcast(broadcast *model_struct.LocalBroadcast) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(broadcast).Error, "InsertMessage failed")
}

func (d *DataBase) GetBroadcastList(userID string) ([]model_struct.LocalBroadcast, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var broadcasts []model_struct.LocalBroadcast
	err := utils.Wrap(d.conn.Model(model_struct.LocalBroadcast{}).Where("sender_id=?", userID).Find(&broadcasts).Error, utils.GetSelfFuncName()+" failed")
	return broadcasts, err
}

func (d *DataBase) ClearBroadcastList(userID string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()

	err := utils.Wrap(d.conn.Model(model_struct.LocalBroadcastChatLog{}).Delete(&model_struct.LocalBroadcast{}).Error, utils.GetSelfFuncName()+" failed")
	return err
}

func (d *DataBase) InsertBroadcastMessage(Message *model_struct.LocalBroadcastChatLog) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(Message).Error, "InsertMessage failed")
}

func (d *DataBase) InsertLocalBroadcastMsgReceiver(msgReceiver []model_struct.LocalBroadcastMsgReceiver) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(msgReceiver).Error, "InsertMessage failed")
}

func (d *DataBase) GetBroadcastMessageReceiverListDB(clientMsgID string) ([]model_struct.BroadcastMessageList, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var broadcastMessageList []model_struct.BroadcastMessageList
	var broadcastsMsg model_struct.LocalBroadcastChatLog
	err := utils.Wrap(d.conn.Model(model_struct.LocalBroadcastChatLog{}).Where("client_msg_id = ? ", clientMsgID).Find(&broadcastsMsg).Error, utils.GetSelfFuncName()+" failed")
	if err == nil {
		temp := model_struct.BroadcastMessageList{LocalBroadcastChatLog: broadcastsMsg}
		var localBroadcastMsgReceiver []model_struct.LocalBroadcastMsgReceiver
		err := utils.Wrap(d.conn.Model(model_struct.LocalBroadcastMsgReceiver{}).Where("client_msg_id = ? ", broadcastsMsg.ClientMsgID).Find(&localBroadcastMsgReceiver).Error, utils.GetSelfFuncName()+" failed")
		if err == nil {
			temp.LocalBroadcastMsgReceiverList = append(temp.LocalBroadcastMsgReceiverList, localBroadcastMsgReceiver...)
		}
		broadcastMessageList = append(broadcastMessageList, temp)

	}
	return broadcastMessageList, err
}

func (d *DataBase) GetBroadcastMessageListDB(pageNumber int32) ([]model_struct.BroadcastMessageList, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var broadcastMessageList []model_struct.BroadcastMessageList
	var broadcastsMsgs []model_struct.LocalBroadcastChatLog
	err := utils.Wrap(d.conn.Model(model_struct.LocalBroadcastChatLog{}).Order("create_time desc").Offset(int(((pageNumber - 1) * 20))).Limit(20).Find(&broadcastsMsgs).Error, utils.GetSelfFuncName()+" failed")
	if err == nil {
		for _, msg := range broadcastsMsgs {
			temp := model_struct.BroadcastMessageList{LocalBroadcastChatLog: msg}
			var totalCountReceiver int64 = 0
			var localBroadcastMsgReceiver []model_struct.LocalBroadcastMsgReceiver
			err := utils.Wrap(d.conn.Model(model_struct.LocalBroadcastMsgReceiver{}).Where("client_msg_id = ? ", msg.ClientMsgID).Count(&totalCountReceiver).Limit(5).Find(&localBroadcastMsgReceiver).Error, utils.GetSelfFuncName()+" failed")
			if err == nil {
				temp.LocalBroadcastMsgReceiverList = append(temp.LocalBroadcastMsgReceiverList, localBroadcastMsgReceiver...)
			}
			temp.TotalReceiverCount = int32(totalCountReceiver)
			broadcastMessageList = append(broadcastMessageList, temp)
		}
	}
	return broadcastMessageList, err
}

func (d *DataBase) GetCallingMessages(searchKey string, errCode, startTime, endTime int64) ([]model_struct.LocalChatLog, error) {
	var friendList []string
	queryFriend := false
	if searchKey != "" {
		err := utils.Wrap(d.conn.Model(model_struct.LocalFriend{}).Select("friend_user_id").Where("name LIKE ?", "%"+searchKey+"%").Find(&friendList).Error, utils.GetSelfFuncName()+" failed")
		if err != nil {
			return nil, err
		}
		queryFriend = true
	}

	dbSub := d.conn.Model(model_struct.LocalChatLog{}).Where("content_type=?", constant.Calling)
	if startTime != 0 {
		dbSub = dbSub.Where("send_time>=?", startTime)
	}
	if endTime != 0 {
		dbSub = dbSub.Where("send_time<=?", endTime)
	}
	if queryFriend {
		dbSub = dbSub.Where("send_id IN (?) or recv_id IN (?)", friendList, friendList)
	}
	if errCode != 0 {
		dbSub = dbSub.Where("content LIKE ?", fmt.Sprintf(`%%err_code":%d%%`, errCode))
	}
	var result []model_struct.LocalChatLog
	err := utils.Wrap(dbSub.Order("send_time DESC").Find(&result).Error, utils.GetSelfFuncName()+" failed")
	if err != nil {
		return nil, err
	}
	return result, nil
}





