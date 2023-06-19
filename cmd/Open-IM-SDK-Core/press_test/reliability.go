package main

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"fmt"
	"io/ioutil"
	log2 "log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	TESTIP   string
	HTTPADDR string
	APIADDR  string

	WSADDR                string
	REGISTERADDR          string
	TOKENADDR             string
	TESTUSERLOGIN         string
	SECRET                string
	GETSELFUSERINFO       string
	CREATEGROUP           string
	TESTINVITEGROUPMEMBER string
	ACCOUNTCHECK          string
	GETGROUPSINFOROUTER   string
	ADMINLOGIN            string
)

func GetFileContentAsStringLines(filePath string) ([]string, error) {
	result := []string{}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return result, err
	}
	s := string(b)
	for _, lineStr := range strings.Split(s, "\n") {
		lineStr = strings.TrimSpace(lineStr)
		if lineStr == "" {
			continue
		}
		result = append(result, lineStr)
	}
	return result, nil
}

func GetCmd(myUid int, filename string) int {
	cmd, err := GetFileContentAsStringLines("cmd.txt")
	if err != nil {
		fmt.Println("GetFileContentAsStringLines failed")
		return -1
	}
	if len(cmd) < myUid {
		fmt.Println("len failed")
		return -1
	}
	return int(utils.StringToInt64(cmd[myUid-1]))
}

//func ReliabilityTest(msgNumOneClient int, intervalSleepMS int, randSleepMaxSecond int, clientNum int) {
//	msgNumInOneClient = msgNumOneClient
//	timeStamp := utils.Int64ToString(time.Now().Unix())
//
//	var wg sync.WaitGroup
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		go func(idx int) {
//			RegisterReliabilityUser(idx, timeStamp)
//			wg.Done()
//		}(i)
//	}
//	wg.Wait()
//
//	log.Warn("", "RegisterReliabilityUser finished, clientNum: ", clientNum)
//	log.Warn("", " init, login, send msg, start ")
//	rand.Seed(time.Now().UnixNano())
//
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		rdSleep := rand.Intn(randSleepMaxSecond) + 1
//		isSend := 0
//		if isSend == 0 {
//			go func(idx int) {
//				log.Warn("", " send msg flag true ", idx)
//				ReliabilityOne(idx, rdSleep, true, intervalSleepMS)
//				wg.Done()
//			}(i)
//			sendMsgClient++
//		} else {
//			go func(idx int) {
//				log.Warn("", " send msg flag false ", idx)
//				ReliabilityOne(idx, rdSleep, false, intervalSleepMS)
//				wg.Done()
//			}(i)
//		}
//	}
//	wg.Wait()
//	log.Warn("send msg finish,  CheckReliabilityResult")
//
//	for {
//		if CheckReliabilityResult(msgNumOneClient, clientNum) {
//			log.Warn("", "CheckReliabilityResult ok, exit")
//			os.Exit(0)
//			return
//		} else {
//			log.Warn("", "CheckReliabilityResult failed , wait.... ")
//		}
//		time.Sleep(time.Duration(5) * time.Second)
//	}
//}

//func WorkGroupReliabilityTest(msgNumOneClient int, intervalSleepMS int, randSleepMaxSecond int, clientNum int, groupID string) {
//	msgNumInOneClient = msgNumOneClient
//	//timeStamp := utils.Int64ToString(time.Now().Unix())
//
//	var wg sync.WaitGroup
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		go func(idx int) {
//			WorkGroupRegisterReliabilityUser(idx)
//			wg.Done()
//		}(i)
//	}
//	wg.Wait()
//
//	log.Warn("", "RegisterReliabilityUser finished, clientNum: ", clientNum)
//	log.Warn("", " init, login, send msg, start ")
//	rand.Seed(time.Now().UnixNano())
//
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		rdSleep := rand.Intn(randSleepMaxSecond) + 1
//		isSend := 0
//		if isSend == 0 {
//			go func(idx int) {
//				log.Warn("", " send msg flag true ", idx)
//				WorkGroupReliabilityOne(idx, rdSleep, true, intervalSleepMS, groupID)
//				wg.Done()
//			}(i)
//			sendMsgClient++
//		} else {
//			go func(idx int) {
//				log.Warn("", " send msg flag false ", idx)
//				ReliabilityOne(idx, rdSleep, false, intervalSleepMS)
//				wg.Done()
//			}(i)
//		}
//	}
//	wg.Wait()
//	log.Warn("send msg finish,  CheckReliabilityResult")
//
//	for {
//		if CheckReliabilityResult(msgNumOneClient, clientNum) {
//			log.Warn("", "CheckReliabilityResult ok, exit")
//			os.Exit(0)
//			return
//		} else {
//			log.Warn("", "CheckReliabilityResult failed , wait.... ")
//		}
//		time.Sleep(time.Duration(5) * time.Second)
//	}
//}

//func WorkGroupMsgDelayTest(msgNumOneClient int, intervalSleepMS int, randSleepMaxSecond int, clientBegin int, clientEnd int, groupID string) {
//	msgNumInOneClient = msgNumOneClient
//
//	var wg sync.WaitGroup
//
//	wg.Add(clientEnd - clientBegin + 1)
//	for i := clientBegin; i <= clientEnd; i++ {
//		go func(idx int) {
//			WorkGroupRegisterReliabilityUser(idx)
//			wg.Done()
//		}(i)
//	}
//	wg.Wait()
//
//	log.Warn("", "RegisterReliabilityUser finished, client: ", clientBegin, clientEnd)
//	log.Warn("", " init, login, send msg, start ")
//	rand.Seed(time.Now().UnixNano())
//
//	wg.Add(clientEnd - clientBegin + 1)
//	for i := clientBegin; i <= clientEnd; i++ {
//		rdSleep := rand.Intn(randSleepMaxSecond) + 1
//		isSend := 0
//		if isSend == 0 {
//			go func(idx int) {
//				log.Warn("", " send msg flag true ", idx)
//				WorkGroupReliabilityOne(idx, rdSleep, true, intervalSleepMS, groupID)
//				wg.Done()
//			}(i)
//			sendMsgClient++
//		} else {
//			go func(idx int) {
//				log.Warn("", " send msg flag false ", idx)
//				WorkGroupReliabilityOne(idx, rdSleep, false, intervalSleepMS, groupID)
//				wg.Done()
//			}(i)
//		}
//	}
//	wg.Wait()
//	log.Warn("send msg finish,  CheckReliabilityResult")
//
//	for {
//		if CheckReliabilityResult(msgNumOneClient, clientEnd-clientBegin+1) {
//			log.Warn("", "CheckReliabilityResult ok, exit")
//			os.Exit(0)
//			return
//		} else {
//			log.Warn("", "CheckReliabilityResult failed , wait.... ")
//		}
//		time.Sleep(time.Duration(5) * time.Second)
//	}
//}

func PressTest(test *PressureTest) {
	log2.Println("start pressure testing")

	TESTIP = test.ApiUrl
	httpPrefix := "http://"
	wsPrefix := "ws://"
	if test.ISHTTPS {
		httpPrefix = "https://"
		wsPrefix = "wss://"
	}
	HTTPADDR = httpPrefix + TESTIP
	APIADDR = HTTPADDR + "/api"

	WSADDR = wsPrefix + TESTIP + "/msg_gateway"
	REGISTERADDR = APIADDR + "/auth/user_register"
	TOKENADDR = APIADDR + "/auth/user_token"
	TESTUSERLOGIN = HTTPADDR + "/demo/login"
	ADMINLOGIN = HTTPADDR + "/cms/admin/admin_login"
	SECRET = "a123456"
	GETSELFUSERINFO = APIADDR + "/user/get_self_user_info"
	CREATEGROUP = APIADDR + constant.CreateGroupRouter
	TESTINVITEGROUPMEMBER = APIADDR + "/group/invite_user_to_group"
	ACCOUNTCHECK = APIADDR + "/manager/account_check"
	GETGROUPSINFOROUTER = APIADDR + constant.GetGroupsInfoRouter

	msgNumOneClient := test.MessageNum
	intervalSleepMS := test.MessageIntervalTimeMilli
	clientNum := int(test.UserNum)
	groupID := test.GroupID
	groupOwnerID := test.GroupOwnerAccount
	groupOwnerPassword := test.GroupOwnerPassword

	msgNumInOneClient = int(msgNumOneClient)
	//timeStamp := utils.Int64ToString(time.Now().Unix())

	t1 := time.Now()
	var wg sync.WaitGroup

	//login admin
	if adminToken != "" {
		test.AdminToken = adminToken
	} else {
		admin, err := test.AdminLoginFlow()
		if err != nil {
			log2.Println("admin login failed", err.Error())
			log.Info("", "admin login failed", err.Error())
			return
		}
		test.AdminToken = admin.Token
	}

	//register users
	if step == StepRegister || step == StepRegisterAndLogin || step == StepSendMessage {
		wg.Add(clientNum)
		for i := 0; i < clientNum; i++ {
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			go func(idx int) {
				RegisterPressUser(idx, test.AdminToken)
				log2.Println("get user token finish ", idx)
				log.Info("", "get user token finish ", idx)
				wg.Done()
			}(i)
		}
		wg.Wait()
	}

	//get token from registered users
	if step == StepRegisterAndLogin || step == StepSendMessage {
		log2.Println("get token from registered users")
		wg.Add(clientNum)
		for i := 0; i < clientNum; i++ {
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			go func(idx int) {
				defer wg.Done()
				userID := GenUid(idx, PT.UserNamePrefix)
				for {
					token := RunGetToken(userID)
					if token != "" {
						coreMgrLock.Lock()
						defer coreMgrLock.Unlock()
						PT.RegisterSuccessNum++
						allLoginMgr[idx] = &CoreNode{token: token, userID: userID}
						break

					} else {
						time.Sleep(1 * time.Second)
						continue
					}
				}

			}(i)
		}
		wg.Wait()
		log2.Println("", "get all user token finish ", "need register user num", clientNum, "register success num", test.RegisterSuccessNum, " cost time: ", time.Since(t1))
		log.Warn("", "get all user token finish ", "need register user num", clientNum, "register success num", test.RegisterSuccessNum, " cost time: ", time.Since(t1))

		//login users
		log2.Println("", "init and login begin ")
		log.Warn("", "init and login begin ")
		t1 = time.Now()
		wg.Add(clientNum)
		for i := 0; i < clientNum; i++ {
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			go func(idx int) {
				strMyUid := allLoginMgr[idx].userID
				token := allLoginMgr[idx].token
				PressInitAndLogin(idx, strMyUid, token, WSADDR, APIADDR, test.DataDir)
				wg.Done()
			}(i)
		}
		wg.Wait()
		log2.Println("", "init and login end ", " cost time: ", time.Since(t1))
		log.Warn("", "init and login end ", " cost time: ", time.Since(t1))
	}

	if step == StepSendMessage {
		//start adding group members
		log2.Println("start adding group member")
		//login group owner
		log2.Println("login group owner")
		user, err := test.UserLoginFlow(groupOwnerID, groupOwnerPassword)
		if err != nil {
			log2.Println("group owner login failed", err.Error())
			return
		} else {
			log2.Println("group owner login success", user.UserID, user.UserName, user.Token)
		}
		test.GroupOwnerToken = user.Token

		//invite group members
		var times int
		onceNum := 300
		if clientNum > onceNum {
			times = int(math.Abs(float64(clientNum) / float64(onceNum)))
			modNum := math.Mod(float64(clientNum), float64(onceNum))
			if modNum > 0 {
				times++
			}
		} else {
			times = 1
		}

		log2.Println("add group members times", times)
		var wg2 sync.WaitGroup
		wg2.Add(times)
		for i := 0; i < times; i++ {
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			go func(idx int) {
				log2.Println("running", idx)
				defer wg2.Done()
				var startNum, endNum int
				if idx == 0 {
					startNum = 0
					endNum = (idx + 1) * onceNum
				} else {
					startNum = idx * onceNum
					endNum = (idx + 1) * onceNum
				}
				var inviteMemberList []string
				for j := startNum; j < endNum; j++ {
					if j < clientNum {
						user := allLoginMgr[j]
						inviteMemberList = append(inviteMemberList, user.userID)
					}
				}
				log2.Println("inviteMemberList", inviteMemberList)

				resp, err := test.InviteGroupMember(inviteMemberList)
				if err != nil {
					log2.Println("invite group members failed", err.Error())
					return
				} else {
					log2.Println("invite group members finish", resp.ErrCode, resp.ErrMsg)
				}
			}(i)
		}
		wg2.Wait()
		log2.Println("Invite group members finish")

		//start sending messages to the group
		log2.Println("", "send msg begin ")
		log.Warn("", "send msg begin ")
		t1 = time.Now()
		wg.Add(clientNum)
		for i := 0; i < clientNum; i++ {
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			go func(idx int) {
				PressGroupMsg(idx, 0, true, intervalSleepMS, groupID)
				log2.Println("", "press finished  ", idx)
				log.Warn("", "press finished  ", idx)
				wg.Done()
			}(i)
		}
		wg.Wait()
		//sendMsgTotalSuccessNum := uint32(0)
		//sendMsgTotalFailedNum := uint32(0)
		//for _, v := range allLoginMgr {
		//	sendMsgTotalSuccessNum += v.sendMsgSuccessNum
		//	sendMsgTotalFailedNum += v.sendMsgFailedNum
		//}
		//time.Sleep(time.Second * time.Duration(10))
		log2.Println("send msg end  ", "number of messages expected to be sent: ", clientNum*msgNumInOneClient, " sendMsgTotalSuccessNum: ", PT.SendSuccMsgNum, " sendMsgTotalFailedNum: ", PT.SendFailedMsgNum, "cost time: ", time.Since(t1))
		log.Warn("send msg end  ", "number of messages expected to be sent: ", clientNum*msgNumInOneClient, " sendMsgTotalSuccessNum: ", PT.SendSuccMsgNum, " sendMsgTotalFailedNum: ", PT.SendFailedMsgNum, "cost time: ", time.Since(t1))
	}

}

//func WorkGroupPressTest(msgNumOneClient int, intervalSleepMS int, clientNum int, groupID string) {
//	msgNumInOneClient = msgNumOneClient
//	t1 := time.Now()
//	var wg sync.WaitGroup
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		go func(idx int) {
//			WorkGroupRegisterReliabilityUser(idx)
//			log.Info("", "get user token finish ", idx)
//			wg.Done()
//		}(i)
//	}
//	wg.Wait()
//	log.Warn("", "get all user token finish ", clientNum, " cost time: ", time.Since(t1))
//
//	log.Warn("", "init and login begin ")
//	t1 = time.Now()
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		go func(idx int) {
//			strMyUid := allLoginMgr[idx].userID
//			token := allLoginMgr[idx].token
//			ReliabilityInitAndLogin(idx, strMyUid, token, WSADDR, APIADDR)
//			wg.Done()
//		}(i)
//	}
//	wg.Wait()
//	log.Warn("", "init and login end ", " cost time: ", time.Since(t1))
//
//	log.Warn("", "send msg begin ")
//	t1 = time.Now()
//	wg.Add(clientNum)
//	for i := 0; i < clientNum; i++ {
//		go func(idx int) {
//			WorkGroupPressOne(idx, 0, true, intervalSleepMS, groupID)
//			wg.Done()
//		}(i)
//	}
//	wg.Wait()
//	sendMsgTotalSuccessNum := uint32(0)
//	sendMsgTotalFailedNum := uint32(0)
//	for _, v := range allLoginMgr {
//		sendMsgTotalSuccessNum += v.sendMsgSuccessNum
//		sendMsgTotalFailedNum += v.sendMsgFailedNum
//	}
//	log.Warn("send msg end  ", "number of messages expected to be sent: ", clientNum*msgNumOneClient, " sendMsgTotalSuccessNum: ", sendMsgTotalSuccessNum, " sendMsgTotalFailedNum: ", sendMsgTotalFailedNum, "cost time: ", time.Since(t1))
//}

//func CheckReliabilityResult(msgNumOneClient int, clientNum int) bool {
//	log.Info("", "start check map send -> map recv")
//	sameNum := 0
//
//	if len(SendSuccAllMsg)+len(SendFailedAllMsg) != msgNumOneClient*clientNum {
//		log.Warn("", utils.GetSelfFuncName(), " send msg succ number: ", len(SendSuccAllMsg), " send  msg failed number: ", len(SendFailedAllMsg), " all: ", msgNumOneClient*clientNum)
//		return false
//	}
//
//	for ksend, _ := range SendSuccAllMsg {
//		_, ok := RecvAllMsg[ksend]
//		if ok {
//			sameNum++
//			//x := vsend
//			//y := krecv
//			//x = x + x
//			//y = y + y
//
//		} else {
//			log.Error("", "check failed  not in recv ", ksend)
//			log.Error("", "send failed num: ", len(SendFailedAllMsg), " send success num: ", len(SendSuccAllMsg), " recv num: ", len(RecvAllMsg))
//			return false
//		}
//	}
//	log.Info("", "check map send -> map recv ok ", sameNum)
//	//log.Info("", "start check map recv -> map send ")
//	//sameNum = 0
//
//	//for k1, _ := range RecvAllMsg {
//	//	_, ok := SendSuccAllMsg[k1]
//	//	if ok {
//	//		sameNum++
//	//		//x := v1 + v2
//	//		//x = x + x
//	//
//	//	} else {
//	//		log.Error("", "check failed  not in send ", k1, len(SendFailedAllMsg), len(SendSuccAllMsg), len(RecvAllMsg))
//	//		//	return false
//	//	}
//	//}
//	maxCostMsgID := ""
//	minCostTime := int64(1000000)
//	maxCostTime := int64(0)
//	totalCostTime := int64(0)
//	for ksend, vsend := range SendSuccAllMsg {
//		krecv, ok := RecvAllMsg[ksend]
//		if ok {
//			sameNum++
//			costTime := krecv.RecvTime - vsend.SendTime
//			totalCostTime += costTime
//			if costTime > maxCostTime {
//				maxCostMsgID = ksend
//				maxCostTime = costTime
//			}
//			if minCostTime > costTime {
//				minCostTime = costTime
//			}
//		}
//	}
//
//	log.Warn("", "need send msg num : ", sendMsgClient*msgNumInOneClient)
//	log.Warn("", "send msg succ num ", len(SendSuccAllMsg))
//	log.Warn("", "send msg failed num ", len(SendFailedAllMsg))
//	log.Warn("", "recv msg succ num ", len(RecvAllMsg))
//	log.Warn("", "minCostTime: ", minCostTime, "ms, maxCostTime: ", maxCostTime, "ms, average cost time: ", totalCostTime/(int64(sendMsgClient*msgNumInOneClient)), "ms", " maxCostMsgID: ", maxCostMsgID)
//
//	return true
//}

func ReliabilityOne(index int, beforeLoginSleep int, isSendMsg bool, intervalSleepMS int) {
	//	time.Sleep(time.Duration(beforeLoginSleep) * time.Second)
	strMyUid := allLoginMgr[index].userID
	token := allLoginMgr[index].token
	ReliabilityInitAndLogin(index, strMyUid, token, WSADDR, APIADDR)
	log.Info("", "login ok client num: ", len(allLoginMgr))
	log.Warn("start One", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
	msgnum := msgNumInOneClient
	uidNum := len(allLoginMgr)
	var recvId string
	var idx string
	rand.Seed(time.Now().UnixNano())
	if msgnum == 0 {
		os.Exit(0)
	}
	if !isSendMsg {
		//	Msgwg.Done()
	} else {
		for i := 0; i < msgnum; i++ {
			var r int
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			for {
				r = rand.Intn(uidNum)
				if r == index {
					continue
				} else {

					break
				}

			}

			recvId = allLoginMgr[r].userID

			idx = strconv.FormatInt(int64(i), 10)
			for {
				if runtime.NumGoroutine() > MaxNumGoroutine {
					time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
					log.Warn("", "NumGoroutine > max  ", runtime.NumGoroutine(), MaxNumGoroutine)
					continue
				} else {
					break
				}
			}

			DoTestSendMsg(index, strMyUid, recvId, "", idx)

		}
		//Msgwg.Done()
	}
}

func WorkGroupReliabilityOne(index int, beforeLoginSleep int, isSendMsg bool, intervalSleepMS int, groupID string) {
	//	time.Sleep(time.Duration(beforeLoginSleep) * time.Second)
	strMyUid := allLoginMgr[index].userID
	token := allLoginMgr[index].token
	ReliabilityInitAndLogin(index, strMyUid, token, WSADDR, APIADDR)
	log.Info("", "login ok client num: ", len(allLoginMgr))
	log.Warn("start One", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
	msgnum := msgNumInOneClient
	uidNum := len(allLoginMgr)
	var idx string
	rand.Seed(time.Now().UnixNano())
	if msgnum == 0 {
		os.Exit(0)
	}
	if !isSendMsg {
		//	Msgwg.Done()
	} else {
		for i := 0; i < msgnum; i++ {
			var r int
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			for {
				r = rand.Intn(uidNum)
				if r == index {
					continue
				} else {

					break
				}

			}

			idx = strconv.FormatInt(int64(i), 10)
			for {
				if runtime.NumGoroutine() > MaxNumGoroutine {
					time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
					log.Warn("", "NumGoroutine > max  ", runtime.NumGoroutine(), MaxNumGoroutine)
					continue
				} else {
					break
				}
			}

			DoTestSendMsg(index, strMyUid, "", groupID, idx)

		}
		//Msgwg.Done()
	}
}

func WorkGroupMsgDelayOne(index int, beforeLoginSleep int, isSendMsg bool, intervalSleepMS int, groupID string) {
	//	time.Sleep(time.Duration(beforeLoginSleep) * time.Second)
	strMyUid := allLoginMgr[index].userID
	token := allLoginMgr[index].token
	ReliabilityInitAndLogin(index, strMyUid, token, WSADDR, APIADDR)
	log.Info("", "login ok client num: ", len(allLoginMgr))
	log.Warn("start One", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
	msgnum := msgNumInOneClient
	uidNum := len(allLoginMgr)
	var idx string
	rand.Seed(time.Now().UnixNano())
	if msgnum == 0 {
		os.Exit(0)
	}
	if !isSendMsg {
		//	Msgwg.Done()
	} else {
		for i := 0; i < msgnum; i++ {
			var r int
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			for {
				r = rand.Intn(uidNum)
				if r == index {
					continue
				} else {

					break
				}

			}

			idx = strconv.FormatInt(int64(i), 10)
			for {
				if runtime.NumGoroutine() > MaxNumGoroutine {
					time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
					log.Warn("", "NumGoroutine > max  ", runtime.NumGoroutine(), MaxNumGoroutine)
					continue
				} else {
					break
				}
			}

			DoTestSendMsg(index, strMyUid, "", groupID, idx)

		}
		//Msgwg.Done()
	}
}

//
//func WorkGroupMsgDelayOne(sendID1 string, beforeLoginSleep int, isSendMsg bool, intervalSleepMS int, groupID string) {
//	//	time.Sleep(time.Duration(beforeLoginSleep) * time.Second)
//	strMyUid := allLoginMgr[index].userID
//	token := allLoginMgr[index].token
//	ReliabilityInitAndLogin(index, strMyUid, token, WSADDR, APIADDR)
//	log.Info("", "login ok client num: ", len(allLoginMgr))
//	log.Warn("start One", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
//	msgnum := msgNumInOneClient
//	uidNum := len(allLoginMgr)
//	var idx string
//	rand.Seed(time.Now().UnixNano())
//	if msgnum == 0 {
//		os.Exit(0)
//	}
//	if !isSendMsg {
//		//	Msgwg.Done()
//	} else {
//		for i := 0; i < msgnum; i++ {
//			var r int
//			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
//			for {
//				r = rand.Intn(uidNum)
//				if r == index {
//					continue
//				} else {
//
//					break
//				}
//
//			}
//
//			idx = strconv.FormatInt(int64(i), 10)
//			for {
//				if runtime.NumGoroutine() > MaxNumGoroutine {
//					time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
//					log.Warn("", "NumGoroutine > max  ", runtime.NumGoroutine(), MaxNumGoroutine)
//					continue
//				} else {
//					break
//				}
//			}
//
//			DoTestSendMsg(index, strMyUid, "", groupID, idx)
//
//		}
//		//Msgwg.Done()
//	}
//}

func PressGroupMsg(index int, beforeLoginSleep int, isSendMsg bool, intervalSleepMS int, groupId string) {
	if beforeLoginSleep != 0 {
		time.Sleep(time.Duration(beforeLoginSleep) * time.Millisecond)
	}

	strMyUid := allLoginMgr[index].userID
	token := allLoginMgr[index].token
	//	ReliabilityInitAndLogin(index, strMyUid, token, WSADDR, APIADDR)
	log2.Println("", "login ok client num: ", len(allLoginMgr))
	log2.Println("start One", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
	log.Info("", "login ok client num: ", len(allLoginMgr))
	log.Info("start One", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
	msgnum := msgNumInOneClient
	var idx string
	rand.Seed(time.Now().UnixNano())
	if msgnum == 0 {
		log2.Println("testing msg num", msgnum)
		os.Exit(0)
	}
	if !isSendMsg {
		//	Msgwg.Done()
	} else {
		for i := 0; i < msgnum; i++ {
			log2.Println("start testing sending message:", i)
			idx = strconv.FormatInt(int64(i), 10)
			for {
				if runtime.NumGoroutine() > MaxNumGoroutine {
					time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
					log2.Println("", " NumGoroutine > max ", runtime.NumGoroutine(), MaxNumGoroutine)
					log.Warn("", " NumGoroutine > max ", runtime.NumGoroutine(), MaxNumGoroutine)
					continue
				} else {
					break
				}
			}
			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
			//DoTestSendMsg(index, strMyUid, recvId, idx)
			sc := sendPressMsg(index, strMyUid, "", groupId, idx, i)
			log2.Println("end testing sending message:", i, sc.result)
		}
		//Msgwg.Done()
	}
}

//func WorkGroupPressOne(index int, beforeLoginSleep int, isSendMsg bool, intervalSleepMS int, groupID string) {
//	if beforeLoginSleep != 0 {
//		time.Sleep(time.Duration(beforeLoginSleep) * time.Millisecond)
//	}
//	strMyUid := allLoginMgr[index].userID
//	token := allLoginMgr[index].token
//	//ReliabilityInitAndLogin(index, strMyUid, token, WSADDR, APIADDR)
//	log.Info("", "login ok, client num: ", len(allLoginMgr))
//	log.Info("start One ", index, beforeLoginSleep, isSendMsg, strMyUid, token, WSADDR, APIADDR)
//	msgnum := msgNumInOneClient
//	var idx string
//	rand.Seed(time.Now().UnixNano())
//	if msgnum == 0 {
//		os.Exit(0)
//	}
//	if !isSendMsg {
//	} else {
//		for i := 0; i < msgnum; i++ {
//			idx = strconv.FormatInt(int64(i), 10)
//
//			for {
//				if runtime.NumGoroutine() > MaxNumGoroutine {
//					time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
//					log.Warn("", " NumGoroutine > max ", runtime.NumGoroutine(), MaxNumGoroutine)
//					continue
//				} else {
//					break
//				}
//			}
//			log.Info("sendPressMsg begin", index, strMyUid, groupID)
//			if sendPressMsg(index, strMyUid, "", groupID, idx, i) {
//				allLoginMgr[index].sendMsgSuccessNum++
//			} else {
//				allLoginMgr[index].sendMsgFailedNum++
//			}
//			log.Info("sendPressMsg end")
//			time.Sleep(time.Duration(intervalSleepMS) * time.Millisecond)
//		}
//	}
//}
