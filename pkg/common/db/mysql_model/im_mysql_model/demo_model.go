package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"time"
)

func GetRegisterFromPhone(account string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var r db.User
	return &r, dbConn.Table("users").Where("phone_number = ? and delete_time = 0",
		account).Take(&r).Error
}

func GetRegisterFromUserId(account string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var r db.User
	return &r, dbConn.Table("users").Where("user_id = ? and delete_time = 0",
		account).Take(&r).Error
}

func GetHistoryRegisterFromUserId(account string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var r db.User
	return &r, dbConn.Table("users").Where("user_id = ?", account).Take(&r).Error
}

func GetRegisterFromEmail(account string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var r db.User
	return &r, dbConn.Table("users").Where("email = ? and delete_time = 0",
		account).Take(&r).Error
}

func GetRegisterFromEmailOrUserIdOrPhone(account string) (*db.User, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var r db.User
	return &r, dbConn.Table("users").Where("(email = ? or user_id= ? or phone_number=?) and delete_time = 0",
		account, account, account).Take(&r).Error
}

func SetPassword(account, password, ex, userID string) error {
	r := db.Register{
		Account:  account,
		Password: password,
		Ex:       ex,
		UserID:   userID,
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table("registers").Create(&r).Error
}

func ResetPassword(account, password string) error {
	r := db.User{
		Password: password,
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()

	if err != nil {
		return err
	}
	return dbConn.Table("users").Where("user_id = ? and delete_time = 0", account).Updates(&r).Error
}

func UpdateUserLastLoginIP(userID, ip string) error {
	r := db.User{
		LoginIp:       ip,
		LastLoginTime: time.Now().Unix(),
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table("users").Debug().Where("user_id = ? and delete_time = 0", userID).Updates(&r).Error
}

func UpdateUserLastLoginDevice(userID string, platform int32) error {
	r := db.User{
		LastLoginDevice: int8(platform),
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table("users").Where("user_id = ? and delete_time = 0", userID).Updates(&r).Error
}
