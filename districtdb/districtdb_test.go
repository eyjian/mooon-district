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

// go test -v -run="TestGetDistrictId" -args 'username:password@tcp(host:port)/dbname?charset=utf8mb4'
func TestGetDistrictId(t *testing.T) {
    ctx := context.Background()
    dsn := os.Args[len(os.Args)-1]
    t.Logf("%s\n", dsn)

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Error("failed to connect database")
    } else {
        tableName := "t_dict_district"
        name := &DistrictName{
            ProvinceName: "广东省",
            CityName:     "珠海市",
            CountyName:   "香洲区",
        }
        queryDistrictId(t, ctx, db, tableName, name, 1)

        name.CountyName = "香洲区X"
        queryDistrictId(t, ctx, db, tableName, name, 2)

        tableName = "t_dict_districtX"
        name.CountyName = "香洲区"
        queryDistrictId(t, ctx, db, tableName, name, 3)

        tableName = "t_dict_district"
        name.ProvinceName = "广东省X"
        queryDistrictId(t, ctx, db, tableName, name, 2)

        name.ProvinceName = "广东省"
        name.CityName = "珠海市X"
        queryDistrictId(t, ctx, db, tableName, name, 2)

        name.ProvinceName = "海南省"
        name.CityName = "文昌市"
        name.CountyName = ""
        queryDistrictId(t, ctx, db, tableName, name, 1)

        name.ProvinceName = "北京市"
        name.CityName = "海淀区"
        queryDistrictId(t, ctx, db, tableName, name, 1)

        name.ProvinceName = "河南省"
        name.CityName = "济源市"
        queryDistrictId(t, ctx, db, tableName, name, 1)
    }
}

// queryDistrictId 查询行政区代码
// expect 取值：
// 1）期待成功
// 2）期待不存在
// 3）期待出错
func queryDistrictId(t *testing.T, ctx context.Context, db *gorm.DB, tableName string, name *DistrictName, expect int) {
    result, err := GetDistrictId(ctx, db, tableName, name)
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
