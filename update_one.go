package dbsync

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// 插入增量数据时的配置信息
type UpdateOptions struct {
	Columns      []string               // 列名称
	ColumnTypes  []string               // 列类型
	FixedFields  map[string]interface{} // 固定的插入列
	UniqueFields []string               // 唯一键或主键的列名称列表
}

// 插入单条数据
func DoUpdateOne(db executor, tableName string, data []interface{}, options UpdateOptions) (err error) {
	// 列名和值的映射关系
	mapItem := map[string]interface{}{}
	for i, fieldName := range options.Columns {
		mapItem[fieldName] = convertUpdateType(data[i], options.ColumnTypes[i])
	}
	for k, v := range options.FixedFields {
		mapItem[k] = v
	}
	// 生成SQL语句的列名、问号、值列表
	uniqueMap := make(map[string]bool)
	for _, k := range options.UniqueFields {
		uniqueMap[k] = true
	}
	columns := make([]string, len(mapItem))
	updateColumns := make([]string, 0)
	questions := make([]string, len(mapItem))
	values := make([]interface{}, len(mapItem))
	num := 0
	for k, v := range mapItem {
		columns[num] = k
		if uniqueMap[k] != true {
			updateColumns = append(updateColumns, fmt.Sprintf("%s=VALUES(%s)", k, k))
		}
		questions[num] = "?"
		values[num] = v
		num++
	}
	colStr, questionStr, updateStr := strings.Join(columns, ","), strings.Join(questions, ","), strings.Join(updateColumns, ",")
	// 生成并执行SQL语句
	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s", tableName, colStr, questionStr, updateStr)
	_, err = db.Exec(sqlStr, values...)
	return
}

// 类型转换方法
func convertUpdateType(data interface{}, columnType string) interface{} {
	switch columnType {
	case "time":
		if data != nil {
			return time.Unix(int64(data.(float64)), 0)
		} else {
			return nil
		}
	default:
		return data
	}
}
