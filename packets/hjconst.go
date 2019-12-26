package packets

// CN (Host command and slave req command)
const (
	// init cmd
	HsetTimeoutRepeats = 1000
	HgetSlaveTime      = 1011
	SupDate            = 1011
	HsetSlaveTime      = 1012
	SsetTimeReq        = 1013
	HgetRealInterval   = 1061
	SupRealInterval    = 1061
	HsetRealInterval   = 1062
	HgetMinuteInterval = 1063
	SupMinuteInterval  = 1063
	HsetMinuteInterval = 1064
	HsetSlaveRestart   = 1070
	HresetSlavePasswd  = 1072
	// data cmd
	// real time data
	HgetRealTimeData  = 2011
	SupRealTimeData   = 2011
	HStopRealTimeData = 2012
	// device status
	HgetDevStat  = 2021
	SupDevSrat   = 2021
	HStopDevStat = 2022
	// day data
	HgetDayHistory           = 2031
	SupDayHistory            = 2031
	HgetDevRunTimeDayHistory = 2041
	SupDevRunTimeDayHistory  = 2041
	// minute data
	HgetMinHistory = 2051
	SupMinHistory  = 2051
	// hour data
	HgetHourHistory = 2061
	SupHourHistory  = 2061
	// orther data
	SupSCYupTime = 2081
	// control cmd
	HzeroCal      = 3011
	HrealSampling = 3012
	HcleanStart   = 3013
	HcompSampling = 3014
	HkeepSample   = 3015
	//.....
	HgetDevID        = 3019
	SupDevID         = 3019
	HgetXCJinfo      = 3020
	SupXCJinfo       = 3020
	HsetXCJparameter = 3021
	// interactive cmd
	Sresponse   = 9011
	SexecResult = 9012
	HnoticeACK  = 9013
	HdataACK    = 9014
)

// ST
var systemCodeStr = map[uint16]string{
	21: "地表水质量监测",
	22: "空气质量监测",
	23: "声环境质量监测",
	31: "大气环境污染源",
	32: "地表水质污染源",
	99: "餐饮油烟污染源",
	91: "系统交互",
}

var execResultCodeStr = map[uint16]string{
	1:   "执行成功",
	2:   "执行失败,但不知道原因",
	3:   "命令请求条件错误",
	100: "没有数据",
}

var reqCmdReturn = map[uint16]string{
	1: "准备执行请求",
	2: "请求被拒绝",
	3: "密码错误",
}

var dataTag = map[byte]string{
	'N': "在线监控仪器仪表工作正常",
	'P': "电源故障",
	'p': "电源故障",
	'B': "监测仪器发生故障",
	'b': "监测仪器发生故障",
	'D': "数据采集通道关闭",
	'd': "数据采集通道关闭",
	'C': "监测仪处于校准状态",
	'c': "监测仪处于校准状态",
	'H': "数据超出仪器量程上限",
	'h': "数据超出仪器量程上限",
	'L': "数据低于仪器量程上限",
	'l': "数据低于仪器量程上限",
	']': "数据高于用户自定义设定的上限",
	')': "数据高于用户自定义设定的上限",
	'[': "数据低于用户自定义设定的上限",
	'(': "数据低于用户自定义设定的上限",
	'F': "在线监控仪器仪表停运",
}

var cpkn = map[string]string{
	"SystemTime":  "系统时间",
	"QN":          "请求编号",
	"QnRtn":       "请求回应码",
	"ExeRtn":      "执行结果回应代码",
	"RtdInterval": "实时数据上报间隔",
}
