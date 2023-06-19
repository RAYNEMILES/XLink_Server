package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"gorm.io/gorm/clause"
)

func InitUserConfig(userId string) bool {
	var insertData []db.PrivacySetting
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	// allow to add me by phone
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyAddByPhone,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// allow to add me by account
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyAddByAccount,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// allow to add me by email
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyAddByEmail,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// allow to see my wooms
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacySeeWooms,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// Allow non-friends to chat privately
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyPrivateChat,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// group chat
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyAddByGroup,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// qr code
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyAddByQr,
		SettingVal: constant.PrivacyStatusOpen,
	})
	// group chat
	insertData = append(insertData, db.PrivacySetting{
		UserId:     userId,
		SettingKey: constant.PrivacyAddByContactCard,
		SettingVal: constant.PrivacyStatusOpen,
	})

	err = dbConn.Create(&insertData).Error
	if err != nil {
		log.Error("InitUserConfig", "dbConn.Create", err.Error())
		return false
	}
	return true
}

func GetPrivacySetting(userId string) ([]db.PrivacySetting, error) {
	var privacySetting []db.PrivacySetting
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	err = dbConn.Where("user_id = ?", userId).Find(&privacySetting).Error
	if err != nil {
		InitUserConfig(userId)

		// allow to add me by phone
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyAddByPhone,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// allow to add me by account
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyAddByAccount,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// allow to add me by email
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyAddByEmail,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// allow to see my wooms
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacySeeWooms,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// Allow non-friends to chat privately
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyPrivateChat,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// group chat
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyAddByGroup,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// qr code
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyAddByQr,
			SettingVal: constant.PrivacyStatusOpen,
		})
		// group chat
		privacySetting = append(privacySetting, db.PrivacySetting{
			UserId:     userId,
			SettingKey: constant.PrivacyAddByContactCard,
			SettingVal: constant.PrivacyStatusOpen,
		})
		return privacySetting, nil
	}
	return privacySetting, nil
}

func GetUserPrivacyByUserIdList(userIdList []string) ([]db.PrivacySetting, error) {
	var privacySetting []db.PrivacySetting
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	err = dbConn.Where("user_id in (?)", userIdList).Find(&privacySetting).Error
	if err != nil {
		return nil, err
	}
	return privacySetting, nil
}

func UpdatePrivacySetting(userId string, settingKey string, settingVal string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	data := db.PrivacySetting{
		UserId:     userId,
		SettingKey: settingKey,
		SettingVal: settingVal,
	}
	err = dbConn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"setting_val"}),
	}).Create(&data).Error

	if err != nil {
		return err
	}
	return nil
}
