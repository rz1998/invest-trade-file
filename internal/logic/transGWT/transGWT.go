package transGWT

import (
	"fmt"
	invest "github.com/rz1998/invest-basic"
	"github.com/rz1998/invest-basic/types/investBasic"
	"github.com/rz1998/invest-trade-basic/types/tradeBasic"
	"github.com/rz1998/invest-trade-file/internal/config"
	"github.com/rz1998/invest-trade-file/internal/logic"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
	"strings"
	"time"
)

func getUniqueCode(mapTransed map[string]string) string {
	strExchangeCD := mapTransed["exchangeCD"]
	exchangeCD := ""
	switch strExchangeCD {
	case "SH", "上证所", "上交所":
		exchangeCD = "SSE"
	case "SZ", "深证所", "深交所":
		exchangeCD = "SZSE"
	default:
		logx.Errorf("getUniqueCode unhandled exchangeCD %v", strExchangeCD)
	}
	return fmt.Sprintf("%v.%s", mapTransed["uniqueCode"], exchangeCD)
}

type ApiTransGWT struct {
	conf config.Config
}

func (api ApiTransGWT) Init(conf config.Config) {
	api.conf = conf
}

func (api ApiTransGWT) TransAcFund(mapTransed map[string]string) *tradeBasic.SAcFund {
	return (*(logic.MarshStrMap2Struct(&tradeBasic.SAcFund{}, mapTransed).(*interface{}))).(*tradeBasic.SAcFund)
}

func (api ApiTransGWT) TransAcPos(mapTransed map[string]string) *tradeBasic.SAcPos {
	// 代码
	mapTransed["uniqueCode"] = getUniqueCode(mapTransed)
	// 价格
	open, _ := strconv.ParseFloat(mapTransed["priceOpen"], 64)
	open *= 10000
	mapTransed["priceOpen"] = fmt.Sprintf("%.0f", open)
	mapTransed["priceSettle"] = mapTransed["priceOpen"]
	// 冻结股份
	volTotal, _ := strconv.ParseInt(mapTransed["volTotal"], 10, 64)
	volTd, _ := strconv.ParseInt(mapTransed["volTd"], 10, 64)
	volFrozenYd, _ := strconv.ParseInt(mapTransed["volFrozenYd"], 10, 64)
	volFrozenYd = volTotal - volFrozenYd - volTd
	mapTransed["volFrozenYd"] = fmt.Sprintf("%d", volFrozenYd)
	mapTransed["volYd"] = fmt.Sprintf("%d", volTotal-volTd)
	mapTransed["tradeDir"] = fmt.Sprintf("%d", investBasic.LONG)
	mapTransed["volFrozenTotal"] = fmt.Sprintf("%d", volFrozenYd)
	return (*(logic.MarshStrMap2Struct(&tradeBasic.SAcPos{}, mapTransed).(*interface{}))).(*tradeBasic.SAcPos)
}

func (api ApiTransGWT) TransInfoOrder(mapTransed map[string]string) *tradeBasic.SOrderInfo {
	orderRef := mapTransed["orderRef"]
	idOrderSys := mapTransed["idOrderSys"]
	// 代码
	uniqueCode := getUniqueCode(mapTransed)
	// date
	date, _ := time.Parse("20060102 15:04:05", fmt.Sprintf("%v %v", mapTransed["date"], mapTransed["timestamp"]))
	// dir
	strDir := mapTransed["dir"]
	var dir investBasic.EDirTrade
	var flagOffset tradeBasic.EFlagOffset
	if strings.Contains(strDir, "买入") {
		dir = investBasic.LONG
		flagOffset = tradeBasic.Open
	} else if strings.Contains(strDir, "卖出") {
		dir = investBasic.SHORT
		flagOffset = tradeBasic.Close
	} else {
		logx.Errorf("TransInfoOrder unhandled type %s", strDir)
	}
	// reqOrder
	var reqOrder *tradeBasic.PReqOrder
	price, _ := strconv.ParseFloat(mapTransed["price"], 64)
	price *= 10000
	vol, _ := strconv.ParseInt(mapTransed["vol"], 10, 64)
	reqOrder = &tradeBasic.PReqOrder{
		Timestamp:  date.UnixMilli(),
		OrderRef:   orderRef,
		Dir:        dir,
		FlagOffset: flagOffset,
		UniqueCode: uniqueCode,
		Price:      int64(price),
		Vol:        vol,
	}
	// orderSys
	var orderSys *tradeBasic.SOrderSys
	orderSys = &tradeBasic.SOrderSys{
		Timestamp:    time.Now().UnixMilli(),
		OrderRef:     orderRef,
		IdOrderLocal: mapTransed["idOrderLocal"],
		IdOrderSys:   idOrderSys,
	}
	// tradeStatus
	var statusOrderSubmit tradeBasic.EStatusOrderSubmit
	var statusOrder tradeBasic.EStatusOrder
	statusMsg := mapTransed["statusMsg"]
	var timeCancel int64 = 0
	switch statusMsg {
	case "待报", "未报":
		statusOrderSubmit = tradeBasic.EStatusOrderSubmit_NONE
		statusOrder = tradeBasic.EStatusOrder_NONE
	case "已报":
		statusOrderSubmit = tradeBasic.InsertSubmitted
		statusOrder = tradeBasic.NotTraded
	case "已报待撤":
		statusOrderSubmit = tradeBasic.CancelSubmitted
		statusOrder = tradeBasic.NotTraded
	case "部成待撤":
		statusOrderSubmit = tradeBasic.CancelSubmitted
		statusOrder = tradeBasic.PartialTraded
	case "部撤":
		statusOrderSubmit = tradeBasic.CancelSubmitted
		statusOrder = tradeBasic.PartialCanceled
		timeCancel = time.Now().UnixMilli()
	case "已撤":
		statusOrderSubmit = tradeBasic.CancelSubmitted
		statusOrder = tradeBasic.Canceled
		timeCancel = time.Now().UnixMilli()
	case "部成":
		statusOrderSubmit = tradeBasic.Accepted
		statusOrder = tradeBasic.PartialTraded
	case "已成":
		statusOrderSubmit = tradeBasic.Accepted
		statusOrder = tradeBasic.AllTraded
	case "废单":
		statusOrderSubmit = tradeBasic.InsertRejected
		statusOrder = tradeBasic.Canceled
	case "已确认":
		statusOrderSubmit = tradeBasic.Accepted
		statusOrder = tradeBasic.NotTraded
	case "未知":
		statusOrderSubmit = tradeBasic.EStatusOrderSubmit_NONE
		statusOrder = tradeBasic.Unknown
	default:
		logx.Errorf("TransInfoOrder tradeStatus unhandled type %s", statusMsg)
	}
	volTotal, _ := strconv.ParseInt(mapTransed["vol"], 10, 64)
	volTrade, _ := strconv.ParseInt(mapTransed["volTraded"], 10, 64)
	orderStatus := &tradeBasic.SOrderStatus{
		Timestamp:         time.Now().UnixMilli(),
		TimeCancel:        timeCancel,
		StatusOrderSubmit: statusOrderSubmit,
		StatusOrder:       statusOrder,
		StatusMsg:         statusMsg,
		VolTraded:         volTrade,
		VolTotal:          volTotal,
	}
	return &tradeBasic.SOrderInfo{
		OrderSys:    orderSys,
		ReqOrder:    reqOrder,
		OrderStatus: orderStatus,
	}
}

func (api ApiTransGWT) TransInfoTrade(mapTransed map[string]string) *tradeBasic.STradeInfo {
	// 代码
	uniqueCode := getUniqueCode(mapTransed)
	orderRef := mapTransed["orderRef"]
	idOrderSys := mapTransed["idOrderSys"]
	// date
	date, _ := time.Parse("20060102 15:04:05", fmt.Sprintf("%v %v", mapTransed["date"], mapTransed["timestamp"]))
	// sourcePrice
	strTypeTrade := mapTransed["typeTrade"]
	switch strTypeTrade {
	case "普通", "普通成交":
		mapTransed["typeTrade"] = fmt.Sprintf("%d", tradeBasic.Common)
	default:
		logx.Errorf("TransInfoTrade sourcePrice unhandled type %v", strTypeTrade)
	}
	// dir
	strDir := mapTransed["dir"]
	var dir investBasic.EDirTrade
	var flagOffset tradeBasic.EFlagOffset
	if strings.Contains(strDir, "买入") {
		dir = investBasic.LONG
		flagOffset = tradeBasic.Open
	} else if strings.Contains(strDir, "卖出") {
		dir = investBasic.SHORT
		flagOffset = tradeBasic.Close
	} else {
		logx.Errorf("TransInfoTrade dir unhandled type %s", strDir)
	}
	// price
	price, _ := strconv.ParseFloat(mapTransed["price"], 64)
	price *= 10000
	mapTransed["price"] = fmt.Sprintf("%.0f", price)

	// reqOrder
	var reqOrder *tradeBasic.PReqOrder
	reqOrder = &tradeBasic.PReqOrder{
		OrderRef:   orderRef,
		Dir:        dir,
		FlagOffset: flagOffset,
		UniqueCode: uniqueCode,
	}
	// orderSys
	var orderSys *tradeBasic.SOrderSys
	orderSys = &tradeBasic.SOrderSys{
		OrderRef:     orderRef,
		IdOrderLocal: mapTransed["idOrderLocal"],
		IdOrderSys:   idOrderSys,
	}
	// tradeStatus
	mapTransed["margin"] = mapTransed["val"]
	mapTransed["TradeSource"] = fmt.Sprintf("%d", tradeBasic.QUERY)
	mapTransed["timestamp"] = fmt.Sprintf("%d", date.UnixMilli())
	tradeStatus := (*(logic.MarshStrMap2Struct(&tradeBasic.STradeStatus{}, mapTransed).(*interface{}))).(*tradeBasic.STradeStatus)

	return &tradeBasic.STradeInfo{
		Date:        date.Format("2006-01-02"),
		ReqOrder:    reqOrder,
		OrderSys:    orderSys,
		TradeStatus: tradeStatus,
	}
}

func (api ApiTransGWT) TransReqOrder(reqOrder *tradeBasic.PReqOrder) map[string]string {
	mapTransed := make(map[string]string)
	if reqOrder == nil || len(reqOrder.UniqueCode) == 0 {
		return mapTransed
	}
	// 下单类型 order_type
	switch reqOrder.Dir {
	case investBasic.LONG:
		mapTransed["order_type"] = "23"
	case investBasic.SHORT:
		mapTransed["order_type"] = "24"
	default:
		mapTransed["order_type"] = ""
		logx.Errorf("TransReqOrder unhandled type %d", reqOrder.Dir)
	}
	// 委托价格类型 price_type
	mapTransed["price_type"] = "3"
	// 委托价格 mode_price
	if reqOrder.UniqueCode[0] == '0' || reqOrder.UniqueCode[0] == '3' || reqOrder.UniqueCode[0] == '6' {
		mapTransed["mode_price"] = fmt.Sprintf("%.2f", float64(reqOrder.Price)/10000)
	} else {
		mapTransed["mode_price"] = fmt.Sprintf("%.3f", float64(reqOrder.Price)/10000)
	}
	// 证券代码 stock_code
	code, _ := invest.GetSecInfo(reqOrder.UniqueCode)
	mapTransed["stock_code"] = code
	// 委托数量 volume
	mapTransed["volume"] = fmt.Sprintf("%d", reqOrder.Vol)
	// 下单资金账号 account_id
	mapTransed["account_id"] = ""
	// 账号类别 act_type
	mapTransed["act_type"] = ""
	// 账号类型 brokertype
	mapTransed["brokertype"] = ""
	// 策略备注 strategy
	mapTransed["strategy"] = ""
	// 投资备注 note
	mapTransed["note"] = reqOrder.OrderRef
	// 交易参数 tradeparam
	mapTransed["tradeparam"] = ""
	// 写入时间 inserttime
	mapTransed["inserttime"] = ""
	return mapTransed
}

func (api ApiTransGWT) TransReqOrderAction(reqOrderAction *tradeBasic.PReqOrderAction) map[string]string {
	mapTransed := make(map[string]string)
	if reqOrderAction == nil || reqOrderAction.OrderSys == nil {
		return mapTransed
	}
	// 撤单指令名称 order_type
	//mapTransed["order_type"] = "cancel_order"
	mapTransed["order_type"] = "cancel_order_number"
	// 资金账号
	mapTransed["price_type"] = reqOrderAction.OrderSys.IdOrderLocal
	mapTransed["mode_price"] = reqOrderAction.OrderSys.IdOrderSys
	mapTransed["stock_code"] = ""
	//mapTransed["volume"] = reqOrderAction.OrderSys.IdOrderSys
	//mapTransed["account_id"] = reqOrderAction.OrderSys.IdOrderLocal
	// 账号类别 act_type
	mapTransed["act_type"] = ""
	// 账号类型 brokertype
	mapTransed["brokertype"] = ""
	return mapTransed
}
