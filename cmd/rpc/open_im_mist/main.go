package main

import (
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"github.com/nanopack/mist/server"
	"strconv"
)

func main() {

	log.NewPrivateLog(constant.OpenImMistLog)
	defaultPorts := config.Config.RpcPort.OpenImMistPort
	serverURl := "tcp://0.0.0.0:" + strconv.Itoa(defaultPorts[0])

	if err := server.Start([]string{serverURl}, ""); err == nil {
		log.Error("Expecting error - %s\n", err.Error())
	}
}
