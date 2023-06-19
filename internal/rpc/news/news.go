package news

import (
	"Open_IM/internal/rpc/admin_cms"
	chat "Open_IM/internal/rpc/msg"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdminCMS "Open_IM/pkg/proto/admin_cms"
	pbNews "Open_IM/pkg/proto/news"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gorm.io/gorm"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"

	"Open_IM/pkg/common/config"

	"google.golang.org/grpc"
)

const (
	nicknameUpdateFirstWait  = 604800000  // 1 week in milliseconds
	nicknameUpdateSecondWait = 2629800000 // 1 month in milliseconds
)

type rpcNews struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func NewRpcNewsServer(port int) *rpcNews {
	return &rpcNews{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImNewsName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (rpc *rpcNews) Run() {
	log.NewPrivateLog(constant.OpenImNewsLog)
	log.NewInfo("0", "rpc news start...")

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
	pbNews.RegisterNewsServer(srv, rpc)
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
	log.NewInfo("0", "RegisterNewsServer ok ", rpc.etcdSchema, strings.Join(rpc.etcdAddr, ","), rpcRegisterIP, rpc.rpcPort, rpc.rpcRegisterName)

	UpdateRedisWithLatestArticlesByOfficialAccount()
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error())
		return
	}
	log.NewInfo("0", "rpc News ok")
}

func UpdateRedisWithLatestArticlesByOfficialAccount() {
	log.NewInfo("UpdateRedisWithLatestArticlesByOfficialAccount", "called")
	listOfArticles, err := imdb.GetLatestArticlesByOfficialID()
	if err != nil {
		log.NewError("UpdateRedisWithLatestArticlesByOfficialAccount", err.Error())
	}
	if listOfArticles != nil {
		log.NewInfo("UpdateRedisWithLatestArticlesByOfficialAccount", len(listOfArticles))
		for _, article := range listOfArticles {
			article.Content = ""
			//if len(article.TextContent) > 100 {
			//	article.TextContent = article.TextContent[0:98]
			//}
			err := db.DB.SetLatestArticlesByOfficialAccount(article)
			if err != nil {
				log.NewError("UpdateRedisWithLatestArticlesByOfficialAccount", err.Error())
			}

		}
	}

}

// RegisterOfficial RPC method for registering new official accounts
func (rpc *rpcNews) RegisterOfficial(_ context.Context, req *pbNews.RegisterOfficialRequest) (*pbNews.CommonResponse, error) {
	// get user
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	// abort if user already have an official account pending/verified
	if user.OfficialID != 0 {
		// get official account
		official, err := imdb.GetOfficialByOfficialID(user.OfficialID)
		if err == gorm.ErrRecordNotFound {
			// ignore and assume user have not official account registered
		} else if err != nil {
			errMsg := req.OperationID + " imdb.GetOfficialByOfficialID failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		} else {
			// official account is pending or verified
			if official.ProcessStatus == 0 || official.ProcessStatus == 1 {
				return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
			}
		}
	}

	officialFind, err := imdb.GetOfficialByName(req.Type, req.Nickname)
	if err != nil {
		log.NewError("", "query official error")
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialNameExist.ErrCode, ErrMsg: constant.ErrOfficialNameExist.ErrMsg}, nil
	}
	if officialFind.Id != 0 {
		log.NewError("", "official is already existed")
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialNameExist.ErrCode, ErrMsg: constant.ErrOfficialNameExist.ErrMsg}, nil
	}

	var official imdb.InsertOfficialParams
	if err = utils.CopyStructFields(&official, req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}
	official.UserGender = user.Gender

	IDNumberExists := imdb.CheckOfficialIDNumberAvailable(official.IdNumber, official.IdType, 0)
	if IDNumberExists {
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "id_number already used"}, err
	}

	var officialId int64 = 0
	if officialId, err = imdb.InsertOfficial(official); err != nil {
		errMsg := req.OperationID + " imdb.InsertOfficial failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err = db.DB.AddOfficialFollow(officialId, user.UserID); err != nil {
		errMsg := req.OperationID + " AddOfficialFollow failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

// GetSelfOfficialInfo RPC method for returning user official info
func (rpc *rpcNews) GetSelfOfficialInfo(_ context.Context, req *pbNews.GetSelfOfficialInfoRequest) (*pbNews.GetSelfOfficialInfoResponse, error) {
	var (
		responseData pbNews.GetSelfOfficialInfoResponse_Data
	)

	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetSelfOfficialInfoResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	responseData.UserInfo = &pbNews.SelfUserInfo{
		UserID:     user.UserID,
		OfficialID: user.OfficialID,
		Nickname:   user.Nickname,
		FaceURL:    user.FaceURL,
	}

	if user.OfficialID == 0 {
		return &pbNews.GetSelfOfficialInfoResponse{
			Data: &responseData,
		}, nil
	}

	official, err := imdb.GetOfficialByOfficialID(user.OfficialID)
	if err == gorm.ErrRecordNotFound {
		return &pbNews.GetSelfOfficialInfoResponse{
			Data: &responseData,
		}, nil
	} else if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialByOfficialID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetSelfOfficialInfoResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	responseData.OfficialInfo = &pbNews.SelfOfficialInfo{
		Nickname:            official.Nickname,
		NicknameUpdateTime:  official.NicknameUpdateTime,
		NicknameUpdateCount: int32(official.NicknameUpdateCount),
		Bio:                 official.Bio,
		FaceURL:             official.FaceURL,
		ProcessStatus:       int32(official.ProcessStatus),
		ProcessFeedback:     official.ProcessFeedback,
		PostCounts:          official.PostCounts,
		FollowCounts:        official.FollowCounts,
		LikeCounts:          official.LikeCounts,
	}

	officialInterests, err := imdb.GetOfficialInterestsByOfficialID(user.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialInterestsByOfficialID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetSelfOfficialInfoResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	responseData.OfficialInfo.Interests = officialInterests

	return &pbNews.GetSelfOfficialInfoResponse{
		Data: &responseData,
	}, nil
}

func (rpc *rpcNews) SetSelfOfficialInfo(_ context.Context, req *pbNews.SetSelfOfficialInfoRequest) (*pbNews.CommonResponse, error) {
	// get user
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	// user have no official account
	if user.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	// get official
	official, err := imdb.GetOfficialByOfficialID(user.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialByOfficialID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	// official account not verified
	if official.ProcessStatus != 1 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	// official account to be updated
	now := time.Now().Unix()
	updatedOfficial := imdb.UpdateSelfOfficialParams{
		OfficialID:          official.Id,
		FaceURL:             req.FaceURL,
		Nickname:            req.Nickname,
		Bio:                 req.Bio,
		NicknameUpdateTime:  official.NicknameUpdateTime,
		NicknameUpdateCount: official.NicknameUpdateCount,
		Interests:           req.Interests,
	}

	client, err := admin_cms.GetTencentCloudClient(true)
	if err != nil {
		errMsg := req.OperationID + " get cloud link failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if req.FaceURL != official.FaceURL {
		err = admin_cms.DeleteFileForPersistent(client, official.FaceURL)
		if err != nil {
			log.NewError("", "delete face url failed, url: ", official.FaceURL)
		}
		failedList, err := admin_cms.RemoveDeleteTagForPersistent(client, []string{req.FaceURL})
		if err != nil {
			errMsg := req.OperationID + " Upload face url failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}
		if len(failedList) > 0 {
			errMsg := req.OperationID + " Upload face url failed "
			log.NewError(req.OperationID, errMsg, failedList)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}

	}

	// check for nickname update limits
	if official.Nickname != updatedOfficial.Nickname {
		if (official.NicknameUpdateCount == 1 && official.NicknameUpdateTime+nicknameUpdateFirstWait > now) ||
			(official.NicknameUpdateCount == 2 && official.NicknameUpdateTime+nicknameUpdateSecondWait > now) ||
			official.NicknameUpdateCount > 2 {
			return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialNicknameUpdateLocked.ErrCode, ErrMsg: constant.ErrOfficialNicknameUpdateLocked.ErrMsg}, nil
		}
		updatedOfficial.NicknameUpdateCount++
		updatedOfficial.NicknameUpdateTime = now
	}

	if err = imdb.UpdateSelfOfficial(updatedOfficial); err != nil {
		errMsg := req.OperationID + " imdb.UpdateSelfOfficial failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) GetOfficialAccounts(_ context.Context, req *pbNews.GetOfficialAccountsReq) (*pbNews.GetOfficialAccountsResp, error) {
	resp := &pbNews.GetOfficialAccountsResp{OfficialAccount: []*pbNews.OfficialAccountResponse{}}
	where := map[string]string{}
	where["official_account"] = req.OfficialAccount
	where["account_type"] = strconv.Itoa(int(req.AccountType))
	where["id_type"] = strconv.Itoa(int(req.IdType))
	where["id_number"] = req.IdNumber
	where["process_status"] = strconv.Itoa(int(req.ProcessStatus))
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["is_system"] = strconv.Itoa(int(req.IsSystem))
	where["bio"] = req.Bio

	log.Debug("", "req.TagsId: [", req.TagsId, "]")
	var tagsIdList []int64
	if req.TagsId != "" {
		tagsIdListStr := strings.Split(req.TagsId, ",")

		for _, tagId := range tagsIdListStr {
			id, err := strconv.ParseInt(tagId, 10, 64)
			if err != nil {
				continue
			}
			tagsIdList = append(tagsIdList, id)
		}
	}

	officialAccountRes, officialCounts, err := imdb.GetOfficialAccountsByWhere(where, tagsIdList, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}
	for _, officialAccount := range officialAccountRes {

		log.Debug("", "imdb result: ", officialAccount.Interests)

		officialAccountResp := pbNews.OfficialAccountResponse{}
		official := pbNews.Official{}
		utils.CopyStructFields(&official, &officialAccount)
		officialAccountResp.Official = &official
		officialAccountResp.Interests = []*pbAdminCMS.InterestResp{}
		utils.CopyStructFields(&officialAccountResp.Interests, &officialAccount.Interests)
		resp.OfficialAccount = append(resp.OfficialAccount, &officialAccountResp)
	}
	log.Debug("", "resp.OfficialAccount len: ", len(resp.OfficialAccount))

	// Get total interests count
	official := &db.Official{}

	pendingWhere := map[string]string{}
	pendingWhere["process_status"] = "0"
	pendingCount, err := imdb.GetOfficialCountsByWhere(official, pendingWhere, []int64{})
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetOfficialCounts failed", err.Error(), officialCounts)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.PendingNums = pendingCount
	resp.OfficialNums = int32(officialCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcNews) DeleteOfficialAccounts(_ context.Context, req *pbNews.DeleteOfficialAccountsReq) (*pbNews.DeleteOfficialAccountsResp, error) {
	log.Debug(req.OperationID, "Delete official req: ", req.String())
	resp := &pbNews.DeleteOfficialAccountsResp{}
	if len(req.Officials) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	client, err := admin_cms.GetTencentCloudClient(true)
	if err != nil {
		log.NewError("", "get tencent cloud error", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrServer)
	}
	for _, official := range req.Officials {
		officialID, err := strconv.ParseInt(official, 10, 64)
		if err != nil {
			return resp, openIMHttp.WrapError(constant.ErrArgs)
		}
		officialObj, err := imdb.GetOfficialByOfficialID(officialID)
		if err != nil {
			return resp, openIMHttp.WrapError(constant.ErrArgs)
		}
		err = admin_cms.DeleteFileForPersistent(client, officialObj.FaceURL)
		if err != nil {
			log.NewError("", "delete official face url failed: ", officialObj.FaceURL)
		}

		count, err := imdb.ClearAllFollowers(officialID)
		if err != nil {
			return resp, openIMHttp.WrapError(constant.ErrArgs)
		}
		if count == 0 {
			log.NewError("", "clear official count = 0")
		}

		err = db.DB.ClearAllFollowers(officialID, req.OpUserId)
		if err != nil {
			return resp, openIMHttp.WrapError(constant.ErrArgs)
		}
	}

	if row := imdb.DeleteOfficialAccounts(req.Officials, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	for _, official := range req.Officials {
		officialID, err := strconv.ParseInt(official, 10, 64)
		if err == nil {
			if err = imdb.DeleteAllArticlesByOfficialId(officialID, req.OpUserId); err != nil {
				errMsg := req.OperationID + " imdb. All Delete Article failed " + err.Error() + req.OpUserId
				log.NewError(req.OperationID, errMsg)
			}
			if err = db.DB.DeleteAllArticlesByOfficialId(officialID, req.OpUserId); err != nil {
				errMsg := req.OperationID + " db.DB.DeleteArticle failed " + err.Error() + req.OpUserId
				log.NewError(req.OperationID, errMsg)
			}
			err := db.DB.RemoveLatestArticlesByOfficialAccount(officialID)
			if err != nil {
				log.NewError("UpdateRedisWithLatestArticlesByOfficialAccount", err.Error())
			}
		}
	}

	return resp, nil
}

func (rpc *rpcNews) AlterOfficialAccount(_ context.Context, req *pbNews.AlterOfficialAccountReq) (*pbNews.AlterOfficialAccountResp, error) {
	resp := &pbNews.AlterOfficialAccountResp{CommonResp: &pbNews.CommonResponse{}}
	official := req.Official
	log.Debug("", "alter official account: ", req.Official)

	oldOfficial, err := imdb.GetOfficialByOfficialID(req.Official.Id)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "official account isn't exist"
		return resp, nil
	}
	if oldOfficial.Id == 0 {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "official account isn't exist"
		return resp, nil
	}

	officialFind, err := imdb.GetOfficialByName(req.Official.Type, req.Official.Nickname)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "official name is already existed"
		return resp, nil
	}
	if officialFind.Id != 0 && officialFind.Id != oldOfficial.Id {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "official name is already existed"
		return resp, nil
	}

	IsAvailable := imdb.CheckOfficialIDNumberAvailable(req.Official.IdNumber, int8(req.Official.IdType), req.Official.Id)
	if IsAvailable {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "id number is already registered"
		return resp, nil
	}

	if req.Official.FaceURL != "" && oldOfficial.FaceURL != req.Official.FaceURL {
		client, err := admin_cms.GetTencentCloudClient(true)
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
			return resp, err
		}
		failedList, err := admin_cms.RemoveDeleteTagForPersistent(client, []string{req.Official.FaceURL})
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
			return resp, err
		}

		if len(failedList) != 0 {
			resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
			resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
			log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
			return resp, errors.New("upload profile photo failed")
		}
	}

	if row := imdb.AlterOfficialAccount(official, req.Interests); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		if req.Official.IsSystem == 1 {
			// if set to system
			if oldOfficial.IsSystem == 0 {

				allUserCount, err := imdb.GetAllUserCount()
				if err != nil {
					log.NewError("", "get all user failed")
					return
				}
				// allFollowers, err := db.DB.GetAllOfficialFollowersData(official.Id)
				allFollowers, err := imdb.GetAllOfficialFollowersData(official.Id)
				if err != nil {
					log.NewError("", "GetAllOfficialFollowersData failed", err)
					return
				}

				allUser, err := imdb.GetAllUserExcludeUser(allFollowers)

				err = addSystemOfficialAccount(official.Id, allUserCount, allUser, req.OperationID)
				if err != nil {
					log.NewError("", "addSystemOfficialAccount failed", err)
					return
				}

				err = imdb.DeleteOfficialInterest(oldOfficial.Id)
				if err != nil {
					errMsg := req.OperationID + "delete official interest failed"
					log.NewError(req.OperationID, errMsg)
				}

			}
		} else if oldOfficial.IsSystem == 1 {
			count, err := imdb.ClearAllFollowers(official.Id)
			if err != nil {
				return
			}
			if count == 0 {
				log.NewError("", "clear official count = 0", official.Id)
			}

			err = db.DB.ClearAllFollowers(official.Id, req.OpUserID)
			if err != nil {
				log.NewError("", "clear official count = 0", official.Id)
			}
		}
	}()

	return resp, nil
}

func (rpc *rpcNews) AddOfficialAccount(_ context.Context, req *pbNews.AddOfficialAccountReq) (*pbNews.AddOfficialAccountResp, error) {
	log.Debug(req.OperationID, "AddOfficialAccount req: ", req.String())
	resp := &pbNews.AddOfficialAccountResp{CommonResp: &pbNews.CommonResponse{}}
	official := &db.Official{}
	utils.CopyStructFields(official, req)
	official.FaceURL = req.ProfilePhoto

	// check id, name, account
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Add official failed:", err)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "user id doesn't exist"
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}
	if user == nil {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "user doesn't exist"
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}
	if user.OfficialID != 0 {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "this user has been registered an official"
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}

	officialFind, err := imdb.GetOfficialByName(req.Type, req.InitialNickname)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = "official is already existed"
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}
	if officialFind.Id != 0 {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "InitialNickname"
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}

	if req.Nickname != "" {
		officialFind, err = imdb.GetOfficialByName(req.Type, req.Nickname)
		if err != nil {
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommonResp.ErrMsg = "official initial nickname is already existed"
			log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
			return resp, nil
		}
		if officialFind.Id != 0 {
			resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
			resp.CommonResp.ErrMsg = "Nickname"
			log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
			return resp, nil
		}
	}

	IsAvailable := imdb.CheckOfficialIDNumberAvailable(req.IdNumber, int8(req.IdType), 0)
	if IsAvailable {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = "id number is already existed"
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}

	if req.ProfilePhoto == "" {
		resp.CommonResp.ErrCode = constant.ErrArgs.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrArgs.ErrMsg
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, nil
	}

	client, err := admin_cms.GetTencentCloudClient(true)
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, err
	}
	failedList, err := admin_cms.RemoveDeleteTagForPersistent(client, []string{req.ProfilePhoto})
	if err != nil {
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, err
	}

	if len(failedList) != 0 {
		resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, errors.New("upload profile photo failed")
	}

	if err = imdb.AddOfficialAccount(official, req.Interests, user); err != nil {
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, resp.CommonResp.ErrMsg)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {

		if req.IsSystem == 1 {
			allUserCount, err := imdb.GetAllUserCount()
			if err != nil {
				log.NewError("", "get all user failed")
				return
			}
			// allFollowers, err := db.DB.GetAllOfficialFollowersData(official.Id)
			allFollowers, err := imdb.GetAllOfficialFollowersData(official.Id)
			if err != nil {
				log.NewError("", "GetAllOfficialFollowersData failed", err)
				return
			}

			allUser, err := imdb.GetAllUserExcludeUser(allFollowers)

			err = addSystemOfficialAccount(official.Id, allUserCount, allUser, req.OperationID)
			if err != nil {
				log.NewError("", "addSystemOfficialAccount failed", err)
				return
			}
		} else {
			err := db.DB.AddOfficialFollow(official.Id, user.UserID)
			if err != nil {
				log.Debug("add official follow failed")
			}
		}

		if imdb.InsertThirdInfo(constant.OauthTypeOfficial, utils.Int64ToString(official.Id), official.Nickname, official.UserID) == false {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "imdb.InsertThirdInfo official failed")
		}
	}()

	return resp, nil
}

func (rpc *rpcNews) Process(_ context.Context, req *pbNews.ProcessReq) (*pbNews.ProcessResp, error) {
	resp := &pbNews.ProcessResp{}

	if req.ProcessStatus == 2 && req.ProcessFeedback == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Feedback shouldn't be empty!")
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	params := imdb.OfficialProcessParams{
		OpUserId:        req.OpUserId,
		OfficialId:      req.OfficialId,
		ProcessStatus:   req.ProcessStatus,
		ProcessFeedback: req.ProcessFeedback,
	}

	official, err := imdb.GetOfficialByOfficialID(req.OfficialId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "process failed")
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	if row := imdb.Process(&params); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "process failed", "process rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	if imdb.InsertThirdInfo(constant.OauthTypeOfficial, utils.Int64ToString(official.Id), official.Nickname, official.UserID) == false {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "imdb.InsertThirdInfo official failed")
	}

	return resp, nil
}

func (rpc *rpcNews) LikeArticle(_ context.Context, req *pbNews.LikeArticleRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	//insert in Mongo DB
	if err = db.DB.AddArticleLike(req.ArticleID, req.UserID); err != nil {
		errMsg := "db.DB.AddArticleLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	log.Debug("", "err: ", err)
	go func() {
		article, err := imdb.GetArticleByArticleID(req.ArticleID)
		if err != nil {
			errMsg := req.OperationID + " imdb.GetArticleByArticleID failed " + err.Error() + req.UserID
			log.NewError(req.OperationID, errMsg)
			return
		}

		//insert in MYSQL DB
		if err = imdb.AddArticleLike(article.OfficialID, req.ArticleID, req.UserID, user.Gender); err != nil {
			errMsg := "imdb.AddArticleLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return
		}
		UpdateArticleCount(true, false, article.ArticleID, 1, 0, 0, 0, 0)

	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) UnlikeArticle(_ context.Context, req *pbNews.UnlikeArticleRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err = db.DB.DeleteArticleLike(req.ArticleID, req.UserID, req.UserID); err != nil {
		errMsg := "db.DB.DeleteArticleLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {
		article, err := imdb.GetArticleByArticleID(req.ArticleID)
		if err != nil {
			errMsg := req.OperationID + " imdb.GetArticleByArticleID failed " + err.Error() + req.UserID
			log.NewError(req.OperationID, errMsg)
			return
		}

		if err = imdb.DeleteArticleLike(article.OfficialID, req.ArticleID, req.UserID, user.Gender, req.UserID); err != nil {
			errMsg := "imdb.DeleteArticleLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
		}

		UpdateArticleCount(true, false, article.ArticleID, -1, 0, 0, 0, 0)
	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) DeleteArticleLike(_ context.Context, req *pbNews.DeleteArticleLikeRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.OpUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.OpUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	article, err := db.DB.GetArticlesByID(req.ArticleID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetArticleByArticleID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	// return if the official account is not the owner of article
	if user.OfficialID != article.OfficialID {
		return &pbNews.CommonResponse{}, nil
	}

	if err = db.DB.DeleteArticleLike(req.ArticleID, req.UserID, req.OpUserID); err != nil {
		errMsg := "db.DB.DeleteArticleLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {
		if err = imdb.DeleteArticleLike(article.OfficialID, req.ArticleID, req.UserID, user.Gender, req.OpUserID); err != nil {
			errMsg := "imdb.DeleteArticleLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
		}
		UpdateArticleCount(true, false, article.ArticleID, -1, 0, 0, 0, 0)
	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) AddArticleComment(_ context.Context, req *pbNews.AddArticleCommentRequest) (*pbNews.AddArticleCommentResponse, error) {
	article, err := imdb.GetArticleByArticleID(req.ArticleID)
	if err != nil {
		errMsg := "imdb.GetArticleByArticleID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.AddArticleCommentResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	user, err := imdb.GetUserByUserID(req.OpUserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.AddArticleCommentResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	var officialID int64
	if req.UserID == "" {
		if user.OfficialID == 0 {
			return &pbNews.AddArticleCommentResponse{
				CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg},
			}, nil
		}
		officialID = user.OfficialID
	} else {

	}

	commentID, err := imdb.AddArticleComment(req.ArticleID, article.OfficialID, officialID, req.UserID, user.Gender, req.ParentCommentID, req.ReplyOfficialID, req.ReplyUserID, req.Content, req.OpUserID)
	if err != nil {
		errMsg := "imdb.AddArticleComment failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.AddArticleCommentResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	err = db.DB.AddArticleComment(commentID, req.ArticleID, officialID, req.UserID, req.ParentCommentID, req.ReplyOfficialID, req.ReplyUserID, req.Content, req.OpUserID)
	if err != nil {
		errMsg := "db.DB.AddArticleComment failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.AddArticleCommentResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}
	go func() {
		if req.ParentCommentID == 0 {
			// no parent comment
			UpdateArticleCount(true, false, article.ArticleID, 0, 1, 0, 0, 0)
		}
	}()

	return &pbNews.AddArticleCommentResponse{CommentID: commentID}, nil
}

func (rpc *rpcNews) ListArticlesTimeLine(_ context.Context, req *pbNews.ListArticlesTimeLineRequest) (*pbNews.ListArticlesTimeLineResponse, error) {

	// Get following articles or recommend articles, Get followed official account by user first(all unblock official)
	officialFollows, err := db.DB.GetSelfOfficialAccountFollows(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetSelfOfficialAccountFollows failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticlesTimeLineResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}
	log.Debug("", "get follows: ", officialFollows)
	var officialIDList = make([]int64, len(officialFollows))
	var blockedOfficialIDList []int64
	officialFollowsMap := make(map[int64]db.OfficialFollow, len(officialFollows))
	for index, officialFollow := range officialFollows {
		officialFollowsMap[officialFollow.OfficialID] = officialFollow
		officialIDList[index] = officialFollow.OfficialID
	}
	// var systemOfficialList []int64 = nil
	var xlinkOfficialID int64 = 0

	blockedOfficialList, err := db.DB.GetBlockedOfficialFollowList(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserInterestList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticlesTimeLineResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}
	blockedOfficialIDList = make([]int64, len(blockedOfficialList))
	for index, follow := range blockedOfficialList {
		blockedOfficialIDList[index] = follow.OfficialID
	}

	log.Debug("", "ListArticlesTimeLine official id: ", req.OfficialID)
	log.Debug("", "ListArticlesTimeLine Source: ", req.Source)
	log.Debug("", "officialFollows: ", officialFollows)
	if req.OfficialID == 0 {
		// if list recommend, from unblocked officials list select the last 3 ,condition: interest or follow list
		if req.Source == 1 {
			//interests, err := imdb.GetUserInterestList(req.UserID)
			//if err != nil {
			//	errMsg := req.OperationID + " imdb.GetUserInterestList failed " + err.Error() + req.UserID
			//	log.NewError(req.OperationID, errMsg)
			//	return &pbNews.ListArticlesTimeLineResponse{
			//		CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
			//	}, nil
			//}
			//if len(interests) == 1 && interests[0] == -1 {
			//	interests = imdb.GetDefaultInterestTypeList()
			//}
			//
			//log.Debug("", "interests: ", interests)
			//blockedOfficialList, err := db.DB.GetBlockedOfficialFollowList(req.UserID)
			//if err != nil {
			//	errMsg := req.OperationID + " imdb.GetBlockedOfficialFollowList failed " + err.Error() + req.UserID
			//	log.NewError(req.OperationID, errMsg)
			//}
			//blockedOfficialIDList := make([]int64, len(blockedOfficialList))
			//for index, follow := range blockedOfficialList {
			//	blockedOfficialIDList[index] = follow.OfficialID
			//}
			//
			//officialIDList, err = imdb.GetOfficialIDListByInterestAndFollowList(officialIDList, blockedOfficialIDList, interests)
			//if err != nil {
			//	errMsg := req.OperationID + " imdb.GetOfficialIDListByInterestList failed " + err.Error() + req.UserID
			//	log.NewError(req.OperationID, errMsg)
			//	return &pbNews.ListArticlesTimeLineResponse{
			//		CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
			//	}, nil
			//}
			// recommendation
			officialIDList = nil
			xlinkOfficial, err := imdb.GetOfficialByName(config.Config.Official.SystemOfficialType, config.Config.Official.SystemOfficialName)
			if err != nil {
				return nil, err
			}
			xlinkOfficialID = xlinkOfficial.Id
			//systemOfficialList, err = imdb.GetSystemOfficialIDList(req.Limit)
			//if err != nil {
			//	errMsg := req.OperationID + " GetSystemOfficialIDList failed " + err.Error() + req.UserID
			//	log.NewError(req.OperationID, errMsg)
			//	return &pbNews.ListArticlesTimeLineResponse{
			//		CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
			//	}, nil
			//}
		} else {
			officialIDList = make([]int64, len(officialFollows))
			for index, officialFollow := range officialFollows {
				log.Debug("", "officialFollow.OfficialID:", officialFollow.OfficialID)
				officialIDList[index] = officialFollow.OfficialID
			}
		}
	} else {
		officialIDList = []int64{req.OfficialID}
		//follow, err := db.DB.GetOfficialFollow(req.OfficialID, req.UserID)
		//if err != nil {
		//	errMsg := req.OperationID + " db.DB.GetOfficialFollow failed " + err.Error() + req.UserID
		//	log.NewError(req.OperationID, errMsg)
		//	return &pbNews.ListArticlesTimeLineResponse{
		//		CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		//	}, nil
		//}
		//if follow != nil && follow.BlockTime != 0 {
		//	// blocked
		//	return &pbNews.ListArticlesTimeLineResponse{
		//		CommonResp: &pbNews.CommonResponse{ErrCode: constant.OK.ErrCode, ErrMsg: "no record"},
		//	}, nil
		//} else {
		//	// unblocked or unfollowed
		//
		//}
	}

	deletedIdList, err := imdb.GetDeletedOfficialIDList()
	if err != nil {
		errMsg := req.OperationID + " imdb.GetDeletedOfficialIDList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticlesTimeLineResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	articles, count, err := db.DB.GetArticlesByFollowedOfficialIDListV2(xlinkOfficialID, officialIDList, blockedOfficialIDList, req.Offset, req.Limit, deletedIdList)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetArticlesByFollowedOfficialIDList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticlesTimeLineResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	for _, article := range articles {
		officialIDList = append(officialIDList, article.OfficialID)
	}

	// get official map
	officialsMap, err := imdb.GetOfficialsByOfficialIDList(officialIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialsByOfficialIDList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticlesTimeLineResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	entries := make([]*pbNews.ListArticlesTimeLineResponseEntry, len(articles))
	for i, article := range articles {
		entries[i] = &pbNews.ListArticlesTimeLineResponseEntry{
			Article: &pbNews.UserArticleSummary{
				ArticleID:        article.ArticleID,
				Title:            article.Title,
				CoverPhoto:       article.CoverPhoto,
				TextContent:      article.TextContent,
				ReadCounts:       article.ReadCounts,
				UniqueReadCounts: article.UniqueReadCounts,
				LikeCounts:       article.LikeCounts,
				CommentCounts:    article.CommentCounts,
				RepostCounts:     article.RepostCounts,
				CreateTime:       article.CreateTime,
			},
		}
		if official, exist := officialsMap[article.OfficialID]; exist {
			entries[i].Official = &pbNews.UserFollow{
				OfficialID: official.Id,
				Nickname:   official.Nickname,
				FaceURL:    official.FaceURL,
				Type:       int32(official.Type),
				Bio:        official.Bio,
			}

			if officialFollow, followExists := officialFollowsMap[article.OfficialID]; followExists {
				entries[i].Official.FollowTime = officialFollow.FollowTime
				entries[i].Official.Muted = officialFollow.Muted
				entries[i].Official.Enabled = officialFollow.Enabled
			}
		}
	}

	return &pbNews.ListArticlesTimeLineResponse{
		Entries: entries,
		Count:   count,
	}, nil
}

func (rpc *rpcNews) LikeArticleComment(_ context.Context, req *pbNews.LikeArticleCommentRequest) (*pbNews.CommonResponse, error) {
	//insert in Mongo DB
	var err error
	if err = db.DB.AddArticleCommentLike(req.CommentID, req.UserID, 0, req.UserID); err != nil {
		errMsg := "db.DB.AddArticleCommentLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {
		//insert in MYSQL DB
		if err := imdb.AddArticleCommentLike(req.CommentID, req.UserID, 0, req.UserID); err != nil {
			errMsg := "imdb.AddArticleCommentLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
		}
		go UpdateArticleCommentCount(true, false, req.CommentID, 0, 1)
	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) UnlikeArticleComment(_ context.Context, req *pbNews.UnlikeArticleCommentRequest) (*pbNews.CommonResponse, error) {

	if err := db.DB.DeleteArticleCommentLike(req.CommentID, req.UserID, 0, req.UserID); err != nil {
		errMsg := "db.DB.DeleteArticleCommentLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {
		if err := imdb.DeleteArticleCommentLike(req.CommentID, req.UserID, 0, req.UserID); err != nil {
			errMsg := "imdb.DeleteArticleCommentLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
		}
		go UpdateArticleCommentCount(true, false, req.CommentID, 0, -1)
	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) OfficialLikeArticleComment(_ context.Context, req *pbNews.OfficialLikeArticleCommentRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	//insert in MYSQL DB
	if err = imdb.AddArticleCommentLike(req.CommentID, "", user.OfficialID, user.UserID); err != nil {
		errMsg := "imdb.AddArticleCommentLike failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	//insert in Mongo DB
	if err = db.DB.AddArticleCommentLike(req.CommentID, "", user.OfficialID, user.UserID); err != nil {
		errMsg := "db.DB.AddArticleCommentLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go UpdateArticleCommentCount(true, false, req.CommentID, 0, 1)

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) OfficialUnlikeArticleComment(_ context.Context, req *pbNews.OfficialUnlikeArticleCommentRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	if err = db.DB.DeleteArticleCommentLike(req.CommentID, "", user.OfficialID, user.UserID); err != nil {
		errMsg := "db.DB.DeleteArticleCommentLike failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {

		if err = imdb.DeleteArticleCommentLike(req.CommentID, "", user.OfficialID, user.UserID); err != nil {
			errMsg := "imdb.DeleteArticleCommentLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return
		}
		UpdateArticleCommentCount(true, false, req.CommentID, 0, -1)

	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) OfficialDeleteArticleComment(_ context.Context, req *pbNews.OfficialDeleteArticleCommentRequest) (*pbNews.CommonResponse, error) {
	comment, err := imdb.GetCommentByCommentID(req.CommentID)
	if err != nil {
		errMsg := "imdb.GetCommentByCommentID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	article, err := imdb.GetArticleByArticleID(comment.ArticleID)
	if err != nil {
		errMsg := "imdb.GetArticleByArticleID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	opUser, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	var commentUser *db.User
	if comment.UserID != "" {
		commentUser, err = imdb.GetUserByUserID(comment.UserID)
		if err != nil {
			errMsg := "imdb.GetUserByUserID failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}
	}

	if comment.OfficialID != 0 {
		commentUser, err = imdb.GetUserByOfficialID(comment.OfficialID)
		if err != nil {
			errMsg := "imdb.GetUserByOfficialID failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}
	}

	if article.OfficialID != opUser.OfficialID && comment.OfficialID != opUser.OfficialID {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialActionForbidden.ErrCode, ErrMsg: constant.ErrOfficialActionForbidden.ErrMsg}, nil
	}

	if err = imdb.OfficialDeleteArticleComment(comment.CommentID, comment.ParentCommentID, article.ArticleID, article.OfficialID, commentUser.Gender, opUser.UserID); err != nil {
		errMsg := "imdb.OfficialDeleteArticleComment failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err = db.DB.OfficialDeleteArticleComment(req.CommentID, req.UserID); err != nil {
		errMsg := "db.DB.OfficialDeleteArticleComment failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {
		if comment.ParentCommentID == 0 {
			UpdateArticleCount(true, false, article.ArticleID, 0, -1, 0, 0, 0)
		}
	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) OfficialHideArticleComment(_ context.Context, req *pbNews.OfficialHideArticleCommentRequest) (*pbNews.CommonResponse, error) {
	comment, err := imdb.GetCommentByCommentID(req.CommentID)
	if err != nil {
		errMsg := "imdb.GetCommentByCommentID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	article, err := imdb.GetArticleByArticleID(comment.ArticleID)
	if err != nil {
		errMsg := "imdb.GetArticleByArticleID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	opUser, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if article.OfficialID != opUser.OfficialID && comment.OfficialID != opUser.OfficialID {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialActionForbidden.ErrCode, ErrMsg: constant.ErrOfficialActionForbidden.ErrMsg}, nil
	}

	if err = imdb.OfficialHideArticleComment(req.CommentID, req.UserID); err != nil {
		errMsg := "imdb.OfficialHideArticleComment failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err = db.DB.OfficialHideArticleComment(req.CommentID, req.UserID); err != nil {
		errMsg := "db.DB.OfficialHideArticleComment failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) OfficialShowArticleComment(_ context.Context, req *pbNews.OfficialShowArticleCommentRequest) (*pbNews.CommonResponse, error) {
	comment, err := imdb.GetCommentByCommentID(req.CommentID)
	if err != nil {
		errMsg := "imdb.GetCommentByCommentID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	article, err := imdb.GetArticleByArticleID(comment.ArticleID)
	if err != nil {
		errMsg := "imdb.GetArticleByArticleID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	opUser, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if article.OfficialID != opUser.OfficialID && comment.OfficialID != opUser.OfficialID {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialActionForbidden.ErrCode, ErrMsg: constant.ErrOfficialActionForbidden.ErrMsg}, nil
	}

	if err = imdb.OfficialShowArticleComment(req.CommentID, req.UserID); err != nil {
		errMsg := "imdb.OfficialShowArticleComment failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err = db.DB.OfficialShowArticleComment(req.CommentID, req.UserID); err != nil {
		errMsg := "db.DB.OfficialShowArticleComment failed " + err.Error()
		log.Error(errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) GetNews(_ context.Context, req *pbNews.GetNewsReq) (*pbNews.GetNewsResp, error) {
	resp := &pbNews.GetNewsResp{Articles: []*pbNews.Article{}}
	where := map[string]string{}
	where["official_account"] = req.OfficialAccount
	where["account_type"] = strconv.Itoa(int(req.AccountType))
	where["ip"] = req.Ip
	where["title"] = req.Title
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	dbNewsList, articleCounts, err := imdb.GetNewsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}
	err = utils.CopyStructFields(&resp.Articles, &dbNewsList)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Copy Struct Fields failed", err.Error())
		return nil, err
	}

	resp.ArticlesNums = int32(articleCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcNews) DeleteNews(_ context.Context, req *pbNews.DeleteNewsReq) (*pbNews.DeleteNewsResp, error) {
	log.Debug(req.OperationID, "Delete official req: ", req.String())
	resp := &pbNews.DeleteNewsResp{}
	if len(req.Articles) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	var articles []*db.Article
	for _, article := range req.Articles {
		art, err := db.DB.GetArticlesByID(article)
		if err != nil {
			return resp, err
		}
		articles = append(articles, art)
	}
	err := removeArticleFiles(articles)
	if err != nil {
		return resp, err
	}

	if row := imdb.DeleteArticles(req.Articles, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		nowTime := time.Now().Unix()
		for _, article := range articles {
			article.DeletedBy = req.OpUserId
			article.DeleteTime = nowTime
			err = db.DB.DeleteArticle(*article)
			if err != nil {
				log.Error("", "delete article error, article: ", article.ArticleID)
			}
		}
	}()

	return resp, nil
}

func (rpc *rpcNews) AlterNews(_ context.Context, req *pbNews.AlterNewsReq) (*pbNews.AlterNewsResp, error) {
	log.Debug(req.OperationID, "AlterNews req: ", req.String())
	resp := &pbNews.AlterNewsResp{CommonResp: &pbNews.CommonResponse{}}

	articleDB := &db.ArticleSQL{
		ArticleID: req.ArticleId,
		Title:     req.Title,
		Content:   req.Content,
	}

	originalArticle, err := db.DB.GetArticlesByID(req.ArticleId)
	if err != nil {
		resp.CommonResp.ErrMsg = "get article failed"
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		return resp, err
	}

	article := db.Article{}
	err = utils.CopyStructFields(&article, articleDB)
	if err != nil {
		log.Debug("", "copy article to mongo db article failed")
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	originalMediaMap := getArticleMediaUrlMap([]*db.Article{originalArticle})
	newMediaMap := getArticleMediaUrlMap([]*db.Article{&article})
	var oldRemoveMedias []string
	var newAddMedias []string
	for url, _ := range originalMediaMap {
		if _, ok := newMediaMap[url]; !ok {
			// the new medias don't have this original media, then need to delete this media
			oldRemoveMedias = append(oldRemoveMedias, url)
		}
	}
	for url, _ := range newMediaMap {
		if _, ok := originalMediaMap[url]; !ok {
			// the original media don't have the new media, need to remove the tag.
			newAddMedias = append(newAddMedias, url)
		}
	}

	if len(oldRemoveMedias) > 0 {
		// remove these files.
		err = admin_cms.RemoveFiles(oldRemoveMedias)
		if err != nil {
			resp.CommonResp.ErrMsg = "remove article file failed"
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			return resp, err
		}
	}
	if len(newAddMedias) > 0 {
		// remove these file tags
		client, err := admin_cms.GetTencentCloudClient(true)
		if err != nil {
			errMsg := req.OperationID + " get cloud link failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			resp.CommonResp.ErrMsg = errMsg
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			return resp, err
		}
		failList, err := admin_cms.RemoveDeleteTagForPersistent(client, newAddMedias)
		if err != nil {
			log.Error("", "Upload article file failed: ", err.Error())
			resp.CommonResp.ErrMsg = "Upload article file failed"
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			return resp, err
		}
		if len(failList) > 0 {
			log.Error("", "failed file list: ", failList)
			resp.CommonResp.ErrMsg = "Upload article file failed"
			resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
			return resp, err
		}
	}

	if row := imdb.AlterArticle(articleDB, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	originalArticle.Title = req.Title
	originalArticle.Content = req.Content
	originalArticle.UpdatedBy = req.OpUserId
	originalArticle.UpdateTime = time.Now().Unix()
	err = db.DB.UpdateArticle(*originalArticle)
	if err != nil {
		log.Debug("", "update article on mongo db failed")
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (rpc *rpcNews) ChangePrivacy(_ context.Context, req *pbNews.ChangePrivacyReq) (*pbNews.ChangePrivacyResp, error) {
	log.Debug(req.OperationID, "ChangePrivacy req: ", req.String())
	resp := &pbNews.ChangePrivacyResp{}

	article := &db.ArticleSQL{
		ArticleID: req.ArticleId,
		Privacy:   req.Privacy,
	}
	if row := imdb.AlterArticle(article, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "change failed", "change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	go func() {
		articleMongo, err := db.DB.GetArticlesByID(req.ArticleId)
		if err != nil {
			log.Debug("", "copy article to mongo db article failed")
			return
		}
		articleMongo.Privacy = req.Privacy
		articleMongo.UpdatedBy = req.OpUserId
		articleMongo.UpdateTime = time.Now().Unix()
		err = db.DB.ChangeArticlePrivacy(*articleMongo)
		if err != nil {
			log.Debug("", "update article on mongo db failed")
			return
		}
	}()
	return resp, nil

}

func (rpc *rpcNews) GetNewsComments(_ context.Context, req *pbNews.GetNewsCommentsReq) (*pbNews.GetNewsCommentsResp, error) {
	resp := &pbNews.GetNewsCommentsResp{Comments: []*pbNews.ArticleComment{}}
	where := map[string]string{}
	where["official_account"] = req.OfficialAccount
	where["account_type"] = strconv.Itoa(int(req.AccountType))
	where["ip"] = req.Ip
	where["title"] = req.Title
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["comment_user"] = req.CommentUser
	where["comment_key"] = req.CommentKey
	where["article_id"] = strconv.FormatInt(req.ArticleId, 10)

	dbCommentsList, commentsCounts, err := imdb.GetArticleCommentsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}
	err = utils.CopyStructFields(&resp.Comments, &dbCommentsList)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Copy Struct Fields failed", err.Error())
		return nil, err
	}
	log.Debug("", "resp.OfficialAccount len: ", len(resp.Comments))

	resp.CommentsNums = int32(commentsCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcNews) RemoveNewsComments(_ context.Context, req *pbNews.RemoveNewsCommentsReq) (*pbNews.RemoveNewsCommentsResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	resp := &pbNews.RemoveNewsCommentsResp{}
	if len(req.Comments) == 0 {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	if row := imdb.DeleteArticleComments(req.Comments, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		// remove from mongodb
		for index, comment := range req.Comments {
			articleComment := db.ArticleComment{}
			articleComment.CommentID = comment

			err := db.DB.DeleteArticleComment(articleComment)
			if err != nil {
				log.Error("", "delete article comment from mongo db failed, article: ", articleComment.CommentID)
			}

			if req.Parents[index] != "" {
				UpdateArticleCount(true, true, req.Articles[index], 0, -1, 0, 0, 0)
			}
		}

	}()

	return resp, nil
}

func (rpc *rpcNews) AlterNewsComment(_ context.Context, req *pbNews.AlterNewsCommentReq) (*pbNews.AlterNewsCommentResp, error) {
	log.Debug(req.OperationID, "AlterComment req: ", req.String())
	resp := &pbNews.AlterNewsCommentResp{}

	comment := &db.ArticleCommentSQL{}
	comment.CommentID, _ = strconv.ParseInt(req.CommentId, 10, 64)
	comment.UserID = req.UserId
	comment.Content = req.Content
	if comment.CommentID == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed, param is error")
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	if row := imdb.AlterNewsComment(comment, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "alter failed", "alter rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (rpc *rpcNews) ChangeNewsCommentStatus(_ context.Context, req *pbNews.ChangeNewsCommentStatusReq) (*pbNews.ChangeNewsCommentStatusResp, error) {
	log.Debug(req.OperationID, "ChangeNewsCommentStatus req: ", req.String())
	resp := &pbNews.ChangeNewsCommentStatusResp{}

	comment := &db.ArticleCommentSQL{}
	comment.CommentID = req.CommentId
	comment.Status = req.Status
	if row := imdb.AlterNewsComment(comment, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "change failed", "change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		err := db.DB.UpdateArticleComment(req.CommentId, req.Status)
		if err != nil {
			log.NewError("", "update article comment status failed")
		}
	}()

	return resp, nil
}

func (rpc *rpcNews) GetNewsLikes(_ context.Context, req *pbNews.GetNewsLikesReq) (*pbNews.GetNewsLikesResp, error) {

	resp := &pbNews.GetNewsLikesResp{Likes: []*pbNews.ArticleLike{}}
	where := map[string]string{}
	where["official_account"] = req.OfficialAccount
	where["account_type"] = strconv.Itoa(int(req.AccountType))
	where["ip"] = req.Ip
	where["title"] = req.Title
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["like_user"] = req.LikeUser
	where["article_id"] = strconv.FormatInt(req.ArticleId, 10)

	dbLikesList, likesCounts, err := imdb.GetLikesByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}
	err = utils.CopyStructFields(&resp.Likes, &dbLikesList)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Copy Struct Fields failed", err.Error())
		return nil, err
	}

	// Get total interests count
	resp.LikeNums = int32(likesCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil

}

func (rpc *rpcNews) RemoveNewsLikes(_ context.Context, req *pbNews.RemoveNewsLikesReq) (*pbNews.RemoveNewsLikesResp, error) {
	log.Debug(req.OperationID, "remove comments req: ", req.String())
	resp := &pbNews.RemoveNewsLikesResp{}
	if len(req.UserIds) == 0 || len(req.Articles) != len(req.UserIds) {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}

	if row := imdb.DeleteArticleLikes(req.UserIds, req.Articles, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go func() {
		for _, art := range req.Articles {
			UpdateArticleCount(true, true, art, -1, 0, 0, 0, 0)
		}
	}()

	return resp, nil
}

func (rpc *rpcNews) ChangeNewsLikeStatus(_ context.Context, req *pbNews.ChangeNewsLikeStatusReq) (*pbNews.ChangeNewsLikeStatusResp, error) {
	log.Debug(req.OperationID, "ChangeNewsLikeStatus req: ", req.String())
	resp := &pbNews.ChangeNewsLikeStatusResp{}

	like := &db.ArticleLikeSQL{}
	like.ArticleID = req.ArticleId
	like.UserID = req.UserId
	like.Status = req.Status
	if row := imdb.ChangeArticleLikes(like, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "change failed", "change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (rpc *rpcNews) GetRepostArticles(_ context.Context, req *pbNews.GetRepostArticlesReq) (*pbNews.GetRepostArticlesResp, error) {

	resp := &pbNews.GetRepostArticlesResp{Reposts: []*pbNews.ArticleRepost{}}
	where := map[string]string{}
	where["repost_user"] = req.RepostUser
	where["account_type"] = strconv.Itoa(int(req.AccountType))
	where["ip"] = req.Ip
	where["title"] = req.Title
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["original_user"] = req.OriginalUser
	where["article_id"] = strconv.FormatInt(req.ArticleId, 10)

	dbRepostList, repostCounts, err := imdb.GetRepostArticles(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}
	err = utils.CopyStructFields(&resp.Reposts, &dbRepostList)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Copy Struct Fields failed", err.Error())
		return nil, err
	}
	log.Debug("", "resp.Likes len: ", len(resp.Reposts))

	resp.RepostNums = int32(repostCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil

}

func (rpc *rpcNews) CreateArticle(_ context.Context, req *pbNews.CreateArticleReq) (*pbNews.CommonResponse, error) {
	now := time.Now().Unix()
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	// user have no official account
	if user.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	official, err := imdb.GetOfficialByOfficialID(user.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialByOfficialID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	// official account not verified
	if official.ProcessStatus != 1 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}
	articleSql := db.ArticleSQL{
		OfficialID:         official.Id,
		OfficialName:       official.Nickname,
		OfficialProfileImg: official.FaceURL,
		CoverPhoto:         req.CoverPhoto,
		Title:              req.Title,
		Content:            req.Content,
		TextContent:        req.TextContent,
		CreatedBy:          req.UserID,
		CreateTime:         now,
	}

	article := db.Article{}
	if err = utils.CopyStructFields(&article, articleSql); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	urls := getArticleMediaUrls([]*db.Article{&article})
	urls = append(urls, article.CoverPhoto)
	client, err := admin_cms.GetTencentCloudClient(true)
	if err != nil {
		errMsg := req.OperationID + " get cloud link failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	failList, err := admin_cms.RemoveDeleteTagForPersistent(client, urls)
	if err != nil {
		log.Error("", "Upload article file failed: ", err.Error())
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Upload article file failed"}, err
	}
	if len(failList) > 0 {
		log.Error("", "failed file list: ", failList)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Upload article file failed"}, err
	}

	// insert to mysql
	if err = imdb.InsertArticle(&articleSql); err != nil {
		errMsg := req.OperationID + " imdb.InsertArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	//insert in redis
	_ = db.DB.SetLatestArticlesByOfficialAccount(articleSql)

	mongoArticle := db.Article{}
	if err = utils.CopyStructFields(&mongoArticle, articleSql); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	// insert to mongodb
	if err = db.DB.AddArticle(mongoArticle); err != nil {
		errMsg := req.OperationID + " db.DB.AddArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) UpdateArticle(_ context.Context, req *pbNews.UpdateArticleReq) (*pbNews.CommonResponse, error) {
	now := time.Now().Unix()
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	articleSql := db.ArticleSQL{
		ArticleID:   req.ArticleID,
		OfficialID:  user.OfficialID,
		Title:       req.Title,
		CoverPhoto:  req.CoverPhoto,
		Content:     req.Content,
		TextContent: req.TextContent,
		UpdateTime:  now,
		UpdatedBy:   req.UserID,
	}

	article := db.Article{}
	if err = utils.CopyStructFields(&article, articleSql); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	originalArticle, err := db.DB.GetArticlesByID(req.ArticleID)
	if err != nil {
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "get article failed"}, err
	}

	originalMediaMap := getArticleMediaUrlMap([]*db.Article{originalArticle})
	newMediaMap := getArticleMediaUrlMap([]*db.Article{&article})
	var oldRemoveMedias []string
	var newAddMedias []string
	for url, _ := range originalMediaMap {
		if _, ok := newMediaMap[url]; !ok {
			// the new medias don't have this original media, then need to delete this media
			oldRemoveMedias = append(oldRemoveMedias, url)
		}
	}
	for url, _ := range newMediaMap {
		if _, ok := originalMediaMap[url]; !ok {
			// the original media don't have the new media, need to remove the tag.
			newAddMedias = append(newAddMedias, url)
		}
	}
	if originalArticle.CoverPhoto != req.CoverPhoto {
		oldRemoveMedias = append(oldRemoveMedias, req.CoverPhoto)
		newAddMedias = append(newAddMedias, req.CoverPhoto)
	}

	if len(oldRemoveMedias) > 0 {
		// remove these files.
		err = admin_cms.RemoveFiles(oldRemoveMedias)
		if err != nil {
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "remove article file failed"}, err
		}
	}
	if len(newAddMedias) > 0 {
		// remove these file tags
		client, err := admin_cms.GetTencentCloudClient(true)
		if err != nil {
			errMsg := req.OperationID + " get cloud link failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
		}
		failList, err := admin_cms.RemoveDeleteTagForPersistent(client, newAddMedias)
		if err != nil {
			log.Error("", "Upload article file failed: ", err.Error())
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Upload article file failed"}, err
		}
		if len(failList) > 0 {
			log.Error("", "failed file list: ", failList)
			return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "Upload article file failed"}, err
		}
	}

	rowsAffected, err := imdb.UpdateArticle(&articleSql)
	if err != nil {
		errMsg := req.OperationID + " imdb.UpdateArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if rowsAffected == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "update article failed"}, nil
	}

	if err = db.DB.UpdateArticle(article); err != nil {
		errMsg := req.OperationID + " db.DB.UpdateArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) DeleteArticle(_ context.Context, req *pbNews.DeleteArticleReq) (*pbNews.CommonResponse, error) {
	now := time.Now().Unix()
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	articleSql, err := imdb.GetOfficialArticle(req.ArticleID, user.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	articleSql.DeletedBy = req.UserID
	articleSql.DeleteTime = now

	article := db.Article{}
	if err = utils.CopyStructFields(&article, articleSql); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	err = removeArticleFiles([]*db.Article{&article})
	if err != nil {
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: "remove article files error"}, err
	}

	if err = imdb.DeleteArticle(&articleSql); err != nil {
		errMsg := req.OperationID + " imdb.DeleteArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err = db.DB.DeleteArticle(article); err != nil {
		errMsg := req.OperationID + " db.DB.DeleteArticle failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	go func() {
		//update the latest article  in redis if the latest one is deleted
		articleInRedis, err := db.DB.GetLatestArticlesByOfficialAccount(article.OfficialID)
		if err == nil && articleInRedis != nil {
			if articleInRedis.ArticleID == article.ArticleID {
				article, err := imdb.GetLatestArticleByOfficialID(article.OfficialID)
				if err == nil && article.Title != "" {
					_ = db.DB.SetLatestArticlesByOfficialAccount(article)
				}
			}

		}

	}()

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) ListOfficialArticles(_ context.Context, req *pbNews.ListOfficialArticlesReq) (*pbNews.ListOfficialArticlesResp, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListOfficialArticlesResp{
			CommonResp: &pbNews.CommonResponse{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  errMsg,
			},
		}, nil
	}

	resp := &pbNews.ListOfficialArticlesResp{}

	offset := int(req.Offset)
	limit := int(req.Limit)

	articles, count, err := imdb.ListOfficialArticles(user.OfficialID, req.MinCreateTime, offset, limit)
	if err != nil {
		errMsg := req.OperationID + " imdb.ListOfficialArticles failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListOfficialArticlesResp{
			CommonResp: &pbNews.CommonResponse{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  errMsg,
			},
		}, nil
	}

	resp.Articles = make([]*pbNews.ArticleSummary, len(articles))
	for i, article := range articles {
		resp.Articles[i] = &pbNews.ArticleSummary{}
		if err = utils.CopyStructFields(resp.Articles[i], article); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
	}

	resp.Count = count

	return resp, nil
}

func (rpc *rpcNews) GetOfficialArticle(_ context.Context, req *pbNews.GetOfficialArticleReq) (*pbNews.GetOfficialArticleResp, error) {
	articleSql, err := imdb.GetArticleByArticleID(req.ArticleID)
	if err == gorm.ErrRecordNotFound {
		return &pbNews.GetOfficialArticleResp{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialArticleNotExist.ErrCode, ErrMsg: constant.ErrOfficialArticleNotExist.ErrMsg},
		}, nil
	} else if err != nil {
		errMsg := req.OperationID + " imdb.GetArticleByArticleID failed " + err.Error() + utils.Int64ToString(req.ArticleID)
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialArticleResp{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	var article pbNews.Article
	if err = utils.CopyStructFields(&article, articleSql); err != nil {
		log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
	}

	return &pbNews.GetOfficialArticleResp{
		Article: &article,
	}, nil
}

func (rpc *rpcNews) FollowOfficialAccount(_ context.Context, req *pbNews.FollowOfficialAccountRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err = imdb.AddOfficialFollow(req.OfficialID, req.UserID, user.Gender); err != nil {
		errMsg := req.OperationID + " imdb.FollowOfficialAccount failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err = db.DB.AddOfficialFollow(req.OfficialID, req.UserID); err != nil {
		errMsg := req.OperationID + " db.DB.AddOfficialFollow failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	chat.OfficialAccountFollowUnfollowNotification(req.OperationID, req.UserID, strconv.FormatInt(req.OfficialID, 10))
	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) GetUserFollowList(_ context.Context, req *pbNews.GetUserFollowListRequest) (*pbNews.GetUserFollowListResponse, error) {
	follows, count, err := imdb.GetUserFollowList(req.UserID, int(req.Offset), int(req.Limit), req.Keyword)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserFollowList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetUserFollowListResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}
	followEntries := make([]*pbNews.UserFollow, len(follows))
	for i, follow := range follows {
		followEntries[i] = &pbNews.UserFollow{
			OfficialID: follow.OfficialID,
			Nickname:   follow.Nickname,
			FaceURL:    follow.FaceURL,
			Bio:        follow.Bio,
			Type:       follow.Type,
			FollowTime: follow.FollowTime,
			Muted:      follow.Muted,
			Enabled:    follow.Enabled,
		}
	}
	return &pbNews.GetUserFollowListResponse{
		Count:   count,
		Follows: followEntries,
	}, nil
}

func (rpc *rpcNews) UnfollowOfficialAccount(_ context.Context, req *pbNews.UnfollowOfficialAccountRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err = imdb.DeleteOfficialFollows(req.OfficialID, req.UserID, []db.User{*user}); err != nil {
		errMsg := req.OperationID + " imdb.DeleteOfficialFollows failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err = db.DB.DeleteOfficialFollows(req.OfficialID, req.UserID, []string{req.UserID}); err != nil {
		errMsg := req.OperationID + " db.DB.DeleteOfficialFollows failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	chat.OfficialAccountFollowUnfollowNotification(req.OperationID, req.UserID, strconv.FormatInt(req.OfficialID, 10))
	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) UpdateOfficialFollowSettings(_ context.Context, req *pbNews.UpdateOfficialFollowSettingsRequest) (*pbNews.CommonResponse, error) {
	if err := imdb.UpdateOfficialFollow(req.OfficialID, req.UserID, req.Muted, req.Enabled); err != nil {
		errMsg := req.OperationID + " imdb.UpdateOfficialFollow failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if err := db.DB.UpdateOfficialFollow(req.OfficialID, req.UserID, req.Muted, req.Enabled); err != nil {
		errMsg := req.OperationID + " db.DB.UpdateOfficialFollow failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}
func (rpc *rpcNews) GetOfficialFollowSettingsByOfficialID(_ context.Context, req *pbNews.OfficialFollowSettingsByOfficialIDRequest) (*pbNews.OfficialFollowSettingsByOfficialIDResponse, error) {
	resp := pbNews.OfficialFollowSettingsByOfficialIDResponse{}
	resp.CommonResp = new(pbNews.CommonResponse)
	officialFollow, err := imdb.GetOfficialFollowByOfficialAndUserID(req.OfficialID, req.ReqUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.UpdateOfficialFollow failed " + err.Error() + req.ReqUserID
		log.NewError(req.OperationID, errMsg)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = err.Error()
		resp.UserFollow = nil
		return &resp, err
	}
	if officialFollow != nil && officialFollow.OfficialID != 0 {
		resp.UserFollow = new(pbNews.UserFollow)
		resp.UserFollow.OfficialID = officialFollow.OfficialID
		resp.UserFollow.Muted = officialFollow.Muted
		resp.UserFollow.Enabled = officialFollow.Enabled
		resp.UserFollow.FollowTime = officialFollow.FollowTime

		resp.CommonResp.ErrCode = 0
		resp.CommonResp.ErrMsg = ""
		return &resp, nil

	}
	resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
	if err != nil {
		resp.CommonResp.ErrMsg = err.Error()
	}
	resp.UserFollow = nil
	return &resp, err
}

func (rpc *rpcNews) BlockOfficialFollows(_ context.Context, req *pbNews.BlockOfficialFollowsRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.OfficialUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	if err := imdb.AddOfficialFollowBlocks(user.OfficialID, req.OfficialUserID, req.UserIDList); err != nil {
		errMsg := req.OperationID + " imdb.AddOfficialFollowBlock failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err := db.DB.AddOfficialFollowBlock(user.OfficialID, req.OfficialUserID, req.UserIDList); err != nil {
		errMsg := req.OperationID + " db.DB.AddOfficialFollowBlock failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) UnblockOfficialFollows(_ context.Context, req *pbNews.UnblockOfficialFollowsRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.OfficialUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	if err := imdb.DeleteOfficialFollowBlocks(user.OfficialID, req.UserIDList); err != nil {
		errMsg := req.OperationID + " imdb.DeleteOfficialFollowBlock failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err := db.DB.DeleteOfficialFollowBlock(user.OfficialID, req.UserIDList); err != nil {
		errMsg := req.OperationID + " db.DB.DeleteOfficialFollowBlock failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) DeleteOfficialFollows(_ context.Context, req *pbNews.DeleteOfficialFollowsRequest) (*pbNews.CommonResponse, error) {
	opUser, err := imdb.GetUserByUserID(req.OfficialUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	if opUser.OfficialID == 0 {
		return &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg}, nil
	}

	usersMap, err := imdb.GetUsersMapByUserIDList(req.UserIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUsersMapByUserIDList failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	usersList := make([]db.User, 0)
	for _, user := range usersMap {
		usersList = append(usersList, user)
	}

	if err = imdb.DeleteOfficialFollows(opUser.OfficialID, req.OfficialUserID, usersList); err != nil {
		errMsg := req.OperationID + " imdb.DeleteOfficialFollows failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	if err = db.DB.DeleteOfficialFollows(opUser.OfficialID, req.OfficialUserID, req.UserIDList); err != nil {
		errMsg := req.OperationID + " db.DB.DeleteOfficialFollows failed " + err.Error() + req.OfficialUserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}
	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) ListSelfOfficialFollows(_ context.Context, req *pbNews.ListSelfOfficialFollowsRequest) (*pbNews.ListSelfOfficialFollowsResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUserByUserID failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListSelfOfficialFollowsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.ListSelfOfficialFollowsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg},
		}, nil
	}

	follows, count, err := imdb.ListOfficialFollows(user.OfficialID, req.MinFollowTime, req.MinBlockTime, req.BlockFilter, int(req.OrderBy), int(req.Offset), int(req.Limit))
	if err != nil {
		errMsg := req.OperationID + " imdb.ListOfficialFollows failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListSelfOfficialFollowsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	if len(follows) == 0 {
		return &pbNews.ListSelfOfficialFollowsResponse{
			Count: count,
		}, nil
	}

	userIDList := make([]string, len(follows))
	for i, follow := range follows {
		userIDList[i] = follow.UserID
	}

	users, err := imdb.GetUsersMapByUserIDList(userIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUsersMapByUserIDList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListSelfOfficialFollowsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	entries := make([]*pbNews.OfficialFollowEntry, len(follows))
	for i, follow := range follows {
		entries[i] = &pbNews.OfficialFollowEntry{
			UserID:     follow.UserID,
			FollowTime: follow.FollowTime,
			BlockTime:  follow.BlockTime,
		}
		userItem, exists := users[follow.UserID]
		if exists {
			entries[i].FaceURL = userItem.FaceURL
			entries[i].Nickname = userItem.Nickname
			entries[i].Gender = userItem.Gender
		}
	}

	return &pbNews.ListSelfOfficialFollowsResponse{
		Count:   count,
		Follows: entries,
	}, nil
}

func (rpc *rpcNews) ListArticleLikes(_ context.Context, req *pbNews.ListArticleLikesRequest) (*pbNews.ListArticleLikesResponse, error) {
	likes, count, err := imdb.ListArticleLikes(req.ArticleID, req.MinCreateTime, req.Keyword, int(req.Offset), int(req.Limit))
	if err != nil {
		errMsg := req.OperationID + " imdb.ListArticleLikes failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticleLikesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	entries := make([]*pbNews.ArticleLikeEntry, len(likes))
	for i, like := range likes {
		entries[i] = &pbNews.ArticleLikeEntry{
			UserID:     like.UserID,
			Nickname:   like.UserNickname,
			FaceURL:    like.UserFaceURL,
			Gender:     like.UserGender,
			CreateTime: like.CreateTime,
		}
	}

	return &pbNews.ListArticleLikesResponse{
		Count: count,
		Likes: entries,
	}, nil
}

func (rpc *rpcNews) ListArticleComments(_ context.Context, req *pbNews.ListArticleCommentsRequest) (*pbNews.ListArticleCommentsResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.ListArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg},
		}, nil
	}

	commentsMap, countsMap, err := imdb.ListArticleComments(req.ArticleID, []int64{0}, "", user.OfficialID, int(req.Offset), int(req.Limit))
	if err != nil {
		errMsg := req.OperationID + " imdb.ListArticleComments failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	comments, ok := commentsMap[0]
	if !ok {
		return &pbNews.ListArticleCommentsResponse{
			Comments: []*pbNews.ArticleCommentEntry{},
			Count:    0,
		}, err
	}

	parentCommentIdList := make([]int64, len(comments))
	for i, comment := range comments {
		parentCommentIdList[i] = comment.CommentID
	}

	repliesMap, repliesCountsMap, err := imdb.ListArticleComments(req.ArticleID, parentCommentIdList, "", user.OfficialID, 0, 0)
	if err != nil {
		errMsg := req.OperationID + " imdb.ListArticleComments failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	commentEntries := make([]*pbNews.ArticleCommentEntry, len(comments))
	for i, comment := range comments {
		if comment.UserDeleteTime != 0 {
			comment.UserFaceURL = ""
			comment.UserNickname = ""
		}
		if comment.ReplyOfficialAccountDeleteTime != 0 {
			comment.ReplyOfficialFaceURL = ""
			comment.ReplyOfficialNickname = ""
		}
		if comment.ReplyUserDeleteTime != 0 {
			comment.ReplyUserFaceURL = ""
			comment.ReplyUserNickname = ""
		}

		var commentEntry pbNews.CommentEntry
		if err = utils.CopyStructFields(&commentEntry, comment); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}

		var replyEntries []*pbNews.CommentEntry
		replies, ok := repliesMap[comment.CommentID]
		if ok {
			repliesLength := int(math.Min(float64(req.ReplyLimit), float64(len(replies))))
			replyEntries = make([]*pbNews.CommentEntry, repliesLength)
			for j := 0; j < repliesLength; j++ {
				if replies[j].UserDeleteTime != 0 {
					replies[j].UserFaceURL = ""
					replies[j].UserNickname = ""
				}
				if replies[j].ReplyOfficialAccountDeleteTime != 0 {
					replies[j].ReplyOfficialFaceURL = ""
					replies[j].ReplyOfficialNickname = ""
				}
				if replies[j].ReplyUserDeleteTime != 0 {
					replies[j].ReplyUserFaceURL = ""
					replies[j].ReplyUserNickname = ""
				}
				replyEntries[j] = &pbNews.CommentEntry{}
				if err = utils.CopyStructFields(&replyEntries[j], replies[j]); err != nil {
					log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
				}
			}
		}

		commentEntries[i] = &pbNews.ArticleCommentEntry{
			Comment: &commentEntry,
			Replies: &pbNews.ListArticleCommentRepliesResponse{
				Replies: replyEntries,
				Count:   repliesCountsMap[comment.CommentID],
			},
		}
	}

	return &pbNews.ListArticleCommentsResponse{
		Comments: commentEntries,
		Count:    countsMap[0],
	}, nil
}

func (rpc *rpcNews) ListArticleCommentReplies(_ context.Context, req *pbNews.ListArticleCommentRepliesRequest) (*pbNews.ListArticleCommentRepliesResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticleCommentRepliesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.ListArticleCommentRepliesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg},
		}, nil
	}

	commentsMap, countsMap, err := imdb.ListArticleComments(0, []int64{req.ParentCommentID}, "", user.OfficialID, int(req.Offset), int(req.Limit))
	if err != nil {
		errMsg := req.OperationID + " imdb.ListArticleComments failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListArticleCommentRepliesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	replies, ok := commentsMap[req.ParentCommentID]
	if !ok {
		return &pbNews.ListArticleCommentRepliesResponse{
			Replies: []*pbNews.CommentEntry{},
			Count:   0,
		}, err
	}

	replyEntries := make([]*pbNews.CommentEntry, len(replies))
	for i, reply := range replies {
		if reply.UserDeleteTime != 0 {
			reply.UserFaceURL = ""
			reply.UserNickname = ""
		}
		if reply.ReplyOfficialAccountDeleteTime != 0 {
			reply.ReplyOfficialFaceURL = ""
			reply.ReplyOfficialNickname = ""
		}
		if reply.ReplyUserDeleteTime != 0 {
			reply.ReplyUserFaceURL = ""
			reply.ReplyUserNickname = ""
		}
		replyEntries[i] = &pbNews.CommentEntry{}
		if err = utils.CopyStructFields(&replyEntries[i], reply); err != nil {
			log.NewDebug(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
	}

	return &pbNews.ListArticleCommentRepliesResponse{
		Replies: replyEntries,
		Count:   countsMap[req.ParentCommentID],
	}, nil
}

func (rpc *rpcNews) GetOfficialProfile(_ context.Context, req *pbNews.GetOfficialProfileRequest) (*pbNews.GetOfficialProfileResponse, error) {
	profile, err := imdb.GetOfficialProfile(req.UserID, req.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialProfile failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialProfileResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	return &pbNews.GetOfficialProfileResponse{
		Follow: &pbNews.UserFollow{
			OfficialID: profile.OfficialID,
			Nickname:   profile.Nickname,
			FaceURL:    profile.FaceURL,
			Bio:        profile.Bio,
			Type:       profile.Type,
			FollowTime: profile.FollowTime,
			Muted:      profile.Muted,
			Enabled:    profile.Enabled,
		},
	}, nil
}

func transferRecentAnalytics(records []imdb.GetTotalAnalyticsBetweenRecord) *pbNews.RecentAnalyticsEntry {
	entry := &pbNews.RecentAnalyticsEntry{
		LikesByGender:       &pbNews.RecentAnalyticsEntryGender{},
		CommentsByGender:    &pbNews.RecentAnalyticsEntryGender{},
		FollowsByGender:     &pbNews.RecentAnalyticsEntryGender{},
		ReadsByGender:       &pbNews.RecentAnalyticsEntryGender{},
		UniqueReadsByGender: &pbNews.RecentAnalyticsEntryGender{},
	}
	for _, record := range records {
		switch record.Gender {
		case 0:
			entry.LikesByGender.Unknown = record.Likes
			entry.CommentsByGender.Unknown = record.Comments
			entry.FollowsByGender.Unknown = record.Follows
			entry.ReadsByGender.Unknown = record.Reads
			entry.UniqueReadsByGender.Unknown = record.UniqueReads
		case 1:
			entry.LikesByGender.Male = record.Likes
			entry.CommentsByGender.Male = record.Comments
			entry.FollowsByGender.Male = record.Follows
			entry.ReadsByGender.Male = record.Reads
			entry.UniqueReadsByGender.Male = record.UniqueReads
		case 2:
			entry.LikesByGender.Female = record.Likes
			entry.CommentsByGender.Female = record.Comments
			entry.FollowsByGender.Female = record.Follows
			entry.ReadsByGender.Female = record.Reads
			entry.UniqueReadsByGender.Female = record.UniqueReads
		}
	}
	return entry
}

func (rpc *rpcNews) GetOfficialRecentAnalyticsByGender(_ context.Context, req *pbNews.GetOfficialRecentAnalyticsByGenderRequest) (*pbNews.GetOfficialRecentAnalyticsByGenderResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialRecentAnalyticsByGenderResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.GetOfficialRecentAnalyticsByGenderResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg},
		}, nil
	}

	currRes, err := imdb.GetTotalAnalyticsBetween(user.OfficialID, req.StartTime, req.EndTime)
	if err != nil {
		errMsg := "imdb.GetTotalAnalyticsBetween failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialRecentAnalyticsByGenderResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	prevStartTime := req.StartTime - (req.EndTime - req.StartTime)
	prevRes, err := imdb.GetTotalAnalyticsBetween(user.OfficialID, prevStartTime, req.StartTime)
	if err != nil {
		errMsg := "imdb.GetTotalAnalyticsBetween failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialRecentAnalyticsByGenderResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	return &pbNews.GetOfficialRecentAnalyticsByGenderResponse{
		Current:  transferRecentAnalytics(currRes),
		Previous: transferRecentAnalytics(prevRes),
	}, nil
}

func (rpc *rpcNews) GetOfficialAnalyticsByDay(_ context.Context, req *pbNews.GetOfficialAnalyticsByDayRequest) (*pbNews.GetOfficialAnalyticsByDayResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := "imdb.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialAnalyticsByDayResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	if user.OfficialID == 0 {
		return &pbNews.GetOfficialAnalyticsByDayResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrOfficialStatus.ErrCode, ErrMsg: constant.ErrOfficialStatus.ErrMsg},
		}, nil
	}

	records, err := imdb.GetDailyAnalytics(user.OfficialID, req.StartTime, req.EndTime)
	if err != nil {
		errMsg := "imdb.GetDailyAnalytics failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetOfficialAnalyticsByDayResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	entries := make([]*pbNews.AnalyticsByDayEntry, len(records))
	for i, record := range records {
		entries[i] = &pbNews.AnalyticsByDayEntry{
			Day:         record.Day,
			Likes:       record.Likes,
			Comments:    record.Comments,
			Follows:     record.Follows,
			Reads:       record.Reads,
			UniqueReads: record.UniqueReads,
		}
	}

	return &pbNews.GetOfficialAnalyticsByDayResponse{
		Entries: entries,
	}, nil
}

func (rpc *rpcNews) SearchOfficialAccounts(_ context.Context, req *pbNews.SearchOfficialAccountsRequest) (*pbNews.SearchOfficialAccountsResponse, error) {
	officials, count, err := imdb.SearchOfficialAccounts(req.UserID, req.Keyword, int(req.Offset), int(req.Limit))
	if err != nil {
		errMsg := req.OperationID + " imdb.SearchOfficialAccounts failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.SearchOfficialAccountsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}
	officialAccounts := make([]*pbNews.UserFollow, len(officials))
	for i, official := range officials {
		officialAccounts[i] = &pbNews.UserFollow{
			OfficialID: official.OfficialID,
			Nickname:   official.Nickname,
			FaceURL:    official.FaceURL,
			Bio:        official.Bio,
			Type:       official.Type,
			FollowTime: official.FollowTime,
			Muted:      official.Muted,
			Enabled:    official.Enabled,
		}
	}
	return &pbNews.SearchOfficialAccountsResponse{
		Count:   count,
		Entries: officialAccounts,
	}, nil
}

func (rpc *rpcNews) SearchArticles(_ context.Context, req *pbNews.SearchArticlesRequest) (*pbNews.SearchArticlesResponse, error) {
	officialFollows, err := db.DB.GetSelfOfficialAccountFollows(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetSelfOfficialAccountFollows failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.SearchArticlesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	officialFollowsMap := make(map[int64]db.OfficialFollow, len(officialFollows))
	for _, officialFollow := range officialFollows {
		officialFollowsMap[officialFollow.OfficialID] = officialFollow
	}

	articles, count, err := db.DB.SearchArticles(req.Keyword, req.OfficialID, req.MinReadTime, req.MaxReadTime, req.MinCreateTime, req.MaxCreateTime, req.Sort, req.Offset, req.Limit)
	if err != nil {
		errMsg := req.OperationID + " db.DB.SearchArticles failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.SearchArticlesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	officialIDMap := make(map[int64]bool)
	for _, article := range articles {
		officialIDMap[article.Article.OfficialID] = true
	}

	officialIDList := make([]int64, 0)
	for officialID := range officialIDMap {
		officialIDList = append(officialIDList, officialID)
	}

	officialsMap, err := imdb.GetOfficialsByOfficialIDList(officialIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialsByOfficialIDList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.SearchArticlesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	entries := make([]*pbNews.SearchArticlesEntry, len(articles))
	for i, article := range articles {
		entries[i] = &pbNews.SearchArticlesEntry{
			Article: &pbNews.UserArticleSummary{
				ArticleID:        article.Article.ArticleID,
				Title:            article.Article.Title,
				CoverPhoto:       article.Article.CoverPhoto,
				TextContent:      article.Article.TextContent,
				ReadCounts:       article.Article.ReadCounts,
				UniqueReadCounts: article.Article.UniqueReadCounts,
				LikeCounts:       article.Article.LikeCounts,
				CommentCounts:    article.Article.CommentCounts,
				RepostCounts:     article.Article.RepostCounts,
				CreateTime:       article.Article.CreateTime,
			},
			ReadTime: article.ReadTime,
		}
		if official, exist := officialsMap[article.Article.OfficialID]; exist {
			entries[i].Official = &pbNews.UserFollow{
				OfficialID: official.Id,
				Nickname:   official.Nickname,
				FaceURL:    official.FaceURL,
				Type:       int32(official.Type),
				Bio:        official.Bio,
			}

			if officialFollow, followExists := officialFollowsMap[article.Article.OfficialID]; followExists {
				entries[i].Official.FollowTime = officialFollow.FollowTime
				entries[i].Official.Muted = officialFollow.Muted
				entries[i].Official.Enabled = officialFollow.Enabled
			}
		}
	}

	return &pbNews.SearchArticlesResponse{
		Count:   count,
		Entries: entries,
	}, nil
}

func (rpc *rpcNews) GetUserArticleByArticleID(_ context.Context, req *pbNews.GetUserArticleByArticleIDRequest) (*pbNews.GetUserArticleByArticleIDResponse, error) {
	article, err := db.DB.GetArticlesByID(req.ArticleID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetArticlesByID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetUserArticleByArticleIDResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}
	if article.DeleteTime != 0 {
		errMsg := req.OperationID + "the article was deleted"
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetUserArticleByArticleIDResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	official, err := imdb.GetOfficialByOfficialID(article.OfficialID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetOfficialByOfficialID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.GetUserArticleByArticleIDResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	var followTime, likeTime, favoriteTime int64
	var favoriteID string
	var muted, enabled bool
	if req.UserID != nil {
		if follow, err := db.DB.GetOfficialFollow(official.Id, *req.UserID); err != nil {
			errMsg := req.OperationID + " db.DB.GetOfficialFollow failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.GetUserArticleByArticleIDResponse{
				CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
			}, nil
		} else if follow != nil {
			followTime = follow.FollowTime
			muted = follow.Muted
			enabled = follow.Enabled
		}

		if like, err := db.DB.GetArticleLike(article.ArticleID, *req.UserID); err != nil {
			errMsg := req.OperationID + " db.DB.GetArticleLike failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.GetUserArticleByArticleIDResponse{
				CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
			}, nil
		} else if like != nil {
			likeTime = like.CreateTime
		}

		if favorite, err := db.DB.GetArticleFavorite(article.ArticleID, *req.UserID); err != nil {
			errMsg := req.OperationID + " db.DB.GetArticleFavorite failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			return &pbNews.GetUserArticleByArticleIDResponse{
				CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
			}, nil
		} else if favorite != nil {
			favoriteTime = favorite.CreateTime
			favoriteID = favorite.FavoriteId.Hex()
		}
	}

	return &pbNews.GetUserArticleByArticleIDResponse{
		Article: &pbNews.UserArticle{
			ArticleID:        article.ArticleID,
			Title:            article.Title,
			CoverPhoto:       article.CoverPhoto,
			TextContent:      article.TextContent,
			Content:          article.Content,
			ReadCounts:       article.ReadCounts,
			UniqueReadCounts: article.UniqueReadCounts,
			CommentCounts:    article.CommentCounts,
			RepostCounts:     article.RepostCounts,
			CreateTime:       article.CreateTime,
			LikeTime:         likeTime,
			FavoriteTime:     favoriteTime,
			FavoriteID:       favoriteID,
		},
		Official: &pbNews.UserFollow{
			OfficialID: official.Id,
			Nickname:   official.Nickname,
			FaceURL:    official.FaceURL,
			Bio:        official.Bio,
			Type:       int32(official.Type),
			FollowTime: followTime,
			Muted:      muted,
			Enabled:    enabled,
		},
	}, nil
}

func (rpc *rpcNews) InsertArticleRead(_ context.Context, req *pbNews.InsertArticleReadRequest) (*pbNews.CommonResponse, error) {
	user, err := imdb.GetUserByUserID(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetUserByUserID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	article, err := db.DB.GetArticlesByID(req.ArticleID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetArticlesByID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	err = imdb.InsertArticleRead(article.ArticleID, article.OfficialID, user.UserID, user.Gender)
	if err != nil {
		errMsg := req.OperationID + " imdb.InsertArticleRead failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	err = db.DB.InsertArticleRead(article.ArticleID, req.UserID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.InsertArticleRead failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, nil
	}

	return &pbNews.CommonResponse{}, nil
}

func (rpc *rpcNews) ListUserArticleReads(_ context.Context, req *pbNews.ListUserArticleReadsRequest) (*pbNews.ListUserArticleReadsResponse, error) {
	resp := &pbNews.ListUserArticleReadsResponse{}

	follows, err := db.DB.GetSelfOfficialAccountFollows(req.UserID)
	if err != nil {
		errMsg := req.OperationID + " db.DB.GetSelfOfficialAccountFollows failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleReadsResponse{
			CommonResp: &pbNews.CommonResponse{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  errMsg,
			},
		}, nil
	}

	followsMap := make(map[int64]db.OfficialFollow, len(follows))
	for _, follow := range follows {
		followsMap[follow.OfficialID] = follow
	}

	reads, count, err := db.DB.ListUserArticleReads(req.UserID, req.MinCreateTime, req.Offset, req.Limit)
	if err != nil {
		errMsg := req.OperationID + " db.DB.ListUserArticleReads failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleReadsResponse{
			CommonResp: &pbNews.CommonResponse{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  errMsg,
			},
		}, nil
	}

	var officialIDMap = make(map[int64]bool)
	for _, read := range reads {
		officialIDMap[read.Article.OfficialID] = true
	}

	officialIDList := make([]int64, 0)
	for officialID := range officialIDMap {
		officialIDList = append(officialIDList, officialID)
	}

	officialsMap, err := imdb.GetOfficialsByOfficialIDList(officialIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialsByOfficialIDList failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleReadsResponse{
			CommonResp: &pbNews.CommonResponse{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  errMsg,
			},
		}, nil
	}

	resp.Count = count
	resp.Entries = make([]*pbNews.ListUserArticleReadsResponseEntry, len(reads))
	for i, read := range reads {
		resp.Entries[i] = &pbNews.ListUserArticleReadsResponseEntry{
			Article: &pbNews.UserArticleSummary{
				ArticleID:        read.Article.ArticleID,
				Title:            read.Article.Title,
				CoverPhoto:       read.Article.CoverPhoto,
				TextContent:      read.Article.TextContent,
				ReadCounts:       read.Article.ReadCounts,
				UniqueReadCounts: read.Article.UniqueReadCounts,
				LikeCounts:       read.Article.LikeCounts,
				CommentCounts:    read.Article.CommentCounts,
				RepostCounts:     read.Article.RepostCounts,
				CreateTime:       read.Article.CreateTime,
			},
			Official: &pbNews.UserFollow{},
			ReadTime: read.ReadTime,
		}
		if official, ok := officialsMap[read.Article.OfficialID]; ok {
			resp.Entries[i].Official.OfficialID = official.Id
			resp.Entries[i].Official.Nickname = official.Nickname
			resp.Entries[i].Official.FaceURL = official.FaceURL
			resp.Entries[i].Official.Bio = official.Bio
			resp.Entries[i].Official.Type = int32(official.Type)
		}
		if follow, ok := followsMap[read.Article.OfficialID]; ok {
			resp.Entries[i].Official.FollowTime = follow.FollowTime
			resp.Entries[i].Official.Muted = follow.Muted
			resp.Entries[i].Official.Enabled = follow.Enabled
		}
	}

	return resp, nil
}

func (rpc *rpcNews) ClearUserArticleReads(_ context.Context, req *pbNews.ClearUserArticleReadsRequest) (*pbNews.CommonResponse, error) {
	if err := imdb.ClearUserArticleReads(req.UserID); err != nil {
		errMsg := req.OperationID + " imdb.ClearUserArticleReads failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  errMsg,
		}, nil
	}

	if err := db.DB.ClearUserArticleReads(req.UserID); err != nil {
		errMsg := req.OperationID + " db.DB.ClearUserArticleReads failed " + err.Error() + req.UserID
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  errMsg,
		}, nil
	}

	return &pbNews.CommonResponse{}, nil
}
func (rpc *rpcNews) FollowedOfficialConversation(_ context.Context, req *pbNews.FollowedOfficialConversationRequest) (*pbNews.FollowedOfficialConversationResponse, error) {
	response := pbNews.FollowedOfficialConversationResponse{}
	officialAccounts, err := imdb.GetUserFollowOfficialAccountList(req.ReqUserID)
	if err != nil {
		errMsg := req.OperationID + " imdb.ClearUserArticleReads failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		//return &pbNews.CommonResponse{
		//	ErrCode: constant.ErrDB.ErrCode,
		//	ErrMsg:  errMsg,
		//}, nil
	}
	log.NewError("FollowedOfficialConversation Official Account", len(officialAccounts))
	for _, account := range officialAccounts {
		article, err := db.DB.GetLatestArticlesByOfficialAccount(account.OfficialID)
		if err != nil {
			log.NewError("FollowedOfficialConversation Official Account Redis Parser", err.Error())
		}
		if article != nil {
			bpArticle := pbNews.Article{}
			err := utils.CopyStructFields(&bpArticle, article)
			if err == nil {
				//if len(bpArticle.TextContent) > 100 {
				//	bpArticle.TextContent = bpArticle.TextContent[0:99]
				//}
				bpArticle.TextContent = ""
				if bpArticle.OfficialName == "" || bpArticle.OfficialProfileImg == "" {
					officialAccount, err := imdb.GetOfficialByOfficialID(bpArticle.OfficialID)
					if err == nil {
						bpArticle.OfficialName = officialAccount.Nickname
						bpArticle.OfficialProfileImg = officialAccount.FaceURL
					}

				}
				response.Articles = append(response.Articles, &bpArticle)
			} else {
				log.NewError("FollowedOfficialConversation Official Account", err.Error())
			}

		}

	}

	return &response, nil
}
func (rpc *rpcNews) GetOfficialIDNumberAvailability(_ context.Context, req *pbNews.GetOfficialIDNumberAvailabilityRequest) (*pbNews.GetOfficialIDNumberAvailabilityResponse, error) {
	response := pbNews.GetOfficialIDNumberAvailabilityResponse{}
	IsAvailable := imdb.CheckOfficialIDNumberAvailable(req.IDNumber, int8(req.IDType), 0)
	response.IsAvailable = !IsAvailable
	response.IDNumber = req.IDNumber
	return &response, nil
}

func (rpc *rpcNews) ListUserArticleComments(_ context.Context, req *pbNews.ListUserArticleCommentsRequest) (*pbNews.ListUserArticleCommentsResponse, error) {
	comments, count, err := db.DB.ListUserArticleComments(req.UserID, req.ArticleID, req.Offset, req.Limit)
	if err != nil {
		errMsg := req.OperationID + " db.DB.ListUserArticleComments failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	userIDMap := make(map[string]bool)
	officialIDMap := make(map[int64]bool)
	pbComments := make([]*pbNews.UserArticleCommentEntry, len(comments))
	for i, comment := range comments {
		pbComments[i] = &pbNews.UserArticleCommentEntry{
			Comment: &pbNews.CommentEntry{},
		}
		if err = utils.CopyStructFields(&pbComments[i].Comment, comment); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
		if comment.Like != nil {
			pbComments[i].Comment.LikeTime = comment.Like.CreateTime
		}
		if comment.UserID != "" {
			userIDMap[comment.UserID] = true
		}
		if comment.ReplyUserID != "" {
			userIDMap[comment.ReplyUserID] = true
		}
		if comment.OfficialID != 0 {
			officialIDMap[comment.OfficialID] = true
		}
		if comment.ReplyOfficialID != 0 {
			officialIDMap[comment.ReplyOfficialID] = true
		}
		if comment.TopReply != nil {
			pbComments[i].TopReply = &pbNews.CommentEntry{}
			if err = utils.CopyStructFields(&pbComments[i].TopReply, comment.TopReply); err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
			}
			if comment.TopReply.Like != nil {
				pbComments[i].TopReply.LikeTime = comment.TopReply.Like.CreateTime
			}
			if comment.TopReply.UserID != "" {
				userIDMap[comment.TopReply.UserID] = true
			}
			if comment.TopReply.ReplyUserID != "" {
				userIDMap[comment.TopReply.ReplyUserID] = true
			}
			if comment.TopReply.OfficialID != 0 {
				officialIDMap[comment.TopReply.OfficialID] = true
			}
			if comment.TopReply.ReplyOfficialID != 0 {
				officialIDMap[comment.TopReply.ReplyOfficialID] = true
			}
		}
	}

	var userIDList []string
	for userID := range userIDMap {
		userIDList = append(userIDList, userID)
	}

	usersMap, err := imdb.GetUsersMapByUserIDList(userIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUsersMapByUserIDList failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	var officialIDList []int64
	for officialID := range officialIDMap {
		officialIDList = append(officialIDList, officialID)
	}

	officialsMap, err := imdb.GetOfficialsByOfficialIDList(officialIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialsByOfficialIDList failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleCommentsResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	for i := range pbComments {
		if user, ok := usersMap[pbComments[i].Comment.UserID]; ok {
			pbComments[i].Comment.UserNickname = user.Nickname
			pbComments[i].Comment.UserFaceURL = user.FaceURL
		}
		if official, ok := officialsMap[pbComments[i].Comment.OfficialID]; ok {
			pbComments[i].Comment.OfficialNickname = official.Nickname
			pbComments[i].Comment.OfficialFaceURL = official.FaceURL
		}
		if pbComments[i].TopReply != nil {
			if user, ok := usersMap[pbComments[i].TopReply.UserID]; ok {
				pbComments[i].TopReply.UserNickname = user.Nickname
				pbComments[i].TopReply.UserFaceURL = user.FaceURL
			}
			if official, ok := officialsMap[pbComments[i].TopReply.OfficialID]; ok {
				pbComments[i].TopReply.OfficialNickname = official.Nickname
				pbComments[i].TopReply.OfficialFaceURL = official.FaceURL
			}
		}
	}

	return &pbNews.ListUserArticleCommentsResponse{
		Comments: pbComments,
		Count:    count,
	}, nil
}

func (rpc *rpcNews) ListUserArticleCommentReplies(_ context.Context, req *pbNews.ListUserArticleCommentRepliesRequest) (*pbNews.ListUserArticleCommentRepliesResponse, error) {
	comments, count, err := db.DB.ListUserArticleCommentReplies(req.UserID, req.CommentID, req.Offset, req.Limit)
	if err != nil {
		errMsg := req.OperationID + " db.DB.ListUserArticleCommentReplies failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleCommentRepliesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	userIDMap := make(map[string]bool)
	officialIDMap := make(map[int64]bool)
	pbComments := make([]*pbNews.CommentEntry, len(comments))
	for i, comment := range comments {
		pbComments[i] = &pbNews.CommentEntry{}
		if err = utils.CopyStructFields(&pbComments[i], comment); err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", err.Error())
		}
		if comment.Like != nil {
			pbComments[i].LikeTime = comment.Like.CreateTime
		}
		if comment.UserID != "" {
			userIDMap[comment.UserID] = true
		}
		if comment.ReplyUserID != "" {
			userIDMap[comment.ReplyUserID] = true
		}
		if comment.OfficialID != 0 {
			officialIDMap[comment.OfficialID] = true
		}
		if comment.ReplyOfficialID != 0 {
			officialIDMap[comment.ReplyOfficialID] = true
		}
	}

	var userIDList []string
	for userID := range userIDMap {
		userIDList = append(userIDList, userID)
	}

	usersMap, err := imdb.GetUsersMapByUserIDList(userIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetUsersMapByUserIDList failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleCommentRepliesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	var officialIDList []int64
	for officialID := range officialIDMap {
		officialIDList = append(officialIDList, officialID)
	}

	officialsMap, err := imdb.GetOfficialsByOfficialIDList(officialIDList)
	if err != nil {
		errMsg := req.OperationID + " imdb.GetOfficialsByOfficialIDList failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.ListUserArticleCommentRepliesResponse{
			CommonResp: &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg},
		}, nil
	}

	for i := range pbComments {
		if user, ok := usersMap[pbComments[i].UserID]; ok {
			pbComments[i].UserNickname = user.Nickname
			pbComments[i].UserFaceURL = user.FaceURL
		}
		if official, ok := officialsMap[pbComments[i].OfficialID]; ok {
			pbComments[i].OfficialNickname = official.Nickname
			pbComments[i].OfficialFaceURL = official.FaceURL
		}
	}

	return &pbNews.ListUserArticleCommentRepliesResponse{
		Comments: pbComments,
		Count:    count,
	}, nil
}

func (rpc *rpcNews) GetOfficialFollowers(_ context.Context, req *pbNews.GetOfficialFollowersReq) (*pbNews.GetOfficialFollowersResp, error) {
	resp := &pbNews.GetOfficialFollowersResp{}

	where := map[string]string{}
	where["user"] = req.User
	where["official_account"] = req.OfficialAccount
	where["start_time"] = req.StartTime

	where["end_time"] = req.EndTime
	where["muted"] = fmt.Sprintf("%d", req.Muted)

	official, count, err := imdb.GetOfficialFollowerByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}
	err = utils.CopyStructFields(&resp.OfficialFollowers, official)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Copy Struct Fields failed", err.Error())
		return nil, err
	}
	log.Debug("", "resp.Likes len: ", len(resp.OfficialFollowers))

	// Get total interests count
	resp.OfficialFollowersCount = count
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())

	return resp, nil
}

func (rpc *rpcNews) BlockFollower(_ context.Context, req *pbNews.BlockFollowerReq) (*pbNews.BlockFollowerResp, error) {
	resp := &pbNews.BlockFollowerResp{CommResp: &pbNews.CommonResponse{}}

	if req.Block == 1 {
		if err := imdb.AddOfficialFollowBlocks(req.OfficialID, req.OpUserID, []string{req.UserID}); err != nil {
			errMsg := req.OperationID + " imdb.AddOfficialFollowBlock failed " + err.Error() + req.OpUserID
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}
		if err := db.DB.AddOfficialFollowBlock(req.OfficialID, req.OpUserID, []string{req.UserID}); err != nil {
			errMsg := req.OperationID + " db.DB.AddOfficialFollowBlock failed " + err.Error() + req.OpUserID
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}
	} else if req.Block == 2 {
		if err := imdb.DeleteOfficialFollowBlocks(req.OfficialID, []string{req.UserID}); err != nil {
			errMsg := req.OperationID + " imdb.DeleteOfficialFollowBlock failed " + err.Error() + req.UserID
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}
		if err := db.DB.DeleteOfficialFollowBlock(req.OfficialID, []string{req.UserID}); err != nil {
			errMsg := req.OperationID + " db.DB.DeleteOfficialFollowBlock failed " + err.Error() + req.UserID
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}
	}

	return resp, nil

}

func (rpc *rpcNews) MuteFollower(_ context.Context, req *pbNews.MuteFollowerReq) (*pbNews.MuteFollowerResp, error) {
	resp := &pbNews.MuteFollowerResp{}

	err := imdb.MuteFollower(req.OpUserID, req.UserID, req.OfficialID, req.Mute)
	if err != nil {
		log.NewError("", "db error: ", err.Error())
		return nil, err
	}

	return resp, nil
}

func (rpc *rpcNews) RemoveFollowers(_ context.Context, req *pbNews.RemoveFollowersReq) (*pbNews.RemoveFollowersResp, error) {
	resp := &pbNews.RemoveFollowersResp{CommResp: &pbNews.CommonResponse{}}

	for _, user := range req.Users {
		opUser, err := imdb.GetUserByUserID(req.OpUserID)
		if err != nil {
			errMsg := "imdb.GetUserByUserID failed " + err.Error()
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}

		if err = imdb.DeleteOfficialFollows(user.OfficialID, req.OpUserID, []db.User{*opUser}); err != nil {
			errMsg := req.OperationID + " imdb.DeleteOfficialFollows failed " + err.Error() + req.OpUserID
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}
		if err = db.DB.DeleteOfficialFollows(user.OfficialID, req.OpUserID, []string{req.OpUserID}); err != nil {
			errMsg := req.OperationID + " db.DB.DeleteOfficialFollows failed " + err.Error() + req.OpUserID
			log.NewError(req.OperationID, errMsg)
			resp.CommResp.ErrCode = constant.ErrDB.ErrCode
			resp.CommResp.ErrMsg = errMsg
			return resp, nil
		}

		chat.OfficialAccountFollowUnfollowNotification(req.OperationID, req.OpUserID, strconv.FormatInt(user.OfficialID, 10))

	}

	return resp, nil
}

func removeArticleFiles(articles []*db.Article) error {
	var obs []cos.Object
	medias := getArticleMediaUrls(articles)
	for _, url := range medias {
		objName := admin_cms.GetObjNameByURL(url, true)
		obs = append(obs, cos.Object{Key: objName})
	}

	if len(obs) == 0 {
		return nil
	}
	deleteOpt := &cos.ObjectDeleteMultiOptions{Objects: obs}
	client, err := admin_cms.GetTencentCloudClient(true)
	if err != nil {
		log.Error("", "Tencent cloud get failed")
		return err
	}

	_, _, err = client.Object.DeleteMulti(context.Background(), deleteOpt)
	if err != nil {
		client, err = admin_cms.GetTencentCloudClient(false)
		if err != nil {
			log.Error("", "Tencent cloud get failed")
			return err
		}
		_, _, err = client.Object.DeleteMulti(context.Background(), deleteOpt)
		if err != nil {
			log.Error("", "Delete file from tencent cloud failed.")
			return err
		}
		return err
	}

	return nil
}

func getArticleMediaUrls(articles []*db.Article) []string {

	regeTool := regexp.MustCompile(fmt.Sprintf(`https:\/\/%s\.cos\..*?\.myqcloud\.com\/[^"]*`, config.Config.Credential.Tencent.PersistenceBucket))
	var medias []string

	for _, article := range articles {
		urls := regeTool.FindAllString(article.Content, -1)
		medias = append(medias, urls...)
	}

	log.Debug("medias: ")
	return medias
}

func getArticleMediaUrlMap(articles []*db.Article) map[string]struct{} {

	regeTool := regexp.MustCompile(fmt.Sprintf(`https:\/\/%s\.cos\..*?\.myqcloud\.com\/[^"]*`, config.Config.Credential.Tencent.PersistenceBucket))
	var medias = make(map[string]struct{})

	for _, article := range articles {
		urls := regeTool.FindAllString(article.Content, -1)
		for _, url := range urls {
			medias[url] = struct{}{}
		}
	}

	return medias
}

func UpdateArticleCount(updateMongoDB, updateMySQL bool, articleID, likeCount, commentCount, readCount, repostCount, uniqueReadCounts int64) {
	articleMongo := db.Article{ArticleID: articleID, LikeCounts: likeCount, CommentCounts: commentCount,
		RepostCounts: repostCount, ReadCounts: readCount, UniqueReadCounts: uniqueReadCounts}
	article := db.ArticleSQL{ArticleID: articleID, LikeCounts: likeCount, CommentCounts: commentCount,
		RepostCounts: repostCount, ReadCounts: readCount, UniqueReadCounts: uniqueReadCounts}
	if updateMongoDB {
		if err := db.DB.UpdateArticleV2(articleMongo); err != nil {
			errMsg := "Update Article count on mongodb failed " + err.Error()
			log.NewError(errMsg)
		}
	}
	if updateMySQL {
		if _, err := imdb.UpdateArticleV2(&article); err != nil {
			errMsg := "Update Article count on mysql failed " + err.Error()
			log.NewError(errMsg)
		}
	}
}

func UpdateArticleCommentCount(updateMongoDB, updateMySQL bool, commentID, replyCounts, likeCounts int64) {
	articleComment := db.ArticleComment{CommentID: commentID, ReplyCounts: replyCounts, LikeCounts: likeCounts}
	commentSQL := db.ArticleCommentSQL{CommentID: commentID, ReplyCounts: replyCounts, LikeCounts: likeCounts}
	if updateMongoDB {
		if err := db.DB.UpdateArticleCommentV2(articleComment); err != nil {
			errMsg := "Update Article count on mongodb failed " + err.Error()
			log.NewError(errMsg)
		}
	}
	if updateMySQL {
		if _, err := imdb.UpdateArticleCommentV2(&commentSQL); err != nil {
			errMsg := "Update Article count on mysql failed " + err.Error()
			log.NewError(errMsg)
		}
	}
}

func (rpc *rpcNews) DeleteArticleComment(_ context.Context, req *pbNews.DeleteArticleCommentRequest) (*pbNews.CommonResponse, error) {
	articleComment, err := imdb.GetCommentByCommentID(req.CommentID)
	if err != nil {
		errMsg := "imdb.GetArticleByArticleID failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: errMsg}, err
	}
	if articleComment.UserID != req.ReqUserID {
		return &pbNews.CommonResponse{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}, nil
	}
	articleComment.DeletedBy = req.ReqUserID
	articleComment.DeleteTime = time.Now().Unix()
	err = imdb.DeleteArticleComment(articleComment)
	if err != nil {
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}, err
	}

	err = db.DB.DeleteArticleCommentV2(articleComment.CommentID, articleComment.ArticleID, articleComment.ParentCommentID, articleComment.DeleteTime, req.ReqUserID)
	if err != nil {
		errMsg := "db.DB.AddArticleComment failed " + err.Error()
		log.NewError(req.OperationID, errMsg)
		return &pbNews.CommonResponse{ErrCode: constant.ErrDB.ErrCode, ErrMsg: err.Error()}, err
	}
	go func() {
		if articleComment.ParentCommentID == 0 {
			// no parent comment
			UpdateArticleCount(true, false, articleComment.ArticleID, 0, -1, 0, 0, 0)
		}
	}()

	return &pbNews.CommonResponse{}, nil
}

func addSystemOfficialAccount(official int64, allCount int64, users []db.User, operationID string) error {

	var userIdList = make([]string, len(users))
	var genderList = make([]int32, len(users))
	for index, user := range users {
		userIdList[index] = user.UserID
		genderList[index] = user.Gender
	}

	if err := imdb.AddOfficialFollows(official, allCount, userIdList, genderList); err != nil {
		errMsg := " imdb.AddOfficialFollows failed " + err.Error()
		log.NewError(errMsg)
		return err
	}
	if err := db.DB.AddOfficialFollows(official, userIdList); err != nil {
		errMsg := " db.DB.AddOfficialFollow failed " + err.Error()
		log.NewError(errMsg)
		return err
	}

	for _, user := range users {
		chat.OfficialAccountFollowUnfollowNotification(operationID, user.UserID, strconv.FormatInt(official, 10))
	}

	log.Debug("", "xxxx end time: ", time.Now().Unix())

	return nil
}
