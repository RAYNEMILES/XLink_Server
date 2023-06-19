package im_mysql_model

import (
	"Open_IM/pkg/common/db"
)

func GetBindingList(userId string) []db.OauthClient {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	var oauth []db.OauthClient
	dbConn.Model(&oauth).Where("user_id=?", userId).Find(&oauth)

	return oauth
}

func GetUserByTypeAndThirdId(thirdType int, thirdId string) db.OauthClient {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return db.OauthClient{}
	}

	oauth := db.OauthClient{}
	dbConn.Model(&oauth).Where("third_id=? AND type=?", thirdId, thirdType).Take(&oauth)

	return oauth
}

func InsertThirdInfo(thirdType int, thirdId, thirdUserName string, userId ...string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	user := db.OauthClient{
		ThirdId:       thirdId,
		Type:          int8(thirdType),
		ThirdUserName: thirdUserName,
	}

	if len(userId) > 0 {
		user.UserId = userId[0]
	}

	insertErr := dbConn.Table(db.OauthClient{}.TableName()).Create(&user).Error
	if insertErr != nil {
		return false
	}
	return true
}

func UpdateUserIdByThirdInfo(thirdType int, thirdId string, userId string, thirdUserName ...string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	data := GetUserByTypeAndThirdId(thirdType, thirdId)
	updateErr := err
	if data.ThirdId != "" {
		sub := dbConn.Table(db.OauthClient{}.TableName()).Where("third_id=? AND type=?", thirdId, thirdType).Update("user_id", userId)
		if len(thirdUserName) > 0 {
			sub = sub.Update("third_user_name", thirdUserName[0])
		}
		updateErr = sub.Error
	} else {
		updateErr = dbConn.Table(db.OauthClient{}.TableName()).Create(&db.OauthClient{
			ThirdId:       thirdId,
			Type:          int8(thirdType),
			ThirdUserName: thirdUserName[0],
			UserId:        userId,
		}).Error
		if updateErr != nil {
			return false
		}
	}

	if updateErr != nil {
		return false
	}
	return true
}

func UpdateThirdUserNameByThirdInfo(thirdType int, thirdId string, thirdUserName string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	updateErr := dbConn.Table(db.OauthClient{}.TableName()).Where("third_id=? AND type=?", thirdId, thirdType).Update("third_user_name", thirdUserName).Error
	if updateErr != nil {
		return false
	}
	return true
}

func UnbindingOauthByUserIdAndThirdId(thirdType int, userId string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	updateErr := dbConn.Table(db.OauthClient{}.TableName()).Where("user_id=? AND type=?", userId, thirdType).Update("user_id", "").Error
	if updateErr != nil {
		return false
	}
	return true
}
