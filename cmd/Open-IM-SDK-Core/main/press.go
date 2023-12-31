package main

import (
	"Open_IM/cmd/Open-IM-SDK-Core/test"
	"Open_IM/pkg/common/log"
	"flag"
	"os"
)

func main() {
	var onlineNum *int          //Number of online users
	var senderNum *int          //Number of users sending messages
	var singleSenderMsgNum *int //Number of single user send messages
	var intervalTime *int       //Sending time interval, in millisecond
	onlineNum = flag.Int("on", 10, "online num")
	senderNum = flag.Int("sn", 10, "sender num")
	singleSenderMsgNum = flag.Int("mn", 1000, "single sender msg num")
	intervalTime = flag.Int("t", 1, "interval time mill second")
	flag.Parse()

	if *onlineNum < *senderNum {
		log.Error("", "args failed onlineNum < senderNum ", *onlineNum, *senderNum)
		os.Exit(-1)
	}
	log.NewPrivateLog("press.log")
	log.NewWarn("", "online test start, number of online users: ", *onlineNum)
	//test.OnlineTest(*onlineNum)
	log.NewWarn("", "online test finish, number of online users: ", *onlineNum)
	log.NewWarn("", "reliability test start, user: ", *senderNum, "message number: ", *singleSenderMsgNum)
	test.ReliabilityTest(*singleSenderMsgNum, *intervalTime, 10, *senderNum)
	//	test.PressTest(*singleSenderMsgNum, *intervalTime, 10, *senderNum)
}
