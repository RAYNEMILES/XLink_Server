package db

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	log2 "Open_IM/pkg/common/log"
	pbAdminCMS "Open_IM/pkg/proto/admin_cms"
	pbChat "Open_IM/pkg/proto/chat"
	pbRtc "Open_IM/pkg/proto/rtc"
	pbCommon "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	go_redis "github.com/go-redis/redis/v9"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

//	func  (d *  DataBases)pubMessage(channel, msg string) {
//	  d.rdb.Publish(context.Background(),channel,msg)
//	}
//
//	func  (d *  DataBases)pubMessage(channel, msg string) {
//		d.rdb.Publish(context.Background(),channel,msg)
//	}
func (d *DataBases) JudgeAccountEXISTS(account string) (bool, error) {
	key := accountTempCode + account
	n, err := d.rdb.Exists(context.Background(), key).Result()
	if n > 0 {
		return true, err
	} else {
		return false, err
	}
}
func (d *DataBases) SetAccountCode(account string, code, ttl int) (err error) {
	key := accountTempCode + account
	return d.rdb.Set(context.Background(), key, code, time.Duration(ttl)*time.Second).Err()
}
func (d *DataBases) GetAccountCode(account string) (string, error) {
	key := accountTempCode + account
	return d.rdb.Get(context.Background(), key).Result()
}

func (d *DataBases) AccountCodeIsExists(account string) bool {
	key := accountTempCode + account
	resultInt, _ := d.rdb.Exists(context.Background(), key).Result()
	return resultInt > 0
}

func (d *DataBases) SetPrivacySettingByUserId(userID string, m map[string]string) (err error) {
	key := PrivacySettingKey + userID
	defer d.rdb.Expire(context.Background(), key, 3*time.Hour-3*time.Second)
	return d.rdb.HMSet(context.Background(), key, m).Err()
}

func (d *DataBases) GetPrivacySettingByUserId(userID string) (m map[string]string, err error) {
	key := PrivacySettingKey + userID
	return d.rdb.HGetAll(context.Background(), key).Result()
}

func (d *DataBases) DelPrivacySettingByUserId(userID string) (err error) {
	key := PrivacySettingKey + userID
	return d.rdb.Del(context.Background(), key).Err()
}

// Perform seq auto-increment operation of user messages
func (d *DataBases) IncrUserSeq(uid string) (uint64, error) {
	key := userIncrSeq + uid
	seq, err := d.rdb.Incr(context.Background(), key).Result()
	return uint64(seq), err
}

// Get the largest Seq
func (d *DataBases) GetUserMaxSeq(uid string) (uint64, error) {
	key := userIncrSeq + uid
	seq, err := d.rdb.Get(context.Background(), key).Result()
	return uint64(utils.StringToInt(seq)), err
}

// set the largest Seq
func (d *DataBases) SetUserMaxSeq(uid string, maxSeq uint64) error {
	key := userIncrSeq + uid
	return d.rdb.Set(context.Background(), key, maxSeq, 0).Err()
}

// Set the user's minimum seq
func (d *DataBases) SetUserMinSeq(uid string, minSeq uint32) (err error) {
	key := userMinSeq + uid
	return d.rdb.Set(context.Background(), key, minSeq, 0).Err()
}

// Get the smallest Seq
func (d *DataBases) GetUserMinSeq(uid string) (uint32, error) {
	key := userMinSeq + uid
	seq, err := d.rdb.Get(context.Background(), key).Result()
	return uint32(utils.StringToInt(seq)), err
}

// Perform seq auto-increment operation of group messages
func (d *DataBases) IncrGroupSeq(uid string) (uint64, error) {
	key := GroupIncrSeq + uid
	seq, err := d.rdb.Incr(context.Background(), key).Result()
	return uint64(seq), err
}

// Get the largest Seq for group
func (d *DataBases) GetGrupMaxSeq(uid string) (uint64, error) {
	key := GroupIncrSeq + uid
	seq, err := d.rdb.Get(context.Background(), key).Result()
	return uint64(utils.StringToInt(seq)), err
}

// set the largest Seq for group
func (d *DataBases) SetGroupMaxSeq(uid string, maxSeq uint64) error {
	key := GroupIncrSeq + uid
	return d.rdb.Set(context.Background(), key, maxSeq, 0).Err()
}

// Set the group's minimum seq
func (d *DataBases) SetGroupMinSeq(uid string, groupId string, minSeq uint32) (err error) {
	key := GroupMinSeq + uid + "-" + groupId
	return d.rdb.Set(context.Background(), key, minSeq, 0).Err()
}

// Get the smallest Seq for group
func (d *DataBases) GetGroupMinSeq(uid string, groupId string) (uint32, error) {
	key := GroupMinSeq + uid + "-" + groupId
	seq, err := d.rdb.Get(context.Background(), key).Result()
	return uint32(utils.StringToInt(seq)), err
}

// Store userid and platform class to redis
func (d *DataBases) AddTokenFlag(userID string, platformID int, token string, flag int) error {
	key := uidPidToken + userID + ":" + constant.PlatformIDToName(platformID)
	log2.NewDebug("", "add token key is ", key)
	return d.rdb.HSet(context.Background(), key, token, flag).Err()
}

func (d *DataBases) GetTokenMapByUidPid(userID, platformID string) (map[string]interface{}, error) {
	key := uidPidToken + userID + ":" + platformID
	log2.NewDebug("", "get token key is ", key)
	m, err := d.rdb.HGetAll(context.Background(), key).Result()
	mm := make(map[string]interface{})
	for k, v := range m {
		mm[k] = utils.StringToInt(v)
	}
	return mm, err
}
func (d *DataBases) SetTokenMapByUidPid(userID string, platformID int, m map[string]interface{}) error {
	key := uidPidToken + userID + ":" + constant.PlatformIDToName(platformID)
	//for k, v := range m {
	//	err := d.rdb.HSet(context.Background(), key, k, v).Err()
	//	if err != nil {
	//		return err
	//	}
	//}
	//return nil
	return d.rdb.HMSet(context.Background(), key, m).Err()
}
func (d *DataBases) DeleteTokenByUidPid(userID string, platformID int, fields []string) error {
	key := uidPidToken + userID + ":" + constant.PlatformIDToName(platformID)
	return d.rdb.HDel(context.Background(), key, fields...).Err()
}
func (d *DataBases) DeleteTokenByUidWOPid(userID string) error {
	key := uidPidToken + userID
	return d.rdb.Del(context.Background(), key).Err()
}

func (d *DataBases) DeleteTokenByUid(userID string, platformID int) error {
	key := uidPidToken + userID + ":" + constant.PlatformIDToName(platformID)
	return d.rdb.Del(context.Background(), key).Err()
}

func (d *DataBases) SetSingleConversationRecvMsgOpt(userID, conversationID string, opt int32) error {
	key := conversationReceiveMessageOpt + userID
	return d.rdb.HSet(context.Background(), key, conversationID, opt).Err()
}

func (d *DataBases) GetSingleConversationRecvMsgOpt(userID, conversationID string) (int, error) {
	key := conversationReceiveMessageOpt + userID
	result, err := d.rdb.HGet(context.Background(), key, conversationID).Result()
	return utils.StringToInt(result), err
}
func (d *DataBases) SetUserGlobalMsgRecvOpt(userID string, opt int32) error {
	key := conversationReceiveMessageOpt + userID
	return d.rdb.HSet(context.Background(), key, GlobalMsgRecvOpt, opt).Err()
}
func (d *DataBases) GetUserGlobalMsgRecvOpt(userID string) (int, error) {
	key := conversationReceiveMessageOpt + userID
	result, err := d.rdb.HGet(context.Background(), key, GlobalMsgRecvOpt).Result()
	if err != nil {
		if err == go_redis.Nil {
			return 0, nil
		} else {
			return 0, err
		}
	}
	return utils.StringToInt(result), err
}
func (d *DataBases) GetMessageListBySeq(userID string, chatType int, seqList []uint32, operationID string) (seqMsg []*pbCommon.MsgData, failedSeqList []uint32, errResult error) {
	for _, v := range seqList {
		var key string
		if chatType == constant.SingleChatType {
			key = messageCache + userID + "_" + strconv.Itoa(int(v))
		} else if chatType == constant.GroupChatType {
			key = GroupMessageCache + userID + "_" + strconv.Itoa(int(v))
		}

		result, err := d.rdb.Get(context.Background(), key).Result()
		if err != nil {
			errResult = err
			failedSeqList = append(failedSeqList, v)
			log2.NewWarn(operationID, "redis get message error:", err.Error(), v)
		} else {
			msg := pbCommon.MsgData{}
			err = jsonpb.UnmarshalString(result, &msg)
			if err != nil {
				errResult = err
				failedSeqList = append(failedSeqList, v)
				log2.NewWarn(operationID, "Unmarshal err", result, err.Error())
			} else {
				log2.NewDebug(operationID, "redis get msg is ", msg.String())
				seqMsg = append(seqMsg, &msg)
			}

		}
	}
	return seqMsg, failedSeqList, errResult
}

// set private chat and group chat messages to redis
func (d *DataBases) SetMessageToCache(msgList []*pbChat.MsgDataToMQ, uid string, chatType int, operationID string) error {
	ctx := context.Background()
	pipe := d.rdb.Pipeline()
	var failedList []pbChat.MsgDataToMQ
	for _, msg := range msgList {

		var key string
		if chatType == constant.SingleChatType {
			key = messageCache + uid + "_" + strconv.Itoa(int(msg.MsgData.Seq))
		} else if chatType == constant.GroupChatType {
			key = GroupMessageCache + uid + "_" + strconv.Itoa(int(msg.MsgData.Seq))
		}

		if key != "" {
			s, err := utils.Pb2String(msg.MsgData)
			if err != nil {
				log2.NewWarn(operationID, utils.GetSelfFuncName(), "Pb2String failed", msg.MsgData.String(), uid, err.Error())
				continue
			}
			log2.NewDebug(operationID, "convert string is ", s)
			err = pipe.Set(ctx, key, s, time.Duration(config.Config.MsgCacheTimeout)*time.Second).Err()
			//err = d.rdb.HMSet(context.Background(), "12", map[string]interface{}{"1": 2, "343": false}).Err()
			if err != nil {
				log2.NewWarn(operationID, utils.GetSelfFuncName(), "redis failed", "args:", key, *msg, uid, s, err.Error())
				failedList = append(failedList, *msg)
			}
		} else {
			continue
		}

	}
	if len(failedList) != 0 {
		return errors.New(fmt.Sprintf("set msg to cache failed, failed lists: %q,%s", failedList, operationID))
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (d *DataBases) CleanUpOneUserAllMsgFromRedis(userID string, operationID string) error {
	ctx := context.Background()
	key := messageCache + userID + "_" + "*"
	vals, err := d.rdb.Keys(ctx, key).Result()
	log2.Debug(operationID, "vals: ", vals)
	if err == go_redis.Nil {
		return nil
	}
	if err != nil {
		return utils.Wrap(err, "")
	}
	if err = d.rdb.Del(ctx, vals...).Err(); err != nil {
		return utils.Wrap(err, "")
	}
	return nil
}

func (d *DataBases) HandleSignalInfo(operationID string, msg *pbCommon.MsgData) error {
	req := &pbRtc.SignalReq{}
	if err := proto.Unmarshal(msg.Content, req); err != nil {
		return err
	}
	//log.NewDebug(pushMsg.OperationID, utils.GetSelfFuncName(), "SignalReq: ", req.String())
	var inviteeUserIDList []string
	var isInviteSignal bool
	switch signalInfo := req.Payload.(type) {
	case *pbRtc.SignalReq_Invite:
		inviteeUserIDList = signalInfo.Invite.Invitation.InviteeUserIDList
		isInviteSignal = true
	case *pbRtc.SignalReq_InviteInGroup:
		inviteeUserIDList = signalInfo.InviteInGroup.Invitation.InviteeUserIDList
		isInviteSignal = true
	case *pbRtc.SignalReq_HungUp, *pbRtc.SignalReq_Cancel, *pbRtc.SignalReq_Reject, *pbRtc.SignalReq_Accept:
		return errors.New("signalInfo do not need offlinePush")
	default:
		log2.NewDebug(operationID, utils.GetSelfFuncName(), "req invalid type", string(msg.Content))
		return nil
	}
	if isInviteSignal {
		log2.NewInfo(operationID, utils.GetSelfFuncName(), "invite userID list:", inviteeUserIDList)
		for _, userID := range inviteeUserIDList {
			log2.NewInfo(operationID, utils.GetSelfFuncName(), "invite userID:", userID)
			timeout, err := strconv.Atoi(config.Config.Rtc.SignalTimeout)
			if err != nil {
				return err
			}
			keyList := SignalListCache + userID
			err = d.rdb.LPush(context.Background(), keyList, msg.ClientMsgID).Err()
			if err != nil {
				return err
			}
			err = d.rdb.Expire(context.Background(), keyList, time.Duration(timeout)*time.Second).Err()
			if err != nil {
				return err
			}
			key := SignalCache + msg.ClientMsgID
			err = d.rdb.Set(context.Background(), key, msg.Content, time.Duration(timeout)*time.Second).Err()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *DataBases) GetSignalInfoFromCacheByClientMsgID(clientMsgID string) (invitationInfo *pbRtc.SignalInviteReq, err error) {
	key := SignalCache + clientMsgID
	invitationInfo = &pbRtc.SignalInviteReq{}
	bytes, err := d.rdb.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, err
	}
	req := &pbRtc.SignalReq{}
	if err = proto.Unmarshal(bytes, req); err != nil {
		return nil, err
	}
	switch req2 := req.Payload.(type) {
	case *pbRtc.SignalReq_Invite:
		invitationInfo.Invitation = req2.Invite.Invitation
		invitationInfo.OpUserID = req2.Invite.OpUserID
	case *pbRtc.SignalReq_InviteInGroup:
		invitationInfo.Invitation = req2.InviteInGroup.Invitation
		invitationInfo.OpUserID = req2.InviteInGroup.OpUserID
	}
	return invitationInfo, err
}

func (d *DataBases) GetAvailableSignalInvitationInfo(userID string) (invitationInfo *pbRtc.SignalInviteReq, err error) {
	keyList := SignalListCache + userID
	result := d.rdb.LPop(context.Background(), keyList)
	if err = result.Err(); err != nil {
		return nil, utils.Wrap(err, "GetAvailableSignalInvitationInfo failed")
	}
	key, err := result.Result()
	if err != nil {
		return nil, utils.Wrap(err, "GetAvailableSignalInvitationInfo failed")
	}
	log2.NewDebug("", utils.GetSelfFuncName(), result, result.String())
	invitationInfo, err = d.GetSignalInfoFromCacheByClientMsgID(key)
	if err != nil {
		return nil, utils.Wrap(err, "GetSignalInfoFromCacheByClientMsgID")
	}
	err = d.DelUserSignalList(userID)
	if err != nil {
		return nil, utils.Wrap(err, "GetSignalInfoFromCacheByClientMsgID")
	}
	return invitationInfo, nil
}

func (d *DataBases) DelUserSignalList(userID string) error {
	keyList := SignalListCache + userID
	err := d.rdb.Del(context.Background(), keyList).Err()
	return err
}

func (d *DataBases) DelMsgFromCache(uid string, chatType int, seqList []uint32, operationID string) {
	for _, seq := range seqList {
		var key string
		if chatType == constant.SingleChatType {
			key = messageCache + uid + "_" + strconv.Itoa(int(seq))
		} else if chatType == constant.GroupChatType {
			key = GroupMessageCache + uid + "_" + strconv.Itoa(int(seq))
		}

		result := d.rdb.Get(context.Background(), key).String()
		var msg pbCommon.MsgData
		if err := utils.String2Pb(result, &msg); err != nil {
			log2.Error(operationID, utils.GetSelfFuncName(), "String2Pb failed", msg, err.Error())
			continue
		}
		msg.Status = constant.MsgDeleted
		s, err := utils.Pb2String(&msg)
		if err != nil {
			log2.Error(operationID, utils.GetSelfFuncName(), "Pb2String failed", msg, err.Error())
			continue
		}
		if err := d.rdb.Set(context.Background(), key, s, time.Duration(config.Config.MsgCacheTimeout)*time.Second).Err(); err != nil {
			log2.Error(operationID, utils.GetSelfFuncName(), "Set failed", err.Error())
		}
	}
}

func (d *DataBases) SetGetuiToken(token string, expireTime int64) error {
	return d.rdb.Set(context.Background(), getuiToken, token, time.Duration(expireTime)*time.Second).Err()
}

func (d *DataBases) GetGetuiToken() (string, error) {
	result := d.rdb.Get(context.Background(), getuiToken)
	return result.String(), result.Err()
}

func (d *DataBases) AddFriendToCache(userID string, friendIDList ...string) error {
	var IDList []interface{}
	for _, id := range friendIDList {
		IDList = append(IDList, id)
	}
	return d.rdb.SAdd(context.Background(), friendRelationCache+userID, IDList...).Err()
}

func (d *DataBases) ReduceFriendToCache(userID string, friendIDList ...string) error {
	var IDList []interface{}
	for _, id := range friendIDList {
		IDList = append(IDList, id)
	}
	return d.rdb.SRem(context.Background(), friendRelationCache+userID, IDList...).Err()
}

func (d *DataBases) GetFriendIDListFromCache(userID string) ([]string, error) {
	result := d.rdb.SMembers(context.Background(), friendRelationCache+userID)
	return result.Result()
}

func (d *DataBases) AddBlackUserToCache(userID string, blackList ...string) error {
	var IDList []interface{}
	for _, id := range blackList {
		IDList = append(IDList, id)
	}
	return d.rdb.SAdd(context.Background(), blackListCache+userID, IDList...).Err()
}
func (d *DataBases) AddBlackUserForMomentsToCache(userID string, blackList ...string) error {
	var IDList []interface{}
	for _, id := range blackList {
		IDList = append(IDList, id)
	}
	return d.rdb.SAdd(context.Background(), blackListForMomentsCache+userID, IDList...).Err()
}
func (d *DataBases) ReduceBlackUserFromCache(userID string, blackList ...string) error {
	var IDList []interface{}
	for _, id := range blackList {
		IDList = append(IDList, id)
	}
	return d.rdb.SRem(context.Background(), blackListCache+userID, IDList...).Err()
}
func (d *DataBases) ReduceBlackUserForMomentsFromCache(userID string, blackList ...string) error {
	var IDList []interface{}
	for _, id := range blackList {
		IDList = append(IDList, id)
	}
	return d.rdb.SRem(context.Background(), blackListForMomentsCache+userID, IDList...).Err()
}

func (d *DataBases) GetBlackListFromCache(userID string) ([]string, error) {
	result := d.rdb.SMembers(context.Background(), blackListCache+userID)
	return result.Result()
}

func (d *DataBases) GetBlackListForMomentFromCache(userID string) ([]string, error) {
	result := d.rdb.SMembers(context.Background(), blackListForMomentsCache+userID)
	return result.Result()
}

func (d *DataBases) AddGroupMemberToCache(groupID string, userIDList ...string) error {
	var IDList []interface{}
	for _, id := range userIDList {
		IDList = append(IDList, id)
	}
	return d.rdb.SAdd(context.Background(), groupCache+groupID, IDList...).Err()
}

func (d *DataBases) ReduceGroupMemberFromCache(groupID string, userIDList ...string) error {
	var IDList []interface{}
	for _, id := range userIDList {
		IDList = append(IDList, id)
	}
	return d.rdb.SRem(context.Background(), groupCache+groupID, IDList...).Err()
}

func (d *DataBases) GetGroupMemberIDListFromCache(groupID string) ([]string, error) {
	result := d.rdb.SMembers(context.Background(), groupCache+groupID)
	return result.Result()
}

func (d *DataBases) SetUserInfoToCache(userID string, m map[string]interface{}) error {
	return d.rdb.HSet(context.Background(), userInfoCache+userID, m).Err()
}

func (d *DataBases) GetUserInfoFromCache(userID string) (*pbCommon.UserInfo, error) {
	result, err := d.rdb.HGetAll(context.Background(), userInfoCache+userID).Result()
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	userInfo := &pbCommon.UserInfo{}
	if err := proto.Unmarshal(bytes, userInfo); err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, userInfo)
	return userInfo, err
}

func (d *DataBases) SaveVerificationCode(id, answer string) error {
	return d.rdb.SetEx(context.Background(), VerificationCode+id, answer, 179).Err()
}

func (d *DataBases) SaveUserStatus(userID string, status int32) error {
	return d.rdb.Set(context.Background(), constant.UserStatusKey+"_"+userID, status, 0).Err()
}

func (d *DataBases) GetUserStatus(userID string) (string, error) {
	return d.rdb.Get(context.Background(), constant.UserStatusKey+"_"+userID).Result()
}

func (d *DataBases) SaveAdminStatus(userID string, status int32) error {
	return d.rdb.Set(context.Background(), constant.AdminStatusKey+"_"+userID, status, 0).Err()
}

func (d *DataBases) GetAdminStatus(userID string) (string, error) {
	return d.rdb.Get(context.Background(), constant.AdminStatusKey+"_"+userID).Result()
}

func (d *DataBases) SaveInAppLoginPin(userID string, pinCode int) error {
	return d.rdb.Set(context.Background(), constant.InAppLoginPin+":"+userID, pinCode, 5*time.Minute).Err()
}

func (d *DataBases) GetInAppLoginPin(userID string) (string, error) {
	return d.rdb.Get(context.Background(), constant.InAppLoginPin+":"+userID).Result()
}

func (d *DataBases) SetTimeLimitOnActionType(actionType, creatorToken, valueStored string) error {
	creatorTData := []byte(creatorToken)
	creatorTMD5 := fmt.Sprintf("%x", md5.Sum(creatorTData))
	key := actionType + creatorTMD5
	timeLimit := time.Duration(config.Config.RestrictUserActionTimeLimit)
	return d.rdb.Set(context.Background(), key, valueStored, time.Second*timeLimit).Err()
}

func (d *DataBases) GetTimeLimitOnActionType(actionType, creatorToken string) (string, error) {
	creatorTData := []byte(creatorToken)
	creatorTMD5 := fmt.Sprintf("%x", md5.Sum(creatorTData))
	key := actionType + creatorTMD5
	return d.rdb.Get(context.Background(), key).Result()
}

func (d *DataBases) SetConfigCache(name, value string) error {
	return d.rdb.SetEx(context.Background(), ConfigCacheType+name, value, 299*time.Second).Err()
}

func (d *DataBases) GetConfigCache(name string) (string, error) {
	return d.rdb.Get(context.Background(), ConfigCacheType+name).Result()
}

func (d *DataBases) DeleteConfigCache(name string) error {
	return d.rdb.Del(context.Background(), ConfigCacheType+name).Err()
}

func (d *DataBases) SetUserRemarkbyFriend(keyPrefix, owenerUserID, ForFriendUserID, remark string) error {
	key := keyPrefix + owenerUserID + "-" + ForFriendUserID
	return d.rdb.Set(context.Background(), key, remark, 0).Err()
}
func (d *DataBases) GetUserRemarkbyFriend(keyPrefix, owenerUserID, ForFriendUserID string) (string, error) {
	key := keyPrefix + owenerUserID + "-" + ForFriendUserID
	isExsist, _ := d.rdb.Exists(context.Background(), key).Result()
	if isExsist == 1 {
		return d.rdb.Get(context.Background(), key).Result()
	}
	return "", errors.New("not esxixst in redis")

}
func (d *DataBases) DelUserRemarkbyFriend(keyPrefix, owenerUserID, ForFriendUserID string) error {
	key := keyPrefix + owenerUserID + "-" + ForFriendUserID
	return d.rdb.Del(context.Background(), key).Err()
}

func (d *DataBases) SetUserNickNameByGroup(keyPrefix, groupID, userID, nickName string) error {
	key := keyPrefix + groupID + "-" + userID
	return d.rdb.Set(context.Background(), key, nickName, 0).Err()
}
func (d *DataBases) GetUserNickNameByGroup(keyPrefix, groupID, userID string) (string, error) {
	key := keyPrefix + groupID + "-" + userID
	isExsist, _ := d.rdb.Exists(context.Background(), key).Result()
	if isExsist == 1 {
		return d.rdb.Get(context.Background(), key).Result()
	}
	return "", errors.New("not esxixst in redis")

}

func (d *DataBases) SaveUserIPandStatus(userID string, jsonString string) error {
	return d.rdb.Set(context.Background(), constant.UserIPandStatus+"_"+userID, jsonString, 0).Err()
}

func (d *DataBases) GetUserIPandStatus(userID string) (string, error) {
	return d.rdb.Get(context.Background(), constant.UserIPandStatus+"_"+userID).Result()
}

func (d *DataBases) SaveAdminUserStatus(userID string, status int32) error {
	return d.rdb.Set(context.Background(), constant.AdminStatusKey+":"+userID, status, 0).Err()
}

func (d *DataBases) SaveGroupIDListForUser(userID string, groupIDList []interface{}) error {
	return d.rdb.SAdd(context.Background(), constant.UserGroupIDCache+":"+userID, groupIDList).Err()
}

func (d *DataBases) SaveGroupIDForUser(userID, groupID string) error {
	return d.rdb.SAdd(context.Background(), constant.UserGroupIDCache+":"+userID, []interface{}{groupID}).Err()
}

func (d *DataBases) RemoveGroupIDForUser(userID, groupID string) error {
	return d.rdb.SRem(context.Background(), constant.UserGroupIDCache+":"+userID, []interface{}{groupID}).Err()
}

func (d *DataBases) GetGroupIDListForUser(userID string) ([]string, error) {
	return d.rdb.SMembers(context.Background(), constant.UserGroupIDCache+":"+userID).Result()
}

func (d *DataBases) SaveUuid(Uuid string) error {
	return d.rdb.Incr(context.Background(), RegisterByUuid+Uuid).Err()
}

func (d *DataBases) GetRegisterNumByUuid(Uuid string) (int64, error) {
	return d.rdb.Get(context.Background(), RegisterByUuid+Uuid).Int64()
}

func (d *DataBases) StoreAllAdminRolesInRedis(adminRoles []pbAdminCMS.AdminRole) {
	for _, adminRole := range adminRoles {
		key := constant.AdminRolesKey + ":"
		// for _, adminAction := range adminRole.AdminActions {
		for _, allowedAPI := range adminRole.AllowedApis {
			Apikey := key + allowedAPI.ApiPath
			err := d.rdb.SAdd(context.Background(), Apikey, adminRole.Id).Err()
			if err != nil {
				log2.NewDebug("Sandman", utils.GetSelfFuncName(), err.Error())
			}

		}
		for _, allowedPages := range adminRole.AllowedPages {
			Apikey := key + allowedPages.PagePath
			err := d.rdb.SAdd(context.Background(), Apikey, adminRole.Id).Err()
			if err != nil {
				log2.NewDebug("Sandman", utils.GetSelfFuncName(), err.Error())
			}
		}
		// }
	}
}

func (d *DataBases) CheckAdminRolesAllowedInRedis(path string, RoleID string) bool {

	key := constant.AdminRolesKey + ":" + path
	scan := d.rdb.SMembers(context.Background(), key)
	result, err := scan.Result()
	if err != nil {
		log2.NewDebug("Sandman Permission", utils.GetSelfFuncName(), err.Error())
		return false
	}
	if contains(result, RoleID) {
		return true
	}
	log2.NewDebug("Sandman Permission", utils.GetSelfFuncName(), "Role ID not found in results ", key)
	return false

}

func (d *DataBases) SetUsersInfoByUserIdCache(userId, cache string, expireTime int64) error {
	key := PhoneByUid + userId
	return d.rdb.Set(context.Background(), key, cache, time.Duration(expireTime)*time.Second).Err()
}

func (d *DataBases) GetUsersInfoByUserIdCache(userId string) (string, error) {
	key := PhoneByUid + userId
	return d.rdb.Get(context.Background(), key).Result()
}
func (d *DataBases) DelUsersInfoByUserIdCache(userId string) error {
	key := PhoneByUid + userId
	return d.rdb.Del(context.Background(), key).Err()
}

func (d *DataBases) SetInterestListCache(cache string, expireTime int64) error {
	key := Interest + "list"
	return d.rdb.Set(context.Background(), key, cache, time.Duration(expireTime)*time.Second).Err()
}

func (d *DataBases) GetInterestListCache() (string, error) {
	key := Interest + "list"
	return d.rdb.Get(context.Background(), key).Result()
}

func (d *DataBases) SetInterestListByLanguage(language, cache string) error {
	return d.rdb.SetEx(context.Background(), Interest+"list:"+language, cache, time.Duration(597)*time.Second).Err()
}

func (d *DataBases) GetInterestListByLanguage(language string) (string, error) {
	return d.rdb.Get(context.Background(), Interest+"list:"+language).Result()
}

func (d *DataBases) DeleteInterestList() error {
	return d.rdb.Del(context.Background(), Interest+"list:*").Err()
}

// SaveGroupMessageStatistic push群聊列表
func (d *DataBases) SaveGroupMessageStatistic(value string) error {
	key := Statistics + "group"
	return d.rdb.LPush(context.Background(), key, value).Err()
}

// PopGroupMessageStatistic pop群聊列表
func (d *DataBases) PopGroupMessageStatistic() (string, error) {
	key := Statistics + "group"
	return d.rdb.RPop(context.Background(), key).Result()
}

// SaveUsersByGroupMessage 今天有群聊的用户
func (d *DataBases) SaveUsersByGroupMessage(groupId, userId string, date ...string) error {
	key := ""
	if len(date) > 0 {
		key = Statistics + "users" + date[0] + groupId
	} else {
		key = Statistics + "users" + time.Now().Format("0601") + groupId
	}

	defer d.rdb.Expire(context.Background(), key, 72*time.Hour)
	return d.rdb.SAdd(context.Background(), key, userId).Err()
}
func (d *DataBases) GetUsersCountByGroupMessage(groupId string, date ...string) (int64, error) {
	key := ""
	if len(date) > 0 {
		key = Statistics + "users" + date[0] + groupId
	} else {
		key = Statistics + "users" + time.Now().Format("0601") + groupId
	}
	return d.rdb.SCard(context.Background(), key).Result()
}

// IncrGroupMessageCountByDay 群聊次数
func (d *DataBases) IncrGroupMessageCountByDay(groupId string, date ...string) error {
	key := ""
	if len(date) > 0 {
		key = Statistics + "mc" + date[0] + groupId
	} else {
		key = Statistics + "mc" + time.Now().Format("0601") + groupId
	}

	defer d.rdb.Expire(context.Background(), key, 72*time.Hour)
	return d.rdb.Incr(context.Background(), key).Err()
}
func (d *DataBases) GetGroupMessageCountByDay(groupId string, date ...string) (string, error) {
	key := ""
	if len(date) > 0 {
		key = Statistics + "mc" + date[0] + groupId
	} else {
		key = Statistics + "mc" + time.Now().Format("0601") + groupId
	}

	return d.rdb.Get(context.Background(), key).Result()
}

// SaveNeedCountGroup 需要统计的群
func (d *DataBases) SaveNeedCountGroup(groupId string, date ...string) error {
	key := ""
	if len(date) > 0 {
		key = Statistics + "nc" + date[0]
	} else {
		key = Statistics + "nc" + time.Now().Format("0601")
	}

	defer d.rdb.Expire(context.Background(), key, 72*time.Hour)
	return d.rdb.SAdd(context.Background(), key, groupId).Err()
}
func (d *DataBases) GetNeedCountGroup(date ...string) ([]string, error) {
	key := ""
	if len(date) > 0 {
		key = Statistics + "nc" + date[0]
	} else {
		key = Statistics + "nc" + time.Now().Format("0601")
	}

	return d.rdb.SMembers(context.Background(), key).Result()
}
func (d *DataBases) SaveInterestGroupINfoListByUserId(userID, cache string) error {
	return d.rdb.Set(context.Background(), Interest+"interest"+userID, cache, time.Duration(597)*time.Second).Err()
}
func (d *DataBases) GetInterestGroupINfoListByUserId(userID string) (string, error) {
	return d.rdb.Get(context.Background(), Interest+"interest"+userID).Result()
}
func (d *DataBases) DeleteInterestGroupINfoListByUserId(userID string) error {
	return d.rdb.Del(context.Background(), Interest+"interest"+userID).Err()
}

func (d *DataBases) SetProgress(userToken string, progress int, expireTime int64) error {
	key := UploadProgress + userToken
	return d.rdb.Set(context.Background(), key, progress, time.Duration(expireTime)*time.Minute).Err()
}

func (d *DataBases) GetProgress(userToken string) (int, error) {
	return d.rdb.Get(context.Background(), UploadProgress+userToken).Int()
}

func (d *DataBases) SetUserBroadcastStatus(userID string, cache int32) error {
	return d.rdb.Set(context.Background(), BroadcastStatus+userID, cache, time.Duration(-1)*time.Minute).Err()
}

func (d *DataBases) GetUserBroadcastStatus(userID string) (int, error) {
	return d.rdb.Get(context.Background(), BroadcastStatus+userID).Int()
}

func (d *DataBases) SetUserCommunicationStatus(userID string, status int) error {
	return d.rdb.Set(context.Background(), Communication+userID, status, time.Hour*24).Err()
}

func (d *DataBases) GetUserCommunicationStatus(userID string) (int, error) {
	return d.rdb.Get(context.Background(), Communication+userID).Int()
}

func (d *DataBases) SetMemberToCommunication(roomId, userID string) (int64, error) {
	return d.rdb.SAdd(context.Background(), Communication+roomId, userID).Result()
}

func (d *DataBases) GetMemberToCommunication(callID string) ([]string, error) {
	key := Communication + callID
	return d.rdb.SMembers(context.Background(), key).Result()
}

func (d *DataBases) RemoveMemberToCommunication(callID string) error {
	return d.rdb.Del(context.Background(), Communication+callID).Err()
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (d *DataBases) SetLatestArticlesByOfficialAccount(article ArticleSQL) error {
	articleByte, err := json.Marshal(article)
	if err != nil {
		return err
	}
	key := OfficialLatestAccount + strconv.FormatInt(article.OfficialID, 10)
	bo := d.rdb.Set(context.Background(), key, string(articleByte), 0)
	return bo.Err()
}
func (d *DataBases) RemoveLatestArticlesByOfficialAccount(officialID int64) error {
	key := OfficialLatestAccount + strconv.FormatInt(officialID, 10)
	bo := d.rdb.Del(context.Background(), key)
	return bo.Err()
}

func (d *DataBases) GetLatestArticlesByOfficialAccount(officialID int64) (*ArticleSQL, error) {

	key := OfficialLatestAccount + strconv.FormatInt(officialID, 10)
	bo := d.rdb.Get(context.Background(), key)

	article := ArticleSQL{}
	articleString, err := bo.Result()
	if err == nil {
		err := json.Unmarshal([]byte(articleString), &article)
		if err != nil {
			log2.NewError("FollowedOfficialConversation Official Redis ", err.Error())
			return nil, err
		}
		return &article, err
	}

	return nil, err
}
