package logic

import (
	"github.com/zeromicro/go-zero/core/logx"
	"math"
	"reflect"
	"strconv"
	"unicode"
)

// LowerCase 字符首字母小写
func LowerCase(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if unicode.IsUpper(vv[0]) {
				vv[i] += 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

// TransKey 文件中读取的数据， 将key转换为结构体的key
func TransKey(mapTrans map[string]string, mapFile map[string]string) map[string]string {
	mapStruct := make(map[string]string)
	for k, v := range mapTrans {
		mapStruct[v] = mapFile[k]
	}
	return mapStruct
}

// MarshStrMap2Struct 纯字符串map转换为指定数据结构
func MarshStrMap2Struct(s interface{}, strMap map[string]string) interface{} {
	if s == nil {
		return nil
	}
	typeData := reflect.TypeOf(s)
	if typeData.Kind() != reflect.Struct && typeData.Kind() != reflect.Ptr {
		logx.Errorf("MarshValMap2Struct need struct s not %d", typeData.Kind())
		return nil
	}
	var valData reflect.Value
	if typeData.Kind() == reflect.Ptr {
		valData = reflect.ValueOf(s).Elem()
		typeData = reflect.TypeOf(s).Elem()
	} else {
		valData = reflect.ValueOf(s)
	}
	for i := 0; i < valData.NumField(); i++ {
		field := typeData.Field(i)
		value := valData.Field(i)
		key := LowerCase(field.Name)
		strVal := strMap[key]
		switch value.Kind() {
		case reflect.Bool:
			val, _ := strconv.ParseBool(strVal)
			value.SetBool(val)
		case reflect.String:
			value.SetString(strVal)
		case reflect.Float32, reflect.Float64:
			val, _ := strconv.ParseFloat(strVal, 64)
			value.SetFloat(val)
		case reflect.Uint, reflect.Uintptr, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, _ := strconv.ParseFloat(strVal, 64)
			value.SetUint(uint64(math.Round(val)))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, _ := strconv.ParseFloat(strVal, 64)
			value.SetInt(int64(math.Round(val)))
		default:
			logx.Errorf("unhandled type %v key %s", value.Kind(), key)
		}
	}
	if typeData.Kind() == reflect.Ptr {
		return s
	} else {
		return &s
	}
}
