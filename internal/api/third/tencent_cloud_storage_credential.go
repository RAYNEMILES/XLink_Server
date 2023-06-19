package apiThird

import (
	api "Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"image"
	"mime/multipart"
	"net/url"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/structs"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

type uploadFileInfo struct {
	FilePath        string
	NewName         string
	FileType        int
	File            *multipart.FileHeader
	SnapShot        *multipart.FileHeader
	SnapshotNewName string
	ImageWidth      int
	ImageHeight     int
}

type selfListener struct {
	MultiFilesTotalBytes int64
	ConsumedTotalBytes   int64
	mutex                sync.Mutex
	onProgressFun        func(int)
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
						"name/cos:HeadObject",
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

func TencentCloudStorageCredential(c *gin.Context) {
	req := api.TencentCloudStorageCredentialReq{}
	if err := c.BindJSON(&req); err != nil {
		log.NewError("0", "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}

	res, err := getTencentCloudCredential()
	resp := api.TencentCloudStorageCredentialResp{}
	if err != nil {
		resp.ErrCode = constant.ErrTencentCredential.ErrCode
		resp.ErrMsg = err.Error()
	} else {
		resp.CosData.Bucket = config.Config.Credential.Tencent.Bucket
		resp.CosData.Region = config.Config.Credential.Tencent.Region
		if config.Config.Credential.Tencent.Accelerate {
			resp.CosData.Region = "accelerate"
		}
		resp.CosData.CredentialResult = res
	}

	resp.Data = structs.Map(&resp.CosData)
	log.NewInfo(req.OperationID, "TencentCloudStorageCredential return ", resp)

	c.JSON(http.StatusOK, resp)
}

func TencentCloudUploadFile(c *gin.Context) {
	var (
		req  api.TencentCloudUploadFileReq
		resp api.TencentCloudUploadFileResp
	)
	defer func() {
		if r := recover(); r != nil {
			log.NewError(req.OperationID, "panic is", utils.GetSelfFuncName(), r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "Tencent Cloud UploadFile panic", "panic:", string(buf))
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file or snapShot args"})
			return
		}
	}()
	if err := c.ShouldBind(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.Debug(req.OperationID, "Tencent cloud upload file verify user information.")
	log.Debug(req.OperationID, "req.OperationID: ", req.OperationID, " File Type: ", req.FileType)

	log.Debug(req.OperationID, "Tencent cloud upload file, Get cloud client for upload.")
	COSCredential, err := getTencentCloudCredential()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": err.Error()})
		return
	}

	log.Info("upload ", COSCredential.Credentials.SessionToken, "bucket ", COSCredential.Credentials.TmpSecretID)
	log.Debug("", "config.Config.Credential.Tencent.Bucket: ", config.Config.Credential.Tencent.Bucket)
	dir := ""
	if config.Config.Credential.Tencent.Accelerate {
		dir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.Bucket)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.Bucket, config.Config.Credential.Tencent.Region)
	}
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSCredential.Credentials.TmpSecretID,
			SecretKey:    COSCredential.Credentials.TmpSecretKey,
			SessionToken: COSCredential.Credentials.SessionToken,
		},
	})
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": "Create tencent client error"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "FormFile failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file arg: " + err.Error()})
		return
	}

	listener := &selfListener{}
	listener.ConsumedTotalBytes = 0
	//listener.onProgressFun = func(progress int) {
	//}
	listener.MultiFilesTotalBytes = 0
	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
		XOptionHeader: &http.Header{},
		Listener:      listener,
	}}
	opt.XOptionHeader.Add("x-cos-tagging", "delete:1")
	switch req.FileType {
	case constant.VideoType:
		snapShotFile, err := c.FormFile("snapShot")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing snapshot arg: " + err.Error()})
			return
		}
		snapShotFileObj, err := snapShotFile.Open()
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
			return
		}

		snapShotMD5Obj, err := snapShotFile.Open()
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
			return
		}

		snapShotNewName, snapShotNewType := utils.GetNewFileMD5Name(snapShotFile.Filename, &snapShotMD5Obj, constant.ImageType)
		ok, err := client.Object.IsExist(context.Background(), snapShotNewName)
		if err == nil && ok {
			resp.SnapshotNewName = snapShotNewName
			resp.SnapshotURL = fmt.Sprintf("%s/%s", dir, snapShotNewName)
			break
		} else if err != nil {
			fmt.Printf("head object failed: %v\n", err)
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "Tencent connect failed", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
			return
		}

		listener.MultiFilesTotalBytes = file.Size + snapShotFile.Size
		_, err = client.Object.Put(context.Background(), snapShotNewName, snapShotFileObj, opt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "PutFromFile error" + err.Error()})
			return
		}

		resp.SnapshotURL = fmt.Sprintf("%s/%s", dir, snapShotNewName)
		resp.SnapshotNewName = snapShotNewName
		log.Debug(req.OperationID, "Get the snapshot of video file, then upload it, snap shot new name: ", snapShotNewName, "snap shot type:", snapShotNewType)
	}
	fileObj, err := file.Open()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
		return
	}
	fileMD5Obj, err := file.Open()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
		return
	}

	newName, newType := utils.GetNewFileMD5Name(file.Filename, &fileMD5Obj, req.FileType)
	if newType == "img" {
		opt.ContentType = newType
	}

	ok, err := client.Object.IsExist(context.Background(), newName)
	if err == nil && ok {
		resp.NewName = newName
		resp.URL = fmt.Sprintf("%s/%s", dir, newName)
		log.NewInfo(req.OperationID, "upload file to bucket success, snapshot url: ", resp.SnapshotURL, " file url: ", resp.URL)
		c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
		return
	} else if err != nil {
		fmt.Printf("head object failed: %v\n", err)
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Tencent connect failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
		return
	}

	// _, err = client.Object.PutFromFile(context.Background(), newName, file.Filename, opt)
	_, err = client.Object.Put(context.Background(), newName, fileObj, opt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "PutFromFile error " + err.Error()})
		return
	}
	resp.NewName = newName
	resp.URL = fmt.Sprintf("%s/%s", dir, newName)
	log.NewInfo(req.OperationID, "upload file to bucket success, snapshot url: ", resp.SnapshotURL, " file url: ", resp.URL)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
	return
}

func TencentCloudUploadPersistentFile(c *gin.Context) {
	var (
		req  api.TencentCloudUploadFileReq
		resp api.TencentCloudUploadFileResp
	)
	defer func() {
		if r := recover(); r != nil {
			log.NewError(req.OperationID, "panic is", utils.GetSelfFuncName(), r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "Tencent Cloud UploadFile panic", "panic:", string(buf))
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file or snapShot args"})
			return
		}
	}()
	if err := c.Bind(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	log.Debug(req.OperationID, "Tencent cloud upload file verify user information.")

	COSCredential, err := getTencentCloudCredential()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": err.Error()})
		return
	}

	log.Info("upload ", COSCredential.Credentials.SessionToken, "bucket ", COSCredential.Credentials.TmpSecretID)
	dir := ""
	if config.Config.Credential.Tencent.Accelerate {
		dir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket, config.Config.Credential.Tencent.Region)
	}
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSCredential.Credentials.TmpSecretID,
			SecretKey:    COSCredential.Credentials.TmpSecretKey,
			SessionToken: COSCredential.Credentials.SessionToken,
		},
	})
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": "Create tencent client error"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "FormFile failed", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file arg: " + err.Error()})
		return
	}
	listener := &selfListener{}
	listener.MultiFilesTotalBytes = 0
	listener.ConsumedTotalBytes = 0
	//listener.onProgressFun = func(progress int) {
	//	db.DB.SetProgress(token, progress, 60)
	//}
	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
		XOptionHeader: &http.Header{},
		Listener:      listener,
	}}
	opt.XOptionHeader.Add("x-cos-tagging", "delete:1")

	switch req.FileType {
	case constant.VideoType:
		snapShotFile, err := c.FormFile("snapShot")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing snapshot arg: " + err.Error()})
			return
		}
		snapShotFileObj, err := snapShotFile.Open()
		if err != nil {
			log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
			return
		}

		tempSnapObj, err := snapShotFile.Open()
		if err == nil {
			decodeConfig, _, err := image.DecodeConfig(tempSnapObj) // Image Struct
			if err == nil {
				resp.ImageWidth = decodeConfig.Width
				resp.ImageHeight = decodeConfig.Height
			}
		}
		snapShotNewName, _ := utils.GetNewFileNameAndContentType(snapShotFile.Filename, constant.ImageType)
		//snapShotMd5Obj, err := snapShotFile.Open()
		//if err != nil {
		//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
		//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		//	return
		//}
		//
		//snapShotNewName, _ := utils.GetNewFileMD5Name(snapShotFile.Filename, &snapShotMd5Obj, constant.ImageType)
		//
		//ok, err = client.Object.IsExist(context.Background(), snapShotNewName)
		//if err == nil && ok {
		//	resp.SnapshotNewName = snapShotNewName
		//	resp.SnapshotURL = fmt.Sprintf("%s/%s", dir, snapShotNewName)
		//	break
		//} else if err != nil {
		//	fmt.Printf("head object failed: %v\n", err)
		//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "Tencent connect failed", err.Error())
		//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
		//	return
		//}

		listener.MultiFilesTotalBytes = file.Size + snapShotFile.Size
		_, err = client.Object.Put(context.Background(), snapShotNewName, snapShotFileObj, opt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "PutFromFile error" + err.Error()})
			return
		}
		resp.SnapshotURL = fmt.Sprintf("%s/%s", dir, snapShotNewName)
		resp.SnapshotNewName = snapShotNewName
	}
	fileObj, err := file.Open()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
		return
	}
	if req.FileType == constant.ImageType {
		fileObj, err := file.Open()
		if err == nil {
			decodeConfig, _, err := image.DecodeConfig(fileObj) // Image Struct
			if err == nil {
				resp.ImageWidth = decodeConfig.Width
				resp.ImageHeight = decodeConfig.Height
			}
		}

	}

	newName, newType := utils.GetNewFileNameAndContentType(file.Filename, req.FileType)
	// newName, newType := utils.GetNewFileMD5Name(file.Filename, &fileObj, req.FileType)
	if newType == "img" {
		opt.ContentType = newType
	}
	//
	//ok, err = client.Object.IsExist(context.Background(), newName)
	//if err == nil && ok {
	//	resp.NewName = newName
	//	resp.URL = fmt.Sprintf("%s/%s", dir, newName)
	//	log.NewInfo(req.OperationID, "upload file to bucket success, snapshot url: ", resp.SnapshotURL, " file url: ", resp.URL)
	//	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
	//	return
	//} else if err != nil {
	//	fmt.Printf("head object failed: %v\n", err)
	//	log.NewError(req.OperationID, utils.GetSelfFuncName(), "Tencent connect failed", err.Error())
	//	c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
	//	return
	//}

	fileObj, err = file.Open()
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "Open file error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "invalid file path" + err.Error()})
		return
	}
	// _, err = client.Object.PutFromFile(context.Background(), newName, file.Filename, opt)
	_, err = client.Object.Put(context.Background(), newName, fileObj, opt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "PutFromFile error " + err.Error()})
		return
	}

	resp.NewName = newName
	resp.URL = fmt.Sprintf("%s/%s", dir, newName)
	log.NewInfo(req.OperationID, "upload file to bucket success, snapshot url: ", resp.SnapshotURL, " file url: ", resp.URL)
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
	return
}

func TencentCloudMultiUploadFile(c *gin.Context) {
	var (
		req  api.TencentCloudMultiUploadFileReq
		resp api.TencentCloudMultiUploadFileResp
	)
	defer func() {
		if r := recover(); r != nil {
			log.NewError(req.OperationID, "panic is", utils.GetSelfFuncName(), r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "Tencent Cloud Multi Upload Files panic", "panic:", string(buf))
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file or snapShot args"})
			return
		}
	}()
	if err := c.ShouldBind(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	form := c.Request.MultipartForm
	files := form.File["files"]
	snapShots := form.File["snapShots"]
	log.Debug(req.OperationID, "Tencent cloud upload file verify user information.")

	COSCredential, err := getTencentCloudCredential()
	if err != nil {
		log.NewError("", "err: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": err.Error()})
		return
	}
	log.Debug("", "uploaddddd......")

	log.Info("upload ", COSCredential.Credentials.SessionToken, "bucket ", COSCredential.Credentials.TmpSecretID)
	dir := ""
	if config.Config.Credential.Tencent.Accelerate {
		dir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.Bucket)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.Bucket, config.Config.Credential.Tencent.Region)
	}
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSCredential.Credentials.TmpSecretID,
			SecretKey:    COSCredential.Credentials.TmpSecretKey,
			SessionToken: COSCredential.Credentials.SessionToken,
		},
	})
	if client == nil {
		log.NewError("", "Create tencent client error")
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": "Create tencent client error"})
		return
	}
	var newNameDirs []string
	var newNames []string
	var snapShotUrls []string
	var snapShotNames []string

	var fileTotalBytes int64 = 0
	var fileInfoList []uploadFileInfo
	for index, file := range files {
		fileTotalBytes += file.Size
		fileInfo := uploadFileInfo{}
		if req.FileType == constant.VideoType {
			if len(snapShots) < index+1 {
				c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing snapShot args"})
				return
			}
			fileTotalBytes += snapShots[index].Size
			fileInfo.SnapShot = snapShots[index]
		}
		fileInfo.FileType = req.FileType
		fileInfo.NewName = ""
		fileInfo.File = files[index]
		fileInfoList = append(fileInfoList, fileInfo)
	}

	filesCh := make(chan *uploadFileInfo, 50)
	listener := &selfListener{}
	var wg sync.WaitGroup
	threadPool := 10
	var uploadErr error = nil
	listener.MultiFilesTotalBytes = fileTotalBytes
	listener.ConsumedTotalBytes = 0
	//listener.onProgressFun = func(progress int) {
	//	db.DB.SetProgress(token, progress, 60)
	//}
	for thread := 0; thread < threadPool; thread++ {
		wg.Add(1)
		go func() {
			err := uploadFileFromChannel(&wg, client, filesCh, listener)
			if err != nil {
				uploadErr = err
			}
		}()
	}
	for index, _ := range fileInfoList {
		log.Debug("", "put file: ", fileInfoList[index].File.Filename, " size: ", fileInfoList[index].File.Size)
		filesCh <- &fileInfoList[index]
	}
	close(filesCh)
	wg.Wait()
	for _, fileInfo := range fileInfoList {
		newNameDirs = append(newNameDirs, dir+"/"+fileInfo.NewName)
		newNames = append(newNames, fileInfo.NewName)
		if fileInfo.SnapshotNewName != "" {
			snapShotUrls = append(snapShotUrls, dir+"/"+fileInfo.SnapshotNewName)
			snapShotNames = append(snapShotNames, fileInfo.SnapshotNewName)
		}
	}
	if uploadErr != nil {
		log.Debug("", "channel error, err: ", uploadErr.Error())
		var obs []cos.Object

		go func() {
			for _, fileName := range newNames {
				if fileName != "" {
					fileName = strings.Trim(fileName, " ")
					log.Debug("", "delete file name: [", strings.Trim(fileName, " ")+"]")
					obs = append(obs, cos.Object{Key: strings.Trim(fileName, " ")})
				}
			}
			deleteOpt := &cos.ObjectDeleteMultiOptions{
				Objects: obs,
			}
			_, _, _ = client.Object.DeleteMulti(context.Background(), deleteOpt)
		}()
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": uploadErr.Error()})
		return
	}
	resp.NewNames = newNames
	resp.URLs = newNameDirs
	resp.SnapshotNewNames = snapShotNames
	resp.SnapshotURLs = snapShotUrls
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
	return
}

func TencentCloudMultiUploadPersistentFile(c *gin.Context) {
	var (
		req  api.TencentCloudMultiUploadFileReq
		resp api.TencentCloudMultiUploadFileResp
	)
	defer func() {
		if r := recover(); r != nil {
			log.NewError(req.OperationID, "panic is", utils.GetSelfFuncName(), r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "Tencent Cloud Multi Upload Files panic", "panic:", string(buf))
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file or snapShot args"})
			return
		}
	}()
	if err := c.ShouldBind(&req); err != nil {
		log.NewError("0", utils.GetSelfFuncName(), "BindJSON failed ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": err.Error()})
		return
	}
	form := c.Request.MultipartForm
	files := form.File["files"]
	snapShots := form.File["snapShots"]
	log.Debug(req.OperationID, "Tencent cloud upload file verify user information. snapShots: ", snapShots)

	COSCredential, err := getTencentCloudCredential()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": err.Error()})
		return
	}

	log.Info("upload ", COSCredential.Credentials.SessionToken, " bucket ", COSCredential.Credentials.TmpSecretID)
	dir := ""
	if config.Config.Credential.Tencent.Accelerate {
		dir = fmt.Sprintf("https://%s.cos.accelerate.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket)
	} else {
		dir = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", config.Config.Credential.Tencent.PersistenceBucket, config.Config.Credential.Tencent.Region)
	}
	u, _ := url.Parse(dir)
	b := &cos.BaseURL{BucketURL: u}

	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:     COSCredential.Credentials.TmpSecretID,
			SecretKey:    COSCredential.Credentials.TmpSecretKey,
			SessionToken: COSCredential.Credentials.SessionToken,
		},
	})
	if client == nil {
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": "Create tencent client error"})
		return
	}
	var newNameDirs []string
	var newNames []string
	var snapShotUrls []string
	var snapShotNames []string
	var ImagesWidth []int
	var ImagesHeight []int
	// Test
	var fileTotalBytes int64 = 0
	var fileInfoList []uploadFileInfo
	for index, file := range files {
		fileTotalBytes += file.Size
		fileInfo := uploadFileInfo{}
		if req.FileType == constant.VideoType {
			if len(snapShots) < index+1 {
				c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing snapShot args"})
				return
			}
			fileTotalBytes += snapShots[index].Size
			fileInfo.SnapShot = snapShots[index]
		}
		fileInfo.FileType = req.FileType
		fileInfo.NewName = ""
		fileInfo.File = files[index]
		fileInfoList = append(fileInfoList, fileInfo)
	}

	// Test
	filesCh := make(chan *uploadFileInfo, 50)
	listener := &selfListener{}
	var wg sync.WaitGroup
	threadPool := 10
	var uploadErr error = nil
	listener.MultiFilesTotalBytes = fileTotalBytes
	listener.ConsumedTotalBytes = 0
	//listener.onProgressFun = func(progress int) {
	//	db.DB.SetProgress(token, progress, 60)
	//}
	for thread := 0; thread < threadPool; thread++ {
		wg.Add(1)
		go func() {
			err := uploadFileFromChannel(&wg, client, filesCh, listener)
			if err != nil {
				uploadErr = err
			}
		}()
	}
	for index, _ := range fileInfoList {
		filesCh <- &fileInfoList[index]
	}
	close(filesCh)
	wg.Wait()
	for _, fileInfo := range fileInfoList {
		newNameDirs = append(newNameDirs, dir+"/"+fileInfo.NewName)
		newNames = append(newNames, fileInfo.NewName)
		if fileInfo.SnapshotNewName != "" {
			snapShotUrls = append(snapShotUrls, dir+"/"+fileInfo.SnapshotNewName)
			snapShotNames = append(snapShotNames, fileInfo.SnapshotNewName)
		}
		ImagesWidth = append(ImagesWidth, fileInfo.ImageWidth)
		ImagesHeight = append(ImagesHeight, fileInfo.ImageHeight)

	}
	if uploadErr != nil {
		log.Error("", "channel error, err: ", uploadErr.Error())
		var obs []cos.Object

		go func() {
			for _, fileName := range newNames {
				if fileName != "" {
					obs = append(obs, cos.Object{Key: strings.Trim(fileName, " ")})
				}
			}
			deleteOpt := &cos.ObjectDeleteMultiOptions{
				Objects: obs,
			}
			_, _, _ = client.Object.DeleteMulti(context.Background(), deleteOpt)
		}()
		c.JSON(http.StatusBadRequest, gin.H{"errCode": constant.ErrTencentCredential.ErrCode, "errMsg": uploadErr.Error()})
		return
	}
	resp.NewNames = newNames
	resp.URLs = newNameDirs
	resp.SnapshotURLs = snapShotUrls
	resp.SnapshotNewNames = snapShotNames
	resp.ImagesWidth = ImagesWidth
	resp.ImagesHeight = ImagesHeight
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
	return
}

func TencentCloudMultiUploadFileProcess(c *gin.Context) {
	var (
		req  api.TencentCloudMultiUploadFileProgressReq
		resp api.TencentCloudMultiUploadFileProgressResp
	)
	defer func() {
		if r := recover(); r != nil {
			log.NewError(req.OperationID, "panic is", utils.GetSelfFuncName(), r)
			buf := make([]byte, 1<<16)
			runtime.Stack(buf, true)
			log.NewError("", "Tencent Cloud Multi Upload Files panic", "panic:", string(buf))
			c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": "missing file or snapShot args"})
			return
		}
	}()
	token := c.Request.Header.Get("token")
	progress, err := db.DB.GetProgress(token)
	if err != nil {
		resp.Progress = 0
		c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
		return
	}
	resp.Progress = progress
	c.JSON(http.StatusOK, gin.H{"errCode": 0, "errMsg": "", "data": resp})
	return
}

func uploadFileFromChannel(wg *sync.WaitGroup, client *cos.Client, files <-chan *uploadFileInfo, listener *selfListener) error {
	opt := &cos.ObjectPutOptions{ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
		XOptionHeader: &http.Header{},
		Listener:      listener,
	}}
	opt.XOptionHeader.Add("x-cos-tagging", "delete:1")

	defer wg.Done()
	var errRes error = nil
	for file := range files {
		log.Debug("", "get file: ", file.File.Filename, " size: ", file.File.Size)

		switch file.FileType {
		case constant.VideoType:
			if file.SnapShot == nil {
				errRes = errors.New("snap shot can't be empty")
				continue
			}

			snapShotFileObj, err := file.SnapShot.Open()
			if err != nil {
				errRes = err
				continue
			}
			snapShotTempObj, err := file.SnapShot.Open()
			if err == nil {
				decodeConfig, _, err := image.DecodeConfig(snapShotTempObj) // Image Struct
				if err == nil {
					file.ImageWidth = decodeConfig.Width
					file.ImageHeight = decodeConfig.Height
				}
			}

			//snapShotMd5Obj, _ := file.SnapShot.Open()
			//snapShotNewName, newType := utils.GetNewFileMD5Name(file.SnapShot.Filename, &snapShotMd5Obj, constant.ImageType)
			//if newType == "img" {
			//	opt.ContentType = newType
			//}
			//fileObj, err = file.Open()

			snapShotNewName, _ := utils.GetNewFileNameAndContentType(file.SnapShot.Filename, constant.ImageType)
			//ok, err := client.Object.IsExist(context.Background(), snapShotNewName)
			//if err == nil && ok {
			//	file.SnapshotNewName = snapShotNewName
			//	break
			//} else if err != nil {
			//	fmt.Printf("head object failed: %v\n", err)
			//	errRes = err
			//	continue
			//}

			_, err = client.Object.Put(context.Background(), snapShotNewName, snapShotFileObj, opt)
			if err != nil {
				errRes = err
				continue
			}
			file.SnapshotNewName = snapShotNewName
		}

		newName, newType := utils.GetNewFileNameAndContentType(file.File.Filename, file.FileType)

		//fileObjMD5, err := file.File.Open()
		//newName, newType := utils.GetNewFileMD5Name(file.File.Filename, &fileObjMD5, constant.ImageType)
		if newType == "img" {
			opt.ContentType = newType
		} else {
			opt.ContentType = ""
		}
		fileObj, err := file.File.Open()
		if err != nil {
			errRes = err
			continue
		}
		if file.FileType == constant.ImageType {
			fileObj, err := file.File.Open()
			if err == nil {
				decodeConfig, _, err := image.DecodeConfig(fileObj) // Image Struct
				if err == nil {
					file.ImageWidth = decodeConfig.Width
					file.ImageHeight = decodeConfig.Height
				}
			}
		}

		//ok, err := client.Object.IsExist(context.Background(), newName)
		//if err == nil && ok {
		//	file.NewName = newName
		//	continue
		//} else if err != nil {
		//	fmt.Printf("head object failed: %v\n", err)
		//	errRes = err
		//	continue
		//}

		fmt.Println("", "upload file: ", newName)
		_, err = client.Object.Put(context.Background(), newName, fileObj, opt)
		if err != nil {
			// The goroutine is error, stop all upload progressing then delete files that has been uploaded.
			errRes = err
			continue
		}
		file.NewName = newName

	}
	return errRes
}

func (l *selfListener) ProgressChangedCallback(event *cos.ProgressEvent) {
	totalBytes := event.TotalBytes
	consumedBytes := event.ConsumedBytes

	switch event.EventType {
	case cos.ProgressDataEvent:
		l.mutex.Lock()
		l.ConsumedTotalBytes = l.ConsumedTotalBytes + event.RWBytes
		l.mutex.Unlock()
		if l.MultiFilesTotalBytes > 0 {
			totalBytes = l.MultiFilesTotalBytes
			consumedBytes = l.ConsumedTotalBytes
		}
		if consumedBytes == totalBytes {
			if l.onProgressFun != nil {
				l.onProgressFun(int((consumedBytes - 1) * 100 / totalBytes))
			}

		} else {
			if l.onProgressFun != nil {
				l.onProgressFun(int(consumedBytes * 100 / totalBytes))
			}
		}
	case cos.ProgressFailedEvent:
	}
}
