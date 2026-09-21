package lamp

import "streetlight/pkg/query"

// 台账列表允许参与筛选的列名。仅内部常量允许进入 SQL, 不接收外部透传列名。
const (
	ledgerColCode      = "code"
	ledgerColName      = "name"
	ledgerColRoadName  = "road_name"
	ledgerColDistrict  = "district"
	ledgerColAddress   = "address"
	ledgerColLampType  = "lamp_type"
	ledgerColRunStatus = "run_status"
)

// LedgerKeywordColumns 是台账关键词模糊匹配的固定列集合:
// 编号 / 名称 / 道路 / 地址。
func LedgerKeywordColumns() []string {
	return []string{ledgerColCode, ledgerColName, ledgerColRoadName, ledgerColAddress}
}

// LedgerConditions 构造路灯台账列表的统一筛选条件。
//
// 台账列表(/lamps)与维修状态读模型(/status/lamps)必须共用同一份条件,
// 保证传入相同 keyword/road_name/lamp_type/run_status 时,
// 两个页面统计的总数、行的归属与排序完全一致。
//
// prefix 为表别名前缀: 直接查询 lamp 主表时传 "", 带别名的读模型传 "lamp."。
func LedgerConditions(keyword, roadName, district, lampType, runStatus, prefix string) []query.Condition {
	col := func(name string) string { return prefix + name }
	return []query.Condition{
		query.Keyword(keyword, col(ledgerColCode), col(ledgerColName), col(ledgerColRoadName), col(ledgerColAddress)),
		query.Eq(col(ledgerColRoadName), roadName),
		query.Eq(col(ledgerColDistrict), district),
		query.Eq(col(ledgerColLampType), lampType),
		query.Eq(col(ledgerColRunStatus), runStatus),
	}
}

// DeriveRunStatus 依据故障状态计数推导路灯运行状态, 是该字段业务含义的唯一权威实现。
//
// 规则(与故障模块的状态机对应):
//   - 存在"维修中"故障 -> 维修中(maintenance), 维修中优先于待处理;
//   - 否则存在"待处理"故障 -> 故障(fault);
//   - 没有未闭环故障 -> 正常(normal)。
//
// 计数由故障模块依据 fault 表提供, 本函数不依赖故障包, 保持 lamp <- fault 的依赖方向。
// 已闭环(已修复/已关闭)与停用(offline)不参与自动推导: 停用只能由人工设置并保留。
func DeriveRunStatus(processingCount, pendingCount int64) string {
	switch {
	case processingCount > 0:
		return RunStatusMaintenance
	case pendingCount > 0:
		return RunStatusFault
	default:
		return RunStatusNormal
	}
}
