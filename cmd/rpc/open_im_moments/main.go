package main

import (
	rpcMoments "Open_IM/internal/rpc/moments"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"flag"
	"fmt"
)

func main() {

	log.NewPrivateLog(constant.OpenImConversationLog)
	defaultPorts := config.Config.RpcPort.OpenImMomentsPort
	rpcPort := flag.Int("port", defaultPorts[0], "RpcMoments default listen port 11300")
	// rpcPort := flag.Int("port", 10270, "RpcMoments default listen port 11300")
	flag.Parse()
	fmt.Println("start conversation rpc server, port: ", *rpcPort)
	rpcServer := rpcMoments.NewRpcMomentsServer(*rpcPort)
	rpcServer.Run()

}
