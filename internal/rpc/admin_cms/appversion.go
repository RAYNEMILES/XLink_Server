package admin_cms

import (
	"Open_IM/pkg/common/constant"
	imdb "Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	pbAdmin "Open_IM/pkg/proto/appversion"
	sdkws "Open_IM/pkg/proto/sdk_ws"
	"Open_IM/pkg/utils"
	"context"
	"strconv"
)

func (*adminCMSServer) GetAppVersionByID(_ context.Context, req *pbAdmin.GetAppVersionByIDReq) (*pbAdmin.GetAppVersionByIDResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdmin.GetAppVersionByIDResp{}
	appversion, err := imdb.GetAppVersionByID(req.ID)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetAppVersionByID failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.Appversion = &pbAdmin.AppVersion{
		ID:          strconv.FormatInt(int64(appversion.ID), 10),
		Version:     appversion.Version,
		Type:        int32(appversion.Type),
		Status:      int32(appversion.Status),
		Isforce:     int32(appversion.Isforce),
		Title:       appversion.Title,
		DownloadUrl: appversion.DownloadUrl,
		Content:     appversion.Content,
		CreateTime:  appversion.CreateTime,
		CreateUser:  appversion.CreateUser,
		UpdateTime:  appversion.UpdateTime,
		UpdateUser:  appversion.UpdateUser,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (*adminCMSServer) GetLatestAppVersion(_ context.Context, req *pbAdmin.GetLatestAppVersionReq) (*pbAdmin.GetLatestAppVersionResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdmin.GetLatestAppVersionResp{}
	var typeA int
	if req.Client == "ios" {
		typeA = 1
	} else if req.Client == "android" {
		typeA = 2
	}
	appversion, err := imdb.GetLatestAppVersion(typeA)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetLatestAppVersion failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.Appversion = &pbAdmin.AppVersion{
		ID:          strconv.FormatInt(int64(appversion.ID), 10),
		Version:     appversion.Version,
		Type:        int32(appversion.Type),
		Status:      int32(appversion.Status),
		Isforce:     int32(appversion.Isforce),
		Title:       appversion.Title,
		DownloadUrl: appversion.DownloadUrl,
		Content:     appversion.Content,
	}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil
}

func (*adminCMSServer) GetAppVersions(_ context.Context, req *pbAdmin.GetAppVersionsReq) (*pbAdmin.GetAppVersionsResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdmin.GetAppVersionsResp{Appversions: []*pbAdmin.AppVersion{}}
	var (
		client          string
		status          string
		createTimeBegin string
		createTimeEnd   string
	)

	if req.Client == "ios" {
		client = "1"
	} else if req.Client == "android" {
		client = "2"
	} else {
		client = ""
	}

	if req.Status == "off" {
		status = "1"
	} else if req.Status == "on" {
		status = "2"
	} else {
		status = ""
	}

	if req.CreateTimeBegin != "" {
		createTimeBegin = utils.GetTimeStampByFormat(req.CreateTimeBegin)
	}

	if req.CreateTimeEnd != "" {
		createTimeEnd = utils.GetTimeStampByFormat(req.CreateTimeEnd)
	}
	appVersions, err := imdb.GetAppVersionsByPage(client, status, createTimeBegin, createTimeEnd, req.Pagination.ShowNumber, req.Pagination.PageNumber, req.OrderBy)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetAppVersions failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}
	for _, v := range appVersions {
		version := &pbAdmin.AppVersion{
			ID:          strconv.FormatInt(int64(v.ID), 10),
			Version:     v.Version,
			Type:        int32(v.Type),
			Status:      int32(v.Status),
			Isforce:     int32(v.Isforce),
			Title:       v.Title,
			DownloadUrl: v.DownloadUrl,
			Content:     v.Content,
			CreateTime:  v.CreateTime,
			CreateUser:  v.CreateUser,
			UpdateTime:  v.UpdateTime,
			UpdateUser:  v.UpdateUser,
		}
		resp.Appversions = append(resp.Appversions, version)
	}

	total, err := imdb.GetAppVersionsCount(client, status, createTimeBegin, createTimeEnd)
	resp.Total = int32(total)
	if err != nil {
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "GetAppVersionsCount failed:", err.Error())
		return resp, openIMHttp.WrapError(constant.ErrDB)
	}

	resp.Pagination = &sdkws.ResponsePagination{ShowNumber: req.Pagination.ShowNumber, CurrentPage: req.Pagination.PageNumber}
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "resp:", resp.String())
	return resp, nil

}

func (*adminCMSServer) AddAppVersion(_ context.Context, req *pbAdmin.AddAppVersionReq) (*pbAdmin.CommonResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdmin.CommonResp{}
	err := imdb.AddAppVersion(req.Title, req.Remark, req.DownloadUrl, req.Version, req.UserID, int(req.Isforce), int(req.Type), int(req.Status))
	if err != nil {
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "AddAppVersion failed:", err.Error())
		return resp, err
	}

	resp.ErrCode = constant.OK.ErrCode
	resp.ErrMsg = constant.OK.ErrMsg
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "AddAppVersion success")
	return resp, nil
}

func (*adminCMSServer) EditAppVersion(_ context.Context, req *pbAdmin.EditAppVersionReq) (*pbAdmin.CommonResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdmin.CommonResp{}
	err := imdb.EditAppVersion(req.Title, req.Remark, req.DownloadUrl, req.Version, req.UserID, req.ID, int(req.Isforce), int(req.Type), int(req.Status))
	if err != nil {
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "EditAppVersion failed:", err.Error())
		return resp, err
	}
	resp.ErrCode = constant.OK.ErrCode
	resp.ErrMsg = constant.OK.ErrMsg
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "EditAppVersion success")
	return resp, nil
}

func (s adminCMSServer) DeleteAppVersion(_ context.Context, req *pbAdmin.DeleteAppVersionReq) (*pbAdmin.CommonResp, error) {
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "req:", req.String())
	resp := &pbAdmin.CommonResp{}
	err := imdb.DeleteAppVersion(req.ID, req.UserID)
	if err != nil {
		resp.ErrCode = constant.ErrDB.ErrCode
		resp.ErrMsg = constant.ErrDB.ErrMsg
		log.NewError(req.OperationID, utils.GetSelfFuncName(), "DeleteAppVersion failed:", err.Error())
		return resp, err
	}
	resp.ErrCode = constant.OK.ErrCode
	resp.ErrMsg = constant.OK.ErrMsg
	log.NewInfo(req.OperationID, utils.GetSelfFuncName(), "DeleteAppVersion success")
	return resp, nil
}
