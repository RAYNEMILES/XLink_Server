package guest

import (
	"Open_IM/pkg/cms_api_struct"
	"Open_IM/pkg/common/constant"
	"Open_IM/pkg/common/db/mysql_model/im_mysql_model"
	openIMHttp "Open_IM/pkg/common/http"
	"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"github.com/gin-gonic/gin"
)

func GetGuestStatus(c *gin.Context) {
	data := map[string]interface{}{}
	data["allow_guest_login"] = im_mysql_model.GetAllowGuestLogin()
	openIMHttp.RespHttp200(c, constant.OK, data)
}

func SwitchGuestStatus(c *gin.Context) {
	var apiRequest = cms_api_struct.SwitchGuestStatusRequest{}

	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError("", utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	im_mysql_model.SetConfigByName(constant.AllowGuestLogin, apiRequest.Status)
	openIMHttp.RespHttp200(c, constant.OK, nil)
}

func GetGuestLimit(c *gin.Context) {
	data := map[string]interface{}{}
	data["allow_register_by_uuid"] = im_mysql_model.GetAllowGuestLimit()
	openIMHttp.RespHttp200(c, constant.OK, data)
}

func SwitchGuestLimit(c *gin.Context) {
	var apiRequest = cms_api_struct.SwitchGuestStatusRequest{}

	if err := c.BindJSON(&apiRequest); err != nil {
		log.NewError("", utils.GetSelfFuncName(), "ShouldBindQuery failed ", err.Error())
		openIMHttp.RespHttp200(c, constant.ErrArgs, nil)
		return
	}

	im_mysql_model.SetConfigByName(constant.AllowRegisterByUuid, apiRequest.Status)
	openIMHttp.RespHttp200(c, constant.OK, nil)
}
