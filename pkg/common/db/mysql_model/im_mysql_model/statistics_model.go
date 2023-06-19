package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"time"
)

func GetActiveUserNum(from, to int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("chat_logs").Select("count(distinct(send_id))").Where("create_time >= ? and create_time <= ?", from, to).Count(&num).Error
	return num, err
}

func GetIncreaseUserNum(from, to int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("users").Where("create_time >= ? and create_time <= ?", from, to).Count(&num).Error
	return num, err
}

func GetTotalUserNum() (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("users").Count(&num).Error
	return num, err
}

func GetTotalUserNumByDate(to int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("users").Where("create_time <= ?", to).Count(&num).Error
	return num, err
}

func GetPrivateMessageNum(from, to time.Time) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("chat_logs").Where("create_time >= ? and create_time <= ? and session_type = ?", from, to, 1).Count(&num).Error
	return num, err
}

func GetGroupMessageNum(from, to time.Time) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("chat_logs").Where("create_time >= ? and create_time <= ? and session_type = ?", from, to, 2).Count(&num).Error
	return num, err
}

func GetIncreaseGroupNum(from, to time.Time) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("groups").Where("create_time >= ? and create_time <= ?", from, to).Count(&num).Error
	return num, err
}

func GetTotalGroupNum() (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("groups").Count(&num).Error
	return num, err
}

func GetGroupNum(to time.Time) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var num int64
	err = dbConn.Table("groups").Where("create_time <= ?", to).Count(&num).Error
	return num, err
}

type activeGroup struct {
	Name       string
	Id         string `gorm:"column:recv_id"`
	MessageNum int    `gorm:"column:message_num"`
}

func GetActiveGroups(from, to time.Time, limit int) ([]*activeGroup, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var activeGroups []*activeGroup
	if err != nil {
		return activeGroups, err
	}

	err = dbConn.Table("chat_logs").Select("recv_id, count(*) as message_num").Where("create_time >= ? and create_time <= ? and session_type = ?", from, to, 2).Group("recv_id").Limit(limit).Order("message_num DESC").Find(&activeGroups).Error
	for _, activeGroup := range activeGroups {
		group := db.Group{
			GroupID: activeGroup.Id,
		}
		dbConn.Table("groups").Where("group_id= ? ", group.GroupID).Find(&group)
		activeGroup.Name = group.GroupName
	}
	return activeGroups, err
}

type activeUser struct {
	Name       string
	Id         string `gorm:"column:send_id"`
	MessageNum int    `gorm:"column:message_num"`
}

func GetActiveUsers(from, to time.Time, limit int) ([]*activeUser, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var activeUsers []*activeUser
	if err != nil {
		return activeUsers, err
	}

	err = dbConn.Table("chat_logs").Select("send_id, count(*) as message_num").Where("create_time >= ? and create_time <= ? and session_type = ?", from, to, 1).Group("send_id").Limit(limit).Order("message_num DESC").Find(&activeUsers).Error
	for _, activeUser := range activeUsers {
		user := db.User{
			UserID: activeUser.Id,
		}
		dbConn.Table("users").Select("user_id, name").Where("user_id=?", activeUser.Id).Find(&user)
		activeUser.Name = user.Nickname
	}
	return activeUsers, err
}

func GetTodayPlayedGameNumber() (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64
	t := time.Now()
	todayZero := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()

	err = dbConn.Table(db.GamePlayHistory{}.TableName()).Distinct("game_code").Where("create_time >= ?", todayZero).Count(&count).Error

	return count, err
}

func CumulativePlayedGameNumber() (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64
	err = dbConn.Table(db.GamePlayHistory{}.TableName()).Distinct("game_code").Count(&count).Error
	return count, err
}

func GetGamePlayed(from, to int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	err = dbConn.Table(db.GamePlayHistory{}.TableName()).Debug().Select("game_code").Where("create_time >= ? AND create_time <= ?", from, to).Count(&count).Error

	return count, err

}

type activeGame struct {
	GameCode   string
	GameNameEN string
	GameNameCN string
	PlayCounts int64
}

func GetActiveGames() ([]*activeGame, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var activeGames = make([]*activeGame, 0)
	if err != nil {
		return activeGames, err
	}

	err = dbConn.Table(db.Game{}.TableName()).Debug().Select("game_code, game_name_en, game_name_cn, play_counts").
		Where("delete_time=0").Where("play_counts>0").
		Limit(20).Order("play_counts DESC, create_time DESC").Find(&activeGames).Error

	return activeGames, err
}
