package db

import (
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"context"
	"errors"
	"strconv"
	"time"
)

func (d *DataBases) ShortVideoFollowLimit(userId string) bool {
	key := ShortVideoFollowLimit + userId
	return d.rdb.SetNX(context.Background(), key, 1, time.Second).Val()
}

func (d *DataBases) ShortVideoFollowAddMember(userId, followUserId string) bool {
	key := ShortVideoFollowUserIdList + userId
	return d.rdb.SAdd(context.Background(), key, followUserId).Val() > 0
}

func (d *DataBases) ShortVideoFollowRemMember(userId, followUserId string) bool {
	key := ShortVideoFollowUserIdList + userId
	return d.rdb.SRem(context.Background(), key, followUserId).Val() > 0
}

func (d *DataBases) ShortVideoDeleteFollowUserListCache(userId string) bool {
	key := ShortVideoFollowUserIdList + userId
	return d.rdb.Del(context.Background(), key).Val() > 0
}

func (d *DataBases) ShortVideoSetFollowUserListCache(userId string, followUserIds []interface{}) error {
	key := ShortVideoFollowUserIdList + userId
	defer d.rdb.Expire(context.Background(), key, 8*time.Hour-1)
	return d.rdb.SAdd(context.Background(), key, followUserIds...).Err()
}

func (d *DataBases) ShortVideoGetFollowUserListCache(userId string) ([]string, error) {
	key := ShortVideoFollowUserIdList + userId
	return d.rdb.SMembers(context.Background(), key).Result()
}

func (d *DataBases) ShortVideoIsFollow(userId, followUserId string) (bool, error) {
	key := ShortVideoFollowUserIdList + userId

	result, _ := d.rdb.Exists(context.Background(), key).Result()
	if result == 1 {
		defer d.rdb.Expire(context.Background(), key, 8*time.Hour-1)
		return d.rdb.SIsMember(context.Background(), key, followUserId).Val(), nil
	}
	return false, errors.New("key not exist")
}

func (d *DataBases) ShortVideoLikeLimit(userId, FileId string, like bool) bool {
	key := ShortVideoLikeLimit + userId + strconv.FormatBool(like) + FileId
	return d.rdb.SetNX(context.Background(), key, 1, time.Second).Val()
}

func (d *DataBases) ShortVideoRecommendLimit(userId string) bool {
	key := ShortVideoRecommendLimit + userId
	return d.rdb.SetNX(context.Background(), key, 2, time.Second).Val()
}

func (d *DataBases) ShortVideoCommentLimit(userId, FileId string) bool {
	key := ShortVideoCommentLimit + userId + FileId
	return d.rdb.SetNX(context.Background(), key, 1, 3*time.Second).Val()
}

func (d *DataBases) ShortVideoSetLike(userId, FileId string) bool {
	key := ShortVideoLikeSet + FileId
	return d.rdb.SAdd(context.Background(), key, userId).Val() > 0
}

func (d *DataBases) ShortVideoCancelLike(userId, FileId string) bool {
	key := ShortVideoLikeSet + FileId
	return d.rdb.SRem(context.Background(), key, userId).Val() > 0
}

func (d *DataBases) ShortVideoIsLike(userId, FileId string) bool {
	key := ShortVideoLikeSet + FileId
	return d.rdb.SIsMember(context.Background(), key, userId).Val()
}

func (d *DataBases) ShortVideoDeleteLikeListByFileId(FileId string) bool {
	key := ShortVideoLikeSet + FileId
	return d.rdb.Del(context.Background(), key).Val() > 0
}

func (d *DataBases) ShortVideoIncrCommentCount(userId, FileId string, by int64) bool {
	key := ShortVideoCommentNum + FileId
	return d.rdb.HIncrBy(context.Background(), key, userId, by).Val() > 0
}

func (d *DataBases) ShortVideoIsComment(userId, FileId string) bool {
	key := ShortVideoCommentNum + FileId
	result, _ := d.rdb.HGet(context.Background(), key, userId).Result()
	if result == "" || result == "0" {
		return false
	}
	return true
}

func (d *DataBases) ShortVideoDeleteCommentCountByFileId(FileId string) bool {
	key := ShortVideoCommentNum + FileId
	return d.rdb.Del(context.Background(), key).Val() > 0
}

func (d *DataBases) ShortVideoSetCommentListCache(k string, json string) error {
	key := ShortVideoCommentList + k
	return d.rdb.SetEx(context.Background(), key, json, 2*time.Minute-1).Err()
}

func (d *DataBases) ShortVideoGetCommentListCache(k string) (string, error) {
	key := ShortVideoCommentList + k
	return d.rdb.Get(context.Background(), key).Result()
}

func (d *DataBases) ShortVideoDelCommentListCache(k string) bool {
	keys, err := d.rdb.Keys(context.Background(), ShortVideoCommentList+k).Result()
	if err != nil || len(keys) == 0 {
		return false
	}
	return d.rdb.Del(context.Background(), keys...).Val() > 0
}

func (d *DataBases) SetShortVideoLikeListCache(k string, json string) error {
	key := ShortVideoLikeList + k
	return d.rdb.SetEx(context.Background(), key, json, 2*time.Minute-1).Err()
}

func (d *DataBases) ShortVideoGetLikeListCache(k string) (string, error) {
	key := ShortVideoLikeList + k
	return d.rdb.Get(context.Background(), key).Result()
}

func (d *DataBases) SetShortVideoListByUserIdCache(k string, json string) error {
	key := ShortVideoListByUserId + k
	return d.rdb.SetEx(context.Background(), key, json, 2*time.Minute-1).Err()
}

func (d *DataBases) ShortVideoListByUserIdCache(k string) (string, error) {
	key := ShortVideoListByUserId + k
	return d.rdb.Get(context.Background(), key).Result()
}

func (d *DataBases) ShortVideoDelListByUserIdCache(k string) bool {
	keys, err := d.rdb.Keys(context.Background(), ShortVideoListByUserId+k).Result()
	if err != nil || len(keys) == 0 {
		return false
	}
	return d.rdb.Del(context.Background(), keys...).Val() > 0
}

func (d *DataBases) ShortVideoSetCommentLike(userId string, commentId int64) bool {
	key := ShortVideoCommentLikeSet + strconv.FormatInt(commentId, 10)
	return d.rdb.SAdd(context.Background(), key, userId).Val() > 0
}

func (d *DataBases) ShortVideoCancelCommentLike(userId string, commentId int64) bool {
	key := ShortVideoCommentLikeSet + strconv.FormatInt(commentId, 10)
	return d.rdb.SRem(context.Background(), key, userId).Val() > 0
}

func (d *DataBases) ShortVideoIsCommentLike(userId string, commentId int64) bool {
	key := ShortVideoCommentLikeSet + strconv.FormatInt(commentId, 10)
	return d.rdb.SIsMember(context.Background(), key, userId).Val()
}

func (d *DataBases) AddShortVideoNotice(userId string, noticeId int64) bool {
	key := ShortVideoNotices + userId
	return d.rdb.SAdd(context.Background(), key, utils2.Int64ToString(noticeId)).Val() > 0
}

func (d *DataBases) GetShortVideoNoticeCount(userId string) int64 {
	key := ShortVideoNotices + userId
	return d.rdb.SCard(context.Background(), key).Val()
}

func (d *DataBases) DelShortVideoNotice(userId string, noticeId ...int64) bool {
	key := ShortVideoNotices + userId
	var mem []interface{}
	for _, v := range noticeId {
		mem = append(mem, v)
	}
	return d.rdb.SRem(context.Background(), key, mem...).Val() > 0
}
