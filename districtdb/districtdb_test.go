// Package districtdb
// Wrote by yijian on 2024/03/09
package districtdb

import (
    "context"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "os"
    "testing"
)

// go test -v -run="TestGetDistrictCode" -args 'username:password@tcp(host:port)/dbname?charset=utf8mb4'
func TestGetDistrictCode(t *testing.T) {
    ctx := context.Background()
    dsn := os.Args[len(os.Args)-1]
    t.Logf("%s\n", dsn)

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Error("failed to connect database")
    } else {
        tableName := "t_dict_district"
        query := NewQuery(db, tableName)
        name := &DistrictName{
            ProvinceName: "广东省",
            CityName:     "珠海市",
            CountyName:   "香洲区",
        }
        queryDistrictCode(t, ctx, query, tableName, name, 1)

        name.CountyName = "香洲区X"
        queryDistrictCode(t, ctx, query, tableName, name, 2)

        query.TableName = "t_dict_districtX"
        name.CountyName = "香洲区"
        queryDistrictCode(t, ctx, query, tableName, name, 3)

        query.TableName = "t_dict_district"
        name.ProvinceName = "广东省X"
        queryDistrictCode(t, ctx, query, tableName, name, 2)

        name.ProvinceName = "广东省"
        name.CityName = "珠海市X"
        queryDistrictCode(t, ctx, query, tableName, name, 2)

        name.ProvinceName = "海南省"
        name.CityName = "文昌市"
        name.CountyName = ""
        queryDistrictCode(t, ctx, query, tableName, name, 1)

        name.ProvinceName = "北京市"
        name.CityName = "海淀区"
        queryDistrictCode(t, ctx, query, tableName, name, 1)
        queryDistrictCode(t, ctx, query, tableName, name, 1)

        name.ProvinceName = "河南省"
        name.CityName = "济源市"
        queryDistrictCode(t, ctx, query, tableName, name, 1)

        name.ProvinceName = "台湾省"
        name.CityName = "台北市"
        queryDistrictCode(t, ctx, query, tableName, name, 2)

        name.ProvinceName = "香港特别行政区"
        name.CityName = "香港岛"
        queryDistrictCode(t, ctx, query, tableName, name, 2)

        name.ProvinceName = "澳门特别行政区"
        name.CityName = "澳门半岛"
        queryDistrictCode(t, ctx, query, tableName, name, 2)

        CacheMetricFPrintf(os.Stdout)
    }
}

// go test -v -run="TestGetDistrictName" -args 'username:password@tcp(host:port)/dbname?charset=utf8mb4'
func TestGetDistrictName(t *testing.T) {
    ctx := context.Background()
    dsn := os.Args[len(os.Args)-1]
    t.Logf("%s\n", dsn)

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Error("failed to connect database")
    } else {
        tableName := "t_dict_district"
        query := NewQuery(db, tableName)
        code := &DistrictCode{
            ProvinceCode: 440000,
            CityCode:     440400,
            CountyCode:   440402,
        }
        queryDistrictName(t, ctx, query, tableName, code, 1)
        queryDistrictName(t, ctx, query, tableName, code, 1)
        queryDistrictName(t, ctx, query, tableName, code, 1)

        code.CountyCode = 4404020
        queryDistrictName(t, ctx, query, tableName, code, 2)

        query.TableName = "t_dict_districtX"
        queryDistrictName(t, ctx, query, tableName, code, 3)

        query.TableName = "t_dict_district"
        code.ProvinceCode = 4400000
        queryDistrictName(t, ctx, query, tableName, code, 2)

        code.ProvinceCode = 440000
        code.CityCode = 4404000
        queryDistrictName(t, ctx, query, tableName, code, 2)

        code.ProvinceCode = 460000
        code.CityCode = 469000
        code.CountyCode = 0
        queryDistrictName(t, ctx, query, tableName, code, 2)

        code.ProvinceCode = 110000
        code.CityCode = 110100
        queryDistrictName(t, ctx, query, tableName, code, 2)

        CacheMetricFPrintf(os.Stdout)
    }
}

// queryDistrictCode 查询行政区代码
// expect 取值：
// 1）期待成功
// 2）期待不存在
// 3）期待出错
func queryDistrictCode(t *testing.T, ctx context.Context, query *Query, tableName string, name *DistrictName, expect int) {
    result, err := query.GetDistrictCode(ctx, name)
    if err != nil {
        if expect == 3 {
            t.Logf("[%s,%s,expect:%d] error: %s\n", tableName, *name, expect, err.Error())
        } else {
            t.Errorf("[%s,%s,expect:%d] error: %s\n", tableName, *name, expect, err.Error())
        }
    } else {
        if result == nil {
            if expect == 2 {
                t.Logf("[%s,%s,expect:%d] not found", tableName, *name, expect)
            } else {
                t.Errorf("[%s,%s,expect:%d] not found", tableName, *name, expect)
            }
        } else {
            if expect == 1 {
                t.Logf("[%s,%s,expect:%d] ProvinceCode: %d, CityCode: %d, CountyCode: %d\n", tableName, *name, expect, result.ProvinceCode, result.CityCode, result.CountyCode)
            } else {
                t.Errorf("[%s,%s,expect:%d] ProvinceCode: %d, CityCode: %d, CountyCode: %d\n", tableName, *name, expect, result.ProvinceCode, result.CityCode, result.CountyCode)
            }
        }
    }
}

// queryDistrictName 查询行政区名
// expect 取值：
// 1）期待成功
// 2）期待不存在
// 3）期待出错
func queryDistrictName(t *testing.T, ctx context.Context, query *Query, tableName string, code *DistrictCode, expect int) {
    result, err := query.GetDistrictName(ctx, code)
    if err != nil {
        if expect == 3 {
            t.Logf("[%s,%d,%d,%d,expect:%d] error: %s\n", tableName, code.ProvinceCode, code.CityCode, code.CountyCode, expect, err.Error())
        } else {
            t.Errorf("[%s,%d,%d,%d,expect:%d] error: %s\n", tableName, code.ProvinceCode, code.CityCode, code.CountyCode, expect, err.Error())
        }
    } else {
        if result == nil {
            if expect == 2 {
                t.Logf("[%s,%d,%d,%d,expect:%d] not found", tableName, code.ProvinceCode, code.CityCode, code.CountyCode, expect)
            } else {
                t.Errorf("[%s,%d,%d,%d,expect:%d] not found", tableName, code.ProvinceCode, code.CityCode, code.CountyCode, expect)
            }
        } else {
            if expect == 1 {
                t.Logf("[%s,%d,%d,%d,expect:%d] ProvinceName: %s, CityName: %s, CountyName: %s\n", tableName, code.ProvinceCode, code.CityCode, code.CountyCode, expect, result.ProvinceName, result.CityName, result.CountyName)
            } else {
                t.Errorf("[%s,%d,%d,%d,expect:%d] ProvinceName: %s, CityName: %s, CountyName: %s\n", tableName, code.ProvinceCode, code.CityCode, code.CountyCode, expect, result.ProvinceName, result.CityName, result.CountyName)
            }
        }
    }
}
