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
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
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

// 初始化缓存
func init() {
	cacheSize := 1024 * 1024
	districtCache = freecache.NewCache(cacheSize)
	debug.SetGCPercent(20)
}

// NewQuery 新建查询对象
// expireSeconds 缓存时长（单位为秒），值小于 60 时会强制设置为 60，建议为 3600 或者更大值，因为行政区数据更新频率极低
func NewQuery(db *gorm.DB, tableName string, expireSeconds int) *Query {
	seconds := expireSeconds
	if seconds < 60 {
		seconds = 60
	}

	return &Query{
		Db:            db,
		TableName:     tableName,
		ExpireSeconds: seconds,
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

// Load2Cache 从数据库加载数据到缓存
func (q *Query) Load2Cache() (int, error) {
	var results []DictDistrict

	// 从数据库查询数据
	err := q.Db.Table(q.TableName).Find(&results).Error
	if err != nil {
		return 0, fmt.Errorf("load from table://%s error: %s", q.TableName, err.Error())
	}

	// 调用 updateDistrictCodeToCache 和 updateDistrictNameToCache 将数据更新到缓存
	for _, result := range results {
		// 从数据库中取得行政区代码和行政区名
		name := Name{
			ProvinceName: result.ProvinceName,
			CityName:     result.CityName,
			CountyName:   result.CountyName,
		}

		// 从数据库中取得行政区代码和行政区名
		code := Code{
			ProvinceCode: result.ProvinceCode,
			CityCode:     result.CityCode,
			CountyCode:   result.CountyCode,
		}

		// 从缓存中取得行政区代码和行政区名
		err = q.updateDistrictCodeToCache(&name, &code)
		if err != nil {
			return 0, fmt.Errorf("set code to cache error: %s", err.Error())
		}

		// 从缓存中取得行政区代码和行政区名
		err = q.updateDistrictNameToCache(&code, &name)
		if err != nil {
			return 0, fmt.Errorf("set name to cache error: %s", err.Error())
		}
	}

	// 返回成功条数
	return len(results), nil
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
		_ = q.updateDistrictCodeToCache(name, code)
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
		_ = q.updateDistrictNameToCache(code, name)
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
		_ = q.updateCountyCountToCache(provinceName, cityName, count)
	}
	return count, err
}

// getDistrictCodeFromDb 从数据库中取得行政区代码
func (q *Query) getDistrictCodeFromDb(ctx context.Context, name *Name) (*Code, error) {
	var code Code
	err := q.Db.WithContext(ctx).Table(q.TableName).
		Select("f_province_code", "f_city_code", "f_county_code").
		Where("f_province_name = ? AND f_city_name = ? AND f_county_name = ?", name.ProvinceName, name.CityName, name.CountyName).
		First(&code).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 不存在
		}
		return nil, err
	}

	return &code, nil
}

// getDistrictNameFromDb 从数据库中取得行政区名
func (q *Query) getDistrictNameFromDb(ctx context.Context, code *Code) (*Name, error) {
	var result Name
	err := q.Db.WithContext(ctx).Table(q.TableName).
		Select("f_province_name", "f_city_name", "f_county_name").
		Where("f_province_code = ? AND f_city_code = ? AND f_county_code = ?", code.ProvinceCode, code.CityCode, code.CountyCode).
		First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 不存在
		}
		return nil, err
	}

	return &result, nil
}

// getCountyCountFromDb 从数据库中取得区县数量行政区代码
func (q *Query) getCountyCountFromDb(ctx context.Context, provinceName, cityName string) (int, error) {
	var count int64

	err := q.Db.WithContext(ctx).Table(q.TableName).
		Where("f_province_name = ? AND f_city_name = ?", provinceName, cityName).
		Count(&count).
		Error
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// getDistrictCodeFromCache 从缓存中取得行政区代码
func (q *Query) getDistrictCodeFromCache(name *Name) (*Code, error) {
	var code Code
	cacheKey := name.Md5Sum()
	jsonBytes, err := districtCache.Get([]byte(cacheKey))
	if err != nil {
		return nil, fmt.Errorf("cache get error: %s", err.Error())
	}

	err = json.Unmarshal(jsonBytes, &code)
	if err != nil {
		return nil, fmt.Errorf("cache json unmarshal error: %s", err.Error())
	}

	return &code, nil
}

// updateDistrictCodeToCache 将行政区代码更新到缓存
func (q *Query) updateDistrictCodeToCache(name *Name, code *Code) error {
	cacheKey := name.Md5Sum()
	jsonBytes, err := json.Marshal(*code)
	if err != nil {
		return fmt.Errorf("cache json marshal error: %s", err.Error())
	}

	randSeconds := getRandSeconds(q.ExpireSeconds)
	err = districtCache.Set([]byte(cacheKey), jsonBytes, q.ExpireSeconds+randSeconds)
	if err != nil {
		return fmt.Errorf("cache set error: %s", err.Error())
	}

	return nil
}

// getDistrictNameFromCache 从缓存中取得行政区名
func (q *Query) getDistrictNameFromCache(code *Code) (*Name, error) {
	var name Name
	cacheKey := code.Md5Sum()
	jsonBytes, err := districtCache.Get([]byte(cacheKey))
	if err != nil {
		return nil, fmt.Errorf("cache get error: %s", err.Error())
	}

	err = json.Unmarshal(jsonBytes, &name)
	if err != nil {
		return nil, fmt.Errorf("cache json unmarshal error: %s", err.Error())
	}

	return &name, nil
}

// updateDistrictNameToCache 将行政区名更新到缓存
func (q *Query) updateDistrictNameToCache(code *Code, name *Name) error {
	cacheKey := code.Md5Sum()
	jsonBytes, err := json.Marshal(*name)
	if err != nil {
		return fmt.Errorf("cache json marshal error: %s", err.Error())
	}

	randSeconds := getRandSeconds(q.ExpireSeconds)
	err = districtCache.Set([]byte(cacheKey), jsonBytes, q.ExpireSeconds+randSeconds)
	if err != nil {
		return fmt.Errorf("cache set error: %s", err.Error())
	}

	return nil
}

// getCountyCountFromCache 从缓存中取得区县数量
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

// updateCountyCountToCache 将区县数量更新到缓存
func (q *Query) updateCountyCountToCache(provinceName, cityName string, countyCount int) error {
	key := "CountyCount:" + provinceName + ":" + cityName
	cacheKey := Md5Sum(key)

	randSeconds := getRandSeconds(q.ExpireSeconds)
	err := districtCache.Set([]byte(cacheKey), []byte(strconv.Itoa(countyCount)), q.ExpireSeconds+randSeconds)
	if err != nil {
		return fmt.Errorf("cache set error: %s", err.Error())
	}

	return nil
}

// getRandSeconds 随机获取缓存过期时间，防止同一时间过期
func getRandSeconds(expireSeconds int) int {
	randSeconds := 1
	source := rand.NewSource(time.Now().UnixNano())

	if expireSeconds >= 3600 {
		randSeconds = int(source.Int63() % 600)
	} else if expireSeconds >= 600 {
		randSeconds = int(source.Int63() % 60)
	} else if expireSeconds >= 60 {
		randSeconds = int(source.Int63() % 6)
	}

	return randSeconds
}
