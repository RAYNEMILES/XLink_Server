/*
** description("").
** copyright('tuoyun,www.tuoyun.net').
** author("fg,Gordon@tuoyun.net").
** time(2021/5/11 15:37).
 */
package logic

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_msg_model"
	kfk "Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	pbMsg "Open_IM/pkg/proto/chat"
	"Open_IM/pkg/utils"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"strings"
	"time"
)

type PersistentConsumerHandler struct {
	msgHandle               map[string]fcb
	persistentConsumerGroup *kfk.MConsumerGroup
}

func (pc *PersistentConsumerHandler) Init() {
	pc.msgHandle = make(map[string]fcb)
	pc.msgHandle[config.Config.Kafka.Ws2mschat.Topic] = pc.handleChatWs2Mysql
	pc.persistentConsumerGroup = kfk.NewMConsumerGroup(&kfk.MConsumerGroupConfig{KafkaVersion: sarama.V2_0_0_0,
		OffsetsInitial: sarama.OffsetNewest, IsReturnErr: false}, []string{config.Config.Kafka.Ws2mschat.Topic},
		config.Config.Kafka.Ws2mschat.Addr, config.Config.Kafka.ConsumerGroupID.MsgToMySql)

}

func (pc *PersistentConsumerHandler) handleChatWs2Mysql(cMsg *sarama.ConsumerMessage, msgKey string, _ sarama.ConsumerGroupSession) {
	msg := cMsg.Value
	//log.NewInfo("msg come here mysql!!!", "", "msg", string(msg), msgKey)
	var tag bool
	msgFromMQ := pbMsg.MsgDataToMQ{}
	err := proto.Unmarshal(msg, &msgFromMQ)
	if err != nil {
		log.NewError(msgFromMQ.OperationID, "msg_transfer Unmarshal msg err", "msg", string(msg), "err", err.Error())
		return
	}
	//log.Debug(msgFromMQ.OperationID, "proto.Unmarshal MsgDataToMQ", msgFromMQ.String())
	//Control whether to store history messages (mysql)
	isPersist := utils.GetSwitchFromOptions(msgFromMQ.MsgData.Options, constant.IsPersistent)
	//Only process receiver data
	if isPersist {
		switch msgFromMQ.MsgData.SessionType {
		case constant.SingleChatType, constant.NotificationChatType:
			if msgKey == msgFromMQ.MsgData.RecvID {
				tag = true
			}
		case constant.GroupChatType:
			//msgKey == msgFromMQ.MsgData.SendID
			if strings.Contains(msgKey, "-"+msgFromMQ.MsgData.RecvID) {
				tag = true
			}
		case constant.SuperGroupChatType:
			tag = true
		}
		if tag {
			//log.NewError(msgFromMQ.OperationID, "msg_transfer msg persisting", string(msg))
			if err = im_mysql_msg_model.InsertMessageToChatLog(msgFromMQ); err != nil {
				log.NewError(msgFromMQ.OperationID, "Message insert failed", "err", err.Error(), "msg", msgFromMQ.String())
				return
			}
		}

	}

	/**
	//sync the local Database
	log.NewInfo(msgFromMQ.OperationID, "handleChatWs2Mysql Options:", msgFromMQ.MsgData.Options)
	isSyncToLocalDataBase := utils.GetSwitchFromOptions(msgFromMQ.MsgData.Options, constant.IsSyncToLocalDataBase)
	if isSyncToLocalDataBase {
		groupID := msgFromMQ.MsgData.GroupID
		var opUserID string
		tmpContent := string(msgFromMQ.MsgData.Content)
		tmpIndex := strings.Index(tmpContent, "{")

		if tmpIndex != -1 {
			tmpContent2 := tmpContent[tmpIndex:]
			log.NewInfo(msgFromMQ.OperationID, "SyncToLocalDataBase tmpContent2 ", tmpContent2)
			content := make(map[string]interface{})
			err := json.Unmarshal([]byte(tmpContent2), &content)
			if err != nil {
				log.Error(msgFromMQ.OperationID, "SyncToLocalDataBase Unmarshal failed ", err.Error())
				return
			}

			log.Info(msgFromMQ.OperationID, utils.GetSelfFuncName(), "SyncToLocalDataBase Content:", tmpContent2)

			log.Info(msgFromMQ.OperationID, utils.GetSelfFuncName(), "ContentType ", msgFromMQ.MsgData.ContentType)
			switch msgFromMQ.MsgData.ContentType {
			case constant.GroupMemberMutedNotification:
				mutedUser, ok := content["mutedUser"]
				if ok {
					m := mutedUser.(map[string]interface{})
					userID, ok2 := m["userID"]
					if ok2 {
						opUserID = userID.(string)
						log.NewInfo(msgFromMQ.OperationID, "SyncToLocalDataBase groupID userID", groupID, userID)
					} else {
						log.Error(msgFromMQ.OperationID, "SyncToLocalDataBase get userID failed ", err.Error())
						return
					}
				}

				if opUserID != "" && groupID != "" {
					log.NewInfo(msgFromMQ.OperationID, "SyncToLocalDataBase ", msgFromMQ.MsgData.ContentType)
					etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, msgFromMQ.OperationID)
					if etcdConn == nil {
						log.Error(msgFromMQ.OperationID, "OpenImLocalDataName rpc connect failed ", groupID, opUserID)
						return
					}
					client := local_database.NewLocalDataBaseClient(etcdConn)
					reqPb := local_database.SyncGroupMemberInfoReq{
						UserID:  opUserID,
						GroupID: groupID,
					}
					_, err2 := client.SyncGroupMemberInfoToLocal(context.Background(), &reqPb)
					if err2 != nil {
						log.Error(msgFromMQ.OperationID, "SyncGroupMemberInfoToLocal failed ", err2.Error(), groupID, opUserID)
						return
					}
				}
			case constant.GroupCreatedNotification, constant.MemberInvitedNotification:
				opUser, ok := content["opUser"]
				if ok {
					op := opUser.(map[string]interface{})
					if op == nil {
						log.Error(msgFromMQ.OperationID, utils.GetSelfFuncName(), "opUser nil ")
						return
					}

					opUserID, ok = op["userID"].(string)
				}

				if opUserID == "" {
					log.Error(msgFromMQ.OperationID, utils.GetSelfFuncName(), "opUserID nil ")
					return
				}

				var memberList interface{}
				if msgFromMQ.MsgData.ContentType == constant.GroupCreatedNotification {
					memberList, ok = content["memberList"]
					log.Info(msgFromMQ.OperationID, utils.GetSelfFuncName(), "GroupCreatedNotification ok:", ok, opUserID)
				} else if msgFromMQ.MsgData.ContentType == constant.MemberInvitedNotification {
					memberList, ok = content["invitedUserList"]
					log.Info(msgFromMQ.OperationID, utils.GetSelfFuncName(), "MemberInvitedNotification ok:", ok, opUserID)
				}

				if memberList != nil {
					m := memberList.([]interface{})
					mStr, err := json.Marshal(m)
					log.Info(msgFromMQ.OperationID, utils.GetSelfFuncName(), "mStr:", mStr)
					if err != nil {
						log.Error(msgFromMQ.OperationID, utils.GetSelfFuncName(), "json.Marshal failed ")
						return
					}
					etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, msgFromMQ.OperationID)
					if etcdConn == nil {
						log.Error(msgFromMQ.OperationID, utils.GetSelfFuncName(), "OpenImLocalDataName rpc connect failed ", groupID, opUserID)
						return
					}
					client := local_database.NewLocalDataBaseClient(etcdConn)
					reqPb := local_database.SyncGroupMemberListReq{
						UserID:     opUserID,
						MemberList: string(mStr),
					}
					_, err2 := client.SyncGroupMemerListToLocal(context.Background(), &reqPb)
					if err2 != nil {
						log.Error(msgFromMQ.OperationID, utils.GetSelfFuncName(), "SyncToLocal failed ", err2.Error(), groupID, opUserID)
						return
					}
				} else {
					log.Error(msgFromMQ.OperationID, utils.GetSelfFuncName(), "memberlist getting error")
				}
			}
		}

	}
	**/
}
func (PersistentConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (PersistentConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (pc *PersistentConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	for {
		if sess == nil {
			log.NewWarn("", " sess == nil, waiting ")
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	for msg := range claim.Messages() {
		//log.NewError("", utils.GetSelfFuncName(), "kafka get info to mysql", "msgTopic", msg.Topic, "msgPartition", msg.Partition, "msg", string(msg.Value), "key", string(msg.Key))
		if len(msg.Value) != 0 {
			pc.msgHandle[msg.Topic](msg, string(msg.Key), sess)
		} else {
			log.Error("", "msg get from kafka but is nil", msg.Key)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
