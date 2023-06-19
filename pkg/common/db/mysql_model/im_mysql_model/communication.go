package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func StartCommunication(record *db.VideoAudioCommunicationRecord) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	record.StartTime = time.Now().Unix()
	err = dbConn.Table(record.TableName()).Create(record).Error

	return err
}

func GetActiveCommunicationByRoomID(roomId string, roomIdType int64) (db.VideoAudioCommunicationRecord, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var record = db.VideoAudioCommunicationRecord{}
	if err != nil {
		return record, err
	}

	err = dbConn.Table(db.VideoAudioCommunicationRecord{}.TableName()).
		Where("room_id=? AND room_id_type=?", roomId, roomIdType).
		Where("status=1 OR status=2").Debug().First(&record).Error

	return record, err
}

func UpdateCommunication(record *db.VideoAudioCommunicationRecord) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table(record.TableName()).Where("communication_id=?", record.CommunicationID).Updates(record).Error

	return err
}

func InsertCommunicationMembers(communicationID int64, mID []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	gMember := &db.CommunicationGroupMember{CommunicationID: communicationID}
	for _, m := range mID {
		gMember.MemberID = m
		err = dbConn.Table(gMember.TableName()).Create(gMember).Error
		gMember.Id = 0
	}

	return err
}

func GetCommunicationsByWhere(where map[string]string, communicationType, pageNumber, showNumber int32, orderBy string) ([]*db.VideoAudioCommunicationRecordRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var count int64
	if err != nil {
		return nil, count, err
	}

	var originatorUserList []db.User
	var originatorUserIdList []string
	var originatorUserMap = make(map[string]*db.User, 0)

	var memberUserList []db.User
	var memberUserIdList []string

	var communicationIDList []int64
	var communicationMemberUserMapList = make(map[int64][]*db.User, 0)

	sortMap := map[string]string{}
	timeTypeMap := map[int]string{}
	var orderByClause string
	sortMap["start_time"] = "start_time"
	timeTypeMap[1] = "start_time"
	timeTypeMap[2] = "end_time"

	// originator user list
	originatorQuery := false
	if originator, ok := where["originator"]; ok {
		if originator != "" {
			originatorUserList = GetUserByAllCondition(originator)
			originatorQuery = true
			for index, originatorUser := range originatorUserList {
				originatorUserIdList = append(originatorUserIdList, originatorUser.UserID)
				originatorUserMap[originatorUser.UserID] = &originatorUserList[index]
			}
		}
	}

	// member user list
	memberUserQuery := false
	if member, ok := where["member"]; ok {
		if member != "" {
			memberUserList = GetUserByAllCondition(member)
			for _, memberUser := range memberUserList {
				memberUserIdList = append(memberUserIdList, memberUser.UserID)
			}
			memberUserQuery = true
		}
	}

	// call id list
	callIDQuery := false
	if memberUserQuery {
		dbConn.Table(db.CommunicationGroupMember{}.TableName()).Select("distinct communication_id").Where("member_id IN (?)", memberUserIdList).Find(&communicationIDList)
		callIDQuery = true
	}

	// query record
	dbSub := dbConn.Table(db.VideoAudioCommunicationRecord{}.TableName()).Where("delete_time=0")
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

	if communicationType == 1 {
		// group
		dbSub = dbSub.Where("group_id=room_id")
	}
	if originatorPlatform, ok := where["originator_platform"]; ok {
		if originatorPlatform != "" && originatorPlatform != "0" {
			dbSub = dbSub.Where("originator_platform=?", originatorPlatform)
		}
	}
	if chatType, ok := where["chat_type"]; ok {
		if chatType != "" && chatType != "0" {
			dbSub = dbSub.Where("chat_type=?", chatType)
		}
	}
	if duration, ok := where["duration"]; ok {
		if duration != "" && duration != "0" {
			dbSub = dbSub.Where("duration=?", duration)
		}
	}
	if status, ok := where["status"]; ok {
		if status != "" && status != "0" {
			dbSub = dbSub.Where("status=?", status)
		}
	}
	if remark, ok := where["remark"]; ok {
		if remark != "" {
			dbSub = dbSub.Where("remark LIKE ?", "%"+remark+"%")
		}
	}
	if roomId, ok := where["room_id"]; ok {
		if roomId != "" {
			dbSub = dbSub.Where("room_id LIKE ?", "%"+roomId+"%")
		}
	}
	if originatorQuery {
		dbSub = dbSub.Where("originator IN (?)", originatorUserIdList)
	}
	if callIDQuery {
		dbSub = dbSub.Where("communication_id IN (?)", communicationIDList)
	}

	startTime := where["start_time"]
	endTime := where["end_time"]
	timeTypeStr := where["time_type"]
	timeType, _ := strconv.Atoi(timeTypeStr)

	if timeType == 0 {
		if startTime != "" || endTime != "" {
			var sql = ""
			var paramList []interface{}
			isFirst := true
			for _, v := range timeTypeMap {
				beenOr := false
				if startTime != "" {
					if isFirst {
						sql += fmt.Sprintf("%s>=?", v)
						paramList = append(paramList, startTime)
					} else {
						sql += fmt.Sprintf(" OR %s>=?", v)
						paramList = append(paramList, startTime)
						beenOr = true
					}
				}
				if endTime != "" {
					if isFirst || beenOr {
						sql += fmt.Sprintf(" AND %s<=?", v)
						paramList = append(paramList, endTime)
					} else {
						sql += fmt.Sprintf(" OR %s<=?", v)
						paramList = append(paramList, endTime)
					}
				}
				isFirst = false
			}

			log.Debug("", "time sql: ", sql, " param: ", paramList)
			dbSub = dbSub.Where(sql, paramList...)
		}
	} else {
		if startTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s>=?", timeTypeMap[timeType]), startTime)
		}
		if endTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s<=?", timeTypeMap[timeType]), endTime)
		}
	}

	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}

	var recordSQL []db.VideoAudioCommunicationRecord

	dbSub.Debug().Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&recordSQL)
	if !originatorQuery || !callIDQuery {
		for _, record := range recordSQL {
			if !originatorQuery {
				originatorUserIdList = append(originatorUserIdList, record.Originator)
			}
		}
		if !originatorQuery {
			dbConn.Table(db.User{}.TableName()).Where("user_id IN (?)", originatorUserIdList).Find(&originatorUserList)
			for index, originatorUser := range originatorUserList {
				originatorUserMap[originatorUser.UserID] = &originatorUserList[index]
			}
		}
	}
	var memberRelationList []db.CommunicationGroupMember
	var memberIdList []string
	var memberUserMap = make(map[string]*db.User, 0)
	dbSub = dbConn.Table(db.CommunicationGroupMember{}.TableName())
	fmt.Println("callIDQuery::", callIDQuery)
	if callIDQuery {
		dbSub = dbSub.Where("communication_id IN (?)", communicationIDList)
	}
	dbSub.Debug().Find(&memberRelationList)
	fmt.Println("", "memberRelationList: ", memberRelationList)
	for _, relation := range memberRelationList {
		memberIdList = append(memberIdList, relation.MemberID)
	}
	dbConn.Table(db.User{}.TableName()).Where("user_id IN (?)", memberIdList).Find(&memberUserList)
	fmt.Println("", "memberIdList: ", memberIdList)
	fmt.Println("", "memberUserList: ", memberUserList)
	for index, user := range memberUserList {
		memberUserMap[user.UserID] = &memberUserList[index]
	}
	for _, relation := range memberRelationList {
		if _, ok := communicationMemberUserMapList[relation.CommunicationID]; !ok {
			communicationMemberUserMapList[relation.CommunicationID] = make([]*db.User, 0)
		}
		if _, ok := memberUserMap[relation.MemberID]; ok {
			communicationMemberUserMapList[relation.CommunicationID] = append(communicationMemberUserMapList[relation.CommunicationID], memberUserMap[relation.MemberID])
		}
	}
	fmt.Println("", "memberUserMap: ", memberUserMap)
	fmt.Println("", "communicationMemberUserMapList: ", communicationMemberUserMapList)
	var result []*db.VideoAudioCommunicationRecordRes
	for _, record := range recordSQL {
		recordRes := &db.VideoAudioCommunicationRecordRes{}
		err = utils.CopyStructFields(recordRes, record)
		if err != nil {
			log.Error("", "copy result error: ", err.Error())
			return nil, count, err
		}
		if user, ok := originatorUserMap[recordRes.Originator]; ok {
			recordRes.OriginatorName = user.Nickname
		}
		fmt.Println("", "recordRes.CommunicationID: ", recordRes.CommunicationID)
		if mList, ok := communicationMemberUserMapList[recordRes.CommunicationID]; ok {
			for _, user := range mList {
				recordRes.MemberIDs = append(recordRes.MemberIDs, user.UserID)
				recordRes.MemberIDNames = append(recordRes.MemberIDNames, user.Nickname)
			}
		}

		result = append(result, recordRes)
	}

	return result, count, nil

}

func DeleteCommunications(communicatID []int64, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	record := &db.VideoAudioCommunicationRecord{DeleteBy: opUserID, DeleteTime: time.Now().Unix()}
	for _, recordID := range communicatID {
		err = dbConn.Table(db.VideoAudioCommunicationRecord{}.TableName()).Where("communication_id=?", recordID).Updates(record).Error
	}

	return err
}

func UpdateCommunicationRecordByActiveRoomID(record *db.VideoAudioCommunicationRecord) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table(record.TableName()).Where("room_id=? AND  room_id_type=? AND (status=1 OR status=2)", record.RoomID, record.RoomIDType).Updates(record).Error

	return err
}

func GetNoUploadRecord(roomId uint64, roomIDType int64, callbackTs uint64) (db.VideoAudioCommunicationRecord, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var record = db.VideoAudioCommunicationRecord{}
	if err != nil {
		return record, err
	}

	err = dbConn.Table(record.TableName()).
		Where("room_id=? AND  room_id_type=? AND (status=1 OR status=2) AND record_task_id=? AND record_url=''", roomId, callbackTs).
		Find(&record).Error

	return record, err
}

func UpdateCallingMember(communicationId int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	//dbConn.Clauses(clause.Locking{Strength: "UPDATE", Options: "NOWAIT"}).Find()
	var communicationID int64
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		// .Set("gorm:query_option", "FOR UPDATE")
		communicat := db.VideoAudioCommunicationRecord{}
		sql := fmt.Sprintf("SELECT * FROM `video_audio_communication_record` WHERE communication_id=? LIMIT 1 lock in share mode")

		//if err = tx.Set("gorm:query_option", "FOR UPDATE").Where("room_id=? AND room_id_type=? AND (status=1 OR status=2)", roomID, roomIdType).Debug().First(&communicat).Error; err != nil {
		//	return err
		//}
		if err = tx.Raw(sql, communicationId).Debug().First(&communicat).Error; err != nil {
			return err
		}

		specifyMap := make(map[string]interface{})
		specifyMap["status"] = gorm.Expr("3")
		if err = tx.Table(communicat.TableName()).Where("communication_id=?", communicat.CommunicationID).Debug().Updates(specifyMap).Error; err != nil {
			return err
		}
		communicationID = communicat.CommunicationID

		return nil
	})
	if err != nil {
		return communicationID, err
	}

	return communicationID, nil

}

func GetCommunicationById(communicatID int64) (db.VideoAudioCommunicationRecord, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	record := db.VideoAudioCommunicationRecord{}
	if err != nil {
		return record, err
	}

	res := dbConn.Table(record.TableName()).Where("communication_id=?", communicatID).First(&record)

	return record, res.Error
}

func GetCommunicationMemberByCommunicationID(communicatID int64) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var members []string
	res := dbConn.Table("communication_group_member").Select("member_id").Where("communication_id=?", communicatID).Find(&members)

	return members, res.Error
}
