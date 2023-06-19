package im_mysql_model

import (
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	pbAdminCMS "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func mergeInterestTypeToResMap(interestsType []*db.InterestType, interestsLanguage []*db.InterestLanguage) map[int64]*db.InterestTypeRes {
	interestsTypeResMap := make(map[int64]*db.InterestTypeRes)
	if len(interestsType) == 0 {
		return interestsTypeResMap
	}

	if len(interestsType) > 0 && len(interestsLanguage) > 0 {
		for _, interestType := range interestsType {
			interestTypeRes := db.InterestTypeRes{}
			utils.CopyStructFields(&interestTypeRes, &interestType)
			interestTypeRes.Name = []db.InterestLanguage{}
			interestsTypeResMap[interestType.Id] = &interestTypeRes
		}
		for _, language := range interestsLanguage {
			if interest, ok := interestsTypeResMap[language.InterestId]; ok {
				interest.Name = append(interest.Name, *language)
			}
		}
	}

	return interestsTypeResMap

}

func GetAllInterestType() []*db.InterestTypeRes {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	var interestTypeRes []*db.InterestTypeRes
	var interestLanguage []*db.InterestLanguage
	var interestType []*db.InterestType

	dbConn.Model(&db.InterestLanguage{}).Find(&interestLanguage)
	dbConn.Model(&db.InterestType{}).Where("status=1").Where("delete_time=0").Find(&interestType)

	interestsTypeResMap := mergeInterestTypeToResMap(interestType, interestLanguage)
	for _, res := range interestsTypeResMap {
		interestTypeRes = append(interestTypeRes, res)
	}

	return interestTypeRes
}

func GetNotDefaultInterestIdList() []int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	var list []int64
	dbConn.Table("interest_type").Where("is_default!=?", constant.InterestIsDefaultType).Debug().Pluck("id", &list)
	return list
}

func GetDefaultInterestTypeList() []int64 {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	var list []int64
	dbConn.Table("interest_type").Where("is_default = ?", constant.InterestIsDefaultType).Pluck("id", &list)
	return list
}

func AddInterestType(name, user string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	interestType := db.InterestType{
		UpdateUser: user,
		UpdateTime: time.Now().Unix(),
	}

	insertErr := dbConn.Table(db.InterestType{}.TableName()).Create(&interestType).Error
	if insertErr != nil {
		return false
	}
	return true
}

func UpdateInterestType(id int64, name, user string) bool {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false
	}

	updateErr := dbConn.Table(db.InterestType{}.TableName()).Where("id=?", id).Update("name", name).Update("update_user", user).Update("update_time", time.Now().Unix()).Error
	if updateErr != nil {
		return false
	}
	return true
}

func GetInterestsByUserIds(userIdList []string) map[string][]*db.InterestTypeRes {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}

	if userIdList == nil || len(userIdList) == 0 {
		return map[string][]*db.InterestTypeRes{}
	}
	var noRepeatedIdList []string
	var filter = make(map[string]struct{})
	for _, userid := range userIdList {
		filter[userid] = struct{}{}
	}
	noRepeatedIdList = make([]string, len(filter))
	var filterIndex = 0
	for userid, _ := range filter {
		noRepeatedIdList[filterIndex] = userid
		filterIndex++
	}

	result := make(map[string][]*db.InterestTypeRes)

	// query interest user, get interests list, and map
	var interests []int64
	var interestUsers []db.InterestUser
	var interestMap map[string][]*db.InterestUser

	dbConn.Table("interest_user").Debug().Where("user_id IN (?)", noRepeatedIdList).Find(&interestUsers)

	interestMap = make(map[string][]*db.InterestUser)
	for index, interestUser := range interestUsers {
		interests = append(interests, interestUser.InterestId)
		if _, ok := interestMap[interestUser.UserId]; !ok {
			interestMap[interestUser.UserId] = make([]*db.InterestUser, 0)
		}
		interestMap[interestUser.UserId] = append(interestMap[interestUser.UserId], &interestUsers[index])
	}

	interestTypeMap := GetInterestsByIds(interests)
	defaultInterestIds := GetDefaultInterestTypeList()
	defaultInterestTemp, err := GetInterestsById(defaultInterestIds)
	if err != nil {
		return nil
	}
	var defaultInterest = make([]*db.InterestTypeRes, len(defaultInterestTemp))
	for index, _ := range defaultInterestTemp {
		defaultInterest[index] = &defaultInterestTemp[index]
	}

	// Loop interest user, set user's interesting.
	for _, userId := range noRepeatedIdList {
		if interestUser, ok := interestMap[userId]; ok {
			for _, user := range interestUser {
				if interest, ok := interestTypeMap[user.InterestId]; ok {
					var interestType []*db.InterestTypeRes
					_ = utils.CopyStructFields(&interestType, &interest)
					if _, ok = result[userId]; !ok {
						result[userId] = make([]*db.InterestTypeRes, 0)
					}
					result[userId] = append(result[userId], interestType...)
					log.Debug("find", user.InterestId)
				}
			}
		} else {
			fmt.Println("didn't find in interest type map, user id: ", userId)
			result[userId] = defaultInterest
		}
	}

	return result
}

func GetInterestsByIds(interests []int64) map[int64][]*db.InterestTypeRes {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil
	}
	if interests == nil || len(interests) == 0 {
		return nil
	}

	interestMap := make(map[int64][]*db.InterestTypeRes)

	//var defaultInterestTypes []db.InterestType
	//var defaultInterestMapForLink = make(map[int64]*db.InterestTypeRes)

	//for _, interest := range interests {
	//	if interest == -1 {
	//		interestMap[-1] = make([]*db.InterestTypeRes, 0)
	//
	//		defaultInterestLanguage := make([]db.InterestLanguage, 0)
	//		// default
	//		dbConn.Table("interest_type").Where("is_default=2").Find(&defaultInterestTypes)
	//		defaultInterestIdList := make([]int64, len(defaultInterestTypes))
	//
	//		for index, interestType := range defaultInterestTypes {
	//			interestTypeRes := db.InterestTypeRes{Name: []db.InterestLanguage{}}
	//			utils.CopyStructFields(&interestTypeRes, &interestType)
	//			defaultInterestMapForLink[interestType.Id] = &interestTypeRes
	//			defaultInterestIdList[index] = interestType.Id
	//			interestMap[-1] = append(interestMap[-1], &interestTypeRes)
	//		}
	//		dbConn.Table("interest_language").Where("interest_id IN (?)", defaultInterestIdList).Find(&defaultInterestLanguage)
	//		for _, language := range defaultInterestLanguage {
	//			if interestType, ok := defaultInterestMapForLink[language.InterestId]; ok {
	//				interestType.Name = append(interestType.Name, language)
	//			}
	//		}
	//		break
	//	}
	//}

	var interestTypes []db.InterestType
	var interestTypeMap = make(map[int64]*db.InterestTypeRes)
	var interestLanguage []db.InterestLanguage
	// Query personal interest type.
	dbConn.Table("interest_type").Debug().Where("id IN (?)", interests).Where("delete_time=0").Find(&interestTypes)
	for _, interestType := range interestTypes {
		interestTypeRes := db.InterestTypeRes{Name: []db.InterestLanguage{}}
		_ = utils.CopyStructFields(&interestTypeRes, &interestType)
		interestTypeMap[interestType.Id] = &interestTypeRes
		if _, ok := interestMap[interestType.Id]; !ok {
			interestMap[interestType.Id] = make([]*db.InterestTypeRes, 0)
		}

		interestMap[interestType.Id] = append(interestMap[interestType.Id], &interestTypeRes)
		log.Debug("interest map insert: ", interestType.Id)
	}

	// Query interest language.
	dbConn.Table("interest_language").Debug().Where("interest_id IN (?)", interests).Find(&interestLanguage)
	for _, language := range interestLanguage {
		if interestType, ok := interestTypeMap[language.InterestId]; ok {
			interestType.Name = append(interestType.Name, language)
		}
	}

	log.Debug("GetInterestsByIds The map:")

	return interestMap
}

func GetInterestsByWhere(where map[string]string, showNumber int32, pageNumber int32, orderBy string) ([]*db.InterestTypeRes, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var interestsRes []*db.InterestTypeRes
	var count int64
	if err != nil {
		return interestsRes, count, err
	}
	sortMap := map[string]string{}
	timeTypeMap := map[int]string{}
	var orderByClause string
	sortMap["start_time"] = "create_time"
	timeTypeMap[1] = "create_time"
	timeTypeMap[2] = "update_time"

	var nameIdList []int64
	var nameList []db.InterestLanguage
	var name = ""
	ok := false
	if name, ok = where["name"]; !ok {
		name = ""
	}

	// query interest_language
	dbSub := dbConn.Table("interest_language")

	if name != "" {
		dbSub = dbConn.Where("name like ?", "%"+name+"%")
	}
	dbSub.Find(&nameList)

	// Need to query twice.
	if name != "" {
		for _, language := range nameList {
			nameIdList = append(nameIdList, language.InterestId)
		}
		dbConn.Table("interest_language").Where("interest_id IN (?)", nameIdList).Find(&nameList)
	}

	dbSub = dbConn.Table("interest_type").Where("delete_time=0")
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
	if name != "" {
		dbSub = dbSub.Where("id IN (?)", nameIdList)
	}
	if createUser, ok := where["create_user"]; ok {
		log.Debug("inter", "DB create_user", createUser)
		if createUser != "" {
			dbSub = dbSub.Where("create_user like ?", "%"+createUser+"%")
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
	if remark, ok := where["remark"]; ok {
		log.Debug("inter", "DB remark", remark)
		if remark != "" {
			dbSub = dbSub.Where("interest_type.remark like ?", "%"+remark+"%")
		}
	}
	if status, ok := where["status"]; ok {
		log.Debug("inter", "DB status", status)
		if status != "" {
			dbSub = dbSub.Where("status=?", status)
		}
	}
	if isDefault, ok := where["is_default"]; ok {
		if isDefault != "" && isDefault != "0" {
			dbSub = dbSub.Where("is_default=?", isDefault)
		}
	}

	if orderByClause != "" {
		log.Debug("inter", "DB orderByClause", orderByClause)
		dbSub = dbSub.Order(orderByClause)
	}
	var interests []db.InterestType
	dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Debug().Find(&interests)
	log.Debug("", "get interests: ", interests, " len: ", len(interests))
	if len(nameList) > 0 && len(interests) > 0 {
		interestsMap := map[int64]*db.InterestTypeRes{}
		for _, interest := range interests {
			interestTypeRes := db.InterestTypeRes{}
			utils.CopyStructFields(&interestTypeRes, &interest)
			interestTypeRes.Name = []db.InterestLanguage{}
			interestsMap[interest.Id] = &interestTypeRes
			interestsRes = append(interestsRes, &interestTypeRes)
		}
		log.Debug("", "name list: ", nameList)
		for _, language := range nameList {
			if interest, ok := interestsMap[language.InterestId]; ok {
				interest.Name = append(interest.Name, language)
			}
		}
	}
	log.Debug("", "interestsRes: ", interestsRes, " len: ", len(interestsRes))
	return interestsRes, count, err
}

func GetInterestsById(interestIdList []int64) ([]db.InterestTypeRes, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}

	var interests []db.InterestType
	if err = dbConn.Table(db.InterestType{}.TableName()).Where("delete_time = 0").Where("status = 1").Where("id IN (?)", interestIdList).Order("id asc").Find(&interests).Error; err != nil {
		return nil, err
	}

	var interestsLanguage []db.InterestLanguage
	if err = dbConn.Table(db.InterestLanguage{}.TableName()).Where("interest_id IN (?)", interestIdList).Find(&interestsLanguage).Error; err != nil {
		return nil, err
	}

	interestsRes := make([]db.InterestTypeRes, len(interests))

	for i, interest := range interests {
		name := make([]db.InterestLanguage, 0)
		for _, language := range interestsLanguage {
			if language.InterestId == interest.Id {
				name = append(name, language)
			}
		}
		utils.CopyStructFields(&interestsRes[i], interest)
		interestsRes[i].Name = name
	}

	return interestsRes, nil
}

func DeleteInterests(interestIdList []string, opUserId string) (i int64, err error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	interest := db.InterestType{
		DeleteTime: time.Now().Unix(),
		DeletedBy:  opUserId,
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		var allDefaultCount int64
		tx.Table(interest.TableName()).Where("is_default=2").Count(&allDefaultCount)
		var deleteDefaultCount int64
		tx.Table(interest.TableName()).Where("id IN (?)", interestIdList).Where("is_default=2").Count(&deleteDefaultCount)
		if allDefaultCount-deleteDefaultCount == 0 {
			return errors.New("after delete the default interest count will be 0")
		}
		var userUsed int64
		tx.Table(db.InterestUser{}.TableName()).Where("interest_id IN (?)", interestIdList).Count(&userUsed)
		if userUsed != 0 {
			return errors.New("there are users using the interest")
		}
		var groupUsed int64
		tx.Table(db.InterestGroup{}.TableName()).Where("interest_id IN (?)", interestIdList).Count(&groupUsed)
		if groupUsed != 0 {
			return errors.New("there are groups using the interest")
		}

		res := dbConn.Table("interest_type").Where("id IN (?)", interestIdList).Updates(&interest)
		i += res.RowsAffected
		return res.Error
	})
	return i, err
}

func AlterInterest(interest *db.InterestType, names []db.InterestLanguage, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	if names != nil {
		for index := range names {
			names[index].InterestId = interest.Id
		}
	}
	log.Debug("", "names len: ", len(names))

	interest.UpdateTime = time.Now().Unix()
	interest.UpdateUser = opUserId

	err = dbConn.Transaction(func(tx *gorm.DB) error {
		tx.Table("interest_language").Where("interest_id = ?", interest.Id).Delete(&db.InterestLanguage{})
		tx.Table("interest_language").Create(&names)
		i = tx.Table("interest_type").Updates(&interest).RowsAffected
		return nil
	})
	if err != nil {
		return 0
	}
	return i
}

func ChangeInterestStatus(interestId int64, status int32, opUserId string) (i int64) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0
	}
	interest := db.InterestType{
		Id:         interestId,
		Status:     int8(status),
		UpdateTime: time.Now().Unix(),
		UpdateUser: opUserId,
	}
	i = dbConn.Table("interest_type").Updates(&interest).RowsAffected
	return i
}

func AddInterests(interests []*pbAdminCMS.InterestReq, opUserId string) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	err = dbConn.Transaction(func(tx *gorm.DB) error {
		nowTime := time.Now().Unix()
		var languageList []db.InterestLanguage
		for _, interest := range interests {
			interestType := db.InterestType{
				Status:     int8(interest.Status),
				Remark:     interest.Remark,
				CreateUser: opUserId,
				CreateTime: nowTime,
				UpdateUser: interest.UpdateUser,
				UpdateTime: interest.UpdateTime,
				DeleteTime: 0,
				IsDefault:  int8(interest.IsDefault),
			}
			dbConn.Create(&interestType)
			nameList := interest.Name
			for _, name := range nameList {
				language := db.InterestLanguage{
					InterestId:   interestType.Id,
					LanguageType: name.LanguageType,
					Name:         name.Name,
				}
				languageList = append(languageList, language)
			}
		}
		dbConn.Create(&languageList)
		return nil
	})
	return err
}
