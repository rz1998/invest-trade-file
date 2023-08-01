package fileQmtDbf

import (
	"fmt"
	"github.com/rz1998/invest-trade-file/internal/config"
	file "github.com/rz1998/invest-trade-file/internal/logic"
	"github.com/rz1998/invest-trade-file/internal/logic/fileQmtDbf/godbf"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
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
	date := time.Now()
	return fmt.Sprintf("%s%s.dbf", trans.PathFile, date.Format("20060102"))
}

// ReadFileFunc 读取指定方法对应的文件，并将内容以键值对的形式返回
func (api *ApiFileQmtDbf) ReadFileFunc(nameFunc string) []map[string]string {
	var results []map[string]string

	// 读取配置中对应方法的配置信息
	trans, hasTrans := api.conf.MapTrans[nameFunc]
	if !hasTrans {
		logx.Errorf("ReadFileFunc has no func %s", nameFunc)
		return results
	}
	// 生成文件名
	path := api.GetPath(nameFunc)
	if len(path) == 0 {
		logx.Errorf("ReadFileFunc has no path %s", nameFunc)
		return results
	}
	dbfTable, err := godbf.NewFromFile(path, api.conf.Encoding)
	if err != nil {
		logx.Errorf("ReadFileFunc %v", err)
		return results
	}

	nameFields := dbfTable.FieldNames()

	size := dbfTable.NumberOfRecords()
	results = make([]map[string]string, size)

	for i := 0; i < size; i++ {
		results[i] = make(map[string]string)
		for _, nameField := range nameFields {
			results[i][nameField], _ = dbfTable.FieldValueByName(i, nameField)
		}
		results[i] = file.TransKey(trans.JsonTrans, results[i])
	}
	return results
}

// WriteFileFunc 将指定的内容，写入指定的方法对应的文件
func (api *ApiFileQmtDbf) WriteFileFunc(nameFunc string, mapRecords []map[string]string) {
	if mapRecords == nil || len(mapRecords) == 0 {
		return
	}
	// 读取配置中对应方法的配置信息
	trans, hasTrans := api.conf.MapTrans[nameFunc]
	if !hasTrans {
		logx.Errorf("WriteFileFunc has no func %s", nameFunc)
		return
	}
	table, err := godbf.NewFromFile(trans.PathFile, api.conf.Encoding)
	if err != nil {
		logx.Errorf("WriteFileFunc %v", err)
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
					logx.Errorf("WriteFileFunc set field error %v", err)
				}
			}
		}
	}
	err = godbf.SaveToFile(table, trans.PathFile)
	if err != nil {
		logx.Errorf("WriteFileFunc SaveToFile error %v", err)
		return
	}
}
