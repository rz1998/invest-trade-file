package tradeFile

import (
	"github.com/rz1998/invest-trade-basic/types/tradeBasic"
	"github.com/rz1998/invest-trade-file/internal/config"
)

// IApiTrans 对经过配置文件进行字段替换后的内容进行再调整
type IApiTrans interface {
	// Init 初始化，设定参数
	Init(param config.Config)
	// TransAcFund 转换资金数据
	TransAcFund(mapTransed map[string]string) *tradeBasic.SAcFund
	// TransAcPos 转换持仓数据
	TransAcPos(mapTransed map[string]string) *tradeBasic.SAcPos
	// TransInfoOrder 转换委托数据
	TransInfoOrder(mapTransed map[string]string) *tradeBasic.SOrderInfo
	// TransInfoTrade 转换成交数据
	TransInfoTrade(mapTransed map[string]string) *tradeBasic.STradeInfo
	// TransReqOrder 转换报单请求
	TransReqOrder(reqOrder *tradeBasic.PReqOrder) map[string]string
	// TransReqOrderAction 转换撤单请求
	TransReqOrderAction(reqOrderAction *tradeBasic.PReqOrderAction) map[string]string
}

// IApiFile 交易文件读写接口
type IApiFile interface {
	// Init 初始化，设定参数
	Init(param config.Config)
	// 结束并清理
	Stop()
	// GetPath 生成方法对应文件的文件路径
	GetPath(nameFunc string) string
	// ReadFileFunc 读取指定方法对应的文件，并将内容以键值对的形式返回
	ReadFileFunc(nameFunc string) []map[string]string
	// WriteFileFunc 将指定的内容，写入指定的方法对应的文件
	WriteFileFunc(nameFunc string, mapRecords []map[string]string)
}
