/*
** description("").
** copyright('Open_IM,www.Open_IM.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/5/13 10:33).
 */
package logic

import (
	"Open_IM/pkg/common/config"
	kfk "Open_IM/pkg/common/kafka"
	"Open_IM/pkg/common/log"
	pbPush "Open_IM/pkg/proto/push"
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	"time"
)

type fcb func(msg []byte)

type PushConsumerHandler struct {
	msgHandle         map[string]fcb
	pushConsumerGroup *kfk.MConsumerGroup
}

func (ms *PushConsumerHandler) Init() {
	ms.msgHandle = make(map[string]fcb)
	ms.msgHandle[config.Config.Kafka.Ms2pschat.Topic] = ms.handleMs2PsChat
	ms.pushConsumerGroup = kfk.NewMConsumerGroup(&kfk.MConsumerGroupConfig{KafkaVersion: sarama.V2_0_0_0,
		OffsetsInitial: sarama.OffsetNewest, IsReturnErr: false}, []string{config.Config.Kafka.Ms2pschat.Topic}, config.Config.Kafka.Ms2pschat.Addr,
		config.Config.Kafka.ConsumerGroupID.MsgToPush)
}

// anson have changed the pbChat.PushMsgDataToMQ to pbPush.PushMsgReq
func (ms *PushConsumerHandler) handleMs2PsChat(msg []byte) {
	//log.NewError("", "msg come from kafka  And push!!!", "msg", string(msg))
	//msgFromMQ := pbChat.PushMsgDataToMQ{}
	msgFromMQ := pbPush.PushMsgReq{}
	if err := proto.Unmarshal(msg, &msgFromMQ); err != nil {
		log.Error("", "push Unmarshal msg err", "msg", string(msg), "err", err.Error())
		return
	}
	//Call push module to send message to the user
	//MsgToUser((*pbPush.PushMsgReq)(&msgFromMQ))
	MsgToUser(&msgFromMQ)
}
func (PushConsumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (PushConsumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (ms *PushConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession,
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
		//log.NewError("", "kafka get info to mysql", "msgTopic", msg.Topic, "msgPartition", msg.Partition, "msg", string(msg.Value))
		if len(msg.Value) != 0 {
			ms.msgHandle[msg.Topic](msg.Value)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
