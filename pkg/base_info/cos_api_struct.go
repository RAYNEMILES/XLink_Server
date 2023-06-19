package base_info

import sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"

type TencentCloudStorageCredentialReq struct {
	OperationID string `json:"operationID"`
}

type TencentCloudStorageCredentialRespData struct {
	*sts.CredentialResult
	Region string `json:"region"`
	Bucket string `json:"bucket"`
}

type TencentCloudStorageCredentialResp struct {
	CommResp
	CosData TencentCloudStorageCredentialRespData `json:"-"`

	Data map[string]interface{} `json:"data"`
}

type TencentCloudUploadFileReq struct {
	OperationID string `json:"operationID"`
	FileType    int    `form:"fileType" binding:"required"`
}

type TencentCloudUploadFileResp struct {
	CommResp
	URL             string `json:"URL"`
	NewName         string `json:"newName"`
	SnapshotURL     string `json:"snapshotURL,omitempty"`
	SnapshotNewName string `json:"snapshotName,omitempty"`
	ImageWidth      int    `json:"imageWidth,omitempty"`
	ImageHeight     int    `json:"imageHeight,omitempty"`
}

type TencentCloudMultiUploadFileReq struct {
	OperationID string `json:"operationID"`
	FileType    int    `form:"fileType" binding:"required"`
}

type TencentCloudMultiUploadFileResp struct {
	CommResp
	URLs             []string `json:"urls"`
	NewNames         []string `json:"new_names"`
	SnapshotURLs     []string `json:"snapshot_urls,omitempty"`
	SnapshotNewNames []string `json:"snapshot_new_names,omitempty"`
	ImagesWidth      []int    `json:"imageWidth,omitempty"`
	ImagesHeight     []int    `json:"imageHeight,omitempty"`
}

type TencentCloudMultiUploadFileProgressReq struct {
	OperationID string `form:"operationID"`
}

type TencentCloudMultiUploadFileProgressResp struct {
	CommResp
	Progress int `json:"progress"`
}
