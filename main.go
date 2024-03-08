// Wrote by yijian on 2024/03/09
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "github.com/eyjian/mooon-district/district"
    "os"
)

var (
    help             = flag.Bool("h", false, "Display a help message and exit.")
    districtDataFile = flag.String("f", "", "File that stores district data.")
    withJson         = flag.Bool("with-json", false, "Whether to generate json data.")
    withSql          = flag.Bool("with-sql", false, "Whether to generate sql data.")
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
        if !generateJson(districtTable) {
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

    return true
}

func createFile(filepath string) (*os.File, *bufio.Writer) {
    file, err := os.Create(filepath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Create file://%s error: %s.\n", filepath, err.Error())
        return nil, nil
    }

    return file, bufio.NewWriter(file)
}

func generateJson(districtTable *district.Table) bool {
    jsonBytes, err := json.Marshal(*districtTable)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Json marshal error: %s.\n", err.Error())
        return false
    }

    filepath := "district.json"
    file, writer := createFile(filepath)
    if file == nil {
        return false
    }
    defer file.Close()

    _, err = writer.WriteString(string(jsonBytes))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Write file://%s error: %s.\n", filepath, err.Error())
        return false
    }
    err = writer.Flush()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Flush file://%s error: %s.\n", filepath, err.Error())
        return false
    }

    return true
}
