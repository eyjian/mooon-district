// Package district
// Wrote by yijian on 2024/03/09
package district

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "golang.org/x/text/collate"
    "golang.org/x/text/language"
    "io"
    "os"
    "sort"
    "strconv"
    "strings"
)
import (
    excelize "github.com/xuri/excelize/v2"
)

type Table struct {
    ProvinceDistrictTable map[uint32]ProvinceDistrict `json:"-"`
    Provinces             []ProvinceDistrict          `json:"provinces,omitempty"`
}

// ProvinceDistrict 省/自治区/直辖市
type ProvinceDistrict struct {
    Code              uint32                  `json:"code"`
    Name              string                  `json:"name"`         // 行政区名称
    Level             uint32                  `json:"level"`        // 行政区级别（1 省/自治区/直辖市，2 市/州/盟，3 县/县级市/旗）
    Municipality      bool                    `json:"municipality"` // 直辖市
    CityDistrictTable map[uint32]CityDistrict `json:"-"`
    Cities            []CityDistrict          `json:"cities,omitempty"`
}

// CityDistrict 市/州/盟
type CityDistrict struct {
    Code                uint32              `json:"code"`
    Name                string              `json:"name"`        // 行政区名称
    Level               uint32              `json:"level"`       // 行政区级别（1 省/自治区/直辖市，2 市/州/盟，3 县/县级市/旗）
    CountyCity          bool                `json:"county_city"` // 县级市
    CountyDistrictTable map[uint32]District `json:"-"`
    Counties            []District          `json:"counties,omitempty"`
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
            if err == io.EOF {
                break
            }
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
            //fmt.Println(*district)

            provinceCode := getProvinceDistrictCode(district.Code)
            cityCode := getCityDistrictCode(district.Code)
            if IsProvinceDistrictCode(district.Code) {
                // 省/自治区/直辖市
                provinceDistrict := ProvinceDistrict{
                    Code:              district.Code,
                    Name:              district.Name,
                    Level:             district.Level,
                    CityDistrictTable: make(map[uint32]CityDistrict),
                    Municipality:      IsMunicipalityCode(district.Code),
                }
                districtTable.ProvinceDistrictTable[provinceCode] = provinceDistrict
            } else if IsCityDistrictCode(district.Code) {
                // 市/州/盟
                cityDistrict := CityDistrict{
                    Code:                district.Code,
                    Name:                district.Name,
                    Level:               district.Level,
                    CountyDistrictTable: make(map[uint32]District),
                    CountyCity:          false,
                }
                districtTable.ProvinceDistrictTable[provinceCode].CityDistrictTable[cityCode] = cityDistrict
            } else if IsCountyDistrictCode(district.Code) {
                if !IsMunicipalityCode(district.Code) {
                    // 非直辖市
                    if districtTable.ProvinceDistrictTable[provinceCode].CityDistrictTable[cityCode].CountyDistrictTable == nil {
                        // 省直辖县级市（济源市，河南省直辖县级市；五指山市，海南省直辖县级市）
                        cityDistrict := CityDistrict{
                            Code:  district.Code,
                            Name:  district.Name,
                            Level: district.Level,
                            //CountyDistrictTable: make(map[uint32]District),
                            CountyCity: true,
                        }
                        districtTable.ProvinceDistrictTable[provinceCode].CityDistrictTable[district.Code] = cityDistrict
                    } else {
                        // 县/县级市/旗
                        districtTable.ProvinceDistrictTable[provinceCode].CityDistrictTable[cityCode].CountyDistrictTable[district.Code] = *district
                    }
                } else {
                    // 直辖市的区县
                    cityDistrict := CityDistrict{
                        Code:                district.Code,
                        Name:                district.Name,
                        Level:               district.Level - 1,
                        CountyDistrictTable: make(map[uint32]District),
                    }
                    districtTable.ProvinceDistrictTable[provinceCode].CityDistrictTable[district.Code] = cityDistrict
                }
            } else {
                return nil, fmt.Errorf("invalid row data: (%d) %s", lineNo, line)
            }
        }
    }

    perfectTable(&districtTable)
    return &districtTable, nil
}

func GenerateJson(districtTable *Table, jsonFilepath string, withIndent bool, indent, prefix string) error {
    var err error
    var jsonBytes []byte
    filepath := jsonFilepath

    if !withIndent {
        jsonBytes, err = json.Marshal(*districtTable)
    } else {
        jsonBytes, err = json.MarshalIndent(*districtTable, prefix, indent)
    }
    if err != nil {
        return fmt.Errorf("json marshal error: %s", err.Error())
    }

    file, writer := createFile(filepath)
    if file == nil {
        return fmt.Errorf("create file://%s error: %s", filepath, err.Error())
    }
    defer file.Close()

    _, err = writer.WriteString(string(jsonBytes))
    if err != nil {
        return fmt.Errorf("write file://%s error: %s", filepath, err.Error())
    }
    err = writer.Flush()
    if err != nil {
        return fmt.Errorf("flush file://%s error: %s", filepath, err.Error())
    }

    return nil
}

func GenerateCsv(districtTable *Table, csvFilepath, csvDelimiter string, withCode bool) error {
    var err error
    var builder strings.Builder
    filepath := csvFilepath
    file, writer := createFile(filepath)
    if file == nil {
        return fmt.Errorf("create file://%s error: %s", filepath, err.Error())
    }
    defer file.Close()

    for _, provinceDistrict := range districtTable.Provinces {
        if !withCode {
            builder.WriteString(fmt.Sprintf("%s\n", provinceDistrict.Name))
        } else {
            builder.WriteString(fmt.Sprintf("%d%s%s\n", provinceDistrict.Code, csvDelimiter, provinceDistrict.Name))
        }
        for _, cityDistrict := range provinceDistrict.Cities {
            if !withCode {
                builder.WriteString(fmt.Sprintf("%s%s%s\n",
                    provinceDistrict.Name, csvDelimiter, cityDistrict.Name))
            } else {
                builder.WriteString(fmt.Sprintf("%d%s%s%s%s\n",
                    cityDistrict.Code, csvDelimiter,
                    provinceDistrict.Name, csvDelimiter, cityDistrict.Name))
            }

            for _, countyDistrict := range cityDistrict.Counties {
                if !withCode {
                    builder.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
                        provinceDistrict.Name, csvDelimiter,
                        cityDistrict.Name, csvDelimiter, countyDistrict.Name))
                } else {
                    builder.WriteString(fmt.Sprintf("%d%s%s%s%s%s%s\n",
                        countyDistrict.Code, csvDelimiter,
                        provinceDistrict.Name, csvDelimiter,
                        cityDistrict.Name, csvDelimiter, countyDistrict.Name))
                }
                if err != nil {
                    return fmt.Errorf("write file://%s error: %s", filepath, err.Error())
                }
            }
        }
    }
    _, err = writer.WriteString(builder.String())
    if err != nil {
        return fmt.Errorf("write file://%s error: %s", filepath, err.Error())
    }
    err = writer.Flush()
    if err != nil {
        return fmt.Errorf("flush file://%s error: %s", filepath, err.Error())
    }

    return nil
}

func GenerateSql(districtTable *Table, sqlFilepath, tableName string, withIgnore bool) error {
    var err error
    var builder strings.Builder
    filepath := sqlFilepath
    file, writer := createFile(filepath)
    if file == nil {
        return fmt.Errorf("create file://%s error: %s", filepath, err.Error())
    }
    defer file.Close()

    // 要求的表格式：
    builder.WriteString("/*\n")
    builder.WriteString("DROP TABLE t_dict_district;\n")
    builder.WriteString("CREATE TABLE t_dict_district (\n")
    builder.WriteString("  f_province_code INT UNSIGNED NOT NULL,\n")
    builder.WriteString("  f_city_code INT UNSIGNED NOT NULL,\n")
    builder.WriteString("  f_county_code INT UNSIGNED NOT NULL,\n")
    builder.WriteString("  f_level TINYINT UNSIGNED NOT NULL,\n")
    builder.WriteString("  f_province_name VARCHAR(20) NOT NULL,\n")
    builder.WriteString("  f_city_name VARCHAR(20) NOT NULL,\n")
    builder.WriteString("  f_county_name VARCHAR(20) NOT NULL,\n")
    builder.WriteString("  PRIMARY KEY (f_province_code,f_city_code,f_county_code),\n")
    builder.WriteString("  KEY (f_province_name),\n")
    builder.WriteString("  KEY (f_city_name),\n")
    builder.WriteString("  KEY (f_county_name)\n")
    builder.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;\n")
    builder.WriteString("*/\n")

    if withIgnore {
        builder.WriteString(fmt.Sprintf("INSERT IGNORE INTO %s VALUES \n", tableName))
    } else {
        builder.WriteString(fmt.Sprintf("INSERT INTO %s VALUES \n", tableName))
    }
    for _, provinceDistrict := range districtTable.Provinces {
        // 省/自治区/直辖市
        line := fmt.Sprintf("(%d,%d,%d,%d,'%s','%s','%s'),\n",
            provinceDistrict.Code, 0, 0, provinceDistrict.Level,
            provinceDistrict.Name, "", "")
        builder.WriteString(line)

        for _, cityDistrict := range provinceDistrict.Cities {
            // 市/州/盟
            line := fmt.Sprintf("(%d,%d,%d,%d,'%s','%s','%s'),\n",
                provinceDistrict.Code, cityDistrict.Code, 0, cityDistrict.Level,
                provinceDistrict.Name, cityDistrict.Name, "")
            builder.WriteString(line)

            for _, countyDistrict := range cityDistrict.Counties {
                // 县/县级市/旗
                line := fmt.Sprintf("(%d,%d,%d,%d,'%s','%s','%s'),\n",
                    provinceDistrict.Code, cityDistrict.Code, countyDistrict.Code, countyDistrict.Level,
                    provinceDistrict.Name, cityDistrict.Name, countyDistrict.Name)
                builder.WriteString(line)
            }
        }
    }

    sql := strings.Trim(builder.String(), "\n")
    sql = strings.Trim(sql, ",")
    sql = sql + ";"
    _, err = writer.WriteString(sql)
    if err != nil {
        return fmt.Errorf("write file://%s error: %s", filepath, err.Error())
    }

    err = writer.Flush()
    if err != nil {
        return fmt.Errorf("flush file://%s error: %s", filepath, err.Error())
    }

    return nil
}

func GenerateXlsx(districtTable *Table, xlsxFilepath string) error {
    ctx := context.Background()
    sheetName := "mooon-district"

    f, err := createXlsxFile(ctx, sheetName)
    if err != nil {
        return fmt.Errorf("create %s error: %s", xlsxFilepath, err.Error())
    }

    // 取得所有省级行政区名数组
    provinceDistrictNameArray := make([]string, 0)
    for _, provinceDistrict := range districtTable.Provinces {
        provinceDistrictNameArray = append(provinceDistrictNameArray, provinceDistrict.Name)
    }
    sortStrByPinyin(provinceDistrictNameArray)

    // 第一列省级行政区
    columnNumber := 1
    lineNo := 1
    columnName, _ := excelize.ColumnNumberToName(columnNumber)
    provinceDistrictTitle := "省份"
    f.SetCellStr(sheetName, fmt.Sprintf("%s%d", columnName, lineNo), provinceDistrictTitle)
    lineNo++
    for _, provinceDistrictName := range provinceDistrictNameArray {
        f.SetCellStr(sheetName, fmt.Sprintf("%s%d", columnName, lineNo), provinceDistrictName)
        lineNo++
    }

    // 创建省级行政区列的公式名称，名称采用第一行的
    err = f.SetDefinedName(&excelize.DefinedName{
        Name:     provinceDistrictTitle,
        RefersTo: fmt.Sprintf("'%s'!$A$%d:$A$%d", sheetName, 2, lineNo-1),
    })
    if err != nil {
        return fmt.Errorf("set defined name of provinces error: %s", err.Error())
    }

    // 空一行
    columnNumber = columnNumber + 2

    // 市级行政区名
    for _, provinceDistrict := range districtTable.Provinces {
        // 取得所有市级行政区名数组
        cityDistrictNameArray := make([]string, 0)
        for _, cityDistrict := range provinceDistrict.CityDistrictTable {
            cityDistrictNameArray = append(cityDistrictNameArray, cityDistrict.Name)
        }
        sortStrByPinyin(cityDistrictNameArray)

        columnName, _ = excelize.ColumnNumberToName(columnNumber)
        lineNo = 1
        f.SetCellStr(sheetName, fmt.Sprintf("%s%d", columnName, lineNo), provinceDistrict.Name)
        lineNo++

        for _, cityDistrictName := range cityDistrictNameArray {
            f.SetCellStr(sheetName, fmt.Sprintf("%s%d", columnName, lineNo), cityDistrictName)
            lineNo++
        }

        // 创建市级行政区列的公式名称，名称采用第一行的
        provinceDistrictName, err := f.GetCellValue(sheetName, fmt.Sprintf("%s%d", columnName, 1))
        if err != nil {
            return fmt.Errorf("get province district name error: %s", err.Error())
        }
        err = f.SetDefinedName(&excelize.DefinedName{
            Name:     provinceDistrictName,
            RefersTo: fmt.Sprintf("'%s'!$%s$%d:$%s$%d", sheetName, columnName, 2, columnName, lineNo-1),
        })
        if err != nil {
            return fmt.Errorf("set defined name of cities error: (%s) %s", provinceDistrictName, err.Error())
        }

        // 下一列市级行政区名
        columnNumber++
    }

    // 空一列
    columnNumber++

    // 区县级行政区名
    for _, provinceDistrict := range districtTable.Provinces {
        for _, cityDistrict := range provinceDistrict.CityDistrictTable {
            if cityDistrict.CountyDistrictTable == nil || len(cityDistrict.CountyDistrictTable) == 0 {
                continue
            }

            // 取得所有区县级行政区名数组
            countyDistrictNameArray := make([]string, 0)
            for _, countyDistrict := range cityDistrict.CountyDistrictTable {
                countyDistrictNameArray = append(countyDistrictNameArray, countyDistrict.Name)
            }
            sortStrByPinyin(countyDistrictNameArray)

            columnName, _ = excelize.ColumnNumberToName(columnNumber)
            lineNo = 1
            f.SetCellStr(sheetName, fmt.Sprintf("%s%d", columnName, lineNo), cityDistrict.Name)
            lineNo++

            for _, countyDistrictName := range countyDistrictNameArray {
                f.SetCellStr(sheetName, fmt.Sprintf("%s%d", columnName, lineNo), countyDistrictName)
                lineNo++
            }

            // 创建县级行政区列的公式名称，名称采用第一行的
            cityDistrictName, err := f.GetCellValue(sheetName, fmt.Sprintf("%s%d", columnName, 1))
            if err != nil {
                return fmt.Errorf("get city district name error: %s", err.Error())
            }
            err = f.SetDefinedName(&excelize.DefinedName{
                Name:     cityDistrictName,
                RefersTo: fmt.Sprintf("'%s'!$%s$%d:$%s$%d", sheetName, columnName, 2, columnName, lineNo-1),
            })
            if err != nil {
                return fmt.Errorf("set defined name of counties error: %s", err.Error())
            }

            // 下一列区县级行政区名
            columnNumber++
        }

        // 省级行政区间空一行
        columnNumber++
    }

    err = setSheet1(f)
    if err != nil {
        return err
    }
    err = f.SaveAs(xlsxFilepath)
    if err != nil {
        return fmt.Errorf("save %s error: %s", xlsxFilepath, err.Error())
    }
    return nil
}

func parseLine(lineNo int, line string) (*District, error) {
    // 使用逗号分隔每行数据
    parts := strings.Split(line, ",")
    if len(parts) != 2 {
        return nil, fmt.Errorf("invalid row format: (%d) %s, expected format: DistrictCode,DistrictName", lineNo, line)
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

// IsHongKongMacauTaiwan 判断是否为香港/澳门/台湾
func IsHongKongMacauTaiwan(name string) bool {
    return name == "香港" || name == "澳门" || name == "台湾" ||
            name == "香港特别行政区" || name == "澳门特别行政区" || name == "台湾省"
}

// IsMunicipalityCode 是否为直辖市
func IsMunicipalityCode(code uint32) bool {
    provinceCode := (code / 10000) * 10000
    return provinceCode == 110000 || // 北京市
            provinceCode == 310000 || // 上海市
            provinceCode == 120000 || // 天津市
            provinceCode == 500000 // 重庆市
}

func IsProvinceDistrictCode(code uint32) bool {
    return code%10000 == 0
}

func IsCityDistrictCode(code uint32) bool {
    return code%10000 != 0 && code%100 == 0
}

func IsCountyDistrictCode(code uint32) bool {
    return code%10000 != 0 && code%100 != 0
}

func getProvinceDistrictCode(code uint32) uint32 {
    return (code / 10000) * 10000
}

func getCityDistrictCode(code uint32) uint32 {
    return (code / 100) * 100
}

func perfectTable(table *Table) {
    for _, provinceDistrict := range table.ProvinceDistrictTable {
        for _, cityDistrict := range provinceDistrict.CityDistrictTable {
            for _, countyDistrict := range cityDistrict.CountyDistrictTable {
                cityDistrict.Counties = append(cityDistrict.Counties, countyDistrict)
            }
            sort.Slice(cityDistrict.Counties, func(i, j int) bool {
                return cityDistrict.Counties[i].Code < cityDistrict.Counties[j].Code
            })
            provinceDistrict.Cities = append(provinceDistrict.Cities, cityDistrict)
        }
        sort.Slice(provinceDistrict.Cities, func(i, j int) bool {
            return provinceDistrict.Cities[i].Code < provinceDistrict.Cities[j].Code
        })
        table.Provinces = append(table.Provinces, provinceDistrict)
    }
    sort.Slice(table.Provinces, func(i, j int) bool {
        return table.Provinces[i].Code < table.Provinces[j].Code
    })
}

func createFile(filepath string) (*os.File, *bufio.Writer) {
    file, err := os.Create(filepath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Create file://%s error: %s.\n", filepath, err.Error())
        return nil, nil
    }

    return file, bufio.NewWriter(file)
}

func createXlsxFile(ctx context.Context, sheetName string) (*excelize.File, error) {
    // 创建一个新的Excel文件
    f := excelize.NewFile()
    // 创建一个新的工作表
    _, err := f.NewSheet(sheetName)
    if err != nil {
        return nil, err
    }

    f.SetAppProps(&excelize.AppProperties{
        Application: "mooon-district",
        Company:     "mooon",
    })
    f.SetDocProps(&excelize.DocProperties{
        Creator:     "mooon",
        Subject:     "mooon-district",
        Title:       "mooon-district",
        Version:     "0.0.1",
        Description: "mooon-district",
    })
    return f, nil
}

// sortStrByPinyin 中文拼音字母序排序
func sortStrByPinyin(strArray []string) {
    // 创建一个collator，用于中文拼音字母序排序
    collator := collate.New(language.Chinese)

    // 使用sort.Slice()函数和collator对数组进行排序
    sort.Slice(strArray, func(i, j int) bool {
        return collator.CompareString(strArray[i], strArray[j]) < 0
    })
}

func setSheet1(f *excelize.File) error {
    sheetName := "Sheet1"
    // 标题行
    f.SetCellStr(sheetName, "A1", "省份")
    f.SetCellStr(sheetName, "B1", "城市")
    f.SetCellStr(sheetName, "C1", "区县")

    // 设置列宽度
    err := f.SetColWidth(sheetName, "A", "C", 20)
    if err != nil {
        return fmt.Errorf("set column width error: %s", err.Error())
    }

    // 数据行 - 省级行政区
    // 在 A2 单元格中添加数据验证
    dvRange1 := excelize.NewDataValidation(true)
    dvRange1.Sqref = "A2:A2"
    dvRange1.SetSqrefDropList("省份")
    if err := f.AddDataValidation(sheetName, dvRange1); err != nil {
        return fmt.Errorf("add data validation of provinces error: %s", err.Error())
    }

    // 数据行 - 市级行政区
    // 在 B2 单元格中添加数据验证
    dvRange2 := excelize.NewDataValidation(true)
    dvRange2.Sqref = "B2:B2"
    dvRange2.SetSqrefDropList("INDIRECT(A2)")
    if err := f.AddDataValidation(sheetName, dvRange2); err != nil {
        return fmt.Errorf("add data validation of cities error: %s", err.Error())
    }

    // 数据行 - 区县级行政区
    // 在 C2 单元格中添加数据验证
    dvRange3 := excelize.NewDataValidation(true)
    dvRange3.Sqref = "C2:C2"
    dvRange3.SetSqrefDropList("INDIRECT(B2)")
    if err := f.AddDataValidation(sheetName, dvRange3); err != nil {
        return fmt.Errorf("add data validation of counties error: %s", err.Error())
    }

    return nil
}
