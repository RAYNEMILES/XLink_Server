/*
** description("").
** copyright('open-im,www.open-im.io').
** author("fg,Gordon@tuoyun.net").
** time(2021/9/15 15:23).
 */
package manage

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/common/token_verify"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	pbChat "Open_IM/pkg/proto/chat"
	open_im_sdk "Open_IM/pkg/proto/sdk_ws"
	server_api_params "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"image"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"
)

var validate *validator.Validate

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

func newUserSendMsgV2Req(params *api.ManagementSendMsgV2Req, pic, video, snapShot *multipart.FileHeader) *pbChat.SendMegToUsersReq {
	var newContent string
	var err error
	log.Debug("", "params content type:", params.ContentType)
	switch params.ContentType {
	case constant.Text:
		newContent = params.Content["text"].(string)
		if newContent == "" {
			return nil
		}
	case constant.Picture:
		if pic == nil {
			log.Debug("", "pic is nil")
			return nil
		}
		picPath, err := uploadObj(pic)
		if err != nil {
			log.Debug("", "upload pic error, err: ", err.Error())
			return nil
		}
		newContent, err = CreateImageMessageFromFullPath(picPath, pic)
		log.Debug("", "", newContent)
		if err != nil {
			log.Debug("", "CreateImageMessageFromFullPath pic error, err: ", err.Error())
			return nil
		}
	case constant.Custom:
		fallthrough
	case constant.Voice:
		fallthrough
	case constant.Video:
		if video == nil {
			log.Debug("", "video is nil")
			return nil
		}
		videoPath, err := uploadObj(video)
		if err != nil {
			log.Debug("", "upload video error, err: ", err.Error())
			return nil
		}
		snapShotPath, err := uploadObj(snapShot)
		if err != nil {
			return nil
		}
		data := VideoElem{}
		if err := mapstructure.WeakDecode(params.Content["videoElem"], &data); err != nil {
			log.NewError("", "video type error")
			return nil
		} else if err = validate.Struct(data); err != nil {
			log.NewError("", "video type error")
			return nil
		}
		newContent, err = CreateVideoMessageFromFullPath(videoPath, data.VideoType, data.Duration, snapShotPath, video, snapShot)
		log.Debug("", "contenttype video: ", newContent)
		if err != nil {
			log.Debug("", "CreateVideoMessageFromFullPath video error, err: ", err.Error())
			return nil
		}
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

	pbData := pbChat.SendMegToUsersReq{
		OperationID: params.OperationID,
		MsgData: &open_im_sdk.MsgData{
			ClientMsgID:      utils.GetMsgID(params.SendID),
			SendID:           params.SendID,
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
	log.Debug("", "pbData: ", utils.StructToJsonString(pbData))
	return &pbData
}

func init() {
	validate = validator.New()
}

func ManagementSendMsg(c *gin.Context) {
	var data interface{}
	params := api.ManagementSendMsgReq{}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		log.Error(c.PostForm("operationID"), "json unmarshal err", err.Error(), c.PostForm("content"))
		return
	}

	//contentType: 101 Text, 102 picture, 103 voice, 104 video, 105 file, 110 custom, 111 revoke, 1400 OA notification
	switch params.ContentType {
	case constant.Text:
		data = TextElem{}
	case constant.Picture:
		data = PictureElem{}
	case constant.Voice:
		data = SoundElem{}
	case constant.Video:
		data = VideoElem{}
	case constant.File:
		data = FileElem{}
	//case constant.AtText:
	//	data = AtElem{}
	//case constant.Merger:
	//	data =
	//case constant.Card:
	//case constant.Location:
	case constant.Custom:
		data = CustomElem{}
	case constant.Revoke:
		data = RevokeElem{}
	case constant.OANotification:
		data = OANotificationElem{}
		params.SessionType = constant.NotificationChatType
	//case constant.HasReadReceipt:
	//case constant.Typing:
	//case constant.Quote:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 404, "errMsg": "contentType err"})
		log.Error(c.PostForm("operationID"), "contentType err", c.PostForm("content"))
		return
	}
	if err := mapstructure.WeakDecode(params.Content, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": err.Error()})
		log.Error(c.PostForm("operationID"), "content to Data struct  err", err.Error())
		return
	} else if err := validate.Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 403, "errMsg": err.Error()})
		log.Error(c.PostForm("operationID"), "data args validate  err", err.Error())
		return
	}
	log.NewInfo(params.OperationID, data, params)
	token := c.Request.Header.Get("token")
	gAuthTypeToken := false
	_, err := token_verify.ParseToken(token, params.OperationID, gAuthTypeToken)
	if err != nil {
		log.NewError(params.OperationID, "parse token failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "parse token failed", "sendTime": 0, "MsgID": ""})
		return
	}

	switch params.SessionType {
	case constant.SingleChatType:
		if len(params.RecvID) == 0 {
			log.NewError(params.OperationID, "recvID is a null string")
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 405, "errMsg": "recvID is a null string", "sendTime": 0, "MsgID": ""})
			return
		}
	case constant.GroupChatType:
		if len(params.GroupID) == 0 {
			log.NewError(params.OperationID, "groupID is a null string")
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 405, "errMsg": "groupID is a null string", "sendTime": 0, "MsgID": ""})
			return
		}

	}
	log.NewInfo(params.OperationID, "Ws call success to ManagementSendMsgReq", params)

	pbData := newUserSendMsgReq(&params)
	log.Info(params.OperationID, "", "api ManagementSendMsg call start..., [data: %s]", pbData.String())

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, params.OperationID)
	if etcdConn == nil {
		errMsg := params.OperationID + "getcdv3.GetConn == nil"
		log.NewError(params.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}
	client := pbChat.NewChatClient(etcdConn)

	log.Info(params.OperationID, "", "api ManagementSendMsg call, api call rpc...")

	RpcResp, err := client.SendMsg(context.Background(), pbData)
	if err != nil {
		log.NewError(params.OperationID, "call delete UserSendMsg rpc server failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call UserSendMsg  rpc server failed"})
		return
	}
	log.Info(params.OperationID, "", "api ManagementSendMsg call end..., [data: %s] [reply: %s]", pbData.String(), RpcResp.String())
	resp := api.ManagementSendMsgResp{CommResp: api.CommResp{ErrCode: RpcResp.ErrCode, ErrMsg: RpcResp.ErrMsg}, ResultList: server_api_params.UserSendMsgResp{ServerMsgID: RpcResp.ServerMsgID, ClientMsgID: RpcResp.ClientMsgID, SendTime: RpcResp.SendTime}}
	log.Info(params.OperationID, "ManagementSendMsg return", resp)
	c.JSON(http.StatusOK, resp)
}

func ManagementBatchSendMsg(c *gin.Context) {
	var data interface{}
	params := api.ManagementBatchSendMsgReq{}
	resp := api.ManagementBatchSendMsgResp{}
	resp.Data.FailedIDList = make([]string, 0)
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		log.Error(c.PostForm("operationID"), "json unmarshal err", err.Error(), c.PostForm("content"))
		return
	}

	switch params.ContentType {
	case constant.Text:
		data = TextElem{}
	case constant.Picture:
		data = PictureElem{}
	case constant.Voice:
		data = SoundElem{}
	case constant.Video:
		data = VideoElem{}
	case constant.File:
		data = FileElem{}
	//case constant.AtText:
	//	data = AtElem{}
	//case constant.Merger:
	//	data =
	//case constant.Card:
	//case constant.Location:
	case constant.Custom:
		data = CustomElem{}
	case constant.Revoke:
		data = RevokeElem{}
	case constant.OANotification:
		data = OANotificationElem{}
		params.SessionType = constant.NotificationChatType
	//case constant.HasReadReceipt:
	//case constant.Typing:
	//case constant.Quote:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 404, "errMsg": "contentType err"})
		log.Error(c.PostForm("operationID"), "contentType err", c.PostForm("content"))
		return
	}
	if err := mapstructure.WeakDecode(params.Content, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 401, "errMsg": err.Error()})
		log.Error(c.PostForm("operationID"), "content to Data struct  err", err.Error())
		return
	} else if err := validate.Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 403, "errMsg": err.Error()})
		log.Error(c.PostForm("operationID"), "data args validate  err", err.Error())
		return
	}
	log.NewInfo(params.OperationID, data, params)
	token := c.Request.Header.Get("token")
	gAuthTypeToken := false
	_, err := token_verify.ParseToken(token, params.OperationID, gAuthTypeToken)
	if err != nil {
		log.NewError(params.OperationID, "parse token failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "parse token failed", "sendTime": 0, "MsgID": ""})
		return
	}

	log.NewInfo(params.OperationID, "Ws call success to ManagementSendMsgReq", params)
	for _, recvID := range params.RecvIDList {
		pbData := newUserSendMsgReq(&params.ManagementSendMsgReq)
		pbData.MsgData.RecvID = recvID
		log.Info(params.OperationID, "", "api ManagementSendMsg call start..., ", pbData.String())
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, params.OperationID)
		if etcdConn == nil {
			errMsg := params.OperationID + "getcdv3.GetConn == nil"
			log.NewError(params.OperationID, errMsg)
			resp.Data.FailedIDList = append(resp.Data.FailedIDList, recvID)
			continue
		}
		client := pbChat.NewChatClient(etcdConn)
		rpcResp, err := client.SendMsg(context.Background(), pbData)
		if err != nil {
			log.NewError(params.OperationID, "call delete UserSendMsg rpc server failed", err.Error())
			resp.Data.FailedIDList = append(resp.Data.FailedIDList, recvID)
			continue
		}
		if rpcResp.ErrCode != 0 {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "rpc failed", pbData, rpcResp)
			resp.Data.FailedIDList = append(resp.Data.FailedIDList, recvID)
			continue
		}
		resp.Data.ResultList = append(resp.Data.ResultList, server_api_params.UserSendMsgResp{
			ServerMsgID: rpcResp.ServerMsgID,
			ClientMsgID: rpcResp.ClientMsgID,
			SendTime:    rpcResp.SendTime,
		})
	}

	log.NewInfo(params.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

func getTencentCloudCredential() (*sts.CredentialResult, error) {
	cli := sts.NewClient(
		config.Config.Credential.Tencent.SecretID,
		config.Config.Credential.Tencent.SecretKey,
		nil,
	)
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(time.Hour.Seconds()),
		Region:          config.Config.Credential.Tencent.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					Action: []string{
						"name/cos:PostObject",
						"name/cos:PutObject",
						"name/cos:DeleteObject",
						"name/cos:PutObjectTagging",
						"name/cos:DeleteObjectTagging",
					},
					Effect: "allow",
					Resource: []string{
						"qcs::cos:" + config.Config.Credential.Tencent.Region + ":uid/" + config.Config.Credential.Tencent.AppID + ":" + config.Config.Credential.Tencent.Bucket + "/*",
						"qcs::cos:" + config.Config.Credential.Tencent.Region + ":uid/" + config.Config.Credential.Tencent.AppID + ":" + config.Config.Credential.Tencent.PersistenceBucket + "/*",
					},
				},
			},
		},
	}
	return cli.GetCredential(opt)
}

func ManagementBatchSendMsgV2(c *gin.Context) {
	params := api.ManagementBatchSendMsgV2Req{}
	resp := api.ManagementBatchSendMsgV2Resp{}
	resp.Data.FailedIDList = make([]string, 0)
	reqJson := c.PostForm("req_json")
	err := json.Unmarshal([]byte(reqJson), &params)
	if err != nil {
		log.Error(c.PostForm("operationID"), "json unmarshal err", err.Error(), c.PostForm("content"))
		return
	}

	log.Debug("", "params: ", utils.StructToJsonString(params))

	token := c.Request.Header.Get("token")
	_, err = token_verify.ParseToken(token, params.OperationID, false)
	if err != nil {
		log.NewError(params.OperationID, "parse token failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "parse token failed", "sendTime": 0, "MsgID": ""})
		return
	}

	log.NewInfo(params.OperationID, "Ws call success to ManagementSendMsgReq", params)

	pic, err := c.FormFile("image")
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "FormFile failed", err.Error())
	}
	hasVideo := false
	video, err := c.FormFile("video")
	if err != nil {
		log.NewError(params.OperationID, utils.GetSelfFuncName(), "FormFile failed", err.Error())
	} else {
		hasVideo = true
	}
	snapShot, err := c.FormFile("snapShot")
	if err != nil {
		if hasVideo {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "FormFile failed", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file arg: " + err.Error()})
			return
		}
	}

	msgType := []int64{constant.Picture, constant.Video, constant.Text}
	resp.Data.ResultList = []server_api_params.UserSendMsgResp{}
	for _, msgT := range msgType {
		params.ManagementSendMsgV2Req.ContentType = int32(msgT)
		log.Debug("", "msg type: ", msgT)
		pbData := newUserSendMsgV2Req(&params.ManagementSendMsgV2Req, pic, video, snapShot)
		if pbData == nil {
			continue
		}
		pbData.GroupIDList = params.GroupIDList
		pbData.RecvIDList = params.RecvIDList
		log.Info(params.OperationID, "", "api ManagementSendMsg call start..., ", pbData.String())
		etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImOfflineMessageName, params.OperationID)
		if etcdConn == nil {
			errMsg := params.OperationID + "getcdv3.GetConn == nil"
			resp.Data.FailedIDList = append(resp.Data.FailedIDList, params.RecvIDList...)
			log.NewError(params.OperationID, errMsg)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
			return
		}
		client := pbChat.NewChatClient(etcdConn)
		rpcResp, err := client.SendMegToUsers(context.Background(), pbData)
		if err != nil {
			log.NewError(params.OperationID, "call delete UserSendMsg rpc server failed", err.Error())
			if rpcResp != nil {
				resp.Data.FailedIDList = append(resp.Data.FailedIDList, rpcResp.FailedIDList...)
			}
			log.NewError(params.OperationID, "call delete users rpc server failed", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call delete users rpc server failed"})
			return
		}
		if rpcResp.ErrCode != 0 {
			log.NewError(params.OperationID, utils.GetSelfFuncName(), "rpc failed", pbData, rpcResp)
			resp.Data.FailedIDList = append(resp.Data.FailedIDList, rpcResp.FailedIDList...)
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "call send users message rpc server failed"})
			return
		}

		var temp []server_api_params.UserSendMsgResp
		_ = utils.CopyStructFields(&temp, rpcResp.ResultList)
		resp.Data.ResultList = append(resp.Data.ResultList, temp...)
	}

	log.NewInfo(params.OperationID, utils.GetSelfFuncName(), "resp: ", resp)
	c.JSON(http.StatusOK, resp)
}

type PictureBaseInfo struct {
	UUID   string `mapstructure:"uuid"`
	Type   string `mapstructure:"type" `
	Size   int64  `mapstructure:"size" `
	Width  int32  `mapstructure:"width" `
	Height int32  `mapstructure:"height"`
	Url    string `mapstructure:"url" `
}

type PictureElem struct {
	SourcePath      string          `mapstructure:"sourcePath"`
	SourcePicture   PictureBaseInfo `mapstructure:"sourcePicture"`
	BigPicture      PictureBaseInfo `mapstructure:"bigPicture" `
	SnapshotPicture PictureBaseInfo `mapstructure:"snapshotPicture"`
}
type SoundElem struct {
	UUID      string `mapstructure:"uuid"`
	SoundPath string `mapstructure:"soundPath"`
	SourceURL string `mapstructure:"sourceUrl"`
	DataSize  int64  `mapstructure:"dataSize"`
	Duration  int64  `mapstructure:"duration"`
}
type VideoElem struct {
	VideoPath      string `mapstructure:"videoPath"`
	VideoUUID      string `mapstructure:"videoUUID"`
	VideoURL       string `mapstructure:"videoUrl"`
	VideoType      string `mapstructure:"videoType"`
	VideoSize      int64  `mapstructure:"videoSize"`
	Duration       int64  `mapstructure:"duration"`
	SnapshotPath   string `mapstructure:"snapshotPath"`
	SnapshotUUID   string `mapstructure:"snapshotUUID"`
	SnapshotSize   int64  `mapstructure:"snapshotSize"`
	SnapshotURL    string `mapstructure:"snapshotUrl"`
	SnapshotWidth  int32  `mapstructure:"snapshotWidth"`
	SnapshotHeight int32  `mapstructure:"snapshotHeight"`
}
type FileElem struct {
	FilePath  string `mapstructure:"filePath"`
	UUID      string `mapstructure:"uuid"`
	SourceURL string `mapstructure:"sourceUrl"`
	FileName  string `mapstructure:"fileName"`
	FileSize  int64  `mapstructure:"fileSize"`
}
type AtElem struct {
	Text       string   `mapstructure:"text"`
	AtUserList []string `mapstructure:"atUserList"`
	IsAtSelf   bool     `mapstructure:"isAtSelf"`
}
type LocationElem struct {
	Description string  `mapstructure:"description"`
	Longitude   float64 `mapstructure:"longitude"`
	Latitude    float64 `mapstructure:"latitude"`
}
type CustomElem struct {
	Data        string `mapstructure:"data" validate:"required"`
	Description string `mapstructure:"description"`
	Extension   string `mapstructure:"extension"`
}
type TextElem struct {
	Text string `mapstructure:"text" validate:"required"`
}

type RevokeElem struct {
	RevokeMsgClientID string `mapstructure:"revokeMsgClientID" validate:"required"`
}
type OANotificationElem struct {
	NotificationName    string      `mapstructure:"notificationName" json:"notificationName" validate:"required"`
	NotificationFaceURL string      `mapstructure:"notificationFaceURL" json:"notificationFaceURL" validate:"required"`
	NotificationType    int32       `mapstructure:"notificationType" json:"notificationType" validate:"required"`
	Text                string      `mapstructure:"text" json:"text" validate:"required"`
	Url                 string      `mapstructure:"url" json:"url"`
	MixType             int32       `mapstructure:"mixType" json:"mixType"`
	PictureElem         PictureElem `mapstructure:"pictureElem" json:"pictureElem"`
	SoundElem           SoundElem   `mapstructure:"soundElem" json:"soundElem"`
	VideoElem           VideoElem   `mapstructure:"videoElem" json:"videoElem"`
	FileElem            FileElem    `mapstructure:"fileElem" json:"fileElem"`
	Ex                  string      `mapstructure:"ex" json:"ex"`
}

func CreateVideoMessageFromFullPath(videoFullPath string, videoType string, duration int64, snapshotFullPath string, video, snapShot *multipart.FileHeader) (string, error) {
	s := VideoElem{}
	s.VideoPath = videoFullPath
	s.VideoType = videoType
	s.Duration = duration
	if snapshotFullPath == "" {
		s.SnapshotPath = ""
	} else {
		s.SnapshotPath = snapshotFullPath
	}
	s.VideoSize = video.Size
	if snapshotFullPath != "" {
		snapShotObj, err := snapShot.Open()
		if err != nil {
			log.Error("internal", "get Image Attributes error", err)
			return "", err
		}

		img, _, err := image.Decode(snapShotObj)
		if err != nil {
			return "", utils.Wrap(err, "image file  Decode err")
		}
		b := img.Bounds()

		s.SnapshotHeight = int32(b.Max.Y)
		s.SnapshotWidth = int32(b.Max.X)
		s.SnapshotSize = snapShot.Size
	}
	return utils.StructToJsonString(s), nil
}

func CreateImageMessageFromFullPath(imageFullPath string, pic *multipart.FileHeader) (string, error) {
	s := PictureElem{}
	s.SourcePath = imageFullPath
	log.Info("internal", "ImageMessage  path:", s.SourcePath)
	snapShotObj, err := pic.Open()
	if err != nil {
		log.Error("internal", "get Image Attributes error", err)
		return "", err
	}

	img, _, err := image.Decode(snapShotObj)
	if err != nil {
		return "", utils.Wrap(err, "image file  Decode err")
	}
	b := img.Bounds()

	s.SourcePicture.Width = int32(b.Max.X)
	s.SourcePicture.Height = int32(b.Max.Y)
	s.SourcePicture.Size = pic.Size
	s.SourcePicture.Type = pic.Header.Get("Content-Type")

	return utils.StructToJsonString(s), nil
}

func uploadObj(obj *multipart.FileHeader) (path string, err error) {

	dir := ""
	if config.Config.Credential.Tencent.Accelerate {
		dir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.Bucket)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.Bucket, config.Config.Credential.Tencent.Region)
	}
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	COSCredential, err := getTencentCloudCredential()
	if err != nil {
		return path, err
	}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSCredential.Credentials.TmpSecretID,
			SecretKey:    COSCredential.Credentials.TmpSecretKey,
			SessionToken: COSCredential.Credentials.SessionToken,
		},
	})
	if client == nil {
		return path, err
	}

	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
		XOptionHeader: &http.Header{},
	}}
	opt.XOptionHeader.Add("x-cos-tagging", "delete:1")
	objFile, err := obj.Open()
	if err != nil {
		return path, err
	}
	newName, newType := utils.GetNewFileNameAndContentType(obj.Filename, constant.ImageType)
	if newType == "img" {
		opt.ContentType = newType
	}
	_, err = client.Object.Put(context.Background(), newName, objFile, opt)
	if err != nil {
		return path, err
	}

	path = fmt.Sprintf("%s/%s", dir, newName)

	return path, nil

}
