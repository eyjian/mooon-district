// Package districtdb
// Wrote by yijian on 2024/03/09
package districtdb

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "github.com/coocood/freecache"
    "gorm.io/gorm"
    "runtime/debug"
    "strings"
)

var districtCache *freecache.Cache

// Query 行政区查询器
type Query struct {
    Db            *gorm.DB
    TableName     string
    ExpireSeconds int
}

// DistrictName 行政区名（字符集需同 DB 等保持一致）
type DistrictName struct {
    ProvinceName string `gorm:"column:f_province_name" json:"province_name,omitempty"`
    CityName     string `gorm:"column:f_city_name" json:"city_name,omitempty"`
    CountyName   string `gorm:"column:f_county_name" json:"county_name,omitempty"`
}

type DistrictCode struct {
    ProvinceCode uint32 `gorm:"column:f_province_code" json:"province_code,omitempty"`
    CityCode     uint32 `gorm:"column:f_city_code" json:"city_code,omitempty"`
    CountyCode   uint32 `gorm:"column:f_county_code" json:"county_code,omitempty"`
}

func init() {
    cacheSize := 100 * 1024
    districtCache = freecache.NewCache(cacheSize)
    debug.SetGCPercent(20)
}

func NewQuery(db *gorm.DB, tableName string) *Query {
    return &Query{
        Db:            db,
        TableName:     tableName,
        ExpireSeconds: 3600 * 12,
    }
}

func (d *DistrictName) Md5Sum() string {
    data := d.ProvinceName + ":" + d.CityName + ":" + d.CountyName
    return Md5Sum(data)
}

func (d *DistrictCode) Md5Sum() string {
    data := fmt.Sprintf("%d:%d:%d", d.ProvinceCode, d.CityCode, d.CountyCode)
    return Md5Sum(data)
}

func Md5Sum(data string) string {
    hash := md5.Sum([]byte(data))
    return strings.ToLower(hex.EncodeToString(hash[:]))
}

// GetDistrictCode 通过行政区名取得行政区代码
// 返回值：
// 1）成功返回非 nil 的 DistrictCode，同时 error 值为 nil ；
// 2）不存在返回 nil 的 DistrictCode，同时 error 值为 nil ；
// 3）出错返回 nil 的 DistrictCode，同时 error 值不为 nil 。
func (q *Query) GetDistrictCode(ctx context.Context, name *DistrictName) (*DistrictCode, error) {
    code, err := q.getDistrictCodeFromCache(name)
    if err == nil {
        return code, nil
    }

    code, err = q.getDistrictCodeFromDb(ctx, name)
    if err != nil && code != nil {
        q.updateDistrictCodeToCache(name, code)
    }
    return code, err
}

// GetDistrictName 通过行政区代码取得行政区名
// 返回值：
// 1）成功返回非 nil 的 DistrictName，同时 error 值为 nil ；
// 2）不存在返回 nil 的 DistrictName，同时 error 值为 nil ；
// 3）出错返回 nil 的 DistrictName，同时 error 值不为 nil 。
func (q *Query) GetDistrictName(ctx context.Context, code *DistrictCode) (*DistrictName, error) {
    name, err := q.getDistrictNameFromCache(code)
    if err == nil {
        return name, nil
    }

    name, err = q.getDistrictNameFromDb(ctx, code)
    if err != nil && name != nil {
        q.updateDistrictNameToCache(code, name)
    }
    return name, err
}

func (q *Query) getDistrictCodeFromDb(ctx context.Context, name *DistrictName) (*DistrictCode, error) {
    var code DistrictCode
    err := q.Db.Table(q.TableName).
        Select("f_province_code", "f_city_code", "f_county_code").
        Where("f_province_name = ? AND f_city_name = ? AND f_county_name = ?", name.ProvinceName, name.CityName, name.CountyName).
        First(&code).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // 不存在
        }
        return nil, err
    }

    return &code, nil
}

func (q *Query) getDistrictNameFromDb(ctx context.Context, code *DistrictCode) (*DistrictName, error) {
    var result DistrictName
    err := q.Db.Table(q.TableName).
        Select("f_province_name", "f_city_name", "f_county_name").
        Where("f_province_code = ? AND f_city_code = ? AND f_county_code = ?", code.ProvinceCode, code.CityCode, code.CountyCode).
        First(&result).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // 不存在
        }
        return nil, err
    }

    return &result, nil
}

func (q *Query) getDistrictCodeFromCache(name *DistrictName) (*DistrictCode, error) {
    var code DistrictCode
    cacheKey := name.Md5Sum()
    jsonBytes, err := districtCache.Get([]byte(cacheKey))
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(jsonBytes, code)
    if err != nil {
        return nil, err
    }

    return &code, nil
}

func (q *Query) updateDistrictCodeToCache(name *DistrictName, code *DistrictCode) error {
    cacheKey := name.Md5Sum()
    jsonBytes, err := json.Marshal(*code)
    if err != nil {
        return err
    }

    err = districtCache.Set([]byte(cacheKey), jsonBytes, q.ExpireSeconds)
    if err != nil {
        return err
    }

    return nil
}

func (q *Query) getDistrictNameFromCache(code *DistrictCode) (*DistrictName, error) {
    var name DistrictName
    cacheKey := code.Md5Sum()
    jsonBytes, err := districtCache.Get([]byte(cacheKey))
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(jsonBytes, name)
    if err != nil {
        return nil, err
    }

    return &name, nil
}

func (q *Query) updateDistrictNameToCache(code *DistrictCode, name *DistrictName) error {
    cacheKey := code.Md5Sum()
    jsonBytes, err := json.Marshal(*name)
    if err != nil {
        return err
    }

    err = districtCache.Set([]byte(cacheKey), jsonBytes, q.ExpireSeconds)
    if err != nil {
        return err
    }

    return nil
}
