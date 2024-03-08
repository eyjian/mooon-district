// Wrote by yijian on 2024/03/09
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/eyjian/mooon-district/district"
    "os"
)

func main() {
    ctx := context.Background()
    districtTable, err := district.LoadDistrict(ctx, "./district-2022.csv")
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
