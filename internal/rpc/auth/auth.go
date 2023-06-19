package auth

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/network"
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAuth "Open_IM/pkg/proto/auth"
	"Open_IM/pkg/proto/local_database"
	pbRelay "Open_IM/pkg/proto/relay"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"bytes"
	"compress/zlib"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"

	"Open_IM/pkg/common/config"

	"google.golang.org/grpc"
)

func (rpc *RpcAuth) UserRegister(_ context.Context, req *pbAuth.UserRegisterReq) (*pbAuth.UserRegisterResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc args ", req.String())
	var user db.User
	utils.CopyStructFields(&user, req.UserInfo)
	if req.UserInfo.Birth != "" {
		birth, err := utils.TimeStringToTime(req.UserInfo.Birth)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "check the birth format", req.UserInfo.Birth)
			return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "check the birth format"}}, nil
		}
		user.Birth = birth
	}
	log.Debug(req.OperationID, "copy ", user, req.UserInfo)

	//check phone and email
	if user.PhoneNumber != "" {
		userCheck, _ := imdb.GetRegisterFromPhone(user.PhoneNumber)
		if userCheck != nil && userCheck.UserID != "" {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "user exist！!", userCheck.UserID)
			return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "user exist！!"}}, nil
		}
	}

	if user.Email != "" {
		userCheck, _ := imdb.GetRegisterFromEmail(user.Email)
		if userCheck != nil && userCheck.UserID != "" {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "user exist!！", userCheck.UserID)
			return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "user exist!！"}}, nil
		}
	}

	// default female
	if user.Gender == 0 {
		user.Gender = 2
	}

	// check user source
	user.SourceId = constant.UserRegisterSourceTypeOfficial
	user.SourceCode = ""
	if config.Config.Invite.IsOpen == 1 && imdb.GetInviteCodeIsOpen() {
		inviteCodeInfo := imdb.GetCodeInfoByCode(req.UserInfo.SourceCode)
		if inviteCodeInfo != nil && inviteCodeInfo.State == constant.InviteCodeStateValid {
			user.SourceCode = inviteCodeInfo.Code
			user.SourceId = constant.UserRegisterSourceTypeInvite
		}
	}
	if config.Config.Channel.IsOpen == 1 && imdb.GetChannelCodeIsOpen() {
		channelCodeInfo, _ := imdb.GetInviteChannelCodeByCode(req.UserInfo.SourceCode)
		if channelCodeInfo != nil && channelCodeInfo.State == constant.InviteChannelCodeStateValid {
			user.SourceCode = channelCodeInfo.Code
			user.SourceId = channelCodeInfo.SourceId
		}
	}

	user.Password = req.Password
	user.Uuid = req.UserInfo.Uuid
	err := imdb.UserRegister(user)
	if err != nil {
		errMsg := req.OperationID + " imdb.UserRegister failed " + err.Error() + user.UserID
		log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg, user)
		return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}}, nil
	}

	if user.FaceURL != "" {
		syncUserToLocal(req.OperationID, user.UserID)
	}

	//log.NewInfo(req.OperationID, "SyncToLocalDataBase user", user.Nickname, user.UserID)
	//etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.OperationID)
	//if etcdConn == nil {
	//	log.Error(req.OperationID, "OpenImLocalDataName rpc connect failed ")
	//	return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrRPC.ErrCode, ErrMsg: constant.ErrRPC.ErrMsg}}, nil
	//}
	//client := local_database.NewLocalDataBaseClient(etcdConn)
	//reqPb := local_database.SyncUserInfoReq{
	//	OperationID: req.OperationID,
	//	UserID:      user.UserID,
	//}
	//_, err2 := client.SyncUserInfoToLocal(context.Background(), &reqPb)
	//if err2 != nil {
	//	log.Error(req.OperationID, "SyncUserInfoToLocal failed ", err2.Error(), user.UserID)
	//	return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrRPC.ErrCode, ErrMsg: "SyncUserInfoToLocal failed!"}}, nil
	//}

	go func() {
		// private setting
		imdb.InitUserConfig(user.UserID)

		// interest setting
		//imdb.SetUserInterestType(user.UserID, []int64{constant.InterestDefault})

		officials, err := imdb.GetAllSystemOfficials()
		if err != nil {
			log.Debug("", utils.GetSelfFuncName(), "get all official failed", err.Error())
			return
		}
		for _, official := range officials {
			if err = imdb.AddOfficialFollow(official.Id, user.UserID, user.Gender); err != nil {
				errMsg := req.OperationID + " imdb.FollowOfficialAccount failed " + err.Error() + user.UserID
				log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
				continue
			}
			if err = db.DB.AddOfficialFollow(official.Id, user.UserID); err != nil {
				errMsg := req.OperationID + " db.DB.AddOfficialFollow failed " + err.Error() + user.UserID
				log.NewError(req.OperationID, utils.GetSelfFuncName(), errMsg)
				continue
			}
			chat.OfficialAccountFollowUnfollowNotification(req.OperationID, user.UserID, strconv.FormatInt(official.Id, 10))
		}
	}()

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc return ", pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{}})
	return &pbAuth.UserRegisterResp{CommonResp: &pbAuth.CommonResp{}}, nil
}

func (rpc *RpcAuth) UserToken(_ context.Context, req *pbAuth.UserTokenReq) (*pbAuth.UserTokenResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc args ", req.String())
	user, err := imdb.GetUserByUserID(req.FromUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.FromUserID
		//log.NewError(req.OperationID, errMsg)
		return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}}, nil
	}

	tokens, expTime, err := token_verify.CreateToken(req.FromUserID, int(req.Platform), req.GAuthTypeToken, user.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " token_verify.CreateToken failed " + err.Error() + req.FromUserID + utils.Int32ToString(req.Platform)
		log.NewError(req.OperationID, errMsg)
		return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}}, nil
	}

	// update user last login time ang platform
	go func() {
		imdb.UpdateUserInfoByMap(db.User{
			UserID: req.FromUserID,
		}, map[string]interface{}{
			"last_login_device": req.Platform,
			"last_login_time":   time.Now().Unix(),
		})
	}()

	// get user sign for tencent cloud
	secretKey := config.Config.Trtc.SecretKey
	expireTime := int(config.Config.TokenPolicy.AccessExpire * 24 * 60 * 60)
	log.Debug("", "expireTime：", expireTime)
	log.Debug("", "sdkAppId: ", config.Config.Trtc.SdkAppid, " secretKey:", secretKey, " userID:", req.FromUserID, " expireTime: ", expireTime)
	sig, err := GenSig(config.Config.Trtc.SdkAppid, secretKey, req.FromUserID, expireTime, nil)
	if err != nil {
		errMsg := req.OperationID + " token_verify.CreateToken failed " + err.Error() + req.FromUserID + utils.Int32ToString(req.Platform)
		log.NewError(req.OperationID, errMsg)
		return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}}, nil
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc return ", pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{}, Token: tokens, ExpiredTime: expTime})
	return &pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{}, Token: tokens, ExpiredTime: expTime, UserSign: sig}, nil
}

func (rpc *RpcAuth) ForceLogout(_ context.Context, req *pbAuth.ForceLogoutReq) (*pbAuth.ForceLogoutResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc args ", req.String())
	//if !token_verify.IsManagerUserID(req.OpUserID) {
	//	errMsg := req.OperationID + " IsManagerUserID false " + req.OpUserID
	//	log.NewError(req.OperationID, errMsg)
	//	return &pbAuth.ForceLogoutResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}}, nil
	//}
	if err := token_verify.DeleteToken(req.FromUserID, int(req.Platform), false); err != nil { //req.GAuthTypeToken)
		errMsg := req.OperationID + " DeleteToken failed " + err.Error() + req.FromUserID + utils.Int32ToString(req.Platform)
		log.NewError(req.OperationID, errMsg)
		return &pbAuth.ForceLogoutResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}}, nil
	}
	if err := rpc.forceKickOff(req.FromUserID, req.Platform, req.OperationID); err != nil {
		errMsg := req.OperationID + " forceKickOff failed " + err.Error() + req.FromUserID + utils.Int32ToString(req.Platform)
		log.NewError(req.OperationID, errMsg)
		return &pbAuth.ForceLogoutResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}}, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc return ", pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{}})
	return &pbAuth.ForceLogoutResp{CommonResp: &pbAuth.CommonResp{}}, nil
}

func (rpc *RpcAuth) UpdateUserIPandStatus(_ context.Context, req *pbAuth.UpdateUserIPReq) (*pbAuth.UpdateUserIPResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc args ", req.String())
	ipMap := make(map[string]interface{})
	ipMap["IPaddress"] = req.IPaddress
	ipMap["LastOnlineTime"] = time.Now().UTC().String()
	ipMap["OnlineDifValue"] = 15
	jsonString, _ := json.Marshal(ipMap)
	db.DB.SaveUserIPandStatus(req.UserID, string(jsonString))
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), " rpc return ", pbAuth.UserTokenResp{CommonResp: &pbAuth.CommonResp{}})
	return &pbAuth.UpdateUserIPResp{CommonResp: &pbAuth.CommonResp{}}, nil
}

func (rpc *RpcAuth) GetUserIPandStatus(_ context.Context, req *pbAuth.GetUserIPReq) (*pbAuth.GetUserIPResp, error) {

	// userObj, err := imdb.GetUserByUserID(req.FromUserID)
	// if err != nil {
	// 	return &pbAuth.GetUserIPResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "You are not super user"}}, nil
	// }
	// if userObj.SuperUserStatus != 1 {
	// 	return &pbAuth.GetUserIPResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "You are not super user"}}, nil
	// }

	var userIPResp = pbAuth.GetUserIPResp{}
	jsonString, err := db.DB.GetUserIPandStatus(req.ForUserID)
	if err != nil {
		log.NewError("Get User IP 82", req.ForUserID, err.Error())
	}
	_ = json.Unmarshal([]byte(jsonString), &userIPResp)
	if userIPResp.IPaddress != "" {
		requestURl := config.Config.LocationIpServerAddressPreFix + userIPResp.IPaddress + config.Config.LocationIpServerAddressPostFix
		responseBytes, err := network.DoGetRequest(requestURl)
		if err != nil {
			log.NewError("", "GetUserIPandStatus ", userIPResp.IPaddress, err.Error())
			return &pbAuth.GetUserIPResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrEmptyResponse.ErrCode, ErrMsg: "User details not found"}}, nil
		}
		if len(responseBytes) > 0 {
			jsonInterface := make(map[string]interface{})
			err := json.Unmarshal(responseBytes, &jsonInterface)
			if err != nil {
				log.NewError("", "GetUserIPandStatus ", userIPResp.IPaddress, err.Error())
				return &pbAuth.GetUserIPResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrEmptyResponse.ErrCode, ErrMsg: "User details not found"}}, nil
			}
			if val, ok := jsonInterface["city"]; ok {
				userIPResp.City = val.(string)
			}
			userIPResp.UserID = req.ForUserID
			userIPResp.OperationID = req.OperationID
			layout := "2006-01-02 15:04:05 +0000 UTC"
			tLastOnline, err := time.Parse(layout, userIPResp.LastOnlineTime)
			if err == nil {
				currentTime := time.Now().UTC()
				diff := currentTime.Sub(tLastOnline)
				diffInSec := diff.Seconds()
				if diffInSec > 0 && diffInSec < float64(userIPResp.GetOnlineDifValue()) {
					userIPResp.IsOnline = true
				}
				userIPResp.LastOnlineTime = tLastOnline.Format("2006-01-02 15:04:05")
			}

			userIPResp.CommonResp = &pbAuth.CommonResp{ErrCode: constant.NoError, ErrMsg: ""}
			log.NewInfo(req.OperationID, "User City by IP address ", userIPResp.String())
			return &userIPResp, nil
		}
	}

	return &pbAuth.GetUserIPResp{CommonResp: &pbAuth.CommonResp{ErrCode: constant.ErrEmptyResponse.ErrCode, ErrMsg: "User details not found"}}, nil
}
func (rpc *RpcAuth) forceKickOff(userID string, platformID int32, operationID string) error {

	grpcCons := getcdv3.GetConn4Unique(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOnlineMessageRelayName)
	for _, v := range grpcCons {
		client := pbRelay.NewOnlineMessageRelayServiceClient(v)
		kickReq := &pbRelay.KickUserOfflineReq{OperationID: operationID, KickUserIDList: []string{userID}, PlatformID: platformID}
		log.NewInfo(operationID, "KickUserOffline ", client, kickReq.String())
		_, err := client.KickUserOffline(context.Background(), kickReq)
		return utils.Wrap(err, "")
	}

	return errors.New("no rpc node ")
}

type RpcAuth struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (rpc *RpcAuth) ChangePassword(ctx context.Context, request *pbAuth.ChangePasswordRequest) (*pbAuth.ChangePasswordResponse, error) {
	log.NewInfo(request.OperationID, "ChangePassword ", request.String())
	resp := &pbAuth.ChangePasswordResponse{}

	// check user
	user, err := imdb.GetUserByUserID(request.UserID)
	if err != nil {
		resp.CommonResp = &pbAuth.CommonResp{
			ErrCode: constant.ErrAddInviteCodeUserNotExist.ErrCode,
			ErrMsg:  constant.ErrAddInviteCodeUserNotExist.ErrMsg,
		}
		return resp, nil
	}

	oldPasswordFirst := request.OldPassword + user.Salt
	passwordData := []byte(oldPasswordFirst)
	has := md5.Sum(passwordData)
	oldPasswordHas := fmt.Sprintf("%x", has)

	if user.Password != oldPasswordHas {
		resp.CommonResp = &pbAuth.CommonResp{
			ErrCode: constant.PasswordErr,
			ErrMsg:  "password error!",
		}
		return resp, nil
	}

	// check new password
	if validatePasswordString(request.NewPassword) == false {
		resp.CommonResp = &pbAuth.CommonResp{
			ErrCode: constant.FormattingError,
			ErrMsg:  "password error!",
		}
		return resp, nil
	}

	newPasswordFirst := request.NewPassword + user.Salt
	newPasswordData := []byte(newPasswordFirst)
	has = md5.Sum(newPasswordData)
	newPasswordHas := fmt.Sprintf("%x", has)

	err = imdb.UpdateUserInfoByMap(*user, map[string]interface{}{
		"password": newPasswordHas,
	})
	if err != nil {
		resp.CommonResp = &pbAuth.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return resp, err
	}

	if err := token_verify.DeleteAllToken(request.UserID); err != nil { //req.GAuthTypeToken)
		errMsg := request.OperationID + " DeleteToken failed " + err.Error() + request.UserID
		log.NewError(request.OperationID, errMsg)
	}

	if err := rpc.forceKickOff(request.UserID, 0, request.OperationID); err != nil {
		errMsg := request.OperationID + " forceKickOff failed " + err.Error() + request.UserID
		log.NewError(request.OperationID, errMsg)
	}

	resp.CommonResp = &pbAuth.CommonResp{
		ErrCode: constant.NoError,
		ErrMsg:  "",
	}
	return resp, nil
}

func (rpc *RpcAuth) GetDeviceLoginQrCode(ctx context.Context, request *pbAuth.GetDeviceLoginQrCodeRequest) (*pbAuth.GetDeviceLoginQrCodeResponse, error) {
	log.NewInfo(request.OperationID, "GetDeviceLoginQrCode ", request.String())
	response := &pbAuth.GetDeviceLoginQrCodeResponse{}
	response.CommonResp = &pbAuth.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// generate qr code
	qrCodeId := uuid.NewV4().String()

	if db.DB.QrCodeSaveInfo(qrCodeId, map[string]string{
		"qr_code_id": qrCodeId,
		"device_id":  request.DeviceID,
		"state":      strconv.Itoa(constant.QrLoginStateNormal),
	}) == false {
		response.CommonResp.ErrCode = constant.ErrQrLoginSaveFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginSaveFailed.ErrMsg
		return response, nil
	}

	response.QrCode = qrCodeId
	return response, nil
}

func (rpc *RpcAuth) PushDeviceLoginQrCode(ctx context.Context, request *pbAuth.PushDeviceLoginQrCodeRequest) (*pbAuth.PushDeviceLoginQrCodeResponse, error) {
	log.NewInfo(request.OperationID, "PushDeviceLoginQrCode ", request.String())
	response := &pbAuth.PushDeviceLoginQrCodeResponse{}
	response.CommonResp = &pbAuth.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check qr code
	cache, err := db.DB.QrCodeGetInfo(request.QrCode)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
		return response, nil
	}

	if cache == nil || len(cache) == 0 {
		response.CommonResp.ErrCode = constant.ErrQrLoginNotExist.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginNotExist.ErrMsg
		return response, nil
	}

	// check state
	if _, ok := cache["state"]; !ok {
		response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
		return response, nil
	}
	if cache["state"] != strconv.Itoa(constant.QrLoginStateNormal) {
		response.CommonResp.ErrCode = constant.ErrQrLoginStateErr.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginStateErr.ErrMsg
		return response, nil
	}

	// check device id
	if _, ok := cache["device_id"]; !ok {
		response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
		return response, nil
	}

	// bind user
	temporaryToken := uuid.NewV5(uuid.FromStringOrNil(request.QrCode), request.UserId).String()

	if db.DB.QrCodeSaveInfo(request.QrCode, map[string]string{
		"qr_code_id": request.QrCode,
		"device_id":  cache["device_id"],
		"state":      strconv.Itoa(constant.QrLoginStateWaitForConfirmation),
		"user_id":    request.UserId,
		"token":      temporaryToken,
	}) == false {
		response.CommonResp.ErrCode = constant.ErrQrLoginUpdateStateFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginUpdateStateFailed.ErrMsg
		return response, nil
	}

	response.TemporaryToken = temporaryToken
	return response, nil
}

func (rpc *RpcAuth) ConfirmDeviceLoginQrCode(ctx context.Context, request *pbAuth.ConfirmDeviceLoginQrCodeRequest) (*pbAuth.ConfirmDeviceLoginQrCodeResponse, error) {
	log.NewInfo(request.OperationID, "ConfirmDeviceLoginQrCode ", request.String())
	response := &pbAuth.ConfirmDeviceLoginQrCodeResponse{}
	response.CommonResp = &pbAuth.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// agree
	if request.Agree == true {
		// check qr code
		cache, err := db.DB.QrCodeGetInfo(request.QrCode)
		if err != nil {
			response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
			return response, nil
		}

		if cache == nil || len(cache) == 0 {
			response.CommonResp.ErrCode = constant.ErrQrLoginNotExist.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginNotExist.ErrMsg
			return response, nil
		}

		// check state
		if _, ok := cache["state"]; !ok {
			response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
			return response, nil
		}

		if cache["state"] != strconv.Itoa(constant.QrLoginStateWaitForConfirmation) {
			response.CommonResp.ErrCode = constant.ErrQrLoginStateErr.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginStateErr.ErrMsg
			return response, nil
		}

		// check device id
		if _, ok := cache["device_id"]; !ok {
			response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
			return response, nil
		}

		// check token
		if _, ok := cache["token"]; !ok {
			response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
			return response, nil
		}

		if cache["token"] != request.TemporaryToken {
			response.CommonResp.ErrCode = constant.ErrQrLoginTokenErr.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginTokenErr.ErrMsg
			return response, nil
		}

		if db.DB.QrCodeSaveInfo(request.QrCode, map[string]string{
			"qr_code_id": request.QrCode,
			"device_id":  cache["device_id"],
			"state":      strconv.Itoa(constant.QrLoginStateConfirmed),
			"user_id":    request.UserId,
			"token":      cache["token"],
		}) == false {
			response.CommonResp.ErrCode = constant.ErrQrLoginUpdateStateFailed.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginUpdateStateFailed.ErrMsg
			return response, nil
		}
	}

	// refuse
	if request.Agree == false {

	}

	return response, nil
}

func (rpc *RpcAuth) CheckDeviceLoginQrCodeState(ctx context.Context, request *pbAuth.CheckDeviceLoginQrCodeStateRequest) (*pbAuth.CheckDeviceLoginQrCodeStateResponse, error) {
	log.NewInfo(request.OperationID, "CheckDeviceLoginQrCodeState ", request.String())
	response := &pbAuth.CheckDeviceLoginQrCodeStateResponse{}
	response.CommonResp = &pbAuth.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.State = constant.QrLoginStateReserve
	response.UserId = ""

	// check qr code
	cache, err := db.DB.QrCodeGetInfo(request.QrCode)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
		return response, nil
	}

	if cache == nil || len(cache) == 0 {
		response.State = constant.QrLoginStateExpired
		return response, nil
	}

	// check device id
	if _, ok := cache["device_id"]; !ok {
		response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
		return response, nil
	}
	if cache["device_id"] != request.DeviceID {
		response.CommonResp.ErrCode = constant.ErrQrLoginDeviceIdErr.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginDeviceIdErr.ErrMsg
		return response, nil
	}

	// check state
	if _, ok := cache["state"]; !ok {
		response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
		response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
		return response, nil
	}

	state, _ := strconv.Atoi(cache["state"])
	response.State = int32(state)
	response.UserId = ""

	if state == constant.QrLoginStateConfirmed {
		if _, ok := cache["user_id"]; !ok {
			response.CommonResp.ErrCode = constant.ErrQrLoginGetFailed.ErrCode
			response.CommonResp.ErrMsg = constant.ErrQrLoginGetFailed.ErrMsg
			return response, nil
		}
		response.UserId = cache["user_id"]
	}

	return response, nil
}

func NewRpcAuthServer(port int) *RpcAuth {
	return &RpcAuth{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImAuthName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (rpc *RpcAuth) Run() {
	log.NewPrivateLog(constant.OpenImAuthLog)
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, "rpc auth start...")

	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(rpc.rpcPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + rpc.rpcRegisterName)
	}
	log.NewInfo(operationID, "listen network success, ", address, listener)
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd
	pbAuth.RegisterAuthServer(srv, rpc)
	rpcRegisterIP := ""
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.NewError(operationID, "RegisterEtcd failed ", err.Error(),
			rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)
		return
	}
	log.NewInfo(operationID, "RegisterAuthServer ok ", rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)
	err = srv.Serve(listener)
	if err != nil {
		log.NewError(operationID, "Serve failed ", err.Error())
		return
	}
	log.NewInfo(operationID, "rpc auth ok")
}
func syncUserToLocal(operationID, userID string) error {
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)

	log.NewError(operationID, utils.GetSelfFuncName(), userID)
	userInfo, err := imdb.GetUserByUserID(userID)
	if err != nil {
		log.NewError(operationID, "GetUserByUserID failed")
		return err
	}

	localUserInfo := sdkws.UserInfo{}
	err = utils.CopyStructFields(&localUserInfo, &userInfo)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}
	userIDList := []string{userID}

	syncConvReq := &local_database.SyncDataReq{
		OperationID:  operationID,
		MsgType:      constant.SyncUserInfo,
		MemberIDList: userIDList,
		UserInfo:     &localUserInfo,
	}
	localConvResp, err := localDataClient.SyncData(context.Background(), syncConvReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}

	if localConvResp.ErrCode != 0 {
		log.NewError(operationID, "SyncData rpc logic call failed ", localConvResp.String())
		return errors.New("SyncData rpc logic call failed")
	}
	return nil
}

func validatePasswordString(str string) bool {
	length := len(str)
	if length < 6 || length > 20 {
		return false
	}
	var hasLetter, hasNumber bool
	for _, char := range str {
		if unicode.IsLetter(char) {
			hasLetter = true
		} else if unicode.IsNumber(char) {
			hasNumber = true
		}
		if hasLetter && hasNumber {
			return true
		}
	}
	return false
}

func hmacsha256(sdkappid string, key string, identifier string, currTime int64, expire int, base64UserBuf *string) string {
	var contentToBeSigned string
	contentToBeSigned = "TLS.identifier:" + identifier + "\n"
	contentToBeSigned += "TLS.sdkappid:" + sdkappid + "\n"
	contentToBeSigned += "TLS.time:" + strconv.FormatInt(currTime, 10) + "\n"
	contentToBeSigned += "TLS.expire:" + strconv.Itoa(expire) + "\n"
	if nil != base64UserBuf {
		contentToBeSigned += "TLS.userbuf:" + *base64UserBuf + "\n"
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(contentToBeSigned))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func base64urlEncode(data []byte) string {
	str := base64.StdEncoding.EncodeToString(data)
	str = strings.Replace(str, "+", "*", -1)
	str = strings.Replace(str, "/", "-", -1)
	str = strings.Replace(str, "=", "_", -1)
	return str
}

func GenSig(sdkappid string, key string, identifier string, expire int, userbuf []byte) (string, error) {
	currTime := time.Now().Unix()
	sigDoc := make(map[string]interface{})
	sigDoc["TLS.ver"] = "2.0"
	sigDoc["TLS.identifier"] = identifier
	sigDoc["TLS.sdkappid"] = sdkappid
	sigDoc["TLS.expire"] = expire
	sigDoc["TLS.time"] = currTime
	var base64UserBuf string
	if nil != userbuf {
		base64UserBuf = base64.StdEncoding.EncodeToString(userbuf)
		sigDoc["TLS.userbuf"] = base64UserBuf
		sigDoc["TLS.sig"] = hmacsha256(sdkappid, key, identifier, currTime, expire, &base64UserBuf)
	} else {
		sigDoc["TLS.sig"] = hmacsha256(sdkappid, key, identifier, currTime, expire, nil)
	}

	data, err := json.Marshal(sigDoc)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err = w.Write(data); err != nil {
		return "", err
	}
	if err = w.Close(); err != nil {
		return "", err
	}
	return base64urlEncode(b.Bytes()), nil
}
