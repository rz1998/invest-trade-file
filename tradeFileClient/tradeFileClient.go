package tradeFileClient

import (
	"fmt"
	trade "github.com/rz1998/invest-trade-basic"
	"github.com/rz1998/invest-trade-basic/types/tradeBasic"
	"github.com/rz1998/invest-trade-file"
	"github.com/rz1998/invest-trade-file/internal/config"
	"github.com/rz1998/invest-trade-file/internal/logic/fileQmtDbf"
	"github.com/rz1998/invest-trade-file/internal/logic/transGWT"
	"sync/atomic"
	"time"
)

type (
	ConfTransFunc = config.ConfTransFunc
	Config        = config.Config
	ApiFileQmtDbf = fileQmtDbf.ApiFileQmtDbf
	ApiTransGWT   = transGWT.ApiTransGWT
)

// ApiTraderFile 对通用交易接口IApiTrader的实现
type ApiTraderFile struct {
	// 配置信息
	param config.Config
	// 文件读写接口
	apiFile *file.IApiFile
	// 标准转换接口
	apiTrans *file.IApiTrans
	// 交易监听接口
	spi              *trade.ISpiTrader
	countOrder       atomic.Int32
	countOrderAction atomic.Int32
	// ucOrder : reqOrder
	MapReqOrderLocal map[string]*tradeBasic.PReqOrder
	// ucOrder : orderSys
	MapOrderSysLocal map[string]*tradeBasic.SOrderSys
	// ucOrder : reqOrder
	MapReqOrderSys map[string]*tradeBasic.PReqOrder
	// ucOrder : orderSys
	MapOrderSysSys map[string]*tradeBasic.SOrderSys
	// ticker
	tickerOrder    *time.Ticker                        // 定期读取委托
	mapOrderLatest map[string]*tradeBasic.SOrderStatus // 唯一性的保存订单最终状态 ucOrder : SOrderStatus
}

func (api *ApiTraderFile) Init(param config.Config, apiFile *file.IApiFile, apiTrans *file.IApiTrans) {
	api.param = param
	api.apiFile = apiFile
	api.apiTrans = apiTrans
}

func (api *ApiTraderFile) Login(infoAc *tradeBasic.PInfoAc) {
	// 初始化map
	api.MapReqOrderLocal = make(map[string]*tradeBasic.PReqOrder)
	api.MapOrderSysLocal = make(map[string]*tradeBasic.SOrderSys)
	api.MapReqOrderSys = make(map[string]*tradeBasic.PReqOrder)
	api.MapOrderSysSys = make(map[string]*tradeBasic.SOrderSys)
	api.mapOrderLatest = make(map[string]*tradeBasic.SOrderStatus)
	// 加载已保存数据
	api.QryAcFund()
	api.QryAcPos()
	api.QryOrder(nil)
	api.QryTrade()
	// 初始化orderRef
	api.countOrder = atomic.Int32{}
	// 初始化orderActionRef
	api.countOrderAction = atomic.Int32{}
	// 启动ticker
	time.Sleep(10 * time.Second)
	api.tickerOrder = api.tickerQryOrder(2 * time.Millisecond)
}

// Logout 登出
func (api *ApiTraderFile) Logout() {
	if api.tickerOrder != nil {
		api.tickerOrder.Stop()
		api.tickerOrder = nil
	}
	// 清理缓存数据
	if api.MapReqOrderLocal != nil {
		for k := range api.MapReqOrderLocal {
			delete(api.MapReqOrderLocal, k)
		}
	}
	if api.MapOrderSysLocal != nil {
		for k := range api.MapOrderSysLocal {
			delete(api.MapOrderSysLocal, k)
		}
	}
	if api.MapReqOrderSys != nil {
		for k := range api.MapReqOrderSys {
			delete(api.MapReqOrderSys, k)
		}
	}
	if api.MapOrderSysSys != nil {
		for k := range api.MapOrderSysSys {
			delete(api.MapOrderSysSys, k)
		}
	}
}

func (api *ApiTraderFile) generateOrderRef() string {
	return fmt.Sprintf("%d_%d", api.countOrder.Add(1), time.Now().UnixMilli())
}

// ReqOrder 报单请求
func (api *ApiTraderFile) ReqOrder(reqOrder *tradeBasic.PReqOrder) {
	if reqOrder != nil {
		nameFunc := "reqOrder"
		trans := *api.apiTrans
		reqOrder.OrderRef = api.generateOrderRef()
		mapRecord := trans.TransReqOrder(reqOrder)
		(*api.apiFile).WriteFileFunc(nameFunc, []map[string]string{mapRecord})
	}
}

// ReqOrderBatch 批量报单请求
func (api *ApiTraderFile) ReqOrderBatch(reqOrders []*tradeBasic.PReqOrder) {
	if reqOrders != nil && len(reqOrders) > 0 {
		nameFunc := "reqOrderBatch"
		trans := *api.apiTrans
		mapRecords := make([]map[string]string, len(reqOrders))
		for i, reqOrder := range reqOrders {
			reqOrder.OrderRef = api.generateOrderRef()
			mapRecords[i] = trans.TransReqOrder(reqOrder)
		}
		(*api.apiFile).WriteFileFunc(nameFunc, mapRecords)
	}
}

// ReqOrderAction 订单操作请求
func (api *ApiTraderFile) ReqOrderAction(reqOrderAction *tradeBasic.PReqOrderAction) {
	if reqOrderAction != nil {
		nameFunc := "reqOrderAction"
		trans := *api.apiTrans
		reqOrderAction.OrderActionRef = fmt.Sprintf("%d", api.countOrderAction.Add(1))
		mapRecord := trans.TransReqOrderAction(reqOrderAction)
		(*api.apiFile).WriteFileFunc(nameFunc, []map[string]string{mapRecord})
	}
}

// QryAcFund 查询资金
func (api *ApiTraderFile) QryAcFund() {
	nameFunc := "qryAcFund"
	mapList := (*api.apiFile).ReadFileFunc(nameFunc)
	spi := *api.spi
	if mapList != nil && len(mapList) > 0 {
		trans := *api.apiTrans
		spi.OnRtnAcFund(trans.TransAcFund(mapList[0]))
	}
}

// QryAcLiability 查询负债
func (api *ApiTraderFile) QryAcLiability() {

}

// QryAcPos 查询持仓
func (api *ApiTraderFile) QryAcPos() {
	nameFunc := "qryAcPos"
	mapList := (*api.apiFile).ReadFileFunc(nameFunc)
	spi := *api.spi
	if mapList != nil && len(mapList) > 0 {
		trans := *api.apiTrans
		for _, mapTransed := range mapList {
			spi.OnRtnAcPos(trans.TransAcPos(mapTransed), false)
		}
	}
	spi.OnRtnAcPos(nil, true)
}

func (api *ApiTraderFile) tickerQryOrder(d time.Duration) *time.Ticker {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			spi := *api.spi
			trans := *api.apiTrans
			// 检查order信息的长度
			nameFunc := "qryOrder"
			mapListNew := (*api.apiFile).ReadFileFunc(nameFunc)
			if len(mapListNew) > 0 {
				// 逐个报单对比数量和状态的变动情况
				for _, mapTransed := range mapListNew {
					orderInfo := trans.TransInfoOrder(mapTransed)
					// 跟现有的比对
					if orderInfo.OrderStatus != nil && orderInfo.OrderSys != nil &&
						len(orderInfo.OrderSys.IdOrderLocal) != 0 {
						orderStatusOld, has := api.mapOrderLatest[orderInfo.OrderSys.IdOrderLocal]
						if !has || orderStatusOld.StatusOrder < orderInfo.OrderStatus.StatusOrder ||
							orderStatusOld.VolTraded < orderInfo.OrderStatus.VolTraded {
							// 没有现有数据，或者现有数据状态更低，保存最新数据
							api.mapOrderLatest[orderInfo.OrderSys.IdOrderLocal] = orderInfo.OrderStatus
							if orderInfo.OrderSys != nil {
								orderInfo.OrderSys.SourceInfo = tradeBasic.RETURN
							}
							// 回调
							spi.OnRtnOrder(orderInfo, false)
						}
					}
				}
			}
		}
	}()
	return ticker
}

// QryOrder 查询委托
func (api *ApiTraderFile) QryOrder(orderSys *tradeBasic.SOrderSys) {
	nameFunc := "qryOrder"
	mapList := (*api.apiFile).ReadFileFunc(nameFunc)
	spi := *api.spi
	if mapList != nil && len(mapList) > 0 {
		trans := *api.apiTrans
		for _, mapTransed := range mapList {
			orderInfo := trans.TransInfoOrder(mapTransed)
			// 设置为查询
			if orderInfo.OrderSys != nil {
				orderInfo.OrderSys.SourceInfo = tradeBasic.QUERY
			}
			// 跟现有的比对
			if orderInfo.OrderStatus != nil && orderInfo.OrderSys != nil {
				orderStatusOld, has := api.mapOrderLatest[orderInfo.OrderSys.IdOrderLocal]
				if !has || orderStatusOld.StatusOrder < orderInfo.OrderStatus.StatusOrder ||
					orderStatusOld.VolTraded < orderInfo.OrderStatus.VolTraded {
					// 没有现有数据，或者现有数据状态更低，保存最新数据
					api.mapOrderLatest[orderInfo.OrderSys.IdOrderLocal] = orderInfo.OrderStatus
				}
			}
			// 回调
			spi.OnRtnOrder(orderInfo, false)
		}
	}
	spi.OnRtnOrder(nil, true)
}

// QryTrade 查询成交
func (api *ApiTraderFile) QryTrade() {
	nameFunc := "qryTrade"
	mapList := (*api.apiFile).ReadFileFunc(nameFunc)
	spi := *api.spi
	if mapList != nil && len(mapList) > 0 {
		trans := *api.apiTrans
		for _, mapTransed := range mapList {
			spi.OnRtnTrade(trans.TransInfoTrade(mapTransed), false)
		}
	}
	spi.OnRtnTrade(nil, true)
}

// SetSpi 设置回报监听
func (api *ApiTraderFile) SetSpi(spi *trade.ISpiTrader) {
	api.spi = spi
}

// GetSpi 获取回报监听
func (api *ApiTraderFile) GetSpi() *trade.ISpiTrader {
	return api.spi
}

// GetInfoSession 获取会话信息
func (api *ApiTraderFile) GetInfoSession() *tradeBasic.SInfoSessionTrader {
	return nil
}

func (api *ApiTraderFile) GenerateUniqueOrder(orderSys *tradeBasic.SOrderSys) string {
	if orderSys == nil {
		return ""
	}
	return orderSys.IdOrderLocal
}
