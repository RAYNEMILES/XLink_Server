package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"gorm.io/gorm"
)

func GetAllDomains() ([]db.Domain, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var results []db.Domain
	err = dbConn.Table("domain").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func SaveAllDomains(domains []db.Domain) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		tx.Where("1 = 1").Delete(&db.Domain{})
		tx.Table("domain").Create(&domains)
		return nil
	})
	return err
}
