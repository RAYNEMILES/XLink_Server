package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"fmt"
	"strings"
	"time"
)

func GetAppVersionByID(id string) (*db.NewAppVersion, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	appversion := db.NewAppVersion{}
	err = dbConn.Table("new_app_version").Where("id=? AND delete_time = 0", id).First(&appversion).Error
	if err != nil {
		return nil, err
	}
	return &appversion, err
}

func GetLatestAppVersion(typeA int) (*db.NewAppVersion, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	appversion := db.NewAppVersion{}
	err = dbConn.Table("new_app_version").Where("type=? AND delete_time = 0 AND status = 2", typeA).Order("version desc").First(&appversion).Error
	if err != nil {
		return nil, err
	}
	return &appversion, err
}

func GetAppVersionsByPage(client, status, begin, end string, number, page int32, orderBy string) ([]db.NewAppVersion, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var (
		appversions []db.NewAppVersion
		query       string
	)

	if err != nil {
		return appversions, err
	}

	if client != "" {
		query += "type = ?"
	} else {
		query += "type != ?"
	}

	if status != "" {
		query += " AND status = ?"
	} else {
		query += " AND status != ?"
	}

	query += " AND create_time >= ?"

	if end != "" {
		query += " AND create_time <= ?"
	} else {
		query += " AND create_time != ?"
	}

	query += " AND delete_time = 0"

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
		dbConn = dbConn.Order(orderByClause)
	}
	//log.NewInfo("", utils.GetSelfFuncName(), "GetAppVersionsByPage mysql:", query, client, status, begin, end, number, page)
	err = dbConn.Table("new_app_version").Where(query, client, status, begin, end).Limit(int(number)).Offset(int(number * (page - 1))).Find(&appversions).Error
	if err != nil {
		return nil, err
	}

	return appversions, nil
}

func GetAppVersionsCount(client, status, begin, end string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var (
		count int64
		query string
	)

	if err != nil {
		return 0, err
	}

	if client != "" {
		query += "type = ?"
	} else {
		query += "type != ?"
	}

	if status != "" {
		query += " AND status = ?"
	} else {
		query += " AND status != ?"
	}

	query += " AND create_time >= ?"

	if end != "" {
		query += " AND create_time <= ?"
	} else {
		query += " AND create_time != ?"
	}

	query += " AND delete_time = 0"

	err = dbConn.Table("new_app_version").Where(query, client, status, begin, end).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func AddAppVersion(title, content, downloadUrl, version, createUser string, isforce, typeA, status int) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	appVersion := db.NewAppVersion{
		Version:     version,
		Type:        typeA,
		Status:      status,
		Isforce:     isforce,
		Title:       title,
		Content:     content,
		DownloadUrl: downloadUrl,
		CreateTime:  time.Now().Unix(),
		CreateUser:  createUser,
	}
	err = dbConn.Table("new_app_version").Create(&appVersion).Error
	return err
}

func EditAppVersion(title, content, downloadUrl, version, updateUser, id string, isforce, typeA, status int) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	appVersion := db.NewAppVersion{
		Version:     version,
		Type:        typeA,
		Status:      status,
		Isforce:     isforce,
		Title:       title,
		Content:     content,
		DownloadUrl: downloadUrl,
		UpdateTime:  time.Now().Unix(),
		UpdateUser:  updateUser,
	}
	err = dbConn.Table("new_app_version").Where("id=?", id).Updates(&appVersion).Error
	return err
}

func DeleteAppVersion(id, deleteUser string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	appVersion := db.NewAppVersion{
		DeleteTime: time.Now().Unix(),
		DeleteUser: deleteUser,
	}
	err = dbConn.Table("new_app_version").Where("id=?", id).Updates(&appVersion).Error
	return err
}
