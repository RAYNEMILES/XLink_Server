package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"fmt"
	"strings"
	"time"
)

func GetChatLog(chatLog db.ChatLog, pageNumber, showNumber int32, orderBy string) ([]db.ChatLog, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var chatLogs []db.ChatLog
	if err != nil {
		return chatLogs, err
	}

	db := dbConn.Table("chat_logs").
		Where(fmt.Sprintf(" content like '%%%s%%'", chatLog.Content)).
		Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1)))
	if chatLog.SessionType != 0 {
		db = db.Where("session_type = ?", chatLog.SessionType)
	}
	if chatLog.ContentType != 0 {
		db = db.Where("content_type = ?", chatLog.ContentType)
	}
	if chatLog.SendID != "" {
		userIds := GetUserIdByCondition(map[string]string{"user_id": chatLog.SendID})
		if len(userIds) > 0 {
			db = db.Where("send_id IN ?", userIds)
		} else {
			db = db.Where("send_id = ?", chatLog.SendID)
		}
	}
	if chatLog.RecvID != "" {
		groupIds := GetGroupIdByCondition(map[string]string{"group_id": chatLog.RecvID})
		if len(groupIds) > 0 {
			db = db.Where("recv_id IN ?", groupIds)
		} else {
			db = db.Where("recv_id = ?", chatLog.RecvID)
		}
	}
	if chatLog.SendTime.Unix() > 0 {
		db = db.Where("send_time > ? and send_time < ?", chatLog.SendTime, chatLog.SendTime.AddDate(0, 0, 1))
	}
	if chatLog.Ex != "" {
		contextTypes := strings.Split(strings.Trim(chatLog.Ex, ","), ",")
		if len(contextTypes) > 0 {
			db = db.Where("content_type IN ?", contextTypes)
		}
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["create_time"] = "create_time"

	if orderBy != "" {
		direction := "DESC"
		sort := strings.Split(orderBy, ":")
		if len(sort) == 2 {
			if sort[1] == "asc" {
				direction = "ASC"
			}
		}
		col, ok := sortMap[sort[0]]
		if ok {
			orderByClause = fmt.Sprintf("%s %s ", col, direction)
		}

	}
	if orderByClause != "" {
		db = db.Order(orderByClause)
	}
	err = db.Find(&chatLogs).Error
	return chatLogs, err
}

func GetChatLogWithClientMsgID(clientMsgID string) (*db.ChatLog, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	chatLog := db.ChatLog{}
	if err != nil {
		return &chatLog, err
	}

	err = dbConn.Model(&chatLog).
		Where("client_msg_id=?", clientMsgID).First(&chatLog).Error
	return &chatLog, err
}

func SwitchChatLogStatusWithClientMsgID(clientMsgID string, status int) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	chatLog := db.ChatLog{}
	if err != nil {
		return false
	}
	err = dbConn.Model(&chatLog).
		Where("client_msg_id=?", clientMsgID).Update("status", status).Error
	if err != nil {
		return false
	}
	return true
}

func GetChatLogV1(chatLog db.ChatLog, pageNumber, showNumber int32, StartTime, EndTime time.Time) ([]db.ChatLog, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var chatLogs []db.ChatLog
	if err != nil {
		return chatLogs, err
	}

	db := dbConn.Table("chat_logs").
		Where(fmt.Sprintf(" content like '%%%s%%'", chatLog.Content)).
		Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1)))
	if chatLog.SessionType != 0 {
		db = db.Where("session_type = ?", chatLog.SessionType)
	}
	if chatLog.ContentType != 0 {
		db = db.Where("content_type = ?", chatLog.ContentType)
	}
	if chatLog.SendID != "" {
		db = db.Where("send_id = ?", chatLog.SendID)
	}
	if chatLog.RecvID != "" {
		db = db.Where("recv_id = ?", chatLog.RecvID)
	}
	if StartTime.Unix() > 0 {
		log.NewDebug("", utils.GetSelfFuncName(), StartTime)
		db = db.Where("send_time >= ?", StartTime)
	}
	if EndTime.Unix() > 0 {
		log.NewDebug("", utils.GetSelfFuncName(), EndTime)
		db = db.Where("send_time <= ?", EndTime)
	}
	if chatLog.Ex != "" {
		contextTypes := strings.Split(strings.Trim(chatLog.Ex, ","), ",")
		if len(contextTypes) > 0 {
			db = db.Where("content_type IN ?", contextTypes)
		}
	}

	err = db.Find(&chatLogs).Error
	return chatLogs, err
}

func GetChatLogCount(chatLog db.ChatLog) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var chatLogs []db.ChatLog
	var count int64
	if err != nil {
		return count, err
	}

	db := dbConn.Table("chat_logs").
		Where(fmt.Sprintf(" content like '%%%s%%'", chatLog.Content))
	if chatLog.SessionType != 0 {
		db = db.Where("session_type = ?", chatLog.SessionType)
	}
	if chatLog.ContentType != 0 {
		db = db.Where("content_type = ?", chatLog.ContentType)
	}
	if chatLog.SendID != "" {
		userIds := GetUserIdByCondition(map[string]string{"user_id": chatLog.SendID})
		if len(userIds) > 0 {
			db = db.Where("send_id IN ?", userIds)
		} else {
			db = db.Where("send_id = ?", chatLog.SendID)
		}
	}
	if chatLog.RecvID != "" {
		groupIds := GetGroupIdByCondition(map[string]string{"group_id": chatLog.RecvID})
		if len(groupIds) > 0 {
			db = db.Where("recv_id IN ?", groupIds)
		} else {
			db = db.Where("recv_id = ?", chatLog.RecvID)
		}
	}
	if chatLog.SendTime.Unix() > 0 {
		log.NewDebug("", utils.GetSelfFuncName(), chatLog.SendTime, chatLog.SendTime.AddDate(0, 0, 1))
		db = db.Where("send_time > ? and send_time < ?", chatLog.SendTime, chatLog.SendTime.AddDate(0, 0, 1))
	}
	if chatLog.Ex != "" {
		contextTypes := strings.Split(strings.Trim(chatLog.Ex, ","), ",")
		if len(contextTypes) > 0 {
			db = db.Where("content_type IN ?", contextTypes)
		}
	}

	err = db.Find(&chatLogs).Count(&count).Error
	return count, err
}

func GetChatLogCountV1(chatLog db.ChatLog, StartTime, EndTime time.Time) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var chatLogs []db.ChatLog
	var count int64
	if err != nil {
		return count, err
	}

	db := dbConn.Table("chat_logs").
		Where(fmt.Sprintf(" content like '%%%s%%'", chatLog.Content))
	if chatLog.SessionType != 0 {
		db = db.Where("session_type = ?", chatLog.SessionType)
	}
	if chatLog.ContentType != 0 {
		db = db.Where("content_type = ?", chatLog.ContentType)
	}
	if chatLog.SendID != "" {
		db = db.Where("send_id = ?", chatLog.SendID)
	}
	if chatLog.RecvID != "" {
		db = db.Where("recv_id = ?", chatLog.RecvID)
	}
	if StartTime.Unix() > 0 {
		log.NewDebug("", utils.GetSelfFuncName(), StartTime)
		db = db.Where("send_time >= ?", StartTime)
	}
	if EndTime.Unix() > 0 {
		log.NewDebug("", utils.GetSelfFuncName(), EndTime)
		db = db.Where("send_time <= ?", EndTime)
	}

	err = db.Find(&chatLogs).Count(&count).Error
	return count, err
}
