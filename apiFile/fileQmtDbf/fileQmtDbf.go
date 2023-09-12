package fileQmtDbf

import (
	"fmt"
	"time"

	"github.com/rz1998/invest-trade-file/apiFile/fileQmtDbf/godbf"
	"github.com/rz1998/invest-trade-file/internal/config"
	"github.com/rz1998/invest-trade-file/internal/logic"
	"github.com/zeromicro/go-zero/core/logx"
)

type ApiFileQmtDbf struct {
	conf config.Config
}

func (api *ApiFileQmtDbf) Init(conf config.Config) {
	api.conf = conf
}

func (api *ApiFileQmtDbf) GetPath(nameFunc string) string {
	// 读取配置中对应方法的配置信息
	trans, hasTrans := api.conf.MapTrans[nameFunc]
	if !hasTrans {
		logx.Errorf("ReadFileFunc has no func %s", nameFunc)
		return ""
	}
	// 根据日期设置文件名
	date := time.Now()
	return fmt.Sprintf("%s%s.dbf", trans.PathFile, date.Format("20060102"))
}

// ReadFileFunc 读取指定方法对应的文件，并将内容以键值对的形式返回
func (api *ApiFileQmtDbf) ReadFileFunc(nameFunc string) []map[string]string {
	var results []map[string]string
	strMethod := "ReadFileFunc"
	// 读取配置中对应方法的配置信息
	trans, hasTrans := api.conf.MapTrans[nameFunc]
	if !hasTrans {
		logx.Errorf("%s has no func %s", strMethod, nameFunc)
		return results
	}
	// 生成文件名
	path := api.GetPath(nameFunc)
	if len(path) == 0 {
		logx.Errorf("%s has no path %s", strMethod, nameFunc)
		return results
	}
	// 读取文件内容
	dbfTable, err := godbf.NewFromFile(path, api.conf.Encoding)
	if err != nil {
		logx.Errorf("%s %v", strMethod, err)
		return results
	}

	nameFields := dbfTable.FieldNames()

	size := dbfTable.NumberOfRecords()

	for i := 0; i < size; i++ {
		result := make(map[string]string)
		allEmpty := true
		for _, nameField := range nameFields {
			result[nameField], _ = dbfTable.FieldValueByName(i, nameField)
			if len(result[nameField]) > 0 {
				allEmpty = false
			}
		}
		if !allEmpty {
			results = append(results, logic.TransKey(trans.JsonTrans, result))
		}
	}
	return results
}

// WriteFileFunc 将指定的内容，写入指定的方法对应的文件
func (api *ApiFileQmtDbf) WriteFileFunc(nameFunc string, mapRecords []map[string]string) {
	if mapRecords == nil || len(mapRecords) == 0 {
		return
	}
	strMethod := "WriteFileFunc"
	// 读取配置中对应方法的配置信息
	trans, hasTrans := api.conf.MapTrans[nameFunc]
	if !hasTrans {
		logx.Errorf("%s has no func %s", strMethod, nameFunc)
		return
	}
	table, err := godbf.NewFromFile(trans.PathFile, api.conf.Encoding)
	if err != nil {
		logx.Errorf("%s %v", strMethod, err)
		return
	}
	fields := table.Fields()
	size := len(mapRecords)
	for j := 0; j < size; j++ {
		i, _ := table.AddNewRecord()
		for _, field := range fields {
			if record, ok := mapRecords[j][field.Name()]; ok {
				err = table.SetFieldValueByName(i, field.Name(), record)
				if err != nil {
					logx.Errorf("%s set field error %v", strMethod, err)
				}
			}
		}
	}
	err = godbf.SaveToFile(table, trans.PathFile)
	if err != nil {
		logx.Errorf("%s SaveToFile error %v", strMethod, err)
		return
	}
}
