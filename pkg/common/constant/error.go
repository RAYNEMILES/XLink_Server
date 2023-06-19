package constant

import "errors"

// key = errCode, string = errMsg
type ErrInfo struct {
	ErrCode int32
	ErrMsg  string
}

var (
	OK        = ErrInfo{0, ""}
	ErrServer = ErrInfo{500, "server error"}

	//	ErrMysql             = ErrInfo{100, ""}
	//	ErrMongo             = ErrInfo{110, ""}
	//	ErrRedis             = ErrInfo{120, ""}
	ErrParseToken = ErrInfo{700, ParseTokenMsg.Error()}
	//	ErrCreateToken       = ErrInfo{201, "Create token failed"}
	//	ErrAppServerKey      = ErrInfo{300, "key error"}
	ErrTencentCredential = ErrInfo{400, ThirdPartyMsg.Error()}

	//	ErrorUserRegister             = ErrInfo{600, "User registration failed"}
	//	ErrAccountExists              = ErrInfo{601, "The account is already registered and cannot be registered again"}
	//	ErrUserPassword               = ErrInfo{602, "User password error"}
	//	ErrRefreshToken               = ErrInfo{605, "Failed to refresh token"}
	//	ErrAddFriend                  = ErrInfo{606, "Failed to add friends"}
	//	ErrAgreeToAddFriend           = ErrInfo{607, "Failed to agree application"}
	//	ErrAddFriendToBlack           = ErrInfo{608, "Failed to add friends to the blacklist"}
	//	ErrGetBlackList               = ErrInfo{609, "Failed to get blacklist"}
	//	ErrDeleteFriend               = ErrInfo{610, "Failed to delete friend"}
	//	ErrGetFriendApplyList         = ErrInfo{611, "Failed to get friend application list"}
	//	ErrGetFriendList              = ErrInfo{612, "Failed to get friend list"}
	//	ErrRemoveBlackList            = ErrInfo{613, "Failed to remove blacklist"}
	//	ErrSearchUserInfo             = ErrInfo{614, "Can't find the user information"}
	//	ErrDelAppleDeviceToken        = ErrInfo{615, ""}
	//	ErrModifyUserInfo             = ErrInfo{616, "update user some attribute failed"}
	//	ErrSetFriendComment           = ErrInfo{617, "set friend comment failed"}
	//	ErrSearchUserInfoFromTheGroup = ErrInfo{618, "There is no such group or the user not in the group"}
	//	ErrCreateGroup                = ErrInfo{619, "create group chat failed"}
	//	ErrJoinGroupApplication       = ErrInfo{620, "Failed to apply to join the group"}
	//	ErrQuitGroup                  = ErrInfo{621, "Failed to quit the group"}
	//	ErrSetGroupInfo               = ErrInfo{622, "Failed to set group info"}
	//	ErrParam                      = ErrInfo{700, "param failed"}
	ErrTokenExpired                = ErrInfo{701, TokenExpiredMsg.Error()}
	ErrTokenInvalid                = ErrInfo{702, TokenInvalidMsg.Error()}
	ErrTokenMalformed              = ErrInfo{703, TokenMalformedMsg.Error()}
	ErrTokenNotValidYet            = ErrInfo{704, TokenNotValidYetMsg.Error()}
	ErrTokenUnknown                = ErrInfo{705, TokenUnknownMsg.Error()}
	ErrTokenKicked                 = ErrInfo{706, TokenUserKickedMsg.Error()}
	ErrorUserLoginNetDisconnection = ErrInfo{707, "Network disconnection"}

	ErrAccess                 = ErrInfo{ErrCode: 801, ErrMsg: AccessMsg.Error()}
	ErrDB                     = ErrInfo{ErrCode: 802, ErrMsg: DBMsg.Error()}
	ErrArgs                   = ErrInfo{ErrCode: 803, ErrMsg: ArgsMsg.Error()}
	ErrStatus                 = ErrInfo{ErrCode: 804, ErrMsg: StatusMsg.Error()}
	ErrCallback               = ErrInfo{ErrCode: 809, ErrMsg: CallBackMsg.Error()}
	ErrSendLimit              = ErrInfo{ErrCode: 810, ErrMsg: "send msg limit, to many request, try again later"}
	ErrMessageHasReadDisable  = ErrInfo{ErrCode: 811, ErrMsg: "message has read disable"}
	ErrInternal               = ErrInfo{ErrCode: 812, ErrMsg: "internal error"}
	ErrRPC                    = ErrInfo{ErrCode: 813, ErrMsg: "rpc failed"}
	ErrEmptyResponse          = ErrInfo{ErrCode: 814, ErrMsg: "response is empty"}
	LimitExceeded             = ErrInfo{ErrCode: 815, ErrMsg: "Limit Exceeded"}
	ErrGetVersion             = ErrInfo{ErrCode: 816, ErrMsg: "get version failed"}
	ErrUserEmailAlreadyExist  = ErrInfo{ErrCode: 817, ErrMsg: "The email is already registered and cannot be registered again"}
	ErrAddFriendStoped        = ErrInfo{ErrCode: 819, ErrMsg: AccessMsg.Error()}
	ErrUserPhoneAlreadyExsist = ErrInfo{ErrCode: 820, ErrMsg: PhoneAlreadyExsist.Error()}
	ErrUserIDAlreadyExsist    = ErrInfo{ErrCode: 821, ErrMsg: UserIDAlreadyExsist.Error()}
	AddFriendNotSuperUserErr  = ErrInfo{ErrCode: 822, ErrMsg: AddFriendNotSuper.Error()}

	AddFriendLimitGroupErr   = ErrInfo{ErrCode: 823, ErrMsg: "The user turns off friend requests from the source group"}
	AddFriendLimitQrErr      = ErrInfo{ErrCode: 824, ErrMsg: "The user turns off friend requests from the source qr code"}
	AddFriendLimitContactErr = ErrInfo{ErrCode: 825, ErrMsg: "The user turns off friend requests from the source contact card"}

	ErrUserBanned  = ErrInfo{ErrCode: 900, ErrMsg: "The operation was rejected!"}
	ErrUserBanned2 = ErrInfo{ErrCode: 9001, ErrMsg: "The operation was rejected"}
	ErrUserBanned3 = ErrInfo{ErrCode: 9002, ErrMsg: "The operation was rejected!"}

	ErrInviteCode            = ErrInfo{ErrCode: 901, ErrMsg: "Invite code error!"}
	ErrInviteCodeInexistence = ErrInfo{ErrCode: 902, ErrMsg: "Invite code existence!"}

	ErrCaptchaError     = ErrInfo{ErrCode: 903, ErrMsg: "Captcha incorrect!"}
	ErrChannelCodeError = ErrInfo{ErrCode: 904, ErrMsg: "Invitation code incorrect"}

	ErrChannelCodeInexistence  = ErrInfo{ErrCode: 905, ErrMsg: "Channel code inexistence!"}
	ErrChannelCodeIsNull       = ErrInfo{ErrCode: 906, ErrMsg: "Please enter invitation code"}
	OnlyOneOfficialChannelCode = ErrInfo{ErrCode: 907, ErrMsg: "Only one official channel code"}
	ErrChannelCodeIsDelete     = ErrInfo{ErrCode: 908, ErrMsg: "The channel code has been deleted!"}

	ErrInviteCodeLimit = ErrInfo{ErrCode: 1001, ErrMsg: "Invite code limit is failed!"}

	ErrInviteCodeSwitch = ErrInfo{ErrCode: 1101, ErrMsg: "Invite code switch is failed!"}

	ErrInviteCodeMultiDelete = ErrInfo{ErrCode: 1201, ErrMsg: "Invite code Multi delete is failed!"}

	ErrChannelCodeLimit = ErrInfo{ErrCode: 1301, ErrMsg: "Channel code limit is failed!"}

	ErrChannelCodeSwitch = ErrInfo{ErrCode: 1401, ErrMsg: "Channel code switch is failed!"}

	ErrChannelCodeMultiDelete = ErrInfo{ErrCode: 1501, ErrMsg: "Channel code Multi delete is failed!"}

	ErrAddInviteCodeIsExist      = ErrInfo{ErrCode: 1601, ErrMsg: "Add invite code is exist!"}
	ErrAddInviteCodeUserIsExist  = ErrInfo{ErrCode: 1602, ErrMsg: "user is exist!"}
	ErrAddInviteCodeUserNotExist = ErrInfo{ErrCode: 1602, ErrMsg: "user not exist!"}
	ErrAddInviteCodeUserHasExist = ErrInfo{ErrCode: 1602, ErrMsg: "user already has invite code!"}
	ErrAddInviteCodeIsDelete     = ErrInfo{ErrCode: 1603, ErrMsg: "The share code has been deleted!"}

	ErrEditInviteCodeIsNotExist = ErrInfo{ErrCode: 1701, ErrMsg: "Edit invite code is not exist!"}
	ErrEditInviteCodeFailed     = ErrInfo{ErrCode: 1702, ErrMsg: "Edit invite code failed!"}

	ErrRevokeMessageTypeError = ErrInfo{ErrCode: 1801, ErrMsg: "Message type error!"}
	ErrNotASuperUser          = ErrInfo{ErrCode: 1805, ErrMsg: "You are not a super user"}

	ErrUserExistArg              = ErrInfo{ErrCode: 1901, ErrMsg: "Args Error!"}
	ErrUserExistUserIdExist      = ErrInfo{ErrCode: 1902, ErrMsg: "User Id Exist!"}
	ErrUserExistPhoneNumberExist = ErrInfo{ErrCode: 1903, ErrMsg: "Phone Number Exist!"}
	ErrChannelCodeNotExist       = ErrInfo{ErrCode: 1904, ErrMsg: "Channel Code Not Exist!"}
	ErrChannelCodeInvalid        = ErrInfo{ErrCode: 1905, ErrMsg: "Channel Code Invalid!"}
	ErrInviteCodeInvalid         = ErrInfo{ErrCode: 1906, ErrMsg: "Invite Code Invalid!"}

	ErrNotAllowGuestLogin  = ErrInfo{ErrCode: 2001, ErrMsg: "Not allow guest login!"}
	ErrRegisterByUuidLimit = ErrInfo{ErrCode: 2002, ErrMsg: "Register by uuid limit!"}
	ErrPasswordIncorrect   = ErrInfo{ErrCode: PasswordErr, ErrMsg: "Password is incorrect"}

	ErrOauthBoundOver     = ErrInfo{ErrCode: 2101, ErrMsg: "Oauth bound over!"}
	ErrOauthBindingFailed = ErrInfo{ErrCode: 2102, ErrMsg: "Oauth binding failed!"}
	ErrOauthUnbindingFail = ErrInfo{ErrCode: 2103, ErrMsg: "Oauth unbinding fail!"}

	ErrOfficialStatus               = ErrInfo{ErrCode: 2201, ErrMsg: "invalid official account process status"}
	ErrOfficialNicknameUpdateLocked = ErrInfo{ErrCode: 2202, ErrMsg: "official nickname update locked"}
	ErrOfficialNameExist            = ErrInfo{ErrCode: 2203, ErrMsg: "the official name has existed"}

	ErrOfficialArticleNotExist = ErrInfo{ErrCode: 2301, ErrMsg: "official article does not exist"}
	ErrOfficialActionForbidden = ErrInfo{ErrCode: 2302, ErrMsg: "action forbidden"}

	ErrRegisterParamEmailErr   = ErrInfo{ErrCode: 2201, ErrMsg: "Email is incorrect"}
	ErrRegisterParamUidErr     = ErrInfo{ErrCode: 2202, ErrMsg: "UserId is incorrect"}
	ErrRegisterParamPhoneErr   = ErrInfo{ErrCode: 2203, ErrMsg: "Phone is incorrect"}
	ErrRegisterParamEmailExist = ErrInfo{ErrCode: 2204, ErrMsg: "Email is exist"}
	ErrRegisterParamUidExist   = ErrInfo{ErrCode: 2205, ErrMsg: "UserId is exist"}
	ErrRegisterParamPhoneExist = ErrInfo{ErrCode: 2206, ErrMsg: "Phone is exist"}
	ErrAccountNotAvailable     = ErrInfo{ErrCode: 2207, ErrMsg: "This account not available"}
	ErrNickNameLength          = ErrInfo{ErrCode: 2210, ErrMsg: "nick name cant be more then 36 bytes or 12 characters"}
	AdminRoleDeleteError       = ErrInfo{ErrCode: 2220, ErrMsg: "cant delete, admin role already assigned"}
)

var (
	ParseTokenMsg       = errors.New("parse token failed")
	TokenExpiredMsg     = errors.New("token is timed out, please log in again")
	TokenInvalidMsg     = errors.New("token has been invalidated")
	TokenNotValidYetMsg = errors.New("token not active yet")
	TokenMalformedMsg   = errors.New("that's not even a token")
	TokenUnknownMsg     = errors.New("couldn't handle this token")
	TokenUserKickedMsg  = errors.New("user has been kicked")
	AccessMsg           = errors.New("no permission")
	StatusMsg           = errors.New("status is abnormal")
	DBMsg               = errors.New("db failed")
	ArgsMsg             = errors.New("args failed")
	CallBackMsg         = errors.New("callback failed")
	AddFriendNotSuper   = errors.New("sorry, only super user can add friends")

	ThirdPartyMsg = errors.New("third party error")

	PhoneAlreadyExsist  = errors.New("phone number is already used by a user")
	UserIDAlreadyExsist = errors.New("user id is already used by a user")
)

var (
	ErrFileIdAlreadyExist = ErrInfo{ErrCode: 40001, ErrMsg: "FileId is exist"}
	ErrFileIdIsNotExist   = ErrInfo{ErrCode: 40002, ErrMsg: "FileId is not exist"}
	// Like
	ErrLikeLimit     = ErrInfo{ErrCode: 40101, ErrMsg: "Frequent operation"}
	ErrLikeRedisFail = ErrInfo{ErrCode: 40102, ErrMsg: "Operation repeat"}
	ErrLikeMysqlFail = ErrInfo{ErrCode: 40103, ErrMsg: "Operation repeat"}

	// Comment
	ErrCommentLimit              = ErrInfo{ErrCode: 40201, ErrMsg: "Frequent operation"}
	ErrParentIdIsNotExist        = ErrInfo{ErrCode: 40202, ErrMsg: "ParentId is not exist"}
	ErrCommentContentIsSensitive = ErrInfo{ErrCode: 40203, ErrMsg: "Comment content is sensitive"}
	ErrCommentMysqlFail          = ErrInfo{ErrCode: 40204, ErrMsg: "Operation failed"}
	ErrCommentIdIsNotExist       = ErrInfo{ErrCode: 40205, ErrMsg: "CommentId is not exist"}
	ErrAuthority                 = ErrInfo{ErrCode: 40206, ErrMsg: "Authority is not enough"}

	// List
	ErrRecommendLimit = ErrInfo{ErrCode: 40301, ErrMsg: "Frequent operation"}
	ErrRecommendNil   = ErrInfo{ErrCode: 40302, ErrMsg: "Get recommend list failed"}

	// follow
	ErrFollowLimit        = ErrInfo{ErrCode: 40401, ErrMsg: "Frequent operation"}
	ErrFollowSelf         = ErrInfo{ErrCode: 40402, ErrMsg: "Can not follow self"}
	ErrFollowUserIdIsNil  = ErrInfo{ErrCode: 40403, ErrMsg: "UserId is nil"}
	ErrFollowRpcFail      = ErrInfo{ErrCode: 40404, ErrMsg: "Follow failed"}
	ErrFollowMysqlFail    = ErrInfo{ErrCode: 40405, ErrMsg: "Follow failed"}
	ErrFollowRedisFail    = ErrInfo{ErrCode: 40406, ErrMsg: "Follow failed"}
	ErrFollowAlreadyExist = ErrInfo{ErrCode: 40407, ErrMsg: "Already follow"}
	ErrFollowIsNotExist   = ErrInfo{ErrCode: 40408, ErrMsg: "Follow is not exist"}

	// count
	ErrCountUserIdIsNotExist = ErrInfo{ErrCode: 40501, ErrMsg: "UserId is not exist"}
	ErrCountRpcFail          = ErrInfo{ErrCode: 40502, ErrMsg: "Count failed"}
	ErrCountMysqlFail        = ErrInfo{ErrCode: 40503, ErrMsg: "Follow failed"}
)

var (
	ErrQrLoginSaveFailed        = ErrInfo{ErrCode: 50001, ErrMsg: "Get Qr Code Failed"}
	ErrQrLoginGetFailed         = ErrInfo{ErrCode: 50002, ErrMsg: "Get Qr Code Failed"}
	ErrQrLoginNotExist          = ErrInfo{ErrCode: 50003, ErrMsg: "Qr Code Not Exist"}
	ErrQrLoginStateErr          = ErrInfo{ErrCode: 50004, ErrMsg: "Qr Code State Error"}
	ErrQrLoginUpdateStateFailed = ErrInfo{ErrCode: 50005, ErrMsg: "Update Qr Code State Failed"}
	ErrQrLoginTokenErr          = ErrInfo{ErrCode: 50006, ErrMsg: "Qr Code Token Error"}
	ErrQrLoginDeviceIdErr       = ErrInfo{ErrCode: 50007, ErrMsg: "Qr Code DeviceId Error"}
)

const (
	NoError                         = 0
	FormattingError                 = 10001
	HasRegistered                   = 10002
	NotRegistered                   = 10003
	PasswordErr                     = 10004
	GetIMTokenErr                   = 10005
	RepeatSendCode                  = 10006
	MailSendCodeErr                 = 10007
	SmsSendCodeErr                  = 10008
	CodeInvalidOrExpired            = 10009
	RegisterFailed                  = 10010
	ResetPasswordFailed             = 10011
	NotAllowRegisterType            = 10012
	RemoteTokenExpiredErrorCode     = 10013
	RemoteTokenNotAssignedErrorCode = 10014
	DatabaseError                   = 10002
	ServerError                     = 10004
	HttpError                       = 10005
	IoError                         = 10006
	IntentionalError                = 10007
)

func (e ErrInfo) Error() string {
	return e.ErrMsg
}

func (e *ErrInfo) Code() int32 {
	return e.ErrCode
}
