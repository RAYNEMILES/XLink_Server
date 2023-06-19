package utils

import (
	"Open_IM/pkg/common/constant"
	db "Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"fmt"
	"github.com/mojocn/base64Captcha"
	"image/color"
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

func OperationIDGenerator() string {
	return strconv.FormatInt(time.Now().UnixNano()+int64(rand.Uint32()), 10)
}

func FriendOpenIMCopyDB(dst *db.Friend, src *open_im_sdk.FriendInfo) {
	utils.CopyStructFields(dst, src)
	dst.FriendUserID = src.FriendUser.UserID
	dst.CreateTime = utils.UnixSecondToTime(int64(src.CreateTime))
}

func FriendDBCopyOpenIM(dst *open_im_sdk.FriendInfo, src *db.Friend) error {
	utils.CopyStructFields(dst, src)
	user, err := imdb.GetUserByUserIDEvenDeleted(src.FriendUserID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	utils.CopyStructFields(dst.FriendUser, user)
	dst.CreateTime = uint32(src.CreateTime.Unix())
	if dst.FriendUser == nil {
		dst.FriendUser = &open_im_sdk.UserInfo{}
	}
	dst.FriendUser.UserID = user.UserID
	dst.FriendUser.CreateTime = int32(user.CreateTime)
	return nil
}

func FriendRequestOpenIMCopyDB(dst *db.FriendRequest, src *open_im_sdk.FriendRequest) {
	utils.CopyStructFields(dst, src)
	dst.CreateTime = utils.UnixSecondToTime(int64(src.CreateTime))
	tmpTime := utils.UnixSecondToTime(int64(src.HandleTime))
	dst.HandleTime = &tmpTime
}

func FriendRequestDBCopyOpenIM(dst *open_im_sdk.FriendRequest, src *db.FriendRequest) error {
	utils.CopyStructFields(dst, src)
	user, err := imdb.GetUserByUserIDEvenDeleted(src.FromUserID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	dst.FromNickname = user.Nickname
	dst.FromFaceURL = user.FaceURL
	dst.FromGender = user.Gender
	dst.FromPhoneNumber = user.PhoneNumber
	dst.FromEmail = user.Email
	user, err = imdb.GetUserByUserIDEvenDeleted(src.ToUserID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	dst.ToNickname = user.Nickname
	dst.ToFaceURL = user.FaceURL
	dst.ToGender = user.Gender
	dst.ToPhoneNumber = user.PhoneNumber
	dst.ToEmail = user.Email
	dst.CreateTime = uint32(src.CreateTime.Unix())
	if src.HandleTime != nil {
		dst.HandleTime = uint32(src.HandleTime.Unix())
	}

	return nil
}

func BlackOpenIMCopyDB(dst *db.Black, src *open_im_sdk.BlackInfo) {
	utils.CopyStructFields(dst, src)
	dst.BlockUserID = src.BlackUserInfo.UserID
	dst.CreateTime = utils.UnixSecondToTime(int64(src.CreateTime))
}

func BlackDBCopyOpenIM(dst *open_im_sdk.BlackInfo, src *db.Black) error {
	utils.CopyStructFields(dst, src)
	dst.CreateTime = uint32(src.CreateTime.Unix())
	user, err := imdb.GetUserByUserID(src.BlockUserID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	utils.CopyStructFields(dst.BlackUserInfo, user)
	return nil
}

func GroupOpenIMCopyDB(dst *db.Group, src *open_im_sdk.GroupInfo) {
	utils.CopyStructFields(dst, src)
}

func GroupDBCopyOpenIM(dst *open_im_sdk.GroupInfo, src *db.Group) error {
	utils.CopyStructFields(dst, src)
	user, err := imdb.GetGroupOwnerInfoByGroupID(src.GroupID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	dst.OwnerUserID = user.UserID

	memberNum, err := imdb.GetGroupMemberNumByGroupID(src.GroupID)
	dst.MemberCount = uint32(memberNum)
	if err != nil {
		return utils.Wrap(err, "")
	}
	dst.CreateTime = uint32(src.CreateTime.Unix())
	return nil
}

func GroupMemberOpenIMCopyDB(dst *db.GroupMember, src *open_im_sdk.GroupMemberFullInfo) {
	utils.CopyStructFields(dst, src)
}

func GroupMemberDBCopyOpenIM(dst *open_im_sdk.GroupMemberFullInfo, src *db.GroupMember) error {
	utils.CopyStructFields(dst, src)
	//if token_verify.IsManagerUserID(src.UserID) {
	//	u, err := imdb.GetUserByUserID(src.UserID)
	//	if err != nil {
	//		return utils.Wrap(err, "")
	//	}
	//
	//	utils.CopyStructFields(dst, u)
	//
	//	dst.AppMangerLevel = 1
	//}
	dst.DeleteTime = src.DeleteTime
	dst.DeletedBy = src.DeletedBy
	dst.UpdateVersion = src.UpdateVersion
	dst.JoinTime = int32(src.JoinTime.Unix())
	if src.MuteEndTime.Unix() < 0 {
		dst.JoinTime = 0
		return nil
	}
	dst.MuteEndTime = uint32(src.MuteEndTime.Unix())
	if dst.MuteEndTime < uint32(time.Now().Unix()) {
		dst.MuteEndTime = 0
	}
	return nil
}

func GroupRequestOpenIMCopyDB(dst *db.GroupRequest, src *open_im_sdk.GroupRequest) {
	utils.CopyStructFields(dst, src)
}

func GroupRequestDBCopyOpenIM(dst *open_im_sdk.GroupRequest, src *db.GroupRequest) {
	utils.CopyStructFields(dst, src)
	dst.ReqTime = uint32(src.ReqTime.Unix())
	dst.HandleTime = uint32(src.HandledTime.Unix())
}

func UserOpenIMCopyDB(dst *db.User, src *open_im_sdk.UserInfo) {
	utils.CopyStructFields(dst, src)
	dst.Birth, _ = utils.TimeStringToTime(src.Birth)
	dst.CreateTime = utils.UnixSecondToTime(int64(src.CreateTime)).Unix()
}

func UserDBCopyOpenIM(dst *open_im_sdk.UserInfo, src *db.User) {
	utils.CopyStructFields(dst, src)
	dst.CreateTime = int32(src.CreateTime)
	dst.Birth = utils.GetTimeStringFromTime(src.Birth)
}

func UserDBCopyOpenIMPublicUser(dst *open_im_sdk.PublicUserInfo, src *db.User) {
	utils.CopyStructFields(dst, src)
}

func GetKeysFromMap(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

var store = db.RedisStore{}

// CaptMake 生成验证码
func CaptMake() (id, b64s string, err error) {
	var driver base64Captcha.Driver
	var driverString base64Captcha.DriverString

	// 配置验证码信息
	captchaConfig := base64Captcha.DriverString{
		Height:          60,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          4,
		Source:          "1234567890qwertyuioplkjhgfdsazxcvbnm",
		BgColor: &color.RGBA{
			R: 3,
			G: 102,
			B: 214,
			A: 125,
		},
		Fonts: []string{"wqy-microhei.ttc"},
	}

	driverString = captchaConfig
	driver = driverString.ConvertFonts()
	captcha := base64Captcha.NewCaptcha(driver, store)
	lid, lb64s, lerr := captcha.Generate()
	return lid, lb64s, lerr
}

// CaptVerify 验证captcha是否正确
func CaptVerify(id string, capt string) bool {
	if store.Verify(id, capt, false) {
		return true
	} else {
		return false
	}
}

func CheckUserPermissions(userID string) *constant.ErrInfo {
	// check user status
	userStatus, _ := db.DB.GetUserStatus(userID)
	if userStatus == "2" {
		log.NewInfo("", utils.GetSelfFuncName(), "userStatus is", userStatus)
		return &constant.ErrUserBanned
	}

	return nil
}

func CheckAdminPermissions(userID string) *constant.ErrInfo {
	// check Admin status
	userStatus, _ := db.DB.GetUserStatus(userID)
	if userStatus == "2" {
		log.NewInfo("", utils.GetSelfFuncName(), "adminStatus is", userStatus)
		return &constant.ErrUserBanned
	}

	return nil
}

func FindMaxInMap(params map[string]int) (k string) {
	var v = 0
	for key, score := range params {
		if score > v {
			k = key
			v = score
		}
	}
	return k
}
func ArrayColumn(array interface{}, key string) (result map[string]interface{}, err error) {
	result = make(map[string]interface{})
	t := reflect.TypeOf(array)
	v := reflect.ValueOf(array)
	if t.Kind() != reflect.Slice {
		return nil, nil
	}
	if v.Len() == 0 {
		return nil, nil
	}

	for i := 0; i < v.Len(); i++ {
		indexv := v.Index(i)
		if indexv.Type().Kind() != reflect.Struct {
			return nil, nil
		}
		mapKeyInterface := indexv.FieldByName(key)
		if mapKeyInterface.Kind() == reflect.Invalid {
			return nil, nil
		}
		mapKeyString, err := interfaceToString(mapKeyInterface.Interface())
		if err != nil {
			return nil, err
		}
		result[mapKeyString] = indexv.Interface()
	}
	return result, err
}

func interfaceToString(v interface{}) (result string, err error) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		result = fmt.Sprintf("%v", v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result = fmt.Sprintf("%v", v)
	case reflect.String:
		result = v.(string)
	default:
		err = nil
	}
	return result, err
}
func ArrayKey(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
