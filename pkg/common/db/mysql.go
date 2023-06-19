package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"fmt"

	"gorm.io/gorm/schema"

	// "github.com/go-sql-driver/mysql"
	"sync"
	"time"

	// "github.com/jinzhu/gorm"
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mysqlDB struct {
	sync.RWMutex
	dbMap map[string]*gorm.DB
}

func initMysqlDB() {
	// When there is no open IM database, connect to the mysql built-in database to create openIM database
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		config.Config.Mysql.DBUserName, config.Config.Mysql.DBPassword, config.Config.Mysql.DBAddress[0], "mysql")
	var db *gorm.DB
	var err1 error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: log.GetSqlLogger(constant.MySQLLogFileName), NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		fmt.Println("0", "Open failed ", err.Error(), dsn)
		time.Sleep(time.Duration(30) * time.Second)
		db, err1 = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: log.GetSqlLogger(constant.MySQLLogFileName), NamingStrategy: schema.NamingStrategy{SingularTable: true}})
		if err1 != nil {
			fmt.Println("0", "Open failed ", err1.Error(), dsn)
			panic(err1.Error())
		}
	}

	// Check the database and table during initialization
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s default charset utf8mb4 COLLATE utf8mb4_unicode_ci;", config.Config.Mysql.DBDatabaseName)
	err = db.Exec(sql).Error
	if err != nil {
		fmt.Println("0", "Exec failed ", err.Error(), sql)
		panic(err.Error())
	}

	sqlDB, _ := db.DB()
	sqlDB.Close()

	dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		config.Config.Mysql.DBUserName, config.Config.Mysql.DBPassword, config.Config.Mysql.DBAddress[0], config.Config.Mysql.DBDatabaseName)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: log.GetSqlLogger(constant.MySQLLogFileName), NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		fmt.Println("0", "Open failed ", err.Error(), dsn)
		panic(err.Error())
	}

	sqlDB, _ = db.DB()
	fmt.Println("open db ok ", dsn)
	db.AutoMigrate(
		&Register{},
		&Friend{},
		&FriendRequest{},
		&Group{},
		&GroupMember{},
		&GroupRequest{},
		&User{}, &PrivacySetting{},

		&Black{}, &BlackForMoment{}, &ChatLog{}, &Register{}, &Conversation{}, &AppVersion{}, &Department{}, &NewAppVersion{}, &DiscoverUrl{},
		&InviteCodeLog{}, &InviteCode{}, &InviteCodeRelation{}, &InviteChannelCode{}, &Config{}, &AdminUser{},
		&AdminAPIs{}, &AdminPages{}, &AdminRole{}, &MomentSQL{}, &MomentLikeSQL{}, &MomentCommentSQL{}, &OauthClient{},
		&Contact{}, &ContactExclude{},
		&InterestType{},
		&InterestLanguage{},
		&InterestUser{},
		&InterestGroup{}, &InterestGroupExclude{},
		&GroupHeat{},
		&Official{}, &OfficialAnalytics{}, &OfficialInterest{},
		&OfficialFollowSQL{}, &ArticleLikeSQL{}, &ArticleSQL{}, &ArticleReadSQL{}, &ArticleCommentSQL{}, ArticleCommentLikeSQL{},
		&FavoritesSQL{}, &VideoAudioCommunicationRecord{}, &CommunicationGroupMember{},
		&GameCategories{}, &GameLink{}, &Game{}, &GamePlayHistory{}, &GameFavorites{},
		&MePageURL{},

		&ShortVideo{}, &ShortVideoLike{}, &ShortVideoComment{}, &ShortVideoCommentLike{}, &ShortVideoFollow{}, &ShortVideoUserCount{}, &ShortVideoNotice{},
	)

	db.Set("gorm:table_options", "CHARSET=utf8mb4")
	db.Set("gorm:table_options", "collation=utf8mb4_unicode_ci")

	if !db.Migrator().HasTable(&Friend{}) {
		fmt.Println("CreateTable Friend")
		db.Migrator().CreateTable(&Friend{})
	}

	if !db.Migrator().HasTable(&FriendRequest{}) {
		fmt.Println("CreateTable FriendRequest")
		db.Migrator().CreateTable(&FriendRequest{})
	}

	if !db.Migrator().HasTable(&Group{}) {
		fmt.Println("CreateTable Group")
		db.Migrator().CreateTable(&Group{})
	}

	if !db.Migrator().HasTable(&GroupMember{}) {
		fmt.Println("CreateTable GroupMember")
		db.Migrator().CreateTable(&GroupMember{})
	}
	if !db.Migrator().HasTable(&GroupRequest{}) {
		fmt.Println("CreateTable GroupRequest")
		db.Migrator().CreateTable(&GroupRequest{})
	}
	if !db.Migrator().HasTable(&User{}) {
		fmt.Println("CreateTable User")
		db.Migrator().CreateTable(&User{})
	}
	if !db.Migrator().HasTable(&PrivacySetting{}) {
		fmt.Println("CreateTable PrivacySetting")
		db.Migrator().CreateTable(&PrivacySetting{})
	}
	if !db.Migrator().HasTable(&AdminUser{}) {
		fmt.Println("CreateTable AdminUser")
		db.Migrator().CreateTable(&AdminUser{})
	}
	if !db.Migrator().HasTable(&Black{}) {
		fmt.Println("CreateTable Black")
		db.Migrator().CreateTable(&Black{})
	}
	if !db.Migrator().HasTable(&BlackForMoment{}) {
		fmt.Println("CreateTable BlackForMoment")
		db.Migrator().CreateTable(&BlackForMoment{})
	}
	if !db.Migrator().HasTable(&BlackList{}) {
		fmt.Println("CreateTable BlackList")
		db.Migrator().CreateTable(&BlackList{})
	}
	if !db.Migrator().HasTable(&ChatLog{}) {
		fmt.Println("CreateTable ChatLog")
		db.Migrator().CreateTable(&ChatLog{})
	}
	if !db.Migrator().HasTable(&Register{}) {
		fmt.Println("CreateTable Register")
		db.Migrator().CreateTable(&Register{})
	}
	if !db.Migrator().HasTable(&Conversation{}) {
		fmt.Println("CreateTable Conversation")
		db.Migrator().CreateTable(&Conversation{})
	}

	if !db.Migrator().HasTable(&Department{}) {
		fmt.Println("CreateTable Department")
		db.Migrator().CreateTable(&Department{})
	}
	if !db.Migrator().HasTable(&OrganizationUser{}) {
		fmt.Println("CreateTable OrganizationUser")
		db.Migrator().CreateTable(&OrganizationUser{})
	}
	if !db.Migrator().HasTable(&DepartmentMember{}) {
		fmt.Println("CreateTable DepartmentMember")
		db.Migrator().CreateTable(&DepartmentMember{})
	}
	if !db.Migrator().HasTable(&AppVersion{}) {
		fmt.Println("CreateTable DepartmentMember")
		db.Migrator().CreateTable(&AppVersion{})
	}
	if !db.Migrator().HasTable(&NewAppVersion{}) {
		fmt.Println("CreateTable NewAppVersion")
		db.Migrator().CreateTable(&NewAppVersion{})
	}
	if !db.Migrator().HasTable(&DiscoverUrl{}) {
		fmt.Println("CreateTable DiscoverUrl")
		db.Migrator().CreateTable(&DiscoverUrl{})
	}
	if !db.Migrator().HasTable(&InviteCodeLog{}) {
		fmt.Println("CreateTable InviteCode")
		db.Migrator().CreateTable(&InviteCodeLog{})
	}
	if !db.Migrator().HasTable(&InviteCodeRelation{}) {
		fmt.Println("CreateTable InviteCode relation")
		db.Migrator().CreateTable(&InviteCodeRelation{})
	}
	if !db.Migrator().HasTable(&InviteCode{}) {
		fmt.Println("CreateTable InviteCodeCode")
		db.Migrator().CreateTable(&InviteCode{})
	}
	if !db.Migrator().HasTable(&InviteChannelCode{}) {
		fmt.Println("CreateTable InviteChannelCode")
		db.Migrator().CreateTable(&InviteChannelCode{})
	}
	if !db.Migrator().HasTable(&Config{}) {
		fmt.Println("CreateTable Config")
		db.Migrator().CreateTable(&Config{})
	}
	if !db.Migrator().HasTable(&GroupUpdatesVersion{}) {
		fmt.Println("CreateTable GroupUpdatesVersion")
		db.Migrator().CreateTable(&GroupUpdatesVersion{})
	}

	if !db.Migrator().HasTable(&AdminAPIs{}) {
		fmt.Println("CreateTable AdminAPIs")
		db.Migrator().CreateTable(&AdminAPIs{})
	}
	if !db.Migrator().HasTable(&AdminPages{}) {
		fmt.Println("CreateTable AdminPages")
		db.Migrator().CreateTable(&AdminPages{})
	}
	// if !db.Migrator().HasTable(&AdminActions{}) {
	// 	fmt.Println("CreateTable AdminActions")
	// 	db.Migrator().CreateTable(&AdminActions{})
	// }
	if !db.Migrator().HasTable(&AdminRole{}) {
		fmt.Println("CreateTable AdminRole")
		db.Migrator().CreateTable(&AdminRole{})
	}
	if !db.Migrator().HasTable(&MomentSQL{}) {
		fmt.Println("CreateTable Moment")
		db.Migrator().CreateTable(&MomentSQL{})
	}
	if !db.Migrator().HasTable(&MomentLikeSQL{}) {
		fmt.Println("CreateTable MomentLike")
		db.Migrator().CreateTable(&MomentLikeSQL{})
	}
	if !db.Migrator().HasTable(&MomentCommentSQL{}) {
		fmt.Println("CreateTable MomentComment")
		db.Migrator().CreateTable(&MomentCommentSQL{})
	}
	if !db.Migrator().HasTable(&OauthClient{}) {
		fmt.Println("CreateTable oauth_client")
		db.Migrator().CreateTable(&OauthClient{})
	}
	if !db.Migrator().HasTable(&Contact{}) {
		fmt.Println("CreateTable contact")
		db.Migrator().CreateTable(&Contact{})
	}
	if !db.Migrator().HasTable(&ContactExclude{}) {
		fmt.Println("CreateTable " + ContactExclude{}.TableName())
		db.Migrator().CreateTable(&ContactExclude{})
	}
	if !db.Migrator().HasTable(&InterestType{}) {
		fmt.Println("CreateTable " + InterestType{}.TableName())
		db.Migrator().CreateTable(&InterestType{})
	}
	if !db.Migrator().HasTable(&InterestLanguage{}) {
		fmt.Println("CreateTable " + InterestLanguage{}.TableName())
		db.Migrator().CreateTable(&InterestLanguage{})
	}
	if !db.Migrator().HasTable(&InterestUser{}) {
		fmt.Println("CreateTable " + InterestUser{}.TableName())
		db.Migrator().CreateTable(&InterestUser{})
	}
	if !db.Migrator().HasTable(&InterestGroup{}) {
		fmt.Println("CreateTable " + InterestGroup{}.TableName())
		db.Migrator().CreateTable(&InterestGroup{})
	}
	if !db.Migrator().HasTable(&InterestGroupExclude{}) {
		fmt.Println("CreateTable " + InterestGroupExclude{}.TableName())
		db.Migrator().CreateTable(&InterestGroupExclude{})
	}
	if !db.Migrator().HasTable(&GroupHeat{}) {
		fmt.Println("CreateTable " + GroupHeat{}.TableName())
		db.Migrator().CreateTable(&GroupHeat{})
	}
	if !db.Migrator().HasTable(&OfficialFollowSQL{}) {
		fmt.Println("CreateTable " + OfficialFollowSQL{}.TableName())
		db.Migrator().CreateTable(&OfficialFollowSQL{})
	}

	if !db.Migrator().HasTable(&ArticleLikeSQL{}) {
		fmt.Println("CreateTable " + ArticleLikeSQL{}.TableName())
		db.Migrator().CreateTable(&ArticleLikeSQL{})
	}
	if !db.Migrator().HasTable(&ArticleSQL{}) {
		fmt.Println("CreateTable " + ArticleSQL{}.TableName())
		db.Migrator().CreateTable(&ArticleSQL{})
	}
	if !db.Migrator().HasTable(&ArticleReadSQL{}) {
		fmt.Println("CreateTable " + ArticleReadSQL{}.TableName())
		db.Migrator().CreateTable(&ArticleReadSQL{})
	}
	if !db.Migrator().HasTable(&ArticleCommentSQL{}) {
		fmt.Println("CreateTable " + ArticleCommentSQL{}.TableName())
		db.Migrator().CreateTable(&ArticleCommentSQL{})
	}
	if !db.Migrator().HasTable(&ArticleCommentLikeSQL{}) {
		fmt.Println("CreateTable " + ArticleCommentLikeSQL{}.TableName())
		db.Migrator().CreateTable(&ArticleCommentLikeSQL{})
	}
	if !db.Migrator().HasTable(&Official{}) {
		fmt.Println("CreateTable " + Official{}.TableName())
		db.Migrator().CreateTable(&Official{})
	}
	if !db.Migrator().HasTable(&OfficialAnalytics{}) {
		fmt.Println("CreateTable " + OfficialAnalytics{}.TableName())
		db.Migrator().CreateTable(&OfficialAnalytics{})
	}
	if !db.Migrator().HasTable(&OfficialInterest{}) {
		fmt.Println("CreateTable " + OfficialInterest{}.TableName())
		db.Migrator().CreateTable(&OfficialInterest{})
	}
	if !db.Migrator().HasTable(&FavoritesSQL{}) {
		fmt.Println("CreateTable " + FavoritesSQL{}.TableName())
		db.Migrator().CreateTable(&FavoritesSQL{})
	}
	if !db.Migrator().HasTable(&VideoAudioCommunicationRecord{}) {
		fmt.Println("CreateTable " + VideoAudioCommunicationRecord{}.TableName())
		db.Migrator().CreateTable(&VideoAudioCommunicationRecord{})
	}
	if !db.Migrator().HasTable(&CommunicationGroupMember{}) {
		fmt.Println("CreateTable " + CommunicationGroupMember{}.TableName())
		db.Migrator().CreateTable(&CommunicationGroupMember{})
	}
	if !db.Migrator().HasTable(&GameCategories{}) {
		fmt.Println("CreateTable " + GameCategories{}.TableName())
		db.Migrator().CreateTable(&GameCategories{})
	}
	if !db.Migrator().HasTable(&GameLink{}) {
		fmt.Println("CreateTable " + GameLink{}.TableName())
		db.Migrator().CreateTable(&GameLink{})
	}
	if !db.Migrator().HasTable(&Game{}) {
		fmt.Println("CreateTable " + Game{}.TableName())
		db.Migrator().CreateTable(&Game{})
	}
	if !db.Migrator().HasTable(&GamePlayHistory{}) {
		fmt.Println("CreateTable " + GamePlayHistory{}.TableName())
		db.Migrator().CreateTable(&GamePlayHistory{})
	}
	if !db.Migrator().HasTable(&GameFavorites{}) {
		fmt.Println("CreateTable " + GameFavorites{}.TableName())
		db.Migrator().CreateTable(&GameFavorites{})
	}
	if !db.Migrator().HasTable(&MePageURL{}) {
		fmt.Println("CreateTable " + MePageURL{}.TableName())
		db.Migrator().CreateTable(&MePageURL{})
	}

	if !db.Migrator().HasTable(&ShortVideo{}) {
		fmt.Println("CreateTable " + ShortVideo{}.TableName())
		db.Migrator().CreateTable(&ShortVideo{})
	}
	if !db.Migrator().HasTable(&ShortVideoUserCount{}) {
		fmt.Println("CreateTable " + ShortVideoUserCount{}.TableName())
		db.Migrator().CreateTable(&ShortVideoUserCount{})
	}
	if !db.Migrator().HasTable(&ShortVideoLike{}) {
		fmt.Println("CreateTable " + ShortVideoLike{}.TableName())
		db.Migrator().CreateTable(&ShortVideoLike{})
	}
	if !db.Migrator().HasTable(&ShortVideoComment{}) {
		fmt.Println("CreateTable " + ShortVideoComment{}.TableName())
		db.Migrator().CreateTable(&ShortVideoComment{})
	}
	if !db.Migrator().HasTable(&ShortVideoCommentLike{}) {
		fmt.Println("CreateTable " + ShortVideoCommentLike{}.TableName())
		db.Migrator().CreateTable(&ShortVideoCommentLike{})
	}
	if !db.Migrator().HasTable(&ShortVideoFollow{}) {
		fmt.Println("CreateTable " + ShortVideoFollow{}.TableName())
		db.Migrator().CreateTable(&ShortVideoFollow{})
	}
	if !db.Migrator().HasTable(&ShortVideoNotice{}) {
		fmt.Println("CreateTable " + ShortVideoNotice{}.TableName())
		db.Migrator().CreateTable(&ShortVideoNotice{})
	}
	sqlDB.Close()
}

func (m *mysqlDB) DefaultGormDB() (*gorm.DB, error) {
	return m.GormDB(config.Config.Mysql.DBAddress[0], config.Config.Mysql.DBDatabaseName)
}

func (m *mysqlDB) GormDB(dbAddress, dbName string) (*gorm.DB, error) {
	m.Lock()
	defer m.Unlock()

	k := key(dbAddress, dbName)
	if _, ok := m.dbMap[k]; !ok {
		if err := m.open(dbAddress, dbName); err != nil {
			return nil, err
		}
	}
	return m.dbMap[k], nil
}

func (m *mysqlDB) open(dbAddress, dbName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		config.Config.Mysql.DBUserName, config.Config.Mysql.DBPassword, dbAddress, dbName)
	// db, err := gorm.Open("mysql", dsn, &gorm.Config{Logger: log.GetNewLogger(constant.SQLiteLogFileName)})
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: log.GetSqlLogger(constant.MySQLLogFileName), NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		return err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(config.Config.Mysql.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(config.Config.Mysql.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.Config.Mysql.DBMaxLifeTime) * time.Second)

	if m.dbMap == nil {
		m.dbMap = make(map[string]*gorm.DB)
	}
	k := key(dbAddress, dbName)
	m.dbMap[k] = db
	return nil
}
