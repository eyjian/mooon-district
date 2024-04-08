// Package district
// Wrote by yijian on 2024/03/09
package district

import (
    "context"
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "github.com/coocood/freecache"
    "gorm.io/gorm"
    "io"
    "runtime/debug"
    "strconv"
    "strings"
)

var districtCache *freecache.Cache

// Query 行政区查询器
type Query struct {
    Db            *gorm.DB
    TableName     string
    ExpireSeconds int
}

// CacheMetric 缓存的度量数据
type CacheMetric struct {
    EntryCount        int64   `json:"entry_count"`
    ExpiredCount      int64   `json:"expired_count"`
    EvacuateCount     int64   `json:"evacuate_count"`
    LookupCount       int64   `json:"lookup_count"`
    AverageAccessTime int64   `json:"average_access_time"`
    HitCount          int64   `json:"hit_count"`
    MissCount         int64   `json:"miss_count"`
    OverwriteCount    int64   `json:"overwrite_count"`
    TouchedCount      int64   `json:"touched_count"`
    HitRate           float64 `json:"hit_rate"`
}

// Name 行政区名（字符集需同 DB 等保持一致）
type Name struct {
    ProvinceName string `gorm:"column:f_province_name" json:"province_name,omitempty"`
    CityName     string `gorm:"column:f_city_name" json:"city_name,omitempty"`
    CountyName   string `gorm:"column:f_county_name" json:"county_name,omitempty"`
}

type Code struct {
    ProvinceCode uint32 `gorm:"column:f_province_code" json:"province_code,omitempty"`
    CityCode     uint32 `gorm:"column:f_city_code" json:"city_code,omitempty"`
    CountyCode   uint32 `gorm:"column:f_county_code" json:"county_code,omitempty"`
}

func init() {
    cacheSize := 100 * 1024
    districtCache = freecache.NewCache(cacheSize)
    debug.SetGCPercent(20)
}

func NewQuery(db *gorm.DB, tableName string, expireSeconds int) *Query {
    return &Query{
        Db:            db,
        TableName:     tableName,
        ExpireSeconds: expireSeconds,
    }
}

func (d *Name) Md5Sum() string {
    data := d.ProvinceName + ":" + d.CityName + ":" + d.CountyName
    return Md5Sum(data)
}

func (d *Code) Md5Sum() string {
    data := fmt.Sprintf("%d:%d:%d", d.ProvinceCode, d.CityCode, d.CountyCode)
    return Md5Sum(data)
}

func Md5Sum(data string) string {
    hash := md5.Sum([]byte(data))
    return strings.ToLower(hex.EncodeToString(hash[:]))
}

func (c *CacheMetric) String() (string, error) {
    jsonBytes, err := json.Marshal(*c)
    if err != nil {
        return "", fmt.Errorf("json marshal error: %s", err.Error())
    }
    return string(jsonBytes), nil
}

func CacheMetricFPrintf(w io.Writer) {
    cacheMetric := GetCacheMetric()
    str, err := cacheMetric.String()
    if err != nil {
        fmt.Fprintf(w, "%s\n", err.Error())
    } else {
        fmt.Fprintf(w, "%s\n", str)
    }
}

func GetCacheMetric() *CacheMetric {
    return &CacheMetric{
        EntryCount:        districtCache.EntryCount(),
        ExpiredCount:      districtCache.ExpiredCount(),
        EvacuateCount:     districtCache.EvacuateCount(),
        LookupCount:       districtCache.LookupCount(),
        AverageAccessTime: districtCache.AverageAccessTime(),
        HitCount:          districtCache.HitCount(),
        MissCount:         districtCache.MissCount(),
        OverwriteCount:    districtCache.OverwriteCount(),
        TouchedCount:      districtCache.TouchedCount(),
        HitRate:           districtCache.HitRate(),
    }
}

// GetDistrictCode 通过行政区名取得行政区代码
// 返回值：
// 1）成功返回非 nil 的 DistrictCode，同时 error 值为 nil ；
// 2）不存在返回 nil 的 DistrictCode，同时 error 值为 nil ；
// 3）出错返回 nil 的 DistrictCode，同时 error 值不为 nil 。
func (q *Query) GetDistrictCode(ctx context.Context, name *Name) (*Code, error) {
    code, err := q.getDistrictCodeFromCache(name)
    if err == nil {
        return code, nil
    }

    code, err = q.getDistrictCodeFromDb(ctx, name)
    if err == nil && code != nil {
        q.updateDistrictCodeToCache(name, code)
    }
    return code, err
}

// GetDistrictName 通过行政区代码取得行政区名
// 返回值：
// 1）成功返回非 nil 的 DistrictName，同时 error 值为 nil ；
// 2）不存在返回 nil 的 DistrictName，同时 error 值为 nil ；
// 3）出错返回 nil 的 DistrictName，同时 error 值不为 nil 。
func (q *Query) GetDistrictName(ctx context.Context, code *Code) (*Name, error) {
    name, err := q.getDistrictNameFromCache(code)
    if err == nil {
        return name, nil
    }

    name, err = q.getDistrictNameFromDb(ctx, code)
    if err == nil && name != nil {
        q.updateDistrictNameToCache(code, name)
    }
    return name, err
}

// GetCountyCount 取得县/县级市/旗数，像东莞市没有
func (q *Query) GetCountyCount(ctx context.Context, provinceName, cityName string) (int, error) {
    count, err := q.getCountyCountFromCache(provinceName, cityName)
    if err == nil {
        return count, nil
    }

    count, err = q.getCountyCountFromDb(ctx, provinceName, cityName)
    if err == nil {
        q.updateCountyCountToCache(provinceName, cityName, count)
    }
    return count, err
}

func (q *Query) getDistrictCodeFromDb(ctx context.Context, name *Name) (*Code, error) {
    var code Code
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

func (q *Query) getDistrictNameFromDb(ctx context.Context, code *Code) (*Name, error) {
    var result Name
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

func (q *Query) getCountyCountFromDb(ctx context.Context, provinceName, cityName string) (int, error) {
    var count int64

    err := q.Db.Table(q.TableName).
        Where("f_province_name = ? AND f_city_name = ?", provinceName, cityName).
        Count(&count).
        Error
    if err != nil {
        return 0, err
    }

    return int(count), nil
}

func (q *Query) getDistrictCodeFromCache(name *Name) (*Code, error) {
    var code Code
    cacheKey := name.Md5Sum()
    jsonBytes, err := districtCache.Get([]byte(cacheKey))
    if err != nil {
        return nil, fmt.Errorf("cache get error: %s", err.Error())
    }

    err = json.Unmarshal(jsonBytes, code)
    if err != nil {
        return nil, fmt.Errorf("cache json unmarshal error: %s", err.Error())
    }

    return &code, nil
}

func (q *Query) updateDistrictCodeToCache(name *Name, code *Code) error {
    cacheKey := name.Md5Sum()
    jsonBytes, err := json.Marshal(*code)
    if err != nil {
        return fmt.Errorf("cache json marshal error: %s", err.Error())
    }

    err = districtCache.Set([]byte(cacheKey), jsonBytes, q.ExpireSeconds)
    if err != nil {
        return fmt.Errorf("cache set error: %s", err.Error())
    }

    return nil
}

func (q *Query) getDistrictNameFromCache(code *Code) (*Name, error) {
    var name Name
    cacheKey := code.Md5Sum()
    jsonBytes, err := districtCache.Get([]byte(cacheKey))
    if err != nil {
        return nil, fmt.Errorf("cache get error: %s", err.Error())
    }

    err = json.Unmarshal(jsonBytes, name)
    if err != nil {
        return nil, fmt.Errorf("cache json unmarshal error: %s", err.Error())
    }

    return &name, nil
}

func (q *Query) updateDistrictNameToCache(code *Code, name *Name) error {
    cacheKey := code.Md5Sum()
    jsonBytes, err := json.Marshal(*name)
    if err != nil {
        return fmt.Errorf("cache json marshal error: %s", err.Error())
    }

    err = districtCache.Set([]byte(cacheKey), jsonBytes, q.ExpireSeconds)
    if err != nil {
        return fmt.Errorf("cache set error: %s", err.Error())
    }

    return nil
}

func (q *Query) getCountyCountFromCache(provinceName, cityName string) (int, error) {
    key := "CountyCount:" + provinceName + ":" + cityName
    cacheKey := Md5Sum(key)

    data, err := districtCache.Get([]byte(cacheKey))
    if err != nil {
        return 0, fmt.Errorf("cache get error: %s", err.Error())
    }

    intValue, err := strconv.Atoi(string(data))
    if err != nil {
        return 0, fmt.Errorf("cache json strconv error: %s", err.Error())
    }

    return intValue, nil
}

func (q *Query) updateCountyCountToCache(provinceName, cityName string, countyCount int) error {
    key := "CountyCount:" + provinceName + ":" + cityName
    cacheKey := Md5Sum(key)

    err := districtCache.Set([]byte(cacheKey), []byte(strconv.Itoa(countyCount)), q.ExpireSeconds)
    if err != nil {
        return fmt.Errorf("cache set error: %s", err.Error())
    }

    return nil
}