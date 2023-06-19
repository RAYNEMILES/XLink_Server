package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"gorm.io/gorm"
	"time"
)

func GetMePageUrl(pageType int32) ([]db.MePageURL, error){
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	mePageInfo := []db.MePageURL{}
	if err != nil {
		return mePageInfo, err
	}

	dbConn.Table(db.MePageURL{}.TableName()).Where("delete_time = 0").Where("type=?", pageType).Find(&mePageInfo)
	return mePageInfo, nil
}

func GetMePageURLs() ([]db.MePageURL, error){
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	mePageInfo := []db.MePageURL{}
	if err != nil {
		return mePageInfo, err
	}

	dbConn.Table(db.MePageURL{}.TableName()).Where("delete_time = 0").Find(&mePageInfo)
	return mePageInfo, nil
}

func SaveMePageUrl(page map[string]string, pageType int32, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	if len(page) == 0 {
		return nil
	}
	var count int64
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		err = tx.Table(db.MePageURL{}.TableName()).Where("type=?", pageType).Count(&count).Error
		if err != nil {
			return err
		}
		for language, url := range page {
			pageURL := db.MePageURL{}
			pageURL.CreateTime = time.Now().Unix()
			pageURL.CreateUser = opUserID
			pageURL.Status = 1
			pageURL.Language = language
			pageURL.Type = int64(pageType)
			pageURL.Url = url
			if count == 0 {
				err = dbConn.Table(pageURL.TableName()).Create(&pageURL).Error
				if err != nil {
					return err
				}
			} else {
				err = tx.Table(pageURL.TableName()).Where("type=?", pageURL.Type).Where("language=?", language).Updates(&pageURL).Error
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return err
}

func SwitchMePageUrl(page *db.MePageURL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table(page.TableName()).Debug().Where("type = ?", page.Type).Updates(page).Error
	return err
}
