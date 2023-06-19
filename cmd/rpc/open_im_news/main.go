package main

import (
	rpcNews "Open_IM/internal/rpc/news"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"flag"
	"fmt"
)

func main() {

	log.NewPrivateLog(constant.OpenImConversationLog)
	defaultPorts := config.Config.RpcPort.OpenImNewsPort
	rpcPort := flag.Int("port", defaultPorts[0], "RpcNews default listen port 10280")
	// rpcPort := flag.Int("port", 10270, "RpcMoments default listen port 11300")
	flag.Parse()
	fmt.Println("start news rpc server, port: ", *rpcPort)
	rpcServer := rpcNews.NewRpcNewsServer(*rpcPort)
	rpcServer.Run()

}
