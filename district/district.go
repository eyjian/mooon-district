// Package district
// Wrote by yijian on 2024/03/09
package district

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type Table struct {
    ProvinceDistrictTable map[uint32]ProvinceDistrict `json:"province_list"`
}

// ProvinceDistrict 省/自治区/直辖市
type ProvinceDistrict struct {
    Code              uint32                  `json:"code"`
    Name              string                  `json:"name"`  // 行政区名称
    Level             uint32                  `json:"level"` // 行政区级别（1 省/自治区/直辖市，2 市/州/盟，3 县/县级市/旗）
    CityDistrictTable map[uint32]CityDistrict `json:"city_list"`
}

// CityDistrict 市/州/盟
type CityDistrict struct {
    Code                uint32              `json:"code"`
    Name                string              `json:"name"`  // 行政区名称
    Level               uint32              `json:"level"` // 行政区级别（1 省/自治区/直辖市，2 市/州/盟，3 县/县级市/旗）
    CountyDistrictTable map[uint32]District `json:"county_list"`
}

type District struct {
    Code        uint32 `json:"code"`        // 行政区代码
    Name        string `json:"name"`        // 行政区名称
    Level       uint32 `json:"level"`       // 行政区级别（1 省/自治区/直辖市，2 市/州/盟，3 县/县级市/旗）
    Parent      uint32 `json:"parent"`      // 父行政区代码
    Grandparent uint32 `json:"grandparent"` // 父父行政区代码
}

func LoadDistrict(ctx context.Context, filepath string) (*Table, error) {
    var districtTable Table

    // 打开文件
    file, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // 创建一个带缓冲的读取器
    reader := bufio.NewReader(file)

    // 按行读取文件内容
    lineNo := 0
    districtTable.ProvinceDistrictTable = make(map[uint32]ProvinceDistrict)
    for {
        lineNo = lineNo + 1
        line, err := reader.ReadString('\n')
        if err != nil {
            return nil, err
        }

        line = strings.Trim(line, "\n")
        line = strings.Trim(line, "\r")
        line = strings.TrimSpace(line)
        if len(line) == 0 {
            continue
        }

        district, err := parseLine(lineNo, line)
        if err != nil {
            return nil, err
        } else {
            if district == nil {
                continue
            }
            fmt.Println(*district)
            if district.Code == district.Parent && district.Code == district.Grandparent {
                // 省/自治区/直辖市
                provinceDistrict := ProvinceDistrict{
                    Code:              district.Code,
                    Name:              district.Name,
                    Level:             district.Level,
                    CityDistrictTable: make(map[uint32]CityDistrict),
                }
                districtTable.ProvinceDistrictTable[district.Code] = provinceDistrict
            } else if district.Code == district.Parent && district.Code != district.Grandparent {
                // 市/州/盟
                cityDistrict := CityDistrict{
                    Code:                district.Code,
                    Name:                district.Name,
                    Level:               district.Level,
                    CountyDistrictTable: make(map[uint32]District),
                }
                districtTable.ProvinceDistrictTable[district.Grandparent].CityDistrictTable[district.Code] = cityDistrict
            } else {
                countyDistrictTable, ok := districtTable.ProvinceDistrictTable[district.Grandparent].CityDistrictTable[district.Code]
                if ok {
                    // 县/县级市/旗
                    countyDistrictTable.CountyDistrictTable[district.Code] = *district
                } else {
                    // 直辖市的区县
                    cityDistrict := CityDistrict{
                        Code:                district.Code,
                        Name:                district.Name,
                        Level:               district.Level,
                        CountyDistrictTable: make(map[uint32]District),
                    }
                    districtTable.ProvinceDistrictTable[district.Grandparent].CityDistrictTable[district.Code] = cityDistrict
                }
            }
        }
    }

    return &districtTable, nil
}

func parseLine(lineNo int, line string) (*District, error) {
    // 使用逗号分隔每行数据
    parts := strings.Split(line, ",")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid row format: (%d) %s", lineNo, line)
    }

    // 解析行政区代码
    code, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
    if err != nil {
        if lineNo == 1 {
            return nil, nil
        }
        if len(parts[0]) == 0 {
            // 西沙区
            return nil, nil
        }
        return nil, fmt.Errorf("invalid district code: (%d) %s (%s)", lineNo, parts[0], line)
    }

    // 解析行政区名称
    name := strings.TrimSpace(parts[1])
    grandparent := (code / 10000) * 10000
    parent := (code / 100) * 100

    // 计算行政区级别
    level := uint32(3)
    if code%10000 == 0 {
        level = 1
    } else if code%100 == 0 {
        level = 2
    }

    return &District{
        Code:        uint32(code),
        Name:        name,
        Level:       level,
        Parent:      uint32(parent),
        Grandparent: uint32(grandparent),
    }, nil
}
