package admin_cms

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdminCMS "Open_IM/pkg/proto/admin_cms"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (s *adminCMSServer) ManagementShortVideo(ctx context.Context, request *pbAdminCMS.ManagementShortVideoRequest) (*pbAdminCMS.ManagementShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "ManagementShortVideo", request)
	response := &pbAdminCMS.ManagementShortVideoResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &sdkws.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}
	response.TotalCount = 0

	var userIdList []string
	if request.UserId != "" {
		userIdList = imdb.GetUserIdByAllCondition(request.UserId)
		if len(userIdList) == 0 {
			userIdList = append(userIdList, "0")
		}
	}

	list, count, err := imdb.ManageGetShortVideoList(userIdList, int8(request.Status), request.Desc, request.EmptyDesc, int8(request.IsBlock), request.StartTime, request.EndTime, request.Pagination.PageNumber, request.Pagination.ShowNumber, []int64{})
	if err != nil {
		return response, openIMHttp.WrapError(constant.ErrDB)
	}
	response.TotalCount = count

	if len(list) == 0 {
		return response, nil
	}

	userIdList = []string{}
	for _, v := range list {
		userIdList = append(userIdList, v.UserId)
	}

	if len(userIdList) == 0 {
		return response, nil
	}

	connection, err := createUserGRPConnection(request.OperationID)
	if err != nil {
		return response, nil
	}

	rpcRequest := &pbUser.GetUserInfoReq{
		UserIDList: userIdList,
	}
	userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
	if err != nil {
		return response, nil
	}
	userInfoMap := make(map[string]*sdkws.UserInfo)
	for _, user := range userInfo.UserInfoList {
		userInfoMap[user.UserID] = user
	}

	var fileIdList []string
	for _, v := range list {
		fileIdList = append(fileIdList, v.FileId)
	}
	commentCount, _ := imdb.ManagementGetShortVideoCommentCountByFileIdList(fileIdList)

	response.ShortVideoInfo = []*pbAdminCMS.ShortVideoInfoMessage{}
	for _, v := range list {
		shortVideoInfo := &pbAdminCMS.ShortVideoInfoMessage{
			Id:             v.Id,
			UserId:         v.UserId,
			UserName:       v.UserId,
			Status:         int32(v.Status),
			CreateTime:     v.CreateTime,
			MediaUrl:       v.MediaUrl,
			CoverUrl:       v.CoverUrl,
			Desc:           v.Desc,
			LikeNum:        v.LikeNum,
			CommentNum:     0,
			ReplyNum:       v.ReplyNum,
			CommentLikeNum: v.CommentLikeNum,
			Remark:         v.Remark,
			FileId:         v.FileId,
			InterestId:     v.InterestId,
		}
		if _, ok := commentCount[v.FileId]; ok {
			shortVideoInfo.CommentNum = commentCount[v.FileId]
		}

		if userInfoMap[v.UserId] != nil {
			shortVideoInfo.UserName = userInfoMap[v.UserId].Nickname
		}
		response.ShortVideoInfo = append(response.ShortVideoInfo, shortVideoInfo)
	}

	return response, nil
}

func (s *adminCMSServer) DeleteShortVideo(ctx context.Context, request *pbAdminCMS.DeleteShortVideoRequest) (*pbAdminCMS.DeleteShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "DeleteShortVideo", request)
	response := &pbAdminCMS.DeleteShortVideoResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	if len(request.FileId) == 0 {
		return response, nil
	}

	// short video info
	shortVideoInfo, err := imdb.GetShortVideoByFileIdList(request.FileId)
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

	err = imdb.MultiDeleteShortVideoByFileId(request.FileId)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	go func() {
		// gorse
		for _, fileId := range request.FileId {
			utils2.UpdateItem(fileId, true, constant.ShortString, "", strings.Split("", ","))
			utils2.DeleteItem(fileId)
		}

		// count
		for userId, count := range userWorkCountData {
			imdb.IncrShortVideoUserCountByUserId(userId, "work_num", -count)
		}
		for userId, count := range userHarvestedLikes {
			imdb.IncrShortVideoUserCountByUserId(userId, "harvested_likes_number", -count)
		}
	}()

	return response, nil
}

func (s *adminCMSServer) AlterShortVideo(ctx context.Context, request *pbAdminCMS.AlterShortVideoRequest) (*pbAdminCMS.AlterShortVideoResponse, error) {
	log.NewInfo(request.OperationID, "AlterShortVideo", request)
	response := &pbAdminCMS.AlterShortVideoResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	// check short video info
	shortInfo, err := imdb.GetShortVideoByFileId(request.FileId)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	err = imdb.UpdateShortVideoInfoByFileId(request.FileId, &db.ShortVideo{
		Status: int8(request.Status),
		Remark: request.Remark,
		Desc:   request.Desc,
	})
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	go func() {
		// gorse
		utils2.UpdateItem(request.FileId, request.Status != constant.ShortVideoTypeNormal, constant.ShortString, request.Desc, strings.Split(shortInfo.InterestId, ","))
	}()

	return response, nil
}

func (s *adminCMSServer) GetShortVideoLikeList(ctx context.Context, request *pbAdminCMS.GetShortVideoLikeListRequest) (*pbAdminCMS.GetShortVideoLikeListResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoLikeList", request)
	response := &pbAdminCMS.GetShortVideoLikeListResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &sdkws.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}
	response.TotalCount = 0

	var userIdList []string
	if request.UserId != "" {
		userIdList = imdb.GetUserIdByAllCondition(request.UserId)
		if len(userIdList) == 0 {
			userIdList = append(userIdList, "0")
		}
	}

	var likeUserIdList []string
	if request.LikeUserId != "" {
		likeUserIdList = imdb.GetUserIdByAllCondition(request.LikeUserId)
		if len(likeUserIdList) == 0 {
			likeUserIdList = append(likeUserIdList, "0")
		}
	}

	list, count, err := imdb.GetShortVideoLikeList(userIdList, likeUserIdList, request.FileId, request.Status, request.Desc, int32(request.EmptyDesc), request.StartTime, request.EndTime, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}
	response.TotalCount = count

	if len(list) == 0 {
		return response, nil
	}

	userIdList = []string{}
	var fileIdList []string
	for _, v := range list {
		userIdList = append(userIdList, v.UserId)
		fileIdList = append(fileIdList, v.FileId)
	}

	// short video info
	idList, err := imdb.GetShortVideoByFileIdList(fileIdList)
	if err != nil {
		return response, err
	}
	shortVideoInfoMap := make(map[string]*db.ShortVideo)
	for _, v := range idList {
		shortVideoInfoMap[v.FileId] = v
		userIdList = append(userIdList, v.UserId)
	}

	// user info
	connection, err := createUserGRPConnection(request.OperationID)
	if err != nil {
		return response, nil
	}

	rpcRequest := &pbUser.GetUserInfoReq{
		UserIDList: userIdList,
	}
	userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
	if err != nil {
		return response, nil
	}
	userInfoMap := make(map[string]*sdkws.UserInfo)
	for _, user := range userInfo.UserInfoList {
		userInfoMap[user.UserID] = user
	}

	for _, v := range list {
		shortVideoLikeInfo := &pbAdminCMS.ShortVideoLikeMessage{
			Id:         v.Id,
			FileId:     v.FileId,
			UserId:     v.UserId,
			CreateTime: v.CreateTime,
			PostUserId: "",
			FileStatus: 0,
		}
		if userInfoMap[v.UserId] != nil {
			shortVideoLikeInfo.UserName = userInfoMap[v.UserId].Nickname
		}
		if shortVideoInfoMap[v.FileId] != nil {
			shortVideoLikeInfo.MediaUrl = shortVideoInfoMap[v.FileId].MediaUrl
			shortVideoLikeInfo.CoverUrl = shortVideoInfoMap[v.FileId].CoverUrl
			shortVideoLikeInfo.PostUserId = shortVideoInfoMap[v.FileId].UserId
			if userInfoMap[shortVideoInfoMap[v.FileId].UserId] != nil {
				shortVideoLikeInfo.PostUserName = userInfoMap[shortVideoInfoMap[v.FileId].UserId].Nickname
			}
			shortVideoLikeInfo.FileStatus = int32(shortVideoInfoMap[v.FileId].Status)
		}

		response.ShortVideoLike = append(response.ShortVideoLike, shortVideoLikeInfo)
	}

	return response, nil
}

func (s *adminCMSServer) DeleteShortVideoLike(ctx context.Context, request *pbAdminCMS.DeleteShortVideoLikeRequest) (*pbAdminCMS.DeleteShortVideoLikeResponse, error) {
	log.NewInfo(request.OperationID, "DeleteShortVideoLike", request)
	response := &pbAdminCMS.DeleteShortVideoLikeResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	if len(request.LikeIdList) == 0 {
		return response, nil
	}

	// get short video like info
	list, err := imdb.GetShortVideoLikeByLikeIdList(request.LikeIdList)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	waitGroup := sync.WaitGroup{}
	for _, v := range list {
		waitGroup.Add(1)
		go func(v *db.ShortVideoLike) {
			defer waitGroup.Done()
			// mysql
			imdb.DeleteShortVideoLike(v.FileId, v.UserId)
			// redis
			db.DB.ShortVideoCancelLike(v.UserId, v.FileId)
			// count
			imdb.IncrShortVideoUserCountByUserId(v.UserId, "like_num", -1)
		}(v)
	}
	waitGroup.Wait()

	return response, nil
}

func (s *adminCMSServer) GetShortVideoCommentList(ctx context.Context, request *pbAdminCMS.GetShortVideoCommentListRequest) (*pbAdminCMS.GetShortVideoCommentListResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoCommentList", request)
	response := &pbAdminCMS.GetShortVideoCommentListResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &sdkws.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}
	response.TotalCount = 0

	var userIdList []string
	if request.UserId != "" {
		userIdList = imdb.GetUserIdByAllCondition(request.UserId)
		if len(userIdList) == 0 {
			userIdList = append(userIdList, "0")
		}
	}

	var commentUserIdList []string
	if request.CommentUserId != "" {
		commentUserIdList = imdb.GetUserIdByAllCondition(request.CommentUserId)
		if len(commentUserIdList) == 0 {
			commentUserIdList = append(commentUserIdList, "0")
		}
	}

	list, count, err := imdb.ManagementGetShortVideoCommentList(0, userIdList, commentUserIdList, request.FileId, request.Status, request.Desc, request.Content, request.EmptyDesc, request.StartTime, request.EndTime, request.Pagination.PageNumber, request.Pagination.ShowNumber)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}
	response.TotalCount = count

	if len(list) == 0 {
		return response, nil
	}

	userIdList = []string{}
	for _, v := range list {
		if v.UserId != "" {
			userIdList = append(userIdList, v.UserId)
		}
		if v.PostUserId != "" {
			userIdList = append(userIdList, v.PostUserId)
		}
	}

	// user info
	connection, err := createUserGRPConnection(request.OperationID)
	if err != nil {
		return response, nil
	}

	rpcRequest := &pbUser.GetUserInfoReq{
		UserIDList: userIdList,
	}
	userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
	if err != nil {
		return response, nil
	}
	userInfoMap := make(map[string]*sdkws.UserInfo)
	for _, user := range userInfo.UserInfoList {
		userInfoMap[user.UserID] = user
	}
	for _, v := range list {
		shortVideoCommentInfo := &pbAdminCMS.ShortVideoCommentMessage{
			Id:         v.CommentId,
			FileId:     v.FileId,
			MediaUrl:   v.MediaUrl,
			CoverUrl:   v.CoverUrl,
			UserId:     v.UserId,
			CreateTime: v.CreateTime,
			Content:    v.Content,
			LikeNum:    v.CommentLikeCount,
			ReplyNum:   v.CommentReplyCount,
			Remark:     v.Remark,
			Desc:       v.Desc,
			Status:     v.Status,
		}

		if userInfoMap[v.UserId] != nil {
			shortVideoCommentInfo.UserName = userInfoMap[v.UserId].Nickname
		}
		if userInfoMap[v.PostUserId] != nil {
			shortVideoCommentInfo.PostUserId = userInfoMap[v.PostUserId].UserID
			shortVideoCommentInfo.PostUserName = userInfoMap[v.PostUserId].Nickname
		}
		response.ShortVideoComment = append(response.ShortVideoComment, shortVideoCommentInfo)
	}

	return response, nil
}

func (s *adminCMSServer) DeleteShortVideoComment(ctx context.Context, request *pbAdminCMS.DeleteShortVideoCommentRequest) (*pbAdminCMS.DeleteShortVideoCommentResponse, error) {
	log.NewInfo(request.OperationID, "DeleteShortVideoComment", request)
	response := &pbAdminCMS.DeleteShortVideoCommentResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	if len(request.CommentIdList) == 0 {
		return response, nil
	}

	list, err := imdb.GetShortVideoCommentByCommentIdList(request.CommentIdList)
	if err != nil {
		return response, openIMHttp.WrapError(constant.ErrDB)
	}

	waitGroup := sync.WaitGroup{}
	for _, v := range list {
		waitGroup.Add(1)
		go func(v *db.ShortVideoComment) {
			defer waitGroup.Done()

			allCommentInfoList := make([]*db.ShortVideoComment, 0)
			allCommentInfoList = append(allCommentInfoList, v)
			childCommentInfoList, _ := imdb.GetAllReplyCommentIdByCommentId(v.CommentId)
			if len(childCommentInfoList) > 0 {
				childCommentInfoList, _ := imdb.GetCommentInfoListByCommentId(childCommentInfoList)
				for _, childCommentInfo := range childCommentInfoList {
					if childCommentInfo.Status == constant.ShortVideoCommentStatusNormal || childCommentInfo.Status == constant.ShortVideoCommentStatusAudit {
						allCommentInfoList = append(allCommentInfoList, childCommentInfo)
					}
				}
			}

			// mysql
			err = imdb.DeleteShortVideoCommentByCommentList(allCommentInfoList)
			if err == nil {
				// count
				for _, commentInfo := range allCommentInfoList {
					imdb.IncrShortVideoUserCountByUserId(commentInfo.UserID, "comment_num", -1)
				}
				db.DB.ShortVideoDelCommentListCache(v.FileId + "*")
			}
		}(v)
	}
	waitGroup.Wait()

	return response, nil
}

func (s *adminCMSServer) AlterShortVideoComment(ctx context.Context, request *pbAdminCMS.AlterShortVideoCommentRequest) (*pbAdminCMS.AlterShortVideoCommentResponse, error) {
	log.NewInfo(request.OperationID, "AlterShortVideoComment", request)
	response := &pbAdminCMS.AlterShortVideoCommentResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	if request.CommentId == 0 {
		response.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		response.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg
		return response, nil
	}

	info, err := imdb.GetShortVideoCommentByCommentId(request.CommentId)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	err = imdb.UpdateShortVideoCommentByCommentId(request.CommentId, request.Content, request.Remark)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	// cache
	db.DB.ShortVideoDelCommentListCache(info.FileId + "*")

	return response, nil
}

func (s *adminCMSServer) GetShortVideoInterestLabelList(ctx context.Context, request *pbAdminCMS.GetShortVideoInterestLabelListRequest) (*pbAdminCMS.GetShortVideoInterestLabelListResponse, error) {
	log.NewInfo(request.OperationID, "GetShortVideoInterestLabelList", request)
	response := &pbAdminCMS.GetShortVideoInterestLabelListResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	response.Pagination = &sdkws.ResponsePagination{
		CurrentPage: request.Pagination.PageNumber,
		ShowNumber:  request.Pagination.ShowNumber,
	}
	response.TotalCount = 0

	var userIdList []string
	if request.UserId != "" {
		userIdList = imdb.GetUserIdByAllCondition(request.UserId)
		if len(userIdList) == 0 {
			userIdList = append(userIdList, "0")
		}
	}

	var interestIdList []int64
	if request.InterestName != "" || request.Default != 0 {
		var mapData = make(map[string]string, 0)
		if request.InterestName != "" {
			mapData["name"] = request.InterestName
		}
		if request.Default != 0 {
			mapData["default"] = strconv.Itoa(int(request.Default))
		}
		where, intCount, err := imdb.GetInterestsByWhere(mapData, 0, 0, "")
		if intCount == 0 || err != nil {
			interestIdList = append(interestIdList, 0)
		}
		for _, v := range where {
			interestIdList = append(interestIdList, v.Id)
		}
	}

	list, count, err := imdb.ManageGetShortVideoList(userIdList, 0, request.Desc, request.EmptyDesc, 0, 0, 0, request.Pagination.PageNumber, request.Pagination.ShowNumber, interestIdList)

	if err != nil {
		return response, openIMHttp.WrapError(constant.ErrDB)
	}
	response.TotalCount = count

	if len(list) == 0 {
		return response, nil
	}

	userIdList = []string{}
	for _, v := range list {
		userIdList = append(userIdList, v.UserId)
	}

	if len(userIdList) == 0 {
		return response, nil
	}

	// user info
	connection, err := createUserGRPConnection(request.OperationID)
	if err != nil {
		return response, nil
	}

	rpcRequest := &pbUser.GetUserInfoReq{
		UserIDList: userIdList,
	}
	userInfo, err := connection.GetUserInfo(ctx, rpcRequest)
	if err != nil {
		return response, nil
	}
	userInfoMap := make(map[string]*sdkws.UserInfo)
	for _, user := range userInfo.UserInfoList {
		userInfoMap[user.UserID] = user
	}

	// interest info
	interestAllList := imdb.GetAllInterestType()
	var interestTypeMap = make(map[int64]*db.InterestTypeRes, len(interestAllList))
	for index, interestType := range interestAllList {
		interestTypeMap[interestType.Id] = interestAllList[index]
	}

	for _, v := range list {
		vv := &pbAdminCMS.ShortVideoInterestLabelMessage{
			Id:                  v.Id,
			FileId:              v.FileId,
			MediaUrl:            v.MediaUrl,
			CoverUrl:            v.CoverUrl,
			UserId:              v.UserId,
			UserName:            v.UserId,
			InterestId:          v.InterestId,
			InterestChineseName: []string{},
			InterestEnglishName: []string{},
			InterestArabicName:  []string{},
		}

		if userInfoMap[v.UserId] != nil {
			vv.UserName = userInfoMap[v.UserId].Nickname
		}

		if v.InterestId != "" {
			stringArr := strings.Split(v.InterestId, ",")
			for _, v1 := range stringArr {
				vv.InterestIdList = append(vv.InterestIdList, utils.StringToInt64(v1))

				if interestTypeMap[utils.StringToInt64(v1)] != nil {
					for _, v2 := range interestTypeMap[utils.StringToInt64(v1)].Name {
						switch v2.LanguageType {
						case "en":
							vv.InterestEnglishName = append(vv.InterestEnglishName, v2.Name)
						case "cn":
							vv.InterestChineseName = append(vv.InterestChineseName, v2.Name)
						case "ar":
							vv.InterestArabicName = append(vv.InterestArabicName, v2.Name)
						}
					}
				}
			}
		}

		response.ShortVideoInterestLabel = append(response.ShortVideoInterestLabel, vv)
	}

	return response, nil
}

func (s *adminCMSServer) AlterShortVideoInterestLabel(ctx context.Context, request *pbAdminCMS.AlterShortVideoInterestLabelRequest) (*pbAdminCMS.AlterShortVideoInterestLabelResponse, error) {
	log.NewInfo(request.OperationID, "AlterShortVideoInterestLabel", request)
	response := &pbAdminCMS.AlterShortVideoInterestLabelResponse{}
	response.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}

	if request.FileId == "" || len(request.InterestIdList) == 0 {
		response.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		response.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg
		return response, nil
	}

	// check file id
	info, err := imdb.GetShortVideoByFileId(request.FileId)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		response.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg
		return response, nil
	}

	// check interest id
	interestInfo := imdb.GetInterestsByIds(request.InterestIdList)
	if len(interestInfo) == 0 {
		response.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		response.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg
		return response, nil
	}

	var interestIdList []int64
	for i, _ := range interestInfo {
		interestIdList = append(interestIdList, i)
	}
	interestIdStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(interestIdList)), ","), "[]")

	err = imdb.UpdateShortVideoInterestByFileId(request.FileId, interestIdStr)
	if err != nil {
		response.CommonResp.ErrCode = constant.ErrDB.ErrCode
		response.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return response, nil
	}

	// gorse
	utils2.UpdateItem(request.FileId, info.Status != constant.ShortVideoTypeNormal, constant.ShortString, info.Desc, strings.Split(interestIdStr, ","))

	return response, nil
}

func (s *adminCMSServer) GetShortVideoCommentReplies(_ context.Context, req *pbAdminCMS.GetShortVideoCommentRepliesReq) (*pbAdminCMS.GetShortVideoCommentRepliesResp, error) {
	resp := &pbAdminCMS.GetShortVideoCommentRepliesResp{}
	resp.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	resp.Pagination = &sdkws.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	resp.RepliesCount = 0

	where := map[string]string{}
	where["privacy"] = fmt.Sprintf("%d", req.Privacy)
	where["desc"] = req.Content
	where["content_type"] = fmt.Sprintf("%d", req.IsEmpty)
	where["comment_user"] = req.CommentUser
	where["comment"] = req.Comment
	where["reply_user"] = req.ReplyUser
	where["reply_content"] = req.ReplyContent
	where["publisher"] = req.Publisher
	where["comment_id"] = fmt.Sprintf("%d", req.CommentId)
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	dbResList, count, err := imdb.GetShortVideoCommentRepliesByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetMoments failed", err.Error())
		return resp, err
	}

	log.Debug("", "dbResList len: ", len(dbResList))
	_ = utils.CopyStructFields(&resp.CommentReplies, dbResList)

	resp.RepliesCount = count
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (s *adminCMSServer) AlterReply(_ context.Context, req *pbAdminCMS.AlterReplyReq) (*pbAdminCMS.AlterReplyResp, error) {

	resp := &pbAdminCMS.AlterReplyResp{CommonResp: &pbAdminCMS.CommonResp{}}

	shortVideo, err := imdb.GetShortVideoByFileId(req.ShortVideoId)
	if err != nil {
		log.Error("GetShortVideoByFileId error", err.Error())
		return resp, err
	}

	replyComment, err := imdb.GetShortVideoCommentByCommentId(req.ReplyCommentId)
	if err != nil {
		log.Error("GetCommentByCommentID error", err.Error())
		return resp, err
	}
	shortVideo.Desc = req.Content
	replyComment.Content = req.ReplyContent
	replyComment.Remark = req.Remark

	shortVideo.UpdateTime = time.Now().Unix()
	replyComment.UpdateTime = shortVideo.UpdateTime

	err = imdb.UpdateShortVideoInfoByFileId(req.ShortVideoId, shortVideo)
	if err != nil {
		log.Error("GetCommentByCommentID error", err.Error())
		resp.CommonResp.ErrMsg = "GetCommentByCommentID error"
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, err
	}

	raw, err := imdb.UpdateShortVideoCommentReply(replyComment)
	if err != nil {
		log.Error("GetCommentByCommentID replyComment error", err.Error())
		resp.CommonResp.ErrMsg = "GetCommentByCommentID replyComment error"
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, err
	}
	if raw == 0 {
		log.Error("GetCommentByCommentID replyComment update is zero")
		resp.CommonResp.ErrMsg = "GetCommentByCommentID replyComment error"
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, err
	}

	resp.CommonResp.ErrMsg = constant.OK.ErrMsg
	resp.CommonResp.ErrCode = constant.OK.ErrCode

	return resp, nil

}

func (s *adminCMSServer) DeleteReplies(_ context.Context, req *pbAdminCMS.DeleteRepliesReq) (*pbAdminCMS.DeleteRepliesResp, error) {
	resp := &pbAdminCMS.DeleteRepliesResp{CommonResp: &pbAdminCMS.CommonResp{}}

	list, err := imdb.GetShortVideoCommentByCommentIdList(req.CommentIds)
	if err != nil {
		resp.CommonResp.ErrMsg = "delete success"
		resp.CommonResp.ErrCode = constant.OK.ErrCode
		return resp, nil
	}

	for _, commentInfo := range list {
		allCommentInfoList := make([]*db.ShortVideoComment, 0)
		allCommentInfoList = append(allCommentInfoList, commentInfo)
		childCommentInfoList, _ := imdb.GetAllReplyCommentIdByCommentId(commentInfo.CommentId)
		if len(childCommentInfoList) > 0 {
			childCommentInfoList, _ := imdb.GetCommentInfoListByCommentId(childCommentInfoList)
			for _, childCommentInfo := range childCommentInfoList {
				if childCommentInfo.Status == constant.ShortVideoCommentStatusNormal || childCommentInfo.Status == constant.ShortVideoCommentStatusAudit {
					allCommentInfoList = append(allCommentInfoList, childCommentInfo)
				}
			}
		}

		// mysql
		err = imdb.DeleteShortVideoCommentByCommentList(allCommentInfoList)
		if err == nil {
			// count
			for _, commentInfo := range allCommentInfoList {
				imdb.IncrShortVideoUserCountByUserId(commentInfo.UserID, "comment_num", -1)
			}
			db.DB.ShortVideoDelCommentListCache(commentInfo.FileId + "*")
		}
	}

	resp.CommonResp.ErrMsg = "delete success"
	resp.CommonResp.ErrCode = constant.OK.ErrCode
	return resp, nil
}

func (s *adminCMSServer) GetShortVideoCommentLikes(_ context.Context, req *pbAdminCMS.GetShortVideoCommentLikesReq) (*pbAdminCMS.GetShortVideoCommentLikesResp, error) {
	resp := &pbAdminCMS.GetShortVideoCommentLikesResp{}
	resp.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	resp.Pagination = &sdkws.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	resp.LikesCount = 0

	where := map[string]string{}
	where["privacy"] = fmt.Sprintf("%d", req.Privacy)
	where["desc"] = req.Content
	where["content_type"] = fmt.Sprintf("%d", req.IsEmpty)
	where["comment_user"] = req.CommentUser
	where["comment"] = req.Comment
	where["reply_user"] = req.ReplyUser
	where["publisher"] = req.Publisher
	where["reply_content"] = req.ReplyContent
	where["comment_id"] = fmt.Sprintf("%d", req.CommentID)
	where["like_user"] = req.LikeUser
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	dbResList, count, err := imdb.GetShortVideoCommentLikesByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetMoments failed", err.Error())
		return resp, err
	}

	log.Debug("", "dbResList len: ", len(dbResList))
	_ = utils.CopyStructFields(&resp.CommentLikes, dbResList)

	resp.LikesCount = count
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (s *adminCMSServer) AlterLike(_ context.Context, req *pbAdminCMS.AlterLikeReq) (*pbAdminCMS.AlterLikeResp, error) {

	resp := &pbAdminCMS.AlterLikeResp{CommonResp: &pbAdminCMS.CommonResp{}}

	shortVideo, err := imdb.GetShortVideoByFileId(req.ShortVideoID)
	if err != nil {
		log.Error("GetShortVideoByFileId error", err.Error())
		return resp, err
	}

	likes, err := imdb.GetShortVideoCommentLike(req.LikeId)
	if err != nil {
		log.Error("GetShortVideoLikeByLikeIdList error", err.Error())
		return resp, err
	}
	shortVideo.Desc = req.Content
	likes.Remark = req.Remark

	err = imdb.UpdateShortVideoInfoByFileId(req.ShortVideoID, shortVideo)
	if err != nil {
		log.Error("UpdateShortVideoInfoByFileId error", err.Error())
		return resp, err
	}

	err = imdb.UpdateShortVideoCommentLike(&likes)
	if err != nil {
		log.Error("UpdateShortVideoCommentLike error", err.Error())
		return resp, err
	}

	resp.CommonResp.ErrMsg = constant.OK.ErrMsg
	resp.CommonResp.ErrCode = constant.OK.ErrCode

	return resp, nil
}

func (s *adminCMSServer) DeleteLikes(_ context.Context, req *pbAdminCMS.DeleteLikesReq) (*pbAdminCMS.DeleteLikesResp, error) {
	resp := &pbAdminCMS.DeleteLikesResp{CommonResp: &pbAdminCMS.CommonResp{}}

	fmt.Println(req.Likes)
	for _, like := range req.Likes {
		commentLike, err := imdb.GetShortVideoCommentLike(like)
		if err != nil {
			log.Error("GetShortVideoCommentLike error", err.Error())
			fmt.Println("GetShortVideoCommentLike error", err.Error())
			return resp, err
		}

		// redis
		if db.DB.ShortVideoCancelLike(commentLike.UserId, commentLike.FileId) == false {
			resp.CommonResp = &pbAdminCMS.CommonResp{
				ErrCode: constant.ErrLikeRedisFail.ErrCode,
				ErrMsg:  constant.ErrLikeRedisFail.ErrMsg,
			}
			return resp, nil
		}

		if !imdb.DeleteShortVideoCommentLike(commentLike.FileId, commentLike.UserId, commentLike.CommentId) {
			log.Error("DeleteShortVideoCommentLike error")
		}
	}

	resp.CommonResp.ErrMsg = constant.OK.ErrMsg
	resp.CommonResp.ErrCode = constant.OK.ErrCode

	return resp, nil
}

func (s *adminCMSServer) GetFollowers(_ context.Context, req *pbAdminCMS.GetFollowersReq) (*pbAdminCMS.GetFollowersResp, error) {
	resp := &pbAdminCMS.GetFollowersResp{}
	resp.CommonResp = &pbAdminCMS.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	resp.Pagination = &sdkws.ResponsePagination{
		CurrentPage: req.Pagination.PageNumber,
		ShowNumber:  req.Pagination.ShowNumber,
	}
	resp.FollowersCount = 0

	where := map[string]string{}
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["follower"] = req.Follower
	where["followed_user"] = req.FollowedUser

	followers, followerCount, err := imdb.GetFollowersByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetMoments failed", err.Error())
		return resp, err
	}

	log.Debug("", "dbResList len: ", len(followers))
	_ = utils.CopyStructFields(&resp.Followers, followers)

	resp.FollowersCount = followerCount
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (s *adminCMSServer) AlterFollower(_ context.Context, req *pbAdminCMS.AlterFollowerReq) (*pbAdminCMS.AlterFollowerResp, error) {
	resp := &pbAdminCMS.AlterFollowerResp{CommonResp: &pbAdminCMS.CommonResp{}}
	follower := db.ShortVideoFollow{}
	follower.Id = req.Id
	follower.Remark = req.Remark

	var row int64
	var err error
	if row, err = imdb.AlterFollower(&follower); err != nil {
		log.Error("AlterFollower error", err.Error())
		return resp, err
	}
	if row == 0 {
		log.Error("AlterFollower row is zero")
		return resp, err
	}

	resp.CommonResp.ErrMsg = constant.OK.ErrMsg
	resp.CommonResp.ErrCode = constant.OK.ErrCode

	return resp, nil
}

func (s *adminCMSServer) DeleteFollowers(_ context.Context, req *pbAdminCMS.DeleteFollowersReq) (*pbAdminCMS.DeleteFollowersResp, error) {
	resp := &pbAdminCMS.DeleteFollowersResp{CommonResp: &pbAdminCMS.CommonResp{}}

	followers, err := imdb.GetFollowersByIdList(req.Id)
	if err != nil {
		log.Error("GetFollowersByIdList error", err.Error())
		return resp, err
	}

	for _, follower := range followers {
		result := imdb.DeleteShortVideoFollow(follower.UserId, follower.FansId)
		if !result {
			resp.CommonResp = &pbAdminCMS.CommonResp{
				ErrCode: constant.ErrFollowMysqlFail.ErrCode,
				ErrMsg:  constant.ErrFollowMysqlFail.ErrMsg,
			}
			return resp, nil
		}
	}

	go func() {
		for _, follower := range followers {
			db.DB.ShortVideoFollowRemMember(follower.FansId, follower.UserId)

			// gorse
			utils2.InsertUser(follower.UserId, []string{}, []string{})
			followList, _ := db.DB.ShortVideoGetFollowUserListCache(follower.FansId)
			utils2.PatchUser(follower.FansId, imdb.GetStringListByUserId(follower.FansId), followList)

			// notice
			insertData := db.ShortVideoNotice{
				UserId:       follower.UserId,
				SourceUserId: follower.FansId,
				Type:         constant.ShortVideoNoticeTypeFollowMe,
			}
			result := imdb.AddShortVideoNotice(insertData)
			if result != 0 {
				db.DB.AddShortVideoNotice(follower.UserId, result)
			}

			// count
			imdb.IncrShortVideoUserCountByUserId(follower.FansId, "follow_num", -1)
			imdb.IncrShortVideoUserCountByUserId(follower.UserId, "fans_num", -1)
		}

	}()

	resp.CommonResp.ErrMsg = constant.OK.ErrMsg
	resp.CommonResp.ErrCode = constant.OK.ErrCode

	return resp, nil
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
