package cronTask

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbShortVideo "Open_IM/pkg/proto/short_video"
	"Open_IM/pkg/utils"
	"github.com/robfig/cron/v3"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	vod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vod/v20180717"
	"strings"
	"time"
)

const cronTaskOperationID = "cronTaskOperationID-"

func StartCronTask() {
	log.NewPrivateLog("cron")
	c := cron.New()
	// group statistics
	// 接收消息并统计到redis
	log.NewInfo(getCronTaskOperationID(), "cron config", config.Config.Cron.MsgStatistical)
	_, err1 := c.AddFunc(config.Config.Cron.MsgStatistical, func() {
		for {
			result, err := db.DB.PopGroupMessageStatistic()
			if err != nil {
				time.Sleep(time.Second * 5)
				continue
			}

			res := strings.Split(result, " ")
			if len(res) != 3 {
				log.NewError(getCronTaskOperationID(), "split group message statistical error", result)
				time.Sleep(30 * time.Millisecond)
				continue
			}

			groupId := res[0]
			userId := res[1]
			sendTime := res[2] // send time is unix time 1671530408345
			date := time.UnixMilli(utils.StringToInt64(sendTime)).Format("0601")

			_ = db.DB.SaveUsersByGroupMessage(groupId, userId, date)
			_ = db.DB.IncrGroupMessageCountByDay(groupId, date)
			_ = db.DB.SaveNeedCountGroup(groupId, date)
		}
	})
	if err1 != nil {
		log.NewError(getCronTaskOperationID(), "start cron MsgStatistical failed", err1.Error(), config.Config.Cron.MsgStatistical)
		panic(err1)
	}

	// 统计redis中的数据到mysql
	log.NewInfo(getCronTaskOperationID(), "cron config", config.Config.Cron.MsgCountStatistical)
	_, err2 := c.AddFunc(config.Config.Cron.MsgCountStatistical, func() {
		hour := time.Now().Hour()
		minute := time.Now().Minute()
		day := time.Now().Day()

		if day < 3 && hour == 0 && minute == 0 {
			month := time.Now().AddDate(0, 0, -1).Format("0601")
			groupList, _ := db.DB.GetNeedCountGroup(month)
			for _, groupId := range groupList {
				messageCount, _ := db.DB.GetGroupMessageCountByDay(groupId, month)
				userCount, _ := db.DB.GetUsersCountByGroupMessage(groupId, month)

				heat := utils.StringToInt64(messageCount)/20 + userCount
				im_mysql_model.SaveGroupHeat(groupId, month, utils.StringToInt64(messageCount), userCount, heat)
			}
		}

		month := time.Now().Format("0601")
		groupList, _ := db.DB.GetNeedCountGroup(month)

		for _, groupId := range groupList {
			messageCount, _ := db.DB.GetGroupMessageCountByDay(groupId, month)
			userCount, _ := db.DB.GetUsersCountByGroupMessage(groupId, month)

			heat := utils.StringToInt64(messageCount)/20 + userCount
			im_mysql_model.SaveGroupHeat(groupId, month, utils.StringToInt64(messageCount), userCount, heat)
		}
	})
	if err2 != nil {
		log.NewError(getCronTaskOperationID(), "start cron MsgCountStatistical failed", err2.Error(), config.Config.Cron.MsgCountStatistical)
		panic(err2)
	}

	c.Start()
	defer c.Stop()

	// 可靠回调 reliable callback
	if config.Config.Vod.IsReliableCallBack {
		log.NewInfo(getCronTaskOperationID(), "cron config", config.Config.Cron.VodCallback)
		secondCron := cron.New(cron.WithSeconds())
		_, err3 := secondCron.AddFunc(config.Config.Cron.VodCallback, handleVodCallback)
		if err3 != nil {
			log.NewError(getCronTaskOperationID(), "start cron VodCallback failed", err3.Error(), config.Config.Cron.VodCallback)
			panic(err3)
		}
		secondCron.Start()
		defer secondCron.Stop()
	}

	log.NewInfo(getCronTaskOperationID(), "start cron task success")
	for {
		time.Sleep(10 * time.Second)
	}
}

func getCronTaskOperationID() string {
	return cronTaskOperationID + utils.OperationIDGenerator()
}

func handleVodCallback() {
	operationId := getCronTaskOperationID()
	log.NewInfo(operationId, "start handleVodCallback")

	credential := common.NewCredential(
		config.Config.Vod.SecretId,
		config.Config.Vod.SecretKey,
	)

	request := vod.NewPullEventsRequest()
	request.SubAppId = common.Uint64Ptr(uint64(config.Config.Vod.VodSubAppId))

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 5
	cpf.SignMethod = "HmacSHA1"

	client, err := vod.NewClient(credential, "ap-guangzhou", cpf)
	if err != nil {
		log.NewError(getCronTaskOperationID(), "vod client init failed", err.Error())
		return
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImShortVideoName, operationId)
	if etcdConn == nil {
		log.NewError(getCronTaskOperationID(), "etcdConn is nil")
		return
	}
	rpcClient := pbShortVideo.NewShortVideoClient(etcdConn)

	for i := 0; i < 99; i++ {
		events, err := client.PullEvents(request)
		if err != nil {
			log.NewWarn(getCronTaskOperationID(), "vod callback failed", err.Error())
			return
		}

		if len(events.Response.EventSet) > 0 {
			eventHandles := make([]*string, len(events.Response.EventSet))
			for _, event := range events.Response.EventSet {
				eventHandles = append(eventHandles, event.EventHandle)
				switch *event.EventType {
				case "NewFileUpload":
					newFileUpload(rpcClient, event)
					break
				case "ProcedureStateChanged":
					procedureStateChanged(rpcClient, event)
					break
				case "FileDeleted":
					fileDeleted(rpcClient, event)
					break
				default:
					log.NewError(getCronTaskOperationID(), "vod callback failed", "unknown event type", *event.EventType, *event)
					break
				}
			}

			// Confirm Events
			confirmEvents, err := client.ConfirmEvents(&vod.ConfirmEventsRequest{
				SubAppId:     common.Uint64Ptr(uint64(config.Config.Vod.VodSubAppId)),
				EventHandles: eventHandles,
			})
			log.NewInfo(getCronTaskOperationID(), "confirm events", confirmEvents, err)
		}
	}
}
