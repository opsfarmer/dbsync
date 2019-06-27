package dbsync

// 获取增量数据和更新增量数据之间传递的参数
type Params struct {
	Columns []string        `json:"columns"` // 列名称
	Data    [][]interface{} `json:"data"`    // 待同步的数据，每一行是一条数据，与列名称一一对应
}
