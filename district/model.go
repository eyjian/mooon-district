// Package district
// Wrote by yijian on 2024/09/04
package district

// DictDistrict Generated by sql2struct
type DictDistrict struct {
	ProvinceCode uint32 `gorm:"column:f_province_code" json:"province_code"`
	CityCode     uint32 `gorm:"column:f_city_code" json:"city_code"`
	CountyCode   uint32 `gorm:"column:f_county_code" json:"county_code"`
	Level        uint32 `gorm:"column:f_level" json:"level"`
	ProvinceName string `gorm:"column:f_province_name" json:"province_name"`
	CityName     string `gorm:"column:f_city_name" json:"city_name"`
	CountyName   string `gorm:"column:f_county_name" json:"county_name"`
}