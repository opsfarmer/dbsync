package dbsync

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

// 插入数据时的配置信息
type UpdateOptions struct {
	Columns      []string               // 列名称
	ColumnTypes  []string               // 列类型
	FixedFields  map[string]interface{} // 固定的插入列
	UniqueFields []string               // 唯一键或主键的列名称列表
}

// 通用插入数据
func update(db gorm.SQLCommon, tableName string, data [][]interface{}, options UpdateOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	// 列名和索引的映射关系
	mapItem := map[string]int{}
	for i, fieldName := range options.Columns {
		mapItem[fieldName] = i
	}
	i, columnsLen := 0, len(options.Columns)
	for k := range options.FixedFields {
		mapItem[k] = columnsLen + i
		i++
	}
	// 生成SQL语句的列名、问号、值列表
	uniqueMap := make(map[string]bool)
	for _, k := range options.UniqueFields {
		uniqueMap[k] = true
	}
	mapItemLen := len(mapItem)
	columns := make([]string, mapItemLen)
	updateColumns := make([]string, 0)
	questions := make([]string, mapItemLen)
	for k, num := range mapItem {
		columns[num] = k
		if uniqueMap[k] != true {
			updateColumns = append(updateColumns, fmt.Sprintf("%s=VALUES(%s)", k, k))
		}
		questions[num] = "?"
	}
	colStr, questionStr, updateStr := strings.Join(columns, ","), fmt.Sprintf("(%s)", strings.Join(questions, ",")), strings.Join(updateColumns, ",")
	dataLen := len(data)
	values := make([]interface{}, mapItemLen*dataLen)
	allQuestions := make([]string, dataLen)
	for k, dataItem := range data {
		for fieldName, num := range mapItem {
			index := mapItemLen*k + num
			if num >= columnsLen {
				values[index] = options.FixedFields[fieldName]
			} else {
				values[index] = convertUpdateType(dataItem[num], options.ColumnTypes[num])
			}
		}
		allQuestions[k] = questionStr
	}
	allQuestionStr := strings.Join(allQuestions, ",")
	// 生成并执行SQL语句
	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s ON DUPLICATE KEY UPDATE %s", tableName, colStr, allQuestionStr, updateStr)
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
