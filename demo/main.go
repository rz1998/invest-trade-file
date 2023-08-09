package main

import (
	"flag"
	"time"

	trade "github.com/rz1998/invest-trade-basic"
	"github.com/rz1998/invest-trade-basic/demoTradeSpi"
	"github.com/rz1998/invest-trade-file/apiFile/fileQmtDbf"
	"github.com/rz1998/invest-trade-file/apiTrans/transGWT"
	"github.com/rz1998/invest-trade-file/internal/config"
	"github.com/rz1998/invest-trade-file/tradeFileClient"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

func main() {
	// 加载参数
	configFile := flag.String("f", "etc/tradeFile.yaml", "the config tradeFileClient")
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	logx.Infof("%+v", c)
	// apiFileQMT
	apiFile := tradeFileClient.NewApiFile(fileQmtDbf.ApiFileQmtDbf{})
	// apiTransGWT
	apiTrans := tradeFileClient.NewApiTrans(transGWT.ApiTransGWT{})
	// 初始化spi
	spiTrader := trade.NewSpiTrader(demoTradeSpi.SpiTraderPrint{})
	// 初始化api
	apiTraderFile := tradeFileClient.ApiTraderFile{}
	apiTraderFile.Init(c, &apiFile, &apiTrans)
	apiTraderFile.SetSpi(&spiTrader)
	// 查询各项内容
	apiTraderFile.Login(nil)
	apiTraderFile.QryAcFund()
	time.Sleep(1 * time.Second)
}

