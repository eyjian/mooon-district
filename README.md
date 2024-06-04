# mooon-district

一个行政区数据工具，可以用来生成 json 格式数据、SQL 插入语句和 csv 格式的数据。

# 数据说明

数据来源于[民政部官网](https://www.mca.gov.cn/n156/n186/index.html)的公开数据，只支持三级行政区：省/自治区/直辖市、市/州/盟、区/县/县级市/旗，不支持到乡镇和街道这一级行政区。

# 安装工具

```shell
go install github.com/eyjian/mooon-district@latest
```

# 生成 json 格式数据

```shell
mooon-district -f ./district-2022.csv -with-json=true
```

# 生成 csv 格式数据：

```shell
mooon-district -f ./district-2022.csv -with-csv=true
```

# 生成 sql 插入语句：

```shell
mooon-district -f ./district-2022.csv -with-sql=true
```

使用时，可同时指定：-with-json=true、-with-csv=true 和 -with-sql=true：

```shell
mooon-district -f ./district-2022.csv -with-sql=true -with-csv=true -with-json=true
```

如果是新增更新，可指定参数“-with-sql-ignore”值为 true 生成“INSERT IGNORE INTO”语句。

# 特别说明

* 省直辖县/县级市/旗，没有父级行政区地级市，它的行政区代码仍然是县/县级市/旗级的，如河南省的济源市