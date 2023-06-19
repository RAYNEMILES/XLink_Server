package utils

import (
	"Open_IM/pkg/base_info"
	"Open_IM/pkg/common/config"
	"Open_IM/pkg/common/constant"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/importcjj/sensitive"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/speps/go-hashids"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("version_regular", base_info.VersionRegular)
		v.RegisterValidation("utc", base_info.Utc)
		v.RegisterValidation("language", base_info.Language)
	}
}

// copy a by b  b->a
func CopyStructFields(a interface{}, b interface{}, fields ...string) (err error) {
	return copier.Copy(a, b)
}

func Wrap(err error, message string) error {
	return errors.Wrap(err, "==> "+printCallerNameAndLine()+message)
}

func WithMessage(err error, message string) error {
	return errors.WithMessage(err, "==> "+printCallerNameAndLine()+message)
}

func printCallerNameAndLine() string {
	pc, _, line, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name() + "()@" + strconv.Itoa(line) + ": "
}

func GetSelfFuncName() string {
	pc, _, _, _ := runtime.Caller(1)
	return cleanUpFuncName(runtime.FuncForPC(pc).Name())
}
func cleanUpFuncName(funcName string) string {
	end := strings.LastIndex(funcName, ".")
	if end == -1 {
		return ""
	}
	return funcName[end+1:]
}

// Get the intersection of two slices
func Intersect(slice1, slice2 []uint32) []uint32 {
	m := make(map[uint32]bool)
	n := make([]uint32, 0)
	for _, v := range slice1 {
		m[v] = true
	}
	for _, v := range slice2 {
		flag, _ := m[v]
		if flag {
			n = append(n, v)
		}
	}
	return n
}

// Get the diff of two slices
func Difference(slice1, slice2 []uint32) []uint32 {
	m := make(map[uint32]bool)
	n := make([]uint32, 0)
	inter := Intersect(slice1, slice2)
	for _, v := range inter {
		m[v] = true
	}
	for _, v := range slice1 {
		if !m[v] {
			n = append(n, v)
		}
	}

	for _, v := range slice2 {
		if !m[v] {
			n = append(n, v)
		}
	}
	return n
}

// Get the intersection of two slices
func IntersectString(slice1, slice2 []string) []string {
	m := make(map[string]bool)
	n := make([]string, 0)
	for _, v := range slice1 {
		m[v] = true
	}
	for _, v := range slice2 {
		flag, _ := m[v]
		if flag {
			n = append(n, v)
		}
	}
	return n
}

// Get the diff of two slices
func DifferenceString(slice1, slice2 []string) []string {
	m := make(map[string]bool)
	n := make([]string, 0)
	inter := IntersectString(slice1, slice2)
	for _, v := range inter {
		m[v] = true
	}
	for _, v := range slice1 {
		if !m[v] {
			n = append(n, v)
		}
	}

	for _, v := range slice2 {
		if !m[v] {
			n = append(n, v)
		}
	}
	return n
}
func OperationIDGenerator() string {
	return strconv.FormatInt(time.Now().UnixNano()+int64(rand.Uint32()), 10)
}

func RemoveRepeatedStringInList(slc []string) []string {
	var result []string
	tempMap := map[string]byte{}
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l {
			result = append(result, e)
		}
	}
	return result
}

func Pb2String(pb proto.Message) (string, error) {
	marshaler := jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  false,
		EmitDefaults: false,
	}
	return marshaler.MarshalToString(pb)
}

func String2Pb(s string, pb proto.Message) error {
	return proto.Unmarshal([]byte(s), pb)
}

func Map2Pb(m map[string]string) (pb proto.Message, err error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(b, pb)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
func Pb2Map(pb proto.Message) (map[string]interface{}, error) {
	_buffer := bytes.Buffer{}
	jsonbMarshaller := &jsonpb.Marshaler{
		OrigName:     true,
		EnumsAsInts:  true,
		EmitDefaults: false,
	}
	_ = jsonbMarshaller.Marshal(&_buffer, pb)
	jsonCnt := _buffer.Bytes()
	var out map[string]interface{}
	err := json.Unmarshal(jsonCnt, &out)
	return out, err
}
func GetSuperGroupTableName(groupID string) string {
	return constant.SuperGroupChatLogsTableNamePre + groupID
}
func GetErrSuperGroupTableName(groupID string) string {
	return constant.SuperGroupErrChatLogsTableNamePre + groupID
}

func ParseResponse(response *http.Response) (map[string]interface{}, error) {
	var result map[string]interface{}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		err = json.Unmarshal(body, &result)
	}
	return result, err
}

func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode(id int64) string {
	hd := hashids.NewData()
	hd.Salt = config.Config.Invite.Salt
	hd.MinLength = 6
	hd.Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{int(id)})
	return e
}

// DecodeInviteCode 解码邀请码
func DecodeInviteCode(code string) (int64, error) {
	hd := hashids.NewData()
	hd.Salt = config.Config.Invite.Salt
	hd.MinLength = 6
	hd.Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	h, _ := hashids.NewWithData(hd)
	d, err := h.DecodeWithError(code)
	return int64(d[0]), err
}

// GenerateChannelCode 生成渠道码
func GenerateChannelCode(id int64) string {
	hd := hashids.NewData()
	hd.Salt = config.Config.Invite.Salt
	hd.MinLength = 7
	hd.Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	h, _ := hashids.NewWithData(hd)
	e, _ := h.Encode([]int{int(id)})
	return e
}

func CompressStr(str string) string {
	if str == "" {
		return ""
	}
	//匹配一个或多个空白符的正则表达式
	reg := regexp.MustCompile("\\s+")
	return reg.ReplaceAllString(str, "")
}

func RemoveDuplicatesAndEmpty(adders []string) []string {
	result := make([]string, 0, len(adders))
	temp := map[string]struct{}{}
	for _, item := range adders {
		if _, ok := temp[item]; !ok {
			if item != "" {
				temp[item] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return result
}

var filter *sensitive.Filter

func FindSensitiveWord(context string) bool {
	if filter == nil {
		filter = sensitive.New()
		filter.LoadWordDict("../dict.txt")
	}

	result, _ := filter.FindIn(context)

	return result
}

func CleanUpfuncName(funcName string) string {
	end := strings.LastIndex(funcName, ".")
	if end == -1 {
		return ""
	}
	return funcName[end+1:]
}
