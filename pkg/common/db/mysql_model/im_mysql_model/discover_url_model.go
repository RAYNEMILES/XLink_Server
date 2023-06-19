package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"errors"
	"strconv"
	"time"
)

// find discover url
func GetDiscoverUrl(platformId string) (*db.DiscoverUrl, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	//select url where platform_id equals platformId and delete_time is null
	discoverUrl := db.DiscoverUrl{}
	err = dbConn.Model(&discoverUrl).Where("platform_id=? AND delete_time = 0", platformId).First(&discoverUrl).Error
	if err != nil {
		return nil, err
	}
	return &discoverUrl, nil
}

// Save Discover URL
func SaveDiscoverUrl(url string, userId string, platformId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	//update or create url
	discoverUrl := db.DiscoverUrl{}
	if err := dbConn.Model(&discoverUrl).Where("platform_id=?", platformId).First(&discoverUrl).Error; err != nil {
		//create it if it is no existed
		discoverUrl.Url = url
		discoverUrl.PlatformId, err = strconv.Atoi(platformId)
		if err != nil {
			return err
		}
		discoverUrl.CreateUser = userId
		discoverUrl.CreateTime = time.Now().Unix()
		result := dbConn.Model(&discoverUrl).Create(&discoverUrl)
		return result.Error
	}

	//only update it if it is existed.
	//discoverUrl.Url = url
	//discoverUrl.UpdateUser = userId
	//discoverUrl.UpdateTime = time.Now().Unix()
	updateData := map[string]interface{}{
		"url":         url,
		"update_user": userId,
		"update_time": time.Now().Unix(),
	}
	result := dbConn.Model(&discoverUrl).Where("id=?", discoverUrl.ID).Updates(updateData).Error
	return result
}

// switch the visible status of discover page
func SwitchDiscoverStatus(status int, platformId string, userId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	if status > 1 || status < 0 {
		return errors.New("status value is illegale")
	}

	//discoverUrl := db.DiscoverUrl{}
	//discoverUrl.Status = status
	//discoverUrl.UpdateUser = userId
	//discoverUrl.UpdateTime = time.Now().Unix()

	discoverUrl := map[string]interface{}{
		"status":      status,
		"update_user": userId,
		"update_time": time.Now().Unix(),
	}

	err = dbConn.Model(&db.DiscoverUrl{}).Where("platform_id=?", platformId).Updates(discoverUrl).Error
	return err
}
