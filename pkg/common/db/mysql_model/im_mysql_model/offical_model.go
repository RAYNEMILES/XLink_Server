package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	pb "Open_IM/pkg/proto/news"
	"Open_IM/pkg/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
)

type InsertOfficialParams struct {
	UserID      string
	UserGender  int32
	Type        int8
	IdType      int8
	IdName      string
	IdNumber    string
	FaceURL     string
	Nickname    string
	Bio         string
	CountryCode string
	Interests   []int64
}

type UpdateSelfOfficialParams struct {
	OfficialID          int64
	FaceURL             string
	Nickname            string
	Bio                 string
	NicknameUpdateTime  int64
	NicknameUpdateCount int8
	Interests           []int64
}

type OfficialProcessParams struct {
	OpUserId        string
	OfficialId      int64
	ProcessStatus   int32
	ProcessFeedback string
}

func GetOfficialByOfficialID(officialID int64) (db.Official, error) {
	official := db.Official{Id: officialID}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return official, err
	}
	err = dbConn.Table(official.TableName()).Where("delete_time", 0).Take(&official).Error
	return official, err
}

func GetOfficialByOfficialIDAll(officialID int64) (db.Official, error) {
	official := db.Official{Id: officialID}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return official, err
	}
	err = dbConn.Table(official.TableName()).Take(&official).Error
	return official, err
}

func GetOfficialByName(officialType int32, nickName string) (db.Official, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	official := db.Official{}
	if err != nil {
		return official, err
	}
	err = dbConn.Table(official.TableName()).
		Where("(nickname = '' AND initial_nickname = ?) OR (nickname = ?)", nickName, nickName).
		Where("type=?", officialType).
		Take(&official).Error
	return official, nil
}

func GetAllSystemOfficials() ([]db.Official, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	official := []db.Official{}
	if err != nil {
		return official, err
	}
	err = dbConn.Table("official").
		Where("is_system=1").
		Where("delete_time=0").
		Where("status=1").
		Find(&official).Error
	return official, nil
}

func GetDeletedOfficialIDList() ([]int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var result []int64
	err = dbConn.Table("official").Select("id").
		Where("delete_time != 0 OR status=2").
		Find(&result).Error
	return result, nil
}

func GetOfficialsByOfficialIDList(officialIDList []int64) (map[int64]db.Official, error) {
	if len(officialIDList) == 0 {
		officialsMap := make(map[int64]db.Official, 0)
		return officialsMap, nil
	}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	if len(officialIDList) == 0 {
		return map[int64]db.Official{}, nil
	}
	var officials []db.Official
	err = dbConn.Table(db.Official{}.TableName()).
		Where("delete_time", 0).
		Where("id IN (?)", officialIDList).
		Find(&officials).Error
	if err != nil {
		return nil, err
	}

	officialsMap := make(map[int64]db.Official, len(officials))
	for _, official := range officials {
		officialsMap[official.Id] = official
	}

	return officialsMap, err
}

func GetSystemOfficialIDList(limit int64) ([]int64, error) {
	var result []int64
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	err = dbConn.Table(db.Official{}.TableName()).Select("id").Where("is_system = 1").Limit(int(limit)).Find(&result).Error
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetOfficialInterestsByOfficialID(officialID int64) ([]int64, error) {
	var interests []db.OfficialInterest
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	err = dbConn.Table(db.OfficialInterest{}.TableName()).Where("official_id = ?", officialID).Find(&interests).Error
	if err != nil {
		return nil, err
	}

	officialInterest := make([]int64, len(interests))
	for i, interest := range interests {
		officialInterest[i] = interest.InterestTypeID
	}

	return officialInterest, nil
}

func DeleteOfficialInterest(officialID int64) error {
	var interests []db.OfficialInterest
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	err = dbConn.Table(db.OfficialInterest{}.TableName()).Where("official_id = ?", officialID).Delete(&interests).Error
	if err != nil {
		return err
	}

	return nil
}

func CheckOfficialIDNumberAvailable(IDNumber string, IdType int8, selfOfficial int64) bool {
	var foundValueCount int64 = 0
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return true
	}
	dbSub := dbConn.Table(db.Official{}.TableName()).Where("id_number = ? && id_type = ?", IDNumber, IdType)
	if selfOfficial != 0 {
		dbSub = dbSub.Where("id != ?", selfOfficial)
	}
	err = dbSub.Debug().Count(&foundValueCount).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
	}
	if foundValueCount > 0 {
		return true
	}
	return false
}

func InsertOfficial(params InsertOfficialParams) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var official db.Official
	now := time.Now().Unix()
	official.CreateTime = now
	official.InitialNickname = params.Nickname
	official.NicknameUpdateTime = now
	official.FollowCounts = 1
	if err = utils.CopyStructFields(&official, &params); err != nil {
		return 0, err
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		if err = tx.Table(official.TableName()).Create(&official).Error; err != nil {
			return err
		}
		if len(params.Interests) > 0 {
			officialInterests := make([]db.OfficialInterest, len(params.Interests))
			for i, interest := range params.Interests {
				officialInterests[i] = db.OfficialInterest{OfficialID: official.Id, InterestTypeID: interest}
			}
			if err = tx.Table(officialInterests[0].TableName()).Create(&officialInterests).Error; err != nil {
				return err
			}
		}
		follow := db.OfficialFollowSQL{
			OfficialID: official.Id,
			UserID:     params.UserID,
			FollowTime: now,
			Muted:      true,
			Enabled:    true,
		}
		if err = tx.Table(follow.TableName()).Create(&follow).Error; err != nil {
			return err
		}
		analytics := db.OfficialAnalytics{
			OfficialID:   official.Id,
			Time:         utils.FloorTimeToHours(time.Unix(now, 0)),
			Gender:       params.UserGender,
			FollowCounts: 1,
		}
		if err = tx.Table(analytics.TableName()).Create(&analytics).Error; err != nil {
			return err
		}
		return tx.Table(db.User{}.TableName()).Where("user_id = ?", params.UserID).Update("official_id", official.Id).Error
	})
	if err != nil {
		return 0, err
	}
	return official.Id, nil
}

func UpdateSelfOfficial(params UpdateSelfOfficialParams) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	official := map[string]interface{}{
		"face_url":              params.FaceURL,
		"nickname":              params.Nickname,
		"bio":                   params.Bio,
		"nickname_update_time":  params.NicknameUpdateTime,
		"nickname_update_count": params.NicknameUpdateCount,
		"update_time":           time.Now().Unix(),
	}

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		if err = tx.Table(db.Official{}.TableName()).Where("id = ?", params.OfficialID).Updates(official).Error; err != nil {
			return err
		}
		if err = tx.Delete(db.OfficialInterest{}, "official_id = ?", params.OfficialID).Error; err != nil {
			return err
		}
		if len(params.Interests) > 0 {
			officialInterests := make([]db.OfficialInterest, len(params.Interests))
			for i, interest := range params.Interests {
				officialInterests[i] = db.OfficialInterest{OfficialID: params.OfficialID, InterestTypeID: interest}
			}
			if err = tx.Table(officialInterests[0].TableName()).Create(&officialInterests).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func UpdateOfficialAccountIPAndLocationInfo(OfficialID int64, IPAddress, Region, City string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	official := map[string]interface{}{
		"last_login_ip":         IPAddress,
		"last_activity_country": Region,
		"last_activity_city":    City,
		"update_time":           time.Now().Unix(),
	}
	err = dbConn.Table(db.Official{}.TableName()).Where("id = ?", OfficialID).Updates(official).Error
	if err != nil {
		return err
	}
	return nil
}

func GetOfficialAccountsByWhere(where map[string]string, tagIdList []int64, showNumber int32, pageNumber int32, orderBy string) ([]*db.GetOfficialRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var officialAccountRes []*db.GetOfficialRes
	var count int64
	if err != nil {
		return officialAccountRes, count, err
	}
	var officialInterests []db.OfficialInterest
	officialAccountMap := make(map[int64]*db.GetOfficialRes)
	interestsTypeResMap := make(map[int64]*db.InterestTypeRes)
	var officialIdList []int64
	sortMap := map[string]string{}
	timeTypeMap := map[int]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"
	timeTypeMap[1] = "create_time"
	timeTypeMap[2] = "process_time"

	// query interests
	log.Debug("", "tag id list: ", tagIdList)
	queryInterests := false
	if len(tagIdList) > 0 {
		queryInterests = true
		dbConn.Table(db.OfficialInterest{}.TableName()).Where("interest_type_id IN (?)", tagIdList).Find(&officialInterests)
		for _, official := range officialInterests {
			officialIdList = append(officialIdList, official.OfficialID)
		}
	}

	// query official account
	dbSub := dbConn.Table("official").Where("delete_time=0")

	if queryInterests {
		dbSub = dbSub.Where("id IN (?)", officialIdList)
	}
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
	if officialAccount, ok := where["official_account"]; ok {
		if officialAccount != "" {
			dbSub = dbSub.Where("nickname like ? OR initial_nickname like ? OR user_id like ?", "%"+officialAccount+"%", "%"+officialAccount+"%", "%"+officialAccount+"%")
		}
	}
	if accountType, ok := where["account_type"]; ok {
		if accountType != "" {
			accountTypeInt, _ := strconv.Atoi(accountType)
			if accountTypeInt != 0 {
				dbSub = dbSub.Where("type=?", accountTypeInt)
			}
		}
	}
	if idType, ok := where["id_type"]; ok {
		if idType != "" {
			idTypeInt, _ := strconv.Atoi(idType)
			if idTypeInt != 0 {
				dbSub = dbSub.Where("id_type=?", idTypeInt)
			}
		}
	}
	if idNumber, ok := where["id_number"]; ok {
		if idNumber != "" {
			dbSub = dbSub.Where("id_number like ?", "%"+idNumber+"%")
		}
	}
	if bio, ok := where["bio"]; ok {
		if bio != "" {
			dbSub = dbSub.Where("bio like ?", "%"+bio+"%")
		}
	}
	if isSystem, ok := where["is_system"]; ok {
		if isSystem != "" && isSystem != "0" {
			isSystemInt, _ := strconv.ParseInt(isSystem, 10, 64)
			dbSub = dbSub.Where("is_system=?", isSystemInt-1)
		}
	}

	if processStatus, ok := where["process_status"]; ok {
		if processStatus != "" {
			processStatusInt, _ := strconv.Atoi(processStatus)
			if processStatusInt != -1 {
				dbSub = dbSub.Where("process_status=?", processStatusInt)
			}
		}
	}
	startTime := where["start_time"]
	endTime := where["end_time"]
	timeTypeStr := where["time_type"]
	timeType, _ := strconv.Atoi(timeTypeStr)
	if timeType == 0 {
		if startTime != "" || endTime != "" {
			var sql = ""
			var paramList []interface{}
			isFirst := true
			for _, v := range timeTypeMap {
				beenOr := false
				if startTime != "" {
					if isFirst {
						sql += fmt.Sprintf("%s>=?", v)
						paramList = append(paramList, startTime)
					} else {
						sql += fmt.Sprintf(" OR %s>=?", v)
						paramList = append(paramList, startTime)
						beenOr = true
					}
				}
				if endTime != "" {
					if isFirst || beenOr {
						sql += fmt.Sprintf(" AND %s<=?", v)
						paramList = append(paramList, endTime)
					} else {
						sql += fmt.Sprintf(" OR %s<=?", v)
						paramList = append(paramList, endTime)
					}
				}
				isFirst = false
			}

			log.Debug("", "time sql: ", sql, " param: ", paramList)
			dbSub = dbSub.Where(sql, paramList...)
		}
	} else {
		if startTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s>=?", timeTypeMap[timeType]), startTime)
		}
		if endTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s<=?", timeTypeMap[timeType]), endTime)
		}
	}

	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}
	var officialAccounts []db.Official
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&officialAccounts)

	for _, account := range officialAccounts {
		officialRes := db.GetOfficialRes{Interests: []*db.InterestTypeRes{}}
		utils.CopyStructFields(&officialRes, &account)
		officialAccountMap[account.Id] = &officialRes
		officialAccountRes = append(officialAccountRes, &officialRes)
	}

	// if there isn't tags condition,
	tagIdList = make([]int64, 0)
	officialIdList = make([]int64, 0)
	for _, officialAccount := range officialAccountRes {
		officialIdList = append(officialIdList, officialAccount.Id)
		log.Debug("", "officialAccount.Id: ", officialAccount.Id)
	}
	interestsType := make([]*db.InterestType, 0)
	interestsLanguage := make([]*db.InterestLanguage, 0)

	// officialInterests
	dbConn.Table(db.OfficialInterest{}.TableName()).Where("official_id IN (?)", officialIdList).Find(&officialInterests)
	log.Debug("", "Select id list: ", officialInterests)
	for _, interest := range officialInterests {
		tagIdList = append(tagIdList, interest.InterestTypeID)
	}
	dbConn.Table("interest_type").Where("id IN (?)", tagIdList).Find(&interestsType)
	dbConn.Table("interest_language").Where("interest_id IN (?)", tagIdList).Find(&interestsLanguage)
	interestsTypeResMap = mergeInterestTypeToResMap(interestsType, interestsLanguage)

	for _, officialInterest := range officialInterests {
		if official, ok := officialAccountMap[officialInterest.OfficialID]; ok {
			if interest, ok := interestsTypeResMap[officialInterest.InterestTypeID]; ok {
				official.Interests = append(official.Interests, interest)
			}
		}
	}

	return officialAccountRes, count, nil
}

func GetOfficialCountsByWhere(_ *db.Official, where map[string]string, tagIdList []int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	timeTypeMap := map[int]string{}
	var count int64
	var officialIdList []string
	timeTypeMap[1] = "create_time"
	timeTypeMap[2] = "process_time"

	// query official id list from interest id list.
	queryInterests := false
	if len(tagIdList) > 0 {
		queryInterests = true
		dbConn.Table(db.OfficialInterest{}.TableName()).Select("official_id").Where("interest_type_id IN (?)", tagIdList).Find(&officialIdList)
	}

	dbSub := dbConn.Table("official").Where("delete_time=0")
	if queryInterests {
		dbSub = dbSub.Where("id IN (?)", officialIdList)
	}
	if officialAccount, ok := where["official_account"]; ok {
		if officialAccount != "" {
			dbSub = dbSub.Where("nickname like ? OR initial_nickname like ? OR user_id like ?", "%"+officialAccount+"%", "%"+officialAccount+"%", "%"+officialAccount+"%")
		}
	}
	if accountType, ok := where["account_type"]; ok {
		if accountType != "" {
			accountTypeInt, _ := strconv.Atoi(accountType)
			if accountTypeInt != 0 {
				dbSub = dbSub.Where("type=?", accountTypeInt)
			}
		}
	}
	if idType, ok := where["id_type"]; ok {
		if idType != "" {
			idTypeInt, _ := strconv.Atoi(idType)
			if idTypeInt != 0 {
				dbSub = dbSub.Where("id_type=?", idTypeInt)
			}
		}
	}
	if idNumber, ok := where["id_number"]; ok {
		if idNumber != "" {
			dbSub = dbSub.Where("id_number like ?", "%"+idNumber+"%")
		}
	}
	if processStatus, ok := where["process_status"]; ok {
		if processStatus != "" {
			processStatusInt, _ := strconv.Atoi(processStatus)
			if processStatusInt != -1 {
				dbSub = dbSub.Where("process_status=?", processStatusInt)
			}
		}
	}
	startTime := where["start_time"]
	endTime := where["end_time"]
	timeTypeStr := where["time_type"]
	timeType, _ := strconv.Atoi(timeTypeStr)
	if timeType == 0 {
		if startTime != "" || endTime != "" {
			var sql = ""
			var paramList []interface{}
			isFirst := true
			for _, v := range timeTypeMap {
				beenOr := false
				if startTime != "" {
					if isFirst {
						sql += fmt.Sprintf("%s>=?", v)
						paramList = append(paramList, startTime)
					} else {
						sql += fmt.Sprintf(" OR %s>=?", v)
						paramList = append(paramList, startTime)
						beenOr = true
					}
				}
				if endTime != "" {
					if isFirst || beenOr {
						sql += fmt.Sprintf(" AND %s<=?", v)
						paramList = append(paramList, endTime)
					} else {
						sql += fmt.Sprintf(" OR %s<=?", v)
						paramList = append(paramList, endTime)
					}
				}
				isFirst = false
			}

			dbSub = dbSub.Where(sql, paramList...)
		}
	} else {
		if startTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s>=?", timeTypeMap[timeType]), startTime)
		}
		if endTime != "" {
			dbSub = dbSub.Where(fmt.Sprintf("%s<=?", timeTypeMap[timeType]), endTime)
		}
	}
	if err := dbSub.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil

}

func DeleteOfficialAccounts(officialsIdList []string, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	dbOfficial := db.Official{
		DeleteTime: time.Now().Unix(),
		DeleteBy:   opUserId,
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		i += tx.Table("official").Where("id IN (?)", officialsIdList).Debug().Updates(&dbOfficial).RowsAffected
		tx.Table(db.User{}.TableName()).Where("official_id IN (?)", officialsIdList).Debug().UpdateColumn("official_id", 0)
		return nil
	})
	if err != nil {
		return 0
	}

	return i
}

func AlterOfficialAccount(official *pb.Official, interestsIdList []int64) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	i = 0
	var officialInterests []*db.OfficialInterest
	for _, interestId := range interestsIdList {
		officialInterest := db.OfficialInterest{}
		officialInterest.OfficialID = official.Id
		officialInterest.InterestTypeID = interestId
		officialInterests = append(officialInterests, &officialInterest)
	}
	officialDb := db.Official{}
	utils.CopyStructFields(&officialDb, &official)
	officialDb.UpdateTime = time.Now().Unix()
	err = dbConn.Transaction(func(tx *gorm.DB) error {

		// update official information.
		i += tx.Table(officialDb.TableName()).Updates(&official).RowsAffected
		if official.IsSystem == 0 {
			if len(officialInterests) > 0 {
				tx.Table(db.OfficialInterest{}.TableName()).Where("official_id = ?", official.Id).Delete(&db.InterestLanguage{})
				i += tx.Table(db.OfficialInterest{}.TableName()).Create(&officialInterests).RowsAffected
			}

		}

		return nil
	})
	if err != nil {
		return 0
	}
	return i
}

func ClearAllFollowers(officialID int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		result := tx.Table(db.OfficialFollowSQL{}.TableName()).Where("official_id=?", officialID).UpdateColumn("delete_time", time.Now().Unix())
		if result.Error != nil {
			return result.Error
		}
		count += result.RowsAffected

		specifyMap := make(map[string]interface{})
		specifyMap["is_system"] = 0
		specifyMap["follow_counts"] = 0

		result = tx.Table(db.Official{}.TableName()).
			Where("id=?", officialID).
			Updates(specifyMap)
		if result.Error != nil {
			return result.Error
		}
		count += result.RowsAffected

		anay := db.OfficialAnalytics{}
		result = tx.Table(anay.TableName()).Where("official_id=?", officialID).Delete(&anay)
		if result.Error != nil {
			return result.Error
		}
		count += result.RowsAffected

		return nil
	})

	return count, err
}

func AddOfficialAccount(official *db.Official, interests []int64, user *db.User) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	official.CreateTime = now
	official.ProcessStatus = 1
	official.Status = 1
	var registerCount int64 = 0

	// if the user has been registered, return err.
	dbConn.Table("official").Where("delete_time=0").Where("user_id=?", official.UserID).Count(&registerCount)
	if registerCount != 0 {
		return constant.ErrArgs
	}

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		err = tx.Table(official.TableName()).Create(official).Error
		if err != nil {
			return err
		}
		r := tx.Table(db.User{}.TableName()).Where("user_id=?", official.UserID).UpdateColumn("official_id", official.Id)
		if r.Error != nil {
			return err
		}
		if r.RowsAffected == 0 {
			return constant.ErrArgs
		}
		if official.IsSystem == 0 && len(interests) > 0 {
			var officialInterest []db.OfficialInterest
			interestStr := ""
			for _, interest := range interests {
				officialInterest = append(officialInterest, db.OfficialInterest{OfficialID: official.Id, InterestTypeID: interest})
				if interestStr == "" {
					interestStr = strconv.FormatInt(interest, 10)
				} else {
					interestStr = interestStr + "," + strconv.FormatInt(interest, 10)
				}
			}

			err = tx.Table(db.OfficialInterest{}.TableName()).Create(&officialInterest).Error
			if err != nil {
				return err
			}

			follow := db.OfficialFollowSQL{
				OfficialID: official.Id,
				UserID:     official.UserID,
				FollowTime: now,
				Muted:      true,
				Enabled:    true,
			}
			if err = tx.Table(follow.TableName()).Create(&follow).Error; err != nil {
				return err
			}
			analytics := db.OfficialAnalytics{
				OfficialID:   official.Id,
				Time:         utils.FloorTimeToHours(time.Unix(now, 0)),
				Gender:       user.Gender,
				FollowCounts: 1,
			}
			if err = tx.Table(analytics.TableName()).Create(&analytics).Error; err != nil {
				return err
			}
			return tx.Table(db.User{}.TableName()).Where("user_id = ?", user.UserID).Update("official_id", official.Id).Error
		}
		return err
	})
	return err
}

func Process(params *OfficialProcessParams) (i int64) {
	nowTime := time.Now().Unix()
	var official db.Official
	official.Id = params.OfficialId
	official.ProcessStatus = int8(params.ProcessStatus)
	official.ProcessFeedback = params.ProcessFeedback
	official.UpdateTime = nowTime
	official.ProcessBy = params.OpUserId
	official.ProcessTime = nowTime
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	i = dbConn.Table(db.Official{}.TableName()).Updates(&official).RowsAffected
	return i
}

func AddOfficialFollow(officialID int64, userID string, userGender int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		// check if user is already have follow record
		var existingFollow db.OfficialFollowSQL
		err = tx.Table(existingFollow.TableName()).
			Where("official_id", officialID).
			Where("user_id", userID).
			Take(&existingFollow).Error

		// if not create
		if err == gorm.ErrRecordNotFound {
			newFollow := db.OfficialFollowSQL{
				OfficialID: officialID,
				UserID:     userID,
				FollowTime: time.Now().Unix(),
				Muted:      true,
				Enabled:    true,
			}
			createErr := tx.Table(newFollow.TableName()).Create(&newFollow).Error
			if createErr != nil {
				return createErr
			}
		} else if err != nil {
			// other errors
			return err
		} else {
			// skip if already followed
			if existingFollow.DeleteTime == 0 {
				return nil
			}
			// update
			updateErr := tx.Table(existingFollow.TableName()).
				Where("id", existingFollow.Id).
				Update("delete_time", 0).
				Update("deleted_by", "").
				Update("follow_time", time.Now().Unix()).
				Error
			if updateErr != nil {
				return updateErr
			}
		}

		// update official follows count
		err = tx.Table(db.Official{}.TableName()).
			Where("id", officialID).
			Update("follow_counts", gorm.Expr("follow_counts + 1")).
			Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"follow_counts": gorm.Expr("follow_counts + 1")}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID:   officialID,
				Time:         utils.FloorTimeToHours(time.Now()),
				Gender:       userGender,
				FollowCounts: 1,
			}).Error
	})
}

func AddOfficialFollows(officialID int64, allCount int64, userIdList []string, genderList []int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		// check if user is already have follow record
		err = tx.Table("official_follow").
			Where("official_id", officialID).
			UpdateColumn("delete_time", 0).Error

		// replace into test_tbl (id,dr) values (1,'2'),(2,'3'),...(x,'y');
		newFollows := make([]db.OfficialFollowSQL, 1000)
		batchSize := 1000 // 每批次插入的记录数
		total := len(userIdList)
		nowTime := time.Now().Unix()
		nowFloorTime := utils.FloorTimeToHours(time.Now())
		var maleCount int64 = 0
		var famleCount int64 = 0
		log.Debug("totle: ", total)
		for i := 0; i < total; i += batchSize {
			end := i + batchSize
			if end > total {
				end = total
			}
			userIDBatch := userIdList[i:end]
			genderBatch := genderList[i:end]
			batchLen := len(userIDBatch)
			if batchLen < 1000 {
				newFollows = make([]db.OfficialFollowSQL, batchLen)
			}
			for index, userId := range userIDBatch {
				newFollows[index] = db.OfficialFollowSQL{
					OfficialID: officialID,
					UserID:     userId,
					FollowTime: nowTime,
					Muted:      true,
					Enabled:    true,
				}
				if genderBatch[index] == 1 {
					maleCount++
				} else {
					famleCount++
				}
			}

			createErr := tx.Table("official_follow").Debug().Create(&newFollows).Error
			if createErr != nil {
				return createErr
			}
		}

		// update official follows count
		err = tx.Table(db.Official{}.TableName()).
			Where("id=?", officialID).
			Update("follow_counts", allCount).
			Error
		if err != nil {
			return err
		}

		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"follow_counts": maleCount}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID:   officialID,
				Time:         nowFloorTime,
				Gender:       1,
				FollowCounts: maleCount,
			}).Error

		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"follow_counts": famleCount}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID:   officialID,
				Time:         nowFloorTime,
				Gender:       2,
				FollowCounts: famleCount,
			}).Error
	})
}

func DeleteOfficialFollows(officialID int64, userID string, userList []db.User) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	userIDList := make([]string, len(userList))
	analyticsList := make([]db.OfficialAnalytics, len(userList))
	for i, user := range userList {
		userIDList[i] = user.UserID
		analyticsList[i] = db.OfficialAnalytics{
			OfficialID:   officialID,
			Time:         utils.FloorTimeToHours(time.Now()),
			Gender:       user.Gender,
			FollowCounts: -1,
		}
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		deleteFollowsTx := tx.Table(db.OfficialFollowSQL{}.TableName()).
			Where("official_id", officialID).
			Where("user_id", userIDList).
			Where("delete_time", 0).
			Update("deleted_by", userID).
			Update("delete_time", time.Now().Unix())
		if deleteFollowsTx.Error != nil {
			return deleteFollowsTx.Error
		}
		if deleteFollowsTx.RowsAffected == 0 {
			return nil
		}
		err = tx.
			Table(db.Official{}.TableName()).
			Where("id", officialID).
			Update("follow_counts", gorm.Expr("follow_counts - ?", deleteFollowsTx.RowsAffected)).
			Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"follow_counts": gorm.Expr("follow_counts - 1")}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&analyticsList).Error
	})
}

func UpdateOfficialFollow(officialID int64, userID string, muted, enabled bool) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Table(db.OfficialFollowSQL{}.TableName()).
		Where("official_id", officialID).
		Where("user_id", userID).
		Updates(map[string]interface{}{"muted": muted, "enabled": enabled}).Error
}

func GetOfficialFollowByOfficialAndUserID(officialID int64, userID string) (*db.OfficialFollowSQL, error) {
	officialFollow := db.OfficialFollowSQL{}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	err = dbConn.Table(officialFollow.TableName()).
		Where("official_id", officialID).
		Where("user_id", userID).
		Find(&officialFollow).Error
	return &officialFollow, err
}

type OfficialFollow struct {
	OfficialID int64  `gorm:"official_id"`
	Nickname   string `gorm:"nickname"`
	FaceURL    string `gorm:"face_url"`
	Bio        string `gorm:"bio"`
	Type       int32  `gorm:"type"`
	FollowTime int64  `gorm:"follow_time"`
	Muted      bool   `gorm:"muted"`
	Enabled    bool   `gorm:"enabled"`
}

func GetUserFollowList(userID string, offset, limit int, keyword string) ([]OfficialFollow, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	var follows []OfficialFollow
	dbSub := dbConn.Table(db.OfficialFollowSQL{}.TableName()+" f").
		Select(
			"o.id as official_id",
			"o.nickname",
			"o.face_url",
			"o.bio",
			"o.type",
			"f.follow_time",
			"f.muted",
			"f.enabled",
		).
		Where("f.user_id", userID).
		Where("f.delete_time", 0).
		Where("o.delete_time", 0).
		Where("o.process_status", 1).
		Joins("LEFT JOIN " + db.Official{}.TableName() + " o ON o.id = f.official_id").
		Order("f.follow_time desc").
		Offset(offset).
		Limit(limit)

	if keyword != "" {
		dbSub = dbSub.Where("o.nickname like ?", "%"+keyword+"%")
	}

	if err = dbSub.Find(&follows).Error; err != nil {
		return nil, 0, err
	}

	if err = dbSub.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	return follows, count, nil
}

func GetUserFollowOfficialAccountList(userID string) ([]db.OfficialFollowSQL, error) {
	officialFollow := []db.OfficialFollowSQL{}
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	err = dbConn.Table(db.OfficialFollowSQL{}.TableName()).
		Where("user_id", userID).
		Where("delete_time", 0).
		Find(&officialFollow).Error
	return officialFollow, err
}

func GetOfficialProfile(userID string, officialID int64) (OfficialFollow, error) {
	var profile OfficialFollow
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return profile, err
	}

	err = dbConn.Table(db.Official{}.TableName()+" o").
		Select(
			"o.id as official_id",
			"o.nickname",
			"o.face_url",
			"o.bio",
			"o.type",
			"f.follow_time",
			"f.muted",
			"f.enabled",
		).
		Where("o.delete_time", 0).
		Where("o.process_status", 1).
		Where("o.id", officialID).
		Joins("LEFT JOIN "+db.OfficialFollowSQL{}.TableName()+" f ON f.official_id = o.id AND f.user_id = ? AND f.delete_time = 0", userID).
		Take(&profile).Error

	return profile, err
}

func AddOfficialFollowBlocks(officialID int64, officialUserID string, userIDList []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	officialFollow := db.OfficialFollowSQL{
		BlockTime: time.Now().Unix(),
		BlockedBy: officialUserID,
	}

	return dbConn.Table(officialFollow.TableName()).
		Where("official_id", officialID).
		Where("user_id", userIDList).
		Updates(&officialFollow).Error
}

func DeleteOfficialFollowBlocks(officialID int64, userIDList []string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	officialFollow := db.OfficialFollowSQL{}

	return dbConn.Table(officialFollow.TableName()).
		Where("official_id", officialID).
		Where("user_id", userIDList).
		Update("blocked_by", "").
		Update("block_time", 0).
		Error
}

func AddArticleLike(officialID, articleID int64, userID string, userGender int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Transaction(func(tx *gorm.DB) error {
		var articleLike db.ArticleLikeSQL
		err = tx.Table(articleLike.TableName()).
			Where("article_id", articleID).
			Where("user_id", userID).
			Where("delete_time", 0).
			Take(&articleLike).Error
		if err != gorm.ErrRecordNotFound {
			return err
		}
		newLike := db.ArticleLikeSQL{
			ArticleID:  articleID,
			UserID:     userID,
			CreatedBy:  userID,
			CreateTime: time.Now().Unix(),
			Status:     1,
		}
		if err = tx.Table(newLike.TableName()).Create(&newLike).Error; err != nil {
			return err
		}
		if err = tx.Table(db.ArticleSQL{}.TableName()).Where("article_id", articleID).Update("like_counts", gorm.Expr("like_counts + 1")).Error; err != nil {
			return err
		}
		if err = tx.Table(db.Official{}.TableName()).Where("id", officialID).Update("like_counts", gorm.Expr("like_counts + 1")).Error; err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"like_counts": gorm.Expr("like_counts + 1")}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID: officialID,
				Time:       utils.FloorTimeToHours(time.Now()),
				Gender:     userGender,
				LikeCounts: 1,
			}).Error
	})
}

func DeleteArticleLike(officialID, articleID int64, userID string, userGender int32, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		update := map[string]interface{}{
			"deleted_by":  opUserID,
			"delete_time": time.Now().Unix(),
		}

		delTx := tx.Table(db.ArticleLikeSQL{}.TableName()).
			Where("article_id", articleID).
			Where("user_id", userID).
			Where("delete_time", 0).
			Updates(update)

		if delTx.Error != nil {
			return delTx.Error
		}

		if delTx.RowsAffected == 0 {
			return nil
		}

		if err = tx.Table(db.ArticleSQL{}.TableName()).Where("article_id", articleID).Update("like_counts", gorm.Expr("like_counts - 1")).Error; err != nil {
			return err
		}
		err = tx.Table(db.Official{}.TableName()).Where("id", officialID).Update("like_counts", gorm.Expr("like_counts - 1")).Error
		if err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"like_counts": gorm.Expr("like_counts - 1")}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID: officialID,
				Time:       utils.FloorTimeToHours(time.Now()),
				Gender:     userGender,
				LikeCounts: -1,
			}).Error
	})
}

func AddArticleComment(articleID, articleOfficialID, officialID int64, userID string, userGender int32, parentCommentID int64, replyOfficialID int64, replyUserID, content, opUserID string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	comment := db.ArticleCommentSQL{
		ParentCommentID: parentCommentID,
		ArticleID:       articleID,
		OfficialID:      officialID,
		UserID:          userID,
		ReplyUserID:     replyUserID,
		ReplyOfficialID: replyOfficialID,
		Content:         content,
		CreatedBy:       opUserID,
		CreateTime:      time.Now().Unix(),
		Status:          1,
	}

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		if err = tx.Table(comment.TableName()).Create(&comment).Error; err != nil {
			return err
		}
		if parentCommentID != 0 {
			if err = tx.Table(comment.TableName()).Where("comment_id", parentCommentID).Update("reply_counts", gorm.Expr("reply_counts + 1")).Error; err != nil {
				return err
			}
		}
		err = tx.Table(db.ArticleSQL{}.TableName()).Where("article_id", articleID).Update("comment_counts", gorm.Expr("comment_counts + 1")).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"comment_counts": gorm.Expr("comment_counts + 1")}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID:    articleOfficialID,
				Time:          utils.FloorTimeToHours(time.Now()),
				Gender:        userGender,
				CommentCounts: 1,
			}).Error
	})
	if err != nil {
		return 0, err
	}
	return comment.CommentID, nil
}

func AddArticleCommentLike(commentID int64, userID string, officialID int64, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Transaction(func(tx *gorm.DB) error {
		var articleLike db.ArticleCommentLikeSQL
		queryTx := tx.Table(articleLike.TableName()).
			Where("comment_id", commentID).
			Where("delete_time", 0)
		if userID != "" {
			queryTx = queryTx.Where("user_id", userID)
		}
		if officialID != 0 {
			queryTx = queryTx.Where("official_id", officialID)
		}
		if err = queryTx.Take(&articleLike).Error; err != gorm.ErrRecordNotFound {
			return err
		}
		newLike := db.ArticleCommentLikeSQL{
			CommentID:  commentID,
			UserID:     userID,
			OfficialID: officialID,
			CreatedBy:  opUserID,
			CreateTime: time.Now().Unix(),
			Status:     1,
		}
		if err = tx.Table(newLike.TableName()).Create(&newLike).Error; err != nil {
			return err
		}
		return tx.Table(db.ArticleCommentSQL{}.TableName()).Where("comment_id", commentID).Update("like_counts", gorm.Expr("like_counts + 1")).Error
	})
}

func DeleteArticleCommentLike(commentID int64, userID string, officialID int64, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		update := map[string]interface{}{
			"deleted_by":  opUserID,
			"delete_time": time.Now().Unix(),
		}
		delTx := tx.Table(db.ArticleCommentLikeSQL{}.TableName()).
			Where("comment_id", commentID).
			Where("delete_time", 0)

		if userID != "" {
			delTx = delTx.Where("user_id", userID)
		}
		if officialID != 0 {
			delTx = delTx.Where("official_id", officialID)
		}
		if delTx = delTx.Updates(update); delTx.Error != nil {
			return delTx.Error
		} else if delTx.RowsAffected == 0 {
			return nil
		}
		return tx.Table(db.ArticleCommentSQL{}.TableName()).Where("comment_id", commentID).Update("like_counts", gorm.Expr("like_counts - 1")).Error
	})
}

func DeleteArticleCommentLikesByUserID(userIDList []string, opUserID string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		var commentLikes []db.ArticleCommentLikeSQL
		err = tx.Table(db.ArticleCommentLikeSQL{}.TableName()).Where("user_id", userIDList).Find(&commentLikes).Error
		if err != nil {
			log.NewError("", "get article comment like error")
			return err
		}
		commentLike := db.ArticleCommentLikeSQL{
			DeleteTime: time.Now().Unix(),
			DeletedBy:  opUserID,
		}
		i += tx.Table(db.ArticleCommentLikeSQL{}.TableName()).Where("user_id", userIDList).Updates(&commentLike).RowsAffected
		for _, like := range commentLikes {
			articleComment := db.ArticleComment{CommentID: like.CommentID, ReplyCounts: 0, LikeCounts: -1}
			commentSQL := db.ArticleCommentSQL{CommentID: like.CommentID, ReplyCounts: 0, LikeCounts: -1}
			if err = db.DB.UpdateArticleCommentV2(articleComment); err != nil {
				errMsg := "Update Article count on mongodb failed " + err.Error()
				log.NewError(errMsg)
			}
			if _, err = UpdateArticleCommentV2(&commentSQL); err != nil {
				errMsg := "Update Article count on mysql failed " + err.Error()
				log.NewError(errMsg)
			}
		}

		return nil
	})

	return i
}

type GetIdsByCommentIDResult struct {
	CommentUserID         string `gorm:"comment_user_id"`
	CommentOfficialUserID string `gorm:"comment_official_user_id"`
	ArticleOfficialID     int64  `gorm:"article_official_id"`
}

func GetCommentByCommentID(commentID int64) (db.ArticleCommentSQL, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return db.ArticleCommentSQL{}, nil
	}
	var comment db.ArticleCommentSQL
	err = dbConn.Table(comment.TableName()).Where("comment_id", commentID).
		Where("delete_time", 0).
		Take(&comment).Error
	return comment, err
}

func OfficialDeleteArticleComment(commentID, parentCommentID, articleID, articleOfficialID int64, commentUserGender int32, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		update := map[string]interface{}{
			"deleted_by":  opUserID,
			"delete_time": time.Now().Unix(),
		}

		delTx := tx.Table(db.ArticleCommentSQL{}.TableName()).
			Where(tx.Where("comment_id", commentID).Or("parent_comment_id", commentID)).
			Where("delete_time", 0).
			Updates(update)

		if delTx.Error != nil {
			return delTx.Error
		}

		if delTx.RowsAffected == 0 {
			return nil
		}

		if parentCommentID != 0 {
			err = tx.Table(db.ArticleCommentSQL{}.TableName()).Where("comment_id", parentCommentID).Update("reply_counts", gorm.Expr("reply_counts - 1")).Error
			if err != nil {
				return err
			}
		}

		err = tx.Table(db.ArticleSQL{}.TableName()).Where("article_id", articleID).Where("delete_time", 0).Update("comment_counts", gorm.Expr("comment_counts - ?", delTx.RowsAffected)).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"comment_counts": gorm.Expr("comment_counts - ?", delTx.RowsAffected)}),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&db.OfficialAnalytics{
				OfficialID:    articleOfficialID,
				Time:          utils.FloorTimeToHours(time.Now()),
				Gender:        commentUserGender,
				CommentCounts: -delTx.RowsAffected,
			}).Error
	})
}

func OfficialHideArticleComment(commentID int64, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"status":      2,
		"updated_by":  opUserID,
		"update_time": time.Now().Unix(),
	}

	return dbConn.Table(db.ArticleCommentSQL{}.TableName()).
		Where("comment_id", commentID).
		Where("status", 1).
		Where("delete_time", 0).
		Updates(update).Error
}

func OfficialShowArticleComment(commentID int64, opUserID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"status":      1,
		"updated_by":  opUserID,
		"update_time": time.Now().Unix(),
	}

	return dbConn.Table(db.ArticleCommentSQL{}.TableName()).
		Where("comment_id", commentID).
		Where("status", 2).
		Where("delete_time", 0).
		Updates(update).Error
}

func GetOfficialByUserName(officialName string) ([]*db.Official, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var officialList []*db.Official
	if err != nil {
		return officialList, err
	}
	err = dbConn.Table("official").Where("(nickname != '' AND nickname LIKE ?) OR (nickname = '' AND initial_nickname LIKE ?)",
		"%"+officialName+"%", "%"+officialName+"%").Find(&officialList).Error
	return officialList, err
}

func GetOfficialByIds(officialIDs []int64) ([]db.Official, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var officialList []db.Official
	if err != nil {
		return officialList, err
	}
	err = dbConn.Table("official").Where("id", officialIDs).Find(&officialList).Error
	return officialList, err
}

func GetNewsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.ArticleRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var articleResList []*db.ArticleRes
	var count int64
	if err != nil {
		return articleResList, count, err
	}
	var officialList []db.Official
	officialMap := make(map[int64]*db.Official)
	var officialIdList []int64
	var articleList []db.ArticleSQL
	sortMap := map[string]string{}
	timeTypeMap := map[int]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"
	timeTypeMap[1] = "create_time"

	log.Debug("", "query news")

	// query official account
	queryOfficialAccount := false
	dbSub := dbConn.Table(db.Official{}.TableName())
	if officialAccount, ok := where["official_account"]; ok {
		if officialAccount != "" {
			queryOfficialAccount = true
			dbSub = dbSub.Where("nickname like ?", "%"+officialAccount+"%")
		}
	}
	if accountType, ok := where["account_type"]; ok {
		if accountType != "" && accountType != "0" {
			queryOfficialAccount = true
			dbSub = dbSub.Where("type=?", accountType)
		}
	}
	if ip, ok := where["ip"]; ok {
		if ip != "" {
			queryOfficialAccount = true
			dbSub = dbSub.Where("last_login_ip like ?", "%"+ip+"%")
		}
	}
	if queryOfficialAccount {
		dbSub.Debug().Find(&officialList)
		for _, official := range officialList {
			officialIdList = append(officialIdList, official.Id)
			officialMap[official.Id] = &official
		}
	}

	// query article
	dbSub = dbConn.Table(db.ArticleSQL{}.TableName()).Where("delete_time=0")
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

	startTime := where["start_time"]
	endTime := where["end_time"]
	timeTypeStr := where["time_type"]
	timeType, _ := strconv.Atoi(timeTypeStr)
	timeType = 1
	if startTime != "" {
		dbSub = dbSub.Where(fmt.Sprintf("%s>=?", timeTypeMap[timeType]), startTime)
	}
	if endTime != "" {
		dbSub = dbSub.Where(fmt.Sprintf("%s<=?", timeTypeMap[timeType]), endTime)
	}

	if queryOfficialAccount {
		dbSub = dbSub.Where("official_id IN (?)", officialIdList)
	}
	if title, ok := where["title"]; ok {
		if title != "" {
			dbSub = dbSub.Where("title like ?", "%"+title+"%")
		}
	}
	log.Debug("", "orderByClause: ", orderByClause)
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Debug().Find(&articleList)

	// if don't have conditions of official account, need to query official account by article list.
	if !queryOfficialAccount && len(articleList) > 0 {
		var officialAccountIdList []int64
		for _, article := range articleList {
			officialAccountIdList = append(officialAccountIdList, article.OfficialID)
		}
		dbConn.Table(db.Official{}.TableName()).Where("id IN (?)", officialAccountIdList).Find(&officialList)
		for index, official := range officialList {
			officialIdList = append(officialIdList, official.Id)
			officialMap[official.Id] = &officialList[index]
		}
	}

	// link official to article.
	for _, article := range articleList {
		articleRes := &db.ArticleRes{}
		_ = utils.CopyStructFields(articleRes, &article)
		if official, ok := officialMap[article.OfficialID]; ok {
			articleRes.LastLoginTime = official.LastLoginTime
			articleRes.LastLoginIp = official.LastLoginIp
			articleRes.OfficialType = int32(official.Type)
			articleRes.OfficialStatus = official.Status
			if official.DeleteTime != 0 {
				articleRes.OfficialStatus = 3
			}
		}
		articleResList = append(articleResList, articleRes)
	}

	return articleResList, count, nil
}

func DeleteArticles(articles []int64, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}

	nowTime := time.Now().Unix()
	i = 0

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		for index, _ := range articles {
			article := db.Article{
				ArticleID:  articles[index],
				DeletedBy:  opUserId,
				DeleteTime: nowTime,
			}
			articleTemp := db.Article{}
			tx.Table(db.ArticleSQL{}.TableName()).Where("article_id=?", article.ArticleID).First(&articleTemp)
			i += tx.Table(db.ArticleSQL{}.TableName()).Where("article_id=?", article.ArticleID).Debug().Updates(&article).RowsAffected

			officialMap := map[string]interface{}{
				"post_counts": gorm.Expr("post_counts - 1"),
				"like_counts": gorm.Expr("like_counts - ?", articleTemp.LikeCounts),
			}

			tx.Table(db.Official{}.TableName()).
				Where("id", articleTemp.OfficialID).Debug().
				Updates(officialMap)
		}
		return nil
	})
	if err != nil {
		return 0
	}

	return i

}

func AlterArticle(article *db.ArticleSQL, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	article.UpdateTime = time.Now().Unix()
	article.UpdatedBy = opUserId
	i = dbConn.Table(db.ArticleSQL{}.TableName()).Debug().Updates(&article).RowsAffected
	return i
}

func GetArticleCommentsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.ArticleCommentRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var articleComment []*db.ArticleCommentRes
	var count int64
	if err != nil {
		return articleComment, count, err
	}

	var commentUser []db.User
	var commentUserMap = make(map[string]*db.User, 0)
	var commentUserIdList []string

	var commonOfficial []*db.Official
	var commonOfficialNameMap = make(map[int64]*db.Official, 0)
	var commonOfficialIdList []int64

	var officialList []db.Official
	var officialMap = make(map[int64]*db.Official)
	var officialIdList []int64

	var articleList []db.Article
	var articleMap = make(map[int64]*db.Article)
	var articleIdList []int64

	var articleComments []db.ArticleCommentSQL

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"

	// Query comment user
	queryCommentUser := false
	if commentUserKey, ok := where["comment_user"]; ok {
		if commentUserKey != "" {
			queryCommentUser = true
			commentUser = GetUserByAllCondition(commentUserKey)
			for index, user := range commentUser {
				commentUserMap[user.UserID] = &commentUser[index]
				commentUserIdList = append(commentUserIdList, user.UserID)
			}
			commonOfficial, err = GetOfficialByUserName(commentUserKey)
			if err != nil {
				return nil, count, err
			}
			for index, official := range commonOfficial {
				commonOfficialNameMap[official.Id] = commonOfficial[index]
				commonOfficialIdList = append(commonOfficialIdList, official.Id)
			}
		}
	}

	// Query official account list.
	dbSub := dbConn.Table(db.Official{}.TableName())
	queryOfficial := false
	if officialAccount, ok := where["official_account"]; ok {
		if officialAccount != "" {
			queryOfficial = true
			dbSub = dbConn.Where("nickname like ?", "%"+officialAccount+"%")
		}
	}
	if accountType, ok := where["account_type"]; ok {
		if accountType != "" && accountType != "0" {
			accountTypeInt, _ := strconv.Atoi(accountType)
			queryOfficial = true
			dbSub = dbSub.Where("type=?", accountTypeInt)
		}
	}
	if ip, ok := where["ip"]; ok {
		if ip != "" {
			queryOfficial = true
			dbSub = dbSub.Where("last_login_ip=?", ip)
		}
	}
	if queryOfficial {
		dbSub.Find(&officialList)
		for index, official := range officialList {
			officialMap[official.Id] = &officialList[index]
			officialIdList = append(officialIdList, official.Id)
		}
	}

	// Query articles by official id list, title and create time
	dbSub = dbConn.Table(db.ArticleSQL{}.TableName())
	queryArticle := false
	if title, ok := where["title"]; ok {
		if title != "" {
			queryArticle = true
			dbSub = dbSub.Where("title like ?", "%"+title+"%")
		}
	}
	if queryOfficial {
		queryArticle = true
		dbSub = dbSub.Where("official_id IN (?)", officialIdList)
	}
	if articleId, ok := where["article_id"]; ok {
		if articleId != "" && articleId != "0" {
			queryArticle = true
			dbSub = dbSub.Where("article_id=?", articleId)
		}
	}
	if timeType, ok := where["time_type"]; ok {
		// publish time or all time
		if timeType != "" && (timeType == "1" || timeType == "0") {
			if startTime, ok := where["start_time"]; ok {
				if startTime != "" {
					queryArticle = true
					dbSub = dbSub.Where("create_time>=?", startTime)
				}
			}
			if endTime, ok := where["end_time"]; ok {
				if endTime != "" {
					queryArticle = true
					dbSub = dbSub.Where("create_time<=?", endTime)
				}
			}
		}
	}
	if queryArticle {
		dbSub.Find(&articleList)
		for index, article := range articleList {
			articleMap[article.ArticleID] = &articleList[index]
			articleIdList = append(articleIdList, article.ArticleID)
		}
	}

	// Query comments by user id, user name, articles id, comment key.
	dbSub = dbConn.Table(db.ArticleCommentSQL{}.TableName()).Where("delete_time=0")
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
	if timeType, ok := where["time_type"]; ok {
		// publish time or all time
		log.Debug("", "timeType: ", timeType)
		if timeType != "" && (timeType == "2" || timeType == "0") {
			if startTime, ok := where["start_time"]; ok {
				log.Debug("", "startTime: ", startTime)
				if startTime != "" {
					dbSub = dbSub.Where("create_time>=?", startTime)
				}
			}
			if endTime, ok := where["end_time"]; ok {
				log.Debug("", "endTime: ", endTime)
				if endTime != "" {
					dbSub = dbSub.Where("create_time<=?", endTime)
				}
			}
		}
	}
	if queryArticle {
		dbSub = dbSub.Where("article_id IN (?)", articleIdList)
	}
	if queryCommentUser {
		dbSub = dbSub.Where("user_id IN (?) OR official_id IN (?)", commentUserIdList,
			commonOfficialIdList)
	}
	if commentKey, ok := where["comment_key"]; ok {
		if commentKey != "" {
			dbSub = dbSub.Where("content like ?", "%"+commentKey+"%")
		}
	}
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}
	log.Debug("", " comment query...")
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&articleComments)

	// If don't have article and official query conditions, there need to query information by article comments.
	if !queryOfficial || !queryArticle || !queryCommentUser {
		// Get official and article id list.
		for _, comment := range articleComments {
			if !queryArticle {
				articleIdList = append(articleIdList, comment.ArticleID)
			}
			if !queryCommentUser {
				commentUserIdList = append(commentUserIdList, comment.UserID)
				commonOfficialIdList = append(commonOfficialIdList, comment.OfficialID)
			}
		}
		if !queryArticle {
			dbConn.Table(db.ArticleSQL{}.TableName()).Where("article_id IN (?)", articleIdList).Find(&articleList)
			for index, article := range articleList {
				articleMap[article.ArticleID] = &articleList[index]
				articleIdList = append(articleIdList, article.ArticleID)
			}
		}

		if !queryOfficial {
			for _, article := range articleList {
				officialIdList = append(officialIdList, article.OfficialID)
			}

			tempMap, _ := GetOfficialsByOfficialIDList(officialIdList)
			for k, _ := range tempMap {
				official := tempMap[k]
				officialMap[k] = &official
			}
		}
		if !queryCommentUser {
			dbConn.Table(db.User{}.TableName()).Where("user_id IN (?)", commentUserIdList).Find(&commentUser)
			for index, user := range commentUser {
				commentUserMap[user.UserID] = &commentUser[index]
			}
			tempMap, _ := GetOfficialsByOfficialIDList(commonOfficialIdList)
			for k, v := range tempMap {
				commonOfficialNameMap[k] = &v
			}
		}
	}

	// Link official, user, article to comments.
	for _, comment := range articleComments {
		commentRes := &db.ArticleCommentRes{}
		_ = utils.CopyStructFields(&commentRes, &comment)

		log.Debug("user id: ", comment.UserID, " official id: ", comment.OfficialID)
		if user, ok := commentUserMap[comment.UserID]; ok {
			commentRes.UserName = user.Nickname
			commentRes.UserID = user.UserID
			commentRes.CreatedBy = user.UserID
			fmt.Println(user.DeleteTime)
			if user.DeleteTime != 0 {
				commentRes.UserName = ""
			}
		} else {
			if official, ok := commonOfficialNameMap[comment.OfficialID]; ok {
				commentRes.UserID = official.UserID
				commentRes.UserName = official.Nickname
				commentRes.CreatedBy = official.UserID
				if official.Nickname == "" {
					commentRes.UserName = official.InitialNickname
				}
				if official.DeleteTime != 0 {
					commentRes.UserName = ""
				}
			}
		}

		if article, ok := articleMap[comment.ArticleID]; ok {
			if article != nil {
				commentRes.ArticleTitle = article.Title
				commentRes.CoverPhoto = article.CoverPhoto
				commentRes.PostTime = article.CreateTime
				commentRes.DeletedBy = article.DeletedBy
				commentRes.DeleteTime = article.DeleteTime
				if official, ok := officialMap[article.OfficialID]; ok {
					if official != nil {
						commentRes.LastLoginTime = official.LastLoginTime
						commentRes.LastLoginIp = official.LastLoginIp
						commentRes.OfficialType = official.Type
						commentRes.OfficialName = official.Nickname
						if official.Nickname == "" {
							commentRes.OfficialName = official.InitialNickname
						}
					}
				}
			}
		}

		commentRes.CommentLikes = comment.LikeCounts
		articleComment = append(articleComment, commentRes)
	}

	return articleComment, count, nil
}

func DeleteArticleComments(comments []int64, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	dbComment := db.ArticleCommentSQL{
		DeleteTime: time.Now().Unix(),
		DeletedBy:  opUserId,
	}
	i = dbConn.Table(db.ArticleCommentSQL{}.TableName()).
		Where("comment_id IN (?)", comments).
		Where("delete_time=0").
		Updates(&dbComment).RowsAffected
	return i
}

func AlterNewsComment(comment *db.ArticleCommentSQL, OpUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	comment.UpdateTime = time.Now().Unix()
	comment.UpdatedBy = OpUserId

	i = dbConn.Table(db.ArticleCommentSQL{}.TableName()).Where("comment_id=?", comment.CommentID).Updates(&comment).RowsAffected
	return i
}

func GetLikesByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.ArticleLikesRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var articleLikes []*db.ArticleLikesRes
	var count int64
	if err != nil {
		return articleLikes, count, err
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"

	var officialList []db.Official
	var officialMap = make(map[int64]*db.Official)
	var officialIdList []int64

	var articleList []db.Article
	var articleMap = make(map[int64]*db.Article)
	var articleIdList []int64

	var articleLikeList []db.ArticleLikeSQL

	// Query official account by official account name, official type, ip.
	dbSub := dbConn.Table(db.Official{}.TableName())
	queryOfficial := false
	if officialAccount, ok := where["official_account"]; ok {
		if officialAccount != "" {
			queryOfficial = true
			dbSub = dbSub.Where("nickname like ?", "%"+officialAccount+"%")
		}
	}
	if accountType, ok := where["account_type"]; ok {
		accountTypeInt, _ := strconv.Atoi(accountType)
		if accountTypeInt != 0 {
			queryOfficial = true
			dbSub = dbSub.Where("type=?", accountTypeInt)
		}
	}
	if ip, ok := where["ip"]; ok {
		if ip != "" {
			dbSub = dbSub.Where("last_login_ip=?", ip)
		}
	}
	if queryOfficial {
		dbSub.Find(&officialList)
		for index, official := range officialList {
			officialMap[official.Id] = &officialList[index]
			officialIdList = append(officialIdList, official.Id)
		}
	}

	// Query articles by create time, title, official id
	dbSub = dbConn.Table(db.ArticleSQL{}.TableName())
	queryArticle := false
	if queryOfficial {
		queryArticle = true
		dbSub = dbSub.Where("official_id IN (?)", officialIdList)
	}
	if articleId, ok := where["article_id"]; ok {
		if articleId != "" && articleId != "0" {
			queryArticle = true
			dbSub = dbSub.Where("article_id=?", articleId)
		}
	}
	if title, ok := where["title"]; ok {
		if title != "" {
			queryArticle = true
			dbSub = dbSub.Where("title like ?", "%"+title+"%")
		}
	}
	if timeType, ok := where["time_type"]; ok {
		// publish time or all time
		if timeType != "" && (timeType == "1" || timeType == "0") {
			if startTime, ok := where["start_time"]; ok {
				if startTime != "" {
					queryArticle = true
					dbSub = dbSub.Where("create_time>=?", startTime)
				}
			}
			if endTime, ok := where["end_time"]; ok {
				if endTime != "" {
					queryArticle = true
					dbSub = dbSub.Where("create_time<=?", endTime)
				}
			}
		}
	}

	if queryArticle {
		dbSub.Find(&articleList)
		for index, article := range articleList {
			articleMap[article.ArticleID] = &articleList[index]
			articleIdList = append(articleIdList, article.ArticleID)
			if !queryOfficial {
				officialIdList = append(officialIdList, article.OfficialID)
			}
		}
	}

	// Query likes by article id list, like user id and name.
	dbSub = dbConn.Table(db.ArticleLikeSQL{}.TableName()).Where("delete_time=0")
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

	if timeType, ok := where["time_type"]; ok {
		// publish time or all time
		if timeType != "" && (timeType == "2" || timeType == "0") {
			if startTime, ok := where["start_time"]; ok {
				if startTime != "" {
					dbSub = dbSub.Where("create_time>=?", startTime)
				}
			}
			if endTime, ok := where["end_time"]; ok {
				if endTime != "" {
					dbSub = dbSub.Where("create_time<=?", endTime)
				}
			}
		}
	}
	if queryArticle {
		dbSub = dbSub.Where("article_id IN (?)", articleIdList)
	}
	if likeUser, ok := where["like_user"]; ok {
		if likeUser != "" {
			dbSub = dbSub.Where("user_id like ? OR user_name like ?", "%"+likeUser+"%", "%"+likeUser+"%")
		}
	}
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}

	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&articleLikeList)

	// Link like res.
	// If don't have article and official query conditions, there need to query information by article comments.
	if !queryOfficial || !queryArticle {
		// Get official and article id list.
		for _, likes := range articleLikeList {
			if !queryArticle {
				articleIdList = append(articleIdList, likes.ArticleID)
			}
		}
		if !queryArticle {
			dbConn.Table(db.ArticleSQL{}.TableName()).Where("article_id IN (?)", articleIdList).Find(&articleList)
			for index, article := range articleList {
				articleMap[article.ArticleID] = &articleList[index]
				articleIdList = append(articleIdList, article.ArticleID)
				officialIdList = append(officialIdList, article.OfficialID)
			}
		}
		if !queryOfficial {
			dbConn.Table(db.Official{}.TableName()).Where("id IN (?)", officialIdList).Find(&officialList)
			for index, official := range officialList {
				officialMap[official.Id] = &officialList[index]
			}
		}
	}

	// Link official, user, article to comments.
	for _, like := range articleLikeList {
		likeRes := &db.ArticleLikesRes{}
		_ = utils.CopyStructFields(&likeRes, like)
		if article, ok := articleMap[like.ArticleID]; ok {
			likeRes.ArticleTitle = article.Title
			likeRes.PostTime = article.CreateTime
			likeRes.CoverPhoto = article.CoverPhoto
			likeRes.DeletedBy = article.DeletedBy
			likeRes.DeleteTime = article.DeleteTime
			if official, ok := officialMap[article.OfficialID]; ok {
				likeRes.LastLoginIp = official.LastLoginIp
				likeRes.OfficialType = official.Type
				likeRes.OfficialName = official.Nickname
			}
		}

		articleLikes = append(articleLikes, likeRes)
	}

	return articleLikes, count, nil
}

func DeleteArticleLikes(userIdList []string, articleIdList []int64, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	for index, _ := range userIdList {
		like := db.ArticleLikeSQL{
			DeleteTime: time.Now().Unix(),
			DeletedBy:  opUserId,
		}
		i += dbConn.Table(db.ArticleLikeSQL{}.TableName()).
			Where("article_id=?", articleIdList[index]).
			Where("user_id=?", userIdList[index]).
			Where("delete_time=0").
			Updates(&like).RowsAffected
	}
	return i
}

func DeleteArticleLikesByUserID(userIdList []string, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	nowTime := time.Now().Unix()
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		var likes []db.ArticleLikeSQL
		err = tx.Table(db.ArticleLikeSQL{}.TableName()).Where("user_id IN (?)", userIdList).Find(&likes).Error
		if err != nil {
			return err
		}

		articleLike := db.ArticleLikeSQL{
			DeleteTime: nowTime,
			DeletedBy:  opUserId,
		}
		i += tx.Table(db.ArticleLikeSQL{}.TableName()).Where("user_id IN (?)", userIdList).Updates(&articleLike).RowsAffected

		for _, like := range likes {
			articleMongo := db.Article{ArticleID: like.ArticleID, LikeCounts: -1, CommentCounts: 0,
				RepostCounts: 0, ReadCounts: 0, UniqueReadCounts: 0}
			article := db.ArticleSQL{ArticleID: like.ArticleID, LikeCounts: -1, CommentCounts: 0,
				RepostCounts: 0, ReadCounts: 0, UniqueReadCounts: 0}
			if err = db.DB.UpdateArticleV2(articleMongo); err != nil {
				errMsg := "Update Article count on mongodb failed " + err.Error()
				log.NewError(errMsg)
			}
			if _, err = UpdateArticleV2(&article); err != nil {
				errMsg := "Update Article count on mysql failed " + err.Error()
				log.NewError(errMsg)
			}
		}
		return nil
	})

	return i
}

func ChangeArticleLikes(like *db.ArticleLikeSQL, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	like.UpdateTime = time.Now().Unix()
	like.UpdatedBy = opUserId
	i = dbConn.Table(db.ArticleLikeSQL{}.TableName()).Where("article_id=?", like.ArticleID).Where("user_id=?", like.UserID).Updates(&like).RowsAffected
	return i
}

func GetRepostArticles(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.ArticlePostRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var articleRepost []*db.ArticlePostRes
	var count int64
	if err != nil {
		return articleRepost, count, err
	}

	sortMap := map[string]string{}
	var orderByClause string
	sortMap["start_time"] = "m_create_time"

	var sharedUserList []db.User
	var sharedUserMap = make(map[string]*db.User)
	var sharedUserId []string

	var originalUserId []string

	var officialList []db.Official
	var officialMap = make(map[int64]*db.Official)
	var officialIdList []int64

	var articleList []db.Article
	var articleMap = make(map[int64]*db.Article)
	var articleIdList []int64

	var momentList []db.MomentSQL

	// Query Original User id list by original account
	queryOriginUserId := false
	if originalUser, ok := where["original_user"]; ok {
		if originalUser != "" {
			queryOriginUserId = true
			dbConn.Table(db.User{}.TableName()).Select("user_id").
				Where("user_id like ? OR name like ?", "%"+originalUser+"%", "%"+originalUser+"%").
				Find(&originalUserId)
		}
	}

	// Query shared user id
	querySharedUserId := false
	dbSub := dbConn.Table(db.User{}.TableName())
	if ip, ok := where["ip"]; ok {
		if ip != "" {
			querySharedUserId = true
			dbSub = dbSub.Where("update_ip LIKE ?", "%"+ip+"%")
		}
	}
	if repostUser, ok := where["repost_user"]; ok {
		if repostUser != "" {
			querySharedUserId = true
			dbSub = dbSub.Where("user_id like ? OR name like ?", "%"+repostUser+"%", "%"+repostUser+"%")
		}
	}

	if querySharedUserId {
		dbSub.Find(&sharedUserList)
		for _, user := range sharedUserList {
			sharedUserId = append(sharedUserId, user.UserID)
			sharedUserMap[user.UserID] = &user
		}
	}

	// Query official account list by type, ip, userid, user name
	queryOfficial := false
	dbSub = dbConn.Table(db.Official{}.TableName())
	if queryOriginUserId {
		dbSub = dbSub.Where("user_id IN (?)", originalUserId)
	}
	if accountType, ok := where["account_type"]; ok {
		accountTypeInt, _ := strconv.Atoi(accountType)
		if accountTypeInt != 0 {
			queryOfficial = true
			dbSub = dbSub.Where("type=?", accountTypeInt)
		}
	}
	if queryOfficial {
		dbSub.Find(&officialList)
		for index, official := range officialList {
			officialMap[official.Id] = &officialList[index]
			officialIdList = append(officialIdList, official.Id)
		}
	}

	// Query article list by title, official id, create time
	dbSub = dbConn.Table(db.ArticleSQL{}.TableName())
	queryArticle := false
	if queryOfficial {
		queryArticle = true
		dbSub = dbSub.Where("official_id IN (?)", officialIdList)
	}
	if articleId, ok := where["article_id"]; ok {
		if articleId != "" && articleId != "0" {
			queryArticle = true
			dbSub = dbSub.Where("article_id=?", articleId)
		}
	}
	if title, ok := where["title"]; ok {
		if title != "" {
			queryArticle = true
			dbSub = dbSub.Where("title like ?", "%"+title+"%")
		}
	}
	if startTime, ok := where["start_time"]; ok {
		if startTime != "" {
			queryArticle = true
			dbSub = dbSub.Where("m_create_time>=?", startTime)
		}
	}
	if endTime, ok := where["end_time"]; ok {
		if endTime != "" {
			queryArticle = true
			dbSub = dbSub.Where("m_create_time<=?", endTime)
		}
	}

	if queryArticle {
		dbSub.Debug().Find(&articleList)
		for index, article := range articleList {
			articleMap[article.ArticleID] = &articleList[index]
			articleIdList = append(articleIdList, article.ArticleID)
		}
	}

	// Query repost by user id, user name, Original id
	dbSub = dbConn.Table(db.MomentSQL{}.TableName()).Where("delete_time=0").Where("article_id!=0")
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

	if querySharedUserId {
		dbSub = dbSub.Where("creator_id IN (?)", sharedUserId)
	}
	if queryArticle {
		dbSub = dbSub.Where("article_id IN (?)", articleIdList)
	}
	if orderByClause != "" {
		dbSub = dbSub.Order(orderByClause)
	}

	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Debug().Find(&momentList)

	log.Debug("xxx", momentList, "moment len: ", len(momentList))

	// If don't have article and official query conditions, there need to query information by moments.
	if !querySharedUserId || !queryOfficial || !queryArticle {
		// Get official and article id list.
		for _, moment := range momentList {
			if !querySharedUserId {
				sharedUserId = append(sharedUserId, moment.CreatorID)
			}
			if !queryArticle {
				articleIdList = append(articleIdList, moment.ArticleID)
			}
		}
		if !querySharedUserId {
			dbConn.Table(db.User{}.TableName()).Where("user_id IN (?)", sharedUserId).Find(&sharedUserList)
			for index, user := range sharedUserList {
				sharedUserMap[user.UserID] = &sharedUserList[index]
			}
		}
		if !queryArticle {
			dbConn.Table(db.Article{}.TableName()).Where("article_id IN (?)", articleIdList).Debug().Find(&articleList)
			for index, article := range articleList {
				articleMap[article.ArticleID] = &articleList[index]
				articleIdList = append(articleIdList, article.ArticleID)
				if !queryOfficial {
					officialIdList = append(officialIdList, article.OfficialID)
				}
			}
		}
		if !queryOfficial {
			dbConn.Table(db.Official{}.TableName()).Where("id IN (?)", officialIdList).Find(&officialList)
			for index, official := range officialList {
				officialMap[official.Id] = &officialList[index]
				officialIdList = append(officialIdList, official.Id)
			}
		}

	}

	log.Debug("", "...articleMap: ", articleMap)
	log.Debug("", "sharedUserMap: ", sharedUserMap)
	// Link info return
	for _, moment := range momentList {
		repostRes := &db.ArticlePostRes{
			MomentId:      moment.MomentID,
			ShareUser:     moment.UserName,
			CommentCounts: int64(moment.MCommentsCount),
			LikeCounts:    int64(moment.MLikesCount),
			ShareTime:     moment.MCreateTime,
			Privacy:       moment.Privacy,
			ArticleId:     moment.ArticleID,
		}

		if article, ok := articleMap[moment.ArticleID]; ok {
			repostRes.ArticleTitle = article.Title
			repostRes.CoverPhoto = article.CoverPhoto
			repostRes.DeletedBy = article.DeletedBy
			repostRes.DeleteTime = article.DeleteTime
			if official, ok := officialMap[article.OfficialID]; ok {
				repostRes.OfficialType = int32(official.Type)
				repostRes.OriginalUser = official.Nickname
			}
		}
		if user, ok := sharedUserMap[moment.CreatorID]; ok {
			log.Debug("", "creator id: ", moment.CreatorID, "user UpdateIp: ", user.UpdateIp)
			repostRes.LastLoginIp = user.UpdateIp
		}

		articleRepost = append(articleRepost, repostRes)
	}

	return articleRepost, count, nil
}

func GetRepostCountsArticles(_ *db.MomentSQL, where map[string]string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var count int64 = 0
	if err != nil {
		return count, err
	}

	var originalUserId []string
	var officialIdList []int64
	var articleIdList []int64

	// Query User id list
	queryOriginUserId := false
	if originalUser, ok := where["original_user"]; ok {
		if originalUser != "" {
			queryOriginUserId = true
			dbConn.Table(db.User{}.TableName()).Select("user_id").
				Where("user_id like ? OR name like ?", "%"+originalUser+"%", "%"+originalUser+"%").
				Find(&originalUserId)
		}
	}

	// Query official account list by type, ip, userid, user name
	queryOfficial := false
	dbSub := dbConn.Table(db.Official{}.TableName()).Select("id")
	if queryOriginUserId {
		dbSub = dbSub.Where("user_id IN (?)", originalUserId)
	}
	if accountType, ok := where["original_user_type"]; ok {
		accountTypeInt, _ := strconv.Atoi(accountType)
		if accountTypeInt != 0 {
			queryOfficial = true
			dbSub = dbSub.Where("type=?", accountTypeInt)
		}
	}
	if ip, ok := where["ip"]; ok {
		if ip != "" {
			dbSub = dbSub.Where("last_login_ip=?", ip)
		}
	}
	if queryOfficial {
		dbSub.Find(&officialIdList)
	}

	// Query article list by title, official id, create time
	dbSub = dbConn.Table(db.Article{}.TableName()).Select("article_id")
	queryArticle := false
	if queryOfficial {
		queryArticle = true
		dbSub = dbSub.Where("official_id IN (?)", officialIdList)
	}
	if articleId, ok := where["article_id"]; ok {
		if articleId != "" && articleId != "0" {
			queryArticle = true
			dbSub = dbSub.Where("article_id=?", articleId)
		}
	}
	if title, ok := where["title"]; ok {
		if title != "" {
			queryArticle = true
			dbSub = dbSub.Where("title like ?", "%"+title+"%")
		}
	}
	if startTime, ok := where["start_time"]; ok {
		if startTime != "" {
			queryArticle = true
			dbSub = dbSub.Where("m_create_time>=?", startTime)
		}
	}
	if endTime, ok := where["end_time"]; ok {
		if endTime != "" {
			queryArticle = true
			dbSub = dbSub.Where("m_create_time<=?", endTime)
		}
	}

	if queryArticle {
		dbSub.Find(&articleIdList)
	}

	// Query repost by user id, user name, Original id
	dbSub = dbConn.Table(db.MomentSQL{}.TableName()).Where("delete_time=0").Where("moment_type=1")
	if repostUser, ok := where["repost_user"]; ok {
		if repostUser != "" {
			dbSub = dbSub.Where("user_id like ? OR user_name like ?", "%"+repostUser+"%", "%"+repostUser+"%")
		}
	}
	if queryArticle {
		dbSub = dbSub.Where("orignal_id IN (?)", articleIdList)
	}

	if err := dbSub.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func InsertArticle(article *db.ArticleSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		err = tx.Table(article.TableName()).Create(article).Error
		if err != nil {
			return err
		}

		return tx.Table(db.Official{}.TableName()).
			Where("id", article.OfficialID).
			Update("post_counts", gorm.Expr("post_counts + 1")).
			Error
	})
}

func UpdateArticle(article *db.ArticleSQL) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	articleMap := map[string]interface{}{
		"cover_photo":  article.CoverPhoto,
		"title":        article.Title,
		"content":      article.Content,
		"text_content": article.TextContent,
		"updated_by":   article.UpdatedBy,
		"update_time":  article.UpdateTime,
	}

	tx := dbConn.Table(article.TableName()).
		Where("article_id", article.ArticleID).
		Where("official_id", article.OfficialID).
		Where("delete_time", 0).
		Updates(articleMap)

	return tx.RowsAffected, tx.Error
}

func UpdateArticleV2(article *db.ArticleSQL) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var rowsAffected int64 = 0
	if err != nil {
		return rowsAffected, err
	}
	var err1 error

	specifyMap := make(map[string]interface{})
	specifyMap["unique_read_counts"] = gorm.Expr("unique_read_counts+(?)", article.UniqueReadCounts)
	specifyMap["read_counts"] = gorm.Expr("read_counts+(?)", article.ReadCounts)
	specifyMap["like_counts"] = gorm.Expr("like_counts+(?)", article.LikeCounts)
	specifyMap["comment_counts"] = gorm.Expr("comment_counts+(?)", article.CommentCounts)
	specifyMap["repost_counts"] = gorm.Expr("repost_counts+(?)", article.RepostCounts)

	tx := dbConn.Table(db.Article{}.TableName()).Debug().
		Where("article_id=?", article.ArticleID).Updates(specifyMap)

	//UpdateColumn("m_likes_count", gorm.Expr("m_likes_count+(?)", moment.MLikesCount)).
	//UpdateColumn("m_comments_count", gorm.Expr("m_comments_count+(?)", moment.MCommentsCount)).
	//UpdateColumn("m_repost_count", gorm.Expr("m_repost_count+(?)", moment.MRepostCount))
	err1 = tx.Error
	rowsAffected = tx.RowsAffected
	if err1 != nil {
		return rowsAffected, err1
	}

	return rowsAffected, err1
}

func DeleteArticle(article *db.ArticleSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	articleMap := map[string]interface{}{
		"deleted_by":  article.DeletedBy,
		"delete_time": article.DeleteTime,
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		err = tx.Table(db.ArticleSQL{}.TableName()).
			Where("article_id", article.ArticleID).
			Where("official_id", article.OfficialID).
			Where("delete_time", 0).
			Updates(articleMap).
			Error
		if err != nil {
			return err
		}

		officialMap := map[string]interface{}{
			"post_counts": gorm.Expr("post_counts - 1"),
			"like_counts": gorm.Expr("like_counts - ?", article.LikeCounts),
		}

		return tx.Table(db.Official{}.TableName()).
			Where("id", article.OfficialID).
			Updates(officialMap).
			Error
	})
}

func DeleteAllArticlesByOfficialId(officialId int64, OpUserId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	articleMap := map[string]interface{}{
		"deleted_by":  OpUserId,
		"delete_time": time.Now().Unix(),
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		err = tx.Table(db.ArticleSQL{}.TableName()).
			Where("official_id", officialId).
			Updates(articleMap).
			Error
		return err

	})
}

func ListOfficialArticles(officialID int64, minCreateTime int64, offset int, limit int) ([]db.ArticleSQL, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	var articles []db.ArticleSQL
	tx := dbConn.Table(db.ArticleSQL{}.TableName()).
		Where("official_id", officialID).
		Where("delete_time", 0).
		Where("create_time >= ?", minCreateTime).
		Order("article_id desc").
		Count(&count).
		Offset(offset).Limit(limit).
		Find(&articles)

	return articles, count, tx.Error
}

func GetArticleByArticleID(articleID int64) (db.ArticleSQL, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return db.ArticleSQL{}, err
	}

	article := db.ArticleSQL{ArticleID: articleID}
	err = dbConn.Table(article.TableName()).
		Where("delete_time", 0).
		Take(&article).Error

	return article, err
}

func GetLatestArticleByOfficialID(officialID int64) (db.ArticleSQL, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return db.ArticleSQL{}, err
	}
	article := db.ArticleSQL{}
	err = dbConn.Table(article.TableName()).
		Where("delete_time", 0).
		Where("official_id", officialID).
		Order("create_time desc").Limit(1).Find(&article).Error

	return article, err
}

func GetOfficialArticle(articleID, officialID int64) (db.ArticleSQL, error) {
	var article db.ArticleSQL
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return article, err
	}

	err = dbConn.Table(db.ArticleSQL{}.TableName()).
		Where("delete_time", 0).
		Where("article_id", articleID).
		Where("official_id", officialID).
		Take(&article).
		Error
	return article, err
}

func GetAllArticleByArticleID(articleID int64) (db.ArticleSQL, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return db.ArticleSQL{}, err
	}

	article := db.ArticleSQL{ArticleID: articleID}
	err = dbConn.Table(article.TableName()).
		Take(&article).Error

	return article, err
}

func ListOfficialFollows(officialID, minFollowTime, minBlockTime int64, blockedFilter *bool, orderBy, offset, limit int) ([]db.OfficialFollowSQL, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	var follows []db.OfficialFollowSQL
	dbSub := dbConn.Table(db.OfficialFollowSQL{}.TableName()).
		Where("official_id", officialID).
		Where("delete_time", 0)

	if minFollowTime > 0 {
		dbSub = dbSub.Where("follow_time >= ?", minFollowTime)
	}

	if minBlockTime > 0 {
		dbSub = dbSub.Where("block_time >= ?", minBlockTime)
	}

	if blockedFilter != nil {
		if *blockedFilter {
			dbSub = dbSub.Where("block_time > 0")
		} else {
			dbSub = dbSub.Where("block_time", 0)
		}
	}

	switch orderBy {
	case 0:
		dbSub = dbSub.Order("id desc")
	case 1:
		dbSub = dbSub.Order("block_time desc")
	}

	tx := dbSub.Count(&count).Offset(offset).Limit(limit).Find(&follows)

	return follows, count, tx.Error
}

type ArticleLikeEntry struct {
	ArticleID    int64  `gorm:"article_id"`
	UserID       string `gorm:"user_id"`
	UserNickname string `gorm:"user_nickname"`
	UserFaceURL  string `gorm:"user_face_url"`
	UserGender   int32  `gorm:"user_gender"`
	CreatedBy    string `gorm:"created_by"`
	CreateTime   int64  `gorm:"create_time"`
	UpdatedBy    string `gorm:"updated_by"`
	UpdateTime   int64  `gorm:"update_time"`
	DeletedBy    string `gorm:"deleted_by"`
	DeleteTime   int64  `gorm:"delete_time"`
	Status       int32  `gorm:"status"`
}

func ListArticleLikes(articleID, minCreateTime int64, keyword string, offset, limit int) ([]ArticleLikeEntry, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	var articleLikes []ArticleLikeEntry
	dbSub := dbConn.Table(db.ArticleLikeSQL{}.TableName()+" as l").
		Joins("left join "+db.User{}.TableName()+" as u on u.user_id = l.user_id and u.delete_time = 0").
		Select("l.article_id", "l.user_id", "u.name as user_nickname", "u.face_url as user_face_url", "u.gender as user_gender", "l.created_by", "l.update_time", "l.deleted_by", "l.delete_time", "l.status").
		Where("l.article_id", articleID).
		Where("l.delete_time", 0)

	if articleID > 0 {
		dbSub = dbSub.Where("l.create_time >= ?", minCreateTime)
	}

	if keyword != "" {
		dbSub = dbSub.Where("u.name LIKE ?", "%"+keyword+"%")
	}

	tx := dbSub.Order("l.create_time desc").Count(&count).Offset(offset).Limit(limit).Find(&articleLikes)

	return articleLikes, count, tx.Error
}

type commentEntry struct {
	CommentID                      int64  `gorm:"comment_id"`
	ParentCommentID                int64  `gorm:"parent_comment_id"`
	UserID                         string `gorm:"user_id"`
	UserNickname                   string `gorm:"user_nickname"`
	UserFaceURL                    string `gorm:"user_face_url"`
	OfficialID                     int64  `gorm:"official_id"`
	OfficialNickname               string `gorm:"official_nickname"`
	OfficialFaceURL                string `gorm:"official_face_url"`
	ReplyCommentID                 int64  `gorm:"reply_comment_id"`
	ReplyUserID                    string `gorm:"reply_user_id"`
	ReplyUserNickname              string `gorm:"reply_user_nickname"`
	ReplyUserFaceURL               string `gorm:"reply_user_face_url"`
	ReplyOfficialID                int64  `gorm:"reply_official_id"`
	ReplyOfficialNickname          string `gorm:"reply_official_nickname"`
	ReplyOfficialFaceURL           string `gorm:"reply_official_face_url"`
	ReplyCounts                    int64  `gorm:"reply_counts"`
	LikeCounts                     int64  `gorm:"like_counts"`
	Content                        string `gorm:"content"`
	LikeTime                       int64  `gorm:"like_time"`
	CreateTime                     int64  `gorm:"create_time"`
	Status                         int32  `gorm:"status"`
	UserDeleteTime                 int64  `gorm:"user_delete_time"`
	ReplyOfficialAccountDeleteTime int64  `gorm:"reply_official_account_delete_time"`
	ReplyUserDeleteTime            int64  `gorm:"reply_user_delete_time"`
}

type commentsCount struct {
	ParentCommentID int64 `gorm:"parent_comment_id"`
	Count           int64 `gorm:"count"`
}

func ListArticleComments(articleID int64, parentCommentIDList []int64, userID string, officialID int64, offset, limit int) (map[int64][]commentEntry, map[int64]int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, nil, err
	}

	var comments []commentEntry

	tx := dbConn.
		Table(db.ArticleCommentSQL{}.TableName()+" c").
		Select(
			"c.comment_id",
			"c.parent_comment_id",
			"u.user_id as user_id",
			"u.name as user_nickname",
			"u.face_url as user_face_url",
			"o.id as official_id",
			"o.nickname as official_nickname",
			"o.face_url as official_face_url",
			"ru.user_id as reply_user_id",
			"ru.name as reply_user_nickname",
			"ru.face_url as reply_user_face_url",
			"ro.id as reply_official_id",
			"ro.nickname as reply_official_nickname",
			"ro.face_url as reply_official_face_url",
			"c.reply_counts",
			"c.like_counts",
			"c.content",
			"c.create_time",
			"l.create_time as like_time",
			"c.status",
			"u.delete_time as user_delete_time",
			"ro.delete_time as reply_official_account_delete_time",
			"ru.delete_time as reply_user_delete_time",
		)
	if articleID != 0 {
		tx = tx.Where("c.article_id", articleID)
	}
	if parentCommentIDList != nil {
		tx = tx.Where("c.parent_comment_id", parentCommentIDList)
	}
	tx = tx.Where("c.delete_time", 0).
		Joins("LEFT JOIN "+db.User{}.TableName()+" u ON u.user_id = c.user_id").
		Joins("LEFT JOIN "+db.Official{}.TableName()+" o ON o.id = c.official_id AND o.delete_time = 0").
		Joins("LEFT JOIN "+db.User{}.TableName()+" ru ON ru.user_id = c.reply_user_id").
		Joins("LEFT JOIN "+db.Official{}.TableName()+" ro ON ro.id = c.reply_official_id").
		Joins("LEFT JOIN "+db.ArticleCommentLikeSQL{}.TableName()+" l ON l.comment_id = c.comment_id AND l.delete_time = 0 AND l.user_id = ? AND l.official_id = ?", userID, officialID).
		Order("comment_id asc").
		Offset(offset)
	if limit != 0 {
		tx = tx.Limit(limit)
	}
	if err = tx.Find(&comments).Error; err != nil {
		return nil, nil, err
	}

	var counts []commentsCount
	countTx := dbConn.Table(db.ArticleCommentSQL{}.TableName()+" c").
		Select("COUNT(c.parent_comment_id) as count", "c.parent_comment_id")
	if articleID != 0 {
		countTx = countTx.Where("c.article_id", articleID)
	}
	if parentCommentIDList != nil {
		countTx = countTx.Where("c.parent_comment_id", parentCommentIDList)
	}
	err = countTx.Where("c.delete_time", 0).Group("c.parent_comment_id").Find(&counts).Error
	if err != nil {
		return nil, nil, err
	}

	commentsMap := make(map[int64][]commentEntry, len(comments))
	for _, comment := range comments {
		commentsMap[comment.ParentCommentID] = append(commentsMap[comment.ParentCommentID], comment)
	}

	countsMap := make(map[int64]int64, len(counts))
	for _, count := range counts {
		countsMap[count.ParentCommentID] = count.Count
	}

	return commentsMap, countsMap, err
}

func GetOfficialIDListByInterestAndFollowList(followList, blockedOfficialIDList, interestList []int64) ([]int64, error) {
	if len(interestList) == 0 {
		return []int64{}, nil
	}

	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var resOfficialIDList []int64
	dbSub := dbConn.Table(db.OfficialInterest{}.TableName())
	sql := "interest_type_id IN (?)"
	params := []interface{}{interestList}
	if followList != nil {
		sql += " OR official_id IN (?)"
		params = append(params, followList)
	}
	dbSub = dbSub.Where(sql, params)
	if blockedOfficialIDList != nil {
		dbSub = dbSub.Where("official_id NOT IN (?)", blockedOfficialIDList)
	}
	err = dbSub.Pluck("official_id", &resOfficialIDList).Error

	if err != nil {
		return nil, err
	}
	return resOfficialIDList, nil
}

type GetTotalAnalyticsBetweenRecord struct {
	Gender      int32 `gorm:"gender"`
	Likes       int64 `gorm:"likes"`
	Comments    int64 `gorm:"comments"`
	Follows     int64 `gorm:"follows"`
	Reads       int64 `gorm:"reads"`
	UniqueReads int64 `gorm:"unique_reads"`
}

func GetTotalAnalyticsBetween(officialID, startTime, endTime int64) ([]GetTotalAnalyticsBetweenRecord, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var results []GetTotalAnalyticsBetweenRecord

	err = dbConn.Table(db.OfficialAnalytics{}.TableName()).
		Select(
			"gender",
			"SUM(like_counts) as likes",
			"SUM(comment_counts) as comments",
			"SUM(follow_counts) as follows",
			"SUM(read_counts) as `reads`",
			"SUM(unique_read_counts) as unique_reads",
		).
		Where("official_id", officialID).
		Where("time between ? and ?", startTime, endTime).
		Group("official_id, gender").
		Find(&results).
		Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

type GetDailyAnalyticsRecord struct {
	Day         int64 `gorm:"day"`
	Likes       int64 `gorm:"likes"`
	Comments    int64 `gorm:"comments"`
	Follows     int64 `gorm:"follows"`
	Reads       int64 `gorm:"reads"`
	UniqueReads int64 `gorm:"unique_reads"`
}

func GetDailyAnalytics(officialID, startTime, endTime int64) ([]GetDailyAnalyticsRecord, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var results []GetDailyAnalyticsRecord

	err = dbConn.Table(db.OfficialAnalytics{}.TableName()).
		Select(
			"(FLOOR(time / 86400) * 86400) as day",
			"SUM(like_counts) as likes",
			"SUM(comment_counts) as comments",
			"SUM(follow_counts) as follows",
			"SUM(read_counts) as `reads`",
			"SUM(unique_read_counts) as unique_reads",
		).
		Where("official_id", officialID).
		Where("time between ? and ?", startTime, endTime).
		Group("official_id, day").
		Find(&results).
		Error
	if err != nil {
		return nil, err
	}

	return results, nil
}

func SearchOfficialAccounts(userID, keyword string, offset, limit int) ([]OfficialFollow, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, 0, err
	}

	var count int64
	var officialAccounts []OfficialFollow
	searchValue := "%" + strings.Replace(keyword, "%", "\\%", -1) + "%"
	err = dbConn.Table(db.Official{}.TableName()+" o").
		Select(
			"o.id as official_id",
			"o.nickname",
			"o.face_url",
			"o.bio",
			"o.type",
			"f.follow_time",
			"f.muted",
			"f.enabled",
		).
		Where("o.delete_time", 0).
		Where("o.process_status", 1).
		Where(dbConn.Where("o.nickname LIKE ?", searchValue).Or("o.bio LIKE ?", searchValue)).
		Joins("LEFT JOIN "+db.OfficialFollowSQL{}.TableName()+" f ON f.official_id = o.id AND f.user_id = ? AND f.delete_time = 0", userID).
		Order("f.id desc").
		Offset(offset).
		Limit(limit).
		Count(&count).
		Find(&officialAccounts).Error

	if err != nil {
		return nil, 0, err
	}

	return officialAccounts, count, nil
}

func InsertArticleRead(articleID, officialID int64, userID string, userGender int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	var articleRead db.ArticleReadSQL
	err = dbConn.Table(articleRead.TableName()).Where("article_id", articleID).Where("user_id", userID).Take(&articleRead).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		newRead := db.ArticleReadSQL{
			ArticleID:  articleID,
			UserID:     userID,
			CreateTime: time.Now().Unix(),
			Status:     1,
		}

		if err = tx.Table(newRead.TableName()).Create(&newRead).Error; err != nil {
			return err
		}

		updatesMap := map[string]interface{}{"read_counts": gorm.Expr("read_counts + 1")}
		if articleRead.ArticleID == 0 {
			updatesMap["unique_read_counts"] = gorm.Expr("unique_read_counts + 1")
		}

		if err = tx.Table(db.ArticleSQL{}.TableName()).Where("article_id", articleID).Where("delete_time", 0).Updates(updatesMap).Error; err != nil {
			return err
		}

		officialAnalytics := db.OfficialAnalytics{
			OfficialID: officialID,
			Time:       utils.FloorTimeToHours(time.Now()),
			Gender:     userGender,
			ReadCounts: 1,
		}
		clausesUpdates := map[string]interface{}{"read_counts": gorm.Expr("read_counts + 1")}
		if articleRead.ArticleID == 0 {
			clausesUpdates["unique_read_counts"] = gorm.Expr("unique_read_counts + 1")
			officialAnalytics.UniqueReadCounts = 1
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "official_id"}, {Name: "time"}, {Name: "gender"}},
			DoUpdates: clause.Assignments(clausesUpdates),
		}).
			Table(db.OfficialAnalytics{}.TableName()).
			Create(&officialAnalytics).Error
	})
}

func ClearUserArticleReads(userID string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	return dbConn.Table(db.ArticleReadSQL{}.TableName()).Where("user_id", userID).Where("status", 1).Update("status", 2).Error
}

func DeleteArticleComment(comment db.ArticleCommentSQL) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		if err = tx.Table(comment.TableName()).Updates(&comment).Error; err != nil {
			return err
		}
		if comment.ParentCommentID != 0 {
			if err = tx.Table(comment.TableName()).Where("comment_id", comment.ParentCommentID).Update("reply_counts", gorm.Expr("reply_counts - 1")).Error; err != nil {
				return err
			}
		}
		err = tx.Table(db.ArticleSQL{}.TableName()).Where("article_id", comment.ArticleID).Update("comment_counts", gorm.Expr("comment_counts - 1")).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func GetLatestArticlesByOfficialID() ([]db.ArticleSQL, error) {
	var articleList = make([]db.ArticleSQL, 0)
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return articleList, err
	}
	err = dbConn.Raw("SELECT * FROM article WHERE article_id IN (SELECT DISTINCT MAX(article_id) from article where delete_time = 0 GROUP BY official_id)").Scan(&articleList).Error
	return articleList, err

}

func UpdateArticleCommentV2(comment *db.ArticleCommentSQL) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var rowsAffected int64 = 0
	if err != nil {
		return rowsAffected, err
	}
	var err1 error

	specifyMap := make(map[string]interface{})
	specifyMap["reply_counts"] = gorm.Expr("reply_counts+(?)", comment.ReplyCounts)
	specifyMap["like_counts"] = gorm.Expr("like_counts+(?)", comment.LikeCounts)

	tx := dbConn.Table(comment.TableName()).Debug().
		Where("comment_id=?", comment.CommentID).Updates(specifyMap)

	err1 = tx.Error
	rowsAffected = tx.RowsAffected
	if err1 != nil {
		return rowsAffected, err1
	}

	return rowsAffected, err1
}

func GetOfficialFollowerByWhere(where map[string]string, showNumber int32, pageNumber int32) ([]*db.OfficialFollowRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()

	var result []*db.OfficialFollowRes
	var count int64
	if err != nil {
		return result, count, err
	}
	var followerIdList []string
	var followerUsers []db.User
	var followerUserMap map[string]*db.User

	var followedIdList []int64
	var followedOfficial []db.Official
	var followedOfficialMap map[int64]*db.Official

	// query follower user
	queryUser := false
	{
		if user, ok := where["user"]; ok {
			if user != "" {
				queryUser = true
				followerUsers = GetUserByAllCondition(user)
			}
		}
		if queryUser {
			followerIdList = make([]string, len(followerUsers))
			followerUserMap = make(map[string]*db.User, len(followerUsers))
			for index, user := range followerUsers {
				followerIdList[index] = user.UserID
				followerUserMap[user.UserID] = &followerUsers[index]
			}
		}
	}

	// query official account
	queryOfficial := false
	{
		if officialAccount, ok := where["official_account"]; ok {
			if officialAccount != "" {
				queryOfficial = true
				dbConn.Table("official").Where("(nickname != '' AND nickname LIKE ?) OR (nickname = '' AND initial_nickname LIKE ?)",
					"%"+officialAccount+"%", "%"+officialAccount+"%").Find(&followedOfficial)
			}
		}
		if queryOfficial {
			followedIdList = make([]int64, len(followedOfficial))
			followedOfficialMap = make(map[int64]*db.Official)
			for index, official := range followedOfficial {
				followedIdList[index] = official.Id
				followedOfficialMap[official.Id] = &followedOfficial[index]
			}
		}
	}

	// query follow list
	dbSub := dbConn.Table(db.OfficialFollowSQL{}.TableName()).Where("delete_time = 0")
	if queryUser {
		dbSub = dbSub.Where("user_id", followerIdList)
	}
	if queryOfficial {
		dbSub = dbSub.Where("official_id", followedIdList)
	}
	if startTime, ok := where["start_time"]; ok {
		if startTime != "" {
			dbSub = dbSub.Where("follow_time >= ?", startTime)
		}
	}
	if endTime, ok := where["end_time"]; ok {
		if endTime != "" {
			dbSub = dbSub.Where("follow_time <= ?", endTime)
		}
	}
	if muted, ok := where["muted"]; ok {
		if muted != "" && muted != "0" {
			dbSub = dbSub.Where("muted", muted)
		}
	}

	var followDBRes []db.OfficialFollowSQL
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int((pageNumber - 1) * showNumber)).Find(&followDBRes)

	for _, follow := range followDBRes {
		if !queryUser {
			followerIdList = append(followerIdList, follow.UserID)
		}
		if !queryOfficial {
			followedIdList = append(followedIdList, follow.OfficialID)
		}
	}
	if !queryUser {
		followerUsers, _ = GetUsersByUserIDList(followerIdList)
		followerUserMap = make(map[string]*db.User, len(followerUsers))
		for index, user := range followerUsers {
			followerUserMap[user.UserID] = &followerUsers[index]
		}
	}
	if !queryOfficial {
		followedOfficial, _ = GetOfficialByIds(followedIdList)
		followedOfficialMap = make(map[int64]*db.Official)
		for index, official := range followedOfficial {
			followedOfficialMap[official.Id] = &followedOfficial[index]
		}
	}

	result = make([]*db.OfficialFollowRes, len(followDBRes))
	for index, follow := range followDBRes {
		followRes := db.OfficialFollowRes{}
		_ = utils.CopyStructFields(&followRes, follow)
		if user, ok := followerUserMap[follow.UserID]; ok {
			followRes.Username = user.Nickname
		}
		if official, ok := followedOfficialMap[follow.OfficialID]; ok {
			followRes.OfficialName = official.Nickname
			if followRes.OfficialName == "" {
				followRes.OfficialName = official.InitialNickname
			}
		}
		result[index] = &followRes
	}

	return result, count, nil
}

func BlockFollower(OpUserID, userID string, official int64, block int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	follow := db.OfficialFollow{
		UserID:     userID,
		OfficialID: official,
	}
	if block == 1 {
		follow.BlockedBy = ""
		follow.BlockTime = 0
	} else if block == 2 {
		follow.BlockedBy = OpUserID
		follow.BlockTime = time.Now().Unix()
	}

	return dbConn.Table(follow.TableName()).
		Where("user_id", userID).
		Where("official_id", official).Updates(&follow).Error
}

func MuteFollower(OpUserID, userID string, official int64, mute int32) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	specifyMap := make(map[string]interface{})

	follow := db.OfficialFollow{
		UserID:     userID,
		OfficialID: official,
	}
	if mute == 1 {
		specifyMap["muted"] = true
	} else if mute == 2 {
		specifyMap["muted"] = false
	}

	return dbConn.Table(follow.TableName()).
		Where("user_id", userID).
		Where("official_id", official).Updates(&specifyMap).Error
}
