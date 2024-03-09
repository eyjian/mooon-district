# mooon-district

一个行政区数据工具，可以用来生成 json 格式数据、SQL 插入语句和 csv 格式的数据。

# 生成 json 格式数据

```shell
./district_tool -f ./district-2022.csv -with-json=true
```

# 生成 csv 格式数据：

```shell
./district_tool -f ./district-2022.csv -with-csv=true
```

# 生成 sql 插入语句：

```shell
./district_tool -f ./district-2022.csv -with-sql=true
```

使用时，可同时指定：-with-json=true、-with-csv=true 和 -with-sql=true 。

# 特别说明

* 省直辖县/县级市/旗，没有父级行政区地级市，它的行政区代码仍然是县/县级市/旗级的，如河南省的济源市
