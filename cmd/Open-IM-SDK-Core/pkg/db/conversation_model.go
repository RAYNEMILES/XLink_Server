package db

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/utils"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

func (d *DataBase) GetConversationByUserID(userID string) (*model_struct.LocalConversation, error) {
	var conversation model_struct.LocalConversation
	err := utils.Wrap(d.conn.Where("user_id=?", userID).Find(&conversation).Error, "GetConversationByUserID error")
	return &conversation, err
}

func (d *DataBase) GetAllConversationList() ([]*model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var conversationList []model_struct.LocalConversation
	err := utils.Wrap(d.conn.Where("latest_msg_send_time > ?", 0).Order("case when is_pinned=1 then 0 else 1 end, case when is_pinned=1 then pinned_time else max(latest_msg_send_time,draft_text_time) end DESC").Find(&conversationList).Error,
		"GetAllConversationList failed")
	var transfer []*model_struct.LocalConversation
	for _, v := range conversationList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, err
}

func (d *DataBase) GetAllConversationListByType(conversationType []int32) ([]*model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var conversationList []model_struct.LocalConversation
	err := utils.Wrap(d.conn.Where("latest_msg_send_time > ? and conversation_type IN ?", 0, conversationType).Order("case when is_pinned=1 then 0 else 1 end, case when is_pinned=1 then pinned_time else max(latest_msg_send_time,draft_text_time) end DESC").Find(&conversationList).Error,
		"GetAllConversationList failed")
	var transfer []*model_struct.LocalConversation
	for _, v := range conversationList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, err
}

func (d *DataBase) GetHiddenConversationList() ([]*model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var conversationList []model_struct.LocalConversation
	err := utils.Wrap(d.conn.Where("latest_msg_send_time = ?", 0).Find(&conversationList).Error,
		"GetHiddenConversationList failed")
	var transfer []*model_struct.LocalConversation
	for _, v := range conversationList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, err
}

func (d *DataBase) GetAllConversationListToSync() ([]*model_struct.LocalConversation, error) {
	var conversationList []model_struct.LocalConversation
	err := utils.Wrap(d.conn.Find(&conversationList).Error, "GetAllConversationListToSync failed")
	var transfer []*model_struct.LocalConversation
	for _, v := range conversationList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, err
}

func (d *DataBase) GetConversationListSplit(offset, count int) ([]*model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var conversationList []model_struct.LocalConversation
	err := utils.Wrap(d.conn.Where("latest_msg_send_time > ?", 0).Order("case when is_pinned=1 then 0 else 1 end,max(latest_msg_send_time,draft_text_time) DESC").Offset(offset).Limit(count).Find(&conversationList).Error,
		"GetFriendList failed")
	var transfer []*model_struct.LocalConversation
	for _, v := range conversationList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, err
}
func (d *DataBase) BatchInsertConversationList(conversationList []*model_struct.LocalConversation) error {
	if conversationList == nil {
		return nil
	}
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(conversationList).Error, "BatchInsertConversationList failed")
}
func (d *DataBase) InsertConversation(conversationList *model_struct.LocalConversation) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Create(conversationList).Error, "InsertConversation failed")
}

func (d *DataBase) DeleteConversation(conversationID string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	return utils.Wrap(d.conn.Where("conversation_id = ?", conversationID).Delete(&model_struct.LocalConversation{}).Error, "DeleteConversation failed")
}

func (d *DataBase) GetConversation(conversationID string) (*model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var c model_struct.LocalConversation
	return &c, utils.Wrap(d.conn.Where("conversation_id = ?",
		conversationID).Take(&c).Error, "GetConversation failed")
}

func (d *DataBase) GetConversationsByType(conversationType int32) ([]model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var c []model_struct.LocalConversation
	return c, utils.Wrap(d.conn.Where("conversation_type = ?",
		conversationType).Find(&c).Error, "GetConversation failed")
}

func (d *DataBase) UpdateConversation(c *model_struct.LocalConversation) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateConversation failed")
}

func (d *DataBase) UpdateConversationForSync(c *model_struct.LocalConversation) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Model(&model_struct.LocalConversation{}).Where("conversation_id = ?", c.ConversationID).Updates(map[string]interface{}{"recv_msg_opt": c.RecvMsgOpt, "is_pinned": c.IsPinned, "pinned_time": c.PinnedTime, "is_private_chat": c.IsPrivateChat, "group_at_type": c.GroupAtType, "is_not_in_group": c.IsNotInGroup, "ex": c.Ex, "attached_info": c.AttachedInfo, "unread_count": c.UnreadCount})
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateConversation failed")
}
func (d *DataBase) UpdateConversationForGroupIfExists(c *model_struct.LocalConversation) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Model(&model_struct.LocalConversation{}).Where("conversation_id = ?", c.ConversationID).Updates(map[string]interface{}{"show_name": c.ShowName, "face_url": c.FaceURL, "is_not_in_group": false})
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateConversation failed")
}

func (d *DataBase) BatchUpdateConversationList(conversationList []*model_struct.LocalConversation) error {
	for _, v := range conversationList {
		err := d.UpdateConversation(v)
		if err != nil {
			return utils.Wrap(err, "BatchUpdateConversationList failed")
		}

	}
	return nil
}
func (d *DataBase) ConversationIfExists(conversationID string) (bool, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var count int64
	t := d.conn.Model(&model_struct.LocalConversation{}).Where("conversation_id = ?",
		conversationID).Count(&count)
	if t.Error != nil {
		return false, utils.Wrap(t.Error, "ConversationIfExists get failed")
	}
	if count != 1 {
		return false, nil
	} else {
		return true, nil
	}
}
func (d *DataBase) GetLocalMaxSeq() int64 {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var maxSeqLocal int64
	var localSeqSynced *model_struct.LocalSeqSynced
	conversationID := utils.GetConversationIDBySessionType(d.loginUserID, constant.SingleChatType)
	err := d.conn.Model(model_struct.LocalSeqSynced{}).Where("conversation_id = ? ", conversationID).Find(&localSeqSynced).Error
	if err != nil {
		row := d.conn.Model(&model_struct.LocalChatLog{}).Select("max(seq)").Row()
		row.Scan(&maxSeqLocal)
	}
	if localSeqSynced != nil {
		maxSeqLocal = int64(localSeqSynced.Seq)
	}
	return maxSeqLocal
}

func (d *DataBase) GetLocalErrMaxSeq() int64 {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var maxSeqLocal int64
	row := d.conn.Model(&model_struct.LocalErrChatLog{}).Select("max(seq)").Row()
	row.Scan(&maxSeqLocal)
	return maxSeqLocal
}

func (d *DataBase) GetGroupLocalMaxSeq(groupID string) int64 {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var maxSeqLocal int64
	var localSeqSynced *model_struct.LocalSeqSynced
	conversationID := utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
	err := d.conn.Model(model_struct.LocalSeqSynced{}).Where("conversation_id = ? ", conversationID).Find(&localSeqSynced).Error
	if err != nil {
		row := d.conn.Table(utils.GetSuperGroupTableName(groupID)).Select("max(seq)").Row()
		row.Scan(&maxSeqLocal)
	}
	if localSeqSynced != nil {
		maxSeqLocal = int64(localSeqSynced.Seq)
	}
	return maxSeqLocal
}

func (d *DataBase) GetGroupErrLocalMaxSeq(groupID string) int64 {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var maxSeqLocal int64
	row := d.conn.Table(utils.GetErrSuperGroupTableName(groupID)).Select("max(seq)").Row()
	row.Scan(&maxSeqLocal)
	return maxSeqLocal
}

// Reset the conversation is equivalent to deleting the conversation,
// and the GetAllConversation or GetConversationListSplit interface will no longer be obtained.
func (d *DataBase) ResetConversation(conversationID string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{ConversationID: conversationID, UnreadCount: 0, LatestMsg: "", LatestMsgSendTime: 0, DraftText: "", DraftTextTime: 0}
	t := d.conn.Select("unread_count", "latest_msg", "latest_msg_send_time", "draft_text", "draft_text_time").Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "ResetConversation failed")
}

// Reset ALL conversation is equivalent to deleting the conversation,
// and the GetAllConversation or GetConversationListSplit interface will no longer be obtained.
func (d *DataBase) ResetAllConversation() error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{UnreadCount: 0, LatestMsg: "", LatestMsgSendTime: 0, DraftText: "", DraftTextTime: 0}
	t := d.conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Select("unread_count", "latest_msg", "latest_msg_send_time", "draft_text", "draft_text_time").Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "ResetConversation failed")
}

// Clear the conversation, which is used to delete the conversation history message and clear the conversation at the same time.
// The GetAllConversation or GetConversationListSplit interface can still be obtained,
// but there is no latest message.
func (d *DataBase) ClearConversation(conversationID string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{ConversationID: conversationID, UnreadCount: 0, LatestMsg: "", DraftText: "", DraftTextTime: 0}
	t := d.conn.Select("unread_count", "latest_msg", "draft_text", "draft_text_time").Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "ClearConversation failed")
}

// Clear All conversation, which is used to delete the conversation history message and clear the conversation at the same time.
// The GetAllConversation or GetConversationListSplit interface can still be obtained,
// but there is no latest message.
func (d *DataBase) CleaAllConversation() error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{UnreadCount: 0, LatestMsg: "", DraftText: "", DraftTextTime: 0}
	t := d.conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Select("unread_count", "latest_msg", "draft_text", "draft_text_time").Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "ClearConversation failed")
}

func (d *DataBase) MoveConversationToArchived(lc *model_struct.LocalConversation) error {
	lcArchived := model_struct.LocalArchivedConversation{}
	utils2.CopyStructFields(&lcArchived, lc)
	t := d.conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Updates(lcArchived)
	if t.RowsAffected == 0 {
		err := t.Create(&lcArchived).Error
		return utils.Wrap(err, "InsertConversation failed")
	}
	return utils.Wrap(t.Error, "ClearConversation failed")
}

func (d *DataBase) GetArchivedConversation(conversationID string) (*model_struct.LocalArchivedConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var c model_struct.LocalArchivedConversation
	return &c, utils.Wrap(d.conn.Where("conversation_id = ?",
		conversationID).Take(&c).Error, "GetConversation failed")
}

func (d *DataBase) SetConversationDraft(conversationID, draftText string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	nowTime := utils.GetCurrentTimestampByMill()
	t := d.conn.Exec("update local_conversations set draft_text=?,draft_text_time=?,latest_msg_send_time=case when latest_msg_send_time=? then ? else latest_msg_send_time  end where conversation_id=?",
		draftText, nowTime, 0, nowTime, conversationID)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "SetConversationDraft failed")
}
func (d *DataBase) RemoveConversationDraft(conversationID, draftText string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{ConversationID: conversationID, DraftText: draftText, DraftTextTime: 0}
	t := d.conn.Select("draft_text", "draft_text_time").Updates(c)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "RemoveConversationDraft failed")
}
func (d *DataBase) UnPinConversation(conversationID string, isPinned int) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Exec("update local_conversations set is_pinned=?,draft_text_time=case when draft_text=? then ? else draft_text_time  end where conversation_id=?",
		isPinned, "", 0, conversationID)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UnPinConversation failed")
}

func (d *DataBase) UpdateColumnsConversation(conversationID string, args map[string]interface{}) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{ConversationID: conversationID}
	t := d.conn.Model(&c).Updates(args)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateColumnsConversation failed")
}
func (d *DataBase) UpdateAllConversation(conversation *model_struct.LocalConversation) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	if conversation.ConversationID != "" {
		return utils.Wrap(errors.New("not update all conversation"), "UpdateAllConversation failed")
	}
	t := d.conn.Model(conversation).Updates(conversation)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "UpdateColumnsConversation failed")
}
func (d *DataBase) IncrConversationUnreadCount(conversationID string) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	c := model_struct.LocalConversation{ConversationID: conversationID}
	t := d.conn.Model(&c).Update("unread_count", gorm.Expr("unread_count+?", 1))
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "IncrConversationUnreadCount failed")
}
func (d *DataBase) GetTotalUnreadMsgCount() (totalUnreadCount int32, err error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var result []int64
	err = d.conn.Model(&model_struct.LocalConversation{}).Where("recv_msg_opt < ?", constant.ReceiveNotNotifyMessage).Pluck("unread_count", &result).Error
	if err != nil {
		return totalUnreadCount, utils.Wrap(errors.New("GetTotalUnreadMsgCount err"), "GetTotalUnreadMsgCount err")
	}
	for _, v := range result {
		totalUnreadCount += int32(v)
	}
	return totalUnreadCount, nil
}

func (d *DataBase) SetMultipleConversationRecvMsgOpt(conversationIDList []string, opt int) (err error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	t := d.conn.Model(&model_struct.LocalConversation{}).Where("conversation_id IN ?", conversationIDList).Updates(map[string]interface{}{"recv_msg_opt": opt})
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "SetMultipleConversationRecvMsgOpt failed")
}

func (d *DataBase) GetMultipleConversation(conversationIDList []string) (result []*model_struct.LocalConversation, err error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var conversationList []model_struct.LocalConversation
	err = utils.Wrap(d.conn.Where("conversation_id IN ?", conversationIDList).Find(&conversationList).Error, "GetMultipleConversation failed")
	for _, v := range conversationList {
		v1 := v
		result = append(result, &v1)
	}
	return result, err
}

func (d *DataBase) SetForwardMessagesToConversation(receiverID string, receiverType int64) error {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	query := "update local_conversations set  last_forwarded_message_time = " + strconv.FormatInt(time.Now().Unix(), 10) + " where user_id=?"
	if receiverType == 2 {
		query = "update local_conversations set last_forwarded_message_time = " + strconv.FormatInt(time.Now().Unix(), 10) + " where group_id=?"
	}
	log.Error("SetForwardMessagesToConversation", query)
	t := d.conn.Exec(query, receiverID)
	if t.RowsAffected == 0 {
		return utils.Wrap(errors.New("RowsAffected == 0"), "no update")
	}
	return utils.Wrap(t.Error, "SetConversationDraft failed")
}

func (d *DataBase) GetRecentForwardMessagesConversationList() ([]*model_struct.LocalConversation, error) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	var conversationList []model_struct.LocalConversation
	err := utils.Wrap(d.conn.Where("last_forwarded_message_time > ?", 0).Order("last_forwarded_message_time DESC").Limit(6).Find(&conversationList).Error,
		"GetAllConversationList failed")
	var transfer []*model_struct.LocalConversation
	for _, v := range conversationList {
		v1 := v
		transfer = append(transfer, &v1)
	}
	return transfer, err
}
func (d *DataBase) UpdateLocalSeqNumber(localSeqSynced *model_struct.LocalSeqSynced) {
	//d.mRWMutex.Lock()
	//defer d.mRWMutex.Unlock()
	recordUpdated := d.conn.Updates(localSeqSynced).RowsAffected
	if recordUpdated == 0 {
		d.conn.Create(localSeqSynced)
	}
}
