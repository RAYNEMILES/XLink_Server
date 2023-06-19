package im_mysql_model

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/utils"
	"fmt"
	"time"
)

func PushCode(inviteCodeLog *db.InviteCodeLog) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	err = dbConn.Table(inviteCodeLog.TableName()).Create(inviteCodeLog).Error
	if err != nil {
		return err
	}
	return nil
}

func GetCode(params api.InviteCodeReq) ([]db.InviteCodeLog, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var lists []db.InviteCodeLog

	// suggest: Periodically clear table data
	sql := "SELECT *,( channel.code + channel.timezone + channel.mobile + channel.os + channel.version + channel.webkit + channel.screen_width + channel.ip) AS cal FROM (SELECT id, IF ( `code` = ?, 1, 0 ) AS 'code', IF ( `timezone` = ?, 1, 0 ) AS 'timezone', IF ( `mobile` = ?, 1, 0 ) AS 'mobile', IF ( `os` = ?, 1, 0 ) AS 'os', IF ( `version` = ?, 1, 0 ) AS 'version', IF ( `webkit` = ?, 1, 0 ) AS 'webkit', IF ( `screen_width` = ?, 1, 0 ) AS 'screen_width', IF ( `ip` = ?, 1, 0 ) AS 'ip' FROM invite_code_log WHERE create_time >= ? ) AS channel WHERE ( channel.code + channel.timezone + channel.mobile + channel.os + channel.version + channel.webkit + channel.screen_width + channel.ip ) >=3 ORDER BY cal DESC"
	err = dbConn.Raw(sql, params.Code, params.Timezone, params.Mobile, params.Os, params.Version, params.Webkit, params.ScreenWidth, params.Ip, time.Now().AddDate(0, 0, -1).Unix()).Scan(&lists).Error

	if err != nil {
		return nil, err
	}
	return lists, nil
}

func GetCodeByOr(params *api.InviteCodeReq) ([]db.InviteCodeLog, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var lists []db.InviteCodeLog

	nums := []int{0, 1, 2, 3, 4, 5, 6, 7, 8}
	m := 6
	column := []string{"code", "timezone", "mobile", "os", "version", "webkit", "screen_width", "language", "ip"}
	values := []interface{}{params.Code, params.Timezone, params.Mobile, params.Os, params.Version, params.Webkit, params.ScreenWidth, params.Language, params.Ip}

	n := len(nums)
	indexs := utils.ZuheResult(n, m)
	result := utils.FindNumsByIndexs(nums, indexs)

	dbConn = dbConn.Model(&db.InviteCodeLog{})

	for _, ints := range result {
		dbConn = dbConn.Or(fmt.Sprintf("`%s`=? and `%s` =? and `%s` =? and `%s` =? and `%s` =? and `%s` =?", column[ints[0]], column[ints[1]], column[ints[2]], column[ints[3]], column[ints[4]], column[ints[5]]), values[ints[0]], values[ints[1]], values[ints[2]], values[ints[3]], values[ints[4]], values[ints[5]])
	}

	// suggest: Periodically clear table data
	err2 := dbConn.Order("id DESC").Limit(100).Find(&lists).Error
	if err2 != nil {
		return nil, err2
	}
	return lists, nil
}
