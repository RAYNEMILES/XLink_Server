package shortVideo

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbFriend "Open_IM/pkg/proto/friend"
	sdkWs "Open_IM/pkg/proto/sdk_ws"
	pbShortVideo "Open_IM/pkg/proto/short_video"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (s *Server) CreateShortVideo(_ context.Context, request *pbShortVideo.CreateShortVideoRequest) (*pbShortVideo.CreateShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "CreateShortVideo", request)
	response := &pbShortVideo.CreateShortVideoResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// InterestId
	bytes, _ := json.Marshal(request.InterestIds)
	InterestId := string(bytes)

	// file id
	shortVideo, err := im_mysql_model.GetShortVideoByFileId(request.FileId)
	if err == nil && shortVideo.FileId == request.FileId {
		// update
		info := &db.ShortVideo{
			FileId:     request.FileId,
			Status:     constant.ShortVideoTypeNormal,
			Desc:       request.Desc,
			InterestId: InterestId[1 : len(InterestId)-1],
		}

		updateErr := im_mysql_model.AlterShortVideo(info)
		if updateErr != nil {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  constant.ErrDB.ErrMsg,
			}
			return response, nil
		}
		response.FileId = request.FileId
		return response, nil
	}

	insertData := &db.ShortVideo{}
	insertData.FileId = request.FileId
	insertData.Name = request.Name
	insertData.Desc = request.Desc
	insertData.InterestId = InterestId[1 : len(InterestId)-1]
	insertData.ClassId = 0
	insertData.ClassName = "其他"
	insertData.CoverUrl = request.CoverUrl
	insertData.MediaUrl = request.MediaUrl
	insertData.Type = ""
	insertData.Status = constant.ShortVideoTypePrivate
	insertData.Size = 0
	insertData.Height = 0
	insertData.Width = 0
	marshal, _ := json.Marshal(request)
	insertData.Json = string(marshal)

	insertData.UserId = request.UserId
	if insertData.UserId == "" {
		insertData.UserId = config.Config.Manager.AppManagerUid[0]
	}

	insertData.CreateTime = time.Now().Unix()

	dbErr := im_mysql_model.InsertShortVideo(insertData)
	if dbErr != nil {
		log.NewError(request.OperationID, "FileUploadCallBack", "insert short video error", dbErr)
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	go func() {
		// cache
		db.DB.ShortVideoDelListByUserIdCache(request.UserId + "_" + request.UserId + "_*")

		// gorse
		gorseResult := utils2.InsertItem(request.FileId, constant.ShortString, time.Now().String(), strings.Split(insertData.InterestId, ","))
		if gorseResult == false {
			log.NewWarn(request.OperationID, "FileUploadCallBack", "gorseResult", gorseResult, request.FileId)
		}

		// count
		im_mysql_model.IncrShortVideoUserCountByUserId(insertData.UserId, "work_num", 1)

		// notices
		followerList, err := im_mysql_model.GetFansIdByUserId(insertData.UserId)
		if err == nil && len(followerList) > 0 {
			for _, follower := range followerList {
				insertData := db.ShortVideoNotice{
					UserId:       follower,
					SourceUserId: insertData.UserId,
					FileId:       request.FileId,
					Type:         constant.ShortVideoNoticeTypeNewPost,
					Context:      insertData.Desc,
				}
				result := im_mysql_model.AddShortVideoNotice(insertData)
				if result != 0 {
					db.DB.AddShortVideoNotice(follower, result)
				}
			}
		}
	}()

	response.FileId = request.FileId
	return response, nil
}

func (s *Server) GetFollowShortVideoList(ctx context.Context, request *pbShortVideo.GetFollowShortVideoListRequest) (*pbShortVideo.GetFollowShortVideoListResponse, error) {
	log.NewInfo(request.OperationID, "GetFollowShortVideoList", request)
	response := &pbShortVideo.GetFollowShortVideoListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	// mysql
	list, count, err := im_mysql_model.GetFollowShortVideoFileIdListByUserId(request.UserId, request.Pagination.PageNumber, request.Pagination.ShowNumber)

	if err != nil {
		return response, err
	}

	response.FileIdList = list
	response.ShortVideoCount = count

	if len(list) > 0 {
		// Get short video info
		list, _ := s.GetShortVideoByFieldIdList(ctx, &pbShortVideo.GetShortVideoByFileIdListRequest{
			OperationID: request.OperationID,
			UserId:      request.UserId,
			FileIdList:  list,
		})
		response.ShortVideoInfoList = list.ShortVideoInfoList
	}

	return response, nil
}

func (s *Server) GetShortVideoNoticeList(_ context.Context, request *pbShortVideo.GetShortVideoNoticesRequest) (*pbShortVideo.GetShortVideoNoticesResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoNoticeList", request)
	response := &pbShortVideo.GetShortVideoNoticesResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	response.NoticeCount = 0
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	// mysql
	list, count, err := im_mysql_model.GetShortVideoNoticeList(request.UserId, int8(request.NoticeType), int8(request.State), request.Pagination.PageNumber, request.Pagination.ShowNumber)
	if err != nil || count == 0 || len(list) == 0 {
		return response, nil
	}

	response.NoticeCount = count
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	userIds := make([]string, 0)
	noticeIds := make([]int64, 0)
	for _, v := range list {
		userIds = append(userIds, v.SourceUserId)
		noticeIds = append(noticeIds, v.Id)
	}

	updateUserInfo, err := im_mysql_model.GetUserInfoByUserIDs(userIds)
	if err == nil {
		userInfoMap := make(map[string]*pbShortVideo.UserInfoMessage)
		for _, userInfo := range updateUserInfo {
			userInfoMap[userInfo.UserID] = &pbShortVideo.UserInfoMessage{
				UserId:   userInfo.UserID,
				Nickname: userInfo.Nickname,
				FaceURL:  userInfo.FaceURL,
			}
		}

		for _, notice := range list {
			SelfInfo := &pbShortVideo.OperationMessage{
				IsLike:    false,
				IsComment: false,
			}

			// Whether you have already liked the short video
			if notice.Type == constant.ShortVideoNoticeTypeNewPost {
				SelfInfo.IsLike = db.DB.ShortVideoIsLike(request.UserId, notice.FileId)
			}

			// Whether you've already liked the comment
			if notice.Type == constant.ShortVideoNoticeTypeLikeComment || notice.Type == constant.ShortVideoNoticeTypeReplyShort {
				SelfInfo.IsLike = db.DB.ShortVideoIsCommentLike(request.UserId, notice.CommentId)
			}

			response.ShortVideoNoticeList = append(response.ShortVideoNoticeList, &pbShortVideo.ShortVideoNoticeMessage{
				NoticeId:   notice.Id,
				NoticeType: int32(notice.Type),
				FileId:     notice.FileId,
				CommentId:  notice.CommentId,
				Context:    notice.Context,
				State:      int32(notice.State),
				CreateTime: notice.CreateTime,
				UpUserInfo: userInfoMap[notice.SourceUserId],
				SelfInfo:   SelfInfo,
			})
		}
	}

	// read notice
	go func() {
		im_mysql_model.UpdateShortVideoStateByIdList(noticeIds, constant.ShortVideoNoticeStateRead)
		db.DB.DelShortVideoNotice(request.UserId, noticeIds...)
	}()

	return response, nil
}

func (s *Server) SearchShortVideo(ctx context.Context, request *pbShortVideo.SearchShortVideoRequest) (*pbShortVideo.SearchShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "SearchShortVideo", request)
	response := &pbShortVideo.SearchShortVideoResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	searchKey := request.Keyword
	idList := make([]string, 0)
	var err error
	if len(searchKey) == 0 {
		idList, err = im_mysql_model.GetTopLikeShortVideoList()
	} else {
		idList, err = im_mysql_model.SearchShortVideoByKeyword(searchKey)
		if err != nil || len(idList) == 0 {
			return response, nil
		}
	}

	log.NewInfo(request.OperationID, "SearchShortVideo", "idList", idList)

	// Get short video info
	list, err := s.GetShortVideoByFieldIdList(ctx, &pbShortVideo.GetShortVideoByFileIdListRequest{
		OperationID: request.OperationID,
		UserId:      request.UserId,
		FileIdList:  idList,
	})

	response.ShortVideoInfoList = list.ShortVideoInfoList
	return response, nil
}

func (s *Server) GetShortVideoUserCountByUserId(ctx context.Context, request *pbShortVideo.GetShortVideoUserCountByUserIdRequest) (*pbShortVideo.GetShortVideoUserCountByUserIdResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoUserCountByUserId", request)
	response := &pbShortVideo.GetShortVideoUserCountByUserIdResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check user id
	userRpc, err := createUserGRPConnection(request.OperationID)
	if err != nil {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCountRpcFail.ErrCode,
			ErrMsg:  constant.ErrCountRpcFail.ErrMsg,
		}
		return response, nil
	}
	userInfo, err := userRpc.GetUserById(ctx, &pbUser.GetUserByIdReq{
		OperationID: request.OperationID,
		UserId:      request.UserId,
	})
	if err != nil || userInfo.User.UserId == "" || userInfo.User.IsBlock || userInfo.User.Status != 1 {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCountUserIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrCountUserIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	countByUserId, err := im_mysql_model.GetShortVideoUserCountByUserId(request.UserId)
	if err != nil {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCountMysqlFail.ErrCode,
			ErrMsg:  constant.ErrCountMysqlFail.ErrMsg,
		}
		return response, nil
	}

	_ = utils.CopyStructFields(&response, countByUserId)

	// notice
	response.NoticeNum = db.DB.GetShortVideoNoticeCount(request.UserId)

	return response, nil
}

func (s *Server) Follow(ctx context.Context, request *pbShortVideo.FollowRequest) (*pbShortVideo.FollowResponse, error) {
	log.NewInfo(request.OperationID, "FollowUser", request)
	response := &pbShortVideo.FollowResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	if request.UserId == request.FollowUserId {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFollowSelf.ErrCode,
			ErrMsg:  constant.ErrFollowSelf.ErrMsg,
		}
		return response, nil
	}

	// check follow user id
	userRpc, err := createUserGRPConnection(request.OperationID)
	if err != nil {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFollowRpcFail.ErrCode,
			ErrMsg:  constant.ErrFollowRpcFail.ErrMsg,
		}
		return response, nil
	}
	followUserInfo, err := userRpc.GetUserById(ctx, &pbUser.GetUserByIdReq{
		OperationID: request.OperationID,
		UserId:      request.FollowUserId,
	})
	if err != nil || followUserInfo.User.UserId == "" || followUserInfo.User.IsBlock || followUserInfo.User.Status != 1 {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFollowUserIdIsNil.ErrCode,
			ErrMsg:  constant.ErrFollowUserIdIsNil.ErrMsg,
		}
		return response, nil
	}

	// follow list
	follow, _ := s.IsFollow(ctx, &pbShortVideo.IsFollowRequest{
		OperationID:  request.OperationID,
		UserId:       request.UserId,
		FollowUserId: request.FollowUserId,
	})

	if request.Follow == true {
		// follow list
		if follow.IsFollow {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrFollowAlreadyExist.ErrCode,
				ErrMsg:  constant.ErrFollowAlreadyExist.ErrMsg,
			}
			return response, nil
		}

		// mysql
		result := im_mysql_model.InsertShortVideoFollow(request.FollowUserId, request.UserId)
		if !result {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrFollowMysqlFail.ErrCode,
				ErrMsg:  constant.ErrFollowMysqlFail.ErrMsg,
			}
			return response, nil
		}
	}

	if request.Follow == false {
		// follow list
		if follow.IsFollow == false {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrFollowIsNotExist.ErrCode,
				ErrMsg:  constant.ErrFollowIsNotExist.ErrMsg,
			}
			return response, nil
		}

		// mysql
		result := im_mysql_model.DeleteShortVideoFollow(request.FollowUserId, request.UserId)
		if !result {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrFollowMysqlFail.ErrCode,
				ErrMsg:  constant.ErrFollowMysqlFail.ErrMsg,
			}
			return response, nil
		}
	}

	go func() {
		// redis
		if request.Follow == true {
			db.DB.ShortVideoFollowAddMember(request.UserId, request.FollowUserId)
		} else {
			db.DB.ShortVideoFollowRemMember(request.UserId, request.FollowUserId)
		}

		// gorse
		utils2.InsertUser(request.FollowUserId, []string{}, []string{})
		followList, _ := db.DB.ShortVideoGetFollowUserListCache(request.UserId)
		utils2.PatchUser(request.UserId, im_mysql_model.GetStringListByUserId(request.UserId), followList)

		// notice
		if request.Follow == true {
			insertData := db.ShortVideoNotice{
				UserId:       request.FollowUserId,
				SourceUserId: request.UserId,
				Type:         constant.ShortVideoNoticeTypeFollowMe,
			}
			result := im_mysql_model.AddShortVideoNotice(insertData)
			if result != 0 {
				db.DB.AddShortVideoNotice(request.FollowUserId, result)
			}
		}

		// count
		if request.Follow == true {
			im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "follow_num", 1)
			im_mysql_model.IncrShortVideoUserCountByUserId(request.FollowUserId, "fans_num", 1)
		} else {
			im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "follow_num", -1)
			im_mysql_model.IncrShortVideoUserCountByUserId(request.FollowUserId, "fans_num", -1)
		}
	}()

	return response, nil
}

func (s *Server) GetFollowList(_ context.Context, request *pbShortVideo.GetFollowListRequest) (*pbShortVideo.GetFollowListResponse, error) {
	log.NewInfo(request.OperationID, "GetFollowList", request)
	response := &pbShortVideo.GetFollowListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.FollowCount = 0
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	list, count, err := im_mysql_model.GetFollowList("", request.UserId, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	if err != nil || len(list) == 0 || count == 0 {
		return response, nil
	}

	userIds := make([]string, 0)
	for _, v := range list {
		userIds = append(userIds, v.UserId)
	}
	if len(userIds) == 0 {
		return response, nil
	}

	followUserId, err := im_mysql_model.GetUserInfoByUserIDs(userIds)
	if err != nil || len(followUserId) == 0 {
		return response, nil
	}

	userInfoMap := make(map[string]*pbShortVideo.UserInfoMessage)
	for _, userInfo := range followUserId {
		userInfoMap[userInfo.UserID] = &pbShortVideo.UserInfoMessage{
			UserId:    userInfo.UserID,
			Nickname:  userInfo.Nickname,
			FaceURL:   userInfo.FaceURL,
			IsDeleted: false,
		}
	}

	for _, v := range list {
		userInfo := pbShortVideo.UserInfoMessage{
			UserId:    v.UserId,
			Nickname:  v.UserId,
			FaceURL:   "",
			IsDeleted: true,
		}
		if _, ok := userInfoMap[v.UserId]; ok {
			userInfo.UserId = userInfoMap[v.UserId].UserId
			userInfo.Nickname = userInfoMap[v.UserId].Nickname
			userInfo.FaceURL = userInfoMap[v.UserId].FaceURL
			userInfo.IsDeleted = userInfoMap[v.UserId].IsDeleted
		}
		response.FollowList = append(response.FollowList, &userInfo)
	}
	response.FollowCount = count

	return response, nil
}

func (s *Server) GetFansList(_ context.Context, request *pbShortVideo.GetFansListRequest) (*pbShortVideo.GetFansListResponse, error) {
	log.NewInfo(request.OperationID, "GetFansList", request)
	response := &pbShortVideo.GetFansListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.FansCount = 0
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	list, count, err := im_mysql_model.GetFollowList(request.UserId, "", request.Pagination.PageNumber, request.Pagination.ShowNumber)
	if err != nil || len(list) == 0 || count == 0 {
		return response, nil
	}

	userIds := make([]string, 0)
	for _, v := range list {
		userIds = append(userIds, v.FansId)
	}
	if len(userIds) == 0 {
		return response, nil
	}

	fansUserInfo, err := im_mysql_model.GetUserInfoByUserIDs(userIds)
	if err != nil || len(fansUserInfo) == 0 {
		return response, nil
	}

	userInfoMap := make(map[string]*pbShortVideo.UserInfoMessage)
	for _, userInfo := range fansUserInfo {
		userInfoMap[userInfo.UserID] = &pbShortVideo.UserInfoMessage{
			UserId:    userInfo.UserID,
			Nickname:  userInfo.Nickname,
			FaceURL:   userInfo.FaceURL,
			IsDeleted: false,
		}
	}

	for _, v := range list {
		userInfo := pbShortVideo.UserInfoMessage{
			UserId:    v.UserId,
			Nickname:  v.UserId,
			FaceURL:   "",
			IsDeleted: true,
		}
		if _, ok := userInfoMap[v.FansId]; ok {
			userInfo.UserId = userInfoMap[v.FansId].UserId
			userInfo.Nickname = userInfoMap[v.FansId].Nickname
			userInfo.FaceURL = userInfoMap[v.FansId].FaceURL
			userInfo.IsDeleted = userInfoMap[v.FansId].IsDeleted
		}
		response.FansList = append(response.FansList, &userInfo)
	}
	response.FansCount = count

	return response, nil
}

func (s *Server) IsFollow(_ context.Context, request *pbShortVideo.IsFollowRequest) (*pbShortVideo.IsFollowResponse, error) {
	log.NewInfo(request.OperationID, "IsFollow", request)
	response := &pbShortVideo.IsFollowResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.IsFollow = false

	isFollow, err := db.DB.ShortVideoIsFollow(request.UserId, request.FollowUserId)
	if err == nil {
		response.IsFollow = isFollow
		return response, nil
	}

	// mysql
	idList, err := im_mysql_model.GetShortVideoFollowUserIdByFansId(request.UserId)
	if err != nil {
		return response, nil
	}

	followCount := len(idList)
	// count
	im_mysql_model.UpdateShortVideoUserCountByUserId(request.UserId, "follow_num", int64(followCount))

	if followCount == 0 {
		goto returnResponse
	}

	// cache
	_ = db.DB.ShortVideoSetFollowUserListCache(request.UserId, idList)

	// check
	response.IsFollow = utils.IsContain(request.FollowUserId, utils.InterfaceArrayToStringArray(idList))

returnResponse:
	return response, nil
}

func (s *Server) ShortVideoComment(ctx context.Context, request *pbShortVideo.ShortVideoCommentRequest) (*pbShortVideo.ShortVideoCommentResponse, error) {
	log.NewInfo(request.OperationID, "ShortVideoComment", request)
	response := &pbShortVideo.ShortVideoCommentResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check comment content
	if utils.FindSensitiveWord(request.Content) {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCommentContentIsSensitive.ErrCode,
			ErrMsg:  constant.ErrCommentContentIsSensitive.ErrMsg,
		}
		return response, nil
	}

	// check file id
	shortVideoInfo, err := s.GetShortVideoByFieldId(ctx, &pbShortVideo.GetShortVideoByFileIdRequest{
		OperationID: request.OperationID,
		FileId:      request.FileId,
		UserId:      request.UserId,
		IsAdmin:     false,
	})
	if err != nil || shortVideoInfo.CommonResp.ErrCode != constant.OK.ErrCode {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// check parent comment id
	parentInfo := &db.ShortVideoComment{}
	var level int64 = 0
	var replyTo = ""
	var level1CommentId int64 = 0
	var level0CommentId int64 = 0
	if request.ParentId != 0 {
		parentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.ParentId)
		if err != nil || parentInfo == nil || parentInfo.FileId != request.FileId || parentInfo.Status != constant.ShortVideoCommentStatusNormal {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
				ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
			}
			return response, nil
		}
		level = parentInfo.LevelId + 1
		replyTo = parentInfo.UserID

		// get level 0 comment id
		if level == 1 {
			level0CommentId = parentInfo.CommentId
		} else {
			level0CommentId = parentInfo.Level0CommentId
		}

		// get level 1 comment id
		if level == 2 {
			level1CommentId = parentInfo.CommentId
		} else {
			level1CommentId = parentInfo.Level1CommentId
		}
	}

	// mysql
	insertData := &db.ShortVideoComment{}
	insertData.ParentId = request.ParentId
	insertData.Level1CommentId = level1CommentId
	insertData.Level0CommentId = level0CommentId
	insertData.UserID = request.UserId
	insertData.FileId = request.FileId
	insertData.Content = request.Content
	insertData.LevelId = level
	insertData.ReplyTo = replyTo
	commentId, err := im_mysql_model.InsertShortVideoComment(insertData)
	if err != nil || commentId == 0 {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCommentMysqlFail.ErrCode,
			ErrMsg:  constant.ErrCommentMysqlFail.ErrMsg,
		}
		return response, nil
	}
	response.CommentId = commentId

	// redis
	db.DB.ShortVideoIncrCommentCount(request.UserId, request.FileId, 1)

	go func() {
		// gorse
		utils2.InsertFeedback(constant.PositiveFeedbackComment, request.UserId, request.FileId)

		// count
		im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "comment_num", 1)

		// clear cache
		cacheKey := request.FileId + "*"
		db.DB.ShortVideoDelCommentListCache(cacheKey)

		// notice
		if request.UserId != shortVideoInfo.ShortVideoInfo.UserId {
			insertData := db.ShortVideoNotice{
				UserId:       shortVideoInfo.ShortVideoInfo.UserId,
				SourceUserId: request.UserId,
				FileId:       request.FileId,
				CommentId:    commentId,
				Type:         constant.ShortVideoNoticeTypeReplyShort,
				Context:      request.Content,
			}
			result := im_mysql_model.AddShortVideoNotice(insertData)
			if result != 0 {
				db.DB.AddShortVideoNotice(shortVideoInfo.ShortVideoInfo.UserId, result)
			}
		}
		if request.ParentId != 0 && request.UserId != parentInfo.UserID {
			insertData := db.ShortVideoNotice{
				UserId:       parentInfo.UserID,
				SourceUserId: request.UserId,
				FileId:       request.FileId,
				CommentId:    commentId,
				Type:         constant.ShortVideoNoticeTypeReplyComment,
				Context:      request.Content,
			}
			result := im_mysql_model.AddShortVideoNotice(insertData)
			if result != 0 {
				db.DB.AddShortVideoNotice(parentInfo.UserID, result)
			}
		}
	}()

	return response, nil
}

func (s *Server) DeleteShortVideoComment(ctx context.Context, request *pbShortVideo.DeleteShortVideoRequest) (*pbShortVideo.DeleteShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "DeleteShortVideo", request)
	response := &pbShortVideo.DeleteShortVideoResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check file id
	shortVideoInfo, err := s.GetShortVideoByFieldId(ctx, &pbShortVideo.GetShortVideoByFileIdRequest{
		OperationID: request.OperationID,
		FileId:      request.FileId,
		UserId:      request.UserId,
		IsAdmin:     false,
	})
	if err != nil || shortVideoInfo.CommonResp.ErrCode != constant.OK.ErrCode {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// check comment id
	commentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.CommentId)
	if err != nil || commentInfo == nil || commentInfo.FileId != request.FileId || commentInfo.Status != constant.ShortVideoCommentStatusNormal {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCommentIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrCommentIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// authority
	if commentInfo.UserID != request.UserId && !request.IsAdmin {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrAuthority.ErrCode,
			ErrMsg:  constant.ErrAuthority.ErrMsg,
		}
		return response, nil
	}

	// all child id
	allCommentInfoList := make([]*db.ShortVideoComment, 0)
	allCommentInfoList = append(allCommentInfoList, commentInfo)
	childCommentInfoList, _ := im_mysql_model.GetAllReplyCommentIdByCommentId(request.CommentId)
	if len(childCommentInfoList) > 0 {
		childCommentInfoList, _ := im_mysql_model.GetCommentInfoListByCommentId(childCommentInfoList)
		for _, childCommentInfo := range childCommentInfoList {
			if childCommentInfo.Status == constant.ShortVideoCommentStatusNormal || childCommentInfo.Status == constant.ShortVideoCommentStatusAudit {
				allCommentInfoList = append(allCommentInfoList, childCommentInfo)
			}
		}
	}

	// mysql
	err = im_mysql_model.DeleteShortVideoCommentByCommentList(allCommentInfoList)
	if err != nil {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCommentMysqlFail.ErrCode,
			ErrMsg:  constant.ErrCommentMysqlFail.ErrMsg,
		}
		return response, nil
	}

	go func() {
		// count
		for _, commentInfo := range allCommentInfoList {
			im_mysql_model.IncrShortVideoUserCountByUserId(commentInfo.UserID, "comment_num", -1)
		}

		// clear cache
		cacheKey := request.FileId + "*"
		db.DB.ShortVideoDelCommentListCache(cacheKey)
	}()

	return response, nil
}

func (s *Server) ShortVideoCommentLike(_ context.Context, request *pbShortVideo.ShortVideoCommentLikeRequest) (*pbShortVideo.ShortVideoCommentLikeResponse, error) {
	log.NewInfo(request.OperationID, "ShortVideoCommentLike", request)
	response := &pbShortVideo.ShortVideoCommentLikeResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check comment id
	parentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.CommentId)
	if err != nil || parentInfo == nil || parentInfo.Status != constant.ShortVideoCommentStatusNormal {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// like
	if request.Like == true {
		// redis
		if db.DB.ShortVideoSetCommentLike(request.UserId, request.CommentId) == false {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeRedisFail.ErrCode,
				ErrMsg:  constant.ErrLikeRedisFail.ErrMsg,
			}
			return response, nil
		}

		// mysql
		if im_mysql_model.InsertShortVideoCommentLike(request.FileId, request.UserId, request.CommentId) == false {
			db.DB.ShortVideoCancelCommentLike(request.UserId, request.CommentId)
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeMysqlFail.ErrCode,
				ErrMsg:  constant.ErrLikeMysqlFail.ErrMsg,
			}
			return response, nil
		}

		// gorse
		go utils2.InsertFeedback(constant.PositiveFeedbackLike, request.UserId, request.FileId)
	}

	// unlike
	if request.Like == false {
		// redis
		if db.DB.ShortVideoCancelCommentLike(request.UserId, request.CommentId) == false {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeRedisFail.ErrCode,
				ErrMsg:  constant.ErrLikeRedisFail.ErrMsg,
			}
			return response, nil
		}

		// mysql
		if im_mysql_model.DeleteShortVideoCommentLike(request.FileId, request.UserId, request.CommentId) == false {
			db.DB.ShortVideoSetLike(request.UserId, request.FileId)
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeMysqlFail.ErrCode,
				ErrMsg:  constant.ErrLikeMysqlFail.ErrMsg,
			}
			return response, nil
		}
	}

	go func() {
		// count
		if request.Like == true {
			im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "comment_like_num", 1)
		} else {
			im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "comment_like_num", -1)
		}

		// notice
		if request.Like == true && parentInfo.UserID != request.UserId {
			insertData := db.ShortVideoNotice{
				UserId:       parentInfo.UserID,
				SourceUserId: request.UserId,
				FileId:       request.FileId,
				CommentId:    request.CommentId,
				Context:      parentInfo.Content,
				Type:         constant.ShortVideoNoticeTypeLikeComment,
			}
			result := im_mysql_model.AddShortVideoNotice(insertData)
			if result != 0 {
				db.DB.AddShortVideoNotice(parentInfo.UserID, result)
			}
		}

		// clear cache
		cacheKey := request.FileId + "*"
		db.DB.ShortVideoDelCommentListCache(cacheKey)
	}()

	return response, nil
}

func (s *Server) GetShortVideoCommentList(ctx context.Context, request *pbShortVideo.GetShortVideoCommentListRequest) (*pbShortVideo.GetShortVideoCommentListResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoCommentList", request)
	response := &pbShortVideo.GetShortVideoCommentListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// cache
	cacheKey := request.FileId + "_" + strconv.FormatInt(request.ParentId, 10) +
		"_" + strconv.FormatInt(int64(request.Pagination.PageNumber), 10) +
		"_" + strconv.FormatInt(int64(request.Pagination.ShowNumber), 10) +
		"_" + strconv.Itoa(int(request.OrderBy)) + "_" + strconv.Itoa(int(request.SourceCommentId))
	cache, err := db.DB.ShortVideoGetCommentListCache(cacheKey)
	if err == nil {
		_ = json.Unmarshal([]byte(cache), response)
		return response, nil
	}

	// check file id
	response.CommentCount = 0
	var fileCreator = ""
	if request.FileId != "" {
		shortVideoInfo, err := s.GetShortVideoByFieldId(ctx, &pbShortVideo.GetShortVideoByFileIdRequest{
			OperationID: request.OperationID,
			FileId:      request.FileId,
			UserId:      request.UserId,
		})

		if err != nil || shortVideoInfo.CommonResp.ErrCode != constant.OK.ErrCode {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
				ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
			}
			return response, nil
		}
		response.CommentCount = shortVideoInfo.ShortVideoInfo.CommentNum
		fileCreator = shortVideoInfo.ShortVideoInfo.UserId
	}

	// check comment id
	if request.ParentId != 0 {
		parentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.ParentId)
		if err != nil || parentInfo == nil || parentInfo.FileId != request.FileId || parentInfo.Status != constant.ShortVideoCommentStatusNormal {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
				ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
			}
			return response, nil
		}
	}

	// mysql
	commentList, count, err := im_mysql_model.GetShortVideoCommentList(constant.ShortVideoCommentStatusNormal, request.FileId, request.UserId, request.CommentId, request.ParentId, -1, 0, request.Pagination.PageNumber, request.Pagination.ShowNumber, request.OrderBy, request.SourceCommentId)
	if err != nil {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrCommentMysqlFail.ErrCode,
			ErrMsg:  constant.ErrCommentMysqlFail.ErrMsg,
		}
		return nil, err
	}

	response.Level0CommentCount = count
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	if len(commentList) == 0 {
		return response, nil
	}

	userIds := make([]string, 0)
	for _, comment := range commentList {
		commentInfo := &pbShortVideo.ShortVideoCommentInfo{
			CommentId:         comment.CommentId,
			FileId:            comment.FileId,
			Content:           comment.Content,
			ParentId:          comment.ParentId,
			CommentLikeCount:  comment.CommentLikeCount,
			CommentReplyCount: comment.CommentReplyCount,
			CreateTime:        comment.CreateTime,
			Status:            int32(comment.Status),
			SelfOperation: &pbShortVideo.OperationMessage{
				IsLike: db.DB.ShortVideoIsCommentLike(request.OperationUserID, comment.CommentId),
			},
			CommentUserInfo: &pbShortVideo.UserInfoMessage{
				UserId:    comment.UserID,
				IsDeleted: true,
			},
			UpOperation: &pbShortVideo.OperationMessage{
				IsLike: db.DB.ShortVideoIsCommentLike(fileCreator, comment.CommentId),
			},
			ReplyComment: make([]*pbShortVideo.ShortVideoCommentInfo, 0),
		}

		// reply comment
		if comment.ParentId == 0 && comment.CommentReplyCount > 0 {
			replyComment, count, err := im_mysql_model.GetShortVideoCommentList(constant.ShortVideoCommentStatusNormal, "", "", 0, comment.CommentId, -1, 0, 1, 1, 2)
			if err == nil || count > 0 {
				commentInfo.CommentReplyCount = comment.TotalReplyCount
				commentInfo.ReplyComment = append(commentInfo.ReplyComment, &pbShortVideo.ShortVideoCommentInfo{
					CommentId:         replyComment[0].CommentId,
					FileId:            replyComment[0].FileId,
					Content:           replyComment[0].Content,
					ParentId:          replyComment[0].ParentId,
					CommentLikeCount:  replyComment[0].CommentLikeCount,
					CommentReplyCount: replyComment[0].CommentReplyCount,
					CreateTime:        replyComment[0].CreateTime,
					Status:            int32(replyComment[0].Status),
					SelfOperation: &pbShortVideo.OperationMessage{
						IsLike: db.DB.ShortVideoIsCommentLike(request.OperationUserID, replyComment[0].CommentId),
					},
					CommentUserInfo: &pbShortVideo.UserInfoMessage{
						UserId:    replyComment[0].UserID,
						IsDeleted: true,
					},
					UpOperation: &pbShortVideo.OperationMessage{
						IsLike: db.DB.ShortVideoIsCommentLike(fileCreator, replyComment[0].CommentId),
					},
				})
				userIds = append(userIds, replyComment[0].UserID)
			}
		}

		response.CommentList = append(response.CommentList, commentInfo)
		userIds = append(userIds, comment.UserID)
	}

	// user info
	userIds = utils.RemoveDuplicatesAndEmpty(userIds)
	if len(userIds) > 0 {
		connection, err := createUserGRPConnection(request.OperationID)
		if err != nil {
			return response, nil
		}

		rpcRequest := &pbUser.GetUserInfoReq{
			UserIDList: userIds,
		}
		userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
		if err != nil {
			return response, nil
		}

		userInfoMap := make(map[string]*sdkWs.UserInfo)
		for _, user := range userInfo.UserInfoList {
			userInfoMap[user.UserID] = user
		}

		for i, comment := range response.CommentList {
			// comment up user info
			if comment.CommentUserInfo.UserId != "" && userInfoMap[comment.CommentUserInfo.UserId] != nil {
				response.CommentList[i].CommentUserInfo.Nickname = userInfoMap[comment.CommentUserInfo.UserId].Nickname
				response.CommentList[i].CommentUserInfo.FaceURL = userInfoMap[comment.CommentUserInfo.UserId].FaceURL
				response.CommentList[i].CommentUserInfo.IsDeleted = false
			}

			if comment.ParentId == 0 && comment.CommentReplyCount > 0 {
				// reply comment up user info
				if comment.ReplyComment[0].CommentUserInfo.UserId != "" && userInfoMap[comment.ReplyComment[0].CommentUserInfo.UserId] != nil {
					response.CommentList[i].ReplyComment[0].CommentUserInfo.Nickname = userInfoMap[comment.ReplyComment[0].CommentUserInfo.UserId].Nickname
					response.CommentList[i].ReplyComment[0].CommentUserInfo.FaceURL = userInfoMap[comment.ReplyComment[0].CommentUserInfo.UserId].FaceURL
					response.CommentList[i].ReplyComment[0].CommentUserInfo.IsDeleted = false
				}
			}
		}
	}

	// redis
	marshal, _ := json.Marshal(response)
	_ = db.DB.ShortVideoSetCommentListCache(cacheKey, string(marshal))

	return response, nil
}

func (s *Server) GetCommentPage(ctx context.Context, request *pbShortVideo.GetCommentPageRequest) (*pbShortVideo.GetCommentPageResponse, error) {
	log.NewInfo(request.OperationID, "GetCommentPage", request)
	response := &pbShortVideo.GetCommentPageResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.ReplyCount = 0
	response.TotalReplyNum = 0
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	var level1CommentId int64
	var level2CommentId int64
	if request.SourceCommentId != 0 {
		sourceCommentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.SourceCommentId)
		if err != nil || sourceCommentInfo == nil || sourceCommentInfo.Status != constant.ShortVideoCommentStatusNormal {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
				ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
			}
			return response, nil
		}

		request.CommentId = sourceCommentInfo.Level0CommentId
		if sourceCommentInfo.LevelId == 0 {
			request.CommentId = sourceCommentInfo.CommentId
		}

		if sourceCommentInfo.LevelId == 1 {
			level1CommentId = sourceCommentInfo.CommentId
		} else {
			level2CommentId = sourceCommentInfo.CommentId
		}
	}

	// check comment id
	commentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.CommentId)
	if err != nil || commentInfo == nil || commentInfo.Status != constant.ShortVideoCommentStatusNormal {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	if commentInfo.LevelId != 0 {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// get short video info
	shortVideoInfo, err := im_mysql_model.GetShortVideoByFileId(commentInfo.FileId)
	response.CreatorUserId = shortVideoInfo.UserId

	// comment info
	userIds := make([]string, 0)
	response.CommentInfo = &pbShortVideo.ShortVideoCommentInfo{
		CommentId:         commentInfo.CommentId,
		FileId:            commentInfo.FileId,
		Content:           commentInfo.Content,
		ParentId:          commentInfo.ParentId,
		CommentLikeCount:  commentInfo.CommentLikeCount,
		CommentReplyCount: commentInfo.CommentReplyCount,
		CreateTime:        commentInfo.CreateTime,
		Status:            int32(commentInfo.Status),
		SelfOperation: &pbShortVideo.OperationMessage{
			IsLike: db.DB.ShortVideoIsCommentLike(request.UserId, commentInfo.CommentId),
		},
		UpOperation: &pbShortVideo.OperationMessage{
			IsLike: db.DB.ShortVideoIsCommentLike(shortVideoInfo.UserId, commentInfo.CommentId),
		},
		CommentUserInfo: &pbShortVideo.UserInfoMessage{
			UserId:    commentInfo.UserID,
			IsDeleted: true,
		},
	}
	userIds = append(userIds, commentInfo.UserID)

	// comment state
	if commentInfo.Status != constant.ShortVideoCommentStatusNormal {
		response.CommentInfo.CommentUserInfo.UserId = ""
		response.CommentInfo.Content = ""
	}

	// reply level 1 comment
	level1Comment, level1Count, err := im_mysql_model.GetShortVideoCommentList(constant.ShortVideoCommentStatusNormal, "", "", 0, request.CommentId, -1, 0, request.Pagination.PageNumber, request.Pagination.ShowNumber, 1, level1CommentId)
	if err == nil && level1Count > 0 {
		response.ReplyCount = level1Count
		response.TotalReplyNum = im_mysql_model.GetCommentCountByLevel0CommentId(request.CommentId)
		if len(level1Comment) > 0 {
			for _, comment := range level1Comment {
				userIds = append(userIds, comment.UserID)
				commentInfo := &pbShortVideo.ShortVideoCommentInfo{
					CommentId:         comment.CommentId,
					FileId:            comment.FileId,
					Content:           comment.Content,
					ParentId:          comment.ParentId,
					CommentLikeCount:  comment.CommentLikeCount,
					CommentReplyCount: comment.CommentReplyCount,
					CreateTime:        comment.CreateTime,
					Status:            int32(comment.Status),
					SelfOperation: &pbShortVideo.OperationMessage{
						IsLike: db.DB.ShortVideoIsCommentLike(request.UserId, comment.CommentId),
					},
					CommentUserInfo: &pbShortVideo.UserInfoMessage{
						UserId:    comment.UserID,
						IsDeleted: true,
					},
					UpOperation: &pbShortVideo.OperationMessage{
						IsLike: db.DB.ShortVideoIsCommentLike(shortVideoInfo.UserId, comment.CommentId),
					},
					ReplyComment: make([]*pbShortVideo.ShortVideoCommentInfo, 0),
				}
				commentInfo.CommentReplyCount = im_mysql_model.GetCommentCountByLevel1CommentId(comment.CommentId)

				response.ReplyList = append(response.ReplyList, commentInfo)

				// get level 2 comment
				if comment.CommentReplyCount > 0 {
					level2Comment, level2Count, err := im_mysql_model.GetShortVideoCommentList(constant.ShortVideoCommentStatusNormal, "", "", 0, -1, -1, comment.CommentId, 1, 3, 1, level2CommentId)
					if err == nil && level2Count > 0 {
						for _, replyComment := range level2Comment {
							userIds = append(userIds, replyComment.UserID)
							userIds = append(userIds, replyComment.ReplyTo)
							response.ReplyList[len(response.ReplyList)-1].ReplyComment = append(response.ReplyList[len(response.ReplyList)-1].ReplyComment, &pbShortVideo.ShortVideoCommentInfo{
								CommentId:         replyComment.CommentId,
								FileId:            replyComment.FileId,
								Content:           replyComment.Content,
								ParentId:          replyComment.ParentId,
								CommentLikeCount:  replyComment.CommentLikeCount,
								CommentReplyCount: replyComment.CommentReplyCount,
								CreateTime:        replyComment.CreateTime,
								Status:            int32(replyComment.Status),
								SelfOperation: &pbShortVideo.OperationMessage{
									IsLike: db.DB.ShortVideoIsCommentLike(request.UserId, replyComment.CommentId),
								},
								UpOperation: &pbShortVideo.OperationMessage{
									IsLike: db.DB.ShortVideoIsCommentLike(shortVideoInfo.UserId, replyComment.CommentId),
								},
								CommentUserInfo: &pbShortVideo.UserInfoMessage{
									UserId:    replyComment.UserID,
									IsDeleted: true,
								},
								ReplyToUserInfo: &pbShortVideo.UserInfoMessage{
									UserId:    replyComment.ReplyTo,
									IsDeleted: true,
								},
							})
						}
					}
				}
			}
		}
	}

	// user info
	userIds = utils.RemoveDuplicatesAndEmpty(userIds)
	if len(userIds) > 0 {
		connection, err := createUserGRPConnection(request.OperationID)
		if err != nil {
			return response, nil
		}

		rpcRequest := &pbUser.GetUserInfoReq{
			UserIDList: userIds,
		}
		userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
		if err != nil {
			return response, nil
		}
		userInfoMap := make(map[string]*sdkWs.UserInfo)
		for _, user := range userInfo.UserInfoList {
			userInfoMap[user.UserID] = user
		}

		// comment up user info
		if response.CommentInfo.CommentUserInfo.UserId != "" && userInfoMap[response.CommentInfo.CommentUserInfo.UserId] != nil {
			response.CommentInfo.CommentUserInfo.Nickname = userInfoMap[response.CommentInfo.CommentUserInfo.UserId].Nickname
			response.CommentInfo.CommentUserInfo.FaceURL = userInfoMap[response.CommentInfo.CommentUserInfo.UserId].FaceURL
			response.CommentInfo.CommentUserInfo.IsDeleted = false
		}

		for i, reply := range response.ReplyList {
			// level 1 comment user info
			response.ReplyList[i].CommentUserInfo.UserId = reply.CommentUserInfo.UserId
			response.ReplyList[i].CommentUserInfo.Nickname = reply.CommentUserInfo.UserId
			if reply.CommentUserInfo.UserId != "" && userInfoMap[reply.CommentUserInfo.UserId] != nil {
				response.ReplyList[i].CommentUserInfo.Nickname = userInfoMap[reply.CommentUserInfo.UserId].Nickname
				response.ReplyList[i].CommentUserInfo.FaceURL = userInfoMap[reply.CommentUserInfo.UserId].FaceURL
				response.ReplyList[i].CommentUserInfo.IsDeleted = false
			}
			// level 2 comment user info
			if len(reply.ReplyComment) > 0 {
				for j, replyComment := range reply.ReplyComment {
					// comment user info
					response.ReplyList[i].ReplyComment[j].CommentUserInfo.UserId = replyComment.CommentUserInfo.UserId
					response.ReplyList[i].ReplyComment[j].CommentUserInfo.Nickname = replyComment.CommentUserInfo.UserId
					if replyComment.CommentUserInfo.UserId != "" && userInfoMap[replyComment.CommentUserInfo.UserId] != nil {
						response.ReplyList[i].ReplyComment[j].CommentUserInfo.Nickname = userInfoMap[replyComment.CommentUserInfo.UserId].Nickname
						response.ReplyList[i].ReplyComment[j].CommentUserInfo.FaceURL = userInfoMap[replyComment.CommentUserInfo.UserId].FaceURL
						response.ReplyList[i].ReplyComment[j].CommentUserInfo.IsDeleted = false
					}
					// reply to user info
					response.ReplyList[i].ReplyComment[j].ReplyToUserInfo.UserId = replyComment.ReplyToUserInfo.UserId
					response.ReplyList[i].ReplyComment[j].ReplyToUserInfo.Nickname = replyComment.ReplyToUserInfo.UserId
					if replyComment.ReplyToUserInfo.UserId != "" && userInfoMap[replyComment.ReplyToUserInfo.UserId] != nil {
						response.ReplyList[i].ReplyComment[j].ReplyToUserInfo.Nickname = userInfoMap[replyComment.ReplyToUserInfo.UserId].Nickname
						response.ReplyList[i].ReplyComment[j].ReplyToUserInfo.FaceURL = userInfoMap[replyComment.ReplyToUserInfo.UserId].FaceURL
						response.ReplyList[i].ReplyComment[j].ReplyToUserInfo.IsDeleted = false
					}
				}
			}
		}
	}

	return response, nil
}

func (s *Server) GetCommentPageReplyList(ctx context.Context, request *pbShortVideo.GetCommentPageReplyListRequest) (*pbShortVideo.GetCommentPageReplyListResponse, error) {
	log.NewInfo(request.OperationID, "GetCommentPage", request)
	response := &pbShortVideo.GetCommentPageReplyListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.ReplyCount = 0
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	// check comment id
	commentInfo, err := im_mysql_model.GetShortVideoCommentByCommentId(request.CommentId)
	if err != nil || commentInfo == nil || commentInfo.Status != constant.ShortVideoCommentStatusNormal {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrParentIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrParentIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// get short video info
	shortVideoInfo, _ := im_mysql_model.GetShortVideoByFileId(commentInfo.FileId)

	// get reply list
	replyComment, replyCount, err := im_mysql_model.GetShortVideoCommentList(constant.ShortVideoCommentStatusNormal, "", "", 0, -1, -1, request.CommentId, request.Pagination.PageNumber, request.Pagination.ShowNumber, 1, request.SourceCommentId)
	if err != nil || replyCount <= 0 {
		return response, nil
	}
	response.ReplyCount = replyCount
	if len(replyComment) > 0 {
		userIds := make([]string, 0)
		for _, reply := range replyComment {
			userIds = append(userIds, reply.UserID)
			userIds = append(userIds, reply.ReplyTo)
			response.ReplyList = append(response.ReplyList, &pbShortVideo.ShortVideoCommentInfo{
				CommentId:         reply.CommentId,
				FileId:            reply.FileId,
				Content:           reply.Content,
				ParentId:          reply.ParentId,
				CommentLikeCount:  reply.CommentLikeCount,
				CommentReplyCount: reply.CommentReplyCount,
				CreateTime:        reply.CreateTime,
				Status:            int32(reply.Status),
				SelfOperation: &pbShortVideo.OperationMessage{
					IsLike: db.DB.ShortVideoIsCommentLike(request.UserId, reply.CommentId),
				},
				UpOperation: &pbShortVideo.OperationMessage{
					IsLike: db.DB.ShortVideoIsCommentLike(shortVideoInfo.UserId, reply.CommentId),
				},
				CommentUserInfo: &pbShortVideo.UserInfoMessage{
					UserId:    reply.UserID,
					IsDeleted: true,
				},
				ReplyToUserInfo: &pbShortVideo.UserInfoMessage{
					UserId:    reply.ReplyTo,
					IsDeleted: true,
				},
			})
		}

		if len(userIds) > 0 {
			connection, err := createUserGRPConnection(request.OperationID)
			if err != nil {
				return response, nil
			}

			rpcRequest := &pbUser.GetUserInfoReq{
				UserIDList: userIds,
			}
			userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
			if err != nil {
				return response, nil
			}
			userInfoMap := make(map[string]*sdkWs.UserInfo)
			for _, user := range userInfo.UserInfoList {
				userInfoMap[user.UserID] = user
			}

			for i, reply := range response.ReplyList {
				// comment user info
				response.ReplyList[i].CommentUserInfo.UserId = reply.CommentUserInfo.UserId
				response.ReplyList[i].CommentUserInfo.Nickname = reply.CommentUserInfo.UserId
				if reply.CommentUserInfo.UserId != "" && userInfoMap[reply.CommentUserInfo.UserId] != nil {
					response.ReplyList[i].CommentUserInfo.Nickname = userInfoMap[reply.CommentUserInfo.UserId].Nickname
					response.ReplyList[i].CommentUserInfo.FaceURL = userInfoMap[reply.CommentUserInfo.UserId].FaceURL
					response.ReplyList[i].CommentUserInfo.IsDeleted = false
				}

				// reply to user info
				response.ReplyList[i].ReplyToUserInfo.UserId = reply.ReplyToUserInfo.UserId
				response.ReplyList[i].ReplyToUserInfo.Nickname = reply.ReplyToUserInfo.UserId
				if reply.ReplyToUserInfo.UserId != "" && userInfoMap[reply.ReplyToUserInfo.UserId] != nil {
					response.ReplyList[i].ReplyToUserInfo.Nickname = userInfoMap[reply.ReplyToUserInfo.UserId].Nickname
					response.ReplyList[i].ReplyToUserInfo.FaceURL = userInfoMap[reply.ReplyToUserInfo.UserId].FaceURL
					response.ReplyList[i].ReplyToUserInfo.IsDeleted = false
				}
			}
		}
	}

	return response, nil
}

func (s *Server) ShortVideoLike(ctx context.Context, request *pbShortVideo.ShortVideoLikeRequest) (*pbShortVideo.ShortVideoLikeResponse, error) {
	log.NewInfo(request.OperationID, "ShortVideoLike", request)
	response := &pbShortVideo.ShortVideoLikeResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	shortVideoInfo, err := s.GetShortVideoByFieldId(ctx, &pbShortVideo.GetShortVideoByFileIdRequest{
		OperationID: request.OperationID,
		FileId:      request.FileId,
		UserId:      request.UserId,
		IsAdmin:     false,
	})

	if err != nil || shortVideoInfo.CommonResp.ErrCode != constant.OK.ErrCode {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// like
	if request.Like == true {
		// redis
		if db.DB.ShortVideoSetLike(request.UserId, request.FileId) == false {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeRedisFail.ErrCode,
				ErrMsg:  constant.ErrLikeRedisFail.ErrMsg,
			}
			return response, nil
		}

		// mysql
		if im_mysql_model.InsertShortVideoLike(request.FileId, request.UserId) == false {
			db.DB.ShortVideoCancelLike(request.UserId, request.FileId)
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeMysqlFail.ErrCode,
				ErrMsg:  constant.ErrLikeMysqlFail.ErrMsg,
			}
			return response, nil
		}

		// gorse
		go utils2.InsertFeedback(constant.PositiveFeedbackLike, request.UserId, request.FileId)
	}

	// unlike
	if request.Like == false {
		// redis
		if db.DB.ShortVideoCancelLike(request.UserId, request.FileId) == false {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeRedisFail.ErrCode,
				ErrMsg:  constant.ErrLikeRedisFail.ErrMsg,
			}
			return response, nil
		}
		// mysql
		if im_mysql_model.DeleteShortVideoLike(request.FileId, request.UserId) == false {
			db.DB.ShortVideoSetLike(request.UserId, request.FileId)
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrLikeMysqlFail.ErrCode,
				ErrMsg:  constant.ErrLikeMysqlFail.ErrMsg,
			}
			return response, nil
		}
	}

	go func() {
		// count
		if request.Like == true {
			im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "like_num", 1)
			im_mysql_model.IncrShortVideoUserCountByUserId(shortVideoInfo.ShortVideoInfo.UserId, "harvested_likes_number", 1)
		} else {
			im_mysql_model.IncrShortVideoUserCountByUserId(request.UserId, "like_num", -1)
			im_mysql_model.IncrShortVideoUserCountByUserId(shortVideoInfo.ShortVideoInfo.UserId, "harvested_likes_number", -1)
		}

		// notice
		if request.Like == true && request.UserId != shortVideoInfo.ShortVideoInfo.UserId {
			insertData := db.ShortVideoNotice{
				UserId:       shortVideoInfo.ShortVideoInfo.UserId,
				SourceUserId: request.UserId,
				FileId:       request.FileId,
				Context:      shortVideoInfo.ShortVideoInfo.Desc,
				Type:         constant.ShortVideoNoticeTypeLikeShort,
			}
			result := im_mysql_model.AddShortVideoNotice(insertData)
			if result != 0 {
				db.DB.AddShortVideoNotice(shortVideoInfo.ShortVideoInfo.UserId, result)
			}
		}
	}()

	return response, nil
}

func (s *Server) GetLikeShortVideoList(ctx context.Context, request *pbShortVideo.GetLikeShortVideoListRequest) (*pbShortVideo.GetLikeShortVideoListResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoLikeList", request)
	response := &pbShortVideo.GetLikeShortVideoListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	// cache
	cacheKey := request.OperatorUserId + "_" + request.UserId + "_" + strconv.FormatInt(int64(request.Pagination.PageNumber), 10) +
		"_" + strconv.FormatInt(int64(request.Pagination.ShowNumber), 10)
	cache, err := db.DB.ShortVideoGetLikeListCache(cacheKey)
	if err == nil {
		_ = json.Unmarshal([]byte(cache), &response)
		return response, nil
	}

	var idList []string
	var count int64
	// Check out self like short videos
	if request.UserId == request.OperatorUserId {
		idList, count, err = im_mysql_model.GetLikeShortVideoFileIdList(request.OperatorUserId, true, true, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	}

	// Check out other like short videos
	if request.UserId != request.OperatorUserId {
		// authority
		isFriend := false
		connection, err := createFriendGRPConnection(request.OperationID)
		if err == nil {
			rpcRequest := &pbFriend.IsFriendReq{CommID: &pbFriend.CommID{
				OperationID: request.OperationID,
				FromUserID:  request.UserId,
				ToUserID:    request.OperatorUserId,
			}}
			friend, err := connection.IsFriend(ctx, rpcRequest)
			if err == nil {
				isFriend = friend.Response
			}
		}

		idList, count, err = im_mysql_model.GetLikeShortVideoFileIdList(request.UserId, false, isFriend, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	}

	response.ShortVideoCount = count
	if len(idList) == 0 {
		return response, nil
	}

	// Get short video info
	list, err := s.GetShortVideoByFieldIdList(ctx, &pbShortVideo.GetShortVideoByFileIdListRequest{
		OperationID: request.OperationID,
		UserId:      request.UserId,
		FileIdList:  idList,
	})
	response.ShortVideoInfoList = list.ShortVideoInfoList

	// redis
	marshal, _ := json.Marshal(response)
	_ = db.DB.SetShortVideoLikeListCache(cacheKey, string(marshal))

	return response, nil
}

func (s *Server) GetShortVideoListByUserId(ctx context.Context, request *pbShortVideo.GetShortVideoListByUserIdRequest) (*pbShortVideo.GetShortVideoListByUserIdResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoListByUserId", request)
	response := &pbShortVideo.GetShortVideoListByUserIdResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &pbShortVideo.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}

	// cache
	cacheKey := request.OperatorUserId + "_" + request.UserId + "_" + strconv.FormatInt(int64(request.Pagination.PageNumber), 10) +
		"_" + strconv.FormatInt(int64(request.Pagination.ShowNumber), 10)
	cache, err := db.DB.ShortVideoListByUserIdCache(cacheKey)
	if err == nil && len(cache) > 0 {
		_ = json.Unmarshal([]byte(cache), &response)
		return response, nil
	}

	var idList []string
	var count int64
	// Check out self like short videos
	if request.UserId == request.OperatorUserId {
		idList, count, err = im_mysql_model.GetShortVideoFileIdListByUserId(request.OperatorUserId, true, true, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	}

	// Check out other short videos
	if request.UserId != request.OperatorUserId {
		// authority
		isFriend := false
		connection, err := createFriendGRPConnection(request.OperationID)
		if err == nil {
			rpcRequest := &pbFriend.IsFriendReq{CommID: &pbFriend.CommID{
				OperationID: request.OperationID,
				FromUserID:  request.UserId,
				ToUserID:    request.OperatorUserId,
			}}
			friend, err := connection.IsFriend(ctx, rpcRequest)
			if err == nil {
				isFriend = friend.Response
			}
		}

		idList, count, err = im_mysql_model.GetShortVideoFileIdListByUserId(request.UserId, false, isFriend, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	}

	response.ShortVideoCount = count
	if len(idList) == 0 {
		return response, nil
	}

	// Get short video info
	list, err := s.GetShortVideoByFieldIdList(ctx, &pbShortVideo.GetShortVideoByFileIdListRequest{
		OperationID:     request.OperationID,
		UserId:          request.OperatorUserId,
		FileIdList:      idList,
		OperationUserId: request.OperatorUserId,
	})

	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = "get short video failed, err: " + err.Error()
		return response, err
	}

	log.NewInfo(request.OperationID, "GetShortVideoLikeList", list)

	response.ShortVideoInfoList = list.ShortVideoInfoList
	if len(list.ShortVideoInfoList) > 0 {
		response.IsUpFollow = list.ShortVideoInfoList[0].IsUpFollow
	}

	// redis
	marshal, _ := json.Marshal(response)
	_ = db.DB.SetShortVideoListByUserIdCache(cacheKey, string(marshal))

	return response, nil
}

func (s *Server) GetShortVideoByFieldId(ctx context.Context, request *pbShortVideo.GetShortVideoByFileIdRequest) (*pbShortVideo.GetShortVideoByFileIdResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoByField", request)
	response := &pbShortVideo.GetShortVideoByFileIdResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	updateUserInfo := &db.User{}

	shortVideoInfo, err := im_mysql_model.GetShortVideoByFileId(request.FileId)
	if err != nil {
		log.NewError(request.OperationID, "GetShortVideoByField", "get short video error", err)
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// deleted
	if shortVideoInfo.Status == constant.ShortVideoTypeDeleted {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	// Not public
	if shortVideoInfo.Status != constant.ShortVideoTypeNormal {
		// Not Admin
		if request.IsAdmin == false {
			// Not owner
			if (shortVideoInfo.Status == constant.ShortVideoTypePrivate || shortVideoInfo.Status == constant.ShortVideoTypeAudit) && shortVideoInfo.UserId != request.UserId {
				response.CommonResp = &pbShortVideo.CommonResp{
					ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
					ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
				}
				return response, nil
			}
			// Not friend
			if shortVideoInfo.Status == constant.ShortVideoTypeFriend {
				connection, err := createFriendGRPConnection(request.OperationID)
				if err != nil {
					response.CommonResp = &pbShortVideo.CommonResp{
						ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
						ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
					}
					return response, nil
				}

				rpcRequest := &pbFriend.IsFriendReq{CommID: &pbFriend.CommID{
					OperationID: request.OperationID,
					FromUserID:  request.UserId,
					ToUserID:    shortVideoInfo.UserId,
				}}
				friend, err := connection.IsFriend(ctx, rpcRequest)
				if err != nil {
					response.CommonResp = &pbShortVideo.CommonResp{
						ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
						ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
					}
					return response, nil
				}

				if friend.Response == false {
					response.CommonResp = &pbShortVideo.CommonResp{
						ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
						ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
					}
					return response, nil
				}
			}
		}
	}

	response.ShortVideoInfo = &pbShortVideo.ShortVideoInfo{}
	err = utils.CopyStructFields(response.ShortVideoInfo, shortVideoInfo)
	if err != nil {
		return nil, err
	}
	response.ShortVideoInfo.ClassId = int32(shortVideoInfo.ClassId)
	response.ShortVideoInfo.Status = int32(shortVideoInfo.Status)
	response.ShortVideoInfo.CreateTime = shortVideoInfo.CreateTime
	response.ShortVideoInfo.UpdateTime = shortVideoInfo.UpdateTime
	response.ShortVideoInfo.SelfInfo = &pbShortVideo.OperationMessage{
		IsLike:    db.DB.ShortVideoIsLike(request.UserId, request.FileId),
		IsComment: db.DB.ShortVideoIsComment(request.UserId, request.FileId),
	}

	follow, _ := s.IsFollow(ctx, &pbShortVideo.IsFollowRequest{
		OperationID:  request.OperationID,
		UserId:       request.UserId,
		FollowUserId: response.ShortVideoInfo.UserId,
	})
	response.ShortVideoInfo.IsUpFollow = &pbShortVideo.IsFollowMessage{
		IsFollow: follow.IsFollow,
	}

	// get user info by user id
	updateUserInfo, err = im_mysql_model.GetUserByUserID(shortVideoInfo.UserId)
	response.ShortVideoInfo.UpUserInfo = &pbShortVideo.UserInfoMessage{
		IsDeleted: true,
	}
	if err == nil {
		response.ShortVideoInfo.UpUserInfo = &pbShortVideo.UserInfoMessage{
			UserId:    updateUserInfo.UserID,
			Nickname:  updateUserInfo.Nickname,
			FaceURL:   updateUserInfo.FaceURL,
			IsDeleted: false,
		}
	}

	// gorse
	go func() {
		utils2.InsertFeedback(constant.NegativeFeedbackRead, request.UserId, request.FileId)
	}()

	return response, nil
}

func (s *Server) BlockShortVideo(_ context.Context, request *pbShortVideo.BlockShortVideoRequest) (*pbShortVideo.BlockShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "BlockShortVideo", request)
	response := &pbShortVideo.BlockShortVideoResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	go func() {
		// gorse
		utils2.InsertFeedback(constant.NegativeFeedbackBlock, request.UserId, request.FileId)
	}()

	return response, nil
}

func (s *Server) GetShortVideoByFieldIdList(ctx context.Context, request *pbShortVideo.GetShortVideoByFileIdListRequest) (*pbShortVideo.GetShortVideoByFileIdListResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoByFieldIdList", request)
	response := &pbShortVideo.GetShortVideoByFileIdListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	shortVideoInfo, err := im_mysql_model.GetShortVideoByFileIdList(request.FileIdList)
	if err != nil || len(shortVideoInfo) == 0 {
		log.NewError(request.OperationID, "GetShortVideoByField", "get short video error", err)
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdIsNotExist.ErrCode,
			ErrMsg:  constant.ErrFileIdIsNotExist.ErrMsg,
		}
		return response, nil
	}

	response.ShortVideoInfoList = make([]*pbShortVideo.ShortVideoInfo, len(shortVideoInfo))
	_ = utils.CopyStructFields(&response.ShortVideoInfoList, shortVideoInfo)

	userIdList := make([]string, 0)
	for i, shortVideo := range response.ShortVideoInfoList {
		userIdList = append(userIdList, shortVideo.UserId)

		response.ShortVideoInfoList[i].ClassId = shortVideo.ClassId
		response.ShortVideoInfoList[i].Status = shortVideo.Status
		response.ShortVideoInfoList[i].CreateTime = shortVideo.CreateTime
		response.ShortVideoInfoList[i].UpdateTime = shortVideo.UpdateTime
		response.ShortVideoInfoList[i].SelfInfo = &pbShortVideo.OperationMessage{
			IsLike:    db.DB.ShortVideoIsLike(request.UserId, shortVideo.FileId),
			IsComment: db.DB.ShortVideoIsComment(request.UserId, shortVideo.FileId),
		}

		follow, _ := s.IsFollow(ctx, &pbShortVideo.IsFollowRequest{
			OperationID:  request.OperationID,
			UserId:       request.UserId,
			FollowUserId: shortVideo.UserId,
		})
		response.ShortVideoInfoList[i].IsUpFollow = &pbShortVideo.IsFollowMessage{
			IsFollow: follow.IsFollow,
		}

		// clear json
		response.ShortVideoInfoList[i].Json = ""
	}

	updateUserInfo, err := im_mysql_model.GetUserInfoByUserIDs(userIdList)
	if err == nil {
		userInfoMap := make(map[string]*pbShortVideo.UserInfoMessage)
		for _, userInfo := range updateUserInfo {
			userInfoMap[userInfo.UserID] = &pbShortVideo.UserInfoMessage{
				UserId:    userInfo.UserID,
				Nickname:  userInfo.Nickname,
				FaceURL:   userInfo.FaceURL,
				IsDeleted: false,
			}
		}

		for i, shortVideo := range response.ShortVideoInfoList {
			response.ShortVideoInfoList[i].UpUserInfo = &pbShortVideo.UserInfoMessage{
				IsDeleted: true,
			}
			if userInfo, ok := userInfoMap[shortVideo.UserId]; ok {
				response.ShortVideoInfoList[i].UpUserInfo = userInfo
			}
		}
	}
	return response, nil
}

func (s *Server) GetRecommendShortVideoList(ctx context.Context, request *pbShortVideo.GetRecommendShortVideoListRequest) (*pbShortVideo.GetRecommendShortVideoListResponse, error) {
	log.NewInfo(request.OperationID, "GetRecommendShortVideoList", request)
	response := &pbShortVideo.GetRecommendShortVideoListResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// get recommend short video list
	shortVideoIdList := utils2.GetRecommend(request.UserId, constant.ShortString, int(request.Size))
	log.NewInfo(request.OperationID, "GetRecommendShortVideoList", request.UserId, constant.ShortString, int(request.Size), shortVideoIdList)

	if len(shortVideoIdList) == 0 {
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrRecommendNil.ErrCode,
			ErrMsg:  constant.ErrRecommendNil.ErrMsg,
		}
		return response, nil
	}
	response.FileIdList = shortVideoIdList

	go func() {
		// gorse
		for _, fileId := range shortVideoIdList {
			utils2.InsertFeedback(constant.NegativeFeedbackRead, request.UserId, fileId)
		}
	}()

	// Get short video info
	list, _ := s.GetShortVideoByFieldIdList(ctx, &pbShortVideo.GetShortVideoByFileIdListRequest{
		OperationID: request.OperationID,
		UserId:      request.UserId,
		FileIdList:  shortVideoIdList,
	})
	response.ShortVideoInfoList = list.ShortVideoInfoList

	return response, nil
}

func (s *Server) FileUploadCallBack(_ context.Context, request *pbShortVideo.FileUploadCallBackRequest) (*pbShortVideo.FileUploadCallBackResponse, error) {
	log.NewInfo(request.OperationID, "FileUploadCallBack", request)
	response := &pbShortVideo.FileUploadCallBackResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check
	short, getErr := im_mysql_model.GetShortVideoByFileId(request.FileUploadEvent.FileId)
	if getErr == nil && short.FileId == request.FileUploadEvent.FileId {
		// update
		info := &db.ShortVideo{
			FileId:   request.FileUploadEvent.FileId,
			Status:   constant.ShortVideoTypeNormal,
			MediaUrl: request.FileUploadEvent.MediaBasicInfo.MediaUrl,
			CoverUrl: request.FileUploadEvent.MediaBasicInfo.CoverUrl,
			Type:     request.FileUploadEvent.MediaBasicInfo.Type,
			Size:     int64(request.FileUploadEvent.MetaData.Size),
			Width:    int64(request.FileUploadEvent.MetaData.Width),
			Height:   int64(request.FileUploadEvent.MetaData.Height),
			Duration: request.FileUploadEvent.MetaData.Duration,
		}

		updateErr := im_mysql_model.AlterShortVideo(info)
		if updateErr != nil {
			response.CommonResp = &pbShortVideo.CommonResp{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  constant.ErrDB.ErrMsg,
			}
			return response, nil
		}

		return response, nil
	}

	sourceContext := strings.Split(request.FileUploadEvent.MediaBasicInfo.SourceInfo.SourceContext, "丨")
	// vod manager
	if len(sourceContext) != 3 {
		sourceContext = []string{}
		sourceContext = append(sourceContext, "")
		sourceContext = append(sourceContext, "")
		sourceContext = append(sourceContext, "")
	}

	// mysql
	insertData := &db.ShortVideo{}
	insertData.FileId = request.FileUploadEvent.FileId
	insertData.Name = request.FileUploadEvent.MediaBasicInfo.Name
	insertData.Desc = sourceContext[1]
	insertData.InterestId = sourceContext[2]
	insertData.ClassId = int8(request.FileUploadEvent.MediaBasicInfo.ClassId)
	insertData.ClassName = request.FileUploadEvent.MediaBasicInfo.ClassName
	insertData.CoverUrl = request.FileUploadEvent.MediaBasicInfo.CoverUrl
	insertData.MediaUrl = request.FileUploadEvent.MediaBasicInfo.MediaUrl
	insertData.ClassName = request.FileUploadEvent.MediaBasicInfo.ClassName
	insertData.Type = request.FileUploadEvent.MediaBasicInfo.Type

	insertData.Size = int64(request.FileUploadEvent.MetaData.Size)
	insertData.Height = int64(request.FileUploadEvent.MetaData.Height)
	insertData.Width = int64(request.FileUploadEvent.MetaData.Width)
	insertData.Duration = request.FileUploadEvent.MetaData.Duration
	marshal, _ := json.Marshal(request)
	insertData.Json = string(marshal)

	insertData.UserId = sourceContext[0]
	if insertData.UserId == "" {
		insertData.UserId = config.Config.Manager.AppManagerUid[0]
	}

	parse, _ := time.Parse(time.RFC3339, request.FileUploadEvent.MediaBasicInfo.CreateTime)
	insertData.CreateTime = parse.Unix()

	dbErr := im_mysql_model.InsertShortVideo(insertData)
	if dbErr != nil {
		log.NewError(request.OperationID, "FileUploadCallBack", "insert short video error", dbErr)
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	go func() {
		// gorse
		gorseResult := utils2.InsertItem(request.FileUploadEvent.FileId, constant.ShortString, request.FileUploadEvent.MediaBasicInfo.UpdateTime, strings.Split(sourceContext[2], ","))
		if gorseResult == false {
			log.NewWarn(request.OperationID, "FileUploadCallBack", "gorseResult", gorseResult, request.FileUploadEvent.FileId)
		}

		// count
		im_mysql_model.IncrShortVideoUserCountByUserId(insertData.UserId, "work_num", 1)

		// notices
		followerList, err := im_mysql_model.GetFansIdByUserId(insertData.UserId)
		if err == nil && len(followerList) > 0 {
			for _, follower := range followerList {
				insertData := db.ShortVideoNotice{
					UserId:       follower,
					SourceUserId: insertData.UserId,
					FileId:       request.FileUploadEvent.FileId,
					Type:         constant.ShortVideoNoticeTypeNewPost,
					Context:      insertData.Desc,
				}
				result := im_mysql_model.AddShortVideoNotice(insertData)
				if result != 0 {
					db.DB.AddShortVideoNotice(follower, result)
				}
			}
		}
	}()

	return response, nil
}

func (s *Server) ProcedureStateChangeCallBack(_ context.Context, request *pbShortVideo.ProcedureStateChangeCallBackRequest) (*pbShortVideo.ProcedureStateChangeCallBackResponse, error) {
	log.NewInfo(request.OperationID, "ProcedureStateChangeCallBack", request)
	response := &pbShortVideo.ProcedureStateChangeCallBackResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check file id
	_, getErr := im_mysql_model.GetShortVideoByFileId(request.ProcedureStateChangeEvent.FileId)
	if getErr != nil {
		log.NewInfo(request.OperationID, "ProcedureStateChangeEvent", "fileId is not exist")
		response.CommonResp = &pbShortVideo.CommonResp{
			ErrCode: constant.ErrFileIdAlreadyExist.ErrCode,
			ErrMsg:  constant.ErrFileIdAlreadyExist.ErrMsg,
		}
		return response, nil
	}

	if len(request.ProcedureStateChangeEvent.MediaProcessResultSet) > 0 {
		for _, mediaProcessResult := range request.ProcedureStateChangeEvent.MediaProcessResultSet {
			if mediaProcessResult.Type == "Transcode" {
				if mediaProcessResult.TranscodeTask.Status == "SUCCESS" {
					// upload task id, vod configuration
					if mediaProcessResult.TranscodeTask.Input.Definition == 100830 {
						mediaUrl := mediaProcessResult.TranscodeTask.Output.URL
						err := im_mysql_model.UpdateShortVideoMediaUrlByFileId(request.ProcedureStateChangeEvent.FileId, mediaUrl)
						if err != nil {
							log.NewError(request.OperationID, "ProcedureStateChangeEvent", "update short video media url error", request)
							return response, nil
						}
					}
				}
			}
		}
	}

	return response, nil
}

func (s *Server) FileDeletedCallBack(_ context.Context, request *pbShortVideo.FileDeletedCallBackRequest) (*pbShortVideo.FileDeletedCallBackResponse, error) {
	log.NewInfo(request.OperationID, "FileDeletedCallBack", request)
	response := &pbShortVideo.FileDeletedCallBackResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	err := im_mysql_model.MultiDeleteShortVideoByFileId(request.FileDeleteEvent.FileIdSet)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	// short video info
	shortVideoInfo, err := im_mysql_model.GetShortVideoByFileIdList(request.FileDeleteEvent.FileIdSet)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}
	userWorkCountData := make(map[string]int64)
	userHarvestedLikes := make(map[string]int64)
	for _, v := range shortVideoInfo {
		if _, ok := userWorkCountData[v.UserId]; !ok {
			userWorkCountData[v.UserId] = 1
		} else {
			userWorkCountData[v.UserId] = userWorkCountData[v.UserId] + 1
		}

		if _, ok := userHarvestedLikes[v.UserId]; !ok {
			userHarvestedLikes[v.UserId] = v.LikeNum
		} else {
			userHarvestedLikes[v.UserId] = userHarvestedLikes[v.UserId] + v.LikeNum
		}
	}

	go func() {
		// gorse
		for _, fileId := range request.FileDeleteEvent.FileIdSet {
			utils2.UpdateItem(fileId, true, constant.ShortString, "", strings.Split("", ","))
			utils2.DeleteItem(fileId)
		}

		// count
		for userId, count := range userWorkCountData {
			im_mysql_model.IncrShortVideoUserCountByUserId(userId, "work_num", -count)
		}
		for userId, count := range userHarvestedLikes {
			im_mysql_model.IncrShortVideoUserCountByUserId(userId, "harvested_likes_number", -count)
		}
	}()

	return response, nil
}

func (s *Server) GetUpdateShortVideoSign(_ context.Context, request *pbShortVideo.GetUpdateShortVideoSignRequest) (*pbShortVideo.GetUpdateShortVideoSignResponse, error) {
	log.NewInfo(request.OperationID, "GetUpdateShortVideoSign", request)
	response := &pbShortVideo.GetUpdateShortVideoSignResponse{}
	response.CommonResp = &pbShortVideo.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	rand.Seed(time.Now().Unix())
	secretId := config.Config.Vod.SecretId
	secretKey := config.Config.Vod.SecretKey
	timestamp := time.Now().Unix()
	expireTime := timestamp + 86400

	if len(request.InterestId) == 0 {
		request.InterestId = []int32{0}
	}
	interestIdString, _ := json.Marshal(request.InterestId)
	sourceContext := request.UserId + "丨" + request.Desc + "丨" + strings.Trim(string(interestIdString), "[]")
	timestampStr := strconv.FormatInt(timestamp, 10)
	expireTimeStr := strconv.FormatInt(expireTime, 10)
	randomStr := strconv.Itoa(rand.Int())
	original := "secretId=" + secretId + "&currentTimeStamp=" + timestampStr + "&expireTime=" + expireTimeStr + "&random=" + randomStr + "&sourceContext=" + sourceContext
	original += "&procedure=upload&taskPriority=10&vodSubAppId=" + utils.IntToString(config.Config.Vod.VodSubAppId)

	signature := generateHmacSHA1(secretKey, original)
	signature = append(signature, []byte(original)...)
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	response.Sign = signatureB64
	return response, nil
}

func NewShortVideoServer(port int) *Server {
	return &Server{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImShortVideoName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *Server) Run() {
	log.NewPrivateLog(constant.OpenImShortVideoLog)
	log.NewInfo("0", "short video rpc start ")

	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)

	//listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("0", "listen network success, ", address, listener)
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.NewError("0", "listener close err", err)
		}
	}(listener)
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	//Service registers with etcd
	pbShortVideo.RegisterShortVideoServer(srv, s)
	rpcRegisterIP := ""
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error())
		return
	}
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "short video rpc success")
}

func generateHmacSHA1(secretToken, payloadBody string) []byte {
	mac := hmac.New(sha1.New, []byte(secretToken))
	sha1.New()
	mac.Write([]byte(payloadBody))
	return mac.Sum(nil)
}

func createFriendGRPConnection(OperationID string) (pbFriend.FriendClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImFriendName, OperationID)
	if etcdConn == nil {
		errMsg := "etcd3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbFriend.NewFriendClient(etcdConn)
	return client, nil
}

func createUserGRPConnection(OperationID string) (pbUser.UserClient, error) {
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, OperationID)
	if etcdConn == nil {
		errMsg := "etcd3.GetConn is not created"
		log.NewError(errMsg)
		return nil, errors.New(errMsg)
	}
	client := pbUser.NewUserClient(etcdConn)
	return client, nil
}
