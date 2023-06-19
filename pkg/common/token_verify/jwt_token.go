package token_verify

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	commonDB "Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"time"

	go_redis "github.com/go-redis/redis/v9"
	"github.com/golang-jwt/jwt/v4"
)

//var (
//	TokenExpired     = errors.New("token is timed out, please log in again")
//	TokenInvalid     = errors.New("token has been invalidated")
//	TokenNotValidYet = errors.New("token not active yet")
//	TokenMalformed   = errors.New("that's not even a token")
//	TokenUnknown     = errors.New("couldn't handle this token")
//)

type Claims struct {
	UID        string
	Platform   string //login platform
	OfficialID int64
	jwt.RegisteredClaims
}

func BuildClaims(uid, platform string, officialID, ttl int64) Claims {
	now := time.Now()
	return Claims{
		UID:        uid,
		Platform:   platform,
		OfficialID: officialID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttl*24) * time.Hour)), //Expiration time
			IssuedAt:  jwt.NewNumericDate(now),                                        //Issuing time
			NotBefore: jwt.NewNumericDate(now),                                        //Begin Effective time
		}}
}

//	func DeleteToken(userID string, platformID int) error {
//		m, err := commonDB.DB.GetTokenMapByUidPid(userID, constant.PlatformIDToName(platformID))
//		if err != nil && err != go_redis.Nil {
//			return utils.Wrap(err, "")
//		}
//		var deleteTokenKey []string
//		for k, v := range m {
//			_, err = GetClaimFromToken(k)
//			if err != nil || v != constant.NormalToken {
//				deleteTokenKey = append(deleteTokenKey, k)
//			}
//		}
//		if len(deleteTokenKey) != 0 {
//			err = commonDB.DB.DeleteTokenByUidPid(userID, platformID, deleteTokenKey)
//			return utils.Wrap(err, "")
//		}
//		return nil
//	}
func DeleteToken(userID string, platformID int, gAuthTypeToken bool) error {
	m, err := commonDB.DB.GetTokenMapByUidPid(userID, constant.PlatformIDToName(platformID))
	if err != nil && err != go_redis.Nil {
		return utils.Wrap(err, "")
	}
	var deleteTokenKey []string
	for k, v := range m {
		_, err = GetClaimFromToken(k, gAuthTypeToken)
		if err != nil || v != constant.NormalToken {
			deleteTokenKey = append(deleteTokenKey, k)

		}
	}
	if len(deleteTokenKey) != 0 {
		err = commonDB.DB.DeleteTokenByUidPid(userID, platformID, deleteTokenKey)
		return utils.Wrap(err, "")
	}
	return nil
}

// }
func DeleteAllToken(userID string) error {
	for platformID := 1; platformID < len(constant.PlatformID2Name); platformID++ {

		m, err := commonDB.DB.GetTokenMapByUidPid(userID, constant.PlatformIDToName(platformID))
		if err != nil && err != go_redis.Nil {
			return utils.Wrap(err, "")
		}
		var deleteTokenKey []string
		for k, _ := range m {
			deleteTokenKey = append(deleteTokenKey, k)
		}
		if len(deleteTokenKey) != 0 {
			err = commonDB.DB.DeleteTokenByUidPid(userID, platformID, deleteTokenKey)
			//return utils.Wrap(err, "")
		}
	}
	return nil
}

// func CreateToken(userID string, platformID int) (string, int64, error) {
// 	claims := BuildClaims(userID, constant.PlatformIDToName(platformID), config.Config.TokenPolicy.AccessExpire)
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString([]byte(config.Config.TokenPolicy.AccessSecret))
// 	if err != nil {
// 		return "", 0, err
// 	}
// 	//remove Invalid token
// 	m, err := commonDB.DB.GetTokenMapByUidPid(userID, constant.PlatformIDToName(platformID))
// 	if err != nil && err != go_redis.Nil {
// 		return "", 0, err
// 	}
// 	var deleteTokenKey []string
// 	for k, v := range m {
// 		_, err = GetClaimFromToken(k)
// 		if err != nil || v != constant.NormalToken {
// 			deleteTokenKey = append(deleteTokenKey, k)
// 		}
// 	}
// 	if len(deleteTokenKey) != 0 {
// 		err = commonDB.DB.DeleteTokenByUidPid(userID, platformID, deleteTokenKey)
// 		if err != nil {
// 			return "", 0, err
// 		}
// 	}
// 	err = commonDB.DB.AddTokenFlag(userID, platformID, tokenString, constant.NormalToken)
// 	if err != nil {
// 		return "", 0, err
// 	}
// 	return tokenString, claims.ExpiresAt.Time.Unix(), err
// }

func CreateToken(userID string, platformID int, gAuthTypeToken bool, officialID int64) (string, int64, error) {
	claims := BuildClaims(userID, constant.PlatformIDToName(platformID), officialID, config.Config.TokenPolicy.AccessExpire)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var tokenString string
	var err error
	if gAuthTypeToken {
		tokenString, err = token.SignedString([]byte(config.Config.TokenPolicy.AccessSecretGAuth))
	} else {
		tokenString, err = token.SignedString([]byte(config.Config.TokenPolicy.AccessSecret))
	}
	if err != nil {
		return "", 0, err
	}
	//remove Invalid token
	m, err := commonDB.DB.GetTokenMapByUidPid(userID, constant.PlatformIDToName(platformID))
	if err != nil && err != go_redis.Nil {
		return "", 0, err
	}
	var deleteTokenKey []string
	for k, v := range m {
		_, err = GetClaimFromToken(k, gAuthTypeToken)
		if err != nil || v != constant.NormalToken {
			deleteTokenKey = append(deleteTokenKey, k)
		}
	}
	if len(deleteTokenKey) != 0 {
		err = commonDB.DB.DeleteTokenByUidPid(userID, platformID, deleteTokenKey)
		if err != nil {
			return "", 0, err
		}
	}
	err = commonDB.DB.AddTokenFlag(userID, platformID, tokenString, constant.NormalToken)
	if err != nil {
		return "", 0, err
	}
	return tokenString, claims.ExpiresAt.Time.Unix(), err
}

//	func secret() jwt.Keyfunc {
//		return func(token *jwt.Token) (interface{}, error) {
//			return []byte(config.Config.TokenPolicy.AccessSecret), nil
//		}
//	}
func secret(gAuthTypeToken bool) jwt.Keyfunc {
	if gAuthTypeToken {
		return func(token *jwt.Token) (interface{}, error) {
			return []byte(config.Config.TokenPolicy.AccessSecretGAuth), nil
		}
	}
	return func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.TokenPolicy.AccessSecret), nil
	}
}

//	func GetClaimFromToken(tokensString string) (*Claims, error) {
//		token, err := jwt.ParseWithClaims(tokensString, &Claims{}, secret())
//		if err != nil {
//			if ve, ok := err.(*jwt.ValidationError); ok {
//				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
//					return nil, &constant.ErrTokenMalformed
//				} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
//					return nil, &constant.ErrTokenExpired
//				} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
//					return nil, &constant.ErrTokenNotValidYet
//				} else {
//					return nil, &constant.ErrTokenUnknown
//				}
//			} else {
//				return nil, &constant.ErrTokenNotValidYet
//			}
//		} else {
//			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
//				//log.NewDebug("", claims.UID, claims.Platform)
//				return claims, nil
//			}
//			return nil, &constant.ErrTokenNotValidYet
//		}
//	}
func GetClaimFromToken(tokensString string, gAuthTypeToken bool) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokensString, &Claims{}, secret(gAuthTypeToken))
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, &constant.ErrTokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, &constant.ErrTokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, &constant.ErrTokenNotValidYet
			} else {
				return nil, &constant.ErrTokenUnknown
			}
		} else {
			return nil, &constant.ErrTokenNotValidYet
		}
	} else {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			//log.NewDebug("", claims.UID, claims.Platform)
			return claims, nil
		}
		return nil, &constant.ErrTokenNotValidYet
	}
}

func IsAppManagerAccess(token string, OpUserID string) bool {
	gAuthTypeToken := false
	claims, err := ParseToken(token, "", gAuthTypeToken)
	if err != nil {
		return false
	}
	if utils.IsContain(claims.UID, config.Config.Manager.AppManagerUid) && claims.UID == OpUserID {
		return true
	}
	return false
}

func IsManagerUserID(OpUserID string) bool {
	if utils.IsContain(OpUserID, config.Config.Manager.AppManagerUid) {
		return true
	} else {
		return false
	}
}

// func CheckAccess(OpUserID string, OwnerUserID string) bool {
// 	if utils.IsContain(OpUserID, config.Config.Manager.AppManagerUid) {
// 		return true
// 	}
// 	if OpUserID == OwnerUserID {
// 		return true
// 	}
// 	return false
// }

//	func GetUserIDFromToken(token string, operationID string) (bool, string, string) {
//		claims, err := ParseToken(token, operationID)
//		if err != nil {
//			log.Error(operationID, "ParseToken failed, ", err.Error(), token)
//			return false, "", err.Error()
//		}
//		log.Debug(operationID, "token claims.ExpiresAt.Second() ", claims.ExpiresAt.Unix())
//		return true, claims.UID, ""
//	}
func GetUserIDFromToken(token string, operationID string) (bool, string, string) {
	gAuthTypeToken := false
	claims, err := ParseToken(token, operationID, gAuthTypeToken)
	if err != nil {
		log.Error(operationID, "ParseToken failed, ", err.Error(), token)
		return false, "", err.Error()
	}
	log.Debug(operationID, "token claims.ExpiresAt.Second() ", claims.ExpiresAt.Unix())
	return true, claims.UID, ""
}
func GetUserIDFromTokenV2(token string, operationID string) (bool, string, int64, string) {
	gAuthTypeToken := false
	claims, err := ParseToken(token, operationID, gAuthTypeToken)
	if err != nil {
		log.Error(operationID, "ParseToken failed, ", err.Error(), token)
		return false, "", 0, err.Error()
	}
	log.Debug(operationID, "token claims.ExpiresAt.Second() ", claims.ExpiresAt.Unix())
	return true, claims.UID, claims.OfficialID, ""
}

func GetAdminUserIDFromToken(token string, operationID string, gAuthTypeToken bool) (bool, string, string) {
	claims, err := ParseToken(token, operationID, gAuthTypeToken)
	if err != nil {
		log.Error(operationID, "ParseToken failed, ", err.Error(), token)
		return false, "", err.Error()
	}
	log.Debug(operationID, "token claims.ExpiresAt.Second() ", claims.ExpiresAt.Unix())
	return true, claims.UID, ""
}

func GetUserIDFromTokenExpireTime(token string, operationID string) (bool, string, string, int64) {
	gAuthTypeToken := false
	claims, err := ParseToken(token, operationID, gAuthTypeToken)
	if err != nil {
		log.Error(operationID, "ParseToken failed, ", err.Error(), token)
		return false, "", err.Error(), 0
	}
	return true, claims.UID, "", claims.ExpiresAt.Unix()
}

func ParseTokenGetUserID(token string, operationID string) (error, string) {
	gAuthTypeToken := false
	claims, err := ParseToken(token, operationID, gAuthTypeToken)
	if err != nil {
		return utils.Wrap(err, ""), ""
	}
	return nil, claims.UID
}

// func ParseToken(tokensString, operationID string) (claims *Claims, err error) {
// 	claims, err = GetClaimFromToken(tokensString)
// 	if err != nil {
// 		log.NewError(operationID, "token validate err", err.Error(), tokensString)
// 		return nil, utils.Wrap(err, "")
// 	}

// 	m, err := commonDB.DB.GetTokenMapByUidPid(claims.UID, claims.Platform)
// 	if err != nil {
// 		log.NewError(operationID, "get token from redis err", err.Error(), tokensString)
// 		return nil, utils.Wrap(&constant.ErrTokenInvalid, "get token from redis err")
// 	}
// 	if m == nil {
// 		log.NewError(operationID, "get token from redis err", "m is nil", tokensString)
// 		return nil, utils.Wrap(&constant.ErrTokenInvalid, "get token from redis err")
// 	}
// 	if v, ok := m[tokensString]; ok {
// 		switch v {
// 		case constant.NormalToken:
// 			log.NewDebug(operationID, "this is normal return", claims)
// 			return claims, nil
// 		case constant.InValidToken:
// 			return nil, utils.Wrap(&constant.ErrTokenInvalid, "")
// 		case constant.KickedToken:
// 			log.Error(operationID, "this token has been kicked by other same terminal ", constant.ErrTokenKicked)
// 			return nil, utils.Wrap(&constant.ErrTokenKicked, "this token has been kicked by other same terminal ")
// 		case constant.ExpiredToken:
// 			return nil, utils.Wrap(&constant.ErrTokenExpired, "")
// 		default:
// 			return nil, utils.Wrap(&constant.ErrTokenUnknown, "")
// 		}
// 	}
// 	log.NewError(operationID, "redis token map not find", m)
// 	return nil, utils.Wrap(&constant.ErrTokenUnknown, "redis token map not find")
// }

func ParseToken(tokensString, operationID string, gAuthTypeToken bool) (claims *Claims, err error) {
	claims, err = GetClaimFromToken(tokensString, gAuthTypeToken)
	if err != nil {
		log.NewError(operationID, "token validate err", err.Error(), tokensString)
		return nil, utils.Wrap(err, "")
	}

	m, err := commonDB.DB.GetTokenMapByUidPid(claims.UID, claims.Platform)
	if err != nil {
		log.NewError(operationID, "get token from redis err", err.Error(), tokensString)
		return nil, utils.Wrap(&constant.ErrTokenInvalid, "get token from redis err")
	}
	if m == nil {
		log.NewError(operationID, "get token from redis err", "m is nil", tokensString)
		return nil, utils.Wrap(&constant.ErrTokenInvalid, "get token from redis err")
	}
	if v, ok := m[tokensString]; ok {
		switch v {
		case constant.NormalToken:
			log.NewDebug(operationID, "this is normal return", claims)
			return claims, nil
		case constant.InValidToken:
			return nil, utils.Wrap(&constant.ErrTokenInvalid, "")
		case constant.KickedToken:
			log.Error(operationID, "this token has been kicked by other same terminal ", constant.ErrTokenKicked)
			return nil, utils.Wrap(&constant.ErrTokenKicked, "this token has been kicked by other same terminal ")
		case constant.ExpiredToken:
			return nil, utils.Wrap(&constant.ErrTokenExpired, "")
		default:
			return nil, utils.Wrap(&constant.ErrTokenUnknown, "")
		}
	}
	log.NewError(operationID, "redis token map not find", constant.ErrTokenUnknown)
	return nil, utils.Wrap(&constant.ErrTokenUnknown, "redis token map not find")
}

//func MakeTheTokenInvalid(currentClaims *Claims, platformClass string) (bool, error) {
//	storedRedisTokenInterface, err := db.DB.GetPlatformToken(currentClaims.UID, platformClass)
//	if err != nil {
//		return false, err
//	}
//	storedRedisPlatformClaims, err := ParseRedisInterfaceToken(storedRedisTokenInterface)
//	if err != nil {
//		return false, err
//	}
//	//if issue time less than redis token then make this token invalid
//	if currentClaims.IssuedAt.Time.Unix() < storedRedisPlatformClaims.IssuedAt.Time.Unix() {
//		return true, constant.TokenInvalid
//	}
//	return false, nil
//}

//	func ParseRedisInterfaceToken(redisToken interface{}) (*Claims, error) {
//		return GetClaimFromToken(string(redisToken.([]uint8)))
//	}
func ParseRedisInterfaceToken(redisToken interface{}, gAuthTypeToken bool) (*Claims, error) {
	return GetClaimFromToken(string(redisToken.([]uint8)), gAuthTypeToken)
}

func DeleteAdminTokenOnLogout(userID string, gAuthTypeToken bool) error {
	m, err := commonDB.DB.GetTokenMapByUidPid(userID, constant.PlatformIDToName(constant.AdminPlatformID))
	if err != nil && err != go_redis.Nil {
		return utils.Wrap(err, "")
	}
	var deleteTokenKey []string
	for k, v := range m {
		deleteTokenKey = append(deleteTokenKey, k)
		log.Error("Token Added for Delete, ", k, v)
	}
	if len(deleteTokenKey) != 0 {
		err = commonDB.DB.DeleteTokenByUidPid(userID, constant.AdminPlatformID, deleteTokenKey)
		return utils.Wrap(err, "")
	}
	return nil
}

// Validation token, false means failure, true means successful verification
func VerifyToken(token, uid string) (bool, error) {
	gAuthTypeToken := false
	claims, err := ParseToken(token, "", gAuthTypeToken)
	if err != nil {
		return false, utils.Wrap(err, "ParseToken failed")
	}
	if claims.UID != uid {
		return false, &constant.ErrTokenUnknown
	}
	log.NewDebug("", claims.UID, claims.Platform)
	return true, nil
}

func VerifyManagementToken(token, uid string) (bool, error) {
	gAuthTypeToken := false
	claims, err := ParseToken(token, "", gAuthTypeToken)

	if err != nil {
		return false, utils.Wrap(err, "ParseToken failed")
	}
	if claims.UID != uid {
		return false, &constant.ErrTokenUnknown
	}
	if claims.Platform != constant.PlatformIDToName(constant.AdminPlatformID) {
		return false, &constant.ErrTokenInvalid
	}
	log.NewDebug("", claims.UID, claims.Platform)
	return true, nil
}

func WsVerifyToken(token, uid string, platformID string, operationID string) (bool, error, string) {
	argMsg := "token: " + token + " operationID: " + operationID + " userID: " + uid + " platformID: " + platformID
	gAuthTypeToken := false
	claims, err := ParseToken(token, operationID, gAuthTypeToken)
	if err != nil {
		errMsg := "parse token err " + argMsg
		return false, utils.Wrap(err, errMsg), errMsg
	}
	if claims.UID != uid {
		errMsg := " uid is not same to token uid " + " claims.UID " + claims.UID + argMsg
		return false, utils.Wrap(&constant.ErrTokenUnknown, errMsg), errMsg
	}
	if claims.Platform != constant.PlatformIDToName(utils.StringToInt(platformID)) {
		errMsg := " platform is not same to token platform " + argMsg + "claims platformID " + claims.Platform
		return false, utils.Wrap(&constant.ErrTokenUnknown, errMsg), errMsg
	}
	log.NewDebug(operationID, utils.GetSelfFuncName(), " check ok ", claims.UID, uid, claims.Platform)
	return true, nil, ""
}
