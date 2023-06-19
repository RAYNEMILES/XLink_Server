package apiThird

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func CheckStatus(c *gin.Context) {
	var (
		req  api.CheckStatusReq
	)
	if err := c.Bind(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	var user = ""
	userIDInterface, existed := c.Get("userID")
	if existed {
		user = userIDInterface.(string)
	}

	var canCall = true
	var errMsg = ""
	if req.UserID != "" {
		err := checkCanCommunication(req.UserID, req.GroupID, req.ChatType)
		if err != nil {
			canCall = false
			errMsg = "banned"
			//log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
			//c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": "banned", "data": false})
			//return
		}
		err = checkCanCommunication(user, req.GroupID, req.ChatType)
		if err != nil {
			canCall = false
			errMsg = "banned"
			//log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
			//c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": "banned", "data": false})
			//return
		}
	} else {
		// check group
	}

	log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": errMsg, "data": canCall})

}

func StartCommunication(c *gin.Context) {
	var (
		req  api.StartCommunicationReq
		resp api.StartCommunicationResp
	)
	if err := c.Bind(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	err := checkCanCommunication(userID, req.GroupID, req.ChatType)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "video banned", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": err.Error(), "data": struct{}{}})
		return
	}

	// set been in communication
	err = db.DB.SetUserCommunicationStatus(userID, 1)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrServer, "errMsg": "start communication wrong, server failed", "data": struct{}{}})
		return
	}

	//count, err := db.DB.SetMemberToCommunication(req.RoomID, userID)
	//if err != nil {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetMemberToCommunication failed ", err.Error())
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
	//	return
	//}
	//
	//if count == 0 {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "join failed, count: ", count)
	//	c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": "you have another device is in a communication", "data": struct{}{}})
	//	return
	//}

	// can start, add record to mysql and add user to redis
	record := &db.VideoAudioCommunicationRecord{}
	record.Originator = userID
	record.OriginatorPlatform = req.OriginatorPlatform
	record.Supporter = req.Supporter
	record.GroupID = req.GroupID
	record.RoomID = req.RoomID
	record.RoomIDType = req.RoomIdType
	record.RecordStatus = constant.CommunicationRecordStatusWaiting
	record.ChatType = req.ChatType
	err = imdb.StartCommunication(record)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "InsertCommunication", err.Error(), req)
		c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": err.Error(), "data": struct{}{}})
		return
	}

	if req.Receiver != "" {
		err = imdb.InsertCommunicationMembers(record.CommunicationID, []string{req.Receiver})
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "InsertCommunicationMembers")
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "InsertCommunicationMembers error  " + err.Error()})
			return
		}
	}

	//go func() {
	//	startReq := api.StartRecordReq{}
	//	startReq.OperationID = req.OperationID
	//	startReq.RoomIdType = req.RoomIdType
	//	startReq.RoomId = req.RoomID
	//	err = startRecord(startReq)
	//	if err != nil {
	//		log.NewError("", "start record error, startReq: ", startReq, " err msg: ", err.Error())
	//		return
	//	}
	//}()

	resp.CommResp = api.CommResp{ErrCode: 0, ErrMsg: "start success"}
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": "query and update success", "data": gin.H{
		"communication_id": record.CommunicationID,
	}})

}

func JoinCommunication(c *gin.Context) {
	var (
		req  api.JoinCommunicationReq
		resp api.JoinCommunicationResp
	)

	if err := c.Bind(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}

	err := checkCanCommunication(userID, req.GroupID, int8(req.ChatType))
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "video banned", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": err.Error(), "data": struct{}{}})
		return
	}

	record, err := imdb.GetActiveCommunicationByRoomID(req.RoomID, req.RoomIDType)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetActiveCommunicationByRoomID failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	if record.RecordStatus == constant.CommunicationRecordStatusWaiting {
		record.RecordStatus = constant.CommunicationRecordStatusRecording
		err = imdb.UpdateCommunication(&record)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "update status failed ", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
			return
		}
	}

	err = db.DB.SetUserCommunicationStatus(userID, 1)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrServer, "errMsg": "join communication wrong, server failed", "data": struct{}{}})
		return
	}

	// join calling
	//count, err := db.DB.SetMemberToCommunication(req.RoomID, userID)
	//if err != nil {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "SetMemberToCommunication failed ", err.Error())
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
	//	return
	//}
	//
	//if count == 0 {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "join failed, count: ", count)
	//	c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": "you have another device is in a communication", "data": struct{}{}})
	//	return
	//}

	resp.CommResp = api.CommResp{ErrCode: 0, ErrMsg: "join success"}
	c.JSON(http.StatusOK, resp)

}

func EndCommunication(c *gin.Context) {
	var (
		req  api.EndCommunicationReq
		resp api.EndCommunicationResp
	)

	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	record, err := imdb.GetActiveCommunicationByRoomID(req.RoomId, req.RoomIdType)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetActiveCommunicationByCallID failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	if record.RoomID == "" {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "No this communication")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "No this communication"})
		return
	}

	members, err := db.DB.GetMemberToCommunication(req.RoomId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetMemberToCommunication")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "GetMemberToCommunication error  " + err.Error()})
		return
	}

	err = db.DB.RemoveMemberToCommunication(req.RoomId)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "RemoveMemberToCommunication")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "RemoveMemberToCommunication error  " + err.Error()})
		return
	}
	var mIDList = make([]string, len(members))
	for index, member := range members {
		mIDList[index] = member
		err = db.DB.SetUserCommunicationStatus(member, 0)
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
			c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrServer, "errMsg": "join communication wrong, server failed", "data": struct{}{}})
			return
		}
	}

	startTime := time.Unix(record.StartTime, 0)
	endTime := time.Now()
	record.Duration = int64(endTime.Sub(startTime).Seconds())
	record.EndTime = endTime.Unix()
	record.ErrCode = req.ErrCode

	err = imdb.UpdateCommunication(&record)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateCommunication")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "UpdateCommunication error  " + err.Error()})
		return
	}
	log.Debug("", "record.CommunicationID: ", record.CommunicationID)
	err = imdb.InsertCommunicationMembers(record.CommunicationID, mIDList)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "InsertCommunicationMembers")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "InsertCommunicationMembers error  " + err.Error()})
		return
	}

	//go func() {
	//	err = stopRecord(record.RecordTaskID)
	//	if err != nil {
	//		log.NewError("", "record error, task id: ", record.RecordTaskID, " err msg: ", err.Error())
	//		return
	//	}
	//}()

	resp.CommResp = api.CommResp{ErrCode: 0, ErrMsg: "end success"}
	c.JSON(http.StatusOK, resp)

}

func GetMembersByCommunicationID(c *gin.Context) {
	var (
		req  api.GetMembersByCommunicationIDReq
		resp api.GetMembersByCommunicationIDResp
	)

	if err := c.Bind(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	var communicationID int64 = 0
	var userID string
	userIDInter, existed := c.Get("userID")
	if existed {
		userID = userIDInter.(string)
	}
	err := db.DB.SetUserCommunicationStatus(userID, 0)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetUserCommunicationStatus", req)
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrServer, "errMsg": "join communication wrong, server failed", "data": struct{}{}})
		return
	}

	// row lock
	communicationID, err = imdb.UpdateCallingMember(req.CommunicationID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "UpdateAndGetCallingMember failed", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": "UpdateAndGetCallingMember failed", "data": struct{}{}})
		return
	}

	communication, err := imdb.GetCommunicationById(communicationID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "get failed", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": err.Error(), "data": struct{}{}})
		return
	}
	if communication.CommunicationID == 0 {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "invalid calling")
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": "invalid calling", "data": struct{}{}})
		return
	}

	members, err := imdb.GetCommunicationMemberByCommunicationID(communicationID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "get failed", err.Error())
		c.JSON(http.StatusOK, gin.H{"errCode": constant.ErrAccess.ErrCode, "errMsg": err.Error(), "data": struct{}{}})
		return
	}

	resp.MembersID = members
	resp.MasterID = communication.Originator
	resp.GroupID = communication.GroupID
	log.NewError(req.OperationID, utils.GetSelfFuncName(), "query and update success")
	c.JSON(http.StatusOK, gin.H{"errCode": constant.OK.ErrCode, "errMsg": "query and update success",
		"data": gin.H{
			"master_id":  resp.MasterID,
			"group_id":   resp.GroupID,
			"members_id": resp.MembersID,
		}})
	return
}

//func startRecord(req api.StartRecordReq) error {
//	credential := common.NewCredential(
//		config.Config.Credential.Tencent.SecretID,
//		config.Config.Credential.Tencent.SecretKey,
//	)
//	// 实例化要请求产品的client对象,clientProfile是可选的
//	client, _ := trtc.NewClient(credential, config.Config.Trtc.Record.Region, profile.NewClientProfile())
//
//	// 实例化一个请求对象,每个接口都会对应一个request对象
//	request := trtc.NewCreateCloudRecordingRequest()
//	sdkAppIdUint, _ := strconv.ParseUint(config.Config.Trtc.SdkAppid, 10, 64)
//	recordUserID := req.RoomId + "r"
//	expireTime := 3 * 24 * 60 * 60
//	sig, err := register.GenSig(config.Config.Trtc.SdkAppid, config.Config.Trtc.SecretKey, recordUserID, expireTime, nil)
//	if err != nil {
//		return err
//	}
//
//	request.SdkAppId = common.Uint64Ptr(sdkAppIdUint)
//	request.RoomId = common.StringPtr(req.RoomId)
//	request.RoomIdType = common.Uint64Ptr(req.RoomIdType)
//	request.UserId = common.StringPtr(recordUserID)
//	request.UserSig = common.StringPtr(sig)
//	request.RecordParams = &trtc.RecordParams{
//		MaxIdleTime: common.Uint64Ptr(config.Config.Trtc.Record.MaxIdleTime),
//		StreamType:  common.Uint64Ptr(config.Config.Trtc.Record.StreamType),
//		RecordMode:  common.Uint64Ptr(config.Config.Trtc.Record.RecordMode),
//	}
//	request.StorageParams = &trtc.StorageParams{
//		CloudStorage: &trtc.CloudStorage{
//			Vendor:    common.Uint64Ptr(config.Config.Trtc.Record.Vendor),
//			Region:    common.StringPtr(config.Config.Credential.Tencent.Region),
//			Bucket:    common.StringPtr(config.Config.Credential.Tencent.Bucket),
//			AccessKey: common.StringPtr(config.Config.Credential.Tencent.SecretID),
//			SecretKey: common.StringPtr(config.Config.Credential.Tencent.SecretKey),
//		},
//	}
//	request.MixLayoutParams = &trtc.MixLayoutParams {
//		MixLayoutMode: common.Uint64Ptr(1),
//	}
//
//	// 返回的resp是一个CreateCloudRecordingResponse的实例，与请求对象对应
//	response, err := client.CreateCloudRecording(request)
//	if _, ok := err.(*tencentErrors.TencentCloudSDKError); ok {
//		fmt.Printf("An API error has returned: %s", err)
//		log.NewError("", "record error, tencent sdk error")
//		return err
//	}
//	if err != nil {
//		return err
//	}
//	record := &db.VideoAudioCommunicationRecord{}
//	record.RoomID = req.RoomId
//	record.RoomIDType = req.RoomIdType
//	record.RecordTaskID = *response.Response.TaskId
//	record.RecordRequestID = *response.Response.RequestId
//	record.RecordUserID = recordUserID
//
//	err = imdb.UpdateCommunicationRecordByActiveRoomID(record)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func stopRecord(taskId string) error {
//
//	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
//	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。密钥可前往官网控制台 https://console.tencentcloud.com/capi 进行获取
//	credential := common.NewCredential(
//		config.Config.Credential.Tencent.SecretID,
//		config.Config.Credential.Tencent.SecretKey,
//	)
//
//	// 实例化要请求产品的client对象,clientProfile是可选的
//	client, _ := trtc.NewClient(credential, config.Config.Trtc.Record.Region, profile.NewClientProfile())
//
//	// 实例化一个请求对象,每个接口都会对应一个request对象
//	request := trtc.NewDeleteCloudRecordingRequest()
//
//	sdkAppIdUint, _ := strconv.ParseUint(config.Config.Trtc.SdkAppid, 10, 64)
//	request.SdkAppId = common.Uint64Ptr(sdkAppIdUint)
//	request.TaskId = common.StringPtr(taskId)
//
//	// 返回的resp是一个DeleteCloudRecordingResponse的实例，与请求对象对应
//	_, err := client.DeleteCloudRecording(request)
//	if _, ok := err.(*tencentErrors.TencentCloudSDKError); ok {
//		return err
//	}
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func RecordCallBack(c *gin.Context) {
	var (
		req api.RecordCallBackReq
	)
	if err := c.BindJSON(&req); err != nil {
		log.NewError("callback", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	// cloud record
	if req.EventGroupId == 3 {
		if req.EventType == 310 {
			// finish and upload the mp4 file
			if req.EventInfo.Payload.Status == 0 {
				// record success
				record, err := imdb.GetNoUploadRecord(req.EventInfo.RoomId, constant.RoomIDTypeInteger, req.CallbackTs)
				if err != nil {
					log.NewError("callback", utils.GetSelfFuncName(), "not find the avtive record ", err.Error())
					c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "not found the record"})
					return
				}
				if len(req.EventInfo.Payload.FileList) > 0 {
					record.RecordURL = req.EventInfo.Payload.FileList[0]
				}
				record.RecordStatus = constant.CommunicationRecordStatusFinished
				// update record
				err = imdb.UpdateCommunication(&record)
				if err != nil {
					log.NewError("callback", utils.GetSelfFuncName(), "save url to db error", err.Error())
					c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "save url to db error"})
					return
				}
			}
		}
	}
	log.NewError("callback", utils.GetSelfFuncName(), "finish callback")
	c.JSON(http.StatusOK, nil)
	return

}

func getUserStatus(userID, groupID string) (video, audio int8, err error) {
	user, err := imdb.GetUserByUserID(userID)
	if err != nil {
		return 2, 2, err
	}
	audio = user.AudioStatus
	video = user.VideoStatus
	if groupID != "" {
		groupMember, err := imdb.GetGroupMemberByUserIDGroupID(groupID, userID)
		if err != nil {
			return 2, 2, err
		}
		// if user audio is allowed, but in this group isn't allowed
		if audio == 1 && groupMember.AudioStatus == 2 {
			audio = groupMember.AudioStatus
		}
		if video == 1 && groupMember.VideoStatus == 2 {
			video = groupMember.VideoStatus
		}
	}

	return video, audio, nil
}

func checkCanCommunication(userID, groupID string, chatType int8) error {
	video, audio, err := getUserStatus(userID, groupID)
	if err != nil {
		log.NewError("", utils.GetSelfFuncName(), "video banned", err.Error())
		return errors.New("video banned, your account isn't exist")
	}
	if chatType == 1 && video == 2 {
		log.NewError("", utils.GetSelfFuncName(), "video banned")
		return errors.New("your video calling was banned")
	} else if audio == 2 {
		log.NewError("", utils.GetSelfFuncName(), "video banned")
		return errors.New("your audio calling was banned")
	}

	// if you have been in a calling
	status, err := db.DB.GetUserCommunicationStatus(userID)
	if err == nil && status == 1 {
		log.NewError("", utils.GetSelfFuncName(), "GetUserCommunicationStatus")
		return errors.New("you have another device is in a communication")
	}

	// if the user's region is banned.

	//activeRecord, err := imdb.GetActiveCommunicationByCallID(req.CallID)
	//if err != nil {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetActiveCommunicationByCallID failed ", err.Error())
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
	//	return
	//}
	//if activeRecord.CallID == "" {
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "No this communication")
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "No this communication"})
	//	return
	//}

	return nil
}
