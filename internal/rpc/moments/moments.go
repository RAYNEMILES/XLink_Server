package moments

import (
	"Open_IM/internal/api/short_video"
	"Open_IM/internal/rpc/admin_cms"
	"Open_IM/internal/rpc/msg"
	"Open_IM/internal/rpc/news"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbMoments "Open_IM/pkg/proto/moments"
	pbNews "Open_IM/pkg/proto/news"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	pbShortVideo "Open_IM/pkg/proto/short_video"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net"
	"strconv"
	"strings"
	"time"

	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"

	"Open_IM/pkg/common/config"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type rpcMoments struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewRpcMomentsServer(port int) *rpcMoments {
	return &rpcMoments{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImMoemntsName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (rpc *rpcMoments) Run() {
	log.NewPrivateLog(constant.OpenImMomentsLog)
	log.NewInfo("0", "rpc moments start...")

	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(rpc.rpcPort)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + rpc.rpcRegisterName)
	}
	log.NewInfo("0", "listen network success, ", address, listener)
	//grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()

	//service registers with etcd
	pbMoments.RegisterMomentsServer(srv, rpc)
	rpcRegisterIP := ""
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(),
			rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)
		return
	}
	log.NewInfo("0", "RegisterMomentsServer ok ", rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "rpc Moments ok")
}

func checkMomentCreatorIsFriend(momentIDStr, opUserID string) (*pbMoments.CommonResp, *db.Moment, error) {
	momentIDPrimitive, _ := primitive.ObjectIDFromHex(momentIDStr)
	moment, errM := db.DB.GetMoment(momentIDPrimitive)
	if errM != nil {
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Moment Not found in Mongo DB"}, nil, errors.New("moment not found in mongoDB")
	}
	momentOwnerID := moment.CreatorID
	if opUserID != momentOwnerID {
		if !checkUsersAreFriend(momentOwnerID, opUserID) || checkUsersAreInBlackList(momentOwnerID, opUserID) {
			return &pbMoments.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "you are not allowed to perform this action"}, nil, errors.New("you are not allowed to perform this action")
		}
	}
	return nil, moment, nil
}

func checkMomentCreatorIsFrnInReqFrnLst(momentIDStr, opUserID string) (*pbMoments.CommonResp, *db.Moment, error) {
	momentIDPrimitive, _ := primitive.ObjectIDFromHex(momentIDStr)
	moment, errM := db.DB.GetMoment(momentIDPrimitive)
	if errM != nil {
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Moment Not found in Mongo DB"}, nil, errors.New("moment not found in mongoDB")
	}
	momentOwnerID := moment.CreatorID
	if opUserID != momentOwnerID {
		if !checkUsersAreFriend(opUserID, momentOwnerID) {
			return &pbMoments.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "you are not allowed to perform this action"}, nil, errors.New("you are not allowed to perform this action")
		}
	}
	return nil, moment, nil
}

func checkCommentCreatorIsFriend(momentIDStr, commentID, opUserID string) (*pbMoments.CommonResp, error) {
	//Check Moment Owner ID
	momentIDPrimitive, _ := primitive.ObjectIDFromHex(momentIDStr)
	moment, errM := db.DB.GetMoment(momentIDPrimitive)
	if errM != nil {
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Moment Not found in Mongo DB"}, errM
	}
	momentOwnerID := moment.CreatorID
	if opUserID != momentOwnerID {
		if !checkUsersAreFriend(momentOwnerID, opUserID) {
			return &pbMoments.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "you are not allowed to perform this action"}, errors.New("you are not allowed to perform this action")
		}
	}
	//Check Comment Owner ID
	commentIDPrimitive, _ := primitive.ObjectIDFromHex(commentID)
	comment, errM := db.DB.GetMomentComment(commentIDPrimitive)
	if errM != nil {
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Comment Not found in Mongo DB"}, errors.New("comment not found in mongoDB")
	}
	commentOwnerID := comment.UserID
	if opUserID != commentOwnerID {
		if !checkUsersAreFriend(commentOwnerID, opUserID) {
			return &pbMoments.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "you are not allowed to perform this action"}, errors.New("you are not allowed to perform this action")
		}
	}

	return nil, nil
}

func checkUsersAreFriend(userId, friendID string) bool {
	if userId == friendID {
		return true
	}
	myFriendIDListRedis, _ := db.DB.GetFriendIDListFromCache(userId)
	isPass := utils.IsContain(friendID, myFriendIDListRedis)
	return isPass
}
func checkUsersAreInBlackList(userId, friendID string) bool {
	if userId == friendID {
		return true
	}
	blockedUsersList, _ := db.DB.GetBlackListForMomentFromCache(userId)
	isPass := utils.IsContain(friendID, blockedUsersList)
	return isPass
}

// CreateMoment RPC method for creating Moment, will store Moment in NOSQL and SQL at both time
func (rpc *rpcMoments) CreateMoment(c context.Context, req *pbMoments.Moment) (*pbMoments.CreateMomentResponse, error) {

	var moment db.Moment
	var momentSQL db.MomentSQL
	req.Status = 1
	req.Privacy = 1
	req.MCreateTime = time.Now().Unix()

	// no original moment, moment can't be reposted, if add moment reposted, need set original id here, and background management for getting.
	err := utils.CopyStructFields(&moment, req)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	MContentImagesArrayV2Bytes, err := json.Marshal(req.MContentImagesArrayV2)
	if err == nil {
		moment.MContentImagesArray = string(MContentImagesArrayV2Bytes)
	} else {
		moment.MContentImagesArray = ""
	}
	MContentVideosArrayV2Bytes, err := json.Marshal(req.MContentVideosArrayV2)
	if err == nil {
		moment.MContentVideosArray = string(MContentVideosArrayV2Bytes)
	} else {
		moment.MContentVideosArray = ""
	}
	user, err := imdb.GetUserByUserID(req.CreatorID)
	if err != nil {
		log.NewError("GetUserByUserID failed ", err.Error(), req.CreatorID)
	} else {
		moment.UserID = user.UserID
		moment.UserName = user.Nickname
		moment.UserProfileImg = user.FaceURL
	}

	err = utils.CopyStructFields(&momentSQL, moment)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	go func() {
		// remove new files tags
		var media []string
		if len(req.MContentVideosArrayV2) > 0 {
			for _, object := range req.MContentVideosArrayV2 {
				media = append(media, object.VideoUrl, object.SnapShotUrl)
			}
		}
		if len(req.MContentImagesArrayV2) > 0 {
			for _, object := range req.MContentImagesArrayV2 {
				media = append(media, object.ImageUrl)
			}
		}
		client, err := admin_cms.GetTencentCloudClient(true)
		failedList, err := admin_cms.RemoveDeleteTagForPersistent(client, media)
		if err != nil {
			log.NewError("", utils.GetSelfFuncName(), "upload file error")
		}
		if len(failedList) != 0 {
			log.NewError("", utils.GetSelfFuncName(), "upload file error, failed list: ", failedList)
		}
	}()

	mongoID := primitive.NewObjectIDFromTimestamp(time.Now())
	moment.MomentID = mongoID
	momentSQL.MomentID = mongoID.Hex()
	moment.CommentCtl = 1
	moment.Status = 1
	//insert in Mongo DB
	err = db.DB.AddMoment(moment)
	if err != nil {
		errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.CreateMomentResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	//insert in MYSQL DB
	go func() {
		err = imdb.InsertMoment(momentSQL)
		if err != nil {
			log.NewError("CreateMoment", "insert in db failed", err.Error())
		}
		if moment.ArticleID != 0 {
			news.UpdateArticleCount(true, true, moment.ArticleID, 0, 0, 0, 1, 0)
		}
	}()

	req.UserID = moment.UserID
	req.UserName = moment.UserName
	req.UserProfileImg = moment.UserProfileImg
	req.MomentID = moment.MomentID.Hex()
	if req.ArticleID != 0 {
		article, err := db.DB.GetArticlesByID(req.ArticleID)
		if article != nil && err == nil {
			req.ArticleDetailsInMoment = &pbMoments.ArticleDetailsInMoment{Article: &pbNews.UserArticle{}, Official: &pbNews.UserFollow{}}
			_ = utils.CopyStructFields(req.ArticleDetailsInMoment.Article, article)
			req.ArticleDetailsInMoment.Official.OfficialID = article.OfficialID
			req.ArticleDetailsInMoment.Official.Nickname = article.OfficialName
			req.ArticleDetailsInMoment.Official.FaceURL = article.OfficialProfileImg
		}
	}
	if req.WoomFileID != "" {
		var rpcRequest pbShortVideo.GetShortVideoByFileIdRequest
		connection, err := short_video.CreateShortVideoGRPConnection("tempOperationID")
		if err == nil {
			rpcRequest.OperationID = "tempOperationID"
			rpcRequest.UserId = req.UserID
			rpcRequest.FileId = req.WoomFileID
			rpcResponse, err := connection.GetShortVideoByFieldId(context.Background(), &rpcRequest)
			if err == nil && rpcResponse != nil {
				req.WoomDetails = &pbShortVideo.ShortVideoInfo{}
				utils.CopyStructFields(req.WoomDetails, rpcResponse.ShortVideoInfo)
			}
		}
		_ = imdb.IncrShortVideoForwardNum(req.WoomFileID)
	}
	return &pbMoments.CreateMomentResponse{ErrCode: 0, ErrMsg: "Moment added", Moment: req}, nil

}

// CreateMomentLike RPC method for like Moment, will store Moment in NOSQL and SQL at both time
func (rpc *rpcMoments) CreateMomentLike(c context.Context, req *pbMoments.MomentLike) (*pbMoments.CommonResp, error) {

	comResp, _, errRes := checkMomentCreatorIsFriend(req.MomentID, req.CreatorID)
	if errRes != nil {
		return comResp, errRes
	}

	var momentLike db.MomentLike
	var momentLikeSQL db.MomentLikeSQL
	req.Status = 1
	err := utils.CopyStructFields(&momentLike, req)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	user, err := imdb.GetUserByUserID(req.CreatorID)
	if err != nil {
		log.NewError("GetUserByUserID failed ", err.Error(), req.CreatorID)
	} else {
		momentLike.UserID = user.UserID
		momentLike.UserName = user.Nickname
		momentLike.UserProfileImg = user.FaceURL
		momentLike.CreateTime = time.Now().Unix()
		momentLike.CreateBy = user.UserID
	}

	err = utils.CopyStructFields(&momentLikeSQL, momentLike)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	momentLike.MomentID, _ = primitive.ObjectIDFromHex(req.MomentID)
	momentLikeSQL.MomentID = req.MomentID
	//insert in Mongo DB
	err = db.DB.AddMomentLike(momentLike)
	if err != nil {
		errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	//insert in MYSQL DB
	go func() {
		err = imdb.InsertMomentLike(momentLikeSQL)
		if err != nil {
			log.NewError("CreateMoment", "insert in db failed", err.Error())
		}
		// update moment like counts
		updateMomentCounts(req.MomentID, 1, 0, 0)
	}()

	go msg.SendMomentInteractionNotification("MomentNotification_Action_Like"+req.MomentID, req.MomentID, momentLike.UserID, constant.MomentNotification_Action_Like, momentLike)

	return &pbMoments.CommonResp{ErrCode: 0, ErrMsg: "Moment lilked"}, nil

}

// CancelMomentLike RPC method for like Moment, will store Moment in NOSQL and SQL at both time
func (rpc *rpcMoments) CancelMomentLike(c context.Context, req *pbMoments.MomentCancelLike) (*pbMoments.CommonResp, error) {
	var momentLike db.MomentLike
	var momentLikeSQL db.MomentLikeSQL
	err := utils.CopyStructFields(&momentLike, req)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	momentLike.CreateTime = time.Now().Unix()
	momentLike.CreateBy = req.CreatorID

	err = utils.CopyStructFields(&momentLikeSQL, momentLike)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	momentLike.MomentID, _ = primitive.ObjectIDFromHex(req.MomentID)
	momentLikeSQL.MomentID = req.MomentID
	//fetched before delete for notification only
	momentCancelLike, err := db.DB.GetMomentLikeByID(momentLike.MomentID, momentLike.CreateBy)
	if err != nil {
		errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, err
	}
	//insert in Mongo DB
	err = db.DB.CancelMomentLike(momentLike)
	if err != nil {
		errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, err
	}
	//insert in MYSQL DB
	go func() {
		err = imdb.CancelMomentLike(momentLikeSQL)
		if err != nil {
			log.NewError("CreateMoment", "insert in db failed", err.Error())
		}
		// update moment like counts
		updateMomentCounts(req.MomentID, -1, 0, 0)
	}()

	if momentCancelLike != nil {
		go msg.SendMomentInteractionNotification("MomentNotification_Action_CancelLike"+req.MomentID, req.MomentID, momentLike.UserID, constant.MomentNotification_Action_CancelLike, *momentCancelLike)
	}
	return &pbMoments.CommonResp{ErrCode: 0, ErrMsg: "Moment like canceled"}, nil

}

// CreateMomentComment RPC method for creating Comment on Moment, will store Moment in NOSQL and SQL at both time
func (rpc *rpcMoments) CreateMomentComment(c context.Context, req *pbMoments.MomentComment) (*pbMoments.MomentCommentResponse, error) {

	comResp, _, errRes := checkMomentCreatorIsFriend(req.MomentID, req.CreatorID)

	if errRes != nil {
		return &pbMoments.MomentCommentResponse{ErrCode: comResp.ErrCode, ErrMsg: comResp.ErrMsg}, errRes
	}

	var comment db.MomentComment
	var commentSQL db.MomentCommentSQL
	comment.Status = 1
	err := utils.CopyStructFields(&comment, req)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	user, err := imdb.GetUserByUserID(req.CreatorID)
	if err != nil {
		log.NewError("GetUserByUserID failed ", err.Error(), req.CreatorID)
	} else {
		comment.UserID = user.UserID
		comment.UserName = user.Nickname
		comment.UserProfileImg = user.FaceURL
		comment.CreateTime = time.Now().Unix()
		comment.CreateBy = user.UserID
	}

	err = utils.CopyStructFields(&commentSQL, comment)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	comment.MomentID, _ = primitive.ObjectIDFromHex(req.MomentID)
	commentSQL.MomentID = req.MomentID
	mongoID := primitive.NewObjectIDFromTimestamp(time.Now())
	comment.CommentID = mongoID
	commentSQL.CommentID = mongoID.Hex()

	moment, err := db.DB.GetMoment(comment.MomentID)
	if err != nil {
		errMsg := "get moment failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.MomentCommentResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, err
	}

	if moment.CommentCtl != 1 {
		errMsg := "Moment has closed the comment."
		log.Error(errMsg)
		return &pbMoments.MomentCommentResponse{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg}, err
	}

	//insert in Mongo DB
	err = db.DB.AddMomentComment(comment)
	if err != nil {
		errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.MomentCommentResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, err
	}
	//insert in MYSQL DB
	go func() {
		err = imdb.InsertMomentComment(commentSQL)
		if err != nil {
			log.NewError("CreateMoment", "insert in db failed", err.Error())
		}

		// increase comment count to mysql
		updateMomentCounts(req.MomentID, 0, 1, 0)
	}()
	commentObj := &pbMoments.MomentCommentResp{}
	commentObj.CommentID = comment.CommentID.Hex()
	commentObj.MomentID = comment.MomentID.Hex()
	commentObj.CommentContent = comment.CommentContent
	commentObj.CreateBy = comment.CreateBy
	commentObj.CreateTime = comment.CreateTime
	commentObj.Status = int32(comment.Status)
	commentObj.UserID = comment.UserID
	commentObj.UserName = comment.UserName
	commentObj.UserProfileImg = comment.UserProfileImg
	commentObj.AccountStatus = int32(comment.AccountStatus)

	go msg.SendMomentInteractionNotification("MomentNotification_Action_Comment"+req.MomentID, req.MomentID, comment.UserID, constant.MomentNotification_Action_Comment, comment)

	return &pbMoments.MomentCommentResponse{ErrCode: 0, ErrMsg: "Moment comment added", Comment: commentObj}, nil
}

// CreateReplyOfMomentComment RPC method for creating reply of a Comment on Moment, will store Moment in NOSQL and SQL at both time
func (rpc *rpcMoments) CreateReplyOfMomentComment(c context.Context, req *pbMoments.ReplyOfMomentComment) (*pbMoments.MomentCommentResponse, error) {

	log.Debug("", "CreateReplyOfMomentComment req: ", req)

	comResp, errRes := checkCommentCreatorIsFriend(req.MomentID, req.CommentID, req.CreatorID)
	if errRes != nil {
		return &pbMoments.MomentCommentResponse{ErrCode: comResp.ErrCode, ErrMsg: comResp.ErrMsg}, errRes
	}

	var comment db.MomentComment
	var commentSQL db.MomentCommentSQL
	comment.Status = 1
	err := utils.CopyStructFields(&comment, req)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	user, err := imdb.GetUserByUserID(req.CreatorID)
	if err != nil {
		log.NewError("GetUserByUserID failed ", err.Error(), req.CreatorID)
	} else {
		comment.UserID = user.UserID
		comment.UserName = user.Nickname
		comment.UserProfileImg = user.FaceURL
		comment.CreateTime = time.Now().Unix()
		comment.CreateBy = user.UserID
	}
	commentParentObjectID, _ := primitive.ObjectIDFromHex(req.CommentID)
	momentComment, err := db.DB.GetMomentComment(commentParentObjectID)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "Parent Comment not found", err.Error())
		return nil, err
	}
	err = utils.CopyStructFields(&commentSQL, comment)
	if err != nil {
		log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	if momentComment != nil {
		comment.CPUserName = momentComment.UserName
		comment.CPUserID = momentComment.UserID
		comment.CPUserProfileImg = momentComment.UserProfileImg
		// if parent haven't reply comment, then set reply to parent
		if momentComment.ReplyCommentID.IsZero() {
			comment.ReplyCommentID = commentParentObjectID
			commentSQL.ReplyCommentID = req.CommentID
		} else {
			// if parent have reply comment, than set the reply comment id to reply
			comment.ReplyCommentID = momentComment.ReplyCommentID
			commentSQL.ReplyCommentID = momentComment.ReplyCommentID.Hex()
		}
	}

	comment.MomentID, _ = primitive.ObjectIDFromHex(req.MomentID)
	commentSQL.MomentID = req.MomentID

	comment.CommentParentID = commentParentObjectID
	commentSQL.CommentParentID = req.CommentID

	mongoID := primitive.NewObjectIDFromTimestamp(time.Now())
	comment.CommentID = mongoID
	commentSQL.CommentID = mongoID.Hex()

	//insert in Mongo DB
	err = db.DB.AddMomentComment(comment)
	if err != nil {
		errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
		log.Error(errMsg)
		return &pbMoments.MomentCommentResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, err
	}
	//insert in MYSQL DB
	go func() {
		err = imdb.InsertMomentComment(commentSQL)
		if err != nil {
			log.NewError("CreateMoment", "insert in db failed", err.Error())
		}

		updateMomentCommentCounts(commentSQL.ReplyCommentID, 0, 1)
	}()

	commentObj := &pbMoments.MomentCommentResp{}
	commentObj.CommentID = comment.CommentID.Hex()
	commentObj.MomentID = comment.MomentID.Hex()
	commentObj.CommentParentID = comment.CommentParentID.Hex()
	commentObj.CPUserID = comment.CPUserID
	commentObj.CPUserName = comment.CPUserName
	commentObj.CPUserProfileImg = comment.CPUserProfileImg
	commentObj.CommentContent = comment.CommentContent
	commentObj.CreateBy = comment.CreateBy
	commentObj.CreateTime = comment.CreateTime
	commentObj.Status = int32(comment.Status)
	commentObj.UserID = comment.UserID
	commentObj.UserName = comment.UserName
	commentObj.UserProfileImg = comment.UserProfileImg
	commentObj.AccountStatus = int32(comment.AccountStatus)

	go msg.SendMomentInteractionNotification("MomentCommentNotification"+req.MomentID, req.MomentID, comment.UserID, constant.MomentNotification_Action_CommentReply, comment)

	return &pbMoments.MomentCommentResponse{ErrCode: 0, ErrMsg: "Moment comment added", Comment: commentObj}, nil
}

// GetListHomeTimeLineOfMoments RPC method for creating reply of a Comment on Moment, will store Moment in NOSQL and SQL at both time
func (rpc *rpcMoments) GetListHomeTimeLineOfMoments(c context.Context, req *pbMoments.ListHomeTimeLineOfMomentsReq) (*pbMoments.ListHomeTimeLineOfMoments, error) {

	var listHomeTimeLineOfMoments = pbMoments.ListHomeTimeLineOfMoments{}
	var myFriendIDList []string
	myFriendIDListTemp, _ := db.DB.GetFriendIDListFromCache(req.CreatorID)
	blockedUsersList, _ := db.DB.GetBlackListForMomentFromCache(req.CreatorID)
	for _, friendID := range myFriendIDListTemp {
		if !utils.IsContain(friendID, blockedUsersList) {
			user, _ := imdb.GetUserByUserIDEvenDeleted(friendID)
			if user != nil {
				myFriendIDList = append(myFriendIDList, friendID)
			}
		}
	}
	myFriendIDList = append(myFriendIDList, req.CreatorID)
	log.Error(utils.GetSelfFuncName(), "Friend List for Moments fetch ", myFriendIDList)
	moments, err := db.DB.GetMomentsByFriendList(myFriendIDList, req.CreatorID, req.PageNumber, req.MomentLimit)
	log.Error(utils.GetSelfFuncName(), "Moments Fetched ", moments)
	if err == nil {
		// listHomeTimeLineOfMoments.HomeTimeLineOfMoments
		for _, moment := range moments {
			var timelineMoment = pbMoments.HomeTimeLineOfMoments{}
			timelineMoment.Moment = &pbMoments.Moment{}
			timelineMoment.Moment.CreatorID = moment.CreatorID
			timelineMoment.Moment.MomentID = moment.MomentID.Hex()
			timelineMoment.Moment.MContentText = moment.MContentText
			timelineMoment.Moment.MContentImagesArray = moment.MContentImagesArray
			timelineMoment.Moment.MContentVideosArray = moment.MContentVideosArray
			timelineMoment.Moment.MContentThumbnilArray = moment.MContentThumbnilArray
			timelineMoment.Moment.MLikesCount = moment.MLikesCount
			timelineMoment.Moment.MCommentsCount = moment.MCommentsCount
			timelineMoment.Moment.MRepostCount = moment.MRepostCount
			timelineMoment.Moment.MCreateTime = moment.MCreateTime
			timelineMoment.Moment.MUpdateTime = moment.MUpdateTime
			timelineMoment.Moment.OrignalCreatorID = moment.OrignalCreatorID
			timelineMoment.Moment.OriginalCreatorName = moment.OriginalCreatorName
			timelineMoment.Moment.OriginalCreatorProfileImg = moment.OriginalCreatorProfileImg
			timelineMoment.Moment.IsReposted = moment.IsReposted
			timelineMoment.Moment.Status = int32(moment.Status)
			timelineMoment.Moment.Privacy = moment.Privacy
			timelineMoment.Moment.UserID = moment.UserID
			timelineMoment.Moment.UserName = moment.UserName
			timelineMoment.Moment.UserProfileImg = moment.UserProfileImg
			timelineMoment.Moment.ArticleID = moment.ArticleID

			in := []byte(moment.MContentImagesArray)
			var MomentImageRequestObjects []*pbMoments.MomentImageRequestObject
			err := json.Unmarshal(in, &MomentImageRequestObjects)
			if err == nil {
				timelineMoment.Moment.MContentImagesArrayV2 = append(timelineMoment.Moment.MContentImagesArrayV2, MomentImageRequestObjects...)
			}

			in = []byte(moment.MContentVideosArray)
			var MomentVideoRequestObjects []*pbMoments.MomentVideoRequestObject
			err = json.Unmarshal(in, &MomentVideoRequestObjects)
			if err == nil {
				timelineMoment.Moment.MContentVideosArrayV2 = append(timelineMoment.Moment.MContentVideosArrayV2, MomentVideoRequestObjects...)
			}

			//TODO Fetch Article if Moment contain Article
			if moment.ArticleID != 0 {
				article, err := db.DB.GetArticlesByID(moment.ArticleID)
				if article != nil && err == nil {
					timelineMoment.Moment.ArticleDetailsInMoment = &pbMoments.ArticleDetailsInMoment{Article: &pbNews.UserArticle{}, Official: &pbNews.UserFollow{}}
					_ = utils.CopyStructFields(timelineMoment.Moment.ArticleDetailsInMoment.Article, article)
					if article.DeleteTime != 0 {
						timelineMoment.Moment.ArticleDetailsInMoment.Article.Content = ""
						timelineMoment.Moment.ArticleDetailsInMoment.Article.TextContent = ""
					}
					timelineMoment.Moment.ArticleDetailsInMoment.Official.OfficialID = article.OfficialID
					timelineMoment.Moment.ArticleDetailsInMoment.Official.Nickname = article.OfficialName
					timelineMoment.Moment.ArticleDetailsInMoment.Official.FaceURL = article.OfficialProfileImg
				}
			}

			//TODO Fetch Woom Details if Moment contain
			if moment.WoomFileID != "" {
				var rpcRequest pbShortVideo.GetShortVideoByFileIdRequest
				connection, err := short_video.CreateShortVideoGRPConnection("tempOperationID")
				if err == nil {
					rpcRequest.OperationID = "tempOperationID"
					rpcRequest.UserId = moment.UserID
					rpcRequest.FileId = moment.WoomFileID
					rpcResponse, err := connection.GetShortVideoByFieldId(context.Background(), &rpcRequest)
					if err == nil && rpcResponse != nil {
						timelineMoment.Moment.WoomDetails = &pbShortVideo.ShortVideoInfo{}
						utils.CopyStructFields(timelineMoment.Moment.WoomDetails, rpcResponse.ShortVideoInfo)
					}
				}
			}

			//TODO Fetch Comments of this moments
			momentsComments, err := db.DB.GetMomentCommentsByMomentAndFriendIds(moment.MomentID, myFriendIDList, 0, req.CommentsLimit)
			if err == nil {
				for _, comment := range momentsComments {
					var MomentCommentResp = pbMoments.MomentCommentResp{}
					MomentCommentResp.MomentID = comment.MomentID.Hex()
					MomentCommentResp.CommentID = comment.CommentID.Hex()
					MomentCommentResp.UserID = comment.UserID
					MomentCommentResp.UserName = comment.UserName
					MomentCommentResp.UserProfileImg = comment.UserProfileImg
					MomentCommentResp.CommentContent = comment.CommentContent
					MomentCommentResp.CommentParentID = comment.CommentParentID.Hex()
					MomentCommentResp.CPUserID = comment.CPUserID
					MomentCommentResp.CPUserName = comment.CPUserName
					MomentCommentResp.CPUserProfileImg = comment.CPUserProfileImg
					MomentCommentResp.CreateBy = comment.CreateBy
					MomentCommentResp.CreateTime = comment.CreateTime
					MomentCommentResp.UpdateBy = comment.UpdateBy
					MomentCommentResp.UpdatedTime = comment.UpdatedTime
					MomentCommentResp.Status = int32(comment.Status)
					MomentCommentResp.AccountStatus = int32(comment.AccountStatus)

					timelineMoment.MomentComments = append(timelineMoment.MomentComments, &MomentCommentResp)
				}
			}

			//TODO Fetch Likes of this Moment
			momentsLikes, err := db.DB.GetMomentLikesByMomentAndFriendIds(moment.MomentID, myFriendIDList)
			if err == nil {
				for _, momentsLike := range momentsLikes {
					var MomentLikeResp = pbMoments.MomentLikeResponse{}
					MomentLikeResp.MomentID = momentsLike.MomentID.Hex()
					MomentLikeResp.UserID = momentsLike.UserID
					MomentLikeResp.UserName = momentsLike.UserName
					MomentLikeResp.UserProfileImg = momentsLike.UserProfileImg
					MomentLikeResp.CreateBy = momentsLike.CreateBy
					MomentLikeResp.CreateTime = momentsLike.CreateTime
					MomentLikeResp.UpdateBy = momentsLike.UpdateBy
					MomentLikeResp.UpdatedTime = momentsLike.UpdatedTime
					MomentLikeResp.Status = int32(momentsLike.Status)
					timelineMoment.MomentLikes = append(timelineMoment.MomentLikes, &MomentLikeResp)
				}
			}

			listHomeTimeLineOfMoments.HomeTimeLineOfMoments = append(listHomeTimeLineOfMoments.HomeTimeLineOfMoments, &timelineMoment)
		}
	} else {
		log.Error(utils.GetSelfFuncName(), "Moments Not Fetched ", err.Error())
	}

	return &listHomeTimeLineOfMoments, nil
}

// GetMomentDetailsByID RPC method for get moment details and 100 comments
func (rpc *rpcMoments) GetMomentDetailsByID(c context.Context, req *pbMoments.GetMomentDetailsByIDRequest) (*pbMoments.GetMomentDetailsByIDResponse, error) {

	var getMomentDetailsByIDResponse = pbMoments.GetMomentDetailsByIDResponse{}
	var err error
	_, moment, err := checkMomentCreatorIsFrnInReqFrnLst(req.MomentID, req.CreatorID)
	if err != nil {
		return &getMomentDetailsByIDResponse, err
	}
	myFriendIDList, _ := db.DB.GetFriendIDListFromCache(req.CreatorID)
	myFriendIDList = append(myFriendIDList, req.CreatorID)
	log.Error(utils.GetSelfFuncName(), "Friend List for Moments fetch ", myFriendIDList)
	momentObjectID, _ := primitive.ObjectIDFromHex(req.MomentID)
	if moment == nil {
		moment, err = db.DB.GetMoment(momentObjectID)
		log.Error(utils.GetSelfFuncName(), "Moments Fetched ", moment)
	}
	if err == nil {
		// listHomeTimeLineOfMoments.HomeTimeLineOfMoments

		pbMomentObj := &pbMoments.Moment{}
		pbMomentObj.CreatorID = moment.CreatorID
		pbMomentObj.MomentID = moment.MomentID.Hex()
		pbMomentObj.MContentText = moment.MContentText
		pbMomentObj.MContentImagesArray = moment.MContentImagesArray
		pbMomentObj.MContentVideosArray = moment.MContentVideosArray
		pbMomentObj.MContentThumbnilArray = moment.MContentThumbnilArray
		pbMomentObj.MLikesCount = moment.MLikesCount
		pbMomentObj.MCommentsCount = moment.MCommentsCount
		pbMomentObj.MRepostCount = moment.MRepostCount
		pbMomentObj.MCreateTime = moment.MCreateTime
		pbMomentObj.MUpdateTime = moment.MUpdateTime
		pbMomentObj.OrignalCreatorID = moment.OrignalCreatorID
		pbMomentObj.OriginalCreatorName = moment.OriginalCreatorName
		pbMomentObj.OriginalCreatorProfileImg = moment.OriginalCreatorProfileImg
		pbMomentObj.IsReposted = moment.IsReposted
		pbMomentObj.Status = int32(moment.Status)
		pbMomentObj.Privacy = moment.Privacy
		pbMomentObj.UserID = moment.UserID
		pbMomentObj.UserName = moment.UserName
		pbMomentObj.UserProfileImg = moment.UserProfileImg
		pbMomentObj.ArticleID = moment.ArticleID

		in := []byte(moment.MContentImagesArray)
		var MomentImageRequestObjects []*pbMoments.MomentImageRequestObject
		err := json.Unmarshal(in, &MomentImageRequestObjects)
		if err == nil {
			pbMomentObj.MContentImagesArrayV2 = append(pbMomentObj.MContentImagesArrayV2, MomentImageRequestObjects...)
		}

		in = []byte(moment.MContentVideosArray)
		var MomentVideoRequestObjects []*pbMoments.MomentVideoRequestObject
		err = json.Unmarshal(in, &MomentVideoRequestObjects)
		if err == nil {
			pbMomentObj.MContentVideosArrayV2 = append(pbMomentObj.MContentVideosArrayV2, MomentVideoRequestObjects...)
		}

		getMomentDetailsByIDResponse.Moment = pbMomentObj

		//TODO Fetch Article if Moment contain Article
		if moment.ArticleID != 0 {
			article, err := db.DB.GetArticlesByID(moment.ArticleID)
			if article != nil && err == nil {
				getMomentDetailsByIDResponse.Moment.ArticleDetailsInMoment = &pbMoments.ArticleDetailsInMoment{Article: &pbNews.UserArticle{}, Official: &pbNews.UserFollow{}}
				_ = utils.CopyStructFields(getMomentDetailsByIDResponse.Moment.ArticleDetailsInMoment.Article, article)
				getMomentDetailsByIDResponse.Moment.ArticleDetailsInMoment.Official.OfficialID = article.OfficialID
				getMomentDetailsByIDResponse.Moment.ArticleDetailsInMoment.Official.Nickname = article.OfficialName
				getMomentDetailsByIDResponse.Moment.ArticleDetailsInMoment.Official.FaceURL = article.OfficialProfileImg

			}
		}

		//TODO Fetch Woom Details if Moment contain
		if moment.WoomFileID != "" {
			var rpcRequest pbShortVideo.GetShortVideoByFileIdRequest
			connection, err := short_video.CreateShortVideoGRPConnection("tempOperationID")
			if err == nil {
				rpcRequest.OperationID = "tempOperationID"
				rpcRequest.UserId = moment.UserID
				rpcRequest.FileId = moment.WoomFileID
				rpcResponse, err := connection.GetShortVideoByFieldId(context.Background(), &rpcRequest)
				if err == nil && rpcResponse != nil {
					getMomentDetailsByIDResponse.Moment.WoomDetails = &pbShortVideo.ShortVideoInfo{}
					utils.CopyStructFields(getMomentDetailsByIDResponse.Moment.WoomDetails, rpcResponse.ShortVideoInfo)
				}
			}
		}

		//TODO Fetch Comments of this moments
		momentsComments, err := db.DB.GetMomentCommentsByMomentAndFriendIds(moment.MomentID, myFriendIDList, 0, 100)
		if err == nil {
			for _, comment := range momentsComments {
				var MomentCommentResp = pbMoments.MomentCommentResp{}
				MomentCommentResp.MomentID = comment.MomentID.Hex()
				MomentCommentResp.CommentID = comment.CommentID.Hex()
				MomentCommentResp.UserID = comment.UserID
				MomentCommentResp.UserName = comment.UserName
				MomentCommentResp.UserProfileImg = comment.UserProfileImg
				MomentCommentResp.CommentContent = comment.CommentContent
				MomentCommentResp.CommentParentID = comment.CommentParentID.Hex()
				MomentCommentResp.CPUserID = comment.CPUserID
				MomentCommentResp.CPUserName = comment.CPUserName
				MomentCommentResp.CPUserProfileImg = comment.CPUserProfileImg
				MomentCommentResp.CreateBy = comment.CreateBy
				MomentCommentResp.CreateTime = comment.CreateTime
				MomentCommentResp.UpdateBy = comment.UpdateBy
				MomentCommentResp.UpdatedTime = comment.UpdatedTime
				MomentCommentResp.Status = int32(comment.Status)
				MomentCommentResp.AccountStatus = int32(comment.AccountStatus)

				getMomentDetailsByIDResponse.MomentComments = append(getMomentDetailsByIDResponse.MomentComments, &MomentCommentResp)
			}
		} else {
			log.Error(utils.GetSelfFuncName(), "Comments Not Fetched ", err.Error())
		}

		//TODO Fetch Likes of this Moment
		momentsLikes, err := db.DB.GetMomentLikesByMomentAndFriendIds(moment.MomentID, myFriendIDList)
		if err == nil {
			for _, momentsLike := range momentsLikes {
				var MomentLikeResp = pbMoments.MomentLikeResponse{}
				MomentLikeResp.MomentID = momentsLike.MomentID.Hex()
				MomentLikeResp.UserID = momentsLike.UserID
				MomentLikeResp.UserName = momentsLike.UserName
				MomentLikeResp.UserProfileImg = momentsLike.UserProfileImg
				MomentLikeResp.CreateBy = momentsLike.CreateBy
				MomentLikeResp.CreateTime = momentsLike.CreateTime
				MomentLikeResp.UpdateBy = momentsLike.UpdateBy
				MomentLikeResp.UpdatedTime = momentsLike.UpdatedTime
				MomentLikeResp.Status = int32(momentsLike.Status)
				getMomentDetailsByIDResponse.MomentLikes = append(getMomentDetailsByIDResponse.MomentLikes, &MomentLikeResp)
			}
		}

	} else {
		log.Error(utils.GetSelfFuncName(), "Moments Not Fetched ", err.Error())
	}

	return &getMomentDetailsByIDResponse, nil
}

// GetMomentCommentsByID RPC method for get moment comments by pagging
func (rpc *rpcMoments) GetMomentCommentsByID(c context.Context, req *pbMoments.GetMomentCommentsByIDRequest) (*pbMoments.GetMomentCommentsByIDResponse, error) {

	var getMomentCommentsByIDResponse = pbMoments.GetMomentCommentsByIDResponse{}

	_, _, errRes := checkMomentCreatorIsFrnInReqFrnLst(req.MomentID, req.CreatorID)
	if errRes != nil {
		return &getMomentCommentsByIDResponse, errRes
	}
	var myFriendIDList []string
	myFriendIDListTemp, _ := db.DB.GetFriendIDListFromCache(req.CreatorID)
	blockedUsersList, _ := db.DB.GetBlackListForMomentFromCache(req.CreatorID)
	for _, friendID := range myFriendIDListTemp {
		if !utils.IsContain(friendID, blockedUsersList) {
			myFriendIDList = append(myFriendIDList, friendID)
		}
	}
	myFriendIDList = append(myFriendIDList, req.CreatorID)
	log.Error(utils.GetSelfFuncName(), "Friend List for Moments fetch ", myFriendIDList)
	momentObjectID, _ := primitive.ObjectIDFromHex(req.MomentID)
	//TODO Fetch Comments of this moments
	momentsComments, err := db.DB.GetMomentCommentsByMomentAndFriendIds(momentObjectID, myFriendIDList, req.PageNumber, req.CommentsLimit)
	if err == nil {
		for _, comment := range momentsComments {
			var MomentCommentResp = pbMoments.MomentCommentResp{}
			MomentCommentResp.MomentID = comment.MomentID.Hex()
			MomentCommentResp.CommentID = comment.CommentID.Hex()
			MomentCommentResp.UserID = comment.UserID
			MomentCommentResp.UserName = comment.UserName
			MomentCommentResp.UserProfileImg = comment.UserProfileImg
			MomentCommentResp.CommentContent = comment.CommentContent
			MomentCommentResp.CommentParentID = comment.CommentParentID.Hex()
			MomentCommentResp.CPUserID = comment.CPUserID
			MomentCommentResp.CPUserName = comment.CPUserName
			MomentCommentResp.CPUserProfileImg = comment.CPUserProfileImg
			MomentCommentResp.CreateBy = comment.CreateBy
			MomentCommentResp.CreateTime = comment.CreateTime
			MomentCommentResp.UpdateBy = comment.UpdateBy
			MomentCommentResp.UpdatedTime = comment.UpdatedTime
			MomentCommentResp.Status = int32(comment.Status)
			MomentCommentResp.AccountStatus = int32(comment.AccountStatus)

			getMomentCommentsByIDResponse.MomentComments = append(getMomentCommentsByIDResponse.MomentComments, &MomentCommentResp)
		}
	} else {
		log.Error(utils.GetSelfFuncName(), "Moments Not Fetched ", err.Error())
	}

	return &getMomentCommentsByIDResponse, nil
}

// RepostAMoment RPC method for reposting a moment of a friend
func (rpc *rpcMoments) RepostAMoment(c context.Context, req *pbMoments.RepostAMomentRequest) (*pbMoments.CreateMomentResponse, error) {

	comResp, moment, errRes := checkMomentCreatorIsFrnInReqFrnLst(req.MomentID, req.CreatorID)
	if errRes != nil {
		return &pbMoments.CreateMomentResponse{ErrCode: comResp.ErrCode, ErrMsg: comResp.ErrMsg}, errRes
	}
	if moment != nil {

		moment.IsReposted = true
		moment.OrignalID = moment.MomentID
		moment.OrignalCreatorID = moment.CreatorID
		moment.OriginalCreatorName = moment.UserName
		moment.OriginalCreatorProfileImg = moment.UserProfileImg
		moment.CreatorID = req.CreatorID
		moment.MLikesCount = 0
		moment.MCommentsCount = 0

		user, err := imdb.GetUserByUserID(req.CreatorID)
		if err != nil {
			log.NewError("GetUserByUserID failed ", err.Error(), req.CreatorID)
		} else {
			moment.UserID = user.UserID
			moment.UserName = user.Nickname
			moment.UserProfileImg = user.FaceURL
		}
		var momentSQL db.MomentSQL
		err = utils.CopyStructFields(&momentSQL, moment)
		if err != nil {
			log.NewDebug(utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}

		mongoID := primitive.NewObjectIDFromTimestamp(time.Now())
		moment.MomentID = mongoID

		momentSQL.MomentID = mongoID.Hex()
		momentSQL.OrignalID = moment.OrignalID.Hex()
		momentSQL.OrignalCreatorID = moment.OrignalCreatorID
		moment.CommentCtl = 1
		//insert in Mongo DB
		err = db.DB.AddMoment(*moment)
		if err != nil {
			errMsg := "CleanUpUserMsgFromMongo failed " + err.Error()
			log.Error(errMsg)
			return &pbMoments.CreateMomentResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}
		//insert in MYSQL DB
		go func() {
			err = imdb.InsertMoment(momentSQL)
			if err != nil {
				log.NewError("CreateMoment", "insert in db failed", err.Error())
			}
			updateMomentCounts(req.MomentID, 0, 0, 1)
		}()
		momentObj := &pbMoments.Moment{}
		utils.CopyStructFields(momentObj, moment)

		momentObj.MomentID = moment.MomentID.Hex()
		momentObj.OrignalID = moment.OrignalID.Hex()
		momentObj.OrignalCreatorID = moment.OrignalCreatorID

		return &pbMoments.CreateMomentResponse{ErrCode: 0, ErrMsg: "Moment reposted", Moment: momentObj}, nil

	}
	return nil, errors.New("Moment Is not accessble")
}

// DeleteMoment for delete moment by owner ID
func (rpc *rpcMoments) DeleteMoment(_ context.Context, req *pbMoments.DeleteMomentRequest) (*pbMoments.CommonResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	comResp, moment, errRes := checkMomentCreatorIsFriend(req.MomentID, req.CreatorID)
	if errRes != nil {
		return comResp, errRes
	}
	if moment != nil && moment.CreatorID != req.CreatorID {
		return &pbMoments.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: "you are not allowed to perform this action"}, errors.New("you are not allowed to perform this action")
	}
	if moment != nil {
		err := db.DB.DeleteMoment(*moment)
		if err != nil {
			errMsg := "Delete Moment is failed " + err.Error()
			log.Error(errMsg)
			return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}
	}

	go func() {
		imdb.DeleteMoments([]string{req.MomentID}, "")
		momentMongo, err := db.DB.GetMoment(moment.MomentID)
		if err != nil {
			log.Debug("", "get moment from mongodb failed ", err.Error())
			return
		}
		if momentMongo.ArticleID != 0 {
			news.UpdateArticleCount(true, true, momentMongo.ArticleID, 0, 0, 0, -1, 0)
		}

		removeMomentsFiles([]db.Moment{*moment})
	}()

	comResp = &pbMoments.CommonResp{ErrCode: constant.NoError, ErrMsg: ""}
	return comResp, nil
}

func (rpc *rpcMoments) GetAnyUserMomentsByID(_ context.Context, req *pbMoments.GetAnyUserMomentsByIDRequest) (*pbMoments.GetAnyUserMomentsByIDResp, error) {
	log.Debug(req.OperationID, "get comments req: ", req.String())
	resp := &pbMoments.GetAnyUserMomentsByIDResp{CommonResp: &pbMoments.CommonResp{}}
	isFriend := checkUsersAreFriend(req.CurrentUserId, req.UserId)

	moments, err := db.DB.GetMomentsByUserId(req.UserId, isFriend, req.PageNumber, req.ShowNumber)
	if err != nil {
		errMsg := "Get Moment is failed " + err.Error()
		log.Error(errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	for _, detail := range moments {

		momentDetailRes := &pbMoments.Moment{}
		err = utils.CopyStructFields(momentDetailRes, detail)
		if err == nil {
			in := []byte(detail.MContentImagesArray)
			var MomentImageRequestObjects []*pbMoments.MomentImageRequestObject
			err := json.Unmarshal(in, &MomentImageRequestObjects)
			if err == nil {
				momentDetailRes.MContentImagesArrayV2 = append(momentDetailRes.MContentImagesArrayV2, MomentImageRequestObjects...)
			}

			in = []byte(detail.MContentVideosArray)
			var MomentVideoRequestObjects []*pbMoments.MomentVideoRequestObject
			err = json.Unmarshal(in, &MomentVideoRequestObjects)
			if err == nil {
				momentDetailRes.MContentVideosArrayV2 = append(momentDetailRes.MContentVideosArrayV2, MomentVideoRequestObjects...)
			}
			if detail.ArticleID != 0 {
				article, err := db.DB.GetArticlesByID(detail.ArticleID)
				if article != nil && err == nil {
					momentDetailRes.ArticleDetailsInMoment = &pbMoments.ArticleDetailsInMoment{Article: &pbNews.UserArticle{}, Official: &pbNews.UserFollow{}}
					_ = utils.CopyStructFields(momentDetailRes.ArticleDetailsInMoment.Article, article)
					momentDetailRes.ArticleDetailsInMoment.Official.OfficialID = article.OfficialID
					momentDetailRes.ArticleDetailsInMoment.Official.Nickname = article.OfficialName
					momentDetailRes.ArticleDetailsInMoment.Official.FaceURL = article.OfficialProfileImg
				}
			}

			//TODO Fetch Woom Details if Moment contain
			if detail.WoomFileID != "" {
				var rpcRequest pbShortVideo.GetShortVideoByFileIdRequest
				connection, err := short_video.CreateShortVideoGRPConnection("tempOperationID")
				if err == nil {
					rpcRequest.OperationID = "tempOperationID"
					rpcRequest.UserId = detail.UserID
					rpcRequest.FileId = detail.WoomFileID
					rpcResponse, err := connection.GetShortVideoByFieldId(context.Background(), &rpcRequest)
					if err == nil && rpcResponse != nil {
						momentDetailRes.WoomDetails = &pbShortVideo.ShortVideoInfo{}
						utils.CopyStructFields(momentDetailRes.WoomDetails, rpcResponse.ShortVideoInfo)
					}
				}
			}
			resp.Moments = append(resp.Moments, momentDetailRes)
		}
	}
	if err != nil {
		errMsg := "Copy failed " + err.Error()
		log.Error(errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	for index, moment := range moments {
		resp.Moments[index].MomentID = moment.MomentID.Hex()
		resp.Moments[index].OrignalID = moment.OrignalID.Hex()
	}

	resp.CommonResp.ErrCode = constant.NoError
	resp.CommonResp.ErrMsg = ""
	return resp, nil
}

func (rpc *rpcMoments) GetUserMomentCount(_ context.Context, req *pbMoments.GetUserMomentCountRequest) (*pbMoments.GetUserMomentCountResp, error) {

	log.Debug(req.OperationID, "get comments req: ", req.String())
	resp := &pbMoments.GetUserMomentCountResp{CommonResp: &pbMoments.CommonResp{}}
	isFriend := checkUsersAreFriend(req.UserId, req.CurrentUserId)

	posts, err := db.DB.GetUserMomentCount(req.UserId, isFriend)
	if err != nil {
		errMsg := "Get Moment is failed " + err.Error()
		log.Error(errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	resp.Posts = posts

	likeCounts, err := db.DB.GetAllMomentLikeCounts(req.UserId)
	if err != nil {
		errMsg := "GetAllMomentLikeCounts is failed " + err.Error()
		log.Error(errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = errMsg
		return resp, nil
	}
	resp.Likes = likeCounts

	resp.CommonResp.ErrCode = constant.NoError
	resp.CommonResp.ErrMsg = ""
	return resp, nil

}

func (rpc *rpcMoments) GetMoments(_ context.Context, req *pbMoments.GetMomentsReq) (*pbMoments.GetMomentsResp, error) {
	resp := &pbMoments.GetMomentsResp{Moments: []*pbMoments.GetMomentRes{}}
	where := map[string]string{}
	where["account"] = req.Account
	where["privacy"] = strconv.Itoa(int(req.Privacy))
	where["content_type"] = strconv.Itoa(int(req.ContentType))
	where["content"] = req.Content
	where["media_type"] = strconv.Itoa(int(req.MediaType))
	where["is_reposted"] = strconv.Itoa(int(req.IsReposted))
	where["original_user"] = req.OriginalUser
	where["is_blocked"] = strconv.Itoa(int(req.IsBlocked))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	moments, momentCounts, err := imdb.GetMomentsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetMoments failed", err.Error())
		return resp, err
	}
	_ = utils.CopyStructFields(&resp.Moments, &moments)
	for index, moment := range moments {
		_ = json.Unmarshal([]byte(moment.MContentImagesArray), &resp.Moments[index].MContentImagesArrayV2)
		_ = json.Unmarshal([]byte(moment.MContentVideosArray), &resp.Moments[index].MContentVideosArrayV2)
	}

	resp.MomentsNums = int32(momentCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcMoments) DeleteMoments(_ context.Context, req *pbMoments.DeleteMomentsReq) (*pbMoments.DeleteMomentsResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	resp := &pbMoments.DeleteMomentsResp{}
	if len(req.Moments) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	if len(req.Moments) != len(req.ArticleIDs) {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	if row := imdb.DeleteMoments(req.Moments, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	// delete moments in mongodb
	go func() {
		articleCountMap := make(map[int64]int64)
		// delete from mongodb
		var moments []db.Moment
		for index, momentIDStr := range req.Moments {

			// delete moment from mongodb
			momentIDPrimitive, _ := primitive.ObjectIDFromHex(momentIDStr)
			moment, _ := db.DB.GetMoment(momentIDPrimitive)
			moments = append(moments, *moment)
			_ = db.DB.DeleteMoment(db.Moment{MomentID: momentIDPrimitive})
			parentId := req.ArticleIDs[index]

			// accumulate the count for changing
			if parentId != 0 {
				//if req.MomentsType[index] == 0 {
				//	// moment
				//	if _, ok := momentCountMap[parentId]; !ok {
				//		momentCountMap[parentId] = 0
				//	}
				//	momentCountMap[parentId] += 1
				//} else if req.MomentsType[index] == 1 {
				//
				//}
				// article
				if _, ok := articleCountMap[parentId]; !ok {
					articleCountMap[parentId] = 0
				}
				articleCountMap[parentId] += 1
			}
		}
		for k, v := range articleCountMap {
			news.UpdateArticleCount(true, true, k, 0, 0, 0, -v, 0)
		}
		removeMomentsFiles(moments)
	}()

	return resp, nil
}

func (rpc *rpcMoments) AlterMoment(_ context.Context, req *pbMoments.AlterMomentReq) (*pbMoments.AlterMomentResp, error) {
	resp := &pbMoments.AlterMomentResp{}
	if req.MomentId == "" {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	moments := imdb.GetMoment([]string{req.MomentId})
	if len(moments) == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "the moment don't exist", req)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	moment := moments[0]
	momentMongo := db.Moment{}
	_ = utils.CopyStructFields(&momentMongo, moment)
	momentReq := db.MomentSQL{}
	_ = utils.CopyStructFields(&momentReq, moment)
	_ = utils.CopyStructFields(&momentReq, req)
	momentReq.MContentText = req.Content
	MContentImagesArrayV2Bytes, err := json.Marshal(req.MContentImagesArrayV2)
	if err == nil {
		momentReq.MContentImagesArray = string(MContentImagesArrayV2Bytes)
	}
	MContentVideosArrayV2Bytes, err := json.Marshal(req.MContentVideosArrayV2)
	if err == nil {
		momentReq.MContentVideosArray = string(MContentVideosArrayV2Bytes)
	}
	// get old moment url list
	oldMediasMap := getMomentMediaMap(momentMongo)
	var newMediaFiles []string
	var oldNeedDeleted []string
	for _, img := range req.MContentImagesArrayV2 {
		if _, ok := oldMediasMap[img.ImageUrl]; img.ImageUrl != "" && !ok {
			newMediaFiles = append(newMediaFiles, img.ImageUrl)
		}
		if _, ok := oldMediasMap[img.SnapShotUrl]; img.SnapShotUrl != "" && !ok {
			newMediaFiles = append(newMediaFiles, img.SnapShotUrl)
		}
	}
	for _, vid := range req.MContentVideosArrayV2 {
		if _, ok := oldMediasMap[vid.VideoUrl]; vid.VideoUrl != "" && !ok {
			newMediaFiles = append(newMediaFiles, vid.VideoUrl)
		}
		if _, ok := oldMediasMap[vid.SnapShotUrl]; vid.SnapShotUrl != "" && !ok {
			newMediaFiles = append(newMediaFiles, vid.SnapShotUrl)
		}
	}
	for url, _ := range oldMediasMap {
		// if the pic is in old moment, but isn't in alter moment.
		find := false
		for _, img := range req.MContentImagesArrayV2 {
			if img.ImageUrl == url || img.SnapShotUrl == url {
				find = true
				break
			}
		}
		if !find {
			for _, vid := range req.MContentVideosArrayV2 {
				if vid.VideoUrl == url || vid.SnapShotUrl == url {
					find = true
					break
				}
			}
		}
		if !find {
			mediaObjName := admin_cms.GetObjNameByURL(url, true)
			count, err := db.DB.GetFavoriteMediaCount(mediaObjName)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "get favorite media error, won't delete file:", url)
				continue
			}
			if count == 0 {
				oldNeedDeleted = append(oldNeedDeleted, url)
			}
		}
	}
	err = admin_cms.RemoveFiles(oldNeedDeleted)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete old file error, err：", err)
	}

	client, err := admin_cms.GetTencentCloudClient(true)
	failedList, err := admin_cms.RemoveDeleteTagForPersistent(client, newMediaFiles)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "upload file error")
		return resp, nil
	}
	if len(failedList) != 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "upload file error, failed list: ", failedList)
		return resp, nil
	}

	if row := imdb.AlterMoment(momentReq); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		momentIDPrimitive, _ := primitive.ObjectIDFromHex(moment.MomentID)
		originalIDPrimitive, _ := primitive.ObjectIDFromHex(moment.OrignalID)

		momentMongo.MomentID = momentIDPrimitive
		momentMongo.OrignalID = originalIDPrimitive
		momentMongo.MContentText = req.Content
		momentMongo.IsReposted = req.IsReposted
		momentMongo.Privacy = req.Privacy
		MContentImagesArrayV2Bytes, err := json.Marshal(req.MContentImagesArrayV2)
		if err == nil {
			momentMongo.MContentImagesArray = string(MContentImagesArrayV2Bytes)
		}
		MContentVideosArrayV2Bytes, err := json.Marshal(req.MContentVideosArrayV2)
		if err == nil {
			momentMongo.MContentVideosArray = string(MContentVideosArrayV2Bytes)
		}

		log.Debug("", "update moment.")
		err = db.DB.UpdateMoment(momentMongo)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", err.Error())
			return
		}

	}()

	return resp, nil
}

func (rpc *rpcMoments) ChangeMomentStatus(_ context.Context, req *pbMoments.ChangeMomentStatusReq) (*pbMoments.ChangeMomentStatusResp, error) {
	resp := &pbMoments.ChangeMomentStatusResp{}
	if len(req.MomentIds) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	if row := imdb.ChangeMomentStatus(req.MomentIds, int8(req.Status)); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	log.Debug("", "change moment status ok")

	go func() {
		m := db.Moment{}
		for _, moment := range req.MomentIds {
			m.MomentID, _ = primitive.ObjectIDFromHex(moment)
			m.Status = int8(req.Status)
			_ = db.DB.UpdateMomentFields(m)
		}
	}()

	return resp, nil
}

func (rpc *rpcMoments) ModifyVisibility(_ context.Context, req *pbMoments.ModifyVisibilityReq) (*pbMoments.ModifyVisibilityResp, error) {
	log.Debug(req.OperationID, "ModifyVisibility req: ", req.String())
	resp := &pbMoments.ModifyVisibilityResp{}
	if len(req.MomentIds) == 0 {
		return resp, nil
	}

	if row := imdb.ModifyVisibility(req.MomentIds, req.Privacy); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		m := db.Moment{}
		for _, moment := range req.MomentIds {
			m.MomentID, _ = primitive.ObjectIDFromHex(moment)
			m.Privacy = req.Privacy
			_ = db.DB.UpdateMomentFields(m)
		}
	}()

	return resp, nil
}

func (rpc *rpcMoments) GetMomentDetails(_ context.Context, req *pbMoments.GetMomentDetailsReq) (*pbMoments.GetMomentDetailsResp, error) {
	resp := &pbMoments.GetMomentDetailsResp{}
	where := map[string]string{}
	where["account"] = req.Account
	where["content_type"] = strconv.Itoa(int(req.ContentType))
	where["content"] = req.Content
	where["privacy"] = strconv.Itoa(int(req.Privacy))
	where["media_type"] = strconv.Itoa(int(req.MediaType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["moment_id"] = req.MomentID
	where["original_id"] = req.OriginalID

	details, momentCounts, err := imdb.GetMomentDetailsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetMoments failed", err.Error())
		return resp, err
	}
	log.Debug("", "details len: ", len(details))
	for _, detail := range details {

		momentDetailRes := &pbMoments.MomentDetailRes{}
		err = utils.CopyStructFields(momentDetailRes, detail)
		if err == nil {
			in := []byte(detail.MContentImagesArray)
			var MomentImageRequestObjects []*pbMoments.MomentImageRequestObject
			err = json.Unmarshal(in, &MomentImageRequestObjects)
			if err == nil {
				momentDetailRes.MContentImagesArrayV2 = append(momentDetailRes.MContentImagesArrayV2, MomentImageRequestObjects...)
			}

			in = []byte(detail.MContentVideosArray)
			var MomentVideoRequestObjects []*pbMoments.MomentVideoRequestObject
			err = json.Unmarshal(in, &MomentVideoRequestObjects)
			if err == nil {
				momentDetailRes.MContentVideosArrayV2 = append(momentDetailRes.MContentVideosArrayV2, MomentVideoRequestObjects...)
			}
			resp.MomentDetails = append(resp.MomentDetails, momentDetailRes)
		}
	}

	resp.MomentsNums = int32(momentCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcMoments) CtlMomentComment(_ context.Context, req *pbMoments.CtlMomentCommentReq) (*pbMoments.CtlMomentCommentResp, error) {
	log.Debug(req.OperationID, "CtlMomentComment req: ", req.String())
	resp := &pbMoments.CtlMomentCommentResp{}
	if row := imdb.CtlMomentComment(req.MomentId, req.CommentCtl); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "ctl failed", "ctl rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		m := db.Moment{}
		m.MomentID, _ = primitive.ObjectIDFromHex(req.MomentId)
		m.CommentCtl = req.CommentCtl
		_ = db.DB.UpdateMomentFields(m)

	}()

	return resp, nil
}

func (rpc *rpcMoments) GetComments(_ context.Context, req *pbMoments.GetCommentsReq) (*pbMoments.GetCommentsResp, error) {
	resp := &pbMoments.GetCommentsResp{Comments: []*pbMoments.GetMomentComment{}}
	where := map[string]string{}
	where["publish_user"] = req.PublishUser
	where["content_type"] = strconv.Itoa(int(req.ContentType))
	where["m_content_text"] = req.MContentText
	where["comment_user"] = req.CommentUser
	where["comment_content"] = req.CommentContent
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["media_type"] = strconv.Itoa(int(req.MediaType))
	where["privacy"] = strconv.Itoa(int(req.Privacy))

	where["comment_type"] = req.CommentType
	where["commented_user"] = req.CommentedUser
	where["moment_id"] = req.MomentId
	where["reply_comment_id"] = req.ReplyCommentId

	comments, commentCounts, err := imdb.GetMomentCommentsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetMoments failed", err.Error())
		return resp, err
	}

	log.Debug("", "comments len: ", len(comments))
	_ = utils.CopyStructFields(&resp.Comments, &comments)
	for index, comment := range comments {
		_ = json.Unmarshal([]byte(comment.MContentVideosArray), &resp.Comments[index].MContentVideosArrayV2)
		_ = json.Unmarshal([]byte(comment.MContentImagesArray), &resp.Comments[index].MContentImagesArrayV2)
	}

	resp.CommentsNums = int32(commentCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcMoments) RemoveComments(_ context.Context, req *pbMoments.RemoveCommentsReq) (*pbMoments.RemoveCommentsResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	resp := &pbMoments.RemoveCommentsResp{}
	if len(req.CommentIds) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	if len(req.CommentIds) != len(req.MomentIds) {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "len is different")
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	if len(req.CommentIds) != len(req.ParentIds) {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "parent len is different")
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	if len(req.CommentIds) != len(req.ReplyIds) {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "reply len is different")
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	if row := imdb.DeleteMomentComments(req.CommentIds, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		momentCountMap := make(map[string]int32)
		commentCountMap := make(map[string]int64)

		// update mongo db.
		momentComment := db.MomentComment{}
		for index, comment := range req.CommentIds {
			momentComment.CommentID, _ = primitive.ObjectIDFromHex(comment)
			momentComment.MomentID, _ = primitive.ObjectIDFromHex(req.MomentIds[index])
			// comment reply count -1
			removeComment, _ := db.DB.GetMomentComment(momentComment.CommentID)
			if removeComment == nil {
				log.Debug("", "the moment has deleted")
				continue
			}
			// remove comment
			err := db.DB.DeleteMomentComment(momentComment, req.OpUserId)
			if err != nil {
				log.Debug("", "delete moment comment failed", momentComment.CommentID.Hex())
			}
			go msg.SendMomentInteractionNotification("MomentNotification_Action_DeleteComment"+removeComment.MomentID.Hex(), removeComment.MomentID.Hex(), removeComment.CreateBy, constant.MomentNotification_Action_DeleteComment, *removeComment)

			momentId := req.MomentIds[index]
			// moment's comment
			if req.ReplyIds[index] != "" {
				// count -1
				if _, ok := commentCountMap[req.ReplyIds[index]]; !ok {
					commentCountMap[req.ReplyIds[index]] = 1
				} else {
					commentCountMap[req.ReplyIds[index]] += 1
				}
			} else {
				// set moment's comment -1
				if _, ok := momentCountMap[momentId]; !ok {
					momentCountMap[momentId] = 1
				} else {
					momentCountMap[momentId] += 1
				}
			}
		}

		log.Debug("", "update comment count, momentCountMap: ", momentCountMap)
		for comment, val := range commentCountMap {
			log.Debug("", "comment minus ", val, ", comment: ", comment)
			updateMomentCommentCounts(comment, 0, -val)
		}
		// update mysql
		for moment, val := range momentCountMap {
			log.Debug("", "moment minus", val, ", moment: ", moment)
			updateMomentCounts(moment, 0, int64(-val), 0)
		}

	}()

	return resp, nil
}

func (rpc *rpcMoments) AlterComment(_ context.Context, req *pbMoments.AlterCommentReq) (*pbMoments.AlterCommentResp, error) {
	log.Debug(req.OperationID, "alter comments req: ", req.String())
	resp := &pbMoments.AlterCommentResp{}
	if row := imdb.AlterMomentComment(req); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		comments := imdb.GetMomentComment([]string{req.CommentId})
		momentComment := db.MomentComment{}
		err := utils.CopyStructFields(&momentComment, comments[0])
		if err != nil {
			return
		}
		momentComment.CommentID, _ = primitive.ObjectIDFromHex(comments[0].CommentID)

		_ = db.DB.UpdateMomentComment(momentComment)
	}()

	return resp, nil
}

func (rpc *rpcMoments) SwitchCommentHideState(_ context.Context, req *pbMoments.SwitchCommentHideStateReq) (*pbMoments.SwitchCommentHideStateResp, error) {
	log.Debug(req.OperationID, "SwitchCommentHideState req: ", req.String())
	resp := &pbMoments.SwitchCommentHideStateResp{}
	if row := imdb.SwitchMomentCommentHideState(req); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		momentComment := db.MomentComment{}
		momentComment.CommentID, _ = primitive.ObjectIDFromHex(req.CommentId)
		momentComment.Status = int8(req.Status)
		momentComment.UpdateBy = req.OpUserId

		_ = db.DB.UpdateMomentCommentStatus(momentComment)
	}()

	return resp, nil

}

func (rpc *rpcMoments) GetLikes(_ context.Context, req *pbMoments.GetLikesReq) (*pbMoments.GetLikesResp, error) {
	resp := &pbMoments.GetLikesResp{Likes: []*pbMoments.GetMomentLike{}}
	where := map[string]string{}
	where["publish_user"] = req.PublishUser
	where["content_type"] = strconv.Itoa(int(req.ContentType))
	where["m_content_text"] = req.MContentText
	where["like_user"] = req.LikeUser
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["moment_id"] = req.MomentId
	where["media_type"] = strconv.Itoa(int(req.MediaType))
	where["privacy"] = strconv.Itoa(int(req.Privacy))

	likes, likeCounts, err := imdb.GetMomentLikesByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetLikes failed", err.Error())
		return resp, err
	}

	log.Debug("", "likes len: ", len(likes))
	_ = utils.CopyStructFields(&resp.Likes, &likes)

	for index, like := range likes {
		_ = json.Unmarshal([]byte(like.MContentVideosArray), &resp.Likes[index].MContentVideosArrayV2)
		_ = json.Unmarshal([]byte(like.MContentImagesArray), &resp.Likes[index].MContentImagesArrayV2)
	}

	resp.LikeNums = int32(likeCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcMoments) RemoveLikes(_ context.Context, req *pbMoments.RemoveLikesReq) (*pbMoments.RemoveLikesResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	resp := &pbMoments.RemoveLikesResp{}
	if len(req.MomentsId) != len(req.UsersId) {
		log.Error("", "moment id and user id aren't equals")
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	if row := imdb.RemoveMomentsLikes(req.MomentsId, req.UsersId, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {

		likeCountMap := make(map[string]int32)

		// update mongo db.
		momentLike := db.MomentLike{}
		for index, moment := range req.MomentsId {
			momentLike.MomentID, _ = primitive.ObjectIDFromHex(moment)
			momentLike.CreateBy = req.UsersId[index]
			momentCancelLike, err := db.DB.GetMomentLikeByID(momentLike.MomentID, momentLike.CreateBy)
			if err != nil {
				continue
			}

			err = db.DB.CancelMomentLike(momentLike)
			if err != nil {
				log.Error("", "delete moment like: ", momentLike.MomentID, "  user id: ", momentLike.UserID, " error ", err.Error())
				continue
			}
			msg.SendMomentInteractionNotification("MomentNotification_Action_CancelLike"+momentLike.MomentID.Hex(), momentLike.MomentID.Hex(), momentLike.UserID, constant.MomentNotification_Action_CancelLike, *momentCancelLike)

			//filter
			if _, ok := likeCountMap[moment]; !ok {
				likeCountMap[moment] = 1
			} else {
				likeCountMap[moment] += 1
			}
		}

		log.Debug("", "update comment count, momentCountMap: ", likeCountMap)

		log.Debug("", "likeCountMap: ", likeCountMap)
		// update mysql
		for moment, val := range likeCountMap {
			log.Debug("", "moment: ", moment)
			updateMomentCounts(moment, int64(-val), 0, 0)
		}
	}()

	return resp, nil
}

func (rpc *rpcMoments) SwitchLikeHideState(_ context.Context, req *pbMoments.SwitchLikeHideStateReq) (*pbMoments.SwitchLikeHideStateResp, error) {
	resp := &pbMoments.SwitchLikeHideStateResp{}

	if row := imdb.SwitchMomentLikeHideState(req); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "switch like failed", "switch like rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		momentLike := db.MomentLike{}
		momentLike.MomentID, _ = primitive.ObjectIDFromHex(req.MomentId)
		momentLike.UserID = req.UserId
		momentLike.UpdateBy = req.OpUserId
		momentLike.Status = int8(req.Status)

		_ = db.DB.UpdateLikeStatus(momentLike)
	}()

	return resp, nil
}

// DeleteMomentComment for delete moment comment by comment ID
func (rpc *rpcMoments) DeleteMomentComment(_ context.Context, req *pbMoments.DeleteMomentCommentRequest) (*pbMoments.CommonResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	//Check Comment Owner ID
	commentIDPrimitive, _ := primitive.ObjectIDFromHex(req.CommentID)
	comment, errM := db.DB.GetMomentComment(commentIDPrimitive)
	if errM != nil || comment == nil {
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Comment Not found in DB"}, errors.New("comment not found in DB")
	}
	//Check Moment Owner ID
	moment, errM := db.DB.GetMoment(comment.MomentID)
	if errM != nil || moment == nil {
		return &pbMoments.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Moment Not found in DB"}, errors.New("comment parent moment not found in DB")
	}
	if comment.UserID == req.CreatorID || moment.UserID == req.CreatorID {
		_ = db.DB.DeleteMomentComment(*comment, req.CreatorID)
		go func() {
			imdb.DeleteMomentCommentByID(req.CommentID, req.CreatorID)
			if comment.ReplyCommentID.IsZero() {
				updateMomentCounts(comment.MomentID.Hex(), 0, -1, 0)
			} else {
				updateMomentCommentCounts(comment.ReplyCommentID.Hex(), 0, -1)
			}
		}()
		go msg.SendMomentInteractionNotification("MomentNotification_Action_DeleteComment"+comment.MomentID.Hex(), comment.MomentID.Hex(), req.CreatorID, constant.MomentNotification_Action_DeleteComment, *comment)

		comResp := &pbMoments.CommonResp{ErrCode: constant.NoError, ErrMsg: ""}
		return comResp, nil
	}

	comResp := &pbMoments.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}
	return comResp, nil
}

// GlobalSearchInMoments RPC method for search moments by keywords by checking these are by users friends
func (rpc *rpcMoments) GlobalSearchInMoments(_ context.Context, req *pbMoments.GlobalSearchInMomentsRequest) (*pbMoments.ListHomeTimeLineOfMoments, error) {

	var listHomeTimeLineOfMoments = pbMoments.ListHomeTimeLineOfMoments{}
	myFriendIDList, _ := db.DB.GetFriendIDListFromCache(req.CreatorID)
	myFriendIDList = append(myFriendIDList, req.CreatorID)
	log.Error(utils.GetSelfFuncName(), "Friend List for Moments fetch ", myFriendIDList)
	moments, err := db.DB.GetMomentsByFriendListAndSearchKeyword(myFriendIDList, req.SearchKeyWord, req.PageNumber, req.MomentLimit)
	log.Error(utils.GetSelfFuncName(), "Moments Fetched ", moments)
	if err == nil {
		// listHomeTimeLineOfMoments.HomeTimeLineOfMoments
		for _, moment := range moments {
			var timelineMoment = pbMoments.HomeTimeLineOfMoments{}
			timelineMoment.Moment = &pbMoments.Moment{}
			timelineMoment.Moment.CreatorID = moment.CreatorID
			timelineMoment.Moment.MomentID = moment.MomentID.Hex()
			timelineMoment.Moment.MContentText = moment.MContentText
			timelineMoment.Moment.MContentImagesArray = moment.MContentImagesArray
			timelineMoment.Moment.MContentVideosArray = moment.MContentVideosArray
			timelineMoment.Moment.MContentThumbnilArray = moment.MContentThumbnilArray
			_ = json.Unmarshal([]byte(moment.MContentImagesArray), &timelineMoment.Moment.MContentImagesArrayV2)
			_ = json.Unmarshal([]byte(moment.MContentVideosArray), &timelineMoment.Moment.MContentVideosArrayV2)

			timelineMoment.Moment.MLikesCount = moment.MLikesCount
			timelineMoment.Moment.MCommentsCount = moment.MCommentsCount
			timelineMoment.Moment.MRepostCount = moment.MRepostCount
			timelineMoment.Moment.MCreateTime = moment.MCreateTime
			timelineMoment.Moment.MUpdateTime = moment.MUpdateTime
			timelineMoment.Moment.OrignalCreatorID = moment.OrignalCreatorID
			timelineMoment.Moment.OriginalCreatorName = moment.OriginalCreatorName
			timelineMoment.Moment.OriginalCreatorProfileImg = moment.OriginalCreatorProfileImg
			timelineMoment.Moment.IsReposted = moment.IsReposted
			timelineMoment.Moment.Status = int32(moment.Status)
			timelineMoment.Moment.Privacy = moment.Privacy
			timelineMoment.Moment.UserID = moment.UserID
			timelineMoment.Moment.UserName = moment.UserName
			timelineMoment.Moment.UserProfileImg = moment.UserProfileImg

			//TODO Fetch Comments of this moments
			momentsComments, err := db.DB.GetMomentCommentsByMomentAndFriendIds(moment.MomentID, myFriendIDList, 0, req.CommentsLimit)
			if err == nil {
				for _, comment := range momentsComments {
					var MomentCommentResp = pbMoments.MomentCommentResp{}
					MomentCommentResp.MomentID = comment.MomentID.Hex()
					MomentCommentResp.CommentID = comment.CommentID.Hex()
					MomentCommentResp.UserID = comment.UserID
					MomentCommentResp.UserName = comment.UserName
					MomentCommentResp.UserProfileImg = comment.UserProfileImg
					MomentCommentResp.CommentContent = comment.CommentContent
					MomentCommentResp.CommentParentID = comment.CommentParentID.Hex()
					MomentCommentResp.CreateBy = comment.CreateBy
					MomentCommentResp.CreateTime = comment.CreateTime
					MomentCommentResp.UpdateBy = comment.UpdateBy
					MomentCommentResp.UpdatedTime = comment.UpdatedTime
					MomentCommentResp.Status = int32(comment.Status)
					MomentCommentResp.AccountStatus = int32(comment.AccountStatus)

					timelineMoment.MomentComments = append(timelineMoment.MomentComments, &MomentCommentResp)
				}
			}

			//TODO Fetch Likes of this Moment
			momentsLikes, err := db.DB.GetMomentLikesByMomentAndFriendIds(moment.MomentID, myFriendIDList)
			if err == nil {
				for _, momentsLike := range momentsLikes {
					var MomentLikeResp = pbMoments.MomentLikeResponse{}
					MomentLikeResp.MomentID = momentsLike.MomentID.Hex()
					MomentLikeResp.UserID = momentsLike.UserID
					MomentLikeResp.UserName = momentsLike.UserName
					MomentLikeResp.UserProfileImg = momentsLike.UserProfileImg
					MomentLikeResp.CreateBy = momentsLike.CreateBy
					MomentLikeResp.CreateTime = momentsLike.CreateTime
					MomentLikeResp.UpdateBy = momentsLike.UpdateBy
					MomentLikeResp.UpdatedTime = momentsLike.UpdatedTime
					MomentLikeResp.Status = int32(momentsLike.Status)
					timelineMoment.MomentLikes = append(timelineMoment.MomentLikes, &MomentLikeResp)
				}
			}

			listHomeTimeLineOfMoments.HomeTimeLineOfMoments = append(listHomeTimeLineOfMoments.HomeTimeLineOfMoments, &timelineMoment)
		}
	} else {
		log.Error(utils.GetSelfFuncName(), "Moments Not Fetched ", err.Error())
	}

	return &listHomeTimeLineOfMoments, nil
}

func removeMomentsFiles(moments []db.Moment) {
	var obs []cos.Object
	for _, moment := range moments {
		var medias = getMomentMedias(moment)
		if len(medias) > 0 {
			for _, media := range medias {
				mediaObjName := admin_cms.GetObjNameByURL(media, true)
				count, err := db.DB.GetFavoriteMediaCount(mediaObjName)
				if err != nil {
					log.Error("", "Favorite get failed")
					return
				}
				if count == 0 {
					obs = append(obs, cos.Object{Key: mediaObjName})
				}
			}
		}
		msg.SendMomentInteractionNotification("MomentNotification_Action_DeleteMoment"+moment.MomentID.Hex(), moment.MomentID.Hex(), moment.CreatorID, constant.MomentNotification_Action_DeleteMoment, moments)
	}
	if len(obs) == 0 {
		return
	}
	deleteOpt := &cos.ObjectDeleteMultiOptions{Objects: obs}
	client, err := admin_cms.GetTencentCloudClient(true)
	if err != nil {
		log.Error("", "Tencent cloud get failed")
		return
	}

	_, _, err = client.Object.DeleteMulti(context.Background(), deleteOpt)
	if err != nil {

		client, err = admin_cms.GetTencentCloudClient(false)
		if err != nil {
			log.Error("", "Tencent cloud get failed")
			return
		}
		_, _, err = client.Object.DeleteMulti(context.Background(), deleteOpt)
		if err != nil {
			log.Error("", "Delete file from tencent cloud failed.")
		}
		return
	}
	return
}

func getMomentMedias(moment db.Moment) []string {
	var medias []string
	in := []byte(moment.MContentImagesArray)
	var MomentImageRequestObjects []*pbMoments.MomentImageRequestObject
	err := json.Unmarshal(in, &MomentImageRequestObjects)
	if err == nil {
		for _, object := range MomentImageRequestObjects {
			medias = append(medias, object.ImageUrl)
		}
	}

	in = []byte(moment.MContentVideosArray)
	var MomentVideoRequestObjects []*pbMoments.MomentVideoRequestObject
	err = json.Unmarshal(in, &MomentVideoRequestObjects)
	if err == nil {
		for _, object := range MomentVideoRequestObjects {
			medias = append(medias, object.VideoUrl)
			medias = append(medias, object.SnapShotUrl)
		}
	}

	return medias
}

func getMomentMediaMap(moment db.Moment) map[string]struct{} {
	var medias = make(map[string]struct{}, 0)
	in := []byte(moment.MContentImagesArray)
	var MomentImageRequestObjects []*pbMoments.MomentImageRequestObject
	err := json.Unmarshal(in, &MomentImageRequestObjects)
	if err == nil {
		for _, object := range MomentImageRequestObjects {
			medias[object.ImageUrl] = struct{}{}
		}
	}

	in = []byte(moment.MContentVideosArray)
	var MomentVideoRequestObjects []*pbMoments.MomentVideoRequestObject
	err = json.Unmarshal(in, &MomentVideoRequestObjects)
	if err == nil {
		for _, object := range MomentVideoRequestObjects {
			medias[object.VideoUrl] = struct{}{}
			medias[object.SnapShotUrl] = struct{}{}
		}
	}

	return medias
}

func updateMomentCounts(momentID string, likeCount, commentCount, repostCount int64) {
	momentIDPrimitive, _ := primitive.ObjectIDFromHex(momentID)
	moment := db.Moment{MomentID: momentIDPrimitive, MLikesCount: int32(likeCount), MCommentsCount: int32(commentCount), MRepostCount: repostCount}
	momentSQL := db.MomentSQL{MomentID: momentID, MLikesCount: int32(likeCount), MCommentsCount: int32(commentCount), MRepostCount: repostCount}

	// update moment like counts

	if _, err := imdb.UpdateMomentV2(&momentSQL); err != nil {
		log.NewError("UpdateMomentV2", "Update in db failed", err.Error())
	}

	if err := db.DB.UpdateMomentV2(moment); err != nil {
		log.NewError("UpdateMomentV2 Mongodb", "Update in mongodb failed", err.Error())
	}
}

func updateMomentCommentCounts(commentID string, likeCounts int64, commentReplies int64) {
	commentIDPrimitive, _ := primitive.ObjectIDFromHex(commentID)
	comment := db.MomentComment{CommentID: commentIDPrimitive, LikeCounts: likeCounts, CommentReplies: commentReplies}
	commentSQL := db.MomentCommentSQL{CommentID: commentID, LikeCounts: likeCounts, CommentReplies: commentReplies}

	// update moment like counts

	if err := imdb.UpdateMomentCommentV2(&commentSQL); err != nil {
		log.NewError("UpdateMomentCommentV2", "Update in db failed", err.Error())
	}

	if err := db.DB.UpdateMomentCommentV2(comment); err != nil {
		log.NewError("UpdateMomentCommentV2 Mongodb", "Update in mongodb failed", err.Error())
	}
}

func (rpc *rpcMoments) GetMomentAnyUserMediaByID(_ context.Context, req *pbMoments.GetMomentAnyUserMediaByIDRequest) (*pbMoments.GetMomentAnyUserMediaByIDResp, error) {
	resp := &pbMoments.GetMomentAnyUserMediaByIDResp{}

	moments, momentCount, err := db.DB.GetMomentsMediaByID(req.UserID, int64(req.LastCount))
	if err != nil {
		log.NewError("", "get moments error", err.Error())
		return resp, err
	}

	videos := []pbMoments.MomentVideoRequestObject{}
	images := []pbMoments.MomentImageRequestObject{}
	for _, moment := range moments {
		err = json.Unmarshal([]byte(moment.MContentVideosArray), &videos)
		if err != nil || len(videos) == 0 {
			err = json.Unmarshal([]byte(moment.MContentImagesArray), &images)
			if err != nil || len(images) == 0 {
				log.NewError("", "parse moment video error")
				continue
			} else {
				if len(images) > 0 {
					resp.Pics = append(resp.Pics, &pbMoments.Pic{
						URL:  images[0].ImageUrl,
						Type: 1,
					})
				}
			}
		} else {
			if len(videos) > 0 {
				resp.Pics = append(resp.Pics, &pbMoments.Pic{
					URL:  videos[0].SnapShotUrl,
					Type: 2,
				})
			}
		}
	}

	resp.AllMediaMomentCount = momentCount

	return resp, nil
}
