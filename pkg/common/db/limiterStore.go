package db

//import (
//	"Open_IM/pkg/common/config"
//	"context"
//	"github.com/go-redis/redis/v8"
//	"github.com/ulule/limiter/v3"
//	limiterRedis "github.com/ulule/limiter/v3/drivers/store/redis"
//	"time"
//)
//
//var LimiterStore limiter.Store
//
//func init() {
//	var err error
//	var client redis.UniversalClient
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if config.Config.Redis.EnableCluster {
//		client = redis.NewClusterClient(&redis.ClusterOptions{
//			Addrs:    config.Config.Redis.DBAddress,
//			Password: config.Config.Redis.DBPassWord,
//			PoolSize: 50,
//		})
//		if _, err = client.Ping(ctx).Result(); err != nil {
//			panic(err.Error())
//		}
//	} else {
//		client = redis.NewClient(&redis.Options{
//			Addr:     config.Config.Redis.DBAddress[0],
//			Password: config.Config.Redis.DBPassWord,
//			DB:       0,
//			PoolSize: 100,
//		})
//		if _, err = client.Ping(ctx).Result(); err != nil {
//			panic(err.Error())
//		}
//	}
//
//	if LimiterStore, err = limiterRedis.NewStore(client); err != nil {
//		panic(err.Error())
//	}
//}
