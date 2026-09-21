package repair

import "streetlight/pkg/query"

// 维修记录允许参与筛选的列名(内部常量, 不接受外部透传)。
const (
	colRepairNo   = "repair_no"
	colFaultNo    = "fault_no"
	colLampCode   = "lamp_code"
	colRepairman  = "repairman"
	colRepairTeam = "repair_team"
	colFaultID    = "fault_id"
	colLampID     = "lamp_id"
	colStatus     = "status"
	colResult     = "result"
	colStartedAt  = "started_at"
)

// keywordColumns 是维修记录关键词模糊匹配的固定列集合:
// 维修单号 / 故障单号 / 路灯编号 / 维修人员。
func keywordColumns() []string {
	return []string{colRepairNo, colFaultNo, colLampCode, colRepairman}
}

// Conditions 把维修查询条件转换为统一的 Condition 列表。
// 维修记录列表与任何跨模块读模型都必须复用这里, 保证相同入参生成相同 WHERE。
func (f Filter) Conditions() []query.Condition {
	return []query.Condition{
		query.Keyword(f.Keyword, keywordColumns()...),
		query.EqUint(colFaultID, f.FaultID),
		query.EqUint(colLampID, f.LampID),
		query.Eq(colRepairman, f.Repairman),
		query.Eq(colRepairTeam, f.RepairTeam),
		query.Eq(colStatus, f.Status),
		query.Eq(colResult, f.Result),
		query.TimeRange(colStartedAt, f.StartedFrom, f.StartedTo),
	}
}
