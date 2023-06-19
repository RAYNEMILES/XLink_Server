package admin_cms

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	utils3 "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbAdminCMS "Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/proto/appversion"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gotp "github.com/diebietse/gotp/v2"
	"google.golang.org/grpc"
)

type adminCMSServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (s *adminCMSServer) GetInviteCodeList(ctx context.Context, request *pbAdminCMS.GetInviteCodeListRequest) (*pbAdminCMS.GetInviteCodeListResponse, error) {
	response := &pbAdminCMS.GetInviteCodeListResponse{}
	response.CurrentNumber = int64(request.Pagination.PageNumber)
	response.ShowNumber = int64(request.Pagination.ShowNumber)

	inviteCodeLimit, _ := imdb.GetConfigByNameByDefault(constant.ConfigInviteCodeIsLimitKey, strconv.Itoa(constant.ConfigInviteCodeIsLimitFalse))
	limit, _ := strconv.Atoi(inviteCodeLimit.Value)
	response.Limit = int64(limit)

	inviteCodeIsOpen, _ := imdb.GetConfigByNameByDefault(constant.ConfigInviteCodeIsOpenKey, strconv.Itoa(constant.ConfigInviteCodeIsOpenTrue))
	isOpen, _ := strconv.Atoi(inviteCodeIsOpen.Value)
	response.IsOpen = int64(isOpen)

	baseLink, _ := imdb.GetConfigByNameByDefault(constant.ConfigInviteCodeBaseLinkKey, "")
	response.InviteCodeBaseLink = baseLink.Value

	where := map[string]string{
		"code":      request.Conditions.Code,
		"state":     request.Conditions.State,
		"note":      request.Conditions.Note,
		"user_id":   request.Conditions.UserId,
		"user_name": request.Conditions.UserName,
	}

	response.Total, _ = imdb.GetCodeNums(where)
	list, _ := imdb.GetCodeList(where, int(request.Pagination.ShowNumber), int(request.Pagination.PageNumber), request.Conditions.OrderBy)

	usersInfo := make(map[string]string)
	if len(list) > 0 {
		column, _ := utils3.ArrayColumn(list, "UserId")
		userIdArr := utils3.ArrayKey(column)
		usersInfo = imdb.GetSomeUserNameByUserId(userIdArr)
	}

	var InviteCodes []*pbAdminCMS.InviteCodes
	for _, code := range list {
		InviteCodes = append(InviteCodes, &pbAdminCMS.InviteCodes{
			Id:       strconv.Itoa(int(code.ID)),
			Code:     code.Code,
			UserId:   code.UserId,
			Note:     code.Note,
			Greeting: code.Greeting,
			UserName: usersInfo[code.UserId],
			State:    strconv.Itoa(code.State),
		})
	}
	response.List = InviteCodes

	log.NewError("GetInviteCodeList", utils.GetSelfFuncName(), request.String(), "response:", response.String())

	return response, nil
}

func (s *adminCMSServer) MultiDeleteInviteCode(ctx context.Context, request *pbAdminCMS.MultiDeleteInviteCodeRequest) (*pbAdminCMS.MultiDeleteInviteCodeResponse, error) {
	var response = &pbAdminCMS.MultiDeleteInviteCodeResponse{}

	result := imdb.InviteCodeMultiSet(request.Code, strconv.Itoa(constant.InviteCodeStateDelete))

	if result == false {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", result)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) SetInviteCodeLimit(ctx context.Context, request *pbAdminCMS.SetInviteCodeLimitRequest) (*pbAdminCMS.SetInviteCodeLimitResponse, error) {
	var response = &pbAdminCMS.SetInviteCodeLimitResponse{}

	result := imdb.SetConfigByName(constant.ConfigInviteCodeIsLimitKey, request.State)

	if result == false {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", result)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) SetInviteCodeSwitch(ctx context.Context, request *pbAdminCMS.SetInviteCodeSwitchRequest) (*pbAdminCMS.SetInviteCodeSwitchResponse, error) {
	var response = &pbAdminCMS.SetInviteCodeSwitchResponse{}

	result := imdb.SetConfigByName(constant.ConfigInviteCodeIsOpenKey, request.State)

	if result == false {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", result)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) SetChannelCodeSwitch(ctx context.Context, request *pbAdminCMS.SetChannelCodeSwitchRequest) (*pbAdminCMS.SetChannelCodeSwitchResponse, error) {
	var response = &pbAdminCMS.SetChannelCodeSwitchResponse{}

	result := imdb.SetConfigByName(constant.ConfigChannelCodeIsOpenKey, request.State)

	if result == false {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", result)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) SetChannelCodeLimit(ctx context.Context, request *pbAdminCMS.SetChannelCodeLimitRequest) (*pbAdminCMS.SetChannelCodeLimitResponse, error) {
	var response = &pbAdminCMS.SetChannelCodeLimitResponse{}

	result := imdb.SetConfigByName(constant.ConfigChannelCodeIsLimitKey, request.State)

	if result == false {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", result)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) MultiDeleteChannelCode(ctx context.Context, request *pbAdminCMS.MultiDeleteChannelCodeRequest) (*pbAdminCMS.MultiDeleteChannelCodeResponse, error) {
	var response = &pbAdminCMS.MultiDeleteChannelCodeResponse{}

	result := imdb.ChannelCodeMultiSet(request.Code, strconv.Itoa(constant.InviteChannelCodeStateDelete))

	if result == false {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", result)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) AddChannelCode(ctx context.Context, request *pbAdminCMS.AddChannelCodeRequest) (*pbAdminCMS.AddChannelCodeResponse, error) {
	var response = &pbAdminCMS.AddChannelCodeResponse{}

	// code is valid
	isExist := imdb.ChannelCodeIsExpired(request.Code)
	if isExist == true {
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrChannelCodeInexistence.ErrCode,
			ErrMsg:  constant.ErrChannelCodeInexistence.ErrMsg,
		}
		return response, nil
	}

	if imdb.CodeIsExpired(request.Code) {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "code is exist!", request.Code)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrChannelCodeInexistence.ErrCode,
			ErrMsg:  constant.ErrChannelCodeInexistence.ErrMsg,
		}
		return response, nil
	}

	var comment = &db.InviteChannelCode{}
	comment.Code = request.Code
	comment.OperatorUserId = request.OperatorUserId
	comment.FriendId = request.FriendIds
	comment.GroupId = request.GroupIds
	comment.State = constant.InviteChannelCodeStateValid
	comment.Note = request.Note
	comment.Greeting = request.Greeting
	comment.SourceId = request.SourceId

	dbErr := imdb.AddInviteChannelCode(comment)
	if dbErr != nil {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", dbErr)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) EditChannelCode(ctx context.Context, request *pbAdminCMS.EditChannelCodeRequest) (*pbAdminCMS.EditChannelCodeResponse, error) {
	var response = &pbAdminCMS.EditChannelCodeResponse{}

	// code is valid
	isExist := imdb.ChannelCodeIsExpired(request.Code)
	if isExist == false {
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrChannelCodeInexistence.ErrCode,
			ErrMsg:  constant.ErrChannelCodeInexistence.ErrMsg,
		}
		return response, nil
	}

	// only one official channel code
	if request.SourceId == constant.UserRegisterSourceTypeOfficial {
		officialCode := imdb.GetOfficialChannelCode()
		if officialCode != nil && officialCode.Code != request.Code {
			response.CommonResp = &appversion.CommonResp{
				ErrCode: constant.OnlyOneOfficialChannelCode.ErrCode,
				ErrMsg:  constant.OnlyOneOfficialChannelCode.ErrMsg,
			}
			return response, nil
		}
	}

	var comment = db.InviteChannelCode{}

	comment.Code = request.Code
	comment.OperatorUserId = request.OperatorUserId
	comment.FriendId = request.FriendIds
	comment.GroupId = request.GroupIds
	comment.Note = request.Note
	comment.Greeting = request.Greeting
	comment.SourceId = request.SourceId

	dbErr := imdb.EditChannelCodeByCode(comment.Code, comment)
	if dbErr != true {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", dbErr)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) SwitchChannelCodeState(ctx context.Context, request *pbAdminCMS.SwitchChannelCodeStateRequest) (*pbAdminCMS.SwitchChannelCodeStateResponse, error) {
	var response = &pbAdminCMS.SwitchChannelCodeStateResponse{}

	// code is valid
	isExist := imdb.ChannelCodeIsExpired(request.Code)
	if isExist == false {
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrChannelCodeInexistence.ErrCode,
			ErrMsg:  constant.ErrChannelCodeInexistence.ErrMsg,
		}
		return response, nil
	}

	var comment = &db.InviteChannelCode{}
	comment.Code = request.Code
	comment.OperatorUserId = request.OperatorUserId
	comment.State = utils.StringToInt(request.State)
	dbErr := imdb.SwitchInviteChannelCodeState(comment)
	if dbErr != nil {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), request.String(), "dbErr:", dbErr)
		response.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}

		return response, nil
	}

	response.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return response, nil
}

func (s *adminCMSServer) GetChannelCodeList(ctx context.Context, request *pbAdminCMS.GetChannelCodeListRequest) (*pbAdminCMS.GetChannelCodeListResponse, error) {
	response := &pbAdminCMS.GetChannelCodeListResponse{}
	response.CurrentNumber = int64(request.Pagination.PageNumber)
	response.ShowNumber = int64(request.Pagination.ShowNumber)

	channelCodeLimit, _ := imdb.GetConfigByNameByDefault(constant.ConfigChannelCodeIsLimitKey, strconv.Itoa(constant.ConfigChannelCodeIsLimitTrue))
	limit, _ := strconv.Atoi(channelCodeLimit.Value)
	response.Limit = int64(limit)

	channelCodeIsOpen, _ := imdb.GetConfigByNameByDefault(constant.ConfigChannelCodeIsOpenKey, strconv.Itoa(constant.ConfigChannelCodeIsOpenTrue))
	isOpen, _ := strconv.Atoi(channelCodeIsOpen.Value)
	response.IsOpen = int64(isOpen)

	where := map[string]string{
		"code":      request.Conditions.Code,
		"friend_id": request.Conditions.FriendId,
		"group_id":  request.Conditions.GroupId,
		"state":     request.Conditions.State,
		"note":      request.Conditions.Note,
	}
	if request.Conditions.IsOfficial == "1" {
		where["source_id"] = "1"
	}

	list, _ := imdb.GetInviteChannelCodeList(where, int(request.Pagination.PageNumber), int(request.Pagination.ShowNumber), request.Conditions.OrderBy)
	var ChannelCodes []*pbAdminCMS.ChannelCode
	userIDMap := map[string]bool{}
	groupIDMap := map[string]bool{}
	var userIDList, gIDList []string
	for _, code := range list {
		friendIDList := strings.Split(code.FriendId, ",")
		groupIDList := strings.Split(code.GroupId, ",")
		for _, v := range friendIDList {
			_, ok := userIDMap[v]
			if !ok {
				userIDMap[v] = true
				userIDList = append(userIDList, v)
			}
		}
		for _, v := range groupIDList {
			_, ok := groupIDMap[v]
			if !ok {
				groupIDMap[v] = true
				gIDList = append(gIDList, v)
			}
		}
	}
	userDataMap, err := imdb.GetUserByUserIDList(userIDList)
	if err != nil {
		return response, nil
	}
	groupDataMap, err := imdb.GetGroupByIDList(gIDList)
	if err != nil {
		return response, nil
	}
	for _, code := range list {
		friendList := strings.Split(code.FriendId, ",")
		groupList := strings.Split(code.GroupId, ",")
		var friendObjList []*pbAdminCMS.Friends
		var groupObjList []*pbAdminCMS.Groups

		for _, v := range friendList {
			if _, ok := userDataMap[v]; ok {
				friendObjList = append(friendObjList, &pbAdminCMS.Friends{
					Id:   v,
					Name: userDataMap[v],
				})
			}
		}
		for _, v := range groupList {
			if _, ok := groupDataMap[v]; ok {
				groupObjList = append(groupObjList, &pbAdminCMS.Groups{
					Id:   v,
					Name: groupDataMap[v],
				})
			}
		}
		var GroupIds []string
		var FriendIds []string
		// delete invalid group id
		GroupIds = imdb.GetValidGroupIdListByGroupIdList(strings.Split(code.GroupId, ","))
		// delete invalid friend id
		FriendIds, _ = imdb.GetValidUserIdListByUserIdList(strings.Split(code.FriendId, ","))

		ChannelCodes = append(ChannelCodes, &pbAdminCMS.ChannelCode{
			Id:        strconv.Itoa(int(code.ID)),
			Code:      code.Code,
			GroupIds:  GroupIds,
			FriendIds: FriendIds,
			Note:      code.Note,
			Greeting:  code.Greeting,
			State:     strconv.Itoa(code.State),
			SourceId:  code.SourceId,
			Friends:   friendObjList,
			Groups:    groupObjList,
		})
	}
	response.ChannelCodes = ChannelCodes

	response.Total, _ = imdb.GetTotalInviteChannelCodeCount(where)

	return response, nil
}

func (s *adminCMSServer) CheckInviteCode(ctx context.Context, request *pbAdminCMS.CheckInviteCodeRequest) (*pbAdminCMS.CheckInviteCodeResponse, error) {
	code := request.InviteCode
	codeInfo := imdb.GetCodeInfoByCode(code)

	response := &pbAdminCMS.CheckInviteCodeResponse{}

	response.Valid = false
	if codeInfo != nil && codeInfo.State == constant.InviteCodeStateValid {
		response.Valid = true
	}
	return response, nil
}

func (s *adminCMSServer) AddInviteCode(_ context.Context, request *pbAdminCMS.AddInviteCodeRequest) (*pbAdminCMS.AddInviteCodeResponse, error) {
	log.NewInfo(request.OperationID, utils.GetSelfFuncName(), "req:", request.String())

	inviteCode := imdb.GetCodeInfoByCode(request.Code)
	if inviteCode != nil {
		if inviteCode.State == constant.InviteCodeStateDelete {
			return &pbAdminCMS.AddInviteCodeResponse{
				CommonResp: &appversion.CommonResp{
					ErrCode: constant.ErrAddInviteCodeIsDelete.ErrCode,
					ErrMsg:  constant.ErrAddInviteCodeIsDelete.ErrMsg,
				},
			}, nil
		}
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "code is exist!", request.Code)
		return &pbAdminCMS.AddInviteCodeResponse{
			CommonResp: &appversion.CommonResp{
				ErrCode: constant.ErrAddInviteCodeIsExist.ErrCode,
				ErrMsg:  constant.ErrAddInviteCodeIsExist.ErrMsg,
			},
		}, nil
	}

	if imdb.ChannelCodeIsExpired(request.Code) {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "code is exist!", request.Code)
		return &pbAdminCMS.AddInviteCodeResponse{
			CommonResp: &appversion.CommonResp{
				ErrCode: constant.ErrAddInviteCodeIsExist.ErrCode,
				ErrMsg:  constant.ErrAddInviteCodeIsExist.ErrMsg,
			},
		}, nil
	}

	// user is existed
	id, _ := imdb.GetCodeByUserID(request.UserId)
	if id.Code != "" {
		log.NewError(request.OperationID, utils.GetSelfFuncName(), "user is existed!", request.UserId, id.Code, id)
		return &pbAdminCMS.AddInviteCodeResponse{
			CommonResp: &appversion.CommonResp{
				ErrCode: constant.ErrAddInviteCodeUserIsExist.ErrCode,
				ErrMsg:  constant.ErrAddInviteCodeUserIsExist.ErrMsg,
			},
		}, nil
	}

	// get user info id
	userInfo, _ := imdb.GetUserByUserID(request.UserId)
	if userInfo == nil {
		return &pbAdminCMS.AddInviteCodeResponse{
			CommonResp: &appversion.CommonResp{
				ErrCode: constant.ErrAddInviteCodeUserNotExist.ErrCode,
				ErrMsg:  constant.ErrAddInviteCodeUserNotExist.ErrMsg,
			},
		}, nil
	}

	res, _ := imdb.GetCodeByUserID(request.UserId)
	if *res != (db.InviteCode{}) {
		return &pbAdminCMS.AddInviteCodeResponse{
			CommonResp: &appversion.CommonResp{
				ErrCode: constant.ErrAddInviteCodeUserHasExist.ErrCode,
				ErrMsg:  constant.ErrAddInviteCodeUserHasExist.ErrMsg,
			},
		}, nil
	}

	resp := &pbAdminCMS.AddInviteCodeResponse{}
	resp.InviteCode = request.Code
	// auto generate invite code
	if resp.InviteCode == "" {
		resp.InviteCode = utils.GenerateInviteCode(int64(userInfo.ID))
	}

	result := imdb.AddCode(request.UserId, resp.InviteCode, request.Greeting, request.Note)
	if result == false {
		return &pbAdminCMS.AddInviteCodeResponse{
			CommonResp: &appversion.CommonResp{
				ErrCode: constant.ErrDB.ErrCode,
				ErrMsg:  constant.ErrDB.ErrMsg,
			},
		}, nil
	}

	resp.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	return resp, nil
}

func (s *adminCMSServer) GetInviteCodeBaseLink(_ context.Context, req *pbAdminCMS.GetInviteCodeBaseLinkReq) (*pbAdminCMS.GetInviteCodeBaseLinkResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.GetInviteCodeBaseLinkResp{}
	configByName, _ := imdb.GetConfigByName("invite_code_base_link")

	resp.InviteCodeBaseLink = configByName.Value

	return resp, nil
}

func (s *adminCMSServer) SetInviteCodeBaseLink(_ context.Context, req *pbAdminCMS.SetInviteCodeBaseLinkReq) (*pbAdminCMS.GetInviteCodeBaseLinkResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.GetInviteCodeBaseLinkResp{}

	imdb.SetConfigByName("invite_code_base_link", req.Value)
	return resp, nil
}

func (s *adminCMSServer) mustEmbedUnimplementedAdminCMSServer() {
	// TODO implement me
	panic("implement me")
}

func NewAdminCMSServer(port int) *adminCMSServer {
	return &adminCMSServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImAdminCMSName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *adminCMSServer) Run() {
	log.NewPrivateLog(constant.OpenImAdminCmsLog)
	log.NewInfo("0", "AdminCMS rpc start ")
	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)

	// listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("0", "listen network success, ", address, listener)
	defer listener.Close()
	// grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	// Service registers with etcd
	pbAdminCMS.RegisterAdminCMSServer(srv, s)
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
	log.NewInfo("0", "message cms rpc success")
}

// admin login api
func (s *adminCMSServer) AdminLogin(_ context.Context, req *pbAdminCMS.AdminLoginReq) (*pbAdminCMS.AdminLoginResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.AdminLoginResp{}

	user, err := imdb.GetUserByUserID(req.AdminID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Get admin By UserID failed!", "adminID: ", req.AdminID, err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	newPasswordFirst := req.Secret + user.Salt
	passwordData := []byte(newPasswordFirst)
	has := md5.Sum(passwordData)
	password := fmt.Sprintf("%x", has)

	if password == user.Password {
		token, expTime, err := token_verify.CreateToken(req.AdminID, constant.AdminPlatformID, req.GAuthTypeToken, 0)
		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "generate token success", "token: ", token, "expTime:", expTime)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "generate token failed", "adminID: ", req.AdminID, err.Error())
			return resp, openIMHttp.WrapError(constant.ErrTokenUnknown)
		}
		resp.Token = token
	} else {
		return resp, openIMHttp.WrapError(constant.ErrPasswordIncorrect)
	}

	// for i, adminID := range config.Config.Manager.AppManagerUid {
	//	if adminID == req.AdminID && config.Config.Manager.Secrets[i] == req.Secret {
	//		token, expTime, err := token_verify.CreateToken(adminID, constant.SingleChatType)
	//		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "generate token success", "token: ", token, "expTime:", expTime)
	//		if err != nil {
	//			log.NewError(req.OperationID, utils.GetSelfFuncName(), "generate token failed", "adminID: ", adminID, err.Error())
	//			return resp, openIMHttp.WrapError(constant.ErrTokenUnknown)
	//		}
	//		resp.Token = token
	//		break
	//	}
	// }

	if resp.Token == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "failed")
		return resp, openIMHttp.WrapError(constant.ErrTokenMalformed)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

// admin login api
func (s *adminCMSServer) AdminLoginV2(_ context.Context, req *pbAdminCMS.AdminLoginReq) (*pbAdminCMS.AdminLoginResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.AdminLoginResp{}
	user, err := imdb.GetRegAdminUsrByUID(req.AdminID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Get admin By UserID failed!", "adminID: ", req.AdminID, err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	password := req.Secret
	if !req.SecretHashd {
		newPasswordFirst := req.Secret + user.Salt
		passwordData := []byte(newPasswordFirst)
		has := md5.Sum(passwordData)
		password = fmt.Sprintf("%x", has)
	}

	if password == user.Password {
		if user.Status != 1 {
			return resp, errors.New("user is disabled/banned")
		}
		if user.User2FAuthEnable != 1 {
			req.GAuthTypeToken = false
		}
		token, expTime, err := token_verify.CreateToken(req.AdminID, constant.AdminPlatformID, req.GAuthTypeToken, 0)
		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "generate token success", "token: ", token, "expTime:", expTime)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "generate token failed", "adminID: ", req.AdminID, err.Error())
			return resp, openIMHttp.WrapError(constant.ErrTokenUnknown)
		}
		resp.Token = token
	}
	if resp.Token == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "failed")
		return resp, openIMHttp.WrapError(constant.ErrTokenMalformed)
	}
	if user.TwoFactorEnabled == 1 {
		resp.GAuthEnabled = true
	} else {
		resp.GAuthEnabled = false
	}

	resp.GAuthSetupRequired = (user.TwoFactorEnabled != 1) && config.Config.AdminUser2FAuthEnable
	if user.User2FAuthEnable != 1 {
		resp.GAuthSetupRequired = false
		resp.GAuthEnabled = false
	}
	if resp.GAuthSetupRequired {
		resp.GAuthSetupProvUri = genrateTOTPProvisUriQR(*user)
	}
	if req.SecretHashd {
		reqUpdateGauthFlag := pbAdminCMS.AlterAdminUserRequest{}
		reqUpdateGauthFlag.OpUserId = user.UserID
		reqUpdateGauthFlag.UserId = user.UserID
		reqUpdateGauthFlag.TwoFactorEnabled = 1
		imdb.AlterAdminUser(&reqUpdateGauthFlag)
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	go func() {
		if req.RequestIP != "" {
			imdb.AlterAdminUserLoginIP(user.UserID, req.RequestIP)
		}
	}()
	return resp, nil
}

// get discover url
func (s *adminCMSServer) GetDiscoverUrl(_ context.Context, req *pbAdminCMS.GetDiscoverUrlReq) (*pbAdminCMS.GetDiscoverUrlResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.GetDiscoverUrlResp{}
	urlInfo, err := imdb.GetDiscoverUrl(req.PlatformID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetDiscoverUrl failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	// Token
	if urlInfo.Url != "" && config.Config.DiscoverToken.IsOpen == 1 {
		// get user info
		user, err := imdb.GetUserByUserID(req.UserId)
		if err == nil && user.SourceCode != "" {
			token := utils.GetDiscoverToken(req.UserId, user.SourceCode)
			urlInfo.Url = urlInfo.Url + "?token=" + token
		}
	}

	resp.Url = &pbAdminCMS.DiscoverUrl{
		ID:         uint32(urlInfo.ID),
		Url:        urlInfo.Url,
		Status:     int64(urlInfo.Status),
		PlatformId: int64(urlInfo.PlatformId),
		CreateTime: urlInfo.CreateTime,
		CreateUser: urlInfo.CreateUser,
		UpdateTime: urlInfo.UpdateTime,
		UpdateUser: urlInfo.UpdateUser,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil

}

// Save Discover Url
func (s *adminCMSServer) SaveDiscoverUrl(_ context.Context, req *pbAdminCMS.SaveDiscoverUrlReq) (*pbAdminCMS.SaveDiscoverUrlResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.SaveDiscoverUrlResp{}
	err := imdb.SaveDiscoverUrl(req.Url, req.UserID, req.PlatformID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SaveDiscoverUrl failed:", err.Error())
		resp.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	resp.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

// switch the visible status of discover page
func (*adminCMSServer) SwitchDiscoverStatus(_ context.Context, req *pbAdminCMS.SwitchDiscoverStatusReq) (*pbAdminCMS.SwitchDiscoverStatusResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.SwitchDiscoverStatusResp{}
	err := imdb.SwitchDiscoverStatus(int(req.Status), req.PlatformID, req.UserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "SwitchDiscoverStatus failed:", err.Error())
		resp.CommonResp = &appversion.CommonResp{
			ErrCode: constant.ErrDB.ErrCode,
			ErrMsg:  constant.ErrDB.ErrMsg,
		}
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	resp.CommonResp = &appversion.CommonResp{
		ErrCode: constant.OK.ErrCode,
		ErrMsg:  constant.OK.ErrMsg,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (*adminCMSServer) AddAdminUser(ctx context.Context, req *pbAdminCMS.AddAdminUserReq) (*pbAdminCMS.AddAdminUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.AddAdminUserResp{}

	//check the opuser if exist
	_, err := imdb.GetRegAdminUsrByUID(req.OpUserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser Get OpUser", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	err = imdb.AddAdminUser(req)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddUser", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil
}
func (*adminCMSServer) DeleteAdminUser(ctx context.Context, req *pbAdminCMS.DeleteAdminUserReq) (*pbAdminCMS.DeleteAdminUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbAdminCMS.DeleteAdminUserResp{}
	if row := imdb.DeleteAdminUser(req.UserId, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed", "delete rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	if err := token_verify.DeleteAdminTokenOnLogout(req.UserId, false); err != nil { //req.GAuthTypeToken)
		errMsg := req.OperationID + " DeleteToken failed " + err.Error() + req.UserId + "Admin"
		log.NewError(req.OperationID, errMsg)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (*adminCMSServer) GetAdminUsers(ctx context.Context, req *pbAdminCMS.GetAdminUsersReq) (*pbAdminCMS.GetAdminUsersResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAdminUsersResp{User: []*pbAdminCMS.AdminUser{}}
	users, err := imdb.GetAdminUsers(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		// isBlock, err := imdb.UserIsBlock(v.UserID)
		if err == nil {
			user := &pbAdminCMS.AdminUser{
				UserId:           v.UserID,
				CreateTime:       v.CreateTime,
				NickName:         v.Nickname,
				Name:             v.Name,
				CreateUser:       v.CreateUser,
				UpdateUser:       v.UpdateUser,
				Status:           int32(v.Status),
				TwoFactorEnabled: v.TwoFactorEnabled,
				Role:             int64(v.Role),
				IPRangeStart:     v.IPRangeStart,
				IPRangeEnd:       v.IPRangeEnd,
				LastLoginIP:      v.LastLoginIP,
				Remarks:          v.Remarks,
				User2FAuthEnable: v.User2FAuthEnable,
				LastLoginTime:    v.LastLoginTime,
			}

			resp.User = append(resp.User, user)
		} else {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "UserIsBlock failed", err.Error())
		}
	}
	user := db.AdminUser{}
	nums, err := imdb.GetRowsCountCount(user.TableName())
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsersCount failed", err.Error(), user)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	resp.UserNums = int32(nums)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

// SearchAdminUsers(SearchAdminUsersRequest)
func (*adminCMSServer) SearchAdminUsers(ctx context.Context, req *pbAdminCMS.SearchAdminUsersRequest) (*pbAdminCMS.GetAdminUsersResp, error) {
	log.NewInfo(utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAdminUsersResp{User: []*pbAdminCMS.AdminUser{}}
	users, totalRecordsInDb, err := imdb.SearchAdminUsers(req.AccountName, int32(req.RoleID), int32(req.GAuthStatus), int32(req.Status), req.IPAddress, req.DateStart, req.DateEnd, req.PageNumber, req.PageLimit, req.CreateTimeOrLastLogin, req.Remarks)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		// isBlock, err := imdb.UserIsBlock(v.UserID)
		if err == nil {
			user := &pbAdminCMS.AdminUser{
				UserId:           v.UserID,
				CreateTime:       v.CreateTime,
				NickName:         v.Nickname,
				Name:             v.Name,
				CreateUser:       v.CreateUser,
				UpdateUser:       v.UpdateUser,
				Status:           int32(v.Status),
				Role:             int64(v.Role),
				IPRangeStart:     v.IPRangeStart,
				IPRangeEnd:       v.IPRangeEnd,
				LastLoginIP:      v.LastLoginIP,
				Remarks:          v.Remarks,
				TwoFactorEnabled: v.TwoFactorEnabled,
				User2FAuthEnable: v.User2FAuthEnable,
				LastLoginTime:    v.LastLoginTime,
			}
			resp.User = append(resp.User, user)
		} else {
			log.NewError(utils.GetSelfFuncName(), "UserIsBlock failed", err.Error())
		}
	}
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: int32(req.PageLimit), CurrentPage: int32(req.PageNumber)}
	resp.UserNums = int32(totalRecordsInDb)
	return resp, nil
}

func (*adminCMSServer) AlterAdminUser(ctx context.Context, req *pbAdminCMS.AlterAdminUserRequest) (*pbAdminCMS.AlterAdminUserResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())
	resp := &pbAdminCMS.AlterAdminUserResp{}
	if row := imdb.AlterAdminUser(req); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *adminCMSServer) SwitchAdminUserStatus(c context.Context, req *pbAdminCMS.SwitchAdminUserStatusReq) (*pbAdminCMS.SwitchAdminUserStatusResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), req.String())

	resp := &pbAdminCMS.SwitchAdminUserStatusResp{}

	opUserID := req.OpUserId
	userID := req.UserId
	status := req.Status

	_, err := imdb.GetUserByUserID(userID)
	if err != nil {
		return resp, err
	}

	updateData := db.AdminUser{
		UserID:     userID,
		Status:     int64(status),
		UpdateTime: time.Now().Unix(),
		UpdateUser: opUserID,
	}

	if err := imdb.UpdateAdminUserInfo(updateData); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update user status failed!", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	err = db.DB.SaveAdminUserStatus(userID, status)

	if err != nil {
		return resp, err
	}

	//sync local user
	log.NewInfo(req.OperationID, "SyncToLocalDataBase user", userID)
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.OperationID)
	if etcdConn == nil {
		log.Error(req.OperationID, "OpenImLocalDataName rpc connect failed ")
		return resp, openIMHttp.WrapError(constant.ErrRPC)
	}
	// client := local_database.NewLocalDataBaseClient(etcdConn)
	// reqPb := local_database.SyncUserInfoReq{
	// 	OperationID: req.OperationID,
	// 	UserID:      userID,
	// }
	// _, err2 := client.SyncUserInfoToLocal(context.Background(), &reqPb)
	// if err2 != nil {
	// 	log.Error(req.OperationID, "SyncUserInfoToLocal failed ", err2.Error(), userID)
	// 	return resp, err2
	// }

	return resp, nil
}

func (s *adminCMSServer) ChangeAdminUserPassword(_ context.Context, req *pbAdminCMS.ChangeAdminUserPasswordReq) (*pbAdminCMS.ChangeAdminUserPasswordResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.ChangeAdminUserPasswordResp{}

	user, err := imdb.GetRegAdminUsrByUID(req.UserId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Get admin By UserID failed!", "adminID: ", req.UserId, err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	twoFacAuthFalg := user.TwoFactorEnabled == 1 && config.Config.AdminUser2FAuthEnable
	totp := genrateTOTPForNow(*user)
	if req.TOTP == totp || !twoFacAuthFalg {
		password := req.Secret
		newPasswordFirst := password + user.Salt
		passwordData := []byte(newPasswordFirst)
		has := md5.Sum(passwordData)
		password = fmt.Sprintf("%x", has)

		if password == user.Password {
			passwordNew := req.NewSecret
			newPasswordFirst := passwordNew + user.Salt
			passwordData := []byte(newPasswordFirst)
			has := md5.Sum(passwordData)
			passwordNew = fmt.Sprintf("%x", has)

			updateData := db.AdminUser{
				Password: passwordNew,
			}
			if err := imdb.UpdateAdminUserPassword(updateData, user.UserID); err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update user password failed!", err.Error())
				return resp, openIMHttp.WrapError(constant.ErrDB)
			}

			if err := token_verify.DeleteAdminTokenOnLogout(user.UserID, false); err != nil { //req.GAuthTypeToken)
				errMsg := req.OperationID + " DeleteToken failed " + err.Error() + user.UserID + "Admin"
				log.NewError(req.OperationID, errMsg)
				return resp, openIMHttp.WrapError(constant.ErrDB)
			}

			token, expTime, err := token_verify.CreateToken(req.UserId, constant.AdminPlatformID, false, 0)
			log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "generate token success", "token: ", token, "expTime:", expTime)
			if err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "generate token failed", "adminID: ", req.UserId, err.Error())
				return resp, openIMHttp.WrapError(constant.ErrTokenUnknown)
			}
			resp.Token = token
			resp.PasswordUpdated = true
		}
		if resp.Token == "" {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "failed")
			return resp, openIMHttp.WrapError(constant.ErrTokenMalformed)
		}

		log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
		return resp, nil
	}
	return resp, errors.New("TOTP is not correct")

}
func (s *adminCMSServer) GetgAuthQrCode(_ context.Context, req *pbAdminCMS.GetgAuthQrCodeReq) (*pbAdminCMS.GetgAuthQrCodeResp, error) {
	resp := &pbAdminCMS.GetgAuthQrCodeResp{}

	user, err := imdb.GetRegAdminUsrByUID(req.UserId)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Get admin By UserID failed!", "adminID: ", req.UserId, err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	resp.GAuthSetupProvUri = genrateTOTPProvisUriQR(*user)
	if user.User2FAuthEnable == 1 {
		resp.UsergAuthStatus = true
	} else {
		resp.UsergAuthStatus = false
	}
	resp.GAuthAccountID = user.UserID
	resp.GAuthKey = user.Google2fSecretKey
	return resp, nil
}
func (s *adminCMSServer) AlterGAuthStatus(_ context.Context, req *pbAdminCMS.AlterGAuthStatusReq) (*pbAdminCMS.AlterGAuthStatusResp, error) {
	resp := &pbAdminCMS.AlterGAuthStatusResp{}
	err := imdb.AlterUserGAuthSatus(req)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "User GAuth status not updated by admin user id", "adminID: ", req.UserId)
		return resp, err //openIMHttp.WrapError(constant.ErrDB)
	}
	resp.GAuthStatus = req.UserGAuthStatus
	return resp, nil
}

func (s *adminCMSServer) GetAdminPermissionReq(_ context.Context, req *pbAdminCMS.AdminPermissionsReq) (*pbAdminCMS.AdminPermissionsResp, error) {
	resp := &pbAdminCMS.AdminPermissionsResp{}
	user, err := imdb.GetRegAdminUsrByUID(req.UserId)
	if err != nil || user == nil {
		// log.NewError(utils.GetSelfFuncName(), "Get admin By UserID failed!", "adminID: ", req.UserId, err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	adminRole := imdb.GetAdminPermission(*user)
	resp.AdminRole = &adminRole
	return resp, nil

}

// Admin Role CRUD API
func (s *adminCMSServer) AddAdminRole(_ context.Context, req *pbAdminCMS.AddAdminRoleRequest) (*pbAdminCMS.AddAdminRoleResponse, error) {

	resp := &pbAdminCMS.AddAdminRoleResponse{}
	// if req.AdminPagesIDs != "" {
	// 	var adminPagesListTemp []int
	// 	var adminApisListTemp []int
	// 	if err := json.Unmarshal([]byte(req.AdminPagesIDs), &adminPagesListTemp); err == nil {
	// 		fatherAdminPageList := imdb.GetAdminPages(adminPagesListTemp)
	// 		for _, v := range fatherAdminPageList {
	// 			var pageAdminAPITemp []int
	// 			if err := json.Unmarshal([]byte(v.AdminAPIsIDs), &pageAdminAPITemp); err == nil {
	// 				adminApisListTemp = append(adminApisListTemp, pageAdminAPITemp...)
	// 			}
	// 		}
	// 		adminApisList := removeDuplicateInt(adminApisListTemp)
	// 		if len(adminApisList) > 0 {
	// 			adminApisListBytes, err := json.Marshal(adminApisList)
	// 			if err == nil {
	// 				req.AdminAPIsIDs = string(adminApisListBytes[:])
	// 			}
	// 		}
	// 	}
	// }

	err := imdb.AddAdminRole(req)
	return resp, err

}

func removeDuplicateInt(intSlice []int) []int {
	allKeys := make(map[int]bool)
	list := []int{}
	for _, item := range intSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func (s *adminCMSServer) AlterAdminRole(_ context.Context, req *pbAdminCMS.AlterAdminRoleRequest) (*pbAdminCMS.AlterAdminRoleResponse, error) {
	resp := &pbAdminCMS.AlterAdminRoleResponse{}

	// if req.AdminPagesIDs != "" {
	// 	var adminPagesListTemp []int
	// 	var adminApisListTemp []int
	// 	if err := json.Unmarshal([]byte(req.AdminPagesIDs), &adminPagesListTemp); err == nil {
	// 		fatherAdminPageList := imdb.GetAdminPages(adminPagesListTemp)
	// 		for _, v := range fatherAdminPageList {
	// 			var pageAdminAPITemp []int
	// 			if err := json.Unmarshal([]byte(v.AdminAPIsIDs), &pageAdminAPITemp); err == nil {
	// 				adminApisListTemp = append(adminApisListTemp, pageAdminAPITemp...)
	// 			}
	// 		}
	// 		adminApisList := removeDuplicateInt(adminApisListTemp)
	// 		if len(adminApisList) > 0 {
	// 			adminApisListBytes, err := json.Marshal(adminApisList)
	// 			if err == nil {
	// 				req.AdminAPIsIDs = string(adminApisListBytes[:])
	// 			}
	// 		}
	// 	}
	// }

	row := imdb.AlterAdminRole(req)
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil

}

func (s *adminCMSServer) DeleteAdminRole(_ context.Context, req *pbAdminCMS.AlterAdminRoleRequest) (*pbAdminCMS.AlterAdminRoleResponse, error) {
	resp := &pbAdminCMS.AlterAdminRoleResponse{}
	assignedCount := imdb.CheckAdminRoleAssigned(req.AdminRoleID)
	if assignedCount > 0 {
		return resp, openIMHttp.WrapError(constant.AdminRoleDeleteError)
	}
	row := imdb.DeletedminRole(req)
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil

}

func (*adminCMSServer) GetAllAdminRoles(ctx context.Context, req *pbAdminCMS.GetAllAdminRolesReq) (*pbAdminCMS.GetAllAdminRolesResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllAdminRolesResp{AdminRoles: []*pbAdminCMS.AdminRoleResp{}}
	users, err := imdb.GetAllAdminRoles(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		adminRole := &pbAdminCMS.AdminRoleResp{
			AdminRoleID:   v.AdminRoleID,
			AdminRoleName: v.AdminRoleName,
			AdminAPIsIDs:  v.AdminAPIsIDs,
			AdminPagesIDs: v.AdminPagesIDs,
			CreateUser:    v.CreateUser,
			CreateTime:    v.CreateTime,
			// IsBlock:      isBlock,
			Status:               int64(v.Status),
			UpdateUser:           v.UpdateUser,
			UpdateTime:           v.UpdateTime,
			AdminRoleDiscription: v.AdminRoleDiscription,
			AdminRoleRemarks:     v.AdminRoleRemarks,
		}
		resp.AdminRoles = append(resp.AdminRoles, adminRole)
	}
	nums, err := imdb.GetRowsCountCount(db.AdminRole{}.TableName())
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsersCount failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	resp.AdminRolesNums = int32(nums)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}
func (*adminCMSServer) SearchAminRoles(ctx context.Context, req *pbAdminCMS.SearchAminRolesRequest) (*pbAdminCMS.GetAllAdminRolesResp, error) {
	log.NewInfo(utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllAdminRolesResp{AdminRoles: []*pbAdminCMS.AdminRoleResp{}}
	users, totalRecordsInDb, err := imdb.SearchAminRoles(req.RoleName, req.Description, req.PageNumber, req.PageLimit)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		adminRole := &pbAdminCMS.AdminRoleResp{
			AdminRoleID:   v.AdminRoleID,
			AdminRoleName: v.AdminRoleName,
			AdminAPIsIDs:  v.AdminAPIsIDs,
			AdminPagesIDs: v.AdminPagesIDs,
			CreateUser:    v.CreateUser,
			CreateTime:    v.CreateTime,
			// IsBlock:      isBlock,
			Status:               int64(v.Status),
			UpdateUser:           v.UpdateUser,
			UpdateTime:           v.UpdateTime,
			AdminRoleDiscription: v.AdminRoleDiscription,
			AdminRoleRemarks:     v.AdminRoleRemarks,
		}
		resp.AdminRoles = append(resp.AdminRoles, adminRole)
	}
	resp.AdminRolesNums = int32(totalRecordsInDb)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: int32(req.PageLimit), CurrentPage: int32(req.PageNumber)}
	log.NewInfo(utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

// API Admin Role CURD api
func (s *adminCMSServer) AddApiAdminRole(_ context.Context, req *pbAdminCMS.AddApiAdminRoleRequest) (*pbAdminCMS.AddApiAdminRoleResponse, error) {
	resp := &pbAdminCMS.AddApiAdminRoleResponse{}
	err := imdb.AddApiAdminRole(req)
	return resp, err
}

func (s *adminCMSServer) AlterApiAdminRole(_ context.Context, req *pbAdminCMS.AlterApiAdminRoleRequest) (*pbAdminCMS.AlterApiAdminRoleResponse, error) {
	resp := &pbAdminCMS.AlterApiAdminRoleResponse{}
	row := imdb.AlterApiAdminRole(req)
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil
}

func (s *adminCMSServer) DeleteApiAdminRole(_ context.Context, req *pbAdminCMS.AlterApiAdminRoleRequest) (*pbAdminCMS.AlterApiAdminRoleResponse, error) {
	resp := &pbAdminCMS.AlterApiAdminRoleResponse{}
	row := imdb.DeleteApiAdminRole(req)
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil

}

func (*adminCMSServer) GetAllApiAdminRoles(ctx context.Context, req *pbAdminCMS.GetAllApiAdminRolesReq) (*pbAdminCMS.GetAllApiAdminRolesResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllApiAdminRolesResp{ApisAdminRole: []*pbAdminCMS.ApiAdminRoleResp{}}
	users, totalRecordsInDb, err := imdb.GetAllApiAdminRoles(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		adminRole := &pbAdminCMS.ApiAdminRoleResp{
			ApiID:      v.ApiID,
			ApiName:    v.ApiName,
			ApiPath:    v.ApiPath,
			CreateUser: v.CreateUser,
			UpdateUser: v.UpdateUser,
			Status:     int32(v.Status),
			CreateTime: v.CreateTime,
			UpdateTime: v.UpdateTime,
		}
		resp.ApisAdminRole = append(resp.ApisAdminRole, adminRole)
	}
	// nums, err := imdb.GetRowsCountCount(db.AdminAPIs{}.TableName())
	// if err != nil {
	// 	log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsersCount failed", err.Error())
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	resp.ApiNums = int32(totalRecordsInDb)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}
func (*adminCMSServer) SearchApiAdminRoles(ctx context.Context, req *pbAdminCMS.SearchApiAdminRoleRequest) (*pbAdminCMS.GetAllApiAdminRolesResp, error) {
	log.NewInfo(utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllApiAdminRolesResp{ApisAdminRole: []*pbAdminCMS.ApiAdminRoleResp{}}
	users, totalRecordsInDb, err := imdb.SearchApiAdminRoles(req)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range users {
		adminRole := &pbAdminCMS.ApiAdminRoleResp{
			ApiID:      v.ApiID,
			ApiName:    v.ApiName,
			ApiPath:    v.ApiPath,
			CreateUser: v.CreateUser,
			UpdateUser: v.UpdateUser,
			Status:     int32(v.Status),
			CreateTime: v.CreateTime,
			UpdateTime: v.UpdateTime,
		}
		resp.ApisAdminRole = append(resp.ApisAdminRole, adminRole)
	}
	// nums, err := imdb.GetRowsCountCount(db.AdminAPIs{}.TableName())
	// if err != nil {
	// 	log.NewError(utils.GetSelfFuncName(), "GetUsersCount failed", err.Error())
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	resp.ApiNums = int32(totalRecordsInDb)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: int32(req.PageLimit), CurrentPage: int32(req.PageNumber)}
	log.NewInfo(utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

// Pages Admin Role CURD api
func (s *adminCMSServer) AddPageAdminRole(_ context.Context, req *pbAdminCMS.AddPageAdminRoleRequest) (*pbAdminCMS.AddPageAdminRoleResponse, error) {
	resp := &pbAdminCMS.AddPageAdminRoleResponse{}
	err := imdb.AddPageAdminRole(req)
	return resp, err

}

func (s *adminCMSServer) AlterPageAdminRole(_ context.Context, req *pbAdminCMS.AlterPageAdminRoleRequest) (*pbAdminCMS.AlterPageAdminRoleResponse, error) {
	resp := &pbAdminCMS.AlterPageAdminRoleResponse{}
	row := imdb.AlterPageAdminRole(req)
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil

}

func (s *adminCMSServer) DeletePageAdminRole(_ context.Context, req *pbAdminCMS.AlterPageAdminRoleRequest) (*pbAdminCMS.AlterPageAdminRoleResponse, error) {
	resp := &pbAdminCMS.AlterPageAdminRoleResponse{}
	row := imdb.DeletePageAdminRole(req)
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	return resp, nil

}

func (*adminCMSServer) GetAllPageAdminRoles(ctx context.Context, req *pbAdminCMS.GetAllPageAdminRolesReq) (*pbAdminCMS.GetAllPageAdminRolesResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllPageAdminRolesResp{PagesAdminRole: []*pbAdminCMS.PageAdminRoleResp{}}
	users, totalRecordsInDb, err := imdb.GetAllPageAdminRoles(req.FatherIDFilter, req.Pagination.ShowNumber, req.Pagination.PageNumber)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	var adminApisList []int
	var fatherAdminPagesList []int
	for _, v := range users {
		pAdminRole := &pbAdminCMS.PageAdminRoleResp{
			PageID:       v.PageID,
			PageName:     v.PageName,
			PagePath:     v.PagePath,
			CreateUser:   v.CreateUser,
			UpdateUser:   v.UpdateUser,
			Status:       int32(v.Status),
			FatherPageID: v.FatherPageID,
			IsMenu:       v.IsMenu,
			IsButton:     v.IsButton,
			AdminAPIsIDs: v.AdminAPIsIDs,
			SortPriority: v.SortPriority,
			CreateTime:   v.CreateTime,
			UpdateTime:   v.UpdateTime,
		}
		if v.FatherPageID != 0 {
			fatherAdminPagesList = append(fatherAdminPagesList, int(v.FatherPageID))
		}
		resp.PagesAdminRole = append(resp.PagesAdminRole, pAdminRole)

		if v.AdminAPIsIDs != "" {
			var adminApisListTemp []int
			if err := json.Unmarshal([]byte(v.AdminAPIsIDs), &adminApisListTemp); err == nil {
				adminApisList = append(adminApisList, adminApisListTemp...)
			}
		}
	}
	if len(fatherAdminPagesList) > 0 {
		fatherAdminPageList := imdb.GetAdminPages(fatherAdminPagesList)
		for _, v := range fatherAdminPageList {
			pAdminRole := &pbAdminCMS.PageAdminRoleResp{
				PageID:       v.PageID,
				PageName:     v.PageName,
				PagePath:     v.PagePath,
				CreateUser:   v.CreateUser,
				UpdateUser:   v.UpdateUser,
				Status:       int32(v.Status),
				FatherPageID: v.FatherPageID,
				IsMenu:       v.IsMenu,
				IsButton:     v.IsButton,
				AdminAPIsIDs: v.AdminAPIsIDs,
				SortPriority: v.SortPriority,
			}
			resp.FatherPagesAdminRole = append(resp.FatherPagesAdminRole, pAdminRole)
		}
	}

	if len(adminApisList) > 0 {
		adminAPIRoles := imdb.GetAdminAPIs(adminApisList)
		for _, v := range adminAPIRoles {
			adminRole := &pbAdminCMS.ApiAdminRoleResp{
				ApiID:      v.ApiID,
				ApiName:    v.ApiName,
				ApiPath:    v.ApiPath,
				CreateUser: v.CreateUser,
				UpdateUser: v.UpdateUser,
				Status:     int32(v.Status),
			}
			resp.ApisAdminRole = append(resp.ApisAdminRole, adminRole)
		}
	}

	// nums, err := imdb.GetRowsCountCount(db.AdminPages{}.TableName())
	// if err != nil {
	// 	log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsersCount failed", err.Error())
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	resp.TotalRecCount = int32(totalRecordsInDb)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}
func (*adminCMSServer) SearchPageAdminRoles(ctx context.Context, req *pbAdminCMS.SearchPageAdminRolesRequest) (*pbAdminCMS.GetAllPageAdminRolesResp, error) {
	log.NewInfo(utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllPageAdminRolesResp{PagesAdminRole: []*pbAdminCMS.PageAdminRoleResp{}}
	users, totalRecordsInDb, err := imdb.SearchPageAdminRoles(req)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	var adminApisList []int
	var fatherAdminPagesList []int
	for _, v := range users {
		pAdminRole := &pbAdminCMS.PageAdminRoleResp{
			PageID:       v.PageID,
			PageName:     v.PageName,
			PagePath:     v.PagePath,
			CreateUser:   v.CreateUser,
			UpdateUser:   v.UpdateUser,
			Status:       int32(v.Status),
			FatherPageID: v.FatherPageID,
			IsMenu:       v.IsMenu,
			IsButton:     v.IsButton,
			AdminAPIsIDs: v.AdminAPIsIDs,
			SortPriority: v.SortPriority,
		}
		if v.FatherPageID != 0 {
			fatherAdminPagesList = append(fatherAdminPagesList, int(v.FatherPageID))
		}
		resp.PagesAdminRole = append(resp.PagesAdminRole, pAdminRole)

		if v.AdminAPIsIDs != "" {
			var adminApisListTemp []int
			if err := json.Unmarshal([]byte(v.AdminAPIsIDs), &adminApisListTemp); err == nil {
				adminApisList = append(adminApisList, adminApisListTemp...)
			}
		}

	}
	if len(fatherAdminPagesList) > 0 {
		fatherAdminPageList := imdb.GetAdminPages(fatherAdminPagesList)
		for _, v := range fatherAdminPageList {
			pAdminRole := &pbAdminCMS.PageAdminRoleResp{
				PageID:       v.PageID,
				PageName:     v.PageName,
				PagePath:     v.PagePath,
				CreateUser:   v.CreateUser,
				UpdateUser:   v.UpdateUser,
				Status:       int32(v.Status),
				FatherPageID: v.FatherPageID,
				IsMenu:       v.IsMenu,
				IsButton:     v.IsButton,
				AdminAPIsIDs: v.AdminAPIsIDs,
				SortPriority: v.SortPriority,
			}
			resp.FatherPagesAdminRole = append(resp.FatherPagesAdminRole, pAdminRole)
		}
	}

	if len(adminApisList) > 0 {
		adminAPIRoles := imdb.GetAdminAPIs(adminApisList)
		for _, v := range adminAPIRoles {
			adminRole := &pbAdminCMS.ApiAdminRoleResp{
				ApiID:      v.ApiID,
				ApiName:    v.ApiName,
				ApiPath:    v.ApiPath,
				CreateUser: v.CreateUser,
				UpdateUser: v.UpdateUser,
				Status:     int32(v.Status),
			}
			resp.ApisAdminRole = append(resp.ApisAdminRole, adminRole)
		}
	}
	resp.TotalRecCount = int32(totalRecordsInDb)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: int32(req.PageLimit), CurrentPage: int32(req.PageNumber)}
	log.NewInfo(utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

// Admin Actions CRUD API
func (s *adminCMSServer) AddAdminAction(_ context.Context, req *pbAdminCMS.AddAdminActionRequest) (*pbAdminCMS.AddAdminActionResponse, error) {
	resp := &pbAdminCMS.AddAdminActionResponse{}
	// err := imdb.AddAdminAction(req)
	return resp, nil

}

func (s *adminCMSServer) AlterAdminAction(_ context.Context, req *pbAdminCMS.AlterAdminActionRequest) (*pbAdminCMS.AlterAdminActionResponse, error) {
	resp := &pbAdminCMS.AlterAdminActionResponse{}
	// row := imdb.AlterAdminAction(req)
	// if row == 0 {
	// 	log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	return resp, nil

}

func (s *adminCMSServer) DeleteAdminAction(_ context.Context, req *pbAdminCMS.AlterAdminActionRequest) (*pbAdminCMS.AlterAdminActionResponse, error) {
	resp := &pbAdminCMS.AlterAdminActionResponse{}
	// row := imdb.DeletedminAction(req)
	// if row == 0 {
	// 	log.NewError(req.OperationID, utils.GetSelfFuncName(), "Update failed", "update rows:", row)
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	return resp, nil

}

func (*adminCMSServer) GetAllAdminAction(ctx context.Context, req *pbAdminCMS.GetAllAdminActionReq) (*pbAdminCMS.GetAllAdminActionResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req: ", req.String())
	resp := &pbAdminCMS.GetAllAdminActionResp{AdminActions: []*pbAdminCMS.AdminActionResp{}}
	// adminActions, err := imdb.GetAllAdminAction(req.Pagination.ShowNumber, req.Pagination.PageNumber)
	// if err != nil {
	// 	log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsers failed", err.Error())
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	// for _, v := range adminActions {
	// 	adminAction := &pbAdminCMS.AdminActionResp{
	// 		AdminActionID:   v.AdminActionID,
	// 		AdminActionName: v.AdminActionName,
	// 		AdminAPIsIDs:    v.AdminAPIsIDs,
	// 		AdminPagesIDs:   v.AdminPagesIDs,
	// 		CreateUser:      v.CreateUser,
	// 		CreateTime:      v.CreateTime,
	// 		Status:          int64(v.Status),
	// 	}
	// 	resp.AdminActions = append(resp.AdminActions, adminAction)
	// }
	// nums, err := imdb.GetRowsCountCount(db.AdminActions{}.TableName())
	// if err != nil {
	// 	log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUsersCount failed", err.Error())
	// 	return resp, openIMHttp.WrapError(constant.ErrDB)
	// }
	// resp.AdminActionNums = int32(nums)
	// resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	// log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}
func (*adminCMSServer) OperationLog(ctx context.Context, req *pbAdminCMS.OperationLogRequest) (*pbAdminCMS.OperationLogRequest, error) {
	db.DB.AddRequestToOpertaionLog(req)
	return req, nil
}

func (*adminCMSServer) SearchOperationLogs(ctx context.Context, req *pbAdminCMS.SearchOperationLogsRequest) (*pbAdminCMS.SearchOperationLogsResponse, error) {
	log.NewInfo(utils.GetSelfFuncName(), "req: ", req.String())
	var searchOperationLogsResponse pbAdminCMS.SearchOperationLogsResponse
	opLogsResponse, totalRecordsInDb, err := db.DB.SearchOpertaionLog(req)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "GetUsers failed", err.Error())
		return &searchOperationLogsResponse, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range opLogsResponse {
		operationLog := &pbAdminCMS.OperationLogRequest{
			Operator:   v.Operator,
			Action:     v.Action,
			Payload:    v.Payload,
			OperatorIP: v.OperatorIP,
			Executee:   v.Executee,
			CreateTime: v.CreateTime,
		}
		searchOperationLogsResponse.OperationLogs = append(searchOperationLogsResponse.OperationLogs, operationLog)
	}
	log.NewInfo(utils.GetSelfFuncName(), "resp: ", searchOperationLogsResponse.String())
	searchOperationLogsResponse.PageLimit = req.PageLimit
	searchOperationLogsResponse.PageNumber = req.PageNumber
	searchOperationLogsResponse.OperationLogsCount = totalRecordsInDb
	return &searchOperationLogsResponse, nil
}

// genrateTOTPProvisUriQR genrating QR code URI for user
// TOTP -> Time-Based One-Time Password
func genrateTOTPProvisUriQR(user db.AdminUser) string {
	// totp := gotp.NewDefaultTOTP(user.Google2fSecretKey)
	secret, _ := gotp.DecodeBase32(user.Google2fSecretKey)
	totp, _ := gotp.NewTOTP(secret)
	provisingUri, err := totp.ProvisioningURI(user.UserID, config.Config.TotpIssuerName)
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Genrating TOTP Provision URL Failed", err.Error())
	}
	return provisingUri
}

// Utils Function blewThis line
//
// genrateTOTPForNow genrate TOTP code for varification
func genrateTOTPForNow(user db.AdminUser) string {
	// totp := gotp.NewDefaultTOTP(user.Google2fSecretKey)
	secret, _ := gotp.DecodeBase32(user.Google2fSecretKey)
	totp, _ := gotp.NewTOTP(secret)
	totpCode, err := totp.Now()
	if err != nil {
		log.NewError(utils.GetSelfFuncName(), "Genrating TOTP Failed", err.Error())
	}
	return totpCode
}

func (s *adminCMSServer) GetInterests(_ context.Context, req *pbAdminCMS.GetInterestsReq) (*pbAdminCMS.GetInterestsResp, error) {
	log.Debug(req.OperationID, "Get interests, req: ", req.String())
	resp := &pbAdminCMS.GetInterestsResp{Interests: []*pbAdminCMS.InterestResp{}}
	where := map[string]string{}
	where["name"] = req.Name
	where["create_user"] = req.CreateUser
	where["remark"] = req.Remark
	where["time_type"] = strconv.Itoa(int(req.TimeType))
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime
	where["status"] = req.Status
	where["order_by"] = req.OrderBy
	where["is_default"] = strconv.Itoa(int(req.IsDefault))

	interests, interestsCount, err := imdb.GetInterestsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetInterests failed", err.Error())
		return resp, err
	}
	log.Debug("", "interests len: ", len(interests))
	for _, interest := range interests {
		interestPb := pbAdminCMS.InterestResp{}
		_ = utils.CopyStructFields(&interestPb, &interest)
		interestPb.Name = []*pbAdminCMS.InterestLanguage{}
		for _, language := range interest.Name {
			languagePb := pbAdminCMS.InterestLanguage{}
			utils.CopyStructFields(&languagePb, &language)
			interestPb.Name = append(interestPb.Name, &languagePb)
		}
		resp.Interests = append(resp.Interests, &interestPb)
	}

	resp.InterestNums = int32(interestsCount)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func updateInterestsInRedis() {
	allInterests := imdb.GetAllInterestType()

	// key: en/ch/ar  value: [interest id:name, interest id:name...]
	data := make(map[string][]map[string]interface{}, 0)

	for _, interest := range allInterests {
		log.Debug("", "interest status: ", interest.Status)
		for _, language := range interest.Name {
			log.Debug("", "language Name: ", language.Name)
			if _, ok := data[language.LanguageType]; !ok {
				data[language.LanguageType] = make([]map[string]interface{}, 0)
			}
			data[language.LanguageType] = append(data[language.LanguageType], map[string]interface{}{
				"id":   interest.Id,
				"name": language.Name,
			})
		}
	}
	log.Debug("", data)
	err := db.DB.DeleteInterestList()
	if err != nil {
		log.NewError("", "redis delete interest list error")
		return
	}
	for key, oneData := range data {
		dataJson, _ := json.Marshal(oneData)
		err = db.DB.SetInterestListByLanguage(key, string(dataJson))
		if err != nil {
			log.NewError("", "redis set interest list error")
			return
		}
	}

}

func (s *adminCMSServer) DeleteInterests(ctx context.Context, req *pbAdminCMS.DeleteInterestsReq) (*pbAdminCMS.DeleteInterestsResp, error) {
	log.Debug(req.OperationID, "Delete moments req: ", req.String())
	resp := &pbAdminCMS.DeleteInterestsResp{}
	if req.Interests == "" {
		return resp, openIMHttp.WrapError(constant.ErrArgs)
	}
	interestsIdList := strings.Split(req.Interests, ",")
	var row int64
	var err error
	if row, err = imdb.DeleteInterests(interestsIdList, req.OpUserId); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed")
		return resp, err
	}
	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "delete failed")
		return resp, err
	}

	go updateInterestsInRedis()

	return resp, nil
}

func (s *adminCMSServer) AlterInterest(ctx context.Context, req *pbAdminCMS.AlterInterestReq) (*pbAdminCMS.AlterInterestResp, error) {
	log.Debug(req.OperationID, "Change interest status req: ", req.String())
	resp := &pbAdminCMS.AlterInterestResp{}
	interest := db.InterestType{}
	var names []db.InterestLanguage
	utils.CopyStructFields(&interest, req)
	if req.Name != nil {
		for _, reqName := range req.Name {
			language := db.InterestLanguage{}
			utils.CopyStructFields(&language, &reqName)
			names = append(names, language)
		}
	}
	if row := imdb.AlterInterest(&interest, names, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Change interest Status failed", "Change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go updateInterestsInRedis()

	return resp, nil
}

func (s *adminCMSServer) ChangeInterestStatus(ctx context.Context, req *pbAdminCMS.ChangeInterestStatusReq) (*pbAdminCMS.ChangeInterestStatusResp, error) {
	log.Debug(req.OperationID, "Change interest status req: ", req.String())
	resp := &pbAdminCMS.ChangeInterestStatusResp{}
	if row := imdb.ChangeInterestStatus(req.InterestId, req.Status, req.OpUserId); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Change interest Status failed", "Change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go updateInterestsInRedis()

	return resp, nil
}

func (s *adminCMSServer) AddInterests(_ context.Context, req *pbAdminCMS.AddInterestsReq) (*pbAdminCMS.AddInterestsResp, error) {
	log.Debug(req.OperationID, "AddInterests req: ", req.String())
	interests := req.Interests
	resp := &pbAdminCMS.AddInterestsResp{}
	if err := imdb.AddInterests(interests, req.OpUserId); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Add tags failed:", err)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	go updateInterestsInRedis()

	return resp, nil
}

func (s *adminCMSServer) GetUserInterests(_ context.Context, req *pbAdminCMS.GetUserInterestsReq) (*pbAdminCMS.GetUserInterestsResp, error) {
	log.Debug(req.OperationID, "Get interests, req: ", req.String())
	resp := &pbAdminCMS.GetUserInterestsResp{Interests: []*pbAdminCMS.UserInterests{}}
	where := map[string]string{}
	where["interest_name"] = req.InterestName
	where["account"] = req.Account
	where["default"] = strconv.Itoa(int(req.Default))
	where["order_by"] = req.OrderBy
	where["user_id"] = req.UserID

	interests, interestCounts, err := imdb.GetUserInterestsByWhereV3(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetInterests failed", err.Error())
		return resp, err
	}
	_ = utils.CopyStructFields(&resp.Interests, interests)

	resp.InterestNums = int32(interestCounts)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *adminCMSServer) AlterUserInterests(_ context.Context, req *pbAdminCMS.AlterUserInterestsReq) (*pbAdminCMS.AlterUserInterestsResp, error) {
	resp := &pbAdminCMS.AlterUserInterestsResp{}
	var interestList []int64

	if req.Interests != "" {
		interestListStr := strings.Split(req.Interests, ",")
		for _, interest := range interestListStr {
			interestInt, err := strconv.ParseInt(interest, 10, 64)
			if err != nil {
				return resp, err
			}
			interestList = append(interestList, interestInt)
		}
	}

	if row := imdb.AlterUserInterests(req.UserID, interestList); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Alter user interest failed", "Change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (s *adminCMSServer) DeleteUserInterests(_ context.Context, req *pbAdminCMS.DeleteUserInterestsReq) (*pbAdminCMS.DeleteUserInterestsResp, error) {
	resp := &pbAdminCMS.DeleteUserInterestsResp{}
	var userIdList []string

	if req.UsersID != "" {
		userList := strings.Split(req.UsersID, ",")
		for _, userId := range userList {
			userIdList = append(userIdList, userId)
		}
	}

	var row int64 = 0
	for _, userId := range userIdList {
		row += imdb.AlterUserInterests(userId, []int64{})
	}

	if row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Alter user interest failed", "Change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (s *adminCMSServer) GetGroupInterests(_ context.Context, req *pbAdminCMS.GetGroupInterestsReq) (*pbAdminCMS.GetGroupInterestsResp, error) {
	log.Debug(req.OperationID, "Get GroupInterests, req: ", req.String())
	resp := &pbAdminCMS.GetGroupInterestsResp{Interests: []*pbAdminCMS.GroupInterests{}}

	where := map[string]string{}
	where["interest_name"] = req.InterestName
	where["group_name"] = req.GroupName
	where["creator_user"] = req.CreatorUser
	where["order_by"] = req.OrderBy
	where["group_id"] = req.GroupID

	interests, interestsCount, err := imdb.GetGroupInterestsByWhere(where, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "GetInterests failed", err.Error())
		return resp, err
	}
	_ = utils.CopyStructFields(&resp.Interests, interests)

	// Get total interests count
	resp.InterestNums = int32(interestsCount)
	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp: ", resp.String())
	return resp, nil
}

func (s *adminCMSServer) AlterGroupInterests(_ context.Context, req *pbAdminCMS.AlterGroupInterestsReq) (*pbAdminCMS.AlterGroupInterestsResp, error) {
	resp := &pbAdminCMS.AlterGroupInterestsResp{}
	var interestList []int64

	if req.Interests != "" {
		interestListStr := strings.Split(req.Interests, ",")
		for _, interest := range interestListStr {
			interestInt, err := strconv.ParseInt(interest, 10, 64)
			if err != nil {
				return resp, err
			}
			interestList = append(interestList, interestInt)
		}
	}

	if row := imdb.AlterGroupInterests(req.GroupID, interestList); row == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Alter user interest failed", "Change rows:", row)
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	return resp, nil
}

func (s *adminCMSServer) GetMePageURLs(_ context.Context, req *pbAdminCMS.GetMePageURLsReq) (*pbAdminCMS.GetMePageURLsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.GetMePageURLsResp{}

	allMePageList, err := imdb.GetMePageURLs()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetDiscoverUrl failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.MePageURL = make(map[string]*pbAdminCMS.MePageObj)
	for _, pageURL := range allMePageList {
		mePageType := ""
		switch pageURL.Type {
		case constant.MePageTypeOTC:
			mePageType = "otc"
		case constant.MePageTypeDeposit:
			mePageType = "deposit"
		case constant.MePageTypeWithdraw:
			mePageType = "withdraw"
		case constant.MePageTypeExchange:
			mePageType = "exchange"
		case constant.MePageTypeMarket:
			mePageType = "market"
		case constant.MePageTypeEarn:
			mePageType = "earn"
		case constant.MePageGameStore:
			mePageType = "game_store"
		case constant.MePageDiscover:
			mePageType = "discover"
		}
		if pageType, ok := resp.MePageURL[mePageType]; ok {
			pageType.URLMap[pageURL.Language] = pageURL.Url
		} else {
			resp.MePageURL[mePageType] = &pbAdminCMS.MePageObj{
				Status: int32(pageURL.Status),
				URLMap: map[string]string{
					pageURL.Language: pageURL.Url,
				},
			}
		}
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (s *adminCMSServer) GetMePageURL(_ context.Context, req *pbAdminCMS.GetMePageURLReq) (*pbAdminCMS.GetMePageURLResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.GetMePageURLResp{}
	mePageInfo, err := imdb.GetMePageUrl(int32(req.PageType))
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetDiscoverUrl failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	resp.Url = make(map[string]string)
	if len(mePageInfo) == 0 {
		resp.Url["en"] = ""
		resp.Url["cn"] = ""
		resp.Status = 2
		log.Debug("", "No record return empty")
		return resp, nil
	}
	_ = utils.CopyStructFields(resp, mePageInfo[0])
	resp.Status = int32(mePageInfo[0].Status)
	for _, pageURL := range mePageInfo {
		resp.Url[pageURL.Language] = pageURL.Url
	}

	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (s *adminCMSServer) SaveMePageURL(_ context.Context, req *pbAdminCMS.SaveMePageURLReq) (*pbAdminCMS.SaveMePageURLResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.SaveMePageURLResp{}

	err := imdb.SaveMePageUrl(req.Url, int32(req.PageType), req.OpUserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetDiscoverUrl failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.CommonResp = &pbAdminCMS.CommonResp{}
	resp.CommonResp.ErrMsg = "save success"
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (s *adminCMSServer) SwitchMePageURL(_ context.Context, req *pbAdminCMS.SwitchMePageURLReq) (*pbAdminCMS.SwitchMePageURLResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdminCMS.SwitchMePageURLResp{}
	page := &db.MePageURL{
		UpdateUser: req.OpUserID,
		Type:       req.PageType,
		Status:     int(req.Status),
	}
	err := imdb.SwitchMePageUrl(page)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetDiscoverUrl failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.CommonResp = &pbAdminCMS.CommonResp{}
	resp.CommonResp.ErrMsg = "switch success"
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func GetTencentCloudClient(isPersistent bool) (*cos.Client, error) {
	cli := sts.NewClient(
		config.Config.Credential.Tencent.SecretID,
		config.Config.Credential.Tencent.SecretKey,
		nil,
	)
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(time.Hour.Seconds()),
		Region:          config.Config.Credential.Tencent.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					Action: []string{
						"name/cos:PutObjectTagging",
						"name/cos:DeleteObjectTagging",
						"name/cos:HeadObject",
						"name/cos:DeleteObject",
					},
					Effect: "allow",
					Resource: []string{
						"qcs::cos:" + config.Config.Credential.Tencent.Region + ":uid/" + config.Config.Credential.Tencent.AppID + ":" + config.Config.Credential.Tencent.Bucket + "/*",
						"qcs::cos:" + config.Config.Credential.Tencent.Region + ":uid/" + config.Config.Credential.Tencent.AppID + ":" + config.Config.Credential.Tencent.PersistenceBucket + "/*",
					},
				},
			},
		},
	}
	COSCredential, err := cli.GetCredential(opt)
	if err != nil {
		log.NewError("", "get tencent cloud client failed")
		return nil, err
	}

	dir := ""
	if isPersistent {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket, config.Config.Credential.Tencent.Region)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.Bucket, config.Config.Credential.Tencent.Region)
	}
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSCredential.Credentials.TmpSecretID,
			SecretKey:    COSCredential.Credentials.TmpSecretKey,
			SessionToken: COSCredential.Credentials.SessionToken,
		},
	})

	return client, err
}

func RemoveDeleteTagForPersistent(client *cos.Client, urls []string) ([]string, error) {
	var errUrls []string
	for _, delUrl := range urls {
		objName := GetObjNameByURL(delUrl, true)
		_, err := client.Object.DeleteTagging(context.Background(), objName)
		if err != nil {
			fmt.Println("delete tag is failed:", err.Error())
			log.NewError("", "delete tag is failed:", err.Error())
			errUrls = append(errUrls, delUrl)
		}
	}

	return errUrls, nil
}

func DeleteFileForPersistent(client *cos.Client, url string) error {
	objName := GetObjNameByURL(url, true)
	_, err := client.Object.Delete(context.Background(), objName)
	if err != nil {
		log.Error("", "Delete file from tencent cloud failed.")
		return err
	}
	return nil
}

func GetObjNameByURL(url string, isPersistent bool) string {
	dir := ""
	if isPersistent {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket, config.Config.Credential.Tencent.Region)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.Bucket, config.Config.Credential.Tencent.Region)
	}
	if !strings.HasPrefix(url, dir) {
		accelerateDir := ""
		if isPersistent {
			accelerateDir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket)
		} else {
			accelerateDir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.Bucket)
		}
		url = strings.ReplaceAll(url, accelerateDir, dir)
	}
	objName := strings.Replace(url, dir+"/", "", 1)
	return objName
}

func RemoveFiles(urls []string) error {
	var obs []cos.Object
	for _, url := range urls {
		objName := GetObjNameByURL(url, true)
		obs = append(obs, cos.Object{Key: objName})
	}

	if len(obs) == 0 {
		return nil
	}
	deleteOpt := &cos.ObjectDeleteMultiOptions{Objects: obs}
	client, err := GetTencentCloudClient(true)
	if err != nil {
		log.Error("", "Tencent cloud get failed")
		return err
	}

	_, _, err = client.Object.DeleteMulti(context.Background(), deleteOpt)
	if err != nil {
		client, err = GetTencentCloudClient(false)
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
