// Package district
// Wrote by yijian on 2024/03/09
package district

import "strings"

// PerfectNameOfProvincialAdminRegion 完美省级行政区名
// name 省级行政区全名或者简名（非简称）
// 数据参考：https://www.gov.cn/test/2006-04/04/content_244716.htm
func PerfectNameOfProvincialAdminRegion(name string) string {
    // 五大自治区
    if name == "内蒙古" {
        return "内蒙古自治区"
    }
    if name == "广西" {
        return "广西壮族自治区"
    }
    if name == "西藏" {
        return "西藏自治区"
    }
    if name == "宁夏" {
        return "宁夏回族自治区"
    }
    if name == "新疆" {
        return "新疆维吾尔自治区"
    }

    // 四大直辖市
    if name == "北京" {
        return "北京市"
    }
    if name == "上海" {
        return "上海市"
    }
    if name == "天津" {
        return "天津市"
    }
    if name == "重庆" {
        return "重庆市"
    }

    // 其它省份
    if strings.HasSuffix(name, "省") {
        return name
    } else {
        return name + "省"
    }
}
