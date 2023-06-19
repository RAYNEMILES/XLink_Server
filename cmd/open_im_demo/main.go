package main

import (
	"Open_IM/internal/demo/register"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"strconv"
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
	//	if err := http.ListenAndServe(":6063", mux); err != nil {
	//		log1.Fatal(err)
	//		log.NewError("", utils.GetSelfFuncName(), "启动pprof报错：", err.Error())
	//	}
	//	os.Exit(0)
	//}()

	log.NewPrivateLog(constant.OpenImDemoLog)
	gin.SetMode(gin.ReleaseMode)
	f, _ := os.Create("../logs/api.log")
	gin.DefaultWriter = io.MultiWriter(f)

	r := gin.Default()
	r.Use(utils.CorsHandler())

	authRouterGroup := r.Group("/demo")
	{
		authRouterGroup.POST("/allow_guest_login", register.AllowGuestLogin)
		authRouterGroup.POST("/code", register.SendVerificationCode)
		authRouterGroup.POST("/verify", register.Verify)
		authRouterGroup.POST("/password", register.SetPassword)
		authRouterGroup.POST("/login", register.Login)
		authRouterGroup.POST("/reset_password", register.ResetPassword)

		authRouterGroup.POST("/generate_in_app_pin", register.GenerateInAppPin)
		authRouterGroup.POST("/login_with_app_pin", register.LoginWithInAppOTP)

		authRouterGroup.POST("/get_verification_code", register.GetVerificationCode)
		authRouterGroup.POST("/register", register.Register)

		authRouterGroup.POST("/get_register_type", register.GetRegisterType)

		authRouterGroup.POST("/get_invite_switch", register.InviteCodeSwitch)

		// distribution channels
		authRouterGroup.POST("/push_invite_code", register.PushInviteCode)
		authRouterGroup.POST("/get_invite_code", register.GetInviteCode)

		// oauth login
		authRouterGroup.POST("/facebook_login", register.FaceBookLogin)
		authRouterGroup.POST("/google_login", register.GoogleLogin)
		authRouterGroup.POST("/apple_login", register.AppleLogin)

		// remote login
		//createRemoteTokenRate := limiter.Rate{Period: time.Minute, Limit: 60}
		//limiterInstance := limiter.New(db.LimiterStore, createRemoteTokenRate, limiter.WithTrustForwardHeader(true))
		//createRemoteTokenRateLimiter := limiterGin.NewMiddleware(limiterInstance)
		//authRouterGroup.POST("/create_remote_token", createRemoteTokenRateLimiter, register.CreateRemoteToken)
		authRouterGroup.POST("/create_remote_token", register.CreateRemoteToken)
		authRouterGroup.POST("/assign_remote_token", register.AssignRemoteToken)
		authRouterGroup.POST("/consume_remote_token", register.ConsumeRemoteToken)

		// interest
		authRouterGroup.POST("/interest_list", register.InterestList)

		// test
		authRouterGroup.POST("/test", register.Test)
	}
	defaultPorts := config.Config.Demo.Port
	ginPort := flag.Int("port", defaultPorts[0], "get ginServerPort from cmd,default 42233 as port")
	flag.Parse()
	fmt.Println("start demo api server, port: ", *ginPort)
	address := "0.0.0.0:" + strconv.Itoa(*ginPort)
	if config.Config.Api.ListenIP != "" {
		address = config.Config.Api.ListenIP + ":" + strconv.Itoa(*ginPort)
	}
	address = config.Config.CmsApi.ListenIP + ":" + strconv.Itoa(*ginPort)
	fmt.Println("start demo api server address: ", address)
	err := r.Run(address)
	if err != nil {
		log.Error("", "run failed ", *ginPort, err.Error())
	}
}
