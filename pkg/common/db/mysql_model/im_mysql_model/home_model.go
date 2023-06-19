package im_mysql_model

import "Open_IM/pkg/common/db"

func GetHomeVisual() ([]db.HomeVisual, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var results []db.HomeVisual
	err = dbConn.Table("home_visual").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func SetHomeVisual(vis db.HomeVisual) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	visOld := db.HomeVisual{}
	err = dbConn.Table("home_visual").Where("status_name=?", vis.StatusName).Debug().Find(&visOld).Error
	if err != nil || visOld.StatusName == "" {
		dbConn.Table("home_visual").Debug().Create(&vis)
	} else {
		dbConn.Table("home_visual").Where("status_name=?", vis.StatusName).Debug().Updates(&vis)
	}

	return nil
}
