package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"fmt"
	"strings"
)

func ChannelCodeIsExpired(code string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}
	inviteChannelCode := db.InviteChannelCode{}
	err1 := dbConn.Model(&inviteChannelCode).Where("code=?", code).First(&inviteChannelCode).Error
	if err1 != nil {
		return false
	}
	if inviteChannelCode == (db.InviteChannelCode{}) {
		return false
	}
	return true
}

func AddInviteChannelCode(inviteChannelCode *db.InviteChannelCode) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Create(inviteChannelCode).Error
	if err != nil {
		return err
	}
	return nil
}

func EditChannelCodeByCode(code string, data db.InviteChannelCode) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Where("code=?", code).Updates(map[string]interface{}{
		"code":      data.Code,
		"group_id":  data.GroupId,
		"friend_id": data.FriendId,
		"note":      data.Note,
		"greeting":  data.Greeting,
		"source_id": data.SourceId,
	}).Error
	if err != nil {
		return false
	}
	return true
}

func SwitchInviteChannelCodeState(inviteChannelCode *db.InviteChannelCode) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Where("code=?", inviteChannelCode.Code).Updates(map[string]interface{}{"operator_user_id": inviteChannelCode.OperatorUserId, "state": inviteChannelCode.State}).Error
	if err != nil {
		return err
	}
	return nil
}

func MultiDeleteInviteChannelCode(codeList []string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}
	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Where("code IN ?", codeList).Update("state", constant.InviteChannelCodeStateDelete).Error
	if err != nil {
		return false
	}
	return true
}

func GetTotalInviteChannelCodeCount(where map[string]string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64
	subQuery := dbConn.Table(db.InviteChannelCode{}.TableName()).Where("state IN ?", []int{constant.InviteChannelCodeStateValid, constant.InviteChannelCodeStateInvalid})
	if code, ok := where["code"]; ok {
		if code != "" {
			subQuery = subQuery.Where("code=?", where["code"])
		}
	}
	if friendId, ok := where["friend_id"]; ok {
		if friendId != "" {
			subQuery = subQuery.Where("friend_id like ?", "%"+where["friend_id"]+"%")
		}
	}
	if groupId, ok := where["group_id"]; ok {
		if groupId != "" {
			subQuery = subQuery.Where("group_id like ?", "%"+where["group_id"]+"%")
		}
	}
	if state, ok := where["state"]; ok {
		if state != "" {
			subQuery = subQuery.Where("state=?", where["state"])
		}
	}
	if note, ok := where["note"]; ok {
		if note != "" {
			subQuery = subQuery.Where("note like ?", "%"+where["note"]+"%")
		}
	}
	if sourceId, ok := where["source_id"]; ok {
		if sourceId != "" {
			subQuery = subQuery.Where("source_id=?", sourceId)
		}
	}

	dbError := subQuery.Count(&count).Error
	if dbError != nil {
		return 0, dbError
	}
	return count, nil
}

func GetInviteChannelCodeList(where map[string]string, PageNumber, ShowNumber int, orderBy string) ([]*db.InviteChannelCode, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	dbConn = dbConn.Debug()
	if err != nil {
		return nil, err
	}
	inviteChannelCodeList := make([]*db.InviteChannelCode, 0)
	subQuery := dbConn.Table(db.InviteChannelCode{}.TableName()).
		Where("state IN ?", []int{constant.InviteChannelCodeStateValid, constant.InviteChannelCodeStateInvalid})

	if code, ok := where["code"]; ok {
		if code != "" {
			subQuery = subQuery.Where("code=?", where["code"])
		}

	}
	if friendId, ok := where["friend_id"]; ok {
		if friendId != "" {
			subQuery = subQuery.Where("friend_id like ?", "%"+where["friend_id"]+"%")
		}
	}
	if groupId, ok := where["group_id"]; ok {
		if groupId != "" {
			subQuery = subQuery.Where("group_id like ?", "%"+where["group_id"]+"%")
		}
	}
	if state, ok := where["state"]; ok {
		if state != "" {
			subQuery = subQuery.Where("state=?", where["state"])
		}
	}
	if note, ok := where["note"]; ok {
		if note != "" {
			subQuery = subQuery.Where("note like ?", "%"+where["note"]+"%")
		}
	}
	if sourceId, ok := where["source_id"]; ok {
		if sourceId != "" {
			subQuery = subQuery.Where("source_id=?", sourceId)
		}
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["id"] = "id"

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
		subQuery = subQuery.Order(orderByClause)
	}
	dbErr := subQuery.Limit(ShowNumber).Offset((PageNumber - 1) * ShowNumber).Find(&inviteChannelCodeList).Error
	if dbErr != nil {
		return nil, err
	}
	return inviteChannelCodeList, nil
}

func GetInviteChannelCodeByCode(code string) (*db.InviteChannelCode, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	inviteChannelCode := db.InviteChannelCode{}
	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Where("code=?", code).Where("state IN ?", []int{constant.InviteChannelCodeStateValid, constant.InviteChannelCodeStateInvalid}).First(&inviteChannelCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteChannelCode, nil
}

func GetInviteChannelAllCodeByCode(code string) (*db.InviteChannelCode, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	inviteChannelCode := db.InviteChannelCode{}
	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Where("code=?", code).First(&inviteChannelCode).Error
	if err != nil {
		return nil, err
	}
	return &inviteChannelCode, nil
}

func ChannelCodeMultiSet(code []string, state string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	dbConn.Model(&db.InviteChannelCode{}).Where("code in ?", code).Update("state", state)
	return true
}

func GetChannelCodeListByOfficial() []*db.InviteChannelCode {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}
	inviteChannelCodeList := make([]*db.InviteChannelCode, 0)
	dbConn.Table(db.InviteChannelCode{}.TableName()).
		Where("source_id = ?", constant.UserRegisterSourceTypeOfficial).
		Where("state IN ?", []int{constant.InviteChannelCodeStateValid, constant.InviteChannelCodeStateInvalid}).Find(&inviteChannelCodeList)
	return inviteChannelCodeList
}

func GetChannelCodeListByOfficialAndState(state int) []*db.InviteChannelCode {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}
	inviteChannelCodeList := make([]*db.InviteChannelCode, 0)
	dbConn.Table(db.InviteChannelCode{}.TableName()).
		Where("source_id = ?", constant.UserRegisterSourceTypeOfficial).
		Where("state = ?", state).Find(&inviteChannelCodeList)
	return inviteChannelCodeList
}

func IsExistOfficialChannelCode() bool {
	officialChannelCodeList := GetChannelCodeListByOfficial()
	if len(officialChannelCodeList) > 0 {
		return true
	}
	return false
}

func GetOfficialChannelCode() *db.InviteChannelCode {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}
	inviteChannelCode := db.InviteChannelCode{}
	err = dbConn.Table(db.InviteChannelCode{}.TableName()).Where("source_id = ?", constant.UserRegisterSourceTypeOfficial).Where("state IN ?", []int{constant.InviteChannelCodeStateValid, constant.InviteChannelCodeStateInvalid}).First(&inviteChannelCode).Error
	if err != nil {
		return nil
	}
	return &inviteChannelCode
}
