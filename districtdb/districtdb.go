// Package districtdb
// Wrote by yijian on 2024/03/09
package districtdb

import (
    "context"
    "gorm.io/gorm"
)

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

// GetDistrictId 通过行政区名取得行政区代码
// 返回值：
// 1）成功返回非 nil 的 DistrictCode，同时 error 值为 nil ；
// 2）不存在返回 nil 的 DistrictCode，同时 error 值为 nil ；
// 3）出错返回 nil 的 DistrictCode，同时 error 值不为 nil 。
func GetDistrictId(ctx context.Context, db *gorm.DB, tableName string, name *DistrictName) (*DistrictCode, error) {
    var result DistrictCode
    err := db.Table(tableName).
        Select("f_province_code", "f_city_code", "f_county_code").
        Where("f_province_name = ? AND f_city_name = ? AND f_county_name = ?", name.ProvinceName, name.CityName, name.CountyName).
        First(&result).Error

    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // 不存在
        }
        return nil, err
    }

    return &result, nil
}
