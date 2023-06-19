package main

import (
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"flag"
	"fmt"
	log2 "log"
	"runtime"
	"syscall"
)

var PT PressureTest
var step int64
var adminAccount string
var adminPassword string
var adminToken string
var apiUrl string
var dataDir string
var isHttps bool
var timeOutSecond int64
var groupID string
var groupOwnerAccount string
var groupOwnerPassword string
var testUserNum int64
var userNamePrefix string
var messageNum int64
var intervalMilliSecond int64
var cpuNum int

func main() {
	flag.Int64Var(&step, "step", 2, "0 register, 1 login, 2 send message")
	flag.StringVar(&adminAccount, "adminAccount", "xlink001", "admin login account")
	flag.StringVar(&adminPassword, "adminPassword", "a004poi@!", "admin login password")
	flag.StringVar(&adminToken, "adminToken", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVSUQiOiJ4bGluazAwMSIsIlBsYXRmb3JtIjoiQWRtaW4iLCJPZmZpY2lhbElEIjowLCJleHAiOjE5OTQ3NjM4MjMsIm5iZiI6MTY3OTQwMzgyMywiaWF0IjoxNjc5NDAzODIzfQ.LndkmDu-L0mVobKr4BBJyPz4J9Ji9jnMFNffZZJbm0A", "admin token")
	flag.StringVar(&apiUrl, "apiUrl", "web.xlink-test.com", "api domain")
	flag.StringVar(&dataDir, "dataDir", "./test_data", "database directory path")
	flag.BoolVar(&isHttps, "isHttps", true, "http: false, https: true")
	flag.Int64Var(&timeOutSecond, "timeOutSecond", 120, "api request timeout second, 0 means no timeout")
	flag.StringVar(&groupID, "groupID", "presstest1", "group id")
	flag.StringVar(&groupOwnerAccount, "ownerID", "anson1", "the userID of the owner of the specific group")
	flag.StringVar(&groupOwnerPassword, "ownerPassword", "a123456", "the password of the owner of the specific group")
	flag.Int64Var(&testUserNum, "userNum", 5, "number of the test users")
	flag.StringVar(&userNamePrefix, "userNamePrefix", "pt", "prefix for the user name of the test users")
	flag.Int64Var(&messageNum, "messageNum", 5, "number of the messages each user will send")
	flag.Int64Var(&intervalMilliSecond, "interval", 1000, "interval milli second for each message")
	flag.IntVar(&cpuNum, "cpuNum", 1, "number of cpu you will use")
	flag.Parse()

	runtime.GOMAXPROCS(cpuNum)
	fmt.Printf("示例: ./press_test -step 2 -apiUrl web.xlink-test.com -dataDir ./test_data -isHttps=false -timeOutSecond 2000 -adminAccount xlink001 -adminPassword a004poi@! -groupID presstest1 -ownerID anson1 -ownerPassword a123456 -userNum 1 -userNamePrefix pt -messageNum 3 -interval 1000 -cpuNum 1  \n")
	fmt.Printf("当前压测地址：%s \n", apiUrl)
	fmt.Printf("当前使用admin账号：%s \n", adminAccount)
	fmt.Printf("当前请求参数: -step %d -dataDir %s -isHttps %v -timeOutSecond %d -groupID %s -ownerID %s -ownerPassword %s -userNum %d -userNamePrefix %s -messageNum %d -interval %d -cpuNum %d \n", step, dataDir, isHttps, timeOutSecond, groupID, groupOwnerAccount, groupOwnerPassword, testUserNum, userNamePrefix, messageNum, intervalMilliSecond, cpuNum)

	var rlim syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		log2.Println("syscall.Getrlimit failed", err.Error())
		log.NewError("", "syscall.Getrlimit failed", err.Error())
		return
	}
	log2.Println("syscall.Getrlimit result", rlim.Cur, rlim.Max)
	log.NewInfo("", "syscall.Getrlimit result", rlim.Cur, rlim.Max)

	rlim.Cur = 50000
	rlim.Max = 50000
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		log2.Println("syscall.Setrlimit failed", err.Error())
		log.NewError("", "syscall.Setrlimit failed", err.Error())
		return
	}

	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		log2.Println("After updated syscall.Getrlimit failed", err.Error())
		log.NewError("", "After updated syscall.Getrlimit failed", err.Error())
		return
	}
	log2.Println("After updated rlimit result", rlim.Cur, rlim.Max)
	log.NewInfo("", "After updated rlimit result", rlim.Cur, rlim.Max)

	pressTest()

}

func pressTest() {
	PT = PressureTest{
		OperationID:              utils.OperationIDGenerator(),
		Step:                     step,
		ApiUrl:                   apiUrl,
		DataDir:                  dataDir,
		ISHTTPS:                  isHttps,
		ApiTimeOutSecond:         int(timeOutSecond),
		GroupID:                  groupID,
		GroupOwnerAccount:        groupOwnerAccount,
		GroupOwnerPassword:       groupOwnerPassword,
		AdminAccount:             adminAccount,
		AdminPassword:            adminPassword,
		UserNum:                  testUserNum,
		UserNamePrefix:           userNamePrefix,
		MessageNum:               messageNum,
		MessageIntervalTimeMilli: int(intervalMilliSecond),
		SendSuccMsgNum:           0,
		SendFailedMsgNum:         0,
		RecvMsgNum:               0,
	}

	flag.Parse()
	constant.OnlyForTest = 1
	log.NewPrivateLog(constant.PressureTestLogFileName)
	log.Warn("", "press test begin, sender num: ", PT.UserNum, " single sender msg num: ", PT.MessageNum, " send msg total num: ", PT.UserNum*PT.MessageNum)
	PressTest(&PT)
	select {}
}
