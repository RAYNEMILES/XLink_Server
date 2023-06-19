package channel_code

import (
	utils2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/front_api_struct"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/chat"
	server_api_params2 "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func StartingWelcomeMessages(c *gin.Context) {
	token := c.GetHeader("token")
	_, userID, _ := token_verify.GetUserIDFromToken(token, "")
	if userID == "" {
		log.NewError("", utils2.GetSelfFuncName(), "token is illegal")
		apiResponse := front_api_struct.FrontApiResp{
			ErrCode: constant.ErrArgs.ErrCode,
			ErrMsg:  constant.ErrArgs.ErrMsg,
			Data:    nil,
		}
		c.JSON(int(apiResponse.ErrCode), apiResponse)
		return
	}

	user, err := imdb.GetUserByUserID(userID)
	if err != nil {
		log.NewError("", utils2.GetSelfFuncName(), "user is nil")
		apiResponse := front_api_struct.FrontApiResp{
			ErrCode: constant.ErrArgs.ErrCode,
			ErrMsg:  "user is nil",
			Data:    nil,
		}
		c.JSON(int(apiResponse.ErrCode), apiResponse)
		return
	}

	//check if the channelcode and invitecode is opened
	allowChannel := false
	if imdb.GetChannelCodeIsOpen() {
		allowChannel = true
	}

	allowInvite := false
	if imdb.GetInviteCodeIsOpen() {
		allowInvite = true
	}

	//get channel code information
	//'1:official 2:invite 3:channel'
	data := make(map[string]interface{})
	sourceID := user.SourceId
	sourceCode := user.SourceCode
	var officeChannelCode *db.InviteChannelCode
	var inviteCode *db.InviteCode
	log.NewError("", utils2.GetSelfFuncName(), "sourceID", sourceID, "sourceCode", sourceCode, "allowChannel", allowChannel, "allowInvite", allowInvite)
	switch sourceID {
	case 1:
		if allowChannel {
			//include valid and invalid
			officeChannelCode = imdb.GetOfficialChannelCode()
			if officeChannelCode.State == constant.InviteChannelCodeStateValid {
				data["channel"] = officeChannelCode
				log.NewError("", utils2.GetSelfFuncName(), "officeChannelCode", officeChannelCode)
			}
		}
	case 2:
		if allowInvite {
			inviteCode = imdb.GetCodeInfoByCode(sourceCode)
			if inviteCode.State == constant.InviteCodeStateValid {
				data["invite"] = inviteCode
			}
		}
	case 3:
		if allowChannel {
			//include valid and invalid
			officeChannelCode, err = imdb.GetInviteChannelCodeByCode(sourceCode)
			if err == nil && officeChannelCode.State == constant.InviteChannelCodeStateValid {
				data["channel"] = officeChannelCode
			}
		}
	}

	operationID := utils.OperationIDGenerator()
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, operationID)
	if etcdConn == nil {
		errMsg := operationID + "getcdv3.GetConn == nil"
		log.NewError(operationID, errMsg)
		return
	}
	client := pbChat.NewChatClient(etcdConn)

	myFriendIDList, err := db.DB.GetFriendIDListFromCache(userID)
	myGroupIDList, err := db.DB.GetGroupIDListForUser(userID)

	if officeChannelCode != nil {
		friendList := officeChannelCode.FriendId
		groupList := officeChannelCode.GroupId
		log.NewError(operationID, utils.GetSelfFuncName(), "friendList", friendList, "groupList", groupList)
		if friendList != "" && myFriendIDList != nil && len(myFriendIDList) > 0 {
			friendIDList := strings.Split(friendList, ",")
			for _, friendID := range friendIDList {
				isPass := utils.IsContain(friendID, myFriendIDList)
				if friend, err := imdb.GetUserByUserID(friendID); err == nil && isPass {
					var req pbChat.SendMsgReq
					var msg server_api_params2.MsgData
					req.OperationID = operationID
					req.Token = token
					msg.SendID = friendID
					msg.RecvID = userID
					msg.SenderNickname = friend.Nickname
					msg.SenderFaceURL = friend.FaceURL
					msg.Content = utils.String2Bytes(officeChannelCode.Greeting)
					msg.MsgFrom = constant.UserMsgType
					msg.ContentType = constant.Text
					msg.SessionType = constant.SingleChatType
					msg.SendTime = utils.GetCurrentTimestampByMill()
					msg.CreateTime = utils.GetCurrentTimestampByMill()
					msg.ClientMsgID = utils.GetMsgID(friendID)
					req.MsgData = &msg

					_, err := client.SendMsg(context.Background(), &req)
					if err != nil {
						log.NewError(req.OperationID, "SendMsg rpc failed, ", req.String(), err.Error())
					}
				}
			}

		}
		if groupList != "" && myGroupIDList != nil && len(myGroupIDList) > 0 {
			groupIDList := strings.Split(groupList, ",")
			for _, groupID := range groupIDList {
				isPass := utils.IsContain(groupID, myGroupIDList)
				if owner, err := imdb.GetGroupOwnerInfoByGroupID(groupID); err == nil && isPass {
					var req pbChat.SendMsgReq
					var msg server_api_params2.MsgData
					req.OperationID = operationID
					req.Token = token
					msg.SendID = owner.UserID
					msg.GroupID = groupID
					msg.SenderNickname = owner.Nickname
					msg.SenderFaceURL = owner.FaceURL
					msg.Content = utils.String2Bytes(officeChannelCode.Greeting)
					msg.MsgFrom = constant.UserMsgType
					msg.ContentType = constant.Text
					msg.SessionType = constant.GroupChatType
					msg.SendTime = utils.GetCurrentTimestampByMill()
					msg.CreateTime = utils.GetCurrentTimestampByMill()
					msg.ClientMsgID = utils.GetMsgID(owner.UserID)
					req.MsgData = &msg

					_, err := client.SendMsg(context.Background(), &req)
					if err != nil {
						log.NewError(req.OperationID, "SendMsg rpc failed, ", req.String(), err.Error())
					}
				}
			}
		}

	}

	if inviteCode != nil {
		inviteUserID := inviteCode.UserId
		if inviteUserID != "" && myFriendIDList != nil && len(myFriendIDList) > 0 {
			isPass := utils.IsContain(inviteUserID, myFriendIDList)
			if inviteUser, err := imdb.GetUserByUserID(inviteUserID); err == nil && isPass {
				var req pbChat.SendMsgReq
				var msg server_api_params2.MsgData
				req.OperationID = operationID
				req.Token = token
				msg.SendID = inviteUserID
				msg.RecvID = userID
				msg.SenderNickname = inviteUser.Nickname
				msg.SenderFaceURL = inviteUser.FaceURL
				msg.Content = utils.String2Bytes(inviteCode.Greeting)
				msg.MsgFrom = constant.UserMsgType
				msg.ContentType = constant.Text
				msg.SessionType = constant.SingleChatType
				msg.SendTime = utils.GetCurrentTimestampByMill()
				msg.CreateTime = utils.GetCurrentTimestampByMill()
				msg.ClientMsgID = utils.GetMsgID(inviteUserID)
				req.MsgData = &msg

				_, err := client.SendMsg(context.Background(), &req)
				if err != nil {
					log.NewError(req.OperationID, "SendMsg rpc failed, ", req.String(), err.Error())
				}
			}
		}
	}

	log.NewError("", utils2.GetSelfFuncName(), "data", data)
	c.JSON(http.StatusOK, api.GetInvitionTotalRespone{CommResp: api.CommResp{ErrCode: 0, ErrMsg: ""}, Data: data})
	return
}
