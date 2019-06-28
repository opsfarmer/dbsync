# dbsync

[![Build Status](https://travis-ci.org/cuckoopark/dbsync.svg?branch=master)](https://travis-ci.org/cuckoopark/dbsync)
[![Latest Tag](https://img.shields.io/github/tag/cuckoopark/dbsync.svg)](https://github.com/cuckoopark/dbsync/releases/latest)

这是用来实现两个MySQL数据库中的具有相同字段表的增量同步。

* 支持按照某种格式增量获取表中的待同步数据。
* 支持按照列名称向数据库的表中批量插入待同步的数据。

### 安装

```shell
go get -u github.com/cuckoopark/dbsync
```

### 数据库配置

在每一张需要同步的表中，应该有一个`update_time`更新时间的非空字段(名字可以不一样，但是类型必须是时间相关类型)，用来按照更新时间获取最新的更新数据。

这个字段需要在数据更新时，自动更新为当前时间戳，用于记录数据更新的时间。

例如`update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP`的字段设置。

### 获取增量更新数据

可以批量获取一张表的最新更新的数据，方法如下：

```go
func DoFetch(db *sql.DB, tableName string, options FetchOptions) (FetchResult, error)
```

其中参数说明：

* `db`：数据库操作句柄。
* `tableName`：表名称。
* `options`：获取时的配置信息，`FetchOptions`格式如下所示：
  - `IgnoreFields`：需要忽略的列名称，获取数据的结果不包含该列。
  - `PageNumber`：分页获取增量的页码，从1开始。
  - `PageSize`：分页获取增量的页大小，判断分页是否结束，只需要判断获取结果的数量是否小于页大小即可。
  - `UpdateTimeFieldName`：更新时间(即上面说明的`update_time`)所在列的列名称。
  - `LastUpdateTime`：上次更新的时间戳，大于这个时间戳开始查询，如果为0，则表示查询全部数据。
  - `WhereSqlStmt`：自定义SQL查询语句的Where子句，与更新时间的条件(`[UpdateTimeFieldName] > ?`)是`AND`的关系。
  - `WhereSqlArgs`：自定义SQL查询语句的Where子句的参数列表。

获取的结果，是`FetchResult`格式的结构体：

* `columns`：列名称列表。
* `column_types`：列的数据类型列表，与列名称列表一一对应。
* `data`：最新更新的数据，二维数组，每一行是一条数据，里面的值与列名称是一一对应关系。

注：时间类型的列，获取的结果`time.Time`会被转换为时间戳传递，用于节省数据长度。

### 插入单条更新数据

接口为：

```go
func DoUpdateOne(db executor, tableName string, data []interface{}, options UpdateOptions) error
```

其中参数说明：

* `db`：数据库操作句柄。
* `tableName`：表名称。
* `data`：一条数据，里面的值与`options.Columns`一一对应。
* `options`：插入时的配置信息，`UpdateOptions`格式如下所示：
  - `Columns`：列名称列表，参照`DoFetch`返回的结果。
  - `ColumnTypes`：列的数据类型列表，参照`DoFetch`返回的结果。
  - `FixedFields`：固定的插入列，因为在`DoFetch`中会配置忽略一些列，所以这里可以给这些列设置值。
  - `UniqueFields`：唯一键或主键的列名称列表，在插入失败时，更新操作不更新唯一键或主键。

该接口无返回值。

插入或更新会调用`INSERT INTO ... VALUES (...) ON DUPLICATE KEY UPDATE ...`这种SQL语句来执行，不采用`REPLACE INTO`的原因是它的更新会先删除旧数据，再插入新数据，可能导致一些忽略的字段被修改。

### 批量插入更新数据

接口为：

```go
func DoUpdate(db *sql.DB, tableName string, data [][]interface{}, options UpdateMultiOptions) error
```

相比较插入单条更新数据的接口，只是数据变成了二维数组，`options`里面多了一个配置项：

* `BatchCount`：一次批量插入的条数，用于加快插入或更新的执行速度。