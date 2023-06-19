package main

import (
	rpcShortVideo "Open_IM/internal/rpc/short_video"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"flag"
	"fmt"
)

func main() {
	log.NewPrivateLog(constant.OpenImShortVideoLog)
	defaultPorts := config.Config.RpcPort.OpenImShortVideoPort[0]
	rpcPort := flag.Int("port", defaultPorts, "rpc listening port")
	flag.Parse()
	fmt.Println("start short video rpc server, port: ", *rpcPort)
	rpcServer := rpcShortVideo.NewShortVideoServer(*rpcPort)
	rpcServer.Run()
}
