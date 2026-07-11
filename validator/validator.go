package validator

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

// Rules 内置验证规则
var Rules = map[string]string{
	"custom":                "",
	"required":              "^.+$",
	"alpha":                 "^[a-zA-Z]+$",
	"num":                   "^-?[0-9]+$",
	"float":                 "^(-?\\d+)(\\.\\d+)?$",
	"varName":               "^[a-zA-Z][\\w]*$",
	"alphaNum":              "^[a-zA-Z0-9]+$",
	"alphaNumUnline":        "^[\\w]+$",
	"alphaNumUnlineDash":    "^[\\w\\-]+$",
	"chs":                   "^[\u4e00-\u9fa5]+$",
	"chsAlpha":              "^[a-zA-Z\u4e00-\u9fa5]+$",
	"chsAlphaNum":           "^[a-zA-Z0-9\u4e00-\u9fa5]+$",
	"chsAlphaNumUnline":     "^[\\w\u4e00-\u9fa5]+$",
	"chsAlphaNumUnlineDash": "^[\\w\\-\u4e00-\u9fa5]+$",
	"mobile":                "^1\\d{10}$",
	"email":                 "^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$",
	"postalCode":            "^[1-9]\\d{5}$",
	"idCard":                "(^\\d{15}$)|(^\\d{18}$)|(^\\d{17}(\\d|X|x)$)",
	"ip":                    "\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}",
	"website":               "https?://([\\S]+)\\.([\\S]+)",
	"between":               "",
	"equal":                 "",
	"in":                    "",
	"len":                   "",
}

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

// Validator 验证数据
func Validator(data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return fmt.Errorf("validator: expected a non-nil pointer, got nil")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validator: expected a struct, got %v", v.Kind())
	}

	t := v.Type()
	for i := range t.NumField() {
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
			fieldValue = fieldValue.Elem()
		}
		value := fmt.Sprintf("%v", fieldValue)

		for tags := range strings.SplitSeq(t.Field(i).Tag.Get("validate"), ";") {
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
				err = checkComparison(value, arr[1], arr[0], tag[1])
			case "custom":
				err = matchRegex(arr[1], value, tag[1])
			default:
				err = matchRegex(Rules[arr[0]], value, tag[1])
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// matchRegex 正则验证
func matchRegex(pattern string, data string, msg string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	if !re.MatchString(data) {
		return errors.New(msg)
	}
	return nil
}

// checkComparison 数据比较验证
func checkComparison(data1 string, data2 string, compare string, msg string) error {
	hasError := false
	switch compare {
	case "equal":
		hasError = data1 != data2
	case "in":
		arr := strings.Split(data2, ",")
		if len(arr) == 0 {
			return errors.New("in规则错误，示例: in=开启,关闭")
		}
		found := false
		for _, v := range arr {
			if data1 == v {
				found = true
				break
			}
		}
		hasError = !found
	case "between":
		data, _ := strconv.ParseFloat(data1, 64)
		var err error
		hasError, err = checkRange(data, data2)
		if err != nil {
			return err
		}
	case "len":
		length := utf8.RuneCountInString(data1)
		var err error
		hasError, err = checkRange(float64(length), data2)
		if err != nil {
			return err
		}
	}

	if hasError {
		return errors.New(msg)
	}
	return nil
}

// checkRange 判断数据是否在指定范围内
func checkRange(data float64, rangeStr string) (bool, error) {
	arr := strings.Split(rangeStr, ",")
	if len(arr) != 2 {
		return false, fmt.Errorf("范围规则错误，示例: %s=2,5", rangeStr)
	}

	min, err := strconv.ParseFloat(arr[0], 64)
	if err != nil {
		return false, fmt.Errorf("范围最小值格式错误: %v", err)
	}
	max, err := strconv.ParseFloat(arr[1], 64)
	if err != nil {
		return false, fmt.Errorf("范围最大值格式错误: %v", err)
	}

	return data < min || data > max, nil
}
