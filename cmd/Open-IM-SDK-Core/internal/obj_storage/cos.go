package obj_storage

import (
	ws "Open_IM/cmd/Open-IM-SDK-Core/internal/interaction"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/pkg/common/log"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"

	//	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"path"
	"time"
)

type uploadFileInfo struct {
	FilePath string
	NewName  string
	FileType string
}

type COS struct {
	p                  *ws.PostApi
	fileMap            map[string]*os.File
	DeleteMsgIdChannel chan string
}

func NewCOS(p *ws.PostApi) *COS {
	cosTemp := &COS{p: p, fileMap: make(map[string]*os.File), DeleteMsgIdChannel: make(chan string, 200)}
	go func() {
		for {
			select {
			case removeMsgId := <-cosTemp.DeleteMsgIdChannel:
				if file, ok := cosTemp.fileMap[removeMsgId]; ok {
					if file != nil {
						err := file.Close()
						if err != nil {
							log.Error("", "the file has been closed")
							return
						}
					}
				}
			}
		}
	}()
	return cosTemp
}

func (c *COS) tencentCOSCredentials() (*server_api_params.TencentCloudStorageCredentialRespData, error) {
	req := server_api_params.TencentCloudStorageCredentialReq{OperationID: utils.OperationIDGenerator()}
	var resp server_api_params.TencentCloudStorageCredentialResp
	err := c.p.PostReturn(constant.TencentCloudStorageCredentialRouter, req, &resp.CosData)
	if err != nil {
		return nil, utils.Wrap(err, "")
	}
	return &resp.CosData, nil
}

func (c *COS) UploadImage(filePath string, clientMsgId string, onProgressFun func(int)) (string, string, error) {
	return c.uploadObj(filePath, "img", clientMsgId, onProgressFun)
	// return c.uploadObj(filePath, clientMsgId, "img", onProgressFun)
}

func (c *COS) UploadSound(filePath string, clientMsgId string, onProgressFun func(int)) (string, string, error) {
	return c.uploadObj(filePath, "", clientMsgId, onProgressFun)
}

func (c *COS) UploadFile(filePath string, clientMsgId string, onProgressFun func(int)) (string, string, error) {
	return c.uploadObj(filePath, "", clientMsgId, onProgressFun)
}

func (c *COS) UploadVideo(videoPath, snapshotPath string, clientMsgId string, onProgressFun func(int)) (string, string, string, string, error) {
	videoURL, videoUUID, err := c.uploadObj(videoPath, "", clientMsgId, onProgressFun)
	if err != nil {
		return "", "", "", "", utils.Wrap(err, "")
	}
	snapshotURL, snapshotUUID, err := c.uploadObj(snapshotPath, "img", clientMsgId, nil)
	if err != nil {
		return "", "", "", "", utils.Wrap(err, "")
	}
	return snapshotURL, snapshotUUID, videoURL, videoUUID, nil
}

func (c *COS) StopUpload(clientMsgId string) {
	c.DeleteMsgIdChannel <- clientMsgId
}

func (c *COS) getNewFileNameAndContentType(filePath string, fileType string) (string, string, error) {
	suffix := path.Ext(filePath)
	if len(suffix) == 0 {
		return "", "", utils.Wrap(errors.New("no suffix "), filePath)
	}
	newName := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), rand.Int(), suffix)
	contentType := ""
	if fileType == "img" {
		contentType = "image/" + suffix[1:]
	}
	return newName, contentType, nil
}

func (c *COS) getMD5FileName(filePath string, fileType string) (string, string, error) {
	suffix := path.Ext(filePath)
	if len(suffix) == 0 {
		return "", "", utils.Wrap(errors.New("no suffix "), filePath)
	}

	oFile, err := os.Open(filePath)
	if err != nil {
		err = errors.New("open file error")
		return "", "", utils.Wrap(err, "")
	}
	defer oFile.Close()
	md5h := md5.New()
	io.Copy(md5h, oFile)

	newName := hex.EncodeToString(md5h.Sum(nil)) + suffix
	contentType := ""
	if fileType == "img" {
		contentType = "image/" + suffix[1:]
	}
	return newName, contentType, nil
}

func (c *COS) uploadObj(filePath string, fileType string, clientMsgId string, onProgressFun func(int)) (string, string, error) {
	COSResp, err := c.tencentCOSCredentials()
	if err != nil {
		return "", "", utils.Wrap(err, "")
	}
	log.Info("upload ", COSResp.Credentials.SessionToken, "bucket ", COSResp.Credentials.TmpSecretID)
	dir := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", COSResp.Bucket, COSResp.Region)
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSResp.Credentials.TmpSecretID,
			SecretKey:    COSResp.Credentials.TmpSecretKey,
			SessionToken: COSResp.Credentials.SessionToken,
		},
	})
	if client == nil {
		err := errors.New("client == nil")
		return "", "", utils.Wrap(err, "")
	}

	newName, contentType, err := c.getMD5FileName(filePath, fileType)
	if err != nil {
		return "", "", utils.Wrap(err, "")
	}
	var lis = &selfListener{}
	lis.onProgressFun = onProgressFun
	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{Listener: lis, XOptionHeader: &http.Header{}}}
	if fileType == "img" {
		opt.ContentType = contentType
	}
	opt.XOptionHeader.Add("x-cos-tagging", "delete:1")
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", utils.Wrap(err, "")
	}
	if clientMsgId != "" {
		c.fileMap[clientMsgId] = file
	}
	_, err = client.Object.Put(context.Background(), newName, file, opt)
	if clientMsgId != "" {
		delete(c.fileMap, clientMsgId)
	}
	// _, err = client.Object.PutFromFile(context.Background(), newName, filePath, opt)
	if err != nil {
		return "", "", utils.Wrap(err, "")
	}
	return dir + "/" + newName, newName, nil
}

type selfListener struct {
	onProgressFun func(int)
}

func (l *selfListener) ProgressChangedCallback(event *cos.ProgressEvent) {
	switch event.EventType {
	case cos.ProgressDataEvent:
		if event.ConsumedBytes == event.TotalBytes {
			if l.onProgressFun != nil {
				l.onProgressFun(int((event.ConsumedBytes - 1) * 100 / event.TotalBytes))
			}

		} else {
			if l.onProgressFun != nil {
				log.Debug("", "progress: ", int(event.ConsumedBytes*100/event.TotalBytes))
				l.onProgressFun(int(event.ConsumedBytes * 100 / event.TotalBytes))
			}
		}
	case cos.ProgressFailedEvent:
	}
}
