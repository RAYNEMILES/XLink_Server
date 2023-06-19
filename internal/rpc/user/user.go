package user

import (
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/internal/rpc/admin_cms"
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	errors2 "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils3 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	rpcAuth "Open_IM/pkg/proto/auth"
	pbChat "Open_IM/pkg/proto/chat"
	pbFriend "Open_IM/pkg/proto/friend"
	rpc "Open_IM/pkg/proto/friend"
	groupRpc "Open_IM/pkg/proto/group"
	"Open_IM/pkg/proto/local_database"
	pbRelay "Open_IM/pkg/proto/relay"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type userServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (s *userServer) UserIdIsExist(_ context.Context, request *pbUser.UserIdIsExistRequest) (*pbUser.UserIdIsExistResponse, error) {
	log.NewInfo(request.OperationID, "UserIdIsExist args ", request.String())
	response := &pbUser.UserIdIsExistResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.IsExist = false

	if request.UserId == "" {
		return response, nil
	}

	_, err := imdb.GetUserNameByUserID(request.UserId)
	if err != nil {
		return response, nil
	}

	response.IsExist = true
	return response, nil
}

func (s *userServer) GetUserPrivacyByUserIdList(ctx context.Context, request *pbUser.GetUserPrivacyByUserIdListRequest) (*pbUser.GetUserPrivacyByUserIdListResponse, error) {
	log.NewInfo(request.OperationID, "GetUserPrivacyByUserIdList args ", request.String())
	response := &pbUser.GetUserPrivacyByUserIdListResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	privacySetting, err := imdb.GetUserPrivacyByUserIdList(request.UserIdList)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return response, nil
	}

	for _, v := range privacySetting {
		response.Result[v.UserId].UserId = v.UserId
		response.Result[v.UserId].Privacy = append(response.Result[v.UserId].Privacy, &pbUser.Privacy{
			SettingKey:   v.SettingKey,
			SettingValue: v.SettingVal,
		})
	}

	return response, nil
}

func (s *userServer) SetPrivacy(ctx context.Context, request *pbUser.SetUserPrivacyRequest) (*pbUser.SetUserPrivacyResponse, error) {
	log.NewInfo(request.OperationID, "SetPrivacy args ", request.String())
	response := &pbUser.SetUserPrivacyResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	for _, v := range request.Privacy {
		if v.SettingKey == "" {
			continue
		}

		err := imdb.UpdatePrivacySetting(request.UserId, v.SettingKey, v.SettingValue)
		if err != nil {
			response.CommonResp.ErrCode = constant.ErrDB.ErrCode
			return response, nil
		}
	}

	err := db.DB.DelPrivacySettingByUserId(request.UserId)
	if err != nil {
		log.NewError(request.OperationID, "SetPrivacy DelPrivacySettingByUserId err ", err.Error())
	}

	return response, nil
}

func (s *userServer) GetPrivacy(ctx context.Context, request *pbUser.GetUserPrivacyRequest) (*pbUser.GetUserPrivacyResponse, error) {
	log.NewInfo(request.OperationID, "GetPrivacy args ", request.String())
	response := &pbUser.GetUserPrivacyResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// cache
	cache, err := db.DB.GetPrivacySettingByUserId(request.UserId)
	if err == nil && len(cache) > 0 {
		for k, v := range cache {
			response.Privacy = append(response.Privacy, &pbUser.Privacy{
				SettingKey:   k,
				SettingValue: v,
			})
		}
		return response, nil
	}

	// db
	privacySetting, err := imdb.GetPrivacySetting(request.UserId)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return response, nil
	}

	log.NewError(request.OperationID, "GetPrivacy privacySetting ", len(privacySetting), privacySetting)

	if len(privacySetting) == 0 {
		imdb.InitUserConfig(request.UserId)
		privacySetting, err = imdb.GetPrivacySetting(request.UserId)
		if err != nil {
			response.CommonResp.ErrCode = constant.ErrDB.ErrCode
			return response, nil
		}
	}

	cacheMap := map[string]string{}
	for _, v := range privacySetting {
		response.Privacy = append(response.Privacy, &pbUser.Privacy{
			SettingKey:   v.SettingKey,
			SettingValue: v.SettingVal,
		})
		cacheMap[v.SettingKey] = v.SettingVal
	}

	// redis
	err = db.DB.SetPrivacySettingByUserId(request.UserId, cacheMap)
	if err != nil {
		log.NewWarn(request.OperationID, "GetPrivacy set redis err ", err.Error())
	}

	return response, nil
}

func (s *userServer) GetYouKnowUsersByContactList(ctx context.Context, request *pbUser.GetYouKnowUsersByContactListRequest) (*pbUser.GetYouKnowUsersByContactListResponse, error) {
	log.NewInfo(request.OperationID, "GetUsersThirdInfo args ", request.String())
	response := &pbUser.GetYouKnowUsersByContactListResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.User = []*pbUser.User{}

	cache, _ := db.DB.GetUsersInfoByUserIdCache(request.UserId)
	if cache != "" {
		json.Unmarshal([]byte(cache), &response.User)
		return response, nil
	}

	phoneList := request.PhoneNumber
	userId := request.UserId
	nowList := imdb.GetYouKnowUsersByContactList(userId, phoneList)
	for _, v := range nowList {
		response.User = append(response.User, &pbUser.User{
			ProfilePhoto: v.FaceURL,
			UserId:       v.UserID,
			Nickname:     v.Nickname,
			Gender:       v.Gender,
			Email:        v.Email,
			PhoneNumber:  v.PhoneNumber,
		})
	}

	marshal, _ := json.Marshal(response.User)
	_ = db.DB.SetUsersInfoByUserIdCache(userId, string(marshal), 1379)

	return response, nil
}

func (s *userServer) GetUsersThirdInfo(ctx context.Context, request *pbUser.GetUsersThirdInfoRequest) (*pbUser.GetUsersThirdInfoResponse, error) {
	log.NewInfo(request.OperationID, "GetUsersThirdInfo args ", request.String())
	response := &pbUser.GetUsersThirdInfoResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	where := map[string]string{}
	where["user_id"] = request.UserId
	where["third_type"] = request.ThirdType
	where["third_name"] = request.ThirdName

	byWhere, err := imdb.GetUsersThirdInfoByWhere(where, request.Pagination.ShowNumber, request.Pagination.PageNumber)

	if err != nil {
		log.NewInfo(request.OperationID, "GetUsersThirdInfo err ", err.Error())
		response.CommonResp = &pbUser.CommonResp{
			ErrCode: constant.OK.ErrCode,
			ErrMsg:  constant.OK.ErrMsg,
		}
		return response, nil
	}

	for _, v := range byWhere {
		thirdInfo := &pbUser.UserThirdInfo{
			Nickname:     v["name"].(string),
			PhoneNumber:  v["phone_number"].(string),
			UserId:       v["user_id"].(string),
			Email:        v["email"].(string),
			OfficialName: v["official"].(string),
			Facebook:     v["facebook"].(string),
			Google:       v["google"].(string),
			Apple:        v["apple"].(string),
		}
		response.UserThirdInfo = append(response.UserThirdInfo, thirdInfo)
	}

	userNums, _ := imdb.GetUsersThirdInfoCountByWhere(where)
	response.UserNums = int32(userNums)

	return response, nil
}

func (s *userServer) GetUsersInfoByPhoneList(ctx context.Context, request *pbUser.GetUsersInfoByPhoneListRequest) (*pbUser.GetUsersInfoByPhoneListResponse, error) {
	phoneList := request.PhoneNumber

	response := &pbUser.GetUsersInfoByPhoneListResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	result := imdb.GetUsersByPhoneList(phoneList)
	for _, v := range result {
		_, err := imdb.UserIsBlock(v.UserID)
		if err != nil {
			continue
		}

		user := &pbUser.User{
			ProfilePhoto: v.FaceURL,
			UserId:       v.UserID,
			CreateTime:   v.CreateTime,
			Nickname:     v.Nickname,
			PhoneNumber:  v.PhoneNumber,
			Status:       int32(v.Status),
		}
		response.User = append(response.User, user)
	}

	return response, nil
}

func (s *userServer) ExistsUser(ctx context.Context, request *pbUser.ExistsUserRequest) (*pbUser.ExistsUserResponse, error) {
	var response = &pbUser.ExistsUserResponse{}
	userId := request.GetUserId()
	phoneNumber := request.GetPhoneNumber()

	if userId != "" {
		_, err := imdb.GetUserByUserID(userId)
		if err == nil {
			response.CommonResp = &pbUser.CommonResp{
				ErrCode: constant.ErrUserExistUserIdExist.ErrCode,
				ErrMsg:  constant.ErrUserExistUserIdExist.ErrMsg,
			}
			return response, nil
		}
	}

	if phoneNumber != "" {
		_, err := imdb.GetUserByPhoneNumber(phoneNumber)
		if err == nil {
			response.CommonResp = &pbUser.CommonResp{
				ErrCode: constant.ErrUserExistPhoneNumberExist.ErrCode,
				ErrMsg:  constant.ErrUserExistPhoneNumberExist.ErrMsg,
			}
			return response, nil
		}
	}

	if userId == "" && phoneNumber == "" {
		response.CommonResp = &pbUser.CommonResp{
			ErrCode: constant.ErrUserExistArg.ErrCode,
			ErrMsg:  constant.ErrUserExistArg.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *userServer) GetIviteLink(ctx context.Context, request *pbUser.GetInviteLinkRequest) (*pbUser.GetInviteLinkResponse, error) {
	UserId := request.GetUserId()
	CodeInfo, _ := imdb.GetCodeByUserID(UserId)
	if CodeInfo == nil || CodeInfo.Code == "" {
		// get user info
		userInfo, _ := imdb.GetUserByUserID(UserId)
		inviteCode := utils.GenerateInviteCode(int64(userInfo.ID))
		imdb.AddCode(request.UserId, inviteCode, "", "")

		CodeInfo.Code = inviteCode
	}

	configInfo, _ := imdb.GetConfigByName("invite_code_base_link")
	baseLink := configInfo.Value

	return &pbUser.GetInviteLinkResponse{
		CommonResp: &pbUser.CommonResp{},
		InviteLink: baseLink + "/" + CodeInfo.Code,
	}, nil
}

func NewUserServer(port int) *userServer {
	return &userServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImUserName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *userServer) Run() {
	log.NewPrivateLog(constant.OpenImUserLog)
	log.NewInfo("0", "rpc user start...")

	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)

	// listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("0", "listen network success, address ", address, listener)
	defer listener.Close()
	// grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	// Service registers with etcd
	pbUser.RegisterUserServer(srv, s)
	rpcRegisterIP := ""
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(), s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName)
		return
	}
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "rpc  user success")
}

func syncPeerUserConversation(conversation *pbUser.Conversation, operationID string) error {
	peerUserConversation := db.Conversation{
		OwnerUserID:      conversation.UserID,
		ConversationID:   utils.GetConversationIDBySessionType(conversation.OwnerUserID, constant.SingleChatType),
		ConversationType: constant.SingleChatType,
		UserID:           conversation.OwnerUserID,
		GroupID:          "",
		RecvMsgOpt:       0,
		UnreadCount:      0,
		DraftTextTime:    0,
		IsPinned:         false,
		IsPrivateChat:    conversation.IsPrivateChat,
		AttachedInfo:     "",
		Ex:               "",
	}
	err := imdb.PeerUserSetConversation(peerUserConversation)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
		return err
	}
	chat.ConversationSetPrivateNotification(operationID, conversation.OwnerUserID, conversation.UserID, conversation.IsPrivateChat)
	return nil
}

func (s *userServer) SearchUser(ctx context.Context, request *pbUser.SearchUserRequest) (*pbUser.SearchUserResponse, error) {
	log.NewInfo(request.OperationID, "SearchUser args ", request.String())
	response := &pbUser.SearchUserResponse{}
	response.CommonResp = &pbUser.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	var userInfoList []*sdkws.UserInfo
	if len(request.UserIDList) > 0 {
		for _, key := range request.UserIDList {
			var userInfo sdkws.UserInfo
			userList, err := imdb.GetUserByUserIdOrPhoneOrNicknameOrEmail(key)
			if err != nil {
				log.NewError(request.OperationID, "GetUserByUserID failed ", err.Error(), key)
				continue
			}
			if len(userList) > 0 {
				for _, user := range userList {
					// privacy
					userPrivacy, err := s.GetPrivacy(ctx, &pbUser.GetUserPrivacyRequest{
						OperationID: request.OperationID,
						UserId:      user.UserID,
					})
					isPrivacy := false
					isWoomsOpen := true
					if err == nil {
						for _, privacy := range userPrivacy.Privacy {
							if privacy.SettingKey == constant.PrivacyAddByPhone && privacy.SettingValue == constant.PrivacyStatusClose {
								if user.PhoneNumber != "" && strings.ToLower(user.PhoneNumber) == strings.ToLower(key) {
									isPrivacy = true
								}
							}
							if privacy.SettingKey == constant.PrivacyAddByAccount && privacy.SettingValue == constant.PrivacyStatusClose {
								if user.UserID != "" && strings.ToLower(user.UserID) == strings.ToLower(key) {
									isPrivacy = true
								}
							}
							if privacy.SettingKey == constant.PrivacyAddByEmail && privacy.SettingValue == constant.PrivacyStatusClose {
								if user.Email != "" && strings.ToLower(user.Email) == strings.ToLower(key) {
									isPrivacy = true
								}
							}

							// wooms privacy
							if privacy.SettingKey == constant.PrivacySeeWooms && privacy.SettingValue == constant.PrivacyStatusClose {
								isWoomsOpen = false
							}
						}
					}
					if isPrivacy {
						continue
					}

					utils.CopyStructFields(&userInfo, user)

					userInfo.MomentsPicArray = ""
					if isWoomsOpen {
						moments, err := imdb.GetMomentsByUserID(user.UserID)
						if err != nil {
							log.NewError(request.OperationID, "GetMomentsByUserID failed ", err.Error(), user.UserID)
							continue
						}
						for _, moment := range moments {
							userInfo.MomentsPicArray = moment.MContentImagesArray
							if moment.MContentImagesArray == "" {
								userInfo.MomentsPicArray = moment.MContentThumbnilArray
							}
						}
					}

					userInfo.Birth = utils.GetTimeStringFromTime(user.Birth)
					userInfoList = append(userInfoList, &userInfo)
				}
			}
		}
	} else {
		return &pbUser.SearchUserResponse{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}
	log.NewInfo(request.OperationID, "GetUserInfo rpc return ", pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{}, UserInfoList: userInfoList})
	return &pbUser.SearchUserResponse{CommonResp: &pbUser.CommonResp{}, UserInfoList: userInfoList}, nil
}

func (s *userServer) GetUserInfo(ctx context.Context, req *pbUser.GetUserInfoReq) (*pbUser.GetUserInfoResp, error) {
	log.NewInfo(req.OperationID, "GetUserInfo args ", req.String())
	var userInfoList []*sdkws.UserInfo
	connection, FriendGRPErr := createFriendGRPConnection(req.OperationID)
	if len(req.UserIDList) > 0 {
		for _, userID := range req.UserIDList {
			var userInfo sdkws.UserInfo
			user, err := imdb.GetUserByUserID(userID)
			if err != nil {
				log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), userID)
				continue
			}
			utils.CopyStructFields(&userInfo, user)

			// is friend
			userInfo.IsFriend = true
			if FriendGRPErr == nil {
				rpcRequest := &pbFriend.IsFriendReq{CommID: &pbFriend.CommID{
					OperationID: req.OperationID,
					FromUserID:  req.OpUserID,
					ToUserID:    userID,
				}}
				friend, err := connection.IsFriend(ctx, rpcRequest)
				if err == nil {
					userInfo.IsFriend = friend.Response
				}
			}

			moments, err := imdb.GetMomentsByUserID(userID)
			if err != nil {
				log.NewError(req.OperationID, "GetMomentsByUserID failed ", err.Error(), userID)
				continue
			}
			for _, moment := range moments {
				userInfo.MomentsPicArray = moment.MContentImagesArray
				if moment.MContentImagesArray == "" {
					userInfo.MomentsPicArray = moment.MContentThumbnilArray
				}
			}

			isFriend := userInfo.IsFriend
			if req.OpUserID == userID {
				isFriend = true
			}
			posts, err := db.DB.GetUserMomentCount(userID, isFriend)
			if err != nil {
				errMsg := "Get Moment is failed " + err.Error()
				log.Error(errMsg)
				continue
			}
			userInfo.MomentsCount = int32(posts)

			userInfo.Birth = utils.GetTimeStringFromTime(user.Birth)
			userInfoList = append(userInfoList, &userInfo)
		}
	} else {
		return &pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}
	log.NewInfo(req.OperationID, "GetUserInfo rpc return ", pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{}, UserInfoList: userInfoList})
	return &pbUser.GetUserInfoResp{CommonResp: &pbUser.CommonResp{}, UserInfoList: userInfoList}, nil
}

func (s *userServer) BatchSetConversations(ctx context.Context, req *pbUser.BatchSetConversationsReq) (*pbUser.BatchSetConversationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	if req.NotificationType == 0 {
		req.NotificationType = constant.ConversationOptChangeNotification
	}

	resp := &pbUser.BatchSetConversationsResp{}
	for _, v := range req.Conversations {
		conversation := db.Conversation{}
		if err := utils.CopyStructFields(&conversation, v); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), v.String(), "CopyStructFields failed", err.Error())
		}
		// redis op
		if err := db.DB.SetSingleConversationRecvMsgOpt(req.OwnerUserID, v.ConversationID, v.RecvMsgOpt); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
			resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
			return resp, nil
		}
		if err := imdb.SetConversation(conversation); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
			resp.Failed = append(resp.Failed, v.ConversationID)
			continue
		}
		resp.Success = append(resp.Success, v.ConversationID)
		// if is set private chat operation，then peer user need to sync and set tips\
		if v.ConversationType == constant.SingleChatType && req.NotificationType == constant.ConversationPrivateChatNotification {
			if err := syncPeerUserConversation(v, req.OperationID); err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "syncPeerUserConversation", err.Error())
			}
		}

		//data synchronization
		// syncConversationToLocal(req.OperationID, conversation.OwnerUserID, conversation.ConversationID, constant.SyncConversation, conversation.ConversationType)

	}
	//todo check constant.OpFromFrontend if required
	chat.ConversationChangeNotification(req.OperationID, req.OwnerUserID, constant.OpFromFrontend)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) GetAllConversations(ctx context.Context, req *pbUser.GetAllConversationsReq) (*pbUser.GetAllConversationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetAllConversationsResp{Conversations: []*pbUser.Conversation{}}
	conversations, err := imdb.GetUserAllConversations(req.OwnerUserID)
	log.NewDebug(req.OperationID, "conversations: ", conversations)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetConversations error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err = utils.CopyStructFields(&resp.Conversations, conversations); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields error", err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) GetConversation(ctx context.Context, req *pbUser.GetConversationReq) (*pbUser.GetConversationResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetConversationResp{Conversation: &pbUser.Conversation{}}
	conversation, err := imdb.GetConversation(req.OwnerUserID, req.ConversationID)
	log.NewDebug("", utils.GetSelfFuncName(), "conversation", conversation)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetConversation error", err.Error(), req.String())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(resp.Conversation, &conversation); err != nil {
		log.Debug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields error", conversation, err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "CopyStructFields error"}
		return resp, nil
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) GetConversations(ctx context.Context, req *pbUser.GetConversationsReq) (*pbUser.GetConversationsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetConversationsResp{Conversations: []*pbUser.Conversation{}}
	conversations, err := imdb.GetConversations(req.OwnerUserID, req.ConversationIDs)
	log.NewDebug("", utils.GetSelfFuncName(), "conversations", conversations)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetConversations error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	if err := utils.CopyStructFields(&resp.Conversations, conversations); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", conversations, err.Error())
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) SetConversation(ctx context.Context, req *pbUser.SetConversationReq) (*pbUser.SetConversationResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.SetConversationResp{}
	if req.NotificationType == 0 {
		req.NotificationType = constant.ConversationOptChangeNotification
	}
	if req.Conversation.ConversationType == constant.GroupChatType {
		groupInfo, err := imdb.GetGroupInfoByGroupID(req.Conversation.GroupID)
		if err != nil {
			log.NewError(req.OperationID, "GetGroupInfoByGroupID failed ", req.Conversation.GroupID, err.Error())
			resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
			return resp, nil
		}
		if groupInfo.Status == constant.GroupStatusDismissed && !req.Conversation.IsNotInGroup {
			log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "group status is dismissed", groupInfo)
			errMsg := "group status is dismissed"
			resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}
			return resp, nil
		}
	}
	var conversation db.Conversation
	if err := utils.CopyStructFields(&conversation, req.Conversation); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", *req.Conversation, err.Error())
	}
	if err := db.DB.SetSingleConversationRecvMsgOpt(req.Conversation.OwnerUserID, req.Conversation.ConversationID, req.Conversation.RecvMsgOpt); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	err := imdb.SetConversation(conversation)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}

	//Sync Conversation
	//reqConvLocalData := local_database.SyncConversationReq{
	//	OperationID:    req.OperationID,
	//	OwnerID:        req.Conversation.OwnerUserID,
	//	ConversationID: req.Conversation.ConversationID,
	//}
	//etcdConnLocalData := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.OperationID)
	//if etcdConnLocalData == nil {
	//	log.NewError(req.OperationID, "Sync Conversation rpc failed ", req.Conversation.OwnerUserID, req.Conversation.ConversationID)
	//	resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrRPC.ErrCode, ErrMsg: constant.ErrRPC.ErrMsg}
	//	return resp, nil
	//}
	//
	//localDataClient := local_database.NewLocalDataBaseClient(etcdConnLocalData)
	//respLocalData, err := localDataClient.SyncConversationToLocal(context.Background(), &reqConvLocalData)
	//if err != nil {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "SyncConversationToLocal rpc failed, ", respLocalData.String(), err.Error(), req.Conversation.OwnerUserID, req.Conversation.ConversationID)
	//} else {
	//	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "SyncConversationToLocal success", respLocalData.String(), req.Conversation.OwnerUserID, req.Conversation.ConversationID)
	//}

	// notification
	if req.Conversation.ConversationType == constant.SingleChatType && req.NotificationType == constant.ConversationPrivateChatNotification {
		// sync peer user conversation if conversation is singleChatType
		if err := syncPeerUserConversation(req.Conversation, req.OperationID); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "syncPeerUserConversation", err.Error())
			resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
			return resp, nil
		}
	} else {
		chat.ConversationChangeNotification(req.OperationID, req.Conversation.OwnerUserID, constant.OpFromFrontend)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "rpc return", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) SetRecvMsgOpt(ctx context.Context, req *pbUser.SetRecvMsgOptReq) (*pbUser.SetRecvMsgOptResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.SetRecvMsgOptResp{}
	var conversation db.Conversation
	if err := utils.CopyStructFields(&conversation, req); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", *req, err.Error())
	}
	if err := db.DB.SetSingleConversationRecvMsgOpt(req.OwnerUserID, req.ConversationID, req.RecvMsgOpt); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	stringList := strings.Split(req.ConversationID, "_")
	if len(stringList) > 1 {
		switch stringList[0] {
		case "single_":
			conversation.UserID = stringList[1]
			conversation.ConversationType = constant.SingleChatType
		case "group":
			conversation.GroupID = stringList[1]
			conversation.ConversationType = constant.GroupChatType
		}
	}
	err := imdb.SetRecvMsgOpt(conversation)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
		resp.CommonResp = &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}
		return resp, nil
	}
	chat.ConversationChangeNotification(req.OperationID, req.OwnerUserID, constant.OpFromFrontend)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	resp.CommonResp = &pbUser.CommonResp{}
	return resp, nil
}

func (s *userServer) DeleteUsers(_ context.Context, req *pbUser.DeleteUsersReq) (*pbUser.DeleteUsersResp, error) {
	log.NewInfo(req.OperationID, "DeleteUsers args ", req.String())
	var common pbUser.CommonResp
	resp := pbUser.DeleteUsersResp{CommonResp: &common}

	for _, userID := range req.DeleteUserIDList {
		i := imdb.DeleteUser(userID, "", req.OpUserID)
		if i == 0 {
			log.NewError(req.OperationID, "delete user error", userID)
			common.ErrCode = 201
			common.ErrMsg = "some uid deleted failed"
			resp.FailedUserIDList = append(resp.FailedUserIDList, userID)
		} else {
			go UpdateDeletedAccountInFriends(userID)
			go UpdateDeletedAccountInMoments(userID)
			go UpdateDeletedAccountInGroupMembers(userID)
			go UpdateDeletedAccountInOfficialAccountAndArticle(userID)
			go UpdateDeletedAccountInLikes(userID)

			// delete all short video
			go UpdateDeletedAccountInWooms(userID)

		}

		err := imdb.DeleteUserInterests(userID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete user interest failed")
			return nil, err
		}

		if err := token_verify.DeleteAllToken(userID); err != nil { //req.GAuthTypeToken)
			errMsg := req.OperationID + " DeleteToken failed " + err.Error() + userID
			log.NewError(req.OperationID, errMsg)
		}

		if err := forceKickOff(userID, 0, req.OperationID); err != nil {
			errMsg := req.OperationID + " forceKickOff failed " + err.Error() + userID
			log.NewError(req.OperationID, errMsg)
		}
	}
	log.NewInfo(req.OperationID, "DeleteUsers rpc return ", resp.String())
	return &resp, nil
}

func UpdateDeletedAccountInFriends(userID string) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return
	}
	updateInterface := make(map[string]interface{})
	updateInterface["account_status"] = 2
	dbConn.Table(db.Friend{}.TableName()).Where("friend_user_id = ?", userID).UpdateColumns(updateInterface)
	OperationID := "operation ID"
	friendLIst, _ := imdb.GetFriendListByUserID(userID)
	for _, v := range friendLIst {
		log.Info(OperationID, "UserInfoUpdatedNotification ", userID, v.FriendUserID)
		chat.UserInfoUpdatedNotification(OperationID, userID, v.FriendUserID)
	}

	chat.UserInfoUpdatedNotification(OperationID, userID, userID)
	log.Info(OperationID, "UserInfoUpdatedNotification ", userID, userID)
	return
}

func UpdateDeletedAccountInWooms(userId string) {
	// gorse
	utils3.DeleteUser(userId)

	// post short video data
	{
		shortVideo, err := imdb.GetAllShortVideoByUserId(userId)
		if err == nil && len(shortVideo) > 0 {
			// get hasn't deleted wooms
			var fileId []string
			for _, v := range shortVideo {
				if v.Status != constant.ShortVideoTypeDeleted {
					fileId = append(fileId, v.FileId)
				}
			}
			if len(fileId) > 0 {
				// delete in short video
				err = imdb.DeleteShortVideoByFileIdList(fileId)
				if err != nil {
					log.NewError("", "UpdateDeletedAccountInWooms", "DeleteShortVideoByFileIdList", fileId, err.Error())
				}

				// delete in redis cache
				for _, v := range fileId {
					db.DB.ShortVideoDeleteLikeListByFileId(v)
					db.DB.ShortVideoDeleteCommentCountByFileId(v)
				}

				// delete like
				{
					count := imdb.GetLikeCountByFileId(fileId)
					if count > 0 {
						var showNumber int64 = 1000
						times := count / showNumber
						var i int64
						var wg sync.WaitGroup
						for i = 0; i < times; i++ {
							wg.Add(1)
							start := i + 1
							go func() {
								defer wg.Done()
								likeList, err := imdb.GetLikeListByFileId(fileId, start, showNumber)
								if err != nil {
									log.NewError("", "UpdateDeletedAccountInWooms", "GetLikeListByFileId", fileId, err.Error())
									return
								}

								likeIdList := make([]int64, 0)
								likeUserDataList := make(map[string]int64, 0)

								for _, v := range likeList {
									likeIdList = append(likeIdList, v.Id)
									if _, ok := likeUserDataList[v.UserId]; !ok {
										likeUserDataList[v.UserId] = 1
									} else {
										likeUserDataList[v.UserId] += 1
									}
								}

								// Clean like in mysql
								imdb.MultiDeleteLikeByLikeIdList(likeIdList)

								// update count data
								for k, v := range likeUserDataList {
									imdb.IncrShortVideoUserCountByUserId(k, "like_num", -v)
								}
							}()
						}
						wg.Wait()
					}
				}

				// delete comment
				{
					count := imdb.GetCommentCountByFileId(fileId)
					if count > 0 {
						var showNumber int64 = 1000
						times := count / showNumber
						var i int64
						var wg sync.WaitGroup
						for i = 0; i < times; i++ {
							wg.Add(1)
							start := i + 1
							go func() {
								defer wg.Done()
								commentList, err := imdb.GetCommentListByFileId(fileId, start, showNumber)
								if err != nil {
									log.NewError("", "UpdateDeletedAccountInWooms", "GetCommentListByFileId", fileId, err.Error())
									return
								}

								commentIdList := make([]int64, 0)
								commentUserDataList := make(map[string]int64, 0)

								for _, v := range commentList {
									if v.Status != constant.ShortVideoCommentStatusDeleted {
										commentIdList = append(commentIdList, v.CommentId)
										if _, ok := commentUserDataList[v.UserID]; !ok {
											commentUserDataList[v.UserID] = 1
										} else {
											commentUserDataList[v.UserID] += 1
										}
									}
								}

								// Clean comment in mysql
								imdb.MultiDeleteCommentByCommentIdList(commentIdList)

								// update count data
								for k, v := range commentUserDataList {
									imdb.IncrShortVideoUserCountByUserId(k, "comment_num", -v)
								}
							}()
						}
						wg.Wait()
					}

				}

				// delete comment like
				{
					count := imdb.GetCommentLikeCountByFileId(fileId)
					if count > 0 {
						var showNumber int64 = 1000
						times := count / showNumber
						var i int64
						var wg sync.WaitGroup
						for i = 0; i < times; i++ {
							wg.Add(1)
							start := i + 1
							go func() {
								defer wg.Done()
								likeList, err := imdb.GetCommentLikeListByFileId(fileId, start, showNumber)
								if err != nil {
									log.NewError("", "UpdateDeletedAccountInWooms", "GetCommentLikeListByFileId", fileId, err.Error())
									return
								}

								likeIdList := make([]int64, 0)
								likeUserDataList := make(map[string]int64, 0)
								for _, v := range likeList {
									likeIdList = append(likeIdList, v.Id)
									if _, ok := likeUserDataList[v.UserId]; !ok {
										likeUserDataList[v.UserId] = 1
									} else {
										likeUserDataList[v.UserId] += 1
									}
								}

								// Clean like in mysql
								_ = imdb.MultiDeleteCommentLikeByLikeIdList(likeIdList)

								// update count data
								for k, v := range likeUserDataList {
									imdb.IncrShortVideoUserCountByUserId(k, "comment_like_num", -v)
								}
							}()
						}
					}
				}

				// delete file in gorse
				go func() {
					for _, v := range fileId {
						utils3.UpdateItem(v, true, constant.ShortString, "", strings.Split("", ","))
						utils3.DeleteItem(v)
					}
				}()
			}
		}
	}

	// like short video data
	{
		count := imdb.GetShortVideoLikeCountByUserId(userId)
		if count > 0 {
			var showNumber int64 = 1000
			times := count / showNumber
			var i int64
			var wg sync.WaitGroup
			for i = 0; i < times; i++ {
				wg.Add(1)
				start := i + 1
				go func() {
					defer wg.Done()
					likeList, err := imdb.GetLikeListByUserId(userId, start, showNumber)
					if err != nil {
						log.NewError("", "UpdateDeletedAccountInWooms", "GetLikeListByUserId", userId, err.Error())
						return
					}

					likeIdList := make([]int64, 0)
					fileIdList := make([]string, 0)

					for _, v := range likeList {
						likeIdList = append(likeIdList, v.Id)
						fileIdList = append(fileIdList, v.FileId)
					}

					// cleat like in mysql
					imdb.MultiDeleteLikeByLikeIdList(likeIdList)

					// update short video count data
					for _, v := range fileIdList {
						imdb.IncrShortVideoCountByFileId(v, "like_num", -1)
					}

					// update creator user count data
					likeShortVideoInfoList, err := imdb.GetShortVideoByFileIdList(fileIdList)
					if err != nil {
						log.NewError("", "UpdateDeletedAccountInWooms", "GetShortVideoByFileIdList", fileIdList, err.Error())
						return
					}

					createUserIdList := make(map[string]int64, 0)
					for _, v := range likeShortVideoInfoList {
						if _, ok := createUserIdList[v.UserId]; !ok {
							createUserIdList[v.UserId] = 1
						} else {
							createUserIdList[v.UserId] += 1
						}
					}

					for k, v := range createUserIdList {
						imdb.IncrShortVideoUserCountByUserId(k, "harvested_likes_number", -v)
					}
				}()
			}
			wg.Wait()
		}
	}

	return
}

func UpdateDeletedAccountInGroupMembers(userID string) {
	dbConn, err := db.DB.MysqlDB.DefaultGormDB()
	if err != nil {
		return
	}
	groups, err := imdb.GetGroupMemberListByUserID(userID)
	for _, groupMemberInfo := range groups {
		m := make(map[string]interface{})
		updatedVersion, _ := imdb.GetGroupUpdatesVersionNumberGroupID(groupMemberInfo.GroupID)
		m["update_version"] = updatedVersion.VersionNumber + 1
		m["account_status"] = 2
		m["nickname"] = "Deleted Account"
		m["user_group_face_url"] = ""
		dbConn.Table("group_members").Where("user_id=?", userID).Updates(m)
		//Update Group Version Update - We check that for sync process
		go imdb.UpdateGroupUpdatesVersionNumber(groupMemberInfo.GroupID)
		chat.GroupMemberInfoSetNotification("OperationID", userID, groupMemberInfo.GroupID, userID)
	}

}

func UpdateDeletedAccountInOfficialAccountAndArticle(userID string) {
	user, err := imdb.GetUserByUserIDAll(userID)
	if err != nil {
		log.NewError("", "get user by user id error, ", err.Error())
		return
	}

	if user.OfficialID == 0 {
		log.Debug("", "user don't have official account")
		return
	}

	count := imdb.DeleteOfficialAccounts([]string{strconv.FormatInt(user.OfficialID, 10)}, userID)
	if count == 0 {
		log.NewError("", "delete official account failed")
	}

	err = db.DB.DeleteAllArticlesByOfficialId(user.OfficialID, userID)
	if err != nil {
		log.NewError("", "delete article by official id from mongodb failed, err: ", err.Error())
	}

	err = imdb.DeleteAllArticlesByOfficialId(user.OfficialID, userID)
	if err != nil {
		log.NewError("", "delete article by official id failed, err: ", err.Error())
	}
}

func UpdateDeletedAccountInLikes(userID string) {
	count := imdb.DeleteArticleLikesByUserID([]string{userID}, userID)
	if count == 0 {
		log.NewError("", "delete article like failed")
	}
	err := db.DB.DeleteArticleLikeByUserID(userID)
	if err != nil {
		log.NewError("", "delete article like from mongodb failed")
	}

	count = imdb.DeleteArticleCommentLikesByUserID([]string{userID}, userID)
	if count == 0 {
		log.NewError("", "delete article comment like failed")
	}

	err = db.DB.DeleteArticleCommentLikeByUserID(userID, userID)
	if err != nil {
		log.NewError("", "delete article comment like from mongodb failed")
	}
}

func (s *userServer) GetAllUserID(_ context.Context, req *pbUser.GetAllUserIDReq) (*pbUser.GetAllUserIDResp, error) {
	log.NewInfo(req.OperationID, "GetAllUserID args ", req.String())
	if !token_verify.IsManagerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "IsManagerUserID false ", req.OpUserID)
		return &pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	uidList, err := imdb.SelectAllUserID()
	if err != nil {
		log.NewError(req.OperationID, "SelectAllUserID false ", err.Error())
		return &pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	} else {
		log.NewInfo(req.OperationID, "GetAllUserID rpc return ", pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{}, UserIDList: uidList})
		return &pbUser.GetAllUserIDResp{CommonResp: &pbUser.CommonResp{}, UserIDList: uidList}, nil
	}
}

func (s *userServer) AccountCheck(_ context.Context, req *pbUser.AccountCheckReq) (*pbUser.AccountCheckResp, error) {
	log.NewInfo(req.OperationID, "AccountCheck args ", req.String())
	if !token_verify.IsManagerUserID(req.OpUserID) {
		log.NewError(req.OperationID, "IsManagerUserID false ", req.OpUserID)
		return &pbUser.AccountCheckResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	uidList, err := imdb.SelectSomeUserID(req.CheckUserIDList)
	log.NewDebug(req.OperationID, "from db uid list is:", uidList)
	if err != nil {
		log.NewError(req.OperationID, "SelectSomeUserID failed ", err.Error(), req.CheckUserIDList)
		return &pbUser.AccountCheckResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	} else {
		var r []*pbUser.AccountCheckResp_SingleUserStatus
		for _, v := range req.CheckUserIDList {
			temp := new(pbUser.AccountCheckResp_SingleUserStatus)
			temp.UserID = v
			if utils.IsContain(v, uidList) {
				temp.AccountStatus = constant.Registered
			} else {
				temp.AccountStatus = constant.UnRegistered
			}
			r = append(r, temp)
		}
		resp := pbUser.AccountCheckResp{CommonResp: &pbUser.CommonResp{ErrCode: 0, ErrMsg: ""}, ResultList: r}
		log.NewInfo(req.OperationID, "AccountCheck rpc return ", resp.String())
		return &resp, nil
	}

}

func (s *userServer) UpdateUserInfo(ctx context.Context, req *pbUser.UpdateUserInfoReq) (*pbUser.UpdateUserInfoResp, error) {
	log.NewInfo(req.OperationID, "UpdateUserInfo args ", req.String())
	var user db.User
	utils.CopyStructFields(&user, req.UserInfo)
	if req.UserInfo.Birth != "" {
		birth, err := utils.TimeStringToTime(req.UserInfo.Birth)
		if err != nil {
			log.NewError(req.OperationID, "check the birth format", req.UserInfo.Birth)
			return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "check the birth format"}}, nil
		}
		user.Birth = birth
	}

	// check the phone and email if used in the other account
	if user.PhoneNumber != "" {
		oldUser, _ := imdb.GetRegisterFromPhone(user.PhoneNumber)
		if oldUser != nil && oldUser.UserID != "" && oldUser.UserID != user.UserID {
			log.NewError(req.OperationID, "PhoneNumber was used by the other person ", req.OpUserID, req.UserInfo.UserID)
			return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
	} else {
		userStruct, _ := imdb.GetUserByUserID(req.UserInfo.UserID)
		if userStruct != nil {
			user.PhoneNumber = userStruct.PhoneNumber
		}
	}

	if user.Email != "" {
		oldUser, _ := imdb.GetRegisterFromEmail(user.Email)
		if oldUser != nil && oldUser.UserID != "" && oldUser.UserID != user.UserID {
			log.NewError(req.OperationID, "Email was used by the other person ", req.OpUserID, req.UserInfo.UserID)
			return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
		}
	}

	// persistence the face url
	if user.FaceURL != "" {
		client, err := admin_cms.GetTencentCloudClient(true)
		if err == nil {
			errList, _ := admin_cms.RemoveDeleteTagForPersistent(client, []string{user.FaceURL})
			if len(errList) > 0 {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", errList)
			}
		}
	}

	err := imdb.UpdateUserInfo(user)

	if err != nil {
		log.NewError(req.OperationID, "UpdateUserInfo failed ", err.Error(), user)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}

	client := pbFriend.NewFriendClient(etcdConn)
	newReq := &pbFriend.GetFriendListReq{
		CommID: &pbFriend.CommID{OperationID: req.OperationID, FromUserID: req.UserInfo.UserID, OpUserID: req.OpUserID},
	}

	RpcResp, err := client.GetFriendList(context.Background(), newReq)
	if err != nil {
		log.NewError(req.OperationID, "GetFriendList failed ", err.Error(), newReq)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{}}, nil
	}
	for _, v := range RpcResp.FriendInfoList {
		log.Info(req.OperationID, "UserInfoUpdatedNotification ", req.UserInfo.UserID, v.FriendUser.UserID)
		chat.UserInfoUpdatedNotification(req.OperationID, req.UserInfo.UserID, v.FriendUser.UserID)
	}

	chat.UserInfoUpdatedNotification(req.OperationID, req.UserInfo.UserID, req.OpUserID)
	log.Info(req.OperationID, "UserInfoUpdatedNotification ", req.UserInfo.UserID, req.OpUserID)
	if req.UserInfo.FaceURL != "" || req.UserInfo.Nickname != "" {
		// sync local data for group members
		go s.SyncJoinedGroupMemberFaceURL(req.UserInfo.UserID, req.UserInfo.FaceURL, req.UserInfo.Nickname, req.OperationID, req.OpUserID)

		// sync update moments, moments comments, moments like face url.
		go s.syncMomentsFaceURL(req.UserInfo.UserID, req.UserInfo.FaceURL, req.UserInfo.Nickname, req.OperationID, req.OpUserID)
	}

	//log.NewInfo(req.OperationID, "SyncToLocalDataBase user", req.UserInfo.Nickname, req.UserInfo.UserID)
	//etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.OperationID)
	//if etcdConn == nil {
	//	log.Error(req.OperationID, "OpenImLocalDataName rpc connect failed ")
	//	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: "OpenImLocalDataName rpc connect failed "}}, nil
	//}
	//client2 := local_database.NewLocalDataBaseClient(etcdConn)
	//reqPb := local_database.SyncUserInfoReq{
	//	OperationID: req.OperationID,
	//	UserID:      req.UserInfo.UserID,
	//}
	//_, err2 := client2.SyncUserInfoToLocal(context.Background(), &reqPb)
	//if err2 != nil {
	//	log.Error(req.OperationID, "SyncUserInfoToLocal failed ", err2.Error(), req.UserInfo.UserID)
	//	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: "SyncUserInfoToLocal failed "}}, nil
	//}

	// updateUserInfoToCacheReq := &cache.UpdateUserInfoToCacheReq{
	//	OperationID:  req.OperationID,
	//	UserInfoList: []*sdkws.UserInfo{req.UserInfo},
	// }
	// cacheEtcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName)
	// cacheClient := cache.NewCacheClient(cacheEtcdConn)
	// resp, err := cacheClient.UpdateUserInfoToCache(context.Background(), updateUserInfoToCacheReq)
	// if err != nil {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error(), updateUserInfoToCacheReq.String())
	//	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: err.Error()}}, nil
	// }
	// if resp.CommonResp.ErrCode != 0 {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), resp.String())
	//	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: resp.CommonResp.ErrMsg}}, nil
	// }
	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{}}, nil
}

func (s *userServer) RemoveUserFaceUrl(ctx context.Context, req *pbUser.UpdateUserInfoReq) (*pbUser.UpdateUserInfoResp, error) {
	log.NewInfo(req.OperationID, "RemoveUserFaceUrl args ", req.String())
	var user db.User
	utils.CopyStructFields(&user, req.UserInfo)
	err := imdb.RemoveUserFaceUrl(user)

	if err != nil {
		log.NewError(req.OperationID, "RemoveUserFaceUrl failed ", err.Error(), user)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}

	client := pbFriend.NewFriendClient(etcdConn)
	newReq := &pbFriend.GetFriendListReq{
		CommID: &pbFriend.CommID{OperationID: req.OperationID, FromUserID: req.UserInfo.UserID, OpUserID: req.OpUserID},
	}

	RpcResp, err := client.GetFriendList(context.Background(), newReq)
	if err != nil {
		log.NewError(req.OperationID, "GetFriendList failed ", err.Error(), newReq)
		return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{}}, nil
	}
	for _, v := range RpcResp.FriendInfoList {
		log.Info(req.OperationID, "UserInfoUpdatedNotification ", req.UserInfo.UserID, v.FriendUser.UserID)
		chat.UserInfoUpdatedNotification(req.OperationID, req.UserInfo.UserID, v.FriendUser.UserID)
	}

	chat.UserInfoUpdatedNotification(req.OperationID, req.UserInfo.UserID, req.OpUserID)
	log.Info(req.OperationID, "UserInfoUpdatedNotification ", req.UserInfo.UserID, req.OpUserID)
	if req.UserInfo.FaceURL != "" || req.UserInfo.Nickname != "" {
		// sync local data for group members
		go s.SyncJoinedGroupMemberFaceURL(req.UserInfo.UserID, req.UserInfo.FaceURL, req.UserInfo.Nickname, req.OperationID, req.OpUserID)

		// sync update moments, moments comments, moments like face url.
		go s.syncMomentsFaceURL(req.UserInfo.UserID, req.UserInfo.FaceURL, req.UserInfo.Nickname, req.OperationID, req.OpUserID)
	}
	return &pbUser.UpdateUserInfoResp{CommonResp: &pbUser.CommonResp{}}, nil
}
func (s *userServer) SetGlobalRecvMessageOpt(ctx context.Context, req *pbUser.SetGlobalRecvMessageOptReq) (*pbUser.SetGlobalRecvMessageOptResp, error) {
	log.NewInfo(req.OperationID, "SetGlobalRecvMessageOpt args ", req.String())

	var user db.User
	user.UserID = req.UserID
	m := make(map[string]interface{}, 1)

	log.NewDebug(req.OperationID, utils.GetSelfFuncName(), req.GlobalRecvMsgOpt, "set GlobalRecvMsgOpt")
	m["global_recv_msg_opt"] = req.GlobalRecvMsgOpt
	err := db.DB.SetUserGlobalMsgRecvOpt(user.UserID, req.GlobalRecvMsgOpt)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetGlobalRecvMessageOpt failed ", err.Error(), user)
		return &pbUser.SetGlobalRecvMessageOptResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	err = imdb.UpdateUserInfoByMap(user, m)
	if err != nil {
		log.NewError(req.OperationID, "SetGlobalRecvMessageOpt failed ", err.Error(), user)
		return &pbUser.SetGlobalRecvMessageOptResp{CommonResp: &pbUser.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	chat.UserInfoUpdatedNotification(req.OperationID, req.UserID, req.UserID)
	return &pbUser.SetGlobalRecvMessageOptResp{CommonResp: &pbUser.CommonResp{}}, nil
}
func (s *userServer) SyncJoinedGroupMemberFaceURL(userID, faceURL, nickName, operationID, opUserID string) {
	joinedGroupIDList, err := imdb.GetJoinedGroupIDListByUserID(userID)
	if err != nil {
		log.NewWarn(operationID, "GetJoinedGroupIDListByUserID failed ", userID, err.Error())
		return
	}
	for _, v := range joinedGroupIDList {
		groupMemberInfo := db.GroupMember{UserID: userID, GroupID: v, FaceURL: faceURL, Nickname: nickName}
		imdb.UpdateGroupMemberInfo(groupMemberInfo)
		chat.GroupMemberInfoSetNotification(operationID, opUserID, v, userID)
	}
}

func (s *userServer) syncMomentsFaceURL(userID, faceURL, nickName, operationID, opUserID string) {
	nowTime := time.Now().Unix()

	// update mongo db
	err := db.DB.UpdateMomentsUserInfo(db.Moment{UserID: userID, UserName: nickName, UserProfileImg: faceURL, MUpdateTime: nowTime})
	if err != nil {
		log.NewWarn(operationID, "Update moments on mongo db failed", userID, err.Error())
	}
	err = db.DB.UpdateMomentCommentUserInfo(db.MomentComment{UserID: userID, UserName: nickName, UserProfileImg: faceURL, UpdatedTime: nowTime, UpdateBy: opUserID})
	if err != nil {
		log.NewWarn(operationID, "Update moments on mongo db failed", userID, err.Error())
	}
	err = db.DB.UpdateMomentLikeUserInfo(db.MomentLike{UserID: userID, UserName: nickName, UserProfileImg: faceURL, UpdatedTime: nowTime, UpdateBy: opUserID})
	if err != nil {
		log.NewWarn(operationID, "Update moments on mongo db failed", userID, err.Error())
	}

	// update to db
	moment := &db.MomentSQL{
		UserID:         userID,
		UserName:       nickName,
		UserProfileImg: faceURL,
		MUpdateTime:    nowTime,
	}
	_, err = imdb.UpdateMomentByUserId(moment)
	if err != nil {
		log.NewWarn(operationID, "Update moments failed ", userID, err.Error())
	}

	comment := &db.MomentCommentSQL{
		UserID:         userID,
		UserName:       nickName,
		UserProfileImg: faceURL,
		UpdateBy:       opUserID,
		UpdatedTime:    nowTime,
	}
	err = imdb.UpdateMomentComment(comment)
	if err != nil {
		log.NewWarn(operationID, "Update moments failed ", userID, err.Error())
	}

	like := &db.MomentLikeSQL{
		UserID:         userID,
		UserName:       nickName,
		UserProfileImg: faceURL,
		UpdateBy:       opUserID,
		UpdatedTime:    nowTime,
	}
	err = imdb.UpdateMomentLikes(like)
	if err != nil {
		log.NewWarn(operationID, "Update moments failed ", userID, err.Error())
	}

}

func (s *userServer) GetUsersByName(ctx context.Context, req *pbUser.GetUsersByNameReq) (*pbUser.GetUsersByNameResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbUser.GetUsersByNameResp{}
	users, err := imdb.GetUserByName(req.UserName, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserByName failed", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	for _, user := range users {
		isBlock, err := imdb.UserIsBlock(user.UserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
			continue
		}
		resp.Users = append(resp.Users, &pbUser.User{
			ProfilePhoto: user.FaceURL,
			Nickname:     user.Nickname,
			UserId:       user.UserID,
			CreateTime:   user.CreateTime,
			IsBlock:      isBlock,
			Status:       int32(user.Status),
		})
	}
	user := db.User{Nickname: req.UserName}
	userNums, err := imdb.GetUsersCount(user)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	resp.UserNums = int32(userNums)
	resp.Pagination = &sdkws.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	return resp, nil
}

func (s *userServer) GetUserById(ctx context.Context, req *pbUser.GetUserByIdReq) (*pbUser.GetUserByIdResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbUser.GetUserByIdResp{User: &pbUser.User{}}
	user, err := imdb.GetUserByUserID(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	isBlock, err := imdb.UserIsBlock(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "req：", req.String())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	resp.User = &pbUser.User{
		ProfilePhoto: user.FaceURL,
		Nickname:     user.Nickname,
		UserId:       user.UserID,
		CreateTime:   user.CreateTime,
		IsBlock:      isBlock,
		Status:       int32(user.Status),
		PhoneNumber:  user.PhoneNumber,
		Email:        user.Email,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *userServer) GetUsers(_ context.Context, req *pbUser.GetUsersRequest) (*pbUser.GetUsersResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetUsersResp{User: []*pbUser.User{}, CommonResp: &pbUser.CommonResp{}}

	where := map[string]string{}
	where["source_id"] = req.SourceId
	where["source_code"] = req.SourceCode
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["user_id"] = req.UserId
	where["type"] = req.Type
	where["remark"] = req.Remark
	where["last_login_device"] = strconv.Itoa(int(req.LastLoginDevice))
	where["gender"] = strconv.Itoa(int(req.Gender))

	var statusList []string
	if req.AccountStatus != "" {
		err := json.Unmarshal([]byte(req.AccountStatus), &statusList)
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg
			return resp, err
		}
	}
	users, userCount, err := imdb.GetUsersByWhere(where, statusList, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	//absPath, _ := filepath.Abs("../IP2LOCATION-LITE-DB5.BIN")
	//ipDb, ipErr := ip2location.OpenDB(absPath)
	//defer ipDb.Close()

	for _, v := range users {
		isBlock, err := imdb.UserIsBlock(v.UserID)
		address := ""
		if v.UpdateIp != "" {
			//if ipErr == nil {
			//	addressResults, err1 := ipDb.Get_all(v.UpdateIp)
			//	if err1 == nil {
			//		address = addressResults.City
			//	}
			//}
		}

		if err == nil {
			user := &pbUser.User{
				ProfilePhoto:    v.FaceURL,
				UserId:          v.UserID,
				CreateTime:      v.CreateTime,
				Nickname:        v.Nickname,
				LastLoginTime:   v.LastLoginTime,
				SourceId:        strconv.Itoa(int(v.SourceId)),
				SourceCode:      v.SourceCode,
				IsBlock:         isBlock,
				PhoneNumber:     v.PhoneNumber,
				Email:           v.Email,
				Status:          int32(v.Status),
				UpdateIp:        v.UpdateIp,
				Address:         address,
				Uuid:            v.Uuid,
				LastLoginDevice: int32(v.LastLoginDevice),
				VideoStatus:     int32(v.VideoStatus),
				AudioStatus:     int32(v.AudioStatus),
				Gender:          v.Gender,
				Remark:          v.Remark,
				LoginIp:         v.LoginIp,
			}
			resp.User = append(resp.User, user)
		} else {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "UserIsBlock failed", err.Error())
		}
	}

	resp.UserNums = int32(userCount)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *userServer) ResignUser(ctx context.Context, req *pbUser.ResignUserReq) (*pbUser.ResignUserResp, error) {
	log.NewInfo(req.OperationID, "ResignUser args ", req.String())
	return &pbUser.ResignUserResp{}, nil
}

func (s *userServer) AlterUser(ctx context.Context, req *pbUser.AlterUserReq) (*pbUser.AlterUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.AlterUserResp{CommonResp: &pbUser.CommonResp{}}
	// check phone number,only one user can use this phone number
	if req.PhoneNumber != "" {
		oldUser, err := imdb.GetUserByPhoneNumber(req.PhoneNumber)
		if err == nil {
			if oldUser.UserID != req.UserId {
				return resp, errors2.WrapError(constant.ErrUserExistPhoneNumberExist)
			}
		}
	}

	oldUser, err := imdb.GetUserByUserID(req.UserId)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, err
	}
	if oldUser.FaceURL != req.FaceURL {
		client, err := admin_cms.GetTencentCloudClient(true)
		if err != nil {
			return resp, err
		}
		persistent, err := admin_cms.RemoveDeleteTagForPersistent(client, []string{req.FaceURL})
		if err != nil {
			return resp, err
		}
		if len(persistent) != 0 {
			resp.CommonResp.ErrMsg = "Upload face url failed"
			return resp, nil
		}
	}

	password := ""
	if req.Password != "" {
		// get user info
		user, _ := imdb.GetUserByUserID(req.UserId)
		newPasswordFirst := req.Password + user.Salt
		passwordData := []byte(newPasswordFirst)
		has := md5.Sum(passwordData)
		password = fmt.Sprintf("%x", has)
	}

	user := db.User{
		PhoneNumber: req.PhoneNumber,
		Nickname:    req.Nickname,
		UserID:      req.UserId,
		Password:    password,
		Email:       req.Email,
		Gender:      req.Gender,
		Remark:      req.Remark,
		SourceId:    utils.StringToInt64(req.SourceId),
		SourceCode:  req.SourceCode,
		FaceURL:     req.FaceURL,
	}
	if oldUser.FaceURL != req.FaceURL && req.FaceURL != "" {
		client, err := admin_cms.GetTencentCloudClient(true)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
			return resp, err
		}
		errorList, err := admin_cms.RemoveDeleteTagForPersistent(client, []string{req.FaceURL})
		if err != nil {
			return nil, err
		}
		if len(errorList) > 0 {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser")
			return resp, errors.New("add user photo profile failed")
		}
	}
	if err := imdb.UpdateUserInfo(user); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateUserInfo", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}

	if user.Password != "" {
		go func() {
			reqLogout := &rpcAuth.ForceLogoutReq{}
			reqLogout.FromUserID = user.UserID
			reqLogout.OpUserID = req.OpUserId
			reqLogout.OperationID = req.OperationID

			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAuthName, req.OperationID)
			if etcdConn != nil {
				client := rpcAuth.NewAuthClient(etcdConn)
				_, err := client.ForceLogout(context.Background(), reqLogout)
				if err != nil {
					errMsg := req.OperationID + " forceKickOff failed " + err.Error() + user.UserID
					log.NewError(req.OperationID, errMsg)
				}
			}
		}()

	}

	// source manager,can delete user phone number
	if req.PhoneNumber == "" {
		imdb.DeleteUserPhoneNumber(req.UserId)
	}

	//data synchronization
	syncUserToLocal(req.OperationID, req.UserId)

	chat.UserInfoUpdatedNotification(req.OperationID, req.UserId, req.OpUserId)

	var interests []int64
	for _, interest := range req.Interests {
		interest, _ := strconv.ParseInt(interest, 10, 64)
		interests = append(interests, interest)
	}
	if len(interests) > 0 {
		imdb.AlterUserInterests(req.UserId, interests)
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *userServer) AddUser(_ context.Context, req *pbUser.AddUserReq) (*pbUser.AddUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.AddUserResp{CommonResp: &pbUser.CommonResp{}}

	err := imdb.AddUser(req.UserId, req.PhoneNumber, req.Name, req.Password, req.OpUserId, req.SourceId, req.Code, req.Remark, req.Gender, req.Email, req.FaceURL)
	if err != nil {
		if err.Error() == "105" {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
			resp.CommonResp.ErrCode = constant.ErrUserPhoneAlreadyExsist.ErrCode
			resp.CommonResp.ErrMsg = "Error:" + constant.ErrUserPhoneAlreadyExsist.ErrMsg
			return resp, errors2.WrapError(constant.ErrUserPhoneAlreadyExsist)
		}
		if err.Error() == "106" {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
			resp.CommonResp.ErrCode = constant.ErrUserEmailAlreadyExist.ErrCode
			resp.CommonResp.ErrMsg = "Error:" + constant.ErrUserEmailAlreadyExist.ErrMsg
			return resp, errors2.WrapError(constant.ErrUserEmailAlreadyExist)
		}
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
		resp.CommonResp.ErrCode = constant.ErrUserIDAlreadyExsist.ErrCode
		resp.CommonResp.ErrMsg = "Error:" + constant.ErrUserIDAlreadyExsist.ErrMsg
		return resp, errors2.WrapError(constant.ErrUserIDAlreadyExsist)
	}

	if req.FaceURL != "" {
		client, err := admin_cms.GetTencentCloudClient(true)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
			resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
			resp.CommonResp.ErrMsg = "Error:Upload face photo failed"
			return resp, err
		}
		errorList, err := admin_cms.RemoveDeleteTagForPersistent(client, []string{req.FaceURL})
		if err != nil {
			return nil, err
		}
		if len(errorList) > 0 {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser")
			resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
			resp.CommonResp.ErrMsg = "Error:Upload face photo failed"
			return resp, errors.New("add user photo profile failed")
		}
	}

	imdb.SetUserInterestType(req.UserId, req.Interests)

	if req.Code != "" {
		if config.Config.Invite.IsOpen == 1 && imdb.GetInviteCodeIsOpen() {
			inviteCodeInfo := imdb.GetCodeInfoByCode(req.Code)
			if inviteCodeInfo != nil && inviteCodeInfo.State == constant.InviteCodeStateValid {
				go func() {
					// add friend
					friendUserId := inviteCodeInfo.UserId

					etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, req.OperationID)
					if etcdConn == nil {
						errMsg := req.OperationID + "getcdv3.GetConn == nil"
						log.NewError(req.OperationID, errMsg)
						return
					}
					client := rpc.NewFriendClient(etcdConn)
					friend, err := client.AutoAddFriend(context.Background(), &rpc.AutoAddFriendRequset{
						OperationID:      req.OperationID,
						FromUserID:       friendUserId,
						FriendUserIDList: []string{req.UserId},
						Greeting:         inviteCodeInfo.Greeting,
					})
					log.NewInfo(req.OperationID, "AutoAddFriend", friend, err)
				}()
			}
		}
		if config.Config.Channel.IsOpen == 1 && imdb.GetChannelCodeIsOpen() {
			channelInfo, _ := imdb.GetInviteChannelCodeByCode(req.Code)
			if channelInfo != nil {
				if channelInfo.State == constant.InviteChannelCodeStateValid {
					go func() {
						// add friend
						if channelInfo.FriendId != "" {
							etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, req.OperationID)
							if etcdConn == nil {
								errMsg := req.OperationID + "getcdv3.GetConn == nil"
								log.NewError(req.OperationID, errMsg)
								return
							}
							client := rpc.NewFriendClient(etcdConn)
							friend, err := client.ChannelAddFriend(context.Background(), &rpc.ChannelAddFriendRequset{
								OperationID:      req.OperationID,
								FromUserID:       req.UserId,
								FriendUserIDList: strings.Split(channelInfo.FriendId, ","),
								Greeting:         channelInfo.Greeting,
							})
							log.NewInfo(req.OperationID, "AutoAddFriend", friend, err)
						}

						// add group
						if channelInfo.GroupId != "" {
							etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImGroupName, req.OperationID)
							if etcdConn == nil {
								errMsg := req.OperationID + "getcdv3.GetConn == nil"
								log.NewError(req.OperationID, errMsg)
								return
							}
							client := groupRpc.NewGroupClient(etcdConn)
							group, err := client.ChannelCodeInviteUserToGroup(context.Background(), &groupRpc.ChannelCodeInviteUserToGroupRequest{
								OperationID:        req.OperationID,
								UserId:             req.UserId,
								InvitedGroupIDList: strings.Split(channelInfo.GroupId, ","),
								Reason:             "",
							})
							log.NewInfo(req.OperationID, "AutoAddGroup", group, err)
						}
					}()
				}
			}
		}
	}

	return resp, nil
}

func (s *userServer) BlockUser(ctx context.Context, req *pbUser.BlockUserReq) (*pbUser.BlockUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.BlockUserResp{}
	err := imdb.BlockUser(req.UserId, req.EndDisableTime)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BlockUser", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *userServer) UnBlockUser(ctx context.Context, req *pbUser.UnBlockUserReq) (*pbUser.UnBlockUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.UnBlockUserResp{}
	err := imdb.UnBlockUser(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "unBlockUser", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *userServer) GetBlockUsers(ctx context.Context, req *pbUser.GetBlockUsersReq) (*pbUser.GetBlockUsersResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetBlockUsersResp{}
	blockUsers, err := imdb.GetBlockUsers(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.Error(req.OperationID, utils.GetSelfFuncName(), "GetBlockUsers", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	for _, v := range blockUsers {
		resp.BlockUsers = append(resp.BlockUsers, &pbUser.BlockUser{
			User: &pbUser.User{
				ProfilePhoto: v.User.FaceURL,
				Nickname:     v.User.Nickname,
				UserId:       v.User.UserID,
				IsBlock:      true,
			},
			BeginDisableTime: (v.BeginDisableTime).String(),
			EndDisableTime:   (v.EndDisableTime).String(),
		})
	}
	resp.Pagination = &sdkws.ResponsePagination{}
	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	resp.Pagination.CurrentPage = req.Pagination.PageNumber
	nums, err := imdb.GetBlockUsersNumCount()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetBlockUsersNumCount failed", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	resp.UserNums = int32(nums)
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	return resp, nil
}

func (s *userServer) GetBlockUserById(_ context.Context, req *pbUser.GetBlockUserByIdReq) (*pbUser.GetBlockUserByIdResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbUser.GetBlockUserByIdResp{}
	user, err := imdb.GetBlockUserById(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetBlockUserById", err)
		return resp, errors2.WrapError(constant.ErrDB)
	}
	resp.BlockUser = &pbUser.BlockUser{
		User: &pbUser.User{
			ProfilePhoto: user.User.FaceURL,
			Nickname:     user.User.Nickname,
			UserId:       user.User.UserID,
			IsBlock:      true,
		},
		BeginDisableTime: (user.BeginDisableTime).String(),
		EndDisableTime:   (user.EndDisableTime).String(),
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", req.String())
	return resp, nil
}

func (s *userServer) DeleteUser(_ context.Context, req *pbUser.DeleteUserReq) (*pbUser.DeleteUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbUser.DeleteUserResp{}
	if row := imdb.DeleteUser(req.UserId, req.Reason, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, errors2.WrapError(constant.ErrDB)
	} else {
		go UpdateDeletedAccountInFriends(req.UserId)
		go UpdateDeletedAccountInMoments(req.UserId)
		go UpdateDeletedAccountInGroupMembers(req.UserId)
		go UpdateDeletedAccountInOfficialAccountAndArticle(req.UserId)
		go UpdateDeletedAccountInLikes(req.UserId)

		// delete all short video
		go UpdateDeletedAccountInWooms(req.UserId)
	}

	err := imdb.DeleteUserInterests(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete user interest failed")
		return nil, err
	}

	if err := token_verify.DeleteAllToken(req.UserId); err != nil { //req.GAuthTypeToken)
		errMsg := req.OperationID + " DeleteToken failed " + err.Error() + req.UserId
		log.NewError(req.OperationID, errMsg)
	}

	if err := forceKickOff(req.UserId, 0, req.OperationID); err != nil {
		errMsg := req.OperationID + " forceKickOff failed " + err.Error() + req.UserId
		log.NewError(req.OperationID, errMsg)
	}

	return resp, nil
}

func (s *userServer) SwitchStatus(c context.Context, req *pbUser.SwitchStatusReq) (*pbUser.SwitchStatusResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())

	resp := &pbUser.SwitchStatusResp{}

	opUserID := req.OpUserId
	userID := req.UserId
	status := req.Status

	_, err := imdb.GetUserByUserID(userID)
	if err != nil {
		return resp, err
	}

	updateData := db.User{
		UserID:     userID,
		UpdateTime: time.Now().Unix(),
		UpdateUser: opUserID,
	}
	if req.StatusType == 1 {
		updateData.Status = int(status)
	} else if req.StatusType == 2 {
		updateData.VideoStatus = int8(status)
	} else if req.StatusType == 3 {
		updateData.AudioStatus = int8(status)
	}

	if err := imdb.UpdateUserStatus(updateData); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update user status failed!", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}
	if req.StatusType == 1 {
		err = db.DB.SaveUserStatus(userID, status)
		if err != nil {
			return resp, err
		}

		if status == 2 {
			if err := token_verify.DeleteAllToken(req.UserId); err != nil { //req.GAuthTypeToken)
				errMsg := req.OperationID + " DeleteToken failed " + err.Error() + req.UserId
				log.NewError(req.OperationID, errMsg)
			}

			if err := forceKickOff(req.UserId, 0, req.OperationID); err != nil {
				errMsg := req.OperationID + " forceKickOff failed " + err.Error() + req.UserId
				log.NewError(req.OperationID, errMsg)
			}
		}
	}

	//sync local user
	//log.NewInfo(req.OperationID, "SyncToLocalDataBase user", userID)
	//etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.OperationID)
	//if etcdConn == nil {
	//	log.Error(req.OperationID, "OpenImLocalDataName rpc connect failed ")
	//	return resp, errors.WrapError(constant.ErrRPC)
	//}
	//client := local_database.NewLocalDataBaseClient(etcdConn)
	//reqPb := local_database.SyncUserInfoReq{
	//	OperationID: req.OperationID,
	//	UserID:      userID,
	//}
	//_, err2 := client.SyncUserInfoToLocal(context.Background(), &reqPb)
	//if err2 != nil {
	//	log.Error(req.OperationID, "SyncUserInfoToLocal failed ", err2.Error(), userID)
	//	return resp, err2
	//}

	return resp, nil
}

func (s *userServer) GetDeletedUsers(_ context.Context, req *pbUser.GetDeletedUsersReq) (*pbUser.GetDeletedUsersResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbUser.GetDeletedUsersResp{Pagination: &sdkws.ResponsePagination{}}

	where := map[string]string{}
	where["user"] = req.User
	where["gender"] = fmt.Sprintf("%d", req.Gender)
	where["reason"] = req.Reason
	where["location"] = req.Location
	where["last_login_device"] = fmt.Sprintf("%d", req.LastLoginDevice)
	where["deleted_by"] = req.DeletedBy
	where["time_type"] = fmt.Sprintf("%d", req.TimeType)
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	users, userCount, err := imdb.GetDeletedUsersByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber)
	if err != nil {
		log.Error(req.OperationID, utils.GetSelfFuncName(), "GetBlockUsers", err.Error())
		return resp, errors2.WrapError(constant.ErrDB)
	}

	resp.DeletedUsers = make([]*pbUser.DeletedUserRes, len(users))
	for index, user := range users {
		resp.DeletedUsers[index] = &pbUser.DeletedUserRes{
			UserID:       user.UserID,
			Username:     user.Nickname,
			ProfilePhoto: user.FaceURL,
			LastLoginIP:  user.LoginIp,
			DeletedBy:    user.DeleteUser,
			DeleteTime:   user.DeleteTime,
			CreateTime:   user.CreateTime,
			Gender:       user.Gender,
			PhoneNumber:  user.PhoneNumber,
			Reason:       user.DeleteReason,
		}
	}
	resp.DeletedUsersCount = userCount
	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	resp.Pagination.CurrentPage = req.Pagination.PageNumber

	return resp, nil
}

// AlterAddFriendStatus
// func (s *userServer) AlterAddFriendStatus(c context.Context, req *pbUser.SwitchStatusReq) (*pbUser.SwitchStatusResp, error) {
// 	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())

// 	resp := &pbUser.SwitchStatusResp{}

// 	opUserID := req.OpUserId
// 	userID := req.UserId
// 	status := req.Status

// 	updateData := db.User{
// 		UserID:          userID,
// 		SuperUserStatus: int(status),
// 		UpdateTime:      time.Now().Unix(),
// 		UpdateUser:      opUserID,
// 	}

// 	if err := imdb.UpdateSuperUserStatus(updateData); err != nil {
// 		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update user status failed!", err.Error())
// 		return resp, errors2.WrapError(constant.ErrDB)
// 	}

// 	return resp, nil
// }

func syncConversationToLocal(operationID, ownerID, conversationID, msgType string, conversationType int32) error {
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)

	var userIDList []string
	if conversationType == constant.SingleChatType {
		//sync owner
		log.NewError(operationID, utils.GetSelfFuncName(), ownerID, conversationID)
		conversationInfo, err := imdb.GetConversation(ownerID, conversationID)
		if err != nil {
			log.NewError(operationID, "GetConversation failed")
			return err
		}

		localConversation := sdkws.Conversation{}
		err = utils.CopyStructFields(&localConversation, &conversationInfo)
		if err != nil {
			log.NewError(operationID, "CopyStructFields failed")
			return err
		}
		userIDList = []string{ownerID}

		syncConvReq := &local_database.SyncDataReq{
			OperationID:  operationID,
			MsgType:      msgType,
			MemberIDList: userIDList,
			Conversation: &localConversation,
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

		//sync receiver
		otherOwnerID := conversationInfo.UserID
		otherConvID := utils.GetConversationIDBySessionType(conversationInfo.OwnerUserID, constant.SingleChatType)
		log.NewError(operationID, utils.GetSelfFuncName(), otherOwnerID, otherConvID)
		otherConvInfo, err := imdb.GetConversation(otherOwnerID, otherConvID)
		if err != nil {
			log.NewError(operationID, "GetConversation failed")
			return err
		}

		otherlocalConv := sdkws.Conversation{}
		err = utils.CopyStructFields(&otherlocalConv, &otherConvInfo)
		if err != nil {
			log.NewError(operationID, "CopyStructFields failed")
			return err
		}
		userIDList = []string{otherOwnerID}

		syncConvReq = &local_database.SyncDataReq{
			OperationID:  operationID,
			MsgType:      msgType,
			MemberIDList: userIDList,
			Conversation: &otherlocalConv,
		}
		localConvResp, err = localDataClient.SyncData(context.Background(), syncConvReq)
		if err != nil {
			log.NewError(operationID, "SyncData rpc call failed", err.Error())
			return err
		}

		if localConvResp.ErrCode != 0 {
			log.NewError(operationID, "SyncData rpc logic call failed ", localConvResp.String())
			return errors.New("SyncData rpc logic call failed")
		}

	} else if conversationType == constant.GroupChatType {
		//sync for group members
		conversationInfo, err := imdb.GetConversation(ownerID, conversationID)
		if err != nil {
			log.NewError(operationID, "GetConversation failed")
			return err
		}
		userIDList, _ = imdb.GetGroupMemberIDListByGroupID(conversationInfo.GroupID)
		for _, userID := range userIDList {
			log.NewError(operationID, utils.GetSelfFuncName(), userID, conversationID)
			otherConversationInfo, err := imdb.GetConversation(userID, conversationID)
			if err != nil {
				log.NewError(operationID, "GetConversation failed")
				continue
			}
			localConversation := sdkws.Conversation{}
			err = utils.CopyStructFields(&localConversation, &otherConversationInfo)
			if err != nil {
				log.NewError(operationID, "CopyStructFields failed")
				continue
			}

			syncConvReq := &local_database.SyncDataReq{
				OperationID:  operationID,
				MsgType:      msgType,
				MemberIDList: []string{userID},
				Conversation: &localConversation,
			}
			localConvResp, err := localDataClient.SyncData(context.Background(), syncConvReq)
			if err != nil {
				log.NewError(operationID, "SyncData rpc call failed", err.Error())
				continue
			}

			if localConvResp.ErrCode != 0 {
				log.NewError(operationID, "SyncData rpc logic call failed ", localConvResp.String())
				continue
			}
		}

	}
	return nil
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

func (s *userServer) GenerateFriendAndGroupForChannel(c context.Context, req *pbUser.CommonReq) (*pbUser.CommonResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())

	resp := &pbUser.CommonResp{}

	opUserID := req.OpUserID
	user, err := imdb.GetUserByUserID(opUserID)
	if err != nil {
		log.NewError("", utils2.GetSelfFuncName(), "user is nil")
		resp.ErrCode = constant.ErrArgs.ErrCode
		resp.ErrMsg = "user is nil"
		return resp, err
	}

	//check if the channelcode and invitecode is opened
	allowChannel := false
	if imdb.GetChannelCodeIsOpen() {
		allowChannel = true
	}

	allowInvite := false
	if imdb.GetInviteCodeIsOpen() {
		allowInvite = true
	}

	//get channel code information
	//'1:official 2:invite 3:channel'
	sourceID := user.SourceId
	sourceCode := user.SourceCode
	friendList := ""
	groupList := ""
	var channelCode *db.InviteChannelCode
	var inviteCode *db.InviteCode
	var greeting string
	switch sourceID {
	case 1:
		if allowChannel {
			//include valid and invalid
			channelCode = imdb.GetOfficialChannelCode()
			if channelCode.State == constant.InviteChannelCodeStateValid {
				friendList = channelCode.FriendId
				groupList = channelCode.GroupId
				greeting = channelCode.Greeting
			}
		}
	case 2:
		if allowInvite {
			inviteCode = imdb.GetCodeInfoByCode(sourceCode)
			if inviteCode.State == constant.InviteCodeStateValid {
				friendList = inviteCode.UserId
				greeting = inviteCode.Greeting
			}
		}
	case 3:
		if allowChannel {
			//include valid and invalid
			channelCode, err = imdb.GetInviteChannelCodeByCode(sourceCode)
			if err == nil && channelCode.State == constant.InviteChannelCodeStateValid {
				friendList = channelCode.FriendId
				groupList = channelCode.GroupId
				greeting = channelCode.Greeting
			}
		}
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		resp.ErrCode = constant.ErrRPC.ErrCode
		resp.ErrMsg = errMsg
		return resp, err
	}
	client := pbChat.NewChatClient(etcdConn)

	if friendList != "" {
		friendIDList := strings.Split(friendList, ",")
		log.NewError(req.OperationID, "FriendUserIDList ", friendList)
		for _, v := range friendIDList {
			if _, fErr := imdb.GetUserByUserID(v); fErr != nil {
				log.NewError(req.OperationID, "GetUserByUserID failed ", fErr.Error(), v)
				continue
			} else {
				if _, err := imdb.GetFriendRelationshipFromFriend(opUserID, v); err != nil {
					// Establish two single friendship
					toInsertFollow := db.Friend{OwnerUserID: opUserID, FriendUserID: v}
					err1 := imdb.InsertToFriend(&toInsertFollow)
					if err1 != nil {
						log.NewError(req.OperationID, "InsertToFriend failed ", err1.Error(), toInsertFollow)
						continue
					}
					toInsertFollow = db.Friend{OwnerUserID: v, FriendUserID: opUserID}
					err2 := imdb.InsertToFriend(&toInsertFollow)
					if err2 != nil {
						log.NewError(req.OperationID, "InsertToFriend failed ", err2.Error(), toInsertFollow)
						continue
					}
					chat.FriendAddedNotification(req.OperationID, opUserID, opUserID, v)

					// add friend conversation
					var friendConv db.Conversation
					friendConv.OwnerUserID = opUserID
					friendConv.ConversationID = utils.GetConversationIDBySessionType(v, constant.SingleChatType)
					friendConv.ConversationType = constant.SingleChatType
					friendConv.UserID = v
					friendConv.IsNotInGroup = false
					if err := db.DB.SetSingleConversationRecvMsgOpt(friendConv.OwnerUserID, friendConv.ConversationID, friendConv.RecvMsgOpt); err != nil {
						log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
						continue
					}
					err := imdb.SetConversation(friendConv)
					if err != nil {
						log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
						continue
					}

					// add friend cache
					if err := db.DB.AddFriendToCache(opUserID, v); err != nil {
						log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", err.Error())
						continue
					}

					// add owner conversation
					var ownerConv db.Conversation
					ownerConv.OwnerUserID = v
					ownerConv.ConversationID = utils.GetConversationIDBySessionType(opUserID, constant.SingleChatType)
					ownerConv.ConversationType = constant.SingleChatType
					ownerConv.UserID = opUserID
					ownerConv.IsNotInGroup = false
					if err := db.DB.SetSingleConversationRecvMsgOpt(ownerConv.OwnerUserID, ownerConv.ConversationID, ownerConv.RecvMsgOpt); err != nil {
						log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
						continue
					}
					err = imdb.SetConversation(ownerConv)
					if err != nil {
						log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
						continue
					}

					// add owner cache
					if err := db.DB.AddFriendToCache(v, opUserID); err != nil {
						log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", err.Error())
						continue
					}

					// relation
					imdb.AddFriendRalation(v, opUserID, 2)

					//send welcome message
					if friend, err := imdb.GetUserByUserID(v); err == nil {
						var req pbChat.SendMsgReq
						var msg sdkws.MsgData
						req.OperationID = req.OperationID
						req.Token = req.Token
						msg.SendID = v
						msg.RecvID = opUserID
						msg.SenderNickname = friend.Nickname
						msg.SenderFaceURL = friend.FaceURL
						msg.Content = utils.String2Bytes(greeting)
						msg.MsgFrom = constant.UserMsgType
						msg.ContentType = constant.Text
						msg.SessionType = constant.SingleChatType
						msg.SendTime = utils.GetCurrentTimestampByMill()
						msg.CreateTime = utils.GetCurrentTimestampByMill()
						msg.ClientMsgID = utils.GetMsgID(v)
						req.MsgData = &msg

						_, err := client.SendMsg(context.Background(), &req)
						if err != nil {
							log.NewError(req.OperationID, "SendMsg rpc failed, ", req.String(), err.Error())
						}
					}
				} else {
					continue
				}
			}
		}
	}

	if groupList != "" {
		groupIDList := strings.Split(groupList, ",")
		log.NewError(req.OperationID, "groupIDList ", groupIDList)
		for _, v := range groupIDList {
			groupInfo, err := imdb.GetGroupInfoByGroupID(v)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
				continue
			}
			if groupInfo.Status == constant.GroupStatusDismissed {
				continue
			}
			if imdb.IsExistGroupMember(v, opUserID) {
				continue
			}

			// No super group for now
			if groupInfo.GroupType == constant.NormalGroup {
				// overcrowding
				count, err := imdb.GetGroupMembersCount(v, "")
				if err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
					continue
				}
				if int(count) >= config.Config.MembersInGroupMaxLimit {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), "over the group member number limit")
					continue
				}

				// add group
				groupOwnerInfo, err := imdb.GetGroupOwnerInfoByGroupID(v)
				if err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
					continue
				}

				// add group minseq number
				maxSeq, _ := db.DB.GetGrupMaxSeq(v)
				if maxSeq > 0 {
					maxSeq = maxSeq - uint64(config.Config.ServerCtrlGroupPrvMsg)
					if maxSeq < 0 {
						maxSeq = 0
					}
					_ = db.DB.SetGroupMinSeq(opUserID, v, uint32(maxSeq))
				}

				userInfo, err := imdb.GetUserByUserID(opUserID)
				if err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
					continue
				}

				var toInsertInfo db.GroupMember
				utils.CopyStructFields(&toInsertInfo, userInfo)
				toInsertInfo.GroupID = v
				toInsertInfo.RoleLevel = constant.GroupOrdinaryUsers
				toInsertInfo.OperatorUserID = ""
				if groupOwnerInfo != nil {
					toInsertInfo.OperatorUserID = groupOwnerInfo.UserID
				}
				err = imdb.InsertIntoGroupMember(toInsertInfo)
				if err != nil {
					log.NewError(req.OperationID, "channel InsertIntoGroupMember failed ", groupInfo.GroupID, err.Error())
					continue
				}

				// add conversation
				var groupConv db.Conversation
				groupConv.OwnerUserID = opUserID
				groupConv.ConversationID = utils.GetConversationIDBySessionType(v, constant.GroupChatType)
				groupConv.ConversationType = constant.GroupChatType
				groupConv.GroupID = v
				groupConv.IsNotInGroup = false
				if err := db.DB.SetSingleConversationRecvMsgOpt(groupConv.OwnerUserID, groupConv.ConversationID, groupConv.RecvMsgOpt); err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, rpc return", err.Error())
					continue
				}
				err = imdb.SetConversation(groupConv)
				if err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetConversation error", err.Error())
					continue
				}

				// cache
				if err := db.DB.AddGroupMemberToCache(v, []string{opUserID}...); err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddGroupMemberToCache failed", err.Error())
					continue
				}

				//notification
				chat.MemberInvitedNotification(req.OperationID, v, toInsertInfo.OperatorUserID, "", []string{opUserID})

				//welcome message
				//var req pbChat.SendMsgReq
				//var msg sdkws.MsgData
				//req.OperationID = req.OperationID
				//req.Token = req.Token
				//msg.SendID = owner.UserID
				//msg.GroupID = v
				//msg.SenderNickname = owner.Nickname
				//msg.SenderFaceURL = owner.FaceURL
				//msg.Content = utils.String2Bytes(greeting)
				//msg.MsgFrom = constant.UserMsgType
				//msg.ContentType = constant.Text
				//msg.SessionType = constant.GroupChatType
				//msg.SendTime = utils.GetCurrentTimestampByMill()
				//msg.CreateTime = utils.GetCurrentTimestampByMill()
				//msg.ClientMsgID = utils.GetMsgID(owner.UserID)
				//req.MsgData = &msg
				//
				//_, err := client.SendMsg(context.Background(), &req)
				//if err != nil {
				//	log.NewError(req.OperationID, "SendMsg rpc failed, ", req.String(), err.Error())
				//}

			}
		}
	}

	resp.ErrCode = constant.OK.ErrCode
	resp.ErrMsg = constant.OK.ErrMsg
	return resp, nil
}

func UpdateDeletedAccountInMoments(userID string) {

	//Delete All Moments set delete time
	db.DB.DeleteMomentsByUserID(userID)
	//Decrement Like count by this user
	db.DB.DecrementLikeCountInMomentByUserID(userID)
	//Delete Likes on Moments by this user
	db.DB.DeleteMomentLikesByUserID(userID)
	//Delete Comments on Moments by this user
	db.DB.DeleteMomentCommentsByUserID(userID)
}

func createFriendGRPConnection(OperationID string) (pbFriend.FriendClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, OperationID)
	if etcdConn == nil {
		errMsg := "etcd3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbFriend.NewFriendClient(etcdConn)
	return client, nil
}

func forceKickOff(userID string, platformID int32, operationID string) error {

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
