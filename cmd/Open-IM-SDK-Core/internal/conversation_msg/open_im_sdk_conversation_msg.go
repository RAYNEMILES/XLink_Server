package conversation_msg

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/common"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	db2 "Open_IM/cmd/Open-IM-SDK-Core/pkg/db"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/db/model_struct"
	api "Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	"Open_IM/pkg/common/log"
	utils2 "Open_IM/pkg/utils"
	"net/url"
	"strconv"

	"Open_IM/cmd/Open-IM-SDK-Core/pkg/sdk_params_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/cmd/Open-IM-SDK-Core/sdk_struct"
	sdk "Open_IM/pkg/proto/sdk_ws"
	"encoding/json"
	"errors"
	"image"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/copier"
	imgtype "github.com/shamsher31/goimgtype"
)

func (c *Conversation) GetAllConversationList(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "GetAllConversationList args: ")
		result := c.getAllConversationList(callback, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "GetAllConversationList-2 callback: ", utils.StructToJsonStringDefault(result))
	}()
}

func (c *Conversation) GetAllConversationListByType(callback open_im_sdk_callback.Base, conversationType []int32, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "GetAllConversationList args: ")
		result := c.getAllConversationListByType(callback, conversationType, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "GetAllConversationList-2 callback: ", utils.StructToJsonStringDefault(result))
	}()
}

func (c *Conversation) GetConversationListSplit(callback open_im_sdk_callback.Base, offset, count int, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "GetConversationListSplit args: ", offset, count)
		result := c.getConversationListSplit(callback, offset, count, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "GetConversationListSplit callback: ", utils.StructToJsonStringDefault(result))
	}()
}

func (c *Conversation) SetConversationRecvMessageOpt(callback open_im_sdk_callback.Base, conversationIDList string, opt int, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "SetConversationRecvMessageOpt args: ", conversationIDList, opt)
		var unmarshalParams sdk_params_callback.SetConversationRecvMessageOptParams
		common.JsonUnmarshalCallback(conversationIDList, &unmarshalParams, callback, operationID)
		c.setConversationRecvMessageOpt(callback, unmarshalParams, opt, operationID)
		callback.OnSuccess(sdk_params_callback.SetConversationRecvMessageOptCallback)
		log.NewInfo(operationID, "SetConversationRecvMessageOpt callback: ", sdk_params_callback.SetConversationRecvMessageOptCallback)
	}()
}
func (c *Conversation) SetGlobalRecvMessageOpt(callback open_im_sdk_callback.Base, opt int, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "SetGlobalRecvMessageOpt args: ", opt)
		c.setGlobalRecvMessageOpt(callback, int32(opt), operationID)
		callback.OnSuccess(sdk_params_callback.SetGlobalRecvMessageOptCallback)
		log.NewInfo(operationID, "SetGlobalRecvMessageOpt callback: ", sdk_params_callback.SetGlobalRecvMessageOptCallback)
	}()
}

// deprecated
func (c *Conversation) GetConversationRecvMessageOpt(callback open_im_sdk_callback.Base, conversationIDList, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "GetConversationRecvMessageOpt args: ", conversationIDList)
		var unmarshalParams sdk_params_callback.GetConversationRecvMessageOptParams
		common.JsonUnmarshalCallback(conversationIDList, &unmarshalParams, callback, operationID)
		result := c.getConversationRecvMessageOpt(callback, unmarshalParams, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "GetConversationRecvMessageOpt callback: ", utils.StructToJsonStringDefault(result))
	}()
}

func (c *Conversation) GetOneConversation(callback open_im_sdk_callback.Base, sessionType int32, sourceID, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "GetOneConversation args: ", sessionType, sourceID)
		result := c.getOneConversation(callback, sourceID, sessionType, operationID)
		callback.OnSuccess(utils.StructToJsonString(result))
		log.NewInfo(operationID, "GetOneConversation callback: ", utils.StructToJsonString(result))
	}()
}

func (c *Conversation) GetMultipleConversation(callback open_im_sdk_callback.Base, conversationIDList string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "GetMultipleConversation args: ", conversationIDList)
		var unmarshalParams sdk_params_callback.GetMultipleConversationParams
		common.JsonUnmarshalCallback(conversationIDList, &unmarshalParams, callback, operationID)
		result := c.getMultipleConversation(callback, unmarshalParams, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "GetMultipleConversation callback: ", utils.StructToJsonStringDefault(result))
	}()
}
func (c *Conversation) DeleteConversation(callback open_im_sdk_callback.Base, conversationID string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "DeleteConversation args: ", conversationID)
		c.deleteConversation(callback, conversationID, operationID)
		callback.OnSuccess(sdk_params_callback.DeleteConversationCallback)
		log.NewInfo(operationID, "DeleteConversation callback: ", sdk_params_callback.DeleteConversationCallback)
	}()
}
func (c *Conversation) DeleteAllConversationFromLocal(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "DeleteAllConversationFromLocal args: ")
		err := c.db.ResetAllConversation()
		common.CheckDBErrCallback(callback, err, operationID)
		callback.OnSuccess(sdk_params_callback.DeleteAllConversationFromLocalCallback)
		log.NewInfo(operationID, "DeleteConversation callback: ", sdk_params_callback.DeleteAllConversationFromLocalCallback)
	}()
}
func (c *Conversation) SetConversationDraft(callback open_im_sdk_callback.Base, conversationID, draftText string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "SetConversationDraft args: ", conversationID)
		c.setConversationDraft(callback, conversationID, draftText, operationID)
		callback.OnSuccess(sdk_params_callback.SetConversationDraftCallback)
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
		log.NewInfo(operationID, "SetConversationDraft callback: ", sdk_params_callback.SetConversationDraftCallback)
	}()
}
func (c *Conversation) ResetConversationGroupAtType(callback open_im_sdk_callback.Base, conversationID, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "ResetConversationGroupAtType args: ", conversationID)
		c.setOneConversationGroupAtType(callback, conversationID, operationID)
		callback.OnSuccess(sdk_params_callback.ResetConversationGroupAtTypeCallback)
		log.NewInfo(operationID, "ResetConversationGroupAtType callback: ", sdk_params_callback.ResetConversationGroupAtTypeCallback)
	}()
}
func (c *Conversation) PinConversation(callback open_im_sdk_callback.Base, conversationID string, isPinned bool, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "PinConversation args: ", conversationID, isPinned)
		c.pinConversation(callback, conversationID, isPinned, operationID)
		callback.OnSuccess(sdk_params_callback.PinConversationDraftCallback)
		log.NewInfo(operationID, "PinConversation callback: ", sdk_params_callback.PinConversationDraftCallback)
	}()
}

func (c *Conversation) SetOneConversationPrivateChat(callback open_im_sdk_callback.Base, conversationID string, isPrivate bool, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", conversationID, isPrivate)
		c.setOneConversationPrivateChat(callback, conversationID, isPrivate, operationID)
		callback.OnSuccess(sdk_params_callback.SetConversationMessageOptCallback)
		log.NewInfo(operationID, utils.GetSelfFuncName(), "callback: ", sdk_params_callback.SetConversationMessageOptCallback)
	}()
}

func (c *Conversation) SetOneConversationRecvMessageOpt(callback open_im_sdk_callback.Base, conversationID string, opt int, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, utils.GetSelfFuncName(), "args: ", conversationID, opt)
		c.setOneConversationRecvMessageOpt(callback, conversationID, opt, operationID)
		callback.OnSuccess(sdk_params_callback.SetConversationMessageOptCallback)
		log.NewInfo(operationID, utils.GetSelfFuncName(), "callback: ", sdk_params_callback.SetConversationMessageOptCallback)
	}()
}

func (c *Conversation) GetTotalUnreadMsgCount(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "GetTotalUnreadMsgCount args: ")
		count, err := c.db.GetTotalUnreadMsgCount()
		common.CheckDBErrCallback(callback, err, operationID)
		callback.OnSuccess(utils.Int32ToString(count))
		log.NewInfo(operationID, "GetTotalUnreadMsgCount callback: ", utils.Int32ToString(count))
	}()
}

func (c *Conversation) SetConversationListener(listener open_im_sdk_callback.OnConversationListener) {
	if c.ConversationListener != nil {
		log.Error("internal", "just only set on listener")
		return
	}
	c.ConversationListener = listener
}

func (c *Conversation) GetConversationsByUserID(callback open_im_sdk_callback.Base, operationID string, UserID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, utils.GetSelfFuncName())
		conversations, err := c.db.GetAllConversationList()
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
		}
		var conversationIDs []string
		for _, conversation := range conversations {
			conversationIDs = append(conversationIDs, conversation.ConversationID)
		}
	}()
}

//

////
////func (c *Conversation) ForceSyncMsg() bool {
////	if c.syncSeq2Msg() == nil {
////		return true
////	} else {
////		return false
////	}
////}
////
////func (c *Conversation) ForceSyncJoinedGroup() {
////	u.syncJoinedGroupInfo()
////}
////
////func (c *Conversation) ForceSyncJoinedGroupMember() {
////
////	u.syncJoinedGroupMember()
////}
////
////func (c *Conversation) ForceSyncGroupRequest() {
////	u.syncGroupRequest()
////}
////
////func (c *Conversation) ForceSyncSelfGroupRequest() {
////	u.syncSelfGroupRequest()
////}
//

func (c *Conversation) CreateEmojiMessage(text, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Emoji, operationID)
	s.Content = text
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateTextMessage(text, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Text, operationID)
	s.Content = text
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateAdvancedTextMessage(text, messageEntityList, operationID string) string {
	var messageEntitys []*sdk_struct.MessageEntity
	s := sdk_struct.MsgStruct{}
	err := json.Unmarshal([]byte(messageEntityList), &messageEntitys)
	if err != nil {
		log.Error("internal", "messages unmarshal err", err.Error())
		return ""
	}
	c.initBasicInfo(&s, constant.UserMsgType, constant.AdvancedText, operationID)
	s.MessageEntityElem.Text = text
	s.MessageEntityElem.MessageEntityList = messageEntitys
	s.Content = utils.StructToJsonString(s.MessageEntityElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateTextAtMessage(text, atUserList, atUsersInfo, message, operationID string) string {
	var usersInfo []*sdk_struct.AtInfo
	var userIDList []string
	_ = json.Unmarshal([]byte(atUsersInfo), &usersInfo)
	_ = json.Unmarshal([]byte(atUserList), &userIDList)
	s, qs := sdk_struct.MsgStruct{}, sdk_struct.MsgStruct{}
	_ = json.Unmarshal([]byte(message), &qs)
	c.initBasicInfo(&s, constant.UserMsgType, constant.AtText, operationID)
	//Avoid nested references
	if qs.ContentType == constant.Quote {
		qs.Content = qs.QuoteElem.Text
		qs.ContentType = constant.Text
	}
	s.AtElem.Text = text
	s.AtElem.AtUserList = userIDList
	s.AtElem.AtUsersInfo = usersInfo
	s.AtElem.QuoteMessage = &qs
	if message == "" {
		s.AtElem.QuoteMessage = nil
	}
	s.Content = utils.StructToJsonString(s.AtElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateLocationMessage(description string, longitude, latitude float64, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Location, operationID)
	s.LocationElem.Description = description
	s.LocationElem.Longitude = longitude
	s.LocationElem.Latitude = latitude
	s.Content = utils.StructToJsonString(s.LocationElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateCustomMessage(data, extension string, description, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Custom, operationID)
	s.CustomElem.Data = data
	s.CustomElem.Extension = extension
	s.CustomElem.Description = description
	s.Content = utils.StructToJsonString(s.CustomElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateQuoteMessage(text string, message, operationID string) string {
	s, qs := sdk_struct.MsgStruct{}, sdk_struct.MsgStruct{}
	_ = json.Unmarshal([]byte(message), &qs)
	c.initBasicInfo(&s, constant.UserMsgType, constant.Quote, operationID)
	//Avoid nested references
	if qs.ContentType == constant.Quote {
		qs.Content = qs.QuoteElem.Text
		qs.ContentType = constant.Text
	}
	s.QuoteElem.Text = text
	s.QuoteElem.QuoteMessage = &qs
	s.Content = utils.StructToJsonString(s.QuoteElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateAdvancedQuoteMessage(text string, message, messageEntityList, operationID string) string {
	var messageEntitys []*sdk_struct.MessageEntity
	s, qs := sdk_struct.MsgStruct{}, sdk_struct.MsgStruct{}
	_ = json.Unmarshal([]byte(message), &qs)
	_ = json.Unmarshal([]byte(messageEntityList), &messageEntitys)
	c.initBasicInfo(&s, constant.UserMsgType, constant.Quote, operationID)
	//Avoid nested references
	if qs.ContentType == constant.Quote {
		qs.Content = qs.QuoteElem.Text
		qs.ContentType = constant.Text
	}
	s.QuoteElem.Text = text
	s.QuoteElem.MessageEntityList = messageEntitys
	s.QuoteElem.QuoteMessage = &qs
	s.Content = utils.StructToJsonString(s.QuoteElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateCardMessage(cardInfo, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Card, operationID)
	s.Content = cardInfo
	return utils.StructToJsonString(s)

}
func (c *Conversation) CreateVideoMessageFromFullPath(videoFullPath string, videoType string, duration int64, snapshotFullPath, operationID string) string {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		dstFile := utils.FileTmpPath(videoFullPath, c.DataDir) //a->b
		s, err := utils.CopyFile(videoFullPath, dstFile)
		if err != nil {
			log.Error("internal", "open file failed: ", err, videoFullPath)
		}
		log.Info("internal", "videoFullPath dstFile", videoFullPath, dstFile, s)
		dstFile = utils.FileTmpPath(snapshotFullPath, c.DataDir) //a->b
		s, err = utils.CopyFile(snapshotFullPath, dstFile)
		if err != nil {
			log.Error("internal", "open file failed: ", err, snapshotFullPath)
		}
		log.Info("internal", "snapshotFullPath dstFile", snapshotFullPath, dstFile, s)
		wg.Done()
	}()

	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Video, operationID)
	s.VideoElem.VideoPath = videoFullPath
	s.VideoElem.VideoType = videoType
	s.VideoElem.Duration = duration
	if snapshotFullPath == "" {
		s.VideoElem.SnapshotPath = ""
	} else {
		s.VideoElem.SnapshotPath = snapshotFullPath
	}
	fi, err := os.Stat(s.VideoElem.VideoPath)
	if err != nil {
		log.Error("internal", "get file Attributes error", err.Error())
		return ""
	}
	s.VideoElem.VideoSize = fi.Size()
	if snapshotFullPath != "" {
		imageInfo, err := getImageInfo(s.VideoElem.SnapshotPath)
		if err != nil {
			log.Error("internal", "get Image Attributes error", err.Error())
			return ""
		}
		s.VideoElem.SnapshotHeight = imageInfo.Height
		s.VideoElem.SnapshotWidth = imageInfo.Width
		s.VideoElem.SnapshotSize = imageInfo.Size
	}
	wg.Wait()
	s.Content = utils.StructToJsonString(s.VideoElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateFileMessageFromFullPath(fileFullPath string, fileName, operationID string) string {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		dstFile := utils.FileTmpPath(fileFullPath, c.DataDir)
		_, err := utils.CopyFile(fileFullPath, dstFile)
		log.Info(operationID, "copy file, ", fileFullPath, dstFile)
		if err != nil {
			log.Error("internal", "open file failed: ", err.Error(), fileFullPath)
		}
		wg.Done()
	}()
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.File, operationID)
	s.FileElem.FilePath = fileFullPath
	fi, err := os.Stat(fileFullPath)
	if err != nil {
		log.Error("internal", "get file Attributes error", err.Error())
		return ""
	}
	s.FileElem.FileSize = fi.Size()
	s.FileElem.FileName = fileName
	s.Content = utils.StructToJsonString(s.FileElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateImageMessageFromFullPath(imageFullPath, operationID string) string {
	var wg sync.WaitGroup
	wg.Add(1)
	imagePathEncode, err := url.Parse(imageFullPath)
	if err == nil {
		imageFullPath = imagePathEncode.String()
	}
	go func() {
		dstFile := utils.FileTmpPath(imageFullPath, c.DataDir) //a->b
		_, err := utils.CopyFile(imageFullPath, dstFile)
		log.Info("internal", "copy file, ", imageFullPath, dstFile)
		if err != nil {
			log.Error("internal", "open file failed: ", err, imageFullPath)
		}
		wg.Done()
	}()

	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Picture, operationID)
	s.PictureElem.SourcePath = imageFullPath
	log.Info("internal", "ImageMessage  path:", s.PictureElem.SourcePath)
	imageInfo, err := getImageInfo(s.PictureElem.SourcePath)
	if err != nil {
		log.Error("internal", "getImageInfo err:", err.Error())
		return ""
	}
	s.PictureElem.SourcePicture.Width = imageInfo.Width
	s.PictureElem.SourcePicture.Height = imageInfo.Height
	s.PictureElem.SourcePicture.Type = imageInfo.Type
	s.PictureElem.SourcePicture.Size = imageInfo.Size
	wg.Wait()
	s.Content = utils.StructToJsonString(s.PictureElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateSoundMessageFromFullPath(soundPath string, duration int64, operationID string) string {
	var wg sync.WaitGroup
	wg.Add(1)
	soundPathEncode, err := url.Parse(soundPath)
	if err == nil {
		soundPath = soundPathEncode.String()
	}
	go func() {
		dstFile := utils.FileTmpPath(soundPath, c.DataDir) //a->b
		_, err := utils.CopyFile(soundPath, dstFile)
		log.Info("internal", "copy file, ", soundPath, dstFile)
		if err != nil {
			log.Error("internal", "open file failed: ", err, soundPath)
		}
		wg.Done()
	}()
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Voice, operationID)
	s.SoundElem.SoundPath = soundPath
	s.SoundElem.Duration = duration
	fi, err := os.Stat(s.SoundElem.SoundPath)
	if err != nil {
		log.Error("internal", "getSoundInfo err:", err.Error(), s.SoundElem.SoundPath)
		return ""
	}
	s.SoundElem.DataSize = fi.Size()
	wg.Wait()
	s.Content = utils.StructToJsonString(s.SoundElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateImageMessage(imagePath, operationID string) string {
	s := sdk_struct.MsgStruct{}
	imagePathEncode, err := url.Parse(imagePath)
	if err == nil {
		imagePath = imagePathEncode.String()
	}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Picture, operationID)
	s.PictureElem.SourcePath = c.DataDir + imagePath
	log.Debug("internal", "ImageMessage  path:", s.PictureElem.SourcePath)
	imageInfo, err := getImageInfo(s.PictureElem.SourcePath)
	if err != nil {
		log.Error("internal", "get imageInfo err", err.Error())
		return ""
	}
	s.PictureElem.SourcePicture.Width = imageInfo.Width
	s.PictureElem.SourcePicture.Height = imageInfo.Height
	s.PictureElem.SourcePicture.Type = imageInfo.Type
	s.PictureElem.SourcePicture.Size = imageInfo.Size

	s.Content = utils.StructToJsonString(s.PictureElem)
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateArticleMessage(Text, ArticleID, ArticleTitle, ArticleCoverPhoto, ArticleDetailsUrl, OfficialID, OfficialName, OfficialFaceUrl, operationID string) string {
	s := sdk_struct.MsgStruct{}

	c.initBasicInfo(&s, constant.UserMsgType, constant.Article, operationID)
	log.Debug("internal", "Article Message  path:", s.PictureElem.SourcePath)
	s.ArticleElem.Content = Text
	s.ArticleElem.ArticleID = ArticleID
	s.ArticleElem.ArticleTitle = ArticleTitle
	s.ArticleElem.ArticleDetailsUrl = ArticleDetailsUrl
	s.ArticleElem.ArticleCoverPhoto = ArticleCoverPhoto
	s.ArticleElem.OfficialID = OfficialID
	s.ArticleElem.OfficialName = OfficialName
	s.ArticleElem.OfficialFaceUrl = OfficialFaceUrl

	s.Content = utils.StructToJsonString(s.ArticleElem)
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateWoomMessage(fileId, coverUrl, mediaUrl, userId, nickname, faceUrl, operationID, desc string) string {
	s := sdk_struct.MsgStruct{}

	c.initBasicInfo(&s, constant.UserMsgType, constant.Woom, operationID)
	s.WoomElem.FileID = fileId
	s.WoomElem.CoverUrl = coverUrl
	s.WoomElem.MediaUrl = mediaUrl
	s.WoomElem.UserId = userId
	s.WoomElem.Nickname = nickname
	s.WoomElem.FaceUrl = faceUrl
	s.WoomElem.Desc = desc

	s.Content = utils.StructToJsonString(s.WoomElem)
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateImageMessageByURL(sourcePicture, bigPicture, snapshotPicture, operationID string) string {
	log.Info(operationID, "Sandman#8 SourcePicture", sourcePicture)
	s := sdk_struct.MsgStruct{}
	// var p sdk_struct.PictureBaseInfo
	var pSrc sdk_struct.PictureBaseInfo
	_ = json.Unmarshal([]byte(sourcePicture), &pSrc)
	imagePathEncode, err := url.Parse(pSrc.Url)
	if err == nil {
		pSrc.Url = imagePathEncode.String()
	}
	s.PictureElem.SourcePicture = pSrc

	var pBig sdk_struct.PictureBaseInfo
	_ = json.Unmarshal([]byte(bigPicture), &pBig)
	imagePathEncode, err = url.Parse(pBig.Url)
	if err == nil {
		pBig.Url = imagePathEncode.String()
		// newWidth := int32((float32(pSrc.Width) * 1.25))
		// resizeImageUrlGenrator(&pBig, pSrc.Width, pSrc.Height, newWidth)
	}
	s.PictureElem.BigPicture = pBig

	var pSanap sdk_struct.PictureBaseInfo
	_ = json.Unmarshal([]byte(snapshotPicture), &pSanap)
	imagePathEncode, err = url.Parse(pSanap.Url)
	if err == nil {
		pSanap.Url = imagePathEncode.String()
		// newWidth := int32(200)
		// resizeImageUrlGenrator(&pSanap, pSrc.Width, pSrc.Height, newWidth)
	}
	s.PictureElem.SnapshotPicture = pSanap
	c.initBasicInfo(&s, constant.UserMsgType, constant.Picture, operationID)
	s.Content = utils.StructToJsonString(s.PictureElem)
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateCallingMessage(MessageType, BanStatus int8, ErrCode int32, Duration int64, operationID string) string {
	log.Info(operationID, "Create calling message")
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Calling, operationID)
	log.Debug(operationID, "Set basic info finished, start to build calling message")

	s.CallingElem.MessageType = MessageType
	s.CallingElem.ErrCode = ErrCode
	s.CallingElem.Duration = Duration
	s.CallingElem.BanStatus = BanStatus

	s.Content = utils.StructToJsonString(s.CallingElem)

	return utils.StructToJsonString(s)
}

// func resizeImageUrlGenrator(p *sdk_struct.PictureBaseInfo, srcWidth, srcHeight, newWidth int32) {
// 	newHeight := int32((float32(srcHeight) / float32(srcWidth)) * float32(newWidth))
// 	p.Height = newHeight
// 	p.Width = newWidth
// 	imageProxyBaseUrl := config.Config.ImageProxy.ImgProxyAddress
// 	if imageProxyBaseUrl == "" {
// 		imageProxyBaseUrl = "https://img.bytechat-test.com/AfrOrF3gWeDA6VOlDG4TzxMv39O7MXnF4CXpKUwGqRM/resize:fill:"
// 	}
// 	imageProxyUrl := imageProxyBaseUrl + fmt.Sprint(newWidth) + ":" + fmt.Sprint(newHeight) + ":0/gravity:sm/plain/" + p.Url
// 	// imageProxyUrl := config.Config.ImageProxy.ImgProxyAddress + fmt.Sprint(newWidth) + ":" + fmt.Sprint(newHeight) + config.Config.ImageProxy.ImgGravityPrePos + p.Url
// 	p.Url = imageProxyUrl
// }

func msgStructToLocalChatLog(dst *model_struct.LocalChatLog, src *sdk_struct.MsgStruct) {
	copier.Copy(dst, src)
	if src.SessionType == constant.GroupChatType || src.SessionType == constant.SuperGroupChatType {
		dst.RecvID = src.GroupID
	}
}
func localChatLogToMsgStruct(dst *sdk_struct.NewMsgList, src []*model_struct.LocalChatLog) {
	copier.Copy(dst, &src)

}
func (c *Conversation) checkErrAndUpdateMessage(callback open_im_sdk_callback.SendMsgCallBack, errCode int32, err error, s *sdk_struct.MsgStruct, lc *model_struct.LocalConversation, operationID string) {
	if err != nil {
		if callback != nil {
			c.updateMsgStatusAndTriggerConversation(s.ClientMsgID, "", s.CreateTime, constant.MsgStatusSendFailed, s, lc, operationID)
			errInfo := "operationID[" + operationID + "], " + "info[" + err.Error() + "]" + s.ClientMsgID + " " + s.ServerMsgID
			log.NewError(operationID, "checkErr ", errInfo)
			callback.OnError(errCode, errInfo)
			runtime.Goexit()
		}
	}
}

func (c *Conversation) checkErrAndUpdateBroadcastMessage(callback open_im_sdk_callback.SendMsgCallBack, errCode int32, err error, s *sdk_struct.BroadcastMsgStruct, lc *model_struct.LocalConversation, operationID string) {
	if err != nil {
		if callback != nil {
			//TODO update the local db with Message status failed and update conversation
			c.updateBroadcastMsgStatusAndTriggerConversation(s.ClientMsgID, "", s.CreateTime, constant.MsgStatusSendFailed, s, lc, operationID)
			errInfo := "operationID[" + operationID + "], " + "info[" + err.Error() + "]" + s.ClientMsgID + " " + s.ServerMsgID
			log.NewError(operationID, "checkErr ", errInfo)
			callback.OnError(errCode, errInfo)
			runtime.Goexit()
		}
	}
}
func (c *Conversation) updateMsgStatusAndTriggerConversation(clientMsgID, serverMsgID string, sendTime int64, status int32, s *sdk_struct.MsgStruct, lc *model_struct.LocalConversation, operationID string) {
	//log.NewDebug(operationID, "this is test send message ", sendTime, status, clientMsgID, serverMsgID)
	s.SendTime = sendTime
	s.Status = status
	s.ServerMsgID = serverMsgID
	err := c.db.UpdateMessageTimeAndStatusController(s)
	if err != nil {
		log.Error(operationID, "send message update message status error", sendTime, status, clientMsgID, serverMsgID, err.Error())
	}
	lc.LatestMsg = utils.StructToJsonString(s)
	lc.LatestMsgSendTime = sendTime
	log.Info(operationID, "2 send message come here", *lc)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: lc.ConversationID, Action: constant.AddConOrUpLatMsg, Args: *lc}, c.GetCh())

}

func (c *Conversation) updateBroadcastMsgStatusAndTriggerConversation(clientMsgID, serverMsgID string, sendTime int64, status int32, s *sdk_struct.BroadcastMsgStruct, lc *model_struct.LocalConversation, operationID string) {
	//TODO UPDATE IN BROADCAST LOCAL TABLE FOR MESSAGE STATUS

	log.NewDebug(operationID, "this is test send message ", sendTime, status, clientMsgID, serverMsgID)
	s.SendTime = sendTime
	s.Status = status
	s.ServerMsgID = serverMsgID
	err := c.db.UpdateBroadcastMessageTimeAndStatusController(s)
	if err != nil {
		log.Error(operationID, "send message update message status error", sendTime, status, clientMsgID, serverMsgID, err.Error())
	}

	lc.LatestMsg = utils.StructToJsonString(s)
	lc.LatestMsgSendTime = sendTime
	log.Info(operationID, "2 send message come here", *lc)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: lc.ConversationID, Action: constant.AddConOrUpLatMsg, Args: *lc}, c.GetCh())

}

func (c *Conversation) SendMessage(callback open_im_sdk_callback.SendMsgCallBack, message, recvID, groupID string, offlinePushInfo string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		s := sdk_struct.MsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)

		//log.NewError(operationID, utils.GetSelfFuncName(), "字符长度", len(s.Content), s.Content)
		if len(s.Content) > 10000 {
			common.CheckAnyErrCallback(callback, 201, errors.New("message too long:"+strconv.Itoa(len(s.Content))), operationID)
		}

		s.SendID = c.loginUserID
		s.SenderPlatformID = c.platformID
		p := &sdk.OfflinePushInfo{}
		if offlinePushInfo == "" {
			p = nil
		} else {
			common.JsonUnmarshalAndArgsValidate(offlinePushInfo, &p, callback, operationID)
		}
		if recvID == "" && groupID == "" {
			common.CheckAnyErrCallback(callback, 201, errors.New("recvID && groupID not both null"), operationID)
		}
		var localMessage model_struct.LocalChatLog
		var conversationID string
		options := make(map[string]bool, 2)
		lc := &model_struct.LocalConversation{LatestMsgSendTime: s.CreateTime}

		// Update counter for forward message counter, Will fetch list of recent forward conv
		//For Article Share checking the article type because Article share also sent as message & this whole feature build for article first place
		if (recvID != "" && s.RecvID != "" && recvID != s.RecvID) || (groupID != "" && s.GroupID != "" && groupID != s.GroupID) || s.ContentType == constant.Article {
			// it's a forward message
			go func() {
				receiverID := recvID
				receiverType := 1
				if groupID != "" {
					receiverID = groupID
					receiverType = 2
				}
				err := c.db.SetForwardMessagesToConversation(receiverID, int64(receiverType))
				if err != nil {
					log.Error("internal", "Forward counter increment failed", err.Error())
				}
			}()
		}
		//根据单聊群聊类型组装消息和会话
		if recvID == "" {
			s.GroupID = groupID
			lc.GroupID = groupID

			g, err := c.full.GetGroupInfoByGroupID(groupID)
			common.CheckAnyErrCallback(callback, 202, err, operationID)

			//check if the user is owner or admin
			memberInfo, err := c.group.GetOneGroupMemberInfo(groupID, c.loginUserID)
			if err != nil {
				//sync again myself
				c.group.SyncMySelfInTheGroup(groupID, operationID)
				//get group membe again
				memberInfo, err = c.group.GetOneGroupMemberInfo(groupID, c.loginUserID)
				if err != nil {
					common.CheckAnyErrCallback(callback, 301, err, operationID)
				}
			}

			if memberInfo.Nickname != "" {
				s.SenderNickname = memberInfo.Nickname
			}
			if memberInfo.MuteEndTime != 0 && int64(memberInfo.MuteEndTime) > time.Now().Unix() {
				common.CheckAnyErrCallback(callback, 301, errors.New("you are muted!"), operationID)
			}

			//check if the group was muted
			//ownerID, adminIDList, err := c.group.GetGroupOwnerIDAndAdminIDList(groupID, operationID)
			if g.Status != 0 {
				if memberInfo.RoleLevel == constant.GroupOrdinaryUsers {
					common.CheckAnyErrCallback(callback, 302, errors.New("group was invalid!"), operationID)
				}
			}

			//check if the group was dissolved
			if g.Status == 2 {
				common.CheckAnyErrCallback(callback, 302, errors.New("group was dissolved!"), operationID)
			}

			lc.ShowName = g.GroupName
			lc.FaceURL = g.FaceURL
			switch g.GroupType {
			case constant.NormalGroup:
				s.SessionType = constant.GroupChatType
				lc.ConversationType = constant.GroupChatType
				conversationID = utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
			case constant.SuperGroup, constant.WorkingGroup:
				s.SessionType = constant.SuperGroupChatType
				conversationID = utils.GetConversationIDBySessionType(groupID, constant.SuperGroupChatType)
				lc.ConversationType = constant.SuperGroupChatType
			}
			//c.SyncOneConversation(conversationID, operationID)
			//c.group.SyncGroupMemberByGroupID(groupID, operationID, false)

			//todo need to fix if conversation deleted but you are part of that conv or group
			//todo 110423
			oldConversation, err := c.db.GetConversation(conversationID)
			if err == nil {
				if oldConversation.IsNotInGroup {
					common.CheckAnyErrCallback(callback, 301, errors.New("you are not in the group by checking oldConversation.IsNotInGroup!"), operationID)
				}
				//common.CheckAnyErrCallback(callback, 301, err, operationID)
			} else {
				_, err := c.db.GetGroupInfoByGroupID(groupID)
				if err != nil {
					common.CheckAnyErrCallback(callback, 301, err, operationID)
				}
			}

			//groupMemberUidList, err := c.db.GetGroupMemberUIDListByGroupID(groupID)
			//common.CheckAnyErrCallback(callback, 202, err, operationID)
			//if !utils.IsContain(s.SendID, groupMemberUidList) {
			//	common.CheckAnyErrCallback(callback, 208, errors.New("you not exist in this group"), operationID)
			//}
			s.AttachedInfoElem.GroupHasReadInfo.GroupMemberCount = g.MemberCount
			s.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
		} else {
			s.SessionType = constant.SingleChatType
			s.RecvID = recvID
			conversationID = utils.GetConversationIDBySessionType(recvID, constant.SingleChatType)
			lc.UserID = recvID
			lc.ConversationType = constant.SingleChatType
			//check if the friend of sender
			_, err := c.db.GetFriendInfoByFriendUserID(s.RecvID)
			if err != nil {
				//todo check on server for receiver black list, rather then uploading file and check that on server.
				msgStructToLocalChatLog(&localMessage, &s)
				localMessage.ServerMsgID = localMessage.ClientMsgID
				localMessage.Seq = uint32(c.db.GetLocalMaxSeq())
				localMessage.Status = constant.MsgStatusSendFailed
				err := c.db.InsertMessage(&localMessage)

				if err != nil {
					log.Warn(operationID, "Local Message insert Fail", err.Error(), localMessage)
				}

				oldLc, err := c.db.GetConversation(conversationID)
				if err == nil && oldLc != nil {
					oldLc.LatestMsg = utils.StructToJsonString(s)
					oldLc.LatestMsgSendTime = s.SendTime
					oldLc.ConversationID = conversationID
					err := c.db.UpdateConversationForSync(oldLc)
					if err != nil {
						log.Warn(operationID, "Local Conversation insert Fail", err.Error(), oldLc)
					}
				}
				_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.AddConOrUpLatMsg, Args: *oldLc}, c.GetCh())
				log.Warn(operationID, "Local Message insert Success", localMessage)
				common.CheckAnyErrCallback(callback, 302, errors.New("you are not friends"), operationID)
			}
			//faceUrl, name, err := c.friend.GetUserNameAndFaceUrlByUid(recvID, operationID)
			oldLc, err := c.db.GetConversation(conversationID)
			if err == nil && oldLc.IsPrivateChat {
				options[constant.IsNotPrivate] = false
				s.AttachedInfoElem.IsPrivateChat = true
				s.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
			}
			if err != nil {
				faceUrl, name, err := c.cache.GetUserNameAndFaceURL(recvID, operationID)
				common.CheckAnyErrCallback(callback, 301, err, operationID)
				lc.FaceURL = faceUrl
				lc.ShowName = name
			}

		}
		log.Warn(operationID, "before insert  message is ", s)
		oldMessage, err := c.db.GetMessageController(&s)
		if err != nil {
			msgStructToLocalChatLog(&localMessage, &s)
			err := c.db.InsertMessageController(&localMessage)
			common.CheckAnyErrCallback(callback, 201, err, operationID)
		} else {
			if oldMessage.Status != constant.MsgStatusSendFailed {
				common.CheckAnyErrCallback(callback, 202, errors.New("only failed message can be repeatedly send"), operationID)
			} else {
				s.Status = constant.MsgStatusSending
			}
		}
		lc.ConversationID = conversationID
		lc.LatestMsg = utils.StructToJsonString(s)
		log.Info(operationID, "send message come here", *lc)
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.AddConOrUpLatMsg, Args: *lc}, c.GetCh())
		var delFile []string
		//media file handle
		if s.Status != constant.MsgStatusSendSuccess { //filter forward message
			switch s.ContentType {
			case constant.Picture:
				var sourcePath string
				if utils.FileExist(s.PictureElem.SourcePath) {
					sourcePath = s.PictureElem.SourcePath
					delFile = append(delFile, utils.FileTmpPath(s.PictureElem.SourcePath, c.DataDir))
				} else {
					sourcePath = utils.FileTmpPath(s.PictureElem.SourcePath, c.DataDir)
					delFile = append(delFile, sourcePath)
				}
				log.Info(operationID, "file", sourcePath, delFile)
				sourceUrl, uuid, err := c.UploadImage(sourcePath, s.ClientMsgID, callback.OnProgress)
				c.checkErrAndUpdateMessage(callback, 301, err, &s, lc, operationID)
				s.PictureElem.SourcePicture.Url = sourceUrl
				s.PictureElem.SourcePicture.UUID = uuid

				s.PictureElem.SnapshotPicture.Url = sourceUrl
				// s.PictureElem.SnapshotPicture.UUID = uuid
				// newWidth := int32(200)
				// resizeImageUrlGenrator(&s.PictureElem.SnapshotPicture, s.PictureElem.SourcePicture.Width, s.PictureElem.SourcePicture.Height, newWidth)
				s.PictureElem.SnapshotPicture.Url = sourceUrl + "?imageView2/2/w/" + constant.ZoomScale + "/h/" + constant.ZoomScale
				s.PictureElem.SnapshotPicture.Width = int32(utils.StringToInt(constant.ZoomScale))
				s.PictureElem.SnapshotPicture.Height = int32(utils.StringToInt(constant.ZoomScale))
				s.Content = utils.StructToJsonString(s.PictureElem)

			case constant.Voice:
				var sourcePath string
				if utils.FileExist(s.SoundElem.SoundPath) {
					sourcePath = s.SoundElem.SoundPath
					delFile = append(delFile, utils.FileTmpPath(s.SoundElem.SoundPath, c.DataDir))
				} else {
					sourcePath = utils.FileTmpPath(s.SoundElem.SoundPath, c.DataDir)
					delFile = append(delFile, sourcePath)
				}
				log.Info(operationID, "file", sourcePath, delFile)
				soundURL, uuid, err := c.UploadSound(sourcePath, "", callback.OnProgress)
				c.checkErrAndUpdateMessage(callback, 301, err, &s, lc, operationID)
				s.SoundElem.SourceURL = soundURL
				s.SoundElem.UUID = uuid
				s.Content = utils.StructToJsonString(s.SoundElem)

			case constant.Video:
				var videoPath string
				var snapPath string
				if utils.FileExist(s.VideoElem.VideoPath) {
					videoPath = s.VideoElem.VideoPath
					snapPath = s.VideoElem.SnapshotPath
					delFile = append(delFile, utils.FileTmpPath(s.VideoElem.VideoPath, c.DataDir))
					delFile = append(delFile, utils.FileTmpPath(s.VideoElem.SnapshotPath, c.DataDir))
				} else {
					videoPath = utils.FileTmpPath(s.VideoElem.VideoPath, c.DataDir)
					snapPath = utils.FileTmpPath(s.VideoElem.SnapshotPath, c.DataDir)
					delFile = append(delFile, videoPath)
					delFile = append(delFile, snapPath)
				}
				log.Info(operationID, "file: ", videoPath, snapPath, delFile)
				snapshotURL, snapshotUUID, videoURL, videoUUID, err := c.UploadVideo(videoPath, snapPath, s.ClientMsgID, callback.OnProgress)
				c.checkErrAndUpdateMessage(callback, 301, err, &s, lc, operationID)
				s.VideoElem.VideoURL = videoURL
				s.VideoElem.SnapshotUUID = snapshotUUID
				s.VideoElem.SnapshotURL = snapshotURL
				s.VideoElem.VideoUUID = videoUUID
				s.Content = utils.StructToJsonString(s.VideoElem)
			case constant.File:
				fileURL, fileUUID, err := c.UploadFile(s.FileElem.FilePath, s.ClientMsgID, callback.OnProgress)
				c.checkErrAndUpdateMessage(callback, 301, err, &s, lc, operationID)
				s.FileElem.SourceURL = fileURL
				s.FileElem.UUID = fileUUID
				s.Content = utils.StructToJsonString(s.FileElem)
			case constant.Text:
			case constant.AtText:
			case constant.Location:
			case constant.Custom:
			case constant.Merger:
			case constant.Quote:
			case constant.Card:
			case constant.Face:
			case constant.AdvancedText:
			case constant.Article:
			case constant.WalletTransfer:
			case constant.OfficialAccount:
			case constant.Woom:
			case constant.Calling:
			case constant.Emoji:

			default:
				common.CheckAnyErrCallback(callback, 202, errors.New("contentType not currently supported"+utils.Int32ToString(s.ContentType)), operationID)
			}
			oldMessage, err := c.db.GetMessageController(&s)
			if err != nil {
				log.Warn(operationID, "get message err")
			}
			log.Warn(operationID, "before update database message is ", *oldMessage)
			if utils.IsContainInt(int(s.ContentType), []int{constant.Picture, constant.Voice, constant.Video, constant.File}) {
				msgStructToLocalChatLog(&localMessage, &s)
				log.Warn(operationID, "update message is ", s, localMessage)
				err = c.db.UpdateMessageController(&localMessage)
				common.CheckAnyErrCallback(callback, 201, err, operationID)
			}
		}
		//update conversation unread count in local database
		//err = updateConversationUnreadCountInLocal(operationID, s.SendID, recvID, groupID)
		//if err != nil {
		//	log.NewError(operationID, utils.GetSelfFuncName(), "updateConversationUnreadCountInLocal failed:", err.Error())
		//}
		c.sendMessageToServer(&s, lc, callback, delFile, p, options, operationID)
	}()

}

func (c *Conversation) SendCallingMessage(callback open_im_sdk_callback.SendMsgCallBack, message string, communicationID int64, offlinePushInfo, operationID string) {
	if callback == nil {
		return
	}

	// get member and group id from api
	getMemberReq := api.GetMembersByRoomIDReq{}
	getMemberResp := api.GetMembersByRoomIDResp{}
	getMemberReq.OperationID = operationID
	getMemberReq.CommunicationID = communicationID

	c.p.PostFatalCallback(callback, constant.GetMembersByRoomIDReq, getMemberReq, &getMemberResp, operationID)
	if getMemberResp.CommResp.ErrCode != constant.OK.ErrCode || getMemberResp.MembersID == nil {
		callback.OnError(getMemberResp.CommResp.ErrCode, getMemberResp.CommResp.ErrMsg)
		return
	}

	// change status success, then send calling message
	if getMemberResp.GroupID == "" {
		// personal calling, current we just have one member,
		for _, member := range getMemberResp.MembersID {
			c.SendMessage(callback, message, member, getMemberResp.GroupID, offlinePushInfo, operationID)
		}
	} else {
		// group calling, we don't have this currently.
	}

}

func (c *Conversation) GetCallingMessages(callback open_im_sdk_callback.SendMsgCallBack, operationID, userID, searchKey string, errCode, startTime, endTime int64) {
	if callback == nil {
		return
	}

	go func() {
		messages, err := c.db.GetCallingMessages(searchKey, errCode, startTime, endTime)
		if err != nil {
			errMsg := "get calling message list failed " + err.Error()
			callback.OnError(constant.ErrArgs.ErrCode, errMsg)
			return
		}

		type CallingMessage struct {
			ProfilePhoto string `json:"profile_photo"`
			Username     string `json:"username"`
			UserID       string `json:"user_id"`
			CallingType  int8   `json:"calling_type"`
			IsComing     bool   `json:"is_coming"`
			CallTime     int64  `json:"call_time"`
			ErrCode      int32  `json:"error_code"`
			CallingCount int32  `json:"calling_count"`
		}

		friendMap := map[string]*model_struct.LocalFriend{}

		friendIdList := make([]string, len(messages))
		for index, message := range messages {
			if userID == message.SendID {
				friendIdList[index] = message.RecvID
			} else {
				friendIdList[index] = message.SendID
			}
		}
		friendList, err := c.db.GetFriendInfoList(friendIdList)
		if err != nil {
			errMsg := "get friends info error " + err.Error()
			callback.OnError(constant.ErrArgs.ErrCode, errMsg)
			return
		}
		for index, friend := range friendList {
			friendMap[friend.FriendUserID] = friendList[index]
		}

		var result = make([]CallingMessage, 0)
		var multiCount = make([]int, 0)
		for _, message := range messages {
			callingElem := struct {
				MessageType int8  `json:"message_type"`
				BanStatus   int8  `json:"ban_status"`
				ErrCode     int32 `json:"err_code"`
				Duration    int64 `json:"duration"`
			}{}
			_ = json.Unmarshal([]byte(message.Content), &callingElem)

			resLen := len(result)
			currentIsComing := false
			friendId := ""
			if message.SendID == userID {
				currentIsComing = false
				friendId = message.RecvID
			} else {
				currentIsComing = true
				friendId = message.SendID
			}
			var same = false
			if callingElem.ErrCode == constant.CallingMissedCall && resLen > 0 {
				// if current message is same as previous, increase the calling count. otherwise add a new calling message
				// check fields: is coming, user id, calling type, error code.
				for _, index := range multiCount {
					same = (result[index].IsComing == currentIsComing) &&
						(result[index].UserID == friendId) &&
						(result[index].CallingType == callingElem.MessageType)
					preYear, preMonth, preDay := time.Unix(result[index].CallTime, 0).Date()
					curYear, curMonth, curDay := time.Unix(message.SendTime, 0).Date()
					if same && preYear == curYear && preMonth == curMonth && preDay == curDay {
						result[index].CallingCount++
						break
					}
				}
			}
			if !same {
				resMessage := CallingMessage{
					IsComing:     currentIsComing,
					CallingCount: 1,
					CallTime:     message.SendTime,
					CallingType:  callingElem.MessageType,
					ErrCode:      callingElem.ErrCode,
				}
				if friend, ok := friendMap[friendId]; ok {
					resMessage.ProfilePhoto = friend.FaceURL
					resMessage.Username = friend.Nickname
					resMessage.UserID = friendId
					if callingElem.ErrCode == constant.CallingMissedCall {
						multiCount = append(multiCount, resLen)
					}
					result = append(result, resMessage)
				} else {
					// deleted friend
					continue
				}
			}
		}

		callback.OnSuccess(utils.StructToJsonString(result))
	}()

}

// func (c *Conversation) GetUserIPandStatus(callback open_im_sdk_callback.Base, params string, operationID string) {
// 	if callback == nil {
// 		return
// 	}
// 	log.Error("Sandman & 1")
// 	fName := utils.GetSelfFuncName()
// 	go func() {
// 		log.NewInfo(operationID, fName, "args: ", params)
// 		var unmarshalParams server_api_params.GetUserIPAndStatusReq
// 		log.Error("Sandman & 2")
// 		common.JsonUnmarshalAndArgsValidate(params, &unmarshalParams, callback, operationID)
// 		log.Error("Sandman & 3")
// 		response := c.processGetUserIPandStatus(callback, unmarshalParams, 1, operationID)
// 		callback.OnSuccess(utils.StructToJsonString(response))
// 	}()

// }

func (c *Conversation) initMsgStructByMsg(callback open_im_sdk_callback.SendMsgCallBack, s *sdk_struct.MsgStruct, message, operationID string, upload bool) (delFile []string) {
	if callback == nil {
		return
	}
	//log.NewError(operationID, utils.GetSelfFuncName(), "字符长度", len(s.Content), s.Content)
	if len(s.Content) > 10000 {
		common.CheckAnyErrCallback(callback, 201, errors.New("message too long:"+strconv.Itoa(len(s.Content))), operationID)
	}
	common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)

	s.SendID = c.loginUserID
	s.SenderPlatformID = c.platformID
	delFile = []string{}
	//media file handle
	if s.Status != constant.MsgStatusSendSuccess { //filter forward message
		switch s.ContentType {
		case constant.Picture:
			var sourcePath string
			if utils.FileExist(s.PictureElem.SourcePath) {
				sourcePath = s.PictureElem.SourcePath
				delFile = append(delFile, utils.FileTmpPath(s.PictureElem.SourcePath, c.DataDir))
			} else {
				sourcePath = utils.FileTmpPath(s.PictureElem.SourcePath, c.DataDir)
				delFile = append(delFile, sourcePath)
			}
			log.Info(operationID, "file", sourcePath, delFile)
			if upload {
				sourceUrl, uuid, err := c.UploadImage(sourcePath, s.ClientMsgID, callback.OnProgress)
				common.CheckAnyErrCallback(callback, 301, err, operationID)
				s.PictureElem.SourcePicture.Url = sourceUrl
				s.PictureElem.SourcePicture.UUID = uuid

				s.PictureElem.SnapshotPicture.Url = sourceUrl
				// s.PictureElem.SnapshotPicture.UUID = uuid
				// newWidth := int32(200)
				// resizeImageUrlGenrator(&s.PictureElem.SnapshotPicture, s.PictureElem.SourcePicture.Width, s.PictureElem.SourcePicture.Height, newWidth)
				s.PictureElem.SnapshotPicture.Url = sourceUrl + "?imageView2/2/w/" + constant.ZoomScale + "/h/" + constant.ZoomScale
			}
			s.PictureElem.SnapshotPicture.Width = int32(utils.StringToInt(constant.ZoomScale))
			s.PictureElem.SnapshotPicture.Height = int32(utils.StringToInt(constant.ZoomScale))
			s.Content = utils.StructToJsonString(s.PictureElem)

		case constant.Voice:
			var sourcePath string
			if utils.FileExist(s.SoundElem.SoundPath) {
				sourcePath = s.SoundElem.SoundPath
				delFile = append(delFile, utils.FileTmpPath(s.SoundElem.SoundPath, c.DataDir))
			} else {
				sourcePath = utils.FileTmpPath(s.SoundElem.SoundPath, c.DataDir)
				delFile = append(delFile, sourcePath)
			}
			log.Info(operationID, "file", sourcePath, delFile)
			if upload {
				soundURL, uuid, err := c.UploadSound(sourcePath, s.ClientMsgID, callback.OnProgress)
				common.CheckAnyErrCallback(callback, 301, err, operationID)
				s.SoundElem.SourceURL = soundURL
				s.SoundElem.UUID = uuid
				s.Content = utils.StructToJsonString(s.SoundElem)
			}
		case constant.Video:
			var videoPath string
			var snapPath string
			if utils.FileExist(s.VideoElem.VideoPath) {
				videoPath = s.VideoElem.VideoPath
				snapPath = s.VideoElem.SnapshotPath
				delFile = append(delFile, utils.FileTmpPath(s.VideoElem.VideoPath, c.DataDir))
				delFile = append(delFile, utils.FileTmpPath(s.VideoElem.SnapshotPath, c.DataDir))
			} else {
				videoPath = utils.FileTmpPath(s.VideoElem.VideoPath, c.DataDir)
				snapPath = utils.FileTmpPath(s.VideoElem.SnapshotPath, c.DataDir)
				delFile = append(delFile, videoPath)
				delFile = append(delFile, snapPath)
			}
			log.Info(operationID, "file: ", videoPath, snapPath, delFile)
			if upload {
				snapshotURL, snapshotUUID, videoURL, videoUUID, err := c.UploadVideo(videoPath, snapPath, "", callback.OnProgress)
				common.CheckAnyErrCallback(callback, 301, err, operationID)
				s.VideoElem.VideoURL = videoURL
				s.VideoElem.SnapshotUUID = snapshotUUID
				s.VideoElem.SnapshotURL = snapshotURL
				s.VideoElem.VideoUUID = videoUUID
			}
			s.Content = utils.StructToJsonString(s.VideoElem)
		case constant.File:
			if upload {
				fileURL, fileUUID, err := c.UploadFile(s.FileElem.FilePath, "", callback.OnProgress)
				common.CheckAnyErrCallback(callback, 301, err, operationID)
				s.FileElem.SourceURL = fileURL
				s.FileElem.UUID = fileUUID
			}
			s.Content = utils.StructToJsonString(s.FileElem)
		case constant.Article:
			s.Content = utils.StructToJsonString(s.ArticleElem)
		case constant.WalletTransfer:
			s.Content = utils.StructToJsonString(s.WalletTransferElem)
		case constant.OfficialAccount:
			s.Content = utils.StructToJsonString(s.OfficialAccountElem)
		case constant.Woom:
			s.Content = utils.StructToJsonString(s.WoomElem)
		case constant.Text:
		case constant.AtText:
		case constant.Location:
		case constant.Custom:
		case constant.Merger:
		case constant.Quote:
		case constant.Card:
		case constant.Face:
		case constant.AdvancedText:
		case constant.Calling:
		case constant.Emoji:
		default:
			common.CheckAnyErrCallback(callback, 202, errors.New("contentType not currently supported"+utils.Int32ToString(s.ContentType)), operationID)
		}
	}

	return delFile
}

func (c *Conversation) broadcastUpdateAndSendMsg(callback open_im_sdk_callback.SendMsgCallBack, s *sdk_struct.MsgStruct, process int, recvID, groupID, offlinePushInfo, operationID string) (recvName string) {

	ClientMsgID := utils.GetMsgID(s.SendID)
	s.ClientMsgID = ClientMsgID

	p := &sdk.OfflinePushInfo{}
	if offlinePushInfo == "" {
		p = nil
	} else {
		common.JsonUnmarshalAndArgsValidate(offlinePushInfo, &p, callback, operationID)
	}
	if recvID == "" && groupID == "" {
		common.CheckAnyErrCallback(callback, 201, errors.New("recvID && groupID not both null"), operationID)
	}
	var localMessage model_struct.LocalChatLog
	var conversationID string

	options := make(map[string]bool, 2)
	lc := &model_struct.LocalConversation{LatestMsgSendTime: s.CreateTime}

	//根据单聊群聊类型组装消息和会话
	if recvID == "" {
		s.GroupID = groupID
		lc.GroupID = groupID

		g, err := c.full.GetGroupInfoByGroupID(groupID)
		common.CheckAnyErrCallback(callback, 202, err, operationID)

		//check if the user is owner or admin
		memberInfo, err := c.group.GetOneGroupMemberInfo(groupID, c.loginUserID)
		if err != nil {
			//sync again myself
			c.group.SyncMySelfInTheGroup(groupID, operationID)
			//get group membe again
			memberInfo, err = c.group.GetOneGroupMemberInfo(groupID, c.loginUserID)
			if err != nil {
				common.CheckAnyErrCallback(callback, 301, err, operationID)
			}
		}

		if memberInfo.Nickname != "" {
			s.SenderNickname = memberInfo.Nickname
		}
		if memberInfo.MuteEndTime != 0 && int64(memberInfo.MuteEndTime) > time.Now().Unix() {
			common.CheckAnyErrCallback(callback, 301, errors.New("you are muted!"), operationID)
		}

		//check if the group was muted
		//ownerID, adminIDList, err := c.group.GetGroupOwnerIDAndAdminIDList(groupID, operationID)
		if g.Status != 0 {
			if memberInfo.RoleLevel == constant.GroupOrdinaryUsers {
				common.CheckAnyErrCallback(callback, 302, errors.New("group was invalid!"), operationID)
			}
		}

		//check if the group was dissolved
		if g.Status == 2 {
			common.CheckAnyErrCallback(callback, 302, errors.New("group was dissolved!"), operationID)
		}

		lc.ShowName = g.GroupName
		lc.FaceURL = g.FaceURL
		switch g.GroupType {
		case constant.NormalGroup:
			s.SessionType = constant.GroupChatType
			lc.ConversationType = constant.GroupChatType
			conversationID = utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
		case constant.SuperGroup, constant.WorkingGroup:
			s.SessionType = constant.SuperGroupChatType
			conversationID = utils.GetConversationIDBySessionType(groupID, constant.SuperGroupChatType)
			lc.ConversationType = constant.SuperGroupChatType
		}
		//c.SyncOneConversation(conversationID, operationID)
		//c.group.SyncGroupMemberByGroupID(groupID, operationID, false)

		oldConversation, err := c.db.GetConversation(conversationID)
		if err != nil {
			common.CheckAnyErrCallback(callback, 301, err, operationID)
		}
		if oldConversation.IsNotInGroup {
			common.CheckAnyErrCallback(callback, 301, errors.New("you are not in the group by checking oldConversation.IsNotInGroup!"), operationID)
		}

		//groupMemberUidList, err := c.db.GetGroupMemberUIDListByGroupID(groupID)
		//common.CheckAnyErrCallback(callback, 202, err, operationID)
		//if !utils.IsContain(s.SendID, groupMemberUidList) {
		//	common.CheckAnyErrCallback(callback, 208, errors.New("you not exist in this group"), operationID)
		//}
		s.AttachedInfoElem.GroupHasReadInfo.GroupMemberCount = g.MemberCount
		s.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
	} else {
		s.SessionType = constant.SingleChatType
		s.RecvID = recvID
		conversationID = utils.GetConversationIDBySessionType(recvID, constant.SingleChatType)
		lc.UserID = recvID
		lc.ConversationType = constant.SingleChatType
		//check if the friend of sender
		_, err := c.db.GetFriendInfoByFriendUserID(s.RecvID)
		if err != nil {
			msgStructToLocalChatLog(&localMessage, s)
			localMessage.ServerMsgID = localMessage.ClientMsgID
			localMessage.Seq = uint32(c.db.GetLocalMaxSeq())
			localMessage.Status = constant.MsgStatusSendFailed
			err := c.db.InsertMessage(&localMessage)

			if err != nil {
				log.Warn(operationID, "Local Message insert Fail", err.Error(), localMessage)
			}
			log.Warn(operationID, "Local Message insert Success", localMessage)
			common.CheckAnyErrCallback(callback, 302, errors.New("you are not friends"), operationID)
		}
		//faceUrl, name, err := c.friend.GetUserNameAndFaceUrlByUid(recvID, operationID)
		oldLc, err := c.db.GetConversation(conversationID)
		if err == nil && oldLc.IsPrivateChat {
			options[constant.IsNotPrivate] = false
			s.AttachedInfoElem.IsPrivateChat = true
			s.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
		}
		faceUrl, name, err := c.cache.GetUserNameAndFaceURL(recvID, operationID)
		if err != nil {
			common.CheckAnyErrCallback(callback, 301, err, operationID)
		}
		lc.FaceURL = faceUrl
		lc.ShowName = name

	}

	log.Warn(operationID, "before insert  message is ", s)
	oldMessage, err := c.db.GetMessageController(s)
	if err != nil {
		msgStructToLocalChatLog(&localMessage, s)
		err := c.db.InsertMessageController(&localMessage)
		common.CheckAnyErrCallback(callback, 201, err, operationID)
	} else {
		if oldMessage.Status != constant.MsgStatusSendFailed {
			common.CheckAnyErrCallback(callback, 202, errors.New("only failed message can be repeatedly send"), operationID)
		} else {
			s.Status = constant.MsgStatusSending
		}
	}
	lc.ConversationID = conversationID
	lc.LatestMsg = utils.StructToJsonString(s)
	log.Info(operationID, "send message come here", *lc)

	oldMessage, err = c.db.GetMessageController(s)
	if err != nil {
		log.Warn(operationID, "get message err")
	}
	log.Warn(operationID, "before update database message is ", oldMessage)
	if utils.IsContainInt(int(s.ContentType), []int{constant.Picture, constant.Voice, constant.Video, constant.File}) {
		msgStructToLocalChatLog(&localMessage, s)
		log.Warn(operationID, "update message is ", s, localMessage)
		err = c.db.UpdateMessageController(&localMessage)
		common.CheckAnyErrCallback(callback, 201, err, operationID)
	}

	c.sendMessageToServerForBroadcast(s, callback, p, options, operationID, process)

	return lc.ShowName

}

func (c *Conversation) SendBroadCastMessage(callback open_im_sdk_callback.SendMsgCallBack, message string, recvIDs, groupIDs []string, actionType, offlinePushInfo string, operationID string) {
	if callback == nil {
		return
	}
	userIDLen := len(recvIDs)
	groupIDLen := len(groupIDs)

	s := &sdk_struct.MsgStruct{}

	common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)
	upload := false
	msgTotalLen := userIDLen + groupIDLen
	if actionType == "1" {
		// forward
		if userIDLen > 0 && groupIDLen > 0 {
			callback.OnError(constant.ErrSendLimit.ErrCode, "Just can forward to users or groups")
			return
		}
		if msgTotalLen > 100 {
			callback.OnError(constant.ErrSendLimit.ErrCode, "User or group counts is too much")
			return
		}
	} else if actionType == "2" {
		// tools
		if msgTotalLen > 200 {
			callback.OnError(constant.ErrSendLimit.ErrCode, "The total counts of user and group should be less than 200")
			return
		}
		upload = true
	}
	statusReq := api.GetBroadcastStatusReq{}
	statusResp := api.GetBroadcastStatusResp{}
	statusReq.OperationID = operationID
	statusReq.UserID = s.SendID
	c.p.PostFatalCallback(callback, constant.GetBroadcastStatusRouter, statusReq, &statusResp, operationID)
	if statusResp.Status == 0 {
		// disable
		callback.OnError(constant.ErrAccess.ErrCode, "Broadcast is disable, please enable then try again")
		return
	}
	go func() {

		localDatabase, err := db2.NewDataBase(s.SendID, sdk_struct.SvrConf.DataDir)
		if err != nil {
			callback.OnError(constant.ErrServer.ErrCode, constant.ErrServer.ErrMsg)
			return
		}

		type sendInfo struct {
			RecvID   string `json:"recvID"`
			RecvName string `json:"recvName"`
			MsgType  int8   `json:"msgType"`
		}
		broadcastRecord := &model_struct.LocalBroadcast{}
		var sendUserList []sendInfo

		delFile := c.initMsgStructByMsg(callback, s, message, operationID, upload)

		for i := 0; i < groupIDLen; i++ {
			recvName := c.broadcastUpdateAndSendMsg(callback, s, (i+1)*100/msgTotalLen, "", groupIDs[i], offlinePushInfo, operationID)

			sendUserList = append(sendUserList, sendInfo{
				RecvID:   groupIDs[i],
				RecvName: recvName,
				MsgType:  2,
			})
		}

		for i := 0; i < userIDLen; i++ {
			recvName := c.broadcastUpdateAndSendMsg(callback, s, (i+1)*100/msgTotalLen, recvIDs[i], "", offlinePushInfo, operationID)

			sendUserList = append(sendUserList, sendInfo{
				RecvID:   recvIDs[i],
				RecvName: recvName,
				MsgType:  1,
			})
			log.Debug("", "Send to user: ", recvIDs[i])
		}

		messageForDB, err := json.Marshal(s)
		if err != nil {
			return
		}
		broadcastRecord.Receiver = utils.StructToJsonString(sendUserList)
		broadcastRecord.Message = string(messageForDB)
		broadcastRecord.SenderId = s.SendID
		broadcastRecord.ContentType = s.ContentType
		broadcastRecord.CreateTime = time.Now().Unix()

		err = localDatabase.InsertBroadcast(broadcastRecord)
		if err != nil {
			log.NewError("", "Insert broadcast failed: ", err.Error())
			callback.OnError(constant.ErrServer.ErrCode, constant.ErrServer.ErrMsg)
			return
		}

		s.Status = constant.MsgStatusSendSuccess
		callback.OnSuccess(utils.StructToJsonString(broadcastRecord))
		log.Debug(operationID, "callback OnSuccess", s.ClientMsgID, s.ServerMsgID)
		log.Debug(operationID, utils.GetSelfFuncName(), "test s.SendTime and createTime", s.SendTime, s.CreateTime)

		//remove media cache file
		for _, v := range delFile {
			err = os.Remove(v)
			if err != nil {
				log.Error(operationID, "remove failed,", err.Error(), v)
			}
			log.Debug(operationID, "remove file: ", v)
		}

	}()
}

func (c *Conversation) SendBroadCastMessageV3(callback open_im_sdk_callback.SendMsgCallBack, message string, recvIDs, groupIDs []string, actionType, offlinePushInfo string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		var singleChatReciverInvalidIDs = make([]string, 0)
		var groupChatReciverInvalidIDs = make([]string, 0)
		var localBroadcastMsgReceiverList = make([]model_struct.LocalBroadcastMsgReceiver, 0)
		sendMessageToServerStruct := sdk_struct.BroadcastMsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &sendMessageToServerStruct, callback, operationID)

		//log.NewError(operationID, utils.GetSelfFuncName(), "字符长度", len(sendMessageToServerStruct.Content), sendMessageToServerStruct.Content)
		if len(sendMessageToServerStruct.Content) > 10000 {
			common.CheckAnyErrCallback(callback, 201, errors.New("message too long:"+strconv.Itoa(len(sendMessageToServerStruct.Content))), operationID)
		}

		sendMessageToServerStruct.SendID = c.loginUserID
		sendMessageToServerStruct.SenderPlatformID = c.platformID
		p := &sdk.OfflinePushInfo{}
		if offlinePushInfo == "" {
			p = nil
		} else {
			common.JsonUnmarshalAndArgsValidate(offlinePushInfo, &p, callback, operationID)
		}

		common.JsonUnmarshalAndArgsValidate(message, &sendMessageToServerStruct, callback, operationID)

		groupIDs = make([]string, 0)
		// Broadcast Validation, Check
		userIDLen := len(recvIDs)
		groupIDLen := len(groupIDs)
		messageRecipientCount := userIDLen + groupIDLen
		log.Info("Sandman SdK Broadcast Recvers", recvIDs, groupIDs)
		if actionType == "1" {
			// tools
			if messageRecipientCount > 200 {
				callback.OnError(constant.ErrSendLimit.ErrCode, "The total counts of user and group should be less than 200")
				return
			}
		} else if actionType == "2" {
			// it's for Resend the Same Message to Same Users
			sendMessageToServerStruct.ClientMsgID = utils.GetMsgID(sendMessageToServerStruct.SendID)
			sendMessageToServerStruct.SendTime = time.Now().Unix()
			sendMessageToServerStruct.IsRead = false
			sendMessageToServerStruct.Status = constant.MsgStatusSending
		}

		//var localMessage model_struct.LocalChatLog0
		var conversationID string
		options := make(map[string]bool, 2)

		lc := c.getOneConversation(callback, "cUser", constant.BroadcastChatType, operationID)
		conversationID = lc.ConversationID
		lc.LatestMsgSendTime = sendMessageToServerStruct.CreateTime
		sendMessageToServerStruct.SessionType = constant.BroadcastChatType
		lc.ConversationType = constant.BroadcastChatType

		localBrMsg := model_struct.LocalBroadcastChatLog{}
		err := utils2.CopyStructFields(&localBrMsg, &sendMessageToServerStruct)
		localBrMsg.Status = constant.MsgStatusSending
		if err != nil {
			localBrMsg.ClientMsgID = sendMessageToServerStruct.ClientMsgID
		}
		err = c.db.InsertBroadcastMessage(&localBrMsg)
		if err != nil {
			log.Info("Broadcast Message not logged in system", recvIDs, groupIDs)
		}

		c.updateBroadcastMsgStatusAndTriggerConversation(sendMessageToServerStruct.ClientMsgID, "", sendMessageToServerStruct.CreateTime, constant.MsgStatusSending, &sendMessageToServerStruct, lc, operationID)
		for _, reviverID := range recvIDs {
			frnd, err := c.db.GetFriendInfoByFriendUserID(reviverID)
			if err != nil {
				log.Info("Sandman Not A friend", reviverID, err.Error())
				singleChatReciverInvalidIDs = append(singleChatReciverInvalidIDs, reviverID)
				continue
			}
			sendMessageToServerStruct.ReceiverUserIDs = append(sendMessageToServerStruct.ReceiverUserIDs, reviverID)
			if frnd != nil {
				var localBroadcastMsgReceiver = model_struct.LocalBroadcastMsgReceiver{}
				localBroadcastMsgReceiver.ClientMsgID = sendMessageToServerStruct.ClientMsgID
				localBroadcastMsgReceiver.ServerMsgID = ""
				localBroadcastMsgReceiver.ReceiverID = frnd.FriendUserID
				localBroadcastMsgReceiver.ReceiverNickname = frnd.Nickname
				localBroadcastMsgReceiver.ReceiverFaceURL = frnd.FaceURL
				localBroadcastMsgReceiver.Status = constant.MsgStatusSendSuccess
				localBroadcastMsgReceiver.CreateTime = time.Now().Unix()
				localBroadcastMsgReceiver.ReceiverType = constant.SingleChatType
				localBroadcastMsgReceiverList = append(localBroadcastMsgReceiverList, localBroadcastMsgReceiver)
			}

		}

		for _, reviverGroupID := range groupIDs {
			g, err := c.full.GetGroupInfoByGroupID(reviverGroupID)
			if err == nil {
				memberInfo, err := c.group.GetOneGroupMemberInfo(reviverGroupID, c.loginUserID)
				if err != nil {
					if g.Status != 0 {
						if memberInfo.RoleLevel == constant.GroupOrdinaryUsers {
							groupChatReciverInvalidIDs = append(groupChatReciverInvalidIDs, reviverGroupID)
							continue
						}
					}
					//check if the group was dissolved
					if g.Status == 2 {
						groupChatReciverInvalidIDs = append(groupChatReciverInvalidIDs, reviverGroupID)
						continue
					}
				}
				if g != nil {
					var localBroadcastMsgReceiver = model_struct.LocalBroadcastMsgReceiver{}
					localBroadcastMsgReceiver.ClientMsgID = sendMessageToServerStruct.ClientMsgID
					localBroadcastMsgReceiver.ServerMsgID = ""
					localBroadcastMsgReceiver.ReceiverID = g.GroupID
					localBroadcastMsgReceiver.ReceiverNickname = g.GroupName
					localBroadcastMsgReceiver.ReceiverFaceURL = g.FaceURL
					localBroadcastMsgReceiver.Status = constant.MsgStatusSendSuccess
					localBroadcastMsgReceiver.CreateTime = time.Now().Unix()
					localBroadcastMsgReceiver.ReceiverType = constant.GroupChatType
					localBroadcastMsgReceiverList = append(localBroadcastMsgReceiverList, localBroadcastMsgReceiver)
				}
				sendMessageToServerStruct.ReceiverGroupIDs = append(sendMessageToServerStruct.ReceiverGroupIDs, reviverGroupID)
			} else {
				groupChatReciverInvalidIDs = append(groupChatReciverInvalidIDs, reviverGroupID)
			}
		}

		err = c.db.InsertLocalBroadcastMsgReceiver(localBroadcastMsgReceiverList)
		if err != nil {
			log.Info("Broadcast Receiver not logged in system", recvIDs, groupIDs, err.Error())
		}

		//TODO Have to insert message to local DB, will do later to figure out where we have to insert it
		// because it dont need to shown in normal conversation for sender
		lc.ConversationID = conversationID
		lc.LatestMsg = utils.StructToJsonString(sendMessageToServerStruct)
		/////////////////
		//actionType == "1" is for sending a message, 2 is for resend. so no need to upload the content again for resend
		var delFile []string
		if sendMessageToServerStruct.Status != constant.MsgStatusSendSuccess { //filter forward message
			switch sendMessageToServerStruct.ContentType {
			case constant.Picture:
				if actionType == "1" {
					var sourcePath string
					if utils.FileExist(sendMessageToServerStruct.PictureElem.SourcePath) {
						sourcePath = sendMessageToServerStruct.PictureElem.SourcePath
						delFile = append(delFile, utils.FileTmpPath(sendMessageToServerStruct.PictureElem.SourcePath, c.DataDir))
					} else {
						sourcePath = utils.FileTmpPath(sendMessageToServerStruct.PictureElem.SourcePath, c.DataDir)
						delFile = append(delFile, sourcePath)
					}
					log.Info(operationID, "file", sourcePath, delFile)
					sourceUrl, uuid, err := c.UploadImage(sourcePath, "", callback.OnProgress)
					//TODO Need to Update below checkErrAndUpdateMessage according to local storage message logic
					c.checkErrAndUpdateBroadcastMessage(callback, 301, err, &sendMessageToServerStruct, lc, operationID)
					sendMessageToServerStruct.PictureElem.SourcePicture.Url = sourceUrl
					sendMessageToServerStruct.PictureElem.SourcePicture.UUID = uuid

					sendMessageToServerStruct.PictureElem.SnapshotPicture.Url = sourceUrl
					// sendMessageToServerStruct.PictureElem.SnapshotPicture.UUID = uuid
					// newWidth := int32(200)
					// resizeImageUrlGenrator(&sendMessageToServerStruct.PictureElem.SnapshotPicture, sendMessageToServerStruct.PictureElem.SourcePicture.Width, sendMessageToServerStruct.PictureElem.SourcePicture.Height, newWidth)
					sendMessageToServerStruct.PictureElem.SnapshotPicture.Url = sourceUrl + "?imageView2/2/w/" + constant.ZoomScale + "/h/" + constant.ZoomScale
					sendMessageToServerStruct.PictureElem.SnapshotPicture.Width = int32(utils.StringToInt(constant.ZoomScale))
					sendMessageToServerStruct.PictureElem.SnapshotPicture.Height = int32(utils.StringToInt(constant.ZoomScale))
					sendMessageToServerStruct.Content = utils.StructToJsonString(sendMessageToServerStruct.PictureElem)
				}

			case constant.Voice:
				if actionType == "1" {
					var sourcePath string
					if utils.FileExist(sendMessageToServerStruct.SoundElem.SoundPath) {
						sourcePath = sendMessageToServerStruct.SoundElem.SoundPath
						delFile = append(delFile, utils.FileTmpPath(sendMessageToServerStruct.SoundElem.SoundPath, c.DataDir))
					} else {
						sourcePath = utils.FileTmpPath(sendMessageToServerStruct.SoundElem.SoundPath, c.DataDir)
						delFile = append(delFile, sourcePath)
					}
					log.Info(operationID, "file", sourcePath, delFile)
					soundURL, uuid, err := c.UploadSound(sourcePath, "", callback.OnProgress)
					//TODO Need to Update below checkErrAndUpdateMessage according to local storage message logic
					c.checkErrAndUpdateBroadcastMessage(callback, 301, err, &sendMessageToServerStruct, lc, operationID)
					sendMessageToServerStruct.SoundElem.SourceURL = soundURL
					sendMessageToServerStruct.SoundElem.UUID = uuid
					sendMessageToServerStruct.Content = utils.StructToJsonString(sendMessageToServerStruct.SoundElem)
				}

			case constant.Video:
				if actionType == "1" {
					var videoPath string
					var snapPath string
					if utils.FileExist(sendMessageToServerStruct.VideoElem.VideoPath) {
						videoPath = sendMessageToServerStruct.VideoElem.VideoPath
						snapPath = sendMessageToServerStruct.VideoElem.SnapshotPath
						delFile = append(delFile, utils.FileTmpPath(sendMessageToServerStruct.VideoElem.VideoPath, c.DataDir))
						delFile = append(delFile, utils.FileTmpPath(sendMessageToServerStruct.VideoElem.SnapshotPath, c.DataDir))
					} else {
						videoPath = utils.FileTmpPath(sendMessageToServerStruct.VideoElem.VideoPath, c.DataDir)
						snapPath = utils.FileTmpPath(sendMessageToServerStruct.VideoElem.SnapshotPath, c.DataDir)
						delFile = append(delFile, videoPath)
						delFile = append(delFile, snapPath)
					}
					log.Info(operationID, "file: ", videoPath, snapPath, delFile)
					snapshotURL, snapshotUUID, videoURL, videoUUID, err := c.UploadVideo(videoPath, snapPath, "", callback.OnProgress)
					//TODO Need to Update below checkErrAndUpdateMessage according to local storage message logic
					c.checkErrAndUpdateBroadcastMessage(callback, 301, err, &sendMessageToServerStruct, lc, operationID)
					sendMessageToServerStruct.VideoElem.VideoURL = videoURL
					sendMessageToServerStruct.VideoElem.SnapshotUUID = snapshotUUID
					sendMessageToServerStruct.VideoElem.SnapshotURL = snapshotURL
					sendMessageToServerStruct.VideoElem.VideoUUID = videoUUID
					sendMessageToServerStruct.Content = utils.StructToJsonString(sendMessageToServerStruct.VideoElem)
				}
			case constant.File:
				if actionType == "1" {
					fileURL, fileUUID, err := c.UploadFile(sendMessageToServerStruct.FileElem.FilePath, "", callback.OnProgress)
					//TODO Need to Update below checkErrAndUpdateMessage according to local storage message logic
					c.checkErrAndUpdateBroadcastMessage(callback, 301, err, &sendMessageToServerStruct, lc, operationID)
					sendMessageToServerStruct.FileElem.SourceURL = fileURL
					sendMessageToServerStruct.FileElem.UUID = fileUUID
					sendMessageToServerStruct.Content = utils.StructToJsonString(sendMessageToServerStruct.FileElem)
				}
			case constant.Text:
			case constant.AtText:
			case constant.Location:
			case constant.Custom:
			case constant.Merger:
			case constant.Quote:
			case constant.Card:
			case constant.Face:
			case constant.AdvancedText:
			case constant.Article:
			case constant.WalletTransfer:
			case constant.OfficialAccount:
			case constant.Calling:
			case constant.Woom:
			case constant.Emoji:
			default:
				common.CheckAnyErrCallback(callback, 202, errors.New("contentType not currently supported"+utils.Int32ToString(sendMessageToServerStruct.ContentType)), operationID)
			}
			////TODO update the local message with the Uploaded file URL in local message
			//oldMessage, err := c.db.GetMessageController(&sendMessageToServerStruct)
			//if err != nil {
			//	log.Warn(operationID, "get message err")
			//}
			//log.Warn(operationID, "before update database message is ", *oldMessage)
			if utils.IsContainInt(int(sendMessageToServerStruct.ContentType), []int{constant.Picture, constant.Voice, constant.Video, constant.File}) && actionType == "1" {
				err := utils2.CopyStructFields(&localBrMsg, &sendMessageToServerStruct)
				localBrMsg.Status = constant.MsgStatusSending
				if err != nil {
					localBrMsg.ClientMsgID = sendMessageToServerStruct.ClientMsgID
				}
				log.Info(operationID, "update message is ", &sendMessageToServerStruct, localBrMsg)
				err = c.db.UpdateBroadcastMessageController(&localBrMsg)
				common.CheckAnyErrCallback(callback, 201, err, operationID)
			}
		}

		c.sendBroadcastMessageToServer(&sendMessageToServerStruct, lc, callback, delFile, p, options, operationID)
	}()

}

func (c *Conversation) GetBroadcastMessageList(callback open_im_sdk_callback.Base, pageNumber int32, operationID string) {
	if callback == nil {
		return
	}
	if pageNumber == 0 {
		pageNumber = 1
	}

	broadcasts, err := c.db.GetBroadcastMessageListDB(pageNumber)
	if err != nil {
		errMsg := constant.ErrServer.ErrMsg + "  " + err.Error()
		callback.OnError(constant.ErrServer.ErrCode, errMsg)
		return
	}
	callback.OnSuccess(utils.StructToJsonString(broadcasts))

}

func (c *Conversation) GetBroadcastMessageReceiverList(callback open_im_sdk_callback.Base, clientMsgID string, operationID string) {
	if callback == nil {
		return
	}

	broadcasts, err := c.db.GetBroadcastMessageReceiverListDB(clientMsgID)
	if err != nil {
		errMsg := constant.ErrServer.ErrMsg + "  " + err.Error()
		callback.OnError(constant.ErrServer.ErrCode, errMsg)
		return
	}
	callback.OnSuccess(utils.StructToJsonString(broadcasts))

}

func (c *Conversation) GetBroadcastList(callback open_im_sdk_callback.Base, userID, operationID string) {
	if callback == nil {
		return
	}

	localDatabase, err := db2.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
	if err != nil {
		callback.OnError(constant.ErrServer.ErrCode, constant.ErrServer.ErrMsg)
		return
	}
	broadcasts, err := localDatabase.GetBroadcastList(userID)
	if err != nil {
		errMsg := constant.ErrServer.ErrMsg + "  " + err.Error()
		callback.OnError(constant.ErrServer.ErrCode, errMsg)
		return
	}
	callback.OnSuccess(utils.StructToJsonString(broadcasts))

}

func (c *Conversation) ClearBroadcast(callback open_im_sdk_callback.Base, userID, operationID string) {
	if callback == nil {
		return
	}

	conversationID := utils.GetConversationIDBySessionType("cUser", constant.BroadcastChatType)
	err := c.db.DeleteConversation(conversationID)
	if err != nil {
		return
	}
	err = c.db.ClearBroadcastList(userID)
	if err != nil {
		errMsg := constant.ErrServer.ErrMsg + "  " + err.Error()
		callback.OnError(constant.ErrServer.ErrCode, errMsg)
		return
	}
	callback.OnSuccess(userID)
	_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())

}

// update unread count in local database
func updateConversationUnreadCountInLocal(operationID, senderID, recvID, groupID string) error {
	log.NewInfo(operationID, utils.GetSelfFuncName(), "args:", senderID, recvID, groupID)
	if senderID == "" {
		log.NewError(operationID, utils.GetSelfFuncName(), "senderID is nil!")
		return errors.New("senderID is nil!")
	}
	if recvID == "" && groupID == "" {
		log.NewError(operationID, utils.GetSelfFuncName(), "recvID and groupID is nil!")
		return errors.New("recvID and groupID is nil!")
	}
	if recvID != "" {
		//only update receiver data in private chat
		db, err := db2.NewDataBase(recvID, sdk_struct.SvrConf.DataDir)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			return err
		}
		//conversation, err := db.GetConversation(conversationID)
		//conversation.UnreadCount += 1
		return db.IncrConversationUnreadCount(utils.GetConversationIDBySessionType(senderID, constant.SingleChatType))
	} else {
		//update all members data in group chat except sender
		db, err := db2.NewDataBase(senderID, sdk_struct.SvrConf.DataDir)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			return err
		}
		membersList, err := db.GetGroupMemberListByGroupID(groupID)
		if err != nil {
			log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
			return err
		}
		for _, member := range membersList {
			go func(userID, conversationID string) {
				db2, err := db2.NewDataBase(userID, sdk_struct.SvrConf.DataDir)
				if err == nil {
					db2.IncrConversationUnreadCount(conversationID)
				} else {
					log.NewError(operationID, utils.GetSelfFuncName(), err.Error())
				}
			}(member.UserID, utils.GetConversationIDBySessionType(groupID, constant.GroupChatType))
		}
		return nil
	}
}

func (c *Conversation) SendMessageNotOss(callback open_im_sdk_callback.SendMsgCallBack, message, recvID, groupID string, offlinePushInfo string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		s := sdk_struct.MsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)
		s.SendID = c.loginUserID
		s.SenderPlatformID = c.platformID
		p := &sdk.OfflinePushInfo{}
		if offlinePushInfo == "" {
			p = nil
		} else {
			common.JsonUnmarshalAndArgsValidate(offlinePushInfo, &p, callback, operationID)
		}
		if recvID == "" && groupID == "" {
			common.CheckAnyErrCallback(callback, 201, errors.New("recvID && groupID not both null"), operationID)
		}
		var localMessage model_struct.LocalChatLog
		var conversationID string
		options := make(map[string]bool, 2)
		lc := &model_struct.LocalConversation{LatestMsgSendTime: s.CreateTime}
		//根据单聊群聊类型组装消息和会话
		if recvID == "" {
			g, err := c.full.GetGroupInfoByGroupID(groupID)
			common.CheckAnyErrCallback(callback, 202, err, operationID)
			lc.ShowName = g.GroupName
			lc.FaceURL = g.FaceURL
			switch g.GroupType {
			case constant.NormalGroup:
				s.SessionType = constant.GroupChatType
				lc.ConversationType = constant.GroupChatType
				conversationID = utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
			case constant.SuperGroup, constant.WorkingGroup:
				s.SessionType = constant.SuperGroupChatType
				conversationID = utils.GetConversationIDBySessionType(groupID, constant.SuperGroupChatType)
				lc.ConversationType = constant.SuperGroupChatType
			}
			s.GroupID = groupID
			lc.GroupID = groupID
			gm, err := c.db.GetGroupMemberInfoByGroupIDUserID(groupID, c.loginUserID)
			if err == nil && gm != nil {
				log.Debug(operationID, "group chat test", *gm)
				if gm.Nickname != "" {
					s.SenderNickname = gm.Nickname
				}
			}
			//groupMemberUidList, err := c.db.GetGroupMemberUIDListByGroupID(groupID)
			//common.CheckAnyErrCallback(callback, 202, err, operationID)
			//if !utils.IsContain(s.SendID, groupMemberUidList) {
			//	common.CheckAnyErrCallback(callback, 208, errors.New("you not exist in this group"), operationID)
			//}
			s.AttachedInfoElem.GroupHasReadInfo.GroupMemberCount = g.MemberCount
			s.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
		} else {
			s.SessionType = constant.SingleChatType
			s.RecvID = recvID
			conversationID = utils.GetConversationIDBySessionType(recvID, constant.SingleChatType)
			lc.UserID = recvID
			lc.ConversationType = constant.SingleChatType
			//faceUrl, name, err := c.friend.GetUserNameAndFaceUrlByUid(recvID, operationID)
			oldLc, err := c.db.GetConversation(conversationID)
			if err == nil && oldLc.IsPrivateChat {
				options[constant.IsNotPrivate] = false
				s.AttachedInfoElem.IsPrivateChat = true
				s.AttachedInfo = utils.StructToJsonString(s.AttachedInfoElem)
			}
			if err != nil {
				faceUrl, name, err := c.cache.GetUserNameAndFaceURL(recvID, operationID)
				common.CheckAnyErrCallback(callback, 301, err, operationID)
				lc.FaceURL = faceUrl
				lc.ShowName = name
			}

		}
		oldMessage, err := c.db.GetMessageController(&s)
		if err != nil {
			msgStructToLocalChatLog(&localMessage, &s)
			err := c.db.InsertMessageController(&localMessage)
			common.CheckAnyErrCallback(callback, 201, err, operationID)
		} else {
			if oldMessage.Status != constant.MsgStatusSendFailed {
				common.CheckAnyErrCallback(callback, 202, errors.New("only failed message can be repeatedly send"), operationID)
			} else {
				s.Status = constant.MsgStatusSending
			}
		}
		lc.ConversationID = conversationID
		lc.LatestMsg = utils.StructToJsonString(s)
		//u.doUpdateConversation(common.cmd2Value{Value: common.updateConNode{conversationID, constant.AddConOrUpLatMsg,
		//c}})
		//u.doUpdateConversation(cmd2Value{Value: updateConNode{"", ConChange, []string{conversationID}}})
		//_ = u.triggerCmdUpdateConversation(updateConNode{conversationID, ConChange, ""})
		var delFile []string
		if utils.IsContainInt(int(s.ContentType), []int{constant.Picture, constant.Voice, constant.Video, constant.File}) {
			msgStructToLocalChatLog(&localMessage, &s)
			err = c.db.UpdateMessageController(&localMessage)
			common.CheckAnyErrCallback(callback, 201, err, operationID)
		}
		c.sendMessageToServer(&s, lc, callback, delFile, p, options, operationID)
	}()
}

func (c *Conversation) InternalSendMessage(callback open_im_sdk_callback.Base, s *sdk_struct.MsgStruct, recvID, groupID, operationID string, p *sdk.OfflinePushInfo, onlineUserOnly bool, options map[string]bool) (*sdk.UserSendMsgResp, error) {
	if recvID == "" && groupID == "" {
		common.CheckAnyErrCallback(callback, 201, errors.New("recvID && groupID not both null"), operationID)
	}
	//t := time.Now()
	if recvID == "" {
		g, err := c.full.GetGroupInfoByGroupID(groupID)
		common.CheckAnyErrCallback(callback, 202, err, operationID)
		switch g.GroupType {
		case constant.NormalGroup:
			s.SessionType = constant.GroupChatType
			groupMemberUidList, err := c.db.GetGroupMemberUIDListByGroupID(groupID)
			common.CheckAnyErrCallback(callback, 202, err, operationID)
			if !utils.IsContain(s.SendID, groupMemberUidList) {
				common.CheckAnyErrCallback(callback, 208, errors.New("you not exist in this group"), operationID)
			}

		case constant.SuperGroup:
			s.SessionType = constant.SuperGroupChatType
		case constant.WorkingGroup:
			s.SessionType = constant.SuperGroupChatType
			groupMemberUidList, err := c.db.GetGroupMemberUIDListByGroupID(groupID)
			common.CheckAnyErrCallback(callback, 202, err, operationID)
			if !utils.IsContain(s.SendID, groupMemberUidList) {
				common.CheckAnyErrCallback(callback, 208, errors.New("you not exist in this group"), operationID)
			}
		}
		s.GroupID = groupID

	} else {
		s.SessionType = constant.SingleChatType
		s.RecvID = recvID
	}

	if onlineUserOnly {
		options[constant.IsHistory] = false
		options[constant.IsPersistent] = false
		options[constant.IsOfflinePush] = false
		options[constant.IsSenderSync] = false
	}

	var wsMsgData sdk.MsgData
	copier.Copy(&wsMsgData, s)
	wsMsgData.Content = []byte(s.Content)
	wsMsgData.CreateTime = s.CreateTime
	wsMsgData.Options = options
	wsMsgData.OfflinePushInfo = p
	timeout := 15
	retryTimes := 0
	if s.ContentType == constant.Picture || s.ContentType == constant.Video || s.ContentType == constant.Voice || s.ContentType == constant.File {
		timeout = 30
		retryTimes = 0
	}
	g, err := c.SendReqWaitResp(&wsMsgData, constant.WSSendMsg, timeout, retryTimes, c.loginUserID, operationID)
	switch e := err.(type) {
	case *constant.ErrInfo:
		common.CheckAnyErrCallback(callback, e.ErrCode, e, operationID)
	default:
		common.CheckAnyErrCallback(callback, 301, err, operationID)
	}
	var sendMsgResp sdk.UserSendMsgResp
	_ = proto.Unmarshal(g.Data, &sendMsgResp)
	return &sendMsgResp, nil

}

func (c *Conversation) sendMessageToServerForBroadcast(s *sdk_struct.MsgStruct, callback open_im_sdk_callback.SendMsgCallBack,
	offlinePushInfo *sdk.OfflinePushInfo, options map[string]bool, operationID string, process int) {
	log.Debug(operationID, "sendMessageToServer ", s.ServerMsgID, " ", s.ClientMsgID)
	//Protocol conversion
	var wsMsgData sdk.MsgData
	copier.Copy(&wsMsgData, s)
	wsMsgData.Content = []byte(s.Content)
	wsMsgData.CreateTime = s.CreateTime
	wsMsgData.Options = options
	wsMsgData.AtUserIDList = s.AtElem.AtUserList
	wsMsgData.OfflinePushInfo = offlinePushInfo
	timeout := 15
	retryTimes := 2
	if s.ContentType == constant.Picture || s.ContentType == constant.Video || s.ContentType == constant.Voice || s.ContentType == constant.File {
		timeout = 45
		retryTimes = 0
	}
	resp, err := c.SendReqWaitResp(&wsMsgData, constant.WSSendMsg, timeout, retryTimes, c.loginUserID, operationID)
	switch e := err.(type) {
	case *constant.ErrInfo:
		common.CheckAnyErrCallback(callback, e.ErrCode, e, operationID)
	default:
		common.CheckAnyErrCallback(callback, 302, err, operationID)
	}
	var sendMsgResp sdk.UserSendMsgResp
	_ = proto.Unmarshal(resp.Data, &sendMsgResp)
	s.SendTime = sendMsgResp.SendTime
	s.Status = constant.MsgStatusSending
	s.ServerMsgID = sendMsgResp.ServerMsgID
	callback.OnProgress(process)

}

func (c *Conversation) sendMessageToServer(s *sdk_struct.MsgStruct, lc *model_struct.LocalConversation, callback open_im_sdk_callback.SendMsgCallBack,
	delFile []string, offlinePushInfo *sdk.OfflinePushInfo, options map[string]bool, operationID string) {
	log.Debug(operationID, "sendMessageToServer ", s.ServerMsgID, " ", s.ClientMsgID)
	//Protocol conversion
	var wsMsgData sdk.MsgData
	copier.Copy(&wsMsgData, s)
	wsMsgData.Content = []byte(s.Content)
	wsMsgData.CreateTime = s.CreateTime
	wsMsgData.Options = options
	wsMsgData.AtUserIDList = s.AtElem.AtUserList
	wsMsgData.OfflinePushInfo = offlinePushInfo
	timeout := 15
	retryTimes := 2
	if s.ContentType == constant.Picture || s.ContentType == constant.Video || s.ContentType == constant.Voice || s.ContentType == constant.File {
		timeout = 45
		retryTimes = 0
	}
	resp, err := c.SendReqWaitResp(&wsMsgData, constant.WSSendMsg, timeout, retryTimes, c.loginUserID, operationID)
	switch e := err.(type) {
	case *constant.ErrInfo:
		c.checkErrAndUpdateMessage(callback, e.ErrCode, e, s, lc, operationID)
	default:
		c.checkErrAndUpdateMessage(callback, 302, err, s, lc, operationID)
	}
	var sendMsgResp sdk.UserSendMsgResp
	_ = proto.Unmarshal(resp.Data, &sendMsgResp)
	s.SendTime = sendMsgResp.SendTime
	s.Status = constant.MsgStatusSendSuccess
	s.ServerMsgID = sendMsgResp.ServerMsgID
	callback.OnProgress(100)
	callback.OnSuccess(utils.StructToJsonString(s))
	log.Debug(operationID, "callback OnSuccess", s.ClientMsgID, s.ServerMsgID)
	log.Debug(operationID, utils.GetSelfFuncName(), "test s.SendTime and createTime", s.SendTime, s.CreateTime)
	//remove media cache file
	for _, v := range delFile {
		err := os.Remove(v)
		if err != nil {
			log.Error(operationID, "remove failed,", err.Error(), v)
		}
		log.Debug(operationID, "remove file: ", v)
	}
	c.updateMsgStatusAndTriggerConversation(sendMsgResp.ClientMsgID, sendMsgResp.ServerMsgID, sendMsgResp.SendTime, constant.MsgStatusSendSuccess, s, lc, operationID)

}

func (c *Conversation) sendBroadcastMessageToServer(s *sdk_struct.BroadcastMsgStruct, lc *model_struct.LocalConversation, callback open_im_sdk_callback.SendMsgCallBack,
	delFile []string, offlinePushInfo *sdk.OfflinePushInfo, options map[string]bool, operationID string) {
	log.Debug(operationID, "sendMessageToServer ", s.ServerMsgID, " ", s.ClientMsgID)
	//Protocol conversion
	var wsMsgData sdk.MsgData
	copier.Copy(&wsMsgData, s)
	wsMsgData.Content = []byte(s.Content)
	wsMsgData.CreateTime = s.CreateTime
	wsMsgData.Options = options
	wsMsgData.AtUserIDList = s.AtElem.AtUserList
	wsMsgData.OfflinePushInfo = offlinePushInfo
	wsMsgData.BroadcastRecvIDs = append(wsMsgData.BroadcastRecvIDs, s.ReceiverUserIDs...)
	wsMsgData.BroadcastGroupIDs = append(wsMsgData.BroadcastGroupIDs, s.ReceiverGroupIDs...)
	timeout := 15
	retryTimes := 2
	if s.ContentType == constant.Picture || s.ContentType == constant.Video || s.ContentType == constant.Voice || s.ContentType == constant.File {
		timeout = 45
		retryTimes = 0

	}
	log.Debug(operationID, "Sandman sendBroadcastMessageToServer  Req", s.ReceiverUserIDs, " Group ", s.ReceiverGroupIDs)
	log.Debug(operationID, "Sandman sendBroadcastMessageToServer  ", wsMsgData.BroadcastRecvIDs, " Group ", wsMsgData.BroadcastGroupIDs)
	resp, err := c.SendReqWaitResp(&wsMsgData, constant.WSSendBroadcastMsg, timeout, retryTimes, c.loginUserID, operationID)
	if err != nil {
		c.checkErrAndUpdateBroadcastMessage(callback, 302, err, s, lc, operationID)
	}

	//s.SendTime = sendMsgResp.SendTime
	//s.Status = constant.MsgStatusSendSuccess
	//s.ServerMsgID = sendMsgResp.ServerMsgID
	callback.OnProgress(100)
	callback.OnSuccess(string(resp.Data))
	log.Debug(operationID, "callback OnSuccess", s.ClientMsgID, s.ServerMsgID)
	log.Debug(operationID, utils.GetSelfFuncName(), "test s.SendTime and createTime", s.SendTime, s.CreateTime)
	//remove media cache file
	for _, v := range delFile {
		err := os.Remove(v)
		if err != nil {
			log.Error(operationID, "remove failed,", err.Error(), v)
		}
		log.Debug(operationID, "remove file: ", v)
	}
	//TODO need to below function call back according to local Server Logic, this is also called from checkErrAndUpdateMessage() which we also need to fix before this step
	c.updateBroadcastMsgStatusAndTriggerConversation(s.ClientMsgID, s.ServerMsgID, s.SendTime, constant.MsgStatusSendSuccess, s, lc, operationID)

}

func (c *Conversation) CreateSoundMessageByURL(soundBaseInfo, operationID string) string {
	s := sdk_struct.MsgStruct{}
	var soundElem sdk_struct.SoundBaseInfo
	_ = json.Unmarshal([]byte(soundBaseInfo), &soundElem)
	s.SoundElem = soundElem
	c.initBasicInfo(&s, constant.UserMsgType, constant.Voice, operationID)
	s.Content = utils.StructToJsonString(s.SoundElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateSoundMessage(soundPath string, duration int64, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Voice, operationID)
	s.SoundElem.SoundPath = c.DataDir + soundPath
	s.SoundElem.Duration = duration
	fi, err := os.Stat(s.SoundElem.SoundPath)
	if err != nil {
		log.Error("internal", "get sound info err", err.Error())
		return ""
	}
	s.SoundElem.DataSize = fi.Size()
	s.Content = utils.StructToJsonString(s.SoundElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateVideoMessageByURL(videoBaseInfo, operationID string) string {
	s := sdk_struct.MsgStruct{}
	var videoElem sdk_struct.VideoBaseInfo
	_ = json.Unmarshal([]byte(videoBaseInfo), &videoElem)
	s.VideoElem = videoElem
	c.initBasicInfo(&s, constant.UserMsgType, constant.Video, operationID)
	s.Content = utils.StructToJsonString(s.VideoElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateVideoMessage(videoPath string, videoType string, duration int64, snapshotPath, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Video, operationID)
	s.VideoElem.VideoPath = c.DataDir + videoPath
	s.VideoElem.VideoType = videoType
	s.VideoElem.Duration = duration
	if snapshotPath == "" {
		s.VideoElem.SnapshotPath = ""
	} else {
		s.VideoElem.SnapshotPath = c.DataDir + snapshotPath
	}
	fi, err := os.Stat(s.VideoElem.VideoPath)
	if err != nil {
		log.Error("internal", "get video file error", err.Error())
		return ""
	}
	s.VideoElem.VideoSize = fi.Size()
	if snapshotPath != "" {
		imageInfo, err := getImageInfo(s.VideoElem.SnapshotPath)
		if err != nil {
			log.Error("internal", "get snapshot info ", err.Error())
			return ""
		}
		s.VideoElem.SnapshotHeight = imageInfo.Height
		s.VideoElem.SnapshotWidth = imageInfo.Width
		s.VideoElem.SnapshotSize = imageInfo.Size
	}
	s.Content = utils.StructToJsonString(s.VideoElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateFileMessageByURL(fileBaseInfo, operationID string) string {
	s := sdk_struct.MsgStruct{}
	var fileElem sdk_struct.FileBaseInfo
	_ = json.Unmarshal([]byte(fileBaseInfo), &fileElem)
	s.FileElem = fileElem
	c.initBasicInfo(&s, constant.UserMsgType, constant.File, operationID)
	s.Content = utils.StructToJsonString(s.FileElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateFileMessage(filePath string, fileName, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.File, operationID)
	s.FileElem.FilePath = c.DataDir + filePath
	s.FileElem.FileName = fileName
	fi, err := os.Stat(s.FileElem.FilePath)
	if err != nil {
		log.Error("internal", "get file message err", err.Error())
		return ""
	}
	s.FileElem.FileSize = fi.Size()
	s.Content = utils.StructToJsonString(s.FileElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateMergerMessage(messageList, title, summaryList, operationID string) string {
	var messages []*sdk_struct.MsgStruct
	var summaries []string
	s := sdk_struct.MsgStruct{}
	err := json.Unmarshal([]byte(messageList), &messages)
	if err != nil {
		log.Error("internal", "messages Unmarshal err", err.Error())
		return ""
	}
	_ = json.Unmarshal([]byte(summaryList), &summaries)
	c.initBasicInfo(&s, constant.UserMsgType, constant.Merger, operationID)
	s.MergeElem.AbstractList = summaries
	s.MergeElem.Title = title
	s.MergeElem.MultiMessage = messages
	s.Content = utils.StructToJsonString(s.MergeElem)
	return utils.StructToJsonString(s)
}
func (c *Conversation) CreateFaceMessage(index int, data, operationID string) string {
	s := sdk_struct.MsgStruct{}
	c.initBasicInfo(&s, constant.UserMsgType, constant.Face, operationID)
	s.FaceElem.Data = data
	s.FaceElem.Index = index
	s.Content = utils.StructToJsonString(s.FaceElem)
	return utils.StructToJsonString(s)

}
func (c *Conversation) CreateForwardMessage(m, operationID string) string {
	s := sdk_struct.MsgStruct{}
	err := json.Unmarshal([]byte(m), &s)
	if err != nil {
		log.Error("internal", "messages Unmarshal err", err.Error())
		return ""
	}
	if s.Status != constant.MsgStatusSendSuccess {
		log.Error("internal", "only send success message can be Forward")
		return ""
	}
	c.initBasicInfo(&s, constant.UserMsgType, s.ContentType, operationID)
	//Forward message seq is set to 0
	s.Seq = 0
	s.Status = constant.MsgStatusSendSuccess
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateWalletTransferMessage(Text, TransactionID, SenderID, Currency, operationID string, Amount string) string {
	s := sdk_struct.MsgStruct{}

	c.initBasicInfo(&s, constant.UserMsgType, constant.WalletTransfer, operationID)
	log.Debug("internal", "Create Wallet Transfer Message")
	s.WalletTransferElem.Content = Text
	s.WalletTransferElem.TransactionID = TransactionID
	s.WalletTransferElem.SenderID = SenderID
	s.WalletTransferElem.Amount = Amount
	s.WalletTransferElem.Currency = Currency

	s.Content = utils.StructToJsonString(s.WalletTransferElem)
	return utils.StructToJsonString(s)
}

func (c *Conversation) CreateOfficialAccountMessage(Text string, OfficialID int32, Nickname, FaceURL, Bio string, FollowTime, Type int32, operationID string) string {
	s := sdk_struct.MsgStruct{}

	c.initBasicInfo(&s, constant.UserMsgType, constant.OfficialAccount, operationID)
	s.OfficialAccountElem.Content = Text
	s.OfficialAccountElem.OfficialID = OfficialID
	s.OfficialAccountElem.Nickname = Nickname
	s.OfficialAccountElem.FaceURL = FaceURL
	s.OfficialAccountElem.Bio = Bio
	s.OfficialAccountElem.FollowTime = FollowTime
	s.OfficialAccountElem.Type = Type

	s.Content = utils.StructToJsonString(s.OfficialAccountElem)
	return utils.StructToJsonString(s)
}

func (c *Conversation) GetRecentForwardConversations(callback open_im_sdk_callback.Base, operationID string) {
	conversationList, err := c.db.GetRecentForwardMessagesConversationList()
	if err != nil {
		log.Error("internal", "converstations by forward counter", operationID, err.Error())
		return
	}
	callback.OnSuccess(utils.StructToJsonStringDefault(conversationList))

}

func (c *Conversation) FindMessageList(callback open_im_sdk_callback.Base, findMessageOptions, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		t := time.Now()
		log.NewInfo(operationID, "FindMessageList args: ", findMessageOptions)
		var unmarshalParams sdk_params_callback.FindMessageListParams
		common.JsonUnmarshalCallback(findMessageOptions, &unmarshalParams, callback, operationID)
		result := c.findMessageList(callback, unmarshalParams, operationID, false)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "FindMessageList callback: ", utils.StructToJsonStringDefault(result), "cost time", time.Since(t))
	}()
}
func (c *Conversation) GetHistoryMessageList(callback open_im_sdk_callback.Base, getMessageOptions, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		t := time.Now()
		log.NewInfo(operationID, "GetHistoryMessageList args: ", getMessageOptions)
		var unmarshalParams sdk_params_callback.GetHistoryMessageListParams
		common.JsonUnmarshalCallback(getMessageOptions, &unmarshalParams, callback, operationID)
		result := c.getHistoryMessageList(callback, unmarshalParams, operationID, false)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "cost time", time.Since(t), "GetHistoryMessageList callback: ", utils.StructToJsonStringDefault(result))
	}()
}
func (c *Conversation) GetHistoryMessageListByStartMsgId(callback open_im_sdk_callback.Base, getMessageOptions, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		t := time.Now()
		log.NewInfo(operationID, "GetHistoryMessageList args: ", getMessageOptions)
		var unmarshalParams sdk_params_callback.GetHistoryMessageListByStartMsgIdParams
		_ = common.JsonUnmarshalCallback(getMessageOptions, &unmarshalParams, callback, operationID)
		result := c.getHistoryMessageListByStartMsgId(callback, unmarshalParams, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "cost time", time.Since(t), "GetHistoryMessageList callback: ", utils.StructToJsonStringDefault(result))
	}()
}
func (c *Conversation) GetAdvancedHistoryMessageList(callback open_im_sdk_callback.Base, getMessageOptions, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		t := time.Now()
		log.NewInfo(operationID, "GetHistoryMessageList args: ", getMessageOptions)
		var unmarshalParams sdk_params_callback.GetAdvancedHistoryMessageListParams
		common.JsonUnmarshalCallback(getMessageOptions, &unmarshalParams, callback, operationID)
		result := c.getAdvancedHistoryMessageList(callback, unmarshalParams, operationID, false)
		if len(result.MessageList) == 0 {
			s := make([]*sdk_struct.MsgStruct, 0)
			result.MessageList = s
		}
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "cost time", time.Since(t), "GetHistoryMessageList callback: ", utils.StructToJsonStringDefault(result))
	}()
}

func (c *Conversation) GetHistoryMessageListReverse(callback open_im_sdk_callback.Base, getMessageOptions, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "GetHistoryMessageListReverse args: ", getMessageOptions)
		var unmarshalParams sdk_params_callback.GetHistoryMessageListParams
		common.JsonUnmarshalCallback(getMessageOptions, &unmarshalParams, callback, operationID)
		result := c.getHistoryMessageList(callback, unmarshalParams, operationID, true)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "GetHistoryMessageListReverse callback: ", utils.StructToJsonStringDefault(result))
	}()
}
func (c *Conversation) RevokeMessage(callback open_im_sdk_callback.Base, message string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "RevokeMessage args: ", message)
		var unmarshalParams sdk_params_callback.RevokeMessageParams
		common.JsonUnmarshalCallback(message, &unmarshalParams, callback, operationID)
		c.revokeOneMessage(callback, unmarshalParams, operationID)
		callback.OnSuccess(message)
		log.NewInfo(operationID, "RevokeMessage callback: ", sdk_params_callback.RevokeMessageCallback)
	}()
}
func (c *Conversation) NewRevokeMessage(callback open_im_sdk_callback.Base, message string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "RevokeMessage args: ", message)
		var unmarshalParams sdk_params_callback.RevokeMessageParams
		common.JsonUnmarshalCallback(message, &unmarshalParams, callback, operationID)
		c.newRevokeOneMessage(callback, unmarshalParams, operationID)
		callback.OnSuccess(message)
		log.NewInfo(operationID, "RevokeMessage callback: ", sdk_params_callback.RevokeMessageCallback)
	}()
}
func (c *Conversation) TypingStatusUpdate(callback open_im_sdk_callback.Base, recvID, msgTip, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "TypingStatusUpdate args: ", recvID, msgTip)
		c.typingStatusUpdate(callback, recvID, msgTip, operationID)
		callback.OnSuccess(sdk_params_callback.TypingStatusUpdateCallback)
		log.NewInfo(operationID, "TypingStatusUpdate callback: ", sdk_params_callback.TypingStatusUpdateCallback)
	}()
}

func (c *Conversation) MarkC2CMessageAsRead(callback open_im_sdk_callback.Base, userID string, msgIDList, operationID string) {
	if callback == nil {
		return
	}
	//c.Pool.Submit(
	//	func() {
	//		//log.NewError(operationID, "MarkC2CMessageAsRead args: ", userID, msgIDList)
	//		var unmarshalParams sdk_params_callback.MarkC2CMessageAsReadParams
	//		common.JsonUnmarshalCallback(msgIDList, &unmarshalParams, callback, operationID)
	//		if len(unmarshalParams) == 0 {
	//			conversationID := utils.GetConversationIDBySessionType(userID, constant.SingleChatType)
	//			c.setOneConversationUnread(callback, conversationID, 0, operationID)
	//			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UnreadCountSetZero}, c.GetCh())
	//			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
	//			callback.OnSuccess(sdk_params_callback.MarkC2CMessageAsReadCallback)
	//			return
	//		}
	//		c.markC2CMessageAsRead(callback, unmarshalParams, userID, operationID)
	//		callback.OnSuccess(sdk_params_callback.MarkC2CMessageAsReadCallback)
	//		log.NewInfo(operationID, "MarkC2CMessageAsRead callback: ", sdk_params_callback.MarkC2CMessageAsReadCallback)
	//	},
	//)
	//log.NewError("运行线程数:", c.pool.Running()) //当前池中待处理的任务数量
	//
	//gtimer.SetInterval(time.Second, func() {
	//	log.NewError("worker:", c.pool.Size()) //当前工作的协程数
	//	log.NewError("jobs:", c.pool.Jobs())   //当前池中待处理的任务数
	//})

	go func() {
		log.NewInfo(operationID, "MarkC2CMessageAsRead args: ", userID, msgIDList)
		var unmarshalParams sdk_params_callback.MarkC2CMessageAsReadParams
		common.JsonUnmarshalCallback(msgIDList, &unmarshalParams, callback, operationID)
		if len(unmarshalParams) == 0 {
			conversationID := utils.GetConversationIDBySessionType(userID, constant.SingleChatType)
			c.setOneConversationUnread(callback, conversationID, 0, operationID)
			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UnreadCountSetZero}, c.GetCh())
			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
			callback.OnSuccess(sdk_params_callback.MarkC2CMessageAsReadCallback)
			return
		}
		c.markC2CMessageAsRead(callback, unmarshalParams, userID, operationID)
		callback.OnSuccess(sdk_params_callback.MarkC2CMessageAsReadCallback)
		log.NewInfo(operationID, "MarkC2CMessageAsRead callback: ", sdk_params_callback.MarkC2CMessageAsReadCallback)
	}()

}

func (c *Conversation) MarkOfficialMessageAsRead(callback open_im_sdk_callback.Base, userID string, msgIDList, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "MarkC2CMessageAsRead args: ", userID, msgIDList)
		conversationID := utils.GetConversationIDBySessionType(userID, constant.OfficialArticlesChatType)
		c.setOneConversationUnread(callback, conversationID, 0, operationID)
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.NotificationCountSetZero}, c.GetCh())
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
		callback.OnSuccess(sdk_params_callback.MarkC2CMessageAsReadCallback)
		return
	}()

}

func (c *Conversation) MarkMessageAsReadByConID(callback open_im_sdk_callback.Base, conversationID, msgIDList, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "MarkMessageAsReadByConID args: ", conversationID, msgIDList)
		var unmarshalParams sdk_params_callback.MarkMessageAsReadByConIDParams
		common.JsonUnmarshalCallback(msgIDList, &unmarshalParams, callback, operationID)
		if len(unmarshalParams) == 0 {
			c.setOneConversationUnread(callback, conversationID, 0, operationID)
			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UnreadCountSetZero}, c.GetCh())
			_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
			callback.OnSuccess(sdk_params_callback.MarkMessageAsReadByConIDCallback)
			return
		}
		//c.markMessageAsReadByConID(callback, unmarshalParams, conversationID, operationID)
		callback.OnSuccess(sdk_params_callback.MarkMessageAsReadByConIDCallback)
		log.NewInfo(operationID, "MarkMessageAsReadByConID callback: ", sdk_params_callback.MarkMessageAsReadByConIDCallback)
	}()

}

// fixme
func (c *Conversation) MarkAllConversationHasRead(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		var lc model_struct.LocalConversation
		lc.UnreadCount = 0
		err := c.db.UpdateAllConversation(&lc)
		common.CheckDBErrCallback(callback, err, operationID)
		callback.OnSuccess("")

	}()
}

// deprecated
func (c *Conversation) MarkGroupMessageHasRead(callback open_im_sdk_callback.Base, groupID string, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		log.NewInfo(operationID, "MarkGroupMessageHasRead args: ", groupID)
		conversationID := utils.GetConversationIDBySessionType(groupID, constant.GroupChatType)
		c.setOneConversationUnread(callback, conversationID, 0, operationID)
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.UnreadCountSetZero}, c.GetCh())
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversationID, Action: constant.ConChange, Args: []string{conversationID}}, c.GetCh())
		callback.OnSuccess(sdk_params_callback.MarkGroupMessageHasReadCallback)
	}()
}
func (c *Conversation) MarkGroupMessageAsRead(callback open_im_sdk_callback.Base, groupID string, msgIDList, operationID string) {
	if callback == nil {
		return
	}
	//c.Pool.Submit(
	//	func() {
	//		log.NewInfo(operationID, "MarkGroupMessageAsRead args: ", groupID, msgIDList)
	//		var unmarshalParams sdk_params_callback.MarkGroupMessageAsReadParams
	//		common.JsonUnmarshalCallback(msgIDList, &unmarshalParams, callback, operationID)
	//		c.markGroupMessageAsRead(callback, unmarshalParams, groupID, operationID)
	//		callback.OnSuccess(sdk_params_callback.MarkGroupMessageAsReadCallback)
	//		log.NewInfo(operationID, "MarkGroupMessageAsRead callback: ", sdk_params_callback.MarkGroupMessageAsReadCallback)
	//	})
	go func() {
		log.NewInfo(operationID, "MarkGroupMessageAsRead args: ", groupID, msgIDList)
		var unmarshalParams sdk_params_callback.MarkGroupMessageAsReadParams
		common.JsonUnmarshalCallback(msgIDList, &unmarshalParams, callback, operationID)
		c.markGroupMessageAsRead(callback, unmarshalParams, groupID, operationID)
		callback.OnSuccess(sdk_params_callback.MarkGroupMessageAsReadCallback)
		log.NewInfo(operationID, "MarkGroupMessageAsRead callback: ", sdk_params_callback.MarkGroupMessageAsReadCallback)
	}()
}
func (c *Conversation) DeleteMessageFromLocalStorage(callback open_im_sdk_callback.Base, message string, operationID string) {
	go func() {
		s := sdk_struct.MsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)
		c.deleteMessageFromLocalStorage(callback, &s, operationID)
		callback.OnSuccess(message)
	}()
}

func (c *Conversation) ClearC2CHistoryMessage(callback open_im_sdk_callback.Base, userID string, operationID string) {
	go func() {
		c.clearC2CHistoryMessage(callback, userID, operationID)
		callback.OnSuccess(userID)

	}()
}
func (c *Conversation) ClearGroupHistoryMessage(callback open_im_sdk_callback.Base, groupID string, operationID string) {
	go func() {
		c.clearGroupHistoryMessage(callback, groupID, operationID)
		callback.OnSuccess(groupID)

	}()
}
func (c *Conversation) ClearC2CHistoryMessageFromLocalAndSvr(callback open_im_sdk_callback.Base, userID string, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", userID)
		conversationID := utils.GetConversationIDBySessionType(userID, constant.SingleChatType)
		c.deleteConversationAndMsgFromSvr(callback, conversationID, operationID)
		c.clearC2CHistoryMessage(callback, userID, operationID)
		callback.OnSuccess(userID)
	}()
}

// fixme
func (c *Conversation) ClearGroupHistoryMessageFromLocalAndSvr(callback open_im_sdk_callback.Base, groupID string, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", groupID)
		conversationID, _, err := c.getConversationTypeByGroupID(groupID)
		common.CheckDBErrCallback(callback, err, operationID)
		c.deleteConversationAndMsgFromSvr(callback, conversationID, operationID)
		c.clearGroupHistoryMessage(callback, groupID, operationID)
		callback.OnSuccess(groupID)
	}()
}

func (c *Conversation) InsertSingleMessageToLocalStorage(callback open_im_sdk_callback.Base, message, recvID, sendID, operationID string) {
	go func() {
		log.NewInfo(operationID, "InsertSingleMessageToLocalStorage args: ", message, recvID, sendID)
		if recvID == "" || sendID == "" {
			common.CheckAnyErrCallback(callback, 208, errors.New("recvID or sendID is null"), operationID)
		}
		var sourceID string
		s := sdk_struct.MsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)
		if sendID != c.loginUserID {
			faceUrl, name, err := c.cache.GetUserNameAndFaceURL(sendID, operationID)
			if err != nil {
				log.Error(operationID, "GetUserNameAndFaceURL err", err.Error(), sendID)
			}
			sourceID = sendID
			s.SenderFaceURL = faceUrl
			s.SenderNickname = name
		} else {
			sourceID = recvID
		}
		var conversation model_struct.LocalConversation
		conversation.ConversationID = utils.GetConversationIDBySessionType(sourceID, constant.SingleChatType)

		localMessage := model_struct.LocalChatLog{}
		s.SendID = sendID
		s.RecvID = recvID
		s.ClientMsgID = utils.GetMsgID(s.SendID)
		s.SendTime = utils.GetCurrentTimestampByMill()
		s.SessionType = constant.SingleChatType
		s.Status = constant.MsgStatusSendSuccess
		msgStructToLocalChatLog(&localMessage, &s)
		conversation.LatestMsg = utils.StructToJsonString(s)
		conversation.LatestMsgSendTime = s.SendTime
		conversation.FaceURL = s.SenderFaceURL
		conversation.ShowName = s.SenderNickname
		_ = c.insertMessageToLocalStorage(callback, &localMessage, operationID)
		callback.OnSuccess(utils.StructToJsonString(&s))
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversation.ConversationID, Action: constant.AddConOrUpLatMsg, Args: conversation}, c.GetCh())
	}()
}

func (c *Conversation) InsertGroupMessageToLocalStorage(callback open_im_sdk_callback.Base, message, groupID, sendID, operationID string) {
	go func() {
		log.NewInfo(operationID, "InsertSingleMessageToLocalStorage args: ", message, groupID, sendID)
		if groupID == "" || sendID == "" {
			common.CheckAnyErrCallback(callback, 208, errors.New("groupID or sendID is null"), operationID)
		}
		var conversation model_struct.LocalConversation
		var err error
		_, conversation.ConversationType, err = c.getConversationTypeByGroupID(groupID)
		common.CheckAnyErrCallback(callback, 202, err, operationID)
		conversation.ConversationID = utils.GetConversationIDBySessionType(groupID, int(conversation.ConversationType))
		s := sdk_struct.MsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)
		if sendID != c.loginUserID {
			faceUrl, name, _, err, isFromSvr := c.friend.GetUserNameAndFaceUrlByUid(sendID, operationID)
			if err != nil {
				log.Error(operationID, "getUserNameAndFaceUrlByUid err", err.Error(), sendID)
			}
			s.SenderFaceURL = faceUrl
			s.SenderNickname = name
			if isFromSvr {
				c.cache.Update(sendID, faceUrl, name)
			}
		}
		localMessage := model_struct.LocalChatLog{}
		s.SendID = sendID
		s.RecvID = groupID
		s.GroupID = groupID
		s.ClientMsgID = utils.GetMsgID(s.SendID)
		s.SendTime = utils.GetCurrentTimestampByMill()
		s.SessionType = conversation.ConversationType
		s.Status = constant.MsgStatusSendSuccess
		msgStructToLocalChatLog(&localMessage, &s)
		conversation.LatestMsg = utils.StructToJsonString(s)
		conversation.LatestMsgSendTime = s.SendTime
		conversation.FaceURL = s.SenderFaceURL
		conversation.ShowName = s.SenderNickname
		_ = c.insertMessageToLocalStorage(callback, &localMessage, operationID)
		callback.OnSuccess(utils.StructToJsonString(&s))
		_ = common.TriggerCmdUpdateConversation(common.UpdateConNode{ConID: conversation.ConversationID, Action: constant.AddConOrUpLatMsg, Args: conversation}, c.GetCh())
	}()

}

//modifyLocalMessages(callback open_im_sdk_callback.Base, message, groupID, sendID, operationID string)

func (c *Conversation) SetConversationStatus(callback open_im_sdk_callback.Base, operationID string, userID string, status int) {
	if callback == nil {
		log.Error(operationID, "callback is nil")
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", userID, status)
		//var unmarshalParams sdk.SetConversationStatusParams
		//common.JsonUnmarshalAndArgsValidate(userIDRemark, &unmarshalParams, callback, operationID)
		//f.setConversationStatus(unmarshalParams, callback, operationID)
		//callback.OnSuccess(utils.StructToJsonString(sdk.SetFriendRemarkCallback))
		//log.NewInfo(operationID, fName, " callback: ", utils.StructToJsonString(sdk.SetFriendRemarkCallback))
	}()
}

//	func (c *Conversation) FindMessages(callback common.Base, messageIDList string) {
//		go func() {
//			var c []string
//			err := json.Unmarshal([]byte(messageIDList), &c)
//			if err != nil {
//				callback.OnError(200, err.Error())
//				utils.sdkLog("Unmarshal failed, ", err.Error())
//
//			}
//			err, list := u.getMultipleMessageModel(c)
//			if err != nil {
//				callback.OnError(203, err.Error())
//			} else {
//				if list != nil {
//					callback.OnSuccess(utils.structToJsonString(list))
//				} else {
//					callback.OnSuccess(utils.structToJsonString([]utils.MsgStruct{}))
//				}
//			}
//		}()
//	}
func (c *Conversation) SearchLocalMessages(callback open_im_sdk_callback.Base, searchParam, operationID string) {
	if callback == nil {
		return
	}
	go func() {
		s := time.Now()
		log.NewInfo(operationID, "SearchLocalMessages args: ", searchParam)
		var unmarshalParams sdk_params_callback.SearchLocalMessagesParams
		common.JsonUnmarshalCallback(searchParam, &unmarshalParams, callback, operationID)
		unmarshalParams.KeywordList = utils.TrimStringList(unmarshalParams.KeywordList)
		result := c.searchLocalMessages(callback, unmarshalParams, operationID)
		callback.OnSuccess(utils.StructToJsonStringDefault(result))
		log.NewInfo(operationID, "cost time", time.Since(s))
		log.NewInfo(operationID, "SearchLocalMessages callback: ", result.TotalCount, len(result.SearchResultItems))
	}()
}
func getImageInfo(filePath string) (*sdk_struct.ImageInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, utils.Wrap(err, "open file err")
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, utils.Wrap(err, "image file  Decode err")
	}

	datatype, err := imgtype.Get(filePath)
	if err != nil {
		return nil, utils.Wrap(err, "image file  get type err")
	}
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, utils.Wrap(err, "image file  Stat err")
	}

	b := img.Bounds()

	return &sdk_struct.ImageInfo{int32(b.Max.X), int32(b.Max.Y), datatype, fi.Size()}, nil

}

const TimeOffset = 5

func (c *Conversation) initBasicInfo(message *sdk_struct.MsgStruct, msgFrom, contentType int32, operationID string) {
	message.CreateTime = utils.GetCurrentTimestampByMill()
	message.SendTime = message.CreateTime
	message.IsRead = false
	message.Status = constant.MsgStatusSending
	message.SendID = c.loginUserID
	userInfo, err := c.db.GetLoginUser()
	if err != nil {
		log.Error(operationID, "GetLoginUser", err.Error())
	} else {
		message.SenderFaceURL = userInfo.FaceURL
		message.SenderNickname = userInfo.Nickname
	}
	ClientMsgID := utils.GetMsgID(message.SendID)
	message.ClientMsgID = ClientMsgID
	message.MsgFrom = msgFrom
	message.ContentType = contentType
	message.SenderPlatformID = c.platformID

}

func (c *Conversation) DeleteConversationFromLocalAndSvr(callback open_im_sdk_callback.Base, conversationID string, operationID string) {
	defer func() {
		if r := recover(); r != nil {
			log.NewError("", "RunPing panic", " panic is ", r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "RunPing panic", "panic:", string(buf))
		}
	}()
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", conversationID)
		if c != nil {
			c.deleteConversationAndMsgFromSvr(callback, conversationID, operationID)
			c.deleteConversation(callback, conversationID, operationID)
			callback.OnSuccess(conversationID)
		}

		//log.NewInfo(operationID, fName, "callback: ", sdk_params_callback.DeleteConversationCallback)
	}()

}

func (c *Conversation) DeleteMessageFromLocalAndSvr(callback open_im_sdk_callback.Base, message string, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName, "args: ", message)
		s := sdk_struct.MsgStruct{}
		common.JsonUnmarshalAndArgsValidate(message, &s, callback, operationID)
		c.deleteMessageFromSvr(callback, &s, operationID)
		c.deleteMessageFromLocalStorage(callback, &s, operationID)
		callback.OnSuccess(message)
		log.NewInfo(operationID, fName, "callback: ", "")
	}()
}

func (c *Conversation) DeleteAllMsgFromLocalAndSvr(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName)
		//	c.deleteAllMsgFromSvr(callback, operationID)
		c.clearMessageFromSvr(callback, operationID)
		c.deleteAllMsgFromLocal(callback, operationID)
		callback.OnSuccess("")
		log.NewInfo(operationID, fName, "callback: ", "")
	}()
}

func (c *Conversation) DeleteAllMsgFromLocal(callback open_im_sdk_callback.Base, operationID string) {
	if callback == nil {
		return
	}
	fName := utils.GetSelfFuncName()
	go func() {
		log.NewInfo(operationID, fName)
		c.deleteAllMsgFromLocal(callback, operationID)
		callback.OnSuccess("")
		log.NewInfo(operationID, fName, "callback: ", "")
	}()
}
func (c *Conversation) getConversationTypeByGroupID(groupID string) (conversationID string, conversationType int32, err error) {
	g, err := c.full.GetGroupInfoByGroupID(groupID)
	if err != nil {
		return "", 0, utils.Wrap(err, "get group info error")
	}
	switch g.GroupType {
	case constant.NormalGroup:
		return utils.GetConversationIDBySessionType(groupID, constant.GroupChatType), constant.GroupChatType, nil
	case constant.SuperGroup, constant.WorkingGroup:
		return utils.GetConversationIDBySessionType(groupID, constant.SuperGroupChatType), constant.SuperGroupChatType, nil
	default:
		return "", 0, utils.Wrap(errors.New("err groupType"), "group type err")
	}
}

func (c *Conversation) TestSendMessage(callback open_im_sdk_callback.SendMsgCallBack, inputMap map[string]interface{}, operationID string) {
	offlinePushInfo := inputMap["offlinePushInfo"].(string)
	contentType := inputMap["contentType"].(int)
	recvID := inputMap["recvID"].(string)
	groupID := inputMap["groupID"].(string)

	switch contentType {
	case constant.Picture:
		msg := c.CreateImageMessageByURL(inputMap["sourcePicture"].(string), inputMap["bigPicture"].(string), inputMap["snapshotPicture"].(string), operationID)
		c.SendMessageNotOss(callback, msg, recvID, groupID, offlinePushInfo, operationID)
	case constant.Video:
		msg := c.CreateVideoMessageByURL(inputMap["videoBaseInfo"].(string), operationID)
		c.SendMessageNotOss(callback, msg, recvID, groupID, offlinePushInfo, operationID)
	case constant.Voice:
		msg := c.CreateSoundMessageByURL(inputMap["soundBaseInfo"].(string), operationID)
		c.SendMessageNotOss(callback, msg, recvID, groupID, offlinePushInfo, operationID)
	case constant.File:
		msg := c.CreateFileMessageByURL(inputMap["fileBaseInfo"].(string), operationID)
		c.SendMessageNotOss(callback, msg, recvID, groupID, offlinePushInfo, operationID)
	case constant.Text:
		txtMsg := inputMap["textMessage"].(string)
		if txtMsg != "" {
			msg := c.CreateTextMessage(txtMsg, operationID)
			c.SendMessage(callback, msg, recvID, groupID, offlinePushInfo, operationID)
		}
	case constant.AtText:
	case constant.Location:
	case constant.Custom:
	case constant.Merger:
	case constant.Quote:
	case constant.Card:
	case constant.Face:
	case constant.AdvancedText:
	case constant.Emoji:
	default:
		common.CheckAnyErrCallback(callback, 202, errors.New("contentType not currently supported"+utils.IntToString(contentType)), operationID)
	}
}
