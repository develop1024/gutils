package gutils


import (
	"fmt"
	"github.com/gogf/gf/crypto/gmd5"
	"github.com/gogf/gf/database/gdb"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gvalid"
	"log"
	"sort"
	"strings"
)

// 指定返回的状态
const (
	Success     = 0
	ServerError = 50001
	LogicError  = 50002
)

// ResponseError 返回错误
func ResponseError(r *ghttp.Request, err error) {
	log.Println(err)
	g.Log().Error(err.Error())

	_ = r.Response.WriteJsonExit(g.Map{
		"code": ServerError,
		"msg":  "服务器开小差了, 请联系管理员~",
	})
}

// ResponseErrorAndSkip 返回错误
func ResponseErrorAndSkip(r *ghttp.Request, err error) {
	log.Println(err)

	_ = r.Response.WriteJson(g.Map{
		"code": ServerError,
		"msg":  "服务器开小差了, 请联系管理员~",
	})
}

// ResponseErrorWithMsg 返回错误并指定消息
func ResponseErrorWithMsg(msg string, err error, r *ghttp.Request) {
	log.Println(err)
	g.Log().Error(err.Error())

	_ = r.Response.WriteJsonExit(g.Map{
		"code": ServerError,
		"msg":  msg,
	})
}

// ResponseFail 返回失败
func ResponseFail(r *ghttp.Request) {
	_ = r.Response.WriteJsonExit(g.Map{
		"code": LogicError,
		"msg":  "操作失败",
	})
}

// ResponseFailWithMsg 返回失败并指定消息
func ResponseFailWithMsg(msg string, r *ghttp.Request) {
	_ = r.Response.WriteJsonExit(g.Map{
		"code": LogicError,
		"msg":  msg,
	})
}

// ResponseSuccess 返回成功
func ResponseSuccess(r *ghttp.Request) {
	_ = r.Response.WriteJsonExit(g.Map{
		"code": Success,
		"msg":  "操作成功",
	})
}

// ResponseSuccessWithMsg 返回成功并指定消息
func ResponseSuccessWithMsg(msg string, r *ghttp.Request) {
	_ = r.Response.WriteJsonExit(g.Map{
		"code": Success,
		"msg":  msg,
	})
}

// ResponseSuccessWithData 返回成功并指定数据
func ResponseSuccessWithData(data interface{}, r *ghttp.Request) {
	_ = r.Response.WriteJsonExit(g.Map{
		"code": Success,
		"msg":  "操作成功",
		"data": data,
	})
}

// ResponseSuccessWithCustomData 返回成功并指定数据
func ResponseSuccessWithCustomData(data map[string]interface{}, r *ghttp.Request) {
	data["code"] = Success
	data["msg"] = "操作成功"
	_ = r.Response.WriteJsonExit(data)
}

// ResponseSuccessWithDataMsg 返回成功并指定数据和消息
func ResponseSuccessWithDataMsg(data map[string]interface{}, msg string, r *ghttp.Request) {
	_ = r.Response.WriteJsonExit(g.Map{
		"code": Success,
		"msg":  msg,
		"data": data,
	})
}

// GetErrorExit 返回错误并退出
func GetErrorExit(err error, r *ghttp.Request) {
	if err != nil {
		ResponseError(r, err)
	}
}

// GetErrorAndSkip 返回错误并退出
func GetErrorAndSkip(err error, r *ghttp.Request) {
	if err != nil {
		ResponseErrorAndSkip(r, err)
	}
}

// GetStructAndValid 获取结构体中的数据并验证
func GetStructAndValid(params interface{}, r *ghttp.Request) {
	err := r.GetStruct(params)
	GetErrorExit(err, r)

	if err := gvalid.CheckStruct(r.Context(), params, nil); err != nil {
		ResponseFailWithMsg(err.FirstString(), r)
	}
}

// Md5 进行md5加密
func Md5(data string, r *ghttp.Request) string {
	result, err := gmd5.Encrypt(data)
	GetErrorExit(err, r)
	return result
}

// GenerateParamsSign 参数签名
func GenerateParamsSign(params map[string]interface{}) (string, error) {
	// 参数签名，保证参数不被篡改
	paramSecret := g.Cfg().GetString("web.paramSecret")

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var stringList []string
	for _, k := range keys {
		key, val := k, params[k]
		stringList = append(stringList, fmt.Sprintf("%v=%v", key, val))
	}
	stringA := strings.Join(stringList, "&")

	stringSignTemp := fmt.Sprintf("%s&paramsSecret=%s", stringA, paramSecret)
	stringSign, err := gmd5.Encrypt(stringSignTemp)
	sign := strings.ToUpper(stringSign)

	return sign, err
}

// Paginate 分页组件
func Paginate(total int, r *ghttp.Request, function func(page, pageSize int) (gdb.Result, error)) {
	page := r.GetInt("page", 1)
	pageSize := r.GetInt("pageSize", 100)

	totalPage := total / pageSize
	if total % pageSize !=0 {
		totalPage += 1
	}

	result, err := function(page, pageSize)
	GetErrorExit(err, r)

	hasNextPage := true
	if page >= totalPage {
		hasNextPage = false
	}

	ResponseSuccessWithData(g.Map{
		"total": total,
		"totalPage": totalPage,
		"currentPage": page,
		"data": result,
		"hasNextPage": hasNextPage,
	}, r)
}

// LayPage layui分页
func LayPage(r *ghttp.Request, tableName string, where interface{})  {
	page := r.GetInt("page", 1)
	limit := r.GetInt("limit", 10)

	data, err := g.DB().Model(tableName).Where(where).Offset((page-1)*limit).Limit(limit).FindAll()
	GetErrorExit(err, r)

	count, err := g.DB().Model(tableName).Where(where).Count()
	GetErrorExit(err, r)

	_ = r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"msg": "ok",
		"count": count,
		"data": data,
	})
}

// LayPageCallback layui分页支持回调函数
func LayPageCallback(r *ghttp.Request, tableName string, where interface{}, callback func(data *gdb.Result))  {
	page := r.GetInt("page", 1)
	limit := r.GetInt("limit", 10)

	data, err := g.DB().Model(tableName).Where(where).Offset((page-1)*limit).Limit(limit).FindAll()
	GetErrorExit(err, r)

	// 调用回调函数
	callback(&data)

	count, err := g.DB().Model(tableName).Where(where).Count()
	GetErrorExit(err, r)

	_ = r.Response.WriteJsonExit(g.Map{
		"code": 0,
		"msg": "ok",
		"count": count,
		"data": data,
	})
}

// ErgodicParentChild 遍历父子组合关系
func ErgodicParentChild(tableName string, childName string, parentField string, childParentField string, isParentCondition interface{}) *gdb.Result {
	parentData, _ := g.DB().Model(tableName).Where(isParentCondition).FindAll()

	for _, parent := range parentData {
		childData, _ := g.DB().Model(tableName).Where(g.Map{childParentField: parent[parentField]}).FindAll()
		parent[childName] = g.NewVar(childData)
	}
	return &parentData
}

// ErgodicParentChild1 遍历父子组合关系
func ErgodicParentChild1(tableName string, childName string, parentField string, childParentField string) *gdb.Result {
	return ErgodicParentChild(tableName, childName, parentField, childParentField, childParentField + ` is null or `+ childParentField +` =''`)
}

// ErgodicParentChild2 遍历父子组合关系
func ErgodicParentChild2(tableName string, childName string, parentField string, childParentField string) *gdb.Result {
	return ErgodicParentChild(tableName, childName, parentField, childParentField, parentField + `=0`)
}

// HasExists 判断元素是否存在与某列表
func HasExists(dataList interface{}, data interface{}) bool {
	result := false

	switch dataList.(type) {
	case []string:
		for _, item := range dataList.([]string) {
			if item == data {
				result = true
			}
		}
	case []int:
		for _, item := range dataList.([]int) {
			if item == data {
				result = true
			}
		}
	case []int8:
		for _, item := range dataList.([]int8) {
			if item == data {
				result = true
			}
		}
	case []int16:
		for _, item := range dataList.([]int16) {
			if item == data {
				result = true
			}
		}
	case []int32:
		for _, item := range dataList.([]int32) {
			if item == data {
				result = true
			}
		}
	case []int64:
		for _, item := range dataList.([]int64) {
			if item == data {
				result = true
			}
		}
	case []uint:
		for _, item := range dataList.([]int) {
			if item == data {
				result = true
			}
		}
	case []uint8:
		for _, item := range dataList.([]int8) {
			if item == data {
				result = true
			}
		}
	case []uint16:
		for _, item := range dataList.([]int16) {
			if item == data {
				result = true
			}
		}
	case []uint32:
		for _, item := range dataList.([]int32) {
			if item == data {
				result = true
			}
		}
	case []uint64:
		for _, item := range dataList.([]int64) {
			if item == data {
				result = true
			}
		}

	case []float32:
		for _, item := range dataList.([]float32) {
			if item == data {
				result = true
			}
		}
	case []float64:
		for _, item := range dataList.([]float64) {
			if item == data {
				result = true
			}
		}
	case []gdb.Record:
		for _, item := range dataList.([]gdb.Record) {
			if item.Json() == data.(gdb.Record).Json() {
				result = true
			}
		}
	}

	return result
}

