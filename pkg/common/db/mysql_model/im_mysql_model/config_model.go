package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"strconv"
)

func GetConfigByName(name string) (*db.Config, error) {
	config := db.Config{}
	cache, _ := db.DB.GetConfigCache(name)
	if cache != "" {
		config.Name = name
		config.Value = cache
		return &config, nil
	}

	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	err = dbConn.Model(&config).Where("name=?", name).First(&config).Error

	if err != nil {
		return nil, err
	}
	db.DB.SetConfigCache(name, config.Value)
	return &config, nil
}

func GetConfigByNameByDefault(name, acquiescent string) (*db.Config, error) {
	config := db.Config{}
	cache, _ := db.DB.GetConfigCache(name)
	if cache != "" {
		config.Name = name
		config.Value = cache
		return &config, nil
	}

	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	err = dbConn.Model(&config).Where("name=?", name).First(&config).Error

	if err != nil {
		dbConn.Model(&config).Create(&db.Config{Name: name, Value: acquiescent})
		config.Name = name
		config.Value = acquiescent
		return &config, nil
	}

	if config.Name == "" && config.Value == "" {
		config.Name = name
		config.Value = acquiescent

		dbConn.Model(&config).Create(&db.Config{Name: name, Value: acquiescent})
	}

	return &config, nil
}

func GetAllowGuestLogin() bool {
	config, err := GetConfigByNameByDefault(constant.AllowGuestLogin, "1")
	if err != nil {
		return false
	}

	if config.Value == "1" {
		return true
	}
	return false
}

func GetAllowGuestLimit() bool {
	config, err := GetConfigByNameByDefault(constant.AllowRegisterByUuid, "1")
	if err != nil {
		return false
	}

	if config.Value == "1" {
		return true
	}
	return false
}

func GetInviteCodeIsOpen() bool {
	config, _ := GetConfigByNameByDefault(constant.ConfigInviteCodeIsOpenKey, strconv.Itoa(constant.ConfigInviteCodeIsOpenTrue))
	if config.Value == strconv.Itoa(constant.ConfigInviteCodeIsOpenTrue) {
		return true
	}
	return false
}
func GetChannelCodeIsOpen() bool {
	config, _ := GetConfigByNameByDefault(constant.ConfigChannelCodeIsOpenKey, strconv.Itoa(constant.ConfigChannelCodeIsOpenTrue))
	if config.Value == strconv.Itoa(constant.ConfigChannelCodeIsOpenTrue) {
		return true
	}
	return false
}

func GetChannelCodeLimit() bool {
	config, _ := GetConfigByNameByDefault(constant.ConfigChannelCodeIsLimitKey, strconv.Itoa(constant.ConfigChannelCodeIsLimitTrue))
	if config.Value == strconv.Itoa(constant.ConfigChannelCodeIsLimitTrue) {
		return true
	}
	return false
}

func SetConfigByName(name, value string) bool {
	config := db.Config{}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}
	dbConn.Table(config.TableName()).Where("name=?", name).First(&config)

	if config != (db.Config{}) {
		dbConn.Model(&config).Where("name=?", name).Update("value", value)
		db.DB.DeleteConfigCache(name)
		return true
	}
	dbConn.Model(&config).Create(&db.Config{Name: name, Value: value})
	return true
}
