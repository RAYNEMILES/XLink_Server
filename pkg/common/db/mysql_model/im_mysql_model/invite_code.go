package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"fmt"
	"reflect"
	"strings"
)

func CodeIsExpired(code string) bool {
	if code == "" {
		return false
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	inviteCode := db.InviteCode{}
	err1 := dbConn.Model(&inviteCode).Where("code=?", code).First(&inviteCode).Error
	if err1 != nil {
		return false
	}
	if inviteCode == (db.InviteCode{}) {
		return false
	}
	return true
}

func GetCodeByUserID(userID string) (*db.InviteCode, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	inviteCode := db.InviteCode{}
	dbConn.Model(&inviteCode).Where("user_id=?", userID).First(&inviteCode)

	return &inviteCode, nil
}

func GetCodeInfoByCode(code string) *db.InviteCode {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	inviteCode := db.InviteCode{}
	err = dbConn.Table("invite_code").Where("code=?", code).First(&inviteCode).Error

	if err != nil {
		return nil
	}
	return &inviteCode
}

func AddCode(userId, code, greeting, note string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err1 := dbConn.Debug().Table(db.InviteCode{}.TableName()).Create(&db.InviteCode{
		UserId:   userId,
		Code:     code,
		Greeting: greeting,
		Note:     note,
		State:    constant.InviteCodeStateValid,
	}).Error

	if err1 != nil {
		return false
	}
	return true
}

func SwitchCodeState(code string, state int) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err1 := dbConn.Debug().Table(db.InviteCode{}.TableName()).Where("code=?", code).Update("state", state).Error

	if err1 != nil {
		return false
	}
	return true
}

func MultiDeleteCode(codes []string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err = dbConn.Debug().Table(db.InviteCode{}.TableName()).Where("code IN ?", codes).Update("state", constant.InviteCodeStateDelete).Error
	if err != nil {
		return false
	}
	return true
}

func EditInviteCodeByCode(code string, data db.InviteCode) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	err = dbConn.Debug().Table(db.InviteCode{}.TableName()).Where("code=?", code).Updates(map[string]interface{}{
		"code":     data.Code,
		"greeting": data.Greeting,
		"note":     data.Note,
	}).Error
	if err != nil {
		return false
	}
	return true
}

func GetCodeList(where map[string]string, showNumber, pageNumber int, orderBy string) ([]db.InviteCode, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var inviteCode []db.InviteCode
	dbQuery := dbConn.Table("invite_code").Debug().Where("state IN ?", []int{constant.InviteCodeStateValid, constant.InviteCodeStateInvalid}).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1)))

	if code, ok := where["code"]; ok {
		if code != "" {
			dbQuery = dbQuery.Where("code=?", where["code"])
		}

	}
	if state, ok := where["state"]; ok {
		if state != "" && state != "0" {
			dbQuery = dbQuery.Where("state=?", where["state"])
		}
	}
	if note, ok := where["note"]; ok {
		if note != "" {
			dbQuery = dbQuery.Where("note like ?", "%"+where["note"]+"%")
		}
	}
	if userId, ok := where["user_id"]; ok {
		if userId != "" {
			dbQuery = dbQuery.Where("userId = ?", where["user_id"])
		}
	}
	if userName, ok := where["user_name"]; ok {
		if userName != "" {
			userInfo, _ := GetUserByNameOrUserId(userName, -1, 1)
			if len(userInfo) > 0 {
				userId, _ := arrayColumn(userInfo, "UserID")
				dbQuery = dbQuery.Where("user_id in ?", arrayKey(userId))
			} else {
				dbQuery = dbQuery.Where("user_id = ?", "000")
			}
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
		dbQuery = dbQuery.Order(orderByClause)
	}
	err = dbQuery.Find(&inviteCode).Error
	if err != nil {
		return nil, err
	}
	if inviteCode == nil {
		return nil, err
	}
	return inviteCode, nil
}

func GetCodeNums(where map[string]string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	dbQuery := dbConn.Table("invite_code").Where("state IN ?", []int{constant.InviteCodeStateValid, constant.InviteCodeStateInvalid})

	if code, ok := where["code"]; ok {
		if code != "" {
			dbQuery = dbQuery.Where("code=?", where["code"])
		}

	}
	if state, ok := where["state"]; ok {
		if state != "" && state != "0" {
			dbQuery = dbQuery.Where("state=?", where["state"])
		}
	}
	if note, ok := where["note"]; ok {
		if note != "" {
			dbQuery = dbQuery.Where("note like ?", "%"+where["note"]+"%")
		}
	}
	if userId, ok := where["user_id"]; ok {
		if userId != "" {
			dbQuery = dbQuery.Where("userId = ?", where["user_id"])
		}
	}
	if userName, ok := where["user_name"]; ok {
		if userName != "" {
			userInfo, _ := GetUserByNameOrUserId(userName, -1, 1)
			if len(userInfo) > 0 {
				userId, _ := arrayColumn(userInfo, "UserID")
				dbQuery = dbQuery.Where("user_id in ?", arrayKey(userId))
			} else {
				dbQuery = dbQuery.Where("user_id = ?", "000")
			}
		}
	}
	dbErr := dbQuery.Count(&count).Error

	if dbErr != nil {
		return 0, dbErr
	}
	return count, nil
}

func InviteCodeMultiSet(code []string, state string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	dbConn.Model(&db.InviteCode{}).Where("code IN ?", code).Update("state", state)
	return true
}

func arrayColumn(array interface{}, key string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	t := reflect.TypeOf(array)
	v := reflect.ValueOf(array)
	if t.Kind() != reflect.Slice {
		return nil, nil
	}
	if v.Len() == 0 {
		return nil, nil
	}

	for i := 0; i < v.Len(); i++ {
		indexv := v.Index(i)
		if indexv.Type().Kind() != reflect.Struct {
			return nil, nil
		}
		mapKeyInterface := indexv.FieldByName(key)
		if mapKeyInterface.Kind() == reflect.Invalid {
			return nil, nil
		}
		mapKeyString, err := interfaceToString(mapKeyInterface.Interface())
		if err != nil {
			return nil, err
		}
		result[mapKeyString] = indexv.Interface()
	}
	return result, err
}
func interfaceToString(v interface{}) (result string, err error) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		result = fmt.Sprintf("%v", v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = fmt.Sprintf("%v", v)
	case reflect.String:
		result = v.(string)
	default:
		err = nil
	}
	return result, err
}
func arrayKey(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
