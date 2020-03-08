package packets

// CNType ...
type CNType uint16

// CN (Host command and slave req command)
const (
	// init cmd
	HsetTimeoutRepeats CNType = 1000
	HgetSlaveTime      CNType = 1011
	SupDate            CNType = 1011
	HsetSlaveTime      CNType = 1012
	SsetTimeReq        CNType = 1013
	HgetRealInterval   CNType = 1061
	SupRealInterval    CNType = 1061
	HsetRealInterval   CNType = 1062
	HgetMinuteInterval CNType = 1063
	SupMinuteInterval  CNType = 1063
	HsetMinuteInterval CNType = 1064
	HsetSlaveRestart   CNType = 1070
	HresetSlavePasswd  CNType = 1072
	// data cmd
	// real time data
	HgetRealTimeData  CNType = 2011
	SupRealTimeData   CNType = 2011
	HStopRealTimeData CNType = 2012
	// device status
	HgetDevStat  CNType = 2021
	SupDevSrat   CNType = 2021
	HStopDevStat CNType = 2022
	// day data
	HgetDayHistory           CNType = 2031
	SupDayHistory            CNType = 2031
	HgetDevRunTimeDayHistory CNType = 2041
	SupDevRunTimeDayHistory  CNType = 2041
	// minute data
	HgetMinHistory CNType = 2051
	SupMinHistory  CNType = 2051
	// hour data
	HgetHourHistory CNType = 2061
	SupHourHistory  CNType = 2061
	// orther data
	SupSCYupTime CNType = 2081
	// control cmd
	HzeroCal      CNType = 3011
	HrealSampling CNType = 3012
	HcleanStart   CNType = 3013
	HcompSampling CNType = 3014
	HkeepSample   CNType = 3015
	//.....
	HgetDevID        CNType = 3019
	SupDevID         CNType = 3019
	HgetXCJinfo      CNType = 3020
	SupXCJinfo       CNType = 3020
	HsetXCJparameter CNType = 3021
	// interactive cmd
	Sresponse   CNType = 9011
	SexecResult CNType = 9012
	HnoticeACK  CNType = 9013
	HdataACK    CNType = 9014
	HGetConfig  CNType = 9017
	HSetConfig  CNType = 9018
	SPutConfig  CNType = 9018
)

// ST
var systemCodeStr = map[string]string{
	"21": "地表水质量监测",
	"22": "空气质量监测",
	"23": "声环境质量监测",
	"31": "大气环境污染源",
	"32": "地表水质污染源",
	"99": "餐饮油烟污染源",
	"91": "系统交互",
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
