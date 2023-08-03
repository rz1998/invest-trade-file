package main

import (
	"flag"
	trade "github.com/rz1998/invest-trade-basic"
	"github.com/rz1998/invest-trade-basic/demoTradeSpi"
	"github.com/rz1998/invest-trade-file/internal/config"
	file2 "github.com/rz1998/invest-trade-file/internal/logic"
	"github.com/rz1998/invest-trade-file/internal/logic/fileQmtDbf"
	"github.com/rz1998/invest-trade-file/internal/logic/transGWT"
	file "github.com/rz1998/invest-trade-file/tradeFileClient"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

func main() {
	// 加载参数
	configFile := flag.String("f", "etc/tradeFile.yaml", "the config file")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	logx.Infof("%+v", c)
	// apiFileQMT
	var apiFile file2.IApiFile
	apiFileQMT := fileQmtDbf.ApiFileQmtDbf{}
	apiFileQMT.Init(c)
	apiFile = &apiFileQMT
	// apiTransGWT
	var apiTrans file2.IApiTrans
	apiTransGWT := transGWT.ApiTransGWT{}
	apiTrans = &apiTransGWT
	// 初始化spi
	var spiTrader trade.ISpiTrader
	spiTraderPrint := demoTradeSpi.SpiTraderPrint{}
	spiTrader = &spiTraderPrint
	// 初始化api
	apiTraderFile := file.ApiTraderFile{}
	apiTraderFile.Init(c, &apiFile, &apiTrans)
	apiTraderFile.SetSpi(&spiTrader)
	//
	apiTransGWT.ApiTrade = &apiTraderFile
	// 查询各项内容
	apiTraderFile.Login(nil)
	apiTraderFile.QryAcFund()
	time.Sleep(1 * time.Second)
}
