package appversion

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/grpc-etcdv3/getcdv3"
	"Open_IM/pkg/proto/admin_cms"
	"Open_IM/pkg/proto/appversion"
	pbCommon "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// get app version by id
func GetAppVersionByID(c *gin.Context) {
	var (
		req   cms_api_struct.GetAppVersionByIDReq
		resp  cms_api_struct.GetAppVersionByIDResp
		reqPb appversion.GetAppVersionByIDReq
	)

	//check the params from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	//request gRPC client
	reqPb.OperationID = req.OperationID
	reqPb.ID = req.ID
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := admin_cms.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetAppVersionByID(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	resp = cms_api_struct.GetAppVersionByIDResp{
		ID:          respPb.Appversion.ID,
		Version:     respPb.Appversion.Version,
		Status:      int(respPb.Appversion.Status),
		Type:        int(respPb.Appversion.Type),
		Isforce:     int(respPb.Appversion.Isforce),
		Title:       respPb.Appversion.Title,
		DownloadUrl: respPb.Appversion.DownloadUrl,
		Content:     respPb.Appversion.Content,
		CreateTime:  respPb.Appversion.CreateTime,
		CreateUser:  respPb.Appversion.CreateUser,
		UpdateTime:  respPb.Appversion.UpdateTime,
		UpdateUser:  respPb.Appversion.UpdateUser,
	}
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

// get app versions by page
func GetAppVersions(c *gin.Context) {
	var (
		req   cms_api_struct.GetAppVersionsReq
		resp  cms_api_struct.GetAppVersionsResp
		reqPb appversion.GetAppVersionsReq
	)

	//check the params from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	//request gRPC client
	reqPb.OperationID = req.OperationID
	reqPb.Status = req.Status
	reqPb.Client = req.Client
	reqPb.CreateTimeBegin = req.CreateTimeBegin
	reqPb.CreateTimeEnd = req.CreateTimeEnd
	reqPb.Pagination = &pbCommon.RequestPagination{
		PageNumber: int32(req.PageNumber),
		ShowNumber: int32(req.ShowNumber),
	}
	reqPb.OrderBy = req.OrderBy
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := admin_cms.NewAdminCMSClient(etcdConn)
	respPb, err := client.GetAppVersions(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	//log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "GetAppVersions respPb:", respPb.String())
	//utils.CopyStructFields(resp.Appversions, respPb.Appversions)
	resp.Appversions = []*cms_api_struct.GetAppVersionByIDResp{}
	for _, version := range respPb.Appversions {
		resp.Appversions = append(resp.Appversions, &cms_api_struct.GetAppVersionByIDResp{
			ID:          version.ID,
			Version:     version.Version,
			Type:        int(version.Type),
			Status:      int(version.Status),
			Isforce:     int(version.Isforce),
			Title:       version.Title,
			DownloadUrl: version.DownloadUrl,
			Content:     version.Content,
			CreateTime:  version.CreateTime,
			CreateUser:  version.CreateUser,
			UpdateTime:  version.UpdateTime,
			UpdateUser:  version.UpdateUser,
		})
	}
	resp.ShowNumber = int(respPb.Pagination.ShowNumber)
	resp.CurrentPage = int(respPb.Pagination.CurrentPage)
	resp.Total = int64(respPb.Total)
	openIMHttp.RespHttp200(c, constant.OK, resp)
}

// add app version
func AddAppVersion(c *gin.Context) {
	var (
		req   cms_api_struct.AddAppVersionReq
		reqPb appversion.AddAppVersionReq
	)

	//check the params from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)
	}

	//request gRPC client
	reqPb.OperationID = req.OperationID
	reqPb.Version = req.Version
	reqPb.Client = req.Client
	reqPb.Title = req.Title
	reqPb.DownloadUrl = req.DownloadUrl
	reqPb.Remark = req.Remark
	reqPb.UserID = userID

	if req.Client == "ios" {
		reqPb.Type = 1
	} else {
		reqPb.Type = 2
	}

	if req.Status == "on" {
		reqPb.Status = 2
	} else {
		reqPb.Status = 1
	}

	if req.Isforce == "yes" {
		reqPb.Isforce = 2
	} else {
		reqPb.Isforce = 1
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := admin_cms.NewAdminCMSClient(etcdConn)
	_, err := client.AddAppVersion(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

// edit app version
func EditAppVersion(c *gin.Context) {
	var (
		req   cms_api_struct.EditAppVersionReq
		reqPb appversion.EditAppVersionReq
	)

	//check the params from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)

	}

	//request gRPC client
	reqPb.OperationID = req.OperationID
	reqPb.ID = req.ID
	reqPb.Version = req.Version
	reqPb.Client = req.Client
	reqPb.Title = req.Title
	reqPb.DownloadUrl = req.DownloadUrl
	reqPb.Remark = req.Remark
	reqPb.UserID = userID

	if req.Client == "ios" {
		reqPb.Type = 1
	} else {
		reqPb.Type = 2
	}

	if req.Status == "on" {
		reqPb.Status = 2
	} else {
		reqPb.Status = 1
	}

	if req.Isforce == "yes" {
		reqPb.Isforce = 2
	} else {
		reqPb.Isforce = 1
	}
	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := admin_cms.NewAdminCMSClient(etcdConn)
	_, err := client.EditAppVersion(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

// delete app version
func DeleteAppVersion(c *gin.Context) {
	var (
		req   cms_api_struct.DeleteAppVersionReq
		reqPb appversion.DeleteAppVersionReq
	)

	//check the params from request
	if err := c.BindJSON(&req); err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	// get the userId from middleware
	userID := ""
	userIDInter, exsited := c.Get("userID")
	if exsited {
		userID = userIDInter.(string)

	}

	//request gRPC client
	reqPb.OperationID = req.OperationID
	reqPb.ID = req.ID
	reqPb.UserID = userID

	etcdConn := getcdv3.GetConn(config.Config.Etcd.EtcdSchema, strings.Join(config.Config.Etcd.EtcdAddr, ","), config.Config.RpcRegisterName.OpenImAdminCMSName, reqPb.OperationID)
	if etcdConn == nil {
		errMsg := reqPb.OperationID + "getcdv3.GetConn == nil"
		log.NewError(req.OperationID, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{"errCode": 400, "errMsg": errMsg})
		return
	}

	client := admin_cms.NewAdminCMSClient(etcdConn)
	_, err := client.DeleteAppVersion(c, &reqPb)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "rpc failed:", err.Error())
		openIMHttp.RespHttp200(c, err, nil)
		return
	}
	openIMHttp.RespHttp200(c, constant.OK, nil)
}
