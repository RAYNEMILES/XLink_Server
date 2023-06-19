package main

import (
	rpcGameStore "Open_IM/internal/rpc/game_store"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"flag"
	"fmt"
)

func main() {

	log.NewPrivateLog(constant.OpenImGameStoreLog)
	defaultPorts := config.Config.RpcPort.OpenImGameStore
	rpcPort := flag.Int("port", defaultPorts[0], "RpcNews default listen port 10290")
	// rpcPort := flag.Int("port", 10270, "RpcMoments default listen port 11300")
	flag.Parse()
	fmt.Println("start news rpc server, port: ", *rpcPort)
	rpcServer := rpcGameStore.NewRpcGameStoreServer(*rpcPort)
	rpcServer.Run()

}
