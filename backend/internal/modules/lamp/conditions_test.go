package lamp

import "testing"

// TestDeriveRunStatusLocksRules 锁定运行状态推导这一唯一实现的优先级:
// 维修中 > 待处理 > 正常。任何页面展示的 run_status 都应源于此函数,
// 不允许在别处再写一份 switch。
func TestDeriveRunStatusLocksRules(t *testing.T) {
	cases := []struct {
		name       string
		processing int64
		pending    int64
		want       string
	}{
		{"无故障", 0, 0, RunStatusNormal},
		{"仅待处理", 0, 1, RunStatusFault},
		{"仅维修中", 1, 0, RunStatusMaintenance},
		{"维修中优先于待处理", 1, 2, RunStatusMaintenance},
		{"多条待处理仍是故障", 0, 3, RunStatusFault},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeriveRunStatus(tc.processing, tc.pending); got != tc.want {
				t.Fatalf("DeriveRunStatus(%d,%d) = %s, want %s", tc.processing, tc.pending, got, tc.want)
			}
		})
	}
}

// TestLedgerConditionsIgnoreBlank 保证未填写的台账条件不参与 WHERE,
// 这是"同一条件条数一致"的基础: 空值在任何页面语义都是不过滤。
func TestLedgerConditionsCount(t *testing.T) {
	conds := LedgerConditions("", "", "", "", "", "")
	if len(conds) != 5 {
		t.Fatalf("台账条件数量应为 5, 实际 %d", len(conds))
	}
	prefixed := LedgerConditions("", "", "", "", "", "lamp.")
	if len(prefixed) != 5 {
		t.Fatalf("带前缀的台账条件数量应为 5, 实际 %d", len(prefixed))
	}
}
