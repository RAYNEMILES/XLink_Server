package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

var (
	_, b, _, _ = runtime.Caller(0)
	// Root folder of this project
	Root = filepath.Join(filepath.Dir(b), "../../..")
)

var Config config

type callBackConfig struct {
	Enable                 bool `yaml:"enable"`
	CallbackTimeOut        int  `yaml:"callbackTimeOut"`
	CallbackFailedContinue bool `yaml:"callbackFailedContinue"`
}

type config struct {
	Environment    string `yaml:"environment"`
	ServerIP       string `yaml:"serverip"`
	IsSkipDatabase bool   `yaml:"is_skip_database"`

	RpcRegisterIP string `yaml:"rpcRegisterIP"`
	ListenIP      string `yaml:"listenIP"`

	ServerVersion string `yaml:"serverversion"`
	Api           struct {
		GinPort  []int  `yaml:"openImApiPort"`
		ListenIP string `yaml:"listenIP"`
	}
	CmsApi struct {
		GinPort  []int  `yaml:"openImCmsApiPort"`
		ListenIP string `yaml:"listenIP"`
	}
	Sdk struct {
		WsPort  []int    `yaml:"openImSdkWsPort"`
		DataDir []string `yaml:"dataDir"`
	}
	LocalData struct {
		DataDir string `yaml:"dataDir"`
	}
	Credential struct {
		Tencent struct {
			AppID             string `yaml:"appID"`
			Region            string `yaml:"region"`
			Accelerate        bool   `yaml:"accelerate"`
			Bucket            string `yaml:"bucket"`
			PersistenceBucket string `yaml:"persistenceBucket"`
			SecretID          string `yaml:"secretID"`
			SecretKey         string `yaml:"secretKey"`
		} `yaml:"tencent"`
		Ali struct {
			RegionID           string `yaml:"regionID"`
			AccessKeyID        string `yaml:"accessKeyID"`
			AccessKeySecret    string `yaml:"accessKeySecret"`
			StsEndpoint        string `yaml:"stsEndpoint"`
			OssEndpoint        string `yaml:"ossEndpoint"`
			Bucket             string `yaml:"bucket"`
			FinalHost          string `yaml:"finalHost"`
			StsDurationSeconds int64  `yaml:"stsDurationSeconds"`
			OssRoleArn         string `yaml:"OssRoleArn"`
		} `yaml:"ali"`
		Minio struct {
			Bucket              string `yaml:"bucket"`
			AppBucket           string `yaml:"appBucket"`
			PersistenceBucket   string `yaml:"persistenceBucket"`
			Location            string `yaml:"location"`
			Endpoint            string `yaml:"endpoint"`
			AccessKeyID         string `yaml:"accessKeyID"`
			SecretAccessKey     string `yaml:"secretAccessKey"`
			EndpointInner       string `yaml:"endpointInner"`
			EndpointInnerEnable bool   `yaml:"endpointInnerEnable"`
		} `yaml:"minio"`
	}

	Mysql struct {
		DBAddress      []string `yaml:"dbMysqlAddress"`
		DBUserName     string   `yaml:"dbMysqlUserName"`
		DBPassword     string   `yaml:"dbMysqlPassword"`
		DBDatabaseName string   `yaml:"dbMysqlDatabaseName"`
		DBTableName    string   `yaml:"DBTableName"`
		DBMsgTableNum  int      `yaml:"dbMsgTableNum"`
		DBMaxOpenConns int      `yaml:"dbMaxOpenConns"`
		DBMaxIdleConns int      `yaml:"dbMaxIdleConns"`
		DBMaxLifeTime  int      `yaml:"dbMaxLifeTime"`
	}
	Mongo struct {
		DBUri               string `yaml:"dbUri"`
		DBAddress           string `yaml:"dbAddress"`
		DBDirect            bool   `yaml:"dbDirect"`
		DBTimeout           int    `yaml:"dbTimeout"`
		DBDatabase          string `yaml:"dbDatabase"`
		DBSource            string `yaml:"dbSource"`
		DBUserName          string `yaml:"dbUserName"`
		DBPassword          string `yaml:"dbPassword"`
		DBMaxPoolSize       int    `yaml:"dbMaxPoolSize"`
		DBRetainChatRecords int    `yaml:"dbRetainChatRecords"`
	}
	Redis struct {
		DBAddress     []string `yaml:"dbAddress"`
		DBMaxIdle     int      `yaml:"dbMaxIdle"`
		DBMaxActive   int      `yaml:"dbMaxActive"`
		DBIdleTimeout int      `yaml:"dbIdleTimeout"`
		DBPassWord    string   `yaml:"dbPassWord"`
		EnableCluster bool     `yaml:"enableCluster"`
	}
	RpcPort struct {
		OpenImUserPort           []int `yaml:"openImUserPort"`
		OpenImFriendPort         []int `yaml:"openImFriendPort"`
		OpenImMessagePort        []int `yaml:"openImMessagePort"`
		OpenImMessageGatewayPort []int `yaml:"openImMessageGatewayPort"`
		OpenImGroupPort          []int `yaml:"openImGroupPort"`
		OpenImAuthPort           []int `yaml:"openImAuthPort"`
		OpenImPushPort           []int `yaml:"openImPushPort"`
		OpenImStatisticsPort     []int `yaml:"openImStatisticsPort"`
		OpenImMessageCmsPort     []int `yaml:"openImMessageCmsPort"`
		OpenImAdminCmsPort       []int `yaml:"openImAdminCmsPort"`
		OpenImOfficePort         []int `yaml:"openImOfficePort"`
		OpenImOrganizationPort   []int `yaml:"openImOrganizationPort"`
		OpenImConversationPort   []int `yaml:"openImConversationPort"`
		OpenImCachePort          []int `yaml:"openImCachePort"`
		OpenImLocalDataPort      []int `yaml:"openImLocalDataPort"`
		OpenImMomentsPort        []int `yaml:"openImMomentsPort"`
		OpenImNewsPort           []int `yaml:"openImNewsPort"`
		OpenImGameStore          []int `yaml:"openImGameStorePort"`
		OpenImShortVideoPort     []int `yaml:"openImShortVideoPort"`
		OpenImMistPort           []int `yaml:"openImMistPort"`
	}
	RpcRegisterName struct {
		OpenImStatisticsName         string `yaml:"openImStatisticsName"`
		OpenImUserName               string `yaml:"openImUserName"`
		OpenImFriendName             string `yaml:"openImFriendName"`
		OpenImOfflineMessageName     string `yaml:"openImOfflineMessageName"`
		OpenImPushName               string `yaml:"openImPushName"`
		OpenImOnlineMessageRelayName string `yaml:"openImOnlineMessageRelayName"`
		OpenImGroupName              string `yaml:"openImGroupName"`
		OpenImAuthName               string `yaml:"openImAuthName"`
		OpenImMessageCMSName         string `yaml:"openImMessageCMSName"`
		OpenImAdminCMSName           string `yaml:"openImAdminCMSName"`
		OpenImOfficeName             string `yaml:"openImOfficeName"`
		OpenImOrganizationName       string `yaml:"openImOrganizationName"`
		OpenImConversationName       string `yaml:"openImConversationName"`
		OpenImMoemntsName            string `yaml:"openImMoemntsName"`
		OpenImNewsName               string `yaml:"openImNewsName"`
		OpenImCacheName              string `yaml:"openImCacheName"`
		OpenImRealTimeCommName       string `yaml:"openImRealTimeCommName"`
		OpenImLocalDataName          string `yaml:"openImLocalDataName"`
		OpenImGameStoreName          string `yaml:"openImGameStoreName"`
		OpenImShortVideoName         string `yaml:"openImShortVideoName"`
	}
	Etcd struct {
		EtcdSchema string   `yaml:"etcdSchema"`
		EtcdAddr   []string `yaml:"etcdAddr"`
	}
	Log struct {
		StorageLocation       string   `yaml:"storageLocation"`
		RotationTime          int      `yaml:"rotationTime"`
		RemainRotationCount   uint     `yaml:"remainRotationCount"`
		RemainLogLevel        uint     `yaml:"remainLogLevel"`
		GormLogLevel          uint     `yaml:"gormLogLevel"`
		ElasticSearchSwitch   bool     `yaml:"elasticSearchSwitch"`
		ElasticSearchAddr     []string `yaml:"elasticSearchAddr"`
		ElasticSearchUser     string   `yaml:"elasticSearchUser"`
		ElasticSearchPassword string   `yaml:"elasticSearchPassword"`
	}
	ModuleName struct {
		LongConnSvrName string `yaml:"longConnSvrName"`
		MsgTransferName string `yaml:"msgTransferName"`
		PushName        string `yaml:"pushName"`
	}
	LongConnSvr struct {
		WebsocketPort       []int `yaml:"openImWsPort"`
		WebsocketMaxConnNum int   `yaml:"websocketMaxConnNum"`
		WebsocketMaxMsgLen  int   `yaml:"websocketMaxMsgLen"`
		WebsocketTimeOut    int   `yaml:"websocketTimeOut"`
	}

	Push struct {
		Tpns struct {
			Ios struct {
				AccessID  string `yaml:"accessID"`
				SecretKey string `yaml:"secretKey"`
			}
			Android struct {
				AccessID  string `yaml:"accessID"`
				SecretKey string `yaml:"secretKey"`
			}
			Enable bool `yaml:"enable"`
		}
		Jpns struct {
			AppKey       string `yaml:"appKey"`
			MasterSecret string `yaml:"masterSecret"`
			PushUrl      string `yaml:"pushUrl"`
			PushIntent   string `yaml:"pushIntent"`
			IsProduct    bool   `yaml:"isProduct"`
			Enable       bool   `yaml:"enable"`
		}
		Getui struct {
			PushUrl      string `yaml:"pushUrl"`
			AppKey       string `yaml:"appKey"`
			Enable       bool   `yaml:"enable"`
			Intent       string `yaml:"intent"`
			MasterSecret string `yaml:"masterSecret"`
		}
	}
	Manager struct {
		AppManagerUid          []string `yaml:"appManagerUid"`
		Secrets                []string `yaml:"secrets"`
		AppSysNotificationName string   `yaml:"appSysNotificationName"`
	}

	Kafka struct {
		Ws2mschat struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		Ws2mschatOffline struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		MsgToMongo struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		Ms2pschat struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		MsgToSyncData struct {
			Addr  []string `yaml:"addr"`
			Topic string   `yaml:"topic"`
		}
		ConsumerGroupID struct {
			MsgToRedis string `yaml:"msgToTransfer"`
			MsgToMongo string `yaml:"msgToMongo"`
			MsgToMySql string `yaml:"msgToMySql"`
			MsgToPush  string `yaml:"msgToPush"`
			SyncData   string `yaml:"syncData"`
		}
	}
	Secret                            string `yaml:"secret"`
	MultiLoginPolicy                  int    `yaml:"multiloginpolicy"`
	ChatPersistenceMysql              bool   `yaml:"chatpersistencemysql"`
	ReliableStorage                   bool   `yaml:"reliablestorage"`
	MsgCacheTimeout                   int    `yaml:"msgCacheTimeout"`
	GroupMessageHasReadReceiptEnable  bool   `yaml:"groupMessageHasReadReceiptEnable"`
	SingleMessageHasReadReceiptEnable bool   `yaml:"singleMessageHasReadReceiptEnable"`

	LocationIpServerAddressPreFix  string `yaml:"locationIpServerAddressPreFix"`
	LocationIpServerAddressPostFix string `yaml:"locationIpServerAddressPostFix"`

	CallbackAfterSendMsg struct {
		Switch     bool `yaml:"switch"`
		ExpireTime int  `yaml:"expireTime"`
	} `yaml:"callbackAfterSendMsg"`

	TokenPolicy struct {
		AccessSecret      string `yaml:"accessSecret"`
		AccessSecretGAuth string `yaml:"accessSecretGAuth"`
		AccessExpire      int64  `yaml:"accessExpire"`
	}
	MessageVerify struct {
		FriendVerify bool `yaml:"bytechatmyfriendVerify"`
	}
	IOSPush struct {
		PushSound  string `yaml:"pushSound"`
		BadgeCount bool   `yaml:"badgeCount"`
	}

	Callback struct {
		CallbackUrl                 string         `yaml:"callbackUrl"`
		CallbackBeforeSendSingleMsg callBackConfig `yaml:"callbackBeforeSendSingleMsg"`
		CallbackAfterSendSingleMsg  callBackConfig `yaml:"callbackAfterSendSingleMsg"`
		CallbackBeforeSendGroupMsg  callBackConfig `yaml:"callbackBeforeSendGroupMsg"`
		CallbackAfterSendGroupMsg   callBackConfig `yaml:"callbackAfterSendGroupMsg"`
		CallbackWordFilter          callBackConfig `yaml:"callbackWordFilter"`
		CallbackUserOnline          callBackConfig `yaml:"callbackUserOnline"`
		CallbackUserOffline         callBackConfig `yaml:"callbackUserOffline"`
		CallbackOfflinePush         callBackConfig `yaml:"callbackOfflinePush"`
	} `yaml:"callback"`
	Notification struct {
		// /////////////////////group/////////////////////////////
		GroupCreated struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupCreated"`

		GroupInfoSet struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupInfoSet"`

		JoinGroupApplication struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"joinGroupApplication"`

		MemberQuit struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberQuit"`

		GroupApplicationAccepted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupApplicationAccepted"`

		GroupApplicationRejected struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupApplicationRejected"`

		GroupOwnerTransferred struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupOwnerTransferred"`

		MemberKicked struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberKicked"`

		MemberInvited struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberInvited"`

		MemberEnter struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"memberEnter"`

		GroupDismissed struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupDismissed"`

		GroupMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMuted"`

		GroupCancelMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupCancelMuted"`

		GroupMemberMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberMuted"`

		GroupMemberCancelMuted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberCancelMuted"`
		GroupMemberInfoSet struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberInfoSet"`
		GroupMemberSetToAdmin struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberSetToAdmin"`
		GroupMemberSetToOrdinary struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"groupMemberSetToOrdinaryUser"`
		OrganizationChanged struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"organizationChanged"`

		// //////////////////////user///////////////////////
		UserInfoUpdated struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"userInfoUpdated"`

		// ////////////////////friend///////////////////////
		FriendApplication struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendApplicationAdded"`
		FriendApplicationApproved struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendApplicationApproved"`

		FriendApplicationRejected struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendApplicationRejected"`

		FriendAdded struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendAdded"`

		FriendDeleted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendDeleted"`
		FriendRemarkSet struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"friendRemarkSet"`
		BlackAdded struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"blackAdded"`
		BlackDeleted struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"blackDeleted"`
		ConversationOptUpdate struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"conversationOptUpdate"`
		ConversationSetPrivate struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  struct {
				OpenTips  string `yaml:"openTips"`
				CloseTips string `yaml:"closeTips"`
			} `yaml:"defaultTips"`
		} `yaml:"conversationSetPrivate"`
		WorkMomentsNotification struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"workMomentsNotification"`
		JoinDepartmentNotification struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"joinDepartmentNotification"`
		Signal struct {
			OfflinePush struct {
				Title string `yaml:"title"`
			} `yaml:"offlinePush"`
		} `yaml:"signal"`
		UserBlockedYouNotification struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"userBlockedYouNotification"`
		UserNotFriendNotification struct {
			Conversation PConversation `yaml:"conversation"`
			OfflinePush  POfflinePush  `yaml:"offlinePush"`
			DefaultTips  PDefaultTips  `yaml:"defaultTips"`
		} `yaml:"userNotFriendNotification"`
	}
	Demo struct {
		Port         []int  `yaml:"openImDemoPort"`
		ListenIP     string `yaml:"listenIP"`
		AliSMSVerify struct {
			AccessKeyID                  string `yaml:"accessKeyId"`
			AccessKeySecret              string `yaml:"accessKeySecret"`
			SignName                     string `yaml:"signName"`
			VerificationCodeTemplateCode string `yaml:"verificationCodeTemplateCode"`
		}
		SuperCode string `yaml:"superCode"`
		CodeTTL   int    `yaml:"codeTTL"`
		ExpireTTL int    `yaml:"expireTTL"`
		Mail      struct {
			Title                   string `yaml:"title"`
			SenderMail              string `yaml:"senderMail"`
			SenderAuthorizationCode string `yaml:"senderAuthorizationCode"`
			SmtpAddr                string `yaml:"smtpAddr"`
			SmtpPort                int    `yaml:"smtpPort"`
		}
		TestDepartMentID string `yaml:"testDepartMentID"`
	}
	Sms struct {
		Api                string `yaml:"api"`
		SmsRegisterCN      string `yaml:"sms_register_cn"`
		SmsResetpasswordCN string `yaml:"sms_resetpassword_cn"`
		SmsBindEmailCn     string `yaml:"sms_bind_email_cn"`
		SmsRegisterEN      string `yaml:"sms_register_en"`
		SmsResetpasswordEN string `yaml:"sms_resetpassword_en"`
		SmsBindEmailEn     string `yaml:"sms_bind_email_en"`
		Twilio             struct {
			ApiUrl     string `yaml:"api_url"`
			AccountSID string `yaml:"account_sid"`
			ApiSID     string `yaml:"api_sid"`
			ApiSecret  string `yaml:"api_secret"`
			SendNumber string `yaml:"send_number"`
		}
		Huawei struct {
			Api       string `yaml:"api"`
			Signature string `yaml:"signature"`
			Cn        struct {
				SmsRegisterEn      string `yaml:"sms_register_en"`
				SmsResetPasswordEn string `yaml:"sms_reset_password_en"`
				SmsDeleteAccountEn string `yaml:"sms_delete_account_en"`
				SmsRegisterCn      string `yaml:"sms_register_cn"`
				SmsResetPasswordCn string `yaml:"sms_reset_password_cn"`
				SmsDeleteAccountCn string `yaml:"sms_delete_account_cn"`
				AppKey             string `yaml:"app_key"`
				AppSecret          string `yaml:"app_secret"`
				Sender             string `yaml:"sender"`
			} `yaml:"cn"`
			En struct {
				SmsRegisterEn      string `yaml:"sms_register_en"`
				SmsResetPasswordEn string `yaml:"sms_reset_password_en"`
				SmsDeleteAccountEn string `yaml:"sms_delete_account_en"`
				SmsRegisterCn      string `yaml:"sms_register_cn"`
				SmsResetPasswordCn string `yaml:"sms_reset_password_cn"`
				SmsDeleteAccountCn string `yaml:"sms_delete_account_cn"`
				AppKey             string `yaml:"app_key"`
				AppSecret          string `yaml:"app_secret"`
				Sender             string `yaml:"sender"`
			} `yaml:"en"`
		} `yaml:"huawei"`
	} `yaml:"sms"`
	Trtc struct {
		SdkAppid  string `yaml:"sdkappid"`
		SecretKey string `yaml:"secretkey"`
		Record    struct {
			MaxIdleTime uint64 `yaml:"maxIdleTime"`
			StreamType  uint64 `yaml:"streamType"`
			RecordMode  uint64 `yaml:"recordMode"`
			Vendor      uint64 `yaml:"vendor"`
			Region      string `yaml:"region"`
		} `yaml:"record"`
	} `yaml:"trtc"`
	Rtc struct {
		SignalTimeout string `yaml:"signalTimeout"`
	} `yaml:"rtc"`

	//
	ServerCtrlPrivatePrvMsg     int `yaml:"serverCtrlPrivatePrvMsg"`
	ServerCtrlGroupPrvMsg       int `yaml:"serverCtrlGroupPrvMsg"`
	MembersInGroupMaxLimit      int `yaml:"membersInGroupMaxLimit"`
	RestrictUserActionTimeLimit int `yaml:"restrictUserActionTimeLimit"`
	// RestrictUserAddFriendOps    bool  `yaml:"restrictUserAddFriendOps"`
	PhoneRegisterType []int `yaml:"phoneRegisterType"`
	Invite            struct {
		IsOpen int    `yaml:"isOpen"`
		Salt   string `yaml:"salt"`
	} `yaml:"invite"`
	Channel struct {
		IsOpen int `yaml:"isOpen"`
	} `yaml:"channel"`
	Cron struct {
		MsgStatistical      string `yaml:"msgStatistical"`
		MsgCountStatistical string `yaml:"msgCountStatistical"`
		VodCallback         string `yaml:"vodCallback"`
	} `yaml:"cron"`
	ImageProxy struct {
		ImgProxyAddress  string `yaml:"imgProxyAddress"`
		ImgGravityPrePos string `yaml:"imgGravityPrePos"`
	} `yaml:"imageProxy"`
	News struct {
		ArticleUrlTemplate string `yaml:"articleUrlTemplate"`
	} `yaml:"news"`
	DiscoverToken struct {
		IsOpen int `yaml:"isOpen"`
	} `yaml:"discoverToken"`

	Oauth struct {
		IsOpen   int `yaml:"isOpen"`
		Facebook struct {
			ClientID     string `yaml:"clientID"`
			ClientSecret string `yaml:"clientSecret"`
		} `yaml:"facebook"`
		Google struct {
			Web struct {
				ClientID     string `yaml:"clientID"`
				ClientSecret string `yaml:"clientSecret"`
			} `yaml:"web"`
			Ios struct {
				ClientID     string `yaml:"clientID"`
				ClientSecret string `yaml:"clientSecret"`
			} `yaml:"ios"`
			Android struct {
				ClientID     string `yaml:"clientID"`
				ClientSecret string `yaml:"clientSecret"`
			} `yaml:"android"`
		} `yaml:"google"`
		Apple struct {
			AppleId      string `yaml:"appleId"`
			ServerId     string `yaml:"serverId"`
			ClientSecret string `yaml:"clientSecret"`
			TeamId       string `yaml:"teamId"`
			KeyId        string `yaml:"keyId"`
			PrivateKey   string `yaml:"privateKey"`
		} `yaml:"apple"`
	} `yaml:"oauth"`
	Favorite struct {
		MaxCapacity int64 `yaml:"maxCapacity"`
	} `yaml:"favorite"`
	Gorse struct {
		Url   string `yaml:"url"`
		Token string `yaml:"token"`
	} `yaml:"gorse"`
	Vod struct {
		SecretId           string `yaml:"secretId"`
		SecretKey          string `yaml:"secretKey"`
		VodSubAppId        int    `yaml:"vodSubAppId"`
		IsReliableCallBack bool   `yaml:"isReliableCallBack"`
	} `yaml:"vod"`

	AdminUser2FAuthEnable bool   `yaml:"adminUser2FAuthEnable"`
	TotpIssuerName        string `yaml:"totpIssuerName"`
	AllowGuestLogin       bool   `yaml:"allowGuestLogin"`
	Official              struct {
		SystemOfficialType int32  `yaml:"systemOfficialType"`
		SystemOfficialName string `yaml:"systemOfficialName"`
	} `yaml:"official"`
}
type PConversation struct {
	ReliabilityLevel int  `yaml:"reliabilityLevel"`
	UnreadCount      bool `yaml:"unreadCount"`
}

type POfflinePush struct {
	PushSwitch bool   `yaml:"switch"`
	Title      string `yaml:"title"`
	Desc       string `yaml:"desc"`
	Ext        string `yaml:"ext"`
}
type PDefaultTips struct {
	Tips string `yaml:"tips"`
}

func init() {
	cfgName := os.Getenv("CONFIG_NAME")
	if len(cfgName) == 0 {
		cfgName = Root + "/config/config.yaml"
	}

	bytes, err := ioutil.ReadFile(cfgName)
	if err == nil {
		if err = yaml.Unmarshal(bytes, &Config); err != nil {
			panic(err.Error())
		}
	}
	//if err != nil {
	//	panic(err.Error())
	//}
	//if err = yaml.Unmarshal(bytes, &Config); err != nil {
	//	panic(err.Error())
	//}
}
