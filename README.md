# invest-trade-file
各种文件交易接口的go语言通用实现


## 功能
利用IApiFile接口抽象不同交易接口文件的读写过程
利用IApiTrans接口抽象不同数据结构与标准数据结构的转换过程

## 使用
tradeFileClient里的ApiTraderFile是标准交易接口入口
初始化时，传入对应的参数和转换接口（参考internal/logic/transGWT/demo)


