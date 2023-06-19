package common

import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

type ObjectStorage interface {
	UploadImage(filePath string, clientMsgId string, onProgressFun func(int)) (string, string, error)
	UploadSound(filePath string, clientMsgId string, onProgressFun func(int)) (string, string, error)
	UploadFile(filePath string, clientMsgId string, onProgressFun func(int)) (string, string, error)
	UploadVideo(videoPath, snapshotPath string, clientMsgId string, onProgressFun func(int)) (string, string, string, string, error)
	StopUpload(clientMsgId string)
}
