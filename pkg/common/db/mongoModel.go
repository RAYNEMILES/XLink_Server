package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	pbMsg "Open_IM/pkg/proto/chat"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gogo/protobuf/sortkeys"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2"

	pbAdminCMS "Open_IM/pkg/proto/admin_cms"

	//"github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
	"go.mongodb.org/mongo-driver/bson"

	"strconv"
	"time"
)

const privateChat = "msg"
const groupChat = "group_msg"
const cGroup = "group"
const cTag = "tag"
const cSendLog = "send_log"
const cWorkMoment = "work_moment"
const cCommentMsg = "comment_msg"
const cSuperGroup = "super_group"
const cUserToSuperGroup = "user_to_super_group"
const moments = "moments"
const moments_like = "moments_like"
const moments_comments = "moments_comments"
const singleGocMsgNum = 5000
const operationLogs = "operation_logs"
const article_like = "article_like"
const article_comment_like = "article_comment_like"
const favorites = "favorites"
const favoriteMedia = "favorite_media"

func GetSingleGocMsgNum() int {
	return singleGocMsgNum
}

type MsgInfo struct {
	SendTime int64
	Msg      []byte
}

type UserChat struct {
	UID string
	Msg []MsgInfo
}

type GroupMember_x struct {
	GroupID string
	UIDList []string
}

func (d *DataBases) GetMinSeqFromMongo(uid string) (MinSeq uint32, err error) {
	return 1, nil
	//var i, NB uint32
	//var seqUid string
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return MinSeq, errors.New("session == nil")
	//}
	//defer session.Close()
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	//MaxSeq, err := d.GetUserMaxSeq(uid)
	//if err != nil && err != redis.ErrNil {
	//	return MinSeq, err
	//}
	//NB = uint32(MaxSeq / singleGocMsgNum)
	//for i = 0; i <= NB; i++ {
	//	seqUid = indexGen(uid, i)
	//	n, err := c.Find(bson.M{"uid": seqUid}).Count()
	//	if err == nil && n != 0 {
	//		if i == 0 {
	//			MinSeq = 1
	//		} else {
	//			MinSeq = uint32(i * singleGocMsgNum)
	//		}
	//		break
	//	}
	//}
	//return MinSeq, nil
}

func (d *DataBases) GetMinSeqFromMongo2(uid string) (MinSeq uint32, err error) {
	return 1, nil
}

// deleteMsgByLogic
func (d *DataBases) DelMsgBySeqList(userID string, chatType int, seqList []uint32, operationID string) (totalUnexistSeqList []uint32, err error) {
	log.Debug(operationID, utils.GetSelfFuncName(), "args ", userID, seqList)
	sortkeys.Uint32s(seqList)
	suffixUserID2SubSeqList := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(userID, seqList)

	lock := sync.Mutex{}
	var wg sync.WaitGroup
	wg.Add(len(suffixUserID2SubSeqList))
	for k, v := range suffixUserID2SubSeqList {
		go func(suffixUserID string, subSeqList []uint32, operationID string) {
			defer wg.Done()
			unexistSeqList, err := d.DelMsgBySeqListInOneDoc(suffixUserID, chatType, subSeqList, operationID)
			if err != nil {
				log.Error(operationID, "DelMsgBySeqListInOneDoc failed ", err.Error(), suffixUserID, subSeqList)
				return
			}
			lock.Lock()
			totalUnexistSeqList = append(totalUnexistSeqList, unexistSeqList...)
			lock.Unlock()
		}(k, v, operationID)
	}
	return totalUnexistSeqList, err
}

func (d *DataBases) DelMsgBySeqListInOneDoc(suffixUserID string, chatType int, seqList []uint32, operationID string) ([]uint32, error) {
	log.Debug(operationID, utils.GetSelfFuncName(), "args ", suffixUserID, seqList)
	seqMsgList, indexList, unexistSeqList, err := d.GetMsgAndIndexBySeqListInOneMongo2(suffixUserID, chatType, seqList, operationID)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	for i, v := range seqMsgList {
		if err := d.ReplaceMsgByIndex(suffixUserID, chatType, v, operationID, indexList[i]); err != nil {
			return nil, utils.Wrap(err, "")
		}
	}
	return unexistSeqList, nil
}

// deleteMsgByLogic
func (d *DataBases) DelMsgLogic(uid string, chatType int, seqList []uint32, operationID string) error {
	sortkeys.Uint32s(seqList)
	seqMsgs, err := d.GetMsgBySeqListMongo2(uid, chatType, seqList, operationID)
	if err != nil {
		return utils.Wrap(err, "")
	}
	for _, seqMsg := range seqMsgs {
		log.NewDebug(operationID, utils.GetSelfFuncName(), *seqMsg)
		seqMsg.Status = constant.MsgDeleted
		if err = d.ReplaceMsgBySeq(uid, chatType, seqMsg, operationID); err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), "ReplaceMsgListBySeq error", err.Error())
		}
	}
	return nil
}

func (d *DataBases) ReplaceMsgByIndex(suffixUserID string, chatType int, msg *open_im_sdk.MsgData, operationID string, seqIndex int) error {
	log.NewInfo(operationID, utils.GetSelfFuncName(), suffixUserID, *msg)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return errors.New("collection is nil")
	}

	s := fmt.Sprintf("msg.%d.msg", seqIndex)
	log.NewDebug(operationID, utils.GetSelfFuncName(), seqIndex, s)
	msg.Status = constant.MsgDeleted
	bytes, err := proto.Marshal(msg)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "proto marshal failed ", err.Error(), msg.String())
		return utils.Wrap(err, "")
	}
	updateResult, err := c.UpdateOne(ctx, bson.M{"uid": suffixUserID}, bson.M{"$set": bson.M{s: bytes}})
	log.NewInfo(operationID, utils.GetSelfFuncName(), updateResult)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "UpdateOne", err.Error())
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) ReplaceMsgBySeq(uid string, chatType int, msg *open_im_sdk.MsgData, operationID string) error {
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, *msg)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return errors.New("collection is nil")
	}

	uid = getSeqUid(uid, msg.Seq)
	seqIndex := getMsgIndex(msg.Seq)
	s := fmt.Sprintf("msg.%d.msg", seqIndex)
	log.NewDebug(operationID, utils.GetSelfFuncName(), seqIndex, s)
	bytes, err := proto.Marshal(msg)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "proto marshal", err.Error())
		return utils.Wrap(err, "")
	}

	updateResult, err := c.UpdateOne(
		ctx, bson.M{"uid": uid},
		bson.M{"$set": bson.M{s: bytes}})
	log.NewInfo(operationID, utils.GetSelfFuncName(), updateResult)
	if err != nil {
		log.NewError(operationID, utils.GetSelfFuncName(), "UpdateOne", err.Error())
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) GetMsgBySeqList(uid string, chatType int, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	log.NewInfo(operationID, utils.GetSelfFuncName(), uid, seqList)
	var hasSeqList []uint32
	singleCount := 0

	session := d.mgoSession.Clone()
	if session == nil {
		return nil, errors.New("session == nil")
	}
	defer session.Close()

	var c *mgo.Collection
	if chatType == constant.SingleChatType {
		c = session.DB(config.Config.Mongo.DBDatabase).C(privateChat)
	} else if chatType == constant.GroupChatType {
		c = session.DB(config.Config.Mongo.DBDatabase).C(groupChat)
	}
	if c == nil {
		return nil, errors.New("collection is nil")
	}

	m := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(uid, seqList)
	sChat := UserChat{}
	for seqUid, value := range m {
		if err = c.Find(bson.M{"uid": seqUid}).One(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, value, uid, seqList, err.Error())
			continue
		}
		singleCount = 0
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, value, uid, seqList, err.Error())
				return nil, err
			}
			if isContainInt32(msg.Seq, value) {
				seqMsg = append(seqMsg, msg)
				hasSeqList = append(hasSeqList, msg.Seq)
				singleCount++
				if singleCount == len(value) {
					break
				}
			}
		}
	}
	if len(hasSeqList) != len(seqList) {
		var diff []uint32
		diff = utils.Difference(hasSeqList, seqList)
		exceptionMSg := genExceptionMessageBySeqList(diff)
		seqMsg = append(seqMsg, exceptionMSg...)

	}
	return seqMsg, nil
}

func (d *DataBases) GetMsgBySeqListMongo2(uid string, chatType int, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	var hasSeqList []uint32
	singleCount := 0
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return nil, errors.New("collection is nil")
	}

	m := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(uid, seqList)
	sChat := UserChat{}
	for seqUid, value := range m {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqUid", seqUid, value, uid, seqList, err.Error())
			continue
		}
		singleCount = 0
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, value, uid, seqList, err.Error())
				return nil, err
			}
			if isContainInt32(msg.Seq, value) {
				seqMsg = append(seqMsg, msg)
				hasSeqList = append(hasSeqList, msg.Seq)
				singleCount++
				if singleCount == len(value) {
					break
				}
			}
		}
	}
	if len(hasSeqList) != len(seqList) {
		var diff []uint32
		diff = utils.Difference(hasSeqList, seqList)
		exceptionMSg := genExceptionMessageBySeqList(diff)
		seqMsg = append(seqMsg, exceptionMSg...)

	}
	return seqMsg, nil
}
func (d *DataBases) GetSuperGroupMsgBySeqListMongo(groupID string, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, err error) {
	var hasSeqList []uint32
	singleCount := 0
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)

	m := func(uid string, seqList []uint32) map[string][]uint32 {
		t := make(map[string][]uint32)
		for i := 0; i < len(seqList); i++ {
			seqUid := getSeqUid(uid, seqList[i])
			if value, ok := t[seqUid]; !ok {
				var temp []uint32
				t[seqUid] = append(temp, seqList[i])
			} else {
				t[seqUid] = append(value, seqList[i])
			}
		}
		return t
	}(groupID, seqList)
	sChat := UserChat{}
	for seqUid, value := range m {
		if err = c.FindOne(ctx, bson.M{"uid": seqUid}).Decode(&sChat); err != nil {
			log.NewError(operationID, "not find seqGroupID", seqUid, value, groupID, seqList, err.Error())
			continue
		}
		singleCount = 0
		for i := 0; i < len(sChat.Msg); i++ {
			msg := new(open_im_sdk.MsgData)
			if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
				log.NewError(operationID, "Unmarshal err", seqUid, value, groupID, seqList, err.Error())
				return nil, err
			}
			if isContainInt32(msg.Seq, value) {
				seqMsg = append(seqMsg, msg)
				hasSeqList = append(hasSeqList, msg.Seq)
				singleCount++
				if singleCount == len(value) {
					break
				}
			}
		}
	}
	if len(hasSeqList) != len(seqList) {
		var diff []uint32
		diff = utils.Difference(hasSeqList, seqList)
		exceptionMSg := genExceptionSuperGroupMessageBySeqList(diff, groupID)
		seqMsg = append(seqMsg, exceptionMSg...)

	}
	return seqMsg, nil
}

func (d *DataBases) GetMsgAndIndexBySeqListInOneMongo2(suffixUserID string, chatType int, seqList []uint32, operationID string) (seqMsg []*open_im_sdk.MsgData, indexList []int, unexistSeqList []uint32, err error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return nil, nil, nil, errors.New("collection is nil")
	}

	sChat := UserChat{}
	if err = c.FindOne(ctx, bson.M{"uid": suffixUserID}).Decode(&sChat); err != nil {
		log.NewError(operationID, "not find seqUid", suffixUserID, err.Error())
		return nil, nil, nil, utils.Wrap(err, "")
	}
	singleCount := 0
	var hasSeqList []uint32
	for i := 0; i < len(sChat.Msg); i++ {
		msg := new(open_im_sdk.MsgData)
		if err = proto.Unmarshal(sChat.Msg[i].Msg, msg); err != nil {
			log.NewError(operationID, "Unmarshal err", msg.String(), err.Error())
			return nil, nil, nil, err
		}
		if isContainInt32(msg.Seq, seqList) {
			indexList = append(indexList, i)
			seqMsg = append(seqMsg, msg)
			hasSeqList = append(hasSeqList, msg.Seq)
			singleCount++
			if singleCount == len(seqList) {
				break
			}
		}
	}
	for _, i := range seqList {
		if isContainInt32(i, hasSeqList) {
			continue
		}
		unexistSeqList = append(unexistSeqList, i)
	}
	return seqMsg, indexList, unexistSeqList, nil
}

func genExceptionMessageBySeqList(seqList []uint32) (exceptionMsg []*open_im_sdk.MsgData) {
	for _, v := range seqList {
		msg := new(open_im_sdk.MsgData)
		msg.Seq = v
		exceptionMsg = append(exceptionMsg, msg)
	}
	return exceptionMsg
}

func genExceptionSuperGroupMessageBySeqList(seqList []uint32, groupID string) (exceptionMsg []*open_im_sdk.MsgData) {
	for _, v := range seqList {
		msg := new(open_im_sdk.MsgData)
		msg.Seq = v
		msg.GroupID = groupID
		msg.SessionType = constant.SuperGroupChatType
		exceptionMsg = append(exceptionMsg, msg)
	}
	return exceptionMsg
}

func (d *DataBases) SaveUserChatMongo2(uid string, chatType int, sendTime int64, m *pbMsg.MsgDataToDB) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return errors.New("collection is nil")
	}

	newTime := getCurrentTimestampByMill()
	operationID := ""
	seqUid := getSeqUid(uid, m.MsgData.Seq)
	filter := bson.M{"uid": seqUid}
	var err error
	sMsg := MsgInfo{}
	sMsg.SendTime = sendTime
	if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
		return utils.Wrap(err, "")
	}
	err = c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": sMsg}}).Err()
	log.NewWarn(operationID, "get mgoSession cost time", getCurrentTimestampByMill()-newTime)
	if err != nil {
		sChat := UserChat{}
		sChat.UID = seqUid
		sChat.Msg = append(sChat.Msg, sMsg)
		if _, err = c.InsertOne(ctx, &sChat); err != nil {
			log.NewDebug(operationID, "InsertOne failed", filter)
			return utils.Wrap(err, "")
		}
	} else {
		log.NewDebug(operationID, "FindOneAndUpdate ok", filter)
	}

	log.NewDebug(operationID, "find mgo uid cost time", getCurrentTimestampByMill()-newTime)
	return nil
}

//
//func (d *DataBases) SaveUserChatListMongo2(uid string, sendTime int64, msgList []*pbMsg.MsgDataToDB) error {
//	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
//	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cChat)
//	newTime := getCurrentTimestampByMill()
//	operationID := ""
//	seqUid := ""
//	msgListToMongo := make([]MsgInfo, 0)
//
//	for _, m := range msgList {
//		seqUid = getSeqUid(uid, m.MsgData.Seq)
//		var err error
//		sMsg := MsgInfo{}
//		sMsg.SendTime = sendTime
//		if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
//			return utils.Wrap(err, "")
//		}
//		msgListToMongo = append(msgListToMongo, sMsg)
//	}
//
//	filter := bson.M{"uid": seqUid}
//	log.NewDebug(operationID, "filter ", seqUid)
//	err := c.FindOneAndUpdate(ctx, filter, bson.M{"$push": bson.M{"msg": bson.M{"$each": msgListToMongo}}}).Err()
//	log.NewWarn(operationID, "get mgoSession cost time", getCurrentTimestampByMill()-newTime)
//	if err != nil {
//		sChat := UserChat{}
//		sChat.UID = seqUid
//		sChat.Msg = msgListToMongo
//
//		if _, err = c.InsertOne(ctx, &sChat); err != nil {
//			log.NewError(operationID, "InsertOne failed", filter, err.Error(), sChat)
//			return utils.Wrap(err, "")
//		}
//	} else {
//		log.NewDebug(operationID, "FindOneAndUpdate ok", filter)
//	}
//
//	log.NewDebug(operationID, "find mgo uid cost time", getCurrentTimestampByMill()-newTime)
//	return nil
//}

func (d *DataBases) SaveUserChat(uid string, chatType int, sendTime int64, m *pbMsg.MsgDataToDB) error {
	var seqUid string
	newTime := getCurrentTimestampByMill()
	session := d.mgoSession.Clone()
	if session == nil {
		return errors.New("session == nil")
	}
	defer session.Close()
	log.NewDebug("", "get mgoSession cost time", getCurrentTimestampByMill()-newTime)

	var c *mgo.Collection
	if chatType == constant.SingleChatType {
		c = session.DB(config.Config.Mongo.DBDatabase).C(privateChat)
	} else if chatType == constant.GroupChatType {
		c = session.DB(config.Config.Mongo.DBDatabase).C(groupChat)
	}
	if c == nil {
		return errors.New("collection is nil")
	}

	seqUid = getSeqUid(uid, m.MsgData.Seq)
	n, err := c.Find(bson.M{"uid": seqUid}).Count()
	if err != nil {
		return err
	}
	log.NewDebug("", "find mgo uid cost time", getCurrentTimestampByMill()-newTime)
	sMsg := MsgInfo{}
	sMsg.SendTime = sendTime
	if sMsg.Msg, err = proto.Marshal(m.MsgData); err != nil {
		return err
	}
	if n == 0 {
		sChat := UserChat{}
		sChat.UID = seqUid
		sChat.Msg = append(sChat.Msg, sMsg)
		err = c.Insert(&sChat)
		if err != nil {
			return err
		}
	} else {
		err = c.Update(bson.M{"uid": seqUid}, bson.M{"$push": bson.M{"msg": sMsg}})
		if err != nil {
			return err
		}
	}
	log.NewDebug("", "insert mgo data cost time", getCurrentTimestampByMill()-newTime)
	return nil
}

func (d *DataBases) DelUserChat(uid string) error {
	return nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	//
	//delTime := time.Now().Unix() - int64(config.Config.Mongo.DBRetainChatRecords)*24*3600
	//if err := c.Update(bson.M{"uid": uid}, bson.M{"$pull": bson.M{"msg": bson.M{"sendtime": bson.M{"$lte": delTime}}}}); err != nil {
	//	return err
	//}
	//
	//return nil
}

func (d *DataBases) DelUserChatMongo2(uid string, chatType int) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)

	var c *mongo.Collection
	if chatType == constant.SingleChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(privateChat)
	} else if chatType == constant.GroupChatType {
		c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(groupChat)
	}
	if c == nil {
		return errors.New("collection is nil")
	}

	filter := bson.M{"uid": uid}

	delTime := time.Now().Unix() - int64(config.Config.Mongo.DBRetainChatRecords)*24*3600
	if _, err := c.UpdateOne(ctx, filter, bson.M{"$pull": bson.M{"msg": bson.M{"sendtime": bson.M{"$lte": delTime}}}}); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) MgoUserCount() (int, error) {
	return 0, nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return 0, errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	//
	//return c.Find(nil).Count()
}

func (d *DataBases) MgoSkipUID(count int) (string, error) {
	return "", nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return "", errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cChat)
	//
	//sChat := UserChat{}
	//c.Find(nil).Skip(count).Limit(1).One(&sChat)
	//return sChat.UID, nil
}

func (d *DataBases) GetGroupMember(groupID string) []string {
	return nil
	//groupInfo := GroupMember_x{}
	//groupInfo.GroupID = groupID
	//groupInfo.UIDList = make([]string, 0)
	//
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return groupInfo.UIDList
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cGroup)
	//
	//if err := c.Find(bson.M{"groupid": groupInfo.GroupID}).One(&groupInfo); err != nil {
	//	return groupInfo.UIDList
	//}
	//
	//return groupInfo.UIDList
}

func (d *DataBases) AddGroupMember(groupID, uid string) error {
	return nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cGroup)
	//
	//n, err := c.Find(bson.M{"groupid": groupID}).Count()
	//if err != nil {
	//	return err
	//}
	//
	//if n == 0 {
	//	groupInfo := GroupMember_x{}
	//	groupInfo.GroupID = groupID
	//	groupInfo.UIDList = append(groupInfo.UIDList, uid)
	//	err = c.Insert(&groupInfo)
	//	if err != nil {
	//		return err
	//	}
	//} else {
	//	err = c.Update(bson.M{"groupid": groupID}, bson.M{"$addToSet": bson.M{"uidlist": uid}})
	//	if err != nil {
	//		return err
	//	}
	//}
	//
	//return nil
}

func (d *DataBases) DelGroupMember(groupID, uid string) error {
	return nil
	//session := d.mgoSession.Clone()
	//if session == nil {
	//	return errors.New("session == nil")
	//}
	//defer session.Close()
	//
	//c := session.DB(config.Config.Mongo.DBDatabase).C(cGroup)
	//
	//if err := c.Update(bson.M{"groupid": groupID}, bson.M{"$pull": bson.M{"uidlist": uid}}); err != nil {
	//	return err
	//}
	//
	//return nil
}

type Tag struct {
	UserID   string   `bson:"user_id"`
	TagID    string   `bson:"tag_id"`
	TagName  string   `bson:"tag_name"`
	UserList []string `bson:"user_list"`
}

func (d *DataBases) GetUserTags(userID string) ([]Tag, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	var tags []Tag
	cursor, err := c.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return tags, err
	}
	if err = cursor.All(ctx, &tags); err != nil {
		return tags, err
	}
	return tags, nil
}

func (d *DataBases) CreateTag(userID, tagName string, userList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	tagID := generateTagID(tagName, userID)
	tag := Tag{
		UserID:   userID,
		TagID:    tagID,
		TagName:  tagName,
		UserList: userList,
	}
	_, err := c.InsertOne(ctx, tag)
	return err
}

func (d *DataBases) GetTagByID(userID, tagID string) (Tag, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	var tag Tag
	err := c.FindOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}).Decode(&tag)
	return tag, err
}

func (d *DataBases) DeleteTag(userID, tagID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	_, err := c.DeleteOne(ctx, bson.M{"user_id": userID, "tag_id": tagID})
	return err
}

func (d *DataBases) SetTag(userID, tagID, newName string, increaseUserIDList []string, reduceUserIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	var tag Tag
	if err := c.FindOne(ctx, bson.M{"tag_id": tagID, "user_id": userID}).Decode(&tag); err != nil {
		return err
	}
	if newName != "" {
		_, err := c.UpdateOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}, bson.M{"$set": bson.M{"tag_name": newName}})
		if err != nil {
			return err
		}
	}
	tag.UserList = append(tag.UserList, increaseUserIDList...)
	tag.UserList = utils.RemoveRepeatedStringInList(tag.UserList)
	for _, v := range reduceUserIDList {
		for i2, v2 := range tag.UserList {
			if v == v2 {
				tag.UserList[i2] = ""
			}
		}
	}
	var newUserList []string
	for _, v := range tag.UserList {
		if v != "" {
			newUserList = append(newUserList, v)
		}
	}
	_, err := c.UpdateOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}, bson.M{"$set": bson.M{"user_list": newUserList}})
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBases) GetUserIDListByTagID(userID, tagID string) ([]string, error) {
	var tag Tag
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cTag)
	_ = c.FindOne(ctx, bson.M{"user_id": userID, "tag_id": tagID}).Decode(&tag)
	return tag.UserList, nil
}

type TagUser struct {
	UserID   string `bson:"user_id"`
	UserName string `bson:"user_name"`
}

type TagSendLog struct {
	UserList         []TagUser `bson:"tag_list"`
	SendID           string    `bson:"send_id"`
	SenderPlatformID int32     `bson:"sender_platform_id"`
	Content          string    `bson:"content"`
	SendTime         int64     `bson:"send_time"`
}

func (d *DataBases) SaveTagSendLog(tagSendLog *TagSendLog) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSendLog)
	_, err := c.InsertOne(ctx, tagSendLog)
	return err
}

func (d *DataBases) GetTagSendLogs(userID string, showNumber, pageNumber int32) ([]TagSendLog, error) {
	var tagSendLogs []TagSendLog
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSendLog)
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"send_time": -1})
	cursor, err := c.Find(ctx, bson.M{"send_id": userID}, findOpts)
	if err != nil {
		return tagSendLogs, err
	}
	err = cursor.All(ctx, &tagSendLogs)
	if err != nil {
		return tagSendLogs, err
	}
	return tagSendLogs, nil
}

type WorkMoment struct {
	WorkMomentID         string            `bson:"work_moment_id"`
	UserID               string            `bson:"user_id"`
	UserName             string            `bson:"user_name"`
	FaceURL              string            `bson:"face_url"`
	Content              string            `bson:"content"`
	LikeUserList         []*WorkMomentUser `bson:"like_user_list"`
	AtUserList           []*WorkMomentUser `bson:"at_user_list"`
	PermissionUserList   []*WorkMomentUser `bson:"permission_user_list"`
	Comments             []*Comment        `bson:"comments"`
	PermissionUserIDList []string          `bson:"permission_user_id_list"`
	Permission           int32             `bson:"permission"`
	CreateTime           int32             `bson:"create_time"`
}

type WorkMomentUser struct {
	UserID   string `bson:"user_id"`
	UserName string `bson:"user_name"`
}

type Comment struct {
	UserID        string `bson:"user_id" json:"user_id"`
	UserName      string `bson:"user_name" json:"user_name"`
	ReplyUserID   string `bson:"reply_user_id" json:"reply_user_id"`
	ReplyUserName string `bson:"reply_user_name" json:"reply_user_name"`
	ContentID     string `bson:"content_id" json:"content_id"`
	Content       string `bson:"content" json:"content"`
	CreateTime    int32  `bson:"create_time" json:"create_time"`
}

func (d *DataBases) CreateOneWorkMoment(workMoment *WorkMoment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	workMomentID := generateWorkMomentID(workMoment.UserID)
	workMoment.WorkMomentID = workMomentID
	workMoment.CreateTime = int32(time.Now().Unix())
	_, err := c.InsertOne(ctx, workMoment)
	return err
}

func (d *DataBases) DeleteOneWorkMoment(workMomentID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	_, err := c.DeleteOne(ctx, bson.M{"work_moment_id": workMomentID})
	return err
}

func (d *DataBases) DeleteComment(workMomentID, contentID, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	_, err := c.UpdateOne(ctx, bson.D{{"work_moment_id", workMomentID},
		{"$or", bson.A{
			bson.D{{"user_id", opUserID}},
			bson.D{{"comments", bson.M{"$elemMatch": bson.M{"user_id": opUserID}}}},
		},
		}}, bson.M{"$pull": bson.M{"comments": bson.M{"content_id": contentID}}})
	return err
}

func (d *DataBases) GetWorkMomentByID(workMomentID string) (*WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	workMoment := &WorkMoment{}
	err := c.FindOne(ctx, bson.M{"work_moment_id": workMomentID}).Decode(workMoment)
	return workMoment, err
}

func (d *DataBases) LikeOneWorkMoment(likeUserID, userName, workMomentID string) (*WorkMoment, bool, error) {
	workMoment, err := d.GetWorkMomentByID(workMomentID)
	if err != nil {
		return nil, false, err
	}
	var isAlreadyLike bool
	for i, user := range workMoment.LikeUserList {
		if likeUserID == user.UserID {
			isAlreadyLike = true
			workMoment.LikeUserList = append(workMoment.LikeUserList[0:i], workMoment.LikeUserList[i+1:]...)
		}
	}
	if !isAlreadyLike {
		workMoment.LikeUserList = append(workMoment.LikeUserList, &WorkMomentUser{UserID: likeUserID, UserName: userName})
	}
	log.NewDebug("", utils.GetSelfFuncName(), workMoment)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	_, err = c.UpdateOne(ctx, bson.M{"work_moment_id": workMomentID}, bson.M{"$set": bson.M{"like_user_list": workMoment.LikeUserList}})
	return workMoment, !isAlreadyLike, err
}

func (d *DataBases) SetUserWorkMomentsLevel(userID string, level int32) error {
	return nil
}

func (d *DataBases) CommentOneWorkMoment(comment *Comment, workMomentID string) (WorkMoment, error) {
	comment.ContentID = generateWorkMomentCommentID(workMomentID)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMoment WorkMoment
	err := c.FindOneAndUpdate(ctx, bson.M{"work_moment_id": workMomentID}, bson.M{"$push": bson.M{"comments": comment}}).Decode(&workMoment)
	return workMoment, err
}

func (d *DataBases) GetUserSelfWorkMoments(userID string, showNumber, pageNumber int32) ([]WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMomentList []WorkMoment
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"create_time": -1})
	result, err := c.Find(ctx, bson.M{"user_id": userID}, findOpts)
	if err != nil {
		return workMomentList, nil
	}
	err = result.All(ctx, &workMomentList)
	return workMomentList, err
}

func (d *DataBases) GetUserWorkMoments(opUserID, userID string, showNumber, pageNumber int32) ([]WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMomentList []WorkMoment
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"create_time": -1})
	result, err := c.Find(ctx, bson.D{ // 等价条件: select * from
		{"user_id", userID},
		{"$or", bson.A{
			bson.D{{"permission", constant.WorkMomentPermissionCantSee}, {"permission_user_id_list", bson.D{{"$nin", bson.A{userID}}}}},
			bson.D{{"permission", constant.WorkMomentPermissionCanSee}, {"permission_user_id_list", bson.D{{"$in", bson.A{userID}}}}},
			bson.D{{"permission", constant.WorkMomentPublic}},
		}},
	}, findOpts)
	if err != nil {
		return workMomentList, nil
	}
	err = result.All(ctx, &workMomentList)
	return workMomentList, err
}

func (d *DataBases) GetUserFriendWorkMoments(showNumber, pageNumber int32, userID string) ([]WorkMoment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cWorkMoment)
	var workMomentList []WorkMoment
	findOpts := options.Find().SetLimit(int64(showNumber)).SetSkip(int64(showNumber) * (int64(pageNumber) - 1)).SetSort(bson.M{"create_time": -1})
	result, err := c.Find(ctx, bson.D{
		{"$or", bson.A{
			bson.D{{"user_id", userID}}, //self
			bson.D{
				{"$or", bson.A{
					bson.D{{"permission", constant.WorkMomentPermissionCantSee}, {"permission_user_id_list", bson.D{{"$nin", bson.A{userID}}}}},
					bson.D{{"permission", constant.WorkMomentPermissionCanSee}, {"permission_user_id_list", bson.D{{"$in", bson.A{userID}}}}},
					bson.D{{"permission", constant.WorkMomentPublic}},
				}}},
		},
		},
	}, findOpts)
	if err != nil {
		return workMomentList, err
	}
	err = result.All(ctx, &workMomentList)
	return workMomentList, err
}

type SuperGroup struct {
	GroupID string `bson:"group_id"`
	//MemberNumCount int      `bson:"member_num_count"`
	MemberIDList []string `bson:"member_id_list"`
}

type UserToSuperGroup struct {
	UserID      string   `bson:"user_id"`
	GroupIDList []string `bson:"group_id_list"`
}

func (d *DataBases) CreateSuperGroup(groupID string, initMemberIDList []string, memberNumCount int) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	superGroup := SuperGroup{
		GroupID:      groupID,
		MemberIDList: initMemberIDList,
	}
	_, err = c.InsertOne(sCtx, superGroup)
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	var users []UserToSuperGroup
	for _, v := range initMemberIDList {
		users = append(users, UserToSuperGroup{
			UserID: v,
		})
	}
	upsert := true
	opts := &options.UpdateOptions{
		Upsert: &upsert,
	}
	c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	_, err = c.UpdateMany(sCtx, bson.M{"user_id": bson.M{"$in": initMemberIDList}}, bson.M{"$addToSet": bson.M{"group_id_list": groupID}}, opts)
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	session.CommitTransaction(ctx)
	return err
}

func (d *DataBases) GetSuperGroup(groupID string) (SuperGroup, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	superGroup := SuperGroup{}
	err := c.FindOne(ctx, bson.M{"group_id": groupID}).Decode(&superGroup)
	return superGroup, err
}

func (d *DataBases) AddUserToSuperGroup(groupID string, userIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	if err != nil {
		return utils.Wrap(err, "start transaction failed")
	}
	_, err = c.UpdateOne(sCtx, bson.M{"group_id": groupID}, bson.M{"$addToSet": bson.M{"member_id_list": bson.M{"$each": userIDList}}})
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	var users []UserToSuperGroup
	for _, v := range userIDList {
		users = append(users, UserToSuperGroup{
			UserID: v,
		})
	}
	upsert := true
	opts := &options.UpdateOptions{
		Upsert: &upsert,
	}
	for _, userID := range userIDList {
		_, err = c.UpdateOne(sCtx, bson.M{"user_id": userID}, bson.M{"$addToSet": bson.M{"group_id_list": groupID}}, opts)
		if err != nil {
			session.AbortTransaction(ctx)
			return utils.Wrap(err, "transaction failed")
		}
	}
	session.CommitTransaction(ctx)
	return err
}

func (d *DataBases) RemoverUserFromSuperGroup(groupID string, userIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	_, err = c.UpdateOne(ctx, bson.M{"group_id": groupID}, bson.M{"$pull": bson.M{"member_id_list": bson.M{"$in": userIDList}}})
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	err = d.RemoveGroupFromUser(ctx, sCtx, groupID, userIDList)
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	session.CommitTransaction(ctx)
	return err
}

func (d *DataBases) GetSuperGroupByUserID(userID string) (UserToSuperGroup, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	var user UserToSuperGroup
	return user, c.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
}

func (d *DataBases) DeleteSuperGroup(groupID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cSuperGroup)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	superGroup := &SuperGroup{}
	result := c.FindOneAndDelete(sCtx, bson.M{"group_id": groupID})
	err = result.Decode(superGroup)
	if err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	if err = d.RemoveGroupFromUser(ctx, sCtx, groupID, superGroup.MemberIDList); err != nil {
		session.AbortTransaction(ctx)
		return utils.Wrap(err, "transaction failed")
	}
	session.CommitTransaction(ctx)
	return nil
}

func (d *DataBases) RemoveGroupFromUser(ctx, sCtx context.Context, groupID string, userIDList []string) error {
	var users []UserToSuperGroup
	for _, v := range userIDList {
		users = append(users, UserToSuperGroup{
			UserID: v,
		})
	}
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(cUserToSuperGroup)
	_, err := c.UpdateOne(sCtx, bson.M{"user_id": bson.M{"$in": userIDList}}, bson.M{"$pull": bson.M{"group_id_list": groupID}})
	if err != nil {
		return utils.Wrap(err, "UpdateOne transaction failed")
	}
	return err
}

func generateTagID(tagName, userID string) string {
	return utils.Md5(tagName + userID + strconv.Itoa(rand.Int()) + time.Now().String())
}

func generateWorkMomentID(userID string) string {
	return utils.Md5(userID + strconv.Itoa(rand.Int()) + time.Now().String())
}

func generateWorkMomentCommentID(workMomentID string) string {
	return utils.Md5(workMomentID + strconv.Itoa(rand.Int()) + time.Now().String())
}

func getCurrentTimestampByMill() int64 {
	return time.Now().UnixNano() / 1e6
}
func GetCurrentTimestampByMill() int64 {
	return time.Now().UnixNano() / 1e6
}

func getSeqUid(uid string, seq uint32) string {
	seqSuffix := seq / singleGocMsgNum
	return indexGen(uid, seqSuffix)
}

func getSeqUserIDList(userID string, maxSeq uint32) []string {
	seqMaxSuffix := maxSeq / singleGocMsgNum
	var seqUserIDList []string
	for i := 0; i <= int(seqMaxSuffix); i++ {
		seqUserID := indexGen(userID, uint32(i))
		seqUserIDList = append(seqUserIDList, seqUserID)
	}
	return seqUserIDList
}

func getSeqSuperGroupID(groupID string, seq uint32) string {
	seqSuffix := seq / singleGocMsgNum
	return superGroupIndexGen(groupID, seqSuffix)
}

func GetSeqUid(uid string, seq uint32) string {
	return getSeqUid(uid, seq)
}

func getMsgIndex(seq uint32) int {
	seqSuffix := seq / singleGocMsgNum
	var index uint32
	if seqSuffix == 0 {
		index = (seq - seqSuffix*singleGocMsgNum) - 1
	} else {
		index = seq - seqSuffix*singleGocMsgNum
	}
	return int(index)
}

func isContainInt32(target uint32, List []uint32) bool {
	for _, element := range List {
		if target == element {
			return true
		}
	}
	return false
}

func isNotContainInt32(target uint32, List []uint32) bool {
	for _, i := range List {
		if i == target {
			return false
		}
	}
	return true
}

func indexGen(uid string, seqSuffix uint32) string {
	return uid + ":" + strconv.FormatInt(int64(seqSuffix), 10)
}
func superGroupIndexGen(groupID string, seqSuffix uint32) string {
	return "super_group_" + groupID + ":" + strconv.FormatInt(int64(seqSuffix), 10)
}

func (d *DataBases) CleanUpUserMsgFromMongo(userID string, chatType int, operationID string) error {
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

	maxSeq, err := d.GetUserMaxSeq(userID)
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return utils.Wrap(err, "")
	}

	seqUsers := getSeqUserIDList(userID, uint32(maxSeq))
	log.Error(operationID, "getSeqUserIDList", seqUsers)
	_, err = c.DeleteMany(ctx, bson.M{"uid": bson.M{"$in": seqUsers}})
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return utils.Wrap(err, "")
}

func (d *DataBases) AddMoment(moment Moment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(moment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	_, err = c.InsertOne(ctx, moment)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) GetMoment(momentID primitive.ObjectID) (*Moment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": momentID,
		},
	}
	var moment Moment
	momentResult := c.FindOne(ctx, filter)
	err = momentResult.Decode(&moment)

	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return &moment, err
}

func (d *DataBases) GetMomentsMediaByID(userID string, lastCount int64) ([]Moment, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var OrFilter = []bson.M{
		{"m_content_images_array": bson.M{"$ne": "null"}},
		{"m_content_videos_array": bson.M{"$ne": "null"}},
	}

	filter := bson.M{
		"user_id": bson.M{
			"$eq": userID,
		},
		"$or":         OrFilter,
		"delete_time": 0,
	}
	count, err := c.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, utils.Wrap(err, "No Record found")
	}

	options := options.Find()
	options.SetSort(bson.D{{"m_create_time", -1}})
	options.SetLimit(lastCount)
	options.SetSkip(0)

	var moments []Moment
	momentResult, errDB := c.Find(ctx, filter, options)
	if errDB != nil {
		return nil, 0, utils.Wrap(errDB, "No Record found")
	}

	err = momentResult.All(context.TODO(), &moments)
	if err != nil {
		return nil, 0, utils.Wrap(err, "No Record found")
	}

	return moments, count, err

}

func (d *DataBases) UpdateMomentsUserInfo(moment Moment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(moment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"user_id": bson.M{
			"$eq": moment.UserID,
		},
	}

	setM := primitive.M{
		"user_id":       moment.UserID,
		"m_update_time": moment.MUpdateTime,
	}

	if moment.UserName != "" {
		setM["user_name"] = moment.UserName
	}
	if moment.UserProfileImg != "" {
		setM["user_profile_img"] = moment.UserProfileImg
	}

	updateInter := bson.M{
		"$set": setM,
	}

	_, err = c.UpdateMany(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) UpdateMomentCommentUserInfo(comment MomentComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(comment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"user_id": bson.M{
			"$eq": comment.UserID,
		},
	}
	setM := primitive.M{
		"user_id":      comment.UserID,
		"updated_time": comment.UpdatedTime,
		"update_by":    comment.UpdateBy,
	}

	if comment.UserName != "" {
		setM["user_name"] = comment.UserName
	}
	if comment.UserProfileImg != "" {
		setM["user_profile_img"] = comment.UserProfileImg
	}

	updateInter := bson.M{
		"$set": setM,
	}

	_, err = c.UpdateMany(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) UpdateMomentLikeUserInfo(like MomentLike) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(like.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"user_id": bson.M{
			"$eq": like.UserID,
		},
	}

	setM := primitive.M{
		"user_id":      like.UserID,
		"updated_time": like.UpdatedTime,
		"update_by":    like.UpdateBy,
	}

	if like.UserName != "" {
		setM["user_name"] = like.UserName
	}
	if like.UserProfileImg != "" {
		setM["user_profile_img"] = like.UserProfileImg
	}

	updateInter := bson.M{
		"$set": setM,
	}

	_, err = c.UpdateMany(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) AddMomentLike(momentLike MomentLike) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(momentLike.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	_, err = c.InsertOne(ctx, momentLike)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) CancelMomentLike(momentLike MomentLike) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(momentLike.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": momentLike.MomentID,
		},
		"user_id": bson.M{
			"$eq": momentLike.CreateBy,
		},
	}
	// update := bson.M{
	// 	"$set": bson.M{
	// 		"deleted_by":  momentLike.CreateBy,
	// 		"delete_time": time.Now().Unix(),
	// 		"status":      false,
	// 	},
	// }
	deleteResult, err := c.DeleteOne(ctx, filter)
	if err != nil || deleteResult.DeletedCount <= 0 {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) UpdateMomentLike(momentLike MomentLike) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(momentLike.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": momentLike.MomentID,
		},
		"user_id": bson.M{
			"$eq": momentLike.UserID,
		},
	}

	_, err = c.DeleteOne(ctx, filter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	_, err = c.InsertOne(ctx, momentLike)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) GetMomentLikeByID(momentID primitive.ObjectID, userID string) (*MomentLike, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentLike{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": momentID,
		},
		"user_id": bson.M{
			"$eq": userID,
		},
	}
	var momentLike MomentLike
	momentResult := c.FindOne(ctx, filter)
	err = momentResult.Decode(&momentLike)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return &momentLike, err
}

func (d *DataBases) UpdateLikeStatus(momentLike MomentLike) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(momentLike.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": momentLike.MomentID,
		},
		"user_id": bson.M{
			"$eq": momentLike.UserID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"status":       momentLike.Status,
			"updated_time": time.Now().Unix(),
			"update_by":    momentLike.UpdateBy,
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) AddMomentComment(momentComment MomentComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(momentComment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	_, err = c.InsertOne(ctx, momentComment)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) GetMomentComment(commentID primitive.ObjectID) (*MomentComment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"comment_id": bson.M{
			"$eq": commentID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}
	var momentComment MomentComment
	momentResult := c.FindOne(ctx, filter)
	err = momentResult.Decode(&momentComment)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return &momentComment, err
}

func (d *DataBases) DeleteMomentComment(comment MomentComment, opUserId string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(comment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"comment_id": bson.M{
			"$eq": comment.CommentID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"delete_time": time.Now().Unix(),
			"deleted_by":  opUserId,
		},
	}
	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) UpdateMomentCommentV2(comment MomentComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	articleColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(comment.TableName())

	filter := bson.M{"comment_id": comment.CommentID, "delete_time": 0}
	update := bson.M{"$inc": bson.M{
		"comment_replies": comment.CommentReplies,
		"like_counts":     comment.LikeCounts,
	}}
	if _, err = articleColl.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) UpdateMomentComment(comment MomentComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(comment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"comment_id": bson.M{
			"$eq": comment.CommentID,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"comment_content": comment.CommentContent,
			"update_by":       comment.UpdateBy,
			"updated_time":    time.Now().Unix(),
			"status":          comment.Status,
		},
	}
	_, err = c.UpdateOne(ctx, filter, update)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) UpdateMomentCommentStatus(comment MomentComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(comment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"comment_id": bson.M{
			"$eq": comment.CommentID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"status":       comment.Status,
			"update_by":    comment.UpdateBy,
			"updated_time": time.Now().Unix(),
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) GetMomentsByFriendList(friendList []string, creator string, pageNumber, limit int64) ([]Moment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"creator_id": bson.M{
			"$in": friendList,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	blockFilter := bson.M{
		"$or": []bson.M{{
			"creator_id": bson.M{
				"$eq": creator,
			},
		}, {
			"$and": []bson.M{{
				"creator_id": bson.M{
					"$ne": creator,
				}}, {
				"privacy": bson.M{
					"$eq": 1,
				}}},
		},
		},
	}

	var moments []Moment

	cursor, err := c.Aggregate(context.TODO(), mongo.Pipeline{
		bson.D{{"$match", filter}},
		bson.D{{"$match", blockFilter}},
		bson.D{{"$sort", bson.D{{"m_create_time", -1}}}},
		bson.D{{"$skip", pageNumber * limit}},
		bson.D{{"$limit", limit}},
	})
	if err != nil {
		fmt.Println("err111", err.Error())
		return moments, err
	}

	if err = cursor.All(context.TODO(), &moments); err != nil {
		return moments, err
	}

	return moments, err
}

func (d *DataBases) GetMomentCommentsByMomentAndFriendIds(momentID primitive.ObjectID, friendList []string, pageNumber, commentsLimit int64) ([]MomentComment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	options := options.Find()
	options.SetSort(bson.D{{"comment_id", 1}})
	options.SetLimit(commentsLimit)
	options.SetSkip(pageNumber * commentsLimit)

	filter := bson.M{
		"user_id": bson.M{
			"$in": friendList,
		},
		"moment_id": bson.M{
			"$eq": momentID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
		"status": bson.M{
			"$eq": 1,
		},
	}
	var momentComments []MomentComment
	momentResult, errDB := c.Find(ctx, filter, options)
	if errDB != nil {
		return nil, utils.Wrap(errDB, "No Record found")
	}
	err = momentResult.All(context.TODO(), &momentComments)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return momentComments, err
}

func (d *DataBases) GetMomentLikesByMomentAndFriendIds(momentID primitive.ObjectID, friendList []string) ([]MomentLike, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentLike{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	options := options.Find()
	options.SetSort(bson.D{{"comment_id", 1}})

	filter := bson.M{
		"user_id": bson.M{
			"$in": friendList,
		},
		"moment_id": bson.M{
			"$eq": momentID,
		},
		"status": bson.M{
			"$eq": 1,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}
	var momentLikes []MomentLike
	momentResult, errDB := c.Find(ctx, filter, options)
	if errDB != nil {
		return nil, utils.Wrap(errDB, "No Record found")
	}
	err = momentResult.All(context.TODO(), &momentLikes)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return momentLikes, err
}

func (d *DataBases) GetMomentsByFriendListAndSearchKeyword(friendList []string, keyword string, pageNumber, limit int64) ([]Moment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	options := options.Find()
	options.SetSort(bson.D{{"m_create_time", -1}})
	options.SetSkip(pageNumber * limit)
	options.SetLimit(limit)

	OrFilter := []bson.M{}
	OrFilter = append(OrFilter, bson.M{"m_content_text": bson.M{"$regex": keyword}})
	OrFilter = append(OrFilter, bson.M{"creator_id": bson.M{"$regex": keyword}})

	filter := bson.M{
		"creator_id": bson.M{
			"$in": friendList,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
		"$or": OrFilter,
	}
	var moments []Moment
	momentResult, errDB := c.Find(ctx, filter, options)
	if errDB != nil {
		return nil, utils.Wrap(errDB, "No Record found")
	}
	err = momentResult.All(context.TODO(), &moments)
	// err = momentResult.Decode(&moments)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return moments, err
}

func (d *DataBases) GetMomentInteractedUsers(momentID primitive.ObjectID) ([]interface{}, error) {

	moment, _ := d.GetMoment(momentID)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": momentID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}
	var userCommented []interface{}
	if moment != nil {
		userCommented = append(userCommented, moment.CreatorID)
	}
	momentResult, err := c.Distinct(ctx, "user_id", filter)
	if err == nil {
		userCommented = append(userCommented, momentResult...)
	}

	c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentLike{}.TableName())

	filter = bson.M{
		"moment_id": bson.M{
			"$eq": momentID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}
	momentResult, err = c.Distinct(ctx, "user_id", filter)
	if err == nil {
		userCommented = append(userCommented, momentResult...)
	}
	userInteracted := unique(userCommented)
	return userInteracted, err
}

func unique(intSlice []interface{}) []interface{} {
	keys := make(map[interface{}]bool)
	list := make([]interface{}, 0)
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (d *DataBases) AddOfficialFollow(officialID int64, userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	// check if user is already have follow record
	var existingFollow OfficialFollow
	err = c.FindOne(ctx, bson.M{"official_id": officialID, "user_id": userID}).Decode(&existingFollow)
	// if not create
	if err == mongo.ErrNoDocuments {
		newFollow := OfficialFollow{
			OfficialID: officialID,
			UserID:     userID,
			FollowTime: time.Now().Unix(),
			Muted:      true,
			Enabled:    true,
		}
		if _, err = c.InsertOne(ctx, &newFollow); err != nil {
			return utils.Wrap(err, "transaction failed")
		}
	} else if err != nil {
		// other errors
		return utils.Wrap(err, "transaction failed")
	} else {
		// skip if already followed
		if existingFollow.DeleteTime == 0 {
			return nil
		}
		// update otherwise
		filter := bson.M{"official_id": officialID, "user_id": userID}
		update := bson.M{"$set": bson.M{"delete_time": 0, "delete_by": "", "follow_time": time.Now().Unix()}}
		if _, err = c.UpdateOne(ctx, filter, update); err != nil {
			return utils.Wrap(err, "transaction failed")
		}
	}
	return nil
}

func (d *DataBases) AddOfficialFollows(officialID int64, userIdList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	followTime := time.Now().Unix()
	// update otherwise
	filter := bson.M{"official_id": officialID}
	update := bson.M{"$set": bson.M{"delete_time": 0, "delete_by": "", "follow_time": followTime}}
	if _, err = c.UpdateMany(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	// create follows， if the count > 100,000, need to slice the record to insert many.
	batchSize := 1000 // 每批次插入的记录数
	total := len(userIdList)
	nowTime := time.Now().Unix()
	newFollows := make([]interface{}, 1000)
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		userIDBatch := userIdList[i:end]
		batchLen := len(userIDBatch)
		if batchLen < 1000 {
			newFollows = make([]interface{}, batchLen)
		}
		for index, userId := range userIDBatch {
			newFollows[index] = OfficialFollow{
				OfficialID: officialID,
				UserID:     userId,
				FollowTime: nowTime,
				Muted:      true,
				Enabled:    true,
			}
		}

		if _, err = c.InsertMany(ctx, newFollows); err != nil {
			return utils.Wrap(err, "transaction failed")
		}
	}
	return nil
}

func (d *DataBases) DeleteOfficialFollows(officialID int64, userID string, userIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	filter := bson.M{
		"official_id": officialID,
		"user_id": bson.M{
			"$in": userIDList,
		},
	}
	update := bson.M{
		"$set": bson.M{
			"delete_by":   userID,
			"delete_time": time.Now().Unix(),
		},
	}
	if _, err = c.UpdateMany(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) UpdateOfficialFollow(officialID int64, userID string, muted, enabled bool) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	_, err = c.UpdateOne(ctx, bson.M{"user_id": userID, "official_id": officialID, "delete_time": 0}, bson.M{"$set": bson.M{"muted": muted, "enabled": enabled}})
	return err
}

func (d *DataBases) AddOfficialFollowBlock(officialID int64, officialUserID string, UserIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	filter := bson.M{
		"official_id": officialID,
		"user_id":     bson.M{"$in": UserIDList},
	}
	update := bson.M{
		"$set": bson.M{
			"blocked_by": officialUserID,
			"block_time": time.Now().Unix(),
		},
	}
	if _, err = c.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) DeleteOfficialFollowBlock(officialID int64, UserIDList []string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	filter := bson.M{
		"official_id": officialID,
		"user_id":     bson.M{"$in": UserIDList},
	}
	update := bson.M{
		"$set": bson.M{
			"blocked_by": "",
			"block_time": 0,
		},
	}
	if _, err = c.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) AddRequestToOpertaionLog(req *pbAdminCMS.OperationLogRequest) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(operationLogs)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	sCtx := mongo.NewSessionContext(ctx, session)
	if err != nil {
		log.NewError("Request Binding faield for loging DB", err.Error())
		return utils.Wrap(err, "start transaction failed")
	}
	req.CreateTime = time.Now().Unix()
	_, err = c.InsertOne(sCtx, *req)
	if err != nil {
		log.NewError("Request Binding DB", err.Error())
	}

	session.CommitTransaction(ctx)
	return err
}

func (d *DataBases) SearchOpertaionLog(req *pbAdminCMS.SearchOperationLogsRequest) ([]OperationLog, int64, error) {
	var totalRecordsInDb int64 = 0
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(operationLogs)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, totalRecordsInDb, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	findOptions := options.Find()
	findOptions.SetLimit(req.PageLimit)
	findOptions.SetSkip((req.PageNumber - 1) * req.PageLimit)
	findOptions.SetSort(bson.D{{"createtime", -1}})

	Orfilter := []bson.M{}

	if req.Operator != "" {
		Orfilter = append(Orfilter, bson.M{"operator": primitive.Regex{Pattern: req.Operator, Options: ""}})
	}
	if req.Action != "" {
		Orfilter = append(Orfilter, bson.M{"action": primitive.Regex{Pattern: req.Action, Options: ""}})
	}
	if req.Executee != "" {
		Orfilter = append(Orfilter, bson.M{"executee": primitive.Regex{Pattern: req.Executee, Options: ""}})
	}
	if req.DateStart > 0 && req.DateEnd > 0 {
		Orfilter = append(Orfilter, bson.M{"createtime": bson.M{"$gte": req.DateStart, "$lte": req.DateEnd}})
		// Orfilter = append(Orfilter, bson.M{"$gte": bson.M{"createtime": req.DateStart}})
		// Orfilter = append(Orfilter, bson.M{"$lt": bson.M{"createtime": req.DateEnd}})

	}
	filter := bson.M{}
	if len(Orfilter) > 0 {
		filter = bson.M{"$and": Orfilter}
	}

	var operationLogs []OperationLog
	momentResult, errDB := c.Find(ctx, filter, findOptions)
	if errDB != nil {
		return nil, totalRecordsInDb, utils.Wrap(errDB, "No Record found")
	}
	err = momentResult.All(context.TODO(), &operationLogs)
	if err != nil {
		log.NewError("Operation Search # 1", err.Error())
		return nil, totalRecordsInDb, utils.Wrap(err, "No Record found")
	}
	totalRecordsInDb, errDB = c.CountDocuments(ctx, filter)
	if errDB != nil {
		log.NewError("Operation Search # 2", err.Error())
		return nil, totalRecordsInDb, utils.Wrap(errDB, "No Record found")
	}
	log.NewError("Operation Search ", operationLogs)
	return operationLogs, totalRecordsInDb, err

}

func (d *DataBases) AddArticleLike(articleID int64, userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	articleLikeColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleLike{}.TableName())

	var count int64
	filter := bson.M{"article_id": articleID, "user_id": userID, "delete_time": 0}
	if count, err = articleLikeColl.CountDocuments(ctx, filter); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if count != 0 {
		log.Debug("", "liked")
		return errors.New("you have liked")
	}

	log.Debug("", "articleID: ", articleID, "user_id", userID, "delete_time", 0)
	log.Debug("", "LikeArticle count: ", count)
	log.Debug("", "err: ", err)

	article := ArticleLike{
		ArticleID:  articleID,
		UserID:     userID,
		CreatedBy:  userID,
		CreateTime: time.Now().Unix(),
		Status:     1,
	}
	if _, err = articleLikeColl.InsertOne(ctx, article); err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return nil
}

func (d *DataBases) DeleteArticleLike(articleID int64, userID, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	articleLikeColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleLike{}.TableName())

	var count int64
	filter := bson.M{"article_id": articleID, "user_id": userID, "delete_time": 0}
	if count, err = articleLikeColl.CountDocuments(ctx, filter); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if count == 0 {
		return errors.New("you have unliked")
	}

	_, err = articleLikeColl.DeleteOne(ctx, filter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return nil
}

func (d *DataBases) DeleteArticleLikeByUserID(userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	articleLikeColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleLike{}.TableName())

	var count int64
	filter := bson.M{"user_id": userID, "delete_time": 0}
	if count, err = articleLikeColl.CountDocuments(ctx, filter); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if count == 0 {
		return errors.New("you have unliked")
	}

	_, err = articleLikeColl.DeleteOne(ctx, filter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return nil
}

func (d *DataBases) AddArticleCommentLike(commentID int64, userID string, officialID int64, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	commentLikeColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleCommentLike{}.TableName())

	filter := bson.M{"comment_id": commentID, "delete_time": 0}
	if userID != "" {
		filter["user_id"] = userID
	}
	if officialID != 0 {
		filter["official_id"] = officialID
	}
	var count int64
	if count, err = commentLikeColl.CountDocuments(ctx, filter); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if count != 0 {
		return errors.New("you have liked")
	}

	article := ArticleCommentLike{
		CommentID:  commentID,
		UserID:     userID,
		OfficialID: officialID,
		CreatedBy:  opUserID,
		CreateTime: time.Now().Unix(),
		Status:     1,
	}
	if _, err = commentLikeColl.InsertOne(ctx, article); err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return nil
}

func (d *DataBases) DeleteArticleCommentLike(commentID int64, userID string, officialID int64, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	commentLikeColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleCommentLike{}.TableName())

	filter := bson.M{"comment_id": commentID, "delete_time": 0}
	if userID != "" {
		filter["user_id"] = userID
	}
	if officialID != 0 {
		filter["official_id"] = officialID
	}
	var count int64
	if count, err = commentLikeColl.CountDocuments(ctx, filter); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if count == 0 {
		return errors.New("you have unliked")
	}

	update := bson.M{"$set": bson.M{"delete_time": time.Now().Unix(), "deleted_by": opUserID}}
	updateRes, err := commentLikeColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	if updateRes.ModifiedCount == 0 {
		return nil
	}

	return nil
}

func (d *DataBases) DeleteArticleCommentLikeByUserID(userID, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	commentLikeColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleCommentLike{}.TableName())

	filter := bson.M{"user_id": userID, "delete_time": 0}
	var count int64
	if count, err = commentLikeColl.CountDocuments(ctx, filter); err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if count == 0 {
		return errors.New("you have unliked")
	}

	update := bson.M{"$set": bson.M{"delete_time": time.Now().Unix(), "deleted_by": opUserID}}
	updateRes, err := commentLikeColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	if updateRes.ModifiedCount == 0 {
		return nil
	}

	return nil
}

func (d *DataBases) OfficialDeleteArticleComment(commentID int64, userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	commentColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())

	var comment ArticleComment
	err = commentColl.FindOne(ctx, bson.M{"comment_id": commentID, "delete_time": 0}).Decode(&comment)
	if err != nil {
		return err
	}

	filter := bson.M{"$or": []bson.M{{"comment_id": commentID}, {"parent_comment_id": commentID}}, "delete_time": 0}
	update := bson.M{"$set": bson.M{"delete_time": time.Now().Unix(), "deleted_by": userID}}
	updateRes, err := commentColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	if updateRes.ModifiedCount == 0 {
		return nil
	}

	if comment.ParentCommentID != 0 {
		filter = bson.M{"comment_id": comment.ParentCommentID}
		update = bson.M{"$inc": bson.M{"reply_counts": -updateRes.ModifiedCount}}
		if _, err = commentColl.UpdateOne(ctx, filter, update); err != nil {
			return utils.Wrap(err, "transaction failed")
		}
	}

	return nil
}

func (d *DataBases) OfficialHideArticleComment(commentID int64, userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())

	filter := bson.M{"comment_id": commentID, "delete_time": 0}
	update := bson.M{"$set": bson.M{"status": 2, "update_time": time.Now().Unix(), "updated_by": userID}}
	if _, err = c.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) OfficialShowArticleComment(commentID int64, userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())

	filter := bson.M{"comment_id": commentID, "delete_time": 0}
	update := bson.M{"$set": bson.M{"status": 1, "update_time": time.Now().Unix(), "updated_by": userID}}
	if _, err = c.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) UpdateArticleLike(like ArticleLike) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(like.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"article_id": bson.M{
			"$eq": like.ArticleID,
		},
		"user_id": bson.M{
			"$eq": like.UserID,
		},
	}

	_, err = c.DeleteOne(ctx, filter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	_, err = c.InsertOne(ctx, like)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) AddArticleComment(commentID, articleID, officialID int64, userID string, parentCommentID int64, replyOfficialID int64, replyUserID, content, opUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	commentColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	comment := ArticleComment{
		CommentID:       commentID,
		ParentCommentID: parentCommentID,
		ArticleID:       articleID,
		OfficialID:      officialID,
		UserID:          userID,
		ReplyUserID:     replyUserID,
		ReplyOfficialID: replyOfficialID,
		Content:         content,
		CreatedBy:       opUserID,
		CreateTime:      time.Now().Unix(),
		Status:          1,
	}

	if _, err = commentColl.InsertOne(ctx, comment); err != nil {
		return err
	}
	if parentCommentID != 0 {
		_, err = commentColl.UpdateOne(ctx, bson.M{"comment_id": parentCommentID}, bson.M{"$inc": bson.M{"reply_counts": 1}})
		if err != nil {
			return err
		}
	}
	return err
}

func (d *DataBases) DeleteArticleComment(comment ArticleComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(comment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"comment_id": bson.M{
			"$eq": comment.CommentID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"delete_time": time.Now().Unix(),
		},
	}
	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) GetSelfOfficialAccountFollows(userID string) ([]OfficialFollow, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var officialFollows []OfficialFollow
	filter := bson.M{
		"user_id":     userID,
		"delete_time": 0,
		"block_time":  0,
	}
	result, errDB := c.Find(ctx, filter)
	if errDB != nil {
		return nil, utils.Wrap(errDB, "No Record found")
	}
	err = result.All(ctx, &officialFollows)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return officialFollows, nil
}

func (d *DataBases) GetArticlesByID(articleID int64) (*Article, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Article{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	filter := bson.M{"article_id": articleID}
	var article Article
	errDB := c.FindOne(ctx, filter).Decode(&article)
	if errDB != nil {
		log.NewError("Get Article in Moment", "Article DB Not Found ", errDB.Error())
		return nil, utils.Wrap(errDB, "No Record found")
	}
	return &article, err
}

func (d *DataBases) GetArticlesByFollowedOfficialIDList(systemOfficialList, officialIDList, blockedOfficialIDList []int64, offset, limit int64) ([]Article, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Article{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var needCount = limit
	var articles []Article
	var count int64 = 0
	// find system articles

	log.Debug("", "systemOfficialList: ", systemOfficialList)

	// V2: get system official articles
	/*
		if systemOfficialList != nil {
			filterOneOptions := options.FindOneOptions{}
			filterOneOptions.SetSort(bson.D{{"create_time", 1}})
			for index, systemOfficial := range systemOfficialList {
				if index == int(limit) {
					break
				}
				article := Article{}
				if err = c.FindOne(ctx, bson.M{
					"official_id": bson.M{
						"$eq": systemOfficial,
						"$nin": blockedOfficialIDList,
					},
					"delete_time": 0,
					"privacy": 1}, &filterOneOptions).Decode(&article); err != nil {
					log.NewError("", "not find ", err.Error())
				} else if article.ArticleID != 0{
					count++
					needCount--
					articles = append(articles, article)
				}
			}
		}
	*/
	log.Debug("", "articles: ", articles)

	// if system articles not enough, then get user's articles.
	//filterOptions := options.AggregateOptions{}
	//filterOptions.spr(bson.D{{"create_time", -1}})
	//filterOptions.SetSkip(offset)
	//filterOptions.SetLimit(needCount)

	// bson.D{{"created_time", 1}}
	andCondition := []bson.M{
		{"privacy": bson.M{"$eq": 1}},
		{"delete_time": bson.M{"$eq": 0}},
	}
	/*
		var filters []bson.D
			if minCreateTime != 0 {
				filters = append(filters, bson.D{{"create_time", bson.D{{"$gte", minCreateTime}}}})
			}
			if maxCreateTime != 0 {
				filters = append(filters, bson.D{{"create_time", bson.D{{"$lte", maxCreateTime}}}})
			}
			if keyword != "" {
				filters = append(filters, bson.D{{"$or", []bson.D{
					{{"title", bson.M{"$regex": keyword, "$options": "i"}}},
					{{"text_content", bson.M{"$regex": keyword, "$options": "i"}}},
				}}})
			}

	*/
	matchMap := bson.M{"delete_time": bson.M{"$eq": 0}, "privacy": bson.M{"$eq": 1}}
	//userGroupStage := []bson.M{
	//}
	if blockedOfficialIDList != nil && len(blockedOfficialIDList) > 0 {
		allNotIn := []int64{}
		allNotIn = append(allNotIn, blockedOfficialIDList...)
		allNotIn = append(allNotIn, systemOfficialList...)
		andCondition = append(andCondition, bson.M{"official_id": bson.M{"$nin": blockedOfficialIDList}})
		matchMap["official_id"] = bson.M{"$nin": allNotIn}
	} else {
		matchMap["official_id"] = bson.M{"$nin": systemOfficialList}
	}
	log.Debug("", "officialIDList is null: ", officialIDList == nil)
	if officialIDList != nil {
		andCondition = append(andCondition, bson.M{"official_id": bson.M{"$in": officialIDList}})
	} else {
		// if nil, then is recommend, group searching
		var usersArticle []Article

		//userGroupStage = append(userGroupStage,
		//	bson.M{"$group": bson.D{{"_id", "$official_id"}}},
		//	bson.M{"$match": matchMap},
		//	bson.M{"$sort": bson.D{{"create_time", -1}}},
		//	bson.M{"$limit": needCount},
		//)
		// cursor, err := c.Aggregate(context.TODO(), userGroupStage)
		cursor, err := c.Aggregate(context.TODO(), mongo.Pipeline{
			bson.D{{"$match", matchMap}},
			bson.D{{"$group", bson.D{{"_id", "$official_id"}, {"firstDoc", bson.D{{"$first", "$$ROOT"}}}}}},
			bson.D{{"$sort", bson.D{{"create_time", 1}}}},
			{{"$replaceRoot", bson.D{
				{"newRoot", "$firstDoc"},
			}}},
			bson.D{{"$limit", needCount}},
		})
		if err != nil {
			fmt.Println("err111", err.Error())
			return nil, 0, nil
		}

		fmt.Println(len(usersArticle))
		if err = cursor.All(context.TODO(), &usersArticle); err != nil {
			return nil, 0, nil
		}
		count += int64(len(usersArticle))

		log.Debug("", "usersArticle: ", usersArticle)
		articles = append(articles, usersArticle...)
		log.Debug("", "articles: ", articles)
		return articles, count, err
	}

	filter := bson.M{"$and": andCondition}

	filterOptions := options.Find()
	filterOptions.SetSort(bson.D{{"create_time", -1}})
	filterOptions.SetSkip(offset)
	filterOptions.SetLimit(needCount)

	result, errDB := c.Find(ctx, filter, filterOptions)
	if errDB != nil {
		return nil, 0, utils.Wrap(errDB, "No Record found")
	}

	count, err = c.CountDocuments(ctx, filter)

	if err = result.All(ctx, &articles); err != nil {
		return nil, 0, utils.Wrap(err, "No Record found")
	}

	return articles, count, err
}

func (d *DataBases) GetArticlesByFollowedOfficialIDListV2(xlinkOfficialID int64, officialIDList, blockedOfficialIDList []int64, offset, limit int64, deletedIdList []int64) ([]Article, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Article{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var needCount = limit
	var articles []Article
	var count int64 = 0

	for _, blockOfficial := range blockedOfficialIDList {
		if blockOfficial == xlinkOfficialID {
			xlinkOfficialID = 0
			break
		}
	}

	// find system articles
	var xlinkArticle *Article = nil
	if xlinkOfficialID != 0 {
		filterOneOptions := options.FindOneOptions{}
		filterOneOptions.SetSort(bson.D{{"create_time", 1}})
		xlinkArticle = &Article{}
		if err = c.FindOne(ctx, bson.M{
			"official_id": bson.M{
				"$eq": xlinkOfficialID,
			},
			"delete_time": 0,
			"privacy":     1}, &filterOneOptions).Decode(xlinkArticle); err != nil {
			log.NewError("", "not find ", err.Error())
		} else {
			t := time.Now()
			todayTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()
			if xlinkArticle.CreateTime < todayTime {
				xlinkArticle = nil
			} else {
				needCount--
			}
		}
	}

	// if system articles not enough, then get user's articles.
	//filterOptions := options.AggregateOptions{}
	//filterOptions.spr(bson.D{{"create_time", -1}})
	//filterOptions.SetSkip(offset)
	//filterOptions.SetLimit(needCount)

	// bson.D{{"created_time", 1}}
	andCondition := []bson.M{
		{"privacy": bson.M{"$eq": 1}},
		{"delete_time": bson.M{"$eq": 0}},
	}
	if len(deletedIdList) != 0 {
		andCondition = append(andCondition, bson.M{"official_id": bson.M{"$nin": deletedIdList}})
	}
	/*
		var filters []bson.D
			if minCreateTime != 0 {
				filters = append(filters, bson.D{{"create_time", bson.D{{"$gte", minCreateTime}}}})
			}
			if maxCreateTime != 0 {
				filters = append(filters, bson.D{{"create_time", bson.D{{"$lte", maxCreateTime}}}})
			}
			if keyword != "" {
				filters = append(filters, bson.D{{"$or", []bson.D{
					{{"title", bson.M{"$regex": keyword, "$options": "i"}}},
					{{"text_content", bson.M{"$regex": keyword, "$options": "i"}}},
				}}})
			}

	*/
	matchMap := bson.M{"delete_time": bson.M{"$eq": 0}, "privacy": bson.M{"$eq": 1}}
	//userGroupStage := []bson.M{
	//}
	if blockedOfficialIDList != nil && len(blockedOfficialIDList) > 0 {
		allNotIn := []int64{}
		allNotIn = append(allNotIn, blockedOfficialIDList...)
		if xlinkOfficialID != 0 {
			allNotIn = append(allNotIn, xlinkOfficialID)
		}
		andCondition = append(andCondition, bson.M{"official_id": bson.M{"$nin": blockedOfficialIDList}})
		matchMap["official_id"] = bson.M{"$nin": allNotIn}
	} else {
		if xlinkOfficialID != 0 {
			matchMap["official_id"] = bson.M{"$ne": xlinkOfficialID}
		}
	}
	if officialIDList != nil {
		andCondition = append(andCondition, bson.M{"official_id": bson.M{"$in": officialIDList}})
	} else {
		// if nil, then is recommend, group searching
		var usersArticle []Article

		cursor, err := c.Aggregate(context.TODO(), mongo.Pipeline{
			bson.D{{"$sort", bson.D{{"create_time", -1}}}},
			bson.D{{"$match", matchMap}},
			bson.D{{"$group", bson.D{{"_id", "$official_id"}, {"firstDoc", bson.D{{"$first", "$$ROOT"}}}}}},
			{{"$replaceRoot", bson.D{
				{"newRoot", "$firstDoc"},
			}}},
			bson.D{{"$sort", bson.D{{"create_time", -1}}}},
			bson.D{{"$limit", needCount}},
		})
		if err != nil {
			return nil, 0, nil
		}

		if err = cursor.All(context.TODO(), &usersArticle); err != nil {
			return nil, 0, nil
		}
		count += int64(len(usersArticle))

		log.Debug("", "usersArticle: ", usersArticle)

		articles = make([]Article, count)
		added := false

		var index int64 = 0
		var articleIndex = 0
		for ; index < count; index++ {
			if !added && xlinkArticle != nil &&
				xlinkArticle.CreateTime >= usersArticle[articleIndex].CreateTime {
				articles[index] = *xlinkArticle
				index++
				fmt.Println("loop add link article, index: ", index)
				added = true
			}
			articles[index] = usersArticle[articleIndex]
			articleIndex++
		}
		fmt.Println(added)
		if xlinkArticle != nil && !added {
			articles[count-1] = *xlinkArticle
		}
		return articles, count, err
	}

	filter := bson.M{"$and": andCondition}

	filterOptions := options.Find()
	filterOptions.SetSort(bson.D{{"create_time", -1}})
	filterOptions.SetSkip(offset)
	filterOptions.SetLimit(needCount)

	result, errDB := c.Find(ctx, filter, filterOptions)
	if errDB != nil {
		return nil, 0, utils.Wrap(errDB, "No Record found")
	}

	count, err = c.CountDocuments(ctx, filter)

	if err = result.All(ctx, &articles); err != nil {
		return nil, 0, utils.Wrap(err, "No Record found")
	}

	return articles, count, err
}

func (d *DataBases) GetArticleCommentsByArticleID(articleID int64, pageNumber, commentsLimit int64) ([]ArticleComment, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	options := options.Find()
	options.SetSort(bson.D{{"created_time", 1}})
	options.SetLimit(commentsLimit)
	options.SetSkip(pageNumber * commentsLimit)

	filter := bson.M{
		"article_id": bson.M{
			"$eq": articleID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}
	var articleComments []ArticleComment
	momentResult, errDB := c.Find(ctx, filter, options)
	if errDB != nil {
		return nil, utils.Wrap(errDB, "No Record found")
	}
	err = momentResult.All(context.TODO(), &articleComments)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return articleComments, err
}

func (d *DataBases) GetArticleLikesByArticleID(articleID int64) ([]ArticleLike, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleLike{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	options := options.Find()
	options.SetSort(bson.D{{"create_time", 1}})

	filter := bson.M{
		"delete_time": bson.M{
			"$eq": 0,
		},
		"article_id": bson.M{
			"$eq": articleID,
		},
	}
	var articleLikes []ArticleLike
	momentResult, errDB := c.Find(ctx, filter, options)
	if errDB != nil {
		return nil, utils.Wrap(errDB, "No Record found")
	}
	err = momentResult.All(context.TODO(), &articleLikes)
	if err != nil {
		return nil, utils.Wrap(err, "No Record found")
	}
	return articleLikes, err
}

func (d *DataBases) DeleteMoment(moment Moment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(moment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": moment.MomentID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"delete_time": time.Now().Unix(),
			"deleted_by":  moment.CreatorID,
		},
	}
	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) DeleteMomentsByUserID(userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"creator_id": bson.M{
			"$eq": userID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"delete_time": time.Now().Unix(),
			"deleted_by":  userID,
		},
	}
	_, err = c.UpdateMany(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) DeleteMomentLikesByUserID(userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentLike{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"user_id": bson.M{
			"$eq": userID,
		},
	}
	deleteResult, err := c.DeleteMany(ctx, filter)
	if err != nil || deleteResult.DeletedCount <= 0 {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) DeleteMomentCommentsByUserID(userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"user_id": bson.M{
			"$eq": userID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"user_name":        "Deleted Account",
			"user_profile_img": "",
			"account_status":   2,
		},
	}
	_, err = c.UpdateMany(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return err
}

func (d *DataBases) DecrementLikeCountInMomentByUserID(userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(MomentLike{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"user_id": bson.M{
			"$eq": userID,
		},
	}
	var momentsLike []MomentLike
	momentResult, errDB := c.Find(ctx, filter)
	if errDB == nil {
		err = momentResult.All(context.TODO(), &momentsLike)
		if err != nil {
			return err
		}
	}

	var momentsID []primitive.ObjectID
	for _, like := range momentsLike {
		momentsID = append(momentsID, like.MomentID)
	}

	filter = bson.M{"moment_id": bson.M{"$in": momentsID}}
	update := bson.M{"$inc": bson.M{"m_likes_count": -1}}

	c = d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	_, err = c.UpdateMany(ctx, filter, update)
	return nil
}

func (d *DataBases) UpdateMomentV2(moment Moment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	articleColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(moment.TableName())

	filter := bson.M{"moment_id": moment.MomentID, "delete_time": 0}
	update := bson.M{"$inc": bson.M{
		"m_likes_count":    moment.MLikesCount,
		"m_comments_count": moment.MCommentsCount,
		"m_repost_count":   moment.MRepostCount,
	}}
	if _, err = articleColl.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) UpdateMoment(moment Moment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(moment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": moment.MomentID,
		},
	}
	update := bson.M{"$set": bson.M{
		"m_content_text":           moment.MContentText,
		"m_content_images_array":   moment.MContentImagesArray,
		"m_content_videos_array":   moment.MContentVideosArray,
		"m_content_thumbnil_array": moment.MContentThumbnilArray,
		"is_reposted":              moment.IsReposted,
		"privacy":                  moment.Privacy,
		"m_update_time":            time.Now().Unix(),
	}}

	_, err = c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return err
}

// UpdateMomentFields Update status, privacy...
func (d *DataBases) UpdateMomentFields(moment Moment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"moment_id": bson.M{
			"$eq": moment.MomentID,
		},
	}

	updateMap := make(map[string]interface{})
	updateMap["m_update_time"] = moment.MUpdateTime

	if moment.Status != 0 {
		updateMap["status"] = moment.Status
	}
	if moment.Privacy != 0 {
		updateMap["privacy"] = moment.Privacy
	}
	if moment.CommentCtl != 0 {
		updateMap["comment_ctl"] = moment.CommentCtl
	}

	updateInter := bson.M{
		"$set": updateMap,
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return err
}

func (d *DataBases) GetMomentsByUserId(userId string, isFriend bool, pageNumber int64, showNumber int64) ([]Moment, error) {

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	opts := options.Find()
	opts.SetSkip((pageNumber - 1) * showNumber)
	opts.SetLimit(showNumber)
	opts.SetSort(bson.M{"m_create_time": -1})

	filter := map[string]bson.M{
		"creator_id": {
			"$eq": userId,
		},
		"delete_time": {
			"$eq": 0,
		},
	}
	if !isFriend {
		filter["privacy"] = bson.M{
			"$eq": 2,
		}
	}

	var moment []Moment
	momentResult, err := c.Find(ctx, filter, opts)
	if err != nil {
		log.Error("", "find moment error", err.Error())
		return nil, utils.Wrap(err, "No Record found")
	}

	if err = momentResult.All(context.TODO(), &moment); err != nil {
		log.Error("", "all moment error", err.Error())
		return nil, utils.Wrap(err, "No Record found")
	}

	return moment, err

}

func (d *DataBases) GetUserMomentCount(userId string, isFriend bool) (int64, error) {
	var count int64

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := map[string]bson.M{
		"creator_id": {
			"$eq": userId,
		},
		"delete_time": {
			"$eq": 0,
		},
	}
	if !isFriend {
		filter["privacy"] = bson.M{
			"$eq": 2,
		}
	}

	count, err = c.CountDocuments(ctx, filter)
	if err != nil {
		log.Error("", "find moment error", err.Error())
		return 0, utils.Wrap(err, "Find error")
	}

	return count, nil
}

func (d *DataBases) GetAllMomentLikeCounts(userId string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	collection := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Moment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	//, "delete_time": bson.M{"$eq": 0}
	matchMap := bson.M{"user_id": bson.M{"$eq": userId}}
	filter := mongo.Pipeline{
		bson.D{{"$match", matchMap}},
		bson.D{
			{"$group", bson.D{
				{"_id", nil},
				{"total", bson.D{
					{"$sum", "$m_likes_count"},
				}},
			}},
		},
	}

	cursor, err := collection.Aggregate(context.Background(), filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	var result struct {
		Total int64 `bson:"total"`
	}
	if cursor.Next(context.Background()) {
		fmt.Println("has next")
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}

	return result.Total, nil
}

func (d *DataBases) AddArticle(article Article) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(article.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	_, err = c.InsertOne(ctx, article)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

// UpdateArticleV2 update article for repost, like, comment, view counts
func (d *DataBases) UpdateArticleV2(article Article) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	articleColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(article.TableName())

	filter := bson.M{"article_id": article.ArticleID, "delete_time": 0}
	update := bson.M{"$inc": bson.M{
		"read_counts":        article.ReadCounts,
		"unique_read_counts": article.UniqueReadCounts,
		"comment_counts":     article.CommentCounts,
		"like_counts":        article.LikeCounts,
		"repost_counts":      article.RepostCounts,
	}}
	if _, err = articleColl.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) UpdateArticle(article Article) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(article.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"article_id": bson.M{
			"$eq": article.ArticleID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"cover_photo":  article.CoverPhoto,
			"title":        article.Title,
			"content":      article.Content,
			"text_content": article.TextContent,
			"updated_by":   article.UpdatedBy,
			"update_time":  article.UpdateTime,
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) ChangeArticlePrivacy(article Article) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(article.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"article_id": bson.M{
			"$eq": article.ArticleID,
		},
		"official_id": bson.M{
			"$eq": article.OfficialID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"privacy": article.Privacy,
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) DeleteArticle(article Article) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(article.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"article_id": bson.M{
			"$eq": article.ArticleID,
		},
		"official_id": bson.M{
			"$eq": article.OfficialID,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"deleted_by":  article.DeletedBy,
			"delete_time": article.DeleteTime,
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) DeleteAllArticlesByOfficialId(officialID int64, OpUserID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Article{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"official_id": bson.M{
			"$eq": officialID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"deleted_by":  OpUserID,
			"delete_time": time.Now().Unix(),
		},
	}

	_, err = c.UpdateMany(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) AddFavorite(favorite Favorites) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(favorite.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	if favorite.CreateBy == "" {
		favorite.CreateBy = favorite.UserID
	}
	favorite.CreateTime = time.Now().Unix()
	_, err = c.InsertOne(ctx, favorite)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) GetFavoritesById(favoriteId string) (Favorites, error) {
	favorite := Favorites{}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Favorites{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return favorite, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	favID, _ := primitive.ObjectIDFromHex(favoriteId)
	if err = c.FindOne(ctx, bson.M{"favorite_id": bson.M{"$eq": favID}}).Decode(&favorite); err != nil {
		log.NewError("", "not find favorite ", err.Error())
		return favorite, err
	}

	return favorite, nil
}

func (d *DataBases) GetFavorites(userId string, favoriteType int32) ([]Favorites, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Favorites{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	opts := options.Find()
	opts.SetSort(bson.M{"create_time": -1})

	t := time.Now().AddDate(0, 0, -6)
	sevenDaysAgo := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()

	filter := bson.M{
		"user_id": bson.M{
			"$eq": userId,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}
	if favoriteType == 0 {
		// all favorites, no type
		filter["create_time"] = bson.M{
			"$gt": sevenDaysAgo,
		}
	} else {
		// all favorites but has type
		filter["content_type"] = bson.M{
			"$eq": favoriteType,
		}
	}

	var favorites []Favorites

	favoriteResult, err := c.Find(ctx, filter, opts)
	if err != nil {
		log.Error("", "find favorite error", err.Error())
		return nil, utils.Wrap(err, "No Record found")
	}

	if err = favoriteResult.All(context.TODO(), &favorites); err != nil {
		log.Error("", "all favorite error", err.Error())
		return nil, utils.Wrap(err, "No Record found")
	}

	return favorites, nil

}

func (d *DataBases) RemoveFavorites(favorite Favorites) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(favorite.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"favorite_id": bson.M{
			"$eq": favorite.FavoriteId,
		},
		"delete_time": bson.M{
			"$eq": 0,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"deleted_by":  favorite.DeletedBy,
			"delete_time": time.Now().Unix(),
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}

	return nil
}

type FavoriteMedia struct {
	ObjName    string `bson:"obj_name"`
	FavorCount int64  `bson:"favor_count"`
}

func (d *DataBases) AddFavoriteMedia(obj string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	collection := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection("favorite_media")
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	favorMedia := FavoriteMedia{}
	favorMedia.ObjName = obj
	if err = collection.FindOne(ctx, bson.M{"obj_name": obj}).Decode(&favorMedia); err != nil {
		// no document, need to create a new one.
		favorMedia.FavorCount = 1
		_, err = collection.InsertOne(ctx, favorMedia)
		if err != nil {
			return 0, utils.Wrap(err, "transaction failed")
		}
	} else {
		// the media has been favored by other users, update the count +1
		favorMedia.FavorCount += 1
		_, err = collection.UpdateOne(ctx, bson.M{"obj_name": obj}, bson.M{"$inc": bson.M{"favor_count": 1}})
	}

	if err != nil {
		return 0, utils.Wrap(err, "transaction failed")
	}
	return favorMedia.FavorCount, nil
}

func (d *DataBases) GetFavoriteMediaCount(obj string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	collection := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection("favorite_media")
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	favorMedia := FavoriteMedia{}
	if err = collection.FindOne(ctx, bson.M{"obj_name": obj}).Decode(&favorMedia); err != nil && err != mongo.ErrNoDocuments {
		// no document
		return 0, utils.Wrap(err, "transaction failed")
	}

	return favorMedia.FavorCount, nil
}

func (d *DataBases) RemoveFavoriteMedia(obj string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	collection := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection("favorite_media")
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	favorMedia := FavoriteMedia{}
	favorMedia.ObjName = obj
	if err = collection.FindOne(ctx, bson.M{"obj_name": obj}).Decode(&favorMedia); err != nil {
		// no document, need to create a new one.
		return 0, err
	} else {
		// the media has been favored by other users, update the count +1
		favorMedia.FavorCount -= 1
		_, err = collection.UpdateOne(ctx, bson.M{"obj_name": obj}, bson.M{"$inc": bson.M{"favor_count": -1}})
		if favorMedia.FavorCount == 0 {
			_, err = collection.DeleteOne(ctx, bson.M{"obj_name": obj})
			if err != nil {
				return 0, err
			}
		}
	}

	if err != nil {
		return 0, utils.Wrap(err, "transaction failed")
	}

	return favorMedia.FavorCount, nil
}

func (d *DataBases) GetUsedCapacity(userID string) (int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	collection := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection("favorites")
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	matchMap := bson.M{"user_id": bson.M{"$eq": userID}, "delete_time": bson.M{"$eq": 0}}
	filter := mongo.Pipeline{
		bson.D{{"$match", matchMap}},
		bson.D{
			{"$group", bson.D{
				{"_id", nil},
				{"total", bson.D{
					{"$sum", "$file_size"},
				}},
			}},
		},
	}

	countResult, err := collection.Aggregate(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer countResult.Close(context.Background())

	var countStruct struct {
		Count int64 `bson:"total"`
	}

	if countResult.Next(ctx) {
		if err = countResult.Decode(&countStruct); err != nil {
			return 0, nil
		}
	}
	fmt.Println("result: ", countStruct.Count)

	return countStruct.Count, nil
}

type SearchArticle struct {
	Article  Article
	ReadTime int64 `bson:"read_time"`
}

func (d *DataBases) SearchArticles(keyword string, officialID, minReadTime, maxReadTime, minCreateTime, maxCreateTime, sort, offset, limit int64) ([]SearchArticle, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Article{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	matchMap := bson.M{"delete_time": bson.M{"$eq": 0}}
	if officialID != 0 {
		matchMap["official_id"] = bson.M{"$eq": officialID}
	}

	sortBy := bson.D{{"create_time", -1}}
	if maxReadTime != 0 || minReadTime != 0 {
		sortBy = bson.D{{"read_time", -1}}
	}
	switch sort {
	case 1:
		sortBy = bson.D{{"unique_read_counts", -1}}
		break
	case 2:
		sortBy = bson.D{{"like_counts", -1}}
		break
	case 3:
		sortBy = bson.D{{"comment_counts", -1}}
		break
	}

	var filters []bson.D
	if minCreateTime != 0 {
		filters = append(filters, bson.D{{"create_time", bson.D{{"$gte", minCreateTime}}}})
	}
	if maxCreateTime != 0 {
		filters = append(filters, bson.D{{"create_time", bson.D{{"$lte", maxCreateTime}}}})
	}
	if keyword != "" {
		filters = append(filters, bson.D{{"$or", []bson.D{
			{{"title", bson.M{"$regex": keyword, "$options": "i"}}},
			{{"text_content", bson.M{"$regex": keyword, "$options": "i"}}},
		}}})
	}

	readTimeExpr := []bson.M{
		bson.M{"$eq": []interface{}{"$article_id", "$$article_id"}},
		bson.M{"$eq": []interface{}{"$status", 1}},
	}
	if minReadTime != 0 {
		readTimeExpr = append(readTimeExpr, bson.M{"$gte": []interface{}{"$create_time", minReadTime}})
	}
	if maxReadTime != 0 {
		readTimeExpr = append(readTimeExpr, bson.M{"$lte": []interface{}{"$create_time", maxReadTime}})
	}

	pipeline := mongo.Pipeline{
		bson.D{{"$match", matchMap}},
		bson.D{{"$lookup", bson.M{
			"from": "article_read",
			"as":   "article_read",
			"let":  bson.D{{"article_id", "$article_id"}},
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"$expr": bson.M{"$and": readTimeExpr}}}},
			},
		}}},
		bson.D{{"$addFields", bson.D{{"read_time", bson.D{{"$arrayElemAt", []interface{}{"$article_read.create_time", -1}}}}}}},
		bson.D{{"$sort", sortBy}},
	}
	if minReadTime != 0 || maxReadTime != 0 {
		pipeline = append(pipeline, bson.D{{"$match", bson.M{"article_read.0": bson.M{"$exists": true}}}})
	}
	if len(filters) > 0 {
		pipeline = append(pipeline, bson.D{{"$match", bson.D{{"$and", filters}}}})
	}

	articlesPipeline := append(pipeline, bson.D{{"$skip", offset}}, bson.D{{"$limit", limit}})

	articlesResult, err := c.Aggregate(ctx, articlesPipeline)
	if err != nil {
		return nil, 0, utils.Wrap(err, "Query articles failed")
	}

	var searchArticles []SearchArticle
	for articlesResult.Next(ctx) {
		var searchArticle SearchArticle
		if err = articlesResult.Decode(&searchArticle); err != nil {
			return nil, 0, utils.Wrap(err, "Cast article read time failed")
		}
		if err = articlesResult.Decode(&searchArticle.Article); err != nil {
			return nil, 0, utils.Wrap(err, "Cast article failed")
		}
		searchArticles = append(searchArticles, searchArticle)
	}
	if err != nil {
		return nil, 0, utils.Wrap(err, "Cast articles failed")
	}

	countPipeline := append(pipeline, bson.D{{"$count", "count"}})
	countResult, err := c.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, utils.Wrap(err, "Count articles failed")
	}

	var countStruct struct {
		Count int64 `bson:"count"`
	}
	if countResult.Next(ctx) {
		if err = countResult.Decode(&countStruct); err != nil {
			return nil, 0, utils.Wrap(err, "Count articles cast failed")
		}
	}

	return searchArticles, countStruct.Count, nil
}

func (d *DataBases) GetOfficialFollow(officialID int64, userID string) (*OfficialFollow, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var follow OfficialFollow
	if err = c.FindOne(ctx, bson.M{"official_id": officialID, "user_id": userID, "delete_time": 0}).Decode(&follow); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, utils.Wrap(err, "Record not found")
	}

	return &follow, nil
}

func (d *DataBases) GetAllOfficialFollowers(officialId int64) ([]OfficialFollow, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var follow []OfficialFollow
	cursor, err := c.Find(ctx,
		bson.M{
			"official_id": officialId,
			"delete_time": 0})
	if err != nil {
		return follow, err
	}
	if err = cursor.All(ctx, &follow); err != nil {
		return follow, utils.Wrap(err, "Record not found")
	}

	return follow, nil
}

func (d *DataBases) GetAllOfficialFollowersData(officialId int64) ([]string, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	c := d.mongoClient.Database("bytechat").Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	var idList []string
	projection := bson.M{"user_id": 1, "_id": 0}
	selOption := options.Find().SetProjection(projection)
	var results []bson.M
	cursor, err := c.Find(ctx,
		bson.M{"official_id": officialId}, selOption)
	if err != nil {
		return idList, err
	}
	if err = cursor.All(ctx, &results); err != nil {
		return idList, nil
	}

	idList = make([]string, len(results))
	for index, result := range results {
		idList[index] = result["user_id"].(string)
	}

	return idList, nil
}

func (d *DataBases) ClearAllFollowers(officialId int64, opUser string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)
	filter := bson.M{
		"official_id": officialId,
	}
	update := bson.M{
		"$set": bson.M{
			"delete_by":   opUser,
			"delete_time": time.Now().Unix(),
		},
	}
	if _, err = c.UpdateMany(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) GetBlockedOfficialFollowList(userID string) ([]OfficialFollow, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(OfficialFollow{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var follow []OfficialFollow

	cursor, err := c.Find(ctx, bson.M{
		"user_id": userID,
		"block_time": bson.M{
			"$ne": 0,
		},
	})
	if err != nil {
		return follow, err
	}
	if err = cursor.All(ctx, &follow); err != nil {
		return follow, utils.Wrap(err, "Record not found")
	}
	return follow, nil

}

func (d *DataBases) GetArticleLike(articleID int64, userID string) (*ArticleLike, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleLike{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var like ArticleLike
	if err = c.FindOne(ctx, bson.M{"article_id": articleID, "user_id": userID, "delete_time": 0}).Decode(&like); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, utils.Wrap(err, "Record not found")
	}

	return &like, nil
}

func (d *DataBases) GetArticleFavorite(articleID int64, userID string) (*Favorites, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Favorites{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var favorites Favorites
	if err = c.FindOne(ctx, bson.M{"content_id": utils.Int64ToString(articleID), "user_id": userID, "delete_time": 0, "content_type": 2}).Decode(&favorites); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, utils.Wrap(err, "Record not found")
	}

	return &favorites, nil
}

func (d *DataBases) InsertArticleRead(articleID int64, userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	articleColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(Article{}.TableName())
	articleReadColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleRead{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	var articleRead ArticleRead
	err = articleReadColl.FindOne(ctx, bson.M{"article_id": articleID, "user_id": userID}).Decode(&articleRead)
	if err != nil && err != mongo.ErrNoDocuments {
		log.Error("", "get read failed", err.Error())
		return utils.Wrap(err, "get read failed")
	}

	newRead := ArticleRead{
		ArticleID:  articleID,
		UserID:     userID,
		CreateTime: time.Now().Unix(),
		Status:     1,
	}
	if _, err = articleReadColl.InsertOne(ctx, newRead); err != nil {
		log.Error("", "insert read failed", err.Error())
		return utils.Wrap(err, "insert read failed")
	}

	updatesMap := bson.M{"read_counts": 1}
	if articleRead.ArticleID == 0 {
		updatesMap["unique_read_counts"] = 1
	}

	if _, err = articleColl.UpdateOne(ctx, bson.M{"article_id": articleID, "delete_time": 0}, bson.M{"$inc": updatesMap, "$set": bson.M{"update_time": time.Now().Unix()}}); err != nil {
		log.Error("", "update article failed", err.Error())
		return utils.Wrap(err, "update article failed")
	}

	return nil
}

type ListUserArticleReadsResult struct {
	Article  Article `bson:"article"`
	ReadTime int64   `bson:"read_time"`
}

func (d *DataBases) ListUserArticleReads(userID string, minCreateTime, offset, limit int64) ([]ListUserArticleReadsResult, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	articleReadColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleRead{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	res, err := articleReadColl.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"status": 1, "user_id": userID, "create_time": bson.M{"$gte": minCreateTime}}}},
		bson.D{{"$group", bson.D{{"_id", "$article_id"}, {"read_time", bson.D{{"$last", "$create_time"}}}}}},
		bson.D{{"$sort", bson.D{{"read_time", -1}}}},
		bson.D{{"$lookup", bson.M{
			"from":         "article",
			"localField":   "_id",
			"foreignField": "article_id",
			"as":           "article",
		}}},
		bson.D{{"$unwind", "$article"}},
		bson.D{{"$match", bson.M{"article.delete_time": 0}}},
		bson.D{{"$skip", offset}},
		bson.D{{"$limit", limit}},
	})
	if err != nil {
		log.Error("", "get read list failed", err.Error())
		return nil, 0, utils.Wrap(err, "get read list failed")
	}

	var results []ListUserArticleReadsResult
	if err = res.All(ctx, &results); err != nil {
		log.Error("", "decode read list failed", err.Error())
		return nil, 0, utils.Wrap(err, "decode read list failed")
	}

	countRes, err := articleReadColl.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"status": 1, "user_id": userID, "create_time": bson.M{"$gte": minCreateTime}}}},
		bson.D{{"$group", bson.D{{"_id", "$article_id"}, {"read_time", bson.D{{"$last", "$create_time"}}}}}},
		bson.D{{"$lookup", bson.M{
			"from":         "article",
			"localField":   "_id",
			"foreignField": "article_id",
			"as":           "article",
		}}},
		bson.D{{"$unwind", "$article"}},
		bson.D{{"$match", bson.M{"article.delete_time": 0}}},
		bson.D{{"$count", "count"}},
	})

	var countResults struct {
		Count int64 `bson:"count"`
	}

	if countRes.Next(ctx) {
		if err = countRes.Decode(&countResults); err != nil {
			log.Error("", "decode article read count failed", err.Error())
			return nil, 0, utils.Wrap(err, "decode article read count failed")
		}
	}

	return results, countResults.Count, nil
}

func (d *DataBases) ClearUserArticleReads(userID string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	articleReadColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleRead{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	if _, err = articleReadColl.UpdateMany(ctx, bson.M{"user_id": userID, "status": 1}, bson.M{"$set": bson.M{"status": 2}}); err != nil {
		log.Error("", "clear read list failed", err.Error())
		return utils.Wrap(err, "clear read list failed")
	}

	return nil
}

type UserArtucleCommnentReply struct {
	CommentID       int64               `bson:"comment_id"`
	ParentCommentID int64               `bson:"parent_comment_id"`
	ArticleID       int64               `bson:"article_id"`
	OfficialID      int64               `bson:"official_id"`
	UserID          string              `bson:"user_id"`
	ReplyOfficialID int64               `bson:"reply_official_id"`
	ReplyUserID     string              `bson:"reply_user_id"`
	ReplyCounts     int64               `bson:"reply_counts"`
	LikeCounts      int64               `bson:"like_counts"`
	Content         string              `bson:"content"`
	CreateTime      int64               `bson:"created_time"`
	Like            *ArticleCommentLike `bson:"like"`
}

type UserArtucleCommnent struct {
	CommentID       int64                     `bson:"comment_id"`
	ArticleID       int64                     `bson:"article_id"`
	OfficialID      int64                     `bson:"official_id"`
	UserID          string                    `bson:"user_id"`
	ReplyOfficialID int64                     `bson:"reply_official_id"`
	ReplyUserID     string                    `bson:"reply_user_id"`
	ReplyCounts     int64                     `bson:"reply_counts"`
	LikeCounts      int64                     `bson:"like_counts"`
	Content         string                    `bson:"content"`
	CreateTime      int64                     `bson:"created_time"`
	Like            *ArticleCommentLike       `bson:"like"`
	TopReply        *UserArtucleCommnentReply `bson:"top_reply"`
}

func (d *DataBases) UpdateArticleComment(commentID int64, status int32) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	c := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{
		"comment_id": bson.M{
			"$eq": commentID,
		},
	}

	updateInter := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	_, err = c.UpdateOne(ctx, filter, updateInter)
	if err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) UpdateArticleCommentV2(articleComment ArticleComment) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	commentColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(articleComment.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{"comment_id": articleComment.CommentID, "delete_time": 0}
	update := bson.M{"$inc": bson.M{
		"reply_counts": articleComment.ReplyCounts,
		"like_counts":  articleComment.LikeCounts,
	}}
	if _, err = commentColl.UpdateOne(ctx, filter, update); err != nil {
		return utils.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *DataBases) ListUserArticleComments(userID string, articleID, offset, limit int64) ([]UserArtucleCommnent, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	articleCommentColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	cursor, err := articleCommentColl.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"delete_time": 0, "status": 1, "article_id": articleID, "parent_comment_id": 0}}},
		bson.D{{"$sort", bson.M{"created_time": 1}}},
		bson.D{{"$skip", offset}},
		bson.D{{"$limit", limit}},
		bson.D{{"$lookup", bson.M{
			"from": ArticleCommentLike{}.TableName(),
			"as":   "like",
			"let":  bson.M{"comment_id": "$comment_id"},
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"$expr": bson.M{"$and": []bson.M{
					bson.M{"$eq": []interface{}{"$comment_id", "$$comment_id"}},
					bson.M{"$eq": []interface{}{"$delete_time", 0}},
					bson.M{"$eq": []interface{}{"$status", 1}},
					bson.M{"$eq": []interface{}{"$user_id", userID}},
				}}}}},
				bson.D{{"$sort", bson.M{"like_counts": -1}}},
				bson.D{{"$limit", 1}},
			},
		}}},
		bson.D{{"$lookup", bson.M{
			"from": ArticleComment{}.TableName(),
			"as":   "top_reply",
			"let":  bson.M{"comment_id": "$comment_id"},
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"$expr": bson.M{"$and": []bson.M{
					bson.M{"$eq": []interface{}{"$parent_comment_id", "$$comment_id"}},
					bson.M{"$eq": []interface{}{"$delete_time", 0}},
					bson.M{"$eq": []interface{}{"$status", 1}},
					bson.M{"$eq": []interface{}{"$article_id", articleID}},
				}}}}},
				bson.D{{"$sort", bson.M{"like_counts": -1}}},
				bson.D{{"$limit", 2}},
				bson.D{{"$lookup", bson.M{
					"from": ArticleCommentLike{}.TableName(),
					"as":   "like",
					"let":  bson.M{"comment_id": "$comment_id"},
					"pipeline": mongo.Pipeline{
						bson.D{{"$match", bson.M{"$expr": bson.M{"$and": []bson.M{
							bson.M{"$eq": []interface{}{"$comment_id", "$$comment_id"}},
							bson.M{"$eq": []interface{}{"$delete_time", 0}},
							bson.M{"$eq": []interface{}{"$status", 1}},
							bson.M{"$eq": []interface{}{"$user_id", userID}},
						}}}}},
						//bson.D{{"$sort", bson.M{"like_counts": -1}}},
						bson.D{{"$limit", 1}},
					},
				}}},
				bson.D{{"$unwind", bson.M{"path": "$like", "preserveNullAndEmptyArrays": true}}},
			},
		}}},
		bson.D{{"$unwind", bson.M{"path": "$like", "preserveNullAndEmptyArrays": true}}},
		bson.D{{"$unwind", bson.M{"path": "$top_reply", "preserveNullAndEmptyArrays": true}}},
	})
	if err != nil {
		log.Error("", "get read list failed", err.Error())
		return nil, 0, utils.Wrap(err, "get read list failed")
	}

	var results []UserArtucleCommnent
	for cursor.Next(ctx) {
		var comment UserArtucleCommnent
		if err = cursor.Decode(&comment); err != nil {
			log.Error("", "decode comment list failed", err.Error())
			return nil, 0, utils.Wrap(err, "decode comment list failed")
		}
		results = append(results, comment)
	}

	countRes, err := articleCommentColl.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"delete_time": 0, "status": 1, "article_id": articleID, "parent_comment_id": 0}}},
		bson.D{{"$count", "count"}},
	})

	var countResults struct {
		Count int64 `bson:"count"`
	}

	if countRes.Next(ctx) {
		if err = countRes.Decode(&countResults); err != nil {
			log.Error("", "decode article read count failed", err.Error())
			return nil, 0, utils.Wrap(err, "decode article read count failed")
		}
	}

	return results, countResults.Count, nil
}

func (d *DataBases) ListUserArticleCommentReplies(userID string, commentID, offset, limit int64) ([]UserArtucleCommnent, int64, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	articleCommentColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		log.Error("", "start session error", err.Error())
		return nil, 0, utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	cursor, err := articleCommentColl.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"delete_time": 0, "status": 1, "parent_comment_id": commentID}}},
		bson.D{{"$sort", bson.D{{"create_time", -1}}}},
		bson.D{{"$skip", offset}},
		bson.D{{"$limit", limit}},
		bson.D{{"$lookup", bson.M{
			"from": ArticleCommentLike{}.TableName(),
			"as":   "like",
			"let":  bson.M{"comment_id": "$comment_id"},
			"pipeline": mongo.Pipeline{
				bson.D{{"$match", bson.M{"$expr": bson.M{"$and": []bson.M{
					bson.M{"$eq": []interface{}{"$comment_id", "$$comment_id"}},
					bson.M{"$eq": []interface{}{"$delete_time", 0}},
					bson.M{"$eq": []interface{}{"$status", 1}},
					bson.M{"$eq": []interface{}{"$user_id", userID}},
				}}}}},
				bson.D{{"$sort", bson.M{"like_counts": -1}}},
				bson.D{{"$limit", 1}},
			},
		}}},
		bson.D{{"$unwind", bson.M{"path": "$like", "preserveNullAndEmptyArrays": true}}},
	})
	if err != nil {
		log.Error("", "get read list failed", err.Error())
		return nil, 0, utils.Wrap(err, "get read list failed")
	}

	var results []UserArtucleCommnent
	for cursor.Next(ctx) {
		var comment UserArtucleCommnent
		if err = cursor.Decode(&comment); err != nil {
			log.Error("", "decode comment list failed", err.Error())
			return nil, 0, utils.Wrap(err, "decode comment list failed")
		}
		results = append(results, comment)
	}

	countRes, err := articleCommentColl.Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.M{"delete_time": 0, "status": 1, "parent_comment_id": commentID}}},
		bson.D{{"$count", "count"}},
	})

	var countResults struct {
		Count int64 `bson:"count"`
	}

	if countRes.Next(ctx) {
		if err = countRes.Decode(&countResults); err != nil {
			log.Error("", "decode article read count failed", err.Error())
			return nil, 0, utils.Wrap(err, "decode article read count failed")
		}
	}

	return results, countResults.Count, nil
}

func (d *DataBases) DeleteArticleCommentV2(commentID, articleID, parentCommentID, deleteTime int64, reqUser string) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(config.Config.Mongo.DBTimeout)*time.Second)
	commentColl := d.mongoClient.Database(config.Config.Mongo.DBDatabase).Collection(ArticleComment{}.TableName())
	session, err := d.mongoClient.StartSession()
	if err != nil {
		return utils.Wrap(err, "start session failed")
	}
	defer session.EndSession(ctx)

	filter := bson.M{"comment_id": commentID}
	update := bson.M{"$set": bson.M{
		"deleted_by":  reqUser,
		"delete_time": deleteTime,
	}}

	if _, err = commentColl.UpdateOne(ctx, filter, update); err != nil {
		return err
	}
	if parentCommentID != 0 {
		_, err = commentColl.UpdateOne(ctx, bson.M{"comment_id": parentCommentID}, bson.M{"$inc": bson.M{"reply_counts": -1}})
		if err != nil {
			return err
		}
	}
	return err
}
