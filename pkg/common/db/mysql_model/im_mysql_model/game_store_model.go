package im_mysql_model

import (
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func GetGamesByWhere(where map[string]interface{}, pageNumber, showNumber int32, orderBy string) ([]*db.Game, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var games []*db.Game
	var count int64 = 0
	if err != nil {
		return games, count, err
	}

	var orderBySQL = ""
	var categoriesIdList []int64

	dbSub := dbConn.Table(db.GameCategories{}.TableName()).Select("id").Where("delete_time=0")

	// query categories
	queryCategories := false
	if categoriesWhere, ok := where["categories"]; ok {
		categories := categoriesWhere.([]int64)
		if len(categories) > 0 {
			queryCategories = true
			dbSub = dbSub.Where("id IN (?)", categories)
		}
	}
	if queryCategories {
		dbSub = dbSub.Find(&categoriesIdList)
	}

	// query games
	dbSub = dbConn.Table(db.Game{}.TableName())

	if orderBy != "" {
		direction := "DESC"
		isFirst := true
		orderList := strings.Split(orderBy, ",")
		for _, order := range orderList {
			sort := strings.Split(order, ":")
			if len(sort) == 2 {
				if sort[1] == "ASC" {
					direction = "ASC"
				}
			}

			if isFirst {
				orderBySQL += sort[0] + " " + direction
				isFirst = false
			} else {
				orderBySQL += ", " + sort[0] + " " + direction
			}
		}
	}

	if gameCodeWhere, ok := where["game_code"]; ok {
		gameCode := gameCodeWhere.(string)
		if gameCode != "" {
			dbSub = dbSub.Where("game_code=?", gameCode)
		}
	}
	if gameNameWhere, ok := where["game_name"]; ok {
		gameName := gameNameWhere.(string)
		if gameName != "" {
			if nameTypeWhere, ok := where["name_type"]; ok {
				nameType := nameTypeWhere.(string)
				if nameType == "en" {
					dbSub = dbSub.Where("game_name_en LIKE ?", "%"+gameName+"%")
				} else if nameType == "cn" {
					dbSub = dbSub.Where("game_name_cn LIKE ?", "%"+gameName+"%")
				} else {
					dbSub = dbSub.Where("game_name_cn LIKE ? OR game_name_en LIKE ?", "%"+gameName+"%", "%"+gameName+"%")
				}
			} else {
				dbSub = dbSub.Where("game_name_cn LIKE ? OR game_name_en LIKE ?", "%"+gameName+"%", "%"+gameName+"%")
			}
		}
	}
	if publisherWhere, ok := where["publisher"]; ok {
		publisher := publisherWhere.(string)
		if publisher != "" {
			dbSub = dbSub.Where("publisher LIKE ?", "%"+publisher+"%")
		}
	}
	if hotWhere, ok := where["hot"]; ok {
		hot := hotWhere.(int32)
		if hot > 0 && hot <= 5 {
			dbSub = dbSub.Where("hot=?", hot)
		}
	}
	if stateWhere, ok := where["state"]; ok {
		state := stateWhere.(int)
		if state > 0 && state <= 2 {
			dbSub = dbSub.Where("state=?", state)
		}
	}
	if delWhere, ok := where["delete"]; ok {
		del := delWhere.(int)
		if del == 1 {
			dbSub = dbSub.Where("delete_time=0")
		}
	}
	if startTimeWhere, ok := where["start_time"]; ok {
		startTime := startTimeWhere.(string)
		if startTime != "" {
			dbSub = dbSub.Where("create_time>=?", startTime)
		}
	}
	if endTimeWhere, ok := where["end_time"]; ok {
		endTime := endTimeWhere.(string)
		if endTime != "" {
			dbSub = dbSub.Where("create_time<=?", endTime)
		}
	}
	if platformGameCodesWhere, ok := where["platform_game_codes"]; ok {
		platformGameCodes := platformGameCodesWhere.([]string)
		if len(platformGameCodes) > 0 {
			dbSub = dbSub.Where("game_code IN (?)", platformGameCodes)
		}
	}

	if queryCategories {
		categoriesSQL := ""
		var params []interface{}
		isFirst := true
		if len(categoriesIdList) == 0 {
			return games, 0, nil
		}
		for _, categoryId := range categoriesIdList {
			categoryStr := strconv.FormatInt(categoryId, 10)
			if isFirst {
				categoriesSQL += "categories LIKE ?"
				isFirst = false
			} else {
				categoriesSQL += " OR categories LIKE ?"
			}
			params = append(params, "%\""+categoryStr+"\"%")
		}
		dbSub = dbSub.Where(categoriesSQL, params...)
	}
	if classificationWhere, ok := where["classification"]; ok {
		classification := classificationWhere.(string)
		classifications := make([]string, 0)
		_ = json.Unmarshal([]byte(classification), &classifications)
		if classification != "" {
			isFirst := true
			classificationSQL := ""
			var params []interface{}
			for _, cla := range classifications {
				if isFirst {
					classificationSQL += "classification LIKE ?"
					isFirst = false
				} else {
					classificationSQL += " OR classification LIKE ?"
				}
				params = append(params, "%\""+cla+"\"%")
			}
			dbSub = dbSub.Where(classificationSQL, params...)
		}
	}

	if orderBySQL != "" {
		log.Debug("inter", "DB orderByClause", orderBySQL)
		dbSub = dbSub.Order(orderBySQL)
	}

	dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&games)

	return games, count, nil
}

func GetCategoriesByIdList(categories []string) ([]*db.GameCategories, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var categoriesRes []*db.GameCategories
	if err != nil {
		return categoriesRes, err
	}
	err = dbConn.Table("game_categories").Where("delete_time = 0").Where("id IN (?)", categories).Order("priority DESC, create_time DESC").Find(&categoriesRes).Error

	return categoriesRes, err
}

func GetCategoriesList() ([]*db.GameCategories, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var categories []*db.GameCategories
	if err != nil {
		return categories, err
	}
	err = dbConn.Table("game_categories").Where("delete_time = 0").Order("priority DESC, create_time DESC").Where("status=1").Find(&categories).Error

	return categories, err

}

func AddGame(game *db.Game) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	game.CreateTime = time.Now().Unix()

	result := dbConn.Table(game.TableName()).Create(game)
	return result.RowsAffected, err
}

func DeleteGamesByCodes(gamesCode []string, opUserId string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	game := &db.Game{
		DeletedBy:  opUserId,
		DeleteTime: time.Now().Unix(),
	}
	var result int64 = 0
	for _, gameCode := range gamesCode {
		game.GameCode = ""
		result += dbConn.Table(game.TableName()).Where("game_code=?", gameCode).Updates(game).RowsAffected
	}

	return result, nil
}

func UpdateGame(game *db.Game) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	game.UpdateTime = time.Now().Unix()

	result := dbConn.Table(game.TableName()).Where("game_code = ?", game.GameCode).Updates(game)
	return result.RowsAffected, err

}

func UpdateGameV2(game *db.Game) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var rowsAffected int64 = 0
	if err != nil {
		return rowsAffected, err
	}
	var err1 error

	specifyMap := make(map[string]interface{})
	specifyMap["click_counts"] = gorm.Expr("click_counts+(?)", game.ClickCounts)
	specifyMap["play_counts"] = gorm.Expr("play_counts+(?)", game.PlayCounts)

	tx := dbConn.Table("game").
		Where("game_code=?", game.GameCode).Updates(specifyMap)

	err1 = tx.Error
	rowsAffected = tx.RowsAffected
	if err1 != nil {
		return rowsAffected, err1
	}

	return rowsAffected, err1

}

func InsertGameHistory(gameHistory *db.GamePlayHistory) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}
	gameHistory.CreateTime = time.Now().Unix()
	return dbConn.Table(gameHistory.TableName()).Create(gameHistory).Error
}

func GetUserHistoriesGameList(UserID string, pageNumber, showNumber int32) ([]*db.Game, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var res []*db.Game
	if err != nil {
		return res, err
	}
	var historiesGame []db.GamePlayHistory
	var historiesCodes []string
	dbConn.Table("game_play_history").Limit(int(showNumber)).Offset(int(showNumber*(pageNumber-1))).
		Where("user_id = ?", UserID).Where("delete_time=0").Order("create_time DESC").Find(&historiesGame)
	for _, histories := range historiesGame {
		historiesCodes = append(historiesCodes, histories.GameCode)
	}

	var gameList []*db.Game
	var gameMap = make(map[string]*db.Game)
	dbConn.Table("game").Where("game_code IN (?)", historiesCodes).Find(&gameList)
	for index, game := range gameList {
		gameMap[game.GameCode] = gameList[index]
	}
	for _, g := range historiesGame {
		if gMap, ok := gameMap[g.GameCode]; ok {
			game := &db.Game{}
			_ = utils.CopyStructFields(game, gMap)
			game.CreateTime = g.CreateTime
			res = append(res, game)
		}
	}

	return res, err
}

func GetUserHistoryCounts(gameHistory *db.GamePlayHistory, UserID string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64

	err = dbConn.Table(gameHistory.TableName()).Where("user_id=?", UserID).Where("delete_time=0").Count(&count).Error
	return count, err
}

func GetHistories(UserID string) ([]*db.GamePlayHistory, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var histories []*db.GamePlayHistory
	if err != nil {
		return histories, err
	}
	err = dbConn.Table("game_play_history").Where("user_id = ?", UserID).Find(&histories).Error

	return histories, err
}

func GetUserFavoritesGameList(UserID string, pageNumber, showNumber int32) ([]*db.Game, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var res []*db.Game
	if err != nil {
		return res, err
	}
	var favoritesGame []db.GameFavorites
	var favoritesCodes []string
	dbConn.Table("game_favorites").Limit(int(showNumber)).Offset(int(showNumber*(pageNumber-1))).
		Where("user_id = ?", UserID).Where("delete_time=0").Order("create_time DESC").Find(&favoritesGame)
	for _, favorites := range favoritesGame {
		favoritesCodes = append(favoritesCodes, favorites.GameCode)
	}

	var gameList []*db.Game
	var gameMap = make(map[string]*db.Game)
	dbConn.Table("game").Where("game_code IN (?)", favoritesCodes).Find(&gameList)
	for index, game := range gameList {
		gameMap[game.GameCode] = gameList[index]
	}
	for _, g := range favoritesGame {
		if game, ok := gameMap[g.GameCode]; ok {
			game.CreateTime = g.CreateTime
			res = append(res, game)
		}
	}

	return res, err
}

func GetUserFavoritesCounts(gameFavorites *db.GameFavorites, UserID string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	var count int64

	err = dbConn.Table(gameFavorites.TableName()).Where("user_id=?", UserID).Where("delete_time=0").Count(&count).Error
	return count, err
}

func GetFavorites(userID string) ([]*db.GameFavorites, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var favorites []*db.GameFavorites
	if err != nil {
		return favorites, err
	}
	err = dbConn.Table("game_favorites").Where("user_id = ?", userID).Find(&favorites).Error

	return favorites, err
}

func HasFavored(userID, gameCode string) (bool, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var favorites = &db.GameFavorites{}
	if err != nil {
		return false, err
	}
	var count int64
	err = dbConn.Table(favorites.TableName()).Where("user_id=?", userID).Where("game_code=?", gameCode).Where("delete_time=0").Count(&count).Error

	return count != 0, err
}

func RemoveGameFavorite(favorite *db.GameFavorites) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	favorite.DeletedBy = favorite.UserID
	favorite.DeleteTime = time.Now().Unix()
	result := dbConn.Table(favorite.TableName()).Where("user_id=?", favorite.UserID).
		Where("game_code=?", favorite.GameCode).Updates(&favorite)

	return result.RowsAffected, result.Error
}

func AddGameFavorite(favorite *db.GameFavorites) error {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return err
	}

	favorite.CreateTime = time.Now().Unix()
	err = dbConn.Table(favorite.TableName()).Create(favorite).Error

	return err
}

func GetGameByGameCode(gameCode string) (*db.Game, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	game := &db.Game{}
	err = dbConn.Table(game.TableName()).Where("game_code=?", gameCode).Where("delete_time=0").Where("state=1").Find(game).Error
	return game, err
}

func GetGameAllByGameCode(gameCode string) (*db.Game, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	game := &db.Game{}
	err = dbConn.Table(game.TableName()).Where("game_code=?", gameCode).Where("delete_time=0").Find(game).Error
	return game, err
}

func GetGameLink(gameCode string) ([]*db.GameLink, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return nil, err
	}
	var gameLinks []*db.GameLink
	err = dbConn.Table("game_link").Where("game_code=?", gameCode).Where("delete_time=0").Find(&gameLinks).Error
	return gameLinks, err

}

func GetGameCodeListByPlatform(platform int32) ([]string, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var gameCodes []string
	if err != nil {
		return gameCodes, err
	}
	err = dbConn.Table("game_link").Select("game_code").Where("platform=?", platform).Where("delete_time=0").Find(&gameCodes).Error

	return gameCodes, err
}

func HasGameByGameName(gameName, gameCode string) (bool, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return false, err
	}
	var count int64
	err = dbConn.Table("game").
		Select("game_code").
		Where("game_code!=?", gameCode).
		Where("delete_time=0").
		Where("game_name_en=? OR game_name_cn=?", gameName, gameName).
		Count(&count).Error
	return count != 0, err
}

func RealDeleteGameLinksByGameCode(gameLink *db.GameLink, gameCode string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	result := dbConn.Table(gameLink.TableName()).Where("game_code=?", gameCode).Delete(gameLink)

	return result.RowsAffected, result.Error
}

func InsertGameLinks(gameLinks []*db.GameLink) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	result := dbConn.Table("game_link").Create(&gameLinks)

	return result.RowsAffected, result.Error
}

func InsertCategory(category *db.GameCategories) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	// check category name
	var oldCount int64
	var oldCategory db.GameCategories
	dbConn.Table(category.TableName()).Debug().Where("delete_time = 0 AND (category_name_en = ? OR category_name_cn=?)",
		category.CategoryNameCN, category.CategoryNameEN).Count(&oldCount).Find(&oldCategory)
	if oldCount != 0 && oldCategory.Id != category.Id {
		return 0, errors.New("category name already exists")
	}

	dbConn.Table(category.TableName()).Where("delete_time = 0 AND (category_name_en = ? OR category_name_cn=?)",
		category.CategoryNameEN, category.CategoryNameCN).Count(&oldCount).Find(&oldCategory)
	if oldCount != 0 && oldCategory.Id != category.Id {
		return 0, errors.New("category name already exists")
	}

	var count int64
	category.CreateTime = time.Now().Unix()
	dbConn.Table(category.TableName()).Where("category_name_en = ? OR category_name_cn=?",
		category.CategoryNameCN, category.CategoryNameEN).Where("delete_time=0").Count(&count)
	if count != 0 {
		return 0, nil
	}

	result := dbConn.Table(category.TableName()).Create(category)

	return result.RowsAffected, result.Error
}

func UpdateCategory(category *db.GameCategories) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	var oldCategory db.GameCategories
	dbConn.Table(category.TableName()).Debug().Where("delete_time = 0 AND (category_name_en = ? OR category_name_cn=?)",
		category.CategoryNameCN, category.CategoryNameEN).Count(&count).Find(&oldCategory)
	if count != 0 && oldCategory.Id != category.Id {
		return 0, nil
	}

	dbConn.Table(category.TableName()).Where("delete_time = 0 AND (category_name_en = ? OR category_name_cn=?)",
		category.CategoryNameEN, category.CategoryNameCN).Count(&count).Find(&oldCategory)
	if count != 0 && oldCategory.Id != category.Id {
		return 0, nil
	}

	category.UpdateTime = time.Now().Unix()
	result := dbConn.Table(category.TableName()).Where("id", category.Id).Updates(category)

	return result.RowsAffected, result.Error
}

func UpdateCategoryStatus(category *db.GameCategories) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	category.UpdateTime = time.Now().Unix()
	result := dbConn.Table(category.TableName()).Where("id=?", category.Id).Updates(category)

	return result.RowsAffected, result.Error
}

func DeleteCategoryById(categoryId int64, opUserId string) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}
	category := &db.GameCategories{
		DeletedBy:  opUserId,
		DeleteTime: time.Now().Unix(),
	}

	result := dbConn.Table(category.TableName()).Where("id=?", categoryId).Updates(category)

	return result.RowsAffected, result.Error
}

func GetCategoriesByWhere(where map[string]interface{}, pageNumber, showNumber int32, orderBy string) ([]*db.GameCategories, int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	var gamesCategories []*db.GameCategories
	var count int64
	if err != nil {
		return gamesCategories, count, err
	}

	var orderBySQL = ""
	dbSub := dbConn.Table(db.GameCategories{}.TableName()).Where("delete_time=0")

	if orderBy != "" {
		direction := "DESC"
		isFirst := true
		orderList := strings.Split(orderBy, ",")
		for _, order := range orderList {
			sort := strings.Split(order, ":")
			if len(sort) == 2 {
				if sort[1] == "ASC" {
					direction = "ASC"
				}
			}

			if isFirst {
				orderBySQL += sort[0] + " " + direction
				isFirst = false
			} else {
				orderBySQL += ", " + sort[0] + " " + direction
			}
		}
	}

	if categoryNameWhere, ok := where["category_name"]; ok {
		categoryName := categoryNameWhere.(string)
		if categoryName != "" {

			if nameTypeWhere, ok := where["category_name_type"]; ok {
				nameType := nameTypeWhere.(string)
				if nameType == "en" {
					dbSub = dbSub.Where("category_name_en LIKE ?", "%"+categoryName+"%")
				} else if nameType == "cn" {
					dbSub = dbSub.Where("category_name_cn LIKE ?", "%"+categoryName+"%")
				} else {
					dbSub = dbSub.Where("category_name_cn LIKE ? OR category_name_en LIKE ?", "%"+categoryName+"%", "%"+categoryName+"%")
				}
			} else {
				dbSub = dbSub.Where("category_name_cn LIKE ? OR category_name_en LIKE ?", "%"+categoryName+"%", "%"+categoryName+"%")
			}
		}
	}

	//if creatorWhere, ok := where["creator"]; ok {
	//	creator := creatorWhere.([]string)
	//	dbSub = dbSub.Where("create_user IN (?)", creator)
	//}
	//if editorWhere, ok := where["editor"]; ok {
	//	editor := editorWhere.([]string)
	//	dbSub = dbSub.Where("update_user IN (?)", editor)
	//}

	if creatorWhere, ok := where["creator"]; ok {
		creator := creatorWhere.(string)
		dbSub = dbSub.Where("create_user LIKE ?", "%" + creator + "%")
	}
	if editorWhere, ok := where["editor"]; ok {
		editor := editorWhere.(string)
		dbSub = dbSub.Where("update_user LIKE ?", "%" + editor + "%")
	}

	if createStartTimeWhere, ok := where["create_start_time"]; ok {
		createStartTime := createStartTimeWhere.(string)
		if createStartTime != "" {
			dbSub = dbSub.Where("create_time >= ?", createStartTime)
		}
	}
	if createEndTimeWhere, ok := where["create_end_time"]; ok {
		createEndTime := createEndTimeWhere.(string)
		if createEndTime != "" {
			dbSub = dbSub.Where("create_time <= ?", createEndTime)
		}
	}
	if editStartTimeWhere, ok := where["edited_start_time"]; ok {
		editStartTime := editStartTimeWhere.(string)
		if editStartTime != "" {
			dbSub = dbSub.Where("update_time >= ?", editStartTime)
		}
	}
	if editEndTimeWhere, ok := where["edited_end_time"]; ok {
		editEndTime := editEndTimeWhere.(string)
		if editEndTime != "" {
			dbSub = dbSub.Where("update_time <= ?", editEndTime)
		}
	}

	if stateWhere, ok := where["state"]; ok {
		status := stateWhere.(int32)
		if status > 0 && status <= 2 {
			dbSub = dbSub.Where("status = ?", status)
		}
	}

	if orderBySQL != "" {
		log.Debug("inter", "DB orderByClause", orderBySQL)
		dbSub = dbSub.Order(orderBySQL)
	}

	dbSub.Count(&count).Limit(int(showNumber)).Offset(int(showNumber * (pageNumber - 1))).Find(&gamesCategories)

	return gamesCategories, count, nil
}

func GetCategoryUsedAmount(categoryId int64) (int64, error) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return 0, err
	}

	var count int64
	category := strconv.FormatInt(categoryId, 10)
	err = dbConn.Table("game").Where("categories LIKE ?", "%\""+category+"\"%").Where("delete_time=0").Count(&count).Error

	return count, err
}
