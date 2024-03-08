// Wrote by yijian on 2024/03/09
package main

import (
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
        fmt.Fprintf(os.Stderr, "load district error: %s\n", err.Error())
    } else {
        jsonBytes, err := json.Marshal(*districtTable)
        if err != nil {
            fmt.Fprintf(os.Stderr, "json marshal error: %s\n", err.Error())
        } else {
            fmt.Fprintf(os.Stdout, "%s\n", string(jsonBytes))
        }
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
