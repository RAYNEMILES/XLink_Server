package common

import (
	"Open_IM/cmd/Open-IM-SDK-Core/open_im_sdk_callback"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/constant"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/server_api_params"
	"Open_IM/cmd/Open-IM-SDK-Core/pkg/utils"
	"Open_IM/pkg/common/log"
	"encoding/json"
	"errors"
	"runtime"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func CheckAnyErrCallback(callback open_im_sdk_callback.Base, errCode int32, err error, operationID string) {
	if err != nil {
		errInfo := "operationID[" + operationID + "], " + "info[" + err.Error() + "]"
		log.NewError(operationID, "checkErr ", errInfo)
		callback.OnError(errCode, errInfo)
		runtime.Goexit()
	}
}
func CheckConfigErrCallback(callback open_im_sdk_callback.Base, err error, operationID string) {
	CheckAnyErrCallback(callback, constant.ErrConfig.ErrCode, err, operationID)
}

func CheckTokenErrCallback(callback open_im_sdk_callback.Base, err error, operationID string) {
	CheckAnyErrCallback(callback, constant.ErrTokenInvalid.ErrCode, err, operationID)
}

func CheckDBErrCallback(callback open_im_sdk_callback.Base, err error, operationID string) {
	CheckAnyErrCallback(callback, constant.ErrDB.ErrCode, err, operationID)
}

func CheckDataErrCallback(callback open_im_sdk_callback.Base, err error, operationID string) {
	CheckAnyErrCallback(callback, constant.ErrData.ErrCode, err, operationID)
}

func CheckArgsErrCallback(callback open_im_sdk_callback.Base, err error, operationID string) {
	CheckAnyErrCallback(callback, constant.ErrArgs.ErrCode, err, operationID)
}

func CheckErrAndRespCallback(callback open_im_sdk_callback.Base, err0 error, resp []byte, output interface{}, operationID string) {
	log.Debug(operationID, utils.GetSelfFuncName(), "args: ", string(resp))
	log.Debug(operationID, utils.GetSelfFuncName(), "Sandman: ", err0)
	if err := CheckErrAndResp(err0, resp, output); err != nil {
		var c server_api_params.CommDataResp
		err1 := json.Unmarshal(resp, &c)
		if err1 == nil {
			log.Error(operationID, "Custome error thrown on base of response", "input: ", string(resp))
			callback.OnError(c.ErrCode, c.ErrMsg)
			runtime.Goexit()
		}

		log.Error(operationID, "CheckErrAndResp failed ", err.Error(), "input: ", string(resp))
		callback.OnError(constant.ErrServerReturn.ErrCode, "network error")
		runtime.Goexit()
	}
}

//
//func CheckErrAndResp2(err error, resp []byte, output interface{}) error {
//	if err != nil {
//		return utils.Wrap(err, "api resp failed")
//	}
//	var c server_api_params.CommDataResp
//	err = json.Unmarshal(resp, &c)
//	if err == nil {
//		if c.ErrCode != 0 {
//			return utils.Wrap(errors.New(c.ErrMsg), "")
//		}
//		if output != nil {
//			err = mapstructure.Decode(c.Data, output)
//			if err != nil {
//				goto one
//			}
//			return nil
//		}
//		return nil
//	}
//
//	unMarshaler := jsonpb.Unmarshaler{}
//	unMarshaler.Unmarshal()
//	s, _ := marshaler.MarshalToString(pb)
//	out := make(map[string]interface{})
//	json.Unmarshal([]byte(s), &out)
//	if idFix {
//		if _, ok := out["id"]; ok {
//			out["_id"] = out["id"]
//			delete(out, "id")
//		}
//	}
//	return out
//
//one:
//	var c2 server_api_params.CommDataRespOne
//
//	err = json.Unmarshal(resp, &c2)
//	if err != nil {
//		return utils.Wrap(err, "")
//	}
//	if c2.ErrCode != 0 {
//		return utils.Wrap(errors.New(c2.ErrMsg), "")
//	}
//	if output != nil {
//		err = mapstructure.Decode(c2.Data, output)
//		if err != nil {
//			return utils.Wrap(err, "")
//		}
//		return nil
//	}
//	return nil
//}

func CheckErrAndResp(err error, resp []byte, output interface{}) error {
	if err != nil {
		return utils.Wrap(err, "api resp failed")
	}
	var c server_api_params.CommDataResp
	err = json.Unmarshal(resp, &c)
	if err == nil {
		if c.ErrCode != 0 {
			return utils.Wrap(errors.New(c.ErrMsg), "")
		}
		if output != nil {
			//newData := []map[string]interface{}{}
			//for _, datum := range c.Data {
			//	createTime, _ := datum["createTime"]
			//	if createTime != nil {
			//		ctFloat := createTime.(float64)
			//		datum["createTime"] = int(ctFloat)
			//		//ct, ok := createTime.(string)
			//		//if ok {
			//		//datum["createTime"], _ = strconv.ParseInt(createTime.(string), 10, 64)
			//		//}
			//	}
			//	newData = append(newData, datum)
			//}
			//c.Data = newData

			err = mapstructure.Decode(c.Data, output)
			//log.Error("Misip", "Data: ", fmt.Sprintf("%v", c.Data))
			//log.Error("Misip", "OutPut: ", fmt.Sprintf("%v", output))
			if err != nil {
				//log.Error("mapstructure.Decode failed ", "err: ", err.Error(), c.Data, output)
				goto one
			}
			return nil
		}
		return nil
	} else {
		//	log.Error("json.Unmarshal failed ", string(resp), "err: ", err.Error())
	}

one:
	var c2 server_api_params.CommDataRespOne

	err = json.Unmarshal(resp, &c2)
	if err != nil {
		log.Error("json.Unmarshal failed ", string(resp), "err: ", err.Error())
		return utils.Wrap(err, "")
	}
	if c2.ErrCode != 0 {
		return utils.Wrap(errors.New(c2.ErrMsg), "")
	}
	if output != nil {
		err = mapstructure.Decode(c2.Data, output)
		if err != nil {
			return utils.Wrap(err, "")
		}
		return nil
	}
	return nil
}

func JsonUnmarshalAndArgsValidate(s string, args interface{}, callback open_im_sdk_callback.Base, operationID string) error {
	err := json.Unmarshal([]byte(s), args)
	if err != nil {
		if callback != nil {
			log.NewError(operationID, "Unmarshal failed ", err.Error(), s)
			callback.OnError(constant.ErrArgs.ErrCode, err.Error())
			runtime.Goexit()
		} else {
			return utils.Wrap(err, "json Unmarshal failed")
		}
	}
	//err = validate.Struct(args)
	//if err != nil {
	//	if callback != nil {
	//		log.NewError(operationID, "validate failed ", err.Error(), s)
	//		callback.OnError(constant.ErrArgs.ErrCode, constant.ErrArgs.ErrMsg)
	//		runtime.Goexit()
	//	}
	//}
	//return utils.Wrap(err, "args check failed")
	return nil
}

func JsonUnmarshalCallback(s string, args interface{}, callback open_im_sdk_callback.Base, operationID string) error {
	err := json.Unmarshal([]byte(s), args)
	if err != nil {
		if callback != nil {
			log.NewError(operationID, "Unmarshal failed ", err.Error(), s)
			callback.OnError(constant.ErrArgs.ErrCode, err.Error())
			runtime.Goexit()
		} else {
			return utils.Wrap(err, "json Unmarshal failed")
		}
	}
	return nil
}