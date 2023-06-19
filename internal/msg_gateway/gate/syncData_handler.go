package gate

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/proto/local_database"
	"Open_IM/pkg/utils"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"time"
)

type SyncDataConsumerHandler struct {
	msgHandle                func(msg *sarama.ConsumerMessage)
	syncGroupConsumerHandler *kafka.MConsumerGroup
}

func (h *SyncDataConsumerHandler) Init(method func(msg *sarama.ConsumerMessage)) {
	h.msgHandle = method
	h.syncGroupConsumerHandler = kafka.NewMConsumerGroup(&kafka.MConsumerGroupConfig{KafkaVersion: sarama.V2_0_0_0,
		OffsetsInitial: sarama.OffsetNewest, IsReturnErr: false}, []string{config.Config.Kafka.MsgToSyncData.Topic}, config.Config.Kafka.MsgToSyncData.Addr,
		config.Config.Kafka.ConsumerGroupID.SyncData)
	log.NewInfo("", utils.GetSelfFuncName(), "msgHandle", method, "syncGroupConsumerHandler", h.syncGroupConsumerHandler)
}

func (SyncDataConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error {
	log.NewInfo("", utils.GetSelfFuncName(), "syncgroup_handler")
	return nil
}
func (SyncDataConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.NewInfo("", utils.GetSelfFuncName(), "syncgroup_handler")
	return nil
}

func (h *SyncDataConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.NewInfo("", utils.GetSelfFuncName(), "syncgroup_handler")
	for {
		if sess == nil {
			log.NewWarn("", " sess == nil, waiting ")
			time.Sleep(100 * time.Millisecond)
		} else {
			break
		}
	}

	for msg := range claim.Messages() {
		memberID := string(msg.Key)

		log.NewDebug("", "kafka get sync group data", "memberID", memberID, "msgTopic", msg.Topic, "msgPartition", msg.Partition, "msg", string(msg.Value))

		syncMsg := local_database.SyncDataMsg{}
		err := proto.Unmarshal(msg.Value, &syncMsg)
		if err != nil {
			log.NewError("", utils.GetSelfFuncName(), "Unmarshal error", err.Error())
			sess.MarkMessage(msg, "")
			continue
		}

		log.NewDebug("", utils.GetSelfFuncName(), "msg.userID", syncMsg.UserID)

		//send it to websocket
		//mark it is readed before send to websocket, because it take much time
		sess.MarkMessage(msg, "")

		if h.msgHandle != nil {
			//log.NewError("", utils.GetSelfFuncName(), "msg handle:", utils.Bytes2String(msg.Key), utils.Bytes2String(msg.Value))
			h.msgHandle(msg)
		}

	}
	return nil
}
