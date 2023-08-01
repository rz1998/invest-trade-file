package config

// ConfTransFunc 方法转换配置
type ConfTransFunc struct {
	// 调用方法的名称
	NameAct string
	// 方法对应文件的位置
	PathFile string
	// 文件字段名与标准交易字段名的转换规则
	JsonTrans map[string]string
}

// Config 文件交易接口的基本配置
type Config struct {
	// 文件编码
	Encoding string
	// 字段分隔符
	Sep string
	// 方法：方法转换配置对照
	MapTrans map[string]*ConfTransFunc
}
