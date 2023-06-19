package main

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/network"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	log2 "log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type PressureTest struct {
	OperationID              string
	Step                     int64
	DataDir                  string
	ApiUrl                   string
	ISHTTPS                  bool
	ApiTimeOutSecond         int
	GroupID                  string
	GroupOwnerAccount        string
	GroupOwnerPassword       string
	GroupOwnerToken          string
	AdminAccount             string
	AdminPassword            string
	AdminToken               string
	UserNum                  int64
	UserNamePrefix           string
	RegisterSuccessNum       int
	MessageNum               int64
	MessageIntervalTimeMilli int
	SendSuccMsgNum           int
	SendFailedMsgNum         int
	RecvMsgNum               int
}

type TestUser struct {
	UserID   string
	UserName string
	Password string
	Token    string
}

type TestAdmin struct {
	UserName string
	Password string
	Token    string
}

type LoginAndRegisterData struct {
	UserID      string `json:"userID"`
	Token       string `json:"token"`
	ExpiredTime uint64 `json:"expiredTime"`
}

type LoginAndRegisterResponse struct {
	ErrCode int64                `json:"errCode"`
	ErrMsg  string               `json:"errMsg"`
	Data    LoginAndRegisterData `json:"data"`
}

type AdminLoginData struct {
	Token              string `json:"token"`
	GAuthEnabled       bool   `json:"gAuthEnabled"`
	GAuthSetupRequired bool   `json:"gAuthSetupRequired"`
	GAuthSetupProvUri  string `json:"gAuthSetupProvUri"`
}

type AdminLoginResponse struct {
	ErrCode int64          `json:"code"`
	ErrMsg  string         `json:"err_msg"`
	Data    AdminLoginData `json:"data"`
}

type UserIDResult struct {
	UserID string `json:"userID"`
	Result int32  `json:"result"`
}

type InviteUserToGroupResp struct {
	ErrCode          int32           `json:"errCode"`
	ErrMsg           string          `json:"errMsg"`
	UserIDResultList []*UserIDResult `json:"data"`
}

var AllRegisterUsers []*TestUser
var AllLoginUsers []*TestUser
var AllConnectedUsers []*TestUser

func GenUid(uid int, prefix string) string {
	//if getMyIP() == "" {
	//	log.Error("", "getMyIP() failed, exit ")
	//	os.Exit(1)
	//}
	UidPrefix := prefix + "_"
	return UidPrefix + strconv.FormatInt(int64(uid), 10)
}

func RegisterOnlineAccounts(number int, adminToken string) {
	var wg sync.WaitGroup
	wg.Add(number)
	for i := 0; i < number; i++ {
		go func(t int) {
			userID := GenUid(t, "online")
			register(userID, adminToken)
			log.Info("register ", userID)
			wg.Done()
		}(i)

	}
	wg.Wait()
	log.Info("", "RegisterAccounts finish ", number)
}

type GetTokenReq struct {
	Secret   string `json:"secret"`
	Platform int    `json:"platform"`
	Uid      string `json:"uid"`
}

type RegisterReq struct {
	Secret   string `json:"secret"`
	Platform int    `json:"platform"`
	Uid      string `json:"uid"`
	Name     string `json:"name"`
}

type ResToken struct {
	Data struct {
		ExpiredTime int64  `json:"expiredTime"`
		Token       string `json:"token"`
		Uid         string `json:"uid"`
	}
	ErrCode int    `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
}

func init() {
	//AdminToken = getToken("bytechat001")
	//AdminToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVSUQiOiJ4bGluazAwMSIsIlBsYXRmb3JtIjoiQWRtaW4iLCJPZmZpY2lhbElEIjowLCJleHAiOjE5OTM4MjQ2OTgsIm5iZiI6MTY3ODQ2NDY5OCwiaWF0IjoxNjc4NDY0Njk4fQ.lmwnnn_L2QSIIrlQz1lgUZDtvo-vpQmJm867PTYpufg"
}

func register(uid, adminToken string) error {
	//ACCOUNTCHECK
	var req server_api_params.AccountCheckReq
	req.OperationID = utils.OperationIDGenerator()
	req.CheckUserIDList = []string{uid}

	var getSelfUserInfoReq server_api_params.GetSelfUserInfoReq
	getSelfUserInfoReq.OperationID = req.OperationID
	getSelfUserInfoReq.UserID = uid

	var getSelfUserInfoResp server_api_params.AccountCheckResp

	//check account if exist
	r, err := network.Post2ApiWithoutAlives(ACCOUNTCHECK, req, adminToken)
	if err != nil {
		log2.Println(req.OperationID, "post failed, continue ", err.Error(), ACCOUNTCHECK, req, uid)
		log.Error(req.OperationID, "post failed, continue ", err.Error(), ACCOUNTCHECK, req, uid)
		return err
	}
	err = json.Unmarshal(r, &getSelfUserInfoResp)
	if err != nil {
		log2.Println(req.OperationID, "Unmarshal failed ", err.Error(), uid)
		log.Error(req.OperationID, "Unmarshal failed ", err.Error(), uid)
		return err
	}
	if getSelfUserInfoResp.ErrCode == 0 && len(getSelfUserInfoResp.ResultList) == 1 && getSelfUserInfoResp.ResultList[0].AccountStatus == "registered" {
		log2.Println(req.OperationID, "Already registered ", uid, getSelfUserInfoResp)
		log.Warn(req.OperationID, "Already registered ", uid, getSelfUserInfoResp)
		userLock.Lock()
		allUserID = append(allUserID, uid)
		userLock.Unlock()
		return nil
	} else if getSelfUserInfoResp.ErrCode == 0 && len(getSelfUserInfoResp.ResultList) == 1 && getSelfUserInfoResp.ResultList[0].AccountStatus == "unregistered" {
		log2.Println(req.OperationID, "not registered ", uid, getSelfUserInfoResp.ErrCode)
		log.Info(req.OperationID, "not registered ", uid, getSelfUserInfoResp.ErrCode)
	} else {
		log2.Println(req.OperationID, " failed, continue ", err, ACCOUNTCHECK, req, getSelfUserInfoResp, uid)
		log.Error(req.OperationID, " failed, continue ", err, ACCOUNTCHECK, req, getSelfUserInfoResp, uid)
		return errors.New(fmt.Sprintf("register failed, error code: %d, error message: %s", getSelfUserInfoResp.ErrCode, getSelfUserInfoResp.ErrMsg))
	}

	var rreq server_api_params.UserRegisterReq
	rreq.UserID = uid
	rreq.Secret = SECRET
	rreq.UserID = uid
	rreq.Platform = 1
	rreq.UserInfo.Nickname = uid
	rreq.OperationID = req.OperationID
	rreq.OperationID = req.OperationID
	_, err = network.Post2ApiWithoutAlives(REGISTERADDR, rreq, "")
	//if err != nil && !strings.Contains(err.Error(), "status code failed") {
	//	log.Error(req.OperationID, "post failed ,continue ", err.Error(), REGISTERADDR, req)
	//	time.Sleep(100 * time.Millisecond)
	//	continue
	//}
	if err != nil {
		log2.Println(req.OperationID, "post failed ,continue ", err.Error(), REGISTERADDR, req, uid)
		log.Error(req.OperationID, "post failed ,continue ", err.Error(), REGISTERADDR, req, uid)
		time.Sleep(100 * time.Millisecond)
		return err
	} else {
		log2.Println(req.OperationID, "register ok ", REGISTERADDR, req, uid)
		log.Info(req.OperationID, "register ok ", REGISTERADDR, req, uid)
		userLock.Lock()
		allUserID = append(allUserID, uid)
		userLock.Unlock()
		return nil
	}
}

func getToken(uid string) string {
	url := TOKENADDR
	req := api.UserTokenReq{}
	req.OperationID = utils.OperationIDGenerator()
	req.Platform = PlatformID
	req.Secret = SECRET
	req.UserID = uid
	req.GAuthTypeToken = false
	req.LoginIp = "127.0.0.1"

	//var req server_api_params.UserTokenReq
	//req.Platform = PlatformID
	//req.UserID = uid
	//req.Secret = SECRET
	//req.OperationID = utils.OperationIDGenerator()
	r, err := network.Post2ApiWithoutAlives(url, req, "a")
	if err != nil {
		log2.Println(req.OperationID, "Post2Api failed ", err.Error(), url, req)
		log.Error(req.OperationID, "Post2Api failed ", err.Error(), url, req)
		return ""
	}
	var stcResp ResToken
	err = json.Unmarshal(r, &stcResp)
	if stcResp.ErrCode != 0 {
		log2.Println(req.OperationID, "ErrCode failed ", stcResp.ErrCode, stcResp.ErrMsg, url, req)
		log.Error(req.OperationID, "ErrCode failed ", stcResp.ErrCode, stcResp.ErrMsg, url, req)
		return ""
	}
	log2.Println(req.OperationID, "get token: ", stcResp.Data.Token)
	log.Info(req.OperationID, "get token: ", stcResp.Data.Token)
	return stcResp.Data.Token
}

func RunGetToken(strMyUid string) string {
	token := getToken(strMyUid)
	//var token string
	//for true {
	//	token = getToken(strMyUid)
	//	if token == "" {
	//		time.Sleep(time.Duration(100) * time.Millisecond)
	//		continue
	//	} else {
	//		break
	//	}
	//}
	return token
}

func getMyIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error("", "InterfaceAddrs failed ", err.Error())
		os.Exit(1)
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func RegisterReliabilityUser(id int, timeStamp, adminToken string) {
	userID := GenUid(id, "reliability_"+timeStamp+"_")
	register(userID, adminToken)
	token := RunGetToken(userID)
	coreMgrLock.Lock()
	defer coreMgrLock.Unlock()
	allLoginMgr[id] = &CoreNode{token: token, userID: userID}
}

func WorkGroupRegisterReliabilityUser(id int) {
	userID := GenUid(id, "workgroup")
	//	register(userID)
	token := RunGetToken(userID)
	coreMgrLock.Lock()
	defer coreMgrLock.Unlock()
	log.Info("", "WorkGroupRegisterReliabilityUser userID: ", userID, "token: ", token)
	allLoginMgr[id] = &CoreNode{token: token, userID: userID}
}

func RegisterPressUser(id int, adminToken string) {
	userID := GenUid(id, PT.UserNamePrefix)
	log2.Println("start register user", userID)
	register(userID, adminToken)
	//token := RunGetToken(userID)
	//coreMgrLock.Lock()
	//defer coreMgrLock.Unlock()
	//if token != "" {
	//	PT.RegisterSuccessNum++
	//	allLoginMgr[id] = &CoreNode{token: token, userID: userID}
	//	//log2.Println("register user success", userID)
	//} else {
	//	//log2.Println("register user failed", userID)
	//}

}

//func GetGroupMemberNum(groupID string) uint32 {
//	var req server_api_params.GetGroupInfoReq
//	req.OperationID = utils.OperationIDGenerator()
//	req.GroupIDList = []string{groupID}
//
//	var groupInfoList []*server_api_params2.GroupInfo
//
//	r, err := network.Post2Api(GETGROUPSINFOROUTER, req, adminToken)
//	if err != nil {
//		log.Error("", "post failed ", GETGROUPSINFOROUTER, req)
//		return 0
//	}
//	err = common.CheckErrAndResp(nil, r, &groupInfoList)
//	if err != nil {
//		log.Error("", "CheckErrAndResp failed ", err.Error(), string(r))
//		return 0
//	}
//	log.Warn("", "group info", groupInfoList)
//	return groupInfoList[0].MemberCount
//}

func (pt PressureTest) RegisterFlow(username, password string) (*TestUser, error) {
	url := fmt.Sprintf("%s/demo/register", pt.ApiUrl)
	operationID := utils.OperationIDGenerator()
	params := map[string]interface{}{
		"operationID": operationID,
		"platform":    5,
		"userId":      username,
		"password":    password,
	}
	resp, err := http.Post(url, params, pt.ApiTimeOutSecond)
	if err != nil {
		return nil, err
	}

	var registerResp LoginAndRegisterResponse
	err = json.Unmarshal(resp, &registerResp)
	if err != nil {
		return nil, err
	}

	if registerResp.ErrCode != 0 {
		return nil, utils.Wrap(errors.New(registerResp.ErrMsg), fmt.Sprintf("%d", registerResp.ErrCode))
	}
	user := TestUser{
		UserID:   registerResp.Data.UserID,
		UserName: username,
		Password: password,
		Token:    registerResp.Data.Token,
	}

	AllRegisterUsers = append(AllRegisterUsers, &user)
	return &user, nil
}

func (pt PressureTest) UserLoginFlow(username, password string) (*TestUser, error) {
	operationID := utils.OperationIDGenerator()
	params := map[string]interface{}{
		"operationID": operationID,
		"platform":    5,
		"userId":      username,
		"password":    password,
	}
	resp, err := http.Post(TESTUSERLOGIN, params, pt.ApiTimeOutSecond)
	if err != nil {
		log2.Println("user login failed", TESTUSERLOGIN, err.Error())
		return nil, err
	}
	log2.Println("user login success", string(resp))

	var loginResp LoginAndRegisterResponse
	err = json.Unmarshal(resp, &loginResp)
	if err != nil {
		log2.Println("user login json.Unmarshal failed", TESTUSERLOGIN, err.Error())
		return nil, err
	}

	if loginResp.ErrCode != 0 {
		log2.Println("user login failed-2", loginResp.ErrCode, loginResp.ErrMsg)
		return nil, utils.Wrap(errors.New(loginResp.ErrMsg), fmt.Sprintf("%d", loginResp.ErrCode))
	}
	user := TestUser{
		UserID:   loginResp.Data.UserID,
		UserName: username,
		Password: password,
		Token:    loginResp.Data.Token,
	}

	AllLoginUsers = append(AllLoginUsers, &user)
	return &user, nil
}

func (pt PressureTest) InviteGroupMember(userIds []string) (*InviteUserToGroupResp, error) {
	operationID := utils.OperationIDGenerator()
	params := map[string]interface{}{
		"operationID":       operationID,
		"groupID":           pt.GroupID,
		"invitedUserIDList": userIds,
		"reason":            "pressure testing",
	}
	resp, err := http.PostWithToken(TESTINVITEGROUPMEMBER, params, pt.ApiTimeOutSecond, pt.GroupOwnerToken)
	if err != nil {
		return nil, err
	}

	var inviteResp InviteUserToGroupResp
	err = json.Unmarshal(resp, &inviteResp)
	if err != nil {
		return nil, err
	}

	if inviteResp.ErrCode != 0 {
		return nil, utils.Wrap(errors.New(inviteResp.ErrMsg), fmt.Sprintf("%d", inviteResp.ErrCode))
	}

	return &inviteResp, nil
}

func (pt PressureTest) AdminLoginFlow() (*TestAdmin, error) {
	username := pt.AdminAccount
	password := pt.AdminPassword
	params := map[string]interface{}{
		"admin_name": username,
		"secret":     password,
	}
	resp, err := http.Post(ADMINLOGIN, params, pt.ApiTimeOutSecond)
	if err != nil {
		log2.Println("admin login failed", ADMINLOGIN, err.Error())
		return nil, err
	}
	log2.Println("admin login success", string(resp))

	var loginResp AdminLoginResponse
	err = json.Unmarshal(resp, &loginResp)
	if err != nil {
		log2.Println("admin login json.Unmarshal failed", TESTUSERLOGIN, err.Error())
		return nil, err
	}

	if loginResp.ErrCode != 0 {
		log2.Println("admin login failed-2", loginResp.ErrCode, loginResp.ErrMsg)
		return nil, utils.Wrap(errors.New(loginResp.ErrMsg), fmt.Sprintf("%d", loginResp.ErrCode))
	}

	if loginResp.Data.GAuthEnabled || loginResp.Data.GAuthSetupRequired {
		log2.Println("admin login failed-3, closed the google auth first!")
		return nil, errors.New("admin login failed-3, closed the google auth first")
	}

	user := TestAdmin{
		UserName: username,
		Password: password,
		Token:    loginResp.Data.Token,
	}
	return &user, nil
}
