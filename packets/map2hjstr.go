package packets

import (
	"fmt"
	"strconv"
	"strings"
)

func if2string(ifa interface{}) (rs string) {
	switch ifa.(type) {
	case int:
		return strconv.Itoa(ifa.(int))
	case int16:
		return strconv.Itoa(int(ifa.(int16)))
	case uint16:
		return strconv.Itoa(int(ifa.(uint16)))
	case float32:
		return fmt.Sprintf("%v", ifa.(float32))
	case float64:
		//return strconv.FormatFloat(ifa.(float64), 'f', 2, 64), nil
		return fmt.Sprintf("%v", ifa.(float64))
	case string:
		return ifa.(string)
	default:
		return ""
	}
}

func cpkv2str(cpkv CPkv) (rs string) {
	//rtdPattern := regexp.MustCompile(`.+Rtd$`)
	//idPattern := regexp.MustCompile(`.+ID$`)
	//var rtd, id string
	//for k, n := range cpkv {
	//if rtdPattern.MatchString(k) {
	//rtd = fmt.Sprintf("%s=%s%c", k, if2string(n), ',')
	//}
	//if idPattern.MatchString(k) {
	//id = fmt.Sprintf("%s=%s%c", k, if2string(n), ',')
	//}
	//}
	//rs = rtd + id
	for k, n := range cpkv {
		rs += fmt.Sprintf("%s=%s%c", k, if2string(n), ',')
	}
	return
}

func cpkvg2str(cpkvg CPkvg) (rs string) {
	for _, cpkv := range cpkvg {
		temp := strings.TrimRight(cpkv2str(cpkv), ",")
		rs += fmt.Sprintf("%s%c", temp, ';')
	}
	rs = strings.TrimRight(rs, ";")
	return
}
