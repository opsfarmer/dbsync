package dbsync

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// 获取增量数据时的配置信息
type FetchOptions struct {
	IgnoreFields        []string      // 忽略的列名称
	PageNumber          int           // 分页获取增量的页码，从1开始
	PageSize            int           // 分页获取增量的页大小
	UpdateTimeFieldName string        // 更新时间所在列的列名称
	LastUpdateTime      int64         // 从哪个时间戳开始查询，这是大于的关系
	WhereSqlStmt        string        // 自定义SQL查询语句的Where子句
	WhereSqlArgs        []interface{} // 自定义SQL查询语句的Where子句的参数列表
}

// 获取增量更新的数据
func DoFetch(db *sql.DB, tableName string, options FetchOptions) (rsp Params, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			return
		}
	}()
	// 参数处理
	if options.UpdateTimeFieldName == "" {
		err = errors.New("options.UpdateTimeFieldName must be not nil")
		return
	}
	if options.PageNumber <= 0 {
		options.PageNumber = 1
	}
	if options.PageSize <= 0 {
		options.PageSize = 100
	}
	// 拼接SQL语句
	whereStmt := fmt.Sprintf("%s > ?", options.UpdateTimeFieldName)
	whereArgs := []interface{}{time.Unix(options.LastUpdateTime, 0)}
	if options.WhereSqlStmt != "" {
		whereStmt = fmt.Sprintf("%s AND (%s)", whereStmt, options.WhereSqlStmt)
		whereArgs = append(whereArgs, options.WhereSqlArgs...)
	}
	offset, size := (options.PageNumber-1)*options.PageSize, options.PageSize
	sqlStmt := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY %s ASC LIMIT %d OFFSET %d",
		tableName, whereStmt, options.UpdateTimeFieldName, size, offset)
	// 执行查询语句
	rows, err := db.Query(sqlStmt, whereArgs...)
	if err != nil {
		return
	}
	// 处理结果集的有效列
	columns, err := rows.Columns()
	if err != nil {
		return
	}
	columnValidMap := make(map[int]bool)
	for i := range columns {
		columnValidMap[i] = true
	}
	if len(options.IgnoreFields) > 0 {
		ignoreMap := make(map[string]bool)
		for _, ignoreFieldName := range options.IgnoreFields {
			ignoreMap[ignoreFieldName] = true
		}
		for j, columnName := range columns {
			if ignoreMap[columnName] == true {
				columnValidMap[j] = false
			}
		}
	}
	for i, columnName := range columns {
		if columnValidMap[i] {
			rsp.Columns = append(rsp.Columns, columnName)
			rsp.ColumnTypes = append(rsp.ColumnTypes, "")
		}
	}
	// 处理结果集的数据
	validColumnCount := len(rsp.Columns)
	cache := make([]interface{}, len(columns))
	for i := range cache {
		var tmp interface{}
		cache[i] = &tmp
	}
	for rows.Next() {
		if err = rows.Scan(cache...); err != nil {
			return
		}
		item := make([]interface{}, validColumnCount)
		p := 0
		for j, data := range cache {
			if columnValidMap[j] {
				newData, dataType := convertFetchType(data)
				item[p] = newData
				if rsp.ColumnTypes[p] == "" && len(dataType) > 0 {
					rsp.ColumnTypes[p] = dataType
				}
				p++
			}
		}
		rsp.Data = append(rsp.Data, item)
	}
	err = rows.Close()
	return
}

// 类型转换方法
func convertFetchType(data interface{}) (interface{}, string) {
	item := *data.(*interface{})
	switch item.(type) {
	case nil: // 空值
		return nil, ""
	case []uint8: // 字符串
		return string(item.([]uint8)), "string"
	case time.Time: // 时间类型
		return item.(time.Time).Unix(), "time"
	case int, int8, int16, int32, int64, float32, float64, byte: // 数字型
		return item, "number"
	case bool: // 布尔型
		return item, "bool"
	default:
		return item, ""
	}
}
