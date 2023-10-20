package fileQmtDbf

import (
	"fmt"
	"sync"
	"time"

	"github.com/rz1998/invest-trade-file/apiFile/fileQmtDbf/godbf"
	"github.com/rz1998/invest-trade-file/internal/config"
	"github.com/rz1998/invest-trade-file/internal/logic"
	"github.com/zeromicro/go-zero/core/logx"
)

type ApiFileQmtDbf struct {
	conf config.Config
	// 缓存配置map，保证线程安全 funcName string : ConfTransFunc
	mapTransSafe sync.Map
	// 缓存不同文件的写入chan   pathFile string : chan mapRecords []map[string]string
	mapChanWrite sync.Map
}

func (api *ApiFileQmtDbf) Init(conf config.Config) {
	api.conf = conf
	api.mapTransSafe = sync.Map{}
	for k, v := range api.conf.MapTrans {
		api.mapTransSafe.Store(k, v)
	}
	api.mapChanWrite = sync.Map{}
}
func (api *ApiFileQmtDbf) Stop() {
	api.mapChanWrite.Range(func(pathFile, cnVal any) bool {
		cn := cnVal.(chan []map[string]string)
		close(cn)
		api.mapChanWrite.Delete(pathFile)
		return true
	})
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
	if len(mapRecords) == 0 {
		return
	}
	strMethod := "WriteFileFunc"
	// 读取配置中对应方法的配置信息
	transVal, hasTrans := api.mapTransSafe.Load(nameFunc)
	if !hasTrans {
		logx.Errorf("%s has no func %s", strMethod, nameFunc)
		return
	}
	trans := transVal.(*config.ConfTransFunc)
	var cn chan []map[string]string
	cnVal, hasCN := api.mapChanWrite.Load(trans.PathFile)
	if !hasCN {
		// 创建 chan
		cn = make(chan []map[string]string, 10)
		// 放入缓存
		api.mapChanWrite.Store(trans.PathFile, cn)
		// 创建任务
		go func(confTrans *config.ConfTransFunc, chanWrite chan []map[string]string) {
			for mapRec := range chanWrite {
				table, err := godbf.NewFromFile(trans.PathFile, api.conf.Encoding)
				if err != nil {
					logx.Errorf("%s %v", strMethod, err)
					continue
				}
				fields := table.Fields()
				size := len(mapRec)
				for j := 0; j < size; j++ {
					i, _ := table.AddNewRecord()
					for _, field := range fields {
						if record, ok := mapRec[j][field.Name()]; ok {
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
					continue
				}
			}
		}(trans, cn)
	} else {
		cn = cnVal.(chan []map[string]string)
	}
	cn <- mapRecords
}
