package friend

import (
	chat "Open_IM/internal/rpc/msg"
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	cp "Open_IM/pkg/common/utils"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbCache "Open_IM/pkg/proto/cache"
	pbChat "Open_IM/pkg/proto/chat"
	pbFriend "Open_IM/pkg/proto/friend"
	"Open_IM/pkg/proto/local_database"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	pbUser "Open_IM/pkg/proto/user"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"

	"google.golang.org/grpc"
)

type friendServer struct {
	rpcPort         int
	rpcRegisterName string
	etcdSchema      string
	etcdAddr        []string
}

func (s *friendServer) ChannelAddFriend(ctx context.Context, request *pbFriend.ChannelAddFriendRequset) (*pbFriend.ChannelAddFriendResponse, error) {
	log.NewInfo(request.OperationID, "channelFriend args ", request.String())
	resp := pbFriend.ChannelAddFriendResponse{CommonResp: &pbFriend.CommonResp{}}
	var c pbFriend.CommonResp

	if _, err := imdb.GetUserByUserID(request.FromUserID); err != nil {
		log.NewError(request.OperationID, "GetUserByUserID failed ", err.Error(), request.FromUserID)
		c.ErrCode = constant.ErrDB.ErrCode
		c.ErrMsg = "this user not exists,cant not add friend"
		for _, v := range request.FriendUserIDList {
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		}
		resp.CommonResp = &c
		return &resp, nil
	}

	for _, v := range request.FriendUserIDList {
		log.NewDebug(request.OperationID, "FriendUserIDList ", v)
		if _, fErr := imdb.GetUserByUserID(v); fErr != nil {
			log.NewError(request.OperationID, "GetUserByUserID failed ", fErr.Error(), v)
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		} else {
			if _, err := imdb.GetFriendRelationshipFromFriend(request.FromUserID, v); err != nil {
				// Establish two single friendship
				toInsertFollow := db.Friend{OwnerUserID: request.FromUserID, FriendUserID: v}
				err1 := imdb.InsertToFriend(&toInsertFollow)
				if err1 != nil {
					log.NewError(request.OperationID, "InsertToFriend failed ", err1.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				toInsertFollow = db.Friend{OwnerUserID: v, FriendUserID: request.FromUserID}
				err2 := imdb.InsertToFriend(&toInsertFollow)
				if err2 != nil {
					log.NewError(request.OperationID, "InsertToFriend failed ", err2.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
				log.NewDebug(request.OperationID, "UserIDResultList ", resp.UserIDResultList)
				chat.FriendAddedNotification(request.OperationID, request.FromUserID, request.FromUserID, v)

				// add friend conversation
				var channelCodeOwnerConversationPb pbUser.SetConversationReq
				var channelCodeOwnerConversationC pbUser.Conversation
				channelCodeOwnerConversationPb.OperationID = request.OperationID
				channelCodeOwnerConversationC.OwnerUserID = request.FromUserID
				channelCodeOwnerConversationC.ConversationID = utils.GetConversationIDBySessionType(v, constant.SingleChatType)
				channelCodeOwnerConversationC.ConversationType = constant.SingleChatType
				channelCodeOwnerConversationC.UserID = v
				channelCodeOwnerConversationC.IsNotInGroup = false
				channelCodeOwnerConversationPb.Conversation = &channelCodeOwnerConversationC
				etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					client := pbUser.NewUserClient(etcdConn)
					respPb, err := client.SetConversation(context.Background(), &channelCodeOwnerConversationPb)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", channelCodeOwnerConversationPb.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
					}
				}

				// add friend cache
				etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					ownerAddFriendToCacheReq := &pbCache.AddFriendToCacheReq{
						UserID:      request.FromUserID,
						FriendID:    v,
						OperationID: request.OperationID,
					}
					cacheClient := pbCache.NewCacheClient(etcdConn)
					cacheResp, err := cacheClient.AddFriendToCache(context.Background(), ownerAddFriendToCacheReq)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "add owner cache rpc failed, ", ownerAddFriendToCacheReq.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "add owner cache success", cacheResp.String(), v)
					}
				}

				// add owner conversation
				var channelCodeFriendConversationPb pbUser.SetConversationReq
				var channelCodeFriendConversationC pbUser.Conversation
				channelCodeFriendConversationPb.OperationID = request.OperationID
				channelCodeFriendConversationC.OwnerUserID = v
				channelCodeFriendConversationC.ConversationID = utils.GetConversationIDBySessionType(v, constant.SingleChatType)
				channelCodeFriendConversationC.ConversationType = constant.SingleChatType
				channelCodeFriendConversationC.UserID = request.FromUserID
				channelCodeFriendConversationC.IsNotInGroup = false
				channelCodeFriendConversationPb.Conversation = &channelCodeFriendConversationC
				etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					client := pbUser.NewUserClient(etcdConn)
					respPb, err := client.SetConversation(context.Background(), &channelCodeFriendConversationPb)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", channelCodeFriendConversationPb.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
					}
				}

				// add owner cache
				etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					ownerAddFriendToCacheReq := &pbCache.AddFriendToCacheReq{
						UserID:      v,
						FriendID:    request.FromUserID,
						OperationID: request.OperationID,
					}
					cacheClient := pbCache.NewCacheClient(etcdConn)
					cacheResp, err := cacheClient.AddFriendToCache(context.Background(), ownerAddFriendToCacheReq)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "add friend cache rpc failed, ", ownerAddFriendToCacheReq.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "add friend cache success", cacheResp.String(), v)
					}
				}

				// add friend cache

				// send message to friend
				// v -> FromUserID
				if request.Greeting != "" {
					sendTextMsgToFriend(v, request.FromUserID, request.Greeting, request.OperationID)
				}

				// relation
				imdb.AddFriendRalation(v, request.FromUserID, 2)
			} else {
				log.NewWarn(request.OperationID, "GetFriendRelationshipFromFriend ok", request.FromUserID, v)
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
			}
		}
	}

	resp.CommonResp.ErrCode = 0
	log.NewInfo(request.OperationID, "channelAddFriend rpc ok ", resp.String())
	return &resp, nil
}

func (s *friendServer) AutoAddFriend(ctx context.Context, request *pbFriend.AutoAddFriendRequset) (*pbFriend.AutoAddFriendResponse, error) {
	log.NewInfo(request.OperationID, "ImportFriend args ", request.String())
	resp := pbFriend.AutoAddFriendResponse{CommonResp: &pbFriend.CommonResp{}}
	var c pbFriend.CommonResp

	// if config.Config.RestrictUserAddFriendOps {
	// 	log.NewError(request.OperationID, "Hard Check enabled for stoping adding Friend ")
	// 	resp.CommonResp.ErrCode = constant.ErrAccess.ErrCode
	// 	resp.CommonResp.ErrMsg = constant.ErrAccess.ErrMsg
	// 	return &resp, nil

	// }

	if _, err := imdb.GetUserByUserID(request.FromUserID); err != nil {
		log.NewError(request.OperationID, "GetUserByUserID failed ", err.Error(), request.FromUserID)
		c.ErrCode = constant.ErrDB.ErrCode
		c.ErrMsg = "this user not exists,cant not add friend"
		for _, v := range request.FriendUserIDList {
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		}
		resp.CommonResp = &c
		return &resp, nil
	}

	for _, v := range request.FriendUserIDList {
		log.NewDebug(request.OperationID, "FriendUserIDList ", v)
		if _, fErr := imdb.GetUserByUserID(v); fErr != nil {
			log.NewError(request.OperationID, "GetUserByUserID failed ", fErr.Error(), v)
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		} else {
			if _, err := imdb.GetFriendRelationshipFromFriend(request.FromUserID, v); err != nil {
				// Establish two single friendship
				toInsertFollow := db.Friend{OwnerUserID: request.FromUserID, FriendUserID: v}
				err1 := imdb.InsertToFriend(&toInsertFollow)
				if err1 != nil {
					log.NewError(request.OperationID, "InsertToFriend failed ", err1.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				toInsertFollow = db.Friend{OwnerUserID: v, FriendUserID: request.FromUserID}
				err2 := imdb.InsertToFriend(&toInsertFollow)
				if err2 != nil {
					log.NewError(request.OperationID, "InsertToFriend failed ", err2.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
				log.NewDebug(request.OperationID, "UserIDResultList ", resp.UserIDResultList)
				chat.FriendAddedNotification(request.OperationID, request.FromUserID, request.FromUserID, v)

				// add owner conversation
				var channelCodeOwnerConversationPb pbUser.SetConversationReq
				var channelCodeOwnerConversationC pbUser.Conversation
				channelCodeOwnerConversationPb.OperationID = request.OperationID
				channelCodeOwnerConversationC.OwnerUserID = request.FromUserID
				channelCodeOwnerConversationC.ConversationID = utils.GetConversationIDBySessionType(v, constant.SingleChatType)
				channelCodeOwnerConversationC.ConversationType = constant.SingleChatType
				channelCodeOwnerConversationC.UserID = v
				channelCodeOwnerConversationC.IsNotInGroup = false
				channelCodeOwnerConversationPb.Conversation = &channelCodeOwnerConversationC
				etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					client := pbUser.NewUserClient(etcdConn)
					respPb, err := client.SetConversation(context.Background(), &channelCodeOwnerConversationPb)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", channelCodeOwnerConversationPb.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
					}
				}

				// add owner cache
				etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					ownerAddFriendToCacheReq := &pbCache.AddFriendToCacheReq{
						UserID:      request.FromUserID,
						FriendID:    v,
						OperationID: request.OperationID,
					}
					cacheClient := pbCache.NewCacheClient(etcdConn)
					cacheResp, err := cacheClient.AddFriendToCache(context.Background(), ownerAddFriendToCacheReq)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "add owner cache rpc failed, ", ownerAddFriendToCacheReq.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "add owner cache success", cacheResp.String(), v)
					}
				}

				// add owner conversation
				var channelCodeFriendConversationPb pbUser.SetConversationReq
				var channelCodeFriendConversationC pbUser.Conversation
				channelCodeFriendConversationPb.OperationID = request.OperationID
				channelCodeFriendConversationC.OwnerUserID = v
				channelCodeFriendConversationC.ConversationID = utils.GetConversationIDBySessionType(v, constant.SingleChatType)
				channelCodeFriendConversationC.ConversationType = constant.SingleChatType
				channelCodeFriendConversationC.UserID = request.FromUserID
				channelCodeFriendConversationC.IsNotInGroup = false
				channelCodeFriendConversationPb.Conversation = &channelCodeFriendConversationC
				etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImUserName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					client := pbUser.NewUserClient(etcdConn)
					respPb, err := client.SetConversation(context.Background(), &channelCodeFriendConversationPb)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "SetConversation rpc failed, ", channelCodeFriendConversationPb.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "SetConversation success", respPb.String(), v)
					}
				}

				// add friend cache
				etcdConn = getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, request.OperationID)
				if etcdConn == nil {
					errMsg := request.OperationID + "getcdv3.GetConn == nil"
					log.NewError(request.OperationID, errMsg, request.FromUserID, v)
				} else {
					ownerAddFriendToCacheReq := &pbCache.AddFriendToCacheReq{
						UserID:      v,
						FriendID:    request.FromUserID,
						OperationID: request.OperationID,
					}
					cacheClient := pbCache.NewCacheClient(etcdConn)
					cacheResp, err := cacheClient.AddFriendToCache(context.Background(), ownerAddFriendToCacheReq)
					if err != nil {
						log.NewError(request.OperationID, utils.GetSelfFuncName(), "add friend cache rpc failed, ", ownerAddFriendToCacheReq.String(), err.Error(), v)
					} else {
						log.NewDebug(request.OperationID, utils.GetSelfFuncName(), "add friend cache success", cacheResp.String(), v)
					}
				}

				// send message to friend
				// FromUserID -> v
				if request.Greeting != "" {
					sendTextMsgToFriend(request.FromUserID, v, request.Greeting, request.OperationID)
				}

				// relation
				imdb.AddFriendRalation(request.FromUserID, v, 1)
			} else {
				log.NewWarn(request.OperationID, "GetFriendRelationshipFromFriend ok", request.FromUserID, v)
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
			}
		}
	}
	resp.CommonResp.ErrCode = 0
	log.NewInfo(request.OperationID, "AutoAddFriend rpc ok ", resp.String())
	return &resp, nil
}

func (s *friendServer) CheckFriendFromCache(ctx context.Context, req *pbFriend.IsFriendReq) (*pbFriend.IsFriendResp, error) {
	// TODO implement me
	panic("implement me")
}

func (s *friendServer) CheckBlockFromCache(ctx context.Context, req *pbFriend.IsInBlackListReq) (*pbFriend.IsFriendResp, error) {
	// TODO implement me
	panic("implement me")
}

func NewFriendServer(port int) *friendServer {
	return &friendServer{
		rpcPort:         port,
		rpcRegisterName: config.Config.RpcRegisterName.OpenImFriendName,
		etcdSchema:      config.Config.Etcd.EtcdSchema,
		etcdAddr:        config.Config.Etcd.EtcdAddr,
	}
}

func (s *friendServer) Run() {
	log.NewPrivateLog(constant.OpenImFriendLog)
	log.NewInfo("0", "friendServer run...")

	listenIP := ""
	if config.Config.ListenIP == "" {
		listenIP = "0.0.0.0"
	} else {
		listenIP = config.Config.ListenIP
	}
	address := listenIP + ":" + strconv.Itoa(s.rpcPort)

	// listener network
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic("listening err:" + err.Error() + s.rpcRegisterName)
	}
	log.NewInfo("0", "listen ok ", address)
	defer listener.Close()
	// grpc server
	srv := grpc.NewServer()
	defer srv.GracefulStop()
	// User friend related services register to etcd
	pbFriend.RegisterFriendServer(srv, s)
	rpcRegisterIP := config.Config.RpcRegisterIP
	if config.Config.RpcRegisterIP == "" {
		rpcRegisterIP, err = utils.GetLocalIP()
		if err != nil {
			log.Error("", "GetLocalIP failed ", err.Error())
		}
	}
	log.NewInfo("", "rpcRegisterIP", rpcRegisterIP)
	err = getcdv3.RegisterEtcd(s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName, 10)
	if err != nil {
		log.NewError("0", "RegisterEtcd failed ", err.Error(), s.etcdSchema, strings.Join(s.etcdAddr, ","), rpcRegisterIP, s.rpcPort, s.rpcRegisterName)
		return
	}
	err = srv.Serve(listener)
	if err != nil {
		log.NewError("0", "Serve failed ", err.Error(), listener)
		return
	}
}

func (s *friendServer) AddBlacklist(ctx context.Context, req *pbFriend.AddBlacklistReq) (*pbFriend.AddBlacklistResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddBlacklist args ", req.String())

	black := db.Black{OwnerUserID: req.CommID.FromUserID, BlockUserID: req.CommID.ToUserID, OperatorUserID: req.CommID.OpUserID}

	err := imdb.InsertInToUserBlackList(black)
	if err != nil {
		log.NewError(req.CommID.OperationID, "InsertInToUserBlackList failed ", err.Error())
		return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	log.NewInfo(req.CommID.OperationID, "AddBlacklist rpc ok ", req.CommID.FromUserID, req.CommID.ToUserID)
	reqAddBlackUserToCache := &pbCache.AddBlackUserToCacheReq{UserID: req.CommID.FromUserID, BlackUserID: req.CommID.ToUserID, OperationID: req.CommID.OperationID}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.CommID.OperationID)
	if etcdConn == nil {
		errMsg := req.CommID.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.CommID.OperationID, errMsg)
		return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}
	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.AddBlackUserToCache(context.Background(), reqAddBlackUserToCache)
	if err != nil {
		log.NewError(req.CommID.OperationID, "AddBlackUserToCache rpc call failed ", err.Error())
		return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: 500, ErrMsg: "AddBlackUserToCache rpc call failed"}}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.CommID.OperationID, "AddBlackUserToCache rpc logic call failed ", cacheResp.String())
		return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: cacheResp.CommonResp.ErrCode, ErrMsg: cacheResp.CommonResp.ErrMsg}}, nil
	}

	chat.BlackAddedNotification(req)
	return &pbFriend.AddBlacklistResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) AddFriend(ctx context.Context, req *pbFriend.AddFriendReq) (*pbFriend.AddFriendResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddFriend args ", req.String())
	// if config.Config.RestrictUserAddFriendOps {
	// 	fromUserObj, errU := imdb.GetUserByUserID(req.CommID.FromUserID)
	// 	if errU == nil && fromUserObj.SuperUserStatus == 1 {
	// 	} else {
	// 		toUserObj, errU := imdb.GetUserByUserID(req.CommID.ToUserID)
	// 		if errU == nil && toUserObj.SuperUserStatus == 1 {
	// 		} else {
	// 			log.NewError(req.CommID.OperationID, "Hard Check enabled for stoping adding Friend ", req.CommID.OpUserID, req.CommID.FromUserID)
	// 			return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.AddFriendNotSuperUserErr.ErrCode, ErrMsg: constant.AddFriendNotSuperUserErr.ErrMsg}}, errU
	// 		}
	// 	}
	// }
	friend, err := imdb.GetFriendRelationshipFromFriend(req.CommID.ToUserID, req.CommID.FromUserID)
	if err == nil && friend != nil {
		reqAFR := &pbFriend.AddFriendResponseReq{}
		err := utils.CopyStructFields(reqAFR, req)
		if err == nil {
			reqAFR.HandleResult = 1
			_, err := s.AddFriendResponseIfAlreadyFriends(context.Background(), reqAFR)
			if err == nil {
				return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{}}, err
			}
		}
	}
	// cannot add himself
	if req.CommID.FromUserID == req.CommID.ToUserID {
		log.NewError(req.CommID.OperationID, "cannot add himself")
		return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: constant.ErrArgs.ErrMsg}}, nil
	}
	// Cannot add non-existent users
	if _, err := imdb.GetUserByUserID(req.CommID.ToUserID); err != nil {
		log.NewError(req.CommID.OperationID, "GetUserByUserID failed ", err.Error(), req.CommID.ToUserID)
		return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	// Cannot add friends
	toUserPrivacy, _ := imdb.GetPrivacySetting(req.CommID.ToUserID)
	privacyMap := make(map[string]string)
	for _, v := range toUserPrivacy {
		privacyMap[v.SettingKey] = v.SettingVal
	}

	// group chat limit
	if req.Source == constant.AddFriendSourceGroup {
		if privacyMap[constant.PrivacyAddByGroup] == constant.PrivacyStatusClose {
			return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.AddFriendLimitGroupErr.ErrCode, ErrMsg: constant.AddFriendLimitGroupErr.ErrMsg}}, nil
		}
	}

	// qr code limit
	if req.Source == constant.AddFriendSourceQr {
		if privacyMap[constant.PrivacyAddByQr] == constant.PrivacyStatusClose {
			return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.AddFriendLimitQrErr.ErrCode, ErrMsg: constant.AddFriendLimitQrErr.ErrMsg}}, nil
		}
	}

	// contact card limit
	if req.Source == constant.AddFriendSourceCard {
		if privacyMap[constant.PrivacyAddByContactCard] == constant.PrivacyStatusClose {
			return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.AddFriendLimitContactErr.ErrCode, ErrMsg: constant.AddFriendLimitContactErr.ErrMsg}}, nil
		}
	}

	// Establish a latest relationship in the friend request table
	friendRequest := db.FriendRequest{
		HandleResult: 0, ReqMsg: req.ReqMsg, CreateTime: time.Now()}
	utils.CopyStructFields(&friendRequest, req.CommID)
	// {openIM001 openIM002 0 test add friend 0001-01-01 00:00:00 +0000 UTC   0001-01-01 00:00:00 +0000 UTC }]
	log.NewDebug(req.CommID.OperationID, "UpdateFriendApplication args ", friendRequest)
	// err := imdb.InsertFriendApplication(&friendRequest)
	err = imdb.InsertFriendApplication(&friendRequest,
		map[string]interface{}{"handle_result": 0, "req_msg": friendRequest.ReqMsg, "create_time": friendRequest.CreateTime,
			"handler_user_id": "", "handle_msg": "", "ex": ""})
	if err != nil {
		log.NewError(req.CommID.OperationID, "UpdateFriendApplication failed ", err.Error(), friendRequest)
		return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	//data synchronization
	syncFriendRequestToLocal(req.CommID.OperationID, req.CommID.FromUserID, req.CommID.ToUserID, constant.SyncFriendRequest)

	chat.FriendApplicationNotification(req)
	return &pbFriend.AddFriendResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) ImportFriend(ctx context.Context, req *pbFriend.ImportFriendReq) (*pbFriend.ImportFriendResp, error) {
	log.NewInfo(req.OperationID, "ImportFriend args ", req.String())
	resp := pbFriend.ImportFriendResp{CommonResp: &pbFriend.CommonResp{}}
	var c pbFriend.CommonResp

	// if config.Config.RestrictUserAddFriendOps {
	// 	fromUserObj, errU := imdb.GetUserByUserID(req.FromUserID)
	// 	if errU == nil && fromUserObj.SuperUserStatus == 1 {
	// 	} else {
	// 		log.NewError(req.OperationID, "Hard Check enabled for stoping adding Friend ", req.OpUserID, req.FromUserID)
	// 		c.ErrCode = constant.ErrAddFriendStoped.ErrCode
	// 		c.ErrMsg = "Add friend closed."
	// 		for _, v := range req.FriendUserIDList {
	// 			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
	// 		}
	// 		resp.CommonResp = &c
	// 		return &resp, errU
	// 	}
	// }
	if !utils.IsContain(req.OpUserID, config.Config.Manager.AppManagerUid) {

		log.NewError(req.OperationID, "not authorized", req.OpUserID, config.Config.Manager.AppManagerUid)

		c.ErrCode = constant.ErrAccess.ErrCode
		c.ErrMsg = constant.ErrAccess.ErrMsg
		for _, v := range req.FriendUserIDList {
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		}
		resp.CommonResp = &c
		return &resp, nil
	}
	if _, err := imdb.GetUserByUserID(req.FromUserID); err != nil {
		log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), req.FromUserID)
		c.ErrCode = constant.ErrDB.ErrCode
		c.ErrMsg = "this user not exists,cant not add friend"
		for _, v := range req.FriendUserIDList {
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		}
		resp.CommonResp = &c
		return &resp, nil
	}

	for _, v := range req.FriendUserIDList {
		log.NewDebug(req.OperationID, "FriendUserIDList ", v)
		if _, fErr := imdb.GetUserByUserID(v); fErr != nil {
			log.NewError(req.OperationID, "GetUserByUserID failed ", fErr.Error(), v)
			resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
		} else {
			if _, err := imdb.GetFriendRelationshipFromFriend(req.FromUserID, v); err != nil {
				// Establish two single friendship
				toInsertFollow := db.Friend{OwnerUserID: req.FromUserID, FriendUserID: v}
				err1 := imdb.InsertToFriend(&toInsertFollow)
				if err1 != nil {
					log.NewError(req.OperationID, "InsertToFriend failed ", err1.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				toInsertFollow = db.Friend{OwnerUserID: v, FriendUserID: req.FromUserID}
				err2 := imdb.InsertToFriend(&toInsertFollow)
				if err2 != nil {
					log.NewError(req.OperationID, "InsertToFriend failed ", err2.Error(), toInsertFollow)
					resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: -1})
					continue
				}
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
				log.NewDebug(req.OperationID, "UserIDResultList ", resp.UserIDResultList)
				chat.FriendAddedNotification(req.OperationID, req.OpUserID, req.FromUserID, v)
			} else {
				log.NewWarn(req.OperationID, "GetFriendRelationshipFromFriend ok", req.FromUserID, v)
				resp.UserIDResultList = append(resp.UserIDResultList, &pbFriend.UserIDResult{UserID: v, Result: 0})
			}
		}
	}
	resp.CommonResp.ErrCode = 0
	log.NewInfo(req.OperationID, "ImportFriend rpc ok ", resp.String())
	return &resp, nil
}

func (s *friendServer) AddFriendResponseIfAlreadyFriends(ctx context.Context, req *pbFriend.AddFriendResponseReq) (*pbFriend.AddFriendResponseResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddFriendResponse args ", req.String())

	// Change the status of the friend request form
	if req.HandleResult == constant.FriendResponseAgree {
		// Establish friendship after find friend relationship not exists
		_, err := imdb.GetFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
		if err == nil {
			log.NewWarn(req.CommID.OperationID, "GetFriendRelationshipFromFriend exist", req.CommID.FromUserID, req.CommID.ToUserID)
		} else {
			// Establish two single friendship
			toInsertFollow := db.Friend{OwnerUserID: req.CommID.FromUserID, FriendUserID: req.CommID.ToUserID, OperatorUserID: req.CommID.OpUserID}
			err = imdb.InsertToFriend(&toInsertFollow)
			if err != nil {
				log.NewError(req.CommID.OperationID, "InsertToFriend failed ", err.Error(), toInsertFollow)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			}
		}

		_, err = imdb.GetFriendRelationshipFromFriend(req.CommID.ToUserID, req.CommID.FromUserID)
		if err == nil {
			log.NewWarn(req.CommID.OperationID, "GetFriendRelationshipFromFriend exist", req.CommID.ToUserID, req.CommID.FromUserID)
		} else {
			toInsertFollow := db.Friend{OwnerUserID: req.CommID.ToUserID, FriendUserID: req.CommID.FromUserID, OperatorUserID: req.CommID.OpUserID}
			err = imdb.InsertToFriend(&toInsertFollow)
			if err != nil {
				log.NewError(req.CommID.OperationID, "InsertToFriend failed ", err.Error(), toInsertFollow)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			}
		}
		// cache rpc
		addFriendToCacheReq := &pbCache.AddFriendToCacheReq{OperationID: req.CommID.OperationID}
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.CommID.OperationID)
		if etcdConn == nil {
			errMsg := req.CommID.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.CommID.OperationID, errMsg)
			return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
		}
		client := pbCache.NewCacheClient(etcdConn)
		addFriendToCacheReq.UserID = req.CommID.ToUserID
		addFriendToCacheReq.FriendID = req.CommID.FromUserID
		respPb, err := client.AddFriendToCache(context.Background(), addFriendToCacheReq)
		if err != nil {
			log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", err.Error(), addFriendToCacheReq.String())
			return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: constant.ErrServer.ErrMsg}}, nil
		}
		if respPb.CommonResp.ErrCode != 0 {
			log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", addFriendToCacheReq.String())
			return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: respPb.CommonResp.ErrCode, ErrMsg: respPb.CommonResp.ErrMsg}}, nil
		}
		addFriendToCacheReq.UserID = req.CommID.FromUserID
		addFriendToCacheReq.FriendID = req.CommID.ToUserID
		respPb, err = client.AddFriendToCache(context.Background(), addFriendToCacheReq)
		if err != nil {
			log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", err.Error(), addFriendToCacheReq.String())
			return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: constant.ErrServer.ErrMsg}}, nil
		}
		if respPb.CommonResp.ErrCode != 0 {
			log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", addFriendToCacheReq.String())
			return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: respPb.CommonResp.ErrCode, ErrMsg: respPb.CommonResp.ErrMsg}}, nil
		}

		// clear cache
		go func() {
			db.DB.DelUsersInfoByUserIdCache(req.CommID.ToUserID)
			db.DB.DelUsersInfoByUserIdCache(req.CommID.FromUserID)
		}()

		chat.FriendAddedNotification(req.CommID.OperationID, req.CommID.OpUserID, req.CommID.FromUserID, req.CommID.ToUserID)

	}

	if req.HandleResult == constant.FriendResponseAgree {
		//data synchronization
		syncFriendRequestToLocal(req.CommID.OperationID, req.CommID.FromUserID, req.CommID.ToUserID, constant.SyncFriendRequest)

		chat.FriendApplicationApprovedNotification(req)
	} else if req.HandleResult == constant.FriendResponseRefuse {
		//data synchronization
		syncFriendRequestToLocal(req.CommID.OperationID, req.CommID.FromUserID, req.CommID.ToUserID, constant.SyncFriendRequest)

		chat.FriendApplicationRejectedNotification(req)
	} else {
		log.Error(req.CommID.OperationID, "HandleResult failed ", req.HandleResult)
	}

	log.NewInfo(req.CommID.OperationID, "rpc AddFriendResponse ok")
	return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

// process Friend application
func (s *friendServer) AddFriendResponse(ctx context.Context, req *pbFriend.AddFriendResponseReq) (*pbFriend.AddFriendResponseResp, error) {
	log.NewInfo(req.CommID.OperationID, "AddFriendResponse args ", req.String())

	// if config.Config.RestrictUserAddFriendOps {
	// 	fromUserObj, errU := imdb.GetUserByUserID(req.CommID.FromUserID)
	// 	if errU == nil && fromUserObj.SuperUserStatus == 1 {
	// 	} else {
	// 		toUserObj, errU := imdb.GetUserByUserID(req.CommID.ToUserID)
	// 		if errU == nil && toUserObj.SuperUserStatus == 1 {
	// 		} else {
	// 			log.NewError(req.CommID.OperationID, "Hard Check enabled for stoping adding Friend ", req.CommID.OpUserID, req.CommID.FromUserID)
	// 			return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAddFriendStoped.ErrCode, ErrMsg: "Add friend closed."}}, errU
	// 		}
	// 	}
	// }

	// Check there application before agreeing or refuse to a friend's application
	// req.CommID.FromUserID process req.CommID.ToUserID
	friendRequest, err := imdb.GetFriendApplicationByBothUserID(req.CommID.ToUserID, req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetFriendApplicationByBothUserID failed ", err.Error(), req.CommID.ToUserID, req.CommID.FromUserID)
		return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}
	friendRequest.HandleResult = req.HandleResult
	tmpTime := time.Now()
	friendRequest.HandleTime = &tmpTime
	// friendRequest.HandleTime.Unix()
	friendRequest.HandleMsg = req.HandleMsg
	friendRequest.HandlerUserID = req.CommID.OpUserID
	err = imdb.UpdateFriendApplication(friendRequest)
	if err != nil {
		log.NewError(req.CommID.OperationID, "UpdateFriendApplication failed ", err.Error(), friendRequest)
		return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	// Change the status of the friend request form
	if req.HandleResult == constant.FriendResponseAgree {
		// Establish friendship after find friend relationship not exists
		_, err := imdb.GetFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
		if err == nil {
			log.NewWarn(req.CommID.OperationID, "GetFriendRelationshipFromFriend exist", req.CommID.FromUserID, req.CommID.ToUserID)
		} else {
			// Establish two single friendship
			toInsertFollow := db.Friend{OwnerUserID: req.CommID.FromUserID, FriendUserID: req.CommID.ToUserID, OperatorUserID: req.CommID.OpUserID}
			err = imdb.InsertToFriend(&toInsertFollow)
			if err != nil {
				log.NewError(req.CommID.OperationID, "InsertToFriend failed ", err.Error(), toInsertFollow)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			}
		}

		_, err = imdb.GetFriendRelationshipFromFriend(req.CommID.ToUserID, req.CommID.FromUserID)
		if err == nil {
			log.NewWarn(req.CommID.OperationID, "GetFriendRelationshipFromFriend exist", req.CommID.ToUserID, req.CommID.FromUserID)
		} else {
			toInsertFollow := db.Friend{OwnerUserID: req.CommID.ToUserID, FriendUserID: req.CommID.FromUserID, OperatorUserID: req.CommID.OpUserID}
			err = imdb.InsertToFriend(&toInsertFollow)
			if err != nil {
				log.NewError(req.CommID.OperationID, "InsertToFriend failed ", err.Error(), toInsertFollow)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
			}
			// cache friendship to localdata
			//syncAddFriendReq := &local_database.SyncAddFriendReq{
			//	UserID:     req.CommID.FromUserID,
			//	FriendID:   req.CommID.ToUserID,
			//	OperatorID: req.CommID.OpUserID,
			//}
			//etcdLocalConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, req.CommID.OperationID)
			//if etcdLocalConn == nil {
			//	errMsg := req.CommID.OperationID + "etcdLocalConn == nil"
			//	log.NewError(req.CommID.OperationID, errMsg)
			//	return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
			//}
			//localClient := local_database.NewLocalDataBaseClient(etcdLocalConn)
			//respPbLocal, err := localClient.SyncAddFriendToLocal(context.Background(), syncAddFriendReq)
			//if err != nil {
			//	log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "syncAddFriendReq failed", err.Error(), syncAddFriendReq.String())
			//	return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: constant.ErrServer.ErrMsg}}, nil
			//}
			//if respPbLocal.ErrCode != 0 {
			//	log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "syncAddFriendReq failed", respPbLocal.ErrMsg)
			//	return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: respPbLocal.ErrCode, ErrMsg: respPbLocal.ErrMsg}}, nil
			//}

			// cache rpc
			addFriendToCacheReq := &pbCache.AddFriendToCacheReq{OperationID: req.CommID.OperationID}
			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.CommID.OperationID)
			if etcdConn == nil {
				errMsg := req.CommID.OperationID + "getcdv3.GetConn == nil"
				log.NewError(req.CommID.OperationID, errMsg)
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
			}
			client := pbCache.NewCacheClient(etcdConn)
			addFriendToCacheReq.UserID = req.CommID.ToUserID
			addFriendToCacheReq.FriendID = req.CommID.FromUserID
			respPb, err := client.AddFriendToCache(context.Background(), addFriendToCacheReq)
			if err != nil {
				log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", err.Error(), addFriendToCacheReq.String())
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: constant.ErrServer.ErrMsg}}, nil
			}
			if respPb.CommonResp.ErrCode != 0 {
				log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", addFriendToCacheReq.String())
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: respPb.CommonResp.ErrCode, ErrMsg: respPb.CommonResp.ErrMsg}}, nil
			}
			addFriendToCacheReq.UserID = req.CommID.FromUserID
			addFriendToCacheReq.FriendID = req.CommID.ToUserID
			respPb, err = client.AddFriendToCache(context.Background(), addFriendToCacheReq)
			if err != nil {
				log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", err.Error(), addFriendToCacheReq.String())
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: constant.ErrServer.ErrMsg}}, nil
			}
			if respPb.CommonResp.ErrCode != 0 {
				log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed", addFriendToCacheReq.String())
				return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{ErrCode: respPb.CommonResp.ErrCode, ErrMsg: respPb.CommonResp.ErrMsg}}, nil
			}

			// clear cache
			go func() {
				db.DB.DelUsersInfoByUserIdCache(req.CommID.ToUserID)
				db.DB.DelUsersInfoByUserIdCache(req.CommID.FromUserID)
			}()

			chat.FriendAddedNotification(req.CommID.OperationID, req.CommID.OpUserID, req.CommID.FromUserID, req.CommID.ToUserID)
		}
	}

	if req.HandleResult == constant.FriendResponseAgree {
		//data synchronization
		syncFriendRequestToLocal(req.CommID.OperationID, friendRequest.FromUserID, friendRequest.ToUserID, constant.SyncFriendRequest)

		chat.FriendApplicationApprovedNotification(req)
	} else if req.HandleResult == constant.FriendResponseRefuse {
		//data synchronization
		syncFriendRequestToLocal(req.CommID.OperationID, friendRequest.FromUserID, friendRequest.ToUserID, constant.SyncFriendRequest)

		chat.FriendApplicationRejectedNotification(req)
	} else {
		log.Error(req.CommID.OperationID, "HandleResult failed ", req.HandleResult)
	}

	log.NewInfo(req.CommID.OperationID, "rpc AddFriendResponse ok")
	return &pbFriend.AddFriendResponseResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) DeleteFriend(ctx context.Context, req *pbFriend.DeleteFriendReq) (*pbFriend.DeleteFriendResp, error) {
	log.NewInfo(req.CommID.OperationID, "DeleteFriend args ", req.String())
	// Parse token, to find current user information
	err := imdb.DeleteSingleFriendInfo(req.CommID.FromUserID, req.CommID.ToUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "DeleteSingleFriendInfo failed", err.Error(), req.CommID.FromUserID, req.CommID.ToUserID)
		return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil
	}
	log.NewInfo(req.CommID.OperationID, "DeleteFriend rpc ok")
	reduceFriendFromCache := &pbCache.ReduceFriendFromCacheReq{OperationID: req.CommID.OperationID, UserID: req.CommID.FromUserID, FriendID: req.CommID.ToUserID}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.CommID.OperationID)
	if etcdConn == nil {
		errMsg := req.CommID.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.CommID.OperationID, errMsg)
		return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}
	client := pbCache.NewCacheClient(etcdConn)
	respPb, err := client.ReduceFriendFromCache(context.Background(), reduceFriendFromCache)
	if err != nil {
		log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache rpc failed", err.Error())
		return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrServer.ErrCode, ErrMsg: constant.ErrServer.ErrMsg}}, nil
	}

	log.Debug(req.CommID.OperationID, "Delete user req.CommID.FromUserID ", req.CommID.FromUserID, " To user: ", req.CommID.ToUserID)
	go func() {
		if err := db.DB.DelUserRemarkbyFriend(db.RemarksByUserFriends, req.CommID.FromUserID, req.CommID.ToUserID); err != nil {
			log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "delete remark failed, Insertion user remark failed", err.Error())
		}
	}()
	if respPb.CommonResp.ErrCode != 0 {
		log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "AddFriendToCache failed")
		return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{ErrCode: respPb.CommonResp.ErrCode, ErrMsg: respPb.CommonResp.ErrMsg}}, nil
	}
	chat.FriendDeletedNotification(req)
	return &pbFriend.DeleteFriendResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) GetBlacklist(ctx context.Context, req *pbFriend.GetBlacklistReq) (*pbFriend.GetBlacklistResp, error) {
	log.Debug(req.CommID.OperationID, "GetBlacklist args ", req.String())

	// Parse token, to find current user information
	blackListInfo, err := imdb.GetBlackListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetBlackListByUID failed ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetBlacklistResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	var (
		userInfoList []*sdkws.PublicUserInfo
	)
	for _, blackUser := range blackListInfo {
		var blackUserInfo sdkws.PublicUserInfo
		// Find black user information
		us, err := imdb.GetUserByUserID(blackUser.BlockUserID)
		if err != nil {
			log.NewError(req.CommID.OperationID, "GetUserByUserID failed ", err.Error(), blackUser.BlockUserID)
			continue
		}
		utils.CopyStructFields(&blackUserInfo, us)
		userInfoList = append(userInfoList, &blackUserInfo)
	}
	log.Debug(req.CommID.OperationID, "rpc GetBlacklist ok ", pbFriend.GetBlacklistResp{BlackUserInfoList: userInfoList})
	return &pbFriend.GetBlacklistResp{BlackUserInfoList: userInfoList}, nil
}

func (s *friendServer) SetFriendRemark(ctx context.Context, req *pbFriend.SetFriendRemarkReq) (*pbFriend.SetFriendRemarkResp, error) {
	log.NewInfo(req.CommID.OperationID, "SetFriendComment args ", req.String())
	// Parse token, to find current user information
	err := imdb.UpdateFriendComment(req.CommID.FromUserID, req.CommID.ToUserID, req.Remark)
	if err != nil {
		log.NewError(req.CommID.OperationID, "UpdateFriendComment failed ", req.CommID.FromUserID, req.CommID.ToUserID, req.Remark)
		return &pbFriend.SetFriendRemarkResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	chat.FriendRemarkSetNotification(req.CommID.OperationID, req.CommID.OpUserID, req.CommID.FromUserID, req.CommID.ToUserID)
	// set user remark in Redis for use it in chat
	if err := db.DB.SetUserRemarkbyFriend(db.RemarksByUserFriends, req.CommID.OpUserID, req.CommID.ToUserID, req.Remark); err != nil {
		log.NewError(req.CommID.OperationID, utils.GetSelfFuncName(), "cache failed, Insertion user remark failed", err.Error())
	}

	return &pbFriend.SetFriendRemarkResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) GetFriendRemarkOrNick(ctx context.Context, req *pbFriend.GetFriendRemarkOrNickReq) (*pbFriend.GetFriendRemarkOrNickResp, error) {
	log.NewInfo(req.OperationID, "GetFriendRemarkOrNick args ", req.String())

	// set user remark in Redis for use it in chat
	var remark string
	var nickName string
	var err error
	remark, err = db.DB.GetUserRemarkbyFriend(db.RemarksByUserFriends, req.OpUserID, req.ForUserID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, Remark not found in Redis", err.Error())
		remark, err = imdb.GetFriendRemarkByUserID(req.OpUserID, req.ForUserID)
		log.Debug(req.OperationID, "op user ID: ", req.OpUserID, " for user ID: ", req.ForUserID, " remark ", remark)
		if err != nil {
			log.NewError(req.OperationID, "UpdateFriendComment failed ", req.ForUserID, req.GroupID)
		}
		if remark == "" {
			userName, _ := imdb.GetUserNameByUserID(req.ForUserID)
			log.Debug(req.OperationID, "userName ", userName)
			if userName != "" {
				remark = userName
			}
		}
		// set user remark in Redis for use it in chat
		go func() {
			if err := db.DB.SetUserRemarkbyFriend(db.RemarksByUserFriends, req.OpUserID, req.ForUserID, remark); err != nil {
				log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, Insertion user remark failed", err.Error())
			}
		}()
	}
	log.Debug(req.OperationID, "SELECT FROM MYSQL OVER remark: ", remark, " req.GroupID：", req.GroupID)
	if remark == "" && req.GroupID != "" {
		nickName, err = db.DB.GetUserNickNameByGroup(db.NickNameByGroupID, req.GroupID, req.ForUserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, NickName not found in Redis", err.Error())
			groupMember, err := imdb.GetGroupMemberByUserIDGroupID(req.GroupID, req.ForUserID)
			if err != nil {
				log.NewError(req.OperationID, "UpdateFriendComment failed ", req.ForUserID, req.GroupID)
			}
			nickName = groupMember.Nickname
			// set user remark in Redis for use it in chat
			go func() {
				if err := db.DB.SetUserNickNameByGroup(db.NickNameByGroupID, req.GroupID, req.ForUserID, nickName); err != nil {
					log.NewError(req.OperationID, utils.GetSelfFuncName(), "cache failed, Insertion user remark failed", err.Error())
				}
			}()
			remark = nickName
		}
	}
	log.NewInfo(req.OperationID, remark)
	return &pbFriend.GetFriendRemarkOrNickResp{ErrCode: constant.OK.Code(), ErrMsg: constant.OK.ErrMsg, RemarkNickName: remark}, nil
}

func (s *friendServer) RemoveBlacklist(ctx context.Context, req *pbFriend.RemoveBlacklistReq) (*pbFriend.RemoveBlacklistResp, error) {
	log.NewInfo(req.CommID.OperationID, "RemoveBlacklist args ", req.String())
	// Parse token, to find current user information
	err := imdb.RemoveBlackList(req.CommID.FromUserID, req.CommID.ToUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "RemoveBlackList failed", err.Error(), req.CommID.FromUserID, req.CommID.ToUserID)
		return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil

	}
	log.NewInfo(req.CommID.OperationID, "rpc RemoveBlacklist ok ")
	reqReduceBlackUserFromCache := &pbCache.ReduceBlackUserFromCacheReq{UserID: req.CommID.FromUserID, BlackUserID: req.CommID.ToUserID, OperationID: req.CommID.OperationID}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.CommID.OperationID)
	if etcdConn == nil {
		errMsg := req.CommID.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.CommID.OperationID, errMsg)
		return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}

	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.ReduceBlackUserFromCache(context.Background(), reqReduceBlackUserFromCache)
	if err != nil {
		log.NewError(req.CommID.OperationID, "ReduceBlackUserFromCache rpc call failed ", err.Error())
		return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: 500, ErrMsg: "ReduceBlackUserFromCache rpc call failed"}}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.CommID.OperationID, "ReduceBlackUserFromCache rpc logic call failed ", cacheResp.String())
		return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{ErrCode: cacheResp.CommonResp.ErrCode, ErrMsg: cacheResp.CommonResp.ErrMsg}}, nil
	}
	chat.BlackDeletedNotification(req)
	return &pbFriend.RemoveBlacklistResp{CommonResp: &pbFriend.CommonResp{}}, nil
}

func (s *friendServer) IsInBlackList(ctx context.Context, req *pbFriend.IsInBlackListReq) (*pbFriend.IsInBlackListResp, error) {
	log.NewInfo("IsInBlackList args ", req.String())
	var isInBlacklist = false
	err := imdb.CheckBlack(req.CommID.FromUserID, req.CommID.ToUserID)
	if err == nil {
		isInBlacklist = true
	}
	log.NewInfo(req.CommID.OperationID, "IsInBlackList rpc ok ", pbFriend.IsInBlackListResp{Response: isInBlacklist})
	return &pbFriend.IsInBlackListResp{Response: isInBlacklist}, nil
}

func (s *friendServer) IsFriend(ctx context.Context, req *pbFriend.IsFriendReq) (*pbFriend.IsFriendResp, error) {
	log.NewInfo(req.CommID.OperationID, req.String())
	var isFriend bool
	_, err := imdb.GetFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
	if err == nil {
		isFriend = true
	} else {
		isFriend = false
	}
	log.NewInfo(req.CommID.OperationID, pbFriend.IsFriendResp{Response: isFriend})
	return &pbFriend.IsFriendResp{Response: isFriend}, nil
}

func (s *friendServer) GetFriendList(ctx context.Context, req *pbFriend.GetFriendListReq) (*pbFriend.GetFriendListResp, error) {
	log.NewInfo("GetFriendList args ", req.String())

	friends, err := imdb.GetFriendListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "FindUserInfoFromFriend failed ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetFriendListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	var userInfoList []*sdkws.FriendInfo
	for _, friendUser := range friends {
		friendUserInfo := sdkws.FriendInfo{FriendUser: &sdkws.UserInfo{}}
		cp.FriendDBCopyOpenIM(&friendUserInfo, &friendUser)
		log.NewDebug(req.CommID.OperationID, "friends : ", friendUser, "openim friends: ", friendUserInfo)
		userInfoList = append(userInfoList, &friendUserInfo)
	}
	log.NewInfo(req.CommID.OperationID, "rpc GetFriendList ok", pbFriend.GetFriendListResp{FriendInfoList: userInfoList})
	return &pbFriend.GetFriendListResp{FriendInfoList: userInfoList}, nil
}

// received
func (s *friendServer) GetFriendApplyList(ctx context.Context, req *pbFriend.GetFriendApplyListReq) (*pbFriend.GetFriendApplyListResp, error) {
	log.NewInfo(req.CommID.OperationID, "GetFriendApplyList args ", req.String())
	// Parse token, to find current user information
	//	Find the  current user friend applications received
	ApplyUsersInfo, err := imdb.GetReceivedFriendsApplicationListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetReceivedFriendsApplicationListByUserID ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetFriendApplyListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}

	var appleUserList []*sdkws.FriendRequest
	for _, applyUserInfo := range ApplyUsersInfo {
		var userInfo sdkws.FriendRequest
		cp.FriendRequestDBCopyOpenIM(&userInfo, &applyUserInfo)
		//	utils.CopyStructFields(&userInfo, applyUserInfo)
		//	u, err := imdb.GetUserByUserID(userInfo.FromUserID)
		//	if err != nil {
		//		log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.FromUserID)
		//		continue
		//	}
		//	userInfo.FromNickname = u.Nickname
		//	userInfo.FromFaceURL = u.FaceURL
		//	userInfo.FromGender = u.Gender
		//
		//	u, err = imdb.GetUserByUserID(userInfo.ToUserID)
		//	if err != nil {
		//		log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.ToUserID)
		//		continue
		//	}
		//	userInfo.ToNickname = u.Nickname
		//	userInfo.ToFaceURL = u.FaceURL
		//	userInfo.ToGender = u.Gender
		appleUserList = append(appleUserList, &userInfo)
	}

	log.NewInfo(req.CommID.OperationID, "rpc GetFriendApplyList ok", pbFriend.GetFriendApplyListResp{FriendRequestList: appleUserList})
	return &pbFriend.GetFriendApplyListResp{FriendRequestList: appleUserList}, nil
}

func (s *friendServer) GetSelfApplyList(ctx context.Context, req *pbFriend.GetSelfApplyListReq) (*pbFriend.GetSelfApplyListResp, error) {
	log.NewInfo(req.CommID.OperationID, "GetSelfApplyList args ", req.String())
	//	Find the self add other userinfo
	usersInfo, err := imdb.GetSendFriendApplicationListByUserID(req.CommID.FromUserID)
	if err != nil {
		log.NewError(req.CommID.OperationID, "GetSendFriendApplicationListByUserID failed ", err.Error(), req.CommID.FromUserID)
		return &pbFriend.GetSelfApplyListResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}, nil
	}
	var selfApplyOtherUserList []*sdkws.FriendRequest
	for _, selfApplyOtherUserInfo := range usersInfo {
		var userInfo sdkws.FriendRequest // pbFriend.ApplyUserInfo
		cp.FriendRequestDBCopyOpenIM(&userInfo, &selfApplyOtherUserInfo)
		// u, err := imdb.GetUserByUserID(userInfo.FromUserID)
		// if err != nil {
		//	log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.FromUserID)
		//	continue
		// }
		// userInfo.FromNickname = u.Nickname
		// userInfo.FromFaceURL = u.FaceURL
		// userInfo.FromGender = u.Gender
		//
		// u, err = imdb.GetUserByUserID(userInfo.ToUserID)
		// if err != nil {
		//	log.Error(req.CommID.OperationID, "GetUserByUserID", userInfo.ToUserID)
		//	continue
		// }
		// userInfo.ToNickname = u.Nickname
		// userInfo.ToFaceURL = u.FaceURL
		// userInfo.ToGender = u.Gender

		selfApplyOtherUserList = append(selfApplyOtherUserList, &userInfo)
	}
	log.NewInfo(req.CommID.OperationID, "rpc GetSelfApplyList ok", pbFriend.GetSelfApplyListResp{FriendRequestList: selfApplyOtherUserList})
	return &pbFriend.GetSelfApplyListResp{FriendRequestList: selfApplyOtherUserList}, nil
}

func (s *friendServer) GetFriendsInfo(ctx context.Context, req *pbFriend.GetFriendsInfoReq) (*pbFriend.GetFriendsInfoResp, error) {
	log.NewInfo(req.OperationID, "GetFriendsInfo args ", req.String())

	if req.OpUserID == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "req.OpUserID is nil")
		return &pbFriend.GetFriendsInfoResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "OpUserID is nil"}, nil
	}

	if req.FriendUserIDs == nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "FriendUserIDs is nil")
		return &pbFriend.GetFriendsInfoResp{ErrCode: constant.ErrArgs.ErrCode, ErrMsg: "FriendUserIDs is nil"}, nil
	}

	resp := pbFriend.GetFriendsInfoResp{}

	friendList := []*sdkws.FriendInfo{}
	for _, friendUserID := range req.FriendUserIDs {
		friendRelation, err := imdb.GetFriendRelationshipFromFriend(req.OpUserID, friendUserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "friendRelation is nil", req.OpUserID, friendUserID, err.Error())
			continue
		}
		friendInfo, err := imdb.GetUserByUserID(friendRelation.FriendUserID)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "friendInfo is nil", req.OpUserID, friendUserID, err.Error())
			continue
		}

		newFriendInfo := sdkws.FriendInfo{FriendUser: &sdkws.UserInfo{
			UserID:      friendInfo.UserID,
			Nickname:    friendInfo.Nickname,
			FaceURL:     friendInfo.FaceURL,
			Gender:      friendInfo.Gender,
			PhoneNumber: friendInfo.PhoneNumber,
			Birth:       utils.GetTimeStringFromTime(friendInfo.Birth),
			CreateTime:  int32(friendInfo.CreateTime),
		}}
		cp.FriendDBCopyOpenIM(&newFriendInfo, friendRelation)
		err = utils.CopyStructFields(&newFriendInfo.FriendUser, &friendInfo)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "CopyStructFields failed", req.OpUserID, friendUserID, err.Error())
			continue
		}
		friendList = append(friendList, &newFriendInfo)
	}
	resp.FriendInfoList = friendList

	log.NewInfo(req.OperationID, "GetFriendsInfo ok ", resp)
	return &resp, nil

}

func syncFriendRequestToLocal(operationID, fromUserID, toUserID, msgType string) error {
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)
	friendReqInfo, err := imdb.GetFriendApplicationByBothUserID(fromUserID, toUserID)
	if err != nil {
		log.NewError(operationID, "GetFriendApplicationByBothUserID failed")
		return err
	}

	localFriendReq := sdkws.FriendRequest{}
	err = utils.CopyStructFields(&localFriendReq, &friendReqInfo)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}

	fromUser, err := imdb.GetUserByUserID(fromUserID)
	if err == nil {
		localFriendReq.FromGender = fromUser.Gender
		localFriendReq.FromFaceURL = fromUser.FaceURL
		localFriendReq.FromNickname = fromUser.Nickname
		localFriendReq.FromPhoneNumber = fromUser.PhoneNumber
		localFriendReq.FromEmail = fromUser.Email
	}

	toUser, err := imdb.GetUserByUserID(toUserID)
	if err == nil {
		localFriendReq.ToGender = toUser.Gender
		localFriendReq.ToFaceURL = toUser.FaceURL
		localFriendReq.ToNickname = toUser.Nickname
		localFriendReq.ToPhoneNumber = toUser.PhoneNumber
		localFriendReq.ToEmail = toUser.Email
	}

	fromUserIDList := []string{fromUserID}
	toUserIDList := []string{toUserID}

	syncFriendReq := &local_database.SyncDataReq{
		OperationID:   operationID,
		MsgType:       msgType,
		MemberIDList:  toUserIDList,
		FriendRequest: &localFriendReq,
	}
	localFriendResp, err := localDataClient.SyncData(context.Background(), syncFriendReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}

	syncSelfFriendReq := &local_database.SyncDataReq{
		OperationID:   operationID,
		MsgType:       constant.SyncSelfFriendRequest,
		MemberIDList:  fromUserIDList,
		FriendRequest: &localFriendReq,
	}
	localFriendResp, err = localDataClient.SyncData(context.Background(), syncSelfFriendReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}

	if localFriendResp.ErrCode != 0 {
		log.NewError(operationID, "SyncData rpc logic call failed ", localFriendResp.String())
		return errors.New("SyncData rpc logic call failed")
	}
	return nil
}

func syncFriendInfoToLocal(operationID, ownerID, friendID, msgType string) error {
	//data synchronization
	etcdLocalDataConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImLocalDataName, operationID)
	if etcdLocalDataConn == nil {
		errMsg := operationID + "etcdLocalDataConn == nil"
		log.NewError(operationID, errMsg)
		return errors.New(errMsg)
	}
	localDataClient := local_database.NewLocalDataBaseClient(etcdLocalDataConn)
	friendInfo, err := imdb.GetFriendRelationshipFromFriend(ownerID, friendID)
	if err != nil {
		log.NewError(operationID, "GetFriendRelationshipFromFriend failed")
		return err
	}

	localFriend := sdkws.FriendInfo{}
	err = utils.CopyStructFields(&localFriend, &friendInfo)
	if err != nil {
		log.NewError(operationID, "CopyStructFields failed")
		return err
	}

	friendUserInfo, err := imdb.GetUserByUserID(friendID)
	if err == nil {
		userInfo := sdkws.UserInfo{}
		err = utils.CopyStructFields(&userInfo, &friendUserInfo)
		if err != nil {
			log.NewError(operationID, "CopyStructFields failed")
			return err
		}
		localFriend.FriendUser = &userInfo
	}

	toUserIDList := []string{ownerID, friendID}
	syncFriendReq := &local_database.SyncDataReq{
		OperationID:  operationID,
		MsgType:      msgType,
		MemberIDList: toUserIDList,
		FriendInfo:   &localFriend,
	}
	localFriendResp, err := localDataClient.SyncData(context.Background(), syncFriendReq)
	if err != nil {
		log.NewError(operationID, "SyncData rpc call failed", err.Error())
		return err
	}

	if localFriendResp.ErrCode != 0 {
		log.NewError(operationID, "SyncData rpc logic call failed ", localFriendResp.String())
		return errors.New("SyncData rpc logic call failed")
	}
	return nil
}

func sendTextMsgToFriend(sendUid, RecUid string, text string, operationID string) bool {

	sendUserInfo, _ := imdb.GetUserByUserID(sendUid)

	params := api.ManagementSendMsgReq{}
	params.OperationID = operationID
	params.SendID = sendUid
	params.SenderNickname = sendUserInfo.Nickname
	params.SenderFaceURL = sendUserInfo.FaceURL
	params.RecvID = RecUid
	params.ContentType = constant.Text
	params.Content = map[string]interface{}{
		"text": text,
	}

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg, sendUid, RecUid, text)
		return false
	}

	client := pbChat.NewChatClient(etcdConn)
	pbData := newUserSendMsgReq(&params)
	pbData.MsgData.SessionType = constant.SingleChatType
	RpcResp, err := client.SendMsg(context.Background(), pbData)

	log.NewInfo(params.OperationID, "sendTextMsgToFriend", sendUid, RecUid, text, RpcResp, err)
	return true
}

func newUserSendMsgReq(params *api.ManagementSendMsgReq) *pbChat.SendMsgReq {
	var newContent string
	var err error
	switch params.ContentType {
	case constant.Text:
		newContent = params.Content["text"].(string)
	case constant.Picture:
		fallthrough
	case constant.Custom:
		fallthrough
	case constant.Voice:
		fallthrough
	case constant.Video:
		fallthrough
	case constant.File:
		newContent = utils.StructToJsonString(params.Content)
	case constant.Revoke:
		newContent = params.Content["revokeMsgClientID"].(string)
	default:
	}
	options := make(map[string]bool, 5)
	if params.IsOnlineOnly {
		utils.SetSwitchFromOptions(options, constant.IsOfflinePush, false)
		utils.SetSwitchFromOptions(options, constant.IsHistory, false)
		utils.SetSwitchFromOptions(options, constant.IsPersistent, false)
		utils.SetSwitchFromOptions(options, constant.IsSenderSync, false)
	}
	pbData := pbChat.SendMsgReq{
		OperationID: params.OperationID,
		MsgData: &open_im_sdk.MsgData{
			SendID:           params.SendID,
			RecvID:           params.RecvID,
			GroupID:          params.GroupID,
			ClientMsgID:      utils.GetMsgID(params.SendID),
			SenderPlatformID: params.SenderPlatformID,
			SenderNickname:   params.SenderNickname,
			SenderFaceURL:    params.SenderFaceURL,
			SessionType:      params.SessionType,
			MsgFrom:          constant.SysMsgType,
			ContentType:      params.ContentType,
			Content:          []byte(newContent),
			//	ForceList:        params.ForceList,
			CreateTime:      utils.GetCurrentTimestampByMill(),
			Options:         options,
			OfflinePushInfo: params.OfflinePushInfo,
		},
	}
	if params.ContentType == constant.OANotification {
		var tips open_im_sdk.TipsComm
		tips.JsonDetail = utils.StructToJsonString(params.Content)
		pbData.MsgData.Content, err = proto.Marshal(&tips)
		if err != nil {
			log.Error(params.OperationID, "Marshal failed ", err.Error(), tips.String())
		}
	}
	return &pbData
}

// //
// func (s *friendServer) GetFriendsInfo(ctx context.Context, req *pbFriend.GetFriendsInfoReq) (*pbFriend.GetFriendInfoResp, error) {
//	return nil, nil
// //	log.NewInfo(req.CommID.OperationID, "GetFriendsInfo args ", req.String())
// //	var (
// //		isInBlackList int32
// //		//	isFriend      int32
// //		comment string
// //	)
// //
// //	friendShip, err := imdb.FindFriendRelationshipFromFriend(req.CommID.FromUserID, req.CommID.ToUserID)
// //	if err != nil {
// //		log.NewError(req.CommID.OperationID, "FindFriendRelationshipFromFriend failed ", err.Error())
// //		return &pbFriend.GetFriendInfoResp{ErrCode: constant.ErrSearchUserInfo.ErrCode, ErrMsg: constant.ErrSearchUserInfo.ErrMsg}, nil
// //		//	isFriend = constant.FriendFlag
// //	}
// //	comment = friendShip.Remark
// //
// //	friendUserInfo, err := imdb.FindUserByUID(req.CommID.ToUserID)
// //	if err != nil {
// //		log.NewError(req.CommID.OperationID, "FindUserByUID failed ", err.Error())
// //		return &pbFriend.GetFriendInfoResp{ErrCode: constant.ErrSearchUserInfo.ErrCode, ErrMsg: constant.ErrSearchUserInfo.ErrMsg}, nil
// //	}
// //
// //	err = imdb.FindRelationshipFromBlackList(req.CommID.FromUserID, req.CommID.ToUserID)
// //	if err == nil {
// //		isInBlackList = constant.BlackListFlag
// //	}
// //
// //	resp := pbFriend.GetFriendInfoResp{ErrCode: 0, ErrMsg:  "",}
// //
// //	utils.CopyStructFields(resp.FriendInfoList, friendUserInfo)
// //	resp.Data.IsBlack = isInBlackList
// //	resp.Data.OwnerUserID = req.CommID.FromUserID
// //	resp.Data.Remark = comment
// //	resp.Data.CreateTime = friendUserInfo.CreateTime
// //
// //	log.NewInfo(req.CommID.OperationID, "GetFriendsInfo ok ", resp)
// //	return &resp, nil
// //
// }

func (s *friendServer) AddBlackFriends(_ context.Context, req *pbFriend.AddBlackFriendsReq) (*pbFriend.AddBlackFriendsResp, error) {
	log.NewInfo(req.OperationID, "AddBlacklist args ", req.String())

	var blacks = make([]db.Black, len(req.ToUsersID))
	for index, toUser := range req.ToUsersID {
		blacks[index] = db.Black{OwnerUserID: req.FromUserID, BlockUserID: toUser, OperatorUserID: req.OpUserID}
	}
	//TODO did for update moments block user in DB and redis
	go func() {
		var blacksForMoment = make([]db.BlackForMoment, 0)
		for _, toUser := range req.ToUsersID {
			blacksForMoment = append(blacksForMoment, db.BlackForMoment{OwnerUserID: req.FromUserID, BlockUserID: toUser, OperatorUserID: req.OpUserID})
			blacksForMoment = append(blacksForMoment, db.BlackForMoment{OwnerUserID: toUser, BlockUserID: req.FromUserID, OperatorUserID: req.OpUserID})
		}
		err := imdb.InsertUserMomentsBlackList(blacksForMoment)
		if err != nil {
			log.NewError(req.OperationID, "InsertUserBlackList failed ", err.Error())
			return
		}

		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return
		}
		cacheClient := pbCache.NewCacheClient(etcdConn)
		for _, blackMoment := range blacksForMoment {
			reqAddBlackFriendsToCache := &pbCache.AddBlackFriendsToCacheReq{UserID: blackMoment.OwnerUserID, BlackUsersID: []string{blackMoment.BlockUserID}, OperationID: req.OperationID}
			_, err = cacheClient.AddBlackFriendsForMomentsToCache(context.Background(), reqAddBlackFriendsToCache)
			if err != nil {
				log.NewError(req.OperationID, "AddBlackFriendsForMoments ToCache rpc call failed ", err.Error())
				continue
			}
		}

	}()

	err := imdb.InsertUserBlackList(blacks)
	if err != nil {
		log.NewError(req.OperationID, "InsertUserBlackList failed ", err.Error())
		return &pbFriend.AddBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrDB.ErrCode, ErrMsg: constant.ErrDB.ErrMsg}}, nil
	}

	reqAddBlackFriendsToCache := &pbCache.AddBlackFriendsToCacheReq{UserID: req.FromUserID, BlackUsersID: req.ToUsersID, OperationID: req.OperationID}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbFriend.AddBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}
	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.AddBlackFriendsToCache(context.Background(), reqAddBlackFriendsToCache)
	if err != nil {
		log.NewError(req.OperationID, "AddBlackFriendsToCache rpc call failed ", err.Error())
		return &pbFriend.AddBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: 500, ErrMsg: "AddBlackUserToCache rpc call failed"}}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationID, "AddBlackFriendsToCache rpc logic call failed ", cacheResp.String())
		return &pbFriend.AddBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: cacheResp.CommonResp.ErrCode, ErrMsg: cacheResp.CommonResp.ErrMsg}}, nil
	}

	return &pbFriend.AddBlackFriendsResp{CommonResp: &pbFriend.CommonResp{}}, nil

}

func (s *friendServer) GetBlackFriends(_ context.Context, req *pbFriend.GetBlackFriendsReq) (*pbFriend.GetBlackFriendsResp, error) {
	log.NewInfo(req.OperationID, "GetBlacklist args ", req.String())
	resp := &pbFriend.GetBlackFriendsResp{CommonResp: &pbFriend.CommonResp{}}

	// Parse token, to find current user information
	blackListInfo, err := imdb.GetBlackListByUserID(req.FromUserID)
	if err != nil {
		log.NewError(req.OperationID, "GetBlackListByUID failed ", err.Error(), req.FromUserID)
		resp.CommonResp.ErrCode = constant.ErrDB.ErrCode
		resp.CommonResp.ErrMsg = constant.ErrDB.ErrMsg
		return resp, err
	}

	var (
		blackUserInfoList []*pbFriend.BlackUserInfo
	)
	for _, blackUser := range blackListInfo {
		var blackUserInfo pbFriend.BlackUserInfo

		// Find black user information
		us, err := imdb.GetUserByUserID(blackUser.BlockUserID)
		if err != nil {
			log.NewError(req.OperationID, "GetUserByUserID failed ", err.Error(), blackUser.BlockUserID)
			continue
		}
		err = utils.CopyStructFields(&blackUserInfo, us)
		if err != nil {
			log.NewError(req.OperationID, "CopyStructFields failed ", err.Error(), blackUser.BlockUserID)
			continue
		}
		blackUserInfoList = append(blackUserInfoList, &blackUserInfo)
	}

	resp.BlackList = blackUserInfoList

	log.Debug("", "blackUserInfoList len: ", len(blackUserInfoList), " resp.BlackList len: ", len(resp.BlackList))
	return resp, nil
}

func (s *friendServer) RemoveBlackFriends(_ context.Context, req *pbFriend.RemoveBlackFriendsReq) (*pbFriend.RemoveBlackFriendsResp, error) {

	log.NewInfo(req.OperationID, "RemoveBlackFriends args ", req.String())

	// Parse token, to find current user information
	err := imdb.RemoveBlackListUsers(req.FromUserID, req.ToUsersID)
	if err != nil {
		log.NewError(req.OperationID, "RemoveBlackFriends failed", err.Error(), req.FromUserID, req.ToUsersID)
		return &pbFriend.RemoveBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrAccess.ErrCode, ErrMsg: constant.ErrAccess.ErrMsg}}, nil

	}
	//TODO did for remove moments block user in DB and redis
	go func() {
		err := imdb.RemoveMomentsBlackListUsers(req.FromUserID, req.ToUsersID)
		if err != nil {
			log.NewError(req.OperationID, "InsertUserBlackList failed ", err.Error())
		}
		for _, toUser := range req.ToUsersID {
			_ = imdb.RemoveMomentsBlackListUsers(toUser, []string{req.FromUserID})
		}

		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
		if etcdConn == nil {
			errMsg := req.OperationID + "getcdv3.GetConn == nil"
			log.NewError(req.OperationID, errMsg)
			return
		}
		cacheClient := pbCache.NewCacheClient(etcdConn)
		reqAddBlackFriendsToCache := &pbCache.ReduceBlackFriendsFromCacheReq{UserID: req.FromUserID, ToUsersID: req.ToUsersID, OperationID: req.OperationID}
		_, err = cacheClient.ReduceBlackFriendsForMomentsFromCache(context.Background(), reqAddBlackFriendsToCache)
		if err != nil {
			log.NewError(req.OperationID, "AddBlackFriendsForMoments ToCache rpc call failed ", err.Error())
		}
		for _, toUser := range req.ToUsersID {
			reqAddBlackFriendsToCache := &pbCache.ReduceBlackFriendsFromCacheReq{UserID: toUser, ToUsersID: []string{req.FromUserID}, OperationID: req.OperationID}
			_, err = cacheClient.ReduceBlackFriendsForMomentsFromCache(context.Background(), reqAddBlackFriendsToCache)
			if err != nil {
				log.NewError(req.OperationID, "AddBlackFriendsForMoments ToCache rpc call failed ", err.Error())
				continue
			}
		}

	}()

	log.NewInfo(req.OperationID, "rpc RemoveBlackFriends ok ")
	reqReduceBlackFriendsFromCache := &pbCache.ReduceBlackFriendsFromCacheReq{UserID: req.FromUserID, ToUsersID: req.ToUsersID, OperationID: req.OperationID}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbFriend.RemoveBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}

	cacheClient := pbCache.NewCacheClient(etcdConn)
	cacheResp, err := cacheClient.ReduceBlackFriendsFromCache(context.Background(), reqReduceBlackFriendsFromCache)
	if err != nil {
		log.NewError(req.OperationID, "ReduceBlackFriendsFromCache rpc call failed ", err.Error())
		return &pbFriend.RemoveBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: 500, ErrMsg: "ReduceBlackFriendsFromCache rpc call failed"}}, nil
	}
	if cacheResp.CommonResp.ErrCode != 0 {
		log.NewError(req.OperationID, "ReduceBlackFriendsFromCache rpc logic call failed ", cacheResp.String())
		return &pbFriend.RemoveBlackFriendsResp{CommonResp: &pbFriend.CommonResp{ErrCode: cacheResp.CommonResp.ErrCode, ErrMsg: cacheResp.CommonResp.ErrMsg}}, nil
	}

	return &pbFriend.RemoveBlackFriendsResp{CommonResp: &pbFriend.CommonResp{}}, nil

}

func (s *friendServer) GetBlacks(_ context.Context, req *pbFriend.GetBlacksReq) (*pbFriend.GetBlacksResp, error) {
	resp := &pbFriend.GetBlacksResp{}
	resp.Pagination = &sdkws.ResponsePagination{}

	where := make(map[string]interface{}, 0)
	where["owner_user"] = req.OwnerUser
	where["block_user"] = req.BlockUser
	where["remark"] = req.Remark
	where["start_time"] = req.StartTime
	where["end_time"] = req.EndTime

	blackUsers, blackNumbers, err := imdb.GetBlacksByWhere(where, req.Pagination.PageNumber, req.Pagination.ShowNumber, req.OrderBy)
	if err != nil {
		return resp, err
	}

	resp.BlackList = []*pbFriend.BlackListRes{}
	err = utils.CopyStructFields(&resp.BlackList, blackUsers)
	if err != nil {
		return resp, err
	}

	for index, user := range blackUsers {
		resp.BlackList[index].CreateTime = user.CreateTime.Unix()
		resp.BlackList[index].EditUser = user.UpdateUser
		resp.BlackList[index].EditTime = user.UpdateTime
	}

	resp.ListNumber = blackNumbers
	resp.Pagination.ShowNumber = req.Pagination.ShowNumber
	resp.Pagination.CurrentPage = req.Pagination.PageNumber

	return resp, nil
}

func (s *friendServer) RemoveBlack(_ context.Context, req *pbFriend.RemoveBlackReq) (*pbFriend.RemoveBlackResp, error) {
	resp := &pbFriend.RemoveBlackResp{CommResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: ""}}
	friendMap := make(map[string][]string)
	for _, relation := range req.FriendList {
		if _, ok := friendMap[relation.OwnerID]; ok {
			friendMap[relation.OwnerID] = []string{}
		}
		friendMap[relation.OwnerID] = append(friendMap[relation.OwnerID], relation.BlackID)
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
	if etcdConn == nil {
		errMsg := req.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		return &pbFriend.RemoveBlackResp{CommResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: errMsg}}, nil
	}
	cacheClient := pbCache.NewCacheClient(etcdConn)
	for k, v := range friendMap {
		err := imdb.RemoveBlackListUsers(k, v)
		if err != nil {
			log.NewError("", "remove black error, owner:", k, " black user: ", v)
			continue
		}
		reqAddBlackFriendsToCache := &pbCache.ReduceBlackFriendsFromCacheReq{UserID: k, ToUsersID: v, OperationID: req.OperationID}
		cacheResp, err := cacheClient.ReduceBlackFriendsFromCache(context.Background(), reqAddBlackFriendsToCache)
		if err != nil {
			log.NewError(req.OperationID, "ReduceBlackFriendsFromCache rpc call failed ", err.Error())
			continue
		}
		if cacheResp.CommonResp.ErrCode != 0 {
			log.NewError(req.OperationID, "ReduceBlackFriendsFromCache rpc logic call failed ", cacheResp.String())
			continue
		}

		//TODO did for remove moments block user in DB and redis
		go func() {
			err := imdb.RemoveMomentsBlackListUsers(k, v)
			if err != nil {
				log.NewError(req.OperationID, "InsertUserBlackList failed ", err.Error())
			}
			for _, toUser := range v {
				_ = imdb.RemoveMomentsBlackListUsers(toUser, []string{k})
			}

			etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImCacheName, req.OperationID)
			if etcdConn == nil {
				errMsg := req.OperationID + "getcdv3.GetConn == nil"
				log.NewError(req.OperationID, errMsg)
				return
			}
			cacheClient := pbCache.NewCacheClient(etcdConn)
			reqAddBlackFriendsToCache := &pbCache.ReduceBlackFriendsFromCacheReq{UserID: k, ToUsersID: v, OperationID: req.OperationID}
			_, err = cacheClient.ReduceBlackFriendsForMomentsFromCache(context.Background(), reqAddBlackFriendsToCache)
			if err != nil {
				log.NewError(req.OperationID, "AddBlackFriendsForMoments ToCache rpc call failed ", err.Error())
			}
			for _, toUser := range v {
				reqAddBlackFriendsToCache := &pbCache.ReduceBlackFriendsFromCacheReq{UserID: toUser, ToUsersID: []string{k}, OperationID: req.OperationID}
				_, err = cacheClient.ReduceBlackFriendsForMomentsFromCache(context.Background(), reqAddBlackFriendsToCache)
				if err != nil {
					log.NewError(req.OperationID, "AddBlackFriendsForMoments ToCache rpc call failed ", err.Error())
					continue
				}
			}

		}()
	}

	resp.CommResp.ErrMsg = "remove success"
	return resp, nil
}

func (s *friendServer) AlterRemark(_ context.Context, req *pbFriend.AlterRemarkReq) (*pbFriend.AlterRemarkResp, error) {
	resp := &pbFriend.AlterRemarkResp{CommResp: &pbFriend.CommonResp{ErrCode: constant.ErrInternal.ErrCode, ErrMsg: ""}}
	err := imdb.AlterBlackRemark(req.OwnerID, req.BlackID, req.Remark, req.OpUserID)
	if err != nil {
		log.NewError("", "alter black remark failed", err.Error())
		return resp, err
	}

	resp.CommResp.ErrMsg = "alter success"
	return resp, nil
}
