package fault

import "streetlight/pkg/query"

// 故障列表允许参与筛选的列名(内部常量, 不接受外部透传)。
const (
	colFaultNo    = "fault_no"
	colLampCode   = "lamp_code"
	colRoadName   = "road_name"
	colDesc       = "description"
	colStatus     = "status"
	colFaultType  = "fault_type"
	colFaultLevel = "fault_level"
	colSource     = "source"
	colLampID     = "lamp_id"
	colReportedAt = "reported_at"
)

// keywordColumns 是故障关键词模糊匹配的固定列集合:
// 故障单号 / 路灯编号 / 道路 / 描述。
func keywordColumns() []string {
	return []string{colFaultNo, colLampCode, colRoadName, colDesc}
}

// Conditions 把故障查询条件转换为统一的 Condition 列表。
// 任何读取故障列表的页面(故障登记列表、未来的聚合读模型)都必须复用这里,
// 保证同一组入参生成完全相同的 WHERE。
func (f Filter) Conditions() []query.Condition {
	conditions := []query.Condition{
		query.Keyword(f.Keyword, keywordColumns()...),
		query.Eq(colStatus, f.Status),
		query.Eq(colFaultType, f.FaultType),
		query.Eq(colFaultLevel, f.FaultLevel),
		query.Eq(colSource, f.Source),
		query.EqUint(colLampID, f.LampID),
		query.Eq(colRoadName, f.RoadName),
		query.TimeRange(colReportedAt, f.ReportedFrom, f.ReportedTo),
	}
	if f.OnlyOpen {
		conditions = append(conditions, query.In(colStatus, OpenStatuses()))
	}
	return conditions
}
