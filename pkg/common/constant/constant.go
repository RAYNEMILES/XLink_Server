package constant

var OnlyForTest = 0

const (

	// group admin
	//	OrdinaryMember = 0
	//	GroupOwner     = 1
	//	Administrator  = 2
	// group application
	//	Application      = 0
	//	AgreeApplication = 1

	// friend related
	BlackListFlag         = 1
	ApplicationFriendFlag = 0
	FriendFlag            = 1
	RefuseFriendFlag      = -1

	//environment
	DEV  = "DEV"
	TEST = "TEST"
	PROD = "PROD"

	ApiTimeOutSeconds = 60

	// Redis Key For Status
	UserStatusKey    = "user_status"
	AdminStatusKey   = "admin_status"
	UserIPandStatus  = "ip_status:user"
	UserGroupIDCache = "USER_GROUPID_LIST"
	AdminRolesKey    = "admin_roles"
	InAppLoginPin    = "inApp_login_pin"

	// Websocket Protocol
	WSGetNewestSeq     = 1001
	WSPullMsgBySeqList = 1002
	WSSendMsg          = 1003
	WSSendSignalMsg    = 1004
	WSSendBroadcastMsg = 1006
	WSPushMsg          = 2001
	WSKickOnlineMsg    = 2002
	WsLogoutMsg        = 2003
	WsSyncDataMsg      = 2004
	WSDataError        = 3001

	// /ContentType
	// UserRelated
	Text                = 101 // 文本 Text
	Picture             = 102 // 图片 Picture
	Voice               = 103 // 语音 Voice
	Video               = 104 // 视频 Video
	File                = 105 // 文件 File
	AtText              = 106 // @消息 @Text
	Merger              = 107 // 合并消息 Message Merge
	Card                = 108 // 名片 Name Card
	Location            = 109 // 位置 Location
	Custom              = 110 // 自定义 Custom
	Revoke              = 111 // 撤回 Revoke
	HasReadReceipt      = 112 // 已读回执 Has Readed Receipt
	Typing              = 113 // 输入中 Typing
	Quote               = 114 // 引用 Quote
	GroupHasReadReceipt = 116 // 群消息已读回执 Group Message Has Read Receipt
	Common              = 200 // 公共消息 Common
	GroupMsg            = 201 // 群消息 Group Message
	SignalMsg           = 202 // 信号消息 Signal Message

	// SysRelated
	NotificationBegin                     = 1000 // 开始通知
	DeleteMessageNotification             = 1100 // 删除消息通知
	FriendApplicationApprovedNotification = 1201 // 接受好友 Friend Approved add_friend_response
	FriendApplicationRejectedNotification = 1202 // 拒绝好友 Friend Rejected add_friend_response
	FriendApplicationNotification         = 1203 // 申请添加好友 add_friend
	FriendAddedNotification               = 1204 // 添加好友成功 Added Friend
	FriendDeletedNotification             = 1205 // 删除好友 Delete Friend delete_friend
	FriendRemarkSetNotification           = 1206 // 好友备注 Friend Remark set_friend_remark?
	BlackAddedNotification                = 1207 // 拉黑 Add Blacklist add_black
	BlackDeletedNotification              = 1208 // 解除黑名单 Delete Blacklist remove_black

	ConversationOptChangeNotification = 1300 // 会话改变 Conversation Changed change conversation opt

	UserNotificationBegin       = 1301
	UserInfoUpdatedNotification = 1303 // 用户信息更新 User Info Updated
	UserNotFriendNotification   = 1304 // 用户信息更新 User Info Updated
	UserBlockedYouNotification  = 1305 // 用户信息更新 User Info Updated
	UserNotificationEnd         = 1399
	OANotification              = 1400 // OA通知 OA Notification

	GroupNotificationBegin = 1500

	GroupCreatedNotification                 = 1501 // 建群 Create Group
	GroupInfoSetNotification                 = 1502 // 设置群信息 Group Info Setting
	JoinGroupApplicationNotification         = 1503 // 加群 Join Group
	MemberQuitNotification                   = 1504 // 退群 Exit Group
	GroupApplicationAcceptedNotification     = 1505 // 接受入群 Group Accepted
	GroupApplicationRejectedNotification     = 1506 // 拒绝入群 Group Rejected
	GroupOwnerTransferredNotification        = 1507 // 群主转让 Group Owner Transferred
	MemberKickedNotification                 = 1508 // 踢人 Member Kicked
	MemberInvitedNotification                = 1509 // 拉人 Member Invited
	MemberEnterNotification                  = 1510 // 进群 Member Enter
	GroupDismissedNotification               = 1511 // 群解散 Group Dismissed
	GroupMemberMutedNotification             = 1512 // 成员禁言 Member Muted
	GroupMemberCancelMutedNotification       = 1513 // 取消成员禁言 Member Cancel Muted
	GroupMutedNotification                   = 1514 // 群禁言 Group Muted
	GroupCancelMutedNotification             = 1515 // 取消群禁言 Group Cancel Muted
	GroupMemberInfoSetNotification           = 1516 // 设置成员信息 Member Info Setting
	GroupMemberSetToAdminNotification        = 1517 // 设置管理员 Admin Setting
	GroupMemberSetToOrdinaryUserNotification = 1518 // 取消管理员 Cancel Admin
	GroupAnnouncementNotification            = 1519 // 发布群公告 Post Group Announcement
	GroupDeleteNotification                  = 1520 // 群解散 Group Dismissed
	GroupMemberSyncSilentSDKNotification     = 1521 // 拉人 Member Invited
	SignalingNotificationBegin               = 1600
	SignalingNotification                    = 1601
	SignalingNotificationEnd                 = 1649

	SuperGroupNotificationBegin  = 1650
	SuperGroupUpdateNotification = 1651
	SuperGroupNotificationEnd    = 1699

	ConversationPrivateChatNotification = 1701

	OrganizationChangedNotification = 1801

	WorkMomentNotificationBegin = 1900
	WorkMomentNotification      = 1901

	MomentNotification                      = 1950
	MomentNotification_Action_Comment       = 1951
	MomentNotification_Action_DeleteComment = 1952
	MomentNotification_Action_Like          = 1953
	MomentNotification_Action_CancelLike    = 1954
	MomentNotification_Action_CommentReply  = 1955
	MomentNotification_Action_DeleteMoment  = 1956

	OfficialAccountFollowUnfollowNotification = 1970

	NotificationEnd = 3000

	// status
	MsgNormal  = 1
	MsgDeleted = 4

	// MsgStatus
	MsgStatusDefault     = 0
	MsgStatusSending     = 1
	MsgStatusSendSuccess = 2
	MsgStatusSendFailed  = 3
	MsgStatusHasDeleted  = 4
	MsgStatusRevoked     = 5
	MsgStatusFiltered    = 6

	// MsgFrom
	UserMsgType = 100
	SysMsgType  = 200

	// SessionType
	SingleChatType       = 1
	GroupChatType        = 2
	SuperGroupChatType   = 3
	NotificationChatType = 4
	BroadcastChatType    = 5

	// token
	NormalToken  = 0
	InValidToken = 1
	KickedToken  = 2
	ExpiredToken = 3

	// MultiTerminalLogin
	// Full-end login, but the same end is mutually exclusive
	AllLoginButSameTermKick = 1
	// Only one of the endpoints can log in
	SingleTerminalLogin = 2
	// The web side can be online at the same time, and the other side can only log in at one end
	WebAndOther = 3
	// The PC side is mutually exclusive, and the mobile side is mutually exclusive, but the web side can be online at the same time
	PcMobileAndWeb = 4

	OnlineStatus  = "online"
	OfflineStatus = "offline"
	Registered    = "registered"
	UnRegistered  = "unregistered"

	// MsgReceiveOpt
	ReceiveMessage          = 0
	NotReceiveMessage       = 1
	ReceiveNotNotifyMessage = 2

	// OptionsKey
	IsHistory                  = "history"
	IsPersistent               = "persistent"
	IsOfflinePush              = "offlinePush"
	IsUnreadCount              = "unreadCount"
	IsConversationUpdate       = "conversationUpdate"
	IsSenderSync               = "senderSync"
	IsNotPrivate               = "notPrivate"
	IsSenderConversationUpdate = "senderConversationUpdate"
	IsSenderNotificationPush   = "senderNotificationPush"
	IsSyncToLocalDataBase      = "syncToLocalDataBase"

	// GroupStatus
	GroupOk              = 0
	GroupBanChat         = 1
	GroupStatusDismissed = 2
	GroupStatusMuted     = 3

	// GroupType
	NormalGroup     = 0
	SuperGroup      = 1
	DepartmentGroup = 2

	GroupBaned          = 3
	GroupBanPrivateChat = 4

	// UserJoinGroupSource
	JoinByAdmin = 1

	// Minio
	MinioDurationTimes = 3600

	// verificationCode used for
	VerificationCodeForRegister            = 1
	VerificationCodeForReset               = 2
	VerificationCodeForBindEmail           = 3
	VerificationCodeForBindPhone           = 4
	VerificationCodeForDeleteAccount       = 6
	VerificationCodeForRegisterSuffix      = "_forRegister"
	VerificationCodeForResetSuffix         = "_forReset"
	VerificationCodeForBindEmailSuffix     = "_forBind"
	VerificationCodeForBindPhoneSuffix     = "_forBindPhone"
	VerificationCodeForDeleteAccountSuffix = "_forDeleteAccount"

	// callbackCommand
	CallbackBeforeSendSingleMsgCommand = "callbackBeforeSendSingleMsgCommand"
	CallbackAfterSendSingleMsgCommand  = "callbackAfterSendSingleMsgCommand"
	CallbackBeforeSendGroupMsgCommand  = "callbackBeforeSendGroupMsgCommand"
	CallbackAfterSendGroupMsgCommand   = "callbackAfterSendGroupMsgCommand"
	CallbackWordFilterCommand          = "callbackWordFilterCommand"
	CallbackUserOnlineCommand          = "callbackUserOnlineCommand"
	CallbackUserOfflineCommand         = "callbackUserOfflineCommand"
	CallbackOfflinePushCommand         = "callbackOfflinePushCommand"
	// callback actionCode
	ActionAllow     = 0
	ActionForbidden = 1
	// callback callbackHandleCode
	CallbackHandleSuccess = 0
	CallbackHandleFailed  = 1

	// minioUpload
	OtherType = 1
	VideoType = 2
	ImageType = 3

	// workMoment permission
	WorkMomentPublic            = 0
	WorkMomentPrivate           = 1
	WorkMomentPermissionCanSee  = 2
	WorkMomentPermissionCantSee = 3

	// workMoment sdk notification type
	WorkMomentCommentNotification = 0
	WorkMomentLikeNotification    = 1
	WorkMomentAtUserNotification  = 2
)
const (
	SuperGroupTableName               = "local_super_groups"
	SuperGroupErrChatLogsTableNamePre = "local_sg_err_chat_logs_"
	SuperGroupChatLogsTableNamePre    = "local_sg_chat_logs_"
)
const (
	KeywordMatchOr  = 0
	KeywordMatchAnd = 1
)
const (
	AddConOrUpLatMsg          = 2
	UnreadCountSetZero        = 3
	IncrUnread                = 5
	TotalUnreadMessageChanged = 6
	UpdateFaceUrlAndNickName  = 7
	UpdateLatestMessageChange = 8
	ConChange                 = 9
	NewCon                    = 10

	HasRead = 1
	NotRead = 0

	IsFilter  = 1
	NotFilter = 0
)
const (
	AtAllString       = "AtAllTag"
	AtNormal          = 0
	AtMe              = 1
	AtAll             = 2
	AtAllAtMe         = 3
	GroupNotification = 4
)

var ContentType2PushContent = map[int64]string{
	Picture:   "[图片]",
	Voice:     "[语音]",
	Video:     "[视频]",
	File:      "[文件]",
	Text:      "你收到了一条文本消息",
	AtText:    "[有人@你]",
	GroupMsg:  "你收到一条群聊消息",
	Common:    "你收到一条新消息",
	SignalMsg: "音视频通话邀请",
}

const (
	FieldRecvMsgOpt    = 1
	FieldIsPinned      = 2
	FieldAttachedInfo  = 3
	FieldIsPrivateChat = 4
	FieldGroupAtType   = 5
	FieldIsNotInGroup  = 6
	FieldEx            = 7
	FieldUnread        = 8
)

const (
	AppOrdinaryUsers = 1
	AppAdmin         = 2

	GroupOrdinaryUsers = 1
	GroupOwner         = 2
	GroupAdmin         = 3

	GroupResponseAgree  = 1
	GroupResponseRefuse = -1

	FriendResponseAgree  = 1
	FriendResponseRefuse = -1

	Male   = 1
	Female = 2

	OpFromFrontend = 0
	OpFromAdmin    = 1
)

// Verification code type
const (
	SendMsgRegister      = 1
	SendMsgResetPassword = 2
	SendMsgDeleteAccount = 3
)

const (
	InviteCodeStateValid   = 1
	InviteCodeStateInvalid = 2
	InviteCodeStateDelete  = 3
)
const (
	InviteChannelCodeStateValid   = 1
	InviteChannelCodeStateInvalid = 2
	InviteChannelCodeStateDelete  = 3
)

const (
	UnreliableNotification    = 1
	ReliableNotificationNoMsg = 2
	ReliableNotificationMsg   = 3
)

const (
	OauthTypeEmail    = 0
	OauthTypeFaceBook = 1
	OauthTypeGoogle   = 2
	OauthTypeApple    = 3
	OauthTypeOfficial = 5
)

const (
	OauthTypeEmailStr    = "email"
	OauthTypeFaceBookStr = "facebook"
	OauthTypeGoogleStr   = "google"
	OauthTypeAppleStr    = "apple"
	OauthTypeOfficialStr = "official"
)

var OauthTypeId2Name = map[int]string{
	OauthTypeEmail:    OauthTypeEmailStr,
	OauthTypeFaceBook: OauthTypeFaceBookStr,
	OauthTypeGoogle:   OauthTypeGoogleStr,
	OauthTypeApple:    OauthTypeAppleStr,
	OauthTypeOfficial: OauthTypeOfficialStr,
}

var OauthTypeName2Id = map[string]int{
	OauthTypeEmailStr:    OauthTypeEmail,
	OauthTypeFaceBookStr: OauthTypeFaceBook,
	OauthTypeGoogleStr:   OauthTypeGoogle,
	OauthTypeAppleStr:    OauthTypeApple,
	OauthTypeOfficialStr: OauthTypeOfficial,
}

func OauthTypeIdToName(num int) string {
	return OauthTypeId2Name[num]
}
func OauthTypeNameToId(name string) int {
	return OauthTypeName2Id[name]
}

const (
	ConfigInviteCodeBaseLinkKey = "invite_code_base_link"

	ConfigInviteCodeIsOpenKey   = "invite_code_is_open"
	ConfigInviteCodeIsOpenTrue  = 1
	ConfigInviteCodeIsOpenFalse = 0

	ConfigInviteCodeIsLimitKey   = "invite_code_is_limit"
	ConfigInviteCodeIsLimitTrue  = 1
	ConfigInviteCodeIsLimitFalse = 0

	ConfigChannelCodeIsOpenKey    = "channel_code_is_open"
	ConfigChannelCodeIsOpenTrue   = 1
	ConfigChannelCodeIsOpenFalse  = 0
	ConfigChannelCodeIsLimitKey   = "channel_code_is_limit"
	ConfigChannelCodeIsLimitTrue  = 1
	ConfigChannelCodeIsLimitFalse = 0
)

const (
	AllowRegisterByUuid = "allow_register_by_uuid"
	AllowGuestLogin     = "allow_guest_login"
)

const (
	UserRegisterSourceTypeOfficial = 1 // 官方注册
	UserRegisterSourceTypeInvite   = 2 // 邀请注册
	UserRegisterSourceTypeChannel  = 3 // 渠道注册
)

const FriendAcceptTip = "You have successfully become friends, so start chatting"

func GroupIsBanChat(status int32) bool {
	if status == GroupBanChat {
		return true
	}
	return false
}

func GroupIsBanPrivateChat(status int32) bool {
	if status == GroupBanPrivateChat {
		return true
	}
	return false
}

const (
	AddFriendSourceSearch = "1"
	AddFriendSourceQr     = "2"
	AddFriendSourceGroup  = "3"
	AddFriendSourceCard   = "4"
)

const (
	PrivacyAddByPhone   = "privacy_add_by_phone"
	PrivacyAddByAccount = "privacy_add_by_account"
	PrivacyAddByEmail   = "privacy_add_by_email"
	PrivacySeeWooms     = "privacy_see_wooms"
	PrivacyPrivateChat  = "privacy_private_chat"

	PrivacyAddByGroup       = "privacy_add_by_group"
	PrivacyAddByQr          = "privacy_add_by_qr"
	PrivacyAddByContactCard = "privacy_add_by_contact_card"

	PrivacyStatusClose = "0"
	PrivacyStatusOpen  = "1"
)

const (
	InterestDefault       = -1
	InterestIsDefaultType = 2
)

const BigVersion = "v2"

const OpenImAdminCmsLog = "open_im_admin_cms.log"
const OpenImApiLog = "open_im_api.log"
const OpenImAuthLog = "open_im_Auth.log"
const OpenImCacheLog = "open_im_cache.log"
const OpenImCmsApiLog = "open_im_cms_api.log"
const OpenImConversationLog = "open_im_conversation.log"
const OpenImMomentsLog = "open_im_moments.log"
const OpenImNewsLog = "open_im_news.log"
const OpenImDemoLog = "open_im_demo.log"
const OpenImFriendLog = "open_im_friend.log"
const OpenImGroupLog = "open_im_group.log"
const OpenImMessageCmsLog = "open_im_message_cms.log"
const OpenImGameStoreLog = "open_im_game_store.log"
const OpenImShortVideoLog = "open_im_short_video.log"
const OpenImMsgLog = "open_im_msg.log"
const OpenImMsgGatewayLog = "open_im_msg_gateway.log"
const OpenImMsgTransferLog = "open_im_msg_transfer.log"
const OpenImOfficeLog = "open_im_office.log"
const OpenImOrganizationLog = "open_im_organization.log"
const OpenImPushLog = "open_im_push.log"
const OpenImStatisticsLog = "open_im_statistics.log"
const OpenImUserLog = "open_im_user.log"
const OpenImMistLog = "open_im_mist.log"
const OpenImLocalDataLog = "open_im_localdata.log"
const SQLiteLogFileName = "sqlite.log"
const MySQLLogFileName = "mysql.log"
const PressureTestLogFileName = "press_test.log"

const StatisticsTimeInterval = 60

const (
	SyncCreateGroup               = "SyncCreateGroup"
	SyncDeleteGroup               = "SyncDeleteGroup"
	SyncUpdateGroup               = "SyncUpdateGroup"
	SyncInvitedGroupMember        = "SyncInvitedGroupMember"
	SyncKickedGroupMember         = "SyncKickedGroupMember"
	SyncMuteGroupMember           = "SyncMuteGroupMember"
	SyncCancelMuteGroupMember     = "SyncCancelMuteGroupMember"
	SyncGroupMemberInfo           = "SyncGroupMemberInfo"
	SyncMuteGroup                 = "SyncMuteGroup"
	SyncCancelMuteGroup           = "SyncCancelMuteGroup"
	SyncConversation              = "SyncConversation"
	SyncGroupRequest              = "SyncGroupRequest"
	SyncAdminGroupRequest         = "SyncAdminGroupRequest"
	SyncFriendRequest             = "SyncFriendRequest"
	SyncSelfFriendRequest         = "SyncSelfFriendRequest"
	SyncFriendInfo                = "SyncFriendInfo"
	SyncAddBlackList              = "SyncAddBlackList"
	SyncDeleteBlackList           = "SyncDeleteBlackList"
	SyncUserInfo                  = "SyncUserInfo"
	SyncWelcomeMessageFromChannel = "SyncWelcomeMessageFromChannel"
)

// gorse
const (
	PositiveFeedbackLike     = "like"
	PositiveFeedbackFavorite = "favorite"
	PositiveFeedbackComment  = "comment"
	PositiveFeedbackShare    = "share"

	NegativeFeedbackRead  = "read"
	NegativeFeedbackBlock = "block"
)

// short video
const (
	ShortInt    = 1
	ShortString = "Short"
)

const (
	ShortVideoTypeReserve = 0
	ShortVideoTypeNormal  = 1
	ShortVideoTypeAudit   = 2
	ShortVideoTypePrivate = 3
	ShortVideoTypeFriend  = 4
	ShortVideoTypeDeleted = 5
)

const (
	ShortVideoCommentStatusReserve = 0
	ShortVideoCommentStatusNormal  = 1
	ShortVideoCommentStatusAudit   = 2
	ShortVideoCommentStatusDeleted = 3
)

const (
	ShortVideoNoticeTypeLikeShort    = 1 // 点赞短视频
	ShortVideoNoticeTypeLikeComment  = 2 // 点赞评论
	ShortVideoNoticeTypeReplyShort   = 3 // 回复短视频
	ShortVideoNoticeTypeReplyComment = 4 // 回复评论
	ShortVideoNoticeTypeFollowMe     = 5 // 关注我
	ShortVideoNoticeTypeNewPost      = 6 // 新发布

	ShortVideoNoticeStateUnread = 1
	ShortVideoNoticeStateRead   = 2
)

const (
	QrLoginStateReserve             = 0 // 未知状态
	QrLoginStateNormal              = 1 // 正常状态
	QrLoginStateWaitForConfirmation = 2 // 等待确认
	QrLoginStateConfirmed           = 3 // 已确认
	QrLoginStateExpired             = 4 // 已过期
)

// favorite
const (
	FavoriteContentTypeFile     = 1
	FavoriteContentTypeLink     = 2
	FavoriteContentTypeMedia    = 3
	FavoriteContentTypeAudio    = 4
	FavoriteContentTypeChats    = 5
	FavoriteContentTypeLocation = 6
)

const (
	FavoriteSourceTypeChatting    = 1
	FavoriteSourceTypeMoment      = 2
	FavoriteSourceTypeArticle     = 3
	FavoriteSourceTypeWeb         = 4
	FavoriteSourceTypeShortVideo  = 5
	FavoriteSourceCombineChatting = 6
)

// me page
const (
	//1: otc, 2:deposit, 3:withdraw, 4:exchange, 5:market, 6:earn'"
	MePageTypeOTC      = 1
	MePageTypeDeposit  = 2
	MePageTypeWithdraw = 3
	MePageTypeExchange = 4
	MePageTypeMarket   = 5
	MePageTypeEarn     = 6
	MePageGameStore    = 7
	MePageDiscover     = 8
)

// video record status
const (
	CommunicationRecordStatusWaiting   = 1
	CommunicationRecordStatusRecording = 2
	CommunicationRecordStatusFinished  = 3
)

const (
	RoomIDTypeString  = 0
	RoomIDTypeInteger = 1
)

const (
	PinGeneratedMessage = "Dear @nickname@, we received a request from your account to log in on https://www.Xlinkchat.com/deleteaccount. \n\nThis is your login code: @pincode@\n\nDon't give this code to anyone, even if they say they're from Xlink ! It can delete your account. \nWe never ask for it. Ignore this message if you didn't request the code."
)
