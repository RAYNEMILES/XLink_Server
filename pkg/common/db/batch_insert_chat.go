package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	pbMsg "Open_IM/pkg/proto/chat"
	"Open_IM/pkg/utils"
	"context"
	"errors"

	go_redis "github.com/go-redis/redis/v9"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (d *DataBases) BatchDeleteChat2DB(userID string, msgList []*pbMsg.MsgDataToMQ, operationID string) {

}

func (d *DataBases) BatchInsertChat2DB(userID string, msgList []*pbMsg.MsgDataToMQ, chatType int, operationID string, currentMaxSeq uint64) error {
	newTime := getCurrentTimestampByMill()
	if len(msgList) > GetSingleGocMsgNum() {
		return errors.New("too large")
	}
	isInit := false
	var remain uint64
	blk0 := uint64(GetSingleGocMsgNum() - 1)
	if currentMaxSeq < uint64(GetSingleGocMsgNum()) {
		remain = blk0 - currentMaxSeq
	} else {
		excludeBlk0 := currentMaxSeq - blk0
		remain = (uint64(GetSingleGocMsgNum()) - (excludeBlk0 % uint64(GetSingleGocMsgNum()))) % uint64(GetSingleGocMsgNum())
	}
	insertCounter := uint64(0)
	msgListToMongo := make([]MsgInfo, 0)
	msgListToMongoNext := make([]MsgInfo, 0)
	seqUid := ""
	seqUidNext := ""
	log.Debug(operationID, "remain ", remain, "insertCounter ", insertCounter, "currentMaxSeq ", currentMaxSeq, userID, len(msgList))
	var err error
	for _, m := range msgList {
		log.Debug(operationID, "msg node ", m.String(), m.MsgData.ClientMsgID)
		currentMaxSeq++
		sMsg := MsgInfo{}
		sMsg.SendTime = m.MsgData.SendTime
		//setting seq here
		m.MsgData.Seq = uint32(currentMaxSeq)

		if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
			return utils.Wrap(err, "")
		}
		if isInit {
			msgListToMongoNext = append(msgListToMongoNext, sMsg)
			seqUidNext = getSeqUid(userID, uint32(currentMaxSeq))
			log.Debug(operationID, "msgListToMongoNext ", seqUidNext, m.MsgData.Seq, m.MsgData.ClientMsgID, insertCounter, remain)
			continue
		}
		if insertCounter < remain {
			msgListToMongo = append(msgListToMongo, sMsg)
			insertCounter++
			seqUid = getSeqUid(userID, uint32(currentMaxSeq))
			log.Debug(operationID, "msgListToMongo ", seqUid, m.MsgData.Seq, m.MsgData.ClientMsgID, insertCounter, remain)
		} else {
			msgListToMongoNext = append(msgListToMongoNext, sMsg)
			seqUidNext = getSeqUid(userID, uint32(currentMaxSeq))
			log.Debug(operationID, "msgListToMongoNext ", seqUidNext, m.MsgData.Seq, m.MsgData.ClientMsgID, insertCounter, remain)
		}
	}

	ctx := context.Background()

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		log.NewError(operationID, "collection is nil")
		return errors.New("collection is nil")
	}

	if seqUid != "" {
		filter := bson.M{"uid": seqUid}
		log.NewDebug(operationID, "filter ", seqUid, "list ", msgListToMongo)
		err := c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": bson.M{"$each": msgListToMongo}}}).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				filter := bson.M{"uid": seqUid}
				sChat := UserChat{}
				sChat.UID = seqUid
				sChat.Msg = msgListToMongo
				log.NewDebug(operationID, "filter ", seqUid, "list ", msgListToMongo)
				if _, err = c.InsertOne(ctx, &sChat); err != nil {
					log.NewError(operationID, "InsertOne failed", filter, err.Error(), sChat)
					return utils.Wrap(err, "")
				}
			} else {
				log.Error(operationID, "FindOneAndUpdate failed ", err.Error(), filter)
				return utils.Wrap(err, "")
			}
		}
	}
	if seqUidNext != "" {
		filter := bson.M{"uid": seqUidNext}
		sChat := UserChat{}
		sChat.UID = seqUidNext
		sChat.Msg = msgListToMongoNext
		log.NewDebug(operationID, "filter ", seqUidNext, "list ", msgListToMongoNext)
		if _, err = c.InsertOne(ctx, &sChat); err != nil {
			log.NewError(operationID, "InsertOne failed", filter, err.Error(), sChat)
			return utils.Wrap(err, "")
		}
	}
	log.NewWarn(operationID, "batch mgo  cost time ", getCurrentTimestampByMill()-newTime, userID, len(msgList))
	return nil
}

// insert messages to redis
func (d *DataBases) BatchInsertChat2Cache(userID string, chatType int, msgList []*pbMsg.MsgDataToMQ, operationID string) (error, uint64) {
	newTime := getCurrentTimestampByMill()
	if len(msgList) > GetSingleGocMsgNum() {
		return errors.New("too large"), 0
	}

	//get max seq from user or group
	var currentMaxSeq uint64
	var err error
	if chatType == constant.SingleChatType {
		currentMaxSeq, err = d.GetUserMaxSeq(userID)
	} else if chatType == constant.GroupChatType {
		currentMaxSeq, err = d.GetGrupMaxSeq(userID)
	}

	if err != nil {
		if err == go_redis.Nil {
			currentMaxSeq = 0
		} else {
			return utils.Wrap(err, ""), 0
		}
	}

	lastMaxSeq := currentMaxSeq

	for _, m := range msgList {
		log.Debug(operationID, "msg node ", m.String(), m.MsgData.ClientMsgID)
		currentMaxSeq++
		sMsg := MsgInfo{}
		sMsg.SendTime = m.MsgData.SendTime
		m.MsgData.Seq = uint32(currentMaxSeq)
	}
	log.Debug(operationID, "SetMessageToCache ", userID, len(msgList))
	err = d.SetMessageToCache(msgList, userID, chatType, operationID)
	if err != nil {
		log.Error(operationID, "setMessageToCache failed, continue ", err.Error(), len(msgList), userID)
	}
	log.NewWarn(operationID, "batch to redis  cost time ", getCurrentTimestampByMill()-newTime, userID, len(msgList))

	if chatType == constant.SingleChatType {
		return utils.Wrap(d.SetUserMaxSeq(userID, uint64(currentMaxSeq)), ""), lastMaxSeq
	} else {
		return utils.Wrap(d.SetGroupMaxSeq(userID, uint64(currentMaxSeq)), ""), lastMaxSeq
	}

}

//func (d *DataBases) BatchInsertChatBoth(userID string, msgList []*pbMsg.MsgDataToMQ, operationID string) (error, uint64) {
//	err, lastMaxSeq := d.BatchInsertChat2Cache(userID, msgList, operationID)
//	if err != nil {
//		log.Error(operationID, "BatchInsertChat2Cache failed ", err.Error(), userID, len(msgList))
//		return err, 0
//	}
//	for {
//		if runtime.NumGoroutine() > 50000 {
//			log.NewWarn(operationID, "too many NumGoroutine ", runtime.NumGoroutine())
//			time.Sleep(10 * time.Millisecond)
//		} else {
//			break
//		}
//	}
//	return nil, lastMaxSeq
//}

func (d *DataBases) BatchInsertChat(userID string, chatType int, msgList []*pbMsg.MsgDataToMQ, operationID string) error {
	newTime := getCurrentTimestampByMill()
	if len(msgList) > GetSingleGocMsgNum() {
		return errors.New("too large")
	}
	isInit := false
	currentMaxSeq, err := d.GetUserMaxSeq(userID)
	if err == nil {

	} else if err == go_redis.Nil {
		isInit = true
		currentMaxSeq = 0
	} else {
		return utils.Wrap(err, "")
	}
	var remain uint64
	//if currentMaxSeq < uint64(GetSingleGocMsgNum()) {
	//	remain = uint64(GetSingleGocMsgNum()-1) - (currentMaxSeq % uint64(GetSingleGocMsgNum()))
	//} else {
	//	remain = uint64(GetSingleGocMsgNum()) - ((currentMaxSeq - (uint64(GetSingleGocMsgNum()) - 1)) % uint64(GetSingleGocMsgNum()))
	//}

	blk0 := uint64(GetSingleGocMsgNum() - 1)
	if currentMaxSeq < uint64(GetSingleGocMsgNum()) {
		remain = blk0 - currentMaxSeq
	} else {
		excludeBlk0 := currentMaxSeq - blk0
		remain = (uint64(GetSingleGocMsgNum()) - (excludeBlk0 % uint64(GetSingleGocMsgNum()))) % uint64(GetSingleGocMsgNum())
	}

	insertCounter := uint64(0)
	msgListToMongo := make([]MsgInfo, 0)
	msgListToMongoNext := make([]MsgInfo, 0)
	seqUid := ""
	seqUidNext := ""
	log.Debug(operationID, "remain ", remain, "insertCounter ", insertCounter, "currentMaxSeq ", currentMaxSeq, userID, len(msgList))
	//4998 remain ==1
	//4999
	for _, m := range msgList {
		log.Debug(operationID, "msg node ", m.String(), m.MsgData.ClientMsgID)
		currentMaxSeq++
		sMsg := MsgInfo{}
		sMsg.SendTime = m.MsgData.SendTime
		m.MsgData.Seq = uint32(currentMaxSeq)
		if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
			return utils.Wrap(err, "")
		}
		if isInit {
			msgListToMongoNext = append(msgListToMongoNext, sMsg)
			seqUidNext = getSeqUid(userID, uint32(currentMaxSeq))
			log.Debug(operationID, "msgListToMongoNext ", seqUidNext, m.MsgData.Seq, m.MsgData.ClientMsgID, insertCounter, remain)
			continue
		}
		if insertCounter < remain {
			msgListToMongo = append(msgListToMongo, sMsg)
			insertCounter++
			seqUid = getSeqUid(userID, uint32(currentMaxSeq))
			log.Debug(operationID, "msgListToMongo ", seqUid, m.MsgData.Seq, m.MsgData.ClientMsgID, insertCounter, remain)
		} else {
			msgListToMongoNext = append(msgListToMongoNext, sMsg)
			seqUidNext = getSeqUid(userID, uint32(currentMaxSeq))
			log.Debug(operationID, "msgListToMongoNext ", seqUidNext, m.MsgData.Seq, m.MsgData.ClientMsgID, insertCounter, remain)
		}
	}
	//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	ctx := context.Background()
	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return errors.New("collection is nil")
	}

	if seqUid != "" {
		filter := bson.M{"uid": seqUid}
		log.NewDebug(operationID, "filter ", seqUid, "list ", msgListToMongo)
		err := c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": bson.M{"$each": msgListToMongo}}}).Err()
		if err != nil {
			log.Error(operationID, "FindOneAndUpdate failed ", err.Error(), filter)
			return utils.Wrap(err, "")
		}
	}
	if seqUidNext != "" {
		filter := bson.M{"uid": seqUidNext}
		sChat := UserChat{}
		sChat.UID = seqUidNext
		sChat.Msg = msgListToMongoNext
		log.NewDebug(operationID, "filter ", seqUidNext, "list ", msgListToMongoNext)
		if _, err = c.InsertOne(ctx, &sChat); err != nil {
			log.NewError(operationID, "InsertOne failed", filter, err.Error(), sChat)
			return utils.Wrap(err, "")
		}
	}
	log.NewWarn(operationID, "batch mgo  cost time ", getCurrentTimestampByMill()-newTime, userID, len(msgList))
	return utils.Wrap(d.SetUserMaxSeq(userID, uint64(currentMaxSeq)), "")
}

//func (d *DataBases)setMessageToCache(msgList []*pbMsg.MsgDataToMQ, uid string) (err error) {
//
//}

func (d *DataBases) GetFromCacheAndInsertDB(msgUserIDPrefix string) {
	//get value from redis

	//batch insert to db
}
