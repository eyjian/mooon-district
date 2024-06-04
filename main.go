// Wrote by yijian on 2024/03/09
package main

import (
    "context"
    "flag"
    "fmt"
    "github.com/eyjian/mooon-district/district"
    "os"
)

var (
    help             = flag.Bool("h", false, "Display a help message and exit.")
    districtDataFile = flag.String("f", "", "File that stores district data.")

    withJson       = flag.Bool("with-json", false, "Whether to generate json format data.")
    withJsonIndent = flag.Bool("with-json-indent", true, "Whether JSON format is indented.")
    jsonIndent     = flag.String("json-indent", "  ", "Json indent when -with-json-indent is enabled.")
    jsonPrefix     = flag.String("json-prefix", "", "Prefix for each line when -with-json-indent is enabled.")

    withCsv      = flag.Bool("with-csv", false, "Whether to generate csv format data.")
    csvDelimiter = flag.String("csv-delimiter", ",", "Delimiter of csv data.")
    csvWithCode  = flag.Bool("csv-with-code", true, "Whether the csv format outputs the code column.")

    withSql       = flag.Bool("with-sql", false, "Whether to generate sql data.")
    withSqlIgnore = flag.Bool("with-sql-ignore", false, "Use `INSERT IGNORE` to ignore existing.")
    sqlTable      = flag.String("sql-table", "t_dict_district", "Table name for sql data.")
)

func main() {
    flag.Parse()
    if *help {
        usage()
        os.Exit(1)
    }
    if !checkParameters() {
        os.Exit(1)
    }

    ctx := context.Background()
    districtTable, err := district.LoadDistrict(ctx, *districtDataFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Load district error: %s.\n", err.Error())
        os.Exit(2)
    }

    done := false
    if *withJson {
        done = true
        err := district.GenerateJson(districtTable, "example.json", *withJsonIndent, *jsonIndent, *jsonPrefix)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Generate json error: %s.\n", err.Error())
            os.Exit(3)
        }
    }
    if *withCsv {
        done = true
        err := district.GenerateCsv(districtTable, "example.csv", *csvDelimiter, *csvWithCode)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Generate csv error: %s.\n", err.Error())
            os.Exit(3)
        }
    }
    if *withSql {
        done = true
        err := district.GenerateSql(districtTable, "example.sql", *sqlTable, *withSqlIgnore)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Generate sql error: %s.\n", err.Error())
            os.Exit(3)
        }
    }
    if !done {
        fmt.Fprintf(os.Stderr, "Do nothing.\n")
        os.Exit(4)
    }
}

func usage() {
    flag.Usage()
}

func checkParameters() bool {
    if len(*districtDataFile) == 0 {
        fmt.Fprintf(os.Stderr, "Parameter -f is not set.\n")
        return false
    }

    if *withSql {
        if len(*sqlTable) == 0 {
            fmt.Fprintf(os.Stderr, "Parameter -sql-table is not set.\n")
            return false
        }
    }
    return true
}