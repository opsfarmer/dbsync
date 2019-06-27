package dbsync

// 插入单条数据
func DoUpdateOne(db executor, tableName string, data []interface{}, options UpdateOptions) (err error) {
	return update(db, tableName, [][]interface{}{data}, options)
}
