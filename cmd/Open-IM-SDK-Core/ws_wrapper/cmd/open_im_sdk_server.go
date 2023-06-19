/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/9/8 14:35).
 */
package main

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	"Open_IM/cmd/Open-IM-SDK-Core/ws_wrapper/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/ws_wrapper/ws_local_server"
	"Open_IM/pkg/common/log"
	"flag"
	"fmt"
	//	_ "net/http/pprof"
	_ "net/http/pprof"

	"runtime"
	"sync"
)

func main() {
	//runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	//runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪
	//
	//go func() {
	//	// 启动一个自定义mux的http服务器
	//	mux := http.NewServeMux()
	//	mux.HandleFunc("/debug/pprof/", pprof.Index)
	//	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	//
	//	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
	//		w.Write([]byte("hello"))
	//	})
	//	// 启动一个 http server，注意 pprof 相关的 handler 已经自动注册过了
	//	if err := http.ListenAndServe(":6060", mux); err != nil {
	//		log1.Fatal(err)
	//		log.NewError("", utils2.GetSelfFuncName(), "启动pprof报错：", err.Error())
	//	}
	//	os.Exit(0)
	//}()

	var sdkWsPort, openIMApiPort, openIMWsPort *int
	var openIMWsAddress, openIMApiAddress *string
	//
	// openIMTerminalType := flag.String("terminal_type", "web", "different terminal types")

	sdkWsPort = flag.Int("sdk_ws_port", 10003, "openIMSDK ws listening port")
	openIMApiPort = flag.Int("openIM_api_port", 10002, "openIM api listening port")
	openIMWsPort = flag.Int("openIM_ws_port", 10001, "openIM ws listening port")
	flag.Parse()
	// switch *openIMTerminalType {
	// case "pc":
	//	openIMWsAddress = flag.String("openIM_ws_address", "web", "different terminal types")
	//	openIMApiAddress = flag.String("openIM_api_address", "web", "different terminal types")
	//	flag.Parse()
	// case "web":
	//	openIMApiPort = flag.Int("openIM_api_port", 0, "openIM api listening port")
	//	openIMWsPort = flag.Int("openIM_ws_port", 0, "openIM ws listening port")
	//	flag.Parse()
	// }
	APIADDR := "http://43.128.5.63:10000"
	WSADDR := "ws://43.128.5.63:17778"

	sysType := runtime.GOOS
	log.NewPrivateLog(constant.LogFileName)
	open_im_sdk.SetHeartbeatInterval(5)
	switch sysType {

	case "darwin":
		ws_local_server.InitServer(&sdk_struct.IMConfig{ApiAddr: *openIMApiAddress,
			WsAddr: *openIMWsAddress, Platform: utils.OSXPlatformID, DataDir: "./"})
	case "linux":
		// sdkDBDir:= flag.String("sdk_db_dir","","openIMSDK initialization path")
		ws_local_server.InitServer(&sdk_struct.IMConfig{ApiAddr: "http://" + utils.ServerIP + ":" + utils.IntToString(*openIMApiPort),
			WsAddr: "ws://" + utils.ServerIP + ":" + utils.IntToString(*openIMWsPort), Platform: utils.WebPlatformID, DataDir: "../db/sdk/"})

	case "windows":
		//	sdkWsPort = flag.Int("sdk_ws_port", 7799, "openIM ws listening port")
		// flag.Parse()
		ws_local_server.InitServer(&sdk_struct.IMConfig{ApiAddr: APIADDR,
			WsAddr: WSADDR, Platform: utils.WebPlatformID, DataDir: "./"})
	default:
		fmt.Println("this os not support", sysType)

	}
	var wg sync.WaitGroup
	wg.Add(1)
	fmt.Println("ws server is starting")
	ws_local_server.WS.OnInit(*sdkWsPort)
	ws_local_server.WS.Run()
	wg.Wait()

}
