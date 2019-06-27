# dbsync

[![Build Status](https://travis-ci.org/cuckoopark/dbsync.svg?branch=master)](https://travis-ci.org/cuckoopark/dbsync)
[![Latest Tag](https://img.shields.io/github/tag/cuckoopark/dbsync.svg)](https://github.com/cuckoopark/dbsync/releases/latest)

这是用来实现两个数据库的同一张表的增量同步。

* 支持按照某种格式增量获取表中的待同步数据。
* 支持按照列名称向数据库的表中批量插入待同步的数据。

### 安装

```shell
go get -u github.com/cuckoopark/dbsync
```
