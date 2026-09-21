package fault

import (
	"testing"
	"time"
)

// TestIsOverdueLocksDefinition 锁定逾期判定的唯一权威定义:
// 仅"仍待处理"且"登记时刻早于 now-阈值"才算超期。
// 看板的超期计数/列表与任何页面的超期角标都必须与此函数一致。
func TestIsOverdueLocksDefinition(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.Local)
	cutoff := OverdueCutoff(now) // 2026-09-20 12:00

	mk := func(status string, reported time.Time) Fault {
		return Fault{Status: status, ReportedAt: reported}
	}

	cases := []struct {
		name   string
		entity Fault
		want   bool
	}{
		{"恰好超期", mk(StatusPending, cutoff.Add(-time.Minute)), true},
		{"恰好未超期(分界不包含)", mk(StatusPending, cutoff.Add(time.Minute)), false},
		{"整24小时不算超期(严格早于)", mk(StatusPending, cutoff), false},
		{"维修中不算超期", mk(StatusProcessing, cutoff.Add(-time.Hour)), false},
		{"已修复不算超期", mk(StatusRepaired, cutoff.Add(-time.Hour)), false},
		{"已关闭不算超期", mk(StatusClosed, cutoff.Add(-time.Hour)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsOverdue(tc.entity, now); got != tc.want {
				t.Fatalf("IsOverdue(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}

	// 阈值常量必须是 24 小时, 与看板暴露给前端的 overdue_threshold_hours 一致。
	if OverdueThreshold != 24*time.Hour {
		t.Fatalf("逾期阈值应为 24h, 实际 %s", OverdueThreshold)
	}
}
