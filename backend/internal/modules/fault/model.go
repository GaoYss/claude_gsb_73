package fault

import (
	"time"
)

// 故障处理状态。
const (
	StatusPending    = "pending"    // 待处理
	StatusProcessing = "processing" // 维修中
	StatusRepaired   = "repaired"   // 已修复
	StatusClosed     = "closed"     // 已关闭
)

// OverdueThreshold 是"超期未处理"的统一时长阈值:
// 处于待处理状态且上报时间早于 now - 阈值 的故障记为超期。
// 看板概览、超期列表与未来的筛选能力都引用此常量, 保证口径一致。
const OverdueThreshold = 24 * time.Hour

// 故障等级(紧急程度)。
const (
	LevelLow    = "low"    // 一般
	LevelNormal = "normal" // 普通
	LevelHigh   = "high"   // 紧急
	LevelUrgent = "urgent" // 特急
)

// 故障来源。
const (
	SourceInspection = "inspection" // 巡检发现
	SourceCitizen    = "citizen"    // 市民上报
	SourceMonitoring = "monitoring" // 系统告警
	SourceOther      = "other"      // 其它
)

// Statuses 返回全部故障状态取值。
func Statuses() []string {
	return []string{StatusPending, StatusProcessing, StatusRepaired, StatusClosed}
}

// Levels 返回全部故障等级取值。
func Levels() []string {
	return []string{LevelLow, LevelNormal, LevelHigh, LevelUrgent}
}

// Sources 返回全部故障来源取值。
func Sources() []string {
	return []string{SourceInspection, SourceCitizen, SourceMonitoring, SourceOther}
}

// FaultTypes 返回全部故障类型取值。
func FaultTypes() []string {
	return []string{
		"灯不亮", "灯光闪烁", "灯具常亮", "灯杆倾斜",
		"线路故障", "控制箱故障", "灯具破损", "其他",
	}
}

// IsValidStatus 校验故障状态取值。
func IsValidStatus(status string) bool {
	for _, item := range Statuses() {
		if item == status {
			return true
		}
	}
	return false
}

// IsOpen 判断故障是否仍处于未闭环状态。
func IsOpen(status string) bool {
	for _, item := range OpenStatuses() {
		if status == item {
			return true
		}
	}
	return false
}

// OpenStatuses 返回未闭环故障的状态集合(待处理 + 维修中)。
// 所有"仅看未闭环"的筛选、统计、存在性校验都必须引用该集合,
// 禁止在各处内联 []string{pending, processing}。
func OpenStatuses() []string {
	return []string{StatusPending, StatusProcessing}
}

// OverdueCutoff 返回超期判定的分界时刻: 在 now 视角下,
// reported_at 早于该时刻且仍处于待处理状态即为超期。
func OverdueCutoff(now time.Time) time.Time {
	return now.Add(-OverdueThreshold)
}

// IsOverdue 判断某条故障在 now 视角下是否"超期未处理"。
// 这是逾期判定的唯一权威定义: 仅待处理(pending)且登记时长超过阈值才算,
// 维修中/已修复/已关闭均不算超期。
func IsOverdue(entity Fault, now time.Time) bool {
	return entity.Status == StatusPending && entity.ReportedAt.Before(OverdueCutoff(now))
}

// canTransitTo 校验状态流转是否合法。
// 待处理 -> 维修中 / 已关闭, 维修中 -> 已修复 / 已关闭, 已修复 -> 已关闭 / 返修(维修中)。
func canTransitTo(from, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case StatusPending:
		return to == StatusProcessing || to == StatusClosed
	case StatusProcessing:
		return to == StatusRepaired || to == StatusClosed
	case StatusRepaired:
		return to == StatusClosed || to == StatusProcessing
	default:
		return false
	}
}

// Fault 故障登记记录, 串联路灯台账与维修记录。
type Fault struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	FaultNo        string     `gorm:"size:64;uniqueIndex;not null" json:"fault_no"`
	LampID         uint       `gorm:"index;not null" json:"lamp_id"`
	LampCode       string     `gorm:"size:64;index" json:"lamp_code"`
	RoadName       string     `gorm:"size:128;index" json:"road_name"`
	FaultType      string     `gorm:"size:32;index;not null" json:"fault_type"`
	FaultLevel     string     `gorm:"size:32;index;not null;default:normal" json:"fault_level"`
	Source         string     `gorm:"size:32;index" json:"source"`
	Description    string     `gorm:"size:512" json:"description"`
	Reporter       string     `gorm:"size:64" json:"reporter"`
	ReporterPhone  string     `gorm:"size:32" json:"reporter_phone"`
	ReportedAt     time.Time  `gorm:"index;not null" json:"reported_at"`
	Status         string     `gorm:"size:32;index;not null;default:pending" json:"status"`
	RepairCount    int        `gorm:"not null;default:0" json:"repair_count"`
	LatestRepairID *uint      `json:"latest_repair_id"`
	ClosedAt       *time.Time `json:"closed_at"`
	CloseRemark    string     `gorm:"size:255" json:"close_remark"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TableName 指定表名。
func (Fault) TableName() string { return "fault" }
