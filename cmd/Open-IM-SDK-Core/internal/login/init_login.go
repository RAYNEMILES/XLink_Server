package login

import (
	"Open_IM/cmd/Open-IM-SDK-Core/internal/cache"
	comm3 "Open_IM/cmd/Open-IM-SDK-Core/internal/common"
	conv "Open_IM/cmd/Open-IM-SDK-Core/internal/conversation_msg"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/friend"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/full"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/group"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/heartbeart"
	ws "Open_IM/cmd/Open-IM-SDK-Core/internal/interaction"
	comm2 "Open_IM/cmd/Open-IM-SDK-Core/internal/obj_storage"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/organization"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/signaling"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/super_group"
	"Open_IM/cmd/Open-IM-SDK-Core/internal/user"
	workMoments "Open_IM/cmd/Open-IM-SDK-Core/internal/work_moments"
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	sdk "Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	constant2 "Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	server_api_params2 "Open_IM/pkg/proto/sdk_ws"
	utils3 "Open_IM/pkg/utils"
	"encoding/json"
	"github.com/nanopack/mist/clients"
	mist "github.com/nanopack/mist/core"
	log2 "log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LoginMgr struct {
	organization *organization.Organization
	friend       *friend.Friend
	group        *group.Group
	superGroup   *super_group.SuperGroup
	conversation *conv.Conversation
	user         *user.User
	signaling    *signaling.LiveSignaling
	//advancedFunction advanced_interface.AdvancedFunction
	workMoments  *workMoments.WorkMoments
	full         *full.Full
	db           *db.DataBase
	ws           *ws.Ws
	msgSync      *ws.MsgSync
	heartbeat    *heartbeart.Heartbeat
	cache        *cache.Cache
	token        string
	loginUserID  string
	platformID   int32
	connListener open_im_sdk_callback.OnConnListener

	loginTime int64

	justOnceFlag bool

	groupListener        open_im_sdk_callback.OnGroupListener
	friendListener       open_im_sdk_callback.OnFriendshipListener
	conversationListener open_im_sdk_callback.OnConversationListener
	advancedMsgListener  open_im_sdk_callback.OnAdvancedMsgListener
	batchMsgListener     open_im_sdk_callback.OnBatchMsgListener
	userListener         open_im_sdk_callback.OnUserListener
	signalingListener    open_im_sdk_callback.OnSignalingListener
	organizationListener open_im_sdk_callback.OnOrganizationListener
	workMomentsListener  open_im_sdk_callback.OnWorkMomentsListener

	conversationCh     chan common.Cmd2Value
	cmdWsCh            chan common.Cmd2Value
	heartbeatCmdCh     chan common.Cmd2Value
	pushMsgAndMaxSeqCh chan common.Cmd2Value
	joinedSuperGroupCh chan common.Cmd2Value
	syncDataCh         chan common.Cmd2Value

	imConfig    sdk_struct.IMConfig
	tickerCount int
	client      *clients.TCP
}

func (u *LoginMgr) Organization() *organization.Organization {
	return u.organization
}

func (u *LoginMgr) Heartbeat() *heartbeart.Heartbeat {
	return u.heartbeat
}

func (u *LoginMgr) Ws() *ws.Ws {
	return u.ws
}

func (u *LoginMgr) ImConfig() sdk_struct.IMConfig {
	return u.imConfig
}

func (u *LoginMgr) Conversation() *conv.Conversation {
	return u.conversation
}

func (u *LoginMgr) User() *user.User {
	return u.user
}

func (u *LoginMgr) Full() *full.Full {
	return u.full
}

func (u *LoginMgr) Group() *group.Group {
	return u.group
}

func (u *LoginMgr) Friend() *friend.Friend {
	return u.friend
}

func (u *LoginMgr) Signaling() *signaling.LiveSignaling {
	return u.signaling
}

func (u *LoginMgr) WorkMoments() *workMoments.WorkMoments {
	return u.workMoments
}

func (u *LoginMgr) SetConversationListener(conversationListener open_im_sdk_callback.OnConversationListener) {
	u.conversationListener = conversationListener
}

func (u *LoginMgr) SetAdvancedMsgListener(advancedMsgListener open_im_sdk_callback.OnAdvancedMsgListener) {
	u.advancedMsgListener = advancedMsgListener
}

func (u *LoginMgr) SetBatchMsgListener(batchMsgListener open_im_sdk_callback.OnBatchMsgListener) {
	u.batchMsgListener = batchMsgListener
}
func (u *LoginMgr) SetFriendListener(friendListener open_im_sdk_callback.OnFriendshipListener) {
	u.friendListener = friendListener
}

func (u *LoginMgr) SetGroupListener(groupListener open_im_sdk_callback.OnGroupListener) {
	u.groupListener = groupListener
}

func (u *LoginMgr) SetOrganizationListener(listener open_im_sdk_callback.OnOrganizationListener) {
	u.organizationListener = listener
}

func (u *LoginMgr) SetUserListener(userListener open_im_sdk_callback.OnUserListener) {
	u.userListener = userListener
}

func (u *LoginMgr) SetSignalingListener(listener open_im_sdk_callback.OnSignalingListener) {
	u.signalingListener = listener
}

func (u *LoginMgr) SetWorkMomentsListener(listener open_im_sdk_callback.OnWorkMomentsListener) {
	u.workMomentsListener = listener
}

func (u *LoginMgr) wakeUp(cb open_im_sdk_callback.Base, operationID string) {
	log.Info(operationID, utils.GetSelfFuncName(), "args ")
	err := common.TriggerCmdWakeUp(u.heartbeatCmdCh)
	common.CheckAnyErrCallback(cb, 2001, err, operationID)
	cb.OnSuccess("")
}

func (u *LoginMgr) login(userID, token string, cb open_im_sdk_callback.Base, operationID string) {
	//log.NewError(operationID, "login start... ", userID, token, sdk_struct.SvrConf)
	err := common.TriggerCmdLogout(u.cmdWsCh)
	if err != nil {
		log2.Println(operationID, "TriggerCmdLogout failed ", err.Error())
		log.Error(operationID, "TriggerCmdLogout failed ", err.Error())
	}
	log.Info(operationID, "TriggerCmdLogout heartbeat...")
	err = common.TriggerCmdLogout(u.heartbeatCmdCh)
	if err != nil {
		log2.Println(operationID, "TriggerCmdLogout failed ", err.Error())
		log.Error(operationID, "TriggerCmdLogout failed ", err.Error())
	}

	t1 := time.Now()
	//make it compatible with multiple platform
	var platformID int32
	if strings.Contains(userID, "-") {
		platformStr := strings.Split(userID, "-")
		userID = platformStr[0]
		platformID64, _ := strconv.ParseInt(platformStr[1], 10, 64)
		platformID = int32(platformID64)
	} else {
		platformID = u.imConfig.Platform
	}

	sdk_struct.SvrConf.Platform = platformID

	////check token
	//_, err, _ := token_verify.WsVerifyToken(token, userID, strconv.FormatInt(int64(platformID), 10), operationID)
	//if err != nil {
	//	cb.OnError(constant.ErrTokenInvalid.ErrCode, err.Error())
	//	log.Error(operationID, "WsVerifyToken failed ", err.Error())
	//	return
	//}
	//
	////check permissions
	//err2 := utils2.CheckUserPermissions(userID)
	//if err2 != nil {
	//	cb.OnError(err2.ErrCode, err2.Error())
	//	log.Error(operationID, "CheckUserPermissions failed ", err2.Error())
	//	return
	//}

	err, exp := CheckToken(userID, token, platformID, operationID)
	if err != nil {
		log2.Println("CheckToken failed", err.Error(), userID, token, platformID, operationID)
		cb.OnError(constant.ErrTokenInvalid.ErrCode, constant2.ErrorUserLoginNetDisconnection.ErrMsg)
		return
	}
	common.CheckTokenErrCallback(cb, err, operationID)
	log.Info(operationID, "checkToken ok ", userID, platformID, token, exp, "login cost time: ", time.Since(t1))
	u.token = token
	u.loginUserID = userID
	u.platformID = platformID

	db, err := db.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		cb.OnError(constant.ErrDB.ErrCode, err.Error())
		log2.Println(operationID, "NewDataBase failed ", err.Error())
		log.Error(operationID, "NewDataBase failed ", err.Error())
		return
	}
	u.db = db
	log2.Println(operationID, "NewDataBase ok ", userID, sdk_struct.SvrConf.DataDir, "login cost time: ", time.Since(t1))
	log.Info(operationID, "NewDataBase ok ", userID, sdk_struct.SvrConf.DataDir, "login cost time: ", time.Since(t1))

	if u.db != nil {
		localUser, _ := u.db.GetLoginUser()
		if localUser != nil && localUser.Status == 2 {
			cb.OnError(constant2.ErrUserBanned.ErrCode, constant2.ErrUserBanned.ErrMsg)
			log2.Println(operationID, "you were banned", userID)
			log.Error(operationID, "you were banned", userID)
			return
		}
	}

	//init a channel for data synchronization
	u.syncDataCh = make(chan common.Cmd2Value, 1000)

	wsRespAsyn := ws.NewWsRespAsyn()
	wsConn := ws.NewWsConn(u.connListener, token, userID, platformID, u.syncDataCh)
	if wsConn.Conn == nil {
		log2.Println(operationID, "NewWsConn in login is failed!", token, userID, platformID)
		log.Debug(operationID, "NewWsConn in login is failed!", token, userID, platformID)
	}
	u.conversationCh = make(chan common.Cmd2Value, 1000)
	u.cmdWsCh = make(chan common.Cmd2Value, 10)

	u.heartbeatCmdCh = make(chan common.Cmd2Value, 10)
	u.pushMsgAndMaxSeqCh = make(chan common.Cmd2Value, 1000)
	u.ws = ws.NewWs(wsRespAsyn, wsConn, u.cmdWsCh, u.pushMsgAndMaxSeqCh, u.heartbeatCmdCh)
	u.joinedSuperGroupCh = make(chan common.Cmd2Value, 10)
	u.msgSync = ws.NewMsgSync(db, u.ws, userID, u.conversationCh, u.pushMsgAndMaxSeqCh, u.joinedSuperGroupCh)
	id2MinSeq := make(map[string]uint32, 100)
	p := ws.NewPostApi(token, sdk_struct.SvrConf.ApiAddr)

	u.user = user.NewUser(db, p, u.loginUserID, platformID)
	u.user.SetListener(u.userListener)

	u.friend = friend.NewFriend(u.loginUserID, u.db, u.user, p)
	u.friend.SetFriendListener(u.friendListener)

	u.group = group.NewGroup(u.loginUserID, u.db, p, u.joinedSuperGroupCh, u.heartbeatCmdCh)
	u.group.SetGroupListener(u.groupListener)
	u.superGroup = super_group.NewSuperGroup(u.loginUserID, u.db, p, u.joinedSuperGroupCh, u.heartbeatCmdCh)
	u.organization = organization.NewOrganization(u.loginUserID, u.db, p)
	u.organization.SetListener(u.organizationListener)
	u.cache = cache.NewCache(u.user, u.friend)
	u.full = full.NewFull(u.user, u.friend, u.group, u.conversationCh, u.cache, u.db, u.superGroup)
	u.workMoments = workMoments.NewWorkMoments(u.loginUserID, u.db, p)
	u.workMoments.SetListener(u.workMomentsListener)
	log2.Println(operationID, u.imConfig.ObjectStorage, "new obj login cost time: ", time.Since(t1))
	log.NewInfo(operationID, u.imConfig.ObjectStorage, "new obj login cost time: ", time.Since(t1))
	u.user.SyncLoginUserInfo(operationID)
	log2.Println(operationID, u.imConfig.ObjectStorage, "SyncLoginUserInfo login cost time: ", time.Since(t1))
	log.NewInfo(operationID, u.imConfig.ObjectStorage, "SyncLoginUserInfo login cost time: ", time.Since(t1))
	u.loginTime = utils.GetCurrentTimestampByMill()
	u.user.SetLoginTime(u.loginTime)
	u.friend.SetLoginTime(u.loginTime)
	u.group.SetLoginTime(u.loginTime)
	u.superGroup.SetLoginTime(u.loginTime)
	u.organization.SetLoginTime(u.loginTime)
	go u.forcedSynchronization()
	u.heartbeat = heartbeart.NewHeartbeat(u.msgSync, u.heartbeatCmdCh, u.syncDataCh, u.connListener, token, exp, id2MinSeq, u.full)
	log2.Println(operationID, "forcedSynchronization success...", "login cost time: ", time.Since(t1))
	log2.Println(operationID, "all channel ", u.pushMsgAndMaxSeqCh, u.conversationCh, u.heartbeatCmdCh, u.cmdWsCh)
	log2.Println(operationID, u.imConfig.ObjectStorage)
	log.Info(operationID, "forcedSynchronization success...", "login cost time: ", time.Since(t1))
	log.Info(operationID, "all channel ", u.pushMsgAndMaxSeqCh, u.conversationCh, u.heartbeatCmdCh, u.cmdWsCh)
	log.NewInfo(operationID, u.imConfig.ObjectStorage)
	var objStorage comm3.ObjectStorage
	switch u.imConfig.ObjectStorage {
	case "cos":
		objStorage = comm2.NewCOS(p)
	case "minio":
		objStorage = comm2.NewMinio(p)
	case "oss":
		objStorage = comm2.NewOSS(p)
	default:
		objStorage = comm2.NewCOS(p)
	}
	u.signaling = signaling.NewLiveSignaling(u.ws, u.signalingListener, u.loginUserID, u.imConfig.Platform, u.db)

	u.conversation = conv.NewConversation(u.ws, u.db, p, u.conversationCh,
		u.loginUserID, u.imConfig.Platform, u.imConfig.DataDir,
		u.friend, u.group, u.user, objStorage, u.conversationListener, u.advancedMsgListener,
		u.organization, u.signaling, u.workMoments, u.cache, u.full, id2MinSeq)
	if u.batchMsgListener != nil {
		u.conversation.SetBatchMsgListener(u.batchMsgListener)
		log.Info(operationID, "SetBatchMsgListener ", u.batchMsgListener)
	}

	go u.readSyncChanMsg()

	go u.ConnectToSubseriberMist()

	go u.conversation.GetOfficialAccountConversationFormServer(operationID)

	log.Debug(operationID, "SyncConversations begin ")
	//u.conversation.SyncConversations(operationID)
	log.Debug(operationID, "SyncConversations end ")

	go common.DoListener(u.conversation)
	log2.Println(operationID, "login success...", "login cost time: ", time.Since(t1))
	log.Info(operationID, "login success...", "login cost time: ", time.Since(t1))
	cb.OnSuccess("")

	//time.AfterFunc(5*time.Second, func() {
	//	_, err := u.db.GetLocalConfig(u.loginUserID)
	//	if err != nil {
	//		go func(userID string) {
	//			operationID := utils.OperationIDGenerator()
	//			u.loadWelcomeMessages(operationID)
	//			u.db.SetLocalConfig(&model_struct.LocalConfig{UserID: userID, FirstWelcome: 2})
	//		}(u.loginUserID)
	//	}
	//})
}
func (u *LoginMgr) InitSDK(config sdk_struct.IMConfig, listener open_im_sdk_callback.OnConnListener, operationID string) bool {
	u.imConfig = config
	log.NewInfo(operationID, utils.GetSelfFuncName(), config)
	if listener == nil {
		return false
	}
	u.connListener = listener
	return true
}

func (u *LoginMgr) logout(callback open_im_sdk_callback.Base, operationID string) {
	log.Info(operationID, "TriggerCmdLogout ws...")

	if u.friend == nil || u.conversation == nil || u.user == nil || u.full == nil ||
		u.db == nil || u.ws == nil || u.msgSync == nil || u.heartbeat == nil {
		log.Info(operationID, "nil, no TriggerCmdLogout ", *u)
		return
	}

	//if u.conversation != nil && u.conversation.Pool != nil {
	//	u.conversation.Pool.Release()
	//}
	err := common.TriggerCmdLogout(u.cmdWsCh)
	if err != nil {
		log.Error(operationID, "TriggerCmdLogout failed ", err.Error())
	}
	log.Info(operationID, "TriggerCmdLogout heartbeat...")
	err = common.TriggerCmdLogout(u.heartbeatCmdCh)
	if err != nil {
		log.Error(operationID, "TriggerCmdLogout failed ", err.Error())
	}
	log.Info(operationID, "TriggerCmd conversationCh UnInit...")
	common.UnInitAll(u.conversationCh)
	if err != nil {
		log.Error(operationID, "TriggerCmd UnInit conversation failed ", err.Error())
	}

	log.Info(operationID, "TriggerCmd pushMsgAndMaxSeqCh UnInit...")
	common.UnInitAll(u.pushMsgAndMaxSeqCh)
	if err != nil {
		log.Error(operationID, "TriggerCmd UnInit pushMsgAndMaxSeqCh failed ", err.Error())
	}

	timeout := 2
	retryTimes := 0
	log.Info(operationID, "send to svr logout ...", u.loginUserID)
	resp, err := u.ws.SendReqWaitResp(&server_api_params2.GetMaxAndMinSeqReq{}, constant.WsLogoutMsg, timeout, retryTimes, u.loginUserID, operationID)
	if err != nil {
		log.Warn(operationID, "SendReqWaitResp failed ", err.Error(), constant.WsLogoutMsg, timeout, u.loginUserID, resp)
	}
	if callback != nil {
		callback.OnSuccess("")
	}
	u.justOnceFlag = false

	//go func(mgr *LoginMgr) {
	//	time.Sleep(5 * time.Second)
	//	if mgr == nil {
	//		log.Warn(operationID, "login mgr == nil")
	//		return
	//	}
	//	log.Warn(operationID, "logout close   channel ", mgr.heartbeatCmdCh, mgr.cmdWsCh, mgr.pushMsgAndMaxSeqCh, mgr.conversationCh, mgr.loginUserID)
	//	close(mgr.heartbeatCmdCh)
	//	close(mgr.cmdWsCh)
	//	close(mgr.pushMsgAndMaxSeqCh)
	//	close(mgr.conversationCh)
	//	mgr = nil
	//}(u)
}

func (u *LoginMgr) ClearCache(callback open_im_sdk_callback.Base, operationID string) {
	err := db.RemoveAllLocalDatabases(sdk_struct.SvrConf.DataDir)
	if callback != nil && err == nil {
		callback.OnSuccess("")
	}
}

func (u *LoginMgr) GetLoginUser() string {
	return u.loginUserID
}

func (u *LoginMgr) GetLoginStatus() int32 {
	//go u.updateUserIPAndStatus()

	return u.ws.LoginState()
}

func (u *LoginMgr) forcedSynchronization() {
	operationID := utils.OperationIDGenerator()
	log.NewInfo(operationID, utils.GetSelfFuncName(), "start synchronization")
	var wg sync.WaitGroup
	wg.Add(8)
	go func() {
		u.friend.SyncFriendList(operationID)
		wg.Done()
	}()

	go func() {
		u.friend.SyncBlackList(operationID)
		wg.Done()
	}()

	go func() {
		u.friend.SyncFriendApplication(operationID)
		wg.Done()
	}()

	go func() {
		u.friend.SyncSelfFriendApplication(operationID)
		wg.Done()
	}()

	go func() {
		u.group.SyncJoinedGroupList(operationID)
		u.msgSync.CompareGroupSeq()
		wg.Done()
		// go u.group.SyncGroupUpdatesVerion(operationID)
	}()

	go func() {
		u.group.SyncMyInfoInAllGroupForFirstLogin(operationID)
		wg.Done()
	}()

	go func() {
		u.group.SyncAdminGroupApplication(operationID)
		wg.Done()
	}()

	go func() {
		u.group.SyncSelfGroupApplication(operationID)
		wg.Done()
	}()

	// go func() {
	// u.group.SyncJoinedGroupMemberForFirstLogin(operationID)
	// 	wg.Done()
	// }()

	//go func() {
	//	u.organization.SyncOrganization(operationID)
	//	wg.Done()
	//}()

	//go func() {
	//	u.superGroup.SyncJoinedGroupList(operationID)
	//	wg.Done()
	//}()

	wg.Wait()

}

func (u *LoginMgr) GetMinSeqSvr() int64 {
	return u.GetMinSeqSvr()
}

func (u *LoginMgr) SetMinSeqSvr(minSeqSvr int64) {
	u.SetMinSeqSvr(minSeqSvr)
}

func CheckToken(userID, token string, platformID int32, operationID string) (error, uint32) {
	if operationID == "" {
		operationID = utils.OperationIDGenerator()
	}
	log.Debug(operationID, utils.GetSelfFuncName(), userID, token)
	p := ws.NewPostApi(token, sdk_struct.SvrConf.ApiAddr)
	user := user.NewUser(nil, p, userID, platformID)
	//_, err := user.GetSelfUserInfoFromSvr(operationID)
	//if err != nil {
	//	return utils.Wrap(err, "GetSelfUserInfoFromSvr failed "+operationID), 0
	//}
	exp, err := user.ParseTokenFromSvr(operationID)
	return err, exp
}

func (u *LoginMgr) uploadImage(callback open_im_sdk_callback.Base, filePath string, token, obj string, operationID string) string {
	p := ws.NewPostApi(token, u.ImConfig().ApiAddr)
	var o comm3.ObjectStorage
	switch obj {
	case "cos":
		o = comm2.NewCOS(p)
	case "minio":
		o = comm2.NewMinio(p)
	default:
		o = comm2.NewCOS(p)
	}
	url, _, err := o.UploadImage(filePath, "", func(progress int) {
		if progress == 100 {
			callback.OnSuccess("")
		}
	})
	if err != nil {
		log.Error(operationID, "UploadImage failed ", err.Error(), filePath)
		return ""
	}
	return url
}

func (u LoginMgr) uploadFile(callback open_im_sdk_callback.SendMsgCallBack, filePath, operationID string) {
	url, _, err := u.conversation.UploadFile(filePath, "", callback.OnProgress)
	log.NewInfo(operationID, utils.GetSelfFuncName(), url)
	if err != nil {
		log.Error(operationID, "UploadImage failed ", err.Error(), filePath)
		callback.OnError(constant.ErrApi.ErrCode, err.Error())
	}
	callback.OnSuccess(url)
}

func (u LoginMgr) readSyncChanMsg() {
	//operationID2 := utils.OperationIDGenerator()
	//log.NewError(operationID2, utils.GetSelfFuncName(), "进入读取同步数据！")

	operationID := utils.OperationIDGenerator()
	for {
		select {
		case r1 := <-u.cmdWsCh:
			{
				log.Debug(operationID, utils.GetSelfFuncName(), "readSyncChanMsg cmd ...", r1.Value, u.loginUserID)
			}
		case r := <-u.syncDataCh:
			if r.Value != nil {
				switch r.Cmd {
				case constant2.SyncCreateGroup:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localGroup := model_struct.LocalGroup{}
					err := json.Unmarshal(rValueByte, &localGroup)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.InsertGroup(&localGroup)
					if err == nil {
						go u.group.SyncMySelfInTheGroup(localGroup.GroupID, operationID)
						u.group.SyncGroupMemberByGroupID(localGroup.GroupID, operationID, false)
						u.msgSync.CompareGroupSeq()
						callbackData := sdk.GroupInfoChangedCallback(localGroup)
						u.groupListener.OnJoinedGroupAdded(utils.StructToJsonString(callbackData))
						u.conversation.SyncOneConversation(utils.GetConversationIDBySessionType(localGroup.GroupID, constant.GroupChatType), operationID)
					}

				case constant2.SyncKickedGroupMember, constant2.SyncInvitedGroupMember, constant2.SyncMuteGroupMember, constant2.SyncCancelMuteGroupMember, constant2.SyncGroupMemberInfo:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					var localMemberList []model_struct.LocalGroupMember
					err := json.Unmarshal(rValueByte, &localMemberList)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					syncType := 0
					if r.Cmd == constant2.SyncKickedGroupMember {
						syncType = 1
					} else if r.Cmd == constant2.SyncInvitedGroupMember {
						syncType = 2
					} else {
						syncType = 3
					}

					groupID := ""
					isSelf := false
					for _, member := range localMemberList {
						log.NewError(operationID, utils.GetSelfFuncName(), "GroupID", member.GroupID, "UserID", member.UserID)
						u.group.SyncOneGroupMemberInfo(operationID, member.UserID, member.GroupID, syncType)
						if member.UserID == u.loginUserID {
							isSelf = true
						}
						groupID = member.GroupID
					}

					if r.Cmd == constant2.SyncKickedGroupMember {
						//someone was kicked out the group
						localGroup, err := u.group.GetGroupInfoFromSvr(groupID)
						if err == nil && localGroup.GroupID != "" {
							err = u.db.UpdateGroup(localGroup)
							if err == nil {
								callbackData := sdk.GroupInfoChangedCallback(*localGroup)
								u.groupListener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
							}
						}

						if isSelf {
							localGroup, err := u.db.GetGroupInfoByGroupID(groupID)
							if err == nil && localGroup.GroupID != "" {
								u.db.DeleteGroup(groupID)
								callbackData := sdk.GroupInfoChangedCallback(*localGroup)
								u.groupListener.OnJoinedGroupDeleted(utils.StructToJsonString(callbackData))
							}
						}

						u.conversation.SyncOneConversation(utils.GetConversationIDBySessionType(groupID, constant.GroupChatType), operationID)

					} else if r.Cmd == constant2.SyncInvitedGroupMember {
						//someone was invited to the group
						if isSelf {
							localGroup, err := u.group.GetGroupInfoFromSvr(groupID)
							if err == nil && localGroup.GroupID != "" {
								u.db.InsertGroup(localGroup)
								callbackData := sdk.GroupInfoChangedCallback(*localGroup)
								u.msgSync.CompareGroupSeq()
								u.groupListener.OnJoinedGroupAdded(utils.StructToJsonString(callbackData))
							}
						} else {
							localGroup, err := u.group.GetGroupInfoFromSvr(groupID)
							if err == nil && localGroup.GroupID != "" {
								err = u.db.UpdateGroup(localGroup)
								if err == nil {
									callbackData := sdk.GroupInfoChangedCallback(*localGroup)
									u.groupListener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
								}
							}
						}

						u.conversation.SyncOneConversation(utils.GetConversationIDBySessionType(groupID, constant.GroupChatType), operationID)

					} else {
						localGroup, err := u.db.GetGroupInfoByGroupID(groupID)
						if err == nil && localGroup.GroupID != "" {
							callbackData := sdk.GroupInfoChangedCallback(*localGroup)
							u.groupListener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
						}
						u.conversation.SyncOneConversation(utils.GetConversationIDBySessionType(groupID, constant.GroupChatType), operationID)
					}
				case constant2.SyncUpdateGroup:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localGroup := model_struct.LocalGroup{}
					err := json.Unmarshal(rValueByte, &localGroup)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.UpdateGroup(&localGroup)
					if err == nil {
						callbackData := sdk.GroupInfoChangedCallback(localGroup)
						u.groupListener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
					}

				case constant2.SyncDeleteGroup:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localGroup := model_struct.LocalGroup{}
					err := json.Unmarshal(rValueByte, &localGroup)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.DeleteGroup(localGroup.GroupID)
					if err == nil {
						u.db.DeleteGroupAllMembers(localGroup.GroupID)
						callbackData := sdk.GroupInfoChangedCallback(localGroup)
						u.groupListener.OnJoinedGroupDeleted(utils.StructToJsonString(callbackData))
					}

				case constant2.SyncMuteGroup, constant2.SyncCancelMuteGroup:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localGroup := model_struct.LocalGroup{}
					err := json.Unmarshal(rValueByte, &localGroup)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.UpdateGroup(&localGroup)
					if err == nil {
						callbackData := sdk.GroupInfoChangedCallback(localGroup)
						u.groupListener.OnGroupInfoChanged(utils.StructToJsonString(callbackData))
					}

				case constant2.SyncConversation:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localConv := model_struct.LocalConversation{}
					err := json.Unmarshal(rValueByte, &localConv)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}
					u.conversation.SyncOneConversation(localConv.ConversationID, operationID)

					var list []*model_struct.LocalConversation
					oc, err := u.db.GetConversation(localConv.ConversationID)
					if err == nil {
						log.Info("this is old conversation UnreadCount and LatestMsgSendTime", oc.UnreadCount, oc.LatestMsgSendTime)
						log.Info("this is new conversation UnreadCount and LatestMsgSendTime", localConv.UnreadCount, localConv.LatestMsgSendTime)
						if localConv.LatestMsgSendTime >= oc.LatestMsgSendTime {
							//latest msg in local older than passing data
							err := u.db.UpdateColumnsConversation(localConv.ConversationID, map[string]interface{}{"latest_msg_send_time": localConv.LatestMsgSendTime, "latest_msg": localConv.LatestMsg})
							if err != nil {
								log.Error("internal", "updateConversationLatestMsgModel err: ", err)
							} else {
								oc.LatestMsgSendTime = localConv.LatestMsgSendTime
								oc.LatestMsg = localConv.LatestMsg
								oc.ShowName = localConv.ShowName
								list = append(list, oc)
								u.conversationListener.OnConversationChanged(utils.StructToJsonString(list))
							}
						} else {
							//latest msg in local newer than passing data
							err := u.db.UpdateColumnsConversation(localConv.ConversationID, map[string]interface{}{"latest_msg_send_time": oc.LatestMsgSendTime, "latest_msg": oc.LatestMsg})
							if err != nil {
								log.Error("internal", "updateConversationLatestMsgModel err: ", err)
							} else {
								localConv.LatestMsgSendTime = oc.LatestMsgSendTime
								localConv.LatestMsg = oc.LatestMsg
								localConv.ShowName = oc.ShowName
								localConv.FaceURL = oc.FaceURL
								list = append(list, &localConv)
								u.conversationListener.OnConversationChanged(utils.StructToJsonString(list))
							}
						}
					}
					//list := []*model_struct.LocalConversation{&localConv}
					//u.conversationListener.OnConversationChanged(utils.StructToJsonString(list))
				case constant2.SyncGroupRequest:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localGroupReq := model_struct.LocalGroupRequest{}
					err := json.Unmarshal(rValueByte, &localGroupReq)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.UpdateGroupRequest(&localGroupReq)
					if err == nil && localGroupReq.HandleResult == 1 {
						//accept， sync group info and group members
						localGroup, err := u.group.GetGroupInfoFromSvr(localGroupReq.GroupID)
						if err == nil && localGroup.GroupID != "" {
							u.db.InsertGroup(localGroup)
							go u.group.SyncMySelfInTheGroup(localGroup.GroupID, operationID)
							callbackData := sdk.GroupInfoChangedCallback(*localGroup)
							u.groupListener.OnJoinedGroupAdded(utils.StructToJsonString(callbackData))
							u.msgSync.CompareGroupSeq()
							u.group.SyncGroupMemberByGroupID(localGroupReq.GroupID, operationID, false)
						}
					} else {
						//reject
					}
				case constant2.SyncAdminGroupRequest:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localAdminGroupReq := model_struct.LocalAdminGroupRequest{}
					err := json.Unmarshal(rValueByte, &localAdminGroupReq)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.UpdateAdminGroupRequest(&localAdminGroupReq)
					if err == nil && localAdminGroupReq.HandleResult == 1 {
						//accept， sync group members
						u.group.SyncOneGroupMemberInfo(operationID, localAdminGroupReq.UserID, localAdminGroupReq.GroupID, 2)
					} else {
						//reject
					}
				case constant2.SyncFriendRequest:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localFriendReq := model_struct.LocalFriendRequest{}
					err := json.Unmarshal(rValueByte, &localFriendReq)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					_, err = u.db.GetFriendApplicationByBothID(localFriendReq.FromUserID, localFriendReq.ToUserID)
					if err != nil {
						err = u.db.InsertFriendRequest(&localFriendReq)
						if err == nil {
							callbackData := sdk.FriendApplicationAddedCallback(localFriendReq)
							u.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
						}
					} else {
						err = u.db.UpdateFriendRequest(&localFriendReq)
						if err == nil {
							callbackData := sdk.FriendApplicationAddedCallback(localFriendReq)
							if localFriendReq.HandleResult == 1 {
								u.friendListener.OnFriendApplicationAccepted(utils.StructToJsonString(callbackData))
							} else {
								u.friendListener.OnFriendApplicationRejected(utils.StructToJsonString(callbackData))
							}
						}
					}

					if err == nil && localFriendReq.HandleResult == 1 {
						//accept， sync friend list
						svrList, err := u.friend.GetFriendInfoFromSvr(operationID, []string{localFriendReq.FromUserID})
						if err == nil {
							localFriendList := common.TransferToLocalFriend(svrList)
							if localFriendList != nil && len(localFriendList) > 0 {
								err = u.db.InsertFriend(localFriendList[0])
								if err == nil {
									callbackData := sdk.FriendAddedCallback(*localFriendList[0])
									u.friendListener.OnFriendAdded(utils.StructToJsonString(callbackData))
								}
							}
						}

					} else {
						//reject
					}
				case constant2.SyncSelfFriendRequest:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localFriendReq := model_struct.LocalFriendRequest{}
					err := json.Unmarshal(rValueByte, &localFriendReq)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					_, err = u.db.GetFriendApplicationByBothID(localFriendReq.FromUserID, localFriendReq.ToUserID)
					if err != nil {
						err = u.db.InsertFriendRequest(&localFriendReq)
						if err == nil {
							callbackData := sdk.FriendApplicationAddedCallback(localFriendReq)
							u.friendListener.OnFriendApplicationAdded(utils.StructToJsonString(callbackData))
						}
					} else {
						err = u.db.UpdateFriendRequest(&localFriendReq)
						if err == nil {
							callbackData := sdk.FriendApplicationAddedCallback(localFriendReq)
							if localFriendReq.HandleResult == 1 {
								u.friendListener.OnFriendApplicationAccepted(utils.StructToJsonString(callbackData))
							} else {
								u.friendListener.OnFriendApplicationRejected(utils.StructToJsonString(callbackData))
							}
						}
					}

					if err == nil && localFriendReq.HandleResult == 1 {
						//accept， sync friend list
						svrList, err := u.friend.GetFriendInfoFromSvr(operationID, []string{localFriendReq.ToUserID})
						if err == nil {
							localList := common.TransferToLocalFriend(svrList)
							if localList != nil && len(localList) > 0 {
								err = u.db.InsertFriend(localList[0])
								if err == nil {
									callbackData := sdk.FriendAddedCallback(*localList[0])
									u.friendListener.OnFriendAdded(utils.StructToJsonString(callbackData))
								}
							}
						}

					} else {
						//reject
					}
				case constant2.SyncFriendInfo:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localFriend := model_struct.LocalFriend{}
					err := json.Unmarshal(rValueByte, &localFriend)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}
					u.friend.SyncFriendInfo(operationID, []string{localFriend.FriendUserID})
					callbackData := sdk.FriendInfoChangedCallback(localFriend)
					u.friendListener.OnFriendInfoChanged(utils.StructToJsonString(callbackData))
				case constant2.SyncAddBlackList:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localBlack := model_struct.LocalBlack{}
					err := json.Unmarshal(rValueByte, &localBlack)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.InsertBlack(&localBlack)
					if err == nil {
						callbackData := sdk.BlackAddCallback(localBlack)
						u.friendListener.OnBlackAdded(utils.StructToJsonString(callbackData))
					}

				case constant2.SyncDeleteBlackList:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localBlack := model_struct.LocalBlack{}
					err := json.Unmarshal(rValueByte, &localBlack)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}

					err = u.db.DeleteBlack(localBlack.BlockUserID)
					if err == nil {
						callbackData := sdk.BlackDeletedCallback(localBlack)
						u.friendListener.OnBlackDeleted(utils.StructToJsonString(callbackData))
					}
				case constant2.SyncUserInfo:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					localUser := model_struct.LocalUser{}
					err := json.Unmarshal(rValueByte, &localUser)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}
					u.user.SyncLoginUserInfo(operationID)
					callbackData := sdk.SelfInfoUpdatedCallback(localUser)
					u.userListener.OnSelfInfoUpdated(utils.StructToJsonString(callbackData))
				case constant2.SyncWelcomeMessageFromChannel:
					rValueByte := utils3.String2Bytes(r.Value.(string))
					var msgList []*server_api_params2.MsgData
					err := json.Unmarshal(rValueByte, &msgList)
					if err != nil {
						log.NewError(operationID, utils.GetSelfFuncName(), "json.Unmarshal failed", err.Error())
						continue
					}
					u.msgSync.SelfMsgSync.TriggerCmdNewMsgCome(msgList, operationID)
				}

			} else {
				switch r.Cmd {
				case constant.CmdLogout:
					log.NewError(operationID, "sync data channel recv logout cmd...", u.loginUserID)
					return
				}
			}
		}
	}
}

func (lm *LoginMgr) updateUserIPAndStatus() {
	//operationID := utils.OperationIDGenerator()
	// ipAddress :=  ws.Conn.RemoteAddr().String()
	// if ipAddress != "" {
	// 	ipAndPort := strings.Split(ipAddress, ":")
	// 	if len(ipAndPort) > 0 {
	// 		ipAddress = ipAndPort[0]
	// 	}
	// }
	//lm.user.UpdateUserIPandStatus(operationID)
}

func (lm *LoginMgr) ConnectToSubseriberMist() {
	//var err error
	//lm.client, err = clients.New("127.0.0.1:10330", "Token")
	//if err != nil {
	//	fmt.Printf("MSG: %#v\n", err.Error())
	//	//os.Exit(1)
	//	return
	//}
	//err = lm.client.Ping()
	//if err != nil {
	//	return
	//}
	//// example commands (not handling errors for brevity)
	//lm.client.Subscribe([]string{"new_article", "articleID WIll Come Here"})
	//lm.client.Subscribe([]string{lm.loginUserID})
	////lm.client.Publish([]string{"hello"}, "world")
	////lm.client.List()
	//// client.Unsubscribe([]string{"hello"})
	//for {
	//	select {
	//	case msg := <-lm.client.Messages():
	//		fmt.Printf("MSG: %#v\n", msg)
	//		lm.performActionbaseOnMistTrigger(msg)
	//	case <-time.After(time.Second * 1):
	//		// do something if messages are taking too long
	//	}
	//}
}

func (u *LoginMgr) performActionbaseOnMistTrigger(msg mist.Message) {
	//if msg.Command == "publish" && msg.Error == "" {
	//	tags := msg.Tags
	//	if len(tags) > 0 {
	//		if utils.IsContain("new_article", tags) {
	//			newArticlePosted := server_api_params.ArticleForLocalConv{}
	//			err := json.Unmarshal([]byte(msg.Data), &newArticlePosted)
	//			if err == nil {
	//				u.conversation.UpdateOfficialArticleAsLocalConv(newArticlePosted, "MistTrigerMessaeg")
	//			}
	//
	//		}
	//	}
	//
	//}
}

func (u *LoginMgr) loadWelcomeMessages(operationID string) {
	_, err := u.user.StartingWelcomeMessagesFromSvr(operationID)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "load welcome messages error:", err.Error())
	}
}

func (u *LoginMgr) sendMessage(userID, content, operationID string) {
	var wsMsgData server_api_params2.MsgData
	wsMsgData.SendID = u.loginUserID
	wsMsgData.RecvID = userID
	wsMsgData.ClientMsgID = "return-welcome-cmsgid-" + userID
	wsMsgData.ServerMsgID = "return-welcome-smsgid-" + userID
	wsMsgData.MsgFrom = constant.UserMsgType
	wsMsgData.ContentType = constant.Text
	wsMsgData.Content = []byte(content)
	wsMsgData.Seq = 0
	wsMsgData.SortTime = utils.GetCurrentTimestampByMill()
	wsMsgData.SendTime = utils.GetCurrentTimestampByMill()
	wsMsgData.CreateTime = utils.GetCurrentTimestampByMill()
	wsMsgData.Status = constant.MsgStatusSending
	timeout := 15
	retryTimes := 2
	_, err := u.conversation.SendReqWaitResp(&wsMsgData, constant.WSSendMsg, timeout, retryTimes, u.loginUserID, operationID)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
	}
}
