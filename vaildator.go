package goweb

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin/binding"
)

var Rules = map[string]string{
	"custom":                "",                                                    // 自定义规则
	"required":              "^.+$",                                                // 是否必须
	"alpha":                 "^[a-zA-Z]+$",                                         // 字母
	"num":                   "^-?[0-9]+$",                                          // 整数
	"float":                 "^(-?\\d+)(\\.\\d+)?$",                                // 浮点数
	"varName":               "^[a-zA-Z][\\w]*$",                                    // 变量名，字母开头，字母、数字、下划线
	"alphaNum":              "^[a-zA-Z0-9]+$",                                      // 字母、数字
	"alphaNumUnline":        "^[\\w]+$",                                            // 字母、数字、下划线
	"alphaNumUnlineDash":    "^[\\w\\-]+$",                                         // 字母、数字、下划线、破折号（-）
	"chs":                   "^[\u4e00-\u9fa5]+$",                                  // 汉字
	"chsAlpha":              "^[a-zA-Z\u4e00-\u9fa5]+$",                            // 汉字、字母
	"chsAlphaNum":           "^[a-zA-Z0-9\u4e00-\u9fa5]+$",                         // 汉字、字母、数字
	"chsAlphaNumUnline":     "^[\\w\u4e00-\u9fa5]+$",                               // 汉字、字母、数字、下划线
	"chsAlphaNumUnlineDash": "^[\\w\\-\u4e00-\u9fa5]+$",                            // 汉字、字母、数字、下划线、破折号（-）
	"mobile":                "^1\\d{10}$",                                          // 手机号
	"email":                 "^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$", // 邮箱
	"postalCode":            "^[1-9]\\d{5}$",                                       // 中国邮政编码
	"idCard":                "(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)",        // 中国身份证
	"ip":                    "\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}",           // IPv4
	"website":               "https?://([\\S]+)\\.([\\S]+)",                        // 网址
	"between":               "",                                                    // 数字取值范围，示例 between=2,5
	"equal":                 "",                                                    // 判断相等，可以是字符串或数字
	"in":                    "",                                                    // 字符串取值范围，示例 in=是,否
	"len":                   "",                                                    // 文字长度区间,示例 len=2,5
}

// 定义数据示例
//
//	type example struct {
//		State *string `validate:"required:状态必须;in=是,否:状态值错误;"`
//	}

type Request struct {
	Param any
	Error error
}

func NewRequest(param any) *Request {
	return &Request{Param: param}
}

func (r *Request) Bind(req *http.Request) *Request {
	if r.Error != nil {
		return r
	}

	if err := binding.Default(req.Method, req.Header.Get("Content-Type")).Bind(req, r.Param); err != nil {
		r.Error = err
	}
	return r
}

func (r *Request) Validate() *Request {
	if r.Error != nil {
		return r
	}
	if err := Validator(r.Param); err != nil {
		r.Error = err
	}
	return r
}

// 验证数据
func Validator(data any) error {
	// 获取结构体的值
	v := reflect.ValueOf(data)

	// 如果 data 是指针类型，解引用指针
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("validator: expected a non-nil pointer, got nil")
		}
		v = v.Elem()
	}

	// 确保 data 是结构体类型
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validator: expected a struct, got %v", v.Kind())
	}

	t := v.Type()

	// 遍历结构体字段
	for i := range t.NumField() {
		fieldValue := v.Field(i)
		// 如果字段是指针类型，解引用指针
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			fieldValue = fieldValue.Elem()
		}
		value := fmt.Sprintf("%v", fieldValue)

		// 解析字段的 validate 标签
		for _, tags := range strings.Split(t.Field(i).Tag.Get("validate"), ";") {
			tag := strings.Split(tags, ":")
			if len(tag) < 2 {
				continue
			}
			arr := strings.Split(tag[0], "=")
			var err error
			switch arr[0] {
			case "required":
				if fieldValue.IsZero() {
					err = errors.New(tag[1])
				}
			case "len", "between", "equal", "in":
				err = compareSizes(value, arr[1], arr[0], tag[1])
			case "custom":
				err = regularVerification(arr[1], value, tag[1])
			default:
				err = regularVerification(Rules[arr[0]], value, tag[1])
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/**
 * @description: 正则验证
 * @param {string} regular 正则表达式
 * @param {string} data 验证数据
 * @param {map[string]Rule} rule 规则信息
 */
func regularVerification(regular string, data string, msg string) error {
	re, err := regexp.Compile(regular)
	if err != nil {
		return err
	}

	if ok := re.MatchString(data); !ok {
		return errors.New(msg)
	}
	return nil
}

/**
 * @description: 数据比较
 * @param {string} data1
 * @param {string} data2
 * @param {string} compare
 * @param {string} msg
 */
func compareSizes(data1 string, data2 string, compare string, msg string) error {
	ok := false
	switch compare {
	case "==":
		if data1 != data2 {
			ok = true
		}
	case "in":
		arr := strings.Split(data2, ",")
		if len(arr) > 0 {
			in := false
			for _, v := range arr {
				if data1 == v {
					in = true
				}
			}
			if !in {
				ok = true
			}
		} else {
			return errors.New("in规则错误，示例 in=开启,关闭")
		}
	case "between":
		data, _ := strconv.ParseFloat(data1, 64)
		var err error
		ok, err = ifRange(data, data2, "between")
		if err != nil {
			return err
		}
	case "len":
		data := utf8.RuneCountInString(data1)
		var err error
		ok, err = ifRange(float64(data), data2, "len")
		if err != nil {
			return err
		}
	}

	if ok {
		return errors.New(msg)
	} else {
		return nil
	}
}

/**
 * @description: 范围判断
 * @param {float64} data 要判断的数据
 * @param {string} data2 范围元素数据
 * @param {string} rule_name 规则名称
 */
func ifRange(data float64, data2 string, rule_name string) (ok bool, err error) {
	err = errors.New(rule_name + "规则错误，示例 " + rule_name + "=2,5")
	arr := strings.Split(data2, ",")
	if len(arr) == 2 {
		data3, _ := strconv.ParseFloat(arr[0], 64)
		data4, _ := strconv.ParseFloat(arr[1], 64)
		if arr[0] != "" && arr[1] != "" {
			if data < data3 || data > data4 {
				ok = true
			}
		} else if arr[0] != "" && arr[1] == "" {
			if data < data3 {
				ok = true
			}
		} else if arr[0] == "" && arr[1] != "" {
			if data > data4 {
				ok = true
			}
		} else {
			return ok, err
		}
	} else {
		return ok, err
	}
	return ok, nil
}
