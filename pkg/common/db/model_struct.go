package db

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Register struct {
	Account  string `gorm:"column:account;primary_key;type:char(255)" json:"account"`
	Password string `gorm:"column:password;type:varchar(255)" json:"password"`
	Ex       string `gorm:"column:ex;size:1024" json:"ex"`
	UserID   string `gorm:"column:user_id;type:varchar(255)" json:"userID"`
}

func (Register) TableName() string {
	return "registers"
}

// message FriendInfo{
// string OwnerUserID = 1;
// string Remark = 2;
// int64 CreateTime = 3;
// UserInfo FriendUser = 4;
// int32 AddSource = 5;
// string OperatorUserID = 6;
// string Ex = 7;
// }
// open_im_sdk.FriendInfo(FriendUser) != imdb.Friend(FriendUserID)
type Friend struct {
	OwnerUserID    string    `gorm:"column:owner_user_id;primary_key;size:64"`
	FriendUserID   string    `gorm:"column:friend_user_id;primary_key;size:64"`
	Remark         string    `gorm:"column:remark;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreateTime     time.Time `gorm:"column:create_time"`
	AddSource      int32     `gorm:"column:add_source"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	Ex             string    `gorm:"column:ex;size:1024"`
	AccountStatus  int8      `gorm:"column:account_status;size:1;default:1;comment:'1 active 2 Deleted'"`
}

func (Friend) TableName() string {
	return "friends"
}

// message FriendRequest{
// string  FromUserID = 1;
// string ToUserID = 2;
// int32 HandleResult = 3;
// string ReqMsg = 4;
// int64 CreateTime = 5;
// string HandlerUserID = 6;
// string HandleMsg = 7;
// int64 HandleTime = 8;
// string Ex = 9;
// }
// open_im_sdk.FriendRequest(nickname, farce url ...) != imdb.FriendRequest
type FriendRequest struct {
	FromUserID    string     `gorm:"column:from_user_id;primary_key;size:64"`
	ToUserID      string     `gorm:"column:to_user_id;primary_key;size:64"`
	HandleResult  int32      `gorm:"column:handle_result"`
	ReqMsg        string     `gorm:"column:req_msg;size:255"`
	CreateTime    time.Time  `gorm:"column:create_time"`
	HandlerUserID string     `gorm:"column:handler_user_id;size:64"`
	HandleMsg     string     `gorm:"column:handle_msg;size:255"`
	HandleTime    *time.Time `gorm:"column:handle_time"`
	Ex            string     `gorm:"column:ex;size:1024"`
	AccountStatus int8       `gorm:"column:account_status;size:1;default:1;comment:'1 active 2 Deleted'"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}

//	message GroupInfo{
//	 string GroupID = 1;
//	 string GroupName = 2;
//	 string Notification = 3;
//	 string Introduction = 4;
//	 string FaceUrl = 5;
//	 string OwnerUserID = 6;
//	 uint32 MemberCount = 8;
//	 int64 CreateTime = 7;
//	 string Ex = 9;
//	 int32 Status = 10;
//	 string CreatorUserID = 11;
//	 int32 GroupType = 12;
//	}
//
// open_im_sdk.GroupInfo (OwnerUserID ,  MemberCount )> imdb.Group
type Group struct {
	// `json:"operationID" binding:"required"`
	// `protobuf:"bytes,1,opt,name=GroupID" json:"GroupID,omitempty"` `json:"operationID" binding:"required"`
	GroupID       string    `gorm:"column:group_id;primary_key;size:64" json:"groupID" binding:"required"`
	GroupName     string    `gorm:"column:name;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci" json:"groupName"`
	Notification  string    `gorm:"column:notification;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci" json:"notification"`
	Introduction  string    `gorm:"column:introduction;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci" json:"introduction"`
	FaceURL       string    `gorm:"column:face_url;size:255" json:"faceURL"`
	CreateTime    time.Time `gorm:"column:create_time"`
	Ex            string    `gorm:"column:ex" json:"ex;size:1024" json:"ex"`
	Status        int32     `gorm:"column:status;comment:'0 ok 1 banned 2 dismiss 3 muted'"`
	CreatorUserID string    `gorm:"column:creator_user_id;size:64"`
	GroupType     int32     `gorm:"column:group_type;comment:'0 normal group, 1 super group 2 department group'"`
	IsOpen        int32     `gorm:"column:is_open;comment:'0 private 1 public';default:0"`
	Remark        string    `gorm:"column:remark"`
	VideoStatus   int8      `gorm:"column:video_status;size:1;default:1;comment:'1 opened 2 banned'"`
	AudioStatus   int8      `gorm:"column:audio_status;size:1;default:1;comment:'1 opened 2 banned'"`
	CanAddFriend  int8      `gorm:"column:can_add_friend;size:1;default:1;comment:'1 can add 0 banned'"`
}

func (Group) TableName() string {
	return "groups"
}

// message GroupMemberFullInfo {
// string GroupID = 1 ;
// string UserID = 2 ;
// int32 roleLevel = 3;
// int64 JoinTime = 4;
// string NickName = 5;
// string FaceUrl = 6;
// int32 JoinSource = 8;
// string OperatorUserID = 9;
// string Ex = 10;
// int32 AppMangerLevel = 7; //if >0
// }  open_im_sdk.GroupMemberFullInfo(AppMangerLevel) > imdb.GroupMember
type GroupMember struct {
	GroupID        string    `gorm:"column:group_id;primary_key;size:64"`
	UserID         string    `gorm:"column:user_id;primary_key;size:64"`
	Nickname       string    `gorm:"column:nickname;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	FaceURL        string    `gorm:"column:user_group_face_url;size:255"`
	RoleLevel      int32     `gorm:"column:role_level;comment:'1 users 2 owner 3 admin'"`
	JoinTime       time.Time `gorm:"column:join_time"`
	JoinSource     int32     `gorm:"column:join_source"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	MuteEndTime    time.Time `gorm:"column:mute_end_time"`
	Ex             string    `gorm:"column:ex;size:1024"`
	DeleteTime     int32     `gorm:"column:delete_time;default:0"`
	DeletedBy      string    `gorm:"column:deleted_by"`
	UpdateVersion  int32     `gorm:"column:update_version;default:0"`
	Remark         string    `gorm:"column:remark;size:255"`
	VideoStatus    int8      `gorm:"column:video_status"`
	AudioStatus    int8      `gorm:"column:audio_status"`
	AccountStatus  int8      `gorm:"column:account_status;size:1;default:1;comment:'1 active 2 Deleted'"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

// message GroupRequest{
// string UserID = 1;
// string GroupID = 2;
// string HandleResult = 3;
// string ReqMsg = 4;
// string  HandleMsg = 5;
// int64 ReqTime = 6;
// string HandleUserID = 7;
// int64 HandleTime = 8;
// string Ex = 9;
// }open_im_sdk.GroupRequest == imdb.GroupRequest
type GroupRequest struct {
	UserID        string    `gorm:"column:user_id;primary_key;size:64"`
	GroupID       string    `gorm:"column:group_id;primary_key;size:64"`
	HandleResult  int32     `gorm:"column:handle_result"`
	ReqMsg        string    `gorm:"column:req_msg;size:1024"`
	HandledMsg    string    `gorm:"column:handle_msg;size:1024"`
	ReqTime       time.Time `gorm:"column:req_time"`
	HandleUserID  string    `gorm:"column:handle_user_id;size:64"`
	HandledTime   time.Time `gorm:"column:handle_time"`
	Ex            string    `gorm:"column:ex;size:1024"`
	AccountStatus int8      `gorm:"column:account_status;size:1;default:1;comment:'1 active 2 Deleted'"`
}

func (GroupRequest) TableName() string {
	return "group_requests"
}

// string UserID = 1;
// string Nickname = 2;
// string FaceUrl = 3;
// int32 Gender = 4;
// string PhoneNumber = 5;
// string Birth = 6;
// string Email = 7;
// string Ex = 8;
// int64 CreateTime = 9;
// int32 AppMangerLevel = 10;
// open_im_sdk.User == imdb.User
type User struct {
	ID               int       `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserID           string    `gorm:"column:user_id;size:64;uniqueIndex:user_id"`
	OfficialID       int64     `gorm:"column:official_id;Index:official_id;default:0"`
	Nickname         string    `gorm:"column:name;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	FaceURL          string    `gorm:"column:face_url;size:255"`
	Gender           int32     `gorm:"column:gender"`
	PhoneNumber      string    `gorm:"column:phone_number;size:32"`
	Birth            time.Time `gorm:"column:birth"`
	Email            string    `gorm:"column:email;size:64"`
	Ex               string    `gorm:"column:ex;size:1024"`
	Password         string    `gorm:"column:password;size:255"`
	Salt             string    `gorm:"column:salt"`
	Status           int       `gorm:"column:status;size:1;default:1;comment:'1 opened 2 banned'"`
	CreateUser       string    `gorm:"column:create_user"`
	CreateTime       int64     `gorm:"column:create_time"`
	UpdateUser       string    `gorm:"column:update_user"`
	UpdateTime       int64     `gorm:"column:update_time"`
	DeleteUser       string    `gorm:"column:delete_user"`
	DeleteTime       int64     `gorm:"column:delete_time"`
	AppMangerLevel   int32     `gorm:"column:app_manger_level"`
	GlobalRecvMsgOpt int32     `gorm:"column:global_recv_msg_opt"`
	SourceId         int64     `gorm:"column:source_id;Index:source_id;default:1;comment:'1:official 2:invite 3:channel'"`
	SourceCode       string    `gorm:"column:source_code;size:64;Index:source_code;default:''"`
	// SuperUserStatus  int       `gorm:"column:add_friend_status;size:1;default:2;comment:'1 allowed 2 disabled'"`
	CreateIp      string `gorm:"column:create_ip;size:32;default:''"`
	UpdateIp      string `gorm:"column:update_ip;size:32;default:''"`
	LoginIp       string `gorm:"column:login_ip;size:32;default:''"`
	LastLoginTime int64  `gorm:"column:last_login_time"`
	Uuid          string `gorm:"column:uuid;Index:uuid;size:64;default:''"`

	Remark          string `gorm:"column:remark"`
	LastLoginDevice int8   `gorm:"column:last_login_device;size:1;default:1;comment:'1 ios 2 android 3 desktop'"`
	VideoStatus     int8   `gorm:"column:video_status;size:1;default:1;comment:'1 opened 2 banned'"`
	AudioStatus     int8   `gorm:"column:audio_status;size:1;default:1;comment:'1 opened 2 banned'"`
	DeleteReason    string `gorm:"column:delete_reason"`
	AccountStatus   int8   `gorm:"column:account_status;size:1;default:1;comment:'1 active 2 Deleted'"`
}

func (User) TableName() string {
	return "users"
}

// open_im_sdk.BlackInfo(BlackUserInfo) != imdb.Black (BlockUserID)
type Black struct {
	OwnerUserID    string    `gorm:"column:owner_user_id;primary_key;size:64"`
	BlockUserID    string    `gorm:"column:block_user_id;primary_key;size:64"`
	CreateTime     time.Time `gorm:"column:create_time"`
	Remark         string    `gorm:"column:remark"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	UpdateTime     int64     `gorm:"column:update_time"`
	UpdateUser     string    `gorm:"column:update_user;size:64"`
	Ex             string    `gorm:"column:ex;size:1024"`
}

func (Black) TableName() string {
	return "blacks"
}

type BlackRes struct {
	Black
	BlockUserName     string `json:"black_user_name"`
	OwnerUserName     string `json:"owner_user_name"`
	OwnerProfilePhoto string `json:"owner_profile_photo"`
}

type BlackForMoment struct {
	OwnerUserID    string    `gorm:"column:owner_user_id;primary_key;size:64"`
	BlockUserID    string    `gorm:"column:block_user_id;primary_key;size:64"`
	CreateTime     time.Time `gorm:"column:create_time"`
	Remark         string    `gorm:"column:remark"`
	OperatorUserID string    `gorm:"column:operator_user_id;size:64"`
	UpdateTime     int64     `gorm:"column:update_time"`
	UpdateUser     string    `gorm:"column:update_user;size:64"`
	Ex             string    `gorm:"column:ex;size:1024"`
}

func (BlackForMoment) TableName() string {
	return "blacks_moments"
}

type ForMomentRes struct {
	BlackForMoment
	BlockUserName     string `json:"black_user_name"`
	OwnerUserName     string `json:"owner_user_name"`
	OwnerProfilePhoto string `json:"owner_profile_photo"`
}

// ContentType: 101 文本 Text,102 图片 Picture,103 语音 Voice,104 视频 Video,105 文件 File,106 @消息 @Text,
// 107 合并消息 Message Merge,108 名片 Name Card,109 位置 Location,110 自定义 Custom,111 撤回Revoke,
// 112 已读回执 Has Readed Receipt,113 输入中 Typing,114 引用 Quote,200 公共消息 Common,201 群消息 Group Message,
// 1000 开始通知 Notification Begin,1201 接受好友 Friend Approved,1202 拒绝好友 Friend Rejected,1203 好友申请 Friend Application,
// 1204 添加好友 Add Friend,1205 删除好友 Delete Friend,1206 好友备注 Friend Remark,1207 拉黑 Add Blacklist,1208 解除黑名单 Delete Blacklist,
// 1300 会话改变 Conversation Changed,1303 用户信息更新 User Info Updated,1400 OA通知 OA Notification,1501 建群 Create Group,
// 1502 设置群信息 Group Info Setting,1503 加群 Join Group,1504 退群 Exit Group,1505 接受入群 Group Accepted,1506 拒绝入群 Group Rejected,
// 1507 群主转让 Group Owner Transferred,1508 踢人 Member Kicked,1509 拉人 Member Invited,1510 进群 Member Enter,1511 群解散 Group Dismissed,
// 1512 成员禁言 Member Muted,1513 取消成员禁言 Member Cancel Muted,1514 群禁言 Group Muted,1515 取消群禁言 Group Cancel Muted,
// 1516 设置成员信息 Member Info Setting,1517 设置管理员 Admin Setting,1518 取消管理员 Cancel Admin,1519 发布群公告 Post Group Announcement
type ChatLog struct {
	ServerMsgID      string    `gorm:"column:server_msg_id;primary_key;type:char(64)" json:"serverMsgID"`
	ClientMsgID      string    `gorm:"column:client_msg_id;type:char(64);Index:client_msg_id" json:"clientMsgID"`
	SendID           string    `gorm:"column:send_id;type:char(64)" json:"sendID"`
	RecvID           string    `gorm:"column:recv_id;type:char(64)" json:"recvID"`
	SenderPlatformID int32     `gorm:"column:sender_platform_id" json:"senderPlatformID"`
	SenderNickname   string    `gorm:"column:sender_nick_name;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci" json:"senderNickname"`
	SenderFaceURL    string    `gorm:"column:sender_face_url;type:varchar(255)" json:"senderFaceURL"`
	SessionType      int32     `gorm:"column:session_type" json:"sessionType"`
	MsgFrom          int32     `gorm:"column:msg_from" json:"msgFrom"`
	ContentType      int32     `gorm:"column:content_type" json:"contentType"`
	Content          string    `gorm:"column:content;type:TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Status           int32     `gorm:"column:status" json:"status"`
	SendTime         time.Time `gorm:"column:send_time" json:"sendTime"`
	CreateTime       time.Time `gorm:"column:create_time" json:"createTime"`
	Ex               string    `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

func (ChatLog) TableName() string {
	return "chat_logs"
}

type BlackList struct {
	ID               int       `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId           string    `gorm:"column:uid"`
	BeginDisableTime time.Time `gorm:"column:begin_disable_time"`
	EndDisableTime   time.Time `gorm:"column:end_disable_time"`
}

func (BlackList) TableName() string {
	return "black_lists"
}

type Conversation struct {
	OwnerUserID      string `gorm:"column:owner_user_id;primary_key;type:char(128)" json:"OwnerUserID"`
	ConversationID   string `gorm:"column:conversation_id;primary_key;type:char(128)" json:"conversationID"`
	ConversationType int32  `gorm:"column:conversation_type" json:"conversationType"`
	UserID           string `gorm:"column:user_id;type:char(64)" json:"userID"`
	GroupID          string `gorm:"column:group_id;type:char(128)" json:"groupID"`
	RecvMsgOpt       int32  `gorm:"column:recv_msg_opt" json:"recvMsgOpt"`
	UnreadCount      int32  `gorm:"column:unread_count" json:"unreadCount"`
	DraftTextTime    int64  `gorm:"column:draft_text_time" json:"draftTextTime"`
	IsPinned         bool   `gorm:"column:is_pinned" json:"isPinned"`
	PinnedTime       int64  `gorm:"column:pinned_time" json:"pinnedTime"`
	IsPrivateChat    bool   `gorm:"column:is_private_chat" json:"isPrivateChat"`
	GroupAtType      int32  `gorm:"column:group_at_type" json:"groupAtType"`
	IsNotInGroup     bool   `gorm:"column:is_not_in_group" json:"isNotInGroup"`
	AttachedInfo     string `gorm:"column:attached_info;type:varchar(1024)" json:"attachedInfo"`
	Ex               string `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type Department struct {
	DepartmentID   string    `gorm:"column:department_id;primary_key;size:64" json:"departmentID"`
	FaceURL        string    `gorm:"column:face_url;size:255" json:"faceURL"`
	Name           string    `gorm:"column:name;size:256" json:"name" binding:"required"`
	ParentID       string    `gorm:"column:parent_id;size:64" json:"parentID" binding:"required"` // "0" or Real parent id
	Order          int32     `gorm:"column:order" json:"order" `                                  // 1, 2, ...
	DepartmentType int32     `gorm:"column:department_type" json:"departmentType"`                // 1, 2...
	RelatedGroupID string    `gorm:"column:related_group_id;size:64" json:"relatedGroupID"`
	CreateTime     time.Time `gorm:"column:create_time" json:"createTime"`
	Ex             string    `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

func (Department) TableName() string {
	return "departments"
}

type OrganizationUser struct {
	UserID      string    `gorm:"column:user_id;primary_key;size:64"`
	Nickname    string    `gorm:"column:nickname;size:256"`
	EnglishName string    `gorm:"column:english_name;size:256"`
	FaceURL     string    `gorm:"column:face_url;size:256"`
	Gender      int32     `gorm:"column:gender"` // 1 ,2
	Mobile      string    `gorm:"column:mobile;size:32"`
	Telephone   string    `gorm:"column:telephone;size:32"`
	Birth       time.Time `gorm:"column:birth"`
	Email       string    `gorm:"column:email;size:64"`
	CreateTime  time.Time `gorm:"column:create_time"`
	Ex          string    `gorm:"column:ex;size:1024"`
}

func (OrganizationUser) TableName() string {
	return "organization_users"
}

type DepartmentMember struct {
	UserID       string    `gorm:"column:user_id;primary_key;size:64"`
	DepartmentID string    `gorm:"column:department_id;primary_key;size:64"`
	Order        int32     `gorm:"column:order" json:"order"` // 1,2
	Position     string    `gorm:"column:position;size:256" json:"position"`
	Leader       int32     `gorm:"column:leader" json:"leader"` // -1, 1
	Status       int32     `gorm:"column:status" json:"status"` // -1, 1
	CreateTime   time.Time `gorm:"column:create_time"`
	Ex           string    `gorm:"column:ex;type:varchar(1024)" json:"ex"`
}

func (DepartmentMember) TableName() string {
	return "department_members"
}

type AppVersion struct {
	Version     string `gorm:"column:version;size:64" json:"version"`
	Type        int    `gorm:"column:type;primary_key" json:"type"`
	UpdateTime  int    `gorm:"column:update_time" json:"update_time"`
	ForceUpdate bool   `gorm:"column:force_update" json:"force_update"`
	FileName    string `gorm:"column:file_name" json:"file_name"`
	YamlName    string `gorm:"column:yaml_name" json:"yaml_name"`
	UpdateLog   string `gorm:"column:update_log" json:"update_log"`
}

func (AppVersion) TableName() string {
	return "app_version"
}

type DiscoverUrl struct {
	ID         uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Url        string `gorm:"column:url;size:255"`
	Status     int    `gorm:"column:status;size:1;comment: 0 close, 1 open"`
	PlatformId int    `gorm:"column:platform_id;size:11"`
	CreateTime int64  `gorm:"column:create_time"`
	CreateUser string `gorm:"column:create_user;size:64"`
	UpdateTime int64  `gorm:"column:update_time"`
	UpdateUser string `gorm:"column:update_user;size:64"`
	DeleteTime int64  `gorm:"column:delete_time"`
	DeleteUser string `gorm:"column:delete_user;size:64"`
}

func (DiscoverUrl) TableName() string {
	return "discover_url"
}

type NewAppVersion struct {
	ID          uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Version     string `gorm:"column:version;size:64"`
	Type        int    `gorm:"column:type;size:1"`
	Status      int    `gorm:"column:status;size:1"`
	Isforce     int    `gorm:"column:isforce;size:1"`
	Title       string `gorm:"column:title;size:64"`
	DownloadUrl string `gorm:"column:download_url;size:255"`
	Content     string `gorm:"column:content;size:255"`
	CreateTime  int64  `gorm:"column:create_time"`
	CreateUser  string `gorm:"column:create_user;size:64"`
	UpdateTime  int64  `gorm:"column:update_time"`
	UpdateUser  string `gorm:"column:update_user;size:64"`
	DeleteTime  int64  `gorm:"column:delete_time"`
	DeleteUser  string `gorm:"column:delete_user;size:64"`
}

func (NewAppVersion) TableName() string {
	return "new_app_version"
}

type InviteCodeLog struct {
	ID       uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Code     string `gorm:"column:code;size:8;Index:index_code;comment:invite_channel_code"`
	Timezone string `gorm:"column:timezone;size:8;Index:index_timezone"`
	Mobile   string `gorm:"column:mobile;size:26;Index:index_mobile;comment:mobile brand"`
	// The specific model of the phone was not available
	// Device      string `gorm:"column:device;size:32;Index:index_device;comment:device"`
	Os          string `gorm:"column:os;size:16;Index:index_os;comment:os"`
	Version     string `gorm:"column:version;size:16;Index:index_version;comment:mobile version"`
	Webkit      string `gorm:"column:webkit;size:16;Index:index_webkit;comment:webkit version"`
	ScreenWidth int    `gorm:"column:screen_width;Index:index_screen_width;comment:mobile screen width"`
	Language    string `gorm:"column:language;size:16;Index:index_language;comment:moblie system language"`

	Ip         string `gorm:"column:ip;size:64;Index:index_ip"`
	CreateTime int64  `gorm:"column:create_time;Index:index_create_time"`
}

func (InviteCodeLog) TableName() string {
	return "invite_code_log"
}

type InviteCode struct {
	ID       uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId   string `gorm:"column:user_id;size:64;uniqueIndex:index_user_id;comment:user id"`
	Code     string `gorm:"column:code;size:8;uniqueIndex:index_code;comment:invite_code"`
	Greeting string `gorm:"column:greeting;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;comment:greeting"`
	Note     string `gorm:"column:note;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;comment:note"`
	State    int    `gorm:"column:state;size:1;Index:index_state;default:1;comment:state 1:valid 2:invalid 3:deleted"`
}

func (InviteCode) TableName() string {
	return "invite_code"
}

type InviteChannelCode struct {
	ID             uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Code           string `gorm:"column:code;size:8;uniqueIndex:index_code;comment:invite_channel_code"`
	OperatorUserId string `gorm:"column:operator_user_id;size:64;Index:index_user_id;comment:operator user id"`
	FriendId       string `gorm:"column:friend_id;size:256;comment:add friend id"`
	GroupId        string `gorm:"column:group_id;size:256;comment:add group id"`
	Greeting       string `gorm:"column:greeting;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;comment:greeting"`
	Note           string `gorm:"column:note;type:char(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;comment:note"`
	State          int    `gorm:"column:state;size:1;Index:index_state;default:1;comment:state 1:valid 2:invalid 3:deleted"`
	SourceId       int64  `gorm:"column:source_id;Index:source_id;default:3;comment:'1:official 3:channel'"`
}

func (InviteChannelCode) TableName() string {
	return "invite_channel_code"
}

type InviteCodeRelation struct {
	ID          uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	OwnerUserId string `gorm:"column:owner_user_id;size:64;Index:index_owner_user_id;comment:owner user id"`
	InviteUser  string `gorm:"column:invite_user;size:64;Index:index_invite_user;comment:invite user id"`
	Type        int    `gorm:"column:type;size:1;Index:index_type;comment:type 1:invite"`
	CreateTime  int64  `gorm:"column:create_time;Index:index_create_time"`
}

func (InviteCodeRelation) TableName() string {
	return "invite_code_relation"
}

type Config struct {
	ID         uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Name       string `gorm:"column:name;size:64;uniqueIndex:name"`
	Value      string `gorm:"column:value;size:255"`
	Note       string `gorm:"column:note;size:255"`
	UpdateTime int64  `gorm:"column:update_time"`
}

func (Config) TableName() string {
	return "config"
}

type AdminUser struct {
	ID                int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserID            string `gorm:"column:user_id;index:user_id,unique;size:64"`
	Nickname          string `gorm:"column:nick_name;size:255"`
	Name              string `gorm:"column:name;size:255"`
	PhoneNumber       string `gorm:"column:phone_number;size:32"`
	Password          string `gorm:"column:password;size:255"`
	Google2fSecretKey string `gorm:"column:google_2f_secret_key;size:255"`
	Salt              string `gorm:"column:salt"`
	Status            int64  `gorm:"column:status;size:1;default:1;comment:'1 opened 2 banned'"`
	Role              int64  `gorm:"column:role;default:0;comment:'role id from admin_role'"`
	CreateUser        string `gorm:"column:create_user"`
	CreateTime        int64  `gorm:"column:create_time"`
	UpdateUser        string `gorm:"column:update_user"`
	UpdateTime        int64  `gorm:"column:update_time;default:0"`
	DeleteUser        string `gorm:"column:delete_user"`
	DeleteTime        int64  `gorm:"column:delete_time;Index:delete_time;default:0"`
	AppMangerLevel    int32  `gorm:"column:app_manger_level"`
	GlobalRecvMsgOpt  int32  `gorm:"column:global_recv_msg_opt"`
	TwoFactorEnabled  int64  `gorm:"column:two_factor_enabled;default:2"`
	User2FAuthEnable  int64  `gorm:"column:user_two_factor_control_status;default:2"`
	IPRangeStart      string `gorm:"column:ip_range_start;default:0.0.0.0"`
	IPRangeEnd        string `gorm:"column:ip_range_end;default:0.0.0.0"`
	LastLoginIP       string `gorm:"column:last_login_ip;default:0.0.0.0"`
	LastLoginTime     int64  `gorm:"column:last_login_time;default:0"`
	Remarks           string `gorm:"column:remarks;nullable:false"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

type AdminAPIs struct {
	ApiID      int64  `gorm:"column:api_id;primary_key;autoIncrement"`
	ApiName    string `gorm:"column:api_name;size:255"`
	ApiPath    string `gorm:"column:api_path;size:255"`
	CreateUser string `gorm:"column:create_user"`
	Status     int    `gorm:"column:status;size:1;default:1;comment:'1 opened 2 deleted'"`
	CreateTime int64  `gorm:"column:create_time"`
	UpdateUser string `gorm:"column:update_user"`
	UpdateTime int64  `gorm:"column:update_time"`
	DeleteUser string `gorm:"column:delete_user"`
	DeleteTime int64  `gorm:"column:delete_time"`
}

func (AdminAPIs) TableName() string {
	return "admin_apis"
}

type AdminPages struct {
	PageID       int64  `gorm:"column:page_id;primary_key;autoIncrement"`
	PageName     string `gorm:"column:page_name;size:255"`
	PagePath     string `gorm:"column:page_path;size:255"`
	CreateUser   string `gorm:"column:create_user"`
	CreateTime   int64  `gorm:"column:create_time"`
	Status       int    `gorm:"column:status;size:1;default:1;comment:'1 opened 2 deleted'"`
	UpdateUser   string `gorm:"column:update_user"`
	UpdateTime   int64  `gorm:"column:update_time"`
	DeleteUser   string `gorm:"column:delete_user"`
	DeleteTime   int64  `gorm:"column:delete_time"`
	FatherPageID int64  `gorm:"column:father_page_id"`
	IsMenu       int64  `gorm:"column:is_menu"`
	SortPriority int64  `gorm:"column:sort_priority"`
	AdminAPIsIDs string `gorm:"column:admin_apis_ids;size:255"`
	IsButton     int64  `gorm:"column:is_button"`
}

func (AdminPages) TableName() string {
	return "admin_pages"
}

// type AdminActions struct {
// 	AdminActionID   int64  `gorm:"column:admin_action_id;primary_key;autoIncrement:1"`
// 	AdminActionName string `gorm:"column:admin_action_name;size:255"`
// 	AdminAPIsIDs    string `gorm:"column:admin_apis_ids;size:255"`
// 	AdminPagesIDs   string `gorm:"column:admin_pages_ids;size:255"`
// 	Status          int    `gorm:"column:admin_role_status;size:1;default:1;comment:'1 opened 2 deleted'"`
// 	CreateUser      string `gorm:"column:create_user"`
// 	CreateTime      int64  `gorm:"column:create_time"`
// 	UpdateUser      string `gorm:"column:update_user"`
// 	UpdateTime      int64  `gorm:"column:update_time"`
// 	DeleteUser      string `gorm:"column:delete_user"`
// 	DeleteTime      int64  `gorm:"column:delete_time"`
// }

// func (AdminActions) TableName() string {
// 	return "admin_actions"
// }

type AdminRole struct {
	AdminRoleID   int64  `gorm:"column:admin_role_id;primary_key;autoIncrement"`
	AdminRoleName string `gorm:"column:admin_role_name;size:255"`
	// AdminActionsIDs string `gorm:"column:admin_action_ids;size:255"`
	AdminAPIsIDs         string `gorm:"column:admin_apis_ids;type:text"`
	AdminPagesIDs        string `gorm:"column:admin_pages_ids;type:text"`
	Status               int    `gorm:"column:status;size:1;default:1;comment:'1 opened 2 deleted'"`
	CreateUser           string `gorm:"column:create_user"`
	CreateTime           int64  `gorm:"column:create_time"`
	UpdateUser           string `gorm:"column:update_user"`
	UpdateTime           int64  `gorm:"column:update_time"`
	DeleteUser           string `gorm:"column:delete_user"`
	DeleteTime           int64  `gorm:"column:delete_time"`
	AdminRoleDiscription string `gorm:"column:admin_role_discription;nullable:false"`
	AdminRoleRemarks     string `gorm:"column:admin_role_remarks;nullable:false"`
}

func (AdminRole) TableName() string {
	return "admin_role"
}

type GroupUpdatesVersion struct {
	GroupID       string    `gorm:"column:group_id;size:64;primary_key;Index:group_id"`
	VersionNumber int64     `gorm:"column:version_number;default:0"`
	UpdateTime    time.Time `gorm:"column:update_time"`
}

func (GroupUpdatesVersion) TableName() string {
	return "group_updates_version"
}

type Moment struct {
	CreatorID                 string             `json:"creatorID" bson:"creator_id"`
	MomentID                  primitive.ObjectID `json:"momentID" bson:"moment_id"`
	MContentText              string             `json:"mContentText" bson:"m_content_text"`
	MContentImagesArray       string             `json:"mContentImagesArray" bson:"m_content_images_array"`
	MContentVideosArray       string             `json:"mContentVideosArray" bson:"m_content_videos_array"`
	MContentThumbnilArray     string             `json:"mContentThumbnilArray" bson:"m_content_thumbnil_array"`
	MLikesCount               int32              `json:"mLikesCount" bson:"m_likes_count"`
	MCommentsCount            int32              `json:"mCommentsCount" bson:"m_comments_count"`
	MRepostCount              int64              `json:"mRepostCount" bson:"m_repost_count"`
	MCreateTime               int64              `json:"mCreateTime" bson:"m_create_time"`
	MUpdateTime               int64              `json:"mUpdateTime" bson:"m_update_time"`
	MomentType                int32              `json:"momentType" bson:"moment_type"`
	OrignalCreatorID          string             `json:"orignalCreatorID" bson:"orignal_creator_id"`
	OriginalCreatorName       string             `json:"originalCreatorName" bson:"original_creator_name"`
	OriginalCreatorProfileImg string             `json:"originalCreatorProfileImg" bson:"original_creator_profile_img"`
	OrignalID                 primitive.ObjectID `json:"orignalID" bson:"orignal_id"`
	IsReposted                bool               `json:"isReposted" bson:"is_reposted"`
	DeleteTime                int64              `json:"deleteTime" bson:"delete_time"`
	DeletedBy                 string             `json:"deletedBy" bson:"deleted_by"`
	Status                    int8               `json:"status" bson:"status"`
	Privacy                   int32              `json:"privacy" bson:"privacy"`
	UserID                    string             `json:"userID" bson:"user_id"`
	UserName                  string             `json:"userName" bson:"user_name"`
	UserProfileImg            string             `json:"userProfileImg" bson:"user_profile_img"`
	CommentCtl                int32              `json:"commentCtl" bson:"comment_ctl"`
	ArticleID                 int64              `json:"articleID" bson:"article_id"`
	WoomFileID                string             `json:"woomFileID" bson:"woom_file_id"`
}
type MomentSQL struct {
	CreatorID                 string `gorm:"creator_id"`
	MomentID                  string `gorm:"moment_id;primary_key"`
	MContentText              string `gorm:"m_content_text"`
	MContentImagesArray       string `gorm:"m_content_images_array;type:LONGTEXT"`
	MContentVideosArray       string `gorm:"m_content_videos_array;type:LONGTEXT"`
	MContentThumbnilArray     string `gorm:"m_content_thumbnil_array"`
	MLikesCount               int32  `gorm:"m_likes_count;default:0"`
	MCommentsCount            int32  `gorm:"m_comments_count;default:0"`
	MRepostCount              int64  `gorm:"m_repost_count;default:0"`
	MCreateTime               int64  `gorm:"m_create_time;default:0"`
	MUpdateTime               int64  `gorm:"m_update_time;default:0"`
	MomentType                int32  `gorm:"moment_type;default:0; comment: 0:moment 1:article"`
	OrignalCreatorID          string `gorm:"orignal_creator_id"`
	OriginalCreatorName       string `gorm:"original_creator_name"`
	OriginalCreatorProfileImg string `gorm:"original_creator_profile_img"`
	OrignalID                 string `gorm:"orignal_id"`
	IsReposted                bool   `gorm:"is_reposted;default:1"`
	DeleteTime                int64  `gorm:"delete_time;default:0"`
	DeletedBy                 string `gorm:"deleted_by"`
	Status                    int8   `gorm:"status;default:1"`
	Privacy                   int32  `gorm:"privacy;default:1"`
	UserID                    string `gorm:"user_id"`
	UserName                  string `gorm:"user_name"`
	UserProfileImg            string `gorm:"user_profile_img"`
	CommentCtl                int32  `gorm:"comment_ctl;default:1; 1:on 2:off"`
	ArticleID                 int64  `gorm:"article_id;default:0; 0:Not An Article"`
	WoomFileID                string `gorm:"woom_file_id"`
}

func (Moment) TableName() string {
	return "moments"
}

func (MomentSQL) TableName() string {
	return "moments"
}

type MomentsRes struct {
	MomentID              string
	MCreateTime           int64
	MContentText          string
	MContentImagesArray   string
	MContentVideosArray   string
	MContentThumbnilArray string
	OriginalCreatorName   string
	OriginalCreatorID     string
	OrignalID             string
	Privacy               int32
	UserID                string
	UserName              string
	Status                int8
	ArticleID             int64
	Interests             []InterestTypeRes
}

type MomentDetailRes struct {
	MomentSQL
	LastLoginIp string
}

type MomentLike struct {
	MomentID       primitive.ObjectID `json:"momentID" bson:"moment_id"`
	UserID         string             `json:"userID" bson:"user_id"`
	UserName       string             `json:"userName" bson:"user_name"`
	UserProfileImg string             `json:"userProfileImg" bson:"user_profile_img"`
	CreateBy       string             `json:"createBy" bson:"created_by"`
	CreateTime     int64              `json:"createTime" bson:"created_time"`
	UpdateBy       string             `json:"updateBy" bson:"updated_by"`
	UpdatedTime    int64              `json:"updatedTime" bson:"updated_time"`
	DeletedBy      string             `json:"deletedBy" bson:"deleted_by"`
	DeleteTime     int64              `json:"deleteTime" bson:"delete_time"`
	Status         int8               `json:"status" bson:"status"`
}
type MomentLikeSQL struct {
	ID             int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	MomentID       string `gorm:"moment_id"`
	UserID         string `gorm:"user_id"`
	UserName       string `gorm:"user_name"`
	UserProfileImg string `gorm:"user_profile_img"`
	CreateBy       string `gorm:"created_by"`
	CreateTime     int64  `gorm:"created_time;default:0"`
	UpdateBy       string `gorm:"updated_by"`
	UpdatedTime    int64  `gorm:"updated_time;default:0"`
	DeletedBy      string `gorm:"deleted_by"`
	DeleteTime     int64  `gorm:"delete_time;default:0"`
	Status         int8   `gorm:"status;default:1;comment:unhide:1, hide:2"`
}

func (MomentLike) TableName() string {
	return "moments_like"
}
func (MomentLikeSQL) TableName() string {
	return "moments_like"
}

type MomentLikeRes struct {
	MomentID            string
	Account             string
	AccountNickname     string
	MCreateTime         int64
	MContentText        string
	MContentImagesArray string
	MContentVideosArray string
	UserID              string
	UserName            string
	CreateTime          int64
	Privacy             int32
	Status              int8
}

type MomentComment struct {
	MomentID         primitive.ObjectID `json:"momentID" bson:"moment_id"`
	CommentID        primitive.ObjectID `json:"commentID" bson:"comment_id"`
	UserID           string             `json:"userID" bson:"user_id"`
	UserName         string             `json:"userName" bson:"user_name"`
	UserProfileImg   string             `json:"userProfileImg" bson:"user_profile_img"`
	CommentContent   string             `json:"commentContent" bson:"comment_content"`
	ReplyCommentID   primitive.ObjectID `json:"reply_comment_id" bson:"reply_comment_id"`
	CommentParentID  primitive.ObjectID `json:"commentParentID" bson:"comment_parent_id"`
	CPUserID         string             `json:"cpUserID" bson:"cp_user_id"`
	CPUserName       string             `json:"cpUserName" bson:"cp_user_name"`
	CPUserProfileImg string             `json:"cpUserProfileImg" bson:"cp_user_profile_img"`
	CreateBy         string             `json:"createBy" bson:"created_by"`
	CreateTime       int64              `json:"createTime" bson:"created_time"`
	UpdateBy         string             `json:"updateBy" bson:"updated_by"`
	UpdatedTime      int64              `json:"updatedTime" bson:"updated_time"`
	DeletedBy        string             `json:"deletedBy" bson:"deleted_by"`
	DeleteTime       int64              `json:"deleteTime" bson:"delete_time"`
	Status           int8               `json:"status" bson:"status"`
	CommentReplies   int64              `json:"commentReplies" bson:"comment_replies"`
	LikeCounts       int64              `json:"likeCounts" bson:"like_counts"`
	AccountStatus    int8               `json:"accountStatus" bson:"account_status"`
}
type MomentCommentSQL struct {
	ID               int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	MomentID         string `gorm:"moment_id"`
	CommentID        string `gorm:"comment_id"`
	UserID           string `gorm:"user_id"`
	UserName         string `gorm:"user_name"`
	UserProfileImg   string `gorm:"user_profile_img"`
	CommentContent   string `gorm:"comment_content"`
	ReplyCommentID   string `gorm:"reply_comment_id"`
	CommentParentID  string `gorm:"comment_parent_id"`
	CPUserID         string `gorm:"cp_user_id"`
	CPUserName       string `gorm:"cp_user_name"`
	CPUserProfileImg string `gorm:"cp_user_profile_img"`
	CreateBy         string `gorm:"created_by"`
	CreateTime       int64  `gorm:"created_time;default:0"`
	UpdateBy         string `gorm:"updated_by"`
	UpdatedTime      int64  `gorm:"updated_time;default:0"`
	DeletedBy        string `gorm:"deleted_by"`
	DeleteTime       int64  `gorm:"delete_time;default:0"`
	Status           int8   `gorm:"status;default:1;comment: 1 no hide, 2 hide"`
	CommentReplies   int64  `gorm:"comment_replies"`
	LikeCounts       int64  `gorm:"like_counts"`
}

func (MomentComment) TableName() string {
	return "moments_comments"
}
func (MomentCommentSQL) TableName() string {
	return "moments_comments"
}

type MomentCommentRes struct {
	MomentID            string
	CommentID           string
	PublishAccount      string
	PublishName         string
	MCreateTime         int64
	MContentText        string
	MContentImagesArray string
	MContentVideosArray string
	UserID              string
	UserName            string
	CommentContent      string
	CommentUsername     string
	CreateBy            string
	CreateTime          int64
	CommentReplies      int64
	LikeCounts          int64
	Privacy             int32
	Status              int8
	ReplyCommentId      string
	CommentParentId     string
	CommentedUseID      string
	CommentedUserName   string
}

type MomentCommentLike struct {
	MomentID       primitive.ObjectID `json:"momentID" bson:"moment_id"`
	CommentID      primitive.ObjectID `json:"commentID" bson:"comment_id"`
	UserID         string             `json:"userID" bson:"user_id"`
	UserName       string             `json:"userName" bson:"user_name"`
	UserProfileImg string             `json:"userProfileImg" bson:"user_profile_img"`
	CreateBy       string             `json:"createBy" bson:"created_by"`
	CreateTime     int64              `json:"createTime" bson:"created_time"`
	UpdateBy       string             `json:"updateBy" bson:"updated_by"`
	UpdatedTime    int64              `json:"updatedTime" bson:"updated_time"`
	DeletedBy      string             `json:"deletedBy" bson:"deleted_by"`
	DeleteTime     int64              `json:"deleteTime" bson:"delete_time"`
	Status         int8               `json:"status" bson:"status"`
}
type MomentCommentLikeSQL struct {
	ID             int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	MomentID       string `gorm:"moment_id"`
	CommentID      string `gorm:"comment_id"`
	UserID         string `gorm:"user_id"`
	UserName       string `gorm:"user_name"`
	UserProfileImg string `gorm:"user_profile_img"`
	CreateBy       string `gorm:"created_by"`
	CreateTime     int64  `gorm:"created_time;default:0"`
	UpdateBy       string `gorm:"updated_by"`
	UpdatedTime    int64  `gorm:"updated_time;default:0"`
	DeletedBy      string `gorm:"deleted_by"`
	DeleteTime     int64  `gorm:"delete_time;default:0"`
	Status         int8   `gorm:"status;default:1"`
}

type Official struct {
	Id                  int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserID              string `gorm:"column:user_id;size:64;Index:user_id"`
	Type                int8   `gorm:"column:type;comment:1 personal 2 business"`
	IdType              int8   `gorm:"column:id_type;comment:1 passport 2 id_card 3 licence"`
	IdName              string `gorm:"column:id_name;size:255"`
	IdNumber            string `gorm:"column:id_number;size:255"`
	FaceURL             string `gorm:"column:face_url;size:255"`
	Nickname            string `gorm:"column:nickname;size:255"`
	Bio                 string `gorm:"column:bio;size:512"`
	Interests           string `gorm:"column:interests"`
	CountryCode         string `gorm:"column:country_code;type=char(2)"`
	ProcessStatus       int8   `gorm:"column:process_status;default:0;comment:0 pending 1 verified 2 failed"`
	ProcessBy           string `gorm:"column:process_by;size:64"`
	ProcessFeedback     string `gorm:"column:process_feedback;size:255"`
	CreateTime          int64  `gorm:"column:create_time"`
	UpdateTime          int64  `gorm:"column:update_time"`
	ProcessTime         int64  `gorm:"column:process_time"`
	InitialNickname     string `gorm:"column:initial_nickname;size:255"`
	NicknameUpdateCount int8   `gorm:"column:nickname_update_count;default:0"`
	NicknameUpdateTime  int64  `gorm:"column:nickname_update_time"`
	FollowCounts        int64  `gorm:"column:follow_counts;default:0"`
	LikeCounts          int64  `gorm:"column:like_counts;default:0"`
	PostCounts          int64  `gorm:"column:post_counts;default:0"`
	IsSystem            int8   `gorm:"column:is_system"`
	Status              int64  `gorm:"column:status;comment:'1 enable, 2 disable'"`
	DeleteBy            string `gorm:"column:delete_by"`
	DeleteTime          int64  `gorm:"column:delete_time"`
	LastLoginIp         string `gorm:"column:last_login_ip"`
	LastLoginTime       string `gorm:"column:last_login_time"`
	LastActivityCity    string `gorm:"column:last_activity_city"`
	LastActivityCountry string `gorm:"column:last_activity_country"`
}

func (Official) TableName() string {
	return "official"
}

type OfficialAnalytics struct {
	ID               int   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	OfficialID       int64 `gorm:"column:official_id;index:official_analytics,unique;not null"`
	Time             int64 `gorm:"column:time;index:official_analytics,unique;not null"`
	Gender           int32 `gorm:"column:gender;index:official_analytics,unique;not null"`
	FollowCounts     int64 `gorm:"column:follow_counts;default:0"`
	LikeCounts       int64 `gorm:"column:like_counts;default:0"`
	CommentCounts    int64 `gorm:"column:comment_counts;default:0"`
	ReadCounts       int64 `gorm:"column:read_counts;default:0"`
	UniqueReadCounts int64 `gorm:"column:unique_read_counts;default:0"`
}

func (OfficialAnalytics) TableName() string {
	return "official_analytics"
}

type OfficialInterest struct {
	ID             int   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	OfficialID     int64 `gorm:"column:official_id;Index:official_id"`
	InterestTypeID int64 `gorm:"column:interest_type_id"`
}

func (OfficialInterest) TableName() string {
	return "official_interest"
}

type GetOfficialRes struct {
	Official
	Interests []*InterestTypeRes
}

type OauthClient struct {
	ID            int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	ThirdId       string `gorm:"column:third_id;size:64;uniqueIndex:third_id;comment:1facebook 2google 3apple 4wallet 5official"`
	Type          int8   `gorm:"column:type;size:8;uniqueIndex:third_id;"`
	ThirdUserName string `gorm:"column:third_user_name;size:64;default:'';comment:third user name"`
	UserId        string `gorm:"column:user_id;size:64;"`
}

func (OauthClient) TableName() string {
	return "oauth_client"
}

type PrivacySetting struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId     string `gorm:"column:user_id;size:64;Index:user_id;uniqueIndex:user_id_setting_key"`
	SettingKey string `gorm:"column:setting_key;size:64;uniqueIndex:user_id_setting_key"`
	SettingVal string `gorm:"column:setting_val;default:'';comment:0 close 1 open"`
}

func (PrivacySetting) TableName() string {
	return "privacy_setting"
}

type Contact struct {
	Id     int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId string `gorm:"column:user_id;size:64;Index:user_id"`
	Phone  string `gorm:"column:phone;size:64;Index:phone"`
}

func (Contact) TableName() string {
	return "contact"
}

type ContactExclude struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	FromUserId string `gorm:"column:user_id;size:64;Index:user_id;uniqueIndex:from_to"`
	ToUserId   string `gorm:"column:to_user_id;size:64;uniqueIndex:from_to"`
}

func (ContactExclude) TableName() string {
	return "contact_exclude"
}

type InterestType struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	IsDefault  int8   `gorm:"column:is_default;default:1;comment:'1 personal 2 default'"`
	Status     int8   `gorm:"column:status;size:8;default:1;comment:'1 opened 2 deleted'"`
	Remark     string `gorm:"column:remark;size:200'"`
	CreateUser string `gorm:"column:create_user"`
	CreateTime int64  `gorm:"column:create_time"`
	UpdateUser string `gorm:"column:update_user"`
	UpdateTime int64  `gorm:"column:update_time"`
	DeletedBy  string `gorm:"column:deleted_by"`
	DeleteTime int64  `gorm:"column:delete_time;default:0;"`
}

func (InterestType) TableName() string {
	return "interest_type"
}

type InterestLanguage struct {
	Id           int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	InterestId   int64  `gorm:"column:interest_id;"`
	LanguageType string `gorm:"column:language_type;en cn"`
	Name         string `gorm:"column:name;"`
}

func (InterestLanguage) TableName() string {
	return "interest_language"
}

type InterestTypeRes struct {
	Id         int64  `json:"id"`
	Status     int8   `json:"status"`
	Remark     string `json:"remark"`
	CreateUser string `json:"create_user"`
	CreateTime int64  `json:"create_time"`
	UpdateUser string `json:"update_user"`
	UpdateTime int64  `json:"update_time"`
	DeleteTime int64  `json:"delete_time"`
	IsDefault  int8   `json:"is_default"`
	Name       []InterestLanguage
}

type InterestUser struct {
	Id             int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId         string `gorm:"column:user_id;size:64;Index:user_id"`
	InterestId     int64  `gorm:"column:interest_id;Index:interest_id"`
	UpdateTime     int64  `gorm:"column:update_time"`
	UserCreateTime string `gorm:"user_create_time"`
}

func (InterestUser) TableName() string {
	return "interest_user"
}

type UserInterests struct {
	Username  string             `json:"username"`
	UserID    string             `json:"user_id"`
	Type      int32              `json:"type"`
	Interests []*InterestTypeRes `json:"interests"`
}

type InterestGroup struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	GroupId    string `gorm:"column:group_id;size:64;Index:group_id"`
	InterestId int64  `gorm:"column:interest_id;Index:interest_id"`
	UpdateTime int64  `gorm:"column:update_time"`
}

func (InterestGroup) TableName() string {
	return "interest_group"
}

type GroupInterests struct {
	GroupID   string            `json:"group_id"`
	GroupName string            `json:"group_name"`
	GroupType int32             `json:"group_type"`
	Interests []InterestTypeRes `json:"interests"`
}

type InterestGroupExclude struct {
	Id      int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId  string `gorm:"column:user_id;size:64;Index:user_id"`
	GroupId string `gorm:"column:group_id;size:64"`
}

func (InterestGroupExclude) TableName() string {
	return "interest_group_exclude"
}

type GroupHeat struct {
	Id        int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	GroupId   string `gorm:"column:group_id;size:64;uniqueIndex:group_id_month"`
	Month     string `gorm:"column:month;size:8;uniqueIndex:group_id_month"`
	MsgCount  int64  `gorm:"column:msg_count"`
	UserCount int64  `gorm:"column:user_count"`
	Heat      int64  `gorm:"column:heat"`
}

func (GroupHeat) TableName() string {
	return "group_heat"
}

type OfficialFollow struct {
	OfficialID int64  `bson:"official_id"`
	UserID     string `bson:"user_id"`
	FollowTime int64  `bson:"follow_time"`
	DeletedBy  string `bson:"deleted_by"`
	DeleteTime int64  `bson:"delete_time"`
	BlockedBy  string `bson:"blocked_by"`
	BlockTime  int64  `bson:"block_time"`
	Muted      bool   `bson:"muted"`
	Enabled    bool   `bson:"enabled"`
}

func (OfficialFollow) TableName() string {
	return "official_follow"
}

type OfficialFollowSQL struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	OfficialID int64  `gorm:"column:official_id;uniqueIndex:user_official_follow"`
	UserID     string `gorm:"column:user_id;size:64;uniqueIndex:user_official_follow"`
	FollowTime int64  `gorm:"column:follow_time;default:0"`
	DeletedBy  string `gorm:"column:deleted_by;size:64"`
	DeleteTime int64  `gorm:"column:delete_time;default:0"`
	BlockedBy  string `gorm:"column:blocked_by;size:64"`
	BlockTime  int64  `gorm:"column:block_time;default:0"`
	Muted      bool   `bson:"column:muted;default:true"`
	Enabled    bool   `bson:"column:enabled;default:true"`
}

func (OfficialFollowSQL) TableName() string {
	return "official_follow"
}

type OfficialFollowRes struct {
	OfficialFollowSQL
	OfficialName string
	Username     string
}

type OperationLog struct {
	Operator   string `bson:"operator"`
	Action     string `bson:"action"`
	Payload    string `bson:"payload"`
	OperatorIP string `bson:"operatorip"`
	Executee   string `bson:"executee"`
	CreateTime int64  `bson:"createtime"`
}

type ArticleLike struct {
	ArticleID  int64  `bson:"article_id"`
	UserID     string `bson:"user_id"`
	CreatedBy  string `bson:"created_by"`
	CreateTime int64  `bson:"create_time"`
	UpdatedBy  string `bson:"updated_by"`
	UpdateTime int64  `bson:"update_time"`
	DeletedBy  string `bson:"deleted_by"`
	DeleteTime int64  `bson:"delete_time"`
	Status     int32  `bson:"status"`
}

type ArticleLikeSQL struct {
	ID         int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	ArticleID  int64  `gorm:"article_id"`
	UserID     string `gorm:"user_id"`
	CreatedBy  string `gorm:"created_by"`
	CreateTime int64  `gorm:"create_time;default:0"`
	UpdatedBy  string `gorm:"updated_by"`
	UpdateTime int64  `gorm:"update_time;default:0"`
	DeletedBy  string `gorm:"deleted_by"`
	DeleteTime int64  `gorm:"delete_time;default:0"`
	Status     int32  `gorm:"status;default:1"`
}

func (ArticleLike) TableName() string {
	return "article_like"
}

func (ArticleLikeSQL) TableName() string {
	return "article_like"
}

type ArticleRead struct {
	ArticleID  int64  `bson:"article_id"`
	UserID     string `bson:"user_id"`
	Status     int8   `bson:"status"`
	CreateTime int64  `bson:"create_time"`
}

func (ArticleRead) TableName() string {
	return "article_read"
}

type ArticleReadSQL struct {
	ID         int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	ArticleID  int64  `gorm:"article_id"`
	UserID     string `gorm:"user_id;size:64"`
	Status     int8   `bson:"status;default:1"`
	CreateTime int64  `gorm:"create_time;default:0"`
}

func (ArticleReadSQL) TableName() string {
	return "article_read"
}

type ArticleLikesRes struct {
	ArticleLikeSQL
	OfficialName string
	OfficialType int8
	LastLoginIp  string
	ArticleTitle string
	PostTime     int64
	CoverPhoto   string
}

type Article struct {
	ArticleID          int64  `bson:"article_id"`
	CoverPhoto         string `bson:"cover_photo"`
	Title              string `bson:"title"`
	Content            string `bson:"content"`
	TextContent        string `bson:"text_content"`
	OfficialID         int64  `bson:"official_id"`
	OfficialName       string `bson:"official_name"`
	OfficialProfileImg string `bson:"official_profile_img"`
	CreatedBy          string `bson:"created_by"`
	CreateTime         int64  `bson:"create_time"`
	UpdatedBy          string `bson:"updated_by"`
	UpdateTime         int64  `bson:"update_time"`
	DeletedBy          string `bson:"deleted_by"`
	DeleteTime         int64  `bson:"delete_time"`
	Status             int32  `bson:"status"`
	Privacy            int32  `bson:"privacy"`
	ReadCounts         int64  `bson:"read_counts"`
	UniqueReadCounts   int64  `bson:"unique_read_counts"`
	CommentCounts      int64  `bson:"comment_counts"`
	LikeCounts         int64  `bson:"like_counts"`
	RepostCounts       int64  `bson:"repost_counts"`
}

type ArticleSQL struct {
	ArticleID          int64  `gorm:"column:article_id;primary_key;autoIncrement"`
	CoverPhoto         string `gorm:"column:cover_photo;size=1024"`
	Title              string `gorm:"column:title;size=255"`
	Content            string `gorm:"column:content"`
	TextContent        string `gorm:"column:text_content"`
	OfficialID         int64  `gorm:"column:official_id;Index:official_id"`
	OfficialName       string `gorm:"column:official_name"`
	OfficialProfileImg string `gorm:"column:official_profile_img"`
	CreatedBy          string `gorm:"column:created_by;Index:created_by"`
	CreateTime         int64  `gorm:"column:create_time;default:0"`
	UpdatedBy          string `gorm:"column:updated_by;Index:updated_by"`
	UpdateTime         int64  `gorm:"column:update_time;default:0"`
	DeletedBy          string `gorm:"column:deleted_by;Index:deleted_by"`
	DeleteTime         int64  `gorm:"column:delete_time;default:0"`
	Status             int32  `gorm:"column:status;default:1; comment:'1 unhide 2 hide'"`
	Privacy            int32  `gorm:"column:privacy;default:1; comment:'1 private 2 public'"`
	ReadCounts         int64  `gorm:"column:read_counts;default:0"`
	UniqueReadCounts   int64  `gorm:"column:unique_read_counts;default:0"`
	CommentCounts      int64  `gorm:"column:comment_counts;default:0"`
	LikeCounts         int64  `gorm:"column:like_counts;default:0"`
	RepostCounts       int64  `gorm:"column:repost_counts;default:0"`
}

func (Article) TableName() string {
	return "article"
}

func (ArticleSQL) TableName() string {
	return "article"
}

type ArticleRes struct {
	ArticleSQL
	OfficialType   int32
	LastLoginIp    string
	LastLoginTime  string
	OfficialStatus int64
}

type ArticlePostRes struct {
	MomentId      string
	ShareUser     string
	OfficialType  int32
	ArticleTitle  string
	OriginalUser  string
	CommentCounts int64
	LikeCounts    int64
	ShareTime     int64
	LastLoginIp   string
	Privacy       int32
	CoverPhoto    string
	DeletedBy     string
	DeleteTime    int64
	ArticleId     int64
}

type ArticleComment struct {
	CommentID       int64  `bson:"comment_id"`
	ParentCommentID int64  `bson:"parent_comment_id"`
	ArticleID       int64  `bson:"article_id"`
	OfficialID      int64  `bson:"official_id"`
	UserID          string `bson:"user_id"`
	ReplyOfficialID int64  `bson:"reply_official_id"`
	ReplyUserID     string `bson:"reply_user_id"`
	ReplyCounts     int64  `bson:"reply_counts"`
	LikeCounts      int64  `bson:"like_counts"`
	Content         string `bson:"content"`
	CreatedBy       string `bson:"created_by"`
	CreateTime      int64  `bson:"created_time"`
	UpdatedBy       string `bson:"updated_by"`
	UpdateTime      int64  `bson:"update_time"`
	DeletedBy       string `bson:"deleted_by"`
	DeleteTime      int64  `bson:"delete_time"`
	Status          int32  `bson:"status"`
}

func (ArticleComment) TableName() string {
	return "article_comment"
}

type ArticleCommentSQL struct {
	CommentID       int64  `gorm:"comment_id;primary_key;autoIncrement"`
	ParentCommentID int64  `gorm:"parent_comment_id;Index:parent_comment_id"`
	ArticleID       int64  `gorm:"article_id;Index:article_id"`
	OfficialID      int64  `gorm:"official_id;Index:official_id"`
	ReplyOfficialID int64  `gorm:"reply_official_id;Index:reply_official_id"`
	ReplyUserID     string `gorm:"reply_user_id;Index:reply_user_id"`
	UserID          string `gorm:"user_id;size=64;index:user_id"`
	ReplyCounts     int64  `gorm:"reply_counts;default:0"`
	LikeCounts      int64  `gorm:"like_counts;default:0"`
	Content         string `gorm:"content"`
	CreatedBy       string `gorm:"created_by;size=64;Index:created_by"`
	CreateTime      int64  `gorm:"create_time;default:0"`
	UpdatedBy       string `gorm:"updated_by;size=64;Index:updated_by"`
	UpdateTime      int64  `gorm:"update_time;default:0"`
	DeletedBy       string `gorm:"deleted_by;size=64"`
	DeleteTime      int64  `gorm:"delete_time;default:0;Index:deleted_by"`
	Status          int32  `gorm:"status;default:1"`
}

func (ArticleCommentSQL) TableName() string {
	return "article_comment"
}

type ArticleCommentRes struct {
	ArticleCommentSQL
	ArticleTitle  string
	OfficialType  int8
	OfficialName  string
	LastLoginTime string
	LastLoginIp   string
	CoverPhoto    string
	PostTime      int64
	CommentLikes  int64
	UserName      string
}

type ArticleCommentLike struct {
	CommentID  int64  `bson:"comment_id"`
	UserID     string `bson:"user_id"`
	OfficialID int64  `bson:"official_id"`
	CreatedBy  string `bson:"created_by"`
	CreateTime int64  `bson:"create_time"`
	UpdatedBy  string `bson:"updated_by"`
	UpdateTime int64  `bson:"update_time"`
	DeletedBy  string `bson:"deleted_by"`
	DeleteTime int64  `bson:"delete_time"`
	Status     int32  `bson:"status"`
}

type ArticleCommentLikeSQL struct {
	ID         int    `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	CommentID  int64  `gorm:"comment_id"`
	UserID     string `gorm:"user_id;size=64"`
	OfficialID int64  `gorm:"official_id"`
	CreatedBy  string `gorm:"created_by;size=64"`
	CreateTime int64  `gorm:"create_time;default:0"`
	UpdatedBy  string `gorm:"updated_by;size=64"`
	UpdateTime int64  `gorm:"update_time;default:0"`
	DeletedBy  string `gorm:"deleted_by;size=64"`
	DeleteTime int64  `gorm:"delete_time;default:0"`
	Status     int32  `gorm:"status;default:1"`
}

func (ArticleCommentLike) TableName() string {
	return "article_comment_like"
}

func (ArticleCommentLikeSQL) TableName() string {
	return "article_comment_like"
}

type Favorites struct {
	FavoriteId       primitive.ObjectID `bson:"favorite_id"`
	UserID           string             `bson:"user_id"`
	ExKeywords       string             `bson:"ex_keywords"`
	ContentType      int32              `bson:"content_type"`
	Content          string             `bson:"content"`
	ObjName          string             `bson:"obj_name"`
	ThumbnailObjName string             `bson:"thumbnail_obj_name"`
	MediasObjName    string             `bson:"medias_obj_name"`
	ContentID        string             `bson:"content_id"`
	ContentCreatorID string             `bson:"content_creator_id"`
	ContentGroupID   string             `bson:"original_group_id"`
	FileSize         int64              `bson:"file_size"`
	SourceType       int32              `bson:"source_type"`
	PublishTime      int64              `bson:"publish_time"`
	Remark           string             `bson:"remark"`
	UpdateBy         string             `bson:"update_by"`
	UpdateTime       int64              `bson:"update_time"`
	CreateBy         string             `bson:"create_by"`
	CreateTime       int64              `bson:"create_time"`
	DeletedBy        string             `bson:"deleted_by"`
	DeleteTime       int64              `bson:"delete_time"`
}

type FavoritesSQL struct {
	FavoriteId       string `gorm:"favorite_id;primary_key; comment: favorite id"`
	UserID           string `gorm:"user_id; comment: favorite user id"`
	ExKeywords       string `gorm:"ex_keywords; comment: for searching, maybe filename, article title..."`
	ContentType      int32  `gorm:"content_type;default: 1; comment: 1 file, 2 links, 3 media, 4 audio, 5 chats, 6 location"`
	Content          string `gorm:"content; comment: chat info, notes content, link url, file/img/audio/video url"`
	ObjName          string `gorm:"obj_name; comment: if file, media or audio, obj name is uploaded obj name"`
	ThumbnailObjName string `gorm:"thumbnail_obj_name"`
	MediasObjName    string `gorm:"medias_obj_name"`
	ContentID        string `gorm:"content_id;default :''; comment: chat history id, article id or others id"`
	ContentCreatorID string `gorm:"content_creator_id;default:''; comment: official or user id"`
	ContentGroupID   string `gorm:"content_group_id; comment: content group id"`
	FileSize         int64  `gorm:"file_size"`
	SourceType       int32  `gorm:"source_type; default:1; comment: 1 chatting, 2 moment, 3 article, 4 web, 5 short video, 6 combine chatting"`
	PublishTime      int64  `gorm:"publish_time"`
	Remark           string `gorm:"remark"`
	UpdateBy         string `gorm:"update_by"`
	UpdateTime       int64  `gorm:"update_time"`
	CreateBy         string `gorm:"create_by"`
	CreateTime       int64  `gorm:"create_time;default:0"`
	DeletedBy        string `gorm:"deleted_by"`
	DeleteTime       int64  `gorm:"delete_time;default:0"`
}

func (Favorites) TableName() string {
	return "favorites"
}

func (FavoritesSQL) TableName() string {
	return "favorites"
}

type FavoriteRes struct {
	FavoriteId         string
	UserID             string
	UserName           string
	ContentType        int32
	Content            string
	CreateTime         int64
	PublishUser        string
	PublishTime        int64
	Remark             string
	EditUser           string
	UpdateBy           string
	UpdateTime         int64
	ExKeywords         string
	ContentCreatorName string
	FileSize           int64
	SourceType         int32
	CreateBy           string
}

type VideoAudioCommunicationRecord struct {
	CommunicationID    int64  `gorm:"communication_id;primary_key;autoIncrement"`
	Originator         string `gorm:"originator"`
	RoomID             string `gorm:"room_id"`
	RoomIDType         uint64 `gorm:"room_id_type"`
	RecordUserID       string `gorm:"record_user_id"`
	OriginatorPlatform int32  `gorm:"originator_platform"`
	GroupID            string `gorm:"group_id"`
	Status             int32  `gorm:"status;default:1;comment:'1.waiting for the record, 2.on the phone, 3.end, 4 interrupt'"`
	RecordStatus       int32  `gorm:"record_status; comment:'1: waiting, 2: recording, 3: finished'"`
	ErrCode            int64  `gorm:"err_code; comment:'status code'"`
	ErrMsgEN           string `gorm:"err_msg_en; comment:'status message en'"`
	ErrMsgCN           string `gorm:"err_msg_cn; comment:'status message cn'"`
	RecordTaskID       string `gorm:"record_task_id"`
	RecordRequestID    string `gorm:"record_request_id"`
	Duration           int64  `gorm:"duration; comment:'unit is second'"`
	StartTime          int64  `gorm:"start_time"`
	EndTime            int64  `gorm:"end_time"`
	RecordURL          string `gorm:"record_url;default:'';comment:'record video url'"`
	ChatType           int8   `gorm:"chat_type;comment:'1: video, 2: audio'"`
	Supporter          string `gorm:"supporter"`
	Remark             string `gorm:"remark"`
	UpdateBy           string `gorm:"update_by"`
	UpdateTime         int64  `gorm:"update_time"`
	DeleteBy           string `gorm:"delete_by"`
	DeleteTime         int64  `gorm:"delete_time"`
}

func (VideoAudioCommunicationRecord) TableName() string {
	return "video_audio_communication_record"
}

type CommunicationGroupMember struct {
	Id              int64  `gorm:"id;primary_key;autoIncrement"`
	CommunicationID int64  `gorm:"communication_id"`
	MemberID        string `gorm:"member_id"`
}

func (CommunicationGroupMember) TableName() string {
	return "communication_group_member"
}

type VideoAudioCommunicationRecordRes struct {
	VideoAudioCommunicationRecord
	OriginatorName string
	MemberIDs      []string
	MemberIDNames  []string
}

type GameCategories struct {
	Id             int64  `gorm:"id;primary_key;autoIncrement"`
	Status         int8   `gorm:"column:status;size:8;default:1;comment:'1 opened 2 deleted'"`
	Remark         string `gorm:"column:remark;size:200;comment:'maybe someday will be used'"`
	Priority       int32  `gorm:"column:priority;default:0; comment:'for searching sort'"`
	CategoryNameEN string `gorm:"column:category_name_en;size:50"`
	CategoryNameCN string `gorm:"column:category_name_cn;size:50"`
	CreateUser     string `gorm:"column:create_user"`
	CreateTime     int64  `gorm:"column:create_time"`
	UpdateUser     string `gorm:"column:update_user"`
	UpdateTime     int64  `gorm:"column:update_time"`
	DeletedBy      string `gorm:"column:deleted_by"`
	DeleteTime     int64  `gorm:"column:delete_time;default:0;"`
}

func (GameCategories) TableName() string {
	return "game_categories"
}

type GameLink struct {
	Id          int64  `gorm:"id;primary_key;autoIncrement"`
	GameCode    string `gorm:"game_code;size:200;"`
	Platform    int32  `gorm:"platform;comment:'1 H5 2 Android 3 IOS'"`
	PackageURL  string `gorm:"package_url;package url for download package, as like apk"`
	PlayURL     string `gorm:"play_url; play url"`
	PackageName string `gorm:"package_name"`
	PackageSize int64  `gorm:"package_size"`
	CreateUser  string `gorm:"create_user"`
	CreateTime  int64  `gorm:"create_time"`
	UpdateUser  string `gorm:"update_user"`
	UpdateTime  int64  `gorm:"update_time"`
	DeletedBy   string `gorm:"deleted_by"`
	DeleteTime  int64  `gorm:"column:delete_time;default:0;"`
}

func (GameLink) TableName() string {
	return "game_link"
}

type Game struct {
	Id              int64  `gorm:"id;primary_key;autoIncrement"`
	GameCode        string `gorm:"game_code;size:100;not null"`
	CoverImg        string `gorm:"cover_img"`
	HorizontalCover string `gorm:"horizontal_cover;comment:'game horizontal cover, url json array'"`
	Classification  string `gorm:"classification;comment:'1 home banner 2 today recommendation, json array'"`
	Priority        int32  `gorm:"priority"`
	Hot             int32  `gorm:"hot;comment:'hot, 1 to 5'"`
	State           int8   `gorm:"state; comment:'game status 1 on 2 off'"`
	Remark          string `gorm:"remark; comment:'maybe someday need this.'"`
	Categories      string `gorm:"categories; comment: 'json for categories id array list'"`
	GameNameEN      string `gorm:"game_name_en;not null; default:'';"`
	GameNameCN      string `gorm:"game_name_cn;not null; default:'';"`
	DescriptionEN   string `gorm:"column:description_en;size:500;"`
	DescriptionCN   string `gorm:"column:description_cn;size:500;"`
	Publisher       string `gorm:"publisher;size:100;"`
	ClickCounts     int64  `gorm:"click_counts;default:0"`
	PlayCounts      int64  `gorm:"play_counts;default:0"`
	UpdateBy        string `gorm:"update_by"`
	UpdateTime      int64  `gorm:"update_time"`
	CreateTime      int64  `gorm:"create_time;default:0"`
	DeletedBy       string `gorm:"deleted_by"`
	DeleteTime      int64  `gorm:"delete_time;default:0"`
}

func (Game) TableName() string {
	return "game"
}

type GamePlayHistory struct {
	Id         int64  `gorm:"id;primary_key;autoIncrement"`
	UserID     string `gorm:"user_id"`
	GameCode   string `gorm:"game_code"`
	CreateTime int64  `gorm:"create_time;default:0"`
	DeletedBy  string `gorm:"deleted_by"`
	DeleteTime int64  `gorm:"delete_time;default:0"`
}

func (GamePlayHistory) TableName() string {
	return "game_play_history"
}

type GameFavorites struct {
	Id         int64  `gorm:"id;primary_key;autoIncrement"`
	UserID     string `gorm:"user_id"`
	GameCode   string `gorm:"game_code"`
	CreateTime int64  `gorm:"create_time;default:0"`
	DeletedBy  string `gorm:"deleted_by"`
	DeleteTime int64  `gorm:"delete_time;default:0"`
}

func (GameFavorites) TableName() string {
	return "game_favorites"
}

type MePageURL struct {
	ID         uint   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Url        string `gorm:"column:url;size:255"`
	Type       int64  `gorm:"column:type;comment:'1: otc, 2:deposit, 3:withdraw, 4:exchange, 5:market, 6:earn'"`
	Language   string `gorm:"column:language"`
	Status     int    `gorm:"column:status;size:1;default:1"`
	CreateTime int64  `gorm:"column:create_time"`
	CreateUser string `gorm:"column:create_user;size:64"`
	UpdateTime int64  `gorm:"column:update_time"`
	UpdateUser string `gorm:"column:update_user;size:64"`
	DeleteTime int64  `gorm:"column:delete_time;default:0"`
	DeleteUser string `gorm:"column:delete_user;size:64"`
}

func (MePageURL) TableName() string {
	return "me_page_url"
}

type ShortVideo struct {
	Id             int64   `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	FileId         string  `gorm:"column:file_id;type:varchar(32);uniqueIndex:file_id;not null;default:''"`
	Name           string  `gorm:"column:name;type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:''"`
	Desc           string  `gorm:"column:desc;type:varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:''"`
	InterestId     string  `gorm:"column:interest_id;type:varchar(64);not null;default:''"`
	ClassId        int8    `gorm:"column:class_id;type:tinyint(3);not null;default:0"`
	ClassName      string  `gorm:"column:class_name;type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:''"`
	CoverUrl       string  `gorm:"column:cover_url;type:varchar(256);default:''"`
	MediaUrl       string  `gorm:"column:media_url;type:varchar(256);not null;default:''"`
	Type           string  `gorm:"column:type;type:varchar(8);not null;default:''"`
	Size           int64   `gorm:"column:size;type:bigint(20);not null;default:0"`
	Height         int64   `gorm:"column:height;type:int(8);not null;default:0"`
	Width          int64   `gorm:"column:width;type:int(8);not null;default:0"`
	Duration       float64 `gorm:"column:duration;type:decimal(10,4);not null;default:0"`
	Json           string  `gorm:"column:json;type:text;comment:'json格式的数据'"`
	LikeNum        int64   `gorm:"column:like_num;type:int(11);not null;default:0"`
	CommentNum     int64   `gorm:"column:comment_num;type:int(11);not null;default:0"`
	ReplyNum       int64   `gorm:"column:reply_num;type:int(11);not null;default:0"`
	CommentLikeNum int64   `gorm:"column:comment_like_num;type:int(11);not null;default:0"`
	ForwardNum     int64   `gorm:"column:forward_num;type:int(11);not null;default:0"`
	UserId         string  `gorm:"column:user_id;type:varchar(32);Index:user_id;not null;default:''"`
	CreateTime     int64   `gorm:"column:create_time;autoCreateTime"`
	UpdateTime     int64   `gorm:"column:update_time;autoUpdateTime"`
	Remark         string  `gorm:"column:remark;type:varchar(256)CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;not null;default:''"`
	Status         int8    `gorm:"column:status;type:tinyint(3);not null;default:1;comment:'1:正常 2:审核 3:私有 4:好友 5:删除'"`
}

func (ShortVideo) TableName() string {
	return "short_video"
}

type ShortVideoUserCount struct {
	Id                   int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId               string `gorm:"column:user_id;type:varchar(32);uniqueIndex:user_id;not null;default:''"`
	WorkNum              int64  `gorm:"column:work_num;type:int(11);not null;default:0"`
	LikeNum              int64  `gorm:"column:like_num;type:int(11);not null;default:0;comment:'送出去的视频点赞数'"`
	HarvestedLikesNumber int64  `gorm:"column:harvested_likes_number;type:int(11);not null;default:0;comment:'已经收获的点赞数'"`
	CommentNum           int64  `gorm:"column:comment_num;type:int(11);not null;default:0"`
	CommentLikeNum       int64  `gorm:"column:comment_like_num;type:int(11);not null;default:0"`
	FansNum              int64  `gorm:"column:fans_num;type:int(11);not null;default:0"`
	FollowNum            int64  `gorm:"column:follow_num;type:int(11);not null;default:0"`
	CreateTime           int64  `gorm:"column:create_time;autoCreateTime"`
	UpdateTime           int64  `gorm:"column:update_time;autoUpdateTime"`
}

func (ShortVideoUserCount) TableName() string {
	return "short_video_user_count"
}

type ShortVideoLike struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	FileId     string `gorm:"column:file_id;type:varchar(32);uniqueIndex:file_id_user_id;not null"`
	UserId     string `gorm:"column:user_id;type:varchar(32);uniqueIndex:file_id_user_id;not null"`
	CreateTime int64  `gorm:"column:create_time;autoCreateTime"`
	Remark     string `gorm:"column:remark"`
}

func (ShortVideoLike) TableName() string {
	return "short_video_like"
}

type ShortVideoComment struct {
	CommentId         int64  `gorm:"column:comment_id;primary_key;autoIncrement"`
	FileId            string `gorm:"column:file_id;type:varchar(32);not null;Index:file_id_user_id"`
	UserID            string `gorm:"column:user_id;type:varchar(32);not null;Index:file_id_user_id"`
	ParentId          int64  `gorm:"column:parent_id;type:bigint(20);not null;default:0;Index:parent_id"`
	Level0CommentId   int64  `gorm:"column:level0_comment_id;type:bigint(20);not null;default:0;Index:level0_comment_id"`
	Level1CommentId   int64  `gorm:"column:level1_comment_id;type:bigint(20);not null;default:0;Index:level1_comment_id"`
	LevelId           int64  `gorm:"column:level_id;type:bigint(20);not null;default:0"`
	ReplyTo           string `gorm:"column:reply_to;type:varchar(32);not null;default:''"`
	Content           string `gorm:"column:content;type:varchar(256);not null"`
	CommentReplyCount int64  `gorm:"column:comment_reply_count;type:bigint(20);not null;default:0"`
	CommentLikeCount  int64  `gorm:"column:comment_like_count;type:bigint(20);not null;default:0"`
	TotalReplyCount   int64  `gorm:"column:total_reply_count;type:bigint(20);not null;default:0"`
	CreateTime        int64  `gorm:"column:create_time;autoCreateTime"`
	UpdateTime        int64  `gorm:"column:update_time;autoUpdateTime"`
	Status            int8   `gorm:"column:status;type:tinyint(3);not null;default:1;comment:'1:正常 2:审核 3:删除'"`
	Remark            string `gorm:"column:remark;type:varchar(256);not null;default:''"`
}

func (ShortVideoComment) TableName() string {
	return "short_video_comment"
}

type ShortVideoCommentResult struct {
	FileId            string `json:"file_id"`
	CommentId         int64  `json:"comment_id"`
	UserId            string `json:"user_id"`
	CreateTime        int64  `json:"create_time"`
	Content           string `json:"content"`
	CommentReplyCount int64  `json:"comment_reply_count"`
	CommentLikeCount  int64  `json:"comment_like_count"`
	Remark            string `json:"remark"`
	CoverUrl          string `json:"cover_url"`
	MediaUrl          string `json:"media_url"`
	PostUserId        string `json:"post_user_id"`
	Desc              string `json:"desc"`
	Status            int32  `json:"status"`
}

type ShortVideoCommentRepliesRes struct {
	FileId              string `json:"file_id"`
	PublishUserID       string `json:"publisher_id"`
	PublishUser         string `json:"publish_user"`
	ShortVideoStatus    int32  `json:"short_video_status"`
	Content             string `json:"content"`
	CoverUrl            string `json:"cover_url"`
	MediaUrl            string `json:"media_url"`
	Size                int64  `json:"size"`
	Height              int64  `json:"height"`
	Width               int64  `json:"width"`
	CommentId           int64  `json:"comment_id"`
	CommentContent      string `json:"comment_content"`
	CommentStatus       int64  `json:"comment_status"`
	ReplyCommentId      int64  `json:"reply_comment_id"`
	ReplyUserName       string `json:"reply_user_name"`
	ReplyUserID         string `json:"reply_user_id"`
	ReplyTime           int64  `json:"reply_time"`
	ReplyCommentContent string `json:"reply_comment_content"`
	LikeCount           int64  `json:"like_count"`
	CommentCount        int64  `json:"comment_count"`
	Remark              string `json:"remark"`
	Status              int64  `json:"status"`
}

type ShortVideoCommentLike struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId     string `gorm:"column:user_id;type:varchar(32);uniqueIndex:user_id_comment_id;not null"`
	CommentId  int64  `gorm:"column:comment_id;type:bigint(20);uniqueIndex:user_id_comment_id;not null"`
	FileId     string `gorm:"column:file_id;type:varchar(32);not null"`
	CreateTime int64  `gorm:"column:create_time;autoCreateTime"`
	Remark     string `gorm:"column:remark"`
}

func (ShortVideoCommentLike) TableName() string {
	return "short_video_comment_like"
}

type ShortVideoCommentLikeRes struct {
	FileId              string `json:"file_id"`
	ShortVideoStatus    int32  `json:"short_video_status"`
	PublishUserID       string `json:"publisher_id"`
	PublishUser         string `json:"publish_user"`
	Content             string `json:"content"`
	CoverUrl            string `json:"cover_url"`
	MediaUrl            string `json:"media_url"`
	Size                int64  `json:"size"`
	Height              int64  `json:"height"`
	Width               int64  `json:"width"`
	CommentId           int64  `json:"comment_id"`
	CommentContent      string `json:"comment_content"`
	CommentUserName     string `json:"comment_user_name"`
	CommentUserID       string `json:"comment_user_id"`
	ReplyCommentId      int64  `json:"reply_comment_id"`
	ReplyUserName       string `json:"reply_user_name"`
	ReplyUserID         string `json:"reply_user_id"`
	ReplyCommentContent string `json:"reply_comment_content"`
	LikeId              int64  `json:"like_id"`
	LikeUserName        string `json:"like_user_name"`
	LikeUserID          string `json:"like_user_id"`
	LikeTime            int64  `json:"like_time"`
	Remark              string `json:"remark"`
	Status              int64  `json:"status"`
}

type ShortVideoFollow struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId     string `gorm:"column:user_id;type:varchar(32);uniqueIndex:user_id_fans_id;Index:user_id;not null;comment:用户id"`
	FansId     string `gorm:"column:fans_id;type:varchar(32);uniqueIndex:user_id_fans_id;Index:fans_id;not null;comment:粉丝id"`
	CreateTime int64  `gorm:"column:create_time;autoCreateTime"`
	Remark     string `gorm:"column:remark;type:varchar(256);"`
}

func (ShortVideoFollow) TableName() string {
	return "short_video_follow"
}

type ShortVideoFollowRes struct {
	ShortVideoFollow
	UserFace string
	UserName string
	FansFace string
	FansName string
	Remark   string
}

type ShortVideoNotice struct {
	Id           int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	UserId       string `gorm:"column:user_id;type:varchar(32);not null;Index:user_id"`
	SourceUserId string `gorm:"column:source_user_id;type:varchar(32);not null"`
	FileId       string `gorm:"column:file_id;type:varchar(32)"`
	CommentId    int64  `gorm:"column:comment_id;type:bigint(20)"`
	Type         int8   `gorm:"column:type;type:tinyint(3);not null;default:1;comment:'1:点赞视频 2:点赞评论 3:回复视频 4:回复评论 5:关注 6:新视频'"`
	Context      string `gorm:"column:context;type:varchar(128);not null"`
	State        int8   `gorm:"column:state;type:tinyint(3);not null;default:1;comment:'1:未读 2:已读'"`
	CreateTime   int64  `gorm:"column:create_time;autoCreateTime"`
}

func (ShortVideoNotice) TableName() string {
	return "short_video_notice"
}

type HomeVisual struct {
	Id         int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	StatusName string `gorm:"column:status_name"`
	Status     int8   `gorm:"column:status;type:tinyint(3);not null;default:1;comment:'1:开启 2:关闭'"`
}

func (HomeVisual) TableName() string {
	return "home_visual"
}

type Domain struct {
	Id      int64  `gorm:"column:id;primary_key;autoIncrement;type:bigint"`
	Address string `gorm:"column:address"`
}

func (Domain) TableName() string {
	return "domain"
}
